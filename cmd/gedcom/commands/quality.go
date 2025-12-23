package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/cmd/gedcom/internal"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/validator"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"github.com/spf13/cobra"
)

var qualityCmd = &cobra.Command{
	Use:   "quality [input.ged]",
	Short: "Generate data quality reports",
	Long:  "Analyze GEDCOM file data quality and generate comprehensive reports",
	Args:  cobra.ExactArgs(1),
	RunE:  runQuality,
}

func init() {
	qualityCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	qualityCmd.Flags().String("format", "text", "Output format: text or json")
	qualityCmd.Flags().Bool("advanced", false, "Include advanced validation checks")
	qualityCmd.Flags().String("severity", "warning", "Minimum severity to include (severe/warning/info/hint)")
}

func runQuality(cmd *cobra.Command, args []string) error {
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

	// Get flags
	outputFile, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	advanced, _ := cmd.Flags().GetBool("advanced")
	severityStr, _ := cmd.Flags().GetString("severity")

	// Validate format
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format: %s (must be text or json)", format)
	}

	// Parse file
	internal.PrintInfo("ℹ Analyzing: %s\n", inputFile)

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(inputFile)
	if err != nil {
		internal.PrintError("✗ Parse failed: %v\n", err)
		return err
	}

	// Get parsing errors
	parseErrors := p.GetErrors()

	// Run validation
	internal.PrintInfo("ℹ Running validation...\n")
	errorManager := gedcom.NewErrorManager()
	basicValidator := validator.NewGedcomValidator(errorManager)
	if advanced {
		internal.PrintInfo("  Including advanced validation checks\n")
		basicValidator.EnableAdvancedValidation()
	}
	validationErr := basicValidator.Validate(tree)
	if validationErr != nil {
		internal.PrintWarning("⚠ Validation completed with errors\n")
	}

	// Get validation errors
	validationErrors := errorManager.Errors()

	// Build quality report
	report := buildQualityReport(tree, parseErrors, validationErrors, severityStr, advanced)

	// Generate output
	var output string
	if format == "json" {
		jsonData, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			internal.PrintError("✗ Failed to generate JSON report: %v\n", err)
			return fmt.Errorf("failed to generate JSON report: %w", err)
		}
		output = string(jsonData)
	} else {
		output = formatQualityReportText(report)
	}

	// Output report
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			internal.PrintError("✗ Failed to write output file: %v\n", err)
			return fmt.Errorf("failed to write output file: %w", err)
		}
		internal.PrintSuccess("✓ Quality report written to: %s\n", outputFile)
	} else {
		fmt.Print(output)
	}

	return nil
}

// QualityReport represents a comprehensive data quality report
type QualityReport struct {
	Timestamp      time.Time                `json:"timestamp"`
	File           string                   `json:"file"`
	Statistics     QualityStatistics        `json:"statistics"`
	Completeness   CompletenessMetrics      `json:"completeness"`
	Consistency    ConsistencyMetrics       `json:"consistency"`
	Errors         ErrorSummary             `json:"errors"`
	QualityScore   QualityScore             `json:"quality_score"`
	Recommendations []string                `json:"recommendations"`
}

// QualityStatistics provides overall statistics
type QualityStatistics struct {
	TotalIndividuals int `json:"total_individuals"`
	TotalFamilies     int `json:"total_families"`
	TotalNotes        int `json:"total_notes"`
	TotalSources      int `json:"total_sources"`
	TotalErrors       int `json:"total_errors"`
	ParseErrors       int `json:"parse_errors"`
	ValidationErrors  int `json:"validation_errors"`
}

// CompletenessMetrics measures data completeness
type CompletenessMetrics struct {
	IndividualsWithNames      int     `json:"individuals_with_names"`
	IndividualsWithBirthDates int     `json:"individuals_with_birth_dates"`
	IndividualsWithBirthPlaces int    `json:"individuals_with_birth_places"`
	IndividualsWithDeathDates int     `json:"individuals_with_death_dates"`
	FamiliesWithMarriageDates int     `json:"families_with_marriage_dates"`
	NameCompleteness          float64 `json:"name_completeness"`
	BirthDateCompleteness     float64 `json:"birth_date_completeness"`
	BirthPlaceCompleteness     float64 `json:"birth_place_completeness"`
	DeathDateCompleteness     float64 `json:"death_date_completeness"`
	MarriageDateCompleteness   float64 `json:"marriage_date_completeness"`
}

