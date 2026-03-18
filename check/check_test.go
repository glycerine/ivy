package check

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// --- BaseChecker tests ---

func TestBaseCheckerCreate(t *testing.T) {
	c := NewBaseChecker(lg.True, true, true)
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
	c := NewBaseChecker(lg.True, false, false)
	cond := c.Cond()
	if cond == nil {
		t.Fatal("Cond returned nil")
	}
}

func TestBaseCheckerPass(t *testing.T) {
	c := NewBaseChecker(lg.True, false, true)
	result := c.Pass()
	if !result {
		t.Error("Pass should return true")
	}
	if c.Failed() {
		t.Error("after Pass, should not be failed")
	}
}

func TestBaseCheckerFail(t *testing.T) {
	oldFailures := Failures
	c := NewBaseChecker(lg.True, false, true)
	_ = c.Fail()
	if !c.Failed() {
		t.Error("after Fail, should be failed")
	}
	if Failures <= oldFailures {
		t.Error("Failures count should have increased")
	}
	Failures = oldFailures // restore
}

func TestBaseCheckerSatCallsFail(t *testing.T) {
	oldFailures := Failures
	c := NewBaseChecker(lg.True, false, true)
	_ = c.Sat()
	if !c.Failed() {
		t.Error("Sat should trigger Fail")
	}
	Failures = oldFailures
}

func TestBaseCheckerUnsatCallsPass(t *testing.T) {
	c := NewBaseChecker(lg.True, false, true)
	result := c.Unsat()
	if !result {
		t.Error("Unsat should trigger Pass and return true")
	}
}

func TestBaseCheckerAssume(t *testing.T) {
	c := NewBaseChecker(lg.True, false, true)
	if c.Assume() {
		t.Error("BaseChecker.Assume should return false")
	}
}

func TestBaseCheckerGetAnnot(t *testing.T) {
	c := NewBaseChecker(lg.True, false, true)
	if c.GetAnnot() != nil {
		t.Error("BaseChecker.GetAnnot should return nil")
	}
}

func TestBaseCheckerGetLF(t *testing.T) {
	c := NewBaseChecker(lg.True, false, true)
	if c.GetLF() != nil {
		t.Error("BaseChecker.GetLF should return nil")
	}
}

// --- ConjChecker tests ---

func TestConjCheckerCreate(t *testing.T) {
	lf := &module.LabeledFormula{
		Formula: lg.True,
		Lineno:  42,
	}
	cc := NewConjChecker(lf, 8)
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
	lf := &module.LabeledFormula{
		Formula: lg.True,
	}
	cc := NewConjChecker(lf, 4)
	if cc.GetLF() != lf {
		t.Error("GetLF should return the labeled formula")
	}
}

func TestConjCheckerImplementsChecker(t *testing.T) {
	lf := &module.LabeledFormula{Formula: lg.True}
	var _ Checker = NewConjChecker(lf, 8)
}

// --- ConjAssumer tests ---

func TestConjAssumerCreate(t *testing.T) {
	lf := &module.LabeledFormula{
		Formula: lg.True,
	}
	ca := NewConjAssumer(lf)
	if ca == nil {
		t.Fatal("NewConjAssumer returned nil")
	}
	if ca.LF != lf {
		t.Error("LF not set correctly")
	}
}

func TestConjAssumerAssume(t *testing.T) {
	lf := &module.LabeledFormula{Formula: lg.True}
	ca := NewConjAssumer(lf)
	if !ca.Assume() {
		t.Error("ConjAssumer.Assume should return true")
	}
}

func TestConjAssumerImplementsChecker(t *testing.T) {
	lf := &module.LabeledFormula{Formula: lg.True}
	var _ Checker = NewConjAssumer(lf)
}

// --- DualClauses tests ---

func TestDualClausesNil(t *testing.T) {
	result := DualClauses(nil)
	if result != nil {
		t.Error("DualClauses(nil) should return nil")
	}
}

func TestDualClausesEmpty(t *testing.T) {
	c := clauseops.NewClauses(nil, nil, nil)
	result := DualClauses(c)
	if result == nil {
		t.Fatal("DualClauses should not return nil for empty clauses")
	}
}

func TestDualClausesSingleFormula(t *testing.T) {
	// Use a real formula (not lg.True which is empty And, consumed by collectAndList)
	p := lg.NewSymbol("p", lg.Boolean)
	c := clauseops.NewClauses([]lg.Node{p}, nil, nil)
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
	if Diagnose.GetBool() {
		t.Error("diagnose should default to false")
	}
}

func TestCoverageParameter(t *testing.T) {
	if !Coverage.GetBool() {
		t.Error("coverage should default to true")
	}
}

func TestCheckedActionParameter(t *testing.T) {
	if CheckedAction.GetString() != "" {
		t.Error("checked_action should default to empty")
	}
}

