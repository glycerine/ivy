package module

import (
	"testing"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
)

// TestOrClausesIntBareIte verifies that orClausesInt produces bare Ite (not
// SimpIte-simplified) in definition merging, matching Python's or_clauses_int.
func TestOrClausesIntBareIte(t *testing.T) {
	sym := mkConst("x")
	rhs1 := &lg.And{} // true
	rhs2 := mkConst("y")

	def1 := il.NewDefinition(sym, rhs1)
	def2 := il.NewDefinition(sym, rhs2)

	cls1 := NewClauses(nil, []*il.Definition{def1}, nil)
	cls2 := NewClauses(nil, []*il.Definition{def2}, nil)

	// Use OrClausesTyped which goes through orClausesInt
	result := OrClausesTyped(cls1, cls2)

	// Find the definition for sym
	found := false
	for _, d := range result.Defs {
		if d.Lhs.Equal(sym) {
			// The RHS should be a bare Ite, NOT an Or (which SimpIte would produce
			// since rhs1 is true: SimpIte(v, true, y) → Or(v, y))
			if _, ok := d.Rhs.(*lg.Ite); !ok {
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
	skolem := lg.NewConst("__sk", lg.Boolean)
	nonSkolem := mkConst("p")

	// cls1 has a definition for __sk, cls2 does not → __sk is captured
	def1 := il.NewDefinition(skolem, mkConst("a"))
	cls1 := NewClauses(nil, []*il.Definition{def1}, nil)
	cls2 := NewClauses([]lg.Expr{mkConst("b")}, nil, nil)

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
		if c, ok := defSym.(*lg.Const); ok {
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
	sym := mkConst("p")
	def1 := il.NewDefinition(sym, mkConst("a"))
	cls1 := NewClauses(nil, []*il.Definition{def1}, nil)
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
	fmla := mkConst("p")
	nonFalse := NewClauses([]lg.Expr{fmla}, nil, nil)
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
	cls1 := NewClauses([]lg.Expr{mkConst("p")}, nil, nil)
	cls2 := NewClauses([]lg.Expr{mkConst("q")}, nil, nil)

	result := OrClausesTyped(cls1, cls2)

	// Should have at least 3 formulas:
	// 1. Or(v1, v2)
	// 2. Or(Not(v1), p)
	// 3. Or(Not(v2), q)
	if len(result.Fmlas) < 3 {
		t.Fatalf("expected at least 3 formulas from Tseitin encoding, got %d", len(result.Fmlas))
	}

	// First formula should be an Or with 2 terms
	if or, ok := result.Fmlas[0].(*lg.Or); !ok {
		t.Errorf("first formula should be Or, got %T", result.Fmlas[0])
	} else if len(or.Terms) != 2 {
		t.Errorf("Or should have 2 terms, got %d", len(or.Terms))
	}
}

// TestIteClausesSimpIte verifies that IteClauses uses simp_ite for def merging.
func TestIteClausesSimpIte(t *testing.T) {
	sym := mkConst("x")
	rhs1 := &lg.And{} // true
	rhs2 := mkConst("y")

	def1 := il.NewDefinition(sym, rhs1)
	def2 := il.NewDefinition(sym, rhs2)

	cls1 := NewClauses(nil, []*il.Definition{def1}, nil)
	cls2 := NewClauses(nil, []*il.Definition{def2}, nil)
	cond := mkConst("c")

	result := IteClauses(cond, cls1, cls2)

	// Find the definition for sym — should be simplified via SimpIte
	for _, d := range result.Defs {
		if d.Lhs.Equal(sym) {
			// SimpIte(v, true, y) → Or(v, y). So RHS should be Or, not Ite.
			if _, ok := d.Rhs.(*lg.Ite); ok {
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
	v, _ := lg.NewVariable("V", lg.Boolean)
	// ForAll(V, Or()) = ForAll(V, false)
	fmla := &lg.ForAll{Variables: []*lg.Variable{v}, Body: &lg.Or{}}

	cls := NewClauses([]lg.Expr{fmla}, nil, nil)

	// After dropUniversals, ForAll is stripped, exposing Or() = false
	if !cls.IsFalse() {
		t.Error("NewClauses should strip ForAll, exposing Or() as false")
	}
}

// TestNewClausesCollectAndList verifies nested And flattening.
func TestNewClausesCollectAndList(t *testing.T) {
	a := mkConst("a")
	b := mkConst("b")
	c := mkConst("c")

	// And(a, And(b, c))
	inner := &lg.And{Terms: []lg.Expr{b, c}}
	outer := &lg.And{Terms: []lg.Expr{a, inner}}

	cls := NewClauses([]lg.Expr{outer}, nil, nil)

	if len(cls.Fmlas) != 3 {
		t.Fatalf("expected 3 flattened formulas, got %d: %v", len(cls.Fmlas), cls.Fmlas)
	}
	if !cls.Fmlas[0].Equal(a) || !cls.Fmlas[1].Equal(b) || !cls.Fmlas[2].Equal(c) {
		t.Errorf("expected [a, b, c], got %v", cls.Fmlas)
	}
}

// TestNewClausesDropUniversalsNot verifies Not(Exists(...)) handling.
func TestNewClausesDropUniversalsNot(t *testing.T) {
	v, _ := lg.NewVariable("V", lg.Boolean)
	body := mkConst("p")
	// Not(Exists(V, p)) → should strip Exists under Not
	fmla := &lg.Not{Body: &lg.Exists{Variables: []*lg.Variable{v}, Body: body}}

	cls := NewClauses([]lg.Expr{fmla}, nil, nil)

	// After dropUniversals on Not(Exists(...)):
	// → Not(dropExistentials(Exists(V, p)))
	// → Not(p)
	if len(cls.Fmlas) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(cls.Fmlas))
	}
	if not, ok := cls.Fmlas[0].(*lg.Not); !ok {
		t.Errorf("expected Not, got %T", cls.Fmlas[0])
	} else if !not.Body.Equal(body) {
		t.Errorf("expected Not(p), got Not(%v)", not.Body)
	}
}

// --- test helpers ---

func newTestRenamer() *iu.UniqueRenamer {
	return iu.NewUniqueRenamer("__test", nil)
}
