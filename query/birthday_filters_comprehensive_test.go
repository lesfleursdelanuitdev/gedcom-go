package query

import (
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestBirthdayFilters_ByBirthMonth_Comprehensive tests ByBirthMonth with all date types
func TestBirthdayFilters_ByBirthMonth_Comprehensive(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create individuals with various date types
	testCases := []struct {
		xref        string
		dateStr     string
		month       int
		shouldMatch bool
		desc        string
	}{
		{"@I1@", "15 JAN 1800", 1, true, "Exact date - January"},
		{"@I2@", "15 FEB 1800", 2, true, "Exact date - February"},
		{"@I3@", "15 JAN 1800", 2, false, "Exact date - wrong month"},
		{"@I4@", "BET 1 JAN 1800 AND 31 JAN 1800", 1, true, "Range date - within range"},
		{"@I5@", "BET 1 DEC 1800 AND 31 JAN 1801", 1, true, "Range date - year wraparound"},
		{"@I6@", "BET 1 DEC 1800 AND 31 JAN 1801", 12, true, "Range date - year wraparound December"},
		{"@I7@", "BET 1 FEB 1800 AND 28 FEB 1800", 1, false, "Range date - outside range"},
		{"@I8@", "ABT JAN 1800", 1, true, "ABOUT date - with month"},
		{"@I9@", "ABT 1800", 1, false, "ABOUT date - no month"},
		{"@I10@", "BEF JAN 1800", 1, true, "BEFORE date - with month"},
		{"@I11@", "AFT JAN 1800", 1, true, "AFTER date - with month"},
		{"@I12@", "", 1, false, "No date"},
	}

	for _, tc := range testCases {
		indiLine := types.NewGedcomLine(0, "INDI", "", tc.xref)
		nameLine := types.NewGedcomLine(1, "NAME", "Test /Person/", "")
		indiLine.AddChild(nameLine)
		if tc.dateStr != "" {
			birtLine := types.NewGedcomLine(1, "BIRT", "", "")
			dateLine := types.NewGedcomLine(2, "DATE", tc.dateStr, "")
			birtLine.AddChild(dateLine)
			indiLine.AddChild(birtLine)
		}
		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fq := NewFilterQuery(graph)
			results, err := fq.ByBirthMonth(tc.month).Execute()
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			found := false
			for _, result := range results {
				if result.XrefID() == tc.xref {
					found = true
					break
				}
			}

			if found != tc.shouldMatch {
				t.Errorf("Expected match=%v for %s (date: %s, month: %d), got match=%v",
					tc.shouldMatch, tc.xref, tc.dateStr, tc.month, found)
			}
		})
	}
}

// TestBirthdayFilters_ByBirthDay_Comprehensive tests ByBirthDay with all date types
func TestBirthdayFilters_ByBirthDay_Comprehensive(t *testing.T) {
	tree := types.NewGedcomTree()

	testCases := []struct {
		xref        string
		dateStr     string
		day         int
		shouldMatch bool
		desc        string
	}{
		{"@I1@", "15 JAN 1800", 15, true, "Exact date - day 15"},
		{"@I2@", "1 JAN 1800", 1, true, "Exact date - day 1"},
		{"@I3@", "31 JAN 1800", 31, true, "Exact date - day 31"},
		{"@I4@", "15 JAN 1800", 16, false, "Exact date - wrong day"},
		{"@I5@", "BET 10 JAN 1800 AND 20 JAN 1800", 15, true, "Range date - within range"},
		{"@I6@", "BET 10 JAN 1800 AND 20 JAN 1800", 5, false, "Range date - before range"},
		{"@I7@", "BET 10 JAN 1800 AND 20 JAN 1800", 25, false, "Range date - after range"},
		{"@I8@", "BET 25 JAN 1800 AND 5 FEB 1800", 1, true, "Range date - month wraparound"},
		{"@I9@", "BET 25 JAN 1800 AND 5 FEB 1800", 30, true, "Range date - month wraparound end"},
		{"@I10@", "ABT 15 JAN 1800", 15, true, "ABOUT date - with day"},
		{"@I11@", "ABT JAN 1800", 15, false, "ABOUT date - no day"},
		{"@I12@", "BEF 15 JAN 1800", 15, true, "BEFORE date - with day"},
		{"@I13@", "AFT 15 JAN 1800", 15, true, "AFTER date - with day"},
		{"@I14@", "", 15, false, "No date"},
	}

	for _, tc := range testCases {
		indiLine := types.NewGedcomLine(0, "INDI", "", tc.xref)
		nameLine := types.NewGedcomLine(1, "NAME", "Test /Person/", "")
		indiLine.AddChild(nameLine)
		if tc.dateStr != "" {
			birtLine := types.NewGedcomLine(1, "BIRT", "", "")
			dateLine := types.NewGedcomLine(2, "DATE", tc.dateStr, "")
			birtLine.AddChild(dateLine)
			indiLine.AddChild(birtLine)
		}
		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fq := NewFilterQuery(graph)
			results, err := fq.ByBirthDay(tc.day).Execute()
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			found := false
			for _, result := range results {
				if result.XrefID() == tc.xref {
					found = true
					break
				}
			}

			if found != tc.shouldMatch {
				t.Errorf("Expected match=%v for %s (date: %s, day: %d), got match=%v",
					tc.shouldMatch, tc.xref, tc.dateStr, tc.day, found)
			}
		})
	}
}

