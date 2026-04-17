package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

var _ = ast.NewAstConfig // ensure import

// --- ExprListOrLambdaUnion tests ---

func TestExprListOrLambdaUnion_PopSingle(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	u := &ExprListOrLambdaUnion{Items: []lg.Expr{c}}

	// Pop from single-item list returns the item
	got := u.Pop()
	if got != c {
		t.Fatalf("expected c, got %v", got)
	}
	// Pop again — single item stays (Python: if len(save) > 1: del save[0])
	got2 := u.Pop()
	if got2 != c {
		t.Fatalf("second pop should return same item, got %v", got2)
	}
}

func TestExprListOrLambdaUnion_PopMultiple(t *testing.T) {
	s := mkSort("S")
	a := mkConst("a", s)
	b := mkConst("b", s)
	c := mkConst("c", s)
	u := &ExprListOrLambdaUnion{Items: []lg.Expr{a, b, c}}

	// First pop → a
	got1 := u.Pop()
	if got1 != a {
		t.Fatalf("first pop: expected a, got %v", got1)
	}
	// Second pop → b
	got2 := u.Pop()
	if got2 != b {
		t.Fatalf("second pop: expected b, got %v", got2)
	}
	// Third pop → c (last item, stays)
	got3 := u.Pop()
	if got3 != c {
		t.Fatalf("third pop: expected c, got %v", got3)
	}
	// Fourth pop → c again (single item remains)
	got4 := u.Pop()
	if got4 != c {
		t.Fatalf("fourth pop: expected c again, got %v", got4)
	}
}

func TestExprListOrLambdaUnion_PopEmpty(t *testing.T) {
	u := &ExprListOrLambdaUnion{Items: nil}
	got := u.Pop()
	if got != nil {
		t.Fatalf("expected nil from empty Pop, got %v", got)
	}
}

// --- MatchFromDefn tests ---

// mkDefnLF builds a LabeledFormula for a definition: forall X. f(X) = rhs
func mkDefnLF(name string, param *lg.Variable, f *lg.Const, rhs lg.Expr) *ast.LabeledFormula {
	app := lg.MustApply(f, param)
	eq := &lg.Eq{T1: app, T2: rhs}
	body := &lg.ForAll{Variables: []*lg.Variable{param}, Body: eq}
	return mkLF(testAstCfg.NewAtom(name), body)
}

func TestMatchFromDefn_Simple(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)

	defn := mkDefnLF("def_f", x, f, x) // forall X. f(X) = X (identity)

	match, err := MatchFromDefn(defn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lam, ok := match[lg.Key(f)]
	if !ok {
		t.Fatal("expected f in match")
	}
	l, ok := lam.(*lg.Lambda)
	if !ok {
		t.Fatalf("expected *lg.Lambda, got %T", lam)
	}
	if len(l.Variables) != 1 {
		t.Errorf("expected 1 lambda variable, got %d", len(l.Variables))
	}
}

func TestMatchFromDefn_NotADefinition(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	// Not an equation — just a constant
	defn := mkLF(testAstCfg.NewAtom("bad"), c)

	_, err := MatchFromDefn(defn)
	if err == nil {
		t.Fatal("expected error for non-definition")
	}
}

// --- MatchFromDefns tests ---

func TestMatchFromDefns_Single(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)

	defn := mkDefnLF("def_f", x, f, x)

	key, union, err := MatchFromDefns([]*ast.LabeledFormula{defn})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != lg.Key(f) {
		t.Errorf("expected key for f, got %v", key)
	}
	if len(union.Items) != 1 {
		t.Fatalf("expected 1 item in union, got %d", len(union.Items))
	}
	if _, ok := union.Items[0].(*lg.Lambda); !ok {
		t.Errorf("expected lambda, got %T", union.Items[0])
	}
}

