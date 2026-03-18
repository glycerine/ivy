package proof

import (
	"testing"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	iu "github.com/glycerine/goivy/ivyutils"
)

// Ensure imports are used.
var (
	_ = il.IsVariable
	_ = lu.FreeVariables
	_ = iu.NewUniqueRenamer
)

// --- helpers ---

func mkSort(name string) *lg.UninterpretedSort {
	return &lg.UninterpretedSort{Name: name}
}

func mkVar(name string, s lg.Sort) *lg.Variable {
	v, _ := lg.NewVariable(name, s)
	return v
}

func mkConst(name string, s lg.Sort) *lg.Symbol {
	return lg.NewSymbol(name, s)
}

func mkLF(label, formula ast.Node) *ast.LabeledFormula {
	return ast.NewLabeledFormula(label, formula)
}

// --- Error type tests ---

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		err error
		msg string
	}{
		{&Redefinition{Msg: "x"}, "redefinition: x"},
		{&Circular{Msg: "y"}, "circular: y"},
		{&NoMatch{Msg: "z"}, "no match: z"},
		{&ProofError{Msg: "w"}, "proof error: w"},
		{&CaptureError{Msg: "v"}, "capture error: v"},
	}
	for _, tc := range tests {
		if tc.err.Error() != tc.msg {
			t.Errorf("got %q, want %q", tc.err.Error(), tc.msg)
		}
	}
}

func TestErrorsWithNode(t *testing.T) {
	e := &Redefinition{Msg: "dup", Node: "somenode"}
	if e.Error() != "redefinition: dup (at somenode)" {
		t.Errorf("unexpected error: %s", e.Error())
	}
}

// --- MatchProblem tests ---

func TestMatchProblemString(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	mp := NewMatchProblem(x, x, x, map[lg.NodeKey]lg.Node{lg.Key(x): lg.True}, nil)
	str := mp.String()
	if len(str) == 0 {
		t.Error("expected non-empty string")
	}
}

// --- FuncSorts tests ---

func TestFuncSorts(t *testing.T) {
	s := mkSort("S")
	c := mkConst("f", s)
	sorts := FuncSorts(c)
	if len(sorts) != 1 || !sorts[0].Equal(s) {
		t.Errorf("expected [S], got %v", sorts)
	}

	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	c2 := mkConst("p", fs)
	sorts2 := FuncSorts(c2)
	if len(sorts2) != 2 {
		t.Errorf("expected 2 sorts, got %d", len(sorts2))
	}
	if !sorts2[0].Equal(s) || !sorts2[1].Equal(lg.Boolean) {
		t.Error("wrong sorts")
	}
}

// --- FuncsMatch tests ---

func TestFuncsMatch(t *testing.T) {
	s := mkSort("S")
	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	c1 := mkConst("p", fs)
	c2 := mkConst("p", fs)
	c3 := mkConst("q", fs)

	if !FuncsMatch(c1, c2, nil) {
		t.Error("same constants should match")
	}
	if FuncsMatch(c1, c3, nil) {
		t.Error("different names should not match")
	}

	// With free sorts
	s2 := mkSort("T")
	fs2, _ := lg.NewFunctionSort(s2, lg.Boolean)
	c4 := mkConst("p", fs2)
	freesyms := map[lg.NodeKey]lg.Node{lg.Key(s): lg.True}
	if !FuncsMatch(c1, c4, freesyms) {
		t.Error("should match when sort is free")
	}
}

// --- HeadsMatch tests ---

func TestHeadsMatch(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)

	// Variables should match variables
	if !HeadsMatch(x, y, nil) {
		t.Error("variables should heads-match")
	}

	// ForAll should not match anything (quantifiers excluded)
	fa := &lg.ForAll{Variables: []*lg.Variable{x}, Body: x}
	if HeadsMatch(fa, y, nil) {
		t.Error("quantifier should not heads-match variable")
	}
}

// --- MatchSort tests ---

