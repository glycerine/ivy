// Category B: Full Pipeline tests (with Z3).
// These tests parse, compile, and verify .ivy programs end-to-end.
package end2end

import (
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/check"
	"github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/solver"
)

// verifyInitInvariant checks that the module's initialization establishes
// all conjectures. Returns true if all conjectures hold in the initial state.
// Matches Python: ag = AnalysisGraph(initializer=lambda x:None);
//
//	check_conjs_in_state(mod, ag, ag.states[0])
func verifyInitInvariant(t *testing.T, mod *module.Module) bool {
	t.Helper()
	ag := art.NewAnalysisGraph(mod)
	ag.Initialize(func(s *art.State) {}) // no-op abstractor
	if len(ag.States) == 0 {
		t.Fatal("no initial state created")
	}
	return check.CheckConjsInStateWithAG(mod, ag, ag.States[0], 8, nil)
}

// verifyActionPreservation checks that executing an action from a state
// where all conjectures hold still satisfies all conjectures.
// Matches Python: ag = AnalysisGraph(); pre.clauses = get_conjs(mod);
//
//	post = ag.execute(action, pre); check_conjs_in_state(mod, ag, post)
func verifyActionPreservation(t *testing.T, mod *module.Module, actName string) bool {
	t.Helper()
	actionIface, ok := mod.Actions[actName]
	if !ok {
		t.Fatalf("action %q not found", actName)
	}
	action, ok := actionIface.(actions.Action)
	if !ok {
		t.Fatalf("action %q is not actions.Action, got %T", actName, actionIface)
	}

	ag := art.NewAnalysisGraph(mod)

	// Build pre-state from conjectures (matching Python get_conjs)
	conjs := check.GetConjs(mod)
	pre := art.NewState(mod, conjs)
	ag.Add(pre, nil)

	// Execute action
	post := ag.Execute(action, pre, nil, actName)
	if post == nil {
		t.Fatalf("action %q execution returned nil", actName)
	}

	return check.CheckConjsInStateWithAG(mod, ag, post, 8, nil)
}

// --- Category B tests ---

func TestVerify_TrivialPass(t *testing.T) {
	// A trivially true conjecture: r(X) | ~r(X)
	mod := compileIvyFile(t, "trivial_pass.ivy")

	// Init should establish the invariant
	if !verifyInitInvariant(t, mod) {
		t.Error("init should establish trivially true invariant")
	}
}

func TestVerify_RelationSort(t *testing.T) {
	// Two relations with simple conjectures
	mod := compileIvyFile(t, "relation_sort.ivy")
	if !verifyInitInvariant(t, mod) {
		t.Error("init should establish invariant")
	}
}

func TestVerify_ClientServer(t *testing.T) {
	// The full client-server mutual exclusion example.
	// Conjectures:
	//   link(X,Y) -> ~semaphore(Y)
	//   ~link(X,Y) | ~link(X,Z) | Y = Z
	mod := compileIvyFile(t, "client_server.ivy")

	// Init should establish invariant.
	if !verifyInitInvariant(t, mod) {
		t.Error("init should establish client-server invariant")
	}

	// Each exported action should preserve the invariant
	for _, actName := range []string{"connect", "disconnect"} {
		t.Run("preserve_"+actName, func(t *testing.T) {
			if !verifyActionPreservation(t, mod, actName) {
				t.Errorf("action %q should preserve invariant", actName)
			}
		})
	}
}

func TestVerify_PropertyFromAxioms(t *testing.T) {
	// A property provable purely from axioms, no actions needed.
	mod := compileIvySource(t, `#lang ivy1.7

type t

relation r(X:t, Y:t)

axiom r(X,Y) -> r(Y,X)

conjecture r(X,Y) -> r(Y,X)
`)
	// The conjecture is literally the axiom, so it should hold trivially.
	if !verifyInitInvariant(t, mod) {
		t.Error("conjecture identical to axiom should hold")
	}
}

func TestVerify_EnumExhaustive(t *testing.T) {
	// Verify that an enum type exhausts its cases.
	mod := compileIvyFile(t, "enum_types.ivy")
	if !verifyInitInvariant(t, mod) {
		t.Error("init should establish enum exhaustiveness invariant")
	}
}

// TestVerify_Z3SolverBasic does a bare-bones solver check to make sure
// Z3 is working in the test environment.
func TestVerify_Z3SolverBasic(t *testing.T) {
	slv := solver.New()
	if slv == nil {
		t.Fatal("solver.New() returned nil")
	}
	// Check that True is satisfiable
	sat, err := slv.IsSat(logic.True)
	if err != nil {
		t.Fatalf("IsSat error: %v", err)
	}
	if !sat {
		t.Error("True should be satisfiable")
	}
	// Check that False is unsatisfiable
	sat2, err2 := slv.IsSat(logic.False)
	if err2 != nil {
		t.Fatalf("IsSat(False) error: %v", err2)
	}
	if sat2 {
		t.Error("False should be unsatisfiable")
	}
}
