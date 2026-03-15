package parser

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

var v17 = lexer.Version{1, 7}

func parse(t *testing.T, input string) []ast.Node {
	t.Helper()
	p := New(input, v17)
	decls, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return decls
}

func parseExpr(t *testing.T, input string) ast.Node {
	t.Helper()
	p := New(input, v17)
	return p.parseExpr(0)
}

// --- Expression parsing tests ---

func TestParseSymbol(t *testing.T) {
	e := parseExpr(t, "foo")
	sym, ok := e.(*ast.Symbol)
	if !ok {
		t.Fatalf("expected Symbol, got %T", e)
	}
	if sym.Rep != "foo" {
		t.Errorf("got %q", sym.Rep)
	}
}

func TestParseVariable(t *testing.T) {
	e := parseExpr(t, "X")
	v, ok := e.(*ast.Variable)
	if !ok {
		t.Fatalf("expected Variable, got %T", e)
	}
	if v.Rep != "X" {
		t.Errorf("got %q", v.Rep)
	}
}

func TestParseAtom(t *testing.T) {
	e := parseExpr(t, "f(x, y)")
	a, ok := e.(*ast.Atom)
	if !ok {
		t.Fatalf("expected Atom, got %T", e)
	}
	if a.Rep != "f" {
		t.Errorf("got %q", a.Rep)
	}
	if len(a.Terms) != 2 {
		t.Errorf("got %d args", len(a.Terms))
	}
}

func TestParseTrue(t *testing.T) {
	e := parseExpr(t, "true")
	if !ast.IsTrue(e) {
		t.Errorf("expected true, got %T", e)
	}
}

func TestParseFalse(t *testing.T) {
	e := parseExpr(t, "false")
	if !ast.IsFalse(e) {
		t.Errorf("expected false, got %T", e)
	}
}

func TestParseNot(t *testing.T) {
	e := parseExpr(t, "~p")
	n, ok := e.(*ast.Not)
	if !ok {
		t.Fatalf("expected Not, got %T", e)
	}
	if n.String() != "~p" {
		t.Errorf("got %q", n.String())
	}
}

func TestParseAnd(t *testing.T) {
	e := parseExpr(t, "p & q")
	a, ok := e.(*ast.And)
	if !ok {
		t.Fatalf("expected And, got %T", e)
	}
	if len(a.Terms) != 2 {
		t.Errorf("got %d terms", len(a.Terms))
	}
}

func TestParseOr(t *testing.T) {
	e := parseExpr(t, "p | q")
	o, ok := e.(*ast.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", e)
	}
	if len(o.Terms) != 2 {
		t.Errorf("got %d terms", len(o.Terms))
	}
}

func TestParseImplies(t *testing.T) {
	e := parseExpr(t, "p -> q")
	i, ok := e.(*ast.Implies)
	if !ok {
		t.Fatalf("expected Implies, got %T", e)
	}
	_ = i
}

func TestParseIff(t *testing.T) {
	e := parseExpr(t, "p <-> q")
	f, ok := e.(*ast.Iff)
	if !ok {
		t.Fatalf("expected Iff, got %T", e)
	}
	_ = f
}

func TestParseEq(t *testing.T) {
	e := parseExpr(t, "x = y")
	a, ok := e.(*ast.Atom)
	if !ok {
		t.Fatalf("expected Atom, got %T", e)
	}
	if a.Rep != "=" {
		t.Errorf("got rep %q", a.Rep)
	}
}

func TestParseNeq(t *testing.T) {
	e := parseExpr(t, "x ~= y")
	n, ok := e.(*ast.Not)
	if !ok {
		t.Fatalf("expected Not, got %T", e)
	}
	if _, ok := n.Body.(*ast.Atom); !ok {
		t.Errorf("expected Atom inside Not, got %T", n.Body)
	}
}

func TestParseArithmetic(t *testing.T) {
	e := parseExpr(t, "x + y * z")
	// Should parse as x + (y * z) due to precedence
	a, ok := e.(*ast.Atom)
	if !ok {
		t.Fatalf("expected Atom (+), got %T", e)
	}
	if a.Rep != "+" {
		t.Errorf("got rep %q", a.Rep)
	}
	if _, ok := a.Terms[1].(*ast.Atom); !ok {
		t.Errorf("rhs should be Atom (*), got %T", a.Terms[1])
	}
}

