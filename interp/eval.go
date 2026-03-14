package interp

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	"github.com/glycerine/goivy/module"
	tr "github.com/glycerine/goivy/transrel"
)

// ---------------------------------------------------------------------------
// Core post / join
// ---------------------------------------------------------------------------

// ConcretePost applies an update to a state, computing the concrete
// post-image. The result is a new state whose predecessor and update
// fields are set for future analysis.
//
// TODO: the actual composition (compose_state_action) requires the
// solver infrastructure; this implementation captures the data flow.
func ConcretePost(update *tr.Update, state *State, expr ast.Node) (*State, error) {
	if state.Domain == nil {
		return nil, fmt.Errorf("ConcretePost: state has nil domain")
	}
	// In the full implementation:
	//   axioms := state.Domain.BackgroundTheory(state.InScope)
	//   cons := tr.ComposeStateAction(state.Value(), axioms, update, CurrentContext().Check)
	// For now, we propagate the clauses through the update (stub).
	postValue := NewStateValue(
		update.Modified,
		state.Clauses, // placeholder: should be the composed result
		co.FalseClauses(nil),
	)
	res := NewState(state.Domain, postValue, expr, "")
	res.SetPred(state)
	res.SetUpdate(update)
	return res, nil
}

// ConcreteJoin joins two states by taking the disjunction of their
// clauses. The resulting state's JoinOf field records the source states.
//
// TODO: requires join_state from transrel with background theory.
func ConcreteJoin(s1, s2 *State) (*State, error) {
	if s1.Domain == nil || s2.Domain == nil {
		return nil, fmt.Errorf("ConcreteJoin: state has nil domain")
	}
	// In the full implementation:
	//   value := tr.JoinState(s1.Value(), s2.Value(), s1.Domain.BackgroundTheory())
	// For now, we use OrClauses as a stub.
	joined := co.OrClausesTyped(s1.Clauses, s2.Clauses)
	joinExpr := StateJoin(WrapState(s1), WrapState(s2))
	res := NewState(s1.Domain, &StateValue{
		Clauses: joined,
		Precond: co.FalseClauses(nil),
	}, joinExpr, "")
	res.JoinOf = []*State{s1, s2}
	return res, nil
}

// ---------------------------------------------------------------------------
// Action evaluation
// ---------------------------------------------------------------------------

// EvalAction looks up an action by name in the module. If expr is already
// an Action, it is returned directly.
//
// In Python: eval_action in ivy_interp.py.
func EvalAction(expr interface{}, mod *module.Module) (actions.Action, error) {
	// If it's already an Action, return it.
	if a, ok := expr.(actions.Action); ok {
		return a, nil
	}
	// If it's a FailAction, recursively evaluate.
	if fa, ok := expr.(*FailAction); ok {
		inner, err := EvalAction(fa.Inner, mod)
		if err != nil {
			return nil, err
		}
		return NewFailAction(inner), nil
	}
	// If it's a string, look it up in the module.
	name, ok := expr.(string)
	if !ok {
		return nil, fmt.Errorf("EvalAction: unsupported expression type %T", expr)
	}
	act, found := mod.FindAction(name)
	if !found {
		return nil, fmt.Errorf("%s has no value", name)
	}
	a, ok := act.(actions.Action)
	if !ok {
		return nil, fmt.Errorf("%s is not an action", name)
	}
	return a, nil
}

// ApplyAction applies a named action to a state, returning the post-state.
// If the action's precondition fails (ActionFailed), an IvyActionFailedError
// is returned.
func ApplyAction(astNode ast.Node, actionName string, action actions.Action, state *State) (*State, error) {
	// In the full implementation:
	//   upd := action.Update(state.Domain, state.InScope)
	// Stub: use NullUpdate.
	upd := tr.NullUpdate()

	res, err := ConcretePost(upd, state, ActionApp(actionName, WrapState(state)))
	if err != nil {
		// Check if it's an ActionFailed error.
		if af, ok := err.(*tr.ActionFailed); ok {
			return nil, NewIvyActionFailedError(
				astNode, actionName, action, state,
				co.FormulaToClauses(af.Formula, nil),
				af.Trace,
			)
		}
		return nil, err
	}
	res.Action = action
	res.ActionName = actionName
	return res, nil
}

// ---------------------------------------------------------------------------
// State evaluation
// ---------------------------------------------------------------------------

// EvalStateAtom evaluates a leaf expression to a State.
//
// If the expression is already a State (wrapped), it is returned.
// If it is an AST true/false, the corresponding state is created.
// If it is a state symbol, it is looked up.
// If it is an RME, a state is created from the requires/modifies/ensures.
func EvalStateAtom(expr ast.Node, mod *module.Module) (*State, error) {
	// Check if wrapped state.
	if s := UnwrapState(expr); s != nil {
		return s, nil
	}
	// Check true.
	if ast.IsTrue(expr) {
		return NewState(mod, &StateValue{
			Clauses: co.TrueClauses(nil),
			Precond: co.FalseClauses(nil),
		}, nil, ""), nil
	}
	// Check false.
	if ast.IsFalse(expr) {
		return NewState(mod, &StateValue{
			Clauses: co.FalseClauses(nil),
			Precond: co.FalseClauses(nil),
		}, nil, ""), nil
	}
	// State symbol.
	if IsStateSymbol(expr) {
		atom := expr.(*ast.Atom)
		return nil, fmt.Errorf("%s has no value (state symbol lookup not implemented)", atom.Rep)
	}
	return nil, fmt.Errorf("EvalStateAtom: unsupported expression type %T", expr)
}

// EvalState evaluates an expression tree to a State.
//
// If the expression is a state join (Or), the sub-states are joined.
// If it is an action application, the action is applied to the sub-state.
// Otherwise, it is evaluated as an atom.
func EvalState(expr ast.Node, mod *module.Module) (*State, error) {
	if IsStateJoin(expr) {
		or := expr.(*ast.Or)
		if len(or.Terms) == 0 {
			return nil, fmt.Errorf("EvalState: empty state join")
		}
		result, err := EvalState(or.Terms[0], mod)
		if err != nil {
			return nil, err
		}
		for _, term := range or.Terms[1:] {
			s, err := EvalState(term, mod)
			if err != nil {
				return nil, err
			}
			result, err = ConcreteJoin(result, s)
			if err != nil {
				return nil, err
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
		s, err := EvalState(atom.Terms[0], mod)
		if err != nil {
			return nil, err
		}
		return ApplyAction(expr, atom.Rep, act, s)
	}
	return EvalStateAtom(expr, mod)
}

// BottomState creates a state representing the empty set of states.
func BottomState(domain *module.Module) *State {
	return NewState(domain, BottomStateValue(), StateJoin(), "")
}

// NewStateFromClauses creates a state from clauses, mirroring
// module_new_state in Python.
func NewStateFromClauses(mod *module.Module, clauses *co.Clauses) *State {
	if clauses.Annot == nil {
		clauses = co.NewClauses(clauses.Fmlas, clauses.Defs, actions.EmptyAnnotation{})
	}
	return NewState(mod, &StateValue{
		Clauses: clauses,
		Precond: co.FalseClauses(nil),
	}, nil, "")
}

// NewStateWithValue creates a state from a full StateValue triple.
func NewStateWithValue(mod *module.Module, value *StateValue) *State {
	return NewState(mod, value, nil, "")
}
