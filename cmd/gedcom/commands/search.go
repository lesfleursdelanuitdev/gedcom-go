package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourorg/gedcom/cmd/gedcom/internal"
	"github.com/yourorg/gedcom/internal/parser"
	"github.com/yourorg/gedcom/pkg/gedcom"
	"github.com/yourorg/gedcom/pkg/gedcom/query"
)

var searchCmd = &cobra.Command{
	Use:   "search [input.ged]",
	Short: "Advanced search for individuals",
	Long:  "Search individuals with multiple filters, operators, and output options",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	// Name filters
	searchCmd.Flags().String("name", "", "Search by name (contains)")
	searchCmd.Flags().String("name-exact", "", "Search by name (exact match)")
	searchCmd.Flags().String("name-starts", "", "Search by name (starts with)")
	searchCmd.Flags().String("name-ends", "", "Search by name (ends with)")

	// Date filters
	searchCmd.Flags().String("birth-date", "", "Birth date (year, range, or before:YYYY/after:YYYY)")
	searchCmd.Flags().String("birth-year", "", "Birth year (shorthand for --birth-date)")
	searchCmd.Flags().String("birth-date-before", "", "Born before year")
	searchCmd.Flags().String("birth-date-after", "", "Born after year")

	// Place filters
	searchCmd.Flags().String("birth-place", "", "Birth place (contains)")

	// Demographics
	searchCmd.Flags().String("sex", "", "Sex (M, F, U)")

	// Boolean filters
	searchCmd.Flags().Bool("living", false, "Living individuals (has no death date)")
	searchCmd.Flags().Bool("deceased", false, "Deceased individuals (has death date)")
	searchCmd.Flags().Bool("has-children", false, "Has children")
	searchCmd.Flags().Bool("has-spouse", false, "Has spouse")
	searchCmd.Flags().Bool("no-children", false, "Does not have children")
	searchCmd.Flags().Bool("no-spouse", false, "Does not have spouse")

	// Output options
	searchCmd.Flags().StringP("format", "f", "table", "Output format (table, json, yaml, csv, list)")
	searchCmd.Flags().String("fields", "", "Comma-separated fields to display")
	searchCmd.Flags().String("sort", "", "Sort by field (name, birth_date, xref)")
	searchCmd.Flags().Bool("sort-desc", false, "Sort in descending order")
	searchCmd.Flags().IntP("limit", "n", 100, "Limit number of results (0 = no limit)")
	searchCmd.Flags().Bool("count-only", false, "Only show count of results")
	searchCmd.Flags().StringP("output", "o", "", "Output file")
	searchCmd.Flags().Bool("compact", false, "Compact output (xref and name only)")
}

