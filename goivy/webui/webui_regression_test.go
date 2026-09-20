//go:build web

package webui

import (
	"encoding/json"
	goivy "github.com/glycerine/ivy/goivy"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const clientServerIvy = `#lang ivy1.7

type client
type server

relation link(X:client, Y:server)
relation semaphore(X:server)

after init {
    link(X,Y) := false;
    semaphore(Y) := true
}

action connect(x:client, y:server) = {
    require semaphore(y);
    link(x,y) := true;
    semaphore(y) := false
}

action disconnect(x:client, y:server) = {
    require link(x,y);
    link(x,y) := false;
    semaphore(y) := true
}

export connect
export disconnect

conjecture link(X,Y) -> ~semaphore(Y)
conjecture ~link(X,Y) | ~link(X,Z) | Y = Z
`

// loadClientServer creates a session, uploads the client-server example
// via LoadFileContent, and returns the session ID and server.
func loadClientServer(t *testing.T) (*Server, string) {
	t.Helper()
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)

	// Use LoadFileContent directly on the session (multipart upload path).
	sess, err := srv.backend.(*GoBackend).getSession(id)
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}
	if err := sess.LoadFileContent("client_server.ivy", []byte(clientServerIvy)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	return srv, id
}

func loadTutorialClientServer(t *testing.T) (*Server, string) {
	t.Helper()
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	id := createSession(t, srv)

	sess, err := srv.backend.(*GoBackend).getSession(id)
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}
	if err := sess.LoadFileContent("client_server_example.ivy", readClientServerExample(t)); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	return srv, id
}

// getConceptJSON calls GET /api/session/{id}/concept and returns the parsed JSON.
func getConceptJSON(t *testing.T, srv *Server, id string) map[string]interface{} {
	t.Helper()
	w := doReq(t, srv, "GET", "/api/session/"+id+"/concept", "")
	if w.Code != 200 {
		t.Fatalf("concept: status %d, body: %s", w.Code, w.Body.String())
	}
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("concept json parse: %v\nbody: %s", err, w.Body.String())
	}
	return m
}

func getConceptNodeJSON(t *testing.T, srv *Server, id, nodeID string) map[string]interface{} {
	t.Helper()
	w := doReq(t, srv, "GET", "/api/session/"+id+"/concept?node="+nodeID, "")
	if w.Code != 200 {
		t.Fatalf("concept %s: status %d, body: %s", nodeID, w.Code, w.Body.String())
	}
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("concept %s json parse: %v\nbody: %s", nodeID, err, w.Body.String())
	}
	return m
}

func conceptEdgeCount(concept map[string]interface{}, edgeName string) int {
	elements, _ := concept["elements"].([]interface{})
	n := 0
	for _, raw := range elements {
		elem, _ := raw.(map[string]interface{})
		if elem["group"] != "edges" {
			continue
		}
		data, _ := elem["data"].(map[string]interface{})
		if data["obj"] == edgeName || data["label"] == edgeName {
			n++
		}
	}
	return n
}

func conceptToggleValue(concept map[string]interface{}, kind, name, className string) (bool, bool) {
	toggles, _ := concept["toggles"].(map[string]interface{})
	group, _ := toggles[kind].(map[string]interface{})
	entry, _ := group[name].(map[string]interface{})
	val, ok := entry[className].(bool)
	return val, ok
}

func setConceptToggle(t *testing.T, srv *Server, id, edge, displayClass string, value bool) {
	t.Helper()
	body := `{"edge":"` + edge + `","display_class":"` + displayClass + `","value":` + strconv.FormatBool(value) + `}`
	w := doReq(t, srv, "POST", "/api/session/"+id+"/toggles", body)
	if w.Code != 200 {
		t.Fatalf("set toggle: status %d, body: %s", w.Code, w.Body.String())
	}
}

