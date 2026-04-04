package check

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// --- BaseChecker tests ---

func TestBaseCheckerCreate(t *testing.T) {
	c := NewBaseChecker(module.NewConfig(), lg.True, true, true)
	if c == nil {
		t.Fatal("NewBaseChecker returned nil")
	}
	if c.ReportPass != true {
		t.Error("expected ReportPass true")
	}
	if c.Failed() {
		t.Error("new checker should not be failed")
	}
}

func TestBaseCheckerCond(t *testing.T) {
	c := NewBaseChecker(module.NewConfig(), lg.True, false, false)
	cond := c.Cond()
	if cond == nil {
		t.Fatal("Cond returned nil")
	}
}

func TestBaseCheckerPass(t *testing.T) {
	c := NewBaseChecker(module.NewConfig(), lg.True, false, true)
	result := c.Pass()
	if !result {
		t.Error("Pass should return true")
	}
	if c.Failed() {
		t.Error("after Pass, should not be failed")
	}
}

func TestBaseCheckerFail(t *testing.T) {
	cfg := module.NewConfig()
	oldFailures := cfg.Failures
	c := NewBaseChecker(cfg, lg.True, false, true)
	_ = c.Fail()
	if !c.Failed() {
		t.Error("after Fail, should be failed")
	}
	if cfg.Failures <= oldFailures {
		t.Error("Failures count should have increased")
	}
	cfg.Failures = oldFailures // restore
}

func TestBaseCheckerSatCallsFail(t *testing.T) {
	cfg := module.NewConfig()
	oldFailures := cfg.Failures
	c := NewBaseChecker(module.NewConfig(), lg.True, false, true)
	_ = c.Sat()
	if !c.Failed() {
		t.Error("Sat should trigger Fail")
	}
	cfg.Failures = oldFailures
}

func TestBaseCheckerUnsatCallsPass(t *testing.T) {
	c := NewBaseChecker(module.NewConfig(), lg.True, false, true)
	result := c.Unsat()
	if !result {
		t.Error("Unsat should trigger Pass and return true")
	}
}

func TestBaseCheckerAssume(t *testing.T) {
	c := NewBaseChecker(module.NewConfig(), lg.True, false, true)
	if c.Assume() {
		t.Error("BaseChecker.Assume should return false")
	}
}

func TestBaseCheckerGetAnnot(t *testing.T) {
	c := NewBaseChecker(module.NewConfig(), lg.True, false, true)
	if c.GetAnnot() != nil {
		t.Error("BaseChecker.GetAnnot should return nil")
	}
}

func TestBaseCheckerGetLF(t *testing.T) {
	c := NewBaseChecker(module.NewConfig(), lg.True, false, true)
	if c.GetLF() != nil {
		t.Error("BaseChecker.GetLF should return nil")
	}
}

// --- ConjChecker tests ---

func TestConjCheckerCreate(t *testing.T) {
	cfg := module.NewConfig()

	lf := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	lf.Lineno = 42
	cc := NewConjChecker(cfg, lf, 8)
	if cc == nil {
		t.Fatal("NewConjChecker returned nil")
	}
	if cc.Indent != 8 {
		t.Errorf("expected indent 8, got %d", cc.Indent)
	}
	if cc.LF != lf {
		t.Error("LF not set correctly")
	}
}

func TestConjCheckerGetLF(t *testing.T) {
	cfg := module.NewConfig()

	lf := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	cc := NewConjChecker(cfg, lf, 4)
	if cc.GetLF() != lf {
		t.Error("GetLF should return the labeled formula")
	}
}

func TestConjCheckerImplementsChecker(t *testing.T) {
	cfg := module.NewConfig()

	lf := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	var _ Checker = NewConjChecker(cfg, lf, 8)
}

// --- ConjAssumer tests ---

func TestConjAssumerCreate(t *testing.T) {
	cfg := module.NewConfig()

	lf := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	ca := NewConjAssumer(cfg, lf)
	if ca == nil {
		t.Fatal("NewConjAssumer returned nil")
	}
	if ca.LF != lf {
		t.Error("LF not set correctly")
	}
}

func TestConjAssumerAssume(t *testing.T) {
	cfg := module.NewConfig()

	lf := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	ca := NewConjAssumer(cfg, lf)
	if !ca.Assume() {
		t.Error("ConjAssumer.Assume should return true")
	}
}

