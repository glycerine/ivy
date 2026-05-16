//go:build web

package webui

import (
	"encoding/json"
	goivy "github.com/glycerine/ivy/goivy"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- CTI struct tests ---

func TestCTIAnalysisGraphUIFields(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	if ui.HaveCTI {
		t.Error("HaveCTI should be false initially")
	}
	if ui.CurrentConjecture != nil {
		t.Error("CurrentConjecture should be nil initially")
	}
	if ui.CurrentBound != -1 {
		t.Errorf("CurrentBound should be -1, got %d", ui.CurrentBound)
	}
	if ui.RelationsToMinimize != "relations to minimize" {
		t.Errorf("RelationsToMinimize wrong: %q", ui.RelationsToMinimize)
	}
	if ui.Mod != nil {
		t.Error("Mod should be nil when created with nil module")
	}
	if ui.Solver != nil {
		t.Error("Solver should be nil when created with nil module")
	}
}

func TestCTIConjecturesTypeClauses(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	c1 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	c2 := goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil)
	ui.Conjectures = []*goivy.Clauses{c1, c2}

	if len(ui.Conjectures) != 2 {
		t.Fatalf("expected 2 conjectures, got %d", len(ui.Conjectures))
	}
	if ui.Conjectures[0] != c1 {
		t.Error("first conjecture should be c1")
	}
	if ui.Conjectures[1] != c2 {
		t.Error("second conjecture should be c2")
	}
}

// --- AutodetectTransitive tests ---

func TestAutodetectTransitiveNilModule(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ui.AutodetectTransitive()
	if len(ui.TransitiveRelations) != 0 {
		t.Errorf("expected 0 transitive relations with nil module, got %d", len(ui.TransitiveRelations))
	}
}

func TestAutodetectTransitiveEmptySig(t *testing.T) {
	mod := goivy.New()
	ui := NewCTIAnalysisGraphUI(mod)
	ui.AutodetectTransitive()
	if len(ui.TransitiveRelations) != 0 {
		t.Errorf("expected 0 transitive relations with empty sig, got %d", len(ui.TransitiveRelations))
	}
}

// --- CheckInductiveness tests ---

func TestCheckInductivenessNilModule(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ok, msg := ui.CheckInductiveness()
	if !ok {
		t.Error("expected inductive with nil module")
	}
	if !strings.Contains(msg, "Inductive") {
		t.Errorf("expected 'Inductive' in message, got: %s", msg)
	}
	if ui.HaveCTI {
		t.Error("HaveCTI should be false")
	}
}

func TestCheckInductivenessNoConjectures(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ui.Conjectures = nil
	ok, _ := ui.CheckInductiveness()
	if !ok {
		t.Error("expected inductive with no conjectures")
	}
}

// --- BoundedCheck tests ---

func TestBoundedCheckNilModule(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	found, msg := ui.BoundedCheck(0, nil)
	if found {
		t.Error("should not find counter-example with nil module")
	}
	if !strings.Contains(msg, "no module") {
		t.Errorf("expected 'no module' in message, got: %s", msg)
	}
}

func TestBoundedCheckNoConjectures(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ui.Conjectures = nil
	found, msg := ui.BoundedCheck(5, nil)
	if found {
		t.Error("should not find counter-example with no conjectures")
	}
	if !strings.Contains(msg, "no module") {
		t.Errorf("expected 'no module' in message, got: %s", msg)
	}
}

func TestBoundedCheckStoresBound(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ui.BoundedCheck(7, nil)
	if ui.CurrentBound != 7 {
		t.Errorf("CurrentBound should be 7, got %d", ui.CurrentBound)
	}
}

