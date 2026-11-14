package gia

import "testing"

func TestMfManagerInitialization(t *testing.T) {
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	if pars.NLutSize != 6 || pars.NCutNum != 8 || pars.DelayTarget != -1 || pars.FOptEdge != 1 {
		t.Fatalf("defaults %+v", pars)
	}
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	pars.FCutMin = 1
	p := Mf_ManAlloc(g, &pars)
	if p.vTtMem == nil || p.iCur != 2 || len(p.pLfObjs) != int(g.nObjs) || p.pLfObjs[x>>1].nFlowRefs != 1 {
		t.Fatal("incorrect initialized state")
	}
	Mf_ManFree(p)
	if p.vTtMem != nil || p.pLfObjs != nil {
		t.Fatal("manager not released")
	}
}
