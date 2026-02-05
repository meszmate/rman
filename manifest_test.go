package rman

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"

	fb "github.com/google/flatbuffers/go"
	rmfb "github.com/meszmate/rman/flatbuffers"
)

// buildFlatBufferBody builds a FlatBuffers manifest body for testing.
func buildFlatBufferBody(t *testing.T, opts testBodyOpts) []byte {
	t.Helper()
	builder := fb.NewBuilder(1024)

	// Build tags
	tagOffsets := make([]fb.UOffsetT, len(opts.tags))
	for i, tag := range opts.tags {
		name := builder.CreateString(tag.name)
		rmfb.TagStart(builder)
		rmfb.TagAddId(builder, tag.id)
		rmfb.TagAddName(builder, name)
		tagOffsets[i] = rmfb.TagEnd(builder)
	}

	// Build directories
	dirOffsets := make([]fb.UOffsetT, len(opts.dirs))
	for i, dir := range opts.dirs {
		name := builder.CreateString(dir.name)
		rmfb.DirectoryStart(builder)
		rmfb.DirectoryAddId(builder, dir.id)
		rmfb.DirectoryAddParentId(builder, dir.parentID)
		rmfb.DirectoryAddName(builder, name)
		dirOffsets[i] = rmfb.DirectoryEnd(builder)
	}

	// Build bundles with chunks
	bundleOffsets := make([]fb.UOffsetT, len(opts.bundles))
	for i, bundle := range opts.bundles {
		chunkOffsets := make([]fb.UOffsetT, len(bundle.chunks))
		for j, chunk := range bundle.chunks {
			rmfb.ChunkStart(builder)
			rmfb.ChunkAddId(builder, chunk.id)
			rmfb.ChunkAddCompressedSize(builder, chunk.compressedSize)
			rmfb.ChunkAddUncompressedSize(builder, chunk.uncompressedSize)
			chunkOffsets[j] = rmfb.ChunkEnd(builder)
		}
		rmfb.BundleStartChunksVector(builder, len(chunkOffsets))
		for j := len(chunkOffsets) - 1; j >= 0; j-- {
			builder.PrependUOffsetT(chunkOffsets[j])
		}
		chunksVec := builder.EndVector(len(chunkOffsets))

		rmfb.BundleStart(builder)
		rmfb.BundleAddId(builder, bundle.id)
		rmfb.BundleAddChunks(builder, chunksVec)
		bundleOffsets[i] = rmfb.BundleEnd(builder)
	}

	// Build files
	fileOffsets := make([]fb.UOffsetT, len(opts.files))
	for i, file := range opts.files {
		name := builder.CreateString(file.name)

		// Build chunk IDs vector
		rmfb.FileStartChunkIdsVector(builder, len(file.chunkIDs))
		for j := len(file.chunkIDs) - 1; j >= 0; j-- {
			builder.PrependUint64(file.chunkIDs[j])
		}
		chunkIDsVec := builder.EndVector(len(file.chunkIDs))

		var symlinkOff fb.UOffsetT
		if file.symlink != "" {
			symlinkOff = builder.CreateString(file.symlink)
		}

		rmfb.FileStart(builder)
		rmfb.FileAddId(builder, file.id)
		rmfb.FileAddDirectoryId(builder, file.dirID)
		rmfb.FileAddSize(builder, file.size)
		rmfb.FileAddName(builder, name)
		rmfb.FileAddTagBitmask(builder, file.tagBitmask)
		rmfb.FileAddChunkIds(builder, chunkIDsVec)
		if symlinkOff != 0 {
			rmfb.FileAddSymlink(builder, symlinkOff)
		}
		fileOffsets[i] = rmfb.FileEnd(builder)
	}

	// Build manifest root
	rmfb.ManifestStartBundlesVector(builder, len(bundleOffsets))
	for i := len(bundleOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(bundleOffsets[i])
	}
	bundlesVec := builder.EndVector(len(bundleOffsets))

	rmfb.ManifestStartTagsVector(builder, len(tagOffsets))
	for i := len(tagOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(tagOffsets[i])
	}
	tagsVec := builder.EndVector(len(tagOffsets))

	rmfb.ManifestStartFilesVector(builder, len(fileOffsets))
	for i := len(fileOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(fileOffsets[i])
	}
	filesVec := builder.EndVector(len(fileOffsets))

	rmfb.ManifestStartDirectoriesVector(builder, len(dirOffsets))
	for i := len(dirOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(dirOffsets[i])
	}
	dirsVec := builder.EndVector(len(dirOffsets))

	rmfb.ManifestStart(builder)
	rmfb.ManifestAddBundles(builder, bundlesVec)
	rmfb.ManifestAddTags(builder, tagsVec)
	rmfb.ManifestAddFiles(builder, filesVec)
	rmfb.ManifestAddDirectories(builder, dirsVec)
	manifestOff := rmfb.ManifestEnd(builder)
	rmfb.FinishManifestBuffer(builder, manifestOff)

	return builder.FinishedBytes()
}

