package validator

import (
	"sync"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// ParallelGedcomValidator runs validators in parallel for better performance.
type ParallelGedcomValidator struct {
	errorManager        *gedcom.ErrorManager
	individualValidator *IndividualValidator
	familyValidator     *FamilyValidator
	crossRefValidator   *CrossReferenceValidator
	headerValidator     *HeaderValidator
}

// NewParallelGedcomValidator creates a new ParallelGedcomValidator with all sub-validators.
func NewParallelGedcomValidator(errorManager *gedcom.ErrorManager) *ParallelGedcomValidator {
	return &ParallelGedcomValidator{
		errorManager:        errorManager,
		individualValidator: NewIndividualValidator(errorManager),
		familyValidator:     NewFamilyValidator(errorManager),
		crossRefValidator:   NewCrossReferenceValidator(errorManager),
		headerValidator:     NewHeaderValidator(errorManager),
	}
}

// Validate runs all validators in parallel using goroutines.
func (pgv *ParallelGedcomValidator) Validate(tree *gedcom.GedcomTree) error {
	var wg sync.WaitGroup
	var validationErrors []error
	var mu sync.Mutex

	// Run validators in parallel
	wg.Add(4)

	// Header validator (runs first, but in parallel with others)
	go func() {
		defer wg.Done()
		if err := pgv.headerValidator.Validate(tree); err != nil {
			mu.Lock()
			validationErrors = append(validationErrors, err)
			mu.Unlock()
		}
	}()

	// Individual validator
	go func() {
		defer wg.Done()
		if err := pgv.individualValidator.Validate(tree); err != nil {
			mu.Lock()
			validationErrors = append(validationErrors, err)
			mu.Unlock()
		}
	}()

	// Family validator
	go func() {
		defer wg.Done()
		if err := pgv.familyValidator.Validate(tree); err != nil {
			mu.Lock()
			validationErrors = append(validationErrors, err)
			mu.Unlock()
		}
	}()

	// Cross-reference validator (depends on individuals and families being parsed)
	go func() {
		defer wg.Done()
		if err := pgv.crossRefValidator.Validate(tree); err != nil {
			mu.Lock()
			validationErrors = append(validationErrors, err)
			mu.Unlock()
		}
	}()

	// Wait for all validators to complete
	wg.Wait()

	// Check for required SUBM record (sequential, quick check)
	pgv.validateSubmitter(tree)

	// Return first error if any
	if len(validationErrors) > 0 {
		return validationErrors[0]
	}

	return nil
}

// validateSubmitter checks if at least one submitter record exists.
func (pgv *ParallelGedcomValidator) validateSubmitter(tree *gedcom.GedcomTree) {
	header := tree.GetHeader()
	if header != nil {
		submXref := header.GetValue("SUBM")
		if submXref == "" {
			pgv.errorManager.AddError(gedcom.SeveritySevere,
				"Missing SUBM record",
				0,
				"Submitter Validation")
		}
	} else {
		pgv.errorManager.AddError(gedcom.SeveritySevere,
			"Missing SUBM record (no HEAD to check)",
			0,
			"Submitter Validation")
	}
}

// GetErrorManager returns the error manager.
func (pgv *ParallelGedcomValidator) GetErrorManager() *gedcom.ErrorManager {
	return pgv.errorManager
}