func runTutorialInductionCheck(t *testing.T, srv *Server, id string) map[string]interface{} {
	t.Helper()
	w := doReq(t, srv, "POST", "/api/session/"+id+"/check", `{"mode":"induction"}`)
	if w.Code != 200 {
		t.Fatalf("check: status %d, body: %s", w.Code, w.Body.String())
	}
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("check json parse: %v\nbody: %s", err, w.Body.String())
	}
	return m
}

// --- Regression 1: Concept view must show BOTH "client" and "server" ---
// The concept graph nodes come from the sorts declared in the .ivy file.
// The client-server example has: type client, type server.
// Both must appear as concept nodes.

func TestRegression_ConceptViewShowsBothSorts(t *testing.T) {
	srv, id := loadClientServer(t)
	concept := getConceptJSON(t, srv, id)

	// Extract the "nodes" array from the concept JSON.
	nodesRaw, ok := concept["nodes"]
	if !ok {
		t.Fatal("concept JSON missing 'nodes' key")
	}
	nodesArr, ok := nodesRaw.([]interface{})
	if !ok {
		t.Fatalf("nodes is not array, got %T", nodesRaw)
	}
	nodeNames := make([]string, len(nodesArr))
	for i, n := range nodesArr {
		nodeNames[i], _ = n.(string)
	}
	sort.Strings(nodeNames)

	// Both "client" and "server" must be present.
	hasClient := false
	hasServer := false
	for _, n := range nodeNames {
		if n == "client" {
			hasClient = true
		}
		if n == "server" {
			hasServer = true
		}
	}
	if !hasClient {
		t.Errorf("concept nodes missing 'client', got: %v", nodeNames)
	}
	if !hasServer {
		t.Errorf("concept nodes missing 'server', got: %v", nodeNames)
	}
}

// --- Regression 2: Relation list must show "link(X,Y)" not just "link" ---
// The relations list should include parameter names for display.

func TestRegression_RelationListShowsParams(t *testing.T) {
	srv, id := loadClientServer(t)
	concept := getConceptJSON(t, srv, id)

	// Extract "relations" from concept JSON.
	relsRaw, ok := concept["relations"]
	if !ok {
		t.Fatal("concept JSON missing 'relations' key")
	}
	relsArr, ok := relsRaw.([]interface{})
	if !ok {
		t.Fatalf("relations is not array, got %T", relsRaw)
	}
	var relNames []string
	for _, r := range relsArr {
		if s, ok := r.(string); ok {
			relNames = append(relNames, s)
		}
	}

	// "link" should appear as "link(X,Y)" (with parameters), not bare "link".
	foundLinkWithParams := false
	foundBareLinkOnly := false
	for _, r := range relNames {
		if strings.Contains(r, "link") && strings.Contains(r, "(") {
			foundLinkWithParams = true
		}
		if r == "link" {
			foundBareLinkOnly = true
		}
	}
	if foundBareLinkOnly && !foundLinkWithParams {
		t.Errorf("relation 'link' shown without params; expected 'link(X,Y)' but got: %v", relNames)
	}
	if !foundLinkWithParams && !foundBareLinkOnly {
		t.Errorf("relation 'link' not found at all in relations list: %v", relNames)
	}
}

// --- Regression 3: State graph must have initial state node (the "0" circle) ---
// After loading a file, the ARG (analysis graph) should have at least one state.
// The concept view's "elements" (WebUICyElements) should include at least one node.

func TestRegression_StateGraphHasInitialNode(t *testing.T) {
	srv, id := loadClientServer(t)
	concept := getConceptJSON(t, srv, id)

	// Extract "elements" from concept JSON — these are the WebUICyElements
	// that the graph renderer displays.
	elemsRaw, ok := concept["elements"]
	if !ok {
		t.Fatal("concept JSON missing 'elements' key")
	}
	elemsArr, ok := elemsRaw.([]interface{})
	if !ok {
		t.Fatalf("elements is not array, got %T", elemsRaw)
	}

	// There should be at least one element (the initial state node).
	if len(elemsArr) == 0 {
		t.Error("elements array is empty — expected at least one node (the initial state '0')")
	}

	// Check that at least one element is a "nodes" group element.
	hasNodeGroup := false
	for _, e := range elemsArr {
		em, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		if em["group"] == "nodes" {
			hasNodeGroup = true
			break
		}
	}
	if !hasNodeGroup {
		t.Error("no element with group='nodes' found — the state circle should exist")
	}
}

