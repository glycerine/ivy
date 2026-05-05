package goivy

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

var interpTestAstCfg = NewAstConfig()

// ---------------------------------------------------------------------------
// StateValue tests
// ---------------------------------------------------------------------------

func TestNewStateValueDefaults(t *testing.T) {
	sv := NewStateValue(nil, nil, nil)
	if sv.Moded != nil {
		t.Errorf("expected nil Moded, got %v", sv.Moded)
	}
	if sv.Clauses == nil {
		t.Fatal("expected non-nil Clauses")
	}
	if !sv.Clauses.IsTrue() {
		t.Error("expected TrueClauses for default Clauses")
	}
	if sv.Precond == nil {
		t.Fatal("expected non-nil Precond")
	}
	if !sv.Precond.IsFalse() {
		t.Error("expected FalseClauses for default Precond")
	}
}

func TestNewStateValueExplicit(t *testing.T) {
	cls := NewClauses([]Expr{True}, nil, nil)
	pre := FalseClauses(nil)
	sv := NewStateValue([]string{"x", "y"}, cls, pre)
	if len(sv.Moded) != 2 {
		t.Errorf("expected 2 moded symbols, got %d", len(sv.Moded))
	}
	if sv.Clauses != cls {
		t.Error("expected same Clauses pointer")
	}
	if sv.Precond != pre {
		t.Error("expected same Precond pointer")
	}
}

func TestTopStateValue(t *testing.T) {
	sv := TopStateValue()
	if !sv.Clauses.IsTrue() {
		t.Error("TopStateValue should have true clauses")
	}
}

func TestBottomStateValue(t *testing.T) {
	sv := BottomStateValue()
	if !sv.Clauses.IsFalse() {
		t.Error("BottomStateValue should have false clauses")
	}
}

// ---------------------------------------------------------------------------
// State tests
// ---------------------------------------------------------------------------

func TestNewStateDefaults(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	if s.Domain == nil {
		t.Fatal("expected non-nil Domain")
	}
	if s.InScope == nil {
		t.Fatal("expected non-nil InScope")
	}
	if s.Clauses == nil {
		t.Fatal("expected non-nil Clauses")
	}
	if !s.Clauses.IsTrue() {
		t.Error("default state should have TrueClauses")
	}
}

func TestNewStateWithDomain(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "label1")
	if s.Domain != m {
		t.Error("expected same domain")
	}
	if s.Label != "label1" {
		t.Errorf("expected label 'label1', got %q", s.Label)
	}
}

func TestStateValueRoundtrip(t *testing.T) {
	sv := NewStateValue([]string{"a"}, TrueClauses(nil), FalseClauses(nil))
	s := NewInterpState(nil, sv, nil, "")
	got := s.Value()
	if len(got.Moded) != 1 || got.Moded[0] != "a" {
		t.Errorf("Moded roundtrip failed: %v", got.Moded)
	}
}

func TestSetValue(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	sv := NewStateValue([]string{"x"}, FalseClauses(nil), TrueClauses(nil))
	s.SetValue(sv)
	if !s.Clauses.IsFalse() {
		t.Error("SetValue should set Clauses")
	}
	if len(s.Moded) != 1 {
		t.Error("SetValue should set Moded")
	}
}

func TestStateIsBottom(t *testing.T) {
	s := NewInterpState(nil, BottomStateValue(), nil, "")
	if !s.IsBottom() {
		t.Error("state with FalseClauses should be bottom")
	}
	s2 := NewInterpState(nil, TopStateValue(), nil, "")
	if s2.IsBottom() {
		t.Error("state with TrueClauses should not be bottom")
	}
}

func TestStateToFormula(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	f := s.ToFormula()
	if f == nil {
		t.Error("ToFormula should not return nil")
	}
}

func TestStateString(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	str := s.String()
	if str == "" {
		t.Error("String should not be empty")
	}
}

func TestStateStringNilClauses(t *testing.T) {
	s := &InterpState{}
	str := s.String()
	if !strings.Contains(str, "nil") {
		t.Errorf("String for nil clauses should mention nil, got %q", str)
	}
}

func TestStatePredAndUpdate(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	if s.Pred() != nil {
		t.Error("initial Pred should be nil")
	}
	if s.Update() != nil {
		t.Error("initial Update should be nil")
	}
	pred := NewInterpState(nil, nil, nil, "pred")
	s.SetPred(pred)
	if s.Pred() != pred {
		t.Error("SetPred should set predecessor")
	}
	u := NullUpdate()
	s.SetUpdate(u)
	if s.Update() != u {
		t.Error("SetUpdate should set update")
	}
}

func TestStateConjs(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	if s.Conjs() != nil {
		t.Error("initial Conjs should be nil")
	}
	conjs := []*Clauses{TrueClauses(nil)}
	s.SetConjs(conjs)
	got := s.Conjs()
	if len(got) != 1 {
		t.Errorf("expected 1 conjecture, got %d", len(got))
	}
}

