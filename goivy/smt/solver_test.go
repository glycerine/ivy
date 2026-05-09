//go:build cgo && !tinygo && !wasip1 && !web

package smt

import (
	"strings"
	"testing"
)

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

func TestCanonPreservesEnumQuantifierAndFalse(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()

	opSort, opVals := ctx.EnumSort("op_type", []string{"nop", "write", "read"})
	lclockSort := ctx.UninterpretedSort("lclock")
	req := ctx.Function("ref.evs.req", []Z3Sort{lclockSort}, opSort)
	tick := ctx.Const("T", lclockSort)

	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.ForAll([]Z3Expr{tick}, ctx.Eq(req.Apply(tick), opVals[2])))
	solver.Assert(ctx.BoolVal(false))

	canon := solver.CanonZ3Assertions()
	if strings.Contains(canon, "unknown_kind=") {
		t.Fatalf("canon contains unknown AST kind: %s", canon)
	}
	if !strings.Contains(canon, "(a = (a ref.evs.req (v T)) (c read op_type))") {
		t.Fatalf("canon missing enum equality: %s", canon)
	}
	if !strings.Contains(canon, "(c false Bool)") {
		t.Fatalf("canon missing false constant: %s", canon)
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
