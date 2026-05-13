package webui

// Payload conformance tests — lock the JSON wire format that the TypeScript
// data model in frontend/src/models/uiDataModel.ts depends on.
//
// If any of these tests break, update uiDataModel.ts to match.

import (
	"encoding/json"
	"fmt"
	"testing"

	goivy "github.com/glycerine/ivy/goivy"
)

// unmarshalKeys returns the top-level keys of a JSON object.
func unmarshalKeys(t *testing.T, data []byte) map[string]json.RawMessage {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshalKeys: %v", err)
	}
	return m
}

func mustCanonicalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := canonicalJSON(v)
	if err != nil {
		t.Fatalf("canonicalJSON: %v", err)
	}
	return b
}

// TestARGPayloadStructure verifies WebUIARGPayload produces the flat shape
// that TypeScript ARGSnapshot expects.
func TestARGPayloadStructure(t *testing.T) {
	state := goivy.NewAnalysisGraphState()
	state.States = []goivy.ARGNode{
		{ID: 0, Label: "init", IsBottom: false, Info: ""},
	}
	state.Transitions = []goivy.ARGTransition{
		{SourceID: 0, TargetID: 1, Label: "step", IsJoin: false},
	}
	state.Covering = []goivy.ARGCover{
		{CoveredID: 2, CoveringID: 0},
	}

	cy := goivy.RenderARG(state)
	payload := WebUIARGPayload(state, cy)
	data := mustCanonicalJSON(t, payload)
	keys := unmarshalKeys(t, data)

	required := []string{"analysis_graph_state", "elements"}
	for _, k := range required {
		if _, ok := keys[k]; !ok {
			t.Errorf("ARG payload missing key %q", k)
		}
	}
	if _, has := keys["analysis_graph"]; has {
		t.Error("ARG payload must NOT contain 'analysis_graph' key (never sent to browser)")
	}

	// analysis_graph_state must have states/transitions/covering sub-keys.
	var ags map[string]json.RawMessage
	if err := json.Unmarshal(keys["analysis_graph_state"], &ags); err != nil {
		t.Fatalf("analysis_graph_state is not an object: %v", err)
	}
	for _, sub := range []string{"states", "transitions", "covering"} {
		if _, ok := ags[sub]; !ok {
			t.Errorf("analysis_graph_state missing sub-key %q", sub)
		}
	}

	// elements must be a JSON array.
	var elems []json.RawMessage
	if err := json.Unmarshal(keys["elements"], &elems); err != nil {
		t.Errorf("'elements' is not a JSON array: %v", err)
	}
}

// TestCyElementsNodeIDNotSerialized verifies NodeID/EdgeID are never sent.
func TestCyElementsNodeIDNotSerialized(t *testing.T) {
	cy := goivy.NewCyElements()
	cy.AddNode("state_0", "init", nil, "", nil, nil, "ellipse")
	// NodeID map is now populated internally.

	data := mustCanonicalJSON(t, cy)
	keys := unmarshalKeys(t, data)

	for _, forbidden := range []string{"node_id", "NodeID", "edge_id", "EdgeID"} {
		if _, ok := keys[forbidden]; ok {
			t.Errorf("CyElements JSON must not contain key %q (tagged json:\"-\")", forbidden)
		}
	}
	if _, ok := keys["elements"]; !ok {
		t.Error("CyElements JSON must contain 'elements' key")
	}
}

// TestConceptGraphStackPayload verifies the 4-field contract.
func TestConceptGraphStackPayload(t *testing.T) {
	// nil stack → zero values
	nilPayload := conceptGraphStackPayload(nil)
	data := mustCanonicalJSON(t, nilPayload)
	keys := unmarshalKeys(t, data)
	for _, k := range []string{"can_undo", "can_redo", "undo_depth", "redo_depth"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("stack payload (nil) missing key %q", k)
		}
	}
	for _, forbidden := range []string{"current", "undo_stack", "redoStack", "Current"} {
		if _, ok := keys[forbidden]; ok {
			t.Errorf("stack payload must not contain key %q", forbidden)
		}
	}

	// live stack with 2 undo entries
	g := NewGraph([]string{"node"}, nil)
	gs := NewGraphStack(g)
	gs.Checkpoint(false)
	gs.Checkpoint(false)

	livePayload := conceptGraphStackPayload(gs)
	liveData := mustCanonicalJSON(t, livePayload)
	liveKeys := unmarshalKeys(t, liveData)

	var undoDepth int
	if err := json.Unmarshal(liveKeys["undo_depth"], &undoDepth); err != nil {
		t.Fatalf("undo_depth not int: %v", err)
	}
	if undoDepth != 2 {
		t.Errorf("undo_depth: got %d, want 2", undoDepth)
	}

	var canUndo bool
	if err := json.Unmarshal(liveKeys["can_undo"], &canUndo); err != nil {
		t.Fatalf("can_undo not bool: %v", err)
	}
	if !canUndo {
		t.Error("can_undo should be true with 2 undo entries")
	}
}

