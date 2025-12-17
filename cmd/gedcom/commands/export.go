package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourorg/gedcom/cmd/gedcom/internal"
	"github.com/yourorg/gedcom/internal/exporter"
	"github.com/yourorg/gedcom/internal/parser"
	"github.com/yourorg/gedcom/pkg/gedcom"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export GEDCOM to different formats",
	Long:  "Export GEDCOM data to JSON, XML, YAML, or GEDCOM format",
}

var exportJsonCmd = &cobra.Command{
	Use:   "json [input.ged]",
	Short: "Export to JSON",
	Long:  "Export a GEDCOM file to JSON format",
	Args:  cobra.ExactArgs(1),
	RunE:  runExportJSON,
}

var exportXmlCmd = &cobra.Command{
	Use:   "xml [input.ged]",
	Short: "Export to XML",
	Long:  "Export a GEDCOM file to XML format",
	Args:  cobra.ExactArgs(1),
	RunE:  runExportXML,
}

var exportYamlCmd = &cobra.Command{
	Use:   "yaml [input.ged]",
	Short: "Export to YAML",
	Long:  "Export a GEDCOM file to YAML format",
	Args:  cobra.ExactArgs(1),
	RunE:  runExportYAML,
}

var exportGedcomCmd = &cobra.Command{
	Use:   "gedcom [input.ged]",
	Short: "Export to GEDCOM",
	Long:  "Export (re-export) a GEDCOM file to GEDCOM format",
	Args:  cobra.ExactArgs(1),
	RunE:  runExportGEDCOM,
}

func init() {
	// Add flags to all export commands
	for _, cmd := range []*cobra.Command{exportJsonCmd, exportXmlCmd, exportYamlCmd, exportGedcomCmd} {
		cmd.Flags().StringP("output", "o", "", "Output file (required)")
		cmd.MarkFlagRequired("output")
		cmd.Flags().Bool("pretty", true, "Pretty-print output")
		cmd.Flags().Int("indent", 2, "Indentation level")
	}

	// Add subcommands
	exportCmd.AddCommand(exportJsonCmd)
	exportCmd.AddCommand(exportXmlCmd)
	exportCmd.AddCommand(exportYamlCmd)
	exportCmd.AddCommand(exportGedcomCmd)
}

func runExportJSON(cmd *cobra.Command, args []string) error {
	return runExport(cmd, args, "json")
}

func runExportXML(cmd *cobra.Command, args []string) error {
	return runExport(cmd, args, "xml")
}

func runExportYAML(cmd *cobra.Command, args []string) error {
	return runExport(cmd, args, "yaml")
}

func runExportGEDCOM(cmd *cobra.Command, args []string) error {
	return runExport(cmd, args, "gedcom")
}

func runExport(cmd *cobra.Command, args []string, format string) error {
	inputFile := args[0]
	outputFile, _ := cmd.Flags().GetString("output")
	_, _ = cmd.Flags().GetBool("pretty") // Reserved for future use
	_, _ = cmd.Flags().GetInt("indent")  // Reserved for future use

	// Load config
	config, err := internal.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize color
	internal.InitColor(config.Output.Color)

	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		internal.PrintError("✗ File not found: %s\n", inputFile)
		return fmt.Errorf("file not found: %s", inputFile)
	}

	// Parse file
	internal.PrintInfo("ℹ Parsing: %s\n", inputFile)

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(inputFile)
	if err != nil {
		internal.PrintError("✗ Parse failed: %v\n", err)
		return err
	}

	// Show progress
	var progressBar *internal.ProgressBar
	if config.Output.Progress && !internal.IsQuietMode() {
		progressBar = internal.NewProgressBar(100, "Exporting...")
		defer progressBar.Finish()
	}

	// Export
	internal.PrintInfo("ℹ Exporting to %s: %s\n", format, outputFile)

	errorManager := gedcom.NewErrorManager()

	switch format {
	case "json":
		exporter := exporter.NewJsonExporter(errorManager)
		if progressBar != nil {
			progressBar.Set(50)
		}
		err = exporter.ExportToFile(tree, outputFile)
		if progressBar != nil {
			progressBar.Set(100)
		}

	case "xml":
		xmlExporter := exporter.NewXMLExporter(errorManager)
		if progressBar != nil {
			progressBar.Set(50)
		}
		err = xmlExporter.ExportToFile(tree, outputFile)
		if progressBar != nil {
			progressBar.Set(100)
		}

	case "yaml":
		yamlExporter := exporter.NewYAMLExporter(errorManager)
		if progressBar != nil {
			progressBar.Set(50)
		}
		err = yamlExporter.ExportToFile(tree, outputFile)
		if progressBar != nil {
			progressBar.Set(100)
		}

	case "gedcom":
		gedcomExporter := exporter.NewGedcomExporter(errorManager, "gedcom-cli", "1.0.0")
		if progressBar != nil {
			progressBar.Set(50)
		}
		err = gedcomExporter.ExportToFile(tree, outputFile)
		if progressBar != nil {
			progressBar.Set(100)
		}

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		internal.PrintError("✗ Export failed: %v\n", err)
		return err
	}

	internal.PrintSuccess("✓ Exported successfully to: %s\n", outputFile)

	// Get file size
	if fileInfo, err := os.Stat(outputFile); err == nil {
		internal.PrintInfo("  File size: %d bytes\n", fileInfo.Size())
	}

	return nil
}

// GetExportCommand returns the export command
func GetExportCommand() *cobra.Command {
	return exportCmd
}