func TestBoundedCheckUsesAddInitialStateForInitializers(t *testing.T) {
	cmd := pythonIvyCommandForWebTest(t, "-O", "-c", `
import json
from ivy import ivy_art, ivy_module as im, ivy_actions as act, logic as lg

im.module = im.Module()
im.module.initializers = [("init", act.AssumeAction(lg.true))]
ag = ivy_art.AnalysisGraph()
ag.add_initial_state(ag.init_cond)
print(json.dumps({
    "states": len(ag.states),
    "has_initializer_expr": getattr(ag.states[0], "expr", None) is not None,
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python add_initial_state initializer oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		States             int  `json:"states"`
		HasInitializerExpr bool `json:"has_initializer_expr"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python add_initial_state initializer oracle %q: %v", out, err)
	}
	if want.States != 1 || !want.HasInitializerExpr {
		t.Fatalf("python add_initial_state oracle did not run initializer path: %+v", want)
	}

	S := &goivy.UninterpretedSort{Name: "S"}
	X, _ := goivy.NewVariable("X", S)
	pSort, _ := goivy.NewFunctionSort(S, goivy.Boolean)
	P := goivy.NewConst("P", pSort)
	PX, _ := goivy.NewApply(P, X)

	mod := goivy.New()
	mod.Sig = goivy.NewSig()
	mod.Sig.Sorts.Set("S", S)
	mod.Sig.Symbols.Set("P", &goivy.SymbolEntry{Sort: pSort})
	mod.Relations.Set("P", pSort)
	mod.Initializers = append(mod.Initializers, goivy.NamedAction{
		Name:   "init",
		Action: goivy.NewAssignAction(PX, goivy.True),
	})

	conj := goivy.NewClauses([]goivy.Expr{PX}, nil, goivy.EmptyAnnotation{})
	ui := NewCTIAnalysisGraphUI(mod)
	ui.Conjectures = []*goivy.Clauses{conj}

	found, msg := ui.BoundedCheck(0, nil)
	if found {
		t.Fatalf("BoundedCheck found a depth-0 counterexample even though Python add_initial_state runs initializers: %s", msg)
	}
}

func pythonIvyCommandForWebTest(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	pyivyRoot := filepath.Join(repoRoot, "pyivy", "ivy")
	python := filepath.Join(repoRoot, "pyivy", "goivy-venv", "bin", "python3")
	if _, err := os.Stat(python); err != nil {
		if _, lookErr := exec.LookPath("python3"); lookErr != nil {
			t.Skip("python3 not available")
		}
		python = "python3"
	}
	cmd := exec.Command(python, args...)
	cmd.Dir = pyivyRoot
	pythonPath := pyivyRoot
	if existing := os.Getenv("PYTHONPATH"); existing != "" {
		pythonPath += string(os.PathListSeparator) + existing
	}
	cmd.Env = append(os.Environ(),
		"IVY_HOME="+pyivyRoot,
		"PYTHONPATH="+pythonPath,
	)
	return cmd
}

// --- Diagram tests ---

func TestDiagramNilModule(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	_, err := ui.Diagram()
	if err != nil {
		t.Logf("Expected error with nil module: %v", err)
	}
}

func TestDiagramNoCtI(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ui.HaveCTI = false
	result, err := ui.Diagram()
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
	_ = result
}

func TestDiagramUsesLazyStateUpdateNotBMCValue(t *testing.T) {
	mod := goivy.New()
	mod.Actions.Set("act", goivy.NewSequence())
	ui := NewCTIAnalysisGraphUI(mod)
	ui.Solver = nil
	ui.HaveCTI = true
	ui.AG = goivy.NewAnalysisGraph(mod)

	s0 := goivy.NewState(mod, goivy.TrueClauses(nil))
	ui.AG.Add(s0, nil)
	s1 := goivy.NewState(mod, goivy.TrueClauses(nil))
	ui.AG.Add(s1, goivy.NewActionApp("act", s0))

	if s1.Update != nil {
		t.Fatalf("test setup expected no cached update")
	}
	if s1.Value != nil {
		t.Fatalf("test setup expected no BMC value")
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Diagram panicked instead of using Python State.update behavior: %v", r)
		}
	}()

	_, err := ui.Diagram()
	if err == nil || !strings.Contains(err.Error(), "no solver available") {
		t.Fatalf("expected Diagram to reach solver availability check, got %v", err)
	}
	if s1.Update == nil {
		t.Fatalf("Diagram did not compute the CTI successor update")
	}
}

