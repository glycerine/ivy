package util

import (
	"math"
	"testing"
)

func TestCFloatMatchesCPrintf(t *testing.T) {
	for _, c := range []struct {
		v    float64
		want string
	}{{math.Inf(1), "inf"}, {math.Inf(-1), "-inf"}, {math.NaN(), "nan"}, {2.345, "2.35"}, {0, "0.00"}} {
		if got := CFloat(c.v, 2); got != c.want {
			t.Errorf("CFloat(%v) = %q, want %q", c.v, got, c.want)
		}
	}
}