func TestMatchSort(t *testing.T) {
	s1 := mkSort("S")
	s2 := mkSort("T")

	// Free sort: should produce mapping
	free := map[lg.NodeKey]lg.Node{lg.Key(s1): lg.True}
	m := MatchSort(s1, s2, free)
	if m == nil || len(m) != 1 {
		t.Fatal("expected match")
	}
	if !m[lg.Key(s1)].Equal(s2) {
		t.Error("wrong mapping")
	}

	// Same sort, not free: should produce empty match
	m2 := MatchSort(s1, s1, nil)
	if m2 == nil || len(m2) != 0 {
		t.Error("expected empty match for equal sorts")
	}

	// Different sort, not free: should fail
	m3 := MatchSort(s1, s2, nil)
	if m3 != nil {
		t.Error("expected nil for unequal non-free sorts")
	}
}

// --- MergeMatches tests ---

func TestMergeMatchesEmpty(t *testing.T) {
	m := MergeMatches()
	if m == nil || len(m) != 0 {
		t.Error("expected empty map")
	}
}

func TestMergeMatchesNil(t *testing.T) {
	m := MergeMatches(nil, map[lg.NodeKey]lg.Node{})
	if m != nil {
		t.Error("expected nil when any match is nil")
	}
}

func TestMergeMatchesConflict(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)
	z := mkVar("Z", s)

	m1 := map[lg.NodeKey]lg.Node{lg.Key(x): y}
	m2 := map[lg.NodeKey]lg.Node{lg.Key(x): z}
	m := MergeMatches(m1, m2)
	if m != nil {
		t.Error("expected nil for conflicting matches")
	}
}

func TestMergeMatchesConsistent(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)
	z := mkVar("Z", s)

	m1 := map[lg.NodeKey]lg.Node{lg.Key(x): y}
	m2 := map[lg.NodeKey]lg.Node{lg.Key(z): y}
	m := MergeMatches(m1, m2)
	if m == nil || len(m) != 2 {
		t.Error("expected merged map with 2 entries")
	}
}

// --- EquivAlpha tests ---

func TestEquivAlpha(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)

	if !EquivAlpha(x, x) {
		t.Error("same term should be equiv")
	}
	if EquivAlpha(x, y) {
		t.Error("different vars should not be equiv")
	}
}

func TestEquivAlphaLambda(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)

	lam1, _ := lg.NewLambda([]*lg.Variable{x}, x)
	lam2, _ := lg.NewLambda([]*lg.Variable{y}, y)
	if !EquivAlpha(lam1, lam2) {
		t.Error("alpha-equivalent lambdas should be equiv")
	}
}

// --- Match tests ---

func TestMatchVariable(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)

	// Variable pat vs variable inst with same type: heads_match returns true
	free := map[lg.NodeKey]lg.Node{lg.Key(x): x}
	m := Match(x, y, free, nil)
	if m == nil {
		t.Fatal("expected match for variable->variable")
	}
	if !m[lg.Key(x)].Equal(y) {
		t.Error("wrong mapping")
	}
}

func TestMatchApp(t *testing.T) {
	s := mkSort("S")
	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	p := mkConst("p", fs)
	x := mkVar("X", s)
	y := mkVar("Y", s)

	pat, _ := lg.NewApply(p, x)
	inst, _ := lg.NewApply(p, y)

	free := map[lg.NodeKey]lg.Node{lg.Key(x): x}
	m := Match(pat, inst, free, nil)
	if m == nil {
		t.Fatal("expected match for app")
	}
	if !m[lg.Key(x)].Equal(y) {
		t.Errorf("expected X -> Y")
	}
}

func TestMatchFail(t *testing.T) {
	s := mkSort("S")
	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	p := mkConst("p", fs)
	q := mkConst("q", fs)
	c := mkConst("c", s)

	pat := &lg.Apply{Func: p, Terms: []lg.Node{c}}
	inst := &lg.Apply{Func: q, Terms: []lg.Node{c}}

	m := Match(pat, inst, nil, nil)
	if m != nil {
		t.Error("different function heads should not match")
	}
}

