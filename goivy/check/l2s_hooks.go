// l2s_hooks.go implements the L2S diagnostic trace hooks invoked by the
// trace formatter in check.go after a checker fails. These mirror Python's
// renaming_hook (ivy_l2s.py:1349-1350) and auto_hook (ivy_l2s.py:1364-1537).
package check

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// TraceHookFn is the function type stored in ast.LabeledFormula.TraceHook
// (and propagated to module.Module.TraceHook) by L2S tactics. It is invoked
// by the trace formatter in check.go after constructing a MatchHandler from
// the failing checker. Mirrors Python's goal.trace_hook closure
// (ivy_l2s.py:1311-1313).
type TraceHookFn func(handler *MatchHandler, fcs []Checker)

// TemporalAndL2S returns true for symbol names that are temporal/l2s
// auxiliary symbols (to be hidden in traces).
// Python ivy_l2s.py:1352-1354:
//
//	def temporal_and_l2s(sym):
//	    return (sym.name.startswith('l2s') and not sym.name.startswith('l2s_g')
//	            or sym.name.startswith('_old_l2s'))
func TemporalAndL2S(name string) bool {
	return (strings.HasPrefix(name, "l2s") && !strings.HasPrefix(name, "l2s_g")) ||
		strings.HasPrefix(name, "_old_l2s")
}

// L2sGToGlobally converts l2s_g named binders back to Globally operators
// for readable display. Mirrors Python ivy_l2s.py:1520-1526 ls2_g_to_globally.
func L2sGToGlobally(expr lg.Expr) lg.Expr {
	g2g := func(nb *lg.NamedBinder) lg.Expr {
		if nb.Name == "l2s_g" {
			return &lg.Globally{Environ: nb.Environ, Body: nb.Body}
		}
		return nil
	}
	res := lu.ExpandNamedBindersAst(expr, g2g)
	return lu.DenormalizeTemporal(res)
}

// markLoopStart scans MatchHandler's Eqs for l2s_saved = true and sets
// LoopStart. Mirrors Python ivy_l2s.py:113-122 trace_hook.
//
//	def trace_hook(tr,fcs):
//	    for idx,state in enumerate(tr.states):
//	        for c in state.clauses.fmlas:
//	            s1,s2 = list(map(str,c.args))
//	            if s1 == 'l2s_saved' and s2 == 'true':
//	                tr.states[0 if idx == 0 else idx-1].loop_start = True
//	                return tr
//	    print("failed to find loop start!")
//	    return tr
func markLoopStart(handler *MatchHandler) {
	if handler == nil {
		return
	}
	savedKey := lg.Key(lg.NewConst("l2s_saved", lg.Boolean))
	eqs, ok := handler.Eqs[savedKey]
	if !ok {
		fmt.Println("failed to find loop start!")
		return
	}
	for _, eq := range eqs {
		if e, ok := eq.(*lg.Eq); ok {
			if lg.IsTrue(e.T2) {
				handler.LoopStart = 0
				return
			}
		}
	}
	fmt.Println("failed to find loop start!")
}

// applyRenamingToHandler applies subs to the MatchHandler's Lines.
// subs maps {fresh-const-name → original-binder-key}; we rewrite
// occurrences in each line. Mirrors Python ivy_l2s.py:1349-1350 renaming_hook:
//
//	def renaming_hook(subs,tr,fcs):
//	    return tr.rename(dict((x,y) for (y,x) in subs.items()))
func applyRenamingToHandler(handler *MatchHandler, subs map[string]string) {
	if handler == nil || len(subs) == 0 {
		return
	}
	rsubs := make(map[string]string, len(subs))
	for k, v := range subs {
		rsubs[k] = v
	}
	for i, line := range handler.Lines {
		out := line
		for fresh, orig := range rsubs {
			out = strings.ReplaceAll(out, fresh, orig)
		}
		handler.Lines[i] = out
	}
}

