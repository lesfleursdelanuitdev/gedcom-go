package query

// EdgeType represents the type of relationship between nodes.
type EdgeType string

const (
	// Individual <-> Family relationships
	EdgeTypeFAMC EdgeType = "FAMC" // Individual is child of family
	EdgeTypeFAMS EdgeType = "FAMS" // Individual is spouse in family
	EdgeTypeCHIL EdgeType = "CHIL" // Family has child
	EdgeTypeHUSB EdgeType = "HUSB" // Family has husband
	EdgeTypeWIFE EdgeType = "WIFE" // Family has wife

	// Reference relationships
	EdgeTypeNOTE     EdgeType = "NOTE"      // References a note
	EdgeTypeSOUR     EdgeType = "SOUR"      // References a source
	EdgeTypeREPO     EdgeType = "REPO"      // References a repository
	EdgeTypeHasEvent EdgeType = "has_event" // Individual/Family has event

	// Derived/computed edges (for convenience)
	EdgeTypeParent  EdgeType = "parent"  // Computed: parent relationship
	EdgeTypeChild   EdgeType = "child"   // Computed: child relationship
	EdgeTypeSpouse  EdgeType = "spouse"  // Computed: spouse relationship
	EdgeTypeSibling EdgeType = "sibling" // Computed: sibling relationship
)

// Direction represents the direction of an edge.
type Direction string

const (
	DirectionForward       Direction = "forward"       // From -> To
	DirectionBackward      Direction = "backward"      // To -> From
	DirectionBidirectional Direction = "bidirectional" // Both directions
)

// Edge represents a relationship between two nodes in the graph.
type Edge struct {
	// Core data
	ID   string    // Unique edge ID
	From GraphNode // Source node
	To   GraphNode // Target node

	// Edge properties
	EdgeType  EdgeType    // Type of relationship
	Direction Direction   // Forward, Backward, Bidirectional
	Family    *FamilyNode // Family context (for FAMC/FAMS edges)

	// Metadata
	Properties map[string]interface{} // Additional edge properties
	Weight     float64                // Optional weight for algorithms
}

// NewEdge creates a new edge between two nodes.
func NewEdge(id string, from GraphNode, to GraphNode, edgeType EdgeType) *Edge {
	return &Edge{
		ID:         id,
		From:       from,
		To:         to,
		EdgeType:   edgeType,
		Direction:  DirectionForward,
		Properties: make(map[string]interface{}),
		Weight:     1.0,
	}
}

// NewBidirectionalEdge creates a bidirectional edge.
func NewBidirectionalEdge(id string, from GraphNode, to GraphNode, edgeType EdgeType) *Edge {
	edge := NewEdge(id, from, to, edgeType)
	edge.Direction = DirectionBidirectional
	return edge
}

// NewEdgeWithFamily creates an edge with family context.
func NewEdgeWithFamily(id string, from GraphNode, to GraphNode, edgeType EdgeType, family *FamilyNode) *Edge {
	edge := NewEdge(id, from, to, edgeType)
	edge.Family = family
	return edge
}

// IsBidirectional returns true if the edge is bidirectional.
func (e *Edge) IsBidirectional() bool {
	return e.Direction == DirectionBidirectional
}

// Connects returns true if the edge connects the given node.
func (e *Edge) Connects(node GraphNode) bool {
	if e.From != nil && e.From.ID() == node.ID() {
		return true
	}
	if e.To != nil && e.To.ID() == node.ID() {
		return true
	}
	return false
}

// OtherNode returns the other node in the edge (not the given node).
func (e *Edge) OtherNode(node GraphNode) GraphNode {
	if e.From != nil && e.From.ID() == node.ID() {
		return e.To
	}
	if e.To != nil && e.To.ID() == node.ID() {
		return e.From
	}
	return nil
}
