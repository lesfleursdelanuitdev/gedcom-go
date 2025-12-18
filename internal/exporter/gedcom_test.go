package exporter

import (
	"os"
	"strings"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestGedcomExporter_ExportToString(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewGedcomExporter(errorManager, "TestApp", "1.0.0")

	tree := gedcom.NewGedcomTree()

	// Create header
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	gedcLine := gedcom.NewGedcomLine(1, "GEDC", "", "")
	versLine := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")
	gedcLine.AddChild(versLine)
	headerLine.AddChild(gedcLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create submitter
	submLine := gedcom.NewGedcomLine(0, "SUBM", "", "@U1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "Test Submitter", "")
	submLine.AddChild(nameLine)
	submRecord := gedcom.NewSubmitterRecord(submLine)
	tree.AddRecord(submRecord)

	// Create individual
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine2 := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine2)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Export
	result, err := exporter.ExportToString(tree)
	if err != nil {
		t.Fatalf("ExportToString failed: %v", err)
	}

	// Verify structure
	if !strings.Contains(result, "0 HEAD") {
		t.Error("Missing HEAD record")
	}
	if !strings.Contains(result, "0 @U1@ SUBM") {
		t.Error("Missing SUBM record")
	}
	if !strings.Contains(result, "0 @I1@ INDI") {
		t.Error("Missing INDI record")
	}
	if !strings.Contains(result, "1 NAME John /Doe/") {
		t.Error("Missing NAME line")
	}
	if !strings.Contains(result, "0 TRLR") {
		t.Error("Missing TRLR")
	}
}

func TestGedcomExporter_ExportToFile(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewGedcomExporter(errorManager, "TestApp", "1.0.0")

	tree := gedcom.NewGedcomTree()

	// Create minimal valid tree
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	gedcLine := gedcom.NewGedcomLine(1, "GEDC", "", "")
	versLine := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")
	gedcLine.AddChild(versLine)
	headerLine.AddChild(gedcLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	submLine := gedcom.NewGedcomLine(0, "SUBM", "", "@U1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "Test", "")
	submLine.AddChild(nameLine)
	submRecord := gedcom.NewSubmitterRecord(submLine)
	tree.AddRecord(submRecord)

	// Export to file
	tmpFile := "/tmp/test_export.ged"
	defer os.Remove(tmpFile)

	err := exporter.ExportToFile(tree, tmpFile)
	if err != nil {
		t.Fatalf("ExportToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "0 HEAD") {
		t.Error("Exported file missing HEAD")
	}
	if !strings.Contains(contentStr, "0 TRLR") {
		t.Error("Exported file missing TRLR")
	}

	// Verify header was updated
	if !strings.Contains(contentStr, "SOUR TestApp") {
		t.Error("Header should contain SOUR with app name")
	}
	if !strings.Contains(contentStr, "CHAR UTF-8") {
		t.Error("Header should contain CHAR UTF-8")
	}
}

func TestGedcomExporter_UpdateHeader(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewGedcomExporter(errorManager, "MyApp", "2.0.0")

	tree := gedcom.NewGedcomTree()

	// Create header without metadata
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Update header
	err := exporter.updateHeader(tree, "/tmp/test.ged")
	if err != nil {
		t.Fatalf("updateHeader failed: %v", err)
	}

	// Verify header was updated
	header := tree.GetHeader()
	if header == nil {
		t.Fatal("Header should exist")
	}

	// Check GEDC.VERS
	vers := header.GetValue("GEDC.VERS")
	if vers != "5.5.5" {
		t.Errorf("Expected GEDC.VERS to be '5.5.5', got '%s'", vers)
	}

	// Check CHAR
	char := header.GetValue("CHAR")
	if char != "UTF-8" {
		t.Errorf("Expected CHAR to be 'UTF-8', got '%s'", char)
	}

	// Check SOUR
	sour := header.GetValue("SOUR")
	if sour != "MyApp" {
		t.Errorf("Expected SOUR to be 'MyApp', got '%s'", sour)
	}

	// Check SOUR.VERS
	sourVers := header.GetValue("SOUR.VERS")
	if sourVers != "2.0.0" {
		t.Errorf("Expected SOUR.VERS to be '2.0.0', got '%s'", sourVers)
	}

	// Check FILE
	file := header.GetValue("FILE")
	if file != "test.ged" {
		t.Errorf("Expected FILE to be 'test.ged', got '%s'", file)
	}

	// Check DATE exists
	date := header.GetValue("DATE")
	if date == "" {
		t.Error("Expected DATE to be set")
	}

	// Check TIME exists
	time := header.GetValue("TIME")
	if time == "" {
		t.Error("Expected TIME to be set")
	}
}

func TestGedcomExporter_SplitLongLine(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewGedcomExporter(errorManager, "TestApp", "1.0.0")

	// Create a line with a very long value
	longValue := strings.Repeat("A", 500)
	line := gedcom.NewGedcomLine(1, "NOTE", longValue, "")

	lines := exporter.splitLongLine(line)

	// Should have multiple lines
	if len(lines) < 2 {
		t.Errorf("Expected multiple lines for long value, got %d", len(lines))
	}

	// First line should contain the tag
	if !strings.Contains(lines[0], "1 NOTE") {
		t.Error("First line should contain tag")
	}

	// Should have CONC continuation lines
	foundCONC := false
	for _, l := range lines[1:] {
		if strings.Contains(l, "CONC") {
			foundCONC = true
			break
		}
	}
	if !foundCONC {
		t.Error("Should have CONC continuation lines")
	}
}

func TestGedcomExporter_FormatGEDLine(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewGedcomExporter(errorManager, "TestApp", "1.0.0")

	tests := []struct {
		name     string
		line     *gedcom.GedcomLine
		expected string
	}{
		{
			name:     "Level 0 with xref and value",
			line:     gedcom.NewGedcomLine(0, "INDI", "", "@I1@"),
			expected: "0 @I1@ INDI",
		},
		{
			name:     "Level 1 with value",
			line:     gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""),
			expected: "1 NAME John /Doe/",
		},
		{
			name:     "Level 2 with value",
			line:     gedcom.NewGedcomLine(2, "DATE", "1 Jan 1900", ""),
			expected: "2 DATE 1 Jan 1900",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatGEDLine(tt.line)
			if result != tt.expected {
				t.Errorf("formatGEDLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}
