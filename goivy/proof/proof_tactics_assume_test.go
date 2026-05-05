package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// --- helpers for assumeTactic tests ---

// mkPC creates a ProofChecker with the given schemas registered.
func mkPC(schemas ...*ast.LabeledFormula) *ProofChecker {
	pc := NewProofChecker(nil, nil, nil, nil, nil)
	for _, s := range schemas {
		name := s.LabelName()
		pc.Schemata.Set(name, s)
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
	if _, ok := conc.(*ast.AstTemporalModels); !ok {
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
	if _, ok := GoalConc(skolemized).(*ast.AstTemporalModels); !ok {
		t.Fatalf("after skolemize: expected TemporalModels conclusion, got %T", GoalConc(skolemized))
	}

	//        // Verify TemporalModels survived skolemization
	//        t.Logf("after skolemize: Formula type=%T", skolemized.Formula)
	//        concAfterSkolem := GoalConc(skolemized)
	//        t.Logf("after skolemize: GoalConc type=%T", concAfterSkolem)
	//        premsAfterSkolem := GoalPrems(skolemized)
	//        t.Logf("after skolemize: nPrems=%d", len(premsAfterSkolem))
	//        for i, p := range premsAfterSkolem {
	//                t.Logf("  prem[%d] type=%T", i, p)
	//        }
	//        if _, ok := concAfterSkolem.(*ast.TemporalModels); !ok {
	//                t.Fatalf("after skolemize: expected TemporalModels conclusion, got %T", concAfterSkol
	//        }

	// Step 2: instantiate fair_ax
	proof := mkAssumeTactic("fair_ax")
	result, err := pc.assumeTactic([]*ast.LabeledFormula{skolemized}, proof, false)
	if err != nil {
		t.Fatalf("assumeTactic after skolemize failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	//        // Diagnose the result
	//        t.Logf("after assume: Formula type=%T", result[0].Formula)
	//        concFinal := GoalConc(result[0])
	//        t.Logf("after assume: GoalConc type=%T", concFinal)
	//        premsFinal := GoalPrems(result[0])
	//        t.Logf("after assume: nPrems=%d", len(premsFinal))
	//        for i, p := range premsFinal {
	//                t.Logf("  prem[%d] type=%T label=%v", i, p, p)
	//        }

	// The conclusion must STILL be TemporalModels
	concFinal := GoalConc(result[0])
	if _, ok := concFinal.(*ast.AstTemporalModels); !ok {
		t.Fatalf("REGRESSION: after skolemize+instantiate, conclusion is %T, not *ast.TemporalModels", concFinal)
	}
}

// TestIsWitVar_BoundVariableKey verifies that isWitVar correctly identifies a
// match key as a witness-eligible Variable when the key is a bound variable
// of the schema's conclusion AND is not in prob.FreeSyms.
//
// Python: iswit(x) = isinstance(x, il.Variable) and x not in prob.freesyms
// where x is the match KEY. The prior Go port checked val.(*lg.Variable)
// which broke for `instantiate ... with P=_P` when _P is a skolem Const.
// See plan: /Users/jaten/.claude/plans/the-latest-divergence-of-synchronous-creek.md
func TestIsWitVar_BoundVariableKey(t *testing.T) {
	procSort := mkSort("proc")
	p := mkVar("P", procSort)
	underP := mkConst("_P", procSort)

	// Schema conc: forall P:proc. true  (stand-in for the axiom body)
	body := lg.True
	fa, _ := lg.NewForAll([]*lg.Variable{p}, body)
	schema := mkLF(testAstCfg.NewAtom("ifabric_rw_fair_ax"), fa)

	// prob.FreeSyms excludes P (bound), matching Python goal_vocab semantics.
	prob := NewMatchProblem(nil, fa, fa, map[lg.NodeKey]lg.Expr{}, nil)
	prob.SchemaLF = schema

	// Match {P: _P} — key P is a Variable, val _P is a Const.
	pKey := lg.Key(p)

	if !isWitVar(pKey, underP, prob) {
		t.Error("isWitVar should return true for bound-variable key P with Const witness _P")
	}

	// Sanity: a symbol-LHS key that IS in prob.FreeSyms must NOT be a witness.
	sym := mkConst("f", procSort)
	probFS := NewMatchProblem(nil, fa, fa, map[lg.NodeKey]lg.Expr{lg.Key(sym): sym}, nil)
	probFS.SchemaLF = schema
	if isWitVar(lg.Key(sym), underP, probFS) {
		t.Error("isWitVar should return false for symbol-LHS key already in FreeSyms")
	}
}

// TestIsWitVar_DropsForAllViaWitnessAst verifies the end-to-end interaction
// between the fixed isWitVar and module.WitnessAst: when isWitVar correctly
// identifies a bound-Variable key, WitnessAst is called with it in the witness
// map and drops the vacuous ForAll.
//
// Mirrors the `instantiate ifabric_rw_fair_ax with P=_P` scenario from
// ord_live.ivy that caused the neg_prop_init divergence at log.golden.2hr
// index 2411549.
func TestIsWitVar_DropsForAllViaWitnessAst(t *testing.T) {
	procSort := mkSort("proc")
	p := mkVar("P", procSort)
	underP := mkConst("_P", procSort)

	// Build `rd_fair(P) & wr_fair(P)` body.
	boolFs, _ := lg.NewFunctionSort(procSort, lg.Boolean)
	rdFair := mkConst("rd_fair", boolFs)
	wrFair := mkConst("wr_fair", boolFs)
	rdFairP, _ := lg.NewApply(rdFair, p)
	wrFairP, _ := lg.NewApply(wrFair, p)
	body := &lg.And{Terms: []lg.Expr{rdFairP, wrFairP}}

	// Schema conc: forall P:proc. (rd_fair(P) & wr_fair(P))
	fa, _ := lg.NewForAll([]*lg.Variable{p}, body)
	schema := mkLF(testAstCfg.NewAtom("ifabric_rw_fair_ax"), fa)

	// prob with P excluded from FreeSyms (bound variable semantics).
	prob := NewMatchProblem(nil, fa, fa, map[lg.NodeKey]lg.Expr{}, nil)
	prob.SchemaLF = schema

	// Simulate the pmatch split manually: pmatch = {P: _P}.
	pmatch := map[lg.NodeKey]lg.Expr{lg.Key(p): underP}
	witness := make(map[lg.NodeKey]lg.Expr)
	pmatchClean := make(map[lg.NodeKey]lg.Expr)
	for k, v := range pmatch {
		if isWitVar(k, v, prob) {
			witness[k] = v
		} else {
			pmatchClean[k] = v
		}
	}
	if len(witness) != 1 {
		t.Fatalf("expected 1 witness entry, got %d", len(witness))
	}
	if len(pmatchClean) != 0 {
		t.Fatalf("expected 0 pmatchClean entries, got %d", len(pmatchClean))
	}

	// Run WitnessAst on the schema conclusion; the ForAll must be dropped.
	newConc, err := module.WitnessAst(true, nil, witness, fa)
	if err != nil {
		t.Fatalf("WitnessAst failed: %v", err)
	}
	if _, isFA := newConc.(*lg.ForAll); isFA {
		t.Errorf("REGRESSION: WitnessAst did not drop vacuous ForAll; got %v", newConc.Canon())
	}
	// The result should be an And (the substituted body), not a ForAll.
	if _, isAnd := newConc.(*lg.And); !isAnd {
		t.Errorf("expected result to be *lg.And (body), got %T: %v", newConc, newConc.Canon())
	}
}
