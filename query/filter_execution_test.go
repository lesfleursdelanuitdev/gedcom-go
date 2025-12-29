package query

import (
	"errors"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestFilterByBool tests the filterByBool function directly
func TestFilterByBool(t *testing.T) {
	// Create a simple check function that returns true for even IDs
	checkEven := func(id uint32) (bool, error) {
		return id%2 == 0, nil
	}

	// Create a check function that returns errors for certain IDs
	checkWithError := func(id uint32) (bool, error) {
		if id == 5 {
			return false, errors.New("test error")
		}
		return id > 3, nil
	}

	testCases := []struct {
		name     string
		ids      []uint32
		checkFunc func(uint32) (bool, error)
		want     bool
		expected []uint32
		desc     string
	}{
		{
			name:     "Filter even IDs (want true)",
			ids:      []uint32{1, 2, 3, 4, 5, 6},
			checkFunc: checkEven,
			want:     true,
			expected: []uint32{2, 4, 6},
			desc:     "Should return even IDs when want=true",
		},
		{
			name:     "Filter odd IDs (want false)",
			ids:      []uint32{1, 2, 3, 4, 5, 6},
			checkFunc: checkEven,
			want:     false,
			expected: []uint32{1, 3, 5},
			desc:     "Should return odd IDs when want=false",
		},
		{
			name:     "Filter with error handling",
			ids:      []uint32{1, 2, 3, 4, 5, 6},
			checkFunc: checkWithError,
			want:     true,
			expected: []uint32{4, 6}, // ID 5 has error, so it's excluded
			desc:     "Should exclude IDs that return errors",
		},
		{
			name:     "Empty input",
			ids:      []uint32{},
			checkFunc: checkEven,
			want:     true,
			expected: []uint32{},
			desc:     "Should return empty slice for empty input",
		},
		{
			name:     "No matches",
			ids:      []uint32{1, 3, 5},
			checkFunc: checkEven,
			want:     true,
			expected: []uint32{},
			desc:     "Should return empty slice when no matches",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterByBool(tc.ids, tc.checkFunc, tc.want)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d results, got %d", len(tc.expected), len(result))
				return
			}

			// Check that all expected IDs are in the result
			resultMap := make(map[uint32]bool)
			for _, id := range result {
				resultMap[id] = true
			}

			for _, expectedID := range tc.expected {
				if !resultMap[expectedID] {
					t.Errorf("Expected ID %d in result, but it's missing", expectedID)
				}
			}
		})
	}
}

// TestFilterByBool_RealData tests filterByBool indirectly through HasChildren, HasSpouse filters
func TestFilterByBool_RealData(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}

	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
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
			graph, err := BuildGraph(tree)
			if err != nil {
				t.Fatalf("Failed to build graph for %s: %v", filename, err)
			}

			// Test HasChildren filter (uses filterByBool internally in hybrid mode)
			fq := NewFilterQuery(graph)
			results, err := fq.HasChildren().Execute()
			if err != nil {
				t.Errorf("HasChildren filter failed: %v", err)
			}
			_ = results // Just verify it doesn't panic

			// Test NoChildren filter
			fq2 := NewFilterQuery(graph)
			results2, err2 := fq2.NoChildren().Execute()
			if err2 != nil {
				t.Errorf("NoChildren filter failed: %v", err2)
			}
			_ = results2

			// Test HasSpouse filter
			fq3 := NewFilterQuery(graph)
			results3, err3 := fq3.HasSpouse().Execute()
			if err3 != nil {
				t.Errorf("HasSpouse filter failed: %v", err3)
			}
			_ = results3

			// Test NoSpouse filter
			fq4 := NewFilterQuery(graph)
			results4, err4 := fq4.NoSpouse().Execute()
			if err4 != nil {
				t.Errorf("NoSpouse filter failed: %v", err4)
			}
			_ = results4

			// Test Living filter
			fq5 := NewFilterQuery(graph)
			results5, err5 := fq5.Living().Execute()
			if err5 != nil {
				t.Errorf("Living filter failed: %v", err5)
			}
			_ = results5
		})
	}
}

