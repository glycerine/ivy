package goivy

import (
	"strings"
	"testing"
)

// --- helpers ---

func moduleMkConst(name string) *Const {
	return NewConst(name, Boolean)
}

func moduleMkVar(name string) *LogicVariable {
	v, _ := NewVariable(name, &UninterpretedSort{Name: "S"})
	return v
}

func mkBoolVar(name string) *LogicVariable {
	v, _ := NewVariable(name, Boolean)
	return v
}

func mkFuncConst(name string, arity int) *Const {
	sorts := make([]Sort, arity+1)
	for i := 0; i < arity; i++ {
		sorts[i] = &UninterpretedSort{Name: "S"}
	}
	sorts[arity] = Boolean
	fs, _ := NewFunctionSort(sorts...)
	return NewConst(name, fs)
}

// --- Clauses tests ---

func TestModuleClauseOpsNewClausesEmpty(t *testing.T) {
	c := NewClauses(nil, nil, nil)
	if !c.IsTrue() {
		t.Error("empty clauses should be true")
	}
	if c.IsFalse() {
		t.Error("empty clauses should not be false")
	}
}

func TestModuleClauseOpsNewClausesFlattensAnd(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	inner := &LogicAnd{Terms: []Expr{a, b}}
	c := NewClauses([]Expr{inner}, nil, nil)
	if len(c.Fmlas) != 2 {
		t.Errorf("expected 2 formulas after flattening, got %d", len(c.Fmlas))
	}
}

func TestModuleClauseOpsClausesIsFalse(t *testing.T) {
	c := NewClauses([]Expr{False}, nil, nil)
	if !c.IsFalse() {
		t.Error("expected IsFalse")
	}
}

func TestModuleClauseOpsClausesIsTrue(t *testing.T) {
	c := NewClauses([]Expr{True}, nil, nil)
	if !c.IsTrue() {
		t.Error("expected IsTrue")
	}
}

func TestModuleClauseOpsClausesCopy(t *testing.T) {
	a := moduleMkConst("a")
	c := NewClauses([]Expr{a}, nil, nil)
	cp := c.Copy()
	if !c.Equal(cp) {
		t.Error("copy should be equal")
	}
	// Mutation of copy should not affect original
	cp.Fmlas = append(cp.Fmlas, moduleMkConst("b"))
	if len(c.Fmlas) == len(cp.Fmlas) {
		t.Error("original should not be affected by copy mutation")
	}
}

func TestModuleClauseOpsClausesToOpenFormula(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := NewClauses([]Expr{a, b}, nil, nil)
	f := c.ToOpenFormula()
	and, ok := f.(*LogicAnd)
	if !ok {
		t.Fatalf("expected And, got %T", f)
	}
	if len(and.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(and.Terms))
	}
}

func TestModuleClauseOpsClausesToOpenFormulaWithDefs(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	def := NewIvyDefinition(a, b)
	c := NewClauses([]Expr{a}, []*IvyDefinition{def}, nil)
	f := c.ToOpenFormula()
	and, ok := f.(*LogicAnd)
	if !ok {
		t.Fatalf("expected And, got %T", f)
	}
	// Should have def constraint + formula
	if len(and.Terms) != 2 {
		t.Errorf("expected 2 terms (def constraint + fmla), got %d", len(and.Terms))
	}
}

func TestModuleClauseOpsClausesString(t *testing.T) {
	a := moduleMkConst("a")
	c := NewClauses([]Expr{a}, nil, nil)
	s := c.String()
	if !strings.Contains(s, "Clauses") {
		t.Errorf("String should contain 'Clauses': %s", s)
	}
	if !strings.Contains(s, "a") {
		t.Errorf("String should contain 'a': %s", s)
	}
}

