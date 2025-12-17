package validator

import (
	"github.com/yourorg/gedcom/pkg/gedcom"
)

// Validator is the interface that all validators must implement.
type Validator interface {
	// Validate validates the entire GEDCOM tree.
	Validate(tree *gedcom.GedcomTree) error
}

// BaseValidator provides common functionality for all validators.
type BaseValidator struct {
	errorManager *gedcom.ErrorManager
}

// NewBaseValidator creates a new BaseValidator.
func NewBaseValidator(errorManager *gedcom.ErrorManager) *BaseValidator {
	return &BaseValidator{
		errorManager: errorManager,
	}
}

// AddError is a helper method to add errors.
func (bv *BaseValidator) AddError(severity gedcom.ErrorSeverity, message string, lineNumber int, context string) {
	bv.errorManager.AddError(severity, message, lineNumber, context)
}

// GetErrorManager returns the error manager.
func (bv *BaseValidator) GetErrorManager() *gedcom.ErrorManager {
	return bv.errorManager
}

// GedcomValidator orchestrates all validators and validates the entire GEDCOM tree.
type GedcomValidator struct {
	errorManager        *gedcom.ErrorManager
	individualValidator *IndividualValidator
	familyValidator     *FamilyValidator
	crossRefValidator   *CrossReferenceValidator
	headerValidator     *HeaderValidator
	advancedValidator   *AdvancedValidator // Optional advanced validation
}

// NewGedcomValidator creates a new GedcomValidator with all sub-validators.
func NewGedcomValidator(errorManager *gedcom.ErrorManager) *GedcomValidator {
	return &GedcomValidator{
		errorManager:        errorManager,
		individualValidator: NewIndividualValidator(errorManager),
		familyValidator:     NewFamilyValidator(errorManager),
		crossRefValidator:   NewCrossReferenceValidator(errorManager),
		headerValidator:     NewHeaderValidator(errorManager),
	}
}

// Validate runs all validators on the tree.
func (gv *GedcomValidator) Validate(tree *gedcom.GedcomTree) error {
	// Run all basic validators
	if err := gv.headerValidator.Validate(tree); err != nil {
		return err
	}

	if err := gv.individualValidator.Validate(tree); err != nil {
		return err
	}

	if err := gv.familyValidator.Validate(tree); err != nil {
		return err
	}

	if err := gv.crossRefValidator.Validate(tree); err != nil {
		return err
	}

	// Check for required SUBM record
	gv.validateSubmitter(tree)

	// Run advanced validators if configured
	if gv.advancedValidator != nil {
		if err := gv.advancedValidator.Validate(tree); err != nil {
			return err
		}
	}

	return nil
}

// validateSubmitter checks if at least one submitter record exists.
func (gv *GedcomValidator) validateSubmitter(tree *gedcom.GedcomTree) {
	// We'll need to add a method to tree to get all submitters
	// For now, we'll check if header has a SUBM reference
	header := tree.GetHeader()
	if header != nil {
		submXref := header.GetValue("SUBM")
		if submXref == "" {
			gv.errorManager.AddError(gedcom.SeveritySevere,
				"Missing SUBM record",
				0,
				"Submitter Validation")
		}
	} else {
		gv.errorManager.AddError(gedcom.SeveritySevere,
			"Missing SUBM record (no HEAD to check)",
			0,
			"Submitter Validation")
	}
}

// GetErrorManager returns the error manager.
func (gv *GedcomValidator) GetErrorManager() *gedcom.ErrorManager {
	return gv.errorManager
}

// EnableAdvancedValidation enables advanced validation with default rules.
// This adds date consistency, relationship logic, and other advanced checks.
func (gv *GedcomValidator) EnableAdvancedValidation() {
	if gv.advancedValidator == nil {
		gv.advancedValidator = NewAdvancedValidator(gv.errorManager)
		// Add default advanced rules
		gv.advancedValidator.AddRule(NewDateConsistencyValidator(gv.errorManager))
	}
}

// EnableAdvancedValidationWithConfig enables advanced validation with custom configuration.
func (gv *GedcomValidator) EnableAdvancedValidationWithConfig(config *ValidationConfig) {
	if gv.advancedValidator == nil {
		gv.advancedValidator = NewAdvancedValidatorWithConfig(gv.errorManager, config)
		// Add default advanced rules
		gv.advancedValidator.AddRule(NewDateConsistencyValidator(gv.errorManager))
	} else {
		gv.advancedValidator.SetConfig(config)
	}
}

// GetAdvancedValidator returns the advanced validator (if enabled).
func (gv *GedcomValidator) GetAdvancedValidator() *AdvancedValidator {
	return gv.advancedValidator
}
