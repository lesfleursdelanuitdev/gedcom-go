package query

import (
	"testing"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

func TestNewIndividualNode(t *testing.T) {
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi := gedcom.NewIndividualRecord(indiLine)

	node := NewIndividualNode("@I1@", indi)

	if node.ID() != "@I1@" {
		t.Errorf("Expected ID @I1@, got %s", node.ID())
	}

	if node.NodeType() != NodeTypeIndividual {
		t.Errorf("Expected NodeTypeIndividual, got %s", node.NodeType())
	}

	if node.Record() == nil {
		t.Error("Expected record to be set")
	}

	if node.Individual == nil {
		t.Error("Expected Individual to be set")
	}

	if node.Degree() != 0 {
		t.Errorf("Expected degree 0 for new node, got %d", node.Degree())
	}
}

func TestNewFamilyNode(t *testing.T) {
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam := gedcom.NewFamilyRecord(famLine)

	node := NewFamilyNode("@F1@", fam)

	if node.ID() != "@F1@" {
		t.Errorf("Expected ID @F1@, got %s", node.ID())
	}

	if node.NodeType() != NodeTypeFamily {
		t.Errorf("Expected NodeTypeFamily, got %s", node.NodeType())
	}

	if node.Family == nil {
		t.Error("Expected Family to be set")
	}
}

func TestNewNoteNode(t *testing.T) {
	noteLine := gedcom.NewGedcomLine(0, "NOTE", "", "@N1@")
	noteLine.AddChild(gedcom.NewGedcomLine(1, "CONT", "This is a note", ""))
	note := gedcom.NewNoteRecord(noteLine)

	node := NewNoteNode("@N1@", note)

	if node.ID() != "@N1@" {
		t.Errorf("Expected ID @N1@, got %s", node.ID())
	}

	if node.NodeType() != NodeTypeNote {
		t.Errorf("Expected NodeTypeNote, got %s", node.NodeType())
	}

	if node.Note == nil {
		t.Error("Expected Note to be set")
	}
}

func TestNewSourceNode(t *testing.T) {
	sourceLine := gedcom.NewGedcomLine(0, "SOUR", "", "@S1@")
	sourceLine.AddChild(gedcom.NewGedcomLine(1, "TITL", "Source Title", ""))
	source := gedcom.NewSourceRecord(sourceLine)

	node := NewSourceNode("@S1@", source)

	if node.ID() != "@S1@" {
		t.Errorf("Expected ID @S1@, got %s", node.ID())
	}

	if node.NodeType() != NodeTypeSource {
		t.Errorf("Expected NodeTypeSource, got %s", node.NodeType())
	}

	if node.Source == nil {
		t.Error("Expected Source to be set")
	}
}

func TestNewRepositoryNode(t *testing.T) {
	repoLine := gedcom.NewGedcomLine(0, "REPO", "", "@R1@")
	repoLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "Repository Name", ""))
	repo := gedcom.NewRepositoryRecord(repoLine)

	node := NewRepositoryNode("@R1@", repo)

	if node.ID() != "@R1@" {
		t.Errorf("Expected ID @R1@, got %s", node.ID())
	}

	if node.NodeType() != NodeTypeRepository {
		t.Errorf("Expected NodeTypeRepository, got %s", node.NodeType())
	}

	if node.Repository == nil {
		t.Error("Expected Repository to be set")
	}
}

func TestNewEventNode(t *testing.T) {
	eventData := map[string]interface{}{
		"date":  "1 JAN 1900",
		"place": "New York",
	}

	node := NewEventNode("I1_BIRT_0", "BIRT", eventData)

	if node.ID() != "I1_BIRT_0" {
		t.Errorf("Expected ID I1_BIRT_0, got %s", node.ID())
	}

	if node.NodeType() != NodeTypeEvent {
		t.Errorf("Expected NodeTypeEvent, got %s", node.NodeType())
	}

	if node.EventType != "BIRT" {
		t.Errorf("Expected EventType BIRT, got %s", node.EventType)
	}

	if node.Record() != nil {
		t.Error("Expected EventNode.Record() to return nil")
	}
}

func TestBaseNode_Neighbors(t *testing.T) {
	// Create two nodes
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	node1 := NewIndividualNode("@I1@", indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	node2 := NewIndividualNode("@I2@", indi2)

	// Create an edge between them
	edge := NewEdge("E1", node1, node2, EdgeTypeSpouse)
	node1.AddOutEdge(edge)
	node2.AddInEdge(edge)

	// Check neighbors
	neighbors1 := node1.Neighbors()
	if len(neighbors1) != 1 {
		t.Errorf("Expected 1 neighbor for node1, got %d", len(neighbors1))
	}

	if neighbors1[0].ID() != "@I2@" {
		t.Errorf("Expected neighbor @I2@, got %s", neighbors1[0].ID())
	}

	neighbors2 := node2.Neighbors()
	if len(neighbors2) != 1 {
		t.Errorf("Expected 1 neighbor for node2, got %d", len(neighbors2))
	}
}

func TestBaseNode_Degree(t *testing.T) {
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := gedcom.NewIndividualRecord(indiLine)
	node := NewIndividualNode("@I1@", indi)

	// Initially no edges
	if node.Degree() != 0 {
		t.Errorf("Expected degree 0, got %d", node.Degree())
	}

	if node.InDegree() != 0 {
		t.Errorf("Expected in-degree 0, got %d", node.InDegree())
	}

	if node.OutDegree() != 0 {
		t.Errorf("Expected out-degree 0, got %d", node.OutDegree())
	}

	// Add some edges
	edge1 := &Edge{ID: "E1", From: node, To: node}
	edge2 := &Edge{ID: "E2", From: node, To: node}
	node.AddInEdge(edge1)
	node.AddOutEdge(edge2)

	if node.Degree() != 2 {
		t.Errorf("Expected degree 2, got %d", node.Degree())
	}

	if node.InDegree() != 1 {
		t.Errorf("Expected in-degree 1, got %d", node.InDegree())
	}

	if node.OutDegree() != 1 {
		t.Errorf("Expected out-degree 1, got %d", node.OutDegree())
	}
}
