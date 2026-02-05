package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/meszmate/rman"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <old-manifest> <new-manifest>",
	Short: "Show differences between two manifests",
	Long:  "Compare two RMAN manifests and show added, removed, and modified files.",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

var diffJSON bool

func init() {
	diffCmd.Flags().BoolVar(&diffJSON, "json", false, "Output as JSON")
}

func runDiff(cmd *cobra.Command, args []string) error {
	oldM, err := loadManifest(args[0])
	if err != nil {
		return fmt.Errorf("loading old manifest: %w", err)
	}
	newM, err := loadManifest(args[1])
	if err != nil {
		return fmt.Errorf("loading new manifest: %w", err)
	}

	result := rman.Diff(oldM, newM)

	if diffJSON {
		type jsonResult struct {
			Added    []string `json:"added"`
			Removed  []string `json:"removed"`
			Modified []string `json:"modified"`
		}
		jr := jsonResult{
			Added:    make([]string, len(result.Added)),
			Removed:  make([]string, len(result.Removed)),
			Modified: make([]string, len(result.Modified)),
		}
		for i, f := range result.Added {
			jr.Added[i] = f.Name
		}
		for i, f := range result.Removed {
			jr.Removed[i] = f.Name
		}
		for i, f := range result.Modified {
			jr.Modified[i] = f.Name
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(jr)
	}

	if len(result.Added) > 0 {
		fmt.Printf("Added (%d):\n", len(result.Added))
		for _, f := range result.Added {
			fmt.Printf("  + %s (%s)\n", f.Name, formatBytes(uint64(f.FileSize)))
		}
		fmt.Println()
	}
	if len(result.Removed) > 0 {
		fmt.Printf("Removed (%d):\n", len(result.Removed))
		for _, f := range result.Removed {
			fmt.Printf("  - %s (%s)\n", f.Name, formatBytes(uint64(f.FileSize)))
		}
		fmt.Println()
	}
	if len(result.Modified) > 0 {
		fmt.Printf("Modified (%d):\n", len(result.Modified))
		for _, f := range result.Modified {
			fmt.Printf("  ~ %s (%s)\n", f.Name, formatBytes(uint64(f.FileSize)))
		}
		fmt.Println()
	}

	total := len(result.Added) + len(result.Removed) + len(result.Modified)
	if total == 0 {
		fmt.Println("No differences found.")
	} else {
		fmt.Printf("Summary: %d added, %d removed, %d modified\n",
			len(result.Added), len(result.Removed), len(result.Modified))
	}

	return nil
}
