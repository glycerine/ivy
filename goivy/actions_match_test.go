package goivy

import (
	"testing"
)

type recordingAnnotationHandler struct {
	labels []string
}

func (h *recordingAnnotationHandler) Eval(cond Expr) bool { return true }

func (h *recordingAnnotationHandler) Handle(action ActionsAction, env map[NodeKey]Expr) {
	label := ""
	if labeled, ok := action.(interface{ GetLabel() string }); ok {
		label = labeled.GetLabel()
	}
	h.labels = append(h.labels, label)
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