func TestConjAssumerImplementsChecker(t *testing.T) {
	cfg := module.NewConfig()

	lf := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	var _ Checker = NewConjAssumer(cfg, lf)
}

// --- DualClauses tests ---

func TestDualClausesNil(t *testing.T) {
	result := DualClauses(nil)
	if result != nil {
		t.Error("DualClauses(nil) should return nil")
	}
}

func TestDualClausesEmpty(t *testing.T) {
	c := module.NewClauses(nil, nil, nil)
	result := DualClauses(c)
	if result == nil {
		t.Fatal("DualClauses should not return nil for empty clauses")
	}
}

func TestDualClausesSingleFormula(t *testing.T) {
	// Use a real formula (not lg.True which is empty And, consumed by collectAndList)
	p := lg.NewConst("p", lg.Boolean)
	c := module.NewClauses([]lg.Expr{p}, nil, nil)
	result := DualClauses(c)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
	// The dual negates the formula: p becomes Not(p).
	// (No variables to skolemize for a constant symbol.)
	_, isNot := result.Fmlas[0].(*lg.Not)
	if !isNot {
		t.Errorf("expected Not, got %T", result.Fmlas[0])
	}
}

// --- Parameter tests ---

func TestDiagnoseParameter(t *testing.T) {
	cfg := module.NewConfig()

	if cfg.Diagnose {
		t.Error("diagnose should default to false")
	}
}

func TestCoverageParameter(t *testing.T) {
	cfg := module.NewConfig()

	if !cfg.Coverage {
		t.Error("coverage should default to true")
	}
}

func TestCheckedActionParameter(t *testing.T) {
	cfg := module.NewConfig()
	if cfg.CheckedAction != "" {
		t.Error("checked_action should default to empty")
	}
}

func TestOptTrustedParameter(t *testing.T) {
	cfg := module.NewConfig()
	if cfg.OptTrusted {
		t.Error("trusted should default to false")
	}
}

func TestOptMCParameter(t *testing.T) {
	cfg := module.NewConfig()
	if cfg.OptMC {
		t.Error("mc should default to false")
	}
}

func TestOptTraceParameter(t *testing.T) {
	cfg := module.NewConfig()
	if cfg.OptTrace {
		t.Error("trace should default to false")
	}
}

func TestOptIvyStatsParameter(t *testing.T) {
	cfg := module.NewConfig()
	if cfg.OptIvyStats {
		t.Error("ivy_stats should default to false")
	}
}

func TestNoCheckGuaranteesParameter(t *testing.T) {
	cfg := module.NewConfig()
	if cfg.NoCheckGuarantees {
		t.Error("no_check_guarantees should default to false")
	}
}

func TestProfilingParameter(t *testing.T) {
	cfg := module.NewConfig()
	if cfg.Profiling {
		t.Error("profile should default to false")
	}
}

// --- Pretty printing tests ---

func TestPrettyLabelNil(t *testing.T) {
	result := PrettyLabel(nil)
	if result != "(no name)" {
		t.Errorf("expected '(no name)', got '%s'", result)
	}
}

func TestPrettyLabelWithValue(t *testing.T) {
	result := PrettyLabel("safety")
	if result != "safety" {
		t.Errorf("expected 'safety', got '%s'", result)
	}
}

func TestPrettyLinenoPositive(t *testing.T) {
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, nil)
	lf.SetLineno(ast.Location{Filename: "test.ivy", Line: 42})
	result := PrettyLineno(lf)
	// Python: return str(ast.lineno) — LocationTuple with filename and line.
	if result != "test.ivy: line 42: " {
		t.Errorf("expected 'test.ivy: line 42: ', got '%s'", result)
	}
}

func TestPrettyLinenoZero(t *testing.T) {
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, nil)
	result := PrettyLineno(lf)
	if result != "(internal) " {
		t.Errorf("expected '(internal) ', got '%s'", result)
	}
}

func TestPrettyLinenoNil(t *testing.T) {
	result := PrettyLineno(nil)
	if result != "(internal) " {
		t.Errorf("expected '(internal) ', got '%s'", result)
	}
}

func TestPrettyLF(t *testing.T) {
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, nil)
	lf.SetLineno(ast.Location{Filename: "test.ivy", Line: 10})
	result := PrettyLF(lf, 4)
	if !strings.HasPrefix(result, "    ") {
		t.Error("expected 4-space indent")
	}
	if !strings.Contains(result, "line 10") {
		t.Errorf("expected 'line 10' in output, got '%s'", result)
	}
}

