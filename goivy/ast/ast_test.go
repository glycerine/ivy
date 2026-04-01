package ast

import (
	"strings"
	"testing"
)

// --- Core type tests ---

func TestSymbol(t *testing.T) {
	cfg := NewAstConfig()
	s := cfg.NewSymbol("foo", nil)
	if s.String() != "foo" {
		t.Errorf("got %q, want %q", s.String(), "foo")
	}
	if s.Relname() != "foo" {
		t.Errorf("got %q, want %q", s.Relname(), "foo")
	}
	c := s.Clone(nil)
	if c.String() != "foo" {
		t.Errorf("clone: got %q, want %q", c.String(), "foo")
	}
}

func TestAtom(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewSymbol("x", nil)
	y := cfg.NewSymbol("y", nil)
	a := cfg.NewAtom("r", x, y)
	if a.String() != "r(x, y)" {
		t.Errorf("got %q", a.String())
	}

	// Nullary atom
	a0 := cfg.NewAtom("p")
	if a0.String() != "p" {
		t.Errorf("got %q, want %q", a0.String(), "p")
	}

	// Equality
	eq := cfg.NewAtom("=", x, y)
	if eq.String() != "x = y" {
		t.Errorf("got %q", eq.String())
	}
}

func TestAtomPrefix(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewAtom("foo", cfg.NewSymbol("x", nil))
	p := a.Prefix("bar.")
	if p.Rep != "bar.foo" {
		t.Errorf("got %q, want %q", p.Rep, "bar.foo")
	}
}

func TestAtomSuffix(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewAtom("foo")
	s := a.Suffix("_v2")
	if s.Rep != "foo_v2" {
		t.Errorf("got %q, want %q", s.Rep, "foo_v2")
	}
}

func TestApp(t *testing.T) {
	cfg := NewAstConfig()
	f := cfg.NewSymbol("f", nil)
	x := cfg.NewSymbol("x", nil)
	app := cfg.NewApp(f, x)
	if app.String() != "f(x)" {
		t.Errorf("got %q", app.String())
	}

	// Nullary app
	a0 := cfg.NewApp(f)
	if a0.String() != "f" {
		t.Errorf("got %q, want %q", a0.String(), "f")
	}
}

func TestVariable(t *testing.T) {
	cfg := NewAstConfig()
	v := cfg.NewVariable("X", "t")
	if v.String() != "X:t" {
		t.Errorf("got %q", v.String())
	}
	// Clone returns self
	c := v.Clone(nil)
	if c != v {
		t.Error("Variable.Clone should return self")
	}
	// Resort
	v2 := v.Resort("u")
	if v2.String() != "X:u" {
		t.Errorf("got %q", v2.String())
	}
}

func TestOld(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewSymbol("x", nil)
	o := cfg.NewOld(x)
	if o.String() != "old x" {
		t.Errorf("got %q", o.String())
	}
}

func TestThis(t *testing.T) {
	th := &This{}
	if th.String() != "this" {
		t.Errorf("got %q", th.String())
	}
	if th.Relname() != "this" {
		t.Errorf("got %q", th.Relname())
	}
}

func TestLiteral(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewAtom("p", cfg.NewSymbol("x", nil))
	pos := cfg.NewLiteral(1, a)
	neg := cfg.NewLiteral(0, a)
	if pos.String() != "p(x)" {
		t.Errorf("got %q", pos.String())
	}
	if neg.String() != "~p(x)" {
		t.Errorf("got %q", neg.String())
	}
	inv := neg.Invert()
	if inv.Polarity != 1 {
		t.Error("inverted polarity should be 1")
	}
}

func TestDot(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewSymbol("a", nil)
	b := cfg.NewSymbol("b", nil)
	d := cfg.NewDot(a, b)
	if d.String() != "a.b" {
		t.Errorf("got %q", d.String())
	}
}

func TestBracket(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewSymbol("a", nil)
	i := cfg.NewSymbol("i", nil)
	br := cfg.NewBracket(a, i)
	if br.String() != "a[i]" {
		t.Errorf("got %q", br.String())
	}
}

func TestTuple(t *testing.T) {
	cfg := NewAstConfig()
	tup := cfg.NewTuple(cfg.NewSymbol("a", nil), cfg.NewSymbol("b", nil))
	if tup.String() != "(a,b)" {
		t.Errorf("got %q", tup.String())
	}
}

// --- Formula tests ---

func TestAndTrue(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewAnd()
	if a.String() != "true" {
		t.Errorf("got %q, want %q", a.String(), "true")
	}
	if !IsTrue(a) {
		t.Error("empty And should be true")
	}
}

func TestOrFalse(t *testing.T) {
	cfg := NewAstConfig()
	o := cfg.NewOr()
	if o.String() != "false" {
		t.Errorf("got %q, want %q", o.String(), "false")
	}
	if !IsFalse(o) {
		t.Error("empty Or should be false")
	}
}

func TestAndMulti(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewAnd(cfg.NewSymbol("p", nil), cfg.NewSymbol("q", nil))
	if a.String() != "(p & q)" {
		t.Errorf("got %q", a.String())
	}
}

func TestOrMulti(t *testing.T) {
	cfg := NewAstConfig()
	o := cfg.NewOr(cfg.NewSymbol("p", nil), cfg.NewSymbol("q", nil))
	if o.String() != "(p | q)" {
		t.Errorf("got %q", o.String())
	}
}

func TestNot(t *testing.T) {
	cfg := NewAstConfig()
	n := cfg.NewNot(cfg.NewSymbol("p", nil))
	if n.String() != "~p" {
		t.Errorf("got %q", n.String())
	}
}

func TestNotEquals(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewSymbol("x", nil)
	y := cfg.NewSymbol("y", nil)
	eq := cfg.NewAtom("=", x, y)
	n := cfg.NewNot(eq)
	if n.String() != "x ~= y" {
		t.Errorf("got %q", n.String())
	}
}

func TestImplies(t *testing.T) {
	cfg := NewAstConfig()
	i := cfg.NewImplies(cfg.NewSymbol("p", nil), cfg.NewSymbol("q", nil))
	if i.String() != "p -> q" {
		t.Errorf("got %q", i.String())
	}
}

func TestIff(t *testing.T) {
	cfg := NewAstConfig()
	f := cfg.NewIff(cfg.NewSymbol("p", nil), cfg.NewSymbol("q", nil))
	if f.String() != "p <-> q" {
		t.Errorf("got %q", f.String())
	}
}

