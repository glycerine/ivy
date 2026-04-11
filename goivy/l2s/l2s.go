// Package l2s implements liveness-to-safety reduction for temporal property verification.
//
// This is a faithful port of Python's ivy_l2s.py. It transforms temporal (liveness)
// properties into safety properties that can be checked with standard
// IC3/UPDR verification.
//
// The transformation works by:
// 1. Constructing a monitor that tracks whether the temporal property holds
// 2. Adding saved-state copies of all relevant state
// 3. Adding Skolem constants/functions for existentially quantified variables
// 4. Adding fairness constraints
// 5. Proving that the resulting safety property implies the temporal one
//
// Key entry points:
//   - L2STactic: The main tactic for "l2s" proof goals
//   - L2STacticFull: Full version including auxiliary state
//   - L2STacticAuto: Automatic version with trigger inference
package l2s

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/compiler"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/temporal"
	theory "github.com/glycerine/ivy/goivy/theory"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// --- Named constants used by the L2S transformation ---

// L2SWaiting is the "l2s_waiting" boolean flag.
func L2SWaiting() *lg.Const {
	return lg.NewConst("l2s_waiting", lg.Boolean)
}

// L2SFrozen is the "l2s_frozen" boolean flag.
func L2SFrozen() *lg.Const {
	return lg.NewConst("l2s_frozen", lg.Boolean)
}

// L2SSaved is the "l2s_saved" boolean flag.
func L2SSaved() *lg.Const {
	return lg.NewConst("l2s_saved", lg.Boolean)
}

// L2SD creates the l2s_d predicate for a sort (domain tracking).
func L2SD(s lg.Sort) *lg.Const {
	return lg.NewConst("l2s_d", il.RelationSort([]lg.Sort{s}))
}

// L2SA creates the l2s_a predicate for a sort (abstract domain).
func L2SA(s lg.Sort) *lg.Const {
	return lg.NewConst("l2s_a", il.RelationSort([]lg.Sort{s}))
}

// l2sW creates an l2s_w (waited) named binder.
func l2sW(vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_w", Variables: vs, Environ: strPtr(label), Body: t}
}

// l2sS creates an l2s_s (saved) named binder.
func l2sS(vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_s", Variables: vs, Environ: strPtr(label), Body: t}
}

// l2sG creates an l2s_g (globally/safety) named binder.
func l2sG(vs []*lg.Variable, t lg.Expr, environ *string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_g", Variables: vs, Environ: environ, Body: t}
}

// oldL2sG creates an _old_l2s_g named binder.
func oldL2sG(vs []*lg.Variable, t lg.Expr, environ *string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "_old_l2s_g", Variables: vs, Environ: environ, Body: t}
}

// l2sInit creates an l2s_init named binder.
func l2sInit(vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_init", Variables: vs, Environ: strPtr(label), Body: t}
}

// l2sWhen creates an l2s_when<name> named binder.
func l2sWhen(name string, vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_when" + name, Variables: vs, Environ: strPtr(label), Body: t}
}

// l2sOld creates an l2s_old named binder.
func l2sOld(vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_old", Variables: vs, Environ: strPtr(label), Body: t}
}

// --- Helpers ---

func applyNB(nb *lg.NamedBinder, args ...lg.Expr) lg.Expr {
	if len(args) == 0 {
		return nb
	}
	return lg.MustApply(nb, args...)
}

func mustApply(f lg.Expr, args ...lg.Expr) lg.Expr {
	if len(args) == 0 {
		return f
	}
	return lg.MustApply(f, args...)
}

func varsToNodes(vs []*lg.Variable) []lg.Expr {
	nodes := make([]lg.Expr, len(vs))
	for i, v := range vs {
		nodes[i] = v
	}
	return nodes
}

func forall(vs []*lg.Variable, body lg.Expr) lg.Expr {
	if len(vs) == 0 {
		return body
	}
	return &lg.ForAll{Variables: vs, Body: body}
}

func exists(vs []*lg.Variable, body lg.Expr) lg.Expr {
	if len(vs) == 0 {
		return body
	}
	return &lg.Exists{Variables: vs, Body: body}
}

func strPtr(s string) *string { return &s }

func makeAnd(terms ...lg.Expr) lg.Expr {
	if len(terms) == 0 {
		return lg.True
	}
	if len(terms) == 1 {
		return terms[0]
	}
	return &lg.And{Terms: terms}
}

func setLineno(a actions.Action, loc ast.Location) actions.Action {
	a.SetLineno(loc)
	return a
}

// --- l2s_g tracking ---

type l2sGTriple struct {
	Vars    []*lg.Variable
	Body    lg.Expr
	Environ *string
}

