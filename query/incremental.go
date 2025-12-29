package query

import (
	"fmt"
	"sort"
	"strings"
)

// AddNodeIncremental adds a node to the graph and updates relationships incrementally.
// This is more efficient than rebuilding the entire graph.
func (g *Graph) AddNodeIncremental(node GraphNode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Add the node using existing AddNode logic
	if err := g.addNodeInternal(node); err != nil {
		return err
	}

	// Update indexes if it's an individual
	if indiNode, ok := node.(*IndividualNode); ok {
		g.updateIndexesForIndividual(indiNode)
	}

	// Invalidate cache (relationships may have changed)
	g.cache.clear()

	return nil
}

// RemoveNodeIncremental removes a node from the graph and cleans up relationships.
func (g *Graph) RemoveNodeIncremental(xrefID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Get uint32 ID
	internalID := g.xrefToID[xrefID]
	if internalID == 0 {
		return fmt.Errorf("node %s not found", xrefID)
	}

	node := g.nodes[internalID]
	if node == nil {
		return fmt.Errorf("node %s not found", xrefID)
	}

	// Remove all edges connected to this node
	edgesToRemove := make([]*Edge, 0)
	for _, edge := range node.OutEdges() {
		edgesToRemove = append(edgesToRemove, edge)
	}
	for _, edge := range node.InEdges() {
		edgesToRemove = append(edgesToRemove, edge)
	}

	// Remove edges (this will update relationships)
	for _, edge := range edgesToRemove {
		if err := g.removeEdgeInternal(edge.ID); err != nil {
			// Continue even if edge removal fails
			continue
		}
	}

	// Remove from type-specific maps (using uint32 ID)
	switch node.NodeType() {
	case NodeTypeIndividual:
		delete(g.individuals, internalID)
		// Update indexes (remove from all indexes)
		g.removeFromIndexes(xrefID)
	case NodeTypeFamily:
		delete(g.families, internalID)
	case NodeTypeNote:
		delete(g.notes, internalID)
	case NodeTypeSource:
		delete(g.sources, internalID)
	case NodeTypeRepository:
		delete(g.repositories, internalID)
	case NodeTypeEvent:
		delete(g.events, internalID)
	}

	// Remove from main nodes map (using uint32 ID)
	delete(g.nodes, internalID)

	// Invalidate cache
	g.cache.clear()

	return nil
}

// AddEdgeIncremental adds an edge to the graph and updates relationships incrementally.
func (g *Graph) AddEdgeIncremental(edge *Edge) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Add the edge using existing AddEdge logic
	if err := g.addEdgeInternal(edge); err != nil {
		return err
	}

	// For family relationship edges, also create the reverse edge (like in createFamilyEdges)
	// This ensures bidirectional relationships work correctly
	switch edge.EdgeType {
	case EdgeTypeHUSB:
		// Family --[HUSB]--> Individual, also create Individual --[FAMS]--> Family
		if famNode, ok := edge.From.(*FamilyNode); ok {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				edgeID2 := fmt.Sprintf("%s_FAMS_%s", indiNode.ID(), famNode.ID())
				edge2 := NewEdgeWithFamily(edgeID2, indiNode, famNode, EdgeTypeFAMS, famNode)
				if err := g.addEdgeInternal(edge2); err != nil {
					// If reverse edge already exists, that's okay
				}
			}
		}
	case EdgeTypeWIFE:
		// Family --[WIFE]--> Individual, also create Individual --[FAMS]--> Family
		if famNode, ok := edge.From.(*FamilyNode); ok {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				edgeID2 := fmt.Sprintf("%s_FAMS_%s", indiNode.ID(), famNode.ID())
				edge2 := NewEdgeWithFamily(edgeID2, indiNode, famNode, EdgeTypeFAMS, famNode)
				if err := g.addEdgeInternal(edge2); err != nil {
					// If reverse edge already exists, that's okay
				}
			}
		}
	case EdgeTypeCHIL:
		// Family --[CHIL]--> Individual, also create Individual --[FAMC]--> Family
		if famNode, ok := edge.From.(*FamilyNode); ok {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				edgeID2 := fmt.Sprintf("%s_FAMC_%s", indiNode.ID(), famNode.ID())
				edge2 := NewEdgeWithFamily(edgeID2, indiNode, famNode, EdgeTypeFAMC, famNode)
				if err := g.addEdgeInternal(edge2); err != nil {
					// If reverse edge already exists, that's okay
				}
			}
		}
	}

	// Update cached relationships based on edge type (no-op now, but kept for compatibility)
	g.updateRelationshipsForEdge(edge)

	// Invalidate cache
	g.cache.clear()

	return nil
}

