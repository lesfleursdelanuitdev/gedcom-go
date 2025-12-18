package exporter

import (
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"gopkg.in/yaml.v3"
)

func TestYAMLExporter_ExportToString(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewYAMLExporter(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create header
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create individual
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Export
	result, err := exporter.ExportToString(tree)
	if err != nil {
		t.Fatalf("ExportToString failed: %v", err)
	}

	// Verify it's valid YAML
	var yamlData YAMLGedcom
	if err := yaml.Unmarshal([]byte(result), &yamlData); err != nil {
		t.Fatalf("Exported string is not valid YAML: %v", err)
	}

	// Verify structure
	if yamlData.Version != "5.5.5" {
		t.Errorf("Expected version 5.5.5, got %s", yamlData.Version)
	}
	if yamlData.Individuals == nil {
		t.Error("Expected individuals to be present")
	}
}

func TestYAMLExporter_ExportToFile(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	exporter := NewYAMLExporter(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create minimal tree
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Export to file
	tmpFile := "/tmp/test_export.yaml"
	defer os.Remove(tmpFile)

	err := exporter.ExportToFile(tree, tmpFile)
	if err != nil {
		t.Fatalf("ExportToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Read and verify YAML
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	var yamlData YAMLGedcom
	if err := yaml.Unmarshal(content, &yamlData); err != nil {
		t.Fatalf("Exported file is not valid YAML: %v", err)
	}
}
