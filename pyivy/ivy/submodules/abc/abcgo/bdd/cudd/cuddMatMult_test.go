package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func evalADD(t *testing.T, dd *DdManager, f *DdNode, inputs [3]int32) float64 {
	t.Helper()
	n := Cudd_Eval(dd, f, &inputs[0])
	if n == nil {
		t.Fatal("ADD evaluation returned nil")
	}
	return core.Cudd_V(n)
}

func TestCuddADDMatrixMultiplyAndTriangle(t *testing.T) {
	dd := NewBDDManager(3)
	x, z, y := Cudd_addIthVar(dd, 0), Cudd_addIthVar(dd, 1), Cudd_addIthVar(dd, 2)
	c := func(v float64) *DdNode { return Cudd_addConst(dd, v) }
	choose := func(v, hi, lo *DdNode) *DdNode { return Cudd_addIte(dd, v, hi, lo) }
	a := choose(x, choose(z, c(4), c(3)), choose(z, c(2), c(1)))
	b := choose(z, choose(y, c(8), c(7)), choose(y, c(6), c(5)))

	product := Cudd_addMatrixMultiply(dd, a, b, []*DdNode{z}, 1)
	if product == nil {
		t.Fatal("matrix multiply returned nil")
	}
	for _, tc := range []struct {
		x, y int32
		want float64
	}{{0, 0, 19}, {0, 1, 22}, {1, 0, 43}, {1, 1, 50}} {
		got := evalADD(t, dd, product, [3]int32{tc.x, 0, tc.y})
		if got != tc.want {
			t.Errorf("product[%d,%d] = %g, want %g", tc.x, tc.y, got, tc.want)
		}
	}

	triangle := Cudd_addTriangle(dd, a, b, []*DdNode{z}, 1)
	if triangle == nil {
		t.Fatal("triangle product returned nil")
	}
	for _, tc := range []struct {
		x, y int32
		want float64
	}{{0, 0, 6}, {0, 1, 7}, {1, 0, 8}, {1, 1, 9}} {
		got := evalADD(t, dd, triangle, [3]int32{tc.x, 0, tc.y})
		if got != tc.want {
			t.Errorf("triangle[%d,%d] = %g, want %g", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestCuddADDOuterSum(t *testing.T) {
	dd := NewBDDManager(3)
	x, y := Cudd_addIthVar(dd, 0), Cudd_addIthVar(dd, 1)
	c := func(v float64) *DdNode { return Cudd_addConst(dd, v) }
	choose := func(v, hi, lo *DdNode) *DdNode { return Cudd_addIte(dd, v, hi, lo) }
	r := choose(x, c(3), c(1))
	col := choose(y, c(4), c(2))
	m := c(10)
	got := Cudd_addOuterSum(dd, m, r, col)
	if got == nil {
		t.Fatal("outer sum returned nil")
	}
	for _, tc := range []struct {
		x, y int32
		want float64
	}{{0, 0, 3}, {0, 1, 5}, {1, 0, 5}, {1, 1, 7}} {
		value := evalADD(t, dd, got, [3]int32{tc.x, tc.y, 0})
		if value != tc.want {
			t.Errorf("outer sum[%d,%d] = %g, want %g", tc.x, tc.y, value, tc.want)
		}
	}
}