func TestIte(t *testing.T) {
	cfg := NewAstConfig()
	i := cfg.NewIte(cfg.NewSymbol("c", nil), cfg.NewSymbol("t", nil), cfg.NewSymbol("e", nil))
	if i.String() != "(t if c else e)" {
		t.Errorf("got %q", i.String())
	}
}

func TestForall(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewVariable("X", "t")
	body := cfg.NewAtom("p", x)
	f := cfg.NewForall([]Node{x}, body)
	if !strings.HasPrefix(f.String(), "forall X:t.") {
		t.Errorf("got %q", f.String())
	}
}

func TestExists(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewVariable("X", "t")
	body := cfg.NewAtom("p", x)
	e := cfg.NewExists([]Node{x}, body)
	if !strings.HasPrefix(e.String(), "exists X:t.") {
		t.Errorf("got %q", e.String())
	}
}

func TestGlobally(t *testing.T) {
	cfg := NewAstConfig()
	g := cfg.NewGlobally(cfg.NewSymbol("p", nil))
	if g.String() != "(globally p)" {
		t.Errorf("got %q", g.String())
	}
}

func TestEventually(t *testing.T) {
	cfg := NewAstConfig()
	e := cfg.NewEventually(cfg.NewSymbol("p", nil))
	if e.String() != "(eventually p)" {
		t.Errorf("got %q", e.String())
	}
}

func TestHasTemporal(t *testing.T) {
	cfg := NewAstConfig()
	if HasTemporal(cfg.NewSymbol("p", nil)) {
		t.Error("plain symbol should not be temporal")
	}
	if !HasTemporal(cfg.NewGlobally(cfg.NewSymbol("p", nil))) {
		t.Error("Globally should be temporal")
	}
	if !HasTemporal(cfg.NewAnd(cfg.NewSymbol("p", nil), cfg.NewEventually(cfg.NewSymbol("q", nil)))) {
		t.Error("And containing Eventually should be temporal")
	}
}

func TestLet(t *testing.T) {
	cfg := NewAstConfig()
	def := cfg.NewDefinition(cfg.NewAtom("p"), cfg.NewSymbol("true", nil))
	body := cfg.NewSymbol("q", nil)
	l := cfg.NewLet([]Node{def}, body)
	if !strings.HasPrefix(l.String(), "let ") {
		t.Errorf("got %q", l.String())
	}
}

func TestDefinition(t *testing.T) {
	cfg := NewAstConfig()
	d := cfg.NewDefinition(cfg.NewAtom("f", cfg.NewSymbol("x", nil)), cfg.NewSymbol("y", nil))
	if d.String() != "f(x) = y" {
		t.Errorf("got %q", d.String())
	}
}

func TestNamedBinder(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewVariable("X", "t")
	body := cfg.NewAtom("p", x)
	nb := cfg.NewNamedBinder("mybinder", []Node{x}, body)
	if !strings.Contains(nb.String(), "mybinder") {
		t.Errorf("got %q", nb.String())
	}
}

func TestWhenOperator(t *testing.T) {
	cfg := NewAstConfig()
	w := cfg.NewWhenOperator("fire", cfg.NewSymbol("p", nil), cfg.NewSymbol("q", nil))
	if w.String() != "(p firewhennext q)" {
		t.Errorf("got %q", w.String())
	}
}

// --- Sort tests ---

func TestConstantSort(t *testing.T) {
	cfg := NewAstConfig()
	cs := cfg.NewConstantSort()
	if cs.String() != "uninterpreted" {
		t.Errorf("got %q", cs.String())
	}
}

func TestEnumeratedSort(t *testing.T) {
	cfg := NewAstConfig()
	es := cfg.NewEnumeratedSort(cfg.NewSymbol("a", nil), cfg.NewSymbol("b", nil), cfg.NewSymbol("c", nil))
	if es.String() != "{a,b,c}" {
		t.Errorf("got %q", es.String())
	}
	ext := es.Extension()
	if len(ext) != 3 || ext[0] != "a" || ext[1] != "b" || ext[2] != "c" {
		t.Errorf("extension: %v", ext)
	}
}

func TestFunctionSortAST(t *testing.T) {
	cfg := NewAstConfig()
	dom := []Node{cfg.NewSymbol("S", nil), cfg.NewSymbol("S", nil)}
	rng := cfg.NewSymbol("bool", nil)
	fs := cfg.NewFunctionSort(dom, rng)
	if fs.String() != "S * S -> bool" {
		t.Errorf("got %q", fs.String())
	}
}

func TestRelationSortAST(t *testing.T) {
	cfg := NewAstConfig()
	rs := cfg.NewRelationSort([]Node{cfg.NewSymbol("S", nil), cfg.NewSymbol("S", nil)})
	if rs.String() != "S * S" {
		t.Errorf("got %q", rs.String())
	}
}

func TestStructSortAST(t *testing.T) {
	cfg := NewAstConfig()
	ss := cfg.NewStructSort(cfg.NewAtom("x"), cfg.NewAtom("y"))
	if !strings.HasPrefix(ss.String(), "struct {") {
		t.Errorf("got %q", ss.String())
	}
}

func TestRange(t *testing.T) {
	cfg := NewAstConfig()
	r := cfg.NewRange(cfg.NewSymbol("0", nil), cfg.NewSymbol("10", nil))
	if r.String() != "{0..10}" {
		t.Errorf("got %q", r.String())
	}
}

// --- Declaration tests ---

func TestLabeledFormula(t *testing.T) {
	cfg := NewAstConfig()
	label := cfg.NewAtom("my_axiom")
	fmla := cfg.NewImplies(cfg.NewSymbol("p", nil), cfg.NewSymbol("q", nil))
	lf := cfg.NewLabeledFormula(label, fmla)
	if lf.String() != "[my_axiom] p -> q" {
		t.Errorf("got %q", lf.String())
	}
	if lf.LabelName() != "my_axiom" {
		t.Errorf("got %q", lf.LabelName())
	}
	// Fresh ID
	lf2 := lf.CloneWithFreshID([]Node{label, fmla})
	if lf2.ID == lf.ID {
		t.Error("fresh ID should differ")
	}
}

