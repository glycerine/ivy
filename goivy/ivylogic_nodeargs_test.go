package goivy

import (
	"testing"
)

// These tests verify that NodeArgs returns the same elements as Python's
// .args property for each logic type. This is critical for fragment.go's
// mapFmla, makeSkolems, and createMacroMaps which iterate NodeArgs.
//
// Key invariant: Python overrides Apply.args to return only self.terms
// (NOT the function symbol). Go's NodeArgs must match this.

func TestNodeArgsApplyExcludesFunc(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs, _ := NewFunctionSort(S, S, Boolean)
	f := NewConst("f", fs)
	x := NewConst("x", S)
	y := NewConst("y", S)

	app := MustApply(f, x, y)
	args := NodeArgs(app)

	if len(args) != 2 {
		t.Fatalf("NodeArgs(IvyApply(f, [x, y])): expected 2 args, got %d", len(args))
	}
	if args[0] != x {
		t.Errorf("NodeArgs(Apply)[0] should be x, got %v", args[0])
	}
	if args[1] != y {
		t.Errorf("NodeArgs(Apply)[1] should be y, got %v", args[1])
	}
	// Verify function is NOT included
	for i, a := range args {
		if a == f {
			t.Errorf("NodeArgs(Apply)[%d] should not be the function symbol", i)
		}
	}
}

func TestNodeArgsApplyNoTerms(t *testing.T) {
	fs, _ := NewFunctionSort(Boolean)
	f := NewConst("c", fs)
	app := MustApply(f)
	args := NodeArgs(app)

	if len(args) != 0 {
		t.Fatalf("NodeArgs(IvyApply(f, [])): expected 0 args, got %d", len(args))
	}
}

func TestNodeArgsNot(t *testing.T) {
	body := NewConst("p", Boolean)
	not := &Not{Body: body}
	args := NodeArgs(not)

	if len(args) != 1 {
		t.Fatalf("NodeArgs(Not): expected 1 arg, got %d", len(args))
	}
	if args[0] != body {
		t.Errorf("NodeArgs(Not)[0] should be body")
	}
}

func TestNodeArgsImplies(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	imp := &Implies{T1: a, T2: b}
	args := NodeArgs(imp)

	if len(args) != 2 {
		t.Fatalf("NodeArgs(Implies): expected 2 args, got %d", len(args))
	}
	if args[0] != a {
		t.Errorf("NodeArgs(Implies)[0] should be antecedent")
	}
	if args[1] != b {
		t.Errorf("NodeArgs(Implies)[1] should be consequent")
	}
}

func TestNodeArgsAnd(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	c := NewConst("c", Boolean)
	and := &And{Terms: []Expr{a, b, c}}
	args := NodeArgs(and)

	if len(args) != 3 {
		t.Fatalf("NodeArgs(And): expected 3 args, got %d", len(args))
	}
}

func TestNodeArgsOr(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	or := &Or{Terms: []Expr{a, b}}
	args := NodeArgs(or)

	if len(args) != 2 {
		t.Fatalf("NodeArgs(Or): expected 2 args, got %d", len(args))
	}
}

func TestNodeArgsIff(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	iff := &Iff{T1: a, T2: b}
	args := NodeArgs(iff)

	if len(args) != 2 {
		t.Fatalf("NodeArgs(Iff): expected 2 args, got %d", len(args))
	}
}

func TestNodeArgsEq(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	x := NewConst("x", S)
	y := NewConst("y", S)
	eq := &Eq{T1: x, T2: y}
	args := NodeArgs(eq)

	if len(args) != 2 {
		t.Fatalf("NodeArgs(Eq): expected 2 args, got %d", len(args))
	}
	if args[0] != x || args[1] != y {
		t.Errorf("NodeArgs(Eq) should return [t1, t2]")
	}
}

func TestNodeArgsIte(t *testing.T) {
	cond := NewConst("c", Boolean)
	then := NewConst("t", Boolean)
	els := NewConst("e", Boolean)
	ite := &Ite{Cond: cond, Then: then, Else: els}
	args := NodeArgs(ite)

	if len(args) != 3 {
		t.Fatalf("NodeArgs(Ite): expected 3 args, got %d", len(args))
	}
	if args[0] != cond {
		t.Errorf("NodeArgs(Ite)[0] should be cond")
	}
	if args[1] != then {
		t.Errorf("NodeArgs(Ite)[1] should be then")
	}
	if args[2] != els {
		t.Errorf("NodeArgs(Ite)[2] should be else")
	}
}

