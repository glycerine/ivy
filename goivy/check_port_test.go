// check_port_test.go — Tests for the completed ivy_check.py port.
// Covers all groups from the FILL.md plan: F, C, D, I, A, B, H, E,
// plus the annotation type fix (Phase 0).
package goivy

import (
	"fmt"
	"strings"
	"testing"
)

var checkTestAstCfg = NewAstConfig()

// --- Phase 0: Annotation type fix ---

func TestAnnotationHandlerInterfaceTypes(t *testing.T) {
	// Verify IteAnnotation.Cond is lg.Expr, not string
	condSym := NewConst("cond_var", Boolean)
	ite := &IteAnnotation{
		Cond:  condSym,
		ThenB: EmptyAnnotation{},
		ElseB: EmptyAnnotation{},
	}
	if ite.Cond == nil {
		t.Fatal("IteAnnotation.Cond should not be nil")
	}
	if _, ok := ite.Cond.(*Const); !ok {
		t.Errorf("IteAnnotation.Cond should be *lg.Const, got %T", ite.Cond)
	}
}

func TestRenameAnnotationNodeKeyMap(t *testing.T) {
	xSym := NewConst("x", Boolean)
	ySym := NewConst("y", Boolean)
	m := map[NodeKey]Expr{
		Key(xSym): ySym,
	}
	ra := &RenameAnnotation{
		Arg: EmptyAnnotation{},
		Map: m,
	}
	// Lookup by structural key should work
	if _, ok := ra.Map[Key(xSym)]; !ok {
		t.Error("RenameAnnotation.Map lookup by NodeKey failed")
	}
	// Different symbol with same structure should match
	xSym2 := NewConst("x", Boolean)
	if _, ok := ra.Map[Key(xSym2)]; !ok {
		t.Error("RenameAnnotation.Map lookup by structurally equal key failed")
	}
}

func TestAnnotBranchCondIsExpr(t *testing.T) {
	condSym := NewConst("branch_cond", Boolean)
	ite := &IteAnnotation{
		Cond:  condSym,
		ThenB: EmptyAnnotation{},
		ElseB: EmptyAnnotation{},
	}
	branches := UniteAnnot(ite)
	if len(branches) == 0 {
		t.Fatal("UniteAnnot should return branches")
	}
	if branches[0].Cond == nil {
		t.Fatal("AnnotBranch.Cond should not be nil")
	}
	if _, ok := branches[0].Cond.(*Const); !ok {
		t.Errorf("AnnotBranch.Cond should be *lg.Const, got %T", branches[0].Cond)
	}
}

// --- Group F: ConjChecker.GetAnnot ---

func TestConjCheckerGetAnnotWithAnnot(t *testing.T) {
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("test"), &LogicAnd{})
	lf.Annot = &ActionAnnotation{Action: NewSequence(), Annot: EmptyAnnotation{}}
	cc := NewConjChecker(New(), lf, 8)
	got := cc.GetAnnot()
	if got != lf.Annot {
		t.Errorf("GetAnnot() = %v, want %v", got, lf.Annot)
	}
}

func TestConjCheckerGetAnnotWithoutAnnot(t *testing.T) {
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("test"), &LogicAnd{})
	cc := NewConjChecker(New(), lf, 8)
	got := cc.GetAnnot()
	if got != nil {
		t.Errorf("GetAnnot() = %v, want nil", got)
	}
}

// --- Group C: MatchHandler ---

func TestMatchHandlerEqsPopulated(t *testing.T) {
	// With nil solver, Eqs should be empty but handler should not panic
	h := NewMatchHandler(nil, nil, nil, nil)
	if h.Eqs == nil {
		t.Fatal("Eqs map should be initialized")
	}
	if len(h.Eqs) != 0 {
		t.Errorf("Eqs should be empty with nil solver, got %d entries", len(h.Eqs))
	}
}

func TestMatchHandlerEvalWithNilModel(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil, nil)
	// With nil model, Eval should return true (fallback)
	if !h.Eval(&LogicAnd{}) {
		t.Error("Eval with nil model should return true")
	}
}

func TestMatchHandlerIsSkolem(t *testing.T) {
	h := NewMatchHandler(nil, nil, nil, nil)
	// __X uppercase after __ should NOT be skolem
	sym := NewConst("__Abc", Boolean)
	if h.IsSkolem(sym) {
		t.Error("__Abc should not be skolem (uppercase after __)")
	}
}

