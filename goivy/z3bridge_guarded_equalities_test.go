package goivy

import (
	"testing"

	"github.com/glycerine/ivy/goivy/smt"
)

func TestGuardedEqualityBatch(t *testing.T) {
	mod := New()
	s := NewSolver(mod, DefaultSolverOptions())
	defer s.Close()
	left := NewConst("left", Boolean)
	right := NewConst("right", Boolean)
	guard := &LogicAnd{Terms: []Expr{left, right}}
	mk := func(antecedent, consequence Expr) FinalCond {
		neg := &LogicNot{Body: &LogicImplies{T1: antecedent, T2: consequence}}
		return &ConjChecker{BaseChecker: BaseChecker{Mod: mod, FC: NewClauses([]Expr{neg}, nil, nil)}}
	}
	equal := func(a, b Expr) *Eq { return &Eq{T1: a, T2: b} }
	goals := []FinalCond{
		mk(guard, equal(left, right)),
		mk(&LogicAnd{Terms: []Expr{left, right}}, equal(right, True)),
		mk(guard, equal(left, True)),
	}
	if n := guardedEqualityCount(goals); n != len(goals) {
		t.Fatalf("equal guards grouped %d of %d", n, len(goals))
	}
	for _, stop := range []FinalCond{
		mk(left, equal(left, right)),
		mk(&LogicAnd{Terms: []Expr{right, left}}, equal(left, right)),
		mk(guard, left),
		&ConjAssumer{},
		&BaseChecker{},
	} {
		if n := guardedEqualityCount([]FinalCond{goals[0], stop, goals[2]}); n != 1 {
			t.Fatalf("group crossed %T, got %d goals", stop, n)
		}
	}
	var many []FinalCond
	for range 17 {
		many = append(many, goals[0])
	}
	if n := guardedEqualityCount(many); n != 16 {
		t.Fatalf("group size %d", n)
	}

	first, err := s.ClausesToZ3(goals[0].Cond())
	if err != nil {
		t.Fatal(err)
	}
	zs := s.newZ3Solver()
	defer zs.Reset()
	if !s.proveGuardedEqualities(zs, first, goals[1:]) {
		t.Fatal("expected the common guard to establish all three equalities")
	}
	goals[1] = mk(guard, equal(left, False))
	if s.proveGuardedEqualities(zs, first, goals[1:]) {
		t.Fatal("a satisfiable counterexample disjunction was accepted")
	}
	for i, want := range []smt.Z3CheckResult{smt.Unsat, smt.Sat, smt.Unsat} {
		goal, err := s.ClausesToZ3(goals[i].Cond())
		if err != nil {
			t.Fatal(err)
		}
		zs.Push()
		zs.Assert(goal)
		if got := zs.Check(); got != want {
			t.Fatalf("fallback for goal %d got %v, want %v", i, got, want)
		}
		zs.Pop()
	}
	if got := zs.Check(); got != smt.Sat {
		t.Fatalf("temporary assertions survived pop, got %v", got)
	}

	mod.Cfg.Diagnose = true
	goals[2] = mk(guard, equal(right, False))
	result, err := s.GetSmallModelWithCond(TrueClauses(nil), nil, nil, goals, false)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.Model == nil {
		t.Fatal("expected an individual counterexample model")
	}
	if !goals[1].(*ConjChecker).Failed() || goals[2].(*ConjChecker).Failed() || mod.Cfg.Failures != 1 {
		t.Fatal("expected reporting to stop at the first failing equality")
	}
	counterexample, err := s.EvalFormula(result.Model, goals[1].Cond().Fmlas[0])
	if err != nil || !counterexample {
		t.Fatalf("model did not satisfy the individual counterexample: %v, %v", counterexample, err)
	}
}
