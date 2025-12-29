package query

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
)

// BenchmarkHybridStorageComparison benchmarks SQLite vs PostgreSQL hybrid storage
// This test requires DATABASE_URL to be set for PostgreSQL tests
func BenchmarkHybridStorageComparison(b *testing.B) {
	// Skip if DATABASE_URL not set
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		b.Skip("Skipping comparison benchmark: DATABASE_URL not set")
	}

	// Use a small test file for quick benchmarks
	testFile := findTestDataFile("xavier.ged")
	if testFile == "" {
		b.Skip("Skipping: test file not found: xavier.ged")
	}

	// Parse the file once
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(testFile)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}

	b.Run("SQLite_BuildGraph", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tmpDir := b.TempDir()
			sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
			badgerPath := filepath.Join(tmpDir, "test_graph")

			start := time.Now()
			graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
			duration := time.Since(start)

			if err != nil {
				b.Fatalf("Failed to build SQLite graph: %v", err)
			}
			graph.Close()

			b.ReportMetric(float64(duration.Nanoseconds())/1e6, "ms/op")
		}
	})

	b.Run("PostgreSQL_BuildGraph", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tmpDir := b.TempDir()
			badgerPath := filepath.Join(tmpDir, "test_graph")
			fileID := "bench_file_" + time.Now().Format("20060102150405")

			start := time.Now()
			graph, err := BuildGraphHybridPostgres(tree, fileID, badgerPath, databaseURL, nil)
			duration := time.Since(start)

			if err != nil {
				b.Fatalf("Failed to build PostgreSQL graph: %v", err)
			}
			graph.Close()

			b.ReportMetric(float64(duration.Nanoseconds())/1e6, "ms/op")
		}
	})
}

// BenchmarkHybridQueryComparison benchmarks query performance
func BenchmarkHybridQueryComparison(b *testing.B) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		b.Skip("Skipping query comparison: DATABASE_URL not set")
	}

	testFile := findTestDataFile("xavier.ged")
	if testFile == "" {
		b.Skip("Skipping: test file not found: xavier.ged")
	}

	// Parse and build graphs once
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(testFile)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}

	// Build SQLite graph
	tmpDir := b.TempDir()
	sqlitePath := filepath.Join(tmpDir, "sqlite_indexes.db")
	badgerPathSQLite := filepath.Join(tmpDir, "sqlite_graph")
	sqliteGraph, err := BuildGraphHybrid(tree, sqlitePath, badgerPathSQLite, nil)
	if err != nil {
		b.Fatalf("Failed to build SQLite graph: %v", err)
	}
	defer sqliteGraph.Close()

	// Build PostgreSQL graph
	badgerPathPostgres := filepath.Join(tmpDir, "postgres_graph")
	fileID := "bench_query_file"
	postgresGraph, err := BuildGraphHybridPostgres(tree, fileID, badgerPathPostgres, databaseURL, nil)
	if err != nil {
		b.Fatalf("Failed to build PostgreSQL graph: %v", err)
	}
	defer postgresGraph.Close()

	// Get a test XREF for queries
	allIndividuals := sqliteGraph.GetAllIndividuals()
	var testXref string
	for xref := range allIndividuals {
		testXref = xref
		break
	}
	if testXref == "" {
		b.Skip("No individuals found in test file")
	}

	b.Run("SQLite_FindByXref", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := sqliteGraph.queryHelpers.FindByXref(testXref)
			if err != nil {
				b.Fatalf("FindByXref failed: %v", err)
			}
		}
	})

	b.Run("PostgreSQL_FindByXref", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := postgresGraph.queryHelpersPostgres.FindByXref(testXref)
			if err != nil {
				b.Fatalf("FindByXref failed: %v", err)
			}
		}
	})

	b.Run("SQLite_FilterQuery", func(b *testing.B) {
		fq := NewFilterQuery(sqliteGraph)
		for i := 0; i < b.N; i++ {
			_, err := fq.ByName("xavier").Execute()
			if err != nil {
				b.Fatalf("FilterQuery failed: %v", err)
			}
		}
	})

	b.Run("PostgreSQL_FilterQuery", func(b *testing.B) {
		fq := NewFilterQuery(postgresGraph)
		for i := 0; i < b.N; i++ {
			_, err := fq.ByName("xavier").Execute()
			if err != nil {
				b.Fatalf("FilterQuery failed: %v", err)
			}
		}
	})
}

