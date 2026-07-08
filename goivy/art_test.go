package goivy

import (
	"encoding/json"
	"strings"
	"testing"
)

// --- helpers ---

func artTestModule() *Module {
	return New()
}

func artTestClauses() *Clauses {
	return NewClauses([]Expr{True}, nil, nil)
}

func falseClauses() *Clauses {
	return NewClauses([]Expr{False}, nil, nil)
}

func testState(mod *Module) *State {
	return NewState(mod, artTestClauses())
}

func testGraph() *AnalysisGraph {
	return NewAnalysisGraph(artTestModule())
}

// registerAction registers a no-op action in the graph's action map.
// This is needed because Add() asserts that string-rep actions exist
// in the map, matching Python's assert expr.rep in self.actions.
func registerAction(ag *AnalysisGraph, name string) {
	ag.Actions.Set(name, NewSequence())
}

func addStates(ag *AnalysisGraph, n int) []*State {
	states := make([]*State, n)
	for i := 0; i < n; i++ {
		s := testState(ag.Domain)
		ag.Add(s, nil)
		states[i] = s
	}
	return states
}

// --- State tests ---

func TestArtNewState(t *testing.T) {
	mod := artTestModule()
	s := NewState(mod, artTestClauses())
	if s.ID != -1 {
		t.Errorf("expected ID -1, got %d", s.ID)
	}
	if s.Domain != mod {
		t.Error("domain mismatch")
	}
	if s.Clauses == nil {
		t.Error("clauses should not be nil")
	}
}

func TestArtNewStateAddsEmptyAnnotationLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_interp, ivy_module as im, ivy_logic_utils as lut

