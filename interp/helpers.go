package interp

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/solver"
	tr "github.com/glycerine/goivy/transrel"
	"github.com/glycerine/goivy/z3bridge"
)

// ---------------------------------------------------------------------------
// UnsatCoreWithInterpolant
// ---------------------------------------------------------------------------

// UnsatCoreWithInterpolant is returned when a reverse update finds that the
// predecessor cannot reach the given clauses. The Core and Itp fields contain
// the unsat core and interpolant respectively.
// Corresponds to Python's UnsatCoreWithInterpolant exception (ivy_interp.py:219-222).
type UnsatCoreWithInterpolant struct {
	Core *co.Clauses
	Itp  *co.Clauses
}

func (e *UnsatCoreWithInterpolant) Error() string {
	return "unsat core with interpolant"
}

// functionsToInterpreted converts a Functions map (map[string]lg.Sort) to
// the map[string]bool expected by transrel interpolation functions.
func functionsToInterpreted(functions map[string]lg.Sort) map[string]bool {
	if functions == nil {
		return nil
	}
	m := make(map[string]bool, len(functions))
	for k := range functions {
		m[k] = true
	}
	return m
}

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

func (fa *FailAction) ActionArgs() []lg.Expr {
	return fa.Inner.ActionArgs()
}

func (fa *FailAction) ActionClone(args []lg.Expr) actions.Action {
	return &FailAction{
		ActionBase: fa.ActionBase,
		Inner:      fa.Inner.ActionClone(args),
	}
}

func (fa *FailAction) IterCalls() []string {
	return fa.Inner.IterCalls()
}

func (fa *FailAction) IterSubactions() []actions.Action {
	return []actions.Action{fa}
}

func (fa *FailAction) Decompose() [][]actions.Action {
	return [][]actions.Action{{fa}}
}

// --- ast.Node + lg.Expr methods for FailAction ---

