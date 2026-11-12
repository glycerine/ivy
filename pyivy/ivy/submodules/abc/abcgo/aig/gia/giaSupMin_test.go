package gia

import "testing"

func TestGiaManSupExperiment(t *testing.T) {
	g := Gia_ManStart(8)
	a := Gia_ManAppendCi(g)
	b := Gia_ManAppendCi(g)
	and := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, and^1)
	pairs := Vec_IntAlloc(2)
	Vec_IntPush(pairs, 0)
	Vec_IntPush(pairs, 1)
	Gia_ManSupExperiment(g, pairs)
	p := Gia_ManSupStart(6)
	if p.nWordsMax != 2 || len(p.pTruthIn) != 2 || len(p.pTruthOut) != 2 {
		t.Fatalf("unexpected support manager word allocation: %+v", p)
	}
	Gia_ManSupStop(p)
	if p.pTruthIn != nil || p.pTruthOut != nil {
		t.Fatal("support manager retained its truth tables")
	}
}