func TestModuleClauseOpsClausesDefIdx(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	def := NewIvyDefinition(a, b)
	c := NewClauses(nil, []*IvyDefinition{def}, nil)
	idx, ok := c.DefIdx[Key(a)]
	if !ok {
		t.Fatalf("DefIdx should contain key %q", Key(a))
	}
	if c.Defs[idx] != def {
		t.Error("DefIdx should point to the correct definition")
	}
}

func TestModuleClauseOpsClausesEqual(t *testing.T) {
	a := moduleMkConst("a")
	c1 := NewClauses([]Expr{a}, nil, nil)
	c2 := NewClauses([]Expr{a}, nil, nil)
	if !c1.Equal(c2) {
		t.Error("equal clauses should be equal")
	}
	b := moduleMkConst("b")
	c3 := NewClauses([]Expr{b}, nil, nil)
	if c1.Equal(c3) {
		t.Error("different clauses should not be equal")
	}
}

func TestModuleClauseOpsClausesIsUniversalFirstOrder(t *testing.T) {
	a := moduleMkConst("a")
	c := NewClauses([]Expr{a}, nil, nil)
	if !c.IsUniversalFirstOrder() {
		t.Error("expected universal first order")
	}

	// With defs, should not be universal first order
	def := NewIvyDefinition(a, moduleMkConst("b"))
	c2 := NewClauses([]Expr{a}, []*IvyDefinition{def}, nil)
	if c2.IsUniversalFirstOrder() {
		t.Error("clauses with defs should not be universal first order")
	}

	// With Skolem symbol
	sk := NewConst("__sk", Boolean)
	c3 := NewClauses([]Expr{sk}, nil, nil)
	if c3.IsUniversalFirstOrder() {
		t.Error("clauses with skolem should not be universal first order")
	}
}

// --- Constructor tests ---

func TestModuleClauseOpsTrueClauses(t *testing.T) {
	c := TrueClauses(nil)
	if !c.IsTrue() {
		t.Error("TrueClauses should be true")
	}
	if c.IsFalse() {
		t.Error("TrueClauses should not be false")
	}
}

func TestModuleClauseOpsFalseClauses(t *testing.T) {
	c := FalseClauses(nil)
	if !c.IsFalse() {
		t.Error("FalseClauses should be false")
	}
}

func TestModuleClauseOpsFormulaToClauses(t *testing.T) {
	a := moduleMkConst("a")
	c := FormulaToClauses(a, nil)
	if len(c.Fmlas) != 1 {
		t.Errorf("expected 1 formula, got %d", len(c.Fmlas))
	}
	if !c.Fmlas[0].Equal(a) {
		t.Error("formula should be preserved")
	}
}

func TestModuleClauseOpsFormulaToClausesUnwrapsSingleton(t *testing.T) {
	a := moduleMkConst("a")
	wrapped := &LogicAnd{Terms: []Expr{a}}
	c := FormulaToClauses(wrapped, nil)
	if len(c.Fmlas) != 1 {
		t.Errorf("expected 1 formula, got %d", len(c.Fmlas))
	}
	if !c.Fmlas[0].Equal(a) {
		t.Error("singleton And should be unwrapped")
	}
}

func TestModuleClauseOpsFormulaToClausesStripsForAll(t *testing.T) {
	x := mkBoolVar("X")
	body := x
	fa := &ForAll{Variables: []*LogicVariable{x}, Body: body}
	c := FormulaToClauses(fa, nil)
	if len(c.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(c.Fmlas))
	}
	// Should have stripped ForAll
	if _, ok := c.Fmlas[0].(*ForAll); ok {
		t.Error("ForAll should be stripped by dropUniversals")
	}
}

// --- And/Or tests ---

func TestModuleClauseOpsAndClausesTyped(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c1 := NewClauses([]Expr{a}, nil, nil)
	c2 := NewClauses([]Expr{b}, nil, nil)
	result := AndClausesTyped(c1, c2)
	if len(result.Fmlas) != 2 {
		t.Errorf("expected 2 formulas, got %d", len(result.Fmlas))
	}
}

