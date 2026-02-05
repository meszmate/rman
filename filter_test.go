package rman

import "testing"

var testFiles = []FileEntry{
	{ID: 1, Name: "DATA/Maps/map1.bin", FileSize: 1000, Flags: []Flag{{FlagID: 0, Name: "en_US"}}},
	{ID: 2, Name: "DATA/Maps/map2.bin", FileSize: 2000, Flags: []Flag{{FlagID: 1, Name: "de_DE"}}},
	{ID: 3, Name: "DATA/Characters/hero.dll", FileSize: 500, Flags: []Flag{{FlagID: 0, Name: "en_US"}, {FlagID: 1, Name: "de_DE"}}},
	{ID: 4, Name: "Engine/core.dll", FileSize: 5000},
	{ID: 5, Name: "readme.txt", FileSize: 100},
}

func TestFilterFiles_Glob(t *testing.T) {
	result, err := FilterFiles(testFiles, Filter{Glob: "*.dll"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d files, want 2", len(result))
	}
}

func TestFilterFiles_Regex(t *testing.T) {
	result, err := FilterFiles(testFiles, Filter{Regex: `Maps/map\d+\.bin`})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d files, want 2", len(result))
	}
}

func TestFilterFiles_FlagNames(t *testing.T) {
	result, err := FilterFiles(testFiles, Filter{FlagNames: []string{"en_US"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d files, want 2 (map1 and hero)", len(result))
	}
}

func TestFilterFiles_FlagNames_Multiple(t *testing.T) {
	result, err := FilterFiles(testFiles, Filter{FlagNames: []string{"en_US", "de_DE"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d files, want 1 (only hero has both flags)", len(result))
	}
	if result[0].Name != "DATA/Characters/hero.dll" {
		t.Errorf("expected hero.dll, got %s", result[0].Name)
	}
}

func TestFilterFiles_MinSize(t *testing.T) {
	result, err := FilterFiles(testFiles, Filter{MinSize: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 {
		t.Fatalf("got %d files, want 3", len(result))
	}
}

func TestFilterFiles_MaxSize(t *testing.T) {
	result, err := FilterFiles(testFiles, Filter{MaxSize: 500})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d files, want 2 (hero and readme)", len(result))
	}
}

func TestFilterFiles_Combined(t *testing.T) {
	result, err := FilterFiles(testFiles, Filter{
		Glob:    "*.dll",
		MinSize: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d files, want 1 (only core.dll)", len(result))
	}
	if result[0].Name != "Engine/core.dll" {
		t.Errorf("expected core.dll, got %s", result[0].Name)
	}
}

func TestFilterFiles_InvalidRegex(t *testing.T) {
	_, err := FilterFiles(testFiles, Filter{Regex: "[invalid"})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestFilterFiles_EmptyFilter(t *testing.T) {
	result, err := FilterFiles(testFiles, Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != len(testFiles) {
		t.Fatalf("got %d files, want %d (no filter = all files)", len(result), len(testFiles))
	}
}
