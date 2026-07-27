package main

import (
	"encoding/json"
	"time"
)

// dateLayout is the wire and storage format used for calendar dates.
const dateLayout = "2006-01-02"

// Date is a calendar date with no meaningful time-of-day component.
// Parsed values use midnight UTC and are encoded as "YYYY-MM-DD".
type Date struct{ time.Time }

// Equal compares instants without comparing time.Time's internal location data.
func (d Date) Equal(other Date) bool { return d.Time.Equal(other.Time) }

// Compare returns -1, 0, or +1 when d is before, equal to, or after other.
func (d Date) Compare(other Date) int { return d.Time.Compare(other.Time) }

// MarshalJSON renders the date as a bare "YYYY-MM-DD" string.
// A value receiver ensures encoding/json uses this method for Date values.
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(dateLayout) + `"`), nil
}

// UnmarshalJSON parses a "YYYY-MM-DD" JSON string.
func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := parseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// parseDate parses external input in "YYYY-MM-DD" format.
func parseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, err
	}
	return Date{Time: t}, nil
}

// mustDate parses trusted literals used by seeds and tests.
// It should not be used for request or imported data.
func mustDate(s string) Date {
	d, err := parseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

// SharePrice is one fund's share price on one business day.
//
// Data is stored as one record per fund and date so funds can be added without
// changing the schema. FundCode and Date form the natural key.
type SharePrice struct {
	FundCode string  `json:"fund"`  // canonical code: "C", "G", "L2030"
	Date     Date    `json:"date"`  // business day; TSP posts no weekend rows
	Price    float64 `json:"price"` // TSP publishes 4 decimal places
}

// Fund is reference data describing a TSP fund.
type Fund struct {
	Code      string `json:"code"`       // URL-safe and uppercase: "C", "L2030", "LINCOME"
	Name      string `json:"name"`       // official name, per tsp.gov
	ShortName string `json:"short_name"` // display label: "C Fund", "L 2030"
	Kind      string `json:"kind"`       // "core" | "lifecycle"

	// TargetYear is nil for core funds and L Income.
	TargetYear *int `json:"target_year,omitempty"`

	// Inactive funds retain their historical prices.
	Active bool `json:"active"`
}

// ptr makes literal pointer fields concise in seed data.
func ptr[T any](v T) *T { return &v }
