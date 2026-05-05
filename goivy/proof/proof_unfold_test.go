package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

var _ = ast.NewAstConfig // ensure import

// --- ExprListOrLambdaUnion tests ---

func TestExprListOrLambdaUnion_PopSingle(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
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
	s := proofMkSort("S")
	a := proofMkConst("a", s)
	b := proofMkConst("b", s)
	c := proofMkConst("c", s)
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
	return mkLF(proofTestAstCfg.NewAtom(name), body)
}

func TestMatchFromDefn_Simple(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)

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
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	// Not an equation — just a constant
	defn := mkLF(proofTestAstCfg.NewAtom("bad"), c)

	_, err := MatchFromDefn(defn)
	if err == nil {
		t.Fatal("expected error for non-definition")
	}
}

// --- MatchFromDefns tests ---

func TestMatchFromDefns_Single(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)

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
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)

	c1 := proofMkConst("c1", s)
	c2 := proofMkConst("c2", s)

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
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	g := proofMkConst("g", fs)

	defn1 := mkDefnLF("def_f", x, f, x)
	defn2 := mkDefnLF("def_g", x, g, x)

	_, _, err := MatchFromDefns([]*ast.LabeledFormula{defn1, defn2})
	if err == nil {
		t.Fatal("expected error for different LHS symbols")
	}
}

// --- unfoldRhsVars tests ---

func TestUnfoldRhsVars(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	c1 := proofMkConst("c1", s)
	c2 := proofMkConst("c2", s)

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
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	c := proofMkConst("c", s)

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
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	c1 := proofMkConst("c1", s)
	c2 := proofMkConst("c2", s)
	a := proofMkConst("a", s)

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
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	g := proofMkConst("g", fs)
	a := proofMkConst("a", s)

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
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	a := proofMkConst("a", s)

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
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	a := proofMkConst("a", s)

	// Definition: forall X. f(X) = X
	defn := mkDefnLF("def_f", x, f, x)

	// Goal with conclusion f(a)
	app := lg.MustApply(f, a)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), app)

	result := UnfoldGoal(proofTestAstCfg, goal, [][]*ast.LabeledFormula{{defn}})

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
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	y := proofMkVar("Y", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	c := proofMkConst("c", s)

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

// ==========================================================================
// CR-1: MatchFromDefn must reject duplicate parameters (Python: iu.distinct)
// ==========================================================================

func TestMatchFromDefn_DuplicateParams(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s, s) // S x S -> S
	f := proofMkConst("f", fs)
	c := proofMkConst("c", s)

	// Build: forall X. f(X, X) = c — X appears twice as arg to f
	app := lg.MustApply(f, x, x)
	eq := &lg.Eq{T1: app, T2: c}
	body := &lg.ForAll{Variables: []*lg.Variable{x}, Body: eq}
	defn := mkLF(proofTestAstCfg.NewAtom("def"), body)

	_, err := MatchFromDefn(defn)
	if err == nil {
		t.Fatal("expected error for duplicate parameters f(X, X)")
	}
}

func TestMatchFromDefn_DistinctParams(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	y := proofMkVar("Y", s)
	fs, _ := lg.NewFunctionSort(s, s, s) // S x S -> S
	f := proofMkConst("f", fs)
	c := proofMkConst("c", s)

	// Build: forall X, Y. f(X, Y) = c — distinct params, should succeed
	app := lg.MustApply(f, x, y)
	eq := &lg.Eq{T1: app, T2: c}
	body := &lg.ForAll{Variables: []*lg.Variable{x, y}, Body: eq}
	defn := mkLF(proofTestAstCfg.NewAtom("def"), body)

	match, err := MatchFromDefn(defn)
	if err != nil {
		t.Fatalf("unexpected error for distinct params: %v", err)
	}
	lam, ok := match[lg.Key(f)]
	if !ok {
		t.Fatal("expected f in match")
	}
	l, ok := lam.(*lg.Lambda)
	if !ok {
		t.Fatalf("expected *lg.Lambda, got %T", lam)
	}
	if len(l.Variables) != 2 {
		t.Errorf("expected 2 lambda variables, got %d", len(l.Variables))
	}
}

