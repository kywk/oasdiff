package inventory

import (
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Apply computes diff and applies changes to the inventory sheet, returning a new sheet.
func Apply(cfg *Config, spec *openapi3.T, sheet *Sheet) (*Sheet, error) {
	return ApplyWithDate(cfg, spec, sheet, time.Now().Format("2006-01-02"))
}

// ApplyWithDate applies changes using a specific date (useful for testing).
func ApplyWithDate(cfg *Config, spec *openapi3.T, sheet *Sheet, date string) (*Sheet, error) {
	diffResult, err := ComputeDiff(cfg, spec, sheet)
	if err != nil {
		return nil, err
	}

	// Clone the sheet
	result := &Sheet{
		Config:    cfg,
		Records:   make([]Record, len(sheet.Records)),
		MaxSerial: sheet.MaxSerial,
	}
	copy(result.Records, sheet.Records)

	// Apply modifications
	for _, entry := range diffResult.Modified {
		_, idx := result.FindActiveByKey(entry.Method, entry.Path)
		if idx < 0 {
			continue
		}
		result.Records[idx].ChangeType = ChangeTypeUpdate
		result.Records[idx].Date = date

		// Update changed values
		endpoints := ExtractEndpoints(spec)
		for _, ep := range endpoints {
			if ep.Method == entry.Method && ep.Path == entry.Path {
				for _, col := range cfg.Columns {
					result.Records[idx].Values[col.Source] = ExtractValue(ep, col)
				}
				break
			}
		}
	}

	// Apply deletions
	for _, entry := range diffResult.Deleted {
		_, idx := result.FindActiveByKey(entry.Method, entry.Path)
		if idx < 0 {
			continue
		}
		result.Records[idx].ChangeType = ChangeTypeDelete
		result.Records[idx].Date = date
	}

	// Apply additions
	endpoints := ExtractEndpoints(spec)
	for _, entry := range diffResult.Added {
		result.MaxSerial++
		serial := FormatSerial(cfg.Prefix, result.MaxSerial)

		values := make(map[string]string)
		for _, ep := range endpoints {
			if ep.Method == entry.Method && ep.Path == entry.Path {
				for _, col := range cfg.Columns {
					values[col.Source] = ExtractValue(ep, col)
				}
				break
			}
		}

		record := Record{
			Serial:     serial,
			Date:       date,
			ChangeType: ChangeTypeNew,
			Method:     entry.Method,
			Path:       entry.Path,
			Values:     values,
		}
		result.Records = append(result.Records, record)
	}

	return result, nil
}
