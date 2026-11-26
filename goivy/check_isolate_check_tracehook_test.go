package goivy

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSubgoalsMethodFailureAppliesTraceHook(t *testing.T) {
	cfg := NewAstConfig()
	hookCalled := false
	goal := cfg.NewLabeledFormula(cfg.NewAtom("goal"), True)
	goal.TraceHook = TraceHookFn(func(trace *Trace, fcs []Checker) *Trace {
		hookCalled = true
		return trace
	})

	mod := New()
	mod.Cfg = NewConfig()
	err := CheckSubgoals([]*LabeledFormula{goal}, func(_ *Module) error {
		return fmt.Errorf("method counterexample")
	}, mod)
	if err == nil {
		t.Fatal("CheckSubgoals method unexpectedly succeeded")
	}
	if !hookCalled {
		t.Fatal("method subgoal failure did not apply goal.TraceHook before returning the failure")
	}
}

func TestCheckSubgoalsArchivesDecodedLivenessLoop(t *testing.T) {
	mod := New()
	mod.Cfg.TraceDir = filepath.Join(t.TempDir(), "traces")
	files, err := openTraceFiles(mod.Cfg.TraceDir)
	if err != nil {
		t.Fatal(err)
	}
	mod.Cfg.traceFiles = files
	goal := mod.Cfg.AstCfg.NewLabeledFormula(mod.Cfg.AstCfg.NewAtom("live_goal"), True)
	goal.TraceHook = func(trace *Trace, _ []Checker) *Trace { return markLoopStart(trace) }
	decoded := &AigerMatchHandler2{TraceBase: NewTraceBase(mod.Cfg.IuCfg, mod)}
	decoded.TraceBase.AddTraceState([]Expr{&Eq{T1: NewConst("l2s_saved", Boolean), T2: False}})
	decoded.TraceBase.AddTraceState([]Expr{&Eq{T1: NewConst("l2s_saved", Boolean), T2: True}})
	if err := CheckSubgoals([]*LabeledFormula{goal}, func(*Module) error {
		return &MCCounterexampleFailure{Trace: decoded}
	}, mod); err == nil {
		t.Fatal("expected the model-checking failure to remain visible")
	}
	var replay strings.Builder
	path := filepath.Join(mod.Cfg.TraceDir, "live_goal.a2g")
	if err := ReplayTrace(path, &replay, true, "l2s_saved"); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"repeats infinitely", "l2s_saved = false", "l2s_saved = true"} {
		if !strings.Contains(replay.String(), expected) {
			t.Errorf("replay missing %q:\n%s", expected, replay.String())
		}
	}
}

func TestRankingFailureRetainsOriginalFormulaAndAction(t *testing.T) {
	mod := New()
	name := mod.Cfg.AstCfg.NewAtom("ranking_condition")
	original := mod.Cfg.AstCfg.NewLabeledFormula(name, NewConst("original_condition", Boolean))
	transformed := mod.Cfg.AstCfg.NewLabeledFormula(name, NewConst("l2s_g_0", Boolean))
	checker := NewConjChecker(mod, transformed, 0)
	checker.Action = "ext:step"
	checker.FailedFlag = true
	trace := NewTraceForModule(mod.Cfg.IuCfg, mod, nil, nil, nil, true)
	describeRankingFailure(trace, []Checker{checker}, nil, []*LabeledFormula{original})
	if rendered := trace.String(); !strings.Contains(rendered, "postcondition ranking_condition has failed executing action step") ||
		!strings.Contains(rendered, "original_condition") {
		t.Fatalf("missing original proof failure context:\n%s", rendered)
	}
}

func TestCheckSubgoalsMethodFailureUsesTraceHookReturnLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im
from ivy import logic as lg

im.module = im.Module()

class T(ivy_trace.TraceBase):
    def get_universes(self):
        return None

original = T()
original.add_state([lg.Eq(lg.Const("original", lg.Boolean), lg.true)])
replacement = T()
replacement.add_state([lg.Eq(lg.Const("hooked", lg.Boolean), lg.true)])

def method():
    return original

def trace_hook(foo):
    return replacement

foo = method()
if foo:
    foo = trace_hook(foo)
print(json.dumps(str(foo)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python hook-return oracle failed: %v\n%s", err, out)
	}
	var want string
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python hook-return oracle %q: %v", out, err)
	}

	cfg := NewAstConfig()
	goal := cfg.NewLabeledFormula(cfg.NewAtom("goal"), True)
	original := NewTrace(nil, nil, nil, nil, true)
	original.AddState([]Expr{&Eq{T1: NewConst("original", Boolean), T2: True}})
	replacement := NewTrace(nil, nil, nil, nil, true)
	replacement.AddState([]Expr{&Eq{T1: NewConst("hooked", Boolean), T2: True}})
	goal.TraceHook = TraceHookFn(func(trace *Trace, fcs []Checker) *Trace {
		if trace != original {
			t.Fatalf("trace hook received wrong failure object: %p != %p", trace, original)
		}
		return replacement
	})

	mod := New()
	mod.Cfg = NewConfig()
	mod.Cfg.Diagnose = true
	var guiTarget interface{}
	mod.Cfg.GuiArtHook = func(m *Module, target interface{}, isCti *Clauses) error {
		guiTarget = target
		return nil
	}

	err = CheckSubgoals([]*LabeledFormula{goal}, func(_ *Module) error {
		return &TraceFailure{Trace: original}
	}, mod)
	if err == nil {
		t.Fatal("CheckSubgoals method unexpectedly succeeded")
	}
	gotTrace, ok := guiTarget.(*Trace)
	if !ok {
		t.Fatalf("GuiArt target = %T, want *Trace", guiTarget)
	}
	if got := gotTrace.String(); got != want {
		t.Fatalf("Go method hook result differs from Python\nwant:\n%q\ngot:\n%q", want, got)
	}
}
