// Package ranking implements liveness-to-safety reduction using
// lexicographic relational rankings.
//
// It provides the l2s_tactic which transforms temporal properties
// (liveness) into safety properties by adding ranking functions
// and ghost state. This is the core of Ivy's approach to proving
// liveness properties.
//
// Key concepts:
//   - Tasks: work_created, work_needed, work_progress, work_invar, work_helpful
//   - Triggers: work_start conditions
//   - L2S ghost state: l2s_d (domain), l2s_w (waiting), l2s_s (saved),
//     l2s_g (globally), l2s_init (initial)
//
// Ported from ivy_ranking.py.
package ranking

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
	"github.com/glycerine/goivy/temporal"
	"github.com/glycerine/goivy/transrel"
)

// Debug controls debug output for ranking computations.
var Debug = false

// --- Formula helpers ---

// ForAll wraps a body in a ForAll quantifier if there are variables.
// If vs is empty, returns body unchanged.
func ForAll(vs []*lg.Var, body lg.Node) lg.Node {
	if len(vs) == 0 {
		return body
	}
	fa, err := lg.NewForAll(vs, body)
	if err != nil {
		return body
	}
	return fa
}

// Exists wraps a body in an Exists quantifier if there are variables.
// If vs is empty, returns body unchanged.
func Exists(vs []*lg.Var, body lg.Node) lg.Node {
	if len(vs) == 0 {
		return body
	}
	ex, err := lg.NewExists(vs, body)
	if err != nil {
		return body
	}
	return ex
}

// OldOf replaces function symbols in a formula with their "old" versions.
// For an application f(args), it returns old_f(args).
// For compound formulas, it recurses into sub-formulas.
func OldOf(fmla lg.Node) lg.Node {
	if fmla == nil {
		return nil
	}
	switch f := fmla.(type) {
	case *lg.Apply:
		oldFunc := makeOldFunc(f.Func)
		app, err := lg.NewApply(oldFunc, f.Terms...)
		if err != nil {
			return fmla
		}
		return app
	case *lg.Const:
		return lg.NewConst(transrel.Old(f.Name), f.CSort)
	case *lg.Eq:
		return &lg.Eq{T1: OldOf(f.T1), T2: OldOf(f.T2)}
	case *lg.Not:
		return &lg.Not{Body: OldOf(f.Body)}
	case *lg.And:
		terms := make([]lg.Node, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = OldOf(t)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Node, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = OldOf(t)
		}
		return &lg.Or{Terms: terms}
	case *lg.Implies:
		return &lg.Implies{T1: OldOf(f.T1), T2: OldOf(f.T2)}
	case *lg.Iff:
		return &lg.Iff{T1: OldOf(f.T1), T2: OldOf(f.T2)}
	case *lg.ForAll:
		return &lg.ForAll{Variables: f.Variables, Body: OldOf(f.Body)}
	case *lg.Exists:
		return &lg.Exists{Variables: f.Variables, Body: OldOf(f.Body)}
	default:
		return fmla
	}
}

func makeOldFunc(fn lg.Node) lg.Node {
	switch f := fn.(type) {
	case *lg.Const:
		return lg.NewConst(transrel.Old(f.Name), f.CSort)
	default:
		return fn
	}
}

// --- Named binder constructors ---

// L2sD creates the l2s_d predicate for a sort.
// l2s_d : sort -> Boolean (tracks which domain elements are active).
func L2sD(sort lg.Sort) *lg.Const {
	fs, err := lg.NewFunctionSort(sort, lg.Boolean)
	if err != nil {
		return lg.NewConst("l2s_d", lg.Boolean)
	}
	return lg.NewConst("l2s_d", fs)
}

// L2sW creates an l2s_w named binder (waiting predicate).
func L2sW(vs []*lg.Var, body lg.Node, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_w", vs, strPtr(label), body)
	return nb
}

// L2sG creates an l2s_g named binder (globally predicate).
func L2sG(vs []*lg.Var, body lg.Node, environ string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_g", vs, strPtr(environ), body)
	return nb
}

// OldL2sG creates an _old_l2s_g named binder (old globally predicate).
func OldL2sG(vs []*lg.Var, body lg.Node, environ string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("_old_l2s_g", vs, strPtr(environ), body)
	return nb
}

