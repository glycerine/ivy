// ranking.go implements liveness-to-safety reduction using
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
package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/temporal"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// --- Formula helpers ---

// CheckForAll wraps a body in a CheckForAll quantifier if there are variables.
// If vs is empty, returns body unchanged.
func CheckForAll(vs []*lg.Variable, body lg.Expr) lg.Expr {
	if len(vs) == 0 {
		return body
	}
	fa, err := lg.NewForAll(vs, body)
	if err != nil {
		return body
	}
	return fa
}

// CheckExists wraps a body in an CheckExists quantifier if there are variables.
// If vs is empty, returns body unchanged.
func CheckExists(vs []*lg.Variable, body lg.Expr) lg.Expr {
	if len(vs) == 0 {
		return body
	}
	ex, err := lg.NewExists(vs, body)
	if err != nil {
		return body
	}
	return ex
}

// CheckOldOf replaces function symbols in a formula with their "old" versions.
// For an application f(args), it returns old_f(args).
// For compound formulas, it recurses into sub-formulas.
func CheckOldOf(fmla lg.Expr) lg.Expr {
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
		return lg.NewConst(actions.Old(f.Name), f.CSort)
	case *lg.Eq:
		return &lg.Eq{T1: CheckOldOf(f.T1), T2: CheckOldOf(f.T2)}
	case *lg.Not:
		return &lg.Not{Body: CheckOldOf(f.Body)}
	case *lg.And:
		terms := make([]lg.Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = CheckOldOf(t)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = CheckOldOf(t)
		}
		return &lg.Or{Terms: terms}
	case *lg.Implies:
		return &lg.Implies{T1: CheckOldOf(f.T1), T2: CheckOldOf(f.T2)}
	case *lg.Iff:
		return &lg.Iff{T1: CheckOldOf(f.T1), T2: CheckOldOf(f.T2)}
	case *lg.ForAll:
		return &lg.ForAll{Variables: f.Variables, Body: CheckOldOf(f.Body)}
	case *lg.Exists:
		return &lg.Exists{Variables: f.Variables, Body: CheckOldOf(f.Body)}
	default:
		return fmla
	}
}

