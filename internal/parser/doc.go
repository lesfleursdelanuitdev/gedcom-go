// Package parser provides functionality for parsing GEDCOM files.
//
// The parser implements a hierarchical parsing algorithm that builds a complete
// tree structure from GEDCOM files, handling nested levels, continuation lines
// (CONC/CONT), and various character encodings.
//
// # Parsers
//
// The package provides multiple parser implementations:
//
//   - HierarchicalParser: Full hierarchical parser with error handling
//   - BasicParser: Backward-compatible wrapper around HierarchicalParser
//   - ParallelHierarchicalParser: Experimental parallel parser
//   - TwoPhaseParser: Two-phase parser (sequential collection, parallel parsing)
//   - StreamingHierarchicalParser: Streaming parser for very large files (>100MB)
//
// # Features
//
//   - Hierarchical parsing: Builds complete parent-child relationships
//   - Encoding detection: Supports UTF-8, UTF-16 (with BOM), and ANSEL
//   - Continuation handling: Processes CONC (concatenate) and CONT (continue) lines
//   - Error recovery: Continues parsing after non-fatal errors
//   - Line validation: Validates line format and structure
//
// # Usage Example
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
//	)
//
//	func main() {
//		// Create parser
//		p := parser.NewHierarchicalParser()
//
//		// Parse file
//		tree, err := p.Parse("family.ged")
//		if err != nil {
//			panic(err)
//		}
//
//		// Check for errors
//		if p.HasErrors() {
//			errors := p.GetErrors()
//			for _, err := range errors {
//				fmt.Printf("Error: %s\n", err)
//			}
//		}
//
//		// Use tree
//		individuals := tree.GetAllIndividuals()
//		fmt.Printf("Found %d individuals\n", len(individuals))
//	}
//
// # Streaming Parser Example
//
// For very large files (>100MB), use the streaming parser:
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
//	)
//
//	func main() {
//		// Create streaming parser
//		sp := parser.NewStreamingHierarchicalParser()
//
//		// Parse with callback
//		err := sp.ParseWithHandler("large.ged", func(record gedcom.Record) error {
//			// Process record immediately without storing entire tree
//			fmt.Printf("Found record: %s\n", record.Type())
//			return nil // Continue parsing
//		})
//
//		// Or use iterator
//		iterator, err := parser.NewRecordIterator("large.ged")
//		if err != nil {
//			panic(err)
//		}
//		defer iterator.Close()
//
//		for iterator.Next() {
//			record := iterator.Record()
//			fmt.Printf("Found record: %s\n", record.Type())
//		}
//	}
//
// # Algorithm
//
// The parser uses a stack-based algorithm to build hierarchical relationships:
//
//  1. Parse each line to extract level, tag, value, and xref
//  2. For level 0 lines: Create record and add to tree
//  3. For level > 0 lines: Find parent using stack, add as child
//  4. Handle CONC/CONT continuation lines
//  5. Collect errors without stopping parsing
//
// # Error Handling
//
// The parser collects errors during parsing but continues processing. Errors
// are available via GetErrors() and can be filtered by severity.
package parser
