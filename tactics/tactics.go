// Package tactics implements interactive refinement tactics for Ivy.
// This corresponds to Python's tactics.py and tactics_api.py.
//
// Tactics operate on the proof goal stack, refining abstract states
// through the UPDR (Universal Property-Directed Reachability) algorithm.
// They are used in the interactive concept graph UI.
package tactics

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
	"github.com/glycerine/goivy/solver"
	"github.com/glycerine/goivy/transrel"
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
func (tc *TacticsContext) TopGoal() *proof.ProofGoal {
	return tc.Goals.Top()
}

// PushGoal pushes a new goal onto the stack.
func (tc *TacticsContext) PushGoal(goal *proof.ProofGoal) {
	tc.Goals.Push(goal)
}

// RemoveGoal removes a goal from the stack.
func (tc *TacticsContext) RemoveGoal(goal *proof.ProofGoal) {
	tc.Goals.Remove(goal)
}

// GoalAtArgNode creates a new goal at an analysis graph node.
// Corresponds to Python's goal_at_arg_node().
func GoalAtArgNode(formula lg.Node, node *art.State) *proof.ProofGoal {
	return &proof.ProofGoal{Formula: formula, Node: node}
}

// BackgroundTheory returns the background theory for the module.
func (tc *TacticsContext) BackgroundTheory() lg.Node {
	if tc.Mod == nil {
		return lg.True
	}
	clauses := tc.Mod.BackgroundTheory(nil)
	if clauses == nil {
		return lg.True
	}
	return clauses.ToFormula()
}

// RefutedGoal checks if a goal has been refuted.
// A goal is refuted if its node's clauses conjoined with axioms imply
// the negation of the goal formula.
// Corresponds to Python's refuted_goal().
func (tc *TacticsContext) RefutedGoal(goal *proof.ProofGoal) bool {
	if goal == nil || goal.Formula == nil {
		return false
	}
	// Quick check: formula is False
	if or, ok := goal.Formula.(*lg.Or); ok && len(or.Terms) == 0 {
		return true
	}
	// Full check: axioms & node.clauses => ~goal.formula
	node, ok := goal.Node.(*art.State)
	if !ok || node == nil || node.Clauses == nil {
		return false
	}
	axioms := tc.BackgroundTheory()
	premise := conjoinNodes(axioms, node.Clauses.ToFormula())
	negGoal := &lg.Not{Body: goal.Formula}
	slv := solver.New()
	result, err := slv.Implies(premise, negGoal)
	if err != nil {
		return false
	}
	return result
}

// ForwardImage computes the forward image of clauses through an action.
// Corresponds to Python's forward_image(pre_fact, action).
func (tc *TacticsContext) ForwardImage(preFact *clauseops.Clauses, action actions.Action) *clauseops.Clauses {
	if preFact == nil || action == nil {
		return preFact
	}
	axioms := tc.BackgroundTheory()
	update := actions.GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return preFact
	}
	preFormula := preFact.ToFormula()
	resultFormula := transrel.ForwardImage(preFormula, axioms, update)
	return clauseops.FormulaToClauses(resultFormula, preFact.Annot)
}

// BackwardImage computes the backward image (reverse image / weakest precondition)
// of clauses through an action.
// Corresponds to Python's backward_image(post_fact, action).
func (tc *TacticsContext) BackwardImage(postFact *clauseops.Clauses, action actions.Action) *clauseops.Clauses {
	if postFact == nil || action == nil {
		return postFact
	}
	axioms := tc.BackgroundTheory()
	update := actions.GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return postFact
	}
	postFormula := postFact.ToFormula()
	resultFormula := transrel.ReverseImage(postFormula, axioms, update)
	return clauseops.FormulaToClauses(resultFormula, postFact.Annot)
}

