package fraig

import "testing"

func TestMiterReturnsAtFirstUndecidedOutput(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a := Fraig_ManReadIthVar(p, 0)
	clear(a.puSimR)
	Fraig_ManSetPo(p, a)
	Fraig_ManSetPo(p, p.pConst1)
	if got := Fraig_ManCheckMiter(p); got != -1 || p.pModel != nil {
		t.Fatalf("first undecided output must stop the C miter check, got %d model %v", got, p.pModel)
	}
}

// ABC fraigFeed.c tries random patterns before dynamic patterns and copies the
// selected bit for each input. fraigSat.c frees the preceding model each time.
func TestMiterWitnessPatternOrderAndRepeatedChecks(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	clear(a.puSimR)
	clear(b.puSimR)
	a.puSimR[0] = 1 << 3
	a.puSimD[0], b.puSimD[0] = 1<<1, 1<<1
	p.iWordStart = 1
	Fraig_ManSetPo(p, a)
	if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 2 || p.pModel[0] != 1 || p.pModel[1] != 0 {
		t.Fatalf("random witness: status=%d model=%v", got, p.pModel)
	}
	clear(a.puSimR)
	if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 2 || p.pModel[0] != 1 || p.pModel[1] != 1 {
		t.Fatalf("dynamic witness: status=%d model=%v", got, p.pModel)
	}
	p.vOutputs.pArray[0] = Fraig_Not(a)
	b.puSimR[0] = 1
	if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 2 || p.pModel[0] != 0 || p.pModel[1] != 1 {
		t.Fatalf("complemented witness: status=%d model=%v", got, p.pModel)
	}
	p.vOutputs.pArray[0] = a
	clear(a.puSimD)
	if got := Fraig_ManCheckMiter(p); got != -1 || p.pModel != nil {
		t.Fatalf("undecided must clear prior witness: status=%d model=%v", got, p.pModel)
	}
	p.vOutputs.pArray[0] = p.pConst1
	if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 2 || p.pModel[0] != 0 || p.pModel[1] != 0 {
		t.Fatalf("constant-one witness: status=%d model=%v", got, p.pModel)
	}
	p.vOutputs.pArray[0] = Fraig_Not(p.pConst1)
	if got := Fraig_ManCheckMiter(p); got != 1 || p.pModel != nil {
		t.Fatalf("constant-zero must clear witness: status=%d model=%v", got, p.pModel)
	}
}
