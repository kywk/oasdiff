package inventory

import (
	"time"
)

// Merge merges a patch inventory sheet into a base sheet.
// Rules:
//   - Base serial numbers are never modified
//   - New APIs in patch get new serial numbers appended to base
//   - Conflicts resolve by most recent date
//   - Deleted APIs cannot be revived
func Merge(base, patch *Sheet) (*Sheet, error) {
	return MergeWithDate(base, patch, time.Now().Format("2006-01-02"))
}

// MergeWithDate merges using a specific date (useful for testing).
func MergeWithDate(base, patch *Sheet, date string) (*Sheet, error) {
	// Clone the base sheet
	result := &Sheet{
		Config:    base.Config,
		Records:   make([]Record, len(base.Records)),
		MaxSerial: base.MaxSerial,
	}
	copy(result.Records, base.Records)

	// Build a map of base active records by identity key
	// Note: we also track deleted keys to prevent revival
	deletedKeys := make(map[string]bool)
	for _, r := range result.Records {
		if r.ChangeType == ChangeTypeDelete {
			deletedKeys[r.IdentityKey()] = true
		}
	}

	// Process each patch record
	for _, patchRecord := range patch.Records {
		if patchRecord.ChangeType == ChangeTypeDelete {
			// If patch says delete, and base has it active, mark as deleted
			_, idx := result.FindActiveByKey(patchRecord.Method, patchRecord.Path)
			if idx >= 0 {
				result.Records[idx].ChangeType = ChangeTypeDelete
				result.Records[idx].Date = resolveDate(result.Records[idx].Date, patchRecord.Date, date)
			}
			continue
		}

		key := patchRecord.IdentityKey()

		// Check if this key was deleted in base — cannot revive
		if deletedKeys[key] {
			// Treat as a new API with a new serial number
			result.MaxSerial++
			serial := FormatSerial(base.Config.Prefix, result.MaxSerial)
			newRecord := Record{
				Serial:     serial,
				Date:       resolveDate("", patchRecord.Date, date),
				ChangeType: ChangeTypeNew,
				Method:     patchRecord.Method,
				Path:       patchRecord.Path,
				Values:     copyValues(patchRecord.Values),
			}
			result.Records = append(result.Records, newRecord)
			continue
		}

		// Check if it exists in base (active)
		existing, idx := result.FindActiveByKey(patchRecord.Method, patchRecord.Path)
		if existing != nil {
			// Resolve conflict by most recent date
			if patchRecord.Date > existing.Date {
				result.Records[idx].ChangeType = ChangeTypeUpdate
				result.Records[idx].Date = patchRecord.Date
				result.Records[idx].Values = copyValues(patchRecord.Values)
			}
		} else {
			// New API from patch — add with new serial
			result.MaxSerial++
			serial := FormatSerial(base.Config.Prefix, result.MaxSerial)
			newRecord := Record{
				Serial:     serial,
				Date:       resolveDate("", patchRecord.Date, date),
				ChangeType: ChangeTypeNew,
				Method:     patchRecord.Method,
				Path:       patchRecord.Path,
				Values:     copyValues(patchRecord.Values),
			}
			result.Records = append(result.Records, newRecord)
		}
	}

	return result, nil
}

// resolveDate picks the most recent date among the provided values.
// Falls back to defaultDate if no valid date is found.
func resolveDate(dates ...string) string {
	best := ""
	for _, d := range dates {
		if d > best {
			best = d
		}
	}
	if best == "" {
		return time.Now().Format("2006-01-02")
	}
	return best
}

// copyValues creates a shallow copy of a values map.
func copyValues(src map[string]string) map[string]string {
	if src == nil {
		return make(map[string]string)
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