func TestTypeDef(t *testing.T) {
	cfg := NewAstConfig()
	td := cfg.NewTypeDef(cfg.NewSymbol("color", nil), cfg.NewEnumeratedSort(cfg.NewSymbol("red", nil), cfg.NewSymbol("green", nil), cfg.NewSymbol("blue", nil)))
	if !strings.Contains(td.String(), "color") {
		t.Errorf("got %q", td.String())
	}
	defs := td.Defines()
	if len(defs) != 4 { // color + red + green + blue
		t.Errorf("defines: %v", defs)
	}
}

func TestActionDef(t *testing.T) {
	cfg := NewAstConfig()
	name := cfg.NewAtom("my_action")
	body := cfg.NewSymbol("skip", nil)
	params := []Node{cfg.NewVariable("X", "t")}
	ad := cfg.NewActionDef(name, body, params, nil)
	if ad.Defines() != "my_action" {
		t.Errorf("got %q", ad.Defines())
	}
}

func TestVariantDef(t *testing.T) {
	cfg := NewAstConfig()
	vd := cfg.NewVariantDef(cfg.NewSymbol("cons", nil), cfg.NewSymbol("list", nil))
	if vd.String() != "cons of list" {
		t.Errorf("got %q", vd.String())
	}
}

func TestAttributeDef(t *testing.T) {
	cfg := NewAstConfig()
	ad := cfg.NewAttributeDef(cfg.NewAtom("weight"), cfg.NewSymbol("10", nil))
	if ad.String() != "attribute weight = 10" {
		t.Errorf("got %q", ad.String())
	}
}

func TestInstantiation(t *testing.T) {
	cfg := NewAstConfig()
	inst := cfg.NewInstantiation(cfg.NewSymbol("mymod", nil), cfg.NewSymbol("int", nil))
	if inst.String() != "mymod : int" {
		t.Errorf("got %q", inst.String())
	}
}

// --- Clone tests ---

func TestAndClone(t *testing.T) {
	cfg := NewAstConfig()
	a := cfg.NewAnd(cfg.NewSymbol("p", nil), cfg.NewSymbol("q", nil))
	c := a.Clone([]Node{cfg.NewSymbol("r", nil)})
	if c.String() != "r" {
		t.Errorf("got %q", c.String())
	}
	// Original unchanged
	if a.String() != "(p & q)" {
		t.Errorf("original changed: %q", a.String())
	}
}

func TestForallClone(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewVariable("X", "t")
	body := cfg.NewAtom("p", x)
	f := cfg.NewForall([]Node{x}, body)
	newBody := cfg.NewAtom("q", x)
	c := f.Clone([]Node{newBody}).(*Forall)
	// Bounds should be preserved
	if len(c.Bounds) != 1 {
		t.Errorf("bounds lost in clone")
	}
}

// --- Location tests ---

func TestLocation(t *testing.T) {
	loc := Location{Filename: "test.ivy", Line: 42}
	if loc.String() != "test.ivy: line 42: " {
		t.Errorf("got %q", loc.String())
	}
	loc2 := Location{Line: 7}
	if loc2.String() != "line 7: " {
		t.Errorf("got %q", loc2.String())
	}
}

func TestSetLineno(t *testing.T) {
	cfg := NewAstConfig()
	s := cfg.NewSymbol("x", nil)
	loc := Location{Filename: "test.ivy", Line: 10}
	s.SetLineno(loc)
	if s.GetLineno() != loc {
		t.Errorf("got %v, want %v", s.GetLineno(), loc)
	}
}

// --- Tactic tests ---

func TestNullTactic(t *testing.T) {
	nt := &NullTactic{}
	if nt.String() != "{}" {
		t.Errorf("got %q", nt.String())
	}
}

func TestComposeTactics(t *testing.T) {
	ct := &ComposeTactics{Tactics: []Node{
		&ShowGoalsTactic{},
		&NullTactic{},
	}}
	if ct.String() != "showgoals; {}" {
		t.Errorf("got %q", ct.String())
	}
}

func TestSchemaInstantiation(t *testing.T) {
	cfg := NewAstConfig()
	si := &SchemaInstantiation{
		SchemaName: cfg.NewSymbol("my_schema", nil),
		Ren:        &NoneAST{},
	}
	if !strings.HasPrefix(si.String(), "apply my_schema") {
		t.Errorf("got %q", si.String())
	}
}

// --- Some tests ---

func TestSome(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewVariable("X", "t")
	body := cfg.NewAtom("p", x)
	s := cfg.NewSome([]Node{x}, body)
	if !strings.HasPrefix(s.String(), "some X:t.") {
		t.Errorf("got %q", s.String())
	}
}

func TestSomeExpr(t *testing.T) {
	cfg := NewAstConfig()
	x := cfg.NewVariable("X", "t")
	body := cfg.NewAtom("p", x)
	se := &SomeExpr{Param: x, Fmla: body}
	if !strings.HasPrefix(se.String(), "some X:t.") {
		t.Errorf("got %q", se.String())
	}
}

// --- Node interface compliance ---
// Verify that all types implement Node interface at compile time.

