// shared.go contains the L2S instrumentation pipeline steps that are
// shared between the l2s tactic and the ranking tactic.
//
// Each SharedStep* function corresponds to a numbered step in l2sTacticInt.
// The InstrumentationConfig struct carries all state between steps.
package l2s

import (
	"fmt"
	"sort"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	modpkg "github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/temporal"
)

// InstrumentationConfig holds all state for the shared L2S instrumentation pipeline.
type InstrumentationConfig struct {
	ProofLabel         string
	Lineno             ast.Location
	FiniteSorts        map[string]bool
	UninterpretedSorts []lg.Sort
	Mod                *modpkg.Module
	Fmla               lg.Expr // the temporal formula
	Invars             []*ast.LabeledFormula
	Postconds          []*ast.LabeledFormula // nil for l2s

	// Common building blocks (populated by BuildCommonBlocks or caller)
	AddConstsToD     []actions.Action
	ResetW           []actions.Action
	AssumeGAxioms    []actions.Action
	AssumeWhenAxioms []actions.Action
	AssumeWAxioms    []actions.Action
	AssumeInitAxioms []actions.Action

	// Collected state (populated by shared steps)
	L2sGs            map[string]L2sGTriple
	L2sWhensSet      map[string]*lg.NamedBinder
	NamedBindersConjs map[string][]VarBodyPair
	ToWait           []VarBodyPair
	ToSave           []VarBodyPair
	SaveState        []actions.Action
	DoneWaiting      []lg.Expr
	NotLf            lg.Expr

	// Functions built during step 1
	ReplaceTemporals func(lg.Expr) lg.Expr
	// Dependencies closure (built by caller from defnDeps)
	Dependencies func(map[string]bool) map[string]bool
}

// L2sGTriple holds the vars, body, and environ of an l2s_g binder.
type L2sGTriple = l2sGTriple

// VarBodyPair holds a pair of variables and a body expression.
type VarBodyPair = varBodyPair

// --- Exported helper functions used by both tactics ---

// TransformAction walks an action tree, applying transform to all lg.Expr leaves.
func TransformAction(act actions.Action, transform func(lg.Expr) lg.Expr) actions.Action {
	return transformAction(act, transform)
}

// CollectAllNamedBinders collects all named binders from a NormalProgram.
func CollectAllNamedBinders(model *temporal.NormalProgram) map[string][]*lg.NamedBinder {
	return collectAllNamedBinders(model)
}

// SortedSymbols returns the sorted constants from a signature.
func SortedSymbols(sig *il.Sig) []*lg.Symbol {
	return sortedSymbols(sig)
}

// ApplyNB applies a named binder to arguments.
func ApplyNB(nb *lg.NamedBinder, args ...lg.Expr) lg.Expr {
	return applyNB(nb, args...)
}

// VarsToNodes converts a slice of *lg.Variable to []lg.Expr.
func VarsToNodes(vs []*lg.Variable) []lg.Expr {
	return varsToNodes(vs)
}

// Forall wraps body in a ForAll if vs is non-empty.
func Forall(vs []*lg.Variable, body lg.Expr) lg.Expr {
	return forall(vs, body)
}

// MakeAnd creates an And node from terms, or lg.True if empty.
func MakeAnd(terms ...lg.Expr) lg.Expr {
	return makeAnd(terms...)
}

// SetLineno sets the location on an action and returns it.
func SetLineno(a actions.Action, loc ast.Location) actions.Action {
	return setLineno(a, loc)
}

// L2sW creates an l2s_w named binder (exported version).
func ExportL2sW(vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return l2sW(vs, t, label)
}

// L2sS creates an l2s_s named binder (exported version).
func ExportL2sS(vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return l2sS(vs, t, label)
}

// L2sG creates an l2s_g named binder (exported version).
func ExportL2sG(vs []*lg.Variable, t lg.Expr, environ *string) *lg.NamedBinder {
	return l2sG(vs, t, environ)
}

// OldL2sG creates an _old_l2s_g named binder (exported version).
func ExportOldL2sG(vs []*lg.Variable, t lg.Expr, environ *string) *lg.NamedBinder {
	return oldL2sG(vs, t, environ)
}

// L2sInit creates an l2s_init named binder (exported version).
func ExportL2sInit(vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return l2sInit(vs, t, label)
}

// L2sWhen creates an l2s_when named binder (exported version).
func ExportL2sWhen(name string, vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return l2sWhen(name, vs, t, label)
}

