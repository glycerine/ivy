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
	// String() now matches Python lg.Eq.__str__ = pretty_fmla → nary_ugly("=").
	// drop_annotations strips the second arg's sort.
	if eq.String() != "X:S = Y" {
		t.Errorf("String() = %q, want %q", eq.String(), "X:S = Y")
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
	fs := mustFuncSort(t, S, Boolean)
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
	// Python: Not(Eq(a,b)).ugly → nary_ugly("~=", ...) → "a ~= b"
	if notEq.String() != "X:S ~= Y" {
		t.Errorf("Not(Eq) String() = %q, want %q", notEq.String(), "X:S ~= Y")
	}

	and, _ := NewAnd(eq)
	notAnd, _ := NewNot(and)
	// Python: Not(other).ugly → "~" + body.ugly(6)
	// And(eq).ugly → nary_paren("&", [eq], 5, 6) → "(X:S = Y)"  (single-arg And keeps the parens)
	if notAnd.String() != "~(X:S = Y)" {
		t.Errorf("Not(And) String() = %q, want %q", notAnd.String(), "~(X:S = Y)")
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
	leq := NewConst("leq", mustFuncSort(t, S, S, Boolean))
	app1, _ := NewApply(leq, X, Y)
	app2, _ := NewApply(leq, Y, X)

	and, err := NewAnd(app1, app2)
	if err != nil {
		t.Fatal(err)
	}
	// Python And.ugly → nary_paren("&", ...) → always parenthesizes.
	if and.String() != "(leq(X,Y) & leq(Y,X))" {
		t.Errorf("And String() = %q", and.String())
	}

	or, err := NewOr(app1, app2)
	if err != nil {
		t.Fatal(err)
	}
	// Python Or.ugly → nary_ugly("|", ...) — parens only when myprec(4) <= prec(0), which is false.
	if or.String() != "leq(X,Y) | leq(Y,X)" {
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
	// True is empty And → Python ugly → "true". False is empty Or → "false".
	if True.String() != "true" {
		t.Errorf("True = %q, want %q", True.String(), "true")
	}
	if False.String() != "false" {
		t.Errorf("False = %q, want %q", False.String(), "false")
	}
}

func TestImpliesIff(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewConst("leq", mustFuncSort(t, S, S, Boolean))
	a, _ := NewApply(leq, X, Y)
	b, _ := NewApply(leq, Y, X)

	impl, err := NewImplies(a, b)
	if err != nil {
		t.Fatal(err)
	}
	// Python Implies.ugly → nary_ugly("->", ...).
	if impl.String() != "leq(X,Y) -> leq(Y,X)" {
		t.Errorf("Implies = %q", impl.String())
	}

	iff, err := NewIff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	// Python Iff.ugly → nary_ugly("<->", ...).
	if iff.String() != "leq(X,Y) <-> leq(Y,X)" {
		t.Errorf("Iff = %q", iff.String())
	}
}

func TestForAll(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)
	leq := NewConst("leq", mustFuncSort(t, S, S, Boolean))
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
	// Python Ite.ugly → '({} if {} else {})' format.
	if ite.String() != "(X:S if (X = Y) else Y)" {
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
	// Globally is not in Python's monkey-patch list, so its String() body is
	// unchanged — but its child eq.String() now uses pretty_fmla.
	if g.String() != "globally(X:S = Y)" {
		t.Errorf("Globally = %q", g.String())
	}

	env := "env1"
	e, err := NewEventually(&env, eq)
	if err != nil {
		t.Fatal(err)
	}
	if e.String() != "eventually[env1](X:S = Y)" {
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
	// WhenOperator is not in Python's monkey-patch list. Its child variable
	// X.String() and Eq cond.String() now use pretty_fmla.
	if w.String() != "WhenOperator(when,X:S,X:S = X)" {
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
	// Cond is not in Python's monkey-patch list, but its children use pretty_fmla.
	if c.String() != "Cond(X:S = X, X:S)" {
		t.Errorf("Cond = %q", c.String())
	}
}

func TestNamedBinder(t *testing.T) {
	X, _ := NewVariable("X", TopS)
	Y, _ := NewVariable("Y", TopS)
	Z, _ := NewVariable("Z", TopS)

	f := NewConst("f", mustFuncSort(t, TopS, TopS, Boolean))
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
	leq := NewConst("leq", mustFuncSort(t, S, S, Boolean))

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

		terms := make([]Expr, n)
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

// TestDropAnnotationsQuantifierOrder verifies that dropAnnotations processes
// the body BEFORE the bound variables for ForAll, Exists, and Lambda —
// matching Python's quant_drop_annotations (ivy_logic.py:1434-1436).
//
// When a polymorphic operator like <= returns Boolean, the first argument's
// sort annotation is preserved in the body (inferredSort && !Boolean = false),
// and the corresponding bound variable's annotation is stripped because its
// name is already in annotatedVars by the time the bound variables are processed.
//
// Python output:  forall T. T:lclock <= _T
// Wrong Go output (vars-first): forall T:lclock. T <= _T
func TestDropAnnotationsQuantifierOrder(t *testing.T) {
	lclock := &UninterpretedSort{Name: "lclock"}

	// Sub-test 1: ForAll with polymorphic <=
	t.Run("ForAll_polymorphic_le", func(t *testing.T) {
		T, _ := NewVariable("T", lclock)
		_T, _ := NewVariable("_T", lclock)
		leSort := mustFuncSort(t, lclock, lclock, Boolean)
		le := NewConst("<=", leSort)
		body, err := NewApply(le, T, _T)
		if err != nil {
			t.Fatal(err)
		}
		fa, err := NewForAll([]*Variable{T}, body)
		if err != nil {
			t.Fatal(err)
		}
		got := fa.String()
		want := "forall T. T:lclock <= _T"
		if got != want {
			t.Errorf("ForAll String() =\n  %q\nwant:\n  %q\n(annotation must be on body occurrence, not quantifier binding)", got, want)
		}
	})

	// Sub-test 2: Exists with polymorphic <=
	t.Run("Exists_polymorphic_le", func(t *testing.T) {
		T, _ := NewVariable("T", lclock)
		_T, _ := NewVariable("_T", lclock)
		leSort := mustFuncSort(t, lclock, lclock, Boolean)
		le := NewConst("<=", leSort)
		body, err := NewApply(le, T, _T)
		if err != nil {
			t.Fatal(err)
		}
		ex, err := NewExists([]*Variable{T}, body)
		if err != nil {
			t.Fatal(err)
		}
		got := ex.String()
		want := "exists T. T:lclock <= _T"
		if got != want {
			t.Errorf("Exists String() =\n  %q\nwant:\n  %q", got, want)
		}
	})

	// Sub-test 3: Lambda with polymorphic <=
	t.Run("Lambda_polymorphic_le", func(t *testing.T) {
		T, _ := NewVariable("T", lclock)
		_T, _ := NewVariable("_T", lclock)
		leSort := mustFuncSort(t, lclock, lclock, Boolean)
		le := NewConst("<=", leSort)
		body, err := NewApply(le, T, _T)
		if err != nil {
			t.Fatal(err)
		}
		lam, err := NewLambda([]*Variable{T}, body)
		if err != nil {
			t.Fatal(err)
		}
		got := lam.String()
		want := "lambda T. T:lclock <= _T"
		if got != want {
			t.Errorf("Lambda String() =\n  %q\nwant:\n  %q", got, want)
		}
	})

	// Sub-test 4: ForAll with non-polymorphic function (control case).
	// Non-polymorphic Apply passes inferredSort=true to all args, so the body
	// occurrence is stripped and the annotation stays on the bound variable.
	t.Run("ForAll_nonpolymorphic_func", func(t *testing.T) {
		S := &UninterpretedSort{Name: "S"}
		X, _ := NewVariable("X", S)
		fSort := mustFuncSort(t, S, Boolean)
		f := NewConst("f", fSort)
		body, err := NewApply(f, X)
		if err != nil {
			t.Fatal(err)
		}
		fa, err := NewForAll([]*Variable{X}, body)
		if err != nil {
			t.Fatal(err)
		}
		got := fa.String()
		// Non-polymorphic: body args get inferredSort=true, so X annotation
		// is stripped in body. But X is also added to annotatedVars during
		// body processing (inferredSort=true → strip + add). Then bound var
		// X is processed with inferredSort=false, but X is in annotatedVars
		// → also stripped. Both stripped means no annotation anywhere.
		want := "forall X. f(X)"
		if got != want {
			t.Errorf("ForAll (non-polymorphic) String() =\n  %q\nwant:\n  %q", got, want)
		}
	})

	// Sub-test 5: ForAll with multiple bound vars, only first used in polymorphic context.
	// T appears in a polymorphic <=, U appears in a non-polymorphic f().
	t.Run("ForAll_multi_vars_mixed", func(t *testing.T) {
		T, _ := NewVariable("T", lclock)
		U, _ := NewVariable("U", lclock)
		leSort := mustFuncSort(t, lclock, lclock, Boolean)
		le := NewConst("<=", leSort)
		leApp, err := NewApply(le, T, U)
		if err != nil {
			t.Fatal(err)
		}
		fa, err := NewForAll([]*Variable{T, U}, leApp)
		if err != nil {
			t.Fatal(err)
		}
		got := fa.String()
		// <= is polymorphic returning Boolean:
		// arg0 (T) gets inferredSort=false → keeps annotation → T:lclock
		// arg1 (U) gets inferredSort=(name != "*>")=true → stripped
		// Then bound vars: T in annotatedVars → stripped; U in annotatedVars → stripped
		want := "forall T,U. T:lclock <= U"
		if got != want {
			t.Errorf("ForAll (multi-var) String() =\n  %q\nwant:\n  %q", got, want)
		}
	})

	// Sub-test 6: NamedBinder processes vars BEFORE body (matching Python's
	// left-to-right argument evaluation). Annotation stays on bound variable.
	t.Run("NamedBinder_vars_before_body", func(t *testing.T) {
		T, _ := NewVariable("T", lclock)
		_T, _ := NewVariable("_T", lclock)
		leSort := mustFuncSort(t, lclock, lclock, Boolean)
		le := NewConst("<=", leSort)
		body, err := NewApply(le, T, _T)
		if err != nil {
			t.Fatal(err)
		}
		nb, err := NewNamedBinder("l2s_s", []*Variable{T}, nil, body)
		if err != nil {
			t.Fatal(err)
		}
		got := nb.String()
		// NamedBinder processes vars first (like Python), so T:lclock stays
		// on the bound variable, and body T is stripped.
		want := "$l2s_s T:lclock. T <= _T"
		if got != want {
			t.Errorf("NamedBinder String() =\n  %q\nwant:\n  %q", got, want)
		}
	})

	// Sub-test 7: Nested quantifiers — inner forall inside outer body.
	// Ensures body-first ordering works correctly through nesting.
	t.Run("ForAll_nested", func(t *testing.T) {
		T, _ := NewVariable("T", lclock)
		U, _ := NewVariable("U", lclock)
		leSort := mustFuncSort(t, lclock, lclock, Boolean)
		le := NewConst("<=", leSort)
		innerBody, err := NewApply(le, U, T)
		if err != nil {
			t.Fatal(err)
		}
		innerFA, err := NewForAll([]*Variable{U}, innerBody)
		if err != nil {
			t.Fatal(err)
		}
		outerFA, err := NewForAll([]*Variable{T}, innerFA)
		if err != nil {
			t.Fatal(err)
		}
		got := outerFA.String()
		// Outer body (inner forall) is processed first.
		// Inner forall: body (U:lclock <= T:lclock) processed first:
		//   <= polymorphic, Boolean return: arg0 U gets inferredSort=false → keeps U:lclock
		//   arg1 T gets inferredSort=true (name "<=" != "*>") → stripped T
		//   Wait: T not yet in annotatedVars at this point, so:
		//     inferredSort=true → strip + add T to annotatedVars
		//   Inner bound var U: in annotatedVars → stripped
		// Then outer bound var T: in annotatedVars → stripped
		want := "forall T. (forall U. U:lclock <= T)"
		if got != want {
			t.Errorf("ForAll (nested) String() =\n  %q\nwant:\n  %q", got, want)
		}
	})
}
