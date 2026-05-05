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
// Returns (invars, postconds, tasks, triggers, error).
func rankingInvariants(
	goal *LabeledFormula,
	invars []*LabeledFormula,
	proofLabel string,
	fmla Expr,
	finiteSorts map[string]bool,
	uninterpretedSorts []Sort,
	mod *Module,
) ([]*LabeledFormula, []*LabeledFormula, map[string]*Task, map[string]*Trigger, error) {

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
				if impl, ok := gfmla.(*Implies); ok {
					gfmla = impl.T2
				} else {
					break
				}
			}
			if g, ok := gfmla.(*Globally); ok {
				workStart := &Eq{
					T1: NewConst("work_start"+sfx, &BooleanSort{}),
					T2: &Not{Body: g.Body},
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
	eqLHSArgs := func(eq *Eq) []*Variable {
		if app, ok := eq.T1.(*Apply); ok {
			var vars []*Variable
			for _, t := range app.Terms {
				if v, ok := t.(*Variable); ok {
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

	substVars := func(src, dst []*Variable) map[NodeKey]Expr {
		m := make(map[NodeKey]Expr)
		for i, v := range src {
			if i < len(dst) {
				sym := NewConst(v.Name, v.VSort)
				m[Key(sym)] = dst[i]
			}
		}
		return m
	}
	subst := func(node Expr, subs map[NodeKey]Expr) Expr {
		return SubstituteConstantsExpr(node, subs)
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
		rhs := eqRHS(eq)
		if len(cons) == 0 {
			return &And{Terms: nil}
		}
		return &Implies{T1: rhs, T2: rankingMakeAnd(cons...)}
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

		// work_invar definition value
		workInvarVal := eqRHS(workInvar)

		// --- l2s_created invariant ---
		invars = append(invars, mklf("l2s_created"+sfx, allD(workCreated)))

		// --- l2s_needed_implies_created postcond ---
		createdArgs := eqLHSArgs(workCreated)
		s := substVars(neededArgs, createdArgs)
		neededImplCreated := &Implies{
			T1: rankingMakeAnd(workInvarVal, eqRHS(workNeeded)),
			T2: subst(eqRHS(workCreated), s),
		}
		postconds = append(postconds, mklf("l2s_needed_implies_created"+sfx,
			CheckOldOf(neededImplCreated)))

		// --- l2s_invar postcond ---
		wBinder := l2sW(nil, eqRHS(workStart), "")
		notWaitingForTrigger := &Not{Body: wBinder}
		postconds = append(postconds, mklf("l2s_invar"+sfx,
			&Implies{
				T1: &Or{Terms: []Expr{CheckOldOf(workInvarVal), notWaitingForTrigger}},
				T2: workInvarVal,
			}))

		// --- l2s_needed_preserved postcond ---
		noHelp := CheckOldOf(&Not{Body: &Or{Terms: helps}})
		notNeeded := &Implies{
			T1: CheckOldOf(rankingMakeAnd(workInvarVal, &Not{Body: eqRHS(workNeeded)})),
			T2: &Not{Body: eqRHS(workNeeded)},
		}
		postconds = append(postconds, mklf("l2s_needed_preserved"+sfx,
			&Implies{T1: noHelp, T2: notNeeded}))

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
		decreased := CheckExists(wpargs, rankingMakeAnd(
			CheckOldOf(eqRHS(workNeeded)),
			&Not{Body: eqRHS(workNeeded)}))
		progressCond := rankingMakeAnd(
			CheckOldOf(workInvarVal),
			CheckOldOf(eqRHS(workHelpful)),
			&Not{Body: waitingForProgress})
		postconds = append(postconds, mklf("l2s_progress"+sfx,
			&Implies{T1: progressCond, T2: decreased}))

		// --- l2s_progress_eventually postcond ---
		eventuallyProgress := &Eventually{Body: eqRHS(workProgress)}
		postconds = append(postconds, mklf("l2s_progress_eventually"+sfx,
			&Implies{
				T1: rankingMakeAnd(CheckOldOf(workInvarVal), CheckOldOf(eqRHS(workHelpful))),
				T2: eventuallyProgress,
			}))

		// --- l2s_sched_stable postcond ---
		schedStable := CheckForAll(progressArgs,
			&Implies{
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
		workInvarVal := eqRHS(workInvar)

		schedExists := CheckOldOf(&Implies{
			T1: workInvarVal,
			T2: &Or{Terms: helps},
		})
		postconds = append(postconds, mklf("l2s_sched_exists", schedExists))

		// l2s_eventually_start invariant
		invars = append(invars, mklf("l2s_eventually_start",
			&Implies{
				T1: &Not{Body: workInvarVal},
				T2: &Eventually{Body: eqRHS(workStart)},
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

	return invars, postconds, tasks, triggers, nil
}

// rankingMakeAnd, CheckForAll, CheckExists, CheckOldOf are defined in ranking.go

func rankVarsToNodes(vs []*Variable) []Expr {
	nodes := make([]Expr, len(vs))
	for i, v := range vs {
		nodes[i] = v
	}
	return nodes
}

func rankApplyNB(nb *NamedBinder, args ...Expr) Expr {
	if len(args) == 0 {
		return nb
	}
	app, err := NewApply(nb, args...)
	if err != nil {
		return nb
	}
	return app
}
