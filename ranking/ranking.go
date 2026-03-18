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
	"github.com/glycerine/goivy/l2s"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	modpkg "github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
	"github.com/glycerine/goivy/temporal"
	"github.com/glycerine/goivy/transrel"
)

// Debug controls debug output for ranking computations.
var Debug = false

// --- Formula helpers ---

// ForAll wraps a body in a ForAll quantifier if there are variables.
// If vs is empty, returns body unchanged.
func ForAll(vs []*lg.Variable, body lg.Expr) lg.Expr {
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
func Exists(vs []*lg.Variable, body lg.Expr) lg.Expr {
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
func OldOf(fmla lg.Expr) lg.Expr {
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
	case *lg.Symbol:
		return lg.NewSymbol(transrel.Old(f.Name), f.CSort)
	case *lg.Eq:
		return &lg.Eq{T1: OldOf(f.T1), T2: OldOf(f.T2)}
	case *lg.Not:
		return &lg.Not{Body: OldOf(f.Body)}
	case *lg.And:
		terms := make([]lg.Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = OldOf(t)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Expr, len(f.Terms))
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

func makeOldFunc(fn lg.Expr) lg.Expr {
	switch f := fn.(type) {
	case *lg.Symbol:
		return lg.NewSymbol(transrel.Old(f.Name), f.CSort)
	default:
		return fn
	}
}

// --- Named binder constructors ---

// L2sD creates the l2s_d predicate for a sort.
// l2s_d : sort -> Boolean (tracks which domain elements are active).
func L2sD(sort lg.Sort) *lg.Symbol {
	fs, err := lg.NewFunctionSort(sort, lg.Boolean)
	if err != nil {
		return lg.NewSymbol("l2s_d", lg.Boolean)
	}
	return lg.NewSymbol("l2s_d", fs)
}

// L2sW creates an l2s_w named binder (waiting predicate).
func L2sW(vs []*lg.Variable, body lg.Expr, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_w", vs, strPtr(label), body)
	return nb
}

// L2sG creates an l2s_g named binder (globally predicate).
func L2sG(vs []*lg.Variable, body lg.Expr, environ string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_g", vs, strPtr(environ), body)
	return nb
}

// OldL2sG creates an _old_l2s_g named binder (old globally predicate).
func OldL2sG(vs []*lg.Variable, body lg.Expr, environ string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("_old_l2s_g", vs, strPtr(environ), body)
	return nb
}

// L2sInit creates an l2s_init named binder.
func L2sInit(vs []*lg.Variable, body lg.Expr, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_init", vs, strPtr(label), body)
	return nb
}

// L2sWhen creates an l2s_when named binder.
func L2sWhen(name string, vs []*lg.Variable, body lg.Expr, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_when"+name, vs, strPtr(label), body)
	return nb
}

// L2sOld creates an l2s_old named binder.
func L2sOld(vs []*lg.Variable, body lg.Expr, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_old", vs, strPtr(label), body)
	return nb
}

// L2sS creates an l2s_s named binder (saved state).
func L2sS(vs []*lg.Variable, body lg.Expr, label string) *lg.NamedBinder {
	nb, _ := lg.NewNamedBinder("l2s_s", vs, strPtr(label), body)
	return nb
}

func strPtr(s string) *string {
	return &s
}

// --- Task and Trigger types ---

// Task holds the ranking function definitions for one task suffix.
type Task struct {
	WorkCreated  lg.Expr // definition of work_created
	WorkNeeded   lg.Expr // definition of work_needed
	WorkProgress lg.Expr // definition of work_progress
	WorkInvar    lg.Expr // definition of work_invar
	WorkHelpful  lg.Expr // definition of work_helpful
	WorkWitness  lg.Expr // optional work_witness
}

// Trigger holds the trigger condition for a task.
type Trigger struct {
	WorkStart lg.Expr // definition of work_start
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
	lineno := ast.Location{Filename: "ranking", Line: 0}

	// Find the TemporalModels in the goal
	tm := l2s.FindTemporalModels(goal)
	if tm == nil {
		// Fall back to GoalConc check
		conc := proof.GoalConc(goal)
		if conc == nil {
			return nil, fmt.Errorf("goal has no conclusion")
		}
		if !temporal.IsTemporalFormula(conc) && !isTemporalModels(conc) {
			return nil, fmt.Errorf("proof goal is not temporal")
		}
	}

	if cfg.Proof != nil && len(cfg.Proof.TacticLets) > 0 {
		return nil, fmt.Errorf("tactic does not take lets")
	}

	// Extract the model and formula
	m := cfg.Mod
	model := l2s.ExtractNormalProgram(m)

	var fmla lg.Expr
	if tm != nil {
		fmla, _ = tm.Fmla.(lg.Expr)
	}
	if fmla == nil {
		conc := proof.GoalConc(goal)
		if conc != nil {
			fmla, _ = conc.(lg.Expr)
		}
	}
	if fmla == nil {
		return nil, fmt.Errorf("cannot extract formula from goal conclusion")
	}

	// Process temporal premises
	prems := proof.GoalPrems(goal)
	var temporalPrems []lg.Expr
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.Temporal != nil {
				if f, ok := lf.Formula.(lg.Expr); ok {
					temporalPrems = append(temporalPrems, f)
				}
			}
		}
	}
	if len(temporalPrems) > 0 {
		premConj := makeAnd(temporalPrems...)
		fmla = &lg.Implies{T1: premConj, T2: fmla}
	}

	proofLabel := cfg.ProofLabel

	// Build finite sorts map
	finiteSorts := make(map[string]bool)
	var uninterpretedSorts []lg.Sort
	if m != nil && m.Sig != nil {
		for name, s := range m.Sig.Sorts {
			if m.FiniteSorts[name] {
				finiteSorts[name] = true
			} else if _, isUI := s.(*lg.UninterpretedSort); isUI {
				uninterpretedSorts = append(uninterpretedSorts, s)
			}
		}
	}
	sort.Slice(uninterpretedSorts, func(i, j int) bool {
		return uninterpretedSorts[i].String() < uninterpretedSorts[j].String()
	})

	// Add model invariants
	var invars []*module.LabeledFormula
	invars = append(invars, model.Invars...)

	// Generate ranking invariants and postconditions
	invars, postconds, _, _, err := rankingInvariants(
		goal, invars, proofLabel, fmla,
		finiteSorts, uninterpretedSorts, m)
	if err != nil {
		return nil, fmt.Errorf("ranking invariant generation: %w", err)
	}

	// Desugar $was/$happened in invars and postconds
	l2sSaved := l2s.L2SSaved()
	desugarFn := func(n lg.Expr) lg.Expr {
		return l2s.Desugar(n, proofLabel)
	}
	for i, inv := range invars {
		invars[i] = &module.LabeledFormula{
			Label:   inv.Label,
			Formula: desugarFn(inv.Formula),
		}
	}
	for i, pc := range postconds {
		postconds[i] = &module.LabeledFormula{
			Label:   pc.Label,
			Formula: desugarFn(pc.Formula),
		}
	}
	_ = l2sSaved // used by Desugar internally

	// --- Build shared config ---
	defnDeps := l2s.BuildDefnDeps(m)

	icfg := &l2s.InstrumentationConfig{
		ProofLabel:         proofLabel,
		Lineno:             lineno,
		FiniteSorts:        finiteSorts,
		UninterpretedSorts: uninterpretedSorts,
		Mod:                m,
		Fmla:               fmla,
		Invars:             invars,
		Postconds:          postconds,
		Dependencies:       l2s.BuildDependenciesFunc(defnDeps),
	}

	// --- Model pass helper (ranking version: also transforms postconds) ---
	modPass := func(transform func(lg.Expr) lg.Expr) {
		for i, inv := range model.Invars {
			model.Invars[i] = &modpkg.LabeledFormula{
				Label:   inv.Label,
				Formula: transform(inv.Formula),
			}
		}
		for i, asm := range model.Asms {
			model.Asms[i] = &modpkg.LabeledFormula{
				Label:   asm.Label,
				Formula: transform(asm.Formula),
			}
		}
		for i, b := range model.Bindings {
			newStmt := l2s.TransformAction(b.Action.Stmt, transform)
			model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
		}
		if model.Init != nil {
			model.Init = l2s.TransformAction(model.Init, transform)
		}
		for i, inv := range invars {
			invars[i] = &modpkg.LabeledFormula{
				Label:   inv.Label,
				Formula: transform(inv.Formula),
			}
		}
		// Ranking-specific: also transform postconds
		for i, pc := range postconds {
			postconds[i] = &modpkg.LabeledFormula{
				Label:   pc.Label,
				Formula: transform(pc.Formula),
			}
		}
	}

	// ---------------------------------------------------------------
	// Step 1: Convert temporal operators to named binders (shared)
	// ---------------------------------------------------------------
	l2s.SharedStep1_ConvertTemporals(icfg, model, modPass)

	// Normalize named binders in postconds (already done for model by SharedStep1)
	// (SharedStep1 calls modPass which handles postconds)

	// ---------------------------------------------------------------
	// Step 3: Collect named binders from conjectures + postconds (shared)
	// ---------------------------------------------------------------
	l2s.SharedStep3_CollectNamedBinders(icfg, model, false)

	// Build save/wait/reset_w
	l2s.SharedBuildSaveAndWait(icfg)

	// Build addConstsToD
	icfg.AddConstsToD = l2s.BuildAddConstsToD(m, uninterpretedSorts, lineno)

	// ---------------------------------------------------------------
	// Step 6: Tableau construction (shared)
	// ---------------------------------------------------------------
	l2s.SharedStep6_BuildTableau(icfg)

	// ---------------------------------------------------------------
	// Step 7: Action instrumentation (shared)
	// ---------------------------------------------------------------
	l2s.SharedStep7_InstrumentActions(icfg, model)

	// ---------------------------------------------------------------
	// Step 8: Patch exported actions (shared, ranking mode: with postconds)
	// ---------------------------------------------------------------
	l2s.SharedStep8_PatchExports(icfg, model)

	// ---------------------------------------------------------------
	// Step 9: Idle action (ranking-specific: no monitor, no fair cycle)
	// ---------------------------------------------------------------
	var idleParts []actions.Action
	idleParts = append(idleParts, icfg.AssumeGAxioms...)
	idleParts = append(idleParts, icfg.AssumeWhenAxioms...)
	idleParts = append(idleParts, icfg.ResetW...)
	idleParts = append(idleParts, icfg.AddConstsToD...)

	idleAction := l2s.SetLineno(actions.ConcatActions(idleParts...), lineno)
	idleAction.SetFormalParams(nil)
	idleAction.SetFormalReturns(nil)

	model.Bindings = append(model.Bindings, &temporal.ActionTermBinding{
		Name:   "_idle",
		Action: &temporal.ActionTerm{Stmt: idleAction},
	})
	model.Calls = append(model.Calls, "_idle")
	if model.Postconds == nil {
		model.Postconds = make(map[string][]*modpkg.LabeledFormula)
	}
	model.Postconds["_idle"] = postconds

	// ---------------------------------------------------------------
	// Step 10: Init action (ranking-specific: no monitor state vars)
	// ---------------------------------------------------------------
	var rankingInitActions []actions.Action
	rankingInitActions = append(rankingInitActions, icfg.AddConstsToD...)
	rankingInitActions = append(rankingInitActions, icfg.ResetW...)
	rankingInitActions = append(rankingInitActions, icfg.AssumeGAxioms...)
	rankingInitActions = append(rankingInitActions, icfg.AssumeInitAxioms...)
	rankingInitActions = append(rankingInitActions, l2s.SetLineno(actions.NewAssumeAction(icfg.NotLf), lineno))

	if model.Init != nil {
		model.Init = actions.PostfixAction(model.Init, rankingInitActions)
	}

	// ---------------------------------------------------------------
	// Step 11: Replace named binders (shared)
	// ---------------------------------------------------------------
	l2s.SharedStep11_ReplaceNamedBinders(icfg, model, modPass)

	// ---------------------------------------------------------------
	// Step 12: Build new goal (shared)
	// ---------------------------------------------------------------
	if tm != nil {
		return l2s.SharedStep12_BuildGoal(goal, cfg.Goals, prems, tm)
	}
	// Fallback: return goals as-is if no TemporalModels found
	return cfg.Goals, nil
}

func isTemporalModels(n lg.Expr) bool {
	// TemporalModels is an ast.Node, not a logic.Expr.
	// Check by string representation as a fallback.
	if n == nil {
		return false
	}
	s := n.String()
	return strings.Contains(s, "|=")
}

// --- Model pass helpers ---

// ModelPass applies a transformation to all formulas in a NormalProgram.
func ModelPass(model *temporal.NormalProgram, transform func(lg.Expr) lg.Expr) {
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
func TrigGlob(prop lg.Expr, pos bool) []lg.Expr {
	var result []lg.Expr
	trigGlobRec(prop, pos, &result)
	return result
}

func trigGlobRec(prop lg.Expr, pos bool, result *[]lg.Expr) {
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

func newAssignAction(lhs, rhs lg.Expr, lineno ast.Location) actions.Action {
	act := actions.NewAssignAction(lhs, rhs)
	act.SetLineno(lineno)
	return act
}

func newHavocAction(target lg.Expr, lineno ast.Location) actions.Action {
	act := actions.NewHavocAction(target)
	act.SetLineno(lineno)
	return act
}

func makeAssumeForAll(vs []*lg.Variable, body lg.Expr, lineno ast.Location) actions.Action {
	fmla := ForAll(vs, body)
	act := actions.NewAssumeAction(fmla)
	act.SetLineno(lineno)
	return act
}

func makeImplies(t1, t2 lg.Expr) lg.Expr {
	imp, err := lg.NewImplies(t1, t2)
	if err != nil {
		return t2
	}
	return imp
}

func makeAnd(terms ...lg.Expr) lg.Expr {
	and, err := lg.NewAnd(terms...)
	if err != nil {
		return lg.True
	}
	return and
}

func makeNot(body lg.Expr) lg.Expr {
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
func Desugar(expr lg.Expr, proofLabel string, l2sSaved lg.Expr) lg.Expr {
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
	return cloneWithTransform(expr, func(n lg.Expr) lg.Expr {
		return Desugar(n, proofLabel, l2sSaved)
	})
}

func applyWas(expr lg.Expr, proofLabel string) lg.Expr {
	switch e := expr.(type) {
	case *lg.And:
		terms := make([]lg.Expr, len(e.Terms))
		for i, t := range e.Terms {
			terms[i] = applyWas(t, proofLabel)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Expr, len(e.Terms))
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

func applyHappened(expr lg.Expr, proofLabel string) lg.Expr {
	// ~(l2s_w(V, body)(V))
	nb := L2sW(nil, expr, proofLabel)
	return makeNot(nb)
}

// cloneWithTransform recursively applies transform to all children of a node.
func cloneWithTransform(node lg.Expr, transform func(lg.Expr) lg.Expr) lg.Expr {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *lg.And:
		terms := make([]lg.Expr, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = transform(t)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Expr, len(n.Terms))
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
func ConvertToInit(fmla lg.Expr, proofLabel string) lg.Expr {
	if fmla == nil {
		return nil
	}
	switch f := fmla.(type) {
	case *lg.And:
		terms := make([]lg.Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = ConvertToInit(t, proofLabel)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Expr, len(f.Terms))
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
