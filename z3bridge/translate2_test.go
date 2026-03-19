// Tests for translate.go bug fixes: quantifier constraints callback,
// variable naming with :sort_name suffix.
package z3bridge

import (
	"testing"

	"github.com/glycerine/goivy/logic"
)

// TestTranslateVariable_SortSuffix checks that a Variable is translated
// to a Z3 const named "name:sortName".
func TestTranslateVariable_SortSuffix(t *testing.T) {
	tr := NewTranslator()

	sort := &logic.UninterpretedSort{Name: "node"}
	v, err := logic.NewVariable("X", sort)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}

	z3v, err := tr.Translate(v)
	if err != nil {
		t.Fatalf("Translate variable: %v", err)
	}

	name := z3v.String()
	if name != "X:node" {
		t.Fatalf("variable Z3 name = %q, want %q", name, "X:node")
	}
}

// TestTranslateVariable_BoolSort checks Bool variable naming.
func TestTranslateVariable_BoolSort(t *testing.T) {
	tr := NewTranslator()

	v, err := logic.NewVariable("b", logic.Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}

	z3v, err := tr.Translate(v)
	if err != nil {
		t.Fatalf("Translate bool variable: %v", err)
	}

	name := z3v.String()
	if name != "b:Bool" {
		t.Fatalf("bool variable Z3 name = %q, want %q", name, "b:Bool")
	}
}

// TestTranslateSymbol_NoSortSuffix checks that a Symbol (constant) does NOT
// get the :sort suffix — only variables do.
func TestTranslateSymbol_NoSortSuffix(t *testing.T) {
	tr := NewTranslator()

	sort := &logic.UninterpretedSort{Name: "node"}
	sym := logic.NewSymbol("c", sort)

	z3s, err := tr.Translate(sym)
	if err != nil {
		t.Fatalf("Translate symbol: %v", err)
	}

	name := z3s.String()
	if name != "c" {
		t.Fatalf("symbol Z3 name = %q, want %q", name, "c")
	}
}

// TestQuantConstraints_Callback checks that the QuantConstraints callback
// is invoked and its results wrap the quantifier body.
func TestQuantConstraints_Callback(t *testing.T) {
	tr := NewTranslator()

	callCount := 0
	tr.QuantConstraints = func(v *logic.Variable, z3Var Expr) []Expr {
		callCount++
		// Add constraint: z3Var >= 0
		return []Expr{tr.Ctx.Le(tr.Ctx.IntVal(0), z3Var)}
	}

	sort := &logic.UninterpretedSort{Name: "nat"}
	// Register the sort as IntSort for translation
	tr.sorts["nat"] = tr.Ctx.IntSort()

	x, _ := logic.NewVariable("X", sort)

	// Translate a simple predicate P(X)
	pSort, _ := logic.NewFunctionSort(sort, logic.Boolean)
	p := logic.NewSymbol("P", pSort)
	pApp := &logic.Apply{Func: p, Terms: []logic.Expr{x}}

	fmla := &logic.ForAll{Variables: []*logic.Variable{x}, Body: pApp}
	z3expr, err := tr.Translate(fmla)
	if err != nil {
		t.Fatalf("Translate ForAll: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("QuantConstraints called %d times, want 1", callCount)
	}

	// The result should be a ForAll whose body includes Implies
	str := z3expr.String()
	t.Logf("ForAll result: %s", str)

	// Semantic check: ForAll(X, 0<=X => P(X)) should not require P(-1)
	slv := tr.Ctx.NewSolver()
	slv.Assert(z3expr)
	pDecl := tr.Ctx.Function("P", []Sort{tr.Ctx.IntSort()}, tr.Ctx.BoolSort())
	slv.Assert(tr.Ctx.Not(pDecl.Apply(tr.Ctx.IntVal(-1))))
	if slv.Check() == Unsat {
		t.Fatal("ForAll(X, 0<=X => P(X)) + Not(P(-1)) should be SAT")
	}
}

// TestQuantConstraints_Exists checks that Exists wraps with And.
func TestQuantConstraints_Exists(t *testing.T) {
	tr := NewTranslator()

	tr.QuantConstraints = func(v *logic.Variable, z3Var Expr) []Expr {
		return []Expr{tr.Ctx.Le(tr.Ctx.IntVal(0), z3Var)}
	}

	sort := &logic.UninterpretedSort{Name: "nat"}
	tr.sorts["nat"] = tr.Ctx.IntSort()

	x, _ := logic.NewVariable("X", sort)
	pSort, _ := logic.NewFunctionSort(sort, logic.Boolean)
	p := logic.NewSymbol("P", pSort)
	pApp := &logic.Apply{Func: p, Terms: []logic.Expr{x}}

	fmla := &logic.Exists{Variables: []*logic.Variable{x}, Body: pApp}
	z3expr, err := tr.Translate(fmla)
	if err != nil {
		t.Fatalf("Translate Exists: %v", err)
	}

	// Exists(X, And(0<=X, P(X))) with P only true at -1 should be UNSAT
	slv := tr.Ctx.NewSolver()
	slv.Assert(z3expr)
	xConst := tr.Ctx.Const("X:nat", tr.Ctx.IntSort())
	pDecl := tr.Ctx.Function("P", []Sort{tr.Ctx.IntSort()}, tr.Ctx.BoolSort())
	slv.Assert(tr.Ctx.ForAll(
		[]Expr{xConst},
		tr.Ctx.Eq(pDecl.Apply(xConst), tr.Ctx.Eq(xConst, tr.Ctx.IntVal(-1))),
	))
	if slv.Check() != Unsat {
		t.Fatal("Exists(X, 0<=X && P(X)) with P(-1 only) should be UNSAT")
	}
}

// TestQuantConstraints_NilCallback checks no panic when callback is nil.
func TestQuantConstraints_NilCallback(t *testing.T) {
	tr := NewTranslator()
	// QuantConstraints is nil by default

	sort := &logic.UninterpretedSort{Name: "T"}
	x, _ := logic.NewVariable("X", sort)
	body := x // trivial body

	fmla := &logic.ForAll{Variables: []*logic.Variable{x}, Body: body}
	_, err := tr.Translate(fmla)
	if err != nil {
		t.Fatalf("Translate ForAll with nil callback: %v", err)
	}
}
