package query

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// PerformanceBaseline represents baseline performance metrics
type PerformanceBaseline struct {
	Timestamp   time.Time              `json:"timestamp"`
	TestName    string                 `json:"test_name"`
	DatasetSize int                    `json:"dataset_size"`
	Metrics     map[string]interface{}  `json:"metrics"`
}

// PerformanceRegressionTest tracks performance over time
type PerformanceRegressionTest struct {
	baselineFile string
	threshold    float64 // Percentage threshold for regression (default 20%)
}

// NewPerformanceRegressionTest creates a new regression test tracker
func NewPerformanceRegressionTest(baselineFile string) *PerformanceRegressionTest {
	return &PerformanceRegressionTest{
		baselineFile: baselineFile,
		threshold:    20.0, // 20% slower is considered a regression
	}
}

// SetThreshold sets the regression threshold percentage
func (prt *PerformanceRegressionTest) SetThreshold(threshold float64) {
	prt.threshold = threshold
}

// LoadBaseline loads the baseline metrics from file
func (prt *PerformanceRegressionTest) LoadBaseline() (*PerformanceBaseline, error) {
	data, err := os.ReadFile(prt.baselineFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No baseline exists yet
		}
		return nil, fmt.Errorf("failed to read baseline file: %w", err)
	}

	var baseline PerformanceBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse baseline: %w", err)
	}

	return &baseline, nil
}

// SaveBaseline saves the current metrics as a new baseline
func (prt *PerformanceRegressionTest) SaveBaseline(baseline *PerformanceBaseline) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(prt.baselineFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	if err := os.WriteFile(prt.baselineFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write baseline file: %w", err)
	}

	return nil
}

// CompareMetrics compares current metrics with baseline
func (prt *PerformanceRegressionTest) CompareMetrics(baseline *PerformanceBaseline, current map[string]interface{}) (bool, []string) {
	if baseline == nil {
		return true, []string{"No baseline exists - this will become the new baseline"}
	}

	var regressions []string
	hasRegression := false

	// Compare each metric
	for key, baselineValue := range baseline.Metrics {
		currentValue, exists := current[key]
		if !exists {
			continue
		}

		// Convert to float64 for comparison
		baselineFloat := toFloat64(baselineValue)
		currentFloat := toFloat64(currentValue)

		if baselineFloat == 0 {
			continue
		}

		// Calculate percentage change
		percentChange := ((currentFloat - baselineFloat) / baselineFloat) * 100.0

		// Check if it's a regression (slower)
		if percentChange > prt.threshold {
			hasRegression = true
			regressions = append(regressions, fmt.Sprintf(
				"%s: %.2f%% slower (baseline: %v, current: %v)",
				key, percentChange, baselineValue, currentValue,
			))
		}
	}

	return !hasRegression, regressions
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case time.Duration:
		return float64(val.Nanoseconds())
	default:
		return 0
	}
}

// TestPerformanceRegression_GraphBuild tests graph build performance
func TestPerformanceRegression_GraphBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	baselineFile := "testdata/performance_baselines/graph_build.json"
	prt := NewPerformanceRegressionTest(baselineFile)

	// Test with 10K individuals
	datasetSize := 10000
	tree := generateVeryLargeTree(datasetSize)

	start := time.Now()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}
	duration := time.Since(start)

	// Get metrics
	currentMetrics := map[string]interface{}{
		"build_time_ms":     duration.Milliseconds(),
		"nodes":             graph.NodeCount(),
		"edges":             graph.EdgeCount(),
		"dataset_size":      datasetSize,
	}

	// Load baseline
	baseline, err := prt.LoadBaseline()
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Compare
	passed, regressions := prt.CompareMetrics(baseline, currentMetrics)
	if !passed {
		t.Errorf("Performance regression detected:\n%s", fmt.Sprintf("%v", regressions))
	}

	// Save new baseline if it doesn't exist or if we're updating
	if baseline == nil || os.Getenv("UPDATE_BASELINE") == "true" {
		newBaseline := &PerformanceBaseline{
			Timestamp:   time.Now(),
			TestName:    "graph_build",
			DatasetSize: datasetSize,
			Metrics:     currentMetrics,
		}
		if err := prt.SaveBaseline(newBaseline); err != nil {
			t.Logf("Failed to save baseline: %v", err)
		}
	}
}

// TestPerformanceRegression_QueryExecution tests query execution performance
func TestPerformanceRegression_QueryExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	baselineFile := "testdata/performance_baselines/query_execution.json"
	prt := NewPerformanceRegressionTest(baselineFile)

	// Build graph
	datasetSize := 5000
	tree := generateVeryLargeTree(datasetSize)
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Create query from graph
	q := NewQueryFromGraph(graph)

	// Run query
	start := time.Now()
	results, err := q.Filter().ByName("John").Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	duration := time.Since(start)

	// Get metrics
	currentMetrics := map[string]interface{}{
		"query_time_ms":    duration.Milliseconds(),
		"result_count":     len(results),
		"dataset_size":      datasetSize,
	}

	// Load baseline
	baseline, err := prt.LoadBaseline()
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Compare
	passed, regressions := prt.CompareMetrics(baseline, currentMetrics)
	if !passed {
		t.Errorf("Performance regression detected:\n%s", fmt.Sprintf("%v", regressions))
	}

	// Save new baseline if needed
	if baseline == nil || os.Getenv("UPDATE_BASELINE") == "true" {
		newBaseline := &PerformanceBaseline{
			Timestamp:   time.Now(),
			TestName:    "query_execution",
			DatasetSize: datasetSize,
			Metrics:     currentMetrics,
		}
		if err := prt.SaveBaseline(newBaseline); err != nil {
			t.Logf("Failed to save baseline: %v", err)
		}
	}
}