// TestConceptGraphPayload verifies conceptGraphPayload required keys.
func TestConceptGraphPayload(t *testing.T) {
	// nil graph → sentinel zero-value payload
	nilPayload := conceptGraphPayload(nil, nil)
	data := mustCanonicalJSON(t, nilPayload)
	keys := unmarshalKeys(t, data)
	for _, k := range []string{"sorts", "concept_session", "display_checkboxes", "graph_stack"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("conceptGraphPayload(nil) missing key %q", k)
		}
	}

	// live graph
	g := NewGraph([]string{"node", "edge"}, nil)
	gs := NewGraphStack(g)
	livePayload := conceptGraphPayload(g, gs)
	liveData := mustCanonicalJSON(t, livePayload)
	liveKeys := unmarshalKeys(t, liveData)
	for _, k := range []string{"sorts", "concept_session", "display_checkboxes", "graph_stack", "attributes", "state", "concrete"} {
		if _, ok := liveKeys[k]; !ok {
			t.Errorf("conceptGraphPayload(live) missing key %q", k)
		}
	}
}

// TestConceptInteractiveSessionPayloadTagValues verifies the abstract_value
// is serialized as []TagValue (not a flat map) and that axioms/cache are present.
func TestConceptInteractiveSessionPayloadTagValues(t *testing.T) {
	// nil CIS → empty payload with required keys
	nilPayload := conceptInteractiveSessionPayload(nil)
	nilData := mustCanonicalJSON(t, nilPayload)
	nilKeys := unmarshalKeys(t, nilData)
	for _, k := range []string{"abstract_value", "goal_constraints", "suppose_constraints", "undo_depth", "redo_depth"} {
		if _, ok := nilKeys[k]; !ok {
			t.Errorf("CIS payload(nil) missing key %q", k)
		}
	}

	// CIS with known TagValue
	cis := &ConceptInteractiveSession{
		AbstractValue: []TagValue{
			{Tag: Tag{"node_info", "at_least_one", "client"}, Value: true},
			{Tag: Tag{"edge_info", "all_to_all", "link", "client", "client"}, Value: false},
		},
		GoalConstraints:    []goivy.Expr{},
		SupposeConstraints: []goivy.Expr{},
		Cache:              map[string]bool{"node_info|none|server": false},
	}

	payload := conceptInteractiveSessionPayload(cis)
	data := mustCanonicalJSON(t, payload)
	keys := unmarshalKeys(t, data)

	// abstract_value must be a JSON array, not a map.
	var avArr []json.RawMessage
	if err := json.Unmarshal(keys["abstract_value"], &avArr); err != nil {
		t.Fatalf("abstract_value is not a JSON array: %v", err)
	}
	if len(avArr) != 2 {
		t.Errorf("abstract_value array length: got %d, want 2", len(avArr))
	}

	// Each entry must have Tag (array) and Value (bool).
	var entry map[string]json.RawMessage
	if err := json.Unmarshal(avArr[0], &entry); err != nil {
		t.Fatalf("abstract_value[0] is not an object: %v", err)
	}
	if _, ok := entry["Tag"]; !ok {
		t.Error("abstract_value entry missing 'Tag' field")
	}
	if _, ok := entry["Value"]; !ok {
		t.Error("abstract_value entry missing 'Value' field")
	}

	// axioms and cache must be present (Change C).
	if _, ok := keys["axioms"]; !ok {
		t.Error("CIS payload missing 'axioms' key")
	}
	if _, ok := keys["cache"]; !ok {
		t.Error("CIS payload missing 'cache' key")
	}

	// cache must be a bool map.
	var cacheMap map[string]bool
	if err := json.Unmarshal(keys["cache"], &cacheMap); err != nil {
		t.Errorf("cache is not a bool map: %v", err)
	}
	if cacheMap["node_info|none|server"] != false {
		t.Error("cache value mismatch")
	}

	// undo_depth and redo_depth present as ints.
	for _, k := range []string{"undo_depth", "redo_depth"} {
		var n int
		if err := json.Unmarshal(keys[k], &n); err != nil {
			t.Errorf("%s is not an int: %v", k, err)
		}
	}
}