func runSearch(cmd *cobra.Command, args []string) error {
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

	// Parse file
	internal.PrintInfo("ℹ Loading GEDCOM file: %s\n", inputFile)

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(inputFile)
	if err != nil {
		internal.PrintError("✗ Parse failed: %v\n", err)
		return err
	}

	// Build graph
	internal.PrintInfo("ℹ Building graph...\n")
	graph, err := query.BuildGraph(tree)
	if err != nil {
		internal.PrintError("✗ Graph build failed: %v\n", err)
		return err
	}

	// Create query builder
	qb, err := query.NewQuery(tree)
	if err != nil {
		internal.PrintError("✗ Query builder failed: %v\n", err)
		return err
	}

	// Build filter query
	filterQuery := qb.Filter()

	// Apply name filters
	if name, _ := cmd.Flags().GetString("name"); name != "" {
		filterQuery = filterQuery.ByName(name)
	} else if name, _ := cmd.Flags().GetString("name-exact"); name != "" {
		// Exact match - need to add custom filter
		filterQuery = filterQuery.Where(func(indi *gedcom.IndividualRecord) bool {
			return indi.GetName() == name
		})
	} else if name, _ := cmd.Flags().GetString("name-starts"); name != "" {
		filterQuery = filterQuery.Where(func(indi *gedcom.IndividualRecord) bool {
			return strings.HasPrefix(strings.ToLower(indi.GetName()), strings.ToLower(name))
		})
	} else if name, _ := cmd.Flags().GetString("name-ends"); name != "" {
		filterQuery = filterQuery.Where(func(indi *gedcom.IndividualRecord) bool {
			return strings.HasSuffix(strings.ToLower(indi.GetName()), strings.ToLower(name))
		})
	}

	// Apply birth date filters
	if err := applyBirthDateFilters(cmd, filterQuery); err != nil {
		return err
	}

	// Apply birth place filter
	if place, _ := cmd.Flags().GetString("birth-place"); place != "" {
		filterQuery = filterQuery.ByBirthPlace(place)
	}

	// Apply sex filter
	if sex, _ := cmd.Flags().GetString("sex"); sex != "" {
		filterQuery = filterQuery.BySex(sex)
	}

	// Apply boolean filters
	if living, _ := cmd.Flags().GetBool("living"); living {
		filterQuery = filterQuery.Living()
	}
	if deceased, _ := cmd.Flags().GetBool("deceased"); deceased {
		filterQuery = filterQuery.Deceased()
	}
	if hasChildren, _ := cmd.Flags().GetBool("has-children"); hasChildren {
		filterQuery = filterQuery.HasChildren()
	}
	if hasSpouse, _ := cmd.Flags().GetBool("has-spouse"); hasSpouse {
		filterQuery = filterQuery.HasSpouse()
	}
	if noChildren, _ := cmd.Flags().GetBool("no-children"); noChildren {
		filterQuery = filterQuery.Where(func(indi *gedcom.IndividualRecord) bool {
			node := graph.GetIndividual(indi.XrefID())
			return node != nil && len(node.Children) == 0
		})
	}
	if noSpouse, _ := cmd.Flags().GetBool("no-spouse"); noSpouse {
		filterQuery = filterQuery.Where(func(indi *gedcom.IndividualRecord) bool {
			node := graph.GetIndividual(indi.XrefID())
			return node != nil && len(node.Spouses) == 0
		})
	}

	// Execute query
	internal.PrintInfo("ℹ Searching...\n")

	results, err := filterQuery.Execute()
	if err != nil {
		internal.PrintError("✗ Search failed: %v\n", err)
		return err
	}

	totalCount := len(results)

	// Count only mode
	if countOnly, _ := cmd.Flags().GetBool("count-only"); countOnly {
		internal.PrintSuccess("✓ Found %d individuals\n", totalCount)
		return nil
	}

	// Apply limit
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 && totalCount > limit {
		results = results[:limit]
		internal.PrintSuccess("✓ Found %d individuals\n", totalCount)
		internal.PrintWarning("⚠ Showing first %d results (use --limit 0 to see all)\n", limit)
	} else {
		internal.PrintSuccess("✓ Found %d individuals\n", totalCount)
	}

	if len(results) == 0 {
		internal.PrintInfo("  No results found\n")
		return nil
	}

	// Sort results
	if sortField, _ := cmd.Flags().GetString("sort"); sortField != "" {
		sortDesc, _ := cmd.Flags().GetBool("sort-desc")
		sortResults(results, sortField, sortDesc)
	}

	// Format and output results
	format, _ := cmd.Flags().GetString("format")
	fields, _ := cmd.Flags().GetString("fields")
	compact, _ := cmd.Flags().GetBool("compact")
	outputFile, _ := cmd.Flags().GetString("output")

	if err := formatSearchResults(results, format, fields, compact, outputFile); err != nil {
		internal.PrintError("✗ Output failed: %v\n", err)
		return err
	}

	return nil
}

func applyBirthDateFilters(cmd *cobra.Command, filterQuery *query.FilterQuery) error {
	// Check birth-year first (shorthand)
	if yearStr, _ := cmd.Flags().GetString("birth-year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			return fmt.Errorf("invalid birth year: %s", yearStr)
		}
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)
		filterQuery.ByBirthDate(start, end)
		return nil
	}

	// Check birth-date-before
	if beforeStr, _ := cmd.Flags().GetString("birth-date-before"); beforeStr != "" {
		year, err := strconv.Atoi(beforeStr)
		if err != nil {
			return fmt.Errorf("invalid year: %s", beforeStr)
		}
		end := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)
		start := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC) // Very early date
		filterQuery.ByBirthDate(start, end)
		return nil
	}

	// Check birth-date-after
	if afterStr, _ := cmd.Flags().GetString("birth-date-after"); afterStr != "" {
		year, err := strconv.Atoi(afterStr)
		if err != nil {
			return fmt.Errorf("invalid year: %s", afterStr)
		}
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC) // Very late date
		filterQuery.ByBirthDate(start, end)
		return nil
	}

	// Check birth-date (range or single year)
	if dateStr, _ := cmd.Flags().GetString("birth-date"); dateStr != "" {
		// Try to parse as range (YYYY-YYYY)
		if strings.Contains(dateStr, "-") && !strings.HasPrefix(dateStr, "before:") && !strings.HasPrefix(dateStr, "after:") {
			parts := strings.Split(dateStr, "-")
			if len(parts) == 2 {
				startYear, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
				endYear, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err1 == nil && err2 == nil {
					start := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC)
					end := time.Date(endYear, 12, 31, 23, 59, 59, 999999999, time.UTC)
					filterQuery.ByBirthDate(start, end)
					return nil
				}
			}
		}

		// Try to parse as single year
		if year, err := strconv.Atoi(dateStr); err == nil {
			start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			end := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)
			filterQuery.ByBirthDate(start, end)
			return nil
		}

		return fmt.Errorf("invalid birth date format: %s (use YYYY or YYYY-YYYY)", dateStr)
	}

	return nil
}

