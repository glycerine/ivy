package goivy

import (
	"testing"
)

var tacticsTestAstCfg = NewAstConfig()

// testPC creates a minimal ProofChecker with a module.Config for testing.
func testPC() *ProofChecker {
	mod := New()
	mod.Cfg = NewConfig()
	return &ProofChecker{
		Cfg:    TacticNewConfig(),
		AstCfg: tacticsTestAstCfg,
		Mod:    mod,
	}
}

// ---------- helpers ----------

func tacticsMustVar(name string) *LogicVariable {
	v, err := NewVariable(name, Boolean)
	if err != nil {
		panic(err)
	}
	return v
}

func makeSimpleGoal(name string, fmla Expr) *LabeledFormula {
	label := tacticsTestAstCfg.NewAtom(name)
	return MakeGoal(tacticsTestAstCfg, Location{}, label, nil, fmla)
}

func makeTemporalGoal(name string, model Node, fmla Node) *LabeledFormula {
	tm := tacticsTestAstCfg.NewTemporalModels(model, fmla)
	label := tacticsTestAstCfg.NewAtom(name)
	lf := tacticsTestAstCfg.NewLabeledFormula(label, tm)
	lf.Temporal = BoolPtr(true)
	return lf
}

// ---------- Sorry ----------

func TestSorry(t *testing.T) {
	pc := testPC()
	pc.Mod.Cfg.UsedSorry = false
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Sorry(pc, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Sorry returned error: %v", err)
	}
	if !pc.Mod.Cfg.UsedSorry {
		t.Error("Expected UsedSorry to be true")
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 remaining goals, got %d", len(result))
	}
}

func TestSorryPreservesRest(t *testing.T) {
	pc := testPC()
	pc.Mod.Cfg.UsedSorry = false
	x := tacticsMustVar("X")
	g1 := makeSimpleGoal("g1", x)
	g2 := makeSimpleGoal("g2", x)
	g3 := makeSimpleGoal("g3", x)

	result, err := Sorry(pc, []*LabeledFormula{g1, g2, g3}, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Sorry returned error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 remaining goals, got %d", len(result))
	}
}