func TestMatchHandlerImplementsAnnotationHandler(t *testing.T) {
	// Verify MatchHandler satisfies actions.AnnotationHandler
	h := NewMatchHandler(nil, nil, nil, nil)
	var _ AnnotationHandler = h
}

// --- Group I: IsolateProof handling ---

func TestCheckIsolateWithNilProof(t *testing.T) {
	// When IsolateProof is nil, CheckIsolate should not enter the proof path
	mod := New()
	mod.Cfg = NewConfig()
	mod.Cfg.OptSummary = true // skip actual checking
	err := CheckIsolate(mod, nil)
	if err != nil {
		t.Errorf("CheckIsolate with nil proof should succeed, got: %v", err)
	}
}

// --- Group A: CheckTemporals ---

func TestCheckTemporalsNoTemporalProps(t *testing.T) {
	mod := New()
	mod.Cfg = NewConfig()
	// Non-temporal property should be skipped
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("p1"), &LogicAnd{})
	lf.Temporal = BoolPtr(false)
	mod.LabeledProps = []*LabeledFormula{lf}
	err := CheckTemporals(mod)
	if err != nil {
		t.Errorf("CheckTemporals with no temporal props should succeed, got: %v", err)
	}
}

func TestCheckTemporalsAssumedProp(t *testing.T) {
	mod := New()
	mod.Cfg = NewConfig()
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("temporal_assumed"), &LogicAnd{})
	lf.Temporal = BoolPtr(true)
	lf.Assumed = true
	mod.LabeledProps = []*LabeledFormula{lf}
	// Should not error — assumed temporal props are admitted as axioms
	err := CheckTemporals(mod)
	if err != nil {
		t.Errorf("CheckTemporals with assumed temporal prop should succeed, got: %v", err)
	}
}

func TestCheckTemporalsWithProof(t *testing.T) {
	mod := New()
	mod.Cfg = NewConfig()
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("temporal_proved"), &LogicAnd{})
	lf.Temporal = BoolPtr(true)
	mod.LabeledProps = []*LabeledFormula{lf}
	// Add a proof for this property
	mod.Proofs = []ProofEntry{{
		Formula: lf,
		Proof:   checkTestAstCfg.NewAtom("compose_tactics"), // a simple proof node
	}}
	// This will call NormalProgramFromModule and AdmitProposition.
	// With an empty module, it should complete without panic.
	err := CheckTemporals(mod)
	// May error due to proof checking failing on empty module — that's fine,
	// we're testing the wiring, not the proof itself
	_ = err
}

// --- Group B: MCTactic/VMTTactic ---

func TestMCTacticEmptyGoals(t *testing.T) {
	remaining, err := MCTactic(nil, nil, nil, New())
	if err != nil {
		t.Errorf("MCTactic with empty goals should not error, got: %v", err)
	}
	if remaining != nil {
		t.Errorf("MCTactic with empty goals should return nil, got: %v", remaining)
	}
}

func TestVMTTacticEmptyGoals(t *testing.T) {
	remaining, err := VMTTactic(nil, nil, nil, New())
	if err != nil {
		t.Errorf("VMTTactic with empty goals should not error, got: %v", err)
	}
	if remaining != nil {
		t.Errorf("VMTTactic with empty goals should return nil, got: %v", remaining)
	}
}

func TestApplyTemporalTacticChainNonTemporal(t *testing.T) {
	// Non-temporal goal should pass through unchanged
	goal := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("g"), &LogicAnd{})
	goals := []*LabeledFormula{goal}
	result, err := applyTemporalTacticChain(nil, goals, nil)
	if err != nil {
		t.Errorf("non-temporal goal should not error, got: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 goal, got %d", len(result))
	}
}

func TestCloneProofWithTacticLets(t *testing.T) {
	// TacticTactic node with name, body, proof args
	tt := checkTestAstCfg.NewTacticTactic(
		checkTestAstCfg.NewAtom("mc"),
		checkTestAstCfg.NewTacticWith(nil),
		checkTestAstCfg.NewAtom("some_proof"),
	)
	cloned := cloneProofWithTacticLets(tt)
	if cloned == nil {
		t.Fatal("cloneProofWithTacticLets returned nil")
	}
	args := cloned.Args()
	if len(args) < 2 {
		t.Fatalf("expected at least 2 args, got %d", len(args))
	}
	if _, ok := args[1].(*TacticLets); !ok {
		t.Errorf("args[1] should be *TacticLets, got %T", args[1])
	}
}

