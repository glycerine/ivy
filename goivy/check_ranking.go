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
package goivy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// --- Formula helpers ---

// CheckForAll wraps a body in a CheckForAll quantifier if there are variables.
// If vs is empty, returns body unchanged.
func CheckForAll(vs []*LogicVariable, body Expr) Expr {
	if len(vs) == 0 {
		return body
	}
	fa, err := NewForAll(vs, body)
	if err != nil {
		return body
	}
	return fa
}

// CheckExists wraps a body in an CheckExists quantifier if there are variables.
// If vs is empty, returns body unchanged.
func CheckExists(vs []*LogicVariable, body Expr) Expr {
	if len(vs) == 0 {
		return body
	}
	ex, err := NewExists(vs, body)
	if err != nil {
		return body
	}
	return ex
}

// CheckOldOf replaces function symbols in a formula with their "old" versions.
// For an application f(args), it returns old_f(args).
// For compound formulas, it recurses into sub-formulas.
func CheckOldOf(fmla Expr) Expr {
	if fmla == nil {
		return nil
	}
	switch f := fmla.(type) {
	case *Apply:
		oldFunc := makeOldFunc(f.Func)
		app, err := NewApply(oldFunc, f.Terms...)
		if err != nil {
			return fmla
		}
		return app
	case *Const:
		return NewConst(LogicOld(f.Name), f.CSort)
	case *Eq:
		return &Eq{T1: CheckOldOf(f.T1), T2: CheckOldOf(f.T2)}
	case *LogicNot:
		return &LogicNot{Body: CheckOldOf(f.Body)}
	case *LogicAnd:
		terms := make([]Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = CheckOldOf(t)
		}
		return &LogicAnd{Terms: terms}
	case *LogicOr:
		terms := make([]Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = CheckOldOf(t)
		}
		return &LogicOr{Terms: terms}
	case *LogicImplies:
		return &LogicImplies{T1: CheckOldOf(f.T1), T2: CheckOldOf(f.T2)}
	case *LogicIff:
		return &LogicIff{T1: CheckOldOf(f.T1), T2: CheckOldOf(f.T2)}
	case *ForAll:
		return &ForAll{Variables: f.Variables, Body: CheckOldOf(f.Body)}
	case *LogicExists:
		return &LogicExists{Variables: f.Variables, Body: CheckOldOf(f.Body)}
	default:
		return fmla
	}
}

func makeOldFunc(fn Expr) Expr {
	switch f := fn.(type) {
	case *Const:
		return NewConst(LogicOld(f.Name), f.CSort)
	default:
		return fn
	}
}

// --- Named binder constructors ---

// L2sD creates the l2s_d predicate for a sort.
// l2s_d : sort -> Boolean (tracks which domain elements are active).
func L2sD(sort Sort) *Const {
	fs, err := NewFunctionSort(sort, Boolean)
	if err != nil {
		return NewConst("l2s_d", Boolean)
	}
	return NewConst("l2s_d", fs)
}

// --- Task and Trigger types ---

// Task holds the ranking function definitions for one task suffix.
type Task struct {
	WorkCreated  Expr // definition of work_created
	WorkNeeded   Expr // definition of work_needed
	WorkProgress Expr // definition of work_progress
	WorkInvar    Expr // definition of work_invar
	WorkHelpful  Expr // definition of work_helpful
	WorkWitness  Expr // optional work_witness
}

// Trigger holds the trigger condition for a task.
type LogicTrigger struct {
	WorkStart Expr // definition of work_start
}

// --- L2S Tactic ---

// L2STacticConfig holds the configuration for the l2s tactic.
type L2STacticConfig struct {
	TacticName string
	ProofLabel string
	Mod        *Module
	Goals      []*LabeledFormula
	Proof      *LogicProofDecl
}

