package rman

import "testing"

func TestFindGame_BySlug(t *testing.T) {
	game, ok := FindGame("valorant")
	if !ok {
		t.Fatal("expected to find valorant")
	}
	if game.Name != "Valorant" {
		t.Errorf("name = %q, want Valorant", game.Name)
	}
	if game.BundleURL == "" {
		t.Error("BundleURL should not be empty")
	}
}

func TestFindGame_ByName(t *testing.T) {
	game, ok := FindGame("League of Legends")
	if !ok {
		t.Fatal("expected to find League of Legends")
	}
	if game.Slug != "lol" {
		t.Errorf("slug = %q, want lol", game.Slug)
	}
}

func TestFindGame_CaseInsensitive(t *testing.T) {
	game, ok := FindGame("VALORANT")
	if !ok {
		t.Fatal("expected case-insensitive match")
	}
	if game.Slug != "valorant" {
		t.Errorf("slug = %q, want valorant", game.Slug)
	}
}

func TestFindGame_NotFound(t *testing.T) {
	_, ok := FindGame("nonexistent-game")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestFindGame_2XKO(t *testing.T) {
	game, ok := FindGame("2xko")
	if !ok {
		t.Fatal("expected to find 2xko")
	}
	if game.Name != "2XKO" {
		t.Errorf("name = %q, want 2XKO", game.Name)
	}
}

func TestKnownGames_AllHaveURLs(t *testing.T) {
	for _, g := range KnownGames {
		if g.BundleURL == "" {
			t.Errorf("game %q has empty BundleURL", g.Name)
		}
		if g.Slug == "" {
			t.Errorf("game %q has empty Slug", g.Name)
		}
	}
}