func TestModuleClauseOpsAndClausesEmpty(t *testing.T) {
	result := AndClausesTyped()
	if !result.IsTrue() {
		t.Error("empty conjunction should be true")
	}
}

func TestModuleClauseOpsAndClausesWithFalse(t *testing.T) {
	a := moduleMkConst("a")
	c1 := NewClauses([]Expr{a}, nil, nil)
	c2 := FalseClauses(nil)
	result := AndClausesTyped(c1, c2)
	if !result.IsFalse() {
		t.Error("conjunction with false should be false")
	}
}

func TestModuleClauseOpsOrClausesTyped(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c1 := NewClauses([]Expr{a}, nil, nil)
	c2 := NewClauses([]Expr{b}, nil, nil)
	result := OrClausesTyped(c1, c2)
	// Should have introduced fresh variables
	if len(result.Fmlas) == 0 {
		t.Error("or clauses should produce formulas")
	}
	// First formula should be Or(v1, v2)
	firstOr, ok := result.Fmlas[0].(*LogicOr)
	if !ok {
		t.Fatalf("first formula should be Or, got %T", result.Fmlas[0])
	}
	if len(firstOr.Terms) != 2 {
		t.Errorf("expected 2 terms in Or, got %d", len(firstOr.Terms))
	}
}

func TestModuleClauseOpsOrClausesSingleNonFalse(t *testing.T) {
	a := moduleMkConst("a")
	c1 := NewClauses([]Expr{a}, nil, nil)
	c2 := FalseClauses(nil)
	result := OrClausesTyped(c1, c2)
	// Only c1 is non-false, so result should be c1
	if len(result.Fmlas) != 1 || !result.Fmlas[0].Equal(a) {
		t.Error("or of (a, false) should be a")
	}
}

func TestModuleClauseOpsOrClausesAllFalse(t *testing.T) {
	result := OrClausesTyped(FalseClauses(nil), FalseClauses(nil))
	if !result.IsFalse() {
		t.Error("or of all false should be false")
	}
}

// --- IteClauses tests ---

func TestModuleClauseOpsIteClauses(t *testing.T) {
	cond := moduleMkConst("cond")
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	thenCls := NewClauses([]Expr{a}, nil, nil)
	elseCls := NewClauses([]Expr{b}, nil, nil)
	result := IteClauses(cond, thenCls, elseCls)
	if len(result.Fmlas) < 2 {
		t.Errorf("expected at least 2 formulas, got %d", len(result.Fmlas))
	}
	if len(result.Defs) < 1 {
		t.Errorf("expected at least 1 definition (for condition), got %d", len(result.Defs))
	}
}

func TestModuleClauseOpsIteClausesBothFalse(t *testing.T) {
	cond := moduleMkConst("cond")
	result := IteClauses(cond, FalseClauses(nil), FalseClauses(nil))
	if !result.IsFalse() {
		t.Error("ite with both false should be false")
	}
}

// --- NegateClauses tests ---

func TestModuleClauseOpsNegateClauses(t *testing.T) {
	a := moduleMkConst("a")
	c := NewClauses([]Expr{a}, nil, nil)
	neg := NegateClauses(c)
	if neg.IsFalse() && !c.IsFalse() {
		// Negation of non-false should not be trivially false
		// (unless a is True, which it's not)
	}
	// The negated clause should have some formula
	if len(neg.Fmlas) == 0 {
		t.Error("negated clauses should have formulas")
	}
}

// --- ConditionClauses tests ---

func TestModuleClauseOpsConditionClauses(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := NewClauses([]Expr{a}, nil, nil)
	result := ConditionClauses(c, b)
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
	// Should be Or(Not(b), a)
	or, ok := result.Fmlas[0].(*LogicOr)
	if !ok {
		t.Fatalf("expected Or, got %T", result.Fmlas[0])
	}
	if len(or.Terms) != 2 {
		t.Errorf("expected 2 terms in Or, got %d", len(or.Terms))
	}
	// First term should be Not(b)
	notB, ok := or.Terms[0].(*LogicNot)
	if !ok {
		t.Fatalf("expected Not, got %T", or.Terms[0])
	}
	if !notB.Body.Equal(b) {
		t.Error("negated condition should be b")
	}
}