// --- Weaken tests ---

func TestWeakenMultipleIndices(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	c1 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	c2 := goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil)
	c3 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	ui.Conjectures = []*goivy.Clauses{c1, c2, c3}

	removed, err := ui.Weaken([]int{0, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(removed) != 2 {
		t.Fatalf("expected 2 removed, got %d", len(removed))
	}
	if len(ui.Conjectures) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(ui.Conjectures))
	}
	if ui.Conjectures[0] != c2 {
		t.Error("remaining conjecture should be c2")
	}
	if ui.HaveCTI {
		t.Error("HaveCTI should be false after weaken")
	}
}

func TestWeakenOutOfBounds(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	c1 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	ui.Conjectures = []*goivy.Clauses{c1}

	removed, err := ui.Weaken([]int{5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(removed) != 0 {
		t.Errorf("expected 0 removed for out-of-bounds, got %d", len(removed))
	}
	if len(ui.Conjectures) != 1 {
		t.Errorf("conjectures should be unchanged, got %d", len(ui.Conjectures))
	}
}

func TestWeakenClearsHaveCTI(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	c1 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	ui.Conjectures = []*goivy.Clauses{c1}
	ui.HaveCTI = true

	ui.Weaken([]int{0})
	if ui.HaveCTI {
		t.Error("HaveCTI should be false after weaken")
	}
}

// --- SaveConjectures tests ---

func TestSaveConjecturesEmpty(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ui.Conjectures = nil
	result := ui.SaveConjectures()
	if !strings.Contains(result, "generated by ivy") {
		t.Error("should contain header")
	}
	if strings.Contains(result, "invariant") {
		t.Error("should not contain 'invariant' with no conjectures")
	}
}

func TestSaveConjecturesMultiple(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	c1 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	c2 := goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil)
	ui.Conjectures = []*goivy.Clauses{c1, c2}
	result := ui.SaveConjectures()
	count := strings.Count(result, "invariant")
	if count != 2 {
		t.Errorf("expected 2 invariant lines, got %d", count)
	}
}

// --- ShowUsedRelations tests ---

func TestShowUsedRelationsNilGraph(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ui.CurrentConceptGraph = nil
	ui.ShowUsedRelations(nil, false)
}

func TestShowUsedRelationsNilClauses(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	ui.ShowUsedRelations(nil, false)
}

// --- GatherFacts tests ---

func TestGatherFactsNilSession(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)
	w.CISess = nil
	w.GatherFacts()
	if len(w.ActiveFactExprs) != 0 {
		t.Errorf("expected 0 facts with nil session, got %d", len(w.ActiveFactExprs))
	}
}

// --- GetSelectedConjecture tests ---

func TestGetSelectedConjectureEmpty(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)
	w.ActiveFactExprs = nil
	conj := w.GetSelectedConjecture()
	if conj == nil {
		t.Fatal("expected empty selection to match Python's (~true) conjecture")
	}
	if len(conj.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(conj.Fmlas))
	}
	not, ok := conj.Fmlas[0].(*goivy.LogicNot)
	if !ok {
		t.Fatalf("empty selection formula = %T %[1]v, want negated true", conj.Fmlas[0])
	}
	if !goivy.IvyIsTrue(not.Body) {
		t.Fatalf("empty selection formula = %v, want ~true", conj.Fmlas[0])
	}
}

