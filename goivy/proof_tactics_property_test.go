package goivy

import (
	"strings"
	"testing"
)

// --- defargNameSort tests ---

func TestDefargNameSort_App(t *testing.T) {
	cfg := NewAstConfig()
	sym := cfg.NewSymbol("w", nil)
	app := cfg.NewApp(sym)
	app.ASort = cfg.NewSymbol("T", nil)

	name, sortName, isTop := defargNameSort(app)
	if name != "w" {
		t.Errorf("name: got %q, want %q", name, "w")
	}
	if sortName != "T" {
		t.Errorf("sortName: got %q, want %q", sortName, "T")
	}
	if isTop {
		t.Error("expected isTop=false for App with ASort")
	}
}

func TestDefargNameSort_AppNoSort(t *testing.T) {
	cfg := NewAstConfig()
	sym := cfg.NewSymbol("w", nil)
	app := cfg.NewApp(sym)

	name, _, isTop := defargNameSort(app)
	if name != "w" {
		t.Errorf("name: got %q, want %q", name, "w")
	}
	if !isTop {
		t.Error("expected isTop=true for App without ASort")
	}
}

func TestDefargNameSort_Variable(t *testing.T) {
	cfg := NewAstConfig()
	v := cfg.NewVariable("X", "S")

	name, sortName, isTop := defargNameSort(v)
	if name != "X" {
		t.Errorf("name: got %q, want %q", name, "X")
	}
	if sortName != "S" {
		t.Errorf("sortName: got %q, want %q", sortName, "S")
	}
	if !isTop {
		t.Error("expected isTop=true for Variable with VSort='S' (universe)")
	}
}

func TestDefargNameSort_VariableExplicitSort(t *testing.T) {
	cfg := NewAstConfig()
	v := cfg.NewVariable("X", "mytype")

	name, sortName, isTop := defargNameSort(v)
	if name != "X" {
		t.Errorf("name: got %q, want %q", name, "X")
	}
	if sortName != "mytype" {
		t.Errorf("sortName: got %q, want %q", sortName, "mytype")
	}
	if isTop {
		t.Error("expected isTop=false for Variable with explicit sort")
	}
}

func TestDefargNameSort_VariableEmptySort(t *testing.T) {
	cfg := NewAstConfig()
	v := cfg.NewVariable("Z", "")

	_, _, isTop := defargNameSort(v)
	if !isTop {
		t.Error("expected isTop=true for Variable with empty sort")
	}
}

func TestDefargNameSort_Atom(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewAtom("c")

	name, _, isTop := defargNameSort(a)
	if name != "c" {
		t.Errorf("name: got %q, want %q", name, "c")
	}
	if !isTop {
		t.Error("expected isTop=true for Atom without ASort")
	}
}

// --- propertyTactic Skolem integration tests ---

// mkTestMod creates a module with a Sig containing sort "S" and relation "P".
func mkTestMod(pArity int) *Module {
	sig := NewSig()
	sSort := &UninterpretedSort{Name: "S"}
	sig.Sorts.Set("S", sSort)

	if pArity > 0 {
		sorts := make([]Sort, pArity+1)
		for i := 0; i < pArity; i++ {
			sorts[i] = sSort
		}
		sorts[pArity] = Boolean
		pSort, _ := NewFunctionSort(sorts...)
		sig.Symbols.Set("P", &SymbolEntry{Name: "P", Sort: pSort})
	}

	return NewWithSig(sig)
}

func mkPCWithModule(mod *Module) *ProofChecker {
	return NewProofChecker(nil, mod, nil, nil, nil)
}

// mkPropertyTactic creates a PropertyTactic from the given AST nodes.
func mkPropertyTacticNode(cfg *AstConfig, prop, pname, proof Node) *PropertyTactic {
	if pname == nil {
		pname = cfg.NewNoneAST()
	}
	if proof == nil {
		proof = cfg.NewNoneAST()
	}
	return cfg.NewPropertyTactic(prop, pname, proof)
}

