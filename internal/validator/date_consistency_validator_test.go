package validator

import (
	"testing"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

func TestDateConsistencyValidator_DeathBeforeBirth(t *testing.T) {
	// Create individual with death before birth (Severe error)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	birtDateLine := gedcom.NewGedcomLine(2, "DATE", "15 JAN 1900", "")
	deatLine := gedcom.NewGedcomLine(1, "DEAT", "", "")
	deatDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before birth!

	birtLine.AddChild(birtDateLine)
	deatLine.AddChild(deatDateLine)
	indiLine.AddChild(birtLine)
	indiLine.AddChild(deatLine)

	indi := gedcom.NewIndividualRecord(indiLine)
	tree := gedcom.NewGedcomTree()
	tree.AddRecord(indi)

	errorManager := gedcom.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	// Should have at least one severe error
	severeFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeveritySevere {
			severeFound = true
			if err.Message == "" {
				t.Error("Error message should not be empty")
			}
		}
	}

	if !severeFound {
		t.Error("Expected severe error for death before birth")
	}
}

func TestDateConsistencyValidator_MarriageBeforeBirth(t *testing.T) {
	// Create individual and family with marriage before birth
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	birtDateLine := gedcom.NewGedcomLine(2, "DATE", "15 JAN 1900", "")
	birtLine.AddChild(birtDateLine)
	indiLine.AddChild(birtLine)
	indiLine.AddChild(gedcom.NewGedcomLine(1, "FAMS", "@F1@", ""))

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := gedcom.NewGedcomLine(1, "MARR", "", "")
	marrDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before birth!
	marrLine.AddChild(marrDateLine)
	famLine.AddChild(marrLine)
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))

	indi := gedcom.NewIndividualRecord(indiLine)
	fam := gedcom.NewFamilyRecord(famLine)

	tree := gedcom.NewGedcomTree()
	tree.AddRecord(indi)
	tree.AddRecord(fam)

	errorManager := gedcom.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	severeFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeveritySevere {
			severeFound = true
		}
	}

	if !severeFound {
		t.Error("Expected severe error for marriage before birth")
	}
}

func TestDateConsistencyValidator_YoungMarriageAge(t *testing.T) {
	// Create individual with very young marriage age (Warning)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	birtDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1900", "")
	birtLine.AddChild(birtDateLine)
	indiLine.AddChild(birtLine)
	indiLine.AddChild(gedcom.NewGedcomLine(1, "FAMS", "@F1@", ""))

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := gedcom.NewGedcomLine(1, "MARR", "", "")
	marrDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1905", "") // Age 5!
	marrLine.AddChild(marrDateLine)
	famLine.AddChild(marrLine)
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))

	indi := gedcom.NewIndividualRecord(indiLine)
	fam := gedcom.NewFamilyRecord(famLine)

	tree := gedcom.NewGedcomTree()
	tree.AddRecord(indi)
	tree.AddRecord(fam)

	errorManager := gedcom.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()
	config.MinMarriageAge = 12

	errors := validator.Validate(tree, config)

	warningFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeverityWarning {
			warningFound = true
		}
	}

	if !warningFound {
		t.Error("Expected warning for very young marriage age")
	}
}

func TestDateConsistencyValidator_OldParentAge(t *testing.T) {
	// Create family with very old parent (Warning)
	fatherLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	fatherBirtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	fatherBirtDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1800", "")
	fatherBirtLine.AddChild(fatherBirtDateLine)
	fatherLine.AddChild(fatherBirtLine)

	childLine := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	childBirtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	childBirtDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1900", "") // Father would be 100
	childBirtLine.AddChild(childBirtDateLine)
	childLine.AddChild(childBirtLine)
	childLine.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))

	father := gedcom.NewIndividualRecord(fatherLine)
	child := gedcom.NewIndividualRecord(childLine)
	fam := gedcom.NewFamilyRecord(famLine)

	tree := gedcom.NewGedcomTree()
	tree.AddRecord(father)
	tree.AddRecord(child)
	tree.AddRecord(fam)

	errorManager := gedcom.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()
	config.MaxParentAge = 80

	errors := validator.Validate(tree, config)

	warningFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeverityWarning {
			warningFound = true
		}
	}

	if !warningFound {
		t.Error("Expected warning for very old parent age")
	}
}

func TestDateConsistencyValidator_MissingBirthDate(t *testing.T) {
	// Create individual without birth date (Info)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))

	indi := gedcom.NewIndividualRecord(indiLine)
	tree := gedcom.NewGedcomTree()
	tree.AddRecord(indi)

	errorManager := gedcom.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	infoFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeverityInfo {
			infoFound = true
		}
	}

	if !infoFound {
		t.Error("Expected info-level error for missing birth date")
	}
}

func TestDateConsistencyValidator_DivorceBeforeMarriage(t *testing.T) {
	// Create family with divorce before marriage (Severe)
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := gedcom.NewGedcomLine(1, "MARR", "", "")
	marrDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1900", "")
	divLine := gedcom.NewGedcomLine(1, "DIV", "", "")
	divDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before marriage!

	marrLine.AddChild(marrDateLine)
	divLine.AddChild(divDateLine)
	famLine.AddChild(marrLine)
	famLine.AddChild(divLine)

	fam := gedcom.NewFamilyRecord(famLine)
	tree := gedcom.NewGedcomTree()
	tree.AddRecord(fam)

	errorManager := gedcom.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	severeFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeveritySevere {
			severeFound = true
		}
	}

	if !severeFound {
		t.Error("Expected severe error for divorce before marriage")
	}
}

func TestDateConsistencyValidator_ChildBeforeParentBirth(t *testing.T) {
	// Create family where child is born before parent (Severe)
	fatherLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	fatherBirtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	fatherBirtDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1900", "")
	fatherBirtLine.AddChild(fatherBirtDateLine)
	fatherLine.AddChild(fatherBirtLine)

	childLine := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	childBirtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	childBirtDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before father!
	childBirtLine.AddChild(childBirtDateLine)
	childLine.AddChild(childBirtLine)
	childLine.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))

	father := gedcom.NewIndividualRecord(fatherLine)
	child := gedcom.NewIndividualRecord(childLine)
	fam := gedcom.NewFamilyRecord(famLine)

	tree := gedcom.NewGedcomTree()
	tree.AddRecord(father)
	tree.AddRecord(child)
	tree.AddRecord(fam)

	errorManager := gedcom.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	severeFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeveritySevere {
			severeFound = true
		}
	}

	if !severeFound {
		t.Error("Expected severe error for child born before parent")
	}
}

func TestAdvancedValidator_Integration(t *testing.T) {
	// Test the AdvancedValidator with DateConsistencyValidator
	errorManager := gedcom.NewErrorManager()
	advancedValidator := NewAdvancedValidator(errorManager)

	// Add date consistency rule
	dateRule := NewDateConsistencyValidator(errorManager)
	advancedValidator.AddRule(dateRule)

	// Create a tree with date issues
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	birtDateLine := gedcom.NewGedcomLine(2, "DATE", "15 JAN 1900", "")
	deatLine := gedcom.NewGedcomLine(1, "DEAT", "", "")
	deatDateLine := gedcom.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before birth!

	birtLine.AddChild(birtDateLine)
	deatLine.AddChild(deatDateLine)
	indiLine.AddChild(birtLine)
	indiLine.AddChild(deatLine)

	indi := gedcom.NewIndividualRecord(indiLine)
	tree := gedcom.NewGedcomTree()
	tree.AddRecord(indi)

	// Validate
	err := advancedValidator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Check that errors were collected
	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Error("Expected errors to be collected")
	}
}
