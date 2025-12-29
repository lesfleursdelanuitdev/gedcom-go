package query

import (
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestHybridQueryHelpersPostgres_FindByXref(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_query_001"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	// Insert test data
	db := hs.PostgreSQL()
	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, fileID, 1, "@I1@", "individual", "John Doe", "john doe", time.Now().Unix(), time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert test node: %v", err)
	}

	// Create query helpers
	helpers, err := NewHybridQueryHelpersPostgres(db, fileID)
	if err != nil {
		t.Fatalf("Failed to create query helpers: %v", err)
	}
	defer helpers.Close()

	// Test FindByXref
	nodeID, err := helpers.FindByXref("@I1@")
	if err != nil {
		t.Fatalf("FindByXref failed: %v", err)
	}
	if nodeID != 1 {
		t.Errorf("Expected nodeID 1, got %d", nodeID)
	}

	// Test non-existent XREF
	nodeID, err = helpers.FindByXref("@I999@")
	if err != nil {
		t.Fatalf("FindByXref should not error for non-existent XREF: %v", err)
	}
	if nodeID != 0 {
		t.Errorf("Expected nodeID 0 for non-existent XREF, got %d", nodeID)
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
}

func TestHybridQueryHelpersPostgres_FindXrefByID(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_query_002"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	db := hs.PostgreSQL()
	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, fileID, 42, "@I42@", "individual", "Test Person", "test person", time.Now().Unix(), time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert test node: %v", err)
	}

	helpers, err := NewHybridQueryHelpersPostgres(db, fileID)
	if err != nil {
		t.Fatalf("Failed to create query helpers: %v", err)
	}
	defer helpers.Close()

	// Test FindXrefByID
	xref, err := helpers.FindXrefByID(42)
	if err != nil {
		t.Fatalf("FindXrefByID failed: %v", err)
	}
	if xref != "@I42@" {
		t.Errorf("Expected '@I42@', got '%s'", xref)
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
}

func TestHybridQueryHelpersPostgres_FindByName(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_query_003"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	db := hs.PostgreSQL()
	now := time.Now().Unix()

	// Insert multiple test nodes
	testNodes := []struct {
		id   int
		xref string
		name string
	}{
		{1, "@I1@", "John Doe"},
		{2, "@I2@", "Jane Smith"},
		{3, "@I3@", "Johnny Appleseed"},
	}

	for _, node := range testNodes {
		_, err = db.Exec(`
			INSERT INTO nodes (file_id, id, xref, type, name, name_lower, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, fileID, node.id, node.xref, "individual", node.name, toLower(node.name), now, now)
		if err != nil {
			t.Fatalf("Failed to insert test node %s: %v", node.xref, err)
		}
	}

	helpers, err := NewHybridQueryHelpersPostgres(db, fileID)
	if err != nil {
		t.Fatalf("Failed to create query helpers: %v", err)
	}
	defer helpers.Close()

	// Test FindByName (substring match)
	nodeIDs, err := helpers.FindByName("john")
	if err != nil {
		t.Fatalf("FindByName failed: %v", err)
	}
	if len(nodeIDs) != 2 {
		t.Errorf("Expected 2 results for 'john', got %d", len(nodeIDs))
	}

	// Test FindByNameExact
	nodeIDs, err = helpers.FindByNameExact("john doe")
	if err != nil {
		t.Fatalf("FindByNameExact failed: %v", err)
	}
	if len(nodeIDs) != 1 {
		t.Errorf("Expected 1 result for exact match 'john doe', got %d", len(nodeIDs))
	}
	if len(nodeIDs) > 0 && nodeIDs[0] != 1 {
		t.Errorf("Expected nodeID 1, got %d", nodeIDs[0])
	}

	// Test FindByNameStarts
	nodeIDs, err = helpers.FindByNameStarts("jane")
	if err != nil {
		t.Fatalf("FindByNameStarts failed: %v", err)
	}
	if len(nodeIDs) != 1 {
		t.Errorf("Expected 1 result for prefix 'jane', got %d", len(nodeIDs))
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
}

func TestHybridQueryHelpersPostgres_FindByBirthDate(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_query_004"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	db := hs.PostgreSQL()
	now := time.Now().Unix()

	// Insert nodes with birth dates
	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, birth_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, fileID, 1, "@I1@", "individual", "Person 1", "person 1", time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Unix(), now, now)
	if err != nil {
		t.Fatalf("Failed to insert test node: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, birth_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, fileID, 2, "@I2@", "individual", "Person 2", "person 2", time.Date(1950, 6, 15, 0, 0, 0, 0, time.UTC).Unix(), now, now)
	if err != nil {
		t.Fatalf("Failed to insert test node: %v", err)
	}

	helpers, err := NewHybridQueryHelpersPostgres(db, fileID)
	if err != nil {
		t.Fatalf("Failed to create query helpers: %v", err)
	}
	defer helpers.Close()

	// Test FindByBirthDate
	start := time.Date(1940, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(1960, 12, 31, 23, 59, 59, 0, time.UTC)
	nodeIDs, err := helpers.FindByBirthDate(start, end)
	if err != nil {
		t.Fatalf("FindByBirthDate failed: %v", err)
	}
	if len(nodeIDs) != 1 {
		t.Errorf("Expected 1 result for date range, got %d", len(nodeIDs))
	}
	if len(nodeIDs) > 0 && nodeIDs[0] != 2 {
		t.Errorf("Expected nodeID 2, got %d", nodeIDs[0])
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
}

func TestHybridQueryHelpersPostgres_FindBySex(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_query_005"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	db := hs.PostgreSQL()
	now := time.Now().Unix()

	// Insert nodes with different sexes
	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, sex, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, fileID, 1, "@I1@", "individual", "Male Person", "male person", "M", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test node: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, sex, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, fileID, 2, "@I2@", "individual", "Female Person", "female person", "F", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test node: %v", err)
	}

	helpers, err := NewHybridQueryHelpersPostgres(db, fileID)
	if err != nil {
		t.Fatalf("Failed to create query helpers: %v", err)
	}
	defer helpers.Close()

	// Test FindBySex - Male
	nodeIDs, err := helpers.FindBySex("M")
	if err != nil {
		t.Fatalf("FindBySex failed: %v", err)
	}
	if len(nodeIDs) != 1 {
		t.Errorf("Expected 1 male, got %d", len(nodeIDs))
	}

	// Test FindBySex - Female
	nodeIDs, err = helpers.FindBySex("F")
	if err != nil {
		t.Fatalf("FindBySex failed: %v", err)
	}
	if len(nodeIDs) != 1 {
		t.Errorf("Expected 1 female, got %d", len(nodeIDs))
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
}

func TestHybridQueryHelpersPostgres_BooleanFlags(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_query_006"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	db := hs.PostgreSQL()
	now := time.Now().Unix()

	// Insert node with flags set
	_, err = db.Exec(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, has_children, has_spouse, living, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, fileID, 1, "@I1@", "individual", "Test Person", "test person", 1, 1, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test node: %v", err)
	}

	helpers, err := NewHybridQueryHelpersPostgres(db, fileID)
	if err != nil {
		t.Fatalf("Failed to create query helpers: %v", err)
	}
	defer helpers.Close()

	// Test HasChildren
	hasChildren, err := helpers.HasChildren(1)
	if err != nil {
		t.Fatalf("HasChildren failed: %v", err)
	}
	if !hasChildren {
		t.Error("Expected hasChildren to be true")
	}

	// Test HasSpouse
	hasSpouse, err := helpers.HasSpouse(1)
	if err != nil {
		t.Fatalf("HasSpouse failed: %v", err)
	}
	if !hasSpouse {
		t.Error("Expected hasSpouse to be true")
	}

	// Test IsLiving
	isLiving, err := helpers.IsLiving(1)
	if err != nil {
		t.Fatalf("IsLiving failed: %v", err)
	}
	if isLiving {
		t.Error("Expected isLiving to be false")
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
}

func TestHybridQueryHelpersPostgres_GetAllIndividualIDs(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_file_query_007"

	hs, err := NewHybridStoragePostgres(fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL hybrid storage: %v", err)
	}
	defer hs.Close()

	db := hs.PostgreSQL()
	now := time.Now().Unix()

	// Insert multiple individuals
	for i := 1; i <= 5; i++ {
		_, err = db.Exec(`
			INSERT INTO nodes (file_id, id, xref, type, name, name_lower, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, fileID, i, "@I"+string(rune('0'+i))+"@", "individual", "Person "+string(rune('0'+i)), "person "+string(rune('0'+i)), now, now)
		if err != nil {
			t.Fatalf("Failed to insert test node %d: %v", i, err)
		}
	}

	helpers, err := NewHybridQueryHelpersPostgres(db, fileID)
	if err != nil {
		t.Fatalf("Failed to create query helpers: %v", err)
	}
	defer helpers.Close()

	// Test GetAllIndividualIDs
	nodeIDs, err := helpers.GetAllIndividualIDs()
	if err != nil {
		t.Fatalf("GetAllIndividualIDs failed: %v", err)
	}
	if len(nodeIDs) != 5 {
		t.Errorf("Expected 5 individuals, got %d", len(nodeIDs))
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
}
