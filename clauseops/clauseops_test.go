package clauseops

import (
	"strings"
	"testing"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// --- helpers ---

func mkConst(name string) *lg.Const {
	return lg.NewConst(name, lg.Boolean)
}

func mkVar(name string) *lg.Var {
	v, _ := lg.NewVar(name, &lg.UninterpretedSort{Name: "S"})
	return v
}

func mkBoolVar(name string) *lg.Var {
	v, _ := lg.NewVar(name, lg.Boolean)
	return v
}

func mkFuncConst(name string, arity int) *lg.Const {
	sorts := make([]lg.Sort, arity+1)
	for i := 0; i < arity; i++ {
		sorts[i] = &lg.UninterpretedSort{Name: "S"}
	}
	sorts[arity] = lg.Boolean
	fs, _ := lg.NewFunctionSort(sorts...)
	return lg.NewConst(name, fs)
}

// --- Clauses tests ---

func TestNewClausesEmpty(t *testing.T) {
	c := NewClauses(nil, nil, nil)
	if !c.IsTrue() {
		t.Error("empty clauses should be true")
	}
	if c.IsFalse() {
		t.Error("empty clauses should not be false")
	}
}

func TestNewClausesFlattensAnd(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	inner := &lg.And{Terms: []lg.Node{a, b}}
	c := NewClauses([]lg.Node{inner}, nil, nil)
	if len(c.Fmlas) != 2 {
		t.Errorf("expected 2 formulas after flattening, got %d", len(c.Fmlas))
	}
}

func TestClausesIsFalse(t *testing.T) {
	c := NewClauses([]lg.Node{lg.False}, nil, nil)
	if !c.IsFalse() {
		t.Error("expected IsFalse")
	}
}

func TestClausesIsTrue(t *testing.T) {
	c := NewClauses([]lg.Node{lg.True}, nil, nil)
	if !c.IsTrue() {
		t.Error("expected IsTrue")
	}
}

func TestClausesCopy(t *testing.T) {
	a := mkConst("a")
	c := NewClauses([]lg.Node{a}, nil, nil)
	cp := c.Copy()
	if !c.Equal(cp) {
		t.Error("copy should be equal")
	}
	// Mutation of copy should not affect original
	cp.Fmlas = append(cp.Fmlas, mkConst("b"))
	if len(c.Fmlas) == len(cp.Fmlas) {
		t.Error("original should not be affected by copy mutation")
	}
}

func TestClausesToOpenFormula(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := NewClauses([]lg.Node{a, b}, nil, nil)
	f := c.ToOpenFormula()
	and, ok := f.(*lg.And)
	if !ok {
		t.Fatalf("expected And, got %T", f)
	}
	if len(and.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(and.Terms))
	}
}

func TestClausesToOpenFormulaWithDefs(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	def := il.NewDefinition(a, b)
	c := NewClauses([]lg.Node{a}, []*il.Definition{def}, nil)
	f := c.ToOpenFormula()
	and, ok := f.(*lg.And)
	if !ok {
		t.Fatalf("expected And, got %T", f)
	}
	// Should have def constraint + formula
	if len(and.Terms) != 2 {
		t.Errorf("expected 2 terms (def constraint + fmla), got %d", len(and.Terms))
	}
}

func TestClausesString(t *testing.T) {
	a := mkConst("a")
	c := NewClauses([]lg.Node{a}, nil, nil)
	s := c.String()
	if !strings.Contains(s, "Clauses") {
		t.Errorf("String should contain 'Clauses': %s", s)
	}
	if !strings.Contains(s, "a") {
		t.Errorf("String should contain 'a': %s", s)
	}
}

func TestClausesDefIdx(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	def := il.NewDefinition(a, b)
	c := NewClauses(nil, []*il.Definition{def}, nil)
	idx, ok := c.DefIdx["a"]
	if !ok {
		t.Fatal("DefIdx should contain key 'a'")
	}
	if c.Defs[idx] != def {
		t.Error("DefIdx should point to the correct definition")
	}
}

func TestClausesEqual(t *testing.T) {
	a := mkConst("a")
	c1 := NewClauses([]lg.Node{a}, nil, nil)
	c2 := NewClauses([]lg.Node{a}, nil, nil)
	if !c1.Equal(c2) {
		t.Error("equal clauses should be equal")
	}
	b := mkConst("b")
	c3 := NewClauses([]lg.Node{b}, nil, nil)
	if c1.Equal(c3) {
		t.Error("different clauses should not be equal")
	}
}