// ImpliedFacts checks which facts are implied by a premise.
// Returns the subset of factsToCheck that are implied by premise conjoined
// with background axioms.
// Corresponds to Python's implied_facts().
func (tc *TacticsContext) ImpliedFacts(premise *clauseops.Clauses, factsToCheck []*clauseops.Clauses) []*clauseops.Clauses {
	if premise == nil || len(factsToCheck) == 0 {
		return nil
	}
	axioms := tc.BackgroundTheory()
	premFormula := conjoinNodes(axioms, premise.ToFormula())

	var implied []*clauseops.Clauses
	slv := solver.New()
	for _, fact := range factsToCheck {
		if fact == nil {
			continue
		}
		factFormula := fact.ToFormula()
		result, err := slv.Implies(premFormula, factFormula)
		if err == nil && result {
			implied = append(implied, fact)
		}
	}
	return implied
}

// GetDiagram returns the diagram (abstract state representation) for a goal.
// Corresponds to Python's get_diagram().
func (tc *TacticsContext) GetDiagram(goal *proof.ProofGoal) *proof.ProofGoal {
	if goal == nil {
		return nil
	}
	// In the full implementation, this calls ivy_solver.clauses_model_to_diagram
	// to compute a minimal diagram from the goal's node clauses conjoined with
	// the goal formula. For now, return the goal itself.
	return goal
}

// RefineOrReverse attempts to refine or reverse a goal.
// Refinement succeeds if a forward interpolant exists between the predecessor
// state and the goal formula. If refinement fails, the backward image is
// computed and a new goal is pushed.
// Returns (refined bool, result) where result is either the interpolant
// (if refined) or a new proof goal (if reversed).
// Corresponds to Python's refine_or_reverse().
func (tc *TacticsContext) RefineOrReverse(goal *proof.ProofGoal) (bool, interface{}) {
	if goal == nil {
		return false, nil
	}
	node, ok := goal.Node.(*art.State)
	if !ok || node == nil || node.Pred == nil {
		return false, nil
	}
	pred := node.Pred
	if pred.Clauses == nil {
		return false, nil
	}

	// Get the action that transitions from pred to node
	action := node.Action
	if action == nil {
		return false, nil
	}

	update := actions.GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return false, nil
	}

	// Try forward interpolation:
	// Check if pred.clauses & TR => ~goal.formula
	// If so, we can refine (find an interpolant).
	preFmla := pred.Clauses.ToFormula()
	goalFmla := goal.Formula
	postFmla := conjoinNodes(preFmla, update.TRNode())
	negGoal := &lg.Not{Body: goalFmla}

	slv := solver.New()
	implies, err := slv.Implies(postFmla, negGoal)
	if err == nil && implies {
		// Refinement succeeds: the goal is unreachable from pred.
		// The interpolant is the forward image of pred through the action
		// conjoined with the negation of the goal.
		// For now, return the negation of the goal as the "new fact".
		return true, negGoal
	}

	// Refinement fails: compute backward image and push new goal.
	goalClauses := clauseops.FormulaToClauses(goalFmla, nil)
	bi := tc.BackwardImage(goalClauses, action)
	newGoal := GoalAtArgNode(bi.ToFormula(), pred)
	return false, newGoal
}

// -----------------------------------------------------------------------
// ARG helpers — corresponds to tactics_api.py arg_* functions
// -----------------------------------------------------------------------

// ArgGetFact returns the clauses at an ARG node.
func ArgGetFact(node *art.State) *clauseops.Clauses {
	if node == nil {
		return nil
	}
	return node.Clauses
}

// ArgAddFacts adds facts (conjoins clauses) to an ARG node.
func ArgAddFacts(node *art.State, facts ...*clauseops.Clauses) {
	if node == nil {
		return
	}
	for _, f := range facts {
		if f != nil {
			node.Clauses = clauseops.AndClausesTyped(node.Clauses, f)
		}
	}
}

// ArgGetPredAction returns the predecessor and action for a node.
func ArgGetPredAction(node *art.State) (*art.State, actions.Action) {
	if node == nil {
		return nil, nil
	}
	return node.Pred, node.Action
}

// ArgGetConjuncts returns the individual formula conjuncts at a node.
func ArgGetConjuncts(node *art.State) []lg.Node {
	if node == nil || node.Clauses == nil {
		return nil
	}
	return node.Clauses.Fmlas
}

