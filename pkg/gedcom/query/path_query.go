package query

// PathOptions holds configuration for path queries.
type PathOptions struct {
	MaxLength      int  // Maximum path length (0 = unlimited, but defaults to 10)
	IncludeMarital bool // Include marital relationships
	IncludeBlood   bool // Include blood relationships
	ShortestOnly   bool // Return only shortest path
}

// NewPathOptions creates new PathOptions with defaults.
func NewPathOptions() *PathOptions {
	return &PathOptions{
		MaxLength:      10, // Default maximum
		IncludeMarital: true,
		IncludeBlood:   true,
		ShortestOnly:   false,
	}
}

// PathQuery represents a query for paths between two individuals.
type PathQuery struct {
	fromXrefID string
	toXrefID   string
	graph      *Graph
	options    *PathOptions
}

// MaxLength sets the maximum path length.
func (pq *PathQuery) MaxLength(n int) *PathQuery {
	pq.options.MaxLength = n
	return pq
}

// IncludeMarital sets whether to include marital relationships.
func (pq *PathQuery) IncludeMarital(include bool) *PathQuery {
	pq.options.IncludeMarital = include
	return pq
}

// IncludeBlood sets whether to include blood relationships.
func (pq *PathQuery) IncludeBlood(include bool) *PathQuery {
	pq.options.IncludeBlood = include
	return pq
}

// ShortestOnly sets whether to return only the shortest path.
func (pq *PathQuery) ShortestOnly(shortest bool) *PathQuery {
	pq.options.ShortestOnly = shortest
	return pq
}

// Shortest returns the shortest path between the two individuals.
func (pq *PathQuery) Shortest() (*Path, error) {
	return pq.graph.ShortestPath(pq.fromXrefID, pq.toXrefID)
}

// All returns all paths between the two individuals.
func (pq *PathQuery) All() ([]*Path, error) {
	maxLength := pq.options.MaxLength
	if maxLength <= 0 {
		maxLength = 10
	}

	paths, err := pq.graph.AllPaths(pq.fromXrefID, pq.toXrefID, maxLength)
	if err != nil {
		return nil, err
	}

	// Filter by path type if needed
	if !pq.options.IncludeMarital || !pq.options.IncludeBlood {
		filtered := make([]*Path, 0)
		for _, path := range paths {
			if pq.shouldIncludePath(path) {
				filtered = append(filtered, path)
			}
		}
		return filtered, nil
	}

	return paths, nil
}

// shouldIncludePath checks if a path should be included based on options.
func (pq *PathQuery) shouldIncludePath(path *Path) bool {
	if path.Type == PathTypeBlood && !pq.options.IncludeBlood {
		return false
	}
	if path.Type == PathTypeMarital && !pq.options.IncludeMarital {
		return false
	}
	if path.Type == PathTypeMixed {
		// Mixed paths are included if either type is allowed
		return pq.options.IncludeBlood || pq.options.IncludeMarital
	}
	return true
}

// Count returns the number of paths.
func (pq *PathQuery) Count() (int, error) {
	if pq.options.ShortestOnly {
		path, err := pq.Shortest()
		if err != nil {
			return 0, nil // No path found
		}
		if path != nil {
			return 1, nil
		}
		return 0, nil
	}

	paths, err := pq.All()
	if err != nil {
		return 0, err
	}
	return len(paths), nil
}
