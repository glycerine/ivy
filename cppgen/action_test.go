package cppgen

import (
	"strings"
	"testing"

	lg "github.com/glycerine/goivy/logic"
)

// ---------------------------------------------------------------------------
// EmitAssignSimple tests
// ---------------------------------------------------------------------------

func TestEmitAssignSimple(t *testing.T) {
	resetState()
	var buf CodeText
	EmitAssignSimple(&buf, "x", "42")
	got := buf.String()
	if !strings.Contains(got, "x = 42;") {
		t.Errorf("EmitAssignSimple = %q, want x = 42;", got)
	}
}

// ---------------------------------------------------------------------------
// EmitAssign with free vars tests
// ---------------------------------------------------------------------------

func TestEmitAssignWithFreeVars(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	vs := []*lg.Var{{Name: "i", VSort: &lg.EnumeratedSort{Name: "idx", Extension: []string{"0", "1", "2", "3", "4"}}}}
	EmitAssign(ctx, &buf, "arr", "val", vs)
	got := buf.String()
	if !strings.Contains(got, "for (int i") {
		t.Errorf("EmitAssign missing loop: %q", got)
	}
	if !strings.Contains(got, "arr[i] = val[i]") {
		t.Errorf("EmitAssign missing indexed assignment: %q", got)
	}
}

func TestEmitAssignNoFreeVars(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	EmitAssign(ctx, &buf, "x", "1", nil)
	got := buf.String()
	if !strings.Contains(got, "x = 1;") {
		t.Errorf("EmitAssign scalar = %q, want x = 1;", got)
	}
}

// ---------------------------------------------------------------------------
// EmitHavoc tests
// ---------------------------------------------------------------------------

func TestEmitHavoc(t *testing.T) {
	resetState()
	var buf CodeText
	sym := mkConst("x", &lg.UninterpretedSort{Name: "int"})
	EmitHavoc(&buf, sym)
	got := buf.String()
	if !strings.Contains(got, "havoc x") {
		t.Errorf("EmitHavoc = %q, want comment about havoc", got)
	}
}

// ---------------------------------------------------------------------------
// EmitSequence tests
// ---------------------------------------------------------------------------

