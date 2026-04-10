package interp

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// ---------------------------------------------------------------------------
// UnsatCoreWithInterpolant
// ---------------------------------------------------------------------------

// UnsatCoreWithInterpolant is returned when a reverse update finds that the
// predecessor cannot reach the given clauses. The Core and Itp fields contain
// the unsat core and interpolant respectively.
// Corresponds to Python's UnsatCoreWithInterpolant exception (ivy_interp.py:219-222).
type UnsatCoreWithInterpolant struct {
	Core *module.Clauses
	Itp  *module.Clauses
}

func (e *UnsatCoreWithInterpolant) Error() string {
	return "unsat core with interpolant"
}

// functionsToInterpreted converts a Functions InsMap to
// the map[string]bool expected by transrel interpolation functions.
func functionsToInterpreted(functions *iu.InsMap[string, lg.Sort]) map[string]bool {
	if functions == nil {
		return nil
	}
	m := make(map[string]bool, functions.Len())
	for k := range functions.All() {
		m[k] = true
	}
	return m
}

// ---------------------------------------------------------------------------
// Reverse image helpers
// ---------------------------------------------------------------------------

// Reverse computes the reverse image (concrete pre-state) of a state
// through its update. If clauses is nil, the state's own clauses are used.
//
// Corresponds to Python's reverse() in ivy_interp.py.
func Reverse(state *State, clauses *module.Clauses) (*module.Clauses, error) {
	if state.Pred() == nil || state.Update() == nil {
		return nil, fmt.Errorf("Reverse: cannot reverse state without predecessor and update")
	}
	if clauses == nil {
		clauses = state.Clauses
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	revClauses := actions.ReverseImage(clauses, axioms, state.Update())
	return module.AndClausesTyped(revClauses, axioms), nil
}

// ReverseUpdateConcreteClauses reverses an update concretely. If the
// forward interpolant check finds that the predecessor cannot reach
// the given clauses, it returns UnsatCoreWithInterpolant. Otherwise
// returns the reverse image conjoined with axioms.
//
// Corresponds to Python's reverse_update_concrete_clauses() in ivy_interp.py.
func ReverseUpdateConcreteClauses(state *State, clauses *module.Clauses) (*module.Clauses, error) {
	if state.Pred() == nil || state.Update() == nil {
		return nil, fmt.Errorf("ReverseUpdateConcreteClauses: no predecessor or update")
	}
	if clauses == nil {
		clauses = state.Clauses
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	interpreted := functionsToInterpreted(state.Domain.Functions)

	// Check forward interpolant: if the predecessor cannot reach the
	// given clauses, return UnsatCoreWithInterpolant.
	// Python: ivy_interp.py:234-236
	fi := actions.ForwardInterpolant(state.Pred().Clauses, state.Update(), clauses, axioms, interpreted)
	if fi != nil {
		return nil, &UnsatCoreWithInterpolant{Core: fi.Core, Itp: fi.Itp}
	}

	// Compute reverse image: this is the concrete pre-image.
	revClauses := actions.ReverseImage(clauses, axioms, state.Update())
	return module.AndClausesTyped(revClauses, axioms), nil
}

// ---------------------------------------------------------------------------
// Reach state helpers
// ---------------------------------------------------------------------------

// JoinUnders computes the join (disjunction) of all under-approximation
// states. Returns TrueClauses if there are no under-approximations.
func JoinUnders(state *State) *module.Clauses {
	unders := state.Unders()
	if len(unders) == 0 {
		return module.TrueClauses(nil)
	}
	clauses := make([]*module.Clauses, len(unders))
	for i, u := range unders {
		clauses[i] = u.Clauses
	}
	return module.OrClausesTyped(clauses...)
}

// AddUnder adds an under-approximation state to the target state.
func AddUnder(state *State, clauses *module.Clauses, pred *State, universe interface{}) *State {
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
// under-approximation. If reachable, it adds a reachable state to the
// under-approximation and returns it. Otherwise returns nil.
//
// Corresponds to Python's reach_state() in ivy_interp.py.
func ReachState(state *State, clauses *module.Clauses) *State {
	if state.Pred() == nil || state.Update() == nil {
		return nil
	}
	pre := JoinUnders(state.Pred())
	if clauses == nil {
		clauses = state.Clauses
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	// Compute the forward image from the predecessor's under-approximation
	// through the state's update, then conjoin with the target clauses.
	imgClauses := module.AndClausesTyped(
		actions.ForwardImage(pre, axioms, state.Update()),
		axioms,
		clauses,
	)
	// Check satisfiability of the forward image conjoined with the target.
	// If SAT, a reachable state exists.
	solver := z3bridge.NewSolver(nil, nil)
	t := solver.NewTranslator()
	defer t.Close()
	result, err := t.IsSat(imgClauses.ToFormula())
	if err != nil || result != z3bridge.Sat {
		return nil
	}
	// The image is satisfiable, so we can reach the target. Use the image
	// clauses as the under-approximation of the reached state.
	return AddUnder(state, imgClauses, nil, nil)
}

// ReachStateFromPred attempts to reach a state from its predecessor's
// under-approximation. If reachable, updates the under-approximation
// and returns the reachable state. If not reachable, returns nil.
//
// Corresponds to Python's reach_state_from_pred() in ivy_interp.py.
func ReachStateFromPred(state *State, clauses *module.Clauses) (*State, error) {
	post := ReachState(state, clauses)
	if post != nil {
		return post, nil
	}
	// If not reachable, compute a reverse interpolant as abductive inference.
	// Python: ivy_interp.py:316-318
	if clauses == nil {
		clauses = state.Clauses
	}
	if state.Pred() != nil && state.Update() != nil {
		axioms := state.Domain.BackgroundTheory(state.InScope)
		interpreted := functionsToInterpreted(state.Domain.Functions)
		pre := JoinUnders(state.Pred())
		ri := actions.ReverseInterpolantCase(clauses, state.Update(), pre, axioms, interpreted)
		if ri != nil {
			return nil, &UnsatCoreWithInterpolant{Core: ri.Core, Itp: ri.Itp}
		}
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Conjecture helpers
// ---------------------------------------------------------------------------

// UndecidedConjectures returns conjectures of state1 that are not
// implied by the state's clauses and background theory.
//
// Corresponds to Python's undecided_conjectures() in ivy_interp.py.
func UndecidedConjectures(state *State) []*module.Clauses {
	conjs := state.Conjs()
	if len(conjs) == 0 {
		return nil
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	premise := module.AndClausesTyped(state.Clauses, axioms)
	premiseFmla := premise.ToFormula()

	var undecided []*module.Clauses
	for _, c := range conjs {
		solver := z3bridge.NewSolver(nil, nil)
		t := solver.NewTranslator()

		implied, err := t.Implies(premiseFmla, c.ToFormula())
		if err != nil || !implied {
			undecided = append(undecided, c)
		}
	}
	return undecided
}

// FilterConjectures partitions the state's conjectures into those
// implied by the model (kept) and those not implied (lost).
//
// Corresponds to Python's filter_conjectures() in ivy_interp.py.
func FilterConjectures(state *State, model *module.Clauses) []*module.Clauses {
	conjs := state.Conjs()
	if len(conjs) == 0 {
		return nil
	}
	modelFmla := model.ToFormula()
	var keep []*module.Clauses
	var lose []*module.Clauses
	for _, c := range conjs {
		solver := z3bridge.NewSolver(nil, nil)
		t := solver.NewTranslator()

		implied, err := t.Implies(modelFmla, c.ToFormula())
		if err == nil && implied {
			keep = append(keep, c)
		} else {
			lose = append(lose, c)
		}
	}
	state.SetConjs(keep)
	return lose
}

// CaseConjecture conjectures a separator between the state's
// under-approximation and the given clauses. The separator must be
// true of all models of the under-approximation and false in at
// least one model of clauses.
//
// If a separator is found, returns the InterpolantResult and appends
// the interpolant to the state's conjectures.
//
// Faithful port of Python case_conjecture (ivy_interp.py:325-337):
//
//	def case_conjecture(state,clauses):
//	    pre = join_unders(state)
//	    axioms = state.domain.background_theory(state.in_scope)
//	    ri = interpolant_case(pre,clauses,axioms,state.domain.functions)
//	    if ri != None:
//	        core,interp = ri
//	        state.conjs.append(interp)
//	    return ri
func CaseConjecture(state *State, clauses *module.Clauses) *actions.InterpolantResult {
	pre := JoinUnders(state)
	axioms := state.Domain.BackgroundTheory(state.InScope)
	interpreted := functionsToInterpreted(state.Domain.Functions)
	ri := actions.InterpolantCase(pre, clauses, axioms, interpreted)
	if ri != nil {
		state.SetConjs(append(state.Conjs(), ri.Itp))
	}
	return ri
}

// ---------------------------------------------------------------------------
// Diagram
// ---------------------------------------------------------------------------

// Diagram returns the diagram of a single model of clauses in the
// given state, or nil if the clauses are unsatisfiable.
//
// Corresponds to Python's diagram() in ivy_interp.py.
func Diagram(state *State, clauses *module.Clauses, implied *module.Clauses, extraAxioms *module.Clauses, weaken, upwardClose bool) *module.Clauses {
	axioms := state.Domain.BackgroundTheory(state.InScope)
	if extraAxioms != nil {
		axioms = module.AndClausesTyped(axioms, extraAxioms)
	}

	// Use solver to extract a minimal model diagram.
	// Python: ivy_interp.py:337-345 calls clauses_model_to_diagram.
	slv := z3bridge.NewSolver(state.Domain.Sig, nil)
	isSkolem := func(c *lg.Const) bool {
		return actions.IsSkolem(c.Name)
	}
	diag, err := slv.ClausesModelToDiagram(clauses, isSkolem, axioms)
	if err != nil || diag == nil {
		return nil
	}
	return diag
}

// ---------------------------------------------------------------------------
// History (BMC) helpers
// ---------------------------------------------------------------------------

// NewHistory creates a History from a state's value.
func NewHistoryFromState(cfg *iu.IvyUtilsConfig, state *State) *actions.History {
	return actions.NewHistory(cfg, actions.PureState(state.ToFormula()))
}

// HistoryForwardStep advances a history by one step through the
// state's update.
func HistoryForwardStep(history *actions.History, state *State) *actions.History {
	var actionNode lg.Expr
	if state.Expr != nil && IsActionApp(state.Expr) {
		atom := state.Expr.(*ast.Atom)
		actionNode = lg.NewConst(atom.Rep, lg.Boolean)
	}
	pred := state.Pred()
	if pred == nil {
		return history
	}
	bg := pred.Domain.BackgroundTheory(pred.InScope)
	return history.ForwardStep(bg, state.Update(), actionNode)
}

// HistorySatisfy checks whether a history is satisfiable in the
// given state's background theory. Returns the SatisfyResult (universes
// and path) if satisfiable, or nil if unsatisfiable.
//
// Corresponds to Python's history_satisfy() in ivy_interp.py (lines 593-598).
func HistorySatisfy(history *actions.History, state *State) *actions.SatisfyResult {
	axioms := state.Domain.BackgroundTheory(state.InScope)
	return history.Satisfy(axioms)
}

// ---------------------------------------------------------------------------
// Module helper functions (replacing Python monkey-patches)
// ---------------------------------------------------------------------------

// ModuleNewState creates a new State from clauses, adding an empty
// annotation if none is present. Mirrors module_new_state.
func ModuleNewState(mod *module.Module, clauses *module.Clauses) *State {
	return NewStateFromClauses(mod, clauses)
}

// ModuleNewStateWithValue creates a State from a full value triple.
// Mirrors module_new_state_with_value.
func ModuleNewStateWithValue(mod *module.Module, value *StateValue) *State {
	return NewState(mod, value, nil, "")
}

// ModuleTypeCheck type-checks the module's axioms and concept spaces.
//
// Corresponds to Python's module_type_check() in ivy_interp.py:
//
//	type_check_list(self, self.axioms)
//	self.type_check_concepts()
func ModuleTypeCheck(mod *module.Module) error {
	// Python: type_check_list(self, self.axioms)
	axiomExprs := make([]interface{}, 0, len(mod.LabeledAxioms))
	for _, ax := range mod.LabeledAxioms {
		if ax != nil && ax.Formula != nil {
			if expr, ok := ax.Formula.(lg.Expr); ok {
				axiomExprs = append(axiomExprs, expr)
			}
		}
	}
	if err := TypeCheckList(mod, axiomExprs); err != nil {
		return err
	}
	return ModuleTypeCheckConcepts(mod)
}

// ModuleTypeCheckConcepts type-checks concept spaces.
//
// Corresponds to Python's module_type_check_concepts() in ivy_interp.py:
//
//	relations = self.relations
//	self.relations = dict(iter(relations.items()))
//	self.relations.update((x.rep, len(x.args)) for x, y in self.concept_spaces)
//	type_check_list(self, [y for x, y in self.concept_spaces])
//	self.relations = relations
func ModuleTypeCheckConcepts(mod *module.Module) error {
	if len(mod.ConceptSpaces) == 0 {
		return nil
	}
	// Save original relations
	origRelations := mod.Relations

	// Copy and extend with concept space arities.
	// Python temporarily adds (x.rep, len(x.args)) for each concept space relation.
	// Go's Relations is *iu.InsMap[string, lg.Sort]; we add the relation's sort if available.
	newRelations := iu.NewInsMap[string, lg.Sort]()
	for k, v := range origRelations.All() {
		newRelations.Set(k, v)
	}

	// Extract concept space body formulas for type checking
	var formulas []interface{}
	for _, cs := range mod.ConceptSpaces {
		// cs.Label is the compiled relation (lg.Expr), cs.Body is the body
		if cs.Label != nil {
			// Extract name and sort from the relation expression
			switch r := cs.Label.(type) {
			case *lg.Apply:
				if sym, ok := r.Func.(*lg.Const); ok {
					newRelations.Set(sym.Name, sym.CSort)
				}
			case *lg.Const:
				newRelations.Set(r.Name, r.CSort)
			}
		}
		if cs.Body != nil {
			formulas = append(formulas, cs.Body)
		}
	}

	mod.Relations = newRelations
	err := TypeCheckList(mod, formulas)
	mod.Relations = origRelations
	return err
}

// ---------------------------------------------------------------------------
// Property checking
// ---------------------------------------------------------------------------

// FalseProperties returns the labeled properties of the module that
// are false (not implied by the background theory).
//
// Corresponds to Python's false_properties() in ivy_interp.py.
func FalseProperties(mod *module.Module) []*ast.LabeledFormula {
	axioms := mod.BackgroundTheory(nil)
	axiomsFmla := axioms.ToFormula()

	// Build a subgoal map for properties that serve as subgoals.
	subgoalMap := make(map[int64]bool)
	for _, sg := range mod.Subgoals {
		if sg.Formula != nil {
			subgoalMap[sg.Formula.ID] = true
		}
	}

	var falseProps []*ast.LabeledFormula
	// Accumulate: properties that are subgoals are assumed (their
	// truth is accumulated for subsequent checks). Non-subgoal
	// properties are asserted.
	premise := axiomsFmla
	for _, prop := range mod.LabeledProps {
		if prop.IsTemporal() || prop.Formula == nil {
			continue
		}
		if subgoalMap[prop.ID] {
			// Subgoal: assume it for subsequent checks.
			premise = &lg.And{Terms: []lg.Expr{premise, prop.Formula.(lg.Expr)}}
			continue
		}
		// Assert: check if axioms (plus accumulated subgoals) imply this property.
		solver := z3bridge.NewSolver(nil, nil)
		t := solver.NewTranslator()

		implied, err := t.Implies(premise, prop.Formula.(lg.Expr))
		if err != nil || !implied {
			falseProps = append(falseProps, prop)
		}
	}
	return falseProps
}

// GetPropertyContext returns the accumulated context (conjunction of
// prior subgoal properties) for a given property.
//
// Corresponds to Python's get_property_context() in ivy_interp.py.
func GetPropertyContext(mod *module.Module, prop *ast.LabeledFormula) *module.Clauses {
	res := module.TrueClauses(nil)
	// Build subgoal map.
	subgoalMap := make(map[int64]bool)
	for _, sg := range mod.Subgoals {
		if sg.Formula != nil {
			subgoalMap[sg.Formula.ID] = true
		}
	}
	for _, x := range mod.LabeledProps {
		if prop != nil && x.ID == prop.ID {
			break
		}
		if subgoalMap[x.ID] && x.Formula != nil {
			res = module.AndClausesTyped(res, module.FormulaToClauses(x.Formula.(lg.Expr), nil))
		}
	}
	return res
}

// ---------------------------------------------------------------------------
// Additional eval helpers
// ---------------------------------------------------------------------------

// EvalStateFacts evaluates an expression tree, skipping state symbols
// (returning nil for them). Mirrors eval_state_facts.
func EvalStateFacts(checkPrecond bool, expr ast.Node, mod *module.Module) (*State, error) {
	if IsStateJoin(expr) {
		or := expr.(*ast.Or)
		var result *State
		for _, term := range or.Terms {
			s, err := EvalStateFacts(checkPrecond, term, mod)
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
		s, err := EvalStateFacts(checkPrecond, atom.Terms[0], mod)
		if err != nil {
			return nil, err
		}
		if s == nil {
			return nil, nil
		}
		return ApplyAction(checkPrecond, expr, atom.Rep, act, s)
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
				return []*ast.Atom{ActionApp(pre.AstCfg(), atom.Rep, WrapState(pre))}
			}
		}
	}
	return nil
}

// TopAlpha resets a state's clauses to TrueClauses.
func TopAlpha(state *State) {
	state.Clauses = module.TrueClauses(nil)
}

// FailExpr constructs a fail expression from an action application.
func FailExpr(cfg *ast.AstConfig, expr *ast.Atom) *ast.Atom {
	return ActionApp(cfg, "fail_"+expr.Rep, expr.Terms[0])
}

// Ensure unused imports don't cause errors.
var _ = fmt.Sprintf
var _ actions.Action
var _ *module.Module
var _ *actions.Update
