package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// findTestDataFile attempts to locate a test data file given its name.
// It checks multiple common paths relative to the current working directory
// and the Go module root.
func findTestDataFile(filename string) string {
	// Possible paths to check
	possiblePaths := []string{
		filepath.Join("testdata", filename),
		filepath.Join("../testdata", filename),
		filepath.Join("../../testdata", filename),
		filepath.Join("../../../testdata", filename),
		filepath.Join("/apps/gedcom-go/testdata", filename), // Absolute path for specific environments
	}

	// Get current file's directory (for relative paths)
	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		currentDir := filepath.Dir(currentFile)
		possiblePaths = append(possiblePaths,
			filepath.Join(currentDir, "testdata", filename),
			filepath.Join(currentDir, "..", "testdata", filename),
			filepath.Join(currentDir, "..", "..", "testdata", filename),
			filepath.Join(currentDir, "..", "..", "..", "testdata", filename),
		)
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// getRandomIndividuals returns a few random individuals from the tree for testing
func getRandomIndividuals(tree *types.GedcomTree, count int) []string {
	var xrefs []string
	individuals := tree.GetAllIndividuals()
	for xref := range individuals {
		xrefs = append(xrefs, xref)
		if len(xrefs) >= count {
			break
		}
	}
	return xrefs
}

// TestAdvancedQueries_AllTestDataFiles tests all advanced queries on all test data files
func TestAdvancedQueries_AllTestDataFiles(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}

	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged", "pres2020.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
				return
			}

			// Parse the file
			hp := parser.NewHierarchicalParser()
			tree, err := hp.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			// Build query graph
			qb, err := query.NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to build query graph for %s: %v", filename, err)
			}

			// Get some test individuals
			testIndividuals := getRandomIndividuals(tree, 5)
			if len(testIndividuals) < 2 {
				t.Skipf("Not enough individuals in %s for testing", filename)
				return
			}

			fmt.Printf("\n=== Advanced Queries Test - %s ===\n", filename)
			fmt.Printf("Total individuals: %d\n", len(tree.GetAllIndividuals()))
			fmt.Printf("Test individuals: %v\n\n", testIndividuals)

			// Run all query tests
			testRelationshipDetection(t, qb, testIndividuals, filename)
			testOldestAncestor(t, qb, testIndividuals, filename)
			testCommonAncestors(t, qb, testIndividuals, filename)
			testLowestCommonAncestor(t, qb, testIndividuals, filename)
			testPathFinding(t, qb, testIndividuals, filename)
			testSpecificRelationships(t, qb, testIndividuals, filename)
			testBrickWalls(t, qb, filename)
			testEndOfLine(t, qb, filename)
			testMultipleSpouses(t, qb, filename)
			testMissingData(t, qb, filename)
			testGeographicQueries(t, qb, filename)
			testTemporalQueries(t, qb, filename)
			testNameBasedQueries(t, qb, filename)
			testGraphMetrics(t, qb, filename)
		})
	}
}

// testRelationshipDetection tests relationship detection between two individuals
func testRelationshipDetection(t *testing.T, qb *query.QueryBuilder, testIndividuals []string, filename string) {
	fmt.Printf("1. Relationship Detection:\n")

	if len(testIndividuals) < 2 {
		t.Skip("Not enough individuals for relationship detection test")
		return
	}

	fromXref := testIndividuals[0]
	toXref := testIndividuals[1]

	start := time.Now()
	result, err := qb.Individual(fromXref).RelationshipTo(toXref).Execute()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	if result == nil {
		fmt.Printf("  ⚠️  No relationship found between %s and %s\n", fromXref, toXref)
		return
	}

	fmt.Printf("  ✅ Relationship: %s\n", result.RelationshipType)
	fmt.Printf("     Degree: %d, Removal: %d\n", result.Degree, result.Removal)
	fmt.Printf("     Direct: %v, Ancestral: %v, Descendant: %v, Collateral: %v\n",
		result.IsDirect, result.IsAncestral, result.IsDescendant, result.IsCollateral)
	if result.Path != nil {
		fmt.Printf("     Path length: %d\n", result.Path.Length)
	}
	fmt.Printf("     Duration: %v\n", duration)

	// Test a few more pairs
	for i := 2; i < len(testIndividuals) && i < 4; i++ {
		toXref2 := testIndividuals[i]
		result2, err2 := qb.Individual(fromXref).RelationshipTo(toXref2).Execute()
		if err2 == nil && result2 != nil {
			fmt.Printf("  %s -> %s: %s (degree: %d)\n", fromXref, toXref2, result2.RelationshipType, result2.Degree)
		}
	}
	fmt.Printf("\n")
}

