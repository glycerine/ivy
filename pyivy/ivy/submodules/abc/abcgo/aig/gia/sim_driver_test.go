package gia

import "testing"

func TestSimulationDriverCounterexample(t *testing.T) {
	g := Gia_ManStart(10)
	a := Gia_ManAppendCi(g)
	ro := Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, ro)
	Gia_ManAppendCo(g, out)
	Gia_ManAppendCo(g, a)
	Gia_ManSetRegNum(g, 1)
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	pars.NWords = 2
	pars.NIters = 4
	pars.FCheckMiter = 1
	pars.TimeLimit = 0
	equiv := func(*Gia_Man_t, *Gia_ParSim_t) int32 { t.Fatal("unexpected equivalence simulation"); return 0 }
	if giaSimSimulate(g, &pars, equiv) != 1 || g.pCexSeq == nil || g.pCexSeq.IFrame != 1 || Gia_ManVerifyCex(g, g.pCexSeq, 0) != 1 {
		t.Fatalf("invalid counterexample %+v", g.pCexSeq)
	}
}

func TestSimulationDriverNoFailure(t *testing.T) {
	g := Gia_ManStart(4)
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	pars.NIters = 3
	pars.TimeLimit = 0
	if giaSimSimulate(g, &pars, nil) != 0 || g.pCexSeq != nil {
		t.Fatal("unexpected counterexample")
	}
}
