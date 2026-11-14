package cudd

import (
	"testing"

	core "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCuddAddWalshSliceOrderAndValidation(t *testing.T) {
	dd := core.NewBDDManager(2)
	if Cudd_addWalsh(dd, nil, nil, 0) != dd.One {
		t.Fatal("zero-bit Walsh result should be ADD one")
	}
	if Cudd_addWalsh(dd, []*DdNode{dd.Vars[0]}, []*DdNode{dd.Vars[1]}, 2) != nil {
		t.Fatal("short variable tables were accepted")
	}
	x, y := core.Cudd_addIthVar(dd, 0), core.Cudd_addIthVar(dd, 1)
	f := Cudd_addWalsh(dd, []*DdNode{x}, []*DdNode{y}, 1)
	if f == nil {
		t.Fatal("Walsh construction failed")
	}
	for _, tc := range []struct {
		assignment []bool
		want       float64
	}{{[]bool{false, false}, 1}, {[]bool{false, true}, 1}, {[]bool{true, false}, 1}, {[]bool{true, true}, -1}} {
		got, err := dd.EvaluateADD(f, tc.assignment)
		if err != nil || got != tc.want {
			t.Errorf("Walsh(%v) = %g, %v; want %g", tc.assignment, got, err, tc.want)
		}
	}
}
