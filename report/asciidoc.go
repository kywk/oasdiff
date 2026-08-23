package report

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

// GetAsciidocReportAsString returns an Asciidoc diff report as a string
func GetAsciidocReportAsString(d *diff.Diff) string {
	var buf bytes.Buffer
	r := asciidocReport{
		Writer: &buf,
	}
	r.output(d)

	return buf.String()
}

type asciidocReport struct {
	Writer io.Writer
	level  int
}

func (r *asciidocReport) indent() *asciidocReport {
	return &asciidocReport{
		Writer: r.Writer,
		level:  r.level + 1,
	}
}

func (r *asciidocReport) print(output ...any) {
	fmt.Fprintln(r.Writer, r.addPrefix(output)...) //nolint:errcheck
}

func (r *asciidocReport) addPrefix(output []any) []any {
	return append(r.getPrefix(), output...)
}

func (r *asciidocReport) getPrefix() []any {
	if r.level >= 1 {
		return []any{strings.Repeat("*", r.level)}
	}
	return []any{}
}

func (r *asciidocReport) output(d *diff.Diff) {
	if d.Empty() {
		r.print("No changes")
		return
	}

	if d.EndpointsDiff.Empty() {
		r.print("No endpoint changes, but there are some other changes")
	} else {
		r.printEndpoints(d.EndpointsDiff)
	}

	if d.ExtensionsDiff.Empty() &&
		d.SecurityDiff.Empty() &&
		d.ServersDiff.Empty() {
		return
	}

	fmt.Fprintln(r.Writer, "== Other Changes") //nolint:errcheck

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
		fmt.Fprintln(r.Writer) //nolint:errcheck
	}

	r.printValue(d.OpenAPIDiff, "Version")

	if !d.SecurityDiff.Empty() {
		r.print("Security Requirements changed")
		r.indent().printSecurityRequirements(d.SecurityDiff)
		fmt.Fprintln(r.Writer) //nolint:errcheck
	}

	if !d.ServersDiff.Empty() {
		r.print("Servers changed")
		r.indent().printServers(d.ServersDiff)
		fmt.Fprintln(r.Writer) //nolint:errcheck
	}
}

func (r *asciidocReport) printEndpoints(d *diff.EndpointsDiff) {
	r.printTitle("New Endpoints", len(d.Added))
	slices.SortFunc(d.Added, d.Added.SortFunc)
	for _, added := range d.Added {
		r.print(added.Method, added.Path)
	}
	fmt.Fprintln(r.Writer) //nolint:errcheck

	r.printTitle("Deleted Endpoints", len(d.Deleted))
	slices.SortFunc(d.Deleted, d.Deleted.SortFunc)
	for _, deleted := range d.Deleted {
		r.print(deleted.Method, deleted.Path)
	}
	fmt.Fprintln(r.Writer) //nolint:errcheck

	r.printTitle("Modified Endpoints", len(d.Modified))
	keys := d.Modified.ToEndpoints()
	slices.SortFunc(keys, keys.SortFunc)
	for _, endpoint := range keys {
		r.print(endpoint.Method, endpoint.Path)
		r.indent().printMethod(d.Modified[endpoint])
		fmt.Fprintln(r.Writer) //nolint:errcheck
	}
}

func (r *asciidocReport) printTitle(title string, count int) {
	text := ""
	if count == 0 {
		text = fmt.Sprintf("=== %s: None", title)
	} else {
		text = fmt.Sprintf("=== %s: %d", title, count)
	}
	fmt.Fprintln(r.Writer, text) //nolint:errcheck
}

func (r *asciidocReport) printServers(d *diff.ServersDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New server:", added)
	}

	slices.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted server:", deleted)
	}

	for _, server := range getKeys(d.Modified) {
		r.print("Modified server:", server)
		r.indent().printServer(d.Modified[server])
	}
}