func sortResults(results []*gedcom.IndividualRecord, sortField string, desc bool) {
	switch sortField {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			nameI := strings.ToLower(results[i].GetName())
			nameJ := strings.ToLower(results[j].GetName())
			if desc {
				return nameI > nameJ
			}
			return nameI < nameJ
		})
	case "birth_date":
		sort.Slice(results, func(i, j int) bool {
			dateI, errI := results[i].GetBirthDateParsed()
			dateJ, errJ := results[j].GetBirthDateParsed()

			// Handle errors - put invalid dates at end
			if errI != nil && errJ != nil {
				return false
			}
			if errI != nil {
				return desc // Invalid dates go to end
			}
			if errJ != nil {
				return !desc // Invalid dates go to end
			}

			timeI := dateI.Earliest()
			timeJ := dateJ.Earliest()
			if desc {
				return timeI.After(timeJ)
			}
			return timeI.Before(timeJ)
		})
	case "xref":
		sort.Slice(results, func(i, j int) bool {
			xrefI := results[i].XrefID()
			xrefJ := results[j].XrefID()
			if desc {
				return xrefI > xrefJ
			}
			return xrefI < xrefJ
		})
	}
}

func formatSearchResults(results []*gedcom.IndividualRecord, format, fields string, compact bool, outputFile string) error {
	// Determine fields to display
	fieldList := determineFields(fields, compact)

	// Format based on type
	switch format {
	case "table":
		return formatTable(results, fieldList, outputFile)
	case "json":
		return formatJSON(results, fieldList, outputFile)
	case "list":
		return formatList(results, outputFile)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func determineFields(fields string, compact bool) []string {
	if compact {
		return []string{"xref", "name"}
	}

	if fields != "" {
		return strings.Split(fields, ",")
	}

	// Default fields
	return []string{"xref", "name", "sex", "birth_date", "death_date"}
}

func formatTable(results []*gedcom.IndividualRecord, fields []string, outputFile string) error {
	if len(results) == 0 {
		internal.PrintInfo("  No results to display\n")
		return nil
	}

	// Build headers
	headers := make([]string, len(fields))
	for i, field := range fields {
		headers[i] = strings.ToUpper(strings.ReplaceAll(field, "_", " "))
	}

	// Build rows
	rows := make([][]string, len(results))
	for i, indi := range results {
		row := make([]string, len(fields))
		for j, field := range fields {
			value := getFieldValue(indi, field)
			if value == "" {
				value = "-"
			}
			row[j] = value
		}
		rows[i] = row
	}

	// Write to file or stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer file.Close()
		// Write table to file (simplified)
		fmt.Fprintf(file, "%s\n", strings.Join(headers, " | "))
		fmt.Fprintf(file, "%s\n", strings.Repeat("-", len(strings.Join(headers, " | "))))
		for _, row := range rows {
			fmt.Fprintf(file, "%s\n", strings.Join(row, " | "))
		}
		internal.PrintSuccess("✓ Results written to: %s\n", outputFile)
	} else {
		// Use internal table writer
		internal.WriteTable(headers, rows)
	}

	return nil
}

func formatJSON(results []*gedcom.IndividualRecord, fields []string, outputFile string) error {
	// Build JSON structure
	jsonResults := make([]map[string]interface{}, len(results))
	for i, indi := range results {
		result := make(map[string]interface{})
		for _, field := range fields {
			result[field] = getFieldValue(indi, field)
		}
		jsonResults[i] = result
	}

	// Format as JSON
	data := map[string]interface{}{
		"count":   len(results),
		"results": jsonResults,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file or stdout
	if outputFile != "" {
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		internal.PrintSuccess("✓ Results written to: %s\n", outputFile)
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func formatList(results []*gedcom.IndividualRecord, outputFile string) error {
	// Simple list of xref IDs
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer file.Close()
		for _, indi := range results {
			fmt.Fprintf(file, "%s\n", indi.XrefID())
		}
		internal.PrintSuccess("✓ Results written to: %s\n", outputFile)
	} else {
		for _, indi := range results {
			fmt.Println(indi.XrefID())
		}
	}

	return nil
}

func getFieldValue(indi *gedcom.IndividualRecord, field string) string {
	switch field {
	case "xref":
		return indi.XrefID()
	case "name":
		return indi.GetName()
	case "given_name":
		return indi.GetGivenName()
	case "surname":
		return indi.GetSurname()
	case "sex":
		return indi.GetSex()
	case "birth_date":
		return indi.GetBirthDate()
	case "birth_place":
		return indi.GetBirthPlace()
	case "death_date":
		return indi.GetDeathDate()
	case "death_place":
		return indi.GetDeathPlace()
	default:
		return ""
	}
}

// GetSearchCommand returns the search command
func GetSearchCommand() *cobra.Command {
	return searchCmd
}
