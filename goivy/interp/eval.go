package interp

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
	"github.com/glycerine/ivy/goivy/z3bridge"
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
func ConcretePost(checkPrecond bool, update *actions.Update, state *State, expr ast.Node) (*State, error) {
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
	if checkPrecond && preNode != nil && !isNodeFalse(preNode) {
		preCombined := &lg.And{Terms: []lg.Expr{stateTR, axiomsFmla, preNode}}
		solver := z3bridge.NewSolver(nil, nil)
		t := solver.NewTranslator()
		defer t.Close()
		result, err := t.IsSat(preCombined)
		if err == nil && result == z3bridge.Sat {
			return nil, &actions.ActionFailed{
				Formula: preNode,
				Trace:   []lg.Expr{stateTR},
			}
		}
	}

	// Compute forward image.
	postFmla := actions.ForwardImage(stateTR, axiomsFmla, update)
	postClauses := module.FormulaToClauses(postFmla, state.Clauses.Annot)

	postValue := NewStateValue(
		actions.ModifiedNames(update),
		postClauses,
		module.FalseClauses(nil),
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
	joinedUpdate := actions.JoinState(u1, u2, axioms)

	joinedClauses := joinedUpdate.TR
	joinedPrecond := joinedUpdate.Pre

	joinExpr := StateJoin(s1.AstCfg(), WrapState(s1), WrapState(s2))
	res := NewState(s1.Domain, &StateValue{
		Moded:   actions.ModifiedNames(joinedUpdate),
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
	if act == nil {
		return nil, fmt.Errorf("%s is not an action", name)
	}
	return act, nil
}

// ApplyAction applies a named action to a state, returning the post-state.
// If the action's precondition fails (ActionFailed), an IvyActionFailedError
// is returned.
//
// Corresponds to Python's ivy_interp.py apply_action which calls
// action.update(domain, in_scope) to compute the transition relation,
// then compose_state_action to get the post-state.
func ApplyAction(checkPrecond bool, astNode ast.Node, actionName string, action actions.Action, state *State) (*State, error) {
	xtracer.Trace("interp.ApplyAction ENTER actionName=%s", actionName)
	// Compute the action's transition relation update.
	// Python: upd = action.update(state.domain, state.in_scope)
	ctx := &actions.UpdateContext{
		Domain: state.Domain,
		PVars:  state.InScope,
		ActCfg: state.Domain.Cfg.ActCfg,
		GetAction: func(name string) actions.Action {
			if state.Domain != nil {
				if a, ok := state.Domain.Actions.Get2(name); ok {
					if act, ok2 := a.(actions.Action); ok2 {
						return act
					}
				}
			}
			return nil
		},
	}
	xtracer.Trace("interp.ApplyAction post UpdateContext build")
	xtracer.Trace("interp.ApplyAction calling GetUpdate actionName=%s type=%s", actionName, actions.ActionTypeName(action))
	upd := actions.GetUpdate(action, ctx)
	if upd == nil {
		upd = actions.NullUpdate()
	}

	res, err := ConcretePost(checkPrecond, upd, state, ActionApp(state.AstCfg(), actionName, WrapState(state)))
	if err != nil {
		// Check if it's an ActionFailed error.
		if af, ok := err.(*actions.ActionFailed); ok {
			return nil, NewIvyActionFailedError(
				astNode, actionName, action, state,
				module.FormulaToClauses(af.Formula, nil),
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
			Clauses: module.TrueClauses(nil),
			Precond: module.FalseClauses(nil),
		}, nil, ""), nil
	}
	// Check false.
	if ast.IsFalse(expr) {
		return NewState(mod, &StateValue{
			Clauses: module.FalseClauses(nil),
			Precond: module.FalseClauses(nil),
		}, nil, ""), nil
	}
	// State symbol: look up via module.FindAction.
	// Python: res = ivy_actions.context.get(expr.rep) → ivy_module.find_action(symbol)
	// Note: In Python, find_action can return a State (non-action) stored in the
	// module's actions map. In Go, the Actions map is now strongly-typed as
	// map[string]Action, so states won't be stored there. We look up and check
	// if the action wraps a state via the StateAction interface.
	if IsStateSymbol(expr) {
		atom := expr.(*ast.Atom)
		res, ok := mod.FindAction(atom.Rep)
		if !ok || res == nil {
			return nil, fmt.Errorf("%s has no value", atom.Rep)
		}
		// Check if the action provides a state (e.g. via a StateAction wrapper).
		type stateProvider interface {
			GetState() *State
		}
		if sp, ok := res.(stateProvider); ok {
			return sp.GetState(), nil
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
func EvalState(checkPrecond bool, expr ast.Node, mod *module.Module) (*State, error) {
	if IsStateJoin(expr) {
		or := expr.(*ast.Or)
		if len(or.Terms) == 0 {
			return nil, fmt.Errorf("EvalState: empty state join")
		}
		result, err := EvalState(checkPrecond, or.Terms[0], mod)
		if err != nil {
			return nil, err
		}
		for _, term := range or.Terms[1:] {
			s, err := EvalState(checkPrecond, term, mod)
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
		s, err := EvalState(checkPrecond, atom.Terms[0], mod)
		if err != nil {
			return nil, err
		}
		return ApplyAction(checkPrecond, expr, atom.Rep, act, s)
	}
	return EvalStateAtom(expr, mod)
}

// BottomState creates a state representing the empty set of states.
func BottomState(domain *module.Module) *State {
	var acfg *ast.AstConfig
	if domain != nil && domain.Cfg != nil {
		acfg = domain.Cfg.AstCfg
	}
	if acfg == nil {
		acfg = ast.NewAstConfig()
	}
	return NewState(domain, BottomStateValue(), StateJoin(acfg), "")
}

// NewStateFromClauses creates a state from clauses, mirroring
// module_new_state in Python.
func NewStateFromClauses(mod *module.Module, clauses *module.Clauses) *State {
	if clauses.Annot == nil {
		clauses = module.NewClauses(clauses.Fmlas, clauses.Defs, actions.EmptyAnnotation{})
	}
	return NewState(mod, &StateValue{
		Clauses: clauses,
		Precond: module.FalseClauses(nil),
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
func stateValueToUpdate(sv *StateValue) *actions.Update {
	trClauses := module.TrueClauses(nil)
	if sv.Clauses != nil {
		trClauses = sv.Clauses
	}
	preClauses := module.FalseClauses(nil)
	if sv.Precond != nil {
		preClauses = sv.Precond
	}
	// Convert string names to Const
	var modified []*lg.Const
	for _, name := range sv.Moded {
		modified = append(modified, lg.NewConst(name, lg.TopS))
	}
	return &actions.Update{
		Modified: modified,
		TR:       trClauses,
		Pre:      preClauses,
	}
}

// Ensure unused imports don't cause errors.
var _ = z3bridge.Sat
var _ = lg.True
