package abci

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestAbcResCofCountAndCostSourceCases(t *testing.T) {
	dd := core.NewBDDManager(3)
	x0, x1 := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddXor(dd, x0, x1)
	var check, cofactors int32
	if got := Abc_ResCofCount(dd, f, 1, &check); got != 2 {
		t.Fatalf("cofactor count on x0 = %d, want 2", got)
	}
	if check != 1 {
		t.Fatalf("one-variable non-strict check = %d, want 1", check)
	}
	if got := Abc_ResCost(dd, f, 1, &cofactors, &check); got != 10001 || cofactors != 2 {
		t.Fatalf("source partition cost/count = %d/%d, want 10001/2", got, cofactors)
	}
}

func TestAbcNtkBddCofCountComplementAndConstant(t *testing.T) {
	dd := core.NewBDDManager(2)
	x, y := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddXor(dd, x, y)
	if got := Abc_NtkBddCofCount(dd, f, []*core.DdNode{x}, 1); got != 2 {
		t.Fatalf("one-variable cofactor count = %d, want 2", got)
	}
	if got := Abc_NtkBddCofCount(dd, core.Cudd_Not(f), []*core.DdNode{x, y}, 2); got != 2 {
		t.Fatalf("complemented function cofactor count = %d, want 2", got)
	}
	if got := Abc_NtkBddCofCount(dd, f, nil, 0); got != 1 {
		t.Fatalf("empty cofactor set count = %d, want 1", got)
	}
}

func TestCascadePartitionUsesCDefaultRandSequence(t *testing.T) {
	rng := newCascadeCRand(1)
	if got := rng.next(); got != 1804289383 {
		t.Fatalf("first C rand() value = %d, want 1804289383", got)
	}
	if got := rng.next(); got != 846930886 {
		t.Fatalf("second C rand() value = %d, want 846930886", got)
	}
}

func TestCascadeCharacteristicBddAddSourceEncoding(t *testing.T) {
	dd := core.NewBDDManager(1)
	core.Cudd_addNewVarAtLevel(dd, len(dd.Vars))
	core.Cudd_addNewVarAtLevel(dd, len(dd.Vars))
	x := core.Cudd_bddIthVar(dd, 0)
	core.Cudd_bddIthVar(dd, 1)
	core.Cudd_bddIthVar(dd, 2)
	functions := []*core.DdNode{x, core.Cudd_Not(x)}
	relation := Abc_NtkBddDecCharFunc(dd, functions, 2, 3, 2)
	if relation == nil {
		t.Fatal("characteristic BDD construction failed")
	}
	add := Abc_NtkBddToAdd(dd, relation, 2)
	if add == nil {
		t.Fatal("BDD to ADD conversion failed")
	}
	for _, tc := range []struct {
		input bool
		want  float64
	}{{false, 2}, {true, 1}} {
		if got := evalCascadeTestADD(add, []bool{tc.input}); got != tc.want {
			t.Fatalf("encoded output for input %v = %g, want %g", tc.input, got, tc.want)
		}
	}
	reconstructed := Abc_NtkAddToBdd(dd, add, 1, 2)
	if reconstructed == nil {
		t.Fatal("ADD to BDD conversion failed")
	}
	for _, tc := range []struct {
		assignment []bool
		want       bool
	}{
		{[]bool{false, true, false}, true}, // source's MSB-first output code
		{[]bool{true, false, true}, true},
		{[]bool{false, false, true}, false},
	} {
		if got := evalLutTestBDD(dd, reconstructed, tc.assignment); got != tc.want {
			t.Fatalf("decoded assignment %v = %v, want %v", tc.assignment, got, tc.want)
		}
	}
}

func evalCascadeTestADD(edge *core.DdNode, assignment []bool) float64 {
	if core.Cudd_IsConstant(edge) {
		return core.Cudd_Regular(edge).Value
	}
	node := core.Cudd_Regular(edge)
	if assignment[node.Index] {
		return evalCascadeTestADD(node.T, assignment)
	}
	return evalCascadeTestADD(node.E, assignment)
}