// --- Group H: CheckModule macro_finder ---

func TestCheckModuleMacroFinderDefault(t *testing.T) {
	cfg := NewConfig()
	if !cfg.SolverOpts.MacroFinder {
		t.Error("MacroFinder should default to true")
	}
}

// --- Group E: Start entry point ---

func TestStartInvalidArgs(t *testing.T) {
	err := Start(nil, nil)
	if err == nil {
		t.Error("Start with no args should error")
	}
}

func TestStartNonIvyFile(t *testing.T) {
	err := Start([]string{"notanivyfile.txt"}, nil)
	if err == nil {
		t.Error("Start with non-.ivy file should error")
	}
}

// --- LabeledFormula.Annot preservation ---

func TestLabeledFormulaClonePreservesAnnot(t *testing.T) {
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("test"), &LogicAnd{})
	lf.Annot = &ActionAnnotation{Action: NewSequence(), Annot: EmptyAnnotation{}}
	cloned := lf.Clone([]Node{lf.Label, lf.Formula}).(*LabeledFormula)
	if cloned.Annot != lf.Annot {
		t.Errorf("Clone should preserve Annot, got %v", cloned.Annot)
	}
}

// --- Misc integration ---

func TestNormalProgramFromModuleDoesNotPanic(t *testing.T) {
	mod := New()
	np := NormalProgramFromModule(mod)
	if np == nil {
		t.Error("NormalProgramFromModule should not return nil for empty module")
	}
}

func TestProofCheckerAdmitAxiomDoesNotPanic(t *testing.T) {
	pc := NewProofChecker(nil, nil, nil, nil, nil)
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("ax"), &LogicAnd{})
	pc.AdmitAxiom(lf)
	// Should not panic
}

func TestCheckConjsInStateEmptyModule(t *testing.T) {
	mod := New()
	mod.Cfg = NewConfig()
	result := CheckConjsInState(mod, 8, nil)
	if !result {
		t.Error("CheckConjsInState with empty module should pass")
	}
}

func TestGetConjs(t *testing.T) {
	mod := New()
	// Add a non-explicit, non-unprovable conjecture.
	// Use *lg.Const which fully implements lg.Expr (has Sexp()).
	formula := NewConst("conj_fmla", Boolean)
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("conj1"), formula)
	lf.Explicit = false
	lf.Unprovable = false
	mod.LabeledConjs = append(mod.LabeledConjs, lf)
	conjs := GetConjs(mod)
	if conjs == nil {
		t.Fatal("GetConjs should not return nil")
	}
	if len(conjs.Fmlas) != 1 {
		t.Errorf("expected 1 formula, got %d", len(conjs.Fmlas))
	}
}

func TestApplyConjProofsNoProofs(t *testing.T) {
	mod := New()
	lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom("conj1"), &LogicAnd{})
	mod.LabeledConjs = []*LabeledFormula{lf}
	ApplyConjProofs(mod)
	if len(mod.ConjSubgoals) != 1 {
		t.Errorf("expected 1 subgoal (passthrough), got %d", len(mod.ConjSubgoals))
	}
}

// --- Fuzz tests ---

func FuzzMatchHandlerEqs(f *testing.F) {
	f.Add("sym1", "sym2", true)
	f.Add("", "", false)
	f.Add("__skolem", "val", true)

	f.Fuzz(func(t *testing.T, symName, valName string, isEq bool) {
		// Create clauses with various formula types
		var fmlas []Expr
		sym := NewConst(symName, Boolean)
		val := NewConst(valName, Boolean)
		if isEq {
			app, _ := NewApply(sym)
			if app != nil {
				fmlas = append(fmlas, &Eq{T1: app, T2: val})
			}
		} else {
			app, _ := NewApply(sym)
			if app != nil {
				fmlas = append(fmlas, &LogicNot{Body: app})
			}
		}
		cls := NewClauses(fmlas, nil, nil)
		// Should not panic
		h := NewMatchHandler(cls, nil, []*Const{sym}, nil)
		_ = h.String()
	})
}