func TestMatchFromDefn_DuplicateParams_Iff(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s, lg.Boolean) // S x S -> Bool
	p := proofMkConst("p", fs)

	// Build: forall X. p(X, X) <-> true — duplicate params via Iff
	app := lg.MustApply(p, x, x)
	iff := &lg.Iff{T1: app, T2: lg.True}
	body := &lg.ForAll{Variables: []*lg.Variable{x}, Body: iff}
	defn := mkLF(proofTestAstCfg.NewAtom("def"), body)

	_, err := MatchFromDefn(defn)
	if err == nil {
		t.Fatal("expected error for duplicate parameters p(X, X) via Iff")
	}
}

func TestDistinctVars(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	y := proofMkVar("Y", s)

	if !distinctVars([]*lg.Variable{x, y}) {
		t.Error("X, Y should be distinct")
	}
	if !distinctVars([]*lg.Variable{x}) {
		t.Error("single var should be distinct")
	}
	if !distinctVars(nil) {
		t.Error("empty should be distinct")
	}
	if distinctVars([]*lg.Variable{x, x}) {
		t.Error("X, X should NOT be distinct")
	}
}

// ==========================================================================
// CR-2: betaReduce must use capture-detecting substitution
// ==========================================================================

func TestBetaReduce_CaptureDetected(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	y := proofMkVar("Y", s)

	// Lambda(X, forall Y. X = Y)
	eq := &lg.Eq{T1: x, T2: y}
	body := &lg.ForAll{Variables: []*lg.Variable{y}, Body: eq}
	lam := &lg.Lambda{Variables: []*lg.Variable{x}, Body: body}

	// Apply to arg Y — this would capture Y in the forall
	result := betaReduce(lam, []lg.Expr{y})

	// Should NOT produce forall Y. Y = Y (captured).
	// betaReduce returns lam.Body on CaptureError.
	if fa, ok := result.(*lg.ForAll); ok {
		if eq, ok := fa.Body.(*lg.Eq); ok {
			// If both sides are the same variable Y, that's capture corruption
			lv, lOk := eq.T1.(*lg.Variable)
			rv, rOk := eq.T2.(*lg.Variable)
			if lOk && rOk && lv.Name == "Y" && rv.Name == "Y" {
				t.Fatal("capture detected: betaReduce produced forall Y. Y=Y")
			}
		}
	}
	// Result should be the unchanged lambda body (forall Y. X = Y)
	// because capture was detected and substitution was skipped.
	fa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("expected ForAll, got %T", result)
	}
	if eqR, ok := fa.Body.(*lg.Eq); ok {
		if lv, ok := eqR.T1.(*lg.Variable); ok {
			if lv.Name != "X" {
				t.Errorf("expected X in LHS (unsubstituted), got %s", lv.Name)
			}
		}
	}
}

func TestBetaReduce_NoCaptureSucceeds(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	c := proofMkConst("c", s)

	// Lambda(X, X) applied to c — identity, no capture possible
	lam := &lg.Lambda{Variables: []*lg.Variable{x}, Body: x}
	result := betaReduce(lam, []lg.Expr{c})

	rc, ok := result.(*lg.Const)
	if !ok {
		t.Fatalf("expected Const c, got %T: %v", result, result)
	}
	if rc.Name != "c" {
		t.Errorf("expected c, got %s", rc.Name)
	}
}