func TestGetSelectedConjectureBasic(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)

	// Set up simple ground facts (no free variables).
	a := goivy.NewConst("a", mkSort("S"))
	b := goivy.NewConst("b", mkSort("S"))
	eq, _ := goivy.NewEq(a, b)
	w.ActiveFactExprs = []goivy.Expr{eq}

	conj := w.GetSelectedConjecture()
	if conj == nil {
		t.Fatal("expected non-nil conjecture")
	}
	if len(conj.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(conj.Fmlas))
	}
}

func TestGetSelectedConjectureWithFreeVarsReturnsNil(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)

	// A fact with free variables should be rejected.
	x := mkVar("X", mkSort("S"))
	a := goivy.NewConst("a", mkSort("S"))
	eq, _ := goivy.NewEq(x, a)
	w.ActiveFactExprs = []goivy.Expr{eq}

	conj := w.GetSelectedConjecture()
	if conj != nil {
		t.Error("expected nil conjecture when facts have free variables")
	}
}

// --- Strengthen tests ---

func TestStrengthenEmptySelectionAddsPythonDefaultConjecture(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)
	w.ActiveFactExprs = nil

	before := len(ui.Conjectures)
	conj, err := w.Strengthen()
	if err != nil {
		t.Fatalf("unexpected error with Python empty-selection default: %v", err)
	}
	if conj == nil {
		t.Fatal("expected non-nil conjecture")
	}
	if len(ui.Conjectures) != before+1 {
		t.Errorf("expected %d conjectures, got %d", before+1, len(ui.Conjectures))
	}
}

func TestCTIStrengthenAddsConjecture(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)

	a := goivy.NewConst("a", mkSort("S"))
	b := goivy.NewConst("b", mkSort("S"))
	eq, _ := goivy.NewEq(a, b)
	w.ActiveFactExprs = []goivy.Expr{eq}

	before := len(ui.Conjectures)
	conj, err := w.Strengthen()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conj == nil {
		t.Fatal("expected non-nil conjecture")
	}
	if len(ui.Conjectures) != before+1 {
		t.Errorf("expected %d conjectures, got %d", before+1, len(ui.Conjectures))
	}
	if ui.HaveCTI {
		t.Error("HaveCTI should be false after strengthen")
	}
}

// --- MinimizeConjecture tests ---

func TestMinimizeConjectureNilModule(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)

	_, err := w.MinimizeConjecture(0)
	if err == nil {
		t.Error("expected error with nil module")
	}
}

// --- IsSufficient tests ---

func TestIsSufficientEmptySelectionUsesPythonDefaultConjecture(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)
	w.ActiveFactExprs = nil

	ok, msg := w.IsSufficient()
	if ok {
		t.Error("should not be sufficient with no current CTI target")
	}
	if !strings.Contains(msg, "no current CTI") {
		t.Errorf("expected 'no current CTI' in msg, got: %s", msg)
	}
}

func TestIsSufficientNoTarget(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)

	a := goivy.NewConst("a", mkSort("S"))
	b := goivy.NewConst("b", mkSort("S"))
	eq, _ := goivy.NewEq(a, b)
	w.ActiveFactExprs = []goivy.Expr{eq}

	ok, msg := w.IsSufficient()
	if ok {
		t.Error("should not be sufficient with no target")
	}
	if !strings.Contains(msg, "no current CTI") {
		t.Errorf("expected 'no current CTI' in msg, got: %s", msg)
	}
}

// --- IsInductive tests ---

func TestIsInductiveEmptySelectionUsesPythonDefaultConjecture(t *testing.T) {
	gs := NewGraphStack(nil)
	ui := NewCTIAnalysisGraphUI(nil)
	w := NewCTIConceptGraphWidget(gs, ui)
	w.ActiveFactExprs = nil

	ok, msg := w.IsInductive()
	if ok {
		t.Error("should not be inductive without a loaded module")
	}
	if !strings.Contains(msg, "no module") {
		t.Errorf("expected 'no module' in msg, got: %s", msg)
	}
}

// --- FormulaToConceptl tests ---

