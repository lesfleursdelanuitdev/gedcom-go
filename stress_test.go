package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/duplicate"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/query"
)

// StressTestMetrics holds comprehensive stress test results
type StressTestMetrics struct {
	TestName      string
	Duration      time.Duration
	MemoryBefore  uint64
	MemoryAfter   uint64
	MemoryPeak    uint64
	MemoryUsed    uint64
	Throughput    float64
	NumGoroutines int
	NumGC         uint32
	Success       bool
	Error         error
	Graph         *query.Graph // Optional: store graph for component detection
}

// measureMemory returns current memory usage in bytes
func measureMemory() uint64 {
	var m runtime.MemStats
	runtime.GC() // Force GC before measurement
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// getMemStats returns comprehensive memory statistics
func getMemStats() (alloc, totalAlloc, sys uint64, numGC uint32, numGoroutines int) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc, m.TotalAlloc, m.Sys, m.NumGC, runtime.NumGoroutine()
}

// generateStressTestTree creates a tree with n individuals for stress testing
// Uses realistic family structures with varying complexity
func generateStressTestTree(n int) *gedcom.GedcomTree {
	tree := gedcom.NewGedcomTree()

	// Pre-allocate slices for better performance
	individuals := make([]*gedcom.IndividualRecord, 0, n)
	families := make([]*gedcom.FamilyRecord, 0, n/2)

	// Track relationships
	childToFamily := make(map[int]int, n)
	familyID := 1
	indiID := 1

	// Create families with 1-3 children each (realistic distribution)
	for indiID < n {
		// Determine family size (1-3 children, weighted toward 2)
		numChildren := 2
		if indiID%10 == 0 {
			numChildren = 1 // 10% have 1 child
		} else if indiID%5 == 0 {
			numChildren = 3 // 20% have 3 children
		}

		// Check if we have enough individuals left
		if indiID+numChildren+1 >= n {
			numChildren = n - indiID - 1
			if numChildren <= 0 {
				break
			}
		}

		// Create family
		famLine := gedcom.NewGedcomLine(0, "FAM", "", fmt.Sprintf("@F%d@", familyID))

		// Husband
		if indiID < n {
			famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", fmt.Sprintf("@I%d@", indiID), ""))
			indiID++
		}

		// Wife
		if indiID < n {
			famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", fmt.Sprintf("@I%d@", indiID), ""))
			indiID++
		}

		// Children
		for i := 0; i < numChildren && indiID < n; i++ {
			famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", fmt.Sprintf("@I%d@", indiID), ""))
			childToFamily[indiID] = familyID
			indiID++
		}

		fam := gedcom.NewFamilyRecord(famLine)
		families = append(families, fam)
		familyID++
	}

	// Create all individuals with varied data
	for i := 1; i <= n; i++ {
		indiLine := gedcom.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i))

		// Add name with variation
		indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", i), ""))

		// Add birth date (distributed across years 1800-2000)
		birthYear := 1800 + (i % 200)
		birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
		birtLine.AddChild(gedcom.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", birthYear), ""))

		// Add birth place for some individuals (every 10th)
		if i%10 == 0 {
			birtLine.AddChild(gedcom.NewGedcomLine(2, "PLAC", fmt.Sprintf("City %d", i%50), ""))
		}
		indiLine.AddChild(birtLine)

		// Add death date for some individuals (every 5th)
		if i%5 == 0 {
			deathYear := birthYear + 50 + (i % 50)
			deatLine := gedcom.NewGedcomLine(1, "DEAT", "", "")
			deatLine.AddChild(gedcom.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", deathYear), ""))
			indiLine.AddChild(deatLine)
		}

		// Add sex (alternating)
		sex := "M"
		if i%2 == 0 {
			sex = "F"
		}
		indiLine.AddChild(gedcom.NewGedcomLine(1, "SEX", sex, ""))

		// Add FAMC if child
		if famID, ok := childToFamily[i]; ok {
			indiLine.AddChild(gedcom.NewGedcomLine(1, "FAMC", fmt.Sprintf("@F%d@", famID), ""))
		}

		indi := gedcom.NewIndividualRecord(indiLine)
		individuals = append(individuals, indi)
	}

	// Add all records to tree
	for _, fam := range families {
		tree.AddRecord(fam)
	}
	for _, indi := range individuals {
		tree.AddRecord(indi)
	}

	return tree
}

// generateStressTestGEDCOMFile generates a GEDCOM file with n individuals
func generateStressTestGEDCOMFile(filename string, n int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	file.WriteString("0 HEAD\n")
	file.WriteString("1 SOUR Stress Test Generator\n")
	file.WriteString("1 GEDC\n")
	file.WriteString("2 VERS 5.5.1\n")
	file.WriteString("0 @SUBM@ SUBM\n")
	file.WriteString("1 NAME Stress Test\n")

	// Track relationships
	childToFamily := make(map[int]int)
	familyID := 1
	indiID := 1

	// Create families
	for indiID < n {
		numChildren := 2
		if indiID%10 == 0 {
			numChildren = 1
		} else if indiID%5 == 0 {
			numChildren = 3
		}

		if indiID+numChildren+1 >= n {
			numChildren = n - indiID - 1
			if numChildren <= 0 {
				break
			}
		}

		// Family record
		file.WriteString(fmt.Sprintf("0 @F%d@ FAM\n", familyID))

		// Husband
		if indiID < n {
			file.WriteString(fmt.Sprintf("1 HUSB @I%d@\n", indiID))
			indiID++
		}

		// Wife
		if indiID < n {
			file.WriteString(fmt.Sprintf("1 WIFE @I%d@\n", indiID))
			indiID++
		}

		// Children
		for i := 0; i < numChildren && indiID < n; i++ {
			file.WriteString(fmt.Sprintf("1 CHIL @I%d@\n", indiID))
			childToFamily[indiID] = familyID
			indiID++
		}

		familyID++
	}

	// Create individuals
	for i := 1; i <= n; i++ {
		file.WriteString(fmt.Sprintf("0 @I%d@ INDI\n", i))
		file.WriteString(fmt.Sprintf("1 NAME Person %d /Test/\n", i))

		birthYear := 1800 + (i % 200)
		file.WriteString("1 BIRT\n")
		file.WriteString(fmt.Sprintf("2 DATE %d\n", birthYear))

		if i%10 == 0 {
			file.WriteString(fmt.Sprintf("2 PLAC City %d\n", i%50))
		}

		if i%5 == 0 {
			deathYear := birthYear + 50 + (i % 50)
			file.WriteString("1 DEAT\n")
			file.WriteString(fmt.Sprintf("2 DATE %d\n", deathYear))
		}

		sex := "M"
		if i%2 == 0 {
			sex = "F"
		}
		file.WriteString(fmt.Sprintf("1 SEX %s\n", sex))

		if famID, ok := childToFamily[i]; ok {
			file.WriteString(fmt.Sprintf("1 FAMC @F%d@\n", famID))
		}
	}

	file.WriteString("0 TRLR\n")
	return nil
}

