//go:build !js && !web

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

func TestCanonZ3AssertionsIncludesQuantifierWeight(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()

	x := ctx.Const("X", ctx.IntSort())
	q := ctx.ForAll([]Z3Expr{x}, ctx.Le(ctx.IntVal(0), x))
	solver := ctx.NewZ3Solver()
	solver.Assert(q)

	got := solver.CanonZ3Assertions()
	if !strings.Contains(got, "(forall weight:1 ") {
		t.Fatalf("canonical assertions missing quantifier weight: %s", got)
	}
}

func TestSolverToSMT2UsesBenchmarkFormat(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()

	x := ctx.Const("x", ctx.IntSort())
	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.Lt(ctx.IntVal(0), x))
	solver.Assert(ctx.Lt(x, ctx.IntVal(2)))

	got := solver.ToSMT2()
	for _, want := range []string{
		"; benchmark generated from python API",
		"(set-info :status unknown)",
		"(assert",
		"(check-sat)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("ToSMT2() missing %q:\n%s", want, got)
		}
	}
}

func TestSortStringUsesHistoricalSortName(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()
	for _, tt := range []struct {
		sort Z3Sort
		want string
	}{
		{ctx.BoolSort(), "Bool"},
		{ctx.IntSort(), "Int"},
		{ctx.BvSort(8), "bv"},
		{ctx.ArraySort(ctx.IntSort(), ctx.BoolSort()), "Array"},
		{ctx.UninterpretedSort("a b"), "a b"},
	} {
		if got := tt.sort.String(); got != tt.want {
			t.Errorf("sort name = %q, want %q", got, tt.want)
		}
	}
}

func TestBooleanSingletonsPreserveApplicationShape(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()
	p := ctx.BoolConst("p")
	s := ctx.NewZ3Solver()
	s.Assert(ctx.And(p))
	s.Assert(ctx.Or(p))
	want := "(asserts (a and (c p Bool)) (a or (c p Bool)))"
	if got := s.CanonZ3Assertions(); got != want {
		t.Fatalf("singleton connective shape = %s, want %s", got, want)
	}
}