// applyAutoDiagnosticsToHandler dispatches to the auto-failure diagnostic
// printer based on which checker failed. Mirrors Python ivy_l2s.py:1364-1537
// auto_hook.
func applyAutoDiagnosticsToHandler(
	handler *MatchHandler,
	fcs []Checker,
	tasks map[string]map[string]*lg.Eq,
	triggers map[string]map[string]*lg.Eq,
) {
	if handler == nil {
		return
	}
	// Python ivy_l2s.py:1529: tr.pp = ls2_g_to_globally
	handler.PP = L2sGToGlobally

	// Python: failed_fc = [fc for fc in fcs if fc.failed()][0]
	var failedFC Checker
	for _, fc := range fcs {
		if fc != nil && fc.Failed() {
			failedFC = fc
			break
		}
	}
	if failedFC == nil {
		return
	}
	lf := failedFC.GetLF()
	if lf == nil {
		return
	}

	// Python line 1379-1388: extract justice_pred_map from progress_invar
	// checkers, used by the l2s_progress_made case.
	justicePredMap := extractJusticePredMap(fcs, handler)

	name := lfName(lf)
	diagnoseAutoFailure(name, tasks, triggers, lf, handler, justicePredMap)
}

// extractJusticePredMap builds a map from task suffix to justice predicate
// symbol by scanning l2s_progress_invar checkers.
// Python ivy_l2s.py:1379-1388.
func extractJusticePredMap(fcs []Checker, handler *MatchHandler) map[string]*lg.Const {
	result := make(map[string]*lg.Const)
	if handler == nil {
		return result
	}
	for _, fc := range fcs {
		fcLF := fc.GetLF()
		if fcLF == nil {
			continue
		}
		fcName := lfName(fcLF)
		if !strings.HasPrefix(fcName, "l2s_progress_invar") {
			continue
		}
		sfx := fcName[len("l2s_progress_invar"):]
		// Python: gfmla = rsubs[lf.formula.args[1].rep]; jfmla = gfmla.body.args[0]
		// The formula structure is complex; extract the justice predicate if possible.
		if fmla, ok := fcLF.Formula.(lg.Expr); ok {
			if jp := extractJusticePred(fmla); jp != nil {
				result[sfx] = jp
			}
		}
	}
	return result
}

// extractJusticePred attempts to extract a justice predicate Const from
// an l2s_progress_invar formula. Returns nil if structure doesn't match.
func extractJusticePred(fmla lg.Expr) *lg.Const {
	// The formula is typically a ForAll wrapping an Implies.
	// We need to navigate to the justice condition.
	// Python: lf.formula.args[1].rep → gfmla; gfmla.body.args[0] → jfmla
	// This is brittle; return nil if we can't find it.
	switch f := fmla.(type) {
	case *lg.ForAll:
		return extractJusticePred(f.Body)
	case *lg.Implies:
		if eq, ok := f.T2.(*lg.Eq); ok {
			if app, ok := eq.T1.(*lg.Apply); ok {
				if c, ok := app.Func.(*lg.Const); ok {
					return c
				}
			}
		}
	}
	return nil
}

// evalSkolemInHandler evaluates a Skolem symbol (@name) by looking it up
// in the handler's Eqs map. Returns the RHS value if found, else nil.
// Mirrors Python tr.eval_in_state(state, sk) for the post-state.
func evalSkolemInHandler(handler *MatchHandler, sk *lg.Const) lg.Expr {
	if handler == nil || handler.Eqs == nil {
		return nil
	}
	key := lg.Key(sk)
	eqs, ok := handler.Eqs[key]
	if !ok || len(eqs) == 0 {
		return nil
	}
	// Each eq is an equality; extract the RHS.
	for _, eq := range eqs {
		if e, ok := eq.(*lg.Eq); ok {
			return e.T2
		}
	}
	return nil
}

// predLHSArgs extracts the LHS arguments from a predicate Eq definition.
// Python: vs = work_created.args[0].args — returns all args unchanged.
// Python duck-types .name and .sort on whatever types come through.
func predLHSArgs(pred *lg.Eq) []lg.Expr {
	if pred == nil {
		return nil
	}
	if app, ok := pred.T1.(*lg.Apply); ok {
		result := make([]lg.Expr, len(app.Terms))
		copy(result, app.Terms)
		return result
	}
	return nil
}

// predLHSRep extracts the LHS function symbol from a predicate Eq definition.
// Python: work_created.args[0].rep
func predLHSRep(pred *lg.Eq) *lg.Const {
	if pred == nil {
		return nil
	}
	switch lhs := pred.T1.(type) {
	case *lg.Apply:
		if c, ok := lhs.Func.(*lg.Const); ok {
			return c
		}
	case *lg.Const:
		return lhs
	}
	return nil
}