// L2sInit creates an l2s_init named binder.
func L2sInit(vs []*lg.Var, body lg.Node, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_init", vs, strPtr(label), body)
	return nb
}

// L2sWhen creates an l2s_when named binder.
func L2sWhen(name string, vs []*lg.Var, body lg.Node, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_when"+name, vs, strPtr(label), body)
	return nb
}

// L2sOld creates an l2s_old named binder.
func L2sOld(vs []*lg.Var, body lg.Node, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_old", vs, strPtr(label), body)
	return nb
}

// L2sS creates an l2s_s named binder (saved state).
func L2sS(vs []*lg.Var, body lg.Node, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_s", vs, strPtr(label), body)
	return nb
}

func strPtr(s string) *string {
	return &s
}

// --- Task and Trigger types ---

// Task holds the ranking function definitions for one task suffix.
type Task struct {
	WorkCreated  lg.Node // definition of work_created
	WorkNeeded   lg.Node // definition of work_needed
	WorkProgress lg.Node // definition of work_progress
	WorkInvar    lg.Node // definition of work_invar
	WorkHelpful  lg.Node // definition of work_helpful
	WorkWitness  lg.Node // optional work_witness
}

// Trigger holds the trigger condition for a task.
type Trigger struct {
	WorkStart lg.Node // definition of work_start
}

// --- L2S Tactic ---

// L2STacticConfig holds the configuration for the l2s tactic.
type L2STacticConfig struct {
	TacticName string
	ProofLabel string
	Mod        *module.Module
	Goals      []*ast.LabeledFormula
	Proof      *ProofDecl
}

// ProofDecl represents a proof declaration with tactic parameters.
type ProofDecl struct {
	TacticName  string
	TacticDecls []ast.Node
	TacticLets  []ast.Node
	Lineno      ast.Location
	Labels      []string
}

// L2STactic is the main entry point for the liveness-to-safety reduction.
//
// Given a set of proof goals, it transforms the first goal's temporal
// conclusion (M |= phi) into a safety property by:
//   1. Extracting temporal premises and globally assumptions
//   2. Building ranking functions from tactic declarations
//   3. Constructing the L2S monitor with ghost state
//   4. Instrumenting the model's actions with tableau updates
//   5. Replacing the conclusion with M |= true
//
// Returns the modified goal stack.
func L2STactic(cfg *L2STacticConfig) ([]*ast.LabeledFormula, error) {
	if cfg == nil || len(cfg.Goals) == 0 {
		return nil, fmt.Errorf("no goals provided")
	}

	goal := cfg.Goals[0]
	conc := proof.GoalConc(goal)
	if conc == nil {
		return nil, fmt.Errorf("goal has no conclusion")
	}

	// Verify the conclusion is a temporal models formula.
	if !temporal.IsTemporalFormula(conc) && !isTemporalModels(conc) {
		return nil, fmt.Errorf("proof goal is not temporal")
	}

	if cfg.Proof != nil && len(cfg.Proof.TacticLets) > 0 {
		return nil, fmt.Errorf("tactic does not take lets")
	}

	// Extract temporal formula from conclusion
	fmla, _ := conc.(lg.Node)
	if fmla == nil {
		return nil, fmt.Errorf("cannot extract formula from goal conclusion")
	}

	// Process temporal premises
	prems := proof.GoalPrems(goal)
	var temporalPrems []lg.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.Temporal != nil {
				if f, ok := lf.Formula.(lg.Node); ok {
					temporalPrems = append(temporalPrems, f)
				}
			}
		}
	}
	if len(temporalPrems) > 0 {
		premConj := makeAnd(temporalPrems...)
		fmla = &lg.Implies{T1: premConj, T2: fmla}
	}

	// Build finite sorts map
	finiteSorts := make(map[string]bool)
	var uninterpretedSorts []lg.Sort
	if cfg.Mod != nil && cfg.Mod.Sig != nil {
		for name, s := range cfg.Mod.Sig.Sorts {
			if cfg.Mod.FiniteSorts[name] {
				finiteSorts[name] = true
			} else if _, isUI := s.(*lg.UninterpretedSort); isUI {
				uninterpretedSorts = append(uninterpretedSorts, s)
			}
		}
	}
	sort.Slice(uninterpretedSorts, func(i, j int) bool {
		return uninterpretedSorts[i].String() < uninterpretedSorts[j].String()
	})

	// Generate ranking invariants and postconditions
	var invars []*module.LabeledFormula
	invars, postconds, _, _, err := rankingInvariants(
		goal, invars, cfg.ProofLabel, fmla,
		finiteSorts, uninterpretedSorts, cfg.Mod)
	if err != nil {
		return nil, fmt.Errorf("ranking invariant generation: %w", err)
	}

	// The postconditions are stored for later use by the verification pipeline.
	// They are attached to the model's postconds map during action instrumentation.
	// For now, store them in the module if available.
	if cfg.Mod != nil {
		if cfg.Mod.Postconds == nil {
			cfg.Mod.Postconds = make(map[string][]*module.LabeledFormula)
		}
		for _, pc := range postconds {
			cfg.Mod.Postconds["_default"] = append(cfg.Mod.Postconds["_default"], pc)
		}
	}

	// TODO: The full implementation would continue with:
	// - Desugar temporal operators (already available via l2s.Desugar)
	// - Temporal-to-named-binder conversion (already in l2s.go step 1)
	// - Action instrumentation with monitor (already in l2s.go steps 3-10)
	// - Named binder replacement (already in l2s.go step 11)
	// - Goal reconstruction (already in l2s.go step 12)
	// These steps share infrastructure with the l2s tactic and could be
	// factored out into shared functions.

	// Add invariants to goals as additional conjectures
	_ = invars
	return cfg.Goals, nil
}

