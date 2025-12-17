package query

import (
	"fmt"
)

// ShortestPath finds the shortest path between two nodes using bidirectional BFS.
func (g *Graph) ShortestPath(fromID, toID string) (*Path, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	fromNode := g.nodes[fromID]
	toNode := g.nodes[toID]

	if fromNode == nil {
		return nil, fmt.Errorf("node %s not found", fromID)
	}
	if toNode == nil {
		return nil, fmt.Errorf("node %s not found", toID)
	}

	if fromID == toID {
		return &Path{
			Nodes:  []GraphNode{fromNode},
			Edges:  []*Edge{},
			Length: 0,
			Type:   PathTypeBlood,
		}, nil
	}

	// Use memory pools
	visitedFrom := getVisitedMap()
	visitedTo := getVisitedMap()
	parentFrom := getParentMap()
	parentTo := getParentMap()
	queueFrom := getQueue()
	queueTo := getQueue()

	defer func() {
		putVisitedMap(visitedFrom)
		putVisitedMap(visitedTo)
		putParentMap(parentFrom)
		putParentMap(parentTo)
		putQueue(queueFrom)
		putQueue(queueTo)
	}()

	// Initialize bidirectional BFS
	queueFrom = append(queueFrom, fromNode)
	queueTo = append(queueTo, toNode)
	visitedFrom[fromID] = true
	visitedTo[toID] = true

	// Helper function to get neighbors with edges
	getNeighborsWithEdges := func(node GraphNode) []struct {
		node GraphNode
		edge *Edge
	} {
		neighbors := make([]struct {
			node GraphNode
			edge *Edge
		}, 0)
		for _, edge := range node.OutEdges() {
			if edge.To != nil {
				neighbors = append(neighbors, struct {
					node GraphNode
					edge *Edge
				}{edge.To, edge})
			}
		}
		for _, edge := range node.InEdges() {
			if edge.IsBidirectional() && edge.From != nil {
				neighbors = append(neighbors, struct {
					node GraphNode
					edge *Edge
				}{edge.From, edge})
			}
		}
		return neighbors
	}

	// Bidirectional BFS
	for len(queueFrom) > 0 || len(queueTo) > 0 {
		// Expand from forward direction
		if len(queueFrom) > 0 {
			current := queueFrom[0]
			queueFrom = queueFrom[1:]

			for _, neighborData := range getNeighborsWithEdges(current) {
				neighbor := neighborData.node
				edge := neighborData.edge
				neighborID := neighbor.ID()

				// Check if we've met in the middle
				if visitedTo[neighborID] {
					// Found meeting point, reconstruct path
					path := reconstructBidirectionalPath(
						current, neighbor, parentFrom, parentTo, fromNode, toNode,
					)
					// Determine path type using graph
					path.Type = g.determinePathType(path.Edges)
					return path, nil
				}

				if !visitedFrom[neighborID] {
					visitedFrom[neighborID] = true
					parentFrom[neighborID] = struct {
						node GraphNode
						edge *Edge
					}{current, edge}
					queueFrom = append(queueFrom, neighbor)
				}
			}
		}

		// Expand from backward direction
		if len(queueTo) > 0 {
			current := queueTo[0]
			queueTo = queueTo[1:]

			for _, neighborData := range getNeighborsWithEdges(current) {
				neighbor := neighborData.node
				edge := neighborData.edge
				neighborID := neighbor.ID()

				// Check if we've met in the middle
				if visitedFrom[neighborID] {
					// Found meeting point, reconstruct path
					path := reconstructBidirectionalPath(
						neighbor, current, parentFrom, parentTo, fromNode, toNode,
					)
					// Determine path type using graph
					path.Type = g.determinePathType(path.Edges)
					return path, nil
				}

				if !visitedTo[neighborID] {
					visitedTo[neighborID] = true
					parentTo[neighborID] = struct {
						node GraphNode
						edge *Edge
					}{current, edge}
					queueTo = append(queueTo, neighbor)
				}
			}
		}
	}

	// No path found
	return nil, fmt.Errorf("no path found from %s to %s", fromID, toID)
}

