package gia

import "testing"

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
