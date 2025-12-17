package query

import (
	"fmt"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

// CentralityMetric represents the type of centrality measure.
type CentralityMetric string

const (
	CentralityDegree      CentralityMetric = "degree"      // Number of connections
	CentralityBetweenness CentralityMetric = "betweenness" // Betweenness centrality
	CentralityCloseness   CentralityMetric = "closeness"   // Closeness centrality
)

// GraphMetricsQuery provides graph metrics and analytics.
type GraphMetricsQuery struct {
	graph *Graph
}

// NewGraphMetricsQuery creates a new GraphMetricsQuery.
func NewGraphMetricsQuery(graph *Graph) *GraphMetricsQuery {
	return &GraphMetricsQuery{
		graph: graph,
	}
}

// Metrics returns a GraphMetricsQuery for the graph.
func (g *Graph) Metrics() *GraphMetricsQuery {
	return NewGraphMetricsQuery(g)
}

// Degree returns the degree (number of connections) of an individual.
func (gmq *GraphMetricsQuery) Degree(xrefID string) (int, error) {
	node := gmq.graph.GetIndividual(xrefID)
	if node == nil {
		return 0, fmt.Errorf("individual %s not found", xrefID)
	}

	return node.Degree(), nil
}

// InDegree returns the number of incoming connections (parents, spouses).
func (gmq *GraphMetricsQuery) InDegree(xrefID string) (int, error) {
	node := gmq.graph.GetIndividual(xrefID)
	if node == nil {
		return 0, fmt.Errorf("individual %s not found", xrefID)
	}

	return node.InDegree(), nil
}

// OutDegree returns the number of outgoing connections (children, spouses).
func (gmq *GraphMetricsQuery) OutDegree(xrefID string) (int, error) {
	node := gmq.graph.GetIndividual(xrefID)
	if node == nil {
		return 0, fmt.Errorf("individual %s not found", xrefID)
	}

	return node.OutDegree(), nil
}

// Centrality calculates various centrality measures for all individuals.
func (gmq *GraphMetricsQuery) Centrality(metric CentralityMetric) (map[string]float64, error) {
	individuals := gmq.graph.GetAllIndividuals()
	centrality := make(map[string]float64)

	switch metric {
	case CentralityDegree:
		for id, node := range individuals {
			centrality[id] = float64(node.Degree())
		}

	case CentralityBetweenness:
		return gmq.calculateBetweennessCentrality()

	case CentralityCloseness:
		return gmq.calculateClosenessCentrality()

	default:
		return nil, fmt.Errorf("unknown centrality metric: %s", metric)
	}

	return centrality, nil
}

// calculateBetweennessCentrality calculates betweenness centrality.
// Betweenness centrality measures how often a node appears on shortest paths.
func (gmq *GraphMetricsQuery) calculateBetweennessCentrality() (map[string]float64, error) {
	individuals := gmq.graph.GetAllIndividuals()
	betweenness := make(map[string]float64)

	// Initialize
	for id := range individuals {
		betweenness[id] = 0.0
	}

	// For each pair of nodes, find shortest path and count node appearances
	allIDs := make([]string, 0, len(individuals))
	for id := range individuals {
		allIDs = append(allIDs, id)
	}

	// Calculate for all pairs
	for i, fromID := range allIDs {
		for j := i + 1; j < len(allIDs); j++ {
			toID := allIDs[j]

			path, err := gmq.graph.ShortestPath(fromID, toID)
			if err != nil {
				continue // No path
			}

			// Count appearances of each node in the path (excluding endpoints)
			for k := 1; k < len(path.Nodes)-1; k++ {
				nodeID := path.Nodes[k].ID()
				betweenness[nodeID]++
			}
		}
	}

	// Normalize (optional - divide by number of pairs)
	totalPairs := float64(len(allIDs) * (len(allIDs) - 1) / 2)
	if totalPairs > 0 {
		for id := range betweenness {
			betweenness[id] /= totalPairs
		}
	}

	return betweenness, nil
}

// calculateClosenessCentrality calculates closeness centrality.
// Closeness centrality measures average distance to all other nodes.
func (gmq *GraphMetricsQuery) calculateClosenessCentrality() (map[string]float64, error) {
	individuals := gmq.graph.GetAllIndividuals()
	closeness := make(map[string]float64)

	allIDs := make([]string, 0, len(individuals))
	for id := range individuals {
		allIDs = append(allIDs, id)
		closeness[id] = 0.0
	}

	// For each node, calculate average distance to all others
	for _, fromID := range allIDs {
		totalDistance := 0.0
		reachableCount := 0

		for _, toID := range allIDs {
			if fromID == toID {
				continue
			}

			path, err := gmq.graph.ShortestPath(fromID, toID)
			if err != nil {
				continue // Not reachable
			}

			totalDistance += float64(path.Length)
			reachableCount++
		}

		if reachableCount > 0 {
			// Closeness is inverse of average distance
			avgDistance := totalDistance / float64(reachableCount)
			if avgDistance > 0 {
				closeness[fromID] = 1.0 / avgDistance
			}
		}
	}

	return closeness, nil
}

// Diameter returns the diameter of the family tree (longest shortest path).
func (gmq *GraphMetricsQuery) Diameter() (int, error) {
	individuals := gmq.graph.GetAllIndividuals()
	if len(individuals) < 2 {
		return 0, nil
	}

	maxPathLength := 0
	allIDs := make([]string, 0, len(individuals))
	for id := range individuals {
		allIDs = append(allIDs, id)
	}

	// Find longest shortest path between any two nodes
	for i, fromID := range allIDs {
		for j := i + 1; j < len(allIDs); j++ {
			toID := allIDs[j]

			path, err := gmq.graph.ShortestPath(fromID, toID)
			if err != nil {
				continue // No path
			}

			if path.Length > maxPathLength {
				maxPathLength = path.Length
			}
		}
	}

	return maxPathLength, nil
}

// ConnectedComponents returns all connected components in the graph.
func (gmq *GraphMetricsQuery) ConnectedComponents() ([][]*gedcom.IndividualRecord, error) {
	individuals := gmq.graph.GetAllIndividuals()
	visited := make(map[string]bool)
	components := make([][]*gedcom.IndividualRecord, 0)

	for id, node := range individuals {
		if visited[id] {
			continue
		}

		// BFS to find all nodes in this component
		component := make([]*gedcom.IndividualRecord, 0)
		queue := []*IndividualNode{node}
		visited[id] = true

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if current.Individual != nil {
				component = append(component, current.Individual)
			}

			// Find connected individuals through all edges
			connectedIndividuals := make(map[string]*IndividualNode)

			// Through FAMC edges (parents)
			for _, edge := range current.OutEdges() {
				if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
					famNode := edge.Family
					if famNode.Husband != nil {
						connectedIndividuals[famNode.Husband.ID()] = famNode.Husband
					}
					if famNode.Wife != nil {
						connectedIndividuals[famNode.Wife.ID()] = famNode.Wife
					}
					// Also add siblings (other children of same family)
					for _, child := range famNode.Children {
						if child.ID() != current.ID() {
							connectedIndividuals[child.ID()] = child
						}
					}
				}
			}

			// Through FAMS edges (spouses and children)
			for _, edge := range current.OutEdges() {
				if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
					famNode := edge.Family
					if famNode.Husband != nil && famNode.Husband.ID() != current.ID() {
						connectedIndividuals[famNode.Husband.ID()] = famNode.Husband
					}
					if famNode.Wife != nil && famNode.Wife.ID() != current.ID() {
						connectedIndividuals[famNode.Wife.ID()] = famNode.Wife
					}
					for _, child := range famNode.Children {
						connectedIndividuals[child.ID()] = child
					}
				}
			}

			// Add connected individuals to queue
			for neighborID, neighborNode := range connectedIndividuals {
				if !visited[neighborID] {
					visited[neighborID] = true
					queue = append(queue, neighborNode)
				}
			}
		}

		if len(component) > 0 {
			components = append(components, component)
		}
	}

	return components, nil
}

