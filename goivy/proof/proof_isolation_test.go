package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/module"
)

// TestConfigIsolation verifies that two independent ProofConfig instances
// have isolated tactic registries — registering on one does not affect the other.
func TestConfigIsolation(t *testing.T) {
	cfg1 := module.TacticNewConfig()
	cfg2 := module.TacticNewConfig()

	called1 := false
	cfg1.RegisterTactic("only_on_1", func(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
		called1 = true
		return goals, nil
	})

	// cfg1 should have the tactic
	if _, ok := cfg1.Tactics["only_on_1"]; !ok {
		t.Fatal("tactic should be registered on cfg1")
	}

	// cfg2 should NOT have the tactic
	if _, ok := cfg2.Tactics["only_on_1"]; ok {
		t.Fatal("tactic should NOT be registered on cfg2 — configs must be isolated")
	}

	// Register a different tactic on cfg2
	called2 := false
	cfg2.RegisterTactic("only_on_2", func(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
		called2 = true
		return goals, nil
	})

	if _, ok := cfg1.Tactics["only_on_2"]; ok {
		t.Fatal("tactic from cfg2 should NOT leak into cfg1")
	}

	// Call tactics to verify they work
	tac1 := cfg1.Tactics["only_on_1"]
	tac1(nil, nil, nil)
	if !called1 {
		t.Error("cfg1 tactic should have been called")
	}

	tac2 := cfg2.Tactics["only_on_2"]
	tac2(nil, nil, nil)
	if !called2 {
		t.Error("cfg2 tactic should have been called")
	}
}
