package check

import (
	"fmt"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

func TestCheckSubgoalsMethodFailureAppliesTraceHook(t *testing.T) {
	cfg := ast.NewAstConfig()
	hookCalled := false
	goal := cfg.NewLabeledFormula(cfg.NewAtom("goal"), lg.True)
	goal.TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
		hookCalled = true
	})

	mod := module.New()
	mod.Cfg = module.NewConfig()
	err := CheckSubgoals([]*ast.LabeledFormula{goal}, func(_ *module.Module) error {
		return fmt.Errorf("method counterexample")
	}, mod)
	if err == nil {
		t.Fatal("CheckSubgoals method unexpectedly succeeded")
	}
	if !hookCalled {
		t.Fatal("method subgoal failure did not apply goal.TraceHook before returning the failure")
	}
}
