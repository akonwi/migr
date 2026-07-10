package ffi

import (
	"testing"
	"time"
)

// normalize sits between database/sql's driver-specific Go values and the
// Ard side. These tests lock in the conversion contract so that:
//
//   - Ard receives its native `Int` / `Str` / `Float64` for numeric and
//     text columns regardless of driver.
//   - SQL NULLs survive as nil so `decode::nullable` can detect them.
//   - Postgres / MySQL TIMESTAMP columns arrive as RFC3339Nano strings
//     (SQLite already stores dates as TEXT, so this makes all three
//     dialects behave symmetrically from Ard's point of view).
func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want any
	}{
		{"bytes to string", []byte("hello"), "hello"},
		{"int64 to int", int64(42), int(42)},
		{"int32 to int", int32(7), int(7)},
		{"float32 to float64", float32(1.5), float64(1.5)},
		{"string passthrough", "already a string", "already a string"},
		{"int passthrough", int(3), int(3)},
		{"float64 passthrough", float64(2.5), float64(2.5)},
		{"bool passthrough", true, true},
		{"nil preserved for null detection", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalize(tt.in)
			if got != tt.want {
				t.Fatalf("normalize(%v) = %v (%T), want %v (%T)",
					tt.in, got, got, tt.want, tt.want)
			}
		})
	}
}

// time.Time is the Gap 2 case from the FFI normalize follow-up: without
// this conversion, Postgres/MySQL TIMESTAMP columns arrive as an opaque
// time.Time struct that the Ard-side `decode` module can't unwrap. Format
// at the boundary so downstream code just sees a string.
func TestNormalize_TimeIsFormattedRFC3339Nano(t *testing.T) {
	when := time.Date(2025, 3, 14, 15, 9, 26, 535_000_000, time.UTC)

	got := normalize(when)

	want := "2025-03-14T15:09:26.535Z"
	if got != want {
		t.Fatalf("normalize(time.Time) = %q, want %q", got, want)
	}

	// And with sub-millisecond precision, so we know we're using Nano
	// rather than the second-precision RFC3339 constant.
	precise := time.Date(2025, 3, 14, 15, 9, 26, 123_456_789, time.UTC)
	if got := normalize(precise); got != "2025-03-14T15:09:26.123456789Z" {
		t.Fatalf("nanosecond precision lost: got %q", got)
	}
}
