package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"testing"
)

func TestSolveEqnConsistencyAndSubstitution(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	cube := core.Cudd_bddAnd(dd, y, z)
	fs := []*DdNode{x, y, core.Cudd_bddXor(dd, x, y), core.Cudd_bddIte(dd, x, y, z), core.Cudd_bddAnd(dd, y, z), core.Cudd_Not(core.Cudd_bddAnd(dd, x, z))}
	for _, f := range fs {
		got, g, indices := Cudd_SolveEqn(dd, f, cube, 2)
		if got == nil {
			t.Fatal("solve failed")
		}
		if indices[0] != 1 || indices[1] != 2 {
			t.Fatalf("indices %v", indices)
		}
		expected := core.Cudd_Not(core.Cudd_bddExistAbstract(dd, core.Cudd_Not(f), cube))
		if got != expected {
			t.Fatalf("wrong consistency condition")
		}
		if check := Cudd_VerifySol(dd, f, g, indices, 2); check != got {
			t.Fatalf("substitution does not produce consistency condition")
		}
	}
}
