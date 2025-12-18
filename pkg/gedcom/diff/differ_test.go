package diff

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestCompareByXref_AddedRecord(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	tree1 := gedcom.NewGedcomTree()
	tree2 := gedcom.NewGedcomTree()

	// Add individual to tree1
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	// Add same individual plus new one to tree2
	indi1Copy := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1Copy.FirstLine().XrefID = "@I1@"
	tree2.AddRecord(indi1Copy)

	indi2 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi2.FirstLine().XrefID = "@I2@"
	tree2.AddRecord(indi2)

	result, err := differ.Compare(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find one added record
	if len(result.Changes.Added) != 1 {
		t.Errorf("expected 1 added record, got %d", len(result.Changes.Added))
	}

	if result.Changes.Added[0].Xref != "@I2@" {
		t.Errorf("expected added record @I2@, got %s", result.Changes.Added[0].Xref)
	}
}

func TestCompareByXref_RemovedRecord(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	tree1 := gedcom.NewGedcomTree()
	tree2 := gedcom.NewGedcomTree()

	// Add two individuals to tree1
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	indi2 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi2.FirstLine().XrefID = "@I2@"
	tree1.AddRecord(indi2)

	// Add only one to tree2
	indi1Copy := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1Copy.FirstLine().XrefID = "@I1@"
	tree2.AddRecord(indi1Copy)

	result, err := differ.Compare(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find one removed record
	if len(result.Changes.Removed) != 1 {
		t.Errorf("expected 1 removed record, got %d", len(result.Changes.Removed))
	}

	if result.Changes.Removed[0].Xref != "@I2@" {
		t.Errorf("expected removed record @I2@, got %s", result.Changes.Removed[0].Xref)
	}
}

func TestCompareByXref_ModifiedRecord(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	tree1 := gedcom.NewGedcomTree()
	tree2 := gedcom.NewGedcomTree()

	// Add individual to tree1
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	// Add modified individual to tree2
	indi2 := createTestIndividual("John /Doe Jr/", "John", "Doe", "1800", "New York")
	indi2.FirstLine().XrefID = "@I1@"
	tree2.AddRecord(indi2)

	result, err := differ.Compare(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find one modified record
	if len(result.Changes.Modified) != 1 {
		t.Errorf("expected 1 modified record, got %d", len(result.Changes.Modified))
	}

	mod := result.Changes.Modified[0]
	if mod.Xref != "@I1@" {
		t.Errorf("expected modified record @I1@, got %s", mod.Xref)
	}

	// Should have name change
	foundNameChange := false
	for _, change := range mod.Changes {
		if change.Field == "NAME" {
			foundNameChange = true
			if change.OldValue != "John /Doe/" {
				t.Errorf("expected old name 'John /Doe/', got %v", change.OldValue)
			}
			if change.NewValue != "John /Doe Jr/" {
				t.Errorf("expected new name 'John /Doe Jr/', got %v", change.NewValue)
			}
		}
	}

	if !foundNameChange {
		t.Error("expected to find NAME change")
	}
}

func TestCompareByXref_SemanticEquivalence(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	tree1 := gedcom.NewGedcomTree()
	tree2 := gedcom.NewGedcomTree()

	// Add individual with exact date
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	// Add individual with semantically equivalent date
	indi2 := createTestIndividual("John /Doe/", "John", "Doe", "ABT 1800", "New York")
	indi2.FirstLine().XrefID = "@I1@"
	tree2.AddRecord(indi2)

	result, err := differ.Compare(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find one modified record
	if len(result.Changes.Modified) != 1 {
		t.Errorf("expected 1 modified record, got %d", len(result.Changes.Modified))
	}

	mod := result.Changes.Modified[0]

	// Should have date change marked as semantically equivalent
	foundDateChange := false
	for _, change := range mod.Changes {
		if change.Path == "BIRT.DATE" {
			foundDateChange = true
			if change.Type != ChangeTypeSemanticallyEquivalent {
				t.Errorf("expected semantically equivalent change, got %s", change.Type)
			}
		}
	}

	if !foundDateChange {
		t.Error("expected to find BIRT.DATE change")
	}
}

func TestChangeHistory(t *testing.T) {
	config := DefaultConfig()
	config.TrackHistory = true
	differ := NewGedcomDiffer(config)

	tree1 := gedcom.NewGedcomTree()
	tree2 := gedcom.NewGedcomTree()

	// Add individual to tree1
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	// Add modified individual to tree2
	indi2 := createTestIndividual("John /Doe Jr/", "John", "Doe", "1800", "New York")
	indi2.FirstLine().XrefID = "@I1@"
	tree2.AddRecord(indi2)

	result, err := differ.Compare(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have change history
	if len(result.History) == 0 {
		t.Error("expected change history to be populated")
	}

	// Check that history entries have timestamps
	for _, entry := range result.History {
		if entry.Timestamp.IsZero() {
			t.Error("expected history entry to have timestamp")
		}
		if entry.Field == "" {
			t.Error("expected history entry to have field")
		}
	}
}

func TestGenerateReport(t *testing.T) {
	differ := NewGedcomDiffer(DefaultConfig())

	tree1 := gedcom.NewGedcomTree()
	tree2 := gedcom.NewGedcomTree()

	// Add individuals
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	indi2 := createTestIndividual("John /Doe Jr/", "John", "Doe", "1800", "New York")
	indi2.FirstLine().XrefID = "@I1@"
	tree2.AddRecord(indi2)

	indi3 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi3.FirstLine().XrefID = "@I2@"
	tree2.AddRecord(indi3)

	result, err := differ.Compare(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report, err := differ.GenerateReport(result)
	if err != nil {
		t.Fatalf("unexpected error generating report: %v", err)
	}

	if report == "" {
		t.Error("expected report to be non-empty")
	}

	// Check that report contains expected sections
	if !contains(report, "Summary:") {
		t.Error("expected report to contain 'Summary:'")
	}
	if !contains(report, "Modified Records:") {
		t.Error("expected report to contain 'Modified Records:'")
	}
	if !contains(report, "Added Records:") {
		t.Error("expected report to contain 'Added Records:'")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to create test individual (reuse from duplicate tests)
func createTestIndividual(name, givenName, surname, birthDate, birthPlace string) *gedcom.IndividualRecord {
	line := gedcom.NewGedcomLine(0, "INDI", "", "@TEST@")
	indi := gedcom.NewIndividualRecord(line)

	// Set name
	if name != "" {
		nameLine := gedcom.NewGedcomLine(1, "NAME", name, "")
		line.AddChild(nameLine)

		// Set given name
		if givenName != "" {
			givnLine := gedcom.NewGedcomLine(2, "GIVN", givenName, "")
			nameLine.AddChild(givnLine)
		}

		// Set surname
		if surname != "" {
			surnLine := gedcom.NewGedcomLine(2, "SURN", surname, "")
			nameLine.AddChild(surnLine)
		}
	}

	// Set birth date and place
	if birthDate != "" || birthPlace != "" {
		birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
		line.AddChild(birtLine)

		if birthDate != "" {
			dateLine := gedcom.NewGedcomLine(2, "DATE", birthDate, "")
			birtLine.AddChild(dateLine)
		}

		if birthPlace != "" {
			placLine := gedcom.NewGedcomLine(2, "PLAC", birthPlace, "")
			birtLine.AddChild(placLine)
		}
	}

	return indi
}
