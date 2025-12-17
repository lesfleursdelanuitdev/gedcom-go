package query

import (
	"sync"
)

// Memory pools for temporary data structures to reduce allocations.

var (
	// Pool for BFS/DFS queues
	queuePool = sync.Pool{
		New: func() interface{} {
			return make([]GraphNode, 0, 64)
		},
	}

	// Pool for visited maps
	visitedMapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]bool, 64)
		},
	}

	// Pool for parent maps (for path reconstruction)
	parentMapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]struct {
				node GraphNode
				edge *Edge
			}, 64)
		},
	}

	// Pool for path slices
	pathSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]GraphNode, 0, 32)
		},
	}

	// Pool for edge slices
	edgeSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]*Edge, 0, 32)
		},
	}
)

// getQueue gets a queue from the pool.
func getQueue() []GraphNode {
	return queuePool.Get().([]GraphNode)
}

// putQueue returns a queue to the pool.
func putQueue(q []GraphNode) {
	if q == nil {
		return
	}
	// Clear the slice but keep capacity
	q = q[:0]
	queuePool.Put(q)
}

// getVisitedMap gets a visited map from the pool.
func getVisitedMap() map[string]bool {
	m := visitedMapPool.Get().(map[string]bool)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// putVisitedMap returns a visited map to the pool.
func putVisitedMap(m map[string]bool) {
	if m == nil {
		return
	}
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	visitedMapPool.Put(m)
}

// getParentMap gets a parent map from the pool.
func getParentMap() map[string]struct {
	node GraphNode
	edge *Edge
} {
	m := parentMapPool.Get().(map[string]struct {
		node GraphNode
		edge *Edge
	})
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// putParentMap returns a parent map to the pool.
func putParentMap(m map[string]struct {
	node GraphNode
	edge *Edge
}) {
	if m == nil {
		return
	}
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	parentMapPool.Put(m)
}

// getPathSlice gets a path slice from the pool.
func getPathSlice() []GraphNode {
	return pathSlicePool.Get().([]GraphNode)
}

// putPathSlice returns a path slice to the pool.
func putPathSlice(s []GraphNode) {
	if s == nil {
		return
	}
	s = s[:0]
	pathSlicePool.Put(s)
}

// getEdgeSlice gets an edge slice from the pool.
func getEdgeSlice() []*Edge {
	return edgeSlicePool.Get().([]*Edge)
}

// putEdgeSlice returns an edge slice to the pool.
func putEdgeSlice(s []*Edge) {
	if s == nil {
		return
	}
	s = s[:0]
	edgeSlicePool.Put(s)
}