// testOldestAncestor tests finding the oldest/most distant ancestor
func testOldestAncestor(t *testing.T, qb *query.QueryBuilder, testIndividuals []string, filename string) {
	fmt.Printf("2. Oldest Ancestor:\n")

	if len(testIndividuals) == 0 {
		return
	}

	testXref := testIndividuals[0]

	start := time.Now()
	ancestors, err := qb.Individual(testXref).Ancestors().ExecuteWithPaths()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	if len(ancestors) == 0 {
		fmt.Printf("  ⚠️  No ancestors found for %s\n", testXref)
		return
	}

	// Find oldest by depth
	var oldestByDepth *query.AncestorPath
	maxDepth := -1
	for _, ancestor := range ancestors {
		if ancestor.Depth > maxDepth {
			maxDepth = ancestor.Depth
			oldestByDepth = ancestor
		}
	}

	if oldestByDepth != nil {
		fmt.Printf("  ✅ Oldest ancestor by depth: %s (depth: %d)\n",
			oldestByDepth.Ancestor.XrefID(), oldestByDepth.Depth)
	}

	// Find oldest by birth date
	var oldestByDate *query.AncestorPath
	var earliestDateStr string
	for _, ancestor := range ancestors {
		if ancestor.Ancestor != nil {
			birthDateStr := ancestor.Ancestor.GetBirthDate()
			if birthDateStr != "" {
				if earliestDateStr == "" || birthDateStr < earliestDateStr {
					earliestDateStr = birthDateStr
					oldestByDate = ancestor
				}
			}
		}
	}

	if oldestByDate != nil && earliestDateStr != "" {
		fmt.Printf("  ✅ Oldest ancestor by birth date: %s (born: %s)\n",
			oldestByDate.Ancestor.XrefID(), earliestDateStr)
	}

	fmt.Printf("     Total ancestors: %d, Duration: %v\n\n", len(ancestors), duration)
}

// testCommonAncestors tests finding common ancestors
func testCommonAncestors(t *testing.T, qb *query.QueryBuilder, testIndividuals []string, filename string) {
	fmt.Printf("3. Common Ancestors:\n")

	if len(testIndividuals) < 2 {
		return
	}

	fromXref := testIndividuals[0]
	toXref := testIndividuals[1]

	start := time.Now()
	common, err := qb.Individual(fromXref).CommonAncestors(toXref)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Found %d common ancestors between %s and %s\n", len(common), fromXref, toXref)
	if len(common) > 0 {
		fmt.Printf("     First common ancestor: %s\n", common[0].XrefID())
	}
	fmt.Printf("     Duration: %v\n\n", duration)
}

// testLowestCommonAncestor tests finding the lowest common ancestor
func testLowestCommonAncestor(t *testing.T, qb *query.QueryBuilder, testIndividuals []string, filename string) {
	fmt.Printf("4. Lowest Common Ancestor:\n")

	if len(testIndividuals) < 2 {
		return
	}

	fromXref := testIndividuals[0]
	toXref := testIndividuals[1]

	graph := qb.Graph()
	start := time.Now()
	lca, err := graph.LowestCommonAncestor(fromXref, toXref)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  ⚠️  No LCA found: %v\n", err)
		return
	}

	if lca != nil {
		fmt.Printf("  ✅ LCA: %s\n", lca.ID())
		fmt.Printf("     Duration: %v\n", duration)
	}
	fmt.Printf("\n")
}

