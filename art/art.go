// Package art implements the Abstract Reachability Tree (ART) / Analysis Graph
// for Ivy. It provides the core data structures for managing reachability
// analysis states, transitions, and covering relations.
//
// This corresponds to Python's ivy_art.py.
package art

import (
	"fmt"
	"log"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/transrel"
	"github.com/glycerine/goivy/z3bridge"
)

// State represents a reachability analysis state. In the Python code this is
// provided by ivy_interp.State; since the interp package is being created
// in parallel, we define a concrete type here.
type State struct {
	ID       int
	Clauses  *clauseops.Clauses
	Update   *transrel.Update
	Domain   *module.Module
	Label    string
	Expr     Expr        // the expression that produced this state
	Pred     *State      // predecessor state (if derived from action)
	JoinOf   []*State    // predecessor states (if derived from join)
	InScope  map[string]bool
	Unders   []*State    // under-approximations (for exact states)
	Value    interface{} // assigned during BMC
	Universe interface{} // assigned during BMC
	Action   actions.Action
	ArgNode  *State // reference to a state in another graph (for copy_path)
}

// NewState creates a new state with the given domain and clauses.
func NewState(domain *module.Module, clauses *clauseops.Clauses) *State {
	return &State{
		ID:      -1,
		Domain:  domain,
		Clauses: clauses,
		InScope: make(map[string]bool),
	}
}

// StateValue returns the state as a transrel.Update triple (Modified, Clauses, Pre),
// matching Python's State.value property which returns (moded, clauses, precond).
// If Update is set, returns it directly. Otherwise constructs from Clauses with
// moded=nil and pre=FalseClauses (matching Python's default).
func (s *State) StateValue() *transrel.Update {
	if s.Update != nil {
		return s.Update
	}
	return &transrel.Update{
		Modified: nil, // None means "all modified"
		TR:       s.Clauses,
		Pre:      clauseops.FalseClauses(nil),
	}
}

// SetStateValue sets the state from a transrel.Update triple,
// matching Python's State.value setter.
func (s *State) SetStateValue(u *transrel.Update) {
	s.Update = u
	if u != nil {
		s.Clauses = u.TR
	}
}

// IsBottom reports whether this state is the bottom (empty/false) state.
func (s *State) IsBottom() bool {
	return s.Clauses != nil && s.Clauses.IsFalse()
}

// Copy creates a shallow copy of the state.
func (s *State) Copy() *State {
	cp := *s
	return &cp
}

// String returns a human-readable summary of the state.
func (s *State) String() string {
	return fmt.Sprintf("State(%d)", s.ID)
}

// -----------------------------------------------------------------------
// Expression types for recording how states were derived
// -----------------------------------------------------------------------

// Expr is the interface for expressions that record how a state was derived.
type Expr interface {
	exprMarker()
}

// ActionApp records that a state was derived by applying an action to a
// predecessor state.
type ActionApp struct {
	Rep      interface{} // either an actions.Action or a string (action name)
	Args     []*State
	Subgraph *AnalysisGraph // cached decomposition subgraph
}

func (*ActionApp) exprMarker() {}

// IsActionApp reports whether e is an ActionApp.
func IsActionApp(e Expr) bool {
	_, ok := e.(*ActionApp)
	return ok
}

// NewActionApp creates an ActionApp expression.
func NewActionApp(rep interface{}, args ...*State) *ActionApp {
	return &ActionApp{Rep: rep, Args: args}
}

// StateJoin records that a state was derived by joining two or more states.
type StateJoin struct {
	Args []*State
}

func (*StateJoin) exprMarker() {}

// IsStateJoin reports whether e is a StateJoin.
func IsStateJoin(e Expr) bool {
	_, ok := e.(*StateJoin)
	return ok
}

// NewStateJoin creates a StateJoin expression.
func NewStateJoin(args ...*State) *StateJoin {
	return &StateJoin{Args: args}
}

// -----------------------------------------------------------------------
// Counterexample
// -----------------------------------------------------------------------

// Counterexample records why a property is false. Its IsFailed method returns
// false, mirroring Python's __bool__ returning False.
type Counterexample struct {
	Clauses *clauseops.Clauses
	State   *State
	Conc    interface{} // conclusion information
	Msg     string
}

