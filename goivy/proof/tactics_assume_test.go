package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// --- helpers for assumeTactic tests ---

// mkPC creates a ProofChecker with the given schemas registered.
func mkPC(schemas ...*ast.LabeledFormula) *ProofChecker {
	pc := NewProofChecker(nil, nil, nil, nil, nil)
	for _, s := range schemas {
		name := s.LabelName()
		pc.Schemata[name] = s
	}
	return pc
}

// mkAssumeTactic creates an AssumeTactic for "instantiate <name>".
func mkAssumeTactic(name string) *ast.AssumeTactic {
	at := testAstCfg.NewAssumeTactic(testAstCfg.NewAtom(name), nil)
	at.TLabel = testAstCfg.NewNoneAST()
	return at
}

// --- Test 1: Basic instantiate adds premise ---

func TestAssumeTactic_BasicInstantiate(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	// Goal: [goal] c
	goal := mkLF(testAstCfg.NewAtom("goal"), c)

	// Schema: [myax] c (an axiom to instantiate)
	schema := mkLF(testAstCfg.NewAtom("myax"), c)
	pc := mkPC(schema)

	proof := mkAssumeTactic("myax")
	result, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result goal, got %d", len(result))
	}
	// The result should have "myax" as a premise
	prems := GoalPrems(result[0])
	if len(prems) == 0 {
		t.Fatal("expected at least 1 premise after instantiate")
	}
	foundMyax := false
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok && lf.LabelName() == "myax" {
			foundMyax = true
		}
	}
	if !foundMyax {
		t.Error("expected premise 'myax' in result")
	}
}

// --- Test 2: Preserves TemporalModels conclusion ---

func TestAssumeTactic_PreservesTemporalModels(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	// Create a TemporalModels conclusion
	np := testAstCfg.NewNoneAST() // dummy model node (avoids temporal import cycle)
	tm := testAstCfg.NewTemporalModels(np, lg.True)

	// Goal: [goal] TemporalModels(np, true)
	goal := mkLF(testAstCfg.NewAtom("goal"), tm)

	// Schema: [myax] c
	schema := mkLF(testAstCfg.NewAtom("myax"), c)
	pc := mkPC(schema)

	proof := mkAssumeTactic("myax")
	result, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result goal, got %d", len(result))
	}

	// The conclusion must still be *ast.TemporalModels
	conc := GoalConc(result[0])
	if _, ok := conc.(*ast.TemporalModels); !ok {
		t.Fatalf("expected conclusion to be *ast.TemporalModels, got %T", conc)
	}
}

// --- Test 3: Works with existing SchemaBody premises ---

func TestAssumeTactic_WithSchemaBody(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	// Goal already has a premise (SchemaBody)
	existingPrem := mkLF(testAstCfg.NewAtom("existing"), c)
	sb := testAstCfg.NewSchemaBody(existingPrem, c)
	goal := mkLF(testAstCfg.NewAtom("goal"), sb)

	// Schema: [newax] c
	schema := mkLF(testAstCfg.NewAtom("newax"), c)
	pc := mkPC(schema)

	proof := mkAssumeTactic("newax")
	result, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic failed: %v", err)
	}
	prems := GoalPrems(result[0])
	if len(prems) != 2 {
		t.Fatalf("expected 2 premises (existing + new), got %d", len(prems))
	}
}

// --- Test 4: RemoveExplicit clears explicit flag ---

func TestAssumeTactic_RemoveExplicit(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	goal := mkLF(testAstCfg.NewAtom("goal"), c)

	// Schema with explicit=true
	schema := mkLF(testAstCfg.NewAtom("myax"), c)
	schema.Explicit = true
	pc := mkPC(schema)

	proof := mkAssumeTactic("myax")
	result, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic failed: %v", err)
	}
	// The premise should have explicit=false
	for _, p := range GoalPrems(result[0]) {
		if lf, ok := p.(*ast.LabeledFormula); ok && lf.LabelName() == "myax" {
			if lf.Explicit {
				t.Error("expected explicit=false on premise after RemoveExplicit")
			}
		}
	}
}

// --- Test 5: Premise from goal (AssumeTactic, not global) ---

func TestAssumeTactic_PremiseFromGoal(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	// Goal has "myax" as an existing premise
	premInGoal := mkLF(testAstCfg.NewAtom("myax"), c)
	sb := testAstCfg.NewSchemaBody(premInGoal, c)
	goal := mkLF(testAstCfg.NewAtom("goal"), sb)

	pc := mkPC() // no global schemas

	// instantiate myax (which exists in premises)
	proof := mkAssumeTactic("myax")
	result, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic failed: %v", err)
	}
	// When label is NoneAST, the original premise is removed and
	// a fresh copy is added. The net effect may be the same premise
	// re-added. Just verify we get a result without error.
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
}

// --- Test 6: Clash error for AssumeTactic ---

