package actions

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// helper: make a simple logic const for testing.
func mkConst(name string) *lg.Const {
	return lg.NewConst(name, lg.Boolean)
}

// helper: make a sort
func mkSort(name string) lg.Sort {
	return &lg.UninterpretedSort{Name: name}
}

// --- Action interface compliance ---

func TestSequenceBasic(t *testing.T) {
	s := NewSequence()
	if s.Name() != "sequence" {
		t.Errorf("Name() = %q, want %q", s.Name(), "sequence")
	}
	if s.String() != "{}" {
		t.Errorf("String() = %q, want %q", s.String(), "{}")
	}
	if len(s.ActionArgs()) != 0 {
		t.Errorf("Args() len = %d, want 0", len(s.ActionArgs()))
	}
}

func TestAssumeAction(t *testing.T) {
	fmla := mkConst("p")
	a := NewAssumeAction(fmla)
	if a.Name() != "assume" {
		t.Errorf("Name() = %q", a.Name())
	}
	if !strings.Contains(a.String(), "assume") {
		t.Errorf("String() = %q, should contain 'assume'", a.String())
	}
	args := a.ActionArgs()
	if len(args) != 1 {
		t.Fatalf("Args() len = %d, want 1", len(args))
	}
}

func TestAssertAction(t *testing.T) {
	fmla := mkConst("q")
	a := NewAssertAction(fmla)
	if a.Name() != "assert" {
		t.Errorf("Name() = %q", a.Name())
	}
	if !strings.Contains(a.String(), "assert") {
		t.Errorf("String() = %q", a.String())
	}
}

