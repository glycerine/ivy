// l2s_auto.go implements the l2s_auto tactic's task/trigger invariant
// generation, ported from Python ivy_l2s.py lines 196-678.
package goivy

import (
	"fmt"
	"sort"
	"strings"
)

// l2sAutoInvariants generates invariants for l2s_auto tactics.
// It extracts task/trigger definitions from the proof goal premises
// and generates the appropriate L2S invariants. Returns the augmented
// invars list plus the tasks and triggers maps (used by C5 trace_hook).
func l2sAutoInvariants(
	tacticName string,
	goal *LabeledFormula,
	invars []*LabeledFormula,
	proofLabel string,
	fmla Expr,
	finiteSorts map[string]bool,
	uninterpretedSorts []Sort,
	m *Module,
	proofLineno Location, // mirrors Python proof.lineno — source location of the proof tactic
) ([]*LabeledFormula, map[string]map[string]*Eq, map[string]map[string]*Eq, error) {
	if !strings.HasPrefix(tacticName, "l2s_auto") {
		return invars, nil, nil, nil
	}

	autoAcfg := m.Cfg.AstCfg

	// Helper: put into nested dict
	type defnMap map[string]map[string]*Eq // sfx -> name -> definition (Eq node)
	tasks := make(defnMap)
	triggers := make(defnMap)

	dictPut := func(dct defnMap, sfx, name string, dfn *Eq) {
		if _, ok := dct[sfx]; !ok {
			dct[sfx] = make(map[string]*Eq)
		}
		dct[sfx][name] = dfn
	}

	// getAuxDefn extracts definitions like work_created, work_needed, etc.
	// from the goal premises.
	// C18: filter by IsDefinition (Python ivy_l2s.py:208-212) and raise an
	// error on free variables (Python ivy_l2s.py:214) instead of silently
	// skipping.
	getAuxDefn := func(name string, dct defnMap) error {
		prems := GoalPrems(goal)
		for _, prem := range prems {
			// Try to get a definition from the premise.
			premLF, isLF := prem.(*LabeledFormula)
			if !isLF {
				continue
			}
			// C18: only consider premises marked as definitions.
			if !premLF.IsDefinition {
				continue
			}
			f, ok := premLF.Formula.(Expr)
			if !ok || f == nil {
				continue
			}
			// H9 / Python ivy_l2s.py:210-212: strip ONLY one outer CheckForAll,
			// not all levels (DropUniversals would over-strip nested
			// quantifiers, producing ill-formed results).
			var tmp Expr = f
			if fa, isFA := tmp.(*ForAll); isFA {
				tmp = fa.Body
			}
			eq, ok := tmp.(*Eq)
			if !ok {
				continue
			}
			// Get the defined name
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
			// C18: error on free variables instead of silently skipping.
			freeVars := VariablesAST(f)
			if len(freeVars) > 0 {
				return fmt.Errorf("free symbol %s not allowed in definition of %s", freeVars[0].Name, dname)
			}
			// H10 / Python ivy_l2s.py:218-219: rebuild the LHS with the
			// generic name (e.g. "work_created" instead of "work_created0")
			// so downstream code uses a stable lookup.
			renamedEq := eq
			switch lhs := eq.T1.(type) {
			case *Apply:
				if c, ok := lhs.Func.(*Const); ok {
					newC := NewConst(name, c.CSort)
					if newApp, _ := NewApply(newC, lhs.Terms...); newApp != nil {
						renamedEq = &Eq{T1: newApp, T2: eq.T2}
					}
				}
			case *Const:
				renamedEq = &Eq{T1: NewConst(name, lhs.CSort), T2: eq.T2}
			}
			sfx := dname[len(name):]
			dictPut(dct, sfx, name, renamedEq)
		}
		return nil
	}

	if err := getAuxDefn("work_created", tasks); err != nil {
		return nil, nil, nil, err
	}
	if err := getAuxDefn("work_needed", tasks); err != nil {
		return nil, nil, nil, err
	}
	if err := getAuxDefn("work_done", tasks); err != nil {
		return nil, nil, nil, err
	}
	if err := getAuxDefn("work_progress", tasks); err != nil {
		return nil, nil, nil, err
	}
	if err := getAuxDefn("work_end", tasks); err != nil {
		return nil, nil, nil, err
	}
	if err := getAuxDefn("work_invar", tasks); err != nil {
		return nil, nil, nil, err
	}
	if tacticName == "l2s_auto5" {
		if err := getAuxDefn("work_helpful", tasks); err != nil {
			return nil, nil, nil, err
		}
	}
	if err := getAuxDefn("work_start", triggers); err != nil {
		return nil, nil, nil, err
	}

	// Sort task suffixes
	sortedTasks := make([]string, 0, len(tasks))
	for sfx := range tasks {
		sortedTasks = append(sortedTasks, sfx)
	}
	sort.Strings(sortedTasks)

	// Infer work_start from globally formula if not provided
	if len(sortedTasks) > 0 {
		sfx := sortedTasks[0]
		if _, ok := triggers[sfx]; !ok || triggers[sfx]["work_start"] == nil {
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
				dictPut(triggers, sfx, "work_start", workStart)
			}
		}
	}

	// For l2s_auto5: infer work_done if only work_needed is given
	if tacticName == "l2s_auto5" {
		for _, sfx := range sortedTasks {
			task := tasks[sfx]
			if task["work_needed"] != nil && task["work_done"] == nil {
				workNeeded := task["work_needed"]
				workDone := &Eq{
					T1: cloneLHS(workNeeded.T1, "work_done"+sfx),
					T2: &LogicOr{Terms: nil}, // false = empty disjunction
				}
				dictPut(tasks, sfx, "work_done", workDone)
			}
		}
	}

	// Validate required definitions
	for _, sfx := range sortedTasks {
		for _, name := range []string{"work_created", "work_needed", "work_done", "work_progress"} {
			if tasks[sfx][name] == nil {
				return nil, nil, nil, fmt.Errorf("tactic l2s_auto requires a definition of %s%s", name, sfx)
			}
		}
	}

	// H4 / Python ivy_l2s.py:320-321: work_invar must have no arguments.
	for _, sfx := range sortedTasks {
		wi := tasks[sfx]["work_invar"]
		if wi == nil {
			continue
		}
		if app, ok := wi.T1.(*Apply); ok && len(app.Terms) > 0 {
			return nil, nil, nil, fmt.Errorf("work_invar%s may not have arguments", sfx)
		}
	}

	// H5 / Python ivy_l2s.py:314-319: sort-signature consistency checks.
	for _, sfx := range sortedTasks {
		task := tasks[sfx]
		if !sameLHSSort(task["work_created"], task["work_needed"]) {
			return nil, nil, nil, fmt.Errorf("work_created%s and work_needed%s must have same signature", sfx, sfx)
		}
		if !sameLHSSort(task["work_created"], task["work_done"]) {
			return nil, nil, nil, fmt.Errorf("work_created%s and work_done%s must have same signature", sfx, sfx)
		}
		if h := task["work_helpful"]; h != nil {
			if !sameLHSSort(h, task["work_progress"]) {
				return nil, nil, nil, fmt.Errorf("work_helpful%s and work_progress%s must have same signature", sfx, sfx)
			}
		}
	}

	// Monitor symbols
	l2sWaiting := L2SWaiting()
	l2sSaved := L2SSaved()

	// Helpers
	forall := func(vs []*LogicVariable, body Expr) Expr {
		if len(vs) == 0 {
			return body
		}
		return &ForAll{Variables: vs, Body: body}
	}
	exists := func(vs []*LogicVariable, body Expr) Expr {
		if len(vs) == 0 {
			return body
		}
		return &LogicExists{Variables: vs, Body: body}
	}

	// Defn helpers
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

	substVars := func(src, dst []*LogicVariable) map[NodeKey]Expr {
		m := make(map[NodeKey]Expr)
		for i, v := range src {
			if i < len(dst) {
				// Key by Symbol with same name/sort so SubstituteConstantsAST matches
				sym := NewConst(v.Name, v.VSort)
				m[Key(sym)] = dst[i]
			}
		}
		return m
	}
	subst := func(node Expr, subs map[NodeKey]Expr) Expr {
		// Python: lu.substitute (logic_util.substitute), not substitute_constants_ast.
		// lu.substitute does NOT emit the substitute_constants_action trace.
		result, err := Substitute(node, subs)
		if err != nil {
			return node // fallback: return unchanged on error
		}
		return result
	}

	// all_d: all elements in l2s_d
	// M4 / Python ivy_l2s.py:327-329: always returns Implies(rhs, And(...)),
	// even when cons is empty (And() = true).
	allD := func(eq *Eq) Expr {
		args := eqLHSArgs(eq)
		var cons []Expr
		for _, v := range args {
			if v.VSort != nil && !finiteSorts[v.VSort.String()] {
				d := L2SD(v.VSort)
				app, _ := NewApply(d, v)
				if app != nil {
					cons = append(cons, app)
				}
			}
		}
		return &LogicImplies{T1: eqRHS(eq), T2: checkMakeAnd(cons...)}
	}

	// all_a: all elements in l2s_a
	// M4 / Python ivy_l2s.py:333-335: same shape as all_d.
	allA := func(eq *Eq) Expr {
		args := eqLHSArgs(eq)
		var cons []Expr
		for _, v := range args {
			if v.VSort != nil && !finiteSorts[v.VSort.String()] {
				a := L2SA(v.VSort)
				app, _ := NewApply(a, v)
				if app != nil {
					cons = append(cons, app)
				}
			}
		}
		return &LogicImplies{T1: eqRHS(eq), T2: checkMakeAnd(cons...)}
	}

	// all_created: needed elements are in created set
	allCreated := func(defnNeeded *Eq, sfxIdx int) Expr {
		sfx := sortedTasks[sfxIdx]
		workCreated := tasks[sfx]["work_created"]
		createdArgs := eqLHSArgs(workCreated)
		defnArgs := eqLHSArgs(defnNeeded)
		s := substVars(defnArgs, createdArgs)
		return &LogicImplies{T1: subst(eqRHS(defnNeeded), s), T2: eqRHS(workCreated)}
	}

	// eventuallyStartTask
	eventuallyStartTask := func(workStart *Eq) Expr {
		if workStart == nil {
			return &LogicAnd{Terms: nil} // true
		}
		trigRHS := eqRHS(workStart)
		// H11 / Python ivy_l2s.py:384: Eventually with proof_label environ.
		evf := &LogicEventually{Environ: strPtr(proofLabel), Body: trigRHS}
		vs := eqLHSArgs(workStart)
		vsNodes := checkVarsToNodes(vs)
		initNB := l2sInit(vs, evf, proofLabel)
		return forall(vs, applyNB(initNB, vsNodes...))
	}

	// Accumulators per Python ivy_l2s.py:263-265.
	// notAllDonePreds resets at scheduler boundaries (when emitting intermediate
	// l2s_not_all_done<sfx>). notAllWasDonePreds resets when a new task with
	// work_start is encountered (Python ivy_l2s.py:472-473). schedExistsPreds
	// is used only by l2s_auto5.
	var notAllDonePreds []Expr
	var notAllWasDonePreds []Expr
	var schedExistsPreds []Expr

	// Track the last evStart for use in the auto5 final l2s_sched_exists invariant.
	var lastEvStart Expr

	// Generate invariants per task
	for idx, sfx := range sortedTasks {
		task := tasks[sfx]
		workCreated := task["work_created"]
		workNeeded := task["work_needed"]
		workDone := task["work_done"]
		workProgress := task["work_progress"]
		workEnd := task["work_end"]
		workInvar := task["work_invar"]
		var workStart *Eq
		if trig, ok := triggers[sfx]; ok {
			workStart = trig["work_start"]
		}

		doneArgs := eqLHSArgs(workDone)
		workHelpful := task["work_helpful"]

		// notWaitingForStart
		var notWaitingForStart Expr = &LogicAnd{Terms: nil} // true
		if workStart != nil {
			evStart := eventuallyStartTask(workStart)
			trigRHS := eqRHS(workStart)
			wBinder := l2sW(nil, trigRHS, proofLabel)
			notWaitingForStart = checkMakeAnd(evStart,
				&LogicOr{Terms: []Expr{
					&LogicNot{Body: l2sWaiting},
					&LogicNot{Body: wBinder},
				}})
		}

		// --- Closures used by l2s_progress_made and l2s_sched_stable ---
		// These mirror Python ivy_l2s.py:236-372 and capture per-task state.

		// getWorkWasDone: Python ivy_l2s.py:236-245
		// For non-auto3/4/5: l2s_s(done_args, work_done.RHS)(done_args), then Implies(subst(defn.RHS), wasDone)
		// For auto3/4/5: l2s_s(done_args, Implies(subst(defn.RHS), work_done.RHS))(done_args)
		getWorkWasDone := func(defn *Eq, workDoneL *Eq) Expr {
			doneArgsL := eqLHSArgs(workDoneL)
			defnArgs := eqLHSArgs(defn)
			s := substVars(defnArgs, doneArgsL)
			defnsubs := subst(eqRHS(defn), s)
			if tacticName != "l2s_auto3" && tacticName != "l2s_auto4" && tacticName != "l2s_auto5" {
				sNB := l2sS(doneArgsL, eqRHS(workDoneL), proofLabel)
				wd := applyNB(sNB, checkVarsToNodes(doneArgsL)...)
				return &LogicImplies{T1: defnsubs, T2: wd}
			}
			sNB := l2sS(doneArgsL, &LogicImplies{T1: defnsubs, T2: eqRHS(workDoneL)}, proofLabel)
			return applyNB(sNB, checkVarsToNodes(doneArgsL)...)
		}

		// getWasDone: Python ivy_l2s.py:358-360
		getWasDone := func(defn *Eq) Expr {
			return getWorkWasDone(defn, workDone)
		}

		// notAllWasDone: Python ivy_l2s.py:361-363
		// Returns Not(forall(workDone.LHS.Args[skip:], getWasDone(defn)))
		notAllWasDone := func(defn *Eq, skip int) Expr {
			tmp := getWasDone(defn)
			if skip >= len(doneArgs) {
				return &LogicNot{Body: tmp}
			}
			return &LogicNot{Body: forall(doneArgs[skip:], tmp)}
		}

		// getDepends: Python ivy_l2s.py:365-372
		// Substitutes work_helpful.LHS.Args -> work_progress.LHS.Args into work_helpful.RHS.
		getDepends := func() Expr {
			if workHelpful == nil {
				return &LogicAnd{Terms: nil} // true (workHelpful only used for auto5)
			}
			helpfulArgs := eqLHSArgs(workHelpful)
			progressArgsL := eqLHSArgs(workProgress)
			s := substVars(helpfulArgs, progressArgsL)
			return subst(eqRHS(workHelpful), s)
		}

		// nextTaskHasTrigger: Python ivy_l2s.py:462-463
		nextTaskHasTrigger := func() bool {
			if idx+1 >= len(sortedTasks) {
				return false
			}
			_, ok := triggers[sortedTasks[idx+1]]
			return ok
		}

		// nextTaskNotTriggered: Python ivy_l2s.py:465-468
		nextTaskNotTriggered := func() Expr {
			nextSfx := sortedTasks[idx+1]
			trigf := triggers[nextSfx]["work_start"]
			return &LogicNot{Body: eventuallyStartTask(trigf)}
		}

		// Reset notAllWasDonePreds at scheduler boundaries.
		// Python ivy_l2s.py:472-473.
		if workStart != nil {
			notAllWasDonePreds = nil
		}

		// --- l2s_needed_when_start ---
		if tacticName != "l2s_auto5" {
			tmp := &LogicImplies{T1: notWaitingForStart, T2: allCreated(workNeeded, idx)}
			invars = appendLF(autoAcfg, invars, "l2s_needed_when_start"+sfx, tmp, proofLineno)
		} else {
			tmp := &LogicImplies{T1: notWaitingForStart, T2: allD(workNeeded)}
			invars = appendLF(autoAcfg, invars, "l2s_needed_when_start"+sfx, tmp, proofLineno)
		}

		// --- l2s_created ---
		invars = appendLF(autoAcfg, invars, "l2s_created"+sfx, allD(workCreated), proofLineno)

		// --- l2s_needed_are_frozen ---
		evStart := eventuallyStartTask(workStart)
		if tacticName != "l2s_auto4" && tacticName != "l2s_auto5" {
			tmp := &LogicImplies{
				T1: checkMakeAnd(evStart, &LogicNot{Body: l2sWaiting}),
				T2: allA(workNeeded),
			}
			invars = appendLF(autoAcfg, invars, "l2s_needed_are_frozen"+sfx, tmp, proofLineno)
		} else {
			doneSubArgs := eqLHSArgs(workDone)
			neededArgs := eqLHSArgs(workNeeded)
			s := substVars(neededArgs, doneSubArgs)
			isDone := &LogicImplies{T1: subst(eqRHS(workNeeded), s), T2: eqRHS(workDone)}
			notIsDone := &LogicNot{Body: isDone}
			var aCons []Expr
			for _, v := range doneSubArgs {
				if v.VSort != nil && !finiteSorts[v.VSort.String()] {
					a := L2SA(v.VSort)
					app, _ := NewApply(a, v)
					if app != nil {
						aCons = append(aCons, app)
					}
				}
			}
			tmp := &LogicImplies{
				T1: checkMakeAnd(evStart, &LogicNot{Body: l2sWaiting}),
				T2: &LogicImplies{T1: notIsDone, T2: checkMakeAnd(aCons...)},
			}
			invars = appendLF(autoAcfg, invars, "l2s_needed_are_frozen"+sfx, tmp, proofLineno)

			// C12 / Python ivy_l2s.py:422-425: l2s_needed_were_frozen
			// uses get_was_done(work_needed) which for auto4/5 expands to
			//   l2s_s(done_args, Implies(subst(needed.RHS), done.RHS))(done_args).
			isDoneImplied := &LogicImplies{T1: subst(eqRHS(workNeeded), s), T2: eqRHS(workDone)}
			sNBwere := l2sS(doneSubArgs, isDoneImplied, proofLabel)
			wasDoneWere := applyNB(sNBwere, checkVarsToNodes(doneSubArgs)...)
			tmp2 := &LogicImplies{
				T1: checkMakeAnd(evStart, l2sSaved),
				T2: &LogicImplies{T1: &LogicNot{Body: wasDoneWere}, T2: checkMakeAnd(aCons...)},
			}
			invars = appendLF(autoAcfg, invars, "l2s_needed_were_frozen"+sfx, tmp2, proofLineno)
		}

		// --- l2s_done_implies_created ---
		if tacticName != "l2s_auto3" && tacticName != "l2s_auto4" && tacticName != "l2s_auto5" {
			createdArgs := eqLHSArgs(workCreated)
			s := substVars(doneArgs, createdArgs)
			tmp := &LogicImplies{T1: subst(eqRHS(workDone), s), T2: eqRHS(workCreated)}
			invars = appendLF(autoAcfg, invars, "l2s_done_implies_created"+sfx, tmp, proofLineno)
		}

		// --- l2s_needed_implies_created ---
		if tacticName != "l2s_auto5" {
			createdArgs := eqLHSArgs(workCreated)
			neededArgs := eqLHSArgs(workNeeded)
			s := substVars(neededArgs, createdArgs)
			tmp := &LogicImplies{
				T1: notWaitingForStart,
				T2: &LogicImplies{T1: subst(eqRHS(workNeeded), s), T2: eqRHS(workCreated)},
			}
			invars = appendLF(autoAcfg, invars, "l2s_needed_implies_created"+sfx, tmp, proofLineno)
		}

		// --- l2s_work_preserved ---
		var wasDone, isDoneNode Expr
		if tacticName != "l2s_auto4" && tacticName != "l2s_auto5" {
			sNB := l2sS(doneArgs, eqRHS(workDone), proofLabel)
			wasDone = applyNB(sNB, checkVarsToNodes(doneArgs)...)
			isDoneNode = eqRHS(workDone)
		} else {
			neededArgs := eqLHSArgs(workNeeded)
			s := substVars(neededArgs, doneArgs)
			neededSubbed := subst(eqRHS(workNeeded), s)
			isDoneImplied := &LogicImplies{T1: neededSubbed, T2: eqRHS(workDone)}

			sNB := l2sS(doneArgs, isDoneImplied, proofLabel)
			wasDone = applyNB(sNB, checkVarsToNodes(doneArgs)...)
			isDoneNode = isDoneImplied
		}
		tmp := &LogicImplies{T1: checkMakeAnd(l2sSaved, wasDone), T2: isDoneNode}
		if tacticName == "l2s_auto5" {
			tmp = &LogicImplies{T1: evStart, T2: tmp}
		}
		invars = appendLF(autoAcfg, invars, "l2s_work_preserved"+sfx, tmp, proofLineno)

		// --- l2s_progress_made ---
		// C11 / Python ivy_l2s.py:478-505. Three branches:
		//   - l2s_auto5 special form
		//   - non-auto5 with progress_args>0 or len(tasks)>1: forall+exists complex form
		//   - non-auto5 simple form
		progressArgs := eqLHSArgs(workProgress)
		// H3 / Python ivy_l2s.py:475-476: work_progress args must be a prefix
		// of work_done args (except for l2s_auto5).
		if tacticName != "l2s_auto5" {
			if len(progressArgs) > len(doneArgs) {
				return nil, nil, nil, fmt.Errorf("work_progess parameters must be a prefix of work_done parameters")
			}
			for i := range progressArgs {
				if progressArgs[i].Name != doneArgs[i].Name {
					return nil, nil, nil, fmt.Errorf("work_progess parameters must be a prefix of work_done parameters")
				}
			}
		}
		waitingForProgress := l2sW(progressArgs, eqRHS(workProgress), proofLabel)

		if tacticName != "l2s_auto3" {
			var progressInv Expr

			if tacticName == "l2s_auto5" {
				// Python lines 479-490
				nad := getDepends()
				nad = applyNB(l2sS(progressArgs, nad, proofLabel), checkVarsToNodes(progressArgs)...)
				if nextTaskHasTrigger() {
					nad = checkMakeAnd(nextTaskNotTriggered(), nad)
				}
				progressInv = &LogicImplies{
					T1: checkMakeAnd(
						l2sSaved,
						evStart,
						exists(progressArgs, nad),
						notAllWasDone(workNeeded, 0),
						forall(progressArgs, &LogicImplies{
							T1: nad,
							T2: &LogicNot{Body: applyNB(waitingForProgress, checkVarsToNodes(progressArgs)...)},
						}),
					),
					T2: exists(doneArgs, checkMakeAnd(&LogicNot{Body: wasDone}, isDoneNode)),
				}
			} else if len(progressArgs) > 0 || len(tasks) > 1 {
				// Python lines 491-501
				nad := checkMakeAnd(
					notAllWasDone(workNeeded, len(progressArgs)),
					&LogicNot{Body: buildOrExpr(notAllWasDonePreds)},
				)
				if nextTaskHasTrigger() {
					nad = checkMakeAnd(nextTaskNotTriggered(), nad)
				}
				skipDoneArgs := doneArgs
				if len(progressArgs) <= len(doneArgs) {
					skipDoneArgs = doneArgs[len(progressArgs):]
				}
				progressInv = forall(progressArgs, &LogicImplies{
					T1: checkMakeAnd(
						nad,
						l2sSaved,
						evStart,
						&LogicNot{Body: applyNB(waitingForProgress, checkVarsToNodes(progressArgs)...)},
					),
					T2: exists(skipDoneArgs, checkMakeAnd(&LogicNot{Body: wasDone}, isDoneNode)),
				})
			} else {
				// Python lines 502-505 (the simple case)
				progressInv = &LogicImplies{
					T1: checkMakeAnd(l2sSaved, &LogicNot{Body: waitingForProgress}),
					T2: exists(doneArgs, checkMakeAnd(&LogicNot{Body: wasDone}, isDoneNode)),
				}
			}
			invars = appendLF(autoAcfg, invars, "l2s_progress_made"+sfx, progressInv, proofLineno)
		}

		// --- l2s_progress_invar ---
		// H11: ensure Globally/Eventually carry the proof label so dedup
		// matches binders built through SharedStep1.
		gBody := &LogicGlobally{
			Environ: strPtr(proofLabel),
			Body: &LogicEventually{
				Environ: strPtr(proofLabel),
				Body:    eqRHS(workProgress),
			},
		}
		initNB := l2sInit(progressArgs, gBody, proofLabel)
		invars = appendLF(autoAcfg, invars, "l2s_progress_invar"+sfx,
			&LogicImplies{
				T1: applyNB(initNB, checkVarsToNodes(progressArgs)...),
				T2: gBody,
			}, proofLineno)

		// --- l2s_not_all_done ---
		// Builds the per-task `not_all_done(work_needed)` predicate
		// (Python ivy_l2s.py:345-356).
		neededArgs := eqLHSArgs(workNeeded)
		s := substVars(neededArgs, doneArgs)
		var notAllDoneBody Expr = &LogicImplies{T1: subst(eqRHS(workNeeded), s), T2: eqRHS(workDone)}
		if workEnd != nil {
			endArgs := eqLHSArgs(workEnd)
			endSubs := substVars(endArgs, doneArgs)
			notAllDoneBody = &LogicImplies{T1: subst(eqRHS(workEnd), endSubs), T2: notAllDoneBody}
		}
		if workInvar != nil {
			notAllDoneBody = &LogicImplies{T1: eqRHS(workInvar), T2: notAllDoneBody}
		}
		var notAllDone Expr = &LogicNot{Body: forall(doneArgs, notAllDoneBody)}

		// C3 / Python ivy_l2s.py:354-355: wrap with Or(Not(notWaitingForStart),...)
		// for auto2/3/4/5.
		if tacticName == "l2s_auto2" || tacticName == "l2s_auto3" ||
			tacticName == "l2s_auto4" || tacticName == "l2s_auto5" {
			notAllDone = &LogicOr{Terms: []Expr{
				&LogicNot{Body: notWaitingForStart},
				notAllDone,
			}}
		}

		// C4 / Python ivy_l2s.py:519-526: accumulate per-task predicates.
		// At each scheduler boundary (next task has a trigger), emit an
		// intermediate l2s_not_all_done<sfx> and reset the accumulator.
		notAllDonePreds = append(notAllDonePreds, notAllDone)
		if idx+1 < len(sortedTasks) {
			nextSfx := sortedTasks[idx+1]
			if trig, ok := triggers[nextSfx]; ok && trig["work_start"] != nil {
				trigf := trig["work_start"]
				orOfPreds := buildOrExpr(notAllDonePreds)
				tmp := &LogicImplies{
					T1: &LogicNot{Body: eventuallyStartTask(trigf)},
					T2: orOfPreds,
				}
				invars = appendLF(autoAcfg, invars, "l2s_not_all_done"+sfx, tmp, proofLineno)
				notAllDonePreds = nil
			}
		}

		// Python ivy_l2s.py:528: append this task's not_all_was_done to the
		// per-scheduler accumulator (used by next iteration's l2s_progress_made).
		notAllWasDonePreds = append(notAllWasDonePreds, notAllWasDone(workNeeded, 0))

		// C13 / Python ivy_l2s.py:530-543: l2s_sched_stable + sched_exists_preds
		// for auto5.
		if tacticName == "l2s_auto5" {
			nad := getDepends()
			wasNad := applyNB(l2sS(progressArgs, nad, proofLabel), checkVarsToNodes(progressArgs)...)
			waitingForProgressApp := applyNB(waitingForProgress, checkVarsToNodes(progressArgs)...)
			stableInv := forall(progressArgs, &LogicImplies{
				T1: checkMakeAnd(wasNad, l2sSaved, evStart, waitingForProgressApp),
				T2: nad,
			})
			invars = appendLF(autoAcfg, invars, "l2s_sched_stable"+sfx, stableInv, proofLineno)
			schedExistsPreds = append(schedExistsPreds, exists(progressArgs, wasNad))
		}

		// Track the last evStart for use by C14 (l2s_sched_exists, auto5).
		lastEvStart = evStart
	}

	// C4 final / Python ivy_l2s.py:567-568: ALWAYS emit l2s_not_all_done as
	// the OR of all (remaining) accumulated per-task predicates.
	// Python: tmp = lg.Or(*not_all_done_preds) — empty Or = false.
	{
		var nadExpr Expr
		if len(notAllDonePreds) > 0 {
			nadExpr = buildOrExpr(notAllDonePreds)
		} else {
			nadExpr = False
		}
		invars = appendLF(autoAcfg, invars, "l2s_not_all_done", nadExpr, proofLineno)
	}

	// C14 / Python ivy_l2s.py:548-550: l2s_sched_exists invariant for auto5.
	// Python's `eventually_start()` here refers to the last loop iteration's
	// work_start (closure captured), so we use the tracked lastEvStart.
	if tacticName == "l2s_auto5" && len(schedExistsPreds) > 0 {
		evStartHere := lastEvStart
		if evStartHere == nil {
			evStartHere = &LogicAnd{Terms: nil} // true
		}
		invars = appendLF(autoAcfg, invars, "l2s_sched_exists", &LogicImplies{
			T1: checkMakeAnd(l2sSaved, evStartHere),
			T2: buildOrExpr(schedExistsPreds),
		}, proofLineno)
	}

	// --- init_globally: generate l2s_globally invariants ---
	// Key knownInits by Sexp (structural canonical form), mirroring
	// Python's set() at ivy_l2s.py:626 (known_inits.add(prop) at 582,
	// `cond not in known_inits` at 671) which uses Expr struct equality.
	// fmt.Sprint(prop) calls String=PrettyFmla and drops sort annotations.
	knownInits := make(map[string]bool)
	var ninvs []Expr

	var initGlobally func(prop Expr, res *[]Expr, pos bool)
	initGlobally = func(prop Expr, res *[]Expr, pos bool) {
		switch p := prop.(type) {
		case *LogicGlobally:
			knownInits[string(prop.Sexp())] = true
			if pos {
				*res = append(*res, prop)
				initGlobally(p.Body, res, pos)
			} else {
				// ~pos, Globally: treat as Eventually(~body)
				arg := p.Body
				vs := collectVarsSlice(arg)
				wNB := l2sW(vs, &LogicNot{Body: arg}, proofLabel)
				notWaiting := &LogicOr{Terms: []Expr{
					&LogicNot{Body: l2sWaiting},
					&LogicNot{Body: applyNB(wNB, checkVarsToNodes(vs)...)},
				}}
				*res = append(*res, &LogicImplies{T1: prop, T2: notWaiting})
				// C16 / Python ivy_l2s.py:572-576: AG-pattern extra invariant.
				// Must come BEFORE l2s_init to match Python's append order.
				if isEventuallyOrNotGlobally(arg) {
					*res = append(*res, &LogicImplies{T1: notWaiting, T2: &LogicNot{Body: arg}})
				}
				initNB := l2sInit(vs, &LogicNot{Body: prop}, proofLabel)
				*res = append(*res, applyNB(initNB, checkVarsToNodes(vs)...))
			}
		case *LogicEventually:
			knownInits[string(prop.Sexp())] = true
			if !pos {
				*res = append(*res, &LogicNot{Body: prop})
				initGlobally(p.Body, res, pos)
			} else {
				arg := p.Body
				vs := collectVarsSlice(arg)
				wNB := l2sW(vs, arg, proofLabel)
				notWaiting := &LogicOr{Terms: []Expr{
					&LogicNot{Body: l2sWaiting},
					&LogicNot{Body: applyNB(wNB, checkVarsToNodes(vs)...)},
				}}
				*res = append(*res, &LogicImplies{T1: &LogicNot{Body: prop}, T2: notWaiting})
				// C16 / Python ivy_l2s.py:564-568: EF-pattern extra invariant.
				// Must come BEFORE l2s_init to match Python's append order.
				if isGloballyOrNotEventually(arg) {
					*res = append(*res, &LogicImplies{T1: notWaiting, T2: arg})
				}
				initNB := l2sInit(vs, prop, proofLabel)
				*res = append(*res, applyNB(initNB, checkVarsToNodes(vs)...))
			}
		case *LogicImplies:
			if !pos {
				initGlobally(p.T1, res, !pos)
				initGlobally(p.T2, res, pos)
			}
		case *LogicAnd:
			if pos {
				for _, arg := range p.Terms {
					initGlobally(arg, res, pos)
				}
			}
		case *LogicOr:
			if !pos {
				for _, arg := range p.Terms {
					initGlobally(arg, res, pos)
				}
			}
		case *ForAll:
			if pos {
				initGlobally(p.Body, res, pos)
			}
		case *LogicExists:
			if !pos {
				initGlobally(p.Body, res, pos)
			}
		case *LogicNot:
			initGlobally(p.Body, res, !pos)
		}
	}

	initGlobally(fmla, &ninvs, false)

	// Trigger-based invariants
	for _, sfx := range sortedTasks {
		trig, ok := triggers[sfx]
		if !ok {
			continue
		}
		trigDef := trig["work_start"]
		if trigDef == nil {
			continue
		}
		arg := eqRHS(trigDef)
		// H11: Eventually with proof_label environ.
		evf := &LogicEventually{Environ: strPtr(proofLabel), Body: arg}
		vs := eqLHSArgs(trigDef)
		vsNodes := checkVarsToNodes(vs)
		initNB := l2sInit(vs, evf, proofLabel)
		initF := applyNB(initNB, vsNodes...)
		ninvs = append(ninvs, &LogicOr{Terms: []Expr{initF, &LogicNot{Body: evf}}})
		wNB := l2sW(vs, arg, proofLabel)
		notWaiting := &LogicOr{Terms: []Expr{
			&LogicNot{Body: l2sWaiting},
			&LogicNot{Body: applyNB(wNB, vsNodes...)},
		}}
		ninvs = append(ninvs, &LogicImplies{
			T1: checkMakeAnd(initF, &LogicNot{Body: evf}),
			T2: notWaiting,
		})
		var tinvs []Expr
		initGlobally(arg, &tinvs, true)
		for _, tinv := range tinvs {
			ninvs = append(ninvs, &LogicImplies{
				T1: checkMakeAnd(initF, notWaiting),
				T2: tinv,
			})
		}
	}

	for i, ninv := range ninvs {
		invars = appendLF(autoAcfg, invars, fmt.Sprintf("l2s_globally_%d", i), ninv, proofLineno)
	}

	// M3: invariant emission order matches Python
	// (globally → status → consts_d → init_glob → neg_prop → when).

	// --- l2s_status invariants ---
	invars = appendLF(autoAcfg, invars, "l2s_status_0",
		&LogicOr{Terms: []Expr{l2sWaiting, L2SFrozen(), l2sSaved}}, proofLineno)
	invars = appendLF(autoAcfg, invars, "l2s_status_1",
		&LogicOr{Terms: []Expr{&LogicNot{Body: l2sWaiting}, &LogicNot{Body: L2SFrozen()}}}, proofLineno)
	invars = appendLF(autoAcfg, invars, "l2s_status_2",
		&LogicOr{Terms: []Expr{&LogicNot{Body: l2sWaiting}, &LogicNot{Body: l2sSaved}}}, proofLineno)
	invars = appendLF(autoAcfg, invars, "l2s_status_3",
		&LogicOr{Terms: []Expr{&LogicNot{Body: L2SFrozen()}, &LogicNot{Body: l2sSaved}}}, proofLineno)

	// --- l2s_consts_d ---
	var constsDTerms []Expr
	if m != nil && m.Sig != nil {
		for _, s := range uninterpretedSorts {
			sName := s.String()
			if finiteSorts[sName] {
				continue
			}
			for _, sym := range insertionOrderSymbols(m) {
				if sym.CSort != nil && sym.CSort.String() == sName {
					d := L2SD(s)
					app, _ := NewApply(d, sym)
					if app != nil {
						constsDTerms = append(constsDTerms, app)
					}
				}
			}
		}
	}
	// M5 / Python ivy_l2s.py:634-638: always emit l2s_consts_d, even when
	// constsDTerms is empty (And() = true).
	invars = appendLF(autoAcfg, invars, "l2s_consts_d", checkMakeAnd(constsDTerms...), proofLineno)

	// --- convert_to_init: wrap temporal formula with l2s_init ---
	var iinvs []Expr
	var convertToInit func(f Expr) Expr
	convertToInit = func(f Expr) Expr {
		switch n := f.(type) {
		case *LogicAnd:
			terms := make([]Expr, len(n.Terms))
			for i, t := range n.Terms {
				terms[i] = convertToInit(t)
			}
			return &LogicAnd{Terms: terms}
		case *LogicOr:
			terms := make([]Expr, len(n.Terms))
			for i, t := range n.Terms {
				terms[i] = convertToInit(t)
			}
			return &LogicOr{Terms: terms}
		case *LogicNot:
			return &LogicNot{Body: convertToInit(n.Body)}
		case *LogicImplies:
			return &LogicImplies{T1: convertToInit(n.T1), T2: convertToInit(n.T2)}
		case *LogicIff:
			return &LogicIff{T1: convertToInit(n.T1), T2: convertToInit(n.T2)}
		case *ForAll:
			return &ForAll{Variables: n.Variables, Body: convertToInit(n.Body)}
		case *LogicExists:
			return &LogicExists{Variables: n.Variables, Body: convertToInit(n.Body)}
		default:
			vs := collectVarsSlice(f)
			initNB := l2sInit(vs, f, proofLabel)
			ini := applyNB(initNB, checkVarsToNodes(vs)...)
			// Same struct-eq dedup as the initGlobally walker above —
			// keyed by Sexp to match Python ivy_l2s.py:670-673
			// (add_ini_invar uses `cond not in known_inits` / .add).
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

	negPropInit := &LogicNot{Body: convertToInit(fmla)}
	for i, iinv := range iinvs {
		invars = appendLF(autoAcfg, invars, fmt.Sprintf("l2s_init_glob_%d", i), iinv, proofLineno)
	}
	invars = appendLF(autoAcfg, invars, "neg_prop_init", negPropInit, proofLineno)

	// C15 / Python ivy_l2s.py:664-676: l2s_when_<i> invariants for
	// WhenOperator nodes with name=="first" found in invars + property prems.
	{
		// Collect property premises from goal.
		var ntPrems []Expr
		for _, p := range GoalPrems(goal) {
			if lf, ok := p.(*LabeledFormula); ok && GoalIsProperty(lf) {
				if e, ok := lf.Formula.(Expr); ok {
					ntPrems = append(ntPrems, e)
				}
			}
		}
		// Walk all invars + property prems for temporal subnodes.
		var allFmlas []Expr
		for _, inv := range invars {
			if e, ok := inv.Formula.(Expr); ok {
				allFmlas = append(allFmlas, e)
			}
		}
		allFmlas = append(allFmlas, ntPrems...)
		// Collect LogicWhenOperator{Name:"first"} via TemporalsAst, dedup by Sexp.
		// Mirrors Python ivy_l2s.py:695 `iu.unique(ilu.temporals_asts(...))`
		// which uses Expr struct equality (recstruct __hash__/__eq__).
		// fmt.Sprint = String = PrettyFmla drops sort annotations.
		seen := make(map[string]bool)
		var winvs []Expr
		for _, f := range allFmlas {
			for _, t := range TemporalsAst(f) {
				when, ok := t.(*LogicWhenOperator)
				if !ok || when.Name != "first" {
					continue
				}
				key := string(when.Sexp())
				if seen[key] {
					continue
				}
				seen[key] = true
				// nws = LogicOr(Not(l2s_waiting), Not(l2s_w((), when.T2)))
				nws := &LogicOr{Terms: []Expr{
					&LogicNot{Body: l2sWaiting},
					&LogicNot{Body: applyNB(l2sW(nil, when.T2, proofLabel))},
				}}
				// tmp = Implies(Not(nws), Eq(when, LogicWhenOperator{Name:"next", T1:when.T1, T2:when.T2}))
				nextWhen := &LogicWhenOperator{Name: "next", T1: when.T1, T2: when.T2}
				inner := &LogicImplies{
					T1: &LogicNot{Body: nws},
					T2: &Eq{T1: when, T2: nextWhen},
				}
				// Wrap with Implies(l2s_init((), Eventually(when.T2)), inner)
				initNB := l2sInit(nil, &LogicEventually{Environ: strPtr(proofLabel), Body: when.T2}, proofLabel)
				outer := &LogicImplies{
					T1: applyNB(initNB),
					T2: inner,
				}
				winvs = append(winvs, outer)
			}
		}
		for i, w := range winvs {
			invars = appendLF(autoAcfg, invars, fmt.Sprintf("l2s_when_%d", i), w, proofLineno)
		}
	}

	// T1 / Python ivy_l2s.py:679-682: print l2s_auto invariants when in
	// debug mode (the Python code prints unconditionally; we gate behind
	// L2SDebug to avoid noise in normal runs).
	if m != nil && m.Cfg != nil && m.Cfg.L2SDebug {
		fmt.Printf("\n--- begin l2s_auto invariants ---\n\n")
		for _, inv := range invars {
			fmt.Printf("invariant %v\n", inv)
		}
		fmt.Println("\n--- end l2s_auto invariants ---")
	}

	return invars, tasks, triggers, nil
}

// appendLF appends a labeled formula to the invariant list.
// lineno mirrors Python's .sln(proof.lineno) on each LabeledFormula.
func appendLF(cfg *AstConfig, invars []*LabeledFormula, name string, fmla Expr, lineno Location) []*LabeledFormula {
	lf := cfg.NewLabeledFormula(cfg.NewAtom(name), fmla)
	lf.SetLineno(lineno)
	return append(invars, lf)
}

// buildOrExpr builds an Or expression from a slice of terms, mirroring
// Python's `lg.Or(*xs)`. Empty Or is false (per the codebase convention at
// l2s_auto.go:141). Single-element Or returns the element directly.
func buildOrExpr(xs []Expr) Expr {
	return &LogicOr{Terms: xs}
}

// isGloballyOrNotEventually returns true if e is *lg.Globally or *lg.Not{*lg.Eventually}.
// Used for the C16 EF-pattern extra invariant in initGlobally.
// Mirrors Python ivy_l2s.py:566.
func isGloballyOrNotEventually(e Expr) bool {
	if _, ok := e.(*LogicGlobally); ok {
		return true
	}
	if n, ok := e.(*LogicNot); ok {
		if _, ok := n.Body.(*LogicEventually); ok {
			return true
		}
	}
	return false
}

// isEventuallyOrNotGlobally returns true if e is *lg.Eventually or *lg.Not{*lg.Globally}.
// Used for the C16 AG-pattern extra invariant in initGlobally.
// Mirrors Python ivy_l2s.py:574.
func isEventuallyOrNotGlobally(e Expr) bool {
	if _, ok := e.(*LogicEventually); ok {
		return true
	}
	if n, ok := e.(*LogicNot); ok {
		if _, ok := n.Body.(*LogicGlobally); ok {
			return true
		}
	}
	return false
}

// sameLHSSort returns true if a and b have the same LHS sort signature
// (each LHS is either an *lg.Apply or *lg.Const). Used by H5 to validate
// that work_created/work_needed/work_done share a sort and that
// work_helpful/work_progress share a sort.
func sameLHSSort(a, b *Eq) bool {
	if a == nil || b == nil {
		return true
	}
	sortOf := func(e Expr) Sort {
		switch x := e.(type) {
		case *Apply:
			if c, ok := x.Func.(*Const); ok {
				return c.CSort
			}
		case *Const:
			return x.CSort
		}
		return nil
	}
	sa := sortOf(a.T1)
	sb := sortOf(b.T1)
	if sa == nil || sb == nil {
		return true
	}
	return sa.String() == sb.String()
}

// collectVarsSlice collects free variables from a node into a slice.
func collectVarsSlice(n Expr) []*LogicVariable {
	vars := VariablesAST(n)
	return vars
}

// cloneLHS creates a new LHS with a different symbol name.
func cloneLHS(lhs Expr, newName string) Expr {
	switch l := lhs.(type) {
	case *Apply:
		if c, ok := l.Func.(*Const); ok {
			newC := NewConst(newName, c.CSort)
			app, _ := NewApply(newC, l.Terms...)
			return app
		}
	case *Const:
		return NewConst(newName, l.CSort)
	}
	return lhs
}
