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
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	"github.com/glycerine/goivy/interp"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/logicparser"
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
	Expr     Expr     // the expression that produced this state
	Pred     *State   // predecessor state (if derived from action)
	JoinOf   []*State // predecessor states (if derived from join)
	InScope  map[string]bool
	Unders   []*State    // under-approximations (for exact states)
	Value    *transrel.Update // assigned during BMC
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
	Assertions map[string]lg.Expr
	Actions    map[string]module.Action
	Domain     *module.Module
	AddFn      func(*State, Expr)
	NoAdd      bool
}

// NewAC creates a new AC context for the given analysis graph.
func NewAC(ag *AnalysisGraph, noAdd bool) *AC {
	return &AC{
		Assertions: make(map[string]lg.Expr),
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
	PVars         []lg.Expr
	StateGraphs   []interface{}
	Actions       map[string]module.Action
	Predicates    map[string]ast.Node
	Assertions    []*ast.LabeledFormula
	Mixins        map[string][]module.MixinDef
	Isolates      map[string]*ast.IsolateDef
	Exports       []module.Exporter
	Delegates     []module.Delegator
	PublicActions map[string]bool
	InitCond      *clauseops.Clauses
}

// NewAnalysisGraph creates a new AnalysisGraph backed by the given module.
// If mod is nil a fresh empty module is used.
func NewAnalysisGraph(mod *module.Module, pvars ...lg.Expr) *AnalysisGraph {
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
			// Python ivy_art.py:250-264: asserts len(expr.args)==1, isinstance(expr.args[0], State)
			if len(e.Args) != 1 {
				panic(fmt.Sprintf("art.Add: ActionApp must have exactly 1 arg, got %d", len(e.Args)))
			}
			if e.Args[0] == nil {
				panic("art.Add: ActionApp arg[0] must be a State, got nil")
			}
			var action actions.Action
			var label string
			switch rep := e.Rep.(type) {
			case actions.Action:
				action = rep
				label = LabelFromAction(rep)
			case string:
				label = rep
				a, ok := ag.Actions[rep]
				if !ok {
					panic(fmt.Sprintf("art.Add: action %q not found in actions", rep))
				}
				if act, ok2 := a.(actions.Action); ok2 {
					action = act
				}
			}
			ag.Transitions = append(ag.Transitions, Transition{
				Pre:   e.Args[0],
				Op:    action,
				Label: label,
				Post:  state,
			})
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
// Uses transrel.ForwardImage to compute the proper forward image with havocing,
// matching Python ivy_art.py:158-163:
//
//	s = concrete_post(op.update(pre_state.domain, pre_state.in_scope), pre_state)
//	s.action = op
func (ag *AnalysisGraph) PostState(op actions.Action, preState *State, abstractor Abstractor) *State {
	// Compute the update (transition relation) for this action.
	var update *transrel.Update
	if preState.Domain != nil {
		update = actions.GetUpdateForArt(op, preState.Domain, preState.InScope)
	}

	var postClauses *clauseops.Clauses
	if update != nil && preState.Clauses != nil {
		// Get background theory (axioms) for forward image computation.
		var axiomsFmla lg.Expr = lg.True
		if preState.Domain != nil {
			bg := preState.Domain.BackgroundTheory(preState.InScope)
			if bg != nil {
				axiomsFmla = bg.ToFormula()
			}
		}
		preFmla := preState.Clauses.ToFormula()

		// Compute forward image: the proper concrete post operation.
		// This matches Python's concrete_post → compose_state_action → forward_image.
		postFmla := transrel.ForwardImage(preFmla, axiomsFmla, update)
		postClauses = clauseops.FormulaToClauses(postFmla, preState.Clauses.Annot)
	}

	if postClauses == nil {
		if preState.Clauses != nil {
			postClauses = preState.Clauses.Copy()
		}
	}

	s := NewState(preState.Domain, postClauses)
	s.Action = op
	s.Pred = preState
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
// Uses the module's ordering relation (interp.ModuleOrder).
// Python ivy_art.py:217-218: covered_node.domain.order(covered_node, covering_node)
func (ag *AnalysisGraph) Cover(covered, covering *State) bool {
	if covered.Clauses == nil || covering.Clauses == nil {
		return false
	}

	coveredInterp := ArtToInterpState(covered)
	coveringInterp := ArtToInterpState(covering)

	ok, _ := interp.ModuleOrder(coveredInterp, coveringInterp)
	if ok {
		fmt.Printf("Covering succeeded: %d %d\n", covered.ID, covering.ID)
		ag.Covering = append(ag.Covering, CoveringPair{
			Covered:  covered,
			Covering: covering,
		})
		return true
	}
	fmt.Println("Covering failed")
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
	defer t.Close()

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

// ShowCore parses a clause string, gets the unsat core from the state, and
// prints it. Python ivy_art.py:239-243.
func (ag *AnalysisGraph) ShowCore(clauseStr string, state *State) {
	clause, err := logicparser.ToFormula(clauseStr)
	if err != nil {
		fmt.Printf("ShowCore: parse error: %v\n", err)
		return
	}
	if state == nil {
		state = ag.LastState()
	}
	if state == nil {
		fmt.Println("ShowCore: no state")
		return
	}
	interpState := ArtToInterpState(state)
	core := interp.GetCore(interpState, clause.(lg.Expr))
	fmt.Println(core)
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
// Python ivy_art.py:291-298.
func (ag *AnalysisGraph) Delete(state *State) {
	savedID := state.ID
	state.ID = -1
	for i := savedID + 1; i < len(ag.States); i++ {
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
		var formula lg.Expr = lg.True
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
		var axioms lg.Expr = lg.True
		if state.Pred != nil && state.Pred.Domain != nil {
			bgTheory := state.Pred.Domain.BackgroundTheory(state.Pred.InScope)
			if bgTheory != nil {
				axioms = bgTheory.ToFormula()
			}
		}
		var actionNode lg.Expr = lg.True
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
// Uses interp.HistorySatisfy to check satisfiability and extract the
// concrete path and universe. Python ivy_art.py:331-345.
func (ag *AnalysisGraph) BMC(state *State, errorCond lg.Expr, otherArt *AnalysisGraph, bound *int) *AnalysisGraph {
	h := ag.GetHistory(state, bound)
	h = h.Assume(errorCond)

	// Use HistorySatisfy to check and extract path + universe.
	interpState := ArtToInterpState(state)
	bmcRes := interp.HistorySatisfy(h, interpState)
	if bmcRes == nil {
		return nil
	}

	// Counterexample found. Copy the path into otherArt.
	if otherArt == nil {
		otherArt = NewAnalysisGraph(ag.Domain, ag.PVars...)
		otherArt.Actions = ag.Actions
	}
	ag.CopyPath(state, otherArt, bound)

	// Python: for state,value in zip(other_art.states[-len(path):], path):
	//           state.value = value; state.universe = universe
	path := bmcRes.Path
	pathLen := len(path)
	states := otherArt.States
	startIdx := len(states) - pathLen
	if startIdx < 0 {
		startIdx = 0
	}
	for i, s := range states[startIdx:] {
		if i < pathLen {
			s.Value = path[i]
			s.Universe = bmcRes.Universes
		}
	}
	return otherArt
}

// CheckSafety checks safety assertions for the given state.
// Uses interp.CheckStateAssertion for each assertion.
// If the state has an expr, evaluates it under AC(no_add=True) and catches
// IvyActionFailedError.
// Python ivy_art.py:372-383.
func (ag *AnalysisGraph) CheckSafety(state *State) *SafetyResult {
	if len(ag.Assertions) == 0 {
		return &SafetyResult{Safe: true}
	}

	interpState := ArtToInterpState(state)

	// Python: for asn in self.assertions: cex = check_state_assertion(state, asn)
	for _, asn := range ag.Assertions {
		cex := interp.CheckStateAssertion(interpState, asn)
		if !cex {
			return &SafetyResult{Safe: false, Cex: &Counterexample{
				Clauses: state.Clauses,
				State:   state,
				Msg:     "assertion failure",
			}}
		}
	}

	// Python: if hasattr(state,'expr') and state.expr != None:
	//   with AC(self, no_add=True): eval_state(state.expr)
	//   except IvyActionFailedError as err: return Counterexample(...)
	if state.Expr != nil {
		if aa, ok := state.Expr.(*ActionApp); ok {
			ac := NewAC(ag, true)
			_ = ac
			// Construct an ast.Node from the ActionApp for eval_state
			var actionName string
			switch rep := aa.Rep.(type) {
			case string:
				actionName = rep
			case actions.Action:
				actionName = rep.Name()
			}
			if actionName != "" && len(aa.Args) > 0 {
				interpPre := ArtToInterpState(aa.Args[0])
				exprNode := interp.ActionApp(ag.Domain.Cfg.AstCfg, actionName, interp.WrapState(interpPre))
				_, err := interp.EvalState(exprNode, ag.Domain)
				if err != nil {
					if afe, ok := err.(*interp.IvyActionFailedError); ok {
						errState := InterpToArtState(afe.ErrorState)
						var predState *State
						if afe.State != nil {
							predState = InterpToArtState(afe.State)
						}
						return &SafetyResult{Safe: false, Cex: &Counterexample{
							Clauses: errState.Clauses,
							State:   predState,
							Conc:    afe.Conc,
							Msg:     afe.Error(),
						}}
					}
				}
			}
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
		errorCond := &lg.Not{Body: lf.Formula.(lg.Expr)}
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

// DecomposeState creates a new AnalysisGraph showing the decomposed
// sub-steps of the action that produced the given state.
// Python ivy_art.py:392-401.
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

// DecomposeEdge decomposes a transition by decomposing its poststate.
// Python ivy_art.py:408-410.
func (ag *AnalysisGraph) DecomposeEdge(t Transition) *AnalysisGraph {
	return ag.DecomposeState(t.Post)
}

// MakeConcreteTrace is a stub. The Python source (ivy_art.py:404-406) is
// also a stub: the body is just "# TODO\nreturn".
func (ag *AnalysisGraph) MakeConcreteTrace(state *State, conc interface{}) {
	return
}

// StateActions returns the state equations for expanding the given state.
// If the state has a label, returns predicate-based equations.
// Otherwise returns equations for each public action applied to the state.
// Python ivy_art.py:134-137.
func (ag *AnalysisGraph) StateActions(state *State) []*ast.Definition {
	if state.Label != "" {
		// Labeled state: use predicates
		interpState := ArtToInterpState(state)
		var result []*ast.Definition
		for post, e := range ag.Predicates {
			if e == nil {
				continue
			}
			exprs := interp.EvalStateActions(e, interpState)
			for _, expr := range exprs {
				acfg := ag.Domain.Cfg.AstCfg
				lhs := acfg.NewAtom(post) // post label as LHS
				result = append(result, acfg.NewDefinition(lhs, expr))
			}
		}
		return result
	}
	// Unlabeled state: apply each public action
	var result []*ast.Definition
	for actionName := range ag.Actions {
		if !ag.PublicActions[actionName] {
			continue
		}
		interpState := ArtToInterpState(state)
		acfg := ag.Domain.Cfg.AstCfg
		app := interp.ActionApp(acfg, actionName, interp.WrapState(interpState))
		result = append(result, acfg.NewDefinition(nil, app))
	}
	return result
}

// DoStateAction evaluates a state equation and returns the resulting state.
// Python ivy_art.py:139-145.
func (ag *AnalysisGraph) DoStateAction(equation *ast.Definition, abstractor Abstractor) *State {
	ac := ag.Context()
	_ = ac
	rhs := equation.Rhs
	if rhs == nil {
		return nil
	}
	is, err := interp.EvalState(rhs, ag.Domain)
	if err != nil {
		log.Printf("art.DoStateAction: EvalState error: %v", err)
		return nil
	}
	s := InterpToArtState(is)
	if equation.Args() != nil && len(equation.Args()) > 0 {
		if lhs := equation.Args()[0]; lhs != nil {
			if atom, ok := lhs.(*ast.Atom); ok {
				s.Label = atom.Rep
			}
		}
	}
	if abstractor != nil {
		abstractor.Abstract(s)
	}
	return s
}

// RecalculateState recalculates a state from its predecessor or join sources.
// Python ivy_art.py:176-184.
func (ag *AnalysisGraph) RecalculateState(state *State, abstractor Abstractor) {
	if state.Pred != nil && state.Update != nil {
		// Has predecessor: recompute via concrete_post (ForwardImage)
		var axiomsFmla lg.Expr = lg.True
		if state.Pred.Domain != nil {
			bg := state.Pred.Domain.BackgroundTheory(state.Pred.InScope)
			if bg != nil {
				axiomsFmla = bg.ToFormula()
			}
		}
		preFmla := state.Pred.Clauses.ToFormula()
		postFmla := transrel.ForwardImage(preFmla, axiomsFmla, state.Update)
		postClauses := clauseops.FormulaToClauses(postFmla, state.Pred.Clauses.Annot)
		ps := NewState(state.Domain, postClauses)
		if abstractor != nil {
			abstractor.Abstract(ps)
		}
		ag.ReplaceState(state, ps)
	} else if state.JoinOf != nil && len(state.JoinOf) >= 2 {
		// Has join sources: recompute via join
		ps := ag.JoinStates(state.JoinOf[0], state.JoinOf[1], abstractor)
		ag.ReplaceState(state, ps)
	}
}

// StateExtensions yields state equations for extending the given state
// that are not yet covered by the fixpoint candidate.
// Python ivy_art.py:434-440.
func (ag *AnalysisGraph) StateExtensions(state *State, joinFn func(*State, *State) *State) []*ast.Definition {
	sas := ag.StateActions(state)
	fpc := ag.FixedpointCandidate(joinFn)
	var result []*ast.Definition
	for _, equation := range sas {
		// Get the label from the LHS
		label := ""
		if equation.Args() != nil && len(equation.Args()) > 0 {
			if lhs := equation.Args()[0]; lhs != nil {
				if atom, ok := lhs.(*ast.Atom); ok {
					label = atom.Rep
				}
			}
		}
		fpcState := ag.FixedpointCandidateBottomDefault(fpc, label)
		interpFpc := ArtToInterpState(fpcState)

		// Check if the equation's RHS is already covered
		rhs := equation.Rhs
		ok, _ := interp.EvalStateOrder(rhs, interp.WrapState(interpFpc), ag.Domain)
		if !ok {
			result = append(result, equation)
		}
	}
	return result
}

// FixedpointCandidate computes a fixpoint candidate from uncovered states
// grouped by label, joining each group into a single state.
// Returns a map from label to joined state (bottom state for missing labels).
// Python ivy_art.py:426-432.
func (ag *AnalysisGraph) FixedpointCandidate(joinFn func(*State, *State) *State) map[string]*State {
	if joinFn == nil {
		joinFn = func(s1, s2 *State) *State {
			return ag.JoinStates(s1, s2, nil)
		}
	}
	groups := make(map[string][]*State)
	for _, s := range ag.UncoveredStates() {
		if s.Label != "" {
			groups[s.Label] = append(groups[s.Label], s)
		}
	}
	fmt.Printf("fpc = %v\n", groups)
	result := make(map[string]*State)
	for label, states := range groups {
		joined := states[0]
		for i := 1; i < len(states); i++ {
			joined = joinFn(joined, states[i])
		}
		result[label] = joined
	}
	return result
}

// FixedpointCandidateBottomDefault returns the fixpoint candidate for a label,
// defaulting to a bottom state if the label is not found.
// This mirrors Python's defaultdict(bottom_state, ...) behavior.
func (ag *AnalysisGraph) FixedpointCandidateBottomDefault(fpc map[string]*State, label string) *State {
	if s, ok := fpc[label]; ok {
		return s
	}
	return NewState(ag.Domain, clauseops.FalseClauses(nil))
}

// -----------------------------------------------------------------------
// GUI/Visualization methods (Batch 4)
// -----------------------------------------------------------------------

// ConceptGraph creates a concept graph for the given state.
// Python ivy_art.py:307-316.
func (ag *AnalysisGraph) ConceptGraph(state *State, standardGraph func(*State) ConceptGraphView, clauses *clauseops.Clauses) ConceptGraphView {
	if clauses == nil {
		clauses = state.Clauses
	}
	bg := ag.Domain.BackgroundTheory(state.InScope)
	sg := standardGraph(state)
	sg.SetGraphState(clauseops.AndClausesTyped(clauses, bg))
	sg.SetGraphConcrete(nil)
	return sg
}

// AsCyElements converts this AnalysisGraph into Cytoscape elements for
// browser rendering. The dotLayout callback, if provided, is applied to
// the result.
// Python ivy_art.py:442-443: return dot_layout(render_rg(self), edge_labels=True)
func (ag *AnalysisGraph) AsCyElements(dotLayout func(*CyElements) *CyElements) *CyElements {
	argState := &AnalysisGraphState{}
	for _, s := range ag.States {
		info := fmt.Sprintf("%d", s.ID)
		if s.Clauses != nil {
			info = fmt.Sprintf("%d (%d clauses)", s.ID, len(s.Clauses.Fmlas))
		}
		argState.States = append(argState.States, ARGNode{
			ID:       s.ID,
			Label:    fmt.Sprintf("%d", s.ID),
			IsBottom: s.IsBottom(),
			Info:     info,
		})
	}
	for _, t := range ag.Transitions {
		preID, postID := -1, -1
		if t.Pre != nil {
			preID = t.Pre.ID
		}
		if t.Post != nil {
			postID = t.Post.ID
		}
		label := t.Label
		if label == "" {
			label = "(unlabeled)"
		}
		// Python render_rg: label = label.replace('}',']-').replace('{','-[')
		label = strings.ReplaceAll(label, "}", "]-")
		label = strings.ReplaceAll(label, "{", "-[")
		label = strings.ReplaceAll(label, "\n", "\\l") + "\\l"
		argState.Transitions = append(argState.Transitions, ARGTransition{
			SourceID: preID,
			TargetID: postID,
			Label:    label,
			IsJoin:   t.Label == "join",
		})
	}
	for _, c := range ag.Covering {
		coveredID, coveringID := -1, -1
		if c.Covered != nil {
			coveredID = c.Covered.ID
		}
		if c.Covering != nil {
			coveringID = c.Covering.ID
		}
		argState.Covering = append(argState.Covering, ARGCover{
			CoveredID:  coveredID,
			CoveringID: coveringID,
		})
	}
	g := RenderARG(argState)
	if dotLayout != nil {
		g = dotLayout(g)
	}
	return g
}

// CheckConstraints is a stub — not present in the Python ivy_art.py.
// It exists as a placeholder for constraint checking in the ARG.
func (ag *AnalysisGraph) CheckConstraints() bool {
	return true
}

// StratifyGoals is a stub — not present in the Python ivy_art.py.
// It exists as a placeholder for goal stratification in the ARG.
func (ag *AnalysisGraph) StratifyGoals() []interface{} {
	return nil
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
// AddInitialState creates and adds the initial state to the analysis graph.
// Matches Python's AnalysisGraph.add_initial_state (ivy_art.py:102-119).
//
// If the module has initializer actions, they are composed into a Sequence,
// wrapped in an EnvAction, and evaluated as a single action_app with
// EvalContext(check=False). Otherwise, a fresh state from init_cond is added.
func (ag *AnalysisGraph) AddInitialState(ic *clauseops.Clauses, abstractor Abstractor) *State {
	mod := ag.Domain

	if ic == nil {
		if mod.InitCond != nil {
			ic = mod.InitCond
		} else {
			ic = clauseops.TrueClauses(nil)
		}
	}

	s := NewState(mod, ic)

	if len(mod.Initializers) > 0 {
		// Python ivy_art.py:106-114:
		//   action = Sequence(*[a for n,a in domain.initializers])
		//   action = env_action(action, 'init')
		//   s = action_app(action, s)
		//   with AC(self, no_add=True):
		//       with EvalContext(check=False):
		//           s2 = eval_state(s)
		//   s2.expr = s
		//   self.add(s2)
		var seqChildren []lg.Expr
		for _, na := range mod.Initializers {
			if na.Action != nil {
				seqChildren = append(seqChildren, na.Action)
			}
		}
		if len(seqChildren) > 0 {
			// Step 1: Sequence(*initializers)
			seq := actions.NewSequence(seqChildren...)

			// Step 2: env_action(action, 'init') — wrap in EnvAction with label
			retAct := &actions.ReturnAction{}
			innerSeq := actions.NewSequence(seq, retAct)
			env := actions.NewEnvAction(innerSeq)
			env.SetLabels([]string{"init"})

			// Step 3: action_app(action, s) — build expression
			actionAppExpr := NewActionApp(env, s)

			// Step 4: eval_state(s) under AC(no_add=True) + EvalContext(check=False)
			interpState := ArtToInterpState(s)
			ec := interp.NewEvalContext(false)
			ec.Enter()
			s2interp, err := interp.ApplyAction(nil, "init", env, interpState)
			ec.Exit()

			if err != nil {
				log.Printf("art.AddInitialState: ApplyAction error: %v", err)
				// Fall through to no-initializer path
			} else {
				s2 := InterpToArtState(s2interp)
				s2.Expr = actionAppExpr
				ag.Add(s2, nil)
				if abstractor != nil {
					abstractor.Abstract(s2)
				}
				return s2
			}
		}
	}

	// No initializers: add the initial state directly
	// Python: s2 = domain.new_state(ic); self.add(s2, s)
	s2 := NewState(mod, ic)
	ag.Add(s2, nil)
	if abstractor != nil {
		abstractor.Abstract(s2)
	}
	return s2
}

// Initialize creates the analysis graph with initial states.
// Matches Python's AnalysisGraph.initialize (ivy_art.py:121-132).
//
// If predicates exist, evaluates each as state facts with label.
// Otherwise calls AddInitialState.
func (ag *AnalysisGraph) Initialize(abstractor Abstractor) {
	ac := ag.Context()
	_ = ac // AC context

	if len(ag.Predicates) > 0 {
		// Python: if self.predicates:
		//   if not im.module.init_cond.is_true(): raise IvyError
		//   for n,p in self.predicates.items():
		//     s = eval_state_facts(p); if s: s.label = n
		if ag.Domain.InitCond != nil && !ag.Domain.InitCond.IsTrue() {
			panic("init and state declarations are not compatible")
		}
		for name, p := range ag.Predicates {
			if p == nil {
				continue
			}
			is, err := interp.EvalStateFacts(p, ag.Domain)
			if err != nil || is == nil {
				continue
			}
			artState := InterpToArtState(is)
			artState.Label = name
			ag.Add(artState, nil)
		}
	} else {
		ag.AddInitialState(nil, abstractor)
	}
}

// -----------------------------------------------------------------------
// Option: abstract initial state
// -----------------------------------------------------------------------

// OptionAbsInit controls whether the initial state is abstracted.
var OptionAbsInit bool

// -----------------------------------------------------------------------
// State type adapters: art.State <-> interp.State
// -----------------------------------------------------------------------

// ArtToInterpState converts an art.State to an interp.State for calling
// interp functions. This adapter copies the shared fields.
func ArtToInterpState(s *State) *interp.State {
	if s == nil {
		return nil
	}
	sv := interp.NewStateValue(nil, s.Clauses, clauseops.FalseClauses(nil))
	is := interp.NewState(s.Domain, sv, nil, s.Label)
	is.InScope = s.InScope
	is.Action = s.Action
	if s.Update != nil {
		is.SetUpdate(s.Update)
	}
	if s.Pred != nil {
		is.SetPred(ArtToInterpState(s.Pred))
	}
	return is
}

// InterpToArtState converts an interp.State back to an art.State.
func InterpToArtState(is *interp.State) *State {
	if is == nil {
		return nil
	}
	s := NewState(is.Domain, is.Clauses)
	s.Label = is.Label
	s.InScope = is.InScope
	s.Action = is.Action
	if is.Update() != nil {
		s.Update = is.Update()
	}
	if is.Pred() != nil {
		s.Pred = InterpToArtState(is.Pred())
	}
	return s
}
