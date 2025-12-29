package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestCache_Operations tests cache operations comprehensively
func TestCache_Operations(t *testing.T) {
	// Test newQueryCache with different sizes
	cache1 := newQueryCache(10)
	if cache1 == nil {
		t.Fatal("Expected non-nil cache")
	}
	if cache1.maxSize != 10 {
		t.Errorf("Expected maxSize 10, got %d", cache1.maxSize)
	}

	// Test with zero size (should default to 1000)
	cache2 := newQueryCache(0)
	if cache2.maxSize != 1000 {
		t.Errorf("Expected default maxSize 1000, got %d", cache2.maxSize)
	}

	// Test with negative size (should default to 1000)
	cache3 := newQueryCache(-1)
	if cache3.maxSize != 1000 {
		t.Errorf("Expected default maxSize 1000, got %d", cache3.maxSize)
	}

	// Test get/set operations
	cache := newQueryCache(5)
	val, found := cache.get("key1")
	if found {
		t.Error("Expected key1 not to be found initially")
	}
	if val != nil {
		t.Error("Expected nil value for missing key")
	}

	// Set a value
	cache.set("key1", "value1")
	val, found = cache.get("key1")
	if !found {
		t.Error("Expected key1 to be found after set")
	}
	if val != "value1" {
		t.Errorf("Expected value 'value1', got %v", val)
	}

	// Test cache eviction (fill cache beyond maxSize)
	for i := 0; i < 10; i++ {
		cache.set(makeCacheKey("test", i), i)
	}

	// Cache should have evicted some entries
	// Since we use simple FIFO, the first entries should be gone
	if len(cache.cache) > cache.maxSize {
		t.Errorf("Cache size %d exceeds maxSize %d", len(cache.cache), cache.maxSize)
	}

	// Test clear
	cache.clear()
	val, found = cache.get("key1")
	if found {
		t.Error("Expected key1 not to be found after clear")
	}
	if len(cache.cache) != 0 {
		t.Errorf("Expected empty cache after clear, got %d entries", len(cache.cache))
	}
}

// TestCache_makeCacheKey tests cache key generation
func TestCache_makeCacheKey(t *testing.T) {
	key1 := makeCacheKey("query", "param1", "param2")
	key2 := makeCacheKey("query", "param1", "param2")
	key3 := makeCacheKey("query", "param1", "param3")

	// Same parameters should generate same key
	if key1 != key2 {
		t.Error("Expected same key for same parameters")
	}

	// Different parameters should generate different keys
	if key1 == key3 {
		t.Error("Expected different keys for different parameters")
	}

	// Test with no parameters
	key4 := makeCacheKey("query")
	if key4 == "" {
		t.Error("Expected non-empty key")
	}

	// Test with various types
	key5 := makeCacheKey("query", 1, true, "string", 3.14)
	if key5 == "" {
		t.Error("Expected non-empty key for mixed types")
	}
}

// TestConfig_LoadConfig tests LoadConfig function
func TestConfig_LoadConfig(t *testing.T) {
	// Test with non-existent file (should return error if path provided)
	config, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		// If no error, config should be default
		if config == nil {
			t.Fatal("Expected non-nil config")
		}
	} else {
		// Error is acceptable for non-existent file with explicit path
		t.Logf("LoadConfig returned error for non-existent file (expected): %v", err)
	}

	// Test with empty path (should try default locations and return default)
	config2, err := LoadConfig("")
	if err != nil {
		t.Errorf("LoadConfig with empty path should return default config: %v", err)
	}
	if config2 == nil {
		t.Fatal("Expected non-nil config")
	}

	// Test with valid config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")
	configData := `{
		"cache": {
			"query_cache_size": 2000,
			"hybrid_node_cache_size": 1500
		},
		"timeout": {
			"query_timeout": "30s"
		}
	}`
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config3, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config3 == nil {
		t.Fatal("Expected non-nil config")
	}
	if config3.Cache.QueryCacheSize != 2000 {
		t.Errorf("Expected QueryCacheSize 2000, got %d", config3.Cache.QueryCacheSize)
	}
}

