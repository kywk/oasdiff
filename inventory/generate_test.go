package inventory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func loadTestSpec(t *testing.T) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(filepath.Join("..", "data", "inventory", "sample-spec.yaml"))
	require.NoError(t, err)
	return spec
}

func loadTestConfig(t *testing.T) *Config {
	t.Helper()
	cfg, err := LoadConfig(filepath.Join("..", "data", "inventory", "config.yaml"))
	require.NoError(t, err)
	return cfg
}

func TestGenerate(t *testing.T) {
	spec := loadTestSpec(t)
	cfg := loadTestConfig(t)

	sheet, err := Generate(GenerateOptions{
		Config: cfg,
		Spec:   spec,
		Date:   "2024-01-15",
	})
	require.NoError(t, err)
	require.NotNil(t, sheet)

	// Sample spec has 6 operations:
	// GET /users, POST /users, GET /users/{id}, DELETE /users/{id}, GET /orders, POST /orders
	require.Equal(t, 6, len(sheet.Records))
	require.Equal(t, 6, sheet.MaxSerial)

	// Check first record
	r := sheet.Records[0]
	require.Equal(t, "API-001", r.Serial)
	require.Equal(t, "2024-01-15", r.Date)
	require.Equal(t, ChangeTypeNew, r.ChangeType)
	require.Equal(t, "GET", r.Method)
	require.Equal(t, "/users", r.Path)
	require.Equal(t, "GET", r.Values["method"])
	require.Equal(t, "/users", r.Values["path"])
	require.Equal(t, "listUsers", r.Values["operationId"])
	require.Equal(t, "List all users", r.Values["summary"])
	require.Equal(t, "false", r.Values["x-transaction"])

	// Check a POST with x-transaction=true
	r = sheet.Records[1]
	require.Equal(t, "API-002", r.Serial)
	require.Equal(t, "POST", r.Method)
	require.Equal(t, "/users", r.Path)
	require.Equal(t, "true", r.Values["x-transaction"])
}

func TestGenerateCSVOutput(t *testing.T) {
	spec := loadTestSpec(t)
	cfg := loadTestConfig(t)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "inventory.csv")

	_, err := Generate(GenerateOptions{
		Config:     cfg,
		Spec:       spec,
		OutputPath: outputPath,
		Date:       "2024-01-15",
	})
	require.NoError(t, err)

	// Verify file exists and is non-empty
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0))

	// Read it back
	sheet, err := ReadCSV(outputPath, cfg)
	require.NoError(t, err)
	require.Equal(t, 6, len(sheet.Records))
	require.Equal(t, "API-001", sheet.Records[0].Serial)
}

func TestGenerateExcelOutput(t *testing.T) {
	spec := loadTestSpec(t)
	cfg := loadTestConfig(t)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "inventory.xlsx")

	_, err := Generate(GenerateOptions{
		Config:     cfg,
		Spec:       spec,
		OutputPath: outputPath,
		Format:     OutputExcel,
		Date:       "2024-01-15",
	})
	require.NoError(t, err)

	// Verify file exists
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0))

	// Read it back
	sheet, err := ReadExcel(outputPath, cfg)
	require.NoError(t, err)
	require.Equal(t, 6, len(sheet.Records))
}

func TestGenerateNoSpec(t *testing.T) {
	cfg := loadTestConfig(t)
	_, err := Generate(GenerateOptions{
		Config: cfg,
		Spec:   nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "OpenAPI spec is required")
}

func TestGenerateNoConfig(t *testing.T) {
	spec := loadTestSpec(t)
	_, err := Generate(GenerateOptions{
		Config: nil,
		Spec:   spec,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "config is required")
}