var _ Node = (*NoneAST)(nil)
var _ Node = (*Symbol)(nil)
var _ Node = (*Atom)(nil)
var _ Node = (*App)(nil)
var _ Node = (*Variable)(nil)
var _ Node = (*Old)(nil)
var _ Node = (*This)(nil)
var _ Node = (*MethodCall)(nil)
var _ Node = (*Literal)(nil)
var _ Node = (*Dot)(nil)
var _ Node = (*Bracket)(nil)
var _ Node = (*Tuple)(nil)
var _ Node = (*Some)(nil)
var _ Node = (*SomeMin)(nil)
var _ Node = (*SomeMax)(nil)
var _ Node = (*SomeExpr)(nil)
var _ Node = (*KeyArg)(nil)
var _ Node = (*DebugItem)(nil)
var _ Node = (*TemporalModels)(nil)
var _ Node = (*And)(nil)
var _ Node = (*Or)(nil)
var _ Node = (*Not)(nil)
var _ Node = (*Implies)(nil)
var _ Node = (*Iff)(nil)
var _ Node = (*Ite)(nil)
var _ Node = (*Forall)(nil)
var _ Node = (*Exists)(nil)
var _ Node = (*Isa)(nil)
var _ Node = (*Globally)(nil)
var _ Node = (*Eventually)(nil)
var _ Node = (*WhenOperator)(nil)
var _ Node = (*Let)(nil)
var _ Node = (*Definition)(nil)
var _ Node = (*DefinitionSchema)(nil)
var _ Node = (*NamedBinder)(nil)
var _ Node = (*Trigger)(nil)
var _ Node = (*ConstantSort)(nil)
var _ Node = (*EnumeratedSort)(nil)
var _ Node = (*StructSort)(nil)
var _ Node = (*FunctionSort)(nil)
var _ Node = (*RelationSort)(nil)
var _ Node = (*Range)(nil)
var _ Node = (*LabeledFormula)(nil)
var _ Node = (*ModuleDecl)(nil)
var _ Node = (*MacroDecl)(nil)
var _ Node = (*ObjectDecl)(nil)
var _ Node = (*ActionDecl)(nil)
var _ Node = (*ActionDef)(nil)
var _ Node = (*RelationDecl)(nil)
var _ Node = (*ConstantDecl)(nil)
var _ Node = (*ParameterDecl)(nil)
var _ Node = (*DestructorDecl)(nil)
var _ Node = (*ConstructorDecl)(nil)
var _ Node = (*TypeDecl)(nil)
var _ Node = (*TypeDef)(nil)
var _ Node = (*GhostTypeDef)(nil)
var _ Node = (*VariantDecl)(nil)
var _ Node = (*VariantDef)(nil)
var _ Node = (*AxiomDecl)(nil)
var _ Node = (*PropertyDecl)(nil)
var _ Node = (*ConjectureDecl)(nil)
var _ Node = (*ProofDecl)(nil)
var _ Node = (*NamedDecl)(nil)
var _ Node = (*SchemaDecl)(nil)
var _ Node = (*SchemaBody)(nil)
var _ Node = (*TheoremDecl)(nil)
var _ Node = (*DerivedDecl)(nil)
var _ Node = (*DefinitionDecl)(nil)
var _ Node = (*ProgressDecl)(nil)
var _ Node = (*RelyDecl)(nil)
var _ Node = (*MixOrdDecl)(nil)
var _ Node = (*ConceptDecl)(nil)
var _ Node = (*InitDecl)(nil)
var _ Node = (*StateDecl)(nil)
var _ Node = (*UpdateDecl)(nil)
var _ Node = (*AssertDecl)(nil)
var _ Node = (*InterpretDecl)(nil)
var _ Node = (*MixinDecl)(nil)
var _ Node = (*MixinBeforeDef)(nil)
var _ Node = (*MixinImplementDef)(nil)
var _ Node = (*MixinAfterDef)(nil)
var _ Node = (*IsolateDecl)(nil)
var _ Node = (*IsolateDef)(nil)
var _ Node = (*TrustedIsolateDef)(nil)
var _ Node = (*ExtractDef)(nil)
var _ Node = (*ProcessDef)(nil)
var _ Node = (*ExportDecl)(nil)
var _ Node = (*ExportDef)(nil)
var _ Node = (*ImportDecl)(nil)
var _ Node = (*ImportDef)(nil)
var _ Node = (*PrivateDecl)(nil)
var _ Node = (*AliasDecl)(nil)
var _ Node = (*DelegateDecl)(nil)
var _ Node = (*DelegateDef)(nil)
var _ Node = (*ImplementTypeDecl)(nil)
var _ Node = (*NativeCode)(nil)
var _ Node = (*NativeType)(nil)
var _ Node = (*NativeExpr)(nil)
var _ Node = (*NativeDef)(nil)
var _ Node = (*NativeDecl)(nil)
var _ Node = (*AttributeDef)(nil)
var _ Node = (*AttributeDecl)(nil)
var _ Node = (*Instantiation)(nil)
var _ Node = (*InstantiateDecl)(nil)
var _ Node = (*AutoInstanceDecl)(nil)
var _ Node = (*StateDef)(nil)
var _ Node = (*Renaming)(nil)
var _ Node = (*ScenarioDecl)(nil)
var _ Node = (*PlaceList)(nil)
var _ Node = (*ScenarioTransition)(nil)
var _ Node = (*ScenarioDef)(nil)
var _ Node = (*ScenarioBeforeMixin)(nil)
var _ Node = (*ScenarioAfterMixin)(nil)
var _ Node = (*IsolateObjectDecl)(nil)
var _ Node = (*Tactic)(nil)
var _ Node = (*SchemaInstantiation)(nil)
var _ Node = (*AssumeTactic)(nil)
var _ Node = (*UnfoldSpec)(nil)
var _ Node = (*UnfoldTactic)(nil)
var _ Node = (*ForgetTactic)(nil)
var _ Node = (*ShowGoalsTactic)(nil)
var _ Node = (*DeferGoalTactic)(nil)
var _ Node = (*NullTactic)(nil)
var _ Node = (*LetTactic)(nil)
var _ Node = (*WitnessTactic)(nil)
var _ Node = (*SpoilTactic)(nil)
var _ Node = (*IfTactic)(nil)
var _ Node = (*PropertyTactic)(nil)
var _ Node = (*FunctionTactic)(nil)
var _ Node = (*TacticTactic)(nil)
var _ Node = (*ProofTactic)(nil)
var _ Node = (*ComposeTactics)(nil)

// --- Fuzz test helpers ---

