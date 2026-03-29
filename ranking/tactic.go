// tactic.go implements the ranking tactic body: task/trigger extraction,
// invariant and postcondition generation, and delegation to the L2S
// infrastructure for action instrumentation.
//
// Ported from Python ivy_ranking.py l2s_tactic_int (lines 55-1022).
package ranking

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
)

// rankingInvariants generates invariants and postconditions for the ranking tactic.
// This extracts task/trigger definitions from the proof goal premises
// and generates the appropriate ranking invariants and postconditions.
//
// Returns (invars, postconds, tasks, triggers, error).
func rankingInvariants(
	goal *ast.LabeledFormula,
	invars []*ast.LabeledFormula,
	proofLabel string,
	fmla lg.Expr,
	finiteSorts map[string]bool,
	uninterpretedSorts []lg.Sort,
	mod *module.Module,
) ([]*ast.LabeledFormula, []*ast.LabeledFormula, map[string]*Task, map[string]*Trigger, error) {

	// Helper: put into nested dict
	type defnMap map[string]map[string]*lg.Eq
	rawTasks := make(defnMap)
	rawTriggers := make(defnMap)

	dictPut := func(dct defnMap, sfx, name string, dfn *lg.Eq) {
		if _, ok := dct[sfx]; !ok {
			dct[sfx] = make(map[string]*lg.Eq)
		}
		dct[sfx][name] = dfn
	}

	// getAuxDefn extracts definitions from goal premises
	getAuxDefn := func(name string, dct defnMap) {
		prems := proof.GoalPrems(goal)
		for _, prem := range prems {
			var f lg.Expr
			if lf, ok := prem.(*ast.LabeledFormula); ok {
				if n, ok := lf.Formula.(lg.Expr); ok {
					f = n
				}
			}
			if f == nil {
				continue
			}
			tmp := il.DropUniversals(f)
			eq, ok := tmp.(*lg.Eq)
			if !ok {
				continue
			}
			var dname string
			switch lhs := eq.T1.(type) {
			case *lg.Apply:
				if c, ok := lhs.Func.(*lg.Const); ok {
					dname = c.Name
				}
			case *lg.Const:
				dname = lhs.Name
			}
			if dname == "" || !strings.HasPrefix(dname, name) {
				continue
			}
			freeVars := co.VariablesAST(f)
			if len(freeVars) > 0 {
				continue
			}
			sfx := dname[len(name):]
			dictPut(dct, sfx, name, eq)
		}
	}

	getAuxDefn("work_created", rawTasks)
	getAuxDefn("work_needed", rawTasks)
	getAuxDefn("work_progress", rawTasks)
	getAuxDefn("work_invar", rawTasks)
	getAuxDefn("work_helpful", rawTasks)
	getAuxDefn("work_witness", rawTasks)
	getAuxDefn("work_start", rawTriggers)

	// Sort task suffixes
	sortedTasks := make([]string, 0, len(rawTasks))
	for sfx := range rawTasks {
		sortedTasks = append(sortedTasks, sfx)
	}
	sort.Strings(sortedTasks)

	// Infer work_start
	for _, sfx := range sortedTasks {
		if _, ok := rawTriggers[sfx]; !ok || rawTriggers[sfx]["work_start"] == nil {
			gfmla := fmla
			for {
				if impl, ok := gfmla.(*lg.Implies); ok {
					gfmla = impl.T2
				} else {
					break
				}
			}
			if g, ok := gfmla.(*lg.Globally); ok {
				workStart := &lg.Eq{
					T1: lg.NewConst("work_start"+sfx, &lg.BooleanSort{}),
					T2: &lg.Not{Body: g.Body},
				}
				dictPut(rawTriggers, sfx, "work_start", workStart)
			}
		}
	}

	// Validate
	for _, sfx := range sortedTasks {
		for _, name := range []string{"work_created", "work_needed", "work_progress", "work_helpful"} {
			if rawTasks[sfx][name] == nil {
				return nil, nil, nil, nil, fmt.Errorf("tactic requires a definition of %s%s", name, sfx)
			}
		}
	}

	// Build Task/Trigger maps
	tasks := make(map[string]*Task)
	triggers := make(map[string]*Trigger)
	for sfx, defs := range rawTasks {
		t := &Task{}
		if d := defs["work_created"]; d != nil {
			t.WorkCreated = d
		}
		if d := defs["work_needed"]; d != nil {
			t.WorkNeeded = d
		}
		if d := defs["work_progress"]; d != nil {
			t.WorkProgress = d
		}
		if d := defs["work_invar"]; d != nil {
			t.WorkInvar = d
		}
		if d := defs["work_helpful"]; d != nil {
			t.WorkHelpful = d
		}
		if d := defs["work_witness"]; d != nil {
			t.WorkWitness = d
		}
		tasks[sfx] = t
	}
	for sfx, defs := range rawTriggers {
		tr := &Trigger{}
		if d := defs["work_start"]; d != nil {
			tr.WorkStart = d
		}
		triggers[sfx] = tr
	}

	// Helper functions
	eqLHSArgs := func(eq *lg.Eq) []*lg.Variable {
		if app, ok := eq.T1.(*lg.Apply); ok {
			var vars []*lg.Variable
			for _, t := range app.Terms {
				if v, ok := t.(*lg.Variable); ok {
					vars = append(vars, v)
				}
			}
			return vars
		}
		return nil
	}
	eqRHS := func(eq *lg.Eq) lg.Expr {
		return eq.T2
	}

	substVars := func(src, dst []*lg.Variable) map[lg.NodeKey]lg.Expr {
		m := make(map[lg.NodeKey]lg.Expr)
		for i, v := range src {
			if i < len(dst) {
				sym := lg.NewConst(v.Name, v.VSort)
				m[lg.Key(sym)] = dst[i]
			}
		}
		return m
	}
	subst := func(node lg.Expr, subs map[lg.NodeKey]lg.Expr) lg.Expr {
		return co.SubstituteConstantsAST(node, subs)
	}

	mklf := func(name string, fmla lg.Expr) *ast.LabeledFormula {
		return mod.Cfg.AstCfg.NewLabeledFormula(lg.NewConst(name, &lg.BooleanSort{}), fmla)
	}

	allD := func(eq *lg.Eq) lg.Expr {
		args := eqLHSArgs(eq)
		var cons []lg.Expr
		for _, v := range args {
			if v.VSort != nil && !finiteSorts[v.VSort.String()] {
				d := L2sD(v.VSort)
				app, _ := lg.NewApply(d, v)
				if app != nil {
					cons = append(cons, app)
				}
			}
		}
		rhs := eqRHS(eq)
		if len(cons) == 0 {
			return &lg.And{Terms: nil}
		}
		return &lg.Implies{T1: rhs, T2: makeAnd(cons...)}
	}

	// Generate invariants and postconditions
	var postconds []*ast.LabeledFormula
	var helps []lg.Expr

	for _, sfx := range sortedTasks {
		task := rawTasks[sfx]
		workCreated := task["work_created"]
		workNeeded := task["work_needed"]
		workProgress := task["work_progress"]
		workInvar := task["work_invar"]
		workHelpful := task["work_helpful"]
		workStart := rawTriggers[sfx]["work_start"]

		neededArgs := eqLHSArgs(workNeeded)
		helpfulArgs := eqLHSArgs(workHelpful)
		progressArgs := eqLHSArgs(workProgress)

		// work_invar definition value
		workInvarVal := eqRHS(workInvar)

		// --- l2s_created invariant ---
		invars = append(invars, mklf("l2s_created"+sfx, allD(workCreated)))

		// --- l2s_needed_implies_created postcond ---
		createdArgs := eqLHSArgs(workCreated)
		s := substVars(neededArgs, createdArgs)
		neededImplCreated := &lg.Implies{
			T1: makeAnd(workInvarVal, eqRHS(workNeeded)),
			T2: subst(eqRHS(workCreated), s),
		}
		postconds = append(postconds, mklf("l2s_needed_implies_created"+sfx,
			OldOf(neededImplCreated)))

		// --- l2s_invar postcond ---
		wBinder := L2sW(nil, eqRHS(workStart), "")
		notWaitingForTrigger := &lg.Not{Body: wBinder}
		postconds = append(postconds, mklf("l2s_invar"+sfx,
			&lg.Implies{
				T1: &lg.Or{Terms: []lg.Expr{OldOf(workInvarVal), notWaitingForTrigger}},
				T2: workInvarVal,
			}))

		// --- l2s_needed_preserved postcond ---
		noHelp := OldOf(&lg.Not{Body: &lg.Or{Terms: helps}})
		notNeeded := &lg.Implies{
			T1: OldOf(makeAnd(workInvarVal, &lg.Not{Body: eqRHS(workNeeded)})),
			T2: &lg.Not{Body: eqRHS(workNeeded)},
		}
		postconds = append(postconds, mklf("l2s_needed_preserved"+sfx,
			&lg.Implies{T1: noHelp, T2: notNeeded}))

		// Add current task's help to the accumulated helps
		helps = append(helps, makeAnd(workInvarVal,
			Exists(helpfulArgs, eqRHS(workHelpful))))

		// --- l2s_progress postcond ---
		waitingForProgress := rankApplyNB(L2sW(progressArgs, eqRHS(workProgress), ""),
			rankVarsToNodes(progressArgs)...)

		wpargs := neededArgs
		if len(helpfulArgs) > len(progressArgs) {
			wpargs = neededArgs[len(helpfulArgs)-len(progressArgs):]
		}
		decreased := Exists(wpargs, makeAnd(
			OldOf(eqRHS(workNeeded)),
			&lg.Not{Body: eqRHS(workNeeded)}))
		progressCond := makeAnd(
			OldOf(workInvarVal),
			OldOf(eqRHS(workHelpful)),
			&lg.Not{Body: waitingForProgress})
		postconds = append(postconds, mklf("l2s_progress"+sfx,
			&lg.Implies{T1: progressCond, T2: decreased}))

		// --- l2s_progress_eventually postcond ---
		eventuallyProgress := &lg.Eventually{Body: eqRHS(workProgress)}
		postconds = append(postconds, mklf("l2s_progress_eventually"+sfx,
			&lg.Implies{
				T1: makeAnd(OldOf(workInvarVal), OldOf(eqRHS(workHelpful))),
				T2: eventuallyProgress,
			}))

		// --- l2s_sched_stable postcond ---
		schedStable := ForAll(progressArgs,
			&lg.Implies{
				T1: makeAnd(
					OldOf(workInvarVal),
					noHelp,
					OldOf(eqRHS(workHelpful)),
					waitingForProgress),
				T2: eqRHS(workHelpful),
			})
		postconds = append(postconds, mklf("l2s_sched_stable"+sfx, schedStable))
	}

	// --- l2s_sched_exists postcond ---
	if len(sortedTasks) > 0 {
		sfx := sortedTasks[0]
		workInvar := rawTasks[sfx]["work_invar"]
		workStart := rawTriggers[sfx]["work_start"]
		workInvarVal := eqRHS(workInvar)

		schedExists := OldOf(&lg.Implies{
			T1: workInvarVal,
			T2: &lg.Or{Terms: helps},
		})
		postconds = append(postconds, mklf("l2s_sched_exists", schedExists))

		// l2s_eventually_start invariant
		invars = append(invars, mklf("l2s_eventually_start",
			&lg.Implies{
				T1: &lg.Not{Body: workInvarVal},
				T2: &lg.Eventually{Body: eqRHS(workStart)},
			}))
	}

	// --- l2s_consts_d invariant ---
	var constsDTerms []lg.Expr
	if mod != nil && mod.Sig != nil {
		for _, s := range uninterpretedSorts {
			sName := s.String()
			if finiteSorts[sName] {
				continue
			}
			for _, sym := range mod.Sig.Symbols {
				if sym.Sort != nil && sym.Sort.String() == sName {
					d := L2sD(s)
					c := lg.NewConst(sym.Name, sym.Sort)
					app, _ := lg.NewApply(d, c)
					if app != nil {
						constsDTerms = append(constsDTerms, app)
					}
				}
			}
		}
	}
	if len(constsDTerms) > 0 {
		invars = append(invars, mklf("l2s_consts_d", makeAnd(constsDTerms...)))
	}

	return invars, postconds, tasks, triggers, nil
}

// makeAnd, ForAll, Exists, OldOf are defined in ranking.go

func rankVarsToNodes(vs []*lg.Variable) []lg.Expr {
	nodes := make([]lg.Expr, len(vs))
	for i, v := range vs {
		nodes[i] = v
	}
	return nodes
}

func rankApplyNB(nb *lg.NamedBinder, args ...lg.Expr) lg.Expr {
	if len(args) == 0 {
		return nb
	}
	app, err := lg.NewApply(nb, args...)
	if err != nil {
		return nb
	}
	return app
}
