package query

import (
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestBuildGraphHybridPostgres_Basic(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_build_001"

	// Create test data
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	sex1Line := types.NewGedcomLine(1, "SEX", "M", "")
	indi1Line.AddChild(name1Line)
	indi1Line.AddChild(sex1Line)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Build PostgreSQL hybrid graph
	graph, err := BuildGraphHybridPostgres(tree, fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to build PostgreSQL hybrid graph: %v", err)
	}
	defer graph.Close()

	// Verify graph was created
	if graph == nil {
		t.Fatal("Graph is nil")
	}

	// Verify PostgreSQL storage is set
	if graph.hybridStoragePostgres == nil {
		t.Error("PostgreSQL storage should be initialized")
	}

	// Verify query helpers are set
	if graph.queryHelpersPostgres == nil {
		t.Error("PostgreSQL query helpers should be initialized")
	}

	// Verify fileID matches
	if graph.hybridStoragePostgres.FileID() != fileID {
		t.Errorf("Expected fileID '%s', got '%s'", fileID, graph.hybridStoragePostgres.FileID())
	}

	// Test that we can query the data
	helpers := graph.queryHelpersPostgres
	nodeID, err := helpers.FindByXref("@I1@")
	if err != nil {
		t.Fatalf("Failed to find node by XREF: %v", err)
	}
	if nodeID == 0 {
		t.Error("Expected non-zero node ID")
	}

	// Clean up
	db := graph.hybridStoragePostgres.PostgreSQL()
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
	_, _ = db.Exec("DELETE FROM xref_mapping WHERE file_id = $1", fileID)
}

func TestBuildGraphHybridPostgres_WithFamily(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_build_002"

	// Create test data with family
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := types.NewGedcomLine(1, "WIFE", "@I2@", "")
	fam1Line.AddChild(husbLine)
	fam1Line.AddChild(wifeLine)
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Build graph
	graph, err := BuildGraphHybridPostgres(tree, fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}
	defer graph.Close()

	// Verify individuals are in database
	helpers := graph.queryHelpersPostgres
	nodeID1, err := helpers.FindByXref("@I1@")
	if err != nil {
		t.Fatalf("Failed to find @I1@: %v", err)
	}
	if nodeID1 == 0 {
		t.Error("Expected non-zero node ID for @I1@")
	}

	nodeID2, err := helpers.FindByXref("@I2@")
	if err != nil {
		t.Fatalf("Failed to find @I2@: %v", err)
	}
	if nodeID2 == 0 {
		t.Error("Expected non-zero node ID for @I2@")
	}

	// Verify family is in database
	famNodeID, err := helpers.FindByXref("@F1@")
	if err != nil {
		t.Fatalf("Failed to find @F1@: %v", err)
	}
	if famNodeID == 0 {
		t.Error("Expected non-zero node ID for @F1@")
	}

	// Clean up
	db := graph.hybridStoragePostgres.PostgreSQL()
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
	_, _ = db.Exec("DELETE FROM xref_mapping WHERE file_id = $1", fileID)
}

func TestBuildGraphHybridPostgres_FileIDIsolation(t *testing.T) {
	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath1 := filepath.Join(tmpDir, "test_graph1")
	badgerPath2 := filepath.Join(tmpDir, "test_graph2")
	fileID1 := "test_build_003"
	fileID2 := "test_build_004"

	// Create two trees with same XREFs but different data
	tree1 := types.NewGedcomTree()
	indi1Line1 := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line1 := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line1.AddChild(name1Line1)
	indi1 := types.NewIndividualRecord(indi1Line1)
	tree1.AddRecord(indi1)

	tree2 := types.NewGedcomTree()
	indi1Line2 := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line2 := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi1Line2.AddChild(name1Line2)
	indi2 := types.NewIndividualRecord(indi1Line2)
	tree2.AddRecord(indi2)

	// Build both graphs
	graph1, err := BuildGraphHybridPostgres(tree1, fileID1, badgerPath1, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to build graph1: %v", err)
	}
	defer graph1.Close()

	graph2, err := BuildGraphHybridPostgres(tree2, fileID2, badgerPath2, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to build graph2: %v", err)
	}
	defer graph2.Close()

	// Verify file isolation - same XREF but different data
	helpers1 := graph1.queryHelpersPostgres
	helpers2 := graph2.queryHelpersPostgres

	nodeID1, err := helpers1.FindByXref("@I1@")
	if err != nil {
		t.Fatalf("Failed to find @I1@ in graph1: %v", err)
	}

	nodeID2, err := helpers2.FindByXref("@I1@")
	if err != nil {
		t.Fatalf("Failed to find @I1@ in graph2: %v", err)
	}

	// Both should have nodeID 1 (first node in each file)
	if nodeID1 != 1 {
		t.Errorf("Expected nodeID 1 in graph1, got %d", nodeID1)
	}
	if nodeID2 != 1 {
		t.Errorf("Expected nodeID 1 in graph2, got %d", nodeID2)
	}

	// Verify names are different (check in database)
	db := graph1.hybridStoragePostgres.PostgreSQL()
	var name1, name2 string
	err = db.QueryRow("SELECT name FROM nodes WHERE file_id = $1 AND id = $2", fileID1, 1).Scan(&name1)
	if err != nil {
		t.Fatalf("Failed to get name from graph1: %v", err)
	}
	err = db.QueryRow("SELECT name FROM nodes WHERE file_id = $1 AND id = $2", fileID2, 1).Scan(&name2)
	if err != nil {
		t.Fatalf("Failed to get name from graph2: %v", err)
	}

	if name1 == name2 {
		t.Error("Names should be different between files")
	}
	if name1 != "John /Doe/" {
		t.Errorf("Expected 'John /Doe/' in graph1, got '%s'", name1)
	}
	if name2 != "Jane /Smith/" {
		t.Errorf("Expected 'Jane /Smith/' in graph2, got '%s'", name2)
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID1)
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID2)
	_, _ = db.Exec("DELETE FROM xref_mapping WHERE file_id = $1", fileID1)
	_, _ = db.Exec("DELETE FROM xref_mapping WHERE file_id = $1", fileID2)
}

