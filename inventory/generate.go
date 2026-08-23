package inventory

import (
	"fmt"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// GenerateOptions holds options for the generate command.
type GenerateOptions struct {
	Config     *Config
	Spec       *openapi3.T
	OutputPath string
	Format     string // "csv" or "excel"; overrides config if set
	Date       string // override date for testing; defaults to today
}

// Generate produces an inventory sheet from an OpenAPI spec.
func Generate(opts GenerateOptions) (*Sheet, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if opts.Spec == nil {
		return nil, fmt.Errorf("OpenAPI spec is required")
	}

	date := opts.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	format := opts.Format
	if format == "" {
		format = opts.Config.Output
	}

	// Warn about complex columns in CSV mode
	if format == OutputCSV && opts.Config.HasComplexColumns() {
		fmt.Printf("Warning: requestBody and responses columns are not supported in CSV format and will be skipped\n")
	}

	// Extract all endpoints from spec
	endpoints := ExtractEndpoints(opts.Spec)
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no API endpoints found in spec")
	}

	sheet := NewSheet(opts.Config)

	for i, ep := range endpoints {
		serial := FormatSerial(opts.Config.Prefix, i+1)

		values := make(map[string]string)
		for _, col := range opts.Config.Columns {
			values[col.Source] = ExtractValue(ep, col)
		}

		record := Record{
			Serial:     serial,
			Date:       date,
			ChangeType: ChangeTypeNew,
			Method:     ep.Method,
			Path:       ep.Path,
			Values:     values,
		}

		sheet.Records = append(sheet.Records, record)
	}

	sheet.MaxSerial = len(endpoints)

	// Write output if path is specified
	if opts.OutputPath != "" {
		if err := WriteSheet(sheet, opts.OutputPath, format); err != nil {
			return nil, fmt.Errorf("failed to write inventory: %w", err)
		}
	}

	return sheet, nil
}