func TestRegression_ConceptRouteUsesSelectedARGState(t *testing.T) {
	srv, id := loadClientServer(t)
	sess, err := srv.backend.(*GoBackend).getSession(id)
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}

	sess.AG.Add(goivy.NewState(sess.CompiledModule, goivy.TrueClauses(nil)), nil)
	sess.AG.Add(goivy.NewState(sess.CompiledModule, goivy.FalseClauses(nil)), nil)
	sess.syncARGToGraph()

	state0 := getConceptNodeJSON(t, srv, id, "state_0")
	state1 := getConceptNodeJSON(t, srv, id, "state_1")

	if state0["selected_node"] != "state_0" {
		t.Fatalf("state_0 selected_node = %v, want state_0", state0["selected_node"])
	}
	if state1["selected_node"] != "state_1" {
		t.Fatalf("state_1 selected_node = %v, want state_1", state1["selected_node"])
	}
	if state0["state_label"] == state1["state_label"] {
		t.Fatalf("state labels should differ, got %v and %v", state0["state_label"], state1["state_label"])
	}
	if reflect.DeepEqual(state0["abstract_value"], state1["abstract_value"]) {
		t.Fatalf("selected ARG states returned identical abstract_value maps")
	}
}

func TestRegression_TutorialStateOneConceptGraphShowsBothConcreteLinks(t *testing.T) {
	srv, id := loadTutorialClientServer(t)

	check := runTutorialInductionCheck(t, srv, id)
	if check["result"] != "fail" {
		t.Fatalf("tutorial induction check result = %#v, want fail; check=%#v", check["result"], check)
	}

	setConceptToggle(t, srv, id, "link", EdgeDisplayAllToAll, true)
	state1 := getConceptNodeJSON(t, srv, id, "state_1")
	if state1["selected_node"] != "state_1" {
		t.Fatalf("state_1 selected_node = %#v, want state_1", state1["selected_node"])
	}
	graph, _ := state1["graph"].(map[string]interface{})
	graphState, _ := graph["state"].(string)
	for _, want := range []string{"link(0,0) = true", "link(1,0) = true"} {
		if !strings.Contains(graphState, want) {
			t.Fatalf("state_1 graph state missing %q:\n%s", want, graphState)
		}
	}
	if got := conceptEdgeCount(state1, "link"); got != 2 {
		t.Fatalf("state_1 rendered %d link edges, want 2; graph_state=%q elements=%#v", got, graphState, state1["elements"])
	}
}

func TestRegression_ConceptTogglesAreBackendOwnedAndFilterRendering(t *testing.T) {
	srv, id := loadClientServer(t)

	initial := getConceptJSON(t, srv, id)
	if got := conceptEdgeCount(initial, "link"); got != 0 {
		t.Fatalf("link edge rendered before backend checkbox was enabled: %d", got)
	}
	if val, ok := conceptToggleValue(initial, "edges", "link", EdgeDisplayUnknown); !ok || val {
		t.Fatalf("initial link ? toggle = (%v,%v), want present false", val, ok)
	}

	setConceptToggle(t, srv, id, "link", EdgeDisplayUnknown, true)
	shown := getConceptJSON(t, srv, id)
	if got := conceptEdgeCount(shown, "link"); got == 0 {
		t.Fatalf("link edge did not render after enabling backend edge_unknown checkbox")
	}
	if val, ok := conceptToggleValue(shown, "edges", "link", EdgeDisplayUnknown); !ok || !val {
		t.Fatalf("shown link ? toggle = (%v,%v), want present true", val, ok)
	}
}