// TestFilterByBool_SyntheticData tests filterByBool with synthetic data
func TestFilterByBool_SyntheticData(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create individuals with and without children
	// Individual with children
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /One/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Individual without children
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Childless /Person/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Create a family with children
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	fam1Line.AddChild(husbLine)
	wifeLine := types.NewGedcomLine(1, "WIFE", "@I3@", "")
	fam1Line.AddChild(wifeLine)
	childLine := types.NewGedcomLine(1, "CHIL", "@I4@", "")
	fam1Line.AddChild(childLine)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Add wife and child
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Parent /Two/", "")
	indi3Line.AddChild(name3Line)
	tree.AddRecord(types.NewIndividualRecord(indi3Line))

	indi4Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Child /One/", "")
	indi4Line.AddChild(name4Line)
	tree.AddRecord(types.NewIndividualRecord(indi4Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test HasChildren filter
	fq := NewFilterQuery(graph)
	results, err := fq.HasChildren().Execute()
	if err != nil {
		t.Fatalf("HasChildren filter failed: %v", err)
	}

	// @I1@ and @I3@ should have children
	foundI1 := false
	foundI3 := false
	for _, result := range results {
		if result.XrefID() == "@I1@" {
			foundI1 = true
		}
		if result.XrefID() == "@I3@" {
			foundI3 = true
		}
	}

	if !foundI1 {
		t.Error("Expected @I1@ to have children")
	}
	if !foundI3 {
		t.Error("Expected @I3@ to have children")
	}

	// Test NoChildren filter
	fq2 := NewFilterQuery(graph)
	results2, err2 := fq2.NoChildren().Execute()
	if err2 != nil {
		t.Fatalf("NoChildren filter failed: %v", err2)
	}

	// @I2@ and @I4@ should not have children (I4 is a child, not a parent)
	foundI2 := false
	foundI4 := false
	for _, result := range results2 {
		if result.XrefID() == "@I2@" {
			foundI2 = true
		}
		if result.XrefID() == "@I4@" {
			foundI4 = true
		}
	}

	if !foundI2 {
		t.Error("Expected @I2@ to have no children")
	}
	if !foundI4 {
		t.Error("Expected @I4@ to have no children")
	}

	// Test HasSpouse filter
	fq3 := NewFilterQuery(graph)
	results3, err3 := fq3.HasSpouse().Execute()
	if err3 != nil {
		t.Fatalf("HasSpouse filter failed: %v", err3)
	}

	// @I1@ and @I3@ should have spouses
	foundI1Spouse := false
	foundI3Spouse := false
	for _, result := range results3 {
		if result.XrefID() == "@I1@" {
			foundI1Spouse = true
		}
		if result.XrefID() == "@I3@" {
			foundI3Spouse = true
		}
	}

	if !foundI1Spouse {
		t.Error("Expected @I1@ to have a spouse")
	}
	if !foundI3Spouse {
		t.Error("Expected @I3@ to have a spouse")
	}

	// Test NoSpouse filter
	fq4 := NewFilterQuery(graph)
	results4, err4 := fq4.NoSpouse().Execute()
	if err4 != nil {
		t.Fatalf("NoSpouse filter failed: %v", err4)
	}

	// @I2@ and @I4@ should not have spouses
	foundI2NoSpouse := false
	foundI4NoSpouse := false
	for _, result := range results4 {
		if result.XrefID() == "@I2@" {
			foundI2NoSpouse = true
		}
		if result.XrefID() == "@I4@" {
			foundI4NoSpouse = true
		}
	}

	if !foundI2NoSpouse {
		t.Error("Expected @I2@ to have no spouse")
	}
	if !foundI4NoSpouse {
		t.Error("Expected @I4@ to have no spouse")
	}
}

// TestFilterByBool_ErrorHandling tests error handling in filterByBool
func TestFilterByBool_ErrorHandling(t *testing.T) {
	// Create a check function that always returns an error
	alwaysError := func(id uint32) (bool, error) {
		return false, errors.New("always error")
	}

	// Create a check function that returns error for specific IDs
	selectiveError := func(id uint32) (bool, error) {
		if id == 2 || id == 4 {
			return false, errors.New("error for this ID")
		}
		return id%2 == 0, nil
	}

	testCases := []struct {
		name     string
		ids      []uint32
		checkFunc func(uint32) (bool, error)
		want     bool
		expected []uint32
		desc     string
	}{
		{
			name:     "All IDs return error",
			ids:      []uint32{1, 2, 3},
			checkFunc: alwaysError,
			want:     true,
			expected: []uint32{},
			desc:     "Should return empty slice when all IDs return errors",
		},
		{
			name:     "Some IDs return error",
			ids:      []uint32{1, 2, 3, 4, 5, 6},
			checkFunc: selectiveError,
			want:     true,
			expected: []uint32{6}, // Only 6 is even and doesn't error (2 and 4 error)
			desc:     "Should exclude IDs that return errors",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterByBool(tc.ids, tc.checkFunc, tc.want)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d results, got %d", len(tc.expected), len(result))
				return
			}

			// Check that all expected IDs are in the result
			resultMap := make(map[uint32]bool)
			for _, id := range result {
				resultMap[id] = true
			}

			for _, expectedID := range tc.expected {
				if !resultMap[expectedID] {
					t.Errorf("Expected ID %d in result, but it's missing", expectedID)
				}
			}

			// Check that no unexpected IDs are in the result
			for _, resultID := range result {
				found := false
				for _, expectedID := range tc.expected {
					if resultID == expectedID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected ID %d in result", resultID)
				}
			}
		})
	}
}

