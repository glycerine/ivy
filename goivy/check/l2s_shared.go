// l2s_shared.go contains the L2S instrumentation pipeline steps used by
// the L2S and ranking tactics. Each SharedStep* function corresponds to a
// numbered step in l2sTacticInt. The InstrumentationConfig struct carries
// all state between steps.
package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/temporal"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// InstrumentationConfig holds all state for the shared L2S instrumentation pipeline.
type InstrumentationConfig struct {
	ProofLabel         string
	Lineno             ast.Location
	FiniteSorts        map[string]bool
	UninterpretedSorts []lg.Sort
	Mod                *module.Module
	Fmla               lg.Expr               // the temporal formula
	Postconds          []*ast.LabeledFormula // nil for l2s

	// Common building blocks (populated by BuildCommonBlocks or caller)
	AddConstsToD     []actions.Action
	ResetW           []actions.Action
	AssumeGAxioms    []actions.Action
	AssumeWhenAxioms []actions.Action
	AssumeWAxioms    []actions.Action
	AssumeInitAxioms []actions.Action

	// Collected state (populated by shared steps)
	L2sGs             map[lg.NodeKey]L2sGTriple
	L2sWhensSet       map[string]*lg.NamedBinder
	NamedBindersConjs map[string][]VarBodyPair
	ToWait            []VarBodyPair
	ToSave            []VarBodyPair
	SaveState         []actions.Action
	DoneWaiting       []lg.Expr
	NotLf             lg.Expr

	// Functions built during step 1
	ReplaceTemporals func(ast.Node) ast.Node
	// Dependencies closure (built by caller from defnDeps)
	Dependencies func(map[string]bool) map[string]bool

	// C5 trace_hook plumbing: data populated by l2sAutoInvariants and
	// SharedStep11_ReplaceNamedBinders, used to attach a hook to the
	// result goal so that the check package can route diagnostics.
	// Subs is the {fresh-const-name → original-binder-key} renaming map
	// (from SharedStep11). Tasks/Triggers are the per-suffix definition
	// maps from l2sAutoInvariants (only set for l2s_auto* tactics).
	Subs     map[string]string
	Tasks    map[string]map[string]*lg.Eq
	Triggers map[string]map[string]*lg.Eq
}

// L2sGTriple holds the vars, body, and environ of an l2s_g binder.
type L2sGTriple = l2sGTriple

// VarBodyPair holds a pair of variables and a body expression.
type VarBodyPair = varBodyPair

// sortNamedBinderMap extracts values from a map[string]*lg.NamedBinder and
// returns them sorted by Canon() for deterministic cross-language ordering.
func sortNamedBinderMap(m map[string]*lg.NamedBinder) []*lg.NamedBinder {
	sorted := make([]*lg.NamedBinder, 0, len(m))
	for _, v := range m {
		sorted = append(sorted, v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return string(sorted[i].Canon()) < string(sorted[j].Canon())
	})
	return sorted
}

// sortL2sGTriples extracts values from a map[lg.NodeKey]L2sGTriple and
// returns them sorted by Body.Canon() for deterministic cross-language ordering.
func sortL2sGTriples(m map[lg.NodeKey]L2sGTriple) []L2sGTriple {
	sorted := make([]L2sGTriple, 0, len(m))
	for _, v := range m {
		sorted = append(sorted, v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return string(sorted[i].Body.Canon()) < string(sorted[j].Body.Canon())
	})
	return sorted
}

