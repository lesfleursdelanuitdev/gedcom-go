// Package gedcom provides core data structures and types for working with GEDCOM files.
//
// GEDCOM (Genealogical Data Communication) is a specification for exchanging
// genealogical data between different software systems. This package implements
// the core data structures defined in the GEDCOM 5.5.1 specification.
//
// # Core Types
//
// The package provides several core types:
//
//   - GedcomTree: The root container for all parsed GEDCOM records
//   - GedcomLine: Represents a single line in a GEDCOM file with hierarchical structure
//   - Record: Interface for all GEDCOM record types (Individual, Family, Note, etc.)
//
// # Record Types
//
// The package supports the following specialized record types:
//
//   - IndividualRecord: Represents an individual person (INDI)
//   - FamilyRecord: Represents a family unit (FAM)
//   - HeaderRecord: Represents the file header (HEAD)
//   - NoteRecord: Represents a note (NOTE)
//   - SourceRecord: Represents a source citation (SOUR)
//   - RepositoryRecord: Represents a repository (REPO)
//   - SubmitterRecord: Represents a submitter (SUBM)
//   - MultimediaRecord: Represents a multimedia object (OBJE)
//
// # Usage Example
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
//		"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
//	)
//
//	func main() {
//		// Parse a GEDCOM file
//		p := parser.NewHierarchicalParser()
//		tree, err := p.Parse("family.ged")
//		if err != nil {
//			panic(err)
//		}
//
//		// Access individuals
//		individuals := tree.GetAllIndividuals()
//		for xrefID, record := range individuals {
//			indi := record.(*gedcom.IndividualRecord)
//			name := indi.GetName()
//			fmt.Printf("%s: %s\n", xrefID, name)
//		}
//	}
//
// # Thread Safety
//
// GedcomTree is thread-safe and uses sync.RWMutex for concurrent access.
// All public methods are safe to call from multiple goroutines.
//
// # Error Handling
//
// The package uses a centralized ErrorManager for collecting and reporting
// errors during parsing and validation. Errors are categorized by severity
// (Warning, Severe) and include context information like line numbers.
package gedcom
