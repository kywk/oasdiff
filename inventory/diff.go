package inventory

import (
	"fmt"
	"io"

	"github.com/getkin/kin-openapi/openapi3"
)

// DiffResult represents the differences between an OpenAPI spec and an inventory sheet.
type DiffResult struct {
	Added    []DiffEntry // APIs in spec but not in inventory
	Modified []DiffEntry // APIs in both but with field changes
	Deleted  []DiffEntry // APIs in inventory but not in spec
}

// DiffEntry represents a single diff item.
type DiffEntry struct {
	Method  string
	Path    string
	Serial  string            // existing serial (empty for new APIs)
	Changes map[string]Change // field -> change details (for modified)
}

// Change represents a field value change.
type Change struct {
	Field    string
	OldValue string
	NewValue string
}

// ComputeDiff computes differences between an OpenAPI spec and an existing inventory.
func ComputeDiff(cfg *Config, spec *openapi3.T, sheet *Sheet) (*DiffResult, error) {
	endpoints := ExtractEndpoints(spec)

	result := &DiffResult{}

	// Build a map of spec endpoints for quick lookup
	specMap := make(map[string]APIEndpoint)
	for _, ep := range endpoints {
		key := ep.Method + " " + ep.Path
		specMap[key] = ep
	}

	// Build a map of active inventory records
	activeMap := make(map[string]*Record)
	for i := range sheet.Records {
		r := &sheet.Records[i]
		if r.ChangeType != ChangeTypeDelete {
			activeMap[r.IdentityKey()] = r
		}
	}

	// Find added and modified
	for _, ep := range endpoints {
		key := ep.Method + " " + ep.Path
		existing, found := activeMap[key]

		if !found {
			// New API
			result.Added = append(result.Added, DiffEntry{
				Method: ep.Method,
				Path:   ep.Path,
			})
			continue
		}

		// Check for modifications in configured columns
		changes := make(map[string]Change)
		for _, col := range cfg.Columns {
			newVal := ExtractValue(ep, col)
			oldVal := existing.Values[col.Source]
			if newVal != oldVal {
				changes[col.Source] = Change{
					Field:    col.Header,
					OldValue: oldVal,
					NewValue: newVal,
				}
			}
		}
		if len(changes) > 0 {
			result.Modified = append(result.Modified, DiffEntry{
				Method:  ep.Method,
				Path:    ep.Path,
				Serial:  existing.Serial,
				Changes: changes,
			})
		}
	}

	// Find deleted (in inventory but not in spec)
	for key, record := range activeMap {
		if _, found := specMap[key]; !found {
			result.Deleted = append(result.Deleted, DiffEntry{
				Method: record.Method,
				Path:   record.Path,
				Serial: record.Serial,
			})
		}
	}

	return result, nil
}

// Diff computes and prints the differences to the writer.
func Diff(cfg *Config, spec *openapi3.T, sheet *Sheet, w io.Writer) error {
	result, err := ComputeDiff(cfg, spec, sheet)
	if err != nil {
		return err
	}

	if len(result.Added) == 0 && len(result.Modified) == 0 && len(result.Deleted) == 0 {
		fmt.Fprintf(w, "No differences found.\n")
		return nil
	}

	fmt.Fprintf(w, "=== Inventory Diff ===\n\n")

	if len(result.Added) > 0 {
		fmt.Fprintf(w, "--- Added (%d) ---\n", len(result.Added))
		for _, entry := range result.Added {
			fmt.Fprintf(w, "  + %s %s\n", entry.Method, entry.Path)
		}
		fmt.Fprintf(w, "\n")
	}

	if len(result.Modified) > 0 {
		fmt.Fprintf(w, "--- Modified (%d) ---\n", len(result.Modified))
		for _, entry := range result.Modified {
			fmt.Fprintf(w, "  ~ %s %s %s\n", entry.Serial, entry.Method, entry.Path)
			for _, change := range entry.Changes {
				fmt.Fprintf(w, "      %s: %q → %q\n", change.Field, change.OldValue, change.NewValue)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	if len(result.Deleted) > 0 {
		fmt.Fprintf(w, "--- Deleted (%d) ---\n", len(result.Deleted))
		for _, entry := range result.Deleted {
			fmt.Fprintf(w, "  - %s %s %s\n", entry.Serial, entry.Method, entry.Path)
		}
		fmt.Fprintf(w, "\n")
	}

	return nil
}
