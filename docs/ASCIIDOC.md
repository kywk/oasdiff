# Asciidoc Output Format

oasdiff supports [Asciidoc](https://asciidoc.org/) as an output format for the `diff` and `changelog` commands.

## Usage

Use the `--format asciidoc` (or `-f asciidoc`) flag:

```bash
# Full diff in Asciidoc
oasdiff diff base.yaml revision.yaml -f asciidoc

# Changelog in Asciidoc
oasdiff changelog base.yaml revision.yaml -f asciidoc

# Breaking changes in Asciidoc
oasdiff breaking base.yaml revision.yaml -f asciidoc

# Write output to a file
oasdiff changelog base.yaml revision.yaml -f asciidoc > changelog.adoc
```

## Custom Templates

You can provide a custom Asciidoc template for changelog output:

```bash
oasdiff changelog base.yaml revision.yaml --template my-template.adoc -f asciidoc
```

The template uses Go's `text/template` syntax and has access to the same data and functions as the [markdown and html templates](CHANGELOG-TEMPLATE.md):

- `.GroupedChanges` — Changes grouped by endpoint or section
- `.BaseVersion` / `.RevisionVersion` — Spec version strings
- `.GetVersionTitle()` — Formatted version comparison (e.g. `1.0.0 vs. 2.0.0`)
- `pathGroups .GroupedChanges` — API path changes sorted by path and operation
- `sectionGroups .GroupedChanges` — Security and component changes sorted by section
- `capitalize "string"` — Capitalizes the first letter of a string

### Minimal Custom Template Example

```asciidoc
= API Changelog {{ .GetVersionTitle }}
{{ if .GroupedChanges }}
{{ with pathGroups .GroupedChanges }}
== API Changes
{{ range . }}
=== {{ .Group.Operation }} {{ .Group.Path }}
{{ range .Changes }}* {{ if .IsBreaking }}*[BREAKING]* {{ end }}{{ .Text }}
{{ end }}
{{ end }}
{{ end }}
{{ range sectionGroups .GroupedChanges }}
== {{ capitalize .Group.Section }}
{{ range .Changes }}* {{ if .IsBreaking }}*[BREAKING]* {{ end }}{{ .Text }}
{{ end }}
{{ end }}
{{ else }}
No changes detected
{{ end }}
```

## Output Structure

### Diff (`oasdiff diff -f asciidoc`)

The diff output uses Asciidoc section headings and nested unordered lists:

```asciidoc
=== New Endpoints: 1
* GET /pets/{petId}

=== Deleted Endpoints: 1
* POST /pets

=== Modified Endpoints: 2
* PUT /pets
** Request body changed
*** Content changed
```

### Changelog (`oasdiff changelog -f asciidoc`)

The changelog output groups changes by endpoint with breaking change markers:

```asciidoc
= API Changelog 1.0.0 vs. 2.0.0

== API Changes

=== GET /pets
* *[BREAKING]* removed the query parameter 'limit'
* added the query parameter 'page_size'

=== POST /pets
* request body became optional
```
