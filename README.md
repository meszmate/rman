# rman

[![CI](https://github.com/meszmate/rman/actions/workflows/ci.yml/badge.svg)](https://github.com/meszmate/rman/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/meszmate/rman.svg)](https://pkg.go.dev/github.com/meszmate/rman)
[![Latest Release](https://img.shields.io/github/v/release/meszmate/rman)](https://github.com/meszmate/rman/releases/latest)

A Go library and CLI tool for parsing, inspecting, diffing, and downloading [Riot Games RMAN manifests](https://technology.riotgames.com/news/supercharging-data-delivery-new-league-patcher).

## Features

- Parse RMAN v2.x manifest files from local files, URLs, or readers
- Concurrent file downloader with progress reporting and retry support
- Manifest diffing (added/removed/modified files)
- File filtering by glob patterns, regex, flags, and size
- JSON export of manifests and file lists
- Built-in CDN URLs for known Riot games
- Interactive CLI with progress bars
- Resume interrupted downloads via state files
- Cross-platform binaries (Linux, macOS, Windows)

## Installation

### Library

```sh
go get github.com/meszmate/rman
```

### CLI (from source)

```sh
go install github.com/meszmate/rman/cmd/rman@latest
```

### CLI (binary releases)

Download pre-built binaries from the [Releases page](https://github.com/meszmate/rman/releases).

## CLI Quick Start

```sh
# Show manifest info
rman info ./EB9EF8EA7C032A8B.manifest

# List all files, sorted by size
rman list --sort size ./manifest.bin

# List only DLL files as JSON
rman list --filter "*.dll" --json ./manifest.bin

# Compare two manifest versions
rman diff old.manifest new.manifest

# Download with interactive game selection
rman download ./manifest.bin

# Download Valorant files non-interactively
rman download --game valorant --output ./valorant-files ./manifest.bin

# Download with a glob filter
rman download --game lol --output ./lol --filter "*.dll" ./manifest.bin

# Export full manifest as JSON
rman export ./manifest.bin > manifest.json

# Export simplified file list
rman export --files-only ./manifest.bin
```

## Library Usage

### Parsing a Manifest

Load from a local file, a URL, or any `io.Reader`:

```go
// From a file
m, err := rman.LoadFromFile("./EB9EF8EA7C032A8B.manifest")

// From a URL
m, err := rman.LoadFromURL(context.Background(), "https://example.com/manifest.bin")

// From any io.Reader
m, err := rman.LoadFromReader(reader)

// From raw bytes (lowest level)
m, err := rman.ParseManifestData(data)
```

The returned `*rman.Manifest` contains:

```go
type Manifest struct {
    ID      uint64      // Unique manifest identifier
    Version Version     // Format version (Major, Minor)
    Flags   []Flag      // Global flags/tags defined in the manifest
    Files   []FileEntry // All files described by the manifest
}
```

### Inspecting Files and Chunks

Each file contains its full path, size, flags, and the ordered list of chunks that make up its content:

```go
m, _ := rman.LoadFromFile("./manifest.bin")

for _, f := range m.Files {
    fmt.Printf("%s (%d bytes, %d chunks)\n", f.Name, f.FileSize, len(f.Chunks))

    // Each file has flags derived from the manifest's tag bitmask
    for _, flag := range f.Flags {
        fmt.Printf("  flag: [%d] %s\n", flag.FlagID, flag.Name)
    }

    // Chunks describe how the file is split across CDN bundles
    for _, c := range f.Chunks {
        fmt.Printf("  chunk %016X in bundle %016X (offset %d, %d -> %d bytes)\n",
            c.ChunkID, c.BundleID, c.BundleOffset, c.CompressedSize, c.UncompressedSize)
    }
}
```

### Decompressing Chunk Data

Chunks on the CDN are Zstandard-compressed. Use `Decompress` after fetching raw chunk bytes:

```go
decompressed, err := rman.Decompress(compressedChunkBytes)
if err != nil {
    log.Fatal(err)
}
// decompressed contains the original file data for this chunk
```

### Filtering Files

Filter by glob pattern, regex, flag names, and/or size. All criteria are ANDed together:

```go
// Glob pattern (matches against file basename and full path)
dlls, _ := rman.FilterFiles(m.Files, rman.Filter{Glob: "*.dll"})

// Regex
maps, _ := rman.FilterFiles(m.Files, rman.Filter{Regex: `Maps/map\d+\.bin`})

// Files that have specific flags
enUS, _ := rman.FilterFiles(m.Files, rman.Filter{FlagNames: []string{"en_US"}})

// Size range
large, _ := rman.FilterFiles(m.Files, rman.Filter{MinSize: 1024 * 1024}) // >= 1 MiB

// Combined: DLLs larger than 1 MiB with the en_US flag
filtered, _ := rman.FilterFiles(m.Files, rman.Filter{
    Glob:      "*.dll",
    MinSize:   1024 * 1024,
    FlagNames: []string{"en_US"},
})
```

### Diffing Two Manifests

Compare an old and new manifest to find added, removed, and modified files. Files are matched by path; a file is considered modified if its size or chunk composition changed:

```go
old, _ := rman.LoadFromFile("old.manifest")
new, _ := rman.LoadFromFile("new.manifest")

result := rman.Diff(old, new)

fmt.Printf("%d added, %d removed, %d modified\n",
    len(result.Added), len(result.Removed), len(result.Modified))

for _, f := range result.Added {
    fmt.Printf("  + %s (%d bytes)\n", f.Name, f.FileSize)
}
for _, f := range result.Removed {
    fmt.Printf("  - %s\n", f.Name)
}
for _, f := range result.Modified {
    fmt.Printf("  ~ %s (%d bytes)\n", f.Name, f.FileSize)
}
```

### Downloading Files

The concurrent downloader handles bundle fetching, decompression, retries, and writing files to disk:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

err := rman.Download(ctx, m, rman.DownloadConfig{
    CDNURL:    "https://valorant.dyn.riotcdn.net/channels/public/bundles",
    OutputDir: "./output",
    Workers:   8,       // concurrent download goroutines
    Retries:   3,       // retries per chunk on failure
    Client:    myHTTPClient, // optional custom *http.Client
    Progress: func(e rman.ProgressEvent) {
        pct := float64(e.BytesDone) / float64(e.BytesTotal) * 100
        fmt.Printf("\r%.1f%% (%d/%d files)", pct, e.FilesDone, e.FilesTotal)
    },
})
```

Cancel the context (e.g., on SIGINT) for graceful shutdown — workers will finish their current operation and exit.

### Looking Up Game CDN URLs

Use the built-in game registry instead of hardcoding CDN URLs:

```go
game, ok := rman.FindGame("valorant") // by slug
game, ok := rman.FindGame("League of Legends") // by name (case-insensitive)
if ok {
    fmt.Println(game.BundleURL) // https://valorant.dyn.riotcdn.net/channels/public/bundles
}

// Or iterate all known games
for _, g := range rman.KnownGames {
    fmt.Printf("%s (%s): %s\n", g.Name, g.Slug, g.BundleURL)
}
```

### Exporting as JSON

Export the full manifest (with all chunk details) or a simplified file list:

```go
// Full manifest JSON
rman.ExportJSON(os.Stdout, m, true) // true = pretty-print

// Simplified file list (name, size, flags only)
rman.ExportFileList(os.Stdout, m, true)
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `rman info <manifest>` | Show manifest metadata (ID, version, files, size, bundles, flags) |
| `rman list <manifest>` | List files with optional `--filter`, `--regex`, `--sort`, `--json` |
| `rman diff <old> <new>` | Show added/removed/modified files between two manifests |
| `rman download <manifest>` | Download files with interactive prompts or `--game`/`--cdn` flags |
| `rman export <manifest>` | Export manifest as JSON (`--files-only` for simplified output) |

## Download Flags

| Flag | Description |
|------|-------------|
| `-o, --output` | Output directory |
| `-g, --game` | Game slug (auto-sets CDN URL) |
| `-c, --cdn` | Custom CDN base URL |
| `-w, --workers` | Concurrent download workers (default: 8) |
| `-f, --filter` | Glob filter pattern |
| `-r, --regex` | Regex filter pattern |
| `--retries` | Max retries per chunk (default: 3) |
| `--state-file` | Custom state file path for resume |

## Resume Support

When a download is interrupted (Ctrl+C or error), the CLI saves progress to a state file (`.rman-state.json` in the output directory by default). Re-running the same download command will automatically detect and resume from where it left off.

## Supported Games

| Game | Slug | CDN |
|------|------|-----|
| League of Legends | `lol` | `lol.dyn.riotcdn.net` |
| Valorant | `valorant` | `valorant.dyn.riotcdn.net` |
| Legends of Runeterra | `lor` | `lor.dyn.riotcdn.net` |
| Teamfight Tactics | `tft` | `tft.dyn.riotcdn.net` |
| 2XKO | `2xko` | `2xko.dyn.riotcdn.net` |

## Background

RMAN is Riot Games' manifest format for their content delivery system. It describes game files as collections of deduplicated, compressed chunks stored in bundles on a CDN. For a deep dive into the format, see Riot's tech blog: [Supercharging Data Delivery: New League Patcher](https://technology.riotgames.com/news/supercharging-data-delivery-new-league-patcher).

## License

[MIT](LICENSE)
