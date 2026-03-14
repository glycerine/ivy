package interp

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	tr "github.com/glycerine/goivy/transrel"
)

// ---------------------------------------------------------------------------
// FailAction - wraps an action so it computes action_failure(upd)
// ---------------------------------------------------------------------------

// FailAction wraps an inner action so that its update is transformed
// via action_failure. This corresponds to Python's fail_action class.
type FailAction struct {
	actions.ActionBase
	Inner actions.Action
}

// NewFailAction creates a FailAction wrapping the given action.
func NewFailAction(inner actions.Action) *FailAction {
	fa := &FailAction{Inner: inner}
	// Copy lineno if available.
	fa.ActionBase.SetLineno(inner.GetLineno())
	return fa
}

func (fa *FailAction) Name() string { return "fail" }

func (fa *FailAction) String() string {
	return "fail " + fa.Inner.String()
}

func (fa *FailAction) Args() []lg.Node {
	return fa.Inner.Args()
}

func (fa *FailAction) Clone(args []lg.Node) actions.Action {
	return &FailAction{
		ActionBase: fa.ActionBase,
		Inner:      fa.Inner.Clone(args),
	}
}

func (fa *FailAction) IterCalls() []string {
	return fa.Inner.IterCalls()
}

func (fa *FailAction) IterSubactions() []actions.Action {
	return []actions.Action{fa}
}

// FailedAction returns the wrapped inner action.
func (fa *FailAction) FailedAction() actions.Action {
	return fa.Inner
}

// ---------------------------------------------------------------------------
// Reverse image helpers
// ---------------------------------------------------------------------------

// Reverse computes the reverse image (concrete pre-state) of a state
// through its update. If clauses is nil, the state's own clauses are used.
//
// TODO: requires solver infrastructure (reverse_image, forward_interpolant).
func Reverse(state *State, clauses *co.Clauses) (*co.Clauses, error) {
	if state.Pred() == nil || state.Update() == nil {
		return nil, fmt.Errorf("Reverse: cannot reverse state without predecessor and update")
	}
	if clauses == nil {
		clauses = state.Clauses
	}
	// In full implementation:
	//   axioms := state.Domain.BackgroundTheory(state.InScope)
	//   return co.AndClausesTyped(tr.ReverseImage(clauses, axioms, state.Update()), axioms), nil
	// Stub: return the clauses unchanged.
	return clauses, nil
}

// ReverseUpdateConcreteClauses reverses an update concretely. If unsat,
// returns an UnsatCoreWithInterpolant error.
//
// TODO: requires solver infrastructure.
func ReverseUpdateConcreteClauses(state *State, clauses *co.Clauses) (*co.Clauses, error) {
	if state.Pred() == nil || state.Update() == nil {
		return nil, fmt.Errorf("ReverseUpdateConcreteClauses: no predecessor or update")
	}
	if clauses == nil {
		clauses = state.Clauses
	}
	// Stub: return the clauses unchanged.
	return clauses, nil
}

// ---------------------------------------------------------------------------
// Reach state helpers
// ---------------------------------------------------------------------------

// JoinUnders computes the join (disjunction) of all under-approximation
// states. Returns TrueClauses if there are no under-approximations.
func JoinUnders(state *State) *co.Clauses {
	unders := state.Unders()
	if len(unders) == 0 {
		return co.TrueClauses(nil)
	}
	clauses := make([]*co.Clauses, len(unders))
	for i, u := range unders {
		clauses[i] = u.Clauses
	}
	return co.OrClausesTyped(clauses...)
}

// AddUnder adds an under-approximation state to the target state.
func AddUnder(state *State, clauses *co.Clauses, pred *State, universe interface{}) *State {
	s := NewStateFromClauses(state.Domain, clauses)
	if pred != nil {
		s.SetPred(pred)
	}
	if universe != nil {
		s.Universe = universe
	}
	unders := state.Unders()
	unders = append(unders, s)
	state.SetUnders(unders)
	return s
}