func TestAssumeTactic_ClashError(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	// Goal has "myax" as a premise
	premInGoal := mkLF(testAstCfg.NewAtom("myax"), c)
	sb := testAstCfg.NewSchemaBody(premInGoal, c)
	goal := mkLF(testAstCfg.NewAtom("goal"), sb)

	// Schema also named "myax" in global context
	schema := mkLF(testAstCfg.NewAtom("myax"), c)
	pc := mkPC(schema)

	// Use explicit label "myax" to force clash
	proof := testAstCfg.NewAssumeTactic(testAstCfg.NewAtom("myax"), nil)
	proof.TLabel = testAstCfg.NewAtom("myax") // explicit label, not NoneAST

	// isGlobal=false → should error on clash
	_, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, false)
	if err == nil {
		t.Fatal("expected clash error for AssumeTactic")
	}
	if _, ok := err.(*ProofError); !ok {
		t.Fatalf("expected *ProofError, got %T: %v", err, err)
	}
}

// --- Test 7: AssumeGlobalTactic renames on clash ---

func TestAssumeTactic_GlobalRename(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	// Goal has "myax" as a premise
	premInGoal := mkLF(testAstCfg.NewAtom("myax"), c)
	sb := testAstCfg.NewSchemaBody(premInGoal, c)
	goal := mkLF(testAstCfg.NewAtom("goal"), sb)

	// Schema also named "myax" in global context
	schema := mkLF(testAstCfg.NewAtom("myax"), c)
	pc := mkPC(schema)

	proof := mkAssumeTactic("myax")

	// isGlobal=true → should rename instead of erroring
	result, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, true)
	if err != nil {
		t.Fatalf("expected no error for global assume with rename, got: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
}

// --- Test 8: NoneAST label preserves schema label ---

func TestAssumeTactic_NoneASTLabel(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	goal := mkLF(testAstCfg.NewAtom("goal"), c)
	schema := mkLF(testAstCfg.NewAtom("origname"), c)
	pc := mkPC(schema)

	proof := testAstCfg.NewAssumeTactic(testAstCfg.NewAtom("origname"), nil)
	proof.TLabel = testAstCfg.NewNoneAST() // NoneAST → keep schema label

	result, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic failed: %v", err)
	}
	// Premise should keep the schema's label "origname"
	for _, p := range GoalPrems(result[0]) {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.LabelName() != "origname" {
				t.Errorf("expected label 'origname', got '%s'", lf.LabelName())
			}
		}
	}
}

// --- Test 9: Explicit label replaces schema label ---

func TestAssumeTactic_ExplicitLabel(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)

	goal := mkLF(testAstCfg.NewAtom("goal"), c)
	schema := mkLF(testAstCfg.NewAtom("origname"), c)
	pc := mkPC(schema)

	proof := testAstCfg.NewAssumeTactic(testAstCfg.NewAtom("origname"), nil)
	proof.TLabel = testAstCfg.NewAtom("newlabel") // explicit label

	result, err := pc.assumeTactic([]*ast.LabeledFormula{goal}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic failed: %v", err)
	}
	for _, p := range GoalPrems(result[0]) {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.LabelName() != "newlabel" {
				t.Errorf("expected label 'newlabel', got '%s'", lf.LabelName())
			}
		}
	}
}

// --- Test 10: Composed skolemize + instantiate preserves TemporalModels ---
// This is the exact pattern that triggered the original bug.

func TestAssumeTactic_ComposedWithSkolemizePreservesTemporalModels(t *testing.T) {
	s := mkSort("T")
	v := mkVar("X", s)

	// Create a temporal formula: ForAll X. (globally true)
	globally := &lg.Globally{Body: lg.True}
	inner := &lg.ForAll{Variables: []*lg.Variable{v}, Body: globally}

	// Wrap in TemporalModels
	np := testAstCfg.NewNoneAST() // dummy model node (avoids temporal import cycle)
	tm := testAstCfg.NewTemporalModels(np, inner)

	// Goal: [prop] TemporalModels(np, forall X. globally true)
	goal := mkLF(testAstCfg.NewAtom("prop"), tm)

	// Schema (axiom to instantiate): [fair_ax] (globally true)
	fairAx := mkLF(testAstCfg.NewAtom("fair_ax"), globally)

	// Create proof checker with the axiom
	pc := mkPC(fairAx)

	// Step 1: Skolemize
	skolemized := SkolemizeGoal(testAstCfg, goal, true)

	// Verify TemporalModels survived skolemization
	concAfterSkolem := GoalConc(skolemized)
	if _, ok := concAfterSkolem.(*ast.TemporalModels); !ok {
		t.Fatalf("after skolemize: expected TemporalModels conclusion, got %T", concAfterSkolem)
	}

	// Step 2: instantiate fair_ax
	proof := mkAssumeTactic("fair_ax")
	result, err := pc.assumeTactic([]*ast.LabeledFormula{skolemized}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic after skolemize failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	// The conclusion must STILL be TemporalModels
	concFinal := GoalConc(result[0])
	if _, ok := concFinal.(*ast.TemporalModels); !ok {
		t.Fatalf("REGRESSION: after skolemize+instantiate, conclusion is %T, not *ast.TemporalModels", concFinal)
	}
}
