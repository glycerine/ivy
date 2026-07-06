//go:build web

package webui

import (
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
	"testing"
)

// loadARGTestSession creates a session with ivySample loaded and an initial state.
func loadARGTestSession(t *testing.T) (*Session, *AnalysisGraphUI) {
	t.Helper()
	cfg := goivy.NewConfig()
	s := NewSession(cfg, "test-arg")
	if err := s.LoadFileContent("test.ivy", []byte(ivySample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	drainEvents(s)
	if s.AGUI == nil {
		t.Fatal("AGUI should be non-nil after LoadFileContent")
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()
	return s, s.AGUI
}

func requireArgPayload(t *testing.T, result map[string]interface{}) map[string]interface{} {
	t.Helper()
	arg, ok := result["arg"].(map[string]interface{})
	if !ok {
		t.Fatalf("result missing ARG payload: %#v", result)
	}
	return arg
}

func requireArgState(t *testing.T, arg map[string]interface{}) *goivy.AnalysisGraphState {
	t.Helper()
	state, ok := arg["analysis_graph_state"].(*goivy.AnalysisGraphState)
	if !ok {
		t.Fatalf("ARG payload missing typed analysis graph state: %#v", arg["analysis_graph_state"])
	}
	return state
}

func requireCyNodeByObj(t *testing.T, arg map[string]interface{}, obj string) goivy.CyElement {
	t.Helper()
	elements, ok := arg["elements"].([]goivy.CyElement)
	if !ok {
		t.Fatalf("ARG payload elements have unexpected type: %#v", arg["elements"])
	}
	for _, elem := range elements {
		if elem.Group == "nodes" && elem.Data["obj"] == obj {
			return elem
		}
	}
	t.Fatalf("node %q not found in ARG payload", obj)
	return goivy.CyElement{}
}

func countCyNodesWithClass(arg map[string]interface{}, className string) int {
	elements, _ := arg["elements"].([]goivy.CyElement)
	count := 0
	for _, elem := range elements {
		if elem.Group == "nodes" && strings.Contains(elem.Classes, className) {
			count++
		}
	}
	return count
}

// makeMinimalAGUI builds an AnalysisGraphUI backed by a fresh AnalysisGraph
// with one initial state. No module/solver needed.
func makeMinimalAGUI(t *testing.T, mod *goivy.Module) *AnalysisGraphUI {
	t.Helper()
	ui := NewAnalysisGraphUI()
	ui.AG = goivy.NewAnalysisGraph(mod)
	ui.Mod = mod
	ui.AG.AddInitialState(nil, nil)
	return ui
}

// --- Helpers ---

func TestStateByIDValid(t *testing.T) {
	_, ui := loadARGTestSession(t)
	state, err := ui.stateByID(0)
	if err != nil {
		t.Fatalf("stateByID(0): %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil state")
	}
	if state.ID != 0 {
		t.Errorf("expected state ID 0, got %d", state.ID)
	}
}

func TestStateByIDOutOfBounds(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_, err := ui.stateByID(999)
	if err == nil {
		t.Fatal("expected error for out-of-bounds ID")
	}
	_, err = ui.stateByID(-1)
	if err == nil {
		t.Fatal("expected error for negative ID")
	}
}

func TestStateByIDNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	_, err := ui.stateByID(0)
	if err == nil {
		t.Fatal("expected error when AG is nil")
	}
}

func TestTransitionByEndpointsNotFound(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_, err := ui.transitionByEndpoints(0, 99)
	if err == nil {
		t.Fatal("expected error for nonexistent transition")
	}
}

func TestTransitionByEndpointsNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	_, err := ui.transitionByEndpoints(0, 1)
	if err == nil {
		t.Fatal("expected error when AG is nil")
	}
}

func TestArtToGraphState(t *testing.T) {
	s, _ := loadARGTestSession(t)
	gs := ArtToGraphState(s.AG)
	if gs == nil {
		t.Fatal("expected non-nil WebUIAnalysisGraphState")
	}
	if len(gs.States) != len(s.AG.States) {
		t.Errorf("state count mismatch: got %d, want %d", len(gs.States), len(s.AG.States))
	}
}

// --- Stub 1: NodeExecuteCommands ---

func TestNodeExecuteCommands(t *testing.T) {
	_, ui := loadARGTestSession(t)
	entries := ui.NodeExecuteCommands(0)
	if len(entries) == 0 {
		t.Skip("no actions available in this module")
	}
	for i := 1; i < len(entries); i++ {
		if entries[i-1].Label > entries[i].Label {
			t.Errorf("entries not sorted: %q > %q", entries[i-1].Label, entries[i].Label)
		}
	}
}

func TestNodeExecuteCommandsNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	entries := ui.NodeExecuteCommands(0)
	if entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}