func (r *asciidocReport) printServer(d *diff.ServerDiff) {
	if d.Empty() {
		return
	}

	r.printConditional(d.Added, "Server added")
	r.printConditional(d.Deleted, "Server deleted")

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.URLDiff, "URL")
	r.printValue(d.DescriptionDiff, "Description")
	if !d.VariablesDiff.Empty() {
		r.print("Variables changed")
		r.indent().printVariables(d.VariablesDiff)
	}
}

func (r *asciidocReport) printVariables(d *diff.VariablesDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.Added)
	for _, variable := range d.Added {
		r.print("New variable:", variable)
	}

	slices.Sort(d.Deleted)
	for _, variable := range d.Deleted {
		r.print("Deleted variable:", variable)
	}

	for _, variable := range getKeys(d.Modified) {
		r.print("Modified variable:", variable)
		r.indent().printVariable(d.Modified[variable])
	}
}

func (r *asciidocReport) printVariable(d *diff.VariableDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	if !d.EnumDiff.Empty() {
		r.printConditional(len(d.EnumDiff.Added) > 0, "New enum values:", d.EnumDiff.Added)
		r.printConditional(len(d.EnumDiff.Deleted) > 0, "Deleted enum values:", d.EnumDiff.Deleted)
	}
	r.printValue(d.DefaultDiff, "Default")
	r.printValue(d.DescriptionDiff, "Description")
}

func (r *asciidocReport) printMethod(d *diff.MethodDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printStrings(d.TagsDiff, "Tags")
	r.printValue(d.SummaryDiff, "Summary")
	r.printValue(d.DescriptionDiff, "Description")
	r.printValue(d.OperationIDDiff, "OperationID")
	r.printParams(d.ParametersDiff)

	if !d.RequestBodyDiff.Empty() {
		r.print("Request body changed")
		r.indent().printRequestBody(d.RequestBodyDiff)
	}

	if !d.ResponsesDiff.Empty() {
		r.print("Responses changed")
		r.indent().printResponses(d.ResponsesDiff)
	}

	r.printMessage(d.CallbacksDiff, "Callbacks changed")
	r.printValue(d.DeprecatedDiff, "Deprecated")

	if !d.SecurityDiff.Empty() {
		r.print("Security changed")
		r.indent().printSecurityRequirements(d.SecurityDiff)
	}

	if !d.ServersDiff.Empty() {
		r.print("Servers changed")
		r.indent().printServers(d.ServersDiff)
	}
}

func (r *asciidocReport) printParams(d *diff.ParametersDiffByLocation) {
	if d.Empty() {
		return
	}

	for _, location := range diff.ParamLocations {
		params := d.Added[location]
		slices.Sort(params)
		for _, param := range params {
			r.print("New", location, "param:", param)
		}
	}

	for _, location := range diff.ParamLocations {
		params := d.Deleted[location]
		slices.Sort(params)
		for _, param := range params {
			r.print("Deleted", location, "param:", param)
		}
	}

	for _, location := range diff.ParamLocations {
		paramDiffs := d.Modified[location]
		for _, param := range getKeys(paramDiffs) {
			r.print("Modified", location, "param:", param)
			r.indent().printParam(paramDiffs[param])
		}
	}
}

func (r *asciidocReport) printParam(d *diff.ParameterDiff) {
	r.printValue(d.NameDiff, "Name")
	r.printValue(d.InDiff, "In")

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.DescriptionDiff, "Description")
	r.printValue(d.StyleDiff, "Style")
	r.printValue(d.ExplodeDiff, "Explode")
	r.printValue(d.AllowEmptyValueDiff, "AllowEmptyValue")
	r.printValue(d.AllowReservedDiff, "AllowReserved")
	r.printValue(d.DeprecatedDiff, "Deprecated")
	r.printValue(d.RequiredDiff, "Required")

	if !d.SchemaDiff.Empty() {
		r.print("Schema changed")
		r.indent().printSchema(d.SchemaDiff)
	}

	r.printValue(d.ExampleDiff, "Example")

	if !d.ExamplesDiff.Empty() {
		r.print("Examples changed")
		r.indent().printExamples(d.ExamplesDiff)
	}

	if !d.ContentDiff.Empty() {
		r.print("Content changed")
		r.indent().printContent(d.ContentDiff)
	}
}

