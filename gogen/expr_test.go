package gogen

import (
	"strings"
	"testing"
	"unicode/utf8"

	lg "github.com/glycerine/goivy/logic"
)

func makeVar(name string, s lg.Sort) *lg.Variable {
	v, err := lg.NewVariable(name, s)
	if err != nil {
		panic(err)
	}
	return v
}

func TestEmitExpr_Var(t *testing.T) {
	e := NewExprEmitter()
	v := makeVar("X", &lg.BooleanSort{})
	got, err := e.EmitExpr(v)
	if err != nil {
		t.Fatal(err)
	}
	if got != "x" {
		t.Errorf("EmitExpr(Var X) = %q, want %q", got, "x")
	}
}

func TestEmitExpr_Const(t *testing.T) {
	e := NewExprEmitter()
	c := lg.NewConst("red", &lg.EnumeratedSort{Name: "color", Extension: []string{"red"}})
	got, err := e.EmitExpr(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Red" {
		t.Errorf("EmitExpr(Const red) = %q, want %q", got, "Red")
	}
}

func TestEmitExpr_Eq(t *testing.T) {
	e := NewExprEmitter()
	v1 := makeVar("X", &lg.UninterpretedSort{Name: "node"})
	v2 := makeVar("Y", &lg.UninterpretedSort{Name: "node"})
	eq, err := lg.NewEq(v1, v2)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(eq)
	if err != nil {
		t.Fatal(err)
	}
	if got != "(x == y)" {
		t.Errorf("EmitExpr(Eq) = %q, want %q", got, "(x == y)")
	}
}

