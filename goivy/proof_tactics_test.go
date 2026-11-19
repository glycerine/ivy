package goivy

import (
	"strings"
	"testing"
)

func TestLetTacticCompilesDefinitionsAsEqualityAtoms(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	sortT := &UninterpretedSort{Name: "t"}
	if err := mod.Sig.AddSort(sortT); err != nil {
		t.Fatalf("AddSort(t): %v", err)
	}
	if _, err := mod.Sig.AddSymbol("x", sortT); err != nil {
		t.Fatalf("AddSymbol(x): %v", err)
	}
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)
	goal := cfg.NewLabeledFormula(cfg.NewAtom("asrt1"), True)
	def := cfg.NewDefinition(cfg.NewVariable("X", "S"), cfg.NewAtom("x"))
	proof := cfg.NewLetTactic([]Node{def})

	out := captureActionUpdateStdout(t, func() {
		_, _ = pc.letTactic([]*LabeledFormula{goal}, proof)
	})

	if !strings.Contains(out, "XTRACE: proof.CompileExprVocab ENTER exprType=Atom\n") {
		t.Fatalf("let tactic did not compile an equality Atom:\n%s", out)
	}
	if strings.Contains(out, "XTRACE: proof.CompileExprVocab ENTER exprType=Definition\n") {
		t.Fatalf("let tactic compiled the Definition node directly:\n%s", out)
	}
	if !strings.Contains(out, "proof.WrapImplies EXIT type=lgImplies HASH canon=(Implies t1:(And terms:[") {
		t.Fatalf("let tactic did not preserve Python's single-definition And wrapper:\n%s", out)
	}
}

// Python let_tactic returns attrib_goals(proof,[subgoal]) + decls[1:]:
// the new subgoal takes the let proof's line, not the goal's
// (test/havocmc.ivy prints "line 17: invar2", the proof's line).
func TestLetTacticAttributesSubgoalToProof(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	sortT := &UninterpretedSort{Name: "t"}
	if err := mod.Sig.AddSort(sortT); err != nil {
		t.Fatal(err)
	}
	if _, err := mod.Sig.AddSymbol("x", sortT); err != nil {
		t.Fatal(err)
	}
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)
	goal := cfg.NewLabeledFormula(cfg.NewAtom("asrt1"), True)
	goal.SetLineno(Location{Filename: "f.ivy", Line: 16})
	def := cfg.NewDefinition(cfg.NewVariable("X", "S"), cfg.NewAtom("x"))
	proof := cfg.NewLetTactic([]Node{def})
	proof.SetLineno(Location{Filename: "f.ivy", Line: 17})
	res, err := pc.letTactic([]*LabeledFormula{goal}, proof)
	if err != nil || len(res) != 1 {
		t.Fatal(res, err)
	}
	if got := res[0].GetLineno().Line; got != 17 {
		t.Fatalf("subgoal line %d, want the proof's 17", got)
	}
}
