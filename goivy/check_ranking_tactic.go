// ranking_tactic.go implements the ranking tactic body: task/trigger
// extraction, invariant and postcondition generation, and delegation to
// the L2S infrastructure for action instrumentation.
//
// Ported from Python ivy_ranking.py l2s_tactic_int (lines 55-1022).
package goivy

import (
	"fmt"
	"sort"
	"strings"
)

// rankingInvariants generates invariants and postconditions for the ranking tactic.
// This extracts task/trigger definitions from the proof goal premises
// and generates the appropriate ranking invariants and postconditions.
//
// Returns (invars, postconds, strInvarMap, tasks, triggers, error).
// strInvarMap maps symbol names to their strengthened RHS expressions; the
// caller must apply prem-update to the outer prems slice using this map
// (Python ivy_ranking.py:261-269). rankingInvariants only updates its own
// internal rawTasks; it cannot reach the caller's prems.
func rankingInvariants(
	goal *LabeledFormula,
	invars []*LabeledFormula,
	proofLabel string,
	fmla Expr,
	finiteSorts map[string]bool,
	uninterpretedSorts []Sort,
	mod *Module,
) ([]*LabeledFormula, []*LabeledFormula, map[string]Expr, map[string]*Task, map[string]*LogicTrigger, error) {

	// Helper: put into nested dict
	type defnMap map[string]map[string]*Eq
	rawTasks := make(defnMap)
	rawTriggers := make(defnMap)

	dictPut := func(dct defnMap, sfx, name string, dfn *Eq) {
		if _, ok := dct[sfx]; !ok {
			dct[sfx] = make(map[string]*Eq)
		}
		dct[sfx][name] = dfn
	}

	// getAuxDefn extracts definitions from goal premises
	getAuxDefn := func(name string, dct defnMap) {
		prems := GoalPrems(goal)
		for _, prem := range prems {
			var f Expr
			if lf, ok := prem.(*LabeledFormula); ok {
				if n, ok := lf.Formula.(Expr); ok {
					f = n
				}
			}
			if f == nil {
				continue
			}
			tmp := IvyDropUniversals(f)
			eq, ok := tmp.(*Eq)
			if !ok {
				continue
			}
			var dname string
			switch lhs := eq.T1.(type) {
			case *Apply:
				if c, ok := lhs.Func.(*Const); ok {
					dname = c.Name
				}
			case *Const:
				dname = lhs.Name
			}
			if dname == "" || !strings.HasPrefix(dname, name) {
				continue
			}
			freeVars := VariablesAST(f)
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
				if impl, ok := gfmla.(*LogicImplies); ok {
					gfmla = impl.T2
				} else {
					break
				}
			}
			if g, ok := gfmla.(*LogicGlobally); ok {
				workStart := &Eq{
					T1: NewConst("work_start"+sfx, &BooleanSort{}),
					T2: &LogicNot{Body: g.Body},
				}
				dictPut(rawTriggers, sfx, "work_start", workStart)
			}
		}
	}

	// Validate
	for _, sfx := range sortedTasks {
		for _, name := range []string{"work_created", "work_needed", "work_progress", "work_helpful"} {
			if rawTasks[sfx][name] == nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("tactic requires a definition of %s%s", name, sfx)
			}
		}
	}

	// Build Task/Trigger maps
	tasks := make(map[string]*Task)
	triggers := make(map[string]*LogicTrigger)
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
		tr := &LogicTrigger{}
		if d := defs["work_start"]; d != nil {
			tr.WorkStart = d
		}
		triggers[sfx] = tr
	}

	// str_invar: strengthen each task's work_invar by adding globally formulas from work_start.
	// Python ivy_ranking.py:245-259.
	strInvarMap := make(map[string]Expr)
	for _, sfx := range sortedTasks {
		wstart := rawTriggers[sfx]["work_start"]
		winvar := rawTasks[sfx]["work_invar"]
		if wstart == nil || winvar == nil {
			continue
		}
		var globs []Expr
		rankingTrigGlob(wstart.T2, &globs, true)
		rhs := winvar.T2
		if len(globs) > 0 {
			rhs = rankingMakeAnd(append([]Expr{rhs}, globs...)...)
		}
		strengthened := &Eq{T1: winvar.T1, T2: rhs}
		rawTasks[sfx]["work_invar"] = strengthened
		var symName string
		switch lhs := winvar.T1.(type) {
		case *Apply:
			if c, ok := lhs.Func.(*Const); ok {
				symName = c.Name
			}
		case *Const:
			symName = lhs.Name
		}
		if symName != "" {
			strInvarMap[symName] = rhs
		}
	}

	// Helper functions
	eqLHSArgs := func(eq *Eq) []*LogicVariable {
		if app, ok := eq.T1.(*Apply); ok {
			var vars []*LogicVariable
			for _, t := range app.Terms {
				if v, ok := t.(*LogicVariable); ok {
					vars = append(vars, v)
				}
			}
			return vars
		}
		return nil
	}
	eqRHS := func(eq *Eq) Expr {
		return eq.T2
	}

	mklf := func(name string, fmla Expr) *LabeledFormula {
		return mod.Cfg.AstCfg.NewLabeledFormula(mod.Cfg.AstCfg.NewAtom(name), fmla)
	}

	allD := func(eq *Eq) Expr {
		args := eqLHSArgs(eq)
		var cons []Expr
		for _, v := range args {
			if v.VSort != nil && !finiteSorts[v.VSort.String()] {
				d := L2sD(v.VSort)
				app, _ := NewApply(d, v)
				if app != nil {
					cons = append(cons, app)
				}
			}
		}
		if len(cons) == 0 {
			return &LogicAnd{Terms: nil}
		}
		return &LogicImplies{T1: eq.T1, T2: rankingMakeAnd(cons...)}
	}

	// Generate invariants and postconditions
	var postconds []*LabeledFormula
	var helps []Expr

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

		// work_invar predicate application (Python: work_invar = task['work_invar'].args[D] with D=0)
		workInvarVal := workInvar.T1

		// --- l2s_created invariant ---
		invars = append(invars, mklf("l2s_created"+sfx, allD(workCreated)))

		// --- l2s_needed_implies_created postcond ---
		// Python: old_of(Implies(And(work_invar, work_needed.args[D]), work_created.args[D]))
		neededImplCreated := &LogicImplies{
			T1: rankingMakeAnd(workInvarVal, workNeeded.T1),
			T2: workCreated.T1,
		}
		postconds = append(postconds, mklf("l2s_needed_implies_created"+sfx,
			CheckOldOf(neededImplCreated)))

		// --- l2s_invar postcond ---
		wBinder := l2sW(nil, eqRHS(workStart), "")
		notWaitingForTrigger := &LogicNot{Body: wBinder}
		postconds = append(postconds, mklf("l2s_invar"+sfx,
			&LogicImplies{
				T1: &LogicOr{Terms: []Expr{CheckOldOf(workInvarVal), notWaitingForTrigger}},
				T2: workInvarVal,
			}))

		// --- l2s_needed_preserved postcond ---
		noHelp := CheckOldOf(&LogicNot{Body: &LogicOr{Terms: helps}})
		// Python: Implies(old_of(And(work_invar, Not(work_needed.args[D]))), Not(work_needed.args[D]))
		notNeeded := &LogicImplies{
			T1: CheckOldOf(rankingMakeAnd(workInvarVal, &LogicNot{Body: workNeeded.T1})),
			T2: &LogicNot{Body: workNeeded.T1},
		}
		postconds = append(postconds, mklf("l2s_needed_preserved"+sfx,
			&LogicImplies{T1: noHelp, T2: notNeeded}))

		// Add current task's help to the accumulated helps
		helps = append(helps, rankingMakeAnd(workInvarVal,
			CheckExists(helpfulArgs, eqRHS(workHelpful))))

		// --- l2s_progress postcond ---
		waitingForProgress := rankApplyNB(l2sW(progressArgs, eqRHS(workProgress), ""),
			rankVarsToNodes(progressArgs)...)

		wpargs := neededArgs
		if len(helpfulArgs) > len(progressArgs) {
			wpargs = neededArgs[len(helpfulArgs)-len(progressArgs):]
		}
		// Python: decreased = exists(wpargs, And(old_of(work_needed.args[D]), Not(work_needed.args[D])))
		decreased := CheckExists(wpargs, rankingMakeAnd(
			CheckOldOf(workNeeded.T1),
			&LogicNot{Body: workNeeded.T1}))
		progressCond := rankingMakeAnd(
			CheckOldOf(workInvarVal),
			CheckOldOf(eqRHS(workHelpful)),
			&LogicNot{Body: waitingForProgress})
		postconds = append(postconds, mklf("l2s_progress"+sfx,
			&LogicImplies{T1: progressCond, T2: decreased}))

		// --- l2s_progress_eventually postcond ---
		eventuallyProgress := &LogicEventually{Body: eqRHS(workProgress)}
		postconds = append(postconds, mklf("l2s_progress_eventually"+sfx,
			&LogicImplies{
				T1: rankingMakeAnd(CheckOldOf(workInvarVal), CheckOldOf(eqRHS(workHelpful))),
				T2: eventuallyProgress,
			}))

		// --- l2s_sched_stable postcond ---
		schedStable := CheckForAll(progressArgs,
			&LogicImplies{
				T1: rankingMakeAnd(
					CheckOldOf(workInvarVal),
					noHelp,
					CheckOldOf(eqRHS(workHelpful)),
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
		workInvarVal := workInvar.T1

		schedExists := CheckOldOf(&LogicImplies{
			T1: workInvarVal,
			T2: &LogicOr{Terms: helps},
		})
		postconds = append(postconds, mklf("l2s_sched_exists", schedExists))

		// l2s_eventually_start invariant
		// Python: lg.Eventually(proof_label, work_start.args[1]) — environ=proof_label (empty string)
		invars = append(invars, mklf("l2s_eventually_start",
			&LogicImplies{
				T1: &LogicNot{Body: workInvarVal},
				T2: &LogicEventually{Environ: strPtr(proofLabel), Body: eqRHS(workStart)},
			}))
	}

	// --- l2s_consts_d invariant ---
	var constsDTerms []Expr
	if mod != nil && mod.Sig != nil {
		for _, s := range uninterpretedSorts {
			sName := s.String()
			if finiteSorts[sName] {
				continue
			}
			for _, sym := range insertionOrderSymbols(mod) {
				if sym.CSort != nil && sym.CSort.String() == sName {
					d := L2sD(s)
					app, _ := NewApply(d, sym)
					if app != nil {
						constsDTerms = append(constsDTerms, app)
					}
				}
			}
		}
	}
	if len(constsDTerms) > 0 {
		invars = append(invars, mklf("l2s_consts_d", rankingMakeAnd(constsDTerms...)))
	}

	return invars, postconds, strInvarMap, tasks, triggers, nil
}

// rankingMakeAnd, CheckForAll, CheckExists, CheckOldOf are defined in ranking.go

func rankVarsToNodes(vs []*LogicVariable) []Expr {
	nodes := make([]Expr, len(vs))
	for i, v := range vs {
		nodes[i] = v
	}
	return nodes
}

// rankingTrigGlob collects globally/not-eventually formulas from prop in given polarity.
// Python ivy_ranking.py:221-243.
func rankingTrigGlob(prop Expr, res *[]Expr, pos bool) {
	switch p := prop.(type) {
	case *LogicGlobally:
		if pos {
			*res = append(*res, prop)
			rankingTrigGlob(p.Body, res, pos)
		}
	case *LogicEventually:
		if !pos {
			*res = append(*res, &LogicNot{Body: prop})
			rankingTrigGlob(p.Body, res, pos)
		}
	case *LogicImplies:
		if !pos {
			rankingTrigGlob(p.T1, res, !pos)
			rankingTrigGlob(p.T2, res, pos)
		}
	case *LogicAnd:
		if pos {
			for _, arg := range p.Terms {
				rankingTrigGlob(arg, res, pos)
			}
		}
	case *LogicOr:
		if !pos {
			for _, arg := range p.Terms {
				rankingTrigGlob(arg, res, pos)
			}
		}
	case *ForAll:
		if pos {
			rankingTrigGlob(p.Body, res, pos)
		}
	case *LogicExists:
		if !pos {
			rankingTrigGlob(p.Body, res, pos)
		}
	case *LogicNot:
		rankingTrigGlob(p.Body, res, !pos)
	}
}

func rankApplyNB(nb *LogicNamedBinder, args ...Expr) Expr {
	if len(args) == 0 {
		return nb
	}
	app, err := NewApply(nb, args...)
	if err != nil {
		return nb
	}
	return app
}
