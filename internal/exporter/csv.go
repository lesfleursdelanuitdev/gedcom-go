package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// CSVExporter exports GEDCOM data to CSV format
type CSVExporter struct {
	*BaseExporter
}

// NewCSVExporter creates a new CSV exporter
func NewCSVExporter(errorManager *gedcom.ErrorManager) *CSVExporter {
	return &CSVExporter{
		BaseExporter: NewBaseExporter(errorManager),
	}
}

// ExportToFile exports the tree to a CSV file
func (ce *CSVExporter) ExportToFile(tree *gedcom.GedcomTree, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"XREF", "Type", "Name", "Sex", "Birth Date", "Birth Place",
		"Death Date", "Death Place", "Father XREF", "Mother XREF",
		"Spouse XREFs", "Children XREFs", "Notes",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Export individuals
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		row := ce.individualToCSVRow(xrefID, indi, tree)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportToString exports the tree to a CSV string
func (ce *CSVExporter) ExportToString(tree *gedcom.GedcomTree) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)
	defer writer.Flush()

	// Write header
	header := []string{
		"XREF", "Type", "Name", "Sex", "Birth Date", "Birth Place",
		"Death Date", "Death Place", "Father XREF", "Mother XREF",
		"Spouse XREFs", "Children XREFs", "Notes",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Export individuals
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		row := ce.individualToCSVRow(xrefID, indi, tree)
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	return sb.String(), nil
}

// individualToCSVRow converts an individual record to a CSV row
func (ce *CSVExporter) individualToCSVRow(xrefID string, indi *gedcom.IndividualRecord, tree *gedcom.GedcomTree) []string {
	// Get basic info
	name := indi.GetName()
	sex := indi.GetSex()
	birthDate := indi.GetBirthDate()
	birthPlace := indi.GetBirthPlace()
	deathDate := indi.GetDeathDate()
	deathPlace := indi.GetDeathPlace()

	// Get family relationships
	fatherXref := ""
	motherXref := ""
	spouseXrefs := []string{}
	childrenXrefs := []string{}

	// Find families where this individual is a child
	families := tree.GetAllFamilies()
	for famXref, famRecord := range families {
		fam, ok := famRecord.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		// Check if this individual is a child in this family
		children := fam.GetChildren()
		for _, childXref := range children {
			if childXref == xrefID {
				// Found family where this individual is a child
				husband := fam.GetHusband()
				wife := fam.GetWife()
				if husband != "" {
					fatherXref = husband
				}
				if wife != "" {
					motherXref = wife
				}
				break
			}
		}

		// Check if this individual is a spouse in this family
		husband := fam.GetHusband()
		wife := fam.GetWife()
		if husband == xrefID || wife == xrefID {
			spouseXrefs = append(spouseXrefs, famXref)
			// Get children from this family
			children := fam.GetChildren()
			childrenXrefs = append(childrenXrefs, children...)
		}
	}

	// Get notes (from NOTE records referenced by this individual)
	notes := []string{}
	// Note: Individual records may reference NOTE records via XREF
	// For CSV export, we'll collect note XREFs
	noteXrefs := indi.GetValues("NOTE")
	if len(noteXrefs) > 0 {
		notes = noteXrefs
	}

	// Build row
	row := []string{
		xrefID,
		"INDI",
		name,
		sex,
		birthDate,
		birthPlace,
		deathDate,
		deathPlace,
		fatherXref,
		motherXref,
		strings.Join(spouseXrefs, ";"),
		strings.Join(childrenXrefs, ";"),
		strings.Join(notes, " | "),
	}

	return row
}

