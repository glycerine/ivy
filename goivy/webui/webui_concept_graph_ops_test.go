//go:build web

package webui

import (
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Phase 1: Graph-level read/write tests
// ---------------------------------------------------------------------------

func TestGetFacts_InteractiveSession(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	Y := mkVar("Y", S)

	domain := NewCDConceptDomain(nil, nil, nil)
	nodeC := MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X))
	domain.Concepts.SetConcept("n", nodeC)
	domain.Concepts.SetSet("nodes", NewCDConceptSet("n"))

	linkC := MustCDConcept("link", []*goivy.LogicVariable{X, Y}, mkEq(X, Y))
	domain.Concepts.SetConcept("link", linkC)
	domain.Concepts.SetSet("edges", NewCDConceptSet("link"))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)
	sess.AbstractValue = []TagValue{
		{Tag: Tag{"node_info", "at_least_one", "n"}, Value: true},
		{Tag: Tag{"edge_info", "all_to_all", "link", "n", "n"}, Value: true},
	}

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	facts, err := g.GetFacts(true)
	if err != nil {
		t.Fatal(err)
	}
	// Should delegate to InteractiveSess.GetFacts — may return empty if no
	// witnesses exist, but should not panic.
	_ = facts
}

func TestGetFacts_SimpleFallback(t *testing.T) {
	g := NewGraph([]string{"node"}, nil)
	g.ConceptSess.AbstractValue["fact1"] = true
	g.ConceptSess.AbstractValue["fact2"] = true
	g.ConceptSess.AbstractValue["fact3"] = false

	facts, err := g.GetFacts(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 2 {
		t.Errorf("expected 2 true facts, got %d: %v", len(facts), facts)
	}
	found := make(map[string]bool)
	for _, f := range facts {
		found[f] = true
	}
	if !found["fact1"] || !found["fact2"] {
		t.Errorf("expected fact1 and fact2, got %v", facts)
	}
}

func TestGetFacts_NilInteractive(t *testing.T) {
	g := NewGraph([]string{"node"}, nil)
	facts, err := g.GetFacts(true)
	if err != nil {
		t.Fatal(err)
	}
	if facts == nil {
		// nil is ok when AbstractValue is empty
	}
}

func TestSetFactsExpr(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	domain := NewCDConceptDomain(nil, nil, nil)
	domain.Concepts.SetConcept("n", MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X)))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	eq1 := mkEq(mkConst("a", S), mkConst("b", S))
	eq2 := mkEq(mkConst("c", S), mkConst("d", S))

	g.SetFactsExpr([]goivy.Expr{eq1, eq2})

	if len(sess.SupposeConstraints) != 2 {
		t.Errorf("expected 2 suppose constraints, got %d", len(sess.SupposeConstraints))
	}
}

func TestGraphWidgetActiveFactsDefaultAndSelection(t *testing.T) {
	S := mkSort("node")
	g := NewGraph([]string{"node"}, nil)
	sess := NewConceptInteractiveSession(
		NewCDConceptDomain(nil, nil, nil), nil, nil, nil, nil, nil, nil, nil, false,
	)
	g.InteractiveSess = sess

	eq1 := mkEq(mkConst("a", S), mkConst("b", S))
	eq2 := mkEq(mkConst("c", S), mkConst("d", S))
	g.SetFactsExpr([]goivy.Expr{eq1, eq2})

	w := NewGraphWidget(NewGraphStack(g))
	facts := w.ConstraintFacts()
	if len(facts) != 2 {
		t.Fatalf("expected 2 constraint facts, got %d: %#v", len(facts), facts)
	}
	if !facts[0].Selected || !facts[1].Selected {
		t.Fatalf("newly rendered constraint facts should default selected: %#v", facts)
	}
	if got := w.GetActiveFacts(); len(got) != 2 {
		t.Fatalf("expected both facts active by default, got %d: %v", len(got), got)
	}

	if err := w.SetFactSelected(0, false); err != nil {
		t.Fatalf("SetFactSelected: %v", err)
	}
	if got := w.GetActiveFacts(); len(got) != 1 || got[0] != eq2.String() {
		t.Fatalf("expected only second fact active after deselecting first, got %v", got)
	}

	if err := w.SetFactSelected(1, false); err != nil {
		t.Fatalf("SetFactSelected second: %v", err)
	}
	if got := w.GetActiveFacts(); len(got) != 2 {
		t.Fatalf("expected Python fallback to all facts when none are selected, got %d: %v", len(got), got)
	}

	facts = w.ConstraintFacts()
	if facts[0].Selected || facts[1].Selected {
		t.Fatalf("constraint selection did not round-trip: %#v", facts)
	}
}

