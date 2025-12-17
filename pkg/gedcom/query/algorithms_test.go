package query

import (
	"testing"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

func TestGraph_BFS(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create a simple family structure
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	visited := make(map[string]bool)
	err = graph.BFS("@I1@", func(node GraphNode) bool {
		visited[node.ID()] = true
		return true
	})

	if err != nil {
		t.Fatalf("BFS failed: %v", err)
	}

	if !visited["@I1@"] {
		t.Error("Expected @I1@ to be visited")
	}

	// Should visit connected nodes (family and spouse)
	if len(visited) < 2 {
		t.Errorf("Expected at least 2 nodes to be visited, got %d", len(visited))
	}
}

func TestGraph_DFS(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	visited := make(map[string]bool)
	err = graph.DFS("@I1@", func(node GraphNode) bool {
		visited[node.ID()] = true
		return true
	})

	if err != nil {
		t.Fatalf("DFS failed: %v", err)
	}

	if !visited["@I1@"] {
		t.Error("Expected @I1@ to be visited")
	}
}

func TestGraph_ShortestPath(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create parent-child relationship
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

	// Path from child to parent
	path, err := graph.ShortestPath("@I2@", "@I1@")
	if err != nil {
		t.Fatalf("Failed to find path: %v", err)
	}

	if path == nil {
		t.Fatal("Expected path to be found")
	}

	if path.Length == 0 {
		t.Error("Expected path length > 0")
	}

	// Check path contains both nodes
	foundI1 := false
	foundI2 := false
	for _, node := range path.Nodes {
		if node.ID() == "@I1@" {
			foundI1 = true
		}
		if node.ID() == "@I2@" {
			foundI2 = true
		}
	}

	if !foundI1 || !foundI2 {
		t.Error("Expected path to contain both @I1@ and @I2@")
	}
}

func TestGraph_ShortestPath_SameNode(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	path, err := graph.ShortestPath("@I1@", "@I1@")
	if err != nil {
		t.Fatalf("Failed to find path: %v", err)
	}

	if path.Length != 0 {
		t.Errorf("Expected path length 0 for same node, got %d", path.Length)
	}

	if len(path.Nodes) != 1 {
		t.Errorf("Expected 1 node in path, got %d", len(path.Nodes))
	}
}

func TestGraph_AllPaths(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create a simple structure
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	paths, err := graph.AllPaths("@I1@", "@I2@", 5)
	if err != nil {
		t.Fatalf("Failed to find paths: %v", err)
	}

	if len(paths) == 0 {
		t.Error("Expected at least one path")
	}
}

func TestGraph_CommonAncestors(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create grandparent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create parent
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create two siblings
	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	indi4Line := gedcom.NewGedcomLine(0, "INDI", "", "@I4@")
	indi4Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi4 := gedcom.NewIndividualRecord(indi4Line)
	tree.AddRecord(indi4)

	// Family 1: I1 is parent of I2
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I3 and I4
	fam2Line := gedcom.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I4@", ""))
	fam2 := gedcom.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Find common ancestors of I3 and I4 (should be I2 and I1)
	common, err := graph.CommonAncestors("@I3@", "@I4@")
	if err != nil {
		t.Fatalf("Failed to find common ancestors: %v", err)
	}

	if len(common) == 0 {
		t.Error("Expected common ancestors")
	}

	// Should include I2 (parent) and I1 (grandparent)
	foundI2 := false
	foundI1 := false
	for _, ancestor := range common {
		if ancestor.ID() == "@I2@" {
			foundI2 = true
		}
		if ancestor.ID() == "@I1@" {
			foundI1 = true
		}
	}

	if !foundI2 {
		t.Error("Expected @I2@ to be a common ancestor")
	}
	if !foundI1 {
		t.Error("Expected @I1@ to be a common ancestor")
	}
}

func TestGraph_LowestCommonAncestor(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create grandparent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create parent
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create two siblings
	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	indi4Line := gedcom.NewGedcomLine(0, "INDI", "", "@I4@")
	indi4Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi4 := gedcom.NewIndividualRecord(indi4Line)
	tree.AddRecord(indi4)

	// Family 1: I1 is parent of I2
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I3 and I4
	fam2Line := gedcom.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I4@", ""))
	fam2 := gedcom.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// LCA of I3 and I4 should be I2 (their parent)
	lca, err := graph.LowestCommonAncestor("@I3@", "@I4@")
	if err != nil {
		t.Fatalf("Failed to find LCA: %v", err)
	}

	if lca.ID() != "@I2@" {
		t.Errorf("Expected LCA to be @I2@, got %s", lca.ID())
	}
}

func TestGraph_CalculateRelationship_ParentChild(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create parent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Calculate relationship from parent to child
	result, err := graph.CalculateRelationship("@I1@", "@I2@")
	if err != nil {
		t.Fatalf("Failed to calculate relationship: %v", err)
	}

	if !result.IsDirect {
		t.Error("Expected direct relationship")
	}

	if result.RelationshipType != "parent" {
		t.Errorf("Expected relationship type 'parent', got '%s'", result.RelationshipType)
	}
}

func TestGraph_CalculateRelationship_Siblings(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create parent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create two children
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Calculate relationship between siblings
	result, err := graph.CalculateRelationship("@I2@", "@I3@")
	if err != nil {
		t.Fatalf("Failed to calculate relationship: %v", err)
	}

	if !result.IsDirect {
		t.Error("Expected direct relationship")
	}

	if result.RelationshipType != "sibling" {
		t.Errorf("Expected relationship type 'sibling', got '%s'", result.RelationshipType)
	}
}

func TestGraph_CalculateRelationship_Spouses(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create two spouses
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Calculate relationship between spouses
	result, err := graph.CalculateRelationship("@I1@", "@I2@")
	if err != nil {
		t.Fatalf("Failed to calculate relationship: %v", err)
	}

	if !result.IsDirect {
		t.Error("Expected direct relationship")
	}

	if result.RelationshipType != "spouse" {
		t.Errorf("Expected relationship type 'spouse', got '%s'", result.RelationshipType)
	}
}

func TestGraph_CalculateRelationship_Cousins(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create grandparent
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create two parents (siblings)
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Create two cousins (children of siblings)
	indi4Line := gedcom.NewGedcomLine(0, "INDI", "", "@I4@")
	indi4Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi4 := gedcom.NewIndividualRecord(indi4Line)
	tree.AddRecord(indi4)

	indi5Line := gedcom.NewGedcomLine(0, "INDI", "", "@I5@")
	indi5Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F3@", ""))
	indi5 := gedcom.NewIndividualRecord(indi5Line)
	tree.AddRecord(indi5)

	// Family 1: I1 is parent of I2 and I3
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I4
	fam2Line := gedcom.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I4@", ""))
	fam2 := gedcom.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	// Family 3: I3 is parent of I5
	fam3Line := gedcom.NewGedcomLine(0, "FAM", "", "@F3@")
	fam3Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I3@", ""))
	fam3Line.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I5@", ""))
	fam3 := gedcom.NewFamilyRecord(fam3Line)
	tree.AddRecord(fam3)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Calculate relationship between cousins
	result, err := graph.CalculateRelationship("@I4@", "@I5@")
	if err != nil {
		t.Fatalf("Failed to calculate relationship: %v", err)
	}

	if !result.IsCollateral {
		t.Error("Expected collateral relationship")
	}

	if result.Degree != 1 {
		t.Errorf("Expected degree 1 (1st cousins), got %d", result.Degree)
	}

	if result.Removal != 0 {
		t.Errorf("Expected removal 0, got %d", result.Removal)
	}
}
