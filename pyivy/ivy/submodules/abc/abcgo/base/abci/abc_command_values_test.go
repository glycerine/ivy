package abci

import (
	"math"
	"testing"
)

func TestCommandFloatNumericPrefixes(t *testing.T) {
	for _, tc := range []struct {
		text string
		want float64
	}{
		{" \t-1.25e2seconds", -125}, {".5x", .5}, {"1.e-2", .01},
		{"1e+garbage", 1}, {"0x1.fp3suffix", 15.5}, {"0XAp-1", 5},
		{"0x10tail", 16}, {"garbage", 0}, {"+", 0}, {"", 0},
		{"-infinitytail", math.Inf(-1)}, {"1e999", math.Inf(1)},
	} {
		if got := atof(tc.text); got != tc.want {
			t.Errorf("atof(%q)=%v want %v", tc.text, got, tc.want)
		}
	}
	if !math.IsNaN(atof("-nan(payload)tail")) {
		t.Fatal("NaN prefix lost")
	}
}
