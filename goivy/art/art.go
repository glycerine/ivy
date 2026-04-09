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

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/interp"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/logicparser"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

const checkPrecondFalse = false
const checkPrecondTrue = true

// State represents a reachability analysis state. In the Python code this is
// provided by ivy_interp.State; since the interp package is being created
// in parallel, we define a concrete type here.
type State struct {
	ID         int
	Clauses    *module.Clauses
	Update     *actions.Update
	Domain     *module.Module
	Label      string
	Prov       Provenance // provenance: the expression that produced this state
	Pred       *State     // predecessor state (if derived from action)
	JoinOf     []*State   // predecessor states (if derived from join)
	InScope    map[string]bool
	Unders     []*State        // under-approximations (for exact states)
	Value      *actions.Update // assigned during BMC
	Universe   interface{}     // assigned during BMC
	Action     actions.Action
	ActionName string // name of the action (matches interp.State.ActionName)
	ArgNode    *State // reference to a state in another graph (for copy_path)
}

// NewState creates a new state with the given domain and clauses.
func NewState(domain *module.Module, clauses *module.Clauses) *State {
	return &State{
		ID:      -1,
		Domain:  domain,
		Clauses: clauses,
		InScope: make(map[string]bool),
	}
}

// StateValue returns the state as a actions.Update triple (Modified, Clauses, Pre),
// matching Python's State.value property which returns (moded, clauses, precond).
// If Update is set, returns it directly. Otherwise constructs from Clauses with
// moded=nil and pre=FalseClauses (matching Python's default).
func (s *State) StateValue() *actions.Update {
	if s.Update != nil {
		return s.Update
	}
	return &actions.Update{
		Modified: nil, // None means "all modified"
		TR:       s.Clauses,
		Pre:      module.FalseClauses(nil),
	}
}