// TestHybridStorageComparison runs a detailed comparison test
func TestHybridStorageComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comparison test in short mode")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("Skipping comparison test: DATABASE_URL not set")
	}

	testFile := findTestDataFile("xavier.ged")
	if testFile == "" {
		t.Skip("Skipping: test file not found: xavier.ged")
	}

	// Parse the file
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(testFile)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	t.Logf("Test file: %s", testFile)
	t.Logf("Individuals: %d", len(tree.GetAllIndividuals()))
	t.Logf("Families: %d", len(tree.GetAllFamilies()))

	// Test SQLite
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "sqlite_indexes.db")
	badgerPathSQLite := filepath.Join(tmpDir, "sqlite_graph")

	startSQLite := time.Now()
	sqliteGraph, err := BuildGraphHybrid(tree, sqlitePath, badgerPathSQLite, nil)
	sqliteBuildTime := time.Since(startSQLite)
	if err != nil {
		t.Fatalf("Failed to build SQLite graph: %v", err)
	}
	defer sqliteGraph.Close()

	// Test PostgreSQL
	badgerPathPostgres := filepath.Join(tmpDir, "postgres_graph")
	fileID := "comparison_test_file"

	startPostgres := time.Now()
	postgresGraph, err := BuildGraphHybridPostgres(tree, fileID, badgerPathPostgres, databaseURL, nil)
	postgresBuildTime := time.Since(startPostgres)
	if err != nil {
		t.Fatalf("Failed to build PostgreSQL graph: %v", err)
	}
	defer postgresGraph.Close()

	// Compare build times
	t.Logf("\n=== Build Time Comparison ===")
	t.Logf("SQLite:    %v", sqliteBuildTime)
	t.Logf("PostgreSQL: %v", postgresBuildTime)
	if postgresBuildTime > sqliteBuildTime {
		t.Logf("SQLite is %.2fx faster", float64(postgresBuildTime)/float64(sqliteBuildTime))
	} else {
		t.Logf("PostgreSQL is %.2fx faster", float64(sqliteBuildTime)/float64(postgresBuildTime))
	}

	// Compare query performance
	allIndividuals := sqliteGraph.GetAllIndividuals()
	var testXref string
	for xref := range allIndividuals {
		testXref = xref
		break
	}

	if testXref != "" {
		// FindByXref comparison
		startQuery := time.Now()
		for i := 0; i < 100; i++ {
			_, _ = sqliteGraph.queryHelpers.FindByXref(testXref)
		}
		sqliteQueryTime := time.Since(startQuery)

		startQuery = time.Now()
		for i := 0; i < 100; i++ {
			_, _ = postgresGraph.queryHelpersPostgres.FindByXref(testXref)
		}
		postgresQueryTime := time.Since(startQuery)

		t.Logf("\n=== Query Performance (100 iterations) ===")
		t.Logf("SQLite FindByXref:    %v (%.2f μs/op)", sqliteQueryTime, float64(sqliteQueryTime.Nanoseconds())/100/1000)
		t.Logf("PostgreSQL FindByXref: %v (%.2f μs/op)", postgresQueryTime, float64(postgresQueryTime.Nanoseconds())/100/1000)
	}

	// Compare FilterQuery performance
	fqSQLite := NewFilterQuery(sqliteGraph)
	startFilter := time.Now()
	for i := 0; i < 10; i++ {
		_, _ = fqSQLite.ByName("xavier").Execute()
	}
	sqliteFilterTime := time.Since(startFilter)

	fqPostgres := NewFilterQuery(postgresGraph)
	startFilter = time.Now()
	for i := 0; i < 10; i++ {
		_, _ = fqPostgres.ByName("xavier").Execute()
	}
	postgresFilterTime := time.Since(startFilter)

	t.Logf("\n=== FilterQuery Performance (10 iterations) ===")
	t.Logf("SQLite FilterQuery:    %v (%.2f ms/op)", sqliteFilterTime, float64(sqliteFilterTime.Nanoseconds())/10/1e6)
	t.Logf("PostgreSQL FilterQuery: %v (%.2f ms/op)", postgresFilterTime, float64(postgresFilterTime.Nanoseconds())/10/1e6)

	// Verify data consistency
	sqliteResults, _ := fqSQLite.ByName("xavier").Execute()
	postgresResults, _ := fqPostgres.ByName("xavier").Execute()

	if len(sqliteResults) != len(postgresResults) {
		t.Errorf("Result count mismatch: SQLite=%d, PostgreSQL=%d", len(sqliteResults), len(postgresResults))
	} else {
		t.Logf("\n=== Data Consistency ===")
		t.Logf("Both returned %d results for 'xavier' query", len(sqliteResults))
	}
}

