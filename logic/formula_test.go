package logic

import (
	"errors"
	"testing"
)

func TestEqValid(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	eq, err := NewEq(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	if eq.String() != "(X == Y)" {
		t.Errorf("String() = %q, want %q", eq.String(), "(X == Y)")
	}
	if !SortEqual(eq.NodeSort(), Boolean) {
		t.Error("Eq sort should be Boolean")
	}
}

func TestEqDifferentSorts(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	T := &UninterpretedSort{Name: "T"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", T)
	_, err := NewEq(X, Y)
	if err == nil {
		t.Error("Expected error for different sorts")
	}
	var se *SortError
	if !errors.As(err, &se) {
		t.Error("Expected SortError")
	}
}

func TestEqHigherOrder(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs := mustFS(t, S, Boolean)
	X, _ := NewVariable("X", fs)
	Y, _ := NewVariable("Y", fs)
	_, err := NewEq(X, Y)
	if err == nil {
		t.Error("Expected error for higher-order Eq")
	}
}

func TestEqTopSort(t *testing.T) {
	X, _ := NewVariable("X", TopS)
	S := &UninterpretedSort{Name: "S"}
	Y, _ := NewVariable("Y", S)
	_, err := NewEq(X, Y)
	if err != nil {
		t.Error("Eq with TopSort should succeed")
	}
}

func TestNotString(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	eq, _ := NewEq(X, Y)
	notEq, _ := NewNot(eq)
	if notEq.String() != "(X != Y)" {
		t.Errorf("Not(Eq) String() = %q, want %q", notEq.String(), "(X != Y)")
	}

	and, _ := NewAnd(eq)
	notAnd, _ := NewNot(and)
	if notAnd.String() != "~And((X == Y))" {
		t.Errorf("Not(And) String() = %q, want %q", notAnd.String(), "~And((X == Y))")
	}
}

func TestNotBadSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	_, err := NewNot(X) // X is sort S, not Boolean
	if err == nil {
		t.Error("Expected error for Not with non-Boolean")
	}
}

func TestAndOr(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewSymbol("leq", mustFS(t, S, S, Boolean))
	app1, _ := NewApply(leq, X, Y)
	app2, _ := NewApply(leq, Y, X)

	and, err := NewAnd(app1, app2)
	if err != nil {
		t.Fatal(err)
	}
	if and.String() != "And(leq(X,Y), leq(Y,X))" {
		t.Errorf("And String() = %q", and.String())
	}

	or, err := NewOr(app1, app2)
	if err != nil {
		t.Fatal(err)
	}
	if or.String() != "Or(leq(X,Y), leq(Y,X))" {
		t.Errorf("Or String() = %q", or.String())
	}
}

func TestAndBadSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S) // S, not Boolean
	_, err := NewAnd(X)
	if err == nil {
		t.Error("Expected error for And with non-Boolean")
	}
}

func TestTrueFalse(t *testing.T) {
	// True is empty And, False is empty Or
	if True.String() != "And()" {
		t.Errorf("True = %q, want %q", True.String(), "And()")
	}
	if False.String() != "Or()" {
		t.Errorf("False = %q, want %q", False.String(), "Or()")
	}
}

func TestImpliesIff(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewSymbol("leq", mustFS(t, S, S, Boolean))
	a, _ := NewApply(leq, X, Y)
	b, _ := NewApply(leq, Y, X)

	impl, err := NewImplies(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if impl.String() != "(leq(X,Y) -> leq(Y,X))" {
		t.Errorf("Implies = %q", impl.String())
	}

	iff, err := NewIff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if iff.String() != "Iff(leq(X,Y), leq(Y,X))" {
		t.Errorf("Iff = %q", iff.String())
	}
}

func TestForAll(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)
	leq := NewSymbol("leq", mustFS(t, S, S, Boolean))
	leqXY, _ := NewApply(leq, X, Y)
	leqYZ, _ := NewApply(leq, Y, Z)
	leqXZ, _ := NewApply(leq, X, Z)
	andTerm, _ := NewAnd(leqXY, leqYZ)
	impl, _ := NewImplies(andTerm, leqXZ)

	fa, err := NewForAll([]*Variable{X, Y, Z}, impl)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ForAll:", fa.String())
	if !SortEqual(fa.NodeSort(), Boolean) {
		t.Error("ForAll sort should be Boolean")
	}
}

func TestForAllEmpty(t *testing.T) {
	_, err := NewForAll([]*Variable{}, True)
	if err == nil {
		t.Error("Expected error for empty variables")
	}
}

func TestForAllBadBody(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	_, err := NewForAll([]*Variable{X}, X) // body is sort S
	if err == nil {
		t.Error("Expected error for non-Boolean body")
	}
}

func TestExists(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)
	ex, err := NewExists([]*Variable{X}, eq)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Exists:", ex.String())
}

func TestLambda(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)
	lam, err := NewLambda([]*Variable{X}, eq)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Lambda:", lam.String())
}

func TestIte(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	cond, _ := NewEq(X, Y)
	ite, err := NewIte(cond, X, Y)
	if err != nil {
		t.Fatal(err)
	}
	if ite.String() != "Ite((X == Y), X, Y)" {
		t.Errorf("Ite String() = %q", ite.String())
	}
	if !SortEqual(ite.NodeSort(), S) {
		t.Errorf("Ite sort = %s, want S", ite.NodeSort())
	}
}

