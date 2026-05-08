package goivy

import (
	"testing"
)

// mockProofChecker implements module.ProofCheckerInterface for testing.
type mockProofChecker struct {
	subgoals []*LabeledFormula
}

func (m *mockProofChecker) AdmitProposition(prop *LabeledFormula, proof Node, existingSubgoals ...*LabeledFormula) ([]*LabeledFormula, error) {
	return m.subgoals, nil
}
func (m *mockProofChecker) GetSubgoals(prop *LabeledFormula, proof Node) ([]*LabeledFormula, error) {
	return m.subgoals, nil
}
func (m *mockProofChecker) AdmitDefinition(defn *LabeledFormula, proof Node) ([]*LabeledFormula, error) {
	return nil, nil
}
func (m *mockProofChecker) SetLastAxiom(prop *LabeledFormula) {}
func (m *mockProofChecker) SetSchema(name string, prop *LabeledFormula) {
}
func (m *mockProofChecker) GetModule() *Module           { return nil }
func (m *mockProofChecker) GetAstCfg() *AstConfig        { return nil }
func (m *mockProofChecker) GetAxioms() []*LabeledFormula { return nil }

// newTestModule creates a minimal module with a CompilerConfig for testing.
func newTestModule(verifying bool) *Module {
	mod := NewModule()
	mod.Sig = NewSig()
	cc := &CompilerConfig{OptionVerifying: verifying}
	mod.CompCfg = cc
	return mod
}

// TestApplyAssertProofsWithProver_BasicSubgoal verifies that an AssertAction
// with a proof is replaced by Sequence(SubgoalActions... + AssumeAction).
func TestApplyAssertProofsWithProver_BasicSubgoal(t *testing.T) {
	cfg := NewAstConfig()
	mod := newTestModule(true)

	// Create an AssertAction with a proof
	cond := True
	aa := NewAssertAction(cond)
	aa.Proof = WrapTactic(cfg.NewComposeTactics(nil))

	mod.Actions.Set("test_act", aa)

	// Mock prover returns 2 subgoals
	sg1 := cfg.NewLabeledFormula(cfg.NewAtom("sg1"), True)
	sg2 := cfg.NewLabeledFormula(cfg.NewAtom("sg2"), True)
	prover := &mockProofChecker{subgoals: []*LabeledFormula{sg1, sg2}}

	err := ApplyAssertProofsWithProver(mod, prover)
	if err != nil {
		t.Fatalf("ApplyAssertProofsWithProver failed: %v", err)
	}

	result, ok := mod.Actions.Get("test_act").(ActionsAction)
	if !ok {
		t.Fatalf("expected Action, got %T", mod.Actions.Get("test_act"))
	}
	seq, ok := result.(*LogicSequence)
	if !ok {
		t.Fatalf("expected Sequence, got %T: %v", result, result)
	}
	// 2 subgoals + 1 assume = 3
	args := seq.ActionArgs()
	if len(args) != 3 {
		t.Fatalf("expected 3 children (2 subgoals + 1 assume), got %d", len(args))
	}

	// First two should be SubgoalAction
	for i := 0; i < 2; i++ {
		sub, _ := args[i].(ActionsAction)
		if _, ok := sub.(*LogicSubgoalAction); !ok {
			t.Errorf("arg[%d]: expected SubgoalAction, got %T", i, sub)
		}
	}

	// Last should be AssumeAction
	last, _ := args[2].(ActionsAction)
	if _, ok := last.(*LogicAssumeAction); !ok {
		t.Errorf("arg[2]: expected AssumeAction, got %T", last)
	}
}

// TestApplyAssertProofsWithProver_NoProof verifies that an AssertAction
// without a proof is left unchanged.
func TestApplyAssertProofsWithProver_NoProof(t *testing.T) {
	mod := newTestModule(true)

	aa := NewAssertAction(True)
	// No proof set
	mod.Actions.Set("test_act", aa)

	prover := &mockProofChecker{}
	err := ApplyAssertProofsWithProver(mod, prover)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mod.Actions.Get("test_act").(ActionsAction)
	if _, ok := result.(*LogicAssertAction); !ok {
		t.Fatalf("expected AssertAction unchanged, got %T", result)
	}
}

// TestApplyAssertProofsWithProver_NotVerifying verifies that when not
// verifying, an AssertAction with proof is replaced with a plain
// AssertAction (proof stripped).
func TestApplyAssertProofsWithProver_NotVerifying(t *testing.T) {
	cfg := NewAstConfig()
	mod := newTestModule(false) // verifying = false

	aa := NewAssertAction(True)
	aa.Proof = WrapTactic(cfg.NewComposeTactics(nil))
	mod.Actions.Set("test_act", aa)

	prover := &mockProofChecker{}
	err := ApplyAssertProofsWithProver(mod, prover)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mod.Actions.Get("test_act").(ActionsAction)
	ra, ok := result.(*LogicAssertAction)
	if !ok {
		t.Fatalf("expected AssertAction, got %T", result)
	}
	// Proof should be stripped
	if ra.Proof != nil {
		t.Error("expected proof to be stripped when not verifying")
	}
}

