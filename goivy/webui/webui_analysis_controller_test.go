package webui

import (
	"testing"

	goivy "github.com/glycerine/ivy/goivy"
)

func TestAnalysisControllerOptionsPersistOnSession(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "analysis-controller-options")
	s.CTIUI = NewCTIAnalysisGraphUI(nil)

	_ = s.RunCheckWithOptions("induction", CheckOptions{
		Abstractor:          "ta.Abstractors.propagate",
		Bound:               15,
		RelationsToMinimize: "link semaphore",
		TransitionLogFile:   "client_server.log",
	})

	if s.SelectedAbstractor != "ta.Abstractors.propagate" {
		t.Fatalf("SelectedAbstractor = %q", s.SelectedAbstractor)
	}
	if s.TransitionLogFile != "client_server.log" {
		t.Fatalf("TransitionLogFile = %q", s.TransitionLogFile)
	}
	if s.CTIUI.RelationsToMinimize != "link semaphore" {
		t.Fatalf("RelationsToMinimize = %q", s.CTIUI.RelationsToMinimize)
	}
	if s.CTIUI.CurrentBound != 15 {
		t.Fatalf("CurrentBound = %d", s.CTIUI.CurrentBound)
	}

	_, _ = s.ArgNodeAction("state_0", "recalculate", map[string]interface{}{
		"abstractor":            "ta.Abstractors.concept_space",
		"bound":                 float64(5),
		"relations_to_minimize": "semaphore",
		"transition_log_file":   "trace.log",
	})

	if s.SelectedAbstractor != "ta.Abstractors.concept_space" {
		t.Fatalf("SelectedAbstractor after action = %q", s.SelectedAbstractor)
	}
	if s.TransitionLogFile != "trace.log" {
		t.Fatalf("TransitionLogFile after action = %q", s.TransitionLogFile)
	}
	if s.CTIUI.RelationsToMinimize != "semaphore" {
		t.Fatalf("RelationsToMinimize after action = %q", s.CTIUI.RelationsToMinimize)
	}
	if s.CTIUI.CurrentBound != 5 {
		t.Fatalf("CurrentBound after action = %d", s.CTIUI.CurrentBound)
	}
}
