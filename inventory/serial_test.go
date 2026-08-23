package inventory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatSerial(t *testing.T) {
	tests := []struct {
		prefix   string
		num      int
		expected string
	}{
		{"API", 1, "API-001"},
		{"API", 9, "API-009"},
		{"API", 10, "API-010"},
		{"API", 99, "API-099"},
		{"API", 100, "API-100"},
		{"API", 999, "API-999"},
		{"API", 1000, "API-1000"},
		{"API", 1234, "API-1234"},
		{"SVC", 1, "SVC-001"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatSerial(tt.prefix, tt.num)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestParseSerial(t *testing.T) {
	tests := []struct {
		serial   string
		expected int
		wantErr  bool
	}{
		{"API-001", 1, false},
		{"API-099", 99, false},
		{"API-100", 100, false},
		{"API-1000", 1000, false},
		{"SVC-042", 42, false},
		{"INVALID", 0, true},
		{"API-abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.serial, func(t *testing.T) {
			result, err := ParseSerial(tt.serial)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNextSerial(t *testing.T) {
	require.Equal(t, "API-001", NextSerial("API", 0))
	require.Equal(t, "API-002", NextSerial("API", 1))
	require.Equal(t, "API-1000", NextSerial("API", 999))
}
