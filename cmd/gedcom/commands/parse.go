package commands

import (
	"fmt"
	"os"

	"github.com/lesfleursdelanuitdev/gedcom-go/cmd/gedcom/internal"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/exporter"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse GEDCOM files",
	Long:  "Parse and optionally validate GEDCOM files",
}

var parseFileCmd = &cobra.Command{
	Use:   "file [input.ged]",
	Short: "Parse a GEDCOM file",
	Long:  "Parse a GEDCOM file and optionally export to different formats",
	Args:  cobra.ExactArgs(1),
	RunE:  runParseFile,
}

var parseValidateCmd = &cobra.Command{
	Use:   "validate [input.ged]",
	Short: "Parse with validation",
	Long:  "Parse a GEDCOM file with full validation",
	Args:  cobra.ExactArgs(1),
	RunE:  runParseValidate,
}

var parseCheckCmd = &cobra.Command{
	Use:   "check [input.ged]",
	Short: "Quick syntax check",
	Long:  "Perform a quick syntax check on a GEDCOM file",
	Args:  cobra.ExactArgs(1),
	RunE:  runParseCheck,
}

func init() {
	// Add flags to parse file command
	parseFileCmd.Flags().StringP("output", "o", "", "Output file (JSON/XML/YAML)")
	parseFileCmd.Flags().StringP("format", "f", "json", "Output format (json/xml/yaml)")
	parseFileCmd.Flags().Bool("parallel", false, "Use parallel parser")
	parseFileCmd.Flags().Bool("stream", false, "Use streaming parser for large files")
	parseFileCmd.Flags().BoolP("verbose", "v", false, "Show detailed parsing info")

	// Add flags to parse validate command
	parseValidateCmd.Flags().Bool("strict", false, "Fail on errors")

	// Add subcommands
	parseCmd.AddCommand(parseFileCmd)
	parseCmd.AddCommand(parseValidateCmd)
	parseCmd.AddCommand(parseCheckCmd)
}

func runParseFile(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	outputFile, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	useParallel, _ := cmd.Flags().GetBool("parallel")
	useStream, _ := cmd.Flags().GetBool("stream")
	verbose, _ := cmd.Flags().GetBool("verbose")

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

	// Determine parser type
	parserType := "hierarchical"
	if useParallel {
		parserType = "parallel"
	} else if useStream {
		parserType = "stream"
	} else if config.Parser.Type != "" {
		parserType = config.Parser.Type
	}

	// Create parser (for now, only hierarchical is available)
	// Parallel and streaming parsers will be added later
	p := parser.NewHierarchicalParser()
	if parserType != "hierarchical" {
		internal.PrintInfo("  Note: %s parser not yet implemented, using hierarchical\n", parserType)
	}

	// Show progress
	var progressBar *internal.ProgressBar
	if config.Output.Progress && !internal.IsQuietMode() {
		// Estimate file size for progress
		fileInfo, _ := os.Stat(inputFile)
		progressBar = internal.NewProgressBar(fileInfo.Size(), "Parsing...")
		defer progressBar.Finish()
	}

	// Parse file
	internal.PrintInfo("ℹ Parsing file: %s\n", inputFile)
	if verbose {
		internal.PrintInfo("  Parser type: %s\n", parserType)
	}

	tree, err := p.Parse(inputFile)
	if err != nil {
		internal.PrintError("✗ Parse failed: %v\n", err)
		return err
	}

	// Update progress
	if progressBar != nil {
		fileInfo, _ := os.Stat(inputFile)
		progressBar.Set(int(fileInfo.Size()))
	}

	// Get statistics
	individuals := tree.GetAllIndividuals()
	families := tree.GetAllFamilies()

	internal.PrintSuccess("✓ Parsed successfully\n")
	internal.PrintInfo("  Individuals: %d\n", len(individuals))
	internal.PrintInfo("  Families: %d\n", len(families))

	// Export if output specified
	if outputFile != "" {
		internal.PrintInfo("ℹ Exporting to: %s\n", outputFile)
		if err := exportTree(tree, outputFile, format, config); err != nil {
			internal.PrintError("✗ Export failed: %v\n", err)
			return err
		}
		internal.PrintSuccess("✓ Exported successfully\n")
	}

	return nil
}

func runParseValidate(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	strict, _ := cmd.Flags().GetBool("strict")

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
	internal.PrintInfo("ℹ Parsing and validating: %s\n", inputFile)

	p := parser.NewHierarchicalParser()
	_, err = p.Parse(inputFile)
	if err != nil {
		if strict {
			internal.PrintError("✗ Parse failed: %v\n", err)
			return err
		}
		internal.PrintWarning("⚠ Parse completed with errors\n")
	}

	// Get errors from parser
	errors := p.GetErrors()
	if len(errors) == 0 {
		internal.PrintSuccess("✓ No validation errors found\n")
		return nil
	}

	// Report errors
	internal.PrintWarning("⚠ Found %d validation issues\n", len(errors))
	for _, err := range errors {
		switch err.Severity {
		case "severe":
			internal.PrintError("  ✗ [SEVERE] %s\n", err.Message)
		case "warning":
			internal.PrintWarning("  ⚠ [WARNING] %s\n", err.Message)
		case "info":
			internal.PrintInfo("  ℹ [INFO] %s\n", err.Message)
		case "hint":
			internal.PrintHint("  💡 [HINT] %s\n", err.Message)
		}
	}

	if strict && len(errors) > 0 {
		return fmt.Errorf("validation failed with %d errors", len(errors))
	}

	return nil
}

func runParseCheck(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

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

	// Quick syntax check
	internal.PrintInfo("ℹ Checking syntax: %s\n", inputFile)

	// Use parser to check syntax
	p := parser.NewHierarchicalParser()
	_, err = p.Parse(inputFile)
	if err != nil {
		internal.PrintError("✗ Syntax check failed: %v\n", err)
		return err
	}

	internal.PrintSuccess("✓ Syntax check passed\n")
	return nil
}

func exportTree(tree *gedcom.GedcomTree, outputFile string, format string, config *internal.Config) error {
	switch format {
	case "json":
		exporter := exporter.NewJsonExporter(gedcom.NewErrorManager())
		return exporter.ExportToFile(tree, outputFile)
	case "xml":
		// XML export will be implemented in export command
		return fmt.Errorf("XML export not yet implemented in parse command")
	case "yaml":
		// YAML export will be implemented in export command
		return fmt.Errorf("YAML export not yet implemented in parse command")
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetParseCommand returns the parse command
func GetParseCommand() *cobra.Command {
	return parseCmd
}