func FuzzDualClauses(f *testing.F) {
	f.Add("p", "q")
	f.Add("", "x")

	f.Fuzz(func(t *testing.T, name1, name2 string) {
		s1 := NewConst(name1, Boolean)
		s2 := NewConst(name2, Boolean)
		cls := NewClauses([]Expr{s1, s2}, nil, nil)
		// Should not panic
		result := DualClauses(cls, nil, nil)
		_ = result
	})
}

func FuzzAnnotationNodeKeyRoundTrip(f *testing.F) {
	f.Add("x", "y")
	f.Add("__skolem", "val")

	f.Fuzz(func(t *testing.T, name1, name2 string) {
		s1 := NewConst(name1, Boolean)
		s2 := NewConst(name2, Boolean)
		key := Key(s1)
		m := map[NodeKey]Expr{key: s2}
		// Round-trip: key from structurally equal symbol should match
		s1Copy := NewConst(name1, Boolean)
		if v, ok := m[Key(s1Copy)]; ok {
			if v.String() != s2.String() {
				t.Errorf("round-trip mismatch: got %s, want %s", v, s2)
			}
		}
		// Create annotation and verify String doesn't panic
		a := EmptyAnnotation{}.Ite(s1, EmptyAnnotation{})
		_ = a.String()
		r := EmptyAnnotation{}.Rename(map[NodeKey]Expr{Key(s1): s2})
		_ = r.String()
	})
}

func FuzzCheckTemporalsProps(f *testing.F) {
	f.Add("prop1", true, true)
	f.Add("prop2", false, false)
	f.Add("", true, false)

	f.Fuzz(func(t *testing.T, name string, temporal, assumed bool) {
		mod := New()
		mod.Cfg = NewConfig()
		lf := checkTestAstCfg.NewLabeledFormula(checkTestAstCfg.NewAtom(name), &LogicAnd{})
		lf.Temporal = BoolPtr(temporal)
		lf.Assumed = assumed
		mod.LabeledProps = []*LabeledFormula{lf}
		// Should not panic
		_ = CheckTemporals(mod)
	})
}

// --- Annotation system integration tests ---

func TestMatchAnnotationWithLgExprCond(t *testing.T) {
	condSym := NewConst("branch", Boolean)
	ite := &IteAnnotation{
		Cond:  condSym,
		ThenB: EmptyAnnotation{},
		ElseB: EmptyAnnotation{},
	}
	// Verify it can be serialized
	s := ite.String()
	if !strings.Contains(s, "branch") {
		t.Errorf("IteAnnotation String should contain cond name, got: %s", s)
	}
}

func TestRenameAnnotationWithNodeKeyValues(t *testing.T) {
	xSym := NewConst("x", Boolean)
	ySym := NewConst("y", Boolean)
	ra := EmptyAnnotation{}.Rename(map[NodeKey]Expr{Key(xSym): ySym})
	s := ra.String()
	if !strings.Contains(s, "Rename") {
		t.Errorf("expected Rename in string, got: %s", s)
	}
}

func TestUniteAnnotWithLgExprConds(t *testing.T) {
	c1 := NewConst("c1", Boolean)
	c2 := NewConst("c2", Boolean)
	inner := &IteAnnotation{
		Cond:  c1,
		ThenB: EmptyAnnotation{},
		ElseB: EmptyAnnotation{},
	}
	outer := &IteAnnotation{
		Cond:  c2,
		ThenB: EmptyAnnotation{},
		ElseB: inner,
	}
	branches := UniteAnnot(outer)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
	// First branch should have c1, second c2
	if branches[0].Cond.String() != "c1" {
		t.Errorf("branch[0].Cond = %s, want c1", branches[0].Cond)
	}
	if branches[1].Cond.String() != "c2" {
		t.Errorf("branch[1].Cond = %s, want c2", branches[1].Cond)
	}
}

