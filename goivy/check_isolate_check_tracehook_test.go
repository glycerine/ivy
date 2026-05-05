package goivy

import (
	"fmt"
	"testing"
)

func TestCheckSubgoalsMethodFailureAppliesTraceHook(t *testing.T) {
	cfg := NewAstConfig()
	hookCalled := false
	goal := cfg.NewLabeledFormula(cfg.NewAtom("goal"), True)
	goal.TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
		hookCalled = true
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
