package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestAncestorQuery_Count_Additional tests Count method with additional edge cases
func TestAncestorQuery_Count_Additional(t *testing.T) {
	// Create a simple tree
	tree := types.NewGedcomTree()

	// Individual with no parents
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "Root /Person/", ""))
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Individual with parents
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Child /Person/", ""))
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Parents
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(types.NewGedcomLine(1, "NAME", "Parent1 /Person/", ""))
	tree.AddRecord(types.NewIndividualRecord(indi3Line))

	indi4Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	indi4Line.AddChild(types.NewGedcomLine(1, "NAME", "Parent2 /Person/", ""))
	tree.AddRecord(types.NewIndividualRecord(indi4Line))

	// Family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I3@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I4@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(famLine))

	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test Count with individual who has no ancestors
	count, err := qb.Individual("@I1@").Ancestors().Count()
	if err != nil {
		t.Errorf("Count should not return error: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 ancestors for root person, got %d", count)
	}

	// Test Count with individual who has ancestors
	count2, err2 := qb.Individual("@I2@").Ancestors().Count()
	if err2 != nil {
		t.Errorf("Count should not return error: %v", err2)
	}
	if count2 != 2 {
		t.Errorf("Expected 2 ancestors, got %d", count2)
	}

	// Test Count with invalid XREF
	count3, err3 := qb.Individual("@INVALID@").Ancestors().Count()
	if err3 != nil {
		t.Errorf("Count should not return error for invalid XREF: %v", err3)
	}
	if count3 != 0 {
		t.Errorf("Expected 0 count for invalid XREF, got %d", count3)
	}
}

// TestDescendantQuery_Count_Additional tests Count method with additional edge cases
func TestDescendantQuery_Count_Additional(t *testing.T) {
	// Create a simple tree
	tree := types.NewGedcomTree()

	// Root individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "Root /Person/", ""))
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Child
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Child /Person/", ""))
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(famLine))

	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test Count with individual who has no descendants
	count, err := qb.Individual("@I2@").Descendants().Count()
	if err != nil {
		t.Errorf("Count should not return error: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 descendants, got %d", count)
	}

	// Test Count with individual who has descendants
	count2, err2 := qb.Individual("@I1@").Descendants().Count()
	if err2 != nil {
		t.Errorf("Count should not return error: %v", err2)
	}
	if count2 != 1 {
		t.Errorf("Expected 1 descendant, got %d", count2)
	}

	// Test Count with invalid XREF
	count3, err3 := qb.Individual("@INVALID@").Descendants().Count()
	if err3 != nil {
		t.Errorf("Count should not return error for invalid XREF: %v", err3)
	}
	if count3 != 0 {
		t.Errorf("Expected 0 count for invalid XREF, got %d", count3)
	}
}

// TestAncestorQuery_Exists_Additional tests Exists method with additional edge cases
func TestAncestorQuery_Exists_Additional(t *testing.T) {
	tree := types.NewGedcomTree()

	// Individual with no parents
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "Root /Person/", ""))
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Individual with parents
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Child /Person/", ""))
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Parent
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(types.NewGedcomLine(1, "NAME", "Parent /Person/", ""))
	tree.AddRecord(types.NewIndividualRecord(indi3Line))

	// Family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I3@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(famLine))

	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test Exists with individual who has no ancestors
	exists, err := qb.Individual("@I1@").Ancestors().Exists()
	if err != nil {
		t.Errorf("Exists should not return error: %v", err)
	}
	if exists {
		t.Error("Expected false for root person with no ancestors")
	}

	// Test Exists with individual who has ancestors
	exists2, err2 := qb.Individual("@I2@").Ancestors().Exists()
	if err2 != nil {
		t.Errorf("Exists should not return error: %v", err2)
	}
	if !exists2 {
		t.Error("Expected true for person with ancestors")
	}

	// Test Exists with invalid XREF
	exists3, err3 := qb.Individual("@INVALID@").Ancestors().Exists()
	if err3 != nil {
		t.Errorf("Exists should not return error for invalid XREF: %v", err3)
	}
	if exists3 {
		t.Error("Expected false for invalid XREF")
	}
}

// TestDescendantQuery_Exists_Additional tests Exists method with additional edge cases
func TestDescendantQuery_Exists_Additional(t *testing.T) {
	tree := types.NewGedcomTree()

	// Root individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "Root /Person/", ""))
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Child
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Child /Person/", ""))
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(famLine))

	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test Exists with individual who has no descendants
	exists, err := qb.Individual("@I2@").Descendants().Exists()
	if err != nil {
		t.Errorf("Exists should not return error: %v", err)
	}
	if exists {
		t.Error("Expected false for person with no descendants")
	}

	// Test Exists with individual who has descendants
	exists2, err2 := qb.Individual("@I1@").Descendants().Exists()
	if err2 != nil {
		t.Errorf("Exists should not return error: %v", err2)
	}
	if !exists2 {
		t.Error("Expected true for person with descendants")
	}

	// Test Exists with invalid XREF
	exists3, err3 := qb.Individual("@INVALID@").Descendants().Exists()
	if err3 != nil {
		t.Errorf("Exists should not return error for invalid XREF: %v", err3)
	}
	if exists3 {
		t.Error("Expected false for invalid XREF")
	}
}

// TestCount_RealData tests Count methods with real GEDCOM files
func TestCount_RealData(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}

	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse the file
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			// Build graph
			_, err = BuildGraph(tree)
			if err != nil {
				t.Fatalf("Failed to build graph for %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder: %v", err)
			}

			// Get first individual
			allIndividuals := qb.AllIndividuals()
			results, err := allIndividuals.Execute()
			if err != nil || len(results) == 0 {
				t.Skipf("No individuals found in %s", filename)
			}

			firstXref := results[0].XrefID()

			// Test Ancestors().Count()
			count, err := qb.Individual(firstXref).Ancestors().Count()
			if err != nil {
				t.Errorf("Ancestors().Count() failed: %v", err)
			}
			_ = count // Just verify it doesn't panic

			// Test Descendants().Count()
			count2, err2 := qb.Individual(firstXref).Descendants().Count()
			if err2 != nil {
				t.Errorf("Descendants().Count() failed: %v", err2)
			}
			_ = count2
		})
	}
}

// TestExists_RealData tests Exists methods with real GEDCOM files
func TestExists_RealData(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}

	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse the file
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			// Build graph
			_, err = BuildGraph(tree)
			if err != nil {
				t.Fatalf("Failed to build graph for %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder: %v", err)
			}

			// Get first individual
			allIndividuals := qb.AllIndividuals()
			results, err := allIndividuals.Execute()
			if err != nil || len(results) == 0 {
				t.Skipf("No individuals found in %s", filename)
			}

			firstXref := results[0].XrefID()

			// Test Ancestors().Exists()
			exists, err := qb.Individual(firstXref).Ancestors().Exists()
			if err != nil {
				t.Errorf("Ancestors().Exists() failed: %v", err)
			}
			_ = exists // Just verify it doesn't panic

			// Test Descendants().Exists()
			exists2, err2 := qb.Individual(firstXref).Descendants().Exists()
			if err2 != nil {
				t.Errorf("Descendants().Exists() failed: %v", err2)
			}
			_ = exists2
		})
	}
}

