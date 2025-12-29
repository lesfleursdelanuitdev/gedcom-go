package query

import (
	"fmt"
	"log"
)

// GetNode returns a node by XREF ID (external API - still uses strings).
func (g *Graph) GetNode(xrefID string) GraphNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.nodes[id]
}

// GetIndividual returns an IndividualNode by xref ID (external API - still uses strings).
// If lazy mode is enabled, loads the node on-demand.
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetIndividual(xrefID string) *IndividualNode {
	// If hybrid mode, load from BadgerDB (SQLite or PostgreSQL)
	if g.hybridMode && (g.hybridStorage != nil || g.hybridStoragePostgres != nil) {
		if debugHybrid {
			log.Printf("[HYBRID] GetIndividual: xrefID=%s, hybridMode=true", xrefID)
		}
		node := g.getIndividualFromHybrid(xrefID)
		if debugHybrid {
			if node == nil {
				log.Printf("[HYBRID] GetIndividual: returned nil for %s", xrefID)
			} else {
				log.Printf("[HYBRID] GetIndividual: returned node %s", xrefID)
			}
		}
		return node
	}

	g.mu.RLock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		g.mu.RUnlock()
		return nil
	}
	node := g.individuals[id]
	lazyMode := g.lazyMode
	g.mu.RUnlock()

	// If lazy mode and node not loaded, load it
	if lazyMode && node == nil {
		loadedNode, err := g.ensureNodeLoaded(xrefID)
		if err != nil || loadedNode == nil {
			return nil
		}
		if indiNode, ok := loadedNode.(*IndividualNode); ok {
			// Ensure edges are loaded when node is accessed (without holding lock)
			g.ensureEdgesLoadedUnlocked(indiNode)
			return indiNode
		}
		return nil
	}

	// Ensure edges are loaded if lazy mode (without holding lock)
	if lazyMode && node != nil {
		g.ensureEdgesLoadedUnlocked(node)
	}

	return node
}

// GetFamily returns a FamilyNode by xref ID (external API - still uses strings).
// If lazy mode is enabled, loads the node on-demand.
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetFamily(xrefID string) *FamilyNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getFamilyFromHybrid(xrefID)
	}

	g.mu.RLock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		g.mu.RUnlock()
		return nil
	}
	node := g.families[id]
	lazyMode := g.lazyMode
	g.mu.RUnlock()

	// If lazy mode and node not loaded, load it
	if lazyMode && node == nil {
		loadedNode, err := g.ensureNodeLoaded(xrefID)
		if err != nil || loadedNode == nil {
			return nil
		}
		if famNode, ok := loadedNode.(*FamilyNode); ok {
			// Ensure edges are loaded when node is accessed (without holding lock)
			g.ensureEdgesLoadedUnlocked(famNode)
			return famNode
		}
		return nil
	}

	// Ensure edges are loaded if lazy mode (without holding lock)
	if lazyMode && node != nil {
		g.ensureEdgesLoadedUnlocked(node)
	}

	return node
}

// GetNote returns a NoteNode by xref ID (external API - still uses strings).
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetNote(xrefID string) *NoteNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		if debugHybrid {
			log.Printf("[HYBRID] GetNote: xrefID=%s, hybridMode=true", xrefID)
		}
		node := g.getNoteFromHybrid(xrefID)
		if debugHybrid {
			if node == nil {
				log.Printf("[HYBRID] GetNote: returned nil for %s", xrefID)
			} else {
				log.Printf("[HYBRID] GetNote: returned node %s", xrefID)
			}
		}
		return node
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.notes[id]
}

// GetSource returns a SourceNode by xref ID (external API - still uses strings).
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetSource(xrefID string) *SourceNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getSourceFromHybrid(xrefID)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.sources[id]
}

// GetRepository returns a RepositoryNode by xref ID (external API - still uses strings).
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetRepository(xrefID string) *RepositoryNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getRepositoryFromHybrid(xrefID)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.repositories[id]
}

// GetEvent returns an EventNode by event ID (external API - still uses strings).
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetEvent(eventID string) *EventNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getEventFromHybrid(eventID)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[eventID]
	if id == 0 {
		return nil
	}
	return g.events[id]
}

// GetAllNodes returns all nodes in the graph (external API - returns XREF strings).
func (g *Graph) GetAllNodes() map[string]GraphNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]GraphNode)
	for id, node := range g.nodes {
		xrefID := g.idToXref[id]
		if xrefID != "" {
			result[xrefID] = node
		}
	}
	return result
}

// GetAllIndividuals returns all IndividualNodes (external API - returns XREF strings).
// If hybrid mode is enabled, queries SQLite for all individual IDs.
// NOTE: In hybrid mode, this returns an empty map because nodes are loaded on-demand.
// Use FilterQuery.executeHybrid() for efficient querying in hybrid mode.
func (g *Graph) GetAllIndividuals() map[string]*IndividualNode {
	// If hybrid mode, return empty map - nodes loaded on-demand via GetIndividual
	// Note: We check queryHelpers instead of hybridStorage because queryHelpers
	// is required for SQLite queries, and this method is primarily used for
	// in-memory iteration. FilterQuery.executeHybrid() handles hybrid mode queries.
	if g.hybridMode && g.queryHelpers != nil {
		return make(map[string]*IndividualNode)
	}
	return g.getAllIndividualsInMemory()
}

// getAllIndividualsInMemory returns in-memory individuals
func (g *Graph) getAllIndividualsInMemory() map[string]*IndividualNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*IndividualNode)
	for id, node := range g.individuals {
		xrefID := g.idToXref[id]
		if xrefID != "" {
			result[xrefID] = node
		}
	}
	return result
}

// GetAllFamilies returns all FamilyNodes (external API - returns XREF strings).
func (g *Graph) GetAllFamilies() map[string]*FamilyNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*FamilyNode)
	for id, node := range g.families {
		xrefID := g.idToXref[id]
		if xrefID != "" {
			result[xrefID] = node
		}
	}
	return result
}

// getBaseNode extracts the BaseNode from a GraphNode.
// All node types embed *BaseNode, so we can access it via type assertion.
func getBaseNode(node GraphNode) *BaseNode {
	switch n := node.(type) {
	case *IndividualNode:
		return n.BaseNode
	case *FamilyNode:
		return n.BaseNode
	case *NoteNode:
		return n.BaseNode
	case *SourceNode:
		return n.BaseNode
	case *RepositoryNode:
		return n.BaseNode
	case *EventNode:
		return n.BaseNode
	default:
		return nil
	}
}

// AddNode adds a node to the graph (external API - still uses XREF strings).
func (g *Graph) AddNode(node GraphNode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

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