// IsFailed always returns false, indicating that a property check failed.
// In Python this was __bool__ returning False.
func (c *Counterexample) IsFailed() bool {
	return false
}

// String returns a description of the counterexample.
func (c *Counterexample) String() string {
	return fmt.Sprintf("Counterexample: %s", c.Msg)
}

// -----------------------------------------------------------------------
// SafetyResult represents the result of a safety check. It is either
// safe (nil Counterexample) or a Counterexample.
// -----------------------------------------------------------------------

// SafetyResult wraps the outcome of a safety or BMC check.
type SafetyResult struct {
	Safe bool
	Cex  *Counterexample
	Art  *AnalysisGraph // for BMC results
}

// -----------------------------------------------------------------------
// AC (ActionContext)
// -----------------------------------------------------------------------

// AC provides context for evaluating node expressions in the analysis graph.
// It wraps an AnalysisGraph and delegates to its domain (module) and actions.
type AC struct {
	Assertions map[string]lg.Node
	Actions    map[string]interface{}
	Domain     *module.Module
	AddFn      func(*State, Expr)
	NoAdd      bool
}

// NewAC creates a new AC context for the given analysis graph.
func NewAC(ag *AnalysisGraph, noAdd bool) *AC {
	return &AC{
		Assertions: make(map[string]lg.Node),
		Actions:    ag.Actions,
		Domain:     ag.Domain,
		AddFn:      ag.Add,
		NoAdd:      noAdd,
	}
}

// Get looks up an action by symbol name. Returns nil if not found.
func (ac *AC) Get(sym string) interface{} {
	v, ok := ac.Actions[sym]
	if !ok {
		return nil
	}
	return v
}

// NewState creates a new state from clauses, optionally adding it to the graph.
func (ac *AC) NewState(clauses *clauseops.Clauses, exact bool, expr Expr) *State {
	res := NewState(ac.Domain, clauses)
	if !ac.NoAdd {
		ac.AddFn(res, expr)
	}
	if exact {
		res.Unders = []*State{NewState(ac.Domain, clauses)}
	}
	return res
}

// -----------------------------------------------------------------------
// Transition
// -----------------------------------------------------------------------

// Transition records a graph edge: prestate --(action/label)--> poststate.
type Transition struct {
	Pre   *State
	Op    actions.Action // may be nil for joins
	Label string
	Post  *State
}

// -----------------------------------------------------------------------
// CoveringPair
// -----------------------------------------------------------------------

// CoveringPair records that the covered node is subsumed by the covering node.
type CoveringPair struct {
	Covered  *State
	Covering *State
}

// -----------------------------------------------------------------------
// AnalysisGraph
// -----------------------------------------------------------------------

// AnalysisGraph is the core Abstract Reachability Tree structure.
// It maintains a set of states, transitions between them, and a covering
// relation.
type AnalysisGraph struct {
	Domain        *module.Module
	States        []*State
	Transitions   []Transition
	Covering      []CoveringPair
	PVars         []lg.Node
	StateGraphs   []interface{}
	Actions       map[string]interface{}
	Predicates    map[string]interface{}
	Assertions    []*module.LabeledFormula
	Mixins        map[string][]interface{}
	Isolates      map[string]interface{}
	Exports       []interface{}
	Delegates     []interface{}
	PublicActions map[string]bool
	InitCond      *clauseops.Clauses
}

// NewAnalysisGraph creates a new AnalysisGraph backed by the given module.
// If mod is nil a fresh empty module is used.
func NewAnalysisGraph(mod *module.Module, pvars ...lg.Node) *AnalysisGraph {
	if mod == nil {
		mod = module.New()
	}
	ag := &AnalysisGraph{
		Domain:        mod,
		States:        nil,
		Transitions:   nil,
		Covering:      nil,
		PVars:         pvars,
		StateGraphs:   nil,
		Actions:       mod.Actions,
		Predicates:    mod.Predicates,
		Assertions:    mod.Assertions,
		Mixins:        mod.Mixins,
		Isolates:      mod.Isolates,
		Exports:       mod.Exports,
		Delegates:     mod.Delegates,
		PublicActions: mod.PublicActions,
	}
	return ag
}

// Context returns a new AC action context for this graph.
func (ag *AnalysisGraph) Context() *AC {
	return NewAC(ag, false)
}

