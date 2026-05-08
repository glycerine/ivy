// Package interp implements the symbolic interpreter for Ivy.
//
// It corresponds to Python's ivy_interp.py. Key abstractions:
//   - State: an abstract state consisting of a domain (Module), clauses
//     (a moded/clauses/precond triple), optional expression and label,
//     cached update and predecessor, and in-scope symbol map.
//   - EvalContext: provides context (e.g. check flag) for evaluating states
//     and actions.
//   - Core functions for computing concrete post/join, reverse images,
//     applying actions, evaluating expressions, and history (BMC) operations.
//
// Many solver-dependent functions are stubbed because the Z3 bridge
// is not yet fully wired.
package goivy

import (
	"fmt"
	//"sync"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// ---------------------------------------------------------------------------
// StateValue is the triple (Moded, Clauses, Precond) that represents
// the value of an abstract state. In Python this is a plain tuple.
// ---------------------------------------------------------------------------

// StateValue holds the moded symbols, clauses (transition relation /
// assertion), and precondition for a state.
type StateValue struct {
	Moded   []string // modified symbols (nil = all)
	Clauses *Clauses // clauses / assertions
	Precond *Clauses // precondition (negative: action fails when sat)
}

// NewStateValue constructs a StateValue from optional parts. If clauses
// is nil, TopState is used. If precond is nil, FalseClauses is used.
func NewStateValue(moded []string, clauses, precond *Clauses) *StateValue {
	if clauses == nil {
		clauses = TrueClauses(nil)
	}
	if precond == nil {
		precond = FalseClauses(nil)
	}
	return &StateValue{Moded: moded, Clauses: clauses, Precond: precond}
}

// TopStateValue returns a state value representing all states (True clauses).
func TopStateValue() *StateValue {
	return NewStateValue(nil, TrueClauses(nil), FalseClauses(nil))
}

// BottomStateValue returns a state value representing no states (False clauses).
func BottomStateValue() *StateValue {
	return NewStateValue(nil, FalseClauses(nil), FalseClauses(nil))
}

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

// State is an abstract state in the symbolic interpreter.
//
// Python equivalent: class State in ivy_interp.py.
type InterpState struct {
	InScope map[string]bool // symbols in scope (matches Python State.in_scope)
	Domain  *Module         // the module this state belongs to

	// The value triple.
	Moded   []string // modified symbols
	Clauses *Clauses // main clauses
	Precond *Clauses // precondition (negative)

	// Expression tree linking this state to its predecessor.
	Expr  Node   // expression that produced this state
	Label string // optional label

	// Cached analysis results.
	CachedUpdate *Update      // cached update for this state's action
	CachedPred   *InterpState // cached predecessor state

	// Additional fields set during evaluation.
	Action     ActionsAction  // action that produced this state
	ActionName string         // name of the action
	JoinOf     []*InterpState // states this was joined from
	Universe   interface{}    // universe constraint data
}

// NewState constructs a State from a domain and optional value.
// If value is nil, TopStateValue is used. If domain is nil, a fresh
// empty module is created.
func NewInterpState(domain *Module, value *StateValue, expr Node, label string) *InterpState {
	if value == nil {
		value = TopStateValue()
	}
	if domain == nil {
		domain = NewModule()
	}
	return &InterpState{
		InScope: make(map[string]bool),
		Domain:  domain,
		Moded:   value.Moded,
		Clauses: value.Clauses,
		Precond: value.Precond,
		Expr:    expr,
		Label:   label,
	}
}

// AstCfg returns the AstConfig from the state's domain module.
func (s *InterpState) AstCfg() *AstConfig {
	if s.Domain != nil && s.Domain.Cfg != nil && s.Domain.Cfg.AstCfg != nil {
		return s.Domain.Cfg.AstCfg
	}
	return NewAstConfig()
}

// Value returns the state value triple.
func (s *InterpState) Value() *StateValue {
	return &StateValue{
		Moded:   s.Moded,
		Clauses: s.Clauses,
		Precond: s.Precond,
	}
}

// SetValue sets the state value triple.
func (s *InterpState) SetValue(v *StateValue) {
	s.Moded = v.Moded
	s.Clauses = v.Clauses
	s.Precond = v.Precond
}

// Update returns the cached update, computing it lazily on first access.
//
// Mechanical port of Python ivy_interp.py:137-142:
//
//	@property
//	def update(self):
//	    if self.cached_update is None and self.expr is not None and is_action_app(self.expr):
//	        xtracer.trace("interp.InterpState.Update calling GetUpdate type=%s"
//	                      % type(eval_action(self.expr.rep)).__name__)
//	        self.cached_update = eval_action(self.expr.rep).update(self.domain, self.in_scope)
//	    return self.cached_update
//
// Source-of-action lookup differs from Python because ast.Atom.Rep is
// string-only in Go. We use s.Action when set (this is how the
// failState path passes a FailAction object through), and otherwise
// fall back to looking up s.Expr.Rep as an action name in the module.
func (s *InterpState) Update() *Update {
	if s.CachedUpdate != nil {
		return s.CachedUpdate
	}
	var action ActionsAction
	if s.Action != nil {
		action = s.Action
	} else if s.Expr != nil && IsInterpActionApp(s.Expr) {
		if atom, ok := s.Expr.(*Atom); ok && s.Domain != nil {
			if a, found := s.Domain.FindAction(atom.Rep); found {
				if act, ok := a.(ActionsAction); ok {
					action = act
				}
			}
		}
	}
	if action == nil {
		return nil
	}

	xtracer.Trace("interp.State.Update calling GetUpdate type=%s", ActionTypeName(action))

	ctx := &UpdateContext{
		Domain:          s.Domain,
		PVars:           s.InScope,
		ActCfg:          s.Domain.Cfg.ActCfg,
		Instantiator:    s.Domain.Instantiator,
		CheckUnprovable: s.Domain.Cfg.OnlyCheckUnprovable,
		CheckedAssert:   s.Domain.Cfg.CheckLineno,
		GetAction: func(name string) ActionsAction {
			if v, ok := s.Domain.Actions.Get2(name); ok {
				if act, ok := v.(ActionsAction); ok {
					return act
				}
			}
			return nil
		},
	}
	s.CachedUpdate = GetUpdate(action, ctx)
	return s.CachedUpdate
}

// SetUpdate sets the cached update.
func (s *InterpState) SetUpdate(u *Update) {
	s.CachedUpdate = u
}

// Pred returns the cached predecessor state.
func (s *InterpState) Pred() *InterpState {
	if s.CachedPred == nil && s.Expr != nil && IsInterpActionApp(s.Expr) {
		atom := s.Expr.(*Atom)
		if len(atom.Terms) > 0 {
			if pred, ok := atom.Terms[0].(*stateNode); ok {
				s.CachedPred = pred.state
			}
		}
	}
	return s.CachedPred
}

// SetPred sets the cached predecessor state.
func (s *InterpState) SetPred(pred *InterpState) {
	s.CachedPred = pred
}

// IsBottom returns true if the state's clauses are False (no reachable states).
func (s *InterpState) IsBottom() bool {
	return s.Clauses != nil && s.Clauses.IsFalse()
}

// ToFormula converts the state's clauses to a closed formula.
func (s *InterpState) ToFormula() Expr {
	if s.Clauses == nil {
		return True
	}
	return s.Clauses.ToFormula()
}

// String returns a human-readable representation.
func (s *InterpState) String() string {
	if s.Clauses == nil {
		return "State(<nil clauses>)"
	}
	return fmt.Sprintf("%s", s.Clauses)
}

// Conjs returns the conjectures stored on the domain module.
// In Python, conjectures are global via the module.
func (s *InterpState) Conjs() []*Clauses {
	conjs, _ := s.Domain.Attributes["__interp_conjs"].([]*Clauses)
	return conjs
}

// SetConjs sets the conjectures on the domain module.
func (s *InterpState) SetConjs(conjs []*Clauses) {
	s.Domain.Attributes["__interp_conjs"] = conjs
}

// Unders returns the under-approximations stored on the domain module.
func (s *InterpState) Unders() []*InterpState {
	unders, _ := s.Domain.Attributes["__interp_unders"].([]*InterpState)
	return unders
}

// SetUnders sets the under-approximations on the domain module.
func (s *InterpState) SetUnders(unders []*InterpState) {
	s.Domain.Attributes["__interp_unders"] = unders
}

// ---------------------------------------------------------------------------
// stateNode wraps a *State so it can be embedded in an ast.Node tree.
// ---------------------------------------------------------------------------

type stateNode struct {
	Base
	state *InterpState
}

func (sn *stateNode) Args() []Node           { return nil }
func (sn *stateNode) Clone(args []Node) Node { return sn }
func (sn *stateNode) String() string {
	if sn.state == nil {
		return "<nil state>"
	}
	return sn.state.String()
}

// WrapState wraps a State as an ast.Node for use in expression trees.
func WrapState(s *InterpState) Node {
	return &stateNode{state: s}
}

// UnwrapState extracts a *State from an ast.Node, or nil.
func UnwrapState(n Node) *InterpState {
	if sn, ok := n.(*stateNode); ok {
		return sn.state
	}
	return nil
}

// ---------------------------------------------------------------------------
// EvalContext
// ---------------------------------------------------------------------------

// EvalContext provides context for evaluating states and actions.
// The Check flag controls whether preconditions are checked during
// concrete_post.
type EvalContext struct {
	Check      bool
	oldContext *EvalContext
}

// NewEvalContext creates an EvalContext with the given check flag.
func NewEvalContext(check bool) *EvalContext {
	return &EvalContext{Check: check}
}

/* not actually used

// InterpConfig holds per-session interpreter state. Replaces former
// package-level globals contextMu/context for multi-tenancy safety.
type InterpConfig struct {
	mu      sync.Mutex
	context *EvalContext
}

// NewInterpConfig creates a fresh InterpConfig with a default context.
func NewInterpConfig() *InterpConfig {
	return &InterpConfig{
		context: &EvalContext{Check: true},
	}
}

// DefaultInterpConfig is a transitional default for unmigrated callers.
//var DefaultInterpConfig = NewInterpConfig()

// Enter makes ec the current context on this config, saving the old one.
func (ic *InterpConfig) Enter(ec *EvalContext) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ec.oldContext = ic.context
	ic.context = ec
}

// Exit restores the previous context on this config.
func (ic *InterpConfig) Exit(ec *EvalContext) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.context = ec.oldContext
}

// CurrentContext returns the current EvalContext from this config.
func (ic *InterpConfig) CurrentContext() *EvalContext {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.context
}


// Legacy wrappers using DefaultInterpConfig — deprecated.

// Enter makes this context the current global context, saving the old one.
func (ec *EvalContext) Enter() {
	DefaultInterpConfig.Enter(ec)
}

// Exit restores the previous context.
func (ec *EvalContext) Exit() {
	DefaultInterpConfig.Exit(ec)
}

// CurrentContext returns the current global EvalContext.
func CurrentContext() *EvalContext {
	return DefaultInterpConfig.CurrentContext()
}
*/

// ---------------------------------------------------------------------------
// Expression helpers
// ---------------------------------------------------------------------------

// IsInterpActionApp returns true if expr is an Atom with exactly one argument,
// representing an action applied to a predecessor state.
func IsInterpActionApp(expr Node) bool {
	a, ok := expr.(*Atom)
	return ok && len(a.Terms) == 1
}

// IsStateJoin returns true if expr is an Or, representing a join of states.
func IsInterpStateJoin(expr Node) bool {
	_, ok := expr.(*Or)
	return ok
}

// InterpActionApp constructs an action-application expression: action(arg).
func InterpActionApp(cfg *AstConfig, actionName string, arg Node) *Atom {
	return cfg.NewAtom(actionName, arg)
}

// StateJoin constructs a state-join expression (disjunction).
func InterpStateJoin(cfg *AstConfig, args ...Node) *Or {
	return cfg.NewOr(args...)
}

// IsStateSymbol returns true if expr is an Atom with no arguments,
// representing a named state.
func IsStateSymbol(expr Node) bool {
	a, ok := expr.(*Atom)
	return ok && len(a.Terms) == 0
}

// StateEquation constructs a state equation: lhs = rhs.
func StateEquation(cfg *AstConfig, lhs, rhs Node) *Definition {
	return cfg.NewDefinition(lhs, rhs)
}

// StatesInExpr yields all States embedded in an expression tree.
func StatesInExpr(expr Node) []*InterpState {
	var result []*InterpState
	if s := UnwrapState(expr); s != nil {
		result = append(result, s)
		return result
	}
	for _, arg := range expr.Args() {
		result = append(result, StatesInExpr(arg)...)
	}
	return result
}

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

// IvyActionFailedError is returned when an action's precondition fails.
type IvyActionFailedError struct {
	Ast        Node          // AST node where the failure occurred
	ActionName string        // name of the failed action
	Action     ActionsAction // the action
	State      *InterpState  // the state in which the action failed
	ErrorState *InterpState  // the error state
	Conc       interface{}   // concrete trace info (trans tuple)
	Msg        string        // human-readable message
}

func (e *IvyActionFailedError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return fmt.Sprintf("precondition of %q failed", e.ActionName)
}

// NewIvyActionFailedError constructs an IvyActionFailedError, mirroring
// the Python IvyActionFailedError.__init__.
func NewIvyActionFailedError(
	astNode Node,
	actionName string,
	action ActionsAction,
	state *InterpState, failClauses *Clauses,
	trans interface{},
) *IvyActionFailedError {
	errState := NewInterpState(state.Domain, &StateValue{
		Clauses: failClauses,
		Precond: FalseClauses(nil),
	}, nil, "")
	errState.SetPred(state)
	errState.Action = action
	errState.ActionName = actionName
	errState.SetUpdate(NullUpdate())

	return &IvyActionFailedError{
		Ast:        astNode,
		ActionName: actionName,
		Action:     action,
		State:      state,
		ErrorState: errState,
		Conc:       trans,
		Msg:        fmt.Sprintf("precondition of %q failed", actionName),
	}
}

// Ensure interface compliance.
var (
	_ error = (*IvyActionFailedError)(nil)
	_ error = (*UnsatCoreWithInterpolant)(nil)
)
