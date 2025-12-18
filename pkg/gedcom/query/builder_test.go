package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestBuildGraph_FamilyEdges(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create individuals
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "Child /Doe/", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	tree.AddRecord(indi3)

	// Create family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Check nodes (3 individuals, 1 family = 4 nodes minimum, plus any events)
	if graph.NodeCount() < 4 {
		t.Errorf("Expected at least 4 nodes, got %d", graph.NodeCount())
	}

	// Check edges
	if graph.EdgeCount() == 0 {
		t.Error("Expected edges to be created")
	}

	// Check HUSB edge
	famNode := graph.GetFamily("@F1@")
	if famNode == nil {
		t.Fatal("Expected family node")
	}

	husbandEdges := 0
	for _, edge := range famNode.OutEdges() {
		if edge.EdgeType == EdgeTypeHUSB {
			husbandEdges++
			if edge.To.ID() != "@I1@" {
				t.Errorf("Expected HUSB edge to point to @I1@, got %s", edge.To.ID())
			}
		}
	}
	if husbandEdges == 0 {
		t.Error("Expected HUSB edge")
	}

	// Check WIFE edge
	wifeEdges := 0
	for _, edge := range famNode.OutEdges() {
		if edge.EdgeType == EdgeTypeWIFE {
			wifeEdges++
			if edge.To.ID() != "@I2@" {
				t.Errorf("Expected WIFE edge to point to @I2@, got %s", edge.To.ID())
			}
		}
	}
	if wifeEdges == 0 {
		t.Error("Expected WIFE edge")
	}

	// Check CHIL edge
	childEdges := 0
	for _, edge := range famNode.OutEdges() {
		if edge.EdgeType == EdgeTypeCHIL {
			childEdges++
			if edge.To.ID() != "@I3@" {
				t.Errorf("Expected CHIL edge to point to @I3@, got %s", edge.To.ID())
			}
		}
	}
	if childEdges == 0 {
		t.Error("Expected CHIL edge")
	}

	// Check FAMS edges (reverse)
	indi1Node := graph.GetIndividual("@I1@")
	famsEdges := 0
	for _, edge := range indi1Node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMS {
			famsEdges++
			if edge.To.ID() != "@F1@" {
				t.Errorf("Expected FAMS edge to point to @F1@, got %s", edge.To.ID())
			}
		}
	}
	if famsEdges == 0 {
		t.Error("Expected FAMS edge from @I1@")
	}

	// Check FAMC edge (reverse)
	indi3Node := graph.GetIndividual("@I3@")
	famcEdges := 0
	for _, edge := range indi3Node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMC {
			famcEdges++
			if edge.To.ID() != "@F1@" {
				t.Errorf("Expected FAMC edge to point to @F1@, got %s", edge.To.ID())
			}
		}
	}
	if famcEdges == 0 {
		t.Error("Expected FAMC edge from @I3@")
	}
}

func TestBuildGraph_ReferenceEdges(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create individual
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NOTE", "@N1@", ""))
	indiLine.AddChild(gedcom.NewGedcomLine(1, "SOUR", "@S1@", ""))
	indi := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Create note
	noteLine := gedcom.NewGedcomLine(0, "NOTE", "", "@N1@")
	noteLine.AddChild(gedcom.NewGedcomLine(1, "CONT", "This is a note", ""))
	note := gedcom.NewNoteRecord(noteLine)
	tree.AddRecord(note)

	// Create source
	sourceLine := gedcom.NewGedcomLine(0, "SOUR", "", "@S1@")
	sourceLine.AddChild(gedcom.NewGedcomLine(1, "TITL", "Source Title", ""))
	source := gedcom.NewSourceRecord(sourceLine)
	tree.AddRecord(source)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Check NOTE edge
	indiNode := graph.GetIndividual("@I1@")
	noteEdges := 0
	for _, edge := range indiNode.OutEdges() {
		if edge.EdgeType == EdgeTypeNOTE {
			noteEdges++
			if edge.To.ID() != "@N1@" {
				t.Errorf("Expected NOTE edge to point to @N1@, got %s", edge.To.ID())
			}
		}
	}
	if noteEdges == 0 {
		t.Error("Expected NOTE edge")
	}

	// Check SOUR edge
	sourEdges := 0
	for _, edge := range indiNode.OutEdges() {
		if edge.EdgeType == EdgeTypeSOUR {
			sourEdges++
			if edge.To.ID() != "@S1@" {
				t.Errorf("Expected SOUR edge to point to @S1@, got %s", edge.To.ID())
			}
		}
	}
	if sourEdges == 0 {
		t.Error("Expected SOUR edge")
	}
}

