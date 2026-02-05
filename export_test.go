package rman

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestExportJSON(t *testing.T) {
	m := &Manifest{
		ID:      0xABCD,
		Version: Version{Major: 2, Minor: 0},
		Flags:   []Flag{{FlagID: 0, Name: "en_US"}},
		Files: []FileEntry{
			{
				ID:       1,
				Name:     "test.bin",
				FileSize: 1024,
				Flags:    []Flag{{FlagID: 0, Name: "en_US"}},
				Chunks: []Chunk{
					{ChunkID: 100, BundleID: 0xAA, BundleOffset: 0, CompressedSize: 512, UncompressedSize: 1024},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := ExportJSON(&buf, m, false)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if result["version"] != "2.0" {
		t.Errorf("version = %v, want 2.0", result["version"])
	}
	files, ok := result["files"].([]interface{})
	if !ok || len(files) != 1 {
		t.Fatalf("expected 1 file in output")
	}
}

func TestExportJSON_Pretty(t *testing.T) {
	m := &Manifest{
		ID:      1,
		Version: Version{Major: 2, Minor: 0},
		Files:   []FileEntry{{ID: 1, Name: "a.txt", FileSize: 10}},
	}

	var buf bytes.Buffer
	err := ExportJSON(&buf, m, true)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("\n  ")) {
		t.Error("expected indented output for pretty=true")
	}
}

func TestExportFileList(t *testing.T) {
	m := &Manifest{
		Files: []FileEntry{
			{Name: "a.txt", FileSize: 100, Flags: []Flag{{Name: "en_US"}}},
			{Name: "b.txt", FileSize: 200},
		},
	}

	var buf bytes.Buffer
	err := ExportFileList(&buf, m, false)
	if err != nil {
		t.Fatal(err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0]["name"] != "a.txt" {
		t.Errorf("name = %v, want a.txt", result[0]["name"])
	}
}
