package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/meszmate/rman"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <manifest>",
	Short: "Download files from a manifest",
	Long:  "Download game files from an RMAN manifest. Supports interactive mode, filtering, and resume.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDownload,
}

var (
	dlOutput    string
	dlCDN       string
	dlGame      string
	dlWorkers   int
	dlFilter    string
	dlRegex     string
	dlRetries   int
	dlStateFile string
)

func init() {
	downloadCmd.Flags().StringVarP(&dlOutput, "output", "o", "", "Output directory")
	downloadCmd.Flags().StringVarP(&dlCDN, "cdn", "c", "", "CDN base URL")
	downloadCmd.Flags().StringVarP(&dlGame, "game", "g", "", "Game slug (auto-sets CDN)")
	downloadCmd.Flags().IntVarP(&dlWorkers, "workers", "w", 8, "Concurrent workers")
	downloadCmd.Flags().StringVarP(&dlFilter, "filter", "f", "", "Glob filter pattern")
	downloadCmd.Flags().StringVarP(&dlRegex, "regex", "r", "", "Regex filter pattern")
	downloadCmd.Flags().IntVar(&dlRetries, "retries", 3, "Max retries per chunk")
	downloadCmd.Flags().StringVar(&dlStateFile, "state-file", "", "State file path for resume")
}

func runDownload(cmd *cobra.Command, args []string) error {
	m, err := loadManifest(args[0])
	if err != nil {
		return err
	}

	// Resolve CDN URL
	cdnURL := dlCDN
	if dlGame != "" {
		game, ok := rman.FindGame(dlGame)
		if !ok {
			return fmt.Errorf("unknown game: %s", dlGame)
		}
		cdnURL = game.BundleURL
	}

	// Interactive mode: prompt for CDN if not provided
	if cdnURL == "" {
		cdnURL, err = promptCDN()
		if err != nil {
			return err
		}
	}

	// Interactive mode: prompt for output dir if not provided
	outputDir := dlOutput
	if outputDir == "" {
		outputDir, err = promptString("Output directory", "./output")
		if err != nil {
			return err
		}
	}

	// Apply filters
	files := m.Files
	if dlFilter != "" || dlRegex != "" {
		files, err = rman.FilterFiles(files, rman.Filter{
			Glob:  dlFilter,
			Regex: dlRegex,
		})
		if err != nil {
			return fmt.Errorf("filter error: %w", err)
		}
	}

	fmt.Printf("Manifest: %016X (v%s)\n", m.ID, m.Version)
	fmt.Printf("Files: %d\n", len(files))
	fmt.Printf("CDN: %s\n", cdnURL)
	fmt.Printf("Output: %s\n", outputDir)
	fmt.Printf("Workers: %d\n", dlWorkers)
	fmt.Println()

	// Create filtered manifest for download
	filtered := &rman.Manifest{
		ID:      m.ID,
		Version: m.Version,
		Flags:   m.Flags,
		Files:   files,
	}

	// Calculate total size
	var totalSize int64
	for _, f := range files {
		totalSize += int64(f.FileSize)
	}

	// Set up progress bar
	bar := progressbar.NewOptions64(totalSize,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(0),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(os.Stderr)
		}),
	)

	ctx, cancel := signalContext()
	defer cancel()

	err = rman.Download(ctx, filtered, rman.DownloadConfig{
		CDNURL:    cdnURL,
		OutputDir: outputDir,
		Workers:   dlWorkers,
		Retries:   dlRetries,
		StateFile: dlStateFile,
		Progress: func(event rman.ProgressEvent) {
			bar.Set64(event.BytesDone)
		},
	})

	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	bar.Finish()
	fmt.Fprintf(os.Stderr, "Download complete! %d files written to %s\n", len(files), outputDir)
	return nil
}

func promptCDN() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Select a game or enter a custom CDN URL:")
	for i, g := range rman.KnownGames {
		fmt.Printf("  [%d] %s (%s)\n", i+1, g.Name, g.Slug)
	}
	fmt.Printf("  [%d] Custom URL\n", len(rman.KnownGames)+1)
	fmt.Print("\nChoice: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(rman.KnownGames)+1 {
		return "", fmt.Errorf("invalid choice: %s", input)
	}

	if choice <= len(rman.KnownGames) {
		game := rman.KnownGames[choice-1]
		fmt.Printf("Selected: %s (%s)\n", game.Name, game.BundleURL)
		return game.BundleURL, nil
	}

	fmt.Print("Enter CDN URL: ")
	url, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(url), nil
}

func promptString(prompt, defaultVal string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}