// RemoveEdgeIncremental removes an edge from the graph and updates relationships.
func (g *Graph) RemoveEdgeIncremental(edgeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	edge := g.edges[edgeID]
	if edge == nil {
		return fmt.Errorf("edge %s not found", edgeID)
	}

	// Store edge info before removal (for relationship updates)
	fromNode := edge.From
	toNode := edge.To
	edgeType := edge.EdgeType

	// Remove the edge
	if err := g.removeEdgeInternal(edgeID); err != nil {
		return err
	}

	// For family relationship edges, also remove the reverse edge
	switch edgeType {
	case EdgeTypeHUSB:
		// Remove corresponding FAMS edge
		if indiNode, ok := toNode.(*IndividualNode); ok {
			if famNode, ok := fromNode.(*FamilyNode); ok {
				reverseEdgeID := fmt.Sprintf("%s_FAMS_%s", indiNode.ID(), famNode.ID())
				if reverseEdge := g.edges[reverseEdgeID]; reverseEdge != nil {
					g.removeEdgeInternal(reverseEdgeID)
				}
			}
		}
	case EdgeTypeWIFE:
		// Remove corresponding FAMS edge
		if indiNode, ok := toNode.(*IndividualNode); ok {
			if famNode, ok := fromNode.(*FamilyNode); ok {
				reverseEdgeID := fmt.Sprintf("%s_FAMS_%s", indiNode.ID(), famNode.ID())
				if reverseEdge := g.edges[reverseEdgeID]; reverseEdge != nil {
					g.removeEdgeInternal(reverseEdgeID)
				}
			}
		}
	case EdgeTypeCHIL:
		// Remove corresponding FAMC edge (may have index suffix)
		if indiNode, ok := toNode.(*IndividualNode); ok {
			if famNode, ok := fromNode.(*FamilyNode); ok {
				// Find FAMC edge (may have index suffix like _0, _1, etc.)
				edgesToRemove := make([]string, 0)
				for eID, e := range g.edges {
					if e.EdgeType == EdgeTypeFAMC && e.From == indiNode && e.To == famNode {
						edgesToRemove = append(edgesToRemove, eID)
					}
				}
				for _, eID := range edgesToRemove {
					g.removeEdgeInternal(eID)
				}
			}
		}
	case EdgeTypeFAMS:
		// Remove corresponding HUSB or WIFE edge
		if indiNode, ok := fromNode.(*IndividualNode); ok {
			if famNode, ok := toNode.(*FamilyNode); ok {
				// Check for HUSB edge
				husbEdgeID := fmt.Sprintf("%s_HUSB_%s", famNode.ID(), indiNode.ID())
				if husbEdge := g.edges[husbEdgeID]; husbEdge != nil {
					g.removeEdgeInternal(husbEdgeID)
				} else {
					// Check for WIFE edge
					wifeEdgeID := fmt.Sprintf("%s_WIFE_%s", famNode.ID(), indiNode.ID())
					if wifeEdge := g.edges[wifeEdgeID]; wifeEdge != nil {
						g.removeEdgeInternal(wifeEdgeID)
					}
				}
			}
		}
	case EdgeTypeFAMC:
		// Remove corresponding CHIL edge
		if indiNode, ok := fromNode.(*IndividualNode); ok {
			if famNode, ok := toNode.(*FamilyNode); ok {
				// Find CHIL edge (may have index suffix)
				edgesToRemove := make([]string, 0)
				for eID, e := range g.edges {
					if e.EdgeType == EdgeTypeCHIL && e.From == famNode && e.To == indiNode {
						edgesToRemove = append(edgesToRemove, eID)
					}
				}
				for _, eID := range edgesToRemove {
					g.removeEdgeInternal(eID)
				}
			}
		}
	}

	// Update cached relationships (no-op now, but kept for compatibility)
	g.updateRelationshipsAfterEdgeRemoval(fromNode, toNode, edgeType)

	// Invalidate cache
	g.cache.clear()

	return nil
}

