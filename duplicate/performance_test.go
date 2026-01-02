package duplicate

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
)

// generateLargeTreeForDuplicate creates a tree with n individuals for duplicate testing
func generateLargeTreeForDuplicate(n int) *types.GedcomTree {
	tree := types.NewGedcomTree()

	// Create individuals with some intentional duplicates
	for i := 1; i <= n; i++ {
		indiLine := types.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i))

		// Create some duplicates (every 100th person is similar to another)
		name := fmt.Sprintf("Person %d /Test/", i)
		if i%100 == 0 && i > 100 {
			// Make similar to previous person
			name = fmt.Sprintf("Person %d /Test/", i-1)
		}

		indiLine.AddChild(types.NewGedcomLine(1, "NAME", name, ""))

		birthYear := 1800 + (i % 200)
		birtLine := types.NewGedcomLine(1, "BIRT", "", "")
		birtLine.AddChild(types.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", birthYear), ""))
		indiLine.AddChild(birtLine)

		sex := "M"
		if i%2 == 0 {
			sex = "F"
		}
		indiLine.AddChild(types.NewGedcomLine(1, "SEX", sex, ""))

		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	return tree
}

// measureMemory returns current memory usage
func measureMemory() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// TestPerformance_DuplicateDetection_100K tests duplicate detection with 100K individuals
func TestPerformance_DuplicateDetection_100K(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large performance test in short mode")
	}
	// Skip by default in CI - only run if ENABLE_PERFORMANCE_TESTS is set
	if os.Getenv("ENABLE_PERFORMANCE_TESTS") == "" {
		t.Skip("Skipping performance test - set ENABLE_PERFORMANCE_TESTS=1 to run")
	}

	const size = 100000
	t.Logf("Generating tree with %d individuals...", size)
	tree := generateLargeTreeForDuplicate(size)

	// Build graph for relationship matching
	graph, err := query.BuildGraph(tree)
	if err != nil {
		t.Fatalf("Graph construction failed: %v", err)
	}

	before := measureMemory()
	start := time.Now()

	detector := NewDuplicateDetector(DefaultConfig())
	detector.SetTree(tree)
	result, err := detector.FindDuplicates(tree)

	duration := time.Since(start)
	after := measureMemory()

	if err != nil {
		t.Fatalf("Duplicate detection failed: %v", err)
	}

	t.Logf("\n=== Duplicate Detection Performance (100K) ===")
	t.Logf("Duration: %v", duration)
	t.Logf("Throughput: %.2f comparisons/sec", float64(result.TotalComparisons)/duration.Seconds())
	t.Logf("Memory Used: %.2f MB", float64(after-before)/1024/1024)
	t.Logf("Matches Found: %d", len(result.Matches))
	t.Logf("Total Comparisons: %d", result.TotalComparisons)

	if result.Metrics != nil {
		t.Logf("Parallel Workers: %d", result.Metrics.ParallelWorkers)
		t.Logf("Processing Time: %v", result.Metrics.ProcessingTime)
	}

	_ = graph // Use graph to avoid unused variable
}

// TestPerformance_DuplicateDetection_500K tests duplicate detection with 500K individuals
func TestPerformance_DuplicateDetection_500K(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large performance test in short mode")
	}
	// Skip by default in CI - only run if ENABLE_PERFORMANCE_TESTS is set
	if os.Getenv("ENABLE_PERFORMANCE_TESTS") == "" {
		t.Skip("Skipping performance test - set ENABLE_PERFORMANCE_TESTS=1 to run")
	}

	const size = 500000
	t.Logf("Generating tree with %d individuals...", size)
	tree := generateLargeTreeForDuplicate(size)

	// Build graph for relationship matching
	graph, err := query.BuildGraph(tree)
	if err != nil {
		t.Fatalf("Graph construction failed: %v", err)
	}

	before := measureMemory()
	start := time.Now()

	config := DefaultConfig()
	config.UseParallelProcessing = true
	config.NumWorkers = 0 // Auto-detect

	detector := NewDuplicateDetector(config)
	detector.SetTree(tree)
	result, err := detector.FindDuplicates(tree)

	duration := time.Since(start)
	after := measureMemory()

	if err != nil {
		t.Fatalf("Duplicate detection failed: %v", err)
	}

	t.Logf("\n=== Duplicate Detection Performance (500K) ===")
	t.Logf("Duration: %v", duration)
	t.Logf("Throughput: %.2f comparisons/sec", float64(result.TotalComparisons)/duration.Seconds())
	t.Logf("Memory Used: %.2f MB", float64(after-before)/1024/1024)
	t.Logf("Matches Found: %d", len(result.Matches))
	t.Logf("Total Comparisons: %d", result.TotalComparisons)

	if result.Metrics != nil {
		t.Logf("Parallel Workers: %d", result.Metrics.ParallelWorkers)
		t.Logf("Processing Time: %v", result.Metrics.ProcessingTime)
	}

	_ = graph // Use graph to avoid unused variable
}

// BenchmarkDuplicateDetection_100K benchmarks duplicate detection with 100K individuals
func BenchmarkDuplicateDetection_100K(b *testing.B) {
	tree := generateLargeTreeForDuplicate(100000)
	graph, err := query.BuildGraph(tree)
	if err != nil {
		b.Fatal(err)
	}

	detector := NewDuplicateDetector(DefaultConfig())
	detector.SetTree(tree)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := detector.FindDuplicates(tree)
		if err != nil {
			b.Fatal(err)
		}
	}

	_ = graph
}

// BenchmarkDuplicateDetection_500K benchmarks duplicate detection with 500K individuals
func BenchmarkDuplicateDetection_500K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	tree := generateLargeTreeForDuplicate(500000)
	graph, err := query.BuildGraph(tree)
	if err != nil {
		b.Fatal(err)
	}

	config := DefaultConfig()
	config.UseParallelProcessing = true

	detector := NewDuplicateDetector(config)
	detector.SetTree(tree)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := detector.FindDuplicates(tree)
		if err != nil {
			b.Fatal(err)
		}
	}

	_ = graph
}
