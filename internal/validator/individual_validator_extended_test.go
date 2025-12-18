package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestIndividualValidator_ValidateReferences_FAMC(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid FAMC reference format
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	famcLine := gedcom.NewGedcomLine(1, "FAMC", "INVALID_FORMAT", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(famcLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid FAMC reference format INVALID_FORMAT" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid FAMC reference format")
	}
}

func TestIndividualValidator_ValidateReferences_MultipleFAMS(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with multiple FAMS references
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	fams1 := gedcom.NewGedcomLine(1, "FAMS", "@F1@", "")
	fams2 := gedcom.NewGedcomLine(1, "FAMS", "@F2@", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(fams1)
	indiLine.AddChild(fams2)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	// Should not error for multiple FAMS (valid)
	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid FAMS reference format @F1@" {
			t.Error("Should not error for valid FAMS format")
		}
	}
}

func TestIndividualValidator_ValidateEvents_MultipleBirth(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with multiple BIRT events
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	birt1 := gedcom.NewGedcomLine(1, "BIRT", "", "")
	birt2 := gedcom.NewGedcomLine(1, "BIRT", "", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(birt1)
	indiLine.AddChild(birt2)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Multiple BIRT events" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for multiple BIRT events")
	}
}

func TestIndividualValidator_ValidateEvents_MultipleDeath(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with multiple DEAT events
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	deat1 := gedcom.NewGedcomLine(1, "DEAT", "", "")
	deat2 := gedcom.NewGedcomLine(1, "DEAT", "", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(deat1)
	indiLine.AddChild(deat2)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Multiple DEAT events" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for multiple DEAT events")
	}
}

func TestIndividualValidator_ValidateEventStructure_InvalidSubtag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid event subtag
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	invalidLine := gedcom.NewGedcomLine(2, "INVALID_SUBTAG", "value", "")
	birtLine.AddChild(invalidLine)
	indiLine.AddChild(nameLine)
	indiLine.AddChild(birtLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid subtag INVALID_SUBTAG in BIRT event" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for invalid event subtag")
	}
}

func TestIndividualValidator_ValidateEventStructure_UserDefinedTag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with user-defined tag (should not error)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	userTag := gedcom.NewGedcomLine(2, "_CUSTOM", "value", "")
	birtLine.AddChild(userTag)
	indiLine.AddChild(nameLine)
	indiLine.AddChild(birtLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid subtag _CUSTOM in BIRT event" {
			t.Error("Should not error for user-defined tags")
		}
	}
}

func TestIndividualValidator_ValidateNameStructure_InvalidSubtag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid name subtag
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	invalidLine := gedcom.NewGedcomLine(2, "INVALID_SUBTAG", "value", "")
	nameLine.AddChild(invalidLine)
	indiLine.AddChild(nameLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid subtag INVALID_SUBTAG in NAME structure" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for invalid name subtag")
	}
}

func TestIndividualValidator_ValidateEvents_AllEventTypes(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with various event types
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	
	// Add multiple event types
	chrLine := gedcom.NewGedcomLine(1, "CHR", "", "")
	buriLine := gedcom.NewGedcomLine(1, "BURI", "", "")
	adopLine := gedcom.NewGedcomLine(1, "ADOP", "", "")
	
	indiLine.AddChild(nameLine)
	indiLine.AddChild(chrLine)
	indiLine.AddChild(buriLine)
	indiLine.AddChild(adopLine)
	
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	// Should validate without errors for valid event types
	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid tag CHR" {
			t.Error("CHR is a valid event tag")
		}
	}
}


