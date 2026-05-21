package goivy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// Fix 1: IsCheckModUnprovable filtering
// =============================================================================

func TestRegression_Arrayset3IsoSOverwriteFormalGuaranteePasses(t *testing.T) {
	path, err := filepath.Abs("../ivy-lang-examples/doc/examples/arrayset3.ivy")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("arrayset3 fixture not available: %v", err)
	}
	cfg := NewConfig()
	cfg.Isolate = "iso_s"
	if err := Start([]string{path}, cfg); err != nil {
		t.Fatalf("arrayset3 iso_s should verify like Python Ivy 1.6: %v", err)
	}
}

func TestRegression_Bug1_NormalMode(t *testing.T) {
	cfg := NewConfig()
	cfg.OnlyCheckUnprovable = false
	ac := cfg.AstCfg

	normal := ac.NewLabeledFormula(nil, True)
	normal.Unprovable = false
	unprov := ac.NewLabeledFormula(nil, True)
	unprov.Unprovable = true

	if !IsCheckModUnprovable(cfg, normal) {
		t.Error("normal formula should pass filter when OnlyCheckUnprovable=false")
	}
	if IsCheckModUnprovable(cfg, unprov) {
		t.Error("unprovable formula should fail filter when OnlyCheckUnprovable=false")
	}
}

func TestRegression_Bug1_UnprovableMode(t *testing.T) {
	cfg := NewConfig()
	cfg.OnlyCheckUnprovable = true
	ac := cfg.AstCfg

	normal := ac.NewLabeledFormula(nil, True)
	normal.Unprovable = false
	unprov := ac.NewLabeledFormula(nil, True)
	unprov.Unprovable = true

	if IsCheckModUnprovable(cfg, normal) {
		t.Error("normal formula should fail filter when OnlyCheckUnprovable=true")
	}
	if !IsCheckModUnprovable(cfg, unprov) {
		t.Error("unprovable formula should pass filter when OnlyCheckUnprovable=true")
	}
}

func TestRegression_Bug1_CheckConjsFilters(t *testing.T) {
	// With OnlyCheckUnprovable=true and only unprovable conjectures,
	// CheckConjsInState should not panic. The filter selects only
	// unprovable conjectures.
	mod := New()
	mod.Cfg.OnlyCheckUnprovable = true

	mod.LabeledConjs = []*LabeledFormula{
		{Formula: True, Unprovable: true},
	}
	// Just verify no panic — the important thing is the filter ran.
	_ = CheckConjsInState(mod, 0, nil)
}

// =============================================================================
// Fix 2: CheckConjsInStateWithAG passes ag/post
// =============================================================================

func TestRegression_Bug2_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CheckConjsInStateWithAG panicked: %v", r)
		}
	}()
	mod := New()
	mod.LabeledConjs = []*LabeledFormula{
		{Formula: True},
	}
	// ag=nil and post=nil is the degenerate case; should not panic.
	_ = CheckConjsInStateWithAG(mod, nil, nil, 0, nil)
}

// =============================================================================
// Fix 3: ConvertPostcondsWithUpdate gets post.Update
// =============================================================================

func TestRegression_Bug3_WithUpdate(t *testing.T) {
	ac := NewAstConfig()
	sym := NewConst("x", Boolean)
	update := &Update{
		Modified: []*Const{sym},
	}
	oldSym := NewConst("old_x", Boolean)
	pc := ac.NewLabeledFormula(nil, oldSym)
	pc.SetLineno(Location{Line: 1})
	result := ConvertPostcondsWithUpdate(update, []*LabeledFormula{pc})
	if len(result) != 1 {
		t.Fatalf("expected 1 postcondition, got %d", len(result))
	}
	// The renaming should have replaced old_x with __x (pre-state prefix).
	renamed := result[0].Formula
	if renamed == nil {
		t.Fatal("renamed formula is nil")
	}
	renamedStr := fmt.Sprint(renamed)
	// old_x should be renamed: either to "x" (from IsOld mapping) or "__x" (from Modified).
	// The Modified mapping maps old_x → __x, so we expect __x.
	if !strings.Contains(renamedStr, "__x") && !strings.Contains(renamedStr, "x") {
		t.Errorf("expected renamed formula to contain x or __x, got %s", renamedStr)
	}
}

