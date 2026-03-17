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

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	modpkg "github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
	"github.com/glycerine/goivy/temporal"
)

// Debug controls l2s debug output.
var Debug bool

// CurrentModule is set by the caller before invoking the tactic.
// This is the module being verified.
var CurrentModule *modpkg.Module

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
func l2sW(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_w", Variables: vs, Environ: strPtr(label), Body: t}
}

// l2sS creates an l2s_s (saved) named binder.
func l2sS(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_s", Variables: vs, Environ: strPtr(label), Body: t}
}

// l2sG creates an l2s_g (globally/safety) named binder.
func l2sG(vs []*lg.Var, t lg.Node, environ *string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_g", Variables: vs, Environ: environ, Body: t}
}

// oldL2sG creates an _old_l2s_g named binder.
func oldL2sG(vs []*lg.Var, t lg.Node, environ *string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "_old_l2s_g", Variables: vs, Environ: environ, Body: t}
}

// l2sInit creates an l2s_init named binder.
func l2sInit(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_init", Variables: vs, Environ: strPtr(label), Body: t}
}

// l2sWhen creates an l2s_when<name> named binder.
func l2sWhen(name string, vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_when" + name, Variables: vs, Environ: strPtr(label), Body: t}
}

// l2sOld creates an l2s_old named binder.
func l2sOld(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{Name: "l2s_old", Variables: vs, Environ: strPtr(label), Body: t}
}

// --- Helpers ---

func applyNB(nb *lg.NamedBinder, args ...lg.Node) lg.Node {
	if len(args) == 0 {
		return nb
	}
	result, err := lg.NewApply(nb, args...)
	if err != nil {
		return &lg.Apply{Func: nb, Terms: args}
	}
	return result
}

func mustApply(f lg.Node, args ...lg.Node) lg.Node {
	if len(args) == 0 {
		return f
	}
	result, err := lg.NewApply(f, args...)
	if err != nil {
		return &lg.Apply{Func: f, Terms: args}
	}
	return result
}

func varsToNodes(vs []*lg.Var) []lg.Node {
	nodes := make([]lg.Node, len(vs))
	for i, v := range vs {
		nodes[i] = v
	}
	return nodes
}

func forall(vs []*lg.Var, body lg.Node) lg.Node {
	if len(vs) == 0 {
		return body
	}
	return &lg.ForAll{Variables: vs, Body: body}
}

func exists(vs []*lg.Var, body lg.Node) lg.Node {
	if len(vs) == 0 {
		return body
	}
	return &lg.Exists{Variables: vs, Body: body}
}

func strPtr(s string) *string { return &s }

func makeAnd(terms ...lg.Node) lg.Node {
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
	Vars    []*lg.Var
	Body    lg.Node
	Environ *string
}

func (t l2sGTriple) key() string {
	return fmt.Sprintf("%v|%v|%v", t.Vars, t.Body, t.Environ)
}

// --- varBodyPair ---

