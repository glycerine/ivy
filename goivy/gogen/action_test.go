package gogen

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/actions"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// Helper to create a simple lg.Const node.
func testConst(name string, s lg.Sort) *lg.Const {
	return lg.NewConst(name, s)
}

// Helper to create a simple lg.Variable node.
func testVar(name string, s lg.Sort) *lg.Variable {
	v, _ := lg.NewVariable(name, s)
	return v
}

func emitActionToString(act actions.Action) string {
	w := NewCodeWriter()
	e := NewActionEmitter(nil, w)
	e.EmitAction(act)
	return w.String()
}

// --- AssignAction tests ---

func TestEmitAssign_Simple(t *testing.T) {
	lhs := testConst("x", lg.Boolean)
	rhs := testConst("y", lg.Boolean)
	act := actions.NewAssignAction(lhs, rhs)
	out := emitActionToString(act)
	if !strings.Contains(out, "x = y") {
		t.Errorf("expected assignment, got: %s", out)
	}
}

func TestEmitAssign_DottedName(t *testing.T) {
	lhs := testConst("node.link", lg.Boolean)
	rhs := testConst("true_val", lg.Boolean)
	act := actions.NewAssignAction(lhs, rhs)
	out := emitActionToString(act)
	if !strings.Contains(out, "node_link") {
		t.Errorf("expected dotted name converted, got: %s", out)
	}
}

// --- Sequence tests ---

func TestEmitSequence_Empty(t *testing.T) {
	act := actions.NewSequence()
	out := emitActionToString(act)
	// Empty sequence should produce no output.
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty output for empty sequence, got: %q", out)
	}
}

func TestEmitSequence_Multiple(t *testing.T) {
	a1 := actions.NewAssignAction(
		testConst("x", lg.Boolean),
		testConst("y", lg.Boolean),
	)
	a2 := actions.NewAssignAction(
		testConst("a", lg.Boolean),
		testConst("b", lg.Boolean),
	)
	seq := actions.NewSequence(a1, a2)
	out := emitActionToString(seq)
	if !strings.Contains(out, "x = y") {
		t.Errorf("missing first assignment: %s", out)
	}
	if !strings.Contains(out, "a = b") {
		t.Errorf("missing second assignment: %s", out)
	}
}

// --- IfAction tests ---

func TestEmitIf_NoElse(t *testing.T) {
	cond := testConst("c", lg.Boolean)
	body := actions.NewAssignAction(
		testConst("x", lg.Boolean),
		testConst("y", lg.Boolean),
	)
	act := actions.NewIfAction(cond, body)
	out := emitActionToString(act)
	if !strings.Contains(out, "if c {") {
		t.Errorf("expected if header, got: %s", out)
	}
	if strings.Contains(out, "else") {
		t.Errorf("unexpected else: %s", out)
	}
}

func TestEmitIf_WithElse(t *testing.T) {
	cond := testConst("c", lg.Boolean)
	thenBody := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	elseBody := actions.NewAssignAction(testConst("a", lg.Boolean), testConst("b", lg.Boolean))
	act := actions.NewIfAction(cond, thenBody, elseBody)
	out := emitActionToString(act)
	if !strings.Contains(out, "if c {") {
		t.Errorf("expected if header, got: %s", out)
	}
	if !strings.Contains(out, "else {") {
		t.Errorf("expected else clause, got: %s", out)
	}
}

// --- WhileAction tests ---

func TestEmitWhile_Simple(t *testing.T) {
	cond := testConst("running", lg.Boolean)
	body := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	act := actions.NewWhileAction(cond, body)
	out := emitActionToString(act)
	if !strings.Contains(out, "for running {") {
		t.Errorf("expected for loop, got: %s", out)
	}
}

func TestEmitWhile_WithInvariant(t *testing.T) {
	cond := testConst("running", lg.Boolean)
	body := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	inv := testConst("safe", lg.Boolean)
	act := actions.NewWhileAction(cond, body, inv)
	out := emitActionToString(act)
	if !strings.Contains(out, "invariant 0 violated") {
		t.Errorf("expected invariant check, got: %s", out)
	}
}

// --- CallAction tests ---

func TestEmitCall(t *testing.T) {
	callee := testConst("send", lg.Boolean)
	act := actions.NewCallActionOn(actions.NewActionsConfig(), callee)
	out := emitActionToString(act)
	if !strings.Contains(out, "s.Send()") {
		t.Errorf("expected method call, got: %s", out)
	}
}

// --- AssertAction tests ---

func TestEmitAssert(t *testing.T) {
	cond := testConst("valid", lg.Boolean)
	act := actions.NewAssertAction(cond)
	out := emitActionToString(act)
	if !strings.Contains(out, "if !(valid) {") {
		t.Errorf("expected assert check, got: %s", out)
	}
	if !strings.Contains(out, "assertion failed") {
		t.Errorf("expected panic message, got: %s", out)
	}
}

// --- RequiresAction tests ---