type testTag struct {
	id   byte
	name string
}
type testDir struct {
	id       uint64
	parentID uint64
	name     string
}
type testChunk struct {
	id               uint64
	compressedSize   uint32
	uncompressedSize uint32
}
type testBundle struct {
	id     uint64
	chunks []testChunk
}
type testFile struct {
	id         uint64
	dirID      uint64
	size       uint32
	name       string
	tagBitmask uint64
	chunkIDs   []uint64
	symlink    string
}
type testBodyOpts struct {
	tags    []testTag
	dirs    []testDir
	bundles []testBundle
	files   []testFile
}

// buildTestRMANData creates a complete RMAN binary with header + zstd-compressed FlatBuffer body.
func buildTestRMANData(t *testing.T, manifestID uint64, opts testBodyOpts) []byte {
	t.Helper()
	body := buildFlatBufferBody(t, opts)

	// Compress body with zstd
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatalf("creating zstd encoder: %v", err)
	}
	compressed := enc.EncodeAll(body, nil)
	enc.Close()

	// Build RMAN header (28 bytes minimum)
	// Magic: "RMAN"
	// Version: 2.0
	// Flags: 0
	// Content offset: 28
	// Compressed size: len(compressed)
	// Manifest ID: manifestID
	// Uncompressed size: len(body)
	header := make([]byte, 28)
	copy(header[0:4], "RMAN")
	header[4] = 2 // major version
	header[5] = 0 // minor version
	// header[6:8] = flags (0)
	binary.LittleEndian.PutUint32(header[8:12], 28)                // content offset
	binary.LittleEndian.PutUint32(header[12:16], uint32(len(compressed))) // compressed size
	binary.LittleEndian.PutUint64(header[16:24], manifestID)       // manifest ID
	binary.LittleEndian.PutUint32(header[24:28], uint32(len(body)))      // uncompressed size

	return append(header, compressed...)
}

func TestParseManifestData_Valid(t *testing.T) {
	data := buildTestRMANData(t, 0xDEADBEEF, testBodyOpts{
		tags: []testTag{
			{id: 0, name: "en_US"},
			{id: 1, name: "de_DE"},
		},
		dirs: []testDir{
			{id: 1, parentID: 0, name: "DATA"},
			{id: 2, parentID: 1, name: "Characters"},
		},
		bundles: []testBundle{
			{id: 0xABCD, chunks: []testChunk{
				{id: 100, compressedSize: 512, uncompressedSize: 1024},
				{id: 101, compressedSize: 256, uncompressedSize: 512},
			}},
		},
		files: []testFile{
			{id: 1, dirID: 2, size: 1536, name: "hero.bin", tagBitmask: 0b01, chunkIDs: []uint64{100, 101}},
			{id: 2, dirID: 1, size: 1024, name: "data.bin", tagBitmask: 0b11, chunkIDs: []uint64{100}},
		},
	})

	m, err := ParseManifestData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.ID != 0xDEADBEEF {
		t.Errorf("manifest ID = %x, want %x", m.ID, 0xDEADBEEF)
	}
	if m.Version.Major != 2 || m.Version.Minor != 0 {
		t.Errorf("version = %s, want 2.0", m.Version)
	}
	if len(m.Flags) != 2 {
		t.Fatalf("flags count = %d, want 2", len(m.Flags))
	}
	if m.Flags[0].Name != "en_US" {
		t.Errorf("flag[0].Name = %q, want %q", m.Flags[0].Name, "en_US")
	}

	if len(m.Files) != 2 {
		t.Fatalf("files count = %d, want 2", len(m.Files))
	}

	// Check directory path resolution
	heroFile := findFile(m.Files, "DATA/Characters/hero.bin")
	if heroFile == nil {
		t.Fatal("file DATA/Characters/hero.bin not found")
	}
	if heroFile.FileSize != 1536 {
		t.Errorf("hero.bin size = %d, want 1536", heroFile.FileSize)
	}
	if len(heroFile.Chunks) != 2 {
		t.Fatalf("hero.bin chunks = %d, want 2", len(heroFile.Chunks))
	}

	// Check flag assignment via bitmask
	if len(heroFile.Flags) != 1 {
		t.Errorf("hero.bin flags = %d, want 1 (only en_US)", len(heroFile.Flags))
	}

	dataFile := findFile(m.Files, "DATA/data.bin")
	if dataFile == nil {
		t.Fatal("file DATA/data.bin not found")
	}
	if len(dataFile.Flags) != 2 {
		t.Errorf("data.bin flags = %d, want 2 (both flags)", len(dataFile.Flags))
	}

	// Check chunk BundleOffset calculation
	if heroFile.Chunks[0].BundleOffset != 0 {
		t.Errorf("chunk[0] offset = %d, want 0", heroFile.Chunks[0].BundleOffset)
	}
	if heroFile.Chunks[1].BundleOffset != 512 {
		t.Errorf("chunk[1] offset = %d, want 512", heroFile.Chunks[1].BundleOffset)
	}
}

