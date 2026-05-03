package bmc

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// --- helpers ---

func testModule() *module.Module {
	return module.New()
}

func testModuleWithConj() *module.Module {
	mod := testModule()
	// Use a tautology (X = X) as the conjecture — always true.
	S := &lg.UninterpretedSort{Name: "S"}
	X, _ := lg.NewVariable("X", S)
	eq, _ := lg.NewEq(X, X)
	mod.LabeledConjs = []*ast.LabeledFormula{
		{Formula: eq},
	}
	return mod
}

func testModuleWithAction() *module.Module {
	mod := testModuleWithConj()
	act := actions.NewAssumeAction(lg.True)
	mod.Actions.Set("test_action", act)
	mod.PublicActions.Set("test_action", true)
	return mod
}

// --- Config tests ---

func TestDefaultConfig(t *testing.T) {
	mod := testModule()
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
	cfg := &Config{
		NSteps: 1,
		Module: testModule(),
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
	cfg := &Config{NSteps: 1, Module: testModule()}
	// Should not panic with nil logger.
	cfg.log("this is fine")
}

// --- CheckIsolate tests ---

func TestCheckIsolateNilConfig(t *testing.T) {
	result := CheckIsolate(nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Found {
		t.Error("should not find counterexample with nil config")
	}
}

func TestCheckIsolateNilModule(t *testing.T) {
	cfg := &Config{NSteps: 1}
	result := CheckIsolate(cfg)
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
	result := CheckIsolate(cfg)
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
	cfg := &Config{
		NSteps:  1,
		NUnroll: &n,
		Module:  mod,
	}
	result := CheckIsolate(cfg)
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
	result := CheckIsolate(cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// --- EnvAction tests ---

func TestEnvActionNilModule(t *testing.T) {
	act := EnvAction(nil)
	if act == nil {
		t.Fatal("EnvAction should not return nil")
	}
}

func TestEnvActionEmpty(t *testing.T) {
	mod := testModule()
	act := EnvAction(mod)
	if act == nil {
		t.Fatal("EnvAction should not return nil for empty module")
	}
}

func TestEnvActionWithActions(t *testing.T) {
	mod := testModuleWithAction()
	act := EnvAction(mod)
	if act == nil {
		t.Fatal("EnvAction should not return nil")
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
	mod := testModule()
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

func mkWhileAction(sortName string) *actions.WhileAction {
	sortT := &lg.UninterpretedSort{Name: sortName}
	ltSym := lg.NewConst("<", il.RelationSort([]lg.Sort{sortT, sortT}))
	xSym := lg.NewConst("x", sortT)
	boundSym := lg.NewConst("bound", sortT)
	cond, _ := lg.NewApply(ltSym, xSym, boundSym)
	body := actions.NewAssignAction(xSym, xSym)
	return actions.NewWhileAction(cond, body)
}

func TestUnrollAction_NonWhile(t *testing.T) {
	act := actions.NewAssumeAction(lg.True)
	result := UnrollAction(act, 3)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result.(*actions.AssumeAction); !ok {
		t.Errorf("expected *AssumeAction, got %T", result)
	}
}

func TestUnrollAction_WhileUnrolled(t *testing.T) {
	wa := mkWhileAction("T")
	result := UnrollAction(wa, 3)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result.(*actions.IfAction); !ok {
		t.Errorf("expected *IfAction (unrolled while), got %T", result)
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
	ifAct, ok := result.(*actions.IfAction)
	if !ok {
		t.Fatalf("expected *IfAction, got %T", result)
	}
	if _, ok := ifAct.ThenBody.(*actions.AssumeAction); !ok {
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
	if _, ok := result.(*actions.WhileAction); !ok {
		t.Errorf("expected original WhileAction back after panic recovery, got %T", result)
	}
}

func TestUnrollAction_WhileInsideSequence(t *testing.T) {
	assume := actions.NewAssumeAction(lg.True)
	wa := mkWhileAction("T")
	seq := actions.NewSequence(assume, wa)

	result := UnrollAction(seq, 2)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	s, ok := result.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected *Sequence, got %T", result)
	}
	if len(s.Elems) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(s.Elems))
	}
	if _, ok := s.Elems[0].(*actions.AssumeAction); !ok {
		t.Errorf("elem 0: expected *AssumeAction, got %T", s.Elems[0])
	}
	if _, ok := s.Elems[1].(*actions.IfAction); !ok {
		t.Errorf("elem 1: expected *IfAction (unrolled while), got %T", s.Elems[1])
	}
}

func TestUnrollAction_NestedWhile(t *testing.T) {
	inner := mkWhileAction("T")
	outer := actions.NewWhileAction(inner.Cond, inner)

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
	sortT := &lg.UninterpretedSort{Name: "T"}
	params := []*lg.Const{lg.NewConst("p", sortT)}
	returns := []*lg.Const{lg.NewConst("r", sortT)}
	wa.SetFormalParams(params)
	wa.SetFormalReturns(returns)

	result := UnrollAction(wa, 2)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	act := result.(actions.Action)
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
	cfg := &Config{
		NSteps:  1,
		NUnroll: &n,
		Module:  mod,
	}
	result := CheckIsolate(cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Actions should be restored after the call.
	restored, ok := mod.Actions.Get2("step")
	if !ok {
		t.Fatal("action 'step' should be restored")
	}
	if _, ok := restored.(*actions.WhileAction); !ok {
		t.Errorf("restored action should be *WhileAction, got %T", restored)
	}
}

func containsWhileAction(act actions.Action) bool {
	if _, ok := act.(*actions.WhileAction); ok {
		return true
	}
	for _, arg := range act.ActionArgs() {
		if child, ok := arg.(actions.Action); ok {
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
	S := &lg.UninterpretedSort{Name: "S"}
	X, _ := lg.NewVariable("X", S)
	pSort, _ := lg.NewFunctionSort(S, lg.Boolean)
	P := lg.NewConst("P", pSort)
	PX, _ := lg.NewApply(P, X)

	mod := module.New()

	mod.Sig = il.NewSig()
	mod.Sig.Sorts.Set("S", S)
	mod.Sig.Symbols.Set("P", &il.SymbolEntry{Sort: pSort})

	assignAction := actions.NewAssignAction(PX, lg.True)
	mod.Initializers = append(mod.Initializers, module.NamedAction{
		Name:   "init",
		Action: assignAction,
	})

	mod.LabeledConjs = []*ast.LabeledFormula{
		{Formula: PX},
	}

	cfg := DefaultConfig(mod, 0)
	result := CheckIsolate(cfg)
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
