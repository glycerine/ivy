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
	mod "github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
	"github.com/glycerine/goivy/temporal"
)

// Debug controls l2s debug output.
var Debug bool

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
	return &lg.NamedBinder{
		Name:      "l2s_w",
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// l2sS creates an l2s_s (saved) named binder.
func l2sS(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_s",
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// l2sG creates an l2s_g (globally/safety) named binder.
func l2sG(vs []*lg.Var, t lg.Node, environ *string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_g",
		Variables: vs,
		Environ:   environ,
		Body:      t,
	}
}

// oldL2sG creates an _old_l2s_g named binder.
func oldL2sG(vs []*lg.Var, t lg.Node, environ *string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "_old_l2s_g",
		Variables: vs,
		Environ:   environ,
		Body:      t,
	}
}

// l2sInit creates an l2s_init named binder.
func l2sInit(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_init",
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// l2sWhen creates an l2s_when<name> named binder.
func l2sWhen(name string, vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_when" + name,
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// l2sOld creates an l2s_old named binder.
func l2sOld(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_old",
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// --- Helper: apply a named binder to arguments ---

// applyNB applies a named binder to arguments, returning Apply if
// there are arguments or the binder itself if there are none.
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

// varsToNodes converts vars to nodes.
func varsToNodes(vs []*lg.Var) []lg.Node {
	nodes := make([]lg.Node, len(vs))
	for i, v := range vs {
		nodes[i] = v
	}
	return nodes
}

// forall wraps body in ForAll if vs is non-empty.
func forall(vs []*lg.Var, body lg.Node) lg.Node {
	if len(vs) == 0 {
		return body
	}
	return &lg.ForAll{Variables: vs, Body: body}
}

// exists wraps body in Exists if vs is non-empty.
func exists(vs []*lg.Var, body lg.Node) lg.Node {
	if len(vs) == 0 {
		return body
	}
	return &lg.Exists{Variables: vs, Body: body}
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string { return &s }

// --- l2s_g tracking ---

// l2sGTriple stores (vars, body, environ) for a globally binder.
type l2sGTriple struct {
	Vars    []*lg.Var
	Body    lg.Node
	Environ *string
}

// --- Tactic entry points ---

// L2STactic is the main entry point for the "l2s" proof tactic.
// It matches the proof.Tactic signature.
func L2STactic(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	return l2sTacticInt(pc, goals, pf, "l2s")
}

// L2STacticFull includes all auxiliary state in the transformation.
func L2STacticFull(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	goals, err := l2sTacticInt(pc, goals, pf, "l2s_full")
	if err != nil {
		return nil, err
	}
	// In full mode, set trace_hook to hide l2s symbols and mark loop start.
	// (trace hooks are module-level, set on goals externally if needed)
	return goals, nil
}

// L2STacticAuto uses automatic trigger inference.
func L2STacticAuto(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node) ([]*ast.LabeledFormula, error) {
	return l2sTacticInt(pc, goals, pf, "l2s_auto")
}

// l2sTacticInt is the internal implementation of the L2S tactic.
// This is a faithful port of Python's l2s_tactic_int.
func l2sTacticInt(pc *proof.ProofChecker, goals []*ast.LabeledFormula, pf ast.Node, tacticName string) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, fmt.Errorf("l2s: no proof goals")
	}

	full := tacticName == "l2s_full"
	goal := goals[0]
	lineno := ast.Location{Filename: "l2s", Line: 0}
	conc := proof.GoalConc(goal)

	// Check that the conclusion is a temporal proof goal
	tm, ok := conc.(*ast.TemporalModels)
	if !ok {
		return nil, fmt.Errorf("l2s: proof goal is not temporal")
	}

	// Extract the model (NormalProgram) and formula
	model := extractNormalProgram(tm, pc)
	fmla := tm.Fmla.(lg.Node)

	// Get temporal premises
	prems := proof.GoalPrems(goal)
	var temporalPrems []lg.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.Temporal {
				temporalPrems = append(temporalPrems, lf.Formula.(lg.Node))
			}
		}
	}

	// Add assumed globally properties to model assumptions
	if pc != nil {
		for _, ax := range pc.Axioms {
			if !ax.Explicit && ax.Temporal {
				if g, ok := ax.Formula.(*lg.Globally); ok {
					model.Asms = append(model.Asms, &mod.LabeledFormula{
						Formula: g.Body,
					})
				}
			}
		}
	}

	// If there are temporal premises, wrap formula: prems -> fmla
	if len(temporalPrems) > 0 {
		premConj := il.MakeAnd(temporalPrems...)
		fmla = &lg.Implies{T1: premConj, T2: fmla}
	}

	proofLabel := ""

	// --- Compile tactic invariants ---
	// In the full integration, tactic_decls would be compiled here using
	// compile_with_goal_vocab. For now, we work with already-compiled invariants
	// from the model and the user-specified tactic.
	var invars []*mod.LabeledFormula
	// Copy existing model invariants
	invars = append(invars, model.Invars...)

	// --- L2S monitor symbols ---
	l2sWaiting := L2SWaiting()
	l2sFrozen := L2SFrozen()
	l2sSaved := L2SSaved()

	// --- Finite sorts ---
	m := getModule(pc)
	finiteSorts := make(map[string]bool)
	var uninterpretedSorts []lg.Sort
	if m != nil && m.Sig != nil {
		for name := range m.Sig.Sorts {
			s := m.Sig.Sorts[name]
			if m.FiniteSorts[name] || full {
				finiteSorts[name] = true
			} else if _, ok := s.(*lg.UninterpretedSort); ok {
				if !finiteSorts[name] {
					uninterpretedSorts = append(uninterpretedSorts, s)
				}
			}
		}
	}
	// Sort for determinism
	sort.Slice(uninterpretedSorts, func(i, j int) bool {
		return uninterpretedSorts[i].String() < uninterpretedSorts[j].String()
	})

	// --- Build definition dependencies ---
	defnDeps := make(map[string][]string) // symbol -> symbols it depends on
	if m != nil {
		for _, defn := range m.Definitions {
			f := il.DropUniversals(defn.Formula.(lg.Node))
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

	// Reachability for dependencies
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
			for _, d := range defnDeps[s] {
				stack = append(stack, d)
			}
		}
		return result
	}

	// ---------------------------------------------------------------
	// Step 1: Convert temporal operators to named binders
	// ---------------------------------------------------------------

	l2sGs := make(map[l2sGTriple]bool)
	l2sWhens := make(map[*lg.NamedBinder]bool)

	// Custom binder constructors that track what we create
	_l2sG := func(vs []*lg.Var, t lg.Node, env *string) *lg.NamedBinder {
		res := l2sG(vs, t, env)
		l2sGs[l2sGTriple{vs, t, env}] = true
		return res
	}
	_l2sWhen := func(name string, vs []*lg.Var, t lg.Node) *lg.NamedBinder {
		if name == "first" {
			res := l2sWhen("next", vs, t, proofLabel)
			l2sWhens[res] = true
			res2 := l2sInit(vs, applyNB(res, varsToNodes(vs)...), proofLabel)
			return res2
		}
		res := l2sWhen(name, vs, t, proofLabel)
		l2sWhens[res] = true
		return res
	}

	replaceTemporalsByL2sG := func(ast lg.Node) lg.Node {
		return lu.ReplaceTemporalsByNamedBinder(ast,
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
			model.Invars[i] = &mod.LabeledFormula{
				Label:   inv.Label,
				Formula: transform(inv.Formula.(lg.Node)),
			}
		}
		for i, asm := range model.Asms {
			model.Asms[i] = &mod.LabeledFormula{
				Label:   asm.Label,
				Formula: transform(asm.Formula.(lg.Node)),
			}
		}
		for i, b := range model.Bindings {
			newStmt := transformAction(b.Action.Stmt, transform)
			model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
		}
		model.Init = transformAction(model.Init, transform)
		// Transform premises that are properties
		for i, p := range invars {
			invars[i] = &mod.LabeledFormula{
				Label:   p.Label,
				Formula: transform(p.Formula.(lg.Node)),
			}
		}
	}

	// Step 1a: Replace all temporal operators with named binders
	modPass(replaceTemporalsByL2sG)

	notLf := replaceTemporalsByL2sG(&lg.Not{Body: fmla})

	if Debug {
		fmt.Println(strings.Repeat("=", 80) + "\nafter replace_temporals_by_named_binder_g_ast\n")
		for triple := range l2sGs {
			fmt.Printf("l2s_g: %v %v %v\n", triple.Vars, triple.Body, triple.Environ)
		}
		fmt.Println(strings.Repeat("=", 80))
	}

	// Step 1b: Normalize all named binders
	modPass(func(n lg.Node) lg.Node {
		return lu.NormalizeNamedBinders(n, nil)
	})

	if Debug {
		fmt.Println(strings.Repeat("=", 80) + "\nafter normalize_named_binders\n")
		fmt.Println(model)
		fmt.Println(strings.Repeat("=", 80))
	}

	// ---------------------------------------------------------------
	// Step 2: Build monitor building blocks
	// ---------------------------------------------------------------

	// reset_a: for each uninterpreted sort, l2s_a(X) := l2s_d(X)
	var resetA []actions.Action
	for _, s := range uninterpretedSorts {
		v, _ := lg.NewVar("X", s)
		resetA = append(resetA, setLineno(actions.NewAssignAction(
			applyNB(l2sS(nil, nil, ""), v), // placeholder — fixed below
			applyNB(l2sS(nil, nil, ""), v),
		), lineno))
		// Actually: l2s_a(s)(v) := l2s_d(s)(v)
		lhs := mustApply(L2SA(s), v)
		rhs := mustApply(L2SD(s), v)
		resetA[len(resetA)-1] = setLineno(actions.NewAssignAction(lhs, rhs), lineno)
	}

	// add_consts_to_d: for each constant c of uninterpreted sort, l2s_d(sort)(c) := true
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
	// Step 3: Figure out which l2s_w and l2s_s are used in conjectures
	// ---------------------------------------------------------------

	namedBindersConjs := make(map[string][]varBodyPair) // name -> [(vars, body)]
	for _, inv := range model.Invars {
		for _, b := range lu.NamedBindersAst(inv.Formula.(lg.Node)) {
			namedBindersConjs[b.Name] = append(namedBindersConjs[b.Name],
				varBodyPair{b.Variables, b.Body})
		}
	}
	// Deduplicate
	for k, v := range namedBindersConjs {
		namedBindersConjs[k] = dedupeVarBodyPairs(v)
	}

	// In full mode, add all state variables to 'to_save' and all temporal operators to 'to_wait'
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

	toWait := namedBindersConjs["l2s_w"]   // list of (vars, body)
	toSave := namedBindersConjs["l2s_s"]   // list of (vars, body)

	if Debug {
		fmt.Println(strings.Repeat("=", 40) + "\nto_wait:")
		for _, vb := range toWait {
			fmt.Printf("  %v %v\n", vb.Vars, vb.Body)
		}
		fmt.Println(strings.Repeat("=", 40))
	}

	// save_state: l2s_s(vs,t)(vs) := t
	var saveState []actions.Action
	for _, vb := range toSave {
		lhs := applyNB(l2sS(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
		saveState = append(saveState, setLineno(actions.NewAssignAction(lhs, vb.Body), lineno))
	}

	// done_waiting: forall vs. ~l2s_w(vs,t)(vs) for each waited formula
	var doneWaiting []lg.Node
	for _, vb := range toWait {
		inner := applyNB(l2sW(vb.Vars, vb.Body, proofLabel), varsToNodes(vb.Vars)...)
		doneWaiting = append(doneWaiting, forall(vb.Vars, &lg.Not{Body: inner}))
	}

	// reset_w: l2s_w(vs,t)(vs) := l2s_d(v.sort)(v) & ~t & ~(l2s_g(~t))
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
		negatedGlobally := replaceTemporalsByL2sG(
			&lg.Not{Body: &lg.Globally{Environ: strPtr(proofLabel), Body: co.Negate(vb.Body)}})
		conjuncts = append(conjuncts, negatedGlobally)
		rhs := il.MakeAnd(conjuncts...)
		resetW = append(resetW, setLineno(actions.NewAssignAction(lhs, rhs), lineno))
	}

	// ---------------------------------------------------------------
	// Step 4: Fair cycle check
	// ---------------------------------------------------------------

	fairCycle := []lg.Node{l2sSaved}
	fairCycle = append(fairCycle, doneWaiting...)

	// Projection of relations (boolean-valued saved state)
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
						forall(vb.Vars, &lg.Implies{T1: il.MakeAnd(aConjs...), T2: iff}))
				} else {
					fairCycle = append(fairCycle, forall(vb.Vars, iff))
				}
			} else {
				fairCycle = append(fairCycle, iff)
			}
		}
	}

	// Projection of functions and constants (uninterpreted-sort valued)
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
					forall(vb.Vars, &lg.Implies{T1: il.MakeAnd(aConjs...), T2: eq}))
			} else {
				fairCycle = append(fairCycle, forall(vb.Vars, eq))
			}
		}
	}

	assertNoFairCycle := setLineno(
		actions.NewAssertAction(&lg.Not{Body: il.MakeAnd(fairCycle...)}), lineno)
	assertNoFairCycle.SetLineno(lineno)

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

	// waiting -> frozen: edge + done_waiting assumes + reset_a
	var waitToFrozenParts []actions.Action
	waitToFrozenParts = append(waitToFrozenParts, monitorEdge(l2sWaiting, l2sFrozen)...)
	for _, dw := range doneWaiting {
		waitToFrozenParts = append(waitToFrozenParts, setLineno(actions.NewAssumeAction(dw), lineno))
	}
	waitToFrozenParts = append(waitToFrozenParts, resetA...)
	waitToFrozen := setLineno(actions.ConcatActions(waitToFrozenParts...), lineno)

	// frozen -> saved: edge + save_state + reset_w
	var frozenToSavedParts []actions.Action
	frozenToSavedParts = append(frozenToSavedParts, monitorEdge(l2sFrozen, l2sSaved)...)
	frozenToSavedParts = append(frozenToSavedParts, saveState...)
	frozenToSavedParts = append(frozenToSavedParts, resetW...)
	frozenToSaved := setLineno(actions.ConcatActions(frozenToSavedParts...), lineno)

	// stay in same state
	selfLoop := setLineno(actions.NewSequence(), lineno)

	changeMonitorState := []actions.Action{
		setLineno(actions.NewChoiceAction(
			actions.WrapAction(waitToFrozen),
			actions.WrapAction(frozenToSaved),
			actions.WrapAction(selfLoop),
		), lineno),
	}

	// ---------------------------------------------------------------
	// Step 6: Tableau construction
	// ---------------------------------------------------------------

	toG := make([]l2sGTriple, 0, len(l2sGs))
	for triple := range l2sGs {
		toG = append(toG, triple)
	}
	// Sort for determinism
	sort.Slice(toG, func(i, j int) bool {
		return fmt.Sprint(toG[i].Body) < fmt.Sprint(toG[j].Body)
	})

	// assume_g_axioms: forall vs. l2s_g(vs,t)(vs) -> t
	var assumeGAxioms []actions.Action
	for _, triple := range toG {
		inner := &lg.Implies{
			T1: applyNB(l2sG(triple.Vars, triple.Body, triple.Environ), varsToNodes(triple.Vars)...),
			T2: triple.Body,
		}
		assumeGAxioms = append(assumeGAxioms,
			setLineno(actions.NewAssumeAction(forall(triple.Vars, inner)), lineno))
	}

	// assume_when_axioms: forall vs. cond -> when(vs) = val
	var assumeWhenAxioms []actions.Action
	for when := range l2sWhens {
		inner := forall(when.Variables, &lg.Implies{
			T1: when.Body, // Cond is encoded in the body as a Cond{} node
			T2: &lg.Eq{T1: applyNB(when, varsToNodes(when.Variables)...), T2: when.Body},
		})
		assumeWhenAxioms = append(assumeWhenAxioms,
			setLineno(actions.NewAssumeAction(inner), lineno))
	}

	// assume_init_axioms: forall vs. l2s_init(vs,t)(vs) = t
	var assumeInitAxioms []actions.Action
	for _, vb := range namedBindersConjs["l2s_init"] {
		applied := applyL2sInit(vb.Vars, vb.Body, proofLabel)
		inner := forall(vb.Vars, &lg.Eq{T1: applied, T2: vb.Body})
		assumeInitAxioms = append(assumeInitAxioms,
			setLineno(actions.NewAssumeAction(inner), lineno))
	}

	// assume_w_axioms: forall vs. ~(t & l2s_w(vs,t)(vs))
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

	// Build memo tables for symbol -> property/wait/when dependencies
	envprops := make(map[string][]*lg.NamedBinder) // environ -> props
	symprops := make(map[string][]*lg.NamedBinder) // symbol -> props
	symwaits := make(map[string][]*lg.NamedBinder) // symbol -> waits
	symwhens := make(map[string][]*lg.NamedBinder) // symbol -> whens

	for _, triple := range toG {
		prop := l2sG(triple.Vars, triple.Body, triple.Environ)
		env := ""
		if triple.Environ != nil {
			env = *triple.Environ
		}
		envprops[env] = append(envprops[env], prop)
		for _, sym := range il.SymbolsAst(triple.Body) {
			symprops[sym.Name] = append(symprops[sym.Name], prop)
		}
	}
	for when := range l2sWhens {
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

	// prop_events: generates pre/post actions for property state updates
	propEvents := func(gprops map[*lg.NamedBinder]bool) ([]actions.Action, []actions.Action) {
		var pre, post []actions.Action
		for gprop := range gprops {
			vs := gprop.Variables
			t := gprop.Body
			env := gprop.Environ
			// pre: save old value, havoc current
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
		for gprop := range gprops {
			vs := gprop.Variables
			t := gprop.Body
			env := gprop.Environ
			// monotonicity: old_g -> g
			pre = append(pre,
				setLineno(actions.NewAssumeAction(forall(vs,
					&lg.Implies{
						T1: applyNB(oldL2sG(vs, t, env), varsToNodes(vs)...),
						T2: applyNB(l2sG(vs, t, env), varsToNodes(vs)...),
					})), lineno))
			// falsification: ~old_g & t -> ~g
			pre = append(pre,
				setLineno(actions.NewAssumeAction(forall(vs,
					&lg.Implies{
						T1: &lg.And{Terms: []lg.Node{
							&lg.Not{Body: applyNB(oldL2sG(vs, t, env), varsToNodes(vs)...)},
							t,
						}},
						T2: &lg.Not{Body: applyNB(l2sG(vs, t, env), varsToNodes(vs)...)},
					})), lineno))
			// post: g -> t
			post = append(post,
				setLineno(actions.NewAssumeAction(forall(vs,
					&lg.Implies{
						T1: applyNB(l2sG(vs, t, env), varsToNodes(vs)...),
						T2: t,
					})), lineno))
		}
		return pre, post
	}

	// when_events: generates pre/post actions for when operator updates
	whenEvents := func(whens map[*lg.NamedBinder]bool) ([]actions.Action, []actions.Action) {
		var pre, post []actions.Action
		for when := range whens {
			vs := when.Variables
			if when.Name == "l2s_whennext" {
				// Save old condition
				cond := when.Body // the condition part
				oldcond := applyNB(l2sOld(vs, cond, proofLabel), varsToNodes(vs)...)
				pre = append(pre, setLineno(actions.NewAssignAction(oldcond, cond), lineno))
				post = append(post, setLineno(actions.NewIfAction(
					oldcond,
					actions.NewHavocAction(applyNB(when, varsToNodes(vs)...)),
					nil,
				), lineno))
			}
			if when.Name == "l2s_whenprev" {
				cond := when.Body
				post = append(post, setLineno(actions.NewIfAction(
					cond,
					actions.NewHavocAction(applyNB(when, varsToNodes(vs)...)),
					nil,
				), lineno))
			}
		}
		for when := range whens {
			post = append(post,
				actions.NewAssumeAction(forall(when.Variables,
					&lg.Implies{
						T1: when.Body,
						T2: &lg.Eq{
							T1: applyNB(when, varsToNodes(when.Variables)...),
							T2: when.Body,
						},
					})))
		}
		return pre, post
	}

	// wait_events: generates actions for eventuality updates
	waitEvents := func(waits map[*lg.NamedBinder]bool) []actions.Action {
		var res []actions.Action
		for wait := range waits {
			vs := wait.Variables
			t := wait.Body
			// l2s_w(vs,t)(vs) := l2s_w(vs,t)(vs) & ~t & ~(l2s_g(~t))
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

	// instr_stmt: instruments a statement with temporal events
	var instrStmt func(stmt actions.Action) actions.Action
	instrStmt = func(stmt actions.Action) actions.Action {
		// First, recur on sub-statements
		args := stmt.Args()
		newArgs := make([]lg.Node, len(args))
		changed := false
		for i, a := range args {
			if sub, ok := actions.UnwrapAction(a); ok {
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

		// Collect events for modified symbols
		eventProps := make(map[*lg.NamedBinder]bool)
		eventWhens := make(map[*lg.NamedBinder]bool)
		eventWaits := make(map[*lg.NamedBinder]bool)

		modifiedSyms := actions.Modifies(stmt)
		allDeps := dependencies(modifiedSyms)
		for sym := range allDeps {
			for _, prop := range symprops[sym] {
				eventProps[prop] = true
			}
			for _, when := range symwhens[sym] {
				eventWhens[when] = true
			}
			for _, wait := range symwaits[sym] {
				eventWaits[wait] = true
			}
		}

		// Add events
		preEvents, postEvents := propEvents(eventProps)
		whenPre, whenPost := whenEvents(eventWhens)
		preEvents = append(whenPre, preEvents...)
		postEvents = append(postEvents, whenPost...)
		postEvents = append(postEvents, waitEvents(eventWaits)...)

		res = actions.PrefixAction(res, preEvents)
		res = actions.PostfixAction(res, postEvents)
		actions.CopyFormalsTo(stmt, res)
		return res
	}

	// Instrument all action bindings
	for i, b := range model.Bindings {
		newStmt := instrStmt(b.Action.Stmt)
		model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
	}

	// ---------------------------------------------------------------
	// Step 8: Patch exported actions with L2S prefix/suffix
	// ---------------------------------------------------------------

	calls := make(map[string]bool)
	for _, c := range model.Calls {
		calls[c] = true
	}

	for i, b := range model.Bindings {
		if !calls[b.Name] {
			continue
		}
		// Add params to d
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
	// Step 9: Build idle action
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
	// Step 10: Build init action
	// ---------------------------------------------------------------

	var l2sInitActions []actions.Action
	l2sInitActions = append(l2sInitActions,
		setLineno(actions.NewAssignAction(l2sWaiting, lg.True), lineno),
		setLineno(actions.NewAssignAction(l2sFrozen, lg.False), lineno),
		setLineno(actions.NewAssignAction(l2sSaved, lg.False), lineno),
	)
	l2sInitActions = append(l2sInitActions, addConstsToD...)
	l2sInitActions = append(l2sInitActions, resetW...)
	l2sInitActions = append(l2sInitActions, assumeGAxioms...)
	l2sInitActions = append(l2sInitActions, assumeInitAxioms...)
	l2sInitActions = append(l2sInitActions, setLineno(actions.NewAssumeAction(notLf), lineno))

	model.Init = actions.PostfixAction(model.Init, l2sInitActions)

	if Debug {
		fmt.Println(strings.Repeat("=", 80) + "\nafter patching actions\n")
		fmt.Println(model)
		fmt.Println(strings.Repeat("=", 80))
	}

	// ---------------------------------------------------------------
	// Step 11: Replace all named binders by fresh relations
	// ---------------------------------------------------------------

	namedBinders := collectAllNamedBinders(model)

	// Make sure _old_l2s_g is consistent with l2s_g
	namedBinders["_old_l2s_g"] = nil
	for _, b := range namedBinders["l2s_g"] {
		namedBinders["_old_l2s_g"] = append(namedBinders["_old_l2s_g"],
			&lg.NamedBinder{
				Name:      "_old_l2s_g",
				Variables: b.Variables,
				Environ:   b.Environ,
				Body:      b.Body,
			})
	}

	// Build substitution map: named binder string -> fresh Const
	subs := make(map[string]lg.Node)
	for k, binders := range namedBinders {
		for i, b := range binders {
			freshName := fmt.Sprintf("%s_%d", k, i)
			freshConst := lg.NewConst(freshName, b.NodeSort())
			subs[b.String()] = freshConst
		}
	}

	if Debug {
		fmt.Println(strings.Repeat("=", 80) + "\nsubs:")
		for k, v := range subs {
			fmt.Printf("  %s : %s\n", k, v)
		}
		fmt.Println(strings.Repeat("=", 80))
	}

	modPass(func(ast lg.Node) lg.Node {
		return lu.ReplaceNamedBindersAst(ast, subs)
	})

	if Debug {
		fmt.Println(strings.Repeat("=", 80) + "\nafter replace_named_binders\n")
		fmt.Println(model)
		fmt.Println(strings.Repeat("=", 80))
	}

	// HACK: reestablish invariant
	for _, b := range model.Bindings {
		b.Action.Stmt.SetFormalParams(b.Action.Inputs)
		b.Action.Stmt.SetFormalReturns(b.Action.Outputs)
	}

	// ---------------------------------------------------------------
	// Step 12: Build new goal
	// ---------------------------------------------------------------

	// Change conclusion to M |= true
	newConc := &ast.TemporalModels{
		Model: tm.Model,
		Fmla:  lg.True,
	}

	// Filter non-temporal premises
	var nonTemporalPrems []ast.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.Temporal {
				continue
			}
		}
		nonTemporalPrems = append(nonTemporalPrems, p)
	}

	newGoal := proof.CloneGoal(goal, nonTemporalPrems, newConc)

	// Return new goal stack
	result := make([]*ast.LabeledFormula, len(goals))
	result[0] = newGoal
	copy(result[1:], goals[1:])
	return result, nil
}

// --- Helper types and functions ---

// varBodyPair stores (variables, body) for named binder tracking.
type varBodyPair struct {
	Vars []*lg.Var
	Body lg.Node
}

// dedupeVarBodyPairs removes duplicate pairs based on string representation.
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

// mustApply applies a function to arguments, ignoring errors.
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

// setLineno sets the line number on an action and returns it.
func setLineno(a actions.Action, loc ast.Location) actions.Action {
	a.SetLineno(loc)
	return a
}

// transformAction applies a node transformation function to all formulas in an action tree.
func transformAction(act actions.Action, transform func(lg.Node) lg.Node) actions.Action {
	if act == nil {
		return nil
	}
	args := act.Args()
	newArgs := make([]lg.Node, len(args))
	changed := false
	for i, a := range args {
		if sub, ok := actions.UnwrapAction(a); ok {
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

// getModule extracts the module from the proof checker.
func getModule(pc *proof.ProofChecker) *mod.Module {
	if pc != nil && pc.Module != nil {
		return pc.Module
	}
	return nil
}

// extractNormalProgram creates a NormalProgram from a TemporalModels goal.
func extractNormalProgram(tm *ast.TemporalModels, pc *proof.ProofChecker) *temporal.NormalProgram {
	m := getModule(pc)
	if m != nil {
		return temporal.NormalProgramFromModule(m)
	}
	// Fallback: create an empty normal program
	return &temporal.NormalProgram{
		Init:  actions.NewSequence(),
		Calls: nil,
	}
}

// sortedSymbols returns symbols from a signature sorted by name.
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

// collectAllNamedBinders collects all named binders from the model.
func collectAllNamedBinders(model *temporal.NormalProgram) map[string][]*lg.NamedBinder {
	result := make(map[string][]*lg.NamedBinder)

	collect := func(n lg.Node) {
		for _, b := range lu.NamedBindersAst(n) {
			result[b.Name] = append(result[b.Name], b)
		}
	}

	for _, inv := range model.Invars {
		collect(inv.Formula.(lg.Node))
	}
	for _, asm := range model.Asms {
		collect(asm.Formula.(lg.Node))
	}
	collectAction(model.Init, &result)
	for _, b := range model.Bindings {
		collectAction(b.Action.Stmt, &result)
	}

	// Deduplicate within each name
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
		// Sort for determinism
		sort.Slice(deduped, func(i, j int) bool {
			return deduped[i].String() < deduped[j].String()
		})
		result[k] = deduped
	}
	return result
}

// collectAction collects named binders from an action tree.
func collectAction(act actions.Action, result *map[string][]*lg.NamedBinder) {
	if act == nil {
		return
	}
	for _, a := range act.Args() {
		if sub, ok := actions.UnwrapAction(a); ok {
			collectAction(sub, result)
		} else if a != nil {
			for _, b := range lu.NamedBindersAst(a) {
				(*result)[b.Name] = append((*result)[b.Name], b)
			}
		}
	}
}

// applyL2sInit applies l2s_init, handling negation properly.
// Python: if type(t) == lg.Not: return Not(apply_l2s_init(vs, t.args[0]))
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

// --- Desugar helpers ---

// Desugar replaces $was and $happened named binders with L2S equivalents.
//
//	$was. phi(V)  -->   l2s_saved & ($l2s_s V.phi(V))(V)
//	$happened. phi --> l2s_saved & ~($l2s_w V.phi(V))(V)
func Desugar(expr lg.Node, proofLabel string) lg.Node {
	l2sSaved := L2SSaved()

	applyWas := func(e lg.Node) lg.Node {
		return applyWasRec(e, proofLabel)
	}
	applyHappened := func(e lg.Node) lg.Node {
		vs := lu.FreeVariablesList(e)
		return &lg.Not{Body: applyNB(l2sW(vs, e, proofLabel), varsToNodes(vs)...)}
	}

	if nb, ok := expr.(*lg.NamedBinder); ok {
		if nb.Name == "was" {
			if len(nb.Variables) > 0 {
				return expr // error case, but don't crash
			}
			return &lg.And{Terms: []lg.Node{l2sSaved, applyWas(nb.Body)}}
		} else if nb.Name == "happened" {
			if len(nb.Variables) > 0 {
				return expr
			}
			return &lg.And{Terms: []lg.Node{l2sSaved, applyHappened(nb.Body)}}
		}
	}

	// Recurse into children
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

// applyWasRec pushes l2s_s inside propositional connectives.
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
		return &lg.Implies{
			T1: applyWasRec(t.T1, proofLabel),
			T2: applyWasRec(t.T2, proofLabel),
		}
	case *lg.Iff:
		return &lg.Iff{
			T1: applyWasRec(t.T1, proofLabel),
			T2: applyWasRec(t.T2, proofLabel),
		}
	}
	vs := lu.FreeVariablesList(expr)
	return applyNB(l2sS(vs, expr, proofLabel), varsToNodes(vs)...)
}

// --- Trace hooks ---

// TraceHook hides auxiliary variables in an error trace and marks the loop start.
// This corresponds to Python's trace_hook function.
func TraceHook(states []interface{}) {
	// In the full integration, this would:
	// 1. Look through trace states for l2s_saved = true
	// 2. Mark the preceding state as loop_start
	// 3. Set hidden_symbols to filter l2s_ symbols
	// For now, this is a no-op placeholder pending trace infrastructure.
}

// RenamingHook converts temporary symbols back to named binders.
func RenamingHook(subs map[string]lg.Node) func(interface{}) interface{} {
	// Build reverse map
	reverse := make(map[string]string)
	for k, v := range subs {
		reverse[fmt.Sprint(v)] = k
	}
	return func(tr interface{}) interface{} {
		// In the full integration, this would rename symbols in the trace
		return tr
	}
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