// reconstructBidirectionalPath reconstructs the path from bidirectional BFS.
func reconstructBidirectionalPath(
	meetingFrom, meetingTo GraphNode,
	parentFrom, parentTo map[string]struct {
		node GraphNode
		edge *Edge
	},
	fromNode, toNode GraphNode,
) *Path {
	path := &Path{
		Nodes:  make([]GraphNode, 0),
		Edges:  make([]*Edge, 0),
		Length: 0,
		Type:   PathTypeBlood,
	}

	// Build path from start to meeting point
	pathFromStart := make([]GraphNode, 0)
	current := meetingFrom
	for current != nil && current.ID() != fromNode.ID() {
		pathFromStart = append([]GraphNode{current}, pathFromStart...)
		if p, ok := parentFrom[current.ID()]; ok {
			current = p.node
		} else {
			break
		}
	}
	pathFromStart = append([]GraphNode{fromNode}, pathFromStart...)

	// Build path from meeting point to end
	pathToEnd := make([]GraphNode, 0)
	current = meetingTo
	for current != nil && current.ID() != toNode.ID() {
		pathToEnd = append(pathToEnd, current)
		if p, ok := parentTo[current.ID()]; ok {
			current = p.node
		} else {
			break
		}
	}
	pathToEnd = append(pathToEnd, toNode)

	// Combine paths (avoid duplicate meeting point)
	if len(pathToEnd) > 0 && pathToEnd[0].ID() == pathFromStart[len(pathFromStart)-1].ID() {
		pathToEnd = pathToEnd[1:]
	}
	path.Nodes = append(pathFromStart, pathToEnd...)
	path.Length = len(path.Nodes) - 1

	// Reconstruct edges from nodes
	path.Edges = make([]*Edge, 0, path.Length)
	for i := 0; i < len(path.Nodes)-1; i++ {
		current := path.Nodes[i]
		next := path.Nodes[i+1]
		// Find edge between current and next
		for _, edge := range current.OutEdges() {
			if edge.To != nil && edge.To.ID() == next.ID() {
				path.Edges = append(path.Edges, edge)
				break
			}
		}
		// Also check bidirectional edges
		if len(path.Edges) <= i {
			for _, edge := range current.InEdges() {
				if edge.IsBidirectional() && edge.From != nil && edge.From.ID() == next.ID() {
					path.Edges = append(path.Edges, edge)
					break
				}
			}
		}
	}

	return path
}

// AllPaths finds all paths between two nodes using DFS.
// maxLength limits the maximum path length to avoid infinite loops.
func (g *Graph) AllPaths(fromID, toID string, maxLength int) ([]*Path, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	fromNode := g.nodes[fromID]
	toNode := g.nodes[toID]

	if fromNode == nil {
		return nil, fmt.Errorf("node %s not found", fromID)
	}
	if toNode == nil {
		return nil, fmt.Errorf("node %s not found", toID)
	}

	if maxLength <= 0 {
		maxLength = 10 // Default maximum
	}

	paths := make([]*Path, 0)
	currentPath := make([]GraphNode, 0)
	currentEdges := make([]*Edge, 0)
	visited := make(map[string]bool)

	g.allPathsDFS(fromNode, toNode, currentPath, currentEdges, visited, &paths, maxLength)

	return paths, nil
}

// allPathsDFS is the recursive helper for AllPaths.
func (g *Graph) allPathsDFS(current, target GraphNode, currentPath []GraphNode, currentEdges []*Edge, visited map[string]bool, paths *[]*Path, maxLength int) {
	if len(currentPath) > maxLength {
		return
	}

	currentPath = append(currentPath, current)
	visited[current.ID()] = true

	if current.ID() == target.ID() {
		// Found a path
		path := &Path{
			Nodes:  make([]GraphNode, len(currentPath)),
			Edges:  make([]*Edge, len(currentEdges)),
			Length: len(currentEdges),
			Type:   PathTypeBlood,
		}
		copy(path.Nodes, currentPath)
		copy(path.Edges, currentEdges)
		path.Type = g.determinePathType(path.Edges)
		*paths = append(*paths, path)
	} else {
		// Continue searching
		for _, edge := range current.OutEdges() {
			neighbor := edge.To
			if neighbor != nil && !visited[neighbor.ID()] {
				g.allPathsDFS(neighbor, target, currentPath, append(currentEdges, edge), visited, paths, maxLength)
			}
		}

		// Check bidirectional edges
		for _, edge := range current.InEdges() {
			if edge.IsBidirectional() {
				neighbor := edge.From
				if neighbor != nil && !visited[neighbor.ID()] {
					g.allPathsDFS(neighbor, target, currentPath, append(currentEdges, edge), visited, paths, maxLength)
				}
			}
		}
	}

	// Backtrack
	visited[current.ID()] = false
}

// determinePathType determines the type of path based on edge types.
func (g *Graph) determinePathType(edges []*Edge) PathType {
	hasBlood := false
	hasMarital := false

	for _, edge := range edges {
		switch edge.EdgeType {
		case EdgeTypeFAMC, EdgeTypeCHIL, EdgeTypeParent, EdgeTypeChild, EdgeTypeSibling:
			hasBlood = true
		case EdgeTypeFAMS, EdgeTypeHUSB, EdgeTypeWIFE, EdgeTypeSpouse:
			hasMarital = true
		}
	}

	if hasBlood && hasMarital {
		return PathTypeMixed
	}
	if hasMarital {
		return PathTypeMarital
	}
	return PathTypeBlood
}
