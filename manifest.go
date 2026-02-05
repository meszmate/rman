package rman

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/klauspost/compress/zstd"
	"github.com/meszmate/rman/flatbuffers"
)

// Version represents a manifest format version.
type Version struct {
	Major uint8
	Minor uint8
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// Manifest represents a parsed RMAN manifest.
type Manifest struct {
	ID      uint64
	Version Version
	Flags   []Flag
	Files   []FileEntry
}

// Chunk represents a compressed data chunk within a bundle.
type Chunk struct {
	CompressedSize   uint32
	UncompressedSize uint32
	ChunkID          uint64
	BundleOffset     uint32
	BundleID         uint64
}

// Flag represents a tag/flag with an ID and name.
type Flag struct {
	FlagID uint8
	Name   string
}

// FileEntry represents a file within the manifest.
type FileEntry struct {
	ID       uint64
	FileSize uint32
	Name     string
	SymLink  string
	Flags    []Flag
	Chunks   []Chunk
}

// LoadFromFile reads and parses a manifest from a local file.
func LoadFromFile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest file: %w", err)
	}
	return ParseManifestData(data)
}

// LoadFromURL fetches and parses a manifest from a URL.
func LoadFromURL(ctx context.Context, url string) (*Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return ParseManifestData(data)
}

// LoadFromReader reads and parses a manifest from an io.Reader.
func LoadFromReader(r io.Reader) (*Manifest, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading manifest data: %w", err)
	}
	return ParseManifestData(data)
}

// Decompress decompresses Zstandard-compressed data.
func Decompress(data []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("creating zstd decoder: %w", err)
	}
	defer decoder.Close()

	result, err := decoder.DecodeAll(data, nil)
	if err != nil {
		return nil, fmt.Errorf("zstd decompression: %w", err)
	}
	return result, nil
}

// ParseManifestData parses raw RMAN manifest bytes into a Manifest.
func ParseManifestData(data []byte) (*Manifest, error) {
	if len(data) < 28 {
		return nil, fmt.Errorf("data too short to contain necessary headers")
	}

	if string(data[:4]) != "RMAN" {
		return nil, fmt.Errorf("not a valid RMAN file: missing magic bytes")
	}

	ver := Version{
		Major: data[4],
		Minor: data[5],
	}

	if ver.Major != 2 {
		return nil, fmt.Errorf("unsupported manifest version %s", ver)
	}

	contentOffset := binary.LittleEndian.Uint32(data[8:12])
	compressedSize := binary.LittleEndian.Uint32(data[12:16])
	uncompressedSize := binary.LittleEndian.Uint32(data[24:28])

	if int(contentOffset+compressedSize) > len(data) {
		return nil, fmt.Errorf("compressed data exceeds manifest size")
	}

	compressedData := data[contentOffset : contentOffset+compressedSize]

	reader := bytes.NewReader(compressedData)
	decoder, err := zstd.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("zstd decompression failed: %w", err)
	}
	defer decoder.Close()

	uncompressedData, err := io.ReadAll(decoder)
	if err != nil {
		return nil, fmt.Errorf("reading decompressed data failed: %w", err)
	}

	if len(uncompressedData) != int(uncompressedSize) {
		return nil, fmt.Errorf("decompressed data size mismatch: expected %d, got %d", uncompressedSize, len(uncompressedData))
	}

	m := &Manifest{
		ID:      binary.LittleEndian.Uint64(data[16:24]),
		Version: ver,
	}
	if err := parseBody(m, uncompressedData); err != nil {
		return nil, fmt.Errorf("parsing body: %w", err)
	}

	return m, nil
}

// rawDir is an intermediate representation used during directory path resolution.
type rawDir struct {
	id       uint64
	parentID uint64
	name     string
}

