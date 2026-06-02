package goivy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --- BaseChecker tests ---

func TestBaseCheckerCreate(t *testing.T) {
	c := NewBaseChecker(New(), True, true, true)
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
	c := NewBaseChecker(New(), True, false, false)
	cond := c.Cond()
	if cond == nil {
		t.Fatal("Cond returned nil")
	}
}

func TestBaseCheckerStartPrintsDotsWithNewlineLikePython(t *testing.T) {
	c := NewBaseChecker(New(), True, true, false)
	out := captureActionUpdateStdout(t, func() {
		c.Start()
	})
	if out != "...\n" {
		t.Fatalf("BaseChecker.Start output mismatch\n got: %q\nwant: %q", out, "...\n")
	}
}

func TestBaseCheckerPass(t *testing.T) {
	c := NewBaseChecker(New(), True, false, true)
	result := c.Pass()
	if !result {
		t.Error("Pass should return true")
	}
	if c.Failed() {
		t.Error("after Pass, should not be failed")
	}
}

func TestBaseCheckerFail(t *testing.T) {
	mod := New()
	oldFailures := mod.Cfg.Failures
	c := NewBaseChecker(mod, True, false, true)
	_ = c.Fail()
	if !c.Failed() {
		t.Error("after Fail, should be failed")
	}
	if mod.Cfg.Failures <= oldFailures {
		t.Error("Failures count should have increased")
	}
	mod.Cfg.Failures = oldFailures // restore
}

func TestBaseCheckerSatCallsFail(t *testing.T) {
	mod := New()
	oldFailures := mod.Cfg.Failures
	c := NewBaseChecker(mod, True, false, true)
	_ = c.Sat()
	if !c.Failed() {
		t.Error("Sat should trigger Fail")
	}
	mod.Cfg.Failures = oldFailures
}

func TestBaseCheckerUnsatCallsPass(t *testing.T) {
	c := NewBaseChecker(New(), True, false, true)
	result := c.Unsat()
	if !result {
		t.Error("Unsat should trigger Pass and return true")
	}
}

func TestBaseCheckerAssume(t *testing.T) {
	c := NewBaseChecker(New(), True, false, true)
	if c.Assume() {
		t.Error("BaseChecker.Assume should return false")
	}
}

func TestBaseCheckerGetAnnot(t *testing.T) {
	c := NewBaseChecker(New(), True, false, true)
	if c.GetAnnot() != nil {
		t.Error("BaseChecker.GetAnnot should return nil")
	}
}

func TestBaseCheckerGetLF(t *testing.T) {
	c := NewBaseChecker(New(), True, false, true)
	if c.GetLF() != nil {
		t.Error("BaseChecker.GetLF should return nil")
	}
}

// --- ConjChecker tests ---

func TestConjCheckerCreate(t *testing.T) {
	mod := New()

	lf := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	lf.SetLineno(Location{Line: 42})
	cc := NewConjChecker(mod, lf, 8)
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
	mod := New()

	lf := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	cc := NewConjChecker(mod, lf, 4)
	if cc.GetLF() != lf {
		t.Error("GetLF should return the labeled formula")
	}
}

func TestConjCheckerImplementsChecker(t *testing.T) {
	mod := New()

	lf := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	var _ Checker = NewConjChecker(mod, lf, 8)
}

// --- ConjAssumer tests ---

func TestConjAssumerCreate(t *testing.T) {
	mod := New()

	lf := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	ca := NewConjAssumer(mod, lf)
	if ca == nil {
		t.Fatal("NewConjAssumer returned nil")
	}
	if ca.LF != lf {
		t.Error("LF not set correctly")
	}
}

func TestConjAssumerAssume(t *testing.T) {
	mod := New()

	lf := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	ca := NewConjAssumer(mod, lf)
	if !ca.Assume() {
		t.Error("ConjAssumer.Assume should return true")
	}
}

func TestConjAssumerImplementsChecker(t *testing.T) {
	mod := New()

	lf := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	var _ Checker = NewConjAssumer(mod, lf)
}

// --- Parameter tests ---

func TestDiagnoseParameter(t *testing.T) {
	cfg := NewConfig()

	if cfg.Diagnose {
		t.Error("diagnose should default to false")
	}
}