func TestEmitRequire(t *testing.T) {
	cond := testConst("precond", lg.Boolean)
	act := actions.NewRequiresAction(cond)
	out := emitActionToString(act)
	if !strings.Contains(out, "precondition failed") {
		t.Errorf("expected precondition panic, got: %s", out)
	}
}

// --- EnsuresAction tests ---

func TestEmitEnsure(t *testing.T) {
	cond := testConst("postcond", lg.Boolean)
	act := actions.NewEnsuresAction(cond)
	out := emitActionToString(act)
	if !strings.Contains(out, "postcondition failed") {
		t.Errorf("expected postcondition panic, got: %s", out)
	}
}

// --- AssumeAction tests ---

func TestEmitAssume(t *testing.T) {
	cond := testConst("premise", lg.Boolean)
	act := actions.NewAssumeAction(cond)
	out := emitActionToString(act)
	if !strings.Contains(out, "// assume:") {
		t.Errorf("expected assume comment, got: %s", out)
	}
}

// --- HavocAction tests ---

func TestEmitHavoc(t *testing.T) {
	target := testConst("x", lg.Boolean)
	act := actions.NewHavocAction(target)
	out := emitActionToString(act)
	if !strings.Contains(out, "havoc") {
		t.Errorf("expected havoc comment, got: %s", out)
	}
	if !strings.Contains(out, "false") {
		t.Errorf("expected default value for bool, got: %s", out)
	}
}

// --- ChoiceAction tests ---

func TestEmitChoice_Single(t *testing.T) {
	body := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	act := actions.NewChoiceActionOn(actions.NewActionsConfig(), body)
	out := emitActionToString(act)
	// Single branch should not use switch.
	if strings.Contains(out, "switch") {
		t.Errorf("single branch should not use switch, got: %s", out)
	}
	if !strings.Contains(out, "x = y") {
		t.Errorf("expected assignment, got: %s", out)
	}
}

func TestEmitChoice_Multiple(t *testing.T) {
	b1 := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	b2 := actions.NewAssignAction(testConst("a", lg.Boolean), testConst("b", lg.Boolean))
	act := actions.NewChoiceActionOn(actions.NewActionsConfig(), b1, b2)
	out := emitActionToString(act)
	if !strings.Contains(out, "switch rand.Intn(2)") {
		t.Errorf("expected switch with rand, got: %s", out)
	}
	if !strings.Contains(out, "case 0:") {
		t.Errorf("expected case 0, got: %s", out)
	}
}

func TestEmitChoice_Empty(t *testing.T) {
	act := actions.NewChoiceActionOn(actions.NewActionsConfig())
	out := emitActionToString(act)
	if !strings.Contains(out, "empty choice") {
		t.Errorf("expected empty choice comment, got: %s", out)
	}
}

// --- LocalAction tests ---

func TestEmitLocal(t *testing.T) {
	local := testConst("tmp", lg.Boolean)
	body := actions.NewAssignAction(
		testConst("tmp", lg.Boolean),
		testConst("x", lg.Boolean),
	)
	act := actions.NewLocalActionOn(actions.NewActionsConfig(), "test", local, body)
	out := emitActionToString(act)
	if !strings.Contains(out, "var tmp bool") {
		t.Errorf("expected local var declaration, got: %s", out)
	}
}

// --- LetAction tests ---

func TestEmitLet(t *testing.T) {
	binding := testConst("val", lg.Boolean)
	body := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("val", lg.Boolean))
	act := actions.NewLetAction(binding, body)
	out := emitActionToString(act)
	if !strings.Contains(out, "val := val") {
		t.Errorf("expected let binding, got: %s", out)
	}
}

// --- NativeAction tests ---

func TestEmitNative(t *testing.T) {
	code := testConst("some_native_code", lg.Boolean)
	act := actions.NewNativeAction(code)
	out := emitActionToString(act)
	if !strings.Contains(out, "// native:") {
		t.Errorf("expected native comment, got: %s", out)
	}
}

// --- CrashAction tests ---

func TestEmitCrash(t *testing.T) {
	target := testConst("node", lg.Boolean)
	act := actions.NewCrashAction(target)
	out := emitActionToString(act)
	if !strings.Contains(out, `panic("crash")`) {
		t.Errorf("expected crash panic, got: %s", out)
	}
}

// --- SetAction tests ---

func TestEmitSet(t *testing.T) {
	lit := testConst("link", lg.Boolean)
	act := actions.NewSetAction(lit)
	out := emitActionToString(act)
	if !strings.Contains(out, "// set:") {
		t.Errorf("expected set comment, got: %s", out)
	}
}

// --- EnvAction tests ---

func TestEmitEnv(t *testing.T) {
	b1 := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	act := actions.NewEnvActionOn(actions.NewActionsConfig(), b1)
	out := emitActionToString(act)
	if !strings.Contains(out, "x = y") {
		t.Errorf("expected assignment from env, got: %s", out)
	}
}

