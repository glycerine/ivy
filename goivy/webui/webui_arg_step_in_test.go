//go:build web

package webui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	goivy "github.com/glycerine/ivy/goivy"
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

func readClientServerExample(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "ivy-lang-examples", "doc", "examples", "client_server_example.ivy")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read client_server_example.ivy: %v", err)
	}
	return content
}

func clientServerExampleWithIndividualC0(t *testing.T) []byte {
	t.Helper()
	content := string(readClientServerExample(t))
	withC0 := strings.Replace(content, "type server\n", "type server\n\nindividual c0 : client\n", 1)
	if withC0 == content {
		t.Fatalf("client_server_example.ivy did not contain expected type server declaration")
	}
	return []byte(withC0)
}

func clientServerExampleWithConcreteIndividuals(t *testing.T) []byte {
	t.Helper()
	content := string(readClientServerExample(t))
	withIndividuals := strings.Replace(content, "type server\n", "type server\n\nindividual c0 : client\nindividual c1 : client\nindividual s0 : server\n", 1)
	if withIndividuals == content {
		t.Fatalf("client_server_example.ivy did not contain expected type server declaration")
	}
	return []byte(withIndividuals)
}

func pythonIvyDiagramDomainNodes(t *testing.T, content []byte) []string {
	t.Helper()
	pyivyRoot, err := filepath.Abs(filepath.Join("..", "..", "pyivy", "ivy"))
	if err != nil {
		t.Fatalf("resolve pyivy root: %v", err)
	}
	python, err := filepath.Abs(filepath.Join("..", "..", "pyivy", "goivy-venv", "bin", "python3"))
	if err != nil {
		t.Fatalf("resolve pyivy python: %v", err)
	}
	if _, err := os.Stat(python); err != nil {
		var lookErr error
		python, lookErr = exec.LookPath("python3")
		if lookErr != nil {
			t.Skip("python3 not available")
		}
	}
	script := `
import json
import sys
from ivy import ivy_module as im
from ivy.ivy_compiler import ivy_from_string
from ivy.concept import get_diagram_concept_domain
from ivy.logic import And

with im.Module():
    ivy_from_string(sys.stdin.read(), create_isolate=False)
    cd = get_diagram_concept_domain(im.module.sig, And())
    print(json.dumps(cd.concepts['nodes']))
`
	cmd := exec.Command(python, "-c", script)
	cmd.Dir = pyivyRoot
	cmd.Stdin = strings.NewReader(string(content))
	pythonPath := pyivyRoot
	if existing := os.Getenv("PYTHONPATH"); existing != "" {
		pythonPath += string(os.PathListSeparator) + existing
	}
	cmd.Env = append(os.Environ(),
		"IVY_HOME="+pyivyRoot,
		"PYTHONPATH="+pythonPath,
		"XTRACE_OFF=1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python Diagram Domain probe failed: %v\n%s", err, out)
	}
	var nodes []string
	if err := json.Unmarshal(out, &nodes); err != nil {
		t.Fatalf("decode python Diagram Domain nodes: %v\n%s", err, out)
	}
	return nodes
}

