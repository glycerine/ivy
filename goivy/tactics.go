// Package tactics implements interactive refinement tactics for Ivy.
// This corresponds to Python's tactics.py and tactics_api.py.
//
// Tactics operate on the proof goal stack, refining abstract states
// through the UPDR (Universal Property-Directed Reachability) algorithm.
// They are used in the interactive concept graph UI.
package goivy

import (
	"fmt"
	"strings"
)

const tacticsCheckPrecondTrue = true
const tacticsCheckPrecondFalse = false

// Tactic is the interface for proof refinement tactics.
type LogicTactic interface {
	Apply(goal *ProofGoal) (bool, error)
	Name() string
}

// TacticStep records a single step in the tactic execution.
type TacticStep struct {
	Goal   *ProofGoal
	Active *ProofGoal
	Info   map[string]interface{}
}

// -----------------------------------------------------------------------
// Tactics API — corresponds to Python tactics_api.py
// -----------------------------------------------------------------------

// TacticsContext holds the state for interactive proof exploration.
type TacticsContext struct {
	AG    *AnalysisGraph
	Mod   *Module
	Goals *ProofGoalStack
}

// NewTacticsContext creates a new tactics context.
func NewTacticsContext(ag *AnalysisGraph, mod *Module) *TacticsContext {
	return &TacticsContext{
		AG:    ag,
		Mod:   mod,
		Goals: NewProofGoalStack(),
	}
}

// modSig returns tc.Mod.Sig if available, nil otherwise.
func (tc *TacticsContext) modSig() *Sig {
	if tc.Mod != nil {
		return tc.Mod.Sig
	}
	return nil
}

// modSolverOpts returns tc.Mod.Cfg.SolverOpts if available, nil otherwise.
func (tc *TacticsContext) modSolverOpts() *SolverOptions {
	if tc.Mod != nil && tc.Mod.Cfg != nil {
		return tc.Mod.Cfg.SolverOpts
	}
	return nil
}

// TopGoal returns the current top goal.
func (tc *TacticsContext) TopGoal() *ProofGoal {
	return tc.Goals.Top()
}

// PushGoal pushes a new goal onto the stack.
func (tc *TacticsContext) PushGoal(goal *ProofGoal) {
	tc.Goals.Push(goal)
}

// RemoveGoal removes a goal from the stack.
func (tc *TacticsContext) RemoveGoal(goal *ProofGoal) {
	tc.Goals.Remove(goal)
}

// GoalAtArgNode creates a new goal at an analysis graph node.
// Corresponds to Python's goal_at_arg_node().
func GoalAtArgNode(formula Expr, node *State) *ProofGoal {
	return &ProofGoal{Formula: formula, Node: node}
}

// BackgroundTheory returns the background theory for the module as a formula.
func (tc *TacticsContext) BackgroundTheory() Expr {
	if tc.Mod == nil {
		return True
	}
	clauses := tc.Mod.BackgroundTheory(nil)
	if clauses == nil {
		return True
	}
	return clauses.ToFormula()
}

// BackgroundTheoryClauses returns the background theory as Clauses.
// Preserves Clauses structure for and_clauses operations.
// Corresponds to Python _ivy_interp.background_theory().
func (tc *TacticsContext) BackgroundTheoryClauses() *Clauses {
	if tc.Mod == nil {
		return TrueClauses(nil)
	}
	clauses := tc.Mod.BackgroundTheory(nil)
	if clauses == nil {
		return TrueClauses(nil)
	}
	return clauses
}

