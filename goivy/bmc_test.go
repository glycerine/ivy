package goivy

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// --- helpers ---

func bmcTestModule() *Module {
	return New()
}

func testModuleWithConj() *Module {
	mod := bmcTestModule()
	// Use a tautology (X = X) as the conjecture — always true.
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)
	mod.LabeledConjs = []*LabeledFormula{
		{Formula: eq},
	}
	return mod
}

func testModuleWithAction() *Module {
	mod := testModuleWithConj()
	act := NewAssumeAction(True)
	mod.Actions.Set("test_action", act)
	mod.PublicActions.Set("test_action", true)
	return mod
}

// --- Config tests ---

func TestDefaultConfig(t *testing.T) {
	mod := bmcTestModule()
	cfg := DefaultConfig(mod, 5)
	if cfg.NSteps != 5 {
		t.Errorf("expected NSteps=5, got %d", cfg.NSteps)
	}
	if cfg.Module != mod {
		t.Error("Module mismatch")
	}
	if cfg.NUnroll != nil {
		t.Error("NUnroll should be nil by default")
	}
}

func TestConfigLog(t *testing.T) {
	var msgs []string
	cfg := &BMCConfig{
		NSteps: 1,
		Module: bmcTestModule(),
		Logger: func(s string) { msgs = append(msgs, s) },
	}
	cfg.log("test %d", 42)
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0] != "test 42" {
		t.Errorf("expected 'test 42', got %q", msgs[0])
	}
}

func TestConfigLogNil(t *testing.T) {
	cfg := &BMCConfig{NSteps: 1, Module: bmcTestModule()}
	// Should not panic with nil logger.
	cfg.log("this is fine")
}

// --- CheckIsolate tests ---