func TestShowCounterexampleConsumesSatisfyResultLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_check, ivy_module as im

im.module = im.Module()

captured = {}

class State(object):
    pass

class AnalysisGraph(object):
    def copy_path(self, state, other_art, bound):
        other_art.states = [State()]

def fake_gui_art(other_art):
    captured["state_count"] = len(other_art.states)
    captured["value_assigned"] = other_art.states[0].value == "path0"
    captured["universe_assigned"] = other_art.states[0].universe == {"S": ["0"]}

ivy_check.gui_art = fake_gui_art
universe = {"S": ["0"]}
path = ["path0"]
ivy_check.show_counterexample(AnalysisGraph(), object(), (universe, path))
print(json.dumps(captured, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python show_counterexample oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		StateCount       int  `json:"state_count"`
		ValueAssigned    bool `json:"value_assigned"`
		UniverseAssigned bool `json:"universe_assigned"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python show_counterexample oracle %q: %v", out, err)
	}
	if !want.ValueAssigned || !want.UniverseAssigned {
		t.Fatalf("python show_counterexample oracle did not assign value/universe: %+v", want)
	}

	mod := New()
	ag := NewAnalysisGraph(mod)
	source := NewState(mod, TrueClauses(EmptyAnnotation{}))
	ag.Add(source, nil)
	pathValue := PureStateClauses(FalseClauses(EmptyAnnotation{}))
	sortS := &UninterpretedSort{Name: "S"}
	universes := map[string][]Expr{"S": {NewConst("0:S", sortS)}}

	var captured *AnalysisGraph
	mod.Cfg.GuiArtHook = func(_ *Module, target interface{}, _ *Clauses) error {
		var ok bool
		captured, ok = target.(*AnalysisGraph)
		if !ok {
			t.Fatalf("GuiArtHook target type = %T, want *AnalysisGraph", target)
		}
		return nil
	}

	ShowCounterexample(ag, source, &SatisfyResult{
		Universes: universes,
		Path:      []*Update{pathValue},
	})

	if captured == nil {
		t.Fatal("GuiArtHook was not called")
	}
	if len(captured.States) != want.StateCount {
		t.Fatalf("ShowCounterexample copied state count differs from Python\nwant: %d\ngot:  %d", want.StateCount, len(captured.States))
	}
	got := captured.States[len(captured.States)-1]
	if got.Value != pathValue {
		t.Fatalf("ShowCounterexample did not assign SatisfyResult path value to copied state")
	}
	gotUniverses, ok := got.Universe.(map[string][]Expr)
	if !ok || len(gotUniverses["S"]) != 1 {
		t.Fatalf("ShowCounterexample did not assign SatisfyResult universes to copied state: %#v", got.Universe)
	}
}

func TestInitializerGuaranteesUsePythonLeakedActionSemantics(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_module as im, logic as lg

im.module = im.Module()
first = act.AssertAction(lg.Const("A", lg.Boolean))
last = act.AssertAction(lg.Const("B", lg.Boolean))
im.module.initializers = [("a", first), ("z", last)]

for actname, action in sorted(im.module.initializers, key=lambda x: x[0]):
    pass
guarantees = [sub for sub in action.iter_subactions()
              if isinstance(sub, (act.AssertAction, act.Ranking))
              for action in im.module.initializers]
print(json.dumps([str(g.args[0]) for g in guarantees]))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python initializer guarantee oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python initializer guarantee oracle %q: %v", out, err)
	}
	if !reflect.DeepEqual(want, []string{"B", "B"}) {
		t.Fatalf("python initializer guarantee oracle changed: %#v", want)
	}

	mod := New()
	first := NewAssertAction(NewConst("A", Boolean))
	last := NewAssertAction(NewConst("B", Boolean))
	mod.Initializers = []NamedAction{
		{Name: "a", Action: first},
		{Name: "z", Action: last},
	}

	guarantees := initializerGuaranteesForCheckIsolate(mod)
	got := make([]string, 0, len(guarantees))
	for _, guarantee := range guarantees {
		if assert, ok := guarantee.(*LogicAssertAction); ok {
			got = append(got, fmt.Sprint(assert.Formula))
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Go initializer guarantees differ from Python leaked-action semantics\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestCoverageParameter(t *testing.T) {
	cfg := NewConfig()

	if !cfg.Coverage {
		t.Error("coverage should default to true")
	}
}

func TestCheckedActionParameter(t *testing.T) {
	cfg := NewConfig()
	if cfg.CheckedAction != "" {
		t.Error("checked_action should default to empty")
	}
}

func TestOptTrustedParameter(t *testing.T) {
	cfg := NewConfig()
	if cfg.OptTrusted {
		t.Error("trusted should default to false")
	}
}

func TestOptMCParameter(t *testing.T) {
	cfg := NewConfig()
	if cfg.OptMC {
		t.Error("mc should default to false")
	}
}

func TestOptTraceParameter(t *testing.T) {
	cfg := NewConfig()
	if cfg.OptTrace {
		t.Error("trace should default to false")
	}
}

func TestOptIvyStatsParameter(t *testing.T) {
	cfg := NewConfig()
	if cfg.OptIvyStats {
		t.Error("ivy_stats should default to false")
	}
}

func TestNoCheckGuaranteesParameter(t *testing.T) {
	cfg := NewConfig()
	if cfg.NoCheckGuarantees {
		t.Error("no_check_guarantees should default to false")
	}
}

func TestProfilingParameter(t *testing.T) {
	cfg := NewConfig()
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
	acfg := NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, nil)
	lf.SetLineno(Location{Filename: "test.ivy", Line: 42})
	result := CheckPrettyLineno(lf)
	// Python: return str(ast.lineno) — LocationTuple with filename and line.
	if result != "test.ivy: line 42: " {
		t.Errorf("expected 'test.ivy: line 42: ', got '%s'", result)
	}
}

func TestPrettyLinenoZero(t *testing.T) {
	acfg := NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, nil)
	result := CheckPrettyLineno(lf)
	if result != "(internal) " {
		t.Errorf("expected '(internal) ', got '%s'", result)
	}
}

func TestPrettyLinenoNil(t *testing.T) {
	result := CheckPrettyLineno(nil)
	if result != "(internal) " {
		t.Errorf("expected '(internal) ', got '%s'", result)
	}
}

func TestPrettyLF(t *testing.T) {
	acfg := NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, nil)
	lf.SetLineno(Location{Filename: "test.ivy", Line: 10})
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
	mod := New()
	result := GetCheckedActions(mod)
	if len(result) != 0 {
		t.Errorf("expected 0 actions, got %d", len(result))
	}
}

func TestGetCheckedActionsAll(t *testing.T) {
	mod := New()
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
	cfg := NewConfig()

	result := GetPrioritizedActions(cfg)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

// --- GetConjs tests ---

func TestGetConjsEmpty(t *testing.T) {
	mod := New()
	result := GetConjs(mod)
	if result == nil {
		t.Fatal("expected non-nil clauses")
	}
	if len(result.Fmlas) != 0 {
		t.Errorf("expected 0 formulas, got %d", len(result.Fmlas))
	}
}

func TestGetConjsFiltersExplicit(t *testing.T) {
	mod := New()
	p := NewConst("p", Boolean)
	mod.LabeledConjs = []*LabeledFormula{
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
	mod := New()
	result := FindAssertions("", mod)
	if len(result) != 0 {
		t.Errorf("expected 0 assertions, got %d", len(result))
	}
}

func TestFindAssertionsWithAsserts(t *testing.T) {
	mod := New()
	assertAct := NewAssertAction(True)
	seq := NewSequence(assertAct)
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
	act := NewAssertAction(True)
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

func TestMatchHandlerFormatsActionLocationLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_logic as il, ivy_utils as iu

action = act.AssertAction(il.And())
action.lineno = iu.Location("sample.ivy", 7)
print(json.dumps("{}{}".format(action.lineno, action)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python MatchHandler location oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python MatchHandler location oracle %q: %v", out, err)
	}

	h := NewMatchHandler(nil, nil, nil, nil)
	action := NewAssertAction(True)
	action.SetLineno(Location{Filename: "sample.ivy", Line: 7})
	h.Handle(action, nil)

	if len(h.Lines) != 1 {
		t.Fatalf("MatchHandler line count=%d, want 1", len(h.Lines))
	}
	if got := h.Lines[0]; got != want {
		t.Fatalf("MatchHandler formatted line differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestMatchHandlerHandlesLineZeroLocationLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_logic as il, ivy_utils as iu

action = act.AssertAction(il.And())
action.lineno = iu.Location("nowhere", 0)
print(json.dumps("{}{}".format(action.lineno, action)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python MatchHandler line-zero oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python MatchHandler line-zero oracle %q: %v", out, err)
	}

	h := NewMatchHandler(nil, nil, nil, nil)
	action := NewAssertAction(True)
	action.SetLineno(Location{Filename: "nowhere", Line: 0})
	h.Handle(action, nil)

	if len(h.Lines) != 1 {
		t.Fatalf("MatchHandler line count=%d, want 1", len(h.Lines))
	}
	if got := h.Lines[0]; got != want {
		t.Fatalf("MatchHandler line-zero formatted line differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestMatchHandlerString(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil, nil)
	act1 := NewAssertAction(True)
	loc1 := act1.GetLineno()
	loc1.Line = 5
	act1.SetLineno(loc1)
	act2 := NewAssumeAction(True)
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
	mod := New()
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
	mod := New()

	lf := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	lf.SetLineno(Location{Line: 10})
	checkers := []Checker{NewConjChecker(mod, lf, 8)}
	result := FilterCheckers(checkers, "")
	if len(result) != 1 {
		t.Errorf("expected 1 checker, got %d", len(result))
	}
}

func TestFilterCheckersWithLineFilter(t *testing.T) {
	mod := New()

	lf1 := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	lf1.SetLineno(Location{Filename: "test.ivy", Line: 10})
	lf2 := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	lf2.SetLineno(Location{Filename: "test.ivy", Line: 20})
	checkers := []Checker{NewConjChecker(mod, lf1, 8), NewConjChecker(mod, lf2, 8)}
	result := FilterCheckers(checkers, "test.ivy:10")
	if len(result) != 1 {
		t.Errorf("expected 1 checker after filter, got %d", len(result))
	}
}

// --- CheckFcsInState tests ---

func TestCheckFcsInStateEmpty(t *testing.T) {
	mod := New()
	result := CheckFcsInState(mod, nil)
	if !result {
		t.Error("empty checkers should pass")
	}
}

func TestCheckFcsInStateWithChecker(t *testing.T) {
	mod := New()
	c := NewBaseChecker(mod, True, false, true)
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
	mod := New()
	mod.LabeledConjs = []*LabeledFormula{
		{Formula: True},
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
	mod := New()
	result := GetIsolateMethod("", mod)
	if result != "ic" {
		t.Errorf("expected 'ic', got '%s'", result)
	}
}

func TestGetIsolateMethodMC(t *testing.T) {
	// Save and restore
	mod := New()
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
	mod := New()
	if CheckSeparately("test", mod) {
		t.Error("should default to false")
	}
}

func TestAllAssertLinenosEmpty(t *testing.T) {
	mod := New()
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
	cfg := NewConfig()

	mod := New()
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
	pcs := []*LabeledFormula{{Formula: True}}
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
	mod := New()
	// No properties, nil method → should be a no-op success
	err := MCIsolate("test", mod, nil)
	if err != nil {
		t.Errorf("MCIsolate with nil method should succeed, got: %v", err)
	}
}

func TestMCIsolateNonTemporalPropertyRejects(t *testing.T) {
	mod := New()
	falseVal := false
	mod.LabeledProps = []*LabeledFormula{
		{Formula: True, Temporal: &falseVal},
	}
	err := MCIsolate("test", mod, func(_ *Module) error { return nil })
	if err == nil {
		t.Fatal("MCIsolate should reject non-temporal properties")
	}
	if !strings.Contains(err.Error(), "model checking not supported") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMCIsolateNonTemporalMixedRejects(t *testing.T) {
	mod := New()
	trueVal := true
	falseVal := false
	mod.LabeledProps = []*LabeledFormula{
		{Formula: True, Temporal: &trueVal},
		{Formula: True, Temporal: &falseVal},
	}
	err := MCIsolate("test", mod, func(_ *Module) error { return nil })
	if err == nil {
		t.Fatal("MCIsolate should reject when any property is non-temporal")
	}
}

func TestMCIsolateAllTemporalPasses(t *testing.T) {
	mod := New()
	trueVal1 := true
	trueVal2 := true
	mod.LabeledProps = []*LabeledFormula{
		{Formula: True, Temporal: &trueVal1},
		{Formula: True, Temporal: &trueVal2},
	}
	called := false
	err := MCIsolate("test", mod, func(_ *Module) error {
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
	mod := New()
	// No LabeledProps → the temporal check loop has no iterations,
	// so it passes. Method should be called.
	called := false
	err := MCIsolate("test", mod, func(_ *Module) error {
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
	mod := New()
	err := MCIsolate("test", mod, func(_ *Module) error {
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
	mod := New()
	callCount := 0
	err := MCIsolate("test", mod, func(_ *Module) error {
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
	cfg := NewConfig()

	mod := New()
	// Set up assertions so AllAssertLinenos returns something
	assertAct := NewAssertAction(True)
	loc := assertAct.GetLineno()
	loc.Line = 42
	assertAct.SetLineno(loc)
	seq := NewSequence(assertAct)
	mod.Actions.Set("test_action", seq)

	// Force separate mode
	oldVal := cfg.OptSeparate
	cfg.OptSeparate = true
	defer func() { cfg.OptSeparate = oldVal }()

	callCount := 0
	oldCheckLineno := cfg.CheckLineno
	defer func() { cfg.CheckLineno = oldCheckLineno }()

	err := MCIsolate("test", mod, func(_ *Module) error {
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
	cfg := NewConfig()

	mod := New()
	assertAct := NewAssertAction(True)
	loc := assertAct.GetLineno()
	loc.Line = 10
	assertAct.SetLineno(loc)
	seq := NewSequence(assertAct)
	mod.Actions.Set("act1", seq)

	oldVal := cfg.OptSeparate
	cfg.OptSeparate = true
	defer func() { cfg.OptSeparate = oldVal }()

	originalLineno := "original"
	cfg.CheckLineno = originalLineno
	defer func() { cfg.CheckLineno = "" }()

	err := MCIsolate("test", mod, func(_ *Module) error { return nil })
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.CheckLineno != originalLineno {
		t.Errorf("CheckLineno should be restored to %q, got %q", originalLineno, cfg.CheckLineno)
	}
}

func TestMCIsolateSeparateStopsOnError(t *testing.T) {
	cfg := NewConfig()

	mod := New()
	// Add two assertions at different lines
	a1 := NewAssertAction(True)
	loc1 := a1.GetLineno()
	loc1.Line = 10
	a1.SetLineno(loc1)
	a2 := NewAssertAction(True)
	loc2 := a2.GetLineno()
	loc2.Line = 20
	a2.SetLineno(loc2)
	seq := NewSequence(a1, a2)
	mod.Actions.Set("act1", seq)

	oldVal := cfg.OptSeparate
	cfg.OptSeparate = true
	defer func() { cfg.OptSeparate = oldVal }()
	defer func() { cfg.CheckLineno = "" }()

	callCount := 0
	err := MCIsolate("test", mod, func(_ *Module) error {
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
	mod := New()
	result := GetIsolateAttr("", "method", "default", mod)
	if result != "default" {
		t.Errorf("expected 'default', got %q", result)
	}
}

func TestGetIsolateAttrNotFound(t *testing.T) {
	mod := New()
	result := GetIsolateAttr("myiso", "method", "ic", mod)
	if result != "ic" {
		t.Errorf("expected default 'ic', got %q", result)
	}
}

func TestGetIsolateAttrStringValue(t *testing.T) {
	mod := New()
	mod.Attributes["myiso.method"] = "mc"
	result := GetIsolateAttr("myiso", "method", "ic", mod)
	if result != "mc" {
		t.Errorf("expected 'mc', got %q", result)
	}
}

func TestGetIsolateAttrFallbackParentIso(t *testing.T) {
	mod := New()
	// If child is "iso", fall back to parent.attrName
	mod.Attributes["parent.method"] = "vmt"
	result := GetIsolateAttr("parent.iso", "method", "ic", mod)
	if result != "vmt" {
		t.Errorf("expected 'vmt' via parent fallback, got %q", result)
	}
}

func TestGetIsolateAttrNoFallbackNonIso(t *testing.T) {
	mod := New()
	mod.Attributes["parent.method"] = "vmt"
	// child is "other", not "iso", so no fallback
	result := GetIsolateAttr("parent.other", "method", "ic", mod)
	if result != "ic" {
		t.Errorf("expected default 'ic' (no iso fallback), got %q", result)
	}
}

func TestGetIsolateAttrStringer(t *testing.T) {
	mod := New()
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
	mod := New()
	// Non-string, non-Stringer: uses fmt.Sprint
	mod.Attributes["myiso.cardinality"] = 42
	result := GetIsolateAttr("myiso", "cardinality", "0", mod)
	if result != "42" {
		t.Errorf("expected '42' from Sprint, got %q", result)
	}
}

// --- GetIsolateMethod tests ---

func TestGetIsolateMethodFromAttribute(t *testing.T) {
	mod := New()
	mod.Attributes["myiso.method"] = "vmt"
	result := GetIsolateMethod("myiso", mod)
	if result != "vmt" {
		t.Errorf("expected 'vmt', got %q", result)
	}
}

func TestGetIsolateMethodOptMCOverrides(t *testing.T) {
	mod := New()
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
	mod := New()
	mod.Attributes["myiso.method"] = "bmc[5]"
	result := GetIsolateMethod("myiso", mod)
	if result != "bmc[5]" {
		t.Errorf("expected 'bmc[5]', got %q", result)
	}
}

// --- CheckSeparately tests ---

func TestCheckSeparatelyFromAttribute(t *testing.T) {
	mod := New()
	mod.Attributes["myiso.separate"] = "true"
	if !CheckSeparately("myiso", mod) {
		t.Error("should return true when attribute is 'true'")
	}
}

func TestCheckSeparatelyFalseAttribute(t *testing.T) {
	mod := New()
	mod.Attributes["myiso.separate"] = "false"
	if CheckSeparately("myiso", mod) {
		t.Error("should return false when attribute is 'false'")
	}
}

func TestCheckSeparatelyOptOverrides(t *testing.T) {
	mod := New()
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

// --- SomeBounded (BMC flag) tests ---
//
// Mirrors Python ivy_check.py's `some_bounded` global, which is set when an
// isolate uses a bmc[N] verification method and read at the end of start()
// to print "BOUNDED" before "OK".

func TestSomeBoundedZeroByDefault(t *testing.T) {
	cfg := NewConfig()
	if cfg.SomeBounded {
		t.Error("fresh Config should have SomeBounded == false")
	}
}

// loadBMCFixture parses test_vectors/bmc_minimal.ivy into a fresh module
// and optionally sets the "method" attribute. Returns the module ready for
// CheckModule. The fixture uses the implicit "this" isolate, which is the
// only way the BMC dispatch can be reached without a more complex isolate
// graph (a hand-built IsolateDef without going through the compiler fails
// CreateIsolate).
func loadBMCFixture(t *testing.T, method string) *Module {
	t.Helper()
	mod := New()
	mod.Cfg = NewConfig()
	mod.Cfg.Isolate = "this"
	err := SourceFile(
		"test_vectors/bmc_minimal.ivy",
		mod, mod.Sig,
		map[string]interface{}{"create_isolate": false},
	)
	if err != nil {
		t.Fatalf("loadBMCFixture: SourceFile: %v", err)
	}
	if method != "" {
		// ComposeNames("this", "method") strips "this" and returns "method",
		// so the attribute key for the implicit isolate is just "method".
		mod.Attributes["method"] = method
	}
	return mod
}

func TestSomeBoundedSetByBMCBranch(t *testing.T) {
	// CheckModule's BMC branch should set mod.Cfg.SomeBounded = true on the
	// PARENT mod.Cfg (not isoMod.Cfg, which would be lost by Module.Copy).
	mod := loadBMCFixture(t, "bmc[1]")
	_ = CheckModule(mod) // ignore the BMC verdict; we only care about the flag
	if !mod.Cfg.SomeBounded {
		t.Error("expected mod.Cfg.SomeBounded == true after BMC isolate")
	}
}

func TestSomeBoundedNotSetForDefaultMethod(t *testing.T) {
	// No method attribute → defaults to "ic" (induction check) → BMC branch
	// is never reached → SomeBounded stays false.
	mod := loadBMCFixture(t, "")
	_ = CheckModule(mod)
	if mod.Cfg.SomeBounded {
		t.Error("SomeBounded should remain false when no isolate uses BMC")
	}
}

func TestSomeBoundedResetByStartWithConfig(t *testing.T) {
	// Pre-set the flag to true on a Config, then call StartWithConfig.
	// The reset at the entry of StartWithConfig should clear it before
	// any work happens. We use an invalid filename so StartWithConfig
	// returns immediately after the reset and before CheckModule runs.
	cfg := NewConfig()
	cfg.SomeBounded = true
	// Use an empty args slice; StartWithConfig should return Usage error
	// after the reset has already executed.
	_ = Start([]string{}, cfg)
	// The reset happens after the args check (lines 1240-1242), so for an
	// empty args slice the reset is NOT yet reached. We instead use a
	// non-existent .ivy file so the args check passes and the reset runs.
	cfg2 := NewConfig()
	cfg2.SomeBounded = true
	_ = Start([]string{"/nonexistent/path/that/will/fail.ivy"}, cfg2)
	if cfg2.SomeBounded {
		t.Error("StartWithConfig should reset SomeBounded to false at entry")
	}
}

// --- AllAssertLinenos tests (more comprehensive) ---

func TestAllAssertLinenosDeduplicates(t *testing.T) {
	mod := New()
	a1 := NewAssertAction(True)
	loc1 := a1.GetLineno()
	loc1.Line = 10
	a1.SetLineno(loc1)
	a2 := NewAssertAction(True)
	loc2 := a2.GetLineno()
	loc2.Line = 10 // same line
	a2.SetLineno(loc2)
	seq := NewSequence(a1, a2)
	mod.Actions.Set("act1", seq)

	result, _ := AllAssertLinenos(mod)
	if len(result) != 1 {
		t.Errorf("expected 1 unique line, got %d", len(result))
	}
}

func TestAllAssertLinenosIncludesConjs(t *testing.T) {
	mod := New()
	lf := &LabeledFormula{Formula: True}
	lf.SetLineno(Location{Line: 55})
	mod.LabeledConjs = []*LabeledFormula{lf}
	result, _ := AllAssertLinenos(mod)
	found := false
	for _, l := range result {
		if l.Line == 55 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected lineno 55 from conjectures, got %v", result)
	}
}

func TestAllAssertLinenosMultipleActions(t *testing.T) {
	mod := New()
	a1 := NewAssertAction(True)
	loc1 := a1.GetLineno()
	loc1.Line = 10
	a1.SetLineno(loc1)
	seq1 := NewSequence(a1)
	mod.Actions.Set("act1", seq1)

	a2 := NewAssertAction(True)
	loc2 := a2.GetLineno()
	loc2.Line = 20
	a2.SetLineno(loc2)
	seq2 := NewSequence(a2)
	mod.Actions.Set("act2", seq2)

	result, _ := AllAssertLinenos(mod)
	if len(result) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(result), result)
	}
}

// --- Integration-level tests ---

func TestCheckerInterfaceCompliance(t *testing.T) {
	mod := New()

	// All checker types implement the Checker interface
	lf := mod.Cfg.AstCfg.NewLabeledFormula(nil, True)
	checkers := []Checker{
		NewBaseChecker(mod, True, true, true),
		NewConjChecker(mod, lf, 8),
		NewConjAssumer(mod, lf),
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
	mod := New()
	// Reset failures counter
	oldFailures := mod.Cfg.Failures
	result := CheckSafetyInState(mod, false)
	_ = result
	// Restore failures counter (this is an expected failure)
	mod.Cfg.Failures = oldFailures
}

func TestCheckConjsInState(t *testing.T) {
	mod := New()
	mod.LabeledConjs = []*LabeledFormula{
		{Formula: True},
	}
	ApplyConjProofs(mod)
	result := CheckConjsInState(mod, 8, nil)
	if !result {
		t.Error("should pass (stub)")
	}
}

// --- Failures propagation tests ---
//
// Python's ivy_check.py uses a module-level global `failures` counter
// (line 208). All Checker.fail() calls increment it regardless of which
// module copy is active. Go stores Failures on mod.Cfg, so Module.Copy()
// creates a separate counter. CheckModule and CheckSubgoals must propagate
// the delta back after every CheckIsolate(copy) call.

func TestCheckModuleReportsFailures(t *testing.T) {
	// Integration test: exercises the full CheckModule → isoMod := mod.Copy()
	// → CheckIsolate(isoMod) → ConjChecker.Fail() → propagation chain.
	//
	// 3 false conjectures (empty Or = False). ConjChecker inverts (dualizes)
	// each → True → trivially SAT → Fail(). Failures accumulate on
	// isoMod.Cfg.Failures inside the copy; the propagation code at
	// isolate_check.go must bring them back to mod.Cfg.Failures.
	mod := New()
	acfg := mod.Cfg.AstCfg
	for i := 0; i < 3; i++ {
		label := acfg.NewAtom(fmt.Sprintf("false_conj_%d", i))
		lf := acfg.NewLabeledFormula(label, &LogicOr{})
		mod.LabeledConjs = append(mod.LabeledConjs, lf)
	}

	err := CheckModule(mod)
	if err == nil {
		t.Fatal("expected error from CheckModule with false conjectures, got nil")
	}
	if !strings.Contains(err.Error(), "failed checks:") {
		t.Errorf("expected error containing 'failed checks:', got %q", err.Error())
	}
	if mod.Cfg.Failures != 3 {
		t.Errorf("expected mod.Cfg.Failures=3 (propagated from isoMod copy), got %d", mod.Cfg.Failures)
	}
}

func TestIssue74ExternalTraceDoesNotResolveEnvActionToBody(t *testing.T) {
	const src = `#lang ivy1.8

individual x : bool

after init {
    x := false
}

action flip = {
    x := true
}
export flip

conjecture [x_stays_false] ~x
`
	dir := t.TempDir()
	ivyFile := filepath.Join(dir, "issue74_min.ivy")
	if err := os.WriteFile(ivyFile, []byte(src), 0o600); err != nil {
		t.Fatalf("write Ivy fixture: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestIssue74ExternalTraceHelper$", "-test.v")
	cmd.Env = append(os.Environ(),
		"GOIVY_ISSUE74_HELPER=1",
		"GOIVY_ISSUE74_FILE="+ivyFile,
		"XTRACE_OFF=1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("issue74 helper failed: %v\n%s", err, out)
	}
	text := string(out)
	if !strings.Contains(text, "FAIL") {
		t.Fatalf("regression fixture should fail and print a trace, got:\n%s", text)
	}
	if !strings.Contains(text, "searching for a small model... done") {
		t.Fatalf("trace failure should report Python-style small-model shrinking status, got:\n%s", text)
	}
	if strings.Contains(text, "annotation error:") {
		t.Fatalf("external trace resolved the EnvAction annotation against the action body:\n%s", text)
	}
	if !strings.Contains(text, "call flip") {
		t.Fatalf("trace should include the exported action call, got:\n%s", text)
	}
}

func TestIssue74ExternalTraceHelper(t *testing.T) {
	if os.Getenv("GOIVY_ISSUE74_HELPER") != "1" {
		t.Skip("helper process only")
	}
	ivyFile := os.Getenv("GOIVY_ISSUE74_FILE")
	if ivyFile == "" {
		t.Fatal("GOIVY_ISSUE74_FILE is not set")
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(filepath.Dir(ivyFile)); err != nil {
		t.Fatalf("chdir fixture: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	cfg := NewConfig()
	cfg.OptTrace = true
	if err := Start([]string{filepath.Base(ivyFile)}, cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
}

func TestCheckModuleNoFailuresReturnsNil(t *testing.T) {
	mod := New()
	err := CheckModule(mod)
	if err != nil {
		t.Errorf("expected nil error when no conjectures fail, got: %v", err)
	}
}