// Internal helper methods (must be called with lock held)

func (g *Graph) addNodeInternal(node GraphNode) error {
	xrefID := node.ID()
	if xrefID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	// Get or create uint32 ID for this XREF
	internalID := g.getOrCreateID(xrefID)

	// Check if node already exists
	if _, exists := g.nodes[internalID]; exists {
		return fmt.Errorf("node with ID %s already exists", xrefID)
	}

	// Phase 3: Store nodeID on BaseNode for fast access (eliminates GetNodeID() overhead)
	if baseNode := getBaseNode(node); baseNode != nil {
		baseNode.nodeID = internalID
	}

	// Add to nodes map (using uint32 ID)
	g.nodes[internalID] = node

	// Add to type-specific map (using uint32 ID)
	switch node.NodeType() {
	case NodeTypeIndividual:
		if indiNode, ok := node.(*IndividualNode); ok {
			g.individuals[internalID] = indiNode
		}
	case NodeTypeFamily:
		if famNode, ok := node.(*FamilyNode); ok {
			g.families[internalID] = famNode
		}
	case NodeTypeNote:
		if noteNode, ok := node.(*NoteNode); ok {
			g.notes[internalID] = noteNode
		}
	case NodeTypeSource:
		if sourceNode, ok := node.(*SourceNode); ok {
			g.sources[internalID] = sourceNode
		}
	case NodeTypeRepository:
		if repoNode, ok := node.(*RepositoryNode); ok {
			g.repositories[internalID] = repoNode
		}
	case NodeTypeEvent:
		if eventNode, ok := node.(*EventNode); ok {
			g.events[internalID] = eventNode
		}
	}

	return nil
}

func (g *Graph) addEdgeInternal(edge *Edge) error {
	if edge == nil {
		return fmt.Errorf("edge cannot be nil")
	}

	if edge.ID == "" {
		return fmt.Errorf("edge ID cannot be empty")
	}

	if edge.From == nil || edge.To == nil {
		return fmt.Errorf("edge must have both From and To nodes")
	}

	// Check if edge already exists
	if _, exists := g.edges[edge.ID]; exists {
		return fmt.Errorf("edge with ID %s already exists", edge.ID)
	}

	// Add to edges map
	g.edges[edge.ID] = edge

	// Add to edge index (using uint32 IDs)
	fromXref := edge.From.ID()
	toXref := edge.To.ID()
	fromID := g.getOrCreateID(fromXref)
	toID := g.getOrCreateID(toXref)

	g.edgeIndex[fromID] = append(g.edgeIndex[fromID], edge)
	g.edgeIndex[toID] = append(g.edgeIndex[toID], edge)

	// Add to node's edge lists
	edge.From.AddOutEdge(edge)
	edge.To.AddInEdge(edge)

	// If bidirectional, also add reverse
	if edge.IsBidirectional() {
		edge.To.AddOutEdge(edge)
		edge.From.AddInEdge(edge)
	}

	return nil
}