func (t l2sGTriple) key() string {
	return fmt.Sprintf("%v|%v|%v", t.Vars, t.Body, t.Environ)
}

// --- varBodyPair ---

type varBodyPair struct {
	Vars []*lg.Variable
	Body lg.Expr
}

func dedupeVarBodyPairs(pairs []varBodyPair) []varBodyPair {
	seen := make(map[string]bool)
	var result []varBodyPair
	for _, p := range pairs {
		key := fmt.Sprintf("%v:%v", p.Vars, p.Body)
		if !seen[key] {
			seen[key] = true
			result = append(result, p)
		}
	}
	return result
}

// --- Tactic entry points ---

// l2sTactic mirrors Python's l2s_tactic (ivy_l2s.py:75-79).
// It enters WithSymbols and WithSorts contexts before calling l2sTacticInt.
// All three public entry points (L2STactic, L2STacticFull, L2STacticAuto)
// route through this helper.
func l2sTactic(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node, tacticName string) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, fmt.Errorf("l2s: no proof goals")
	}
	m := pc.GetModule()
	if m == nil || m.Sig == nil {
		return nil, fmt.Errorf("l2s: module or sig is nil")
	}
	vocab := proof.GoalVocab(goals[0])
	ws := il.NewWithSymbols(m.Sig, vocab.Symbols)
	ws.Enter()
	defer ws.Exit()
	wso := il.NewWithSorts(m.Sig, vocab.Sorts)
	wso.Enter()
	defer wso.Exit()
	return l2sTacticInt(pc, goals, pf, tacticName)
}

// L2STactic is the main tactic for "l2s" proof goals.
func L2STactic(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	return l2sTactic(pc, goals, pf, "l2s")
}

// L2STacticFull includes all auxiliary state in the transformation.
func L2STacticFull(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	return l2sTactic(pc, goals, pf, "l2s_full")
}

// L2STacticAuto uses automatic trigger inference. It extracts the user's
// actual tactic name (l2s_auto, l2s_auto2, l2s_auto3, l2s_auto4, l2s_auto5)
// from the proof object, mirroring Python's l2s_tactic_auto (ivy_l2s.py:91-93)
// which passes proof.tactic_name through.
func L2STacticAuto(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	tacticName := "l2s_auto" // fallback
	if tt, ok := pf.(*ast.TacticTactic); ok && tt.TName != nil {
		tacticName = l2sNodeToString(tt.TName)
	}
	return l2sTactic(pc, goals, pf, tacticName)
}

// l2sNodeToString extracts a string name from an AST node (mirrors
// proof.nodeToString which is unexported).
func l2sNodeToString(n ast.Node) string {
	if n == nil {
		return ""
	}
	if a, ok := n.(*ast.Atom); ok {
		return a.Relname()
	}
	return fmt.Sprint(n)
}

// pyBool formats a Go bool as Python's True/False, for trace parity.
func pyBool(b bool) string {
	if b {
		return "True"
	}
	return "False"
}

// isTheoryFiniteSort returns true if the named sort has a theory
// interpretation that is finite (e.g. bv[N]).
// C17 / Python ivy_l2s.py:194 calls thy.get_sort_theory(sort).is_finite().
func isTheoryFiniteSort(name string, m *module.Module) bool {
	if m == nil || m.Sig == nil {
		return false
	}
	s, ok := m.Sig.Sorts[name]
	if !ok {
		return false
	}
	th := theory.GetSortTheory(s, m.Sig.Interp)
	if t, ok := th.(*theory.Theory); ok {
		return t.Finite
	}
	return false
}