func TestApplyMatch_CaptureNotCorrupted(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	y := proofMkVar("Y", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)

	// match: {f: Lambda(X, forall Y. X = Y)}
	eq := &lg.Eq{T1: x, T2: y}
	body := &lg.ForAll{Variables: []*lg.Variable{y}, Body: eq}
	lam := &lg.Lambda{Variables: []*lg.Variable{x}, Body: body}
	match := map[lg.NodeKey]lg.Expr{lg.Key(f): lam}

	// formula: f(Y) — applying match would substitute Y for X,
	// but Y is also bound in the lambda body's forall.
	fmla := lg.MustApply(f, y)

	result := ApplyMatch(match, fmla)

	// Result should NOT have Y captured by the forall.
	// With capture detection, betaReduce returns the lambda body unchanged.
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

// ==========================================================================
// CR-3: applyMatchAltRec must handle LambdaApply errors (not return nil)
// ==========================================================================

func TestApplyMatchAltRec_LambdaApplyError_NotNil(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	y := proofMkVar("Y", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)

	// match: {f: Lambda(X, forall Y. X = Y)}
	eq := &lg.Eq{T1: x, T2: y}
	body := &lg.ForAll{Variables: []*lg.Variable{y}, Body: eq}
	lam := &lg.Lambda{Variables: []*lg.Variable{x}, Body: body}
	match := map[lg.NodeKey]lg.Expr{lg.Key(f): lam}

	// formula: f(Y) — capture scenario
	fmla := lg.MustApply(f, y)
	env := make(map[lg.NodeKey]bool)

	result := applyMatchAltRec(match, fmla, env)

	// Must NOT be nil — that would cause a crash downstream
	if result == nil {
		t.Fatal("applyMatchAltRec returned nil on CaptureError — should return fmla")
	}
}

// ==========================================================================
// CR-4: applyUnfoldRec must handle LambdaApply errors
// ==========================================================================

func TestApplyUnfoldRec_CaptureReturnsOriginal(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	y := proofMkVar("Y", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)

	// Lambda whose body has bound Y: Lambda(X, forall Y. X = Y)
	eq := &lg.Eq{T1: x, T2: y}
	body := &lg.ForAll{Variables: []*lg.Variable{y}, Body: eq}
	lam := &lg.Lambda{Variables: []*lg.Variable{x}, Body: body}
	union := &ExprListOrLambdaUnion{Items: []lg.Expr{lam}}

	// Formula: f(Y) — applying lambda would capture Y
	fmla := lg.MustApply(f, y)

	result := applyUnfoldRec(lg.Key(f), union, fmla)

	// Must NOT be nil
	if result == nil {
		t.Fatal("applyUnfoldRec returned nil on capture — should return original fmla")
	}
}

// ==========================================================================
// CR-5: applyUnfoldGoal must filter ConstantDecl matching unfold key
// ==========================================================================

func TestApplyUnfoldGoal_FiltersConstantDeclPremise(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	c := proofMkConst("c", s)
	a := proofMkConst("a", s)

	// Lambda: f(X) = c
	lam, _ := lg.NewLambda([]*lg.Variable{x}, c)
	union := &ExprListOrLambdaUnion{Items: []lg.Expr{lam}}
	freeVars := unfoldRhsVars(union)

	// Goal with ConstantDecl(f) premise + conclusion f(a)
	cd := proofTestAstCfg.NewConstantDecl(f)
	app := lg.MustApply(f, a)
	sb := proofTestAstCfg.NewSchemaBody(cd, app)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	result := applyUnfoldGoal(proofTestAstCfg, lg.Key(f), union, freeVars, goal)

	// The ConstantDecl(f) premise should be FILTERED OUT
	// (Python: transformed to lambda-typed, then filtered by is_lambda check)
	prems := GoalPrems(result)
	for _, p := range prems {
		if cd, ok := p.(*ast.ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if rc, ok := args[0].(*lg.Const); ok && rc.Name == "f" {
					t.Error("ConstantDecl(f) should have been filtered from premises")
				}
			}
		}
	}

	// Conclusion should be unfolded: f(a) → c
	conc := GoalConc(result)
	if rc, ok := conc.(*lg.Const); ok {
		if rc.Name != "c" {
			t.Errorf("expected conclusion c, got %s", rc.Name)
		}
	} else {
		t.Errorf("expected *lg.Const conclusion, got %T", conc)
	}
}