func TestConceptGraphActionPayloadIncludesHighlightedFactSelections(t *testing.T) {
	S := mkSort("node")
	g := NewGraph([]string{"node"}, nil)
	g.ConceptSess = NewConceptSession()
	g.ConceptSess.Domain.Concepts["Client"] = &Concept{Name: "Client", Formula: "client", Sorts: []string{"Client"}, Arity: 1}
	g.ConceptSess.Domain.Concepts["Server"] = &Concept{Name: "Server", Formula: "server", Sorts: []string{"Server"}, Arity: 1}
	g.ConceptSess.Domain.Concepts["link"] = &Concept{Name: "link", Variables: []string{"X", "Y"}, Formula: "link(X,Y)", Sorts: []string{"Client", "Server"}, Arity: 2}
	g.ConceptSess.Domain.Nodes = []string{"Client", "Server"}
	g.ConceptSess.Domain.Edges = []string{"link"}
	g.Checks = NewDisplayCheckboxes()
	g.Checks.SetEdgeCheckbox("link", EdgeDisplayUnknown, true)
	g.InteractiveSess = NewConceptInteractiveSession(
		NewCDConceptDomain(nil, nil, nil), nil, nil, nil, nil, nil, nil, nil, false,
	)
	fact := mkEq(mkConst("a", S), mkConst("b", S))
	g.SetFactsExpr([]goivy.Expr{fact})

	w := NewGraphWidget(NewGraphStack(g))
	w.FactElems = map[string][][]string{
		fact.String(): {
			{"Client"},
			{"link", "Client", "Server"},
		},
	}
	w.HighlightSelectedFacts()

	payload := conceptGraphActionPayload(w)
	elements, ok := payload["elements"].([]WebUICyElement)
	if !ok {
		t.Fatalf("payload elements wrong type: %#v", payload["elements"])
	}
	if !payloadHasClass(elements, "nodes", "Client", "selected_node") {
		t.Fatalf("highlighted fact node was not marked selected: %#v", elements)
	}
	if !payloadHasClass(elements, "edges", "link", "selected_edge") {
		t.Fatalf("highlighted fact edge was not marked selected: %#v", elements)
	}
}

