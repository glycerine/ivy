package goivy

import (
	"encoding/json"
	"strings"
	"testing"
)

type recordingAnnotationHandler struct {
	labels  []string
	records []matchHandleRecord
}

type matchHandleRecord struct {
	Type       string `json:"type"`
	Label      string `json:"label"`
	Argc       int    `json:"argc"`
	ChildType  string `json:"child_type"`
	ChildLabel string `json:"child_label"`
}

func (h *recordingAnnotationHandler) Eval(cond Expr) bool { return true }

func (h *recordingAnnotationHandler) Handle(action ActionsAction, env map[NodeKey]Expr) {
	label := ""
	if labeled, ok := action.(interface{ GetLabel() string }); ok {
		label = labeled.GetLabel()
	}
	h.labels = append(h.labels, label)

	record := matchHandleRecord{Type: pythonMatchActionName(action), Label: label}
	args := action.ActionArgs()
	record.Argc = len(args)
	if len(args) > 0 {
		if child := extractActionFromNode(args[0]); child != nil {
			record.ChildType = pythonMatchActionName(child)
			if labeled, ok := child.(interface{ GetLabel() string }); ok {
				record.ChildLabel = labeled.GetLabel()
			}
		}
	}
	h.records = append(h.records, record)
}

func (h *recordingAnnotationHandler) DoReturn(action ActionsAction, env map[NodeKey]Expr) {}
func (h *recordingAnnotationHandler) Fail()                                               {}

func TestMatchAnnotationEnvActionUsesChosenBranchLabel(t *testing.T) {
	branch := NewSequence(NewAssumeAction(True), NewReturnAction())
	branch.SetLabel("ext")
	env := NewEnvActionOn(NewActionsConfig(), branch)

	annot := &IteAnnotation{
		Cond:  NewConst("branch", Boolean),
		ThenB: EmptyAnnotation{},
	}
	handler := &recordingAnnotationHandler{}

	MatchAnnotation(env, annot, handler, nil)

	if len(handler.labels) == 0 {
		t.Fatalf("MatchAnnotation did not handle the chosen EnvAction branch")
	}
	if got := handler.labels[0]; got != "call ext" {
		t.Fatalf("first handled label = %q, want %q", got, "call ext")
	}
}

func TestMatchAnnotationUnlabeledEnvActionWrapsChosenBranchLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_logic as il, logic as lg

p = il.Symbol("p", lg.Boolean)
branch = act.Sequence(act.AssumeAction(p), act.ReturnAction())
branch.label = "ext"
env = act.EnvAction(branch)
cond = il.Symbol("branch", lg.Boolean)

class H(object):
    def __init__(self):
        self.records = []
    def eval(self, cond):
        return True
    def handle(self, action, env):
        args = list(getattr(action, "args", []))
        child = args[0] if args else None
        self.records.append({
            "type": type(action).__name__,
            "label": getattr(action, "label", ""),
            "argc": len(args),
            "child_type": type(child).__name__ if child is not None else "",
            "child_label": getattr(child, "label", "") if child is not None else "",
        })
    def do_return(self, action, env):
        pass
    def fail(self):
        pass

h = H()
annot = act.IteAnnotation(cond, act.EmptyAnnotation(), act.EmptyAnnotation())
act.match_annotation(env, annot, h)
print(json.dumps(h.records))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python MatchAnnotation EnvAction oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []matchHandleRecord
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python MatchAnnotation EnvAction oracle %q: %v", out, err)
	}

	branch := NewSequence(NewAssumeAction(True), NewReturnAction())
	branch.SetLabel("ext")
	env := NewEnvActionOn(NewActionsConfig(), branch)
	annot := &IteAnnotation{
		Cond:  NewConst("branch", Boolean),
		ThenB: EmptyAnnotation{},
	}
	handler := &recordingAnnotationHandler{}

	MatchAnnotation(env, annot, handler, nil)

	if len(handler.records) == 0 {
		t.Fatalf("MatchAnnotation did not handle the chosen EnvAction branch")
	}
	if got := handler.records[:1]; strings.Join(recordsToComparableJSON(t, got), "\n") != strings.Join(recordsToComparableJSON(t, want[:1]), "\n") {
		t.Fatalf("Go MatchAnnotation EnvAction record differs from Python\nwant: %#v\ngot:  %#v", want[:1], got)
	}
}

func recordsToComparableJSON(t *testing.T, records []matchHandleRecord) []string {
	t.Helper()
	result := make([]string, len(records))
	for i, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("marshal record: %v", err)
		}
		result[i] = string(data)
	}
	return result
}

type eventRecordingAnnotationHandler struct {
	events []string
}

func (h *eventRecordingAnnotationHandler) Eval(cond Expr) bool { return true }

func (h *eventRecordingAnnotationHandler) Handle(action ActionsAction, env map[NodeKey]Expr) {
	h.events = append(h.events, "handle:"+pythonMatchActionName(action))
}

func (h *eventRecordingAnnotationHandler) DoReturn(action ActionsAction, env map[NodeKey]Expr) {
	h.events = append(h.events, "return:"+pythonMatchActionName(action))
}

func (h *eventRecordingAnnotationHandler) Fail() {
	h.events = append(h.events, "fail")
}

func pythonMatchActionName(action ActionsAction) string {
	switch action.(type) {
	case *LogicAssumeAction:
		return "AssumeAction"
	case *FailAction:
		return "fail_action"
	default:
		return ActionTypeName(action)
	}
}

func TestMatchAnnotationFailActionRecursesThenFailsLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_interp as itp, ivy_logic as il, logic as lg

