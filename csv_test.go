package main

import (
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