func TestClausesIsUniversalFirstOrder(t *testing.T) {
	a := mkConst("a")
	c := NewClauses([]lg.Node{a}, nil, nil)
	if !c.IsUniversalFirstOrder() {
		t.Error("expected universal first order")
	}

	// With defs, should not be universal first order
	def := il.NewDefinition(a, mkConst("b"))
	c2 := NewClauses([]lg.Node{a}, []*il.Definition{def}, nil)
	if c2.IsUniversalFirstOrder() {
		t.Error("clauses with defs should not be universal first order")
	}

	// With Skolem symbol
	sk := lg.NewConst("__sk", lg.Boolean)
	c3 := NewClauses([]lg.Node{sk}, nil, nil)
	if c3.IsUniversalFirstOrder() {
		t.Error("clauses with skolem should not be universal first order")
	}
}

// --- Constructor tests ---

func TestTrueClauses(t *testing.T) {
	c := TrueClauses(nil)
	if !c.IsTrue() {
		t.Error("TrueClauses should be true")
	}
	if c.IsFalse() {
		t.Error("TrueClauses should not be false")
	}
}

func TestFalseClauses(t *testing.T) {
	c := FalseClauses(nil)
	if !c.IsFalse() {
		t.Error("FalseClauses should be false")
	}
}

func TestFormulaToClauses(t *testing.T) {
	a := mkConst("a")
	c := FormulaToClauses(a, nil)
	if len(c.Fmlas) != 1 {
		t.Errorf("expected 1 formula, got %d", len(c.Fmlas))
	}
	if !c.Fmlas[0].Equal(a) {
		t.Error("formula should be preserved")
	}
}

func TestFormulaToClausesUnwrapsSingleton(t *testing.T) {
	a := mkConst("a")
	wrapped := &lg.And{Terms: []lg.Node{a}}
	c := FormulaToClauses(wrapped, nil)
	if len(c.Fmlas) != 1 {
		t.Errorf("expected 1 formula, got %d", len(c.Fmlas))
	}
	if !c.Fmlas[0].Equal(a) {
		t.Error("singleton And should be unwrapped")
	}
}

func TestFormulaToClausesStripsForAll(t *testing.T) {
	x := mkBoolVar("X")
	body := x
	fa := &lg.ForAll{Variables: []*lg.Var{x}, Body: body}
	c := FormulaToClauses(fa, nil)
	if len(c.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(c.Fmlas))
	}
	// Should have stripped ForAll
	if _, ok := c.Fmlas[0].(*lg.ForAll); ok {
		t.Error("ForAll should be stripped by dropUniversals")
	}
}

// --- And/Or tests ---

func TestAndClausesTyped(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c1 := NewClauses([]lg.Node{a}, nil, nil)
	c2 := NewClauses([]lg.Node{b}, nil, nil)
	result := AndClausesTyped(c1, c2)
	if len(result.Fmlas) != 2 {
		t.Errorf("expected 2 formulas, got %d", len(result.Fmlas))
	}
}

func TestAndClausesEmpty(t *testing.T) {
	result := AndClausesTyped()
	if !result.IsTrue() {
		t.Error("empty conjunction should be true")
	}
}

func TestAndClausesWithFalse(t *testing.T) {
	a := mkConst("a")
	c1 := NewClauses([]lg.Node{a}, nil, nil)
	c2 := FalseClauses(nil)
	result := AndClausesTyped(c1, c2)
	if !result.IsFalse() {
		t.Error("conjunction with false should be false")
	}
}

func TestOrClausesTyped(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c1 := NewClauses([]lg.Node{a}, nil, nil)
	c2 := NewClauses([]lg.Node{b}, nil, nil)
	result := OrClausesTyped(c1, c2)
	// Should have introduced fresh variables
	if len(result.Fmlas) == 0 {
		t.Error("or clauses should produce formulas")
	}
	// First formula should be Or(v1, v2)
	firstOr, ok := result.Fmlas[0].(*lg.Or)
	if !ok {
		t.Fatalf("first formula should be Or, got %T", result.Fmlas[0])
	}
	if len(firstOr.Terms) != 2 {
		t.Errorf("expected 2 terms in Or, got %d", len(firstOr.Terms))
	}
}

