//go:build web

package webui

import (
	"strings"
	"testing"

	iu "github.com/glycerine/goivy/ivyutils"
)

// --- AnalysisGraphUI tests ---

func TestNewAnalysisGraphUI(t *testing.T) {
	ui := NewAnalysisGraphUI()
	if ui == nil {
		t.Fatal("NewAnalysisGraphUI returned nil")
	}
	if ui.Mode != DefaultMode {
		t.Errorf("expected default mode %s, got %s", DefaultMode, ui.Mode)
	}
	if ui.G == nil {
		t.Error("G is nil")
	}
}

func TestAnalysisGraphUISetMode(t *testing.T) {
	ui := NewAnalysisGraphUI()
	for _, mode := range AllModes {
		ui.SetMode(mode)
		if ui.GetMode() != mode {
			t.Errorf("expected mode %s, got %s", mode, ui.GetMode())
		}
	}
}

func TestAnalysisGraphUIMenus(t *testing.T) {
	ui := NewAnalysisGraphUI()
	menus := ui.Menus()
	if len(menus) != 3 {
		t.Fatalf("expected 3 menus, got %d", len(menus))
	}
	labels := []string{"File", "Mode", "Action"}
	for i, m := range menus {
		if m.Label != labels[i] {
			t.Errorf("menu %d: expected %q, got %q", i, labels[i], m.Label)
		}
	}
}

func TestAnalysisGraphUIStart(t *testing.T) {
	ui := NewAnalysisGraphUI()
	ui.Start()
	if len(ui.G.States) == 0 {
		t.Error("Start should create initial state")
	}
}

func TestAnalysisGraphUINodeColor(t *testing.T) {
	ui := NewAnalysisGraphUI()
	ref := &ARGStateRef{IsSafe: true}
	if ui.NodeColor(ref) != "green" {
		t.Error("safe node should be green")
	}
	ref.IsSafe = false
	if ui.NodeColor(ref) != "black" {
		t.Error("unsafe node should be black")
	}
	if ui.NodeColor(nil) != "black" {
		t.Error("nil node should be black")
	}
}

func TestAnalysisGraphUIGetNodeActions(t *testing.T) {
	ui := NewAnalysisGraphUI()
	actions := ui.GetNodeActions(0, "left")
	if len(actions) != 1 || actions[0].Action != "view_state" {
		t.Error("left click should return view_state action")
	}
	actions = ui.GetNodeActions(0, "right")
	if len(actions) == 0 {
		t.Error("right click should return actions")
	}
}

func TestAnalysisGraphUIGetEdgeActions(t *testing.T) {
	ui := NewAnalysisGraphUI()
	actions := ui.GetEdgeActions("left")
	if actions != nil {
		t.Error("left click on edge should return nil")
	}
	actions = ui.GetEdgeActions("right")
	if len(actions) != 4 {
		t.Errorf("right click should return 4 actions, got %d", len(actions))
	}
}

func TestAnalysisGraphUINodeCommands(t *testing.T) {
	ui := NewAnalysisGraphUI()
	cmds := ui.NodeCommands()
	if len(cmds) != 8 {
		t.Errorf("expected 8 node commands, got %d", len(cmds))
	}
	// Check that expected commands are present.
	found := false
	for _, c := range cmds {
		if c.Action == "check_safety" {
			found = true
		}
	}
	if !found {
		t.Error("check_safety action not found")
	}
}

func TestAnalysisGraphUIMarkNode(t *testing.T) {
	ui := NewAnalysisGraphUI()
	ref := &ARGStateRef{ID: 42}
	ui.MarkNode(ref)
	mark := ui.GetMark()
	if mark == nil || mark.ID != 42 {
		t.Error("MarkNode/GetMark failed")
	}
}

func TestAnalysisGraphUIRememberGraph(t *testing.T) {
	ui := NewAnalysisGraphUI()
	g := NewGraph([]string{"s"}, nil)
	ui.RememberGraph("test_goal", g)
	names := ui.RememberedGraphNames()
	if len(names) != 1 || names[0] != "test_goal" {
		t.Errorf("expected [test_goal], got %v", names)
	}
}

func TestAnalysisGraphUIDeleteNode(t *testing.T) {
	ui := NewAnalysisGraphUI()
	ui.G.States = []ARGNode{{ID: 0}, {ID: 1}, {ID: 2}}
	ui.DeleteNode(1)
	if len(ui.G.States) != 2 {
		t.Errorf("expected 2 states after delete, got %d", len(ui.G.States))
	}
}

