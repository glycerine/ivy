package z3bridge

import (
	"testing"

	"github.com/glycerine/goivy/logic"
)

func mustFS(t *testing.T, sorts ...logic.Sort) *logic.FunctionSort {
	t.Helper()
	fs, err := logic.NewFunctionSort(sorts...)
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

// --- Low-level Z3 wrapper tests ---

func TestContextCreateDestroy(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
}

func TestBoolOps(t *testing.T) {
	ctx := NewContext()
	a := ctx.BoolVal(true)
	b := ctx.BoolVal(false)
	_ = ctx.And(a, b)
	_ = ctx.Or(a, b)
	_ = ctx.Not(a)
	_ = ctx.Implies(a, b)
	_ = ctx.Iff(a, b)
}

func TestConstAndEq(t *testing.T) {
	ctx := NewContext()
	s := ctx.UninterpretedSort("S")
	x := ctx.Const("x", s)
	y := ctx.Const("y", s)
	eq := ctx.Eq(x, y)
	_ = eq.String()
}

func TestFuncDeclApply(t *testing.T) {
	ctx := NewContext()
	s := ctx.UninterpretedSort("S")
	bs := ctx.BoolSort()
	f := ctx.Function("leq", []Sort{s, s}, bs)
	x := ctx.Const("x", s)
	y := ctx.Const("y", s)
	result := f.Apply(x, y)
	t.Log("f(x,y):", result.String())
}

func TestSolverSat(t *testing.T) {
	ctx := NewContext()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	s := ctx.NewSolver()
	s.Assert(x)
	if s.Check() != Sat {
		t.Error("expected sat")
	}
}

func TestSolverUnsat(t *testing.T) {
	ctx := NewContext()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)
	s := ctx.NewSolver()
	s.Assert(x)
	s.Assert(ctx.Not(x))
	if s.Check() != Unsat {
		t.Error("expected unsat")
	}
}

func TestForAllQuantifier(t *testing.T) {
	ctx := NewContext()
	s := ctx.UninterpretedSort("S")
	x := ctx.Const("x", s)
	eq := ctx.Eq(x, x)
	fa := ctx.ForAll([]Expr{x}, eq)
	t.Log("ForAll:", fa.String())

	solver := ctx.NewSolver()
	solver.Assert(fa)
	if solver.Check() != Sat {
		t.Error("ForAll x. x==x should be sat")
	}
}

func TestExistsQuantifier(t *testing.T) {
	ctx := NewContext()
	s := ctx.UninterpretedSort("S")
	x := ctx.Const("x", s)
	y := ctx.Const("y", s)
	eq := ctx.Eq(x, y)
	ex := ctx.Exists([]Expr{x}, eq)
	t.Log("Exists:", ex.String())

	solver := ctx.NewSolver()
	solver.Assert(ex)
	if solver.Check() != Sat {
		t.Error("Exists x. x==y should be sat")
	}
}

func TestPushPop(t *testing.T) {
	ctx := NewContext()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)

	solver := ctx.NewSolver()
	solver.Assert(x)
	if solver.Check() != Sat {
		t.Error("expected sat before push")
	}

	solver.Push()
	solver.Assert(ctx.Not(x))
	if solver.Check() != Unsat {
		t.Error("expected unsat after contradictory assert")
	}

	solver.Pop()
	if solver.Check() != Sat {
		t.Error("expected sat after pop")
	}
}

func TestModel(t *testing.T) {
	ctx := NewContext()
	bs := ctx.BoolSort()
	x := ctx.Const("x", bs)

	solver := ctx.NewSolver()
	solver.Assert(x)
	if solver.Check() != Sat {
		t.Fatal("expected sat")
	}
	m := solver.Model()
	if m == nil {
		t.Fatal("expected model")
	}
	t.Log("Model:", m.String())

	val, ok := m.Eval(x, true)
	if !ok {
		t.Fatal("eval failed")
	}
	t.Log("x =", val.String())
}

// --- Translation tests ---

