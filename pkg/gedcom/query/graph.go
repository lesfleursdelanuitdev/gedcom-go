package query

import (
	"fmt"
	"sync"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

// Graph represents the graph structure of a GEDCOM tree.
type Graph struct {
	// Core data
	tree *gedcom.GedcomTree

	// Node storage
	nodes        map[string]GraphNode // All nodes by ID
	individuals  map[string]*IndividualNode
	families     map[string]*FamilyNode
	notes        map[string]*NoteNode
	sources      map[string]*SourceNode
	repositories map[string]*RepositoryNode
	events       map[string]*EventNode

	// Edge storage
	edges     map[string]*Edge   // All edges by ID
	edgeIndex map[string][]*Edge // Edges by node ID (for fast lookup)

	// Thread safety
	mu sync.RWMutex

	// Metadata
	properties map[string]interface{}

	// Performance optimizations
	cache   *queryCache
	indexes *FilterIndexes
}

// NewGraph creates a new empty graph.
func NewGraph(tree *gedcom.GedcomTree) *Graph {
	return &Graph{
		tree:         tree,
		nodes:        make(map[string]GraphNode),
		individuals:  make(map[string]*IndividualNode),
		families:     make(map[string]*FamilyNode),
		notes:        make(map[string]*NoteNode),
		sources:      make(map[string]*SourceNode),
		repositories: make(map[string]*RepositoryNode),
		events:       make(map[string]*EventNode),
		edges:        make(map[string]*Edge),
		edgeIndex:    make(map[string][]*Edge),
		properties:   make(map[string]interface{}),
		cache:        newQueryCache(1000), // Default cache size
		indexes:      newFilterIndexes(),
	}
}

// GetNode returns a node by ID.
func (g *Graph) GetNode(id string) GraphNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.nodes[id]
}

// GetIndividual returns an IndividualNode by xref ID.
func (g *Graph) GetIndividual(xrefID string) *IndividualNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.individuals[xrefID]
}

// GetFamily returns a FamilyNode by xref ID.
func (g *Graph) GetFamily(xrefID string) *FamilyNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.families[xrefID]
}

// GetNote returns a NoteNode by xref ID.
func (g *Graph) GetNote(xrefID string) *NoteNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.notes[xrefID]
}

// GetSource returns a SourceNode by xref ID.
func (g *Graph) GetSource(xrefID string) *SourceNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.sources[xrefID]
}

// GetRepository returns a RepositoryNode by xref ID.
func (g *Graph) GetRepository(xrefID string) *RepositoryNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.repositories[xrefID]
}

// GetEvent returns an EventNode by event ID.
func (g *Graph) GetEvent(eventID string) *EventNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.events[eventID]
}

// GetAllNodes returns all nodes in the graph.
func (g *Graph) GetAllNodes() map[string]GraphNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]GraphNode)
	for k, v := range g.nodes {
		result[k] = v
	}
	return result
}

// GetAllIndividuals returns all IndividualNodes.
func (g *Graph) GetAllIndividuals() map[string]*IndividualNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*IndividualNode)
	for k, v := range g.individuals {
		result[k] = v
	}
	return result
}

// GetAllFamilies returns all FamilyNodes.
func (g *Graph) GetAllFamilies() map[string]*FamilyNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*FamilyNode)
	for k, v := range g.families {
		result[k] = v
	}
	return result
}

// AddNode adds a node to the graph.
func (g *Graph) AddNode(node GraphNode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := node.ID()
	if id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	// Check if node already exists
	if _, exists := g.nodes[id]; exists {
		return fmt.Errorf("node with ID %s already exists", id)
	}

	// Add to nodes map
	g.nodes[id] = node

	// Add to type-specific map
	switch node.NodeType() {
	case NodeTypeIndividual:
		if indiNode, ok := node.(*IndividualNode); ok {
			g.individuals[id] = indiNode
		}
	case NodeTypeFamily:
		if famNode, ok := node.(*FamilyNode); ok {
			g.families[id] = famNode
		}
	case NodeTypeNote:
		if noteNode, ok := node.(*NoteNode); ok {
			g.notes[id] = noteNode
		}
	case NodeTypeSource:
		if sourceNode, ok := node.(*SourceNode); ok {
			g.sources[id] = sourceNode
		}
	case NodeTypeRepository:
		if repoNode, ok := node.(*RepositoryNode); ok {
			g.repositories[id] = repoNode
		}
	case NodeTypeEvent:
		if eventNode, ok := node.(*EventNode); ok {
			g.events[id] = eventNode
		}
	}

	return nil
}

// AddEdge adds an edge to the graph.
func (g *Graph) AddEdge(edge *Edge) error {
	g.mu.Lock()
	defer g.mu.Unlock()

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

	// Add to edge index
	fromID := edge.From.ID()
	toID := edge.To.ID()

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

// GetEdges returns all edges for a given node ID.
func (g *Graph) GetEdges(nodeID string) []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edgeIndex[nodeID]
}

// GetEdge returns an edge by ID.
func (g *Graph) GetEdge(edgeID string) *Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edges[edgeID]
}

// GetAllEdges returns all edges in the graph.
func (g *Graph) GetAllEdges() map[string]*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*Edge)
	for k, v := range g.edges {
		result[k] = v
	}
	return result
}

// NodeCount returns the total number of nodes.
func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the total number of edges.
func (g *Graph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.edges)
}

// Tree returns the underlying GEDCOM tree.
func (g *Graph) Tree() *gedcom.GedcomTree {
	return g.tree
}
