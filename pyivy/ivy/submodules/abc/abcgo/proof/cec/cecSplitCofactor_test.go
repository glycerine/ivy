package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestSplitSelectsMostReferencedPrimaryInput(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a, b := gia.Gia_ManAppendAnd(g, x, y), gia.Gia_ManAppendAnd(g, x, z)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	order := Gia_PermuteSpecialOrder(g)
	if len(order) != 3 || order[0] != 0 {
		t.Fatalf("reference order %v", order)
	}
	var fanouts, cost int32
	if v := Gia_SplitCofVar2(g, &fanouts, &cost); v != 0 || fanouts != 2 || cost != -1 {
		t.Fatalf("variable=%d fanouts=%d cost=%d", v, fanouts, cost)
	}
}

func TestSplitLookaheadUsesMinimumCofactorCost(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	b := gia.Gia_ManAppendAnd(g, a, z)
	gia.Gia_ManAppendCo(g, b)
	var want int32 = -1
	var best int32 = 1 << 30
	for _, v := range Gia_PermuteSpecialOrder(g) {
		zero := gia.Gia_ManDupCofactorVar(g, v, 0)
		one := gia.Gia_ManDupCofactorVar(g, v, 1)
		c := gia.Gia_ManAndNum(zero) + gia.Gia_ManAndNum(one)
		if c < best {
			want, best = v, c
		}
	}
	var fanouts, cost int32
	if got := Gia_SplitCofVar(g, 3, &fanouts, &cost); got != want || cost != best {
		t.Fatalf("variable=%d cost=%d, want variable=%d cost=%d", got, cost, want, best)
	}
}