// ReachState tries to reach a state in one step from its predecessor's
// under-approximation. Returns nil if not reachable.
//
// TODO: requires solver infrastructure (get_model_clauses, forward_image).
func ReachState(state *State, clauses *co.Clauses) *State {
	if state.Pred() == nil || state.Update() == nil {
		return nil
	}
	// Stub: not yet implemented (requires solver).
	return nil
}

// ReachStateFromPred attempts to reach a state from its predecessor's
// under-approximation. If not reachable, returns an
// UnsatCoreWithInterpolant error.
//
// TODO: requires solver infrastructure.
func ReachStateFromPred(state *State, clauses *co.Clauses) (*State, error) {
	post := ReachState(state, clauses)
	if post != nil {
		return post, nil
	}
	// Stub: would call reverse_interpolant_case here.
	return nil, nil
}

// ---------------------------------------------------------------------------
// Conjecture helpers
// ---------------------------------------------------------------------------

// UndecidedConjectures returns conjectures of state1 that are not
// implied by the state's clauses and background theory.
//
// TODO: requires solver infrastructure (clauses_imply_list).
func UndecidedConjectures(state *State) []*co.Clauses {
	// Stub: return all conjectures as undecided.
	return state.Conjs()
}

// FilterConjectures partitions the state's conjectures into those
// implied by the model (kept) and those not implied (lost).
//
// TODO: requires solver infrastructure (clauses_imply).
func FilterConjectures(state *State, model *co.Clauses) []*co.Clauses {
	// Stub: keep all conjectures.
	conjs := state.Conjs()
	state.SetConjs(conjs)
	return nil // no lost conjectures
}

// CaseConjecture conjectures a separator between the state's
// under-approximation and the given clauses.
//
// TODO: requires solver infrastructure (interpolant_case).
func CaseConjecture(state *State, clauses *co.Clauses) (interface{}, interface{}, bool) {
	// Stub: not implemented.
	return nil, nil, false
}

// ---------------------------------------------------------------------------
// Diagram
// ---------------------------------------------------------------------------

// Diagram returns the diagram of a single model of clauses in the
// given state, or nil if the clauses are unsatisfiable.
//
// TODO: requires solver infrastructure (clauses_model_to_diagram).
func Diagram(state *State, clauses *co.Clauses, implied *co.Clauses, extraAxioms *co.Clauses, weaken, upwardClose bool) *co.Clauses {
	// Stub: not implemented.
	return nil
}

// ---------------------------------------------------------------------------
// History (BMC) helpers
// ---------------------------------------------------------------------------

// NewHistory creates a History from a state's value.
func NewHistoryFromState(state *State) *tr.History {
	return tr.NewHistory(&tr.Update{
		Modified: nil, // pure state
		TR:       state.ToFormula(),
		Pre:      lg.False,
	})
}

// HistoryForwardStep advances a history by one step through the
// state's update.
func HistoryForwardStep(history *tr.History, state *State) *tr.History {
	var actionNode lg.Node
	if state.Expr != nil && IsActionApp(state.Expr) {
		atom := state.Expr.(*ast.Atom)
		actionNode = lg.NewConst(atom.Rep, lg.Boolean)
	}
	pred := state.Pred()
	if pred == nil {
		return history
	}
	// In full implementation:
	//   bg := pred.Domain.BackgroundTheory(pred.InScope)
	// Stub: use True.
	bg := lg.True
	return history.ForwardStep(bg, state.Update(), actionNode)
}

// HistorySatisfy checks whether a history is satisfiable in the
// given state's background theory.
//
// TODO: requires solver infrastructure.
func HistorySatisfy(history *tr.History, state *State) (interface{}, []interface{}) {
	// Stub: not implemented.
	return nil, nil
}

// ---------------------------------------------------------------------------
// Module helper functions (replacing Python monkey-patches)
// ---------------------------------------------------------------------------

// ModuleNewState creates a new State from clauses, adding an empty
// annotation if none is present. Mirrors module_new_state.
func ModuleNewState(mod *module.Module, clauses *co.Clauses) *State {
	return NewStateFromClauses(mod, clauses)
}