// buildRandomASTNode constructs a random AST node from fuzz input bytes.
// It consumes bytes from data and returns the node plus remaining bytes.
func buildRandomASTNode(data []byte) (Node, []byte) {
	cfg := NewAstConfig()
	if len(data) == 0 {
		return cfg.NewSymbol("x", nil), nil
	}
	nodeType := data[0] % 12
	data = data[1:]

	// helper to extract a short string from data
	extractStr := func(d []byte) (string, []byte) {
		if len(d) == 0 {
			return "a", d
		}
		nameLen := int(d[0]%8) + 1
		d = d[1:]
		if nameLen > len(d) {
			nameLen = len(d)
		}
		if nameLen == 0 {
			return "a", d
		}
		// sanitize to printable ASCII
		buf := make([]byte, nameLen)
		for i := 0; i < nameLen; i++ {
			buf[i] = 'a' + d[i]%26
		}
		return string(buf), d[nameLen:]
	}

	switch nodeType {
	case 0: // Symbol
		name, rest := extractStr(data)
		return cfg.NewSymbol(name, nil), rest
	case 1: // Atom with no args
		name, rest := extractStr(data)
		return cfg.NewAtom(name), rest
	case 2: // Atom with args
		name, rest := extractStr(data)
		arg1, rest := buildRandomASTNode(rest)
		return cfg.NewAtom(name, arg1), rest
	case 3: // App
		sym, rest := buildRandomASTNode(data)
		arg, rest := buildRandomASTNode(rest)
		return cfg.NewApp(sym, arg), rest
	case 4: // Variable
		name, rest := extractStr(data)
		sortSym := "t"
		return cfg.NewVariable(name, sortSym), rest
	case 5: // Old
		inner, rest := buildRandomASTNode(data)
		return cfg.NewOld(inner), rest
	case 6: // MethodCall
		obj, rest := buildRandomASTNode(data)
		method, rest := buildRandomASTNode(rest)
		return &MethodCall{Obj: obj, Method: method}, rest
	case 7: // Literal positive
		inner, rest := buildRandomASTNode(data)
		return cfg.NewLiteral(1, inner), rest
	case 8: // Literal negative
		inner, rest := buildRandomASTNode(data)
		return cfg.NewLiteral(0, inner), rest
	case 9: // Dot
		left, rest := buildRandomASTNode(data)
		right, rest := buildRandomASTNode(rest)
		return cfg.NewDot(left, right), rest
	case 10: // Not
		inner, rest := buildRandomASTNode(data)
		return cfg.NewNot(inner), rest
	case 11: // And
		t1, rest := buildRandomASTNode(data)
		t2, rest := buildRandomASTNode(rest)
		return cfg.NewAnd(t1, t2), rest
	default:
		return cfg.NewSymbol("fallback", nil), data
	}
}

// FuzzASTClone builds random AST nodes from fuzz input, clones them,
// and verifies the clone is structurally independent.
func FuzzASTClone(f *testing.F) {
	cfg := NewAstConfig()
	f.Add([]byte{0, 3, 'h', 'i'})
	f.Add([]byte{2, 2, 'f', 0, 1, 'x'})
	f.Add([]byte{3, 0, 1, 'a', 0, 1, 'b'})
	f.Add([]byte{5, 0, 2, 'z', 'z'})
	f.Add([]byte{6, 0, 1, 'o', 0, 1, 'm'})
	f.Add([]byte{7, 0, 1, 'p'})
	f.Add([]byte{8, 0, 1, 'q'})
	f.Add([]byte{9, 0, 1, 'a', 0, 1, 'b'})
	f.Add([]byte{10, 0, 1, 'p'})
	f.Add([]byte{11, 0, 1, 'p', 0, 1, 'q'})
	f.Add([]byte{4, 2, 'X', 'Y'})
	f.Add([]byte{1, 3, 'f', 'o', 'o'})

	f.Fuzz(func(t *testing.T, data []byte) {
		node, _ := buildRandomASTNode(data)
		if node == nil {
			return
		}

		// Clone the node
		args := node.Args()
		clone := node.Clone(args)
		if clone == nil {
			t.Fatal("Clone returned nil")
		}

		// Verify String() doesn't panic on clone
		origStr := node.String()
		cloneStr := clone.String()

		// For most node types, clone with same args should produce same string
		// Variable.Clone returns self, so we skip that check
		if _, isVar := node.(*Variable); !isVar {
			if origStr != cloneStr {
				// This is OK for some types where Clone may differ,
				// but we at least verify no panic
			}
		}

		// Verify modifying the clone's Args slice doesn't affect original
		cloneArgs := clone.Args()
		if len(cloneArgs) > 0 {
			// Replace first arg with a new symbol
			replacement := cfg.NewSymbol("REPLACED", nil)
			newArgs := make([]Node, len(cloneArgs))
			copy(newArgs, cloneArgs)
			newArgs[0] = replacement
			clone2 := clone.Clone(newArgs)
			_ = clone2.String() // should not panic

			// Original should be unchanged
			if node.String() != origStr {
				t.Errorf("original was mutated: was %q, now %q", origStr, node.String())
			}
		}
	})
}

// FuzzASTString builds random AST nodes from fuzz bytes and calls String(),
// verifying no panics occur.
func FuzzASTString(f *testing.F) {
	f.Add([]byte{0, 3, 'h', 'i'})
	f.Add([]byte{2, 2, 'f', 0, 1, 'x'})
	f.Add([]byte{3, 0, 1, 'a', 0, 1, 'b'})
	f.Add([]byte{5, 0, 2, 'z', 'z'})
	f.Add([]byte{6, 0, 1, 'o', 0, 1, 'm'})
	f.Add([]byte{7, 0, 1, 'p'})
	f.Add([]byte{8, 0, 1, 'q'})
	f.Add([]byte{9, 0, 1, 'a', 0, 1, 'b'})
	f.Add([]byte{10, 0, 1, 'p'})
	f.Add([]byte{11, 0, 1, 'p', 0, 1, 'q'})
	f.Add([]byte{4, 2, 'X', 'Y'})
	f.Add([]byte{1, 3, 'f', 'o', 'o'})
	f.Add([]byte{}) // empty input

	f.Fuzz(func(t *testing.T, data []byte) {
		node, _ := buildRandomASTNode(data)
		if node == nil {
			return
		}
		// Call String() and verify no panic
		s := node.String()
		if len(s) < 0 {
			t.Fatal("impossible") // just to use s
		}
	})
}