// GetBigAction creates a nondeterministic choice over all public actions.
// Corresponds to Python's get_big_action().
func GetBigAction(ag *art.AnalysisGraph) actions.Action {
	var branches []lg.Node
	for name := range ag.PublicActions {
		if act, ok := ag.Actions[name]; ok {
			if a, ok := act.(actions.Action); ok {
				branches = append(branches, actions.WrapAction(a))
			}
		}
	}
	if len(branches) == 0 {
		return actions.NewSequence()
	}
	return actions.NewEnvAction(branches...)
}

// -----------------------------------------------------------------------
// Concrete tactic implementations
// -----------------------------------------------------------------------

// RemoveIfRefuted removes a goal if it has been refuted.
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
type RemoveGoalTactic struct {
	TC *TacticsContext
}

func (t *RemoveGoalTactic) Name() string { return "RemoveGoal" }

func (t *RemoveGoalTactic) Apply(goal *proof.ProofGoal) (bool, error) {
	t.TC.RemoveGoal(goal)
	return true, nil
}

// RefineOrReverseTactic attempts refinement or reversal.
type RefineOrReverseTactic struct {
	TC         *TacticsContext
	AutoRemove bool
}

func (t *RefineOrReverseTactic) Name() string { return "RefineOrReverse" }

func (t *RefineOrReverseTactic) Apply(goal *proof.ProofGoal) (bool, error) {
	refined, result := t.TC.RefineOrReverse(goal)
	if refined {
		// Add the new fact to the goal's node
		if newFact, ok := result.(lg.Node); ok {
			node, ok := goal.Node.(*art.State)
			if ok && node != nil {
				factClauses := clauseops.FormulaToClauses(newFact, nil)
				ArgAddFacts(node, factClauses)
			}
		}
		// Remove refuted goals in the chain
		if t.AutoRemove {
			g := goal
			for g != nil && t.TC.RefutedGoal(g) {
				t.TC.RemoveGoal(g)
				g = t.TC.TopGoal()
			}
		}
		return true, nil
	}
	// Reversed: push the new goal
	if newGoal, ok := result.(*proof.ProofGoal); ok {
		t.TC.PushGoal(newGoal)
	}
	return false, nil
}

// UPDR implements the Universal Property-Directed Reachability tactic.
// This iteratively builds a sequence of frames (over-approximations of
// reachable states) and tries to block counterexamples by refining frames.
type UPDR struct {
	TC        *TacticsContext
	MaxFrames int // 0 means unlimited
}

func (t *UPDR) Name() string { return "UPDR" }

func (t *UPDR) Apply(goal *proof.ProofGoal) (bool, error) {
	if t.TC.AG == nil {
		return false, fmt.Errorf("UPDR: no analysis graph")
	}

	frames := t.TC.AG.States
	if len(frames) == 0 {
		return false, fmt.Errorf("UPDR: no initial state")
	}

	// The safety property (negated) is the bad states
	badStates := goal.Formula
	if badStates == nil {
		return false, fmt.Errorf("UPDR: no goal formula")
	}

	action := GetBigAction(t.TC.AG)
	initFrame := frames[0]
	lastFrame := frames[len(frames)-1]

	maxIter := 100
	if t.MaxFrames > 0 {
		maxIter = t.MaxFrames
	}

	for iter := 0; iter < maxIter; iter++ {
		// Check if we found an inductive invariant
		for i := 0; i < len(frames)-1; i++ {
			if t.TC.AG.Cover(frames[i+1], frames[i]) {
				return true, nil // Found inductive invariant
			}
		}

		// Add new frame
		lastFrame = t.TC.AG.Execute(action, lastFrame, nil, "")
		if lastFrame == nil {
			break
		}
		frames = t.TC.AG.States

		// Push bad states as goal for the new frame
		newGoal := GoalAtArgNode(badStates, lastFrame)
		t.TC.PushGoal(newGoal)

		// Process goals
		for t.TC.Goals.Len() > 0 {
			currentGoal := t.TC.TopGoal()
			if currentGoal == nil {
				break
			}

			// Remove if refuted
			if t.TC.RefutedGoal(currentGoal) {
				t.TC.RemoveGoal(currentGoal)
				continue
			}

			// Check if goal is at initial frame (no invariant)
			if goalNode, ok := currentGoal.Node.(*art.State); ok {
				if goalNode == initFrame {
					return false, fmt.Errorf("UPDR: no invariant found (counterexample reaches initial state)")
				}
			}

			// Try to refine or reverse
			refined, result := t.TC.RefineOrReverse(currentGoal)
			if refined {
				// Add the learned fact
				if newFact, ok := result.(lg.Node); ok {
					if goalNode, ok := currentGoal.Node.(*art.State); ok {
						factClauses := clauseops.FormulaToClauses(newFact, nil)
						ArgAddFacts(goalNode, factClauses)
					}
				}
				// Remove refuted goals
				for t.TC.Goals.Len() > 0 {
					g := t.TC.TopGoal()
					if g == nil || !t.TC.RefutedGoal(g) {
						break
					}
					t.TC.RemoveGoal(g)
				}
			} else {
				// Push new goal from backward image
				if newGoal, ok := result.(*proof.ProofGoal); ok {
					t.TC.PushGoal(newGoal)
				} else {
					// Can't make progress
					break
				}
			}
		}
	}

	return false, fmt.Errorf("UPDR: reached iteration limit without finding invariant")
}

