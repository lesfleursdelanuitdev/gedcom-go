package query

import (
	"github.com/yourorg/gedcom/pkg/gedcom"
)

// DescendantOptions holds configuration for descendant queries.
type DescendantOptions struct {
	MaxGenerations int                                 // Limit depth (0 = unlimited)
	IncludeSelf    bool                                // Include starting individual
	Filter         func(*gedcom.IndividualRecord) bool // Custom filter function
	Order          Order                               // BFS or DFS order
}

// NewDescendantOptions creates new DescendantOptions with defaults.
func NewDescendantOptions() *DescendantOptions {
	return &DescendantOptions{
		MaxGenerations: 0, // Unlimited
		IncludeSelf:    false,
		Filter:         nil,
		Order:          OrderBFS,
	}
}

// DescendantQuery represents a query for descendants.
type DescendantQuery struct {
	startXrefID string
	graph       *Graph
	options     *DescendantOptions
}

// MaxGenerations limits the depth of the descendant search.
func (dq *DescendantQuery) MaxGenerations(n int) *DescendantQuery {
	dq.options.MaxGenerations = n
	return dq
}

// IncludeSelf includes the starting individual in results.
func (dq *DescendantQuery) IncludeSelf() *DescendantQuery {
	dq.options.IncludeSelf = true
	return dq
}

// Filter applies a custom filter function to results.
func (dq *DescendantQuery) Filter(fn func(*gedcom.IndividualRecord) bool) *DescendantQuery {
	dq.options.Filter = fn
	return dq
}

// Execute runs the query and returns descendant records.
func (dq *DescendantQuery) Execute() ([]*gedcom.IndividualRecord, error) {
	startNode := dq.graph.GetIndividual(dq.startXrefID)
	if startNode == nil {
		return nil, nil
	}

	descendants := make(map[string]*IndividualNode)
	visited := make(map[string]bool)

	// Add self if requested
	if dq.options.IncludeSelf {
		descendants[startNode.ID()] = startNode
	}

	// Find descendants recursively
	dq.findDescendants(startNode, descendants, visited, 0)

	// Convert to records
	records := make([]*gedcom.IndividualRecord, 0, len(descendants))
	for _, node := range descendants {
		if node.Individual != nil {
			// Apply filter if provided
			if dq.options.Filter == nil || dq.options.Filter(node.Individual) {
				records = append(records, node.Individual)
			}
		}
	}

	return records, nil
}

// findDescendants recursively finds descendants.
func (dq *DescendantQuery) findDescendants(node *IndividualNode, descendants map[string]*IndividualNode, visited map[string]bool, depth int) {
	if visited[node.ID()] {
		return
	}

	// Check max generations limit
	if dq.options.MaxGenerations > 0 && depth >= dq.options.MaxGenerations {
		return
	}

	visited[node.ID()] = true

	// Find children via FAMS -> Family -> CHIL edges
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
			famNode := edge.Family
			for _, childNode := range famNode.Children {
				descendants[childNode.ID()] = childNode
				dq.findDescendants(childNode, descendants, visited, depth+1)
			}
		}
	}
}

// Count returns the number of descendants.
func (dq *DescendantQuery) Count() (int, error) {
	descendants, err := dq.Execute()
	if err != nil {
		return 0, err
	}
	return len(descendants), nil
}

// Exists checks if any descendants exist.
func (dq *DescendantQuery) Exists() (bool, error) {
	count, err := dq.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