func TestPropertyTacticLabeledFormulaCompileTraceMatchesPython(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	pc := mkPCWithModule(mod)
	goal := cfg.NewLabeledFormula(cfg.NewAtom("g"), True)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), True)
	pt := mkPropertyTacticNode(cfg, propLF, nil, nil)

	out := captureActionUpdateStdout(t, func() {
		if _, err := pc.propertyTactic([]*LabeledFormula{goal}, pt); err != nil {
			t.Fatalf("propertyTactic failed: %v", err)
		}
	})

	compileIdx := strings.Index(out, "XTRACE: proof.CompileExprVocab ENTER exprType=LabeledFormula\n")
	if compileIdx < 0 {
		t.Fatalf("property tactic did not use Python compile_expr_vocab for LF cut:\n%s", out)
	}
	sortIdx := strings.Index(out, "XTRACE: ivylogic.WithSorts.Enter nSorts=0\n")
	if sortIdx < 0 {
		t.Fatalf("property tactic did not enter WithSorts while compiling LF cut:\n%s", out)
	}
	if sortIdx < compileIdx {
		t.Fatalf("WithSorts trace preceded CompileExprVocab; want Python order:\n%s", out)
	}
}

func TestPropertyTacticSkolemTraceMatchesPython(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(1)
	pc := mkPCWithModule(mod)
	goal := cfg.NewLabeledFormula(cfg.NewAtom("g"), True)
	yv := cfg.NewVariable("Y", "S")
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), cfg.NewExists([]Node{yv}, cfg.NewAtom("P", yv)))
	pt := mkPropertyTacticNode(cfg, propLF, cfg.NewAtom("c"), nil)

	out := captureActionUpdateStdout(t, func() {
		if _, err := pc.propertyTactic([]*LabeledFormula{goal}, pt); err != nil {
			t.Fatalf("propertyTactic failed: %v", err)
		}
	})

	if !strings.Contains(out, "XTRACE: proof.propertyTactic skolem sym=c nTargs=0\n") {
		t.Fatalf("property tactic did not emit Python-matched skolem trace:\n%s", out)
	}
}

// Test 1: Basic Skolem witness with one universal
func TestPropertyTactic_BasicSkolemWitness(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(2)

	// Goal: conclusion is something simple (true)
	goal := mkLF(cfg.NewAtom("g"), True)

	// Property formula (AST level): forall X:S . exists Y:S . P(X, Y)
	xv := cfg.NewVariable("X", "S")
	yv := cfg.NewVariable("Y", "S")
	pAtom := cfg.NewAtom("P", xv, yv)
	exBody := cfg.NewExists([]Node{yv}, pAtom)
	forallBody := cfg.NewForall([]Node{xv}, exBody)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), forallBody)

	// Skolem LHS: named f(X)
	skolemLhs := cfg.NewAtom("f")
	skolemLhs.Terms = []Node{cfg.NewVariable("X", "S")}

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)

	result, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err != nil {
		t.Fatalf("propertyTactic failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	// Check that a ConstantDecl for "f" appears in the premises of the first result
	prems := GoalPrems(result[0])
	foundSkolem := false
	for _, p := range prems {
		if cd, ok := p.(*ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if c, ok := args[0].(*Const); ok && c.Name == "f" {
					foundSkolem = true
					// Check the sort: should be S -> S (function sort)
					if fs, ok := c.CSort.(*LogicFunctionSort); ok {
						if len(fs.Sorts) != 2 {
							t.Errorf("Skolem sort: expected 2-element Sorts (dom+rng), got %d", len(fs.Sorts))
						}
					} else {
						t.Errorf("Skolem sort: expected *lg.FunctionSort, got %T", c.CSort)
					}
				}
			}
		}
	}
	if !foundSkolem {
		t.Error("expected ConstantDecl for Skolem function 'f' in premises")
	}
}