// ProofDecl represents a proof declaration with tactic parameters.
type LogicProofDecl struct {
	TacticName  string
	TacticDecls []Node
	TacticLets  []Node
	Lineno      Location
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
func RankingL2STactic(cfg *L2STacticConfig) ([]*LabeledFormula, error) {
	if cfg == nil || len(cfg.Goals) == 0 {
		return nil, fmt.Errorf("no goals provided")
	}

	goal := cfg.Goals[0]
	lineno := Location{Filename: "ranking", Line: 0}

	// Find the TemporalModels in the goal
	tm := FindTemporalModels(goal)
	if tm == nil {
		// Fall back to GoalConc check; only lg.Expr conclusions can be
		// "temporal formulas" (the alternative is *ast.TemporalModels which
		// FindTemporalModels would have caught above).
		conc := GoalConcExpr(goal)
		if conc == nil {
			return nil, fmt.Errorf("goal has no conclusion")
		}
		if !IsTemporalFormula(conc) && !isTemporalModels(conc) {
			return nil, fmt.Errorf("check/ranking: [1]proof goal is not temporal")
		}
	}

	if cfg.Proof != nil && len(cfg.Proof.TacticLets) > 0 {
		return nil, fmt.Errorf("tactic does not take lets")
	}

	m := cfg.Mod

	// Python ivy_ranking.py:95-107 — split TacticDecls, compile definitions into goal
	var tacticDefns []Node
	var tacticInvars []*LabeledFormula
	if cfg.Proof != nil {
		for _, d := range cfg.Proof.TacticDecls {
			if dd, ok := d.(*DerivedDecl); ok {
				tacticDefns = append(tacticDefns, dd)
			} else if lf, ok := d.(*LabeledFormula); ok {
				tacticInvars = append(tacticInvars, lf)
			}
		}
	}
	for _, defn := range tacticDefns {
		var compErr error
		goal, compErr = CompileDefinitionGoalVocab(m.Cfg.AstCfg, defn, goal, m)
		if compErr != nil {
			return nil, compErr
		}
	}

	// Extract the model and formula.
	// Python ivy_ranking.py:63: model = conc.model.clone([])
	// Use the model already embedded in the TemporalModels goal conclusion,
	// NOT NormalProgramFromModule(m), which emits a spurious XTRACE.
	var model *NormalProgram
	if tm != nil && tm.Model != nil {
		model = NormalProgramClone(tm.Model.(*NormalProgram))
	} else {
		model = ExtractNormalProgram(m)
	}

	var fmla Expr
	if tm != nil {
		fmla, _ = tm.Fmla.(Expr)
	}
	if fmla == nil {
		conc := GoalConc(goal)
		if conc != nil {
			fmla, _ = conc.(Expr)
		}
	}
	if fmla == nil {
		return nil, fmt.Errorf("cannot extract formula from goal conclusion")
	}

	// Process temporal premises
	prems := GoalPrems(goal)
	var temporalPrems []Expr
	for _, p := range prems {
		if lf, ok := p.(*LabeledFormula); ok {
			if lf.IsTemporal() {
				if f, ok := lf.Formula.(Expr); ok {
					temporalPrems = append(temporalPrems, f)
				}
			}
		}
	}
	if len(temporalPrems) > 0 {
		premConj := rankingMakeAnd(temporalPrems...)
		fmla = &LogicImplies{T1: premConj, T2: fmla}
	}

	proofLabel := cfg.ProofLabel

	finiteSorts, uninterpretedSorts := rankingFiniteSortsAndUninterpreted(m)

	// Add model invariants
	var invars []*LabeledFormula
	invars = append(invars, model.Invars...)

	// Python ivy_ranking.py:126 — compile user-supplied tactic invariants
	for idx, inv := range tacticInvars {
		xtracer.Trace("ranking.RankingL2STactic compileInvar[%d] pre-compile HASH canon=%v", idx, inv.Canon())
		vocab := GoalVocab(goal)
		compiledLF := CompileExprVocabExtLF(inv, vocab, m)
		if compiledLF == nil {
			xtracer.Trace("ranking.RankingL2STactic compileInvar[%d] compiled=nil", idx)
			continue
		}
		labeled := LabelTemporalNode(compiledLF, proofLabel).(*LabeledFormula)
		xtracer.Trace("ranking.RankingL2STactic compileInvar[%d] post-compile HASH canon=%v", idx, labeled.Canon())
		invars = append(invars, labeled)
	}

	// Generate ranking invariants and postconditions.
	// Pass outer prems so that the internal prem-update (Python ivy_ranking.py:261-269)
	// modifies the correct slice that modPass (line 467) closes over.
	invars, postconds, _, _, err := rankingInvariants(
		goal, invars, proofLabel, fmla,
		finiteSorts, uninterpretedSorts, m, prems)
	if err != nil {
		return nil, fmt.Errorf("ranking invariant generation: %w", err)
	}

	// Python ivy_ranking.py:433-457: convert_to_init + iinvs + neg_prop_init.
	// These LFs must be created BEFORE winvs and desugar to match Python's order.
	{
		knownInits := make(map[string]bool)
		var iinvs []Expr
		var localCTI func(f Expr) Expr
		localCTI = func(f Expr) Expr {
			switch n := f.(type) {
			case *LogicAnd:
				terms := make([]Expr, len(n.Terms))
				for i, t := range n.Terms {
					terms[i] = localCTI(t)
				}
				return &LogicAnd{Terms: terms}
			case *LogicOr:
				terms := make([]Expr, len(n.Terms))
				for i, t := range n.Terms {
					terms[i] = localCTI(t)
				}
				return &LogicOr{Terms: terms}
			case *LogicNot:
				return &LogicNot{Body: localCTI(n.Body)}
			case *LogicImplies:
				return &LogicImplies{T1: localCTI(n.T1), T2: localCTI(n.T2)}
			case *LogicIff:
				return &LogicIff{T1: localCTI(n.T1), T2: localCTI(n.T2)}
			case *ForAll:
				return &ForAll{Variables: n.Variables, Body: localCTI(n.Body)}
			case *LogicExists:
				return &LogicExists{Variables: n.Variables, Body: localCTI(n.Body)}
			default:
				vs := collectVarsSlice(f)
				ini := applyNB(l2sInit(vs, f, proofLabel), checkVarsToNodes(vs)...)
				key := string(f.Sexp())
				if _, ok := f.(*LogicGlobally); ok && !knownInits[key] {
					iinvs = append(iinvs, &LogicImplies{T1: ini, T2: f})
					knownInits[key] = true
				}
				if _, ok := f.(*LogicEventually); ok && !knownInits[key] {
					iinvs = append(iinvs, &LogicImplies{T1: f, T2: ini})
					knownInits[key] = true
				}
				return ini
			}
		}
		negPropInit := &LogicNot{Body: localCTI(fmla)}
		acfg := cfg.Mod.Cfg.AstCfg
		for i, iinv := range iinvs {
			invars = appendLF(acfg, invars, fmt.Sprintf("l2s_init_glob_%d", i), iinv, lineno)
		}
		invars = appendLF(acfg, invars, "neg_prop_init", negPropInit, lineno)
	}

	// Python ivy_ranking.py:461-471: WhenOperator invariant generation from pprems.
	// This must come BEFORE desugar to match Python's execution order.
	{
		var pprems []Expr
		for _, p := range prems {
			if lf, ok := p.(*LabeledFormula); ok && GoalIsProperty(lf) {
				if e, ok := lf.Formula.(Expr); ok {
					pprems = append(pprems, e)
				}
			}
		}
		var allFmlas []Expr
		for _, inv := range invars {
			if e, ok := inv.Formula.(Expr); ok {
				allFmlas = append(allFmlas, e)
			}
		}
		allFmlas = append(allFmlas, pprems...)
		seen := make(map[string]bool)
		var winvs []Expr
		for _, f := range allFmlas {
			for _, t := range TemporalsAst(f) {
				wo, ok := t.(*LogicWhenOperator)
				if !ok || wo.Name != "first" {
					continue
				}
				key := string(wo.Sexp())
				if seen[key] {
					continue
				}
				seen[key] = true
				nws := &LogicOr{Terms: []Expr{
					&LogicNot{Body: L2SWaiting()},
					&LogicNot{Body: applyNB(l2sW(nil, wo.T2, proofLabel))},
				}}
				tmp := &LogicImplies{
					T1: &LogicNot{Body: nws},
					T2: &Eq{T1: wo, T2: &LogicWhenOperator{Name: "next", T1: wo.T1, T2: wo.T2}},
				}
				tmp2 := &LogicImplies{
					T1: applyNB(l2sInit(nil, &LogicEventually{Environ: strPtr(proofLabel), Body: wo.T2}, proofLabel)),
					T2: tmp,
				}
				winvs = append(winvs, tmp2)
			}
		}
		acfg := cfg.Mod.Cfg.AstCfg
		for i, winv := range winvs {
			label := NewConst(fmt.Sprintf("l2s_when_%d", i), Boolean)
			invars = append(invars, acfg.NewLabeledFormula(label, winv))
		}
	}

	// Python ivy_ranking.py:515: desugar invars only (not postconds).
	// Use Clone (not NewLabeledFormula) to preserve LF id + metadata,
	// matching Python's `expr.clone(...)` recursion in desugar (ivy_ranking.py:505).
	// This comes AFTER neg_prop_init and winvs to match Python's execution order.
	{
		desugarFn := func(n Expr) (Expr, error) {
			return Desugar(n, proofLabel)
		}
		for i, inv := range invars {
			desugared, err := desugarFn(inv.Formula.(Expr))
			if err != nil {
				return nil, err
			}
			invars[i] = inv.Clone([]Node{inv.Label, desugared}).(*LabeledFormula)
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
	// Python ivy_ranking.py:113: defn_deps includes prover.definitions + prem_defns.
	// Pass goalPrems so BuildDefnDeps includes definition premises from the goal.
	defnDeps := BuildDefnDeps(m, GoalPrems(goal)...)

	icfg := &InstrumentationConfig{
		ProofLabel:         proofLabel,
		Lineno:             lineno,
		FiniteSorts:        finiteSorts,
		UninterpretedSorts: uninterpretedSorts,
		Mod:                m,
		Fmla:               fmla,
		Postconds:          postconds,
		Dependencies:       BuildDependenciesFunc(defnDeps),
		IsRankingTactic:    true,
	}

	// --- Model pass helper (ranking version: also transforms postconds) ---
	// Use Clone (not NewLabeledFormula) to preserve LF id + metadata,
	// matching Python ivy_ranking.py:526-527 where each `transform(x)` ends
	// up calling `LF.clone(args)` via the recursive `ast.clone` dispatch.
	modPass := func(transformName string, transform func(Node) Node) {
		if xtracer.Enabled {
			nPropPrems := 0
			for _, p := range prems {
				if lf, ok := p.(*LabeledFormula); ok && GoalIsProperty(lf) {
					nPropPrems++
				}
			}
			xtracer.Trace("ranking.modPass ENTER transform=%s nInvars=%d nAsms=%d nBindings=%d nPrems=%d nPropPrems=%d nPostconds=%d",
				transformName, len(model.Invars), len(model.Asms), len(model.Bindings), len(prems), nPropPrems, len(postconds))
		}
		for i, inv := range model.Invars {
			xtracer.Trace("ranking.modPass clone invar[%d] ENTER HASH canon=%v", i, inv.Canon())
			model.Invars[i] = transform(inv).(*LabeledFormula)
			xtracer.Trace("ranking.modPass clone invar[%d] EXIT HASH canon=%v", i, model.Invars[i].Canon())
		}
		for i, asm := range model.Asms {
			xtracer.Trace("ranking.modPass clone asm[%d] ENTER HASH canon=%v", i, asm.Canon())
			model.Asms[i] = transform(asm).(*LabeledFormula)
			xtracer.Trace("ranking.modPass clone asm[%d] EXIT HASH canon=%v", i, model.Asms[i].Canon())
		}
		for i, b := range model.Bindings {
			xtracer.Trace("ranking.modPass clone binding[%d] ENTER name=%s", i, b.Name)
			// Python: model.bindings[i] = b.clone([transform(b.action)])
			newAction := transform(b.Action).(*ActionTerm)
			model.Bindings[i] = b.CloneAction(newAction)
			xtracer.Trace("ranking.modPass clone binding[%d] EXIT name=%s", i, b.Name)
		}
		if model.Init != nil {
			xtracer.Trace("ranking.modPass clone init ENTER")
			// Python: model.init = transform(model.init)
			model.Init = transform(model.Init).(ActionsAction)
			xtracer.Trace("ranking.modPass clone init EXIT")
		}
		// Python ivy_ranking.py:532: list_transform(prems, transform)
		for i, p := range prems {
			lf, ok := p.(*LabeledFormula)
			if !ok || !GoalIsProperty(lf) {
				continue
			}
			xtracer.Trace("ranking.modPass clone prem[%d] ENTER HASH canon=%v", i, lf.Canon())
			prems[i] = transform(lf).(*LabeledFormula)
			xtracer.Trace("ranking.modPass clone prem[%d] EXIT HASH canon=%v", i, prems[i].(*LabeledFormula).Canon())
		}
		// Ranking-specific: also transform postconds
		for i, pc := range postconds {
			xtracer.Trace("ranking.modPass clone postcond[%d] ENTER HASH canon=%v", i, pc.Canon())
			postconds[i] = transform(pc).(*LabeledFormula)
			xtracer.Trace("ranking.modPass clone postcond[%d] EXIT HASH canon=%v", i, postconds[i].Canon())
		}
		if xtracer.Enabled {
			nPropPrems := 0
			for _, p := range prems {
				if lf, ok := p.(*LabeledFormula); ok && GoalIsProperty(lf) {
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
	var idleParts []ActionsAction
	idleParts = append(idleParts, icfg.AssumeGAxioms...)
	idleParts = append(idleParts, icfg.AssumeWhenAxioms...)
	idleParts = append(idleParts, icfg.ResetW...)
	idleParts = append(idleParts, icfg.AddConstsToD...)

	idleAction := setLineno(ConcatActions(idleParts...), lineno)
	idleAction.SetFormalParams(nil)
	idleAction.SetFormalReturns(nil)

	model.Bindings = append(model.Bindings, &ActionTermBinding{
		Name:   "_idle",
		Action: &ActionTerm{Stmt: idleAction},
	})
	model.Calls = append(model.Calls, "_idle")
	if model.Postconds == nil {
		model.Postconds = make(map[string][]*LabeledFormula)
	}
	model.Postconds["_idle"] = postconds

	// ---------------------------------------------------------------
	// Step 10: Init action (ranking-specific: no monitor state vars)
	// ---------------------------------------------------------------
	var rankingInitActions []ActionsAction
	rankingInitActions = append(rankingInitActions, icfg.AddConstsToD...)
	rankingInitActions = append(rankingInitActions, icfg.ResetW...)
	rankingInitActions = append(rankingInitActions, icfg.AssumeGAxioms...)
	rankingInitActions = append(rankingInitActions, icfg.AssumeInitAxioms...)
	rankingInitActions = append(rankingInitActions, setLineno(NewAssumeAction(icfg.NotLf), lineno))

	if model.Init != nil {
		model.Init = PostfixAction(model.Init, rankingInitActions)
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
			result[0].TraceHook = TraceHookFn(func(trace *Trace, fcs []Checker) *Trace {
				if subs != nil {
					trace = applyRenamingToTrace(trace, subs)
				}
				return applyAutoDiagnosticsToTrace(trace, fcs, tasks, triggers, rsubs, fullSubs)
			})
		}
		return result, nil
	}
	// Fallback: return goals as-is if no TemporalModels found
	return cfg.Goals, nil
}

func isTemporalModels(n Expr) bool {
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
func ModelPass(model *NormalProgram, transform func(Node) Node) {
	if model == nil {
		return
	}
	for i, inv := range model.Invars {
		if inv.Formula != nil {
			model.Invars[i] = transform(inv).(*LabeledFormula)
		}
	}
	for i, asm := range model.Asms {
		if asm.Formula != nil {
			model.Asms[i] = transform(asm).(*LabeledFormula)
		}
	}
}

func rankingFiniteSortsAndUninterpreted(m *Module) (map[string]bool, []Sort) {
	finiteSorts := make(map[string]bool)
	var uninterpretedSorts []Sort
	if m != nil && m.Sig != nil {
		for name, s := range m.Sig.Sorts.All() {
			if m.FiniteSorts[name] || isTheoryFiniteSort(name, m) {
				finiteSorts[name] = true
			} else if _, isUI := s.(*UninterpretedSort); isUI {
				uninterpretedSorts = append(uninterpretedSorts, s)
			}
		}
	}
	sort.Slice(uninterpretedSorts, func(i, j int) bool {
		return uninterpretedSorts[i].String() < uninterpretedSorts[j].String()
	})
	return finiteSorts, uninterpretedSorts
}

// --- Diagnostics ---

// DiagnosticInfo holds diagnostic information about a failed ranking check.
type DiagnosticInfo struct {
	Name    string
	Suffix  string
	Message string
}

// DiagnoseFailure provides diagnostic information when a ranking check fails.
func DiagnoseFailure(name string, tasks map[string]*Task, triggers map[string]*LogicTrigger) *DiagnosticInfo {
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
func TrigGlob(prop Expr, pos bool) []Expr {
	var result []Expr
	trigGlobRec(prop, pos, &result)
	return result
}

func trigGlobRec(prop Expr, pos bool, result *[]Expr) {
	if prop == nil {
		return
	}
	switch p := prop.(type) {
	case *LogicGlobally:
		if pos {
			*result = append(*result, p)
			trigGlobRec(p.Body, pos, result)
		}
	case *LogicEventually:
		if !pos {
			*result = append(*result, &LogicNot{Body: p})
			trigGlobRec(p.Body, pos, result)
		}
	case *LogicImplies:
		if !pos {
			trigGlobRec(p.T1, !pos, result)
			trigGlobRec(p.T2, pos, result)
		}
	case *LogicAnd:
		if pos {
			for _, t := range p.Terms {
				trigGlobRec(t, pos, result)
			}
		}
	case *LogicOr:
		if !pos {
			for _, t := range p.Terms {
				trigGlobRec(t, pos, result)
			}
		}
	case *ForAll:
		if pos {
			trigGlobRec(p.Body, pos, result)
		}
	case *LogicExists:
		if !pos {
			trigGlobRec(p.Body, pos, result)
		}
	case *LogicNot:
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
	PreActions  []ActionsAction
	PostActions []ActionsAction
}

// NewPropEvents creates the pre/post actions for a set of globally properties.
// For each property:
//   - Pre: save old value, havoc current value, assume monotonicity constraints
//   - Post: assume semantic constraint (g(V) -> body)
func NewPropEvents(gprops []*LogicNamedBinder, lineno Location) *PropEvent {
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
func WaitEvent(waits []*LogicNamedBinder, proofLabel string, lineno Location) []ActionsAction {
	var result []ActionsAction
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

func newAssignAction(lhs, rhs Expr, lineno Location) ActionsAction {
	act := NewAssignAction(lhs, rhs)
	act.SetLineno(lineno)
	return act
}

func newHavocAction(target Expr, lineno Location) ActionsAction {
	act := NewHavocAction(target)
	act.SetLineno(lineno)
	return act
}

func makeAssumeForAll(vs []*LogicVariable, body Expr, lineno Location) ActionsAction {
	fmla := CheckForAll(vs, body)
	act := NewAssumeAction(fmla)
	act.SetLineno(lineno)
	return act
}

func makeImplies(t1, t2 Expr) Expr {
	imp, err := NewImplies(t1, t2)
	if err != nil {
		return t2
	}
	return imp
}

// rankingMakeAnd is the ranking-tactic And constructor. It differs from
// l2s.go's makeAnd: l2s builds &lg.And{Terms: terms} directly, while ranking
// goes through lg.NewAnd which validates and may return lg.True on error.
func rankingMakeAnd(terms ...Expr) Expr {
	and, err := NewAnd(terms...)
	if err != nil {
		return True
	}
	return and
}

func makeNot(body Expr) Expr {
	not, err := NewNot(body)
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
func RankingDesugar(expr Expr, proofLabel string, l2sSaved Expr) Expr {
	if expr == nil {
		return nil
	}
	if nb, ok := expr.(*LogicNamedBinder); ok {
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
	return cloneWithTransform(expr, func(n Expr) Expr {
		return RankingDesugar(n, proofLabel, l2sSaved)
	})
}

func applyWas(expr Expr, proofLabel string) Expr {
	switch e := expr.(type) {
	case *LogicAnd:
		terms := make([]Expr, len(e.Terms))
		for i, t := range e.Terms {
			terms[i] = applyWas(t, proofLabel)
		}
		return &LogicAnd{Terms: terms}
	case *LogicOr:
		terms := make([]Expr, len(e.Terms))
		for i, t := range e.Terms {
			terms[i] = applyWas(t, proofLabel)
		}
		return &LogicOr{Terms: terms}
	case *LogicNot:
		return &LogicNot{Body: applyWas(e.Body, proofLabel)}
	case *LogicImplies:
		return &LogicImplies{T1: applyWas(e.T1, proofLabel), T2: applyWas(e.T2, proofLabel)}
	case *LogicIff:
		return &LogicIff{T1: applyWas(e.T1, proofLabel), T2: applyWas(e.T2, proofLabel)}
	default:
		// Atomic: create l2s_s binder
		nb := l2sS(nil, expr, proofLabel)
		return nb
	}
}

func applyHappened(expr Expr, proofLabel string) Expr {
	// ~(l2s_w(V, body)(V))
	nb := l2sW(nil, expr, proofLabel)
	return makeNot(nb)
}

// cloneWithTransform recursively applies transform to all children of a node.
func cloneWithTransform(node Expr, transform func(Expr) Expr) Expr {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *LogicAnd:
		terms := make([]Expr, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = transform(t)
		}
		return &LogicAnd{Terms: terms}
	case *LogicOr:
		terms := make([]Expr, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = transform(t)
		}
		return &LogicOr{Terms: terms}
	case *LogicNot:
		return &LogicNot{Body: transform(n.Body)}
	case *LogicImplies:
		return &LogicImplies{T1: transform(n.T1), T2: transform(n.T2)}
	case *LogicIff:
		return &LogicIff{T1: transform(n.T1), T2: transform(n.T2)}
	case *ForAll:
		return &ForAll{Variables: n.Variables, Body: transform(n.Body)}
	case *LogicExists:
		return &LogicExists{Variables: n.Variables, Body: transform(n.Body)}
	case *Eq:
		return &Eq{T1: transform(n.T1), T2: transform(n.T2)}
	default:
		return node
	}
}

// --- Registration ---

// RankingTactic adapts RankingL2STactic to the LogicProofTactic interface.
// Mirrors Python ivy_ranking.py:48-53 (l2s_tactic registered as 'ranking').
func RankingTactic(pc ProofCheckerInterface, goals []*LabeledFormula, pf Node) ([]*LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, fmt.Errorf("ranking: no proof goals")
	}
	m := pc.GetModule()
	if m == nil || m.Sig == nil {
		return nil, fmt.Errorf("ranking: module or sig is nil")
	}
	vocab := GoalVocab(goals[0])
	ws := NewWithSymbols(m.Sig, vocab.Symbols)
	ws.Enter()
	defer ws.Exit()
	wso := NewWithSorts(m.Sig, vocab.Sorts)
	wso.Enter()
	defer wso.Exit()

	var proofDecl *LogicProofDecl
	if tt, ok := pf.(*TacticTactic); ok {
		proofDecl = &LogicProofDecl{
			TacticLets:  tt.TacticLetsList(),
			TacticDecls: tt.TacticDeclsList(),
		}
	}
	cfg := &L2STacticConfig{
		TacticName: "ranking",
		Goals:      goals,
		Mod:        m,
		ProofLabel: "",
		Proof:      proofDecl,
	}
	return RankingL2STactic(cfg)
}

// TacticFunc is the signature for tactic implementations.
type TacticFunc func(cfg *L2STacticConfig) ([]*LabeledFormula, error)

// RegisteredTactics maps tactic names to implementations.
var RegisteredTactics = map[string]TacticFunc{
	"ranking": func(cfg *L2STacticConfig) ([]*LabeledFormula, error) {
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
func ConvertToInit(fmla Expr, proofLabel string) Expr {
	if fmla == nil {
		return nil
	}
	switch f := fmla.(type) {
	case *LogicAnd:
		terms := make([]Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = ConvertToInit(t, proofLabel)
		}
		return &LogicAnd{Terms: terms}
	case *LogicOr:
		terms := make([]Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = ConvertToInit(t, proofLabel)
		}
		return &LogicOr{Terms: terms}
	case *LogicNot:
		return &LogicNot{Body: ConvertToInit(f.Body, proofLabel)}
	case *LogicImplies:
		return &LogicImplies{
			T1: ConvertToInit(f.T1, proofLabel),
			T2: ConvertToInit(f.T2, proofLabel),
		}
	case *LogicIff:
		return &LogicIff{
			T1: ConvertToInit(f.T1, proofLabel),
			T2: ConvertToInit(f.T2, proofLabel),
		}
	case *ForAll:
		return &ForAll{Variables: f.Variables, Body: ConvertToInit(f.Body, proofLabel)}
	case *LogicExists:
		return &LogicExists{Variables: f.Variables, Body: ConvertToInit(f.Body, proofLabel)}
	default:
		// Wrap in l2s_init
		return l2sInit(nil, fmla, proofLabel)
	}
}
