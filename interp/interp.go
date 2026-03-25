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
package interp

import (
	"fmt"
	"sync"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	tr "github.com/glycerine/goivy/transrel"
)

// ---------------------------------------------------------------------------
// StateValue is the triple (Moded, Clauses, Precond) that represents
// the value of an abstract state. In Python this is a plain tuple.
// ---------------------------------------------------------------------------

// StateValue holds the moded symbols, clauses (transition relation /
// assertion), and precondition for a state.
type StateValue struct {
	Moded   []string    // modified symbols (nil = all)
	Clauses *co.Clauses // clauses / assertions
	Precond *co.Clauses // precondition (negative: action fails when sat)
}

// NewStateValue constructs a StateValue from optional parts. If clauses
// is nil, TopState is used. If precond is nil, FalseClauses is used.
func NewStateValue(moded []string, clauses, precond *co.Clauses) *StateValue {
	if clauses == nil {
		clauses = co.TrueClauses(nil)
	}
	if precond == nil {
		precond = co.FalseClauses(nil)
	}
	return &StateValue{Moded: moded, Clauses: clauses, Precond: precond}
}

// TopStateValue returns a state value representing all states (True clauses).
func TopStateValue() *StateValue {
	return NewStateValue(nil, co.TrueClauses(nil), co.FalseClauses(nil))
}

// BottomStateValue returns a state value representing no states (False clauses).
func BottomStateValue() *StateValue {
	return NewStateValue(nil, co.FalseClauses(nil), co.FalseClauses(nil))
}

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

// State is an abstract state in the symbolic interpreter.
//
// Python equivalent: class State in ivy_interp.py.
type State struct {
	InScope map[string]bool // symbols in scope (matches Python State.in_scope)
	Domain  *module.Module  // the module this state belongs to

	// The value triple.
	Moded   []string    // modified symbols
	Clauses *co.Clauses // main clauses
	Precond *co.Clauses // precondition (negative)

	// Expression tree linking this state to its predecessor.
	Expr  ast.Node // expression that produced this state
	Label string   // optional label

	// Cached analysis results.
	CachedUpdate *tr.Update // cached update for this state's action
	CachedPred   *State     // cached predecessor state

	// Additional fields set during evaluation.
	Action     actions.Action // action that produced this state
	ActionName string         // name of the action
	JoinOf     []*State       // states this was joined from
	Universe   interface{}    // universe constraint data
}