func TestBuildGraph_EventNodes(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create individual with birth event
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
	birtLine.AddChild(gedcom.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	birtLine.AddChild(gedcom.NewGedcomLine(2, "PLAC", "New York", ""))
	indiLine.AddChild(birtLine)
	indi := gedcom.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Check event node was created
	eventID := "@I1@_BIRT_0"
	eventNode := graph.GetEvent(eventID)
	if eventNode == nil {
		t.Error("Expected event node to be created")
	}

	if eventNode.EventType != "BIRT" {
		t.Errorf("Expected event type BIRT, got %s", eventNode.EventType)
	}

	// Check has_event edge
	indiNode := graph.GetIndividual("@I1@")
	hasEventEdges := 0
	for _, edge := range indiNode.OutEdges() {
		if edge.EdgeType == EdgeTypeHasEvent {
			hasEventEdges++
			if edge.To.ID() != eventID {
				t.Errorf("Expected has_event edge to point to %s, got %s", eventID, edge.To.ID())
			}
		}
	}
	if hasEventEdges == 0 {
		t.Error("Expected has_event edge")
	}
}

func TestBuildGraph_CachedRelationships(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create individuals
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Create family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Check cached relationships
	famNode := graph.GetFamily("@F1@")
	if famNode.Husband == nil {
		t.Error("Expected Husband to be cached")
	}
	if famNode.Husband.ID() != "@I1@" {
		t.Errorf("Expected Husband @I1@, got %s", famNode.Husband.ID())
	}

	if famNode.Wife == nil {
		t.Error("Expected Wife to be cached")
	}
	if famNode.Wife.ID() != "@I2@" {
		t.Errorf("Expected Wife @I2@, got %s", famNode.Wife.ID())
	}

	if len(famNode.Children) == 0 {
		t.Error("Expected Children to be cached")
	}
	if len(famNode.Children) > 0 && famNode.Children[0].ID() != "@I3@" {
		t.Errorf("Expected Child @I3@, got %s", famNode.Children[0].ID())
	}

	// Check individual cached relationships
	indi3Node := graph.GetIndividual("@I3@")
	if len(indi3Node.Parents) == 0 {
		t.Error("Expected Parents to be cached")
	}
}

func TestBuildGraph_CompleteExample(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create a complete family structure
	// Individual 1 (father)
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	birt1Line := gedcom.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(gedcom.NewGedcomLine(2, "DATE", "1 JAN 1850", ""))
	indi1Line.AddChild(birt1Line)
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Individual 2 (mother)
	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Individual 3 (child)
	indi3Line := gedcom.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "Child /Doe/", ""))
	indi3Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi3 := gedcom.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family
	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I3@", ""))
	marrLine := gedcom.NewGedcomLine(1, "MARR", "", "")
	marrLine.AddChild(gedcom.NewGedcomLine(2, "DATE", "1 JAN 1875", ""))
	famLine.AddChild(marrLine)
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Verify structure
	if graph.NodeCount() < 3 {
		t.Errorf("Expected at least 3 nodes, got %d", graph.NodeCount())
	}

	if graph.EdgeCount() == 0 {
		t.Error("Expected edges to be created")
	}

	// Check event node for birth
	birtEventID := "@I1@_BIRT_0"
	birtEvent := graph.GetEvent(birtEventID)
	if birtEvent == nil {
		t.Error("Expected birth event node")
	}

	// Check event node for marriage
	marrEventID := "@F1@_MARR_0"
	marrEvent := graph.GetEvent(marrEventID)
	if marrEvent == nil {
		t.Error("Expected marriage event node")
	}
}