func TestStateUnders(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	if s.Unders() != nil {
		t.Error("initial Unders should be nil")
	}
	under := NewInterpState(m, nil, nil, "under1")
	s.SetUnders([]*InterpState{under})
	got := s.Unders()
	if len(got) != 1 {
		t.Errorf("expected 1 under, got %d", len(got))
	}
}

func TestStateAction(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	if s.Action != nil {
		t.Error("initial Action should be nil")
	}
	if s.ActionName != "" {
		t.Error("initial ActionName should be empty")
	}
}

func TestStateJoinOf(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	if s.JoinOf != nil {
		t.Error("initial JoinOf should be nil")
	}
}

// ---------------------------------------------------------------------------
// WrapState / UnwrapState tests
// ---------------------------------------------------------------------------

func TestWrapUnwrapState(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "test")
	wrapped := WrapState(s)
	if wrapped == nil {
		t.Fatal("WrapState returned nil")
	}
	unwrapped := UnwrapState(wrapped)
	if unwrapped != s {
		t.Error("UnwrapState should return original state")
	}
}

func TestUnwrapStateNonState(t *testing.T) {
	atom := interpTestAstCfg.NewAtom("foo")
	if UnwrapState(atom) != nil {
		t.Error("UnwrapState of non-state should return nil")
	}
}

func TestStateNodeString(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	sn := &stateNode{state: s}
	str := sn.String()
	if str == "" {
		t.Error("stateNode String should not be empty")
	}
}

func TestStateNodeNilState(t *testing.T) {
	sn := &stateNode{state: nil}
	str := sn.String()
	if !strings.Contains(str, "nil") {
		t.Errorf("stateNode String for nil should mention nil, got %q", str)
	}
}

/*
// ---------------------------------------------------------------------------
// EvalContext tests
// ---------------------------------------------------------------------------

func TestNewEvalContext(t *testing.T) {
	ec := NewEvalContext(true)
	if !ec.Check {
		t.Error("Check should be true")
	}
	ec2 := NewEvalContext(false)
	if ec2.Check {
		t.Error("Check should be false")
	}
}

func TestEvalContextEnterExit(t *testing.T) {
	origCtx := CurrentContext()
	checkPrecond := true
	ec := NewEvalContext(false)
	ec.Enter()
	if CurrentContext() != ec {
		t.Error("Enter should set current context")
	}
	ec.Exit()
	if CurrentContext() != origCtx {
		t.Error("Exit should restore original context")
	}
}

func TestEvalContextNested(t *testing.T) {
	origCtx := CurrentContext()
	ec1 := NewEvalContext(false)
	ec2 := NewEvalContext(true)
	ec1.Enter()
	ec2.Enter()
	if CurrentContext() != ec2 {
		t.Error("nested Enter should set innermost context")
	}
	ec2.Exit()
	if CurrentContext() != ec1 {
		t.Error("exit inner should restore outer")
	}
	ec1.Exit()
	if CurrentContext() != origCtx {
		t.Error("exit outer should restore original")
	}
}

func TestDefaultContextCheckIsTrue(t *testing.T) {
	// Use a fresh InterpConfig to avoid interference with other tests.
	ic := NewInterpConfig()
	if !ic.CurrentContext().Check {
		t.Error("default context Check should be true")
	}
}
*/

// ---------------------------------------------------------------------------
// Expression helper tests
// ---------------------------------------------------------------------------

func TestIsActionApp(t *testing.T) {
	atom1 := interpTestAstCfg.NewAtom("act", interpTestAstCfg.NewAtom("state"))
	if !IsInterpActionApp(atom1) {
		t.Error("Atom with 1 arg should be action app")
	}
	atom0 := interpTestAstCfg.NewAtom("state")
	if IsInterpActionApp(atom0) {
		t.Error("Atom with 0 args should not be action app")
	}
	or := interpTestAstCfg.NewOr()
	if IsInterpActionApp(or) {
		t.Error("Or should not be action app")
	}
}

func TestIsStateJoin(t *testing.T) {
	or := interpTestAstCfg.NewOr()
	if !IsInterpStateJoin(or) {
		t.Error("Or should be state join")
	}
	atom := interpTestAstCfg.NewAtom("x")
	if IsInterpStateJoin(atom) {
		t.Error("Atom should not be state join")
	}
}

func TestActionApp(t *testing.T) {
	result := InterpActionApp(interpTestAstCfg, "doAction", interpTestAstCfg.NewAtom("s"))
	if result.Rep != "doAction" {
		t.Errorf("expected rep 'doAction', got %q", result.Rep)
	}
	if len(result.Terms) != 1 {
		t.Errorf("expected 1 term, got %d", len(result.Terms))
	}
}

func TestStateJoinFunc(t *testing.T) {
	result := InterpStateJoin(interpTestAstCfg, interpTestAstCfg.NewAtom("a"), interpTestAstCfg.NewAtom("b"))
	if len(result.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(result.Terms))
	}
}

