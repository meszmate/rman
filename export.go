package rman

import (
	"encoding/json"
	"io"
)

type exportManifest struct {
	ID      uint64            `json:"id"`
	Version string            `json:"version"`
	Flags   []exportFlag      `json:"flags"`
	Files   []exportFileEntry `json:"files"`
}

type exportFlag struct {
	ID   uint8  `json:"id"`
	Name string `json:"name"`
}

type exportFileEntry struct {
	ID       uint64        `json:"id"`
	Name     string        `json:"name"`
	Size     uint32        `json:"size"`
	SymLink  string        `json:"symlink,omitempty"`
	Flags    []string      `json:"flags,omitempty"`
	Chunks   []exportChunk `json:"chunks"`
}

type exportChunk struct {
	ChunkID          uint64 `json:"chunk_id"`
	BundleID         uint64 `json:"bundle_id"`
	BundleOffset     uint32 `json:"bundle_offset"`
	CompressedSize   uint32 `json:"compressed_size"`
	UncompressedSize uint32 `json:"uncompressed_size"`
}

type exportFileListEntry struct {
	Name  string   `json:"name"`
	Size  uint32   `json:"size"`
	Flags []string `json:"flags,omitempty"`
}

// ExportJSON writes the full manifest as JSON to w.
func ExportJSON(w io.Writer, m *Manifest, pretty bool) error {
	em := exportManifest{
		ID:      m.ID,
		Version: m.Version.String(),
		Flags:   make([]exportFlag, len(m.Flags)),
		Files:   make([]exportFileEntry, len(m.Files)),
	}

	for i, f := range m.Flags {
		em.Flags[i] = exportFlag{ID: f.FlagID, Name: f.Name}
	}

	for i, f := range m.Files {
		flags := make([]string, len(f.Flags))
		for j, fl := range f.Flags {
			flags[j] = fl.Name
		}
		chunks := make([]exportChunk, len(f.Chunks))
		for j, c := range f.Chunks {
			chunks[j] = exportChunk{
				ChunkID:          c.ChunkID,
				BundleID:         c.BundleID,
				BundleOffset:     c.BundleOffset,
				CompressedSize:   c.CompressedSize,
				UncompressedSize: c.UncompressedSize,
			}
		}
		em.Files[i] = exportFileEntry{
			ID:     f.ID,
			Name:   f.Name,
			Size:   f.FileSize,
			SymLink: f.SymLink,
			Flags:  flags,
			Chunks: chunks,
		}
	}

	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(em)
}

// ExportFileList writes a simplified file list as JSON to w.
func ExportFileList(w io.Writer, m *Manifest, pretty bool) error {
	entries := make([]exportFileListEntry, len(m.Files))
	for i, f := range m.Files {
		flags := make([]string, len(f.Flags))
		for j, fl := range f.Flags {
			flags[j] = fl.Name
		}
		entries[i] = exportFileListEntry{
			Name:  f.Name,
			Size:  f.FileSize,
			Flags: flags,
		}
	}

	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(entries)
}
