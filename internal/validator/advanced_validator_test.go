package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestAdvancedValidator_AddRule(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewAdvancedValidator(errorManager)

	rule := NewDateConsistencyValidator(errorManager)
	validator.AddRule(rule)

	rules := validator.GetRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	if rules[0].Name() != "Date Consistency" {
		t.Errorf("Expected rule name 'Date Consistency', got %q", rules[0].Name())
	}
}

func TestAdvancedValidator_SeverityFiltering(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	config := NewValidationConfig()
	config.MinSeverity = gedcom.SeverityWarning // Only warnings and above

	validator := NewAdvancedValidatorWithConfig(errorManager, config)
	rule := NewDateConsistencyValidator(errorManager)
	validator.AddRule(rule)

	// Create tree with info-level error (missing birth date)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))

	indi := gedcom.NewIndividualRecord(indiLine)
	tree := gedcom.NewGedcomTree()
	tree.AddRecord(indi)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Info errors should be filtered out
	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Severity == gedcom.SeverityInfo {
			t.Error("Info errors should be filtered out when MinSeverity is Warning")
		}
	}
}

func TestAdvancedValidator_ShowAllSeverities(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	config := NewValidationConfig()
	config.MinSeverity = gedcom.SeverityHint // Show all

	validator := NewAdvancedValidatorWithConfig(errorManager, config)
	rule := NewDateConsistencyValidator(errorManager)
	validator.AddRule(rule)

	// Create tree with info-level error (missing birth date)
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))

	indi := gedcom.NewIndividualRecord(indiLine)
	tree := gedcom.NewGedcomTree()
	tree.AddRecord(indi)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Info errors should be included
	errors := errorManager.Errors()
	infoFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeverityInfo {
			infoFound = true
			break
		}
	}

	if !infoFound {
		t.Error("Info errors should be included when MinSeverity is Hint")
	}
}

func TestGedcomValidator_EnableAdvancedValidation(t *testing.T) {
	errorManager := gedcom.NewErrorManager()
	validator := NewGedcomValidator(errorManager)

	// Enable advanced validation
	validator.EnableAdvancedValidation()

	// Create tree with date consistency issue
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
	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should have advanced validation errors
	errors := errorManager.Errors()
	severeFound := false
	for _, err := range errors {
		if err.Severity == gedcom.SeveritySevere && err.Context == "Date Consistency" {
			severeFound = true
			break
		}
	}

	if !severeFound {
		t.Error("Expected severe error from advanced validation")
	}
}
