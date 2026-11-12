package gia

import "testing"

func TestLutBalancePreservesFunction(t *testing.T) {
	for _, useMuxes := range []int32{0, 1} {
		p := Gia_ManStart(20)
		a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b^1)
		y := Gia_ManAppendMux(p, c, x, d)
		Gia_ManAppendCo(p, y)
		result := Gia_ManLutBalance(p, 4, useMuxes, 0, 0, 0)
		if result == nil || Gia_ManVerifyTwoTruths(p, result) != 1 {
			t.Fatalf("mux mode %d changed function", useMuxes)
		}
	}
}
