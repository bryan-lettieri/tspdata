package main

import (
	"os"
	"slices"
	"testing"
)

func TestFundCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"C Fund", "C"},
		{"L 2030", "L2030"},
		{"L Income", "LINCOME"},
		{"L 2080", "L2080"},
	}

	for _, tt := range tests {
		got := fundCode(tt.input)
		if got != tt.want {
			t.Errorf("%q got %q, expected %q",
				tt.input,
				got,
				tt.want,
			)
		}
	}
}

func TestParseCSVFixture(t *testing.T) {
	const (
		wantRecords = 68
		wantCodes   = 16
	)
	// open input file
	f, err := os.Open("testdata/prices.csv")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	// close input on exit
	defer f.Close()

	prices, err := parseCSV(f)
	if err != nil {
		t.Fatalf("parseCSV() returned error: %v", err)
	}

	// Test for correct record count
	t.Run("record count", func(t *testing.T) {
		if len(prices) != wantRecords {
			t.Errorf("got record count = %d, want %d", len(prices), wantRecords)
		}
	})

	// Test for correct number of distinct codes
	t.Run("canonical codes", func(t *testing.T) {
		distinctCodes := []string{}
		for _, p := range prices {

			code := p.FundCode
			if !slices.Contains(distinctCodes, code) {
				distinctCodes = append(distinctCodes, code)
			}
		}
		if len(distinctCodes) != wantCodes {
			t.Errorf("got code count = %d, want %d", len(distinctCodes), wantCodes)
		}
	})

	// Test no record has price of 0
	t.Run("no zero prices", func(t *testing.T) {
		for _, p := range prices {
			if p.Price == 0 {
				t.Errorf("%s on %s has a Price of 0", p.FundCode, p.Date.Format(dateLayout))
			}
		}
	})
}
