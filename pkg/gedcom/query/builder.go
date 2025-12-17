package query

import (
	"fmt"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

// BuildGraph constructs a graph from a GEDCOM tree.
// This includes Phase 1 (nodes) and Phase 2 (edges).
func BuildGraph(tree *gedcom.GedcomTree) (*Graph, error) {
	graph := NewGraph(tree)

	// Phase 1: Create all nodes
	if err := createNodes(graph, tree); err != nil {
		return nil, fmt.Errorf("failed to create nodes: %w", err)
	}

	// Phase 2: Create all edges
	if err := createEdges(graph, tree); err != nil {
		return nil, fmt.Errorf("failed to create edges: %w", err)
	}

	// Phase 2: Build cached relationships
	buildCachedRelationships(graph)

	// Phase 3: Build indexes for fast filtering
	graph.indexes.buildIndexes(graph)

	return graph, nil
}

// createNodes creates all nodes from the GEDCOM tree.
func createNodes(graph *Graph, tree *gedcom.GedcomTree) error {
	// Create IndividualNodes
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}
		node := NewIndividualNode(xrefID, indi)
		if err := graph.AddNode(node); err != nil {
			return fmt.Errorf("failed to add individual node %s: %w", xrefID, err)
		}
	}

	// Create FamilyNodes
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		fam, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}
		node := NewFamilyNode(xrefID, fam)
		if err := graph.AddNode(node); err != nil {
			return fmt.Errorf("failed to add family node %s: %w", xrefID, err)
		}
	}

	// Create NoteNodes
	notes := tree.GetAllNotes()
	for xrefID, record := range notes {
		note, ok := record.(*gedcom.NoteRecord)
		if !ok {
			continue
		}
		node := NewNoteNode(xrefID, note)
		if err := graph.AddNode(node); err != nil {
			return fmt.Errorf("failed to add note node %s: %w", xrefID, err)
		}
	}

	// Create SourceNodes
	sources := tree.GetAllSources()
	for xrefID, record := range sources {
		source, ok := record.(*gedcom.SourceRecord)
		if !ok {
			continue
		}
		node := NewSourceNode(xrefID, source)
		if err := graph.AddNode(node); err != nil {
			return fmt.Errorf("failed to add source node %s: %w", xrefID, err)
		}
	}

	// Create RepositoryNodes
	repositories := tree.GetAllRepositories()
	for xrefID, record := range repositories {
		repo, ok := record.(*gedcom.RepositoryRecord)
		if !ok {
			continue
		}
		node := NewRepositoryNode(xrefID, repo)
		if err := graph.AddNode(node); err != nil {
			return fmt.Errorf("failed to add repository node %s: %w", xrefID, err)
		}
	}

	// EventNodes will be created in Phase 2 when we process edges
	// (they're embedded in Individual/Family records)

	return nil
}

// createEdges creates all edges in the graph.
// This includes Individual ↔ Family edges, reference edges, and event edges.
func createEdges(graph *Graph, tree *gedcom.GedcomTree) error {
	// Phase 2.1: Create Individual ↔ Family edges
	if err := createFamilyEdges(graph, tree); err != nil {
		return fmt.Errorf("failed to create family edges: %w", err)
	}

	// Phase 2.2: Create reference edges (NOTE, SOUR, REPO)
	if err := createReferenceEdges(graph, tree); err != nil {
		return fmt.Errorf("failed to create reference edges: %w", err)
	}

	// Phase 2.3: Create event nodes and edges
	if err := createEventNodesAndEdges(graph, tree); err != nil {
		return fmt.Errorf("failed to create event nodes and edges: %w", err)
	}

	return nil
}

