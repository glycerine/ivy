package core

import (
	"slices"
	"testing"
)

// Construct the row-major queens BDD in the same operation order as the C driver.
func buildQueensBDD(side int) (*DdManager, *DdNode) {
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
	return dd, f
}

func queensBDD(side int) (solutions float64, nodes int) {
	dd, f := buildQueensBDD(side)
	solutions, nodes = Cudd_CountMinterm(dd, f, side*side), Cudd_DagSize(f)
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

func BenchmarkQueensCheckZeroRef(b *testing.B) {
	dd, root := buildQueensBDD(6)
	Cudd_RecursiveDeref(dd, root)
	defer Cudd_Quit(dd)
	b.ReportAllocs()
	for b.Loop() {
		if leaks := Cudd_CheckZeroRef(dd); leaks != 0 {
			b.Fatalf("%d remaining external roots", leaks)
		}
	}
}

func BenchmarkQueensAPACount(b *testing.B) {
	for _, tc := range []struct {
		name                   string
		side, width, solutions int
	}{
		{"FiveQueens", 5, 4121, 10},
		{"SixQueensCrossDigit", 6, 4096, 4},
	} {
		b.Run(tc.name, func(b *testing.B) {
			dd, root := buildQueensBDD(tc.side)
			defer Cudd_Quit(dd)
			defer Cudd_RecursiveDeref(dd, root)
			want := make(DdApaNumber, Cudd_ApaNumberOfDigits(tc.width+1))
			for n, power := tc.solutions, tc.width-tc.side*tc.side; n > 0; n, power = n>>1, power+1 {
				if n&1 != 0 {
					want[len(want)-1-power/32] |= 1 << (power % 32)
				}
			}
			b.ReportAllocs()
			for b.Loop() {
				var digits int
				got := Cudd_ApaCountMinterm(dd, root, tc.width, &digits)
				if digits != len(want) || !slices.Equal(got, want) {
					b.Fatal("incorrect exact count")
				}
			}
		})
	}
}
