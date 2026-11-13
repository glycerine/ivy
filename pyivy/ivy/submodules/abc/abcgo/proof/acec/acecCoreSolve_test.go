package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestFallbackCecTransfersCounterexample(t *testing.T) {
	makeGraph := func(inv int32) *Gia_Man_t {
		p := gia.Gia_ManStart(8)
		a := gia.Gia_ManAppendCi(p)
		gia.Gia_ManAppendCo(p, a^inv)
		return p
	}
	pars := &Acec_ParCec_t{}
	Acec_ManCecSetDefaultParams(pars)
	a, b := makeGraph(0), makeGraph(1)
	if got := Acec_Solve(a, b, pars); got != 0 {
		t.Fatalf("different graphs status %d", got)
	}
	if a.CombinationalCounterexample() == nil {
		t.Fatal("missing transferred witness")
	}
	a, b = makeGraph(0), makeGraph(0)
	if got := Acec_Solve(a, b, pars); got != 1 {
		t.Fatalf("equivalent graphs status %d", got)
	}
}