func TestNodeExecuteCommandsBadNode(t *testing.T) {
	_, ui := loadARGTestSession(t)
	entries := ui.NodeExecuteCommands(999)
	if entries != nil {
		t.Errorf("expected nil for invalid nodeID, got %v", entries)
	}
}

// --- Stub 2: CheckLocalSafety ---

func TestCheckLocalSafetySafe(t *testing.T) {
	_, ui := loadARGTestSession(t)
	safe, msg := ui.CheckLocalSafety(0)
	if !safe {
		t.Errorf("expected initial state to be safe, got msg: %s", msg)
	}
}

func TestCheckLocalSafetyNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	safe, msg := ui.CheckLocalSafety(0)
	if safe {
		t.Error("expected false when AG is nil")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

// --- Stub 3: CheckBoundedSafety ---

func TestCheckBoundedSafetySafe(t *testing.T) {
	_, ui := loadARGTestSession(t)
	safe, msg := ui.CheckBoundedSafety(0)
	if !safe {
		t.Errorf("expected initial state bounded-safe, got msg: %s", msg)
	}
}

func TestCheckBoundedSafetyNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	safe, msg := ui.CheckBoundedSafety(0)
	if safe {
		t.Error("expected false when AG is nil")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

// --- Stub 4: FindExtension ---

func TestFindExtensionNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	_, err := ui.FindExtension(0)
	if err == nil {
		t.Fatal("expected error when AG is nil")
	}
}

// --- Stub 5: ExecuteAction ---

func TestExecuteAction(t *testing.T) {
	s, ui := loadARGTestSession(t)
	before := len(s.AG.States)
	err := ui.ExecuteAction(0, "ext:connect")
	if err != nil {
		t.Fatalf("ExecuteAction: %v", err)
	}
	after := len(s.AG.States)
	if after != before+1 {
		t.Errorf("expected state count %d, got %d", before+1, after)
	}
	if len(s.AG.Transitions) == 0 {
		t.Error("expected at least one transition after execute")
	}
}

func TestExecuteActionBadName(t *testing.T) {
	_, ui := loadARGTestSession(t)
	err := ui.ExecuteAction(0, "nonexistent_action")
	if err == nil {
		t.Fatal("expected error for unknown action name")
	}
}

func TestExecuteActionBadNode(t *testing.T) {
	_, ui := loadARGTestSession(t)
	err := ui.ExecuteAction(999, "ext:connect")
	if err == nil {
		t.Fatal("expected error for invalid node ID")
	}
}

func TestExecuteActionSyncsGraph(t *testing.T) {
	s, ui := loadARGTestSession(t)
	synced := false
	ui.SyncCallback = func() { synced = true; s.syncARGToGraph() }
	_ = ui.ExecuteAction(0, "ext:connect")
	if !synced {
		t.Error("SyncCallback was not called")
	}
	if len(s.Graph.States) != len(s.AG.States) {
		t.Errorf("Graph states (%d) != AG states (%d)", len(s.Graph.States), len(s.AG.States))
	}
}

// --- Stub 6: RecalculateAll ---

func TestRecalculateAll(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_ = ui.ExecuteAction(0, "ext:connect")
	before := len(ui.AG.States)
	ui.RecalculateAll()
	after := len(ui.AG.States)
	if after != before {
		t.Errorf("state count changed: %d → %d", before, after)
	}
}

func TestRecalculateAllNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	ui.RecalculateAll() // should not panic
}

