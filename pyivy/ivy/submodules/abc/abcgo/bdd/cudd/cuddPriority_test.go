package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCuddXgtySelectsStrictlyLargerWord(t *testing.T) {
	dd := core.NewBDDManager(4)
	x := []*DdNode{core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 2)}
	y := []*DdNode{core.Cudd_bddIthVar(dd, 1), core.Cudd_bddIthVar(dd, 3)}
	got := Cudd_Xgty(dd, 2, nil, x, y)
	if got == nil {
		t.Fatal("Cudd_Xgty returned nil")
	}
	for xv := 0; xv < 4; xv++ {
		for yv := 0; yv < 4; yv++ {
			assignment := []bool{xv&2 != 0, yv&2 != 0, xv&1 != 0, yv&1 != 0}
			value, err := dd.Evaluate(got, assignment)
			if err != nil {
				t.Fatal(err)
			}
			if value != (xv > yv) {
				t.Fatalf("x=%d y=%d: got %v, want %v", xv, yv, value, xv > yv)
			}
		}
	}
}