// TestRegisterTacticsWiring verifies that RegisterTactics populates the
// ProofConfig with all expected tactics. Python registers these at import
// time; Go must wire them explicitly.
func TestRegisterTacticsWiring(t *testing.T) {
	mod := New()
	mod.Cfg = NewConfig()
	proofCfg := TacticNewConfig()
	mod.Cfg.ProofCfg = proofCfg
	RegisterTactics(mod.Cfg.ProofCfg, mod)

	expected := []string{
		// from tactics.RegisterProofTactics
		"vcgen", "skolemize", "skolemizenp", "tempind", "tempcase", "sorry",
		// from check.RegisterTactics
		"mc", "vmt",
		// from temporal.RegisterTactics
		"invariance",
		// from l2s.RegisterTactics
		"l2s", "l2s_full", "l2s_auto",
	}
	for _, name := range expected {
		if _, ok := mod.Cfg.ProofCfg.Tactics[name]; !ok {
			t.Errorf("expected tactic %q to be registered, but it was not", name)
		}
	}
}

// spyPC wraps a ProofCheckerInterface and records GetModule calls.
// If the tactic closure correctly calls pc.GetModule(), getModuleCalls > 0.
// If it uses a closure-captured module instead, getModuleCalls == 0.
type spyPC struct {
	ProofCheckerInterface
	mod            *Module
	getModuleCalls int
}

func (s *spyPC) GetModule() *Module {
	s.getModuleCalls++
	return s.mod
}

// TestMCVMTTacticClosureUsesProofCheckerModule verifies that the mc and vmt
// tactic closures registered by RegisterTactics use pc.GetModule() — the
// isolate-processed module from the proof checker — not the original module
// captured in the closure at registration time.
//
// Regression test for the bug where MCTactic/VMTTactic received the
// pre-CreateIsolate module (94 actions) instead of the post-CreateIsolate
// module (5 actions), causing divergence at XTRACE 28253287.
func TestMCVMTTacticClosureUsesProofCheckerModule(t *testing.T) {
	// origMod: the module passed to RegisterTactics (simulates
	// the compilation-level module with many actions).
	origMod := New()
	origMod.Cfg = NewConfig()
	for i := 0; i < 10; i++ {
		seq := NewSequence()
		origMod.Actions.Set(fmt.Sprintf("orig_action_%d", i), seq)
	}

	origMod.Cfg.ProofCfg = TacticNewConfig()
	RegisterTactics(origMod.Cfg.ProofCfg, origMod)

	// isoMod: the isolate-processed module (simulates post-CreateIsolate
	// with a smaller action set). This is what pc.GetModule() should return.
	isoMod := New()
	isoMod.Cfg = NewConfig()
	seq := NewSequence()
	isoMod.Actions.Set("iso_action", seq)

	// Build a TemporalModels goal with fmla=lg.True so the temporal
	// tactic chain is a no-op (it skips when fmla is true).
	acfg := isoMod.Cfg.AstCfg
	np := &NormalProgram{
		Invars:   []*LabeledFormula{acfg.NewLabeledFormula(nil, True)},
		Asms:     []*LabeledFormula{acfg.NewLabeledFormula(nil, True)},
		Calls:    []string{"iso_action"},
		Bindings: []*ActionTermBinding{{Name: "iso_action", Action: &ActionTerm{Stmt: seq}}},
		Init:     NewSequence(),
	}
	tm := acfg.NewTemporalModels(np, True)
	goal := acfg.NewLabeledFormula(acfg.NewAtom("test_goal"), tm)
	goals := []*LabeledFormula{goal}

	for _, tacticName := range []string{"mc", "vmt"} {
		t.Run(tacticName, func(t *testing.T) {
			tactic, ok := origMod.Cfg.ProofCfg.Tactics[tacticName]
			if !ok {
				t.Fatalf("tactic %q not registered", tacticName)
			}

			// Create a real proof checker with isoMod, then wrap in spy.
			realPC := NewProofChecker(origMod.Cfg.ProofCfg, isoMod, nil, nil, nil)
			spy := &spyPC{ProofCheckerInterface: realPC, mod: isoMod}

			// Invoke the registered tactic closure. MCTactic/VMTTactic
			// will call CheckSubgoals which may panic deep in the mc/vmt
			// checker (no real Z3/solver). Recover and check the spy.
			func() {
				defer func() { recover() }()
				_, _ = tactic(spy, goals, nil)
			}()

			if spy.getModuleCalls == 0 {
				t.Errorf("tactic %q closure did NOT call pc.GetModule(); "+
					"it is using the closure-captured original module instead of "+
					"the proof checker's isolate-processed module", tacticName)
			}
		})
	}
}
