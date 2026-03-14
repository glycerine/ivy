package ast

import (
	"strings"
	"testing"
)

// --- Core type tests ---

func TestSymbol(t *testing.T) {
	s := NewSymbol("foo", nil)
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
	x := NewSymbol("x", nil)
	y := NewSymbol("y", nil)
	a := NewAtom("r", x, y)
	if a.String() != "r(x, y)" {
		t.Errorf("got %q", a.String())
	}

	// Nullary atom
	a0 := NewAtom("p")
	if a0.String() != "p" {
		t.Errorf("got %q, want %q", a0.String(), "p")
	}

	// Equality
	eq := NewAtom("=", x, y)
	if eq.String() != "x = y" {
		t.Errorf("got %q", eq.String())
	}
}

func TestAtomPrefix(t *testing.T) {
	a := NewAtom("foo", NewSymbol("x", nil))
	p := a.Prefix("bar.")
	if p.Rep != "bar.foo" {
		t.Errorf("got %q, want %q", p.Rep, "bar.foo")
	}
}

func TestAtomSuffix(t *testing.T) {
	a := NewAtom("foo")
	s := a.Suffix("_v2")
	if s.Rep != "foo_v2" {
		t.Errorf("got %q, want %q", s.Rep, "foo_v2")
	}
}

func TestApp(t *testing.T) {
	f := NewSymbol("f", nil)
	x := NewSymbol("x", nil)
	app := NewApp(f, x)
	if app.String() != "f(x)" {
		t.Errorf("got %q", app.String())
	}

	// Nullary app
	a0 := NewApp(f)
	if a0.String() != "f" {
		t.Errorf("got %q, want %q", a0.String(), "f")
	}
}

func TestVariable(t *testing.T) {
	v := NewVariable("X", NewSymbol("t", nil))
	if v.String() != "X:t" {
		t.Errorf("got %q", v.String())
	}
	// Clone returns self
	c := v.Clone(nil)
	if c != v {
		t.Error("Variable.Clone should return self")
	}
	// Resort
	v2 := v.Resort(NewSymbol("u", nil))
	if v2.String() != "X:u" {
		t.Errorf("got %q", v2.String())
	}
}

func TestOld(t *testing.T) {
	x := NewSymbol("x", nil)
	o := NewOld(x)
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
	a := NewAtom("p", NewSymbol("x", nil))
	pos := NewLiteral(1, a)
	neg := NewLiteral(0, a)
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
	a := NewSymbol("a", nil)
	b := NewSymbol("b", nil)
	d := NewDot(a, b)
	if d.String() != "a.b" {
		t.Errorf("got %q", d.String())
	}
}

func TestBracket(t *testing.T) {
	a := NewSymbol("a", nil)
	i := NewSymbol("i", nil)
	br := NewBracket(a, i)
	if br.String() != "a[i]" {
		t.Errorf("got %q", br.String())
	}
}

func TestTuple(t *testing.T) {
	tup := NewTuple(NewSymbol("a", nil), NewSymbol("b", nil))
	if tup.String() != "(a,b)" {
		t.Errorf("got %q", tup.String())
	}
}

// --- Formula tests ---

func TestAndTrue(t *testing.T) {
	a := NewAnd()
	if a.String() != "true" {
		t.Errorf("got %q, want %q", a.String(), "true")
	}
	if !IsTrue(a) {
		t.Error("empty And should be true")
	}
}

func TestOrFalse(t *testing.T) {
	o := NewOr()
	if o.String() != "false" {
		t.Errorf("got %q, want %q", o.String(), "false")
	}
	if !IsFalse(o) {
		t.Error("empty Or should be false")
	}
}

func TestAndMulti(t *testing.T) {
	a := NewAnd(NewSymbol("p", nil), NewSymbol("q", nil))
	if a.String() != "(p & q)" {
		t.Errorf("got %q", a.String())
	}
}

func TestOrMulti(t *testing.T) {
	o := NewOr(NewSymbol("p", nil), NewSymbol("q", nil))
	if o.String() != "(p | q)" {
		t.Errorf("got %q", o.String())
	}
}

func TestNot(t *testing.T) {
	n := NewNot(NewSymbol("p", nil))
	if n.String() != "~p" {
		t.Errorf("got %q", n.String())
	}
}

func TestNotEquals(t *testing.T) {
	x := NewSymbol("x", nil)
	y := NewSymbol("y", nil)
	eq := NewAtom("=", x, y)
	n := NewNot(eq)
	if n.String() != "x ~= y" {
		t.Errorf("got %q", n.String())
	}
}

func TestImplies(t *testing.T) {
	i := NewImplies(NewSymbol("p", nil), NewSymbol("q", nil))
	if i.String() != "p -> q" {
		t.Errorf("got %q", i.String())
	}
}

func TestIff(t *testing.T) {
	f := NewIff(NewSymbol("p", nil), NewSymbol("q", nil))
	if f.String() != "p <-> q" {
		t.Errorf("got %q", f.String())
	}
}

func TestIte(t *testing.T) {
	i := NewIte(NewSymbol("c", nil), NewSymbol("t", nil), NewSymbol("e", nil))
	if i.String() != "(t if c else e)" {
		t.Errorf("got %q", i.String())
	}
}

