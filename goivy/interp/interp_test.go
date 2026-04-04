package interp

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

var testAstCfg = ast.NewAstConfig()

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
	cls := module.NewClauses([]lg.Expr{lg.True}, nil, nil)
	pre := module.FalseClauses(nil)
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
	s := NewState(nil, nil, nil, "")
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
	m := module.New()
	s := NewState(m, nil, nil, "label1")
	if s.Domain != m {
		t.Error("expected same domain")
	}
	if s.Label != "label1" {
		t.Errorf("expected label 'label1', got %q", s.Label)
	}
}

func TestStateValueRoundtrip(t *testing.T) {
	sv := NewStateValue([]string{"a"}, module.TrueClauses(nil), module.FalseClauses(nil))
	s := NewState(nil, sv, nil, "")
	got := s.Value()
	if len(got.Moded) != 1 || got.Moded[0] != "a" {
		t.Errorf("Moded roundtrip failed: %v", got.Moded)
	}
}

func TestSetValue(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	sv := NewStateValue([]string{"x"}, module.FalseClauses(nil), module.TrueClauses(nil))
	s.SetValue(sv)
	if !s.Clauses.IsFalse() {
		t.Error("SetValue should set Clauses")
	}
	if len(s.Moded) != 1 {
		t.Error("SetValue should set Moded")
	}
}

func TestStateIsBottom(t *testing.T) {
	s := NewState(nil, BottomStateValue(), nil, "")
	if !s.IsBottom() {
		t.Error("state with FalseClauses should be bottom")
	}
	s2 := NewState(nil, TopStateValue(), nil, "")
	if s2.IsBottom() {
		t.Error("state with TrueClauses should not be bottom")
	}
}

func TestStateToFormula(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	f := s.ToFormula()
	if f == nil {
		t.Error("ToFormula should not return nil")
	}
}

func TestStateString(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	str := s.String()
	if str == "" {
		t.Error("String should not be empty")
	}
}

func TestStateStringNilClauses(t *testing.T) {
	s := &State{}
	str := s.String()
	if !strings.Contains(str, "nil") {
		t.Errorf("String for nil clauses should mention nil, got %q", str)
	}
}

func TestStatePredAndUpdate(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	if s.Pred() != nil {
		t.Error("initial Pred should be nil")
	}
	if s.Update() != nil {
		t.Error("initial Update should be nil")
	}
	pred := NewState(nil, nil, nil, "pred")
	s.SetPred(pred)
	if s.Pred() != pred {
		t.Error("SetPred should set predecessor")
	}
	u := actions.NullUpdate()
	s.SetUpdate(u)
	if s.Update() != u {
		t.Error("SetUpdate should set update")
	}
}

func TestStateConjs(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	if s.Conjs() != nil {
		t.Error("initial Conjs should be nil")
	}
	conjs := []*module.Clauses{module.TrueClauses(nil)}
	s.SetConjs(conjs)
	got := s.Conjs()
	if len(got) != 1 {
		t.Errorf("expected 1 conjecture, got %d", len(got))
	}
}

func TestStateUnders(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	if s.Unders() != nil {
		t.Error("initial Unders should be nil")
	}
	under := NewState(m, nil, nil, "under1")
	s.SetUnders([]*State{under})
	got := s.Unders()
	if len(got) != 1 {
		t.Errorf("expected 1 under, got %d", len(got))
	}
}

func TestStateAction(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	if s.Action != nil {
		t.Error("initial Action should be nil")
	}
	if s.ActionName != "" {
		t.Error("initial ActionName should be empty")
	}
}

func TestStateJoinOf(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	if s.JoinOf != nil {
		t.Error("initial JoinOf should be nil")
	}
}

// ---------------------------------------------------------------------------
// WrapState / UnwrapState tests
// ---------------------------------------------------------------------------

func TestWrapUnwrapState(t *testing.T) {
	s := NewState(nil, nil, nil, "test")
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
	atom := testAstCfg.NewAtom("foo")
	if UnwrapState(atom) != nil {
		t.Error("UnwrapState of non-state should return nil")
	}
}