mod = im.Module()
state = mod.new_state(lut.true_clauses())
print(json.dumps(type(state.clauses.annot).__name__))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python module.new_state annotation oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python module.new_state annotation oracle %q: %v", out, err)
	}

	state := NewState(New(), TrueClauses(nil))
	if got := TypeName(state.Clauses.Annot); got != want {
		t.Fatalf("NewState annotation differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestStateUpdateComputesFromActionAppProvenance(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	ag.Add(s1, NewActionApp("act", s0))

	if s1.Update != nil {
		t.Fatalf("test setup expected no cached update")
	}
	if s1.Value != nil {
		t.Fatalf("test setup expected no BMC value")
	}

	update := StateUpdate(s1)
	if update == nil {
		t.Fatalf("StateUpdate returned nil for ActionApp provenance")
	}
	if s1.Update != update {
		t.Fatalf("StateUpdate did not cache the computed update")
	}
}

func TestArtStateIsBottom(t *testing.T) {
	mod := artTestModule()
	s := NewState(mod, falseClauses())
	if !s.IsBottom() {
		t.Error("expected state to be bottom")
	}
	s2 := NewState(mod, artTestClauses())
	if s2.IsBottom() {
		t.Error("expected state to not be bottom")
	}
}

func TestArtStateCopy(t *testing.T) {
	mod := artTestModule()
	s := NewState(mod, artTestClauses())
	s.ID = 42
	s.Label = "test"
	cp := s.Copy()
	if cp.ID != 42 || cp.Label != "test" {
		t.Error("copy did not preserve fields")
	}
	cp.ID = 99
	if s.ID == 99 {
		t.Error("copy should be independent")
	}
}

func TestArtStateString(t *testing.T) {
	s := &State{ID: 7}
	str := s.String()
	if str != "State(7)" {
		t.Errorf("expected State(7), got %s", str)
	}
}

func TestArtStateNilClauses(t *testing.T) {
	s := &State{ID: 0}
	if s.IsBottom() {
		t.Error("nil clauses should not be bottom")
	}
}

// --- Expression tests ---

func TestArtActionAppExpr(t *testing.T) {
	s := &State{ID: 0}
	expr := NewActionApp("myaction", s)
	if !IsActionApp(expr) {
		t.Error("expected ActionApp")
	}
	if IsStateJoin(expr) {
		t.Error("should not be StateJoin")
	}
	aa := expr
	if aa.Rep != "myaction" {
		t.Error("wrong rep")
	}
	if len(aa.Args) != 1 || aa.Args[0] != s {
		t.Error("wrong args")
	}
}

func TestArtStateJoinExpr(t *testing.T) {
	s1 := &State{ID: 0}
	s2 := &State{ID: 1}
	expr := NewStateJoin(s1, s2)
	if !IsStateJoin(expr) {
		t.Error("expected StateJoin")
	}
	if IsActionApp(expr) {
		t.Error("should not be ActionApp")
	}
	if len(expr.Args) != 2 {
		t.Error("expected 2 args")
	}
}

func TestArtIsActionAppNil(t *testing.T) {
	if IsActionApp(nil) {
		t.Error("nil should not be ActionApp")
	}
}

func TestArtIsStateJoinNil(t *testing.T) {
	if IsStateJoin(nil) {
		t.Error("nil should not be StateJoin")
	}
}

// --- Counterexample tests ---

func TestArtCounterexample(t *testing.T) {
	cex := &Counterexample{
		Clauses: artTestClauses(),
		State:   &State{ID: 3},
		Msg:     "assertion failure",
	}
	if cex.IsFailed() {
		t.Error("IsFailed should return false")
	}
	str := cex.String()
	if str != "Counterexample: assertion failure" {
		t.Errorf("unexpected string: %s", str)
	}
}

func TestArtCounterexampleNilState(t *testing.T) {
	cex := &Counterexample{Msg: "test"}
	if cex.IsFailed() {
		t.Error("IsFailed should return false regardless")
	}
}

// --- AC tests ---

func TestArtACCreation(t *testing.T) {
	ag := testGraph()
	ac := NewAC(ag, false)
	if ac.NoAdd {
		t.Error("NoAdd should be false")
	}
	if ac.Domain != ag.Domain {
		t.Error("domain mismatch")
	}
}

func TestArtACGet(t *testing.T) {
	ag := testGraph()
	ag.Actions.Set("foo", NewAssumeAction(True))
	ac := NewAC(ag, false)
	if ac.Get("foo") == nil {
		t.Error("expected to find action foo")
	}
	if ac.Get("bar") != nil {
		t.Error("expected nil for unknown action")
	}
}

func TestArtACNewState(t *testing.T) {
	ag := testGraph()
	ac := NewAC(ag, false)
	s := ac.NewState(artTestClauses(), false, nil)
	if s == nil {
		t.Fatal("state should not be nil")
	}
	if len(ag.States) != 1 {
		t.Error("state should have been added to graph")
	}
}

func TestArtACNewStateNoAdd(t *testing.T) {
	ag := testGraph()
	ac := NewAC(ag, true)
	s := ac.NewState(artTestClauses(), false, nil)
	if s == nil {
		t.Fatal("state should not be nil")
	}
	if len(ag.States) != 0 {
		t.Error("state should NOT have been added to graph")
	}
}

func TestArtACNewStateExact(t *testing.T) {
	ag := testGraph()
	ac := NewAC(ag, true)
	s := ac.NewState(artTestClauses(), true, nil)
	if len(s.Unders) != 1 {
		t.Error("exact state should have 1 under-approximation")
	}
}

// --- AnalysisGraph tests ---

func TestArtNewAnalysisGraph(t *testing.T) {
	ag := testGraph()
	if ag.Domain == nil {
		t.Error("domain should not be nil")
	}
	if ag.StateCount() != 0 {
		t.Error("should start with 0 states")
	}
}

func TestArtNewAnalysisGraphNilModule(t *testing.T) {
	ag := NewAnalysisGraph(nil, nil)
	if ag.Domain == nil {
		t.Error("nil module should result in a fresh module")
	}
}

func TestArtAnalysisGraphAdd(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	ag.Add(s, nil)
	if ag.StateCount() != 1 {
		t.Error("expected 1 state")
	}
	if s.ID != 0 {
		t.Errorf("expected ID 0, got %d", s.ID)
	}
}

func TestArtAnalysisGraphAddMultiple(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 5)
	for i, s := range states {
		if s.ID != i {
			t.Errorf("state %d has wrong ID %d", i, s.ID)
		}
	}
	if ag.StateCount() != 5 {
		t.Error("expected 5 states")
	}
}

func TestArtAnalysisGraphLastState(t *testing.T) {
	ag := testGraph()
	if ag.LastState() != nil {
		t.Error("empty graph should return nil")
	}
	states := addStates(ag, 3)
	if ag.LastState() != states[2] {
		t.Error("LastState should return the last state")
	}
}

func TestArtAnalysisGraphAddWithActionApp(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post := testState(ag.Domain)
	ag.Add(post, NewActionApp("act", pre))

	if len(ag.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(ag.Transitions))
	}
	tr := ag.Transitions[0]
	if tr.Pre != pre || tr.Post != post || tr.Label != "act" {
		t.Error("transition fields mismatch")
	}
}