type varBodyPair struct {
	Vars []*lg.Var
	Body lg.Node
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

// L2STactic is the main tactic for "l2s" proof goals.
func L2STactic(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	return l2sTacticInt(pc, goals, pf, "l2s")
}

// L2STacticFull includes all auxiliary state in the transformation.
func L2STacticFull(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	return l2sTacticInt(pc, goals, pf, "l2s_full")
}

// L2STacticAuto uses automatic trigger inference.
func L2STacticAuto(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	return l2sTacticInt(pc, goals, pf, "l2s_auto")
}

// l2sTacticInt is the internal implementation of the L2S tactic.
// Faithful port of Python's l2s_tactic_int.
func l2sTacticInt(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node, tacticName string) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, fmt.Errorf("l2s: no proof goals")
	}

	full := tacticName == "l2s_full"
	goal := goals[0]
	lineno := ast.Location{Filename: "l2s", Line: 0}
	// Check that the conclusion is a temporal proof goal.
	// TemporalModels is an ast.Node, not a lg.Node, so we must check
	// the goal's formula directly (GoalConc won't find it).
	tm := findTemporalModels(goal)
	if tm == nil {
		return nil, fmt.Errorf("l2s: proof goal is not temporal")
	}

	// Extract the model (NormalProgram) and formula
	m := CurrentModule
	model := extractNormalProgram(m)
	fmla, _ := tm.Fmla.(lg.Node)
	if fmla == nil {
		return nil, fmt.Errorf("l2s: could not extract temporal formula from goal")
	}

	// Get temporal premises
	prems := proof.GoalPrems(goal)
	var temporalPrems []lg.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.Temporal != nil { // Temporal is a Node, nil means non-temporal
				if f, ok := lf.Formula.(lg.Node); ok {
					temporalPrems = append(temporalPrems, f)
				}
			}
		}
	}

	// Add assumed globally properties to model assumptions
	if pc != nil {
		for _, ax := range pc.Axioms {
			if !ax.Explicit && ax.Temporal != nil {
				if f, ok := ax.Formula.(lg.Node); ok {
					if g, ok := f.(*lg.Globally); ok {
						model.Asms = append(model.Asms, &modpkg.LabeledFormula{
							Formula: g.Body,
						})
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

	// --- Invariants from model ---
	var invars []*modpkg.LabeledFormula
	invars = append(invars, model.Invars...)

	// --- L2S monitor symbols ---
	l2sWaitingSym := L2SWaiting()
	l2sFrozenSym := L2SFrozen()
	l2sSavedSym := L2SSaved()

	// --- Finite sorts ---
	finiteSorts := make(map[string]bool)
	var uninterpretedSorts []lg.Sort
	if m != nil && m.Sig != nil {
		for name, s := range m.Sig.Sorts {
			if m.FiniteSorts[name] || full {
				finiteSorts[name] = true
			} else if _, isUI := s.(*lg.UninterpretedSort); isUI {
				uninterpretedSorts = append(uninterpretedSorts, s)
			}
		}
	}
	sort.Slice(uninterpretedSorts, func(i, j int) bool {
		return uninterpretedSorts[i].String() < uninterpretedSorts[j].String()
	})

	// --- Build definition dependency map ---
	defnDeps := make(map[string][]string)
	if m != nil {
		for _, defn := range m.Definitions {
			f := il.DropUniversals(defn.Formula)
			if eq, ok := f.(*lg.Eq); ok {
				if app, ok := eq.T1.(*lg.Apply); ok {
					if c, ok := app.Func.(*lg.Const); ok {
						for _, sym := range il.SymbolsAst(eq.T2) {
							defnDeps[sym.Name] = append(defnDeps[sym.Name], c.Name)
						}
					}
				}
			}
		}
	}

	dependencies := func(syms map[string]bool) map[string]bool {
		result := make(map[string]bool)
		var stack []string
		for s := range syms {
			stack = append(stack, s)
		}
		for len(stack) > 0 {
			s := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if result[s] {
				continue
			}
			result[s] = true
			stack = append(stack, defnDeps[s]...)
		}
		return result
	}

	// ---------------------------------------------------------------
	// L2S Auto: generate task/trigger invariants (before main steps)
	// ---------------------------------------------------------------
	if strings.HasPrefix(tacticName, "l2s_auto") {
		var err error
		invars, err = l2sAutoInvariants(tacticName, goal, invars, proofLabel,
			fmla, finiteSorts, uninterpretedSorts, m)
		if err != nil {
			return nil, fmt.Errorf("l2s_auto: %w", err)
		}
	}

	// ---------------------------------------------------------------
	// Step 1: Convert temporal operators to named binders
	// ---------------------------------------------------------------

	l2sGs := make(map[string]l2sGTriple) // key -> triple
	l2sWhensSet := make(map[string]*lg.NamedBinder)

	_l2sG := func(vs []*lg.Var, t lg.Node, env *string) *lg.NamedBinder {
		res := l2sG(vs, t, env)
		triple := l2sGTriple{vs, t, env}
		l2sGs[triple.key()] = triple
		return res
	}
	_l2sWhen := func(name string, vs []*lg.Var, t lg.Node) *lg.NamedBinder {
		if name == "first" {
			res := l2sWhen("next", vs, t, proofLabel)
			l2sWhensSet[res.String()] = res
			return l2sInit(vs, applyNB(res, varsToNodes(vs)...), proofLabel)
		}
		res := l2sWhen(name, vs, t, proofLabel)
		l2sWhensSet[res.String()] = res
		return res
	}

	replaceTemporalsByL2sG := func(n lg.Node) lg.Node {
		return lu.ReplaceTemporalsByNamedBinder(n,
			func(vs []*lg.Var, body lg.Node, env *string) *lg.NamedBinder {
				return _l2sG(vs, body, env)
			},
			func(name string, vs []*lg.Var, body lg.Node) *lg.NamedBinder {
				return _l2sWhen(name, vs, body)
			},
		)
	}

	// --- Model pass helper ---
	modPass := func(transform func(lg.Node) lg.Node) {
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
			newStmt := transformAction(b.Action.Stmt, transform)
			model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
		}
		if model.Init != nil {
			model.Init = transformAction(model.Init, transform)
		}
		for i, inv := range invars {
			invars[i] = &modpkg.LabeledFormula{
				Label:   inv.Label,
				Formula: transform(inv.Formula),
			}
		}
	}

	modPass(replaceTemporalsByL2sG)
	notLf := replaceTemporalsByL2sG(&lg.Not{Body: fmla})

	if Debug {
		fmt.Println(strings.Repeat("=", 80) + "\nafter replace_temporals_by_named_binder_g_ast")
		for _, triple := range l2sGs {
			fmt.Printf("l2s_g: %v %v %v\n", triple.Vars, triple.Body, triple.Environ)
		}
		fmt.Println(strings.Repeat("=", 80))
	}

	// Normalize named binders
	modPass(func(n lg.Node) lg.Node {
		return lu.NormalizeNamedBinders(n, nil)
	})

	// ---------------------------------------------------------------
	// Step 2: Build monitor building blocks
	// ---------------------------------------------------------------

	// reset_a: l2s_a(s)(X) := l2s_d(s)(X) for each uninterpreted sort
	var resetA []actions.Action
	for _, s := range uninterpretedSorts {
		v, _ := lg.NewVar("X", s)
		resetA = append(resetA,
			setLineno(actions.NewAssignAction(mustApply(L2SA(s), v), mustApply(L2SD(s), v)), lineno))
	}

	// add_consts_to_d: l2s_d(s)(c) := true for each constant c of sort s
	var addConstsToD []actions.Action
	if m != nil && m.Sig != nil {
		for _, s := range uninterpretedSorts {
			for _, sym := range sortedSymbols(m.Sig) {
				if sym.CSort != nil && sym.CSort.String() == s.String() {
					addConstsToD = append(addConstsToD,
						setLineno(actions.NewAssignAction(mustApply(L2SD(s), sym), lg.True), lineno))
				}
			}
		}
	}

	// ---------------------------------------------------------------
	// Step 3: Collect used l2s_w and l2s_s from conjectures
	// ---------------------------------------------------------------

	namedBindersConjs := make(map[string][]varBodyPair)
	for _, inv := range model.Invars {
		for _, b := range lu.NamedBindersAst(inv.Formula) {
			namedBindersConjs[b.Name] = append(namedBindersConjs[b.Name],
				varBodyPair{b.Variables, b.Body})
		}
	}
	for k, v := range namedBindersConjs {
		namedBindersConjs[k] = dedupeVarBodyPairs(v)
	}

	// In full mode, add all state variables to 'to_save'
	if full {
		seenSave := make(map[string]bool)
		for _, vb := range namedBindersConjs["l2s_s"] {
			seenSave[fmt.Sprint(vb.Body)] = true
		}
		for _, bnd := range model.Bindings {
			for _, act := range bnd.Action.Stmt.IterSubactions() {
				mods := actions.Modifies(act)
				for symName := range mods {
					if m != nil && m.Sig != nil {
						if entry, ok := m.Sig.Symbols[symName]; ok {
							vs := co.SymPlaceholders(lg.NewConst(symName, entry.Sort))
							var expr lg.Node
							if len(vs) > 0 {
								expr = mustApply(lg.NewConst(symName, entry.Sort), varsToNodes(vs)...)
							} else {
								expr = lg.NewConst(symName, entry.Sort)
							}
							key := fmt.Sprint(expr)
							if !seenSave[key] {
								seenSave[key] = true
								namedBindersConjs["l2s_s"] = append(namedBindersConjs["l2s_s"],
									varBodyPair{vs, expr})
							}
						}
					}
				}
			}
		}

		seenWait := make(map[string]bool)
		for _, vb := range namedBindersConjs["l2s_w"] {
			seenWait[fmt.Sprint(vb.Body)] = true
		}
		normNotLf := lu.NormalizeNamedBinders(notLf, nil)
		for _, b := range lu.NamedBindersAst(normNotLf) {
			if b.Name == "l2s_g" {
				negBody := co.Negate(b.Body)
				key := fmt.Sprint(negBody)
				if !seenWait[key] {
					seenWait[key] = true
					namedBindersConjs["l2s_w"] = append(namedBindersConjs["l2s_w"],
						varBodyPair{b.Variables, negBody})
				}
			}
			if b.Name == "l2s_init" {
				namedBindersConjs["l2s_init"] = append(namedBindersConjs["l2s_init"],
					varBodyPair{b.Variables, b.Body})
			}
		}
		namedBindersConjs["l2s_init"] = dedupeVarBodyPairs(namedBindersConjs["l2s_init"])
	}

	toWait := namedBindersConjs["l2s_w"]
	toSave := namedBindersConjs["l2s_s"]

	// save_state actions
	var saveState []actions.Action
	for _, vb := range toSave {
		lhs := applyNB(l2sS(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
		saveState = append(saveState, setLineno(actions.NewAssignAction(lhs, vb.Body), lineno))
	}

	// done_waiting formulas
	var doneWaiting []lg.Node
	for _, vb := range toWait {
		inner := applyNB(l2sW(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
		doneWaiting = append(doneWaiting, forall(vb.Vars, &lg.Not{Body: inner}))
	}

	// reset_w actions
	var resetW []actions.Action
	for _, vb := range toWait {
		lhs := applyNB(l2sW(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
		var conjuncts []lg.Node
		for _, v := range vb.Vars {
			if !finiteSorts[v.VSort.String()] {
				conjuncts = append(conjuncts, mustApply(L2SD(v.VSort), v))
			}
		}
		conjuncts = append(conjuncts, &lg.Not{Body: vb.Body})
		negGlob := replaceTemporalsByL2sG(
			&lg.Not{Body: &lg.Globally{Environ: strPtr(proofLabel), Body: co.Negate(vb.Body)}})
		conjuncts = append(conjuncts, negGlob)
		resetW = append(resetW, setLineno(actions.NewAssignAction(lhs, makeAnd(conjuncts...)), lineno))
	}

	// ---------------------------------------------------------------
	// Step 4: Fair cycle check
	// ---------------------------------------------------------------

	fairCycle := []lg.Node{l2sSavedSym}
	fairCycle = append(fairCycle, doneWaiting...)

	// Projection of relations
	for _, vb := range toSave {
		bodySort := vb.Body.NodeSort()
		isRelation := bodySort == lg.Boolean
		if fs, ok := bodySort.(*lg.FunctionSort); ok && fs.Range() == lg.Boolean {
			isRelation = true
		}
		if isRelation {
			savedApp := applyNB(l2sS(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
			iff := &lg.Iff{T1: savedApp, T2: vb.Body}
			if len(vb.Vars) > 0 {
				var aConjs []lg.Node
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
	for _, vb := range toSave {
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
			var aConjs []lg.Node
			for _, v := range vb.Vars {
				if !finiteSorts[v.VSort.String()] {
					aConjs = append(aConjs, mustApply(L2SA(v.VSort), v))
				}
			}
			if !finiteSorts[bodySort.String()] {
				aConjs = append(aConjs,
					&lg.Or{Terms: []lg.Node{
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

	assertNoFairCycle := setLineno(
		actions.NewAssertAction(&lg.Not{Body: makeAnd(fairCycle...)}), lineno)

	// ---------------------------------------------------------------
	// Step 5: Monitor state machine
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
	for _, dw := range doneWaiting {
		waitToFrozenParts = append(waitToFrozenParts, setLineno(actions.NewAssumeAction(dw), lineno))
	}
	waitToFrozenParts = append(waitToFrozenParts, resetA...)

	// frozen -> saved
	var frozenToSavedParts []actions.Action
	frozenToSavedParts = append(frozenToSavedParts, monitorEdge(l2sFrozenSym, l2sSavedSym)...)
	frozenToSavedParts = append(frozenToSavedParts, saveState...)
	frozenToSavedParts = append(frozenToSavedParts, resetW...)

	changeMonitorState := []actions.Action{
		setLineno(actions.NewChoiceAction(
			actions.WrapAction(setLineno(actions.ConcatActions(waitToFrozenParts...), lineno)),
			actions.WrapAction(setLineno(actions.ConcatActions(frozenToSavedParts...), lineno)),
			actions.WrapAction(setLineno(actions.NewSequence(), lineno)),
		), lineno),
	}

	// ---------------------------------------------------------------
	// Step 6: Tableau construction
	// ---------------------------------------------------------------

	toG := make([]l2sGTriple, 0, len(l2sGs))
	for _, triple := range l2sGs {
		toG = append(toG, triple)
	}
	sort.Slice(toG, func(i, j int) bool {
		return fmt.Sprint(toG[i].Body) < fmt.Sprint(toG[j].Body)
	})

	// assume_g_axioms
	var assumeGAxioms []actions.Action
	for _, triple := range toG {
		inner := &lg.Implies{
			T1: applyNB(l2sG(triple.Vars, triple.Body, triple.Environ), varsToNodes(triple.Vars)...),
			T2: triple.Body,
		}
		assumeGAxioms = append(assumeGAxioms,
			setLineno(actions.NewAssumeAction(forall(triple.Vars, inner)), lineno))
	}

	// assume_when_axioms
	var assumeWhenAxioms []actions.Action
	for _, when := range l2sWhensSet {
		inner := forall(when.Variables, &lg.Implies{
			T1: when.Body,
			T2: &lg.Eq{T1: applyNB(when, varsToNodes(when.Variables)...), T2: when.Body},
		})
		assumeWhenAxioms = append(assumeWhenAxioms,
			setLineno(actions.NewAssumeAction(inner), lineno))
	}

	// assume_init_axioms
	var assumeInitAxioms []actions.Action
	for _, vb := range namedBindersConjs["l2s_init"] {
		applied := applyL2sInit(vb.Vars, vb.Body, proofLabel)
		inner := forall(vb.Vars, &lg.Eq{T1: applied, T2: vb.Body})
		assumeInitAxioms = append(assumeInitAxioms,
			setLineno(actions.NewAssumeAction(inner), lineno))
	}

	// assume_w_axioms
	var assumeWAxioms []actions.Action
	for _, vb := range namedBindersConjs["l2s_w"] {
		wApp := applyNB(l2sW(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
		inner := forall(vb.Vars, &lg.Not{Body: &lg.And{Terms: []lg.Node{vb.Body, wApp}}})
		assumeWAxioms = append(assumeWAxioms,
			setLineno(actions.NewAssumeAction(inner), lineno))
	}

	// ---------------------------------------------------------------
	// Step 7: Action instrumentation
	// ---------------------------------------------------------------

	symprops := make(map[string][]*lg.NamedBinder)
	symwaits := make(map[string][]*lg.NamedBinder)
	symwhens := make(map[string][]*lg.NamedBinder)

	for _, triple := range toG {
		prop := l2sG(triple.Vars, triple.Body, triple.Environ)
		for _, sym := range il.SymbolsAst(triple.Body) {
			symprops[sym.Name] = append(symprops[sym.Name], prop)
		}
	}
	for _, when := range l2sWhensSet {
		for _, sym := range il.SymbolsAst(when.Body) {
			symwhens[sym.Name] = append(symwhens[sym.Name], when)
		}
	}
	for _, vb := range toWait {
		wait := l2sW(vb.Vars, vb.Body, proofLabel)
		for _, sym := range il.SymbolsAst(vb.Body) {
			symwaits[sym.Name] = append(symwaits[sym.Name], wait)
		}
	}

	propEventsFunc := func(gprops map[string]*lg.NamedBinder) ([]actions.Action, []actions.Action) {
		var pre, post []actions.Action
		for _, gprop := range gprops {
			vs, t, env := gprop.Variables, gprop.Body, gprop.Environ
			pre = append(pre,
				setLineno(actions.NewAssignAction(
					applyNB(oldL2sG(vs, t, env), varsToNodes(vs)...),
					applyNB(l2sG(vs, t, env), varsToNodes(vs)...),
				), lineno))
			pre = append(pre,
				setLineno(actions.NewHavocAction(
					applyNB(l2sG(vs, t, env), varsToNodes(vs)...),
				), lineno))
		}
		for _, gprop := range gprops {
			vs, t, env := gprop.Variables, gprop.Body, gprop.Environ
			pre = append(pre,
				setLineno(actions.NewAssumeAction(forall(vs,
					&lg.Implies{
						T1: applyNB(oldL2sG(vs, t, env), varsToNodes(vs)...),
						T2: applyNB(l2sG(vs, t, env), varsToNodes(vs)...),
					})), lineno))
			pre = append(pre,
				setLineno(actions.NewAssumeAction(forall(vs,
					&lg.Implies{
						T1: &lg.And{Terms: []lg.Node{
							&lg.Not{Body: applyNB(oldL2sG(vs, t, env), varsToNodes(vs)...)},
							t,
						}},
						T2: &lg.Not{Body: applyNB(l2sG(vs, t, env), varsToNodes(vs)...)},
					})), lineno))
			post = append(post,
				setLineno(actions.NewAssumeAction(forall(vs,
					&lg.Implies{
						T1: applyNB(l2sG(vs, t, env), varsToNodes(vs)...),
						T2: t,
					})), lineno))
		}
		return pre, post
	}

	whenEventsFunc := func(whens map[string]*lg.NamedBinder) ([]actions.Action, []actions.Action) {
		var pre, post []actions.Action
		for _, when := range whens {
			vs := when.Variables
			if when.Name == "l2s_whennext" {
				cond := when.Body
				oldcond := applyNB(l2sOld(vs, cond, proofLabel), varsToNodes(vs)...)
				pre = append(pre, setLineno(actions.NewAssignAction(oldcond, cond), lineno))
				post = append(post, setLineno(actions.NewIfAction(
					oldcond,
					actions.WrapAction(actions.NewHavocAction(applyNB(when, varsToNodes(vs)...))),
				), lineno))
			}
			if when.Name == "l2s_whenprev" {
				cond := when.Body
				post = append(post, setLineno(actions.NewIfAction(
					cond,
					actions.WrapAction(actions.NewHavocAction(applyNB(when, varsToNodes(vs)...))),
				), lineno))
			}
		}
		for _, when := range whens {
			post = append(post,
				actions.NewAssumeAction(forall(when.Variables,
					&lg.Implies{
						T1: when.Body,
						T2: &lg.Eq{T1: applyNB(when, varsToNodes(when.Variables)...), T2: when.Body},
					})))
		}
		return pre, post
	}

	waitEventsFunc := func(waits map[string]*lg.NamedBinder) []actions.Action {
		var res []actions.Action
		for _, wait := range waits {
			vs, t := wait.Variables, wait.Body
			waitApp := applyNB(wait, varsToNodes(vs)...)
			rhs := &lg.And{Terms: []lg.Node{
				waitApp,
				&lg.Not{Body: t},
				replaceTemporalsByL2sG(&lg.Not{Body: &lg.Globally{
					Environ: strPtr(proofLabel),
					Body:    co.Negate(t),
				}}),
			}}
			res = append(res, setLineno(actions.NewAssignAction(waitApp, rhs), lineno))
		}
		return res
	}

	var instrStmt func(stmt actions.Action) actions.Action
	instrStmt = func(stmt actions.Action) actions.Action {
		args := stmt.Args()
		newArgs := make([]lg.Node, len(args))
		changed := false
		for i, a := range args {
			if sub := actions.UnwrapAction(a); sub != nil {
				newSub := instrStmt(sub)
				newArgs[i] = actions.WrapAction(newSub)
				if newSub != sub {
					changed = true
				}
			} else {
				newArgs[i] = a
			}
		}
		var res actions.Action
		if changed {
			res = stmt.Clone(newArgs)
		} else {
			res = stmt
		}

		eventProps := make(map[string]*lg.NamedBinder)
		eventWhens := make(map[string]*lg.NamedBinder)
		eventWaits := make(map[string]*lg.NamedBinder)

		modifiedSyms := actions.Modifies(stmt)
		allDeps := dependencies(modifiedSyms)
		for sym := range allDeps {
			for _, prop := range symprops[sym] {
				eventProps[prop.String()] = prop
			}
			for _, when := range symwhens[sym] {
				eventWhens[when.String()] = when
			}
			for _, wait := range symwaits[sym] {
				eventWaits[wait.String()] = wait
			}
		}

		preEvents, postEvents := propEventsFunc(eventProps)
		whenPre, whenPost := whenEventsFunc(eventWhens)
		preEvents = append(whenPre, preEvents...)
		postEvents = append(postEvents, whenPost...)
		postEvents = append(postEvents, waitEventsFunc(eventWaits)...)

		res = actions.PrefixAction(res, preEvents)
		res = actions.PostfixAction(res, postEvents)
		actions.CopyFormalsTo(stmt, res)
		return res
	}

	// Instrument all bindings
	for i, b := range model.Bindings {
		newStmt := instrStmt(b.Action.Stmt)
		model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
	}

	// ---------------------------------------------------------------
	// Step 8: Patch exported actions
	// ---------------------------------------------------------------

	calls := make(map[string]bool)
	for _, c := range model.Calls {
		calls[c] = true
	}

	for i, b := range model.Bindings {
		if !calls[b.Name] {
			continue
		}
		var addParamsToD []actions.Action
		for _, p := range b.Action.Inputs {
			if p.CSort != nil && !finiteSorts[p.CSort.String()] {
				if _, ok := p.CSort.(*lg.UninterpretedSort); ok {
					addParamsToD = append(addParamsToD,
						setLineno(actions.NewAssignAction(mustApply(L2SD(p.CSort), p), lg.True), lineno))
				}
			}
		}

		var stmtParts []actions.Action
		stmtParts = append(stmtParts, addParamsToD...)
		stmtParts = append(stmtParts, assumeGAxioms...)
		stmtParts = append(stmtParts, assumeWhenAxioms...)
		stmtParts = append(stmtParts, assumeWAxioms...)
		stmtParts = append(stmtParts, b.Action.Stmt)
		stmtParts = append(stmtParts, addConstsToD...)

		newStmt := setLineno(actions.ConcatActions(stmtParts...), lineno)
		actions.CopyFormalsTo(b.Action.Stmt, newStmt)
		model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
	}

	// ---------------------------------------------------------------
	// Step 9: Idle action
	// ---------------------------------------------------------------

	var idleParts []actions.Action
	idleParts = append(idleParts, changeMonitorState...)
	idleParts = append(idleParts, assumeGAxioms...)
	idleParts = append(idleParts, addConstsToD...)
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
	// Step 10: Init action
	// ---------------------------------------------------------------

	var l2sInitActions []actions.Action
	l2sInitActions = append(l2sInitActions,
		setLineno(actions.NewAssignAction(l2sWaitingSym, lg.True), lineno),
		setLineno(actions.NewAssignAction(l2sFrozenSym, lg.False), lineno),
		setLineno(actions.NewAssignAction(l2sSavedSym, lg.False), lineno),
	)
	l2sInitActions = append(l2sInitActions, addConstsToD...)
	l2sInitActions = append(l2sInitActions, resetW...)
	l2sInitActions = append(l2sInitActions, assumeGAxioms...)
	l2sInitActions = append(l2sInitActions, assumeInitAxioms...)
	l2sInitActions = append(l2sInitActions, setLineno(actions.NewAssumeAction(notLf), lineno))

	if model.Init != nil {
		model.Init = actions.PostfixAction(model.Init, l2sInitActions)
	}

	// ---------------------------------------------------------------
	// Step 11: Replace named binders with fresh relations
	// ---------------------------------------------------------------

	namedBinders := collectAllNamedBinders(model)

	// Ensure _old_l2s_g is consistent with l2s_g
	namedBinders["_old_l2s_g"] = nil
	for _, b := range namedBinders["l2s_g"] {
		namedBinders["_old_l2s_g"] = append(namedBinders["_old_l2s_g"],
			&lg.NamedBinder{Name: "_old_l2s_g", Variables: b.Variables, Environ: b.Environ, Body: b.Body})
	}

	subs := make(map[string]lg.Node)
	for k, binders := range namedBinders {
		for i, b := range binders {
			freshName := fmt.Sprintf("%s_%d", k, i)
			subs[b.String()] = lg.NewConst(freshName, b.NodeSort())
		}
	}

	modPass(func(n lg.Node) lg.Node {
		return lu.ReplaceNamedBindersAst(n, subs)
	})

	// Reestablish formals invariant
	for _, b := range model.Bindings {
		b.Action.Stmt.SetFormalParams(b.Action.Inputs)
		b.Action.Stmt.SetFormalReturns(b.Action.Outputs)
	}

	// ---------------------------------------------------------------
	// Step 12: Build new goal
	// ---------------------------------------------------------------

	// Build M |= true as the new conclusion.
	// TemporalModels.Fmla is ast.Node, so wrap lg.True.
	newConc := &ast.TemporalModels{Model: tm.Model, Fmla: wrapLogicAsAST(lg.True)}

	var nonTemporalPrems []ast.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok && lf.Temporal != nil {
			continue
		}
		nonTemporalPrems = append(nonTemporalPrems, p)
	}

	// Build the new goal directly since CloneGoal expects lg.Node
	// but our conclusion is ast.Node (TemporalModels).
	newGoal := cloneGoalWithASTConc(goal, nonTemporalPrems, newConc)

	result := make([]*ast.LabeledFormula, len(goals))
	result[0] = newGoal
	copy(result[1:], goals[1:])
	return result, nil
}

// --- Adapter: wraps lg.Node as ast.Node ---

type logicASTAdapter struct {
	ast.Base
	Node lg.Node
}

func (a *logicASTAdapter) Args() []ast.Node        { return nil }
func (a *logicASTAdapter) Clone([]ast.Node) ast.Node { return a }
func (a *logicASTAdapter) String() string           { return a.Node.String() }

func wrapLogicAsAST(n lg.Node) ast.Node {
	if an, ok := n.(ast.Node); ok {
		return an
	}
	return &logicASTAdapter{Node: n}
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

// cloneGoalWithASTConc clones a goal with an ast.Node conclusion
// (instead of lg.Node which proof.CloneGoal requires).
func cloneGoalWithASTConc(goal *ast.LabeledFormula, prems []ast.Node, conc ast.Node) *ast.LabeledFormula {
	var formula ast.Node
	if len(prems) > 0 {
		elems := make([]ast.Node, len(prems)+1)
		copy(elems, prems)
		elems[len(prems)] = conc
		formula = ast.NewSchemaBody(elems...)
	} else {
		formula = conc
	}
	return goal.CloneWithFreshID([]ast.Node{goal.Label, formula})
}

func transformAction(act actions.Action, transform func(lg.Node) lg.Node) actions.Action {
	if act == nil {
		return nil
	}
	args := act.Args()
	newArgs := make([]lg.Node, len(args))
	changed := false
	for i, a := range args {
		if sub := actions.UnwrapAction(a); sub != nil {
			newSub := transformAction(sub, transform)
			newArgs[i] = actions.WrapAction(newSub)
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
		return act.Clone(newArgs)
	}
	return act
}

func extractNormalProgram(m *modpkg.Module) *temporal.NormalProgram {
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

	collect := func(n lg.Node) {
		for _, b := range lu.NamedBindersAst(n) {
			result[b.Name] = append(result[b.Name], b)
		}
	}

	for _, inv := range model.Invars {
		collect(inv.Formula)
	}
	for _, asm := range model.Asms {
		collect(asm.Formula)
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
	for _, a := range act.Args() {
		if sub := actions.UnwrapAction(a); sub != nil {
			collectActionNBs(sub, result)
		} else if a != nil {
			for _, b := range lu.NamedBindersAst(a) {
				(*result)[b.Name] = append((*result)[b.Name], b)
			}
		}
	}
}

func applyL2sInit(vs []*lg.Var, t lg.Node, label string) lg.Node {
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
func L2SGToGlobally(n lg.Node) lg.Node {
	if nb, ok := n.(*lg.NamedBinder); ok && nb.Name == "l2s_g" {
		return &lg.Globally{Environ: nb.Environ, Body: L2SGToGlobally(nb.Body)}
	}
	children := n.Children()
	if len(children) == 0 {
		return n
	}
	newChildren := make([]lg.Node, len(children))
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
func Desugar(expr lg.Node, proofLabel string) lg.Node {
	l2sSaved := L2SSaved()

	if nb, ok := expr.(*lg.NamedBinder); ok {
		if nb.Name == "was" {
			return &lg.And{Terms: []lg.Node{l2sSaved, applyWasRec(nb.Body, proofLabel)}}
		} else if nb.Name == "happened" {
			vs := lu.FreeVariablesList(nb.Body)
			return &lg.And{Terms: []lg.Node{
				l2sSaved,
				&lg.Not{Body: applyNB(l2sW(vs, nb.Body, proofLabel), varsToNodes(vs)...)},
			}}
		}
	}

	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]lg.Node, len(children))
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

func applyWasRec(expr lg.Node, proofLabel string) lg.Node {
	switch t := expr.(type) {
	case *lg.And:
		terms := make([]lg.Node, len(t.Terms))
		for i, a := range t.Terms {
			terms[i] = applyWasRec(a, proofLabel)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Node, len(t.Terms))
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

func init() {
	proof.RegisterTactic("l2s", L2STactic)
	proof.RegisterTactic("l2s_full", L2STacticFull)
	proof.RegisterTactic("l2s_auto", L2STacticAuto)
	proof.RegisterTactic("l2s_auto2", L2STacticAuto)
	proof.RegisterTactic("l2s_auto3", L2STacticAuto)
	proof.RegisterTactic("l2s_auto4", L2STacticAuto)
	proof.RegisterTactic("l2s_auto5", L2STacticAuto)
}
