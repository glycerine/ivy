package goivy

import (
	"testing"
)

// TestOrClausesIntBareIte verifies that orClausesInt produces bare Ite (not
// SimpIte-simplified) in definition merging, matching Python's or_clauses_int.
func TestOrClausesIntBareIte(t *testing.T) {
	sym := moduleMkConst("x")
	rhs1 := &LogicAnd{} // true
	rhs2 := moduleMkConst("y")

	def1 := NewIvyDefinition(sym, rhs1)
	def2 := NewIvyDefinition(sym, rhs2)

	cls1 := NewClauses(nil, []*IvyDefinition{def1}, nil)
	cls2 := NewClauses(nil, []*IvyDefinition{def2}, nil)

	// Use OrClausesTyped which goes through orClausesInt
	result := OrClausesTyped(cls1, cls2)

	// Find the definition for sym
	found := false
	for _, d := range result.Defs {
		if d.Lhs.Equal(sym) {
			// The RHS should be a bare Ite, NOT an Or (which SimpIte would produce
			// since rhs1 is true: SimpIte(v, true, y) → Or(v, y))
			if _, ok := d.Rhs.(*LogicIte); !ok {
				t.Errorf("expected bare Ite in def RHS, got %T: %v", d.Rhs, d.Rhs)
			}
			found = true
		}
	}
	if !found {
		t.Error("expected to find definition for sym 'x' in result")
	}
}

// TestElimDeadDefinitionsSkolemRename verifies that skolem captured definitions
// are renamed (not eliminated), matching Python's elim_dead_definitions.
func TestElimDeadDefinitionsSkolemRename(t *testing.T) {
	// Create a skolem symbol (starts with __)
	skolem := NewConst("__sk", Boolean)
	nonSkolem := moduleMkConst("p")

	// cls1 has a definition for __sk, cls2 does not → __sk is captured
	def1 := NewIvyDefinition(skolem, moduleMkConst("a"))
	cls1 := NewClauses(nil, []*IvyDefinition{def1}, nil)
	cls2 := NewClauses([]Expr{moduleMkConst("b")}, nil, nil)

	rn := newTestRenamer()
	result := elimDeadDefinitions(rn, []*Clauses{cls1, cls2})

	// The skolem should be RENAMED, not eliminated.
	// cls1 should still have a definition (with a renamed symbol).
	if len(result[0].Defs) == 0 {
		t.Error("skolem definition should be renamed, not eliminated")
	}

	// The definition symbol should have a different name (renamed by rn)
	if len(result[0].Defs) > 0 {
		defSym := result[0].Defs[0].Defines()
		if c, ok := defSym.(*Const); ok {
			if c.Name == "__sk" {
				t.Error("skolem should have been renamed to a fresh name")
			}
		}
	}

	// cls2 should NOT have the skolem definition (it never did)
	_ = nonSkolem
}

// TestElimDeadDefinitionsNonSkolemElim verifies that non-skolem captured
// definitions are eliminated (converted to constraints), matching Python.
func TestElimDeadDefinitionsNonSkolemElim(t *testing.T) {
	sym := moduleMkConst("p")
	def1 := NewIvyDefinition(sym, moduleMkConst("a"))
	cls1 := NewClauses(nil, []*IvyDefinition{def1}, nil)
	cls2 := NewClauses(nil, nil, nil)

	rn := newTestRenamer()
	result := elimDeadDefinitions(rn, []*Clauses{cls1, cls2})

	// Non-skolem should be eliminated: no definition, but a formula constraint
	if len(result[0].Defs) > 0 {
		t.Error("non-skolem captured def should be eliminated, not kept")
	}
	if len(result[0].Fmlas) == 0 {
		t.Error("non-skolem captured def should be converted to constraint formula")
	}
}

// TestOrClausesOneFalse verifies that OrClausesTyped with one false branch
// returns the non-false branch (no Tseitin encoding).
func TestOrClausesOneFalse(t *testing.T) {
	fmla := moduleMkConst("p")
	nonFalse := NewClauses([]Expr{fmla}, nil, nil)
	falseCls := FalseClauses(nil)

	result := OrClausesTyped(falseCls, nonFalse)

	// Should contain p in fmlas (returned directly, no Tseitin __ts0_ variables)
	if len(result.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(result.Fmlas))
	}
	if !result.Fmlas[0].Equal(fmla) {
		t.Errorf("expected formula p, got %v", result.Fmlas[0])
	}
}