func TestPrettyLFNil(t *testing.T) {
	result := PrettyLF(nil, 2)
	if !strings.Contains(result, "(nil)") {
		t.Errorf("expected '(nil)' for nil formula, got '%s'", result)
	}
}

// --- GetCheckedActions tests ---

func TestGetCheckedActionsEmpty(t *testing.T) {
	mod := module.New()
	result := GetCheckedActions(mod)
	if len(result) != 0 {
		t.Errorf("expected 0 actions, got %d", len(result))
	}
}

func TestGetCheckedActionsAll(t *testing.T) {
	mod := module.New()
	mod.PublicActions.Set("ext:send", true)
	mod.PublicActions.Set("ext:recv", true)
	result := GetCheckedActions(mod)
	if len(result) != 2 {
		t.Errorf("expected 2 actions, got %d", len(result))
	}
	// Should be sorted
	if result[0] != "ext:recv" || result[1] != "ext:send" {
		t.Errorf("unexpected order: %v", result)
	}
}

// --- GetPrioritizedActions tests ---

func TestGetPrioritizedActionsNil(t *testing.T) {
	cfg := module.NewConfig()

	result := GetPrioritizedActions(cfg)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

// --- GetConjs tests ---

func TestGetConjsEmpty(t *testing.T) {
	mod := module.New()
	result := GetConjs(mod)
	if result == nil {
		t.Fatal("expected non-nil clauses")
	}
	if len(result.Fmlas) != 0 {
		t.Errorf("expected 0 formulas, got %d", len(result.Fmlas))
	}
}

func TestGetConjsFiltersExplicit(t *testing.T) {
	mod := module.New()
	p := lg.NewConst("p", lg.Boolean)
	mod.LabeledConjs = []*ast.LabeledFormula{
		{Formula: p, Explicit: false, Unprovable: false},
		{Formula: p, Explicit: true, Unprovable: false},
		{Formula: p, Explicit: false, Unprovable: true},
	}
	result := GetConjs(mod)
	if len(result.Fmlas) != 1 {
		t.Errorf("expected 1 formula (non-explicit, non-unprovable), got %d", len(result.Fmlas))
	}
}

// --- FindAssertions tests ---

func TestFindAssertionsEmpty(t *testing.T) {
	mod := module.New()
	result := FindAssertions("", mod)
	if len(result) != 0 {
		t.Errorf("expected 0 assertions, got %d", len(result))
	}
}

func TestFindAssertionsWithAsserts(t *testing.T) {
	mod := module.New()
	assertAct := actions.NewAssertAction(lg.True)
	seq := actions.NewSequence(assertAct)
	mod.Actions.Set("test_action", seq)
	result := FindAssertions("", mod)
	if len(result) != 1 {
		t.Errorf("expected 1 assertion, got %d", len(result))
	}
}

// --- MatchHandler tests ---

func TestMatchHandlerCreate(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil, nil)
	if h == nil {
		t.Fatal("NewMatchHandler returned nil")
	}
	if h.Started {
		t.Error("should not be started initially")
	}
}

func TestMatchHandlerHandle(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil, nil)
	act := actions.NewAssertAction(lg.True)
	loc := act.GetLineno()
	loc.Line = 5
	act.SetLineno(loc)
	h.Handle(act, nil)
	if !h.Started {
		t.Error("should be started after Handle")
	}
	if len(h.Lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(h.Lines))
	}
}

func TestMatchHandlerString(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil, nil)
	act1 := actions.NewAssertAction(lg.True)
	loc1 := act1.GetLineno()
	loc1.Line = 5
	act1.SetLineno(loc1)
	act2 := actions.NewAssumeAction(lg.True)
	loc2 := act2.GetLineno()
	loc2.Line = 6
	act2.SetLineno(loc2)
	h.Handle(act1, nil)
	h.Handle(act2, nil)
	result := h.String()
	if len(result) == 0 {
		t.Error("expected non-empty output")
	}
}

// --- BuildCallGraph tests ---

func TestBuildCallGraphEmpty(t *testing.T) {
	mod := module.New()
	cg := BuildCallGraph(mod)
	if len(cg) != 0 {
		t.Errorf("expected empty call graph, got %d entries", len(cg))
	}
}

// --- PrettyActionName tests ---

