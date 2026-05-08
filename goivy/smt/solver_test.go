//go:build cgo && !tinygo && !wasip1 && !web

package smt

import "testing"

func TestSolverBoolRoundTrip(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	satSolver := ctx.NewSolver()
	satSolver.Assert(ctx.BoolVal(true))
	if got := satSolver.Check(); got != Sat {
		t.Fatalf("true check = %v, want %v", got, Sat)
	}

	unsatSolver := ctx.NewSolver()
	truth := ctx.BoolVal(true)
	unsatSolver.Assert(truth)
	unsatSolver.Assert(ctx.Not(truth))
	if got := unsatSolver.Check(); got != Unsat {
		t.Fatalf("true and not true check = %v, want %v", got, Unsat)
	}
}