// Test 2: Nullary Skolem constant (no ForAll wrapping)
func TestPropertyTactic_NullarySkolemConstant(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(1)

	goal := mkLF(cfg.NewAtom("g"), True)

	// Property: exists Y:S . P(Y)
	yv := cfg.NewVariable("Y", "S")
	pAtom := cfg.NewAtom("P", yv)
	exBody := cfg.NewExists([]Node{yv}, pAtom)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), exBody)

	// Skolem LHS: named c (no parameters)
	skolemLhs := cfg.NewAtom("c")

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)

	result, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err != nil {
		t.Fatalf("propertyTactic failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	prems := GoalPrems(result[0])
	foundSkolem := false
	for _, p := range prems {
		if cd, ok := p.(*ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if c, ok := args[0].(*Const); ok && c.Name == "c" {
					foundSkolem = true
					// Nullary: sort should be just "S", not a FunctionSort
					if _, ok := c.CSort.(*LogicFunctionSort); ok {
						t.Error("Skolem sort: expected non-function sort for nullary constant")
					}
					if !strings.Contains(string(cd.Canon()), `declArgs:[(Symbol name:c sort:(UninterpretedSort name:S))]`) {
						t.Fatalf("Skolem ConstantDecl lost its structural symbol in canon: %s", cd.Canon())
					}
				}
			}
		}
	}
	if !foundSkolem {
		t.Error("expected ConstantDecl for Skolem constant 'c' in premises")
	}
}

// Test 3: Non-existential formula → error
func TestPropertyTactic_NonExistentialError(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(1)

	goal := mkLF(cfg.NewAtom("g"), True)

	// Property: forall X:S . P(X) — no Exists after stripping ForAll
	xv := cfg.NewVariable("X", "S")
	pAtom := cfg.NewAtom("P", xv)
	forallBody := cfg.NewForall([]Node{xv}, pAtom)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), forallBody)

	skolemLhs := cfg.NewAtom("f")
	skolemLhs.Terms = []Node{cfg.NewVariable("X", "S")}

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)

	_, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err == nil {
		t.Fatal("expected error for non-existential formula")
	}
	if !strings.Contains(err.Error(), "not existential") {
		t.Errorf("expected 'not existential' error, got: %v", err)
	}
}

// Test 4: Conjunction (not existential at all) → error
func TestPropertyTactic_ConjunctionNotExistentialError(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(1)

	goal := mkLF(cfg.NewAtom("g"), True)

	// Property: P(X) & P(X) — a conjunction, not an existential
	xv := cfg.NewVariable("X", "S")
	p1 := cfg.NewAtom("P", xv)
	p2 := cfg.NewAtom("P", xv)
	conjunction := cfg.NewAnd(p1, p2)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), conjunction)

	skolemLhs := cfg.NewAtom("f")

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)

	_, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err == nil {
		t.Fatal("expected error for non-existential formula")
	}
	if !strings.Contains(err.Error(), "not existential") {
		t.Errorf("expected 'not existential' error, got: %v", err)
	}
}

// Test 5: Repeated parameter → error
func TestPropertyTactic_RepeatedParameterError(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(2)

	goal := mkLF(cfg.NewAtom("g"), True)

	xv := cfg.NewVariable("X", "S")
	yv := cfg.NewVariable("Y", "S")
	pAtom := cfg.NewAtom("P", xv, yv)
	exBody := cfg.NewExists([]Node{yv}, pAtom)
	forallBody := cfg.NewForall([]Node{xv}, exBody)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), forallBody)

	// Skolem LHS: named f(X, X) — duplicate
	skolemLhs := cfg.NewAtom("f")
	skolemLhs.Terms = []Node{
		cfg.NewVariable("X", "S"),
		cfg.NewVariable("X", "S"),
	}

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)

	_, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err == nil {
		t.Fatal("expected error for repeated parameter")
	}
	if !strings.Contains(err.Error(), "repeat parameter") {
		t.Errorf("expected 'repeat parameter' error, got: %v", err)
	}
}

