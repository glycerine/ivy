package webui

import (
	"encoding/json"
	goivy "github.com/glycerine/ivy/goivy"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const executeActionMenuSample = `#lang ivy1.7
type client
type server
relation link(X:client, Y:server)
relation semaphore(X:server)
after init { semaphore(W) := true; link(X,Y) := false }
action connect(x:client,y:server) = { require semaphore(y); link(x,y) := true; semaphore(y) := false }
export connect
conjecture link(X,Y) -> ~semaphore(Y)
`

const ctiUsedRelationSample = `#lang ivy1.7
type node
relation p(X:node)
relation q(X:node)
relation unused(X:node)
relation spare(X:node)
after init { p(X) := false; q(X) := true; unused(X) := false }
action bad(x:node) = { p(x) := true }
export bad
conjecture p(X) -> ~q(X)
`

func argStepInTestDomainSetup() (*CDConceptDomain, *goivy.UninterpretedSort) {
	mustVar := func(name string, sort goivy.Sort) *goivy.LogicVariable {
		v, err := goivy.NewVariable(name, sort)
		if err != nil {
			panic(err)
		}
		return v
	}
	mustFuncSort := func(sorts ...goivy.Sort) *goivy.LogicFunctionSort {
		fs, err := goivy.NewFunctionSort(sorts...)
		if err != nil {
			panic(err)
		}
		return fs
	}
	mustApply := func(fn goivy.Expr, args ...goivy.Expr) goivy.Expr {
		app, err := goivy.NewApply(fn, args...)
		if err != nil {
			panic(err)
		}
		return app
	}
	mustNot := func(expr goivy.Expr) goivy.Expr {
		not, err := goivy.NewNot(expr)
		if err != nil {
			panic(err)
		}
		return not
	}
	mustAnd := func(exprs ...goivy.Expr) goivy.Expr {
		and, err := goivy.NewAnd(exprs...)
		if err != nil {
			panic(err)
		}
		return and
	}

	S := &goivy.UninterpretedSort{Name: "S"}
	X := mustVar("X", S)
	Y := mustVar("Y", S)
	unaryRel := mustFuncSort(S, goivy.Boolean)
	binaryRel := mustFuncSort(S, S, goivy.Boolean)
	p := goivy.NewConst("p", unaryRel)
	q := goivy.NewConst("q", unaryRel)
	r := goivy.NewConst("r", binaryRel)

	concepts := NewCDConceptDict()
	concepts.SetConcept("both", MustCDConcept("both", []*goivy.LogicVariable{X}, mustAnd(mustApply(p, X), mustApply(q, X))))
	concepts.SetConcept("none", MustCDConcept("none", []*goivy.LogicVariable{X}, mustAnd(mustNot(mustApply(p, X)), mustNot(mustApply(q, X)))))
	concepts.SetConcept("onlyp", MustCDConcept("onlyp", []*goivy.LogicVariable{X}, mustAnd(mustApply(p, X), mustNot(mustApply(q, X)))))
	concepts.SetConcept("onlyq", MustCDConcept("onlyq", []*goivy.LogicVariable{X}, mustAnd(mustNot(mustApply(p, X)), mustApply(q, X))))
	concepts.SetConcept("r", MustCDConcept("r", []*goivy.LogicVariable{X, Y}, mustApply(r, X, Y)))
	concepts.SetList("nodes", []string{"both", "none", "onlyp", "onlyq"})
	concepts.SetList("edges", []string{"r"})
	concepts.SetList("node_labels", []string{})

	return NewCDConceptDomain(concepts, GetStandardCombiners(), GetStandardCombinations()), S
}

const saveInvariantSample = `#lang ivy1.7
type node
relation p(X:node)
relation q(X:node)
after init { p(X) := true; q(X) := true }
conjecture p(X) -> q(X)
conjecture q(X) -> q(X)
`

func TestArgStepInClientServerDiagnosticEdge(t *testing.T) {
	path := filepath.Join("..", "..", "ivy-lang-examples", "doc", "examples", "client_server_example.ivy")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}

	s := NewSession(goivy.NewConfig(), "test-step-in")
	if err := s.LoadFileContent("client_server_example.ivy", content); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	cr := s.RunCheck("induction")
	if cr.Result != "fail" {
		t.Fatalf("RunCheck induction result = %q, want fail; message: %s", cr.Result, cr.Message)
	}
	if cr.FailedConjecture == "" {
		t.Fatalf("RunCheck induction failed without reporting failed conjecture; message: %s", cr.Message)
	}
	if !strings.Contains(cr.FailedConjecture, "link") {
		t.Fatalf("failed conjecture %q does not mention expected relation link", cr.FailedConjecture)
	}
	if s.AG == nil || len(s.AG.Transitions) == 0 {
		t.Fatalf("induction failure did not populate ARG transitions")
	}
	if got := s.AG.Transitions[0].Label; got != "call ext" {
		t.Fatalf("ARG transition label = %q, want %q", got, "call ext")
	}
	if cr.CounterexampleDetails == "" || !strings.Contains(cr.CounterexampleDetails, "Counterexample trace") {
		t.Fatalf("induction failure did not return counterexample details: %#v", cr.CounterexampleDetails)
	}
	if s.AGUI == nil || s.AGUI.CurrentConceptGraph == nil {
		t.Fatalf("induction failure did not install a current concept graph")
	}
	if s.AGUI.CurrentConceptGraph.G().ParentState != s.AG.States[0] {
		t.Fatalf("concept graph parent state was not switched to the CTI predecessor")
	}
	if len(s.SimpleSess.AbstractValue) == 0 {
		t.Fatalf("concept graph abstract value was not recomputed for the CTI predecessor")
	}
	if got := countSimpleConceptNodesOfSort(s.SimpleSess, "client"); got < 2 {
		t.Fatalf("CTI concept graph has %d client nodes, want at least 2 concrete witnesses", got)
	}
	if got := countSimpleConceptNodesOfSort(s.SimpleSess, "server"); got < 1 {
		t.Fatalf("CTI concept graph has %d server nodes, want at least 1 concrete witness", got)
	}
	for _, want := range []string{"=@X", "=@Y", "=@Z", "link(X,Y)", "semaphore"} {
		if !stringSliceContains(s.SimpleSess.RelationNames(), want) {
			t.Fatalf("CTI relation rows = %v, missing %q", s.SimpleSess.RelationNames(), want)
		}
	}
	if !hasConcreteAllToAllEdge(s.SimpleSess, "link") {
		t.Fatalf("CTI concept graph did not expose a concrete all_to_all link edge; abstract value=%v", s.SimpleSess.AbstractValue)
	}

	result, err := s.ArgNodeAction("state_0", "decompose", map[string]interface{}{"target": "state_1"})
	if err != nil {
		t.Fatalf("ArgNodeAction decompose: %v", err)
	}
	if result["decomposed"] != true {
		t.Fatalf("decomposed = %v, want true; result=%#v", result["decomposed"], result)
	}
	subARG, ok := result["sub_arg"].(map[string]interface{})
	if !ok {
		t.Fatalf("sub_arg missing or wrong type: %#v", result["sub_arg"])
	}
	elements, ok := subARG["elements"].([]WebUICyElement)
	if !ok {
		t.Fatalf("sub_arg elements missing or wrong type: %#v", subARG["elements"])
	}
	if len(elements) == 0 {
		t.Fatal("sub_arg elements empty")
	}
}

