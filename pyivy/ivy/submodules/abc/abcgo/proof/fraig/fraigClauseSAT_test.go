package fraig

import "testing"

func TestSatClauseValidity(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	both := Fraig_NodeAnd(p, a, b)
	if Fraig_ManCheckClauseUsingSat(p, Fraig_Not(both), a, -1) != 1 {
		t.Fatal("AND implies its input")
	}
	if Fraig_ManCheckClauseUsingSat(p, a, b, -1) != 0 {
		t.Fatal("independent inputs need not satisfy their disjunction")
	}
	if p.nSatCounterImp == 0 {
		t.Fatal("satisfiable clause negation not recorded")
	}
}
