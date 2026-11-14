package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestSubsetHeavyBranchReturnsSubsetWithinBound(t *testing.T) {
	dd := core.NewBDDManager(3)
	x0, x1 := Cudd_bddIthVar(dd, 0), Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddOr(dd, x0, x1)
	got := Cudd_SubsetHeavyBranch(dd, f, 3, 1)
	if got == nil {
		t.Fatal("heavy branch subset returned nil")
	}
	if !core.Cudd_bddLeq(dd, got, f) {
		t.Fatal("result is not a subset of the input")
	}
}

func TestSupersetHeavyBranchContainsInput(t *testing.T) {
	dd := core.NewBDDManager(3)
	x0, x1 := Cudd_bddIthVar(dd, 0), Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x0, x1)
	got := Cudd_SupersetHeavyBranch(dd, f, 3, 1)
	if got == nil {
		t.Fatal("heavy branch superset returned nil")
	}
	if !core.Cudd_bddLeq(dd, f, got) {
		t.Fatal("result is not a superset of the input")
	}
}