func TestIsStateSymbol(t *testing.T) {
	atom := interpTestAstCfg.NewAtom("s")
	if !IsStateSymbol(atom) {
		t.Error("Atom with 0 args should be state symbol")
	}
	atom2 := interpTestAstCfg.NewAtom("act", interpTestAstCfg.NewAtom("s"))
	if IsStateSymbol(atom2) {
		t.Error("Atom with args should not be state symbol")
	}
}

func TestStateEquation(t *testing.T) {
	eq := StateEquation(interpTestAstCfg, interpTestAstCfg.NewAtom("lhs"), interpTestAstCfg.NewAtom("rhs"))
	if eq == nil {
		t.Fatal("StateEquation should not return nil")
	}
	if eq.Defines() != "lhs" {
		t.Errorf("expected defines 'lhs', got %q", eq.Defines())
	}
}

func TestStatesInExpr(t *testing.T) {
	s1 := NewInterpState(nil, nil, nil, "s1")
	s2 := NewInterpState(nil, nil, nil, "s2")
	expr := interpTestAstCfg.NewOr(WrapState(s1), WrapState(s2))
	states := StatesInExpr(expr)
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
}

func TestStatesInExprEmpty(t *testing.T) {
	expr := interpTestAstCfg.NewAtom("foo")
	states := StatesInExpr(expr)
	if len(states) != 0 {
		t.Errorf("expected 0 states, got %d", len(states))
	}
}

// ---------------------------------------------------------------------------
// Error type tests
// ---------------------------------------------------------------------------

func TestIvyActionFailedError(t *testing.T) {
	err := &IvyActionFailedError{
		ActionName: "do_something",
		Msg:        "precondition of \"do_something\" failed",
	}
	msg := err.Error()
	if !strings.Contains(msg, "do_something") {
		t.Errorf("error message should contain action name, got %q", msg)
	}
}

func TestIvyActionFailedErrorDefaultMsg(t *testing.T) {
	err := &IvyActionFailedError{
		ActionName: "act",
	}
	msg := err.Error()
	if !strings.Contains(msg, "act") {
		t.Errorf("error message should contain action name, got %q", msg)
	}
}

func TestNewIvyActionFailedError(t *testing.T) {
	m := New()
	state := NewInterpState(m, nil, nil, "")
	seq := NewSequence()
	err := NewIvyActionFailedError(
		interpTestAstCfg.NewAtom("test"),
		"myAction",
		seq,
		state,
		TrueClauses(nil),
		nil,
	)
	if err.ActionName != "myAction" {
		t.Errorf("expected 'myAction', got %q", err.ActionName)
	}
	if err.ErrorState == nil {
		t.Fatal("ErrorState should not be nil")
	}
	if err.ErrorState.Pred() != state {
		t.Error("ErrorState pred should be the original state")
	}
}

func TestUnsatCoreWithInterpolant(t *testing.T) {
	core := NewClauses([]Expr{True}, nil, nil)
	itp := NewClauses([]Expr{False}, nil, nil)
	err := &UnsatCoreWithInterpolant{Core: core, Itp: itp}
	msg := err.Error()
	if !strings.Contains(msg, "interpolant") {
		t.Errorf("error should mention interpolant, got %q", msg)
	}
}

// ---------------------------------------------------------------------------
// ConcretePost / ConcreteJoin tests
// ---------------------------------------------------------------------------

func TestConcretePostBasic(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	upd := NullUpdate()
	result, err := ConcretePost(true, upd, s, nil)
	if err != nil {
		t.Fatalf("ConcretePost returned error: %v", err)
	}
	if result.Pred() != s {
		t.Error("result pred should be original state")
	}
	if result.Update() != upd {
		t.Error("result update should be the provided update")
	}
}

func TestConcretePostNilDomain(t *testing.T) {
	s := &InterpState{}
	upd := NullUpdate()
	_, err := ConcretePost(true, upd, s, nil)
	if err == nil {
		t.Error("expected error for nil domain")
	}
}

func TestConcreteJoinBasic(t *testing.T) {
	m := New()
	s1 := NewInterpState(m, nil, nil, "")
	s2 := NewInterpState(m, nil, nil, "")
	result, err := ConcreteJoin(s1, s2)
	if err != nil {
		t.Fatalf("ConcreteJoin returned error: %v", err)
	}
	if len(result.JoinOf) != 2 {
		t.Errorf("expected 2 JoinOf, got %d", len(result.JoinOf))
	}
}

func TestConcreteJoinNilDomain(t *testing.T) {
	s1 := &InterpState{}
	s2 := NewInterpState(nil, nil, nil, "")
	_, err := ConcreteJoin(s1, s2)
	if err == nil {
		t.Error("expected error for nil domain")
	}
}

// ---------------------------------------------------------------------------
// EvalAction tests
// ---------------------------------------------------------------------------