// L2sOld creates an l2s_old named binder (exported version).
func ExportL2sOld(vs []*lg.Variable, t lg.Expr, label string) *lg.NamedBinder {
	return l2sOld(vs, t, label)
}

// StrPtr returns a pointer to a string.
func StrPtr(s string) *string {
	return strPtr(s)
}

// ApplyL2sInit applies l2s_init, handling negation.
func ExportApplyL2sInit(vs []*lg.Variable, t lg.Expr, label string) lg.Expr {
	return applyL2sInit(vs, t, label)
}

// DedupeVarBodyPairs removes duplicate VarBodyPairs.
func DedupeVarBodyPairs(pairs []VarBodyPair) []VarBodyPair {
	return dedupeVarBodyPairs(pairs)
}

// SharedStep1_ConvertTemporals converts temporal operators to named binders.
// This is Step 1 of l2sTacticInt. It populates cfg.L2sGs, cfg.L2sWhensSet,
// cfg.ReplaceTemporals, and cfg.NotLf.
//
// modPass should apply a transform to the entire model (invars, asms, bindings, init, invars list, and postconds if applicable).
func SharedStep1_ConvertTemporals(cfg *InstrumentationConfig, model *temporal.NormalProgram, modPass func(func(lg.Expr) lg.Expr)) {
	cfg.L2sGs = make(map[string]L2sGTriple)
	cfg.L2sWhensSet = make(map[string]*lg.NamedBinder)

	_l2sG := func(vs []*lg.Variable, t lg.Expr, env *string) *lg.NamedBinder {
		res := l2sG(vs, t, env)
		triple := L2sGTriple{vs, t, env}
		cfg.L2sGs[triple.key()] = triple
		return res
	}
	_l2sWhen := func(name string, vs []*lg.Variable, t lg.Expr) *lg.NamedBinder {
		if name == "first" {
			res := l2sWhen("next", vs, t, cfg.ProofLabel)
			cfg.L2sWhensSet[res.String()] = res
			return l2sInit(vs, applyNB(res, varsToNodes(vs)...), cfg.ProofLabel)
		}
		res := l2sWhen(name, vs, t, cfg.ProofLabel)
		cfg.L2sWhensSet[res.String()] = res
		return res
	}

	cfg.ReplaceTemporals = func(n lg.Expr) lg.Expr {
		return lu.ReplaceTemporalsByNamedBinder(n,
			func(vs []*lg.Variable, body lg.Expr, env *string) *lg.NamedBinder {
				return _l2sG(vs, body, env)
			},
			func(name string, vs []*lg.Variable, body lg.Expr) *lg.NamedBinder {
				return _l2sWhen(name, vs, body)
			},
		)
	}

	modPass(cfg.ReplaceTemporals)
	cfg.NotLf = cfg.ReplaceTemporals(&lg.Not{Body: cfg.Fmla})

	// Normalize named binders
	modPass(func(n lg.Expr) lg.Expr {
		return lu.NormalizeNamedBinders(n, nil)
	})
}

