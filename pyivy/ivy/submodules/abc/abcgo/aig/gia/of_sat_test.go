package gia

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestOfSATMappingBound(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	out := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	ofDefaultParameters(&pars)
	cuts := newOfCutManager(g, &pars)
	cuts.computeCuts()
	mapper := newOfMapper(cuts)
	if got := mapper.satProblem(0).Solve(); got != satoko.SATOKO_UNSAT {
		t.Fatalf("zero LUTs: %d", got)
	}
	if got := mapper.satProblem(1).Solve(); got != satoko.SATOKO_SAT {
		t.Fatalf("one LUT: %d", got)
	}
	pars.NLutSize = 2
	cuts = newOfCutManager(g, &pars)
	cuts.computeCuts()
	mapper = newOfMapper(cuts)
	if got := mapper.satProblem(1).Solve(); got != satoko.SATOKO_UNSAT {
		t.Fatalf("two-input LUT bound: %d", got)
	}
	if got := mapper.satProblem(2).Solve(); got != satoko.SATOKO_SAT {
		t.Fatalf("two LUTs: %d", got)
	}
}
