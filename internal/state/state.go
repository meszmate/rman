package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DownloadState tracks the progress of a manifest download for resume support.
type DownloadState struct {
	ManifestID uint64                `json:"manifest_id"`
	CDNURL     string                `json:"cdn_url"`
	OutputDir  string                `json:"output_dir"`
	StartedAt  time.Time             `json:"started_at"`
	Files      map[string]*FileState `json:"files"`

	mu   sync.Mutex
	path string
}

// FileState tracks download progress for a single file.
type FileState struct {
	TotalChunks int      `json:"total_chunks"`
	DoneChunks  []uint64 `json:"done_chunks"`
	Complete    bool     `json:"complete"`
}

// New creates a new DownloadState.
func New(path string, manifestID uint64, cdnURL, outputDir string) *DownloadState {
	return &DownloadState{
		ManifestID: manifestID,
		CDNURL:     cdnURL,
		OutputDir:  outputDir,
		StartedAt:  time.Now(),
		Files:      make(map[string]*FileState),
		path:       path,
	}
}

// Load reads a DownloadState from disk.
func Load(path string) (*DownloadState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}
	var s DownloadState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}
	s.path = path
	return &s, nil
}

// Save writes the state to disk atomically (write to temp + rename).
func (s *DownloadState) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *DownloadState) saveLocked() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing temp state file: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("renaming state file: %w", err)
	}
	return nil
}

// MarkChunkDone records that a chunk has been downloaded for a file.
func (s *DownloadState) MarkChunkDone(fileName string, chunkID uint64, totalChunks int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fs, ok := s.Files[fileName]
	if !ok {
		fs = &FileState{TotalChunks: totalChunks}
		s.Files[fileName] = fs
	}
	fs.DoneChunks = append(fs.DoneChunks, chunkID)
	if len(fs.DoneChunks) >= fs.TotalChunks {
		fs.Complete = true
	}
}

// MarkFileComplete marks a file as fully downloaded.
func (s *DownloadState) MarkFileComplete(fileName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fs, ok := s.Files[fileName]; ok {
		fs.Complete = true
	}
}

// IsFileComplete returns true if the file has been fully downloaded.
func (s *DownloadState) IsFileComplete(fileName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fs, ok := s.Files[fileName]; ok {
		return fs.Complete
	}
	return false
}

// IsChunkDone returns true if the chunk has already been downloaded for a file.
func (s *DownloadState) IsChunkDone(fileName string, chunkID uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	fs, ok := s.Files[fileName]
	if !ok {
		return false
	}
	for _, id := range fs.DoneChunks {
		if id == chunkID {
			return true
		}
	}
	return false
}

// Path returns the file path of the state file.
func (s *DownloadState) Path() string {
	return s.path
}

// Remove deletes the state file from disk.
func (s *DownloadState) Remove() error {
	return os.Remove(s.path)
}
