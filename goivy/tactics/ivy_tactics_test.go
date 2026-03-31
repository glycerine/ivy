package tactics

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/clauseops"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/temporal"
)

var testAstCfg = ast.NewAstConfig()

// testPC creates a minimal ProofChecker with a module.Config for testing.
func testPC() *proof.ProofChecker {
	mod := module.New()
	mod.Cfg = module.NewConfig()
	return &proof.ProofChecker{
		Cfg:    module.TacticNewConfig(),
		AstCfg: testAstCfg,
		Mod:    mod,
	}
}

// ---------- helpers ----------

func mustVar(name string) *lg.Variable {
	v, err := lg.NewVariable(name, lg.Boolean)
	if err != nil {
		panic(err)
	}
	return v
}

func makeSimpleGoal(name string, fmla lg.Expr) *ast.LabeledFormula {
	label := testAstCfg.NewAtom(name)
	return proof.MakeGoal(testAstCfg, ast.Location{}, label, nil, fmla)
}

func makeTemporalGoal(name string, model ast.Node, fmla ast.Node) *ast.LabeledFormula {
	tm := testAstCfg.NewTemporalModels(model, fmla)
	label := testAstCfg.NewAtom(name)
	lf := testAstCfg.NewLabeledFormula(label, tm)
	lf.Temporal = ast.BoolPtr(true)
	return lf
}

// ---------- Sorry ----------

