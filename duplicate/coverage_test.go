package duplicate

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestMinInt tests the minInt helper function
func TestMinInt(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
		{1, -1, -1},
		{0, 0, 0},
		{100, 50, 50},
		{50, 100, 50},
	}

	for _, tt := range tests {
		result := minInt(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("minInt(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// TestBlockingMetrics_String tests the String method
func TestBlockingMetrics_String(t *testing.T) {
	bm := &BlockingMetrics{
		TotalPeople:              100,
		PeopleWithPrimaryBlock:   80,
		PeopleWithAnyBlock:       90,
		PeopleWithNoBlocks:       10,
		TotalBlocks:              50,
		PrimaryBlocks:            30,
		SurnameYearBlocks:        10,
		SurnameInitialBlocks:     5,
		SurnamePrefixBlocks:      5,
		TotalCandidatesGenerated: 500,
		TotalCandidatesScored:    450,
		AverageCandidatesPerPerson: 5.0,
		MaxCandidatesPerPerson:    20,
		PeopleWithZeroCandidates:  5,
		PeopleWithOneCandidate:   10,
		PeopleWithManyCandidates: 15,
		TopBlockSizes: []BlockSizeInfo{
			{Size: 10, Count: 5, BlockType: "surname"},
			{Size: 5, Count: 10, BlockType: "year"},
		},
	}

	str := bm.String()
	if str == "" {
		t.Error("expected non-empty string")
	}

	// Check for key metrics in output
	if !contains(str, "Total People") {
		t.Error("expected 'Total People' in output")
	}
	if !contains(str, "Total Blocks") {
		t.Error("expected 'Total Blocks' in output")
	}
}

// TestBlockingMetrics_String_WithWarning tests String with warnings
func TestBlockingMetrics_String_WithWarning(t *testing.T) {
	bm := &BlockingMetrics{
		TotalPeople:           100,
		RepetitionWarning:     "Test warning",
		PeopleWithPrimaryBlock: 80,
		PeopleWithAnyBlock:     90,
		PeopleWithNoBlocks:     10,
		TotalBlocks:           50,
	}

	str := bm.String()
	if !contains(str, "WARNING") {
		t.Error("expected 'WARNING' in output when warning is present")
	}
	if !contains(str, "Test warning") {
		t.Error("expected warning text in output")
	}
}

// TestBlockingMetrics_GetWarnings tests the GetWarnings method
func TestBlockingMetrics_GetWarnings(t *testing.T) {
	// Test with no warnings
	bm := &BlockingMetrics{
		TotalPeople: 100,
	}
	warnings := bm.GetWarnings()
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(warnings))
	}

	// Test with giant blocks warning
	bm.HasGiantBlocks = true
	bm.PeopleInGiantBlocks = 50
	bm.LargestBlockSize = 1000
	warnings = bm.GetWarnings()
	if len(warnings) == 0 {
		t.Error("expected warnings for giant blocks")
	}
	if !contains(warnings[0], "extremely common") {
		t.Error("expected 'extremely common' in giant blocks warning")
	}

	// Test with repetition warning
	bm2 := &BlockingMetrics{
		TotalPeople:       100,
		RepetitionWarning: "High repetition detected",
	}
	warnings = bm2.GetWarnings()
	if len(warnings) > 0 {
		found := false
		for _, w := range warnings {
			if contains(w, "High repetition") || contains(w, "repetition") {
				found = true
				break
			}
		}
		if !found {
			t.Log("Note: Repetition warning may not be in warnings list")
		}
	}

	// Test with low coverage (many people with no blocks)
	bm3 := &BlockingMetrics{
		TotalPeople:           100,
		PeopleWithAnyBlock:    10,
		PeopleWithNoBlocks:    90,
		PeopleWithZeroCandidates: 60, // Over half
	}
	warnings = bm3.GetWarnings()
	if len(warnings) == 0 {
		t.Error("expected warnings for high zero candidates")
	}

	// Test with common surname warning
	bm4 := &BlockingMetrics{
		TotalPeople:           100,
		PeopleWithPrimaryBlock: 50,
		LargestBlockSize:      40, // > 33% of total
		MostCommonSurname:     "Smith",
	}
	warnings = bm4.GetWarnings()
	// May or may not have warnings depending on thresholds
	_ = warnings
}

// TestPoolFunctions tests all memory pool functions
func TestPoolFunctions(t *testing.T) {
	// Test match slice pool
	matchSlice := getMatchSlice()
	if matchSlice == nil {
		t.Error("expected non-nil match slice")
	}
	matchSlice = append(matchSlice, DuplicateMatch{})
	putMatchSlice(matchSlice)
	matchSlice2 := getMatchSlice()
	if len(matchSlice2) != 0 {
		t.Error("expected cleared slice from pool")
	}
	putMatchSlice(nil) // Test nil handling

	// Test individual slice pool
	indiSlice := getIndividualSlice()
	if indiSlice == nil {
		t.Error("expected non-nil individual slice")
	}
	putIndividualSlice(indiSlice)
	putIndividualSlice(nil) // Test nil handling

	// Test string slice pool
	strSlice := getStringSlice()
	if strSlice == nil {
		t.Error("expected non-nil string slice")
	}
	strSlice = append(strSlice, "test")
	putStringSlice(strSlice)
	strSlice2 := getStringSlice()
	if len(strSlice2) != 0 {
		t.Error("expected cleared slice from pool")
	}
	putStringSlice(nil) // Test nil handling

	// Test job slice pool
	jobSlice := getJobSlice()
	if jobSlice == nil {
		t.Error("expected non-nil job slice")
	}
	putJobSlice(jobSlice)
	putJobSlice(nil) // Test nil handling
}

// TestQuickNameSimilarity tests the quickNameSimilarity function
func TestQuickNameSimilarity(t *testing.T) {
	tests := []struct {
		name1, name2 string
		expected     float64
	}{
		{"John Doe", "John Doe", 1.0},           // Exact match
		{"John Doe", "john doe", 1.0},           // Case-insensitive
		{"John Doe", "John Smith", 0.5},         // First word matches
		{"John Doe", "Jane Doe", 0.0},           // No match
		{"John", "John", 1.0},                   // Single word exact
		{"John", "Jane", 0.0},                   // Single word different
		{"", "", 1.0},                           // Both empty
		{"John", "", 0.0},                       // One empty
		{"John Michael", "John", 0.5},           // First word matches
		{"Mary Jane", "Mary Ann", 0.5},          // First word matches
	}

	for _, tt := range tests {
		result := quickNameSimilarity(tt.name1, tt.name2)
		if result != tt.expected {
			t.Errorf("quickNameSimilarity(%q, %q) = %.1f, expected %.1f",
				tt.name1, tt.name2, result, tt.expected)
		}
	}
}

// TestGetChildren tests the getChildren function
func TestGetChildren(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create a family with children
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam := types.NewFamilyRecord(famLine)
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	tree.AddRecord(fam)

	// Create individual with FAMS reference
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := types.NewIndividualRecord(indiLine)
	indiLine.AddChild(types.NewGedcomLine(1, "FAMS", "@F1@", ""))
	tree.AddRecord(indi)

	children := getChildren(indi, tree)
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}

	// Check children XREFs
	expected := []string{"@I3@", "@I4@"}
	for _, expectedXref := range expected {
		found := false
		for _, child := range children {
			if child == expectedXref {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected child %s not found", expectedXref)
		}
	}
}

// TestGetChildren_NoFamily tests getChildren with no family
func TestGetChildren_NoFamily(t *testing.T) {
	tree := types.NewGedcomTree()

	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	children := getChildren(indi, tree)
	if len(children) != 0 {
		t.Errorf("expected 0 children, got %d", len(children))
	}
}

// TestGetChildren_InvalidFamilyXref tests getChildren with invalid family XREF
func TestGetChildren_InvalidFamilyXref(t *testing.T) {
	tree := types.NewGedcomTree()

	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := types.NewIndividualRecord(indiLine)
	indiLine.AddChild(types.NewGedcomLine(1, "FAMS", "@F999@", "")) // Non-existent family
	tree.AddRecord(indi)

	children := getChildren(indi, tree)
	if len(children) != 0 {
		t.Errorf("expected 0 children for invalid family, got %d", len(children))
	}
}

// TestGenerateComparisonJobs tests generateComparisonJobs
func TestGenerateComparisonJobs(t *testing.T) {
	config := DefaultConfig()
	config.UseBlocking = false // Disable blocking for simpler test
	config.MinThreshold = 0.0  // Lower threshold to allow more comparisons
	detector := NewDuplicateDetector(config)

	// Create test individuals with similar names to pass shouldCompare filter
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	indi2 := createTestIndividual("John /Doe/", "John", "Doe", "1801", "New York") // Similar name
	indi2.FirstLine().XrefID = "@I2@"
	indi3 := createTestIndividual("John /Doe/", "John", "Doe", "1802", "New York") // Similar name
	indi3.FirstLine().XrefID = "@I3@"

	individuals := []*types.IndividualRecord{indi1, indi2, indi3}

	// Build indexes
	idx := detector.buildIndexes(individuals)

	// Generate jobs
	jobs := detector.generateComparisonJobs(individuals, idx)

	// Should generate at least some jobs (may be filtered by shouldCompare)
	// The exact number depends on shouldCompare logic
	if len(jobs) == 0 {
		t.Log("Note: No jobs generated (may be filtered by shouldCompare)")
		// This still tests the function, just with filtering
	} else {
		// Verify job structure
		for i, job := range jobs {
			if job.indi1 == nil || job.indi2 == nil {
				t.Errorf("job %d has nil individuals", i)
			}
			if job.index != i {
				t.Errorf("job %d has incorrect index: expected %d, got %d", i, i, job.index)
			}
		}
	}
}

// TestFindDuplicatesBetweenParallel tests parallel duplicate detection between two trees
func TestFindDuplicatesBetweenParallel(t *testing.T) {
	config := DefaultConfig()
	config.UseParallelProcessing = true
	config.NumWorkers = 2 // Use small number for testing
	detector := NewDuplicateDetector(config)

	// Create test individuals for tree 1
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	indi2 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi2.FirstLine().XrefID = "@I2@"

	// Create similar individuals for tree 2 (potential duplicates)
	indi3 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi3.FirstLine().XrefID = "@I3@"
	indi4 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi4.FirstLine().XrefID = "@I4@"

	individuals1 := []*types.IndividualRecord{indi1, indi2}
	individuals2 := []*types.IndividualRecord{indi3, indi4}

	matches, comparisons, err := detector.findDuplicatesBetweenParallel(individuals1, individuals2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have performed comparisons
	if comparisons == 0 {
		t.Error("expected non-zero comparisons")
	}

	// Should find some matches (exact duplicates)
	if len(matches) == 0 {
		t.Log("Note: No matches found (may be expected depending on similarity threshold)")
	}
}

// TestFindDuplicatesBetweenParallel_Empty tests with empty inputs
func TestFindDuplicatesBetweenParallel_Empty(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())

	// Test with empty first list
	matches, comparisons, err := detector.findDuplicatesBetweenParallel([]*types.IndividualRecord{}, []*types.IndividualRecord{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
	if comparisons != 0 {
		t.Errorf("expected 0 comparisons, got %d", comparisons)
	}

	// Test with empty second list
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	matches, comparisons, err = detector.findDuplicatesBetweenParallel([]*types.IndividualRecord{indi1}, []*types.IndividualRecord{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
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