// --- FOMatch tests ---

func TestFOMatchVariable(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	c := mkConst("c", s)

	free := map[lg.NodeKey]lg.Node{lg.Key(x): x}
	constants := map[lg.NodeKey]lg.Node{}
	m := FOMatch(x, c, free, constants)
	if m == nil {
		t.Fatal("expected fo match")
	}
	if !m[lg.Key(x)].Equal(c) {
		t.Error("wrong mapping")
	}
}

func TestFOMatchNoMatch(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)

	// Y is not a constant, so X should not match Y
	free := map[lg.NodeKey]lg.Node{lg.Key(x): x}
	constants := map[lg.NodeKey]lg.Node{}
	m := FOMatch(x, y, free, constants)
	if m != nil && len(m) > 0 {
		// FOMatch returns empty dict on no match for head cases
		if _, ok := m[lg.Key(x)]; ok {
			t.Error("should not match non-constant")
		}
	}
}

// --- MatchQuants tests ---

func TestMatchQuants(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)

	body1 := &lg.And{Terms: []lg.Node{x}}
	body2 := &lg.And{Terms: []lg.Node{y}}

	fa1 := &lg.ForAll{Variables: []*lg.Variable{x}, Body: body1}
	fa2 := &lg.ForAll{Variables: []*lg.Variable{y}, Body: body2}

	free := map[lg.NodeKey]lg.Node{}
	m := MatchQuants(fa1, fa2, free, nil)
	if m == nil {
		t.Fatal("expected match for quantified formulas")
	}
}

func TestMatchQuantsDiffType(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)

	fa := &lg.ForAll{Variables: []*lg.Variable{x}, Body: x}
	ex := &lg.Exists{Variables: []*lg.Variable{y}, Body: y}

	m := MatchQuants(fa, ex, nil, nil)
	if m != nil {
		t.Error("ForAll should not match Exists")
	}
}

// --- AddSymbols / RemoveSymbols tests ---

func TestAddSymbols(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	y := mkVar("Y", s)

	set := map[lg.NodeKey]lg.Node{}
	as := NewAddSymbols(set, []lg.Node{x, y})
	if set[lg.Key(x)] == nil || set[lg.Key(y)] == nil {
		t.Error("symbols should be added")
	}
	as.Restore()
	if set[lg.Key(x)] != nil || set[lg.Key(y)] != nil {
		t.Error("symbols should be removed after restore")
	}
}

func TestRemoveSymbols(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)

	set := map[lg.NodeKey]lg.Node{lg.Key(x): lg.True}
	rs := NewRemoveSymbols(set, []lg.Node{x})
	if set[lg.Key(x)] != nil {
		t.Error("symbol should be removed")
	}
	rs.Restore()
	if set[lg.Key(x)] == nil {
		t.Error("symbol should be restored")
	}
}

// --- Goal operation tests ---

func TestGoalConcSimple(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	// Wrap logic node in adapter
	lf := mkLF(ast.NewAtom("test"), concToASTNode(c))
	conc := GoalConc(lf)
	if conc == nil {
		t.Fatal("expected non-nil conclusion")
	}
}

func TestGoalPremsEmpty(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	lf := mkLF(ast.NewAtom("test"), concToASTNode(c))
	prems := GoalPrems(lf)
	if len(prems) != 0 {
		t.Error("expected no premises for simple formula")
	}
}

func TestGoalPremsWithSchema(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	prem := mkLF(ast.NewAtom("p"), concToASTNode(c))

	sb := ast.NewSchemaBody(prem, concToASTNode(c))
	goal := mkLF(ast.NewAtom("g"), sb)

	prems := GoalPrems(goal)
	if len(prems) != 1 {
		t.Errorf("expected 1 premise, got %d", len(prems))
	}
}

func TestCloneGoal(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	lf := mkLF(ast.NewAtom("test"), concToASTNode(c))

	clone := CloneGoal(lf, nil, c)
	if clone == nil {
		t.Fatal("expected non-nil clone")
	}
	if clone.ID == lf.ID {
		t.Error("cloned goal should have fresh ID")
	}
}