func TestArtAnalysisGraphAddWithJoin(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)

	s3 := testState(ag.Domain)
	ag.Add(s3, NewStateJoin(s1, s2))

	if len(ag.Transitions) != 2 {
		t.Fatalf("expected 2 transitions (one per join arg), got %d", len(ag.Transitions))
	}
	for _, tr := range ag.Transitions {
		if tr.Label != "join" {
			t.Error("join transitions should have label 'join'")
		}
	}
}

func TestArtAnalysisGraphCover(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	ok := ag.Cover(states[0], states[1])
	if !ok {
		t.Error("cover should succeed (stubbed)")
	}
	if len(ag.Covering) != 1 {
		t.Error("expected 1 covering pair")
	}
}

func TestArtAnalysisGraphIsCovered(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 3)
	if ag.IsCovered(states[0]) {
		t.Error("should not be covered initially")
	}
	ag.Cover(states[0], states[1])
	if !ag.IsCovered(states[0]) {
		t.Error("should be covered after covering")
	}
	if ag.IsCovered(states[2]) {
		t.Error("state 2 should not be covered")
	}
}

func TestArtAnalysisGraphUnreachable(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 1)
	if ag.Unreachable(states[0]) {
		t.Error("stub should return false")
	}
}

func TestArtAnalysisGraphTransitionTo(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := testState(ag.Domain)
	ag.Add(post, NewActionApp("act", pre))

	tr := ag.TransitionTo(post)
	if tr == nil {
		t.Fatal("should find transition")
	}
	if tr.Pre != pre {
		t.Error("wrong prestate")
	}

	if ag.TransitionTo(pre) != nil {
		t.Error("pre has no incoming transition")
	}
}

func TestArtAnalysisGraphJoin(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	joined := ag.Join(states[0], states[1], nil)
	if joined == nil {
		t.Fatal("join result should not be nil")
	}
	if ag.StateCount() != 3 {
		t.Error("joined state should be added to graph")
	}
}

func TestArtAnalysisGraphJoinDefaultState2(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	joined := ag.Join(states[0], nil, nil)
	if joined == nil {
		t.Fatal("join result should not be nil")
	}
	// state2 should default to last state = states[1]
}

func TestArtAnalysisGraphRemoveMarkedStates(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 4)
	states[1].ID = -1
	states[3].ID = -1
	ag.RemoveMarkedStates()
	if ag.StateCount() != 2 {
		t.Errorf("expected 2 states, got %d", ag.StateCount())
	}
	for i, s := range ag.States {
		if s.ID != i {
			t.Errorf("state %d has wrong ID %d after renumber", i, s.ID)
		}
	}
}

func TestArtAnalysisGraphReplaceState(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	states[0].Label = "original"

	replacement := testState(ag.Domain)
	replacement.Label = "replaced"
	ag.ReplaceState(states[0], replacement)
	if states[0].Label != "original" {
		t.Error("label should be preserved")
	}
	if states[0].ID != 0 {
		t.Error("ID should be preserved")
	}
}

func TestArtAnalysisGraphUncoveredStates(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 3)
	uncov := ag.UncoveredStates()
	if len(uncov) != 3 {
		t.Errorf("expected 3 uncovered, got %d", len(uncov))
	}
	ag.Cover(states[0], states[1])
	uncov = ag.UncoveredStates()
	if len(uncov) != 2 {
		t.Errorf("expected 2 uncovered, got %d", len(uncov))
	}
}

func TestArtAnalysisGraphUncoveredWithJoin(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	s3 := testState(ag.Domain)
	ag.Add(s3, NewStateJoin(states[0], states[1]))
	// states[0] and states[1] are joined into s3, so they should be
	// marked as "joined" and excluded from uncovered.
	uncov := ag.UncoveredStates()
	if len(uncov) != 1 {
		t.Errorf("expected 1 uncovered (the joined state), got %d", len(uncov))
	}
	if uncov[0] != s3 {
		t.Error("the only uncovered state should be s3")
	}
}

func TestArtAnalysisGraphContext(t *testing.T) {
	ag := testGraph()
	ctx := ag.Context()
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	if ctx.Domain != ag.Domain {
		t.Error("domain mismatch")
	}
}

