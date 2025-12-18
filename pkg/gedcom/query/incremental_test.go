package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestGraph_AddNodeIncremental(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create initial graph
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Add a new individual incrementally
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	indi2Node := NewIndividualNode("@I2@", indi2)

	err = graph.AddNodeIncremental(indi2Node)
	if err != nil {
		t.Fatalf("Failed to add node incrementally: %v", err)
	}

	// Verify node was added
	if graph.GetIndividual("@I2@") == nil {
		t.Error("Expected @I2@ to be in graph")
	}

	// Verify node count increased
	if graph.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes, got %d", graph.NodeCount())
	}
}

func TestGraph_RemoveNodeIncremental(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Remove the node
	err = graph.RemoveNodeIncremental("@I1@")
	if err != nil {
		t.Fatalf("Failed to remove node: %v", err)
	}

	// Verify node was removed
	if graph.GetIndividual("@I1@") != nil {
		t.Error("Expected @I1@ to be removed from graph")
	}

	// Verify node count decreased
	if graph.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes, got %d", graph.NodeCount())
	}
}

func TestGraph_AddEdgeIncremental(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Add an edge between them
	indi1Node := graph.GetIndividual("@I1@")
	indi2Node := graph.GetIndividual("@I2@")

	edge := NewEdge("@I1@_spouse_@I2@", indi1Node, indi2Node, EdgeTypeSpouse)

	err = graph.AddEdgeIncremental(edge)
	if err != nil {
		t.Fatalf("Failed to add edge incrementally: %v", err)
	}

	// Verify edge was added
	if graph.GetEdge("@I1@_spouse_@I2@") == nil {
		t.Error("Expected edge to be in graph")
	}

	// Verify edge count increased
	if graph.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", graph.EdgeCount())
	}
}

func TestGraph_RemoveEdgeIncremental(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Add an edge
	indi1Node := graph.GetIndividual("@I1@")
	indi2Node := graph.GetIndividual("@I2@")
	edge := NewEdge("@I1@_spouse_@I2@", indi1Node, indi2Node, EdgeTypeSpouse)
	graph.AddEdge(edge)

	// Remove the edge
	err = graph.RemoveEdgeIncremental("@I1@_spouse_@I2@")
	if err != nil {
		t.Fatalf("Failed to remove edge: %v", err)
	}

	// Verify edge was removed
	if graph.GetEdge("@I1@_spouse_@I2@") != nil {
		t.Error("Expected edge to be removed from graph")
	}

	// Verify edge count decreased
	if graph.EdgeCount() != 0 {
		t.Errorf("Expected 0 edges, got %d", graph.EdgeCount())
	}
}

func TestGraph_AddEdgeIncremental_UpdatesRelationships(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create family structure
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Verify relationships were cached
	indi1Node := graph.GetIndividual("@I1@")
	indi3Node := graph.GetIndividual("@I3@")

	if len(indi1Node.Spouses) == 0 {
		t.Error("Expected @I1@ to have spouse relationship cached")
	}

	if len(indi1Node.Children) == 0 {
		t.Error("Expected @I1@ to have children relationship cached")
	}

	if len(indi3Node.Parents) == 0 {
		t.Error("Expected @I3@ to have parents relationship cached")
	}

	// Add a new child incrementally
	indi4Line := gedcom.NewGedcomLine(0, "INDI", "", "@I4@")
	indi4 := gedcom.NewIndividualRecord(indi4Line)
	indi4Node := NewIndividualNode("@I4@", indi4)
	graph.AddNodeIncremental(indi4Node)

	// Add CHIL edge from family to new child
	famNode := graph.GetFamily("@F1@")
	edge := NewEdgeWithFamily("@F1@_CHIL_@I4@", famNode, indi4Node, EdgeTypeCHIL, famNode)
	err = graph.AddEdgeIncremental(edge)
	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	// Verify relationships were updated
	if len(famNode.Children) != 2 {
		t.Errorf("Expected family to have 2 children, got %d", len(famNode.Children))
	}

	if len(indi1Node.Children) != 2 {
		t.Errorf("Expected @I1@ to have 2 children, got %d", len(indi1Node.Children))
	}

	if len(indi4Node.Parents) != 2 {
		t.Errorf("Expected @I4@ to have 2 parents, got %d", len(indi4Node.Parents))
	}
}

func TestGraph_RemoveEdgeIncremental_UpdatesRelationships(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	indi1Node := graph.GetIndividual("@I1@")
	indi2Node := graph.GetIndividual("@I2@")

	// Verify initial relationships
	if len(indi1Node.Children) != 1 {
		t.Errorf("Expected @I1@ to have 1 child initially, got %d", len(indi1Node.Children))
	}

	// Find and remove the CHIL edge
	famNode := graph.GetFamily("@F1@")
	var chilEdge *Edge
	for _, edge := range famNode.OutEdges() {
		if edge.EdgeType == EdgeTypeCHIL && edge.To.ID() == "@I2@" {
			chilEdge = edge
			break
		}
	}

	if chilEdge == nil {
		t.Fatal("Could not find CHIL edge")
	}

	// Remove the edge
	err = graph.RemoveEdgeIncremental(chilEdge.ID)
	if err != nil {
		t.Fatalf("Failed to remove edge: %v", err)
	}

	// Verify relationships were updated
	if len(indi1Node.Children) != 0 {
		t.Errorf("Expected @I1@ to have 0 children after removal, got %d", len(indi1Node.Children))
	}

	if len(indi2Node.Parents) != 0 {
		t.Errorf("Expected @I2@ to have 0 parents after removal, got %d", len(indi2Node.Parents))
	}
}

func TestGraph_AddNodeIncremental_UpdatesIndexes(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Add an individual with name
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi := gedcom.NewIndividualRecord(indiLine)
	indiNode := NewIndividualNode("@I1@", indi)

	err = graph.AddNodeIncremental(indiNode)
	if err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}

	// Verify index was updated
	results := graph.indexes.findByName("John")
	if len(results) == 0 {
		t.Error("Expected name index to be updated")
	}

	found := false
	for _, xrefID := range results {
		if xrefID == "@I1@" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected @I1@ to be in name index")
	}
}

func TestGraph_RemoveNodeIncremental_UpdatesIndexes(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Verify index has the individual
	results := graph.indexes.findByName("John")
	if len(results) == 0 {
		t.Fatal("Expected name index to contain @I1@")
	}

	// Remove the node
	err = graph.RemoveNodeIncremental("@I1@")
	if err != nil {
		t.Fatalf("Failed to remove node: %v", err)
	}

	// Verify index was updated
	results = graph.indexes.findByName("John")
	found := false
	for _, xrefID := range results {
		if xrefID == "@I1@" {
			found = true
			break
		}
	}
	if found {
		t.Error("Expected @I1@ to be removed from name index")
	}
}

func TestGraph_IncrementalUpdates_CacheInvalidation(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Cache a query result (access Parents field directly)
	indi1Node := graph.GetIndividual("@I1@")
	parents1 := indi1Node.Parents
	_ = parents1

	// Add a new node
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	indi2Node := NewIndividualNode("@I2@", indi2)

	err = graph.AddNodeIncremental(indi2Node)
	if err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}

	// Cache should be cleared (we can't directly test, but verify it doesn't crash)
	parents2 := indi1Node.Parents
	_ = parents2
}
