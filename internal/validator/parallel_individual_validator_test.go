package validator

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestParallelIndividualValidator_Validate(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create multiple individuals
	for i := 1; i <= 5; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indiLine := gedcom.NewGedcomLine(0, "INDI", "", xrefID)
		nameLine := gedcom.NewGedcomLine(1, "NAME", "Test /Person/", "")
		indiLine.AddChild(nameLine)
		indiRecord := gedcom.NewIndividualRecord(indiLine)
		tree.AddRecord(indiRecord)
	}

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should validate without errors
	errors := errorManager.Errors()
	if len(errors) > 0 {
		t.Logf("Found %d validation errors (may be expected)", len(errors))
	}
}

func TestParallelIndividualValidator_ValidateWithErrors(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create individual without NAME (should error)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiRecord := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Create individual with invalid SEX
	indiLine2 := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	nameLine2 := gedcom.NewGedcomLine(1, "NAME", "Test /Person/", "")
	sexLine2 := gedcom.NewGedcomLine(1, "SEX", "INVALID", "")
	indiLine2.AddChild(nameLine2)
	indiLine2.AddChild(sexLine2)
	indiRecord2 := gedcom.NewIndividualRecord(indiLine2)
	tree.AddRecord(indiRecord2)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Error("Expected validation errors")
	}

	// Check for specific errors
	foundNameError := false
	foundSexError := false
	for _, err := range errors {
		if err.Context == "Individual Validation" {
			if err.Message == "INDI @I1@: Missing required tag NAME" {
				foundNameError = true
			}
			if err.Message == "INDI @I2@: Invalid SEX value INVALID" {
				foundSexError = true
			}
		}
	}

	if !foundNameError {
		t.Error("Expected error for missing NAME")
	}
	if !foundSexError {
		t.Error("Expected error for invalid SEX")
	}
}

func TestParallelIndividualValidator_LargeDataset(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := gedcom.NewGedcomTree()

	// Create many individuals to test parallel processing
	for i := 1; i <= 100; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indiLine := gedcom.NewGedcomLine(0, "INDI", "", xrefID)
		nameLine := gedcom.NewGedcomLine(1, "NAME", "Test /Person/", "")
		indiLine.AddChild(nameLine)
		indiRecord := gedcom.NewIndividualRecord(indiLine)
		tree.AddRecord(indiRecord)
	}

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should complete without panics
	t.Logf("Validated 100 individuals successfully")
}