func TestTranslateSorts(t *testing.T) {
	tr := NewTranslator()

	_, err := tr.TranslateSort(logic.Boolean)
	if err != nil {
		t.Fatal(err)
	}

	S := &logic.UninterpretedSort{Name: "S"}
	_, err = tr.TranslateSort(S)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTranslateVar(t *testing.T) {
	tr := NewTranslator()
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVar("X", S)

	zx, err := tr.Translate(X)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("X:", zx.String())
}

func TestTranslateEq(t *testing.T) {
	tr := NewTranslator()
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVar("X", S)
	Y, _ := logic.NewVar("Y", S)

	eq, _ := logic.NewEq(X, Y)
	zeq, err := tr.Translate(eq)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Eq:", zeq.String())
}

func TestTranslateApply(t *testing.T) {
	tr := NewTranslator()
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVar("X", S)
	Y, _ := logic.NewVar("Y", S)
	leq := logic.NewConst("leq", mustFS(t, S, S, logic.Boolean))

	app, _ := logic.NewApply(leq, X, Y)
	zapp, err := tr.Translate(app)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("leq(X,Y):", zapp.String())
}

func TestTranslateForAll(t *testing.T) {
	tr := NewTranslator()
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVar("X", S)
	Y, _ := logic.NewVar("Y", S)

	eq, _ := logic.NewEq(X, Y)
	fa, _ := logic.NewForAll([]*logic.Var{X, Y}, eq)

	zfa, err := tr.Translate(fa)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ForAll:", zfa.String())
}

// Reproduce Python z3_utils.py __main__ examples
func TestTransitiveImplication(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVar("X", S)
	Y, _ := logic.NewVar("Y", S)
	Z, _ := logic.NewVar("Z", S)
	BinRel := mustFS(t, S, S, logic.Boolean)
	leq := logic.NewConst("leq", BinRel)

	leqXY, _ := logic.NewApply(leq, X, Y)
	leqYZ, _ := logic.NewApply(leq, Y, Z)
	leqXZ, _ := logic.NewApply(leq, X, Z)

	// transitive1: ForAll (X,Y,Z). Implies(And(leq(X,Y), leq(Y,Z)), leq(X,Z))
	andTerm, _ := logic.NewAnd(leqXY, leqYZ)
	impl, _ := logic.NewImplies(andTerm, leqXZ)
	transitive1, _ := logic.NewForAll([]*logic.Var{X, Y, Z}, impl)

	// transitive2: ForAll (X,Y,Z). Or(Not(leq(X,Y)), Not(leq(Y,Z)), leq(X,Z))
	notXY, _ := logic.NewNot(leqXY)
	notYZ, _ := logic.NewNot(leqYZ)
	orTerm, _ := logic.NewOr(notXY, notYZ, leqXZ)
	transitive2, _ := logic.NewForAll([]*logic.Var{X, Y, Z}, orTerm)

	// transitive3: Not(Exists (X,Y,Z). And(leq(X,Y), leq(Y,Z), Not(leq(X,Z))))
	notXZ, _ := logic.NewNot(leqXZ)
	andTerm3, _ := logic.NewAnd(leqXY, leqYZ, notXZ)
	existsTerm, _ := logic.NewExists([]*logic.Var{X, Y, Z}, andTerm3)
	transitive3, _ := logic.NewNot(existsTerm)

	tr := NewTranslator()

	// transitive1 => transitive2 (should be true — they're equivalent)
	result, err := tr.Implies(transitive1, transitive2)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("transitive1 should imply transitive2")
	}

	// transitive2 => transitive3 (should be true)
	tr2 := NewTranslator()
	result, err = tr2.Implies(transitive2, transitive3)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("transitive2 should imply transitive3")
	}

	// transitive3 => transitive1 (should be true — all three are equivalent)
	tr3 := NewTranslator()
	result, err = tr3.Implies(transitive3, transitive1)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("transitive3 should imply transitive1")
	}
}

