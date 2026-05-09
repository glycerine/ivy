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

func TestCanonPreservesEnumFunctionCopy(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()

	opSort, opVals := ctx.EnumSort("op_type", []string{"nop", "write", "read", "rrsp", "wr_cmp"})
	lclockSort := ctx.UninterpretedSort("lclock")
	req := ctx.Function("ref.evs.req", []Z3Sort{lclockSort}, opSort)
	oldReq := ctx.Function("__ref.evs.req", []Z3Sort{lclockSort}, opSort)
	tick := ctx.Const("V0:lclock", lclockSort)

	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.ForAll([]Z3Expr{tick}, ctx.Eq(req.Apply(tick), oldReq.Apply(tick))))
	solver.Assert(ctx.ForAll([]Z3Expr{tick}, ctx.Eq(req.Apply(tick), ctx.Ite(ctx.BoolVal(true), oldReq.Apply(tick), opVals[0]))))

	canon := solver.CanonZ3Assertions()
	if strings.Contains(canon, "unknown_kind=") {
		t.Fatalf("canon contains unknown AST kind: %s", canon)
	}
	if !strings.Contains(canon, "(a = (a ref.evs.req (v V0:lclock)) (a __ref.evs.req (v V0:lclock)))") {
		t.Fatalf("canon missing enum function copy: %s", canon)
	}
	if !strings.Contains(canon, "(a if (c true Bool) (a __ref.evs.req (v V0:lclock)) (c nop op_type))") {
		t.Fatalf("canon missing enum function ite: %s", canon)
	}
}

func TestCanonPreservesMultiBoundEnumPremise(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()

	opSort, opVals := ctx.EnumSort("op_type", []string{"nop", "write", "read", "rrsp", "wr_cmp"})
	locSort, locVals := ctx.EnumSort("loc_type", []string{"init_l", "cf_mem_l", "cf_pio_l", "cf_cmp_l", "if_l", "memc_l", "dramc_l", "arm_l"})
	lclockSort := ctx.UninterpretedSort("lclock")
	tarCFClockSort := ctx.UninterpretedSort("tar_cf_clock")

	req := ctx.Function("__ref.evs.req", []Z3Sort{lclockSort}, opSort)
	lReq := ctx.Function("__ref.evs.l_req", []Z3Sort{lclockSort}, locSort)
	arrMin := ctx.Const("cfabric.t_rd_arr_min", tarCFClockSort)
	tick := ctx.Const("T:lclock", lclockSort)
	tarr := ctx.Const("Tarr:tar_cf_clock", tarCFClockSort)

	body := ctx.Implies(
		ctx.And(
			ctx.Eq(arrMin, tarr),
			ctx.Eq(req.Apply(tick), opVals[2]),
			ctx.Eq(lReq.Apply(tick), locVals[1]),
		),
		ctx.Eq(arrMin, tarr),
	)

	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.ForAll([]Z3Expr{tick, tarr}, body))

	canon := solver.CanonZ3Assertions()
	if strings.Contains(canon, "unknown_kind=") {
		t.Fatalf("canon contains unknown AST kind: %s", canon)
	}
	if !strings.Contains(canon, "(a = (a __ref.evs.req (v T:lclock)) (c read op_type))") {
		t.Fatalf("canon missing req premise: %s", canon)
	}
	if !strings.Contains(canon, "(a = (a __ref.evs.l_req (v T:lclock)) (c cf_mem_l loc_type))") {
		t.Fatalf("canon missing l_req premise: %s", canon)
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