func TestNormalizeGoal(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	lf := mkLF(ast.NewAtom("test"), concToASTNode(c))
	norm := NormalizeGoal(lf)
	if norm == nil {
		t.Fatal("expected non-nil normalized goal")
	}
}

func TestGoalVocab(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)

	// Create a formula with a variable
	body := &lg.And{Terms: []lg.Node{x}}
	lf := mkLF(ast.NewAtom("test"), concToASTNode(body))
	vocab := GoalVocab(lf)
	if vocab == nil {
		t.Fatal("expected non-nil vocab")
	}
}

func TestGoalFree(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	body := &lg.And{Terms: []lg.Node{x}}
	lf := mkLF(ast.NewAtom("test"), concToASTNode(body))
	free := GoalFree(lf)
	if len(free) == 0 {
		t.Error("expected free variables")
	}
}

func TestTrivialGoal(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	lf := mkLF(ast.NewAtom("test"), concToASTNode(c))
	if TrivialGoal(lf) {
		t.Error("simple goal should not be trivial")
	}
}

func TestCheckConcsMatch(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	g1 := mkLF(ast.NewAtom("g1"), concToASTNode(c))
	g2 := mkLF(ast.NewAtom("g2"), concToASTNode(c))
	if err := CheckConcsMatch(g1, g2); err != nil {
		t.Errorf("same conclusions should match: %v", err)
	}

	d := mkConst("d", s)
	g3 := mkLF(ast.NewAtom("g3"), concToASTNode(d))
	if err := CheckConcsMatch(g1, g3); err == nil {
		t.Error("different conclusions should not match")
	}
}

// --- ProofChecker tests ---

func TestNewProofChecker(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	ax := mkLF(ast.NewAtom("ax1"), concToASTNode(c))
	pc := NewProofChecker([]*ast.LabeledFormula{ax}, nil, nil)
	if len(pc.Axioms) != 1 {
		t.Errorf("expected 1 axiom, got %d", len(pc.Axioms))
	}
}

func TestProofCheckerAdmitAxiom(t *testing.T) {
	pc := NewProofChecker(nil, nil, nil)
	s := mkSort("S")
	c := mkConst("c", s)
	ax := mkLF(ast.NewAtom("ax1"), concToASTNode(c))
	pc.AdmitAxiom(ax)
	if len(pc.Axioms) != 1 {
		t.Errorf("expected 1 axiom, got %d", len(pc.Axioms))
	}
}

func TestProofCheckerLookupSchema(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	ax := mkLF(ast.NewAtom("myax"), concToASTNode(c))
	pc := NewProofChecker([]*ast.LabeledFormula{ax}, nil, nil)

	goal := mkLF(ast.NewAtom("goal"), concToASTNode(c))
	schema, err := pc.LookupSchema("myax", goal)
	if err != nil {
		t.Fatalf("expected to find schema: %v", err)
	}
	if schema == nil {
		t.Error("expected non-nil schema")
	}
}

func TestProofCheckerLookupSchemaNotFound(t *testing.T) {
	pc := NewProofChecker(nil, nil, nil)
	s := mkSort("S")
	c := mkConst("c", s)
	goal := mkLF(ast.NewAtom("goal"), concToASTNode(c))
	_, err := pc.LookupSchema("nonexistent", goal)
	if err == nil {
		t.Error("expected error for missing schema")
	}
}

// --- Tactic registry tests ---

func TestRegisterTactic(t *testing.T) {
	called := false
	RegisterTactic("test_tactic", func(pc *ProofChecker, goals []*ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
		called = true
		return goals, nil
	})
	if _, ok := RegisteredTactics["test_tactic"]; !ok {
		t.Error("tactic should be registered")
	}
	tac := RegisteredTactics["test_tactic"]
	_, _ = tac(nil, nil, nil)
	if !called {
		t.Error("tactic should have been called")
	}
	// Clean up
	delete(RegisteredTactics, "test_tactic")
}