func TestNodeArgsForAllExcludesVars(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	v, _ := NewVariable("X", S)
	body := NewConst("p", Boolean)
	fa := &ForAll{Variables: []*Variable{v}, Body: body}
	args := NodeArgs(fa)

	if len(args) != 1 {
		t.Fatalf("NodeArgs(ForAll): expected 1 arg (body only), got %d", len(args))
	}
	if args[0] != body {
		t.Errorf("NodeArgs(ForAll)[0] should be body, not variables")
	}
}

func TestNodeArgsExistsExcludesVars(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	v, _ := NewVariable("X", S)
	body := NewConst("p", Boolean)
	ex := &Exists{Variables: []*Variable{v}, Body: body}
	args := NodeArgs(ex)

	if len(args) != 1 {
		t.Fatalf("NodeArgs(Exists): expected 1 arg (body only), got %d", len(args))
	}
	if args[0] != body {
		t.Errorf("NodeArgs(Exists)[0] should be body, not variables")
	}
}

func TestNodeArgsConstNil(t *testing.T) {
	c := NewConst("c", Boolean)
	args := NodeArgs(c)
	if args != nil {
		t.Errorf("NodeArgs(Const): expected nil, got %v", args)
	}
}

func TestNodeArgsVariableNil(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	v, _ := NewVariable("X", S)
	args := NodeArgs(v)
	if args != nil {
		t.Errorf("NodeArgs(Variable): expected nil, got %v", args)
	}
}

// TestNodeArgsPolarConsistency verifies that Polar(fmla, i, pol) uses
// the same position indices as NodeArgs, matching Python's
// il.polar(fmla, pos, pol) with enumerate(fmla.args).
func TestNodeArgsPolarConsistency(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)

	// Implies: pos=0 gets negated polarity, pos=1 gets same polarity
	imp := &Implies{T1: a, T2: b}
	impArgs := NodeArgs(imp)
	if len(impArgs) != 2 {
		t.Fatal("Implies should have 2 args")
	}
	if Polar(imp, 0, 0) != 1 { // antecedent: negated (0 → 1)
		t.Errorf("Polar(Implies, 0, 0) should be 1 (negated), got %d", Polar(imp, 0, 0))
	}
	if Polar(imp, 1, 0) != 0 { // consequent: same polarity
		t.Errorf("Polar(Implies, 1, 0) should be 0 (same), got %d", Polar(imp, 1, 0))
	}

	// Ite: pos=0 gets "both" (-1), pos=1,2 get same polarity
	cond := NewConst("c", Boolean)
	then := NewConst("t", Boolean)
	els := NewConst("e", Boolean)
	ite := &Ite{Cond: cond, Then: then, Else: els}
	if Polar(ite, 0, 0) != -1 { // condition: both polarities
		t.Errorf("Polar(Ite, 0, 0) should be -1 (both), got %d", Polar(ite, 0, 0))
	}
	if Polar(ite, 1, 0) != 0 { // then: same polarity
		t.Errorf("Polar(Ite, 1, 0) should be 0, got %d", Polar(ite, 1, 0))
	}

	// Apply: all positions get "both" (-1) — Apply falls through to default
	S := &UninterpretedSort{Name: "S"}
	fs2, _ := NewFunctionSort(S, Boolean)
	f := NewConst("f", fs2)
	x := NewConst("x", S)
	app := MustApply(f, x)
	if Polar(app, 0, 0) != -1 {
		t.Errorf("Polar(Apply, 0, 0) should be -1 (both), got %d", Polar(app, 0, 0))
	}

	// Not: single arg gets negated polarity
	not := &Not{Body: a}
	if Polar(not, 0, 0) != 1 {
		t.Errorf("Polar(Not, 0, 0) should be 1 (negated), got %d", Polar(not, 0, 0))
	}

	// And/Or/ForAll/Exists: all positions keep same polarity
	and := &And{Terms: []Expr{a, b}}
	if Polar(and, 0, 0) != 0 {
		t.Errorf("Polar(And, 0, 0) should be 0 (same), got %d", Polar(and, 0, 0))
	}
}
