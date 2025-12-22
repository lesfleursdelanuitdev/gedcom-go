package exporter

import (
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestExporter_Integration_ParseExportParse(t *testing.T) {
	testFiles := []string{
		"/apps/family-tree/flask-backend/gedcom/sample.ged",
		"/apps/family-tree/gedcom/gracis.ged",
		"/apps/family-tree/gedcom/xavier.ged",
	}

	for _, filePath := range testFiles {
		t.Run(filePath, func(t *testing.T) {
			// Skip if file doesn't exist (these are optional external test files)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Skipping test - file does not exist: %s", filePath)
			}

			// Parse original file
			hp := parser.NewHierarchicalParser()
			tree, err := hp.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filePath, err)
			}

			// Export to GEDCOM
			errorManager := gedcom.NewErrorManager()
			gedExporter := NewGedcomExporter(errorManager, "TestApp", "1.0.0")
			
			tmpFile := "/tmp/test_roundtrip.ged"
			defer os.Remove(tmpFile)

			err = gedExporter.ExportToFile(tree, tmpFile)
			if err != nil {
				t.Fatalf("Failed to export %s: %v", filePath, err)
			}

			// Parse exported file
			hp2 := parser.NewHierarchicalParser()
			tree2, err := hp2.Parse(tmpFile)
			if err != nil {
				t.Fatalf("Failed to parse exported file: %v", err)
			}

			// Verify structure is preserved
			originalIndis := tree.GetAllIndividuals()
			exportedIndis := tree2.GetAllIndividuals()

			if len(originalIndis) != len(exportedIndis) {
				t.Errorf("Individual count mismatch: original=%d, exported=%d",
					len(originalIndis), len(exportedIndis))
			}

			originalFams := tree.GetAllFamilies()
			exportedFams := tree2.GetAllFamilies()

			if len(originalFams) != len(exportedFams) {
				t.Errorf("Family count mismatch: original=%d, exported=%d",
					len(originalFams), len(exportedFams))
			}
		})
	}
}

func TestExporter_Integration_JSONExport(t *testing.T) {
	testFiles := []string{
		"/apps/family-tree/flask-backend/gedcom/sample.ged",
		"/apps/family-tree/gedcom/gracis.ged",
	}

	for _, filePath := range testFiles {
		t.Run(filePath, func(t *testing.T) {
			// Skip if file doesn't exist (these are optional external test files)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Skipping test - file does not exist: %s", filePath)
			}

			// Parse file
			hp := parser.NewHierarchicalParser()
			tree, err := hp.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filePath, err)
			}

			// Export to JSON
			errorManager := gedcom.NewErrorManager()
			jsonExporter := NewJsonExporter(errorManager)

			tmpFile := "/tmp/test_export.json"
			defer os.Remove(tmpFile)

			err = jsonExporter.ExportToFile(tree, tmpFile)
			if err != nil {
				t.Fatalf("Failed to export to JSON: %v", err)
			}

			// Verify file exists and is valid JSON
			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read JSON file: %v", err)
			}

			// Basic validation - file should not be empty
			if len(content) == 0 {
				t.Error("Exported JSON file is empty")
			}

			// Verify it contains expected keys
			contentStr := string(content)
			expectedKeys := []string{"header", "individuals", "families", "metadata"}
			for _, key := range expectedKeys {
				if !contains(contentStr, `"`+key+`"`) {
					t.Errorf("Exported JSON missing key: %s", key)
				}
			}
		})
	}
}

func TestExporter_Integration_ComplexTree(t *testing.T) {
	// Create a complex tree with multiple record types
	tree := gedcom.NewGedcomTree()

	// Header
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	gedcLine := gedcom.NewGedcomLine(1, "GEDC", "", "")
	versLine := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")
	gedcLine.AddChild(versLine)
	headerLine.AddChild(gedcLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Submitter
	submLine := gedcom.NewGedcomLine(0, "SUBM", "", "@U1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "Test Submitter", "")
	submLine.AddChild(nameLine)
	submRecord := gedcom.NewSubmitterRecord(submLine)
	tree.AddRecord(submRecord)

	// Individual with events
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine2 := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	dateLine := gedcom.NewGedcomLine(2, "DATE", "1 Jan 1900", "")
	birtLine.AddChild(dateLine)
	indiLine.AddChild(nameLine2)
	indiLine.AddChild(birtLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	famLine.AddChild(husbLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Export to GEDCOM
	errorManager := gedcom.NewErrorManager()
	gedExporter := NewGedcomExporter(errorManager, "TestApp", "1.0.0")
	
	gedStr, err := gedExporter.ExportToString(tree)
	if err != nil {
		t.Fatalf("Failed to export: %v", err)
	}

	// Verify all record types are present
	if !contains(gedStr, "0 HEAD") {
		t.Error("Missing HEAD in export")
	}
	if !contains(gedStr, "0 @U1@ SUBM") {
		t.Error("Missing SUBM in export")
	}
	if !contains(gedStr, "0 @I1@ INDI") {
		t.Error("Missing INDI in export")
	}
	if !contains(gedStr, "0 @F1@ FAM") {
		t.Error("Missing FAM in export")
	}
	if !contains(gedStr, "0 TRLR") {
		t.Error("Missing TRLR in export")
	}

	// Export to JSON
	jsonExporter := NewJsonExporter(errorManager)
	jsonStr, err := jsonExporter.ExportToString(tree)
	if err != nil {
		t.Fatalf("Failed to export to JSON: %v", err)
	}

	// Verify JSON structure
	if !contains(jsonStr, `"header"`) {
		t.Error("Missing header in JSON")
	}
	if !contains(jsonStr, `"individuals"`) {
		t.Error("Missing individuals in JSON")
	}
	if !contains(jsonStr, `"families"`) {
		t.Error("Missing families in JSON")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

