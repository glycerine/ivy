package gia

import (
	"fmt"
	"testing"
)

func TestGiarfSolverAssumptions(t *testing.T) {
	for _, mini := range []bool{false, true} {
		g := Gia_ManStart(10)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		Gia_ManCreateRefs(g)
		Gia_ManFillValue(g)
		s := newGiarfSolver(g, 1000, mini)
		status, model, err := s.check(x, nil)
		if err != nil || status != 0 || model == nil {
			t.Fatalf("mini=%v sat=%d err=%v", mini, status, err)
		}
		neg := x ^ 1
		status, _, err = s.check(x, &neg)
		if err != nil || status != 1 {
			t.Fatalf("mini=%v unsat=%d err=%v", mini, status, err)
		}
		status, _, err = s.check(0, nil)
		if err != nil || status != 1 {
			t.Fatalf("mini=%v constant=%d err=%v", mini, status, err)
		}
		s.close()
	}
}

func TestGiarfSolverRepeatedAssumptionsAndWitnesses(t *testing.T) {
	for _, mini := range []bool{false, true} {
		t.Run(fmt.Sprintf("cec%t", mini), func(t *testing.T) {
			g := Gia_ManStart(1)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			Gia_ManHashAlloc(g)
			x := Gia_ManHashAnd(g, a, b^1)
			y := Gia_ManHashXor(g, b, c)
			z := Gia_ManHashMux(g, a, x, y)
			Gia_ManHashStop(g)
			gates := []int32{0, 1, a, a ^ 1, b, b ^ 1, c, c ^ 1, x, x ^ 1, y, y ^ 1, z, z ^ 1}
			truth := make([][]bool, 8)
			for mask := range truth {
				var err error
				truth[mask], _, err = g.evaluateObjects([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
				if err != nil {
					t.Fatal(err)
				}
			}
			Gia_ManCreateRefs(g)
			Gia_ManFillValue(g)
			s := newGiarfSolver(g, 1000, mini)
			defer s.close()
			check := func(a int32, b *int32) {
				t.Helper()
				value := func(row []bool, lit int32) bool { return row[lit>>1] != (lit&1 != 0) }
				want := int32(1)
				for _, row := range truth {
					if value(row, a) && (b == nil || value(row, *b)) {
						want = 0
					}
				}
				status, model, err := s.check(a, b)
				if err != nil || status != want {
					t.Fatalf("query %d/%v: status %d, %v; want %d", a, b, status, err, want)
				}
				if status == 1 {
					if model != nil {
						t.Fatal("UNSAT query retained the prior witness")
					}
					return
				}
				if model == nil {
					t.Fatal("SAT query omitted its witness")
				}
				mask := 0
				for _, lit := range Vec_IntArray(model) {
					if lit < 0 || lit>>1 >= 3 {
						t.Fatalf("witness literal %d is not a CI ordinal", lit)
					}
					if lit&1 == 0 {
						mask |= 1 << (lit >> 1)
					}
				}
				if !value(truth[mask], a) || b != nil && !value(truth[mask], *b) {
					t.Fatalf("witness %v does not satisfy query", Vec_IntArray(model))
				}
			}
			for round := 0; round < 2; round++ {
				for i, a := range gates {
					check(a, nil)
					for j := range gates {
						b := gates[(i+j+round)%len(gates)]
						check(a, &b)
					}
				}
			}
		})
	}
}

func BenchmarkGiarfRepeatedCECQueries(b *testing.B) {
	g := Gia_ManStart(1)
	root := Gia_ManAppendCi(g)
	for i := 0; i < 63; i++ {
		root = Gia_ManAppendAnd(g, root, Gia_ManAppendCi(g))
	}
	Gia_ManCreateRefs(g)
	Gia_ManFillValue(g)
	s := newGiarfSolver(g, 1000, true)
	defer s.close()
	// Warm the cone once so the measurement covers repeated queries.
	if _, _, err := s.check(root, nil); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if status, _, err := s.check(root, nil); err != nil || status != 0 {
			b.Fatalf("status %d: %v", status, err)
		}
	}
}
