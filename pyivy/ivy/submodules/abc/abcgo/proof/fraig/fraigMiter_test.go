package fraig

import "testing"

func TestMiterChecksLaterSatisfiableOutput(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a := Fraig_ManReadIthVar(p, 0)
	clear(a.puSimR)
	Fraig_ManSetPo(p, a)
	Fraig_ManSetPo(p, p.pConst1)
	if got := Fraig_ManCheckMiter(p); got != 0 || p.pModel == nil {
		t.Fatalf("later constant-one output must prove SAT, got %d model %v", got, p.pModel)
	}
}
