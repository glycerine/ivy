package goivy

import (
	"encoding/json"
	"fmt"
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

	mod := NewModule()
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

	mod := NewModule()
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
