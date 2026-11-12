package gia

import "testing"

func TestQbfCounterexampleGuidedSolver(t *testing.T) {
	for _, unsat := range []bool{false, true} {
		g := Gia_ManStart(10)
		p, x := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		Gia_ManHashAlloc(g)
		var f int32
		if unsat {
			f = Gia_ManHashXor(g, p, x) ^ 1
		} else {
			f = Gia_ManHashOr(g, p, x)
		}
		Gia_ManAppendCo(g, f)
		status, values := qbfSolve(g, 1, 0, 0, 0)
		if unsat && status != 1 || !unsat && (status != 0 || len(values) != 1 || !values[0]) {
			t.Fatalf("unsat=%t status=%d values=%v", unsat, status, values)
		}
	}
}

func TestQbfSolverIterationLimit(t *testing.T) {
	g := Gia_ManStart(8)
	p, x := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashAlloc(g)
	Gia_ManAppendCo(g, Gia_ManHashOr(g, p, x))
	if status, _ := qbfSolve(g, 1, 1, 0, 0); status != -1 {
		t.Fatalf("expected iteration limit, got %d", status)
	}
}