// TestCISPayloadNoDomainKey verifies the CIS payload never sends a "domain" key.
// CDConceptDomain has unexported fields and serializes to {}, so the key was removed.
func TestCISPayloadNoDomainKey(t *testing.T) {
	// nil CIS
	nilPayload := conceptInteractiveSessionPayload(nil)
	nilData := mustCanonicalJSON(t, nilPayload)
	nilKeys := unmarshalKeys(t, nilData)
	if _, ok := nilKeys["domain"]; ok {
		t.Error("nil CIS payload must NOT contain 'domain' key")
	}

	// live CIS
	cis := &ConceptInteractiveSession{
		AbstractValue:      []TagValue{},
		GoalConstraints:    []goivy.Expr{},
		SupposeConstraints: []goivy.Expr{},
	}
	livePayload := conceptInteractiveSessionPayload(cis)
	liveData := mustCanonicalJSON(t, livePayload)
	liveKeys := unmarshalKeys(t, liveData)
	if _, ok := liveKeys["domain"]; ok {
		t.Error("live CIS payload must NOT contain 'domain' key")
	}
	// Required keys remain present.
	for _, k := range []string{"abstract_value", "goal_constraints", "suppose_constraints", "undo_depth", "redo_depth", "axioms", "cache", "state", "info"} {
		if _, ok := liveKeys[k]; !ok {
			t.Errorf("CIS payload missing required key %q", k)
		}
	}
}

// TestARGNodeInfoIsFormula verifies ArtToGraphState populates ARGNode.Info from
// state.Clauses.String() rather than the hardcoded "State N" fallback.
func TestARGNodeInfoIsFormula(t *testing.T) {
	st := &goivy.State{
		ID:      7,
		Clauses: goivy.TrueClauses(nil),
	}
	ag := &goivy.AnalysisGraph{States: []*goivy.State{st}}
	gs := ArtToGraphState(ag)

	if len(gs.States) != 1 {
		t.Fatalf("expected 1 ARGNode, got %d", len(gs.States))
	}
	info := gs.States[0].Info
	if info == "" {
		t.Error("ARGNode.Info must not be empty when Clauses is non-nil")
	}
	fallback := fmt.Sprintf("State %d", st.ID)
	if info == fallback {
		t.Errorf("ARGNode.Info must carry Clauses string, got fallback %q", info)
	}
}

// TestFullARGNodeUniverse verifies argNodeUniverse converts map[string][]Expr correctly.
func TestFullARGNodeUniverse(t *testing.T) {
	nodeSort := &goivy.UninterpretedSort{Name: "node"}
	n0 := goivy.NewConst("n0", nodeSort)
	n1 := goivy.NewConst("n1", nodeSort)
	st := &goivy.State{
		Universe: map[string][]goivy.Expr{"node": {n0, n1}},
	}
	u := argNodeUniverse(st)
	if u == nil {
		t.Fatal("argNodeUniverse: expected non-nil map")
	}
	nodes, ok := u["node"]
	if !ok {
		t.Fatal("argNodeUniverse: missing 'node' key")
	}
	if len(nodes) != 2 {
		t.Errorf("argNodeUniverse: got %d entries, want 2", len(nodes))
	}
}

// TestFullARGNodeUniverseNil verifies argNodeUniverse returns nil when Universe is nil.
func TestFullARGNodeUniverseNil(t *testing.T) {
	st := &goivy.State{}
	u := argNodeUniverse(st)
	if u != nil {
		t.Errorf("argNodeUniverse: expected nil for nil Universe, got %v", u)
	}
}

