package query

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/dgraph-io/badger/v4"
)

// HybridStorage manages both SQLite and BadgerDB databases
type HybridStorage struct {
	sqliteDB *sql.DB
	badgerDB *badger.DB
	sqlitePath string
	badgerPath string
}

// NewHybridStorage creates a new hybrid storage instance
// If config is nil, DefaultConfig() is used.
func NewHybridStorage(sqlitePath, badgerPath string, config *Config) (*HybridStorage, error) {
	// Use default config if none provided
	if config == nil {
		config = DefaultConfig()
	}

	hs := &HybridStorage{
		sqlitePath: sqlitePath,
		badgerPath: badgerPath,
	}

	// Initialize SQLite
	if err := hs.initSQLite(config); err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite: %w", err)
	}

	// Initialize BadgerDB
	if err := hs.initBadgerDB(config); err != nil {
		hs.Close() // Clean up SQLite if BadgerDB fails
		return nil, fmt.Errorf("failed to initialize BadgerDB: %w", err)
	}

	return hs, nil
}

// initSQLite initializes the SQLite database and creates schema
func (hs *HybridStorage) initSQLite(config *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(hs.sqlitePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create SQLite directory: %w", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", hs.sqlitePath+"?_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Set connection pool settings from config
	db.SetMaxOpenConns(config.Database.SQLiteMaxOpenConns)
	db.SetMaxIdleConns(config.Database.SQLiteMaxIdleConns)

	hs.sqliteDB = db

	// Create schema
	if err := hs.createSQLiteSchema(); err != nil {
		db.Close()
		return fmt.Errorf("failed to create SQLite schema: %w", err)
	}

	return nil
}

// createSQLiteSchema creates all tables, indexes, and FTS5 virtual table
func (hs *HybridStorage) createSQLiteSchema() error {
	schema := `
	-- Nodes table
	CREATE TABLE IF NOT EXISTS nodes (
		id INTEGER PRIMARY KEY,
		xref TEXT UNIQUE NOT NULL,
		type TEXT NOT NULL,
		name TEXT,
		name_lower TEXT,
		birth_date INTEGER,
		birth_place TEXT,
		sex TEXT,
		has_children INTEGER DEFAULT 0,
		has_spouse INTEGER DEFAULT 0,
		living INTEGER DEFAULT 0,
		created_at INTEGER,
		updated_at INTEGER
	);

	-- Indexes for fast lookups
	CREATE INDEX IF NOT EXISTS idx_nodes_xref ON nodes(xref);
	CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);
	CREATE INDEX IF NOT EXISTS idx_nodes_name_lower ON nodes(name_lower);
	CREATE INDEX IF NOT EXISTS idx_nodes_birth_date ON nodes(birth_date);
	CREATE INDEX IF NOT EXISTS idx_nodes_birth_place ON nodes(birth_place);
	CREATE INDEX IF NOT EXISTS idx_nodes_sex ON nodes(sex);
	CREATE INDEX IF NOT EXISTS idx_nodes_has_children ON nodes(has_children);
	CREATE INDEX IF NOT EXISTS idx_nodes_has_spouse ON nodes(has_spouse);
	CREATE INDEX IF NOT EXISTS idx_nodes_living ON nodes(living);

	-- Composite indexes for common queries
	CREATE INDEX IF NOT EXISTS idx_nodes_name_date ON nodes(name_lower, birth_date);
	CREATE INDEX IF NOT EXISTS idx_nodes_place_date ON nodes(birth_place, birth_date);

	-- XREF mapping table
	CREATE TABLE IF NOT EXISTS xref_mapping (
		xref TEXT PRIMARY KEY,
		node_id INTEGER NOT NULL,
		FOREIGN KEY (node_id) REFERENCES nodes(id)
	);

	CREATE INDEX IF NOT EXISTS idx_xref_mapping_node_id ON xref_mapping(node_id);

	-- Components table
	CREATE TABLE IF NOT EXISTS components (
		component_id INTEGER NOT NULL,
		node_id INTEGER NOT NULL,
		FOREIGN KEY (node_id) REFERENCES nodes(id),
		PRIMARY KEY (component_id, node_id)
	);

	CREATE INDEX IF NOT EXISTS idx_components_node_id ON components(node_id);
	CREATE INDEX IF NOT EXISTS idx_components_component_id ON components(component_id);

	-- Full-text search virtual table (FTS5 - optional, may not be available in all SQLite builds)
	-- Note: This will fail silently if FTS5 is not available
	-- We'll handle this gracefully in the application
	CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
		name,
		birth_place,
		content='nodes',
		content_rowid='id'
	);

	-- Triggers to keep FTS in sync (only created if FTS5 table exists)
	CREATE TRIGGER IF NOT EXISTS nodes_fts_insert AFTER INSERT ON nodes BEGIN
		INSERT INTO nodes_fts(rowid, name, birth_place)
		VALUES (new.id, new.name, new.birth_place);
	END;

	CREATE TRIGGER IF NOT EXISTS nodes_fts_update AFTER UPDATE ON nodes BEGIN
		UPDATE nodes_fts SET name = new.name, birth_place = new.birth_place
		WHERE rowid = new.id;
	END;

	CREATE TRIGGER IF NOT EXISTS nodes_fts_delete AFTER DELETE ON nodes BEGIN
		DELETE FROM nodes_fts WHERE rowid = old.id;
	END;

	-- Enable memory-mapped I/O
	PRAGMA mmap_size = 268435456;  -- 256 MB
	`

	// Execute schema - FTS5 may not be available, so we'll try to create it
	// and continue if it fails
	_, err := hs.sqliteDB.Exec(schema)
	if err != nil {
		// If FTS5 is not available, try creating schema without FTS5
		// This is a fallback for systems without FTS5 support
		schemaWithoutFTS5 := `
		-- Nodes table
		CREATE TABLE IF NOT EXISTS nodes (
			id INTEGER PRIMARY KEY,
			xref TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL,
			name TEXT,
			name_lower TEXT,
			birth_date INTEGER,
			birth_place TEXT,
			sex TEXT,
			has_children INTEGER DEFAULT 0,
			has_spouse INTEGER DEFAULT 0,
			living INTEGER DEFAULT 0,
			created_at INTEGER,
			updated_at INTEGER
		);

		-- Indexes for fast lookups
		CREATE INDEX IF NOT EXISTS idx_nodes_xref ON nodes(xref);
		CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);
		CREATE INDEX IF NOT EXISTS idx_nodes_name_lower ON nodes(name_lower);
		CREATE INDEX IF NOT EXISTS idx_nodes_birth_date ON nodes(birth_date);
		CREATE INDEX IF NOT EXISTS idx_nodes_birth_place ON nodes(birth_place);
		CREATE INDEX IF NOT EXISTS idx_nodes_sex ON nodes(sex);
		CREATE INDEX IF NOT EXISTS idx_nodes_has_children ON nodes(has_children);
		CREATE INDEX IF NOT EXISTS idx_nodes_has_spouse ON nodes(has_spouse);
		CREATE INDEX IF NOT EXISTS idx_nodes_living ON nodes(living);

		-- Composite indexes for common queries
		CREATE INDEX IF NOT EXISTS idx_nodes_name_date ON nodes(name_lower, birth_date);
		CREATE INDEX IF NOT EXISTS idx_nodes_place_date ON nodes(birth_place, birth_date);

		-- XREF mapping table
		CREATE TABLE IF NOT EXISTS xref_mapping (
			xref TEXT PRIMARY KEY,
			node_id INTEGER NOT NULL,
			FOREIGN KEY (node_id) REFERENCES nodes(id)
		);

		CREATE INDEX IF NOT EXISTS idx_xref_mapping_node_id ON xref_mapping(node_id);

		-- Components table
		CREATE TABLE IF NOT EXISTS components (
			component_id INTEGER NOT NULL,
			node_id INTEGER NOT NULL,
			FOREIGN KEY (node_id) REFERENCES nodes(id),
			PRIMARY KEY (component_id, node_id)
		);

		CREATE INDEX IF NOT EXISTS idx_components_node_id ON components(node_id);
		CREATE INDEX IF NOT EXISTS idx_components_component_id ON components(component_id);

		-- Enable memory-mapped I/O
		PRAGMA mmap_size = 268435456;  -- 256 MB
		`
		
		if _, err2 := hs.sqliteDB.Exec(schemaWithoutFTS5); err2 != nil {
			return fmt.Errorf("failed to execute schema (with and without FTS5): %w (original: %v)", err2, err)
		}
		// FTS5 not available, but core schema created successfully
		// This is acceptable - we can still use regular indexes
	}

	return nil
}

// initBadgerDB initializes the BadgerDB database
func (hs *HybridStorage) initBadgerDB(config *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(hs.badgerPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create BadgerDB directory: %w", err)
	}

	// Configure BadgerDB options
	opts := badger.DefaultOptions(hs.badgerPath)
	// Note: BadgerDB v4 uses different options - memory mapping is handled automatically
	// We can configure other options like compression, etc. here if needed
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
func (hs *HybridStorage) Close() error {
	var errs []error

	if hs.sqliteDB != nil {
		if err := hs.sqliteDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close SQLite: %w", err))
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

// SQLite returns the SQLite database connection
func (hs *HybridStorage) SQLite() *sql.DB {
	return hs.sqliteDB
}

// BadgerDB returns the BadgerDB database instance
func (hs *HybridStorage) BadgerDB() *badger.DB {
	return hs.badgerDB
}