func TestRequiresAction(t *testing.T) {
	a := NewRequiresAction(mkConst("r"))
	if a.Name() != "assert" { // Python: RequiresAction inherits name() → "assert"
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestEnsuresAction(t *testing.T) {
	a := NewEnsuresAction(mkConst("e"))
	if a.Name() != "assert" { // Python: EnsuresAction inherits name() → "assert"
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestAssignAction(t *testing.T) {
	lhs := mkConst("x")
	rhs := mkConst("y")
	a := NewAssignAction(lhs, rhs)
	if a.Name() != "assign" {
		t.Errorf("Name() = %q", a.Name())
	}
	if !strings.Contains(a.String(), ":=") {
		t.Errorf("String() = %q, should contain ':='", a.String())
	}
	if len(a.ActionArgs()) != 2 {
		t.Errorf("Args() len = %d, want 2", len(a.ActionArgs()))
	}
}

func TestHavocAction(t *testing.T) {
	a := NewHavocAction(mkConst("x"))
	if a.Name() != "havoc" {
		t.Errorf("Name() = %q", a.Name())
	}
	if !strings.Contains(a.String(), ":= *") {
		t.Errorf("String() = %q, should contain ':= *'", a.String())
	}
}

func TestSetAction(t *testing.T) {
	a := NewSetAction(mkConst("lit"))
	if a.Name() != "set" {
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestIfAction(t *testing.T) {
	cond := mkConst("c")
	thenB := NewSequence()
	a := NewIfAction(cond, thenB)
	if a.Name() != "if" {
		t.Errorf("Name() = %q", a.Name())
	}
	if len(a.ActionArgs()) != 2 {
		t.Errorf("Args() len = %d, want 2", len(a.ActionArgs()))
	}

	// With else branch
	elseB := NewSequence()
	a2 := NewIfAction(cond, thenB, elseB)
	if len(a2.ActionArgs()) != 3 {
		t.Errorf("Args() len = %d, want 3", len(a2.ActionArgs()))
	}
	if !strings.Contains(a2.String(), "else") {
		t.Errorf("String() = %q, should contain 'else'", a2.String())
	}
}

func TestWhileAction(t *testing.T) {
	cond := mkConst("c")
	body := NewSequence()
	a := NewWhileAction(cond, body)
	if a.Name() != "while" {
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestChoiceAction(t *testing.T) {
	b1 := NewSequence()
	b2 := NewSequence()
	a := NewChoiceAction(b1, b2)
	if a.Name() != "choice" {
		t.Errorf("Name() = %q", a.Name())
	}
	if len(a.ActionArgs()) != 2 {
		t.Errorf("Args() len = %d, want 2", len(a.ActionArgs()))
	}
}

func TestCallAction(t *testing.T) {
	callee := mkConst("myaction")
	a := NewCallActionOn(NewActionsConfig(), callee)
	if a.Name() != "call" {
		t.Errorf("Name() = %q", a.Name())
	}
	calls := a.IterCalls()
	if len(calls) != 1 || calls[0] != "myaction" {
		t.Errorf("IterCalls() = %v, want [myaction]", calls)
	}
}

func TestLocalAction(t *testing.T) {
	body := NewSequence()
	local := mkConst("v")
	actCfg := NewActionsConfig()
	a := NewLocalActionOn(actCfg, "test", local, body)
	if a.Name() != "local" {
		t.Errorf("Name() = %q", a.Name())
	}
	if len(a.ActionArgs()) != 2 {
		t.Errorf("Args() len = %d, want 2", len(a.ActionArgs()))
	}
}

func TestLetAction(t *testing.T) {
	body := NewSequence()
	binding := mkConst("b")
	a := NewLetAction(binding, body)
	if a.Name() != "let" {
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestBindOldsAction(t *testing.T) {
	inner := NewSequence()
	a := NewBindOldsAction(inner)
	if a.Name() != "bindolds" {
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestNativeAction(t *testing.T) {
	code := mkConst("code_blob")
	a := NewNativeAction(code)
	if a.Name() != "native" {
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestCrashAction(t *testing.T) {
	a := NewCrashAction(mkConst("target"))
	if a.Name() != "crash" {
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestThunkAction(t *testing.T) {
	a := NewThunkAction(mkConst("a"), mkConst("b"))
	if a.Name() != "thunk" {
		t.Errorf("Name() = %q", a.Name())
	}
}

func TestEnvAction(t *testing.T) {
	a := NewEnvAction(NewSequence())
	if a.Name() != "env" {
		t.Errorf("Name() = %q", a.Name())
	}
	// EnvAction should always return empty formals.
	if a.GetFormalParams() != nil {
		t.Error("EnvAction should have nil formal params")
	}
	if a.GetFormalReturns() != nil {
		t.Error("EnvAction should have nil formal returns")
	}
}

func TestReturnAction(t *testing.T) {
	a := NewReturnAction()
	if a.Name() != "return" {
		t.Errorf("Name() = %q", a.Name())
	}
	if a.String() != "return" {
		t.Errorf("String() = %q", a.String())
	}
}

func TestIgnoreAction(t *testing.T) {
	a := NewIgnoreAction()
	if a.Name() != "ignore" {
		t.Errorf("Name() = %q", a.Name())
	}
}

// --- Clone ---

func TestClonePreservesFormals(t *testing.T) {
	fmla := mkConst("p")
	a := NewAssumeAction(fmla)
	params := []*lg.Const{mkConst("x"), mkConst("y")}
	returns := []*lg.Const{mkConst("r")}
	a.SetFormalParams(params)
	a.SetFormalReturns(returns)
	a.SetLineno(ast.Location{Filename: "test.ivy", Line: 42})

	cloned := a.ActionClone(a.ActionArgs())
	if cloned.GetLineno().Line != 42 {
		t.Errorf("Clone lost lineno: got %d", cloned.GetLineno().Line)
	}
}

func TestSequenceClone(t *testing.T) {
	c1 := mkConst("a")
	c2 := mkConst("b")
	s := NewSequence(c1, c2)
	s.SetLineno(ast.Location{Line: 10})

	cloned := s.ActionClone([]lg.Expr{mkConst("x")})
	if len(cloned.ActionArgs()) != 1 {
		t.Errorf("Cloned args len = %d, want 1", len(cloned.ActionArgs()))
	}
}

// --- IterCalls / IterSubactions ---

func TestIterCallsNested(t *testing.T) {
	call := NewCallActionOn(NewActionsConfig(), mkConst("foo"))
	seq := NewSequence(call)
	calls := seq.IterCalls()
	if len(calls) != 1 || calls[0] != "foo" {
		t.Errorf("IterCalls() = %v, want [foo]", calls)
	}
}

func TestIterSubactions(t *testing.T) {
	inner := NewAssumeAction(mkConst("p"))
	seq := NewSequence(inner)
	subs := seq.IterSubactions()
	// Should include seq itself + the assume
	if len(subs) != 2 {
		t.Errorf("IterSubactions() len = %d, want 2", len(subs))
	}
}

// --- Formal params/returns ---

func TestFormalParams(t *testing.T) {
	a := NewSequence()
	if a.GetFormalParams() != nil {
		t.Error("Initial formal params should be nil")
	}
	params := []*lg.Const{mkConst("x")}
	a.SetFormalParams(params)
	got := a.GetFormalParams()
	if len(got) != 1 || got[0].Name != "x" {
		t.Errorf("GetFormalParams() = %v", got)
	}
}

// --- Lineno ---

func TestLineno(t *testing.T) {
	a := NewSequence()
	loc := ast.Location{Filename: "test.ivy", Line: 5}
	a.SetLineno(loc)
	got := a.GetLineno()
	if got.Filename != "test.ivy" || got.Line != 5 {
		t.Errorf("GetLineno() = %v", got)
	}
}

// --- CopyFormalsTo ---

func TestCopyFormalsTo(t *testing.T) {
	src := NewSequence()
	src.SetFormalParams([]*lg.Const{mkConst("a")})
	src.SetFormalReturns([]*lg.Const{mkConst("b")})

	dst := NewSequence()
	src.CopyFormalsTo(dst)

	if len(dst.GetFormalParams()) != 1 || dst.GetFormalParams()[0].Name != "a" {
		t.Error("CopyFormalsTo did not copy params")
	}
	if len(dst.GetFormalReturns()) != 1 || dst.GetFormalReturns()[0].Name != "b" {
		t.Error("CopyFormalsTo did not copy returns")
	}
}

// --- Schema ---

func TestSchema(t *testing.T) {
	s := NewSchema(mkConst("defn"))
	if !strings.Contains(s.String(), "defn") {
		t.Errorf("Schema.String() = %q", s.String())
	}
}

// --- RME ---

func TestRME(t *testing.T) {
	r := NewRME(mkConst("pre"), []string{"x", "y"}, mkConst("post"))
	s := r.String()
	if !strings.Contains(s, "requires") || !strings.Contains(s, "modifies") || !strings.Contains(s, "ensures") {
		t.Errorf("RME.String() = %q", s)
	}
}

// --- Helpers ---

func TestConcatActions(t *testing.T) {
	a1 := NewAssumeAction(mkConst("p"))
	a2 := NewAssertAction(mkConst("q"))
	seq := ConcatActions(a1, a2)
	if len(seq.Elems) != 2 {
		t.Errorf("ConcatActions len = %d, want 2", len(seq.Elems))
	}

	// Flattening: concat with an existing sequence.
	a3 := NewAssumeAction(mkConst("r"))
	seq2 := ConcatActions(seq, a3)
	if len(seq2.Elems) != 3 {
		t.Errorf("ConcatActions (flatten) len = %d, want 3", len(seq2.Elems))
	}
}

func TestHasCode(t *testing.T) {
	empty := NewSequence()
	if HasCode(empty) {
		t.Error("Empty sequence should have no code")
	}

	withCode := NewSequence(NewAssumeAction(mkConst("p")))
	if !HasCode(withCode) {
		t.Error("Sequence with assume should have code")
	}
}

func TestCallSet(t *testing.T) {
	env := map[string]Action{
		"a": NewCallActionOn(NewActionsConfig(), mkConst("b")),
		"b": NewCallActionOn(NewActionsConfig(), mkConst("c")),
		"c": NewSequence(),
	}
	result := CallSet("a", env)
	// Should include a, b, c
	if len(result) != 3 {
		t.Errorf("CallSet len = %d, want 3", len(result))
	}
}

func TestPrefixAction(t *testing.T) {
	body := NewSequence()
	body.SetFormalParams([]*lg.Const{mkConst("p")})
	body.SetLineno(ast.Location{Line: 5})

	stmt := NewAssumeAction(mkConst("pre"))
	result := PrefixAction(body, []Action{stmt})

	if _, ok := result.(*Sequence); !ok {
		t.Error("PrefixAction should return a Sequence")
	}
	if result.GetLineno().Line != 5 {
		t.Errorf("PrefixAction should preserve lineno, got %d", result.GetLineno().Line)
	}
}

func TestPostfixAction(t *testing.T) {
	body := NewSequence()
	// No stmts -> returns body unchanged.
	result := PostfixAction(body, nil)
	if result != body {
		t.Error("PostfixAction with no stmts should return original")
	}

	stmt := NewAssumeAction(mkConst("post"))
	result = PostfixAction(body, []Action{stmt})
	if _, ok := result.(*Sequence); !ok {
		t.Error("PostfixAction should return a Sequence")
	}
}

func TestParamsToStr(t *testing.T) {
	params := []*lg.Const{
		lg.NewConst("fml:x", mkSort("S")),
		lg.NewConst("y", lg.Boolean),
	}
	s := ParamsToStr(params)
	if !strings.Contains(s, "x:S") {
		t.Errorf("ParamsToStr should strip fml: prefix, got %q", s)
	}
	if !strings.Contains(s, "y:") {
		t.Errorf("ParamsToStr should include y, got %q", s)
	}
}

func TestActionDefToStr(t *testing.T) {
	a := NewSequence()
	a.SetFormalParams([]*lg.Const{lg.NewConst("x", lg.Boolean)})
	a.SetFormalReturns([]*lg.Const{lg.NewConst("r", lg.Boolean)})
	s := ActionDefToStr("myact", a)
	if !strings.Contains(s, "action myact") {
		t.Errorf("ActionDefToStr missing 'action myact': %q", s)
	}
	if !strings.Contains(s, "returns") {
		t.Errorf("ActionDefToStr missing 'returns': %q", s)
	}
}

func TestApplyMixin(t *testing.T) {
	a1 := NewAssumeAction(mkConst("pre"))
	a1.SetLineno(ast.Location{Line: 1})
	a1.SetFormalParams([]*lg.Const{mkConst("p")})
	a1.SetFormalReturns(nil)

	a2 := NewSequence()
	a2.SetFormalParams([]*lg.Const{mkConst("p")})
	a2.SetFormalReturns(nil)

	// After mixin: a1 appended after a2
	result := ApplyMixin(a1, a2, true)
	if result.GetLineno().Line != 1 {
		t.Errorf("ApplyMixin should use action1 lineno, got %d", result.GetLineno().Line)
	}
}

// --- Annotations ---

func TestEmptyAnnotation(t *testing.T) {
	a := EmptyAnnotation{}
	if a.String() != "()" {
		t.Errorf("EmptyAnnotation.String() = %q", a.String())
	}
}

func TestConjAnnotation(t *testing.T) {
	a := EmptyAnnotation{}
	c := a.Conj(EmptyAnnotation{})
	if !strings.Contains(c.String(), "And") {
		t.Errorf("ConjAnnotation.String() = %q", c.String())
	}
}

func TestComposeAnnotation(t *testing.T) {
	a := EmptyAnnotation{}
	c := a.Compose(EmptyAnnotation{})
	if !strings.Contains(c.String(), "Compose") {
		t.Errorf("ComposeAnnotation.String() = %q", c.String())
	}
}

func TestRenameAnnotation(t *testing.T) {
	a := EmptyAnnotation{}
	xSym := lg.NewConst("x", lg.Boolean)
	ySym := lg.NewConst("y", lg.Boolean)
	r := a.Rename(map[lg.NodeKey]lg.Expr{lg.Key(xSym): ySym})
	if !strings.Contains(r.String(), "Rename") {
		t.Errorf("RenameAnnotation.String() = %q", r.String())
	}
}

func TestRenameAnnotationEmpty(t *testing.T) {
	a := EmptyAnnotation{}
	r := a.Rename(map[lg.NodeKey]lg.Expr{})
	// Empty map should return self.
	if _, ok := r.(EmptyAnnotation); !ok {
		t.Errorf("Rename with empty map should return self, got %T", r)
	}
}

func TestIteAnnotation(t *testing.T) {
	a := EmptyAnnotation{}
	i := a.Ite(lg.NewConst("cond", lg.Boolean), EmptyAnnotation{})
	if !strings.Contains(i.String(), "Ite") {
		t.Errorf("IteAnnotation.String() = %q", i.String())
	}
}

func TestComposeAnnotationWithLineno(t *testing.T) {
	loc := ast.Location{Filename: "test.ivy", Line: 10}
	c := &ComposeAnnotation{
		Args:   []Annotation{EmptyAnnotation{}, EmptyAnnotation{}},
		Lineno: &loc,
	}
	if !strings.Contains(c.String(), "test.ivy: line 10: ") {
		t.Errorf("ComposeAnnotation with lineno: %q", c.String())
	}
}

// --- Fuzz ---

func FuzzActionClone(f *testing.F) {
	f.Add("assume", "p")
	f.Add("assert", "q")
	f.Add("assign", "x")
	f.Add("havoc", "h")
	f.Add("sequence", "a")
	f.Add("call", "myaction")

	f.Fuzz(func(t *testing.T, actionType, argName string) {
		if argName == "" {
			return
		}
		c := mkConst(argName)
		var a Action
		switch actionType {
		case "assume":
			a = NewAssumeAction(c)
		case "assert":
			a = NewAssertAction(c)
		case "assign":
			a = NewAssignAction(c, mkConst("rhs"))
		case "havoc":
			a = NewHavocAction(c)
		case "sequence":
			a = NewSequence(c)
		case "call":
			a = NewCallActionOn(NewActionsConfig(), c)
		case "set":
			a = NewSetAction(c)
		case "crash":
			a = NewCrashAction(c)
		default:
			a = NewAssumeAction(c)
		}

		// Set some formals.
		a.SetFormalParams([]*lg.Const{mkConst("fp")})
		a.SetFormalReturns([]*lg.Const{mkConst("fr")})
		a.SetLineno(ast.Location{Filename: "fuzz.ivy", Line: 1})

		// Clone should not panic.
		cloned := a.ActionClone(a.ActionArgs())
		if cloned == nil {
			t.Error("Clone returned nil")
		}
		if cloned.Name() != a.Name() {
			t.Errorf("Clone changed Name: %q -> %q", a.Name(), cloned.Name())
		}

		// String should not panic.
		_ = a.String()
		_ = cloned.String()

		// IterCalls should not panic.
		_ = a.IterCalls()

		// IterSubactions should include self.
		subs := a.IterSubactions()
		if len(subs) == 0 {
			t.Error("IterSubactions should include self")
		}
	})
}

func FuzzAnnotation(f *testing.F) {
	f.Add("conj", "x", "y")
	f.Add("compose", "a", "b")
	f.Add("rename", "old", "new")
	f.Add("ite", "cond", "val")

	f.Fuzz(func(t *testing.T, op, s1, s2 string) {
		a := EmptyAnnotation{}
		b := EmptyAnnotation{}

		var result Annotation
		switch op {
		case "conj":
			result = a.Conj(b)
		case "compose":
			result = a.Compose(b)
		case "rename":
			keySym := lg.NewConst(s1, lg.Boolean)
			valSym := lg.NewConst(s2, lg.Boolean)
			result = a.Rename(map[lg.NodeKey]lg.Expr{lg.Key(keySym): valSym})
		case "ite":
			result = a.Ite(lg.NewConst(s1, lg.Boolean), b)
		default:
			result = a.Conj(b)
		}

		// Should not panic.
		_ = result.String()
	})
}