func (r *asciidocReport) printExamples(d *diff.ExamplesDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.Added)
	for _, example := range d.Added {
		r.print("New example:", example)
	}

	slices.Sort(d.Deleted)
	for _, example := range d.Deleted {
		r.print("Deleted example:", example)
	}

	for _, example := range getKeys(d.Modified) {
		r.print("Modified example:", example)
		r.indent().printExample(d.Modified[example])
	}
}

func (r *asciidocReport) printExample(d *diff.ExampleDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.SummaryDiff, "Summary")
	r.printValue(d.DescriptionDiff, "Description")
	r.printValue(d.ValueDiff, "Value")
	r.printValue(d.ExternalValueDiff, "ExternalValue")
}

func (r *asciidocReport) printRequiredProperties(d *diff.RequiredPropertiesDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New required property:", added)
	}

	slices.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted required property:", deleted)
	}
}

func (r *asciidocReport) printExtensions(d *diff.ExtensionsDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New extension:", added)
	}

	slices.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted extension:", deleted)
	}

	for extension, patch := range d.Modified {
		r.print("Modified extension:", extension)
		r.indent().printExtension(patch)
	}
}

func (r *asciidocReport) printExtension(d diff.JsonPatch) {
	if d.Empty() {
		return
	}

	for _, op := range d {
		r.print(op.String())
	}
}

func (r *asciidocReport) printSchema(d *diff.SchemaDiff) {
	if d.Empty() {
		return
	}

	r.printConditional(d.SchemaAdded, "Schema added")
	r.printConditional(d.SchemaDeleted, "Schema deleted")
	r.printConditional(d.CircularRefDiff, "Schema circular referecnce changed")

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	if !d.OneOfDiff.Empty() {
		r.print("Property 'OneOf' changed")
		r.indent().printSchemaListDiff(d.OneOfDiff)
	}
	if !d.AnyOfDiff.Empty() {
		r.print("Property 'AnyOf' changed")
		r.indent().printSchemaListDiff(d.AnyOfDiff)
	}
	if !d.AllOfDiff.Empty() {
		r.print("Property 'AllOf' changed")
		r.indent().printSchemaListDiff(d.AllOfDiff)
	}

	if !d.NotDiff.Empty() {
		r.print("Property 'Not' changed")
		r.indent().printSchema(d.NotDiff)
	}

	r.printStrings(d.TypeDiff, "Type")
	r.printValue(d.TitleDiff, "Title")
	r.printValue(d.FormatDiff, "Format")
	r.printValue(d.DescriptionDiff, "Description")

	if !d.EnumDiff.Empty() {
		r.printConditional(len(d.EnumDiff.Added) > 0, "New enum values:", d.EnumDiff.Added)
		r.printConditional(len(d.EnumDiff.Deleted) > 0, "Deleted enum values:", d.EnumDiff.Deleted)
	}

	r.printValue(d.DefaultDiff, "Default")
	r.printValue(d.ExampleDiff, "Example")
	r.printValue(d.AdditionalPropertiesAllowedDiff, "AdditionalProperties")
	r.printValue(d.UniqueItemsDiff, "UniqueItems")
	r.printValue(d.ExclusiveMinDiff, "ExclusiveMin")
	r.printValue(d.ExclusiveMaxDiff, "ExclusiveMax")
	r.printValue(d.NullableDiff, "Nullable")
	r.printValue(d.ReadOnlyDiff, "ReadOnly")
	r.printValue(d.WriteOnlyDiff, "WriteOnly")
	r.printValue(d.AllowEmptyValueDiff, "AllowEmptyValue")
	r.printValue(d.XMLDiff, "XML")
	r.printValue(d.DeprecatedDiff, "Deprecated")
	r.printValue(d.MinDiff, "Min")
	r.printValue(d.MaxDiff, "Max")
	r.printValue(d.MultipleOfDiff, "MultipleOf")
	r.printValue(d.MinLengthDiff, "MinLength")
	r.printValue(d.MaxLengthDiff, "MaxLength")
	r.printValue(d.PatternDiff, "Pattern")
	r.printValue(d.MinItemsDiff, "MinItems")
	r.printValue(d.MaxItemsDiff, "MaxItems")

	if !d.ItemsDiff.Empty() {
		r.print("Items changed")
		r.indent().printSchema(d.ItemsDiff)
	}

	if !d.RequiredDiff.Empty() {
		r.print("Required changed")
		r.indent().printRequiredProperties(d.RequiredDiff)
	}

	r.printValue(d.MinPropsDiff, "MinProps")
	r.printValue(d.MaxPropsDiff, "MaxProps")

	if !d.PropertiesDiff.Empty() {
		r.print("Properties changed")
		r.indent().printProperties(d.PropertiesDiff)
	}

	if !d.AdditionalPropertiesDiff.Empty() {
		r.print("AdditionalProperties changed")
		r.indent().printSchema(d.AdditionalPropertiesDiff)
	}

	r.printMessage(d.DiscriminatorDiff, "Discriminator changed")
}