// TestFullARGNodeUniverseUnknownType verifies argNodeUniverse returns nil for
// unsupported Universe types (e.g., map[string]interface{} from mc_phase7).
func TestFullARGNodeUniverseUnknownType(t *testing.T) {
	st := &goivy.State{
		Universe: map[string]interface{}{"node": []interface{}{"n0", "n1"}},
	}
	u := argNodeUniverse(st)
	if u != nil {
		t.Errorf("argNodeUniverse: expected nil for unsupported Universe type, got %v", u)
	}
}

// TestFullARGPayloadStructure verifies FullAnalysisUIARGPayload JSON shape.
func TestFullARGPayloadStructure(t *testing.T) {
	st := &goivy.State{
		ID:      3,
		Clauses: goivy.TrueClauses(nil),
	}
	st.ActionName = "send"
	ag := &goivy.AnalysisGraph{States: []*goivy.State{st}}

	ui := NewAnalysisGraphUI()
	ui.AG = ag

	payload := FullAnalysisUIARGPayload(ui)
	data := mustCanonicalJSON(t, payload)
	keys := unmarshalKeys(t, data)

	if _, ok := keys["elements"]; !ok {
		t.Error("FullARG payload missing 'elements' key")
	}
	if _, ok := keys["analysis_graph_state"]; !ok {
		t.Error("FullARG payload missing 'analysis_graph_state' key")
	}

	var ags map[string]json.RawMessage
	if err := json.Unmarshal(keys["analysis_graph_state"], &ags); err != nil {
		t.Fatalf("analysis_graph_state not an object: %v", err)
	}

	var states []map[string]json.RawMessage
	if err := json.Unmarshal(ags["states"], &states); err != nil {
		t.Fatalf("states not an array: %v", err)
	}
	if len(states) == 0 {
		t.Fatal("states array is empty")
	}
	node := states[0]
	if _, ok := node["clauses"]; !ok {
		t.Error("FullARGNode missing 'clauses' key")
	}
	if _, ok := node["action_name"]; !ok {
		t.Error("FullARGNode missing 'action_name' key")
	}
	// universe must be absent when nil (omitempty)
	if _, ok := node["universe"]; ok {
		t.Error("FullARGNode must not have 'universe' when nil (omitempty)")
	}
}

// TestCTIARGPayloadShape verifies GetCTIARG produces the expected JSON fields.
func TestCTIARGPayloadShape(t *testing.T) {
	be := NewGoBackend(nil)
	sessionData, err := be.NewSession(&goivy.Config{})
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	var sess struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(sessionData, &sess); err != nil {
		t.Fatalf("unmarshal session: %v", err)
	}

	data, err := be.GetCTIARG(sess.SessionID)
	if err != nil {
		t.Fatalf("GetCTIARG: %v", err)
	}
	keys := unmarshalKeys(t, data)

	for _, k := range []string{"have_cti", "current_conjecture", "elements"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("CTI payload missing key %q", k)
		}
	}
	var haveCTI bool
	if err := json.Unmarshal(keys["have_cti"], &haveCTI); err != nil {
		t.Errorf("have_cti is not bool: %v", err)
	}
	var conjecture string
	if err := json.Unmarshal(keys["current_conjecture"], &conjecture); err != nil {
		t.Errorf("current_conjecture is not string: %v", err)
	}
}

// TestConceptDomainNeverNull verifies the GetConcept response map always has a
// non-null concept_domain value.
func TestConceptDomainNeverNull(t *testing.T) {
	// Simulate the pattern used in GetConcept: build conceptDomain defensively.
	conceptDomain := NewConceptDomain()
	response := map[string]interface{}{
		"concept_domain": conceptDomain,
	}
	data := mustCanonicalJSON(t, response)
	keys := unmarshalKeys(t, data)

	raw, ok := keys["concept_domain"]
	if !ok {
		t.Fatal("concept_domain key missing from response")
	}
	// Must not be JSON null.
	if string(raw) == "null" {
		t.Error("concept_domain must never be JSON null")
	}
	// Must be an object.
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Errorf("concept_domain must be a JSON object, got: %s", raw)
	}
}
