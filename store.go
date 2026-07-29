package main

import (
	"cmp"
	"slices"
)

// The seed slices stand in for SQLite tables while the storage layer is developed.
// Handlers access them through the query functions below.

// seedFunds is placeholder fund metadata, hand-copied from tsp.gov.
// TODO: decide at the SQLite step whether this list should be derived from the
// CSV header instead, so a newly added fund cannot have prices but no metadata.
var seedFunds = []Fund{
	{Code: "C", Name: "Common Stock Index Investment Fund", ShortName: "C Fund", Kind: "core", Active: true},
	{Code: "S", Name: "Small cap stock Index Investment Fund", ShortName: "S Fund", Kind: "core", Active: true},
	{Code: "I", Name: "International Stock Index Investment Fund", ShortName: "I Fund", Kind: "core", Active: true},
	{Code: "G", Name: "Government Securities Investment Fund", ShortName: "G Fund", Kind: "core", Active: true},
	{Code: "F", Name: "Fixed Income Index Investment Fund", ShortName: "F Fund", Kind: "core", Active: true},
	{Code: "L2030", Name: "Lifecycle 2030 Fund", ShortName: "L 2030", Kind: "lifecycle", TargetYear: ptr(2030), Active: true},
	{Code: "L2035", Name: "Lifecycle 2035 Fund", ShortName: "L 2035", Kind: "lifecycle", TargetYear: ptr(2035), Active: true},
	{Code: "L2040", Name: "Lifecycle 2040 Fund", ShortName: "L 2040", Kind: "lifecycle", TargetYear: ptr(2040), Active: true},
	{Code: "L2045", Name: "Lifecycle 2045 Fund", ShortName: "L 2045", Kind: "lifecycle", TargetYear: ptr(2045), Active: true},
	{Code: "L2050", Name: "Lifecycle 2050 Fund", ShortName: "L 2050", Kind: "lifecycle", TargetYear: ptr(2050), Active: true},
	{Code: "L2055", Name: "Lifecycle 2055 Fund", ShortName: "L 2055", Kind: "lifecycle", TargetYear: ptr(2055), Active: true},
	{Code: "L2060", Name: "Lifecycle 2060 Fund", ShortName: "L 2060", Kind: "lifecycle", TargetYear: ptr(2060), Active: true},
	{Code: "L2065", Name: "Lifecycle 2065 Fund", ShortName: "L 2065", Kind: "lifecycle", TargetYear: ptr(2065), Active: true},
	{Code: "L2070", Name: "Lifecycle 2070 Fund", ShortName: "L 2070", Kind: "lifecycle", TargetYear: ptr(2070), Active: true},
	{Code: "L2075", Name: "Lifecycle 2075 Fund", ShortName: "L 2075", Kind: "lifecycle", TargetYear: ptr(2075), Active: true},
	{Code: "LINCOME", Name: "Lifecycle Income Fund", ShortName: "L Income", Kind: "lifecycle", TargetYear: nil, Active: true},
}

// seedPrices contains two complete daily snapshots copied from the CSV fixture.
// The oldest snapshot is stored first so queries cannot rely on insertion order
// to return the newest price first. Delete it once the nightly ingest job is
// implemented.
var seedPrices = []SharePrice{
	{FundCode: "C", Date: mustDate("2026-07-23"), Price: 119.2733},
	{FundCode: "S", Date: mustDate("2026-07-23"), Price: 114.3141},
	{FundCode: "I", Date: mustDate("2026-07-23"), Price: 63.1339},
	{FundCode: "G", Date: mustDate("2026-07-23"), Price: 20.0714},
	{FundCode: "F", Date: mustDate("2026-07-23"), Price: 20.7658},
	{FundCode: "L2030", Date: mustDate("2026-07-23"), Price: 62.3082},
	{FundCode: "L2035", Date: mustDate("2026-07-23"), Price: 19.1213},
	{FundCode: "L2040", Date: mustDate("2026-07-23"), Price: 73.6486},
	{FundCode: "L2045", Date: mustDate("2026-07-23"), Price: 20.4583},
	{FundCode: "L2050", Date: mustDate("2026-07-23"), Price: 45.5054},
	{FundCode: "L2055", Date: mustDate("2026-07-23"), Price: 23.8351},
	{FundCode: "L2060", Date: mustDate("2026-07-23"), Price: 23.8319},
	{FundCode: "L2065", Date: mustDate("2026-07-23"), Price: 23.8287},
	{FundCode: "L2070", Date: mustDate("2026-07-23"), Price: 14.1226},
	{FundCode: "L2075", Date: mustDate("2026-07-23"), Price: 12.3361},
	{FundCode: "LINCOME", Date: mustDate("2026-07-23"), Price: 30.6476},
	{FundCode: "C", Date: mustDate("2026-07-24"), Price: 119.3426},
	{FundCode: "S", Date: mustDate("2026-07-24"), Price: 113.9880},
	{FundCode: "I", Date: mustDate("2026-07-24"), Price: 62.9110},
	{FundCode: "G", Date: mustDate("2026-07-24"), Price: 20.0739},
	{FundCode: "F", Date: mustDate("2026-07-24"), Price: 20.7899},
	{FundCode: "L2030", Date: mustDate("2026-07-24"), Price: 62.2717},
	{FundCode: "L2035", Date: mustDate("2026-07-24"), Price: 19.1069},
	{FundCode: "L2040", Date: mustDate("2026-07-24"), Price: 73.5874},
	{FundCode: "L2045", Date: mustDate("2026-07-24"), Price: 20.4399},
	{FundCode: "L2050", Date: mustDate("2026-07-24"), Price: 45.4611},
	{FundCode: "L2055", Date: mustDate("2026-07-24"), Price: 23.8043},
	{FundCode: "L2060", Date: mustDate("2026-07-24"), Price: 23.8012},
	{FundCode: "L2065", Date: mustDate("2026-07-24"), Price: 23.7980},
	{FundCode: "L2070", Date: mustDate("2026-07-24"), Price: 14.1044},
	{FundCode: "L2075", Date: mustDate("2026-07-24"), Price: 12.3203},
	{FundCode: "LINCOME", Date: mustDate("2026-07-24"), Price: 30.6409},
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
