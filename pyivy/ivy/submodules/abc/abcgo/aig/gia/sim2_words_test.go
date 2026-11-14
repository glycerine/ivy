package gia

import "testing"

func TestEquivSimulationRound(t *testing.T) {
	g := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ro := Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b^1)
	y := Gia_ManAppendAnd(g, x, ro)
	Gia_ManAppendCo(g, y^1)
	Gia_ManAppendCo(g, x)
	Gia_ManSetRegNum(g, 1)
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	pars.NWords = 2
	fast := Gia_ManSimCreate(g, &pars)
	defer Gia_ManSimDelete(fast)
	eq := Gia_Sim2Create(g, &pars)
	defer Gia_Sim2Delete(eq)
	Gia_ManResetRandom(&pars)
	Gia_ManSimInfoInit(fast)
	Gia_ManResetRandom(&pars)
	Gia_Sim2SimulateRound(eq)
	Gia_ManSimulateRound(fast)
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		if Gia_Sim2CompareEqual(Gia_SimDataCo(fast, i), Gia_Sim2Data(eq, Gia_ManCo(g, i).id), pars.NWords, 0) != 1 {
			t.Fatalf("first frame output %d", i)
		}
	}
	Gia_Sim2InfoTransfer(eq)
	Gia_ManResetRandom(&pars)
	Gia_ManSimInfoTransfer(fast)
	Gia_ManResetRandom(&pars)
	Gia_Sim2SimulateRound(eq)
	Gia_ManSimulateRound(fast)
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		if Gia_Sim2CompareEqual(Gia_SimDataCo(fast, i), Gia_Sim2Data(eq, Gia_ManCo(g, i).id), pars.NWords, 0) != 1 {
			t.Fatalf("second frame output %d", i)
		}
	}
}