func TestArgStepInClientServerDiagnosticEdge(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-step-in")
	if err := s.LoadFileContent("cti_used_relation.ivy", []byte(ctiUsedRelationSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	cr := s.RunCheck("induction")
	if cr.Result != "fail" {
		t.Fatalf("RunCheck induction result = %q, want fail; message: %s", cr.Result, cr.Message)
	}
	if cr.FailedConjecture == "" {
		t.Fatalf("RunCheck induction failed without reporting failed conjecture; message: %s", cr.Message)
	}
	if !strings.Contains(cr.FailedConjecture, "p") || !strings.Contains(cr.FailedConjecture, "q") {
		t.Fatalf("failed conjecture %q does not mention expected relations p and q", cr.FailedConjecture)
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
	if got := countSimpleConceptNodesOfSort(s.SimpleSess, "node"); got < 1 {
		t.Fatalf("CTI concept graph has %d node witnesses, want at least 1 concrete witness", got)
	}
	relationRows := strings.Join(s.SimpleSess.RelationNames(), "\n")
	for _, want := range []string{"=@X", "p", "q"} {
		if !strings.Contains(relationRows, want) {
			t.Fatalf("CTI relation rows = %v, missing %q", s.SimpleSess.RelationNames(), want)
		}
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

func TestClientServerStrengtheningInvariantMakesInductionPass(t *testing.T) {
	content := append([]byte{}, readClientServerExample(t)...)
	content = append(content, []byte(`

private {
    invariant ~(link(X,Y) & semaphore(Y))
}
`)...)

	s := NewSession(goivy.NewConfig(), "test-client-server-strengthened")
	if err := s.LoadFileContent("client_server_example.ivy", content); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	if s.ActiveIsolate != NoIsolatesFoundChoice {
		t.Fatalf("ActiveIsolate = %q, want %q", s.ActiveIsolate, NoIsolatesFoundChoice)
	}
	cr := s.RunCheck("induction")
	if cr.Result != "pass" {
		t.Fatalf("RunCheck induction result = %q, want pass; message: %s; failed conjecture: %s", cr.Result, cr.Message, cr.FailedConjecture)
	}
}

func TestLoadWithoutDeclaredIsolatesUsesSentinelChoice(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-no-isolates-choice")
	if err := s.LoadFileContent("client_server_example.ivy", readClientServerExample(t)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	if s.ActiveIsolate != NoIsolatesFoundChoice {
		t.Fatalf("ActiveIsolate = %q, want %q", s.ActiveIsolate, NoIsolatesFoundChoice)
	}
	if len(s.AvailableIsolates) != 1 || s.AvailableIsolates[0] != NoIsolatesFoundChoice {
		t.Fatalf("AvailableIsolates = %#v, want [%q]", s.AvailableIsolates, NoIsolatesFoundChoice)
	}
	if _, ok := s.CompiledModule.Actions.Get2("ext:connect"); !ok {
		t.Fatalf("no-isolates load should still compile Ivy's implicit ext:connect action")
	}
	if _, ok := s.CompiledModule.Actions.Get2("connect"); ok {
		t.Fatalf("no-isolates load should expose the compiled implicit isolate, not raw source action connect")
	}
}

func TestLoadWithoutDeclaredIsolatesIgnoresStaleThisChoice(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-no-isolates-stale-this")
	if err := s.LoadFileContentWithIsolate("client_server_example.ivy", readClientServerExample(t), "this"); err != nil {
		t.Fatalf("LoadFileContentWithIsolate: %v", err)
	}
	if s.ActiveIsolate != NoIsolatesFoundChoice {
		t.Fatalf("ActiveIsolate = %q, want %q", s.ActiveIsolate, NoIsolatesFoundChoice)
	}
	if len(s.AvailableIsolates) != 1 || s.AvailableIsolates[0] != NoIsolatesFoundChoice {
		t.Fatalf("AvailableIsolates = %#v, want [%q]", s.AvailableIsolates, NoIsolatesFoundChoice)
	}
	if _, ok := s.CompiledModule.Actions.Get2("ext:connect"); !ok {
		t.Fatalf("no-isolates load should still compile Ivy's implicit ext:connect action")
	}
	if _, ok := s.CompiledModule.Actions.Get2("connect"); ok {
		t.Fatalf("no-isolates load should expose the compiled implicit isolate, not raw source action connect")
	}
}

func TestLoadWithDeclaredIsolatesIgnoresStaleUnknownChoice(t *testing.T) {
	const content = `#lang ivy1.7

isolate alpha = {
    type t
}

isolate beta = {
    type u
}
`
	s := NewSession(goivy.NewConfig(), "test-stale-unknown-isolate")
	if err := s.LoadFileContentWithIsolate("two_isolates.ivy", []byte(content), "dramc_nb2"); err != nil {
		t.Fatalf("LoadFileContentWithIsolate: %v", err)
	}
	if s.ActiveIsolate != "alpha" {
		t.Fatalf("ActiveIsolate = %q, want alpha", s.ActiveIsolate)
	}
	if len(s.AvailableIsolates) != 2 || s.AvailableIsolates[0] != "alpha" || s.AvailableIsolates[1] != "beta" {
		t.Fatalf("AvailableIsolates = %#v, want [alpha beta]", s.AvailableIsolates)
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
	s := NewSession(goivy.NewConfig(), "test-step-in-sheet")
	if err := s.LoadFileContent("cti_used_relation.ivy", []byte(ctiUsedRelationSample)); err != nil {
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
		t.Fatalf("state_0 actions do not include ext:connect execute_action: %#v", actions)
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

func TestTryConjecturePDRBrowsesSourceAndShowsConceptGraph(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-try-conjecture-pdr")
	if err := s.LoadFileContent("test.ivy", []byte(executeActionMenuSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()
	s.AGUI.SetMode(ModePDR)

	choices, err := s.ArgNodeAction("state_0", "try_conjecture_choices", map[string]interface{}{"sheet_id": "sheet-1"})
	if err != nil {
		t.Fatalf("try_conjecture_choices: %v", err)
	}
	conjChoices, ok := choices["choices"].([]ChoiceItem)
	if !ok || len(conjChoices) == 0 {
		t.Fatalf("expected conjecture choices, got %#v", choices["choices"])
	}

	result, err := s.ArgNodeAction("state_0", "try_conjecture", map[string]interface{}{
		"sheet_id":   "sheet-1",
		"conjecture": conjChoices[0].Value,
	})
	if err != nil {
		t.Fatalf("try_conjecture: %v", err)
	}
	if result["view"] != "concept" {
		t.Fatalf("try_conjecture view = %#v, want concept", result["view"])
	}
	if _, ok := result["concept"].(map[string]interface{}); !ok {
		t.Fatalf("try_conjecture result missing concept payload: %#v", result["concept"])
	}
	if result["source"] != string(executeActionMenuSample) {
		t.Fatalf("try_conjecture did not return source for browsing: %#v", result["source"])
	}
	if result["file"] != "test.ivy" {
		t.Fatalf("try_conjecture source file = %#v, want test.ivy", result["file"])
	}
	if line, ok := result["lineno"].(int); !ok || line <= 0 {
		t.Fatalf("try_conjecture lineno = %#v, want positive int", result["lineno"])
	}
	if s.AGUI.CurrentConceptGraph == nil || s.AGUI.CurrentConceptGraph.G() == nil {
		t.Fatal("try_conjecture did not install a current concept graph")
	}
	if got := len(s.AGUI.CurrentConceptGraph.G().InteractiveSess.SupposeConstraints); got == 0 {
		t.Fatal("try_conjecture did not add dual conjecture constraints to concept graph")
	}
}

func TestTryConjectureBoundedAndInductionReturnResultViewsAndSource(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode VerificationMode
	}{
		{name: "bounded", mode: ModeBounded},
		{name: "induction", mode: ModeInduction},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSession(goivy.NewConfig(), "test-try-conjecture-"+tc.name)
			if err := s.LoadFileContent("test.ivy", []byte(executeActionMenuSample)); err != nil {
				t.Fatalf("LoadFileContent: %v", err)
			}
			s.AG.AddInitialState(nil, nil)
			s.syncARGToGraph()
			s.AGUI.SetMode(tc.mode)

			choices, err := s.ArgNodeAction("state_0", "try_conjecture_choices", map[string]interface{}{"sheet_id": "sheet-1"})
			if err != nil {
				t.Fatalf("try_conjecture_choices: %v", err)
			}
			conjChoices, ok := choices["choices"].([]ChoiceItem)
			if !ok || len(conjChoices) == 0 {
				t.Fatalf("expected conjecture choices, got %#v", choices["choices"])
			}

			result, err := s.ArgNodeAction("state_0", "try_conjecture", map[string]interface{}{
				"sheet_id":   "sheet-1",
				"conjecture": conjChoices[0].Value,
			})
			if err != nil {
				t.Fatalf("try_conjecture: %v", err)
			}
			if result["mode"] != string(tc.mode) {
				t.Fatalf("mode = %#v, want %s", result["mode"], tc.mode)
			}
			if result["source"] != string(executeActionMenuSample) {
				t.Fatalf("try_conjecture did not return source for browsing: %#v", result["source"])
			}
			if _, ok := result["lineno"].(int); !ok {
				t.Fatalf("lineno = %#v, want int", result["lineno"])
			}
			if _, ok := result["reachable"].(bool); !ok {
				t.Fatalf("reachable = %#v, want bool", result["reachable"])
			}
			if _, ok := result["concept"]; ok {
				t.Fatalf("%s try_conjecture should not return a concept graph: %#v", tc.name, result["concept"])
			}
			switch result["view"] {
			case "message":
				if msg, _ := result["message"].(string); msg == "" {
					t.Fatalf("%s message view missing message: %#v", tc.name, result)
				}
				if _, ok := result["trace_arg"]; ok {
					t.Fatalf("%s message view should not include trace_arg: %#v", tc.name, result["trace_arg"])
				}
			case "trace":
				if _, ok := result["trace_arg"].(map[string]interface{}); !ok {
					t.Fatalf("%s trace view missing trace_arg: %#v", tc.name, result["trace_arg"])
				}
				if sheetID, _ := result["trace_sheet_id"].(string); sheetID == "" || sheetID == "sheet-1" {
					t.Fatalf("%s trace_sheet_id = %#v, want registered trace sheet", tc.name, result["trace_sheet_id"])
				}
			default:
				t.Fatalf("%s view = %#v, want message or trace", tc.name, result["view"])
			}
		})
	}
}

func TestCheckFailureCarriesTraceARGForViewAction(t *testing.T) {
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
	if _, err := be.Load(session["session_id"], "cti_used_relation.ivy", []byte(ctiUsedRelationSample), ""); err != nil {
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
		"action_name": "ext:bad",
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
	mod.Relations.Set(goivy.RelationKey("p", goivy.Boolean), goivy.Boolean)
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

func TestReachReportsEliminatedConjectures(t *testing.T) {
	mod := goivy.New()
	p := goivy.NewConst("p", goivy.Boolean)
	mod.Relations.Set(goivy.RelationKey("p", goivy.Boolean), goivy.Boolean)
	conj := goivy.NewClauses([]goivy.Expr{p}, nil, nil)
	mod.SetAttribute("__interp_conjs", []*goivy.Clauses{conj})

	pred := goivy.NewState(mod, goivy.TrueClauses(nil))
	pred.Unders = []*goivy.State{goivy.NewState(mod, goivy.TrueClauses(nil))}
	target := goivy.NewState(mod, goivy.TrueClauses(nil))
	target.Pred = pred
	target.Update = goivy.GetUpdate(goivy.NewAssignAction(p, goivy.False), &goivy.UpdateContext{
		Domain: mod,
		ActCfg: goivy.NewActionsConfig(),
	})
	ag := goivy.NewAnalysisGraph(mod)
	ag.Add(pred, nil)
	ag.Add(target, nil)

	ui := NewAnalysisGraphUI()
	ui.AG = ag
	ui.Mod = mod
	w := NewGraphWidget(StandardGraph([]string{"node"}, target))
	w.G().ParentState = target
	ui.CurrentConceptGraph = w

	s := NewSession(goivy.NewConfig(), "test-reach-eliminated-conjectures")
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
	eliminated, ok := result["eliminated_conjectures"].([]string)
	if !ok || len(eliminated) != 1 || !strings.Contains(eliminated[0], "p") {
		t.Fatalf("eliminated_conjectures = %#v, want [p]", result["eliminated_conjectures"])
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
	if result["view"] != "trace" {
		t.Fatalf("view = %#v, want trace for CTI BMC counterexample", result["view"])
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

func TestCTIStrengthenPreviewDoesNotAppendAndAcceptAppendsOnce(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-cti-strengthen-confirm")
	ui := NewCTIAnalysisGraphUI(goivy.New())
	widget := NewGraphWidget(StandardGraph([]string{"S"}, nil))
	ui.CurrentConceptGraph = widget
	s.CTIUI = ui

	before := len(ui.Conjectures)
	preview, err := s.ExecuteAction("cti_strengthen_preview", map[string]interface{}{
		"sheet_id": rootSheetID,
	})
	if err != nil {
		t.Fatalf("cti_strengthen_preview: %v", err)
	}
	if len(ui.Conjectures) != before {
		t.Fatalf("preview appended conjectures: got %d, want %d", len(ui.Conjectures), before)
	}
	previewConj, _ := preview["conjecture"].(string)
	if previewConj == "" {
		t.Fatalf("preview conjecture missing: %#v", preview)
	}

	result, err := s.ExecuteAction("cti_strengthen", map[string]interface{}{
		"sheet_id": rootSheetID,
	})
	if err != nil {
		t.Fatalf("cti_strengthen: %v", err)
	}
	if len(ui.Conjectures) != before+1 {
		t.Fatalf("strengthen conjecture count = %d, want %d", len(ui.Conjectures), before+1)
	}
	if result["conjecture"] != previewConj {
		t.Fatalf("strengthen conjecture = %#v, want preview %q", result["conjecture"], previewConj)
	}
}

func TestCTIMinimizeReportsBoundAndCoreFacts(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-cti-minimize-details")
	mod := goivy.New()
	ui := NewCTIAnalysisGraphUI(mod)
	ui.CurrentBound = 2
	widget := NewGraphWidget(StandardGraph([]string{"S"}, nil))
	domain, _ := argStepInTestDomainSetup()
	widget.G().InteractiveSess = NewConceptInteractiveSession(
		domain,
		goivy.True,
		goivy.True,
		nil,
		[]goivy.Expr{goivy.False, goivy.True},
		nil,
		nil,
		nil,
		false,
	)
	ui.CurrentConceptGraph = widget
	s.CTIUI = ui

	result, err := s.ExecuteAction("cti_minimize", map[string]interface{}{
		"sheet_id": rootSheetID,
	})
	if err != nil {
		t.Fatalf("cti_minimize: %v", err)
	}
	if result["bound"] != 2 {
		t.Fatalf("bound = %#v, want 2; result=%#v", result["bound"], result)
	}
	inputFacts, ok := result["input_facts"].([]string)
	if !ok || !stringSliceContains(inputFacts, "false") || !stringSliceContains(inputFacts, "true") {
		t.Fatalf("input_facts = %#v, want false and true", result["input_facts"])
	}
	coreFacts, ok := result["core_facts"].([]string)
	if !ok || !stringSliceContains(coreFacts, "false") || stringSliceContains(coreFacts, "true") {
		t.Fatalf("core_facts = %#v, want only false retained", result["core_facts"])
	}
	removedFacts, ok := result["removed_facts"].([]string)
	if !ok || !stringSliceContains(removedFacts, "true") || stringSliceContains(removedFacts, "false") {
		t.Fatalf("removed_facts = %#v, want only true removed", result["removed_facts"])
	}
	if message, _ := result["message"].(string); !strings.Contains(message, "BMC bound 2") || !strings.Contains(message, "kept 1 of 2") {
		t.Fatalf("message = %q, want visible bound/core detail", message)
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

func TestInductionCheckAcceptsRelationsToMinimizeOption(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-relations-to-minimize-option")
	if err := s.LoadFileContent("test.ivy", []byte(ctiUsedRelationSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	cr := s.RunCheckWithOptions("induction", CheckOptions{RelationsToMinimize: "q"})
	if cr.Result != "fail" {
		t.Fatalf("RunCheck induction result = %q, want fail; message=%s", cr.Result, cr.Message)
	}
	if s.CTIUI == nil {
		t.Fatal("CTIUI missing after load")
	}
	if s.CTIUI.RelationsToMinimize != "q" {
		t.Fatalf("RelationsToMinimize = %q, want q", s.CTIUI.RelationsToMinimize)
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

func TestCTIDiagramActionReportsPreStateContext(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-cti-diagram-pre-state")
	if err := s.LoadFileContent("test.ivy", []byte(ctiUsedRelationSample)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	cr := s.RunCheck("induction")
	if cr.Result != "fail" {
		t.Fatalf("RunCheck induction result = %q, want fail; message=%s", cr.Result, cr.Message)
	}
	result, err := s.ExecuteAction("diagram", map[string]interface{}{
		"sheet_id": rootSheetID,
		"ui_mode":  "cti",
	})
	if err != nil {
		t.Fatalf("diagram action: %v", err)
	}
	if result["cti_state_label"] != "CTI pre-state 0" {
		t.Fatalf("cti_state_label = %#v, want CTI pre-state 0; result=%#v", result["cti_state_label"], result)
	}
	concept, ok := result["concept"].(map[string]interface{})
	if !ok {
		t.Fatalf("concept payload missing or wrong type: %#v", result["concept"])
	}
	if concept["state_label"] != "CTI pre-state 0" {
		t.Fatalf("concept state_label = %#v, want CTI pre-state 0", concept["state_label"])
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
}

func TestConceptDomainSaveLoadReplaceActionsReturnSnapshots(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-domain-save-load-replace")
	if err := s.LoadFileContent("client_server_example.ivy", clientServerExampleWithIndividualC0(t)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()
	if _, err := s.AGUI.ViewState(0, "", false); err != nil {
		t.Fatalf("ViewState: %v", err)
	}

	diagram, err := s.ExecuteAction("diagram_domain", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("diagram_domain: %v", err)
	}
	if got := diagram["type"]; got == "diagram_domain_empty" {
		t.Fatalf("diagram_domain unexpectedly empty: %#v", diagram)
	}
	savedNodes := append([]string{}, s.AGUI.CurrentConceptGraph.G().ConceptSess.Domain.Nodes...)
	if len(savedNodes) == 0 {
		t.Fatalf("diagram_domain produced no saved nodes")
	}

	save, err := s.ExecuteAction("save_domain", map[string]interface{}{"sheet_id": rootSheetID, "name": "diagram"})
	if err != nil {
		t.Fatalf("save_domain: %v", err)
	}
	if got := save["type"]; got != "save_domain" {
		t.Fatalf("save_domain type = %v, want save_domain; result=%#v", got, save)
	}
	if got := save["name"]; got != "diagram" {
		t.Fatalf("save_domain name = %v, want diagram; result=%#v", got, save)
	}

	if _, err := s.ExecuteAction("reset_domain", map[string]interface{}{"sheet_id": rootSheetID}); err != nil {
		t.Fatalf("reset_domain: %v", err)
	}
	resetNodes := s.AGUI.CurrentConceptGraph.G().ConceptSess.Domain.Nodes
	if strings.Join(resetNodes, "\x00") == strings.Join(savedNodes, "\x00") {
		t.Fatalf("reset_domain left saved diagram nodes in place: %v", resetNodes)
	}

	load, err := s.ExecuteAction("load_domain", map[string]interface{}{"sheet_id": rootSheetID, "name": "diagram"})
	if err != nil {
		t.Fatalf("load_domain: %v", err)
	}
	if got := load["type"]; got != "load_domain" {
		t.Fatalf("load_domain type = %v, want load_domain; result=%#v", got, load)
	}
	if got := s.AGUI.CurrentConceptGraph.G().ConceptSess.Domain.Nodes; strings.Join(got, "\x00") != strings.Join(savedNodes, "\x00") {
		t.Fatalf("load_domain nodes = %v, want saved nodes %v", got, savedNodes)
	}
	loadConcept, ok := load["concept"].(map[string]interface{})
	if !ok {
		t.Fatalf("load_domain concept payload missing: %#v", load["concept"])
	}
	if stack, ok := loadConcept["graph_stack"].(map[string]interface{}); !ok || stack["can_undo"] != true {
		t.Fatalf("load_domain graph stack should expose undo after replacement: %#v", loadConcept["graph_stack"])
	}

	if _, err := s.ExecuteAction("reset_domain", map[string]interface{}{"sheet_id": rootSheetID}); err != nil {
		t.Fatalf("reset_domain before replace: %v", err)
	}
	replace, err := s.ExecuteAction("replace_domain", map[string]interface{}{"sheet_id": rootSheetID, "name": "diagram"})
	if err != nil {
		t.Fatalf("replace_domain: %v", err)
	}
	if got := replace["type"]; got != "replace_domain" {
		t.Fatalf("replace_domain type = %v, want replace_domain; result=%#v", got, replace)
	}
	if got := s.AGUI.CurrentConceptGraph.G().ConceptSess.Domain.Nodes; strings.Join(got, "\x00") != strings.Join(savedNodes, "\x00") {
		t.Fatalf("replace_domain nodes = %v, want saved nodes %v", got, savedNodes)
	}
	if _, ok := replace["concept"].(map[string]interface{}); !ok {
		t.Fatalf("replace_domain concept payload missing: %#v", replace["concept"])
	}
}

func TestSplatterActionUsesSignatureConstantsAndGraphStackUndoRedo(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-splatter-action-constants")
	if err := s.LoadFileContent("client_server_example.ivy", clientServerExampleWithConcreteIndividuals(t)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	w := s.ensureConceptGraphWidgetLocked()
	if w == nil || w.G() == nil {
		t.Fatalf("concept graph widget was not created")
	}

	splatter, err := s.ExecuteAction("splatter", map[string]interface{}{"concept": "client", "sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("splatter: %v", err)
	}
	if _, ok := splatter["concept"].(map[string]interface{}); !ok {
		t.Fatalf("splatter did not return concept snapshot: %#v", splatter)
	}
	if !w.GraphStack.CanUndo() {
		t.Fatalf("splatter did not create graph-stack undo state")
	}
	if !conceptSessionNodesContain(s.SimpleSess, "c0") || !conceptSessionNodesContain(s.SimpleSess, "c1") {
		t.Fatalf("SimpleSess nodes after splatter = %v, want c0/c1 split nodes", s.SimpleSess.Domain.Nodes)
	}

	if _, err := s.ExecuteAction("undo", map[string]interface{}{"sheet_id": rootSheetID}); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if !w.GraphStack.CanRedo() {
		t.Fatalf("undo did not create graph-stack redo state")
	}
	if conceptSessionNodesContain(s.SimpleSess, "c0") || conceptSessionNodesContain(s.SimpleSess, "c1") {
		t.Fatalf("SimpleSess nodes after undo = %v, want original unsplattered nodes", s.SimpleSess.Domain.Nodes)
	}

	if _, err := s.ExecuteAction("redo", map[string]interface{}{"sheet_id": rootSheetID}); err != nil {
		t.Fatalf("redo: %v", err)
	}
	if !w.GraphStack.CanUndo() || w.GraphStack.CanRedo() {
		t.Fatalf("redo graph stack canUndo=%v canRedo=%v, want true/false", w.GraphStack.CanUndo(), w.GraphStack.CanRedo())
	}
	if !conceptSessionNodesContain(s.SimpleSess, "c0") || !conceptSessionNodesContain(s.SimpleSess, "c1") {
		t.Fatalf("SimpleSess nodes after redo = %v, want c0/c1 split nodes", s.SimpleSess.Domain.Nodes)
	}
}

func TestEmptyActionUsesGraphStackUndoRedo(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-empty-action-graph-stack")
	if err := s.LoadFileContent("client_server_example.ivy", clientServerExampleWithConcreteIndividuals(t)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	w := s.ensureConceptGraphWidgetLocked()
	if w == nil || w.G() == nil {
		t.Fatalf("concept graph widget was not created")
	}

	empty, err := s.ExecuteAction("empty", map[string]interface{}{"concept": "client", "sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if _, ok := empty["concept"].(map[string]interface{}); !ok {
		t.Fatalf("empty did not return concept snapshot: %#v", empty)
	}
	if !w.GraphStack.CanUndo() {
		t.Fatalf("empty did not create graph-stack undo state")
	}
	if !s.SimpleSess.AbstractValue["node_info|none|client"] {
		t.Fatalf("SimpleSess abstract value after empty = %v, want client marked none", s.SimpleSess.AbstractValue)
	}

	if _, err := s.ExecuteAction("undo", map[string]interface{}{"sheet_id": rootSheetID}); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if !w.GraphStack.CanRedo() {
		t.Fatalf("undo did not create graph-stack redo state")
	}
	if s.SimpleSess.AbstractValue["node_info|none|client"] {
		t.Fatalf("SimpleSess abstract value after undo = %v, want client restored", s.SimpleSess.AbstractValue)
	}

	if _, err := s.ExecuteAction("redo", map[string]interface{}{"sheet_id": rootSheetID}); err != nil {
		t.Fatalf("redo: %v", err)
	}
	if !s.SimpleSess.AbstractValue["node_info|none|client"] {
		t.Fatalf("SimpleSess abstract value after redo = %v, want client marked none", s.SimpleSess.AbstractValue)
	}
}

func TestGoBackendConceptMaterializeNodeUsesGraphStackUndoRedo(t *testing.T) {
	cfg := goivy.NewConfig()
	be := NewGoBackend(cfg)
	defer be.Close()
	sessionJSON, err := be.NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	var session map[string]string
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		t.Fatalf("session json: %v", err)
	}
	sessionID := session["session_id"]
	if sessionID == "" {
		t.Fatalf("missing session_id in %s", sessionJSON)
	}
	if _, err := be.Load(sessionID, "client_server_example.ivy", clientServerExampleWithConcreteIndividuals(t), ""); err != nil {
		t.Fatalf("Load: %v", err)
	}

	materializedJSON, err := be.ConceptMaterialize(sessionID, ConceptMaterializeRequest{Concept: "client"})
	if err != nil {
		t.Fatalf("ConceptMaterialize: %v", err)
	}
	materialized := decodeMapJSON(t, materializedJSON)
	if got, _ := materialized["witness"].(string); got == "" {
		t.Fatalf("witness missing/empty; body=%s", materializedJSON)
	}
	materializedConcept := requireConceptPayload(t, materialized)
	if !conceptPayloadConceptNamesContain(materializedConcept, "=__c0") {
		t.Fatalf("materialized concept domain does not contain =__c0: %#v", materializedConcept["concept_domain"])
	}
	assertConceptGraphStack(t, materializedConcept, true, false)

	undoneJSON, err := be.ConceptUndo(sessionID)
	if err != nil {
		t.Fatalf("ConceptUndo: %v", err)
	}
	undoneConcept := requireConceptPayload(t, decodeMapJSON(t, undoneJSON))
	if conceptPayloadConceptNamesContain(undoneConcept, "=__c0") {
		t.Fatalf("undone concept domain still contains =__c0: %#v", undoneConcept["concept_domain"])
	}
	assertConceptGraphStack(t, undoneConcept, false, true)

	redoneJSON, err := be.Action(sessionID, "redo", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("redo action: %v", err)
	}
	redoneConcept := requireConceptPayload(t, decodeMapJSON(t, redoneJSON))
	if !conceptPayloadConceptNamesContain(redoneConcept, "=__c0") {
		t.Fatalf("redone concept domain does not contain =__c0: %#v", redoneConcept["concept_domain"])
	}
	assertConceptGraphStack(t, redoneConcept, true, false)
}

func TestGoBackendConceptMaterializeEdgeReturnsConceptFactAndGraphStackUndoRedo(t *testing.T) {
	cfg := goivy.NewConfig()
	be := NewGoBackend(cfg)
	defer be.Close()
	sessionJSON, err := be.NewSession(cfg)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	var session map[string]string
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		t.Fatalf("session json: %v", err)
	}
	sessionID := session["session_id"]
	if sessionID == "" {
		t.Fatalf("missing session_id in %s", sessionJSON)
	}
	if _, err := be.Load(sessionID, "client_server_example.ivy", clientServerExampleWithConcreteIndividuals(t), ""); err != nil {
		t.Fatalf("Load: %v", err)
	}

	materializedJSON, err := be.ConceptMaterialize(sessionID, ConceptMaterializeRequest{
		Type:     "edge",
		Relation: "link",
		Source:   "client",
		Target:   "server",
		Positive: true,
	})
	if err != nil {
		t.Fatalf("ConceptMaterialize edge: %v", err)
	}
	materialized := decodeMapJSON(t, materializedJSON)
	witnesses, ok := materialized["witnesses"].([]interface{})
	if !ok || len(witnesses) != 2 {
		t.Fatalf("witnesses = %#v, want two materialized endpoints; body=%s", materialized["witnesses"], materializedJSON)
	}
	materializedConcept := requireConceptPayload(t, materialized)
	if !conceptPayloadFactsContain(materializedConcept, "link") {
		t.Fatalf("materialized edge facts do not mention link: %#v", materializedConcept["facts"])
	}
	assertConceptGraphStack(t, materializedConcept, true, false)

	undoneJSON, err := be.Action(sessionID, "undo", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("undo action: %v", err)
	}
	undoneConcept := requireConceptPayload(t, decodeMapJSON(t, undoneJSON))
	if conceptPayloadFactsContain(undoneConcept, "link") {
		t.Fatalf("undone concept facts still mention link: %#v", undoneConcept["facts"])
	}
	assertConceptGraphStack(t, undoneConcept, false, true)

	redoneJSON, err := be.Action(sessionID, "redo", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("redo action: %v", err)
	}
	redoneConcept := requireConceptPayload(t, decodeMapJSON(t, redoneJSON))
	if !conceptPayloadFactsContain(redoneConcept, "link") {
		t.Fatalf("redone concept facts do not mention link: %#v", redoneConcept["facts"])
	}
	assertConceptGraphStack(t, redoneConcept, true, false)
}

func conceptSessionNodesContain(cs *ConceptSession, fragment string) bool {
	if cs == nil || cs.Domain == nil {
		return false
	}
	for _, node := range cs.Domain.Nodes {
		if strings.Contains(node, fragment) {
			return true
		}
	}
	return false
}

func decodeMapJSON(t *testing.T, data []byte) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("decode json %s: %v", data, err)
	}
	return result
}

func requireConceptPayload(t *testing.T, result map[string]interface{}) map[string]interface{} {
	t.Helper()
	concept, ok := result["concept"].(map[string]interface{})
	if !ok {
		t.Fatalf("concept payload missing/wrong type: %#v", result["concept"])
	}
	return concept
}

func conceptPayloadValuesContain(concept map[string]interface{}, fragment string) bool {
	elements, _ := concept["elements"].([]interface{})
	for _, raw := range elements {
		elem, _ := raw.(map[string]interface{})
		data, _ := elem["data"].(map[string]interface{})
		for _, key := range []string{"id", "obj", "label", "source", "target"} {
			if strings.Contains(fmt.Sprint(data[key]), fragment) {
				return true
			}
		}
	}
	return false
}

func conceptPayloadFactsContain(concept map[string]interface{}, fragments ...string) bool {
	facts, _ := concept["facts"].([]interface{})
	for _, raw := range facts {
		fact, _ := raw.(map[string]interface{})
		text := fmt.Sprint(fact["text"])
		matched := true
		for _, fragment := range fragments {
			if !strings.Contains(text, fragment) {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func conceptPayloadConceptNamesContain(concept map[string]interface{}, name string) bool {
	domain, _ := concept["concept_domain"].(map[string]interface{})
	concepts, _ := domain["concepts"].(map[string]interface{})
	_, ok := concepts[name]
	return ok
}

func assertConceptGraphStack(t *testing.T, concept map[string]interface{}, wantUndo, wantRedo bool) {
	t.Helper()
	stack, ok := concept["graph_stack"].(map[string]interface{})
	if !ok {
		t.Fatalf("graph_stack missing/wrong type: %#v", concept["graph_stack"])
	}
	if got, _ := stack["can_undo"].(bool); got != wantUndo {
		t.Fatalf("graph_stack can_undo = %v, want %v; stack=%#v", got, wantUndo, stack)
	}
	if got, _ := stack["can_redo"].(bool); got != wantRedo {
		t.Fatalf("graph_stack can_redo = %v, want %v; stack=%#v", got, wantRedo, stack)
	}
}

func TestDiagramDomainWithoutConstantsLeavesConceptGraphAlone(t *testing.T) {
	s := NewSession(goivy.NewConfig(), "test-empty-diagram-domain")
	if err := s.LoadFileContent("client_server_example.ivy", readClientServerExample(t)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()
	if _, err := s.AGUI.ViewState(0, "", false); err != nil {
		t.Fatalf("ViewState: %v", err)
	}

	before := append([]string{}, s.AGUI.CurrentConceptGraph.G().ConceptSess.Domain.Nodes...)
	diagram, err := s.ExecuteAction("diagram_domain", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("diagram_domain: %v", err)
	}
	if got := diagram["type"]; got != "diagram_domain_empty" {
		t.Fatalf("diagram_domain type = %v, want diagram_domain_empty; result=%#v", got, diagram)
	}
	if got := diagram["status"]; got != "warning" {
		t.Fatalf("diagram_domain status = %v, want warning; result=%#v", got, diagram)
	}
	wantMessage := "no first-order constants in 'client_server_example.ivy' found. Diagram Domain would give an empty graph. Leaving existing graph alone."
	if got := diagram["message"]; got != wantMessage {
		t.Fatalf("diagram_domain message = %q, want %q", got, wantMessage)
	}
	if _, ok := diagram["concept"]; ok {
		t.Fatalf("diagram_domain_empty should not return a concept snapshot: %#v", diagram["concept"])
	}
	after := s.AGUI.CurrentConceptGraph.G().ConceptSess.Domain.Nodes
	if strings.Join(after, "\x00") != strings.Join(before, "\x00") {
		t.Fatalf("diagram_domain_empty changed concept graph nodes: before=%v after=%v", before, after)
	}
}

func TestDiagramDomainIndividualConstantMatchesPython(t *testing.T) {
	content := clientServerExampleWithIndividualC0(t)
	wantNodes := pythonIvyDiagramDomainNodes(t, content)
	if len(wantNodes) == 0 {
		t.Fatalf("python Diagram Domain returned no nodes for client_server_example.ivy plus individual c0 : client")
	}

	s := NewSession(goivy.NewConfig(), "test-diagram-domain-c0")
	if err := s.LoadFileContent("client_server_example.ivy", content); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	s.AG.AddInitialState(nil, nil)
	s.syncARGToGraph()
	if _, err := s.AGUI.ViewState(0, "", false); err != nil {
		t.Fatalf("ViewState: %v", err)
	}

	diagram, err := s.ExecuteAction("diagram_domain", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("diagram_domain: %v", err)
	}
	if got := diagram["type"]; got == "diagram_domain_empty" {
		t.Fatalf("Go Diagram Domain was empty, but Python returned nodes %v; result=%#v", wantNodes, diagram)
	}
	gotNodes := s.AGUI.CurrentConceptGraph.G().ConceptSess.Domain.Nodes
	if strings.Join(gotNodes, "\x00") != strings.Join(wantNodes, "\x00") {
		t.Fatalf("Go Diagram Domain nodes = %v, want Python nodes %v", gotNodes, wantNodes)
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
	content := []byte(ctiUsedRelationSample)

	s := NewSession(goivy.NewConfig(), "test-view-source")
	if err := s.LoadFileContent("cti_used_relation.ivy", content); err != nil {
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
