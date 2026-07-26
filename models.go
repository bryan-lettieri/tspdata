package main

import (
	"encoding/json"
	"time"
)

type Date struct{ time.Time }

func (d Date) Equal(other Date) bool  { return d.Time.Equal(other.Time) }
func (d Date) Compare(other Date) int { return d.Time.Compare(other.Time) }

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format("2006-01-02") + `"`), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
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

func mustDate(s string) Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return Date{t}
}

type SharePrice struct {
	FundCode string  `json:"fund"` // "C", "G", "L2030"
	Date     Date    `json:"date"` // date only, no clock
	Price    float64 `json:"price"`
}

type Fund struct {
	Code       string `json:"code"`                  // "C"
	Name       string `json:"name"`                  // Common Stock Index Investment Fund
	ShortName  string `json:"short_name"`            // "C Fund"
	Kind       string `json:"kind"`                  // "core" | "lifecycle"
	TargetYear *int   `json:"target_year,omitempty"` // L funds only
	Active     bool   `json:"active"`
}

func ptr[T any](v T) *T { return &v }