// testPathFinding tests path finding between individuals
func testPathFinding(t *testing.T, qb *query.QueryBuilder, testIndividuals []string, filename string) {
	fmt.Printf("5. Path Finding:\n")

	if len(testIndividuals) < 2 {
		return
	}

	fromXref := testIndividuals[0]
	toXref := testIndividuals[1]

	// Shortest path
	start := time.Now()
	path, err := qb.Individual(fromXref).PathTo(toXref).Shortest()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  ⚠️  No path found: %v\n", err)
	} else if path != nil {
		fmt.Printf("  ✅ Shortest path: length %d\n", path.Length)
		fmt.Printf("     Duration: %v\n", duration)
	}

	// All paths (limited)
	start = time.Now()
	paths, err := qb.Individual(fromXref).PathTo(toXref).MaxLength(5).All()
	duration = time.Since(start)

	if err == nil {
		fmt.Printf("  ✅ Found %d paths (max length 5)\n", len(paths))
		fmt.Printf("     Duration: %v\n", duration)
	}
	fmt.Printf("\n")
}

// testSpecificRelationships tests specific relationship queries
func testSpecificRelationships(t *testing.T, qb *query.QueryBuilder, testIndividuals []string, filename string) {
	fmt.Printf("6. Specific Relationships:\n")

	if len(testIndividuals) == 0 {
		return
	}

	testXref := testIndividuals[0]

	// Cousins
	cousins, err := qb.Individual(testXref).Cousins(1)
	if err == nil {
		fmt.Printf("  ✅ 1st cousins: %d\n", len(cousins))
	}

	// Uncles
	uncles, err := qb.Individual(testXref).Uncles()
	if err == nil {
		fmt.Printf("  ✅ Uncles: %d\n", len(uncles))
	}

	// Nephews
	nephews, err := qb.Individual(testXref).Nephews()
	if err == nil {
		fmt.Printf("  ✅ Nephews: %d\n", len(nephews))
	}

	// Grandparents
	grandparents, err := qb.Individual(testXref).Grandparents()
	if err == nil {
		fmt.Printf("  ✅ Grandparents: %d\n", len(grandparents))
	}

	// Grandchildren
	grandchildren, err := qb.Individual(testXref).Grandchildren()
	if err == nil {
		fmt.Printf("  ✅ Grandchildren: %d\n", len(grandchildren))
	}

	fmt.Printf("\n")
}

// testBrickWalls tests finding individuals with no known parents
func testBrickWalls(t *testing.T, qb *query.QueryBuilder, filename string) {
	fmt.Printf("7. Brick Walls (No Parents):\n")

	start := time.Now()
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	var brickWalls []*types.IndividualRecord
	for _, indi := range allIndividuals {
		parents, err := qb.Individual(indi.XrefID()).Parents()
		if err == nil && len(parents) == 0 {
			brickWalls = append(brickWalls, indi)
		}
	}
	duration := time.Since(start)

	fmt.Printf("  ✅ Found %d individuals with no known parents\n", len(brickWalls))
	if len(brickWalls) > 0 && len(brickWalls) <= 5 {
		for _, indi := range brickWalls {
			fmt.Printf("     - %s: %s\n", indi.XrefID(), indi.GetName())
		}
	}
	fmt.Printf("     Duration: %v\n\n", duration)
}