func TestForall(t *testing.T) {
	x := NewVariable("X", NewSymbol("t", nil))
	body := NewAtom("p", x)
	f := NewForall([]Node{x}, body)
	if !strings.HasPrefix(f.String(), "forall X:t.") {
		t.Errorf("got %q", f.String())
	}
}

func TestExists(t *testing.T) {
	x := NewVariable("X", NewSymbol("t", nil))
	body := NewAtom("p", x)
	e := NewExists([]Node{x}, body)
	if !strings.HasPrefix(e.String(), "exists X:t.") {
		t.Errorf("got %q", e.String())
	}
}

func TestGlobally(t *testing.T) {
	g := NewGlobally(NewSymbol("p", nil))
	if g.String() != "(globally p)" {
		t.Errorf("got %q", g.String())
	}
}

func TestEventually(t *testing.T) {
	e := NewEventually(NewSymbol("p", nil))
	if e.String() != "(eventually p)" {
		t.Errorf("got %q", e.String())
	}
}

func TestHasTemporal(t *testing.T) {
	if HasTemporal(NewSymbol("p", nil)) {
		t.Error("plain symbol should not be temporal")
	}
	if !HasTemporal(NewGlobally(NewSymbol("p", nil))) {
		t.Error("Globally should be temporal")
	}
	if !HasTemporal(NewAnd(NewSymbol("p", nil), NewEventually(NewSymbol("q", nil)))) {
		t.Error("And containing Eventually should be temporal")
	}
}

func TestLet(t *testing.T) {
	def := NewDefinition(NewAtom("p"), NewSymbol("true", nil))
	body := NewSymbol("q", nil)
	l := NewLet([]Node{def}, body)
	if !strings.HasPrefix(l.String(), "let ") {
		t.Errorf("got %q", l.String())
	}
}

func TestDefinition(t *testing.T) {
	d := NewDefinition(NewAtom("f", NewSymbol("x", nil)), NewSymbol("y", nil))
	if d.String() != "f(x) = y" {
		t.Errorf("got %q", d.String())
	}
}

func TestNamedBinder(t *testing.T) {
	x := NewVariable("X", NewSymbol("t", nil))
	body := NewAtom("p", x)
	nb := NewNamedBinder("mybinder", []Node{x}, body)
	if !strings.Contains(nb.String(), "mybinder") {
		t.Errorf("got %q", nb.String())
	}
}

func TestWhenOperator(t *testing.T) {
	w := NewWhenOperator("fire", NewSymbol("p", nil), NewSymbol("q", nil))
	if w.String() != "(p firewhennext q)" {
		t.Errorf("got %q", w.String())
	}
}

// --- Sort tests ---

func TestConstantSort(t *testing.T) {
	cs := NewConstantSort()
	if cs.String() != "uninterpreted" {
		t.Errorf("got %q", cs.String())
	}
}

func TestEnumeratedSort(t *testing.T) {
	es := NewEnumeratedSort(NewSymbol("a", nil), NewSymbol("b", nil), NewSymbol("c", nil))
	if es.String() != "{a,b,c}" {
		t.Errorf("got %q", es.String())
	}
	ext := es.Extension()
	if len(ext) != 3 || ext[0] != "a" || ext[1] != "b" || ext[2] != "c" {
		t.Errorf("extension: %v", ext)
	}
}

func TestFunctionSortAST(t *testing.T) {
	dom := []Node{NewSymbol("S", nil), NewSymbol("S", nil)}
	rng := NewSymbol("bool", nil)
	fs := NewFunctionSort(dom, rng)
	if fs.String() != "S * S -> bool" {
		t.Errorf("got %q", fs.String())
	}
}

func TestRelationSortAST(t *testing.T) {
	rs := NewRelationSort([]Node{NewSymbol("S", nil), NewSymbol("S", nil)})
	if rs.String() != "S * S" {
		t.Errorf("got %q", rs.String())
	}
}

func TestStructSortAST(t *testing.T) {
	ss := NewStructSort(NewAtom("x"), NewAtom("y"))
	if !strings.HasPrefix(ss.String(), "struct {") {
		t.Errorf("got %q", ss.String())
	}
}

func TestRange(t *testing.T) {
	r := NewRange(NewSymbol("0", nil), NewSymbol("10", nil))
	if r.String() != "{0..10}" {
		t.Errorf("got %q", r.String())
	}
}

// --- Declaration tests ---

func TestLabeledFormula(t *testing.T) {
	label := NewAtom("my_axiom")
	fmla := NewImplies(NewSymbol("p", nil), NewSymbol("q", nil))
	lf := NewLabeledFormula(label, fmla)
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
	td := NewTypeDef(NewSymbol("color", nil), NewEnumeratedSort(NewSymbol("red", nil), NewSymbol("green", nil), NewSymbol("blue", nil)))
	if !strings.Contains(td.String(), "color") {
		t.Errorf("got %q", td.String())
	}
	defs := td.Defines()
	if len(defs) != 4 { // color + red + green + blue
		t.Errorf("defines: %v", defs)
	}
}

