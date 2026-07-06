package webui

import (
	"testing"

	goivy "github.com/glycerine/ivy/goivy"
)

func TestExtensionArgNodeActionDescriptorDispatchesThroughBrowserAPI(t *testing.T) {
	ag := goivy.NewAnalysisGraph(nil)
	ag.States = append(ag.States, &goivy.State{ID: 0})

	ui := NewAnalysisGraphUI()
	ui.AG = ag
	ui.ExtensionConfig = NewExtConfig()
	ui.ExtensionConfig.ArgNodeActions = NewExtensionPoint("arg_node_actions")

	called := false
	ui.ExtensionConfig.ArgNodeActions.Action("fake extension", func(args ...interface{}) error {
		called = true
		if len(args) < 2 {
			t.Fatalf("extension args = %#v, want node id and selections", args)
		}
		if args[0] != 0 {
			t.Fatalf("node id arg = %#v, want 0", args[0])
		}
		selection, _ := args[1].([]string)
		if len(selection) != 1 || selection[0] != "fact-a" {
			t.Fatalf("selection arg = %#v, want fact-a", args[1])
		}
		ag.States[0].Label = "mutated by extension"
		return nil
	})

	entries := ui.GetNodeActions(0, "right")
	var found *ActionEntry
	for i := range entries {
		if entries[i].Label == "fake extension" {
			found = &entries[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("fake extension action not found in %#v", entries)
	}
	if found.Action != "extension_arg_node" {
		t.Fatalf("extension action id = %q", found.Action)
	}

	s := NewSession(goivy.NewConfig(), "extension-browser-api")
	s.AG = ag
	s.AGUI = ui
	s.SheetUIs = map[string]*AnalysisGraphUI{rootSheetID: ui}

	result, err := s.ArgNodeAction("state_0", found.Action, map[string]interface{}{
		"sheet_id":        rootSheetID,
		"extension_label": "fake extension",
		"selection":       []interface{}{"fact-a"},
	})
	if err != nil {
		t.Fatalf("ArgNodeAction extension: %v", err)
	}
	if !called {
		t.Fatal("extension callback was not called")
	}
	if result["extension"] != "fake extension" {
		t.Fatalf("result extension = %#v", result["extension"])
	}
	if ag.States[0].Label != "mutated by extension" {
		t.Fatalf("state label = %q", ag.States[0].Label)
	}
}