func TestAnalysisGraphUICheckSafety(t *testing.T) {
	ui := NewAnalysisGraphUI()
	safe, msg := ui.CheckSafetyNode(0)
	if !safe {
		t.Errorf("stub should return safe, got msg: %s", msg)
	}
}

func TestStateEquationLabel(t *testing.T) {
	if StateEquationLabel("act", "trans") != "trans -> act" {
		t.Error("unexpected label format")
	}
	if StateEquationLabel("", "trans") != "trans" {
		t.Error("empty action should return just transLabel")
	}
}

// --- GraphWidget tests ---

func TestGraphWidgetMenus(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	menus := w.Menus()
	if len(menus) != 2 {
		t.Fatalf("expected 2 menus, got %d", len(menus))
	}
	if menus[0].Label != "Action" {
		t.Errorf("expected Action menu, got %q", menus[0].Label)
	}
}

func TestGraphWidgetUndoRedo(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	w.Checkpoint(false)
	w.G().State = "changed"
	w.Undo()
	if w.G().State == "changed" {
		t.Error("undo should revert state")
	}
	w.Redo()
	if w.G().State != "changed" {
		t.Error("redo should restore state")
	}
}

func TestGraphWidgetBacktrack(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	w.Checkpoint(true) // backtrack point
	w.G().State = "s1"
	w.Checkpoint(false)
	w.G().State = "s2"
	w.Backtrack()
	if w.G().State == "s2" {
		t.Error("backtrack should revert past s2")
	}
}

func TestGraphWidgetGetNodeActions(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	actions := w.GetNodeActions("n1", "left")
	if len(actions) == 0 {
		t.Error("expected node actions for left click")
	}
	actions = w.GetNodeActions("n1", "right")
	if actions != nil {
		t.Error("expected nil for right click")
	}
}

func TestGraphWidgetGetEdgeActions(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	actions := w.GetEdgeActions("left")
	if len(actions) != 3 {
		t.Errorf("expected 3 edge actions, got %d", len(actions))
	}
}

func TestGraphWidgetSelectNode(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	w.SelectNode("n1", true)
	if !w.NodeSelection["n1"] {
		t.Error("n1 should be selected")
	}
	w.SelectNode("n1", false)
	if w.NodeSelection["n1"] {
		t.Error("n1 should be deselected")
	}
}

func TestGraphWidgetClearSelection(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	w.SelectNode("n1", true)
	w.SelectEdge("e1", true)
	w.ClearElemSelection()
	if len(w.NodeSelection) != 0 || len(w.EdgeSelection) != 0 {
		t.Error("selections should be cleared")
	}
}

func TestGraphWidgetShowRelation(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	c := &Concept{Name: "rel1"}
	w.ShowRelation(c, "+", true, false)
	if !w.G().Checks.EdgeVisible("rel1", EdgeDisplayAllToAll) {
		t.Error("rel1 all_to_all should be visible after ShowRelation '+'")
	}
	w.ShowRelation(c, "-", true, false)
	if !w.G().Checks.EdgeVisible("rel1", EdgeDisplayNoneToNone) {
		t.Error("rel1 none_to_none should be visible after ShowRelation '-'")
	}
	w.ShowRelation(c, "T", true, false)
	if !w.G().Checks.EdgeVisible("rel1", EdgeDisplayTransitive) {
		t.Error("rel1 transitive should be visible after ShowRelation 'T'")
	}
}

func TestGraphWidgetClearEdges(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	w.G().Checks.SetEdgeCheckbox("r1", EdgeDisplayAllToAll, true)
	w.ClearEdges()
	if w.G().Checks.EdgeVisible("r1", EdgeDisplayAllToAll) {
		t.Error("r1 should not be visible after ClearEdges")
	}
}

func TestGraphWidgetApplyStructureRenaming(t *testing.T) {
	gs := StandardGraph([]string{"s"}, nil)
	w := NewGraphWidget(gs)
	w.StructureRenaming["old_name"] = "new_name"
	result := w.ApplyStructureRenaming("the old_name is here")
	if result != "the new_name is here" {
		t.Errorf("expected renaming, got %q", result)
	}
}

// --- ExtensionPoint tests ---

