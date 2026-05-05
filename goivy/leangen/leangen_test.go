package leangen

import (
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
	"testing"
)

func TestSortToStringBool(t *testing.T) {
	got, err := SortToString(goivy.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	if got != "sort.bool" {
		t.Errorf("expected sort.bool, got %q", got)
	}
}

func TestSortToStringUninterpreted(t *testing.T) {
	s := &goivy.UninterpretedSort{Name: "node"}
	got, err := SortToString(s)
	if err != nil {
		t.Fatal(err)
	}
	if got != `(sort.ui "node")` {
		t.Errorf("unexpected: %q", got)
	}
}

func TestSortToStringFunction(t *testing.T) {
	fs, err := goivy.NewFunctionSort(
		&goivy.UninterpretedSort{Name: "a"},
		&goivy.UninterpretedSort{Name: "b"}, goivy.
			Boolean,
	)
	if err != nil {
		t.Fatal(err)
	}
	got, err2 := SortToString(fs)
	if err2 != nil {
		t.Fatal(err2)
	}
	if !strings.Contains(got, "\u00d7") {
		t.Errorf("expected cross product symbol, got %q", got)
	}
	if !strings.Contains(got, "\u21a6") {
		t.Errorf("expected mapsto arrow, got %q", got)
	}
}

func TestSortToStringEnumerated(t *testing.T) {
	s := &goivy.LogicEnumeratedSort{Name: "e", Extension: []string{"a", "b"}}
	_, err := SortToString(s)
	if err == nil {
		t.Error("expected error for enumerated sort")
	}
}

func TestEmitSymbolDef(t *testing.T) {
	g := NewGenerator()
	err := g.EmitSymbolDef("myrel", goivy.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "def myrel := mk_cnst") {
		t.Errorf("unexpected: %q", got)
	}
	if !strings.Contains(got, "sort.bool") {
		t.Errorf("missing sort: %q", got)
	}
}

func TestEmitExprConst(t *testing.T) {
	g := NewGenerator()
	c := goivy.NewConst("foo", goivy.Boolean)
	err := g.EmitExpr(c)
	if err != nil {
		t.Fatal(err)
	}
	if g.String() != "foo" {
		t.Errorf("expected foo, got %q", g.String())
	}
}

func TestEmitExprVar(t *testing.T) {
	g := NewGenerator()
	v, _ := goivy.NewVariable("X", goivy.Boolean)
	err := g.EmitExpr(v)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, `"X"`) {
		t.Errorf("expected var name, got %q", got)
	}
}

func TestEmitExprNot(t *testing.T) {
	g := NewGenerator()
	c := goivy.NewConst("p", goivy.Boolean)
	n, _ := goivy.NewNot(c)
	err := g.EmitExpr(n)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.HasPrefix(got, "\u00ac") {
		t.Errorf("expected NOT sign prefix, got %q", got)
	}
}

func TestEmitExprAndEmpty(t *testing.T) {
	g := NewGenerator()
	a, _ := goivy.NewAnd()
	err := g.EmitExpr(a)
	if err != nil {
		t.Fatal(err)
	}
	if g.String() != "ltrue" {
		t.Errorf("expected ltrue, got %q", g.String())
	}
}

func TestEmitExprOrEmpty(t *testing.T) {
	g := NewGenerator()
	o, _ := goivy.NewOr()
	err := g.EmitExpr(o)
	if err != nil {
		t.Fatal(err)
	}
	if g.String() != "lfalse" {
		t.Errorf("expected lfalse, got %q", g.String())
	}
}

func TestEmitExprAnd(t *testing.T) {
	g := NewGenerator()
	c1 := goivy.NewConst("p", goivy.Boolean)
	c2 := goivy.NewConst("q", goivy.Boolean)
	a, _ := goivy.NewAnd(c1, c2)
	err := g.EmitExpr(a)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "\u2227") {
		t.Errorf("expected AND sign, got %q", got)
	}
}

func TestEmitExprOr(t *testing.T) {
	g := NewGenerator()
	c1 := goivy.NewConst("p", goivy.Boolean)
	c2 := goivy.NewConst("q", goivy.Boolean)
	o, _ := goivy.NewOr(c1, c2)
	err := g.EmitExpr(o)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "\u2228") {
		t.Errorf("expected OR sign, got %q", got)
	}
}

func TestEmitExprImplies(t *testing.T) {
	g := NewGenerator()
	c1 := goivy.NewConst("p", goivy.Boolean)
	c2 := goivy.NewConst("q", goivy.Boolean)
	imp, _ := goivy.NewImplies(c1, c2)
	err := g.EmitExpr(imp)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "\u21d2") {
		t.Errorf("expected implies arrow, got %q", got)
	}
}

func TestEmitExprEq(t *testing.T) {
	g := NewGenerator()
	c1 := goivy.NewConst("a", goivy.Boolean)
	c2 := goivy.NewConst("b", goivy.Boolean)
	eq, _ := goivy.NewEq(c1, c2)
	err := g.EmitExpr(eq)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "fmla.eq") {
		t.Errorf("expected fmla.eq, got %q", got)
	}
}

