// Package tactics implements interactive refinement tactics for Ivy.
// This corresponds to Python's tactics.py and tactics_api.py.
//
// Tactics operate on the proof goal stack, refining abstract states
// through the UPDR (Universal Property-Directed Reachability) algorithm.
// They are used in the interactive concept graph UI.
package tactics

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/art"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

const checkPrecondTrue = true
const checkPrecondFalse = false

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

// modSig returns tc.Mod.Sig if available, nil otherwise.
func (tc *TacticsContext) modSig() *il.Sig {
	if tc.Mod != nil {
		return tc.Mod.Sig
	}
	return nil
}

// modSolverOpts returns tc.Mod.Cfg.SolverOpts if available, nil otherwise.
func (tc *TacticsContext) modSolverOpts() *module.SolverOptions {
	if tc.Mod != nil && tc.Mod.Cfg != nil {
		return tc.Mod.Cfg.SolverOpts
	}
	return nil
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
func GoalAtArgNode(formula lg.Expr, node *art.State) *proof.ProofGoal {
	return &proof.ProofGoal{Formula: formula, Node: node}
}

// BackgroundTheory returns the background theory for the module as a formula.
func (tc *TacticsContext) BackgroundTheory() lg.Expr {
	if tc.Mod == nil {
		return lg.True
	}
	clauses := tc.Mod.BackgroundTheory(nil)
	if clauses == nil {
		return lg.True
	}
	return clauses.ToFormula()
}

// BackgroundTheoryClauses returns the background theory as Clauses.
// Preserves Clauses structure for and_clauses operations.
// Corresponds to Python _ivy_interp.background_theory().
func (tc *TacticsContext) BackgroundTheoryClauses() *module.Clauses {
	if tc.Mod == nil {
		return module.TrueClauses(nil)
	}
	clauses := tc.Mod.BackgroundTheory(nil)
	if clauses == nil {
		return module.TrueClauses(nil)
	}
	return clauses
}

// RefutedGoal checks if a goal has been refuted.
// A goal is refuted if its node's clauses conjoined with axioms imply
// the negation of the goal formula.
// Corresponds to Python tactics_api.py:refuted_goal (lines 336-341).
func (tc *TacticsContext) RefutedGoal(goal *proof.ProofGoal) bool {
	if goal == nil || goal.Formula == nil {
		return false
	}
	// Quick check: formula is False (empty Or)
	if or, ok := goal.Formula.(*lg.Or); ok && len(or.Terms) == 0 {
		return true
	}
	node, ok := goal.Node.(*art.State)
	if !ok || node == nil || node.Clauses == nil {
		return false
	}

	// Python: axioms = _ivy_interp.background_theory()
	axioms := tc.BackgroundTheoryClauses()

	// Python: premise = (and_clauses(axioms, goal.node.clauses)).to_formula()
	combined := module.AndClausesTyped(axioms, node.Clauses)
	premise := combined.ToFormula()

	// Python: f = Not(goal.formula.to_formula())
	negGoal := &lg.Not{Body: goal.Formula}

	// Python: return z3_implies(premise, f)
	slv := z3bridge.NewSolver(tc.Mod, tc.modSolverOpts())
	result, err := slv.Z3Implies(premise, negGoal, false)
	if err != nil {
		return false
	}
	return result
}

// ForwardImage computes the forward image of clauses through an action.
// Corresponds to Python's forward_image(pre_fact, action).
func (tc *TacticsContext) ForwardImage(preFact *module.Clauses, action actions.Action) *module.Clauses {
	if preFact == nil || action == nil {
		return preFact
	}
	axioms := tc.BackgroundTheoryClauses()
	update := actions.GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return preFact
	}
	return actions.ForwardImage(preFact, axioms, update)
}

// BackwardImage computes the backward image (reverse image / weakest precondition)
// of clauses through an action.
// Corresponds to Python's backward_image(post_fact, action).
func (tc *TacticsContext) BackwardImage(postFact *module.Clauses, action actions.Action) *module.Clauses {
	if postFact == nil || action == nil {
		return postFact
	}
	axioms := tc.BackgroundTheoryClauses()
	update := actions.GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return postFact
	}
	return actions.ReverseImage(postFact, axioms, update)
}