func TestExtensionPointRegisterAndInvoke(t *testing.T) {
	ep := NewExtensionPoint("test")
	called := false
	ep.Register(func(ctx interface{}, args ...interface{}) ([]ExtensionAction, error) {
		called = true
		return []ExtensionAction{{Label: "test_action"}}, nil
	})
	if ep.Len() != 1 {
		t.Errorf("expected 1 callback, got %d", ep.Len())
	}
	actions, errs := ep.Invoke(nil)
	if !called {
		t.Error("callback was not called")
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}
	if len(actions) != 1 || actions[0].Label != "test_action" {
		t.Error("unexpected action result")
	}
}

func TestExtensionPointAction(t *testing.T) {
	ep := NewExtensionPoint("test")
	actionCalled := false
	ep.Action("my action", func(args ...interface{}) error {
		actionCalled = true
		return nil
	})
	actions, _ := ep.Invoke(nil)
	if len(actions) != 1 || actions[0].Label != "my action" {
		t.Error("Action registration failed")
	}
	if actions[0].Callback != nil {
		_ = actions[0].Callback()
		if !actionCalled {
			t.Error("action callback not invoked")
		}
	}
}

func TestExtensionPointUnregister(t *testing.T) {
	ep := NewExtensionPoint("test")
	ep.Register(func(ctx interface{}, args ...interface{}) ([]ExtensionAction, error) {
		return nil, nil
	})
	if ep.Len() != 1 {
		t.Fatal("expected 1 callback")
	}
	err := ep.Unregister(0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ep.Len() != 0 {
		t.Error("expected 0 callbacks after unregister")
	}
}

func TestExtensionPointUnregisterOutOfRange(t *testing.T) {
	ep := NewExtensionPoint("test")
	if err := ep.Unregister(5); err == nil {
		t.Error("expected error for out of range index")
	}
}

func TestArgNodeActionsInitRegistered(t *testing.T) {
	// The init() function should have registered 2 default callbacks.
	if ArgNodeActions.Len() < 2 {
		t.Errorf("expected at least 2 default arg_node_actions callbacks, got %d", ArgNodeActions.Len())
	}
}

func TestGoalNodeActionsEmpty(t *testing.T) {
	if GoalNodeActions.Len() != 0 {
		t.Errorf("expected 0 goal_node_actions callbacks, got %d", GoalNodeActions.Len())
	}
}

// --- CTI UI tests ---

func TestNewCTIAnalysisGraphUI(t *testing.T) {
	ui := NewCTIAnalysisGraphUI()
	if ui == nil {
		t.Fatal("NewCTIAnalysisGraphUI returned nil")
	}
	if ui.HaveCTI {
		t.Error("should not have CTI initially")
	}
}

func TestCTIMenus(t *testing.T) {
	ui := NewCTIAnalysisGraphUI()
	menus := ui.CTIMenus()
	if len(menus) != 2 {
		t.Fatalf("expected 2 CTI menus, got %d", len(menus))
	}
	if menus[1].Label != "Invariant" {
		t.Errorf("expected Invariant menu, got %q", menus[1].Label)
	}
}

func TestCTIWeaken(t *testing.T) {
	ui := NewCTIAnalysisGraphUI()
	ui.Conjectures = []string{"conj1", "conj2", "conj3"}
	removed, err := ui.Weaken([]int{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(removed) != 1 || removed[0] != "conj2" {
		t.Errorf("expected [conj2] removed, got %v", removed)
	}
	if len(ui.Conjectures) != 2 {
		t.Errorf("expected 2 conjectures remaining, got %d", len(ui.Conjectures))
	}
}

func TestCTIWeakenEmpty(t *testing.T) {
	ui := NewCTIAnalysisGraphUI()
	_, err := ui.Weaken(nil)
	if err == nil {
		t.Error("expected error for empty indices")
	}
}

func TestCTISaveConjectures(t *testing.T) {
	ui := NewCTIAnalysisGraphUI()
	ui.Conjectures = []string{"p(X)", "q(X,Y)"}
	result := ui.SaveConjectures()
	if !strings.Contains(result, "invariant p(X)") {
		t.Error("save should contain 'invariant p(X)'")
	}
	if !strings.Contains(result, "invariant q(X,Y)") {
		t.Error("save should contain 'invariant q(X,Y)'")
	}
}

func TestWriteConjecture(t *testing.T) {
	s := WriteConjecture("mylab", "p(X)")
	if s != "invariant [mylab] p(X)\n" {
		t.Errorf("unexpected: %q", s)
	}
	s = WriteConjecture("", "p(X)")
	if s != "invariant p(X)\n" {
		t.Errorf("unexpected: %q", s)
	}
}

func TestConceptGraphUIMenus(t *testing.T) {
	menus := ConceptGraphUIMenus()
	if len(menus) != 2 {
		t.Fatalf("expected 2 menus, got %d", len(menus))
	}
	if menus[0].Label != "Conjecture" {
		t.Errorf("expected Conjecture menu, got %q", menus[0].Label)
	}
}

// --- EventTraceViewer tests ---

func TestEventTraceViewerNewSheet(t *testing.T) {
	v := NewEventTraceViewer()
	evs := []*TraceEvent{
		NewTraceEvent("event1"),
		NewTraceEvent("event2"),
	}
	name := v.NewSheet(evs)
	if name == "" {
		t.Error("NewSheet returned empty name")
	}
	sheet := v.GetSheet(name)
	if sheet == nil {
		t.Fatal("sheet not found")
	}
	if len(sheet.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(sheet.Events))
	}
}

func TestEventTraceViewerLookup(t *testing.T) {
	evs := []*TraceEvent{
		NewTraceEvent("top"),
	}
	sub := NewTraceEvent("sub")
	evs[0].AddSub(sub)
	subsub := NewTraceEvent("subsub")
	sub.AddSub(subsub)

	found := LookupEvent(evs, "0")
	if found == nil || found.Text != "top" {
		t.Error("lookup 0 failed")
	}
	found = LookupEvent(evs, "0/0")
	if found == nil || found.Text != "sub" {
		t.Error("lookup 0/0 failed")
	}
	found = LookupEvent(evs, "0/0/0")
	if found == nil || found.Text != "subsub" {
		t.Error("lookup 0/0/0 failed")
	}
	found = LookupEvent(evs, "99")
	if found != nil {
		t.Error("lookup 99 should return nil")
	}
}

func TestFilterEvents(t *testing.T) {
	evs := []*TraceEvent{
		NewTraceEvent("hello world"),
		NewTraceEvent("foo bar"),
	}
	evs[0].AddSub(NewTraceEvent("nested hello"))
	result := FilterEvents(evs, "hello")
	if len(result) != 2 {
		t.Errorf("expected 2 matching events, got %d", len(result))
	}
}

func TestFindEvent(t *testing.T) {
	evs := []*TraceEvent{
		NewTraceEvent("first"),
		NewTraceEvent("target"),
		NewTraceEvent("last"),
	}
	ev, addr := FindEvent(evs, "target", false)
	if ev == nil || ev.Text != "target" {
		t.Error("FindEvent forward failed")
	}
	if addr != "1" {
		t.Errorf("expected addr '1', got %q", addr)
	}

	ev, addr = FindEvent(evs, "target", true)
	if ev == nil || ev.Text != "target" {
		t.Error("FindEvent reverse failed")
	}
}

func TestFormatEvent(t *testing.T) {
	if FormatEvent(nil) != "" {
		t.Error("FormatEvent(nil) should be empty")
	}
	ev := NewTraceEvent("hello")
	if FormatEvent(ev) != "hello" {
		t.Error("FormatEvent returned wrong text")
	}
}

func TestFormatTrace(t *testing.T) {
	evs := []*TraceEvent{
		NewTraceEvent("top"),
	}
	evs[0].AddSub(NewTraceEvent("sub"))
	result := FormatTrace(evs)
	if !strings.Contains(result, "top") || !strings.Contains(result, "  sub") {
		t.Errorf("FormatTrace unexpected output: %q", result)
	}
}

func TestEventTraceViewerPatterns(t *testing.T) {
	v := NewEventTraceViewer()
	v.AddPattern("pat1")
	v.AddPattern("pat2")
	if len(v.Patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(v.Patterns))
	}
	saved := v.SavePatterns()
	if !strings.Contains(saved, "pat1") {
		t.Error("saved patterns should contain pat1")
	}

	v2 := NewEventTraceViewer()
	v2.LoadPatterns(saved)
	if len(v2.Patterns) != 2 {
		t.Errorf("loaded %d patterns, expected 2", len(v2.Patterns))
	}

	v.RemovePattern(0)
	if len(v.Patterns) != 1 || v.Patterns[0] != "pat2" {
		t.Error("RemovePattern failed")
	}

	v.ClearPatterns()
	if len(v.Patterns) != 0 {
		t.Error("ClearPatterns failed")
	}
}

// --- UI Util tests ---

func TestCenterWindow(t *testing.T) {
	x, y := CenterWindow(1920, 1080, 800, 600)
	if x != 560 || y != 240 {
		t.Errorf("expected (560, 240), got (%d, %d)", x, y)
	}
}

func TestCenterWindowOnWindow(t *testing.T) {
	x, y := CenterWindowOnWindow(100, 100, 800, 600, 400, 300)
	if x != 300 || y != 250 {
		t.Errorf("expected (300, 250), got (%d, %d)", x, y)
	}
}

func TestConvertToInt(t *testing.T) {
	v, err := ConvertToInt("42", nil, nil)
	if err != nil || v != 42 {
		t.Errorf("expected 42, got %d, err %v", v, err)
	}

	_, err = ConvertToInt("abc", nil, nil)
	if err == nil {
		t.Error("expected error for non-integer")
	}

	min := 10
	_, err = ConvertToInt("5", &min, nil)
	if err == nil {
		t.Error("expected error for below minimum")
	}

	max := 100
	_, err = ConvertToInt("200", nil, &max)
	if err == nil {
		t.Error("expected error for above maximum")
	}
}

func TestFileBrowserState(t *testing.T) {
	fb := NewFileBrowserState()
	fb.SetFile("test.ivy", "line1\nline2\nline3", 2)
	if fb.LineCount() != 3 {
		t.Errorf("expected 3 lines, got %d", fb.LineCount())
	}
	if fb.GetLine(2) != "line2" {
		t.Errorf("expected 'line2', got %q", fb.GetLine(2))
	}
	if fb.GetLine(0) != "" {
		t.Error("GetLine(0) should return empty")
	}
	if fb.GetLine(99) != "" {
		t.Error("GetLine(99) should return empty")
	}
	if fb.HighlightLine != 2 {
		t.Error("HighlightLine should be 2")
	}
}

func TestRunContext(t *testing.T) {
	var capturedErr error
	rc := NewRunContext(func(err error) {
		capturedErr = err
	})

	err := rc.Run(func() error {
		return nil
	})
	if err != nil || capturedErr != nil {
		t.Error("no error expected")
	}

	err = rc.Run(func() error {
		panic("test panic")
	})
	if err == nil {
		t.Error("expected error from panic")
	}
	if capturedErr == nil {
		t.Error("OnError should have been called")
	}
}

func TestBuildMenuBar(t *testing.T) {
	menus := []MenuDef{{Type: "menu", Label: "Test"}}
	mb := BuildMenuBar(menus)
	if len(mb.Menus) != 1 {
		t.Error("BuildMenuBar failed")
	}
}

// --- Show tests ---

func TestShowVerification(t *testing.T) {
	_, err := ShowVerification("")
	if err == nil {
		t.Error("expected error for empty path")
	}

	sess, err := ShowVerification("test.ivy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess == nil {
		t.Fatal("session is nil")
	}
}

func TestLaunchUI(t *testing.T) {
	cfg := iu.NewConfig()
	_, err := LaunchUI(cfg, nil, ":0")
	if err == nil {
		t.Error("expected error for nil session")
	}

	sess := NewSession("test")
	srv, err := LaunchUI(cfg, sess, ":0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv == nil {
		t.Fatal("server is nil")
	}
}

// --- Fuzz tests ---

func FuzzConvertToInt(f *testing.F) {
	f.Add("42")
	f.Add("0")
	f.Add("-1")
	f.Add("abc")
	f.Add("")
	f.Add("99999999999999999999")
	f.Fuzz(func(t *testing.T, s string) {
		v, err := ConvertToInt(s, nil, nil)
		if err == nil {
			// If no error, should be a valid int.
			if s == "" {
				t.Error("empty string should fail")
			}
			_ = v
		}
	})
}

func FuzzFilterEvents(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add(".*")
	f.Add("a very long pattern that probably matches nothing")
	f.Fuzz(func(t *testing.T, pattern string) {
		evs := []*TraceEvent{
			NewTraceEvent("hello world"),
			NewTraceEvent("test event"),
		}
		// Should not panic.
		_ = FilterEvents(evs, pattern)
	})
}
