package inventory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigYAML(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join("..", "data", "inventory", "config.yaml"))
	require.NoError(t, err)
	require.Equal(t, "API", cfg.Prefix)
	require.Equal(t, "csv", cfg.Output)
	require.Len(t, cfg.Columns, 8)
	require.Equal(t, "method", cfg.Columns[0].Source)
	require.Equal(t, "HTTP方法", cfg.Columns[0].Header)
	require.Equal(t, "x-transaction", cfg.Columns[7].Source)
	require.Equal(t, "boolean", cfg.Columns[7].Type)
}

func TestLoadConfigJSON(t *testing.T) {
	// Create a temp JSON config
	content := `{
		"prefix": "SVC",
		"output": "excel",
		"columns": [
			{"source": "method", "header": "Method"},
			{"source": "path", "header": "Path"}
		]
	}`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	require.Equal(t, "SVC", cfg.Prefix)
	require.Equal(t, "excel", cfg.Output)
	require.Len(t, cfg.Columns, 2)
}

func TestLoadConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name:    "missing prefix",
			content: `{"prefix": "", "columns": [{"source": "method", "header": "Method"}]}`,
			errMsg:  "prefix is required",
		},
		{
			name:    "no columns",
			content: `{"prefix": "API", "columns": []}`,
			errMsg:  "at least one column is required",
		},
		{
			name:    "column missing source",
			content: `{"prefix": "API", "columns": [{"source": "", "header": "X"}]}`,
			errMsg:  "column[0] source is required",
		},
		{
			name:    "column missing header",
			content: `{"prefix": "API", "columns": [{"source": "method", "header": ""}]}`,
			errMsg:  "column[0] header is required",
		},
		{
			name:    "invalid output format",
			content: `{"prefix": "API", "output": "pdf", "columns": [{"source": "method", "header": "M"}]}`,
			errMsg:  "output must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "config.json")
			require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o644))

			_, err := LoadConfig(path)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestHasComplexColumns(t *testing.T) {
	cfg := &Config{
		Prefix:  "API",
		Columns: []ColumnConfig{{Source: "method", Header: "M"}, {Source: "path", Header: "P"}},
	}
	require.False(t, cfg.HasComplexColumns())

	cfg.Columns = append(cfg.Columns, ColumnConfig{Source: "requestBody", Header: "RB"})
	require.True(t, cfg.HasComplexColumns())
}

func TestCSVColumns(t *testing.T) {
	cfg := &Config{
		Prefix: "API",
		Columns: []ColumnConfig{
			{Source: "method", Header: "M"},
			{Source: "requestBody", Header: "RB"},
			{Source: "path", Header: "P"},
			{Source: "responses", Header: "R"},
		},
	}
	csvCols := cfg.CSVColumns()
	require.Len(t, csvCols, 2)
	require.Equal(t, "method", csvCols[0].Source)
	require.Equal(t, "path", csvCols[1].Source)
}