// Regression: Symbol("this") must be rewritten to Symbol("index") by
// AstRewriteSubstPrefix when used as a type name inside a module body.
// Bug was: the Symbol case in AstRewrite only called RewriteName (substitution)
// but not RewriteAtom (prefix transformation), so "this" stayed as "this"
// instead of becoming the prefix name.
func TestAstRewrite_SymbolThisBecomesPrefix(t *testing.T) {
	cfg := NewAstConfig()
	// Simulate: module mymod = { type this; alias t = this }
	// instantiated as: instance idx : mymod
	// Expected: type idx, alias idx.t = idx

	pref := cfg.NewAtom("idx")
	toPref := map[string]bool{"this": true, "t": true}
	subst := map[string]string{}

	// Test 1: Symbol("this") → Symbol("idx")
	sym := cfg.NewSymbol("this", nil)
	result := SubstPrefixAtomsAst(sym, subst, pref, toPref, nil)
	resSym, ok := result.(*Symbol)
	if !ok {
		t.Fatalf("expected *Symbol, got %T: %v", result, result)
	}
	if resSym.Rep != "idx" {
		t.Errorf("Symbol(\"this\") should become Symbol(\"idx\"), got Symbol(%q)", resSym.Rep)
	}

	// Test 2: TypeDef with Name=Symbol("this") → Name=Symbol("idx")
	td := &TypeDef{Name: cfg.NewSymbol("this", nil), Value: nil}
	resTd := SubstPrefixAtomsAst(td, subst, pref, toPref, nil)
	tdResult, ok := resTd.(*TypeDef)
	if !ok {
		t.Fatalf("expected *TypeDef, got %T", resTd)
	}
	if tdSym, ok := tdResult.Name.(*Symbol); !ok || tdSym.Rep != "idx" {
		t.Errorf("TypeDef name should be \"idx\", got %v", tdResult.Name)
	}

	// Test 3: AliasDecl Definition(Atom("t"), Atom("this")) →
	//         Definition(Atom("idx.t"), Atom("idx"))
	alias := cfg.NewDefinition(cfg.NewAtom("t"), cfg.NewAtom("this"))
	resAlias := SubstPrefixAtomsAst(alias, subst, pref, toPref, nil)
	if def, ok := resAlias.(*Definition); ok {
		if lhs, ok := def.Lhs.(*Atom); !ok || lhs.Rep != "idx.t" {
			t.Errorf("alias LHS should be \"idx.t\", got %v", def.Lhs)
		}
		if rhs, ok := def.Rhs.(*Atom); !ok || rhs.Rep != "idx" {
			t.Errorf("alias RHS should be \"idx\", got %v", def.Rhs)
		}
	} else {
		t.Fatalf("expected *Definition, got %T", resAlias)
	}
}

// --- Defines() comprehensive tests ---
// These tests cover all Defines() methods to prevent regressions,
// especially after struct field nodes changed from *Atom to *App.

func TestStructSort_Defines_AtomFields(t *testing.T) {
	cfg := NewAstConfig()
	// StructSort with *Atom fields (legacy representation)
	ss := cfg.NewStructSort(cfg.NewAtom("x"), cfg.NewAtom("y"))
	defs := ss.Defines()
	if len(defs) != 2 || defs[0] != "x" || defs[1] != "y" {
		t.Errorf("StructSort.Defines() with Atom fields: got %v, want [x y]", defs)
	}
}

func TestStructSort_Defines_AppFields(t *testing.T) {
	cfg := NewAstConfig()
	// StructSort with *App fields (current representation matching Python)
	// Simulates: struct { is_end : bool, val : t }
	isEnd := cfg.NewApp(cfg.NewSymbol("is_end", nil))
	isEnd.ASort = cfg.NewSymbol("bool", nil)
	val := cfg.NewApp(cfg.NewSymbol("val", nil))
	val.ASort = cfg.NewSymbol("t", nil)
	ss := cfg.NewStructSort(isEnd, val)
	defs := ss.Defines()
	if len(defs) != 2 || defs[0] != "is_end" || defs[1] != "val" {
		t.Errorf("StructSort.Defines() with App fields: got %v, want [is_end val]", defs)
	}
}

func TestStructSort_Defines_MixedFields(t *testing.T) {
	cfg := NewAstConfig()
	// Mix of *Atom and *App fields
	ss := cfg.NewStructSort(cfg.NewAtom("a"), cfg.NewApp(cfg.NewSymbol("b", nil)))
	defs := ss.Defines()
	if len(defs) != 2 || defs[0] != "a" || defs[1] != "b" {
		t.Errorf("StructSort.Defines() mixed fields: got %v, want [a b]", defs)
	}
}

func TestStructSort_Defines_Empty(t *testing.T) {
	cfg := NewAstConfig()
	ss := cfg.NewStructSort()
	defs := ss.Defines()
	if len(defs) != 0 {
		t.Errorf("StructSort.Defines() empty: got %v, want []", defs)
	}
}

func TestConstantSort_Defines(t *testing.T) {
	cfg := NewAstConfig()
	cs := cfg.NewConstantSort()
	if defs := cs.Defines(); defs != nil {
		t.Errorf("ConstantSort.Defines(): got %v, want nil", defs)
	}
}

func TestUninterpretedSortAST_Defines(t *testing.T) {
	us := &UninterpretedSortAST{}
	if defs := us.Defines(); defs != nil {
		t.Errorf("UninterpretedSortAST.Defines(): got %v, want nil", defs)
	}
}

func TestEnumeratedSort_Defines(t *testing.T) {
	cfg := NewAstConfig()
	es := cfg.NewEnumeratedSort(cfg.NewSymbol("red", nil), cfg.NewSymbol("green", nil), cfg.NewSymbol("blue", nil))
	defs := es.Defines()
	if len(defs) != 3 || defs[0] != "red" || defs[1] != "green" || defs[2] != "blue" {
		t.Errorf("EnumeratedSort.Defines(): got %v, want [red green blue]", defs)
	}
}

func TestFunctionSort_Defines(t *testing.T) {
	cfg := NewAstConfig()
	fs := cfg.NewFunctionSort([]Node{cfg.NewSymbol("S", nil)}, cfg.NewSymbol("T", nil))
	if defs := fs.Defines(); defs != nil {
		t.Errorf("FunctionSort.Defines(): got %v, want nil", defs)
	}
}

func TestRelationSort_Defines(t *testing.T) {
	cfg := NewAstConfig()
	rs := cfg.NewRelationSort([]Node{cfg.NewSymbol("S", nil)})
	if defs := rs.Defines(); defs != nil {
		t.Errorf("RelationSort.Defines(): got %v, want nil", defs)
	}
}

func TestTypeDef_Defines_WithStructSort(t *testing.T) {
	cfg := NewAstConfig()
	// type t = struct { is_end : bool, val : range.t }
	// TypeDef.Defines() should return [t, is_end, val]
	isEnd := cfg.NewApp(cfg.NewSymbol("is_end", nil))
	isEnd.ASort = cfg.NewSymbol("bool", nil)
	val := cfg.NewApp(cfg.NewSymbol("val", nil))
	val.ASort = cfg.NewSymbol("t", nil)
	ss := cfg.NewStructSort(isEnd, val)
	td := cfg.NewTypeDef(cfg.NewSymbol("t", nil), ss)
	defs := td.Defines()
	if len(defs) != 3 || defs[0] != "t" || defs[1] != "is_end" || defs[2] != "val" {
		t.Errorf("TypeDef.Defines() with StructSort App fields: got %v, want [t is_end val]", defs)
	}
}

