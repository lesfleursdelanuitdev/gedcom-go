package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestFamilyValidator_ValidateReferences_WIFE(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with invalid WIFE reference format
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := gedcom.NewGedcomLine(1, "WIFE", "INVALID_FORMAT", "")
	famLine.AddChild(husbLine)
	famLine.AddChild(wifeLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Invalid WIFE reference format INVALID_FORMAT" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid WIFE reference format")
	}
}

func TestFamilyValidator_ValidateReferences_CHIL(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with invalid CHIL reference format
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	chil1 := gedcom.NewGedcomLine(1, "CHIL", "@I2@", "")
	chil2 := gedcom.NewGedcomLine(1, "CHIL", "INVALID_FORMAT", "")
	famLine.AddChild(husbLine)
	famLine.AddChild(chil1)
	famLine.AddChild(chil2)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Invalid CHIL reference format INVALID_FORMAT" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid CHIL reference format")
	}
}

func TestFamilyValidator_ValidateEvents_MultipleMarriage(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with multiple MARR events
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	marr1 := gedcom.NewGedcomLine(1, "MARR", "", "")
	marr2 := gedcom.NewGedcomLine(1, "MARR", "", "")
	famLine.AddChild(husbLine)
	famLine.AddChild(marr1)
	famLine.AddChild(marr2)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Multiple MARR events" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for multiple MARR events")
	}
}

func TestFamilyValidator_ValidateEvents_MultipleDivorce(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with multiple DIV events
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	div1 := gedcom.NewGedcomLine(1, "DIV", "", "")
	div2 := gedcom.NewGedcomLine(1, "DIV", "", "")
	famLine.AddChild(husbLine)
	famLine.AddChild(div1)
	famLine.AddChild(div2)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Multiple DIV events" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for multiple DIV events")
	}
}

func TestFamilyValidator_ValidateEventStructure_InvalidSubtag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with invalid event subtag
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	marrLine := gedcom.NewGedcomLine(1, "MARR", "", "")
	invalidLine := gedcom.NewGedcomLine(2, "INVALID_SUBTAG", "value", "")
	marrLine.AddChild(invalidLine)
	famLine.AddChild(husbLine)
	famLine.AddChild(marrLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Invalid subtag INVALID_SUBTAG in MARR event" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for invalid event subtag")
	}
}

func TestFamilyValidator_ValidateEvents_AllEventTypes(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with various event types
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	
	// Add multiple event types
	engaLine := gedcom.NewGedcomLine(1, "ENGA", "", "")
	anulLine := gedcom.NewGedcomLine(1, "ANUL", "", "")
	censLine := gedcom.NewGedcomLine(1, "CENS", "", "")
	
	famLine.AddChild(husbLine)
	famLine.AddChild(engaLine)
	famLine.AddChild(anulLine)
	famLine.AddChild(censLine)
	
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	// Should validate without errors for valid event types
	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Invalid tag ENGA" {
			t.Error("ENGA is a valid event tag")
		}
	}
}

func TestFamilyValidator_ValidateStructure_OnlyHusband(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with only HUSB (should be valid)
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	famLine.AddChild(husbLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Missing required tags (must have at least HUSB or WIFE)" {
			t.Error("Family with only HUSB should be valid")
		}
	}
}

func TestFamilyValidator_ValidateStructure_OnlyWife(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create family with only WIFE (should be valid)
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	wifeLine := gedcom.NewGedcomLine(1, "WIFE", "@I1@", "")
	famLine.AddChild(wifeLine)
	famRecord := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Missing required tags (must have at least HUSB or WIFE)" {
			t.Error("Family with only WIFE should be valid")
		}
	}
}


