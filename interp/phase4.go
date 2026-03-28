// phase4.go implements Phase 4.2 functions from the Ivy port.
// These correspond to Python ivy_interp.py helper functions.
package interp

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/solver"
	tr "github.com/glycerine/goivy/transrel"
)

// --- TypeCheckList ---

// TypeCheckList recursively type-checks a list of expressions.
// Items can be nested lists ([]interface{}) or lg.Expr values.
// Corresponds to Python's type_check_list.
func TypeCheckList(domain *module.Module, items []interface{}) error {
	for _, x := range items {
		switch v := x.(type) {
		case []interface{}:
			if err := TypeCheckList(domain, v); err != nil {
				return err
			}
		case lg.Expr:
			if err := actions.TypeCheck(domain, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- ModuleOrder ---

// ModuleOrder checks if state1's clauses are a subset of state2's clauses,
// i.e., state1 implies state2 in state style.
// Returns true if state1 ⊆ state2.
// Corresponds to Python's module_order.
func ModuleOrder(state1, state2 *State) (bool, *tr.CounterExample) {
	axioms := state1.Domain.BackgroundTheory(state1.InScope)
	u1 := stateValueToUpdate(state1.Value())
	u2 := stateValueToUpdate(state2.Value())
	return tr.ImpliesState(u1, u2, axioms)
}

// --- ModuleSkolemizer ---

// ModuleSkolemizer returns a skolemizer function for the module.
// The skolemizer converts variables to fresh constants using a unique renamer.
// Corresponds to Python's module_skolemizer.
func ModuleSkolemizer(mod *module.Module) func(*lg.Variable) *lg.Symbol {
	// Build the list of existing function names
	var funcNames []string
	if mod.Functions != nil {
		for name := range mod.Functions {
			funcNames = append(funcNames, name)
		}
	}
	rn := iu.NewUniqueRenamer("", funcNames)
	return func(v *lg.Variable) *lg.Symbol {
		name := rn.Rename(v.Name)
		return lg.NewSymbol(name, v.VSort)
	}
}

// --- GetCore ---

// GetCore checks whether a clause (as a formula) is implied by a state.
// If so, returns an unsat core (a subset of state clauses that implies
// the clause). Otherwise returns nil.
// Corresponds to Python's get_core.
func GetCore(state *State, clause lg.Expr) *co.Clauses {
	// Python:
	//   clauses1 = and_clauses(state_clauses, background_theory)
	//   clauses2 = [[~lit] for lit in clause]
	//   return unsat_core(clauses1, clauses2)
	stateClauses := state.Clauses
	clauses1 := co.AndClausesTyped(stateClauses, state.Domain.BackgroundTheory(state.InScope))

	// Negate the clause: each literal becomes a singleton clause with its negation
	clauses2 := co.NegateClauses(co.FormulaToClauses(clause, nil))

	slv := solver.New()
	core, err := slv.UnsatCore(clauses1, clauses2, nil, nil)
	if err != nil {
		return nil
	}
	return core
}

// --- ReverseJoinConcreteClauses ---

// ReverseJoinConcreteClauses reverses a join operation by finding which
// joined state is compatible with the given clauses.
// Corresponds to Python's reverse_join_concrete_clauses.
func ReverseJoinConcreteClauses(state *State, joinOf []*State, clauses *co.Clauses) (*co.Clauses, *State, error) {
	if clauses == nil {
		clauses = state.Clauses
	}
	for _, s := range joinOf {
		combined := co.AndClausesTyped(s.Clauses, clauses)
		if !combined.IsFalse() {
			return combined, s, nil
		}
	}

	// No compatible joined state found. Compute interpolant from
	// the disjunction of all joined states vs the target clauses.
	// Python: ivy_interp.py:258-262
	axioms := state.Domain.BackgroundTheory(state.InScope)
	interpreted := functionsToInterpreted(state.Domain.Functions)
	clausesOfStates := make([]*co.Clauses, len(joinOf))
	for i, s := range joinOf {
		clausesOfStates[i] = s.Clauses
	}
	pre := co.OrClausesTyped(clausesOfStates...)
	itp := tr.Interpolant(pre, clauses, axioms, interpreted)
	if itp != nil {
		return nil, nil, &UnsatCoreWithInterpolant{Core: itp.Core, Itp: itp.Itp}
	}
	return nil, nil, fmt.Errorf("decision procedure incompleteness")
}

// --- UnderapproximateState ---

// UnderapproximateState builds an under-approximation of reachable states
// using model extraction from the state's clauses.
// Corresponds to Python's underapproximate_state.
func UnderapproximateState(state *State, implied *co.Clauses) {
	// Python:
	//   axioms = state.domain.background_theory(state.in_scope)
	//   under = clauses_model_to_clauses(and_clauses(state.clauses, axioms), is_skolem, implied)
	//   if under != None:
	//       state.unders.append(self.new_state(under))
	axioms := state.Domain.BackgroundTheory(state.InScope)
	combined := co.AndClausesTyped(state.Clauses, axioms)

	slv := solver.New()
	under, err := slv.ClausesModelToClauses(
		combined,
		func(s *lg.Symbol) bool {
			return tr.IsSkolem(s.Name)
		},
	)
	if err != nil || under == nil {
		return
	}
	AddUnder(state, under, nil, nil)
}

// --- StatesStateExpr ---

// StatesStateExpr yields all State objects embedded in an expression tree.
// This is an alias for StatesInExpr which was already implemented.
// Corresponds to Python's states_state_expr.
func StatesStateExpr(expr ast.Node) []*State {
	return StatesInExpr(expr)
}

// --- DecomposeActionApp ---

// DecomposeActionApp decomposes an action application into intermediate
// states using BMC/History and action decomposition.
// Corresponds to Python's decompose_action_app.
func DecomposeActionApp(checkPrecond bool, cfg *iu.IvyUtilsConfig, state2 *State, expr ast.Node) (*State, error) {
	if !IsActionApp(expr) {
		return nil, nil
	}
	atom := expr.(*ast.Atom)
	act, err := EvalAction(atom.Rep, state2.Domain)
	if err != nil {
		return nil, err
	}
	state1, err := EvalState(checkPrecond, atom.Terms[0], state2.Domain)
	if err != nil {
		return nil, err
	}

	// Compute update
	ctx := &actions.UpdateContext{
		Domain: state1.Domain,
		PVars:  state1.InScope,
		GetAction: func(name string) actions.Action {
			if state1.Domain != nil {
				if a, ok := state1.Domain.Actions[name]; ok {
					if act, ok2 := a.(actions.Action); ok2 {
						return act
					}
				}
			}
			return nil
		},
	}
	upd := actions.IntUpdate(act, ctx)
	if upd == nil {
		upd = tr.NullUpdate()
	}

	// Use decomposition: try each decomposition path
	comps := actions.DecomposeWithState(act, state1.Clauses.ToFormula(), state2.Clauses.ToFormula(), false)
	bg := state1.Domain.BackgroundTheory(state1.InScope)

	for _, comp := range comps {
		// Compute updates for each action in the decomposition
		upds := make([]*tr.Update, len(comp.Actions))
		for i, subAct := range comp.Actions {
			subUpd := actions.IntUpdate(subAct, ctx)
			if subUpd == nil {
				subUpd = tr.NullUpdate()
			}
			upds[i] = subUpd
		}

		// Build a history from pre-state
		h := tr.NewHistory(cfg, tr.PureState(comp.Pre))

		// Forward-step through each update
		for _, upd := range upds {
			h = h.ForwardStep(bg.ToFormula(), upd, nil)
		}

		// Assume the post-state
		if comp.Post != nil {
			h = h.Assume(comp.Post)
		}

		// Check satisfiability
		bmcRes := h.Satisfy(bg.ToFormula())
		if bmcRes == nil {
			continue
		}

		// Build per-step states from the satisfying path.
		// Python ivy_interp.py:504-515
		var states []*State
		for i, value := range bmcRes.Path {
			state := NewStateFromClauses(state1.Domain, value.TR)
			if i != 0 {
				state.Expr = ActionApp(state1.AstCfg(), comp.Actions[i-1].Name(), WrapState(states[len(states)-1]))
				state.SetUpdate(upds[i-1])
				state.SetPred(states[len(states)-1])
			}
			state.Label = ""
			state.Universe = bmcRes.Universes
			states = append(states, state)
		}
		if len(states) > 0 {
			return states[len(states)-1], nil
		}
		return nil, nil
	}
	return nil, nil
}

// --- StateImpliesFormula ---

// StateImpliesFormula checks if a state logically implies a formula.
// Corresponds to Python's state_implies_formula.
func StateImpliesFormula(state *State, fmla lg.Expr) bool {
	axioms := state.Domain.BackgroundTheory(state.InScope)
	combined := co.AndClausesTyped(state.Clauses, axioms)
	ok, _ := tr.ClausesImplyFormulaCex(combined, fmla)
	return ok
}

// --- NewHistory (already implemented as NewHistoryFromState) ---

// --- EvalAssertRhs ---

// EvalAssertRhs evaluates the right-hand side of a state assertion.
// If the RHS is not already an RME, wraps it in one.
// Corresponds to Python's eval_assert_rhs.
func EvalAssertRhs(checkPrecond bool, rhs interface{}, domain *module.Module) (*State, error) {
	// Python:
	//   if not isinstance(rhs, ivy_actions.RME):
	//       rhs = ivy_actions.RME(And(), None, rhs)
	//   with ivy_actions.ActionContext(domain):
	//       return eval_state(rhs)
	rmeVal, ok := rhs.(*actions.RME)
	if !ok {
		// Wrap non-RME in RME(And(), nil, rhs)
		var rhsNode lg.Expr
		if n, ok2 := rhs.(lg.Expr); ok2 {
			rhsNode = n
		} else if n, ok2 := rhs.(ast.Node); ok2 {
			// For ast.Node, evaluate directly within ActionContext
			ctx := actions.NewActionContext(domain)
			_ = ctx
			return EvalState(checkPrecond, n, domain)
		}
		rmeVal = actions.NewRME(&lg.And{}, nil, rhsNode)
	}

	// Evaluate within an ActionContext
	ctx := actions.NewActionContext(domain)
	_ = ctx
	// Convert RME to state: the RME's ensures formula becomes the state constraint
	if rmeVal.Ensures != nil {
		cls := co.FormulaToClauses(rmeVal.Ensures, nil)
		return NewStateFromClauses(domain, cls), nil
	}
	return NewStateFromClauses(domain, co.TrueClauses(nil)), nil
}

// --- EvalStateOrder ---

// EvalStateOrder evaluates a state ordering relation between two expressions.
// Returns true if lhs ⊆ rhs.
// Corresponds to Python's eval_state_order.
func EvalStateOrder(checkPrecond bool, lhs, rhs ast.Node, mod *module.Module) (bool, error) {
	state, err := EvalState(checkPrecond, lhs, mod)
	if err != nil {
		return false, err
	}
	if IsStateJoin(rhs) {
		or := rhs.(*ast.Or)
		for _, r := range or.Terms {
			rState, err := EvalState(checkPrecond, r, mod)
			if err != nil {
				continue
			}
			ok, _ := ModuleOrder(state, rState)
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	rState, err := EvalState(checkPrecond, rhs, mod)
	if err != nil {
		return false, err
	}
	ok, _ := ModuleOrder(state, rState)
	return ok, nil
}

// --- CheckStateAssertion ---

// CheckStateAssertion checks if a state satisfies an assertion.
// Returns true if the state satisfies the assertion (or if no check is needed).
// Corresponds to Python's check_state_assertion.
//
// Note that checkPrecond defaults to true in the python, so
// use true if not certain what to pass here.
//
// This comes from ivy_interp.py, where Python Ivy has a single global
// `context = EvalContext(check=True)`. The `.check` field is
// read in exactly **one place**: `concrete_post()` ivy_interp.py:202,
// which passes it to `compose_state_action(..., check=context.check)`.
// That function checks whether action preconditions are
// satisfiable and raises `ActionFailed` if check=True and
// the precondition is violated.
// See also the file 'goivy/already_applied_plans/PRECOND_CHECK.md'.
func CheckStateAssertion(checkPrecond bool, state *State, assertion *ast.LabeledFormula) bool {
	if state.Label == "" {
		return true
	}
	// Check if the assertion's label matches the state's label
	if assertion.Label == nil {
		return true
	}
	labelSym, ok := assertion.Label.(*lg.Symbol)
	if !ok {
		return true
	}
	if state.Label != labelSym.Name {
		return true
	}
	// Evaluate the RHS and check ordering
	rhsState, err := EvalAssertRhs(checkPrecond, assertion.Formula, state.Domain)
	if err != nil {
		return true
	}
	ok2, _ := ModuleOrder(state, rhsState)
	return ok2
}

// --- GetStateAssertions ---

// GetStateAssertions collects all assertions matching a state's label
// and returns their conjunction. Returns nil if no assertions match
// or all are trivially true.
// Corresponds to Python's get_state_assertions.
func GetStateAssertions(checkPrecond bool, state *State, mod *module.Module) *co.Clauses {
	if state.Label == "" {
		return nil
	}
	res := co.TrueClauses(nil)
	for _, assertion := range mod.Assertions {
		if assertion.Label == nil {
			continue
		}
		labelSym, ok := assertion.Label.(*lg.Symbol)
		if !ok {
			continue
		}
		if state.Label == labelSym.Name {
			rhsState, err := EvalAssertRhs(checkPrecond, assertion.Formula, state.Domain)
			if err != nil {
				continue
			}
			res = co.AndClausesTyped(res, rhsState.Clauses)
		}
	}
	if res.IsTrue() {
		return nil
	}
	return res
}

// --- UniverseConstraint ---

// SortUniverse pairs a Sort with its universe values.
// Use []SortUniverse instead of map[lg.Sort][]lg.Expr to avoid
// pointer-equality misses on lg.Sort interface map keys.
type SortUniverse struct {
	Sort   lg.Sort
	Values []lg.Expr
}

// UniverseConstraint creates clauses constraining universe values.
// If the state has universe data (from model finding), generates
// equality constraints for each sort.
// Corresponds to Python's universe_constraint.
func UniverseConstraint(state *State) *co.Clauses {
	if state.Universe == nil {
		return co.TrueClauses(nil)
	}
	// Universe is a map from sort -> []values.
	// Accept both the legacy map[lg.Sort][]lg.Expr and the
	// safe []SortUniverse (which avoids pointer-equality misses
	// on lg.Sort interface map keys).
	type sortEntry struct {
		sort   lg.Sort
		values []lg.Expr
	}
	var entries []sortEntry
	switch um := state.Universe.(type) {
	case []SortUniverse:
		for _, su := range um {
			entries = append(entries, sortEntry{su.Sort, su.Values})
		}
	case map[lg.Sort][]lg.Expr:
		for s, vals := range um {
			entries = append(entries, sortEntry{s, vals})
		}
	default:
		return co.TrueClauses(nil)
	}
	var fmlas []lg.Expr
	for _, e := range entries {
		s, values := e.sort, e.values
		if len(values) == 0 {
			continue
		}
		x, _ := lg.NewVariable("X", s)
		var disjuncts []lg.Expr
		for _, v := range values {
			disjuncts = append(disjuncts, &lg.Eq{T1: x, T2: v})
		}
		fmla := il.ForAll([]*lg.Variable{x}, &lg.Or{Terms: disjuncts})
		fmlas = append(fmlas, fmla)
	}
	if len(fmlas) == 0 {
		return co.TrueClauses(nil)
	}
	return co.NewClauses(fmlas, nil, nil)
}