// TestBirthdayFilters_ByBirthMonthAndDay_Comprehensive tests ByBirthMonthAndDay with all date types
func TestBirthdayFilters_ByBirthMonthAndDay_Comprehensive(t *testing.T) {
	tree := types.NewGedcomTree()

	testCases := []struct {
		xref        string
		dateStr     string
		month       int
		day         int
		shouldMatch bool
		desc        string
	}{
		{"@I1@", "15 JAN 1800", 1, 15, true, "Exact date - match"},
		{"@I2@", "15 JAN 1800", 1, 16, false, "Exact date - wrong day"},
		{"@I3@", "15 JAN 1800", 2, 15, false, "Exact date - wrong month"},
		{"@I4@", "BET 10 JAN 1800 AND 20 JAN 1800", 1, 15, true, "Range date - within range"},
		{"@I5@", "BET 10 JAN 1800 AND 20 JAN 1800", 1, 5, false, "Range date - before range"},
		{"@I6@", "BET 10 JAN 1800 AND 20 JAN 1800", 1, 25, false, "Range date - after range"},
		{"@I7@", "BET 25 JAN 1800 AND 5 FEB 1800", 1, 30, true, "Range date - month wraparound"},
		{"@I8@", "BET 25 JAN 1800 AND 5 FEB 1800", 2, 3, true, "Range date - month wraparound February"},
		{"@I9@", "BET 1 JAN 1800 AND 31 DEC 1800", 6, 15, true, "Range date - full year"},
		{"@I10@", "ABT 15 JAN 1800", 1, 15, true, "ABOUT date - match"},
		{"@I11@", "ABT JAN 1800", 1, 15, false, "ABOUT date - no day"},
		{"@I12@", "BEF 15 JAN 1800", 1, 15, true, "BEFORE date - match"},
		{"@I13@", "AFT 15 JAN 1800", 1, 15, true, "AFTER date - match"},
		{"@I14@", "", 1, 15, false, "No date"},
	}

	for _, tc := range testCases {
		indiLine := types.NewGedcomLine(0, "INDI", "", tc.xref)
		nameLine := types.NewGedcomLine(1, "NAME", "Test /Person/", "")
		indiLine.AddChild(nameLine)
		if tc.dateStr != "" {
			birtLine := types.NewGedcomLine(1, "BIRT", "", "")
			dateLine := types.NewGedcomLine(2, "DATE", tc.dateStr, "")
			birtLine.AddChild(dateLine)
			indiLine.AddChild(birtLine)
		}
		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fq := NewFilterQuery(graph)
			results, err := fq.ByBirthMonthAndDay(tc.month, tc.day).Execute()
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			found := false
			for _, result := range results {
				if result.XrefID() == tc.xref {
					found = true
					break
				}
			}

			if found != tc.shouldMatch {
				t.Errorf("Expected match=%v for %s (date: %s, month: %d, day: %d), got match=%v",
					tc.shouldMatch, tc.xref, tc.dateStr, tc.month, tc.day, found)
			}
		})
	}
}

