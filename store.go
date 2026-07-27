package main

import "slices"

// Everything in this file is scaffolding for a database that does not exist yet.
//
// The seed slices below stand in for SQLite tables while the schema settles.
// Handlers reach the data only through the functions at the bottom of the file,
// never through the slices directly, so swapping in a real store should not
// require touching main.go.

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

// seedPrices is placeholder price data, hand-copied from the TSP share price CSV.
//
// The rows deliberately span a weekend — Friday 07-24 to Monday 07-27 — so that
// any "latest price" logic written against them cannot quietly assume yesterday
// exists. TSP posts on business days only.
//
// They are also deliberately out of date order. Do not tidy them: sorting the
// seed data would let an ordering bug in the query functions pass unnoticed,
// because the output would look correct without the sort ever running.
//
// Delete once the nightly ingest job lands.
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

// latestPrices returns every fund's price for the most recent date on record,
// or an empty slice when there is no data at all.
//
// This makes two full passes over the data: one to find the newest date, one to
// collect the rows carrying it. That is O(n) per request — around 0.6ms against
// a full 23-year dataset, which is fine for this endpoint but will not survive
// the date-range queries still to come. In SQLite it becomes a single
// WHERE date = (SELECT MAX(date) FROM prices) against an index.
//
// TODO: unlike fundPrices, the rows here come back in seedPrices order, which is
// incidental rather than promised. Sort by FundCode so the response has a stable
// documented order — SQLite will not preserve insertion order either.
func latestPrices() []SharePrice {
	if len(seedPrices) == 0 {
		// Two reasons this guard is not optional: slices.MaxFunc panics on an
		// empty slice, and a nil slice marshals to JSON null rather than [].
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

// allFunds returns metadata for every fund.
//
// The returned slice shares its backing array with seedFunds, so a caller that
// modifies an element modifies the seed data itself. That is safe today because
// the only caller serializes it and discards it; wrap this in slices.Clone the
// moment that stops being true.
func allFunds() []Fund { return seedFunds }

// fundPrices returns every recorded price for one fund, newest date first. The
// code must already be normalized — callers uppercase at the HTTP boundary.
//
// The comparator's operands are reversed on purpose — b before a — which is what
// makes the sort descending. Newest-first is the ordering that makes a limit
// useful: "the most recent 100" is nearly always what a caller wants, where the
// oldest 100 essentially never is.
//
// TODO: from/to/limit are still unimplemented. Apply them after the sort, so
// that limiting takes the most recent rows rather than an arbitrary slice.
func fundPrices(code string) []SharePrice {
	result := []SharePrice{}
	for _, p := range seedPrices {
		if p.FundCode == code {
			result = append(result, p)
		}
	}
	slices.SortFunc(result, func(a, b SharePrice) int {
		return b.Date.Compare(a.Date)
	})
	return result
}

// fundByCode returns the fund with the given code and reports whether it exists.
// The comma-ok shape mirrors a map lookup: on a miss it returns the zero Fund,
// so callers must check the bool before trusting the value.
func fundByCode(code string) (Fund, bool) {
	for _, f := range seedFunds {
		if f.Code == code {
			return f, true
		}
	}
	return Fund{}, false
}
