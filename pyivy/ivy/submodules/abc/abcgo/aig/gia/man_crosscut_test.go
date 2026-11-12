package gia

import "testing"

func TestGiaCrossCutOrders(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	Gia_ManAppendCo(g, x)
	for _, reverse := range []int32{0, 1} {
		if got := Gia_ManCrossCut(g, reverse); got < 2 || got > 4 {
			t.Fatalf("reverse %d crosscut %d", reverse, got)
		}
	}
}
