package gia

import "testing"

func TestEquivSimulationDriverCounterexample(t *testing.T) {
	g := Gia_ManStart(10)
	a := Gia_ManAppendCi(g)
	ro := Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, ro)
	Gia_ManAppendCo(g, out)
	Gia_ManAppendCo(g, a)
	Gia_ManSetRegNum(g, 1)
	g.PReprs = make([]Gia_Rpr_t, g.nObjs)
	g.PNexts = make([]int32, g.nObjs)
	for id := int32(1); id < g.nObjs; id++ {
		Gia_ObjSetRepr(g, id, GIA_VOID)
	}
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	pars.NWords = 2
	pars.NIters = 4
	pars.FCheckMiter = 1
	pars.TimeLimit = 0
	if Gia_ManSimSimulate(g, &pars) != 1 || g.pCexSeq == nil || g.pCexSeq.IFrame != 1 || Gia_ManVerifyCex(g, g.pCexSeq, 0) != 1 {
		t.Fatalf("invalid counterexample %+v", g.pCexSeq)
	}
}
func TestEquivSimulationDriverRefinesConstants(t *testing.T) {
	g := Gia_ManStart(6)
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	g.PReprs = make([]Gia_Rpr_t, g.nObjs)
	g.PNexts = make([]int32, g.nObjs)
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	pars.NWords = 1
	pars.NIters = 1
	pars.TimeLimit = 0
	if Gia_ManSimSimulate(g, &pars) != 0 || Gia_ObjRepr(g, a>>1) != GIA_VOID {
		t.Fatal("input remained a constant candidate")
	}
}
