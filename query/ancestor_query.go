package query

import (
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// AncestorOptions holds configuration for ancestor queries.
type AncestorOptions struct {
	MaxGenerations int                                 // Limit depth (0 = unlimited)
	IncludeSelf    bool                                // Include starting individual
	Filter         func(*types.IndividualRecord) bool // Custom filter function
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
func (aq *AncestorQuery) Filter(fn func(*types.IndividualRecord) bool) *AncestorQuery {
	aq.options.Filter = fn
	return aq
}

// Execute runs the query and returns ancestor records.
func (aq *AncestorQuery) Execute() ([]*types.IndividualRecord, error) {
	// Record metrics if available
	start := time.Now()
	defer func() {
		if aq.graph.metrics != nil {
			duration := time.Since(start)
			aq.graph.metrics.RecordQuery(duration)
		}
	}()

	startNode := aq.graph.GetIndividual(aq.startXrefID)
	if startNode == nil {
		return nil, nil
	}

	// Phase 3: Use cached nodeID directly (eliminates GetNodeID() lock overhead)
	startNodeID := startNode.BaseNode.nodeID
	if startNodeID == 0 {
		return nil, nil
	}

	ancestors := make(map[uint32]*IndividualNode)
	visited := make(map[uint32]bool)

	// Add self if requested
	if aq.options.IncludeSelf {
		ancestors[startNodeID] = startNode
	}

	// Find ancestors recursively (pass nodeID to avoid repeated lookups)
	aq.findAncestors(startNode, startNodeID, ancestors, visited, 0)

	// Convert to records
	records := make([]*types.IndividualRecord, 0, len(ancestors))
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
// Optimized with Phase 1 (indexed edges, uint32 IDs), Phase 2 (cached parents), and Phase 3 (cached nodeID).
// Phase 3: Accepts nodeID parameter to eliminate repeated GetNodeID() calls.
func (aq *AncestorQuery) findAncestors(node *IndividualNode, nodeID uint32, ancestors map[uint32]*IndividualNode, visited map[uint32]bool, depth int) {
	// Phase 3: nodeID already provided - no lookup needed!
	if nodeID == 0 || visited[nodeID] {
		return
	}

	// Check max generations limit
	if aq.options.MaxGenerations > 0 && depth >= aq.options.MaxGenerations {
		return
	}

	visited[nodeID] = true

	// Phase 2: Use cached parents for O(1) access (fastest path)
	if len(node.parents) > 0 {
		for _, parent := range node.parents {
			// Phase 3: Use cached nodeID directly - no lock acquisition!
			parentID := parent.BaseNode.nodeID
			if parentID != 0 {
				ancestors[parentID] = parent
				// Phase 3: Pass parentID through recursion to avoid repeated lookups
				aq.findAncestors(parent, parentID, ancestors, visited, depth+1)
			}
		}
		return
	}

	// Fallback: Use indexed FAMC edges (Phase 1 optimization)
	// This path is used if parent cache is not populated (shouldn't happen in normal flow)
	for _, edge := range node.famcEdges {
		if edge.Family != nil {
			famNode := edge.Family
			// Phase 1: Use indexed edges for O(1) access
			if famNode.husbandEdge != nil {
				if husband, ok := famNode.husbandEdge.To.(*IndividualNode); ok {
					// Phase 3: Use cached nodeID directly - no lock acquisition!
					husbandID := husband.BaseNode.nodeID
					if husbandID != 0 {
						ancestors[husbandID] = husband
						// Phase 3: Pass husbandID through recursion
						aq.findAncestors(husband, husbandID, ancestors, visited, depth+1)
					}
				}
			}
			if famNode.wifeEdge != nil {
				if wife, ok := famNode.wifeEdge.To.(*IndividualNode); ok {
					// Phase 3: Use cached nodeID directly - no lock acquisition!
					wifeID := wife.BaseNode.nodeID
					if wifeID != 0 {
						ancestors[wifeID] = wife
						// Phase 3: Pass wifeID through recursion
						aq.findAncestors(wife, wifeID, ancestors, visited, depth+1)
					}
				}
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
	Ancestor *types.IndividualRecord
	Path     *Path
	Depth    int
}

// ExecuteWithPaths returns ancestors with path information.
func (aq *AncestorQuery) ExecuteWithPaths() ([]*AncestorPath, error) {
	startNode := aq.graph.GetIndividual(aq.startXrefID)
	if startNode == nil {
		return nil, nil
	}

	// Phase 3: Use cached nodeID directly (eliminates GetNodeID() lock overhead)
	startNodeID := startNode.BaseNode.nodeID
	if startNodeID == 0 {
		return nil, nil
	}

	ancestors := make(map[uint32]*IndividualNode)
	visited := make(map[uint32]bool)
	depths := make(map[uint32]int)

	// Add self if requested
	if aq.options.IncludeSelf {
		ancestors[startNodeID] = startNode
		depths[startNodeID] = 0
	}

	// Find ancestors with depth tracking (pass nodeID to avoid repeated lookups)
	aq.findAncestorsWithDepth(startNode, startNodeID, ancestors, visited, depths, 0)

	// Build paths and convert to AncestorPath
	result := make([]*AncestorPath, 0, len(ancestors))
	for id, node := range ancestors {
		if node.Individual != nil {
			// Apply filter if provided
			if aq.options.Filter == nil || aq.options.Filter(node.Individual) {
				// Convert uint32 ID back to XREF for path finding
				xrefID := aq.graph.GetXrefFromID(id)
				if xrefID != "" {
					// Find path to this ancestor
					path, err := aq.graph.ShortestPath(aq.startXrefID, xrefID)
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
	}

	return result, nil
}

// findAncestorsWithDepth recursively finds ancestors with depth tracking.
// Optimized with Phase 1 (indexed edges, uint32 IDs), Phase 2 (cached parents), and Phase 3 (cached nodeID).
// Phase 3: Accepts nodeID parameter to eliminate repeated GetNodeID() calls.
func (aq *AncestorQuery) findAncestorsWithDepth(node *IndividualNode, nodeID uint32, ancestors map[uint32]*IndividualNode, visited map[uint32]bool, depths map[uint32]int, depth int) {
	// Phase 3: nodeID already provided - no lookup needed!
	if nodeID == 0 || visited[nodeID] {
		return
	}

	// Check max generations limit
	if aq.options.MaxGenerations > 0 && depth >= aq.options.MaxGenerations {
		return
	}

	visited[nodeID] = true

	// Phase 2: Use cached parents for O(1) access (fastest path)
	if len(node.parents) > 0 {
		for _, parent := range node.parents {
			// Phase 3: Use cached nodeID directly - no lock acquisition!
			parentID := parent.BaseNode.nodeID
			if parentID != 0 {
				ancestors[parentID] = parent
				depths[parentID] = depth + 1
				// Phase 3: Pass parentID through recursion to avoid repeated lookups
				aq.findAncestorsWithDepth(parent, parentID, ancestors, visited, depths, depth+1)
			}
		}
		return
	}

	// Fallback: Use indexed FAMC edges (Phase 1 optimization)
	for _, edge := range node.famcEdges {
		if edge.Family != nil {
			famNode := edge.Family
			// Phase 1: Use indexed edges for O(1) access
			if famNode.husbandEdge != nil {
				if husband, ok := famNode.husbandEdge.To.(*IndividualNode); ok {
					// Phase 3: Use cached nodeID directly - no lock acquisition!
					husbandID := husband.BaseNode.nodeID
					if husbandID != 0 {
						ancestors[husbandID] = husband
						depths[husbandID] = depth + 1
						// Phase 3: Pass husbandID through recursion
						aq.findAncestorsWithDepth(husband, husbandID, ancestors, visited, depths, depth+1)
					}
				}
			}
			if famNode.wifeEdge != nil {
				if wife, ok := famNode.wifeEdge.To.(*IndividualNode); ok {
					// Phase 3: Use cached nodeID directly - no lock acquisition!
					wifeID := wife.BaseNode.nodeID
					if wifeID != 0 {
						ancestors[wifeID] = wife
						depths[wifeID] = depth + 1
						// Phase 3: Pass wifeID through recursion
						aq.findAncestorsWithDepth(wife, wifeID, ancestors, visited, depths, depth+1)
					}
				}
			}
		}
	}
}