func TestParsePrecedence(t *testing.T) {
	// p & q -> r should be (p & q) -> r
	e := parseExpr(t, "p & q -> r")
	imp, ok := e.(*ast.Implies)
	if !ok {
		t.Fatalf("expected Implies at top, got %T", e)
	}
	if _, ok := imp.T1.(*ast.And); !ok {
		t.Errorf("lhs should be And, got %T", imp.T1)
	}
}

func TestParseForall(t *testing.T) {
	e := parseExpr(t, "forall X:t. p(X)")
	f, ok := e.(*ast.Forall)
	if !ok {
		t.Fatalf("expected Forall, got %T", e)
	}
	if len(f.Bounds) != 1 {
		t.Errorf("got %d bounds", len(f.Bounds))
	}
}

func TestParseExists(t *testing.T) {
	e := parseExpr(t, "exists X:t, Y:t. f(X) = f(Y)")
	ex, ok := e.(*ast.Exists)
	if !ok {
		t.Fatalf("expected Exists, got %T", e)
	}
	if len(ex.Bounds) != 2 {
		t.Errorf("got %d bounds", len(ex.Bounds))
	}
}

func TestParseIte(t *testing.T) {
	e := parseExpr(t, "x if c else y")
	ite, ok := e.(*ast.Ite)
	if !ok {
		t.Fatalf("expected Ite, got %T", e)
	}
	_ = ite
}

func TestParseDot(t *testing.T) {
	e := parseExpr(t, "a.b")
	d, ok := e.(*ast.Dot)
	if !ok {
		t.Fatalf("expected Dot, got %T", e)
	}
	_ = d
}

func TestParseOld(t *testing.T) {
	e := parseExpr(t, "old x")
	o, ok := e.(*ast.Old)
	if !ok {
		t.Fatalf("expected Old, got %T", e)
	}
	_ = o
}

func TestParseGlobally(t *testing.T) {
	e := parseExpr(t, "globally p")
	g, ok := e.(*ast.Globally)
	if !ok {
		t.Fatalf("expected Globally, got %T", e)
	}
	_ = g
}

func TestParseSortAnnotation(t *testing.T) {
	e := parseExpr(t, "X:t")
	v, ok := e.(*ast.Variable)
	if !ok {
		t.Fatalf("expected Variable, got %T", e)
	}
	if v.VSort == nil {
		t.Error("expected sort annotation")
	}
}

func TestParseParenthesized(t *testing.T) {
	e := parseExpr(t, "(p & q) | r")
	o, ok := e.(*ast.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", e)
	}
	if _, ok := o.Terms[0].(*ast.And); !ok {
		t.Errorf("first term should be And, got %T", o.Terms[0])
	}
}

func TestParseNaryAnd(t *testing.T) {
	e := parseExpr(t, "a & b & c & d")
	a, ok := e.(*ast.And)
	if !ok {
		t.Fatalf("expected And, got %T", e)
	}
	if len(a.Terms) != 4 {
		t.Errorf("got %d terms, want 4", len(a.Terms))
	}
}

func TestParseComparison(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"x <= y", "<="},
		{"x < y", "<"},
		{"x >= y", ">="},
		{"x > y", ">"},
	}
	for _, tc := range tests {
		e := parseExpr(t, tc.input)
		a, ok := e.(*ast.Atom)
		if !ok {
			t.Errorf("%s: expected Atom, got %T", tc.input, e)
			continue
		}
		if a.Rep != tc.op {
			t.Errorf("%s: got op %q, want %q", tc.input, a.Rep, tc.op)
		}
	}
}

func TestParseSome(t *testing.T) {
	e := parseExpr(t, "some X:t. p(X)")
	se, ok := e.(*ast.SomeExpr)
	if !ok {
		t.Fatalf("expected SomeExpr, got %T", e)
	}
	_ = se
}

func TestParseThis(t *testing.T) {
	e := parseExpr(t, "this")
	_, ok := e.(*ast.This)
	if !ok {
		t.Fatalf("expected This, got %T", e)
	}
}

// --- Declaration parsing tests ---