func makeOldFunc(fn lg.Expr) lg.Expr {
	switch f := fn.(type) {
	case *lg.Const:
		return lg.NewConst(actions.Old(f.Name), f.CSort)
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
//  1. Extracting temporal premises and globally assumptions
//  2. Building ranking functions from tactic declarations
//  3. Constructing the L2S monitor with ghost state
//  4. Instrumenting the model's actions with tableau updates
//  5. Replacing the conclusion with M |= true
//
// Returns the modified goal stack.
// RankingL2STactic is the ranking-tactic entry point. The "Ranking" prefix
// distinguishes it from the proof-tactic L2STactic in l2s.go.
func RankingL2STactic(cfg *L2STacticConfig) ([]*ast.LabeledFormula, error) {
	if cfg == nil || len(cfg.Goals) == 0 {
		return nil, fmt.Errorf("no goals provided")
	}

	goal := cfg.Goals[0]
	lineno := ast.Location{Filename: "ranking", Line: 0}

	// Find the TemporalModels in the goal
	tm := FindTemporalModels(goal)
	if tm == nil {
		// Fall back to GoalConc check; only lg.Expr conclusions can be
		// "temporal formulas" (the alternative is *ast.TemporalModels which
		// FindTemporalModels would have caught above).
		conc := proof.GoalConcExpr(goal)
		if conc == nil {
			return nil, fmt.Errorf("goal has no conclusion")
		}
		if !temporal.IsTemporalFormula(conc) && !isTemporalModels(conc) {
			return nil, fmt.Errorf("check/ranking: [1]proof goal is not temporal")
		}
	}

	if cfg.Proof != nil && len(cfg.Proof.TacticLets) > 0 {
		return nil, fmt.Errorf("tactic does not take lets")
	}

	// Extract the model and formula
	m := cfg.Mod
	model := ExtractNormalProgram(m)

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
			if lf.IsTemporal() {
				if f, ok := lf.Formula.(lg.Expr); ok {
					temporalPrems = append(temporalPrems, f)
				}
			}
		}
	}
	if len(temporalPrems) > 0 {
		premConj := rankingMakeAnd(temporalPrems...)
		fmla = &lg.Implies{T1: premConj, T2: fmla}
	}

	proofLabel := cfg.ProofLabel

	// Build finite sorts map
	finiteSorts := make(map[string]bool)
	var uninterpretedSorts []lg.Sort
	if m != nil && m.Sig != nil {
		for name, s := range m.Sig.Sorts.All() {
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
	var invars []*ast.LabeledFormula
	invars = append(invars, model.Invars...)

	// Generate ranking invariants and postconditions
	invars, postconds, _, _, err := rankingInvariants(
		goal, invars, proofLabel, fmla,
		finiteSorts, uninterpretedSorts, m)
	if err != nil {
		return nil, fmt.Errorf("ranking invariant generation: %w", err)
	}

	// Desugar $was/$happened in invars and postconds.
	// Use Clone (not NewLabeledFormula) to preserve LF id + metadata,
	// matching Python's `expr.clone(...)` recursion in desugar
	// (ivy_ranking.py:505).
	l2sSaved := L2SSaved()
	desugarFn := func(n lg.Expr) (lg.Expr, error) {
		return Desugar(n, proofLabel)
	}
	for i, inv := range invars {
		desugared, err := desugarFn(inv.Formula.(lg.Expr))
		if err != nil {
			return nil, err
		}
		invars[i] = inv.Clone([]ast.Node{inv.Label, desugared}).(*ast.LabeledFormula)
	}
	for i, pc := range postconds {
		desugared, err := desugarFn(pc.Formula.(lg.Expr))
		if err != nil {
			return nil, err
		}
		postconds[i] = pc.Clone([]ast.Node{pc.Label, desugared}).(*ast.LabeledFormula)
	}
	_ = l2sSaved // used by Desugar internally

	// Python ivy_ranking.py:449-461: WhenOperator invariant generation from pprems.
	{
		var pprems []lg.Expr
		for _, p := range prems {
			if lf, ok := p.(*ast.LabeledFormula); ok && proof.GoalIsProperty(lf) {
				if e, ok := lf.Formula.(lg.Expr); ok {
					pprems = append(pprems, e)
				}
			}
		}
		var allFmlas []lg.Expr
		for _, inv := range invars {
			if e, ok := inv.Formula.(lg.Expr); ok {
				allFmlas = append(allFmlas, e)
			}
		}
		allFmlas = append(allFmlas, pprems...)
		seen := make(map[string]bool)
		var winvs []lg.Expr
		for _, f := range allFmlas {
			for _, t := range lu.TemporalsAst(f) {
				wo, ok := t.(*lg.WhenOperator)
				if !ok || wo.Name != "first" {
					continue
				}
				// Key by Sexp (structural canonical form). The Python
				// ranking equivalent uses Expr struct equality on
				// WhenOperator. wo.String() = PrettyFmla drops sort
				// annotations and would collapse sort-distinct WhenOps.
				key := string(wo.Sexp())
				if seen[key] {
					continue
				}
				seen[key] = true
				nws := &lg.Or{Terms: []lg.Expr{
					&lg.Not{Body: L2SWaiting()},
					&lg.Not{Body: applyNB(l2sW(nil, wo.T2, proofLabel))},
				}}
				tmp := &lg.Implies{
					T1: &lg.Not{Body: nws},
					T2: &lg.Eq{T1: wo, T2: &lg.WhenOperator{Name: "next", T1: wo.T1, T2: wo.T2}},
				}
				tmp2 := &lg.Implies{
					T1: applyNB(l2sInit(nil, &lg.Eventually{Environ: strPtr(proofLabel), Body: wo.T2}, proofLabel)),
					T2: tmp,
				}
				winvs = append(winvs, tmp2)
			}
		}
		acfg := cfg.Mod.Cfg.AstCfg
		for i, winv := range winvs {
			label := lg.NewConst(fmt.Sprintf("l2s_when_%d", i), lg.Boolean)
			invars = append(invars, acfg.NewLabeledFormula(label, winv))
		}
	}

	// Python ivy_ranking.py:465-473: print invariants and postconditions.
	if m != nil && m.Cfg != nil && m.Cfg.L2SDebug {
		fmt.Println("--- invariants ---")
		for _, inv := range invars {
			fmt.Printf("invariant %v\n", inv)
		}
		fmt.Println("---------------------------")

		fmt.Println("--- postconditions ---")
		for _, pc := range postconds {
			fmt.Printf("assert %v\n", pc)
		}
		fmt.Println("---------------------------")
	}

	// Python ivy_ranking.py:511: model.invars = model.invars + invars
	model.Invars = append(model.Invars, invars...)

	// --- Build shared config ---
	defnDeps := BuildDefnDeps(m)

	icfg := &InstrumentationConfig{
		ProofLabel:         proofLabel,
		Lineno:             lineno,
		FiniteSorts:        finiteSorts,
		UninterpretedSorts: uninterpretedSorts,
		Mod:                m,
		Fmla:               fmla,
		Postconds:          postconds,
		Dependencies:       BuildDependenciesFunc(defnDeps),
	}

	// --- Model pass helper (ranking version: also transforms postconds) ---
	// Use Clone (not NewLabeledFormula) to preserve LF id + metadata,
	// matching Python ivy_ranking.py:526-527 where each `transform(x)` ends
	// up calling `LF.clone(args)` via the recursive `ast.clone` dispatch.
	modPass := func(transformName string, transform func(ast.Node) ast.Node) {
		if xtracer.Enabled {
			nPropPrems := 0
			for _, p := range prems {
				if lf, ok := p.(*ast.LabeledFormula); ok && proof.GoalIsProperty(lf) {
					nPropPrems++
				}
			}
			xtracer.Trace("ranking.modPass ENTER transform=%s nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPropPrems=%d nPostconds=%d",
				transformName, len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), nPropPrems, len(postconds))
		}
		for i, inv := range model.Invars {
			xtracer.Trace("ranking.modPass clone invar[%d] ENTER HASH canon=%v", i, inv.Canon())
			model.Invars[i] = transform(inv).(*ast.LabeledFormula)
			xtracer.Trace("ranking.modPass clone invar[%d] EXIT HASH canon=%v", i, model.Invars[i].Canon())
		}
		for i, asm := range model.Asms {
			xtracer.Trace("ranking.modPass clone asm[%d] ENTER HASH canon=%v", i, asm.Canon())
			model.Asms[i] = transform(asm).(*ast.LabeledFormula)
			xtracer.Trace("ranking.modPass clone asm[%d] EXIT HASH canon=%v", i, model.Asms[i].Canon())
		}
		for i, b := range model.Bindings {
			xtracer.Trace("ranking.modPass clone binding[%d] ENTER name=%s", i, b.Name)
			// Python: model.bindings[i] = b.clone([transform(b.action)])
			newAction := transform(b.Action).(*temporal.ActionTerm)
			model.Bindings[i] = b.CloneAction(newAction)
			xtracer.Trace("ranking.modPass clone binding[%d] EXIT name=%s", i, b.Name)
		}
		if model.Init != nil {
			xtracer.Trace("ranking.modPass clone init ENTER")
			// Python: model.init = transform(model.init)
			model.Init = transform(model.Init).(actions.ActionsAction)
			xtracer.Trace("ranking.modPass clone init EXIT")
		}
		// Python ivy_ranking.py:532: list_transform(prems, transform)
		for i, p := range prems {
			lf, ok := p.(*ast.LabeledFormula)
			if !ok || !proof.GoalIsProperty(lf) {
				continue
			}
			xtracer.Trace("ranking.modPass clone prem[%d] ENTER HASH canon=%v", i, lf.Canon())
			prems[i] = transform(lf).(*ast.LabeledFormula)
			xtracer.Trace("ranking.modPass clone prem[%d] EXIT HASH canon=%v", i, prems[i].(*ast.LabeledFormula).Canon())
		}
		// Ranking-specific: also transform postconds
		for i, pc := range postconds {
			xtracer.Trace("ranking.modPass clone postcond[%d] ENTER HASH canon=%v", i, pc.Canon())
			postconds[i] = transform(pc).(*ast.LabeledFormula)
			xtracer.Trace("ranking.modPass clone postcond[%d] EXIT HASH canon=%v", i, postconds[i].Canon())
		}
		if xtracer.Enabled {
			nPropPrems := 0
			for _, p := range prems {
				if lf, ok := p.(*ast.LabeledFormula); ok && proof.GoalIsProperty(lf) {
					nPropPrems++
				}
			}
			xtracer.Trace("ranking.modPass EXIT transform=%s nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPropPrems=%d nPostconds=%d",
				transformName, len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), nPropPrems, len(postconds))
		}
	}

	// ---------------------------------------------------------------
	// Step 1: Convert temporal operators to named binders (shared)
	// ---------------------------------------------------------------
	xtracer.Trace("ranking.SharedStep1 ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))
	SharedStep1_ConvertTemporals(icfg, model, modPass)
	xtracer.Trace("ranking.SharedStep1 EXIT nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))

	// Normalize named binders in postconds (already done for model by SharedStep1)
	// (SharedStep1 calls modPass which handles postconds)

	// ---------------------------------------------------------------
	// Step 3: Collect named binders from conjectures + postconds (shared)
	// ---------------------------------------------------------------
	xtracer.Trace("ranking.SharedStep3 ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))
	SharedStep3_CollectNamedBinders(icfg, model, false)
	xtracer.Trace("ranking.SharedStep3 EXIT nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))

	// Build save/wait/reset_w
	xtracer.Trace("ranking.SharedBuildSaveAndWait ENTER")
	SharedBuildSaveAndWait(icfg)
	xtracer.Trace("ranking.SharedBuildSaveAndWait EXIT")

	// Build addConstsToD
	icfg.AddConstsToD = BuildAddConstsToD(m, uninterpretedSorts, lineno)

	// ---------------------------------------------------------------
	// Step 6: Tableau construction (shared)
	// ---------------------------------------------------------------
	xtracer.Trace("ranking.SharedStep6 ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))
	SharedStep6_BuildTableau(icfg)
	xtracer.Trace("ranking.SharedStep6 EXIT nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))

	// ---------------------------------------------------------------
	// Step 7: Action instrumentation (shared)
	// ---------------------------------------------------------------
	xtracer.Trace("ranking.SharedStep7 ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))
	SharedStep7_InstrumentActions(icfg, model)
	xtracer.Trace("ranking.SharedStep7 EXIT nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))

	// ---------------------------------------------------------------
	// Step 8: Patch exported actions (shared, ranking mode: with postconds)
	// ---------------------------------------------------------------
	xtracer.Trace("ranking.SharedStep8 ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))
	SharedStep8_PatchExports(icfg, model)
	xtracer.Trace("ranking.SharedStep8 EXIT nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))

	// ---------------------------------------------------------------
	// Step 9: Idle action (ranking-specific: no monitor, no fair cycle)
	// ---------------------------------------------------------------
	var idleParts []actions.ActionsAction
	idleParts = append(idleParts, icfg.AssumeGAxioms...)
	idleParts = append(idleParts, icfg.AssumeWhenAxioms...)
	idleParts = append(idleParts, icfg.ResetW...)
	idleParts = append(idleParts, icfg.AddConstsToD...)

	idleAction := setLineno(actions.ConcatActions(idleParts...), lineno)
	idleAction.SetFormalParams(nil)
	idleAction.SetFormalReturns(nil)

	model.Bindings = append(model.Bindings, &temporal.ActionTermBinding{
		Name:   "_idle",
		Action: &temporal.ActionTerm{Stmt: idleAction},
	})
	model.Calls = append(model.Calls, "_idle")
	if model.Postconds == nil {
		model.Postconds = make(map[string][]*ast.LabeledFormula)
	}
	model.Postconds["_idle"] = postconds

	// ---------------------------------------------------------------
	// Step 10: Init action (ranking-specific: no monitor state vars)
	// ---------------------------------------------------------------
	var rankingInitActions []actions.ActionsAction
	rankingInitActions = append(rankingInitActions, icfg.AddConstsToD...)
	rankingInitActions = append(rankingInitActions, icfg.ResetW...)
	rankingInitActions = append(rankingInitActions, icfg.AssumeGAxioms...)
	rankingInitActions = append(rankingInitActions, icfg.AssumeInitAxioms...)
	rankingInitActions = append(rankingInitActions, setLineno(actions.NewAssumeAction(icfg.NotLf), lineno))

	if model.Init != nil {
		model.Init = actions.PostfixAction(model.Init, rankingInitActions)
	}

	// ---------------------------------------------------------------
	// Step 11: Replace named binders (shared)
	// ---------------------------------------------------------------
	xtracer.Trace("ranking.SharedStep11 ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))
	SharedStep11_ReplaceNamedBinders(icfg, model, modPass)
	xtracer.Trace("ranking.SharedStep11 EXIT nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))

	// ---------------------------------------------------------------
	// Step 12: Build new goal (shared)
	// ---------------------------------------------------------------
	xtracer.Trace("ranking.SharedStep12 ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPostconds=%d",
		len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), len(postconds))
	if tm != nil {
		// Python: conc = ivy_ast.TemporalModels(model, lg.And())
		// Pass the modified model, not tm (which has the original model).
		result, err := SharedStep12_BuildGoal(cfg.Mod.Cfg.AstCfg, goal, cfg.Goals, prems, model)
		if err != nil {
			return nil, err
		}
		// Python ivy_ranking.py:1016: goal.trace_hook = lambda tr,fcs: auto_hook(tasks,triggers,subs,tr,fcs)
		if len(result) > 0 && result[0] != nil {
			subs := icfg.Subs
			tasks := icfg.Tasks
			triggers := icfg.Triggers
			rsubs := icfg.RSubs
			fullSubs := icfg.FullSubs
			result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
				if subs != nil {
					applyRenamingToHandler(handler, subs)
				}
				applyAutoDiagnosticsToHandler(handler, fcs, tasks, triggers, rsubs, fullSubs)
			})
		}
		return result, nil
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
// Mirrors Python ivy_ranking.py:526-527 — uses LF.Clone so the resulting LFs
// retain their original ids and metadata flags.
func ModelPass(model *temporal.NormalProgram, transform func(ast.Node) ast.Node) {
	if model == nil {
		return
	}
	for i, inv := range model.Invars {
		if inv.Formula != nil {
			model.Invars[i] = transform(inv).(*ast.LabeledFormula)
		}
	}
	for i, asm := range model.Asms {
		if asm.Formula != nil {
			model.Asms[i] = transform(asm).(*ast.LabeledFormula)
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
	PreActions  []actions.ActionsAction
	PostActions []actions.ActionsAction
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

		oldG := oldL2sG(vs, body, gprop.Environ)
		curG := l2sG(vs, body, gprop.Environ)

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

		oldG := oldL2sG(vs, body, gprop.Environ)
		curG := l2sG(vs, body, gprop.Environ)

		// Pre: assume monotonicity: old_g(V) -> g(V)
		if oldG != nil && curG != nil {
			monoFmla := makeImplies(oldG, curG)
			pe.PreActions = append(pe.PreActions,
				makeAssumeForAll(vs, monoFmla, lineno))

			// Pre: assume ~old_g(V) & body -> ~g(V)
			newFmla := makeImplies(
				rankingMakeAnd(makeNot(oldG), body),
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
func WaitEvent(waits []*lg.NamedBinder, proofLabel string, lineno ast.Location) []actions.ActionsAction {
	var result []actions.ActionsAction
	for _, wait := range waits {
		vs := wait.Variables
		body := wait.Body

		// w(V) := w(V) & ~body & ~g(~body)
		wApp := wait // The waiting predicate itself
		negBody := makeNot(body)
		gNegBody := l2sG(vs, negBody, strPtr(proofLabel))

		newVal := rankingMakeAnd(wApp, negBody, makeNot(gNegBody))
		result = append(result, newAssignAction(wApp, newVal, lineno))
	}
	return result
}

// --- helper action constructors ---

func newAssignAction(lhs, rhs lg.Expr, lineno ast.Location) actions.ActionsAction {
	act := actions.NewAssignAction(lhs, rhs)
	act.SetLineno(lineno)
	return act
}

func newHavocAction(target lg.Expr, lineno ast.Location) actions.ActionsAction {
	act := actions.NewHavocAction(target)
	act.SetLineno(lineno)
	return act
}

func makeAssumeForAll(vs []*lg.Variable, body lg.Expr, lineno ast.Location) actions.ActionsAction {
	fmla := CheckForAll(vs, body)
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

// rankingMakeAnd is the ranking-tactic And constructor. It differs from
// l2s.go's makeAnd: l2s builds &lg.And{Terms: terms} directly, while ranking
// goes through lg.NewAnd which validates and may return lg.True on error.
func rankingMakeAnd(terms ...lg.Expr) lg.Expr {
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
// RankingDesugar is the ranking-tactic Desugar. The "Ranking" prefix
// distinguishes it from l2s.go's two-argument Desugar.
func RankingDesugar(expr lg.Expr, proofLabel string, l2sSaved lg.Expr) lg.Expr {
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
			return rankingMakeAnd(l2sSaved, applyWas(nb.Body, proofLabel))
		case "happened":
			if len(nb.Variables) > 0 {
				return expr
			}
			return rankingMakeAnd(l2sSaved, applyHappened(nb.Body, proofLabel))
		}
	}
	return cloneWithTransform(expr, func(n lg.Expr) lg.Expr {
		return RankingDesugar(n, proofLabel, l2sSaved)
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
		nb := l2sS(nil, expr, proofLabel)
		return nb
	}
}

func applyHappened(expr lg.Expr, proofLabel string) lg.Expr {
	// ~(l2s_w(V, body)(V))
	nb := l2sW(nil, expr, proofLabel)
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
		return RankingL2STactic(cfg)
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
		return l2sInit(nil, fmla, proofLabel)
	}
}
