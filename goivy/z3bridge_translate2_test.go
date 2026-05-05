// Tests for translate.go bug fixes: quantifier constraints callback,
// variable naming with :sort_name suffix.
package goivy

import (
	"strings"
	"testing"
)

// TestTranslateVariable_SortSuffix checks that a Variable is translated
// to a Z3 const named "name:sortName" (Z3 may quote it as |name:sort|).
func TestTranslateVariable_SortSuffix(t *testing.T) {
	//vv("TestTranslateVariable_SortSuffix start")
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	sort := &UninterpretedSort{Name: "node"}
	v, err := NewVariable("X", sort)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}

	z3v, err := tr.Translate(v)
	if err != nil {
		t.Fatalf("Translate variable: %v", err)
	}

	name := z3v.String()
	// Z3 may quote identifiers containing ':' as |X:node|
	stripped := strings.Trim(name, "|")
	if stripped != "X:node" {
		t.Fatalf("variable Z3 name = %q, want %q (possibly quoted)", name, "X:node")
	}
}

// TestTranslateVariable_BoolSort checks Bool variable naming.
func TestTranslateVariable_BoolSort(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	v, err := NewVariable("Flag", Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}

	z3v, err := tr.Translate(v)
	if err != nil {
		t.Fatalf("Translate bool variable: %v", err)
	}

	name := z3v.String()
	stripped := strings.Trim(name, "|")
	if stripped != "Flag:Bool" {
		t.Fatalf("bool variable Z3 name = %q, want %q", name, "Flag:Bool")
	}
}

// TestTranslateSymbol_NoSortSuffix checks that a Symbol (constant) does NOT
// get the :sort suffix — only variables do.
func TestTranslateSymbol_NoSortSuffix(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	sort := &UninterpretedSort{Name: "node"}
	sym := NewConst("c", sort)

	z3s, err := tr.Translate(sym)
	if err != nil {
		t.Fatalf("Translate symbol: %v", err)
	}

	name := z3s.String()
	if name != "c" {
		t.Fatalf("symbol Z3 name = %q, want %q", name, "c")
	}
}