func TestParseTypeDecl(t *testing.T) {
	decls := parse(t, "type node")
	if len(decls) != 1 {
		t.Fatalf("got %d decls, want 1", len(decls))
	}
	_, ok := decls[0].(*ast.TypeDecl)
	if !ok {
		t.Fatalf("expected TypeDecl, got %T", decls[0])
	}
}

func TestParseEnumType(t *testing.T) {
	decls := parse(t, "type color = {red, green, blue}")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	td, ok := decls[0].(*ast.TypeDecl)
	if !ok {
		t.Fatalf("expected TypeDecl, got %T", decls[0])
	}
	_ = td
}

func TestParseRelationDecl(t *testing.T) {
	// In Python Ivy, "relation" produces ConstantDecl (relations are just
	// constants with Boolean range). Our Go parser matches this.
	decls := parse(t, "relation link(X:node, Y:node)")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("expected ConstantDecl (matching Python), got %T", decls[0])
	}
}

func TestParseIndividualDecl(t *testing.T) {
	decls := parse(t, "individual x : node")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("expected ConstantDecl, got %T", decls[0])
	}
}

func TestParseAxiomDecl(t *testing.T) {
	decls := parse(t, "axiom [refl] forall X:t. r(X, X)")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	ax, ok := decls[0].(*ast.AxiomDecl)
	if !ok {
		t.Fatalf("expected AxiomDecl, got %T", decls[0])
	}
	_ = ax
}