func TestArtAnalysisGraphExecuteAction(t *testing.T) {
	ag := testGraph()
	act := NewAssumeAction(True)
	ag.Actions.Set("test_action", act)
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post, err := ag.ExecuteAction(artCheckPrecondTrue, "test_action", pre, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post == nil {
		t.Fatal("post state should not be nil")
	}
	if ag.StateCount() != 2 {
		t.Error("should have 2 states after execute")
	}
}

func TestAnalysisGraphAddActionObjectRecordsCanonicalActionName(t *testing.T) {
	mod := New()
	act := NewSequence()
	act.SetLabel("*goivy.LogicSequence@0xdb35e0")
	mod.Actions.Set("ext:connect", act)
	ag := NewAnalysisGraph(mod)
	pre := testState(mod)
	post := testState(mod)
	ag.Add(pre, nil)

	expr := NewActionApp(act, pre)
	ag.Add(post, expr)

	if expr.ActionName != "ext:connect" {
		t.Fatalf("ActionApp.ActionName = %q, want ext:connect", expr.ActionName)
	}
	if post.ActionName != "ext:connect" {
		t.Fatalf("post.ActionName = %q, want ext:connect", post.ActionName)
	}
	if len(ag.Transitions) != 1 {
		t.Fatalf("transition count = %d, want 1", len(ag.Transitions))
	}
	tr := ag.Transitions[0]
	if tr.ActionName != "ext:connect" {
		t.Fatalf("Transition.ActionName = %q, want ext:connect", tr.ActionName)
	}
	if tr.Label != "ext:connect" {
		t.Fatalf("Transition.Label = %q, want ext:connect", tr.Label)
	}
}

func TestAnalysisGraphAddWrappedSequenceRecordsCanonicalActionName(t *testing.T) {
	mod := New()
	act := NewAssumeAction(True)
	mod.Actions.Set("ext:connect", act)
	wrapper := NewSequence(act, NewReturnAction())
	wrapper.SetLabel("*goivy.LogicSequence@0xdeb400")

	ag := NewAnalysisGraph(mod)
	pre := testState(mod)
	post := testState(mod)
	ag.Add(pre, nil)
	expr := NewActionApp(wrapper, pre)
	ag.Add(post, expr)

	if expr.ActionName != "ext:connect" {
		t.Fatalf("ActionApp.ActionName = %q, want ext:connect", expr.ActionName)
	}
	if post.ActionName != "ext:connect" {
		t.Fatalf("post.ActionName = %q, want ext:connect", post.ActionName)
	}
	if len(ag.Transitions) != 1 {
		t.Fatalf("transition count = %d, want 1", len(ag.Transitions))
	}
	if got := ag.Transitions[0].ActionName; got != "ext:connect" {
		t.Fatalf("Transition.ActionName = %q, want ext:connect", got)
	}
}

func TestAnalysisGraphExecuteActionRecordsCanonicalActionName(t *testing.T) {
	mod := New()
	act := NewAssumeAction(True)
	mod.Actions.Set("ext:connect", act)
	ag := NewAnalysisGraph(mod)
	pre := testState(mod)
	ag.Add(pre, nil)

	post, err := ag.ExecuteAction(false, "ext:connect", pre, nil)
	if err != nil {
		t.Fatalf("ExecuteAction: %v", err)
	}
	if post.ActionName != "ext:connect" {
		t.Fatalf("post.ActionName = %q, want ext:connect", post.ActionName)
	}
	if len(ag.Transitions) != 1 {
		t.Fatalf("transition count = %d, want 1", len(ag.Transitions))
	}
	if got := ag.Transitions[0].ActionName; got != "ext:connect" {
		t.Fatalf("Transition.ActionName = %q, want ext:connect", got)
	}
}

func TestAnalysisGraphExecuteEnvActionCreatesModelNamedTransitions(t *testing.T) {
	mod := New()
	mod.Actions.Set("ext:connect", NewAssumeAction(True))
	mod.Actions.Set("ext:disconnect", NewAssumeAction(True))
	mod.PublicActions.Set("ext:connect", true)
	mod.PublicActions.Set("ext:disconnect", true)
	ag := NewAnalysisGraph(mod)
	pre := testState(mod)

	env := BuildEnvAction(mod.Cfg.ActCfg, mod.PublicActions, mod.Actions, "", "")
	post, err := ag.Execute(false, env, pre, nil, "")
	if err != nil {
		t.Fatalf("Execute env action: %v", err)
	}
	if post == nil {
		t.Fatal("Execute env action returned nil post")
	}
	got := make(map[string]bool)
	for _, tr := range ag.Transitions {
		got[tr.ActionName] = true
	}
	if !got["ext:connect"] || !got["ext:disconnect"] || len(got) != 2 {
		t.Fatalf("transition action names = %#v, want ext:connect and ext:disconnect", got)
	}
}

func TestAnalysisGraphAddUnregisteredActionObjectPanics(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	post := testState(ag.Domain)
	ag.Add(pre, nil)
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("Add with unregistered action object did not panic")
		}
		if !strings.Contains(got.(string), "canonical model action name") {
			t.Fatalf("panic = %q, want canonical model action name", got)
		}
	}()
	ag.Add(post, NewActionApp(NewAssumeAction(True), pre))
}