// Add adds a state to the graph, assigning it an ID. If expr is non-nil,
// it records the derivation expression and creates appropriate transitions.
func (ag *AnalysisGraph) Add(state *State, expr Expr) {
	state.ID = len(ag.States)
	ag.States = append(ag.States, state)
	if expr != nil {
		state.Expr = expr
		switch e := expr.(type) {
		case *ActionApp:
			if len(e.Args) == 1 {
				var action actions.Action
				var label string
				switch rep := e.Rep.(type) {
				case actions.Action:
					action = rep
					label = LabelFromAction(rep)
				case string:
					label = rep
					if a, ok := ag.Actions[rep]; ok {
						if act, ok2 := a.(actions.Action); ok2 {
							action = act
						}
					}
				}
				ag.Transitions = append(ag.Transitions, Transition{
					Pre:   e.Args[0],
					Op:    action,
					Label: label,
					Post:  state,
				})
			}
		case *StateJoin:
			for _, js := range e.Args {
				ag.Transitions = append(ag.Transitions, Transition{
					Pre:   js,
					Op:    nil,
					Label: "join",
					Post:  state,
				})
			}
		}
	}
}

// StateCount returns the number of states in the graph.
func (ag *AnalysisGraph) StateCount() int {
	return len(ag.States)
}

// LastState returns the last state in the graph, or nil if empty.
func (ag *AnalysisGraph) LastState() *State {
	if len(ag.States) == 0 {
		return nil
	}
	return ag.States[len(ag.States)-1]
}

// Execute executes an action on a prestate (defaulting to the last state),
// computes the post-state, and adds it to the graph.
func (ag *AnalysisGraph) Execute(op actions.Action, prestate *State, abstractor Abstractor, label string) *State {
	if prestate == nil {
		prestate = ag.LastState()
	}
	if prestate == nil {
		return nil
	}
	poststate := ag.PostState(op, prestate, abstractor)
	var exprRep interface{} = op
	if label != "" {
		exprRep = label
	}
	expr := NewActionApp(exprRep, prestate)
	ag.Add(poststate, expr)
	return poststate
}

// ExecuteAction executes a named action from the graph's action map.
func (ag *AnalysisGraph) ExecuteAction(name string, prestate *State, abstractor Abstractor) *State {
	a, ok := ag.Actions[name]
	if !ok {
		return nil
	}
	action, ok := a.(actions.Action)
	if !ok {
		return nil
	}
	return ag.Execute(action, prestate, abstractor, name)
}

// PostState computes the post-state of applying an action to a pre-state.
// If the action provides an update (via the Updater interface), the
// transition relation is composed with the pre-state clauses. Otherwise
// the pre-state clauses are carried forward unchanged.
func (ag *AnalysisGraph) PostState(op actions.Action, preState *State, abstractor Abstractor) *State {
	// Compute the update (transition relation) for this action.
	// Matches Python art.py:159: s = concrete_post(op.update(domain, in_scope), pre)
	// which calls Action.update → hide_formals(bind_olds(int_update(domain, in_scope)))
	var update *transrel.Update
	if preState.Domain != nil {
		update = actions.GetUpdateForArt(op, preState.Domain, preState.InScope)
	}

	// Compose pre-state clauses with the transition relation.
	// Matches Python ivy_interp.py:202:
	//   cons = compose_state_action(state.value, axioms, update, check=context.check)
	var postClauses *clauseops.Clauses
	if update != nil && preState.Clauses != nil {
		trNode := update.TRNode()
		if trNode != nil && trNode != lg.True {
			preFmla := preState.Clauses.ToFormula()
			composed, _ := lg.NewAnd(preFmla, trNode)
			postClauses = clauseops.FormulaToClauses(composed, preState.Clauses.Annot)
		} else {
			postClauses = preState.Clauses
		}
	}

	if postClauses == nil {
		// Fallback: carry pre-state clauses forward.
		if preState.Clauses != nil {
			postClauses = preState.Clauses.Copy()
		}
	}

	s := NewState(preState.Domain, postClauses)
	s.Action = op
	s.Pred = preState
	// Store the update (transition relation) for history reconstruction.
	// Matches Python ivy_interp.py:206: res.update = update
	s.Update = update
	if abstractor != nil {
		abstractor.Abstract(s)
	}
	return s
}

