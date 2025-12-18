package validator

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestParallelIndividualValidator_ValidateReferences(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid FAMS reference format
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	famsLine := gedcom.NewGedcomLine(1, "FAMS", "INVALID_FORMAT", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(famsLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid FAMS reference format INVALID_FORMAT" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid FAMS reference format")
	}
}

func TestParallelIndividualValidator_ValidateReferences_FAMC(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid FAMC reference format
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	famcLine := gedcom.NewGedcomLine(1, "FAMC", "INVALID_FORMAT", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(famcLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

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

func TestParallelIndividualValidator_ValidateReferences_MultipleFAMS(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

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

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should validate format (but not existence in parallel validator)
	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid FAMS reference format @F1@" {
			t.Error("Should not error for valid FAMS format")
		}
	}
}

func TestParallelIndividualValidator_ValidateStructure_InvalidTag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with invalid tag
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	invalidLine := gedcom.NewGedcomLine(1, "INVALID_TAG", "value", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(invalidLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid tag INVALID_TAG" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid tag")
	}
}

func TestParallelIndividualValidator_ValidateStructure_UserDefinedTag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual with user-defined tag (should not error)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	userTag := gedcom.NewGedcomLine(1, "_CUSTOM", "value", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(userTag)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid tag _CUSTOM" {
			t.Error("Should not error for user-defined tags")
		}
	}
}

func TestParallelIndividualValidator_ValidateStructure_MissingRequiredTag(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual without NAME (required)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	sexLine := gedcom.NewGedcomLine(1, "SEX", "M", "")
	indiLine.AddChild(sexLine)
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

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

func TestParallelIndividualValidator_ConcurrentValidation(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create many individuals to test concurrent processing
	for i := 1; i <= 50; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indiLine := gedcom.NewGedcomLine(0, "INDI", "", xrefID)
		nameLine := gedcom.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", i), "")
		indiLine.AddChild(nameLine)
		indiRecord := gedcom.NewIndividualRecord(indiLine)
		tree.AddRecord(indiRecord)
	}

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should complete without panics
	individuals := tree.GetAllIndividuals()
	if len(individuals) != 50 {
		t.Errorf("Expected 50 individuals, got %d", len(individuals))
	}

	t.Logf("Successfully validated %d individuals concurrently", len(individuals))
}