func countSimpleConceptNodesOfSort(cs *ConceptSession, sortName string) int {
	if cs == nil || cs.Domain == nil {
		return 0
	}
	count := 0
	for _, node := range cs.Domain.Nodes {
		c := cs.Domain.Concepts[node]
		if c != nil && len(c.Sorts) > 0 && c.Sorts[0] == sortName {
			count++
		}
	}
	return count
}

func hasConcreteAllToAllEdge(cs *ConceptSession, edgeName string) bool {
	if cs == nil || cs.Domain == nil {
		return false
	}
	nodeSet := make(map[string]bool)
	for _, node := range cs.Domain.Nodes {
		nodeSet[node] = true
	}
	prefix := "edge_info|all_to_all|" + edgeName + "|"
	for key, value := range cs.AbstractValue {
		if !value || !strings.HasPrefix(key, prefix) {
			continue
		}
		parts := strings.Split(key, "|")
		if len(parts) == 5 && nodeSet[parts[3]] && nodeSet[parts[4]] {
			return true
		}
	}
	return false
}

func TestArgStepInRegistersIndependentAnalysisSheet(t *testing.T) {
	path := filepath.Join("..", "..", "ivy-lang-examples", "doc", "examples", "client_server_example.ivy")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}

	s := NewSession(goivy.NewConfig(), "test-step-in-sheet")
	if err := s.LoadFileContent("client_server_example.ivy", content); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	cr := s.RunCheck("induction")
	if cr.Result != "fail" {
		t.Fatalf("RunCheck induction result = %q, want fail; message: %s", cr.Result, cr.Message)
	}

	result, err := s.ArgNodeAction("state_0", "decompose", map[string]interface{}{
		"sheet_id": "sheet-1",
		"target":   "state_1",
	})
	if err != nil {
		t.Fatalf("ArgNodeAction decompose: %v", err)
	}
	sheetID, _ := result["sheet_id"].(string)
	if sheetID == "" || sheetID == "sheet-1" {
		t.Fatalf("decompose sheet_id = %#v, want new sheet id; result=%#v", result["sheet_id"], result)
	}

	s.mu.Lock()
	ui := s.analysisUIForSheetLocked(sheetID)
	root := s.analysisUIForSheetLocked("sheet-1")
	s.mu.Unlock()
	if ui == nil {
		t.Fatalf("decompose did not register backend analysis UI for %q", sheetID)
	}
	if ui == root {
		t.Fatalf("decompose sheet %q reuses root AnalysisGraphUI", sheetID)
	}
	if ui.AG == nil || len(ui.AG.States) == 0 {
		t.Fatalf("decompose sheet %q has no analysis graph states", sheetID)
	}

	selected, stateLabel, err := s.selectConceptARGNode(sheetID, "state_0")
	if err != nil {
		t.Fatalf("select concept node in decomposed sheet: %v", err)
	}
	if selected != "state_0" {
		t.Fatalf("selected = %q, want state_0", selected)
	}
	if stateLabel == "" {
		t.Fatalf("state label is empty for decomposed sheet")
	}
	s.mu.Lock()
	if ui.CurrentConceptGraph == nil {
		t.Fatalf("decomposed sheet did not receive its own concept graph")
	}
	if root != nil && root.CurrentConceptGraph == ui.CurrentConceptGraph {
		t.Fatalf("decomposed sheet concept graph aliases root concept graph")
	}
	s.mu.Unlock()
}