// testEndOfLine tests finding individuals with no known children
func testEndOfLine(t *testing.T, qb *query.QueryBuilder, filename string) {
	fmt.Printf("8. End of Line (No Children):\n")

	start := time.Now()
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	var endOfLine []*types.IndividualRecord
	for _, indi := range allIndividuals {
		children, err := qb.Individual(indi.XrefID()).Children()
		if err == nil && len(children) == 0 {
			endOfLine = append(endOfLine, indi)
		}
	}
	duration := time.Since(start)

	fmt.Printf("  ✅ Found %d individuals with no known children\n", len(endOfLine))
	if len(endOfLine) > 0 && len(endOfLine) <= 5 {
		for _, indi := range endOfLine {
			fmt.Printf("     - %s: %s\n", indi.XrefID(), indi.GetName())
		}
	}
	fmt.Printf("     Duration: %v\n\n", duration)
}

// testMultipleSpouses tests finding individuals with multiple spouses
func testMultipleSpouses(t *testing.T, qb *query.QueryBuilder, filename string) {
	fmt.Printf("9. Multiple Spouses:\n")

	start := time.Now()
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	var multipleSpouses []*types.IndividualRecord
	for _, indi := range allIndividuals {
		spouses, err := qb.Individual(indi.XrefID()).Spouses()
		if err == nil && len(spouses) > 1 {
			multipleSpouses = append(multipleSpouses, indi)
		}
	}
	duration := time.Since(start)

	fmt.Printf("  ✅ Found %d individuals with multiple spouses\n", len(multipleSpouses))
	if len(multipleSpouses) > 0 && len(multipleSpouses) <= 5 {
		for _, indi := range multipleSpouses {
			spouses, _ := qb.Individual(indi.XrefID()).Spouses()
			fmt.Printf("     - %s: %s (%d spouses)\n", indi.XrefID(), indi.GetName(), len(spouses))
		}
	}
	fmt.Printf("     Duration: %v\n\n", duration)
}

// testMissingData tests finding individuals with missing data
func testMissingData(t *testing.T, qb *query.QueryBuilder, filename string) {
	fmt.Printf("10. Missing Data:\n")

	start := time.Now()
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	var noBirthDate, noDeathDate, noBirthPlace int
	for _, indi := range allIndividuals {
		if indi.GetBirthDate() == "" {
			noBirthDate++
		}
		if indi.GetDeathDate() == "" {
			noDeathDate++
		}
		if indi.GetBirthPlace() == "" {
			noBirthPlace++
		}
	}
	duration := time.Since(start)

	fmt.Printf("  ✅ No birth date: %d\n", noBirthDate)
	fmt.Printf("  ✅ No death date: %d (potentially living)\n", noDeathDate)
	fmt.Printf("  ✅ No birth place: %d\n", noBirthPlace)
	fmt.Printf("     Duration: %v\n\n", duration)
}

// testGeographicQueries tests geographic queries
func testGeographicQueries(t *testing.T, qb *query.QueryBuilder, filename string) {
	fmt.Printf("11. Geographic Queries:\n")

	// Get all places
	start := time.Now()
	places, err := qb.AllPlaces()
	duration := time.Since(start)

	if err == nil {
		fmt.Printf("  ✅ Total unique places: %d\n", len(places))
		if len(places) > 0 && len(places) <= 10 {
			for i, place := range places {
				if i >= 5 {
					break
				}
				fmt.Printf("     - %s\n", place)
			}
		}
		fmt.Printf("     Duration: %v\n", duration)
	}

	// Test filtering by birth place (if we have places)
	if len(places) > 0 {
		testPlace := places[0]
		start = time.Now()
		results, err := qb.Filter().ByBirthPlace(testPlace).Execute()
		duration = time.Since(start)

		if err == nil {
			fmt.Printf("  ✅ Born in '%s': %d individuals\n", testPlace, len(results))
			fmt.Printf("     Duration: %v\n", duration)
		}
	}

	fmt.Printf("\n")
}

