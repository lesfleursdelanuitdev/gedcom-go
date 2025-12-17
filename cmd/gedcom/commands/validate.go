package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourorg/gedcom/cmd/gedcom/internal"
	"github.com/yourorg/gedcom/internal/parser"
	"github.com/yourorg/gedcom/internal/validator"
	"github.com/yourorg/gedcom/pkg/gedcom"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate GEDCOM files",
	Long:  "Validate GEDCOM files with severity levels",
}

var validateBasicCmd = &cobra.Command{
	Use:   "basic [input.ged]",
	Short: "Basic validation",
	Long:  "Perform basic validation on a GEDCOM file",
	Args:  cobra.ExactArgs(1),
	RunE:  runValidateBasic,
}

var validateAdvancedCmd = &cobra.Command{
	Use:   "advanced [input.ged]",
	Short: "Advanced validation",
	Long:  "Perform advanced validation with configurable severity levels",
	Args:  cobra.ExactArgs(1),
	RunE:  runValidateAdvanced,
}

func init() {
	// Add flags to basic command
	validateBasicCmd.Flags().Bool("fix", false, "Attempt to fix common issues")
	validateBasicCmd.Flags().String("fix-output", "", "Output file for fixed GEDCOM")

	// Add flags to advanced command
	validateAdvancedCmd.Flags().String("severity", "warning", "Minimum severity (severe/warning/info/hint)")
	validateAdvancedCmd.Flags().StringP("output", "o", "", "Report output file")
	validateAdvancedCmd.Flags().String("format", "text", "Report format (text/json/html)")

	// Add subcommands
	validateCmd.AddCommand(validateBasicCmd)
	validateCmd.AddCommand(validateAdvancedCmd)
}

func runValidateBasic(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	fix, _ := cmd.Flags().GetBool("fix")
	fixOutput, _ := cmd.Flags().GetString("fix-output")

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
	internal.PrintInfo("ℹ Validating: %s\n", inputFile)

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(inputFile)
	if err != nil {
		internal.PrintError("✗ Parse failed: %v\n", err)
		return err
	}

	// Get parsing errors
	parseErrors := p.GetErrors()
	if len(parseErrors) > 0 {
		internal.PrintWarning("⚠ Found %d parsing issues\n", len(parseErrors))
		for _, err := range parseErrors {
			switch err.Severity {
			case gedcom.SeveritySevere:
				internal.PrintError("  ✗ [SEVERE] %s\n", err.Message)
			case gedcom.SeverityWarning:
				internal.PrintWarning("  ⚠ [WARNING] %s\n", err.Message)
			case gedcom.SeverityInfo:
				internal.PrintInfo("  ℹ [INFO] %s\n", err.Message)
			case gedcom.SeverityHint:
				internal.PrintHint("  💡 [HINT] %s\n", err.Message)
			}
		}
	}

	// Run basic validator
	internal.PrintInfo("ℹ Running basic validation...\n")

	errorManager := gedcom.NewErrorManager()
	basicValidator := validator.NewGedcomValidator(errorManager)
	validationErr := basicValidator.Validate(tree)
	if validationErr != nil {
		internal.PrintError("✗ Validation failed: %v\n", validationErr)
		return validationErr
	}

	// Get validation errors
	var errors []*gedcom.GedcomError
	if errorManager != nil {
		errors = errorManager.Errors()
		if len(errors) > 0 {
			internal.PrintWarning("⚠ Found %d validation issues\n", len(errors))
			for _, err := range errors {
				switch err.Severity {
				case gedcom.SeveritySevere:
					internal.PrintError("  ✗ [SEVERE] %s\n", err.Message)
				case gedcom.SeverityWarning:
					internal.PrintWarning("  ⚠ [WARNING] %s\n", err.Message)
				case gedcom.SeverityInfo:
					internal.PrintInfo("  ℹ [INFO] %s\n", err.Message)
				case gedcom.SeverityHint:
					internal.PrintHint("  💡 [HINT] %s\n", err.Message)
				}
			}
		} else {
			internal.PrintSuccess("✓ No validation errors found\n")
		}
	} else {
		internal.PrintSuccess("✓ Basic validation passed\n")
	}

	// Fix mode (placeholder for future implementation)
	if fix {
		internal.PrintInfo("ℹ Fix mode not yet implemented\n")
		if fixOutput != "" {
			internal.PrintInfo("  Would write fixed file to: %s\n", fixOutput)
		}
	}

	return nil
}

