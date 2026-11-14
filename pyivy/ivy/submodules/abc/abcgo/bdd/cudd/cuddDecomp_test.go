package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestVariableDecompositionReconstructsFunction(t *testing.T) {
	dd := NewBDDManager(3)
	x, y := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, y)
	var factors []*DdNode
	if n := Cudd_bddVarConjDecomp(dd, f, &factors); n != 2 {
		t.Fatalf("conjunct count = %d, want 2", n)
	}
	if core.Cudd_bddAnd(dd, factors[0], factors[1]) != f {
		t.Fatal("conjunct factors do not reconstruct f")
	}
	var disjuncts []*DdNode
	if n := Cudd_bddVarDisjDecomp(dd, f, &disjuncts); n != 1 {
		t.Fatalf("disjunct count = %d, want 1 for the source's trivial split", n)
	}
	reconstructed := disjuncts[0]
	for _, term := range disjuncts[1:] {
		reconstructed = core.Cudd_bddOr(dd, reconstructed, term)
	}
	if reconstructed != f {
		t.Fatal("disjunct factors do not reconstruct f")
	}
}
