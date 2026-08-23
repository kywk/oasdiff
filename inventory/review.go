package inventory

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Review summarizes the OpenAPI spec information based on config columns.
// It outputs statistics to the writer including custom extension attribute counts.
func Review(cfg *Config, spec *openapi3.T, w io.Writer) error {
	endpoints := ExtractEndpoints(spec)
	if len(endpoints) == 0 {
		fmt.Fprintf(w, "No API endpoints found in spec.\n")
		return nil
	}

	fmt.Fprintf(w, "=== API Inventory Review ===\n\n")
	fmt.Fprintf(w, "Total APIs: %d\n\n", len(endpoints))

	// Group by tags
	tagCounts := make(map[string]int)
	for _, ep := range endpoints {
		if len(ep.Op.Tags) == 0 {
			tagCounts["(untagged)"]++
		} else {
			for _, tag := range ep.Op.Tags {
				tagCounts[tag]++
			}
		}
	}

	fmt.Fprintf(w, "--- By Tag ---\n")
	tags := sortedKeys(tagCounts)
	for _, tag := range tags {
		fmt.Fprintf(w, "  %s: %d APIs\n", tag, tagCounts[tag])
	}
	fmt.Fprintf(w, "\n")

	// Group by method
	methodCounts := make(map[string]int)
	for _, ep := range endpoints {
		methodCounts[ep.Method]++
	}
	fmt.Fprintf(w, "--- By Method ---\n")
	methods := sortedKeys(methodCounts)
	for _, method := range methods {
		fmt.Fprintf(w, "  %s: %d APIs\n", method, methodCounts[method])
	}
	fmt.Fprintf(w, "\n")

	// Custom extension statistics
	for _, col := range cfg.Columns {
		if !strings.HasPrefix(col.Source, "x-") && !strings.HasPrefix(col.Source, "X-") {
			continue
		}

		fmt.Fprintf(w, "--- %s (%s) ---\n", col.Header, col.Source)
		valueCounts := make(map[string][]string)
		for _, ep := range endpoints {
			val := ExtractValue(ep, col)
			if val == "" {
				val = "(empty)"
			}
			valueCounts[val] = append(valueCounts[val], fmt.Sprintf("%s %s", ep.Method, ep.Path))
		}

		values := sortedKeys(valueCounts)
		for _, val := range values {
			apis := valueCounts[val]
			fmt.Fprintf(w, "  %s=%s: %d APIs\n", col.Source, val, len(apis))
			for _, api := range apis {
				fmt.Fprintf(w, "    - %s\n", api)
			}
		}
		fmt.Fprintf(w, "\n")
	}

	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