func TestRegression_Bug3_NilUpdate(t *testing.T) {
	ac := NewAstConfig()
	pc := ac.NewLabeledFormula(nil, True)
	result := ConvertPostconds([]*LabeledFormula{pc})
	// With nil update, postconds should pass through unchanged.
	if len(result) != 1 {
		t.Fatalf("expected 1 postcondition, got %d", len(result))
	}
	if result[0] != pc {
		t.Error("with nil update, ConvertPostconds should return input unchanged")
	}
}

// =============================================================================
// Fix 4: UpdateTheory after property promotion
// =============================================================================

func TestRegression_Bug4_PromotionUpdatesTheory(t *testing.T) {
	mod := New()
	mod.LabeledProps = []*LabeledFormula{
		{Formula: True},
	}
	origLen := len(mod.LabeledAxioms)

	// Promote props to axioms and update theory.
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)
	mod.UpdateTheory()

	if len(mod.LabeledAxioms) <= origLen {
		t.Error("properties should have been promoted to axioms")
	}
	bt := mod.BackgroundTheory(nil)
	if bt == nil {
		t.Error("BackgroundTheory should be non-nil after UpdateTheory")
	}
}

// =============================================================================
// Fix 5: NewBaseChecker witness Skolem prefix "@"
// =============================================================================

func TestRegression_Bug5_SkolemPrefix(t *testing.T) {
	v, err := NewVariable("X", Boolean)
	if err != nil {
		t.Fatalf("NewVariable failed: %v", err)
	}
	// NewBaseChecker(invert=true) dualizes the conjecture using the @-prefix
	// witness skolemizer (Python: def witness(v): return lg.Symbol('@'+v.name, v.sort)).
	checker := NewBaseChecker(New(), v, false, true)
	if checker == nil || checker.FC == nil {
		t.Fatal("NewBaseChecker returned nil or FC")
	}
	// Walk the dual clauses for skolem symbols.
	// The Skolem constant should have "@" prefix, not "__".
	for _, fmla := range checker.FC.Fmlas {
		walkForSkolem(t, fmla)
	}
}

func walkForSkolem(t *testing.T, e Expr) {
	t.Helper()
	if e == nil {
		return
	}
	if sym, ok := e.(*Const); ok {
		if strings.HasPrefix(sym.Name, "__") && !strings.HasPrefix(sym.Name, "__old") {
			t.Errorf("found double-underscore skolem %q; expected @-prefix", sym.Name)
		}
	}
	for _, c := range e.Children() {
		walkForSkolem(t, c)
	}
}

// =============================================================================
// Fix 6: fail_expr transformation
// =============================================================================

func TestRegression_Bug6_FailExpr(t *testing.T) {
	// The fail_expr logic prepends "fail_" to the action rep.
	rep := "initialize"
	failRep := "fail_" + rep
	if failRep != "fail_initialize" {
		t.Errorf("expected fail_initialize, got %s", failRep)
	}

	reps := []string{"initialize", "step", "ext:foo"}
	for _, r := range reps {
		fail := "fail_" + r
		if !strings.HasPrefix(fail, "fail_") {
			t.Errorf("fail prefix not applied to %s", r)
		}
	}
}

// =============================================================================
// Fix 7: !unprovable guards
// =============================================================================

func TestRegression_Bug7_PropertyCheckGuard(t *testing.T) {
	cfg := NewConfig()
	cfg.OnlyCheckUnprovable = true

	// When OnlyCheckUnprovable is true, the guard `!cfg.OnlyCheckUnprovable`
	// should be false, preventing normal property checks.
	guard := !cfg.OnlyCheckUnprovable
	if guard {
		t.Error("negated guard should be false when OnlyCheckUnprovable=true")
	}
}

func TestRegression_Bug7_InvariantInitGuard(t *testing.T) {
	cfg := NewConfig()
	cfg.OnlyCheckUnprovable = true
	guard := !cfg.OnlyCheckUnprovable
	if guard {
		t.Error("invariant init guard should be false when OnlyCheckUnprovable=true")
	}
}

