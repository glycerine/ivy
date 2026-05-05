package goivy

import (
	"testing"
)

var autoTestAstCfg = NewAstConfig()

// --- buildOrExpr ---

func TestBuildOrExpr_Empty(t *testing.T) {
	result := buildOrExpr(nil)
	or, ok := result.(*Or)
	if !ok {
		t.Fatalf("expected *lg.Or, got %T", result)
	}
	if len(or.Terms) != 0 {
		t.Errorf("expected 0 terms, got %d", len(or.Terms))
	}
}

func TestBuildOrExpr_Single(t *testing.T) {
	a := NewConst("a", Boolean)
	result := buildOrExpr([]Expr{a})
	// Single element should be wrapped in Or (matching Python behavior).
	or, ok := result.(*Or)
	if !ok {
		t.Fatalf("expected *lg.Or, got %T", result)
	}
	if len(or.Terms) != 1 || or.Terms[0] != a {
		t.Error("expected Or with single term")
	}
}

func TestBuildOrExpr_Multiple(t *testing.T) {
	a := NewConst("a", Boolean)
	b := NewConst("b", Boolean)
	result := buildOrExpr([]Expr{a, b})
	or, ok := result.(*Or)
	if !ok {
		t.Fatalf("expected *lg.Or, got %T", result)
	}
	if len(or.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(or.Terms))
	}
}

// --- isGloballyOrNotEventually ---

func TestIsGloballyOrNotEventually_Globally(t *testing.T) {
	g := &Globally{Body: True}
	if !isGloballyOrNotEventually(g) {
		t.Error("expected true for Globally")
	}
}

func TestIsGloballyOrNotEventually_NotEventually(t *testing.T) {
	ne := &Not{Body: &Eventually{Body: True}}
	if !isGloballyOrNotEventually(ne) {
		t.Error("expected true for Not{Eventually}")
	}
}

func TestIsGloballyOrNotEventually_Eventually(t *testing.T) {
	e := &Eventually{Body: True}
	if isGloballyOrNotEventually(e) {
		t.Error("expected false for Eventually")
	}
}

func TestIsGloballyOrNotEventually_Plain(t *testing.T) {
	c := NewConst("c", Boolean)
	if isGloballyOrNotEventually(c) {
		t.Error("expected false for plain Const")
	}
}

func TestIsGloballyOrNotEventually_NotGlobally(t *testing.T) {
	ng := &Not{Body: &Globally{Body: True}}
	if isGloballyOrNotEventually(ng) {
		t.Error("expected false for Not{Globally}")
	}
}

// --- isEventuallyOrNotGlobally ---

func TestIsEventuallyOrNotGlobally_Eventually(t *testing.T) {
	e := &Eventually{Body: True}
	if !isEventuallyOrNotGlobally(e) {
		t.Error("expected true for Eventually")
	}
}

func TestIsEventuallyOrNotGlobally_NotGlobally(t *testing.T) {
	ng := &Not{Body: &Globally{Body: True}}
	if !isEventuallyOrNotGlobally(ng) {
		t.Error("expected true for Not{Globally}")
	}
}

func TestIsEventuallyOrNotGlobally_Globally(t *testing.T) {
	g := &Globally{Body: True}
	if isEventuallyOrNotGlobally(g) {
		t.Error("expected false for Globally")
	}
}

func TestIsEventuallyOrNotGlobally_Plain(t *testing.T) {
	c := NewConst("c", Boolean)
	if isEventuallyOrNotGlobally(c) {
		t.Error("expected false for plain Const")
	}
}

// --- sameLHSSort ---

func TestSameLHSSort_BothNil(t *testing.T) {
	if !sameLHSSort(nil, nil) {
		t.Error("expected true for nil, nil")
	}
}

func TestSameLHSSort_SameSort(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	fs, _ := NewFunctionSort(s, Boolean)
	f := NewConst("f", fs)
	g := NewConst("g", fs)
	x := hooksVar("X", s)
	y := hooksVar("Y", s)

	a := &Eq{T1: MustApply(f, x), T2: True}
	b := &Eq{T1: MustApply(g, y), T2: True}

	if !sameLHSSort(a, b) {
		t.Error("expected true for same sort")
	}
}

func TestSameLHSSort_DiffSort(t *testing.T) {
	s1 := &UninterpretedSort{Name: "S"}
	s2 := &UninterpretedSort{Name: "T"}
	fs1, _ := NewFunctionSort(s1, Boolean)
	fs2, _ := NewFunctionSort(s2, Boolean)
	f := NewConst("f", fs1)
	g := NewConst("g", fs2)
	x := hooksVar("X", s1)
	y := hooksVar("Y", s2)

	a := &Eq{T1: MustApply(f, x), T2: True}
	b := &Eq{T1: MustApply(g, y), T2: True}

	if sameLHSSort(a, b) {
		t.Error("expected false for different sorts")
	}
}

func TestSameLHSSort_ConstLHS(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	c := NewConst("c", s)
	d := NewConst("d", s)
	a := &Eq{T1: c, T2: True}
	b := &Eq{T1: d, T2: True}
	if !sameLHSSort(a, b) {
		t.Error("expected true for same Const sort")
	}
}

// --- cloneLHS ---

func TestCloneLHS_Const(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	c := NewConst("f", s)
	result := cloneLHS(c, "g")
	rc, ok := result.(*Const)
	if !ok {
		t.Fatalf("expected *lg.Const, got %T", result)
	}
	if rc.Name != "g" {
		t.Errorf("expected name 'g', got %q", rc.Name)
	}
	if !rc.CSort.Equal(s) {
		t.Errorf("sort should be preserved")
	}
}

func TestCloneLHS_Apply(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	fs, _ := NewFunctionSort(s, Boolean)
	f := NewConst("f", fs)
	x := hooksVar("X", s)
	app := MustApply(f, x)

	result := cloneLHS(app, "g")
	ra, ok := result.(*Apply)
	if !ok {
		t.Fatalf("expected *lg.Apply, got %T", result)
	}
	if rc, ok := ra.Func.(*Const); !ok || rc.Name != "g" {
		t.Error("expected Func name 'g'")
	}
	if len(ra.Terms) != 1 {
		t.Errorf("expected 1 term, got %d", len(ra.Terms))
	}
}

func TestCloneLHS_Other(t *testing.T) {
	s := &UninterpretedSort{Name: "S"}
	v := hooksVar("X", s)
	result := cloneLHS(v, "g")
	// Variable is not Const or Apply; should be returned unchanged.
	if result != v {
		t.Error("expected same variable back")
	}
}

// --- appendLF ---

func TestAppendLF(t *testing.T) {
	result := appendLF(autoTestAstCfg, nil, "foo", True, Location{})
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result))
	}
	lf := result[0]
	if lf == nil {
		t.Fatal("expected non-nil LabeledFormula")
	}
	if a, ok := lf.Label.(*Atom); !ok || a.Rep != "foo" {
		t.Errorf("expected label Atom with rep 'foo', got %v", lf.Label)
	}
}
