package trace

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// --- helpers ---

func testModule() *module.Module {
	return module.New()
}

func testClauses(fmlas ...lg.Node) *clauseops.Clauses {
	if len(fmlas) == 0 {
		fmlas = []lg.Node{lg.True}
	}
	return clauseops.NewClauses(fmlas, nil, nil)
}

func makeEq(name string, val string) lg.Node {
	c := lg.NewConst(name, lg.Boolean)
	v := lg.NewConst(val, lg.Boolean)
	eq, _ := lg.NewEq(c, v)
	return eq
}

// --- TraceBase tests ---

func TestNewTraceBase(t *testing.T) {
	tb := NewTraceBase(nil)
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

func TestNewTraceBaseWithModule(t *testing.T) {
	mod := testModule()
	tb := NewTraceBase(mod)
	if tb.Domain != mod {
		t.Error("domain should match provided module")
	}
}

func TestTraceBaseRename(t *testing.T) {
	tb := NewTraceBase(nil)
	m := map[string]string{"old_x": "x", "old_y": "y"}
	result := tb.Rename(m)
	if result != tb {
		t.Error("Rename should return same TraceBase")
	}
	if len(tb.Renaming) != 2 {
		t.Errorf("expected 2 entries, got %d", len(tb.Renaming))
	}
}

func TestIsSkolem(t *testing.T) {
	tests := []struct {
		name   string
		expect bool
	}{
		{"x", false},
		{"foo__bar", true},
		{"__X", false},    // global skolem, uppercase
		{"__xbar", true},  // regular skolem, lowercase
		{"__Ytest", false}, // global skolem
		{"abc", false},
		{"a__b", true},
	}
	for _, tt := range tests {
		got := IsSkolem(tt.name)
		if got != tt.expect {
			t.Errorf("IsSkolem(%q) = %v, want %v", tt.name, got, tt.expect)
		}
	}
}

func TestAddTraceState(t *testing.T) {
	tb := NewTraceBase(nil)
	tb.AddTraceState(nil)
	if len(tb.TraceStates) != 1 {
		t.Errorf("expected 1 trace state, got %d", len(tb.TraceStates))
	}
	if len(tb.States) != 1 {
		t.Errorf("expected 1 art state, got %d", len(tb.States))
	}
}

func TestAddTraceStateWithAction(t *testing.T) {
	tb := NewTraceBase(nil)
	tb.AddTraceState(nil) // first state
	tb.LastAction = actions.NewAssumeAction(lg.True)
	tb.AddTraceState(nil) // second state linked by action
	if len(tb.TraceStates) != 2 {
		t.Errorf("expected 2 trace states, got %d", len(tb.TraceStates))
	}
	if tb.LastAction != nil {
		t.Error("LastAction should be nil after AddTraceState")
	}
}

func TestTraceBaseString(t *testing.T) {
	tb := NewTraceBase(nil)
	s := tb.String()
	// Empty trace should produce empty string
	if s != "" {
		t.Errorf("empty trace String() should be empty, got %q", s)
	}
}

func TestTraceBaseClone(t *testing.T) {
	tb := NewTraceBase(nil)
	tb.AddTraceState(nil)
	clone := tb.Clone()
	if clone == tb {
		t.Error("Clone should return a new TraceBase")
	}
	if len(clone.TraceStates) != 0 {
		t.Error("cloned trace should have no states")
	}
}

func TestTraceBaseHandle(t *testing.T) {
	tb := NewTraceBase(nil)
	action := actions.NewAssumeAction(lg.True)
	env := map[string]string{}
	tb.Handle(action, env)
	if tb.LastAction != action {
		t.Error("Handle should set LastAction")
	}
}

func TestTraceBaseFail(t *testing.T) {
	tb := NewTraceBase(nil)
	origAction := actions.NewAssertAction(lg.True)
	tb.LastAction = origAction
	tb.Fail()
	fa, ok := tb.LastAction.(*FailAction)
	if !ok {
		t.Fatal("Fail should wrap action in FailAction")
	}
	if fa.Action != origAction {
		t.Error("FailAction should wrap original action")
	}
}

func TestTraceBaseEnd(t *testing.T) {
	tb := NewTraceBase(nil)
	tb.End()
	// Should have added a final state
	if len(tb.TraceStates) != 1 {
		t.Errorf("End should add final state, got %d states", len(tb.TraceStates))
	}
}

func TestTraceBaseEndWithSub(t *testing.T) {
	tb := NewTraceBase(nil)
	tb.Sub = NewTraceBase(nil)
	tb.End()
	if tb.Sub != nil {
		t.Error("End should clear Sub")
	}
	if tb.Returned == nil {
		t.Error("End should set Returned from Sub")
	}
}

// --- FailAction tests ---

func TestFailActionString(t *testing.T) {
	fa := &FailAction{Action: actions.NewAssertAction(lg.True)}
	s := fa.String()
	if !strings.Contains(s, "FAIL") {
		t.Errorf("FailAction.String() should contain FAIL, got %q", s)
	}
}

func TestFailActionNilAction(t *testing.T) {
	fa := &FailAction{}
	s := fa.String()
	if s != "FAIL" {
		t.Errorf("FailAction with nil action should return 'FAIL', got %q", s)
	}
}

func TestFailActionClone(t *testing.T) {
	orig := &FailAction{Action: actions.NewAssertAction(lg.True)}
	clone := orig.Clone(nil)
	if clone == nil {
		t.Fatal("Clone returned nil")
	}
}

func TestFailActionName(t *testing.T) {
	fa := &FailAction{}
	if fa.Name() != "fail" {
		t.Errorf("expected name 'fail', got %q", fa.Name())
	}
}

// --- Pretty / label helpers ---

func TestPretty(t *testing.T) {
	s := "line1\nline2\nline3\nline4\nline5\nline6"
	result := Pretty(s, 3)
	lines := strings.Split(result, "\n")
	if len(lines) != 4 { // 3 original + "..."
		t.Errorf("expected 4 lines (3 + ...), got %d", len(lines))
	}
}

func TestPrettyShort(t *testing.T) {
	s := "line1\nline2"
	result := Pretty(s, 5)
	if result != s {
		t.Errorf("short string should be unchanged, got %q", result)
	}
}

func TestLabelFromAction(t *testing.T) {
	action := actions.NewAssumeAction(lg.True)
	label := LabelFromAction(action, nil)
	if label == "" {
		t.Error("LabelFromAction should return non-empty string")
	}
}

// --- EvalInState ---

func TestEvalInState(t *testing.T) {
	param := lg.NewConst("X", lg.Boolean)
	val := lg.NewConst("true_val", lg.Boolean)
	eq, _ := lg.NewEq(param, val)
	clauses := clauseops.NewClauses([]lg.Node{eq}, nil, nil)
	state := art.NewState(nil, clauses)
	result := EvalInState(state, param)
	if result == nil {
		t.Fatal("EvalInState should find matching parameter")
	}
	if !result.Equal(val) {
		t.Errorf("expected %v, got %v", val, result)
	}
}

func TestEvalInStateNotFound(t *testing.T) {
	clauses := clauseops.NewClauses([]lg.Node{lg.True}, nil, nil)
	state := art.NewState(nil, clauses)
	param := lg.NewConst("missing", lg.Boolean)
	result := EvalInState(state, param)
	if result != nil {
		t.Error("EvalInState should return nil for missing param")
	}
}

func TestEvalInStateNilClauses(t *testing.T) {
	state := art.NewState(nil, nil)
	param := lg.NewConst("X", lg.Boolean)
	result := EvalInState(state, param)
	if result != nil {
		t.Error("EvalInState should return nil for nil clauses")
	}
}

// --- ValueToStr ---

func TestValueToStrNil(t *testing.T) {
	s := ValueToStr(nil, nil)
	if s != "..." {
		t.Errorf("expected '...', got %q", s)
	}
}

func TestValueToStrConst(t *testing.T) {
	c := lg.NewConst("my_value", lg.Boolean)
	s := ValueToStr(c, nil)
	if s != "my_value" {
		t.Errorf("expected 'my_value', got %q", s)
	}
}

func TestValueToStrNode(t *testing.T) {
	n := lg.True
	s := ValueToStr(n, nil)
	if s == "" {
		t.Error("ValueToStr should return non-empty string for True")
	}
}

// --- Trace (model-based) ---

func TestNewTrace(t *testing.T) {
	clauses := testClauses()
	tr := NewTrace(clauses, nil, nil, true)
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

func TestNewTraceNilClauses(t *testing.T) {
	tr := NewTrace(nil, nil, nil, false)
	if tr == nil {
		t.Fatal("NewTrace returned nil")
	}
	if len(tr.Eqs) != 0 {
		t.Error("Eqs should be empty for nil clauses")
	}
}

// --- MakeCheckArt ---

func TestMakeCheckArt(t *testing.T) {
	mod := testModule()
	ag, pre, _ := MakeCheckArt(mod, "test_action", nil)
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

func TestMakeCheckArtWithPrecond(t *testing.T) {
	mod := testModule()
	precond := []*clauseops.Clauses{testClauses(), testClauses()}
	ag, pre, _ := MakeCheckArt(mod, "test", precond)
	if ag == nil || pre == nil {
		t.Fatal("MakeCheckArt with preconds returned nil")
	}
}

// --- CheckFinalCond ---

func TestCheckFinalCondNilPost(t *testing.T) {
	ag := art.NewAnalysisGraph(nil)
	result := CheckFinalCond(ag, nil, testClauses(), nil, false)
	if result != nil {
		t.Error("CheckFinalCond with nil post should return nil")
	}
}

func TestCheckFinalCondNilFinalCond(t *testing.T) {
	ag := art.NewAnalysisGraph(nil)
	state := art.NewState(nil, testClauses())
	result := CheckFinalCond(ag, state, nil, nil, false)
	if result != nil {
		t.Error("CheckFinalCond with nil final cond should return nil")
	}
}

// --- CheckVC ---

func TestCheckVCNilClauses(t *testing.T) {
	result := CheckVC(nil, nil, nil, nil, false)
	if result != nil {
		t.Error("CheckVC with nil clauses should return nil")
	}
}

func TestCheckVCNoAnnot(t *testing.T) {
	clauses := clauseops.NewClauses([]lg.Node{lg.True}, nil, nil)
	// Annot is nil
	result := CheckVC(clauses, nil, nil, nil, false)
	if result != nil {
		t.Error("CheckVC with nil annot should return nil")
	}
}

// --- MakeVC ---

func TestMakeVC(t *testing.T) {
	action := actions.NewAssumeAction(lg.True)
	pre := []*clauseops.Clauses{testClauses()}
	post := []*clauseops.Clauses{testClauses()}
	vc := MakeVC(action, pre, post, true)
	if vc == nil {
		t.Fatal("MakeVC returned nil")
	}
}

// --- ToLines ---

func TestToLinesEmpty(t *testing.T) {
	tb := NewTraceBase(nil)
	var lines []string
	hash := make(map[string]string)
	tb.ToLines(&lines, hash, 0, func(s string) bool { return false }, false, nil, nil)
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for empty trace, got %d", len(lines))
	}
}

func TestToLinesWithState(t *testing.T) {
	tb := NewTraceBase(nil)
	eq := makeEq("X", "val1")
	tb.AddTraceState([]lg.Node{eq})
	var lines []string
	hash := make(map[string]string)
	tb.ToLines(&lines, hash, 0, func(s string) bool { return false }, false, nil, nil)
	joined := strings.Join(lines, "")
	if !strings.Contains(joined, "[") {
		t.Error("ToLines should contain state bracket for detailed mode")
	}
}

func TestToLinesHidden(t *testing.T) {
	tb := NewTraceBase(nil)
	eq := makeEq("hidden_sym", "val")
	tb.AddTraceState([]lg.Node{eq})
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

func TestToLinesLoopStart(t *testing.T) {
	tb := NewTraceBase(nil)
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

func TestIsCallOrEnv(t *testing.T) {
	if isCallOrEnv(nil) {
		t.Error("nil should not be call or env")
	}
	assume := actions.NewAssumeAction(lg.True)
	if isCallOrEnv(assume) {
		t.Error("AssumeAction should not be call or env")
	}
}

func TestIsCallAction(t *testing.T) {
	if isCallAction(nil) {
		t.Error("nil should not be call action")
	}
	assume := actions.NewAssumeAction(lg.True)
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
		c := lg.NewConst(name, lg.Boolean)
		s := ValueToStr(c, nil)
		if s != name {
			t.Errorf("ValueToStr(Const(%q)) = %q, want %q", name, s, name)
		}
	})
}