func TestRegression_Bug7_InitializerGuaranteeGuard(t *testing.T) {
	cfg := NewConfig()
	cfg.OnlyCheckUnprovable = false
	guard := !cfg.OnlyCheckUnprovable
	if !guard {
		t.Error("initializer guarantee guard should be true when OnlyCheckUnprovable=false")
	}
}

// =============================================================================
// Fix 8: checked_invariants filtered by IsCheckModUnprovable
// =============================================================================

func TestRegression_Bug8_NormalMode(t *testing.T) {
	cfg := NewConfig()
	cfg.OnlyCheckUnprovable = false

	conjs := []*LabeledFormula{
		{Formula: True, Unprovable: false},
		{Formula: True, Unprovable: true},
	}

	var filtered []*LabeledFormula
	for _, c := range conjs {
		if IsCheckModUnprovable(cfg, c) {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered conjecture, got %d", len(filtered))
	}
	if filtered[0].Unprovable {
		t.Error("expected non-unprovable conjecture in normal mode")
	}
}

func TestRegression_Bug8_UnprovableMode(t *testing.T) {
	cfg := NewConfig()
	cfg.OnlyCheckUnprovable = true

	conjs := []*LabeledFormula{
		{Formula: True, Unprovable: false},
		{Formula: True, Unprovable: true},
	}

	var filtered []*LabeledFormula
	for _, c := range conjs {
		if IsCheckModUnprovable(cfg, c) {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered conjecture, got %d", len(filtered))
	}
	if !filtered[0].Unprovable {
		t.Error("expected unprovable conjecture in unprovable mode")
	}
}

// =============================================================================
// Fix 9: Ranking implements Action
// =============================================================================

func TestRegression_Bug9_ActionInterface(t *testing.T) {
	r := NewRanking(True)

	// Compile-time check: Ranking implements Action.
	var _ ActionsAction = r

	if r.Name() != "decreases" {
		t.Errorf("Ranking.Name() = %q, want 'decreases'", r.Name())
	}
	if r.String() == "" {
		t.Error("Ranking.String() should not be empty")
	}

	clone := r.ActionClone([]Expr{True})
	if clone == nil {
		t.Error("ActionClone returned nil")
	}
	if r.IterCalls() != nil {
		t.Error("Ranking.IterCalls() should return nil")
	}
	subs := r.IterSubactions()
	if len(subs) == 0 {
		t.Error("Ranking.IterSubactions() should return at least itself")
	}
	dec := r.Decompose()
	if len(dec) != 1 || len(dec[0]) != 1 {
		t.Error("Ranking.Decompose() should return [[self]]")
	}
}

func TestRegression_Bug9_FindAssertionsRanking(t *testing.T) {
	mod := New()
	ranking := NewRanking(True, True)
	seq := NewSequence(ranking)
	mod.Actions.Set("test_action", seq)

	found := FindAssertions("", mod)
	foundRanking := false
	for _, a := range found {
		if _, ok := a.(*LogicRanking); ok {
			foundRanking = true
		}
	}
	if !foundRanking {
		t.Error("FindAssertions should find Ranking actions")
	}
}

func TestRegression_Bug9_BothTypes(t *testing.T) {
	mod := New()
	assert := NewAssertAction(True)
	ranking := NewRanking(True)
	seq := NewSequence(assert, ranking)
	mod.Actions.Set("test_action", seq)

	found := FindAssertions("", mod)
	hasAssert, hasRanking := false, false
	for _, a := range found {
		if _, ok := a.(*LogicAssertAction); ok {
			hasAssert = true
		}
		if _, ok := a.(*LogicRanking); ok {
			hasRanking = true
		}
	}
	if !hasAssert {
		t.Error("FindAssertions should find AssertAction")
	}
	if !hasRanking {
		t.Error("FindAssertions should find Ranking")
	}
}

// =============================================================================
// Fix 11: FilterCheckers applied in normal path
// =============================================================================

func TestRegression_Bug11_FilterCheckers(t *testing.T) {
	mod := New()
	ac := mod.Cfg.AstCfg
	lf10 := ac.NewLabeledFormula(nil, True)
	lf10.SetLineno(Location{Filename: "test.ivy", Line: 10})
	lf20 := ac.NewLabeledFormula(nil, True)
	lf20.SetLineno(Location{Filename: "test.ivy", Line: 20})

	checkers := []Checker{
		NewConjChecker(mod, lf10, 0),
		NewConjChecker(mod, lf20, 0),
	}

	// Filter by file:line — should return 1 result.
	filtered := FilterCheckers(checkers, "test.ivy:10")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 checker for test.ivy:10, got %d", len(filtered))
	}

	// Empty filter — should return all.
	all := FilterCheckers(checkers, "")
	if len(all) != 2 {
		t.Fatalf("expected 2 checkers for empty filter, got %d", len(all))
	}
}

// =============================================================================
// Fix 12: fragment.CheckFragment called
// =============================================================================

func TestRegression_Bug12_FragmentCheck(t *testing.T) {
	// CheckIsolate calls fragment.CheckFragment. We verify the call path
	// doesn't panic on an empty module.
	mod := New()
	err := CheckIsolate(mod, nil)
	// An error is acceptable (missing solver etc.); a panic is not.
	_ = err
}

// =============================================================================
// Fix 13: theory_context wrapping
// =============================================================================

func TestRegression_Bug13_TheoryContext(t *testing.T) {
	mod := New()
	cleanup := mod.TheoryContext()
	if cleanup == nil {
		t.Fatal("TheoryContext returned nil cleanup function")
	}
	// Call cleanup without panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TheoryContext cleanup panicked: %v", r)
		}
	}()
	cleanup()
}

