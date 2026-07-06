package webui

import (
	"testing"

	goivy "github.com/glycerine/ivy/goivy"
)

func menuLabelsForTest(menus []MenuDef) map[string]bool {
	labels := make(map[string]bool, len(menus))
	for _, menu := range menus {
		labels[menu.Label] = true
	}
	return labels
}

func requireMenuLabelForTest(t *testing.T, menus []MenuDef, label string) {
	t.Helper()
	if !menuLabelsForTest(menus)[label] {
		t.Fatalf("missing menu label %q in %#v", label, menus)
	}
}

func rejectMenuLabelForTest(t *testing.T, menus []MenuDef, label string) {
	t.Helper()
	if menuLabelsForTest(menus)[label] {
		t.Fatalf("unexpected menu label %q in %#v", label, menus)
	}
}

func TestSessionMenuDescriptorsFollowActiveSheetAndMode(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "menu-context")
	if err := s.LoadFileContent("test.ivy", []byte(ivySample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}

	reach := BuildBrowserMenuDescriptorsForSession(s, MenuRequest{
		SheetID: rootSheetID,
		UIMode:  "reachability",
	})
	requireMenuLabelForTest(t, reach.Arg, "File")
	requireMenuLabelForTest(t, reach.Arg, "Action")
	rejectMenuLabelForTest(t, reach.Arg, "Invariant")
	rejectMenuLabelForTest(t, reach.Concept, "Conjecture")

	cti := BuildBrowserMenuDescriptorsForSession(s, MenuRequest{
		SheetID: rootSheetID,
		UIMode:  "cti",
	})
	requireMenuLabelForTest(t, cti.Arg, "File")
	requireMenuLabelForTest(t, cti.Arg, "Invariant")
	rejectMenuLabelForTest(t, cti.Arg, "Action")
	requireMenuLabelForTest(t, cti.Concept, "Conjecture")

	eventSheetID := s.EventViewer.NewSheet([]*TraceEvent{NewTraceEvent("execute_action(foo,bar)")})
	events := BuildBrowserMenuDescriptorsForSession(s, MenuRequest{
		SheetID: eventSheetID,
		UIMode:  "reachability",
	})
	requireMenuLabelForTest(t, events.Arg, "File")
	rejectMenuLabelForTest(t, events.Arg, "Action")
	rejectMenuLabelForTest(t, events.Arg, "Invariant")
	if len(events.Concept) != 0 {
		t.Fatalf("event sheet concept menus = %#v, want none", events.Concept)
	}
}