func TestRegression_ConceptTogglesRoundTripThroughUndoRedo(t *testing.T) {
	srv, id := loadClientServer(t)
	sess, err := srv.backend.(*GoBackend).getSession(id)
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}
	sess.AG.Add(goivy.NewState(sess.CompiledModule, goivy.TrueClauses(nil)), nil)
	sess.syncARGToGraph()

	_ = getConceptNodeJSON(t, srv, id, "state_0")
	setConceptToggle(t, srv, id, "link", EdgeDisplayUnknown, true)
	sess.AGUI.CurrentConceptGraph.Checkpoint(false)
	setConceptToggle(t, srv, id, "link", EdgeDisplayUnknown, false)

	w := doReq(t, srv, "POST", "/api/session/"+id+"/action", `{"action":"undo"}`)
	if w.Code != 200 {
		t.Fatalf("undo: status %d, body: %s", w.Code, w.Body.String())
	}
	undone := getConceptNodeJSON(t, srv, id, "state_0")
	if val, ok := conceptToggleValue(undone, "edges", "link", EdgeDisplayUnknown); !ok || !val {
		t.Fatalf("undo link ? toggle = (%v,%v), want present true", val, ok)
	}

	w = doReq(t, srv, "POST", "/api/session/"+id+"/action", `{"action":"redo"}`)
	if w.Code != 200 {
		t.Fatalf("redo: status %d, body: %s", w.Code, w.Body.String())
	}
	redone := getConceptNodeJSON(t, srv, id, "state_0")
	if val, ok := conceptToggleValue(redone, "edges", "link", EdgeDisplayUnknown); !ok || val {
		t.Fatalf("redo link ? toggle = (%v,%v), want present false", val, ok)
	}
}

func TestRegression_ConceptFactsAreBackendOwnedAndSelectable(t *testing.T) {
	srv, id := loadClientServer(t)
	sess, err := srv.backend.(*GoBackend).getSession(id)
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}

	S := mkSort("client")
	first := mkEq(mkConst("a", S), mkConst("b", S))
	second := mkEq(mkConst("c", S), mkConst("d", S))
	sess.mu.Lock()
	widget := sess.ensureConceptGraphWidgetLocked()
	if widget.G().InteractiveSess == nil {
		widget.G().InteractiveSess = NewConceptInteractiveSession(
			NewCDConceptDomain(nil, nil, nil), nil, nil, nil, nil, nil, nil, nil, false,
		)
	}
	widget.G().SetFactsExpr([]goivy.Expr{first, second})
	sess.mu.Unlock()

	concept := getConceptJSON(t, srv, id)
	factsRaw, ok := concept["facts"].([]interface{})
	if !ok || len(factsRaw) != 2 {
		t.Fatalf("concept facts missing or wrong length: %#v", concept["facts"])
	}
	firstFact, _ := factsRaw[0].(map[string]interface{})
	if selected, _ := firstFact["selected"].(bool); !selected {
		t.Fatalf("first fact should default selected: %#v", firstFact)
	}

	w := doReq(t, srv, "POST", "/api/session/"+id+"/action", `{"action":"set_fact_selection","args":{"index":0,"selected":false}}`)
	if w.Code != 200 {
		t.Fatalf("set_fact_selection: status %d, body: %s", w.Code, w.Body.String())
	}
	w = doReq(t, srv, "POST", "/api/session/"+id+"/action", `{"action":"get_active_facts","args":{}}`)
	if w.Code != 200 {
		t.Fatalf("get_active_facts: status %d, body: %s", w.Code, w.Body.String())
	}
	var active struct {
		Facts []string `json:"facts"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &active); err != nil {
		t.Fatalf("active facts json: %v", err)
	}
	if len(active.Facts) != 1 || active.Facts[0] != second.String() {
		t.Fatalf("expected only second fact active, got %#v", active.Facts)
	}
}
