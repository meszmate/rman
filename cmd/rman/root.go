package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/meszmate/rman"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "rman",
	Short:   "RMAN - Riot Games Manifest Tool",
	Long:    "A CLI tool for parsing, inspecting, and downloading Riot Games RMAN manifests.",
	Version: version,
}

func init() {
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(downloadCmd)
}

// loadManifest loads a manifest from a file path or URL.
func loadManifest(pathOrURL string) (*rman.Manifest, error) {
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		return rman.LoadFromURL(context.Background(), pathOrURL)
	}

	if _, err := os.Stat(pathOrURL); err != nil {
		return nil, fmt.Errorf("manifest not found: %s", pathOrURL)
	}
	return rman.LoadFromFile(pathOrURL)
}

// signalContext returns a context that is cancelled on SIGINT/SIGTERM.
func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			fmt.Fprintln(os.Stderr, "\nInterrupted. Shutting down gracefully...")
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(sigCh)
	}()
	return ctx, cancel
}
