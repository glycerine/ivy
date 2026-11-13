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