// --- ClausesUsingSymbols tests ---

func TestModuleClauseOpsClausesUsingSymbols(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := NewClauses([]Expr{a, b}, nil, nil)
	syms := NewInsMap[NodeKey, Expr]()
	syms.Set(Key(a), a)
	result := ClausesUsingSymbols(syms, c)
	if len(result.Fmlas) != 1 {
		t.Errorf("expected 1 formula using symbol 'a', got %d", len(result.Fmlas))
	}
	if !result.Fmlas[0].Equal(a) {
		t.Error("filtered formula should be 'a'")
	}
}

// --- AST utility tests ---

func TestModuleClauseOpsSymbolsAST(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	fmla := &LogicAnd{Terms: []Expr{a, b}}
	syms := SymbolsAST(fmla)
	if len(syms) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(syms))
	}
}

func TestModuleClauseOpsUsedSymbolsAST(t *testing.T) {
	f := mkFuncConst("f", 1)
	x := moduleMkVar("X")
	app := MustApply(f, x)
	syms := UsedSymbolsAST(app)
	if _, ok := syms.Get2(Key(f)); !ok {
		t.Error("should contain function symbol f")
	}
}

func TestModuleClauseOpsVariablesAST(t *testing.T) {
	x := moduleMkVar("X")
	y := moduleMkVar("Y")
	fmla := &LogicAnd{Terms: []Expr{x, y}}
	vars := VariablesAST(fmla)
	if len(vars) != 2 {
		t.Errorf("expected 2 variables, got %d", len(vars))
	}
}

func TestModuleClauseOpsVariablesASTSkipsBound(t *testing.T) {
	x := mkBoolVar("X")
	y := mkBoolVar("Y")
	body := &LogicAnd{Terms: []Expr{x, y}}
	fa := &ForAll{Variables: []*LogicVariable{x}, Body: body}
	vars := VariablesAST(fa)
	// Only Y should be free
	if len(vars) != 1 {
		t.Errorf("expected 1 free variable, got %d", len(vars))
	}
}

func TestModuleClauseOpsUsedVariablesAST(t *testing.T) {
	x := moduleMkVar("X")
	fmla := &LogicAnd{Terms: []Expr{x}}
	vars := UsedVariablesAST(fmla)
	if len(vars) != 1 {
		t.Errorf("expected 1 variable, got %d", len(vars))
	}
}

func TestModuleClauseOpsSubstituteConstantsAST(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	fmla := &LogicAnd{Terms: []Expr{a}}
	result := SubstituteConstantsAST(fmla, map[NodeKey]Expr{Key(a): b})
	and, ok := result.(*LogicAnd)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	if !and.Terms[0].Equal(b) {
		t.Error("a should be substituted with b")
	}
}

func TestSubstituteConstantsActionPreservesNativeAtomExpr(t *testing.T) {
	cfg := NewAstConfig()
	act := NewNativeAction(
		cfg.NewNativeCode("byte random `8`"),
		newNativeAtomExpr(cfg.NewAtom("8")),
	)

	result := SubstituteConstantsAction(act, nil)
	native, ok := result.(*LogicNativeAction)
	if !ok {
		t.Fatalf("SubstituteConstantsAction = %T, want *LogicNativeAction", result)
	}
	if len(native.Params) != 1 {
		t.Fatalf("native params = %d, want 1", len(native.Params))
	}
	if _, ok := native.Params[0].(*nativeAtomExpr); !ok {
		t.Fatalf("native param = %T, want *nativeAtomExpr", native.Params[0])
	}
	if got := TypeName(native.Params[0]); got != "Atom" {
		t.Fatalf("TypeName(native param) = %q, want Atom", got)
	}
}