// JoinStates computes the join (disjunction) of two states' clauses.
func (ag *AnalysisGraph) JoinStates(state1, state2 *State, abstractor Abstractor) *State {
	var joinedClauses *clauseops.Clauses
	if state1.Clauses != nil && state2.Clauses != nil {
		joinedClauses = clauseops.OrClausesTyped(state1.Clauses, state2.Clauses)
	} else if state1.Clauses != nil {
		joinedClauses = state1.Clauses
	} else {
		joinedClauses = state2.Clauses
	}

	joined := NewState(state1.Domain, joinedClauses)
	joined.JoinOf = []*State{state1, state2}
	joined.Label = state1.Label
	if abstractor != nil {
		abstractor.Abstract(joined)
	}
	return joined
}

// Join joins two states and adds the result to the graph.
func (ag *AnalysisGraph) Join(state1, state2 *State, abstractor Abstractor) *State {
	if state2 == nil {
		state2 = ag.LastState()
	}
	if state2 == nil {
		return nil
	}
	joined := ag.JoinStates(state1, state2, abstractor)
	ag.Add(joined, NewStateJoin(state1, state2))
	return joined
}

// Cover attempts to cover the covered node by the covering node.
// Returns true if covering succeeded (i.e., covered's clauses imply
// covering's clauses, meaning covered is a subset of covering).
func (ag *AnalysisGraph) Cover(covered, covering *State) bool {
	if covered.Clauses == nil || covering.Clauses == nil {
		return false
	}

	// Trivial cases: if covered is false (bottom), it implies anything.
	if covered.Clauses.IsFalse() {
		ag.Covering = append(ag.Covering, CoveringPair{
			Covered:  covered,
			Covering: covering,
		})
		return true
	}
	// If covering is true (top), anything implies it.
	if covering.Clauses.IsTrue() {
		ag.Covering = append(ag.Covering, CoveringPair{
			Covered:  covered,
			Covering: covering,
		})
		return true
	}

	coveredFmla := covered.Clauses.ToFormula()
	coveringFmla := covering.Clauses.ToFormula()

	t := z3bridge.NewTranslator()
	implies, err := t.Implies(coveredFmla, coveringFmla)
	if err != nil {
		log.Printf("art.Cover: z3bridge.Implies error: %v; returning false", err)
		return false
	}
	if implies {
		ag.Covering = append(ag.Covering, CoveringPair{
			Covered:  covered,
			Covering: covering,
		})
		return true
	}
	return false
}

// IsCovered reports whether the given node is covered by some other node.
func (ag *AnalysisGraph) IsCovered(node *State) bool {
	for _, c := range ag.Covering {
		if c.Covered.ID == node.ID {
			return true
		}
	}
	return false
}

// Unreachable checks whether a node is unreachable by testing if its
// clauses are unsatisfiable. If UNSAT, the node's clauses are replaced
// with false and the method returns true.
func (ag *AnalysisGraph) Unreachable(node *State) bool {
	if node.Clauses == nil {
		return false
	}
	// Already false is trivially unreachable.
	if node.Clauses.IsFalse() {
		return true
	}

	fmla := node.Clauses.ToFormula()
	t := z3bridge.NewTranslator()
	result, err := t.IsSat(fmla)
	if err != nil {
		log.Printf("art.Unreachable: z3bridge.IsSat error: %v; returning false", err)
		return false
	}
	if result == z3bridge.Unsat {
		node.Clauses = clauseops.FalseClauses(node.Clauses.Annot)
		return true
	}
	return false
}

// TransitionTo finds the transition whose post-state is the given state.
// Returns nil if not found.
func (ag *AnalysisGraph) TransitionTo(state *State) *Transition {
	for i := range ag.Transitions {
		if ag.Transitions[i].Post == state {
			return &ag.Transitions[i]
		}
	}
	return nil
}

// ReplaceState replaces the contents of poststate with those from ps,
// preserving the ID and label. It also updates covering relations.
func (ag *AnalysisGraph) ReplaceState(poststate, ps *State) {
	oldID := poststate.ID
	oldLabel := poststate.Label
	*poststate = *ps
	poststate.ID = oldID
	poststate.Label = oldLabel

	// Re-check covering for this state: remove old covering pairs
	// where this state was the covered node, then try to re-cover.
	// Corresponds to Python's replace_state which calls self.cover()
	// for each removed pair.
	var kept []CoveringPair
	var reCoverTargets []*State
	for _, c := range ag.Covering {
		if c.Covered.ID == poststate.ID {
			reCoverTargets = append(reCoverTargets, c.Covering)
		} else {
			kept = append(kept, c)
		}
	}
	ag.Covering = kept
	for _, covering := range reCoverTargets {
		ag.Cover(poststate, covering)
	}
}

