package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestQueryBuilder_Individual(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	indiQuery := query.Individual("@I1@")
	if indiQuery == nil {
		t.Fatal("Expected IndividualQuery to be created")
	}

	if indiQuery.xrefID != "@I1@" {
		t.Errorf("Expected xrefID @I1@, got %s", indiQuery.xrefID)
	}
}

func TestIndividualQuery_Parents(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create parent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	parents, err := query.Individual("@I2@").Parents()
	if err != nil {
		t.Fatalf("Failed to get parents: %v", err)
	}

	if len(parents) == 0 {
		t.Error("Expected at least one parent")
	}

	found := false
	for _, parent := range parents {
		if parent.XrefID() == "@I1@" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected @I1@ to be a parent")
	}
}

func TestIndividualQuery_Ancestors(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create grandparent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create parent
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create child
	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family 1: I1 is parent of I2
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I3
	fam2Line := gedcom.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2 := gedcom.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	ancestors, err := query.Individual("@I3@").Ancestors().Execute()
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}

	if len(ancestors) < 2 {
		t.Errorf("Expected at least 2 ancestors, got %d", len(ancestors))
	}

	// Check that I1 and I2 are ancestors
	foundI1 := false
	foundI2 := false
	for _, ancestor := range ancestors {
		if ancestor.XrefID() == "@I1@" {
			foundI1 = true
		}
		if ancestor.XrefID() == "@I2@" {
			foundI2 = true
		}
	}

	if !foundI1 {
		t.Error("Expected @I1@ to be an ancestor")
	}
	if !foundI2 {
		t.Error("Expected @I2@ to be an ancestor")
	}
}

func TestAncestorQuery_MaxGenerations(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create 3-generation family
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family 1: I1 is parent of I2
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I3
	fam2Line := gedcom.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2 := gedcom.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Limit to 1 generation
	ancestors, err := query.Individual("@I3@").Ancestors().MaxGenerations(1).Execute()
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}

	// Should only get I2 (parent), not I1 (grandparent)
	if len(ancestors) != 1 {
		t.Errorf("Expected 1 ancestor with MaxGenerations(1), got %d", len(ancestors))
	}

	if len(ancestors) > 0 && ancestors[0].XrefID() != "@I2@" {
		t.Errorf("Expected @I2@, got %s", ancestors[0].XrefID())
	}
}

func TestIndividualQuery_Descendants(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create parent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create grandchild
	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family 1: I1 is parent of I2
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I3
	fam2Line := gedcom.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2 := gedcom.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	descendants, err := query.Individual("@I1@").Descendants().Execute()
	if err != nil {
		t.Fatalf("Failed to get descendants: %v", err)
	}

	if len(descendants) < 2 {
		t.Errorf("Expected at least 2 descendants, got %d", len(descendants))
	}
}

func TestIndividualQuery_RelationshipTo(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create parent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	result, err := query.Individual("@I1@").RelationshipTo("@I2@").Execute()
	if err != nil {
		t.Fatalf("Failed to calculate relationship: %v", err)
	}

	if !result.IsDirect {
		t.Error("Expected direct relationship")
	}

	if result.RelationshipType != "parent" {
		t.Errorf("Expected relationship type 'parent', got '%s'", result.RelationshipType)
	}
}

func TestIndividualQuery_PathTo(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create parent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	path, err := query.Individual("@I1@").PathTo("@I2@").Shortest()
	if err != nil {
		t.Fatalf("Failed to find path: %v", err)
	}

	if path == nil {
		t.Fatal("Expected path to be found")
	}

	if path.Length == 0 {
		t.Error("Expected path length > 0")
	}
}

func TestFilterQuery_ByName(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "Jane /Smith/", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().ByName("John").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_BySex(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "SEX", "M", ""))
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "SEX", "F", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().BySex("M").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestFilterQuery_HasChildren(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create parent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create individual without children
	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().HasChildren().Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_MultipleFilters(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "SEX", "M", ""))
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "SEX", "F", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Filter by name "John" AND sex "M"
	results, err := query.Filter().ByName("John").BySex("M").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Filter by name "John" AND sex "F" (should return 0)
	results2, err := query.Filter().ByName("John").BySex("F").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results2) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results2))
	}
}

func TestAncestorQuery_IncludeSelf(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	ancestors, err := query.Individual("@I1@").Ancestors().IncludeSelf().Execute()
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}

	found := false
	for _, ancestor := range ancestors {
		if ancestor.XrefID() == "@I1@" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected @I1@ to be included when IncludeSelf() is called")
	}
}

func TestPathQuery_Shortest(t *testing.T) {
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

	path, err := query.Individual("@I1@").PathTo("@I2@").Shortest()
	if err != nil {
		t.Fatalf("Failed to find shortest path: %v", err)
	}

	if path == nil {
		t.Fatal("Expected path to be found")
	}

	if path.Length == 0 {
		t.Error("Expected path length > 0")
	}
}

func TestPathQuery_All(t *testing.T) {
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

	paths, err := query.Individual("@I1@").PathTo("@I2@").All()
	if err != nil {
		t.Fatalf("Failed to find paths: %v", err)
	}

	if len(paths) == 0 {
		t.Error("Expected at least one path")
	}
}

func TestAncestorQuery_Count(t *testing.T) {
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

	count, err := query.Individual("@I2@").Ancestors().Count()
	if err != nil {
		t.Fatalf("Failed to count ancestors: %v", err)
	}

	if count == 0 {
		t.Error("Expected at least one ancestor")
	}
}

func TestFilterQuery_Count(t *testing.T) {
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

	count, err := query.Filter().ByName("John").Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	// Should be 0 since no individuals named "John"
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}