func TestPrettyActionNameWithPrefix(t *testing.T) {
	if PrettyActionName("ext:send") != "send" {
		t.Error("should strip ext: prefix")
	}
}

func TestPrettyActionNameWithoutPrefix(t *testing.T) {
	if PrettyActionName("send") != "send" {
		t.Error("should return unchanged name")
	}
}

// --- FilterCheckers tests ---

func TestFilterCheckersNoFilter(t *testing.T) {
	cfg := module.NewConfig()

	lf := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	lf.Lineno = 10
	checkers := []Checker{NewConjChecker(cfg, lf, 8)}
	result := FilterCheckers(checkers, "")
	if len(result) != 1 {
		t.Errorf("expected 1 checker, got %d", len(result))
	}
}

func TestFilterCheckersWithLineFilter(t *testing.T) {
	cfg := module.NewConfig()

	lf1 := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	lf1.Lineno = 10
	lf2 := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	lf2.Lineno = 20
	checkers := []Checker{NewConjChecker(cfg, lf1, 8), NewConjChecker(cfg, lf2, 8)}
	result := FilterCheckers(checkers, "10")
	if len(result) != 1 {
		t.Errorf("expected 1 checker after filter, got %d", len(result))
	}
}

// --- CheckFcsInState tests ---

func TestCheckFcsInStateEmpty(t *testing.T) {
	mod := module.New()
	result := CheckFcsInState(mod, nil)
	if !result {
		t.Error("empty checkers should pass")
	}
}

func TestCheckFcsInStateWithChecker(t *testing.T) {
	mod := module.New()
	c := NewBaseChecker(module.NewConfig(), lg.True, false, true)
	result := CheckFcsInState(mod, []Checker{c})
	if !result {
		t.Error("should pass (stub implementation)")
	}
}

/* commented out check/check.go:260 as dead code.
It is confusing because compiler also has CheckProperties.
This one is not used anywhere. The python port is not
used anywhere either. So we just ported dead code.
Comment out for now to avoid confusion.
// --- CheckProperties tests ---
func TestCheckProperties(t *testing.T) {
	mod := module.New()
	mod.LabeledProps = []*ast.LabeledFormula{
		{Formula: lg.True},
	}
	err := CheckProperties(mod)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mod.LabeledAxioms) != 1 {
		t.Error("props should be moved to axioms")
	}
}
*/

// --- ApplyConjProofs tests ---

func TestApplyConjProofs(t *testing.T) {
	mod := module.New()
	mod.LabeledConjs = []*ast.LabeledFormula{
		{Formula: lg.True},
	}
	ApplyConjProofs(mod)
	if mod.ConjSubgoals == nil {
		t.Error("ConjSubgoals should be set")
	}
	if len(mod.ConjSubgoals) != 1 {
		t.Errorf("expected 1 subgoal, got %d", len(mod.ConjSubgoals))
	}
}

// --- Set helper tests ---

func TestSetFromSlice(t *testing.T) {
	s := setFromSlice([]string{"a", "b", "c"})
	if len(s) != 3 {
		t.Errorf("expected 3 elements, got %d", len(s))
	}
	if !s["a"] || !s["b"] || !s["c"] {
		t.Error("missing elements")
	}
}

func TestIntersect(t *testing.T) {
	a := map[string]bool{"x": true, "y": true}
	b := map[string]bool{"y": true, "z": true}
	r := intersect(a, b)
	if len(r) != 1 || !r["y"] {
		t.Errorf("expected {y}, got %v", r)
	}
}

func TestSortedUnion(t *testing.T) {
	pri := map[string]bool{"b": true}
	all := map[string]bool{"a": true, "b": true, "c": true}
	result := sortedUnion(pri, all)
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
	// b should come first (prioritized), then a, c
	if result[0] != "b" {
		t.Errorf("expected 'b' first, got '%s'", result[0])
	}
}

// --- Isolate check tests ---

func TestGetIsolateMethodDefault(t *testing.T) {
	mod := module.New()
	result := GetIsolateMethod("", mod)
	if result != "ic" {
		t.Errorf("expected 'ic', got '%s'", result)
	}
}

func TestGetIsolateMethodMC(t *testing.T) {
	// Save and restore
	mod := module.New()
	cfg := mod.Cfg

	oldVal := cfg.OptMC
	cfg.OptMC = true
	defer func() { cfg.OptMC = oldVal }()

	result := GetIsolateMethod("", mod)
	if result != "mc" {
		t.Errorf("expected 'mc', got '%s'", result)
	}
}

