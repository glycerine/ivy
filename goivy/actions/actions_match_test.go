package actions

import (
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
)

type recordingAnnotationHandler struct {
	labels []string
}

func (h *recordingAnnotationHandler) Eval(cond lg.Expr) bool { return true }

func (h *recordingAnnotationHandler) Handle(action ActionsAction, env map[lg.NodeKey]lg.Expr) {
	label := ""
	if labeled, ok := action.(interface{ GetLabel() string }); ok {
		label = labeled.GetLabel()
	}
	h.labels = append(h.labels, label)
}

func (h *recordingAnnotationHandler) DoReturn(action ActionsAction, env map[lg.NodeKey]lg.Expr) {}
func (h *recordingAnnotationHandler) Fail()                                                     {}

func TestMatchAnnotationEnvActionUsesChosenBranchLabel(t *testing.T) {
	branch := NewSequence(NewAssumeAction(lg.True), NewReturnAction())
	branch.SetLabel("ext")
	env := NewEnvActionOn(NewActionsConfig(), branch)

	annot := &IteAnnotation{
		Cond:  lg.NewConst("branch", lg.Boolean),
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