// createFamilyEdges creates edges between Individual and Family nodes.
func createFamilyEdges(graph *Graph, tree *gedcom.GedcomTree) error {
	families := tree.GetAllFamilies()

	for famXref, famRecord := range families {
		fam, ok := famRecord.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		famNode := graph.GetFamily(famXref)
		if famNode == nil {
			continue
		}

		// Get HUSB xref
		husbandXref := fam.GetHusband()
		if husbandXref != "" {
			husbandNode := graph.GetIndividual(husbandXref)
			if husbandNode != nil {
				// Family --[HUSB]--> Individual
				edgeID := fmt.Sprintf("%s_HUSB_%s", famXref, husbandXref)
				edge := NewEdgeWithFamily(edgeID, famNode, husbandNode, EdgeTypeHUSB, famNode)
				if err := graph.AddEdge(edge); err != nil {
					return fmt.Errorf("failed to add HUSB edge: %w", err)
				}

				// Individual --[FAMS]--> Family (reverse)
				edgeID2 := fmt.Sprintf("%s_FAMS_%s", husbandXref, famXref)
				edge2 := NewEdgeWithFamily(edgeID2, husbandNode, famNode, EdgeTypeFAMS, famNode)
				if err := graph.AddEdge(edge2); err != nil {
					return fmt.Errorf("failed to add FAMS edge: %w", err)
				}
			}
		}

		// Get WIFE xref
		wifeXref := fam.GetWife()
		if wifeXref != "" {
			wifeNode := graph.GetIndividual(wifeXref)
			if wifeNode != nil {
				// Family --[WIFE]--> Individual
				edgeID := fmt.Sprintf("%s_WIFE_%s", famXref, wifeXref)
				edge := NewEdgeWithFamily(edgeID, famNode, wifeNode, EdgeTypeWIFE, famNode)
				if err := graph.AddEdge(edge); err != nil {
					return fmt.Errorf("failed to add WIFE edge: %w", err)
				}

				// Individual --[FAMS]--> Family (reverse)
				edgeID2 := fmt.Sprintf("%s_FAMS_%s", wifeXref, famXref)
				edge2 := NewEdgeWithFamily(edgeID2, wifeNode, famNode, EdgeTypeFAMS, famNode)
				if err := graph.AddEdge(edge2); err != nil {
					return fmt.Errorf("failed to add FAMS edge: %w", err)
				}
			}
		}

		// Get CHIL xrefs
		children := fam.GetChildren()
		for i, childXref := range children {
			childNode := graph.GetIndividual(childXref)
			if childNode == nil {
				continue
			}

			// Family --[CHIL]--> Individual
			edgeID := fmt.Sprintf("%s_CHIL_%s_%d", famXref, childXref, i)
			edge := NewEdgeWithFamily(edgeID, famNode, childNode, EdgeTypeCHIL, famNode)
			if err := graph.AddEdge(edge); err != nil {
				return fmt.Errorf("failed to add CHIL edge: %w", err)
			}

			// Individual --[FAMC]--> Family (reverse)
			edgeID2 := fmt.Sprintf("%s_FAMC_%s_%d", childXref, famXref, i)
			edge2 := NewEdgeWithFamily(edgeID2, childNode, famNode, EdgeTypeFAMC, famNode)
			if err := graph.AddEdge(edge2); err != nil {
				return fmt.Errorf("failed to add FAMC edge: %w", err)
			}
		}
	}

	return nil
}

// createReferenceEdges creates NOTE, SOUR, and REPO reference edges.
func createReferenceEdges(graph *Graph, tree *gedcom.GedcomTree) error {
	// Individual -> NOTE edges
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		indiNode := graph.GetIndividual(xrefID)
		if indiNode == nil {
			continue
		}

		// NOTE references
		notes := indi.GetNotes()
		for i, noteXref := range notes {
			noteNode := graph.GetNote(noteXref)
			if noteNode != nil {
				edgeID := fmt.Sprintf("%s_NOTE_%s_%d", xrefID, noteXref, i)
				edge := NewEdge(edgeID, indiNode, noteNode, EdgeTypeNOTE)
				if err := graph.AddEdge(edge); err != nil {
					return fmt.Errorf("failed to add NOTE edge: %w", err)
				}
			}
		}

		// SOUR references
		sources := indi.GetSources()
		for i, sourceXref := range sources {
			sourceNode := graph.GetSource(sourceXref)
			if sourceNode != nil {
				edgeID := fmt.Sprintf("%s_SOUR_%s_%d", xrefID, sourceXref, i)
				edge := NewEdge(edgeID, indiNode, sourceNode, EdgeTypeSOUR)
				if err := graph.AddEdge(edge); err != nil {
					return fmt.Errorf("failed to add SOUR edge: %w", err)
				}
			}
		}
	}

	// Family -> NOTE and SOUR edges
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		fam, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		famNode := graph.GetFamily(xrefID)
		if famNode == nil {
			continue
		}

		// NOTE references
		notes := fam.GetNotes()
		for i, noteXref := range notes {
			noteNode := graph.GetNote(noteXref)
			if noteNode != nil {
				edgeID := fmt.Sprintf("%s_NOTE_%s_%d", xrefID, noteXref, i)
				edge := NewEdge(edgeID, famNode, noteNode, EdgeTypeNOTE)
				if err := graph.AddEdge(edge); err != nil {
					return fmt.Errorf("failed to add NOTE edge: %w", err)
				}
			}
		}

		// SOUR references
		sources := fam.GetSources()
		for i, sourceXref := range sources {
			sourceNode := graph.GetSource(sourceXref)
			if sourceNode != nil {
				edgeID := fmt.Sprintf("%s_SOUR_%s_%d", xrefID, sourceXref, i)
				edge := NewEdge(edgeID, famNode, sourceNode, EdgeTypeSOUR)
				if err := graph.AddEdge(edge); err != nil {
					return fmt.Errorf("failed to add SOUR edge: %w", err)
				}
			}
		}
	}

	// Note -> SOUR edges
	notes := tree.GetAllNotes()
	for xrefID, record := range notes {
		note, ok := record.(*gedcom.NoteRecord)
		if !ok {
			continue
		}

		noteNode := graph.GetNote(xrefID)
		if noteNode == nil {
			continue
		}

		sources := note.GetValues("SOUR")
		for i, sourceXref := range sources {
			sourceNode := graph.GetSource(sourceXref)
			if sourceNode != nil {
				edgeID := fmt.Sprintf("%s_SOUR_%s_%d", xrefID, sourceXref, i)
				edge := NewEdge(edgeID, noteNode, sourceNode, EdgeTypeSOUR)
				if err := graph.AddEdge(edge); err != nil {
					return fmt.Errorf("failed to add SOUR edge: %w", err)
				}
			}
		}
	}

	// Source -> REPO edges
	sources := tree.GetAllSources()
	for xrefID, record := range sources {
		source, ok := record.(*gedcom.SourceRecord)
		if !ok {
			continue
		}

		sourceNode := graph.GetSource(xrefID)
		if sourceNode == nil {
			continue
		}

		repoXref := source.GetRepository()
		if repoXref != "" {
			repoNode := graph.GetRepository(repoXref)
			if repoNode != nil {
				edgeID := fmt.Sprintf("%s_REPO_%s", xrefID, repoXref)
				edge := NewEdge(edgeID, sourceNode, repoNode, EdgeTypeREPO)
				if err := graph.AddEdge(edge); err != nil {
					return fmt.Errorf("failed to add REPO edge: %w", err)
				}
			}
		}
	}

	return nil
}