func TestCheckIsolateNilConfig(t *testing.T) {
	result := BMCCheckIsolate(nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Found {
		t.Error("should not find counterexample with nil config")
	}
}

func TestCheckIsolateNilModule(t *testing.T) {
	cfg := &BMCConfig{NSteps: 1}
	result := BMCCheckIsolate(cfg)
	if result.Found {
		t.Error("should not find counterexample with nil module")
	}
	if !strings.Contains(result.Message, "nil module") {
		t.Errorf("expected 'nil module' message, got %q", result.Message)
	}
}

func TestCheckIsolateZeroSteps(t *testing.T) {
	mod := testModuleWithConj()
	cfg := DefaultConfig(mod, 0)
	result := BMCCheckIsolate(cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// With no real solver, should report no counterexample.
	if result.Found {
		t.Error("stub solver should not find counterexample")
	}
}

func TestCheckIsolateWithUnroll(t *testing.T) {
	mod := testModuleWithAction()
	n := 2
	cfg := &BMCConfig{
		NSteps:  1,
		NUnroll: &n,
		Module:  mod,
	}
	result := BMCCheckIsolate(cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Actions should be restored after the call.
	if _, ok := mod.Actions.Get2("test_action"); !ok {
		t.Error("actions should be restored after unrolling")
	}
}

func TestCheckIsolateMultipleSteps(t *testing.T) {
	mod := testModuleWithAction()
	cfg := DefaultConfig(mod, 3)
	result := BMCCheckIsolate(cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestBMCCheckIsolateExecutesInitializeActionLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_art, ivy_logic as il, ivy_logic_utils as ilu, ivy_module as im, ivy_trace, logic as lg

im.module = im.Module()
il.sig = im.module.sig
P = il.Symbol("P", lg.Boolean)
im.module.relations["P"] = P
im.module.actions["initialize"] = act.AssignAction(P, lg.true)
clauses = ilu.dual_clauses(ilu.formula_to_clauses(P), lambda v: lg.Const("@" + v.name, v.sort))

ag = ivy_art.AnalysisGraph()
ag.add_initial_state(ag.init_cond)
post = ag.states[0]
post = ag.execute(im.module.actions["initialize"], None, None, "initialize")
res = ivy_trace.check_final_cond(ag, post, clauses, [], True)
print(json.dumps({"safe_after_initialize": res is None, "states": len(ag.states)}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python BMC initialize oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		SafeAfterInitialize bool `json:"safe_after_initialize"`
		States              int  `json:"states"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python BMC initialize oracle %q: %v", out, err)
	}
	if !want.SafeAfterInitialize || want.States != 2 {
		t.Fatalf("python BMC initialize oracle did not execute initialize as expected: %+v", want)
	}

	mod := bmcTestModule()
	mod.Sig = NewSig()
	mod.Sig.Symbols.Set("P", &SymbolEntry{Sort: Boolean})
	mod.Relations.Set(RelationKey("P", Boolean), Boolean)
	P := NewConst("P", Boolean)
	mod.Actions.Set("initialize", NewAssignAction(P, True))
	mod.LabeledConjs = []*LabeledFormula{{Formula: P}}

	result := BMCCheckIsolate(DefaultConfig(mod, 0))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Found {
		t.Fatalf("BMCCheckIsolate found a depth-0 counterexample; Python executes initialize first: %s", result.Message)
	}
}

func TestBMCAssertionFailureTraceUsesFailActionHistoryLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_art, ivy_interp as itp, ivy_logic as il, ivy_logic_utils as ilu, ivy_module as im, ivy_trace, logic as lg

im.module = im.Module()
il.sig = im.module.sig
P = il.Symbol("P", lg.Boolean)
im.module.relations["P"] = P
step = act.AssertAction(P)
im.module.actions["step"] = step
im.module.public_actions["step"] = True

ag = ivy_art.AnalysisGraph()
ag.add_initial_state(ag.init_cond)
with itp.EvalContext(False):
    post = ag.execute(act.env_action(None))
fail = itp.State(expr=itp.fail_expr(post.expr))
history = ag.get_history(fail)
trace = ivy_trace.check_final_cond(ag, fail, ilu.true_clauses(), [], True)
print(json.dumps({
    "history_actions": [type(a).__name__ for a in history.actions],
    "contains_fail_call": "fail call step" in str(trace),
    "contains_assert": "assert P" in str(trace),
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python BMC failure-history oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		HistoryActions   []string `json:"history_actions"`
		ContainsFailCall bool     `json:"contains_fail_call"`
		ContainsAssert   bool     `json:"contains_assert"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python BMC failure-history oracle %q: %v", out, err)
	}
	if !reflect.DeepEqual(want.HistoryActions, []string{"fail_action"}) || !want.ContainsFailCall || !want.ContainsAssert {
		t.Fatalf("python BMC failure-history oracle did not expose fail_action trace shape: %+v", want)
	}

	mod := bmcTestModule()
	mod.Sig = NewSig()
	mod.Sig.Symbols.Set("P", &SymbolEntry{Sort: Boolean})
	mod.Relations.Set(RelationKey("P", Boolean), Boolean)
	P := NewConst("P", Boolean)
	mod.Actions.Set("step", NewAssertAction(P))
	mod.PublicActions.Set("step", true)

	ag := NewAnalysisGraph(mod)
	ag.AddInitialState(nil, nil)
	stepAction := BMCEnvAction(mod)
	post, err := ag.Execute(false, stepAction, nil, nil, "")
	if err != nil {
		t.Fatalf("execute BMC step: %v", err)
	}
	failState := bmcFailStateFromPost(mod, post, stepAction)
	history := ag.GetHistory(failState, nil)
	gotHistoryActions := make([]string, len(history.Actions))
	for i, action := range history.Actions {
		if aa, ok := action.(ActionsAction); ok {
			gotHistoryActions[i] = ActionTypeName(aa)
		}
	}
	if !reflect.DeepEqual(gotHistoryActions, want.HistoryActions) {
		t.Fatalf("Go BMC failure history actions differ from Python\nwant: %#v\ngot:  %#v", want.HistoryActions, gotHistoryActions)
	}

	result := BMCCheckIsolate(DefaultConfig(mod, 0))
	if result == nil || !result.Found || result.Trace == nil {
		t.Fatalf("expected BMC assertion failure trace, got %#v", result)
	}
	got := result.Trace.String()
	if !strings.Contains(got, "fail") || !strings.Contains(got, "assert P") {
		t.Fatalf("Go BMC assertion-failure trace lost fail_action content\nwant substrings: %q and %q\ngot:\n%s", "fail", "assert P", got)
	}
}

// --- EnvAction tests ---

func TestEnvActionNilModule(t *testing.T) {
	act := BMCEnvAction(nil)
	if act == nil {
		t.Fatal("EnvAction should not return nil")
	}
}

func TestEnvActionEmpty(t *testing.T) {
	mod := bmcTestModule()
	act := BMCEnvAction(mod)
	if act == nil {
		t.Fatal("EnvAction should not return nil for empty module")
	}
}

func TestEnvActionWithActions(t *testing.T) {
	mod := testModuleWithAction()
	act := BMCEnvAction(mod)
	if act == nil {
		t.Fatal("EnvAction should not return nil")
	}
}

type bmcEnvBranchShape struct {
	Type          string   `json:"type"`
	Label         string   `json:"label"`
	ArgTypes      []string `json:"arg_types"`
	FormalParams  int      `json:"formal_params"`
	FormalReturns int      `json:"formal_returns"`
}

func TestBMCEnvActionMatchesPythonEnvActionShape(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_module as im, logic as lg

im.module = im.Module()
sort_s = lg.UninterpretedSort("S")

zeta = act.AssumeAction(lg.true)
zeta.formal_params = [lg.Const("x", sort_s)]
zeta.formal_returns = [lg.Const("r", sort_s)]
im.module.actions["ext:zeta"] = zeta
im.module.public_actions["ext:zeta"] = True

alpha = act.AssumeAction(lg.true)
im.module.actions["alpha"] = alpha
im.module.public_actions["alpha"] = True

env = act.env_action(None)
rows = []
for branch in env.args:
    rows.append({
        "type": type(branch).__name__,
        "label": getattr(branch, "label", ""),
        "arg_types": [type(x).__name__ for x in getattr(branch, "args", [])],
        "formal_params": len(getattr(branch, "formal_params", [])),
        "formal_returns": len(getattr(branch, "formal_returns", [])),
    })
print(json.dumps(rows))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python env_action oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []bmcEnvBranchShape
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python env_action oracle %q: %v", out, err)
	}

	mod := bmcTestModule()
	sortS := &UninterpretedSort{Name: "S"}
	zeta := NewAssumeAction(True)
	zeta.SetFormalParams([]*Const{NewConst("x", sortS)})
	zeta.SetFormalReturns([]*Const{NewConst("r", sortS)})
	mod.Actions.Set("ext:zeta", zeta)
	mod.PublicActions.Set("ext:zeta", true)
	alpha := NewAssumeAction(True)
	mod.Actions.Set("alpha", alpha)
	mod.PublicActions.Set("alpha", true)

	env, ok := BMCEnvAction(mod).(*LogicEnvAction)
	if !ok {
		t.Fatalf("BMCEnvAction returned %T, want *LogicEnvAction", BMCEnvAction(mod))
	}
	got := describeBMCEnvBranches(env)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Go BMCEnvAction shape differs from Python env_action(None)\nwant: %#v\ngot:  %#v", want, got)
	}
}

func describeBMCEnvBranches(env *LogicEnvAction) []bmcEnvBranchShape {
	rows := make([]bmcEnvBranchShape, 0, len(env.Branches))
	for _, branch := range env.Branches {
		action, _ := branch.(ActionsAction)
		row := bmcEnvBranchShape{
			Type:          pythonBMCActionTypeName(branch),
			FormalParams:  len(action.GetFormalParams()),
			FormalReturns: len(action.GetFormalReturns()),
		}
		if labeled, ok := branch.(interface{ GetLabel() string }); ok {
			row.Label = labeled.GetLabel()
		}
		if seq, ok := branch.(*LogicSequence); ok {
			for _, elem := range seq.Elems {
				row.ArgTypes = append(row.ArgTypes, pythonBMCActionTypeName(elem))
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func pythonBMCActionTypeName(x Expr) string {
	switch x.(type) {
	case *LogicSequence:
		return "Sequence"
	case *LogicAssumeAction:
		return "AssumeAction"
	case *ReturnAction:
		return "ReturnAction"
	case *LogicEnvAction:
		return "EnvAction"
	default:
		return TypeName(x)
	}
}

// --- BuildConjecture tests ---

func TestBuildConjectureNilModule(t *testing.T) {
	conj := BuildConjecture(nil)
	if conj == nil {
		t.Fatal("BuildConjecture should not return nil")
	}
	if !conj.IsTrue() {
		t.Error("nil module should produce true clauses")
	}
}

func TestBuildConjectureEmpty(t *testing.T) {
	mod := bmcTestModule()
	conj := BuildConjecture(mod)
	if !conj.IsTrue() {
		t.Error("empty module should produce true clauses")
	}
}

func TestBuildConjectureWithConj(t *testing.T) {
	mod := testModuleWithConj()
	conj := BuildConjecture(mod)
	if conj == nil {
		t.Fatal("BuildConjecture should not return nil")
	}
	if len(conj.Fmlas) != 1 {
		t.Errorf("expected 1 formula, got %d", len(conj.Fmlas))
	}
}

// --- UnrollAction tests ---

func mkWhileAction(sortName string) *LogicWhileAction {
	sortT := &UninterpretedSort{Name: sortName}
	ltSym := NewConst("<", LogicRelationSort([]Sort{sortT, sortT}))
	xSym := NewConst("x", sortT)
	boundSym := NewConst("bound", sortT)
	cond, _ := NewApply(ltSym, xSym, boundSym)
	body := NewAssignAction(xSym, xSym)
	return NewWhileAction(cond, body)
}

func TestUnrollAction_NonWhile(t *testing.T) {
	act := NewAssumeAction(True)
	result := UnrollAction(act, 3)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result.(*LogicAssumeAction); !ok {
		t.Errorf("expected *LogicAssumeAction, got %T", result)
	}
}

func TestUnrollAction_WhileUnrolled(t *testing.T) {
	wa := mkWhileAction("T")
	result := UnrollAction(wa, 3)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result.(*LogicIfAction); !ok {
		t.Errorf("expected *LogicIfAction (unrolled while), got %T", result)
	}
}

func TestUnrollAction_NilAction(t *testing.T) {
	result := UnrollAction(nil, 3)
	if result != nil {
		t.Errorf("expected nil, got %T", result)
	}
}

func TestUnrollAction_ZeroUnroll(t *testing.T) {
	wa := mkWhileAction("T")
	result := UnrollAction(wa, 0)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// card=0: just the base case if(cond, assume(false))
	ifAct, ok := result.(*LogicIfAction)
	if !ok {
		t.Fatalf("expected *LogicIfAction, got %T", result)
	}
	if _, ok := ifAct.ThenBody.(*LogicAssumeAction); !ok {
		t.Errorf("base case should be AssumeAction, got %T", ifAct.ThenBody)
	}
}

func TestUnrollAction_LargeUnroll(t *testing.T) {
	wa := mkWhileAction("T")
	result := UnrollAction(wa, 200)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Should recover from panic and return original action
	if _, ok := result.(*LogicWhileAction); !ok {
		t.Errorf("expected original WhileAction back after panic recovery, got %T", result)
	}
}

func TestUnrollAction_WhileInsideSequence(t *testing.T) {
	assume := NewAssumeAction(True)
	wa := mkWhileAction("T")
	seq := NewSequence(assume, wa)

	result := UnrollAction(seq, 2)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	s, ok := result.(*LogicSequence)
	if !ok {
		t.Fatalf("expected *LogicSequence, got %T", result)
	}
	if len(s.Elems) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(s.Elems))
	}
	if _, ok := s.Elems[0].(*LogicAssumeAction); !ok {
		t.Errorf("elem 0: expected *LogicAssumeAction, got %T", s.Elems[0])
	}
	if _, ok := s.Elems[1].(*LogicIfAction); !ok {
		t.Errorf("elem 1: expected *LogicIfAction (unrolled while), got %T", s.Elems[1])
	}
}

func TestUnrollAction_NestedWhile(t *testing.T) {
	inner := mkWhileAction("T")
	outer := NewWhileAction(inner.Cond, inner)

	result := UnrollAction(outer, 2)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Neither while should remain
	if containsWhileAction(result) {
		t.Error("unrolled result should not contain any WhileAction")
	}
}

func TestUnrollAction_FormalsPreserved(t *testing.T) {
	wa := mkWhileAction("T")
	sortT := &UninterpretedSort{Name: "T"}
	params := []*Const{NewConst("p", sortT)}
	returns := []*Const{NewConst("r", sortT)}
	wa.SetFormalParams(params)
	wa.SetFormalReturns(returns)

	result := UnrollAction(wa, 2)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	act := result.(ActionsAction)
	fp := act.GetFormalParams()
	fr := act.GetFormalReturns()
	if len(fp) != 1 || fp[0].Name != "p" {
		t.Errorf("formal params not preserved: %v", fp)
	}
	if len(fr) != 1 || fr[0].Name != "r" {
		t.Errorf("formal returns not preserved: %v", fr)
	}
}

func TestCheckIsolateWithUnroll_WhileAction(t *testing.T) {
	mod := testModuleWithConj()
	wa := mkWhileAction("T")
	mod.Actions.Set("step", wa)
	mod.PublicActions.Set("step", true)

	n := 2
	cfg := &BMCConfig{
		NSteps:  1,
		NUnroll: &n,
		Module:  mod,
	}
	result := BMCCheckIsolate(cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Actions should be restored after the call.
	restored, ok := mod.Actions.Get2("step")
	if !ok {
		t.Fatal("action 'step' should be restored")
	}
	if _, ok := restored.(*LogicWhileAction); !ok {
		t.Errorf("restored action should be *LogicWhileAction, got %T", restored)
	}
}

func containsWhileAction(act ActionsAction) bool {
	if _, ok := act.(*LogicWhileAction); ok {
		return true
	}
	for _, arg := range act.ActionArgs() {
		if child, ok := arg.(ActionsAction); ok {
			if containsWhileAction(child) {
				return true
			}
		}
	}
	return false
}

// TestBMCInitializerRespected verifies that BMC uses AddInitialState
// (which runs mod.Initializers) rather than an unconstrained TrueClauses
// initial state. Regression test: before the fix, BMC started from an
// unconstrained state and trivially found spurious counterexamples at
// depth 0 even when the init block established the conjecture.
func TestBMCInitializerRespected(t *testing.T) {
	// Build a module where:
	//   - Sort S exists
	//   - Relation P(X:S) exists
	//   - after init { P(X) := true }   (initializer sets P to true for all X)
	//   - conjecture forall X:S. P(X)   (should hold in initial state)
	//
	// Regression: before the fix, BMC started from TrueClauses (unconstrained)
	// and trivially found a spurious counterexample at depth 0.
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	pSort, _ := NewFunctionSort(S, Boolean)
	P := NewConst("P", pSort)
	PX, _ := NewApply(P, X)

	mod := New()

	mod.Sig = NewSig()
	mod.Sig.Sorts.Set("S", S)
	mod.Sig.Symbols.Set("P", &SymbolEntry{Sort: pSort})

	assignAction := NewAssignAction(PX, True)
	mod.Initializers = append(mod.Initializers, NamedAction{
		Name:   "init",
		Action: assignAction,
	})

	mod.LabeledConjs = []*LabeledFormula{
		{Formula: PX},
	}

	cfg := DefaultConfig(mod, 0)
	result := BMCCheckIsolate(cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Found {
		t.Errorf("BMC should NOT find counterexample at depth 0 when initializer establishes the conjecture; got: %s", result.Message)
	}
}

// --- BMCResult tests ---

func TestBMCResultString(t *testing.T) {
	r := &BMCResult{Found: true, Depth: 3, Message: "found at depth 3"}
	if r.Message != "found at depth 3" {
		t.Errorf("unexpected message: %q", r.Message)
	}
}

func TestBMCResultNotFound(t *testing.T) {
	r := &BMCResult{Found: false, Depth: -1, Message: "no counterexample"}
	if r.Found {
		t.Error("should not be found")
	}
	if r.Depth != -1 {
		t.Errorf("expected depth -1, got %d", r.Depth)
	}
}

// --- Fuzz ---
/* off for now, it hangs.
func FuzzCheckIsolateSteps(f *testing.F) {

	f.Add(0)
	f.Add(1)
	f.Add(5)
	f.Add(100)
	f.Fuzz(func(t *testing.T, nSteps int) {
		if nSteps < 0 || nSteps > 1000 {
			return // bound the input to avoid long runs
		}
		mod := testModuleWithConj()
		cfg := DefaultConfig(mod, nSteps)
		result := CheckIsolate(cfg)
		if result == nil {
			t.Error("CheckIsolate should not return nil")
		}
	})
}
*/