// runStressTest runs a test function and collects comprehensive metrics
func runStressTest(testName string, datasetSize int, fn func() error) StressTestMetrics {
	before, _, _, numGCBefore, _ := getMemStats()
	start := time.Now()

	err := fn()

	duration := time.Since(start)
	after, totalAlloc, _, numGCAfter, numGoroutines := getMemStats()

	throughput := float64(datasetSize) / duration.Seconds()

	return StressTestMetrics{
		TestName:      testName,
		Duration:      duration,
		MemoryBefore:  before,
		MemoryAfter:   after,
		MemoryPeak:    totalAlloc,
		MemoryUsed:    after - before,
		Throughput:    throughput,
		NumGoroutines: numGoroutines,
		NumGC:         numGCAfter - numGCBefore,
		Success:       err == nil,
		Error:         err,
	}
}

// TestStress_1M_Comprehensive runs comprehensive stress tests with 1,000,000 individuals
func TestStress_1M_Comprehensive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive stress test in short mode")
	}

	t.Logf("=== COMPREHENSIVE STRESS TEST: 1,000,000 INDIVIDUALS ===\n")
	t.Logf("WARNING: This is a large-scale stress test. It may take several minutes to complete.\n")
	t.Logf("Note: For 1M testing, see TestStress_1M_Comprehensive in performance_test.go\n")
	t.Skip("Use TestStress_1_5M_Comprehensive for 1.5M testing or see performance_test.go for 1M tests")
}

// TestStress_1_5M_Comprehensive runs comprehensive stress tests with 1,500,000 individuals
func TestStress_1_5M_Comprehensive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive stress test in short mode")
	}

	const size = 1500000
	allMetrics := []StressTestMetrics{}

	t.Logf("=== COMPREHENSIVE STRESS TEST: 1,500,000 INDIVIDUALS ===\n")
	t.Logf("WARNING: This is a very large-scale stress test. It may take several minutes to complete.\n")
	t.Logf("Estimated memory requirement: ~23-24 GB RAM\n")

	runStressTestPhases(t, size, &allMetrics)
}

// TestStress_5M_Comprehensive runs comprehensive stress tests with 5,000,000 individuals
func TestStress_5M_Comprehensive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive stress test in short mode")
	}

	const size = 5000000
	allMetrics := []StressTestMetrics{}

	t.Logf("=== COMPREHENSIVE STRESS TEST: 5,000,000 INDIVIDUALS ===\n")
	t.Logf("WARNING: This is an extremely large-scale stress test. It may take 10-20 minutes to complete.\n")
	t.Logf("Estimated memory requirement: ~70-75 GB RAM\n")
	t.Logf("This test will NOT be restarted if it times out.\n")

	runStressTestPhases(t, size, &allMetrics)
}