// l2sTacticInt is the internal implementation of the L2S tactic.
// Faithful port of Python's l2s_tactic_int.
func l2sTacticInt(pc module.ProofCheckerInterface, goals []*ast.LabeledFormula, pf ast.Node, tacticName string) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, fmt.Errorf("l2s: no proof goals")
	}

	full := tacticName == "l2s_full"
	goal := goals[0]
	xtracer.Trace("l2s.l2sTacticInt ENTER tactic='%s' ngoals=%d goal.Formula type=%s", tacticName, len(goals), iu.TypeName(goal.Formula))
	if goal.Formula != nil {
		// Also check if it's a SchemaBody
		if sb, ok := goal.Formula.(*ast.SchemaBody); ok {
			conc := sb.Conc()
			xtracer.Trace("l2s.l2sTacticInt goal.Formula is SchemaBody, conc type=%s", iu.TypeName(conc))
		}
	}
	lineno := ast.Location{Filename: "l2s", Line: 0}
	// Get the goal's conclusion. After the TemporalModels API broadening,
	// proof.GoalConc returns the inner conclusion (TemporalModels in this case)
	// regardless of whether the goal's Formula is a SchemaBody or a direct
	// TemporalModels. Mirrors Python ivy_l2s.py:117-118.
	conc := proof.GoalConc(goal)
	tm, isTM := conc.(*ast.TemporalModels)
	xtracer.Trace("l2s.l2sTacticInt goalConc result type=%s (isTemporalModels=%s)", iu.TypeName(conc), pyBool(isTM))
	if !isTM {
		return nil, fmt.Errorf("l2s: proof goal is not temporal")
	}

	// Extract the model (NormalProgram) and formula
	m := pc.GetModule()
	model := extractNormalProgram(m)
	fmla, _ := tm.Fmla.(lg.Expr)
	if fmla == nil {
		return nil, fmt.Errorf("l2s: could not extract temporal formula from goal")
	}

	// Get temporal premises
	prems := proof.GoalPrems(goal)
	var temporalPrems []lg.Expr
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.IsTemporal() { // Temporal is a *bool, nil means non-temporal
				if f, ok := lf.Formula.(lg.Expr); ok {
					temporalPrems = append(temporalPrems, f)
				}
			}
		}
	}
	// C10: also include non-explicit temporal axioms in temporalPrems
	// (mirrors Python ivy_l2s.py:134-135).
	if pc != nil {
		for _, ax := range pc.GetAxioms() {
			if !ax.Explicit && ax.IsTemporal() {
				if f, ok := ax.Formula.(lg.Expr); ok {
					temporalPrems = append(temporalPrems, f)
				}
			}
		}
	}

	// Add assumed globally properties to model assumptions
	// H13: preserve the original axiom label (Python ivy_l2s.py:132).
	if pc != nil {
		for _, ax := range pc.GetAxioms() {
			if !ax.Explicit && ax.IsTemporal() {
				if f, ok := ax.Formula.(lg.Expr); ok {
					if g, ok := f.(*lg.Globally); ok {
						model.Asms = append(model.Asms, m.Cfg.AstCfg.NewLabeledFormula(ax.Label, g.Body))
					}
				}
			}
		}
	}

	// If there are temporal premises, wrap: prems -> fmla
	if len(temporalPrems) > 0 {
		premConj := makeAnd(temporalPrems...)
		fmla = &lg.Implies{T1: premConj, T2: fmla}
	}

	proofLabel := ""

	// --- Invariants ---
	// C6/C8: invars holds tactic-level invariants (user-supplied + auto-generated).
	// At the end of this function we commit invars into model.Invars (Python
	// ivy_l2s.py:722). Do NOT seed from model.Invars — that's the original
	// C6/C8 bug.
	var invars []*ast.LabeledFormula

	// C6/C7/M1: process user-supplied tactic_decls.
	// Python ivy_l2s.py:124-125, 141-153, 175.
	if tt, ok := pf.(*ast.TacticTactic); ok {
		// M1: reject tactic_lets (Python line 124-125).
		if tt.Body != nil {
			if _, isLets := tt.Body.(*ast.TacticLets); isLets {
				return nil, fmt.Errorf("tactic does not take lets")
			}
		}

		decls := tt.TacticDeclsList()
		var tacticInvars []*ast.LabeledFormula
		var tacticDefns []ast.Node
		for _, d := range decls {
			if dd, isDerived := d.(*ast.DerivedDecl); isDerived {
				tacticDefns = append(tacticDefns, dd)
			} else if lf, isLF := d.(*ast.LabeledFormula); isLF {
				tacticInvars = append(tacticInvars, lf)
			}
		}

		// C7: compile definitions into goal premises (Python lines 152-153).
		for _, defn := range tacticDefns {
			goal = proof.CompileDefinitionGoalVocab(m.Cfg.AstCfg, defn, goal)
		}

		// C6: compile user invariants and seed `invars` (Python line 175).
		for _, inv := range tacticInvars {
			compiled := proof.CompileWithGoalVocab(inv.Formula, goal)
			if compiled == nil {
				continue
			}
			labeled := il.LabelTemporal(compiled, proofLabel)
			invars = append(invars, m.Cfg.AstCfg.NewLabeledFormula(inv.Label, labeled))
		}
	}

	// --- L2S monitor symbols ---
	l2sWaitingSym := L2SWaiting()
	l2sFrozenSym := L2SFrozen()
	l2sSavedSym := L2SSaved()

	// --- Finite sorts ---
	// C17 / Python ivy_l2s.py:194: include sorts whose theory is finite
	// (e.g. bv[N]) in addition to mod.FiniteSorts and the `full` override.
	finiteSorts := make(map[string]bool)
	var uninterpretedSorts []lg.Sort
	if m != nil && m.Sig != nil {
		for name, s := range m.Sig.Sorts {
			if m.FiniteSorts[name] || full || isTheoryFiniteSort(name, m) {
				finiteSorts[name] = true
			} else if _, isUI := s.(*lg.UninterpretedSort); isUI {
				uninterpretedSorts = append(uninterpretedSorts, s)
			}
		}
	}
	sort.Slice(uninterpretedSorts, func(i, j int) bool {
		return uninterpretedSorts[i].String() < uninterpretedSorts[j].String()
	})

	// ---------------------------------------------------------------
	// L2S Auto: generate task/trigger invariants (before main steps)
	// ---------------------------------------------------------------
	var autoTasks, autoTriggers map[string]map[string]*lg.Eq
	if strings.HasPrefix(tacticName, "l2s_auto") {
		var err error
		invars, autoTasks, autoTriggers, err = l2sAutoInvariants(tacticName, goal, invars, proofLabel,
			fmla, finiteSorts, uninterpretedSorts, m)
		if err != nil {
			return nil, fmt.Errorf("l2s_auto: %w", err)
		}
	}

	// C9: desugar $was/$happened operators in invars (Python ivy_l2s.py:716)
	l2sAcfg := m.Cfg.AstCfg
	for i, inv := range invars {
		if expr, ok := inv.Formula.(lg.Expr); ok {
			invars[i] = l2sAcfg.NewLabeledFormula(inv.Label, Desugar(expr, proofLabel))
		}
	}

	// C8: commit auto-generated/user invariants into model.Invars
	// (Python ivy_l2s.py:722 `model.invars = model.invars + invars`).
	model.Invars = append(model.Invars, invars...)

	// M7 / Python ivy_l2s.py:724: re-fetch prems AFTER tactic_decls have
	// mutated the goal, so the resulting prems include user-supplied
	// definition premises. modPass below transforms these in place.
	prems = proof.GoalPrems(goal)

	// --- Build shared config ---
	// H12: include user-supplied definition premises from the goal in defnDeps.
	defnDeps := BuildDefnDeps(m, prems...)

	cfg := &InstrumentationConfig{
		ProofLabel:         proofLabel,
		Lineno:             lineno,
		FiniteSorts:        finiteSorts,
		UninterpretedSorts: uninterpretedSorts,
		Mod:                m,
		Fmla:               fmla,
		Postconds:          nil, // l2s has no postconds
		Dependencies:       BuildDependenciesFunc(defnDeps),
		Tasks:              autoTasks,    // C5: for trace_hook routing
		Triggers:           autoTriggers, // C5: for trace_hook routing
	}

	// --- Model pass helper (l2s version: no postconds) ---
	// All tactic-generated invariants now live in model.Invars (per C8 above),
	// so modPass only iterates model.Invars (no separate local invars loop).
	// M7 / Python ivy_l2s.py:744: also transform property prems via list_transform.
	modPass := func(transform func(lg.Expr) lg.Expr) {
		for i, inv := range model.Invars {
			model.Invars[i] = l2sAcfg.NewLabeledFormula(inv.Label, transform(inv.Formula.(lg.Expr)))
		}
		for i, asm := range model.Asms {
			model.Asms[i] = l2sAcfg.NewLabeledFormula(asm.Label, transform(asm.Formula.(lg.Expr)))
		}
		for i, b := range model.Bindings {
			newStmt := transformAction(b.Action.Stmt, transform)
			model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
		}
		if model.Init != nil {
			model.Init = transformAction(model.Init, transform)
		}
		// M7: list_transform on property prems.
		for i, p := range prems {
			lf, ok := p.(*ast.LabeledFormula)
			if !ok || !proof.GoalIsProperty(lf) {
				continue
			}
			if e, ok := lf.Formula.(lg.Expr); ok {
				prems[i] = l2sAcfg.NewLabeledFormula(lf.Label, transform(e))
			}
		}
	}

	// ---------------------------------------------------------------
	// Step 1: Convert temporal operators to named binders (shared)
	// ---------------------------------------------------------------
	SharedStep1_ConvertTemporals(cfg, model, modPass)

	if cfg.Mod.Cfg.L2SDebug {
		fmt.Println(strings.Repeat("=", 80) + "\nafter replace_temporals_by_named_binder_g_ast")
		for _, triple := range cfg.L2sGs {
			fmt.Printf("l2s_g: %v %v %v\n", triple.Vars, triple.Body, triple.Environ)
		}
		fmt.Println(strings.Repeat("=", 80))
	}

	// ---------------------------------------------------------------
	// Step 2: Build monitor building blocks (l2s-specific)
	// ---------------------------------------------------------------

	// reset_a: l2s_a(s)(X) := l2s_d(s)(X) for each uninterpreted sort
	var resetA []actions.Action
	for _, s := range uninterpretedSorts {
		v, _ := lg.NewVariable("X", s)
		resetA = append(resetA,
			setLineno(actions.NewAssignAction(mustApply(L2SA(s), v), mustApply(L2SD(s), v)), lineno))
	}

	// add_consts_to_d
	cfg.AddConstsToD = BuildAddConstsToD(m, uninterpretedSorts, lineno)

	// ---------------------------------------------------------------
	// Step 3: Collect used l2s_w and l2s_s from conjectures (shared)
	// ---------------------------------------------------------------
	SharedStep3_CollectNamedBinders(cfg, model, full)

	// Build save/wait/reset_w from collected binders
	SharedBuildSaveAndWait(cfg)

	// ---------------------------------------------------------------
	// Step 4: Fair cycle check (l2s-specific)
	// ---------------------------------------------------------------

	fairCycle := []lg.Expr{l2sSavedSym}
	fairCycle = append(fairCycle, cfg.DoneWaiting...)

	// Projection of relations
	for _, vb := range cfg.ToSave {
		bodySort := vb.Body.NodeSort()
		isRelation := lg.SortEqual(bodySort, lg.Boolean)
		if fs, ok := bodySort.(*lg.FunctionSort); ok && lg.SortEqual(fs.Range(), lg.Boolean) {
			isRelation = true
		}
		if isRelation {
			savedApp := applyNB(l2sS(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
			iff := &lg.Iff{T1: savedApp, T2: vb.Body}
			if len(vb.Vars) > 0 {
				var aConjs []lg.Expr
				for _, v := range vb.Vars {
					if !finiteSorts[v.VSort.String()] {
						aConjs = append(aConjs, mustApply(L2SA(v.VSort), v))
					}
				}
				if len(aConjs) > 0 {
					fairCycle = append(fairCycle,
						forall(vb.Vars, &lg.Implies{T1: makeAnd(aConjs...), T2: iff}))
				} else {
					fairCycle = append(fairCycle, forall(vb.Vars, iff))
				}
			} else {
				fairCycle = append(fairCycle, iff)
			}
		}
	}

	// Projection of functions/constants (uninterpreted-sort valued)
	for _, vb := range cfg.ToSave {
		bodySort := vb.Body.NodeSort()
		isUninterp := false
		if _, ok := bodySort.(*lg.UninterpretedSort); ok {
			isUninterp = true
		}
		if fs, ok := bodySort.(*lg.FunctionSort); ok {
			if _, ok := fs.Range().(*lg.UninterpretedSort); ok {
				isUninterp = true
			}
		}
		if isUninterp {
			savedApp := applyNB(l2sS(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
			eq := &lg.Eq{T1: savedApp, T2: vb.Body}
			var aConjs []lg.Expr
			for _, v := range vb.Vars {
				if !finiteSorts[v.VSort.String()] {
					aConjs = append(aConjs, mustApply(L2SA(v.VSort), v))
				}
			}
			if !finiteSorts[bodySort.String()] {
				aConjs = append(aConjs,
					&lg.Or{Terms: []lg.Expr{
						mustApply(L2SA(bodySort), savedApp),
						mustApply(L2SA(bodySort), vb.Body),
					}})
			}
			if len(aConjs) > 0 {
				fairCycle = append(fairCycle,
					forall(vb.Vars, &lg.Implies{T1: makeAnd(aConjs...), T2: eq}))
			} else {
				fairCycle = append(fairCycle, forall(vb.Vars, eq))
			}
		}
	}

	var assertNoFairCycleAction actions.Action = setLineno(
		actions.NewAssertAction(&lg.Not{Body: makeAnd(fairCycle...)}), lineno)

	// H1 / Python ivy_l2s.py:910-911: if the user supplied a tactic_proof
	// (e.g. `tactic l2s_auto2 proof { <subproof> }`), apply it to the
	// no-fair-cycle assertion.
	if tt, ok := pf.(*ast.TacticTactic); ok {
		if tp := tt.TacticProofNode(); tp != nil {
			if assertAct, ok := assertNoFairCycleAction.(*actions.AssertAction); ok {
				assertNoFairCycleAction = compiler.ApplyAssertProofWith(m, assertAct, tp, pc)
			}
		}
	}
	assertNoFairCycle := assertNoFairCycleAction

	// ---------------------------------------------------------------
	// Step 5: Monitor state machine (l2s-specific)
	// ---------------------------------------------------------------

	monitorEdge := func(s1, s2 *lg.Const) []actions.Action {
		return []actions.Action{
			setLineno(actions.NewAssumeAction(s1), lineno),
			setLineno(actions.NewAssignAction(s1, lg.False), lineno),
			setLineno(actions.NewAssignAction(s2, lg.True), lineno),
		}
	}

	// waiting -> frozen
	var waitToFrozenParts []actions.Action
	waitToFrozenParts = append(waitToFrozenParts, monitorEdge(l2sWaitingSym, l2sFrozenSym)...)
	for _, dw := range cfg.DoneWaiting {
		waitToFrozenParts = append(waitToFrozenParts, setLineno(actions.NewAssumeAction(dw), lineno))
	}
	waitToFrozenParts = append(waitToFrozenParts, resetA...)

	// frozen -> saved
	var frozenToSavedParts []actions.Action
	frozenToSavedParts = append(frozenToSavedParts, monitorEdge(l2sFrozenSym, l2sSavedSym)...)
	frozenToSavedParts = append(frozenToSavedParts, cfg.SaveState...)
	frozenToSavedParts = append(frozenToSavedParts, cfg.ResetW...)

	changeMonitorState := []actions.Action{
		setLineno(actions.NewChoiceAction(
			setLineno(actions.ConcatActions(waitToFrozenParts...), lineno),
			setLineno(actions.ConcatActions(frozenToSavedParts...), lineno),
			setLineno(actions.NewSequence(), lineno),
		), lineno),
	}

	// ---------------------------------------------------------------
	// Step 6: Tableau construction (shared)
	// ---------------------------------------------------------------
	SharedStep6_BuildTableau(cfg)

	// ---------------------------------------------------------------
	// Step 7: Action instrumentation (shared)
	// ---------------------------------------------------------------
	SharedStep7_InstrumentActions(cfg, model)

	// ---------------------------------------------------------------
	// Step 8: Patch exported actions (shared, l2s mode: no postconds)
	// ---------------------------------------------------------------
	SharedStep8_PatchExports(cfg, model)

	// ---------------------------------------------------------------
	// Step 9: Idle action (l2s-specific)
	// ---------------------------------------------------------------

	var idleParts []actions.Action
	idleParts = append(idleParts, changeMonitorState...)
	idleParts = append(idleParts, cfg.AssumeGAxioms...)
	idleParts = append(idleParts, cfg.AddConstsToD...)
	idleParts = append(idleParts, assertNoFairCycle)

	idleAction := setLineno(actions.ConcatActions(idleParts...), lineno)
	idleAction.SetFormalParams(nil)
	idleAction.SetFormalReturns(nil)

	model.Bindings = append(model.Bindings, &temporal.ActionTermBinding{
		Name:   "idle",
		Action: &temporal.ActionTerm{Stmt: idleAction},
	})
	model.Calls = append(model.Calls, "idle")

	// ---------------------------------------------------------------
	// Step 10: Init action (l2s-specific)
	// ---------------------------------------------------------------

	var l2sInitActions []actions.Action
	l2sInitActions = append(l2sInitActions,
		setLineno(actions.NewAssignAction(l2sWaitingSym, lg.True), lineno),
		setLineno(actions.NewAssignAction(l2sFrozenSym, lg.False), lineno),
		setLineno(actions.NewAssignAction(l2sSavedSym, lg.False), lineno),
	)
	l2sInitActions = append(l2sInitActions, cfg.AddConstsToD...)
	l2sInitActions = append(l2sInitActions, cfg.ResetW...)
	l2sInitActions = append(l2sInitActions, cfg.AssumeGAxioms...)
	l2sInitActions = append(l2sInitActions, cfg.AssumeInitAxioms...)
	l2sInitActions = append(l2sInitActions, setLineno(actions.NewAssumeAction(cfg.NotLf), lineno))

	if model.Init != nil {
		model.Init = actions.PostfixAction(model.Init, l2sInitActions)
	}

	// ---------------------------------------------------------------
	// Step 11: Replace named binders (shared)
	// ---------------------------------------------------------------
	SharedStep11_ReplaceNamedBinders(cfg, model, modPass)

	// M2 / Python ivy_l2s.py:1308: remove unused definitions from goal.
	goal = proof.RemoveUnusedDefinitionsGoal(m.Cfg.AstCfg, goal)

	// ---------------------------------------------------------------
	// Step 12: Build new goal (shared)
	// ---------------------------------------------------------------
	result, err := SharedStep12_BuildGoal(pc.GetAstCfg(), goal, goals, prems, tm)
	if err != nil {
		return nil, err
	}

	// C5 / Python ivy_l2s.py:1310-1313: attach a trace hook to the result
	// goal so that the check package can route diagnostics on failure.
	// The hook is opaque (interface{}); check/ type-asserts it to
	// *l2s.L2STraceHookData.
	if len(result) > 0 && result[0] != nil {
		if strings.HasPrefix(tacticName, "l2s_auto5") {
			result[0].TraceHook = &L2STraceHookData{
				Kind:     HookKindAuto,
				Subs:     cfg.Subs,
				Tasks:    cfg.Tasks,
				Triggers: cfg.Triggers,
			}
		} else if strings.HasPrefix(tacticName, "l2s_auto") {
			result[0].TraceHook = &L2STraceHookData{
				Kind: HookKindRenaming,
				Subs: cfg.Subs,
			}
		} else if tacticName == "l2s_full" {
			result[0].TraceHook = &L2STraceHookData{
				Kind: HookKindFull,
			}
		}
	}
	return result, nil
}

// --- Internal helpers ---

// findTemporalModels looks through the goal formula for a TemporalModels node.
func findTemporalModels(goal *ast.LabeledFormula) *ast.TemporalModels {
	if goal == nil || goal.Formula == nil {
		return nil
	}
	// Check the formula directly
	if tm, ok := goal.Formula.(*ast.TemporalModels); ok {
		return tm
	}
	// Check if it's a SchemaBody and the conclusion is TemporalModels
	if sb, ok := goal.Formula.(*ast.SchemaBody); ok {
		conc := sb.Conc()
		if tm, ok := conc.(*ast.TemporalModels); ok {
			return tm
		}
	}
	return nil
}

// (cloneGoalWithASTConc was deleted — proof.CloneGoal now takes ast.Node
// for conc and supersedes this duplicate helper.)

func transformAction(act actions.Action, transform func(lg.Expr) lg.Expr) actions.Action {
	if act == nil {
		return nil
	}
	args := act.ActionArgs()
	newArgs := make([]lg.Expr, len(args))
	changed := false
	for i, a := range args {
		if sub, ok := a.(actions.Action); ok {
			newSub := transformAction(sub, transform)
			newArgs[i] = newSub
			if newSub != sub {
				changed = true
			}
		} else if a != nil {
			newA := transform(a)
			newArgs[i] = newA
			if newA != a {
				changed = true
			}
		} else {
			newArgs[i] = a
		}
	}
	if changed {
		return act.ActionClone(newArgs)
	}
	return act
}

func extractNormalProgram(m *module.Module) *temporal.NormalProgram {
	if m != nil {
		return temporal.NormalProgramFromModule(m)
	}
	return &temporal.NormalProgram{
		Init:  actions.NewSequence(),
		Calls: nil,
	}
}

func sortedSymbols(sig *il.Sig) []*lg.Const {
	if sig == nil {
		return nil
	}
	var result []*lg.Const
	for name, entry := range sig.Symbols {
		result = append(result, lg.NewConst(name, entry.Sort))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func collectAllNamedBinders(model *temporal.NormalProgram) map[string][]*lg.NamedBinder {
	result := make(map[string][]*lg.NamedBinder)

	collect := func(n lg.Expr) {
		for _, b := range lu.NamedBindersAst(n) {
			result[b.Name] = append(result[b.Name], b)
		}
	}

	for _, inv := range model.Invars {
		collect(inv.Formula.(lg.Expr))
	}
	for _, asm := range model.Asms {
		collect(asm.Formula.(lg.Expr))
	}
	collectActionNBs(model.Init, &result)
	for _, b := range model.Bindings {
		collectActionNBs(b.Action.Stmt, &result)
	}

	// Deduplicate and sort within each name
	for k, v := range result {
		seen := make(map[string]bool)
		var deduped []*lg.NamedBinder
		for _, b := range v {
			key := b.String()
			if !seen[key] {
				seen[key] = true
				deduped = append(deduped, b)
			}
		}
		sort.Slice(deduped, func(i, j int) bool {
			return deduped[i].String() < deduped[j].String()
		})
		result[k] = deduped
	}
	return result
}

func collectActionNBs(act actions.Action, result *map[string][]*lg.NamedBinder) {
	if act == nil {
		return
	}
	for _, a := range act.ActionArgs() {
		if sub, ok := a.(actions.Action); ok {
			collectActionNBs(sub, result)
		} else if a != nil {
			for _, b := range lu.NamedBindersAst(a) {
				(*result)[b.Name] = append((*result)[b.Name], b)
			}
		}
	}
}

func applyL2sInit(vs []*lg.Variable, t lg.Expr, label string) lg.Expr {
	if not, ok := t.(*lg.Not); ok {
		return &lg.Not{Body: applyL2sInit(vs, not.Body, label)}
	}
	return applyNB(l2sInit(vs, t, label), varsToNodes(vs)...)
}

// --- Public utility functions ---

// CreateSavedCopy creates a "saved" version of a symbol name.
func CreateSavedCopy(name string) string {
	return "l2s_saved_" + name
}

// IsSavedSymbol returns true if a symbol name is a saved copy.
func IsSavedSymbol(name string) bool {
	return strings.HasPrefix(name, "l2s_saved_")
}

// IsL2SSymbol returns true if a symbol name is an L2S auxiliary symbol.
func IsL2SSymbol(name string) bool {
	return strings.HasPrefix(name, "l2s_")
}

// TemporalAndL2S checks if a symbol is temporal-related or L2S-related.
func TemporalAndL2S(sym *lg.Const) bool {
	return (strings.HasPrefix(sym.Name, "l2s") && !strings.HasPrefix(sym.Name, "l2s_g")) ||
		strings.HasPrefix(sym.Name, "_old_l2s")
}

// L2SGToGlobally converts l2s_g named binders back to Globally operators.
func L2SGToGlobally(n lg.Expr) lg.Expr {
	if nb, ok := n.(*lg.NamedBinder); ok && nb.Name == "l2s_g" {
		return &lg.Globally{Environ: nb.Environ, Body: L2SGToGlobally(nb.Body)}
	}
	children := n.Children()
	if len(children) == 0 {
		return n
	}
	newChildren := make([]lg.Expr, len(children))
	changed := false
	for i, c := range children {
		nc := L2SGToGlobally(c)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return n
	}
	return il.CloneNode(n, newChildren)
}

// Desugar replaces $was and $happened named binders with L2S equivalents.
func Desugar(expr lg.Expr, proofLabel string) lg.Expr {
	l2sSaved := L2SSaved()

	if nb, ok := expr.(*lg.NamedBinder); ok {
		if nb.Name == "was" {
			return &lg.And{Terms: []lg.Expr{l2sSaved, applyWasRec(nb.Body, proofLabel)}}
		} else if nb.Name == "happened" {
			vs := lu.FreeVariablesList(nb.Body)
			return &lg.And{Terms: []lg.Expr{
				l2sSaved,
				&lg.Not{Body: applyNB(l2sW(vs, nb.Body, proofLabel), varsToNodes(vs)...)},
			}}
		}
	}

	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]lg.Expr, len(children))
	changed := false
	for i, c := range children {
		nc := Desugar(c, proofLabel)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return expr
	}
	return il.CloneNode(expr, newChildren)
}

func applyWasRec(expr lg.Expr, proofLabel string) lg.Expr {
	switch t := expr.(type) {
	case *lg.And:
		terms := make([]lg.Expr, len(t.Terms))
		for i, a := range t.Terms {
			terms[i] = applyWasRec(a, proofLabel)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Expr, len(t.Terms))
		for i, a := range t.Terms {
			terms[i] = applyWasRec(a, proofLabel)
		}
		return &lg.Or{Terms: terms}
	case *lg.Not:
		return &lg.Not{Body: applyWasRec(t.Body, proofLabel)}
	case *lg.Implies:
		return &lg.Implies{T1: applyWasRec(t.T1, proofLabel), T2: applyWasRec(t.T2, proofLabel)}
	case *lg.Iff:
		return &lg.Iff{T1: applyWasRec(t.T1, proofLabel), T2: applyWasRec(t.T2, proofLabel)}
	}
	vs := lu.FreeVariablesList(expr)
	return applyNB(l2sS(vs, expr, proofLabel), varsToNodes(vs)...)
}

// --- Registration ---

// RegisterTactics registers the l2s tactics on the given proof config.
// Replaces the old init()-based global registration.
func RegisterTactics(proofCfg *module.ProofConfig) {
	proofCfg.RegisterTactic("l2s", L2STactic)
	proofCfg.RegisterTactic("l2s_full", L2STacticFull)
	proofCfg.RegisterTactic("l2s_auto", L2STacticAuto)
	proofCfg.RegisterTactic("l2s_auto2", L2STacticAuto)
	proofCfg.RegisterTactic("l2s_auto3", L2STacticAuto)
	proofCfg.RegisterTactic("l2s_auto4", L2STacticAuto)
	proofCfg.RegisterTactic("l2s_auto5", L2STacticAuto)
}
