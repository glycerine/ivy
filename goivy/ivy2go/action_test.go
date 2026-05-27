package ivy2go

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// --- M4: action emission tests ---------------------------------------
//
// These tests exercise emitAction for individual statement shapes by
// constructing the action AST in-process. Where Ivy source is the
// shorter path (e.g. for fully-typed assignments), we compile a small
// module and reach into mod.Actions.

func newActionGen(t *testing.T, src string) *Generator {
	return newExprGen(t, src)
}

// --- Sequence + AssertLike -------------------------------------------

func TestEmitAction_AssertEmitsIvyAssertCall(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	assertion := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	a := &goivy.LogicAssertAction{Formula: assertion}
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "ivyAssert(true,") {
		t.Errorf("assert did not emit ivyAssert call:\n%s", got)
	}
}

func TestEmitAction_AssumeEmitsIvyAssumeCall(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	a := &goivy.LogicAssumeAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "ivyAssume(true,") {
		t.Errorf("assume did not emit ivyAssume call:\n%s", got)
	}
}

func TestEmitAction_SequenceEmitsBracedBlock(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	a := &goivy.LogicSequence{Elems: []goivy.Expr{
		&goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}},
		&goivy.LogicAssumeAction{Formula: &goivy.Const{Name: "false", CSort: goivy.Boolean}},
	}}
	g.emitAction(&w, a)
	got := w.String()
	// braced wrapper + both nested calls
	if !strings.HasPrefix(strings.TrimSpace(got), "{") {
		t.Errorf("sequence missing opening {:\n%s", got)
	}
	if !strings.Contains(got, "ivyAssert(true,") {
		t.Errorf("sequence body missing assert:\n%s", got)
	}
	if !strings.Contains(got, "ivyAssume(false,") {
		t.Errorf("sequence body missing assume:\n%s", got)
	}
}

// --- If --------------------------------------------------------------

func TestEmitAction_IfEmitsGoStyleHeader(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	cond := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	then := &goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	a := &goivy.LogicIfAction{Cond: cond, ThenBody: then}
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "if true {") {
		t.Errorf("if header should be Go-style 'if true {':\n%s", got)
	}
}

func TestEmitAction_IfElseChain(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	cond := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	then := &goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	els := &goivy.LogicAssumeAction{Formula: &goivy.Const{Name: "false", CSort: goivy.Boolean}}
	a := &goivy.LogicIfAction{Cond: cond, ThenBody: then, ElseBody: els}
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "} else {") {
		t.Errorf("if-else missing 'else' brace:\n%s", got)
	}
}

// --- While -----------------------------------------------------------

func TestEmitAction_WhileLowersToFor(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	cond := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	body := &goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	a := &goivy.LogicWhileAction{Cond: cond, Body: body}
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "for true {") {
		t.Errorf("while should lower to 'for true {':\n%s", got)
	}
}

// --- Return ----------------------------------------------------------

func TestEmitAction_ReturnEmitsBareReturn(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	g.emitAction(&w, &goivy.ReturnAction{})
	got := strings.TrimSpace(w.String())
	if got != "return" {
		t.Errorf("return = %q, want %q", got, "return")
	}
}

func TestEmitAction_ReturnUsesNamedReturnWhenSingle(t *testing.T) {
	g := newActionGen(t, "")
	g.currentReturns = []*goivy.Const{{Name: "result", CSort: goivy.Boolean}}
	w := newGoWriter(NewGoText())
	g.emitAction(&w, &goivy.ReturnAction{})
	got := strings.TrimSpace(w.String())
	if got != "return result" {
		t.Errorf("return = %q, want %q", got, "return result")
	}
}

func TestEmitPrintExpr_BoolVariableUsesFirstSentinel(t *testing.T) {
	g := newActionGen(t, "")
	b, err := goivy.NewVariable("B", goivy.Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	w := newGoWriter(NewGoText())
	g.emitPrintExpr(&w, b)
	got := w.String()
	if strings.Contains(got, "B != 0") {
		t.Fatalf("debug print over bool must not compare the loop value to zero:\n%s", got)
	}
	if !strings.Contains(got, "if !__temp__0") || !strings.Contains(got, "__temp__0 = false") {
		t.Fatalf("debug print should use a per-loop first sentinel, got:\n%s", got)
	}
}

func TestEmitPrintExpr_NonzeroRangeVariableUsesFirstSentinel(t *testing.T) {
	g := newActionGen(t, "")
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "5"}, Ub: goivy.NumeralBound{Value: "7"}}
	i, err := goivy.NewVariable("I", rng)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	w := newGoWriter(NewGoText())
	g.emitPrintExpr(&w, i)
	got := w.String()
	if strings.Contains(got, "I != 0") {
		t.Fatalf("debug print over nonzero range must not use zero as comma sentinel:\n%s", got)
	}
	if !strings.Contains(got, "if !__temp__0") || !strings.Contains(got, "__temp__0 = false") {
		t.Fatalf("debug print should use a per-loop first sentinel, got:\n%s", got)
	}
}