func TestEmitExpr_Not(t *testing.T) {
	e := NewExprEmitter()
	v := makeVar("X", &lg.BooleanSort{})
	not, err := lg.NewNot(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(not)
	if err != nil {
		t.Fatal(err)
	}
	if got != "(!x)" {
		t.Errorf("EmitExpr(Not) = %q, want %q", got, "(!x)")
	}
}

func TestEmitExpr_And_Empty(t *testing.T) {
	e := NewExprEmitter()
	and := &lg.And{}
	got, err := e.EmitExpr(and)
	if err != nil {
		t.Fatal(err)
	}
	if got != "true" {
		t.Errorf("EmitExpr(And{}) = %q, want %q", got, "true")
	}
}

func TestEmitExpr_And(t *testing.T) {
	e := NewExprEmitter()
	a := makeVar("A", &lg.BooleanSort{})
	b := makeVar("B", &lg.BooleanSort{})
	and, err := lg.NewAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(and)
	if err != nil {
		t.Fatal(err)
	}
	if got != "(a && b)" {
		t.Errorf("EmitExpr(And) = %q, want %q", got, "(a && b)")
	}
}

func TestEmitExpr_Or_Empty(t *testing.T) {
	e := NewExprEmitter()
	or := &lg.Or{}
	got, err := e.EmitExpr(or)
	if err != nil {
		t.Fatal(err)
	}
	if got != "false" {
		t.Errorf("EmitExpr(Or{}) = %q, want %q", got, "false")
	}
}

func TestEmitExpr_Or(t *testing.T) {
	e := NewExprEmitter()
	a := makeVar("A", &lg.BooleanSort{})
	b := makeVar("B", &lg.BooleanSort{})
	or, err := lg.NewOr(a, b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(or)
	if err != nil {
		t.Fatal(err)
	}
	if got != "(a || b)" {
		t.Errorf("EmitExpr(Or) = %q, want %q", got, "(a || b)")
	}
}

func TestEmitExpr_Implies(t *testing.T) {
	e := NewExprEmitter()
	a := makeVar("A", &lg.BooleanSort{})
	b := makeVar("B", &lg.BooleanSort{})
	imp, err := lg.NewImplies(a, b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(imp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "(!a || b)" {
		t.Errorf("EmitExpr(Implies) = %q, want %q", got, "(!a || b)")
	}
}

func TestEmitExpr_Iff(t *testing.T) {
	e := NewExprEmitter()
	a := makeVar("A", &lg.BooleanSort{})
	b := makeVar("B", &lg.BooleanSort{})
	iff, err := lg.NewIff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(iff)
	if err != nil {
		t.Fatal(err)
	}
	if got != "(a == b)" {
		t.Errorf("EmitExpr(Iff) = %q, want %q", got, "(a == b)")
	}
}

func TestEmitExpr_Ite_Bool(t *testing.T) {
	e := NewExprEmitter()
	cond := makeVar("C", &lg.BooleanSort{})
	then := makeVar("A", &lg.BooleanSort{})
	els := makeVar("B", &lg.BooleanSort{})
	ite, err := lg.NewIte(cond, then, els)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(ite)
	if err != nil {
		t.Fatal(err)
	}
	if got != "ite(c, a, b)" {
		t.Errorf("EmitExpr(Ite bool) = %q, want %q", got, "ite(c, a, b)")
	}
}

func TestEmitExpr_Ite_Int(t *testing.T) {
	e := NewExprEmitter()
	cond := makeVar("C", &lg.BooleanSort{})
	intSort := &lg.UninterpretedSort{Name: "val"}
	then := makeVar("A", intSort)
	els := makeVar("B", intSort)
	ite, err := lg.NewIte(cond, then, els)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(ite)
	if err != nil {
		t.Fatal(err)
	}
	if got != "iteVal[int](c, a, b)" {
		t.Errorf("EmitExpr(Ite int) = %q, want %q", got, "iteVal[int](c, a, b)")
	}
}

func TestEmitExpr_Apply_SingleArg(t *testing.T) {
	e := NewExprEmitter()
	nodeSort := &lg.UninterpretedSort{Name: "node"}
	fs, _ := lg.NewFunctionSort(nodeSort, &lg.BooleanSort{})
	fn := lg.NewConst("visited", fs)
	arg := makeVar("N", nodeSort)
	app, err := lg.NewApply(fn, arg)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(app)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Visited[n]" {
		t.Errorf("EmitExpr(Apply unary) = %q, want %q", got, "Visited[n]")
	}
}

func TestEmitExpr_Apply_MultiArg(t *testing.T) {
	e := NewExprEmitter()
	nodeSort := &lg.UninterpretedSort{Name: "node"}
	fs, _ := lg.NewFunctionSort(nodeSort, nodeSort, &lg.BooleanSort{})
	fn := lg.NewConst("edge", fs)
	a := makeVar("A", nodeSort)
	b := makeVar("B", nodeSort)
	app, err := lg.NewApply(fn, a, b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(app)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Edge[[2]int{a, b}]" {
		t.Errorf("EmitExpr(Apply binary) = %q, want %q", got, "Edge[[2]int{a, b}]")
	}
}

func TestEmitExpr_ForAll(t *testing.T) {
	e := NewExprEmitter()
	colorSort := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	v := makeVar("C", colorSort)
	body := makeVar("X", &lg.BooleanSort{})
	fa, err := lg.NewForAll([]*lg.Variable{v}, body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(fa)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "forAll_Color") {
		t.Errorf("EmitExpr(ForAll) = %q, missing forAll_Color", got)
	}
	if !strings.Contains(got, "func(c Color) bool") {
		t.Errorf("EmitExpr(ForAll) = %q, missing func signature", got)
	}
	// Check helper sort was registered.
	if _, ok := e.HelperSorts["Color"]; !ok {
		t.Error("HelperSorts missing Color")
	}
}

func TestEmitExpr_Exists(t *testing.T) {
	e := NewExprEmitter()
	nodeSort := &lg.UninterpretedSort{Name: "node"}
	v := makeVar("N", nodeSort)
	body := makeVar("X", &lg.BooleanSort{})
	ex, err := lg.NewExists([]*lg.Variable{v}, body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(ex)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "exists_Node") {
		t.Errorf("EmitExpr(Exists) = %q, missing exists_Node", got)
	}
}

func TestEmitExpr_ForAll_MultiVar(t *testing.T) {
	e := NewExprEmitter()
	nodeSort := &lg.UninterpretedSort{Name: "node"}
	v1 := makeVar("X", nodeSort)
	v2 := makeVar("Y", nodeSort)
	body := makeVar("Z", &lg.BooleanSort{})
	fa, err := lg.NewForAll([]*lg.Variable{v1, v2}, body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(fa)
	if err != nil {
		t.Fatal(err)
	}
	// Should have nested forAll calls.
	if strings.Count(got, "forAll_Node") != 2 {
		t.Errorf("EmitExpr(ForAll multi) = %q, expected 2 forAll_Node calls", got)
	}
}

func TestEmitExpr_Nil(t *testing.T) {
	e := NewExprEmitter()
	_, err := e.EmitExpr(nil)
	if err == nil {
		t.Error("expected error for nil node")
	}
}

func TestEmitExpr_Lambda(t *testing.T) {
	e := NewExprEmitter()
	nodeSort := &lg.UninterpretedSort{Name: "node"}
	v := makeVar("X", nodeSort)
	body := makeVar("Y", &lg.BooleanSort{})
	lam, err := lg.NewLambda([]*lg.Variable{v}, body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.EmitExpr(lam)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "func(x int)") {
		t.Errorf("EmitExpr(Lambda) = %q, missing func signature", got)
	}
}

func TestEmitForAllHelper(t *testing.T) {
	w := NewCodeWriter()
	s := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	EmitForAllHelper(w, s)
	out := w.String()
	if !strings.Contains(out, "func forAll_Color(f func(Color) bool) bool") {
		t.Errorf("EmitForAllHelper missing signature:\n%s", out)
	}
	if !strings.Contains(out, "allColor") {
		t.Errorf("EmitForAllHelper missing allColor iteration:\n%s", out)
	}
	if !strings.Contains(out, "return false") {
		t.Errorf("EmitForAllHelper missing return false:\n%s", out)
	}
}

func TestEmitExistsHelper(t *testing.T) {
	w := NewCodeWriter()
	s := &lg.EnumeratedSort{Name: "color", Extension: []string{"red"}}
	EmitExistsHelper(w, s)
	out := w.String()
	if !strings.Contains(out, "func exists_Color(f func(Color) bool) bool") {
		t.Errorf("EmitExistsHelper missing signature:\n%s", out)
	}
	if !strings.Contains(out, "return true") {
		t.Errorf("EmitExistsHelper missing return true:\n%s", out)
	}
}

func TestEmitIteHelper(t *testing.T) {
	w := NewCodeWriter()
	EmitIteHelper(w)
	out := w.String()
	if !strings.Contains(out, "func ite(cond, a, b bool) bool") {
		t.Errorf("EmitIteHelper missing signature:\n%s", out)
	}
}

func TestCodeWriter_Basic(t *testing.T) {
	w := NewCodeWriter()
	w.Line("package main")
	w.BlankLine()
	w.OpenBlock("func main() {")
	w.Line(`fmt.Println("hello")`)
	w.CloseBlock()

	out := w.String()
	if !strings.Contains(out, "package main") {
		t.Error("missing package line")
	}
	if !strings.Contains(out, "\tfmt.Println") {
		t.Error("missing indented println")
	}
	if !strings.Contains(out, "}") {
		t.Error("missing close brace")
	}
}

func TestCodeWriter_NestedBlocks(t *testing.T) {
	w := NewCodeWriter()
	w.OpenBlock("if true {")
	w.OpenBlock("for i := 0; i < 10; i++ {")
	w.Line("x++")
	w.CloseBlock()
	w.CloseBlock()
	out := w.String()
	if !strings.Contains(out, "\t\tx++") {
		t.Errorf("expected double-indented x++:\n%s", out)
	}
}

func TestCodeWriter_Linef(t *testing.T) {
	w := NewCodeWriter()
	w.Linef("x := %d", 42)
	if !strings.Contains(w.String(), "x := 42") {
		t.Errorf("Linef didn't format: %q", w.String())
	}
}

func TestCodeWriter_IndentDedent(t *testing.T) {
	w := NewCodeWriter()
	w.Indent()
	w.Indent()
	w.Line("deep")
	w.Dedent()
	w.Line("less")
	out := w.String()
	if !strings.Contains(out, "\t\tdeep") {
		t.Errorf("expected double-indented deep:\n%s", out)
	}
	if !strings.Contains(out, "\tless") {
		t.Errorf("expected single-indented less:\n%s", out)
	}
}

func TestCodeWriter_DedentFloor(t *testing.T) {
	w := NewCodeWriter()
	w.Dedent() // should not go negative
	w.Dedent()
	w.Line("ok")
	if w.IndentLevel() != 0 {
		t.Errorf("indent level should be 0 after double dedent, got %d", w.IndentLevel())
	}
}

func TestSortHelperName(t *testing.T) {
	tests := []struct {
		sort lg.Sort
		want string
	}{
		{&lg.BooleanSort{}, "Bool"},
		{&lg.EnumeratedSort{Name: "color"}, "Color"},
		{&lg.RangeSort{Name: "idx", Lb: lg.NumeralBound{Value: "0"}, Ub: lg.NumeralBound{Value: "0"}}, "Idx"},
		{&lg.UninterpretedSort{Name: "node"}, "Node"},
	}
	for _, tt := range tests {
		got := sortHelperName(tt.sort)
		if got != tt.want {
			t.Errorf("sortHelperName(%T{%s}) = %q, want %q", tt.sort, tt.sort, got, tt.want)
		}
	}
}

func TestAllValsExpr(t *testing.T) {
	tests := []struct {
		sort lg.Sort
		want string
	}{
		{&lg.BooleanSort{}, "[2]bool{false, true}"},
		{&lg.EnumeratedSort{Name: "color"}, "allColor"},
		{&lg.UninterpretedSort{Name: "node"}, "allNode"},
	}
	for _, tt := range tests {
		got := allValsExpr(tt.sort)
		if got != tt.want {
			t.Errorf("allValsExpr(%T) = %q, want %q", tt.sort, got, tt.want)
		}
	}
}

// FuzzGoExportedName ensures goExportedName doesn't panic on arbitrary input.
func FuzzGoExportedName(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("X")
	f.Add("foo.bar.baz")
	f.Add("123")
	f.Add("_under")

	f.Fuzz(func(t *testing.T, s string) {
		result := goExportedName(s)
		if s == "" {
			if result != "" {
				t.Errorf("goExportedName(%q) = %q, want empty", s, result)
			}
			return
		}
		// Result should be valid UTF-8.
		if !utf8.ValidString(result) {
			t.Errorf("goExportedName(%q) produced invalid UTF-8: %q", s, result)
		}
		// Should not contain dots.
		if strings.Contains(result, ".") {
			t.Errorf("goExportedName(%q) still contains dot: %q", s, result)
		}
	})
}