// ConsistencyMetrics measures data consistency
type ConsistencyMetrics struct {
	DateConsistencyIssues int `json:"date_consistency_issues"`
	RelationshipIssues    int `json:"relationship_issues"`
	CrossReferenceIssues  int `json:"cross_reference_issues"`
}

// ErrorSummary provides error breakdown
type ErrorSummary struct {
	Severe   int `json:"severe"`
	Warning  int `json:"warning"`
	Info     int `json:"info"`
	Hint     int `json:"hint"`
	ByType   map[string]int `json:"by_type"`
}

// QualityScore provides overall quality scores
type QualityScore struct {
	Overall      float64 `json:"overall"`
	Completeness float64 `json:"completeness"`
	Consistency  float64 `json:"consistency"`
	Accuracy     float64 `json:"accuracy"`
}

func buildQualityReport(tree *gedcom.GedcomTree, parseErrors []*gedcom.GedcomError, validationErrors []*gedcom.GedcomError, minSeverity string, advanced bool) *QualityReport {
	// Filter errors by severity
	filteredErrors := filterErrorsBySeverity(validationErrors, minSeverity)
	allErrors := append(parseErrors, filteredErrors...)

	// Get all records
	allIndi := tree.GetAllIndividuals()
	allFam := tree.GetAllFamilies()
	allNotes := tree.GetAllNotes()
	allSources := tree.GetAllSources()

	// Calculate statistics
	stats := QualityStatistics{
		TotalIndividuals: len(allIndi),
		TotalFamilies:     len(allFam),
		TotalNotes:        len(allNotes),
		TotalSources:      len(allSources),
		TotalErrors:       len(allErrors),
		ParseErrors:       len(parseErrors),
		ValidationErrors:  len(filteredErrors),
	}

	// Calculate completeness
	completeness := calculateCompleteness(allIndi, allFam)

	// Calculate consistency
	consistency := calculateConsistency(allErrors)

	// Calculate error summary
	errorSummary := calculateErrorSummary(allErrors)

	// Calculate quality scores
	qualityScore := calculateQualityScore(completeness, consistency, errorSummary, stats)

	// Generate recommendations
	recommendations := generateRecommendations(completeness, consistency, errorSummary, qualityScore)

	return &QualityReport{
		Timestamp:      time.Now(),
		File:           "", // Will be set by caller if needed
		Statistics:     stats,
		Completeness:   completeness,
		Consistency:    consistency,
		Errors:         errorSummary,
		QualityScore:   qualityScore,
		Recommendations: recommendations,
	}
}

func calculateCompleteness(allIndi map[string]gedcom.Record, allFam map[string]gedcom.Record) CompletenessMetrics {
	var withNames, withBirthDates, withBirthPlaces, withDeathDates int
	var withMarriageDates int

	for _, record := range allIndi {
		indi := record.(*gedcom.IndividualRecord)
		if indi.GetName() != "" {
			withNames++
		}
		if indi.GetBirthDate() != "" {
			withBirthDates++
		}
		if indi.GetBirthPlace() != "" {
			withBirthPlaces++
		}
		if indi.GetDeathDate() != "" {
			withDeathDates++
		}
	}

	for _, record := range allFam {
		fam := record.(*gedcom.FamilyRecord)
		if fam.GetMarriageDate() != "" {
			withMarriageDates++
		}
	}

	totalIndi := len(allIndi)
	totalFam := len(allFam)

	var nameCompleteness, birthDateCompleteness, birthPlaceCompleteness, deathDateCompleteness, marriageDateCompleteness float64
	if totalIndi > 0 {
		nameCompleteness = float64(withNames) / float64(totalIndi) * 100
		birthDateCompleteness = float64(withBirthDates) / float64(totalIndi) * 100
		birthPlaceCompleteness = float64(withBirthPlaces) / float64(totalIndi) * 100
		deathDateCompleteness = float64(withDeathDates) / float64(totalIndi) * 100
	}
	if totalFam > 0 {
		marriageDateCompleteness = float64(withMarriageDates) / float64(totalFam) * 100
	}

	return CompletenessMetrics{
		IndividualsWithNames:      withNames,
		IndividualsWithBirthDates:  withBirthDates,
		IndividualsWithBirthPlaces: withBirthPlaces,
		IndividualsWithDeathDates:  withDeathDates,
		FamiliesWithMarriageDates:  withMarriageDates,
		NameCompleteness:           nameCompleteness,
		BirthDateCompleteness:      birthDateCompleteness,
		BirthPlaceCompleteness:     birthPlaceCompleteness,
		DeathDateCompleteness:      deathDateCompleteness,
		MarriageDateCompleteness:   marriageDateCompleteness,
	}
}