func TestParseActionDecl(t *testing.T) {
	decls := parse(t, `action send(src:node, dst:node) = {
		link(src, dst) := true
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.ActionDecl)
	if !ok {
		t.Fatalf("expected ActionDecl, got %T", decls[0])
	}
}

func TestParseActionWithReturns(t *testing.T) {
	decls := parse(t, `action get(x:t) returns (y:t) = {
		y := x
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseInitDecl(t *testing.T) {
	decls := parse(t, `init {
		x := 0
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.InitDecl)
	if !ok {
		t.Fatalf("expected InitDecl, got %T", decls[0])
	}
}

func TestParseModuleDecl(t *testing.T) {
	decls := parse(t, `module counter(t) = {
		individual val : t
		action increment = {
			val := val + 1
		}
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.ModuleDecl)
	if !ok {
		t.Fatalf("expected ModuleDecl, got %T", decls[0])
	}
}

func TestParseObjectDecl(t *testing.T) {
	decls := parse(t, `object counter = {
		individual val : nat
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.ObjectDecl)
	if !ok {
		t.Fatalf("expected ObjectDecl, got %T", decls[0])
	}
}

func TestParseExportDecl(t *testing.T) {
	decls := parse(t, "export send")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.ExportDecl)
	if !ok {
		t.Fatalf("expected ExportDecl, got %T", decls[0])
	}
}

func TestParseImportDecl(t *testing.T) {
	decls := parse(t, "import recv")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.ImportDecl)
	if !ok {
		t.Fatalf("expected ImportDecl, got %T", decls[0])
	}
}

func TestParseInstantiateDecl(t *testing.T) {
	decls := parse(t, "instantiate counter(nat)")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.InstantiateDecl)
	if !ok {
		t.Fatalf("expected InstantiateDecl, got %T", decls[0])
	}
}

func TestParsePropertyDecl(t *testing.T) {
	decls := parse(t, "property [safe] forall X:node. ~link(X, X)")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.PropertyDecl)
	if !ok {
		t.Fatalf("expected PropertyDecl, got %T", decls[0])
	}
}

func TestParseConjectureDecl(t *testing.T) {
	decls := parse(t, "conjecture forall X:node. ~link(X, X)")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseInterpretDecl(t *testing.T) {
	decls := parse(t, "interpret t -> int")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.InterpretDecl)
	if !ok {
		t.Fatalf("expected InterpretDecl, got %T", decls[0])
	}
}

func TestParseVariantDecl(t *testing.T) {
	decls := parse(t, "variant msg of packet")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseDefinitionDecl(t *testing.T) {
	decls := parse(t, "definition succ(X:nat) = X + 1")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.DefinitionDecl)
	if !ok {
		t.Fatalf("expected DefinitionDecl, got %T", decls[0])
	}
}

// --- Action parsing tests ---

func TestParseIfAction(t *testing.T) {
	decls := parse(t, `action test = {
		if x = 0 {
			y := 1
		} else {
			y := 2
		}
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseWhileAction(t *testing.T) {
	decls := parse(t, `action test = {
		while x > 0 {
			x := x - 1
		}
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseLocalAction(t *testing.T) {
	decls := parse(t, `action test = {
		local tmp:t {
			tmp := x
		}
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseAssignment(t *testing.T) {
	decls := parse(t, `action test = {
		x := y + 1
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseMultipleStatements(t *testing.T) {
	decls := parse(t, `action test = {
		x := 1;
		y := 2;
		z := x + y
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseAssumeAssert(t *testing.T) {
	decls := parse(t, `action test = {
		assume x > 0;
		assert x > 0
	}`)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

// --- Full program tests ---

func TestParseFullProgram(t *testing.T) {
	input := `
type node
relation link(X:node, Y:node)
individual root : node

axiom [refl] forall X:node. link(X, X)
axiom [trans] forall X:node, Y:node, Z:node.
    link(X, Y) & link(Y, Z) -> link(X, Z)

action connect(src:node, dst:node) = {
    link(src, dst) := true
}

export connect

conjecture forall X:node. link(root, X)
`
	decls := parse(t, input)
	if len(decls) < 5 {
		t.Errorf("got only %d decls, expected at least 5", len(decls))
	}
}

func TestParseIsolateWithBody(t *testing.T) {
	input := `
isolate server = {
    type request
    action handle(r:request) = {
        assume true
    }
}
`
	decls := parse(t, input)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

func TestParseAttributeDecl(t *testing.T) {
	decls := parse(t, "attribute server.weight = 10")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
	_, ok := decls[0].(*ast.AttributeDecl)
	if !ok {
		t.Fatalf("expected AttributeDecl, got %T", decls[0])
	}
}

func TestParseBeforeAction(t *testing.T) {
	// Python produces both ActionDecl and MixinDecl for "before name { ... }"
	input := `before send {
		assert link(src, dst)
	}`
	decls := parse(t, input)
	if len(decls) != 2 {
		t.Fatalf("got %d decls, want 2 (ActionDecl + MixinDecl)", len(decls))
	}
	if _, ok := decls[0].(*ast.ActionDecl); !ok {
		t.Fatalf("decls[0]: expected ActionDecl, got %T", decls[0])
	}
	if _, ok := decls[1].(*ast.MixinDecl); !ok {
		t.Fatalf("decls[1]: expected MixinDecl, got %T", decls[1])
	}
}

func TestParseAfterAction(t *testing.T) {
	// Python produces both ActionDecl and MixinDecl for "after name { ... }"
	input := `after recv {
		link(src, dst) := true
	}`
	decls := parse(t, input)
	if len(decls) != 2 {
		t.Fatalf("got %d decls, want 2 (ActionDecl + MixinDecl)", len(decls))
	}
	if _, ok := decls[0].(*ast.ActionDecl); !ok {
		t.Fatalf("decls[0]: expected ActionDecl, got %T", decls[0])
	}
	if _, ok := decls[1].(*ast.MixinDecl); !ok {
		t.Fatalf("decls[1]: expected MixinDecl, got %T", decls[1])
	}
}

func TestParseImplement(t *testing.T) {
	input := `implement send {
		link(src, dst) := true
	}`
	decls := parse(t, input)
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

// --- Error handling ---

func TestParseError(t *testing.T) {
	p := New("type = bad", v17)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("got %q", err.Error())
	}
}

// --- Struct type ---

func TestParseStructType(t *testing.T) {
	decls := parse(t, "type point = struct { x:nat, y:nat }")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

// --- Range type ---

func TestParseRangeType(t *testing.T) {
	decls := parse(t, "interpret idx -> {0..15}")
	if len(decls) != 1 {
		t.Fatalf("got %d decls", len(decls))
	}
}

// --- Fuzz tests ---

// FuzzParser feeds arbitrary strings to the parser and verifies no panics occur.
func FuzzParser(f *testing.F) {
	// Seed corpus with valid Ivy snippets and edge cases.
	seeds := []string{
		"",
		"type t",
		"type color = {red, green, blue}",
		"relation r(X:t)",
		"relation link(X:node, Y:node)",
		"axiom forall X:t. r(X, X)",
		"axiom [refl] forall X:t. r(X, X)",
		"individual x : node",
		"action send(src:node, dst:node) = {\n\tlink(src, dst) := true\n}",
		"export send",
		"import recv",
		"property [safe] forall X:node. ~link(X, X)",
		"conjecture forall X:node. ~link(X, X)",
		"definition succ(X:nat) = X + 1",
		"interpret t -> int",
		"interpret idx -> {0..15}",
		"type point = struct { x:nat, y:nat }",
		"module counter(t) = {\n\tindividual val : t\n}",
		"object counter = {\n\tindividual val : nat\n}",
		"instantiate counter(nat)",
		"variant msg of packet",
		"attribute server.weight = 10",
		"init {\n\tx := 0\n}",
		"before send {\n\tassert link(src, dst)\n}",
		"after recv {\n\tlink(src, dst) := true\n}",
		"implement send {\n\tlink(src, dst) := true\n}",
		"isolate server = {\n\ttype request\n}",
		"action test = {\n\tif x = 0 {\n\t\ty := 1\n\t} else {\n\t\ty := 2\n\t}\n}",
		"action test = {\n\twhile x > 0 {\n\t\tx := x - 1\n\t}\n}",
		"action get(x:t) returns (y:t) = {\n\ty := x\n}",
		// Edge cases
		"type = bad",
		"!!!@@@###",
		"((((()))))",
		"{}",
		"{,,,}",
		"forall",
		"exists",
		"~ ~ ~ ~",
		"a & b & c | d -> e <-> f",
		"type\ntype\ntype",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on input %q: %v", input, r)
			}
		}()
		p := New(input, lexer.Version{1, 7})
		_, _ = p.Parse()
	})
}

// FuzzParserExpr feeds arbitrary expression-like strings to parseExpr
// and verifies no panics occur.
func FuzzParserExpr(f *testing.F) {
	// Seed corpus with valid expressions and edge cases.
	seeds := []string{
		"",
		"a",
		"X",
		"true",
		"false",
		"this",
		"f(x, y)",
		"a & b | c",
		"a & b & c & d",
		"~a -> b <-> c",
		"forall X:t. f(X)",
		"exists X:t, Y:t. f(X) = f(Y)",
		"x + y * z",
		"x - y / z",
		"x <= y",
		"x < y",
		"x >= y",
		"x > y",
		"x = y",
		"x ~= y",
		"~p",
		"~~p",
		"p -> q -> r",
		"p <-> q",
		"(p & q) | r",
		"x if c else y",
		"a.b",
		"a.b.c",
		"old x",
		"globally p",
		"some X:t. p(X)",
		"X:t",
		"-x",
		"(a, b, c)",
		// Edge cases
		"(",
		")",
		"()",
		"&",
		"|",
		"->",
		"<->",
		"~",
		".",
		"::",
		"+ + +",
		"a b c",
		"f(,)",
		"f(x,,y)",
		"forall .",
		"exists .",
		"(((a)))",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on input %q: %v", input, r)
			}
		}()
		p := New(input, lexer.Version{1, 7})
		_ = p.parseExpr(0)
	})
}

// FuzzParserRoundTrip parses valid Ivy code snippets, calls String() on the
// resulting AST, re-parses the String() output, and verifies no crash on the
// re-parse. Exact structural equivalence is NOT required.
func FuzzParserRoundTrip(f *testing.F) {
	seeds := []string{
		"type t",
		"relation r(X:t)",
		"axiom forall X:t. r(X)",
		"individual x : t",
		"type color = {red, green, blue}",
		"relation link(X:node, Y:node)",
		"axiom [refl] forall X:t. r(X, X)",
		"property [safe] forall X:node. ~link(X, X)",
		"conjecture forall X:node. ~link(X, X)",
		"export send",
		"import recv",
		"interpret t -> int",
		"variant msg of packet",
		"definition succ(X:nat) = X + 1",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on input %q: %v", input, r)
			}
		}()

		// First parse
		p1 := New(input, lexer.Version{1, 7})
		decls, err := p1.Parse()
		if err != nil || len(decls) == 0 {
			return // skip invalid inputs
		}

		// Convert each decl to string and try re-parsing
		for _, decl := range decls {
			s := decl.String()
			if s == "" {
				continue
			}
			p2 := New(s, lexer.Version{1, 7})
			_, _ = p2.Parse() // errors are OK, panics are not
		}
	})
}