func TestSorry(t *testing.T) {
	pc := testPC()
	pc.Mod.Cfg.UsedSorry = false
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Sorry(pc, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
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
	x := mustVar("X")
	g1 := makeSimpleGoal("g1", x)
	g2 := makeSimpleGoal("g2", x)
	g3 := makeSimpleGoal("g3", x)

	result, err := Sorry(pc, []*ast.LabeledFormula{g1, g2, g3}, testAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Sorry returned error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 remaining goals, got %d", len(result))
	}
}

func TestSorryEmpty(t *testing.T) {
	result, err := Sorry(nil, nil, testAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Sorry returned error: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// ---------- Skolemize / Skolemizenp ----------

func TestSkolemize(t *testing.T) {
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Skolemize(nil, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Skolemize returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestSkolemizenp(t *testing.T) {
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Skolemizenp(nil, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Skolemizenp returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestSkolemizeEmpty(t *testing.T) {
	_, err := Skolemize(nil, nil, testAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for empty decls")
	}
}

// ---------- TempindFmla ----------

func TestTempindFmlaDefault(t *testing.T) {
	x := mustVar("X")
	result := TempindFmla(x, nil, nil, nil)
	if result != x {
		t.Errorf("Expected identity when no vs, got %v", result)
	}
}

func TestTempindFmlaWithVs(t *testing.T) {
	x := mustVar("X")
	vs := []*lg.Variable{x}
	result := TempindFmla(x, nil, nil, vs)
	fa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("Expected ForAll, got %T: %v", result, result)
	}
	if len(fa.Variables) != 1 || fa.Variables[0].Name != "X" {
		t.Errorf("Expected ForAll over X, got %v", fa.Variables)
	}
}

func TestTempindFmlaForAll(t *testing.T) {
	x := mustVar("X")
	y := mustVar("Y")
	fa, _ := lg.NewForAll([]*lg.Variable{y}, x)
	result := TempindFmla(fa, nil, nil, []*lg.Variable{x})
	resultFa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("Expected ForAll, got %T: %v", result, result)
	}
	if len(resultFa.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(resultFa.Variables))
	}
}

func TestTempindFmlaGlobally(t *testing.T) {
	x := mustVar("X")
	gb, _ := lg.NewGlobally(nil, x)
	vs := []*lg.Variable{x}
	result := TempindFmla(gb, nil, nil, vs)
	fa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("Expected ForAll, got %T: %v", result, result)
	}
	if len(fa.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(fa.Variables))
	}
	or, ok := fa.Body.(*lg.Or)
	if !ok {
		t.Fatalf("Expected Or body, got %T: %v", fa.Body, fa.Body)
	}
	if len(or.Terms) != 2 {
		t.Errorf("Expected 2 terms in Or, got %d", len(or.Terms))
	}
	if _, ok := or.Terms[0].(*lg.Globally); !ok {
		t.Errorf("Expected Globally as first Or term, got %T", or.Terms[0])
	}
	if _, ok := or.Terms[1].(*lg.WhenOperator); !ok {
		t.Errorf("Expected WhenOperator as second Or term, got %T", or.Terms[1])
	}
}

// ---------- TempcaseFmla ----------

func TestTempcaseFmlaDefault(t *testing.T) {
	x := mustVar("X")
	result, err := TempcaseFmla(x, nil, nil, testAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("TempcaseFmla returned error: %v", err)
	}
	if result != x {
		t.Errorf("Expected identity, got %v", result)
	}
}

func TestTempcaseFmlaCapture(t *testing.T) {
	x := mustVar("X")
	fa, _ := lg.NewForAll([]*lg.Variable{x}, x)
	vs := []lg.Expr{x}
	_, err := TempcaseFmla(fa, nil, vs, testAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected variable capture error")
	}
}

func TestTempcaseFmlaGlobally(t *testing.T) {
	x := mustVar("X")
	gb, _ := lg.NewGlobally(nil, x)
	vs := []lg.Expr{x}
	result, err := TempcaseFmla(gb, x, vs, testAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("TempcaseFmla returned error: %v", err)
	}
	fa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("Expected ForAll, got %T: %v", result, result)
	}
	if _, ok := fa.Body.(*lg.Globally); !ok {
		t.Errorf("Expected Globally body, got %T", fa.Body)
	}
}

// ---------- Vcgen ----------

func TestVcgenNonTemporalError(t *testing.T) {
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	_, err := Vcgen(nil, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for non-temporal goal")
	}
}

func TestVcgenNonTrueError(t *testing.T) {
	x := mustVar("X")
	np := &temporal.NormalProgram{}
	goal := makeTemporalGoal("g", np, x)

	_, err := Vcgen(nil, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for non-true temporal formula")
	}
}

func TestVcgenEmpty(t *testing.T) {
	_, err := Vcgen(nil, nil, testAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for empty decls")
	}
}

// ---------- Tempind ----------

func TestTempindNonTemporalError(t *testing.T) {
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	tt := testAstCfg.NewTacticTactic(testAstCfg.NewAtom("tempind"), testAstCfg.NewNoneAST(), nil)
	_, err := Tempind(nil, []*ast.LabeledFormula{goal}, tt)
	if err == nil {
		t.Error("Expected error for non-temporal goal")
	}
}

func TestTempindWithTemporal(t *testing.T) {
	x := mustVar("X")
	gb, _ := lg.NewGlobally(nil, x)
	np := &temporal.NormalProgram{}
	goal := makeTemporalGoal("g", np, gb)

	tt := testAstCfg.NewTacticTactic(testAstCfg.NewAtom("tempind"), testAstCfg.NewNoneAST(), nil)
	result, err := Tempind(nil, []*ast.LabeledFormula{goal}, tt)
	if err != nil {
		t.Fatalf("Tempind returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

// ---------- Registration ----------

func TestRegisterProofTactics(t *testing.T) {
	cfg := module.TacticNewConfig()
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
	x := mustVar("X")
	cls := clauseops.NewClauses([]lg.Expr{x}, nil, nil)
	goal := VcToGoal(testAstCfg, ast.Location{}, "test", cls, nil)
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
	pc := &proof.ProofChecker{
		Cfg:    module.TacticNewConfig(),
		AstCfg: nil, // deliberately nil
		Mod:    module.New(),
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
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Sorry(pc, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
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
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Skolemize(pc, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Skolemize returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestSkolemizenp_WithProofChecker(t *testing.T) {
	pc := testPC()
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	result, err := Skolemizenp(pc, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
	if err != nil {
		t.Fatalf("Skolemizenp returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestTempindWithProofChecker(t *testing.T) {
	pc := testPC()
	x := mustVar("X")
	gb, _ := lg.NewGlobally(nil, x)
	np := &temporal.NormalProgram{}
	goal := makeTemporalGoal("g", np, gb)

	tt := testAstCfg.NewTacticTactic(testAstCfg.NewAtom("tempind"), testAstCfg.NewNoneAST(), nil)
	result, err := Tempind(pc, []*ast.LabeledFormula{goal}, tt)
	if err != nil {
		t.Fatalf("Tempind returned error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result))
	}
}

func TestVcgenWithProofChecker(t *testing.T) {
	pc := testPC()
	x := mustVar("X")
	goal := makeSimpleGoal("g", x)

	// Vcgen requires a temporal goal; non-temporal should error even with a real checker.
	_, err := Vcgen(pc, []*ast.LabeledFormula{goal}, testAstCfg.NewNoneAST())
	if err == nil {
		t.Error("Expected error for non-temporal goal")
	}
}
