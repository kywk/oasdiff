package formatters

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	_ "embed"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/report"
)

type AsciidocFormatter struct {
	notImplementedFormatter
	Localizer       checker.Localizer
	BaseVersion     string
	RevisionVersion string
}

func newAsciidocFormatter(l checker.Localizer, baseVersion, revisionVersion string) AsciidocFormatter {
	return AsciidocFormatter{
		Localizer:       l,
		BaseVersion:     baseVersion,
		RevisionVersion: revisionVersion,
	}
}

func (f AsciidocFormatter) RenderDiff(diff *diff.Diff, opts RenderOpts) ([]byte, error) {
	return []byte(report.GetAsciidocReportAsString(diff)), nil
}

//go:embed templates/changelog.adoc
var changelogAsciidoc string

func (f AsciidocFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts) ([]byte, error) {
	var tmpl *template.Template
	var err error

	if opts.TemplatePath != "" {
		tmpl, err = f.loadCustomTemplate(opts.TemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom template: %w", err)
		}
	} else {
		tmpl = template.Must(template.New("changelog").Funcs(AsciidocTemplateFuncs()).Parse(changelogAsciidoc))
	}

	return ExecuteTextTemplate(tmpl, GroupChanges(changes, f.Localizer), f.BaseVersion, f.RevisionVersion, opts.DiffEmpty, opts.IsBreaking)
}

func (f AsciidocFormatter) loadCustomTemplate(templatePath string) (*template.Template, error) {
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	tmpl, err := template.New("custom-changelog").Funcs(AsciidocTemplateFuncs()).Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// AsciidocTemplateFuncs returns the FuncMap available to Asciidoc changelog templates.
func AsciidocTemplateFuncs() template.FuncMap {
	return template.FuncMap(changelogTemplateFuncs())
}

func ExecuteAsciidocTemplate(tmpl *template.Template, changes ChangesByGroup, baseVersion, revisionVersion string, diffEmpty, isBreaking bool) ([]byte, error) {
	var out bytes.Buffer
	if err := tmpl.Execute(&out, TemplateData{changes, baseVersion, revisionVersion, diffEmpty, isBreaking}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (f AsciidocFormatter) SupportedOutputs() []Output {
	return []Output{OutputDiff, OutputChangelog}
}

func (f AsciidocFormatter) SupportsTemplate() bool {
	return true
}