func (fa *FailAction) Args() []ast.Node {
	args := fa.ActionArgs()
	nodes := make([]ast.Node, len(args))
	for i, e := range args {
		nodes[i] = e
	}
	return nodes
}
func (fa *FailAction) Clone(args []ast.Node) ast.Node {
	exprs := make([]lg.Expr, len(args))
	for i, n := range args {
		exprs[i] = n.(lg.Expr)
	}
	return fa.ActionClone(exprs).(ast.Node)
}
func (fa *FailAction) Children() []lg.Expr          { return fa.ActionArgs() }
func (fa *FailAction) NodeSort() lg.Sort            { return lg.ActionS }
func (fa *FailAction) Equal(other lg.Expr) bool     { return fa.Sexp() == other.Sexp() }
func (fa *FailAction) GetAstConfig() *ast.AstConfig { return nil }
func (fa *FailAction) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(FailAction inner:%v)", fa.Inner.Sexp()))
}
func (fa *FailAction) Canon() iu.Canonical { return iu.Canonical(fa.Sexp()) }

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
// Corresponds to Python's reverse() in ivy_interp.py.
func Reverse(state *State, clauses *co.Clauses) (*co.Clauses, error) {
	if state.Pred() == nil || state.Update() == nil {
		return nil, fmt.Errorf("Reverse: cannot reverse state without predecessor and update")
	}
	if clauses == nil {
		clauses = state.Clauses
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	revImage := tr.ReverseImage(clauses.ToFormula(), axioms.ToFormula(), state.Update())
	revClauses := co.FormulaToClauses(revImage, clauses.Annot)
	return co.AndClausesTyped(revClauses, axioms), nil
}

// ReverseUpdateConcreteClauses reverses an update concretely. If the
// forward interpolant check finds that the predecessor cannot reach
// the given clauses, it returns UnsatCoreWithInterpolant. Otherwise
// returns the reverse image conjoined with axioms.
//
// Corresponds to Python's reverse_update_concrete_clauses() in ivy_interp.py.
func ReverseUpdateConcreteClauses(state *State, clauses *co.Clauses) (*co.Clauses, error) {
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
	fi := tr.ForwardInterpolant(state.Pred().Clauses, state.Update(), clauses, axioms, interpreted)
	if fi != nil {
		return nil, &UnsatCoreWithInterpolant{Core: fi.Core, Itp: fi.Itp}
	}

	// Compute reverse image: this is the concrete pre-image.
	revImage := tr.ReverseImage(clauses.ToFormula(), axioms.ToFormula(), state.Update())
	revClauses := co.FormulaToClauses(revImage, clauses.Annot)
	return co.AndClausesTyped(revClauses, axioms), nil
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
// under-approximation. If reachable, it adds a reachable state to the
// under-approximation and returns it. Otherwise returns nil.
//
// Corresponds to Python's reach_state() in ivy_interp.py.
func ReachState(state *State, clauses *co.Clauses) *State {
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
	img := tr.ForwardImage(pre.ToFormula(), axioms.ToFormula(), state.Update())
	imgClauses := co.AndClausesTyped(
		co.FormulaToClauses(img, nil),
		axioms,
		clauses,
	)
	// Check satisfiability of the forward image conjoined with the target.
	// If SAT, a reachable state exists.
	t := z3bridge.NewTranslator()
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
func ReachStateFromPred(state *State, clauses *co.Clauses) (*State, error) {
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
		ri := tr.ReverseInterpolantCase(clauses, state.Update(), pre, axioms, interpreted)
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
func UndecidedConjectures(state *State) []*co.Clauses {
	conjs := state.Conjs()
	if len(conjs) == 0 {
		return nil
	}
	axioms := state.Domain.BackgroundTheory(state.InScope)
	premise := co.AndClausesTyped(state.Clauses, axioms)
	premiseFmla := premise.ToFormula()

	var undecided []*co.Clauses
	for _, c := range conjs {
		t := z3bridge.NewTranslator()
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
func FilterConjectures(state *State, model *co.Clauses) []*co.Clauses {
	conjs := state.Conjs()
	if len(conjs) == 0 {
		return nil
	}
	modelFmla := model.ToFormula()
	var keep []*co.Clauses
	var lose []*co.Clauses
	for _, c := range conjs {
		t := z3bridge.NewTranslator()
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
// Returns (core, interpolant, ok). If ok is true, the interpolant
// is appended to the state's conjectures.
//
// Corresponds to Python's case_conjecture() in ivy_interp.py.
func CaseConjecture(state *State, clauses *co.Clauses) (interface{}, interface{}, bool) {
	pre := JoinUnders(state)
	axioms := state.Domain.BackgroundTheory(state.InScope)

	// Check if the under-approximation (pre) implies the clauses.
	// If pre AND NOT clauses is UNSAT, then pre implies clauses and
	// there's no separating conjecture to find.
	premiseFmla := co.AndClausesTyped(pre, axioms).ToFormula()
	clausesFmla := clauses.ToFormula()

	t := z3bridge.NewTranslator()
	implied, err := t.Implies(premiseFmla, clausesFmla)
	if err != nil {
		return nil, nil, false
	}
	if implied {
		// pre already implies clauses; no separator needed.
		return nil, nil, false
	}

	// Check if NOT clauses AND pre is satisfiable (to find a model
	// that separates them). If the negation conjoined with pre is
	// SAT, we have found something the under-approximation satisfies
	// but clauses does not — use the negation of clauses as a conjecture.
	negClauses := co.FormulaToClauses(&lg.Not{Body: clausesFmla}, nil)
	interp := negClauses
	state.SetConjs(append(state.Conjs(), interp))
	return nil, interp, true
}

// ---------------------------------------------------------------------------
// Diagram
// ---------------------------------------------------------------------------

// Diagram returns the diagram of a single model of clauses in the
// given state, or nil if the clauses are unsatisfiable.
//
// Corresponds to Python's diagram() in ivy_interp.py.
func Diagram(state *State, clauses *co.Clauses, implied *co.Clauses, extraAxioms *co.Clauses, weaken, upwardClose bool) *co.Clauses {
	axioms := state.Domain.BackgroundTheory(state.InScope)
	if extraAxioms != nil {
		axioms = co.AndClausesTyped(axioms, extraAxioms)
	}

	// Use solver to extract a minimal model diagram.
	// Python: ivy_interp.py:337-345 calls clauses_model_to_diagram.
	slv := solver.New()
	isSkolem := func(c *lg.Symbol) bool {
		return tr.IsSkolem(c.Name)
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
func NewHistoryFromState(cfg *iu.IvyUtilsConfig, state *State) *tr.History {
	return tr.NewHistory(cfg, tr.PureState(state.ToFormula()))
}

// HistoryForwardStep advances a history by one step through the
// state's update.
func HistoryForwardStep(history *tr.History, state *State) *tr.History {
	var actionNode lg.Expr
	if state.Expr != nil && IsActionApp(state.Expr) {
		atom := state.Expr.(*ast.Atom)
		actionNode = lg.NewSymbol(atom.Rep, lg.Boolean)
	}
	pred := state.Pred()
	if pred == nil {
		return history
	}
	bg := pred.Domain.BackgroundTheory(pred.InScope).ToFormula()
	return history.ForwardStep(bg, state.Update(), actionNode)
}

// HistorySatisfy checks whether a history is satisfiable in the
// given state's background theory. Returns the SatisfyResult (universes
// and path) if satisfiable, or nil if unsatisfiable.
//
// Corresponds to Python's history_satisfy() in ivy_interp.py (lines 593-598).
func HistorySatisfy(history *tr.History, state *State) *tr.SatisfyResult {
	axioms := state.Domain.BackgroundTheory(state.InScope)
	return history.Satisfy(axioms.ToFormula())
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
	// Go's Relations is map[string]lg.Sort; we add the relation's sort if available.
	newRelations := make(map[string]lg.Sort, len(origRelations))
	for k, v := range origRelations {
		newRelations[k] = v
	}

	// Extract concept space body formulas for type checking
	var formulas []interface{}
	for _, cs := range mod.ConceptSpaces {
		// cs.Label is the compiled relation (lg.Expr), cs.Body is the body
		if cs.Label != nil {
			// Extract name and sort from the relation expression
			switch r := cs.Label.(type) {
			case *lg.Apply:
				if sym, ok := r.Func.(*lg.Symbol); ok {
					newRelations[sym.Name] = sym.CSort
				}
			case *lg.Symbol:
				newRelations[r.Name] = r.CSort
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
		t := z3bridge.NewTranslator()
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
func GetPropertyContext(mod *module.Module, prop *ast.LabeledFormula) *co.Clauses {
	res := co.TrueClauses(nil)
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
			res = co.AndClausesTyped(res, co.FormulaToClauses(x.Formula.(lg.Expr), nil))
		}
	}
	return res
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
				return []*ast.Atom{ActionApp(pre.AstCfg(), atom.Rep, WrapState(pre))}
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
func FailExpr(cfg *ast.AstConfig, expr *ast.Atom) *ast.Atom {
	return ActionApp(cfg, "fail_"+expr.Rep, expr.Terms[0])
}

// Ensure unused imports don't cause errors.
var _ = fmt.Sprintf
var _ actions.Action
var _ *module.Module
var _ *tr.Update
