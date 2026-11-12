package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSatImprovesPatternCoverage(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b^1)
	Gia_ManAppendCo(p, x)
	p.vSimsPi = vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(p.vSimsPi, 0xaaaaaaaaaaaaaaaa)
	vec.Vec_WrdPush(p.vSimsPi, 0xaaaaaaaaaaaaaaaa)
	Gia_ManPatSatImprove(p, 1, 0)
	if vec.Vec_WrdSize(p.vSimsPi) != 4 {
		t.Fatalf("expected an additional word per input, got %d", vec.Vec_WrdSize(p.vSimsPi))
	}
	sims := Gia_ManSimPatSim(p)
	words := vec.Vec_WrdArray(sims)
	if words[(x>>1)*2+1] == 0 {
		t.Fatal("new patterns did not expose miter counterexample")
	}
}