func TestActionAppPointerStringDoesNotBecomeActionName(t *testing.T) {
	app := NewActionApp("*goivy.LogicSequence@0xdb35e0")
	if app.ActionName != "" {
		t.Fatalf("ActionName = %q, want empty for Go implementation string", app.ActionName)
	}
}

func TestActionAppInterpNamePrefersCanonicalActionName(t *testing.T) {
	mod := New()
	act := NewSequence()
	act.SetLabel("*goivy.LogicSequence@0xdb35e0")
	mod.Actions.Set("ext:connect", act)
	pre := NewState(mod, TrueClauses(nil))
	post := NewState(mod, TrueClauses(nil))
	post.Prov = NewActionApp(act, pre)
	post.Prov.(*ActionApp).ActionName = "ext:connect"
	post.ActionName = "ext:connect"

	interp := ArtToInterpState(post)
	atom, ok := interp.Expr.(*Atom)
	if !ok {
		t.Fatalf("interp expr = %T, want *Atom", interp.Expr)
	}
	if atom.Rep != "ext:connect" {
		t.Fatalf("interp action app rep = %q, want ext:connect", atom.Rep)
	}
}

func TestArtAnalysisGraphExecuteActionNotFound(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	_, err := ag.ExecuteAction(artCheckPrecondTrue, "nonexistent", pre, nil)
	if err == nil {
		t.Error("should return error for unknown action")
	}
}

func TestArtAnalysisGraphPostState(t *testing.T) {
	ag := testGraph()
	act := NewAssumeAction(True)
	pre := testState(ag.Domain)
	post, err := ag.PostState(artCheckPrecondFalse, act, pre, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post == nil {
		t.Fatal("post state should not be nil")
	}
	if post.Action != act {
		t.Error("action should be set")
	}
}

func TestArtAnalysisGraphPostStateWithAbstractor(t *testing.T) {
	ag := testGraph()
	act := NewAssumeAction(True)
	pre := testState(ag.Domain)
	called := false
	abs := AbstractorFunc(func(s *State) { called = true })
	_, err := ag.PostState(artCheckPrecondFalse, act, pre, abs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("abstractor should have been called")
	}
}

func TestArtAnalysisGraphCopyPath(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Pred = s0
	s1.Prov = NewActionApp("act", s0)
	ag.Add(s1, s1.Prov)

	other := NewAnalysisGraph(ag.Domain)
	result := ag.CopyPath(s1, other, nil)
	if result == nil {
		t.Fatal("copy path result should not be nil")
	}
	if other.StateCount() < 2 {
		t.Errorf("other graph should have at least 2 states, got %d", other.StateCount())
	}
}

func TestArtAnalysisGraphCopyPathBounded(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Pred = s0
	s1.Prov = NewActionApp("act", s0)
	ag.Add(s1, s1.Prov)

	other := NewAnalysisGraph(ag.Domain)
	bound := 0
	result := ag.CopyPath(s1, other, &bound)
	if result == nil {
		t.Fatal("copy path result should not be nil")
	}
	// With bound=0, should only copy the leaf state
	if other.StateCount() != 1 {
		t.Errorf("expected 1 state with bound=0, got %d", other.StateCount())
	}
}

func TestArtAnalysisGraphBMC(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)

	// BMC with error condition True on a satisfiable state should find
	// a counterexample (True AND True is SAT).
	result := ag.BMC(ag.States[0], True, nil, nil)
	if result == nil {
		t.Error("BMC with True error on SAT state should find counterexample")
	}

	// BMC with error condition False should not find a counterexample
	// (True AND False is UNSAT).
	result2 := ag.BMC(ag.States[0], False, nil, nil)
	if result2 != nil {
		t.Error("BMC with False error should return nil (no counterexample)")
	}
}

func TestArtAnalysisGraphCheckSafety(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)
	res := ag.CheckSafety(true, ag.States[0])
	if !res.Safe {
		t.Error("stub should return safe")
	}
}

func TestArtAnalysisGraphCheckBoundedSafety(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)
	res := ag.CheckBoundedSafety(ag.States[0], nil)
	if !res.Safe {
		t.Error("stub should return safe")
	}
}