func TestARGExecuteActionMenuEntriesRenderAndDispatch(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-execute-menu")
	if err := s.LoadFileContent("test.ivy", []byte(executeActionMenuSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()

	cy := RenderAnalysisUIARG(s.AGUI)
	var actions []any
	for _, ele := range cy.Elements {
		if ele.Group == "nodes" && ele.Data["obj"] == "state_0" {
			raw, ok := ele.Data["actions"].([]goivy.NodeAction)
			if !ok {
				t.Fatalf("state_0 actions missing or wrong type: %#v", ele.Data["actions"])
			}
			for _, act := range raw {
				actions = append(actions, act)
			}
		}
	}
	if len(actions) == 0 {
		t.Fatalf("state_0 has no rendered actions")
	}
	foundExecute := false
	for _, raw := range actions {
		act := raw.(goivy.NodeAction)
		if act.Label == "ext:connect" && act.Action == "execute_action" {
			if got := act.Args["action_name"]; got != "ext:connect" {
				t.Fatalf("ext:connect action_name = %#v, want ext:connect", got)
			}
			foundExecute = true
			break
		}
	}
	if !foundExecute {
		t.Fatalf("state_0 actions do not include connect execute_action: %#v", actions)
	}

	beforeStates := len(s.AG.States)
	beforeTransitions := len(s.AG.Transitions)
	result, err := s.ArgNodeAction("state_0", "execute_action", map[string]interface{}{
		"sheet_id":    "sheet-1",
		"action_name": "ext:connect",
	})
	if err != nil {
		t.Fatalf("ArgNodeAction execute_action: %v", err)
	}
	if len(s.AG.States) != beforeStates+1 {
		t.Fatalf("states after execute = %d, want %d", len(s.AG.States), beforeStates+1)
	}
	if len(s.AG.Transitions) != beforeTransitions+1 {
		t.Fatalf("transitions after execute = %d, want %d", len(s.AG.Transitions), beforeTransitions+1)
	}
	if _, ok := result["arg"]; !ok {
		t.Fatalf("execute_action result missing arg update: %#v", result)
	}
}

func TestARGChoiceBackedCommandsExposeConjecturesAndRememberedGraphs(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-choice-backed")
	if err := s.LoadFileContent("test.ivy", []byte(executeActionMenuSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()

	choices, err := s.ArgNodeAction("state_0", "try_conjecture_choices", map[string]interface{}{"sheet_id": "sheet-1"})
	if err != nil {
		t.Fatalf("try_conjecture_choices: %v", err)
	}
	conjChoices, ok := choices["choices"].([]ChoiceItem)
	if !ok {
		t.Fatalf("choices have type %T, want []ChoiceItem: %#v", choices["choices"], choices["choices"])
	}
	if len(conjChoices) == 0 {
		t.Fatalf("expected at least one conjecture choice")
	}
	if !strings.Contains(conjChoices[0].Label, "link") {
		t.Fatalf("conjecture choice label = %q, want link conjecture", conjChoices[0].Label)
	}

	if _, _, err := s.selectConceptARGNode("sheet-1", "state_0"); err != nil {
		t.Fatalf("select concept graph: %v", err)
	}
	remember, err := s.ExecuteAction("remember", map[string]interface{}{
		"sheet_id": "sheet-1",
		"name":     "goal-a",
	})
	if err != nil {
		t.Fatalf("remember: %v", err)
	}
	if got := remember["remembered"]; got != "goal-a" {
		t.Fatalf("remembered = %#v, want goal-a", got)
	}
	rememberedChoices, err := s.ArgNodeAction("state_0", "try_remembered_choices", map[string]interface{}{"sheet_id": "sheet-1"})
	if err != nil {
		t.Fatalf("try_remembered_choices: %v", err)
	}
	goalChoices, ok := rememberedChoices["choices"].([]ChoiceItem)
	if !ok {
		t.Fatalf("remembered choices have type %T, want []ChoiceItem", rememberedChoices["choices"])
	}
	if len(goalChoices) != 1 || goalChoices[0].Value != "goal-a" {
		t.Fatalf("remembered choices = %#v, want goal-a", goalChoices)
	}
	result, err := s.ArgNodeAction("state_0", "try_remembered", map[string]interface{}{
		"sheet_id": "sheet-1",
		"goal":     "goal-a",
	})
	if err != nil {
		t.Fatalf("try_remembered: %v", err)
	}
	if _, ok := result["concept"]; !ok {
		t.Fatalf("try_remembered result missing concept graph update: %#v", result)
	}
}

func TestCheckFailureCarriesTraceARGForViewAction(t *testing.T) {
	path := filepath.Join("..", "..", "ivy-lang-examples", "doc", "examples", "client_server_example.ivy")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}

	cfg := goivy.NewConfig()
	be := NewGoBackend(cfg)
	sessionJSON, err := be.NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	var session map[string]string
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		t.Fatalf("session json: %v", err)
	}
	if _, err := be.Load(session["session_id"], "client_server_example.ivy", content, ""); err != nil {
		t.Fatalf("Load: %v", err)
	}
	resultJSON, err := be.Check(session["session_id"], "induction", CheckOptions{})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("check json: %v", err)
	}
	if result["result"] != "fail" {
		t.Fatalf("result = %#v, want fail; body=%s", result["result"], resultJSON)
	}
	trace, ok := result["trace_arg"].(map[string]interface{})
	if !ok {
		t.Fatalf("trace_arg missing or wrong type: %#v", result["trace_arg"])
	}
	elements, ok := trace["elements"].([]interface{})
	if !ok || len(elements) == 0 {
		t.Fatalf("trace_arg elements missing/empty: %#v", trace["elements"])
	}
	traceSheetID, _ := result["trace_sheet_id"].(string)
	if traceSheetID == "" || traceSheetID == "sheet-1" {
		t.Fatalf("trace_sheet_id = %#v, want registered non-root sheet", result["trace_sheet_id"])
	}
	if label, _ := result["trace_label"].(string); label == "" {
		t.Fatalf("trace_label missing: %#v", result["trace_label"])
	}
	if _, err := be.ArgAction(session["session_id"], "state_0", "execute_action", map[string]interface{}{
		"sheet_id":    traceSheetID,
		"action_name": "ext:connect",
	}); err != nil {
		t.Fatalf("registered trace sheet ArgAction: %v", err)
	}
	if details, _ := result["counterexample_details"].(string); details == "" || !strings.Contains(details, "Counterexample trace") {
		t.Fatalf("counterexample_details missing from check JSON: %#v", result["counterexample_details"])
	}
}

func TestShowReachableStatesOpensReachableARGSheet(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-show-reachable")
	if err := s.LoadFileContent("test.ivy", []byte(executeActionMenuSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	result, err := s.ExecuteAction("show_reachable", nil)
	if err != nil {
		t.Fatalf("show_reachable: %v", err)
	}
	sheetID, _ := result["sheet_id"].(string)
	if sheetID == "" || sheetID == "sheet-1" {
		t.Fatalf("sheet_id = %#v, want new reachable sheet", result["sheet_id"])
	}
	arg, ok := result["arg"].(map[string]interface{})
	if !ok {
		t.Fatalf("arg missing/wrong type: %#v", result["arg"])
	}
	elements, ok := arg["elements"].([]goivy.CyElement)
	if !ok || len(elements) == 0 {
		t.Fatalf("reachable arg elements missing/empty: %#v", arg["elements"])
	}
	s.mu.Lock()
	ui := s.analysisUIForSheetLocked(sheetID)
	s.mu.Unlock()
	if ui == nil || ui != s.ReachableUI {
		t.Fatalf("reachable sheet %q not registered to reachable UI", sheetID)
	}
}

func TestConceptResponseIncludesEdgeSortsForMaterializeFromSelected(t *testing.T) {
	cfg := goivy.NewConfig()
	be := NewGoBackend(cfg)
	sessionJSON, err := be.NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	var session map[string]string
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		t.Fatalf("session json: %v", err)
	}
	if _, err := be.Load(session["session_id"], "test.ivy", []byte(executeActionMenuSample), ""); err != nil {
		t.Fatalf("Load: %v", err)
	}
	conceptJSON, err := be.GetConcept(session["session_id"], "sheet-1", "")
	if err != nil {
		t.Fatalf("GetConcept: %v", err)
	}
	var concept map[string]interface{}
	if err := json.Unmarshal(conceptJSON, &concept); err != nil {
		t.Fatalf("concept json: %v", err)
	}
	edgeSorts, ok := concept["edge_sorts"].(map[string]interface{})
	if !ok {
		t.Fatalf("edge_sorts missing/wrong type: %#v", concept["edge_sorts"])
	}
	linkSorts, ok := edgeSorts["link"].([]interface{})
	if !ok || len(linkSorts) != 2 || linkSorts[0] != "client" || linkSorts[1] != "server" {
		t.Fatalf("edge_sorts[link] = %#v, want [client server]", edgeSorts["link"])
	}
}

func TestBacktrackUsesGraphWidgetBacktrackPoint(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-backtrack-action")
	s.AGUI = NewAnalysisGraphUI()
	s.SheetUIs = map[string]*AnalysisGraphUI{"sheet-1": s.AGUI}
	w := NewGraphWidget(StandardGraph([]string{"node"}, nil))
	s.AGUI.CurrentConceptGraph = w

	w.Checkpoint(true)
	w.G().State = "at-backtrack"
	w.Checkpoint(false)
	w.G().State = "past-backtrack"
	_, err := s.ExecuteAction("backtrack", map[string]interface{}{"sheet_id": "sheet-1"})
	if err != nil {
		t.Fatalf("backtrack: %v", err)
	}
	if got := w.G().State; got == "past-backtrack" {
		t.Fatalf("backtrack left state at %q", got)
	}
	if got := w.G().State; got != "" {
		t.Fatalf("backtrack state = %q, want initial backtrack point", got)
	}
}

func TestConcreteUsesCurrentConceptGraph(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-concrete-graph")
	s.AGUI = NewAnalysisGraphUI()
	s.SheetUIs = map[string]*AnalysisGraphUI{"sheet-1": s.AGUI}
	w := NewGraphWidget(StandardGraph([]string{"node"}, nil))
	w.G().State = "base_state"
	w.G().Concrete = "concrete_constraints"
	s.AGUI.CurrentConceptGraph = w

	result, err := s.ExecuteAction("concrete", map[string]interface{}{"sheet_id": "sheet-1"})
	if err != nil {
		t.Fatalf("concrete: %v", err)
	}
	if got := w.G().State; got != "base_state & concrete_constraints" {
		t.Fatalf("concrete graph state = %q, want combined state", got)
	}
	if _, ok := result["concept"]; !ok {
		t.Fatalf("concrete result missing concept graph update: %#v", result)
	}
}

func TestGatherUsesRequestedConceptGraphSheet(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-gather-sheet")
	rootUI := NewAnalysisGraphUI()
	otherUI := NewAnalysisGraphUI()
	rootW := NewGraphWidget(StandardGraph([]string{"root"}, nil))
	otherW := NewGraphWidget(StandardGraph([]string{"other"}, nil))
	rootW.G().ConceptSess.AbstractValue["root_fact"] = true
	otherW.G().ConceptSess.AbstractValue["other_fact"] = true
	rootUI.CurrentConceptGraph = rootW
	otherUI.CurrentConceptGraph = otherW
	s.AGUI = rootUI
	s.SheetUIs = map[string]*AnalysisGraphUI{
		"sheet-1": rootUI,
		"sheet-2": otherUI,
	}

	result, err := s.ExecuteAction("gather", map[string]interface{}{"sheet_id": "sheet-2"})
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	facts, ok := result["facts"].([]string)
	if !ok {
		t.Fatalf("facts have type %T, want []string: %#v", result["facts"], result["facts"])
	}
	if len(facts) != 1 || facts[0] != "other_fact" {
		t.Fatalf("facts = %#v, want only sheet-2 fact", facts)
	}
	if _, ok := result["concept"]; !ok {
		t.Fatalf("gather result missing concept graph update: %#v", result)
	}
}

func TestReachUsesCurrentConceptGraphParentState(t *testing.T) {
	mod := goivy.New()
	p := goivy.NewConst("p", goivy.Boolean)
	mod.Relations.Set("p", goivy.Boolean)
	pred := goivy.NewState(mod, goivy.TrueClauses(nil))
	pred.Unders = []*goivy.State{
		goivy.NewState(mod, goivy.FalseClauses(nil)),
		goivy.NewState(mod, goivy.TrueClauses(nil)),
	}
	target := goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{p}, nil, nil))
	target.Pred = pred
	target.Update = goivy.GetUpdate(goivy.NewAssignAction(p, goivy.True), &goivy.UpdateContext{
		Domain: mod,
		ActCfg: goivy.NewActionsConfig(),
	})
	unrelated := goivy.NewState(mod, goivy.TrueClauses(nil))
	ag := goivy.NewAnalysisGraph(mod)
	ag.Add(pred, nil)
	ag.Add(target, nil)
	ag.Add(unrelated, nil)

	ui := NewAnalysisGraphUI()
	ui.AG = ag
	ui.Mod = mod
	w := NewGraphWidget(StandardGraph([]string{"node"}, target))
	w.G().ParentState = target
	ui.CurrentConceptGraph = w

	s := NewSession(goivy.NewConfig(), "test-reach-sheet")
	s.CompiledModule = mod
	s.AG = ag
	s.AGUI = ui
	s.SheetUIs = map[string]*AnalysisGraphUI{"sheet-1": ui}

	result, err := s.ExecuteAction("reach", map[string]interface{}{"sheet_id": "sheet-1"})
	if err != nil {
		t.Fatalf("reach: %v", err)
	}
	if result["reachable"] != true {
		t.Fatalf("reachable = %#v, want true; result=%#v", result["reachable"], result)
	}
	if len(target.Unders) != 1 {
		t.Fatalf("target unders = %d, want reached under appended", len(target.Unders))
	}
}