// Recalculate recalculates the post-state of a transition.
func (ag *AnalysisGraph) Recalculate(t Transition, abstractor Abstractor) *State {
	var ps *State
	if t.Op == nil && t.Label == "join" {
		if t.Post.JoinOf != nil && len(t.Post.JoinOf) >= 2 {
			ps = ag.JoinStates(t.Post.JoinOf[0], t.Post.JoinOf[1], abstractor)
		} else {
			ps = t.Post
		}
	} else {
		if t.Label != "" {
			if a, ok := ag.Actions[t.Label]; ok {
				if act, ok2 := a.(actions.Action); ok2 {
					ps = ag.PostState(act, t.Pre, abstractor)
				}
			}
		}
		if ps == nil {
			ps = ag.PostState(t.Op, t.Pre, abstractor)
		}
	}
	ag.ReplaceState(t.Post, ps)
	return t.Post
}

// Delete marks a state (and its dependents) for removal, then removes them.
func (ag *AnalysisGraph) Delete(state *State) {
	state.ID = -1
	for i := state.ID + 1; i < len(ag.States); i++ {
		s := ag.States[i]
		deps := ag.dependencies(s)
		for _, d := range deps {
			if d.ID == -1 {
				s.ID = -1
				break
			}
		}
	}
	ag.RemoveMarkedStates()
}

// dependencies returns the predecessor states of s.
func (ag *AnalysisGraph) dependencies(s *State) []*State {
	if s.Pred != nil {
		return []*State{s.Pred}
	}
	if s.JoinOf != nil {
		return s.JoinOf
	}
	return nil
}

// RemoveMarkedStates removes all states with ID == -1 and updates
// transitions and covering relations accordingly.
func (ag *AnalysisGraph) RemoveMarkedStates() {
	var kept []*State
	for _, s := range ag.States {
		if s.ID != -1 {
			kept = append(kept, s)
		}
	}
	ag.States = kept
	// Re-number
	for i, s := range ag.States {
		s.ID = i
	}
	// Filter transitions
	var keptT []Transition
	for _, t := range ag.Transitions {
		if t.Pre.ID != -1 && t.Post.ID != -1 {
			keptT = append(keptT, t)
		}
	}
	ag.Transitions = keptT
	// Filter covering
	var keptC []CoveringPair
	for _, c := range ag.Covering {
		if c.Covered.ID != -1 && c.Covering.ID != -1 {
			keptC = append(keptC, c)
		}
	}
	ag.Covering = keptC
}

// UncoveredStates returns all states that are not covered and not part of
// a join expression.
func (ag *AnalysisGraph) UncoveredStates() []*State {
	covered := make(map[int]bool)
	for _, c := range ag.Covering {
		covered[c.Covered.ID] = true
	}

	joined := make(map[int]bool)
	for _, s := range ag.States {
		if s.Expr != nil {
			// If any state in this expression's args is covered,
			// mark the current state as covered too.
			if aa, ok := s.Expr.(*ActionApp); ok {
				for _, arg := range aa.Args {
					if covered[arg.ID] {
						covered[s.ID] = true
					}
				}
			}
			if sj, ok := s.Expr.(*StateJoin); ok {
				for _, arg := range sj.Args {
					joined[arg.ID] = true
				}
			}
		}
	}

	var result []*State
	for _, s := range ag.States {
		if !covered[s.ID] && !joined[s.ID] {
			result = append(result, s)
		}
	}
	return result
}