// SetStateValue sets the state from a actions.Update triple,
// matching Python's State.value setter.
func (s *State) SetStateValue(u *actions.Update) {
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
// Provenance types for recording how states were derived
// -----------------------------------------------------------------------

// Provenance records how a state was derived (its provenance).
// Named to avoid confusion with lg.Expr, ast.Node, and interp.State.Expr.
type Provenance interface {
	provenanceMarker()
}

// ActionApp records that a state was derived by applying an action to a
// predecessor state.
type ActionApp struct {
	Rep      interface{} // either an actions.Action or a string (action name)
	Args     []*State
	Subgraph *AnalysisGraph // cached decomposition subgraph
}

func (*ActionApp) provenanceMarker() {}

// IsActionApp reports whether p is an ActionApp.
func IsActionApp(p Provenance) bool {
	_, ok := p.(*ActionApp)
	return ok
}

// NewActionApp creates an ActionApp provenance.
func NewActionApp(rep interface{}, args ...*State) *ActionApp {
	return &ActionApp{Rep: rep, Args: args}
}

// StateJoin records that a state was derived by joining two or more states.
type StateJoin struct {
	Args []*State
}

func (*StateJoin) provenanceMarker() {}

// IsStateJoin reports whether p is a StateJoin.
func IsStateJoin(p Provenance) bool {
	_, ok := p.(*StateJoin)
	return ok
}

// NewStateJoin creates a StateJoin provenance.
func NewStateJoin(args ...*State) *StateJoin {
	return &StateJoin{Args: args}
}

// -----------------------------------------------------------------------
// Counterexample
// -----------------------------------------------------------------------

// Counterexample records why a property is false. Its IsFailed method returns
// false, mirroring Python's __bool__ returning False.
type Counterexample struct {
	Clauses *module.Clauses
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
	Actions    *iu.InsMap[string, module.Action]
	Domain     *module.Module
	AddFn      func(*State, Provenance)
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
	v, ok := ac.Actions.Get2(sym)
	if !ok {
		return nil
	}
	return v
}

// NewState creates a new state from clauses, optionally adding it to the graph.
func (ac *AC) NewState(clauses *module.Clauses, exact bool, prov Provenance) *State {
	res := NewState(ac.Domain, clauses)
	if !ac.NoAdd {
		ac.AddFn(res, prov)
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
	Actions       *iu.InsMap[string, module.Action]
	Predicates    map[string]ast.Node
	Assertions    []*ast.LabeledFormula
	Mixins        *iu.InsMap[string, []module.MixinDef]
	Isolates      map[string]*ast.IsolateDef
	Exports       []module.Exporter
	Delegates     []module.Delegator
	PublicActions *iu.InsMap[string, bool]
	InitCond      *module.Clauses
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
func (ag *AnalysisGraph) Add(state *State, prov Provenance) {
	state.ID = len(ag.States)
	ag.States = append(ag.States, state)
	if prov != nil {
		state.Prov = prov
		switch e := prov.(type) {
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
				a, ok := ag.Actions.Get2(rep)
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
// Returns an error if the action's precondition fails (when checkPrecond is true).
func (ag *AnalysisGraph) Execute(checkPrecond bool, op actions.Action, prestate *State, abstractor Abstractor, label string) (*State, error) {
	xtracer.Trace("art.Execute ENTER")
	if prestate == nil {
		prestate = ag.LastState()
	}
	if prestate == nil {
		return nil, nil
	}
	poststate, err := ag.PostState(checkPrecond, op, prestate, abstractor)
	if err != nil {
		return nil, err
	}
	var exprRep interface{} = op
	if label != "" {
		exprRep = label
	}
	expr := NewActionApp(exprRep, prestate)
	ag.Add(poststate, expr)
	return poststate, nil
}

// ExecuteAction executes a named action from the graph's action map.
func (ag *AnalysisGraph) ExecuteAction(checkPrecond bool, name string, prestate *State, abstractor Abstractor) (*State, error) {
	a, ok := ag.Actions.Get2(name)
	if !ok {
		return nil, fmt.Errorf("art.ExecuteAction: action %q not found", name)
	}
	action, ok := a.(actions.Action)
	if !ok {
		return nil, fmt.Errorf("art.ExecuteAction: %q is not an Action", name)
	}
	return ag.Execute(checkPrecond, action, prestate, abstractor, name)
}

// PostState computes the post-state of applying an action to a pre-state.
// Delegates to interp.ApplyAction which calls interp.ConcretePost, matching
// Python ivy_art.py:158-163:
//
//	s = concrete_post(op.update(pre_state.domain, pre_state.in_scope), pre_state)
//	s.action = op
//
// When checkPrecond is true the action's precondition is checked; if it is
// violated an error wrapping interp.IvyActionFailedError is returned.
func (ag *AnalysisGraph) PostState(checkPrecond bool, op actions.Action, preState *State, abstractor Abstractor) (*State, error) {
	xtracer.Trace("art.PostState ENTER opName=%s", actions.ActionTypeName(op))
	xtracer.Trace("art.PostState calling GetUpdate type=%s", actions.ActionTypeName(op))
	interpPre := ArtToInterpState(preState)
	xtracer.Trace("art.PostState post ArtToInterpState")

	// interp.ApplyAction computes the update (action.update(domain, in_scope)),
	// then calls ConcretePost which:
	//   1. Checks the precondition when checkPrecond is true
	//   2. Handles moded-symbol renaming (compose_state_action logic)
	//   3. Computes actions.ForwardImage
	interpPost, err := interp.ApplyAction(checkPrecond, nil, op.Name(), op, interpPre)
	if err != nil {
		return nil, err
	}

	s := InterpToArtState(interpPost)
	s.Action = op
	if abstractor != nil {
		abstractor.Abstract(s)
	}
	return s, nil
}

// JoinStates computes the join (disjunction) of two states using
// interp.ConcreteJoin, which applies actions.JoinState with proper
// differential frame conditions. Matches Python's concrete_join().
func (ag *AnalysisGraph) JoinStates(state1, state2 *State, abstractor Abstractor) *State {
	interp1 := ArtToInterpState(state1)
	interp2 := ArtToInterpState(state2)

	interpJoined, err := interp.ConcreteJoin(interp1, interp2)
	if err != nil {
		log.Printf("art.JoinStates: %v", err)
		// Fallback: simple OR (defensive, matches old behavior)
		var joinedClauses *module.Clauses
		if state1.Clauses != nil && state2.Clauses != nil {
			joinedClauses = module.OrClausesTyped(state1.Clauses, state2.Clauses)
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

	s := InterpToArtState(interpJoined)
	s.JoinOf = []*State{state1, state2}
	s.Label = state1.Label
	if abstractor != nil {
		abstractor.Abstract(s)
	}
	return s
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
	solver := z3bridge.NewSolver(nil, nil)
	t := solver.NewTranslator()
	defer t.Close()

	result, err := t.IsSat(fmla)
	if err != nil {
		log.Printf("art.Unreachable: z3bridge.IsSat error: %v; returning false", err)
		return false
	}
	if result == z3bridge.Unsat {
		node.Clauses = module.FalseClauses(node.Clauses.Annot)
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
	clauseExpr, ok := clause.(lg.Expr)
	if !ok {
		fmt.Printf("ShowCore: formula is not an lg.Expr: %T\n", clause)
		return
	}
	core := interp.GetCore(interpState, clauseExpr)
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
func (ag *AnalysisGraph) Recalculate(checkPrecond bool, t Transition, abstractor Abstractor) (*State, error) {
	var ps *State
	var err error
	if t.Op == nil && t.Label == "join" {
		if t.Post.JoinOf != nil && len(t.Post.JoinOf) >= 2 {
			ps = ag.JoinStates(t.Post.JoinOf[0], t.Post.JoinOf[1], abstractor)
		} else {
			ps = t.Post
		}
	} else {
		if t.Label != "" {
			if a, ok := ag.Actions.Get2(t.Label); ok {
				if act, ok2 := a.(actions.Action); ok2 {
					ps, err = ag.PostState(checkPrecond, act, t.Pre, abstractor)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		if ps == nil {
			ps, err = ag.PostState(checkPrecond, t.Op, t.Pre, abstractor)
			if err != nil {
				return nil, err
			}
		}
	}
	ag.ReplaceState(t.Post, ps)
	return t.Post, nil
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
		if s.Prov != nil {
			// If any state in this expression's args is covered,
			// mark the current state as covered too.
			if aa, ok := s.Prov.(*ActionApp); ok {
				for _, arg := range aa.Args {
					if covered[arg.ID] {
						covered[s.ID] = true
					}
				}
			}
			if sj, ok := s.Prov.(*StateJoin); ok {
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
func (ag *AnalysisGraph) GetHistory(state *State, bound *int) *actions.History {
	// Base case: no predecessor or bound exhausted.
	if state.Pred == nil || (bound != nil && *bound <= 0) {
		// Use the state's clauses as the initial pure state.
		// Python: new_history(state) stores state.value (Clauses) directly.
		clauses := state.Clauses
		if clauses == nil {
			clauses = module.TrueClauses(nil)
		}
		u := actions.PureStateClauses(clauses)
		return actions.NewHistory(ag.Domain.Cfg.IuCfg, u)
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
		var axioms *module.Clauses = module.TrueClauses(nil)
		if state.Pred != nil && state.Pred.Domain != nil {
			bgTheory := state.Pred.Domain.BackgroundTheory(state.Pred.InScope)
			if bgTheory != nil {
				axioms = bgTheory
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
		if state.Prov != nil {
			if aa, ok := state.Prov.(*ActionApp); ok {
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
	h = h.Assume(module.FormulaToClauses(errorCond, nil))

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
func (ag *AnalysisGraph) CheckSafety(checkPrecond bool, state *State) *SafetyResult {
	if len(ag.Assertions) == 0 {
		return &SafetyResult{Safe: true}
	}

	interpState := ArtToInterpState(state)

	// Python: for asn in self.assertions: cex = check_state_assertion(state, asn)
	for _, asn := range ag.Assertions {
		cex := interp.CheckStateAssertion(checkPrecond, interpState, asn)
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
	if state.Prov != nil {
		if aa, ok := state.Prov.(*ActionApp); ok {
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
				_, err := interp.EvalState(checkPrecond, exprNode, ag.Domain)
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
		fmla, ok := lf.Formula.(lg.Expr)
		if !ok {
			continue
		}
		errorCond := &lg.Not{Body: fmla}
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

// Call executes an operation in a sub-graph, adds the result to this graph,
// and creates a transition. Python ivy_art.py:278-284.
func (ag *AnalysisGraph) Call(name string, op func(*AnalysisGraph), prestate *State) *State {
	if prestate == nil {
		prestate = ag.LastState()
	}
	if prestate == nil {
		return nil
	}
	poststate := ag.CallAction(name, op, prestate)
	if poststate == nil {
		return nil
	}
	ag.Add(poststate, nil)
	ag.Transitions = append(ag.Transitions, Transition{
		Pre:   prestate,
		Op:    nil,
		Label: name,
		Post:  poststate,
	})
	return poststate
}

// DecomposeState creates a new AnalysisGraph showing the decomposed
// sub-steps of the action that produced the given state.
// Delegates to interp.DecomposeActionApp which does proper BMC-based
// decomposition matching Python ivy_art.py:392-401 → ivy_interp.py:477-516.
func (ag *AnalysisGraph) DecomposeState(state *State) *AnalysisGraph {
	if state == nil || state.Prov == nil {
		return nil
	}

	// Check if there's a cached subgraph
	if aa, ok := state.Prov.(*ActionApp); ok && aa.Subgraph != nil {
		return aa.Subgraph
	}

	// Build the ast.Node expression from the ActionApp for interp.DecomposeActionApp
	aa, ok := state.Prov.(*ActionApp)
	if !ok {
		return nil
	}
	var actionName string
	switch rep := aa.Rep.(type) {
	case string:
		actionName = rep
	case actions.Action:
		actionName = rep.Name()
	}
	if actionName == "" || len(aa.Args) == 0 {
		return nil
	}

	interpState := ArtToInterpState(state)
	interpPre := ArtToInterpState(aa.Args[0])
	exprNode := interp.ActionApp(ag.Domain.Cfg.AstCfg, actionName, interp.WrapState(interpPre))

	resultState, err := interp.DecomposeActionApp(true, ag.Domain.Cfg.IuCfg, interpState, exprNode)
	if err != nil || resultState == nil {
		return nil
	}

	// Python: other_art = AnalysisGraph(self.domain)
	//         with AC(other_art): res = decompose_action_app(state, state.expr)
	//         other_art.construct_transitions_from_expressions()
	otherArt := NewAnalysisGraph(ag.Domain)

	// Walk the chain of states from resultState back to the root,
	// converting each interp.State to an art.State and adding to the sub-graph.
	var interpStates []*interp.State
	for cur := resultState; cur != nil; cur = cur.Pred() {
		interpStates = append(interpStates, cur)
	}
	// Reverse to get root-first order.
	for i, j := 0, len(interpStates)-1; i < j; i, j = i+1, j-1 {
		interpStates[i], interpStates[j] = interpStates[j], interpStates[i]
	}
	for _, is := range interpStates {
		artSt := InterpToArtState(is)
		otherArt.Add(artSt, artSt.Prov)
	}
	otherArt.ConstructTransitionsFromExpressions()

	// Cache the subgraph on the expression
	aa.Subgraph = otherArt
	return otherArt
}

// DecomposeEdge decomposes a transition by decomposing its poststate.
// Python ivy_art.py:408-410.
func (ag *AnalysisGraph) DecomposeEdge(t Transition) *AnalysisGraph {
	return ag.DecomposeState(t.Post)
}

// ConstructTransitionsFromExpressions rebuilds the transitions list from
// the state expressions. Python ivy_art.py:385-390.
func (ag *AnalysisGraph) ConstructTransitionsFromExpressions() {
	for _, state := range ag.States {
		if state.Prov == nil {
			continue
		}
		aa, ok := state.Prov.(*ActionApp)
		if !ok {
			continue
		}
		if len(aa.Args) == 0 {
			continue
		}
		prestate := aa.Args[0]
		var action actions.Action
		var label string
		switch rep := aa.Rep.(type) {
		case actions.Action:
			action = rep
			label = LabelFromAction(rep)
		case string:
			label = rep
			if a, ok := ag.Actions.Get2(rep); ok {
				if act, ok2 := a.(actions.Action); ok2 {
					action = act
				}
			}
		}
		ag.Transitions = append(ag.Transitions, Transition{
			Pre:   prestate,
			Op:    action,
			Label: label,
			Post:  state,
		})
	}
}

// MakeConcreteTrace is a stub. The Python source (ivy_art.py:404-406) is
// also a stub: the body is just "# TODO\nreturn".
func (ag *AnalysisGraph) MakeConcreteTrace(state *State, conc interface{}) {
	panic("TODO: AnalysisGraph.MakeConcreteTrace() is stubbed.")
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
	for actionName := range ag.Actions.All() {
		if !ag.PublicActions.Get(actionName) {
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
//
// Note: python UI path uses false for checkPrecond here (for
// when we get to that point in the porting).
func (ag *AnalysisGraph) DoStateAction(checkPrecond bool, equation *ast.Definition, abstractor Abstractor) *State {
	ac := ag.Context()
	_ = ac
	rhs := equation.Rhs
	if rhs == nil {
		return nil
	}
	is, err := interp.EvalState(checkPrecond, rhs, ag.Domain)
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
func (ag *AnalysisGraph) RecalculateState(checkPrecond bool, state *State, abstractor Abstractor) {
	if state.Pred != nil && state.Update != nil {
		// Has predecessor: recompute via interp.ConcretePost
		// Python: ps = concrete_post(state.update, state.pred)
		interpPre := ArtToInterpState(state.Pred)
		interpPost, err := interp.ConcretePost(checkPrecond, state.Update, interpPre, nil)
		if err != nil {
			log.Printf("art.RecalculateState: %v", err)
			return
		}
		ps := InterpToArtState(interpPost)
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
		ok, _ := interp.EvalStateOrder(checkPrecondTrue, rhs, interp.WrapState(interpFpc), ag.Domain)
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
	return NewState(ag.Domain, module.FalseClauses(nil))
}

// -----------------------------------------------------------------------
// GUI/Visualization methods (Batch 4)
// -----------------------------------------------------------------------

// ConceptGraph creates a concept graph for the given state.
// Python ivy_art.py:307-316.
func (ag *AnalysisGraph) ConceptGraph(state *State, standardGraph func(*State) ConceptGraphView, clauses *module.Clauses) ConceptGraphView {
	if clauses == nil {
		clauses = state.Clauses
	}
	bg := ag.Domain.BackgroundTheory(state.InScope)
	sg := standardGraph(state)
	sg.SetGraphState(module.AndClausesTyped(clauses, bg))
	sg.SetGraphConcrete(nil)
	return sg
}

// AsCyElements converts this AnalysisGraph into Cytoscape elements for
// browser rendering, matching Python's render_rg (ivy_art.py:459-522).
// The dotLayout callback, if provided, is applied to the result.
func (ag *AnalysisGraph) AsCyElements(dotLayout func(*CyElements) *CyElements) *CyElements {
	g := NewCyElements()

	// Add nodes for states — Python ivy_art.py:467-479
	for _, s := range ag.States {
		var classes []string
		if s.IsBottom() {
			classes = []string{"bottom_state"}
		} else {
			classes = []string{"state"}
		}

		shortInfo := fmt.Sprintf("%d", s.ID)
		// Python: long_info=[str(x) for x in s.clauses.to_open_formula()]
		var longInfo interface{} = shortInfo
		if s.Clauses != nil {
			openFmla := s.Clauses.ToOpenFormula()
			if and, ok := openFmla.(*lg.And); ok && len(and.Terms) > 0 {
				fmlaStrings := make([]string, len(and.Terms))
				for i, term := range and.Terms {
					fmlaStrings[i] = term.String()
				}
				longInfo = fmlaStrings
			}
		}

		g.AddNode(
			fmt.Sprintf("state_%d", s.ID),
			fmt.Sprintf("%d", s.ID),
			classes,
			shortInfo,
			longInfo,
			nil,
			"ellipse",
		)
	}

	// Add edges for transitions — Python ivy_art.py:482-506
	for _, t := range ag.Transitions {
		sourceKey := "state_-1"
		targetKey := "state_-1"
		if t.Pre != nil {
			sourceKey = fmt.Sprintf("state_%d", t.Pre.ID)
		}
		if t.Post != nil {
			targetKey = fmt.Sprintf("state_%d", t.Post.ID)
		}

		var label, info string
		var classes []string

		if t.Label == "join" {
			classes = []string{"transition_join"}
			label = "join"
			info = "join"
		} else {
			classes = []string{"transition_action"}
			label = t.Label
			if label == "" {
				label = "(unlabeled)"
			}
			// Python: label.replace('}',']-').replace('{','-[')
			label = strings.ReplaceAll(label, "}", "]-")
			label = strings.ReplaceAll(label, "{", "-[")
			// Python: label.replace('\n','\\l')+'\\l'
			label = strings.ReplaceAll(label, "\n", "\\l") + "\\l"
			// Python: info = str(op)
			if t.Op != nil {
				info = t.Op.String()
			} else {
				info = label
			}
		}

		edgeKey := fmt.Sprintf("tr_%s_%s", sourceKey, targetKey)
		g.AddEdge(edgeKey, sourceKey, targetKey, label, classes, info, info)
	}

	// Add edges for covering — Python ivy_art.py:509-520
	for _, c := range ag.Covering {
		coveredKey := "state_-1"
		coveringKey := "state_-1"
		if c.Covered != nil {
			coveredKey = fmt.Sprintf("state_%d", c.Covered.ID)
		}
		if c.Covering != nil {
			coveringKey = fmt.Sprintf("state_%d", c.Covering.ID)
		}
		edgeKey := fmt.Sprintf("cover_%s_%s", coveredKey, coveringKey)
		g.AddEdge(edgeKey, coveredKey, coveringKey, "", []string{"cover"}, "", "")
	}

	if dotLayout != nil {
		g = dotLayout(g)
	}
	return g
}

// CheckConstraints is a stub — not present in the Python ivy_art.py.
// It exists as a placeholder for constraint checking in the ARG.
func (ag *AnalysisGraph) CheckConstraints() bool {
	panic("TODO: implement if needed AnalysisGraph.CheckConstraints()")
	return true
}

// StratifyGoals is a stub — not present in the Python ivy_art.py.
// It exists as a placeholder for goal stratification in the ARG.
func (ag *AnalysisGraph) StratifyGoals() []interface{} {
	panic("TODO: implement if needed AnalysisGraph.StratigyGoals()")
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
func (ag *AnalysisGraph) AddInitialState(ic *module.Clauses, abstractor Abstractor) *State {
	mod := ag.Domain

	if ic == nil {
		if mod.InitCond != nil {
			ic = mod.InitCond
		} else {
			ic = module.TrueClauses(nil)
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
			//ec := interp.NewEvalContext(false)
			//ec.Enter()
			s2interp, err := interp.ApplyAction(checkPrecondFalse, nil, "init", env, interpState)
			//ec.Exit()

			if err != nil {
				log.Printf("art.AddInitialState: ApplyAction error: %v", err)
				// Fall through to no-initializer path
			} else {
				s2 := InterpToArtState(s2interp)
				s2.Prov = actionAppExpr
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
			is, err := interp.EvalStateFacts(checkPrecondTrue, p, ag.Domain)
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

// -----------------------------------------------------------------------
// State type adapters: art.State <-> interp.State
// -----------------------------------------------------------------------

// ArtToInterpState converts an art.State to an interp.State for calling
// interp functions. Uses memoization to preserve pointer identity and
// converts art.State.Prov (art.Provenance) to interp.State.Expr (ast.Node).
func ArtToInterpState(s *State) *interp.State {
	return artToInterpMemo(s, make(map[*State]*interp.State))
}

func artToInterpMemo(s *State, memo map[*State]*interp.State) *interp.State {
	if s == nil {
		return nil
	}
	if is, ok := memo[s]; ok {
		return is
	}
	sv := interp.NewStateValue(nil, s.Clauses, module.FalseClauses(nil))
	is := interp.NewState(s.Domain, sv, nil, s.Label)
	memo[s] = is // memo before recursing to break cycles
	is.InScope = s.InScope
	is.Action = s.Action
	is.ActionName = s.ActionName
	if s.Update != nil {
		is.SetUpdate(s.Update)
	}
	if s.Pred != nil {
		is.SetPred(artToInterpMemo(s.Pred, memo))
	}
	if s.JoinOf != nil {
		for _, jo := range s.JoinOf {
			is.JoinOf = append(is.JoinOf, artToInterpMemo(jo, memo))
		}
	}
	// Convert art.State.Prov (art.Provenance) → interp.State.Expr (ast.Node)
	is.Expr = provenanceToInterpExpr(s.Prov, s.Domain, memo)
	return is
}

// provenanceToInterpExpr converts an art.Provenance to an ast.Node
// suitable for interp.State.Expr.
func provenanceToInterpExpr(prov Provenance, domain *module.Module, memo map[*State]*interp.State) ast.Node {
	if prov == nil {
		return nil
	}
	cfg := ast.NewAstConfig()
	if domain != nil && domain.Cfg != nil && domain.Cfg.AstCfg != nil {
		cfg = domain.Cfg.AstCfg
	}
	switch p := prov.(type) {
	case *ActionApp:
		if len(p.Args) > 0 {
			interpPred := artToInterpMemo(p.Args[0], memo)
			actionName := fmt.Sprintf("%v", p.Rep)
			return interp.ActionApp(cfg, actionName, interp.WrapState(interpPred))
		}
	case *StateJoin:
		var terms []ast.Node
		for _, s := range p.Args {
			terms = append(terms, interp.WrapState(artToInterpMemo(s, memo)))
		}
		acfg := ast.NewAstConfig()
		if domain != nil && domain.Cfg != nil && domain.Cfg.AstCfg != nil {
			acfg = domain.Cfg.AstCfg
		}
		return acfg.NewOr(terms...)
	}
	return nil
}

// InterpToArtState converts an interp.State back to an art.State.
// Uses memoization to preserve pointer identity and converts
// interp.State.Expr (ast.Node) to art.State.Prov (art.Provenance).
func InterpToArtState(is *interp.State) *State {
	return interpToArtMemo(is, make(map[*interp.State]*State))
}

func interpToArtMemo(is *interp.State, memo map[*interp.State]*State) *State {
	if is == nil {
		return nil
	}
	if s, ok := memo[is]; ok {
		return s
	}
	s := NewState(is.Domain, is.Clauses)
	memo[is] = s // memo before recursing to break cycles
	s.Label = is.Label
	s.InScope = is.InScope
	s.Action = is.Action
	s.ActionName = is.ActionName
	if is.Update() != nil {
		s.Update = is.Update()
	}
	if is.Pred() != nil {
		s.Pred = interpToArtMemo(is.Pred(), memo)
	}
	if is.JoinOf != nil {
		for _, jo := range is.JoinOf {
			s.JoinOf = append(s.JoinOf, interpToArtMemo(jo, memo))
		}
	}
	// Convert interp.State.Expr (ast.Node) → art.State.Prov (art.Provenance)
	s.Prov = interpExprToProvenance(is.Expr, memo)
	return s
}

// interpExprToProvenance converts an interp-package ast.Node expression
// into an art.Provenance value. Returns nil for unrecognized expressions.
//
// Mapping:
//
//	interp ActionApp (ast.Atom, 1 stateNode arg) → *art.ActionApp
//	interp StateJoin (ast.Or, stateNode args)     → *art.StateJoin
func interpExprToProvenance(expr ast.Node, memo map[*interp.State]*State) Provenance {
	if expr == nil {
		return nil
	}
	if interp.IsActionApp(expr) {
		atom := expr.(*ast.Atom)
		var rep interface{} = atom.Rep
		var args []*State
		for _, term := range atom.Terms {
			if is := interp.UnwrapState(term); is != nil {
				args = append(args, interpToArtMemo(is, memo))
			}
		}
		return &ActionApp{Rep: rep, Args: args}
	}
	if interp.IsStateJoin(expr) {
		or := expr.(*ast.Or)
		var args []*State
		for _, term := range or.Terms {
			if is := interp.UnwrapState(term); is != nil {
				args = append(args, interpToArtMemo(is, memo))
			}
		}
		return &StateJoin{Args: args}
	}
	return nil
}
