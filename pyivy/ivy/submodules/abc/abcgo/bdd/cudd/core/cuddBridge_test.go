package core

import (
	"math"
	"testing"
)

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

func TestADDThresholdIntervalAndBits(t *testing.T) {
	leaves := []float64{-2147483648, -17.75, -1.25, 0, 0.75, 2.5, 17, 2147483647}
	dd := NewBDDManager(3)
	var build func(int, int, int) *DdNode
	build = func(level, start, end int) *DdNode {
		if end-start == 1 {
			return Cudd_addConst(dd, leaves[start])
		}
		middle := (start + end) / 2
		return Cudd_addIte(dd, Cudd_addIthVar(dd, level), build(level+1, middle, end), build(level+1, start, middle))
	}
	f := build(0, 0, len(leaves))
	type test struct {
		f    *DdNode
		want func(float64) bool
	}
	tests := []test{}
	for _, threshold := range []float64{-20, -1, 0, 0.75, 17, 100} {
		v := threshold
		tests = append(tests, test{Cudd_addBddThreshold(dd, f, v), func(x float64) bool { return !(x < v) }}, test{Cudd_addBddStrictThreshold(dd, f, v), func(x float64) bool { return !(x <= v) }})
	}
	for _, interval := range [][2]float64{{-20, 20}, {-1, 0}, {0, 0}, {2, -2}} {
		lo, hi := interval[0], interval[1]
		tests = append(tests, test{Cudd_addBddInterval(dd, f, lo, hi), func(x float64) bool { return !(x < lo || x > hi) }})
	}
	for bit := 0; bit < 32; bit++ {
		b := bit
		tests = append(tests, test{Cudd_addBddIthBit(dd, f, b), func(x float64) bool { return uint32(int32(x))&(1<<uint(b)) != 0 }})
	}
	for i, tc := range tests {
		if tc.f == nil {
			t.Fatal("conversion failed", i)
		}
		for a, x := range leaves {
			got, err := dd.Evaluate(tc.f, []bool{a&4 != 0, a&2 != 0, a&1 != 0})
			if err != nil || got != tc.want(x) {
				t.Fatal("conversion", i, a, x, got, err)
			}
		}
	}
	if Cudd_addBddIthBit(dd, f, -1) != nil || Cudd_addBddIthBit(dd, f, 32) != nil {
		t.Fatal("invalid bit accepted")
	}
	if Cudd_addBddThreshold(dd, dd.LogicZero, 0) != nil {
		t.Fatal("BDD false accepted as ADD")
	}
}

// This checksum comes from test_vectors/bridge_oracle.c, linked to original
// CUDD. NaN comparisons intentionally follow the original negated predicates.
func TestOriginalCBridgeFixture(t *testing.T) {
	leaves := []float64{-2147483648, -17.75, -1.25, 0, .75, 2.5, 17, 2147483647}
	dd := NewBDDManager(3)
	var build func(int, int, int) *DdNode
	build = func(level, start, end int) *DdNode {
		if end-start == 1 {
			return Cudd_addConst(dd, leaves[start])
		}
		middle := (start + end) / 2
		return Cudd_addIte(dd, Cudd_addIthVar(dd, level), build(level+1, middle, end), build(level+1, start, middle))
	}
	f := build(0, 0, 8)
	var hash uint64 = 14695981039346656037
	collect := func(f *DdNode) {
		if f == nil {
			t.Fatal("conversion failed")
		}
		for a := 0; a < 8; a++ {
			value, err := dd.Evaluate(f, []bool{a&4 != 0, a&2 != 0, a&1 != 0})
			if err != nil {
				t.Fatal(err)
			}
			if value {
				hash ^= 1
			}
			hash *= 1099511628211
		}
	}
	for _, v := range []float64{-20, -1, 0, .75, 17, 100, math.NaN(), math.Inf(-1), math.Inf(1)} {
		collect(Cudd_addBddThreshold(dd, f, v))
		collect(Cudd_addBddStrictThreshold(dd, f, v))
	}
	for _, v := range [][2]float64{{-20, 20}, {-1, 0}, {0, 0}, {2, -2}, {math.NaN(), 0}, {0, math.NaN()}} {
		collect(Cudd_addBddInterval(dd, f, v[0], v[1]))
	}
	for bit := 0; bit < 32; bit++ {
		collect(Cudd_addBddIthBit(dd, f, bit))
	}
	if hash != 0x2c12e60c54a5bb25 {
		t.Fatalf("original C bridge checksum: %016x", hash)
	}
}
func TestBridgeExceptionalLeavesAndOwnership(t *testing.T) {
	dd := NewBDDManager(1)
	for _, v := range []float64{math.NaN(), math.Inf(-1), math.Inf(1), 2147483648, -2147483649} {
		f := Cudd_addConst(dd, v)
		if Cudd_addBddIthBit(dd, f, 0) != nil {
			t.Fatal("undefined integer conversion accepted", v)
		}
		if Cudd_addBddThreshold(dd, f, 0) == nil || Cudd_addBddStrictThreshold(dd, f, 0) == nil || Cudd_addBddInterval(dd, f, -1, 1) == nil {
			t.Fatal("valid floating conversion rejected", v)
		}
	}
	nan := Cudd_addConst(dd, math.NaN())
	if Cudd_addBddThreshold(dd, nan, 0) != dd.One || Cudd_addBddStrictThreshold(dd, nan, 0) != dd.One || Cudd_addBddInterval(dd, nan, -1, 1) != dd.One {
		t.Fatal("lost original NaN comparison semantics")
	}
	other := NewBDDManager(1)
	if Cudd_addBddThreshold(dd, Cudd_addIthVar(other, 0), 0) != nil {
		t.Fatal("accepted foreign diagram")
	}
}

func TestNumericBridgeAfterReordering(t *testing.T) {
	dd := NewBDDManager(3)
	f := Cudd_addIte(dd, Cudd_addIthVar(dd, 0), Cudd_addIte(dd, Cudd_addIthVar(dd, 2), Cudd_addConst(dd, 9.5), Cudd_addConst(dd, -3)), Cudd_addIthVar(dd, 1))
	Cudd_Ref(f)
	if Cudd_ShuffleHeap(dd, []int{2, 0, 1}) == 0 {
		t.Fatal("shuffle failed")
	}
	b := Cudd_addBddThreshold(dd, f, 1)
	for a := 0; a < 8; a++ {
		values := []bool{a&1 != 0, a&2 != 0, a&4 != 0}
		number, err := dd.EvaluateADD(f, values)
		value, berr := dd.Evaluate(b, values)
		if err != nil || berr != nil || value != (number >= 1) {
			t.Fatal("reordered bridge", a, number, value, err, berr)
		}
	}
}