func TestApplyUnfoldGoal_KeepsNonMatchingConstantDecl(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	g := proofMkConst("g", fs)
	c := proofMkConst("c", s)
	a := proofMkConst("a", s)

	// Lambda: f(X) = c
	lam, _ := lg.NewLambda([]*lg.Variable{x}, c)
	union := &ExprListOrLambdaUnion{Items: []lg.Expr{lam}}
	freeVars := unfoldRhsVars(union)

	// Goal with ConstantDecl(g) premise (NOT the unfold key f) + conclusion f(a)
	cd := proofTestAstCfg.NewConstantDecl(g)
	app := lg.MustApply(f, a)
	sb := proofTestAstCfg.NewSchemaBody(cd, app)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	result := applyUnfoldGoal(proofTestAstCfg, lg.Key(f), union, freeVars, goal)

	// ConstantDecl(g) should be KEPT (different symbol from unfold key)
	prems := GoalPrems(result)
	foundG := false
	for _, p := range prems {
		if cd, ok := p.(*ast.ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if rc, ok := args[0].(*lg.Const); ok && rc.Name == "g" {
					foundG = true
				}
			}
		}
	}
	if !foundG {
		t.Error("ConstantDecl(g) should have been kept in premises")
	}
}

// ==========================================================================
// CR-6: applyUnfoldRec non-lambda Pop must apply to args (Python Symbol.__call__)
// ==========================================================================

func TestApplyUnfoldRec_NonLambdaPopAppliesArgs(t *testing.T) {
	s := proofMkSort("S")
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	g := proofMkConst("g", fs)
	a := proofMkConst("a", s)

	// Union contains a non-lambda Const (g) instead of a Lambda.
	// Python: Symbol.__call__(*args) creates Apply(g, args).
	union := &ExprListOrLambdaUnion{Items: []lg.Expr{g}}

	// Formula: f(a)
	fmla := lg.MustApply(f, a)

	result := applyUnfoldRec(lg.Key(f), union, fmla)

	// Result should be g(a) — the popped g applied to the processed arg a.
	// NOT just g with args dropped.
	app, ok := result.(*lg.Apply)
	if !ok {
		t.Fatalf("expected Apply (g applied to a), got %T: %v", result, result)
	}
	if rc, ok := app.Func.(*lg.Const); !ok || rc.Name != "g" {
		t.Errorf("expected function g, got %v", app.Func)
	}
	if len(app.Terms) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(app.Terms))
	}
	if rc, ok := app.Terms[0].(*lg.Const); !ok || rc.Name != "a" {
		t.Errorf("expected arg a, got %v", app.Terms[0])
	}
}

func TestApplyUnfoldRec_NonLambdaPopSortMismatch(t *testing.T) {
	s := proofMkSort("S")
	fs, _ := lg.NewFunctionSort(s, s)
	f := proofMkConst("f", fs)
	c := proofMkConst("c", s) // sort S, NOT a function sort

	// Union contains a non-lambda, non-function-sort Const.
	// NewApply would fail (sort mismatch), so the fallback returns c itself.
	union := &ExprListOrLambdaUnion{Items: []lg.Expr{c}}

	// Formula: f(a)
	a := proofMkConst("a", s)
	fmla := lg.MustApply(f, a)
	result := applyUnfoldRec(lg.Key(f), union, fmla)

	// Must not panic, must not be nil
	if result == nil {
		t.Fatal("result should not be nil")
	}
	// Result is c itself (NewApply fails due to sort mismatch)
	if rc, ok := result.(*lg.Const); ok {
		if rc.Name != "c" {
			t.Errorf("expected c, got %s", rc.Name)
		}
	}
}
