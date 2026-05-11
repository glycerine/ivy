//go:build cgo && !wasip1 && !web

package smt

import "testing"

func TestSolverBoolRoundTrip(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()

	satSolver := ctx.NewZ3Solver()
	satSolver.Assert(ctx.BoolVal(true))
	if got := satSolver.Check(); got != Sat {
		t.Fatalf("true check = %v, want %v", got, Sat)
	}

	unsatSolver := ctx.NewZ3Solver()
	truth := ctx.BoolVal(true)
	unsatSolver.Assert(truth)
	unsatSolver.Assert(ctx.Not(truth))
	if got := unsatSolver.Check(); got != Unsat {
		t.Fatalf("true and not true check = %v, want %v", got, Unsat)
	}

	sort := ctx.UninterpretedSort("Thing")
	if got := sort.String(); got != "Thing" {
		t.Fatalf("uninterpreted sort name = %q, want %q", got, "Thing")
	}
}

func TestZ3ErrorCallbackBoundary(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("BvSort(0) did not report a Z3 error")
		}
		err, ok := r.(*ErrMsg)
		if !ok {
			t.Fatalf("recovered %T, want *ErrMsg", r)
		}
		if err.Code == 0 {
			t.Fatalf("error code = %d, want a Z3 error", err.Code)
		}
		if err.Msg == "" {
			t.Fatal("empty Z3 error message")
		}
	}()

	_ = ctx.BvSort(0)
}
