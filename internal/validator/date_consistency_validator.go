package validator

import (
	"fmt"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// DateConsistencyValidator validates date consistency across records.
// This implements Phase 1 of advanced validation: Date Consistency.
type DateConsistencyValidator struct {
	*BaseValidator
}

// NewDateConsistencyValidator creates a new DateConsistencyValidator.
func NewDateConsistencyValidator(errorManager *gedcom.ErrorManager) *DateConsistencyValidator {
	return &DateConsistencyValidator{
		BaseValidator: NewBaseValidator(errorManager),
	}
}

// Name returns the name of this validation rule.
func (dcv *DateConsistencyValidator) Name() string {
	return "Date Consistency"
}

// Description returns a description of what this rule checks.
func (dcv *DateConsistencyValidator) Description() string {
	return "Validates that dates are logically consistent (birth before death, reasonable ages, etc.)"
}

// Validate validates date consistency across the GEDCOM tree.
func (dcv *DateConsistencyValidator) Validate(tree *gedcom.GedcomTree, config *ValidationConfig) []*gedcom.GedcomError {
	errors := make([]*gedcom.GedcomError, 0)

	// Validate individual date consistency
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}
		errors = append(errors, dcv.validateIndividualDates(indi, xrefID, tree, config)...)
	}

	// Validate family date consistency
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		fam, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}
		errors = append(errors, dcv.validateFamilyDates(fam, xrefID, tree, config)...)
	}

	// Validate cross-record date consistency
	errors = append(errors, dcv.validateCrossRecordDates(tree, config)...)

	return errors
}

// validateIndividualDates validates date consistency within an individual record.
func (dcv *DateConsistencyValidator) validateIndividualDates(indi *gedcom.IndividualRecord, xrefID string, tree *gedcom.GedcomTree, config *ValidationConfig) []*gedcom.GedcomError {
	errors := make([]*gedcom.GedcomError, 0)

	birthDateStr := indi.GetBirthDate()
	deathDateStr := indi.GetDeathDate()

	birthDate, _ := indi.GetBirthDateParsed()
	deathDate, _ := indi.GetDeathDateParsed()

	// Check for missing birth date (Info) - check string first
	if birthDateStr == "" {
		errors = append(errors, &gedcom.GedcomError{
			Severity:   gedcom.SeverityInfo,
			Message:    fmt.Sprintf("INDI %s: Missing birth date", xrefID),
			LineNumber: indi.FirstLine().LineNumber,
			Context:    "Date Consistency",
		})
		// Skip other date validations if no birth date
		return errors
	}

	// Check birth before death (Severe)
	if birthDate != nil && deathDate != nil && birthDate.IsValid() && deathDate.IsValid() {
		if deathDate.Earliest().Before(birthDate.Earliest()) {
			errors = append(errors, &gedcom.GedcomError{
				Severity:   gedcom.SeveritySevere,
				Message:    fmt.Sprintf("INDI %s: Death date (%s) is before birth date (%s)", xrefID, deathDate.String(), birthDate.String()),
				LineNumber: indi.FirstLine().LineNumber,
				Context:    "Date Consistency",
			})
		} else {
			// Check age at death (Warning for very old ages)
			age := dcv.calculateAge(birthDate, deathDate)
			if age > config.MaxDeathAge {
				errors = append(errors, &gedcom.GedcomError{
					Severity:   gedcom.SeverityWarning,
					Message:    fmt.Sprintf("INDI %s: Age at death (%d years) exceeds reasonable maximum (%d)", xrefID, age, config.MaxDeathAge),
					LineNumber: indi.FirstLine().LineNumber,
					Context:    "Date Consistency",
				})
			}
		}
	}

	// Check for missing death date if individual is very old (Info)
	if deathDateStr == "" {
		if birthDate != nil && birthDate.IsValid() {
			// Check if birth date is very old
			birthYear := birthDate.Year
			if birthYear > 0 {
				currentYear := time.Now().Year()
				age := currentYear - birthYear
				if age > 100 {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeverityInfo,
						Message:    fmt.Sprintf("INDI %s: Missing death date (individual would be %d years old)", xrefID, age),
						LineNumber: indi.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				}
			}
		}
	}

	// Check birth before marriage events
	marriageFamilies := indi.GetFamiliesAsSpouse()
	for _, famXref := range marriageFamilies {
		famRecord := tree.GetFamily(famXref)
		if famRecord == nil {
			continue
		}
		fam, ok := famRecord.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}
		marriageDate, err := fam.GetMarriageDateParsed()
		if err == nil && marriageDate != nil && marriageDate.IsValid() {
			if birthDate != nil && birthDate.IsValid() {
				// Check birth before marriage (Severe)
				if marriageDate.Earliest().Before(birthDate.Earliest()) {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeveritySevere,
						Message:    fmt.Sprintf("INDI %s: Marriage date (%s) is before birth date (%s)", xrefID, marriageDate.String(), birthDate.String()),
						LineNumber: indi.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				} else {
					// Check age at marriage (Warning for very young/old)
					age := dcv.calculateAge(birthDate, marriageDate)
					if age < config.MinMarriageAge {
						errors = append(errors, &gedcom.GedcomError{
							Severity:   gedcom.SeverityWarning,
							Message:    fmt.Sprintf("INDI %s: Age at marriage (%d years) is below minimum (%d)", xrefID, age, config.MinMarriageAge),
							LineNumber: indi.FirstLine().LineNumber,
							Context:    "Date Consistency",
						})
					} else if age > config.MaxMarriageAge {
						errors = append(errors, &gedcom.GedcomError{
							Severity:   gedcom.SeverityWarning,
							Message:    fmt.Sprintf("INDI %s: Age at marriage (%d years) exceeds maximum (%d)", xrefID, age, config.MaxMarriageAge),
							LineNumber: indi.FirstLine().LineNumber,
							Context:    "Date Consistency",
						})
					}
				}
			}
		}
	}

	return errors
}

