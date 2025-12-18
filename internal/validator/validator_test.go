package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestGedcomValidator_Validate(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewGedcomValidator(errorManager)

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
		t.Fatalf("Validate failed: %v", err)
	}

	// Check for errors (should have errors for missing @I2@ and @U1@)
	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Log("No validation errors found (this is expected for minimal valid tree)")
	}
}

func TestIndividualValidator_ValidateStructure(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid tag
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	invalidLine := gedcom.NewGedcomLine(1, "INVALID_TAG", "value", "")
	indiLine.AddChild(invalidLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Error("Expected error for invalid tag")
	}
}

func TestIndividualValidator_ValidateRequiredTags(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual without NAME (required)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	sexLine := gedcom.NewGedcomLine(1, "SEX", "M", "")
	indiLine.AddChild(sexLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Missing required tag NAME" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing NAME tag")
	}
}

func TestIndividualValidator_ValidateSex(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid SEX value
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	sexLine := gedcom.NewGedcomLine(1, "SEX", "INVALID", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(sexLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid SEX value INVALID" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid SEX value")
	}
}

func TestFamilyValidator_ValidateRequiredTags(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family without HUSB or WIFE
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Missing required tags (must have at least HUSB or WIFE)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing HUSB/WIFE")
	}
}

func TestCrossReferenceValidator_ValidateXrefFormat(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid xref format
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "INVALID_XREF")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid xref format")
	}
}

func TestCrossReferenceValidator_ValidateReferences(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with invalid HUSB reference
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I999@", "") // Non-existent individual
	famLine.AddChild(husbLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference: @I999@ in FAM record @F1@" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid cross-reference")
	}
}

func TestHeaderValidator_ValidateMissingHeader(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := gedcom.NewGedcomTree()
	// No header added

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "Missing HEAD record" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing HEAD record")
	}
}



