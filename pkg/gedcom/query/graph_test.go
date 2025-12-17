package query

import (
	"testing"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

func TestNewGraph(t *testing.T) {
	tree := gedcom.NewGedcomTree()
	graph := NewGraph(tree)

	if graph == nil {
		t.Fatal("Expected graph to be created")
	}

	if graph.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes, got %d", graph.NodeCount())
	}

	if graph.EdgeCount() != 0 {
		t.Errorf("Expected 0 edges, got %d", graph.EdgeCount())
	}
}

func TestGraph_AddNode(t *testing.T) {
	tree := gedcom.NewGedcomTree()
	graph := NewGraph(tree)

	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := gedcom.NewIndividualRecord(indiLine)
	node := NewIndividualNode("@I1@", indi)

	err := graph.AddNode(node)
	if err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}

	if graph.NodeCount() != 1 {
		t.Errorf("Expected 1 node, got %d", graph.NodeCount())
	}

	retrieved := graph.GetIndividual("@I1@")
	if retrieved == nil {
		t.Fatal("Expected to retrieve node")
	}

	if retrieved.ID() != "@I1@" {
		t.Errorf("Expected ID @I1@, got %s", retrieved.ID())
	}
}

func TestGraph_AddNode_Duplicate(t *testing.T) {
	tree := gedcom.NewGedcomTree()
	graph := NewGraph(tree)

	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := gedcom.NewIndividualRecord(indiLine)
	node1 := NewIndividualNode("@I1@", indi)
	node2 := NewIndividualNode("@I1@", indi)

	err := graph.AddNode(node1)
	if err != nil {
		t.Fatalf("Failed to add first node: %v", err)
	}

	err = graph.AddNode(node2)
	if err == nil {
		t.Error("Expected error when adding duplicate node")
	}
}

func TestGraph_AddEdge(t *testing.T) {
	tree := gedcom.NewGedcomTree()
	graph := NewGraph(tree)

	// Create two nodes
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	node1 := NewIndividualNode("@I1@", indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	node2 := NewIndividualNode("@I2@", indi2)

	// Add nodes to graph
	graph.AddNode(node1)
	graph.AddNode(node2)

	// Create edge
	edge := NewEdge("E1", node1, node2, EdgeTypeSpouse)

	err := graph.AddEdge(edge)
	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	if graph.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", graph.EdgeCount())
	}

	// Check edges are added to nodes
	if node1.OutDegree() != 1 {
		t.Errorf("Expected node1 out-degree 1, got %d", node1.OutDegree())
	}

	if node2.InDegree() != 1 {
		t.Errorf("Expected node2 in-degree 1, got %d", node2.InDegree())
	}
}

func TestGraph_AddEdge_Bidirectional(t *testing.T) {
	tree := gedcom.NewGedcomTree()
	graph := NewGraph(tree)

	// Create two nodes
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	node1 := NewIndividualNode("@I1@", indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	node2 := NewIndividualNode("@I2@", indi2)

	// Add nodes to graph
	graph.AddNode(node1)
	graph.AddNode(node2)

	// Create bidirectional edge
	edge := NewBidirectionalEdge("E1", node1, node2, EdgeTypeSpouse)

	err := graph.AddEdge(edge)
	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	// Both nodes should have out and in edges
	if node1.OutDegree() != 1 || node1.InDegree() != 1 {
		t.Errorf("Expected node1 to have 1 in and 1 out edge, got in=%d out=%d", node1.InDegree(), node1.OutDegree())
	}

	if node2.OutDegree() != 1 || node2.InDegree() != 1 {
		t.Errorf("Expected node2 to have 1 in and 1 out edge, got in=%d out=%d", node2.InDegree(), node2.OutDegree())
	}
}

func TestBuildGraph(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Add an individual
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Add a family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	if graph.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes, got %d", graph.NodeCount())
	}

	// Check individual node
	indiNode := graph.GetIndividual("@I1@")
	if indiNode == nil {
		t.Fatal("Expected to find individual node")
	}

	// Check family node
	famNode := graph.GetFamily("@F1@")
	if famNode == nil {
		t.Fatal("Expected to find family node")
	}
}

func TestBuildGraph_WithAllNodeTypes(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Add individual
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Add family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	// Add note
	noteLine := gedcom.NewGedcomLine(0, "NOTE", "", "@N1@")
	note := gedcom.NewNoteRecord(noteLine)
	tree.AddRecord(note)

	// Add source
	sourceLine := gedcom.NewGedcomLine(0, "SOUR", "", "@S1@")
	source := gedcom.NewSourceRecord(sourceLine)
	tree.AddRecord(source)

	// Add repository
	repoLine := gedcom.NewGedcomLine(0, "REPO", "", "@R1@")
	repo := gedcom.NewRepositoryRecord(repoLine)
	tree.AddRecord(repo)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	if graph.NodeCount() != 5 {
		t.Errorf("Expected 5 nodes, got %d", graph.NodeCount())
	}

	// Verify all node types
	if graph.GetIndividual("@I1@") == nil {
		t.Error("Expected individual node")
	}
	if graph.GetFamily("@F1@") == nil {
		t.Error("Expected family node")
	}
	if graph.GetNote("@N1@") == nil {
		t.Error("Expected note node")
	}
	if graph.GetSource("@S1@") == nil {
		t.Error("Expected source node")
	}
	if graph.GetRepository("@R1@") == nil {
		t.Error("Expected repository node")
	}
}
