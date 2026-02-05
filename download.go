package rman

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/meszmate/rman/internal/state"
)

// ProgressEvent contains information about download progress.
type ProgressEvent struct {
	BytesDone  int64
	BytesTotal int64
	FilesDone  int64
	FilesTotal int64
}

// ProgressFunc is called with progress updates during download.
type ProgressFunc func(event ProgressEvent)

// DownloadConfig configures the concurrent downloader.
type DownloadConfig struct {
	// CDNURL is the base URL for bundle downloads (e.g., "https://valorant.dyn.riotcdn.net/channels/public/bundles")
	CDNURL string
	// OutputDir is the directory to write downloaded files to
	OutputDir string
	// Workers is the number of concurrent download workers (default: 8)
	Workers int
	// Retries is the number of retry attempts per bundle fetch (default: 3)
	Retries int
	// Client is the HTTP client to use (default: http.DefaultClient)
	Client *http.Client
	// Progress is called with progress updates (optional)
	Progress ProgressFunc
	// StateFile is the path to the state file for resume support (optional).
	// If empty, defaults to ".rman-state.json" in OutputDir.
	StateFile string
}

func (c *DownloadConfig) defaults() {
	if c.Workers <= 0 {
		c.Workers = 8
	}
	if c.Retries <= 0 {
		c.Retries = 3
	}
	if c.Client == nil {
		c.Client = http.DefaultClient
	}
	if c.StateFile == "" && c.OutputDir != "" {
		c.StateFile = filepath.Join(c.OutputDir, ".rman-state.json")
	}
}

// Download downloads all files from the manifest using the given configuration.
// It uses a bundle-oriented strategy: fetching chunk data via HTTP Range requests.
// Supports context-based cancellation for graceful shutdown.
// If a StateFile is configured, download progress is persisted so interrupted
// downloads can be resumed.
func Download(ctx context.Context, m *Manifest, cfg DownloadConfig) error {
	cfg.defaults()

	if cfg.CDNURL == "" {
		return fmt.Errorf("CDN URL is required")
	}
	if cfg.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Load or create download state
	ds := loadOrCreateState(cfg.StateFile, m.ID, cfg.CDNURL, cfg.OutputDir)

	// Calculate totals
	var totalBytes int64
	for _, f := range m.Files {
		totalBytes += int64(f.FileSize)
	}

	var bytesDone atomic.Int64
	var filesDone atomic.Int64
	filesTotal := int64(len(m.Files))

	// Account for already-completed files in progress
	for _, f := range m.Files {
		if ds.IsFileComplete(f.Name) {
			bytesDone.Add(int64(f.FileSize))
			filesDone.Add(1)
		}
	}

	reportProgress := func() {
		if cfg.Progress != nil {
			cfg.Progress(ProgressEvent{
				BytesDone:  bytesDone.Load(),
				BytesTotal: totalBytes,
				FilesDone:  filesDone.Load(),
				FilesTotal: filesTotal,
			})
		}
	}

	// Group chunks by file for ordered writing
	type fileWork struct {
		file   FileEntry
		chunks []Chunk
	}

	fileCh := make(chan fileWork, cfg.Workers)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	// Start workers that process files
	for range cfg.Workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fw := range fileCh {
				if ctx.Err() != nil {
					return
				}
				if err := downloadFile(ctx, cfg, fw.file, ds, &bytesDone); err != nil {
					select {
					case errCh <- fmt.Errorf("downloading %s: %w", fw.file.Name, err):
					default:
					}
					return
				}
				ds.MarkFileComplete(fw.file.Name)
				filesDone.Add(1)
				ds.Save()
				reportProgress()
			}
		}()
	}

	// Feed files to workers, skipping completed files
	go func() {
		for _, f := range m.Files {
			if ctx.Err() != nil {
				break
			}
			if ds.IsFileComplete(f.Name) {
				continue
			}
			fileCh <- fileWork{file: f, chunks: f.Chunks}
		}
		close(fileCh)
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		ds.Save()
		return err
	default:
	}

	if ctx.Err() != nil {
		ds.Save()
		return ctx.Err()
	}

	// All done — clean up state file
	ds.Remove()
	return nil
}

// loadOrCreateState loads an existing state file if it matches the current
// download parameters, otherwise creates a new one.
func loadOrCreateState(path string, manifestID uint64, cdnURL, outputDir string) *state.DownloadState {
	if path == "" {
		return state.New("", manifestID, cdnURL, outputDir)
	}
	ds, err := state.Load(path)
	if err == nil && ds.ManifestID == manifestID && ds.CDNURL == cdnURL && ds.OutputDir == outputDir {
		return ds
	}
	return state.New(path, manifestID, cdnURL, outputDir)
}

func downloadFile(ctx context.Context, cfg DownloadConfig, file FileEntry, ds *state.DownloadState, bytesDone *atomic.Int64) error {
	fpath := filepath.Join(cfg.OutputDir, file.Name)
	if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	f, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	totalChunks := len(file.Chunks)
	for _, chunk := range file.Chunks {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if ds.IsChunkDone(file.Name, chunk.ChunkID) {
			bytesDone.Add(int64(chunk.UncompressedSize))
			continue
		}

		data, err := fetchChunk(ctx, cfg, chunk)
		if err != nil {
			return fmt.Errorf("fetching chunk %d: %w", chunk.ChunkID, err)
		}

		decompressed, err := Decompress(data)
		if err != nil {
			return fmt.Errorf("decompressing chunk %d: %w", chunk.ChunkID, err)
		}

		if _, err := f.Write(decompressed); err != nil {
			return fmt.Errorf("writing chunk: %w", err)
		}

		bytesDone.Add(int64(chunk.UncompressedSize))
		ds.MarkChunkDone(file.Name, chunk.ChunkID, totalChunks)
	}

	return nil
}

func fetchChunk(ctx context.Context, cfg DownloadConfig, chunk Chunk) ([]byte, error) {
	url := fmt.Sprintf("%s/%016X.bundle", cfg.CDNURL, chunk.BundleID)
	rangeEnd := chunk.BundleOffset + chunk.CompressedSize - 1

	var lastErr error
	for attempt := range cfg.Retries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", chunk.BundleOffset, rangeEnd))

		resp, err := cfg.Client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status %d (attempt %d/%d)", resp.StatusCode, attempt+1, cfg.Retries)
			continue
		}

		return data, nil
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", cfg.Retries, lastErr)
}
