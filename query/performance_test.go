package query

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// generateVeryLargeTree creates a tree with n individuals efficiently
// Uses a more realistic family structure with varying family sizes
func generateVeryLargeTree(n int) *types.GedcomTree {
	tree := types.NewGedcomTree()

	// Pre-allocate slices for better performance
	individuals := make([]*types.IndividualRecord, 0, n)
	families := make([]*types.FamilyRecord, 0, n/2)

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
		famLine := types.NewGedcomLine(0, "FAM", "", fmt.Sprintf("@F%d@", familyID))

		// Husband
		if indiID < n {
			famLine.AddChild(types.NewGedcomLine(1, "HUSB", fmt.Sprintf("@I%d@", indiID), ""))
			indiID++
		}

		// Wife
		if indiID < n {
			famLine.AddChild(types.NewGedcomLine(1, "WIFE", fmt.Sprintf("@I%d@", indiID), ""))
			indiID++
		}

		// Children
		for i := 0; i < numChildren && indiID < n; i++ {
			famLine.AddChild(types.NewGedcomLine(1, "CHIL", fmt.Sprintf("@I%d@", indiID), ""))
			childToFamily[indiID] = familyID
			indiID++
		}

		fam := types.NewFamilyRecord(famLine)
		families = append(families, fam)
		familyID++
	}

	// Create all individuals
	for i := 1; i <= n; i++ {
		indiLine := types.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i))

		// Add name
		indiLine.AddChild(types.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", i), ""))

		// Add birth date (distributed across years 1800-2000)
		birthYear := 1800 + (i % 200)
		birtLine := types.NewGedcomLine(1, "BIRT", "", "")
		birtLine.AddChild(types.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", birthYear), ""))
		indiLine.AddChild(birtLine)

		// Add sex (alternating)
		sex := "M"
		if i%2 == 0 {
			sex = "F"
		}
		indiLine.AddChild(types.NewGedcomLine(1, "SEX", sex, ""))

		// Add FAMC if child
		if famID, ok := childToFamily[i]; ok {
			indiLine.AddChild(types.NewGedcomLine(1, "FAMC", fmt.Sprintf("@F%d@", famID), ""))
		}

		indi := types.NewIndividualRecord(indiLine)
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

// PerformanceMetrics holds performance test results
type PerformanceMetrics struct {
	Operation    string
	DatasetSize  int
	Duration     time.Duration
	MemoryBefore uint64
	MemoryAfter  uint64
	MemoryPeak   uint64
	Throughput   float64 // operations per second
}

// measureMemory returns current memory usage in bytes
func measureMemory() uint64 {
	var m runtime.MemStats
	runtime.GC() // Force GC before measurement
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// runPerformanceTest runs a test function and collects metrics
func runPerformanceTest(operation string, datasetSize int, fn func()) PerformanceMetrics {
	before := measureMemory()
	start := time.Now()

	fn()

	duration := time.Since(start)
	after := measureMemory()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	peak := m.TotalAlloc

	throughput := float64(datasetSize) / duration.Seconds()

	return PerformanceMetrics{
		Operation:    operation,
		DatasetSize:  datasetSize,
		Duration:     duration,
		MemoryBefore: before,
		MemoryAfter:  after,
		MemoryPeak:   peak,
		Throughput:   throughput,
	}
}

// TestPerformance_100K runs comprehensive performance tests with 100,000 individuals
func TestPerformance_100K(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large performance test in short mode")
	}
	// Skip by default in CI - only run if ENABLE_PERFORMANCE_TESTS is set
	if os.Getenv("ENABLE_PERFORMANCE_TESTS") == "" {
		t.Skip("Skipping performance test - set ENABLE_PERFORMANCE_TESTS=1 to run")
	}

	const size = 100000
	t.Logf("Generating tree with %d individuals...", size)
	tree := generateVeryLargeTree(size)

	metrics := []PerformanceMetrics{}

	// Test 1: Graph Construction
	t.Log("Testing graph construction...")
	metrics = append(metrics, runPerformanceTest("GraphConstruction", size, func() {
		_, err := BuildGraph(tree)
		if err != nil {
			t.Fatalf("Graph construction failed: %v", err)
		}
	}))

	// Build graph once for subsequent tests
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Graph construction failed: %v", err)
	}

	// Test 2: Query Creation
	t.Log("Testing query creation...")
	metrics = append(metrics, runPerformanceTest("QueryCreation", size, func() {
		_ = NewQueryFromGraph(graph)
	}))

	query := NewQueryFromGraph(graph)

	// Test 3: Filter Query (indexed)
	t.Log("Testing filter query (indexed)...")
	metrics = append(metrics, runPerformanceTest("FilterQuery_ByName", size, func() {
		results, err := query.Filter().ByName("Person").Execute()
		if err != nil {
			t.Fatalf("Filter query failed: %v", err)
		}
		if len(results) == 0 {
			t.Log("Warning: No results found")
		}
	}))

	// Test 4: Filter Query (exact)
	t.Log("Testing filter query (exact)...")
	metrics = append(metrics, runPerformanceTest("FilterQuery_ByNameExact", size, func() {
		results, err := query.Filter().ByNameExact("Person 50000 /Test/").Execute()
		if err != nil {
			t.Fatalf("Filter query failed: %v", err)
		}
		if len(results) == 0 {
			t.Log("Warning: No results found")
		}
	}))

	// Test 5: Filter Query (by sex)
	t.Log("Testing filter query (by sex)...")
	metrics = append(metrics, runPerformanceTest("FilterQuery_BySex", size, func() {
		results, err := query.Filter().BySex("M").Execute()
		if err != nil {
			t.Fatalf("Filter query failed: %v", err)
		}
		if len(results) == 0 {
			t.Log("Warning: No results found")
		}
	}))

	// Test 6: Ancestor Query
	t.Log("Testing ancestor query...")
	metrics = append(metrics, runPerformanceTest("AncestorQuery", size, func() {
		results, err := query.Individual("@I50000@").Ancestors().MaxGenerations(5).Execute()
		if err != nil {
			t.Fatalf("Ancestor query failed: %v", err)
		}
		if len(results) == 0 {
			t.Log("Warning: No ancestors found")
		}
	}))

	// Test 7: Path Finding
	t.Log("Testing path finding...")
	metrics = append(metrics, runPerformanceTest("PathFinding", size, func() {
		_, err := query.Individual("@I1@").PathTo("@I50000@").Shortest()
		if err != nil {
			// Path might not exist, that's okay
			t.Logf("Path finding returned error (expected): %v", err)
		}
	}))

	// Print results
	t.Log("\n=== Performance Test Results (100K) ===")
	for _, m := range metrics {
		t.Logf("%s:", m.Operation)
		t.Logf("  Duration: %v", m.Duration)
		t.Logf("  Throughput: %.2f ops/sec", m.Throughput)
		t.Logf("  Memory Used: %.2f MB", float64(m.MemoryAfter-m.MemoryBefore)/1024/1024)
		t.Logf("  Peak Memory: %.2f MB", float64(m.MemoryPeak)/1024/1024)
	}
}

// TestPerformance_500K runs comprehensive performance tests with 500,000 individuals
func TestPerformance_500K(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large performance test in short mode")
	}
	// Skip by default in CI - only run if ENABLE_PERFORMANCE_TESTS is set
	if os.Getenv("ENABLE_PERFORMANCE_TESTS") == "" {
		t.Skip("Skipping performance test - set ENABLE_PERFORMANCE_TESTS=1 to run")
	}

	const size = 500000
	t.Logf("Generating tree with %d individuals...", size)
	tree := generateVeryLargeTree(size)

	metrics := []PerformanceMetrics{}

	// Test 1: Graph Construction
	t.Log("Testing graph construction...")
	metrics = append(metrics, runPerformanceTest("GraphConstruction", size, func() {
		_, err := BuildGraph(tree)
		if err != nil {
			t.Fatalf("Graph construction failed: %v", err)
		}
	}))

	// Build graph once for subsequent tests
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Graph construction failed: %v", err)
	}

	// Test 2: Query Creation
	t.Log("Testing query creation...")
	metrics = append(metrics, runPerformanceTest("QueryCreation", size, func() {
		_ = NewQueryFromGraph(graph)
	}))

	query := NewQueryFromGraph(graph)

	// Test 3: Filter Query (indexed)
	t.Log("Testing filter query (indexed)...")
	metrics = append(metrics, runPerformanceTest("FilterQuery_ByName", size, func() {
		results, err := query.Filter().ByName("Person").Execute()
		if err != nil {
			t.Fatalf("Filter query failed: %v", err)
		}
		if len(results) == 0 {
			t.Log("Warning: No results found")
		}
	}))

	// Test 4: Filter Query (by sex)
	t.Log("Testing filter query (by sex)...")
	metrics = append(metrics, runPerformanceTest("FilterQuery_BySex", size, func() {
		results, err := query.Filter().BySex("M").Execute()
		if err != nil {
			t.Fatalf("Filter query failed: %v", err)
		}
		if len(results) == 0 {
			t.Log("Warning: No results found")
		}
	}))

	// Print results
	t.Log("\n=== Performance Test Results (500K) ===")
	for _, m := range metrics {
		t.Logf("%s:", m.Operation)
		t.Logf("  Duration: %v", m.Duration)
		t.Logf("  Throughput: %.2f ops/sec", m.Throughput)
		t.Logf("  Memory Used: %.2f MB", float64(m.MemoryAfter-m.MemoryBefore)/1024/1024)
		t.Logf("  Peak Memory: %.2f MB", float64(m.MemoryPeak)/1024/1024)
	}
}

// BenchmarkGraphConstruction_100K benchmarks graph construction with 100K individuals
func BenchmarkGraphConstruction_100K(b *testing.B) {
	tree := generateVeryLargeTree(100000)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := BuildGraph(tree)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGraphConstruction_500K benchmarks graph construction with 500K individuals
func BenchmarkGraphConstruction_500K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	tree := generateVeryLargeTree(500000)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := BuildGraph(tree)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFilterQuery_100K benchmarks filter queries with 100K individuals
func BenchmarkFilterQuery_100K(b *testing.B) {
	tree := generateVeryLargeTree(100000)
	graph, err := BuildGraph(tree)
	if err != nil {
		b.Fatal(err)
	}

	query := NewQueryFromGraph(graph)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := query.Filter().ByName("Person").Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFilterQuery_500K benchmarks filter queries with 500K individuals
func BenchmarkFilterQuery_500K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	tree := generateVeryLargeTree(500000)
	graph, err := BuildGraph(tree)
	if err != nil {
		b.Fatal(err)
	}

	query := NewQueryFromGraph(graph)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := query.Filter().ByName("Person").Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}
