package core

import (
	"strings"
	"testing"
)

func TestBddReadSparseGraph(t *testing.T) {
	dd := NewBDDManager(0)
	var graph *DdNode
	var x, y []*DdNode
	var nx, ny, rows, cols int32
	if Cudd_bddRead(strings.NewReader("3 3\n0 2\n2 1\n"), dd, &graph, &x, &y, &nx, &ny, &rows, &cols, 0, 2, 1, 2) != 1 {
		t.Fatal("sparse graph read failed")
	}
	if nx != 2 || ny != 2 || rows != 3 || cols != 3 {
		t.Fatalf("dimensions %d %d, bits %d %d", rows, cols, nx, ny)
	}
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			assignment := []bool{row&2 != 0, col&2 != 0, row&1 != 0, col&1 != 0}
			got, err := dd.Evaluate(graph, assignment)
			want := row == 0 && col == 2 || row == 2 && col == 1
			if err != nil || got != want {
				t.Fatalf("row %d col %d: got %v want %v, %v", row, col, got, want, err)
			}
		}
	}
}

func TestAddReadSparseMatrixAndRepeatedEntry(t *testing.T) {
	dd := NewBDDManager(0)
	var matrix *DdNode
	var x, y, xn, yn []*DdNode
	var nx, ny, rows, cols int32
	if Cudd_addRead(strings.NewReader("2 3\n0 1 1.5\n1 2 -4\n0 1 2.5\n"), dd, &matrix, &x, &y, &xn, &yn, &nx, &ny, &rows, &cols, 0, 1, 1, 1) != 1 {
		t.Fatal("sparse matrix read failed")
	}
	for row := 0; row < 2; row++ {
		for col := 0; col < 4; col++ {
			assignment := []bool{row != 0, col&2 != 0, col&1 != 0}
			got, err := dd.EvaluateADD(matrix, assignment)
			want := 0.0
			if row == 0 && col == 1 {
				want = 2.5
			}
			if row == 1 && col == 2 {
				want = -4
			}
			if err != nil || got != want {
				t.Fatalf("row %d col %d: got %v want %v, %v", row, col, got, want, err)
			}
		}
	}
}

func TestReadRejectsMalformedAndOutOfRangeEntries(t *testing.T) {
	for _, input := range []string{"", "2", "2 2\n0", "2 2\n-1 0", "2 2\n0 2"} {
		dd := NewBDDManager(0)
		var result *DdNode
		var x, y []*DdNode
		var nx, ny, rows, cols int32
		if Cudd_bddRead(strings.NewReader(input), dd, &result, &x, &y, &nx, &ny, &rows, &cols, 0, 1, 1, 1) != 0 {
			t.Fatalf("accepted %q", input)
		}
	}
}
