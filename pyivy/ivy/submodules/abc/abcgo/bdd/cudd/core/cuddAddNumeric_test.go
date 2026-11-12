package core

import (
	"math"
	"testing"
)

func numericFromLeaves(dd *DdManager, values []float64) *DdNode {
	var build func(int, int, int) *DdNode
	build = func(level, start, end int) *DdNode {
		if end-start == 1 {
			return Cudd_addConst(dd, values[start])
		}
		middle := (start + end) / 2
		return Cudd_addIte(dd, Cudd_addIthVar(dd, level), build(level+1, middle, end), build(level+1, start, middle))
	}
	return build(0, 0, len(values))
}
func numericCube(dd *DdManager, mask, n int) *DdNode {
	cube := dd.One
	for i := 0; i < n; i++ {
		if mask&(1<<i) != 0 {
			cube = Cudd_addApply(dd, Cudd_addTimes, cube, Cudd_addIthVar(dd, i))
		}
	}
	return cube
}
func numericValues(assignment, n int) []bool {
	values := make([]bool, n)
	for i := range values {
		values[i] = assignment&(1<<uint(n-1-i)) != 0
	}
	return values
}
func TestADDAbstractionEnumeratedSumsProductsAndOr(t *testing.T) {
	seed := uint64(729)
	for sample := 0; sample < 32; sample++ {
		dd := NewBDDManager(4)
		leaves := make([]float64, 8)
		for i := range leaves {
			seed ^= seed << 13
			seed ^= seed >> 7
			seed ^= seed << 17
			leaves[i] = float64(int(seed%7) - 3)
		}
		f := numericFromLeaves(dd, leaves)
		boolean := Cudd_BddToAdd(dd, Cudd_addBddPattern(dd, f))
		for mask := 0; mask < 16; mask++ {
			cube := numericCube(dd, mask, 4)
			results := []*DdNode{Cudd_addExistAbstract(dd, f, cube), Cudd_addUnivAbstract(dd, f, cube), Cudd_addOrAbstract(dd, boolean, cube)}
			for assignment := 0; assignment < 16; assignment++ {
				sum, product, or := 0., 1., 0.
				for quantified := 0; quantified < 16; quantified++ {
					match := true
					for v := 0; v < 4; v++ {
						if mask&(1<<v) == 0 && (assignment&(1<<uint(3-v))) != (quantified&(1<<uint(3-v))) {
							match = false
						}
					}
					if !match {
						continue
					}
					value := leaves[quantified>>1]
					sum += value
					product *= value
					if value != 0 {
						or = 1
					}
				}
				for mode, want := range []float64{sum, product, or} {
					got, err := dd.EvaluateADD(results[mode], numericValues(assignment, 4))
					if err != nil || got != want {
						t.Fatal("abstract", sample, mask, assignment, mode, got, want, err)
					}
				}
			}
		}
	}
}
func TestADDAbstractionInvalidCubes(t *testing.T) {
	dd := NewBDDManager(2)
	f := Cudd_addIthVar(dd, 0)
	other := NewBDDManager(2)
	bad := []*DdNode{dd.Zero, dd.LogicZero, Cudd_bddIthVar(dd, 0), Cudd_addConst(dd, 2), Cudd_addIte(dd, Cudd_addIthVar(dd, 0), dd.Zero, dd.One), Cudd_addIthVar(other, 0)}
	for _, cube := range bad {
		if Cudd_addExistAbstract(dd, f, cube) != nil || Cudd_addUnivAbstract(dd, f, cube) != nil || Cudd_addOrAbstract(dd, f, cube) != nil {
			t.Fatal("accepted invalid cube", cube)
		}
	}
}
func TestADDScalarInverseAndExtrema(t *testing.T) {
	dd := NewBDDManager(2)
	f := numericFromLeaves(dd, []float64{-4, -.5, 2, 8})
	if Cudd_addFindMax(dd, f).Value != 8 || Cudd_addFindMin(dd, f).Value != -4 {
		t.Fatal("extrema")
	}
	r := Cudd_addScalarInverse(dd, f, Cudd_addConst(dd, .5))
	if r == nil {
		t.Fatal("equal epsilon rejected")
	}
	for a, want := range []float64{-.25, -2, .5, .125} {
		value, err := dd.EvaluateADD(r, numericValues(a, 2))
		if err != nil || value != want {
			t.Fatal("reciprocal", a, value, err)
		}
	}
	if Cudd_addScalarInverse(dd, f, Cudd_addConst(dd, .50001)) != nil || Cudd_addScalarInverse(dd, f, Cudd_addIthVar(dd, 0)) != nil {
		t.Fatal("invalid inverse accepted")
	}
	if Cudd_addScalarInverse(dd, dd.Zero, dd.Zero) != dd.PlusInfinity {
		t.Fatal("zero epsilon semantics")
	}
	for bit := 0; bit < 32; bit++ {
		add := Cudd_addIthBit(dd, f, bit)
		bdd := Cudd_addBddIthBit(dd, f, bit)
		if Cudd_addBddPattern(dd, add) != bdd {
			t.Fatal("bit conversion disagrees", bit)
		}
	}
}
func TestADDExtremaNaNAndInfinitySourceOrder(t *testing.T) {
	dd := NewBDDManager(1)
	v := Cudd_addIthVar(dd, 0)
	nan := Cudd_addConst(dd, math.NaN())
	two := Cudd_addConst(dd, 2)
	f := Cudd_addIte(dd, v, nan, two)
	if Cudd_addFindMax(dd, f) != two || Cudd_addFindMin(dd, f) != two {
		t.Fatal("then NaN comparison")
	}
	f = Cudd_addIte(dd, v, two, nan)
	if !math.IsNaN(Cudd_addFindMax(dd, f).Value) || !math.IsNaN(Cudd_addFindMin(dd, f).Value) {
		t.Fatal("else NaN comparison")
	}
	f = Cudd_addIte(dd, v, dd.PlusInfinity, nan)
	if Cudd_addFindMax(dd, f) != dd.PlusInfinity {
		t.Fatal("max infinity short circuit")
	}
	f = Cudd_addIte(dd, v, dd.MinusInfinity, nan)
	if Cudd_addFindMin(dd, f) != dd.MinusInfinity {
		t.Fatal("min infinity short circuit")
	}
}
func TestADDNumericAfterShuffle(t *testing.T) {
	dd := NewBDDManager(3)
	f := numericFromLeaves(dd, []float64{1, 2, 3, 4, 5, 6, 7, 8})
	cube := numericCube(dd, 5, 3)
	Cudd_Ref(f)
	Cudd_Ref(cube)
	sum := Cudd_addExistAbstract(dd, f, cube)
	Cudd_Ref(sum)
	if Cudd_ShuffleHeap(dd, []int{2, 0, 1}) == 0 {
		t.Fatal("shuffle")
	}
	if Cudd_addExistAbstract(dd, f, cube) != sum || Cudd_addFindMax(dd, f).Value != 8 {
		t.Fatal("reordered abstraction")
	}
}

