package main

import "testing"

// TestFundPricesSortedNewestFirst pins the ordering fundPrices promises.
//
// It checks the invariant — each row is at least as new as the one after it —
// rather than comparing against a hardcoded list of dates. An expected-values
// test would have to be edited every time the seed data changes; this one keeps
// working, and it is the actual property callers depend on.
func TestFundPricesSortedNewestFirst(t *testing.T) {
	got := fundPrices("C")

	if len(got) < 2 {
		t.Fatalf("got %d rows, need at least 2 to check ordering", len(got))
	}

	for i := 1; i < len(got); i++ {
		if got[i-1].Date.Compare(got[i].Date) < 0 {
			t.Errorf("row %d (%s) is older than row %d (%s), want newest first",
				i-1, got[i-1].Date.Format("2006-01-02"),
				i, got[i].Date.Format("2006-01-02"))
		}
	}
}
