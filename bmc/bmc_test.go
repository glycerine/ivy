package bmc

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// --- helpers ---

func testModule() *module.Module {
	return module.New()
}

func testModuleWithConj() *module.Module {
	mod := testModule()
	// Use a tautology (X = X) as the conjecture — always true.
	S := &lg.UninterpretedSort{Name: "S"}
	X, _ := lg.NewVar("X", S)
	eq, _ := lg.NewEq(X, X)
	mod.LabeledConjs = []*module.LabeledFormula{
		{Formula: eq},
	}
	return mod
}

func testModuleWithAction() *module.Module {
	mod := testModuleWithConj()
	act := actions.NewAssumeAction(lg.True)
	mod.Actions["test_action"] = act
	mod.PublicActions["test_action"] = true
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
	if _, ok := mod.Actions["test_action"]; !ok {
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

// --- DualClauses tests ---

func TestDualClausesNil(t *testing.T) {
	dual := DualClauses(nil)
	if dual == nil {
		t.Fatal("DualClauses should not return nil")
	}
}

func TestDualClausesEmpty(t *testing.T) {
	clauses := clauseops.NewClauses(nil, nil, nil)
	dual := DualClauses(clauses)
	if dual == nil {
		t.Fatal("DualClauses should not return nil")
	}
}

func TestDualClausesSingle(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	clauses := clauseops.NewClauses([]lg.Node{p}, nil, nil)
	dual := DualClauses(clauses)
	if dual == nil {
		t.Fatal("DualClauses should not return nil")
	}
	if len(dual.Fmlas) != 1 {
		t.Errorf("expected 1 formula in dual, got %d", len(dual.Fmlas))
	}
	or, ok := dual.Fmlas[0].(*lg.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", dual.Fmlas[0])
	}
	if len(or.Terms) != 1 {
		t.Errorf("expected 1 term in Or, got %d", len(or.Terms))
	}
}

func TestDualClausesMultiple(t *testing.T) {
	c := lg.NewConst("P", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)
	clauses := clauseops.NewClauses([]lg.Node{q, c}, nil, nil)
	dual := DualClauses(clauses)
	if len(dual.Fmlas) != 1 {
		t.Errorf("expected 1 formula in dual, got %d", len(dual.Fmlas))
	}
	or, ok := dual.Fmlas[0].(*lg.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", dual.Fmlas[0])
	}
	if len(or.Terms) != 2 {
		t.Errorf("expected 2 terms in Or, got %d", len(or.Terms))
	}
}

// --- UnrollAction tests ---

func TestUnrollAction(t *testing.T) {
	act := actions.NewAssumeAction(lg.True)
	result := UnrollAction(act, 3)
	if result != act {
		t.Error("stub UnrollAction should return same action")
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