func (g *Graph) removeEdgeInternal(edgeID string) error {
	edge := g.edges[edgeID]
	if edge == nil {
		return fmt.Errorf("edge %s not found", edgeID)
	}

	fromXref := edge.From.ID()
	toXref := edge.To.ID()
	fromID := g.getID(fromXref)
	toID := g.getID(toXref)

	// Remove from edge index (using uint32 IDs)
	if fromID != 0 {
		g.removeFromEdgeIndex(fromID, edge)
	}
	if toID != 0 {
		g.removeFromEdgeIndex(toID, edge)
	}

	// Remove from node's edge lists
	edge.From.RemoveOutEdge(edge)
	edge.To.RemoveInEdge(edge)

	// If bidirectional, also remove reverse
	if edge.IsBidirectional() {
		edge.To.RemoveOutEdge(edge)
		edge.From.RemoveInEdge(edge)
	}

	// Remove from edges map
	delete(g.edges, edgeID)

	return nil
}

func (g *Graph) removeFromEdgeIndex(nodeID uint32, edge *Edge) {
	edges := g.edgeIndex[nodeID]
	for i, e := range edges {
		if e.ID == edge.ID {
			// Remove by swapping with last element and truncating
			edges[i] = edges[len(edges)-1]
			g.edgeIndex[nodeID] = edges[:len(edges)-1]
			return
		}
	}
}

// updateRelationshipsForEdge is no longer needed.
// Relationships are now computed on-demand from edges, so we don't need to maintain cached relationships.
func (g *Graph) updateRelationshipsForEdge(edge *Edge) {
	// No-op: relationships computed on-demand from edges
}

// updateRelationshipsAfterEdgeRemoval is no longer needed.
// Relationships are now computed on-demand from edges, so we don't need to maintain cached relationships.
func (g *Graph) updateRelationshipsAfterEdgeRemoval(fromNode, toNode GraphNode, edgeType EdgeType) {
	// No-op: relationships computed on-demand from edges
}

// Helper functions for relationship management are no longer needed.
// Relationships are now computed on-demand from edges.
// These functions are kept as no-ops for compatibility but are not used.
// func (g *Graph) addSpouseRelationship(indi1, indi2 *IndividualNode) { }
// func (g *Graph) removeSpouseRelationship(indi1 *IndividualNode, indi2 GraphNode) { }
// func (g *Graph) addParentChildRelationship(parent, child *IndividualNode) { }
// func (g *Graph) removeParentChildRelationship(parent *IndividualNode, child GraphNode) { }
// func (g *Graph) addSiblingRelationship(sib1, sib2 *IndividualNode) { }
// func (g *Graph) removeSiblingRelationship(sib1 *IndividualNode, sib2 GraphNode) { }

// Contains checks and remove helpers are no longer needed.
// Relationships are now computed on-demand from edges, so these helper functions
// that operated on cached relationship slices are obsolete.
// They are kept as no-ops for compatibility but are not used.
// func (g *Graph) containsSpouse(spouses []*IndividualNode, spouse *IndividualNode) bool { return false }
// func (g *Graph) containsChild(children []*IndividualNode, child *IndividualNode) bool { return false }
// func (g *Graph) containsParent(parents []*IndividualNode, parent *IndividualNode) bool { return false }
// func (g *Graph) containsSibling(siblings []*IndividualNode, sibling *IndividualNode) bool { return false }
// func (g *Graph) removeSpouse(spouses []*IndividualNode, spouse *IndividualNode) []*IndividualNode { return nil }
// func (g *Graph) removeChild(children []*IndividualNode, child GraphNode) []*IndividualNode { return nil }
// func (g *Graph) removeParent(parents []*IndividualNode, parent *IndividualNode) []*IndividualNode { return nil }
// func (g *Graph) removeSibling(siblings []*IndividualNode, sibling *IndividualNode) []*IndividualNode { return nil }

// Index management