func calculateConsistency(errors []*gedcom.GedcomError) ConsistencyMetrics {
	var dateIssues, relationshipIssues, crossRefIssues int

	for _, err := range errors {
		context := err.Context
		if contains(context, "date") || contains(context, "Date") {
			dateIssues++
		}
		if contains(context, "relationship") || contains(context, "Relationship") || contains(context, "family") {
			relationshipIssues++
		}
		if contains(context, "xref") || contains(context, "XREF") || contains(context, "cross-reference") {
			crossRefIssues++
		}
	}

	return ConsistencyMetrics{
		DateConsistencyIssues: dateIssues,
		RelationshipIssues:    relationshipIssues,
		CrossReferenceIssues:  crossRefIssues,
	}
}

func calculateErrorSummary(errors []*gedcom.GedcomError) ErrorSummary {
	summary := ErrorSummary{
		ByType: make(map[string]int),
	}

	for _, err := range errors {
		switch err.Severity {
		case gedcom.SeveritySevere:
			summary.Severe++
		case gedcom.SeverityWarning:
			summary.Warning++
		case gedcom.SeverityInfo:
			summary.Info++
		case gedcom.SeverityHint:
			summary.Hint++
		}

		// Count by context/type
		context := err.Context
		if context == "" {
			context = "unknown"
		}
		summary.ByType[context]++
	}

	return summary
}

func calculateQualityScore(completeness CompletenessMetrics, consistency ConsistencyMetrics, errors ErrorSummary, stats QualityStatistics) QualityScore {
	// Completeness score (average of all completeness metrics)
	completenessScore := (completeness.NameCompleteness +
		completeness.BirthDateCompleteness +
		completeness.BirthPlaceCompleteness +
		completeness.DeathDateCompleteness +
		completeness.MarriageDateCompleteness) / 5.0

	// Consistency score (inverse of issues, normalized)
	// Lower issues = higher score
	totalIssues := consistency.DateConsistencyIssues + consistency.RelationshipIssues + consistency.CrossReferenceIssues
	consistencyScore := 100.0
	if stats.TotalIndividuals > 0 {
		issueRate := float64(totalIssues) / float64(stats.TotalIndividuals)
		consistencyScore = 100.0 - (issueRate * 100.0)
		if consistencyScore < 0 {
			consistencyScore = 0
		}
	}

	// Accuracy score (inverse of errors)
	accuracyScore := 100.0
	if stats.TotalIndividuals > 0 {
		errorRate := float64(errors.Severe+errors.Warning) / float64(stats.TotalIndividuals)
		accuracyScore = 100.0 - (errorRate * 100.0)
		if accuracyScore < 0 {
			accuracyScore = 0
		}
	}

	// Overall score (weighted average)
	overallScore := (completenessScore*0.4 + consistencyScore*0.3 + accuracyScore*0.3)

	return QualityScore{
		Overall:      overallScore,
		Completeness: completenessScore,
		Consistency:  consistencyScore,
		Accuracy:     accuracyScore,
	}
}

func generateRecommendations(completeness CompletenessMetrics, consistency ConsistencyMetrics, errors ErrorSummary, score QualityScore) []string {
	var recommendations []string

	// Completeness recommendations
	if completeness.NameCompleteness < 90 {
		recommendations = append(recommendations, fmt.Sprintf("Add names for %.1f%% of individuals", 100-completeness.NameCompleteness))
	}
	if completeness.BirthDateCompleteness < 80 {
		recommendations = append(recommendations, fmt.Sprintf("Add birth dates for %.1f%% of individuals", 100-completeness.BirthDateCompleteness))
	}
	if completeness.BirthPlaceCompleteness < 70 {
		recommendations = append(recommendations, fmt.Sprintf("Add birth places for %.1f%% of individuals", 100-completeness.BirthPlaceCompleteness))
	}

	// Consistency recommendations
	if consistency.DateConsistencyIssues > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d date consistency issues", consistency.DateConsistencyIssues))
	}
	if consistency.RelationshipIssues > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d relationship issues", consistency.RelationshipIssues))
	}
	if consistency.CrossReferenceIssues > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Fix %d cross-reference issues", consistency.CrossReferenceIssues))
	}

	// Error recommendations
	if errors.Severe > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Address %d severe errors", errors.Severe))
	}
	if errors.Warning > 10 {
		recommendations = append(recommendations, fmt.Sprintf("Review %d warnings", errors.Warning))
	}

	// Overall score recommendation
	if score.Overall < 70 {
		recommendations = append(recommendations, "Overall quality score is below 70% - consider comprehensive data cleanup")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Data quality is good! No major issues detected.")
	}

	return recommendations
}