func (r *asciidocReport) printSchemaListDiff(d *diff.SubschemasDiff) {
	if d.Empty() {
		return
	}

	r.printConditional(len(d.Added) > 0, "Schemas added:", d.Added)
	r.printConditional(len(d.Deleted) > 0, "Schemas deleted:", d.Deleted)

	if len(d.Modified) > 0 {
		for _, schemaDiff := range d.Modified {
			r.print("Modified schema:", schemaDiff.String())
			r.indent().printSchema(schemaDiff.Diff)
		}
	}
}

func (r *asciidocReport) printProperties(d *diff.SchemasDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.Added)
	for _, property := range d.Added {
		r.print("New property:", property)
	}

	slices.Sort(d.Deleted)
	for _, property := range d.Deleted {
		r.print("Deleted property:", property)
	}

	for _, property := range getKeys(d.Modified) {
		r.print("Modified property:", property)
		r.indent().printSchema(d.Modified[property])
	}
}

func (r *asciidocReport) printResponses(d *diff.ResponsesDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New response:", added)
	}

	slices.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted response:", deleted)
	}

	for _, response := range getKeys(d.Modified) {
		r.print("Modified response:", response)
		r.indent().printResponse(d.Modified[response])
	}
}

func (r *asciidocReport) printResponse(d *diff.ResponseDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.DescriptionDiff, "Description")

	if !d.ContentDiff.Empty() {
		r.print("Content changed")
		r.indent().printContent(d.ContentDiff)
	}

	if !d.HeadersDiff.Empty() {
		r.print("Headers changed")
		r.indent().printHeaders(d.HeadersDiff)
	}
}

func (r *asciidocReport) printRequestBody(d *diff.RequestBodyDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.DescriptionDiff, "Description")

	if !d.ContentDiff.Empty() {
		r.print("Content changed")
		r.indent().printContent(d.ContentDiff)
	}
}

func (r *asciidocReport) printContent(d *diff.ContentDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.MediaTypeAdded)
	for _, name := range d.MediaTypeAdded {
		r.print("New media type:", name)
	}

	slices.Sort(d.MediaTypeDeleted)
	for _, name := range d.MediaTypeDeleted {
		r.print("Deleted media type:", name)
	}

	for _, name := range getKeys(d.MediaTypeModified) {
		r.print("Modified media type:", name)
		r.indent().printMediaType(d.MediaTypeModified[name])
	}
}