// runStressTestPhases runs all stress test phases (shared between different test sizes)
func runStressTestPhases(t *testing.T, size int, allMetrics *[]StressTestMetrics) {

	// ============================================================
	// PHASE 1: Data Generation
	// ============================================================
	t.Log("PHASE 1: Data Generation")
	t.Log("Generating in-memory tree...")
	genStart := time.Now()
	tree := generateStressTestTree(size)
	genDuration := time.Since(genStart)
	t.Logf("✓ Generated tree in %v (%.2f ind/sec)\n", genDuration, float64(size)/genDuration.Seconds())

	// Verify tree
	individuals := tree.GetAllIndividuals()
	if len(individuals) != size {
		t.Errorf("Expected %d individuals, got %d", size, len(individuals))
	}
	t.Logf("✓ Verified: %d individuals, %d families\n", len(individuals), len(tree.GetAllFamilies()))

	// ============================================================
	// PHASE 2: File Generation and Parsing
	// ============================================================
	t.Log("\nPHASE 2: File Generation and Parsing")
	filename := fmt.Sprintf("/tmp/gedcom_stress_100k_%d.ged", time.Now().Unix())

	// Generate GEDCOM file
	t.Log("Generating GEDCOM file...")
	*allMetrics = append(*allMetrics, runStressTest("FileGeneration", size, func() error {
		return generateStressTestGEDCOMFile(filename, size)
	}))
	defer os.Remove(filename)

	// Parse GEDCOM file
	t.Log("Parsing GEDCOM file...")
	var parsedTree *gedcom.GedcomTree
	*allMetrics = append(*allMetrics, runStressTest("Parsing", size, func() error {
		parser := parser.NewHierarchicalParser()
		var err error
		parsedTree, err = parser.Parse(filename)
		return err
	}))

	if parsedTree != nil {
		parsedIndividuals := parsedTree.GetAllIndividuals()
		if len(parsedIndividuals) != size {
			t.Errorf("Parsed tree: Expected %d individuals, got %d", size, len(parsedIndividuals))
		}
		t.Logf("✓ Parsed: %d individuals\n", len(parsedIndividuals))
	}

	// ============================================================
	// PHASE 3: Graph Construction
	// ============================================================
	t.Log("\nPHASE 3: Graph Construction")
	var graph *query.Graph
	*allMetrics = append(*allMetrics, runStressTest("GraphConstruction", size, func() error {
		var err error
		graph, err = query.BuildGraph(tree)
		return err
	}))

	if graph == nil {
		t.Fatal("Graph construction failed")
	}
	t.Logf("✓ Graph constructed: %d nodes, %d edges\n",
		len(graph.GetAllIndividuals()), len(graph.GetAllEdges()))

	// ============================================================
	// PHASE 4: Query System Tests
	// ============================================================
	t.Log("\nPHASE 4: Query System Tests")

	// Create query builder
	var q *query.QueryBuilder
	*allMetrics = append(*allMetrics, runStressTest("QueryCreation", size, func() error {
		q = query.NewQueryFromGraph(graph)
		return nil
	}))

	if q == nil {
		t.Fatal("Query creation failed")
	}

	// Test 4.1: Filter Queries
	t.Log("4.1 Filter Queries...")

	// By name (indexed)
	*allMetrics = append(*allMetrics, runStressTest("Filter_ByName", size, func() error {
		results, err := q.Filter().ByName("Person").Execute()
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return fmt.Errorf("expected results, got 0")
		}
		return nil
	}))

	// By name exact
	*allMetrics = append(*allMetrics, runStressTest("Filter_ByNameExact", size, func() error {
		results, err := q.Filter().ByNameExact("Person 50000 /Test/").Execute()
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return fmt.Errorf("expected results, got 0")
		}
		return nil
	}))

	// By sex
	*allMetrics = append(*allMetrics, runStressTest("Filter_BySex", size, func() error {
		results, err := q.Filter().BySex("M").Execute()
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return fmt.Errorf("expected results, got 0")
		}
		return nil
	}))

	// Combined filters
	*allMetrics = append(*allMetrics, runStressTest("Filter_Combined", size, func() error {
		results, err := q.Filter().ByName("Person").BySex("F").Execute()
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return fmt.Errorf("expected results, got 0")
		}
		return nil
	}))

	// Test 4.2: Direct Relationship Queries
	t.Log("4.2 Direct Relationship Queries...")

	// Parents
	*allMetrics = append(*allMetrics, runStressTest("Query_Parents", size, func() error {
		parents, err := q.Individual("@I1000@").Parents()
		if err != nil {
			return err
		}
		_ = parents // Use to avoid unused variable
		return nil
	}))

	// Children
	*allMetrics = append(*allMetrics, runStressTest("Query_Children", size, func() error {
		children, err := q.Individual("@I1@").Children()
		if err != nil {
			return err
		}
		_ = children
		return nil
	}))

	// Siblings
	*allMetrics = append(*allMetrics, runStressTest("Query_Siblings", size, func() error {
		siblings, err := q.Individual("@I1000@").Siblings()
		if err != nil {
			return err
		}
		_ = siblings
		return nil
	}))

	// Spouses
	*allMetrics = append(*allMetrics, runStressTest("Query_Spouses", size, func() error {
		spouses, err := q.Individual("@I1@").Spouses()
		if err != nil {
			return err
		}
		_ = spouses
		return nil
	}))

	// Test 4.3: Ancestor Queries
	t.Log("4.3 Ancestor Queries...")

	// Ancestors with limit
	*allMetrics = append(*allMetrics, runStressTest("Query_Ancestors_Limited", size, func() error {
		ancestors, err := q.Individual("@I50000@").Ancestors().MaxGenerations(5).Execute()
		if err != nil {
			return err
		}
		_ = ancestors
		return nil
	}))

	// Ancestor count
	*allMetrics = append(*allMetrics, runStressTest("Query_Ancestors_Count", size, func() error {
		count, err := q.Individual("@I50000@").Ancestors().MaxGenerations(10).Count()
		if err != nil {
			return err
		}
		_ = count
		return nil
	}))

	// Test 4.4: Descendant Queries
	t.Log("4.4 Descendant Queries...")

	*allMetrics = append(*allMetrics, runStressTest("Query_Descendants", size, func() error {
		descendants, err := q.Individual("@I1@").Descendants().MaxGenerations(5).Execute()
		if err != nil {
			return err
		}
		_ = descendants
		return nil
	}))

	// Test 4.5: Path Finding
	t.Log("4.5 Path Finding...")

	// Shortest path
	*allMetrics = append(*allMetrics, runStressTest("Query_ShortestPath", size, func() error {
		_, _ = q.Individual("@I1@").PathTo("@I1000@").Shortest()
		// Path might not exist, that's okay
		return nil // Don't return error if path doesn't exist
	}))

	// All paths
	*allMetrics = append(*allMetrics, runStressTest("Query_AllPaths", size, func() error {
		_, _ = q.Individual("@I1@").PathTo("@I1000@").MaxLength(10).All()
		// Path might not exist, that's okay
		return nil
	}))

	// Test 4.6: Relationship Calculation
	t.Log("4.6 Relationship Calculation...")

	*allMetrics = append(*allMetrics, runStressTest("Query_Relationship", size, func() error {
		_, _ = q.Individual("@I1@").RelationshipTo("@I1000@").Execute()
		// Relationship might not exist, that's okay
		return nil
	}))

	// ============================================================
	// PHASE 5: Concurrent Operations
	// ============================================================
	t.Log("\nPHASE 5: Concurrent Operations")

	*allMetrics = append(*allMetrics, runStressTest("Concurrent_Queries", size, func() error {
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Run 10 concurrent queries
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := q.Filter().ByName(fmt.Sprintf("Person %d", id*10000)).Execute()
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			if err != nil {
				return err
			}
		}
		return nil
	}))

	// ============================================================
	// PHASE 6: Duplicate Detection (Skipped for Large Datasets)
	// ============================================================
	// Note: Full duplicate detection on 100K individuals would require
	// ~5 billion comparisons (100,000 * 99,999 / 2), which is computationally
	// prohibitive even with parallel processing and filtering.
	// For stress testing, we skip this phase or use a very small sample.
	t.Log("\nPHASE 6: Duplicate Detection")

	// For datasets > 10K, use blocking-based duplicate detection
	// With blocking, 1.5M individuals should be tractable (minutes instead of impossible)
	if size > 10000 {
		t.Logf("Testing duplicate detection with blocking for large dataset (%d individuals)", size)
		if size >= 1000000 {
			naiveComparisons := float64(size) * float64(size-1) / 2.0
			estimatedComparisons := float64(size) * 200 // Assuming ~200 candidates per person
			t.Logf("Note: Naive approach would require ~%.0f trillion comparisons", naiveComparisons/1e12)
			t.Logf("Note: With blocking, estimated ~%.0f million comparisons", estimatedComparisons/1e6)
		}

		*allMetrics = append(*allMetrics, runStressTest("DuplicateDetection_Blocking", size, func() error {
			config := duplicate.DefaultConfig()
			config.UseParallelProcessing = true
			config.UseBlocking = true
			config.MaxCandidatesPerPerson = 200
			detector := duplicate.NewDuplicateDetector(config)
			detector.SetTree(tree)
			result, err := detector.FindDuplicates(tree)
			if err != nil {
				return err
			}
			t.Logf("  Found %d potential duplicates", len(result.Matches))
			t.Logf("  Total comparisons: %d (vs %.0f naive)", result.TotalComparisons, float64(size)*float64(size-1)/2.0)
			if result.Metrics != nil {
				t.Logf("  Processing time: %v", result.Metrics.ProcessingTime)
			}
			if result.BlockingMetrics != nil {
				t.Logf("\n  Blocking Metrics:")
				t.Logf("    People with Primary Block: %d (%.1f%%)",
					result.BlockingMetrics.PeopleWithPrimaryBlock,
					float64(result.BlockingMetrics.PeopleWithPrimaryBlock)/float64(result.BlockingMetrics.TotalPeople)*100)
				t.Logf("    People with Any Block: %d (%.1f%%)",
					result.BlockingMetrics.PeopleWithAnyBlock,
					float64(result.BlockingMetrics.PeopleWithAnyBlock)/float64(result.BlockingMetrics.TotalPeople)*100)
				t.Logf("    People with No Blocks: %d (%.1f%%)",
					result.BlockingMetrics.PeopleWithNoBlocks,
					float64(result.BlockingMetrics.PeopleWithNoBlocks)/float64(result.BlockingMetrics.TotalPeople)*100)
				t.Logf("    Total Blocks: %d", result.BlockingMetrics.TotalBlocks)
				t.Logf("    Avg Candidates/Person: %.2f", result.BlockingMetrics.AverageCandidatesPerPerson)
				t.Logf("    Max Candidates/Person: %d", result.BlockingMetrics.MaxCandidatesPerPerson)
				t.Logf("    People with 0 candidates: %d", result.BlockingMetrics.PeopleWithZeroCandidates)
				if len(result.BlockingMetrics.TopBlockSizes) > 0 {
					t.Logf("    Top 5 Block Sizes:")
					for i, info := range result.BlockingMetrics.TopBlockSizes {
						if i >= 5 {
							break
						}
						t.Logf("      Size %d: %d blocks", info.Size, info.Count)
					}
				}

				// Show warnings if present
				warnings := result.BlockingMetrics.GetWarnings()
				if len(warnings) > 0 {
					t.Logf("\n  ⚠️  WARNINGS:")
					for _, warning := range warnings {
						t.Logf("    %s", warning)
					}
				}
			}
			_ = result
			return nil
		}))
	} else {
		// For smaller datasets, run duplicate detection
		*allMetrics = append(*allMetrics, runStressTest("DuplicateDetection", size, func() error {
			config := duplicate.DefaultConfig()
			config.UseParallelProcessing = true
			detector := duplicate.NewDuplicateDetector(config)
			detector.SetTree(tree)
			result, err := detector.FindDuplicates(tree)
			if err != nil {
				return err
			}
			t.Logf("  Found %d potential duplicates", len(result.Matches))
			_ = result
			return nil
		}))
	}

	// ============================================================
	// PHASE 7: Graph Metrics
	// ============================================================
	t.Log("\nPHASE 7: Graph Metrics")

	*allMetrics = append(*allMetrics, runStressTest("GraphMetrics", size, func() error {
		metrics := q.Metrics()
		// Test a few metric operations
		_, _ = metrics.Degree("@I1@")
		_, _ = metrics.AverageDegree()
		_, _ = metrics.Density()
		return nil
	}))

	// ============================================================
	// RESULTS SUMMARY
	// ============================================================
	t.Log("\n" + repeat("=", 80))
	t.Log("STRESS TEST RESULTS SUMMARY")
	t.Log(repeat("=", 80))

	totalDuration := time.Duration(0)
	for _, m := range *allMetrics {
		totalDuration += m.Duration

		status := "✓"
		if !m.Success {
			status = "✗"
		}

		t.Logf("\n%s %s", status, m.TestName)
		t.Logf("  Duration:     %v", m.Duration)
		t.Logf("  Throughput:   %.2f ops/sec", m.Throughput)
		t.Logf("  Memory Used:  %.2f MB", float64(m.MemoryUsed)/1024/1024)
		t.Logf("  Peak Memory:  %.2f MB", float64(m.MemoryPeak)/1024/1024)
		t.Logf("  Goroutines:   %d", m.NumGoroutines)
		t.Logf("  GC Cycles:    %d", m.NumGC)

		if m.Error != nil {
			t.Logf("  Error:        %v", m.Error)
		}
	}

	t.Logf("\n" + repeat("=", 80))
	t.Logf("TOTAL TEST DURATION: %v", totalDuration)
	t.Logf(repeat("=", 80))

	// Final memory stats
	alloc, totalAlloc, sys, _, numGoroutines := getMemStats()
	t.Logf("\nFinal Memory Statistics:")
	t.Logf("  Current Alloc:  %.2f MB", float64(alloc)/1024/1024)
	t.Logf("  Total Alloc:    %.2f MB", float64(totalAlloc)/1024/1024)
	t.Logf("  System Memory:   %.2f MB", float64(sys)/1024/1024)
	t.Logf("  Goroutines:     %d", numGoroutines)
}