// ImpliedFacts checks which facts are implied by a premise.
// Returns the subset of factsToCheck that are implied by premise conjoined
// with background axioms.
// Corresponds to Python tactics_api.py:implied_facts (lines 310-320).
func (tc *TacticsContext) ImpliedFacts(premise *module.Clauses, factsToCheck []*module.Clauses) []*module.Clauses {
	if premise == nil || len(factsToCheck) == 0 {
		return nil
	}

	// Python: axioms = _ivy_interp.background_theory()
	axioms := tc.BackgroundTheoryClauses()

	// Python: premise = normalize_quantifiers((and_clauses(axioms, premise)).to_formula())
	combined := module.AndClausesTyped(axioms, premise)
	premFormula := lu.NormalizeQuantifiers(combined.ToFormula())

	// Python: facts_to_check = [f.to_formula() if type(f) is Clauses else f ...]
	formulas := make([]lg.Expr, 0, len(factsToCheck))
	indices := make([]int, 0, len(factsToCheck))
	for i, fact := range factsToCheck {
		if fact != nil {
			formulas = append(formulas, fact.ToFormula())
			indices = append(indices, i)
		}
	}

	// Python: result = z3_implies_batch(premise, facts_to_check, False)
	slv := z3bridge.NewSolver(tc.Mod, tc.modSolverOpts())
	results, err := slv.ImpliesBatch(premFormula, formulas, false)
	if err != nil {
		return nil
	}

	// Python: return [f for f, x in zip(facts_to_check, result) if x]
	var implied []*module.Clauses
	for j, isImplied := range results {
		if isImplied {
			implied = append(implied, factsToCheck[indices[j]])
		}
	}
	return implied
}

// GetDiagram returns the diagram (abstract state representation) for a goal.
// Python tactics_api.py:323-333 get_diagram(goal, weaken=False).
func (tc *TacticsContext) GetDiagram(goal *proof.ProofGoal, weaken bool) *proof.ProofGoal {
	if goal == nil {
		return nil
	}
	node, ok := goal.Node.(*art.State)
	if !ok || node == nil {
		return nil
	}

	// Python: axioms = _ivy_interp.background_theory()
	axioms := tc.BackgroundTheoryClauses()

	// Python: and_clauses(goal.node.clauses, goal.formula)
	goalClauses := module.FormulaToClauses(goal.Formula, nil)
	combined := module.AndClausesTyped(node.Clauses, goalClauses)

	// Python: is_skolem
	isSkolem := func(c *lg.Const) bool {
		return actions.IsSkolem(c.Name)
	}

	// Python: ivy_solver.clauses_model_to_diagram(
	//     combined, is_skolem, false_clauses(), axioms=axioms, weaken=weaken)
	slv := z3bridge.NewSolver(tc.Mod, tc.modSolverOpts())
	d, err := slv.ClausesModelToDiagramFull(
		combined,
		isSkolem,
		module.FalseClauses(nil),
		nil,
		axioms,
		weaken,
		true,
		true,
	)
	if err != nil || d == nil {
		return nil
	}

	// Python: return goal_at_arg_node(d, goal.node)
	return GoalAtArgNode(d.ToFormula(), node)
}

