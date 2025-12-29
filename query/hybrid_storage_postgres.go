package query

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/dgraph-io/badger/v4"
)

// HybridStoragePostgres manages both PostgreSQL and BadgerDB databases
// This is similar to HybridStorage but uses PostgreSQL instead of SQLite
type HybridStoragePostgres struct {
	postgresDB *sql.DB
	badgerDB   *badger.DB
	badgerPath string
	fileID     string // File ID for this graph (used in PostgreSQL queries)
}

// NewHybridStoragePostgres creates a new PostgreSQL-based hybrid storage instance
// If config is nil, DefaultConfig() is used.
// databaseURL can be empty, in which case it will use DATABASE_URL environment variable
func NewHybridStoragePostgres(fileID, badgerPath, databaseURL string, config *Config) (*HybridStoragePostgres, error) {
	// Use default config if none provided
	if config == nil {
		config = DefaultConfig()
	}

	// Get database URL from config or environment
	if databaseURL == "" {
		databaseURL = config.Database.PostgreSQLDatabaseURL
		if databaseURL == "" {
			databaseURL = os.Getenv("DATABASE_URL")
		}
		if databaseURL == "" {
			return nil, fmt.Errorf("DATABASE_URL environment variable not set and no database URL provided")
		}
	}

	hs := &HybridStoragePostgres{
		fileID:     fileID,
		badgerPath: badgerPath,
	}

	// Initialize PostgreSQL
	if err := hs.initPostgreSQL(config, databaseURL); err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Initialize BadgerDB
	if err := hs.initBadgerDB(config); err != nil {
		hs.Close() // Clean up PostgreSQL if BadgerDB fails
		return nil, fmt.Errorf("failed to initialize BadgerDB: %w", err)
	}

	return hs, nil
}

