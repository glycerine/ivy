package dafnygen

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// UniqueRenamer tests
// ---------------------------------------------------------------------------

func TestUniqueRenamerNoHint(t *testing.T) {
	r := NewUniqueRenamer("tmp")
	n1 := r.Next("")
	n2 := r.Next("")
	if n1 == n2 {
		t.Error("names should be unique")
	}
	if !strings.HasPrefix(n1, "tmp") {
		t.Errorf("expected prefix tmp, got %q", n1)
	}
}

func TestUniqueRenamerWithHint(t *testing.T) {
	r := NewUniqueRenamer("pre:")
	n := r.Next("x")
	if !strings.Contains(n, "x") {
		t.Errorf("expected hint in name, got %q", n)
	}
}

// ---------------------------------------------------------------------------
// DafnyType tests
// ---------------------------------------------------------------------------

func TestDafnyTypeIsBool(t *testing.T) {
	bt := &DafnyType{Rep: "bool"}
	if !bt.IsBool() {
		t.Error("expected bool")
	}
	nt := &DafnyType{Rep: "int"}
	if nt.IsBool() {
		t.Error("int is not bool")
	}
}

// ---------------------------------------------------------------------------
// OpMap tests
// ---------------------------------------------------------------------------

func TestMapOp(t *testing.T) {
	if MapOp("==") != "=" {
		t.Errorf("expected = for ==")
	}
	if MapOp("+") != "+" {
		t.Errorf("expected + for +")
	}
}

// ---------------------------------------------------------------------------
// IvyNode rendering tests
// ---------------------------------------------------------------------------

func TestAtomNoArgs(t *testing.T) {
	a := NewAtom("foo")
	if a.IvyString() != "foo" {
		t.Errorf("got %q", a.IvyString())
	}
}

func TestAtomWithArgs(t *testing.T) {
	a := NewAtom("r", NewApp("x"), NewApp("y"))
	got := a.IvyString()
	if got != "r(x,y)" {
		t.Errorf("got %q", got)
	}
}

func TestAppWithSort(t *testing.T) {
	a := NewApp("x")
	a.Sort = "int"
	if a.IvyString() != "x:int" {
		t.Errorf("got %q", a.IvyString())
	}
}

func TestAssignActionString(t *testing.T) {
	a := NewAssignAction(NewApp("x"), NewApp("y"))
	got := a.IvyString()
	if got != "x := y" {
		t.Errorf("got %q", got)
	}
}

func TestSequenceString(t *testing.T) {
	s := NewSequence(
		NewAssignAction(NewApp("x"), NewApp("1")),
		NewAssignAction(NewApp("y"), NewApp("2")),
	)
	got := s.IvyString()
	if !strings.Contains(got, "; ") {
		t.Errorf("expected semicolons, got %q", got)
	}
}

func TestAndEmpty(t *testing.T) {
	a := NewAnd()
	if a.IvyString() != "true" {
		t.Errorf("got %q", a.IvyString())
	}
}

func TestOrEmpty(t *testing.T) {
	o := NewOr()
	if o.IvyString() != "false" {
		t.Errorf("got %q", o.IvyString())
	}
}

func TestAndMultiple(t *testing.T) {
	a := NewAnd(NewApp("p"), NewApp("q"))
	got := a.IvyString()
	if !strings.Contains(got, "&") {
		t.Errorf("expected &, got %q", got)
	}
}

func TestOrMultiple(t *testing.T) {
	o := NewOr(NewApp("p"), NewApp("q"))
	got := o.IvyString()
	if !strings.Contains(got, "|") {
		t.Errorf("expected |, got %q", got)
	}
}

func TestNotString(t *testing.T) {
	n := &Not{Body: NewApp("p")}
	got := n.IvyString()
	if !strings.HasPrefix(got, "~") {
		t.Errorf("expected ~, got %q", got)
	}
}

func TestImpliesString(t *testing.T) {
	imp := &IvyImplies{LHS: NewApp("a"), RHS: NewApp("b")}
	got := imp.IvyString()
	if !strings.Contains(got, "~") || !strings.Contains(got, "|") {
		t.Errorf("expected ~A | B encoding, got %q", got)
	}
}

