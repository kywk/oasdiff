package inventory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config defines the inventory configuration for field extraction and output.
type Config struct {
	Prefix  string         `yaml:"prefix" json:"prefix"`
	Output  string         `yaml:"output" json:"output"`
	Columns []ColumnConfig `yaml:"columns" json:"columns"`
}

// ColumnConfig defines a single column mapping from OpenAPI to inventory sheet.
type ColumnConfig struct {
	Source string `yaml:"source" json:"source"` // OpenAPI field name or x- extension
	Header string `yaml:"header" json:"header"` // Display header in inventory sheet
	Type   string `yaml:"type" json:"type"`     // Optional: boolean, string, array, etc.
}

// OutputFormat enumerates supported output formats.
const (
	OutputCSV   = "csv"
	OutputExcel = "excel"
)

// LoadConfig loads an inventory config from a YAML or JSON file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	cfg := &Config{}

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := unmarshalYAML(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := unmarshalYAML(data, cfg); err != nil {
			if err2 := json.Unmarshal(data, cfg); err2 != nil {
				return nil, fmt.Errorf("failed to parse config (tried YAML and JSON): YAML error: %w", err)
			}
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that the config has all required fields.
func (c *Config) Validate() error {
	if c.Prefix == "" {
		return fmt.Errorf("config: prefix is required")
	}
	if len(c.Columns) == 0 {
		return fmt.Errorf("config: at least one column is required")
	}
	for i, col := range c.Columns {
		if col.Source == "" {
			return fmt.Errorf("config: column[%d] source is required", i)
		}
		if col.Header == "" {
			return fmt.Errorf("config: column[%d] header is required", i)
		}
	}
	// Default output format
	if c.Output == "" {
		c.Output = OutputCSV
	}
	if c.Output != OutputCSV && c.Output != OutputExcel {
		return fmt.Errorf("config: output must be %q or %q, got %q", OutputCSV, OutputExcel, c.Output)
	}
	return nil
}

// HasComplexColumns returns true if the config contains columns that are only
// supported in Excel format (requestBody, responses).
func (c *Config) HasComplexColumns() bool {
	for _, col := range c.Columns {
		if col.Source == "requestBody" || col.Source == "responses" {
			return true
		}
	}
	return false
}

// CSVColumns returns columns that are suitable for CSV output
// (excludes requestBody and responses).
func (c *Config) CSVColumns() []ColumnConfig {
	result := make([]ColumnConfig, 0, len(c.Columns))
	for _, col := range c.Columns {
		if col.Source != "requestBody" && col.Source != "responses" {
			result = append(result, col)
		}
	}
	return result
}
