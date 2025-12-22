package query

import (
	"sync"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// Graph represents the graph structure of a GEDCOM tree.
type Graph struct {
	// Core data
	tree *gedcom.GedcomTree

	// ID mapping: XREF string <-> uint32 ID (for memory efficiency)
	xrefToID map[string]uint32 // XREF string -> uint32 ID
	idToXref map[uint32]string // uint32 ID -> XREF string
	nextID   uint32            // Next available ID

	// Node storage (using uint32 IDs internally for memory efficiency)
	nodes        map[uint32]GraphNode // All nodes by ID
	individuals  map[uint32]*IndividualNode
	families     map[uint32]*FamilyNode
	notes        map[uint32]*NoteNode
	sources      map[uint32]*SourceNode
	repositories map[uint32]*RepositoryNode
	events       map[uint32]*EventNode

	// Edge storage
	edges     map[string]*Edge   // All edges by ID (edge IDs remain strings)
	edgeIndex map[uint32][]*Edge // Edges by node ID (using uint32 for consistency)

	// Lazy loading support
	nodeMetadata map[uint32]*NodeMetadata // Node metadata (skeleton) - always loaded
	edgesLoaded  map[uint32]bool          // Track which nodes have edges loaded
	edgesLoading map[uint32]bool          // Track which nodes are currently loading edges (prevents recursion)
	lazyMode     bool                     // If true, use lazy loading; if false, load everything upfront

	// Graph partitioning support
	components     map[uint32][]uint32 // ComponentID -> []nodeIDs
	componentCount uint32              // Number of components found

	// Thread safety
	mu sync.RWMutex

	// Metadata
	properties map[string]interface{}

	// Performance optimizations
	cache   *queryCache
	indexes *FilterIndexes

	// Hybrid storage support
	hybridStorage *HybridStorage
	hybridMode    bool
	queryHelpers  *HybridQueryHelpers // SQLite query helpers
	hybridCache   *HybridCache        // LRU cache for hybrid storage
}

// NewGraph creates a new empty graph with default configuration.
func NewGraph(tree *gedcom.GedcomTree) *Graph {
	return NewGraphWithConfig(tree, DefaultConfig())
}

// NewGraphWithConfig creates a new empty graph with the provided configuration.
func NewGraphWithConfig(tree *gedcom.GedcomTree, config *Config) *Graph {
	// Use default config if none provided
	if config == nil {
		config = DefaultConfig()
	}

	return &Graph{
		tree:           tree,
		xrefToID:       make(map[string]uint32),
		idToXref:       make(map[uint32]string),
		nextID:         1, // Start IDs at 1 (0 reserved for invalid)
		nodes:          make(map[uint32]GraphNode),
		individuals:    make(map[uint32]*IndividualNode),
		families:       make(map[uint32]*FamilyNode),
		notes:          make(map[uint32]*NoteNode),
		sources:        make(map[uint32]*SourceNode),
		repositories:   make(map[uint32]*RepositoryNode),
		events:         make(map[uint32]*EventNode),
		edges:          make(map[string]*Edge),
		edgeIndex:      make(map[uint32][]*Edge),
		nodeMetadata:   make(map[uint32]*NodeMetadata),
		edgesLoaded:    make(map[uint32]bool),
		edgesLoading:   make(map[uint32]bool),
		lazyMode:       false, // Default: eager loading (backward compatible)
		components:     make(map[uint32][]uint32),
		componentCount: 0,
		properties:     make(map[string]interface{}),
		cache:          newQueryCache(config.Cache.QueryCacheSize),
		indexes:        newFilterIndexes(),
	}
}

// NodeCount returns the total number of nodes.
func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the total number of edges.
func (g *Graph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.edges)
}

// Tree returns the underlying GEDCOM tree.
func (g *Graph) Tree() *gedcom.GedcomTree {
	return g.tree
}
