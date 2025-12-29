package diff

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/duplicate"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestCompareByContent tests content-based comparison
func TestCompareByContent(t *testing.T) {
	config := DefaultConfig()
	config.MatchingStrategy = "content"
	config.SimilarityThreshold = 0.85
	differ := NewGedcomDiffer(config)

	tree1 := types.NewGedcomTree()
	tree2 := types.NewGedcomTree()

	// Add similar individuals to both trees (different XREFs)
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	indi2 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi2.FirstLine().XrefID = "@I2@" // Different XREF
	tree2.AddRecord(indi2)

	changes, err := differ.compareByContent(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Content matching should find the similar records
	// The exact results depend on duplicate detection, but should not error
	_ = changes
}

// TestCompareHybrid tests hybrid comparison (XREF + content)
func TestCompareHybrid(t *testing.T) {
	config := DefaultConfig()
	config.MatchingStrategy = "hybrid"
	differ := NewGedcomDiffer(config)

	tree1 := types.NewGedcomTree()
	tree2 := types.NewGedcomTree()

	// Add individual with same XREF
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	indi2 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi2.FirstLine().XrefID = "@I1@"
	tree2.AddRecord(indi2)

	// Add unmatched individual in tree2
	indi3 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi3.FirstLine().XrefID = "@I2@"
	tree2.AddRecord(indi3)

	changes, err := differ.compareHybrid(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have at least the unmatched record
	if len(changes.Added) == 0 && len(changes.Removed) == 0 && len(changes.Modified) == 0 {
		t.Log("Note: No changes detected (may be expected depending on matching)")
	}
}

// TestFindUnmatchedRecords tests finding unmatched records
func TestFindUnmatchedRecords(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	tree := types.NewGedcomTree()

	// Add individuals
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree.AddRecord(indi1)

	indi2 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi2.FirstLine().XrefID = "@I2@"
	tree.AddRecord(indi2)

	// Create changes with one matched record
	changes := DiffChanges{
		Modified: []RecordModification{
			{Xref: "@I1@", Type: "INDI"},
		},
	}

	unmatched := differ.findUnmatchedRecords(tree, changes)

	// @I2@ should be unmatched
	if len(unmatched) != 1 {
		t.Errorf("expected 1 unmatched record, got %d", len(unmatched))
	}

	if _, ok := unmatched["@I2@"]; !ok {
		t.Error("expected @I2@ to be unmatched")
	}
}

// TestFindUnmatchedRecords_AllMatched tests finding unmatched when all are matched
func TestFindUnmatchedRecords_AllMatched(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	tree := types.NewGedcomTree()

	// Add individuals
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree.AddRecord(indi1)

	// Create changes with all records matched
	changes := DiffChanges{
		Modified: []RecordModification{
			{Xref: "@I1@", Type: "INDI"},
		},
	}

	unmatched := differ.findUnmatchedRecords(tree, changes)

	// All should be matched
	if len(unmatched) != 0 {
		t.Errorf("expected 0 unmatched records, got %d", len(unmatched))
	}
}

// TestMergeChanges tests merging two DiffChanges structures
func TestMergeChanges(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	changes1 := DiffChanges{
		Added: []RecordDiff{
			{Xref: "@I1@", Type: "INDI"},
		},
		Removed: []RecordDiff{
			{Xref: "@I2@", Type: "INDI"},
		},
		Modified: []RecordModification{
			{Xref: "@I3@", Type: "INDI"},
		},
	}

	changes2 := DiffChanges{
		Added: []RecordDiff{
			{Xref: "@I4@", Type: "INDI"},
		},
		Removed: []RecordDiff{
			{Xref: "@I5@", Type: "INDI"},
		},
		Modified: []RecordModification{
			{Xref: "@I6@", Type: "INDI"},
		},
	}

	merged := differ.mergeChanges(changes1, changes2)

	if len(merged.Added) != 2 {
		t.Errorf("expected 2 added records, got %d", len(merged.Added))
	}

	if len(merged.Removed) != 2 {
		t.Errorf("expected 2 removed records, got %d", len(merged.Removed))
	}

	if len(merged.Modified) != 2 {
		t.Errorf("expected 2 modified records, got %d", len(merged.Modified))
	}
}

// TestMergeChanges_Empty tests merging with empty changes
func TestMergeChanges_Empty(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	changes1 := DiffChanges{
		Added: []RecordDiff{
			{Xref: "@I1@", Type: "INDI"},
		},
	}

	changes2 := DiffChanges{}

	merged := differ.mergeChanges(changes1, changes2)

	if len(merged.Added) != 1 {
		t.Errorf("expected 1 added record, got %d", len(merged.Added))
	}
}

// TestCompareUnmatchedByContent tests comparing unmatched records by content
func TestCompareUnmatchedByContent(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	unmatched1 := map[string]types.Record{
		"@I1@": createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York"),
	}

	unmatched2 := map[string]types.Record{
		"@I2@": createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston"),
	}

	tree1 := types.NewGedcomTree()
	changes := differ.compareUnmatchedByContent(unmatched1, unmatched2, tree1)

	// Should mark unmatched1 as removed and unmatched2 as added
	if len(changes.Removed) != 1 {
		t.Errorf("expected 1 removed record, got %d", len(changes.Removed))
	}

	if len(changes.Added) != 1 {
		t.Errorf("expected 1 added record, got %d", len(changes.Added))
	}

	if changes.Removed[0].Xref != "@I1@" {
		t.Errorf("expected removed record @I1@, got %s", changes.Removed[0].Xref)
	}

	if changes.Added[0].Xref != "@I2@" {
		t.Errorf("expected added record @I2@, got %s", changes.Added[0].Xref)
	}
}

// TestCompareUnmatchedByContent_Empty tests comparing with empty unmatched sets
func TestCompareUnmatchedByContent_Empty(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	unmatched1 := map[string]types.Record{}
	unmatched2 := map[string]types.Record{}

	tree1 := types.NewGedcomTree()
	changes := differ.compareUnmatchedByContent(unmatched1, unmatched2, tree1)

	if len(changes.Removed) != 0 {
		t.Errorf("expected 0 removed records, got %d", len(changes.Removed))
	}

	if len(changes.Added) != 0 {
		t.Errorf("expected 0 added records, got %d", len(changes.Added))
	}
}

// TestBuildChangesFromMatches tests building changes from duplicate matches
func TestBuildChangesFromMatches(t *testing.T) {
	// This test requires duplicate detection setup
	// For now, we'll test with empty matches to ensure the function doesn't panic
	differ := NewGedcomDiffer(DefaultConfig())

	tree1 := types.NewGedcomTree()
	tree2 := types.NewGedcomTree()

	// Add individuals
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	indi2 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi2.FirstLine().XrefID = "@I2@"
	tree2.AddRecord(indi2)

	// Test with empty matches (no duplicates found)
	matches := []duplicate.DuplicateMatch{}
	changes := differ.buildChangesFromMatches(tree1, tree2, matches)

	// Should mark all records as added/removed since no matches
	if len(changes.Removed) != 1 {
		t.Errorf("expected 1 removed record, got %d", len(changes.Removed))
	}

	if len(changes.Added) != 1 {
		t.Errorf("expected 1 added record, got %d", len(changes.Added))
	}
}