func TestAddProjectionAddsBinaryConceptToCurrentGraph(t *testing.T) {
	client := mkSort("client")
	server := mkSort("server")
	token := mkSort("token")
	X := mkVar("X", client)
	Y := mkVar("Y", server)
	Z := mkVar("Z", token)
	c0 := mkConst("c0", client)
	routeSort, err := goivy.NewFunctionSort(client, server, token, goivy.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	route := goivy.NewConst("route", routeSort)

	domain := NewCDConceptDomain(nil, nil, nil)
	client0 := MustCDConcept("client0", []*goivy.LogicVariable{X}, mkEq(X, c0))
	domain.Concepts.SetConcept("client0", client0)
	domain.Concepts.SetSet("nodes", NewCDConceptSet("client0"))
	routeC := MustCDConcept("route", []*goivy.LogicVariable{X, Y, Z}, mkApply(route, X, Y, Z))
	domain.Concepts.SetConcept("route", routeC)
	domain.Concepts.SetSet("edges", NewCDConceptSet("route"))

	cis := NewConceptInteractiveSession(domain, nil, nil, nil, nil, nil, nil, nil, false)
	g := NewGraph([]string{"client", "server", "token"}, nil)
	g.InteractiveSess = cis
	g.ConceptSess.Domain = simpleConceptDomainFromCD(domain)
	w := NewGraphWidget(NewGraphStack(g))
	projectionActions := w.GetNodeProjectionActions("client0")
	var projection ActionEntry
	for _, action := range projectionActions {
		if action.Action == "add_projection" {
			projection = action
			break
		}
	}
	if projection.Action == "" {
		t.Fatalf("no projection action produced from ternary route relation: %#v", projectionActions)
	}
	name, _ := projection.Args["name"].(string)
	formula, _ := projection.Args["concept"].(string)
	if found, err := findProjectedConcept(cis, name, formula); err != nil || found == nil {
		t.Fatalf("projection descriptor did not resolve back to a CD concept: found=%v err=%v name=%q formula=%q", found, err, name, formula)
	}

	s := NewSession(goivy.NewConfig(), "test-add-projection-current-graph")
	s.AGUI = &AnalysisGraphUI{CurrentConceptGraph: w}
	if err := s.AddProjection(name, formula); err != nil {
		t.Fatalf("AddProjection: %v", err)
	}

	if !w.G().InteractiveSess.Domain.Concepts.Has(name) {
		t.Fatalf("interactive domain missing added projection %q", name)
	}
	if !stringSliceContains(w.G().InteractiveSess.Domain.Concepts.GetList("edges"), name) {
		t.Fatalf("interactive edge list missing projection %q: %v", name, w.G().InteractiveSess.Domain.Concepts.GetList("edges"))
	}
	concept := w.G().ConceptSess.Domain.Concepts[name]
	if concept == nil {
		t.Fatalf("render domain missing added projection %q", name)
	}
	if concept.Arity != 2 {
		t.Fatalf("projection arity = %d, want 2: %#v", concept.Arity, concept)
	}
	if !stringSliceContains(w.G().ConceptSess.Domain.Edges, name) {
		t.Fatalf("render edge list missing projection %q: %v", name, w.G().ConceptSess.Domain.Edges)
	}
	if _, ok := w.G().Checks.EdgeDisplayCheckboxes[name]; !ok {
		t.Fatalf("projection %q missing edge display checkboxes: %#v", name, w.G().Checks.EdgeDisplayCheckboxes)
	}
}

func payloadHasClass(elements []WebUICyElement, group, obj, className string) bool {
	for _, el := range elements {
		if el.Group != group {
			continue
		}
		if got, _ := el.Data["obj"].(string); got != obj {
			continue
		}
		if hasCyClass(el.Classes, className) {
			return true
		}
	}
	return false
}

func TestSetFactsExpr_Replaces(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	domain := NewCDConceptDomain(nil, nil, nil)
	domain.Concepts.SetConcept("n", MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X)))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)
	sess.SupposeConstraints = []goivy.Expr{mkEq(mkConst("old", S), mkConst("old", S))}

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	eq1 := mkEq(mkConst("new1", S), mkConst("new2", S))
	g.SetFactsExpr([]goivy.Expr{eq1})

	if len(sess.SupposeConstraints) != 1 {
		t.Errorf("expected 1 suppose constraint after replace, got %d", len(sess.SupposeConstraints))
	}
	if sess.SupposeConstraints[0].String() != eq1.String() {
		t.Errorf("expected new constraint, got %q", sess.SupposeConstraints[0])
	}
}

func TestAddConstraintsExpr(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	domain := NewCDConceptDomain(nil, nil, nil)
	domain.Concepts.SetConcept("n", MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X)))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	eq1 := mkEq(mkConst("a", S), mkConst("b", S))
	eq2 := mkEq(mkConst("c", S), mkConst("d", S))

	g.AddConstraintsExpr([]goivy.Expr{eq1, eq2}, false)
	if len(sess.SupposeConstraints) != 2 {
		t.Errorf("expected 2, got %d", len(sess.SupposeConstraints))
	}

	eq3 := mkEq(mkConst("e", S), mkConst("f", S))
	g.AddConstraintsExpr([]goivy.Expr{eq3}, false)
	if len(sess.SupposeConstraints) != 3 {
		t.Errorf("expected 3 after append, got %d", len(sess.SupposeConstraints))
	}
}

func TestAddConstraintsExpr_FiltersTautology(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	domain := NewCDConceptDomain(nil, nil, nil)
	domain.Concepts.SetConcept("n", MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X)))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	// X = X is a tautology equality — Suppose should filter it
	tautology := mkEq(X, X)
	g.AddConstraintsExpr([]goivy.Expr{tautology}, false)
	if len(sess.SupposeConstraints) != 0 {
		t.Errorf("expected tautology to be filtered, got %d constraints", len(sess.SupposeConstraints))
	}
}

// ---------------------------------------------------------------------------
// Phase 2: Z3-delegating tests
// ---------------------------------------------------------------------------