// SharedStep3_CollectNamedBinders collects l2s_w and l2s_s binders from
// model.Invars (and postconds if cfg.Postconds != nil).
// This is Step 3 of l2sTacticInt.
//
// The 'full' parameter controls whether to add all state variables (l2s_full mode).
// It populates cfg.NamedBindersConjs, cfg.ToWait, cfg.ToSave.
func SharedStep3_CollectNamedBinders(cfg *InstrumentationConfig, model *temporal.NormalProgram, full bool) {
	cfg.NamedBindersConjs = make(map[string][]VarBodyPair)

	// Collect from model.Invars
	sources := make([]lg.Expr, 0, len(model.Invars)+len(cfg.Postconds))
	for _, inv := range model.Invars {
		sources = append(sources, inv.Formula.(lg.Expr))
	}
	// Ranking also collects from postconds
	for _, pc := range cfg.Postconds {
		sources = append(sources, pc.Formula.(lg.Expr))
	}
	for _, src := range sources {
		for _, b := range lu.NamedBindersAst(src) {
			cfg.NamedBindersConjs[b.Name] = append(cfg.NamedBindersConjs[b.Name],
				VarBodyPair{b.Variables, b.Body})
		}
	}
	for k, v := range cfg.NamedBindersConjs {
		cfg.NamedBindersConjs[k] = dedupeVarBodyPairs(v)
	}

	// In full mode, add all state variables to 'to_save'
	if full {
		seenSave := make(map[string]bool)
		for _, vb := range cfg.NamedBindersConjs["l2s_s"] {
			seenSave[fmt.Sprint(vb.Body)] = true
		}
		m := cfg.Mod
		for _, bnd := range model.Bindings {
			for _, act := range bnd.Action.Stmt.IterSubactions() {
				mods := actions.Modifies(act)
				for _, modSym := range mods {
					symName := modSym.Name
					if m != nil && m.Sig != nil {
						if entry, ok := m.Sig.Symbols[symName]; ok {
							vs := co.SymPlaceholders(lg.NewSymbol(symName, entry.Sort))
							var expr lg.Expr
							if len(vs) > 0 {
								expr = mustApply(lg.NewSymbol(symName, entry.Sort), varsToNodes(vs)...)
							} else {
								expr = lg.NewSymbol(symName, entry.Sort)
							}
							key := fmt.Sprint(expr)
							if !seenSave[key] {
								seenSave[key] = true
								cfg.NamedBindersConjs["l2s_s"] = append(cfg.NamedBindersConjs["l2s_s"],
									VarBodyPair{vs, expr})
							}
						}
					}
				}
			}
		}

		seenWait := make(map[string]bool)
		for _, vb := range cfg.NamedBindersConjs["l2s_w"] {
			seenWait[fmt.Sprint(vb.Body)] = true
		}
		normNotLf := lu.NormalizeNamedBinders(cfg.NotLf, nil)
		for _, b := range lu.NamedBindersAst(normNotLf) {
			if b.Name == "l2s_g" {
				negBody := co.Negate(b.Body)
				key := fmt.Sprint(negBody)
				if !seenWait[key] {
					seenWait[key] = true
					cfg.NamedBindersConjs["l2s_w"] = append(cfg.NamedBindersConjs["l2s_w"],
						VarBodyPair{b.Variables, negBody})
				}
			}
			if b.Name == "l2s_init" {
				cfg.NamedBindersConjs["l2s_init"] = append(cfg.NamedBindersConjs["l2s_init"],
					VarBodyPair{b.Variables, b.Body})
			}
		}
		cfg.NamedBindersConjs["l2s_init"] = dedupeVarBodyPairs(cfg.NamedBindersConjs["l2s_init"])
	}

	cfg.ToWait = cfg.NamedBindersConjs["l2s_w"]
	cfg.ToSave = cfg.NamedBindersConjs["l2s_s"]
}

// SharedBuildSaveAndWait builds saveState, doneWaiting, and resetW from ToWait/ToSave.
// Call after SharedStep3 and before SharedStep6.
func SharedBuildSaveAndWait(cfg *InstrumentationConfig) {
	// save_state actions
	cfg.SaveState = nil
	for _, vb := range cfg.ToSave {
		lhs := applyNB(l2sS(vb.Vars, vb.Body, cfg.ProofLabel), varsToNodes(vb.Vars)...)
		cfg.SaveState = append(cfg.SaveState, setLineno(actions.NewAssignAction(lhs, vb.Body), cfg.Lineno))
	}

	// done_waiting formulas
	cfg.DoneWaiting = nil
	for _, vb := range cfg.ToWait {
		inner := applyNB(l2sW(vb.Vars, vb.Body, cfg.ProofLabel), varsToNodes(vb.Vars)...)
		cfg.DoneWaiting = append(cfg.DoneWaiting, forall(vb.Vars, &lg.Not{Body: inner}))
	}

	// reset_w actions
	cfg.ResetW = nil
	for _, vb := range cfg.ToWait {
		lhs := applyNB(l2sW(vb.Vars, vb.Body, cfg.ProofLabel), varsToNodes(vb.Vars)...)
		var conjuncts []lg.Expr
		for _, v := range vb.Vars {
			if !cfg.FiniteSorts[v.VSort.String()] {
				conjuncts = append(conjuncts, mustApply(L2SD(v.VSort), v))
			}
		}
		conjuncts = append(conjuncts, &lg.Not{Body: vb.Body})
		negGlob := cfg.ReplaceTemporals(
			&lg.Not{Body: &lg.Globally{Environ: strPtr(cfg.ProofLabel), Body: co.Negate(vb.Body)}})
		conjuncts = append(conjuncts, negGlob)
		cfg.ResetW = append(cfg.ResetW, setLineno(actions.NewAssignAction(lhs, makeAnd(conjuncts...)), cfg.Lineno))
	}
}

