package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// BuildGraphHybrid builds a graph using hybrid storage (SQLite + BadgerDB)
// This function coordinates the building process by delegating to:
// - buildGraphInSQLite: Builds indexes in SQLite (see hybrid_sqlite_builder.go)
// - buildGraphInBadgerDB: Stores graph structure in BadgerDB (see hybrid_badger_builder.go)
// If config is nil, DefaultConfig() is used.
func BuildGraphHybrid(tree *gedcom.GedcomTree, sqlitePath, badgerPath string, config *Config) (*Graph, error) {
	// Use default config if none provided
	if config == nil {
		config = DefaultConfig()
	}

	// Initialize hybrid storage
	storage, err := NewHybridStorage(sqlitePath, badgerPath, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hybrid storage: %w", err)
	}

	// Create graph structure (will use hybrid storage)
	graph := NewGraphWithConfig(tree, config)
	graph.hybridStorage = storage
	graph.hybridMode = true

	// Initialize query helpers with prepared statements
	queryHelpers, err := NewHybridQueryHelpers(storage.SQLite())
	if err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to create query helpers: %w", err)
	}
	graph.queryHelpers = queryHelpers

	// Initialize hybrid cache with configured sizes
	hybridCache, err := NewHybridCache(
		config.Cache.HybridNodeCacheSize,
		config.Cache.HybridXrefCacheSize,
		config.Cache.HybridQueryCacheSize,
	)
	if err != nil {
		queryHelpers.Close()
		storage.Close()
		return nil, fmt.Errorf("failed to create hybrid cache: %w", err)
	}
	graph.hybridCache = hybridCache

	// Build graph in both databases
	if err := buildGraphInSQLite(storage, tree, graph); err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to build SQLite indexes: %w", err)
	}

	if err := buildGraphInBadgerDB(storage, tree, graph); err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to build BadgerDB graph: %w", err)
	}

	return graph, nil
}
