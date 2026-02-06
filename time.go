// ABOUTME: Time formatting and parsing helpers for consistent storage.
// ABOUTME: Uses RFC3339Nano as primary format with RFC3339 fallback for parsing.
package mdstore

import (
	"fmt"
	"time"
)

// FormatTime formats a time in RFC3339Nano for consistent storage.
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

// ParseTime parses a time string, trying RFC3339Nano first, then RFC3339.
// Returns error if neither format matches.
func ParseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("mdstore: unable to parse time %q: expected RFC3339 or RFC3339Nano format", s)
}