// TestApplyAssertProofsWithProver_WhileInvariantFlattening verifies that
// WhileAction invariants containing AssertActions with proofs are flattened
// when the prover returns multiple subgoals.
func TestApplyAssertProofsWithProver_WhileInvariantFlattening(t *testing.T) {
	cfg := NewAstConfig()
	mod := newTestModule(true)

	// Build a WhileAction with one invariant that is an AssertAction with proof
	invAssert := NewAssertAction(True)
	invAssert.Proof = WrapTactic(cfg.NewComposeTactics(nil))

	body := NewSequence() // empty body
	w := NewWhileAction(True, body,
		invAssert)

	mod.Actions.Set("test_act", w)

	// Mock prover returns 2 subgoals → will produce a Sequence with 3 children
	sg1 := cfg.NewLabeledFormula(cfg.NewAtom("sg1"), True)
	sg2 := cfg.NewLabeledFormula(cfg.NewAtom("sg2"), True)
	prover := &mockProofChecker{subgoals: []*LabeledFormula{sg1, sg2}}

	err := ApplyAssertProofsWithProver(mod, prover)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mod.Actions.Get("test_act").(ActionsAction)
	rw, ok := result.(*LogicWhileAction)
	if !ok {
		t.Fatalf("expected WhileAction, got %T", result)
	}

	// The original 1 invariant (assert with proof → Sequence of 3) should be
	// flattened into 3 invariant entries (2 subgoals + 1 assume).
	if len(rw.Invariants) != 3 {
		t.Fatalf("expected 3 flattened invariants (2 subgoals + 1 assume), got %d", len(rw.Invariants))
	}
}

// TestApplyAssertProofsWithProver_Nested verifies that only AssertActions
// with proofs are transformed; other nested actions are left alone.
func TestApplyAssertProofsWithProver_Nested(t *testing.T) {
	cfg := NewAstConfig()
	mod := newTestModule(true)

	// AssertAction with proof (inside a LocalAction)
	assertWithProof := NewAssertAction(True)
	assertWithProof.Proof = WrapTactic(cfg.NewComposeTactics(nil))

	// LocalAction wrapping the assert-with-proof
	actCfg := NewActionsConfig()
	localAct := NewLocalActionOn(actCfg, "test", assertWithProof)

	// AssertAction without proof
	assertNoProof := NewAssertAction(True)

	// Sequence of both
	seq := NewSequence(
		localAct,
		assertNoProof,
	)
	mod.Actions.Set("test_act", seq)

	sg1 := cfg.NewLabeledFormula(cfg.NewAtom("sg1"), True)
	prover := &mockProofChecker{subgoals: []*LabeledFormula{sg1}}

	err := ApplyAssertProofsWithProver(mod, prover)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := mod.Actions.Get("test_act").(ActionsAction)
	rSeq, ok := result.(*LogicSequence)
	if !ok {
		t.Fatalf("expected Sequence, got %T", result)
	}

	args := rSeq.ActionArgs()
	if len(args) != 2 {
		t.Fatalf("expected 2 children, got %d", len(args))
	}

	// First child: LocalAction whose body should now be a Sequence (proof was expanded)
	firstAct, _ := args[0].(ActionsAction)
	la, ok := firstAct.(*LogicLocalAction)
	if !ok {
		t.Fatalf("arg[0]: expected LocalAction, got %T", firstAct)
	}
	bodyAct, _ := la.Body.(ActionsAction)
	if _, ok := bodyAct.(*LogicSequence); !ok {
		t.Errorf("LocalAction body: expected Sequence (expanded proof), got %T", bodyAct)
	}

	// Second child: AssertAction should be unchanged (no proof)
	secondAct, _ := args[1].(ActionsAction)
	if _, ok := secondAct.(*LogicAssertAction); !ok {
		t.Errorf("arg[1]: expected AssertAction, got %T", secondAct)
	}
}

// TestCompileAssertFormula_WithProof verifies that compiling an assert AST
// node with a proof tactic populates AssertAction.Proof.
func TestCompileAssertFormula_WithProof(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	// Build: Atom("assert", trueAtom, ComposeTactics{})
	formula := cfg.NewAtom("true")
	proof := cfg.NewComposeTactics(nil)
	assertAtom := cfg.NewAtom("assert", formula, proof)

	act, err := c.CompileActionBody(assertAtom)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	aa, ok := act.(*LogicAssertAction)
	if !ok {
		t.Fatalf("expected AssertAction, got %T", act)
	}
	if aa.Proof == nil {
		t.Fatal("expected Proof to be set, got nil")
	}
}

// TestCompileAssertFormula_NoProof verifies that compiling an assert AST
// node without a proof leaves AssertAction.Proof nil.
func TestCompileAssertFormula_NoProof(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	formula := cfg.NewAtom("true")
	assertAtom := cfg.NewAtom("assert", formula)

	act, err := c.CompileActionBody(assertAtom)
	if err != nil {
		t.Fatalf("CompileActionBody failed: %v", err)
	}

	aa, ok := act.(*LogicAssertAction)
	if !ok {
		t.Fatalf("expected AssertAction, got %T", act)
	}
	if aa.Proof != nil {
		t.Errorf("expected Proof to be nil, got %v", aa.Proof)
	}
}
