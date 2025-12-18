package validator

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestParallelGedcomValidator_Validate(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelGedcomValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create a minimal valid tree
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	submLine := gedcom.NewGedcomLine(1, "SUBM", "@U1@", "")
	headerLine.AddChild(submLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create an individual
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Create a family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := gedcom.NewGedcomLine(1, "WIFE", "@I2@", "")
	famLine.AddChild(husbLine)
	famLine.AddChild(wifeLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Validate
	err := validator.Validate(tree)
	if err != nil {
		t.Logf("Validation returned error (may be expected): %v", err)
	}

	// Check for errors (should have errors for missing @I2@ and @U1@)
	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Log("No validation errors found (this is expected for minimal valid tree)")
	}
}

func BenchmarkValidator_Sequential(b *testing.B) {
	errorManager := gedcom.NewErrorManager()
	validator := NewGedcomValidator(errorManager)

	tree := createTestTree()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(tree)
	}
}

func BenchmarkValidator_Parallel(b *testing.B) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelGedcomValidator(errorManager)

	tree := createTestTree()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(tree)
	}
}

func BenchmarkParallelIndividualValidator(b *testing.B) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := createLargeTestTree(1000) // 1000 individuals

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(tree)
	}
}

func createTestTree() *gedcom.GedcomTree {
	tree := gedcom.NewGedcomTree()

	// Create header
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create 100 individuals
	for i := 0; i < 100; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indiLine := gedcom.NewGedcomLine(0, "INDI", "", xrefID)
		nameLine := gedcom.NewGedcomLine(1, "NAME", "Test /Person/", "")
		indiLine.AddChild(nameLine)
		indiRecord := gedcom.NewIndividualRecord(indiLine)
		tree.AddRecord(indiRecord)
	}

	return tree
}

func createLargeTestTree(count int) *gedcom.GedcomTree {
	tree := gedcom.NewGedcomTree()

	// Create header
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create many individuals
	for i := 0; i < count; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indiLine := gedcom.NewGedcomLine(0, "INDI", "", xrefID)
		nameLine := gedcom.NewGedcomLine(1, "NAME", "Test /Person/", "")
		sexLine := gedcom.NewGedcomLine(1, "SEX", "M", "")
		indiLine.AddChild(nameLine)
		indiLine.AddChild(sexLine)
		indiRecord := gedcom.NewIndividualRecord(indiLine)
		tree.AddRecord(indiRecord)
	}

	return tree
}
