package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestSplitSetTrivialAndExactCover(t *testing.T) {
	dd := core.NewBDDManager(3)
	x, y, z := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1), core.Cudd_bddIthVar(dd, 2)
	f := core.Cudd_bddOr(dd, core.Cudd_bddAnd(dd, x, y), core.Cudd_bddAnd(dd, x, z))
	vars := []*DdNode{x, y, z}

	if got := Cudd_SplitSet(dd, f, vars, 3, 0); got != core.Cudd_ReadLogicZero(dd) {
		t.Fatalf("zero-sized split = %p, want logic zero", got)
	}
	if got := Cudd_SplitSet(dd, f, vars, 3, 3); got != f {
		t.Fatalf("full split = %p, want original %p", got, f)
	}
	part := Cudd_SplitSet(dd, f, vars, 3, 1)
	if part == nil {
		t.Fatal("one-minterm split returned nil")
	}
	if count := core.Cudd_CountMinterm(dd, part, 3); count != 1 {
		t.Fatalf("split contains %g minterms, want 1", count)
	}
	if remainder := core.Cudd_bddAnd(dd, part, core.Cudd_Not(f)); remainder != core.Cudd_ReadLogicZero(dd) {
		t.Fatal("split contains minterms outside input")
	}
}

func TestSplitSetRejectsOversizedRequest(t *testing.T) {
	dd := core.NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	if got := Cudd_SplitSet(dd, x, []*DdNode{x}, 1, 3); got != nil {
		t.Fatalf("oversized split = %p, want nil", got)
	}
}