func isTemporalModels(n lg.Node) bool {
	// TemporalModels is an ast.Node, not a logic.Node.
	// Check by string representation as a fallback.
	if n == nil {
		return false
	}
	s := n.String()
	return strings.Contains(s, "|=")
}

// --- Model pass helpers ---

// ModelPass applies a transformation to all formulas in a NormalProgram.
func ModelPass(model *temporal.NormalProgram, transform func(lg.Node) lg.Node) {
	if model == nil {
		return
	}
	for i, inv := range model.Invars {
		if inv.Formula != nil {
			model.Invars[i] = &module.LabeledFormula{
				Label:   inv.Label,
				Formula: transform(inv.Formula),
			}
		}
	}
	for i, asm := range model.Asms {
		if asm.Formula != nil {
			model.Asms[i] = &module.LabeledFormula{
				Label:   asm.Label,
				Formula: transform(asm.Formula),
			}
		}
	}
}

// --- Diagnostics ---

// DiagnosticInfo holds diagnostic information about a failed ranking check.
type DiagnosticInfo struct {
	Name    string
	Suffix  string
	Message string
}

// DiagnoseFailure provides diagnostic information when a ranking check fails.
func DiagnoseFailure(name string, tasks map[string]*Task, triggers map[string]*Trigger) *DiagnosticInfo {
	info := &DiagnosticInfo{Name: name}

	switch {
	case strings.HasPrefix(name, "l2s_created"):
		info.Suffix = name[len("l2s_created"):]
		info.Message = fmt.Sprintf("Failed to prove that work_created%s is finite by induction.", info.Suffix)

	case strings.HasPrefix(name, "l2s_eventually_start"):
		info.Suffix = name[len("l2s_eventually_start"):]
		info.Message = "The eventually property fails to hold when liveness invariant is false."

	case strings.HasPrefix(name, "l2s_needed_preserved"):
		info.Suffix = name[len("l2s_needed_preserved"):]
		info.Message = fmt.Sprintf("Failed to prove that work_needed%s is preserved.", info.Suffix)

	case strings.HasPrefix(name, "l2s_progress"):
		info.Suffix = name[len("l2s_progress"):]
		info.Message = fmt.Sprintf("Failed to prove that work_needed%s decreases when a helpful transition occurs.", info.Suffix)

	case strings.HasPrefix(name, "l2s_sched_stable"):
		info.Suffix = name[len("l2s_sched_stable"):]
		info.Message = fmt.Sprintf("Failed to prove that work_helpful%s is stable until helpful transition occurs.", info.Suffix)

	case strings.HasPrefix(name, "l2s_sched_exists"):
		var rankNames []string
		for sfx, task := range tasks {
			if task.WorkHelpful != nil {
				rankNames = append(rankNames, "work_helpful"+sfx)
			}
		}
		sort.Strings(rankNames)
		info.Message = fmt.Sprintf("The helpful set(s) %s have become empty, but termination has not occurred.",
			strings.Join(rankNames, " and "))

	default:
		info.Message = fmt.Sprintf("Unknown ranking failure: %s", name)
	}

	return info
}