// SharedStep6_BuildTableau builds the tableau axiom actions.
// Populates cfg.AssumeGAxioms, cfg.AssumeWhenAxioms, cfg.AssumeInitAxioms, cfg.AssumeWAxioms.
func SharedStep6_BuildTableau(cfg *InstrumentationConfig) {
	toG := make([]L2sGTriple, 0, len(cfg.L2sGs))
	for _, triple := range cfg.L2sGs {
		toG = append(toG, triple)
	}
	sort.Slice(toG, func(i, j int) bool {
		return fmt.Sprint(toG[i].Body) < fmt.Sprint(toG[j].Body)
	})

	// assume_g_axioms
	cfg.AssumeGAxioms = nil
	for _, triple := range toG {
		inner := &lg.Implies{
			T1: applyNB(l2sG(triple.Vars, triple.Body, triple.Environ), varsToNodes(triple.Vars)...),
			T2: triple.Body,
		}
		cfg.AssumeGAxioms = append(cfg.AssumeGAxioms,
			setLineno(actions.NewAssumeAction(forall(triple.Vars, inner)), cfg.Lineno))
	}

	// assume_when_axioms
	cfg.AssumeWhenAxioms = nil
	for _, when := range cfg.L2sWhensSet {
		inner := forall(when.Variables, &lg.Implies{
			T1: when.Body,
			T2: &lg.Eq{T1: applyNB(when, varsToNodes(when.Variables)...), T2: when.Body},
		})
		cfg.AssumeWhenAxioms = append(cfg.AssumeWhenAxioms,
			setLineno(actions.NewAssumeAction(inner), cfg.Lineno))
	}

	// assume_init_axioms
	cfg.AssumeInitAxioms = nil
	for _, vb := range cfg.NamedBindersConjs["l2s_init"] {
		applied := applyL2sInit(vb.Vars, vb.Body, cfg.ProofLabel)
		inner := forall(vb.Vars, &lg.Eq{T1: applied, T2: vb.Body})
		cfg.AssumeInitAxioms = append(cfg.AssumeInitAxioms,
			setLineno(actions.NewAssumeAction(inner), cfg.Lineno))
	}

	// assume_w_axioms
	cfg.AssumeWAxioms = nil
	for _, vb := range cfg.NamedBindersConjs["l2s_w"] {
		wApp := applyNB(l2sW(vb.Vars, vb.Body, cfg.ProofLabel), varsToNodes(vb.Vars)...)
		inner := forall(vb.Vars, &lg.Not{Body: &lg.And{Terms: []lg.Expr{vb.Body, wApp}}})
		cfg.AssumeWAxioms = append(cfg.AssumeWAxioms,
			setLineno(actions.NewAssumeAction(inner), cfg.Lineno))
	}
}