func TestSmoke_BuildEmittedDebugPrintBoolFreeVariable(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
relation seen(B: bool)
action report = {}
export report
`)
	seenEntry, ok := mod.Sig.Symbols.Get2("seen")
	if !ok {
		t.Fatal("missing seen symbol")
	}
	b, err := goivy.NewVariable("B", goivy.Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	seen := goivy.NewConst("seen", seenEntry.Sort)
	seenB, err := goivy.NewApply(seen, b)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	dbg := goivy.NewDebugAction(goivy.NewConst(`"myvar"`, goivy.TopS), seenB)
	dbg.WithNames = []string{"values"}
	mod.SetAction("report", dbg)
	buildEmittedImplPackage(t, mod, "ivygo_debug_bool_free_var")
}

// --- Choice ----------------------------------------------------------

func TestEmitAction_ChoiceUsesIvyChooseSwitch(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	a := &goivy.LogicChoiceAction{Branches: []goivy.Expr{
		&goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}},
		&goivy.LogicAssumeAction{Formula: &goivy.Const{Name: "false", CSort: goivy.Boolean}},
	}}
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "switch ivyChoose(2)") {
		t.Errorf("choice should emit 'switch ivyChoose(2)':\n%s", got)
	}
	if !strings.Contains(got, "case 0:") {
		t.Errorf("choice missing 'case 0:':\n%s", got)
	}
	if !strings.Contains(got, "default:") {
		t.Errorf("choice missing default branch:\n%s", got)
	}
}

// --- Local -----------------------------------------------------------

func TestEmitAction_LocalDeclaresVar(t *testing.T) {
	g := newActionGen(t, "")
	w := newGoWriter(NewGoText())
	local := &goivy.Const{Name: "tmp", CSort: goivy.Boolean}
	body := &goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	a := &goivy.LogicLocalAction{Locals: []goivy.Expr{local}, Body: body}
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "var tmp bool") {
		t.Errorf("local missing 'var tmp bool':\n%s", got)
	}
}

// --- Assign ----------------------------------------------------------

func TestEmitAction_AssignScalarBoolean(t *testing.T) {
	// state symbol `flag : bool` so the LHS resolves through
	// goStorageAccess and produces `s.Flag`. We build the Const
	// directly because Sig.Symbols stores SymbolEntry entries, not
	// the Const expressions emitAction wants.
	g := newActionGen(t, `
relation flag
`)
	flag := goivy.NewConst("flag", goivy.Boolean)
	rhs := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	a := &goivy.LogicAssignAction{LHS: flag, RHS: rhs}
	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := strings.TrimSpace(w.String())
	if got != "s.Flag = true" {
		t.Errorf("assign emission = %q, want %q", got, "s.Flag = true")
	}
}

func TestEmitActions_RangeArithmeticClampFileIsValidGo(t *testing.T) {
	mod := compileIvySource(t, `
type idx = {5..7}
individual current : idx
action bump = {
	current := current + current
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	requireHasLineWithAllTerms(t, actions, "func() Idx")
	if strings.Contains(actions, "return if") {
		t.Fatalf("range arithmetic should not emit statement text in return expression:\n%s", actions)
	}
	for name, text := range out.Files {
		assertGoSourceGofmt(t, name, text)
	}
}

// --- emitActionMethods integration ----------------------------------

func TestEmitActions_EmptyModuleProducesNoActionsFile(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, has := out.Files["actions.go"]; has {
		t.Errorf("empty module should not produce actions.go, got:\n%s", out.Files["actions.go"])
	}
}

func TestEmitActions_BasicActionEmitsMethod(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["actions.go"]
	if text == "" {
		t.Fatalf("actions.go is empty:\n%v", out.Files)
	}
	requireHasLineWithAllTerms(t, text, "func", "(s *State)", "SetFlag")
}

func TestEmitActions_GeneratedFileIsGofmtClean(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for name, text := range out.Files {
		assertGoSourceGofmt(t, name, text)
	}
}

// --- linenoStr unit test --------------------------------------------

func TestLinenoStr(t *testing.T) {
	// Location.String() ends in ": " by default; linenoStr drops it.
	var loc goivy.Location
	loc.Filename = "file.ivy"
	loc.Line = 42
	got := linenoStr(loc)
	if strings.HasSuffix(got, ": ") {
		t.Errorf("linenoStr should drop trailing ': ', got %q", got)
	}
	if !strings.Contains(got, "file.ivy") || !strings.Contains(got, "42") {
		t.Errorf("linenoStr = %q, want filename:line", got)
	}
}
