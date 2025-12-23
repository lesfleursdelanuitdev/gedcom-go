package exporter

import (
	"os"
	"strings"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestCSVExporter_ExportToFile(t *testing.T) {
	// Parse a test file
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse("../../testdata/royal92.ged")
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	// Create CSV exporter
	errorManager := gedcom.NewErrorManager()
	csvExporter := NewCSVExporter(errorManager)

	// Export to temporary file
	tmpFile := "/tmp/test_export.csv"
	defer os.Remove(tmpFile)

	err = csvExporter.ExportToFile(tree, tmpFile)
	if err != nil {
		t.Fatalf("Failed to export to CSV: %v", err)
	}

	// Verify file exists and has content
	fileInfo, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("CSV file was not created: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("CSV file is empty")
	}

	// Read and verify header
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		t.Error("CSV file should have at least header and one data row")
	}

	// Check header
	header := lines[0]
	expectedColumns := []string{"XREF", "Type", "Name", "Sex", "Birth Date", "Birth Place"}
	for _, col := range expectedColumns {
		if !strings.Contains(header, col) {
			t.Errorf("CSV header missing column: %s", col)
		}
	}
}

func TestCSVExporter_ExportToString(t *testing.T) {
	// Parse a test file
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse("../../testdata/royal92.ged")
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	// Create CSV exporter
	errorManager := gedcom.NewErrorManager()
	csvExporter := NewCSVExporter(errorManager)

	// Export to string
	csvString, err := csvExporter.ExportToString(tree)
	if err != nil {
		t.Fatalf("Failed to export to CSV string: %v", err)
	}

	if csvString == "" {
		t.Error("CSV string is empty")
	}

	// Verify header
	lines := strings.Split(csvString, "\n")
	if len(lines) < 2 {
		t.Error("CSV string should have at least header and one data row")
	}

	header := lines[0]
	expectedColumns := []string{"XREF", "Type", "Name"}
	for _, col := range expectedColumns {
		if !strings.Contains(header, col) {
			t.Errorf("CSV header missing column: %s", col)
		}
	}
}