func TestOrClausesSingleNonFalse(t *testing.T) {
	a := mkConst("a")
	c1 := NewClauses([]lg.Node{a}, nil, nil)
	c2 := FalseClauses(nil)
	result := OrClausesTyped(c1, c2)
	// Only c1 is non-false, so result should be c1
	if len(result.Fmlas) != 1 || !result.Fmlas[0].Equal(a) {
		t.Error("or of (a, false) should be a")
	}
}

func TestOrClausesAllFalse(t *testing.T) {
	result := OrClausesTyped(FalseClauses(nil), FalseClauses(nil))
	if !result.IsFalse() {
		t.Error("or of all false should be false")
	}
}

// --- IteClauses tests ---

func TestIteClauses(t *testing.T) {
	cond := mkConst("cond")
	a := mkConst("a")
	b := mkConst("b")
	thenCls := NewClauses([]lg.Node{a}, nil, nil)
	elseCls := NewClauses([]lg.Node{b}, nil, nil)
	result := IteClauses(cond, thenCls, elseCls)
	if len(result.Fmlas) < 2 {
		t.Errorf("expected at least 2 formulas, got %d", len(result.Fmlas))
	}
	if len(result.Defs) < 1 {
		t.Errorf("expected at least 1 definition (for condition), got %d", len(result.Defs))
	}
}

func TestIteClausesBothFalse(t *testing.T) {
	cond := mkConst("cond")
	result := IteClauses(cond, FalseClauses(nil), FalseClauses(nil))
	if !result.IsFalse() {
		t.Error("ite with both false should be false")
	}
}

// --- NegateClauses tests ---

func TestNegateClauses(t *testing.T) {
	a := mkConst("a")
	c := NewClauses([]lg.Node{a}, nil, nil)
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

func TestConditionClauses(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := NewClauses([]lg.Node{a}, nil, nil)
	result := ConditionClauses(c, b)
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
	// Should be Or(Not(b), a)
	or, ok := result.Fmlas[0].(*lg.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", result.Fmlas[0])
	}
	if len(or.Terms) != 2 {
		t.Errorf("expected 2 terms in Or, got %d", len(or.Terms))
	}
	// First term should be Not(b)
	notB, ok := or.Terms[0].(*lg.Not)
	if !ok {
		t.Fatalf("expected Not, got %T", or.Terms[0])
	}
	if !notB.Body.Equal(b) {
		t.Error("negated condition should be b")
	}
}

// --- ClausesUsingSymbols tests ---

func TestClausesUsingSymbols(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := NewClauses([]lg.Node{a, b}, nil, nil)
	syms := map[lg.NodeKey]lg.Node{lg.Key(a): a}
	result := ClausesUsingSymbols(syms, c)
	if len(result.Fmlas) != 1 {
		t.Errorf("expected 1 formula using symbol 'a', got %d", len(result.Fmlas))
	}
	if !result.Fmlas[0].Equal(a) {
		t.Error("filtered formula should be 'a'")
	}
}

// --- AST utility tests ---

func TestSymbolsAST(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	fmla := &lg.And{Terms: []lg.Node{a, b}}
	syms := SymbolsAST(fmla)
	if len(syms) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(syms))
	}
}

func TestUsedSymbolsAST(t *testing.T) {
	f := mkFuncConst("f", 1)
	x := mkVar("X")
	app := &lg.Apply{Func: f, Terms: []lg.Node{x}}
	syms := UsedSymbolsAST(app)
	if _, ok := syms[lg.Key(f)]; !ok {
		t.Error("should contain function symbol f")
	}
}

func TestVariablesAST(t *testing.T) {
	x := mkVar("X")
	y := mkVar("Y")
	fmla := &lg.And{Terms: []lg.Node{x, y}}
	vars := VariablesAST(fmla)
	if len(vars) != 2 {
		t.Errorf("expected 2 variables, got %d", len(vars))
	}
}

func TestVariablesASTSkipsBound(t *testing.T) {
	x := mkBoolVar("X")
	y := mkBoolVar("Y")
	body := &lg.And{Terms: []lg.Node{x, y}}
	fa := &lg.ForAll{Variables: []*lg.Var{x}, Body: body}
	vars := VariablesAST(fa)
	// Only Y should be free
	if len(vars) != 1 {
		t.Errorf("expected 1 free variable, got %d", len(vars))
	}
}

