package query

import (
	"fmt"
)

// BFS performs breadth-first search starting from a node.
// visitor function is called for each visited node. Return false to stop traversal.
func (g *Graph) BFS(startID string, visitor func(GraphNode) bool) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	startNode := g.nodes[startID]
	if startNode == nil {
		return fmt.Errorf("node %s not found", startID)
	}

	visited := make(map[string]bool)
	queue := []GraphNode{startNode}
	visited[startID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if !visitor(current) {
			return nil // Stop traversal
		}

		// Add neighbors to queue
		for _, edge := range current.OutEdges() {
			neighbor := edge.To
			if neighbor != nil && !visited[neighbor.ID()] {
				visited[neighbor.ID()] = true
				queue = append(queue, neighbor)
			}
		}

		// Also check bidirectional edges
		for _, edge := range current.InEdges() {
			if edge.IsBidirectional() {
				neighbor := edge.From
				if neighbor != nil && !visited[neighbor.ID()] {
					visited[neighbor.ID()] = true
					queue = append(queue, neighbor)
				}
			}
		}
	}

	return nil
}

// DFS performs depth-first search starting from a node.
// visitor function is called for each visited node. Return false to stop traversal.
func (g *Graph) DFS(startID string, visitor func(GraphNode) bool) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	startNode := g.nodes[startID]
	if startNode == nil {
		return fmt.Errorf("node %s not found", startID)
	}

	visited := make(map[string]bool)
	return g.dfsRecursive(startNode, visited, visitor)
}

// dfsRecursive is the recursive helper for DFS.
func (g *Graph) dfsRecursive(node GraphNode, visited map[string]bool, visitor func(GraphNode) bool) error {
	if visited[node.ID()] {
		return nil
	}

	visited[node.ID()] = true

	if !visitor(node) {
		return nil // Stop traversal
	}

	// Visit neighbors
	for _, edge := range node.OutEdges() {
		neighbor := edge.To
		if neighbor != nil && !visited[neighbor.ID()] {
			if err := g.dfsRecursive(neighbor, visited, visitor); err != nil {
				return err
			}
		}
	}

	// Also check bidirectional edges
	for _, edge := range node.InEdges() {
		if edge.IsBidirectional() {
			neighbor := edge.From
			if neighbor != nil && !visited[neighbor.ID()] {
				if err := g.dfsRecursive(neighbor, visited, visitor); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
