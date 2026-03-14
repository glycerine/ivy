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
	Rep  interface{} // either an actions.Action or a string (action name)
	Args []*State
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
// This is a stub; full implementation requires the interp package's
// concrete_post function.
func (ag *AnalysisGraph) PostState(op actions.Action, preState *State, abstractor Abstractor) *State {
	// In full implementation: s = concrete_post(op.update(preState.Domain, preState.InScope), preState)
	s := NewState(preState.Domain, preState.Clauses)
	s.Action = op
	s.Pred = preState
	if abstractor != nil {
		abstractor.Abstract(s)
	}
	return s
}

// JoinStates computes the join of two states. This is a stub; full
// implementation requires concrete_join from the interp package.
func (ag *AnalysisGraph) JoinStates(state1, state2 *State, abstractor Abstractor) *State {
	// Stub: in full implementation, concrete_join(state1, state2)
	joined := NewState(state1.Domain, state1.Clauses)
	joined.JoinOf = []*State{state1, state2}
	if abstractor != nil {
		abstractor.Abstract(joined)
	}
	joined.Label = state1.Label
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
// Returns true if covering succeeded.
//
// In the full implementation this checks domain.order(covered, covering).
// Here it is stubbed to always succeed.
func (ag *AnalysisGraph) Cover(covered, covering *State) bool {
	ag.Covering = append(ag.Covering, CoveringPair{
		Covered:  covered,
		Covering: covering,
	})
	return true
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

// Unreachable checks whether a node is unreachable (coverable by bottom).
// This is a stub; the full implementation checks domain.order against an
// empty state.
func (ag *AnalysisGraph) Unreachable(node *State) bool {
	// Stub: always returns false
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

	// Re-check covering for this state
	var kept []CoveringPair
	for _, c := range ag.Covering {
		if c.Covered.ID == poststate.ID {
			// Try to re-cover (stubbed: always re-add)
			kept = append(kept, CoveringPair{Covered: poststate, Covering: c.Covering})
		} else {
			kept = append(kept, c)
		}
	}
	ag.Covering = kept
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
// from state for at most bound steps. This is a stub.
func (ag *AnalysisGraph) GetHistory(state *State, bound *int) *transrel.History {
	// Stub: return a simple history from the state's update
	u := transrel.TopState()
	return transrel.NewHistory(u)
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
// This is a stub; full implementation requires history_satisfy from interp.
func (ag *AnalysisGraph) BMC(state *State, errorCond lg.Node, otherArt *AnalysisGraph, bound *int) *AnalysisGraph {
	// Stub: always returns nil (no counterexample found)
	return nil
}

// CheckSafety checks safety assertions for the given state.
// Returns a SafetyResult indicating whether the state is safe.
// This is a stub; full implementation requires check_state_assertion.
func (ag *AnalysisGraph) CheckSafety(state *State) *SafetyResult {
	// Stub: always returns safe
	return &SafetyResult{Safe: true}
}

// CheckBoundedSafety checks safety using bounded model checking.
// This is a stub.
func (ag *AnalysisGraph) CheckBoundedSafety(state *State, bound *int) *SafetyResult {
	// Stub: always returns safe
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
func (ag *AnalysisGraph) DecomposeState(state *State) *AnalysisGraph {
	return nil
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
// Option: abstract initial state
// -----------------------------------------------------------------------

// OptionAbsInit controls whether the initial state is abstracted.
var OptionAbsInit bool
