package core

import (
	"strings"
	"testing"
)

// Cudd_bddRead preserves the caller's existing variable pointers and appends
// only the additional indices required by the input dimensions.
func TestBddReadExtendsExistingVariableArrays(t *testing.T) {
	dd := NewBDDManager(0)
	first := Cudd_bddIthVar(dd, 0)
	x, y := []*DdNode{first}, []*DdNode(nil)
	var result *DdNode
	nx, ny, rows, cols := int32(1), int32(0), int32(0), int32(0)
	if Cudd_bddRead(strings.NewReader("4 2\n3 1\n"), dd, &result, &x, &y, &nx, &ny, &rows, &cols, 0, 2, 1, 2) != 1 {
		t.Fatal("read failed")
	}
	if nx != 2 || ny != 1 || x[0] != first || x[1].Index != 2 || y[0].Index != 1 {
		t.Fatalf("existing variables were not extended in source index order: nx=%d ny=%d", nx, ny)
	}
	for mask := 0; mask < 8; mask++ {
		value, err := dd.Evaluate(result, []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
		if err != nil || value != (mask == 7) {
			t.Fatalf("mask %d: got %v, error %v", mask, value, err)
		}
	}
}

func TestAddReadExtendsExistingVariableArrays(t *testing.T) {
	dd := NewBDDManager(0)
	first := Cudd_addIthVar(dd, 0)
	firstNot := Cudd_addIte(dd, first, dd.Zero, dd.One)
	x, xn := []*DdNode{first}, []*DdNode{firstNot}
	var y, yn []*DdNode
	var result *DdNode
	nx, ny, rows, cols := int32(1), int32(0), int32(0), int32(0)
	if Cudd_addRead(strings.NewReader("4 2\n3 1 7.5\n"), dd, &result, &x, &y, &xn, &yn, &nx, &ny, &rows, &cols, 0, 2, 1, 2) != 1 {
		t.Fatal("read failed")
	}
	if nx != 2 || ny != 1 || x[0] != first || xn[0] != firstNot || x[1].Index != 2 || y[0].Index != 1 {
		t.Fatal("existing ADD variables were not extended in source index order")
	}
	for mask := 0; mask < 8; mask++ {
		value, err := dd.EvaluateADD(result, []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
		want := 0.0
		if mask == 7 {
			want = 7.5
		}
		if err != nil || value != want {
			t.Fatalf("mask %d: got %v want %v, error %v", mask, value, want, err)
		}
	}
}
