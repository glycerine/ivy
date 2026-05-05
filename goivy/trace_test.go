package goivy

import (
	"strings"
	"testing"
)

// --- helpers ---

func traceTestModule() *Module {
	return New()
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

func TestTraceTraceBaseClone(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	tb.AddTraceState(nil)
	clone := tb.Clone()
	if clone == tb {
		t.Error("Clone should return a new TraceBase")
	}
	if len(clone.TraceStates) != 0 {
		t.Error("cloned trace should have no states")
	}
}

func TestTraceTraceBaseHandle(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	action := NewAssumeAction(True)
	env := map[string]string{}
	tb.Handle(action, env)
	if tb.LastAction != action {
		t.Error("Handle should set LastAction")
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

func TestTraceTraceBaseEnd(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	tb.End()
	// Should have added a final state
	if len(tb.TraceStates) != 1 {
		t.Errorf("End should add final state, got %d states", len(tb.TraceStates))
	}
}

func TestTraceTraceBaseEndWithSub(t *testing.T) {
	tb := NewTraceBase(nil, nil)
	tb.Sub = NewTraceBase(nil, nil)
	tb.End()
	if tb.Sub != nil {
		t.Error("End should clear Sub")
	}
	if tb.Returned == nil {
		t.Error("End should set Returned from Sub")
	}
}

// --- Pretty / label helpers ---

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

// --- MakeCheckArt ---

func TestTraceMakeCheckArt(t *testing.T) {
	mod := traceTestModule()
	ag, pre, _, err := MakeCheckArt(mod, "test_action", nil)
	if err != nil {
		t.Fatalf("MakeCheckArt error: %v", err)
	}
	if ag == nil {
		t.Fatal("MakeCheckArt returned nil graph")
	}
	if pre == nil {
		t.Fatal("MakeCheckArt returned nil pre-state")
	}
	if len(ag.States) < 1 {
		t.Errorf("expected at least 1 state, got %d", len(ag.States))
	}
}

func TestTraceMakeCheckArtWithPrecond(t *testing.T) {
	mod := traceTestModule()
	precond := []*Clauses{traceTestClauses(), traceTestClauses()}
	ag, pre, _, err := MakeCheckArt(mod, "test", precond)
	if err != nil {
		t.Fatalf("MakeCheckArt error: %v", err)
	}
	if ag == nil || pre == nil {
		t.Fatal("MakeCheckArt with preconds returned nil")
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

func TestTraceCheckFinalCondNilFinalCond(t *testing.T) {
	ag := NewAnalysisGraph(nil, nil)
	state := NewState(nil, traceTestClauses())
	result := CheckFinalCond(ag, state, nil, nil, false)
	if result != nil {
		t.Error("CheckFinalCond with nil final cond should return nil")
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