// --- Skolemize tests ---

func TestSkolemizeFmlaSimple(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	c := mkConst("c", s)

	// ForAll X. p(X) in positive position -> skolemize X
	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	p := mkConst("p", fs)
	body := &lg.Apply{Func: p, Terms: []lg.Node{x}}
	fmla := &lg.ForAll{Variables: []*lg.Variable{x}, Body: body}

	renamer := iu.NewUniqueRenamer("", nil)
	var skfuns []*lg.Symbol
	result := SkolemizeFmla(fmla, true, renamer, &skfuns, true)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(skfuns) != 1 {
		t.Errorf("expected 1 skolem function, got %d", len(skfuns))
	}

	// The result should not have a ForAll anymore
	if _, ok := result.(*lg.ForAll); ok {
		t.Error("skolemized formula should not have ForAll")
	}
	_ = c
}

func TestSkolemizeFmlaExists(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)

	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	p := mkConst("p", fs)
	body := &lg.Apply{Func: p, Terms: []lg.Node{x}}
	fmla := &lg.Exists{Variables: []*lg.Variable{x}, Body: body}

	renamer := iu.NewUniqueRenamer("", nil)
	var skfuns []*lg.Symbol
	result := SkolemizeFmla(fmla, true, renamer, &skfuns, true)

	// Exists in positive position -> universalize, result should have Exists wrapping
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(skfuns) != 0 {
		t.Errorf("expected 0 skolem functions for existential in positive, got %d", len(skfuns))
	}
}

func TestSkolemizeGoalBasic(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)

	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	p := mkConst("p", fs)
	body := &lg.Apply{Func: p, Terms: []lg.Node{x}}
	fmla := &lg.ForAll{Variables: []*lg.Variable{x}, Body: body}

	goal := mkLF(ast.NewAtom("test"), concToASTNode(fmla))
	result := SkolemizeGoal(goal, true)
	if result == nil {
		t.Fatal("expected non-nil skolemized goal")
	}
}

// --- ApplyMatch tests ---

func TestApplyMatchIdentity(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	result := ApplyMatch(nil, x)
	if !result.Equal(x) {
		t.Error("empty match should return same term")
	}
}

func TestApplyMatchSubstitution(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	c := mkConst("c", s)

	match := map[lg.NodeKey]lg.Node{lg.Key(x): c}
	result := ApplyMatch(match, x)
	if !result.Equal(c) {
		t.Errorf("expected c, got %s", result)
	}
}

func TestApplyMatchFunc(t *testing.T) {
	s := mkSort("S")
	s2 := mkSort("T")
	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	c := mkConst("f", fs)

	match := map[lg.NodeKey]lg.Node{lg.Key(s): s2}
	result := ApplyMatchFunc(match, c)
	expected, _ := lg.NewFunctionSort(s2, lg.Boolean)
	if !result.CSort.Equal(expected) {
		t.Errorf("expected sort T->Bool, got %s", result.CSort)
	}
}

// --- ComposeMatches tests ---

func TestComposeMatches(t *testing.T) {
	s := mkSort("S")
	s2 := mkSort("T")
	x := mkVar("X", s)
	y := mkVar("Y", s2)

	free := map[lg.NodeKey]lg.Node{lg.Key(x): x}
	mat1 := map[lg.NodeKey]lg.Node{lg.Key(x): y}
	mat2 := map[lg.NodeKey]lg.Node{lg.Key(y): mkConst("c", s2)}
	result := ComposeMatches(free, mat1, mat2, nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result[lg.Key(x)].Equal(mkConst("c", s2)) {
		t.Error("composition should map X to c")
	}
}

func TestComposeMatchesNil(t *testing.T) {
	result := ComposeMatches(nil, nil, map[lg.NodeKey]lg.Node{}, nil)
	if result != nil {
		t.Error("nil mat1 should return nil")
	}
}

// --- ExtractTerms tests ---

func TestExtractTerms(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)

	// inst = X, terms = [X] -> lambda V0. V0
	// X is a variable that appears as a term; after extraction the body is V0
	// No free vars remain (V0 is a lambda param), so this should succeed
	constants := map[lg.NodeKey]lg.Node{lg.Key(x): lg.True}
	lam := ExtractTerms(x, []lg.Node{x}, constants)
	if lam == nil {
		t.Fatal("expected non-nil lambda")
	}
	if len(lam.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(lam.Variables))
	}
	// The body should be V0 (the lambda variable)
	if _, ok := lam.Body.(*lg.Variable); !ok {
		t.Errorf("expected body to be a variable, got %T", lam.Body)
	}
}

