package query

// Path represents a path between two nodes in the graph.
type Path struct {
	Nodes  []GraphNode
	Edges  []*Edge
	Length int
	Type   PathType
}

// PathType represents the type of path (blood, marital, mixed).
type PathType string

const (
	PathTypeBlood   PathType = "blood"   // Only blood relationships
	PathTypeMarital PathType = "marital" // Only marital relationships
	PathTypeMixed   PathType = "mixed"   // Mixed relationships
)
