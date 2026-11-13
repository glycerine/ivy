package fraig

import "testing"

func TestSatClauseWithConstantLiterals(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	x := Fraig_ManReadIthVar(p, 0)
	if got := Fraig_ManCheckClauseUsingSat(p, p.pConst1, x, -1); got != 1 {
		t.Fatalf("true OR x must hold, got %d", got)
	}
	if got := Fraig_ManCheckClauseUsingSat(p, Fraig_Not(p.pConst1), x, -1); got != 0 {
		t.Fatalf("false OR x is not tautological, got %d", got)
	}
}