// --- Stub 7: RecalculateEdge ---

func TestRecalculateEdge(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_ = ui.ExecuteAction(0, "ext:connect")
	if len(ui.AG.Transitions) == 0 {
		t.Skip("no transitions to recalculate")
	}
	tr := ui.AG.Transitions[0]
	ui.RecalculateEdge(tr.Pre.ID, tr.Post.ID) // should not panic
}

func TestRecalculateEdgeBadIDs(t *testing.T) {
	_, ui := loadARGTestSession(t)
	ui.RecalculateEdge(99, 100) // should not panic, just a no-op
}

// --- Stub 8: DecomposeEdge ---

func TestDecomposeEdgeBadIDs(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_, err := ui.DecomposeEdge(99, 100)
	if err == nil {
		t.Fatal("expected error for nonexistent edge")
	}
}

func TestDecomposeEdgeNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	_, err := ui.DecomposeEdge(0, 1)
	if err == nil {
		t.Fatal("expected error when AG is nil")
	}
}

// --- Stub 9: ViewSourceEdge ---

func TestViewSourceEdgeBadIDs(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_, _, err := ui.ViewSourceEdge(99, 100)
	if err == nil {
		t.Fatal("expected error for nonexistent edge")
	}
}

func TestViewSourceEdgeNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	_, _, err := ui.ViewSourceEdge(0, 1)
	if err == nil {
		t.Fatal("expected error when AG is nil")
	}
}

// --- Stub 10: CoverNode ---

func TestCoverNodeNoMark(t *testing.T) {
	_, ui := loadARGTestSession(t)
	ok, err := ui.CoverNode(0)
	if ok {
		t.Error("expected false without mark")
	}
	if err == nil {
		t.Fatal("expected error without mark")
	}
}

func TestCoverNodeBadID(t *testing.T) {
	_, ui := loadARGTestSession(t)
	ui.MarkNode(&ARGStateRef{ID: 0})
	ok, err := ui.CoverNode(999)
	if ok {
		t.Error("expected false for invalid ID")
	}
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestCoverNodeWithMark(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_ = ui.ExecuteAction(0, "ext:connect")
	if len(ui.AG.States) < 2 {
		t.Skip("need at least 2 states for cover test")
	}
	ui.MarkNode(&ARGStateRef{ID: 0})
	ok, err := ui.CoverNode(1)
	// Covering may or may not succeed depending on state ordering.
	// The important thing is it doesn't panic.
	_ = ok
	_ = err
}

// --- Stub 11: JoinNode ---

func TestJoinNodeNoMark(t *testing.T) {
	_, ui := loadARGTestSession(t)
	err := ui.JoinNode(0)
	if err == nil {
		t.Fatal("expected error without mark")
	}
}

func TestJoinNodeBadID(t *testing.T) {
	_, ui := loadARGTestSession(t)
	ui.MarkNode(&ARGStateRef{ID: 0})
	err := ui.JoinNode(999)
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestJoinNodeWithMark(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_ = ui.ExecuteAction(0, "ext:connect")
	if len(ui.AG.States) < 2 {
		t.Skip("need at least 2 states for join test")
	}
	before := len(ui.AG.States)
	ui.MarkNode(&ARGStateRef{ID: 0})
	err := ui.JoinNode(1)
	if err != nil {
		t.Fatalf("JoinNode: %v", err)
	}
	after := len(ui.AG.States)
	if after != before+1 {
		t.Errorf("expected state count %d after join, got %d", before+1, after)
	}
}

// --- Stub 12: TryConjecture ---

func TestTryConjectureEmpty(t *testing.T) {
	_, ui := loadARGTestSession(t)
	err := ui.TryConjecture(0, "")
	if err == nil {
		t.Fatal("expected error for empty conjecture")
	}
}

func TestTryConjectureNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	err := ui.TryConjecture(0, "true")
	if err == nil {
		t.Fatal("expected error when AG is nil")
	}
}

// --- Stub 13: TryRememberedGraph ---

func TestTryRememberedGraphEmpty(t *testing.T) {
	_, ui := loadARGTestSession(t)
	err := ui.TryRememberedGraph(0, "")
	if err != nil {
		t.Fatalf("empty goalName should return nil, got: %v", err)
	}
}

func TestTryRememberedGraphNotFound(t *testing.T) {
	_, ui := loadARGTestSession(t)
	err := ui.TryRememberedGraph(0, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent graph")
	}
}

// --- Stub 14: BMC ---

func TestBMCBadNode(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_, err := ui.BMC(999, "true", 5)
	if err == nil {
		t.Fatal("expected error for invalid node ID")
	}
}

func TestBMCNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	_, err := ui.BMC(0, "true", 5)
	if err == nil {
		t.Fatal("expected error when AG is nil")
	}
}

