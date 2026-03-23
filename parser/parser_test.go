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
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return result.Decls
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
	// Should parse as x + (y * z) due to precedence.
	// Arithmetic operators produce App (term-level), not Atom (formula-level),
	// matching Python's App for +, -, *, /.
	a, ok := e.(*ast.App)
	if !ok {
		t.Fatalf("expected App (+), got %T", e)
	}
	if a.Relname() != "+" {
		t.Errorf("got rep %q", a.Relname())
	}
	if _, ok := a.Terms[1].(*ast.App); !ok {
		t.Errorf("rhs should be App (*), got %T", a.Terms[1])
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
	// Python flattens a.b into Atom("a.b") via compose_atoms, not Dot(a, b).
	// ivy_logic_parser.py line 151-154: compose_atoms(p[1], p[3])
	e := parseExpr(t, "a.b")
	a, ok := e.(*ast.Atom)
	if !ok {
		t.Fatalf("expected Atom (flattened dot), got %T", e)
	}
	if a.Rep != "a.b" {
		t.Fatalf("expected Atom(\"a.b\"), got Atom(%q)", a.Rep)
	}
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
	// Python inlines object contents: ObjectDecl + prefixed inner declarations
	decls := parse(t, `object counter = {
		individual val : nat
	}`)
	if len(decls) != 2 {
		t.Fatalf("got %d decls, want 2 (ObjectDecl + ConstantDecl)", len(decls))
	}
	_, ok := decls[0].(*ast.ObjectDecl)
	if !ok {
		t.Fatalf("decls[0]: expected ObjectDecl, got %T", decls[0])
	}
	cd, ok := decls[1].(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("decls[1]: expected ConstantDecl, got %T", decls[1])
	}
	// The inner individual should be prefixed with "counter."
	if len(cd.Args()) > 0 {
		if a, ok := cd.Args()[0].(*ast.Atom); ok {
			if a.Rep != "counter.val" {
				t.Errorf("expected name counter.val, got %s", a.Rep)
			}
		}
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
	// Python produces TypeDecl + VariantDecl for "variant msg of packet"
	decls := parse(t, "variant msg of packet")
	if len(decls) != 2 {
		t.Fatalf("got %d decls, want 2 (TypeDecl + VariantDecl)", len(decls))
	}
	if _, ok := decls[0].(*ast.TypeDecl); !ok {
		t.Fatalf("decl[0]: expected TypeDecl, got %T", decls[0])
	}
	if _, ok := decls[1].(*ast.VariantDecl); !ok {
		t.Fatalf("decl[1]: expected VariantDecl, got %T", decls[1])
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
	// Python inlines isolate contents:
	//   [0] ObjectDecl server
	//   [1] TypeDecl server.request
	//   [2] ActionDecl server.handle
	//   [3] IsolateObjectDecl
	input := `
isolate server = {
    type request
    action handle(r:request) = {
        assume true
    }
}
`
	decls := parse(t, input)
	if len(decls) != 4 {
		t.Fatalf("got %d decls, want 4 (ObjectDecl + TypeDecl + ActionDecl + IsolateObjectDecl)", len(decls))
	}
	if _, ok := decls[0].(*ast.ObjectDecl); !ok {
		t.Fatalf("decls[0]: expected ObjectDecl, got %T", decls[0])
	}
	if _, ok := decls[1].(*ast.TypeDecl); !ok {
		t.Fatalf("decls[1]: expected TypeDecl, got %T", decls[1])
	}
	if _, ok := decls[2].(*ast.ActionDecl); !ok {
		t.Fatalf("decls[2]: expected ActionDecl, got %T", decls[2])
	}
	if _, ok := decls[3].(*ast.IsolateObjectDecl); !ok {
		t.Fatalf("decls[3]: expected IsolateObjectDecl, got %T", decls[3])
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
	// Python produces both ActionDecl and MixinDecl for "implement name { ... }"
	input := `implement send {
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
		result, err := p1.Parse()
		if err != nil || len(result.Decls) == 0 {
			return // skip invalid inputs
		}

		// Convert each decl to string and try re-parsing
		for _, decl := range result.Decls {
			s := decl.String()
			if s == "" {
				continue
			}
			p2 := New(s, lexer.Version{1, 7})
			_, _ = p2.Parse() // errors are OK, panics are not
		}
	})
}

// --- Scenario parsing tests ---

var v16 = lexer.Version{1, 6}

// TestScenarioParseScen1 parses the scenario from scen1.ivy and checks AST structure.
// Scenario has 2 "before a" transitions: s0->s1 and s1->s0.
func TestScenarioParseScen1(t *testing.T) {
	input := `#lang ivy1.6

action a = {
}

var q : bool

after init {
    q := false;
}

export a

scenario {
    -> s0;
    s0 -> s1 : before a {
	q := true
    }
    s1 -> s0 : before a {
	q := false
    }
}
`
	p := New(input, v16)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Find the ScenarioDecl
	var scenDecl *ast.ScenarioDecl
	for _, d := range result.Decls {
		if sd, ok := d.(*ast.ScenarioDecl); ok {
			scenDecl = sd
			break
		}
	}
	if scenDecl == nil {
		t.Fatal("expected ScenarioDecl in parsed output")
	}

	// The ScenarioDecl should wrap a ScenarioDef
	if len(scenDecl.DeclArgs) != 1 {
		t.Fatalf("expected 1 DeclArg, got %d", len(scenDecl.DeclArgs))
	}
	sdef, ok := scenDecl.DeclArgs[0].(*ast.ScenarioDef)
	if !ok {
		t.Fatalf("expected ScenarioDef, got %T", scenDecl.DeclArgs[0])
	}

	// Check init places
	initP := sdef.InitPlaces()
	if initP == nil {
		t.Fatal("expected non-nil InitPlaces")
	}
	if len(initP.Elems) != 1 {
		t.Fatalf("expected 1 init place, got %d", len(initP.Elems))
	}
	if atom, ok := initP.Elems[0].(*ast.Atom); !ok || atom.Rep != "s0" {
		t.Errorf("expected init place 's0', got %v", initP.Elems[0])
	}

	// Check transitions
	transs := sdef.Transitions()
	if len(transs) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(transs))
	}

	// Transition 0: s0 -> s1 : before a
	tr0 := transs[0]
	from0, _ := tr0.From.(*ast.PlaceList)
	to0, _ := tr0.To.(*ast.PlaceList)
	if from0 == nil || len(from0.Elems) != 1 || from0.Elems[0].(*ast.Atom).Rep != "s0" {
		t.Errorf("tr0 from: expected [s0], got %v", tr0.From)
	}
	if to0 == nil || len(to0.Elems) != 1 || to0.Elems[0].(*ast.Atom).Rep != "s1" {
		t.Errorf("tr0 to: expected [s1], got %v", tr0.To)
	}
	bm0, ok := tr0.Action.(*ast.ScenarioBeforeMixin)
	if !ok {
		t.Fatalf("tr0 action: expected ScenarioBeforeMixin, got %T", tr0.Action)
	}
	// Check mixer name — v1.6: "a[before]"
	mixer0, _ := bm0.Mixer.(*ast.Atom)
	if mixer0 == nil || mixer0.Rep != "a[before]" {
		t.Errorf("tr0 mixer: expected 'a[before]', got %v", bm0.Mixer)
	}
	// Check ActionDef name
	adef0, ok := bm0.Def.(*ast.ActionDef)
	if !ok {
		t.Fatalf("tr0 def: expected ActionDef, got %T", bm0.Def)
	}
	if adef0.Defines() != "a" {
		t.Errorf("tr0 def name: expected 'a', got %q", adef0.Defines())
	}

	// Transition 1: s1 -> s0 : before a
	tr1 := transs[1]
	bm1, ok := tr1.Action.(*ast.ScenarioBeforeMixin)
	if !ok {
		t.Fatalf("tr1 action: expected ScenarioBeforeMixin, got %T", tr1.Action)
	}
	// v1.6: second transition also gets "a[before]" (no counter)
	mixer1, _ := bm1.Mixer.(*ast.Atom)
	if mixer1 == nil || mixer1.Rep != "a[before]" {
		t.Errorf("tr1 mixer: expected 'a[before]', got %v", bm1.Mixer)
	}

	// Check Places() helper
	places := sdef.Places()
	if len(places) != 2 {
		t.Errorf("expected 2 unique places, got %d", len(places))
	}
	placeNames := map[string]bool{}
	for _, p := range places {
		placeNames[p.Name] = true
	}
	if !placeNames["s0"] || !placeNames["s1"] {
		t.Errorf("expected places {s0, s1}, got %v", placeNames)
	}
}

// TestScenarioParseScen2 parses scen2.ivy — transitions on different actions a and b.
func TestScenarioParseScen2(t *testing.T) {
	input := `#lang ivy1.6

action a = {}
action b = {}
var q : bool

after init {
    q := false;
}

export a
export b

scenario {
    -> s0;
    s0 -> s1 : before a {
        q := true
    }
    s1 -> s0 : before b {
        q := false
    }
}
`
	p := New(input, v16)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var scenDecl *ast.ScenarioDecl
	for _, d := range result.Decls {
		if sd, ok := d.(*ast.ScenarioDecl); ok {
			scenDecl = sd
			break
		}
	}
	if scenDecl == nil {
		t.Fatal("expected ScenarioDecl")
	}

	sdef := scenDecl.DeclArgs[0].(*ast.ScenarioDef)
	transs := sdef.Transitions()
	if len(transs) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(transs))
	}

	// tr0: before a
	bm0 := transs[0].Action.(*ast.ScenarioBeforeMixin)
	adef0 := bm0.Def.(*ast.ActionDef)
	if adef0.Defines() != "a" {
		t.Errorf("tr0: expected action 'a', got %q", adef0.Defines())
	}

	// tr1: before b
	bm1 := transs[1].Action.(*ast.ScenarioBeforeMixin)
	adef1 := bm1.Def.(*ast.ActionDef)
	if adef1.Defines() != "b" {
		t.Errorf("tr1: expected action 'b', got %q", adef1.Defines())
	}
}

// TestScenarioParseScen3 parses scen3.ivy — has both before and after mixins.
func TestScenarioParseScen3(t *testing.T) {
	input := `#lang ivy1.6

type foo = struct {
    y : bool
}

action a(x:foo) = {
    call b(x)
}

action b(z:foo) = {}
var q : bool

after init {
    q := false;
}

export a
import b

object scen = {
    scenario {
        -> s0;
        s0 -> s1 : before a {
            q := x.y
        }
        s1 -> s0 : after b {
            q := z.y
        }
    }
}
`
	p := New(input, v16)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	decls := result.Decls

	// The scenario is inside object scen, so find it in nested decls
	var scenDecl *ast.ScenarioDecl
	var findScenario func(nodes []ast.Node)
	findScenario = func(nodes []ast.Node) {
		for _, d := range nodes {
			if sd, ok := d.(*ast.ScenarioDecl); ok {
				scenDecl = sd
				return
			}
			// Check DeclArgs for nested decls
			if hasArgs, ok := d.(interface{ GetDeclArgs() []ast.Node }); ok {
				findScenario(hasArgs.GetDeclArgs())
			}
			// Also recurse into children
			for _, child := range d.Args() {
				if sd, ok := child.(*ast.ScenarioDecl); ok {
					scenDecl = sd
					return
				}
			}
		}
	}
	findScenario(decls)
	if scenDecl == nil {
		t.Fatal("expected ScenarioDecl somewhere in parsed output")
	}

	sdef := scenDecl.DeclArgs[0].(*ast.ScenarioDef)
	transs := sdef.Transitions()
	if len(transs) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(transs))
	}

	// tr0: before a
	_, isBefore := transs[0].Action.(*ast.ScenarioBeforeMixin)
	if !isBefore {
		t.Errorf("tr0: expected ScenarioBeforeMixin, got %T", transs[0].Action)
	}

	// tr1: after b
	am, isAfter := transs[1].Action.(*ast.ScenarioAfterMixin)
	if !isAfter {
		t.Fatalf("tr1: expected ScenarioAfterMixin, got %T", transs[1].Action)
	}
	adef := am.Def.(*ast.ActionDef)
	if adef.Defines() != "b" {
		t.Errorf("tr1: expected action 'b', got %q", adef.Defines())
	}
}

// Regression: dotted name expressions must be flattened into a single Atom
// with a composed name, matching Python's compose_atoms behavior.
// Bug was: Go created Dot(Dot(Symbol("ref"), Atom("evs",[T])), Symbol("req"))
// instead of Atom("ref.evs.req", [T]).
func TestDotFlattening_QualifiedNameWithArgs(t *testing.T) {
	// ref.evs(T).req should become Atom("ref.evs.req", [T])
	e := parseExpr(t, "ref.evs(T).req")
	a, ok := e.(*ast.Atom)
	if !ok {
		t.Fatalf("expected Atom (flattened), got %T: %v", e, e)
	}
	if a.Rep != "ref.evs.req" {
		t.Errorf("expected name \"ref.evs.req\", got %q", a.Rep)
	}
	if len(a.Terms) != 1 {
		t.Errorf("expected 1 term (T), got %d terms", len(a.Terms))
	}
}

// Regression: dotted names without args should also flatten.
func TestDotFlattening_SimpleQualifiedName(t *testing.T) {
	e := parseExpr(t, "a.b.c")
	a, ok := e.(*ast.Atom)
	if !ok {
		t.Fatalf("expected Atom (flattened), got %T: %v", e, e)
	}
	if a.Rep != "a.b.c" {
		t.Errorf("expected name \"a.b.c\", got %q", a.Rep)
	}
}

// Regression: old(x.y) should flatten the inner dotted name.
func TestDotFlattening_OldDot(t *testing.T) {
	e := parseExpr(t, "old x.y")
	old, ok := e.(*ast.Old)
	if !ok {
		t.Fatalf("expected Old, got %T: %v", e, e)
	}
	inner, ok := old.Term.(*ast.Atom)
	if !ok {
		t.Fatalf("expected Atom inside Old, got %T: %v", old.Term, old.Term)
	}
	if inner.Rep != "x.y" {
		t.Errorf("expected inner name \"x.y\", got %q", inner.Rep)
	}
}

// Regression: parameterized instance must propagate prefix parameters
// to inner declarations. E.g., "instance abs(P:proc) : mymod(nat)"
// expanding "var begun(X:n) : bool" must produce "abs.begun(P:proc, X:nat)"
// with 2 terms, not "abs.begun(X:n)" with 1 term.
// Bug was two-fold:
//   1. pref was created as Atom("abs") without the P:proc parameter
//   2. all names were marked as static, stripping pref args in ComposeAtoms
func TestParameterizedInstanceExpansion(t *testing.T) {
	src := `
type nat
type proc
type bool
module mymod(n) = {
    var begun(X:n) : bool
}
instance abs(P:proc) : mymod(nat)
`
	v18 := lexer.Version{1, 8}
	p := New(src, v18)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Find the ConstantDecl for abs.begun
	var found *ast.Atom
	for _, d := range result.Decls {
		cd, ok := d.(*ast.ConstantDecl)
		if !ok {
			continue
		}
		for _, arg := range cd.DeclArgs {
			if a, ok := arg.(*ast.Atom); ok && a.Rep == "abs.begun" {
				found = a
			}
		}
	}
	if found == nil {
		t.Fatal("abs.begun not found in expanded declarations")
	}
	if len(found.Terms) != 2 {
		t.Errorf("abs.begun should have 2 terms (P and X), got %d: %v", len(found.Terms), found)
	}
	// First term should be the prefix parameter (P:proc)
	if len(found.Terms) >= 1 {
		if v, ok := found.Terms[0].(*ast.Variable); ok {
			if v.Rep != "P" {
				t.Errorf("first term should be Variable P, got Variable %q", v.Rep)
			}
		} else {
			t.Errorf("first term should be *ast.Variable, got %T: %v", found.Terms[0], found.Terms[0])
		}
	}
}

// --- Argument absorption tests ---
// These tests verify that the parser correctly separates names from
// parameter lists in grammar productions. The bug that motivated these:
// parseCallatom() greedily absorbed "(nat)" in "module foo(nat) = ...",
// leaving FormalParams empty and breaking module instantiation.

func TestModuleParamsSingleParam(t *testing.T) {
	src := `module foo(t) = {
		var x : t
	}`
	p := New(src, v17)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// Find the ModuleDecl
	var md *ast.ModuleDecl
	for _, d := range result.Decls {
		if m, ok := d.(*ast.ModuleDecl); ok {
			md = m
			break
		}
	}
	if md == nil {
		t.Fatal("ModuleDecl not found")
	}
	if len(md.FormalParams) != 1 {
		t.Fatalf("FormalParams: got %d, want 1: %v", len(md.FormalParams), md.FormalParams)
	}
	if a, ok := md.FormalParams[0].(*ast.Atom); ok {
		if a.Rep != "t" {
			t.Errorf("FormalParams[0] name: got %q, want %q", a.Rep, "t")
		}
	} else {
		t.Errorf("FormalParams[0]: got %T, want *ast.Atom", md.FormalParams[0])
	}
	if len(md.BodyDecls) == 0 {
		t.Error("BodyDecls should not be empty")
	}
}

func TestModuleParamsMultipleParams(t *testing.T) {
	src := `module foo(a,b) = {
		var x : a
	}`
	p := New(src, v17)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	var md *ast.ModuleDecl
	for _, d := range result.Decls {
		if m, ok := d.(*ast.ModuleDecl); ok {
			md = m
			break
		}
	}
	if md == nil {
		t.Fatal("ModuleDecl not found")
	}
	if len(md.FormalParams) != 2 {
		t.Fatalf("FormalParams: got %d, want 2: %v", len(md.FormalParams), md.FormalParams)
	}
	names := make([]string, len(md.FormalParams))
	for i, fp := range md.FormalParams {
		if a, ok := fp.(*ast.Atom); ok {
			names[i] = a.Rep
		}
	}
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("FormalParams names: got %v, want [a b]", names)
	}
}

func TestModuleParamsNone(t *testing.T) {
	src := `module foo = {
		var x : nat
	}`
	p := New(src, v17)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	var md *ast.ModuleDecl
	for _, d := range result.Decls {
		if m, ok := d.(*ast.ModuleDecl); ok {
			md = m
			break
		}
	}
	if md == nil {
		t.Fatal("ModuleDecl not found")
	}
	if len(md.FormalParams) != 0 {
		t.Errorf("FormalParams: got %d, want 0: %v", len(md.FormalParams), md.FormalParams)
	}
}

func TestModuleIsolateParams(t *testing.T) {
	src := `module isolate foo(t) with t = {
		var x : t
	}`
	p := New(src, v17)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	var md *ast.ModuleDecl
	for _, d := range result.Decls {
		if m, ok := d.(*ast.ModuleDecl); ok {
			md = m
			break
		}
	}
	if md == nil {
		t.Fatal("ModuleDecl not found")
	}
	if len(md.FormalParams) != 1 {
		t.Fatalf("FormalParams: got %d, want 1: %v", len(md.FormalParams), md.FormalParams)
	}
	if a, ok := md.FormalParams[0].(*ast.Atom); ok {
		if a.Rep != "t" {
			t.Errorf("FormalParams[0] name: got %q, want %q", a.Rep, "t")
		}
	} else {
		t.Errorf("FormalParams[0]: got %T, want *ast.Atom", md.FormalParams[0])
	}
	// Body should include IsolateDecl injected by modcat=="isolate"
	hasIso := false
	for _, bd := range md.BodyDecls {
		if _, ok := bd.(*ast.IsolateDecl); ok {
			hasIso = true
		}
	}
	if !hasIso {
		t.Error("module isolate body should contain an IsolateDecl")
	}
}

func TestModuleIsolateMultipleParams(t *testing.T) {
	src := `module isolate foo(a,b) with a,b = {
		var x : a
		var y : b
	}`
	p := New(src, v17)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	var md *ast.ModuleDecl
	for _, d := range result.Decls {
		if m, ok := d.(*ast.ModuleDecl); ok {
			md = m
			break
		}
	}
	if md == nil {
		t.Fatal("ModuleDecl not found")
	}
	if len(md.FormalParams) != 2 {
		t.Fatalf("FormalParams: got %d, want 2: %v", len(md.FormalParams), md.FormalParams)
	}
}

func TestInstanceExpansionSubstitutesSorts(t *testing.T) {
	src := `
module mymod(t) = {
	var x : t
}
instance bar : mymod(nat)
`
	p := New(src, v17)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// Find ConstantDecl for bar.x
	var found *ast.Atom
	for _, d := range result.Decls {
		cd, ok := d.(*ast.ConstantDecl)
		if !ok {
			continue
		}
		for _, arg := range cd.DeclArgs {
			if a, ok := arg.(*ast.Atom); ok && a.Rep == "bar.x" {
				found = a
			}
		}
	}
	if found == nil {
		t.Fatal("bar.x not found in expanded declarations")
	}
	sortStr := ""
	if found.ASort != nil {
		sortStr = strings.TrimSpace(found.ASort.String())
	}
	if sortStr != "nat" {
		t.Errorf("bar.x sort: got %q, want %q", sortStr, "nat")
	}
}

func TestInstanceExpansionWithVariableParam(t *testing.T) {
	src := `
type nat
type proc
module mymod(t) = {
	var x : t
}
instance bar(P:proc) : mymod(nat)
`
	v18 := lexer.Version{1, 8}
	p := New(src, v18)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	var found *ast.Atom
	for _, d := range result.Decls {
		cd, ok := d.(*ast.ConstantDecl)
		if !ok {
			continue
		}
		for _, arg := range cd.DeclArgs {
			if a, ok := arg.(*ast.Atom); ok && a.Rep == "bar.x" {
				found = a
			}
		}
	}
	if found == nil {
		t.Fatal("bar.x not found in expanded declarations")
	}
	sortStr := ""
	if found.ASort != nil {
		sortStr = strings.TrimSpace(found.ASort.String())
	}
	if sortStr != "nat" {
		t.Errorf("bar.x sort: got %q, want %q", sortStr, "nat")
	}
}

func TestIsolateParamsNotAbsorbed(t *testing.T) {
	// isolate foo = this with bar
	// Name should be "foo", not "foo(something)"
	src := `isolate foo = this with bar`
	p := New(src, v17)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// Should produce IsolateDecl
	var isoDecl *ast.IsolateDecl
	for _, d := range result.Decls {
		if id, ok := d.(*ast.IsolateDecl); ok {
			isoDecl = id
			break
		}
	}
	if isoDecl == nil {
		t.Fatal("IsolateDecl not found")
	}
}

func TestObjectParamsCreatePrefixedDecls(t *testing.T) {
	src := `object foo = {
		var x : nat
	}`
	p := New(src, v17)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// Find ConstantDecl with prefixed name
	var found *ast.Atom
	for _, d := range result.Decls {
		cd, ok := d.(*ast.ConstantDecl)
		if !ok {
			continue
		}
		for _, arg := range cd.DeclArgs {
			if a, ok := arg.(*ast.Atom); ok && a.Rep == "foo.x" {
				found = a
			}
		}
	}
	if found == nil {
		t.Error("foo.x not found — object body should be prefixed with object name")
	}
}