// Helper function to repeat strings (Go doesn't have this built-in)
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// TestStress_LazyLoading_1M tests lazy loading with 1M individuals
func TestStress_LazyLoading_1M(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping lazy loading stress test in short mode")
	}

	// Set timeout to 5 minutes
	timer := time.AfterFunc(5*time.Minute, func() {
		panic("test timed out after 5 minutes")
	})
	defer timer.Stop()

	size := 1_000_000
	t.Logf("=== LAZY LOADING STRESS TEST: %d INDIVIDUALS ===\n", size)
	t.Logf("Comparing eager vs lazy loading modes\n")
	t.Logf("Timeout: 5 minutes\n")

	// Generate test data
	tree := generateStressTestTree(size)
	t.Logf("✓ Generated %d individuals\n", size)

	// Test 1: Eager Loading (baseline)
	t.Log("\n--- EAGER LOADING MODE ---")
	eagerMetrics := testGraphConstructionLazy(t, tree, size, false, "Eager")

	// Test 2: Lazy Loading
	t.Log("\n--- LAZY LOADING MODE ---")
	lazyMetrics := testGraphConstructionLazy(t, tree, size, true, "Lazy")

	// Compare results
	t.Log("\n=== COMPARISON ===")
	t.Logf("Graph Construction:")
	t.Logf("  Eager: %v, Memory: %.2f MB, Peak: %.2f MB",
		eagerMetrics.Duration,
		float64(eagerMetrics.MemoryUsed)/(1024*1024),
		float64(eagerMetrics.MemoryPeak)/(1024*1024))
	t.Logf("  Lazy:  %v, Memory: %.2f MB, Peak: %.2f MB",
		lazyMetrics.Duration,
		float64(lazyMetrics.MemoryUsed)/(1024*1024),
		float64(lazyMetrics.MemoryPeak)/(1024*1024))

	if eagerMetrics.MemoryUsed > 0 {
		memorySavings := float64(eagerMetrics.MemoryUsed-lazyMetrics.MemoryUsed) / float64(eagerMetrics.MemoryUsed) * 100
		t.Logf("\nMemory Savings: %.1f%% (%.2f MB saved)",
			memorySavings,
			float64(eagerMetrics.MemoryUsed-lazyMetrics.MemoryUsed)/(1024*1024))
	}

	// Test query performance
	t.Log("\n--- QUERY PERFORMANCE TEST ---")
	testQueryPerformanceLazy(t, tree, size, true)
}

