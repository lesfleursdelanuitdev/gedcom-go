package query

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test cache defaults
	if config.Cache.HybridNodeCacheSize != 50000 {
		t.Errorf("Expected HybridNodeCacheSize 50000, got %d", config.Cache.HybridNodeCacheSize)
	}
	if config.Cache.HybridXrefCacheSize != 25000 {
		t.Errorf("Expected HybridXrefCacheSize 25000, got %d", config.Cache.HybridXrefCacheSize)
	}
	if config.Cache.HybridQueryCacheSize != 5000 {
		t.Errorf("Expected HybridQueryCacheSize 5000, got %d", config.Cache.HybridQueryCacheSize)
	}
	if config.Cache.QueryCacheSize != 1000 {
		t.Errorf("Expected QueryCacheSize 1000, got %d", config.Cache.QueryCacheSize)
	}

	// Test timeout defaults
	if config.Timeout.SQLiteQueryTimeout != 30*time.Second {
		t.Errorf("Expected SQLiteQueryTimeout 30s, got %v", config.Timeout.SQLiteQueryTimeout)
	}
	if config.Timeout.BadgerDBTimeout != 10*time.Second {
		t.Errorf("Expected BadgerDBTimeout 10s, got %v", config.Timeout.BadgerDBTimeout)
	}
	if config.Timeout.BuildTimeout != 5*time.Minute {
		t.Errorf("Expected BuildTimeout 5m, got %v", config.Timeout.BuildTimeout)
	}
	if config.Timeout.QueryTimeout != 1*time.Minute {
		t.Errorf("Expected QueryTimeout 1m, got %v", config.Timeout.QueryTimeout)
	}

	// Test database defaults
	if config.Database.SQLiteMaxOpenConns != 10 {
		t.Errorf("Expected SQLiteMaxOpenConns 10, got %d", config.Database.SQLiteMaxOpenConns)
	}
	if config.Database.SQLiteMaxIdleConns != 5 {
		t.Errorf("Expected SQLiteMaxIdleConns 5, got %d", config.Database.SQLiteMaxIdleConns)
	}
	if config.Database.BadgerDBValueLogFileSize != 1<<30 {
		t.Errorf("Expected BadgerDBValueLogFileSize 1GB, got %d", config.Database.BadgerDBValueLogFileSize)
	}
}

func TestLoadConfig_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	// Create a test config file
	testConfig := &Config{
		Cache: CacheConfig{
			HybridNodeCacheSize:  10000,
			HybridXrefCacheSize:  5000,
			HybridQueryCacheSize: 1000,
			QueryCacheSize:       500,
		},
		Timeout: TimeoutConfig{
			SQLiteQueryTimeout: 60 * time.Second,
			BadgerDBTimeout:    20 * time.Second,
			BuildTimeout:       10 * time.Minute,
			QueryTimeout:       2 * time.Minute,
		},
		Database: DatabaseConfig{
			SQLiteMaxOpenConns:    20,
			SQLiteMaxIdleConns:    10,
			BadgerDBValueLogFileSize: 2 << 30, // 2GB
		},
	}

	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if loadedConfig.Cache.HybridNodeCacheSize != 10000 {
		t.Errorf("Expected HybridNodeCacheSize 10000, got %d", loadedConfig.Cache.HybridNodeCacheSize)
	}
	if loadedConfig.Cache.HybridXrefCacheSize != 5000 {
		t.Errorf("Expected HybridXrefCacheSize 5000, got %d", loadedConfig.Cache.HybridXrefCacheSize)
	}
	if loadedConfig.Timeout.SQLiteQueryTimeout != 60*time.Second {
		t.Errorf("Expected SQLiteQueryTimeout 60s, got %v", loadedConfig.Timeout.SQLiteQueryTimeout)
	}
	if loadedConfig.Database.SQLiteMaxOpenConns != 20 {
		t.Errorf("Expected SQLiteMaxOpenConns 20, got %d", loadedConfig.Database.SQLiteMaxOpenConns)
	}
}

func TestLoadConfig_FileNotExists(t *testing.T) {
	// Try to load from non-existent file
	config, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
	if config != nil {
		t.Error("Expected nil config when file doesn't exist")
	}
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	// Load with empty path should return default (no error)
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig with empty path should not error: %v", err)
	}
	if config == nil {
		t.Fatal("LoadConfig should return default config, not nil")
	}

	// Should have default values
	if config.Cache.HybridNodeCacheSize != 50000 {
		t.Errorf("Expected default HybridNodeCacheSize 50000, got %d", config.Cache.HybridNodeCacheSize)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved-config.json")

	config := DefaultConfig()
	config.Cache.HybridNodeCacheSize = 75000

	// Save config
	if err := SaveConfig(config, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load and verify
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Cache.HybridNodeCacheSize != 75000 {
		t.Errorf("Expected HybridNodeCacheSize 75000, got %d", loadedConfig.Cache.HybridNodeCacheSize)
	}
}

func TestValidateAndSetDefaults(t *testing.T) {
	config := &Config{
		Cache: CacheConfig{
			HybridNodeCacheSize:  0, // Should be set to default
			HybridXrefCacheSize:  1000, // Valid, should be kept
		},
		Timeout: TimeoutConfig{
			SQLiteQueryTimeout: 0, // Should be set to default
		},
	}

	config.validateAndSetDefaults()

	defaults := DefaultConfig()

	if config.Cache.HybridNodeCacheSize != defaults.Cache.HybridNodeCacheSize {
		t.Errorf("Expected HybridNodeCacheSize to be set to default %d, got %d",
			defaults.Cache.HybridNodeCacheSize, config.Cache.HybridNodeCacheSize)
	}

	if config.Cache.HybridXrefCacheSize != 1000 {
		t.Errorf("Expected HybridXrefCacheSize to remain 1000, got %d", config.Cache.HybridXrefCacheSize)
	}

	if config.Timeout.SQLiteQueryTimeout != defaults.Timeout.SQLiteQueryTimeout {
		t.Errorf("Expected SQLiteQueryTimeout to be set to default %v, got %v",
			defaults.Timeout.SQLiteQueryTimeout, config.Timeout.SQLiteQueryTimeout)
	}
}

func TestBuildGraphHybrid_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	// Create a custom config
	config := DefaultConfig()
	config.Cache.HybridNodeCacheSize = 1000
	config.Cache.HybridXrefCacheSize = 500
	config.Cache.HybridQueryCacheSize = 100

	// Create test data
	tree := gedcom.NewGedcomTree()
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Build graph with custom config
	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, config)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}
	defer graph.Close()

	// Verify cache was created with custom sizes
	if graph.hybridCache == nil {
		t.Fatal("Hybrid cache should be initialized")
	}

	stats := graph.hybridCache.Stats()
	// Cache should be empty initially, but capacity should match config
	// Note: LRU cache doesn't expose capacity directly, so we can't verify exact sizes
	// But we can verify the cache works
	if stats.NodeCacheSize < 0 {
		t.Error("Node cache size should be non-negative")
	}
}