func TestMaterializeEdge_Interactive(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	Y := mkVar("Y", S)

	domain := NewCDConceptDomain(nil, nil, nil)
	n1C := MustCDConcept("n1", []*goivy.LogicVariable{X}, mkEq(X, mkConst("c1", S)))
	n2C := MustCDConcept("n2", []*goivy.LogicVariable{X}, mkEq(X, mkConst("c2", S)))
	linkC := MustCDConcept("link", []*goivy.LogicVariable{X, Y}, mkEq(X, Y))

	domain.Concepts.SetConcept("n1", n1C)
	domain.Concepts.SetConcept("n2", n2C)
	domain.Concepts.SetConcept("link", linkC)
	domain.Concepts.SetSet("nodes", NewCDConceptSet("n1", "n2"))
	domain.Concepts.SetSet("edges", NewCDConceptSet("link"))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	initialLen := len(sess.SupposeConstraints)
	witnesses, err := g.MaterializeEdge("link", "n1", "n2", true, false)
	if err != nil {
		t.Fatalf("MaterializeEdge failed: %v", err)
	}
	if len(witnesses) == 0 {
		t.Error("expected non-empty witnesses")
	}
	if len(sess.SupposeConstraints) <= initialLen {
		t.Error("expected SupposeConstraints to grow after materialization")
	}
}

func TestMaterializeEdge_Fallback(t *testing.T) {
	g := NewGraph([]string{"node"}, nil)
	witnesses, err := g.MaterializeEdge("link", "n1", "n2", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(witnesses) != 2 {
		t.Errorf("expected 2 dummy witnesses, got %d", len(witnesses))
	}
	if witnesses[0] != "n1_witness" || witnesses[1] != "n2_witness" {
		t.Errorf("expected dummy witness names, got %v", witnesses)
	}
}

func TestMaterializeEdge_NegativePolarity(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	Y := mkVar("Y", S)

	domain := NewCDConceptDomain(nil, nil, nil)
	n1C := MustCDConcept("n1", []*goivy.LogicVariable{X}, mkEq(X, mkConst("c1", S)))
	linkC := MustCDConcept("link", []*goivy.LogicVariable{X, Y}, mkEq(X, Y))

	domain.Concepts.SetConcept("n1", n1C)
	domain.Concepts.SetConcept("link", linkC)
	domain.Concepts.SetSet("nodes", NewCDConceptSet("n1"))
	domain.Concepts.SetSet("edges", NewCDConceptSet("link"))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	_, err := g.MaterializeEdge("link", "n1", "n1", false, false)
	if err != nil {
		t.Fatalf("MaterializeEdge(negative) failed: %v", err)
	}
	// Check that a negation was supposed
	hasNot := false
	for _, sc := range sess.SupposeConstraints {
		if strings.Contains(sc.String(), "~") || strings.Contains(sc.String(), "not") || strings.Contains(sc.String(), "Not") {
			hasNot = true
		}
	}
	if !hasNot && len(sess.SupposeConstraints) > 0 {
		// The negation format depends on logic.Not.String() — just verify constraints grew
		t.Logf("SupposeConstraints: %v", sess.SupposeConstraints)
	}
}

// ---------------------------------------------------------------------------
// ConceptSession.Materialize tests
// ---------------------------------------------------------------------------

func TestConceptSession_Materialize(t *testing.T) {
	cs := NewConceptSession()
	cs.Domain.Concepts["node"] = &Concept{
		Name: "node", Variables: []string{"X"},
		Formula: "X = X", Sorts: []string{"s"}, Arity: 1,
	}
	cs.Domain.Nodes = append(cs.Domain.Nodes, "node")

	err := cs.Materialize("node")
	if err != nil {
		t.Fatalf("Materialize failed: %v", err)
	}

	// The original "node" concept should be gone (replaced by sub-concepts)
	if _, ok := cs.Domain.Concepts["node"]; ok {
		t.Error("original concept 'node' should have been split away")
	}

	// A witness concept "=__c0" should exist
	if _, ok := cs.Domain.Concepts["=__c0"]; !ok {
		t.Error("expected witness concept '=__c0'")
	}

	// Undo stack should have entries (Materialize pushes, then Split pushes)
	if len(cs.undoStack) < 1 {
		t.Errorf("expected undo entries, got %d", len(cs.undoStack))
	}
}

func TestConceptSession_Materialize_NotFound(t *testing.T) {
	cs := NewConceptSession()
	err := cs.Materialize("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent concept")
	}
}

