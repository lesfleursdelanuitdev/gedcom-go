package parser

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/validator"
)

// ParseMetrics holds parser performance metrics
type ParseMetrics struct {
	Filename           string
	FileSize           int64
	NumLines           int
	NumIndividuals     int
	NumFamilies        int
	NumOtherRecords    int
	ParseDuration      time.Duration
	ValidationDuration time.Duration
	GraphBuildDuration time.Duration
	TotalDuration      time.Duration
	MemoryBefore       uint64
	MemoryAfterParse   uint64
	MemoryAfterValid   uint64
	MemoryAfterGraph   uint64
	MemoryPeak         uint64
	ParseThroughput    float64 // individuals/sec
	ValidThroughput    float64 // records/sec
	ParseErrors        int
	ValidationErrors   int
	ValidationWarnings int
}

// measureMemoryTestData returns current memory usage in bytes
func measureMemoryTestData() uint64 {
	var m runtime.MemStats
	runtime.GC() // Force GC before measurement
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// getPeakMemoryTestData returns peak memory usage
func getPeakMemoryTestData() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.TotalAlloc
}

// measureParsePerformance measures parsing performance for a test data file
func measureParsePerformance(filename string) (*ParseMetrics, error) {
	filePath := findTestDataFile(filename)
	if filePath == "" {
		return nil, fmt.Errorf("test data file not found: %s", filename)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Count lines (approximate)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Simple line count
	numLines := 0
	buf := make([]byte, 32*1024)
	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			if buf[i] == '\n' {
				numLines++
			}
		}
		if err != nil {
			break
		}
	}

	metrics := &ParseMetrics{
		Filename: filename,
		FileSize: fileInfo.Size(),
		NumLines: numLines,
	}

	// Measure memory before
	metrics.MemoryBefore = measureMemoryTestData()
	var peakBefore uint64
	{
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		peakBefore = m.TotalAlloc
	}

	// Phase 1: Parse
	parseStart := time.Now()
	parser := NewHierarchicalParser()
	tree, err := parser.Parse(filePath)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}
	metrics.ParseDuration = time.Since(parseStart)
	metrics.MemoryAfterParse = measureMemoryTestData()

	// Count records
	individuals := tree.GetAllIndividuals()
	families := tree.GetAllFamilies()
	notes := tree.GetAllNotes()
	sources := tree.GetAllSources()
	repositories := tree.GetAllRepositories()
	submitters := tree.GetAllSubmitters()
	multimedia := tree.GetAllMultimedia()

	metrics.NumIndividuals = len(individuals)
	metrics.NumFamilies = len(families)
	metrics.NumOtherRecords = len(notes) + len(sources) + len(repositories) + len(submitters) + len(multimedia)

	// Calculate parse throughput
	if metrics.ParseDuration > 0 {
		metrics.ParseThroughput = float64(metrics.NumIndividuals) / metrics.ParseDuration.Seconds()
	}

	// Count parse errors
	errors := parser.GetErrors()
	metrics.ParseErrors = len(errors)

	// Phase 2: Basic Validation
	validStart := time.Now()
	errorManager := types.NewErrorManager()
	gedcomValidator := validator.NewGedcomValidator(errorManager)
	err = gedcomValidator.Validate(tree)
	metrics.ValidationDuration = time.Since(validStart)
	metrics.MemoryAfterValid = measureMemoryTestData()

	// Count validation errors/warnings
	allErrors := errorManager.Errors()
	metrics.ValidationErrors = 0
	metrics.ValidationWarnings = 0
	for _, e := range allErrors {
		if e.Severity == types.SeveritySevere {
			metrics.ValidationErrors++
		} else if e.Severity == types.SeverityWarning {
			metrics.ValidationWarnings++
		}
	}

	// Calculate validation throughput
	totalRecords := metrics.NumIndividuals + metrics.NumFamilies + metrics.NumOtherRecords
	if metrics.ValidationDuration > 0 {
		metrics.ValidThroughput = float64(totalRecords) / metrics.ValidationDuration.Seconds()
	}

	// Phase 3: Graph Construction
	graphStart := time.Now()
	graph, err := query.BuildGraph(tree)
	if err != nil {
		return nil, fmt.Errorf("graph construction failed: %w", err)
	}
	metrics.GraphBuildDuration = time.Since(graphStart)
	metrics.MemoryAfterGraph = measureMemoryTestData()

	// Get peak memory
	metrics.MemoryPeak = getPeakMemoryTestData()
	if metrics.MemoryPeak < peakBefore {
		metrics.MemoryPeak = peakBefore
	}

	// Calculate total duration
	metrics.TotalDuration = metrics.ParseDuration + metrics.ValidationDuration + metrics.GraphBuildDuration

	// Keep graph in scope to prevent GC
	_ = graph

	return metrics, nil
}

