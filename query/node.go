package query

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// NodeType represents the type of a graph node.
type NodeType string

const (
	NodeTypeIndividual NodeType = "individual"
	NodeTypeFamily     NodeType = "family"
	NodeTypeNote       NodeType = "note"
	NodeTypeSource     NodeType = "source"
	NodeTypeRepository NodeType = "repository"
	NodeTypeEvent      NodeType = "event"
)

// GraphNode is the interface that all graph nodes must implement.
type GraphNode interface {
	// ID returns the unique identifier of the node.
	ID() string

	// NodeType returns the type of the node.
	NodeType() NodeType

	// Record returns the original GEDCOM record (if applicable).
	// Returns nil for EventNode which doesn't have a top-level record.
	Record() types.Record

	// InEdges returns all edges pointing TO this node.
	InEdges() []*Edge

	// OutEdges returns all edges pointing FROM this node.
	OutEdges() []*Edge

	// AddInEdge adds an incoming edge to this node.
	AddInEdge(*Edge)

	// AddOutEdge adds an outgoing edge from this node.
	AddOutEdge(*Edge)

	// RemoveInEdge removes an incoming edge from this node.
	RemoveInEdge(*Edge)

	// RemoveOutEdge removes an outgoing edge from this node.
	RemoveOutEdge(*Edge)

	// Neighbors returns all nodes connected to this node (via in or out edges).
	Neighbors() []GraphNode

	// Degree returns the total number of edges (in + out).
	Degree() int

	// InDegree returns the number of incoming edges.
	InDegree() int

	// OutDegree returns the number of outgoing edges.
	OutDegree() int
}

// BaseNode provides common functionality for all graph nodes.
type BaseNode struct {
	xrefID   string
	nodeID   uint32 // Phase 3: Cached uint32 ID for fast access (eliminates GetNodeID() overhead)
	nodeType NodeType
	record   types.Record
	inEdges  []*Edge
	outEdges []*Edge
}

// ID returns the unique identifier of the node.
func (bn *BaseNode) ID() string {
	return bn.xrefID
}

// NodeType returns the type of the node.
func (bn *BaseNode) NodeType() NodeType {
	return bn.nodeType
}

// Record returns the original GEDCOM record.
func (bn *BaseNode) Record() types.Record {
	return bn.record
}

// InEdges returns all incoming edges.
func (bn *BaseNode) InEdges() []*Edge {
	return bn.inEdges
}

// OutEdges returns all outgoing edges.
// If lazy mode is enabled, triggers edge loading if not already loaded.
func (bn *BaseNode) OutEdges() []*Edge {
	// Note: Edge loading is handled by Graph.ensureEdgesLoaded()
	// which is called before accessing edges
	return bn.outEdges
}

// AddInEdge adds an incoming edge.
func (bn *BaseNode) AddInEdge(edge *Edge) {
	bn.inEdges = append(bn.inEdges, edge)
}

// AddOutEdge adds an outgoing edge.
func (bn *BaseNode) AddOutEdge(edge *Edge) {
	bn.outEdges = append(bn.outEdges, edge)
}

// RemoveInEdge removes an incoming edge.
func (bn *BaseNode) RemoveInEdge(edge *Edge) {
	for i, e := range bn.inEdges {
		if e.ID == edge.ID {
			// Remove by swapping with last element and truncating
			bn.inEdges[i] = bn.inEdges[len(bn.inEdges)-1]
			bn.inEdges = bn.inEdges[:len(bn.inEdges)-1]
			return
		}
	}
}

// RemoveOutEdge removes an outgoing edge.
func (bn *BaseNode) RemoveOutEdge(edge *Edge) {
	for i, e := range bn.outEdges {
		if e.ID == edge.ID {
			// Remove by swapping with last element and truncating
			bn.outEdges[i] = bn.outEdges[len(bn.outEdges)-1]
			bn.outEdges = bn.outEdges[:len(bn.outEdges)-1]
			return
		}
	}
}

// Neighbors returns all nodes connected to this node.
func (bn *BaseNode) Neighbors() []GraphNode {
	neighbors := make([]GraphNode, 0)
	neighborMap := make(map[string]GraphNode)

	// Add nodes from incoming edges
	for _, edge := range bn.inEdges {
		if edge.From != nil {
			if _, exists := neighborMap[edge.From.ID()]; !exists {
				neighborMap[edge.From.ID()] = edge.From
				neighbors = append(neighbors, edge.From)
			}
		}
	}

	// Add nodes from outgoing edges
	for _, edge := range bn.outEdges {
		if edge.To != nil {
			if _, exists := neighborMap[edge.To.ID()]; !exists {
				neighborMap[edge.To.ID()] = edge.To
				neighbors = append(neighbors, edge.To)
			}
		}
	}

	return neighbors
}

// Degree returns the total number of edges.
func (bn *BaseNode) Degree() int {
	return len(bn.inEdges) + len(bn.outEdges)
}

// InDegree returns the number of incoming edges.
func (bn *BaseNode) InDegree() int {
	return len(bn.inEdges)
}

// OutDegree returns the number of outgoing edges.
func (bn *BaseNode) OutDegree() int {
	return len(bn.outEdges)
}