func TestMatchFromDefns_Multiple(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)

	c1 := mkConst("c1", s)
	c2 := mkConst("c2", s)

	// Two definitions for f with different RHS
	defn1 := mkDefnLF("def_f1", x, f, c1) // forall X. f(X) = c1
	defn2 := mkDefnLF("def_f2", x, f, c2) // forall X. f(X) = c2

	key, union, err := MatchFromDefns([]*ast.LabeledFormula{defn1, defn2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != lg.Key(f) {
		t.Errorf("expected key for f, got %v", key)
	}
	if len(union.Items) != 2 {
		t.Fatalf("expected 2 items in union, got %d", len(union.Items))
	}
}

func TestMatchFromDefns_Empty(t *testing.T) {
	_, _, err := MatchFromDefns(nil)
	if err == nil {
		t.Fatal("expected error for empty defns")
	}
}

func TestMatchFromDefns_DifferentLHS(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)
	g := mkConst("g", fs)

	defn1 := mkDefnLF("def_f", x, f, x)
	defn2 := mkDefnLF("def_g", x, g, x)

	_, _, err := MatchFromDefns([]*ast.LabeledFormula{defn1, defn2})
	if err == nil {
		t.Fatal("expected error for different LHS symbols")
	}
}

// --- unfoldRhsVars tests ---

func TestUnfoldRhsVars(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	c1 := mkConst("c1", s)
	c2 := mkConst("c2", s)

	// Lambdas whose bodies contain free constants c1, c2
	lam1, _ := lg.NewLambda([]*lg.Variable{x}, c1) // Lambda(X, c1) — c1 is free
	lam2, _ := lg.NewLambda([]*lg.Variable{x}, c2) // Lambda(X, c2) — c2 is free

	union := &ExprListOrLambdaUnion{Items: []lg.Expr{lam1, lam2}}
	vars := unfoldRhsVars(union)

	// Should contain c1 and c2 from the lambda bodies
	if len(vars) == 0 {
		t.Error("expected non-empty free vars from lambda bodies")
	}
	if _, ok := vars[lg.Key(c1)]; !ok {
		t.Error("expected c1 in free vars")
	}
	if _, ok := vars[lg.Key(c2)]; !ok {
		t.Error("expected c2 in free vars")
	}
}

// --- UnfoldFmla tests ---

func TestUnfoldFmla_SingleOccurrence(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)
	c := mkConst("c", s)

	// Definition: forall X. f(X) = c (constant function)
	defn := mkDefnLF("def_f", x, f, c)

	// Formula: f(c) — should unfold to c
	app := lg.MustApply(f, c)

	result := UnfoldFmla(app, [][]*ast.LabeledFormula{{defn}})

	// After unfolding f(c) with f(X) = c, we should get c
	if rc, ok := result.(*lg.Const); ok {
		if rc.Name != "c" {
			t.Errorf("expected const c, got %s", rc.Name)
		}
	} else {
		t.Errorf("expected *lg.Const after unfold, got %T: %v", result, result)
	}
}

func TestUnfoldFmla_MultipleOccurrences_DestructivePop(t *testing.T) {
	// This is the KEY BUG-11 test: multiple occurrences get different lambdas.
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)
	c1 := mkConst("c1", s)
	c2 := mkConst("c2", s)
	a := mkConst("a", s)

	// Two definitions for f with different RHS (simulating renamed versions)
	defn1 := mkDefnLF("def_f1", x, f, c1) // forall X. f(X) = c1
	defn2 := mkDefnLF("def_f2", x, f, c2) // forall X. f(X) = c2

	// Formula: f(a) = f(a) — two occurrences of f
	app1 := lg.MustApply(f, a)
	app2 := lg.MustApply(f, a)
	fmla := &lg.Eq{T1: app1, T2: app2}

	result := UnfoldFmla(fmla, [][]*ast.LabeledFormula{{defn1, defn2}})

	// After unfolding with destructive pop:
	// first f(a) → c1 (from defn1's lambda)
	// second f(a) → c2 (from defn2's lambda)
	// Result should be: c1 = c2
	eq, ok := result.(*lg.Eq)
	if !ok {
		t.Fatalf("expected *lg.Eq, got %T: %v", result, result)
	}
	lhs, lhsOk := eq.T1.(*lg.Const)
	rhs, rhsOk := eq.T2.(*lg.Const)
	if !lhsOk || !rhsOk {
		t.Fatalf("expected Const = Const, got %T = %T", eq.T1, eq.T2)
	}
	if lhs.Name != "c1" {
		t.Errorf("expected LHS = c1, got %s", lhs.Name)
	}
	if rhs.Name != "c2" {
		t.Errorf("expected RHS = c2, got %s", rhs.Name)
	}
}