func TestEmitSequence(t *testing.T) {
	resetState()
	var buf CodeText
	EmitSequence(&buf, []string{"a = 1;", "b = 2;"})
	got := buf.String()
	if !strings.Contains(got, "{") || !strings.Contains(got, "}") {
		t.Errorf("EmitSequence missing braces: %q", got)
	}
	if !strings.Contains(got, "a = 1;") || !strings.Contains(got, "b = 2;") {
		t.Errorf("EmitSequence missing actions: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitAssert / EmitAssume tests
// ---------------------------------------------------------------------------

func TestEmitAssert(t *testing.T) {
	resetState()
	var buf CodeText
	EmitAssert(&buf, "cond", "file.ivy:10")
	got := buf.String()
	if !strings.Contains(got, "ivy_assert(cond,") {
		t.Errorf("EmitAssert = %q, missing ivy_assert", got)
	}
}

func TestEmitAssume(t *testing.T) {
	resetState()
	var buf CodeText
	EmitAssume(&buf, "cond", "file.ivy:20")
	got := buf.String()
	if !strings.Contains(got, "ivy_assume(cond,") {
		t.Errorf("EmitAssume = %q, missing ivy_assume", got)
	}
}

// ---------------------------------------------------------------------------
// EmitCall tests
// ---------------------------------------------------------------------------

func TestEmitCallNoReturn(t *testing.T) {
	resetState()
	var buf CodeText
	args := []CallArg{
		{Code: "a", FormalSort: &lg.UninterpretedSort{Name: "int"}, ActualSort: &lg.UninterpretedSort{Name: "int"}},
		{Code: "b", FormalSort: &lg.UninterpretedSort{Name: "int"}, ActualSort: &lg.UninterpretedSort{Name: "int"}},
	}
	EmitCall(&buf, "do.something", args, "", false)
	got := buf.String()
	if !strings.Contains(got, "do__something(a, b)") {
		t.Errorf("EmitCall = %q, unexpected", got)
	}
}

func TestEmitCallWithReturn(t *testing.T) {
	resetState()
	var buf CodeText
	EmitCall(&buf, "compute", nil, "result", true)
	got := buf.String()
	if !strings.Contains(got, "result = compute()") {
		t.Errorf("EmitCall = %q, missing return assignment", got)
	}
}

// ---------------------------------------------------------------------------
// LocalStart / LocalEnd tests
// ---------------------------------------------------------------------------

func TestLocalStartEnd(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	params := []*lg.Const{mkConst("tmp", &lg.UninterpretedSort{Name: "int"})}
	LocalStart(ctx, &buf, params, -1)
	codeLine(&buf, "tmp = 0")
	LocalEnd(&buf)
	got := buf.String()
	if !strings.Contains(got, "{\n") {
		t.Errorf("LocalStart missing open brace: %q", got)
	}
	if !strings.Contains(got, "}\n") {
		t.Errorf("LocalEnd missing close brace: %q", got)
	}
}

func TestLocalStartWithNondet(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	params := []*lg.Const{mkConst("x", &lg.UninterpretedSort{Name: "int"})}
	LocalStart(ctx, &buf, params, 42)
	LocalEnd(&buf)
	got := buf.String()
	if !strings.Contains(got, "___ivy_choose") {
		t.Errorf("LocalStart nondet missing choose: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitIf tests
// ---------------------------------------------------------------------------

func TestEmitIfOnly(t *testing.T) {
	resetState()
	var buf CodeText
	EmitIf(&buf, "x > 0", "    y = 1;\n", "")
	got := buf.String()
	if !strings.Contains(got, "if(x > 0)") {
		t.Errorf("EmitIf missing condition: %q", got)
	}
	if strings.Contains(got, "else") {
		t.Errorf("EmitIf unexpected else: %q", got)
	}
}

func TestEmitIfElse(t *testing.T) {
	resetState()
	var buf CodeText
	EmitIf(&buf, "x > 0", "    y = 1;\n", "    y = 0;\n")
	got := buf.String()
	if !strings.Contains(got, "else") {
		t.Errorf("EmitIfElse missing else: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitWhile tests
// ---------------------------------------------------------------------------

func TestEmitWhileSimple(t *testing.T) {
	resetState()
	var buf CodeText
	EmitWhile(&buf, "i < n", "", "    i++;\n")
	got := buf.String()
	if !strings.Contains(got, "while(i < n)") {
		t.Errorf("EmitWhile = %q, missing while", got)
	}
}

func TestEmitWhileWithPreamble(t *testing.T) {
	resetState()
	var buf CodeText
	EmitWhile(&buf, "cond", "    compute_cond();\n", "    body();\n")
	got := buf.String()
	if !strings.Contains(got, "while(true)") {
		t.Errorf("EmitWhile preamble = %q, missing while(true)", got)
	}
	if !strings.Contains(got, "break") {
		t.Errorf("EmitWhile preamble = %q, missing break", got)
	}
}

// ---------------------------------------------------------------------------
// EmitChoice tests
// ---------------------------------------------------------------------------

func TestEmitChoiceSingle(t *testing.T) {
	resetState()
	var buf CodeText
	EmitChoice(&buf, []string{"    a = 1;\n"}, 0)
	got := buf.String()
	if strings.Contains(got, "if(") {
		t.Errorf("EmitChoice single branch has if: %q", got)
	}
}

func TestEmitChoiceMultiple(t *testing.T) {
	resetState()
	var buf CodeText
	EmitChoice(&buf, []string{"    a();\n", "    b();\n", "    c();\n"}, 99)
	got := buf.String()
	if !strings.Contains(got, "___ivy_choose") {
		t.Errorf("EmitChoice multi missing choose: %q", got)
	}
	if !strings.Contains(got, "else") {
		t.Errorf("EmitChoice multi missing else: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitCrash / EmitDebug tests
// ---------------------------------------------------------------------------

func TestEmitCrash(t *testing.T) {
	resetState()
	var buf CodeText
	EmitCrash(&buf)
	if buf.String() != "" {
		t.Errorf("EmitCrash should emit nothing, got %q", buf.String())
	}
}

func TestEmitDebug(t *testing.T) {
	resetState()
	var buf CodeText
	EmitDebug(&buf, "step", []DebugField{
		{Name: "x", Code: "x"},
	})
	got := buf.String()
	if !strings.Contains(got, "step") {
		t.Errorf("EmitDebug missing event name: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitNativeAction tests
// ---------------------------------------------------------------------------

func TestEmitNativeAction(t *testing.T) {
	resetState()
	var buf CodeText
	EmitNativeAction(&buf, "printf(\"hello\");")
	got := buf.String()
	if !strings.Contains(got, "printf") {
		t.Errorf("EmitNativeAction = %q, missing printf", got)
	}
}

// ---------------------------------------------------------------------------
// EmitQuant tests
// ---------------------------------------------------------------------------

func TestEmitQuantEmpty(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	EmitQuant(ctx, &buf, nil, "body_expr", false)
	if !strings.Contains(buf.String(), "body_expr") {
		t.Errorf("EmitQuant empty vars = %q, missing body", buf.String())
	}
}

func TestEmitQuantForall(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	vs := []*lg.Var{{Name: "i", VSort: &lg.EnumeratedSort{Name: "node", Extension: []string{"n0", "n1", "n2"}}}}
	EmitQuant(ctx, &buf, vs, "pred(i)", false)
	got := buf.String()
	if !strings.Contains(got, "for (") {
		t.Errorf("EmitQuant forall missing for: %q", got)
	}
	if !strings.Contains(got, "= 1") {
		t.Errorf("EmitQuant forall should init to 1: %q", got)
	}
}

func TestEmitQuantExists(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	vs := []*lg.Var{{Name: "i", VSort: &lg.EnumeratedSort{Name: "node", Extension: []string{"n0", "n1", "n2"}}}}
	EmitQuant(ctx, &buf, vs, "pred(i)", true)
	got := buf.String()
	if !strings.Contains(got, "= 0") {
		t.Errorf("EmitQuant exists should init to 0: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EmitSome tests
// ---------------------------------------------------------------------------

func TestEmitSome(t *testing.T) {
	resetState()
	ctx := testCtx()
	var buf CodeText
	vs := []*lg.Var{{Name: "i", VSort: &lg.EnumeratedSort{Name: "node", Extension: []string{"n0", "n1", "n2"}}}}
	EmitSome(ctx, &buf, vs, "check(i)", "", "result")
	got := buf.String()
	if !strings.Contains(got, "for (int i") {
		t.Errorf("EmitSome missing loop: %q", got)
	}
	if !strings.Contains(got, "if(check(i))") {
		t.Errorf("EmitSome missing if: %q", got)
	}
}

// ---------------------------------------------------------------------------
// GetBounds tests
// ---------------------------------------------------------------------------

func TestGetBoundsFinite(t *testing.T) {
	ctx := testCtx()
	s := &lg.EnumeratedSort{Name: "color", Extension: []string{"r", "g", "b"}}
	bds, err := GetBounds(ctx, s)
	if err != nil {
		t.Fatalf("GetBounds error: %v", err)
	}
	if bds[0] != "0" || bds[1] != "3" {
		t.Errorf("GetBounds = %v, want [0, 3]", bds)
	}
}