func TestParseManifestData_BadMagic(t *testing.T) {
	data := make([]byte, 28)
	copy(data[0:4], "XMAN")
	_, err := ParseManifestData(data)
	if err == nil {
		t.Fatal("expected error for bad magic bytes")
	}
}

func TestParseManifestData_TooShort(t *testing.T) {
	_, err := ParseManifestData([]byte("RM"))
	if err == nil {
		t.Fatal("expected error for too short data")
	}
}

func TestParseManifestData_UnsupportedVersion(t *testing.T) {
	data := make([]byte, 28)
	copy(data[0:4], "RMAN")
	data[4] = 3 // unsupported major version
	_, err := ParseManifestData(data)
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
}

func TestParseManifestData_DirectoryCycle(t *testing.T) {
	// Create directories that reference each other in a cycle
	data := buildTestRMANData(t, 1, testBodyOpts{
		dirs: []testDir{
			{id: 1, parentID: 2, name: "a"},
			{id: 2, parentID: 1, name: "b"},
		},
	})

	_, err := ParseManifestData(data)
	if err == nil {
		t.Fatal("expected error for directory cycle")
	}
}

func TestParseManifestData_DecompressedSizeMismatch(t *testing.T) {
	body := buildFlatBufferBody(t, testBodyOpts{})
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}
	compressed := enc.EncodeAll(body, nil)
	enc.Close()

	header := make([]byte, 28)
	copy(header[0:4], "RMAN")
	header[4] = 2
	binary.LittleEndian.PutUint32(header[8:12], 28)
	binary.LittleEndian.PutUint32(header[12:16], uint32(len(compressed)))
	binary.LittleEndian.PutUint64(header[16:24], 1)
	binary.LittleEndian.PutUint32(header[24:28], uint32(len(body)+999)) // wrong size

	data := append(header, compressed...)
	_, err = ParseManifestData(data)
	if err == nil {
		t.Fatal("expected error for decompressed size mismatch")
	}
}

func TestDecompress(t *testing.T) {
	original := []byte("hello world, this is test data for zstd compression")
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}
	compressed := enc.EncodeAll(original, nil)
	enc.Close()

	result, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("decompressed data mismatch")
	}
}

func TestDecompress_Invalid(t *testing.T) {
	_, err := Decompress([]byte("not zstd data"))
	if err == nil {
		t.Fatal("expected error for invalid data")
	}
}

func TestLoadFromFile(t *testing.T) {
	data := buildTestRMANData(t, 42, testBodyOpts{
		files: []testFile{
			{id: 1, size: 100, name: "test.txt"},
		},
		bundles: []testBundle{},
	})

	tmp := filepath.Join(t.TempDir(), "test.manifest")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := LoadFromFile(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != 42 {
		t.Errorf("ID = %d, want 42", m.ID)
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/manifest.bin")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFromURL(t *testing.T) {
	data := buildTestRMANData(t, 99, testBodyOpts{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer srv.Close()

	m, err := LoadFromURL(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != 99 {
		t.Errorf("ID = %d, want 99", m.ID)
	}
}

func TestLoadFromReader(t *testing.T) {
	data := buildTestRMANData(t, 77, testBodyOpts{})
	m, err := LoadFromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != 77 {
		t.Errorf("ID = %d, want 77", m.ID)
	}
}

func TestLoadFromReader_Error(t *testing.T) {
	_, err := LoadFromReader(&errorReader{})
	if err == nil {
		t.Fatal("expected error from bad reader")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func findFile(files []FileEntry, name string) *FileEntry {
	for i := range files {
		if files[i].Name == name {
			return &files[i]
		}
	}
	return nil
}
