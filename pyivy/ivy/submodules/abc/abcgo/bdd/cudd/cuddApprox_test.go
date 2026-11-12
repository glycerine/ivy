package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestApproximationContracts(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1), core.Cudd_bddIthVar(dd, 2)
	f := core.Cudd_bddOr(dd, core.Cudd_bddAnd(dd, x, y), core.Cudd_bddAnd(dd, Cudd_Not(x), z))
	under := Cudd_UnderApprox(dd, f, 3, 1, 0, 0.1)
	if under == nil || core.Cudd_bddAnd(dd, under, Cudd_Not(f)) != Cudd_ReadLogicZero(dd) {
		t.Fatal("under-approximation is not a subset")
	}
	over := Cudd_OverApprox(dd, f, 3, 1, 0, 0.1)
	if over == nil || core.Cudd_bddAnd(dd, f, Cudd_Not(over)) != Cudd_ReadLogicZero(dd) {
		t.Fatal("over-approximation is not a superset")
	}
	biased := Cudd_BiasedUnderApprox(dd, f, x, 3, 1, 0.1, 0.1)
	if biased == nil || core.Cudd_bddAnd(dd, biased, Cudd_Not(f)) != Cudd_ReadLogicZero(dd) {
		t.Fatal("biased under-approximation is not a subset")
	}
	if core.Cudd_bddAnd(dd, x, core.Cudd_bddXor(dd, biased, f)) != Cudd_ReadLogicZero(dd) {
		t.Fatal("biased under-approximation changed f where the bias is true")
	}
}