func TestModuleClauseOpsRenameAST(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	fmla := &LogicAnd{Terms: []Expr{a}}
	result := RenameAST(fmla, map[NodeKey]*Const{Key(a): b})
	and, ok := result.(*LogicAnd)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	if !and.Terms[0].Equal(b) {
		t.Error("a should be renamed to b")
	}
}

func TestModuleClauseOpsNegate(t *testing.T) {
	a := moduleMkConst("a")
	notA := Negate(a)
	not, ok := notA.(*LogicNot)
	if !ok {
		t.Fatalf("expected Not, got %T", notA)
	}
	if !not.Body.Equal(a) {
		t.Error("body should be a")
	}

	// Double negation elimination
	result := Negate(notA)
	if !result.Equal(a) {
		t.Error("double negation should yield a")
	}
}

func TestModuleClauseOpsIsTrue(t *testing.T) {
	if !ModuleIsTrue(True) {
		t.Error("True should be true")
	}
	if ModuleIsTrue(False) {
		t.Error("False should not be true")
	}
}

func TestModuleClauseOpsIsFalse(t *testing.T) {
	if !ModuleIsFalse(False) {
		t.Error("False should be false")
	}
	if ModuleIsFalse(True) {
		t.Error("True should not be false")
	}
}

func TestModuleClauseOpsSymPlaceholders(t *testing.T) {
	f := mkFuncConst("f", 2)
	phs := SymPlaceholders(f)
	if len(phs) != 2 {
		t.Errorf("expected 2 placeholders, got %d", len(phs))
	}
	if phs[0].Name != "V0" || phs[1].Name != "V1" {
		t.Errorf("expected V0, V1, got %s, %s", phs[0].Name, phs[1].Name)
	}
}

func TestModuleClauseOpsSymPlaceholdersNoArgs(t *testing.T) {
	c := moduleMkConst("c")
	phs := SymPlaceholders(c)
	if len(phs) != 0 {
		t.Errorf("expected 0 placeholders, got %d", len(phs))
	}
}

func TestModuleClauseOpsSymInst(t *testing.T) {
	f := mkFuncConst("f", 1)
	inst := SymInst(f)
	app, ok := inst.(*Apply)
	if !ok {
		t.Fatalf("expected Apply, got %T", inst)
	}
	if !app.Func.Equal(f) {
		t.Error("func should be f")
	}
	if len(app.Terms) != 1 {
		t.Errorf("expected 1 arg, got %d", len(app.Terms))
	}
}

func TestModuleClauseOpsSymInstNoArgs(t *testing.T) {
	c := moduleMkConst("c")
	inst := SymInst(c)
	if !inst.Equal(c) {
		t.Error("0-arity sym inst should return the symbol itself")
	}
}

func TestModuleClauseOpsEqAtom(t *testing.T) {
	x := moduleMkVar("X")
	y := moduleMkVar("Y")
	eq := EqAtom(x, y)
	e, ok := eq.(*Eq)
	if !ok {
		t.Fatalf("expected Eq, got %T", eq)
	}
	if !e.T1.Equal(x) || !e.T2.Equal(y) {
		t.Error("EqAtom should wrap x and y")
	}
}

func TestModuleClauseOpsEqLit(t *testing.T) {
	x := moduleMkVar("X")
	y := moduleMkVar("Y")
	lit := EqLit(x, y)
	if lit.Polarity != 1 {
		t.Error("EqLit should have positive polarity")
	}
	e, ok := lit.Atom.(*Eq)
	if !ok {
		t.Fatalf("expected Eq atom, got %T", lit.Atom)
	}
	if !e.T1.Equal(x) || !e.T2.Equal(y) {
		t.Error("EqLit should have x == y as atom")
	}
}

