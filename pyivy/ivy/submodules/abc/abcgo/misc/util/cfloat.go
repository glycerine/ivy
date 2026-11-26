package util

import (
	"math"
	"strconv"
)

// CFloat formats v like C's printf("%.*f", prec, v), which writes infinities
// and NaNs as "inf", "-inf" and "nan" (Go writes "+Inf", "-Inf", "NaN").
func CFloat(v float64, prec int) string {
	switch {
	case math.IsInf(v, 1):
		return "inf"
	case math.IsInf(v, -1):
		return "-inf"
	case math.IsNaN(v):
		return "nan"
	}
	return strconv.FormatFloat(v, 'f', prec, 64)
}