// RefineOrReverse attempts to refine or reverse a goal.
// Returns (true, *module.Clauses) with the interpolant if refinement succeeds,
// or (false, *proof.ProofGoal) with the backward image goal if it fails.
// Python tactics_api.py:286-307 refine_or_reverse(goal).
func (tc *TacticsContext) RefineOrReverse(goal *proof.ProofGoal) (bool, interface{}) {
	if goal == nil {
		return false, nil
	}
	node, ok := goal.Node.(*art.State)
	if !ok || node == nil {
		return false, nil
	}

	// Python: preds, action = arg_get_preds_action(goal.node)
	pred := node.Pred
	action := node.Action
	if pred == nil || action == nil {
		return false, nil
	}

	// Python: axioms = _ivy_interp.background_theory()
	axioms := tc.BackgroundTheoryClauses()

	// Python: action.update(_ivy_interp, None)
	update := actions.GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return false, nil
	}

	// Python: x = ivy_transrel.forward_interpolant(
	//     pred.clauses, action.update(...), goal.formula, axioms, None)
	goalClauses := module.FormulaToClauses(goal.Formula, nil)
	fmt.Println("calling ForwardInterpolant")
	x := actions.ForwardInterpolant(tc.Mod, pred.Clauses, update, goalClauses, axioms, nil)
	fmt.Printf("    got: %v\n", x)

	if x == nil {
		// Python: bi = backward_image(goal.formula, action)
		//         return False, goal_at_arg_node(bi, pred)
		bi := tc.BackwardImage(goalClauses, action)
		newGoal := GoalAtArgNode(bi.ToFormula(), pred)
		return false, newGoal
	}
	// Python: return True, x[1]  # x is (core, interpolant)
	return true, x.Itp
}

// -----------------------------------------------------------------------
// ARG helpers — corresponds to tactics_api.py arg_* functions
// -----------------------------------------------------------------------

// ArgGetFact returns the clauses at an ARG node.
func ArgGetFact(node *art.State) *module.Clauses {
	if node == nil {
		return nil
	}
	return node.Clauses
}