func (g *Graph) updateIndexesForIndividual(indiNode *IndividualNode) {
	if indiNode.Individual == nil {
		return
	}

	indi := indiNode.Individual
	xrefID := indiNode.ID()

	// Update name index
	name := strings.ToLower(indi.GetName())
	if name != "" {
		g.indexes.nameIndex[name] = append(g.indexes.nameIndex[name], xrefID)
		words := strings.Fields(name)
		for _, word := range words {
			if len(word) > 2 {
				g.indexes.nameIndex[word] = append(g.indexes.nameIndex[word], xrefID)
			}
		}
	}

	// Update birth date index
	birthDate, err := indi.GetBirthDateParsed()
	if err == nil && birthDate != nil && birthDate.IsValid() {
		g.indexes.birthDateIndex = append(g.indexes.birthDateIndex, &dateIndexEntry{
			xrefID:    xrefID,
			birthDate: birthDate,
		})
		// Re-sort (could be optimized with insertion sort)
		sort.Slice(g.indexes.birthDateIndex, func(i, j int) bool {
			dateI := g.indexes.birthDateIndex[i].birthDate.Earliest()
			dateJ := g.indexes.birthDateIndex[j].birthDate.Earliest()
			return dateI.Before(dateJ)
		})
	}

	// Update place index
	birthPlace := strings.ToLower(indi.GetBirthPlace())
	if birthPlace != "" {
		g.indexes.placeIndex[birthPlace] = append(g.indexes.placeIndex[birthPlace], xrefID)
		words := strings.Fields(birthPlace)
		for _, word := range words {
			if len(word) > 2 {
				g.indexes.placeIndex[word] = append(g.indexes.placeIndex[word], xrefID)
			}
		}
	}

	// Update sex index
	sex := strings.ToUpper(indi.GetSex())
	if sex != "" {
		g.indexes.sexIndex[sex] = append(g.indexes.sexIndex[sex], xrefID)
	}

	// Update boolean indexes
	// Compute from edges instead of cached fields
	children := indiNode.getChildrenFromEdges()
	g.indexes.hasChildrenIndex[xrefID] = len(children) > 0
	spouses := indiNode.getSpousesFromEdges()
	g.indexes.hasSpouseIndex[xrefID] = len(spouses) > 0
	g.indexes.livingIndex[xrefID] = indi.GetDeathDate() == ""
}

func (g *Graph) removeFromIndexes(xrefID string) {
	// Remove from name index
	for key, xrefIDs := range g.indexes.nameIndex {
		g.indexes.nameIndex[key] = g.removeFromSlice(xrefIDs, xrefID)
		if len(g.indexes.nameIndex[key]) == 0 {
			delete(g.indexes.nameIndex, key)
		}
	}

	// Remove from birth date index
	newDateIndex := make([]*dateIndexEntry, 0)
	for _, entry := range g.indexes.birthDateIndex {
		if entry.xrefID != xrefID {
			newDateIndex = append(newDateIndex, entry)
		}
	}
	g.indexes.birthDateIndex = newDateIndex

	// Remove from place index
	for key, xrefIDs := range g.indexes.placeIndex {
		g.indexes.placeIndex[key] = g.removeFromSlice(xrefIDs, xrefID)
		if len(g.indexes.placeIndex[key]) == 0 {
			delete(g.indexes.placeIndex, key)
		}
	}

	// Remove from sex index
	for key, xrefIDs := range g.indexes.sexIndex {
		g.indexes.sexIndex[key] = g.removeFromSlice(xrefIDs, xrefID)
		if len(g.indexes.sexIndex[key]) == 0 {
			delete(g.indexes.sexIndex, key)
		}
	}

	// Remove from boolean indexes
	delete(g.indexes.hasChildrenIndex, xrefID)
	delete(g.indexes.hasSpouseIndex, xrefID)
	delete(g.indexes.livingIndex, xrefID)
}

func (g *Graph) removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
