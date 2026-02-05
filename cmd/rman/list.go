package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/meszmate/rman"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <manifest>",
	Short: "List files in a manifest",
	Long:  "List all files in an RMAN manifest with optional filtering and sorting.",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

var (
	listFilter string
	listRegex  string
	listJSON   bool
	listSort   string
)

func init() {
	listCmd.Flags().StringVarP(&listFilter, "filter", "f", "", "Glob filter pattern")
	listCmd.Flags().StringVarP(&listRegex, "regex", "r", "", "Regex filter pattern")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().StringVar(&listSort, "sort", "name", "Sort by: name, size")
}

func runList(cmd *cobra.Command, args []string) error {
	m, err := loadManifest(args[0])
	if err != nil {
		return err
	}

	files := m.Files
	if listFilter != "" || listRegex != "" {
		files, err = rman.FilterFiles(files, rman.Filter{
			Glob:  listFilter,
			Regex: listRegex,
		})
		if err != nil {
			return fmt.Errorf("filter error: %w", err)
		}
	}

	switch listSort {
	case "size":
		sort.Slice(files, func(i, j int) bool {
			return files[i].FileSize > files[j].FileSize
		})
	default:
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name < files[j].Name
		})
	}

	if listJSON {
		type entry struct {
			Name  string   `json:"name"`
			Size  uint32   `json:"size"`
			Flags []string `json:"flags,omitempty"`
		}
		entries := make([]entry, len(files))
		for i, f := range files {
			flags := make([]string, len(f.Flags))
			for j, fl := range f.Flags {
				flags[j] = fl.Name
			}
			entries[i] = entry{Name: f.Name, Size: f.FileSize, Flags: flags}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	for _, f := range files {
		fmt.Printf("%10s  %s\n", formatBytes(uint64(f.FileSize)), f.Name)
	}
	fmt.Printf("\n%d files\n", len(files))
	return nil
}
