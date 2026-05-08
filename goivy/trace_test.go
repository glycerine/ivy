package goivy

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

// --- helpers ---

func traceTestModule() *Module {
	return NewModule()
}

func traceTestClauses(fmlas ...Expr) *Clauses {
	if len(fmlas) == 0 {
		fmlas = []Expr{True}
	}
	return NewClauses(fmlas, nil, nil)
}

func makeEq(name string, val string) Expr {
	c := NewConst(name, Boolean)
	v := NewConst(val, Boolean)
	eq, _ := NewEq(c, v)
	return eq
}

type traceTestModel struct{}

func (traceTestModel) EvalToConstant(Expr) Expr { return True }
func (traceTestModel) Universes(bool) map[string][]Expr {
	return nil
}

type traceUniverseModel struct{}

func (traceUniverseModel) EvalToConstant(Expr) Expr { return True }
func (traceUniverseModel) Universes(bool) map[string][]Expr {
	sort := &UninterpretedSort{Name: "S"}
	return map[string][]Expr{
		"S": {
			NewConst("0:S", sort),
			NewConst("1:S", sort),
		},
	}
}

// --- TraceBase tests ---

func TestTraceNewTraceBase(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	if tb == nil {
		t.Fatal("NewTraceBase returned nil")
	}
	if tb.AnalysisGraph == nil {
		t.Fatal("AnalysisGraph is nil")
	}
	if tb.HiddenSymbols == nil {
		t.Fatal("HiddenSymbols function is nil")
	}
	if tb.HiddenSymbols("anything") {
		t.Error("default HiddenSymbols should return false")
	}
}

func TestTraceNewTraceBaseWithModule(t *testing.T) {
	mod := traceTestModule()
	tb := NewTraceBase(nil, mod)
	if tb.Domain != mod {
		t.Error("domain should match provided module")
	}
}

func TestTraceTraceBaseRename(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	m := map[string]string{"old_x": "x", "old_y": "y"}
	result := tb.Rename(m)
	if result != tb {
		t.Error("Rename should return same TraceBase")
	}
	if len(tb.Renaming) != 2 {
		t.Errorf("expected 2 entries, got %d", len(tb.Renaming))
	}
}

func TestTraceIsSkolem(t *testing.T) {
	tests := []struct {
		name   string
		expect bool
	}{
		{"x", false},
		{"foo__bar", true},
		{"__X", false},     // global skolem, uppercase
		{"__xbar", true},   // regular skolem, lowercase
		{"__Ytest", false}, // global skolem
		{"abc", false},
		{"a__b", true},
	}
	for _, tt := range tests {
		got := TraceIsSkolem(tt.name)
		if got != tt.expect {
			t.Errorf("TraceIsSkolem(%q) = %v, want %v", tt.name, got, tt.expect)
		}
	}
}

func TestTraceAddTraceState(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	tb.AddTraceState(nil)
	if len(tb.TraceStates) != 1 {
		t.Errorf("expected 1 trace state, got %d", len(tb.TraceStates))
	}
	if len(tb.States) != 1 {
		t.Errorf("expected 1 art state, got %d", len(tb.States))
	}
}

func TestTraceAddTraceStateWithAction(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	tb.AddTraceState(nil) // first state
	tb.LastAction = NewAssumeAction(True)
	tb.AddTraceState(nil) // second state linked by action
	if len(tb.TraceStates) != 2 {
		t.Errorf("expected 2 trace states, got %d", len(tb.TraceStates))
	}
	if tb.LastAction != nil {
		t.Error("LastAction should be nil after AddTraceState")
	}
}

func TestTraceTraceBaseString(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	s := tb.String()
	// Empty trace should produce empty string
	if s != "" {
		t.Errorf("empty trace String() should be empty, got %q", s)
	}
}

func TestTraceTraceBaseFail(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	origAction := NewAssertAction(True)
	tb.LastAction = origAction
	tb.Fail()
	fa, ok := tb.LastAction.(*FailAction)
	if !ok {
		t.Fatal("Fail should wrap action in FailAction")
	}
	if fa.Inner != origAction {
		t.Error("FailAction should wrap original action")
	}
}

// --- Pretty / label helpers ---

func TestTraceDetailedActionRenamingAndLinenoMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_actions as act, logic as lg
from ivy import ivy_module as im, ivy_utils as iu

im.module = im.Module()

class DummyTrace(ivy_trace.TraceBase):
    def get_universes(self):
        return None

old = lg.Const("old", lg.Boolean)
new = lg.Const("new", lg.Boolean)
action = act.AssignAction(old, lg.true)
action.lineno = iu.Location("sample.ivy", 7)