// RefutedGoal checks if a goal has been refuted.
// A goal is refuted if its node's clauses conjoined with axioms imply
// the negation of the goal formula.
// Corresponds to Python tactics_api.py:refuted_goal (lines 336-341).
func (tc *TacticsContext) RefutedGoal(goal *ProofGoal) bool {
	if goal == nil || goal.Formula == nil {
		return false
	}
	// Quick check: formula is False (empty Or)
	if or, ok := goal.Formula.(*LogicOr); ok && len(or.Terms) == 0 {
		return true
	}
	node, ok := goal.Node.(*State)
	if !ok || node == nil || node.Clauses == nil {
		return false
	}

	// Python: axioms = _ivy_interp.background_theory()
	axioms := tc.BackgroundTheoryClauses()

	// Python: premise = (and_clauses(axioms, goal.node.clauses)).to_formula()
	combined := AndClausesTyped(axioms, node.Clauses)
	premise := combined.ToFormula()

	// Python: f = LogicNot(goal.formula.to_formula())
	negGoal := &LogicNot{Body: goal.Formula}

	// Python: return z3_implies(premise, f)
	slv := NewSolver(tc.Mod, tc.modSolverOpts())
	result, err := slv.Z3Implies(premise, negGoal, false)
	if err != nil {
		return false
	}
	return result
}

// ForwardImage computes the forward image of clauses through an action.
// Corresponds to Python's forward_image(pre_fact, action).
func (tc *TacticsContext) ForwardImage(preFact *Clauses, action ActionsAction) *Clauses {
	if preFact == nil || action == nil {
		return preFact
	}
	axioms := tc.BackgroundTheoryClauses()
	update := GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return preFact
	}
	return ForwardImage(preFact, axioms, update)
}

// BackwardImage computes the backward image (reverse image / weakest precondition)
// of clauses through an action.
// Corresponds to Python's backward_image(post_fact, action).
func (tc *TacticsContext) BackwardImage(postFact *Clauses, action ActionsAction) *Clauses {
	if postFact == nil || action == nil {
		return postFact
	}
	axioms := tc.BackgroundTheoryClauses()
	update := GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return postFact
	}
	return ReverseImage(postFact, axioms, update)
}

// ImpliedFacts checks which facts are implied by a premise.
// Returns the subset of factsToCheck that are implied by premise conjoined
// with background axioms.
// Corresponds to Python tactics_api.py:implied_facts (lines 310-320).
func (tc *TacticsContext) ImpliedFacts(premise *Clauses, factsToCheck []*Clauses) []*Clauses {
	if premise == nil || len(factsToCheck) == 0 {
		return nil
	}

	// Python: axioms = _ivy_interp.background_theory()
	axioms := tc.BackgroundTheoryClauses()

	// Python: premise = normalize_quantifiers((and_clauses(axioms, premise)).to_formula())
	combined := AndClausesTyped(axioms, premise)
	premFormula := NormalizeQuantifiers(combined.ToFormula())

	// Python: facts_to_check = [f.to_formula() if type(f) is Clauses else f ...]
	formulas := make([]Expr, 0, len(factsToCheck))
	indices := make([]int, 0, len(factsToCheck))
	for i, fact := range factsToCheck {
		if fact != nil {
			formulas = append(formulas, fact.ToFormula())
			indices = append(indices, i)
		}
	}

	// Python: result = z3_implies_batch(premise, facts_to_check, False)
	slv := NewSolver(tc.Mod, tc.modSolverOpts())
	results, err := slv.ImpliesBatch(premFormula, formulas, false)
	if err != nil {
		return nil
	}

	// Python: return [f for f, x in zip(facts_to_check, result) if x]
	var implied []*Clauses
	for j, isImplied := range results {
		if isImplied {
			implied = append(implied, factsToCheck[indices[j]])
		}
	}
	return implied
}