// GetHistory returns a History for bounded model checking, tracing back
// from state through its predecessors for at most bound steps.
func (ag *AnalysisGraph) GetHistory(state *State, bound *int) *transrel.History {
	// Base case: no predecessor or bound exhausted.
	if state.Pred == nil || (bound != nil && *bound <= 0) {
		// Use the state's clauses as the initial pure state.
		var formula lg.Node = lg.True
		if state.Clauses != nil {
			formula = state.Clauses.ToFormula()
		}
		u := transrel.PureState(formula)
		return transrel.NewHistory(u)
	}

	// Recursive case: get history from predecessor.
	var nextBound *int
	if bound != nil {
		nb := *bound - 1
		nextBound = &nb
	}
	h := ag.GetHistory(state.Pred, nextBound)

	// If the state has an Update, use it for the forward step.
	// Matches Python ivy_interp.py:591:
	//   history.forward_step(state.pred.domain.background_theory(...), state.update, action)
	if state.Update != nil {
		var axioms lg.Node = lg.True
		if state.Pred != nil && state.Pred.Domain != nil {
			bgTheory := state.Pred.Domain.BackgroundTheory(state.Pred.InScope)
			if bgTheory != nil {
				axioms = bgTheory.ToFormula()
			}
		}
		var actionNode lg.Node = lg.True
		h = h.ForwardStep(axioms, state.Update, actionNode)
	}

	return h
}

// CopyPath copies the path leading to state into another analysis graph.
func (ag *AnalysisGraph) CopyPath(state *State, other *AnalysisGraph, bound *int) *State {
	otherState := NewState(other.Domain, nil)
	otherState.ArgNode = state
	if state.Pred != nil && (bound == nil || *bound != 0) {
		var nextBound *int
		if bound != nil {
			nb := *bound - 1
			nextBound = &nb
		}
		pred := ag.CopyPath(state.Pred, other, nextBound)
		var rep interface{}
		if state.Expr != nil {
			if aa, ok := state.Expr.(*ActionApp); ok {
				rep = aa.Rep
			}
		}
		if rep == nil {
			rep = "unknown"
		}
		other.Add(otherState, NewActionApp(rep, pred))
	} else {
		other.Add(otherState, nil)
	}
	return otherState
}

// BMC performs bounded model checking on the graph from the given state.
// It builds a history from the state, assumes the error condition, and
// checks satisfiability. If SAT, a counterexample trace is constructed
// in otherArt. If UNSAT, returns nil.
func (ag *AnalysisGraph) BMC(state *State, errorCond lg.Node, otherArt *AnalysisGraph, bound *int) *AnalysisGraph {
	h := ag.GetHistory(state, bound)
	h = h.Assume(errorCond)

	// Check if the history's post formula (conjoined with error) is satisfiable.
	t := z3bridge.NewTranslator()
	result, err := t.IsSat(h.Post)
	if err != nil {
		log.Printf("art.BMC: z3bridge.IsSat error: %v; returning nil", err)
		return nil
	}
	if result == z3bridge.Sat {
		// Counterexample found. Copy the path into otherArt.
		if otherArt == nil {
			otherArt = NewAnalysisGraph(ag.Domain, ag.PVars...)
		}
		ag.CopyPath(state, otherArt, bound)
		return otherArt
	}
	return nil
}

// CheckSafety checks safety assertions for the given state.
// For each assertion in the graph, it checks whether the state's
// clauses imply the assertion formula. If any assertion is not implied,
// a counterexample is returned.
func (ag *AnalysisGraph) CheckSafety(state *State) *SafetyResult {
	if state.Clauses == nil || len(ag.Assertions) == 0 {
		return &SafetyResult{Safe: true}
	}

	stateFmla := state.Clauses.ToFormula()

	for _, lf := range ag.Assertions {
		if lf.Formula == nil {
			continue
		}
		t := z3bridge.NewTranslator()
		implies, err := t.Implies(stateFmla, lf.Formula)
		if err != nil {
			log.Printf("art.CheckSafety: z3bridge.Implies error: %v; treating as safe for this assertion", err)
			continue
		}
		if !implies {
			// The state does not satisfy this assertion.
			labelStr := ""
			if lf.Label != nil {
				labelStr = fmt.Sprintf(" (%s)", lf.Label)
			}
			cex := &Counterexample{
				Clauses: state.Clauses,
				State:   state,
				Msg:     fmt.Sprintf("assertion%s not satisfied", labelStr),
			}
			return &SafetyResult{Safe: false, Cex: cex}
		}
	}
	return &SafetyResult{Safe: true}
}

