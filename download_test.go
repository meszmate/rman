package rman

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestDownload_Basic(t *testing.T) {
	// Create test data: a chunk with known content
	content := []byte("hello world from chunk")
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}
	compressed := enc.EncodeAll(content, nil)
	enc.Close()

	bundleID := uint64(0xABCD)
	bundleData := compressed // single chunk at offset 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := fmt.Sprintf("/%016X.bundle", bundleID)
		if r.URL.Path != expected {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", 404)
			return
		}
		w.WriteHeader(http.StatusPartialContent)
		w.Write(bundleData)
	}))
	defer srv.Close()

	m := &Manifest{
		ID:      1,
		Version: Version{Major: 2},
		Files: []FileEntry{
			{
				ID:       1,
				Name:     "test/output.txt",
				FileSize: uint32(len(content)),
				Chunks: []Chunk{
					{
						ChunkID:          100,
						BundleID:         bundleID,
						BundleOffset:     0,
						CompressedSize:   uint32(len(compressed)),
						UncompressedSize: uint32(len(content)),
					},
				},
			},
		},
	}

	outputDir := t.TempDir()

	err = Download(context.Background(), m, DownloadConfig{
		CDNURL:    srv.URL,
		OutputDir: outputDir,
		Workers:   2,
		Retries:   1,
	})
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	// Verify file was written correctly
	got, err := os.ReadFile(filepath.Join(outputDir, "test", "output.txt"))
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", string(got), string(content))
	}
}

func TestDownload_MissingCDN(t *testing.T) {
	err := Download(context.Background(), &Manifest{}, DownloadConfig{
		OutputDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error for missing CDN URL")
	}
}

func TestDownload_MissingOutputDir(t *testing.T) {
	err := Download(context.Background(), &Manifest{}, DownloadConfig{
		CDNURL: "http://example.com",
	})
	if err == nil {
		t.Fatal("expected error for missing output dir")
	}
}

func TestDownload_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	m := &Manifest{
		Files: []FileEntry{
			{ID: 1, Name: "test.txt", Chunks: []Chunk{
				{ChunkID: 1, BundleID: 1, CompressedSize: 100, UncompressedSize: 200},
			}},
		},
	}

	err := Download(ctx, m, DownloadConfig{
		CDNURL:    "http://localhost:1",
		OutputDir: t.TempDir(),
		Workers:   1,
		Retries:   1,
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestDownload_Progress(t *testing.T) {
	content := []byte("progress test data")
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}
	compressed := enc.EncodeAll(content, nil)
	enc.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPartialContent)
		w.Write(compressed)
	}))
	defer srv.Close()

	m := &Manifest{
		Files: []FileEntry{
			{
				ID:       1,
				Name:     "test.txt",
				FileSize: uint32(len(content)),
				Chunks: []Chunk{
					{ChunkID: 1, BundleID: 1, CompressedSize: uint32(len(compressed)), UncompressedSize: uint32(len(content))},
				},
			},
		},
	}

	var progressCalled bool
	err = Download(context.Background(), m, DownloadConfig{
		CDNURL:    srv.URL,
		OutputDir: t.TempDir(),
		Workers:   1,
		Retries:   1,
		Progress: func(event ProgressEvent) {
			progressCalled = true
			if event.FilesTotal != 1 {
				t.Errorf("FilesTotal = %d, want 1", event.FilesTotal)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !progressCalled {
		t.Error("progress callback was not called")
	}
}

func TestDownload_Resume(t *testing.T) {
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create 3 files with distinct content
	fileContents := []string{"file-one-content", "file-two-content", "file-three-content"}
	fileCompressed := make([][]byte, 3)
	for i, c := range fileContents {
		fileCompressed[i] = enc.EncodeAll([]byte(c), nil)
	}
	enc.Close()

	bundleIDs := []uint64{0x1001, 0x1002, 0x1003}

	// Track how many HTTP requests are made per file
	var fetchCounts [3]atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i, bid := range bundleIDs {
			expected := fmt.Sprintf("/%016X.bundle", bid)
			if r.URL.Path == expected {
				fetchCounts[i].Add(1)
				w.WriteHeader(http.StatusPartialContent)
				w.Write(fileCompressed[i])
				return
			}
		}
		http.Error(w, "not found", 404)
	}))
	defer srv.Close()

	m := &Manifest{
		ID: 42,
		Files: []FileEntry{
			{
				ID: 1, Name: "a.txt", FileSize: uint32(len(fileContents[0])),
				Chunks: []Chunk{{ChunkID: 10, BundleID: bundleIDs[0], CompressedSize: uint32(len(fileCompressed[0])), UncompressedSize: uint32(len(fileContents[0]))}},
			},
			{
				ID: 2, Name: "b.txt", FileSize: uint32(len(fileContents[1])),
				Chunks: []Chunk{{ChunkID: 20, BundleID: bundleIDs[1], CompressedSize: uint32(len(fileCompressed[1])), UncompressedSize: uint32(len(fileContents[1]))}},
			},
			{
				ID: 3, Name: "c.txt", FileSize: uint32(len(fileContents[2])),
				Chunks: []Chunk{{ChunkID: 30, BundleID: bundleIDs[2], CompressedSize: uint32(len(fileCompressed[2])), UncompressedSize: uint32(len(fileContents[2]))}},
			},
		},
	}

	outputDir := t.TempDir()
	stateFile := filepath.Join(outputDir, ".rman-state.json")

	// First download: cancel after first file completes
	var filesDone atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	err = Download(ctx, m, DownloadConfig{
		CDNURL:    srv.URL,
		OutputDir: outputDir,
		Workers:   1, // single worker so files are processed in order
		Retries:   1,
		StateFile: stateFile,
		Progress: func(event ProgressEvent) {
			filesDone.Store(event.FilesDone)
			// Cancel after at least 1 file has completed
			if event.FilesDone >= 1 {
				cancel()
			}
		},
	})
	// Should return context error
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}

	// State file should exist on disk
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("state file should exist after interruption: %v", err)
	}

	// Reset fetch counts for second run
	for i := range fetchCounts {
		fetchCounts[i].Store(0)
	}

	// Second download: should resume and complete
	err = Download(context.Background(), m, DownloadConfig{
		CDNURL:    srv.URL,
		OutputDir: outputDir,
		Workers:   1,
		Retries:   1,
		StateFile: stateFile,
	})
	if err != nil {
		t.Fatalf("resume download failed: %v", err)
	}

	// The first file should NOT have been re-fetched
	if fetchCounts[0].Load() != 0 {
		t.Errorf("file a.txt was re-fetched %d times, expected 0 (should have been skipped)", fetchCounts[0].Load())
	}

	// All files should exist with correct content
	for i, name := range []string{"a.txt", "b.txt", "c.txt"} {
		got, err := os.ReadFile(filepath.Join(outputDir, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		if string(got) != fileContents[i] {
			t.Errorf("%s content = %q, want %q", name, string(got), fileContents[i])
		}
	}

	// State file should be cleaned up after successful completion
	if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
		t.Error("state file should be removed after successful download")
	}
}