func TestUsedVariablesAST(t *testing.T) {
	x := mkVar("X")
	fmla := &lg.And{Terms: []lg.Node{x}}
	vars := UsedVariablesAST(fmla)
	if len(vars) != 1 {
		t.Errorf("expected 1 variable, got %d", len(vars))
	}
}

func TestSubstituteConstantsAST(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	fmla := &lg.And{Terms: []lg.Node{a}}
	result := SubstituteConstantsAST(fmla, map[string]lg.Node{"a": b})
	and, ok := result.(*lg.And)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	if !and.Terms[0].Equal(b) {
		t.Error("a should be substituted with b")
	}
}

func TestRenameAST(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	fmla := &lg.And{Terms: []lg.Node{a}}
	result := RenameAST(fmla, map[string]*lg.Const{"a": b})
	and, ok := result.(*lg.And)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	if !and.Terms[0].Equal(b) {
		t.Error("a should be renamed to b")
	}
}

func TestNegate(t *testing.T) {
	a := mkConst("a")
	notA := Negate(a)
	not, ok := notA.(*lg.Not)
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

func TestIsTrue(t *testing.T) {
	if !IsTrue(lg.True) {
		t.Error("True should be true")
	}
	if IsTrue(lg.False) {
		t.Error("False should not be true")
	}
}

func TestIsFalse(t *testing.T) {
	if !IsFalse(lg.False) {
		t.Error("False should be false")
	}
	if IsFalse(lg.True) {
		t.Error("True should not be false")
	}
}

func TestSymPlaceholders(t *testing.T) {
	f := mkFuncConst("f", 2)
	phs := SymPlaceholders(f)
	if len(phs) != 2 {
		t.Errorf("expected 2 placeholders, got %d", len(phs))
	}
	if phs[0].Name != "V0" || phs[1].Name != "V1" {
		t.Errorf("expected V0, V1, got %s, %s", phs[0].Name, phs[1].Name)
	}
}

func TestSymPlaceholdersNoArgs(t *testing.T) {
	c := mkConst("c")
	phs := SymPlaceholders(c)
	if len(phs) != 0 {
		t.Errorf("expected 0 placeholders, got %d", len(phs))
	}
}

func TestSymInst(t *testing.T) {
	f := mkFuncConst("f", 1)
	inst := SymInst(f)
	app, ok := inst.(*lg.Apply)
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

func TestSymInstNoArgs(t *testing.T) {
	c := mkConst("c")
	inst := SymInst(c)
	if !inst.Equal(c) {
		t.Error("0-arity sym inst should return the symbol itself")
	}
}

func TestEqAtom(t *testing.T) {
	x := mkVar("X")
	y := mkVar("Y")
	eq := EqAtom(x, y)
	e, ok := eq.(*lg.Eq)
	if !ok {
		t.Fatalf("expected Eq, got %T", eq)
	}
	if !e.T1.Equal(x) || !e.T2.Equal(y) {
		t.Error("EqAtom should wrap x and y")
	}
}

func TestEqLit(t *testing.T) {
	x := mkVar("X")
	y := mkVar("Y")
	lit := EqLit(x, y)
	if lit.Polarity != 1 {
		t.Error("EqLit should have positive polarity")
	}
	e, ok := lit.Atom.(*lg.Eq)
	if !ok {
		t.Fatalf("expected Eq atom, got %T", lit.Atom)
	}
	if !e.T1.Equal(x) || !e.T2.Equal(y) {
		t.Error("EqLit should have x == y as atom")
	}
}

func TestCollectAndList(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := mkConst("c")
	nested := &lg.And{Terms: []lg.Node{a, &lg.And{Terms: []lg.Node{b, c}}}}
	result := CollectAndList([]lg.Node{nested})
	if len(result) != 3 {
		t.Errorf("expected 3 after flattening, got %d", len(result))
	}
}

func TestCollectOr(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := mkConst("c")
	nested := &lg.Or{Terms: []lg.Node{a, &lg.Or{Terms: []lg.Node{b, c}}}}
	result := CollectOr(nested)
	if len(result) != 3 {
		t.Errorf("expected 3 after flattening, got %d", len(result))
	}
}

func TestDropUniversals(t *testing.T) {
	x := mkBoolVar("X")
	body := x
	fa := &lg.ForAll{Variables: []*lg.Var{x}, Body: body}
	result := DropUniversals(fa)
	if _, ok := result.(*lg.ForAll); ok {
		t.Error("ForAll should be stripped")
	}
	if !result.Equal(x) {
		t.Error("should return body")
	}
}

func TestDropUniversalsNested(t *testing.T) {
	x := mkBoolVar("X")
	y := mkBoolVar("Y")
	body := &lg.And{Terms: []lg.Node{x, y}}
	inner := &lg.ForAll{Variables: []*lg.Var{y}, Body: body}
	outer := &lg.ForAll{Variables: []*lg.Var{x}, Body: inner}
	result := DropUniversals(outer)
	// Should strip both ForAlls
	if _, ok := result.(*lg.ForAll); ok {
		t.Error("all ForAlls should be stripped")
	}
}

func TestIsGroundAST(t *testing.T) {
	a := mkConst("a")
	if !IsGroundAST(a) {
		t.Error("constant should be ground")
	}
	x := mkVar("X")
	if IsGroundAST(x) {
		t.Error("variable should not be ground")
	}
}

func TestNormalizeFreeVariables(t *testing.T) {
	x := mkBoolVar("X")
	y := mkBoolVar("Y")
	fmla := &lg.And{Terms: []lg.Node{x, y}}
	oldVars, newVars, result := NormalizeFreeVariables(fmla)
	if len(oldVars) != 2 || len(newVars) != 2 {
		t.Fatalf("expected 2 old and 2 new vars, got %d, %d", len(oldVars), len(newVars))
	}
	if newVars[0].Name != "V0" || newVars[1].Name != "V1" {
		t.Errorf("expected V0, V1, got %s, %s", newVars[0].Name, newVars[1].Name)
	}
	and, ok := result.(*lg.And)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	for _, term := range and.Terms {
		v, ok := term.(*lg.Var)
		if !ok {
			t.Fatalf("expected Var, got %T", term)
		}
		if !strings.HasPrefix(v.Name, "V") {
			t.Errorf("expected normalized name starting with V, got %s", v.Name)
		}
	}
}

// --- Rename/Substitute on Clauses tests ---

func TestRenameClauses(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := NewClauses([]lg.Node{a}, nil, nil)
	result := RenameClauses(c, map[string]*lg.Const{"a": b})
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
	if !result.Fmlas[0].Equal(b) {
		t.Error("a should be renamed to b")
	}
}

func TestSubstituteConstantsClauses(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := NewClauses([]lg.Node{a}, nil, nil)
	result := SubstituteConstantsClauses(c, map[string]lg.Node{"a": b})
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
	if !result.Fmlas[0].Equal(b) {
		t.Error("a should be substituted with b")
	}
}

func TestClausesSymbols(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := NewClauses([]lg.Node{a, b}, nil, nil)
	syms := c.Symbols()
	if len(syms) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(syms))
	}
}