func TestEvalActionDirect(t *testing.T) {
	seq := NewSequence()
	act, err := EvalAction(seq, nil)
	if err != nil {
		t.Fatalf("EvalAction returned error: %v", err)
	}
	if act != seq {
		t.Error("EvalAction should return the action directly")
	}
}

func TestEvalActionFromModule(t *testing.T) {
	m := New()
	seq := NewSequence()
	m.Actions.Set("myAct", seq)
	act, err := EvalAction("myAct", m)
	if err != nil {
		t.Fatalf("EvalAction returned error: %v", err)
	}
	if act != seq {
		t.Error("EvalAction should return the action from module")
	}
}

func TestEvalActionNotFound(t *testing.T) {
	m := New()
	_, err := EvalAction("missing", m)
	if err == nil {
		t.Error("expected error for missing action")
	}
}

// TestEvalActionNotAction removed: mod.Actions is now map[string]module.Action,
// so non-Action values cannot be inserted.

func TestEvalActionUnsupportedType(t *testing.T) {
	_, err := EvalAction(42, nil)
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

// ---------------------------------------------------------------------------
// EvalState / EvalStateAtom tests
// ---------------------------------------------------------------------------

func TestEvalStateAtomWrappedState(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	wrapped := WrapState(s)
	result, err := EvalStateAtom(wrapped, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != s {
		t.Error("should return original state")
	}
}

func TestEvalStateAtomTrue(t *testing.T) {
	m := New()
	trueNode := interpTestAstCfg.NewAnd() // empty And = true
	result, err := EvalStateAtom(trueNode, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Clauses.IsTrue() {
		t.Error("true atom should produce true clauses")
	}
}

func TestEvalStateAtomFalse(t *testing.T) {
	m := New()
	falseNode := interpTestAstCfg.NewOr() // empty Or = false
	result, err := EvalStateAtom(falseNode, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Clauses.IsFalse() {
		t.Error("false atom should produce false clauses")
	}
}

func TestEvalStateAtomSymbol(t *testing.T) {
	m := New()
	sym := interpTestAstCfg.NewAtom("s")
	_, err := EvalStateAtom(sym, m)
	if err == nil {
		t.Error("state symbol lookup should fail (not implemented)")
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestJoinUnders(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	result := JoinUnders(s)
	if result == nil {
		t.Fatal("JoinUnders should not return nil")
	}
}

func TestAddUnder(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	cls := TrueClauses(nil)
	under := AddUnder(s, cls, nil, nil)
	if under == nil {
		t.Fatal("AddUnder should return a state")
	}
	if len(s.Unders()) != 1 {
		t.Errorf("expected 1 under, got %d", len(s.Unders()))
	}
}

func TestAddUnderWithPred(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	pred := NewInterpState(m, nil, nil, "pred")
	under := AddUnder(s, TrueClauses(nil), pred, "universe")
	if under.Pred() != pred {
		t.Error("under pred should be set")
	}
	if under.Universe != "universe" {
		t.Error("under universe should be set")
	}
}

func TestReachStateNoPred(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	result := ReachState(s, nil)
	if result != nil {
		t.Error("ReachState without pred should return nil")
	}
}

func TestReachStateRecordsTaggedModelUnderLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act
from ivy import ivy_interp as interp
from ivy import ivy_logic as lg
from ivy import ivy_logic_utils as lu
from ivy import ivy_module as im

mod = im.Module()
im.module = mod
p = lg.Symbol('p', lg.RelationSort([]))
mod.relations[p] = p.sort
pred = interp.State(mod, lu.true_clauses())
mod.unders = [mod.new_state(lu.false_clauses()), mod.new_state(lu.true_clauses())]
state = interp.State(mod, lu.Clauses([p]))
state.pred = pred
state.update = act.AssignAction(p, lg.And()).update(mod, {})
post = interp.reach_state(state)
text = str(post.clauses) if post is not None else ''
print(json.dumps({
    "post_none": post is None,
    "unders_len": len(mod.unders),
    "pred_index": mod.unders.index(post.pred) if post is not None and hasattr(post, 'pred') else -1,
    "has_universe": hasattr(post, 'universe') if post is not None else False,
    "universe_keys": [str(k) for k in post.universe.keys()] if post is not None and hasattr(post, 'universe') else [],
    "mentions_new_p": "new_p" in text,
    "mentions_p": "p" in text,
}, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python ReachState oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		PostNone     bool     `json:"post_none"`
		UndersLen    int      `json:"unders_len"`
		PredIndex    int      `json:"pred_index"`
		HasUniverse  bool     `json:"has_universe"`
		UniverseKeys []string `json:"universe_keys"`
		MentionsNewP bool     `json:"mentions_new_p"`
		MentionsP    bool     `json:"mentions_p"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python ReachState oracle %q: %v", out, err)
	}
	if want.PostNone || want.UndersLen != 3 || want.PredIndex != 1 || !want.HasUniverse || len(want.UniverseKeys) != 0 || want.MentionsNewP || !want.MentionsP {
		t.Fatalf("python ReachState oracle changed: %+v", want)
	}

	mod := New()
	p := NewConst("p", Boolean)
	mod.Relations.Set("p", Boolean)
	pred := NewInterpState(mod, nil, nil, "")
	under0 := NewStateFromClauses(mod, FalseClauses(nil))
	under1 := NewStateFromClauses(mod, TrueClauses(nil))
	pred.SetUnders([]*InterpState{under0, under1})

	state := NewStateFromClauses(mod, NewClauses([]Expr{p}, nil, nil))
	state.SetPred(pred)
	state.SetUpdate(GetUpdate(NewAssignAction(p, True), &UpdateContext{
		Domain: mod,
		PVars:  nil,
		ActCfg: NewActionsConfig(),
	}))

	post := ReachState(state, nil)
	gotPostNone := post == nil
	gotUndersLen := len(state.Unders())
	gotPredIndex := -1
	if post != nil {
		for i, u := range []*InterpState{under0, under1} {
			if post.Pred() == u {
				gotPredIndex = i
			}
		}
	}
	gotHasUniverse := post != nil && post.Universe != nil
	var gotUniverseKeys []string
	if post != nil {
		if universe, ok := post.Universe.(map[string][]Expr); ok {
			for k := range universe {
				gotUniverseKeys = append(gotUniverseKeys, k)
			}
		}
	}
	gotText := ""
	if post != nil && post.Clauses != nil {
		gotText = post.Clauses.String()
	}
	gotMentionsNewP := strings.Contains(gotText, "new_p")
	gotMentionsP := strings.Contains(gotText, "p")

	if gotPostNone != want.PostNone ||
		gotUndersLen != want.UndersLen ||
		gotPredIndex != want.PredIndex ||
		gotHasUniverse != want.HasUniverse ||
		fmt.Sprint(gotUniverseKeys) != fmt.Sprint(want.UniverseKeys) ||
		gotMentionsNewP != want.MentionsNewP ||
		gotMentionsP != want.MentionsP {
		t.Fatalf("Go ReachState differs from Python\nwant none=%v unders=%d pred=%d universe=%v keys=%v new_p=%v p=%v\ngot  none=%v unders=%d pred=%d universe=%v keys=%v new_p=%v p=%v\nclauses: %s",
			want.PostNone, want.UndersLen, want.PredIndex, want.HasUniverse, want.UniverseKeys, want.MentionsNewP, want.MentionsP,
			gotPostNone, gotUndersLen, gotPredIndex, gotHasUniverse, gotUniverseKeys, gotMentionsNewP, gotMentionsP, gotText)
	}
}

func TestReachStateFromPredNoPred(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	result, err := ReachStateFromPred(s, nil)
	if err == nil || !strings.Contains(err.Error(), "name 'pre' is not defined") {
		t.Fatalf("ReachStateFromPred without pred should match Python's undefined pre failure, got result=%v err=%v", result, err)
	}
	if result != nil {
		t.Error("ReachStateFromPred without pred should return nil")
	}
}

func TestReachStateFromPredUnreachableRaisesUndefinedPreLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act
from ivy import ivy_interp as interp
from ivy import ivy_logic as lg
from ivy import ivy_logic_utils as lu
from ivy import ivy_module as im

mod = im.Module()
im.module = mod
p = lg.Symbol('p', lg.RelationSort([]))
mod.relations[p] = p.sort
pred = interp.State(mod, lu.true_clauses())
state = interp.State(mod, lu.Clauses([p]))
state.pred = pred
state.update = act.AssignAction(p, lg.And()).update(mod, {})

try:
    post = interp.reach_state_from_pred(state)
    res = {"errored": False, "type": "", "message": "", "post_none": post is None}
except Exception as err:
    res = {"errored": True, "type": type(err).__name__, "message": str(err), "post_none": True}

print(json.dumps(res, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python ReachStateFromPred oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		Errored  bool   `json:"errored"`
		Type     string `json:"type"`
		Message  string `json:"message"`
		PostNone bool   `json:"post_none"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python ReachStateFromPred oracle %q: %v", out, err)
	}
	if !want.Errored || want.Type != "NameError" || !strings.Contains(want.Message, "name 'pre' is not defined") || !want.PostNone {
		t.Fatalf("python ReachStateFromPred oracle changed: %+v", want)
	}

	mod := New()
	p := NewConst("p", Boolean)
	mod.Relations.Set("p", Boolean)
	pred := NewInterpState(mod, nil, nil, "")
	state := NewStateFromClauses(mod, NewClauses([]Expr{p}, nil, nil))
	state.SetPred(pred)
	state.SetUpdate(GetUpdate(NewAssignAction(p, True), &UpdateContext{
		Domain: mod,
		PVars:  nil,
		ActCfg: NewActionsConfig(),
	}))

	got, gotErr := ReachStateFromPred(state, nil)
	if got != nil || gotErr == nil ||
		!strings.Contains(gotErr.Error(), "name 'pre' is not defined") {
		t.Fatalf("Go ReachStateFromPred differs from Python\nwant error %s\n got result=%v err=%v", want.Message, got, gotErr)
	}
}

func TestUndecidedConjectures(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	// TrueClauses state implies TrueClauses conjecture, so it should
	// be decided (not undecided). Use FalseClauses as conjecture to
	// get an undecided one (True does not imply False).
	conjs := []*Clauses{FalseClauses(nil)}
	s.SetConjs(conjs)
	result := UndecidedConjectures(s)
	if len(result) != 1 {
		t.Errorf("expected 1 undecided conjecture, got %d", len(result))
	}
}

func TestFilterConjectures(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	conjs := []*Clauses{TrueClauses(nil)}
	s.SetConjs(conjs)
	lost := FilterConjectures(s, TrueClauses(nil))
	if len(lost) != 0 {
		t.Errorf("stub should lose no conjectures, got %d", len(lost))
	}
}

func TestCaseConjecture(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	// Smoke test for the rewritten CaseConjecture, which now delegates
	// to actions.InterpolantCase (Python ivy_interp.py:325-337). With an
	// empty state and TrueClauses input, the call should not panic.
	// Deeper coverage requires real under-approximations.
	_ = CaseConjecture(s, TrueClauses(nil))
}

func TestDiagram(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	// TrueClauses is satisfiable, so Diagram should return non-nil.
	result := Diagram(s, TrueClauses(nil), nil, nil, true, true)
	if result == nil {
		t.Error("Diagram of satisfiable clauses should return non-nil")
	}
	// FalseClauses is unsatisfiable, so Diagram should return nil.
	result2 := Diagram(s, FalseClauses(nil), nil, nil, true, true)
	if result2 != nil {
		t.Error("Diagram of unsatisfiable clauses should return nil")
	}
}

func TestDiagramHonorsUpwardCloseLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_interp as interp
from ivy import ivy_logic as lg
from ivy import ivy_logic_utils as lu
from ivy import ivy_module as im

mod = im.Module()
im.module = mod
sort_s = lg.UninterpretedSort('S')
X = lg.Variable('X', sort_s)
p = lg.Symbol('p', lg.RelationSort([sort_s]))
mod.relations[p] = p.sort
state = interp.State(mod)
clauses = lu.Clauses([lg.Exists([X], p(X))])
diag_open = interp.diagram(state, clauses, upward_close=True)
diag_closed = interp.diagram(state, clauses, upward_close=False)
print(json.dumps({
    "open_fmlas": len(diag_open.fmlas) if diag_open is not None else -1,
    "closed_fmlas": len(diag_closed.fmlas) if diag_closed is not None else -1,
    "closed_has_eq": "=" in str(diag_closed),
    "strings_equal": str(diag_open) == str(diag_closed),
}, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python Diagram oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		OpenFmlas    int  `json:"open_fmlas"`
		ClosedFmlas  int  `json:"closed_fmlas"`
		ClosedHasEq  bool `json:"closed_has_eq"`
		StringsEqual bool `json:"strings_equal"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python Diagram oracle %q: %v", out, err)
	}
	if want.OpenFmlas != 1 || want.ClosedFmlas != 2 || !want.ClosedHasEq || want.StringsEqual {
		t.Fatalf("python Diagram oracle changed: %+v", want)
	}

	mod := New()
	sortS := &UninterpretedSort{Name: "S"}
	if err := mod.Sig.AddSort(sortS); err != nil {
		t.Fatalf("AddSort(S): %v", err)
	}
	p, err := mod.Sig.AddSymbol("p", LogicRelationSort([]Sort{sortS}))
	if err != nil {
		t.Fatalf("AddSymbol(p): %v", err)
	}
	mod.Relations.Set("p", p.CSort)
	x, _ := NewVariable("X", sortS)
	exists, err := NewExists([]*LogicVariable{x}, MustApply(p, x))
	if err != nil {
		t.Fatalf("NewExists: %v", err)
	}
	clauses := NewClauses([]Expr{exists}, nil, nil)
	state := NewInterpState(mod, nil, nil, "")

	diagOpen := Diagram(state, clauses, nil, nil, true, true)
	diagClosed := Diagram(state, clauses, nil, nil, true, false)
	if diagOpen == nil || diagClosed == nil {
		t.Fatalf("Go Diagram returned nil: open=%v closed=%v", diagOpen, diagClosed)
	}
	gotOpenFmlas := len(diagOpen.Fmlas)
	gotClosedFmlas := len(diagClosed.Fmlas)
	gotClosedHasEq := strings.Contains(diagClosed.String(), "=")
	gotStringsEqual := diagOpen.String() == diagClosed.String()

	if gotOpenFmlas != want.OpenFmlas ||
		gotClosedFmlas != want.ClosedFmlas ||
		gotClosedHasEq != want.ClosedHasEq ||
		gotStringsEqual != want.StringsEqual {
		t.Fatalf("Go Diagram differs from Python upward_close handling\nwant open=%d closed=%d eq=%v equal=%v\ngot  open=%d closed=%d eq=%v equal=%v\nopen: %s\nclosed: %s",
			want.OpenFmlas, want.ClosedFmlas, want.ClosedHasEq, want.StringsEqual,
			gotOpenFmlas, gotClosedFmlas, gotClosedHasEq, gotStringsEqual,
			diagOpen, diagClosed)
	}
}

func TestTopAlpha(t *testing.T) {
	s := NewInterpState(nil, BottomStateValue(), nil, "")
	if !s.IsBottom() {
		t.Fatal("should start as bottom")
	}
	TopAlpha(s)
	if s.IsBottom() {
		t.Error("after TopAlpha, should not be bottom")
	}
	if !s.Clauses.IsTrue() {
		t.Error("after TopAlpha, clauses should be true")
	}
}

func TestFailExpr(t *testing.T) {
	expr := interpTestAstCfg.NewAtom("myAction", interpTestAstCfg.NewAtom("s"))
	result := FailExpr(interpTestAstCfg, expr)
	if result.Rep != "fail_myAction" {
		t.Errorf("expected 'fail_myAction', got %q", result.Rep)
	}
}

// ---------------------------------------------------------------------------
// History helper tests
// ---------------------------------------------------------------------------

func TestNewHistoryFromState(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	h := NewHistoryFromState(nil, s)
	if h == nil {
		t.Fatal("NewHistoryFromState should not return nil")
	}
}

func TestHistorySatisfy(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	h := NewHistoryFromState(nil, s)
	// TrueClauses state with True history post is satisfiable,
	// so we should get a non-nil result with a non-empty path.
	result := HistorySatisfy(h, s)
	if result == nil {
		t.Error("HistorySatisfy of satisfiable history should return non-nil result")
	}
	if result != nil && len(result.Path) != 1 {
		t.Errorf("expected path length 1 (single state), got %d", len(result.Path))
	}
}

// ---------------------------------------------------------------------------
// Module helper tests
// ---------------------------------------------------------------------------

func TestModuleNewState(t *testing.T) {
	m := New()
	cls := TrueClauses(nil)
	s := ModuleNewState(m, cls)
	if s == nil {
		t.Fatal("ModuleNewState should not return nil")
	}
	if s.Domain != m {
		t.Error("domain should match")
	}
}

func TestModuleNewStateWithValue(t *testing.T) {
	m := New()
	sv := TopStateValue()
	s := ModuleNewStateWithValue(m, sv)
	if s == nil {
		t.Fatal("should not be nil")
	}
}

func TestModuleTypeCheck(t *testing.T) {
	m := New()
	if err := ModuleTypeCheck(m); err != nil {
		t.Errorf("stub should return nil: %v", err)
	}
}

func TestModuleTypeCheckConcepts(t *testing.T) {
	m := New()
	if err := ModuleTypeCheckConcepts(m); err != nil {
		t.Errorf("stub should return nil: %v", err)
	}
}

func TestFalseProperties(t *testing.T) {
	m := New()
	result := FalseProperties(m)
	if len(result) != 0 {
		t.Errorf("stub should return empty, got %d", len(result))
	}
}

func TestGetPropertyContext(t *testing.T) {
	m := New()
	result := GetPropertyContext(m, nil)
	if result == nil {
		t.Fatal("should not be nil")
	}
	if !result.IsTrue() {
		t.Error("stub should return TrueClauses")
	}
}

func TestBottomState(t *testing.T) {
	s := InterpBottomState(nil)
	if !s.IsBottom() {
		t.Error("InterpBottomState should be bottom")
	}
}

func TestNewStateFromClauses(t *testing.T) {
	m := New()
	cls := NewClauses(nil, nil, nil) // annot is nil
	s := NewStateFromClauses(m, cls)
	if s == nil {
		t.Fatal("should not be nil")
	}
	// Should have set annotation to EmptyAnnotation.
	if s.Clauses.Annot == nil {
		t.Error("annotation should not be nil after NewStateFromClauses")
	}
}

func TestNewStateFromClausesWithAnnot(t *testing.T) {
	m := New()
	existingAnnot := EmptyAnnotation{}
	cls := NewClauses(nil, nil, existingAnnot)
	s := NewStateFromClauses(m, cls)
	if s.Clauses.Annot != existingAnnot {
		t.Error("should preserve existing annotation")
	}
}

func TestEvalStateActions(t *testing.T) {
	pre := NewInterpState(nil, nil, nil, "")
	pre.Label = "startState"
	expr := interpTestAstCfg.NewAtom("act", interpTestAstCfg.NewAtom("startState"))
	result := EvalStateActions(expr, pre)
	if len(result) != 1 {
		t.Errorf("expected 1 action, got %d", len(result))
	}
}

func TestEvalStateActionsNoMatch(t *testing.T) {
	pre := NewInterpState(nil, nil, nil, "")
	pre.Label = "startState"
	expr := interpTestAstCfg.NewAtom("act", interpTestAstCfg.NewAtom("otherState"))
	result := EvalStateActions(expr, pre)
	if len(result) != 0 {
		t.Errorf("expected 0 actions, got %d", len(result))
	}
}

func TestEvalStateActionsJoin(t *testing.T) {
	pre := NewInterpState(nil, nil, nil, "")
	pre.Label = "s"
	expr := interpTestAstCfg.NewOr(
		interpTestAstCfg.NewAtom("a1", interpTestAstCfg.NewAtom("s")),
		interpTestAstCfg.NewAtom("a2", interpTestAstCfg.NewAtom("s")),
	)
	result := EvalStateActions(expr, pre)
	if len(result) != 2 {
		t.Errorf("expected 2 actions, got %d", len(result))
	}
}

func TestReverseNoPred(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	_, err := Reverse(s, nil)
	if err == nil {
		t.Error("Reverse without pred should error")
	}
}

func TestReverseUpdateConcreteClauses(t *testing.T) {
	s := NewInterpState(nil, nil, nil, "")
	_, err := ReverseUpdateConcreteClauses(s, nil)
	if err == nil {
		t.Error("should error without pred")
	}
}

// ---------------------------------------------------------------------------
// ApplyAction test
// ---------------------------------------------------------------------------

func TestApplyAction(t *testing.T) {
	m := New()
	s := NewInterpState(m, nil, nil, "")
	seq := NewSequence()
	result, err := ApplyAction(true, interpTestAstCfg.NewAtom("test"), "test", seq, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ActionName != "test" {
		t.Errorf("expected 'test', got %q", result.ActionName)
	}
	if result.Action != seq {
		t.Error("action should be set")
	}
}

// ---------------------------------------------------------------------------
// FailAction with EvalAction
// ---------------------------------------------------------------------------

func TestEvalActionFailAction(t *testing.T) {
	inner := NewSequence()
	fa := NewFailAction(inner)
	m := New()
	act, err := EvalAction(fa, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := act.(*FailAction); !ok {
		t.Error("should return a FailAction")
	}
}

// ---------------------------------------------------------------------------
// Fuzz tests
// ---------------------------------------------------------------------------

func FuzzStateValueCreation(f *testing.F) {
	f.Add("", true, true)
	f.Add("x", false, false)
	f.Add("a,b,c", true, false)
	f.Add("symbol_with_underscore", false, true)

	f.Fuzz(func(t *testing.T, modedStr string, useTrue bool, useFalsePrecond bool) {
		var moded []string
		if modedStr != "" {
			moded = strings.Split(modedStr, ",")
		}
		var cls *Clauses
		if useTrue {
			cls = TrueClauses(nil)
		} else {
			cls = FalseClauses(nil)
		}
		var pre *Clauses
		if useFalsePrecond {
			pre = FalseClauses(nil)
		} else {
			pre = TrueClauses(nil)
		}

		sv := NewStateValue(moded, cls, pre)
		if sv == nil {
			t.Fatal("NewStateValue returned nil")
		}
		s := NewInterpState(nil, sv, nil, "")
		if s == nil {
			t.Fatal("NewState returned nil")
		}
		// Value roundtrip.
		got := s.Value()
		if got == nil {
			t.Fatal("Value() returned nil")
		}
		if len(got.Moded) != len(sv.Moded) {
			t.Errorf("Moded length mismatch: %d vs %d", len(got.Moded), len(sv.Moded))
		}
	})
}

func FuzzExpressionHelpers(f *testing.F) {
	f.Add("action", "state", true)
	f.Add("", "s", false)
	f.Add("act_with_underscore", "s1", true)

	f.Fuzz(func(t *testing.T, actionName string, stateName string, useJoin bool) {
		// Test IsInterpActionApp / IsStateSymbol / InterpActionApp.
		stateAtom := interpTestAstCfg.NewAtom(stateName)
		if !IsStateSymbol(stateAtom) {
			t.Error("Atom with 0 args should be state symbol")
		}

		if actionName != "" {
			app := InterpActionApp(interpTestAstCfg, actionName, stateAtom)
			if !IsInterpActionApp(app) {
				t.Error("InterpActionApp result should be action app")
			}
			if app.Rep != actionName {
				t.Errorf("Rep mismatch: %q vs %q", app.Rep, actionName)
			}
		}

		if useJoin {
			join := InterpStateJoin(interpTestAstCfg, stateAtom)
			if !IsInterpStateJoin(join) {
				t.Error("StateJoin should produce state join")
			}
		}
	})
}

// Ensure we use all imports.
var (
	_ = True
	_ = NullUpdate
)