func TestUnfoldFmla_NoMatch(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)
	g := mkConst("g", fs)
	a := mkConst("a", s)

	// Definition for f
	defn := mkDefnLF("def_f", x, f, x)

	// Formula uses g, not f — should be unchanged
	app := lg.MustApply(g, a)

	result := UnfoldFmla(app, [][]*ast.LabeledFormula{{defn}})

	if rapp, ok := result.(*lg.Apply); ok {
		if rc, ok := rapp.Func.(*lg.Const); ok {
			if rc.Name != "g" {
				t.Errorf("expected g unchanged, got %s", rc.Name)
			}
		}
	} else {
		t.Errorf("expected *lg.Apply unchanged, got %T", result)
	}
}

func TestUnfoldFmla_NestedOccurrence(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)
	a := mkConst("a", s)

	// Definition: forall X. f(X) = X (identity)
	defn := mkDefnLF("def_f", x, f, x)

	// Formula: f(f(a)) — nested occurrence
	inner := lg.MustApply(f, a)
	outer := lg.MustApply(f, inner)

	result := UnfoldFmla(outer, [][]*ast.LabeledFormula{{defn}})

	// f(f(a)): inner f(a) → a, then outer f(a) → a
	// Both use the same single lambda (no pop since len==1)
	if rc, ok := result.(*lg.Const); ok {
		if rc.Name != "a" {
			t.Errorf("expected a after nested unfold, got %s", rc.Name)
		}
	} else {
		t.Errorf("expected *lg.Const after nested unfold, got %T: %v", result, result)
	}
}

// --- UnfoldGoal tests ---

func TestUnfoldGoal_Simple(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)
	a := mkConst("a", s)

	// Definition: forall X. f(X) = X
	defn := mkDefnLF("def_f", x, f, x)

	// Goal with conclusion f(a)
	app := lg.MustApply(f, a)
	goal := mkLF(testAstCfg.NewAtom("g"), app)

	result := UnfoldGoal(testAstCfg, goal, [][]*ast.LabeledFormula{{defn}})

	conc := GoalConc(result)
	if rc, ok := conc.(*lg.Const); ok {
		if rc.Name != "a" {
			t.Errorf("expected conclusion a, got %s", rc.Name)
		}
	} else {
		t.Errorf("expected *lg.Const conclusion, got %T", conc)
	}
}

// --- applyUnfoldRec tests ---

func TestApplyUnfoldRec_UnderQuantifier(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := mkConst("f", fs)
	c := mkConst("c", s)

	// Lambda for f: f(X) = c
	lam, _ := lg.NewLambda([]*lg.Variable{x}, c)
	union := &ExprListOrLambdaUnion{Items: []lg.Expr{lam}}

	// Formula: forall Y. f(Y)
	app := lg.MustApply(f, y)
	fmla := &lg.ForAll{Variables: []*lg.Variable{y}, Body: app}

	result := applyUnfoldRec(lg.Key(f), union, fmla)

	// Should become: forall Y. c
	fa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("expected ForAll, got %T", result)
	}
	if rc, ok := fa.Body.(*lg.Const); !ok || rc.Name != "c" {
		t.Errorf("expected body c, got %T: %v", fa.Body, fa.Body)
	}
}