// makeSkolems creates Skolem constants for each arg: @name with sort.
// Python: sks = [ilg.Symbol('@'+v.name, v.sort) for v in vs]
// Python duck-types .name and .sort; we extract from Variable or Const.
func makeSkolems(vs []lg.Expr) []*lg.Const {
	sks := make([]*lg.Const, len(vs))
	for i, v := range vs {
		name, sort := exprNameSort(v)
		sks[i] = lg.NewConst("@"+name, sort)
	}
	return sks
}

// exprNameSort extracts name and sort from a Variable or Const,
// mirroring Python's duck-typed access to .name and .sort.
func exprNameSort(e lg.Expr) (string, lg.Sort) {
	switch t := e.(type) {
	case *lg.Variable:
		return t.Name, t.VSort
	case *lg.Const:
		return t.Name, t.CSort
	default:
		panic(fmt.Sprintf("exprNameSort: unhandled type %T (python duck-types .name/.sort)", e))
	}
}

// evalSkolems evaluates all Skolem symbols in the handler, returning the
// values. Returns nil if any value is missing.
// Python: vals = [tr.eval_in_state(post_state, sk) for sk in sks]
//
//	if None not in vals: ...
func evalSkolems(handler *MatchHandler, sks []*lg.Const) []lg.Expr {
	vals := make([]lg.Expr, len(sks))
	for i, sk := range sks {
		val := evalSkolemInHandler(handler, sk)
		if val == nil {
			return nil // "None in vals"
		}
		vals[i] = val
	}
	return vals
}

// applyPredToVals builds pred.rep(*vals) — applies the predicate function
// to the evaluated values. Returns a string representation.
func applyPredToVals(rep *lg.Const, vals []lg.Expr) string {
	if rep == nil {
		return "<nil>"
	}
	if len(vals) == 0 {
		return rep.Name
	}
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprint(v)
	}
	return fmt.Sprintf("%s(%s)", rep.Name, strings.Join(parts, ","))
}

// termsKey encodes up to 8 ground terms as a fixed-size string array.
// Mirrors Python tuple(eqn.args[0].args) used as a dict key.
type termsKey [8]string

// makeTermsKey builds a termsKey from a slice of ground-term expressions.
func makeTermsKey(terms []lg.Expr) termsKey {
	var k termsKey
	for i, t := range terms {
		if i >= len(k) {
			break
		}
		k[i] = fmt.Sprint(t)
	}
	return k
}

// applyFuncName returns the Const name of an Apply's function symbol, or "".
// Apply.Func is lg.Expr; this helper type-asserts to *lg.Const.
func applyFuncName(app *lg.Apply) string {
	if c, ok := app.Func.(*lg.Const); ok {
		return c.Name
	}
	return ""
}

// formatKeyWithRep formats "rep(arg0,arg1,...)" from a termsKey and a predicate rep.
// Used in diagnostic loops to display counterexample arguments.
// Mirrors Python: pred_eq.args[0].rep(*args).
func formatKeyWithRep(rep *lg.Const, k termsKey) string {
	var parts []string
	for _, s := range k {
		if s == "" {
			break
		}
		parts = append(parts, s)
	}
	if rep == nil {
		return strings.Join(parts, ",")
	}
	if len(parts) == 0 {
		return rep.Name
	}
	return fmt.Sprintf("%s(%s)", rep.Name, strings.Join(parts, ","))
}

