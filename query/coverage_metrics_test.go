package query

import (
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestGraphMetrics_BetweennessCentrality tests betweenness centrality calculation
func TestGraphMetrics_BetweennessCentrality(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build graph
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test betweenness centrality
	metrics := graph.Metrics()
	centrality, err := metrics.Centrality(CentralityBetweenness)
	if err != nil {
		t.Fatalf("Failed to calculate betweenness centrality: %v", err)
	}

	if len(centrality) == 0 {
		t.Error("Expected at least one centrality value")
	}

	// Verify all values are non-negative
	for id, value := range centrality {
		if value < 0 {
			t.Errorf("Betweenness centrality for %s is negative: %f", id, value)
		}
	}
}

// TestGraphMetrics_ClosenessCentrality tests closeness centrality calculation
func TestGraphMetrics_ClosenessCentrality(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build graph
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test closeness centrality
	metrics := graph.Metrics()
	centrality, err := metrics.Centrality(CentralityCloseness)
	if err != nil {
		t.Fatalf("Failed to calculate closeness centrality: %v", err)
	}

	if len(centrality) == 0 {
		t.Error("Expected at least one centrality value")
	}

	// Verify all values are non-negative
	for id, value := range centrality {
		if value < 0 {
			t.Errorf("Closeness centrality for %s is negative: %f", id, value)
		}
	}
}

// TestGraphMetrics_Diameter tests diameter calculation
func TestGraphMetrics_Diameter(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build graph
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			graph, err := BuildGraph(tree)
			if err != nil {
				t.Fatalf("Failed to build graph: %v", err)
			}

			// Test diameter
			metrics := graph.Metrics()
			diameter, err := metrics.Diameter()
			if err != nil {
				t.Fatalf("Failed to calculate diameter: %v", err)
			}

			// Diameter should be non-negative
			if diameter < 0 {
				t.Errorf("Diameter is negative: %d", diameter)
			}
		})
	}
}

// TestGraphMetrics_Diameter_SingleNode tests diameter with single node
func TestGraphMetrics_Diameter_SingleNode(t *testing.T) {
	// Create a minimal tree with one individual
	tree := types.NewGedcomTree()
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "Test /Person/", "")
	indiLine.AddChild(nameLine)
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	metrics := graph.Metrics()
	diameter, err := metrics.Diameter()
	if err != nil {
		t.Fatalf("Failed to calculate diameter: %v", err)
	}

	// With less than 2 nodes, diameter should be 0
	if diameter != 0 {
		t.Errorf("Expected diameter 0 for single node, got %d", diameter)
	}
}

// TestGraphMetrics_OutDegree_Error tests OutDegree with invalid XREF
func TestGraphMetrics_OutDegree_Error(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build graph
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test OutDegree with invalid XREF
	metrics := graph.Metrics()
	outDegree, err := metrics.OutDegree("@INVALID@")
	if err == nil {
		t.Error("Expected error for invalid XREF, got nil")
	}
	if outDegree != 0 {
		t.Errorf("Expected out-degree 0 for invalid XREF, got %d", outDegree)
	}
}

