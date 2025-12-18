package exporter

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestJsonExporter_ExportToString(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewJsonExporter(errorManager)

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
	sexLine := gedcom.NewGedcomLine(1, "SEX", "M", "")
	indiLine.AddChild(nameLine2)
	indiLine.AddChild(sexLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Export
	result, err := exporter.ExportToString(tree)
	if err != nil {
		t.Fatalf("ExportToString failed: %v", err)
	}

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Exported string is not valid JSON: %v", err)
	}

	// Verify structure
	if jsonData["header"] == nil {
		t.Error("Missing header in JSON")
	}
	if jsonData["individuals"] == nil {
		t.Error("Missing individuals in JSON")
	}
	if jsonData["families"] == nil {
		t.Error("Missing families in JSON")
	}
}

func TestJsonExporter_ExportToFile(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewJsonExporter(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create minimal tree
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	submLine := gedcom.NewGedcomLine(0, "SUBM", "", "@U1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "Test", "")
	submLine.AddChild(nameLine)
	submRecord := gedcom.NewSubmitterRecord(submLine)
	tree.AddRecord(submRecord)

	// Export to file
	tmpFile := "/tmp/test_export.json"
	defer os.Remove(tmpFile)

	err := exporter.ExportToFile(tree, tmpFile)
	if err != nil {
		t.Fatalf("ExportToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Read and verify JSON
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		t.Fatalf("Exported file is not valid JSON: %v", err)
	}
}

func TestJsonExporter_IndividualToJSON(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewJsonExporter(errorManager)

	// Create individual with name, birth, death
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	givnLine := gedcom.NewGedcomLine(2, "GIVN", "John", "")
	surnLine := gedcom.NewGedcomLine(2, "SURN", "Doe", "")
	nameLine.AddChild(givnLine)
	nameLine.AddChild(surnLine)
	indiLine.AddChild(nameLine)

	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	dateLine := gedcom.NewGedcomLine(2, "DATE", "1 Jan 1900", "")
	placLine := gedcom.NewGedcomLine(2, "PLAC", "New York", "")
	birtLine.AddChild(dateLine)
	birtLine.AddChild(placLine)
	indiLine.AddChild(birtLine)

	indiRecord := gedcom.NewIndividualRecord(indiLine)

	jsonData := exporter.individualToJSON(indiRecord)

	if jsonData["id"] != "@I1@" {
		t.Errorf("Expected id @I1@, got %v", jsonData["id"])
	}

	names, ok := jsonData["names"].([]map[string]interface{})
	if !ok || len(names) == 0 {
		t.Error("Expected names array")
	} else {
		if names[0]["full"] != "John /Doe/" {
			t.Errorf("Expected full name 'John /Doe/', got %v", names[0]["full"])
		}
	}

	birth, ok := jsonData["birth"].(map[string]interface{})
	if !ok {
		t.Error("Expected birth event")
	} else {
		if birth["date"] != "1 Jan 1900" {
			t.Errorf("Expected birth date '1 Jan 1900', got %v", birth["date"])
		}
	}
}

func TestJsonExporter_FamilyToJSON(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewJsonExporter(errorManager)

	// Create family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := gedcom.NewGedcomLine(1, "WIFE", "@I2@", "")
	chilLine := gedcom.NewGedcomLine(1, "CHIL", "@I3@", "")
	famLine.AddChild(husbLine)
	famLine.AddChild(wifeLine)
	famLine.AddChild(chilLine)

	famRecord := gedcom.NewFamilyRecord(famLine)

	jsonData := exporter.familyToJSON(famRecord)

	if jsonData["id"] != "@F1@" {
		t.Errorf("Expected id @F1@, got %v", jsonData["id"])
	}
	if jsonData["husband"] != "@I1@" {
		t.Errorf("Expected husband @I1@, got %v", jsonData["husband"])
	}
	if jsonData["wife"] != "@I2@" {
		t.Errorf("Expected wife @I2@, got %v", jsonData["wife"])
	}

	children, ok := jsonData["children"].([]string)
	if !ok || len(children) != 1 || children[0] != "@I3@" {
		t.Errorf("Expected children [@I3@], got %v", children)
	}
}