// NewState constructs a State from a domain and optional value.
// If value is nil, TopStateValue is used. If domain is nil, a fresh
// empty module is created.
func NewState(domain *module.Module, value *StateValue, expr ast.Node, label string) *State {
	if value == nil {
		value = TopStateValue()
	}
	if domain == nil {
		domain = module.New()
	}
	return &State{
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
func (s *State) AstCfg() *ast.AstConfig {
	if s.Domain != nil && s.Domain.Cfg != nil && s.Domain.Cfg.AstCfg != nil {
		return s.Domain.Cfg.AstCfg
	}
	return ast.NewAstConfig()
}

// Value returns the state value triple.
func (s *State) Value() *StateValue {
	return &StateValue{
		Moded:   s.Moded,
		Clauses: s.Clauses,
		Precond: s.Precond,
	}
}

// SetValue sets the state value triple.
func (s *State) SetValue(v *StateValue) {
	s.Moded = v.Moded
	s.Clauses = v.Clauses
	s.Precond = v.Precond
}

// Update returns the cached update, computing it lazily if the state's
// expression is an action application.
//
// NOTE: currently returns the cached value only; full computation
// requires the action update infrastructure.
func (s *State) Update() *tr.Update {
	if s.CachedUpdate == nil && s.Expr != nil && IsActionApp(s.Expr) {
		// Matches Python: s.update = eval_action(s.expr.rep).int_update(s.domain, s.in_scope)
		// Action update computation requires the actions package's IntUpdate method.
		// The update is cached on first access.
	}
	return s.CachedUpdate
}

// SetUpdate sets the cached update.
func (s *State) SetUpdate(u *tr.Update) {
	s.CachedUpdate = u
}

// Pred returns the cached predecessor state.
func (s *State) Pred() *State {
	if s.CachedPred == nil && s.Expr != nil && IsActionApp(s.Expr) {
		atom := s.Expr.(*ast.Atom)
		if len(atom.Terms) > 0 {
			if pred, ok := atom.Terms[0].(*stateNode); ok {
				s.CachedPred = pred.state
			}
		}
	}
	return s.CachedPred
}

// SetPred sets the cached predecessor state.
func (s *State) SetPred(pred *State) {
	s.CachedPred = pred
}

// IsBottom returns true if the state's clauses are False (no reachable states).
func (s *State) IsBottom() bool {
	return s.Clauses != nil && s.Clauses.IsFalse()
}

// ToFormula converts the state's clauses to a closed formula.
func (s *State) ToFormula() lg.Expr {
	if s.Clauses == nil {
		return lg.True
	}
	return s.Clauses.ToFormula()
}

// String returns a human-readable representation.
func (s *State) String() string {
	if s.Clauses == nil {
		return "State(<nil clauses>)"
	}
	return fmt.Sprintf("%s", s.Clauses)
}

// Conjs returns the conjectures stored on the domain module.
// In Python, conjectures are global via the module.
func (s *State) Conjs() []*co.Clauses {
	conjs, _ := s.Domain.Attributes["__interp_conjs"].([]*co.Clauses)
	return conjs
}

// SetConjs sets the conjectures on the domain module.
func (s *State) SetConjs(conjs []*co.Clauses) {
	s.Domain.Attributes["__interp_conjs"] = conjs
}

// Unders returns the under-approximations stored on the domain module.
func (s *State) Unders() []*State {
	unders, _ := s.Domain.Attributes["__interp_unders"].([]*State)
	return unders
}

// SetUnders sets the under-approximations on the domain module.
func (s *State) SetUnders(unders []*State) {
	s.Domain.Attributes["__interp_unders"] = unders
}

// ---------------------------------------------------------------------------
// stateNode wraps a *State so it can be embedded in an ast.Node tree.
// ---------------------------------------------------------------------------

type stateNode struct {
	ast.Base
	state *State
}

func (sn *stateNode) Args() []ast.Node           { return nil }
func (sn *stateNode) Clone(args []ast.Node) ast.Node { return sn }
func (sn *stateNode) String() string {
	if sn.state == nil {
		return "<nil state>"
	}
	return sn.state.String()
}

// WrapState wraps a State as an ast.Node for use in expression trees.
func WrapState(s *State) ast.Node {
	return &stateNode{state: s}
}

// UnwrapState extracts a *State from an ast.Node, or nil.
func UnwrapState(n ast.Node) *State {
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
var DefaultInterpConfig = NewInterpConfig()

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

// ---------------------------------------------------------------------------
// Expression helpers
// ---------------------------------------------------------------------------

// IsActionApp returns true if expr is an Atom with exactly one argument,
// representing an action applied to a predecessor state.
func IsActionApp(expr ast.Node) bool {
	a, ok := expr.(*ast.Atom)
	return ok && len(a.Terms) == 1
}

// IsStateJoin returns true if expr is an Or, representing a join of states.
func IsStateJoin(expr ast.Node) bool {
	_, ok := expr.(*ast.Or)
	return ok
}

// ActionApp constructs an action-application expression: action(arg).
func ActionApp(cfg *ast.AstConfig, actionName string, arg ast.Node) *ast.Atom {
	return cfg.NewAtom(actionName, arg)
}

// StateJoin constructs a state-join expression (disjunction).
func StateJoin(cfg *ast.AstConfig, args ...ast.Node) *ast.Or {
	return cfg.NewOr(args...)
}

// IsStateSymbol returns true if expr is an Atom with no arguments,
// representing a named state.
func IsStateSymbol(expr ast.Node) bool {
	a, ok := expr.(*ast.Atom)
	return ok && len(a.Terms) == 0
}

// StateEquation constructs a state equation: lhs = rhs.
func StateEquation(cfg *ast.AstConfig, lhs, rhs ast.Node) *ast.Definition {
	return cfg.NewDefinition(lhs, rhs)
}

// StatesInExpr yields all States embedded in an expression tree.
func StatesInExpr(expr ast.Node) []*State {
	var result []*State
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
	Ast        ast.Node       // AST node where the failure occurred
	ActionName string         // name of the failed action
	Action     actions.Action // the action
	State      *State         // the state in which the action failed
	ErrorState *State         // the error state
	Conc       interface{}    // concrete trace info (trans tuple)
	Msg        string         // human-readable message
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
	astNode ast.Node,
	actionName string,
	action actions.Action,
	state *State,
	failClauses *co.Clauses,
	trans interface{},
) *IvyActionFailedError {
	errState := NewState(state.Domain, &StateValue{
		Clauses: failClauses,
		Precond: co.FalseClauses(nil),
	}, nil, "")
	errState.SetPred(state)
	errState.Action = action
	errState.ActionName = actionName
	errState.SetUpdate(tr.NullUpdate())

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