func TestArgNodeActionBMCReturnsReachabilityAndTrace(t *testing.T) {
	cfg := goivy.NewConfig()
	s := NewSession(cfg, "test-arg-node-bmc")
	mod := goivy.New()
	ag := goivy.NewAnalysisGraph(mod)
	ag.Add(goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)), nil)
	ui := NewAnalysisGraphUI()
	ui.AG = ag
	ui.Mod = mod
	ui.G = AnalysisUIARGState(ui)
	ui.SyncCallback = func() { ui.G = AnalysisUIARGState(ui) }
	s.AG = ag
	s.AGUI = ui
	s.CompiledModule = mod
	drainEvents(s)

	unreachable, err := s.ArgNodeAction("state_0", "bmc", map[string]interface{}{
		"bound":    float64(0),
		"err_cond": "false",
	})
	if err != nil {
		t.Fatalf("ArgNodeAction bmc unreachable: %v", err)
	}
	if got := unreachable["reachable"]; got != false {
		t.Fatalf("unreachable reachable = %#v, want false", got)
	}
	if got := unreachable["result"]; got != "pass" {
		t.Fatalf("unreachable result = %#v, want pass", got)
	}
	if _, ok := unreachable["trace_arg"]; ok {
		t.Fatalf("unreachable BMC should not include trace_arg: %#v", unreachable)
	}

	reachable, err := s.ArgNodeAction("state_0", "bmc", map[string]interface{}{
		"bound":    float64(0),
		"err_cond": "true",
	})
	if err != nil {
		t.Fatalf("ArgNodeAction bmc reachable: %v", err)
	}
	if got := reachable["bound"]; got != 0 {
		t.Fatalf("reachable bound = %#v, want 0", got)
	}
	if got := reachable["err_cond"]; got != "true" {
		t.Fatalf("reachable err_cond = %#v, want true", got)
	}
	if got := reachable["reachable"]; got != true {
		t.Fatalf("reachable reachable = %#v, want true", got)
	}
	if got := reachable["result"]; got != "fail" {
		t.Fatalf("reachable result = %#v, want fail", got)
	}
	if _, ok := reachable["trace_sheet_id"].(string); !ok {
		t.Fatalf("missing trace_sheet_id in result: %#v", reachable)
	}
	traceArg, ok := reachable["trace_arg"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing trace_arg in result: %#v", reachable)
	}
	traceState := requireArgState(t, traceArg)
	if len(traceState.States) == 0 {
		t.Fatalf("trace ARG has no states: %#v", traceState)
	}
	if !traceState.States[len(traceState.States)-1].IsMarked {
		t.Fatalf("final trace state is not marked: %#v", traceState.States)
	}
}

