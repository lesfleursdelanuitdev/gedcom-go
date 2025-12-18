package query

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
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
	Record() gedcom.Record

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
	nodeType NodeType
	record   gedcom.Record
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
func (bn *BaseNode) Record() gedcom.Record {
	return bn.record
}

// InEdges returns all incoming edges.
func (bn *BaseNode) InEdges() []*Edge {
	return bn.inEdges
}

// OutEdges returns all outgoing edges.
func (bn *BaseNode) OutEdges() []*Edge {
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
	Individual *gedcom.IndividualRecord

	// Cached relationships (for performance)
	Parents  []*IndividualNode
	Children []*IndividualNode
	Spouses  []*IndividualNode
	Siblings []*IndividualNode
}

// NewIndividualNode creates a new IndividualNode.
func NewIndividualNode(xrefID string, record *gedcom.IndividualRecord) *IndividualNode {
	var recordInterface gedcom.Record = record
	return &IndividualNode{
		BaseNode: &BaseNode{
			xrefID:   xrefID,
			nodeType: NodeTypeIndividual,
			record:   recordInterface,
			inEdges:  make([]*Edge, 0),
			outEdges: make([]*Edge, 0),
		},
		Individual: record,
		Parents:    make([]*IndividualNode, 0),
		Children:   make([]*IndividualNode, 0),
		Spouses:    make([]*IndividualNode, 0),
		Siblings:   make([]*IndividualNode, 0),
	}
}

// FamilyNode represents a family unit (marriage/partnership and children).
type FamilyNode struct {
	*BaseNode
	Family *gedcom.FamilyRecord

	// Cached relationships (for performance)
	Husband  *IndividualNode
	Wife     *IndividualNode
	Children []*IndividualNode
}

// NewFamilyNode creates a new FamilyNode.
func NewFamilyNode(xrefID string, record *gedcom.FamilyRecord) *FamilyNode {
	var recordInterface gedcom.Record = record
	return &FamilyNode{
		BaseNode: &BaseNode{
			xrefID:   xrefID,
			nodeType: NodeTypeFamily,
			record:   recordInterface,
			inEdges:  make([]*Edge, 0),
			outEdges: make([]*Edge, 0),
		},
		Family:   record,
		Children: make([]*IndividualNode, 0),
	}
}

// NoteNode represents a note record.
type NoteNode struct {
	*BaseNode
	Note *gedcom.NoteRecord

	// Cached relationships
	ReferencedBy []GraphNode
}

// NewNoteNode creates a new NoteNode.
func NewNoteNode(xrefID string, record *gedcom.NoteRecord) *NoteNode {
	var recordInterface gedcom.Record = record
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
	Source *gedcom.SourceRecord

	// Cached relationships
	ReferencedBy []GraphNode
	Repository   *RepositoryNode
}

// NewSourceNode creates a new SourceNode.
func NewSourceNode(xrefID string, record *gedcom.SourceRecord) *SourceNode {
	var recordInterface gedcom.Record = record
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
	Repository *gedcom.RepositoryRecord

	// Cached relationships
	Sources []*SourceNode
}

// NewRepositoryNode creates a new RepositoryNode.
func NewRepositoryNode(xrefID string, record *gedcom.RepositoryRecord) *RepositoryNode {
	var recordInterface gedcom.Record = record
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
func (en *EventNode) Record() gedcom.Record {
	return nil
}