func (r *asciidocReport) printMediaType(d *diff.MediaTypeDiff) {
	if d.Empty() {
		return
	}

	if !d.NameDiff.Empty() {
		r.printValue(d.NameDiff.NameDiff, "Name")
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	if !d.SchemaDiff.Empty() {
		r.print("Schema changed")
		r.indent().printSchema(d.SchemaDiff)
	}

	r.printValue(d.ExampleDiff, "Example")

	if !d.ExamplesDiff.Empty() {
		r.print("Examples changed")
		r.indent().printExamples(d.ExamplesDiff)
	}

	r.printMessage(d.EncodingsDiff, "Encodings changed")
}

func (r *asciidocReport) printHeaders(d *diff.HeadersDiff) {
	if d.Empty() {
		return
	}

	slices.Sort(d.Added)
	for _, added := range d.Added {
		r.print("New header:", added)
	}

	slices.Sort(d.Deleted)
	for _, deleted := range d.Deleted {
		r.print("Deleted header:", deleted)
	}

	for _, header := range getKeys(d.Modified) {
		r.print("Modified header:", header)
		r.indent().printHeader(d.Modified[header])
	}
}

func (r *asciidocReport) printHeader(d *diff.HeaderDiff) {
	if d.Empty() {
		return
	}

	if !d.ExtensionsDiff.Empty() {
		r.print("Extensions changed")
		r.indent().printExtensions(d.ExtensionsDiff)
	}

	r.printValue(d.DescriptionDiff, "Description")
	r.printValue(d.DeprecatedDiff, "Deprecated")
	r.printValue(d.RequiredDiff, "Required")

	r.printValue(d.ExampleDiff, "Example")

	if !d.ExamplesDiff.Empty() {
		r.print("Examples changed")
		r.indent().printExamples(d.ExamplesDiff)
	}

	if !d.SchemaDiff.Empty() {
		r.print("Schema changed")
		r.indent().printSchema(d.SchemaDiff)
	}

	if !d.ContentDiff.Empty() {
		r.print("Content changed")
		r.indent().printContent(d.ContentDiff)
	}
}

func (r *asciidocReport) printSecurityRequirements(d *diff.SecurityRequirementsDiff) {
	if d.Empty() {
		return
	}

	for _, added := range d.Added {
		r.print("New security requirements:", added.String())
	}

	for _, deleted := range d.Deleted {
		r.print("Deleted security requirements:", deleted.String())
	}

	for _, modified := range d.Modified {
		r.print("Modified security requirements:", modified.Base.SchemeNames())
		r.indent().printSecurityScopes(modified.Scopes)
	}
}

func (r *asciidocReport) printSecurityScopes(d diff.SecurityScopesDiff) {
	for _, scheme := range getKeys(d) {
		scopeDiff := d[scheme]
		r.printConditional(len(scopeDiff.Added) > 0, "Scheme", scheme, "Added scopes:", scopeDiff.Added)
		r.printConditional(len(scopeDiff.Deleted) > 0, "Scheme", scheme, "Deleted scopes:", scopeDiff.Deleted)
	}
}

func (r *asciidocReport) printValue(d *diff.ValueDiff, title string) {
	if d.Empty() {
		return
	}

	r.print(title, "changed from", asciidocQuote(d.From), "to", asciidocQuote(d.To))
}

func (r *asciidocReport) printStrings(d *diff.StringsDiff, title string) {
	if d.Empty() {
		return
	}

	r.print(title, "changed from", asciidocQuote(strings.Join(d.Deleted, ", ")), "to", asciidocQuote(strings.Join(d.Added, ", ")))
}

func (r *asciidocReport) printMessage(d diff.IDiff, output ...any) {
	r.printConditional(!d.Empty(), output...)
}

func (r *asciidocReport) printConditional(b bool, output ...any) {
	if b {
		r.print(output...)
	}
}

func asciidocQuote(value any) any {
	if value == nil {
		return "`null`"
	}
	if reflect.ValueOf(value).Kind() == reflect.String {
		return "`" + value.(string) + "`"
	}
	return value
}