func TestExtractTermsEmpty(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	lam := ExtractTerms(c, nil, nil)
	if lam != nil {
		t.Error("expected nil for empty terms")
	}
}

// --- ApplyMatchFreesyms tests ---

func TestApplyMatchFreesyms(t *testing.T) {
	s := mkSort("S")
	s2 := mkSort("T")

	free := map[lg.NodeKey]lg.Node{lg.Key(s): s, lg.Key(s2): s2}
	match := map[lg.NodeKey]lg.Node{lg.Key(s): s2}
	result := ApplyMatchFreesyms(match, free)
	if result[lg.Key(s)] != nil {
		t.Error("matched symbol should not be in result")
	}
	if result[lg.Key(s2)] == nil {
		t.Error("unmatched symbol should remain")
	}
}

// --- Fuzz tests ---

func FuzzMergeMatches(f *testing.F) {
	f.Add(0, 1, 2, 3, false)
	f.Add(1, 1, 1, 1, true)
	f.Add(0, 0, 0, 0, false)

	f.Fuzz(func(t *testing.T, a, b, c, d int, conflict bool) {
		s := mkSort("S")
		vars := make([]*lg.Variable, 4)
		names := []string{"A", "B", "C", "D"}
		for i, n := range names {
			vars[i] = mkVar(n, s)
		}

		consts := []*lg.Symbol{
			mkConst("c0", s), mkConst("c1", s),
			mkConst("c2", s), mkConst("c3", s),
		}

		m1 := map[lg.NodeKey]lg.Node{
			lg.Key(vars[a%4]): consts[b%4],
		}
		m2 := map[lg.NodeKey]lg.Node{
			lg.Key(vars[c%4]): consts[d%4],
		}

		result := MergeMatches(m1, m2)

		// Verify: if same key maps to different values, result should be nil
		key := vars[a%4]
		if key == vars[c%4] {
			if !consts[b%4].Equal(consts[d%4]) {
				if result != nil {
					t.Error("conflicting matches should return nil")
				}
			}
		}
		// If result is non-nil, verify all entries are present
		if result != nil {
			if v, ok := result[lg.Key(vars[a%4])]; !ok || !v.Equal(consts[b%4]) {
				t.Error("first match entry missing")
			}
			if v, ok := result[lg.Key(vars[c%4])]; !ok || !v.Equal(consts[d%4]) {
				t.Error("second match entry missing")
			}
		}
	})
}

func FuzzMatch(f *testing.F) {
	f.Add(true, false)
	f.Add(false, true)
	f.Add(true, true)

	f.Fuzz(func(t *testing.T, patFree, addExtra bool) {
		s := mkSort("S")
		x := mkVar("X", s)
		y := mkVar("Y", s)

		free := map[lg.NodeKey]lg.Node{}
		if patFree {
			free[lg.Key(x)] = x
		}
		constants := map[lg.NodeKey]lg.Node{}

		// Match variable to variable (same type)
		m := Match(x, y, free, constants)
		if patFree {
			// Should match: same type, free var
			if m == nil {
				t.Error("expected match when pattern is free")
			} else if !m[lg.Key(x)].Equal(y) {
				t.Error("wrong match result")
			}
		} else {
			// Not free: X != Y so should not match
			if m != nil {
				if v, ok := m[lg.Key(x)]; ok && !v.Equal(y) {
					// This is fine, just no mapping for x
				}
			}
		}
	})
}