func TestSorryEmpty(t *testing.T) {
	result, err := Sorry(nil, nil, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Sorry returned error: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// ---------- Skolemize / Skolemizenp ----------

func TestSkolemize(t *testing.T) {
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Skolemize(nil, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Skolemize returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestSkolemizenp(t *testing.T) {
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Skolemizenp(nil, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Skolemizenp returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestSkolemizeEmpty(t *testing.T) {
	_, err := Skolemize(nil, nil, tacticsTestAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for empty decls")
	}
}

// ---------- TempindFmla ----------

func TestTempindFmlaDefault(t *testing.T) {
	x := tacticsMustVar("X")
	result := TempindFmla(x, nil, nil, nil)
	if result != x {
		t.Errorf("Expected identity when no vs, got %v", result)
	}
}

func TestTempindFmlaWithVs(t *testing.T) {
	x := tacticsMustVar("X")
	vs := []*LogicVariable{x}
	result := TempindFmla(x, nil, nil, vs)
	fa, ok := result.(*ForAll)
	if !ok {
		t.Fatalf("Expected ForAll, got %T: %v", result, result)
	}
	if len(fa.Variables) != 1 || fa.Variables[0].Name != "X" {
		t.Errorf("Expected ForAll over X, got %v", fa.Variables)
	}
}

func TestTempindFmlaForAll(t *testing.T) {
	x := tacticsMustVar("X")
	y := tacticsMustVar("Y")
	fa, _ := NewForAll([]*LogicVariable{y}, x)
	result := TempindFmla(fa, nil, nil, []*LogicVariable{x})
	resultFa, ok := result.(*ForAll)
	if !ok {
		t.Fatalf("Expected ForAll, got %T: %v", result, result)
	}
	if len(resultFa.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(resultFa.Variables))
	}
}

func TestTempindFmlaGlobally(t *testing.T) {
	x := tacticsMustVar("X")
	gb, _ := NewGlobally(nil, x)
	vs := []*LogicVariable{x}
	result := TempindFmla(gb, nil, nil, vs)
	fa, ok := result.(*ForAll)
	if !ok {
		t.Fatalf("Expected ForAll, got %T: %v", result, result)
	}
	if len(fa.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(fa.Variables))
	}
	or, ok := fa.Body.(*LogicOr)
	if !ok {
		t.Fatalf("Expected Or body, got %T: %v", fa.Body, fa.Body)
	}
	if len(or.Terms) != 2 {
		t.Errorf("Expected 2 terms in Or, got %d", len(or.Terms))
	}
	if _, ok := or.Terms[0].(*LogicGlobally); !ok {
		t.Errorf("Expected Globally as first Or term, got %T", or.Terms[0])
	}
	if _, ok := or.Terms[1].(*LogicWhenOperator); !ok {
		t.Errorf("Expected WhenOperator as second Or term, got %T", or.Terms[1])
	}
}

// ---------- TempcaseFmla ----------

func TestTempcaseFmlaDefault(t *testing.T) {
	x := tacticsMustVar("X")
	result, err := TempcaseFmla(x, nil, nil, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("TempcaseFmla returned error: %v", err)
	}
	if result != x {
		t.Errorf("Expected identity, got %v", result)
	}
}

func TestTempcaseFmlaCapture(t *testing.T) {
	x := tacticsMustVar("X")
	fa, _ := NewForAll([]*LogicVariable{x}, x)
	vs := []Expr{x}
	_, err := TempcaseFmla(fa, nil, vs, tacticsTestAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected variable capture error")
	}
}

func TestTempcaseFmlaGlobally(t *testing.T) {
	x := tacticsMustVar("X")
	gb, _ := NewGlobally(nil, x)
	vs := []Expr{x}
	result, err := TempcaseFmla(gb, x, vs, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("TempcaseFmla returned error: %v", err)
	}
	fa, ok := result.(*ForAll)
	if !ok {
		t.Fatalf("Expected ForAll, got %T: %v", result, result)
	}
	if _, ok := fa.Body.(*LogicGlobally); !ok {
		t.Errorf("Expected Globally body, got %T", fa.Body)
	}
}

// ---------- Vcgen ----------

func TestVcgenNonTemporalError(t *testing.T) {
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	_, err := Vcgen(nil, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for non-temporal goal")
	}
}

func TestVcgenNonTrueError(t *testing.T) {
	x := tacticsMustVar("X")
	np := &NormalProgram{}
	goal := makeTemporalGoal("g", np, x)

	_, err := Vcgen(nil, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for non-true temporal formula")
	}
}

func TestVcgenEmpty(t *testing.T) {
	_, err := Vcgen(nil, nil, tacticsTestAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for empty decls")
	}
}

// ---------- Tempind ----------

func TestTempindNonTemporalError(t *testing.T) {
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	tt := tacticsTestAstCfg.NewTacticTactic(tacticsTestAstCfg.NewAtom("tempind"), tacticsTestAstCfg.NewNoneAST(), nil)
	_, err := Tempind(nil, []*LabeledFormula{goal}, tt)
	if err == nil {
		t.Error("Expected error for non-temporal goal")
	}
}

func TestTempindWithTemporal(t *testing.T) {
	x := tacticsMustVar("X")
	gb, _ := NewGlobally(nil, x)
	np := &NormalProgram{}
	goal := makeTemporalGoal("g", np, gb)

	tt := tacticsTestAstCfg.NewTacticTactic(tacticsTestAstCfg.NewAtom("tempind"), tacticsTestAstCfg.NewNoneAST(), nil)
	result, err := Tempind(nil, []*LabeledFormula{goal}, tt)
	if err != nil {
		t.Fatalf("Tempind returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestTempindSchemaBodyTemporalModelsConclusion(t *testing.T) {
	x := tacticsMustVar("X")
	gb, _ := NewGlobally(nil, x)
	tm := tacticsTestAstCfg.NewTemporalModels(&NormalProgram{}, gb)
	prem := tacticsTestAstCfg.NewConstantDecl(NewConst("c", Boolean))
	goal := MakeGoal(tacticsTestAstCfg, Location{}, tacticsTestAstCfg.NewAtom("g"), []Node{prem}, tm)

	tt := tacticsTestAstCfg.NewTacticTactic(tacticsTestAstCfg.NewAtom("tempind"), tacticsTestAstCfg.NewNoneAST(), nil)
	result, err := Tempind(nil, []*LabeledFormula{goal}, tt)
	if err != nil {
		t.Fatalf("Tempind returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 goal, got %d", len(result))
	}
	if _, ok := GoalConc(result[0]).(*TemporalModels); !ok {
		t.Fatalf("expected SchemaBody conclusion to remain TemporalModels, got %T", GoalConc(result[0]))
	}
	if len(GoalPrems(result[0])) != 1 {
		t.Fatalf("expected original premise to remain, got %d", len(GoalPrems(result[0])))
	}
}

func TestCompileTacticLetsEmptyUsesTrueCondition(t *testing.T) {
	x := tacticsMustVar("X")
	gb, _ := NewGlobally(nil, x)
	goal := makeTemporalGoal("g", &NormalProgram{}, gb)
	tt := tacticsTestAstCfg.NewTacticTactic(tacticsTestAstCfg.NewAtom("tempind"), tacticsTestAstCfg.NewNoneAST(), nil)

	defs, cond, params, err := compileTacticLets(nil, tacticsTestAstCfg, goal, tt)
	if err != nil {
		t.Fatalf("compileTacticLets: %v", err)
	}
	if len(defs) != 0 {
		t.Fatalf("expected no defs, got %d", len(defs))
	}
	if !IsTrue(cond) {
		t.Fatalf("empty tactic lets should compile to true condition, got %T %v", cond, cond)
	}
	if len(params) != 0 {
		t.Fatalf("expected no params, got %d", len(params))
	}
}

// ---------- Registration ----------

func TestRegisterProofTactics(t *testing.T) {
	cfg := TacticNewConfig()
	RegisterProofTactics(cfg)

	expected := []string{"vcgen", "skolemize", "skolemizenp", "tempind", "tempcase", "sorry"}
	for _, name := range expected {
		if _, ok := cfg.Tactics[name]; !ok {
			t.Errorf("Expected tactic %q to be registered", name)
		}
	}
}

// ---------- VcToGoal ----------

func TestVcToGoal(t *testing.T) {
	x := tacticsMustVar("X")
	cls := NewClauses([]Expr{x}, nil, nil)
	goal := VcToGoal(tacticsTestAstCfg, Location{}, "test", cls, nil)
	if goal == nil {
		t.Fatal("VcToGoal returned nil")
	}
}

// ---------- pcAstCfg ----------

func TestPcAstCfg_NilChecker(t *testing.T) {
	// nil ProofChecker must not panic, must return a valid AstConfig.
	cfg := pcAstCfg(nil)
	if cfg == nil {
		t.Fatal("pcAstCfg(nil) returned nil, expected fallback AstConfig")
	}
}

func TestPcAstCfg_CheckerWithNilAstCfg(t *testing.T) {
	// ProofChecker whose GetAstCfg() returns nil must still produce a valid config.
	pc := &ProofChecker{
		Cfg:    TacticNewConfig(),
		AstCfg: nil, // deliberately nil
		Mod:    New(),
	}
	cfg := pcAstCfg(pc)
	if cfg == nil {
		t.Fatal("pcAstCfg returned nil for checker with nil AstCfg")
	}
}

func TestPcAstCfg_CheckerWithAstCfg(t *testing.T) {
	// When the checker has a real AstConfig, pcAstCfg must return it —
	// NOT a fresh default. This is the bug the infinite recursion masked:
	// callers expect to get the session's AstConfig, not a throwaway one.
	pc := testPC()
	cfg := pcAstCfg(pc)
	if cfg == nil {
		t.Fatal("pcAstCfg returned nil")
	}
	if cfg != pc.AstCfg {
		t.Error("pcAstCfg returned a different AstConfig than the checker's — " +
			"should return the checker's config, not a fresh default")
	}
}

// ---------- Tactic functions with non-nil ProofChecker ----------

func TestSorryWithProofChecker(t *testing.T) {
	pc := testPC()
	pc.Mod.Cfg.UsedSorry = false
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Sorry(pc, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Sorry returned error: %v", err)
	}
	if !pc.Mod.Cfg.UsedSorry {
		t.Error("Expected UsedSorry to be true")
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 remaining goals, got %d", len(result))
	}
}

func TestSkolemizeWithProofChecker(t *testing.T) {
	pc := testPC()
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Skolemize(pc, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Skolemize returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestSkolemizenp_WithProofChecker(t *testing.T) {
	pc := testPC()
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Skolemizenp(pc, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Skolemizenp returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestSkolemizenpSchemaBodyConstantPremisePythonError(t *testing.T) {
	tm := tacticsTestAstCfg.NewTemporalModels(&NormalProgram{}, True)
	prem := tacticsTestAstCfg.NewConstantDecl(NewConst("c", Boolean))
	goal := MakeGoal(tacticsTestAstCfg, Location{}, tacticsTestAstCfg.NewAtom("g"), []Node{prem}, tm)

	_, err := Skolemizenp(nil, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err == nil {
		t.Fatal("expected Python-compatible ConstantDecl label error")
	}
	want := "'ConstantDecl' object has no attribute 'label'"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestTempindWithProofChecker(t *testing.T) {
	pc := testPC()
	x := tacticsMustVar("X")
	gb, _ := NewGlobally(nil, x)
	np := &NormalProgram{}
	goal := makeTemporalGoal("g", np, gb)

	tt := tacticsTestAstCfg.NewTacticTactic(tacticsTestAstCfg.NewAtom("tempind"), tacticsTestAstCfg.NewNoneAST(), nil)
	result, err := Tempind(pc, []*LabeledFormula{goal}, tt)
	if err != nil {
		t.Fatalf("Tempind returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestVcgenWithProofChecker(t *testing.T) {
	pc := testPC()
	x := tacticsMustVar("X")
	goal := makeSimpleGoal("g", x)

	// Vcgen requires a temporal goal; non-temporal should error even with a real checker.
	_, err := Vcgen(pc, []*LabeledFormula{goal}, tacticsTestAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for non-temporal goal")
	}
}

// --- Tests for ImpliedFacts and RefutedGoal (PLAN219) ---

func TestImpliedFactsNilPremise(t *testing.T) {
	tc := NewTacticsContext(nil, New())
	result := tc.ImpliedFacts(nil, []*Clauses{TrueClauses(nil)})
	if result != nil {
		t.Error("nil premise should return nil")
	}
}

func TestImpliedFactsEmptyFacts(t *testing.T) {
	tc := NewTacticsContext(nil, New())
	p := NewConst("p", Boolean)
	premClauses := NewClauses([]Expr{p}, nil, nil)
	result := tc.ImpliedFacts(premClauses, nil)
	if result != nil {
		t.Error("empty factsToCheck should return nil")
	}
}

func TestImpliedFactsBasic(t *testing.T) {
	tc := NewTacticsContext(nil, New())

	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	r := NewConst("r", Boolean)

	// premise: p AND q
	premClauses := NewClauses([]Expr{p, q}, nil, nil)

	// facts: {p}, {q}, {r}
	factP := NewClauses([]Expr{p}, nil, nil)
	factQ := NewClauses([]Expr{q}, nil, nil)
	factR := NewClauses([]Expr{r}, nil, nil)

	implied := tc.ImpliedFacts(premClauses, []*Clauses{factP, factQ, factR})
	// Should find p and q implied, not r
	if len(implied) != 2 {
		t.Fatalf("expected 2 implied facts, got %d", len(implied))
	}
}

func TestRefutedGoalNilGoal(t *testing.T) {
	tc := NewTacticsContext(nil, New())
	if tc.RefutedGoal(nil) {
		t.Error("nil goal should not be refuted")
	}
}

func TestRefutedGoalFalseFormula(t *testing.T) {
	tc := NewTacticsContext(nil, New())
	// Empty Or is False
	goal := &ProofGoal{Formula: &LogicOr{Terms: []Expr{}}}
	if !tc.RefutedGoal(goal) {
		t.Error("goal with False formula should be refuted")
	}
}

// TestUPDR_InductiveWithInitializer is in updr_tactics_test.go
// (external test package to avoid import cycle with check/).
