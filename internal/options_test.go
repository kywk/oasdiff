package internal_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func TestOptionsCmd_Default(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "options", "../data/openapi-test1.yaml"}, &stdout, &stderr)
	require.Equal(t, 0, code)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(stdout.Bytes())
	require.NoError(t, err)

	// In openapi-test1.yaml:
	// /api/{domain}/{project}/badges/security-score has GET
	// /subscribe has POST
	// /register has POST
	require.Nil(t, doc.Paths.Find("/api/{domain}/{project}/badges/security-score").Options)
	require.NotNil(t, doc.Paths.Find("/subscribe").Options)
	require.NotNil(t, doc.Paths.Find("/register").Options)
	require.Equal(t, "CORS preflight", doc.Paths.Find("/subscribe").Options.Summary)
}

func TestOptionsCmd_Alias(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "add-options", "../data/openapi-test1.yaml"}, &stdout, &stderr)
	require.Equal(t, 0, code)
	require.Contains(t, stdout.String(), "options:")
}

func TestOptionsCmd_OutputFlag(t *testing.T) {
	tempDir := t.TempDir()
	outFile := filepath.Join(tempDir, "out.yaml")

	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "options", "../data/openapi-test1.yaml", "-o", outFile}, &stdout, &stderr)
	require.Equal(t, 0, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "Added OPTIONS to")

	data, err := os.ReadFile(outFile)
	require.NoError(t, err)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	require.NoError(t, err)
	require.NotNil(t, doc.Paths.Find("/subscribe").Options)
}

func TestOptionsCmd_InPlace(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "spec.yaml")

	// Read original and write to tempFile
	original, err := os.ReadFile("../data/openapi-test1.yaml")
	require.NoError(t, err)
	err = os.WriteFile(targetFile, original, 0644)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "options", targetFile, "-i"}, &stdout, &stderr)
	require.Equal(t, 0, code)
	require.Contains(t, stderr.String(), "Added OPTIONS to")

	data, err := os.ReadFile(targetFile)
	require.NoError(t, err)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	require.NoError(t, err)
	require.NotNil(t, doc.Paths.Find("/subscribe").Options)
}

func TestOptionsCmd_InPlace_ConflictWithOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "options", "../data/openapi-test1.yaml", "-i", "-o", "foo.yaml"}, &stdout, &stderr)
	require.NotEqual(t, 0, code)
	require.Contains(t, stderr.String(), "cannot use both --output and --in-place")
}

func TestOptionsCmd_InPlace_InvalidSource(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "options", "-", "-i"}, &stdout, &stderr)
	require.NotEqual(t, 0, code)
	require.Contains(t, stderr.String(), "--in-place can only be used with existing local files")
}

func TestOptionsCmd_FormatJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "options", "../data/openapi-test1.yaml", "-f", "json"}, &stdout, &stderr)
	require.Equal(t, 0, code)

	var parsed map[string]any
	err := json.Unmarshal(stdout.Bytes(), &parsed)
	require.NoError(t, err)
	require.Equal(t, "3.0.3", parsed["openapi"])
}

func TestOptionsCmd_AllFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "options", "../data/openapi-test1.yaml", "--all"}, &stdout, &stderr)
	require.Equal(t, 0, code)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(stdout.Bytes())
	require.NoError(t, err)

	// With --all, even GET-only endpoint gets OPTIONS
	require.NotNil(t, doc.Paths.Find("/api/{domain}/{project}/badges/security-score").Options)
}

func TestOptionsCmd_CustomMethods(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := internal.Run([]string{"oasdiff", "options", "../data/openapi-test1.yaml", "--methods", "PUT"}, &stdout, &stderr)
	require.Equal(t, 0, code)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(stdout.Bytes())
	require.NoError(t, err)

	// Since openapi-test1 has no PUT, no OPTIONS should be added
	require.Nil(t, doc.Paths.Find("/subscribe").Options)
	require.Nil(t, doc.Paths.Find("/register").Options)
}
