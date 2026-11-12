package core

import "testing"

func TestBooleanADDBridge(t *testing.T) {
	dd := NewBDDManager(3)
	for _, f := range []*DdNode{dd.One, dd.LogicZero, dd.Vars[0], Cudd_Not(dd.Vars[0]), Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2])} {
		a := Cudd_BddToAdd(dd, f)
		if a == nil || Cudd_addBddPattern(dd, a) != f {
			t.Fatal("roundtrip", f)
		}
	}
	a := Cudd_addIte(dd, Cudd_addIthVar(dd, 1), Cudd_addConst(dd, -3), dd.Zero)
	if Cudd_addBddPattern(dd, a) != dd.Vars[1] {
		t.Fatal("numeric pattern")
	}
	if Cudd_BddToAdd(dd, a) != nil || Cudd_addBddPattern(dd, dd.LogicZero) != nil {
		t.Fatal("accepted wrong family")
	}
}
func TestNewVariableAtLevelPreservesReferencedGraphs(t *testing.T) {
	dd := NewBDDManager(3)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2])
	a := Cudd_BddToAdd(dd, f)
	Cudd_Ref(f)
	Cudd_Ref(a)
	for level := 0; level < 3; level++ {
		v := Cudd_bddNewVarAtLevel(dd, level)
		if v == nil || dd.Perm[v.Index] != level {
			t.Fatal("level", level)
		}
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := make([]bool, len(dd.Vars))
		for i := 0; i < 3; i++ {
			values[i] = assignment&(1<<i) != 0
		}
		b, err := dd.Evaluate(f, values)
		number, aerr := dd.EvaluateADD(a, values)
		want := values[0] != values[2]
		if err != nil || aerr != nil || b != want || (number != 0) != want {
			t.Fatal("lost graph", assignment)
		}
	}
	if Cudd_bddNewVarAtLevel(dd, -1) != nil || Cudd_bddNewVarAtLevel(dd, len(dd.Vars)+1) != nil {
		t.Fatal("invalid level")
	}
}

func TestVariableInsertionDoesNotCollectOrReorder(t *testing.T) {
	dd := NewBDDManager(3)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2]) // deliberately unreferenced
	before := Cudd_ReadReorderings(dd)
	if Cudd_bddNewVarAtLevel(dd, 1) == nil {
		t.Fatal("insertion failed")
	}
	if Cudd_ReadReorderings(dd) != before {
		t.Fatal("insertion counted as reordering")
	}
	if Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2]) != f {
		t.Fatal("insertion collected a live unreferenced node")
	}
}
