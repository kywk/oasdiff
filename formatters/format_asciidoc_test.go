package formatters_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var asciidocFormatter = formatters.AsciidocFormatter{
	Localizer: MockLocalizer,
}

func TestAsciidocLookup(t *testing.T) {
	f, err := formatters.Lookup(string(formatters.FormatAsciidoc), formatters.DefaultFormatterOpts())
	require.NoError(t, err)
	require.IsType(t, formatters.AsciidocFormatter{}, f)
}

func TestAsciidocFormatter_RenderDiff(t *testing.T) {
	out, err := asciidocFormatter.RenderDiff(nil, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Contains(t, string(out), "No changes")
}

func TestAsciidocFormatter_RenderChangelog_NoVersions(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Path:      "/test",
			Operation: "GET",
			Id:        "change_id",
			Level:     checker.ERR,
		},
	}

	out, err := asciidocFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Contains(t, string(out), "= API Changelog")
	require.NotContains(t, string(out), "vs.")
}

func TestAsciidocFormatter_RenderChangelog_NoBaseVersion(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Path:      "/test",
			Operation: "GET",
			Id:        "change_id",
			Level:     checker.ERR,
		},
	}

	f := formatters.AsciidocFormatter{Localizer: asciidocFormatter.Localizer, RevisionVersion: "2.0.0"}
	out, err := f.RenderChangelog(testChanges, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Contains(t, string(out), "= API Changelog")
	require.NotContains(t, string(out), "vs.")
}

func TestAsciidocFormatter_RenderChangelog_WithVersions(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Path:      "/test",
			Operation: "GET",
			Id:        "change_id",
			Level:     checker.ERR,
		},
	}

	f := formatters.AsciidocFormatter{Localizer: asciidocFormatter.Localizer, BaseVersion: "1.0.0", RevisionVersion: "2.0.0"}
	out, err := f.RenderChangelog(testChanges, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Contains(t, string(out), "= API Changelog 1.0.0 vs. 2.0.0")
}

func TestAsciidocFormatter_RenderChangelog_BreakingChange(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Path:      "/test",
			Operation: "GET",
			Id:        "change_id",
			Level:     checker.ERR,
		},
	}

	out, err := asciidocFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts())
	require.NoError(t, err)
	require.Contains(t, string(out), "*[BREAKING]*")
}

func TestAsciidocFormatter_RenderChangelog_SecurityAndComponentChanges(t *testing.T) {
	testChanges := checker.Changes{
		checker.SecurityChange{
			Id:    "change_id",
			Level: checker.INFO,
		},
		checker.ComponentChange{
			Id:        "change_id",
			Level:     checker.INFO,
			Component: "securitySchemes",
		},
	}

	out, err := asciidocFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts())
	require.NoError(t, err)

	result := string(out)
	require.Contains(t, result, "== Security")
	require.Contains(t, result, "== Components")
}

func TestAsciidocFormatter_NotImplemented(t *testing.T) {
	var err error

	_, err = asciidocFormatter.RenderChecks(formatters.Checks{}, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = asciidocFormatter.RenderFlatten(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = asciidocFormatter.RenderSummary(nil, formatters.NewRenderOpts())
	assert.Error(t, err)
}

func TestAsciidocFormatter_RenderChangelog_WithCustomTemplate(t *testing.T) {
	customTemplate := `= Changes for {{ .GetVersionTitle }}
{{ range $endpoint, $changes := .APIChanges }}
== {{ $endpoint.Operation }} {{ $endpoint.Path }}
{{ range $changes }}* {{ if .IsBreaking }}*BREAKING* {{ end }}{{ .Text }}
{{ end }}
{{ end }}`

	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "custom-changelog.adoc")
	err := os.WriteFile(templatePath, []byte(customTemplate), 0644)
	require.NoError(t, err)

	testChanges := checker.Changes{
		checker.ApiChange{
			Path:      "/test",
			Operation: "GET",
			Id:        "change_id",
			Level:     checker.ERR,
		},
	}

	opts := formatters.NewRenderOpts()
	opts.TemplatePath = templatePath

	f := formatters.AsciidocFormatter{Localizer: asciidocFormatter.Localizer, BaseVersion: "1.0.0", RevisionVersion: "2.0.0"}
	out, err := f.RenderChangelog(testChanges, opts)
	require.NoError(t, err)

	result := string(out)
	require.Contains(t, result, "= Changes for 1.0.0 vs. 2.0.0")
	require.Contains(t, result, "== GET /test")
	require.Contains(t, result, "* ")
}

func TestAsciidocFormatter_RenderChangelog_CustomTemplateNotFound(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Path:      "/test",
			Operation: "GET",
			Id:        "change_id",
			Level:     checker.ERR,
		},
	}

	opts := formatters.NewRenderOpts()
	opts.TemplatePath = "/nonexistent/template.adoc"

	f := formatters.AsciidocFormatter{Localizer: asciidocFormatter.Localizer, BaseVersion: "1.0.0", RevisionVersion: "2.0.0"}
	_, err := f.RenderChangelog(testChanges, opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load custom template")
}

func TestAsciidocFormatter_RenderChangelog_InvalidCustomTemplate(t *testing.T) {
	invalidTemplate := `{{ .InvalidField `

	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "invalid-template.adoc")
	err := os.WriteFile(templatePath, []byte(invalidTemplate), 0644)
	require.NoError(t, err)

	testChanges := checker.Changes{
		checker.ApiChange{
			Path:      "/test",
			Operation: "GET",
			Id:        "change_id",
			Level:     checker.ERR,
		},
	}

	opts := formatters.NewRenderOpts()
	opts.TemplatePath = templatePath

	f := formatters.AsciidocFormatter{Localizer: asciidocFormatter.Localizer, BaseVersion: "1.0.0", RevisionVersion: "2.0.0"}
	_, err = f.RenderChangelog(testChanges, opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load custom template")
}

func TestAsciidocFormatter_SupportsTemplate(t *testing.T) {
	require.True(t, asciidocFormatter.SupportsTemplate())
}

func TestAsciidocFormatter_SupportedOutputs(t *testing.T) {
	outputs := asciidocFormatter.SupportedOutputs()
	require.Contains(t, outputs, formatters.OutputDiff)
	require.Contains(t, outputs, formatters.OutputChangelog)
}
