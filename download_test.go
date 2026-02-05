package rman

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