func TestClausesToFormula(t *testing.T) {
	x := mkVar("X")
	c := NewClauses([]lg.Node{x}, nil, nil)
	f := c.ToFormula()
	// Should be ForAll X. X (since X is free)
	fa, ok := f.(*lg.ForAll)
	if !ok {
		t.Fatalf("expected ForAll, got %T: %s", f, f)
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
		consts := make([]lg.Node, 0, numConsts)
		for i := 0; i < numConsts; i++ {
			consts = append(consts, mkConst("c"+strings.Repeat("x", i)))
		}

		// Build a nested And
		inner := &lg.And{Terms: consts}
		numExtra := abs(n2)%3 + 1
		outerTerms := []lg.Node{inner}
		for i := 0; i < numExtra; i++ {
			outerTerms = append(outerTerms, mkConst("d"+strings.Repeat("y", i)))
		}
		outer := &lg.And{Terms: outerTerms}

		result := CollectAndList([]lg.Node{outer})
		// Result should have at least as many elements as inner + outer extras
		expectedMin := numConsts + numExtra
		if len(result) < expectedMin {
			t.Errorf("flattening produced %d elements, expected at least %d", len(result), expectedMin)
		}
		// No And should remain at the top level of result
		for _, r := range result {
			if a, ok := r.(*lg.And); ok && len(a.Terms) > 0 {
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
			fmlas := make([]lg.Node, numFmlas)
			for j := 0; j < numFmlas; j++ {
				fmlas[j] = mkConst("f" + strings.Repeat("z", (i+j)%5))
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
