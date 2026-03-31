//go:build web

package webui

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/module"
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
	cfg := module.NewConfig()
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
// The concept view's "elements" (CyElements) should include at least one node.

func TestRegression_StateGraphHasInitialNode(t *testing.T) {
	srv, id := loadClientServer(t)
	concept := getConceptJSON(t, srv, id)

	// Extract "elements" from concept JSON — these are the CyElements
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