func TestNewGraphWithConfig(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Test with custom config
	config := DefaultConfig()
	config.Cache.QueryCacheSize = 2000

	graph := NewGraphWithConfig(tree, config)
	if graph == nil {
		t.Fatal("Graph should not be nil")
	}

	// Graph should be created successfully
	// We can't directly verify cache size, but we can verify graph works
	if graph.tree != tree {
		t.Error("Graph tree should match input tree")
	}
}

func TestNewGraph_DefaultConfig(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Test with nil config (should use defaults)
	graph := NewGraphWithConfig(tree, nil)
	if graph == nil {
		t.Fatal("Graph should not be nil")
	}

	// Test with NewGraph (should also use defaults)
	graph2 := NewGraph(tree)
	if graph2 == nil {
		t.Fatal("Graph should not be nil")
	}
}

func TestConfig_JSONSerialization(t *testing.T) {
	config := DefaultConfig()
	config.Cache.HybridNodeCacheSize = 75000
	config.Timeout.SQLiteQueryTimeout = 45 * time.Second

	// Marshal to JSON
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal from JSON
	var loadedConfig Config
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify values
	if loadedConfig.Cache.HybridNodeCacheSize != 75000 {
		t.Errorf("Expected HybridNodeCacheSize 75000, got %d", loadedConfig.Cache.HybridNodeCacheSize)
	}

	// Duration should be preserved
	if loadedConfig.Timeout.SQLiteQueryTimeout != 45*time.Second {
		t.Errorf("Expected SQLiteQueryTimeout 45s, got %v", loadedConfig.Timeout.SQLiteQueryTimeout)
	}
}

func TestConfig_DurationStringParsing(t *testing.T) {
	// Test JSON with duration strings
	jsonData := `{
		"cache": {
			"hybrid_node_cache_size": 10000,
			"hybrid_xref_cache_size": 5000,
			"hybrid_query_cache_size": 1000,
			"query_cache_size": 500
		},
		"timeout": {
			"sqlite_query_timeout": "60s",
			"badgerdb_timeout": "20s",
			"build_timeout": "10m",
			"query_timeout": "2m"
		},
		"database": {
			"sqlite_max_open_conns": 20,
			"sqlite_max_idle_conns": 10,
			"badgerdb_value_log_file_size": 2147483648
		}
	}`

	var config Config
	if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
		t.Fatalf("Failed to unmarshal config with duration strings: %v", err)
	}

	// Verify durations were parsed correctly
	if config.Timeout.SQLiteQueryTimeout != 60*time.Second {
		t.Errorf("Expected SQLiteQueryTimeout 60s, got %v", config.Timeout.SQLiteQueryTimeout)
	}
	if config.Timeout.BadgerDBTimeout != 20*time.Second {
		t.Errorf("Expected BadgerDBTimeout 20s, got %v", config.Timeout.BadgerDBTimeout)
	}
	if config.Timeout.BuildTimeout != 10*time.Minute {
		t.Errorf("Expected BuildTimeout 10m, got %v", config.Timeout.BuildTimeout)
	}
	if config.Timeout.QueryTimeout != 2*time.Minute {
		t.Errorf("Expected QueryTimeout 2m, got %v", config.Timeout.QueryTimeout)
	}
}

func TestConfig_DurationNumberParsing(t *testing.T) {
	// Test JSON with duration as nanoseconds (number)
	jsonData := `{
		"timeout": {
			"sqlite_query_timeout": 30000000000,
			"badgerdb_timeout": 10000000000
		}
	}`

	var config Config
	if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
		t.Fatalf("Failed to unmarshal config with duration numbers: %v", err)
	}

	// Verify durations were parsed correctly (30000000000 ns = 30s)
	if config.Timeout.SQLiteQueryTimeout != 30*time.Second {
		t.Errorf("Expected SQLiteQueryTimeout 30s, got %v", config.Timeout.SQLiteQueryTimeout)
	}
	if config.Timeout.BadgerDBTimeout != 10*time.Second {
		t.Errorf("Expected BadgerDBTimeout 10s, got %v", config.Timeout.BadgerDBTimeout)
	}
}