func TestFormulaToConceptlBasic(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	Y := mkVar("Y", S)
	r := mkConst("r", mkFuncSort(S, S, goivy.Boolean))
	fmla := mkApply(r, X, Y)

	concept := FormulaToConceptl(fmla)
	if concept == nil {
		t.Fatal("expected non-nil concept")
	}
	if concept.Arity() != 2 {
		t.Errorf("expected arity 2, got %d", concept.Arity())
	}
	if concept.Name == "" {
		t.Error("concept name should not be empty")
	}
	if !strings.Contains(concept.Name, "r") {
		t.Errorf("concept name should reference 'r', got: %s", concept.Name)
	}
}

func TestFormulaToConceptlUnary(t *testing.T) {
	S := mkSort("S")
	X := mkVar("X", S)
	p := mkConst("p", mkFuncSort(S, goivy.Boolean))
	fmla := mkApply(p, X)

	concept := FormulaToConceptl(fmla)
	if concept == nil {
		t.Fatal("expected non-nil concept")
	}
	if concept.Arity() != 1 {
		t.Errorf("expected arity 1, got %d", concept.Arity())
	}
}

func TestFormulaToConceptlGround(t *testing.T) {
	concept := FormulaToConceptl(goivy.True)
	if concept == nil {
		t.Fatal("expected non-nil concept")
	}
	if concept.Arity() != 0 {
		t.Errorf("expected arity 0 for ground formula, got %d", concept.Arity())
	}
}

// --- shouldFilterFact tests ---

func TestShouldFilterFactNotEqOrdered(t *testing.T) {
	a := goivy.NewConst("a", mkSort("S"))
	b := goivy.NewConst("b", mkSort("S"))

	// Not(a = b): should filter if "a" >= "b"
	eq, _ := goivy.NewEq(a, b)
	notEq, _ := goivy.NewNot(eq)

	result := shouldFilterFact(notEq)
	// "a" < "b", so a >= b is false, should NOT filter
	if result {
		t.Error("should not filter Not(a=b) since 'a' < 'b'")
	}

	// Not(b = a): "b" >= "a" is true, should filter
	eq2, _ := goivy.NewEq(b, a)
	notEq2, _ := goivy.NewNot(eq2)
	result2 := shouldFilterFact(notEq2)
	if !result2 {
		t.Error("should filter Not(b=a) since 'b' >= 'a'")
	}
}

func TestShouldFilterFactPositive(t *testing.T) {
	a := goivy.NewConst("a", mkSort("S"))
	b := goivy.NewConst("b", mkSort("S"))
	eq, _ := goivy.NewEq(a, b)

	if shouldFilterFact(eq) {
		t.Error("should not filter positive equalities")
	}
}

func TestShouldFilterFactTrue(t *testing.T) {
	if shouldFilterFact(goivy.True) {
		t.Error("should not filter True")
	}
}

// --- ctiWitness tests ---

func TestCtiWitness(t *testing.T) {
	witness := ctiWitness(nil)
	v := mkVar("V", mkSort("S"))
	result := witness(v)

	c, ok := result.(*goivy.Const)
	if !ok {
		t.Fatal("witness should return a Const")
	}
	if c.Name != "@V" {
		t.Errorf("expected '@V', got %q", c.Name)
	}
}

func TestCtiWitnessWithUsedNames(t *testing.T) {
	used := map[string]bool{"@V": true}
	witness := ctiWitness(used)
	v := mkVar("V", mkSort("S"))

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for name collision")
		}
	}()
	witness(v)
}

// --- Integration: Weaken then save ---

func TestWeakenThenSave(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	c1 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	c2 := goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil)
	c3 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	ui.Conjectures = []*goivy.Clauses{c1, c2, c3}

	ui.Weaken([]int{1})
	result := ui.SaveConjectures()
	count := strings.Count(result, "invariant")
	if count != 2 {
		t.Errorf("expected 2 invariant lines after weaken, got %d", count)
	}
}