// SharedStep7_InstrumentActions instruments all binding actions with
// prop events, when events, and wait events.
func SharedStep7_InstrumentActions(cfg *InstrumentationConfig, model *temporal.NormalProgram) {
	symprops := make(map[lg.NodeKey][]*lg.NamedBinder)
	symwaits := make(map[lg.NodeKey][]*lg.NamedBinder)
	symwhens := make(map[lg.NodeKey][]*lg.NamedBinder)

	for _, triple := range cfg.L2sGs {
		prop := l2sG(triple.Vars, triple.Body, triple.Environ)
		for _, sym := range il.SymbolsAst(triple.Body) {
			symprops[lg.Key(sym)] = append(symprops[lg.Key(sym)], prop)
		}
	}
	for _, when := range cfg.L2sWhensSet {
		for _, sym := range il.SymbolsAst(when.Body) {
			symwhens[lg.Key(sym)] = append(symwhens[lg.Key(sym)], when)
		}
	}
	for _, vb := range cfg.ToWait {
		wait := l2sW(vb.Vars, vb.Body, cfg.ProofLabel)
		for _, sym := range il.SymbolsAst(vb.Body) {
			symwaits[lg.Key(sym)] = append(symwaits[lg.Key(sym)], wait)
		}
	}

	lineno := cfg.Lineno

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
						T1: &lg.And{Terms: []lg.Expr{
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
				oldcond := applyNB(l2sOld(vs, cond, cfg.ProofLabel), varsToNodes(vs)...)
				pre = append(pre, setLineno(actions.NewAssignAction(oldcond, cond), lineno))
				post = append(post, setLineno(actions.NewIfAction(
					oldcond,
					actions.NewHavocAction(applyNB(when, varsToNodes(vs)...)),
				), lineno))
			}
			if when.Name == "l2s_whenprev" {
				cond := when.Body
				post = append(post, setLineno(actions.NewIfAction(
					cond,
					actions.NewHavocAction(applyNB(when, varsToNodes(vs)...)),
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
			rhs := &lg.And{Terms: []lg.Expr{
				waitApp,
				&lg.Not{Body: t},
				cfg.ReplaceTemporals(&lg.Not{Body: &lg.Globally{
					Environ: strPtr(cfg.ProofLabel),
					Body:    co.Negate(t),
				}}),
			}}
			res = append(res, setLineno(actions.NewAssignAction(waitApp, rhs), lineno))
		}
		return res
	}

	var instrStmt func(stmt actions.Action) actions.Action
	instrStmt = func(stmt actions.Action) actions.Action {
		args := stmt.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		changed := false
		for i, a := range args {
			if sub, ok := a.(actions.Action); ok {
				newSub := instrStmt(sub)
				newArgs[i] = newSub
				if newSub != sub {
					changed = true
				}
			} else {
				newArgs[i] = a
			}
		}
		var res actions.Action
		if changed {
			res = stmt.ActionClone(newArgs)
		} else {
			res = stmt
		}

		eventProps := make(map[string]*lg.NamedBinder)
		eventWhens := make(map[string]*lg.NamedBinder)
		eventWaits := make(map[string]*lg.NamedBinder)

		modifiedSyms := actions.Modifies(stmt)
		modSet := make(map[string]bool, len(modifiedSyms))
		for _, sym := range modifiedSyms {
			modSet[sym.Name] = true
		}
		allDeps := cfg.Dependencies(modSet)
		for sym := range allDeps {
			symKey := lg.NodeKey(sym)
			for _, prop := range symprops[symKey] {
				eventProps[prop.String()] = prop
			}
			for _, when := range symwhens[symKey] {
				eventWhens[when.String()] = when
			}
			for _, wait := range symwaits[symKey] {
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

	for i, b := range model.Bindings {
		newStmt := instrStmt(b.Action.Stmt)
		model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))
	}
}

// SharedStep8_PatchExports patches exported actions with tableau axioms.
// If cfg.Postconds is non-nil (ranking mode), uses resetW instead of assumeWAxioms
// and attaches postconds to each exported action.
func SharedStep8_PatchExports(cfg *InstrumentationConfig, model *temporal.NormalProgram) {
	calls := make(map[string]bool)
	for _, c := range model.Calls {
		calls[c] = true
	}

	if model.Postconds == nil && cfg.Postconds != nil {
		model.Postconds = make(map[string][]*ast.LabeledFormula)
	}

	for i, b := range model.Bindings {
		if !calls[b.Name] {
			continue
		}
		var addParamsToD []actions.Action
		for _, p := range b.Action.Inputs {
			if p.CSort != nil && !cfg.FiniteSorts[p.CSort.String()] {
				if _, ok := p.CSort.(*lg.UninterpretedSort); ok {
					addParamsToD = append(addParamsToD,
						setLineno(actions.NewAssignAction(mustApply(L2SD(p.CSort), p), lg.True), cfg.Lineno))
				}
			}
		}

		var stmtParts []actions.Action
		stmtParts = append(stmtParts, addParamsToD...)
		stmtParts = append(stmtParts, cfg.AssumeGAxioms...)
		stmtParts = append(stmtParts, cfg.AssumeWhenAxioms...)
		if cfg.Postconds != nil {
			// Ranking mode: use reset_w instead of assume_w_axioms
			stmtParts = append(stmtParts, cfg.ResetW...)
		} else {
			// L2S mode: use assume_w_axioms
			stmtParts = append(stmtParts, cfg.AssumeWAxioms...)
		}
		stmtParts = append(stmtParts, b.Action.Stmt)
		stmtParts = append(stmtParts, cfg.AddConstsToD...)

		newStmt := setLineno(actions.ConcatActions(stmtParts...), cfg.Lineno)
		actions.CopyFormalsTo(b.Action.Stmt, newStmt)
		model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))

		if cfg.Postconds != nil {
			model.Postconds[b.Name] = cfg.Postconds
		}
	}
}

// SharedStep11_ReplaceNamedBinders replaces named binders with fresh relation constants.
func SharedStep11_ReplaceNamedBinders(cfg *InstrumentationConfig, model *temporal.NormalProgram, modPass func(func(lg.Expr) lg.Expr)) {
	namedBinders := collectAllNamedBinders(model)

	// Ensure _old_l2s_g is consistent with l2s_g
	namedBinders["_old_l2s_g"] = nil
	for _, b := range namedBinders["l2s_g"] {
		namedBinders["_old_l2s_g"] = append(namedBinders["_old_l2s_g"],
			&lg.NamedBinder{Name: "_old_l2s_g", Variables: b.Variables, Environ: b.Environ, Body: b.Body})
	}

	subs := make(map[string]lg.Expr)
	for k, binders := range namedBinders {
		for i, b := range binders {
			freshName := fmt.Sprintf("%s_%d", k, i)
			subs[b.String()] = lg.NewSymbol(freshName, b.NodeSort())
		}
	}

	modPass(func(n lg.Expr) lg.Expr {
		return lu.ReplaceNamedBindersAst(n, subs)
	})

	// Reestablish formals invariant
	for _, b := range model.Bindings {
		b.Action.Stmt.SetFormalParams(b.Action.Inputs)
		b.Action.Stmt.SetFormalReturns(b.Action.Outputs)
	}
}

// SharedStep12_BuildGoal builds the new goal with M |= true as conclusion.
func SharedStep12_BuildGoal(acfg *ast.AstConfig, goal *ast.LabeledFormula, goals []*ast.LabeledFormula, prems []ast.Node, tm *ast.TemporalModels) ([]*ast.LabeledFormula, error) {
	newConc := &ast.TemporalModels{Model: tm.Model, Fmla: lg.True}

	var nonTemporalPrems []ast.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok && lf.IsTemporal() {
			continue
		}
		nonTemporalPrems = append(nonTemporalPrems, p)
	}

	newGoal := cloneGoalWithASTConc(acfg, goal, nonTemporalPrems, newConc)

	result := make([]*ast.LabeledFormula, len(goals))
	result[0] = newGoal
	copy(result[1:], goals[1:])
	return result, nil
}

// BuildAddConstsToD builds the addConstsToD action list.
func BuildAddConstsToD(mod *modpkg.Module, uninterpretedSorts []lg.Sort, lineno ast.Location) []actions.Action {
	var addConstsToD []actions.Action
	if mod != nil && mod.Sig != nil {
		for _, s := range uninterpretedSorts {
			for _, sym := range sortedSymbols(mod.Sig) {
				if sym.CSort != nil && sym.CSort.String() == s.String() {
					addConstsToD = append(addConstsToD,
						setLineno(actions.NewAssignAction(mustApply(L2SD(s), sym), lg.True), lineno))
				}
			}
		}
	}
	return addConstsToD
}

// BuildDefnDeps builds the definition dependency map from a module.
func BuildDefnDeps(mod *modpkg.Module) map[string][]string {
	defnDeps := make(map[string][]string)
	if mod != nil {
		for _, defn := range mod.Definitions {
			f := il.DropUniversals(defn.Formula.(lg.Expr))
			if eq, ok := f.(*lg.Eq); ok {
				if app, ok := eq.T1.(*lg.Apply); ok {
					if c, ok := app.Func.(*lg.Symbol); ok {
						for _, sym := range il.SymbolsAst(eq.T2) {
							defnDeps[sym.Name] = append(defnDeps[sym.Name], c.Name)
						}
					}
				}
			}
		}
	}
	return defnDeps
}

// BuildDependenciesFunc builds a dependency closure function from defnDeps.
func BuildDependenciesFunc(defnDeps map[string][]string) func(map[string]bool) map[string]bool {
	return func(syms map[string]bool) map[string]bool {
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
}

// FindTemporalModels looks through the goal formula for a TemporalModels node.
func FindTemporalModels(goal *ast.LabeledFormula) *ast.TemporalModels {
	return findTemporalModels(goal)
}

// ExtractNormalProgram extracts a NormalProgram from a module.
func ExtractNormalProgram(m *modpkg.Module) *temporal.NormalProgram {
	return extractNormalProgram(m)
}

// CloneGoalWithASTConc clones a goal with an ast.Node conclusion.
func CloneGoalWithASTConc(acfg *ast.AstConfig, goal *ast.LabeledFormula, prems []ast.Node, conc ast.Node) *ast.LabeledFormula {
	return cloneGoalWithASTConc(acfg, goal, prems, conc)
}