func formatQualityReportText(report *QualityReport) string {
	var output string

	output += "GEDCOM Data Quality Report\n"
	output += "==========================\n\n"
	output += fmt.Sprintf("Generated: %s\n\n", report.Timestamp.Format(time.RFC3339))

	// Statistics
	output += "Statistics:\n"
	output += fmt.Sprintf("  Total Individuals: %d\n", report.Statistics.TotalIndividuals)
	output += fmt.Sprintf("  Total Families:     %d\n", report.Statistics.TotalFamilies)
	output += fmt.Sprintf("  Total Notes:        %d\n", report.Statistics.TotalNotes)
	output += fmt.Sprintf("  Total Sources:      %d\n", report.Statistics.TotalSources)
	output += fmt.Sprintf("  Total Errors:       %d\n", report.Statistics.TotalErrors)
	output += fmt.Sprintf("    Parse Errors:     %d\n", report.Statistics.ParseErrors)
	output += fmt.Sprintf("    Validation Errors: %d\n\n", report.Statistics.ValidationErrors)

	// Quality Scores
	output += "Quality Scores:\n"
	output += fmt.Sprintf("  Overall:      %.1f%%\n", report.QualityScore.Overall)
	output += fmt.Sprintf("  Completeness: %.1f%%\n", report.QualityScore.Completeness)
	output += fmt.Sprintf("  Consistency:  %.1f%%\n", report.QualityScore.Consistency)
	output += fmt.Sprintf("  Accuracy:     %.1f%%\n\n", report.QualityScore.Accuracy)

	// Completeness
	output += "Completeness Metrics:\n"
	output += fmt.Sprintf("  Names:        %.1f%% (%d/%d)\n",
		report.Completeness.NameCompleteness,
		report.Completeness.IndividualsWithNames,
		report.Statistics.TotalIndividuals)
	output += fmt.Sprintf("  Birth Dates:  %.1f%% (%d/%d)\n",
		report.Completeness.BirthDateCompleteness,
		report.Completeness.IndividualsWithBirthDates,
		report.Statistics.TotalIndividuals)
	output += fmt.Sprintf("  Birth Places: %.1f%% (%d/%d)\n",
		report.Completeness.BirthPlaceCompleteness,
		report.Completeness.IndividualsWithBirthPlaces,
		report.Statistics.TotalIndividuals)
	output += fmt.Sprintf("  Death Dates:  %.1f%% (%d/%d)\n",
		report.Completeness.DeathDateCompleteness,
		report.Completeness.IndividualsWithDeathDates,
		report.Statistics.TotalIndividuals)
	output += fmt.Sprintf("  Marriage Dates: %.1f%% (%d/%d)\n\n",
		report.Completeness.MarriageDateCompleteness,
		report.Completeness.FamiliesWithMarriageDates,
		report.Statistics.TotalFamilies)

	// Consistency
	output += "Consistency Metrics:\n"
	output += fmt.Sprintf("  Date Issues:        %d\n", report.Consistency.DateConsistencyIssues)
	output += fmt.Sprintf("  Relationship Issues: %d\n", report.Consistency.RelationshipIssues)
	output += fmt.Sprintf("  Cross-Reference Issues: %d\n\n", report.Consistency.CrossReferenceIssues)

	// Errors
	output += "Error Summary:\n"
	output += fmt.Sprintf("  Severe:  %d\n", report.Errors.Severe)
	output += fmt.Sprintf("  Warning: %d\n", report.Errors.Warning)
	output += fmt.Sprintf("  Info:    %d\n", report.Errors.Info)
	output += fmt.Sprintf("  Hint:    %d\n\n", report.Errors.Hint)

	// Recommendations
	output += "Recommendations:\n"
	for i, rec := range report.Recommendations {
		output += fmt.Sprintf("  %d. %s\n", i+1, rec)
	}

	return output
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetQualityCommand returns the quality command
func GetQualityCommand() *cobra.Command {
	return qualityCmd
}