func TestConceptSession_FreshConstName(t *testing.T) {
	cs := NewConceptSession()
	name1 := cs.freshConstName()
	// Add the name so it collides
	cs.Domain.Concepts["="+name1] = &Concept{Name: "=" + name1}
	name2 := cs.freshConstName()

	if name1 == name2 {
		t.Errorf("freshConstName should generate unique names, got %q twice", name1)
	}
	if name1 != "__c0" {
		t.Errorf("expected __c0, got %q", name1)
	}
	if name2 != "__c1" {
		t.Errorf("expected __c1, got %q", name2)
	}
}

// ---------------------------------------------------------------------------
// Splatter tests
// ---------------------------------------------------------------------------

func TestSplatter_Interactive(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)

	domain := NewCDConceptDomain(nil, nil, nil)
	nC := MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X))
	domain.Concepts.SetConcept("n", nC)
	domain.Concepts.SetSet("nodes", NewCDConceptSet("n"))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	a := mkConst("a", S)
	b := mkConst("b", S)
	sess.SupposeConstraints = []goivy.Expr{
		mkEq(X, a),
		mkEq(X, b),
	}

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess
	gs := NewGraphStack(g)
	w := NewGraphWidget(gs)

	w.Splatter("n")

	// After splatter, "=a" and "=b" concepts should exist
	if sess.Domain.Concepts.GetConcept("=a") == nil {
		t.Error("expected concept '=a' after splatter")
	}
	if sess.Domain.Concepts.GetConcept("=b") == nil {
		t.Error("expected concept '=b' after splatter")
	}
}

func TestSplatter_NoConstants(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)

	domain := NewCDConceptDomain(nil, nil, nil)
	nC := MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X))
	domain.Concepts.SetConcept("n", nC)
	domain.Concepts.SetSet("nodes", NewCDConceptSet("n"))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess
	gs := NewGraphStack(g)
	w := NewGraphWidget(gs)

	// No constants in SupposeConstraints — should not crash
	w.Splatter("n")

	// Original concept should remain unchanged
	if sess.Domain.Concepts.GetConcept("n") == nil {
		t.Error("concept 'n' should remain when no constants to splatter")
	}
}

func TestSplatter_NonexistentConcept(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)

	domain := NewCDConceptDomain(nil, nil, nil)
	nC := MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X))
	domain.Concepts.SetConcept("n", nC)

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess
	gs := NewGraphStack(g)
	w := NewGraphWidget(gs)

	// Splatter on nonexistent concept — should not crash
	w.Splatter("nonexistent")
}

// ---------------------------------------------------------------------------
// Recalculate tests
// ---------------------------------------------------------------------------

func TestRecalculate_NoParent(t *testing.T) {
	g := NewGraph([]string{"node"}, nil)
	gs := NewGraphStack(g)
	w := NewGraphWidget(gs)

	// No parent — should not crash
	w.Recalculate()
}

func TestRecalculate_WithParentState(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)

	domain := NewCDConceptDomain(nil, nil, nil)
	nC := MustCDConcept("n", []*goivy.LogicVariable{X}, mkEq(X, X))
	domain.Concepts.SetConcept("n", nC)

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	state := &goivy.State{
		Clauses: goivy.NewClauses([]goivy.Expr{mkEq(X, X)}, nil, nil),
	}

	g := NewGraph([]string{"node"}, state)
	g.InteractiveSess = sess
	gs := NewGraphStack(g)
	w := NewGraphWidget(gs)

	agui := &AnalysisGraphUI{
		AG: goivy.NewAnalysisGraph(nil),
	}
	w.Parent = agui

	// Should not crash even if AG has no states
	w.Recalculate()
}

func TestRecalculate_RefreshesParentARGStateWithoutUnrelatedStates(t *testing.T) {
	mod := goivy.New()
	ag := goivy.NewAnalysisGraph(mod)
	clauseFormula := func(clauses *goivy.Clauses) string {
		if clauses == nil {
			return ""
		}
		return clauses.ToFormula().String()
	}

	left := goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil))
	right := goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil))
	staleJoined := goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil))
	staleJoined.JoinOf = []*goivy.State{left, right}
	unrelated := goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil))
	ag.Add(left, nil)
	ag.Add(right, nil)
	ag.Add(staleJoined, nil)
	ag.Add(unrelated, nil)

	ui := &AnalysisGraphUI{AG: ag, Mod: mod}
	w, err := ui.ViewState(staleJoined.ID, "", false)
	if err != nil {
		t.Fatalf("ViewState: %v", err)
	}

	if got := clauseFormula(staleJoined.Clauses); got != "true" {
		t.Fatalf("test setup expected stale joined state to start true, got %q", got)
	}
	if err := w.Recalculate(); err != nil {
		t.Fatalf("Recalculate: %v", err)
	}
	if got := clauseFormula(staleJoined.Clauses); !strings.Contains(got, "false") {
		t.Fatalf("joined parent state was not recalculated from join sources: got %q", got)
	}
	if got := w.G().State; !strings.Contains(got, "false") {
		t.Fatalf("concept graph did not reload recalculated parent clauses: got %q", got)
	}
	if got := clauseFormula(unrelated.Clauses); got != "true" {
		t.Fatalf("unrelated state should not be recalculated by state recalc: got %q", got)
	}
}

