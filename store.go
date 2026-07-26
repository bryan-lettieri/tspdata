package main

import "slices"

// seedFunds is placeholder data
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
	return result
}

func allFunds() []Fund { return seedFunds }