// ModuleNewStateWithValue creates a State from a full value triple.
// Mirrors module_new_state_with_value.
func ModuleNewStateWithValue(mod *module.Module, value *StateValue) *State {
	return NewState(mod, value, nil, "")
}

// ModuleTypeCheck type-checks the module's axioms.
//
// TODO: requires type_check_list from actions package.
func ModuleTypeCheck(mod *module.Module) error {
	// Stub: no-op for now.
	return nil
}

// ModuleTypeCheckConcepts type-checks concept spaces.
//
// TODO: requires concept space iteration.
func ModuleTypeCheckConcepts(mod *module.Module) error {
	// Stub: no-op for now.
	return nil
}

// ---------------------------------------------------------------------------
// Property checking
// ---------------------------------------------------------------------------

// FalseProperties returns the labeled properties of the module that
// are false (not implied by the background theory).
//
// TODO: requires solver infrastructure.
func FalseProperties(mod *module.Module) []*module.LabeledFormula {
	// Stub: return empty.
	return nil
}

// GetPropertyContext returns the accumulated context (conjunction of
// prior subgoal properties) for a given property.
func GetPropertyContext(mod *module.Module, prop *module.LabeledFormula) *co.Clauses {
	res := co.TrueClauses(nil)
	// Stub: would iterate over LabeledProps and accumulate subgoals.
	_ = res
	return co.TrueClauses(nil)
}

// ---------------------------------------------------------------------------
// Additional eval helpers
// ---------------------------------------------------------------------------

// EvalStateFacts evaluates an expression tree, skipping state symbols
// (returning nil for them). Mirrors eval_state_facts.
func EvalStateFacts(expr ast.Node, mod *module.Module) (*State, error) {
	if IsStateJoin(expr) {
		or := expr.(*ast.Or)
		var result *State
		for _, term := range or.Terms {
			s, err := EvalStateFacts(term, mod)
			if err != nil {
				return nil, err
			}
			if s == nil {
				continue
			}
			if result == nil {
				result = s
			} else {
				result, err = ConcreteJoin(result, s)
				if err != nil {
					return nil, err
				}
			}
		}
		return result, nil
	}
	if IsActionApp(expr) {
		atom := expr.(*ast.Atom)
		act, err := EvalAction(atom.Rep, mod)
		if err != nil {
			return nil, err
		}
		s, err := EvalStateFacts(atom.Terms[0], mod)
		if err != nil {
			return nil, err
		}
		if s == nil {
			return nil, nil
		}
		return ApplyAction(expr, atom.Rep, act, s)
	}
	if IsStateSymbol(expr) {
		return nil, nil
	}
	return EvalStateAtom(expr, mod)
}

// EvalStateActions extracts action applications from an expression
// that reference the given predecessor state. Mirrors eval_state_actions.
func EvalStateActions(expr ast.Node, pre *State) []*ast.Atom {
	if IsStateJoin(expr) {
		or := expr.(*ast.Or)
		var result []*ast.Atom
		for _, term := range or.Terms {
			result = append(result, EvalStateActions(term, pre)...)
		}
		return result
	}
	if IsActionApp(expr) {
		atom := expr.(*ast.Atom)
		if IsStateSymbol(atom.Terms[0]) {
			inner := atom.Terms[0].(*ast.Atom)
			if inner.Rep == pre.Label {
				return []*ast.Atom{ActionApp(atom.Rep, WrapState(pre))}
			}
		}
	}
	return nil
}

// TopAlpha resets a state's clauses to TrueClauses.
func TopAlpha(state *State) {
	state.Clauses = co.TrueClauses(nil)
}

// FailExpr constructs a fail expression from an action application.
func FailExpr(expr *ast.Atom) *ast.Atom {
	return ActionApp("fail_"+expr.Rep, expr.Terms[0])
}

// Ensure unused imports don't cause errors.
var _ = fmt.Sprintf
var _ actions.Action
var _ *module.Module
var _ *tr.Update