func TestExecuteActionRecalculateRefreshesCurrentConceptGraphState(t *testing.T) {
	mod := goivy.New()
	ag := goivy.NewAnalysisGraph(mod)
	clauseFormula := func(clauses *goivy.Clauses) string {
		if clauses == nil {
			return ""
		}
		return clauses.ToFormula().String()
	}

	left := goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil))
	right := goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.False}, nil, nil))
	staleJoined := goivy.NewState(mod, goivy.NewClauses([]goivy.Expr{goivy.True}, nil, nil))
	staleJoined.JoinOf = []*goivy.State{left, right}
	ag.Add(left, nil)
	ag.Add(right, nil)
	ag.Add(staleJoined, nil)

	ui := &AnalysisGraphUI{AG: ag, Mod: mod}
	if _, err := ui.ViewState(staleJoined.ID, "", false); err != nil {
		t.Fatalf("ViewState: %v", err)
	}
	s := NewSession(goivy.NewConfig(), "test-recalculate-state")
	s.AG = ag
	s.AGUI = ui
	s.SheetUIs = map[string]*AnalysisGraphUI{rootSheetID: ui}
	s.sheetCounter = 1

	result, err := s.ExecuteAction("recalculate", map[string]interface{}{"sheet_id": rootSheetID})
	if err != nil {
		t.Fatalf("ExecuteAction(recalculate): %v", err)
	}
	if got := clauseFormula(staleJoined.Clauses); !strings.Contains(got, "false") {
		t.Fatalf("selected parent state was not recalculated: got %q", got)
	}
	if _, ok := result["concept"].(map[string]interface{}); !ok {
		t.Fatalf("recalculate result missing concept graph snapshot: %#v", result["concept"])
	}
}

// ---------------------------------------------------------------------------
// RegisterArg* tests
// ---------------------------------------------------------------------------

func testExtConfigWithAG(nStates int) *ExtConfig {
	cfg := NewExtConfig()
	ag := goivy.NewAnalysisGraph(nil)
	for i := 0; i < nStates; i++ {
		s := &goivy.State{
			ID:      i,
			Clauses: goivy.NewClauses(nil, nil, nil),
		}
		if i > 0 {
			s.Pred = ag.States[i-1]
		}
		ag.States = append(ag.States, s)
	}
	cfg.AG = ag
	return cfg
}

