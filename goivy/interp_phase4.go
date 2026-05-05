// phase4.go implements Phase 4.2 functions from the Ivy port.
// These correspond to Python ivy_interp.py helper functions.
package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// --- TypeCheckList ---

// TypeCheckList recursively type-checks a list of expressions.
// Items can be nested lists ([]interface{}) or lg.Expr values.
// Corresponds to Python's type_check_list.
func TypeCheckList(domain *Module, items []interface{}) error {
	for _, x := range items {
		switch v := x.(type) {
		case []interface{}:
			if err := TypeCheckList(domain, v); err != nil {
				return err
			}
		case Expr:
			if err := TypeCheck(domain, v); err != nil {
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
func ModuleOrder(state1, state2 *InterpState) (bool, *CounterExample) {
	axioms := state1.Domain.BackgroundTheory(state1.InScope)
	u1 := stateValueToUpdate(state1.Value())
	u2 := stateValueToUpdate(state2.Value())
	return ImpliesState(state1.Domain, u1, u2, axioms)
}

// --- ModuleSkolemizer ---

// ModuleSkolemizer returns a skolemizer function for the module.
// The skolemizer converts variables to fresh constants using a unique renamer.
// Corresponds to Python's module_skolemizer.
func ModuleSkolemizer(mod *Module) func(*LogicVariable) *Const {
	// Build the list of existing function names
	var funcNames []string
	if mod.Functions != nil {
		for name := range mod.Functions.All() {
			funcNames = append(funcNames, name)
		}
	}
	rn := NewUniqueRenamer("", funcNames)
	return func(v *LogicVariable) *Const {
		name := rn.Rename(v.Name)
		return NewConst(name, v.VSort)
	}
}

// --- GetCore ---

// GetCore checks whether a clause (as a formula) is implied by a state.
// If so, returns an unsat core (a subset of state clauses that implies
// the clause). Otherwise returns nil.
// Corresponds to Python's get_core.
func GetCore(state *InterpState, clause Expr) *Clauses {
	// Python:
	//   clauses1 = and_clauses(state_clauses, background_theory)
	//   clauses2 = [[~lit] for lit in clause]
	//   return unsat_core(clauses1, clauses2)
	stateClauses := state.Clauses
	clauses1 := AndClausesTyped(stateClauses, state.Domain.BackgroundTheory(state.InScope))

	// Negate the clause: each literal becomes a singleton clause with its negation
	clauses2 := NegateClauses(FormulaToClauses(clause, nil))

	slv := NewSolver(state.Domain, nil)
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
func ReverseJoinConcreteClauses(state *InterpState, joinOf []*InterpState, clauses *Clauses) (*Clauses, *InterpState, error) {
	if clauses == nil {
		clauses = state.Clauses
	}
	for _, s := range joinOf {
		combined := AndClausesTyped(s.Clauses, clauses)
		if !combined.IsFalse() {
			return combined, s, nil
		}
	}

	// No compatible joined state found. Compute interpolant from
	// the disjunction of all joined states vs the target clauses.
	// Python: ivy_interp.py:258-262
	axioms := state.Domain.BackgroundTheory(state.InScope)
	interpreted := functionsToInterpreted(state.Domain.Functions)
	clausesOfStates := make([]*Clauses, len(joinOf))
	for i, s := range joinOf {
		clausesOfStates[i] = s.Clauses
	}
	pre := OrClausesTyped(clausesOfStates...)
	itp := Interpolant(state.Domain, pre, clauses, axioms, interpreted)
	if itp != nil {
		return nil, nil, &UnsatCoreWithInterpolant{Core: itp.Core, Itp: itp.Itp}
	}
	return nil, nil, fmt.Errorf("decision procedure incompleteness")
}

// --- UnderapproximateState ---

// UnderapproximateState builds an under-approximation of reachable states
// using model extraction from the state's clauses.
// Corresponds to Python's underapproximate_state.
func UnderapproximateState(state *InterpState, implied *Clauses) {
	// Python:
	//   axioms = state.domain.background_theory(state.in_scope)
	//   under = clauses_model_to_clauses(and_clauses(state.clauses, axioms), is_skolem, implied)
	//   if under != None:
	//       state.unders.append(self.new_state(under))
	axioms := state.Domain.BackgroundTheory(state.InScope)
	combined := AndClausesTyped(state.Clauses, axioms)

	slv := NewSolver(state.Domain, nil)
	under, err := slv.ClausesModelToClauses(
		combined,
		func(s *Const) bool {
			return IsSkolem(s.Name)
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
func StatesStateExpr(expr Node) []*InterpState {
	return StatesInExpr(expr)
}

// --- DecomposeActionApp ---

// DecomposeActionApp decomposes an action application into intermediate
// states using BMC/History and action decomposition.
// Corresponds to Python's decompose_action_app.
func DecomposeActionApp(checkPrecond bool, cfg *IvyUtilsConfig, state2 *InterpState, expr Node) (*InterpState, error) {
	if !IsInterpActionApp(expr) {
		return nil, nil
	}
	atom := expr.(*Atom)
	act, err := EvalAction(atom.Rep, state2.Domain)
	if err != nil {
		return nil, err
	}
	state1, err := EvalState(checkPrecond, atom.Terms[0], state2.Domain)
	if err != nil {
		return nil, err
	}
	return DecomposeAction(checkPrecond, cfg, state2, state1, act)
}

// DecomposeAction decomposes an already-resolved action between two states.
// This is the same implementation as DecomposeActionApp after evaluating the
// action and pre-state expression, but it also supports action provenance that
// is not addressable by a module action name, such as env actions.
func DecomposeAction(checkPrecond bool, cfg *IvyUtilsConfig, state2, state1 *InterpState, act ActionsAction) (*InterpState, error) {
	// Compute update
	ctx := &UpdateContext{
		Domain:          state1.Domain,
		PVars:           state1.InScope,
		ActCfg:          state1.Domain.Cfg.ActCfg,
		Instantiator:    state1.Domain.Instantiator,
		CheckUnprovable: state1.Domain.Cfg.OnlyCheckUnprovable,
		CheckedAssert:   state1.Domain.Cfg.CheckLineno,
		GetAction: func(name string) ActionsAction {
			if state1.Domain != nil {
				if a, ok := state1.Domain.Actions.Get2(name); ok {
					if act, ok2 := a.(ActionsAction); ok2 {
						return act
					}
				}
			}
			return nil
		},
	}
	xtracer.Trace("phase4 calling IntUpdate type=%s", ActionTypeName(act))
	upd := IntUpdate(act, ctx)
	if upd == nil {
		upd = NullUpdate()
	}

	// Use decomposition: try each decomposition path. Python passes the full
	// state value triples, not just the visible formulas.
	comps := DecomposeWithUpdate(ctx, act, stateValueToUpdate(state1.Value()), stateValueToUpdate(state2.Value()), false)
	bg := state1.Domain.BackgroundTheory(state1.InScope)

	for _, comp := range comps {
		// Compute updates for each action in the decomposition
		upds := make([]*Update, len(comp.Actions))
		for i, subAct := range comp.Actions {
			xtracer.Trace("phase4 calling IntUpdate decomp type=%s", ActionTypeName(subAct))
			subUpd := IntUpdate(subAct, ctx)
			if subUpd == nil {
				subUpd = NullUpdate()
			}
			upds[i] = subUpd
		}

		// Build a history from pre-state
		h := NewHistory(cfg, comp.Pre)

		// Forward-step through each update
		for _, upd := range upds {
			h = h.ForwardStep(bg, upd, nil)
		}

		// Assume the post-state
		if comp.Post != nil {
			h = h.Assume(comp.Post.TR)
		}

		// Check satisfiability
		bmcRes := h.Satisfy(bg)
		if bmcRes == nil {
			continue
		}

		// Build per-step states from the satisfying path.
		// Python ivy_interp.py:504-515
		var states []*InterpState
		for i, value := range bmcRes.Path {
			state := NewStateFromClauses(state1.Domain, value.TR)
			if i != 0 {
				state.Expr = InterpActionApp(state1.AstCfg(), comp.Actions[i-1].Name(), WrapState(states[len(states)-1]))
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
func StateImpliesFormula(state *InterpState, fmla Expr) bool {
	axioms := state.Domain.BackgroundTheory(state.InScope)
	combined := AndClausesTyped(state.Clauses, axioms)
	ok, _ := ClausesImplyFormulaCex(state.Domain, combined, fmla)
	return ok
}

// --- NewHistory (already implemented as NewHistoryFromState) ---

// --- EvalAssertRhs ---

// EvalAssertRhs evaluates the right-hand side of a state assertion.
// If the RHS is not already an RME, wraps it in one.
// Corresponds to Python's eval_assert_rhs.
func EvalAssertRhs(checkPrecond bool, rhs interface{}, domain *Module) (*InterpState, error) {
	// Python:
	//   if not isinstance(rhs, ivy_actions.RME):
	//       rhs = ivy_actions.RME(And(), None, rhs)
	//   with ivy_actions.ActionContext(domain):
	//       return eval_state(rhs)
	rmeVal, ok := rhs.(*LogicRME)
	if !ok {
		// Wrap non-RME in RME(And(), nil, rhs)
		var rhsNode Expr
		if n, ok2 := rhs.(Expr); ok2 {
			rhsNode = n
		} else if n, ok2 := rhs.(Node); ok2 {
			// For ast.Node, evaluate directly within ActionContext
			ctx := NewActionContext(domain)
			_ = ctx
			return EvalState(checkPrecond, n, domain)
		}
		rmeVal = NewRME(&LogicAnd{}, nil, rhsNode)
	}

	// Evaluate within an ActionContext
	ctx := NewActionContext(domain)
	_ = ctx
	// Convert RME to state: the RME's ensures formula becomes the state constraint
	if rmeVal.Ensures != nil {
		cls := FormulaToClauses(rmeVal.Ensures, nil)
		return NewStateFromClauses(domain, cls), nil
	}
	return NewStateFromClauses(domain, TrueClauses(nil)), nil
}

// --- EvalStateOrder ---

// EvalStateOrder evaluates a state ordering relation between two expressions.
// Returns true if lhs ⊆ rhs.
// Corresponds to Python's eval_state_order.
func EvalStateOrder(checkPrecond bool, lhs, rhs Node, mod *Module) (bool, error) {
	state, err := EvalState(checkPrecond, lhs, mod)
	if err != nil {
		return false, err
	}
	if IsInterpStateJoin(rhs) {
		or := rhs.(*Or)
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
func CheckStateAssertion(checkPrecond bool, state *InterpState, assertion *LabeledFormula) bool {
	if state.Label == "" {
		return true
	}
	// Check if the assertion's label matches the state's label
	if assertion.Label == nil {
		return true
	}
	labelSym, ok := assertion.Label.(*Const)
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
func GetStateAssertions(checkPrecond bool, state *InterpState, mod *Module) *Clauses {
	if state.Label == "" {
		return nil
	}
	res := TrueClauses(nil)
	for _, assertion := range mod.Assertions {
		if assertion.Label == nil {
			continue
		}
		labelSym, ok := assertion.Label.(*Const)
		if !ok {
			continue
		}
		if state.Label == labelSym.Name {
			rhsState, err := EvalAssertRhs(checkPrecond, assertion.Formula, state.Domain)
			if err != nil {
				continue
			}
			res = AndClausesTyped(res, rhsState.Clauses)
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
	Sort   Sort
	Values []Expr
}

// UniverseConstraint creates clauses constraining universe values.
// If the state has universe data (from model finding), generates
// equality constraints for each sort.
// Corresponds to Python's universe_constraint.
func UniverseConstraint(state *InterpState) *Clauses {
	if state.Universe == nil {
		return TrueClauses(nil)
	}
	// Universe is a map from sort -> []values.
	// Accept both the legacy map[lg.Sort][]lg.Expr and the
	// safe []SortUniverse (which avoids pointer-equality misses
	// on lg.Sort interface map keys).
	type sortEntry struct {
		sort   Sort
		values []Expr
	}
	var entries []sortEntry
	switch um := state.Universe.(type) {
	case []SortUniverse:
		for _, su := range um {
			entries = append(entries, sortEntry{su.Sort, su.Values})
		}
	case map[Sort][]Expr:
		for s, vals := range um {
			entries = append(entries, sortEntry{s, vals})
		}
	default:
		return TrueClauses(nil)
	}
	var fmlas []Expr
	for _, e := range entries {
		s, values := e.sort, e.values
		if len(values) == 0 {
			continue
		}
		x, _ := NewVariable("X", s)
		var disjuncts []Expr
		for _, v := range values {
			disjuncts = append(disjuncts, &Eq{T1: x, T2: v})
		}
		fmla := IvyForAll([]*LogicVariable{x}, &LogicOr{Terms: disjuncts})
		fmlas = append(fmlas, fmla)
	}
	if len(fmlas) == 0 {
		return TrueClauses(nil)
	}
	return NewClauses(fmlas, nil, nil)
}
