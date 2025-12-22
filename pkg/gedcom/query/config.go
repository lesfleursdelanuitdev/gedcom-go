package query

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the configuration for query operations
type Config struct {
	// Cache configuration
	Cache CacheConfig `json:"cache"`

	// Timeout configuration
	Timeout TimeoutConfig `json:"timeout"`

	// Database configuration
	Database DatabaseConfig `json:"database"`
}

// CacheConfig holds cache size configurations
type CacheConfig struct {
	// Hybrid cache sizes
	HybridNodeCacheSize  int `json:"hybrid_node_cache_size"`  // Default: 50000
	HybridXrefCacheSize  int `json:"hybrid_xref_cache_size"`  // Default: 25000
	HybridQueryCacheSize int `json:"hybrid_query_cache_size"` // Default: 5000

	// Query cache size (for in-memory graph)
	QueryCacheSize int `json:"query_cache_size"` // Default: 1000
}

// TimeoutConfig holds timeout configurations
type TimeoutConfig struct {
	// Database operation timeouts
	SQLiteQueryTimeout time.Duration `json:"sqlite_query_timeout"` // Default: 30s
	BadgerDBTimeout   time.Duration `json:"badgerdb_timeout"`     // Default: 10s

	// Graph building timeout
	BuildTimeout time.Duration `json:"build_timeout"` // Default: 5m

	// Query execution timeout
	QueryTimeout time.Duration `json:"query_timeout"` // Default: 1m
}

// UnmarshalJSON implements custom JSON unmarshaling for TimeoutConfig
// to support duration strings like "30s", "5m", "1h", etc.
func (tc *TimeoutConfig) UnmarshalJSON(data []byte) error {
	// First, try to unmarshal as a map
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// Parse each duration field
	if v, ok := m["sqlite_query_timeout"]; ok {
		if d, err := parseDuration(v); err == nil {
			tc.SQLiteQueryTimeout = d
		}
	}
	if v, ok := m["badgerdb_timeout"]; ok {
		if d, err := parseDuration(v); err == nil {
			tc.BadgerDBTimeout = d
		}
	}
	if v, ok := m["build_timeout"]; ok {
		if d, err := parseDuration(v); err == nil {
			tc.BuildTimeout = d
		}
	}
	if v, ok := m["query_timeout"]; ok {
		if d, err := parseDuration(v); err == nil {
			tc.QueryTimeout = d
		}
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for TimeoutConfig
// to output duration as human-readable strings
func (tc TimeoutConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		SQLiteQueryTimeout string `json:"sqlite_query_timeout"`
		BadgerDBTimeout    string `json:"badgerdb_timeout"`
		BuildTimeout       string `json:"build_timeout"`
		QueryTimeout       string `json:"query_timeout"`
	}{
		SQLiteQueryTimeout: tc.SQLiteQueryTimeout.String(),
		BadgerDBTimeout:    tc.BadgerDBTimeout.String(),
		BuildTimeout:       tc.BuildTimeout.String(),
		QueryTimeout:       tc.QueryTimeout.String(),
	})
}

// parseDuration parses a duration from various formats:
// - String: "30s", "5m", "1h", etc.
// - Number: nanoseconds (int64 or float64)
func parseDuration(v interface{}) (time.Duration, error) {
	switch val := v.(type) {
	case string:
		// Try parsing as duration string
		return time.ParseDuration(val)
	case float64:
		// JSON numbers are float64, convert to int64 nanoseconds
		return time.Duration(int64(val)), nil
	case int64:
		// Already nanoseconds
		return time.Duration(val), nil
	case int:
		// Already nanoseconds
		return time.Duration(val), nil
	default:
		return 0, fmt.Errorf("cannot parse duration from type %T", v)
	}
}

// DatabaseConfig holds database-specific configurations
type DatabaseConfig struct {
	// SQLite configuration
	SQLiteMaxOpenConns int `json:"sqlite_max_open_conns"` // Default: 10
	SQLiteMaxIdleConns int `json:"sqlite_max_idle_conns"` // Default: 5

	// BadgerDB configuration
	BadgerDBValueLogFileSize int64 `json:"badgerdb_value_log_file_size"` // Default: 1GB
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Cache: CacheConfig{
			HybridNodeCacheSize:  50000,
			HybridXrefCacheSize:  25000,
			HybridQueryCacheSize: 5000,
			QueryCacheSize:       1000,
		},
		Timeout: TimeoutConfig{
			SQLiteQueryTimeout: 30 * time.Second,
			BadgerDBTimeout:    10 * time.Second,
			BuildTimeout:       5 * time.Minute,
			QueryTimeout:       1 * time.Minute,
		},
		Database: DatabaseConfig{
			SQLiteMaxOpenConns:    10,
			SQLiteMaxIdleConns:    5,
			BadgerDBValueLogFileSize: 1 << 30, // 1GB
		},
	}
}