// --- Stub 15: TryProperty ---

func TestTryPropertyEmpty(t *testing.T) {
	ui := NewIvyUI()
	err := ui.TryProperty("")
	if err == nil {
		t.Fatal("expected error for empty property")
	}
}

func TestTryPropertyNoModule(t *testing.T) {
	ui := NewIvyUI()
	err := ui.TryProperty("true")
	if err == nil {
		t.Fatal("expected error when no module loaded")
	}
}

// --- DeleteNode ---

func TestDeleteNodeDelegates(t *testing.T) {
	_, ui := loadARGTestSession(t)
	_ = ui.ExecuteAction(0, "ext:connect")
	before := len(ui.AG.States)
	if before < 2 {
		t.Skip("need at least 2 states for delete test")
	}
	ui.DeleteNode(before - 1)
	after := len(ui.AG.States)
	if after >= before {
		t.Errorf("expected state count to decrease after delete: %d → %d", before, after)
	}
}

func TestDeleteNodeClearsMark(t *testing.T) {
	_, ui := loadARGTestSession(t)
	ui.MarkNode(&ARGStateRef{ID: 0})
	ui.DeleteNode(0)
	if ui.GetMark() != nil {
		t.Error("expected mark to be cleared after deleting marked node")
	}
}

func TestDeleteNodeNilAG(t *testing.T) {
	ui := NewAnalysisGraphUI()
	ui.DeleteNode(0) // should not panic
}

// --- Session ArgNodeAction dispatch ---

func TestArgNodeActionCheckSafety(t *testing.T) {
	s, _ := loadARGTestSession(t)
	drainEvents(s)
	result, err := s.ArgNodeAction("state_0", "check_safety", nil)
	if err != nil {
		t.Fatalf("ArgNodeAction: %v", err)
	}
	if _, ok := result["safe"]; !ok {
		t.Error("expected 'safe' key in result")
	}
	arg := requireArgPayload(t, result)
	state := requireArgState(t, arg)
	if len(state.States) == 0 || !state.States[0].IsSafe {
		t.Fatalf("typed ARG state did not record safe node: %#v", state.States)
	}
	node := requireCyNodeByObj(t, arg, "state_0")
	if !strings.Contains(node.Classes, "safe_state") {
		t.Fatalf("safe node is missing safe_state class: %q", node.Classes)
	}
	if node.Data["is_safe"] != true {
		t.Fatalf("safe node is missing is_safe metadata: %#v", node.Data)
	}
}

func TestArgNodeActionBoundedSafetyFailureReturnsTrace(t *testing.T) {
	cfg := goivy.NewConfig()
	s := NewSession(cfg, "test-unsafe-bounded")
	mod := goivy.New()
	ag := goivy.NewAnalysisGraph(mod)
	ag.Add(goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)), nil)
	ag.Assertions = append(ag.Assertions, &goivy.LabeledFormula{Formula: goivy.False})
	ui := NewAnalysisGraphUI()
	ui.AG = ag
	ui.Mod = mod
	ui.G = AnalysisUIARGState(ui)
	ui.SyncCallback = func() { ui.G = AnalysisUIARGState(ui) }
	s.AG = ag
	s.AGUI = ui
	s.CompiledModule = mod
	drainEvents(s)

	result, err := s.ArgNodeAction("state_0", "check_safety", map[string]interface{}{"mode": "bounded"})
	if err != nil {
		t.Fatalf("ArgNodeAction bounded check_safety: %v", err)
	}
	if got := result["safe"]; got != false {
		t.Fatalf("safe = %#v, want false", got)
	}
	if got := result["result"]; got != "fail" {
		t.Fatalf("result = %#v, want fail", got)
	}
	if got := result["message"]; got != "The node is unsafe: View error trace?" {
		t.Fatalf("message = %#v", got)
	}
	if _, ok := result["trace_sheet_id"].(string); !ok {
		t.Fatalf("missing trace_sheet_id in result: %#v", result)
	}
	traceArg, ok := result["trace_arg"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing trace_arg in result: %#v", result)
	}
	traceState := requireArgState(t, traceArg)
	if len(traceState.States) == 0 {
		t.Fatalf("trace ARG has no states: %#v", traceState)
	}
	if !traceState.States[len(traceState.States)-1].IsMarked {
		t.Fatalf("final trace state is not marked: %#v", traceState.States)
	}
	if countCyNodesWithClass(traceArg, "marked_state") != 1 {
		t.Fatalf("trace ARG should render exactly one marked node: %#v", traceArg["elements"])
	}
}