func TestArtAnalysisGraphCallAction(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post := ag.CallAction("test", func(sub *AnalysisGraph) {
		s := testState(sub.Domain)
		sub.Add(s, nil)
	}, pre)
	if post == nil {
		t.Fatal("should get a post state")
	}
}

func TestArtAnalysisGraphDecomposeState(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)
	sub := ag.DecomposeState(ag.States[0])
	if sub != nil {
		t.Error("stub should return nil")
	}
}

func TestArtAnalysisGraphFixedpointCandidate(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 3)
	states[0].Label = "A"
	states[1].Label = "A"
	states[2].Label = "B"
	fpc := ag.FixedpointCandidate(nil)
	// After joining, each label maps to a single joined state.
	if fpc["A"] == nil {
		t.Error("expected a joined state for label A, got nil")
	}
	if fpc["B"] == nil {
		t.Error("expected a joined state for label B, got nil")
	}
}

func TestArtAnalysisGraphGetHistory(t *testing.T) {
	ag := testGraph()
	addStates(ag, 1)
	h := ag.GetHistory(ag.States[0], nil)
	if h == nil {
		t.Fatal("history should not be nil")
	}
}

func TestArtAnalysisGraphRecalculate(t *testing.T) {
	ag := testGraph()
	act := NewAssumeAction(True)
	ag.Actions.Set("myact", act)
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := testState(ag.Domain)
	ag.Add(post, NewActionApp("myact", pre))

	tr := ag.Transitions[0]
	result, err := ag.Recalculate(artCheckPrecondFalse, tr, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("recalculate should return a state")
	}
}

// --- AnalysisSubgraph tests ---

func TestArtAnalysisSubgraph(t *testing.T) {
	ag := testGraph()
	sub := NewAnalysisSubgraph("myop", ag)
	if sub.Op != "myop" {
		t.Error("wrong op")
	}
	if sub.Graph != ag {
		t.Error("wrong graph")
	}
	if sub.String() != "myop" {
		t.Errorf("expected 'myop', got '%s'", sub.String())
	}
}

// --- LabelFromAction tests ---

func TestArtLabelFromAction(t *testing.T) {
	act := NewAssumeAction(True)
	label := LabelFromAction(act)
	if label == "" {
		t.Error("label should not be empty")
	}
}

func TestArtLabelFromActionUsesSingularActionLabel(t *testing.T) {
	act := NewAssumeAction(True)
	act.SetLabel("call ext")
	if got := LabelFromAction(act); got != "call ext" {
		t.Fatalf("LabelFromAction() = %q, want %q", got, "call ext")
	}
}

// --- SafetyResult tests ---

func TestArtSafetyResult(t *testing.T) {
	r := &SafetyResult{Safe: true}
	if !r.Safe {
		t.Error("should be safe")
	}
	r2 := &SafetyResult{Safe: false, Cex: &Counterexample{Msg: "fail"}}
	if r2.Safe {
		t.Error("should not be safe")
	}
	if r2.Cex.Msg != "fail" {
		t.Error("wrong msg")
	}
}

// --- Transition tests ---

func TestArtTransition(t *testing.T) {
	pre := &State{ID: 0}
	post := &State{ID: 1}
	act := NewAssumeAction(True)
	tr := Transition{Pre: pre, Op: act, Label: "act", Post: post}
	if tr.Pre != pre || tr.Post != post {
		t.Error("wrong states")
	}
	if tr.Label != "act" {
		t.Error("wrong label")
	}
}

// --- CoveringPair tests ---

func TestArtCoveringPair(t *testing.T) {
	s1 := &State{ID: 0}
	s2 := &State{ID: 1}
	cp := CoveringPair{Covered: s1, Covering: s2}
	if cp.Covered != s1 || cp.Covering != s2 {
		t.Error("wrong states")
	}
}

// --- Integration-style tests ---

func TestArtGraphBuildAndCover(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "step1")
	registerAction(ag, "step2")
	// Build a small graph: s0 -> s1 -> s2, cover s2 by s0
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)

	s1 := testState(ag.Domain)
	ag.Add(s1, NewActionApp("step1", s0))

	s2 := testState(ag.Domain)
	ag.Add(s2, NewActionApp("step2", s1))

	ag.Cover(s2, s0)

	if !ag.IsCovered(s2) {
		t.Error("s2 should be covered")
	}
	if ag.IsCovered(s0) {
		t.Error("s0 should not be covered")
	}

	uncov := ag.UncoveredStates()
	// s0 is uncovered, s1 is uncovered (s2's arg), s2 is covered
	for _, u := range uncov {
		if u.ID == s2.ID {
			t.Error("s2 should not be in uncovered states")
		}
	}
}