// initPostgreSQL initializes the PostgreSQL database and creates schema
func (hs *HybridStoragePostgres) initPostgreSQL(config *Config, databaseURL string) error {
	// Open PostgreSQL database
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Set connection pool settings from config
	db.SetMaxOpenConns(config.Database.PostgreSQLMaxOpenConns)
	db.SetMaxIdleConns(config.Database.PostgreSQLMaxIdleConns)
	db.SetConnMaxLifetime(config.Database.PostgreSQLConnMaxLifetime)
	db.SetConnMaxIdleTime(config.Database.PostgreSQLConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	hs.postgresDB = db

	// Create schema
	if err := hs.createPostgreSQLSchema(); err != nil {
		db.Close()
		return fmt.Errorf("failed to create PostgreSQL schema: %w", err)
	}

	return nil
}

// createPostgreSQLSchema creates all tables and indexes for PostgreSQL
// Note: Uses file_id column for multi-file support (shared database)
func (hs *HybridStoragePostgres) createPostgreSQLSchema() error {
	schema := `
	-- Nodes table (with file_id for multi-file support)
	CREATE TABLE IF NOT EXISTS nodes (
		file_id TEXT NOT NULL,
		id INTEGER NOT NULL,
		xref TEXT NOT NULL,
		type TEXT NOT NULL,
		name TEXT,
		name_lower TEXT,
		birth_date BIGINT,
		birth_place TEXT,
		sex TEXT,
		has_children INTEGER DEFAULT 0,
		has_spouse INTEGER DEFAULT 0,
		living INTEGER DEFAULT 0,
		created_at BIGINT,
		updated_at BIGINT,
		PRIMARY KEY (file_id, id),
		UNIQUE (file_id, xref)
	);

	-- Indexes for fast lookups (all include file_id for efficient filtering)
	CREATE INDEX IF NOT EXISTS idx_nodes_file_id ON nodes(file_id);
	CREATE INDEX IF NOT EXISTS idx_nodes_xref ON nodes(file_id, xref);
	CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(file_id, type);
	CREATE INDEX IF NOT EXISTS idx_nodes_name_lower ON nodes(file_id, name_lower);
	CREATE INDEX IF NOT EXISTS idx_nodes_birth_date ON nodes(file_id, birth_date);
	CREATE INDEX IF NOT EXISTS idx_nodes_birth_place ON nodes(file_id, birth_place);
	CREATE INDEX IF NOT EXISTS idx_nodes_sex ON nodes(file_id, sex);
	CREATE INDEX IF NOT EXISTS idx_nodes_has_children ON nodes(file_id, has_children);
	CREATE INDEX IF NOT EXISTS idx_nodes_has_spouse ON nodes(file_id, has_spouse);
	CREATE INDEX IF NOT EXISTS idx_nodes_living ON nodes(file_id, living);

	-- Composite indexes for common queries
	CREATE INDEX IF NOT EXISTS idx_nodes_name_date ON nodes(file_id, name_lower, birth_date);
	CREATE INDEX IF NOT EXISTS idx_nodes_place_date ON nodes(file_id, birth_place, birth_date);

	-- XREF mapping table
	CREATE TABLE IF NOT EXISTS xref_mapping (
		file_id TEXT NOT NULL,
		xref TEXT NOT NULL,
		node_id INTEGER NOT NULL,
		PRIMARY KEY (file_id, xref),
		FOREIGN KEY (file_id, node_id) REFERENCES nodes(file_id, id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_xref_mapping_file_id ON xref_mapping(file_id);
	CREATE INDEX IF NOT EXISTS idx_xref_mapping_node_id ON xref_mapping(file_id, node_id);

	-- Components table
	CREATE TABLE IF NOT EXISTS components (
		file_id TEXT NOT NULL,
		component_id INTEGER NOT NULL,
		node_id INTEGER NOT NULL,
		PRIMARY KEY (file_id, component_id, node_id),
		FOREIGN KEY (file_id, node_id) REFERENCES nodes(file_id, id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_components_file_id ON components(file_id);
	CREATE INDEX IF NOT EXISTS idx_components_node_id ON components(file_id, node_id);
	CREATE INDEX IF NOT EXISTS idx_components_component_id ON components(file_id, component_id);

	-- Full-text search using PostgreSQL's built-in full-text search
	-- Create GIN index for full-text search on name and birth_place
	CREATE INDEX IF NOT EXISTS idx_nodes_name_fts ON nodes USING gin(to_tsvector('english', COALESCE(name, '')));
	CREATE INDEX IF NOT EXISTS idx_nodes_place_fts ON nodes USING gin(to_tsvector('english', COALESCE(birth_place, '')));
	`

	// Execute schema
	_, err := hs.postgresDB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute PostgreSQL schema: %w", err)
	}

	return nil
}

// initBadgerDB initializes the BadgerDB database (same as SQLite version)
func (hs *HybridStoragePostgres) initBadgerDB(config *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(hs.badgerPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create BadgerDB directory: %w", err)
	}

	// Configure BadgerDB options
	opts := badger.DefaultOptions(hs.badgerPath)
	opts.Logger = nil // Disable logging for now
	opts.ValueLogFileSize = config.Database.BadgerDBValueLogFileSize

	// Open BadgerDB
	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	hs.badgerDB = db
	return nil
}

// Close closes both databases
func (hs *HybridStoragePostgres) Close() error {
	var errs []error

	if hs.postgresDB != nil {
		if err := hs.postgresDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close PostgreSQL: %w", err))
		}
	}

	if hs.badgerDB != nil {
		if err := hs.badgerDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close BadgerDB: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}

	return nil
}

// PostgreSQL returns the PostgreSQL database connection
func (hs *HybridStoragePostgres) PostgreSQL() *sql.DB {
	return hs.postgresDB
}

// BadgerDB returns the BadgerDB database instance
func (hs *HybridStoragePostgres) BadgerDB() *badger.DB {
	return hs.badgerDB
}

// FileID returns the file ID for this storage instance
func (hs *HybridStoragePostgres) FileID() string {
	return hs.fileID
}

