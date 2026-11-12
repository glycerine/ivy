package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestEssentialLiteralsAndTwoLiteralClauses(t *testing.T) {
	dd := NewBDDManager(2)
	x, y := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, Cudd_Not(y))
	if Cudd_bddIsVarEssential(dd, f, 0, 1) != 1 || Cudd_bddIsVarEssential(dd, f, 1, 0) != 1 {
		t.Fatal("failed to find implied literal phases")
	}
	if Cudd_FindEssential(dd, f) != f {
		t.Fatal("essential cube differs from the function")
	}
	xor := core.Cudd_bddXor(dd, x, y)
	info := Cudd_FindTwoLiteralClauses(dd, xor)
	if info == nil || len(info.clauses) != 2 {
		t.Fatalf("XOR clause count = %d, want 2", len(info.clauses))
	}
}
