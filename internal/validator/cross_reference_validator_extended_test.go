package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestCrossReferenceValidator_ValidateIndividualReferences_FAMC(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with non-existent FAMC reference
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	famcLine := gedcom.NewGedcomLine(1, "FAMC", "@F999@", "") // Non-existent family
	indiLine.AddChild(nameLine)
	indiLine.AddChild(famcLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference: @F999@ in INDI record @I1@" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid FAMC cross-reference")
	}
}

func TestCrossReferenceValidator_ValidateIndividualReferences_MultipleFAMS(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with multiple FAMS references (one valid, one invalid)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	fams1 := gedcom.NewGedcomLine(1, "FAMS", "@F1@", "")
	fams2 := gedcom.NewGedcomLine(1, "FAMS", "@F999@", "") // Non-existent
	indiLine.AddChild(nameLine)
	indiLine.AddChild(fams1)
	indiLine.AddChild(fams2)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Create valid family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference: @F999@ in INDI record @I1@" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid FAMS cross-reference")
	}
}

func TestCrossReferenceValidator_ValidateFamilyReferences_WIFE(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with non-existent WIFE reference
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := gedcom.NewGedcomLine(1, "WIFE", "@I999@", "") // Non-existent
	famLine.AddChild(husbLine)
	famLine.AddChild(wifeLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Create valid husband
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

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
		t.Error("Expected error for invalid WIFE cross-reference")
	}
}

func TestCrossReferenceValidator_ValidateFamilyReferences_CHIL(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with non-existent CHIL reference
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	chil1 := gedcom.NewGedcomLine(1, "CHIL", "@I2@", "")      // Valid
	chil2 := gedcom.NewGedcomLine(1, "CHIL", "@I999@", "")   // Non-existent
	famLine.AddChild(husbLine)
	famLine.AddChild(chil1)
	famLine.AddChild(chil2)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Create valid husband and one child
	husbIndiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	husbNameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	husbIndiLine.AddChild(husbNameLine)
	husbIndiRecord := gedcom.NewIndividualRecord(husbIndiLine)
	tree.AddRecord(husbIndiRecord)

	chilIndiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	chilNameLine := gedcom.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	chilIndiLine.AddChild(chilNameLine)
	chilIndiRecord := gedcom.NewIndividualRecord(chilIndiLine)
	tree.AddRecord(chilIndiRecord)

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
		t.Error("Expected error for invalid CHIL cross-reference")
	}
}

func TestCrossReferenceValidator_ValidateXrefIDs_InvalidFormat(t *testing.T) {
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
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference ID: INVALID_XREF in INDI record" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid xref ID format")
	}
}

func TestCrossReferenceValidator_ValidateXrefIDs_FamilyInvalidFormat(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with invalid xref format
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "INVALID_XREF")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	famLine.AddChild(husbLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference ID: INVALID_XREF in FAM record" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid family xref ID format")
	}
}