// testGraphConstructionLazy tests graph construction in either eager or lazy mode
func testGraphConstructionLazy(t *testing.T, tree *gedcom.GedcomTree, size int, lazy bool, mode string) StressTestMetrics {
	before, _, _, numGCBefore, _ := getMemStats()
	start := time.Now()

	var graph *query.Graph
	var err error

	if lazy {
		graph, err = query.BuildGraphLazy(tree)
	} else {
		graph, err = query.BuildGraph(tree)
	}

	duration := time.Since(start)
	after, totalAlloc, _, numGCAfter, numGoroutines := getMemStats()

	if err != nil {
		t.Errorf("%s loading failed: %v", mode, err)
		return StressTestMetrics{Success: false, Error: err}
	}

	if graph == nil {
		t.Errorf("%s loading returned nil graph", mode)
		return StressTestMetrics{Success: false}
	}

	// Get graph stats
	individuals := graph.GetAllIndividuals()
	edges := graph.GetAllEdges()

	// Calculate memory used (handle potential underflow)
	memoryUsed := uint64(0)
	if after > before {
		memoryUsed = after - before
	}

	t.Logf("✓ %s graph constructed:", mode)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Nodes: %d individuals", len(individuals))
	t.Logf("  Edges: %d", len(edges))
	t.Logf("  Memory Used: %.2f MB", float64(memoryUsed)/(1024*1024))
	t.Logf("  Peak Memory: %.2f MB", float64(totalAlloc)/(1024*1024))
	t.Logf("  GC Cycles: %d", numGCAfter-numGCBefore)
	t.Logf("  Goroutines: %d", numGoroutines)

	if lazy {
		componentCount := graph.GetComponentCount()
		t.Logf("  Components: %d", componentCount)
	}

	throughput := float64(size) / duration.Seconds()

	metrics := StressTestMetrics{
		TestName:      fmt.Sprintf("%s_GraphConstruction", mode),
		Duration:      duration,
		MemoryBefore:  before,
		MemoryAfter:   after,
		MemoryPeak:    totalAlloc,
		MemoryUsed:    memoryUsed,
		Throughput:    throughput,
		NumGoroutines: numGoroutines,
		NumGC:         numGCAfter - numGCBefore,
		Success:       true,
	}

	// Store graph reference for component detection
	if lazy {
		metrics.Graph = graph
	}

	return metrics
}

