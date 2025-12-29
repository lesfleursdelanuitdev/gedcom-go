package diff

import (
	"strings"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestCompareFamily_HusbandChange tests family comparison with husband change
func TestCompareFamily_HusbandChange(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	fam1 := createTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"}, "1800", "New York")
	fam2 := createTestFamily("@F1@", "@I10@", "@I2@", []string{"@I3@"}, "1800", "New York")

	changes := differ.compareFamily(fam1, fam2)

	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}

	if changes[0].Field != "HUSB" {
		t.Errorf("expected HUSB field, got %s", changes[0].Field)
	}

	if changes[0].OldValue != "@I1@" {
		t.Errorf("expected old value @I1@, got %v", changes[0].OldValue)
	}

	if changes[0].NewValue != "@I10@" {
		t.Errorf("expected new value @I10@, got %v", changes[0].NewValue)
	}
}

// TestCompareFamily_WifeChange tests family comparison with wife change
func TestCompareFamily_WifeChange(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	fam1 := createTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"}, "1800", "New York")
	fam2 := createTestFamily("@F1@", "@I1@", "@I20@", []string{"@I3@"}, "1800", "New York")

	changes := differ.compareFamily(fam1, fam2)

	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}

	if changes[0].Field != "WIFE" {
		t.Errorf("expected WIFE field, got %s", changes[0].Field)
	}
}

// TestCompareFamily_MarriageDateChange tests family comparison with marriage date change
func TestCompareFamily_MarriageDateChange(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	fam1 := createTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"}, "1800", "New York")
	fam2 := createTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"}, "1801", "New York")

	changes := differ.compareFamily(fam1, fam2)

	// Should have one date change
	foundDateChange := false
	for _, change := range changes {
		if change.Path == "MARR.DATE" {
			foundDateChange = true
			if change.OldValue != "1800" {
				t.Errorf("expected old date 1800, got %v", change.OldValue)
			}
			if change.NewValue != "1801" {
				t.Errorf("expected new date 1801, got %v", change.NewValue)
			}
		}
	}

	if !foundDateChange {
		t.Error("expected to find MARR.DATE change")
	}
}

// TestCompareFamily_MarriagePlaceChange tests family comparison with marriage place change
func TestCompareFamily_MarriagePlaceChange(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	fam1 := createTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"}, "1800", "New York")
	fam2 := createTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"}, "1800", "Boston")

	changes := differ.compareFamily(fam1, fam2)

	// Should have one place change
	foundPlaceChange := false
	for _, change := range changes {
		if change.Path == "MARR.PLAC" {
			foundPlaceChange = true
			if change.OldValue != "New York" {
				t.Errorf("expected old place New York, got %v", change.OldValue)
			}
			if change.NewValue != "Boston" {
				t.Errorf("expected new place Boston, got %v", change.NewValue)
			}
		}
	}

	if !foundPlaceChange {
		t.Error("expected to find MARR.PLAC change")
	}
}

// TestCompareFamily_NoChanges tests family comparison with no changes
func TestCompareFamily_NoChanges(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	fam1 := createTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"}, "1800", "New York")
	fam2 := createTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"}, "1800", "New York")

	changes := differ.compareFamily(fam1, fam2)

	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

// TestCompareChildren_AddedChild tests children comparison with added child
func TestCompareChildren_AddedChild(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	children1 := []string{"@I3@", "@I4@"}
	children2 := []string{"@I3@", "@I4@", "@I5@"}

	changes := differ.compareChildren(children1, children2)

	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}

	if changes[0].Type != ChangeTypeAdded {
		t.Errorf("expected Added type, got %s", changes[0].Type)
	}

	if changes[0].NewValue != "@I5@" {
		t.Errorf("expected new value @I5@, got %v", changes[0].NewValue)
	}
}

// TestCompareChildren_RemovedChild tests children comparison with removed child
func TestCompareChildren_RemovedChild(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	children1 := []string{"@I3@", "@I4@", "@I5@"}
	children2 := []string{"@I3@", "@I4@"}

	changes := differ.compareChildren(children1, children2)

	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}

	if changes[0].Type != ChangeTypeRemoved {
		t.Errorf("expected Removed type, got %s", changes[0].Type)
	}

	if changes[0].OldValue != "@I5@" {
		t.Errorf("expected old value @I5@, got %v", changes[0].OldValue)
	}
}

// TestCompareChildren_MultipleChanges tests children comparison with multiple changes
func TestCompareChildren_MultipleChanges(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	children1 := []string{"@I3@", "@I4@", "@I5@"}
	children2 := []string{"@I3@", "@I6@", "@I7@"}

	changes := differ.compareChildren(children1, children2)

	// Should have 2 removed (@I4@, @I5@) and 2 added (@I6@, @I7@)
	if len(changes) != 4 {
		t.Errorf("expected 4 changes, got %d", len(changes))
	}

	removedCount := 0
	addedCount := 0
	for _, change := range changes {
		if change.Type == ChangeTypeRemoved {
			removedCount++
		} else if change.Type == ChangeTypeAdded {
			addedCount++
		}
	}

	if removedCount != 2 {
		t.Errorf("expected 2 removed, got %d", removedCount)
	}

	if addedCount != 2 {
		t.Errorf("expected 2 added, got %d", addedCount)
	}
}