func runValidateAdvanced(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	severityStr, _ := cmd.Flags().GetString("severity")
	outputFile, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")

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
	internal.PrintInfo("ℹ Validating (advanced): %s\n", inputFile)

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(inputFile)
	if err != nil {
		internal.PrintError("✗ Parse failed: %v\n", err)
		return err
	}

	// Run advanced validator
	internal.PrintInfo("ℹ Running advanced validation...\n")
	internal.PrintInfo("  Severity threshold: %s\n", severityStr)

	errorManager := gedcom.NewErrorManager()
	basicValidator := validator.NewGedcomValidator(errorManager)
	basicValidator.EnableAdvancedValidation()
	validationErr := basicValidator.Validate(tree)
	if validationErr != nil {
		internal.PrintError("✗ Validation failed: %v\n", validationErr)
		return validationErr
	}

	// Filter errors by severity
	errorManager = basicValidator.GetErrorManager()
	if errorManager != nil {
		allErrors := errorManager.Errors()
		filteredErrors := filterErrorsBySeverity(allErrors, severityStr)

		if len(filteredErrors) > 0 {
			internal.PrintWarning("⚠ Found %d validation issues (severity >= %s)\n", len(filteredErrors), severityStr)
			for _, err := range filteredErrors {
				switch err.Severity {
				case gedcom.SeveritySevere:
					internal.PrintError("  ✗ [SEVERE] %s\n", err.Message)
				case gedcom.SeverityWarning:
					internal.PrintWarning("  ⚠ [WARNING] %s\n", err.Message)
				case gedcom.SeverityInfo:
					internal.PrintInfo("  ℹ [INFO] %s\n", err.Message)
				case gedcom.SeverityHint:
					internal.PrintHint("  💡 [HINT] %s\n", err.Message)
				}
			}

			// Export report if requested
			if outputFile != "" {
				if err := exportValidationReport(filteredErrors, outputFile, format); err != nil {
					internal.PrintError("✗ Failed to export report: %v\n", err)
					return err
				}
				internal.PrintSuccess("✓ Report exported to: %s\n", outputFile)
			}
		} else {
			internal.PrintSuccess("✓ No validation errors found (severity >= %s)\n", severityStr)
		}
	} else {
		internal.PrintSuccess("✓ Advanced validation passed\n")
	}

	return nil
}

func filterErrorsBySeverity(errors []*gedcom.GedcomError, minSeverity string) []*gedcom.GedcomError {
	severityOrder := map[gedcom.ErrorSeverity]int{
		gedcom.SeverityHint:    0,
		gedcom.SeverityInfo:    1,
		gedcom.SeverityWarning: 2,
		gedcom.SeveritySevere:  3,
	}

	// Convert string to ErrorSeverity
	var minSeverityEnum gedcom.ErrorSeverity
	switch minSeverity {
	case "hint":
		minSeverityEnum = gedcom.SeverityHint
	case "info":
		minSeverityEnum = gedcom.SeverityInfo
	case "warning":
		minSeverityEnum = gedcom.SeverityWarning
	case "severe":
		minSeverityEnum = gedcom.SeveritySevere
	default:
		minSeverityEnum = gedcom.SeverityWarning
	}

	minLevel, ok := severityOrder[minSeverityEnum]
	if !ok {
		minLevel = severityOrder[gedcom.SeverityWarning]
	}

	filtered := make([]*gedcom.GedcomError, 0)
	for _, err := range errors {
		if level, ok := severityOrder[err.Severity]; ok && level >= minLevel {
			filtered = append(filtered, err)
		}
	}

	return filtered
}

func exportValidationReport(errors []*gedcom.GedcomError, outputFile string, format string) error {
	// Simple JSON export for now
	if format == "json" {
		data := map[string]interface{}{
			"total_errors": len(errors),
			"errors":       errors,
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		return nil
	}

	// Text format
	content := fmt.Sprintf("Validation Report\n")
	content += fmt.Sprintf("=================\n\n")
	content += fmt.Sprintf("Total Errors: %d\n\n", len(errors))

	for i, err := range errors {
		content += fmt.Sprintf("%d. [%s] %s\n", i+1, err.Severity, err.Message)
		if err.LineNumber > 0 {
			content += fmt.Sprintf("   Line: %d\n", err.LineNumber)
		}
		content += "\n"
	}

	return os.WriteFile(outputFile, []byte(content), 0644)
}

// GetValidateCommand returns the validate command
func GetValidateCommand() *cobra.Command {
	return validateCmd
}
