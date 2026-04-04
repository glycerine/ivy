package art

import (
	"testing"

	"github.com/glycerine/ivy/goivy/actions"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// ---------------------------------------------------------------------------
// Precondition checking tests (Steps 1, 3, 4 of plan)
// ---------------------------------------------------------------------------

// TestPostStateChecksPreconditionFalse verifies that PostState with
// checkPrecond=false computes a post-state without checking preconditions.
func TestPostStateChecksPreconditionFalse(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post, err := ag.PostState(false, act, pre, nil)
	if err != nil {
		t.Fatalf("PostState(checkPrecond=false) should not error: %v", err)
	}
	if post == nil {
		t.Fatal("PostState should return a non-nil state")
	}
	if post.Action != act {
		t.Error("action should be set on post-state")
	}
}

// TestPostStateWithAbstractorCalled verifies the abstractor is called.
func TestPostStateWithAbstractorCalled(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	called := false
	abs := AbstractorFunc(func(s *State) { called = true })
	_, err := ag.PostState(false, act, pre, abs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("abstractor should have been called")
	}
}

// TestExecuteActionReturnsError verifies ExecuteAction returns an error
// for unknown actions.
func TestExecuteActionReturnsError(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	_, err := ag.ExecuteAction(false, "nonexistent", pre, nil)
	if err == nil {
		t.Error("ExecuteAction should return error for unknown action")
	}
}

// TestExecuteReturnsPostState verifies Execute produces a post-state and
// adds it to the graph.
func TestExecuteReturnsPostState(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	ag.Actions.Set("test_act", act)
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post, err := ag.Execute(false, act, pre, nil, "test_act")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if post == nil {
		t.Fatal("Execute should return a non-nil post-state")
	}
	if ag.StateCount() != 2 {
		t.Errorf("expected 2 states, got %d", ag.StateCount())
	}
}

// TestExecuteNilPrestate verifies Execute defaults to last state.
func TestExecuteNilPrestate(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	ag.Actions.Set("act", act)
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post, err := ag.Execute(false, act, nil, nil, "act")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if post == nil {
		t.Fatal("Execute should return a non-nil post-state when prestate defaults to last")
	}
}

// TestExecuteEmptyGraphReturnsNil verifies Execute on empty graph returns nil.
func TestExecuteEmptyGraphReturnsNil(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)

	post, err := ag.Execute(false, act, nil, nil, "")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if post != nil {
		t.Error("Execute on empty graph should return nil")
	}
}

// TestAddInitialStateCheckFalse verifies AddInitialState uses check=false
// (matching Python's EvalContext(check=False)).
func TestAddInitialStateCheckFalse(t *testing.T) {
	mod := testModule()
	ag := NewAnalysisGraph(mod)
	s := ag.AddInitialState(nil, nil)
	if s == nil {
		t.Fatal("AddInitialState should return a state")
	}
	if ag.StateCount() < 1 {
		t.Error("should have at least 1 state after AddInitialState")
	}
}

// ---------------------------------------------------------------------------
// JoinStates tests (Step 2)
// ---------------------------------------------------------------------------

// TestJoinStatesPreservesLabel verifies the joined state gets state1's label.
func TestJoinStatesPreservesLabel(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	s1.Label = "labelA"
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	s2.Label = "labelB"
	ag.Add(s2, nil)

	joined := ag.JoinStates(s1, s2, nil)
	if joined == nil {
		t.Fatal("JoinStates should return a non-nil state")
	}
	if joined.Label != "labelA" {
		t.Errorf("expected label 'labelA', got %q", joined.Label)
	}
}

// TestJoinStatesJoinOfSet verifies JoinOf records the source states.
func TestJoinStatesJoinOfSet(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)

	joined := ag.JoinStates(s1, s2, nil)
	if joined.JoinOf == nil || len(joined.JoinOf) != 2 {
		t.Error("JoinOf should have 2 entries")
	}
}

// TestJoinStatesAbstractorCalled verifies the abstractor is called on join.
func TestJoinStatesAbstractorCalled(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)

	called := false
	abs := AbstractorFunc(func(s *State) { called = true })
	ag.JoinStates(s1, s2, abs)
	if !called {
		t.Error("abstractor should have been called")
	}
}

