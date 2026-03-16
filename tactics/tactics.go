// Package tactics implements interactive refinement tactics for Ivy.
// This corresponds to Python's tactics.py and tactics_api.py.
//
// Tactics operate on the proof goal stack, refining abstract states
// through the UPDR (Universal Property-Directed Reachability) algorithm.
// They are used in the interactive concept graph UI.
package tactics

import (
	"fmt"

	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
)

// Tactic is the interface for proof refinement tactics.
type Tactic interface {
	Apply(goal *proof.ProofGoal) (bool, error)
	Name() string
}

// TacticStep records a single step in the tactic execution.
type TacticStep struct {
	Goal   *proof.ProofGoal
	Active *proof.ProofGoal
	Info   map[string]interface{}
}

// -----------------------------------------------------------------------
// Tactics API — corresponds to Python tactics_api.py
// -----------------------------------------------------------------------

// TacticsContext holds the state for interactive proof exploration.
type TacticsContext struct {
	AG    *art.AnalysisGraph
	Mod   *module.Module
	Goals *proof.ProofGoalStack
}

// NewTacticsContext creates a new tactics context.
func NewTacticsContext(ag *art.AnalysisGraph, mod *module.Module) *TacticsContext {
	return &TacticsContext{
		AG:    ag,
		Mod:   mod,
		Goals: proof.NewProofGoalStack(),
	}
}

// TopGoal returns the current top goal.
// Corresponds to Python's top_goal().
func (tc *TacticsContext) TopGoal() *proof.ProofGoal {
	return tc.Goals.Top()
}

// PushGoal pushes a new goal onto the stack.
// Corresponds to Python's push_goal().
func (tc *TacticsContext) PushGoal(goal *proof.ProofGoal) {
	tc.Goals.Push(goal)
}

// RemoveGoal removes a goal from the stack.
// Corresponds to Python's remove_goal().
func (tc *TacticsContext) RemoveGoal(goal *proof.ProofGoal) {
	tc.Goals.Remove(goal)
}

// RefutedGoal checks if a goal has been refuted (found to be unreachable).
// Corresponds to Python's refuted_goal().
func (tc *TacticsContext) RefutedGoal(goal *proof.ProofGoal) bool {
	// A goal is refuted if its formula is False (empty Or)
	if goal == nil || goal.Formula == nil {
		return false
	}
	if or, ok := goal.Formula.(*lg.Or); ok && len(or.Terms) == 0 {
		return true
	}
	return false
}

// ForwardImage computes the forward image of clauses through an action.
// Corresponds to Python's forward_image().
func (tc *TacticsContext) ForwardImage(clauses *clauseops.Clauses) *clauseops.Clauses {
	// Placeholder — full implementation requires transrel.Update integration
	return clauses
}

// BackwardImage computes the backward image of clauses through an action.
// Corresponds to Python's backward_image().
func (tc *TacticsContext) BackwardImage(clauses *clauseops.Clauses) *clauseops.Clauses {
	// Placeholder — full implementation requires transrel.Update integration
	return clauses
}

// ImpliedFacts returns facts implied by the current state and axioms.
// Corresponds to Python's implied_facts().
func (tc *TacticsContext) ImpliedFacts(state *art.State) []lg.Node {
	// Placeholder
	return nil
}

// GetDiagram returns the diagram (abstract state representation) for a goal.
// Corresponds to Python's get_diagram().
func (tc *TacticsContext) GetDiagram(goal *proof.ProofGoal) lg.Node {
	if goal == nil {
		return nil
	}
	return goal.Formula
}

// RefineOrReverse attempts to refine or reverse a goal.
// Returns (refined, new_facts).
// Corresponds to Python's refine_or_reverse().
func (tc *TacticsContext) RefineOrReverse(goal *proof.ProofGoal) (bool, []lg.Node) {
	// Placeholder — needs full UPDR integration
	return false, nil
}

// -----------------------------------------------------------------------
// Concrete tactic implementations
// -----------------------------------------------------------------------

// RemoveIfRefuted removes a goal if it has been refuted.
// Corresponds to Python's RemoveIfRefuted.
type RemoveIfRefuted struct {
	TC *TacticsContext
}

func (t *RemoveIfRefuted) Name() string { return "RemoveIfRefuted" }

func (t *RemoveIfRefuted) Apply(goal *proof.ProofGoal) (bool, error) {
	if t.TC.RefutedGoal(goal) {
		t.TC.RemoveGoal(goal)
		return true, nil
	}
	return false, nil
}

// RemoveGoalTactic removes a goal unconditionally.
// Corresponds to Python's RemoveGoal.
type RemoveGoalTactic struct {
	TC *TacticsContext
}

func (t *RemoveGoalTactic) Name() string { return "RemoveGoal" }

func (t *RemoveGoalTactic) Apply(goal *proof.ProofGoal) (bool, error) {
	t.TC.RemoveGoal(goal)
	return true, nil
}

// RefineOrReverseTactic attempts refinement or reversal.
// Corresponds to Python's RefineOrReverse.
type RefineOrReverseTactic struct {
	TC         *TacticsContext
	AutoRemove bool
}

func (t *RefineOrReverseTactic) Name() string { return "RefineOrReverse" }

func (t *RefineOrReverseTactic) Apply(goal *proof.ProofGoal) (bool, error) {
	refined, _ := t.TC.RefineOrReverse(goal)
	if !refined {
		return false, nil
	}
	if t.AutoRemove {
		// Remove refuted goals in the chain
		g := goal
		for g != nil && t.TC.RefutedGoal(g) {
			t.TC.RemoveGoal(g)
			g = t.TC.TopGoal()
		}
	}
	return true, nil
}

// UPDR implements the Universal Property-Directed Reachability tactic.
// Corresponds to Python's UPDR.
type UPDR struct {
	TC *TacticsContext
}

func (t *UPDR) Name() string { return "UPDR" }

func (t *UPDR) Apply(goal *proof.ProofGoal) (bool, error) {
	// UPDR is an iterative algorithm that maintains a sequence of
	// over-approximations of reachable states and repeatedly tries
	// to push/block counterexamples.
	// Full implementation requires:
	//   - Backward image computation
	//   - Interpolation or generalization
	//   - Inductive invariant checking
	return false, fmt.Errorf("UPDR tactic not yet fully implemented")
}

// CheckCover checks if one abstract state covers another.
// Corresponds to Python's CheckCover.
type CheckCover struct {
	TC *TacticsContext
}

func (t *CheckCover) Name() string { return "CheckCover" }

func (t *CheckCover) Apply(goal *proof.ProofGoal) (bool, error) {
	// Check if the goal's state is covered by existing states
	return false, nil
}