func TestArtGraphRemoveMarkedWithTransitions(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	registerAction(ag, "act2")
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	ag.Add(s1, NewActionApp("act", s0))
	s2 := testState(ag.Domain)
	ag.Add(s2, NewActionApp("act2", s1))

	// Mark s1 for deletion
	s1.ID = -1
	ag.RemoveMarkedStates()

	if ag.StateCount() != 2 {
		t.Errorf("expected 2 states, got %d", ag.StateCount())
	}
	// Transition from s0->s1 should be gone (s1 was deleted)
	// Transition from s1->s2 should be gone (s1 was deleted)
	if len(ag.Transitions) != 0 {
		t.Errorf("expected 0 transitions, got %d", len(ag.Transitions))
	}
}

func TestArtGraphExecuteAndTraverse(t *testing.T) {
	ag := testGraph()
	act := NewAssumeAction(True)
	ag.Actions.Set("step", act)
	ag.PublicActions.Set("step", true)

	s0 := testState(ag.Domain)
	ag.Add(s0, nil)

	s1, err := ag.ExecuteAction(artCheckPrecondTrue, "step", s0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s1 == nil {
		t.Fatal("execute should produce a state")
	}

	tr := ag.TransitionTo(s1)
	if tr == nil {
		t.Fatal("should find transition to s1")
	}
	if tr.Pre != s0 {
		t.Error("transition pre should be s0")
	}
}

// --- Fuzz tests ---

func FuzzAnalysisGraphAddRemove(f *testing.F) {
	f.Add(uint8(3), uint8(1))
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(10), uint8(5))

	f.Fuzz(func(t *testing.T, numStates uint8, removeIdx uint8) {
		n := int(numStates) % 20 // cap at 20
		ag := testGraph()
		for i := 0; i < n; i++ {
			s := testState(ag.Domain)
			ag.Add(s, nil)
		}
		if ag.StateCount() != n {
			t.Errorf("expected %d states, got %d", n, ag.StateCount())
		}
		if n > 0 {
			idx := int(removeIdx) % n
			ag.States[idx].ID = -1
			ag.RemoveMarkedStates()
			if ag.StateCount() != n-1 {
				t.Errorf("expected %d states after remove, got %d", n-1, ag.StateCount())
			}
			// Check IDs are contiguous
			for i, s := range ag.States {
				if s.ID != i {
					t.Errorf("state at index %d has ID %d", i, s.ID)
				}
			}
		}
	})
}

func FuzzAnalysisGraphCoverUncover(f *testing.F) {
	f.Add(uint8(5), uint8(0), uint8(1))
	f.Add(uint8(2), uint8(0), uint8(1))

	f.Fuzz(func(t *testing.T, numStates uint8, coverIdx uint8, coverByIdx uint8) {
		n := int(numStates)%20 + 2 // at least 2
		ag := testGraph()
		for i := 0; i < n; i++ {
			s := testState(ag.Domain)
			ag.Add(s, nil)
		}
		ci := int(coverIdx) % n
		cbi := int(coverByIdx) % n
		if ci == cbi {
			cbi = (cbi + 1) % n
		}
		ag.Cover(ag.States[ci], ag.States[cbi])
		if !ag.IsCovered(ag.States[ci]) {
			t.Error("covered state should be covered")
		}
		if ci != cbi && ag.IsCovered(ag.States[cbi]) {
			// cbi might also be covered if ci covers cbi in some other pair,
			// but we only added one pair, so it shouldn't be.
			// Only check if cbi wasn't also used as covered.
		}
		uncov := ag.UncoveredStates()
		for _, u := range uncov {
			if u.ID == ci {
				t.Error("covered state should not be in uncovered list")
			}
		}
	})
}

// --- Z3-backed method tests ---

func TestArtCoverWithZ3Implication(t *testing.T) {
	ag := testGraph()

	// State with True clauses should be covered by another True-clauses state
	// because True => True.
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)
	if !ag.Cover(s1, s2) {
		t.Error("True should imply True")
	}
}

func TestArtCoverFalseImpliesAnything(t *testing.T) {
	ag := testGraph()

	// False => anything should be true.
	s1 := NewState(ag.Domain, falseClauses())
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)
	if !ag.Cover(s1, s2) {
		t.Error("False should imply anything (bottom covers everything)")
	}
}

