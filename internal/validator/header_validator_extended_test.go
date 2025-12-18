package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestHeaderValidator_ValidateGedc_MissingGEDC(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create header without GEDC
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "HEAD: Missing GEDC tag" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing GEDC tag")
	}
}

func TestHeaderValidator_ValidateGedc_MissingVERS(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create header with GEDC but without VERS
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	gedcLine := gedcom.NewGedcomLine(1, "GEDC", "", "")
	headerLine.AddChild(gedcLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "HEAD: Missing GEDC.VERS" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for missing GEDC.VERS")
	}
}

func TestHeaderValidator_ValidateStructure_InvalidTag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create header with invalid tag
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	invalidLine := gedcom.NewGedcomLine(1, "INVALID_TAG", "value", "")
	headerLine.AddChild(invalidLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "HEAD: Invalid tag INVALID_TAG" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for invalid header tag")
	}
}

func TestHeaderValidator_ValidateStructure_UserDefinedTag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create header with user-defined tag (should not error)
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	userTag := gedcom.NewGedcomLine(1, "_CUSTOM", "value", "")
	gedcLine := gedcom.NewGedcomLine(1, "GEDC", "", "")
	versLine := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")
	gedcLine.AddChild(versLine)
	headerLine.AddChild(userTag)
	headerLine.AddChild(gedcLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "HEAD: Invalid tag _CUSTOM" {
			t.Error("Should not error for user-defined tags")
		}
	}
}

func TestHeaderValidator_ValidateGedc_ValidStructure(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create header with valid GEDC structure
	headerLine := gedcom.NewGedcomLine(0, "HEAD", "", "")
	gedcLine := gedcom.NewGedcomLine(1, "GEDC", "", "")
	versLine := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")
	formLine := gedcom.NewGedcomLine(2, "FORM", "LINEAGE-LINKED", "")
	gedcLine.AddChild(versLine)
	gedcLine.AddChild(formLine)
	headerLine.AddChild(gedcLine)
	headerRecord := gedcom.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Header Validation" && (err.Message == "HEAD: Missing GEDC tag" || err.Message == "HEAD: Missing GEDC.VERS") {
			t.Error("Should not error for valid GEDC structure")
		}
	}
}


