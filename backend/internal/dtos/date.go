/*
Package dtos - Custom Date Type for JSON Parsing

==============================================================================
FILE: internal/dtos/date.go
==============================================================================

DESCRIPTION:
    Provides a custom Date type that can parse both "YYYY-MM-DD" and RFC3339
    formats from JSON. This is needed because frontend sends dates as
    "YYYY-MM-DD" but Go's default time.Time expects RFC3339.

==============================================================================
*/
package dtos

import (
	"strings"
	"time"
)

// Date is a custom time type that can parse "YYYY-MM-DD" format
type Date struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for Date
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		d.Time = time.Time{}
		return nil
	}

	// Try parsing as date-only format first (YYYY-MM-DD)
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try RFC3339 format
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try RFC3339 without timezone
	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		d.Time = t
		return nil
	}

	return err
}

// MarshalJSON implements json.Marshaler for Date
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte("\"" + d.Time.Format("2006-01-02") + "\""), nil
}

// ToTime converts Date to time.Time
func (d Date) ToTime() time.Time {
	return d.Time
}

// DatePtr is a pointer version for optional dates
type DatePtr struct {
	*time.Time
}

// UnmarshalJSON implements json.Unmarshaler for DatePtr
func (d *DatePtr) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		d.Time = nil
		return nil
	}

	// Try parsing as date-only format first (YYYY-MM-DD)
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		d.Time = &t
		return nil
	}

	// Try RFC3339 format
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		d.Time = &t
		return nil
	}

	// Try RFC3339 without timezone
	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		d.Time = &t
		return nil
	}

	return err
}

// MarshalJSON implements json.Marshaler for DatePtr
func (d DatePtr) MarshalJSON() ([]byte, error) {
	if d.Time == nil {
		return []byte("null"), nil
	}
	return []byte("\"" + d.Time.Format("2006-01-02") + "\""), nil
}
