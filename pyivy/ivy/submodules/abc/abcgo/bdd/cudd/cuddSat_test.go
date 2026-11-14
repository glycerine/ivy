package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCuddEvalFollowsComplementedEdges(t *testing.T) {
	dd := NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, core.Cudd_Not(y))

	for _, tc := range []struct {
		inputs [2]int32
		want   *DdNode
	}{
		{[2]int32{0, 0}, dd.LogicZero},
		{[2]int32{0, 1}, dd.LogicZero},
		{[2]int32{1, 0}, dd.One},
		{[2]int32{1, 1}, dd.LogicZero},
	} {
		if got := Cudd_Eval(dd, f, &tc.inputs[0]); got != tc.want {
			t.Errorf("Cudd_Eval(%v) = %p, want %p", tc.inputs, got, tc.want)
		}
	}
}

func TestCuddShortestPathAndLength(t *testing.T) {
	dd := NewBDDManager(2)
	x, y := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddOr(dd, x, y)
	weights := [2]int32{1, 3}
	support := [2]int32{}
	var length int32
	path := Cudd_ShortestPath(dd, f, &weights[0], &support[0], &length)
	if path == nil || !core.Cudd_bddLeq(dd, path, f) {
		t.Fatalf("shortest path %p is not an implicant", path)
	}
	if length != 1 || Cudd_ShortestLength(dd, f, &weights[0]) != 1 {
		t.Fatalf("shortest length = %d, want 1", length)
	}
	if support != [2]int32{1, 1} {
		t.Fatalf("support = %v, want both variables", support)
	}
	if !core.Cudd_bddLeq(dd, x, path) {
		t.Fatalf("weighted shortest path should select x")
	}
}

func TestCuddLargestCube(t *testing.T) {
	dd := NewBDDManager(2)
	x, y := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddOr(dd, x, y)
	var length int32
	cube := Cudd_LargestCube(dd, f, &length)
	if cube == nil || !core.Cudd_bddLeq(dd, cube, f) {
		t.Fatalf("largest cube %p is not an implicant", cube)
	}
	if length != 1 {
		t.Fatalf("largest cube length = %d, want 1", length)
	}
}

func TestCuddShortestPathComplementedFunction(t *testing.T) {
	dd := NewBDDManager(2)
	x, y := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_Not(core.Cudd_bddAnd(dd, x, y))
	var length int32
	path := Cudd_ShortestPath(dd, f, nil, nil, &length)
	if path == nil || !core.Cudd_bddLeq(dd, path, f) {
		t.Fatalf("complemented shortest path %p is not an implicant", path)
	}
	if length != 0 {
		t.Fatalf("complemented shortest path length = %d, want 0", length)
	}
}
