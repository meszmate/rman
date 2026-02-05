package main

import (
	"os"

	"github.com/meszmate/rman"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <manifest>",
	Short: "Export manifest as JSON",
	Long:  "Export the full manifest or a file list as JSON.",
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

var (
	exportFileList bool
	exportPretty   bool
)

func init() {
	exportCmd.Flags().BoolVar(&exportFileList, "files-only", false, "Export only the file list (simplified)")
	exportCmd.Flags().BoolVar(&exportPretty, "pretty", true, "Pretty-print JSON output")
}

func runExport(cmd *cobra.Command, args []string) error {
	m, err := loadManifest(args[0])
	if err != nil {
		return err
	}

	if exportFileList {
		return rman.ExportFileList(os.Stdout, m, exportPretty)
	}
	return rman.ExportJSON(os.Stdout, m, exportPretty)
}
