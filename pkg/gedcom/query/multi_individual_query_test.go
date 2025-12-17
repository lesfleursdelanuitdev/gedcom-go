package query

import (
	"testing"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

func TestMultiIndividualQuery_Ancestors(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create grandparent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create two parents
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Create two children
	indi4Line := gedcom.NewGedcomLine(0, "INDI", "", "@I4@")
	indi4Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F3@", ""))
	indi4 := gedcom.NewIndividualRecord(indi4Line)
	tree.AddRecord(indi4)

	indi5Line := gedcom.NewGedcomLine(0, "INDI", "", "@I5@")
	indi5Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F4@", ""))
	indi5 := gedcom.NewIndividualRecord(indi5Line)
	tree.AddRecord(indi5)

	// Family 1: I1 is parent of I2
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I1 is parent of I3
	fam2Line := gedcom.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2 := gedcom.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	// Family 3: I2 is parent of I4
	fam3Line := gedcom.NewGedcomLine(0, "FAM", "", "@F3@")
	fam3Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam3Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I4@", ""))
	fam3 := gedcom.NewFamilyRecord(fam3Line)
	tree.AddRecord(fam3)

	// Family 4: I3 is parent of I5
	fam4Line := gedcom.NewGedcomLine(0, "FAM", "", "@F4@")
	fam4Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I3@", ""))
	fam4Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I5@", ""))
	fam4 := gedcom.NewFamilyRecord(fam4Line)
	tree.AddRecord(fam4)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Find ancestors of both I4 and I5
	ancestors, err := query.Individuals("@I4@", "@I5@").Ancestors()
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}

	// Should include I1, I2, I3 (union of ancestors)
	if len(ancestors) < 3 {
		t.Errorf("Expected at least 3 ancestors, got %d", len(ancestors))
	}
}

func TestMultiIndividualQuery_CommonAncestors(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create grandparent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create two siblings
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family: I1 is parent of I2 and I3
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Find common ancestors of I2 and I3
	common, err := query.Individuals("@I2@", "@I3@").CommonAncestors()
	if err != nil {
		t.Fatalf("Failed to get common ancestors: %v", err)
	}

	// Should include I1 (their parent)
	if len(common) == 0 {
		t.Error("Expected at least one common ancestor")
	}

	foundI1 := false
	for _, ancestor := range common {
		if ancestor.XrefID() == "@I1@" {
			foundI1 = true
			break
		}
	}

	if !foundI1 {
		t.Error("Expected @I1@ to be a common ancestor")
	}
}

func TestMultiIndividualQuery_Execute(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Individuals("@I1@", "@I2@").Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestMultiIndividualQuery_Count(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	count := query.Individuals("@I1@", "@I2@").Count()
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}
