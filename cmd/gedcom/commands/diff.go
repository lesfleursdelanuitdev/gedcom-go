package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lesfleursdelanuitdev/gedcom-go/cmd/gedcom/internal"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/diff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [file1.ged] [file2.ged]",
	Short: "Compare two GEDCOM files",
	Long:  "Compare two GEDCOM files and show differences at a semantic level",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

func init() {
	diffCmd.Flags().String("strategy", "hybrid", "Matching strategy: xref, content, or hybrid")
	diffCmd.Flags().Float64("similarity", 0.85, "Similarity threshold for content matching (0.0-1.0)")
	diffCmd.Flags().Int("date-tolerance", 2, "Date tolerance in years")
	diffCmd.Flags().String("detail", "field", "Detail level: summary, field, or full")
	diffCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	diffCmd.Flags().String("format", "text", "Output format: text or json")
	diffCmd.Flags().Bool("no-history", false, "Disable change history tracking")
	diffCmd.Flags().Bool("include-unchanged", false, "Include unchanged records in output")
}

func runDiff(cmd *cobra.Command, args []string) error {
	file1 := args[0]
	file2 := args[1]

	// Load config
	config, err := internal.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize color
	internal.InitColor(config.Output.Color)

	// Check if files exist
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		internal.PrintError("✗ File not found: %s\n", file1)
		return fmt.Errorf("file not found: %s", file1)
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		internal.PrintError("✗ File not found: %s\n", file2)
		return fmt.Errorf("file not found: %s", file2)
	}

	// Get flags
	strategy, _ := cmd.Flags().GetString("strategy")
	similarity, _ := cmd.Flags().GetFloat64("similarity")
	dateTolerance, _ := cmd.Flags().GetInt("date-tolerance")
	detailLevel, _ := cmd.Flags().GetString("detail")
	outputFile, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	noHistory, _ := cmd.Flags().GetBool("no-history")
	includeUnchanged, _ := cmd.Flags().GetBool("include-unchanged")

	// Validate strategy
	if strategy != "xref" && strategy != "content" && strategy != "hybrid" {
		return fmt.Errorf("invalid strategy: %s (must be xref, content, or hybrid)", strategy)
	}

	// Validate similarity threshold
	if similarity < 0.0 || similarity > 1.0 {
		return fmt.Errorf("similarity threshold must be between 0.0 and 1.0")
	}

	// Validate detail level
	if detailLevel != "summary" && detailLevel != "field" && detailLevel != "full" {
		return fmt.Errorf("invalid detail level: %s (must be summary, field, or full)", detailLevel)
	}

	// Validate format
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format: %s (must be text or json)", format)
	}

	// Parse both files
	internal.PrintInfo("ℹ Parsing: %s\n", file1)
	p := parser.NewHierarchicalParser()
	tree1, err := p.Parse(file1)
	if err != nil {
		internal.PrintError("✗ Failed to parse %s: %v\n", file1, err)
		return fmt.Errorf("failed to parse %s: %w", file1, err)
	}

	internal.PrintInfo("ℹ Parsing: %s\n", file2)
	tree2, err := p.Parse(file2)
	if err != nil {
		internal.PrintError("✗ Failed to parse %s: %v\n", file2, err)
		return fmt.Errorf("failed to parse %s: %w", file2, err)
	}

	// Create diff configuration
	diffConfig := diff.DefaultConfig()
	diffConfig.MatchingStrategy = strategy
	diffConfig.SimilarityThreshold = similarity
	diffConfig.DateTolerance = dateTolerance
	diffConfig.DetailLevel = detailLevel
	diffConfig.TrackHistory = !noHistory
	diffConfig.IncludeUnchanged = includeUnchanged
	diffConfig.OutputFormat = format

	// Create differ
	internal.PrintInfo("ℹ Comparing files (strategy: %s)...\n", strategy)
	differ := diff.NewGedcomDiffer(diffConfig)

	// Compare
	result, err := differ.Compare(tree1, tree2)
	if err != nil {
		internal.PrintError("✗ Comparison failed: %v\n", err)
		return fmt.Errorf("comparison failed: %w", err)
	}

	// Generate report
	var report string
	if format == "json" {
		// JSON format
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			internal.PrintError("✗ Failed to generate JSON report: %v\n", err)
			return fmt.Errorf("failed to generate JSON report: %w", err)
		}
		report = string(jsonData)
	} else {
		// Text format
		report, err = differ.GenerateReport(result)
		if err != nil {
			internal.PrintError("✗ Failed to generate report: %v\n", err)
			return fmt.Errorf("failed to generate report: %w", err)
		}
	}

	// Output report
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(report), 0644); err != nil {
			internal.PrintError("✗ Failed to write output file: %v\n", err)
			return fmt.Errorf("failed to write output file: %w", err)
		}
		internal.PrintSuccess("✓ Diff report written to: %s\n", outputFile)
	} else {
		fmt.Print(report)
	}

	// Print summary
	summary := result.Summary
	internal.PrintInfo("\nℹ Summary:\n")
	internal.PrintInfo("  Added:     %d individuals, %d families\n",
		summary.Changes.Added.Individuals,
		summary.Changes.Added.Families)
	internal.PrintInfo("  Removed:   %d individuals, %d families\n",
		summary.Changes.Removed.Individuals,
		summary.Changes.Removed.Families)
	internal.PrintInfo("  Modified:  %d individuals, %d families\n",
		summary.Changes.Modified.Individuals,
		summary.Changes.Modified.Families)

	return nil
}

// GetDiffCommand returns the diff command
func GetDiffCommand() *cobra.Command {
	return diffCmd
}