/*
// TestQuantConstraints_Callback checks that the QuantConstraints callback
// is invoked and its results wrap the quantifier body.
func TestQuantConstraints_Callback(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	callCount := 0
	tr.QuantConstraints = func(v *logic.Variable, z3Var Expr) []Expr {
		callCount++
		// The variable has been translated to IntSort (since we registered
		// the sort as IntSort). Add constraint: 0 <= z3Var.
		return []Expr{tr.Ctx.Le(tr.Ctx.IntVal(0), z3Var)}
	}

	// Use an uninterpreted sort but register it as IntSort via LookupNative
	// so the Z3 variable will be of IntSort, matching the Le constraint.
	sort := &logic.UninterpretedSort{Name: "mynat"}
	tr.LookupNative = func(name string, s logic.Sort, kind string) any {
		if kind == "sort" && name == "mynat" {
			return tr.Ctx.IntSort()
		}
		return nil
	}

	x, _ := logic.NewVariable("X", sort)

	pSort, _ := logic.NewFunctionSort(sort, logic.Boolean)
	p := logic.NewConst("P", pSort)
	pApp := logic.MustApply(p, x)

	fmla := &logic.ForAll{Variables: []*logic.Variable{x}, Body: pApp}
	z3expr, err := tr.Translate(fmla)
	if err != nil {
		t.Fatalf("Translate ForAll: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("QuantConstraints called %d times, want 1", callCount)
	}

	t.Logf("ForAll result: %s", z3expr.String())

	// Semantic check: ForAll(X, 0<=X => P(X)) should not require P(-1).
	// We need to use the same P function declaration from the translator.
	slv := tr.Ctx.NewZ3Solver()
	slv.Assert(z3expr)
	// Create a fresh P decl to negate P(-1)
	pDecl := tr.Ctx.Function("P", []Sort{tr.Ctx.IntSort()}, tr.Ctx.BoolSort())
	slv.Assert(tr.Ctx.Not(pDecl.Apply(tr.Ctx.IntVal(-1))))
	if slv.Check() == Unsat {
		t.Fatal("ForAll(X, 0<=X => P(X)) + Not(P(-1)) should be SAT")
	}
}

// TestQuantConstraints_Exists checks that Exists wraps body with And.
func TestQuantConstraints_Exists(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	tr.QuantConstraints = func(v *logic.Variable, z3Var Expr) []Expr {
		return []Expr{tr.Ctx.Le(tr.Ctx.IntVal(0), z3Var)}
	}

	sort := &logic.UninterpretedSort{Name: "mynat"}
	tr.LookupNative = func(name string, s logic.Sort, kind string) any {
		if kind == "sort" && name == "mynat" {
			return tr.Ctx.IntSort()
		}
		return nil
	}

	x, _ := logic.NewVariable("X", sort)
	pSort, _ := logic.NewFunctionSort(sort, logic.Boolean)
	p := logic.NewConst("P", pSort)
	pApp := logic.MustApply(p, x)

	fmla := &logic.Exists{Variables: []*logic.Variable{x}, Body: pApp}
	z3expr, err := tr.Translate(fmla)
	if err != nil {
		t.Fatalf("Translate Exists: %v", err)
	}

	t.Logf("Exists result: %s", z3expr.String())

	// Exists(X, And(0<=X, P(X))) with P only true at -1 should be UNSAT
	slv := tr.Ctx.NewZ3Solver()
	slv.Assert(z3expr)
	xConst := tr.Ctx.Const("|X:mynat|", tr.Ctx.IntSort())
	pDecl := tr.Ctx.Function("P", []Sort{tr.Ctx.IntSort()}, tr.Ctx.BoolSort())
	slv.Assert(tr.Ctx.ForAll(
		[]Expr{xConst},
		tr.Ctx.Eq(pDecl.Apply(xConst), tr.Ctx.Eq(xConst, tr.Ctx.IntVal(-1))),
	))
	if slv.Check() != Unsat {
		t.Fatal("Exists(X, 0<=X && P(X)) with P(-1 only) should be UNSAT")
	}
}
*/

// TestQuantConstraints_NilCallback checks no panic when callback is nil
// and body is a non-Bool expression (should return an error, not panic).
func TestQuantConstraints_NilCallback(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()
	// QuantConstraints is nil by default

	sort := &UninterpretedSort{Name: "T"}
	x, _ := NewVariable("X", sort)
	body := x // non-Bool body — invalid ForAll

	fmla := &ForAll{Variables: []*LogicVariable{x}, Body: body}
	_, err := tr.Translate(fmla)
	if err == nil {
		t.Fatal("Translate ForAll with non-Bool body should return an error, not succeed")
	}
	t.Logf("got expected error: %v", err)
}

// --- parseBfeParams format tests ---

func TestParseBfeParams_BracketFormat(t *testing.T) {
	lo, hi, ok := parseBfeParams("bfe[0][15]")
	if !ok {
		t.Fatal("parseBfeParams failed for bfe[0][15]")
	}
	if lo != 0 || hi != 15 {
		t.Fatalf("expected (0, 15), got (%d, %d)", lo, hi)
	}
}

func TestParseBfeParams_ColonFormat(t *testing.T) {
	lo, hi, ok := parseBfeParams("bfe[0:15]")
	if !ok {
		t.Fatal("parseBfeParams failed for bfe[0:15]")
	}
	if lo != 0 || hi != 15 {
		t.Fatalf("expected (0, 15), got (%d, %d)", lo, hi)
	}
}

func TestParseBfeParams_CommaFormat(t *testing.T) {
	lo, hi, ok := parseBfeParams("bfe[3,7]")
	if !ok {
		t.Fatal("parseBfeParams failed for bfe[3,7]")
	}
	if lo != 3 || hi != 7 {
		t.Fatalf("expected (3, 7), got (%d, %d)", lo, hi)
	}
}