// CheckBoundedSafety checks safety using bounded model checking.
// For each assertion, it runs BMC with the negation of the assertion
// as the error condition. If BMC finds a counterexample, it is returned.
func (ag *AnalysisGraph) CheckBoundedSafety(state *State, bound *int) *SafetyResult {
	if state.Clauses == nil || len(ag.Assertions) == 0 {
		return &SafetyResult{Safe: true}
	}

	for _, lf := range ag.Assertions {
		if lf.Formula == nil {
			continue
		}
		// Error condition is the negation of the assertion.
		errorCond := &lg.Not{Body: lf.Formula}
		cexArt := ag.BMC(state, errorCond, nil, bound)
		if cexArt != nil {
			labelStr := ""
			if lf.Label != nil {
				labelStr = fmt.Sprintf(" (%s)", lf.Label)
			}
			cex := &Counterexample{
				Clauses: state.Clauses,
				State:   state,
				Msg:     fmt.Sprintf("bounded safety check failed for assertion%s", labelStr),
			}
			return &SafetyResult{Safe: false, Cex: cex, Art: cexArt}
		}
	}
	return &SafetyResult{Safe: true}
}

// CallAction executes an operation in a sub-graph and returns the resulting
// post-state.
func (ag *AnalysisGraph) CallAction(name string, op func(*AnalysisGraph), prestate *State) *State {
	sub := NewAnalysisGraph(ag.Domain, ag.PVars...)
	sub.Add(prestate.Copy(), nil)
	op(sub)
	poststate := sub.LastState()
	if poststate == nil {
		return nil
	}
	return poststate.Copy()
}

// DecomposeState decomposes a state into a subgraph if it has a decomposable
// expression. This is a stub.
// DecomposeState creates a new AnalysisGraph showing the decomposed
// sub-steps of the action that produced the given state.
// Matches Python ivy_art.py AnalysisGraph.decompose_state().
func (ag *AnalysisGraph) DecomposeState(state *State) *AnalysisGraph {
	if state == nil || state.Expr == nil {
		return nil
	}

	// Check if there's a cached subgraph
	if aa, ok := state.Expr.(*ActionApp); ok && aa.Subgraph != nil {
		return aa.Subgraph
	}

	// Get the action from the expression
	var action actions.Action
	if aa, ok := state.Expr.(*ActionApp); ok {
		if act, ok2 := aa.Rep.(actions.Action); ok2 {
			action = act
		}
	}
	if action == nil {
		return nil
	}

	// Decompose the action into sub-steps
	decomps := action.Decompose()
	if len(decomps) == 0 {
		return nil
	}

	// Build a new AnalysisGraph with the decomposed steps.
	// Use the first decomposition path (for Choice/If, could offer selection).
	subActions := decomps[0]
	subArt := NewAnalysisGraph(ag.Domain)

	// Create states: one per sub-action boundary (n+1 states for n actions)
	var prevState *State
	for i := 0; i <= len(subActions); i++ {
		st := NewState(ag.Domain, nil)
		if i == 0 && state.Pred != nil {
			// First state inherits pre-state clauses
			st.Clauses = state.Pred.Clauses
		} else if i == len(subActions) {
			// Last state inherits post-state clauses
			st.Clauses = state.Clauses
		}
		st.Label = fmt.Sprintf("%d", i)

		if i > 0 && prevState != nil {
			// Create transition edge from previous state
			expr := NewActionApp(subActions[i-1], prevState)
			subArt.Add(st, expr)
			subArt.Transitions = append(subArt.Transitions, Transition{
				Pre:   prevState,
				Op:    subActions[i-1],
				Label: subActions[i-1].Name(),
				Post:  st,
			})
		} else {
			subArt.Add(st, nil)
		}
		prevState = st
	}

	// Cache the subgraph on the expression
	if aa, ok := state.Expr.(*ActionApp); ok {
		aa.Subgraph = subArt
	}

	return subArt
}

// FixedpointCandidate computes a fixpoint candidate from uncovered states
// grouped by label. This is a stub.
func (ag *AnalysisGraph) FixedpointCandidate() map[string][]*State {
	fpc := make(map[string][]*State)
	for _, s := range ag.UncoveredStates() {
		if s.Label != "" {
			fpc[s.Label] = append(fpc[s.Label], s)
		}
	}
	return fpc
}

// -----------------------------------------------------------------------
// Abstractor interface
// -----------------------------------------------------------------------

