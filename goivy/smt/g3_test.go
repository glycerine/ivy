//go:build !js && !classic_z3

package smt

import "testing"

func TestG3SolverOwnershipAndReset(t *testing.T) {
	ctx := NewZ3Context()
	x := ctx.Const("x", ctx.IntSort())
	s := ctx.NewZ3Solver()
	s.SetParam("smt.random_seed", "1")
	s.Assert(ctx.Eq(x, ctx.IntVal(7)))
	if s.Check() != Sat {
		t.Fatal("x=7 must be satisfiable")
	}
	m := s.Model()
	if m == nil {
		t.Fatal("missing model")
	}
	s.Reset()
	s.Assert(ctx.Eq(x, ctx.IntVal(8)))
	if s.Check() != Sat {
		t.Fatal("reset solver should admit x=8")
	}
	// A model acquired before reset must keep its original interpretation.
	v, ok := m.Eval(x, true)
	if !ok || v.String() != "7" {
		t.Fatalf("retained model: %v, %v", v, ok)
	}
	m.Release()
	m.Release()
	if err := ctx.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestG3AssumptionsCoreAndQuantifierCanon(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()
	s := ctx.NewZ3Solver()
	a := ctx.BoolConst("a")
	b := ctx.BoolConst("b")
	s.Assert(ctx.Not(a))
	if s.CheckAssumptions([]Z3Expr{a, b}) != Unsat {
		t.Fatal("contradictory assumption should be unsat")
	}
	core := s.UnsatCore()
	if len(core) != 1 || !core[0].Equal(a) {
		t.Fatalf("unexpected core: %v", core)
	}
	s.Reset()
	sort := ctx.UninterpretedSort("T")
	x := ctx.Const("X", sort)
	y := ctx.Const("Y", sort)
	p := ctx.Function("p", []Z3Sort{sort, sort}, ctx.BoolSort())
	s.Assert(ctx.ForAll([]Z3Expr{y, x}, p.Apply(x, y)))
	want := "(asserts (forall weight:1 ((X T) (Y T)) (a p (v X) (v Y))))"
	if got := s.CanonZ3Assertions(); got != want {
		t.Fatalf("canonical assertions:\n%s\nwant:\n%s", got, want)
	}
}

func TestG3SMTBindingsPreserveDeclarationIdentity(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()
	sort := ctx.UninterpretedSort("Thing")
	x := ctx.Const("x", sort)
	p := ctx.Function("p", []Z3Sort{sort}, ctx.BoolSort())
	expression, err := ctx.ParseSMTLIB2String("(assert (p x))",
		[]SMTLIBSortDecl{{Name: "Thing", Sort: sort}},
		[]SMTLIBFuncDecl{{Name: "x", Decl: x.Decl()}, {Name: "p", Decl: p}})
	if err != nil {
		t.Fatal(err)
	}
	s := ctx.NewZ3Solver()
	s.Assert(expression)
	s.Assert(ctx.Not(p.Apply(x)))
	if s.Check() != Unsat {
		t.Fatal("parsed and constructed declarations must be identical")
	}
	if _, err := ctx.ParseSMTLIB2String("(assert missing)", nil, nil); err == nil {
		t.Fatal("invalid SMT must return an error")
	}
}

func TestG3VerificationConditionFallsBackToConcreteModel(t *testing.T) {
	for _, mode := range []string{"Single", "Batch"} {
		t.Run(mode, func(t *testing.T) {
			ctx := NewZ3Context()
			defer ctx.Close()
			s := ctx.NewZ3Solver()
			check := s.CheckVerificationCondition
			if mode == "Batch" {
				var release func()
				check, release = s.BeginVerificationChecks()
				defer release()
			}
			domain := ctx.BvSort(2)
			x := ctx.Const("x", domain)
			a := ctx.Const("a", domain)
			p := ctx.Function("p", []Z3Sort{domain}, ctx.BoolSort())
			s.Assert(ctx.ForAll([]Z3Expr{x}, p.Apply(x)))

			s.Push()
			s.Assert(ctx.Not(p.Apply(a)))
			if got := check(); got != Unsat {
				t.Fatal("expected refutation", got)
			}
			s.Pop()
			three := ctx.BvVal(3, 2)
			s.Assert(ctx.Eq(a, three))
			if got := check(); got != Sat {
				t.Fatal("expected counterexample", got)
			}
			m := s.Model()
			if m == nil {
				t.Fatal("missing concrete counterexample")
			}
			defer m.Release()
			got, ok := m.Eval(a, true)
			if !ok || !got.Equal(three) {
				t.Fatal("counterexample lost the two-bit value 3", got, ok)
			}
		})
	}
}

func TestG3VerificationConditionFallsBackUnderMemoryPressure(t *testing.T) {
	for _, mode := range []string{"Single", "Batch"} {
		t.Run(mode, func(t *testing.T) {
			ctx := NewZ3Context()
			defer ctx.Close()
			s := ctx.NewZ3Solver()
			check := s.CheckVerificationCondition
			if mode == "Batch" {
				var release func()
				check, release = s.BeginVerificationChecks()
				defer release()
			}
			domain := ctx.BvSort(2)
			x := ctx.Const("x", domain)
			a := ctx.Const("a", domain)
			p := ctx.Function("p", []Z3Sort{domain}, ctx.BoolSort())
			s.Assert(ctx.ForAll([]Z3Expr{x}, p.Apply(x)))
			s.Push()
			s.Assert(ctx.Not(p.Apply(a)))
			if got := s.Check(); got != Unsat {
				t.Fatal("concrete control", got)
			}
			m, err := ctx.c.Manager()
			if err != nil {
				t.Fatal(err)
			}
			budget := m.Budget()
			held := budget.Limit() - budget.Used() - 32*1024
			if err = budget.Reserve(held); err != nil {
				t.Fatal(err)
			}
			defer budget.Release(held)
			if got := check(); got != Unsat {
				t.Fatal("cached concrete result", got)
			}
		})
	}
}

func TestG3EnumSortReleasesUnusedTesters(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()
	names := []string{"north", "south", "west"}
	sort, values := ctx.EnumSort("Direction", names)
	manager, err := ctx.c.Manager()
	if err != nil {
		t.Fatal(err)
	}
	before := manager.Stats().LiveNodes
	// Each tester must be recreated on demand. Its constructor is still
	// owned by the corresponding enumeration value.
	for i, value := range values {
		tester, err := ctx.c.DatatypeRecognizer(sort.c, uint32(i))
		if err != nil {
			t.Fatal(err)
		}
		if err := ctx.c.IncASTRef(tester); err != nil {
			t.Fatal(err)
		}
		if got := manager.Stats().LiveNodes; got != before+1 {
			t.Fatalf("tester %d: got %d live nodes, want %d", i, got, before+1)
		}
		if err := ctx.c.DecASTRef(tester); err != nil {
			t.Fatal(err)
		}
		if got := manager.Stats().LiveNodes; got != before {
			t.Fatalf("tester %d left %d live nodes, want %d", i, got, before)
		}
		if got := value.String(); got != names[i] {
			t.Fatalf("enum value %d: %s", i, got)
		}
	}
}

func TestG3VerificationGoalPreservesCounterexample(t *testing.T) {
	ctx := NewZ3Context()
	defer ctx.Close()
	s := ctx.NewZ3Solver()
	check, proveGoal, release := s.BeginVerificationGoalChecks()
	if proveGoal == nil {
		t.Skip("pinned g3 release has no temporary-goal proof")
	}
	defer release()
	sort := ctx.BvSort(2)
	x := ctx.Const("x", sort)
	a := ctx.Const("a", sort)
	pred := ctx.Function("p", []Z3Sort{sort}, ctx.BoolSort())
	s.Assert(ctx.ForAll([]Z3Expr{x}, pred.Apply(x)))
	base := s.CanonZ3Assertions()
	if !proveGoal(ctx.Not(pred.Apply(a))) {
		t.Fatal("failed to refute an instance of the universal premise")
	}
	if got := s.CanonZ3Assertions(); got != base {
		t.Fatalf("goal changed source assertions: %s", got)
	}
	three := ctx.BvVal(3, 2)
	counterexample := ctx.Eq(a, three)
	if proveGoal(counterexample) {
		t.Fatal("refuted a satisfiable goal")
	}
	s.Push()
	s.Assert(counterexample)
	if got := check(); got != Sat {
		t.Fatal("concrete fallback", got)
	}
	model := s.Model()
	if model == nil {
		t.Fatal("missing diagnostic model")
	}
	defer model.Release()
	if val, ok := model.Eval(a, true); !ok || !val.Equal(three) {
		t.Fatalf("model value %v, %t", val, ok)
	}
	s.Pop()
	if got := s.CanonZ3Assertions(); got != base {
		t.Fatalf("fallback pop changed base: %s", got)
	}
}
