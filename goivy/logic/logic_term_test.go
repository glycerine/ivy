package logic

import (
	"errors"
	"testing"
)

func TestVarValid(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	v, err := NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	// String() now matches Python lg.Var.__str__ = pretty_fmla, which
	// includes the sort for non-TopSort variables.
	if v.String() != "X:S" {
		t.Errorf("String() = %q, want %q", v.String(), "X:S")
	}
	if !SortEqual(v.NodeSort(), S) {
		t.Error("Sort should be S")
	}
}

func TestVarInvalidName(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	_, err := NewVariable("x", S)
	if err == nil {
		t.Error("Expected error for lowercase variable name")
	}
	var ivyErr *IvyError
	if !errors.As(err, &ivyErr) {
		t.Error("Expected IvyError")
	}
}

func TestVarEmpty(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	_, err := NewVariable("", S)
	if err == nil {
		t.Error("Expected error for empty variable name")
	}
}

func TestVarEqual(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	v1, _ := NewVariable("X", S)
	v2, _ := NewVariable("X", S)
	v3, _ := NewVariable("Y", S)
	if !v1.Equal(v2) {
		t.Error("Same vars should be equal")
	}
	if v1.Equal(v3) {
		t.Error("Different vars should not be equal")
	}
}

func TestConst(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	c := NewConst("leq", logicMustFuncSort(t, S, S, Boolean))
	if c.String() != "leq" {
		t.Errorf("String() = %q, want %q", c.String(), "leq")
	}
}

func TestApplyValid(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewConst("leq", logicMustFuncSort(t, S, S, Boolean))

	app, err := NewApply(leq, X, Y)
	if err != nil {
		t.Fatal(err)
	}
	// String() matches Python lg.Apply.__str__ = pretty_fmla → app_ugly,
	// which uses ',' (no space) and drops inferable variable sorts.
	if app.String() != "leq(X,Y)" {
		t.Errorf("String() = %q, want %q", app.String(), "leq(X,Y)")
	}
	if !SortEqual(app.NodeSort(), Boolean) {
		t.Errorf("Sort should be Boolean, got %s", app.NodeSort())
	}
}

func TestApplyArityMismatch(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	leq := NewConst("leq", logicMustFuncSort(t, S, S, Boolean))

	_, err := NewApply(leq, X) // expects 2 args, got 1
	if err == nil {
		t.Error("Expected arity error")
	}
	var sortErr *SortError
	if !errors.As(err, &sortErr) {
		t.Errorf("Expected SortError, got %T", err)
	}
}

func TestApplySortMismatch(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	T := &UninterpretedSort{Name: "T"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", T)
	leq := NewConst("leq", logicMustFuncSort(t, S, S, Boolean))

	_, err := NewApply(leq, X, Y) // Y is T, expects S
	if err == nil {
		t.Error("Expected sort mismatch error")
	}
}

func TestApplyTopSortPassthrough(t *testing.T) {
	X, _ := NewVariable("X", TopS)
	Y, _ := NewVariable("Y", TopS)
	h := NewConst("h", TopS)

	app, err := NewApply(h, X, Y)
	if err != nil {
		t.Fatal(err)
	}
	if !SortEqual(app.NodeSort(), TopS) {
		t.Error("Apply with TopSort func should have TopSort result")
	}
}

func TestApplyNonFunction(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	c := NewConst("c", S) // not a function sort

	_, err := NewApply(c, X)
	if err == nil {
		t.Error("Expected error for non-function application")
	}
}

func TestVarCall(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs := logicMustFuncSort(t, S, Boolean)
	V, _ := NewVariable("V", fs)
	X, _ := NewVariable("X", S)

	// Call with args
	result, err := V.Call(X)
	if err != nil {
		t.Fatal(err)
	}
	if result.String() != "V(X)" {
		t.Errorf("String() = %q, want %q", result.String(), "V(X)")
	}

	// Call with no args returns self
	result, err = V.Call()
	if err != nil {
		t.Fatal(err)
	}
	if result != V {
		t.Error("Call() with no args should return self")
	}
}

func TestConstCall(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	leq := NewConst("leq", logicMustFuncSort(t, S, S, Boolean))
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)

	result, err := leq.Call(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	if result.String() != "leq(X,Y)" {
		t.Errorf("String() = %q, want %q", result.String(), "leq(X,Y)")
	}
}

func TestApplyEqual(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewConst("leq", logicMustFuncSort(t, S, S, Boolean))

	a1, _ := NewApply(leq, X, Y)
	a2, _ := NewApply(leq, X, Y)
	if !a1.Equal(a2) {
		t.Error("Same applies should be equal")
	}
}

func TestApplyChildren(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	leq := NewConst("leq", logicMustFuncSort(t, S, S, Boolean))
	app, _ := NewApply(leq, X, Y)

	children := app.Children()
	if len(children) != 2 { // terms only (Func excluded — matches Python Apply.args)
		t.Errorf("Children() len = %d, want 2", len(children))
	}
}

// Reproduce Python __main__ example
func TestPythonExample(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)

	BinRel := logicMustFuncSort(t, S, S, Boolean)
	leq := NewConst("leq", BinRel)

	leqXY, _ := NewApply(leq, X, Y)
	leqYZ, _ := NewApply(leq, Y, Z)
	leqXZ, _ := NewApply(leq, X, Z)

	// transitive1: ForAll (X,Y,Z). Implies(And(leq(X,Y), leq(Y,Z)), leq(X,Z))
	andTerm, _ := NewAnd(leqXY, leqYZ)
	impl, _ := NewImplies(andTerm, leqXZ)
	trans1, err := NewForAll([]*Variable{X, Y, Z}, impl)
	if err != nil {
		t.Fatal(err)
	}
	if !SortEqual(trans1.NodeSort(), Boolean) {
		t.Error("ForAll sort should be Boolean")
	}
	t.Log("transitive1:", trans1.String())
}

func FuzzVarName(f *testing.F) {
	f.Add("X")
	f.Add("x")
	f.Add("")
	f.Add("123")
	f.Fuzz(func(t *testing.T, name string) {
		S := &UninterpretedSort{Name: "S"}
		v, err := NewVariable(name, S)
		if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' {
			if err == nil {
				t.Error("Expected error for invalid name")
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for name %q: %v", name, err)
			}
			// PrettyFmla now appends ":S" for the non-TopSort variable.
			want := name + ":S"
			if v.String() != want {
				t.Errorf("String() = %q, want %q", v.String(), want)
			}
		}
	})
}