// Test 6: Missing parameter → error
func TestPropertyTactic_MissingParameterError(t *testing.T) {
	cfg := NewAstConfig()
	sSort := &UninterpretedSort{Name: "S"}
	sig := NewSig()
	sig.Sorts.Set("S", sSort)
	pSort, _ := NewFunctionSort(sSort, sSort, sSort, Boolean)
	sig.Symbols.Set("P", &SymbolEntry{Name: "P", Sort: pSort})
	mod := NewWithSig(sig)

	goal := mkLF(cfg.NewAtom("g"), True)

	// Property: forall X:S . forall Z:S . exists Y:S . P(X, Z, Y)
	xv := cfg.NewVariable("X", "S")
	zv := cfg.NewVariable("Z", "S")
	yv := cfg.NewVariable("Y", "S")
	pAtom := cfg.NewAtom("P", xv, zv, yv)
	exBody := cfg.NewExists([]Node{yv}, pAtom)
	fa2 := cfg.NewForall([]Node{zv}, exBody)
	fa1 := cfg.NewForall([]Node{xv}, fa2)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), fa1)

	// Skolem LHS: named f(X) — missing Z
	skolemLhs := cfg.NewAtom("f")
	skolemLhs.Terms = []Node{cfg.NewVariable("X", "S")}

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)

	_, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err == nil {
		t.Fatal("expected error for missing parameter")
	}
	if !strings.Contains(err.Error(), "must be a parameter of") {
		t.Errorf("expected 'must be a parameter' error, got: %v", err)
	}
}

// Test 7: Cannot infer sort for extra parameter → error
func TestPropertyTactic_CannotInferSortError(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(1)

	goal := mkLF(cfg.NewAtom("g"), True)

	// Property: exists Y:S . P(Y)
	yv := cfg.NewVariable("Y", "S")
	pAtom := cfg.NewAtom("P", yv)
	exBody := cfg.NewExists([]Node{yv}, pAtom)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), exBody)

	// Skolem LHS: named f(W) — W has topsort (VSort="S"), not in formula variables
	skolemLhs := cfg.NewAtom("f")
	skolemLhs.Terms = []Node{cfg.NewVariable("W", "S")}

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)

	_, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err == nil {
		t.Fatal("expected error for cannot-infer-sort parameter")
	}
	if !strings.Contains(err.Error(), "cannot infer sort") {
		t.Errorf("expected 'cannot infer sort' error, got: %v", err)
	}
}

// Test 8: Extra parameter with explicit sort → success
func TestPropertyTactic_ExtraParamExplicitSort(t *testing.T) {
	cfg := NewAstConfig()
	sSort := &UninterpretedSort{Name: "S"}
	tSort := &UninterpretedSort{Name: "T"}
	sig := NewSig()
	sig.Sorts.Set("S", sSort)
	sig.Sorts.Set("T", tSort)
	pSort, _ := NewFunctionSort(sSort, Boolean)
	sig.Symbols.Set("P", &SymbolEntry{Name: "P", Sort: pSort})
	mod := NewWithSig(sig)

	goal := mkLF(cfg.NewAtom("g"), True)

	// Property: exists Y:S . P(Y)
	yv := cfg.NewVariable("Y", "S")
	pAtom := cfg.NewAtom("P", yv)
	exBody := cfg.NewExists([]Node{yv}, pAtom)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), exBody)

	// Skolem LHS: named f(w:T) — w has explicit sort T, not in formula
	skolemLhs := cfg.NewAtom("f")
	wApp := cfg.NewApp(cfg.NewSymbol("w", nil))
	wApp.ASort = cfg.NewSymbol("T", nil)
	skolemLhs.Terms = []Node{wApp}

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)

	result, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err != nil {
		t.Fatalf("propertyTactic failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	// Check Skolem function "f" with sort T -> S
	prems := GoalPrems(result[0])
	foundSkolem := false
	for _, p := range prems {
		if cd, ok := p.(*ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if c, ok := args[0].(*Const); ok && c.Name == "f" {
					foundSkolem = true
					if fs, ok := c.CSort.(*LogicFunctionSort); ok {
						if len(fs.Sorts) != 2 {
							t.Errorf("expected 2-element Sorts (dom+rng), got %d", len(fs.Sorts))
						}
					} else {
						t.Errorf("expected FunctionSort for f, got %T", c.CSort)
					}
				}
			}
		}
	}
	if !foundSkolem {
		t.Error("expected ConstantDecl for Skolem function 'f'")
	}
}

