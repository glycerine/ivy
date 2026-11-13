package fraig

import "testing"

func TestFunctionalReductionHonorsInspectionLimit(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 1, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b, c, d := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2), Fraig_ManReadIthVar(p, 3)
	first := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), c)
	if got := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, c), b); got != first {
		t.Fatal("initial equivalence not proved")
	}
	if Fraig_ManReadInspects(p) <= 0 {
		t.Fatal("initial proof did not inspect clauses")
	}
	p.nInspLimit = 1
	left := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), d)
	before := p.nSatCalls
	right := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, d), b)
	if p.nSatCalls != before {
		t.Fatal("SAT called after inspection budget was exhausted")
	}
	if left == right {
		t.Fatal("unproved candidate unexpectedly merged")
	}
}