func TestStateNodeString(t *testing.T) {
	s := NewState(nil, nil, nil, "")
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
	atom1 := testAstCfg.NewAtom("act", testAstCfg.NewAtom("state"))
	if !IsActionApp(atom1) {
		t.Error("Atom with 1 arg should be action app")
	}
	atom0 := testAstCfg.NewAtom("state")
	if IsActionApp(atom0) {
		t.Error("Atom with 0 args should not be action app")
	}
	or := testAstCfg.NewOr()
	if IsActionApp(or) {
		t.Error("Or should not be action app")
	}
}

func TestIsStateJoin(t *testing.T) {
	or := testAstCfg.NewOr()
	if !IsStateJoin(or) {
		t.Error("Or should be state join")
	}
	atom := testAstCfg.NewAtom("x")
	if IsStateJoin(atom) {
		t.Error("Atom should not be state join")
	}
}

func TestActionApp(t *testing.T) {
	result := ActionApp(testAstCfg, "doAction", testAstCfg.NewAtom("s"))
	if result.Rep != "doAction" {
		t.Errorf("expected rep 'doAction', got %q", result.Rep)
	}
	if len(result.Terms) != 1 {
		t.Errorf("expected 1 term, got %d", len(result.Terms))
	}
}

func TestStateJoinFunc(t *testing.T) {
	result := StateJoin(testAstCfg, testAstCfg.NewAtom("a"), testAstCfg.NewAtom("b"))
	if len(result.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(result.Terms))
	}
}

func TestIsStateSymbol(t *testing.T) {
	atom := testAstCfg.NewAtom("s")
	if !IsStateSymbol(atom) {
		t.Error("Atom with 0 args should be state symbol")
	}
	atom2 := testAstCfg.NewAtom("act", testAstCfg.NewAtom("s"))
	if IsStateSymbol(atom2) {
		t.Error("Atom with args should not be state symbol")
	}
}

func TestStateEquation(t *testing.T) {
	eq := StateEquation(testAstCfg, testAstCfg.NewAtom("lhs"), testAstCfg.NewAtom("rhs"))
	if eq == nil {
		t.Fatal("StateEquation should not return nil")
	}
	if eq.Defines() != "lhs" {
		t.Errorf("expected defines 'lhs', got %q", eq.Defines())
	}
}

func TestStatesInExpr(t *testing.T) {
	s1 := NewState(nil, nil, nil, "s1")
	s2 := NewState(nil, nil, nil, "s2")
	expr := testAstCfg.NewOr(WrapState(s1), WrapState(s2))
	states := StatesInExpr(expr)
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
}