// validateFamilyDates validates date consistency within a family record.
func (dcv *DateConsistencyValidator) validateFamilyDates(fam *gedcom.FamilyRecord, xrefID string, tree *gedcom.GedcomTree, config *ValidationConfig) []*gedcom.GedcomError {
	errors := make([]*gedcom.GedcomError, 0)

	marriageDate, _ := fam.GetMarriageDateParsed()
	divorceDate, _ := fam.GetDivorceDateParsed()

	// Check marriage before divorce (Severe)
	if marriageDate != nil && divorceDate != nil && marriageDate.IsValid() && divorceDate.IsValid() {
		if divorceDate.Earliest().Before(marriageDate.Earliest()) {
			errors = append(errors, &gedcom.GedcomError{
				Severity:   gedcom.SeveritySevere,
				Message:    fmt.Sprintf("FAM %s: Divorce date (%s) is before marriage date (%s)", xrefID, divorceDate.String(), marriageDate.String()),
				LineNumber: fam.FirstLine().LineNumber,
				Context:    "Date Consistency",
			})
		} else {
			// Check marriage duration (Warning for very long marriages)
			duration := dcv.calculateAge(marriageDate, divorceDate)
			if duration > 80 {
				errors = append(errors, &gedcom.GedcomError{
					Severity:   gedcom.SeverityWarning,
					Message:    fmt.Sprintf("FAM %s: Marriage duration (%d years) is unusually long", xrefID, duration),
					LineNumber: fam.FirstLine().LineNumber,
					Context:    "Date Consistency",
				})
			}
		}
	}

	// Check marriage before children's births
	if marriageDate != nil && marriageDate.IsValid() {
		children := fam.GetChildren()
		for _, childXref := range children {
			childRecord := tree.GetIndividual(childXref)
			if childRecord == nil {
				continue
			}
			child, ok := childRecord.(*gedcom.IndividualRecord)
			if !ok {
				continue
			}
			birthDate, err := child.GetBirthDateParsed()
			if err == nil && birthDate != nil && birthDate.IsValid() {
				// Allow some flexibility for pre-marital births (9 months before)
				marriageTime := marriageDate.Earliest()
				birthTime := birthDate.Earliest()
				monthsBefore := int(marriageTime.Sub(birthTime).Hours() / 24 / 30)
				if monthsBefore > 9 {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeverityInfo,
						Message:    fmt.Sprintf("FAM %s: Child %s born %d months before marriage", xrefID, childXref, monthsBefore),
						LineNumber: fam.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				}
			}
		}
	}

	return errors
}