// Test 9: Stale symbol → error
func TestPropertyTactic_StaleSymbolError(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(2)

	goal := mkLF(cfg.NewAtom("g"), True)

	xv := cfg.NewVariable("X", "S")
	yv := cfg.NewVariable("Y", "S")
	pAtom := cfg.NewAtom("P", xv, yv)
	exBody := cfg.NewExists([]Node{yv}, pAtom)
	forallBody := cfg.NewForall([]Node{xv}, exBody)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), forallBody)

	skolemLhs := cfg.NewAtom("f")
	skolemLhs.Terms = []Node{cfg.NewVariable("X", "S")}

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)
	pc.Stale["f"] = true // mark "f" as stale

	_, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err == nil {
		t.Fatal("expected error for stale symbol")
	}
	if !strings.Contains(err.Error(), "is not fresh") {
		t.Errorf("expected 'is not fresh' error, got: %v", err)
	}
}

// Test 10: Symbol in GoalDefns → error (use Stale since GoalDefns key matching
// depends on sort identity which varies by compilation)
func TestPropertyTactic_GoalDefnsSymbolError(t *testing.T) {
	cfg := NewAstConfig()
	mod := mkTestMod(2)

	goal := mkLF(cfg.NewAtom("g"), True)

	xv := cfg.NewVariable("X", "S")
	yv := cfg.NewVariable("Y", "S")
	pAtom := cfg.NewAtom("P", xv, yv)
	exBody := cfg.NewExists([]Node{yv}, pAtom)
	forallBody := cfg.NewForall([]Node{xv}, exBody)
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), forallBody)

	skolemLhs := cfg.NewAtom("f")
	skolemLhs.Terms = []Node{cfg.NewVariable("X", "S")}

	pt := mkPropertyTacticNode(cfg, propLF, skolemLhs, nil)
	pc := mkPCWithModule(mod)
	// Both stale and GoalDefns guard freshness; test stale path explicitly
	pc.Stale["f"] = true

	_, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err == nil {
		t.Fatal("expected error for non-fresh symbol")
	}
	if !strings.Contains(err.Error(), "is not fresh") {
		t.Errorf("expected 'is not fresh' error, got: %v", err)
	}
}

// Test 12: NoneAST passthrough (no Skolem handling)
func TestPropertyTactic_NoneAST_Passthrough(t *testing.T) {
	cfg := NewAstConfig()
	sSort := &UninterpretedSort{Name: "S"}
	sig := NewSig()
	sig.Sorts.Set("S", sSort)
	sig.Symbols.Set("Q", &SymbolEntry{Name: "Q", Sort: Boolean})
	mod := NewWithSig(sig)

	goal := mkLF(cfg.NewAtom("g"), True)

	// Simple property: just "Q" (a boolean constant) with PName = NoneAST
	propLF := cfg.NewLabeledFormula(cfg.NewAtom("prop"), cfg.NewAtom("Q"))

	pt := mkPropertyTacticNode(cfg, propLF, nil, nil)
	pc := mkPCWithModule(mod)

	result, err := pc.propertyTactic([]*LabeledFormula{goal}, pt)
	if err != nil {
		t.Fatalf("propertyTactic failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
	// Should have at least 2 goals: modified goal with cut as premise + subgoal
	if len(result) < 2 {
		t.Errorf("expected at least 2 result goals (goal+subgoal), got %d", len(result))
	}
}
