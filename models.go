package main

import (
	"encoding/json"
	"time"
)

// Date is a calendar date with no meaningful time-of-day component.
//
// TSP publishes one share price per business day, so a bare time.Time would
// carry a clock reading and a location that mean nothing here — and would leak
// into the API as "2026-07-24T00:00:00Z". Every Date is built by parsing a
// "YYYY-MM-DD" string, which yields midnight UTC and keeps comparison,
// formatting, and eventual SQLite storage predictable.
type Date struct{ time.Time }

// Equal reports whether d and other fall on the same day.
//
// Always prefer this over ==. Comparing Dates with == compares time.Time's
// internal fields, including its *Location pointer, so two Dates for the same
// instant can compare unequal if they were constructed in different zones —
// a silent wrong answer rather than an error. time.Time.Equal compares instants.
func (d Date) Equal(other Date) bool { return d.Time.Equal(other.Time) }

// Compare returns -1, 0, or +1 as d falls before, on, or after other.
// This is the three-way contract slices.SortFunc and slices.MaxFunc expect.
func (d Date) Compare(other Date) int { return d.Time.Compare(other.Time) }

// MarshalJSON renders the date as a bare "YYYY-MM-DD" string.
//
// The value receiver is deliberate. encoding/json reaches these structs through
// an interface, where they are not addressable, so a pointer receiver would not
// be found at all — json would silently fall back to the MarshalJSON promoted
// from the embedded time.Time and emit RFC 3339 instead.
func (d Date) MarshalJSON() ([]byte, error) {
	// The return value is raw JSON, not a Go string, so the surrounding quotes
	// are ours to supply.
	return []byte(`"` + d.Format("2006-01-02") + `"`), nil
}

// UnmarshalJSON parses a "YYYY-MM-DD" JSON string.
//
// The pointer receiver is required: a value receiver would parse correctly into
// a copy and then discard it, leaving the caller with a zero Date and a nil
// error. Nothing decodes SharePrice yet, but CSV ingest will.
func (d *Date) UnmarshalJSON(b []byte) error {
	// b arrives with its surrounding quotes intact, so unquote before parsing.
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// mustDate parses a "YYYY-MM-DD" literal and panics if it fails.
//
// For programmer-supplied constants only — seed data, tests, fixtures. A bad
// literal there is a typo, not a runtime condition, and panicking during package
// initialization surfaces it before the server binds a port. Never use this on
// CSV rows or request input: that data is untrusted and belongs on an error path.
func mustDate(s string) Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return Date{Time: t}
}

// SharePrice is one fund's share price on one business day.
//
// Prices are stored long/tall — one record per (fund, date) — rather than
// mirroring the wide CSV layout, because TSP retires and adds Lifecycle funds
// every few years. In this shape a new fund is new rows; in the wide shape it
// would be a schema change.
//
// FundCode and Date together form the natural key. There is deliberately no
// surrogate ID, so that re-ingesting the same night's CSV is idempotent once
// this lives in SQLite under PRIMARY KEY (fund_code, date).
type SharePrice struct {
	FundCode string  `json:"fund"`  // canonical code: "C", "G", "L2030"
	Date     Date    `json:"date"`  // business day; TSP posts no weekend rows
	Price    float64 `json:"price"` // TSP publishes 4 decimal places
}

// Fund is reference data describing a single TSP fund. It changes at most a few
// times a decade, which is why it is separate from SharePrice.
type Fund struct {
	Code      string `json:"code"`       // URL-safe and uppercase: "C", "L2030", "LINCOME"
	Name      string `json:"name"`       // official name, per tsp.gov
	ShortName string `json:"short_name"` // display label: "C Fund", "L 2030"
	Kind      string `json:"kind"`       // "core" | "lifecycle"

	// TargetYear is nil for every fund without a glide path: all five core funds,
	// and L Income, which is the terminal fund the others roll into at maturity.
	//
	// A pointer rather than a plain int so that "no target year" is distinct from
	// a target year of zero, and so the field disappears from JSON entirely.
	// Anything reading it must nil-check first — dereferencing L Income panics.
	TargetYear *int `json:"target_year,omitempty"`

	// Active is false for matured funds such as L 2025. They keep their full
	// price history and stay served by the API, but drop off the front page.
	Active bool `json:"active"`
}

// ptr returns a pointer to v.
//
// Go has no address-of for literals — &2030 does not compile, because constants
// are not addressable. Copying into a parameter produces a variable that is.
func ptr[T any](v T) *T { return &v }
