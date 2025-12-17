package internal

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
	FormatCSV   OutputFormat = "csv"
)

// FormatOutput formats data according to the specified format
func FormatOutput(data interface{}, format OutputFormat, pretty bool) error {
	switch format {
	case FormatTable:
		return formatTable(data)
	case FormatJSON:
		return formatJSON(data, pretty)
	case FormatYAML:
		return formatYAML(data)
	case FormatCSV:
		return formatCSV(data)
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

// formatTable formats data as a table
func formatTable(data interface{}) error {
	// For now, simple table formatting
	// Can be extended for specific data types
	fmt.Printf("%+v\n", data)
	return nil
}

// formatJSON formats data as JSON
func formatJSON(data interface{}, pretty bool) error {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// formatYAML formats data as YAML
func formatYAML(data interface{}) error {
	output, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	fmt.Print(string(output))
	return nil
}

// formatCSV formats data as CSV
func formatCSV(data interface{}) error {
	// Simple CSV formatting
	// Can be extended for specific data types
	fmt.Printf("%+v\n", data)
	return nil
}

// WriteTable writes data as a formatted table
func WriteTable(headers []string, rows [][]string) {
	// Simple table implementation for now
	// Can be enhanced later with full tablewriter features
	if len(headers) > 0 {
		// Print headers
		for i, h := range headers {
			if i > 0 {
				fmt.Print(" | ")
			}
			if IsColorEnabled() {
				Info.Print(h)
			} else {
				fmt.Print(h)
			}
		}
		fmt.Println()
		// Print separator
		for i := 0; i < len(headers); i++ {
			if i > 0 {
				fmt.Print("---")
			}
			fmt.Print("---")
		}
		fmt.Println()
	}

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Print(cell)
		}
		fmt.Println()
	}
}