func TestArgNodeActionExtendClosedNodeReturnsDialogPayload(t *testing.T) {
	cfg := goivy.NewConfig()
	s := NewSession(cfg, "test-extend-closed")
	mod := goivy.New()
	ag := goivy.NewAnalysisGraph(mod)
	ag.Add(goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil)), nil)
	ui := NewAnalysisGraphUI()
	ui.AG = ag
	ui.Mod = mod
	ui.G = AnalysisUIARGState(ui)
	s.AG = ag
	s.AGUI = ui
	s.SheetUIs = map[string]*AnalysisGraphUI{rootSheetID: ui}
	s.sheetCounter = 1

	result, err := s.ArgNodeAction("state_0", "find_extension", nil)
	if err != nil {
		t.Fatalf("find_extension should return a closed-node payload, not an error: %v", err)
	}
	if result["closed"] != true {
		t.Fatalf("closed node payload missing closed=true: %#v", result)
	}
	if result["result"] != "closed" {
		t.Fatalf("closed node result = %#v, want closed", result["result"])
	}
	if result["message"] != "State 0 is closed." {
		t.Fatalf("closed node message = %#v", result["message"])
	}
}

func TestArgNodeActionExtendAddsStateAndViewsConceptGraph(t *testing.T) {
	s, ui := loadARGTestSession(t)
	drainEvents(s)
	beforeStates := len(s.AG.States)
	result, err := s.ArgNodeAction("state_0", "find_extension", nil)
	if err != nil {
		t.Fatalf("find_extension: %v", err)
	}
	if len(s.AG.States) <= beforeStates {
		t.Fatalf("extend should add a new ARG state, before=%d after=%d", beforeStates, len(s.AG.States))
	}
	if _, ok := result["arg"]; !ok {
		t.Fatalf("extend result missing ARG payload: %#v", result)
	}
	if _, ok := result["concept"]; !ok {
		t.Fatalf("extend result missing concept payload for viewed extension state: %#v", result)
	}
	if result["extension"] == "" {
		t.Fatalf("extend result missing extension label: %#v", result)
	}
	if ui.CurrentConceptGraph == nil || ui.CurrentConceptGraph.G() == nil {
		t.Fatal("extend did not install a current concept graph")
	}
	if got, want := ui.CurrentConceptGraph.G().ParentState, s.AG.States[len(s.AG.States)-1]; got != want {
		t.Fatalf("concept graph parent state = %#v, want new extension state %#v", got, want)
	}
}

func TestArgNodeActionMark(t *testing.T) {
	s, _ := loadARGTestSession(t)
	drainEvents(s)
	result, err := s.ArgNodeAction("state_0", "mark", nil)
	if err != nil {
		t.Fatalf("ArgNodeAction: %v", err)
	}
	if result["marked"] != true {
		t.Error("expected marked=true")
	}
	if s.AGUI.GetMark() == nil {
		t.Error("expected AGUI mark to be set")
	}
	arg := requireArgPayload(t, result)
	state := requireArgState(t, arg)
	if len(state.States) == 0 || !state.States[0].IsMarked {
		t.Fatalf("typed ARG state did not record marked node: %#v", state.States)
	}
	node := requireCyNodeByObj(t, arg, "state_0")
	if !strings.Contains(node.Classes, "marked_state") {
		t.Fatalf("marked node is missing marked_state class: %q", node.Classes)
	}
	if node.Data["is_marked"] != true {
		t.Fatalf("marked node is missing is_marked metadata: %#v", node.Data)
	}
}