// TestGraphRelationships_ErrorCases tests error cases for relationship helper methods
func TestGraphRelationships_ErrorCases(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build graph
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetChildren with invalid XREF
	children, err := graph.GetChildren("@INVALID@")
	if err == nil {
		t.Error("Expected error for GetChildren with invalid XREF, got nil")
	}
	if children != nil {
		t.Errorf("Expected nil children for invalid XREF, got %v", children)
	}

	// Test GetParents with invalid XREF
	parents, err := graph.GetParents("@INVALID@")
	if err == nil {
		t.Error("Expected error for GetParents with invalid XREF, got nil")
	}
	if parents != nil {
		t.Errorf("Expected nil parents for invalid XREF, got %v", parents)
	}

	// Test GetSiblings with invalid XREF
	siblings, err := graph.GetSiblings("@INVALID@")
	if err == nil {
		t.Error("Expected error for GetSiblings with invalid XREF, got nil")
	}
	if siblings != nil {
		t.Errorf("Expected nil siblings for invalid XREF, got %v", siblings)
	}

	// Test GetSpouses with invalid XREF
	spouses, err := graph.GetSpouses("@INVALID@")
	if err == nil {
		t.Error("Expected error for GetSpouses with invalid XREF, got nil")
	}
	if spouses != nil {
		t.Errorf("Expected nil spouses for invalid XREF, got %v", spouses)
	}

	// Test GetFamilyHusband with invalid XREF
	husband, err := graph.GetFamilyHusband("@INVALID@")
	if err == nil {
		t.Error("Expected error for GetFamilyHusband with invalid XREF, got nil")
	}
	if husband != nil {
		t.Errorf("Expected nil husband for invalid XREF, got %v", husband)
	}

	// Test GetFamilyWife with invalid XREF
	wife, err := graph.GetFamilyWife("@INVALID@")
	if err == nil {
		t.Error("Expected error for GetFamilyWife with invalid XREF, got nil")
	}
	if wife != nil {
		t.Errorf("Expected nil wife for invalid XREF, got %v", wife)
	}

	// Test GetFamilyChildren with invalid XREF
	familyChildren, err := graph.GetFamilyChildren("@INVALID@")
	if err == nil {
		t.Error("Expected error for GetFamilyChildren with invalid XREF, got nil")
	}
	if familyChildren != nil {
		t.Errorf("Expected nil children for invalid XREF, got %v", familyChildren)
	}
}

// TestGraphMetrics_BetweennessCentrality_LargeDataset tests on larger dataset
func TestGraphMetrics_BetweennessCentrality_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}
	if os.Getenv("ENABLE_PERFORMANCE_TESTS") == "" {
		t.Skip("Skipping large dataset test - set ENABLE_PERFORMANCE_TESTS=1 to run")
	}

	filePath := findTestDataFile("royal92.ged")
	if filePath == "" {
		t.Skip("Test data file not found: royal92.ged")
	}

	// Parse and build graph
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test betweenness centrality
	metrics := graph.Metrics()
	centrality, err := metrics.Centrality(CentralityBetweenness)
	if err != nil {
		t.Fatalf("Failed to calculate betweenness centrality: %v", err)
	}

	if len(centrality) == 0 {
		t.Error("Expected at least one centrality value")
	}
}

// TestGraphMetrics_ClosenessCentrality_LargeDataset tests on larger dataset
func TestGraphMetrics_ClosenessCentrality_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}
	if os.Getenv("ENABLE_PERFORMANCE_TESTS") == "" {
		t.Skip("Skipping large dataset test - set ENABLE_PERFORMANCE_TESTS=1 to run")
	}

	filePath := findTestDataFile("royal92.ged")
	if filePath == "" {
		t.Skip("Test data file not found: royal92.ged")
	}

	// Parse and build graph
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test closeness centrality
	metrics := graph.Metrics()
	centrality, err := metrics.Centrality(CentralityCloseness)
	if err != nil {
		t.Fatalf("Failed to calculate closeness centrality: %v", err)
	}

	if len(centrality) == 0 {
		t.Error("Expected at least one centrality value")
	}
}

// TestGraphMetrics_Diameter_LargeDataset tests diameter on larger dataset
func TestGraphMetrics_Diameter_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}
	if os.Getenv("ENABLE_PERFORMANCE_TESTS") == "" {
		t.Skip("Skipping large dataset test - set ENABLE_PERFORMANCE_TESTS=1 to run")
	}

	filePath := findTestDataFile("royal92.ged")
	if filePath == "" {
		t.Skip("Test data file not found: royal92.ged")
	}

	// Parse and build graph
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test diameter
	metrics := graph.Metrics()
	diameter, err := metrics.Diameter()
	if err != nil {
		t.Fatalf("Failed to calculate diameter: %v", err)
	}

	// Diameter should be non-negative
	if diameter < 0 {
		t.Errorf("Diameter is negative: %d", diameter)
	}
}