// diagnoseAutoFailure prints diagnostic information based on the failed
// invariant name. Faithful port of the dispatch table in Python's auto_hook
// (ivy_l2s.py:1393-1537).
func diagnoseAutoFailure(
	name string,
	tasks, triggers map[string]map[string]*lg.Eq,
	lf *ast.LabeledFormula,
	handler *MatchHandler,
	justicePredMap map[string]*lg.Const,
) {
	switch {

	// Python ivy_l2s.py:1393-1405
	case strings.HasPrefix(name, "l2s_created"):
		sfx := name[len("l2s_created"):]
		fmt.Printf("\n\nFailed to prove that work_created%s is finite by induction.\n", sfx)
		task := tasks[sfx]
		if task == nil {
			break
		}
		wc := task["work_created"]
		if wc == nil {
			break
		}
		vs := predLHSArgs(wc)
		sks := makeSkolems(vs)
		vals := evalSkolems(handler, sks)
		if vals != nil {
			rep := predLHSRep(wc)
			pred := applyPredToVals(rep, vals)
			fmt.Printf("Note: %s is true in the post-state of the action, but not in the pre-state,\n", pred)
			fmt.Println("and its argument(s) are not visited during the action execution.")
		}
		if handler != nil {
			handler.HiddenSymbols = TemporalAndL2S
		}

	// Python ivy_l2s.py:1407-1423
	case strings.HasPrefix(name, "l2s_needed_when_start"):
		sfx := name[len("l2s_needed_when_start"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is a subset of work_created%s when the start condition has occurred.\n", sfx, sfx)
		task := tasks[sfx]
		if task == nil {
			break
		}
		wn := task["work_needed"]
		wc := task["work_created"]
		if wn == nil || wc == nil {
			break
		}
		vs := predLHSArgs(wn)
		sks := makeSkolems(vs)
		vals := evalSkolems(handler, sks)
		if vals != nil {
			wnRep := predLHSRep(wn)
			wcRep := predLHSRep(wc)
			pred1 := applyPredToVals(wnRep, vals)
			pred2 := applyPredToVals(wcRep, vals)
			fmt.Printf("Note: the start condition occurs during the action and %s is true in the post-state of the action, but %s is not true.\n", pred1, pred2)
		}
		if handler != nil {
			handler.HiddenSymbols = TemporalAndL2S
		}

	// Python ivy_l2s.py:1425-1437
	case strings.HasPrefix(name, "l2s_work_preserved"):
		sfx := name[len("l2s_work_preserved"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)
		task := tasks[sfx]
		if task == nil {
			break
		}
		wn := task["work_needed"]
		if wn == nil {
			break
		}
		vs := predLHSArgs(wn)
		sks := makeSkolems(vs)
		vals := evalSkolems(handler, sks)
		if vals != nil {
			rep := predLHSRep(wn)
			pred := applyPredToVals(rep, vals)
			fmt.Printf("Note: work_invar%s is true and %s changes from false to true.\n\n", sfx, pred)
		}
		if handler != nil {
			handler.HiddenSymbols = TemporalAndL2S
		}

	// Python ivy_l2s.py:1439-1451
	case strings.HasPrefix(name, "l2s_needed_are_frozen"):
		sfx := name[len("l2s_needed_are_frozen"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)
		task := tasks[sfx]
		if task == nil {
			break
		}
		wn := task["work_needed"]
		if wn == nil {
			break
		}
		vs := predLHSArgs(wn)
		sks := makeSkolems(vs)
		vals := evalSkolems(handler, sks)
		if vals != nil {
			rep := predLHSRep(wn)
			pred := applyPredToVals(rep, vals)
			fmt.Printf("Note: work_invar%s is true and %s changes from false to true.\n\n", sfx, pred)
		}
		if handler != nil {
			handler.HiddenSymbols = TemporalAndL2S
		}

	// Python ivy_l2s.py:1617-1673
	case strings.HasPrefix(name, "l2s_progress_made"):
		sfx := name[len("l2s_progress_made"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s decreases when a helpful transition occurs\n", sfx)

		task := tasks[sfx]
		if task == nil {
			break
		}

		// --- Extract nonce predicate names from invar formula ---
		// Python ivy_l2s.py:1621-1636
		// lf.Formula = Implies(And(l2s_saved, eventually_start, exists, not_all_was_done,
		//   ForAll(progress_args, Implies(nad, Not(waiting_for_progress)))), ...)
		// For l2s_auto5 (only tactic that calls applyAutoDiagnosticsToHandler):
		//   Terms[4] = ForAll(progress_args, Implies(nad, Not(waiting)))
		//   body.args[0] = nad = Apply(l2s_s_i, ...) → Func.(*lg.Const).Name = was_helpful_pred_nonce
		//   body.args[1].args[0] = Apply(l2s_w_j, ...) → Func.(*lg.Const).Name = trigger_happened_pred_nonce
		var wasHelpfulNonce, triggerNonce string
		if impl, ok := lf.Formula.(*lg.Implies); ok {
			if ant, ok := impl.T1.(*lg.And); ok && len(ant.Terms) >= 5 {
				allHH := ant.Terms[4]
				if fa, ok := allHH.(*lg.ForAll); ok {
					// ForAll case (l2s_auto5): body = Implies(nad, Not(waiting_for_progress))
					if bodyImpl, ok := fa.Body.(*lg.Implies); ok {
						// was_helpful_pred_nonce = body.args[0].rep
						if app, ok := bodyImpl.T1.(*lg.Apply); ok {
							if c, ok := app.Func.(*lg.Const); ok {
								wasHelpfulNonce = c.Name
							}
						}
						// trigger_happened_pred_nonce = body.args[1].args[0].rep
						if notExpr, ok := bodyImpl.T2.(*lg.Not); ok {
							if app, ok := notExpr.Body.(*lg.Apply); ok {
								if c, ok := app.Func.(*lg.Const); ok {
									triggerNonce = c.Name
								}
							}
						}
					}
				} else {
					// Non-forall variant: all_helpful_happened.args[0].rep
					if app, ok := allHH.(*lg.Apply); ok {
						if c, ok := app.Func.(*lg.Const); ok {
							wasHelpfulNonce = c.Name
						}
					}
				}
			}
		}
		xtracer.Trace("l2s.diagnoseAutoFailure l2s_progress_made sfx=%s wasHelpfulNonce=%s triggerNonce=%s", sfx, wasHelpfulNonce, triggerNonce)

		// --- Build helpful_map from pre-state (Python: tr.states[0].clauses.fmlas) ---
		// Python ivy_l2s.py:1627-1632
		// In Go: pre-state equalities in handler.Eqs are keyed by the bare symbol name.
		// handler.Eqs is populated from the combined 2-state clauses model, which includes
		// both pre-state atoms (bare names) and post-state atoms ("new_" prefix).
		wh := task["work_helpful"]
		helpfulMap := make(map[termsKey]bool)
		if wasHelpfulNonce != "" {
			for _, eqs := range handler.Eqs {
				for _, eqExpr := range eqs {
					eq, ok := eqExpr.(*lg.Eq)
					if !ok {
						continue
					}
					app, ok := eq.T1.(*lg.Apply)
					if !ok || applyFuncName(app) != wasHelpfulNonce {
						continue
					}
					k := makeTermsKey(app.Terms)
					helpfulMap[k] = lg.IsTrue(eq.T2)
					if wh != nil {
						rep := predLHSRep(wh)
						fmt.Printf("%s = %v\n", applyPredToVals(rep, app.Terms), eq.T2)
					}
				}
			}
		}

		// --- Build happened_maps for both states ---
		// Python ivy_l2s.py:1637-1644: for idx in range(2) over tr.states[idx]
		// In Go: tr.states[0] = bare name, tr.states[1] = "new_" + name (actions.New)
		wp := task["work_progress"]
		happenedMaps := [2]map[termsKey]bool{
			make(map[termsKey]bool),
			make(map[termsKey]bool),
		}
		if triggerNonce != "" {
			triggerNames := [2]string{triggerNonce, actions.New(triggerNonce)}
			for idx := 0; idx < 2; idx++ {
				tname := triggerNames[idx]
				fmt.Println()
				for _, eqs := range handler.Eqs {
					for _, eqExpr := range eqs {
						eq, ok := eqExpr.(*lg.Eq)
						if !ok {
							continue
						}
						app, ok := eq.T1.(*lg.Apply)
						if !ok || applyFuncName(app) != tname {
							continue
						}
						k := makeTermsKey(app.Terms)
						happenedMaps[idx][k] = lg.IsTrue(eq.T2)
						if wp != nil {
							rep := predLHSRep(wp)
							fmt.Printf("~happened %s = %v\n", applyPredToVals(rep, app.Terms), eq.T2)
						}
					}
				}
			}
		}

		// --- Build justice_map from pre-state (Python: tr.states[0].clauses.fmlas) ---
		// Python ivy_l2s.py:1645-1652
		justiceMap := make(map[termsKey]bool)
		if jp, ok := justicePredMap[sfx]; ok && jp != nil {
			fmt.Println()
			jpKey := lg.Key(jp)
			for _, eqExpr := range handler.Eqs[jpKey] {
				eq, ok := eqExpr.(*lg.Eq)
				if !ok {
					continue
				}
				app, ok := eq.T1.(*lg.Apply)
				if !ok {
					continue
				}
				k := makeTermsKey(app.Terms)
				justiceMap[k] = lg.IsTrue(eq.T2)
				if wp != nil {
					rep := predLHSRep(wp)
					fmt.Printf("~eventually %s = %v\n", applyPredToVals(rep, app.Terms), eq.T2)
				}
			}
		}

		// --- Diagnostic loop 1 ---
		// Python ivy_l2s.py:1653-1658:
		// if helpful[args] and happened[0][args]=true and happened[1][args]=false
		// → trigger occurred during action but work_needed not reduced
		if wh != nil && wp != nil {
			whRep := predLHSRep(wh)
			wpRep := predLHSRep(wp)
		outer1:
			for k, isHelpful := range helpfulMap {
				if !isHelpful {
					continue
				}
				h0, ok0 := happenedMaps[0][k]
				h1, ok1 := happenedMaps[1][k]
				if ok0 && ok1 && h0 && !h1 {
					fmt.Printf("\nNote: %s is true and %s occurs during the action, but work_needed is not reduced.\n\n",
						formatKeyWithRep(whRep, k), formatKeyWithRep(wpRep, k))
					break outer1
				}
			}
		}

		// --- Diagnostic loop 2 ---
		// Python ivy_l2s.py:1659-1664:
		// if helpful[args] and happened[0][args]=true and justice[args]=true
		// → helpful condition is eventually enabled but not satisfied
		if wh != nil && wp != nil {
			whRep := predLHSRep(wh)
			wpRep := predLHSRep(wp)
		outer2:
			for k, isHelpful := range helpfulMap {
				if !isHelpful {
					continue
				}
				h0, ok0 := happenedMaps[0][k]
				j, okJ := justiceMap[k]
				if ok0 && okJ && h0 && j {
					fmt.Printf("\nNote: %s is true and eventually %s is false.\n\n",
						formatKeyWithRep(whRep, k), formatKeyWithRep(wpRep, k))
					break outer2
				}
			}
		}

		// --- work_needed post-state eval (Python ivy_l2s.py:1665-1672) ---
		wn := task["work_needed"]
		if wn != nil {
			vs := predLHSArgs(wn)
			sks := makeSkolems(vs)
			vals := evalSkolems(handler, sks)
			if vals != nil {
				rep := predLHSRep(wn)
				pred := applyPredToVals(rep, vals)
				fmt.Printf("Note: work_invar%s is true and %s changes from false to true.\n\n", sfx, pred)
			}
		}
		// Python line 1673: tr.hidden_symbols = temporal_and_l2s (commented out in Python, omit)

	// Python ivy_l2s.py:1511-1525
	case strings.HasPrefix(name, "l2s_sched_stable"):
		sfx := name[len("l2s_sched_stable"):]
		fmt.Printf("\n\nFailed to prove that work_helpful%s is stable until helpful transition occurs\n", sfx)
		task := tasks[sfx]
		if task == nil {
			break
		}
		// Python: evaluate work_progress and work_helpful in post-state
		wp := task["work_progress"]
		wh := task["work_helpful"]
		if wp != nil && wh != nil {
			vs := predLHSArgs(wh)
			sks := makeSkolems(vs)
			vals := evalSkolems(handler, sks)
			if vals != nil {
				whRep := predLHSRep(wh)
				wpRep := predLHSRep(wp)
				pred1 := applyPredToVals(whRep, vals)
				pred2 := applyPredToVals(wpRep, vals)
				fmt.Printf("Note: %s changes and %s does not occur during the action.\n\n", pred1, pred2)
			}
		}
		if handler != nil {
			handler.HiddenSymbols = TemporalAndL2S
		}

	// Python ivy_l2s.py:1527-1530
	case strings.HasPrefix(name, "l2s_not_all_done"):
		var rankNames []string
		for sfx, task := range tasks {
			if task["work_needed"] != nil {
				rankNames = append(rankNames, "work_needed"+sfx)
			}
		}
		fmt.Printf("The ranking(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))
		if handler != nil {
			handler.HiddenSymbols = TemporalAndL2S
		}

	// Python ivy_l2s.py:1532-1535
	case strings.HasPrefix(name, "l2s_sched_exists"):
		var rankNames []string
		for sfx, task := range tasks {
			if task["work_helpful"] != nil {
				rankNames = append(rankNames, "work_helpful"+sfx)
			}
		}
		fmt.Printf("The helpful set(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))
		if handler != nil {
			handler.HiddenSymbols = TemporalAndL2S
		}
	}
}

// lfName extracts the name from a labeled formula's label.
func lfName(lf *ast.LabeledFormula) string {
	if lf == nil || lf.Label == nil {
		return ""
	}
	if c, ok := lf.Label.(*lg.Const); ok {
		return c.Name
	}
	return fmt.Sprint(lf.Label)
}