// TestOrClausesBothNonFalse verifies Tseitin encoding with 2 non-false branches.
func TestOrClausesBothNonFalse(t *testing.T) {
	cls1 := NewClauses([]Expr{moduleMkConst("p")}, nil, nil)
	cls2 := NewClauses([]Expr{moduleMkConst("q")}, nil, nil)

	result := OrClausesTyped(cls1, cls2)

	// Should have at least 3 formulas:
	// 1. Or(v1, v2)
	// 2. Or(Not(v1), p)
	// 3. Or(Not(v2), q)
	if len(result.Fmlas) < 3 {
		t.Fatalf("expected at least 3 formulas from Tseitin encoding, got %d", len(result.Fmlas))
	}

	// First formula should be an Or with 2 terms
	if or, ok := result.Fmlas[0].(*LogicOr); !ok {
		t.Errorf("first formula should be Or, got %T", result.Fmlas[0])
	} else if len(or.Terms) != 2 {
		t.Errorf("Or should have 2 terms, got %d", len(or.Terms))
	}
}

// TestIteClausesSimpIte verifies that IteClauses uses simp_ite for def merging.
func TestIteClausesSimpIte(t *testing.T) {
	sym := moduleMkConst("x")
	rhs1 := &LogicAnd{} // true
	rhs2 := moduleMkConst("y")

	def1 := NewIvyDefinition(sym, rhs1)
	def2 := NewIvyDefinition(sym, rhs2)

	cls1 := NewClauses(nil, []*IvyDefinition{def1}, nil)
	cls2 := NewClauses(nil, []*IvyDefinition{def2}, nil)
	cond := moduleMkConst("c")

	result := IteClauses(cond, cls1, cls2)

	// Find the definition for sym — should be simplified via SimpIte
	for _, d := range result.Defs {
		if d.Lhs.Equal(sym) {
			// SimpIte(v, true, y) → Or(v, y). So RHS should be Or, not Ite.
			if _, ok := d.Rhs.(*LogicIte); ok {
				t.Error("IteClauses should use SimpIte which simplifies Ite(v, true, y) to Or(v, y)")
			}
		}
	}

	// Should have a definition for the condition variable
	hasCondDef := false
	for _, d := range result.Defs {
		if !d.Lhs.Equal(sym) {
			hasCondDef = true
		}
	}
	if !hasCondDef {
		t.Error("IteClauses should add a definition for the condition variable")
	}
}

// TestNewClausesDropUniversals verifies that NewClauses strips ForAll from fmlas.
func TestNewClausesDropUniversals(t *testing.T) {
	v, _ := NewVariable("V", Boolean)
	// ForAll(V, Or()) = ForAll(V, false)
	fmla := &ForAll{Variables: []*LogicVariable{v}, Body: &LogicOr{}}

	cls := NewClauses([]Expr{fmla}, nil, nil)

	// After dropUniversals, ForAll is stripped, exposing Or() = false
	if !cls.IsFalse() {
		t.Error("NewClauses should strip ForAll, exposing Or() as false")
	}
}

// TestNewClausesCollectAndList verifies nested And flattening.
func TestNewClausesCollectAndList(t *testing.T) {
	a := moduleMkConst("a")
	b := moduleMkConst("b")
	c := moduleMkConst("c")

	// And(a, And(b, c))
	inner := &LogicAnd{Terms: []Expr{b, c}}
	outer := &LogicAnd{Terms: []Expr{a, inner}}

	cls := NewClauses([]Expr{outer}, nil, nil)

	if len(cls.Fmlas) != 3 {
		t.Fatalf("expected 3 flattened formulas, got %d: %v", len(cls.Fmlas), cls.Fmlas)
	}
	if !cls.Fmlas[0].Equal(a) || !cls.Fmlas[1].Equal(b) || !cls.Fmlas[2].Equal(c) {
		t.Errorf("expected [a, b, c], got %v", cls.Fmlas)
	}
}