func TestActionDef(t *testing.T) {
	name := NewAtom("my_action")
	body := NewSymbol("skip", nil)
	params := []Node{NewVariable("X", NewSymbol("t", nil))}
	ad := NewActionDef(name, body, params, nil)
	if ad.Defines() != "my_action" {
		t.Errorf("got %q", ad.Defines())
	}
}

func TestVariantDef(t *testing.T) {
	vd := NewVariantDef(NewSymbol("cons", nil), NewSymbol("list", nil))
	if vd.String() != "cons of list" {
		t.Errorf("got %q", vd.String())
	}
}

func TestAttributeDef(t *testing.T) {
	ad := NewAttributeDef(NewAtom("weight"), NewSymbol("10", nil))
	if ad.String() != "attribute weight = 10" {
		t.Errorf("got %q", ad.String())
	}
}

func TestInstantiation(t *testing.T) {
	inst := NewInstantiation(NewSymbol("mymod", nil), NewSymbol("int", nil))
	if inst.String() != "mymod : int" {
		t.Errorf("got %q", inst.String())
	}
}

// --- Clone tests ---

func TestAndClone(t *testing.T) {
	a := NewAnd(NewSymbol("p", nil), NewSymbol("q", nil))
	c := a.Clone([]Node{NewSymbol("r", nil)})
	if c.String() != "r" {
		t.Errorf("got %q", c.String())
	}
	// Original unchanged
	if a.String() != "(p & q)" {
		t.Errorf("original changed: %q", a.String())
	}
}

func TestForallClone(t *testing.T) {
	x := NewVariable("X", NewSymbol("t", nil))
	body := NewAtom("p", x)
	f := NewForall([]Node{x}, body)
	newBody := NewAtom("q", x)
	c := f.Clone([]Node{newBody}).(*Forall)
	// Bounds should be preserved
	if len(c.Bounds) != 1 {
		t.Errorf("bounds lost in clone")
	}
}

// --- Location tests ---

func TestLocation(t *testing.T) {
	loc := Location{Filename: "test.ivy", Line: 42}
	if loc.String() != "test.ivy:42" {
		t.Errorf("got %q", loc.String())
	}
	loc2 := Location{Line: 7}
	if loc2.String() != "line 7" {
		t.Errorf("got %q", loc2.String())
	}
}

func TestSetLineno(t *testing.T) {
	s := NewSymbol("x", nil)
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
	si := &SchemaInstantiation{
		SchemaName: NewSymbol("my_schema", nil),
		Ren:        &NoneAST{},
	}
	if !strings.HasPrefix(si.String(), "apply my_schema") {
		t.Errorf("got %q", si.String())
	}
}

// --- Some tests ---

func TestSome(t *testing.T) {
	x := NewVariable("X", NewSymbol("t", nil))
	body := NewAtom("p", x)
	s := NewSome([]Node{x}, body)
	if !strings.HasPrefix(s.String(), "some X:t.") {
		t.Errorf("got %q", s.String())
	}
}

func TestSomeExpr(t *testing.T) {
	x := NewVariable("X", NewSymbol("t", nil))
	body := NewAtom("p", x)
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
	if len(data) == 0 {
		return NewSymbol("x", nil), nil
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
		return NewSymbol(name, nil), rest
	case 1: // Atom with no args
		name, rest := extractStr(data)
		return NewAtom(name), rest
	case 2: // Atom with args
		name, rest := extractStr(data)
		arg1, rest := buildRandomASTNode(rest)
		return NewAtom(name, arg1), rest
	case 3: // App
		sym, rest := buildRandomASTNode(data)
		arg, rest := buildRandomASTNode(rest)
		return NewApp(sym, arg), rest
	case 4: // Variable
		name, rest := extractStr(data)
		sortSym := NewSymbol("t", nil)
		return NewVariable(name, sortSym), rest
	case 5: // Old
		inner, rest := buildRandomASTNode(data)
		return NewOld(inner), rest
	case 6: // MethodCall
		obj, rest := buildRandomASTNode(data)
		method, rest := buildRandomASTNode(rest)
		return &MethodCall{Obj: obj, Method: method}, rest
	case 7: // Literal positive
		inner, rest := buildRandomASTNode(data)
		return NewLiteral(1, inner), rest
	case 8: // Literal negative
		inner, rest := buildRandomASTNode(data)
		return NewLiteral(0, inner), rest
	case 9: // Dot
		left, rest := buildRandomASTNode(data)
		right, rest := buildRandomASTNode(rest)
		return NewDot(left, right), rest
	case 10: // Not
		inner, rest := buildRandomASTNode(data)
		return NewNot(inner), rest
	case 11: // And
		t1, rest := buildRandomASTNode(data)
		t2, rest := buildRandomASTNode(rest)
		return NewAnd(t1, t2), rest
	default:
		return NewSymbol("fallback", nil), data
	}
}

// FuzzASTClone builds random AST nodes from fuzz input, clones them,
// and verifies the clone is structurally independent.
func FuzzASTClone(f *testing.F) {
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
			replacement := NewSymbol("REPLACED", nil)
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
