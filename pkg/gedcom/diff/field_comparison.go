package diff

import (
	"strings"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// compareIndividual compares two individual records field by field.
func (gd *GedcomDiffer) compareIndividual(indi1, indi2 *gedcom.IndividualRecord) []FieldChange {
	changes := make([]FieldChange, 0)

	// Compare name
	name1 := indi1.GetName()
	name2 := indi2.GetName()
	if name1 != name2 {
		history := []ChangeHistory{}
		if gd.config.TrackHistory {
			history = append(history, gd.createChangeHistory(
				ChangeTypeModified,
				"NAME",
				name1,
				name2,
			))
		}
		changes = append(changes, FieldChange{
			Field:    "NAME",
			Path:     "NAME",
			OldValue: name1,
			NewValue: name2,
			Type:     ChangeTypeModified,
			History:  history,
		})
	}

	// Compare given name
	given1 := indi1.GetGivenName()
	given2 := indi2.GetGivenName()
	if given1 != given2 {
		history := []ChangeHistory{}
		if gd.config.TrackHistory {
			history = append(history, gd.createChangeHistory(
				ChangeTypeModified,
				"NAME.GIVN",
				given1,
				given2,
			))
		}
		changes = append(changes, FieldChange{
			Field:    "GIVN",
			Path:     "NAME.GIVN",
			OldValue: given1,
			NewValue: given2,
			Type:     ChangeTypeModified,
			History:  history,
		})
	}

	// Compare surname
	surname1 := indi1.GetSurname()
	surname2 := indi2.GetSurname()
	if surname1 != surname2 {
		history := []ChangeHistory{}
		if gd.config.TrackHistory {
			history = append(history, gd.createChangeHistory(
				ChangeTypeModified,
				"NAME.SURN",
				surname1,
				surname2,
			))
		}
		changes = append(changes, FieldChange{
			Field:    "SURN",
			Path:     "NAME.SURN",
			OldValue: surname1,
			NewValue: surname2,
			Type:     ChangeTypeModified,
			History:  history,
		})
	}

	// Compare sex
	sex1 := indi1.GetSex()
	sex2 := indi2.GetSex()
	if sex1 != sex2 {
		history := []ChangeHistory{}
		if gd.config.TrackHistory {
			history = append(history, gd.createChangeHistory(
				ChangeTypeModified,
				"SEX",
				sex1,
				sex2,
			))
		}
		changes = append(changes, FieldChange{
			Field:    "SEX",
			Path:     "SEX",
			OldValue: sex1,
			NewValue: sex2,
			Type:     ChangeTypeModified,
			History:  history,
		})
	}

	// Compare birth date
	birthDate1 := indi1.GetBirthDate()
	birthDate2 := indi2.GetBirthDate()
	dateChange := gd.compareDate(birthDate1, birthDate2, "BIRT.DATE")
	if dateChange != nil {
		changes = append(changes, *dateChange)
	}

	// Compare birth place
	birthPlace1 := indi1.GetBirthPlace()
	birthPlace2 := indi2.GetBirthPlace()
	placeChange := gd.comparePlace(birthPlace1, birthPlace2, "BIRT.PLAC")
	if placeChange != nil {
		changes = append(changes, *placeChange)
	}

	// Compare death date
	deathDate1 := indi1.GetDeathDate()
	deathDate2 := indi2.GetDeathDate()
	if deathDate1 != deathDate2 {
		// Check if one is empty (added/removed)
		if deathDate1 == "" && deathDate2 != "" {
			history := []ChangeHistory{}
			if gd.config.TrackHistory {
				history = append(history, gd.createChangeHistory(
					ChangeTypeAdded,
					"DEAT.DATE",
					"",
					deathDate2,
				))
			}
			changes = append(changes, FieldChange{
				Field:    "DATE",
				Path:     "DEAT.DATE",
				OldValue: nil,
				NewValue: deathDate2,
				Type:     ChangeTypeAdded,
				History:  history,
			})
		} else if deathDate1 != "" && deathDate2 == "" {
			history := []ChangeHistory{}
			if gd.config.TrackHistory {
				history = append(history, gd.createChangeHistory(
					ChangeTypeRemoved,
					"DEAT.DATE",
					deathDate1,
					"",
				))
			}
			changes = append(changes, FieldChange{
				Field:    "DATE",
				Path:     "DEAT.DATE",
				OldValue: deathDate1,
				NewValue: nil,
				Type:     ChangeTypeRemoved,
				History:  history,
			})
		} else {
			dateChange := gd.compareDate(deathDate1, deathDate2, "DEAT.DATE")
			if dateChange != nil {
				changes = append(changes, *dateChange)
			}
		}
	}

	// Compare death place
	deathPlace1 := indi1.GetDeathPlace()
	deathPlace2 := indi2.GetDeathPlace()
	if deathPlace1 != deathPlace2 {
		if deathPlace1 == "" && deathPlace2 != "" {
			history := []ChangeHistory{}
			if gd.config.TrackHistory {
				history = append(history, gd.createChangeHistory(
					ChangeTypeAdded,
					"DEAT.PLAC",
					"",
					deathPlace2,
				))
			}
			changes = append(changes, FieldChange{
				Field:    "PLAC",
				Path:     "DEAT.PLAC",
				OldValue: nil,
				NewValue: deathPlace2,
				Type:     ChangeTypeAdded,
				History:  history,
			})
		} else if deathPlace1 != "" && deathPlace2 == "" {
			history := []ChangeHistory{}
			if gd.config.TrackHistory {
				history = append(history, gd.createChangeHistory(
					ChangeTypeRemoved,
					"DEAT.PLAC",
					deathPlace1,
					"",
				))
			}
			changes = append(changes, FieldChange{
				Field:    "PLAC",
				Path:     "DEAT.PLAC",
				OldValue: deathPlace1,
				NewValue: nil,
				Type:     ChangeTypeRemoved,
				History:  history,
			})
		} else {
			placeChange := gd.comparePlace(deathPlace1, deathPlace2, "DEAT.PLAC")
			if placeChange != nil {
				changes = append(changes, *placeChange)
			}
		}
	}

	return changes
}

// compareFamily compares two family records field by field.
func (gd *GedcomDiffer) compareFamily(fam1, fam2 *gedcom.FamilyRecord) []FieldChange {
	changes := make([]FieldChange, 0)

	// Compare husband
	husb1 := fam1.GetHusband()
	husb2 := fam2.GetHusband()
	if husb1 != husb2 {
		history := []ChangeHistory{}
		if gd.config.TrackHistory {
			history = append(history, gd.createChangeHistory(
				ChangeTypeModified,
				"HUSB",
				husb1,
				husb2,
			))
		}
		changes = append(changes, FieldChange{
			Field:    "HUSB",
			Path:     "HUSB",
			OldValue: husb1,
			NewValue: husb2,
			Type:     ChangeTypeModified,
			History:  history,
		})
	}

	// Compare wife
	wife1 := fam1.GetWife()
	wife2 := fam2.GetWife()
	if wife1 != wife2 {
		history := []ChangeHistory{}
		if gd.config.TrackHistory {
			history = append(history, gd.createChangeHistory(
				ChangeTypeModified,
				"WIFE",
				wife1,
				wife2,
			))
		}
		changes = append(changes, FieldChange{
			Field:    "WIFE",
			Path:     "WIFE",
			OldValue: wife1,
			NewValue: wife2,
			Type:     ChangeTypeModified,
			History:  history,
		})
	}

	// Compare children
	children1 := fam1.GetChildren()
	children2 := fam2.GetChildren()
	childChanges := gd.compareChildren(children1, children2)
	changes = append(changes, childChanges...)

	// Compare marriage date
	marrDate1 := fam1.GetMarriageDate()
	marrDate2 := fam2.GetMarriageDate()
	dateChange := gd.compareDate(marrDate1, marrDate2, "MARR.DATE")
	if dateChange != nil {
		changes = append(changes, *dateChange)
	}

	// Compare marriage place
	marrPlace1 := fam1.GetMarriagePlace()
	marrPlace2 := fam2.GetMarriagePlace()
	placeChange := gd.comparePlace(marrPlace1, marrPlace2, "MARR.PLAC")
	if placeChange != nil {
		changes = append(changes, *placeChange)
	}

	return changes
}

// compareChildren compares children lists.
func (gd *GedcomDiffer) compareChildren(children1, children2 []string) []FieldChange {
	changes := make([]FieldChange, 0)

	// Create maps for quick lookup
	map1 := make(map[string]bool)
	for _, child := range children1 {
		map1[child] = true
	}

	map2 := make(map[string]bool)
	for _, child := range children2 {
		map2[child] = true
	}

	// Find added children
	for _, child := range children2 {
		if !map1[child] {
			history := []ChangeHistory{}
			if gd.config.TrackHistory {
				history = append(history, gd.createChangeHistory(
					ChangeTypeAdded,
					"CHIL",
					"",
					child,
				))
			}
			changes = append(changes, FieldChange{
				Field:    "CHIL",
				Path:     "CHIL",
				OldValue: nil,
				NewValue: child,
				Type:     ChangeTypeAdded,
				History:  history,
			})
		}
	}

	// Find removed children
	for _, child := range children1 {
		if !map2[child] {
			history := []ChangeHistory{}
			if gd.config.TrackHistory {
				history = append(history, gd.createChangeHistory(
					ChangeTypeRemoved,
					"CHIL",
					child,
					"",
				))
			}
			changes = append(changes, FieldChange{
				Field:    "CHIL",
				Path:     "CHIL",
				OldValue: child,
				NewValue: nil,
				Type:     ChangeTypeRemoved,
				History:  history,
			})
		}
	}

	return changes
}

// compareBasicRecord compares basic record fields.
func (gd *GedcomDiffer) compareBasicRecord(record1, record2 gedcom.Record) []FieldChange {
	changes := make([]FieldChange, 0)

	// Compare basic value fields
	// This is a simplified comparison - can be enhanced based on record type
	return changes
}

// compareDate compares two date strings with semantic equivalence.
func (gd *GedcomDiffer) compareDate(date1, date2, path string) *FieldChange {
	if date1 == date2 {
		return nil
	}

	// Check for semantic equivalence
	if gd.areDatesSemanticallyEquivalent(date1, date2) {
		history := []ChangeHistory{}
		if gd.config.TrackHistory {
			history = append(history, gd.createChangeHistory(
				ChangeTypeSemanticallyEquivalent,
				path,
				date1,
				date2,
			))
		}
		return &FieldChange{
			Field:    "DATE",
			Path:     path,
			OldValue: date1,
			NewValue: date2,
			Type:     ChangeTypeSemanticallyEquivalent,
			History:  history,
		}
	}

	// Modified date
	history := []ChangeHistory{}
	if gd.config.TrackHistory {
		history = append(history, gd.createChangeHistory(
			ChangeTypeModified,
			path,
			date1,
			date2,
		))
	}
	return &FieldChange{
		Field:    "DATE",
		Path:     path,
		OldValue: date1,
		NewValue: date2,
		Type:     ChangeTypeModified,
		History:  history,
	}
}

// comparePlace compares two place strings with semantic equivalence.
func (gd *GedcomDiffer) comparePlace(place1, place2, path string) *FieldChange {
	if place1 == place2 {
		return nil
	}

	// Normalize and compare
	norm1 := normalizePlace(place1)
	norm2 := normalizePlace(place2)

	if norm1 == norm2 {
		// Semantically equivalent (different format, same place)
		history := []ChangeHistory{}
		if gd.config.TrackHistory {
			history = append(history, gd.createChangeHistory(
				ChangeTypeSemanticallyEquivalent,
				path,
				place1,
				place2,
			))
		}
		return &FieldChange{
			Field:    "PLAC",
			Path:     path,
			OldValue: place1,
			NewValue: place2,
			Type:     ChangeTypeSemanticallyEquivalent,
			History:  history,
		}
	}

	// Modified place
	history := []ChangeHistory{}
	if gd.config.TrackHistory {
		history = append(history, gd.createChangeHistory(
			ChangeTypeModified,
			path,
			place1,
			place2,
		))
	}
	return &FieldChange{
		Field:    "PLAC",
		Path:     path,
		OldValue: place1,
		NewValue: place2,
		Type:     ChangeTypeModified,
		History:  history,
	}
}

// areDatesSemanticallyEquivalent checks if two dates are semantically equivalent.
func (gd *GedcomDiffer) areDatesSemanticallyEquivalent(date1, date2 string) bool {
	if date1 == "" || date2 == "" {
		return false
	}

	// Try to parse dates
	parsed1, err1 := gedcom.ParseDate(date1)
	parsed2, err2 := gedcom.ParseDate(date2)

	if err1 != nil || err2 != nil {
		// Fallback to string comparison
		return false
	}

	// Extract years
	year1 := parsed1.Year
	if year1 == 0 {
		year1 = parsed1.StartYear
	}
	year2 := parsed2.Year
	if year2 == 0 {
		year2 = parsed2.StartYear
	}

	if year1 == 0 || year2 == 0 {
		return false
	}

	// Check if years are within tolerance
	diff := abs(year1 - year2)
	return diff <= gd.config.DateTolerance
}

// normalizePlace normalizes a place string for comparison.
func normalizePlace(place string) string {
	// Simple normalization: lowercase, trim
	// Can be enhanced with place parsing
	return strings.ToLower(strings.TrimSpace(place))
}

// abs returns absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