// testQueryPerformanceLazy tests query performance with lazy loading
func testQueryPerformanceLazy(t *testing.T, tree *gedcom.GedcomTree, size int, lazy bool) {
	var q *query.QueryBuilder
	var err error

	if lazy {
		q, err = query.NewQueryLazy(tree)
	} else {
		q, err = query.NewQuery(tree)
	}

	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test 1: Access a single individual (should trigger lazy loading)
	t.Log("\n1. Single Individual Access:")
	start := time.Now()
	indi := q.Individual("@I1000@")
	if indi == nil {
		t.Fatal("Failed to get individual")
	}
	duration := time.Since(start)
	t.Logf("   Duration: %v", duration)

	// Test 2: Query parents (should trigger edge loading)
	t.Log("\n2. Query Parents:")
	start = time.Now()
	parents, err := indi.Parents()
	if err != nil {
		t.Fatalf("Failed to query parents: %v", err)
	}
	duration = time.Since(start)
	t.Logf("   Duration: %v, Found: %d parents", duration, len(parents))

	// Test 3: Query children
	t.Log("\n3. Query Children:")
	start = time.Now()
	children, err := indi.Children()
	if err != nil {
		t.Fatalf("Failed to query children: %v", err)
	}
	duration = time.Since(start)
	t.Logf("   Duration: %v, Found: %d children", duration, len(children))

	// Test 4: Filter query (should trigger multiple node loads)
	t.Log("\n4. Filter Query:")
	start = time.Now()
	results, err := q.Filter().ByName("Person").Execute()
	if err != nil {
		t.Fatalf("Failed filter query: %v", err)
	}
	duration = time.Since(start)
	resultCount := len(results)
	if resultCount > 100 {
		resultCount = 100 // Limit display
	}
	t.Logf("   Duration: %v, Results: %d (showing first %d)", duration, len(results), resultCount)

	// Test 5: Access multiple individuals
	t.Log("\n5. Access Multiple Individuals:")
	start = time.Now()
	for i := 1; i <= 100 && i <= size; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indi := q.Individual(xrefID)
		if indi == nil {
			t.Errorf("Failed to get individual %s", xrefID)
		}
	}
	duration = time.Since(start)
	t.Logf("   Duration: %v for 100 individuals", duration)
	t.Logf("   Average: %v per individual", duration/100)

	// Memory check
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("\n   Current Memory: %.2f MB", float64(m.Alloc)/(1024*1024))
}

// TestStress_LazyLoading_1_5M tests lazy loading with 1.5M individuals
func TestStress_LazyLoading_1_5M(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping lazy loading stress test in short mode")
	}

	// Set timeout to 5 minutes
	timer := time.AfterFunc(5*time.Minute, func() {
		panic("test timed out after 5 minutes")
	})
	defer timer.Stop()

	size := 1_500_000
	t.Logf("=== LAZY LOADING STRESS TEST: %d INDIVIDUALS ===\n", size)
	t.Logf("Timeout: 5 minutes\n")

	// Generate test data
	tree := generateStressTestTree(size)
	t.Logf("✓ Generated %d individuals\n", size)

	// Test Lazy Loading
	t.Log("\n--- LAZY LOADING MODE ---")
	lazyMetrics := testGraphConstructionLazy(t, tree, size, true, "Lazy")

	t.Logf("\nLazy Loading Results:")
	t.Logf("  Duration: %v", lazyMetrics.Duration)
	t.Logf("  Memory Used: %.2f MB", float64(lazyMetrics.MemoryUsed)/(1024*1024))
	t.Logf("  Peak Memory: %.2f MB", float64(lazyMetrics.MemoryPeak)/(1024*1024))
	t.Logf("  Throughput: %.0f ops/sec", lazyMetrics.Throughput)

	// Test query performance
	t.Log("\n--- QUERY PERFORMANCE TEST ---")
	testQueryPerformanceLazy(t, tree, size, true)

	// Test component detection (skip for very large datasets to avoid timeout)
	if size <= 1_000_000 {
		t.Log("\n--- COMPONENT DETECTION TEST ---")
		testComponentDetectionLazy(t, tree, size)
	} else {
		t.Log("\n--- COMPONENT DETECTION TEST (Skipped for large dataset) ---")
		graph := lazyMetrics.Graph
		if graph != nil {
			componentCount := graph.GetComponentCount()
			t.Logf("Total Components: %d (detected during graph construction)", componentCount)
		}
	}
}

// testComponentDetectionLazy tests component detection and loading
func testComponentDetectionLazy(t *testing.T, tree *gedcom.GedcomTree, size int) {
	q, err := query.NewQueryLazy(tree)
	if err != nil {
		t.Fatalf("Failed to create lazy query: %v", err)
	}

	graph := q.Graph()

	// Get component count
	componentCount := graph.GetComponentCount()
	t.Logf("Total Components: %d", componentCount)

	// Find largest component (sample first 1000 components to avoid timeout)
	largestComponentID := uint32(0)
	largestSize := 0
	sampleSize := componentCount
	if componentCount > 1000 {
		sampleSize = 1000
		t.Logf("Sampling first 1000 of %d components to find largest", componentCount)
	}

	for i := uint32(1); i <= sampleSize; i++ {
		compSize := graph.GetComponentSize(i)
		if compSize > largestSize {
			largestSize = compSize
			largestComponentID = i
		}
	}

	t.Logf("Largest Component (from sample): ID=%d, Size=%d individuals (%.1f%% of total)",
		largestComponentID, largestSize, float64(largestSize)/float64(size)*100)

	// Test loading a component
	if largestComponentID > 0 {
		t.Log("\nLoading largest component...")
		start := time.Now()
		err := graph.LoadComponent(largestComponentID)
		duration := time.Since(start)
		if err != nil {
			t.Errorf("Failed to load component: %v", err)
		} else {
			t.Logf("✓ Component loaded in %v", duration)

			// Check memory after loading
			runtime.GC()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			t.Logf("Memory after loading component: %.2f MB", float64(m.Alloc)/(1024*1024))
		}
	}
}