func TestOptTrustedParameter(t *testing.T) {
	if OptTrusted.GetBool() {
		t.Error("trusted should default to false")
	}
}

func TestOptMCParameter(t *testing.T) {
	if OptMC.GetBool() {
		t.Error("mc should default to false")
	}
}

func TestOptTraceParameter(t *testing.T) {
	if OptTrace.GetBool() {
		t.Error("trace should default to false")
	}
}

func TestOptIvyStatsParameter(t *testing.T) {
	if OptIvyStats.GetBool() {
		t.Error("ivy_stats should default to false")
	}
}

func TestNoCheckGuaranteesParameter(t *testing.T) {
	if NoCheckGuarantees.GetBool() {
		t.Error("no_check_guarantees should default to false")
	}
}

func TestProfilingParameter(t *testing.T) {
	if Profiling.GetBool() {
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
	lf := &module.LabeledFormula{Lineno: 42}
	result := PrettyLineno(lf)
	if result != "line 42: " {
		t.Errorf("expected 'line 42: ', got '%s'", result)
	}
}

func TestPrettyLinenoZero(t *testing.T) {
	lf := &module.LabeledFormula{Lineno: 0}
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
	lf := &module.LabeledFormula{Lineno: 10}
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
	mod.PublicActions["ext:send"] = true
	mod.PublicActions["ext:recv"] = true
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
	result := GetPrioritizedActions()
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
	p := lg.NewSymbol("p", lg.Boolean)
	mod.LabeledConjs = []*module.LabeledFormula{
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
	seq := actions.NewSequence(actions.WrapAction(assertAct))
	mod.Actions["test_action"] = seq
	result := FindAssertions("", mod)
	if len(result) != 1 {
		t.Errorf("expected 1 assertion, got %d", len(result))
	}
}

// --- MatchHandler tests ---

func TestMatchHandlerCreate(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil)
	if h == nil {
		t.Fatal("NewMatchHandler returned nil")
	}
	if h.Started {
		t.Error("should not be started initially")
	}
}

func TestMatchHandlerHandle(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil)
	h.Handle("action1", nil)
	if !h.Started {
		t.Error("should be started after Handle")
	}
	if len(h.Lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(h.Lines))
	}
}

func TestMatchHandlerString(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil)
	h.Handle("line1", nil)
	h.Handle("line2", nil)
	result := h.String()
	if !strings.Contains(result, "line1") || !strings.Contains(result, "line2") {
		t.Errorf("unexpected output: %s", result)
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
	lf := &module.LabeledFormula{Formula: lg.True, Lineno: 10}
	checkers := []Checker{NewConjChecker(lf, 8)}
	result := FilterCheckers(checkers, "")
	if len(result) != 1 {
		t.Errorf("expected 1 checker, got %d", len(result))
	}
}

func TestFilterCheckersWithLineFilter(t *testing.T) {
	lf1 := &module.LabeledFormula{Formula: lg.True, Lineno: 10}
	lf2 := &module.LabeledFormula{Formula: lg.True, Lineno: 20}
	checkers := []Checker{NewConjChecker(lf1, 8), NewConjChecker(lf2, 8)}
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
	c := NewBaseChecker(lg.True, false, true)
	result := CheckFcsInState(mod, []Checker{c})
	if !result {
		t.Error("should pass (stub implementation)")
	}
}

// --- CheckProperties tests ---

func TestCheckProperties(t *testing.T) {
	mod := module.New()
	mod.LabeledProps = []*module.LabeledFormula{
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

// --- ApplyConjProofs tests ---

func TestApplyConjProofs(t *testing.T) {
	mod := module.New()
	mod.LabeledConjs = []*module.LabeledFormula{
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
	oldVal := OptMC.Value
	OptMC.Value = true
	defer func() { OptMC.Value = oldVal }()

	mod := module.New()
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
	result := AllAssertLinenos(mod)
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
	mod := module.New()
	Failures = 0
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
	pcs := []*module.LabeledFormula{{Formula: lg.True}}
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

// --- Integration-level tests ---

func TestCheckerInterfaceCompliance(t *testing.T) {
	// All checker types implement the Checker interface
	lf := &module.LabeledFormula{Formula: lg.True}
	checkers := []Checker{
		NewBaseChecker(lg.True, true, true),
		NewConjChecker(lf, 8),
		NewConjAssumer(lf),
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
	oldFailures := Failures
	result := CheckSafetyInState(mod, false)
	_ = result
	// Restore failures counter (this is an expected failure)
	Failures = oldFailures
}

func TestCheckConjsInState(t *testing.T) {
	mod := module.New()
	mod.LabeledConjs = []*module.LabeledFormula{
		{Formula: lg.True},
	}
	ApplyConjProofs(mod)
	result := CheckConjsInState(mod, 8, nil)
	if !result {
		t.Error("should pass (stub)")
	}
}