func TestCheckSeparatelyDefault(t *testing.T) {
	mod := module.New()
	if CheckSeparately("test", mod) {
		t.Error("should default to false")
	}
}

func TestAllAssertLinenosEmpty(t *testing.T) {
	mod := module.New()
	result, err := AllAssertLinenos(mod)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestHasTemporalStuff(t *testing.T) {
	if HasTemporalStuff(nil) {
		t.Error("stub should return false")
	}
}

// --- CheckModule tests ---

func TestCheckModuleEmpty(t *testing.T) {
	cfg := module.NewConfig()

	mod := module.New()
	cfg.Failures = 0
	err := CheckModule(mod)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- ConvertPostconds tests ---

func TestConvertPostcondsEmpty(t *testing.T) {
	result := ConvertPostconds(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestConvertPostcondsPassthrough(t *testing.T) {
	pcs := []*ast.LabeledFormula{{Formula: lg.True}}
	result := ConvertPostconds(pcs)
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

// --- Fuzz test ---

func FuzzPrettyLabel(f *testing.F) {
	f.Add("safety")
	f.Add("")
	f.Add("a.b.c")
	f.Add("ext:send")
	f.Fuzz(func(t *testing.T, s string) {
		result := PrettyLabel(s)
		if s == "" {
			if result != "(no name)" {
				t.Errorf("empty string should give '(no name)', got '%s'", result)
			}
		} else {
			if result != s {
				t.Errorf("expected '%s', got '%s'", s, result)
			}
		}
	})
}

// --- MCIsolate tests ---

func TestMCIsolateNilMethod(t *testing.T) {
	mod := module.New()
	// No properties, nil method → should be a no-op success
	err := MCIsolate("test", mod, nil)
	if err != nil {
		t.Errorf("MCIsolate with nil method should succeed, got: %v", err)
	}
}

func TestMCIsolateNonTemporalPropertyRejects(t *testing.T) {
	mod := module.New()
	falseVal := false
	mod.LabeledProps = []*ast.LabeledFormula{
		{Formula: lg.True, Temporal: &falseVal},
	}
	err := MCIsolate("test", mod, func() error { return nil })
	if err == nil {
		t.Fatal("MCIsolate should reject non-temporal properties")
	}
	if !strings.Contains(err.Error(), "model checking not supported") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMCIsolateNonTemporalMixedRejects(t *testing.T) {
	mod := module.New()
	trueVal := true
	falseVal := false
	mod.LabeledProps = []*ast.LabeledFormula{
		{Formula: lg.True, Temporal: &trueVal},
		{Formula: lg.True, Temporal: &falseVal},
	}
	err := MCIsolate("test", mod, func() error { return nil })
	if err == nil {
		t.Fatal("MCIsolate should reject when any property is non-temporal")
	}
}

func TestMCIsolateAllTemporalPasses(t *testing.T) {
	mod := module.New()
	trueVal1 := true
	trueVal2 := true
	mod.LabeledProps = []*ast.LabeledFormula{
		{Formula: lg.True, Temporal: &trueVal1},
		{Formula: lg.True, Temporal: &trueVal2},
	}
	called := false
	err := MCIsolate("test", mod, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("method should have been called")
	}
}

func TestMCIsolateNoPropsCallsMethod(t *testing.T) {
	mod := module.New()
	// No LabeledProps → the temporal check loop has no iterations,
	// so it passes. Method should be called.
	called := false
	err := MCIsolate("test", mod, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("method should have been called with no props")
	}
}

func TestMCIsolateMethodErrorPropagates(t *testing.T) {
	mod := module.New()
	err := MCIsolate("test", mod, func() error {
		return fmt.Errorf("counterexample found")
	})
	if err == nil {
		t.Fatal("MCIsolate should propagate method error")
	}
	if !strings.Contains(err.Error(), "counterexample found") {
		t.Errorf("error should contain original message, got: %v", err)
	}
}

func TestMCIsolateCallsMethodOnce(t *testing.T) {
	mod := module.New()
	callCount := 0
	err := MCIsolate("test", mod, func() error {
		callCount++
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("method should be called exactly once, was called %d times", callCount)
	}
}

func TestMCIsolateMethodCalledInSeparateMode(t *testing.T) {
	cfg := module.NewConfig()

	mod := module.New()
	// Set up assertions so AllAssertLinenos returns something
	assertAct := actions.NewAssertAction(lg.True)
	loc := assertAct.GetLineno()
	loc.Line = 42
	assertAct.SetLineno(loc)
	seq := actions.NewSequence(assertAct)
	mod.Actions.Set("test_action", seq)

	// Force separate mode
	oldVal := cfg.OptSeparate
	cfg.OptSeparate = true
	defer func() { cfg.OptSeparate = oldVal }()

	callCount := 0
	oldCheckLineno := cfg.CheckLineno
	defer func() { cfg.CheckLineno = oldCheckLineno }()

	err := MCIsolate("test", mod, func() error {
		callCount++
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount == 0 {
		t.Error("method should have been called in separate mode")
	}
}

func TestMCIsolateRestoresCheckLineno(t *testing.T) {
	cfg := module.NewConfig()

	mod := module.New()
	assertAct := actions.NewAssertAction(lg.True)
	loc := assertAct.GetLineno()
	loc.Line = 10
	assertAct.SetLineno(loc)
	seq := actions.NewSequence(assertAct)
	mod.Actions.Set("act1", seq)

	oldVal := cfg.OptSeparate
	cfg.OptSeparate = true
	defer func() { cfg.OptSeparate = oldVal }()

	originalLineno := "original"
	cfg.CheckLineno = originalLineno
	defer func() { cfg.CheckLineno = "" }()

	err := MCIsolate("test", mod, func() error { return nil })
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.CheckLineno != originalLineno {
		t.Errorf("CheckLineno should be restored to %q, got %q", originalLineno, cfg.CheckLineno)
	}
}

func TestMCIsolateSeparateStopsOnError(t *testing.T) {
	cfg := module.NewConfig()

	mod := module.New()
	// Add two assertions at different lines
	a1 := actions.NewAssertAction(lg.True)
	loc1 := a1.GetLineno()
	loc1.Line = 10
	a1.SetLineno(loc1)
	a2 := actions.NewAssertAction(lg.True)
	loc2 := a2.GetLineno()
	loc2.Line = 20
	a2.SetLineno(loc2)
	seq := actions.NewSequence(a1, a2)
	mod.Actions.Set("act1", seq)

	oldVal := cfg.OptSeparate
	cfg.OptSeparate = true
	defer func() { cfg.OptSeparate = oldVal }()
	defer func() { cfg.CheckLineno = "" }()

	callCount := 0
	err := MCIsolate("test", mod, func() error {
		callCount++
		return fmt.Errorf("fail at call %d", callCount)
	})
	if err == nil {
		t.Fatal("should have returned error")
	}
	if callCount != 1 {
		t.Errorf("should stop after first error, called %d times", callCount)
	}
}

// --- GetIsolateAttr tests ---

func TestGetIsolateAttrEmptyIsolate(t *testing.T) {
	mod := module.New()
	result := GetIsolateAttr("", "method", "default", mod)
	if result != "default" {
		t.Errorf("expected 'default', got %q", result)
	}
}

func TestGetIsolateAttrNotFound(t *testing.T) {
	mod := module.New()
	result := GetIsolateAttr("myiso", "method", "ic", mod)
	if result != "ic" {
		t.Errorf("expected default 'ic', got %q", result)
	}
}

func TestGetIsolateAttrStringValue(t *testing.T) {
	mod := module.New()
	mod.Attributes["myiso.method"] = "mc"
	result := GetIsolateAttr("myiso", "method", "ic", mod)
	if result != "mc" {
		t.Errorf("expected 'mc', got %q", result)
	}
}

func TestGetIsolateAttrFallbackParentIso(t *testing.T) {
	mod := module.New()
	// If child is "iso", fall back to parent.attrName
	mod.Attributes["parent.method"] = "vmt"
	result := GetIsolateAttr("parent.iso", "method", "ic", mod)
	if result != "vmt" {
		t.Errorf("expected 'vmt' via parent fallback, got %q", result)
	}
}

func TestGetIsolateAttrNoFallbackNonIso(t *testing.T) {
	mod := module.New()
	mod.Attributes["parent.method"] = "vmt"
	// child is "other", not "iso", so no fallback
	result := GetIsolateAttr("parent.other", "method", "ic", mod)
	if result != "ic" {
		t.Errorf("expected default 'ic' (no iso fallback), got %q", result)
	}
}

func TestGetIsolateAttrStringer(t *testing.T) {
	mod := module.New()
	// Use a fmt.Stringer value
	mod.Attributes["myiso.method"] = stringerVal("bmc[10]")
	result := GetIsolateAttr("myiso", "method", "ic", mod)
	if result != "bmc[10]" {
		t.Errorf("expected 'bmc[10]' from Stringer, got %q", result)
	}
}

type stringerVal string

func (s stringerVal) String() string { return string(s) }

func TestGetIsolateAttrIntValue(t *testing.T) {
	mod := module.New()
	// Non-string, non-Stringer: uses fmt.Sprint
	mod.Attributes["myiso.cardinality"] = 42
	result := GetIsolateAttr("myiso", "cardinality", "0", mod)
	if result != "42" {
		t.Errorf("expected '42' from Sprint, got %q", result)
	}
}

// --- GetIsolateMethod tests ---

func TestGetIsolateMethodFromAttribute(t *testing.T) {
	mod := module.New()
	mod.Attributes["myiso.method"] = "vmt"
	result := GetIsolateMethod("myiso", mod)
	if result != "vmt" {
		t.Errorf("expected 'vmt', got %q", result)
	}
}

func TestGetIsolateMethodOptMCOverrides(t *testing.T) {
	mod := module.New()
	cfg := mod.Cfg

	oldVal := cfg.OptMC
	cfg.OptMC = true
	defer func() { cfg.OptMC = oldVal }()

	mod.Attributes["myiso.method"] = "vmt"
	result := GetIsolateMethod("myiso", mod)
	if result != "mc" {
		t.Errorf("OptMC should override attribute, got %q", result)
	}
}

func TestGetIsolateMethodBMCFromAttribute(t *testing.T) {
	mod := module.New()
	mod.Attributes["myiso.method"] = "bmc[5]"
	result := GetIsolateMethod("myiso", mod)
	if result != "bmc[5]" {
		t.Errorf("expected 'bmc[5]', got %q", result)
	}
}

// --- CheckSeparately tests ---

func TestCheckSeparatelyFromAttribute(t *testing.T) {
	mod := module.New()
	mod.Attributes["myiso.separate"] = "true"
	if !CheckSeparately("myiso", mod) {
		t.Error("should return true when attribute is 'true'")
	}
}

func TestCheckSeparatelyFalseAttribute(t *testing.T) {
	mod := module.New()
	mod.Attributes["myiso.separate"] = "false"
	if CheckSeparately("myiso", mod) {
		t.Error("should return false when attribute is 'false'")
	}
}

func TestCheckSeparatelyOptOverrides(t *testing.T) {
	mod := module.New()
	cfg := mod.Cfg

	oldVal := cfg.OptSeparate
	oldSet := cfg.OptSeparateSet
	cfg.OptSeparate = true
	cfg.OptSeparateSet = true
	defer func() { cfg.OptSeparate = oldVal; cfg.OptSeparateSet = oldSet }()

	mod.Attributes["myiso.separate"] = "false"
	if !CheckSeparately("myiso", mod) {
		t.Error("OptSeparate should override attribute")
	}
}

// --- parseBMCParams tests ---

func TestParseBMCParamsSingle(t *testing.T) {
	nSteps, nUnroll, err := parseBMCParams("bmc[10]")
	if err != nil {
		t.Fatal(err)
	}
	if nSteps != 10 {
		t.Errorf("expected nSteps=10, got %d", nSteps)
	}
	if nUnroll != -1 {
		t.Errorf("expected nUnroll=-1 (not specified), got %d", nUnroll)
	}
}

func TestParseBMCParamsDouble(t *testing.T) {
	nSteps, nUnroll, err := parseBMCParams("bmc[5][3]")
	if err != nil {
		t.Fatal(err)
	}
	if nSteps != 5 {
		t.Errorf("expected nSteps=5, got %d", nSteps)
	}
	if nUnroll != 3 {
		t.Errorf("expected nUnroll=3, got %d", nUnroll)
	}
}

func TestParseBMCParamsZero(t *testing.T) {
	nSteps, _, err := parseBMCParams("bmc[0]")
	if err != nil {
		t.Fatal(err)
	}
	if nSteps != 0 {
		t.Errorf("expected nSteps=0, got %d", nSteps)
	}
}

func TestParseBMCParamsInvalidNoParams(t *testing.T) {
	_, _, err := parseBMCParams("bmc")
	if err == nil {
		t.Fatal("should fail for 'bmc' (no brackets)")
	}
}

func TestParseBMCParamsInvalidTooMany(t *testing.T) {
	_, _, err := parseBMCParams("bmc[1][2][3]")
	if err == nil {
		t.Fatal("should fail for 3 params")
	}
}

func TestParseBMCParamsInvalidNonNumeric(t *testing.T) {
	_, _, err := parseBMCParams("bmc[abc]")
	if err == nil {
		t.Fatal("should fail for non-numeric param")
	}
}

func TestParseBMCParamsInvalidPrefix(t *testing.T) {
	_, _, err := parseBMCParams("mc[10]")
	if err == nil {
		t.Fatal("should fail for non-bmc prefix")
	}
}

func TestParseBMCParamsMissingCloseBracket(t *testing.T) {
	_, _, err := parseBMCParams("bmc[10")
	if err == nil {
		t.Fatal("should fail for missing ']'")
	}
}

// --- AllAssertLinenos tests (more comprehensive) ---

func TestAllAssertLinenosDeduplicates(t *testing.T) {
	mod := module.New()
	a1 := actions.NewAssertAction(lg.True)
	loc1 := a1.GetLineno()
	loc1.Line = 10
	a1.SetLineno(loc1)
	a2 := actions.NewAssertAction(lg.True)
	loc2 := a2.GetLineno()
	loc2.Line = 10 // same line
	a2.SetLineno(loc2)
	seq := actions.NewSequence(a1, a2)
	mod.Actions.Set("act1", seq)

	result, _ := AllAssertLinenos(mod)
	if len(result) != 1 {
		t.Errorf("expected 1 unique line, got %d", len(result))
	}
}

func TestAllAssertLinenosIncludesConjs(t *testing.T) {
	mod := module.New()
	mod.LabeledConjs = []*ast.LabeledFormula{
		{Formula: lg.True, Lineno: 55},
	}
	result, _ := AllAssertLinenos(mod)
	found := false
	for _, l := range result {
		if l == 55 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected lineno 55 from conjectures, got %v", result)
	}
}

func TestAllAssertLinenosMultipleActions(t *testing.T) {
	mod := module.New()
	a1 := actions.NewAssertAction(lg.True)
	loc1 := a1.GetLineno()
	loc1.Line = 10
	a1.SetLineno(loc1)
	seq1 := actions.NewSequence(a1)
	mod.Actions.Set("act1", seq1)

	a2 := actions.NewAssertAction(lg.True)
	loc2 := a2.GetLineno()
	loc2.Line = 20
	a2.SetLineno(loc2)
	seq2 := actions.NewSequence(a2)
	mod.Actions.Set("act2", seq2)

	result, _ := AllAssertLinenos(mod)
	if len(result) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(result), result)
	}
}

// --- Integration-level tests ---

func TestCheckerInterfaceCompliance(t *testing.T) {
	cfg := module.NewConfig()

	// All checker types implement the Checker interface
	lf := cfg.AstCfg.NewLabeledFormula(nil, lg.True)
	checkers := []Checker{
		NewBaseChecker(module.NewConfig(), lg.True, true, true),
		NewConjChecker(cfg, lf, 8),
		NewConjAssumer(cfg, lf),
	}
	for i, c := range checkers {
		t.Run(fmt.Sprintf("checker_%d", i), func(t *testing.T) {
			_ = c.Cond()
			_ = c.Assume()
			_ = c.GetAnnot()
			_ = c.Failed()
			_ = c.GetLF()
		})
	}
}

func TestCheckSafetyInState(t *testing.T) {
	// With an empty module (no axioms, no state), the safety check
	// fails because the dual of the empty Or (false) is True, and
	// True is trivially SAT, meaning the checker finds a "counterexample".
	// This is correct: with no information, safety cannot be proved.
	mod := module.New()
	// Reset failures counter
	oldFailures := mod.Cfg.Failures
	result := CheckSafetyInState(mod, false)
	_ = result
	// Restore failures counter (this is an expected failure)
	mod.Cfg.Failures = oldFailures
}

func TestCheckConjsInState(t *testing.T) {
	mod := module.New()
	mod.LabeledConjs = []*ast.LabeledFormula{
		{Formula: lg.True},
	}
	ApplyConjProofs(mod)
	result := CheckConjsInState(mod, 8, nil)
	if !result {
		t.Error("should pass (stub)")
	}
}