// CheckCover checks if one abstract state covers another.
type CheckCover struct {
	TC *TacticsContext
}

func (t *CheckCover) Name() string { return "CheckCover" }

func (t *CheckCover) Apply(goal *proof.ProofGoal) (bool, error) {
	if t.TC.AG == nil || goal == nil {
		return false, nil
	}
	goalNode, ok := goal.Node.(*art.State)
	if !ok || goalNode == nil {
		return false, nil
	}
	// Check if any other state covers this one
	for _, state := range t.TC.AG.States {
		if state != goalNode && t.TC.AG.Cover(goalNode, state) {
			return true, nil
		}
	}
	return false, nil
}

// PathReach checks if a goal is reachable via BMC.
type PathReach struct {
	TC    *TacticsContext
	Bound *int // nil for unbounded
}

func (t *PathReach) Name() string { return "PathReach" }

func (t *PathReach) Apply(goal *proof.ProofGoal) (bool, error) {
	if t.TC.AG == nil || goal == nil {
		return false, nil
	}
	goalNode, ok := goal.Node.(*art.State)
	if !ok || goalNode == nil {
		return false, nil
	}
	result := t.TC.AG.BMC(goalNode, goal.Formula, nil, t.Bound)
	return result != nil, nil
}

// PushDiagram pushes the diagram of a goal onto the goal stack.
type PushDiagram struct {
	TC     *TacticsContext
	Weaken bool
}

func (t *PushDiagram) Name() string { return "PushDiagram" }

func (t *PushDiagram) Apply(goal *proof.ProofGoal) (bool, error) {
	dg := t.TC.GetDiagram(goal)
	if dg != nil {
		t.TC.PushGoal(dg)
	}
	return true, nil
}

// ExecuteAction executes an action at a node.
type ExecuteAction struct {
	TC         *TacticsContext
	Action     actions.Action
	Abstractor art.Abstractor
}

func (t *ExecuteAction) Name() string { return "ExecuteAction" }

func (t *ExecuteAction) Apply(goal *proof.ProofGoal) (bool, error) {
	if t.TC.AG == nil || goal == nil || t.Action == nil {
		return false, nil
	}
	goalNode, ok := goal.Node.(*art.State)
	if !ok || goalNode == nil {
		return false, nil
	}
	t.TC.AG.Execute(t.Action, goalNode, t.Abstractor, "")
	return true, nil
}

// -----------------------------------------------------------------------
// Helper functions
// -----------------------------------------------------------------------

func conjoinNodes(a, b lg.Node) lg.Node {
	if a == nil || lg.IsTrue(a) {
		return b
	}
	if b == nil || lg.IsTrue(b) {
		return a
	}
	and, _ := lg.NewAnd(a, b)
	return and
}