// ArgAddFacts adds facts (conjoins clauses) to an ARG node.
func ArgAddFacts(node *art.State, facts ...*module.Clauses) {
	if node == nil {
		return
	}
	for _, f := range facts {
		if f != nil {
			node.Clauses = module.AndClausesTyped(node.Clauses, f)
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
func ArgGetConjuncts(node *art.State) []lg.Expr {
	if node == nil || node.Clauses == nil {
		return nil
	}
	return node.Clauses.Fmlas
}

// GetBigAction creates a nondeterministic choice over all public actions.
// Corresponds to Python's get_big_action().
func GetBigAction(ag *art.AnalysisGraph) actions.Action {
	var branches []lg.Expr
	for name := range ag.PublicActions.All() {
		if act, ok := ag.Actions.Get2(name); ok {
			if a, ok := act.(actions.Action); ok {
				branches = append(branches, a)
			}
		}
	}
	if len(branches) == 0 {
		return actions.NewSequence()
	}
	return actions.NewEnvActionOn(ag.Domain.Cfg.ActCfg, branches...)
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
		// Python: ta.arg_add_facts(goal.node, y)
		if newFact, ok := result.(*module.Clauses); ok {
			node, ok := goal.Node.(*art.State)
			if ok && node != nil {
				ArgAddFacts(node, newFact)
			}
		}
		// Python: walk g = g.parent removing refuted goals
		if t.AutoRemove {
			g := goal
			for g != nil && t.TC.RefutedGoal(g) {
				t.TC.RemoveGoal(g)
				g = g.Parent
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
		var execErr error
		lastFrame, execErr = t.TC.AG.Execute(checkPrecondTrue, action, lastFrame, nil, "")
		if execErr != nil {
			return false, fmt.Errorf("Execute failed: %w", execErr)
		}
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
				if newFact, ok := result.(lg.Expr); ok {
					if goalNode, ok := currentGoal.Node.(*art.State); ok {
						factClauses := module.FormulaToClauses(newFact, nil)
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
	_, err := t.TC.AG.Execute(checkPrecondTrue, t.Action, goalNode, t.Abstractor, "")
	if err != nil {
		return false, fmt.Errorf("ExecuteAction: %w", err)
	}
	return true, nil
}

// -----------------------------------------------------------------------
// Additional tactics_api / tactics functions for the iupdr.py port
// -----------------------------------------------------------------------

// GetSafetyProperty mirrors Python tactics_api.py:60-64:
//
//	def get_safety_property():
//	    """Return the safety property"""
//	    return _ivy_interp.conjs[0]
//
// In Python this returns the first conjecture from the interpreter; in
// Go the conjectures live on the module and are accessible via
// Module.Conjs().
func (tc *TacticsContext) GetSafetyProperty() *module.Clauses {
	if tc.AG == nil || tc.AG.Domain == nil {
		return module.TrueClauses(nil)
	}
	conjs := tc.AG.Domain.Conjs()
	if len(conjs) == 0 {
		return module.TrueClauses(nil)
	}
	return conjs[0]
}

// ArgAddActionNode mirrors Python tactics_api.py:109-120:
//
//	def arg_add_action_node(pre, action, abstractor=None):
//	    """Add a new node with an action edge from pre, and return it."""
//	    try:
//	        label = [k for k, v in _ivy_ag.actions.items() if v == action][0]
//	    except IndexError:
//	        label = None
//	    node = _ivy_ag.execute(action, pre, abstractor, label)
//	    if abstractor is None:
//	        node.clauses = true_clauses()
//	    return node
func (tc *TacticsContext) ArgAddActionNode(pre *art.State, action actions.Action, abstractor art.Abstractor) *art.State {
	if tc.AG == nil {
		return nil
	}
	// Python: try to find the action's registered label.
	label := ""
	for k := range tc.AG.Actions.All() {
		if act, ok := tc.AG.Actions.Get2(k); ok {
			if act == action {
				label = k
				break
			}
		}
	}
	node, err := tc.AG.Execute(checkPrecondTrue, action, pre, abstractor, label)
	if err != nil || node == nil {
		return nil
	}
	if abstractor == nil {
		// Python: node.clauses = true_clauses()
		node.Clauses = module.TrueClauses(nil)
	}
	return node
}

// ArgGetPred mirrors Python tactics_api.py:251-253:
//
//	def arg_get_pred(node):
//	    assert node.pred is not None
//	    return node.pred
func ArgGetPred(node *art.State) *art.State {
	if node == nil || node.Pred == nil {
		panic("arg_get_pred: node.pred is nil")
	}
	return node.Pred
}

// ArgIsCovered mirrors Python tactics_api.py:102-106:
//
//	def arg_is_covered(covered, by):
//	    """Returns True if covered is covered by 'by' in the arg"""
//	    return _ivy_ag.cover(covered, by)
func (tc *TacticsContext) ArgIsCovered(covered, by *art.State) bool {
	if tc.AG == nil {
		return false
	}
	return tc.AG.Cover(covered, by)
}

// CheckCover mirrors Python tactics.py:206-211 (the @tactic
// CheckCover.apply body, which becomes a top-level function via the
// @tactic decorator):
//
//	@tactic
//	class CheckCover(Tactic):
//	    def apply(self, covered, by):
//	        result = ta.arg_is_covered(covered, by)
//	        self.step(covered=covered, by=by, result=result)
//	        return result
//
// (The struct CheckCover above is the goal-tactic form; this function
// is the iupdr-style direct call.)
func CheckCoverFn(tc *TacticsContext, covered, by *art.State) bool {
	return tc.ArgIsCovered(covered, by)
}

// RemoveIfRefutedFn mirrors Python tactics.py:29-36 (the @tactic
// RemoveIfRefuted.apply body, exposed as a function via the decorator):
//
//	@goal_tactic
//	@tactic
//	class RemoveIfRefuted(Tactic):
//	    def apply(self, goal):
//	        if ta.refuted_goal(goal):
//	            ta.remove_goal(goal)
//	            self.step(goal=goal, active=ta.top_goal())
//	            return True
//	        else:
//	            return False
func RemoveIfRefutedFn(tc *TacticsContext, goal *proof.ProofGoal) bool {
	if tc.RefutedGoal(goal) {
		tc.RemoveGoal(goal)
		return true
	}
	return false
}

// RecalculateFactsFn mirrors Python tactics.py:156-177 (the @tactic
// RecalculateFacts.apply body):
//
//	@tactic
//	class RecalculateFacts(Tactic):
//	    def apply(self, node, facts):
//	        preds, action = ta.arg_get_preds_action(node)
//	        assert action != 'join'
//	        assert len(preds) == 1
//	        pred = preds[0]
//	        already_implied = ta.implied_facts(arg_get_fact(node), facts)
//	        implied = ta.implied_facts(
//	            forward_image(arg_get_fact(pred), action),
//	            list(set(facts) - set(already_implied)),
//	        )
//	        if len(implied) > 0:
//	            for fact in implied:
//	                ta.arg_add_facts(node, fact)
//	            self.step(node=node, facts=facts, implied=implied)
//	            return True
//	        else:
//	            return False
//
// `facts` here is a slice of *Clauses, mirroring the Python list-of-Clauses
// usage in iupdr.interactive_updr.
func RecalculateFactsFn(tc *TacticsContext, node *art.State, facts []*module.Clauses) bool {
	pred, action := ArgGetPredAction(node)
	if action == nil {
		// Python: assert action != 'join'
		panic("recalculate_facts: action is 'join'")
	}
	if pred == nil {
		// Python: assert len(preds) == 1
		panic("recalculate_facts: no single predecessor")
	}

	// Python: already_implied = ta.implied_facts(arg_get_fact(node), facts)
	alreadyImplied := tc.ImpliedFacts(ArgGetFact(node), facts)
	alreadySet := make(map[*module.Clauses]bool, len(alreadyImplied))
	for _, c := range alreadyImplied {
		alreadySet[c] = true
	}

	// Python: list(set(facts) - set(already_implied))
	remaining := make([]*module.Clauses, 0, len(facts))
	for _, f := range facts {
		if !alreadySet[f] {
			remaining = append(remaining, f)
		}
	}

	// Python: forward_image(arg_get_fact(pred), action)
	predClauses := tc.ForwardImage(ArgGetFact(pred), action)

	// Python: implied = ta.implied_facts(forward_image(...), remaining)
	implied := tc.ImpliedFacts(predClauses, remaining)

	if len(implied) > 0 {
		for _, fact := range implied {
			ArgAddFacts(node, fact)
		}
		return true
	}
	return false
}

// CustomRefineOrReverse mirrors Python tactics.py:89-125 (the @tactic
// CustomRefineOrReverse.apply body):
//
//	@tactic
//	class CustomRefineOrReverse(Tactic):
//	    def apply(self, goal, x, y, auto_remove=True):
//	        info = dict(goal=goal, x=x, y=y)
//	        if x:
//	            ta.arg_add_facts(goal.node, y)
//	            removed = []
//	            removed_str = []
//	            if auto_remove:
//	                g = goal
//	                while g is not None and ta.refuted_goal(g):
//	                    removed.append(g)
//	                    ta.remove_goal(g)
//	                    g = g.parent
//	                ...
//	        else:
//	            ta.push_goal(y)
//	            ...
//	        self.step(**info)
//	        return x
//
// `x` is a bool: True if we learned a new fact (refined), False if we
// pushed a new goal (reversed). `y` is the learned fact (a *Clauses) or
// the new goal (a *ProofGoal), depending on x.
func CustomRefineOrReverse(tc *TacticsContext, goal *proof.ProofGoal, x bool, y interface{}, autoRemove bool) bool {
	if x {
		// Python: ta.arg_add_facts(goal.node, y)
		if node, ok := goal.Node.(*art.State); ok {
			if yc, ok := y.(*module.Clauses); ok {
				ArgAddFacts(node, yc)
			} else if ye, ok := y.(lg.Expr); ok {
				ArgAddFacts(node, module.FormulaToClauses(ye, nil))
			}
		}
		if autoRemove {
			// Python: walk up via g.parent removing refuted goals.
			g := goal
			for g != nil && tc.RefutedGoal(g) {
				tc.RemoveGoal(g)
				g = g.Parent
			}
		}
	} else {
		// Python: ta.push_goal(y)
		if ng, ok := y.(*proof.ProofGoal); ok {
			tc.PushGoal(ng)
		}
	}
	return x
}

// -----------------------------------------------------------------------
// Helper functions
// -----------------------------------------------------------------------

func conjoinNodes(a, b lg.Expr) lg.Expr {
	if a == nil || lg.IsTrue(a) {
		return b
	}
	if b == nil || lg.IsTrue(b) {
		return a
	}
	and, _ := lg.NewAnd(a, b)
	return and
}
