package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := New(path, 0xDEAD, "http://cdn.example.com", "/output")
	if err := s.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("state file not created: %v", err)
	}
}

func TestLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := New(path, 0xBEEF, "http://cdn.example.com", "/output")
	s.MarkChunkDone("file1.txt", 100, 3)
	s.MarkChunkDone("file1.txt", 101, 3)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.ManifestID != 0xBEEF {
		t.Errorf("ManifestID = %x, want %x", loaded.ManifestID, 0xBEEF)
	}
	if loaded.CDNURL != "http://cdn.example.com" {
		t.Errorf("CDNURL = %q", loaded.CDNURL)
	}

	fs := loaded.Files["file1.txt"]
	if fs == nil {
		t.Fatal("file1.txt state not found")
	}
	if len(fs.DoneChunks) != 2 {
		t.Errorf("DoneChunks = %d, want 2", len(fs.DoneChunks))
	}
	if fs.Complete {
		t.Error("should not be complete yet (2/3 chunks)")
	}
}

func TestMarkChunkDone_Complete(t *testing.T) {
	s := New("", 1, "", "")
	s.MarkChunkDone("f.txt", 1, 2)
	s.MarkChunkDone("f.txt", 2, 2)

	if !s.IsFileComplete("f.txt") {
		t.Error("expected file to be complete")
	}
}

func TestIsChunkDone(t *testing.T) {
	s := New("", 1, "", "")
	if s.IsChunkDone("f.txt", 1) {
		t.Error("chunk should not be done yet")
	}

	s.MarkChunkDone("f.txt", 1, 2)
	if !s.IsChunkDone("f.txt", 1) {
		t.Error("chunk should be done")
	}
	if s.IsChunkDone("f.txt", 2) {
		t.Error("chunk 2 should not be done")
	}
}

func TestMarkFileComplete(t *testing.T) {
	s := New("", 1, "", "")
	s.MarkChunkDone("f.txt", 1, 5)
	s.MarkFileComplete("f.txt")

	if !s.IsFileComplete("f.txt") {
		t.Error("expected file to be complete")
	}
}

func TestIsFileComplete_NotTracked(t *testing.T) {
	s := New("", 1, "", "")
	if s.IsFileComplete("unknown.txt") {
		t.Error("unknown file should not be complete")
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("/nonexistent/state.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	os.WriteFile(path, []byte("not json"), 0o644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := New(path, 1, "", "")
	s.Save()

	if err := s.Remove(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("state file should be removed")
	}
}