func TestEmitExprForAll(t *testing.T) {
	g := NewGenerator()
	v, _ := goivy.NewVariable("X", goivy.Boolean)
	body := goivy.NewConst("p", goivy.Boolean)
	fa, _ := goivy.NewForAll([]*goivy.LogicVariable{v}, body)
	err := g.EmitExpr(fa)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "fmla.proj") {
		t.Errorf("expected fmla.proj, got %q", got)
	}
	// ForAll uses double negation encoding
	if strings.Count(got, "\u00ac") < 2 {
		t.Errorf("expected double negation for forall, got %q", got)
	}
}

func TestEmitExprExists(t *testing.T) {
	g := NewGenerator()
	v, _ := goivy.NewVariable("X", goivy.Boolean)
	body := goivy.NewConst("p", goivy.Boolean)
	ex, _ := goivy.NewExists([]*goivy.LogicVariable{v}, body)
	err := g.EmitExpr(ex)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "fmla.proj") {
		t.Errorf("expected fmla.proj, got %q", got)
	}
}

func TestEmitExprLambda(t *testing.T) {
	g := NewGenerator()
	v, _ := goivy.NewVariable("X", goivy.Boolean)
	body := goivy.NewConst("p", goivy.Boolean)
	lam, _ := goivy.NewLambda([]*goivy.LogicVariable{v}, body)
	err := g.EmitExpr(lam)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "fmla.lambda") {
		t.Errorf("expected fmla.lambda, got %q", got)
	}
}

func TestEmitExprIte(t *testing.T) {
	g := NewGenerator()
	cond := goivy.NewConst("c", goivy.Boolean)
	then := goivy.NewConst("t", goivy.Boolean)
	els := goivy.NewConst("e", goivy.Boolean)
	ite, _ := goivy.NewIte(cond, then, els)
	err := g.EmitExpr(ite)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "ite_fmla") {
		t.Errorf("expected ite_fmla, got %q", got)
	}
}

func TestEmitActionAssign(t *testing.T) {
	g := NewGenerator()
	lhs := goivy.NewConst("x", goivy.Boolean)
	rhs := goivy.NewConst("y", goivy.Boolean)
	a := goivy.NewAssignAction(lhs, rhs)
	err := g.EmitAction(a)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "::=") {
		t.Errorf("expected ::= in assign, got %q", got)
	}
}

func TestEmitActionIf(t *testing.T) {
	g := NewGenerator()
	cond := goivy.NewConst("c", goivy.Boolean)
	thenAct := goivy.NewAssignAction(goivy.
		NewConst("x", goivy.Boolean), goivy.
		NewConst("y", goivy.Boolean),
	)
	ifAct := goivy.NewIfAction(cond, thenAct)
	err := g.EmitAction(ifAct)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "PL.pterm.ite") {
		t.Errorf("expected PL.pterm.ite, got %q", got)
	}
	if !strings.Contains(got, "PL.pterm.skip") {
		t.Errorf("expected PL.pterm.skip for missing else, got %q", got)
	}
}

func TestGenerateProgram(t *testing.T) {
	g := NewGenerator()
	syms := []SymbolDef{
		{Name: "r", Sort: goivy.Boolean},
	}
	actMap := map[string]goivy.ActionsAction{
		"act1": goivy.NewAssignAction(goivy.
			NewConst("x", goivy.Boolean), goivy.
			NewConst("y", goivy.Boolean),
		),
	}
	exports := []string{"act1"}

	err := g.GenerateProgram(syms, actMap, exports, "test")
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "namespace ivy") {
		t.Error("missing preamble")
	}
	if !strings.Contains(got, "end ivy") {
		t.Error("missing postamble")
	}
	if !strings.Contains(got, "PL.pterm.call") {
		t.Error("missing export call")
	}
}

func TestGenerateProgramNoExports(t *testing.T) {
	g := NewGenerator()
	err := g.GenerateProgram(nil, map[string]goivy.ActionsAction{}, nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "PL.pterm.skip") {
		t.Error("expected skip for empty exports")
	}
}

func TestEmitExprIff(t *testing.T) {
	g := NewGenerator()
	c1 := goivy.NewConst("a", goivy.Boolean)
	c2 := goivy.NewConst("b", goivy.Boolean)
	iff, _ := goivy.NewIff(c1, c2)
	err := g.EmitExpr(iff)
	if err != nil {
		t.Fatal(err)
	}
	got := g.String()
	if !strings.Contains(got, "fmla.eq") {
		t.Errorf("expected fmla.eq for Iff, got %q", got)
	}
}

func FuzzSortToString(f *testing.F) {
	f.Add("alpha")
	f.Add("")
	f.Add("node.link")
	f.Add(`with"quotes`)

	f.Fuzz(func(t *testing.T, name string) {
		s := &goivy.UninterpretedSort{Name: name}
		got, err := SortToString(s)
		if err != nil {
			t.Skip()
		}
		if !strings.Contains(got, "sort.ui") {
			t.Errorf("expected sort.ui, got %q", got)
		}
	})
}