// LoadConfig loads configuration from file or returns default
// It searches for config files in the following order:
// 1. The provided configPath (if not empty)
// 2. ./gedcom-query-config.json (current directory)
// 3. ~/.gedcom/query-config.json (user home)
// 4. ~/.config/gedcom/query-config.json (XDG config)
func LoadConfig(configPath string) (*Config, error) {
	// If path provided, use it directly
	if configPath != "" {
		return loadConfigFromFile(configPath)
	}

	// Try current directory
	if config, err := loadConfigFromFile("./gedcom-query-config.json"); err == nil {
		return config, nil
	}

	// Try user home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		// Try ~/.gedcom/query-config.json
		if config, err := loadConfigFromFile(filepath.Join(homeDir, ".gedcom", "query-config.json")); err == nil {
			return config, nil
		}

		// Try ~/.config/gedcom/query-config.json
		if config, err := loadConfigFromFile(filepath.Join(homeDir, ".config", "gedcom", "query-config.json")); err == nil {
			return config, nil
		}
	}

	// Return default if no config found
	return DefaultConfig(), nil
}

// loadConfigFromFile loads configuration from a specific file
func loadConfigFromFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", configPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and apply defaults for zero values
	config.validateAndSetDefaults()

	return config, nil
}

// validateAndSetDefaults ensures all values are set to defaults if zero
func (c *Config) validateAndSetDefaults() {
	defaults := DefaultConfig()

	if c.Cache.HybridNodeCacheSize <= 0 {
		c.Cache.HybridNodeCacheSize = defaults.Cache.HybridNodeCacheSize
	}
	if c.Cache.HybridXrefCacheSize <= 0 {
		c.Cache.HybridXrefCacheSize = defaults.Cache.HybridXrefCacheSize
	}
	if c.Cache.HybridQueryCacheSize <= 0 {
		c.Cache.HybridQueryCacheSize = defaults.Cache.HybridQueryCacheSize
	}
	if c.Cache.QueryCacheSize <= 0 {
		c.Cache.QueryCacheSize = defaults.Cache.QueryCacheSize
	}

	if c.Timeout.SQLiteQueryTimeout <= 0 {
		c.Timeout.SQLiteQueryTimeout = defaults.Timeout.SQLiteQueryTimeout
	}
	if c.Timeout.BadgerDBTimeout <= 0 {
		c.Timeout.BadgerDBTimeout = defaults.Timeout.BadgerDBTimeout
	}
	if c.Timeout.BuildTimeout <= 0 {
		c.Timeout.BuildTimeout = defaults.Timeout.BuildTimeout
	}
	if c.Timeout.QueryTimeout <= 0 {
		c.Timeout.QueryTimeout = defaults.Timeout.QueryTimeout
	}

	if c.Database.SQLiteMaxOpenConns <= 0 {
		c.Database.SQLiteMaxOpenConns = defaults.Database.SQLiteMaxOpenConns
	}
	if c.Database.SQLiteMaxIdleConns <= 0 {
		c.Database.SQLiteMaxIdleConns = defaults.Database.SQLiteMaxIdleConns
	}
	if c.Database.BadgerDBValueLogFileSize <= 0 {
		c.Database.BadgerDBValueLogFileSize = defaults.Database.BadgerDBValueLogFileSize
	}
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// If no path provided, use default location
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".gedcom", "query-config.json")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a custom struct for marshaling to ensure durations are strings
	type ConfigJSON struct {
		Cache    CacheConfig    `json:"cache"`
		Timeout  TimeoutConfig  `json:"timeout"`
		Database DatabaseConfig `json:"database"`
	}
	configJSON := ConfigJSON{
		Cache:    config.Cache,
		Timeout:  config.Timeout,
		Database: config.Database,
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(configJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