func TestModuleClauseOpsCollectAndList(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := moduleMkConst("c")
	nested := &LogicAnd{Terms: []Expr{a, &LogicAnd{Terms: []Expr{b, c}}}}
	result := CollectAndList([]Expr{nested})
	if len(result) != 3 {
		t.Errorf("expected 3 after flattening, got %d", len(result))
	}
}

func TestModuleClauseOpsCollectOr(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := moduleMkConst("c")
	nested := &LogicOr{Terms: []Expr{a, &LogicOr{Terms: []Expr{b, c}}}}
	result := CollectOr(nested)
	if len(result) != 3 {
		t.Errorf("expected 3 after flattening, got %d", len(result))
	}
}

func TestModuleClauseOpsDropUniversals(t *testing.T) {
	x := mkBoolVar("X")
	body := x
	fa := &ForAll{Variables: []*LogicVariable{x}, Body: body}
	result := DropUniversals(fa)
	if _, ok := result.(*ForAll); ok {
		t.Error("ForAll should be stripped")
	}
	if !result.Equal(x) {
		t.Error("should return body")
	}
}

func TestModuleClauseOpsDropUniversalsNested(t *testing.T) {
	x := mkBoolVar("X")
	y := mkBoolVar("Y")
	body := &LogicAnd{Terms: []Expr{x, y}}
	inner := &ForAll{Variables: []*LogicVariable{y}, Body: body}
	outer := &ForAll{Variables: []*LogicVariable{x}, Body: inner}
	result := DropUniversals(outer)
	// Should strip both ForAlls
	if _, ok := result.(*ForAll); ok {
		t.Error("all ForAlls should be stripped")
	}
}

func TestModuleClauseOpsIsGroundAST(t *testing.T) {
	a := moduleMkConst("a")
	if !IsGroundAST(a) {
		t.Error("constant should be ground")
	}
	x := moduleMkVar("X")
	if IsGroundAST(x) {
		t.Error("variable should not be ground")
	}
}

func TestModuleClauseOpsNormalizeFreeVariables(t *testing.T) {
	x := mkBoolVar("X")
	y := mkBoolVar("Y")
	fmla := &LogicAnd{Terms: []Expr{x, y}}
	oldVars, newVars, result := NormalizeFreeVariables(fmla)
	if len(oldVars) != 2 || len(newVars) != 2 {
		t.Fatalf("expected 2 old and 2 new vars, got %d, %d", len(oldVars), len(newVars))
	}
	if newVars[0].Name != "V0" || newVars[1].Name != "V1" {
		t.Errorf("expected V0, V1, got %s, %s", newVars[0].Name, newVars[1].Name)
	}
	and, ok := result.(*LogicAnd)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	for _, term := range and.Terms {
		v, ok := term.(*LogicVariable)
		if !ok {
			t.Fatalf("expected Var, got %T", term)
		}
		if !strings.HasPrefix(v.Name, "V") {
			t.Errorf("expected normalized name starting with V, got %s", v.Name)
		}
	}
}

// --- Rename/Substitute on Clauses tests ---

func TestModuleClauseOpsRenameClauses(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := NewClauses([]Expr{a}, nil, nil)
	result := RenameClauses(c, map[NodeKey]*Const{Key(a): b})
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
	if !result.Fmlas[0].Equal(b) {
		t.Error("a should be renamed to b")
	}
}

func TestModuleClauseOpsSubstituteConstantsClauses(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := NewClauses([]Expr{a}, nil, nil)
	result := SubstituteConstantsClauses(c, map[NodeKey]Expr{Key(a): b})
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
	if !result.Fmlas[0].Equal(b) {
		t.Error("a should be substituted with b")
	}
}

func TestModuleClauseOpsClausesSymbols(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := NewClauses([]Expr{a, b}, nil, nil)
	syms := c.Symbols()
	if syms.Len() != 2 {
		t.Errorf("expected 2 symbols, got %d", syms.Len())
	}
}

