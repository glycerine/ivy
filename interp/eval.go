package interp

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	tr "github.com/glycerine/goivy/transrel"
	"github.com/glycerine/goivy/z3bridge"
)

// ---------------------------------------------------------------------------
// Core post / join
// ---------------------------------------------------------------------------

// ConcretePost applies an update to a state, computing the concrete
// post-image. The result is a new state whose predecessor and update
// fields are set for future analysis.
//
// Corresponds to Python's concrete_post() in ivy_interp.py:
//
//	axioms = state.domain.background_theory(state.in_scope)
//	cons = compose_state_action(state.value, axioms, update, check=context.check)
func ConcretePost(update *tr.Update, state *State, expr ast.Node) (*State, error) {
	if state.Domain == nil {
		return nil, fmt.Errorf("ConcretePost: state has nil domain")
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)

	// compose_state_action: compute the forward image of the state
	// through the action update, producing the post-state.
	//
	// The state is in "state style" (Modified=nil for pure states).
	// The update is in "action style". We:
	// 1. Compute the forward image of the state's TR through the update.
	// 2. If check is enabled and the precondition is satisfiable with
	//    the state, raise ActionFailed.
	stateTR := state.Clauses.ToFormula()
	axiomsFmla := axioms.ToFormula()

	// Check precondition if requested.
	preNode := update.PreNode()
	if CurrentContext().Check && preNode != nil && !isNodeFalse(preNode) {
		preCombined := &lg.And{Terms: []lg.Expr{stateTR, axiomsFmla, preNode}}
		t := z3bridge.NewTranslator()
		defer t.Close()
		result, err := t.IsSat(preCombined)
		if err == nil && result == z3bridge.Sat {
			return nil, &tr.ActionFailed{
				Formula: preNode,
				Trace:   []lg.Expr{stateTR},
			}
		}
	}

	// Compute forward image.
	postFmla := tr.ForwardImage(stateTR, axiomsFmla, update)
	postClauses := co.FormulaToClauses(postFmla, state.Clauses.Annot)

	postValue := NewStateValue(
		tr.ModifiedNames(update),
		postClauses,
		co.FalseClauses(nil),
	)
	res := NewState(state.Domain, postValue, expr, "")
	res.SetPred(state)
	res.SetUpdate(update)
	return res, nil
}

// ConcreteJoin joins two states by computing the join (disjunction)
// of their state values using the transrel JoinState operation.
// The resulting state's JoinOf field records the source states.
//
// Corresponds to Python's concrete_join() in ivy_interp.py.
func ConcreteJoin(s1, s2 *State) (*State, error) {
	if s1.Domain == nil || s2.Domain == nil {
		return nil, fmt.Errorf("ConcreteJoin: state has nil domain")
	}
	axioms := s1.Domain.BackgroundTheory(nil)

	// Use transrel.JoinState to compute the state-style join.
	// This adds differential frame conditions and takes the disjunction.
	u1 := stateValueToUpdate(s1.Value())
	u2 := stateValueToUpdate(s2.Value())
	joinedUpdate := tr.JoinState(u1, u2, axioms)

	joinedClauses := joinedUpdate.TR
	joinedPrecond := joinedUpdate.Pre

	joinExpr := StateJoin(WrapState(s1), WrapState(s2))
	res := NewState(s1.Domain, &StateValue{
		Moded:   tr.ModifiedNames(joinedUpdate),
		Clauses: joinedClauses,
		Precond: joinedPrecond,
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
//
// Corresponds to Python's ivy_interp.py apply_action which calls
// action.update(domain, in_scope) to compute the transition relation,
// then compose_state_action to get the post-state.
func ApplyAction(astNode ast.Node, actionName string, action actions.Action, state *State) (*State, error) {
	// Compute the action's transition relation update.
	// Python: upd = action.update(state.domain, state.in_scope)
	ctx := &actions.UpdateContext{
		Domain: state.Domain,
		PVars:  state.InScope,
		GetAction: func(name string) actions.Action {
			if state.Domain != nil {
				if a, ok := state.Domain.Actions[name]; ok {
					if act, ok2 := a.(actions.Action); ok2 {
						return act
					}
				}
			}
			return nil
		},
	}
	upd := actions.IntUpdate(action, ctx)
	if upd == nil {
		upd = tr.NullUpdate()
	}

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
	// State symbol: look up via module.FindAction.
	// Python: res = ivy_actions.context.get(expr.rep) → ivy_module.find_action(symbol)
	if IsStateSymbol(expr) {
		atom := expr.(*ast.Atom)
		res, ok := mod.FindAction(atom.Rep)
		if !ok || res == nil {
			return nil, fmt.Errorf("%s has no value", atom.Rep)
		}
		if s, ok := res.(*State); ok {
			return s, nil
		}
		return nil, fmt.Errorf("%s is not a state", atom.Rep)
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

// isNodeFalse checks if a logic node is the False constant.
func isNodeFalse(n lg.Expr) bool {
	return lg.IsFalse(n)
}

// stateValueToUpdate converts a StateValue to a transrel.Update.
// This maps the state representation to the transrel format.
func stateValueToUpdate(sv *StateValue) *tr.Update {
	trClauses := co.TrueClauses(nil)
	if sv.Clauses != nil {
		trClauses = sv.Clauses
	}
	preClauses := co.FalseClauses(nil)
	if sv.Precond != nil {
		preClauses = sv.Precond
	}
	// Convert string names to Const
	var modified []*lg.Symbol
	for _, name := range sv.Moded {
		modified = append(modified, lg.NewSymbol(name, lg.TopS))
	}
	return &tr.Update{
		Modified: modified,
		TR:       trClauses,
		Pre:      preClauses,
	}
}

// Ensure unused imports don't cause errors.
var _ = z3bridge.Sat
var _ = lg.True