func TestExportConceptGraphReturnsDOT(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-export-dot")
	ui := NewAnalysisGraphUI()
	w := NewGraphWidget(StandardGraph([]string{"node"}, nil))
	w.G().ConceptSess.Domain.Nodes = []string{"node"}
	w.G().ConceptSess.Domain.Concepts["node"] = &Concept{
		Name:    "node",
		Formula: "X:node = X:node",
		Sorts:   []string{"node"},
		Arity:   1,
	}
	ui.CurrentConceptGraph = w
	s.AGUI = ui
	s.SheetUIs = map[string]*AnalysisGraphUI{"sheet-1": ui}

	result, err := s.ExecuteAction("export", map[string]interface{}{"sheet_id": "sheet-1"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	content, _ := result["content"].(string)
	if !strings.HasPrefix(content, "digraph concept_graph") {
		t.Fatalf("export content = %q, want DOT graph", content)
	}
	if !strings.Contains(content, "node") {
		t.Fatalf("export content %q does not mention rendered node sort", content)
	}
	if got := result["filename"]; got != "concept_graph.dot" {
		t.Fatalf("filename = %#v, want concept_graph.dot", got)
	}
}

func TestLoadFileContentInitializesCTIUI(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-load-cti")
	if err := s.LoadFileContent("test.ivy", []byte(executeActionMenuSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	if s.CTIUI == nil {
		t.Fatal("CTIUI should be initialized after loading a module")
	}
	if len(s.CTIUI.Conjectures) == 0 {
		t.Fatal("CTIUI should load module conjectures")
	}
	if s.CTIUI.CurrentConceptGraph == nil {
		t.Fatal("CTIUI should view state 0 and create a concept graph")
	}
	if s.CTIUI.AnalysisGraphUI == nil || s.CTIUI.AnalysisGraphUI.AG == nil || len(s.CTIUI.AnalysisGraphUI.AG.States) == 0 {
		t.Fatal("CTIUI should have its own initialized analysis graph")
	}
	if len(s.AG.States) != 0 {
		t.Fatalf("generic session ARG states = %d, want lazy empty ARG preserved", len(s.AG.States))
	}
}

func TestCTIConceptBoundedCheckUsesSelectedFacts(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-cti-concept-bmc")
	mod := goivy.New()
	ui := NewCTIAnalysisGraphUI(mod)
	widget := NewGraphWidget(StandardGraph([]string{"S"}, nil))
	domain, _ := argStepInTestDomainSetup()
	widget.G().InteractiveSess = NewConceptInteractiveSession(
		domain,
		goivy.True,
		goivy.True,
		nil,
		[]goivy.Expr{goivy.True},
		nil,
		nil,
		nil,
		false,
	)
	ui.CurrentConceptGraph = widget
	s.CTIUI = ui

	result, err := s.ExecuteAction("cti_bounded_check", map[string]interface{}{
		"sheet_id": "sheet-1",
		"bound":    float64(0),
	})
	if err != nil {
		t.Fatalf("cti_bounded_check: %v", err)
	}
	if s.CTIUI.CurrentBound != 0 {
		t.Fatalf("CurrentBound = %d, want 0", s.CTIUI.CurrentBound)
	}
	if got := result["bound"]; got != 0 {
		t.Fatalf("bound result = %#v, want 0", got)
	}
	message, _ := result["message"].(string)
	if !strings.Contains(message, "BMC with bound 0") {
		t.Fatalf("message = %q, want BMC bound message", message)
	}
	if _, ok := result["conjecture"].(string); !ok {
		t.Fatalf("result missing conjecture string: %#v", result)
	}
	if found, _ := result["found"].(bool); found {
		if _, ok := result["trace_arg"].(map[string]interface{}); !ok {
			t.Fatalf("found CTI BMC result missing trace_arg: %#v", result)
		}
		if sheetID, _ := result["trace_sheet_id"].(string); !strings.HasPrefix(sheetID, "sheet-") {
			t.Fatalf("trace_sheet_id = %q, want registered sheet id", sheetID)
		}
	}
}

func TestCTIConceptBoundedCheckEmptySelectionUsesPythonDefault(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-cti-empty-concept-bmc")
	ui := NewCTIAnalysisGraphUI(goivy.New())
	widget := NewGraphWidget(StandardGraph([]string{"S"}, nil))
	domain, _ := argStepInTestDomainSetup()
	widget.G().InteractiveSess = NewConceptInteractiveSession(
		domain,
		goivy.True,
		goivy.True,
		nil,
		nil,
		nil,
		nil,
		nil,
		false,
	)
	ui.CurrentConceptGraph = widget
	s.CTIUI = ui

	result, err := s.ExecuteAction("cti_bounded_check", map[string]interface{}{
		"sheet_id": "sheet-1",
		"bound":    float64(10),
	})
	if err != nil {
		t.Fatalf("cti_bounded_check with empty selection: %v", err)
	}
	message, _ := result["message"].(string)
	if !strings.Contains(message, "BMC with bound 0") {
		t.Fatalf("message = %q, want Python-style bound-0 counterexample message", message)
	}
	conjecture, _ := result["conjecture"].(string)
	if !strings.Contains(conjecture, "true") {
		t.Fatalf("conjecture = %q, want empty-selection ~true conjecture", conjecture)
	}
	trace, ok := result["trace_arg"].(map[string]interface{})
	if !ok {
		t.Fatalf("empty-selection counterexample missing trace_arg: %#v", result)
	}
	if elements, ok := trace["elements"].([]goivy.CyElement); !ok || len(elements) == 0 {
		t.Fatalf("trace_arg elements missing/empty: %#v", trace["elements"])
	}
	if sheetID, _ := result["trace_sheet_id"].(string); sheetID == "" || s.analysisUIForSheetLocked(sheetID) == nil {
		t.Fatalf("trace sheet %q was not registered", sheetID)
	}
	if label, _ := result["trace_label"].(string); label == "" {
		t.Fatalf("trace_label missing: %#v", result)
	}
}

func TestInductionFailureUsedRelationsExcludeUnusedSignatureRelations(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-used-relations")
	if err := s.LoadFileContent("test.ivy", []byte(ctiUsedRelationSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	cr := s.RunCheck("induction")
	if cr.Result != "fail" {
		t.Fatalf("RunCheck induction result = %q, want fail; message=%s", cr.Result, cr.Message)
	}
	for _, rel := range cr.UsedRelations {
		if rel == "unused" || rel == "spare" {
			t.Fatalf("used relations include unused relation: %#v", cr.UsedRelations)
		}
	}
	if !stringSliceContains(cr.UsedRelations, "p") || !stringSliceContains(cr.UsedRelations, "q") {
		t.Fatalf("used relations = %#v, want p and q", cr.UsedRelations)
	}
	toggles := s.GetToggles()
	for _, rel := range []string{"p", "q"} {
		if toggles.Edges[rel] == nil || !toggles.Edges[rel]["all_to_all"] {
			t.Fatalf("used relation %q was not enabled in concept graph toggles: %#v", rel, toggles.Edges[rel])
		}
	}
	if toggles.Edges["unused"] != nil && toggles.Edges["unused"]["all_to_all"] {
		t.Fatalf("unused relation was enabled in concept graph toggles: %#v", toggles.Edges["unused"])
	}
}

func TestConceptDiagramRoutesThroughCTIUI(t *testing.T) {
	be := NewGoBackend(goivy.NewConfig())
	s := NewSession(goivy.NewConfig(), "test-cti-diagram")
	s.CTIUI = NewCTIAnalysisGraphUI(nil)
	be.sessions[s.ID] = s

	_, err := be.ConceptDiagram(s.ID)
	if err == nil || !strings.Contains(err.Error(), "no module loaded") {
		t.Fatalf("ConceptDiagram error = %v, want CTI Diagram no-module error", err)
	}
}

func TestReachabilityDomainActionsReturnConceptSnapshots(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-domain-actions")
	if err := s.LoadFileContent("test.ivy", []byte(executeActionMenuSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()
	if _, err := s.AGUI.ViewState(0, "", false); err != nil {
		t.Fatalf("ViewState: %v", err)
	}

	reset, err := s.ExecuteAction("reset_domain", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("reset_domain: %v", err)
	}
	if got := reset["type"]; got != "reset_domain" {
		t.Fatalf("reset_domain type = %v, want reset_domain; result=%#v", got, reset)
	}
	resetConcept, ok := reset["concept"].(map[string]interface{})
	if !ok {
		t.Fatalf("reset_domain concept payload missing: %#v", reset["concept"])
	}
	if elements, ok := resetConcept["elements"].([]WebUICyElement); !ok || len(elements) == 0 {
		t.Fatalf("reset_domain elements missing/empty: %#v", resetConcept["elements"])
	}

	if _, err := s.ExecuteAction("diagram", map[string]interface{}{"sheet_id": rootSheetID}); err != nil {
		t.Fatalf("diagram: %v", err)
	}

	diagram, err := s.ExecuteAction("diagram_domain", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("diagram_domain: %v", err)
	}
	if got := diagram["type"]; got != "diagram_domain" {
		t.Fatalf("diagram_domain type = %v, want diagram_domain; result=%#v", got, diagram)
	}
	diagramConcept, ok := diagram["concept"].(map[string]interface{})
	if !ok {
		t.Fatalf("diagram_domain concept payload missing: %#v", diagram["concept"])
	}
	diagramDomain, ok := diagramConcept["concept_domain"].(*ConceptDomain)
	if !ok || diagramDomain == nil || len(diagramDomain.Concepts) == 0 {
		t.Fatalf("diagram_domain concept domain missing/empty: %#v", diagramConcept["concept_domain"])
	}
	if s.AGUI.CurrentConceptGraph == nil || s.AGUI.CurrentConceptGraph.G() == nil {
		t.Fatalf("diagram_domain lost current concept graph")
	}
}

func TestDiagramVacuousReturnsDialogPayload(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-vacuous-diagram")
	if err := s.LoadFileContent("test.ivy", []byte(executeActionMenuSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()
	if _, err := s.AGUI.ViewState(0, "false", false); err != nil {
		t.Fatalf("ViewState false: %v", err)
	}

	result, err := s.ExecuteAction("diagram", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("diagram: %v", err)
	}
	if got := result["type"]; got != "vacuous" {
		t.Fatalf("diagram type = %v, want vacuous; result=%#v", got, result)
	}
	if got := result["status"]; got != "vacuous" {
		t.Fatalf("diagram status = %v, want vacuous; result=%#v", got, result)
	}
	message, _ := result["message"].(string)
	if !strings.Contains(message, "The current state is vacuous.") {
		t.Fatalf("diagram message = %q, want vacuous dialog text", message)
	}
	if _, ok := result["concept"].(map[string]interface{}); !ok {
		t.Fatalf("diagram concept payload missing: %#v", result["concept"])
	}
}

func TestSaveInvariantUsesPythonKeptDroppedSections(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-save-invariant")
	if err := s.LoadFileContent("test.ivy", []byte(saveInvariantSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	_, err := s.ExecuteAction("weaken", map[string]interface{}{
		"indices": []interface{}{float64(0)},
	})
	if err != nil {
		t.Fatalf("weaken: %v", err)
	}
	result, err := s.ExecuteAction("save_invariant", nil)
	if err != nil {
		t.Fatalf("save_invariant: %v", err)
	}
	content, _ := result["content"].(string)
	for _, want := range []string{
		"# original conjectures kept",
		"# original conjectures dropped",
		"# invariant",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("save invariant content missing %q:\n%s", want, content)
		}
	}
}

func TestArgViewSourceReturnsLoadedSourceAndLine(t *testing.T) {
	path := filepath.Join("..", "..", "ivy-lang-examples", "doc", "examples", "client_server_example.ivy")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}

	s := NewSession(goivy.NewConfig(), "test-view-source")
	if err := s.LoadFileContent("client_server_example.ivy", content); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	cr := s.RunCheck("induction")
	if cr.Result != "fail" {
		t.Fatalf("RunCheck induction result = %q, want fail; message: %s", cr.Result, cr.Message)
	}

	result, err := s.ArgNodeAction("state_0", "view_source", map[string]interface{}{"target": "state_1"})
	if err != nil {
		t.Fatalf("ArgNodeAction view_source: %v", err)
	}
	if got := result["source"]; got != string(content) {
		t.Fatalf("source mismatch: got %T %q", got, got)
	}
	if file, _ := result["file"].(string); file == "" {
		t.Fatalf("file missing: %#v", result)
	}
}

func TestArgViewSourceLocatedActionReturnsLine(t *testing.T) {
	const source = "line1\naction go = {\n}\n"
	mod := goivy.New()
	ag := goivy.NewAnalysisGraph(mod)
	pre := goivy.NewState(mod, goivy.TrueClauses(nil))
	post := goivy.NewState(mod, goivy.TrueClauses(nil))
	act := goivy.NewAssumeAction(goivy.True)
	act.SetLineno(goivy.Location{Filename: "sample.ivy", Line: 2})
	ag.Add(pre, nil)
	ag.Add(post, goivy.NewActionApp(act, pre))

	s := NewSession(goivy.NewConfig(), "test-view-source-located")
	s.FilePath = "sample.ivy"
	s.FileContent = source
	s.AG = ag
	s.AGUI = NewAnalysisGraphUI()
	s.AGUI.AG = ag
	s.AGUI.Mod = mod

	result, err := s.ArgNodeAction("state_0", "view_source", map[string]interface{}{"target": "state_1"})
	if err != nil {
		t.Fatalf("ArgNodeAction view_source: %v", err)
	}
	if got := result["source"]; got != source {
		t.Fatalf("source mismatch: got %T %q", got, got)
	}
	if got := result["file"]; got != "sample.ivy" {
		t.Fatalf("file = %#v, want sample.ivy", got)
	}
	if got := result["lineno"]; got != 2 {
		t.Fatalf("lineno = %#v, want 2", got)
	}
}
