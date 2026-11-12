package util

import "testing"

func TestXSatFloatArithmetic(t *testing.T) {
	for _, v := range []float32{1, 1.5, 2, 2.5, 3, 12, 32768} {
		a := xSat_FloatFromFloat(v)
		if got := xSat_Float2Float(a); got != v {
			t.Fatalf("roundtrip %g -> %g", v, got)
		}
		if xSat_Uint2Float(xSat_Float2Uint(a)) != a {
			t.Fatal("packing")
		}
	}
	a, b := xSat_FloatFromFloat(1.5), xSat_FloatFromFloat(2.5)
	if v := xSat_Float2Float(xSat_FloatAdd(a, b)); v != 4 {
		t.Fatal(v)
	}
	if v := xSat_Float2Float(xSat_FloatMul(a, b)); v != 3.75 {
		t.Fatal(v)
	}
	if xSat_FloatDiv(a, 2) != xSat_FloatCreateConst1() {
		t.Fatal("underflow")
	}
	if xSat_FloatAdd(xSat_FloatCreateMaximum(), xSat_FloatCreateMaximum()) != xSat_FloatCreateMaximum() {
		t.Fatal("overflow")
	}
	if !xSat_LessThan(a, b) || !xSat_Equal(a, a) {
		t.Fatal("comparison")
	}
}