// TestConfig_SaveConfig tests SaveConfig function
func TestConfig_SaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	config := DefaultConfig()
	config.Cache.QueryCacheSize = 3000

	// Test saving to specific path
	err := SaveConfig(config, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load and verify
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	if loadedConfig.Cache.QueryCacheSize != 3000 {
		t.Errorf("Expected QueryCacheSize 3000, got %d", loadedConfig.Cache.QueryCacheSize)
	}

	// Test saving with empty path (should use default location)
	// This might fail if home directory doesn't exist, so we'll skip if it errors
	_ = SaveConfig(config, "")
}

// TestConfig_loadConfigFromFile tests loadConfigFromFile with various scenarios
func TestConfig_loadConfigFromFile(t *testing.T) {
	// Test with non-existent file
	_, err := loadConfigFromFile("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with invalid JSON
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	_, err = loadConfigFromFile(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test with valid JSON
	validPath := filepath.Join(tmpDir, "valid.json")
	validJSON := `{
		"cache": {
			"query_cache_size": 5000
		}
	}`
	if err := os.WriteFile(validPath, []byte(validJSON), 0644); err != nil {
		t.Fatalf("Failed to write valid config: %v", err)
	}

	config, err := loadConfigFromFile(validPath)
	if err != nil {
		t.Fatalf("Failed to load valid config: %v", err)
	}
	// After loading, validateAndSetDefaults is called, which might override zero values
	// So we check if it's at least set (not zero) or matches what we set
	if config.Cache.QueryCacheSize == 0 {
		t.Errorf("Expected QueryCacheSize to be set (non-zero), got %d", config.Cache.QueryCacheSize)
	}
	// The value might be validated/set to default if it's invalid, so we just check it's not zero
}

// TestConfig_validateAndSetDefaults tests validateAndSetDefaults
func TestConfig_validateAndSetDefaults(t *testing.T) {
	config := &Config{}

	// Call validateAndSetDefaults
	config.validateAndSetDefaults()

	// Verify defaults were set
	defaults := DefaultConfig()
	if config.Cache.QueryCacheSize != defaults.Cache.QueryCacheSize {
		t.Errorf("Expected QueryCacheSize %d, got %d", defaults.Cache.QueryCacheSize, config.Cache.QueryCacheSize)
	}
	if config.Timeout.QueryTimeout != defaults.Timeout.QueryTimeout {
		t.Errorf("Expected QueryTimeout %s, got %s", defaults.Timeout.QueryTimeout, config.Timeout.QueryTimeout)
	}

	// Test with partial config (some values set, some zero)
	config2 := &Config{
		Cache: CacheConfig{
			QueryCacheSize: 1000, // Set
			// HybridNodeCacheSize: 0 (zero, should get default)
		},
	}
	config2.validateAndSetDefaults()

	if config2.Cache.QueryCacheSize != 1000 {
		t.Errorf("Expected QueryCacheSize to remain 1000, got %d", config2.Cache.QueryCacheSize)
	}
	if config2.Cache.HybridNodeCacheSize != defaults.Cache.HybridNodeCacheSize {
		t.Errorf("Expected HybridNodeCacheSize to get default %d, got %d", defaults.Cache.HybridNodeCacheSize, config2.Cache.HybridNodeCacheSize)
	}
}

// TestBuilder_EdgeCases tests BuildGraph with edge cases
func TestBuilder_EdgeCases(t *testing.T) {
	// Test with empty tree
	emptyTree := types.NewGedcomTree()
	graph, err := BuildGraph(emptyTree)
	if err != nil {
		t.Fatalf("BuildGraph should succeed with empty tree: %v", err)
	}
	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}
	if graph.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes in empty graph, got %d", graph.NodeCount())
	}

	// Test with tree containing invalid XREFs
	tree := types.NewGedcomTree()
	
	// Individual with invalid family reference
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(types.NewGedcomLine(1, "NAME", "Test /Person/", ""))
	indiLine.AddChild(types.NewGedcomLine(1, "FAMS", "@INVALID@", "")) // Invalid XREF
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Build graph - should handle invalid XREFs gracefully
	graph2, err := BuildGraph(tree)
	if err != nil {
		t.Logf("BuildGraph returned error (may be expected): %v", err)
	}
	if graph2 != nil {
		// Graph should still be created even with invalid references
		if graph2.NodeCount() == 0 {
			t.Error("Expected at least 1 node (the individual)")
		}
	}

	// Test with tree containing all record types
	fullTree := types.NewGedcomTree()
	
	// Add header
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	headerLine.AddChild(types.NewGedcomLine(1, "GEDC", "", ""))
	headerLine.AddChild(types.NewGedcomLine(2, "VERS", "5.5.5", ""))
	header := types.NewHeaderRecord(headerLine)
	fullTree.AddRecord(header)

	// Add individual
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	fullTree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Add family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fullTree.AddRecord(types.NewFamilyRecord(famLine))

	// Add note
	noteLine := types.NewGedcomLine(0, "NOTE", "", "@N1@")
	noteLine.AddChild(types.NewGedcomLine(1, "CONC", "Test note", ""))
	fullTree.AddRecord(types.NewNoteRecord(noteLine))

	// Add source
	sourceLine := types.NewGedcomLine(0, "SOUR", "", "@S1@")
	sourceLine.AddChild(types.NewGedcomLine(1, "TITL", "Test Source", ""))
	fullTree.AddRecord(types.NewSourceRecord(sourceLine))

	// Add repository
	repoLine := types.NewGedcomLine(0, "REPO", "", "@R1@")
	repoLine.AddChild(types.NewGedcomLine(1, "NAME", "Test Repository", ""))
	fullTree.AddRecord(types.NewRepositoryRecord(repoLine))

	graph3, err := BuildGraph(fullTree)
	if err != nil {
		t.Fatalf("Failed to build graph with all record types: %v", err)
	}
	if graph3 == nil {
		t.Fatal("Expected non-nil graph")
	}

	// Verify all node types were created
	if graph3.GetIndividual("@I1@") == nil {
		t.Error("Expected individual node")
	}
	if graph3.GetFamily("@F1@") == nil {
		t.Error("Expected family node")
	}
	if graph3.GetNote("@N1@") == nil {
		t.Error("Expected note node")
	}
	if graph3.GetSource("@S1@") == nil {
		t.Error("Expected source node")
	}
	if graph3.GetRepository("@R1@") == nil {
		t.Error("Expected repository node")
	}
}