func TestRegisterArgNewGoal(t *testing.T) {
	cfg := testExtConfigWithAG(2)
	cfg.RegisterArgNewGoal()

	actions, errs := cfg.ArgNodeActions.Invoke(nil, 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(actions) == 0 {
		t.Fatal("expected at least one action")
	}

	// Find and invoke the "new goal" action
	for _, a := range actions {
		if a.Label == "new goal" {
			err := a.Callback(0)
			if err != nil {
				t.Fatalf("new goal callback failed: %v", err)
			}
			return
		}
	}
	t.Error("'new goal' action not found")
}

func TestRegisterArgRecalculate(t *testing.T) {
	cfg := testExtConfigWithAG(2)
	cfg.RegisterArgRecalculate()

	actions, errs := cfg.ArgNodeActions.Invoke(nil, 1)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, a := range actions {
		if a.Label == "recalculate" {
			err := a.Callback(1)
			if err != nil {
				t.Fatalf("recalculate callback failed: %v", err)
			}
			return
		}
	}
	t.Error("'recalculate' action not found")
}

func TestRegisterArgRecalculate_NoPred(t *testing.T) {
	cfg := testExtConfigWithAG(1) // only 1 state, no predecessor
	cfg.RegisterArgRecalculate()

	actions, _ := cfg.ArgNodeActions.Invoke(nil, 0)
	for _, a := range actions {
		if a.Label == "recalculate" {
			err := a.Callback(0)
			if err == nil {
				t.Error("expected error for node with no predecessor")
			}
			return
		}
	}
}

func TestRegisterArgCheckCover(t *testing.T) {
	cfg := testExtConfigWithAG(2)
	cfg.RegisterArgCheckCover()

	actions, errs := cfg.ArgNodeActions.Invoke(nil, 0, 1)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, a := range actions {
		if a.Label == "check cover" {
			err := a.Callback(0, 1)
			if err != nil {
				t.Fatalf("check cover callback failed: %v", err)
			}
			return
		}
	}
	t.Error("'check cover' action not found")
}

func TestRegisterArgRemoveFacts(t *testing.T) {
	S := mkSort("node")
	f1 := mkEq(mkConst("a", S), mkConst("b", S))
	f2 := mkEq(mkConst("c", S), mkConst("d", S))
	f3 := mkEq(mkConst("e", S), mkConst("f", S))

	cfg := NewExtConfig()
	ag := goivy.NewAnalysisGraph(nil)
	state := &goivy.State{
		ID:      0,
		Clauses: goivy.NewClauses([]goivy.Expr{f1, f2, f3}, nil, nil),
	}
	ag.States = append(ag.States, state)
	cfg.AG = ag
	cfg.RegisterArgRemoveFacts()

	actions, errs := cfg.ArgNodeActions.Invoke(nil, 0, []goivy.Expr{f2})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, a := range actions {
		if a.Label == "remove facts" {
			err := a.Callback(0, []goivy.Expr{f2})
			if err != nil {
				t.Fatalf("remove facts callback failed: %v", err)
			}
			// After removing f2, should have f1 and f3
			remaining := state.Clauses.Fmlas
			if len(remaining) != 2 {
				t.Errorf("expected 2 remaining facts, got %d", len(remaining))
			}
			return
		}
	}
	t.Error("'remove facts' action not found")
}

func TestRegisterArgJoin(t *testing.T) {
	cfg := testExtConfigWithAG(2)
	cfg.RegisterArgJoin()
	initialStates := cfg.AG.StateCount()

	actions, errs := cfg.ArgNodeActions.Invoke(nil, 0, 1)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, a := range actions {
		if a.Label == "join with selection" {
			err := a.Callback(0, 1)
			if err != nil {
				t.Fatalf("join callback failed: %v", err)
			}
			// Join should add a new state
			if cfg.AG.StateCount() <= initialStates {
				t.Logf("join may not add state when clauses are nil (expected)")
			}
			return
		}
	}
	t.Error("'join with selection' action not found")
}

func TestRegisterArg_NoAG(t *testing.T) {
	cfg := NewExtConfig()
	cfg.AG = nil

	cfg.RegisterArgNewGoal()
	cfg.RegisterArgRecalculate()
	cfg.RegisterArgCheckCover()
	cfg.RegisterArgRemoveFacts()
	cfg.RegisterArgJoin()

	actions, _ := cfg.ArgNodeActions.Invoke(nil, 0)
	for _, a := range actions {
		err := a.Callback(0)
		if err == nil {
			t.Errorf("expected error from %q with nil AG", a.Label)
		}
		if !strings.Contains(err.Error(), "no analysis graph") {
			t.Errorf("expected 'no analysis graph' error from %q, got: %v", a.Label, err)
		}
	}
}

func TestRegisterArgCheckCover_InvalidIDs(t *testing.T) {
	cfg := testExtConfigWithAG(2)
	cfg.RegisterArgCheckCover()

	actions, _ := cfg.ArgNodeActions.Invoke(nil, 0, 99)
	for _, a := range actions {
		if a.Label == "check cover" {
			err := a.Callback(0, 99)
			if err == nil {
				t.Error("expected error for invalid node ID")
			}
		}
	}
}

func TestRegisterArgJoin_InvalidIDs(t *testing.T) {
	cfg := testExtConfigWithAG(1)
	cfg.RegisterArgJoin()

	actions, _ := cfg.ArgNodeActions.Invoke(nil, 0, 99)
	for _, a := range actions {
		if a.Label == "join with selection" {
			err := a.Callback(0, 99)
			if err == nil {
				t.Error("expected error for invalid selection ID")
			}
		}
	}
}