// SharedStep1_ConvertTemporals converts temporal operators to named binders.
// This is Step 1 of l2sTacticInt. It populates cfg.L2sGs, cfg.L2sWhensSet,
// cfg.ReplaceTemporals, and cfg.NotLf.
//
// modPass should apply a transform to the entire model (invars, asms, bindings, init, invars list, and postconds if applicable).
func SharedStep1_ConvertTemporals(cfg *InstrumentationConfig, model *temporal.NormalProgram, modPass func(string, func(ast.Node) ast.Node)) {
	cfg.L2sGs = make(map[lg.NodeKey]L2sGTriple)
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

	cfg.ReplaceTemporals = func(n ast.Node) ast.Node {
		return lu.ReplaceTemporalsByNamedBinder(n,
			func(vs []*lg.Variable, body lg.Expr, env *string) *lg.NamedBinder {
				return _l2sG(vs, body, env)
			},
			func(name string, vs []*lg.Variable, body lg.Expr) *lg.NamedBinder {
				return _l2sWhen(name, vs, body)
			},
		)
	}

	modPass("ReplaceTemporals", cfg.ReplaceTemporals)
	xtracer.Trace("l2s.SharedStep1 TOPLEVEL_NotLf_START fmla HASH canon=%s", cfg.Fmla.Canon())
	lu.ResetRtrDepth()
	cfg.NotLf = cfg.ReplaceTemporals(&lg.Not{Body: cfg.Fmla}).(lg.Expr)
	xtracer.Trace("l2s.SharedStep1 notLf HASH canon=%s", cfg.NotLf.Canon())

	// Normalize named binders
	modPass("NormalizeNamedBinders", func(n ast.Node) ast.Node {
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
	for srcIdx, src := range sources {
		for _, b := range lu.NamedBindersAst(src) {
			xtracer.Trace("l2s.SharedStep3 collecting binder name=%s fromSource=%d nVars=%d HASH canon=%s", b.Name, srcIdx, len(b.Variables), b.Body.Canon())
			cfg.NamedBindersConjs[b.Name] = append(cfg.NamedBindersConjs[b.Name],
				VarBodyPair{b.Variables, b.Body})
		}
	}
	for k, v := range cfg.NamedBindersConjs {
		cfg.NamedBindersConjs[k] = dedupeVarBodyPairs(v)
	}
	// Trace named_binders_conjs after dedup (sorted keys for determinism)
	{
		keys := make([]string, 0, len(cfg.NamedBindersConjs))
		for k := range cfg.NamedBindersConjs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			xtracer.Trace("l2s.SharedStep3 namedBindersConjs key=%s nEntries=%d", k, len(cfg.NamedBindersConjs[k]))
		}
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
						if entry, ok := m.Sig.Symbols.Get2(symName); ok {
							vs := module.SymPlaceholders(lg.NewConst(symName, entry.Sort))
							var expr lg.Expr
							if len(vs) > 0 {
								expr = mustApply(lg.NewConst(symName, entry.Sort), varsToNodes(vs)...)
							} else {
								expr = lg.NewConst(symName, entry.Sort)
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
		normNotLf := lu.NormalizeNamedBinders(cfg.NotLf, nil).(lg.Expr)
		for _, b := range lu.NamedBindersAst(normNotLf) {
			if b.Name == "l2s_g" {
				negBody := module.Negate(b.Body)
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
	for i, vb := range cfg.ToWait {
		xtracer.Trace("l2s.SharedStep3 toWait[%d] nVars=%d HASH canon=%s", i, len(vb.Vars), vb.Body.Canon())
	}
	for i, vb := range cfg.ToSave {
		xtracer.Trace("l2s.SharedStep3 toSave[%d] nVars=%d HASH canon=%s", i, len(vb.Vars), vb.Body.Canon())
	}
}

// SharedBuildSaveAndWait builds saveState, doneWaiting, and resetW from ToWait/ToSave.
// Call after SharedStep3 and before SharedStep6.
func SharedBuildSaveAndWait(cfg *InstrumentationConfig) {
	// save_state actions
	cfg.SaveState = nil
	for i, vb := range cfg.ToSave {
		xtracer.Trace("l2s.SharedBuildSaveAndWait saveState[%d] nVars=%d HASH canon=%s", i, len(vb.Vars), vb.Body.Canon())
		lhs := applyNB(l2sS(vb.Vars, vb.Body, cfg.ProofLabel), varsToNodes(vb.Vars)...)
		cfg.SaveState = append(cfg.SaveState, setLineno(actions.NewAssignAction(lhs, vb.Body), cfg.Lineno))
	}

	// done_waiting formulas
	cfg.DoneWaiting = nil
	for i, vb := range cfg.ToWait {
		inner := applyNB(l2sW(vb.Vars, vb.Body, cfg.ProofLabel), varsToNodes(vb.Vars)...)
		xtracer.Trace("l2s.SharedBuildSaveAndWait doneWaiting[%d] nVars=%d HASH canon=%s", i, len(vb.Vars), inner.Canon())
		cfg.DoneWaiting = append(cfg.DoneWaiting, forall(vb.Vars, &lg.Not{Body: inner}))
	}

	// reset_w actions
	cfg.ResetW = nil
	for i, vb := range cfg.ToWait {
		xtracer.Trace("l2s.SharedBuildSaveAndWait resetW[%d] nVars=%d body HASH canon=%s", i, len(vb.Vars), vb.Body.Canon())
		lhs := applyNB(l2sW(vb.Vars, vb.Body, cfg.ProofLabel), varsToNodes(vb.Vars)...)
		var conjuncts []lg.Expr
		for _, v := range vb.Vars {
			if !cfg.FiniteSorts[v.VSort.String()] {
				conjuncts = append(conjuncts, mustApply(L2SD(v.VSort), v))
			}
		}
		conjuncts = append(conjuncts, &lg.Not{Body: vb.Body})
		negatedBody := module.Negate(vb.Body)
		xtracer.Trace("l2s.SharedBuildSaveAndWait resetW[%d] negatedBody HASH canon=%s", i, negatedBody.Canon())
		preReplaceInput := &lg.Not{Body: &lg.Globally{Environ: strPtr(cfg.ProofLabel), Body: negatedBody}}
		xtracer.Trace("l2s.SharedBuildSaveAndWait resetW[%d] preReplace HASH canon=%s", i, preReplaceInput.Canon())
		lu.ResetRtrDepth()
		negGlob := cfg.ReplaceTemporals(preReplaceInput).(lg.Expr)
		xtracer.Trace("l2s.SharedBuildSaveAndWait resetW[%d] postReplace HASH canon=%s", i, negGlob.Canon())
		conjuncts = append(conjuncts, negGlob)
		cfg.ResetW = append(cfg.ResetW, setLineno(actions.NewAssignAction(lhs, makeAnd(conjuncts...)), cfg.Lineno))
	}
	xtracer.Trace("l2s.SharedBuildSaveAndWait EXIT nSaveState=%d nDoneWaiting=%d nResetW=%d", len(cfg.SaveState), len(cfg.DoneWaiting), len(cfg.ResetW))
}

// SharedStep6_BuildTableau builds the tableau axiom actions.
// Populates cfg.AssumeGAxioms, cfg.AssumeWhenAxioms, cfg.AssumeInitAxioms, cfg.AssumeWAxioms.
func SharedStep6_BuildTableau(cfg *InstrumentationConfig) {
	toG := make([]L2sGTriple, 0, len(cfg.L2sGs))
	for _, triple := range cfg.L2sGs {
		toG = append(toG, triple)
	}
	sort.Slice(toG, func(i, j int) bool {
		return string(toG[i].Body.Canon()) < string(toG[j].Body.Canon())
	})

	// assume_g_axioms
	for i, triple := range toG {
		xtracer.Trace("l2s.SharedStep6 toG[%d] nVars=%d HASH canon=%s", i, len(triple.Vars), triple.Body.Canon())
	}
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
	// Python ivy_l2s.py:1005-1008:
	//   AssumeAction(forall(when.variables, lg.Implies(when.body.t1, lg.Eq(when(*when.variables), when.body.t2))))
	// when.body is a Cond(condition, value); decompose T1=condition, T2=value.
	cfg.AssumeWhenAxioms = nil
	sortedWhens := sortNamedBinderMap(cfg.L2sWhensSet)
	for _, when := range sortedWhens {
		cond, ok := when.Body.(*lg.Cond)
		if !ok {
			continue
		}
		inner := forall(when.Variables, &lg.Implies{
			T1: cond.T1,
			T2: &lg.Eq{T1: applyNB(when, varsToNodes(when.Variables)...), T2: cond.T2},
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
	xtracer.Trace("l2s.SharedStep6 EXIT nAssumeG=%d nAssumeWhen=%d nAssumeInit=%d nAssumeW=%d",
		len(cfg.AssumeGAxioms), len(cfg.AssumeWhenAxioms), len(cfg.AssumeInitAxioms), len(cfg.AssumeWAxioms))
}

// SharedStep7_InstrumentActions instruments all binding actions with
// prop events, when events, and wait events.
func SharedStep7_InstrumentActions(cfg *InstrumentationConfig, model *temporal.NormalProgram) {
	symprops := make(map[string][]*lg.NamedBinder)
	symwaits := make(map[string][]*lg.NamedBinder)
	symwhens := make(map[string][]*lg.NamedBinder)

	sortedTriples := sortL2sGTriples(cfg.L2sGs)
	for _, triple := range sortedTriples {
		prop := l2sG(triple.Vars, triple.Body, triple.Environ)
		for sym := range il.SymbolsIluAst(triple.Body) {
			if c, ok := sym.(*lg.Const); ok {
				symprops[c.Name] = append(symprops[c.Name], prop)
			}
		}
	}
	sortedWhens7 := sortNamedBinderMap(cfg.L2sWhensSet)
	for _, when := range sortedWhens7 {
		for sym := range il.SymbolsIluAst(when.Body) {
			if c, ok := sym.(*lg.Const); ok {
				symwhens[c.Name] = append(symwhens[c.Name], when)
			}
		}
	}
	for wi, vb := range cfg.ToWait {
		wait := l2sW(vb.Vars, vb.Body, cfg.ProofLabel)
		si := 0
		for sym := range il.SymbolsIluAst(vb.Body) {
			if c, ok := sym.(*lg.Const); ok {
				xtracer.Trace("l2s.SharedStep7 symwaits toWait[%d] sym[%d]=%s HASH canon=%s", wi, si, c.Name, vb.Body.Canon())
				symwaits[c.Name] = append(symwaits[c.Name], wait)
				si++
			}
		}
	}

	lineno := cfg.Lineno

	propEventsFunc := func(gprops map[string]*lg.NamedBinder) ([]actions.Action, []actions.Action) {
		var pre, post []actions.Action
		sortedProps := sortNamedBinderMap(gprops)
		for _, gprop := range sortedProps {
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
		for _, gprop := range sortedProps {
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

	// Python ivy_l2s.py:1070-1085: when_events.
	// when.body is a Cond(condition, value); decompose T1=cond, T2=val.
	whenEventsFunc := func(whens map[string]*lg.NamedBinder) ([]actions.Action, []actions.Action) {
		var pre, post []actions.Action
		sortedWhens := sortNamedBinderMap(whens)
		for _, when := range sortedWhens {
			condVal, ok := when.Body.(*lg.Cond)
			if !ok {
				continue
			}
			vs := when.Variables
			cond := condVal.T1
			if when.Name == "l2s_whennext" {
				oldcond := applyNB(l2sOld(vs, cond, cfg.ProofLabel), varsToNodes(vs)...)
				pre = append(pre, setLineno(actions.NewAssignAction(oldcond, cond), lineno))
				post = append(post, setLineno(actions.NewIfAction(
					oldcond,
					actions.NewHavocAction(applyNB(when, varsToNodes(vs)...)),
				), lineno))
			}
			if when.Name == "l2s_whenprev" {
				post = append(post, setLineno(actions.NewIfAction(
					cond,
					actions.NewHavocAction(applyNB(when, varsToNodes(vs)...)),
				), lineno))
			}
		}
		for _, when := range sortedWhens {
			condVal, ok := when.Body.(*lg.Cond)
			if !ok {
				continue
			}
			post = append(post,
				actions.NewAssumeAction(forall(when.Variables,
					&lg.Implies{
						T1: condVal.T1,
						T2: &lg.Eq{T1: applyNB(when, varsToNodes(when.Variables)...), T2: condVal.T2},
					})))
		}
		return pre, post
	}

	waitEventsFunc := func(waits map[string]*lg.NamedBinder) []actions.Action {
		var res []actions.Action
		sortedWaits := sortNamedBinderMap(waits)
		xtracer.Trace("l2s.SharedStep7 waitEventsFunc nWaits=%d", len(sortedWaits))
		for wi, wait := range sortedWaits {
			xtracer.Trace("l2s.SharedStep7 waitEventsFunc wait[%d] HASH canon=%s", wi, wait.Canon())
			vs, t := wait.Variables, wait.Body
			waitApp := applyNB(wait, varsToNodes(vs)...)
			rhs := &lg.And{Terms: []lg.Expr{
				waitApp,
				&lg.Not{Body: t},
				cfg.ReplaceTemporals(&lg.Not{Body: &lg.Globally{
					Environ: strPtr(cfg.ProofLabel),
					Body:    module.Negate(t),
				}}).(lg.Expr),
			}}
			res = append(res, setLineno(actions.NewAssignAction(waitApp, rhs), lineno))
		}
		return res
	}

	var instrStmt func(stmt actions.Action) actions.Action
	instrStmt = func(stmt actions.Action) actions.Action {
		// H7 / Python ivy_l2s.py:1127-1131: if a CallAction's returns
		// include any monitored symbol (in symprops, symwhens, or symwaits),
		// split the call so the assignment is a separate statement.
		if call, ok := stmt.(*actions.CallAction); ok {
			callArgs := call.ActionArgs()
			if len(callArgs) > 1 {
				returns := callArgs[1:]
				monitored := false
				for _, r := range returns {
					if c, ok := r.(*lg.Const); ok {
						k := c.Name
						if len(symprops[k]) > 0 || len(symwhens[k]) > 0 || len(symwaits[k]) > 0 {
							monitored = true
							break
						}
					}
				}
				if monitored {
					split := call.SplitReturns(actions.NewActionsConfig())
					return instrStmt(split)
				}
			}
		}

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
		{
			sortedDeps := make([]string, 0, len(allDeps))
			for sym := range allDeps {
				sortedDeps = append(sortedDeps, sym)
			}
			sort.Strings(sortedDeps)
			sortedMods := make([]string, 0, len(modSet))
			for sym := range modSet {
				sortedMods = append(sortedMods, sym)
			}
			sort.Strings(sortedMods)
			xtracer.Trace("l2s.SharedStep7 instrStmt mods=[%s] deps=[%s]", strings.Join(sortedMods, ","), strings.Join(sortedDeps, ","))
		}
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

	for i, b := range model.Bindings {
		newStmt := instrStmt(b.Action.Stmt)
		model.Bindings[i] = b.CloneAction(b.Action.CloneStmt(newStmt))
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
		// Python ivy_l2s.py:1228-1231: add l2s_d for all non-finite-sort inputs.
		var addParamsToD []actions.Action
		for _, p := range b.Action.Inputs {
			if p.CSort != nil && !cfg.FiniteSorts[p.CSort.String()] {
				addParamsToD = append(addParamsToD,
					setLineno(actions.NewAssignAction(mustApply(L2SD(p.CSort), p), lg.True), cfg.Lineno))
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
		model.Bindings[i] = b.CloneAction(b.Action.CloneStmt(newStmt))

		if cfg.Postconds != nil {
			model.Postconds[b.Name] = cfg.Postconds
		}
	}
}

// SharedStep11_ReplaceNamedBinders replaces named binders with fresh relation constants.
func SharedStep11_ReplaceNamedBinders(cfg *InstrumentationConfig, model *temporal.NormalProgram, modPass func(string, func(ast.Node) ast.Node)) {
	namedBinders := collectAllNamedBinders(model)
	{
		keys := make([]string, 0, len(namedBinders))
		for k := range namedBinders {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			xtracer.Trace("l2s.SharedStep11 namedBinders key=%s count=%d", k, len(namedBinders[k]))
		}
	}

	// Ensure _old_l2s_g is consistent with l2s_g
	namedBinders["_old_l2s_g"] = nil
	for _, b := range namedBinders["l2s_g"] {
		namedBinders["_old_l2s_g"] = append(namedBinders["_old_l2s_g"],
			&lg.NamedBinder{Name: "_old_l2s_g", Variables: b.Variables, Environ: b.Environ, Body: b.Body})
	}

	subs := make(map[string]lg.Expr)
	// C5: also build a string-keyed inverse map for the trace hook.
	// Maps fresh-const-name → original-binder-key string. The renaming
	// hook later builds the reverse for trace display.
	if cfg.Subs == nil {
		cfg.Subs = make(map[string]string)
	}
	for k, binders := range namedBinders {
		for i, b := range binders {
			freshName := fmt.Sprintf("%s_%d", k, i)
			subs[b.String()] = lg.NewConst(freshName, b.NodeSort())
			cfg.Subs[freshName] = b.String()
			xtracer.Trace("l2s.SharedStep11 sub freshName=%s binderKey=%s", freshName, b.String())
		}
	}

	modPass("ReplaceNamedBindersAst", func(n ast.Node) ast.Node {
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
	newConc := acfg.NewTemporalModels(tm.Model, lg.True)

	var nonTemporalPrems []ast.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok && lf.IsTemporal() {
			continue
		}
		nonTemporalPrems = append(nonTemporalPrems, p)
	}

	newGoal := proof.CloneGoal(acfg, goal, nonTemporalPrems, newConc)

	result := make([]*ast.LabeledFormula, len(goals))
	result[0] = newGoal
	copy(result[1:], goals[1:])
	return result, nil
}

// BuildAddConstsToD builds the addConstsToD action list.
func BuildAddConstsToD(mod *module.Module, uninterpretedSorts []lg.Sort, lineno ast.Location) []actions.Action {
	var addConstsToD []actions.Action
	if mod != nil && mod.Sig != nil {
		for _, s := range uninterpretedSorts {
			for _, sym := range insertionOrderSymbols(mod) {
				if sym.CSort != nil && sym.CSort.String() == s.String() {
					addConstsToD = append(addConstsToD,
						setLineno(actions.NewAssignAction(mustApply(L2SD(s), sym), lg.True), lineno))
				}
			}
		}
	}
	return addConstsToD
}

// insertionOrderSymbols returns module symbols in insertion order (matching
// Python's ilg.sig.symbols.values()). Sig.Symbols is an InsMap that
// preserves insertion order and deduplicates by key automatically.
func insertionOrderSymbols(mod *module.Module) []*lg.Const {
	if mod == nil || mod.Sig == nil || mod.Sig.Symbols.Len() == 0 {
		return nil
	}
	result := make([]*lg.Const, 0, mod.Sig.Symbols.Len())
	for name, entry := range mod.Sig.Symbols.All() {
		result = append(result, lg.NewConst(name, entry.Sort))
	}
	return result
}

// BuildDefnDeps builds the definition dependency map from a module.
// H12 / Python ivy_l2s.py:159-166: also include premise definitions
// from the goal (premises with IsDefinition == true), not just the
// module-level definitions.
func BuildDefnDeps(mod *module.Module, goalPrems ...ast.Node) map[string][]string {
	defnDeps := make(map[string][]string)
	addEq := func(formula lg.Expr) {
		f := il.DropUniversals(formula)
		if eq, ok := f.(*lg.Eq); ok {
			if app, ok := eq.T1.(*lg.Apply); ok {
				if c, ok := app.Func.(*lg.Const); ok {
					for x := range il.SymbolsIluAst(eq.T2) {
						if sym, ok := x.(*lg.Const); ok {
							defnDeps[sym.Name] = append(defnDeps[sym.Name], c.Name)
						}
					}
				}
			}
		}
	}
	if mod != nil {
		for _, defn := range mod.Definitions {
			if e, ok := defn.Formula.(lg.Expr); ok {
				addEq(e)
			}
		}
	}
	// H12: include user-supplied definition premises from the goal.
	for _, p := range goalPrems {
		lf, ok := p.(*ast.LabeledFormula)
		if !ok || !lf.IsDefinition {
			continue
		}
		if e, ok := lf.Formula.(lg.Expr); ok {
			addEq(e)
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
func ExtractNormalProgram(m *module.Module) *temporal.NormalProgram {
	return extractNormalProgram(m)
}

// (CloneGoalWithASTConc was deleted — callers should use proof.CloneGoal
// directly, which now accepts ast.Node for conc.)
