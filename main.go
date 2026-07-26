package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
)

type Date struct{ time.Time }

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

type SharePrice struct {
	FundCode string  `json:"fund"` // "C", "G", "L2030"
	Date     Date    `json:"date"` // date only, no clock
	Price    float64 `json:"price"`
}

// The Must prefix means "panics instead of returning an error, for known-good compile-time input"
func mustDate(s string) Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return Date{t}
}

func ptr[T any](v T) *T { return &v }

type Fund struct {
	Code       string `json:"code"`                  // "C"
	Name       string `json:"name"`                  // Common Stock Index Investment Fund
	ShortName  string `json:"short_name"`            // "C Fund"
	Kind       string `json:"kind"`                  // "core" | "lifecycle"
	TargetYear *int   `json:"target_year,omitempty"` // L funds only
	Active     bool   `json:"active"`
}

// seedFunds is placeholder data,
var seedFunds = []Fund{
	{Code: "C", Name: "Common Stock Index Investment Fund", ShortName: "C Fund", Kind: "core", Active: true},
	{Code: "S", Name: "Small cap stock Index Investment Fund", ShortName: "S Fund", Kind: "core", Active: true},
	{Code: "I", Name: "International Stock Index Investment Fund", ShortName: "I Fund", Kind: "core", Active: true},
	{Code: "G", Name: "Government Securities Investment Fund", ShortName: "G Fund", Kind: "core", Active: true},
	{Code: "F", Name: "Fixed Income Index Investment Fund", ShortName: "F Fund", Kind: "core", Active: true},
	{Code: "L2030", Name: "Lifecycle 2030 Fund", ShortName: "L 2030", Kind: "lifecycle", TargetYear: ptr(2030), Active: true},
	{Code: "L2035", Name: "Lifecycle 2035 Fund", ShortName: "L 2035", Kind: "lifecycle", TargetYear: ptr(2035), Active: true},
	{Code: "L2040", Name: "Lifecycle 2040 Fund", ShortName: "L 2040", Kind: "lifecycle", TargetYear: ptr(2040), Active: true},
	{Code: "LINCOME", Name: "Lifecycle Income Fund", ShortName: "L Income", Kind: "lifecycle", TargetYear: nil, Active: true},
}

// seedPrices is placeholder data, hand-copied from the TSP share price CSV.
// Replaced by the SQLite store - delete when the ingest job lands.
var seedPrices = []SharePrice{
	{FundCode: "C", Date: mustDate("2026-07-27"), Price: 119.3425},
	{FundCode: "G", Date: mustDate("2026-07-27"), Price: 20.0740},
	{FundCode: "L2030", Date: mustDate("2026-07-27"), Price: 62.2718},
	{FundCode: "C", Date: mustDate("2026-07-24"), Price: 119.3426},
	{FundCode: "G", Date: mustDate("2026-07-24"), Price: 20.0739},
	{FundCode: "L2030", Date: mustDate("2026-07-24"), Price: 62.2717},
	{FundCode: "C", Date: mustDate("2026-07-23"), Price: 119.2733},
	{FundCode: "G", Date: mustDate("2026-07-23"), Price: 20.0714},
	{FundCode: "L2030", Date: mustDate("2026-07-23"), Price: 62.3082},
}

func main() {
	router := gin.Default()
	router.GET("/api/v1/funds", getFunds)                // -> []Fund
	router.GET("/api/v1/prices/latest", getLatestPrices) // all funds, most recent date
	// router.GET("/api/v1/funds/:code/prices?from=&to=&limit=") // -> []SharePRice

	router.Run("localhost:8080")
}

// getFunds responds with the list of all funds as JSON
func getFunds(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, seedFunds)
}

func getLatestPrices(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, latestPrices())
}

func (d Date) Equal(other Date) bool  { return d.Time.Equal(other.Time) }
func (d Date) Compare(other Date) int { return d.Time.Compare(other.Time) }

func latestPrices() []SharePrice {
	// if we don't have any data in our seedPrices return an empty slice
	if len(seedPrices) == 0 {
		return []SharePrice{}
	}

	newest := slices.MaxFunc(seedPrices, func(a, b SharePrice) int {
		return a.Date.Compare(b.Date)
	})

	result := []SharePrice{}
	for _, p := range seedPrices {
		if p.Date.Equal(newest.Date) {
			result = append(result, p)
		}
	}
	return result
}