// TestBirthdayFilters_EdgeCases tests edge cases for invalid inputs
func TestBirthdayFilters_EdgeCases(t *testing.T) {
	tree := types.NewGedcomTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test invalid month values
	fq := NewFilterQuery(graph)
	fq2 := fq.ByBirthMonth(0) // Invalid
	if fq2 != fq {
		t.Error("Expected unchanged query for month 0")
	}

	fq3 := fq.ByBirthMonth(13) // Invalid
	if fq3 != fq {
		t.Error("Expected unchanged query for month 13")
	}

	fq4 := fq.ByBirthMonth(-1) // Invalid
	if fq4 != fq {
		t.Error("Expected unchanged query for month -1")
	}

	// Test invalid day values
	fq5 := fq.ByBirthDay(0) // Invalid
	if fq5 != fq {
		t.Error("Expected unchanged query for day 0")
	}

	fq6 := fq.ByBirthDay(32) // Invalid
	if fq6 != fq {
		t.Error("Expected unchanged query for day 32")
	}

	fq7 := fq.ByBirthDay(-1) // Invalid
	if fq7 != fq {
		t.Error("Expected unchanged query for day -1")
	}

	// Test invalid month and day combinations
	fq8 := fq.ByBirthMonthAndDay(0, 15) // Invalid month
	if fq8 != fq {
		t.Error("Expected unchanged query for month 0")
	}

	fq9 := fq.ByBirthMonthAndDay(1, 0) // Invalid day
	if fq9 != fq {
		t.Error("Expected unchanged query for day 0")
	}

	fq10 := fq.ByBirthMonthAndDay(13, 15) // Invalid month
	if fq10 != fq {
		t.Error("Expected unchanged query for month 13")
	}

	fq11 := fq.ByBirthMonthAndDay(1, 32) // Invalid day
	if fq11 != fq {
		t.Error("Expected unchanged query for day 32")
	}
}

// TestBirthdayFilters_RealData tests birthday filters with real GEDCOM files
func TestBirthdayFilters_RealData(t *testing.T) {
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

			// Test ByBirthMonth with real data
			fq := NewFilterQuery(graph)
			results, err := fq.ByBirthMonth(1).Execute() // January
			if err != nil {
				t.Errorf("ByBirthMonth failed: %v", err)
			}
			_ = results // Just verify it doesn't panic

			// Test ByBirthDay with real data
			fq2 := NewFilterQuery(graph)
			results2, err2 := fq2.ByBirthDay(15).Execute() // Day 15
			if err2 != nil {
				t.Errorf("ByBirthDay failed: %v", err2)
			}
			_ = results2

			// Test ByBirthMonthAndDay with real data
			fq3 := NewFilterQuery(graph)
			results3, err3 := fq3.ByBirthMonthAndDay(1, 1).Execute() // January 1
			if err3 != nil {
				t.Errorf("ByBirthMonthAndDay failed: %v", err3)
			}
			_ = results3
		})
	}
}

// TestBirthdayFilters_ByBirthDateRange_Comprehensive tests ByBirthDateRange
func TestBirthdayFilters_ByBirthDateRange_Comprehensive(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create individuals with various birth dates
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1800", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Doe/", "")
	indi2Line.AddChild(name2Line)
	birt2Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt2Line.AddChild(types.NewGedcomLine(2, "DATE", "20 FEB 1850", ""))
	indi2Line.AddChild(birt2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test date range that includes both
	start := time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(1900, 12, 31, 23, 59, 59, 999999999, time.UTC)

	fq := NewFilterQuery(graph)
	results, err := fq.ByBirthDateRange(start, end).Execute()
	if err != nil {
		t.Fatalf("ByBirthDateRange failed: %v", err)
	}

	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}
}
