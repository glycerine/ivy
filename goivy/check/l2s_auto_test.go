package check

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

var autoTestAstCfg = ast.NewAstConfig()

// --- buildOrExpr ---

func TestBuildOrExpr_Empty(t *testing.T) {
	result := buildOrExpr(nil)
	or, ok := result.(*lg.Or)
	if !ok {
		t.Fatalf("expected *lg.Or, got %T", result)
	}
	if len(or.Terms) != 0 {
		t.Errorf("expected 0 terms, got %d", len(or.Terms))
	}
}

func TestBuildOrExpr_Single(t *testing.T) {
	a := lg.NewConst("a", lg.Boolean)
	result := buildOrExpr([]lg.Expr{a})
	// Single element should return the element directly, not wrapped in Or.
	if result != a {
		t.Errorf("expected unwrapped element, got %T", result)
	}
}

func TestBuildOrExpr_Multiple(t *testing.T) {
	a := lg.NewConst("a", lg.Boolean)
	b := lg.NewConst("b", lg.Boolean)
	result := buildOrExpr([]lg.Expr{a, b})
	or, ok := result.(*lg.Or)
	if !ok {
		t.Fatalf("expected *lg.Or, got %T", result)
	}
	if len(or.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(or.Terms))
	}
}

// --- isGloballyOrNotEventually ---

func TestIsGloballyOrNotEventually_Globally(t *testing.T) {
	g := &lg.Globally{Body: lg.True}
	if !isGloballyOrNotEventually(g) {
		t.Error("expected true for Globally")
	}
}

func TestIsGloballyOrNotEventually_NotEventually(t *testing.T) {
	ne := &lg.Not{Body: &lg.Eventually{Body: lg.True}}
	if !isGloballyOrNotEventually(ne) {
		t.Error("expected true for Not{Eventually}")
	}
}

func TestIsGloballyOrNotEventually_Eventually(t *testing.T) {
	e := &lg.Eventually{Body: lg.True}
	if isGloballyOrNotEventually(e) {
		t.Error("expected false for Eventually")
	}
}

func TestIsGloballyOrNotEventually_Plain(t *testing.T) {
	c := lg.NewConst("c", lg.Boolean)
	if isGloballyOrNotEventually(c) {
		t.Error("expected false for plain Const")
	}
}

func TestIsGloballyOrNotEventually_NotGlobally(t *testing.T) {
	ng := &lg.Not{Body: &lg.Globally{Body: lg.True}}
	if isGloballyOrNotEventually(ng) {
		t.Error("expected false for Not{Globally}")
	}
}

// --- isEventuallyOrNotGlobally ---

func TestIsEventuallyOrNotGlobally_Eventually(t *testing.T) {
	e := &lg.Eventually{Body: lg.True}
	if !isEventuallyOrNotGlobally(e) {
		t.Error("expected true for Eventually")
	}
}

func TestIsEventuallyOrNotGlobally_NotGlobally(t *testing.T) {
	ng := &lg.Not{Body: &lg.Globally{Body: lg.True}}
	if !isEventuallyOrNotGlobally(ng) {
		t.Error("expected true for Not{Globally}")
	}
}

func TestIsEventuallyOrNotGlobally_Globally(t *testing.T) {
	g := &lg.Globally{Body: lg.True}
	if isEventuallyOrNotGlobally(g) {
		t.Error("expected false for Globally")
	}
}

func TestIsEventuallyOrNotGlobally_Plain(t *testing.T) {
	c := lg.NewConst("c", lg.Boolean)
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
	s := &lg.UninterpretedSort{Name: "S"}
	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	f := lg.NewConst("f", fs)
	g := lg.NewConst("g", fs)
	x := hooksVar("X", s)
	y := hooksVar("Y", s)

	a := &lg.Eq{T1: lg.MustApply(f, x), T2: lg.True}
	b := &lg.Eq{T1: lg.MustApply(g, y), T2: lg.True}

	if !sameLHSSort(a, b) {
		t.Error("expected true for same sort")
	}
}

func TestSameLHSSort_DiffSort(t *testing.T) {
	s1 := &lg.UninterpretedSort{Name: "S"}
	s2 := &lg.UninterpretedSort{Name: "T"}
	fs1, _ := lg.NewFunctionSort(s1, lg.Boolean)
	fs2, _ := lg.NewFunctionSort(s2, lg.Boolean)
	f := lg.NewConst("f", fs1)
	g := lg.NewConst("g", fs2)
	x := hooksVar("X", s1)
	y := hooksVar("Y", s2)

	a := &lg.Eq{T1: lg.MustApply(f, x), T2: lg.True}
	b := &lg.Eq{T1: lg.MustApply(g, y), T2: lg.True}

	if sameLHSSort(a, b) {
		t.Error("expected false for different sorts")
	}
}

func TestSameLHSSort_ConstLHS(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "S"}
	c := lg.NewConst("c", s)
	d := lg.NewConst("d", s)
	a := &lg.Eq{T1: c, T2: lg.True}
	b := &lg.Eq{T1: d, T2: lg.True}
	if !sameLHSSort(a, b) {
		t.Error("expected true for same Const sort")
	}
}

// --- cloneLHS ---

func TestCloneLHS_Const(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "S"}
	c := lg.NewConst("f", s)
	result := cloneLHS(c, "g")
	rc, ok := result.(*lg.Const)
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
	s := &lg.UninterpretedSort{Name: "S"}
	fs, _ := lg.NewFunctionSort(s, lg.Boolean)
	f := lg.NewConst("f", fs)
	x := hooksVar("X", s)
	app := lg.MustApply(f, x)

	result := cloneLHS(app, "g")
	ra, ok := result.(*lg.Apply)
	if !ok {
		t.Fatalf("expected *lg.Apply, got %T", result)
	}
	if rc, ok := ra.Func.(*lg.Const); !ok || rc.Name != "g" {
		t.Error("expected Func name 'g'")
	}
	if len(ra.Terms) != 1 {
		t.Errorf("expected 1 term, got %d", len(ra.Terms))
	}
}

func TestCloneLHS_Other(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "S"}
	v := hooksVar("X", s)
	result := cloneLHS(v, "g")
	// Variable is not Const or Apply; should be returned unchanged.
	if result != v {
		t.Error("expected same variable back")
	}
}

// --- appendLF ---

func TestAppendLF(t *testing.T) {
	result := appendLF(autoTestAstCfg, nil, "foo", lg.True)
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result))
	}
	lf := result[0]
	if lf == nil {
		t.Fatal("expected non-nil LabeledFormula")
	}
	if c, ok := lf.Label.(*lg.Const); !ok || c.Name != "foo" {
		t.Errorf("expected label name 'foo', got %v", lf.Label)
	}
}