// TestCompareChildren_NoChanges tests children comparison with no changes
func TestCompareChildren_NoChanges(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	children1 := []string{"@I3@", "@I4@"}
	children2 := []string{"@I3@", "@I4@"}

	changes := differ.compareChildren(children1, children2)

	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

// TestCompareChildren_EmptyLists tests children comparison with empty lists
func TestCompareChildren_EmptyLists(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	children1 := []string{}
	children2 := []string{}

	changes := differ.compareChildren(children1, children2)

	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

// TestCompareDate_SameDate tests date comparison with identical dates
func TestCompareDate_SameDate(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	change := differ.compareDate("1800", "1800", "BIRT.DATE")

	if change != nil {
		t.Error("expected nil for identical dates")
	}
}

// TestCompareDate_DifferentDate tests date comparison with different dates
func TestCompareDate_DifferentDate(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	change := differ.compareDate("1800", "1900", "BIRT.DATE")

	if change == nil {
		t.Fatal("expected non-nil change for different dates")
	}

	if change.Type != ChangeTypeModified {
		t.Errorf("expected Modified type, got %s", change.Type)
	}

	if change.OldValue != "1800" {
		t.Errorf("expected old value 1800, got %v", change.OldValue)
	}

	if change.NewValue != "1900" {
		t.Errorf("expected new value 1900, got %v", change.NewValue)
	}
}

// TestCompareDate_SemanticallyEquivalent tests date comparison with semantically equivalent dates
func TestCompareDate_SemanticallyEquivalent(t *testing.T) {
	config := DefaultConfig()
	config.DateTolerance = 2
	differ := NewGedcomDiffer(config)

	// Dates within tolerance should be semantically equivalent
	change := differ.compareDate("1800", "1801", "BIRT.DATE")

	if change == nil {
		t.Fatal("expected non-nil change")
	}

	if change.Type != ChangeTypeSemanticallyEquivalent {
		t.Errorf("expected SemanticallyEquivalent type, got %s", change.Type)
	}
}

// TestCompareDate_EmptyDates tests date comparison with empty dates
func TestCompareDate_EmptyDates(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	// Both empty
	change := differ.compareDate("", "", "BIRT.DATE")
	if change != nil {
		t.Error("expected nil for both empty dates")
	}

	// One empty
	change = differ.compareDate("", "1800", "BIRT.DATE")
	if change == nil {
		t.Fatal("expected non-nil change when one date is empty")
	}

	if change.Type != ChangeTypeModified {
		t.Errorf("expected Modified type, got %s", change.Type)
	}
}

// TestComparePlace_SamePlace tests place comparison with identical places
func TestComparePlace_SamePlace(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	change := differ.comparePlace("New York", "New York", "BIRT.PLAC")

	if change != nil {
		t.Error("expected nil for identical places")
	}
}

// TestComparePlace_DifferentPlace tests place comparison with different places
func TestComparePlace_DifferentPlace(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	change := differ.comparePlace("New York", "Boston", "BIRT.PLAC")

	if change == nil {
		t.Fatal("expected non-nil change for different places")
	}

	if change.Type != ChangeTypeModified {
		t.Errorf("expected Modified type, got %s", change.Type)
	}
}

// TestComparePlace_SemanticallyEquivalent tests place comparison with semantically equivalent places
func TestComparePlace_SemanticallyEquivalent(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	// Normalized places should be equivalent (case-insensitive, trimmed)
	change := differ.comparePlace("New York", "new york", "BIRT.PLAC")

	if change == nil {
		t.Fatal("expected non-nil change")
	}

	if change.Type != ChangeTypeSemanticallyEquivalent {
		t.Errorf("expected SemanticallyEquivalent type, got %s", change.Type)
	}
}

// TestComparePlace_WithWhitespace tests place comparison with whitespace differences
func TestComparePlace_WithWhitespace(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	// Places with different whitespace should be semantically equivalent
	change := differ.comparePlace("New York", "  New York  ", "BIRT.PLAC")

	if change == nil {
		t.Fatal("expected non-nil change")
	}

	if change.Type != ChangeTypeSemanticallyEquivalent {
		t.Errorf("expected SemanticallyEquivalent type, got %s", change.Type)
	}
}

// TestAreDatesSemanticallyEquivalent_WithinTolerance tests date semantic equivalence within tolerance
func TestAreDatesSemanticallyEquivalent_WithinTolerance(t *testing.T) {
	config := DefaultConfig()
	config.DateTolerance = 2
	differ := NewGedcomDiffer(config)

	// Dates within 2 years should be equivalent
	if !differ.areDatesSemanticallyEquivalent("1800", "1801") {
		t.Error("expected dates within tolerance to be equivalent")
	}

	if !differ.areDatesSemanticallyEquivalent("1800", "1802") {
		t.Error("expected dates within tolerance to be equivalent")
	}
}

// TestAreDatesSemanticallyEquivalent_OutsideTolerance tests date semantic equivalence outside tolerance
func TestAreDatesSemanticallyEquivalent_OutsideTolerance(t *testing.T) {
	config := DefaultConfig()
	config.DateTolerance = 2
	differ := NewGedcomDiffer(config)

	// Dates outside tolerance should not be equivalent
	if differ.areDatesSemanticallyEquivalent("1800", "1803") {
		t.Error("expected dates outside tolerance to not be equivalent")
	}
}

// TestAreDatesSemanticallyEquivalent_EmptyDates tests date semantic equivalence with empty dates
func TestAreDatesSemanticallyEquivalent_EmptyDates(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	// Empty dates should not be equivalent
	if differ.areDatesSemanticallyEquivalent("", "1800") {
		t.Error("expected empty date to not be equivalent")
	}

	if differ.areDatesSemanticallyEquivalent("1800", "") {
		t.Error("expected empty date to not be equivalent")
	}
}

// TestAreDatesSemanticallyEquivalent_InvalidDates tests date semantic equivalence with invalid dates
func TestAreDatesSemanticallyEquivalent_InvalidDates(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	// Invalid dates should not be equivalent
	if differ.areDatesSemanticallyEquivalent("invalid", "1800") {
		t.Error("expected invalid date to not be equivalent")
	}
}

// TestNormalizePlace tests place normalization
func TestNormalizePlace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"New York", "new york"},
		{"  New York  ", "new york"},
		{"BOSTON", "boston"},
		{"  Boston  ", "boston"},
		{"", ""},
	}

	for _, tt := range tests {
		result := normalizePlace(tt.input)
		if result != tt.expected {
			t.Errorf("normalizePlace(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// TestCompareBasicRecord tests basic record comparison
func TestCompareBasicRecord(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	// Create two basic records
	record1 := types.NewIndividualRecord(types.NewGedcomLine(0, "INDI", "", "@I1@"))
	record2 := types.NewIndividualRecord(types.NewGedcomLine(0, "INDI", "", "@I2@"))

	changes := differ.compareBasicRecord(record1, record2)

	// compareBasicRecord currently returns empty changes (stub implementation)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for basic record comparison, got %d", len(changes))
	}
}

// TestWriteRemovedRecords tests report generation for removed records
func TestWriteRemovedRecords(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	// Create a removed record
	removed := []RecordDiff{
		{
			Xref:   "@I1@",
			Type:   "INDI",
			Record: createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York"),
		},
		{
			Xref:   "@I2@",
			Type:   "INDI",
			Record: createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston"),
		},
	}

	var sb strings.Builder
	differ.writeRemovedRecords(&sb, removed)

	report := sb.String()

	if report == "" {
		t.Error("expected non-empty report")
	}

	if !strings.Contains(report, "Removed Records:") {
		t.Error("expected report to contain 'Removed Records:'")
	}

	if !strings.Contains(report, "@I1@") {
		t.Error("expected report to contain @I1@")
	}

	if !strings.Contains(report, "@I2@") {
		t.Error("expected report to contain @I2@")
	}

	if !strings.Contains(report, "John /Doe/") {
		t.Error("expected report to contain individual name")
	}
}

// Helper function to create test family
func createTestFamily(xref, husband, wife string, children []string, marriageDate, marriagePlace string) *types.FamilyRecord {
	line := types.NewGedcomLine(0, "FAM", "", xref)
	fam := types.NewFamilyRecord(line)

	// Add husband
	if husband != "" {
		husbLine := types.NewGedcomLine(1, "HUSB", husband, "")
		line.AddChild(husbLine)
	}

	// Add wife
	if wife != "" {
		wifeLine := types.NewGedcomLine(1, "WIFE", wife, "")
		line.AddChild(wifeLine)
	}

	// Add children
	for _, child := range children {
		childLine := types.NewGedcomLine(1, "CHIL", child, "")
		line.AddChild(childLine)
	}

	// Add marriage date and place
	if marriageDate != "" || marriagePlace != "" {
		marrLine := types.NewGedcomLine(1, "MARR", "", "")
		line.AddChild(marrLine)

		if marriageDate != "" {
			dateLine := types.NewGedcomLine(2, "DATE", marriageDate, "")
			marrLine.AddChild(dateLine)
		}

		if marriagePlace != "" {
			placLine := types.NewGedcomLine(2, "PLAC", marriagePlace, "")
			marrLine.AddChild(placLine)
		}
	}

	return fam
}

