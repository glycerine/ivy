package art

import (
	"reflect"
	"testing"
)

// TestPortCompleteness_AllPythonMethodsExist verifies that every method
// from Python ivy_art.py has a corresponding Go method on *AnalysisGraph.
func TestPortCompleteness_AllPythonMethodsExist(t *testing.T) {
	// Every Python method name from ivy_art.py mapped to Go PascalCase.
	expectedMethods := []string{
		// Python: context (property) → Go: Context
		"Context",
		// Python: add_initial_state → Go: AddInitialState
		"AddInitialState",
		// Python: initialize → Go: Initialize
		"Initialize",
		// Python: state_actions → Go: StateActions
		"StateActions",
		// Python: do_state_action → Go: DoStateAction
		"DoStateAction",
		// Python: recalculate → Go: Recalculate
		"Recalculate",
		// Python: post_state → Go: PostState
		"PostState",
		// Python: replace_state → Go: ReplaceState
		"ReplaceState",
		// Python: recalculate_state → Go: RecalculateState
		"RecalculateState",
		// Python: execute → Go: Execute
		"Execute",
		// Python: execute_action → Go: ExecuteAction
		"ExecuteAction",
		// Python: join_states → Go: JoinStates
		"JoinStates",
		// Python: join → Go: Join
		"Join",
		// Python: cover → Go: Cover
		"Cover",
		// Python: is_covered → Go: IsCovered
		"IsCovered",
		// Python: unreachable → Go: Unreachable
		"Unreachable",
		// Python: show_core → Go: ShowCore
		"ShowCore",
		// Python: add → Go: Add
		"Add",
		// Python: call_action → Go: CallAction
		"CallAction",
		// Python: call → Go: Call
		"Call",
		// Python: dependencies → private, tested via Delete
		// Python: delete → Go: Delete
		"Delete",
		// Python: remove_marked_states → Go: RemoveMarkedStates
		"RemoveMarkedStates",
		// Python: concept_graph → Go: ConceptGraph
		"ConceptGraph",
		// Python: transition_to → Go: TransitionTo
		"TransitionTo",
		// Python: get_history → Go: GetHistory
		"GetHistory",
		// Python: bmc → Go: BMC
		"BMC",
		// Python: check_bounded_safety → Go: CheckBoundedSafety
		"CheckBoundedSafety",
		// Python: copy_path → Go: CopyPath
		"CopyPath",
		// Python: check_safety → Go: CheckSafety
		"CheckSafety",
		// Python: construct_transitions_from_expressions → Go: ConstructTransitionsFromExpressions
		"ConstructTransitionsFromExpressions",
		// Python: decompose_state → Go: DecomposeState
		"DecomposeState",
		// Python: make_concrete_trace → Go: MakeConcreteTrace
		"MakeConcreteTrace",
		// Python: decompose_edge → Go: DecomposeEdge
		"DecomposeEdge",
		// Python: uncovered_states (property) → Go: UncoveredStates
		"UncoveredStates",
		// Python: fixedpoint_candidate → Go: FixedpointCandidate
		"FixedpointCandidate",
		// Python: state_extensions → Go: StateExtensions
		"StateExtensions",
		// Python: as_cy_elements → Go: AsCyElements
		"AsCyElements",
	}

	agType := reflect.TypeOf(&AnalysisGraph{})
	for _, name := range expectedMethods {
		_, ok := agType.MethodByName(name)
		if !ok {
			t.Errorf("missing method *AnalysisGraph.%s (required by Python ivy_art.py)", name)
		}
	}
}

// TestPortCompleteness_StandaloneFunctions verifies that standalone
// Python functions from ivy_art.py exist as Go package-level functions.
func TestPortCompleteness_StandaloneFunctions(t *testing.T) {
	// Verify the functions exist by referencing them.
	// If any is missing, this file won't compile.
	_ = LabelFromAction
	_ = NewAC
	_ = NewAnalysisSubgraph
	_ = IsActionApp
	_ = IsStateJoin
	_ = NewActionApp
	_ = NewStateJoin
	_ = NewAnalysisGraph
	_ = NewState
}

// TestPortCompleteness_Types verifies that all Python classes from
// ivy_art.py have corresponding Go types.
func TestPortCompleteness_Types(t *testing.T) {
	// Python: AC → Go: AC
	_ = AC{}
	// Python: Counterexample → Go: Counterexample
	_ = Counterexample{}
	// Python: AnalysisGraph → Go: AnalysisGraph
	_ = AnalysisGraph{}
	// Python: AnalysisSubgraph → Go: AnalysisSubgraph
	_ = AnalysisSubgraph{}
	// Python: ActionApp expression → Go: ActionApp
	_ = ActionApp{}
	// Python: state_join expression → Go: StateJoin
	_ = StateJoin{}
}

// TestPortCompleteness_CounterexampleBoolFalse verifies Python's
// __bool__ returning False is ported as IsFailed returning false.
func TestPortCompleteness_CounterexampleBoolFalse(t *testing.T) {
	cex := &Counterexample{Msg: "test"}
	// Python: bool(Counterexample(...)) == False
	if cex.IsFailed() {
		t.Error("IsFailed should return false (matching Python __bool__ → False)")
	}
}