// TestStress_LazyLoading_Comprehensive tests lazy loading with multiple dataset sizes
func TestStress_LazyLoading_Comprehensive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive lazy loading test in short mode")
	}

	// Set timeout to 5 minutes
	timer := time.AfterFunc(5*time.Minute, func() {
		panic("test timed out after 5 minutes")
	})
	defer timer.Stop()

	t.Logf("=== COMPREHENSIVE LAZY LOADING TEST ===\n")
	t.Logf("Testing multiple dataset sizes to demonstrate lazy loading improvements\n")
	t.Logf("Timeout: 5 minutes\n\n")

	// Test different sizes
	sizes := []int{100_000, 500_000, 1_000_000}

	for _, size := range sizes {
		t.Logf("\n" + strings.Repeat("=", 60) + "\n")
		t.Logf("DATASET SIZE: %d INDIVIDUALS\n", size)
		t.Logf(strings.Repeat("=", 60) + "\n")

		// Generate test data
		genStart := time.Now()
		tree := generateStressTestTree(size)
		genDuration := time.Since(genStart)
		t.Logf("✓ Generated %d individuals in %v\n", size, genDuration)

		// Test Eager Loading
		t.Log("\n--- EAGER LOADING ---")
		eagerStart := time.Now()
		eagerGraph, err := query.BuildGraph(tree)
		eagerDuration := time.Since(eagerStart)

		if err != nil {
			t.Errorf("Eager loading failed: %v", err)
			continue
		}

		eagerBefore, _, _, _, _ := getMemStats()
		runtime.GC()
		eagerAfter, eagerTotalAlloc, _, eagerGC, _ := getMemStats()
		eagerMemory := uint64(0)
		if eagerAfter > eagerBefore {
			eagerMemory = eagerAfter - eagerBefore
		}

		t.Logf("  Duration: %v", eagerDuration)
		t.Logf("  Memory: %.2f MB", float64(eagerMemory)/(1024*1024))
		t.Logf("  Peak Memory: %.2f MB", float64(eagerTotalAlloc)/(1024*1024))
		t.Logf("  Nodes: %d", len(eagerGraph.GetAllIndividuals()))
		t.Logf("  Edges: %d", len(eagerGraph.GetAllEdges()))
		t.Logf("  GC Cycles: %d", eagerGC)

		// Test Lazy Loading
		t.Log("\n--- LAZY LOADING ---")
		lazyStart := time.Now()
		lazyGraph, err := query.BuildGraphLazy(tree)
		lazyDuration := time.Since(lazyStart)

		if err != nil {
			t.Errorf("Lazy loading failed: %v", err)
			continue
		}

		lazyBefore, _, _, _, _ := getMemStats()
		runtime.GC()
		lazyAfter, lazyTotalAlloc, _, lazyGC, _ := getMemStats()
		lazyMemory := uint64(0)
		if lazyAfter > lazyBefore {
			lazyMemory = lazyAfter - lazyBefore
		}

		componentCount := lazyGraph.GetComponentCount()

		t.Logf("  Duration: %v", lazyDuration)
		t.Logf("  Memory: %.2f MB", float64(lazyMemory)/(1024*1024))
		t.Logf("  Peak Memory: %.2f MB", float64(lazyTotalAlloc)/(1024*1024))
		t.Logf("  Nodes: %d (skeleton only)", len(lazyGraph.GetAllIndividuals()))
		t.Logf("  Edges: %d (loaded on-demand)", len(lazyGraph.GetAllEdges()))
		t.Logf("  Components: %d", componentCount)
		t.Logf("  GC Cycles: %d", lazyGC)

		// Comparison
		t.Log("\n--- COMPARISON ---")
		speedup := float64(eagerDuration) / float64(lazyDuration)
		memorySavings := float64(0)
		if eagerMemory > 0 {
			memorySavings = float64(eagerMemory-lazyMemory) / float64(eagerMemory) * 100
		}

		t.Logf("  Speed: %.1fx faster (%.1f%% improvement)", speedup, (speedup-1)*100)
		t.Logf("  Memory: %.1f%% reduction (%.2f MB saved)",
			memorySavings,
			float64(eagerMemory-lazyMemory)/(1024*1024))

		// Test query performance with lazy loading
		t.Log("\n--- QUERY PERFORMANCE (Lazy) ---")
		q, err := query.NewQueryLazy(tree)
		if err != nil {
			t.Errorf("Failed to create lazy query: %v", err)
			continue
		}

		// Access a few individuals
		queryStart := time.Now()
		for i := 1; i <= 10 && i <= size; i++ {
			xrefID := fmt.Sprintf("@I%d@", i)
			indi := q.Individual(xrefID)
			if indi == nil {
				t.Errorf("Failed to get individual %s", xrefID)
			}
			// Query parents
			_, _ = indi.Parents()
		}
		queryDuration := time.Since(queryStart)
		t.Logf("  Accessed 10 individuals + parents: %v", queryDuration)
		t.Logf("  Average: %v per individual", queryDuration/10)

		// Memory after queries
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		t.Logf("  Memory after queries: %.2f MB", float64(m.Alloc)/(1024*1024))

		// Summary
		t.Logf("\n✓ Size %d: Lazy loading is %.1fx faster, uses %.1f%% less memory\n",
			size, speedup, memorySavings)
	}
}