func TestIffTranslation(t *testing.T) {
	got := TranslateIff(NewApp("a"), NewApp("b"))
	s := got.IvyString()
	if !strings.Contains(s, "&") || !strings.Contains(s, "|") {
		t.Errorf("expected & and | in iff, got %q", s)
	}
}

// ---------------------------------------------------------------------------
// RME tests
// ---------------------------------------------------------------------------

func TestRMEFull(t *testing.T) {
	rme := NewRME(NewApp("req"), []string{"x", "y"}, NewApp("ens"))
	got := rme.IvyString()
	if !strings.Contains(got, "requires") {
		t.Error("missing requires")
	}
	if !strings.Contains(got, "modifies") {
		t.Error("missing modifies")
	}
	if !strings.Contains(got, "ensures") {
		t.Error("missing ensures")
	}
}

func TestRMEEmpty(t *testing.T) {
	rme := NewRME(nil, nil, nil)
	if rme.IvyString() != "" {
		t.Errorf("expected empty, got %q", rme.IvyString())
	}
}

// ---------------------------------------------------------------------------
// Compiler tests
// ---------------------------------------------------------------------------

func TestCompilerDeclareVarBool(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.CompileVarDecl([]*TypedSymbol{
		{Rep: "p", Type: &DafnyType{Rep: "bool"}},
	})
	got := mod.String()
	if !strings.Contains(got, "relation p") {
		t.Errorf("expected relation, got %q", got)
	}
}

func TestCompilerDeclareVarObject(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.CompileVarDecl([]*TypedSymbol{
		{Rep: "x", Type: &DafnyType{Rep: "object"}},
	})
	got := mod.String()
	if !strings.Contains(got, "individual x") {
		t.Errorf("expected individual, got %q", got)
	}
}

func TestCompilerDeclareVarTyped(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.CompileVarDecl([]*TypedSymbol{
		{Rep: "n", Type: &DafnyType{Rep: "int"}},
	})
	got := mod.String()
	if !strings.Contains(got, "int") {
		t.Errorf("expected sort annotation, got %q", got)
	}
}

func TestCompilerAssignSingle(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.MethCtx = NewMethodContext("m", nil, nil)
	c.ScopeCtx = NewScopeContext(nil)

	node, err := c.CompileAssign(
		[]IvyNode{NewApp("x")},
		[]IvyNode{NewApp("y")},
	)
	if err != nil {
		t.Fatal(err)
	}
	got := node.IvyString()
	if got != "x := y" {
		t.Errorf("got %q", got)
	}
}

func TestCompilerAssignMulti(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.MethCtx = NewMethodContext("m", nil, nil)
	c.ScopeCtx = NewScopeContext(nil)

	node, err := c.CompileAssign(
		[]IvyNode{NewApp("a"), NewApp("b")},
		[]IvyNode{NewApp("x"), NewApp("y")},
	)
	if err != nil {
		t.Fatal(err)
	}
	got := node.IvyString()
	if !strings.Contains(got, ":=") {
		t.Errorf("expected assignments, got %q", got)
	}
}

func TestCompilerAssignMismatch(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	_, err := c.CompileAssign(
		[]IvyNode{NewApp("a")},
		[]IvyNode{NewApp("x"), NewApp("y")},
	)
	if err == nil {
		t.Error("expected error for mismatched count")
	}
}

func TestCompilerAssume(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	got := c.CompileAssume(NewApp("p")).IvyString()
	if got != "assume p" {
		t.Errorf("got %q", got)
	}
}

func TestCompilerAssert(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	got := c.CompileAssert(NewApp("q")).IvyString()
	if got != "assert q" {
		t.Errorf("got %q", got)
	}
}

func TestCompilerIf(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	got := c.CompileIf(NewApp("c"), NewApp("t"), NewApp("e")).IvyString()
	if !strings.Contains(got, "if") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "else") {
		t.Errorf("expected else, got %q", got)
	}
}

func TestCompilerReturn(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.MethCtx = NewMethodContext("m", nil, []*TypedSymbol{
		{Rep: "out1", Type: &DafnyType{Rep: "int"}},
	})
	c.ScopeCtx = NewScopeContext(nil)

	got, err := c.CompileReturn([]IvyNode{NewApp("42")})
	if err != nil {
		t.Fatal(err)
	}
	s := got.IvyString()
	if !strings.Contains(s, "out1") {
		t.Errorf("expected out param, got %q", s)
	}
	if !c.ScopeCtx.Returns {
		t.Error("scope should be marked as returning")
	}
}