func TestArtCoverNilClausesFails(t *testing.T) {
	ag := testGraph()
	s1 := &State{ID: -1, Domain: ag.Domain}
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)
	if ag.Cover(s1, s2) {
		t.Error("nil clauses should not cover")
	}
}

func TestArtUnreachableFalseState(t *testing.T) {
	ag := testGraph()
	s := NewState(ag.Domain, falseClauses())
	ag.Add(s, nil)
	if !ag.Unreachable(s) {
		t.Error("false state should be unreachable")
	}
}

func TestArtUnreachableTrueState(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	ag.Add(s, nil)
	if ag.Unreachable(s) {
		t.Error("true state should not be unreachable")
	}
}

func TestArtUnreachableNilClauses(t *testing.T) {
	ag := testGraph()
	s := &State{ID: -1, Domain: ag.Domain}
	ag.Add(s, nil)
	if ag.Unreachable(s) {
		t.Error("nil clauses should not be unreachable")
	}
}

func TestArtUnreachableUsesModuleOrderBackgroundTheoryLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_ast as ia
from ivy import ivy_interp
from ivy import ivy_logic as il
from ivy import ivy_logic_utils as lu
from ivy import ivy_module as im

mod = im.Module()
mod.labeled_axioms.append(ia.LabeledFormula(ia.Atom('ax'), il.Or()))
mod.update_theory()
state = mod.new_state(lu.true_clauses())
false_state = ivy_interp.State(mod, lu.false_clauses())
print(json.dumps({"order": bool(mod.order(state, false_state))}, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python module.order unreachable oracle failed: %v\n%s", err, out)
	}
	var want struct {
		Order bool `json:"order"`
	}
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python module.order unreachable oracle %q: %v", out, err)
	}
	if !want.Order {
		t.Fatal("python module.order oracle no longer covers true by false under inconsistent background theory")
	}

	mod := New()
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.Cfg.AstCfg.NewLabeledFormula(mod.Cfg.AstCfg.NewAtom("ax"), False))
	mod.UpdateTheory()
	ag := NewAnalysisGraph(mod)
	s := NewState(mod, TrueClauses(nil))
	ag.Add(s, nil)

	if !ag.Unreachable(s) {
		t.Fatal("Unreachable ignored module background theory; Python module.order covers the state by false")
	}
	if s.Clauses == nil || !s.Clauses.IsFalse() {
		t.Fatalf("Unreachable should collapse the state to false clauses, got %v", s.Clauses)
	}
}

func TestArtJoinStatesDisjunction(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := NewState(ag.Domain, falseClauses())
	ag.Add(s2, nil)

	// Join of True and False should be satisfiable (True OR False = True).
	joined := ag.JoinStates(s1, s2, nil)
	if joined == nil {
		t.Fatal("joined state should not be nil")
	}
	if joined.JoinOf == nil || len(joined.JoinOf) != 2 {
		t.Error("JoinOf should have 2 entries")
	}
	if joined.Clauses == nil {
		t.Fatal("joined clauses should not be nil")
	}
}

func TestArtCheckSafetyNoAssertions(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	ag.Add(s, nil)
	res := ag.CheckSafety(true, s)
	if !res.Safe {
		t.Error("no assertions should mean safe")
	}
}

func TestArtCheckSafetySatisfiedAssertion(t *testing.T) {
	ag := testGraph()
	// Add an assertion that is True.
	acfg := NewAstConfig()
	ag.Assertions = append(ag.Assertions, acfg.NewLabeledFormula(nil, True))
	s := testState(ag.Domain)
	ag.Add(s, nil)
	res := ag.CheckSafety(true, s)
	if !res.Safe {
		t.Error("True assertion on True state should be safe")
	}
}

func TestArtGetHistoryWithPredecessor(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Pred = s0
	ag.Add(s1, NewActionApp("act", s0))

	h := ag.GetHistory(s1, nil)
	if h == nil {
		t.Fatal("history should not be nil")
	}
}

func TestArtGetHistoryBounded(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Pred = s0
	ag.Add(s1, NewActionApp("act", s0))

	bound := 0
	h := ag.GetHistory(s1, &bound)
	if h == nil {
		t.Fatal("history should not be nil")
	}
	// With bound 0, should stop at s1 (no predecessor expansion).
	if len(h.Actions) != 0 {
		t.Error("bound=0 should produce no action steps")
	}
}

// Verify that unused imports are used.
var _ = NullUpdate