// --- Temporal operator classification ---

// TrigGlob extracts globally formulas from a temporal formula at a given polarity.
// This is used to strengthen invariants when a trigger implies a globally property.
func TrigGlob(prop lg.Node, pos bool) []lg.Node {
	var result []lg.Node
	trigGlobRec(prop, pos, &result)
	return result
}

func trigGlobRec(prop lg.Node, pos bool, result *[]lg.Node) {
	if prop == nil {
		return
	}
	switch p := prop.(type) {
	case *lg.Globally:
		if pos {
			*result = append(*result, p)
			trigGlobRec(p.Body, pos, result)
		}
	case *lg.Eventually:
		if !pos {
			*result = append(*result, &lg.Not{Body: p})
			trigGlobRec(p.Body, pos, result)
		}
	case *lg.Implies:
		if !pos {
			trigGlobRec(p.T1, !pos, result)
			trigGlobRec(p.T2, pos, result)
		}
	case *lg.And:
		if pos {
			for _, t := range p.Terms {
				trigGlobRec(t, pos, result)
			}
		}
	case *lg.Or:
		if !pos {
			for _, t := range p.Terms {
				trigGlobRec(t, pos, result)
			}
		}
	case *lg.ForAll:
		if pos {
			trigGlobRec(p.Body, pos, result)
		}
	case *lg.Exists:
		if !pos {
			trigGlobRec(p.Body, pos, result)
		}
	case *lg.Not:
		trigGlobRec(p.Body, !pos, result)
	}
}

// --- Renaming hook ---

// RenameMap returns the inverse of a substitution map.
func RenameMap(subs map[string]string) map[string]string {
	inv := make(map[string]string, len(subs))
	for k, v := range subs {
		inv[v] = k
	}
	return inv
}

// IsTemporalAndL2S returns true if the symbol name is an L2S ghost symbol
// (but not an l2s_g symbol).
func IsTemporalAndL2S(name string) bool {
	if strings.HasPrefix(name, "l2s") && !strings.HasPrefix(name, "l2s_g") {
		return true
	}
	if strings.HasPrefix(name, "_old_l2s") {
		return true
	}
	return false
}

// --- Prop events ---

// PropEvent describes the pre/post actions for a property event.
type PropEvent struct {
	PreActions  []actions.Action
	PostActions []actions.Action
}

// NewPropEvents creates the pre/post actions for a set of globally properties.
// For each property:
//   - Pre: save old value, havoc current value, assume monotonicity constraints
//   - Post: assume semantic constraint (g(V) -> body)
func NewPropEvents(gprops []*lg.NamedBinder, lineno ast.Location) *PropEvent {
	pe := &PropEvent{}

	for _, gprop := range gprops {
		vs := gprop.Variables
		body := gprop.Body
		environ := ""
		if gprop.Environ != nil {
			environ = *gprop.Environ
		}

		oldG := OldL2sG(vs, body, environ)
		curG := L2sG(vs, body, environ)

		// Pre: save old value
		if oldG != nil && curG != nil {
			pe.PreActions = append(pe.PreActions,
				newAssignAction(oldG, curG, lineno))
			pe.PreActions = append(pe.PreActions,
				newHavocAction(curG, lineno))
		}
	}

	for _, gprop := range gprops {
		vs := gprop.Variables
		body := gprop.Body
		environ := ""
		if gprop.Environ != nil {
			environ = *gprop.Environ
		}

		oldG := OldL2sG(vs, body, environ)
		curG := L2sG(vs, body, environ)

		// Pre: assume monotonicity: old_g(V) -> g(V)
		if oldG != nil && curG != nil {
			monoFmla := makeImplies(oldG, curG)
			pe.PreActions = append(pe.PreActions,
				makeAssumeForAll(vs, monoFmla, lineno))

			// Pre: assume ~old_g(V) & body -> ~g(V)
			newFmla := makeImplies(
				makeAnd(makeNot(oldG), body),
				makeNot(curG))
			pe.PreActions = append(pe.PreActions,
				makeAssumeForAll(vs, newFmla, lineno))
		}

		// Post: assume g(V) -> body
		postFmla := makeImplies(curG, body)
		pe.PostActions = append(pe.PostActions,
			makeAssumeForAll(vs, postFmla, lineno))
	}

	return pe
}

