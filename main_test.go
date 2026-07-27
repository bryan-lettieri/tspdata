package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Handler tests use the router so route registration is covered as well.
// They currently depend on the package seed data being non-empty.

// TestLatestPricesReturnsOK verifies the latest-prices route is available.
func TestLatestPricesReturnsOK(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/prices/latest", nil)

	newMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestLatestPricesShareOneDate verifies the response is a single daily snapshot.
func TestLatestPricesShareOneDate(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/prices/latest", nil)

	newMux().ServeHTTP(rec, req)

	var prices []SharePrice
	if err := json.NewDecoder(rec.Body).Decode(&prices); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if len(prices) == 0 {
		t.Fatal("got 0 prices, want at least 1")
	}

	want := prices[0].Date
	for _, p := range prices[1:] {
		if !p.Date.Equal(want) {
			t.Errorf("fund %s has date %s, want %s",
				p.FundCode, p.Date.Format("2006-01-02"), want.Format("2006-01-02"))
		}
	}
}
