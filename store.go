package main

import (
	"cmp"
	"slices"
)

// The seed slices stand in for SQLite tables while the storage layer is developed.
// Handlers access them through the query functions below.

// seedFunds is placeholder fund metadata, hand-copied from tsp.gov.
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

// seedPrices is intentionally unordered and spans a weekend. This ensures query
// behavior does not depend on insertion order or consecutive calendar dates.
// Delete it once the nightly ingest job is implemented.
var seedPrices = []SharePrice{
	{FundCode: "C", Date: mustDate("2026-07-23"), Price: 119.2733},
	{FundCode: "G", Date: mustDate("2026-07-23"), Price: 20.0714},
	{FundCode: "L2030", Date: mustDate("2026-07-23"), Price: 62.3082},
	{FundCode: "C", Date: mustDate("2026-07-27"), Price: 119.3425},
	{FundCode: "G", Date: mustDate("2026-07-27"), Price: 20.0740},
	{FundCode: "L2030", Date: mustDate("2026-07-27"), Price: 62.2718},
	{FundCode: "C", Date: mustDate("2026-07-24"), Price: 119.3426},
	{FundCode: "G", Date: mustDate("2026-07-24"), Price: 20.0739},
	{FundCode: "L2030", Date: mustDate("2026-07-24"), Price: 62.2717},
}

// priceQuery narrows a request for one fund's price history. From and To are
// inclusive bounds; nil means unbounded on that side. Limit has already been
// defaulted and capped by the caller.
type priceQuery struct {
	Code  string
	From  *Date
	To    *Date
	Limit int
}

// allFunds returns metadata for every fund.
// The returned slice shares its backing array with the seed data.
func allFunds() []Fund { return seedFunds }

// fundByCode returns the matching fund and whether it exists.
func fundByCode(code string) (Fund, bool) {
	for _, f := range seedFunds {
		if f.Code == code {
			return f, true
		}
	}
	return Fund{}, false
}

// latestPrices returns every fund's price for the most recent date on record,
// sorted by fund code, or an empty slice when there is no data at all.
func latestPrices() []SharePrice {
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

	slices.SortFunc(result, func(a, b SharePrice) int {
		return cmp.Compare(a.FundCode, b.FundCode)
	})

	return result
}

// fundPrices returns one fund's prices matching q, newest date first. The code
// must already be normalized — callers uppercase at the HTTP boundary.
func fundPrices(q priceQuery) []SharePrice {
	result := []SharePrice{}
	for _, p := range seedPrices {
		if p.FundCode != q.Code {
			continue
		}
		if q.From != nil && p.Date.Compare(*q.From) < 0 {
			continue
		}
		if q.To != nil && p.Date.Compare(*q.To) > 0 {
			continue
		}

		result = append(result, p)
	}
	slices.SortFunc(result, func(a, b SharePrice) int {
		return b.Date.Compare(a.Date)
	})
	// Applied after the sort, so a limit returns the most recent
	if q.Limit > 0 && q.Limit < len(result) {
		result = result[:q.Limit]
	}
	return result
}