// --- BindOldsAction tests ---

func TestEmitBindOlds(t *testing.T) {
	inner := actions.NewAssignAction(testConst("x", lg.Boolean), testConst("y", lg.Boolean))
	act := actions.NewBindOldsAction(inner)
	out := emitActionToString(act)
	if !strings.Contains(out, "x = y") {
		t.Errorf("expected inner action emitted, got: %s", out)
	}
}

// --- ReturnAction tests ---

func TestEmitReturn(t *testing.T) {
	act := actions.NewReturnAction()
	out := emitActionToString(act)
	if !strings.Contains(out, "return") {
		t.Errorf("expected return, got: %s", out)
	}
}

// --- IgnoreAction tests ---

func TestEmitIgnore(t *testing.T) {
	act := actions.NewIgnoreAction()
	out := emitActionToString(act)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected no output for ignore, got: %q", out)
	}
}

// --- Nil action ---

func TestEmitNilAction(t *testing.T) {
	out := emitActionToString(nil)
	if !strings.Contains(out, "nil action") {
		t.Errorf("expected nil action comment, got: %s", out)
	}
}

// --- ExprToGo tests ---

func TestExprToGo_Const(t *testing.T) {
	c := testConst("myVar", lg.Boolean)
	got := ExprToGo(c)
	if got != "myVar" {
		t.Errorf("expected myVar, got: %s", got)
	}
}

func TestExprToGo_Eq(t *testing.T) {
	eq, _ := lg.NewEq(testConst("a", lg.Boolean), testConst("b", lg.Boolean))
	got := ExprToGo(eq)
	if !strings.Contains(got, "==") {
		t.Errorf("expected ==, got: %s", got)
	}
}

func TestExprToGo_Not(t *testing.T) {
	not, _ := lg.NewNot(testConst("a", lg.Boolean))
	got := ExprToGo(not)
	if !strings.Contains(got, "!(") {
		t.Errorf("expected negation, got: %s", got)
	}
}

func TestExprToGo_And(t *testing.T) {
	and, _ := lg.NewAnd(testConst("a", lg.Boolean), testConst("b", lg.Boolean))
	got := ExprToGo(and)
	if !strings.Contains(got, "&&") {
		t.Errorf("expected &&, got: %s", got)
	}
}

func TestExprToGo_Or(t *testing.T) {
	or, _ := lg.NewOr(testConst("a", lg.Boolean), testConst("b", lg.Boolean))
	got := ExprToGo(or)
	if !strings.Contains(got, "||") {
		t.Errorf("expected ||, got: %s", got)
	}
}

func TestExprToGo_Implies(t *testing.T) {
	imp, _ := lg.NewImplies(testConst("a", lg.Boolean), testConst("b", lg.Boolean))
	got := ExprToGo(imp)
	if !strings.Contains(got, "||") {
		t.Errorf("expected implication encoding, got: %s", got)
	}
}

func TestExprToGo_Nil(t *testing.T) {
	got := ExprToGo(nil)
	if got != "nil" {
		t.Errorf("expected nil, got: %s", got)
	}
}

// --- GoIdentifier tests ---

func TestGoIdentifier_DottedName(t *testing.T) {
	got := GoIdentifier("node.link")
	if got != "node_link" {
		t.Errorf("expected node_link, got: %s", got)
	}
}

func TestGoIdentifier_Empty(t *testing.T) {
	got := GoIdentifier("")
	if got != "_" {
		t.Errorf("expected _, got: %s", got)
	}
}

func TestGoIdentifier_NumericStart(t *testing.T) {
	got := GoIdentifier("3way")
	if got != "_3way" {
		t.Errorf("expected _3way, got: %s", got)
	}
}

func TestGoExportedIdentifier(t *testing.T) {
	got := GoExportedIdentifier("send")
	if got != "Send" {
		t.Errorf("expected Send, got: %s", got)
	}
}

// --- goTypeForSort tests ---

func TestGoTypeForSort_Bool(t *testing.T) {
	got := goTypeForSort(lg.Boolean)
	if got != "bool" {
		t.Errorf("expected bool, got: %s", got)
	}
}

func TestGoTypeForSort_Nil(t *testing.T) {
	got := goTypeForSort(nil)
	if got != "interface{}" {
		t.Errorf("expected interface{}, got: %s", got)
	}
}

// --- Fuzz test ---

func FuzzGoIdentifier(f *testing.F) {
	f.Add("hello")
	f.Add("node.link")
	f.Add("a:b:c")
	f.Add("")
	f.Add("123")
	f.Add("__x__")
	f.Fuzz(func(t *testing.T, input string) {
		result := GoIdentifier(input)
		if len(result) == 0 {
			t.Error("GoIdentifier returned empty string")
		}
		// Should not contain dots or colons.
		if strings.ContainsAny(result, ".:") {
			t.Errorf("GoIdentifier result contains dots or colons: %q", result)
		}
	})
}