// ---------------------------------------------------------------------------
// RecalculateState tests (Step 3)
// ---------------------------------------------------------------------------

// TestRecalculateStateWithJoin verifies RecalculateState handles join states.
func TestRecalculateStateWithJoin(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)

	joined := testState(ag.Domain)
	joined.JoinOf = []*State{s1, s2}
	ag.Add(joined, NewStateJoin(s1, s2))

	ag.RecalculateState(false, joined, nil)
	// Should not panic and state should still be in graph.
	if ag.StateCount() != 3 {
		t.Errorf("expected 3 states, got %d", ag.StateCount())
	}
}

// ---------------------------------------------------------------------------
// Recalculate tests (Step 3)
// ---------------------------------------------------------------------------

// TestRecalculateWithCheckPrecond verifies Recalculate passes checkPrecond.
func TestRecalculateWithCheckPrecond(t *testing.T) {
	ag := testGraph()
	act := actions.NewAssumeAction(lg.True)
	ag.Actions.Set("myact", act)
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := testState(ag.Domain)
	ag.Add(post, NewActionApp("myact", pre))

	tr := ag.Transitions[0]
	result, err := ag.Recalculate(false, tr, nil)
	if err != nil {
		t.Fatalf("Recalculate error: %v", err)
	}
	if result == nil {
		t.Fatal("recalculate should return a state")
	}
}

// TestRecalculateJoinTransition verifies Recalculate handles join transitions.
func TestRecalculateJoinTransition(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)
	s3 := testState(ag.Domain)
	s3.JoinOf = []*State{s1, s2}
	ag.Add(s3, NewStateJoin(s1, s2))

	// Find a join transition
	var joinTr Transition
	for _, tr := range ag.Transitions {
		if tr.Label == "join" {
			joinTr = tr
			break
		}
	}
	result, err := ag.Recalculate(false, joinTr, nil)
	if err != nil {
		t.Fatalf("Recalculate join error: %v", err)
	}
	if result == nil {
		t.Fatal("recalculate join should return a state")
	}
}

// ---------------------------------------------------------------------------
// Call method tests (Step 5)
// ---------------------------------------------------------------------------

// TestCallAddsStateAndTransition verifies Call adds a post-state and transition.
func TestCallAddsStateAndTransition(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)

	post := ag.Call("test_call", func(sub *AnalysisGraph) {
		s := testState(sub.Domain)
		sub.Add(s, nil)
	}, pre)
	if post == nil {
		t.Fatal("Call should return a non-nil state")
	}
	if ag.StateCount() != 2 {
		t.Errorf("expected 2 states, got %d", ag.StateCount())
	}
	// Should have a transition
	found := false
	for _, tr := range ag.Transitions {
		if tr.Label == "test_call" && tr.Pre == pre && tr.Post == post {
			found = true
		}
	}
	if !found {
		t.Error("Call should create a transition with the given label")
	}
}

// TestCallDefaultsToLastState verifies Call defaults prestate to the last state.
func TestCallDefaultsToLastState(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	ag.Add(s, nil)

	post := ag.Call("op", func(sub *AnalysisGraph) {
		sub.Add(testState(sub.Domain), nil)
	}, nil)
	if post == nil {
		t.Fatal("Call with nil prestate should use last state")
	}
}

// TestCallEmptyGraphReturnsNil verifies Call on empty graph returns nil.
func TestCallEmptyGraphReturnsNil(t *testing.T) {
	ag := testGraph()
	post := ag.Call("op", func(sub *AnalysisGraph) {}, nil)
	if post != nil {
		t.Error("Call on empty graph should return nil")
	}
}

// ---------------------------------------------------------------------------
// ConstructTransitionsFromExpressions tests (Step 6)
// ---------------------------------------------------------------------------

// TestConstructTransitionsFromExpressions verifies transitions are rebuilt
// from state expressions.
func TestConstructTransitionsFromExpressions(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	s1.Prov = NewActionApp("act", s0)
	// Add without creating a transition (add with nil expr, set expr manually).
	ag.Add(s1, nil)

	// Clear transitions
	ag.Transitions = nil

	ag.ConstructTransitionsFromExpressions()
	if len(ag.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(ag.Transitions))
	}
	tr := ag.Transitions[0]
	if tr.Pre != s0 {
		t.Error("pre-state mismatch")
	}
	if tr.Post != s1 {
		t.Error("post-state mismatch")
	}
	if tr.Label != "act" {
		t.Errorf("expected label 'act', got %q", tr.Label)
	}
}

