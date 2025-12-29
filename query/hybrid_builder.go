package query

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// BuildGraphHybrid builds a graph using hybrid storage (SQLite + BadgerDB)
// This function coordinates the building process by delegating to:
// - buildGraphInSQLite: Builds indexes in SQLite (see hybrid_sqlite_builder.go)
// - buildGraphInBadgerDB: Stores graph structure in BadgerDB (see hybrid_badger_builder.go)
// If config is nil, DefaultConfig() is used.
func BuildGraphHybrid(tree *types.GedcomTree, sqlitePath, badgerPath string, config *Config) (*Graph, error) {
	return BuildGraphHybridWithStorage(tree, sqlitePath, badgerPath, "", "", config, false)
}

// BuildGraphHybridPostgres builds a graph using hybrid storage (PostgreSQL + BadgerDB)
// This function coordinates the building process by delegating to:
// - buildGraphInPostgreSQL: Builds indexes in PostgreSQL (see hybrid_postgres_builder.go)
// - buildGraphInBadgerDB: Stores graph structure in BadgerDB (see hybrid_badger_builder.go)
// If config is nil, DefaultConfig() is used.
// fileID is required for PostgreSQL to identify which file the data belongs to.
// databaseURL can be empty, in which case it will use DATABASE_URL environment variable or config.
func BuildGraphHybridPostgres(tree *types.GedcomTree, fileID, badgerPath, databaseURL string, config *Config) (*Graph, error) {
	return BuildGraphHybridWithStorage(tree, "", badgerPath, fileID, databaseURL, config, true)
}

// BuildGraphHybridWithStorage is an internal function that supports both SQLite and PostgreSQL
func BuildGraphHybridWithStorage(tree *types.GedcomTree, sqlitePath, badgerPath, fileID, databaseURL string, config *Config, usePostgres bool) (*Graph, error) {
	// Use default config if none provided
	if config == nil {
		config = DefaultConfig()
	}

	// Create graph structure (will use hybrid storage)
	graph := NewGraphWithConfig(tree, config)
	graph.hybridMode = true

	var storage interface {
		Close() error
		BadgerDB() *badger.DB
	}

	var queryHelpers interface {
		Close() error
	}

	if usePostgres {
		// Initialize PostgreSQL hybrid storage
		if fileID == "" {
			return nil, fmt.Errorf("fileID is required for PostgreSQL storage")
		}
		postgresStorage, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, config)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize PostgreSQL hybrid storage: %w", err)
		}
		storage = postgresStorage
		graph.hybridStoragePostgres = postgresStorage

		// Initialize PostgreSQL query helpers with prepared statements
		postgresQueryHelpers, err := NewHybridQueryHelpersPostgres(postgresStorage.PostgreSQL(), fileID)
		if err != nil {
			postgresStorage.Close()
			return nil, fmt.Errorf("failed to create PostgreSQL query helpers: %w", err)
		}
		queryHelpers = postgresQueryHelpers
		graph.queryHelpersPostgres = postgresQueryHelpers
	} else {
		// Initialize SQLite hybrid storage
		sqliteStorage, err := NewHybridStorage(sqlitePath, badgerPath, config)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize SQLite hybrid storage: %w", err)
		}
		storage = sqliteStorage
		graph.hybridStorage = sqliteStorage

		// Initialize SQLite query helpers with prepared statements
		sqliteQueryHelpers, err := NewHybridQueryHelpers(sqliteStorage.SQLite())
		if err != nil {
			sqliteStorage.Close()
			return nil, fmt.Errorf("failed to create SQLite query helpers: %w", err)
		}
		queryHelpers = sqliteQueryHelpers
		graph.queryHelpers = sqliteQueryHelpers
	}

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
	if usePostgres {
		postgresStorage := graph.hybridStoragePostgres
		if err := buildGraphInPostgreSQL(postgresStorage, tree, graph); err != nil {
			storage.Close()
			return nil, fmt.Errorf("failed to build PostgreSQL indexes: %w", err)
		}
	} else {
		sqliteStorage := graph.hybridStorage
		if err := buildGraphInSQLite(sqliteStorage, tree, graph); err != nil {
			storage.Close()
			return nil, fmt.Errorf("failed to build SQLite indexes: %w", err)
		}
	}

	if err := buildGraphInBadgerDB(storage, tree, graph); err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to build BadgerDB graph: %w", err)
	}

	return graph, nil
}
