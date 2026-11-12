package gia

import "testing"

func TestStructuredNormalizationRoundTrip(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b^1)
	y := Gia_ManAppendMux(p, c, x, d)
	Gia_ManAppendCo(p, y)
	ntk := Str_ManNormalize(p)
	defer Str_NtkDelete(ntk)
	result := Str_NtkToGia(p, ntk)
	if result == nil || Gia_ManVerifyTwoTruths(p, result) != 1 {
		t.Fatal("normalization changed Boolean function")
	}
}