// IsConnected checks if two individuals are connected (path exists).
func (gmq *GraphMetricsQuery) IsConnected(xrefID1, xrefID2 string) (bool, error) {
	_, err := gmq.graph.ShortestPath(xrefID1, xrefID2)
	if err != nil {
		return false, nil // Not connected
	}
	return true, nil
}

// LongestPath finds the longest path in the tree.
func (gmq *GraphMetricsQuery) LongestPath() (*Path, error) {
	individuals := gmq.graph.GetAllIndividuals()
	if len(individuals) < 2 {
		return nil, fmt.Errorf("need at least 2 individuals for longest path")
	}

	var longestPath *Path
	maxLength := -1

	allIDs := make([]string, 0, len(individuals))
	for id := range individuals {
		allIDs = append(allIDs, id)
	}

	// Find longest path between any two nodes
	for i, fromID := range allIDs {
		for j := i + 1; j < len(allIDs); j++ {
			toID := allIDs[j]

			// Try to find longest path (use AllPaths with high limit)
			paths, err := gmq.graph.AllPaths(fromID, toID, 20)
			if err != nil {
				continue
			}

			for _, path := range paths {
				if path.Length > maxLength {
					maxLength = path.Length
					longestPath = path
				}
			}
		}
	}

	if longestPath == nil {
		return nil, fmt.Errorf("no path found")
	}

	return longestPath, nil
}

