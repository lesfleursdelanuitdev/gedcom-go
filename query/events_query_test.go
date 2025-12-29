package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestEventsQuery_GetEventsOnDate tests GetEventsOnDate with real GEDCOM files
func TestEventsQuery_GetEventsOnDate(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
		"tree1.ged",
	}

	if !testing.Short() {
		testFiles = append(testFiles, "royal92.ged")
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

			// Test with year only
			events, err := graph.GetEventsOnDate(1800, 0, 0)
			if err != nil {
				t.Errorf("GetEventsOnDate failed: %v", err)
			}
			_ = events // Just verify it doesn't panic

			// Test with year and month
			events, err = graph.GetEventsOnDate(1800, 1, 0)
			if err != nil {
				t.Errorf("GetEventsOnDate failed: %v", err)
			}
			_ = events

			// Test with full date
			events, err = graph.GetEventsOnDate(1800, 1, 1)
			if err != nil {
				t.Errorf("GetEventsOnDate failed: %v", err)
			}
			_ = events

			// Test with no matches (future date)
			events, err = graph.GetEventsOnDate(3000, 1, 1)
			if err != nil {
				t.Errorf("GetEventsOnDate failed: %v", err)
			}
			if len(events) > 0 {
				t.Logf("Found %d events in year 3000 (unexpected but not an error)", len(events))
			}
		})
	}
}

// TestEventsQuery_GetEventsOnDateByType tests GetEventsOnDateByType
func TestEventsQuery_GetEventsOnDateByType(t *testing.T) {
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

			// Test with BIRT event type
			events, err := graph.GetEventsOnDateByType("BIRT", 1800, 0, 0)
			if err != nil {
				t.Errorf("GetEventsOnDateByType failed: %v", err)
			}
			_ = events

			// Test with DEAT event type
			events, err = graph.GetEventsOnDateByType("DEAT", 1900, 0, 0)
			if err != nil {
				t.Errorf("GetEventsOnDateByType failed: %v", err)
			}
			_ = events

			// Test with MARR event type
			events, err = graph.GetEventsOnDateByType("MARR", 1850, 0, 0)
			if err != nil {
				t.Errorf("GetEventsOnDateByType failed: %v", err)
			}
			_ = events

			// Test with non-existent event type
			events, err = graph.GetEventsOnDateByType("NONEXISTENT", 1800, 0, 0)
			if err != nil {
				t.Errorf("GetEventsOnDateByType failed: %v", err)
			}
			if len(events) > 0 {
				t.Logf("Found %d events of type NONEXISTENT (unexpected)", len(events))
			}
		})
	}
}

// TestEventsQuery_GetRecordsForEvent tests GetRecordsForEvent
func TestEventsQuery_GetRecordsForEvent(t *testing.T) {
	// Create a test tree with events
	tree := types.NewGedcomTree()

	// Create an individual with a birth event
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	birtLine := types.NewGedcomLine(1, "BIRT", "", "")
	dateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1800", "")
	birtLine.AddChild(dateLine)
	indiLine.AddChild(birtLine)
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get all individuals to find event nodes
	allIndividuals := graph.GetAllIndividuals()
	if len(allIndividuals) == 0 {
		t.Fatal("No individuals found in graph")
	}

	// Get events for the first individual
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	events, err := qb.Individual("@I1@").GetEvents()
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) > 0 {
		// Try to find the event node
		eventID := events[0].EventID
		records, err := graph.GetRecordsForEvent(eventID)
		if err != nil {
			// Event might not be in graph as EventNode, test error case
			if err.Error() == "event "+eventID+" not found" {
				t.Logf("Event not found as EventNode (expected for record-based events)")
			} else {
				t.Errorf("Unexpected error: %v", err)
			}
		} else {
			if len(records) == 0 {
				t.Error("Expected at least one record for event")
			}
		}
	}

	// Test with non-existent event ID
	_, err = graph.GetRecordsForEvent("nonexistent_event_id")
	if err == nil {
		t.Error("Expected error for non-existent event ID")
	}
}