func TestIteBadCond(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	_, err := NewIte(X, X, X) // cond is S, not Boolean
	if err == nil {
		t.Error("Expected error for non-Boolean condition")
	}
}

func TestGloballyEventually(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	eq, _ := NewEq(X, Y)

	g, err := NewGlobally(nil, eq)
	if err != nil {
		t.Fatal(err)
	}
	if g.String() != "globally((X == Y))" {
		t.Errorf("Globally = %q", g.String())
	}

	env := "env1"
	e, err := NewEventually(&env, eq)
	if err != nil {
		t.Fatal(err)
	}
	if e.String() != "eventually[env1]((X == Y))" {
		t.Errorf("Eventually = %q", e.String())
	}
}

func TestWhenOperator(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	cond, _ := NewEq(X, X)
	w, err := NewWhenOperator("when", X, cond)
	if err != nil {
		t.Fatal(err)
	}
	if w.String() != "WhenOperator(when,X,(X == X))" {
		t.Errorf("WhenOperator = %q", w.String())
	}
	if !SortEqual(w.NodeSort(), S) {
		t.Error("WhenOperator sort should be S")
	}
}

func TestCond(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	cond, _ := NewEq(X, X)
	c, err := NewCond(cond, X)
	if err != nil {
		t.Fatal(err)
	}
	if c.String() != "Cond((X == X), X)" {
		t.Errorf("Cond = %q", c.String())
	}
}

func TestNamedBinder(t *testing.T) {
	X, _ := NewVariable("X", TopS)
	Y, _ := NewVariable("Y", TopS)
	Z, _ := NewVariable("Z", TopS)

	f := NewSymbol("f", mustFS(t, TopS, TopS, Boolean))
	fXY, _ := NewApply(f, X, Y)
	fXZ, _ := NewApply(f, X, Z)
	andTerm, _ := NewAnd(fXY, fXZ)
	eqYZ, _ := NewEq(Y, Z)
	impl, _ := NewImplies(andTerm, eqYZ)

	b, err := NewNamedBinder("mybinder", []*Variable{X, Y, Z}, nil, impl)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("NamedBinder:", b.String())
	t.Log("NamedBinder sort:", b.NodeSort().String())

	// Sort should be TopSort * TopSort * TopSort -> Boolean
	fs, ok := b.NodeSort().(*FunctionSort)
	if !ok {
		t.Fatalf("Expected FunctionSort, got %T", b.NodeSort())
	}
	if fs.Arity() != 3 {
		t.Errorf("Arity = %d, want 3", fs.Arity())
	}
}

func TestNamedBinderNoVars(t *testing.T) {
	X, _ := NewVariable("X", TopS)
	S := &UninterpretedSort{Name: "S"}
	Z, _ := NewVariable("Z", S)

	b, err := NewNamedBinder("mybinder", []*Variable{X, Y_(), Z}, nil, Z)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("NamedBinder:", b.String())
	t.Log("NamedBinder sort:", b.NodeSort().String())
}

func TestNamedBinderCall(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)
	nb, _ := NewNamedBinder("nb", []*Variable{X}, nil, eq)

	// Call with no args returns self
	result, err := nb.Call()
	if err != nil {
		t.Fatal(err)
	}
	if result != nb {
		t.Error("Call with no args should return self")
	}
}

func TestFormulaChildren(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)

	// ForAll children should be [body] only, not variables
	fa, _ := NewForAll([]*Variable{X}, eq)
	children := fa.Children()
	if len(children) != 1 {
		t.Errorf("ForAll Children() len = %d, want 1", len(children))
	}
}

func TestFormulaEquality(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	eq1, _ := NewEq(X, Y)
	eq2, _ := NewEq(X, Y)
	if !eq1.Equal(eq2) {
		t.Error("Same Eq nodes should be equal")
	}

	not1, _ := NewNot(eq1)
	not2, _ := NewNot(eq2)
	if !not1.Equal(not2) {
		t.Error("Same Not nodes should be equal")
	}
}

// Y_ helper for creating Y variable with TopSort
func Y_() *Variable {
	v, _ := NewVariable("Y", TopS)
	return v
}

// Python __main__ antisymmetric example
func TestAntisymmetric(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewSymbol("leq", mustFS(t, S, S, Boolean))

	leqXY, _ := NewApply(leq, X, Y)
	leqYX, _ := NewApply(leq, Y, X)
	andTerm, _ := NewAnd(leqXY, leqYX)
	eqYX, _ := NewEq(Y, X)
	impl, _ := NewImplies(andTerm, eqYX)
	antisym, err := NewForAll([]*Variable{X, Y}, impl)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("antisymmetric:", antisym.String())
}

func FuzzAndConstruction(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(5)
	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 || n > 20 {
			return
		}
		S := &UninterpretedSort{Name: "S"}
		X, _ := NewVariable("X", S)
		Y, _ := NewVariable("Y", S)
		eq, _ := NewEq(X, Y)

		terms := make([]Node, n)
		for i := range terms {
			terms[i] = eq
		}
		and, err := NewAnd(terms...)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(and.Children()) != n {
			t.Errorf("Children len = %d, want %d", len(and.Children()), n)
		}
	})
}