// GetDiagram returns the diagram (abstract state representation) for a goal.
// Python tactics_api.py:323-333 get_diagram(goal, weaken=False).
func (tc *TacticsContext) GetDiagram(goal *ProofGoal, weaken bool) *ProofGoal {
	if goal == nil {
		return nil
	}
	node, ok := goal.Node.(*State)
	if !ok || node == nil {
		return nil
	}

	// Python: axioms = _ivy_interp.background_theory()
	axioms := tc.BackgroundTheoryClauses()

	// Python: and_clauses(goal.node.clauses, goal.formula)
	goalClauses := FormulaToClauses(goal.Formula, nil)
	combined := AndClausesTyped(node.Clauses, goalClauses)

	// Python: is_skolem
	isSkolem := func(c *Const) bool {
		return IsSkolem(c.Name)
	}

	// Python: ivy_solver.clauses_model_to_diagram(
	//     combined, is_skolem, false_clauses(), axioms=axioms, weaken=weaken)
	slv := NewSolver(tc.Mod, tc.modSolverOpts())
	d, err := slv.ClausesModelToDiagramFull(
		combined,
		isSkolem,
		FalseClauses(nil),
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
func (tc *TacticsContext) RefineOrReverse(goal *ProofGoal) (bool, interface{}) {
	if goal == nil {
		return false, nil
	}
	node, ok := goal.Node.(*State)
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
	update := GetUpdateForArt(action, tc.Mod, nil)
	if update == nil {
		return false, nil
	}

	// Python: x = ivy_transrel.forward_interpolant(
	//     pred.clauses, action.update(...), goal.formula, axioms, None)
	goalClauses := FormulaToClauses(goal.Formula, nil)
	x := ForwardInterpolant(tc.Mod, pred.Clauses, update, goalClauses, axioms, nil)

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
func ArgGetFact(node *State) *Clauses {
	if node == nil {
		return nil
	}
	return node.Clauses
}

// ArgAddFacts adds facts (conjoins clauses) to an ARG node.
func ArgAddFacts(node *State, facts ...*Clauses) {
	if node == nil {
		return
	}
	for _, f := range facts {
		if f != nil {
			node.Clauses = AndClausesTyped(node.Clauses, f)
		}
	}
}

// ArgGetPredAction returns the predecessor and action for a node.
func ArgGetPredAction(node *State) (*State, ActionsAction) {
	if node == nil {
		return nil, nil
	}
	return node.Pred, node.Action
}

// ArgGetConjuncts returns the individual formula conjuncts at a node.
func ArgGetConjuncts(node *State) []Expr {
	if node == nil || node.Clauses == nil {
		return nil
	}
	return node.Clauses.Fmlas
}

// GetBigAction creates a nondeterministic choice over all exported actions.
// Python tactics_api.py:71-80:
//
//	def get_big_action():
//	    exported_action_names = [e.exported() for e in _ivy_ag.exports]
//	    exported_actions = [_ivy_ag.actions[k] for k in exported_action_names]
//	    result = LogicChoiceAction(*exported_actions)
//	    result.label = ' + '.join(exported_action_names)
//	    return result
func GetBigAction(ag *AnalysisGraph) ActionsAction {
	var names []string
	var branches []Expr
	for _, e := range ag.Exports {
		name := e.Exported()
		if actionName, ok := ag.actionNameFromActionLabel(name); ok {
			names = append(names, actionName)
		}
		act, ok := ag.Actions.Get2(name)
		if !ok {
			continue
		}
		if a, ok := act.(ActionsAction); ok {
			branches = append(branches, a)
		}
	}
	var result ActionsAction
	if len(branches) == 0 {
		result = NewSequence()
	} else {
		choice := NewChoiceActionOn(ag.Domain.Cfg.ActCfg, branches...)
		choice.Label = strings.Join(names, " + ")
		result = choice
	}
	ag.RegisterActionNames(result, names)
	return result
}

// -----------------------------------------------------------------------
// Concrete tactic implementations
// -----------------------------------------------------------------------

// RemoveIfRefuted removes a goal if it has been refuted.
type RemoveIfRefuted struct {
	TC *TacticsContext
}

func (t *RemoveIfRefuted) Name() string { return "RemoveIfRefuted" }

func (t *RemoveIfRefuted) Apply(goal *ProofGoal) (bool, error) {
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

func (t *RemoveGoalTactic) Apply(goal *ProofGoal) (bool, error) {
	t.TC.RemoveGoal(goal)
	return true, nil
}

// RefineOrReverseTactic attempts refinement or reversal.
type RefineOrReverseTactic struct {
	TC         *TacticsContext
	AutoRemove bool
}

func (t *RefineOrReverseTactic) Name() string { return "RefineOrReverse" }

func (t *RefineOrReverseTactic) Apply(goal *ProofGoal) (bool, error) {
	refined, result := t.TC.RefineOrReverse(goal)
	if refined {
		// Python: ta.arg_add_facts(goal.node, y)
		if newFact, ok := result.(*Clauses); ok {
			node, ok := goal.Node.(*State)
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
	if newGoal, ok := result.(*ProofGoal); ok {
		t.TC.PushGoal(newGoal)
	}
	return false, nil
}

// UPDR implements the Universal Property-Directed Reachability tactic.
// This iteratively builds a sequence of frames (over-approximations of
// reachable states) and tries to block counterexamples by refining frames.
type UPDR struct {
	TC           *TacticsContext
	MaxFrames    int // 0 means unlimited
	MaxGoalSteps int // 0 uses a bounded default derived from MaxFrames
}

func (t *UPDR) Name() string { return "UPDR" }

func proofGoalNodeClausesKey(goal *ProofGoal) string {
	if goal == nil {
		return ""
	}
	node, ok := goal.Node.(*State)
	if !ok || node == nil || node.Clauses == nil {
		return ""
	}
	return node.Clauses.String()
}

func proofGoalNodeLabel(goal *ProofGoal) string {
	if goal == nil {
		return "<nil>"
	}
	node, ok := goal.Node.(*State)
	if !ok || node == nil {
		return "<unknown>"
	}
	if node.ID >= 0 {
		return fmt.Sprintf("state %d", node.ID)
	}
	if node.Label != "" {
		return node.Label
	}
	return "unlabeled state"
}

// Apply runs the UPDR algorithm. Python tactics.py:222-292.
func (t *UPDR) Apply(goal *ProofGoal) (bool, error) {
	if t.TC.AG == nil {
		return false, fmt.Errorf("UPDR: no analysis graph")
	}

	frames := t.TC.AG.States
	if len(frames) == 0 {
		return false, fmt.Errorf("UPDR: no initial state")
	}

	// Python: bad_states = negate_clauses(get_safety_property())
	badStates := goal.Formula
	if badStates == nil {
		return false, fmt.Errorf("UPDR: no goal formula")
	}

	// Python: action = get_big_action()
	action := GetBigAction(t.TC.AG)

	// Python: ta._ivy_ag.actions[repr(action)] = action
	actionKey := fmt.Sprintf("%T@%p", action, action)
	if _, exists := t.TC.AG.Actions.Get2(actionKey); !exists {
		t.TC.AG.Actions.Set(actionKey, action)
	}

	initFrame := frames[0]
	lastFrame := frames[len(frames)-1]

	maxIter := 100
	if t.MaxFrames > 0 {
		maxIter = t.MaxFrames
	}
	maxGoalSteps := t.MaxGoalSteps
	if maxGoalSteps <= 0 {
		maxGoalSteps = maxIter * 100
		if maxGoalSteps < 100 {
			maxGoalSteps = 100
		}
	}
	goalSteps := 0

	for iter := 0; iter < maxIter; iter++ {
		// Python: check if we found an inductive invariant
		for i := 0; i < len(frames)-1; i++ {
			if t.TC.AG.Cover(frames[i+1], frames[i]) {
				return true, nil
			}
		}

		// Python: last_frame = ta.arg_add_action_node(last_frame, action, None)
		lastFrame = t.TC.ArgAddActionNode(lastFrame, action, nil)
		if lastFrame == nil {
			break
		}
		frames = t.TC.AG.States

		// Python: push_goal(goal_at_arg_node(bad_states, last_frame))
		t.TC.PushGoal(GoalAtArgNode(badStates, lastFrame))

		// Python: recalculate_facts(last_frame, ta.arg_get_conjuncts(arg_get_pred(last_frame)))
		predConjs := updrExprsToClausesList(ArgGetConjuncts(ArgGetPred(lastFrame)))
		RecalculateFactsFn(t.TC, lastFrame, predConjs)

		// Python: while not ic.interrupted:
		for t.TC.Goals.Len() > 0 {
			goalSteps++
			if goalSteps > maxGoalSteps {
				return false, fmt.Errorf("UPDR: reached goal step limit (%d) without proving or reversing the active goal", maxGoalSteps)
			}
			currentGoal := t.TC.TopGoal()
			if currentGoal == nil {
				break
			}

			// Python: if remove_if_refuted(current_goal): continue
			if t.TC.RefutedGoal(currentGoal) {
				t.TC.RemoveGoal(currentGoal)
				continue
			}

			// Python: if current_goal.node == init_frame: return False
			if goalNode, ok := currentGoal.Node.(*State); ok {
				if goalNode == initFrame {
					fmt.Println("No Invariant!")
					return false, fmt.Errorf("UPDR: no invariant found (counterexample reaches initial state)")
				}
			}

			// Python: push_diagram(current_goal, False)
			beforeLen := t.TC.Goals.Len()
			beforeTop := currentGoal
			beforeClauses := proofGoalNodeClausesKey(currentGoal)
			dgGoal := t.TC.GetDiagram(currentGoal, false)
			if dgGoal == nil {
				break
			}
			t.TC.PushGoal(dgGoal)

			// Python: dg = top_goal()
			dg := t.TC.TopGoal()

			// Python: if refine_or_reverse(dg, False):
			rorTactic := &RefineOrReverseTactic{TC: t.TC, AutoRemove: false}
			refined, rorErr := rorTactic.Apply(dg)
			if rorErr != nil {
				return false, rorErr
			}
			if refined {
				t.TC.RemoveGoal(dg)
			}
			if t.TC.Goals.Len() == beforeLen && t.TC.TopGoal() == beforeTop && proofGoalNodeClausesKey(currentGoal) == beforeClauses {
				return false, fmt.Errorf("UPDR: no progress while processing goal at %s", proofGoalNodeLabel(currentGoal))
			}
		}

		// Python: propagate phase
		// for i in range(1, len(frames)):
		//     facts_to_check = set(arg_get_conjuncts(frames[i-1])) - set(arg_get_conjuncts(frames[i]))
		//     recalculate_facts(frames[i], list(facts_to_check))
		frames = t.TC.AG.States
		for i := 1; i < len(frames); i++ {
			prev := ArgGetConjuncts(frames[i-1])
			cur := ArgGetConjuncts(frames[i])
			diff := updrSetDifference(prev, cur)
			RecalculateFactsFn(t.TC, frames[i], updrExprsToClausesList(diff))
		}
	}

	return false, fmt.Errorf("UPDR: reached iteration limit without finding invariant")
}

// CheckCover checks if one abstract state covers another.
type CheckCover struct {
	TC *TacticsContext
}

func (t *CheckCover) Name() string { return "CheckCover" }

func (t *CheckCover) Apply(goal *ProofGoal) (bool, error) {
	if t.TC.AG == nil || goal == nil {
		return false, nil
	}
	goalNode, ok := goal.Node.(*State)
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

func (t *PathReach) Apply(goal *ProofGoal) (bool, error) {
	if t.TC.AG == nil || goal == nil {
		return false, nil
	}
	goalNode, ok := goal.Node.(*State)
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

func (t *PushDiagram) Apply(goal *ProofGoal) (bool, error) {
	dg := t.TC.GetDiagram(goal, t.Weaken)
	if dg != nil {
		t.TC.PushGoal(dg)
	}
	return true, nil
}

// ExecuteAction executes an action at a node.
type ExecuteAction struct {
	TC         *TacticsContext
	Action     ActionsAction
	Abstractor Abstractor
}

func (t *ExecuteAction) Name() string { return "ExecuteAction" }

func (t *ExecuteAction) Apply(goal *ProofGoal) (bool, error) {
	if t.TC.AG == nil || goal == nil || t.Action == nil {
		return false, nil
	}
	goalNode, ok := goal.Node.(*State)
	if !ok || goalNode == nil {
		return false, nil
	}
	_, err := t.TC.AG.Execute(tacticsCheckPrecondTrue, t.Action, goalNode, t.Abstractor, "")
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
func (tc *TacticsContext) GetSafetyProperty() *Clauses {
	if tc.AG == nil || tc.AG.Domain == nil {
		return TrueClauses(nil)
	}
	conjs := tc.AG.Domain.Conjs()
	if len(conjs) == 0 {
		return TrueClauses(nil)
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
func (tc *TacticsContext) ArgAddActionNode(pre *State, action ActionsAction, abstractor Abstractor) *State {
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
	node, err := tc.AG.Execute(tacticsCheckPrecondTrue, action, pre, abstractor, label)
	if err != nil || node == nil {
		return nil
	}
	if abstractor == nil {
		// Python: node.clauses = true_clauses()
		node.Clauses = TrueClauses(nil)
	}
	return node
}

// ArgGetPred mirrors Python tactics_api.py:251-253:
//
//	def arg_get_pred(node):
//	    assert node.pred is not None
//	    return node.pred
func ArgGetPred(node *State) *State {
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
func (tc *TacticsContext) ArgIsCovered(covered, by *State) bool {
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
func CheckCoverFn(tc *TacticsContext, covered, by *State) bool {
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
func RemoveIfRefutedFn(tc *TacticsContext, goal *ProofGoal) bool {
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
func RecalculateFactsFn(tc *TacticsContext, node *State, facts []*Clauses) bool {
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
	alreadySet := make(map[*Clauses]bool, len(alreadyImplied))
	for _, c := range alreadyImplied {
		alreadySet[c] = true
	}

	// Python: list(set(facts) - set(already_implied))
	remaining := make([]*Clauses, 0, len(facts))
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
func CustomRefineOrReverse(tc *TacticsContext, goal *ProofGoal, x bool, y interface{}, autoRemove bool) bool {
	if x {
		// Python: ta.arg_add_facts(goal.node, y)
		if node, ok := goal.Node.(*State); ok {
			if yc, ok := y.(*Clauses); ok {
				ArgAddFacts(node, yc)
			} else if ye, ok := y.(Expr); ok {
				ArgAddFacts(node, FormulaToClauses(ye, nil))
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
		if ng, ok := y.(*ProofGoal); ok {
			tc.PushGoal(ng)
		}
	}
	return x
}

// -----------------------------------------------------------------------
// Helper functions
// -----------------------------------------------------------------------

// updrExprsToClausesList wraps each formula as a single-formula Clauses.
func updrExprsToClausesList(es []Expr) []*Clauses {
	if len(es) == 0 {
		return nil
	}
	out := make([]*Clauses, 0, len(es))
	for _, e := range es {
		out = append(out, NewClauses([]Expr{e}, nil, nil))
	}
	return out
}

// updrSetDifference returns elements in prev not in cur (by pointer identity).
func updrSetDifference(prev, cur []Expr) []Expr {
	curSet := make(map[Expr]bool, len(cur))
	for _, c := range cur {
		curSet[c] = true
	}
	out := make([]Expr, 0, len(prev))
	for _, p := range prev {
		if !curSet[p] {
			out = append(out, p)
		}
	}
	return out
}