// TestEventsQuery_matchesDate tests the matchesDate function with various date formats
func TestEventsQuery_matchesDate(t *testing.T) {
	// Create a test tree to access matchesDate indirectly through GetEventsOnDate
	tree := types.NewGedcomTree()

	// Create individuals with various date formats
	testCases := []struct {
		name     string
		xref     string
		dateStr  string
		year     int
		month    int
		day      int
		shouldMatch bool
	}{
		{"Exact date match", "@I1@", "1 JAN 1800", 1800, 1, 1, true},
		{"Exact date - year only", "@I2@", "1800", 1800, 0, 0, true},
		{"Exact date - year and month", "@I3@", "JAN 1800", 1800, 1, 0, true},
		{"Exact date - no match year", "@I4@", "1 JAN 1800", 1801, 1, 1, false},
		{"Exact date - no match month", "@I5@", "1 JAN 1800", 1800, 2, 1, false},
		{"Exact date - no match day", "@I6@", "1 JAN 1800", 1800, 1, 2, false},
		{"Range date - within range", "@I7@", "BET 1 JAN 1800 AND 31 DEC 1800", 1800, 6, 15, true},
		{"Range date - before range", "@I8@", "BET 1 JAN 1800 AND 31 DEC 1800", 1799, 6, 15, false},
		{"Range date - after range", "@I9@", "BET 1 JAN 1800 AND 31 DEC 1800", 1801, 6, 15, false},
		{"ABOUT date", "@I10@", "ABT 1800", 1800, 0, 0, true},
		{"BEFORE date - matches parsed year", "@I11@", "BEF 1800", 1800, 0, 0, true}, // matchesDate checks if parsed year matches target
		{"AFTER date - matches parsed year", "@I12@", "AFT 1800", 1800, 0, 0, true}, // matchesDate checks if parsed year matches target
		{"BEFORE date - no match different year", "@I15@", "BEF 1800", 1799, 0, 0, false}, // parsed year is 1800, target is 1799
		{"AFTER date - no match different year", "@I16@", "AFT 1800", 1801, 0, 0, false},   // parsed year is 1800, target is 1801
		{"Empty date", "@I13@", "", 1800, 1, 1, false},
		{"Invalid date", "@I14@", "INVALID DATE", 1800, 1, 1, false},
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

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test each case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			events, err := graph.GetEventsOnDate(tc.year, tc.month, tc.day)
			if err != nil {
				t.Errorf("GetEventsOnDate failed: %v", err)
				return
			}

			// Check if we found an event for this individual
			found := false
			for _, event := range events {
				if event.Owner != nil && event.Owner.ID() == tc.xref {
					found = true
					break
				}
			}

			if found != tc.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v for %s with date %s", tc.shouldMatch, found, tc.xref, tc.dateStr)
			}
		})
	}
}

// TestEventsQuery_GetEventsOnDate_EdgeCases tests edge cases for GetEventsOnDate
func TestEventsQuery_GetEventsOnDate_EdgeCases(t *testing.T) {
	// Create empty graph
	tree := types.NewGedcomTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with empty graph
	events, err := graph.GetEventsOnDate(1800, 1, 1)
	if err != nil {
		t.Errorf("GetEventsOnDate failed on empty graph: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events in empty graph, got %d", len(events))
	}

	// Test with zero values (should match any date)
	events, err = graph.GetEventsOnDate(0, 0, 0)
	if err != nil {
		t.Errorf("GetEventsOnDate failed with zero values: %v", err)
	}
	_ = events
}

// TestEventsQuery_GetEventsOnDateByType_EdgeCases tests edge cases
func TestEventsQuery_GetEventsOnDateByType_EdgeCases(t *testing.T) {
	tree := types.NewGedcomTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with empty graph
	events, err := graph.GetEventsOnDateByType("BIRT", 1800, 1, 1)
	if err != nil {
		t.Errorf("GetEventsOnDateByType failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(events))
	}

	// Test with empty event type
	events, err = graph.GetEventsOnDateByType("", 1800, 0, 0)
	if err != nil {
		t.Errorf("GetEventsOnDateByType failed: %v", err)
	}
	_ = events
}

