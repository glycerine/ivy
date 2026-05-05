//go:build web

package webui

import (
	"encoding/json"
	goivy "github.com/glycerine/ivy/goivy"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	_, ui := loadARGTestSession(t)
	// Execute an action to get a second state (ID 1).
	err := ui.ExecuteAction(0, "ext:connect")
	if err != nil {
		t.Fatalf("ExecuteAction: %v", err)
	}
	before := len(ui.AG.States)
	if before < 2 {
		t.Fatalf("expected at least 2 states, got %d", before)
	}
	ui.DeleteNode(before - 1)
	after := len(ui.AG.States)
	if after >= before {
		t.Errorf("expected fewer states after delete, got %d → %d", before, after)
	}
}

func TestAnalysisGraphUICheckSafety(t *testing.T) {
	_, ui := loadARGTestSession(t)
	safe, msg := ui.CheckSafetyNode(0)
	if !safe {
		t.Errorf("initial state should be safe, got msg: %s", msg)
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
	cfg := NewExtConfig()
	if cfg.ArgNodeActions.Len() < 2 {
		t.Errorf("expected at least 2 default arg_node_actions callbacks, got %d", cfg.ArgNodeActions.Len())
	}
}

func TestGoalNodeActionsEmpty(t *testing.T) {
	cfg := NewExtConfig()
	if cfg.GoalNodeActions.Len() != 0 {
		t.Errorf("expected 0 goal_node_actions callbacks, got %d", cfg.GoalNodeActions.Len())
	}
}

// --- CTI UI tests ---

func TestNewCTIAnalysisGraphUI(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	if ui == nil {
		t.Fatal("NewCTIAnalysisGraphUI returned nil")
	}
	if ui.HaveCTI {
		t.Error("should not have CTI initially")
	}
}

func TestCTIMenus(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	menus := ui.CTIMenus()
	if len(menus) != 2 {
		t.Fatalf("expected 2 CTI menus, got %d", len(menus))
	}
	if menus[1].Label != "Invariant" {
		t.Errorf("expected Invariant menu, got %q", menus[1].Label)
	}
}

func TestCTIWeaken(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	c1 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	c2 := goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil)
	c3 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	ui.Conjectures = []*goivy.Clauses{c1, c2, c3}
	removed, err := ui.Weaken([]int{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(removed) != 1 || removed[0] != c2 {
		t.Errorf("expected c2 removed, got %v", removed)
	}
	if len(ui.Conjectures) != 2 {
		t.Errorf("expected 2 conjectures remaining, got %d", len(ui.Conjectures))
	}
}

func TestCTIWeakenEmpty(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	_, err := ui.Weaken(nil)
	if err == nil {
		t.Error("expected error for empty indices")
	}
}

func TestCTISaveConjectures(t *testing.T) {
	ui := NewCTIAnalysisGraphUI(nil)
	c1 := goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)
	c2 := goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil)
	ui.Conjectures = []*goivy.Clauses{c1, c2}
	result := ui.SaveConjectures()
	if !strings.Contains(result, "invariant") {
		t.Error("save should contain 'invariant'")
	}
	if !strings.Contains(result, "# conjectures") {
		t.Error("save should contain '# conjectures' header")
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

func TestBrowserMenuDescriptors(t *testing.T) {
	menus := BuildBrowserMenuDescriptors()
	if len(menus.Arg) == 0 {
		t.Fatal("missing ARG menu descriptors")
	}
	if len(menus.Concept) == 0 {
		t.Fatal("missing concept menu descriptors")
	}
	var foundUndo bool
	for _, menu := range menus.Concept {
		if menu.Label != "Action" {
			continue
		}
		for _, item := range menu.Items {
			if item.Action == "undo" {
				foundUndo = true
				if item.Dispatch != "action" {
					t.Fatalf("undo dispatch = %q, want action", item.Dispatch)
				}
				if !item.Enabled {
					t.Fatal("undo descriptor should be enabled")
				}
			}
		}
	}
	if !foundUndo {
		t.Fatal("concept Action menu missing undo descriptor")
	}
}

// --- Show tests ---

func TestShowVerificationEmptyPath(t *testing.T) {
	cfg := goivy.NewConfig()
	_, err := ShowVerification(cfg, "")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestShowVerificationNonexistentFile(t *testing.T) {
	cfg := goivy.NewConfig()
	_, err := ShowVerification(cfg, "nonexistent_file_9999.ivy")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("error should mention file read failure, got: %v", err)
	}
}

func TestShowVerificationInvalidContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.ivy")
	if err := os.WriteFile(path, []byte("@@@ not valid ivy @@@"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := goivy.NewConfig()
	_, err := ShowVerification(cfg, path)
	if err == nil {
		t.Error("expected error for invalid ivy content")
	}
	if !strings.Contains(err.Error(), "failed to compile file") {
		t.Errorf("error should mention compile failure, got: %v", err)
	}
}

func TestShowVerificationSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "client_server.ivy")
	if err := os.WriteFile(path, []byte(ivySample), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := goivy.NewConfig()
	sess, err := ShowVerification(cfg, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess == nil {
		t.Fatal("session is nil")
	}
	if sess.CompiledModule == nil {
		t.Error("CompiledModule should be non-nil")
	}
	if sess.CompiledSig == nil {
		t.Error("CompiledSig should be non-nil")
	}
	if sess.AG == nil {
		t.Error("AnalysisGraph should be non-nil")
	}
	if sess.AGUI == nil {
		t.Error("AGUI should be non-nil")
	}
	if sess.ConceptSess == nil {
		t.Error("ConceptSess should be non-nil")
	}
	if sess.SimpleSess == nil {
		t.Error("SimpleSess should be non-nil")
	}
	if sess.FileContent != string(ivySample) {
		t.Error("FileContent should match input")
	}
	if !cfg.IsolateCfg.ShowCompiled {
		t.Error("ShowCompiled should be true after ShowVerification")
	}
}

func TestShowVerificationFromTestVectors(t *testing.T) {
	path := filepath.Join("..", "test_vectors", "bmc_minimal.ivy")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("test vector not available: %v", err)
	}
	cfg := goivy.NewConfig()
	sess, err := ShowVerification(cfg, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.CompiledModule == nil {
		t.Error("CompiledModule should be non-nil")
	}
	if sess.AG == nil {
		t.Error("AG should be non-nil")
	}
}

func TestShowVerificationNoHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noheader.ivy")
	src := "type t\nrelation r(X:t)\nafter init { r(X) := false }\n"
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := goivy.NewConfig()
	sess, err := ShowVerification(cfg, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.CompiledModule == nil {
		t.Error("CompiledModule should be non-nil")
	}
}

func TestCheckModuleAndShowComponents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.ivy")
	if err := os.WriteFile(path, []byte(ivySample), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := goivy.NewConfig()
	sess, err := ShowVerification(cfg, path)
	if err != nil {
		t.Fatalf("ShowVerification: %v", err)
	}

	srv, err := LaunchUI(cfg, sess, ":0")
	if err != nil {
		t.Fatalf("LaunchUI: %v", err)
	}
	if srv == nil {
		t.Fatal("server is nil")
	}
}

func TestLaunchUI(t *testing.T) {
	cfg := goivy.NewConfig()
	_, err := LaunchUI(cfg, nil, ":0")
	if err == nil {
		t.Error("expected error for nil session")
	}

	sess := NewSession(cfg, "test")
	srv, err := LaunchUI(cfg, sess, ":0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv == nil {
		t.Fatal("server is nil")
	}
}

// --- Proof stack tests ---

func loadProofTestSession(t *testing.T) *Session {
	t.Helper()
	cfg := goivy.NewConfig()
	s := NewSession(cfg, "test-proof")
	if err := s.LoadFileContent("test.ivy", []byte(ivySample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	return s
}

func TestProofManagerCreatedAfterLoad(t *testing.T) {
	s := loadProofTestSession(t)
	if s.ProofMgr == nil {
		t.Fatal("ProofMgr should be non-nil after LoadFileContent")
	}
	if s.ProofMgr.Goals == nil {
		t.Fatal("ProofMgr.Goals should be non-nil")
	}
}

func TestProofStackPopulatedFromConjectures(t *testing.T) {
	s := loadProofTestSession(t)
	// ivySample has 1 conjecture: "conjecture link(X,Y) -> ~semaphore(Y)"
	ps := s.ProofStackData()
	if ps == nil {
		t.Fatal("ProofStackData() returned nil")
	}
	if len(ps.Goals) != 1 {
		t.Fatalf("expected 1 proof goal from 1 conjecture, got %d", len(ps.Goals))
	}
}

func TestProofStackDataNonNilAfterLoad(t *testing.T) {
	s := loadProofTestSession(t)
	ps := s.ProofStackData()
	if ps == nil {
		t.Fatal("ProofStackData() should not be nil")
	}
}

func TestProofStackRenderFromSync(t *testing.T) {
	s := loadProofTestSession(t)
	ps := s.ProofStackData()
	cy := RenderProofStack(ps)
	if cy == nil {
		t.Fatal("RenderProofStack returned nil")
	}
	// 1 conjecture → 1 goal node, no parent edges
	if len(cy.Elements) < 1 {
		t.Fatalf("expected at least 1 element, got %d", len(cy.Elements))
	}
}

func TestProofGoalActionView(t *testing.T) {
	s := loadProofTestSession(t)
	result, err := s.ProofGoalAction("goal_0", "view")
	if err != nil {
		t.Fatalf("view error: %v", err)
	}
	info, ok := result["info"]
	if !ok {
		t.Fatal("result should contain 'info' key")
	}
	infoStr, ok := info.(string)
	if !ok || infoStr == "" {
		t.Error("info should be a non-empty string with formula text")
	}
}

func TestProofGoalActionRefute(t *testing.T) {
	s := loadProofTestSession(t)
	ps := s.ProofStackData()
	initialCount := len(ps.Goals)
	if initialCount == 0 {
		t.Skip("no goals to refute")
	}

	result, err := s.ProofGoalAction("goal_0", "refute")
	if err != nil {
		t.Fatalf("refute error: %v", err)
	}
	if result["refuted"] != true {
		t.Error("expected refuted=true")
	}
	ps = s.ProofStackData()
	if len(ps.Goals) != initialCount-1 {
		t.Errorf("expected %d goals after refute, got %d", initialCount-1, len(ps.Goals))
	}
}

func TestProofGoalActionPushPop(t *testing.T) {
	s := loadProofTestSession(t)
	ps := s.ProofStackData()
	initialCount := len(ps.Goals)

	// Push a sub-goal from goal_0
	result, err := s.ProofGoalAction("goal_0", "push")
	if err != nil {
		t.Fatalf("push error: %v", err)
	}
	if _, ok := result["new_goal_id"]; !ok {
		t.Error("push result should contain new_goal_id")
	}
	ps = s.ProofStackData()
	if len(ps.Goals) != initialCount+1 {
		t.Errorf("expected %d goals after push, got %d", initialCount+1, len(ps.Goals))
	}

	// Pop the top goal
	result, err = s.ProofGoalAction("", "pop")
	if err != nil {
		t.Fatalf("pop error: %v", err)
	}
	if _, ok := result["popped_id"]; !ok {
		t.Error("pop result should contain popped_id")
	}
	ps = s.ProofStackData()
	if len(ps.Goals) != initialCount {
		t.Errorf("expected %d goals after pop, got %d", initialCount, len(ps.Goals))
	}
}

func TestEmptyModuleEmptyProofStack(t *testing.T) {
	cfg := goivy.NewConfig()
	s := NewSession(cfg, "test-empty")
	src := "type t\nrelation r(X:t)\nafter init { r(X) := false }\n"
	if err := s.LoadFileContent("empty.ivy", []byte(src)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	ps := s.ProofStackData()
	if ps == nil {
		t.Fatal("ProofStackData() should not be nil even with no conjectures")
	}
	if len(ps.Goals) != 0 {
		t.Errorf("expected 0 goals for module with no conjectures, got %d", len(ps.Goals))
	}
}

func TestProofGoalParentRelationships(t *testing.T) {
	s := loadProofTestSession(t)
	// Push a second goal to create a parent relationship
	s.ProofGoalAction("goal_0", "push")
	ps := s.ProofStackData()
	if len(ps.Goals) < 2 {
		t.Fatalf("expected at least 2 goals, got %d", len(ps.Goals))
	}
	// The pushed goal should have the previous top as parent
	lastGoal := ps.Goals[len(ps.Goals)-1]
	if lastGoal.ParentID < 0 {
		t.Errorf("pushed goal should have a parent, got ParentID=%d", lastGoal.ParentID)
	}
}

func TestGetProofNonEmptyJSON(t *testing.T) {
	cfg := goivy.NewConfig()
	be := NewGoBackend(cfg)
	defer be.Close()

	byNew, err := be.NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	// Extract session ID
	var resp map[string]string
	if err := json.Unmarshal(byNew, &resp); err != nil {
		t.Fatalf("unmarshal session: %v", err)
	}
	sid := resp["session_id"]

	_, err = be.Load(sid, "test.ivy", []byte(ivySample))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	data, err := be.GetProof(sid)
	if err != nil {
		t.Fatalf("GetProof: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetProof returned empty data")
	}
	// Should contain at least 1 element (the conjecture goal)
	if !strings.Contains(string(data), "elements") {
		t.Error("GetProof JSON should contain 'elements' key")
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