// createEventNodesAndEdges creates EventNodes from embedded events and their edges.
func createEventNodesAndEdges(graph *Graph, tree *gedcom.GedcomTree) error {
	// Process Individual events
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}

		indiNode := graph.GetIndividual(xrefID)
		if indiNode == nil {
			continue
		}

		events := indi.GetEvents()
		for i, eventData := range events {
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)

			// Check if event node already exists
			eventNode := graph.GetEvent(eventID)
			if eventNode == nil {
				// Create new event node
				eventNode = NewEventNode(eventID, eventType, eventData)
				if err := graph.AddNode(eventNode); err != nil {
					return fmt.Errorf("failed to add event node: %w", err)
				}
				eventNode.Owner = indiNode
			}

			// Create has_event edge
			edgeID := fmt.Sprintf("%s_has_event_%s", xrefID, eventID)
			edge := NewEdge(edgeID, indiNode, eventNode, EdgeTypeHasEvent)
			if err := graph.AddEdge(edge); err != nil {
				return fmt.Errorf("failed to add has_event edge: %w", err)
			}

			// Create SOUR edges from event
			// Get SOUR references from event lines
			eventLines := indi.GetLines(eventType)
			for lineIdx, eventLine := range eventLines {
				sourLines := eventLine.GetLines("SOUR")
				for j, sourLine := range sourLines {
					sourceXref := sourLine.Value
					if sourceXref != "" {
						sourceNode := graph.GetSource(sourceXref)
						if sourceNode != nil {
							edgeID := fmt.Sprintf("%s_SOUR_%s_%d_%d", eventID, sourceXref, lineIdx, j)
							edge := NewEdge(edgeID, eventNode, sourceNode, EdgeTypeSOUR)
							if err := graph.AddEdge(edge); err != nil {
								return fmt.Errorf("failed to add SOUR edge from event: %w", err)
							}
						}
					}
				}
			}

			// Create NOTE edges from event
			// Events can have NOTE references, need to check the event lines
			for lineIdx, eventLine := range eventLines {
				// Get NOTE lines from this event line
				noteLines := eventLine.GetLines("NOTE")
				for j, noteLine := range noteLines {
					noteXref := noteLine.Value
					if noteXref != "" {
						noteNode := graph.GetNote(noteXref)
						if noteNode != nil {
							edgeID := fmt.Sprintf("%s_NOTE_%s_%d_%d", eventID, noteXref, lineIdx, j)
							edge := NewEdge(edgeID, eventNode, noteNode, EdgeTypeNOTE)
							if err := graph.AddEdge(edge); err != nil {
								return fmt.Errorf("failed to add NOTE edge from event: %w", err)
							}
						}
					}
				}
			}
		}
	}

	// Process Family events
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		fam, ok := record.(*gedcom.FamilyRecord)
		if !ok {
			continue
		}

		famNode := graph.GetFamily(xrefID)
		if famNode == nil {
			continue
		}

		events := fam.GetEvents()
		for i, eventData := range events {
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)

			// Check if event node already exists
			eventNode := graph.GetEvent(eventID)
			if eventNode == nil {
				// Create new event node
				eventNode = NewEventNode(eventID, eventType, eventData)
				if err := graph.AddNode(eventNode); err != nil {
					return fmt.Errorf("failed to add event node: %w", err)
				}
				eventNode.Owner = famNode
			}

			// Create has_event edge
			edgeID := fmt.Sprintf("%s_has_event_%s", xrefID, eventID)
			edge := NewEdge(edgeID, famNode, eventNode, EdgeTypeHasEvent)
			if err := graph.AddEdge(edge); err != nil {
				return fmt.Errorf("failed to add has_event edge: %w", err)
			}

			// Create SOUR edges from event
			// Get SOUR references from event lines
			eventLines := fam.GetLines(eventType)
			for lineIdx, eventLine := range eventLines {
				sourLines := eventLine.GetLines("SOUR")
				for j, sourLine := range sourLines {
					sourceXref := sourLine.Value
					if sourceXref != "" {
						sourceNode := graph.GetSource(sourceXref)
						if sourceNode != nil {
							edgeID := fmt.Sprintf("%s_SOUR_%s_%d_%d", eventID, sourceXref, lineIdx, j)
							edge := NewEdge(edgeID, eventNode, sourceNode, EdgeTypeSOUR)
							if err := graph.AddEdge(edge); err != nil {
								return fmt.Errorf("failed to add SOUR edge from event: %w", err)
							}
						}
					}
				}
			}

			// Create NOTE edges from event
			for lineIdx, eventLine := range eventLines {
				// Get NOTE lines from this event line
				noteLines := eventLine.GetLines("NOTE")
				for j, noteLine := range noteLines {
					noteXref := noteLine.Value
					if noteXref != "" {
						noteNode := graph.GetNote(noteXref)
						if noteNode != nil {
							edgeID := fmt.Sprintf("%s_NOTE_%s_%d_%d", eventID, noteXref, lineIdx, j)
							edge := NewEdge(edgeID, eventNode, noteNode, EdgeTypeNOTE)
							if err := graph.AddEdge(edge); err != nil {
								return fmt.Errorf("failed to add NOTE edge from event: %w", err)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// buildCachedRelationships builds cached relationships for performance.
func buildCachedRelationships(graph *Graph) {
	// First, build cached relationships for FamilyNodes (needed for IndividualNodes)
	families := graph.GetAllFamilies()
	for _, famNode := range families {
		// Set husband and wife from edges
		for _, edge := range famNode.OutEdges() {
			if edge.EdgeType == EdgeTypeHUSB {
				if indiNode, ok := edge.To.(*IndividualNode); ok {
					famNode.Husband = indiNode
				}
			}
			if edge.EdgeType == EdgeTypeWIFE {
				if indiNode, ok := edge.To.(*IndividualNode); ok {
					famNode.Wife = indiNode
				}
			}
			if edge.EdgeType == EdgeTypeCHIL {
				if indiNode, ok := edge.To.(*IndividualNode); ok {
					famNode.Children = append(famNode.Children, indiNode)
				}
			}
		}
	}

	// Then, build cached relationships for IndividualNodes
	individuals := graph.GetAllIndividuals()
	for _, indiNode := range individuals {
		// Find parents via FAMC edges
		for _, edge := range indiNode.OutEdges() {
			if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
				// Get parents from the family
				famNode := edge.Family
				if famNode.Husband != nil {
					indiNode.Parents = append(indiNode.Parents, famNode.Husband)
				}
				if famNode.Wife != nil {
					indiNode.Parents = append(indiNode.Parents, famNode.Wife)
				}
			}
		}

		// Find children via FAMS -> Family -> CHIL edges
		for _, edge := range indiNode.OutEdges() {
			if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
				famNode := edge.Family
				indiNode.Children = append(indiNode.Children, famNode.Children...)
			}
		}

		// Find spouses via FAMS edges
		for _, edge := range indiNode.OutEdges() {
			if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
				famNode := edge.Family
				if famNode.Husband != nil && famNode.Husband.ID() != indiNode.ID() {
					indiNode.Spouses = append(indiNode.Spouses, famNode.Husband)
				}
				if famNode.Wife != nil && famNode.Wife.ID() != indiNode.ID() {
					indiNode.Spouses = append(indiNode.Spouses, famNode.Wife)
				}
			}
		}

		// Find siblings via shared FAMC
		parentFamilies := make(map[string]*FamilyNode)
		for _, edge := range indiNode.OutEdges() {
			if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
				parentFamilies[edge.Family.ID()] = edge.Family
			}
		}
		for _, famNode := range parentFamilies {
			for _, child := range famNode.Children {
				if child.ID() != indiNode.ID() {
					indiNode.Siblings = append(indiNode.Siblings, child)
				}
			}
		}
	}
}
