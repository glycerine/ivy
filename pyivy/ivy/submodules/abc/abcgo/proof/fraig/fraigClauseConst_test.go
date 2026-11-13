package fraig

import "testing"

func TestSatClauseTreatsConstantAsUnconstrainedVariable(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	x := Fraig_ManReadIthVar(p, 0)
	// The C solver does not install a unit clause for FRAIG's constant node.
	if got := Fraig_ManCheckClauseUsingSat(p, p.pConst1, x, -1); got != 0 {
		t.Fatalf("expected an unconstrained constant SAT variable, got %d", got)
	}
}
