package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestFamilyQuery_Husband(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	husband, err := query.Family("@F1@").Husband()
	if err != nil {
		t.Fatalf("Failed to get husband: %v", err)
	}

	if husband == nil {
		t.Fatal("Expected husband to be found")
	}

	if husband.XrefID() != "@I1@" {
		t.Errorf("Expected husband @I1@, got %s", husband.XrefID())
	}
}

func TestFamilyQuery_Wife(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I1@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	wife, err := query.Family("@F1@").Wife()
	if err != nil {
		t.Fatalf("Failed to get wife: %v", err)
	}

	if wife == nil {
		t.Fatal("Expected wife to be found")
	}

	if wife.XrefID() != "@I1@" {
		t.Errorf("Expected wife @I1@, got %s", wife.XrefID())
	}
}

func TestFamilyQuery_Children(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	children, err := query.Family("@F1@").Children()
	if err != nil {
		t.Fatalf("Failed to get children: %v", err)
	}

	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}

	if len(children) > 0 && children[0].XrefID() != "@I2@" {
		t.Errorf("Expected child @I2@, got %s", children[0].XrefID())
	}
}

func TestFamilyQuery_Parents(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	parents, err := query.Family("@F1@").Parents()
	if err != nil {
		t.Fatalf("Failed to get parents: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(parents))
	}
}

func TestFamilyQuery_MarriageDate(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := gedcom.NewGedcomLine(1, "MARR", "", "")
	marrLine.AddChild(gedcom.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	famLine.AddChild(marrLine)
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	date, err := query.Family("@F1@").MarriageDate()
	if err != nil {
		t.Fatalf("Failed to get marriage date: %v", err)
	}

	if date == "" {
		t.Error("Expected marriage date to be found")
	}
}

func TestFamilyQuery_Events(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := gedcom.NewGedcomLine(1, "MARR", "", "")
	marrLine.AddChild(gedcom.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	famLine.AddChild(marrLine)
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	events, err := query.Family("@F1@").Events()
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) == 0 {
		t.Error("Expected at least one event")
	}
}

func TestFamilyQuery_Exists(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	if !query.Family("@F1@").Exists() {
		t.Error("Expected family to exist")
	}

	if query.Family("@F2@").Exists() {
		t.Error("Expected family @F2@ not to exist")
	}
}
