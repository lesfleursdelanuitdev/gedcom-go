// Package exporter provides functionality for exporting GEDCOM trees to various formats.
//
// The exporter package supports multiple output formats, allowing you to convert
// GEDCOM data to different representations for integration with other systems
// or for data transformation purposes.
//
// # Export Formats
//
// The package provides exporters for the following formats:
//
//   - GedcomExporter: Exports back to GEDCOM format
//   - JsonExporter: Exports to JSON format
//   - XMLExporter: Exports to XML format
//   - YAMLExporter: Exports to YAML format
//   - CSVExporter: Exports to CSV format (tabular, spreadsheet-friendly)
//
// # Features
//
//   - Format conversion: Convert between GEDCOM and other formats
//   - Header management: Automatically updates header metadata
//   - Continuation handling: Properly handles long lines with CONC/CONT
//   - Pretty printing: Human-readable output for JSON, XML, YAML
//
// # Usage Example
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/lesfleursdelanuitdev/gedcom-go/internal/exporter"
//		"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
//	)
//
//	func main() {
//		// Parse file
//		p := parser.NewHierarchicalParser()
//		tree, err := p.Parse("family.ged")
//		if err != nil {
//			panic(err)
//		}
//
//		// Export to JSON
//		jsonExporter := exporter.NewJsonExporter()
//		json, err := jsonExporter.ExportToString(tree)
//		if err != nil {
//			panic(err)
//		}
//		fmt.Println(json)
//
//		// Export to file
//		err = jsonExporter.ExportToFile(tree, "family.json")
//		if err != nil {
//			panic(err)
//		}
//	}
//
// # Round-trip Conversion
//
// You can convert between formats:
//
//	GEDCOM → JSON → XML → YAML → GEDCOM
//
// The exporters preserve the data structure and relationships.
//
// # Header Updates
//
// When exporting to GEDCOM format, the exporter automatically updates:
//
//   - GEDC.VERS: GEDCOM version
//   - CHAR: Character encoding
//   - SOUR: Source system
//   - DATE: Export date
//   - TIME: Export time
package exporter