// TestNewClausesDropUniversalsNot verifies Not(Exists(...)) handling.
func TestNewClausesDropUniversalsNot(t *testing.T) {
	v, _ := NewVariable("V", Boolean)
	body := moduleMkConst("p")
	// Not(Exists(V, p)) → should strip Exists under Not
	fmla := &LogicNot{Body: &LogicExists{Variables: []*LogicVariable{v}, Body: body}}

	cls := NewClauses([]Expr{fmla}, nil, nil)

	// After dropUniversals on Not(Exists(...)):
	// → Not(dropExistentials(Exists(V, p)))
	// → Not(p)
	if len(cls.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(cls.Fmlas))
	}
	if not, ok := cls.Fmlas[0].(*LogicNot); !ok {
		t.Errorf("expected Not, got %T", cls.Fmlas[0])
	} else if !not.Body.Equal(body) {
		t.Errorf("expected Not(p), got Not(%v)", not.Body)
	}
}

// TestIteClausesIntAnnotation verifies that iteClausesInt preserves
// annotations, matching Python: annot = None if a0 is None or a1 is None else a0.ite(v,a1).
func TestIteClausesIntAnnotation(t *testing.T) {
	a0 := EmptyAnnotation{}
	a1 := EmptyAnnotation{}
	cls1 := NewClauses([]Expr{moduleMkConst("p")}, nil, a0)
	cls2 := NewClauses([]Expr{moduleMkConst("q")}, nil, a1)
	cond := moduleMkConst("c")

	result := IteClauses(cond, cls1, cls2)

	ite, ok := result.Annot.(*IteAnnotation)
	if !ok {
		t.Fatalf("expected IteAnnotation, got %T", result.Annot)
	}
	if ite.Cond == nil {
		t.Fatal("expected IteAnnotation condition")
	}
	if _, ok := ite.ThenB.(EmptyAnnotation); !ok {
		t.Errorf("expected then annotation to be EmptyAnnotation, got %T", ite.ThenB)
	}
	if _, ok := ite.ElseB.(EmptyAnnotation); !ok {
		t.Errorf("expected else annotation to be EmptyAnnotation, got %T", ite.ElseB)
	}

	// With one nil annotation, result should be nil
	cls3 := NewClauses([]Expr{moduleMkConst("r")}, nil, nil)
	result2 := IteClauses(cond, cls1, cls3)
	if result2.Annot != nil {
		t.Errorf("expected nil annotation when one arg annot is nil, got %v", result2.Annot)
	}
}

// TestElimDeadDefinitionsPerArgRename verifies that each arg gets different
// fresh names for captured skolem symbols, matching Python's per-arg renaming.
// Python: args = [rename_symbols(rn, arg, to_rename) for arg in args]
// Each call generates DIFFERENT fresh names because rn is stateful.
func TestElimDeadDefinitionsPerArgRename(t *testing.T) {
	skolem := NewConst("__sk", Boolean)

	// cls1 defines __sk; cls2 does NOT define __sk but references it in a formula.
	// This makes __sk "captured" (defined in some but not all args).
	def1 := NewIvyDefinition(skolem, moduleMkConst("a"))
	cls1 := NewClauses(nil, []*IvyDefinition{def1}, nil)
	cls2 := NewClauses([]Expr{skolem}, nil, nil) // references __sk in fmla

	rn := newTestRenamer()
	result := elimDeadDefinitions(rn, []*Clauses{cls1, cls2})

	// cls1 should have its definition renamed (not __sk anymore)
	if len(result[0].Defs) == 0 {
		t.Fatal("cls1 should retain its (renamed) definition")
	}
	defName := result[0].Defs[0].Defines().(*Const).Name
	if defName == "__sk" {
		t.Error("skolem definition should have been renamed from __sk")
	}

	// cls2's formula should reference a different renamed version of __sk.
	// With per-arg renaming, cls1 and cls2 get DIFFERENT fresh names.
	if len(result[1].Fmlas) == 0 {
		t.Fatal("cls2 should still have its formula")
	}
	fmlaConst, ok := result[1].Fmlas[0].(*Const)
	if !ok {
		t.Fatalf("cls2 formula should be a renamed Const, got %T", result[1].Fmlas[0])
	}
	if fmlaConst.Name == "__sk" {
		t.Error("cls2's reference to __sk should have been renamed")
	}
	if fmlaConst.Name == defName {
		t.Errorf("per-arg rename should produce different names: cls1 def=%q, cls2 fmla=%q (should differ)", defName, fmlaConst.Name)
	}
}

