package ivy2go

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// --- M8: hash-thunk emission tests ----------------------------------

func TestMakeThunk_SingleVariableEmitsStruct(t *testing.T) {
	g := newExprGen(t, "")
	// Build `lambda(v: bool) -> not v`.
	v, err := goivy.NewVariable("V", goivy.Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body := &goivy.LogicNot{Body: v}
	w := newGoWriter(NewGoText())
	g.thunks = w
	expr, err := g.makeThunk(&w, []*goivy.LogicVariable{v}, body)
	if err != nil {
		t.Fatalf("makeThunk: %v", err)
	}
	// Construction expression returned to caller.
	if !strings.HasPrefix(expr, "newthunk_0(") {
		t.Errorf("thunk construction = %q, want prefix newthunk_0(", expr)
	}
	// Struct definition emitted into thunks stream.
	text := g.Ctx.Thunks.GetFile()
	if !strings.Contains(text, "type thunk_0 struct") {
		t.Errorf("thunk struct not emitted, got:\n%s", text)
	}
	if !strings.Contains(text, "memo map[bool]bool") {
		t.Errorf("thunk memo field missing or wrong type:\n%s", text)
	}
	if !strings.Contains(text, "func newthunk_0() *thunk_0") {
		t.Errorf("thunk constructor not emitted, got:\n%s", text)
	}
	if !strings.Contains(text, "func (t *thunk_0) get(k bool) bool") {
		t.Errorf("thunk get method not emitted, got:\n%s", text)
	}
}

func TestMakeThunk_MemoizesIdenticalRequests(t *testing.T) {
	g := newExprGen(t, "")
	v, err := goivy.NewVariable("V", goivy.Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body := &goivy.LogicNot{Body: v}
	w := newGoWriter(NewGoText())
	g.thunks = w
	expr1, err := g.makeThunk(&w, []*goivy.LogicVariable{v}, body)
	if err != nil {
		t.Fatalf("makeThunk #1: %v", err)
	}
	expr2, err := g.makeThunk(&w, []*goivy.LogicVariable{v}, body)
	if err != nil {
		t.Fatalf("makeThunk #2: %v", err)
	}
	if expr1 != expr2 {
		t.Errorf("identical thunks should share name; got %q vs %q", expr1, expr2)
	}
	// Only one struct definition should appear.
	text := g.Ctx.Thunks.GetFile()
	if got := strings.Count(text, "type thunk_0 struct"); got != 1 {
		t.Errorf("thunk struct definition count = %d, want 1", got)
	}
	// And no second-counter struct should have leaked.
	if strings.Contains(text, "type thunk_1 struct") {
		t.Errorf("identical thunks should not produce thunk_1, got:\n%s", text)
	}
}

func TestNextThunkName_Increments(t *testing.T) {
	g := &Generator{}
	if n := g.nextThunkName(); n != "thunk_0" {
		t.Errorf("first thunk = %q, want thunk_0", n)
	}
	if n := g.nextThunkName(); n != "thunk_1" {
		t.Errorf("second thunk = %q, want thunk_1", n)
	}
}

func TestThunkEnvSymbols_FiltersLoopVarsAndEnums(t *testing.T) {
	g := newExprGen(t, `
type color = {red, green, blue}
relation pick(C: color)
function rank(C: color) : color
`)
	// expr: rank(C) where C is a loop var (uppercase per Ivy convention).
	c, err := goivy.NewVariable("C", g.Mod.Sig.Sorts.Get("color"))
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	rankSort, _ := goivy.NewFunctionSort(g.Mod.Sig.Sorts.Get("color"), g.Mod.Sig.Sorts.Get("color"))
	rank := goivy.NewConst("rank", rankSort)
	apply, err := goivy.NewApply(rank, c)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	env := g.thunkEnvSymbols([]*goivy.LogicVariable{c}, apply)
	if len(env) != 1 || env[0].Name != "rank" {
		t.Errorf("expected env = [rank], got %v", envNames(env))
	}
}

func envNames(env []*goivy.Const) []string {
	out := make([]string, len(env))
	for i, c := range env {
		out[i] = c.Name
	}
	return out
}

func TestThunkStruct_GofmtClean(t *testing.T) {
	// Force a real Generate so finalize runs format.Source on the
	// thunks stream alongside everything else.
	g := newExprGen(t, "")
	v, err := goivy.NewVariable("V", goivy.Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body := &goivy.LogicNot{Body: v}
	w := newGoWriter(NewGoText())
	g.thunks = w
	if _, err := g.makeThunk(&w, []*goivy.LogicVariable{v}, body); err != nil {
		t.Fatalf("makeThunk: %v", err)
	}
	// Compose a stand-alone file and assert gofmt-clean.
	text := "package p\n\n" + g.Ctx.Thunks.GetFile()
	assertGoSourceGofmt(t, "thunk.go", text)
}
