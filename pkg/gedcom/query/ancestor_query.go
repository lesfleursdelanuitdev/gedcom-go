package query

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// AncestorOptions holds configuration for ancestor queries.
type AncestorOptions struct {
	MaxGenerations int                                 // Limit depth (0 = unlimited)
	IncludeSelf    bool                                // Include starting individual
	Filter         func(*gedcom.IndividualRecord) bool // Custom filter function
	Order          Order                               // BFS or DFS order
}

// Order represents the traversal order.
type Order string

const (
	OrderBFS Order = "BFS" // Breadth-first search
	OrderDFS Order = "DFS" // Depth-first search
)

// NewAncestorOptions creates new AncestorOptions with defaults.
func NewAncestorOptions() *AncestorOptions {
	return &AncestorOptions{
		MaxGenerations: 0, // Unlimited
		IncludeSelf:    false,
		Filter:         nil,
		Order:          OrderBFS,
	}
}

// AncestorQuery represents a query for ancestors.
type AncestorQuery struct {
	startXrefID string
	graph       *Graph
	options     *AncestorOptions
}

// MaxGenerations limits the depth of the ancestor search.
func (aq *AncestorQuery) MaxGenerations(n int) *AncestorQuery {
	aq.options.MaxGenerations = n
	return aq
}

// IncludeSelf includes the starting individual in results.
func (aq *AncestorQuery) IncludeSelf() *AncestorQuery {
	aq.options.IncludeSelf = true
	return aq
}

// Filter applies a custom filter function to results.
func (aq *AncestorQuery) Filter(fn func(*gedcom.IndividualRecord) bool) *AncestorQuery {
	aq.options.Filter = fn
	return aq
}

// Execute runs the query and returns ancestor records.
func (aq *AncestorQuery) Execute() ([]*gedcom.IndividualRecord, error) {
	startNode := aq.graph.GetIndividual(aq.startXrefID)
	if startNode == nil {
		return nil, nil
	}

	ancestors := make(map[string]*IndividualNode)
	visited := make(map[string]bool)

	// Add self if requested
	if aq.options.IncludeSelf {
		ancestors[startNode.ID()] = startNode
	}

	// Find ancestors recursively
	aq.findAncestors(startNode, ancestors, visited, 0)

	// Convert to records
	records := make([]*gedcom.IndividualRecord, 0, len(ancestors))
	for _, node := range ancestors {
		if node.Individual != nil {
			// Apply filter if provided
			if aq.options.Filter == nil || aq.options.Filter(node.Individual) {
				records = append(records, node.Individual)
			}
		}
	}

	return records, nil
}

// findAncestors recursively finds ancestors.
func (aq *AncestorQuery) findAncestors(node *IndividualNode, ancestors map[string]*IndividualNode, visited map[string]bool, depth int) {
	if visited[node.ID()] {
		return
	}

	// Check max generations limit
	if aq.options.MaxGenerations > 0 && depth >= aq.options.MaxGenerations {
		return
	}

	visited[node.ID()] = true

	// Find parents via FAMC edges
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
			famNode := edge.Family
			if famNode.Husband != nil {
				ancestors[famNode.Husband.ID()] = famNode.Husband
				aq.findAncestors(famNode.Husband, ancestors, visited, depth+1)
			}
			if famNode.Wife != nil {
				ancestors[famNode.Wife.ID()] = famNode.Wife
				aq.findAncestors(famNode.Wife, ancestors, visited, depth+1)
			}
		}
	}
}

// Count returns the number of ancestors.
func (aq *AncestorQuery) Count() (int, error) {
	ancestors, err := aq.Execute()
	if err != nil {
		return 0, err
	}
	return len(ancestors), nil
}

// Exists checks if any ancestors exist.
func (aq *AncestorQuery) Exists() (bool, error) {
	count, err := aq.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AncestorPath represents an ancestor with path information.
type AncestorPath struct {
	Ancestor *gedcom.IndividualRecord
	Path     *Path
	Depth    int
}

// ExecuteWithPaths returns ancestors with path information.
func (aq *AncestorQuery) ExecuteWithPaths() ([]*AncestorPath, error) {
	startNode := aq.graph.GetIndividual(aq.startXrefID)
	if startNode == nil {
		return nil, nil
	}

	ancestors := make(map[string]*IndividualNode)
	visited := make(map[string]bool)
	depths := make(map[string]int)

	// Add self if requested
	if aq.options.IncludeSelf {
		ancestors[startNode.ID()] = startNode
		depths[startNode.ID()] = 0
	}

	// Find ancestors with depth tracking
	aq.findAncestorsWithDepth(startNode, ancestors, visited, depths, 0)

	// Build paths and convert to AncestorPath
	result := make([]*AncestorPath, 0, len(ancestors))
	for id, node := range ancestors {
		if node.Individual != nil {
			// Apply filter if provided
			if aq.options.Filter == nil || aq.options.Filter(node.Individual) {
				// Find path to this ancestor
				path, err := aq.graph.ShortestPath(aq.startXrefID, id)
				if err == nil {
					result = append(result, &AncestorPath{
						Ancestor: node.Individual,
						Path:     path,
						Depth:    depths[id],
					})
				}
			}
		}
	}

	return result, nil
}

// findAncestorsWithDepth recursively finds ancestors with depth tracking.
func (aq *AncestorQuery) findAncestorsWithDepth(node *IndividualNode, ancestors map[string]*IndividualNode, visited map[string]bool, depths map[string]int, depth int) {
	if visited[node.ID()] {
		return
	}

	// Check max generations limit
	if aq.options.MaxGenerations > 0 && depth >= aq.options.MaxGenerations {
		return
	}

	visited[node.ID()] = true

	// Find parents via FAMC edges
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
			famNode := edge.Family
			if famNode.Husband != nil {
				ancestors[famNode.Husband.ID()] = famNode.Husband
				depths[famNode.Husband.ID()] = depth + 1
				aq.findAncestorsWithDepth(famNode.Husband, ancestors, visited, depths, depth+1)
			}
			if famNode.Wife != nil {
				ancestors[famNode.Wife.ID()] = famNode.Wife
				depths[famNode.Wife.ID()] = depth + 1
				aq.findAncestorsWithDepth(famNode.Wife, ancestors, visited, depths, depth+1)
			}
		}
	}
}