// TestToFormulaCloseEPR verifies that ToFormula distributes ForAll through And,
// matching Python's close_epr behavior.
func TestToFormulaCloseEPR(t *testing.T) {
	x := moduleMkVar("X")
	y := moduleMkVar("Y")
	// Two formulas with different free variables
	cls := NewClauses([]Expr{x, y}, nil, nil)
	f := cls.ToFormula()

	// CloseEPR(And(X, Y)) → And(CloseEPR(X), CloseEPR(Y)) → And(ForAll(X,X), ForAll(Y,Y))
	and, ok := f.(*LogicAnd)
	if !ok {
		t.Fatalf("expected And at top level, got %T", f)
	}
	if len(and.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(and.Terms))
	}
	for i, term := range and.Terms {
		if _, ok := term.(*ForAll); !ok {
			t.Errorf("term %d: expected ForAll, got %T", i, term)
		}
	}
}

// TestNegateClausesPanic verifies that NegateClauses panics on non-universal-first-order input.
func TestNegateClausesPanic(t *testing.T) {
	// Create clauses with a definition (not universal first order)
	sym := moduleMkConst("p")
	def := NewIvyDefinition(sym, moduleMkConst("a"))
	cls := NewClauses(nil, []*IvyDefinition{def}, nil)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for non-universal-first-order input")
		}
	}()
	NegateClauses(cls)
}

// TestDualClausesCustomSkolemizer verifies that a custom skolemizer is applied.
func TestDualClausesCustomSkolemizer(t *testing.T) {
	x := moduleMkVar("X")
	cls := NewClauses([]Expr{x}, nil, nil)

	customPrefix := "@test_"
	skolemizer := func(v *LogicVariable) Expr {
		return NewConst(customPrefix+v.Name, v.VSort)
	}

	result := DualClauses(cls, skolemizer, nil)

	// The result should reference the custom-prefixed skolem
	syms := result.Symbols()
	found := false
	for _, s := range syms.All() {
		if ExprName(s) == customPrefix+"X" {
			found = true
		}
	}
	if !found {
		t.Error("custom skolemizer should produce @test_X, but symbol not found in result")
	}
}

// TestDualFormulaInstantiator verifies that DualFormula conjoins definition
// instances from the instantiator, matching Python's dual_formula.
func TestDualFormulaInstantiator(t *testing.T) {
	p := moduleMkConst("p")

	// With nil instantiator: just negate
	result1 := DualFormula(p, nil, nil)
	if _, ok := result1.(*LogicNot); !ok {
		t.Fatalf("DualFormula(p, nil, nil) should be Not, got %T", result1)
	}

	// With non-nil instantiator: should conjoin instances
	instantiator := func(gts []Expr) *Clauses {
		return NewClauses([]Expr{moduleMkConst("inst")}, nil, nil)
	}
	result2 := DualFormula(p, nil, instantiator)
	and, ok := result2.(*LogicAnd)
	if !ok {
		t.Fatalf("DualFormula with instantiator should produce And, got %T", result2)
	}
	if len(and.Terms) != 2 {
		t.Fatalf("expected 2 And terms, got %d", len(and.Terms))
	}
	// First term should be Not(p), second should be the instantiated formula
	if _, ok := and.Terms[0].(*LogicNot); !ok {
		t.Errorf("first And term should be Not, got %T", and.Terms[0])
	}
}

