package rman

import "testing"

func TestDiff_Added(t *testing.T) {
	old := &Manifest{Files: []FileEntry{
		{Name: "a.txt", FileSize: 100},
	}}
	new := &Manifest{Files: []FileEntry{
		{Name: "a.txt", FileSize: 100},
		{Name: "b.txt", FileSize: 200},
	}}

	result := Diff(old, new)
	if len(result.Added) != 1 || result.Added[0].Name != "b.txt" {
		t.Errorf("expected 1 added file (b.txt), got %v", result.Added)
	}
	if len(result.Removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(result.Removed))
	}
	if len(result.Modified) != 0 {
		t.Errorf("expected 0 modified, got %d", len(result.Modified))
	}
}

func TestDiff_Removed(t *testing.T) {
	old := &Manifest{Files: []FileEntry{
		{Name: "a.txt", FileSize: 100},
		{Name: "b.txt", FileSize: 200},
	}}
	new := &Manifest{Files: []FileEntry{
		{Name: "a.txt", FileSize: 100},
	}}

	result := Diff(old, new)
	if len(result.Removed) != 1 || result.Removed[0].Name != "b.txt" {
		t.Errorf("expected 1 removed file (b.txt), got %v", result.Removed)
	}
}

func TestDiff_Modified_Size(t *testing.T) {
	old := &Manifest{Files: []FileEntry{
		{Name: "a.txt", FileSize: 100},
	}}
	new := &Manifest{Files: []FileEntry{
		{Name: "a.txt", FileSize: 200},
	}}

	result := Diff(old, new)
	if len(result.Modified) != 1 || result.Modified[0].Name != "a.txt" {
		t.Errorf("expected 1 modified file, got %v", result.Modified)
	}
}

func TestDiff_Modified_Chunks(t *testing.T) {
	old := &Manifest{Files: []FileEntry{
		{Name: "a.txt", FileSize: 100, Chunks: []Chunk{{ChunkID: 1}}},
	}}
	new := &Manifest{Files: []FileEntry{
		{Name: "a.txt", FileSize: 100, Chunks: []Chunk{{ChunkID: 2}}},
	}}

	result := Diff(old, new)
	if len(result.Modified) != 1 {
		t.Errorf("expected 1 modified (chunk change), got %d", len(result.Modified))
	}
}

func TestDiff_NoChanges(t *testing.T) {
	files := []FileEntry{
		{Name: "a.txt", FileSize: 100, Chunks: []Chunk{{ChunkID: 1}}},
	}
	old := &Manifest{Files: files}
	new := &Manifest{Files: files}

	result := Diff(old, new)
	if len(result.Added) != 0 || len(result.Removed) != 0 || len(result.Modified) != 0 {
		t.Error("expected no differences")
	}
}

func TestDiff_Empty(t *testing.T) {
	result := Diff(&Manifest{}, &Manifest{})
	if len(result.Added) != 0 || len(result.Removed) != 0 || len(result.Modified) != 0 {
		t.Error("expected no differences for empty manifests")
	}
}
