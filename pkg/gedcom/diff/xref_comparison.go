package diff

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// compareByXref performs XREF-based comparison.
func (gd *GedcomDiffer) compareByXref(tree1, tree2 *gedcom.GedcomTree) (DiffChanges, error) {
	changes := DiffChanges{
		Added:    make([]RecordDiff, 0),
		Removed:  make([]RecordDiff, 0),
		Modified: make([]RecordModification, 0),
	}

	// Compare individuals
	indiChanges := gd.compareRecordsByXref(
		tree1.GetAllIndividuals(),
		tree2.GetAllIndividuals(),
		"INDI",
	)
	changes.Added = append(changes.Added, indiChanges.Added...)
	changes.Removed = append(changes.Removed, indiChanges.Removed...)
	changes.Modified = append(changes.Modified, indiChanges.Modified...)

	// Compare families
	famChanges := gd.compareRecordsByXref(
		tree1.GetAllFamilies(),
		tree2.GetAllFamilies(),
		"FAM",
	)
	changes.Added = append(changes.Added, famChanges.Added...)
	changes.Removed = append(changes.Removed, famChanges.Removed...)
	changes.Modified = append(changes.Modified, famChanges.Modified...)

	// Compare notes
	noteChanges := gd.compareRecordsByXref(
		tree1.GetAllNotes(),
		tree2.GetAllNotes(),
		"NOTE",
	)
	changes.Added = append(changes.Added, noteChanges.Added...)
	changes.Removed = append(changes.Removed, noteChanges.Removed...)
	changes.Modified = append(changes.Modified, noteChanges.Modified...)

	// Compare sources
	sourceChanges := gd.compareRecordsByXref(
		tree1.GetAllSources(),
		tree2.GetAllSources(),
		"SOUR",
	)
	changes.Added = append(changes.Added, sourceChanges.Added...)
	changes.Removed = append(changes.Removed, sourceChanges.Removed...)
	changes.Modified = append(changes.Modified, sourceChanges.Modified...)

	// Compare repositories
	repoChanges := gd.compareRecordsByXref(
		tree1.GetAllRepositories(),
		tree2.GetAllRepositories(),
		"REPO",
	)
	changes.Added = append(changes.Added, repoChanges.Added...)
	changes.Removed = append(changes.Removed, repoChanges.Removed...)
	changes.Modified = append(changes.Modified, repoChanges.Modified...)

	// Compare submitters
	submChanges := gd.compareRecordsByXref(
		tree1.GetAllSubmitters(),
		tree2.GetAllSubmitters(),
		"SUBM",
	)
	changes.Added = append(changes.Added, submChanges.Added...)
	changes.Removed = append(changes.Removed, submChanges.Removed...)
	changes.Modified = append(changes.Modified, submChanges.Modified...)

	// Compare multimedia
	multChanges := gd.compareRecordsByXref(
		tree1.GetAllMultimedia(),
		tree2.GetAllMultimedia(),
		"OBJE",
	)
	changes.Added = append(changes.Added, multChanges.Added...)
	changes.Removed = append(changes.Removed, multChanges.Removed...)
	changes.Modified = append(changes.Modified, multChanges.Modified...)

	return changes, nil
}

// compareRecordsByXref compares records of the same type by XREF.
func (gd *GedcomDiffer) compareRecordsByXref(
	records1, records2 map[string]gedcom.Record,
	recordType string,
) DiffChanges {
	changes := DiffChanges{
		Added:    make([]RecordDiff, 0),
		Removed:  make([]RecordDiff, 0),
		Modified: make([]RecordModification, 0),
	}

	// Find added and modified records
	for xref, record2 := range records2 {
		record1, exists := records1[xref]
		if !exists {
			// Added record
			history := []ChangeHistory{
				gd.createChangeHistory(ChangeTypeAdded, recordType, "", xref),
			}
			changes.Added = append(changes.Added, RecordDiff{
				Xref:    xref,
				Type:    recordType,
				Record:  record2,
				History: history,
			})
		} else {
			// Check if modified
			modification := gd.compareRecordContent(record1, record2, xref, recordType)
			if modification != nil && len(modification.Changes) > 0 {
				changes.Modified = append(changes.Modified, *modification)
			}
		}
	}

	// Find removed records
	for xref, record1 := range records1 {
		_, exists := records2[xref]
		if !exists {
			// Removed record
			history := []ChangeHistory{
				gd.createChangeHistory(ChangeTypeRemoved, recordType, xref, ""),
			}
			changes.Removed = append(changes.Removed, RecordDiff{
				Xref:    xref,
				Type:    recordType,
				Record:  record1,
				History: history,
			})
		}
	}

	return changes
}

// compareRecordContent compares the content of two records with the same XREF.
func (gd *GedcomDiffer) compareRecordContent(
	record1, record2 gedcom.Record,
	xref, recordType string,
) *RecordModification {
	changes := make([]FieldChange, 0)
	history := make([]ChangeHistory, 0)

	// Compare based on record type
	switch recordType {
	case "INDI":
		indi1, ok1 := record1.(*gedcom.IndividualRecord)
		indi2, ok2 := record2.(*gedcom.IndividualRecord)
		if ok1 && ok2 {
			fieldChanges := gd.compareIndividual(indi1, indi2)
			changes = append(changes, fieldChanges...)
		}
	case "FAM":
		fam1, ok1 := record1.(*gedcom.FamilyRecord)
		fam2, ok2 := record2.(*gedcom.FamilyRecord)
		if ok1 && ok2 {
			fieldChanges := gd.compareFamily(fam1, fam2)
			changes = append(changes, fieldChanges...)
		}
	case "NOTE", "SOUR", "REPO", "SUBM", "OBJE":
		// For other record types, compare basic fields
		fieldChanges := gd.compareBasicRecord(record1, record2)
		changes = append(changes, fieldChanges...)
	}

	// If no changes, return nil
	if len(changes) == 0 {
		return nil
	}

	// Collect history from field changes
	for _, change := range changes {
		history = append(history, change.History...)
	}

	// Add overall record modification history
	if gd.config.TrackHistory {
		history = append(history, gd.createChangeHistory(
			ChangeTypeModified,
			recordType,
			xref,
			xref,
		))
	}

	return &RecordModification{
		Xref:    xref,
		Type:    recordType,
		Changes: changes,
		History: history,
	}
}
