package inventory

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatSerial generates a serial string from a prefix and number.
// Numbers below 1000 are zero-padded to 3 digits; above that they expand naturally.
func FormatSerial(prefix string, num int) string {
	if num < 1000 {
		return fmt.Sprintf("%s-%03d", prefix, num)
	}
	return fmt.Sprintf("%s-%d", prefix, num)
}

// ParseSerial extracts the numeric part from a serial string.
// Returns the number and an error if parsing fails.
func ParseSerial(serial string) (int, error) {
	idx := strings.LastIndex(serial, "-")
	if idx < 0 {
		return 0, fmt.Errorf("invalid serial format: %q (no dash found)", serial)
	}
	numStr := serial[idx+1:]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid serial number in %q: %w", serial, err)
	}
	return num, nil
}

// NextSerial returns the next serial number to use, given the current max.
func NextSerial(prefix string, currentMax int) string {
	return FormatSerial(prefix, currentMax+1)
}