func TestStatesInExprEmpty(t *testing.T) {
	expr := testAstCfg.NewAtom("foo")
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
	m := module.New()
	state := NewState(m, nil, nil, "")
	seq := actions.NewSequence()
	err := NewIvyActionFailedError(
		testAstCfg.NewAtom("test"),
		"myAction",
		seq,
		state,
		module.TrueClauses(nil),
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
	core := module.NewClauses([]lg.Expr{lg.True}, nil, nil)
	itp := module.NewClauses([]lg.Expr{lg.False}, nil, nil)
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
	m := module.New()
	s := NewState(m, nil, nil, "")
	upd := actions.NullUpdate()
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
	s := &State{}
	upd := actions.NullUpdate()
	_, err := ConcretePost(true, upd, s, nil)
	if err == nil {
		t.Error("expected error for nil domain")
	}
}

func TestConcreteJoinBasic(t *testing.T) {
	m := module.New()
	s1 := NewState(m, nil, nil, "")
	s2 := NewState(m, nil, nil, "")
	result, err := ConcreteJoin(s1, s2)
	if err != nil {
		t.Fatalf("ConcreteJoin returned error: %v", err)
	}
	if len(result.JoinOf) != 2 {
		t.Errorf("expected 2 JoinOf, got %d", len(result.JoinOf))
	}
}

func TestConcreteJoinNilDomain(t *testing.T) {
	s1 := &State{}
	s2 := NewState(nil, nil, nil, "")
	_, err := ConcreteJoin(s1, s2)
	if err == nil {
		t.Error("expected error for nil domain")
	}
}

// ---------------------------------------------------------------------------
// EvalAction tests
// ---------------------------------------------------------------------------

func TestEvalActionDirect(t *testing.T) {
	seq := actions.NewSequence()
	act, err := EvalAction(seq, nil)
	if err != nil {
		t.Fatalf("EvalAction returned error: %v", err)
	}
	if act != seq {
		t.Error("EvalAction should return the action directly")
	}
}

func TestEvalActionFromModule(t *testing.T) {
	m := module.New()
	seq := actions.NewSequence()
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
	m := module.New()
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
	m := module.New()
	s := NewState(m, nil, nil, "")
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
	m := module.New()
	trueNode := testAstCfg.NewAnd() // empty And = true
	result, err := EvalStateAtom(trueNode, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Clauses.IsTrue() {
		t.Error("true atom should produce true clauses")
	}
}

func TestEvalStateAtomFalse(t *testing.T) {
	m := module.New()
	falseNode := testAstCfg.NewOr() // empty Or = false
	result, err := EvalStateAtom(falseNode, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Clauses.IsFalse() {
		t.Error("false atom should produce false clauses")
	}
}

func TestEvalStateAtomSymbol(t *testing.T) {
	m := module.New()
	sym := testAstCfg.NewAtom("s")
	_, err := EvalStateAtom(sym, m)
	if err == nil {
		t.Error("state symbol lookup should fail (not implemented)")
	}
}

// ---------------------------------------------------------------------------
// FailAction tests
// ---------------------------------------------------------------------------

func TestNewFailAction(t *testing.T) {
	inner := actions.NewSequence()
	fa := NewFailAction(inner)
	if fa.Inner != inner {
		t.Error("Inner should be the provided action")
	}
	if fa.Name() != "fail" {
		t.Errorf("Name() should be 'fail', got %q", fa.Name())
	}
}

func TestFailActionString(t *testing.T) {
	inner := actions.NewSequence()
	fa := NewFailAction(inner)
	str := fa.String()
	if !strings.HasPrefix(str, "fail ") {
		t.Errorf("String() should start with 'fail ', got %q", str)
	}
}

func TestFailActionFailedAction(t *testing.T) {
	inner := actions.NewSequence()
	fa := NewFailAction(inner)
	if fa.FailedAction() != inner {
		t.Error("FailedAction should return inner")
	}
}

func TestFailActionClone(t *testing.T) {
	inner := actions.NewSequence()
	fa := NewFailAction(inner)
	cloned := fa.ActionClone(nil)
	if _, ok := cloned.(*FailAction); !ok {
		t.Error("Clone should return a *FailAction")
	}
}

func TestFailActionIterCalls(t *testing.T) {
	inner := actions.NewSequence()
	fa := NewFailAction(inner)
	calls := fa.IterCalls()
	// Sequence has no calls
	if len(calls) != 0 {
		t.Errorf("expected 0 calls, got %d", len(calls))
	}
}

func TestFailActionIterSubactions(t *testing.T) {
	inner := actions.NewSequence()
	fa := NewFailAction(inner)
	subs := fa.IterSubactions()
	if len(subs) != 1 {
		t.Errorf("expected 1 subaction, got %d", len(subs))
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestJoinUnders(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	result := JoinUnders(s)
	if result == nil {
		t.Fatal("JoinUnders should not return nil")
	}
}

func TestAddUnder(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	cls := module.TrueClauses(nil)
	under := AddUnder(s, cls, nil, nil)
	if under == nil {
		t.Fatal("AddUnder should return a state")
	}
	if len(s.Unders()) != 1 {
		t.Errorf("expected 1 under, got %d", len(s.Unders()))
	}
}

func TestAddUnderWithPred(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	pred := NewState(m, nil, nil, "pred")
	under := AddUnder(s, module.TrueClauses(nil), pred, "universe")
	if under.Pred() != pred {
		t.Error("under pred should be set")
	}
	if under.Universe != "universe" {
		t.Error("under universe should be set")
	}
}

func TestReachStateNoPred(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	result := ReachState(s, nil)
	if result != nil {
		t.Error("ReachState without pred should return nil")
	}
}

func TestReachStateFromPredNoPred(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	result, err := ReachStateFromPred(s, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("ReachStateFromPred without pred should return nil")
	}
}

func TestUndecidedConjectures(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	// TrueClauses state implies TrueClauses conjecture, so it should
	// be decided (not undecided). Use FalseClauses as conjecture to
	// get an undecided one (True does not imply False).
	conjs := []*module.Clauses{module.FalseClauses(nil)}
	s.SetConjs(conjs)
	result := UndecidedConjectures(s)
	if len(result) != 1 {
		t.Errorf("expected 1 undecided conjecture, got %d", len(result))
	}
}

func TestFilterConjectures(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	conjs := []*module.Clauses{module.TrueClauses(nil)}
	s.SetConjs(conjs)
	lost := FilterConjectures(s, module.TrueClauses(nil))
	if len(lost) != 0 {
		t.Errorf("stub should lose no conjectures, got %d", len(lost))
	}
}

func TestCaseConjecture(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	_, _, ok := CaseConjecture(s, module.TrueClauses(nil))
	if ok {
		t.Error("stub CaseConjecture should return false")
	}
}

func TestDiagram(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	// TrueClauses is satisfiable, so Diagram should return non-nil.
	result := Diagram(s, module.TrueClauses(nil), nil, nil, true, true)
	if result == nil {
		t.Error("Diagram of satisfiable clauses should return non-nil")
	}
	// FalseClauses is unsatisfiable, so Diagram should return nil.
	result2 := Diagram(s, module.FalseClauses(nil), nil, nil, true, true)
	if result2 != nil {
		t.Error("Diagram of unsatisfiable clauses should return nil")
	}
}

func TestTopAlpha(t *testing.T) {
	s := NewState(nil, BottomStateValue(), nil, "")
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
	expr := testAstCfg.NewAtom("myAction", testAstCfg.NewAtom("s"))
	result := FailExpr(testAstCfg, expr)
	if result.Rep != "fail_myAction" {
		t.Errorf("expected 'fail_myAction', got %q", result.Rep)
	}
}

// ---------------------------------------------------------------------------
// History helper tests
// ---------------------------------------------------------------------------

func TestNewHistoryFromState(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	h := NewHistoryFromState(nil, s)
	if h == nil {
		t.Fatal("NewHistoryFromState should not return nil")
	}
}

func TestHistorySatisfy(t *testing.T) {
	s := NewState(nil, nil, nil, "")
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
	m := module.New()
	cls := module.TrueClauses(nil)
	s := ModuleNewState(m, cls)
	if s == nil {
		t.Fatal("ModuleNewState should not return nil")
	}
	if s.Domain != m {
		t.Error("domain should match")
	}
}

func TestModuleNewStateWithValue(t *testing.T) {
	m := module.New()
	sv := TopStateValue()
	s := ModuleNewStateWithValue(m, sv)
	if s == nil {
		t.Fatal("should not be nil")
	}
}

func TestModuleTypeCheck(t *testing.T) {
	m := module.New()
	if err := ModuleTypeCheck(m); err != nil {
		t.Errorf("stub should return nil: %v", err)
	}
}

func TestModuleTypeCheckConcepts(t *testing.T) {
	m := module.New()
	if err := ModuleTypeCheckConcepts(m); err != nil {
		t.Errorf("stub should return nil: %v", err)
	}
}

func TestFalseProperties(t *testing.T) {
	m := module.New()
	result := FalseProperties(m)
	if len(result) != 0 {
		t.Errorf("stub should return empty, got %d", len(result))
	}
}

func TestGetPropertyContext(t *testing.T) {
	m := module.New()
	result := GetPropertyContext(m, nil)
	if result == nil {
		t.Fatal("should not be nil")
	}
	if !result.IsTrue() {
		t.Error("stub should return TrueClauses")
	}
}

func TestBottomState(t *testing.T) {
	s := BottomState(nil)
	if !s.IsBottom() {
		t.Error("BottomState should be bottom")
	}
}

func TestNewStateFromClauses(t *testing.T) {
	m := module.New()
	cls := module.NewClauses(nil, nil, nil) // annot is nil
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
	m := module.New()
	cls := module.NewClauses(nil, nil, "existing_annot")
	s := NewStateFromClauses(m, cls)
	if s.Clauses.Annot != "existing_annot" {
		t.Error("should preserve existing annotation")
	}
}

func TestEvalStateActions(t *testing.T) {
	pre := NewState(nil, nil, nil, "")
	pre.Label = "startState"
	expr := testAstCfg.NewAtom("act", testAstCfg.NewAtom("startState"))
	result := EvalStateActions(expr, pre)
	if len(result) != 1 {
		t.Errorf("expected 1 action, got %d", len(result))
	}
}

func TestEvalStateActionsNoMatch(t *testing.T) {
	pre := NewState(nil, nil, nil, "")
	pre.Label = "startState"
	expr := testAstCfg.NewAtom("act", testAstCfg.NewAtom("otherState"))
	result := EvalStateActions(expr, pre)
	if len(result) != 0 {
		t.Errorf("expected 0 actions, got %d", len(result))
	}
}

func TestEvalStateActionsJoin(t *testing.T) {
	pre := NewState(nil, nil, nil, "")
	pre.Label = "s"
	expr := testAstCfg.NewOr(
		testAstCfg.NewAtom("a1", testAstCfg.NewAtom("s")),
		testAstCfg.NewAtom("a2", testAstCfg.NewAtom("s")),
	)
	result := EvalStateActions(expr, pre)
	if len(result) != 2 {
		t.Errorf("expected 2 actions, got %d", len(result))
	}
}

func TestReverseNoPred(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	_, err := Reverse(s, nil)
	if err == nil {
		t.Error("Reverse without pred should error")
	}
}

func TestReverseUpdateConcreteClauses(t *testing.T) {
	s := NewState(nil, nil, nil, "")
	_, err := ReverseUpdateConcreteClauses(s, nil)
	if err == nil {
		t.Error("should error without pred")
	}
}

// ---------------------------------------------------------------------------
// ApplyAction test
// ---------------------------------------------------------------------------

func TestApplyAction(t *testing.T) {
	m := module.New()
	s := NewState(m, nil, nil, "")
	seq := actions.NewSequence()
	result, err := ApplyAction(true, testAstCfg.NewAtom("test"), "test", seq, s)
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
	inner := actions.NewSequence()
	fa := NewFailAction(inner)
	m := module.New()
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
		var cls *module.Clauses
		if useTrue {
			cls = module.TrueClauses(nil)
		} else {
			cls = module.FalseClauses(nil)
		}
		var pre *module.Clauses
		if useFalsePrecond {
			pre = module.FalseClauses(nil)
		} else {
			pre = module.TrueClauses(nil)
		}

		sv := NewStateValue(moded, cls, pre)
		if sv == nil {
			t.Fatal("NewStateValue returned nil")
		}
		s := NewState(nil, sv, nil, "")
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
		// Test IsActionApp / IsStateSymbol / ActionApp.
		stateAtom := testAstCfg.NewAtom(stateName)
		if !IsStateSymbol(stateAtom) {
			t.Error("Atom with 0 args should be state symbol")
		}

		if actionName != "" {
			app := ActionApp(testAstCfg, actionName, stateAtom)
			if !IsActionApp(app) {
				t.Error("ActionApp result should be action app")
			}
			if app.Rep != actionName {
				t.Errorf("Rep mismatch: %q vs %q", app.Rep, actionName)
			}
		}

		if useJoin {
			join := StateJoin(testAstCfg, stateAtom)
			if !IsStateJoin(join) {
				t.Error("StateJoin should produce state join")
			}
		}
	})
}

// Ensure we use all imports.
var (
	_ = lg.True
	_ = actions.NullUpdate
)