func TestArgNodeActionMarkReplacesCurrentMarkedNode(t *testing.T) {
	s, _ := loadARGTestSession(t)
	drainEvents(s)
	if _, err := s.ArgNodeAction("state_0", "execute_action", map[string]interface{}{"action_name": "ext:connect"}); err != nil {
		t.Fatalf("execute_action: %v", err)
	}
	if len(s.AG.States) < 2 {
		t.Skip("need at least 2 states for mark replacement test")
	}
	if _, err := s.ArgNodeAction("state_0", "mark", nil); err != nil {
		t.Fatalf("first mark: %v", err)
	}
	recalc, err := s.ArgNodeAction("state_0", "recalculate", nil)
	if err != nil {
		t.Fatalf("recalculate: %v", err)
	}
	if countCyNodesWithClass(requireArgPayload(t, recalc), "marked_state") != 1 {
		t.Fatalf("marked node should survive recalculation as a single visual mark")
	}
	result, err := s.ArgNodeAction("state_1", "mark", nil)
	if err != nil {
		t.Fatalf("second mark: %v", err)
	}
	arg := requireArgPayload(t, result)
	if countCyNodesWithClass(arg, "marked_state") != 1 {
		t.Fatalf("expected exactly one marked node after replacing mark")
	}
	if node := requireCyNodeByObj(t, arg, "state_0"); strings.Contains(node.Classes, "marked_state") {
		t.Fatalf("old mark is still visually marked: %q", node.Classes)
	}
	if node := requireCyNodeByObj(t, arg, "state_1"); !strings.Contains(node.Classes, "marked_state") {
		t.Fatalf("new mark is not visually marked: %q", node.Classes)
	}
	if mark := s.AGUI.GetMark(); mark == nil || mark.ID != 1 {
		t.Fatalf("backend mark = %#v, want state_1", mark)
	}
	_, err = s.ArgNodeAction("state_0", "cover", nil)
	if err != nil && strings.Contains(err.Error(), "no marked node") {
		t.Fatalf("cover did not use the backend mark: %v", err)
	}
}

func TestArgNodeActionUnknown(t *testing.T) {
	s, _ := loadARGTestSession(t)
	drainEvents(s)
	_, err := s.ArgNodeAction("state_0", "totally_unknown_action", nil)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestArgNodeActionDelete(t *testing.T) {
	s, _ := loadARGTestSession(t)
	drainEvents(s)
	before := len(s.AG.States)
	_, err := s.ArgNodeAction("state_0", "delete", nil)
	if err != nil {
		t.Fatalf("ArgNodeAction delete: %v", err)
	}
	after := len(s.AG.States)
	if after >= before {
		t.Errorf("expected state count to decrease: %d → %d", before, after)
	}
}

// --- AGUI wiring on Session ---

func TestSessionAGUIWired(t *testing.T) {
	s := loadTestSession(t)
	if s.AGUI == nil {
		t.Fatal("AGUI should be set after LoadFileContent")
	}
	if s.AGUI.AG != s.AG {
		t.Error("AGUI.AG should point to session AG")
	}
	if s.AGUI.Mod != s.CompiledModule {
		t.Error("AGUI.Mod should point to session CompiledModule")
	}
	if s.AGUI.SyncCallback == nil {
		t.Error("AGUI.SyncCallback should be set")
	}
}

func TestSyncARGToGraphUsesArtToGraphState(t *testing.T) {
	s, _ := loadARGTestSession(t)
	s.syncARGToGraph()
	if len(s.Graph.States) != len(s.AG.States) {
		t.Errorf("Graph states (%d) != AG states (%d)", len(s.Graph.States), len(s.AG.States))
	}
}