func TestDownload_StateFileMismatch(t *testing.T) {
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("test content")
	compressed := enc.EncodeAll(content, nil)
	enc.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPartialContent)
		w.Write(compressed)
	}))
	defer srv.Close()

	outputDir := t.TempDir()
	stateFile := filepath.Join(outputDir, ".rman-state.json")

	m1 := &Manifest{
		ID: 100,
		Files: []FileEntry{
			{ID: 1, Name: "f.txt", FileSize: uint32(len(content)),
				Chunks: []Chunk{{ChunkID: 1, BundleID: 1, CompressedSize: uint32(len(compressed)), UncompressedSize: uint32(len(content))}}},
		},
	}

	// Download with manifest ID 100
	err = Download(context.Background(), m1, DownloadConfig{
		CDNURL: srv.URL, OutputDir: outputDir, Workers: 1, Retries: 1, StateFile: stateFile,
	})
	if err != nil {
		t.Fatal(err)
	}

	// State file removed on success, write a fake one with different manifest ID
	os.WriteFile(stateFile, []byte(`{"manifest_id":999,"cdn_url":"`+srv.URL+`","output_dir":"`+outputDir+`","started_at":"2024-01-01T00:00:00Z","files":{"f.txt":{"total_chunks":1,"done_chunks":[1],"complete":true}}}`), 0o644)

	// Download with manifest ID 200 — old state should be discarded
	var fetchCount atomic.Int64
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount.Add(1)
		w.WriteHeader(http.StatusPartialContent)
		w.Write(compressed)
	}))
	defer srv2.Close()

	m2 := &Manifest{
		ID: 200,
		Files: []FileEntry{
			{ID: 1, Name: "f.txt", FileSize: uint32(len(content)),
				Chunks: []Chunk{{ChunkID: 1, BundleID: 1, CompressedSize: uint32(len(compressed)), UncompressedSize: uint32(len(content))}}},
		},
	}

	err = Download(context.Background(), m2, DownloadConfig{
		CDNURL: srv2.URL, OutputDir: outputDir, Workers: 1, Retries: 1, StateFile: stateFile,
	})
	if err != nil {
		t.Fatal(err)
	}

	// File should have been fetched since state was discarded
	if fetchCount.Load() == 0 {
		t.Error("expected file to be re-fetched when state manifest ID doesn't match")
	}
}
