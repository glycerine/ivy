package gia

import "testing"

func TestPreparedCombinationalVerification(t *testing.T) {
	makeGraph := func(invert bool, register bool) *Gia_Man_t {
		p := Gia_ManStart(8)
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b)
		if invert {
			x ^= 1
		}
		Gia_ManAppendCo(p, x)
		Gia_ManAppendCo(p, b)
		if register {
			Gia_ManSetRegNum(p, 1)
		}
		return p
	}
	a, b, c := makeGraph(false, true), makeGraph(false, false), makeGraph(true, false)
	if status := giaVerifyPreparedCombinational(a, b, 1000, 0, 0); status != 1 {
		t.Fatalf("equivalent status %d", status)
	}
	if status := giaVerifyPreparedCombinational(a, c, 1000, 0, 0); status != 0 {
		t.Fatalf("different status %d", status)
	}
}