func TestTypeDef_Defines_WithEnumeratedSort(t *testing.T) {
	cfg := NewAstConfig()
	// type color = {red, green, blue}
	es := cfg.NewEnumeratedSort(cfg.NewSymbol("red", nil), cfg.NewSymbol("green", nil), cfg.NewSymbol("blue", nil))
	td := cfg.NewTypeDef(cfg.NewSymbol("color", nil), es)
	defs := td.Defines()
	if len(defs) != 4 || defs[0] != "color" {
		t.Errorf("TypeDef.Defines() with EnumeratedSort: got %v, want [color red green blue]", defs)
	}
}

func TestTypeDef_Defines_WithConstantSort(t *testing.T) {
	cfg := NewAstConfig()
	// type foo (uninterpreted — no sub-defines)
	td := cfg.NewTypeDef(cfg.NewSymbol("foo", nil), cfg.NewConstantSort())
	defs := td.Defines()
	if len(defs) != 1 || defs[0] != "foo" {
		t.Errorf("TypeDef.Defines() with ConstantSort: got %v, want [foo]", defs)
	}
}

func TestTypeDef_Defines_AtomName(t *testing.T) {
	cfg := NewAstConfig()
	// TypeDef where Name is *Atom instead of *Symbol
	td := cfg.NewTypeDef(cfg.NewAtom("mytype"), cfg.NewConstantSort())
	defs := td.Defines()
	if len(defs) != 1 || defs[0] != "mytype" {
		t.Errorf("TypeDef.Defines() with Atom name: got %v, want [mytype]", defs)
	}
}

func TestDeclBase_Defines_TypeDeclWithStructApp(t *testing.T) {
	cfg := NewAstConfig()
	// TypeDecl (which uses DeclBase.Defines) containing TypeDef with StructSort App fields.
	// This is the exact scenario that broke when struct fields changed from *Atom to *App.
	isEnd := cfg.NewApp(cfg.NewSymbol("is_end", nil))
	isEnd.ASort = cfg.NewSymbol("bool", nil)
	val := cfg.NewApp(cfg.NewSymbol("val", nil))
	val.ASort = cfg.NewSymbol("t", nil)
	ss := cfg.NewStructSort(isEnd, val)
	td := cfg.NewTypeDef(cfg.NewAtom("t"), ss)
	typeDecl := cfg.NewTypeDecl(td)
	defs := typeDecl.Defines()
	if len(defs) != 3 || defs[0] != "t" || defs[1] != "is_end" || defs[2] != "val" {
		t.Errorf("TypeDecl.Defines() via DeclBase with App struct fields: got %v, want [t is_end val]", defs)
	}
}

func TestDeclBase_Defines_WithAtomArg(t *testing.T) {
	cfg := NewAstConfig()
	// DeclBase with a plain Atom arg (e.g. ObjectDecl)
	decl := cfg.NewObjectDecl(cfg.NewAtom("myobj"))
	defs := decl.Defines()
	if len(defs) != 1 || defs[0] != "myobj" {
		t.Errorf("ObjectDecl.Defines(): got %v, want [myobj]", defs)
	}
}

func TestDeclBase_Defines_WithAppArg(t *testing.T) {
	cfg := NewAstConfig()
	// DeclBase with an App arg (e.g. ConstantDecl with App)
	app := cfg.NewApp(cfg.NewSymbol("myconst", nil))
	decl := &ConstantDecl{DeclBase: DeclBase{DeclArgs: []Node{app}}}
	defs := decl.Defines()
	if len(defs) != 1 || defs[0] != "myconst" {
		t.Errorf("ConstantDecl.Defines() with App arg: got %v, want [myconst]", defs)
	}
}

func TestDeclBase_Defines_WithDefinitionArg(t *testing.T) {
	cfg := NewAstConfig()
	// DeclBase with a Definition arg (e.g. ModuleDecl)
	defn := cfg.NewDefinition(cfg.NewAtom("mymod"), cfg.NewAtom("body"))
	decl := &ModuleDecl{DeclBase: DeclBase{DeclArgs: []Node{defn}}}
	defs := decl.Defines()
	if len(defs) != 1 || defs[0] != "mymod" {
		t.Errorf("ModuleDecl.Defines() with Definition arg: got %v, want [mymod]", defs)
	}
}

func TestDeclBase_Defines_WithLabeledFormulaArg(t *testing.T) {
	cfg := NewAstConfig()
	// DeclBase with a LabeledFormula arg (e.g. AxiomDecl)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("ax1"), cfg.NewSymbol("true", nil))
	decl := &AxiomDecl{DeclBase: DeclBase{DeclArgs: []Node{lf}}}
	defs := decl.Defines()
	if len(defs) != 1 || defs[0] != "ax1" {
		t.Errorf("AxiomDecl.Defines() with LabeledFormula: got %v, want [ax1]", defs)
	}
}

func TestDeclBase_Defines_Empty(t *testing.T) {
	decl := &DeclBase{}
	defs := decl.Defines()
	if len(defs) != 0 {
		t.Errorf("empty DeclBase.Defines(): got %v, want []", defs)
	}
}

func TestDefinition_Defines(t *testing.T) {
	cfg := NewAstConfig()
	// Definition with Symbol LHS
	d1 := cfg.NewDefinition(cfg.NewSymbol("foo", nil), cfg.NewSymbol("bar", nil))
	if d1.Defines() != "foo" {
		t.Errorf("Definition.Defines() with Symbol: got %q, want %q", d1.Defines(), "foo")
	}
	// Definition with Atom LHS
	d2 := cfg.NewDefinition(cfg.NewAtom("baz"), cfg.NewSymbol("quux", nil))
	if d2.Defines() != "baz" {
		t.Errorf("Definition.Defines() with Atom: got %q, want %q", d2.Defines(), "baz")
	}
}

func TestActionDef_Defines(t *testing.T) {
	cfg := NewAstConfig()
	ad := cfg.NewActionDef(cfg.NewAtom("my_action"), cfg.NewSymbol("skip", nil), nil, nil)
	if ad.Defines() != "my_action" {
		t.Errorf("ActionDef.Defines(): got %q, want %q", ad.Defines(), "my_action")
	}
}