trace = DummyTrace()
trace.renaming = {old: new}
trace.add_state([])
trace.last_action = action
trace.add_state([])
print(json.dumps(str(trace)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python trace action-label oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python trace oracle %q: %v", out, err)
	}

	mod := NewModule()
	mod.Cfg.TraceDetailed = true
	old := NewConst("old", Boolean)
	action := NewAssignAction(old, True)
	action.SetLineno(Location{Filename: "sample.ivy", Line: 7})

	trace := NewTraceBase(nil, mod)
	trace.Renaming = map[string]string{"old": "new"}
	trace.AddTraceState(nil)
	trace.LastAction = action
	trace.AddTraceState(nil)

	if got := trace.String(); got != want {
		t.Fatalf("Go trace action label differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestTraceNonDetailedFailureSummaryMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_actions as act, logic as lg
from ivy import ivy_module as im, ivy_utils as iu, ivy_interp as itp

im.module = im.Module()
ivy_trace.option_detailed.set("false")

class DummyTrace(ivy_trace.TraceBase):
    def get_universes(self):
        return None

action = act.AssertAction(lg.false)
action.lineno = iu.Location("sample.ivy", 9)
trace = DummyTrace()
trace.add_state([])
trace.last_action = itp.fail_action(action)
trace.add_state([])
print(json.dumps(str(trace)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python non-detailed failure oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python non-detailed failure oracle %q: %v", out, err)
	}

	mod := NewModule()
	mod.Cfg.TraceDetailed = false
	action := NewAssertAction(False)
	action.SetLineno(Location{Filename: "sample.ivy", Line: 9})
	trace := NewTraceBase(nil, mod)
	trace.AddTraceState(nil)
	trace.LastAction = NewFailAction(action)
	trace.AddTraceState(nil)

	if got := trace.String(); got != want {
		t.Fatalf("Go non-detailed failure trace differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestTraceNonDetailedEnvSummaryMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_actions as act, logic as lg
from ivy import ivy_module as im

im.module = im.Module()
ivy_trace.option_detailed.set("false")

class DummyTrace(ivy_trace.TraceBase):
    def get_universes(self):
        return None

x = lg.Const("x", lg.Boolean)
branch = act.Sequence()
branch.label = "ext"
branch.formal_params = [x]
trace = DummyTrace()
trace.add_state([lg.Eq(x, lg.true)])
trace.last_action = act.EnvAction(branch)
trace.add_state([lg.Eq(x, lg.true)])
print(json.dumps(str(trace)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python non-detailed env oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python non-detailed env oracle %q: %v", out, err)
	}

	mod := NewModule()
	mod.Cfg.TraceDetailed = false
	x := NewConst("x", Boolean)
	branch := NewSequence()
	branch.SetLabel("ext")
	branch.SetFormalParams([]*Const{x})
	trace := NewTraceBase(nil, mod)
	trace.AddTraceState([]Expr{&Eq{T1: x, T2: True}})
	trace.LastAction = NewEnvActionOn(NewActionsConfig(), branch)
	trace.AddTraceState([]Expr{&Eq{T1: x, T2: True}})

	if got := trace.String(); got != want {
		t.Fatalf("Go non-detailed env trace differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestTraceNonDetailedCallSummaryMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_actions as act, logic as lg
from ivy import ivy_module as im, ivy_ast

im.module = im.Module()
ivy_trace.option_detailed.set("false")

class DummyTrace(ivy_trace.TraceBase):
    def get_universes(self):
        return None

x = lg.Const("x", lg.Boolean)
body = act.Sequence()
body.formal_params = [x]
im.module.actions["ext"] = body
im.module.imports = [ivy_ast.ImportDef(ivy_ast.Atom("ext"), ivy_ast.Atom(""))]
trace = DummyTrace()
trace.add_state([lg.Eq(x, lg.true)])
trace.last_action = act.CallAction(ivy_ast.Atom("ext"))
trace.add_state([lg.Eq(x, lg.true)])
print(json.dumps(str(trace)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python non-detailed call oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python non-detailed call oracle %q: %v", out, err)
	}

	mod := NewModule()
	mod.Cfg.TraceDetailed = false
	x := NewConst("x", Boolean)
	body := NewSequence()
	body.SetFormalParams([]*Const{x})
	mod.Actions.Set("ext", body)
	mod.Imports = []Node{NewConst("ext", TopS)}
	trace := NewTraceBase(nil, mod)
	trace.AddTraceState([]Expr{&Eq{T1: x, T2: True}})
	trace.LastAction = NewCallActionOn(NewActionsConfig(), NewConst("ext", ActionS))
	trace.AddTraceState([]Expr{&Eq{T1: x, T2: True}})

	if got := trace.String(); got != want {
		t.Fatalf("Go non-detailed call trace differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestTraceNonDetailedDebugEventMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_actions as act, logic as lg
from ivy import ivy_module as im, ivy_ast

im.module = im.Module()
ivy_trace.option_detailed.set("false")

class DummyTrace(ivy_trace.TraceBase):
    def get_universes(self):
        return None

x = lg.Const("x", lg.Boolean)
event = ivy_ast.Atom("event")
field = ivy_ast.Atom("field")
trace = DummyTrace()
trace.add_state([lg.Eq(x, lg.true)])
trace.last_action = act.DebugAction(event, ivy_ast.Equals(field, x))
trace.add_state([lg.Eq(x, lg.true)])
print(json.dumps(str(trace)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python non-detailed debug oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python non-detailed debug oracle %q: %v", out, err)
	}

	mod := NewModule()
	mod.Cfg.TraceDetailed = false
	x := NewConst("x", Boolean)
	event := NewConst("event", TopS)
	field := NewConst("field", TopS)
	trace := NewTraceBase(nil, mod)
	trace.AddTraceState([]Expr{&Eq{T1: x, T2: True}})
	trace.LastAction = NewDebugAction(event, &Eq{T1: field, T2: x})
	trace.AddTraceState([]Expr{&Eq{T1: x, T2: True}})

	if got := trace.String(); got != want {
		t.Fatalf("Go non-detailed debug trace differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestTracePretty(t *testing.T) {
	// Test truncation: maxLines=3 means lines[0:2] + "..."
	s := "line1\nline2\nline3\nline4\nline5\nline6"
	result := TracePretty(s, 3)
	lines := strings.Split(result, "\n")
	if len(lines) != 3 { // 2 original + "..."
		t.Errorf("expected 3 lines (2 + ...), got %d: %q", len(lines), result)
	}
	if lines[len(lines)-1] != "..." {
		t.Errorf("expected last line to be '...', got %q", lines[len(lines)-1])
	}
}

func TestTracePrettyFormatting(t *testing.T) {
	// Test brace indentation: "a { b; c }" should be reformatted
	s := "a { b; c }"
	result := TracePretty(s, 0)
	if !strings.Contains(result, "    b") {
		t.Errorf("expected indented content inside braces, got %q", result)
	}
}

func TestTracePrettyShort(t *testing.T) {
	s := "line1\nline2"
	result := TracePretty(s, 5)
	if result != s {
		t.Errorf("short string should be unchanged, got %q", result)
	}
}

func TestTraceLabelFromAction(t *testing.T) {
	action := NewAssumeAction(True)
	label := TraceLabelFromAction(action, nil)
	if label == "" {
		t.Error("TraceLabelFromAction should return non-empty string")
	}
}

func TestTraceLabelFromActionUsesSingularActionLabel(t *testing.T) {
	action := NewAssumeAction(True)
	action.SetLabel("call ext")
	if got := TraceLabelFromAction(action, nil); got != "call ext\n" {
		t.Fatalf("TraceLabelFromAction() = %q, want %q", got, "call ext\n")
	}
}

// --- EvalInState ---

func TestTraceEvalInState(t *testing.T) {
	param := NewConst("X", Boolean)
	val := NewConst("true_val", Boolean)
	eq, _ := NewEq(param, val)
	clauses := NewClauses([]Expr{eq}, nil, nil)
	state := NewState(nil, clauses)
	result := EvalInState(state, param)
	if result == nil {
		t.Fatal("EvalInState should find matching parameter")
	}
	if !result.Equal(val) {
		t.Errorf("expected %v, got %v", val, result)
	}
}

func TestTraceEvalInStateNotFound(t *testing.T) {
	clauses := NewClauses([]Expr{True}, nil, nil)
	state := NewState(nil, clauses)
	param := NewConst("missing", Boolean)
	result := EvalInState(state, param)
	if result != nil {
		t.Error("EvalInState should return nil for missing param")
	}
}

func TestTraceEvalInStateNilClauses(t *testing.T) {
	state := NewState(nil, nil)
	param := NewConst("X", Boolean)
	result := EvalInState(state, param)
	if result != nil {
		t.Error("EvalInState should return nil for nil clauses")
	}
}

// --- ValueToStr ---

func TestTraceValueToStrNil(t *testing.T) {
	s := ValueToStr(nil, nil)
	if s != "..." {
		t.Errorf("expected '...', got %q", s)
	}
}

func TestTraceValueToStrConst(t *testing.T) {
	c := NewConst("my_value", Boolean)
	s := ValueToStr(c, nil)
	if s != "my_value" {
		t.Errorf("expected 'my_value', got %q", s)
	}
}

func TestTraceValueToStrNode(t *testing.T) {
	n := True
	s := ValueToStr(n, nil)
	if s == "" {
		t.Error("ValueToStr should return non-empty string for True")
	}
}

func TestTraceValueToStrPythonArrayAndDestructor(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im
from ivy import ivy_logic as il

im.module = im.Module()

arr = il.UninterpretedSort("arr")
idx = il.UninterpretedSort("idx")
elem = il.UninterpretedSort("elem")
a = il.Symbol("a", arr)
end = il.Symbol("arr.end", il.FunctionSort(arr, idx))
value = il.Symbol("arr.value", il.FunctionSort(arr, idx, elem))
il.sig.symbols["arr.end"] = end
il.sig.symbols["arr.value"] = value
zero = il.Symbol("0", idx)
one = il.Symbol("1", idx)
two = il.Symbol("2", idx)
x = il.Symbol("x", elem)
y = il.Symbol("y", elem)
array_values = {
    str(end(a)): two,
    str(value(a, zero)): x,
    str(value(a, one)): y,
}
def eval_array(e):
    return array_values.get(str(e))

rec = il.UninterpretedSort("rec")
field_sort = il.UninterpretedSort("field")
r = il.Symbol("r", rec)
field = il.Symbol("field", il.FunctionSort(rec, field_sort))
im.module.sort_destructors["rec"].append(field)
v = il.Symbol("v", field_sort)
def eval_destructor(e):
    return v if str(e) == str(field(r)) else None

print(json.dumps({
    "array": ivy_trace.value_to_str(a, eval_array),
    "destructor": ivy_trace.value_to_str(r, eval_destructor),
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python ValueToStr oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want map[string]string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python ValueToStr oracle %q: %v", out, err)
	}

	mod := NewModule()
	arr := &UninterpretedSort{Name: "arr"}
	idx := &UninterpretedSort{Name: "idx"}
	elem := &UninterpretedSort{Name: "elem"}
	a := NewConst("a", arr)
	endSort, err := NewFunctionSort(arr, idx)
	if err != nil {
		t.Fatalf("end sort: %v", err)
	}
	end, err := mod.Sig.AddSymbol("arr.end", endSort)
	if err != nil {
		t.Fatalf("arr.end symbol: %v", err)
	}
	mod.Functions.Set(end.Name, end.CSort)
	valueSort, err := NewFunctionSort(arr, idx, elem)
	if err != nil {
		t.Fatalf("value sort: %v", err)
	}
	value, err := mod.Sig.AddSymbol("arr.value", valueSort)
	if err != nil {
		t.Fatalf("arr.value symbol: %v", err)
	}
	mod.Functions.Set(value.Name, value.CSort)

	arrayVals := map[NodeKey]Expr{
		Key(MustApply(end, a)):                       NewConst("2", idx),
		Key(MustApply(value, a, NewConst("0", idx))): NewConst("x", elem),
		Key(MustApply(value, a, NewConst("1", idx))): NewConst("y", elem),
	}
	gotArray := ValueToStrWithModule(mod, a, func(e Expr) Expr {
		return arrayVals[Key(e)]
	})
	if gotArray != want["array"] {
		t.Fatalf("Go array ValueToStr=%q, want Python %q", gotArray, want["array"])
	}

	rec := &UninterpretedSort{Name: "rec"}
	fieldSort := &UninterpretedSort{Name: "field"}
	r := NewConst("r", rec)
	destrSort, err := NewFunctionSort(rec, fieldSort)
	if err != nil {
		t.Fatalf("destructor sort: %v", err)
	}
	field, err := mod.Sig.AddSymbol("field", destrSort)
	if err != nil {
		t.Fatalf("field symbol: %v", err)
	}
	mod.SortDestructors["rec"] = []*Const{field}
	destrVals := map[NodeKey]Expr{
		Key(MustApply(field, r)): NewConst("v", fieldSort),
	}
	gotDestructor := ValueToStrWithModule(mod, r, func(e Expr) Expr {
		return destrVals[Key(e)]
	})
	if gotDestructor != want["destructor"] {
		t.Fatalf("Go destructor ValueToStr=%q, want Python %q", gotDestructor, want["destructor"])
	}
}

// --- Trace (model-based) ---

func TestTraceNewTrace(t *testing.T) {
	clauses := traceTestClauses()
	tr := NewTrace(nil, clauses, nil, nil, true)
	if tr == nil {
		t.Fatal("NewTrace returned nil")
	}
	if !tr.TopLevel {
		t.Error("TopLevel should be true")
	}
	if tr.Eqs == nil {
		t.Error("Eqs should be initialized")
	}
}

func TestTraceNewTraceNilClauses(t *testing.T) {
	tr := NewTrace(nil, nil, nil, nil, false)
	if tr == nil {
		t.Fatal("NewTrace returned nil")
	}
	if len(tr.Eqs) != 0 {
		t.Error("Eqs should be empty for nil clauses")
	}
}

func TestTraceImplementsAnnotationHandlerLikePython(t *testing.T) {
	var _ AnnotationHandler = (*Trace)(nil)
}

func TestTraceClonePreservesModelStateLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_logic_utils as lut, ivy_module as im, ivy_solver as slv
from ivy import logic as lg

im.module = im.Module()
x = lg.Const("x", lg.Boolean)
clauses = lut.Clauses([x])
model = slv.get_small_model(clauses, [], [], shrink=False)
vocab = [x]
tr = ivy_trace.Trace(clauses, model, vocab, True)
clone = tr.clone()
print(json.dumps({
    "same_clauses": clone.clauses is clauses,
    "same_model": clone.model is model,
    "same_vocab": clone.vocab is vocab,
    "top_level": clone.top_level,
    "eq_count": sum(len(v) for v in clone.eqs.values()),
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python trace clone oracle failed: %v\n%s", err, out)
	}
	var want struct {
		SameClauses bool `json:"same_clauses"`
		SameModel   bool `json:"same_model"`
		SameVocab   bool `json:"same_vocab"`
		TopLevel    bool `json:"top_level"`
		EqCount     int  `json:"eq_count"`
	}
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python trace clone oracle %q: %v", out, err)
	}

	x := NewConst("x", Boolean)
	clauses := NewClauses([]Expr{x}, nil, nil)
	model := traceTestModel{}
	vocab := []Expr{x}
	tr := NewTrace(nil, clauses, model, vocab, true)
	clone := tr.Clone()

	if want.SameClauses != (clone.Clauses == clauses) {
		t.Fatalf("clone clauses identity = %v, want %v", clone.Clauses == clauses, want.SameClauses)
	}
	if want.SameModel != (clone.Model == model) {
		t.Fatalf("clone model identity = %v, want %v", clone.Model == model, want.SameModel)
	}
	if want.SameVocab != (&clone.Vocab[0] == &vocab[0]) {
		t.Fatalf("clone vocab backing identity = %v, want %v", &clone.Vocab[0] == &vocab[0], want.SameVocab)
	}
	if clone.TopLevel != want.TopLevel {
		t.Fatalf("clone TopLevel = %v, want %v", clone.TopLevel, want.TopLevel)
	}
	gotEqCount := 0
	for _, eqs := range clone.Eqs {
		gotEqCount += len(eqs)
	}
	if gotEqCount != want.EqCount {
		t.Fatalf("clone equation count = %d, want %d", gotEqCount, want.EqCount)
	}
}

func TestTraceHandleSkipsNowhereActionLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_actions as act, ivy_utils as iu
from ivy import logic as lg

im.module = im.Module()

class T(ivy_trace.TraceBase):
    def get_universes(self):
        return None
    def new_state(self, env):
        self.add_state([lg.Eq(lg.Const("seen", lg.Boolean), lg.true)])

tr = T()
action = act.AssumeAction(lg.true)
action.lineno = iu.Location("nowhere", 0)
tr.handle(action, {})
print(json.dumps({
    "states": len(tr.states),
    "has_last_action": tr.last_action is not None,
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python nowhere-action oracle failed: %v\n%s", err, out)
	}
	var want struct {
		States        int  `json:"states"`
		HasLastAction bool `json:"has_last_action"`
	}
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python nowhere-action oracle %q: %v", out, err)
	}

	tr := NewTrace(nil, nil, nil, nil, true)
	action := NewAssumeAction(True)
	action.SetLineno(Location{Filename: "nowhere", Line: 0})
	tr.Handle(action, nil)

	if got := len(tr.TraceStates); got != want.States {
		t.Fatalf("Trace.Handle nowhere action states=%d, want %d", got, want.States)
	}
	if got := tr.LastAction != nil; got != want.HasLastAction {
		t.Fatalf("Trace.Handle nowhere action has LastAction=%v, want %v", got, want.HasLastAction)
	}
}

func TestTraceStateReceivesModelUniverseLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_logic_utils as lut, ivy_module as im, ivy_solver as slv
from ivy import logic as lg

im.module = im.Module()
S = lg.UninterpretedSort("S")
a = lg.Const("a", S)
b = lg.Const("b", S)
clauses = lut.Clauses([lg.Not(lg.Eq(a, b))])
model = slv.get_small_model(clauses, [S], [], shrink=True)
tr = ivy_trace.Trace(clauses, model, [a, b], True)
tr.add_state([])
universe = getattr(tr.states[0], "universe", None)
print(json.dumps({
    "has_universe": universe is not None,
    "count": sum(len(v) for v in universe.values()) if universe is not None else 0,
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python trace universe oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		HasUniverse bool `json:"has_universe"`
		Count       int  `json:"count"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python trace universe oracle %q: %v", out, err)
	}

	tr := NewTrace(nil, nil, traceUniverseModel{}, nil, true)
	tr.AddState(nil)
	if len(tr.TraceStates) != 1 {
		t.Fatalf("expected one trace state, got %d", len(tr.TraceStates))
	}
	universe, ok := tr.TraceStates[0].State.Universe.(map[string][]Expr)
	if ok != want.HasUniverse {
		t.Fatalf("trace state universe present=%v, want %v", ok, want.HasUniverse)
	}
	gotCount := 0
	for _, elems := range universe {
		gotCount += len(elems)
	}
	if gotCount != want.Count {
		t.Fatalf("trace state universe element count=%d, want %d", gotCount, want.Count)
	}
}

// --- MakeCheckArt ---

func TestTraceMakeCheckArt(t *testing.T) {
	mod := traceTestModule()
	ag, post, fail, err := MakeCheckArt(mod, "", nil)
	if err != nil {
		t.Fatalf("MakeCheckArt error: %v", err)
	}
	if ag == nil {
		t.Fatal("MakeCheckArt returned nil graph")
	}
	if post == nil {
		t.Fatal("MakeCheckArt returned nil post-state")
	}
	if fail == nil {
		t.Fatal("MakeCheckArt returned nil fail-state")
	}
	if len(ag.States) < 1 {
		t.Errorf("expected at least 1 state, got %d", len(ag.States))
	}
}

func TestTraceMakeCheckArtWithPrecond(t *testing.T) {
	mod := traceTestModule()
	precond := []*Clauses{traceTestClauses(), traceTestClauses()}
	ag, post, fail, err := MakeCheckArt(mod, "", precond)
	if err != nil {
		t.Fatalf("MakeCheckArt error: %v", err)
	}
	if ag == nil || post == nil || fail == nil {
		t.Fatal("MakeCheckArt with preconds returned nil")
	}
}

func TestTraceMakeCheckArtPythonTupleAndFailState(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_actions as act

im.module = im.Module()
im.module.actions["ext:foo"] = act.Sequence()
im.module.public_actions["ext:foo"] = None
ag, post, fail = ivy_trace.make_check_art()

print(json.dumps({
    "states": len(ag.states),
    "post_id": post.id,
    "post_is_graph_state": post is ag.states[0],
    "fail_pred_is_post_pred": fail.pred is post.pred,
    "post_expr_rep_type": type(post.expr.rep).__name__,
    "fail_expr_rep_type": type(fail.expr.rep).__name__,
    "fail_expr_rep_str": str(fail.expr.rep),
    "post_clauses": str(post.clauses),
    "fail_clauses": str(fail.clauses),
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python MakeCheckArt oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		States             int    `json:"states"`
		PostID             int    `json:"post_id"`
		PostIsGraphState   bool   `json:"post_is_graph_state"`
		FailPredIsPostPred bool   `json:"fail_pred_is_post_pred"`
		PostExprRepType    string `json:"post_expr_rep_type"`
		FailExprRepType    string `json:"fail_expr_rep_type"`
		FailExprRepStr     string `json:"fail_expr_rep_str"`
		PostClauses        string `json:"post_clauses"`
		FailClauses        string `json:"fail_clauses"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python MakeCheckArt oracle %q: %v", out, err)
	}

	mod := NewModule()
	mod.Actions.Set("ext:foo", NewSequence())
	mod.PublicActions.Set("ext:foo", true)

	ag, succeed, fail, err := MakeCheckArt(mod, "", nil)
	if err != nil {
		t.Fatalf("MakeCheckArt failed: %v", err)
	}
	if len(ag.States) != want.States {
		t.Fatalf("Go MakeCheckArt graph state count=%d, want Python %d", len(ag.States), want.States)
	}
	if succeed == nil || fail == nil {
		t.Fatalf("MakeCheckArt returned nil succeed/fail: succeed=%#v fail=%#v", succeed, fail)
	}
	if succeed.ID != want.PostID {
		t.Fatalf("Go MakeCheckArt succeed ID=%d, want Python post ID=%d", succeed.ID, want.PostID)
	}
	if (len(ag.States) > 0 && ag.States[0] == succeed) != want.PostIsGraphState {
		t.Fatalf("Go MakeCheckArt succeed graph-state membership differs from Python")
	}
	if (fail.Pred == succeed.Pred) != want.FailPredIsPostPred {
		t.Fatalf("Go MakeCheckArt fail predecessor does not match Python fail_expr predecessor")
	}
	succeedApp, ok := succeed.Prov.(*ActionApp)
	if !ok {
		t.Fatalf("Go MakeCheckArt succeed provenance is not ActionApp like Python: %#v", succeed.Prov)
	}
	if _, ok := succeedApp.Rep.(*LogicEnvAction); !ok || want.PostExprRepType != "EnvAction" {
		t.Fatalf("Go MakeCheckArt succeed provenance is not EnvAction like Python: %#v", succeed.Prov)
	}
	failApp, ok := fail.Prov.(*ActionApp)
	if !ok {
		t.Fatalf("Go MakeCheckArt fail provenance is not ActionApp: %#v", fail.Prov)
	}
	failAction, ok := failApp.Rep.(*FailAction)
	if !ok || want.FailExprRepType != "fail_action" {
		t.Fatalf("Go MakeCheckArt fail provenance is not FailAction like Python: %#v", failApp.Rep)
	}
	if got := failAction.String(); got != want.FailExprRepStr {
		t.Fatalf("Go MakeCheckArt fail action string=%q, want Python %q", got, want.FailExprRepStr)
	}
	if want.PostClauses != "true" || succeed.Clauses == nil || !succeed.Clauses.IsTrue() {
		t.Fatalf("Go MakeCheckArt post clauses=%q, want Python %q", succeed.Clauses, want.PostClauses)
	}
	if want.FailClauses != "true" || fail.Clauses == nil || !fail.Clauses.IsTrue() {
		t.Fatalf("Go MakeCheckArt fail clauses=%q, want Python %q", fail.Clauses, want.FailClauses)
	}
	for _, st := range ag.States {
		if st == fail {
			t.Fatalf("Go MakeCheckArt fail state is in the graph; Python fail state is standalone")
		}
	}
}

// --- CheckFinalCond ---

func TestTraceCheckFinalCondNilPost(t *testing.T) {
	ag := NewAnalysisGraph(nil, nil)
	result := CheckFinalCond(ag, nil, traceTestClauses(), nil, false)
	if result != nil {
		t.Error("CheckFinalCond with nil post should return nil")
	}
}

func TestCheckFinalCondNilFinalCondMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_logic_utils as lut, ivy_actions as act

im.module = im.Module()

class History(object):
    def __init__(self):
        self.post = lut.true_clauses(act.EmptyAnnotation())
        self.actions = []

class AnalysisGraph(object):
    def get_history(self, post):
        return History()

res = ivy_trace.check_final_cond(AnalysisGraph(), object(), None)
print(json.dumps({
    "is_none": res is None,
    "state_count": None if res is None else len(res.states),
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CheckFinalCond nil-final oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		IsNone     bool `json:"is_none"`
		StateCount *int `json:"state_count"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python CheckFinalCond nil-final oracle %q: %v", out, err)
	}

	mod := NewModule()
	ag := NewAnalysisGraph(mod)
	state := NewState(mod, TrueClauses(nil))
	result := CheckFinalCond(ag, state, nil, nil, false)

	gotIsNone := result == nil
	if gotIsNone != want.IsNone {
		t.Fatalf("CheckFinalCond nil final-cond nilness differs from Python\nwant nil: %v\ngot nil:  %v", want.IsNone, gotIsNone)
	}
	if result != nil && want.StateCount != nil && len(result.TraceStates) != *want.StateCount {
		t.Fatalf("CheckFinalCond nil final-cond state count differs from Python\nwant: %d\ngot:  %d", *want.StateCount, len(result.TraceStates))
	}
}

func TestCheckFinalCondPanicsOnMissingHistoryAnnotationLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_logic_utils as lut, logic as lg

im.module = im.Module()

class History(object):
    def __init__(self):
        self.post = lut.Clauses([lg.true])
        self.actions = []

class AnalysisGraph(object):
    def get_history(self, post):
        return History()

try:
    ivy_trace.check_final_cond(AnalysisGraph(), object(), lut.true_clauses())
    print(json.dumps("no-panic"))
except AssertionError:
    print(json.dumps("AssertionError"))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CheckFinalCond annotation oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python CheckFinalCond annotation oracle %q: %v", out, err)
	}
	if want != "AssertionError" {
		t.Fatalf("python CheckFinalCond oracle returned %q, want AssertionError", want)
	}

	mod := NewModule()
	ag := NewAnalysisGraph(mod)
	post := &State{
		ID:      -1,
		Domain:  mod,
		Clauses: NewClauses([]Expr{True}, nil, nil),
		InScope: make(map[string]bool),
	}

	didPanic := false
	func() {
		defer func() {
			if recover() != nil {
				didPanic = true
			}
		}()
		_ = CheckFinalCond(ag, post, TrueClauses(EmptyAnnotation{}), nil, false)
	}()
	if !didPanic {
		t.Fatalf("Go CheckFinalCond did not panic on missing history annotation; Python raised %s", want)
	}
}

// --- CheckVC ---

func TestTraceCheckVCNilClauses(t *testing.T) {
	result := CheckVC(nil, nil, nil, nil, nil, false)
	if result != nil {
		t.Error("CheckVC with nil clauses should return nil")
	}
}

func TestTraceCheckVCNoAnnot(t *testing.T) {
	clauses := NewClauses([]Expr{True}, nil, nil)
	// Annot is nil
	result := CheckVC(nil, clauses, nil, nil, nil, false)
	if result != nil {
		t.Error("CheckVC with nil annot should return nil")
	}
}

func TestCheckVCBuildsModelValuationTraceLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_logic_utils as lut, ivy_actions as act
from ivy import logic as lg

im.module = im.Module()
p = lg.Const("p", lg.Boolean)
q = lg.Const("q", lg.Boolean)
clauses = lut.Clauses([lg.Or(p, q)], annot=act.EmptyAnnotation())
tr = ivy_trace.check_vc(clauses, act.Sequence())
print(json.dumps([[str(f) for f in st.clauses.fmlas] for st in tr.states]))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CheckVC oracle failed: %v\n%s", err, out)
	}
	var want [][]string
	outLines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if err := json.Unmarshal([]byte(outLines[len(outLines)-1]), &want); err != nil {
		t.Fatalf("decode python CheckVC oracle %q: %v", out, err)
	}
	if len(want) == 0 {
		t.Fatal("python CheckVC oracle returned no states")
	}

	mod := NewModule()
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	clauses := NewClauses([]Expr{&LogicOr{Terms: []Expr{p, q}}}, nil, EmptyAnnotation{})
	gotTrace := CheckVC(mod, clauses, NewSequence(), nil, nil, false)
	if gotTrace == nil || len(gotTrace.TraceStates) == 0 {
		t.Fatalf("Go CheckVC returned no trace states: %#v", gotTrace)
	}
	var got []string
	for _, fmla := range gotTrace.TraceStates[0].State.Clauses.Fmlas {
		got = append(got, fmla.String())
	}
	if strings.Join(got, "\n") != strings.Join(want[0], "\n") {
		t.Fatalf("Go CheckVC trace state differs from Python\nwant: %#v\ngot:  %#v", want[0], got)
	}
}

func TestCheckVCReturnsDefaultHiddenSymbolsLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_logic_utils as lut, ivy_actions as act
from ivy import logic as lg

im.module = im.Module()
p = lg.Const("p", lg.Boolean)
tr = ivy_trace.check_vc(lut.Clauses([p], annot=act.EmptyAnnotation()), act.Sequence())
print(json.dumps(tr.hidden_symbols(p)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CheckVC hidden-symbol oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want bool
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python CheckVC hidden-symbol oracle %q: %v", out, err)
	}

	p := NewConst("p", Boolean)
	gotTrace := CheckVC(NewModule(), NewClauses([]Expr{p}, nil, EmptyAnnotation{}), NewSequence(), nil, nil, false)
	if gotTrace == nil {
		t.Fatal("Go CheckVC returned nil trace")
	}
	if gotTrace.HiddenSymbols == nil {
		t.Fatal("Go CheckVC returned trace with nil HiddenSymbols")
	}
	if got := gotTrace.HiddenSymbols("p"); got != want {
		t.Fatalf("Go CheckVC default HiddenSymbols(\"p\")=%v, want %v", got, want)
	}
	_ = gotTrace.String()
}

func TestCheckVCUsesRelationsToMinimizeLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_logic_utils as lut, ivy_actions as act
from ivy import logic as lg

im.module = im.Module()
S = lg.UninterpretedSort("S")
a = lg.Const("a", S)
b = lg.Const("b", S)
r = lg.Const("r", lg.FunctionSort(S, lg.Boolean))
clauses = lut.Clauses([lg.Not(lg.Eq(a, b)), lg.Or(r(a), r(b))], annot=act.EmptyAnnotation())
tr = ivy_trace.check_vc(clauses, act.Sequence(), rels_to_min=[r], shrink=True)
print(json.dumps([[str(f) for f in st.clauses.fmlas] for st in tr.states]))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CheckVC relation-min oracle failed: %v\n%s", err, out)
	}
	outLines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want [][]string
	if err := json.Unmarshal([]byte(outLines[len(outLines)-1]), &want); err != nil {
		t.Fatalf("decode python CheckVC relation-min oracle %q: %v", out, err)
	}
	if len(want) == 0 {
		t.Fatal("python CheckVC relation-min oracle returned no states")
	}
	sort.Strings(want[0])

	mod := NewModule()
	S := &UninterpretedSort{Name: "S"}
	a := NewConst("a", S)
	b := NewConst("b", S)
	fs, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	r := NewConst("r", fs)
	ra := MustApply(r, a)
	rb := MustApply(r, b)
	clauses := NewClauses([]Expr{
		&LogicNot{Body: &Eq{T1: a, T2: b}},
		&LogicOr{Terms: []Expr{ra, rb}},
	}, nil, EmptyAnnotation{})
	gotTrace := CheckVC(mod, clauses, NewSequence(), nil, []string{"r"}, true)
	if gotTrace == nil || len(gotTrace.TraceStates) == 0 {
		t.Fatalf("Go CheckVC returned no trace states: %#v", gotTrace)
	}
	var got []string
	for _, fmla := range gotTrace.TraceStates[0].State.Clauses.Fmlas {
		got = append(got, fmla.String())
	}
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want[0], "\n") {
		t.Fatalf("Go CheckVC relation-min trace differs from Python\nwant: %#v\ngot:  %#v", want[0], got)
	}
}

func TestTraceHistoryRelationsToMinimizeAppliesHistoryMapLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import logic as lg

r = lg.Const("r", lg.Boolean)
renamed = lg.Const("__r", lg.Boolean)
history = type("History", (), {"maps": [{r: renamed}]})()
rels_to_min = []
for x in "r missing".split():
    relation = {"r": r}.get(x, lg.Const(x, lg.Boolean))
    relation = history.maps[0].get(relation, relation)
    rels_to_min.append(str(relation))
print(json.dumps(rels_to_min))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CTI relation-min history-map oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python CTI relation-min history-map oracle %q: %v", out, err)
	}

	mod := NewModule()
	mod.Relations.Set("r", Boolean)
	history := &History{Maps: []LogicRenaming{LogicRenaming{}}}
	history.Maps[0].Set(NewConst("r", Boolean), NewConst("__r", Boolean))
	got := traceHistoryRelationsToMinimize(mod, history, []string{"r", "missing"})
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("history relation minimization rewrite differs from Python\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestCheckVCMinimizesAllSignatureUninterpretedSortsLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_logic as il, ivy_logic_utils as lut, ivy_actions as act
from ivy import logic as lg

im.module = im.Module()
S = il.UninterpretedSort("S")
il.sig.sorts["S"] = S
p = il.Symbol("p", lg.Boolean)
il.sig.symbols["p"] = p
tr = ivy_trace.check_vc(lut.Clauses([p], annot=act.EmptyAnnotation()), act.Sequence(), shrink=True)
keys = [] if tr is None or not tr.states else sorted(s.name for s in tr.states[0].universe.keys())
print(json.dumps(keys))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CheckVC signature-sort oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python CheckVC signature-sort oracle %q: %v", out, err)
	}

	mod := NewModule()
	sortS := &UninterpretedSort{Name: "S"}
	mod.Sig.Sorts.Set("S", sortS)
	p, err := mod.Sig.AddSymbol("p", Boolean)
	if err != nil {
		t.Fatalf("AddSymbol(p): %v", err)
	}
	tr := CheckVC(mod, NewClauses([]Expr{p}, nil, EmptyAnnotation{}), NewSequence(), nil, nil, true)
	if tr == nil || len(tr.TraceStates) == 0 {
		t.Fatalf("Go CheckVC returned no trace: %#v", tr)
	}
	gotUniverse, _ := tr.TraceStates[0].State.Universe.(map[string][]Expr)
	got := make([]string, 0, len(gotUniverse))
	for name := range gotUniverse {
		got = append(got, name)
	}
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Go CheckVC minimized universe keys differ from Python\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestCheckVCRespectsShrinkFalseLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_logic as il, ivy_logic_utils as lut, ivy_actions as act
from ivy import logic as lg

im.module = im.Module()
S = il.UninterpretedSort("S")
il.sig.sorts["S"] = S
p = il.Symbol("p", lg.Boolean)
il.sig.symbols["p"] = p
tr = ivy_trace.check_vc(lut.Clauses([p], annot=act.EmptyAnnotation()), act.Sequence(), shrink=False)
keys = [] if tr is None or not tr.states else sorted(s.name for s in tr.states[0].universe.keys())
print(json.dumps(keys))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CheckVC shrink oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python CheckVC shrink oracle %q: %v", out, err)
	}

	mod := NewModule()
	sortS := &UninterpretedSort{Name: "S"}
	mod.Sig.Sorts.Set("S", sortS)
	p, err := mod.Sig.AddSymbol("p", Boolean)
	if err != nil {
		t.Fatalf("AddSymbol(p): %v", err)
	}
	tr := CheckVC(mod, NewClauses([]Expr{p}, nil, EmptyAnnotation{}), NewSequence(), nil, nil, false)
	if tr == nil || len(tr.TraceStates) == 0 {
		t.Fatalf("Go CheckVC returned no trace: %#v", tr)
	}
	gotUniverse, _ := tr.TraceStates[0].State.Universe.(map[string][]Expr)
	got := make([]string, 0, len(gotUniverse))
	for name := range gotUniverse {
		got = append(got, name)
	}
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Go CheckVC shrink=false universe keys differ from Python\nwant: %#v\ngot:  %#v", want, got)
	}
}

// --- MakeVC ---

func TestTraceMakeVC(t *testing.T) {
	action := NewAssumeAction(True)
	pre := []*Clauses{traceTestClauses()}
	post := []*Clauses{traceTestClauses()}
	vc := MakeVC(action, pre, post, true)
	if vc == nil {
		t.Fatal("MakeVC returned nil")
	}
}

func TestTraceMakeVCPythonExecutesActionAndDualizesPost(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_actions as act, ivy_logic as il, logic as lg

im.module = im.Module()
p = il.Symbol("p", lg.Boolean)
vc = ivy_trace.make_vc(act.AssumeAction(p), [], [])
print(json.dumps({
    "fmlas": [str(f) for f in vc.fmlas],
    "defs": [str(d) for d in vc.defs],
    "annot": type(vc.annot).__name__,
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python MakeVC oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		Fmlas []string `json:"fmlas"`
		Defs  []string `json:"defs"`
		Annot string   `json:"annot"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python MakeVC oracle %q: %v", out, err)
	}

	p := NewConst("p", Boolean)
	vc := MakeVC(NewAssumeAction(p), nil, nil, true)
	gotFmlas := make([]string, len(vc.Fmlas))
	for i, f := range vc.Fmlas {
		gotFmlas[i] = f.String()
	}
	gotDefs := make([]string, len(vc.Defs))
	for i, d := range vc.Defs {
		gotDefs[i] = d.String()
	}
	if strings.Join(gotFmlas, "\n") != strings.Join(want.Fmlas, "\n") {
		t.Fatalf("Go MakeVC formulas differ from Python\nwant: %#v\ngot:  %#v", want.Fmlas, gotFmlas)
	}
	if strings.Join(gotDefs, "\n") != strings.Join(want.Defs, "\n") {
		t.Fatalf("Go MakeVC defs differ from Python\nwant: %#v\ngot:  %#v", want.Defs, gotDefs)
	}
	if _, ok := vc.Annot.(EmptyAnnotation); !ok || want.Annot != "EmptyAnnotation" {
		t.Fatalf("Go MakeVC annot=%T, want Python %s", vc.Annot, want.Annot)
	}
}

// --- ToLines ---

func TestTraceToLinesEmpty(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	var lines []string
	hash := make(map[string]string)
	tb.ToLines(&lines, hash, 0, func(s string) bool { return false }, false, nil, nil)
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for empty trace, got %d", len(lines))
	}
}

func TestTraceToLinesWithState(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	eq := makeEq("X", "val1")
	tb.AddTraceState([]Expr{eq})
	var lines []string
	hash := make(map[string]string)
	tb.ToLines(&lines, hash, 0, func(s string) bool { return false }, false, nil, nil)
	joined := strings.Join(lines, "")
	if !strings.Contains(joined, "[") {
		t.Error("ToLines should contain state bracket for detailed mode")
	}
}

func TestTraceToLinesHidden(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	eq := makeEq("hidden_sym", "val")
	tb.AddTraceState([]Expr{eq})
	var lines []string
	hash := make(map[string]string)
	tb.ToLines(&lines, hash, 0, func(s string) bool {
		return s == "hidden_sym"
	}, false, nil, nil)
	joined := strings.Join(lines, "")
	if strings.Contains(joined, "hidden_sym") {
		t.Error("hidden symbols should be excluded from output")
	}
}

func TestTraceToLinesLoopStart(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	tb.AddTraceState(nil)
	tb.TraceStates[0].LoopStart = true
	var lines []string
	hash := make(map[string]string)
	tb.ToLines(&lines, hash, 0, func(s string) bool { return false }, false, nil, nil)
	joined := strings.Join(lines, "")
	if !strings.Contains(joined, "repeats infinitely") {
		t.Error("loop start should show 'repeats infinitely'")
	}
}

func TestL2SMarkLoopStartUsesSavedStateIndexLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im, ivy_l2s
from ivy import logic as lg

im.module = im.Module()

class T(ivy_trace.TraceBase):
    def get_universes(self):
        return None

tr = T()
tr.add_state([lg.Eq(lg.Const("a", lg.Boolean), lg.true)])
tr.add_state([lg.Eq(lg.Const("middle", lg.Boolean), lg.true)])
tr.add_state([lg.Eq(lg.Const("l2s_saved", lg.Boolean), lg.true)])
ivy_l2s.trace_hook(tr, [])
print(json.dumps(str(tr)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python l2s trace hook oracle failed: %v\n%s", err, out)
	}
	var want string
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python l2s trace hook oracle %q: %v", out, err)
	}

	tr := NewTrace(nil, nil, nil, nil, true)
	tr.AddState([]Expr{&Eq{T1: NewConst("a", Boolean), T2: True}})
	tr.AddState([]Expr{&Eq{T1: NewConst("middle", Boolean), T2: True}})
	tr.AddState([]Expr{&Eq{T1: NewConst("l2s_saved", Boolean), T2: True}})
	markLoopStart(tr)

	if !tr.TraceStates[1].LoopStart {
		t.Fatalf("expected l2s_saved in state 2 to mark state 1 as loop start")
	}
	if tr.TraceStates[0].LoopStart {
		t.Fatalf("state 0 was marked, but Python marks the predecessor of the saved state")
	}
	if got := tr.String(); got != want {
		t.Fatalf("Go l2s loop marker differs from Python\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestTraceBaseRenderHookFieldsMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im
from ivy import logic as lg

im.module = im.Module()

class T(ivy_trace.TraceBase):
    def get_universes(self):
        return None

old = lg.Const("old", lg.Boolean)
renamed = lg.Const("renamed", lg.Boolean)
pp_in = lg.Const("pp_in", lg.Boolean)
pp_out = lg.Const("pp_out", lg.Boolean)

tr = T()
tr.add_state([
    lg.Eq(lg.Const("l2s_saved", lg.Boolean), lg.true),
    lg.Eq(old, lg.false),
    lg.Eq(pp_in, lg.false),
])
tr.states[0].loop_start = True
tr.hidden_symbols = lambda sym: sym.name.startswith("l2s")
tr.rename({old: renamed})

def pp(c):
    return lg.Eq(pp_out, lg.true) if c.args[0] == pp_in else c

tr.pp = pp
print(json.dumps(str(tr)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python trace oracle failed: %v\n%s", err, out)
	}
	var want string
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python trace oracle %q: %v", out, err)
	}

	tb := NewTraceBase(nil, NewModule())
	old := NewConst("old", Boolean)
	ppIn := NewConst("pp_in", Boolean)
	tb.AddTraceState([]Expr{
		&Eq{T1: NewConst("l2s_saved", Boolean), T2: True},
		&Eq{T1: old, T2: False},
		&Eq{T1: ppIn, T2: False},
	})
	tb.TraceStates[0].LoopStart = true
	tb.HiddenSymbols = func(name string) bool { return strings.HasPrefix(name, "l2s") }
	tb.Rename(map[string]string{"old": "renamed"})
	tb.PP = func(expr Expr) Expr {
		if eq, ok := expr.(*Eq); ok {
			if c, ok := eq.T1.(*Const); ok && c.Name == "pp_in" {
				return &Eq{T1: NewConst("pp_out", Boolean), T2: True}
			}
		}
		return expr
	}

	if got := tb.String(); got != want {
		t.Fatalf("Go trace rendering differs from Python\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestTraceStateEquationNumericSuppressionMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im
from ivy import logic as lg

im.module = im.Module()

class T(ivy_trace.TraceBase):
    def get_universes(self):
        return None

s = lg.TopSort()
lt = lg.Const("<", lg.FunctionSort(s, s, lg.Boolean))
zero = lg.Const("0", s)
one = lg.Const("1", s)
tr = T()
tr.add_state([lg.Eq(lt(zero, one), lg.true)])
print(json.dumps(str(tr)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python numeric trace oracle failed: %v\n%s", err, out)
	}
	var want string
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python numeric trace oracle %q: %v", out, err)
	}

	fs, err := NewFunctionSort(TopS, TopS, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	lt := NewConst("<", fs)
	zero := NewConst("0", TopS)
	one := NewConst("1", TopS)
	lhs, err := lt.Call(zero, one)
	if err != nil {
		t.Fatal(err)
	}
	tb := NewTraceBase(nil, NewModule())
	tb.AddTraceState([]Expr{&Eq{T1: lhs, T2: True}})
	if got := tb.String(); got != want {
		t.Fatalf("Go numeric trace suppression differs from Python\nwant: %q\ngot:  %q", want, got)
	}
}

func TestTraceRenamingOverlappingNonceNamesMatchesPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_trace, ivy_module as im
from ivy import logic as lg

im.module = im.Module()

class T(ivy_trace.TraceBase):
    def get_universes(self):
        return None

one = lg.Const("l2s_s_1", lg.Boolean)
ten = lg.Const("l2s_s_10", lg.Boolean)
tr = T()
tr.add_state([lg.Eq(ten, lg.false)])
tr.rename({
    one: lg.Const("orig1", lg.Boolean),
    ten: lg.Const("orig10", lg.Boolean),
})
print(json.dumps(str(tr)))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python overlapping rename oracle failed: %v\n%s", err, out)
	}
	var want string
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python overlapping rename oracle %q: %v", out, err)
	}

	tb := NewTraceBase(nil, NewModule())
	tb.AddTraceState([]Expr{&Eq{T1: NewConst("l2s_s_10", Boolean), T2: False}})
	tb.Rename(map[string]string{
		"l2s_s_1":  "orig1",
		"l2s_s_10": "orig10",
	})
	if got := tb.String(); got != want {
		t.Fatalf("Go overlapping rename differs from Python\nwant:\n%q\ngot:\n%q", want, got)
	}
}

// --- isCallOrEnv / isCallAction ---

func TestTraceIsCallOrEnv(t *testing.T) {
	if isCallOrEnv(nil) {
		t.Error("nil should not be call or env")
	}
	assume := NewAssumeAction(True)
	if isCallOrEnv(assume) {
		t.Error("AssumeAction should not be call or env")
	}
}

func TestTraceIsCallAction(t *testing.T) {
	if isCallAction(nil) {
		t.Error("nil should not be call action")
	}
	assume := NewAssumeAction(True)
	if isCallAction(assume) {
		t.Error("AssumeAction should not be call action")
	}
}

// --- Fuzz ---

func FuzzValueToStr(f *testing.F) {
	f.Add("const_name")
	f.Add("")
	f.Add("x.y.z")
	f.Add("__skolem")
	f.Fuzz(func(t *testing.T, name string) {
		c := NewConst(name, Boolean)
		s := ValueToStr(c, nil)
		if s != name {
			t.Errorf("ValueToStr(Const(%q)) = %q, want %q", name, s, name)
		}
	})
}