func TestCompilerReturnMismatch(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.MethCtx = NewMethodContext("m", nil, []*TypedSymbol{
		{Rep: "out1", Type: &DafnyType{Rep: "int"}},
	})
	c.ScopeCtx = NewScopeContext(nil)

	_, err := c.CompileReturn([]IvyNode{NewApp("1"), NewApp("2")})
	if err == nil {
		t.Error("expected error for return count mismatch")
	}
}

func TestCompilerBlock(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.MethCtx = NewMethodContext("m", nil, nil)
	c.ScopeCtx = NewScopeContext(nil)

	got := c.CompileBlock([]IvyNode{
		c.CompileAssume(NewApp("p")),
		c.CompileAssert(NewApp("q")),
	})
	s := got.IvyString()
	if !strings.Contains(s, "assume") || !strings.Contains(s, "assert") {
		t.Errorf("expected assume and assert, got %q", s)
	}
}

func TestCompilerTranslateSymbol(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.ScopeCtx = NewScopeContext(nil)
	c.ScopeCtx.Locals["x"] = &TypedSymbol{Rep: "loc:m:x0", Type: &DafnyType{Rep: "int"}}

	got := c.TranslateSymbol("x").IvyString()
	if !strings.Contains(got, "loc:m:x0") {
		t.Errorf("expected local name, got %q", got)
	}
}

func TestCompilerTranslateSymbolGlobal(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	got := c.TranslateSymbol("y").IvyString()
	if got != "y" {
		t.Errorf("expected y, got %q", got)
	}
}

func TestMakeAppBool(t *testing.T) {
	sym := &TypedSymbol{Rep: "p", Type: &DafnyType{Rep: "bool"}}
	got := MakeApp(sym)
	if _, ok := got.(*Atom); !ok {
		t.Error("expected Atom for bool type")
	}
}

func TestMakeAppNonBool(t *testing.T) {
	sym := &TypedSymbol{Rep: "x", Type: &DafnyType{Rep: "int"}}
	got := MakeApp(sym)
	if _, ok := got.(*App); !ok {
		t.Error("expected App for non-bool type")
	}
}

// ---------------------------------------------------------------------------
// Module tests
// ---------------------------------------------------------------------------

func TestIvyModuleString(t *testing.T) {
	mod := NewIvyModule()
	mod.Declare(&RelationDecl{Atom: NewAtom("p")})
	mod.Declare(&ConstantDecl{App: NewApp("x")})
	got := mod.String()
	if !strings.Contains(got, "relation p") {
		t.Error("missing relation")
	}
	if !strings.Contains(got, "individual x") {
		t.Error("missing individual")
	}
}

// ---------------------------------------------------------------------------
// Declaration rendering tests
// ---------------------------------------------------------------------------

func TestActionDeclString(t *testing.T) {
	d := &ActionDecl{Name: "act", Body: NewSequence(c_assume("p"))}
	got := d.IvyString()
	if !strings.Contains(got, "action act") {
		t.Errorf("got %q", got)
	}
}

func TestStateDeclString(t *testing.T) {
	d := &StateDecl{Name: "s", Def: NewApp("def")}
	got := d.IvyString()
	if !strings.Contains(got, "state s") {
		t.Errorf("got %q", got)
	}
}

func TestCallActionString(t *testing.T) {
	a := &CallAction{Target: NewAtom("foo")}
	got := a.IvyString()
	if got != "call foo" {
		t.Errorf("got %q", got)
	}
}

func TestInstantiateActionString(t *testing.T) {
	a := &InstantiateAction{Target: NewAtom("bar", NewApp("x"))}
	got := a.IvyString()
	if !strings.Contains(got, "instantiate bar") {
		t.Errorf("got %q", got)
	}
}

func TestLocalActionString(t *testing.T) {
	a := &LocalAction{Locals: []string{"a", "b"}, Body: NewApp("skip")}
	got := a.IvyString()
	if !strings.Contains(got, "local a,b") {
		t.Errorf("got %q", got)
	}
}

