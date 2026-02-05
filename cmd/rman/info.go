package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <manifest>",
	Short: "Show manifest metadata",
	Long:  "Display metadata about an RMAN manifest including ID, version, file count, total size, flags, and bundle count.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := loadManifest(args[0])
		if err != nil {
			return err
		}

		var totalSize uint64
		bundles := make(map[uint64]bool)
		var totalChunks int
		for _, f := range m.Files {
			totalSize += uint64(f.FileSize)
			for _, c := range f.Chunks {
				bundles[c.BundleID] = true
				totalChunks++
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Manifest ID:\t%016X\n", m.ID)
		fmt.Fprintf(w, "Version:\t%s\n", m.Version)
		fmt.Fprintf(w, "Files:\t%d\n", len(m.Files))
		fmt.Fprintf(w, "Total Size:\t%s\n", formatBytes(totalSize))
		fmt.Fprintf(w, "Bundles:\t%d\n", len(bundles))
		fmt.Fprintf(w, "Chunks:\t%d\n", totalChunks)
		fmt.Fprintf(w, "Flags:\t%d\n", len(m.Flags))
		w.Flush()

		if len(m.Flags) > 0 {
			fmt.Println("\nFlags:")
			for _, f := range m.Flags {
				fmt.Printf("  [%d] %s\n", f.FlagID, f.Name)
			}
		}

		return nil
	},
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