p = il.Symbol("p", lg.Boolean)
class H(object):
    def __init__(self):
        self.events = []
    def eval(self, cond):
        return True
    def handle(self, action, env):
        self.events.append("handle:" + type(action).__name__)
    def do_return(self, action, env):
        self.events.append("return:" + type(action).__name__)
    def fail(self):
        self.events.append("fail")

h = H()
act.match_annotation(itp.fail_action(act.AssumeAction(p)), act.EmptyAnnotation(), h)
print(json.dumps(h.events))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python MatchAnnotation fail-action oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python MatchAnnotation fail-action oracle %q: %v", out, err)
	}

	p := NewConst("p", Boolean)
	handler := &eventRecordingAnnotationHandler{}
	MatchAnnotation(NewFailAction(NewAssumeAction(p)), EmptyAnnotation{}, handler, nil)
	if strings.Join(handler.events, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Go MatchAnnotation fail-action events differ from Python\nwant: %#v\ngot:  %#v", want, handler.events)
	}
}

func TestMatchAnnotationSomeIfWrapsBranchLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_ast as ia, ivy_logic as il, logic as lg

p = il.Symbol("p", lg.Boolean)
x = il.Variable("X", lg.Boolean)
branch = il.Symbol("branch", lg.Boolean)

class H(object):
    def __init__(self):
        self.events = []
    def eval(self, cond):
        return True
    def handle(self, action, env):
        self.events.append("handle:" + type(action).__name__)
    def do_return(self, action, env):
        self.events.append("return:" + type(action).__name__)
    def fail(self):
        self.events.append("fail")

h = H()
some = ia.Some(x, p)
action = act.IfAction(some, act.AssumeAction(p))
annot = act.IteAnnotation(
    branch,
    act.ComposeAnnotation(
        act.ComposeAnnotation(act.EmptyAnnotation(), act.EmptyAnnotation()),
        act.EmptyAnnotation(),
    ),
    act.EmptyAnnotation(),
)
act.match_annotation(action, annot, h)
print(json.dumps(h.events))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python MatchAnnotation Some-if oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python MatchAnnotation Some-if oracle %q: %v", out, err)
	}

	p := NewConst("p", Boolean)
	x := NewConst("X", Boolean)
	branch := NewConst("branch", Boolean)
	some := &SomeCondition{Params: []*Const{x}, Fmla: p, Kind: "some"}
	action := NewIfAction(some, NewAssumeAction(p))
	annot := &IteAnnotation{
		Cond: branch,
		ThenB: &ComposeAnnotation{Args: []Annotation{
			&ComposeAnnotation{Args: []Annotation{EmptyAnnotation{}, EmptyAnnotation{}}},
			EmptyAnnotation{},
		}},
		ElseB: EmptyAnnotation{},
	}

	handler := &eventRecordingAnnotationHandler{}
	MatchAnnotation(action, annot, handler, nil)
	if strings.Join(handler.events, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Go MatchAnnotation Some-if events differ from Python\nwant: %#v\ngot:  %#v", want, handler.events)
	}
}

func TestTraceExpandWhileRankingChecksUseDecreasesLinenoLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_logic as il, logic as lg, ivy_module as im

cond = il.Symbol("cond", lg.Boolean)
rank = il.Symbol("rank_value", lg.Boolean)
ranking = act.Ranking(rank)
ranking.lineno = 123
while_action = act.WhileAction(cond, act.Sequence(), ranking)
while_action.lineno = 99
expanded = while_action.expand(im.Module(), [])

records = []
def walk(action):
    if isinstance(action, (act.AssumeAction, act.AssertAction)):
        lineno = getattr(action, "lineno", None)
        if lineno is not None:
            records.append({"type": type(action).__name__, "line": lineno})
    for child in getattr(action, "args", []):
        if isinstance(child, act.Action):
            walk(child)

walk(expanded)
print(json.dumps(records))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python WhileAction.expand oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want []lineRecord
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python WhileAction.expand oracle %q: %v", out, err)
	}

	cond := NewConst("cond", Boolean)
	rankExpr := NewConst("rank_value", Boolean)
	ranking := NewRanking(nil, rankExpr)
	ranking.SetLineno(Location{Filename: "test.ivy", Line: 123})
	whileAction := NewWhileAction(cond, NewSequence(), &RankingWrapper{Ranking: ranking})
	whileAction.SetLineno(Location{Filename: "test.ivy", Line: 99})

	expanded := expandWhile(whileAction, New())
	got := collectGeneratedRankingLineRecords(expanded)

	if strings.Join(lineRecordsToComparableJSON(t, got), "\n") != strings.Join(lineRecordsToComparableJSON(t, want), "\n") {
		t.Fatalf("Go trace expandWhile generated ranking line records differ from Python\nwant: %#v\ngot:  %#v", want, got)
	}
}

type lineRecord struct {
	Type string `json:"type"`
	Line int    `json:"line"`
}

func collectGeneratedRankingLineRecords(action ActionsAction) []lineRecord {
	var records []lineRecord
	actionsWalkActions(action, func(a ActionsAction) {
		switch act := a.(type) {
		case *LogicAssumeAction:
			if eq, ok := act.Formula.(*Eq); ok {
				if c, ok := eq.T1.(*Const); ok && c.Name == "$rank" && act.HasLineno() {
					records = append(records, lineRecord{Type: "AssumeAction", Line: act.GetLineno().Line})
				}
			}
		case *LogicAssertAction:
			if act.HasLineno() {
				records = append(records, lineRecord{Type: "AssertAction", Line: act.GetLineno().Line})
			}
		}
	})
	return records
}

func lineRecordsToComparableJSON(t *testing.T, records []lineRecord) []string {
	t.Helper()
	result := make([]string, len(records))
	for i, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("marshal line record: %v", err)
		}
		result[i] = string(data)
	}
	return result
}