func TestLetActionString(t *testing.T) {
	a := &LetAction{
		Bindings: []IvyNode{NewApp("binding")},
		Body:     NewApp("body"),
	}
	got := a.IvyString()
	if !strings.Contains(got, "let") {
		t.Errorf("got %q", got)
	}
}

func TestMacroDeclString(t *testing.T) {
	d := &MacroDecl{Name: "m", Def: NewApp("def")}
	got := d.IvyString()
	if !strings.Contains(got, "macro m") {
		t.Errorf("got %q", got)
	}
}

func TestAssertDeclString(t *testing.T) {
	d := &AssertDecl{Formula: NewApp("f")}
	got := d.IvyString()
	if got != "assert f" {
		t.Errorf("got %q", got)
	}
}

// ---------------------------------------------------------------------------
// SubstSymbol tests
// ---------------------------------------------------------------------------

func TestSubstSymbolHit(t *testing.T) {
	m := SubstMap{"x": {Rep: "P:x", Type: &DafnyType{Rep: "int"}}}
	got := SubstSymbol("x", m)
	if got != "P:x" {
		t.Errorf("got %q", got)
	}
}

func TestSubstSymbolMiss(t *testing.T) {
	m := SubstMap{}
	got := SubstSymbol("y", m)
	if got != "y" {
		t.Errorf("got %q", got)
	}
}

// ---------------------------------------------------------------------------
// InferType tests
// ---------------------------------------------------------------------------

func TestInferTypeFromScope(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	c.ScopeCtx = NewScopeContext(nil)
	c.ScopeCtx.Locals["x"] = &TypedSymbol{Rep: "x", Type: &DafnyType{Rep: "int"}}
	typ, err := c.InferType("x")
	if err != nil {
		t.Fatal(err)
	}
	if typ.Rep != "int" {
		t.Errorf("expected int, got %q", typ.Rep)
	}
}

func TestInferTypeNotFound(t *testing.T) {
	mod := NewIvyModule()
	c := NewCompiler(mod)
	_, err := c.InferType("unknown")
	if err == nil {
		t.Error("expected error")
	}
}

func TestInferInfixTypeArith(t *testing.T) {
	typ := InferInfixType("+", &DafnyType{Rep: "int"})
	if typ.Rep != "int" {
		t.Errorf("expected int, got %q", typ.Rep)
	}
}

func TestInferInfixTypeRelational(t *testing.T) {
	typ := InferInfixType("<=", &DafnyType{Rep: "int"})
	if typ.Rep != "bool" {
		t.Errorf("expected bool, got %q", typ.Rep)
	}
}

// ---------------------------------------------------------------------------
// Preamble test
// ---------------------------------------------------------------------------

func TestPreambleContent(t *testing.T) {
	if !strings.Contains(Preamble, "type int") {
		t.Error("missing type int")
	}
	if !strings.Contains(Preamble, "interpret") {
		t.Error("missing interpret")
	}
}

// ---------------------------------------------------------------------------
// Scope context tests
// ---------------------------------------------------------------------------

func TestScopeContextCopiesParent(t *testing.T) {
	parent := NewScopeContext(nil)
	parent.Locals["x"] = &TypedSymbol{Rep: "x", Type: &DafnyType{Rep: "int"}}

	child := NewScopeContext(parent)
	if _, ok := child.Locals["x"]; !ok {
		t.Error("child should inherit parent locals")
	}

	// Mutating child doesn't affect parent
	child.Locals["y"] = &TypedSymbol{Rep: "y", Type: &DafnyType{Rep: "bool"}}
	if _, ok := parent.Locals["y"]; ok {
		t.Error("parent should not be affected by child")
	}
}

// ---------------------------------------------------------------------------
// Fuzz tests
// ---------------------------------------------------------------------------

func FuzzMapOp(f *testing.F) {
	f.Add("==")
	f.Add("+")
	f.Add("")
	f.Add("<=")
	f.Add("&&")

	f.Fuzz(func(t *testing.T, op string) {
		got := MapOp(op)
		// Should never panic; result should be non-empty if input is "=="
		if op == "==" && got != "=" {
			t.Errorf("expected = for ==, got %q", got)
		}
	})
}

// helpers for test readability
func c_assume(name string) IvyNode {
	return &AssumeAction{Formula: NewApp(name)}
}