// AveragePathLength calculates the average shortest path length between all pairs.
func (gmq *GraphMetricsQuery) AveragePathLength() (float64, error) {
	individuals := gmq.graph.GetAllIndividuals()
	if len(individuals) < 2 {
		return 0.0, nil
	}

	totalLength := 0.0
	pairCount := 0
	allIDs := make([]string, 0, len(individuals))
	for id := range individuals {
		allIDs = append(allIDs, id)
	}

	for i, fromID := range allIDs {
		for j := i + 1; j < len(allIDs); j++ {
			toID := allIDs[j]

			path, err := gmq.graph.ShortestPath(fromID, toID)
			if err != nil {
				continue // No path
			}

			totalLength += float64(path.Length)
			pairCount++
		}
	}

	if pairCount == 0 {
		return 0.0, nil
	}

	return totalLength / float64(pairCount), nil
}

// NodeCount returns the total number of nodes in the graph.
func (gmq *GraphMetricsQuery) NodeCount() int {
	return gmq.graph.NodeCount()
}

// EdgeCount returns the total number of edges in the graph.
func (gmq *GraphMetricsQuery) EdgeCount() int {
	return gmq.graph.EdgeCount()
}

// AverageDegree calculates the average degree of all nodes.
func (gmq *GraphMetricsQuery) AverageDegree() (float64, error) {
	individuals := gmq.graph.GetAllIndividuals()
	if len(individuals) == 0 {
		return 0.0, nil
	}

	totalDegree := 0
	for _, node := range individuals {
		totalDegree += node.Degree()
	}

	return float64(totalDegree) / float64(len(individuals)), nil
}

// Density calculates the graph density (actual edges / possible edges).
func (gmq *GraphMetricsQuery) Density() (float64, error) {
	nodeCount := gmq.graph.NodeCount()
	if nodeCount < 2 {
		return 0.0, nil
	}

	edgeCount := gmq.graph.EdgeCount()
	maxPossibleEdges := float64(nodeCount * (nodeCount - 1))

	if maxPossibleEdges == 0 {
		return 0.0, nil
	}

	return float64(edgeCount) / maxPossibleEdges, nil
}
