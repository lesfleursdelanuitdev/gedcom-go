package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the CLI configuration
type Config struct {
	Parser struct {
		Type     string `json:"type"` // hierarchical, parallel, stream
		Parallel bool   `json:"parallel"`
		Stream   bool   `json:"stream"`
	} `json:"parser"`
	Validation struct {
		SeverityThreshold string `json:"severity_threshold"` // severe, warning, info, hint
		StrictMode        bool   `json:"strict_mode"`
	} `json:"validation"`
	Output struct {
		DefaultFormat string `json:"default_format"` // table, json, yaml, csv
		Color         bool   `json:"color"`
		Progress      bool   `json:"progress"`
	} `json:"output"`
	Graph struct {
		CacheSize     int  `json:"cache_size"`
		EnableIndexes bool `json:"enable_indexes"`
	} `json:"graph"`
	Export struct {
		PrettyPrint bool `json:"pretty_print"`
		Indent      int  `json:"indent"`
	} `json:"export"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	config := &Config{}
	config.Parser.Type = "hierarchical"
	config.Parser.Parallel = true
	config.Parser.Stream = false
	config.Validation.SeverityThreshold = "warning"
	config.Validation.StrictMode = false
	config.Output.DefaultFormat = "table"
	config.Output.Color = true
	config.Output.Progress = true
	config.Graph.CacheSize = 1000
	config.Graph.EnableIndexes = true
	config.Export.PrettyPrint = true
	config.Export.Indent = 2
	return config
}

// LoadConfig loads configuration from file or returns default
func LoadConfig(configPath string) (*Config, error) {
	// If no path provided, try default locations
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return DefaultConfig(), nil
		}

		// Try ~/.gedcom/config.json
		configPath = filepath.Join(homeDir, ".gedcom", "config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Try ~/.config/gedcom/config.json
			configPath = filepath.Join(homeDir, ".config", "gedcom", "config.json")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Return default if no config found
				return DefaultConfig(), nil
			}
		}
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// If no path provided, use default location
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".gedcom", "config.json")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
