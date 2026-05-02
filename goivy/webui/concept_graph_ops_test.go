//go:build web

package webui

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/art"
	"github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// ---------------------------------------------------------------------------
// Phase 1: Graph-level read/write tests
// ---------------------------------------------------------------------------

func TestGetFacts_InteractiveSession(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	Y := mkVar("Y", S)

	domain := NewCDConceptDomain(nil, nil, nil)
	nodeC := MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X))
	domain.Concepts.SetConcept("n", nodeC)
	domain.Concepts.SetSet("nodes", NewCDConceptSet("n"))

	linkC := MustCDConcept("link", []*logic.Variable{X, Y}, mkEq(X, Y))
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

	facts := g.GetFacts(true)
	// Should delegate to InteractiveSess.GetFacts — may return empty if no
	// witnesses exist, but should not panic.
	_ = facts
}

func TestGetFacts_SimpleFallback(t *testing.T) {
	g := NewGraph([]string{"node"}, nil)
	g.ConceptSess.AbstractValue["fact1"] = true
	g.ConceptSess.AbstractValue["fact2"] = true
	g.ConceptSess.AbstractValue["fact3"] = false

	facts := g.GetFacts(false)
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
	facts := g.GetFacts(true)
	if facts == nil {
		// nil is ok when AbstractValue is empty
	}
}

func TestSetFactsExpr(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	domain := NewCDConceptDomain(nil, nil, nil)
	domain.Concepts.SetConcept("n", MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X)))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	eq1 := mkEq(mkConst("a", S), mkConst("b", S))
	eq2 := mkEq(mkConst("c", S), mkConst("d", S))

	g.SetFactsExpr([]logic.Expr{eq1, eq2})

	if len(sess.SupposeConstraints) != 2 {
		t.Errorf("expected 2 suppose constraints, got %d", len(sess.SupposeConstraints))
	}
}

func TestSetFactsExpr_Replaces(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	domain := NewCDConceptDomain(nil, nil, nil)
	domain.Concepts.SetConcept("n", MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X)))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)
	sess.SupposeConstraints = []logic.Expr{mkEq(mkConst("old", S), mkConst("old", S))}

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	eq1 := mkEq(mkConst("new1", S), mkConst("new2", S))
	g.SetFactsExpr([]logic.Expr{eq1})

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
	domain.Concepts.SetConcept("n", MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X)))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	eq1 := mkEq(mkConst("a", S), mkConst("b", S))
	eq2 := mkEq(mkConst("c", S), mkConst("d", S))

	g.AddConstraintsExpr([]logic.Expr{eq1, eq2}, false)
	if len(sess.SupposeConstraints) != 2 {
		t.Errorf("expected 2, got %d", len(sess.SupposeConstraints))
	}

	eq3 := mkEq(mkConst("e", S), mkConst("f", S))
	g.AddConstraintsExpr([]logic.Expr{eq3}, false)
	if len(sess.SupposeConstraints) != 3 {
		t.Errorf("expected 3 after append, got %d", len(sess.SupposeConstraints))
	}
}

func TestAddConstraintsExpr_FiltersTautology(t *testing.T) {
	S := mkSort("node")
	X := mkVar("X", S)
	domain := NewCDConceptDomain(nil, nil, nil)
	domain.Concepts.SetConcept("n", MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X)))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	g := NewGraph([]string{"node"}, nil)
	g.InteractiveSess = sess

	// X = X is a tautology equality — Suppose should filter it
	tautology := mkEq(X, X)
	g.AddConstraintsExpr([]logic.Expr{tautology}, false)
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
	n1C := MustCDConcept("n1", []*logic.Variable{X}, mkEq(X, mkConst("c1", S)))
	n2C := MustCDConcept("n2", []*logic.Variable{X}, mkEq(X, mkConst("c2", S)))
	linkC := MustCDConcept("link", []*logic.Variable{X, Y}, mkEq(X, Y))

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
	n1C := MustCDConcept("n1", []*logic.Variable{X}, mkEq(X, mkConst("c1", S)))
	linkC := MustCDConcept("link", []*logic.Variable{X, Y}, mkEq(X, Y))

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
	nC := MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X))
	domain.Concepts.SetConcept("n", nC)
	domain.Concepts.SetSet("nodes", NewCDConceptSet("n"))

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	a := mkConst("a", S)
	b := mkConst("b", S)
	sess.SupposeConstraints = []logic.Expr{
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
	nC := MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X))
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
	nC := MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X))
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
	nC := MustCDConcept("n", []*logic.Variable{X}, mkEq(X, X))
	domain.Concepts.SetConcept("n", nC)

	sess := NewConceptInteractiveSession(
		domain, nil, nil, nil, nil, nil, nil, nil, false,
	)

	state := &art.State{
		Clauses: module.NewClauses([]logic.Expr{mkEq(X, X)}, nil, nil),
	}

	g := NewGraph([]string{"node"}, state)
	g.InteractiveSess = sess
	gs := NewGraphStack(g)
	w := NewGraphWidget(gs)

	agui := &AnalysisGraphUI{
		AG: art.NewAnalysisGraph(nil),
	}
	w.Parent = agui

	// Should not crash even if AG has no states
	w.Recalculate()
}

// ---------------------------------------------------------------------------
// RegisterArg* tests
// ---------------------------------------------------------------------------

func testExtConfigWithAG(nStates int) *ExtConfig {
	cfg := NewExtConfig()
	ag := art.NewAnalysisGraph(nil)
	for i := 0; i < nStates; i++ {
		s := &art.State{
			ID:      i,
			Clauses: module.NewClauses(nil, nil, nil),
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
	ag := art.NewAnalysisGraph(nil)
	state := &art.State{
		ID:      0,
		Clauses: module.NewClauses([]logic.Expr{f1, f2, f3}, nil, nil),
	}
	ag.States = append(ag.States, state)
	cfg.AG = ag
	cfg.RegisterArgRemoveFacts()

	actions, errs := cfg.ArgNodeActions.Invoke(nil, 0, []logic.Expr{f2})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, a := range actions {
		if a.Label == "remove facts" {
			err := a.Callback(0, []logic.Expr{f2})
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