// Abstractor defines the interface for state abstraction functions.
// In the Python code, this is a callable (function/lambda).
type Abstractor interface {
	Abstract(s *State)
}

// AbstractorFunc is an adapter to allow the use of ordinary functions as
// Abstractors.
type AbstractorFunc func(*State)

// Abstract calls the underlying function.
func (f AbstractorFunc) Abstract(s *State) { f(s) }

// -----------------------------------------------------------------------
// AnalysisSubgraph
// -----------------------------------------------------------------------

// AnalysisSubgraph wraps a subgraph operation.
type AnalysisSubgraph struct {
	Op    string
	Graph *AnalysisGraph
}

// NewAnalysisSubgraph creates a new AnalysisSubgraph.
func NewAnalysisSubgraph(op string, graph *AnalysisGraph) *AnalysisSubgraph {
	return &AnalysisSubgraph{Op: op, Graph: graph}
}

// String returns the operation name.
func (s *AnalysisSubgraph) String() string {
	return s.Op
}

// -----------------------------------------------------------------------
// LabelFromAction
// -----------------------------------------------------------------------

// LabelFromAction returns a label string for the given action.
// If the action has Labels set, returns the first one; otherwise returns
// a truncated string representation.
func LabelFromAction(action actions.Action) string {
	if ab, ok := action.(interface{ GetLabels() []string }); ok {
		labels := ab.GetLabels()
		if len(labels) > 0 {
			return labels[0]
		}
	}
	s := action.String()
	return truncate(s, 200)
}

// truncate shortens a string to at most maxLen characters, appending "..."
// if truncated.
func truncate(s string, maxLen int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > 4 {
		lines = lines[:4]
		s = strings.Join(lines, "\n") + "\n..."
	}
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// -----------------------------------------------------------------------
// Initialization
// -----------------------------------------------------------------------

// AddInitialState creates and adds the initial state to the analysis graph.
// The initial state is computed from the module's initial conditions and
// initializer actions.
//
// Corresponds to Python's AnalysisGraph.add_initial_state.
// AddInitialState creates and adds the initial state to the analysis graph.
// Matches Python's AnalysisGraph.add_initial_state (ivy_art.py:102-119).
//
// If the module has initializer actions, they are composed into a Sequence,
// executed from the initial conditions, and the post-state becomes the
// initial state. Otherwise, a fresh state from init_cond is added directly.
func (ag *AnalysisGraph) AddInitialState() *State {
	mod := ag.Domain

	// Start with the initial conditions from the module
	var initClauses *clauseops.Clauses
	if mod.InitCond != nil {
		initClauses = mod.InitCond
	} else {
		initClauses = clauseops.TrueClauses(nil)
	}

	s := NewState(mod, initClauses)

	if len(mod.Initializers) > 0 {
		// Python: action = Sequence(*[a for n,a in domain.initializers])
		//         action = env_action(action, 'init')
		//         s = action_app(action, s)
		//         s2 = eval_state(s) with EvalContext(check=False)
		var initActs []actions.Action
		for _, na := range mod.Initializers {
			if act, ok := na.Action.(actions.Action); ok {
				initActs = append(initActs, act)
			}
		}
		if len(initActs) > 0 {
			// Execute each initializer sequentially from the initial state.
			// Python: action = Sequence(*[a for n,a in domain.initializers])
			//         s = action_app(action, s)
			current := s
			for _, act := range initActs {
				post := ag.Execute(act, current, nil, "init")
				if post != nil {
					current = post
				}
			}
			return current
		}
	}

	// No initializers: add the initial state directly
	s2 := NewState(mod, initClauses)
	ag.Add(s2, nil)
	return s2
}

// Initialize creates the analysis graph with an initial state and optionally
// runs initializer actions. The initializer parameter, if non-nil, is called
// with the initial state to allow custom initialization.
//
// Corresponds to Python's AnalysisGraph.__init__ with initializer parameter.
func (ag *AnalysisGraph) Initialize(initializer func(*State)) {
	state := ag.AddInitialState()
	if initializer != nil {
		initializer(state)
	}
}

// -----------------------------------------------------------------------
// Option: abstract initial state
// -----------------------------------------------------------------------

// OptionAbsInit controls whether the initial state is abstracted.
var OptionAbsInit bool
