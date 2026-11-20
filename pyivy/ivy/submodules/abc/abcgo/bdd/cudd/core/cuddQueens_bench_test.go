package core

import "testing"

// Construct the row-major queens BDD in the same operation order as the C driver.
func queensBDD(side int) (solutions float64, nodes int) {
	vars := side * side
	dd := Cudd_Init(uint32(vars), 0, 256, 16384, 0)
	f := Cudd_ReadOne(dd)
	Cudd_Ref(f)
	for r := 0; r < side; r++ {
		row := Cudd_ReadLogicZero(dd)
		Cudd_Ref(row)
		for c := 0; c < side; c++ {
			x := Cudd_bddOr(dd, row, Cudd_bddIthVar(dd, r*side+c))
			Cudd_Ref(x)
			Cudd_RecursiveDeref(dd, row)
			row = x
		}
		x := Cudd_bddAnd(dd, f, row)
		Cudd_Ref(x)
		Cudd_RecursiveDeref(dd, row)
		Cudd_RecursiveDeref(dd, f)
		f = x
	}
	for i := 0; i < vars; i++ {
		for j := i + 1; j < vars; j++ {
			dr, dc := j/side-i/side, j%side-i%side
			if dc < 0 {
				dc = -dc
			}
			if dr != 0 && dc != 0 && dr != dc {
				continue
			}
			pair := Cudd_bddAnd(dd, Cudd_bddIthVar(dd, i), Cudd_bddIthVar(dd, j))
			Cudd_Ref(pair)
			x := Cudd_bddAnd(dd, f, Cudd_Not(pair))
			Cudd_Ref(x)
			Cudd_RecursiveDeref(dd, pair)
			Cudd_RecursiveDeref(dd, f)
			f = x
		}
	}
	solutions, nodes = Cudd_CountMinterm(dd, f, vars), Cudd_DagSize(f)
	Cudd_RecursiveDeref(dd, f)
	Cudd_Quit(dd)
	return
}

func TestQueensBDDMatchesNative(t *testing.T) {
	for _, tc := range []struct{ side, solutions, nodes int }{{5, 10, 167}, {6, 4, 130}} {
		got, nodes := queensBDD(tc.side)
		if got != float64(tc.solutions) || nodes != tc.nodes {
			t.Errorf("%d queens got %g solutions and %d nodes, want %d and %d", tc.side, got, nodes, tc.solutions, tc.nodes)
		}
	}
}

func BenchmarkSixQueensBDD(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if solutions, nodes := queensBDD(6); solutions != 4 || nodes != 130 {
			b.Fatalf("solutions %g, nodes %d", solutions, nodes)
		}
	}
}