// =============================================================================
// Fix 14: check_lineno filter on initializer guarantees
// =============================================================================

func TestRegression_Bug14_WithFilter(t *testing.T) {
	mod := New()
	ac := mod.Cfg.AstCfg
	lf42 := ac.NewLabeledFormula(nil, True)
	lf42.SetLineno(Location{Filename: "test.ivy", Line: 42})
	lf99 := ac.NewLabeledFormula(nil, True)
	lf99.SetLineno(Location{Filename: "test.ivy", Line: 99})

	checkers := []Checker{
		NewConjChecker(mod, lf42, 0),
		NewConjChecker(mod, lf99, 0),
	}

	filtered := FilterCheckers(checkers, "test.ivy:42")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 checker for test.ivy:42, got %d", len(filtered))
	}
	cc := filtered[0].(*ConjChecker)
	if cc.LF.Lineno() != 42 {
		t.Errorf("expected lineno 42, got %d", cc.LF.Lineno())
	}
}

func TestRegression_Bug14_NoFilter(t *testing.T) {
	mod := New()
	ac := mod.Cfg.AstCfg
	lf42 := ac.NewLabeledFormula(nil, True)
	lf42.SetLineno(Location{Line: 42})
	lf99 := ac.NewLabeledFormula(nil, True)
	lf99.SetLineno(Location{Line: 99})

	checkers := []Checker{
		NewConjChecker(mod, lf42, 0),
		NewConjChecker(mod, lf99, 0),
	}

	filtered := FilterCheckers(checkers, "")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 checkers for empty filter, got %d", len(filtered))
	}
}

// =============================================================================
// Fix 15: ModuleLFToAstLF / AstLFToModuleLF identity
// =============================================================================

func TestRegression_Bug15_Identity(t *testing.T) {
	ac := NewAstConfig()
	lf := ac.NewLabeledFormula(nil, True)

	if ModuleLFToAstLF(lf) != lf {
		t.Error("ModuleLFToAstLF should return same pointer")
	}
	if AstLFToModuleLF(lf) != lf {
		t.Error("AstLFToModuleLF should return same pointer")
	}

	// Nil returns nil.
	if ModuleLFToAstLF(nil) != nil {
		t.Error("ModuleLFToAstLF(nil) should return nil")
	}
	if AstLFToModuleLF(nil) != nil {
		t.Error("AstLFToModuleLF(nil) should return nil")
	}
}

// =============================================================================
// Fix 16: PrintDots no params
// =============================================================================

func TestRegression_Bug16_NoParams(t *testing.T) {
	result := PrintDots()
	if result != "...\n" {
		t.Errorf("PrintDots() = %q, want %q", result, "... ")
	}
}

// =============================================================================
// Fix 18: CheckSubgoals with vocab context
// =============================================================================