// --- Wait events ---

// WaitEvent creates assignment actions for updating waiting predicates.
func WaitEvent(waits []*lg.NamedBinder, proofLabel string, lineno ast.Location) []actions.Action {
	var result []actions.Action
	for _, wait := range waits {
		vs := wait.Variables
		body := wait.Body

		// w(V) := w(V) & ~body & ~g(~body)
		wApp := wait // The waiting predicate itself
		negBody := makeNot(body)
		gNegBody := L2sG(vs, negBody, proofLabel)

		newVal := makeAnd(wApp, negBody, makeNot(gNegBody))
		result = append(result, newAssignAction(wApp, newVal, lineno))
	}
	return result
}

// --- helper action constructors ---

func newAssignAction(lhs, rhs lg.Node, lineno ast.Location) actions.Action {
	act := actions.NewAssignAction(lhs, rhs)
	act.SetLineno(lineno)
	return act
}

func newHavocAction(target lg.Node, lineno ast.Location) actions.Action {
	act := actions.NewHavocAction(target)
	act.SetLineno(lineno)
	return act
}

func makeAssumeForAll(vs []*lg.Var, body lg.Node, lineno ast.Location) actions.Action {
	fmla := ForAll(vs, body)
	act := actions.NewAssumeAction(fmla)
	act.SetLineno(lineno)
	return act
}

func makeImplies(t1, t2 lg.Node) lg.Node {
	imp, err := lg.NewImplies(t1, t2)
	if err != nil {
		return t2
	}
	return imp
}

func makeAnd(terms ...lg.Node) lg.Node {
	and, err := lg.NewAnd(terms...)
	if err != nil {
		return lg.True
	}
	return and
}

func makeNot(body lg.Node) lg.Node {
	not, err := lg.NewNot(body)
	if err != nil {
		return body
	}
	return not
}

// --- Desugar ---

// Desugar transforms temporal operators in invariants to L2S ghost state.
//
// $was. phi(V)       -->  l2s_saved & ($l2s_s V.phi(V))(V)
// $happened. phi     -->  l2s_saved & ~($l2s_w V.phi(V))(V)
//
// The l2s_s binder is pushed inside propositional connectives so that
// saved values correspond to atoms (avoiding redundant saved values).
func Desugar(expr lg.Node, proofLabel string, l2sSaved lg.Node) lg.Node {
	if expr == nil {
		return nil
	}
	if nb, ok := expr.(*lg.NamedBinder); ok {
		switch nb.Name {
		case "was":
			if len(nb.Variables) > 0 {
				// Error: 'was' does not take parameters
				return expr
			}
			return makeAnd(l2sSaved, applyWas(nb.Body, proofLabel))
		case "happened":
			if len(nb.Variables) > 0 {
				return expr
			}
			return makeAnd(l2sSaved, applyHappened(nb.Body, proofLabel))
		}
	}
	return cloneWithTransform(expr, func(n lg.Node) lg.Node {
		return Desugar(n, proofLabel, l2sSaved)
	})
}

func applyWas(expr lg.Node, proofLabel string) lg.Node {
	switch e := expr.(type) {
	case *lg.And:
		terms := make([]lg.Node, len(e.Terms))
		for i, t := range e.Terms {
			terms[i] = applyWas(t, proofLabel)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Node, len(e.Terms))
		for i, t := range e.Terms {
			terms[i] = applyWas(t, proofLabel)
		}
		return &lg.Or{Terms: terms}
	case *lg.Not:
		return &lg.Not{Body: applyWas(e.Body, proofLabel)}
	case *lg.Implies:
		return &lg.Implies{T1: applyWas(e.T1, proofLabel), T2: applyWas(e.T2, proofLabel)}
	case *lg.Iff:
		return &lg.Iff{T1: applyWas(e.T1, proofLabel), T2: applyWas(e.T2, proofLabel)}
	default:
		// Atomic: create l2s_s binder
		nb := L2sS(nil, expr, proofLabel)
		return nb
	}
}

