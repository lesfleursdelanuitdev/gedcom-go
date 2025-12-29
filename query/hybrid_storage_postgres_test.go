package query

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// getPostgreSQLTestURL returns a PostgreSQL connection URL for testing
// Returns empty string if DATABASE_URL is not set (tests will be skipped)
func getPostgreSQLTestURL(t *testing.T) string {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("Skipping PostgreSQL test: DATABASE_URL environment variable not set")
	}
	return databaseURL
}

// testPostgreSQLConnection tests if we can connect to PostgreSQL
func testPostgreSQLConnection(t *testing.T, databaseURL string) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to open connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to ping database: %v", err)
	}
}

func TestHybridStoragePostgres_Initialization(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	// Create temporary directory for BadgerDB
	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_001"

	// Create PostgreSQL hybrid storage
	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	// Verify PostgreSQL is initialized
	if hs.PostgreSQL() == nil {
		t.Error("PostgreSQL database is nil")
	}

	// Verify BadgerDB is initialized
	if hs.BadgerDB() == nil {
		t.Error("BadgerDB database is nil")
	}

	// Verify fileID
	if hs.FileID() != fileID {
		t.Errorf("Expected fileID '%s', got '%s'", fileID, hs.FileID())
	}

	// Test PostgreSQL schema - try a simple query
	var count int
	err = hs.PostgreSQL().QueryRow("SELECT COUNT(*) FROM nodes WHERE file_id = $1", fileID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query PostgreSQL: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 nodes, got %d", count)
	}

	// Test BadgerDB - try a simple read
	err = hs.BadgerDB().View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("test_key"))
		if err != badger.ErrKeyNotFound {
			// Key not found is expected for empty database
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to access BadgerDB: %v", err)
	}
}

func TestHybridStoragePostgres_SchemaCreation(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_002"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	db := hs.PostgreSQL()

	// Test that all tables exist
	tables := []string{"nodes", "xref_mapping", "components"}
	for _, table := range tables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`
		err := db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check if table %s exists: %v", table, err)
		}
		if !exists {
			t.Errorf("Table %s does not exist", table)
		}
	}

	// Test that indexes exist (check a few key ones)
	indexes := []string{"idx_nodes_file_id", "idx_nodes_xref", "idx_xref_mapping_file_id"}
	for _, index := range indexes {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM pg_indexes 
				WHERE indexname = $1
			)
		`
		err := db.QueryRow(query, index).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check if index %s exists: %v", index, err)
		}
		if !exists {
			t.Errorf("Index %s does not exist", index)
		}
	}
}

func TestHybridStoragePostgres_Cleanup(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_003"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}

	// Close should not error
	if err := hs.Close(); err != nil {
		t.Errorf("Failed to close PostgreSQL hybrid storage: %v", err)
	}

	// Verify BadgerDB directory exists (it should, even after close)
	if _, err := os.Stat(badgerPath); err != nil {
		t.Errorf("BadgerDB directory should exist: %v", err)
	}
}

func TestHybridStoragePostgres_FileIDIsolation(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath1 := filepath.Join(tmpDir, "test_graph1")
	badgerPath2 := filepath.Join(tmpDir, "test_graph2")
	fileID1 := "test_file_004"
	fileID2 := "test_file_005"

	// Create two storage instances with different file IDs
	hs1, err := NewHybridStoragePostgres(fileID1, badgerPath1, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create first PostgreSQL hybrid storage: %v", err)
	}
	defer hs1.Close()

	hs2, err := NewHybridStoragePostgres(fileID2, badgerPath2, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create second PostgreSQL hybrid storage: %v", err)
	}
	defer hs2.Close()

	db := hs1.PostgreSQL() // Both use same database

	// Insert a node in file1
	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, fileID1, 1, "@I1@", "individual", "John Doe", "john doe", 1000, 1000)
	if err != nil {
		t.Fatalf("Failed to insert node in file1: %v", err)
	}

	// Insert a node in file2
	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, fileID2, 1, "@I1@", "individual", "Jane Smith", "jane smith", 1000, 1000)
	if err != nil {
		t.Fatalf("Failed to insert node in file2: %v", err)
	}

	// Verify file_id isolation - file1 should only see its own node
	var count1 int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes WHERE file_id = $1", fileID1).Scan(&count1)
	if err != nil {
		t.Fatalf("Failed to query file1 nodes: %v", err)
	}
	if count1 != 1 {
		t.Errorf("File1 should have 1 node, got %d", count1)
	}

	// Verify file2 can only see its own node
	var count2 int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes WHERE file_id = $1", fileID2).Scan(&count2)
	if err != nil {
		t.Fatalf("Failed to query file2 nodes: %v", err)
	}
	if count2 != 1 {
		t.Errorf("File2 should have 1 node, got %d", count2)
	}

	// Verify the nodes are different
	var name1, name2 string
	err = db.QueryRow("SELECT name FROM nodes WHERE file_id = $1 AND id = $2", fileID1, 1).Scan(&name1)
	if err != nil {
		t.Fatalf("Failed to get name from file1: %v", err)
	}
	err = db.QueryRow("SELECT name FROM nodes WHERE file_id = $1 AND id = $2", fileID2, 1).Scan(&name2)
	if err != nil {
		t.Fatalf("Failed to get name from file2: %v", err)
	}

	if name1 == name2 {
		t.Error("Nodes from different files should be different")
	}
	if name1 != "John Doe" {
		t.Errorf("Expected 'John Doe' in file1, got '%s'", name1)
	}
	if name2 != "Jane Smith" {
		t.Errorf("Expected 'Jane Smith' in file2, got '%s'", name2)
	}

	// Clean up test data
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID1)
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID2)
}