func TestOriginalCADDNumericFixture(t *testing.T) {
	dd := NewBDDManager(4)
	f := numericFromLeaves(dd, []float64{-3, -2, -1, 0, 1, 2, 3, 4})
	boolean := Cudd_BddToAdd(dd, Cudd_addBddPattern(dd, f))
	var hash uint64 = 14695981039346656037
	byteHash := func(b byte) { hash ^= uint64(b); hash *= 1099511628211 }
	count:=0
 collect := func(f *DdNode) {
 defer func(){t.Logf("%d %016x",count,hash);count++}()
		if f == nil {
			byteHash(255)
			return
		}
		byteHash(0)
		for a := 0; a < 16; a++ {
			value, err := dd.EvaluateADD(f, numericValues(a, 4))
			if err != nil {
				t.Fatal(err)
			}
			bits := math.Float64bits(value)
			for i := 0; i < 8; i++ {
				byteHash(byte(bits >> uint(i*8)))
			}
		}
	}
	for mask := 0; mask < 16; mask++ {
		cube := numericCube(dd, mask, 4)
		collect(Cudd_addExistAbstract(dd, f, cube))
		collect(Cudd_addUnivAbstract(dd, f, cube))
		collect(Cudd_addOrAbstract(dd, boolean, cube))
	}
	collect(Cudd_addFindMax(dd, f))
	collect(Cudd_addFindMin(dd, f))
	for _, epsilon := range []float64{0, .5, 1, 4} {
		collect(Cudd_addScalarInverse(dd, f, Cudd_addConst(dd, epsilon)))
	}
	for bit := 0; bit < 5; bit++ {
		collect(Cudd_addIthBit(dd, f, bit))
	}
	collect(Cudd_addIthBit(dd, f, 31))
	if hash != 0x9b9f115f3cf36db6 {
		t.Fatalf("original C numeric fixture: %016x", hash)
	}
}
