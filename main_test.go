package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// These tests drive the router in-process: httptest.NewRecorder stands in for a
// real connection, so nothing binds a port and nothing touches the network.
// Requests go through newMux rather than calling handlers directly, which means
// a broken route registration fails here too.
//
// Both tests currently assume seedPrices is non-empty. That coupling goes away
// when the store becomes a value the tests can construct themselves.

// TestLatestPricesReturnsOK is the smoke test: the route is registered for GET
// and the handler answers without blowing up.
func TestLatestPricesReturnsOK(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/prices/latest", nil)

	newMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestLatestPricesShareOneDate protects the invariant the endpoint actually
// promises: it returns one snapshot in time, so every row must carry the same
// date. A filter bug that leaked older rows would still return 200 and still
// decode cleanly — this is the test that would catch it.
func TestLatestPricesShareOneDate(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/prices/latest", nil)

	newMux().ServeHTTP(rec, req)

	// Fatal rather than Error for both checks below: with no decodable body
	// there is nothing left to assert against.
	var prices []SharePrice
	if err := json.NewDecoder(rec.Body).Decode(&prices); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if len(prices) == 0 {
		t.Fatal("got 0 prices, want at least 1")
	}

	// The first row sets the expectation; every other row must match it. Error
	// rather than Fatal here so a failure names every offending fund at once
	// instead of stopping at the first.
	want := prices[0].Date
	for _, p := range prices[1:] {
		if !p.Date.Equal(want) {
			t.Errorf("fund %s has date %s, want %s",
				p.FundCode, p.Date.Format("2006-01-02"), want.Format("2006-01-02"))
		}
	}
}
