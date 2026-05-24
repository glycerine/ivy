package goivy

import (
	"fmt"
)

// ---------------------------------------------------------------------------
// UnsatCoreWithInterpolant
// ---------------------------------------------------------------------------

// UnsatCoreWithInterpolant is returned when a reverse update finds that the
// predecessor cannot reach the given clauses. The Core and Itp fields contain
// the unsat core and interpolant respectively.
// Corresponds to Python's UnsatCoreWithInterpolant exception (ivy_interp.py:219-222).
type UnsatCoreWithInterpolant struct {
	Core *Clauses
	Itp  *Clauses
}

func (e *UnsatCoreWithInterpolant) Error() string {
	return "unsat core with interpolant"
}

// functionsToInterpreted converts a Functions InsMap to
// the map[string]bool expected by transrel interpolation functions.
func functionsToInterpreted(functions *InsMap[NodeKey, Sort]) map[string]bool {
	if functions == nil {
		return nil
	}
	m := make(map[string]bool, functions.Len())
	for key := range functions.All() {
		m[SymbolNameFromKey(key)] = true
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
func Reverse(state *InterpState, clauses *Clauses) (*Clauses, error) {
	if state.Pred() == nil || state.Update() == nil {
		return nil, fmt.Errorf("Reverse: cannot reverse state without predecessor and update")
	}
	if clauses == nil {
		clauses = state.Clauses
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	revClauses := ReverseImage(clauses, axioms, state.Update())
	return AndClausesTyped(revClauses, axioms), nil
}

// ReverseUpdateConcreteClauses reverses an update concretely. If the
// forward interpolant check finds that the predecessor cannot reach
// the given clauses, it returns UnsatCoreWithInterpolant. Otherwise
// returns the reverse image conjoined with axioms.
//
// Corresponds to Python's reverse_update_concrete_clauses() in ivy_interp.py.
func ReverseUpdateConcreteClauses(state *InterpState, clauses *Clauses) (*Clauses, error) {
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
	fi := ForwardInterpolant(state.Domain, state.Pred().Clauses, state.Update(), clauses, axioms, interpreted)
	if fi != nil {
		return nil, &UnsatCoreWithInterpolant{Core: fi.Core, Itp: fi.Itp}
	}

	// Compute reverse image: this is the concrete pre-image.
	revClauses := ReverseImage(clauses, axioms, state.Update())
	return AndClausesTyped(revClauses, axioms), nil
}

// ---------------------------------------------------------------------------
// Reach state helpers
// ---------------------------------------------------------------------------

// JoinUnders computes the join (disjunction) of all under-approximation
// states. An empty under-approximation set is the empty disjunction, false.
func JoinUnders(state *InterpState) *Clauses {
	unders := state.Unders()
	clauses := make([]*Clauses, len(unders))
	for i, u := range unders {
		clauses[i] = u.Clauses
	}
	return TaggedOrClauses("__pre", clauses...)
}

// AddUnder adds an under-approximation state to the target state.
func AddUnder(state *InterpState, clauses *Clauses, pred *InterpState, universe interface{}) *InterpState {
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
func ReachState(state *InterpState, clauses *Clauses) *InterpState {
	if state.Pred() == nil || state.Update() == nil {
		return nil
	}
	predUnders := state.Pred().Unders()
	if len(predUnders) == 0 {
		return nil
	}
	pre := JoinUnders(state.Pred())
	if clauses == nil {
		clauses = state.Clauses
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	// Compute the forward image from the predecessor's under-approximation
	// through the state's update, then conjoin with the target clauses.
	imgClauses := AndClausesTyped(
		ForwardImage(pre, axioms, state.Update()),
		axioms,
		clauses,
	)
	solver := NewSolver(state.Domain, nil)
	model, err := solver.GetModelClauses(imgClauses)
	if err != nil || model == nil {
		return nil
	}

	idx := FindTrueDisjunct(pre, func(e Expr) bool {
		ok, err := solver.EvalFormula(model.Model, e)
		return err == nil && ok
	})
	if idx < 0 || idx >= len(predUnders) {
		return nil
	}

	ignore := func(c *Const) bool {
		if c == nil {
			return true
		}
		_, inRelations := state.Domain.Relations.Get2(Key(c))
		_, inFunctions := state.Domain.Functions.Get2(Key(c))
		return !inRelations && !inFunctions
	}
	post, hm, err := solver.ClausesModelToClausesWithModelAndHerbrand(imgClauses, model, ignore, false)
	if err != nil || post == nil {
		return nil
	}
	universe := map[string][]Expr{}
	if hm != nil {
		universe = hm.Universes(false)
	}
	if universe == nil {
		universe = map[string][]Expr{}
	}
	return AddUnder(state, post, predUnders[idx], universe)
}

// ReachStateFromPred attempts to reach a state from its predecessor's
// under-approximation. If reachable, updates the under-approximation
// and returns the reachable state. If not reachable, returns nil.
//
// Corresponds to Python's reach_state_from_pred() in ivy_interp.py.
func ReachStateFromPred(state *InterpState, clauses *Clauses) (*InterpState, error) {
	post := ReachState(state, clauses)
	if post != nil {
		return post, nil
	}
	return nil, fmt.Errorf("name 'pre' is not defined")
}

// ---------------------------------------------------------------------------
// Conjecture helpers
// ---------------------------------------------------------------------------

// UndecidedConjectures returns conjectures of state1 that are not
// implied by the state's clauses and background theory.
//
// Corresponds to Python's undecided_conjectures() in ivy_interp.py.
func UndecidedConjectures(state *InterpState) []*Clauses {
	conjs := state.Conjs()
	if len(conjs) == 0 {
		return nil
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	premise := AndClausesTyped(state.Clauses, axioms)
	premiseFmla := premise.ToFormula()

	var undecided []*Clauses
	for _, c := range conjs {
		solver := NewSolver(state.Domain, nil)
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
func FilterConjectures(state *InterpState, model *Clauses) []*Clauses {
	conjs := state.Conjs()
	if len(conjs) == 0 {
		return nil
	}
	modelFmla := model.ToFormula()
	var keep []*Clauses
	var lose []*Clauses
	for _, c := range conjs {
		solver := NewSolver(state.Domain, nil)
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
func CaseConjecture(state *InterpState, clauses *Clauses) *InterpolantResult {
	pre := JoinUnders(state)
	axioms := state.Domain.BackgroundTheory(state.InScope)
	interpreted := functionsToInterpreted(state.Domain.Functions)
	ri := InterpolantCase(state.Domain, pre, clauses, axioms, interpreted)
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
func Diagram(state *InterpState, clauses *Clauses, implied *Clauses, extraAxioms *Clauses, weaken, upwardClose bool) *Clauses {
	axioms := state.Domain.BackgroundTheory(state.InScope)
	if extraAxioms != nil {
		axioms = AndClausesTyped(axioms, extraAxioms)
	}

	// Use solver to extract a minimal model diagram.
	// Python: ivy_interp.py:337-345 calls clauses_model_to_diagram.
	slv := NewSolver(state.Domain, nil)
	isSkolem := func(c *Const) bool {
		return IsSkolem(c.Name)
	}
	if implied == nil {
		implied = FalseClauses(nil)
	}
	diag, err := slv.ClausesModelToDiagramFull(clauses, isSkolem, implied, nil, axioms, weaken, true, upwardClose)
	if err != nil || diag == nil {
		return nil
	}
	return diag
}

// ---------------------------------------------------------------------------
// History (BMC) helpers
// ---------------------------------------------------------------------------

// NewHistory creates a History from a state's value.
func NewHistoryFromState(cfg *IvyUtilsConfig, state *InterpState) *History {
	return NewHistory(cfg, PureState(state.ToFormula()))
}

// HistoryForwardStep advances a history by one step through the
// state's update.
func HistoryForwardStep(history *History, state *InterpState) *History {
	var actionNode Expr
	if state.Expr != nil && IsInterpActionApp(state.Expr) {
		atom := state.Expr.(*Atom)
		actionNode = NewConst(atom.Rep, Boolean)
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
func HistorySatisfy(history *History, state *InterpState) *SatisfyResult {
	axioms := state.Domain.BackgroundTheory(state.InScope)
	return history.Satisfy(axioms)
}

// ---------------------------------------------------------------------------
// Module helper functions (replacing Python monkey-patches)
// ---------------------------------------------------------------------------

// ModuleNewState creates a new State from clauses, adding an empty
// annotation if none is present. Mirrors module_new_state.
func ModuleNewState(mod *Module, clauses *Clauses) *InterpState {
	return NewStateFromClauses(mod, clauses)
}

// ModuleNewStateWithValue creates a State from a full value triple.
// Mirrors module_new_state_with_value.
func ModuleNewStateWithValue(mod *Module, value *StateValue) *InterpState {
	return NewInterpState(mod, value, nil, "")
}

// ModuleTypeCheck type-checks the module's axioms and concept spaces.
//
// Corresponds to Python's module_type_check() in ivy_interp.py:
//
//	type_check_list(self, self.axioms)
//	self.type_check_concepts()
func ModuleTypeCheck(mod *Module) error {
	// Python: type_check_list(self, self.axioms)
	axiomExprs := make([]interface{}, 0, len(mod.LabeledAxioms))
	for _, ax := range mod.LabeledAxioms {
		if ax != nil && ax.Formula != nil {
			if expr, ok := ax.Formula.(Expr); ok {
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
func ModuleTypeCheckConcepts(mod *Module) error {
	if len(mod.ConceptSpaces) == 0 {
		return nil
	}
	// Save original relations
	origRelations := mod.Relations

	// Copy and extend with concept space arities.
	// Python temporarily adds (x.rep, len(x.args)) for each concept space relation.
	// Go's Relations mirrors Python's Symbol-keyed dict.
	newRelations := NewInsMap[NodeKey, Sort]()
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
			case *Apply:
				if sym, ok := r.Func.(*Const); ok {
					newRelations.Set(Key(sym), sym.CSort)
				}
			case *Const:
				newRelations.Set(Key(r), r.CSort)
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
func FalseProperties(mod *Module) []*LabeledFormula {
	axioms := mod.BackgroundTheory(nil)
	axiomsFmla := axioms.ToFormula()

	// Build a subgoal map for properties that serve as subgoals.
	subgoalMap := make(map[int64]bool)
	for _, sg := range mod.Subgoals {
		if sg.Formula != nil {
			subgoalMap[sg.Formula.ID] = true
		}
	}

	var falseProps []*LabeledFormula
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
			premise = &LogicAnd{Terms: []Expr{premise, prop.Formula.(Expr)}}
			continue
		}
		// Assert: check if axioms (plus accumulated subgoals) imply this property.
		solver := NewSolver(mod, nil)
		t := solver.NewTranslator()

		implied, err := t.Implies(premise, prop.Formula.(Expr))
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
func GetPropertyContext(mod *Module, prop *LabeledFormula) *Clauses {
	res := TrueClauses(nil)
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
			res = AndClausesTyped(res, FormulaToClauses(x.Formula.(Expr), nil))
		}
	}
	return res
}

// ---------------------------------------------------------------------------
// Additional eval helpers
// ---------------------------------------------------------------------------

// EvalStateFacts evaluates an expression tree, skipping state symbols
// (returning nil for them). Mirrors eval_state_facts.
func EvalStateFacts(checkPrecond bool, expr Node, mod *Module) (*InterpState, error) {
	if IsInterpStateJoin(expr) {
		or := expr.(*Or)
		var result *InterpState
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
	if IsInterpActionApp(expr) {
		atom := expr.(*Atom)
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
func EvalStateActions(expr Node, pre *InterpState) []*Atom {
	if IsInterpStateJoin(expr) {
		or := expr.(*Or)
		var result []*Atom
		for _, term := range or.Terms {
			result = append(result, EvalStateActions(term, pre)...)
		}
		return result
	}
	if IsInterpActionApp(expr) {
		atom := expr.(*Atom)
		if IsStateSymbol(atom.Terms[0]) {
			inner := atom.Terms[0].(*Atom)
			if inner.Rep == pre.Label {
				return []*Atom{InterpActionApp(pre.AstCfg(), atom.Rep, WrapState(pre))}
			}
		}
	}
	return nil
}

// TopAlpha resets a state's clauses to TrueClauses.
func TopAlpha(state *InterpState) {
	state.Clauses = TrueClauses(nil)
}

// FailExpr constructs a fail expression from an action application.
func FailExpr(cfg *AstConfig, expr *Atom) *Atom {
	return InterpActionApp(cfg, "fail_"+expr.Rep, expr.Terms[0])
}

// Ensure unused imports don't cause errors.
var _ = fmt.Sprintf
var _ ActionsAction
var _ *Module
var _ *Update