func TestBuildGraphHybridPostgres_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseURL := getPostgreSQLTestURL(t)
	testPostgreSQLConnection(t, databaseURL)

	tmpDir := t.TempDir()
	badgerPath := filepath.Join(tmpDir, "test_graph")
	fileID := "test_build_integration"

	// Create test data similar to hybrid_integration_test.go
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	sex1Line := types.NewGedcomLine(1, "SEX", "M", "")
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	date1Line := types.NewGedcomLine(2, "DATE", "15 JAN 1800", "")
	birt1Line.AddChild(date1Line)
	indi1Line.AddChild(name1Line)
	indi1Line.AddChild(sex1Line)
	indi1Line.AddChild(birt1Line)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	sex2Line := types.NewGedcomLine(1, "SEX", "F", "")
	indi2Line.AddChild(name2Line)
	indi2Line.AddChild(sex2Line)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := types.NewGedcomLine(1, "WIFE", "@I2@", "")
	fam1Line.AddChild(husbLine)
	fam1Line.AddChild(wifeLine)
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Build graph
	graph, err := BuildGraphHybridPostgres(tree, fileID, badgerPath, databaseURL, nil)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}
	defer graph.Close()

	// Test 1: Query by name using FilterQuery
	fq := NewFilterQuery(graph)
	results, err := fq.ByName("John").Execute()
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test 2: Get individual
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Error("Expected to find individual @I1@")
	}

	// Test 3: Query helpers work
	helpers := graph.queryHelpersPostgres
	allIDs, err := helpers.GetAllIndividualIDs()
	if err != nil {
		t.Fatalf("GetAllIndividualIDs failed: %v", err)
	}
	if len(allIDs) != 2 {
		t.Errorf("Expected 2 individuals, got %d", len(allIDs))
	}

	// Clean up
	db := graph.hybridStoragePostgres.PostgreSQL()
	_, _ = db.Exec("DELETE FROM nodes WHERE file_id = $1", fileID)
	_, _ = db.Exec("DELETE FROM xref_mapping WHERE file_id = $1", fileID)
}

