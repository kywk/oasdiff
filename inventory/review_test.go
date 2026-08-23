package inventory

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestReview(t *testing.T) {
	spec := loadTestSpec(t)
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Review(cfg, spec, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Check total count
	require.Contains(t, output, "Total APIs: 6")

	// Check tag grouping
	require.Contains(t, output, "users: 4 APIs")
	require.Contains(t, output, "orders: 2 APIs")
	require.Contains(t, output, "admin: 1 APIs")

	// Check method grouping
	require.Contains(t, output, "GET: 3 APIs")
	require.Contains(t, output, "POST: 2 APIs")
	require.Contains(t, output, "DELETE: 1 APIs")

	// Check x-transaction statistics
	require.Contains(t, output, "x-transaction=true: 3 APIs")
	require.Contains(t, output, "x-transaction=false: 3 APIs")

	// Check that specific APIs are listed under their extension values
	require.Contains(t, output, "POST /users")
	require.Contains(t, output, "DELETE /users/{id}")
}

func TestReviewEmptySpec(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "Empty", Version: "1.0.0"},
		Paths:   &openapi3.Paths{},
	}
	cfg := loadTestConfig(t)

	var buf bytes.Buffer
	err := Review(cfg, spec, &buf)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "No API endpoints found")
}

func TestReviewNoExtensions(t *testing.T) {
	// Config that only asks for standard fields, no x- extensions
	cfg := &Config{
		Prefix: "API",
		Output: OutputCSV,
		Columns: []ColumnConfig{
			{Source: "method", Header: "Method"},
			{Source: "path", Header: "Path"},
			{Source: "summary", Header: "Summary"},
		},
	}

	spec := loadTestSpec(t)

	var buf bytes.Buffer
	err := Review(cfg, spec, &buf)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Total APIs: 6")
	// Should not contain x- extension section
	require.NotContains(t, output, "x-transaction")
}

func TestReviewMultipleExtensions(t *testing.T) {
	// Create a spec with multiple custom extensions
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(filepath.Join("..", "data", "inventory", "sample-spec.yaml"))
	require.NoError(t, err)

	cfg := &Config{
		Prefix: "API",
		Output: OutputCSV,
		Columns: []ColumnConfig{
			{Source: "method", Header: "Method"},
			{Source: "path", Header: "Path"},
			{Source: "x-transaction", Header: "交易類別", Type: "boolean"},
		},
	}

	var buf bytes.Buffer
	err = Review(cfg, spec, &buf)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "交易類別 (x-transaction)")
	require.Contains(t, output, "x-transaction=true: 3 APIs")
	require.Contains(t, output, "x-transaction=false: 3 APIs")
}