// validateCrossRecordDates validates date consistency across records (parent-child, siblings).
func (dcv *DateConsistencyValidator) validateCrossRecordDates(tree *gedcom.GedcomTree, config *ValidationConfig) []*gedcom.GedcomError {
	errors := make([]*gedcom.GedcomError, 0)

	families := tree.GetAllFamilies()
	for famXref, famRecord := range families {
		fam, ok := famRecord.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		husbandXref := fam.GetHusband()
		wifeXref := fam.GetWife()
		children := fam.GetChildren()

		// Get parent birth dates
		var husbandBirth, wifeBirth *gedcom.GedcomDate
		if husbandXref != "" {
			husbandRecord := tree.GetIndividual(husbandXref)
			if husbandRecord != nil {
				husband, ok := husbandRecord.(*gedcom.IndividualRecord)
				if ok {
					husbandBirth, _ = husband.GetBirthDateParsed()
				}
			}
		}
		if wifeXref != "" {
			wifeRecord := tree.GetIndividual(wifeXref)
			if wifeRecord != nil {
				wife, ok := wifeRecord.(*gedcom.IndividualRecord)
				if ok {
					wifeBirth, _ = wife.GetBirthDateParsed()
				}
			}
		}

		// Check parent-child age gaps
		for _, childXref := range children {
			childRecord := tree.GetIndividual(childXref)
			if childRecord == nil {
				continue
			}
			child, ok := childRecord.(*gedcom.IndividualRecord)
			if !ok {
				continue
			}
			childBirth, err := child.GetBirthDateParsed()
			if err != nil || childBirth == nil || !childBirth.IsValid() {
				continue
			}

			// Check father-child age gap
			if husbandBirth != nil && husbandBirth.IsValid() {
				age := dcv.calculateAge(husbandBirth, childBirth)
				if age < config.MinParentAge {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeverityWarning,
						Message:    fmt.Sprintf("FAM %s: Father age at child %s birth (%d years) is below minimum (%d)", famXref, childXref, age, config.MinParentAge),
						LineNumber: fam.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				} else if age > config.MaxParentAge {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeverityWarning,
						Message:    fmt.Sprintf("FAM %s: Father age at child %s birth (%d years) exceeds maximum (%d)", famXref, childXref, age, config.MaxParentAge),
						LineNumber: fam.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				}
			}

			// Check mother-child age gap
			if wifeBirth != nil && wifeBirth.IsValid() {
				age := dcv.calculateAge(wifeBirth, childBirth)
				if age < config.MinParentAge {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeverityWarning,
						Message:    fmt.Sprintf("FAM %s: Mother age at child %s birth (%d years) is below minimum (%d)", famXref, childXref, age, config.MinParentAge),
						LineNumber: fam.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				} else if age > config.MaxParentAge {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeverityWarning,
						Message:    fmt.Sprintf("FAM %s: Mother age at child %s birth (%d years) exceeds maximum (%d)", famXref, childXref, age, config.MaxParentAge),
						LineNumber: fam.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				}
			}

			// Check child born before parent birth (Severe)
			if husbandBirth != nil && husbandBirth.IsValid() {
				if childBirth.Earliest().Before(husbandBirth.Earliest()) {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeveritySevere,
						Message:    fmt.Sprintf("FAM %s: Child %s birth date (%s) is before father birth date (%s)", famXref, childXref, childBirth.String(), husbandBirth.String()),
						LineNumber: fam.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				}
			}
			if wifeBirth != nil && wifeBirth.IsValid() {
				if childBirth.Earliest().Before(wifeBirth.Earliest()) {
					errors = append(errors, &gedcom.GedcomError{
						Severity:   gedcom.SeveritySevere,
						Message:    fmt.Sprintf("FAM %s: Child %s birth date (%s) is before mother birth date (%s)", famXref, childXref, childBirth.String(), wifeBirth.String()),
						LineNumber: fam.FirstLine().LineNumber,
						Context:    "Date Consistency",
					})
				}
			}
		}

		// Check sibling age differences (Warning for very large gaps)
		if len(children) > 1 {
			for i := 0; i < len(children)-1; i++ {
				child1Record := tree.GetIndividual(children[i])
				child2Record := tree.GetIndividual(children[i+1])
				if child1Record == nil || child2Record == nil {
					continue
				}
				child1, ok1 := child1Record.(*gedcom.IndividualRecord)
				child2, ok2 := child2Record.(*gedcom.IndividualRecord)
				if !ok1 || !ok2 {
					continue
				}

				birth1, err1 := child1.GetBirthDateParsed()
				birth2, err2 := child2.GetBirthDateParsed()
				if err1 == nil && err2 == nil && birth1 != nil && birth2 != nil && birth1.IsValid() && birth2.IsValid() {
					ageGap := dcv.calculateAge(birth1, birth2)
					if ageGap > 50 {
						errors = append(errors, &gedcom.GedcomError{
							Severity:   gedcom.SeverityWarning,
							Message:    fmt.Sprintf("FAM %s: Siblings %s and %s have large age gap (%d years)", famXref, children[i], children[i+1], ageGap),
							LineNumber: fam.FirstLine().LineNumber,
							Context:    "Date Consistency",
						})
					}
				}
			}
		}
	}

	return errors
}

// calculateAge calculates the age in years between two dates.
// Returns 0 if dates are invalid or cannot be calculated.
func (dcv *DateConsistencyValidator) calculateAge(startDate, endDate *gedcom.GedcomDate) int {
	if startDate == nil || endDate == nil || !startDate.IsValid() || !endDate.IsValid() {
		return 0
	}

	startTime := startDate.Earliest()
	endTime := endDate.Earliest()

	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}

	// Calculate years difference
	years := endTime.Year() - startTime.Year()

	// Adjust if end date hasn't reached start date's month/day yet
	if endTime.Month() < startTime.Month() || (endTime.Month() == startTime.Month() && endTime.Day() < startTime.Day()) {
		years--
	}

	return years
}