// IndividualNode represents an individual person in the genealogy.
type IndividualNode struct {
	*BaseNode
	Individual *types.IndividualRecord

	// Phase 1: Indexed edges for fast access (populated during graph construction)
	famcEdges []*Edge // Only FAMC edges (parent families) - for fast ancestor queries
	famsEdges []*Edge // Only FAMS edges (spouse families) - for fast spouse/child queries

	// Phase 2: Cached parents for O(1) access (populated during graph construction)
	parents []*IndividualNode // Direct parent references - eliminates edge traversal

	// Note: Relationships (Parents, Children, Spouses, Siblings) are now computed
	// on-demand from edges to save memory. Use helper methods or query API.
	// With optimizations, parents are cached for fast access.
}

// NewIndividualNode creates a new IndividualNode.
func NewIndividualNode(xrefID string, record *types.IndividualRecord) *IndividualNode {
	var recordInterface types.Record = record
	return &IndividualNode{
		BaseNode: &BaseNode{
			xrefID:   xrefID,
			nodeType: NodeTypeIndividual,
			record:   recordInterface,
			inEdges:  make([]*Edge, 0),
			outEdges: make([]*Edge, 0),
		},
		Individual: record,
		famcEdges:  make([]*Edge, 0),
		famsEdges:  make([]*Edge, 0),
		parents:    make([]*IndividualNode, 0),
	}
}

// FamilyNode represents a family unit (marriage/partnership and children).
type FamilyNode struct {
	*BaseNode
	Family *types.FamilyRecord

	// Phase 1: Indexed edges for fast access (populated during graph construction)
	husbandEdge *Edge      // Only HUSB edge (or nil) - O(1) access instead of O(n) scan
	wifeEdge    *Edge      // Only WIFE edge (or nil) - O(1) access instead of O(n) scan
	chilEdges   []*Edge    // Only CHIL edges - eliminates filtering

	// Note: Relationships (Husband, Wife, Children) are now computed
	// on-demand from edges to save memory. Use helper methods or query API.
	// With optimizations, husband/wife are directly accessible via indexed edges.
}

// NewFamilyNode creates a new FamilyNode.
func NewFamilyNode(xrefID string, record *types.FamilyRecord) *FamilyNode {
	var recordInterface types.Record = record
	return &FamilyNode{
		BaseNode: &BaseNode{
			xrefID:   xrefID,
			nodeType: NodeTypeFamily,
			record:   recordInterface,
			inEdges:  make([]*Edge, 0),
			outEdges: make([]*Edge, 0),
		},
		Family:      record,
		husbandEdge: nil,
		wifeEdge:    nil,
		chilEdges:   make([]*Edge, 0),
	}
}

// NoteNode represents a note record.
type NoteNode struct {
	*BaseNode
	Note *types.NoteRecord

	// Cached relationships
	ReferencedBy []GraphNode
}

// NewNoteNode creates a new NoteNode.
func NewNoteNode(xrefID string, record *types.NoteRecord) *NoteNode {
	var recordInterface types.Record = record
	return &NoteNode{
		BaseNode: &BaseNode{
			xrefID:   xrefID,
			nodeType: NodeTypeNote,
			record:   recordInterface,
			inEdges:  make([]*Edge, 0),
			outEdges: make([]*Edge, 0),
		},
		Note:         record,
		ReferencedBy: make([]GraphNode, 0),
	}
}

// SourceNode represents a source citation record.
type SourceNode struct {
	*BaseNode
	Source *types.SourceRecord

	// Cached relationships
	ReferencedBy []GraphNode
	Repository   *RepositoryNode
}

// NewSourceNode creates a new SourceNode.
func NewSourceNode(xrefID string, record *types.SourceRecord) *SourceNode {
	var recordInterface types.Record = record
	return &SourceNode{
		BaseNode: &BaseNode{
			xrefID:   xrefID,
			nodeType: NodeTypeSource,
			record:   recordInterface,
			inEdges:  make([]*Edge, 0),
			outEdges: make([]*Edge, 0),
		},
		Source:       record,
		ReferencedBy: make([]GraphNode, 0),
	}
}

// RepositoryNode represents a repository where sources are stored.
type RepositoryNode struct {
	*BaseNode
	Repository *types.RepositoryRecord

	// Cached relationships
	Sources []*SourceNode
}

// NewRepositoryNode creates a new RepositoryNode.
func NewRepositoryNode(xrefID string, record *types.RepositoryRecord) *RepositoryNode {
	var recordInterface types.Record = record
	return &RepositoryNode{
		BaseNode: &BaseNode{
			xrefID:   xrefID,
			nodeType: NodeTypeRepository,
			record:   recordInterface,
			inEdges:  make([]*Edge, 0),
			outEdges: make([]*Edge, 0),
		},
		Repository: record,
		Sources:    make([]*SourceNode, 0),
	}
}

// EventNode represents an event embedded within an Individual or Family record.
type EventNode struct {
	*BaseNode
	EventID   string
	EventType string
	EventData map[string]interface{}

	// Cached relationships
	Owner   GraphNode
	Sources []*SourceNode
	Notes   []*NoteNode
}

// NewEventNode creates a new EventNode.
func NewEventNode(eventID string, eventType string, eventData map[string]interface{}) *EventNode {
	return &EventNode{
		BaseNode: &BaseNode{
			xrefID:   eventID,
			nodeType: NodeTypeEvent,
			record:   nil, // Events don't have top-level records
			inEdges:  make([]*Edge, 0),
			outEdges: make([]*Edge, 0),
		},
		EventID:   eventID,
		EventType: eventType,
		EventData: eventData,
		Sources:   make([]*SourceNode, 0),
		Notes:     make([]*NoteNode, 0),
	}
}

// Record returns nil for EventNode (events don't have top-level records).
func (en *EventNode) Record() types.Record {
	return nil
}