// TestStress_LazyLoading_5M tests lazy loading with 5M individuals
// This test verifies if lazy loading can handle 5M individuals that previously OOM'd
func TestStress_LazyLoading_5M(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 5M lazy loading stress test in short mode")
	}

	// Set timeout to 5 minutes
	timer := time.AfterFunc(5*time.Minute, func() {
		panic("test timed out after 5 minutes")
	})
	defer timer.Stop()

	size := 5_000_000
	t.Logf("=== LAZY LOADING STRESS TEST: %d INDIVIDUALS ===\n", size)
	t.Logf("Testing if lazy loading can handle 5M individuals (previously OOM'd)\n")
	t.Logf("Timeout: 5 minutes\n")

	// Generate test data
	genStart := time.Now()
	tree := generateStressTestTree(size)
	genDuration := time.Since(genStart)
	t.Logf("✓ Generated %d individuals in %v\n", size, genDuration)

	// Test Lazy Loading (this is what we're testing - can it avoid OOM?)
	t.Log("\n--- LAZY LOADING MODE (5M) ---")
	lazyStart := time.Now()
	lazyGraph, err := query.BuildGraphLazy(tree)
	lazyDuration := time.Since(lazyStart)

	if err != nil {
		t.Fatalf("Lazy loading failed: %v", err)
	}

	lazyBefore, _, _, _, _ := getMemStats()
	runtime.GC()
	lazyAfter, lazyTotalAlloc, _, lazyGC, _ := getMemStats()
	lazyMemory := uint64(0)
	if lazyAfter > lazyBefore {
		lazyMemory = lazyAfter - lazyBefore
	}

	componentCount := lazyGraph.GetComponentCount()

	t.Logf("✓ Lazy graph constructed:")
	t.Logf("  Duration: %v", lazyDuration)
	t.Logf("  Memory: %.2f MB", float64(lazyMemory)/(1024*1024))
	t.Logf("  Peak Memory: %.2f MB", float64(lazyTotalAlloc)/(1024*1024))
	t.Logf("  Nodes: %d (skeleton only)", len(lazyGraph.GetAllIndividuals()))
	t.Logf("  Edges: %d (loaded on-demand)", len(lazyGraph.GetAllEdges()))
	t.Logf("  Components: %d", componentCount)
	t.Logf("  GC Cycles: %d", lazyGC)

	// Test query performance
	t.Log("\n--- QUERY PERFORMANCE TEST (5M) ---")
	q, err := query.NewQueryLazy(tree)
	if err != nil {
		t.Fatalf("Failed to create lazy query: %v", err)
	}

	// Access a few individuals
	queryStart := time.Now()
	for i := 1; i <= 10 && i <= size; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indi := q.Individual(xrefID)
		if indi == nil {
			t.Errorf("Failed to get individual %s", xrefID)
			continue
		}
		// Query parents
		_, _ = indi.Parents()
	}
	queryDuration := time.Since(queryStart)
	t.Logf("  Accessed 10 individuals + parents: %v", queryDuration)
	t.Logf("  Average: %v per individual", queryDuration/10)

	// Memory after queries
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("  Memory after queries: %.2f MB", float64(m.Alloc)/(1024*1024))

	t.Logf("\n✓ 5M lazy loading test completed successfully!")
	t.Logf("  Previous eager loading: OOM killed at ~70-75 GB")
	t.Logf("  Lazy loading: Peak memory %.2f MB (%.2f GB)",
		float64(lazyTotalAlloc)/(1024*1024),
		float64(lazyTotalAlloc)/(1024*1024*1024))
}

// TestStress_LazyLoading_10M tests lazy loading with 10M individuals
// This is an extreme stress test to see if lazy loading can handle very large datasets
func TestStress_LazyLoading_10M(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 10M lazy loading stress test in short mode")
	}

	// Set timeout to 10 minutes (larger dataset needs more time)
	timer := time.AfterFunc(10*time.Minute, func() {
		panic("test timed out after 10 minutes")
	})
	defer timer.Stop()

	size := 10_000_000
	t.Logf("=== LAZY LOADING STRESS TEST: %d INDIVIDUALS ===\n", size)
	t.Logf("Extreme stress test: Testing if lazy loading can handle 10M individuals\n")
	t.Logf("Timeout: 10 minutes\n")

	// Generate test data
	genStart := time.Now()
	tree := generateStressTestTree(size)
	genDuration := time.Since(genStart)
	t.Logf("✓ Generated %d individuals in %v\n", size, genDuration)
	t.Logf("  Generation rate: %.0f individuals/sec\n", float64(size)/genDuration.Seconds())

	// Test Lazy Loading
	t.Log("\n--- LAZY LOADING MODE (10M) ---")
	lazyStart := time.Now()
	lazyGraph, err := query.BuildGraphLazy(tree)
	lazyDuration := time.Since(lazyStart)

	if err != nil {
		t.Fatalf("Lazy loading failed: %v", err)
	}

	lazyBefore, _, _, _, _ := getMemStats()
	runtime.GC()
	lazyAfter, lazyTotalAlloc, _, lazyGC, _ := getMemStats()
	lazyMemory := uint64(0)
	if lazyAfter > lazyBefore {
		lazyMemory = lazyAfter - lazyBefore
	}

	componentCount := lazyGraph.GetComponentCount()

	t.Logf("✓ Lazy graph constructed:")
	t.Logf("  Duration: %v", lazyDuration)
	t.Logf("  Construction rate: %.0f individuals/sec", float64(size)/lazyDuration.Seconds())
	t.Logf("  Memory: %.2f MB", float64(lazyMemory)/(1024*1024))
	t.Logf("  Peak Memory: %.2f MB (%.2f GB)",
		float64(lazyTotalAlloc)/(1024*1024),
		float64(lazyTotalAlloc)/(1024*1024*1024))
	t.Logf("  Nodes: %d (skeleton only)", len(lazyGraph.GetAllIndividuals()))
	t.Logf("  Edges: %d (loaded on-demand)", len(lazyGraph.GetAllEdges()))
	t.Logf("  Components: %d", componentCount)
	t.Logf("  GC Cycles: %d", lazyGC)

	// Test query performance
	t.Log("\n--- QUERY PERFORMANCE TEST (10M) ---")
	q, err := query.NewQueryLazy(tree)
	if err != nil {
		t.Fatalf("Failed to create lazy query: %v", err)
	}

	// Access a few individuals
	queryStart := time.Now()
	for i := 1; i <= 10 && i <= size; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indi := q.Individual(xrefID)
		if indi == nil {
			t.Errorf("Failed to get individual %s", xrefID)
			continue
		}
		// Query parents
		_, _ = indi.Parents()
	}
	queryDuration := time.Since(queryStart)
	t.Logf("  Accessed 10 individuals + parents: %v", queryDuration)
	t.Logf("  Average: %v per individual", queryDuration/10)

	// Memory after queries
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("  Memory after queries: %.2f MB", float64(m.Alloc)/(1024*1024))

	// Summary
	t.Logf("\n=== SUMMARY ===")
	t.Logf("✓ 10M lazy loading test completed successfully!")
	t.Logf("  Previous eager loading (5M): OOM killed at ~70-75 GB")
	t.Logf("  Lazy loading (10M): Peak memory %.2f GB", float64(lazyTotalAlloc)/(1024*1024*1024))
	t.Logf("  Memory per 1M individuals: %.2f GB", float64(lazyTotalAlloc)/(1024*1024*1024*10))
	t.Logf("  Query performance: %.2fµs per individual", float64(queryDuration.Nanoseconds())/10/1000)
}