// printParsePerformanceReport prints a formatted performance report
func printParsePerformanceReport(metrics *ParseMetrics) {
	fmt.Printf("\n%s (%d KB, %d lines)\n", metrics.Filename, metrics.FileSize/1024, metrics.NumLines)
	fmt.Printf("  Records: %d individuals, %d families, %d other\n", metrics.NumIndividuals, metrics.NumFamilies, metrics.NumOtherRecords)
	fmt.Printf("  Parse:            %8.2fms (%8.0f individuals/sec)\n", 
		metrics.ParseDuration.Seconds()*1000, metrics.ParseThroughput)
	fmt.Printf("  Basic Validation: %8.2fms (%8.0f records/sec)\n", 
		metrics.ValidationDuration.Seconds()*1000, metrics.ValidThroughput)
	fmt.Printf("  Graph Build:      %8.2fms\n", 
		metrics.GraphBuildDuration.Seconds()*1000)
	fmt.Printf("  Total:            %8.2fms\n", 
		metrics.TotalDuration.Seconds()*1000)
	// Calculate memory used (be careful with uint64 subtraction)
	memoryUsed := uint64(0)
	if metrics.MemoryAfterGraph > metrics.MemoryBefore {
		memoryUsed = metrics.MemoryAfterGraph - metrics.MemoryBefore
	}
	fmt.Printf("  Memory:           %8.2f MB (peak: %8.2f MB)\n", 
		float64(memoryUsed)/1024/1024,
		float64(metrics.MemoryPeak)/1024/1024)
	fmt.Printf("  Errors:           %d parse, %d validation (%d warnings)\n", 
		metrics.ParseErrors, metrics.ValidationErrors, metrics.ValidationWarnings)
}

// TestParserPerformance_AllTestDataFiles tests parser performance on all test data files
func TestParserPerformance_AllTestDataFiles(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
		"tree1.ged",
		"royal92.ged",
	}

	// Only test pres2020.ged if not in short mode (it's very large)
	if !testing.Short() {
		testFiles = append(testFiles, "pres2020.ged")
	}

	fmt.Printf("\n=== Parser Performance Results ===\n")
	fmt.Printf("===================================\n\n")

	allMetrics := make([]*ParseMetrics, 0, len(testFiles))

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			metrics, err := measureParsePerformance(filename)
			if err != nil {
				if err.Error() == fmt.Sprintf("test data file not found: %s", filename) {
					t.Skipf("Test data file not found: %s", filename)
					return
				}
				t.Fatalf("Failed to measure performance: %v", err)
			}

			allMetrics = append(allMetrics, metrics)
			printParsePerformanceReport(metrics)

			// Basic assertions
			if metrics.NumIndividuals == 0 {
				t.Errorf("Expected at least 1 individual, got 0")
			}
			if metrics.ParseDuration <= 0 {
				t.Errorf("Parse duration should be positive, got %v", metrics.ParseDuration)
			}
		})
	}

	// Print summary
	if len(allMetrics) > 0 {
		fmt.Printf("\n=== Summary ===\n")
		fmt.Printf("Files tested: %d\n", len(allMetrics))
		
		totalIndividuals := 0
		totalDuration := time.Duration(0)
		for _, m := range allMetrics {
			totalIndividuals += m.NumIndividuals
			totalDuration += m.TotalDuration
		}
		
		if totalDuration > 0 {
			avgThroughput := float64(totalIndividuals) / totalDuration.Seconds()
			fmt.Printf("Total individuals: %d\n", totalIndividuals)
			fmt.Printf("Total time: %v\n", totalDuration)
			fmt.Printf("Average throughput: %.0f individuals/sec\n", avgThroughput)
		}
	}
}

// TestParserPerformance_IndividualFiles tests individual files separately
func TestParserPerformance_xavier(t *testing.T) {
	metrics, err := measureParsePerformance("xavier.ged")
	if err != nil {
		if err.Error() == "test data file not found: xavier.ged" {
			t.Skip("Test data file not found")
			return
		}
		t.Fatalf("Failed: %v", err)
	}
	printParsePerformanceReport(metrics)
}

func TestParserPerformance_gracis(t *testing.T) {
	metrics, err := measureParsePerformance("gracis.ged")
	if err != nil {
		if err.Error() == "test data file not found: gracis.ged" {
			t.Skip("Test data file not found")
			return
		}
		t.Fatalf("Failed: %v", err)
	}
	printParsePerformanceReport(metrics)
}

func TestParserPerformance_tree1(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	metrics, err := measureParsePerformance("tree1.ged")
	if err != nil {
		if err.Error() == "test data file not found: tree1.ged" {
			t.Skip("Test data file not found")
			return
		}
		t.Fatalf("Failed: %v", err)
	}
	printParsePerformanceReport(metrics)
}

func TestParserPerformance_royal92(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	metrics, err := measureParsePerformance("royal92.ged")
	if err != nil {
		if err.Error() == "test data file not found: royal92.ged" {
			t.Skip("Test data file not found")
			return
		}
		t.Fatalf("Failed: %v", err)
	}
	printParsePerformanceReport(metrics)
}

func TestParserPerformance_pres2020(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	metrics, err := measureParsePerformance("pres2020.ged")
	if err != nil {
		if err.Error() == "test data file not found: pres2020.ged" {
			t.Skip("Test data file not found")
			return
		}
		t.Fatalf("Failed: %v", err)
	}
	printParsePerformanceReport(metrics)
}