func TestAntisymmetricNotImplied(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVar("X", S)
	Y, _ := logic.NewVar("Y", S)
	Z, _ := logic.NewVar("Z", S)
	BinRel := mustFS(t, S, S, logic.Boolean)
	leq := logic.NewConst("leq", BinRel)

	leqXY, _ := logic.NewApply(leq, X, Y)
	leqYZ, _ := logic.NewApply(leq, Y, Z)
	leqXZ, _ := logic.NewApply(leq, X, Z)

	// transitive3: Not(Exists (X,Y,Z). And(leq(X,Y), leq(Y,Z), Not(leq(X,Z))))
	notXZ, _ := logic.NewNot(leqXZ)
	andTerm3, _ := logic.NewAnd(leqXY, leqYZ, notXZ)
	existsTerm, _ := logic.NewExists([]*logic.Var{X, Y, Z}, andTerm3)
	transitive3, _ := logic.NewNot(existsTerm)

	// antisymmetric: ForAll (X,Y). Implies(And(leq(X,Y), leq(Y,X)), Eq(Y,X))
	leqYX, _ := logic.NewApply(leq, Y, X)
	andAS, _ := logic.NewAnd(leqXY, leqYX)
	eqYX, _ := logic.NewEq(Y, X)
	implAS, _ := logic.NewImplies(andAS, eqYX)
	antisymmetric, _ := logic.NewForAll([]*logic.Var{X, Y}, implAS)

	// transitive3 should NOT imply antisymmetric
	tr := NewTranslator()
	result, err := tr.Implies(transitive3, antisymmetric)
	if err != nil {
		t.Fatal(err)
	}
	if result {
		t.Error("transitivity should NOT imply antisymmetry")
	}
}

func TestIteImplication(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	x := logic.NewConst("x", S)
	y := logic.NewConst("y", S)
	b := logic.NewConst("b", logic.Boolean)

	// b => Eq(Ite(b, x, y), x) (should be true)
	ite, _ := logic.NewIte(b, x, y)
	eqIteX, _ := logic.NewEq(ite, x)

	tr := NewTranslator()
	result, err := tr.Implies(b, eqIteX)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("b should imply Ite(b,x,y) == x")
	}

	// Not(b) => Eq(Ite(b, x, y), y) (should be true)
	notB, _ := logic.NewNot(b)
	eqIteY, _ := logic.NewEq(ite, y)

	tr2 := NewTranslator()
	result, err = tr2.Implies(notB, eqIteY)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("Not(b) should imply Ite(b,x,y) == y")
	}
}

func TestTranslateTrue(t *testing.T) {
	tr := NewTranslator()
	zt, err := tr.Translate(logic.True)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("True:", zt.String())
}

func TestTranslateFalse(t *testing.T) {
	tr := NewTranslator()
	zf, err := tr.Translate(logic.False)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("False:", zf.String())
}

func TestIsSat(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVar("X", S)
	eq, _ := logic.NewEq(X, X)

	tr := NewTranslator()
	result, err := tr.IsSat(eq)
	if err != nil {
		t.Fatal(err)
	}
	if result != Sat {
		t.Error("X == X should be satisfiable")
	}
}

func TestIsUnsat(t *testing.T) {
	S := &logic.UninterpretedSort{Name: "S"}
	X, _ := logic.NewVar("X", S)
	Y, _ := logic.NewVar("Y", S)

	eq, _ := logic.NewEq(X, Y)
	neq, _ := logic.NewNot(eq)
	eqSelf, _ := logic.NewEq(X, X)
	neqSelf, _ := logic.NewNot(eqSelf)

	// X != X is unsatisfiable
	tr := NewTranslator()
	result, err := tr.IsSat(neqSelf)
	if err != nil {
		t.Fatal(err)
	}
	if result != Unsat {
		t.Error("X != X should be unsatisfiable")
	}

	// X == Y and X != Y together is not directly one formula,
	// but Not(Eq(X,X)) captures the idea
	_ = neq
}
