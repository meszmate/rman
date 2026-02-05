package rman

import "strings"

// Game represents a known Riot Games title with its CDN bundle URL.
type Game struct {
	Name      string
	Slug      string
	BundleURL string
}

var (
	LeagueOfLegends    = Game{Name: "League of Legends", Slug: "lol", BundleURL: "https://lol.dyn.riotcdn.net/channels/public/bundles"}
	Valorant           = Game{Name: "Valorant", Slug: "valorant", BundleURL: "https://valorant.dyn.riotcdn.net/channels/public/bundles"}
	LegendsOfRuneterra = Game{Name: "Legends of Runeterra", Slug: "lor", BundleURL: "https://lor.dyn.riotcdn.net/channels/public/bundles"}
	TeamfightTactics   = Game{Name: "Teamfight Tactics", Slug: "tft", BundleURL: "https://tft.dyn.riotcdn.net/channels/public/bundles"}
	TwoXKO             = Game{Name: "2XKO", Slug: "2xko", BundleURL: "https://2xko.dyn.riotcdn.net/channels/public/bundles"}
)

// KnownGames is the list of known Riot Games titles with their CDN URLs.
var KnownGames = []Game{
	LeagueOfLegends,
	Valorant,
	LegendsOfRuneterra,
	TeamfightTactics,
	TwoXKO,
}

// FindGame looks up a game by slug or name (case-insensitive).
// Returns the game and true if found, or nil and false otherwise.
func FindGame(query string) (*Game, bool) {
	q := strings.ToLower(query)
	for i := range KnownGames {
		if strings.ToLower(KnownGames[i].Slug) == q || strings.ToLower(KnownGames[i].Name) == q {
			return &KnownGames[i], true
		}
	}
	return nil, false
}