// parseBody parses the FlatBuffers body into the manifest structure.
func parseBody(m *Manifest, data []byte) error {
	root := flatbuffers.GetRootAsManifest(data, 0)

	// Parse flags/tags
	flagsLen := root.TagsLength()
	m.Flags = make([]Flag, 0, flagsLen)
	for i := 0; i < flagsLen; i++ {
		tag := new(flatbuffers.Tag)
		root.Tags(tag, i)
		m.Flags = append(m.Flags, Flag{
			Name:   string(tag.Name()),
			FlagID: tag.Id(),
		})
	}

	// Parse directories using map-based O(n) resolution with cycle detection
	dirsLen := root.DirectoriesLength()
	dirMap := make(map[uint64]*rawDir, dirsLen)
	for i := 0; i < dirsLen; i++ {
		dir := new(flatbuffers.Directory)
		root.Directories(dir, i)
		if dir.Id() == 0 {
			continue
		}
		dirMap[dir.Id()] = &rawDir{
			id:       dir.Id(),
			parentID: dir.ParentId(),
			name:     string(dir.Name()),
		}
	}

	// Resolve full directory paths with memoization and cycle detection
	resolvedPaths := make(map[uint64]string, len(dirMap))
	var resolvePath func(id uint64, visited map[uint64]bool) (string, error)
	resolvePath = func(id uint64, visited map[uint64]bool) (string, error) {
		if path, ok := resolvedPaths[id]; ok {
			return path, nil
		}
		d, ok := dirMap[id]
		if !ok {
			return "", fmt.Errorf("directory ID %d not found", id)
		}
		if visited[id] {
			return "", fmt.Errorf("cycle detected in directory hierarchy at ID %d", id)
		}
		visited[id] = true

		if d.parentID == 0 {
			resolvedPaths[id] = d.name
			return d.name, nil
		}
		parentPath, err := resolvePath(d.parentID, visited)
		if err != nil {
			return "", err
		}
		fullPath := parentPath + "/" + d.name
		resolvedPaths[id] = fullPath
		return fullPath, nil
	}

	for id := range dirMap {
		if _, err := resolvePath(id, make(map[uint64]bool)); err != nil {
			return fmt.Errorf("resolving directory paths: %w", err)
		}
	}

	// Parse bundle chunks
	chunks := make(map[uint64]Chunk)
	bundleLen := root.BundlesLength()
	for i := 0; i < bundleLen; i++ {
		bundle := new(flatbuffers.Bundle)
		root.Bundles(bundle, i)

		chunksLen := bundle.ChunksLength()
		var offset uint32
		for j := 0; j < chunksLen; j++ {
			chunk := new(flatbuffers.Chunk)
			bundle.Chunks(chunk, j)

			c := Chunk{
				BundleID:         bundle.Id(),
				CompressedSize:   chunk.CompressedSize(),
				UncompressedSize: chunk.UncompressedSize(),
				BundleOffset:     offset,
				ChunkID:          chunk.Id(),
			}
			offset += c.CompressedSize
			chunks[c.ChunkID] = c
		}
	}

	// Parse files
	filesLen := root.FilesLength()
	m.Files = make([]FileEntry, 0, filesLen)
	for i := 0; i < filesLen; i++ {
		file := new(flatbuffers.File)
		root.Files(file, i)

		var entry FileEntry

		// Fix: iterate manifest.Flags (not the empty fileEntry.Flags)
		bitmask := file.TagBitmask()
		entry.Flags = make([]Flag, 0)
		for _, f := range m.Flags {
			if bitmask&(1<<f.FlagID) != 0 {
				entry.Flags = append(entry.Flags, f)
			}
		}

		entry.Name = string(file.Name())
		if dirPath, ok := resolvedPaths[file.DirectoryId()]; ok {
			entry.Name = dirPath + "/" + entry.Name
		}

		entry.SymLink = string(file.Symlink())
		entry.FileSize = file.Size()
		entry.ID = file.Id()

		chunkCount := file.ChunkIdsLength()
		entry.Chunks = make([]Chunk, 0, chunkCount)
		for j := 0; j < chunkCount; j++ {
			chunkID := file.ChunkIds(j)
			ch, ok := chunks[chunkID]
			if !ok {
				return fmt.Errorf("chunk ID %d not found for file %s", chunkID, entry.Name)
			}
			entry.Chunks = append(entry.Chunks, ch)
		}

		m.Files = append(m.Files, entry)
	}

	return nil
}