// TestPerformanceRegression_AncestorQuery tests ancestor query performance
func TestPerformanceRegression_AncestorQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	baselineFile := "testdata/performance_baselines/ancestor_query.json"
	prt := NewPerformanceRegressionTest(baselineFile)

	// Build graph
	datasetSize := 5000
	tree := generateVeryLargeTree(datasetSize)
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Create query from graph
	q := NewQueryFromGraph(graph)

	// Get first individual
	allIndi := tree.GetAllIndividuals()
	var firstXref string
	for xref := range allIndi {
		firstXref = xref
		break
	}

	// Run ancestor query
	start := time.Now()
	ancestors, err := q.Individual(firstXref).Ancestors().MaxGenerations(5).Execute()
	if err != nil {
		t.Fatalf("Failed to execute ancestor query: %v", err)
	}
	duration := time.Since(start)

	// Get metrics
	currentMetrics := map[string]interface{}{
		"query_time_ms":    duration.Milliseconds(),
		"ancestor_count":   len(ancestors),
		"dataset_size":     datasetSize,
	}

	// Load baseline
	baseline, err := prt.LoadBaseline()
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Compare
	passed, regressions := prt.CompareMetrics(baseline, currentMetrics)
	if !passed {
		t.Errorf("Performance regression detected:\n%s", fmt.Sprintf("%v", regressions))
	}

	// Save new baseline if needed
	if baseline == nil || os.Getenv("UPDATE_BASELINE") == "true" {
		newBaseline := &PerformanceBaseline{
			Timestamp:   time.Now(),
			TestName:    "ancestor_query",
			DatasetSize: datasetSize,
			Metrics:     currentMetrics,
		}
		if err := prt.SaveBaseline(newBaseline); err != nil {
			t.Logf("Failed to save baseline: %v", err)
		}
	}
}



// TestPerformanceRegression_DescendantQuery tests descendant query performance
func TestPerformanceRegression_DescendantQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	baselineFile := "testdata/performance_baselines/descendant_query.json"
	prt := NewPerformanceRegressionTest(baselineFile)

	// Build graph
	datasetSize := 5000
	tree := generateVeryLargeTree(datasetSize)
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Create query
	q := NewQueryFromGraph(graph)

	// Get first individual
	allIndi := tree.GetAllIndividuals()
	var firstXref string
	for xref := range allIndi {
		firstXref = xref
		break
	}

	// Run descendant query
	start := time.Now()
	descendants, err := q.Individual(firstXref).Descendants().MaxGenerations(5).Execute()
	if err != nil {
		t.Fatalf("Failed to execute descendant query: %v", err)
	}
	duration := time.Since(start)

	// Get metrics
	currentMetrics := map[string]interface{}{
		"query_time_ms":    duration.Milliseconds(),
		"descendant_count": len(descendants),
		"dataset_size":     datasetSize,
	}

	// Load baseline
	baseline, err := prt.LoadBaseline()
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Compare
	passed, regressions := prt.CompareMetrics(baseline, currentMetrics)
	if !passed {
		t.Errorf("Performance regression detected:\n%s", fmt.Sprintf("%v", regressions))
	}

	// Save new baseline if needed
	if baseline == nil || os.Getenv("UPDATE_BASELINE") == "true" {
		newBaseline := &PerformanceBaseline{
			Timestamp:   time.Now(),
			TestName:    "descendant_query",
			DatasetSize: datasetSize,
			Metrics:     currentMetrics,
		}
		if err := prt.SaveBaseline(newBaseline); err != nil {
			t.Logf("Failed to save baseline: %v", err)
		}
	}
}

// TestPerformanceRegression_RelationshipQuery tests relationship query performance
func TestPerformanceRegression_RelationshipQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	baselineFile := "testdata/performance_baselines/relationship_query.json"
	prt := NewPerformanceRegressionTest(baselineFile)

	// Build graph
	datasetSize := 5000
	tree := generateVeryLargeTree(datasetSize)
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Create query
	q := NewQueryFromGraph(graph)

	// Get two individuals
	allIndi := tree.GetAllIndividuals()
	var xref1, xref2 string
	count := 0
	for xref := range allIndi {
		if count == 0 {
			xref1 = xref
		} else if count == 10 {
			xref2 = xref
			break
		}
		count++
	}

	if xref1 == "" || xref2 == "" {
		t.Skip("Not enough individuals for relationship test")
	}

	// Run relationship query
	start := time.Now()
	result, err := q.Individual(xref1).RelationshipTo(xref2).Execute()
	if err != nil {
		t.Fatalf("Failed to execute relationship query: %v", err)
	}
	duration := time.Since(start)

	// Get metrics
	currentMetrics := map[string]interface{}{
		"query_time_ms":    duration.Milliseconds(),
		"relationship":     result.RelationshipType,
		"degree":           result.Degree,
		"dataset_size":     datasetSize,
	}

	// Load baseline
	baseline, err := prt.LoadBaseline()
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Compare
	passed, regressions := prt.CompareMetrics(baseline, currentMetrics)
	if !passed {
		t.Errorf("Performance regression detected:\n%s", fmt.Sprintf("%v", regressions))
	}

	// Save new baseline if needed
	if baseline == nil || os.Getenv("UPDATE_BASELINE") == "true" {
		newBaseline := &PerformanceBaseline{
			Timestamp:   time.Now(),
			TestName:    "relationship_query",
			DatasetSize: datasetSize,
			Metrics:     currentMetrics,
		}
		if err := prt.SaveBaseline(newBaseline); err != nil {
			t.Logf("Failed to save baseline: %v", err)
		}
	}
}