// TestSkolemizeFormulaInstantiator verifies that SkolemizeFormula conjoins
// definition instances from the instantiator, matching Python's skolemize_formula.
func TestSkolemizeFormulaInstantiator(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	body := v
	ex, _ := NewExists([]*LogicVariable{v}, body)

	// With nil instantiator: just skolemize
	result1 := SkolemizeFormula(ex, nil, nil)
	if c, ok := result1.(*Const); !ok || c.Name != "__sk__X" {
		t.Fatalf("SkolemizeFormula without instantiator: expected __sk__X, got %T %v", result1, result1)
	}

	// With non-nil instantiator: should conjoin instances
	instantiator := func(gts []Expr) *Clauses {
		return NewClauses([]Expr{moduleMkConst("inst")}, nil, nil)
	}
	result2 := SkolemizeFormula(ex, nil, instantiator)
	and, ok := result2.(*LogicAnd)
	if !ok {
		t.Fatalf("SkolemizeFormula with instantiator should produce And, got %T", result2)
	}
	if len(and.Terms) != 2 {
		t.Fatalf("expected 2 And terms, got %d", len(and.Terms))
	}
}

// TestTaggedOrClausesPrefix verifies that TaggedOrClauses uses __to0 prefix
// and does NOT filter false branches, matching Python's tagged_or_clauses.
func TestTaggedOrClausesPrefix(t *testing.T) {
	cls1 := NewClauses([]Expr{moduleMkConst("p")}, nil, nil)
	cls2 := NewClauses([]Expr{moduleMkConst("q")}, nil, nil)

	result := TaggedOrClauses("tag", cls1, cls2)

	// Should have the Or disjunction as first formula
	if len(result.Fmlas) < 1 {
		t.Fatal("expected at least 1 formula")
	}
	or, ok := result.Fmlas[0].(*LogicOr)
	if !ok {
		t.Fatalf("first formula should be Or, got %T", result.Fmlas[0])
	}
	if len(or.Terms) != 2 {
		t.Fatalf("expected 2 Or terms, got %d", len(or.Terms))
	}
	// Tag variables should start with __to0, not __ts0
	for _, term := range or.Terms {
		c, ok := term.(*Const)
		if !ok {
			t.Errorf("Or term should be Const, got %T", term)
			continue
		}
		if len(c.Name) < 5 || c.Name[:5] != "__to0" {
			t.Errorf("tag variable should start with __to0, got %q", c.Name)
		}
	}
}

// TestTaggedOrClausesNoFalseFilter verifies that false branches are NOT filtered.
func TestTaggedOrClausesNoFalseFilter(t *testing.T) {
	cls1 := FalseClauses(nil)
	cls2 := NewClauses([]Expr{moduleMkConst("p")}, nil, nil)
	cls3 := NewClauses([]Expr{moduleMkConst("q")}, nil, nil)

	result := TaggedOrClauses("tag", cls1, cls2, cls3)

	// Should have Or with 3 terms (false branch NOT filtered)
	if len(result.Fmlas) < 1 {
		t.Fatal("expected at least 1 formula")
	}
	or, ok := result.Fmlas[0].(*LogicOr)
	if !ok {
		t.Fatalf("first formula should be Or, got %T", result.Fmlas[0])
	}
	if len(or.Terms) != 3 {
		t.Errorf("expected 3 Or terms (false branch not filtered), got %d", len(or.Terms))
	}
}

// TestExistsQuantClausesMapSimple verifies basic equivalence class merging
// and unconditional renaming.
func TestExistsQuantClausesMapSimple(t *testing.T) {
	a := NewConst("a", Boolean)
	c := NewConst("c", Boolean)

	// Definition: a = c
	def := NewIvyDefinition(a, c)
	cls := NewClauses([]Expr{moduleMkConst("p")}, []*IvyDefinition{def}, nil)

	syms := []*Const{a}
	map1, resultCls := ExistsQuantClausesMap(syms, cls)

	// 'a' should be mapped to a fresh name (unconditionally renamed)
	aMapping, ok := map1[Key(a)]
	if !ok {
		t.Fatal("expected mapping for symbol 'a'")
	}
	if aMapping.Name == "a" {
		t.Error("symbol 'a' should be renamed to a fresh name")
	}

	// The definition a=c should have been eliminated by eqcm_upd
	if len(resultCls.Defs) > 0 {
		t.Errorf("expected definition to be eliminated, got %d defs", len(resultCls.Defs))
	}
}

// --- test helpers ---

func newTestRenamer() *UniqueRenamer {
	return NewUniqueRenamer("__test", nil)
}