// testTemporalQueries tests temporal queries
func testTemporalQueries(t *testing.T, qb *query.QueryBuilder, filename string) {
	fmt.Printf("12. Temporal Queries:\n")

	// Test date range queries
	startDate := time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(1900, 12, 31, 23, 59, 59, 0, time.UTC)

	start := time.Now()
	results, err := qb.Filter().ByBirthDate(startDate, endDate).Execute()
	duration := time.Since(start)

	if err == nil {
		fmt.Printf("  ✅ Born 1800-1900: %d individuals\n", len(results))
		fmt.Printf("     Duration: %v\n", duration)
	}

	// Test another range
	startDate2 := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate2 := time.Date(2000, 12, 31, 23, 59, 59, 0, time.UTC)

	start = time.Now()
	results2, err := qb.Filter().ByBirthDate(startDate2, endDate2).Execute()
	duration = time.Since(start)

	if err == nil {
		fmt.Printf("  ✅ Born 1900-2000: %d individuals\n", len(results2))
		fmt.Printf("     Duration: %v\n", duration)
	}

	fmt.Printf("\n")
}

// testNameBasedQueries tests name-based queries
func testNameBasedQueries(t *testing.T, qb *query.QueryBuilder, filename string) {
	fmt.Printf("13. Name-Based Queries:\n")

	// Get unique names
	start := time.Now()
	uniqueNames, err := qb.UniqueNames()
	duration := time.Since(start)

	if err == nil {
		fmt.Printf("  ✅ Unique names: %d\n", len(uniqueNames))
		count := 0
		for name, xrefs := range uniqueNames {
			if count >= 5 {
				break
			}
			fmt.Printf("     - %s: %d individuals\n", name, len(xrefs))
			count++
		}
		fmt.Printf("     Duration: %v\n", duration)
	}

	// Test name filter (if we have names)
	if len(uniqueNames) > 0 {
		// Get first name
		var testName string
		for name := range uniqueNames {
			testName = name
			break
		}

		start = time.Now()
		results, err := qb.Filter().ByName(testName).Execute()
		duration = time.Since(start)

		if err == nil {
			fmt.Printf("  ✅ Name filter '%s': %d individuals\n", testName, len(results))
			fmt.Printf("     Duration: %v\n", duration)
		}
	}

	fmt.Printf("\n")
}

// testGraphMetrics tests graph metrics (skip expensive operations on large files)
func testGraphMetrics(t *testing.T, qb *query.QueryBuilder, filename string) {
	fmt.Printf("14. Graph Metrics:\n")

	metrics := qb.Metrics()

	// Centrality (fast)
	start := time.Now()
	centrality, err := metrics.Centrality(query.CentralityDegree)
	duration := time.Since(start)

	if err == nil {
		maxDegree := 0.0
		mostConnected := ""
		for id, degree := range centrality {
			if degree > maxDegree {
				maxDegree = degree
				mostConnected = id
			}
		}
		fmt.Printf("  ✅ Most connected: %s (degree: %.2f)\n", mostConnected, maxDegree)
		fmt.Printf("     Duration: %v\n", duration)
	}

	// Skip diameter and connected components for large files (too slow)
	// Only run on smaller files
	if filename == "xavier.ged" || filename == "gracis.ged" {
		// Diameter (expensive - skip for large files)
		start = time.Now()
		diameter, err := metrics.Diameter()
		duration = time.Since(start)

		if err == nil {
			fmt.Printf("  ✅ Graph diameter: %d\n", diameter)
			fmt.Printf("     Duration: %v\n", duration)
		}

		// Connected components (expensive - skip for large files)
		start = time.Now()
		components, err := metrics.ConnectedComponents()
		duration = time.Since(start)

		if err == nil {
			fmt.Printf("  ✅ Connected components: %d\n", len(components))
			if len(components) > 0 {
				fmt.Printf("     Largest component: %d individuals\n", len(components[0]))
			}
			fmt.Printf("     Duration: %v\n", duration)
		}
	} else {
		fmt.Printf("  ⚠️  Skipping diameter and connected components (too slow for large files)\n")
	}

	fmt.Printf("\n")
}

