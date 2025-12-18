package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestValidator_Integration_RealFiles(t *testing.T) {
	testFiles := []string{
		"/apps/family-tree/flask-backend/gedcom/sample.ged",
		"/apps/family-tree/gedcom/gracis.ged",
		"/apps/family-tree/gedcom/xavier.ged",
		"/apps/family-tree/gedcom/tree1.ged",
	}

	for _, filePath := range testFiles {
		t.Run(filePath, func(t *testing.T) {
			// Parse the file
			hp := parser.NewHierarchicalParser()
			tree, err := hp.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filePath, err)
			}

			// Create validator with parser's error manager
			errorManager := hp.GetErrorManager()
			validator := NewGedcomValidator(errorManager)

			// Validate
			err = validator.Validate(tree)
			if err != nil {
				t.Logf("Validation returned error (may be expected): %v", err)
			}

			// Check errors
			errors := errorManager.Errors()
			severeCount := 0
			warningCount := 0

			for _, e := range errors {
				if e.Severity == gedcom.SeveritySevere {
					severeCount++
				} else {
					warningCount++
				}
			}

			t.Logf("File: %s", filePath)
			t.Logf("  Severe errors: %d", severeCount)
			t.Logf("  Warnings: %d", warningCount)
			t.Logf("  Total errors: %d", len(errors))

			// Log first few errors for debugging
			for i, e := range errors {
				if i >= 5 {
					break
				}
				t.Logf("  Error %d: %s", i+1, e.Error())
			}
		})
	}
}

func TestValidator_Integration_ValidTree(t *testing.T) {
	// Create a valid tree
	tree := gedcom.NewGedcomTree()

	// Create header with submitter
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	submLine := gedcom.NewGedcomLine(1, "SUBM", "@U1@", "")
	headerLine.AddChild(submLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create submitter
	submLine2 := gedcom.NewGedcomLine(0, "SUBM", "", "@U1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "Test Submitter", "")
	submLine2.AddChild(nameLine)
	submRecord := gedcom.NewSubmitterRecord(submLine2)
	tree.AddRecord(submRecord)

	// Create individual
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine2 := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	sexLine := gedcom.NewGedcomLine(1, "SEX", "M", "")
	indiLine.AddChild(nameLine2)
	indiLine.AddChild(sexLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Create another individual
	indiLine2 := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	nameLine3 := gedcom.NewGedcomLine(1, "NAME", "Jane /Doe/", "")
	sexLine2 := gedcom.NewGedcomLine(1, "SEX", "F", "")
	indiLine2.AddChild(nameLine3)
	indiLine2.AddChild(sexLine2)
	indiRecord2 := gedcom.NewIndividualRecord(indiLine2)
	tree.AddRecord(indiRecord2)

	// Create family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := gedcom.NewGedcomLine(1, "WIFE", "@I2@", "")
	famLine.AddChild(husbLine)
	famLine.AddChild(wifeLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Validate
	errorManager := gedcom.NewErrorManager()
	validator := NewGedcomValidator(errorManager)
	err := validator.Validate(tree)

	if err != nil {
		t.Logf("Validation error (may be expected): %v", err)
	}

	errors := errorManager.Errors()
	if len(errors) > 0 {
		t.Logf("Found %d validation errors:", len(errors))
		for i, e := range errors {
			if i >= 10 {
				t.Logf("  ... and %d more", len(errors)-10)
				break
			}
			t.Logf("  %s", e.Error())
		}
	} else {
		t.Log("No validation errors - tree is valid!")
	}
}

func TestValidator_Integration_InvalidTree(t *testing.T) {
	// Create an invalid tree (missing NAME, invalid SEX, invalid xref)
	tree := gedcom.NewGedcomTree()

	// Individual without NAME (required)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	sexLine := gedcom.NewGedcomLine(1, "SEX", "INVALID", "")
	indiLine.AddChild(sexLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Family with invalid HUSB reference
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I999@", "") // Non-existent
	famLine.AddChild(husbLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Validate
	errorManager := gedcom.NewErrorManager()
	validator := NewGedcomValidator(errorManager)
	validator.Validate(tree)

	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Error("Expected validation errors but found none")
	}

	// Check for specific errors
	foundMissingName := false
	foundInvalidSex := false
	foundInvalidXref := false

	for _, e := range errors {
		if e.Message == "INDI @I1@: Missing required tag NAME" {
			foundMissingName = true
		}
		if e.Message == "INDI @I1@: Invalid SEX value INVALID" {
			foundInvalidSex = true
		}
		if e.Message == "Invalid cross-reference: @I999@ in FAM record @F1@" {
			foundInvalidXref = true
		}
	}

	if !foundMissingName {
		t.Error("Expected error for missing NAME tag")
	}
	if !foundInvalidSex {
		t.Error("Expected error for invalid SEX value")
	}
	if !foundInvalidXref {
		t.Error("Expected error for invalid cross-reference")
	}
}

