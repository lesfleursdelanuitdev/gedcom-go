package diff

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/duplicate"
)

// compareByContent performs content-based comparison using duplicate detection.
func (gd *GedcomDiffer) compareByContent(tree1, tree2 *gedcom.GedcomTree) (DiffChanges, error) {
	// Use duplicate detection for content matching
	config := duplicate.DefaultConfig()
	config.MinThreshold = gd.config.SimilarityThreshold
	detector := duplicate.NewDuplicateDetector(config)
	detector.SetTree(tree1) // Use tree1 as base for relationship matching

	// Find potential matches
	matches, err := detector.FindDuplicatesBetween(tree1, tree2)
	if err != nil {
		return DiffChanges{}, err
	}

	// Build changes from matches
	changes := gd.buildChangesFromMatches(tree1, tree2, matches.Matches)

	return changes, nil
}

// compareHybrid performs hybrid comparison (XREF first, content fallback).
func (gd *GedcomDiffer) compareHybrid(tree1, tree2 *gedcom.GedcomTree) (DiffChanges, error) {
	// First, do XREF-based comparison
	xrefChanges, err := gd.compareByXref(tree1, tree2)
	if err != nil {
		return DiffChanges{}, err
	}

	// Find unmatched records
	unmatched1 := gd.findUnmatchedRecords(tree1, xrefChanges)
	unmatched2 := gd.findUnmatchedRecords(tree2, xrefChanges)

	// Use content matching for unmatched records
	if len(unmatched1) > 0 && len(unmatched2) > 0 {
		contentChanges := gd.compareUnmatchedByContent(unmatched1, unmatched2, tree1)

		// Merge content changes into XREF changes
		xrefChanges = gd.mergeChanges(xrefChanges, contentChanges)
	}

	return xrefChanges, nil
}

// buildChangesFromMatches builds DiffChanges from duplicate matches.
func (gd *GedcomDiffer) buildChangesFromMatches(
	tree1, tree2 *gedcom.GedcomTree,
	matches []duplicate.DuplicateMatch,
) DiffChanges {
	changes := DiffChanges{
		Added:    make([]RecordDiff, 0),
		Removed:  make([]RecordDiff, 0),
		Modified: make([]RecordModification, 0),
	}

	// Track matched records
	matched1 := make(map[string]bool)
	matched2 := make(map[string]bool)

	// Process matches as modifications
	for _, match := range matches {
		if match.SimilarityScore >= gd.config.SimilarityThreshold {
			matched1[match.Individual1.XrefID()] = true
			matched2[match.Individual2.XrefID()] = true

			// Compare the matched records
			modification := gd.compareRecordContent(
				match.Individual1,
				match.Individual2,
				match.Individual1.XrefID(),
				"INDI",
			)
			if modification != nil {
				// Update XREF to show both
				modification.Xref = match.Individual1.XrefID() + " → " + match.Individual2.XrefID()
				changes.Modified = append(changes.Modified, *modification)
			}
		}
	}

	// Find unmatched records in tree1 (removed)
	allIndi1 := tree1.GetAllIndividuals()
	for xref, record := range allIndi1 {
		if !matched1[xref] {
			history := []ChangeHistory{
				gd.createChangeHistory(ChangeTypeRemoved, "INDI", xref, ""),
			}
			changes.Removed = append(changes.Removed, RecordDiff{
				Xref:    xref,
				Type:    "INDI",
				Record:  record,
				History: history,
			})
		}
	}

	// Find unmatched records in tree2 (added)
	allIndi2 := tree2.GetAllIndividuals()
	for xref, record := range allIndi2 {
		if !matched2[xref] {
			history := []ChangeHistory{
				gd.createChangeHistory(ChangeTypeAdded, "INDI", "", xref),
			}
			changes.Added = append(changes.Added, RecordDiff{
				Xref:    xref,
				Type:    "INDI",
				Record:  record,
				History: history,
			})
		}
	}

	return changes
}

// findUnmatchedRecords finds records that weren't matched in XREF comparison.
func (gd *GedcomDiffer) findUnmatchedRecords(
	tree *gedcom.GedcomTree,
	changes DiffChanges,
) map[string]gedcom.Record {
	unmatched := make(map[string]gedcom.Record)

	// Get all individuals
	allIndi := tree.GetAllIndividuals()

	// Track matched XREFs
	matched := make(map[string]bool)
	for _, mod := range changes.Modified {
		matched[mod.Xref] = true
	}

	// Find unmatched
	for xref, record := range allIndi {
		if !matched[xref] {
			unmatched[xref] = record
		}
	}

	return unmatched
}

// compareUnmatchedByContent compares unmatched records using content matching.
func (gd *GedcomDiffer) compareUnmatchedByContent(
	unmatched1, unmatched2 map[string]gedcom.Record,
	tree1 *gedcom.GedcomTree,
) DiffChanges {
	// Create temporary trees for comparison
	// This is a simplified approach - could be optimized
	changes := DiffChanges{
		Added:    make([]RecordDiff, 0),
		Removed:  make([]RecordDiff, 0),
		Modified: make([]RecordModification, 0),
	}

	// For now, mark all unmatched1 as removed and unmatched2 as added
	// Full content matching would require building a temporary tree
	for xref, record := range unmatched1 {
		history := []ChangeHistory{
			gd.createChangeHistory(ChangeTypeRemoved, "INDI", xref, ""),
		}
		changes.Removed = append(changes.Removed, RecordDiff{
			Xref:    xref,
			Type:    "INDI",
			Record:  record,
			History: history,
		})
	}

	for xref, record := range unmatched2 {
		history := []ChangeHistory{
			gd.createChangeHistory(ChangeTypeAdded, "INDI", "", xref),
		}
		changes.Added = append(changes.Added, RecordDiff{
			Xref:    xref,
			Type:    "INDI",
			Record:  record,
			History: history,
		})
	}

	return changes
}

// mergeChanges merges two DiffChanges structures.
func (gd *GedcomDiffer) mergeChanges(changes1, changes2 DiffChanges) DiffChanges {
	return DiffChanges{
		Added:    append(changes1.Added, changes2.Added...),
		Removed:  append(changes1.Removed, changes2.Removed...),
		Modified: append(changes1.Modified, changes2.Modified...),
	}
}