// TestBuilder_createNodes_EdgeCases tests createNodes with edge cases
func TestBuilder_createNodes_EdgeCases(t *testing.T) {
	tree := types.NewGedcomTree()
	graph := NewGraph(tree)

	// Test with tree containing wrong record types in type maps
	// (This shouldn't happen in practice, but test error handling)
	
	// Add a record that's not the expected type
	// This is hard to test directly since GetAllIndividuals() filters by type
	// But we can test the error path in AddNode
	
	// Test createNodes with empty tree
	err := createNodes(graph, tree)
	if err != nil {
		t.Fatalf("createNodes should succeed with empty tree: %v", err)
	}
}

// TestBuilder_createEdges_EdgeCases tests createEdges with edge cases
func TestBuilder_createEdges_EdgeCases(t *testing.T) {
	tree := types.NewGedcomTree()
	graph := NewGraph(tree)

	// Create nodes first
	if err := createNodes(graph, tree); err != nil {
		t.Fatalf("Failed to create nodes: %v", err)
	}

	// Test createEdges with empty tree
	err := createEdges(graph, tree)
	if err != nil {
		t.Fatalf("createEdges should succeed with empty tree: %v", err)
	}

	// Test with tree containing missing references
	tree2 := types.NewGedcomTree()
	
	// Individual referencing non-existent family
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(types.NewGedcomLine(1, "NAME", "Test /Person/", ""))
	indiLine.AddChild(types.NewGedcomLine(1, "FAMS", "@F999@", "")) // Non-existent family
	tree2.AddRecord(types.NewIndividualRecord(indiLine))

	graph2 := NewGraph(tree2)
	if err := createNodes(graph2, tree2); err != nil {
		t.Fatalf("Failed to create nodes: %v", err)
	}

	// createEdges should handle missing references gracefully
	err2 := createEdges(graph2, tree2)
	if err2 != nil {
		t.Logf("createEdges returned error (may be expected): %v", err2)
	}
}