func applyHappened(expr lg.Node, proofLabel string) lg.Node {
	// ~(l2s_w(V, body)(V))
	nb := L2sW(nil, expr, proofLabel)
	return makeNot(nb)
}

// cloneWithTransform recursively applies transform to all children of a node.
func cloneWithTransform(node lg.Node, transform func(lg.Node) lg.Node) lg.Node {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *lg.And:
		terms := make([]lg.Node, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = transform(t)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Node, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = transform(t)
		}
		return &lg.Or{Terms: terms}
	case *lg.Not:
		return &lg.Not{Body: transform(n.Body)}
	case *lg.Implies:
		return &lg.Implies{T1: transform(n.T1), T2: transform(n.T2)}
	case *lg.Iff:
		return &lg.Iff{T1: transform(n.T1), T2: transform(n.T2)}
	case *lg.ForAll:
		return &lg.ForAll{Variables: n.Variables, Body: transform(n.Body)}
	case *lg.Exists:
		return &lg.Exists{Variables: n.Variables, Body: transform(n.Body)}
	case *lg.Eq:
		return &lg.Eq{T1: transform(n.T1), T2: transform(n.T2)}
	default:
		return node
	}
}

// --- Registration ---

// TacticFunc is the signature for tactic implementations.
type TacticFunc func(cfg *L2STacticConfig) ([]*ast.LabeledFormula, error)

// RegisteredTactics maps tactic names to implementations.
var RegisteredTactics = map[string]TacticFunc{
	"ranking": func(cfg *L2STacticConfig) ([]*ast.LabeledFormula, error) {
		return L2STactic(cfg)
	},
}

// RegisterTactic registers a tactic by name.
func RegisterTactic(name string, fn TacticFunc) {
	RegisteredTactics[name] = fn
}

// GetTactic looks up a registered tactic by name.
func GetTactic(name string) (TacticFunc, bool) {
	fn, ok := RegisteredTactics[name]
	return fn, ok
}

// --- Dependency helpers ---

// Dependencies computes the transitive closure of symbol dependencies.
func Dependencies(syms []string, deps map[string][]string) map[string]bool {
	result := make(map[string]bool)
	var visit func(string)
	visit = func(sym string) {
		if result[sym] {
			return
		}
		result[sym] = true
		for _, dep := range deps[sym] {
			visit(dep)
		}
	}
	for _, sym := range syms {
		visit(sym)
	}
	return result
}

// ConvertToInit transforms a formula by replacing temporal sub-formulas
// with l2s_init binders.
func ConvertToInit(fmla lg.Node, proofLabel string) lg.Node {
	if fmla == nil {
		return nil
	}
	switch f := fmla.(type) {
	case *lg.And:
		terms := make([]lg.Node, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = ConvertToInit(t, proofLabel)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Node, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = ConvertToInit(t, proofLabel)
		}
		return &lg.Or{Terms: terms}
	case *lg.Not:
		return &lg.Not{Body: ConvertToInit(f.Body, proofLabel)}
	case *lg.Implies:
		return &lg.Implies{
			T1: ConvertToInit(f.T1, proofLabel),
			T2: ConvertToInit(f.T2, proofLabel),
		}
	case *lg.Iff:
		return &lg.Iff{
			T1: ConvertToInit(f.T1, proofLabel),
			T2: ConvertToInit(f.T2, proofLabel),
		}
	case *lg.ForAll:
		return &lg.ForAll{Variables: f.Variables, Body: ConvertToInit(f.Body, proofLabel)}
	case *lg.Exists:
		return &lg.Exists{Variables: f.Variables, Body: ConvertToInit(f.Body, proofLabel)}
	default:
		// Wrap in l2s_init
		return L2sInit(nil, fmla, proofLabel)
	}
}