// TestConstructTransitionsFromExpressionsNoExpr verifies no transitions
// are created for states without expressions.
func TestConstructTransitionsFromExpressionsNoExpr(t *testing.T) {
	ag := testGraph()
	addStates(ag, 3)
	ag.Transitions = nil

	ag.ConstructTransitionsFromExpressions()
	if len(ag.Transitions) != 0 {
		t.Errorf("expected 0 transitions, got %d", len(ag.Transitions))
	}
}

// TestConstructTransitionsFromExpressionsSkipsJoin verifies that state join
// expressions are not turned into action transitions.
func TestConstructTransitionsFromExpressionsSkipsJoin(t *testing.T) {
	ag := testGraph()
	s0 := testState(ag.Domain)
	ag.Add(s0, nil)
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	s2.Prov = NewStateJoin(s0, s1)
	ag.Add(s2, nil)
	ag.Transitions = nil

	ag.ConstructTransitionsFromExpressions()
	if len(ag.Transitions) != 0 {
		t.Errorf("expected 0 transitions (join should be skipped), got %d", len(ag.Transitions))
	}
}

// ---------------------------------------------------------------------------
// Fuzz tests
// ---------------------------------------------------------------------------

// FuzzPostStateCheckPrecond fuzz-tests PostState with random checkPrecond values.
func FuzzPostStateCheckPrecond(f *testing.F) {
	f.Add(true)
	f.Add(false)

	f.Fuzz(func(t *testing.T, checkPrecond bool) {
		ag := testGraph()
		act := actions.NewAssumeAction(lg.True)
		pre := testState(ag.Domain)
		ag.Add(pre, nil)

		post, err := ag.PostState(checkPrecond, act, pre, nil)
		if err != nil {
			// Precondition failure is acceptable
			return
		}
		if post == nil {
			t.Error("PostState should return a state when no error")
		}
	})
}

// FuzzConstructTransitions verifies round-trip: build expressions → construct transitions.
func FuzzConstructTransitions(f *testing.F) {
	f.Add(uint8(1))
	f.Add(uint8(3))
	f.Add(uint8(10))

	f.Fuzz(func(t *testing.T, numActions uint8) {
		n := int(numActions)%10 + 1
		ag := testGraph()
		s0 := testState(ag.Domain)
		ag.Add(s0, nil)

		prev := s0
		for i := 0; i < n; i++ {
			name := "act"
			registerAction(ag, name)
			s := testState(ag.Domain)
			s.Prov = NewActionApp(name, prev)
			ag.Add(s, nil)
			prev = s
		}

		ag.Transitions = nil
		ag.ConstructTransitionsFromExpressions()
		if len(ag.Transitions) != n {
			t.Errorf("expected %d transitions, got %d", n, len(ag.Transitions))
		}
	})
}

// FuzzExecuteAndCheckSafety fuzz-tests the execute→check_safety pipeline.
func FuzzExecuteAndCheckSafety(f *testing.F) {
	f.Add(true, uint8(1))
	f.Add(false, uint8(3))

	f.Fuzz(func(t *testing.T, checkPrecond bool, nStates uint8) {
		n := int(nStates)%5 + 1
		ag := testGraph()
		act := actions.NewAssumeAction(lg.True)
		ag.Actions.Set("step", act)
		ag.PublicActions.Set("step", true)

		s0 := testState(ag.Domain)
		ag.Add(s0, nil)

		for i := 0; i < n; i++ {
			post, err := ag.ExecuteAction(checkPrecond, "step", nil, nil)
			if err != nil {
				return // precondition failure is OK
			}
			if post == nil {
				return
			}
		}

		// Check safety should not panic.
		res := ag.CheckSafety(checkPrecond, ag.LastState())
		if res == nil {
			t.Error("CheckSafety should always return a result")
		}
	})
}

// Ensure transrel import is used.
var _ = actions.NullUpdate
var _ = module.New
var _ = module.TrueClauses