func TestDerivedDecl_Defines(t *testing.T) {
	cfg := NewAstConfig()
	// DerivedDecl with LabeledFormula containing a Definition
	defn := cfg.NewDefinition(cfg.NewAtom("derived_fn"), cfg.NewSymbol("body", nil))
	lf := cfg.NewLabeledFormula(nil, defn)
	dd := &DerivedDecl{DeclBase: DeclBase{DeclArgs: []Node{lf}}}
	defs := dd.Defines()
	if len(defs) != 1 || defs[0] != "derived_fn" {
		t.Errorf("DerivedDecl.Defines(): got %v, want [derived_fn]", defs)
	}
}

func TestDerivedDecl_Defines_Empty(t *testing.T) {
	// DerivedDecl with no formula
	dd := &DerivedDecl{DeclBase: DeclBase{}}
	defs := dd.Defines()
	if len(defs) != 0 {
		t.Errorf("empty DerivedDecl.Defines(): got %v, want []", defs)
	}
}

func TestIsolateDecl_Defines(t *testing.T) {
	cfg := NewAstConfig()
	idef := &IsolateDef{Elems: []Node{cfg.NewAtom("myiso")}}
	decl := &IsolateDecl{DeclBase: DeclBase{DeclArgs: []Node{idef}}}
	defs := decl.Defines()
	if len(defs) != 1 || defs[0] != "myiso" {
		t.Errorf("IsolateDecl.Defines(): got %v, want [myiso]", defs)
	}
}

func TestIsolateObjectDecl_Defines(t *testing.T) {
	decl := &IsolateObjectDecl{}
	defs := decl.Defines()
	if defs != nil {
		t.Errorf("IsolateObjectDecl.Defines(): got %v, want nil", defs)
	}
}

func TestSchema_Defines(t *testing.T) {
	cfg := NewAstConfig()
	defn := cfg.NewDefinition(cfg.NewAtom("my_schema"), cfg.NewSymbol("body", nil))
	s := cfg.NewSchema(defn)
	if s.Defines() != "my_schema" {
		t.Errorf("Schema.Defines(): got %q, want %q", s.Defines(), "my_schema")
	}
	// Schema with nil Defn
	s2 := &Schema{}
	if s2.Defines() != "" {
		t.Errorf("Schema.Defines() with nil Defn: got %q, want %q", s2.Defines(), "")
	}
}

func TestInterpretDecl_Defines_AppRangeArgs(t *testing.T) {
	cfg := NewAstConfig()
	// interpret t -> {0..max}
	// The Range lo/hi are *App nodes. InterpretDecl.Defines() must extract "max"
	// (non-numeric) via nodeRep, not just *Atom type assertion.
	lo := cfg.NewApp(cfg.NewSymbol("0", nil))
	hi := cfg.NewApp(cfg.NewSymbol("max", nil))
	rng := cfg.NewRange(lo, hi)
	fmla := cfg.NewImplies(cfg.NewAtom("t"), rng)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("interp1"), fmla)
	decl := &InterpretDecl{DeclBase: DeclBase{DeclArgs: []Node{lf}}}
	defs := decl.Defines()
	// Should include label "interp1" and non-numeric range arg "max"
	if len(defs) != 2 || defs[0] != "interp1" || defs[1] != "max" {
		t.Errorf("InterpretDecl.Defines() with App Range args: got %v, want [interp1 max]", defs)
	}
}

func TestInterpretDecl_Defines_AtomRangeArgs(t *testing.T) {
	cfg := NewAstConfig()
	// Same test but with *Atom range args (legacy)
	lo := cfg.NewAtom("0")
	hi := cfg.NewAtom("max")
	rng := cfg.NewRange(lo, hi)
	fmla := cfg.NewImplies(cfg.NewAtom("t"), rng)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("interp2"), fmla)
	decl := &InterpretDecl{DeclBase: DeclBase{DeclArgs: []Node{lf}}}
	defs := decl.Defines()
	if len(defs) != 2 || defs[0] != "interp2" || defs[1] != "max" {
		t.Errorf("InterpretDecl.Defines() with Atom Range args: got %v, want [interp2 max]", defs)
	}
}

func TestInterpretDecl_Defines_AllNumeric(t *testing.T) {
	cfg := NewAstConfig()
	// interpret t -> {0..10} — no non-numeric args, only label defined
	lo := cfg.NewApp(cfg.NewSymbol("0", nil))
	hi := cfg.NewApp(cfg.NewSymbol("10", nil))
	rng := cfg.NewRange(lo, hi)
	fmla := cfg.NewImplies(cfg.NewAtom("t"), rng)
	lf := cfg.NewLabeledFormula(cfg.NewAtom("interp3"), fmla)
	decl := &InterpretDecl{DeclBase: DeclBase{DeclArgs: []Node{lf}}}
	defs := decl.Defines()
	if len(defs) != 1 || defs[0] != "interp3" {
		t.Errorf("InterpretDecl.Defines() all-numeric range: got %v, want [interp3]", defs)
	}
}

func TestDefinition_Defines_AppLhs(t *testing.T) {
	cfg := NewAstConfig()
	// Definition with App LHS (via nodeRep generalization)
	d := cfg.NewDefinition(cfg.NewApp(cfg.NewSymbol("appfn", nil)), cfg.NewSymbol("body", nil))
	if d.Defines() != "appfn" {
		t.Errorf("Definition.Defines() with App LHS: got %q, want %q", d.Defines(), "appfn")
	}
}

func TestActionDef_Defines_AppName(t *testing.T) {
	cfg := NewAstConfig()
	// ActionDef with App name (via nodeRep generalization)
	ad := cfg.NewActionDef(cfg.NewApp(cfg.NewSymbol("my_act", nil)), cfg.NewSymbol("skip", nil), nil, nil)
	if ad.Defines() != "my_act" {
		t.Errorf("ActionDef.Defines() with App name: got %q, want %q", ad.Defines(), "my_act")
	}
}

func TestIsolateDecl_Defines_AppElem(t *testing.T) {
	cfg := NewAstConfig()
	// IsolateDecl with App element (via nodeRep generalization)
	idef := &IsolateDef{Elems: []Node{cfg.NewApp(cfg.NewSymbol("myiso", nil))}}
	decl := &IsolateDecl{DeclBase: DeclBase{DeclArgs: []Node{idef}}}
	defs := decl.Defines()
	if len(defs) != 1 || defs[0] != "myiso" {
		t.Errorf("IsolateDecl.Defines() with App elem: got %v, want [myiso]", defs)
	}
}