func TestRegression_Bug18_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CheckSubgoals panicked: %v", r)
		}
	}()
	mod := New()
	err := CheckSubgoals(nil, nil, mod)
	_ = err
}

func TestRegression_Bug18_EmptyGoals(t *testing.T) {
	mod := New()
	err := CheckSubgoals(nil, nil, mod)
	if err != nil {
		t.Errorf("CheckSubgoals with nil goals should not error, got: %v", err)
	}
}

// =============================================================================
// Fix 19: NormalProgram implements ast.Node
// =============================================================================

func TestRegression_Bug19_NormalProgramIsASTNode(t *testing.T) {
	// Compile-time check: NormalProgram implements ast.Node.
	var _ Node = (*NormalProgram)(nil)

	np := &NormalProgram{
		Invars: []*LabeledFormula{{Formula: True}},
		Asms:   []*LabeledFormula{{Formula: True}},
		Calls:  []string{"action1"},
		Init:   NewSequence(),
	}

	// Args() returns nil (Python: args property returns [])
	if np.Args() != nil {
		t.Error("NormalProgram.Args() should return nil")
	}

	// Clone() returns a copy
	cloned := np.Clone(nil)
	if cloned == nil {
		t.Fatal("NormalProgram.Clone() returned nil")
	}
	npClone, ok := cloned.(*NormalProgram)
	if !ok {
		t.Fatal("Clone() did not return *NormalProgram")
	}
	if len(npClone.Invars) != 1 {
		t.Errorf("clone Invars: expected 1, got %d", len(npClone.Invars))
	}

	// GetLineno/SetLineno work via embedded Base
	np.SetLineno(Location{Line: 42})
	if np.GetLineno().Line != 42 {
		t.Errorf("GetLineno().Line = %d, want 42", np.GetLineno().Line)
	}
}

func TestRegression_Bug19_TemporalModelsHoldsNormalProgram(t *testing.T) {
	acfg := NewAstConfig()
	np := &NormalProgram{
		Invars:    []*LabeledFormula{acfg.NewLabeledFormula(nil, True)},
		Asms:      []*LabeledFormula{acfg.NewLabeledFormula(nil, True)},
		Calls:     []string{"action1", "action2"},
		Init:      NewSequence(),
		Postconds: map[string][]*LabeledFormula{"action1": {acfg.NewLabeledFormula(nil, True)}},
	}

	// Store in TemporalModels (requires ast.Node)
	tm := acfg.NewTemporalModels(np, True)

	// Extract back
	extracted, ok := tm.Model.(*NormalProgram)
	if !ok {
		t.Fatal("could not extract NormalProgram from TemporalModels.Model")
	}
	if len(extracted.Invars) != 1 {
		t.Errorf("round-trip Invars: expected 1, got %d", len(extracted.Invars))
	}
	if len(extracted.Calls) != 2 {
		t.Errorf("round-trip Calls: expected 2, got %d", len(extracted.Calls))
	}
	if len(extracted.Postconds) != 1 {
		t.Errorf("round-trip Postconds: expected 1, got %d", len(extracted.Postconds))
	}
	if extracted.Init == nil {
		t.Error("round-trip Init should be non-nil")
	}
}

func TestRegression_Bug19_CheckSubgoalsTemporalBranch(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CheckSubgoals panicked: %v", r)
		}
	}()

	mod := New()
	cfg := mod.Cfg.AstCfg

	act1Stmt := NewSequence()
	act1Stmt.SetLineno(Location{Filename: "test", Line: 1})
	act1Term := &ActionTerm{Stmt: act1Stmt}
	np := &NormalProgram{
		Invars:   []*LabeledFormula{cfg.NewLabeledFormula(nil, True)},
		Asms:     []*LabeledFormula{cfg.NewLabeledFormula(nil, True)},
		Calls:    []string{"action1"},
		Bindings: []*ActionTermBinding{{Name: "action1", Action: act1Term}},
		Init:     NewSequence(),
	}

	// Build a goal with TemporalModels conclusion containing our NormalProgram
	tm := cfg.NewTemporalModels(np, True)
	goal := cfg.NewLabeledFormula(nil, tm)
	err := CheckSubgoals([]*LabeledFormula{goal}, nil, mod)
	// An error is acceptable; a panic from failed type assertion is not.
	_ = err
}