func TestModuleClauseOpsClausesToFormula(t *testing.T) {
	x := moduleMkVar("X")
	c := NewClauses([]Expr{x}, nil, nil)
	f := c.ToFormula()
	// ToFormula uses CloseEPR which distributes through And.
	// to_open_formula() returns And(X), then CloseEPR(And(X)) → And(ForAll(X, X)).
	// This matches Python: close_epr(And(X)) = LogicAnd(close_epr(X)) = LogicAnd(ForAll([X], X)).
	and, ok := f.(*LogicAnd)
	if !ok {
		t.Fatalf("expected And (CloseEPR distributes through And), got %T: %s", f, f)
	}
	if len(and.Terms) != 1 {
		t.Fatalf("expected 1 term in And, got %d", len(and.Terms))
	}
	fa, ok := and.Terms[0].(*ForAll)
	if !ok {
		t.Fatalf("expected ForAll inside And, got %T: %s", and.Terms[0], and.Terms[0])
	}
	if len(fa.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(fa.Variables))
	}
}

// --- Fuzz tests ---

func FuzzCollectAndList(f *testing.F) {
	f.Add(0, 1, 2)
	f.Fuzz(func(t *testing.T, n1, n2, n3 int) {
		// Create some constants based on the fuzz inputs
		numConsts := abs(n1)%5 + 1
		consts := make([]Expr, 0, numConsts)
		for i := 0; i < numConsts; i++ {
			consts = append(consts, moduleMkConst("c"+strings.Repeat("x", i)))
		}

		// Build a nested And
		inner := &LogicAnd{Terms: consts}
		numExtra := abs(n2)%3 + 1
		outerTerms := []Expr{inner}
		for i := 0; i < numExtra; i++ {
			outerTerms = append(outerTerms, moduleMkConst("d"+strings.Repeat("y", i)))
		}
		outer := &LogicAnd{Terms: outerTerms}

		result := CollectAndList([]Expr{outer})
		// Result should have at least as many elements as inner + outer extras
		expectedMin := numConsts + numExtra
		if len(result) < expectedMin {
			t.Errorf("flattening produced %d elements, expected at least %d", len(result), expectedMin)
		}
		// No And should remain at the top level of result
		for _, r := range result {
			if a, ok := r.(*LogicAnd); ok && len(a.Terms) > 0 {
				t.Error("flattening should not leave non-empty And nodes")
			}
		}
	})
}

func FuzzAndOrClauses(f *testing.F) {
	f.Add(1, 2, true)
	f.Fuzz(func(t *testing.T, n1, n2 int, doOr bool) {
		// Bound the number of clauses to a safe positive range
		numCls := abs(n1)%4 + 1
		args := make([]*Clauses, numCls)
		for i := 0; i < numCls; i++ {
			numFmlas := abs(n2)%3 + 1
			fmlas := make([]Expr, numFmlas)
			for j := 0; j < numFmlas; j++ {
				fmlas[j] = moduleMkConst("f" + strings.Repeat("z", (i+j)%5))
			}
			args[i] = NewClauses(fmlas, nil, nil)
		}

		var result *Clauses
		if doOr {
			result = OrClausesTyped(args...)
		} else {
			result = AndClausesTyped(args...)
		}

		// Basic invariants
		if result == nil {
			t.Fatal("result should not be nil")
		}
		// AndClauses: result should have at least as many formulas as sum of inputs
		// (unless one is false)
		if !doOr {
			total := 0
			for _, a := range args {
				total += len(a.Fmlas)
			}
			if len(result.Fmlas) < total && !result.IsFalse() {
				t.Errorf("and result has %d fmlas, expected at least %d", len(result.Fmlas), total)
			}
		}
		// OrClauses: result should have formulas
		if doOr && len(result.Fmlas) == 0 && !result.IsFalse() {
			t.Error("or result should have formulas or be false")
		}
	})
}

func abs(x int) int {
	if x < 0 {
		if x == -x { // overflow: MinInt
			return 0
		}
		return -x
	}
	return x
}
