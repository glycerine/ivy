// l2s_auto.go implements the l2s_auto tactic's task/trigger invariant
// generation, ported from Python ivy_l2s.py lines 196-678.
package l2s

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	co "github.com/glycerine/ivy/goivy/clauseops"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	modpkg "github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
)

// l2sAutoInvariants generates invariants for l2s_auto tactics.
// It extracts task/trigger definitions from the proof goal premises
// and generates the appropriate L2S invariants.
func l2sAutoInvariants(
	tacticName string,
	goal *ast.LabeledFormula,
	invars []*ast.LabeledFormula,
	proofLabel string,
	fmla lg.Expr,
	finiteSorts map[string]bool,
	uninterpretedSorts []lg.Sort,
	m *modpkg.Module,
) ([]*ast.LabeledFormula, error) {
	if !strings.HasPrefix(tacticName, "l2s_auto") {
		return invars, nil
	}

	autoAcfg := m.Cfg.AstCfg

	// Helper: put into nested dict
	type defnMap map[string]map[string]*lg.Eq // sfx -> name -> definition (Eq node)
	tasks := make(defnMap)
	triggers := make(defnMap)

	dictPut := func(dct defnMap, sfx, name string, dfn *lg.Eq) {
		if _, ok := dct[sfx]; !ok {
			dct[sfx] = make(map[string]*lg.Eq)
		}
		dct[sfx][name] = dfn
	}

	// getAuxDefn extracts definitions like work_created, work_needed, etc.
	// from the goal premises.
	getAuxDefn := func(name string, dct defnMap) {
		prems := proof.GoalPrems(goal)
		for _, prem := range prems {
			// Try to get a definition from the premise
			var f lg.Expr
			switch p := prem.(type) {
			case *ast.LabeledFormula:
				if n, ok := p.Formula.(lg.Expr); ok {
					f = n
				}
			}
			if f == nil {
				continue
			}
			// Drop universals and check for Eq
			tmp := il.DropUniversals(f)
			eq, ok := tmp.(*lg.Eq)
			if !ok {
				continue
			}
			// Get the defined name
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
			// Check no free variables
			freeVars := co.VariablesAST(f)
			if len(freeVars) > 0 {
				continue // skip definitions with free variables
			}
			sfx := dname[len(name):]
			dictPut(dct, sfx, name, eq)
		}
	}

	getAuxDefn("work_created", tasks)
	getAuxDefn("work_needed", tasks)
	getAuxDefn("work_done", tasks)
	getAuxDefn("work_progress", tasks)
	getAuxDefn("work_end", tasks)
	getAuxDefn("work_invar", tasks)
	if tacticName == "l2s_auto5" {
		getAuxDefn("work_helpful", tasks)
	}
	getAuxDefn("work_start", triggers)

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
				workDone := &lg.Eq{
					T1: cloneLHS(workNeeded.T1, "work_done"+sfx),
					T2: &lg.Or{Terms: nil}, // false = empty disjunction
				}
				dictPut(tasks, sfx, "work_done", workDone)
			}
		}
	}

	// Validate required definitions
	for _, sfx := range sortedTasks {
		for _, name := range []string{"work_created", "work_needed", "work_done", "work_progress"} {
			if tasks[sfx][name] == nil {
				return nil, fmt.Errorf("tactic l2s_auto requires a definition of %s%s", name, sfx)
			}
		}
	}

	// Monitor symbols
	l2sWaiting := L2SWaiting()
	l2sSaved := L2SSaved()

	// Helpers
	forall := func(vs []*lg.Variable, body lg.Expr) lg.Expr {
		if len(vs) == 0 {
			return body
		}
		return &lg.ForAll{Variables: vs, Body: body}
	}
	exists := func(vs []*lg.Variable, body lg.Expr) lg.Expr {
		if len(vs) == 0 {
			return body
		}
		return &lg.Exists{Variables: vs, Body: body}
	}

	// Defn helpers
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
				// Key by Symbol with same name/sort so SubstituteConstantsAST matches
				sym := lg.NewConst(v.Name, v.VSort)
				m[lg.Key(sym)] = dst[i]
			}
		}
		return m
	}
	subst := func(node lg.Expr, subs map[lg.NodeKey]lg.Expr) lg.Expr {
		return co.SubstituteConstantsExpr(node, subs)
	}

	// all_d: all elements in l2s_d
	allD := func(eq *lg.Eq) lg.Expr {
		args := eqLHSArgs(eq)
		var cons []lg.Expr
		for _, v := range args {
			if v.VSort != nil && !finiteSorts[v.VSort.String()] {
				d := L2SD(v.VSort)
				app, _ := lg.NewApply(d, v)
				if app != nil {
					cons = append(cons, app)
				}
			}
		}
		rhs := eqRHS(eq)
		if len(cons) == 0 {
			return &lg.And{Terms: nil} // true
		}
		return &lg.Implies{T1: rhs, T2: makeAnd(cons...)}
	}

	// all_a: all elements in l2s_a
	allA := func(eq *lg.Eq) lg.Expr {
		args := eqLHSArgs(eq)
		var cons []lg.Expr
		for _, v := range args {
			if v.VSort != nil && !finiteSorts[v.VSort.String()] {
				a := L2SA(v.VSort)
				app, _ := lg.NewApply(a, v)
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

	// all_created: needed elements are in created set
	allCreated := func(defnNeeded *lg.Eq, sfxIdx int) lg.Expr {
		sfx := sortedTasks[sfxIdx]
		workCreated := tasks[sfx]["work_created"]
		createdArgs := eqLHSArgs(workCreated)
		defnArgs := eqLHSArgs(defnNeeded)
		s := substVars(defnArgs, createdArgs)
		return &lg.Implies{T1: subst(eqRHS(defnNeeded), s), T2: eqRHS(workCreated)}
	}

	// eventuallyStartTask
	eventuallyStartTask := func(workStart *lg.Eq) lg.Expr {
		if workStart == nil {
			return &lg.And{Terms: nil} // true
		}
		trigRHS := eqRHS(workStart)
		evf := &lg.Eventually{Body: trigRHS}
		vs := eqLHSArgs(workStart)
		vsNodes := varsToNodes(vs)
		initNB := l2sInit(vs, evf, proofLabel)
		return forall(vs, applyNB(initNB, vsNodes...))
	}

	// Generate invariants per task
	for idx, sfx := range sortedTasks {
		task := tasks[sfx]
		workCreated := task["work_created"]
		workNeeded := task["work_needed"]
		workDone := task["work_done"]
		workProgress := task["work_progress"]
		workEnd := task["work_end"]
		workInvar := task["work_invar"]
		var workStart *lg.Eq
		if trig, ok := triggers[sfx]; ok {
			workStart = trig["work_start"]
		}

		doneArgs := eqLHSArgs(workDone)

		// notWaitingForStart
		var notWaitingForStart lg.Expr = &lg.And{Terms: nil} // true
		if workStart != nil {
			evStart := eventuallyStartTask(workStart)
			trigRHS := eqRHS(workStart)
			wBinder := l2sW(nil, trigRHS, proofLabel)
			notWaitingForStart = makeAnd(evStart,
				&lg.Or{Terms: []lg.Expr{
					&lg.Not{Body: l2sWaiting},
					&lg.Not{Body: wBinder},
				}})
		}

		// --- l2s_needed_when_start ---
		if tacticName != "l2s_auto5" {
			tmp := &lg.Implies{T1: notWaitingForStart, T2: allCreated(workNeeded, idx)}
			invars = appendLF(autoAcfg, invars, "l2s_needed_when_start"+sfx, tmp)
		} else {
			tmp := &lg.Implies{T1: notWaitingForStart, T2: allD(workNeeded)}
			invars = appendLF(autoAcfg, invars, "l2s_needed_when_start"+sfx, tmp)
		}

		// --- l2s_created ---
		invars = appendLF(autoAcfg, invars, "l2s_created"+sfx, allD(workCreated))

		// --- l2s_needed_are_frozen ---
		evStart := eventuallyStartTask(workStart)
		if tacticName != "l2s_auto4" && tacticName != "l2s_auto5" {
			tmp := &lg.Implies{
				T1: makeAnd(evStart, &lg.Not{Body: l2sWaiting}),
				T2: allA(workNeeded),
			}
			invars = appendLF(autoAcfg, invars, "l2s_needed_are_frozen"+sfx, tmp)
		} else {
			doneSubArgs := eqLHSArgs(workDone)
			neededArgs := eqLHSArgs(workNeeded)
			s := substVars(neededArgs, doneSubArgs)
			isDone := &lg.Implies{T1: subst(eqRHS(workNeeded), s), T2: eqRHS(workDone)}
			notIsDone := &lg.Not{Body: isDone}
			var aCons []lg.Expr
			for _, v := range doneSubArgs {
				if v.VSort != nil && !finiteSorts[v.VSort.String()] {
					a := L2SA(v.VSort)
					app, _ := lg.NewApply(a, v)
					if app != nil {
						aCons = append(aCons, app)
					}
				}
			}
			tmp := &lg.Implies{
				T1: makeAnd(evStart, &lg.Not{Body: l2sWaiting}),
				T2: &lg.Implies{T1: notIsDone, T2: makeAnd(aCons...)},
			}
			invars = appendLF(autoAcfg, invars, "l2s_needed_are_frozen"+sfx, tmp)
		}

		// --- l2s_done_implies_created ---
		if tacticName != "l2s_auto3" && tacticName != "l2s_auto4" && tacticName != "l2s_auto5" {
			createdArgs := eqLHSArgs(workCreated)
			s := substVars(doneArgs, createdArgs)
			tmp := &lg.Implies{T1: subst(eqRHS(workDone), s), T2: eqRHS(workCreated)}
			invars = appendLF(autoAcfg, invars, "l2s_done_implies_created"+sfx, tmp)
		}

		// --- l2s_needed_implies_created ---
		if tacticName != "l2s_auto5" {
			createdArgs := eqLHSArgs(workCreated)
			neededArgs := eqLHSArgs(workNeeded)
			s := substVars(neededArgs, createdArgs)
			tmp := &lg.Implies{
				T1: notWaitingForStart,
				T2: &lg.Implies{T1: subst(eqRHS(workNeeded), s), T2: eqRHS(workCreated)},
			}
			invars = appendLF(autoAcfg, invars, "l2s_needed_implies_created"+sfx, tmp)
		}

		// --- l2s_work_preserved ---
		var wasDone, isDoneNode lg.Expr
		if tacticName != "l2s_auto4" && tacticName != "l2s_auto5" {
			sNB := l2sS(doneArgs, eqRHS(workDone), proofLabel)
			wasDone = applyNB(sNB, varsToNodes(doneArgs)...)
			isDoneNode = eqRHS(workDone)
		} else {
			neededArgs := eqLHSArgs(workNeeded)
			s := substVars(neededArgs, doneArgs)
			neededSubbed := subst(eqRHS(workNeeded), s)
			isDoneImplied := &lg.Implies{T1: neededSubbed, T2: eqRHS(workDone)}

			sNB := l2sS(doneArgs, isDoneImplied, proofLabel)
			wasDone = applyNB(sNB, varsToNodes(doneArgs)...)
			isDoneNode = isDoneImplied
		}
		tmp := &lg.Implies{T1: makeAnd(l2sSaved, wasDone), T2: isDoneNode}
		if tacticName == "l2s_auto5" {
			tmp = &lg.Implies{T1: evStart, T2: tmp}
		}
		invars = appendLF(autoAcfg, invars, "l2s_work_preserved"+sfx, tmp)

		// --- l2s_progress_made ---
		progressArgs := eqLHSArgs(workProgress)
		waitingForProgress := l2sW(progressArgs, eqRHS(workProgress), proofLabel)

		if tacticName != "l2s_auto3" {
			var progressInv lg.Expr
			if len(progressArgs) > 0 || len(tasks) > 1 {
				innerCond := makeAnd(
					l2sSaved,
					&lg.Not{Body: applyNB(waitingForProgress, varsToNodes(progressArgs)...)},
				)
				progressBody := exists(doneArgs,
					makeAnd(&lg.Not{Body: wasDone}, isDoneNode))
				progressInv = forall(progressArgs,
					&lg.Implies{T1: innerCond, T2: progressBody})
			} else {
				progressInv = &lg.Implies{
					T1: makeAnd(l2sSaved, &lg.Not{Body: waitingForProgress}),
					T2: exists(doneArgs, makeAnd(&lg.Not{Body: wasDone}, isDoneNode)),
				}
			}
			invars = appendLF(autoAcfg, invars, "l2s_progress_made"+sfx, progressInv)
		}

		// --- l2s_progress_invar ---
		gBody := &lg.Globally{Body: &lg.Eventually{Body: eqRHS(workProgress)}}
		initNB := l2sInit(progressArgs, gBody, proofLabel)
		invars = appendLF(autoAcfg, invars, "l2s_progress_invar"+sfx,
			&lg.Implies{
				T1: applyNB(initNB, varsToNodes(progressArgs)...),
				T2: gBody,
			})

		// --- l2s_not_all_done ---
		neededArgs := eqLHSArgs(workNeeded)
		s := substVars(neededArgs, doneArgs)
		var notAllDoneBody lg.Expr = &lg.Implies{T1: subst(eqRHS(workNeeded), s), T2: eqRHS(workDone)}
		if workEnd != nil {
			endArgs := eqLHSArgs(workEnd)
			endSubs := substVars(endArgs, doneArgs)
			notAllDoneBody = &lg.Implies{T1: subst(eqRHS(workEnd), endSubs), T2: notAllDoneBody}
		}
		if workInvar != nil {
			notAllDoneBody = &lg.Implies{T1: eqRHS(workInvar), T2: notAllDoneBody}
		}
		notAllDone := &lg.Not{Body: forall(doneArgs, notAllDoneBody)}

		if idx == len(sortedTasks)-1 {
			invars = appendLF(autoAcfg, invars, "l2s_not_all_done", notAllDone)
		}
	}

	// --- init_globally: generate l2s_globally invariants ---
	knownInits := make(map[string]bool)
	var ninvs []lg.Expr

	var initGlobally func(prop lg.Expr, res *[]lg.Expr, pos bool)
	initGlobally = func(prop lg.Expr, res *[]lg.Expr, pos bool) {
		switch p := prop.(type) {
		case *lg.Globally:
			knownInits[fmt.Sprint(prop)] = true
			if pos {
				*res = append(*res, prop)
				initGlobally(p.Body, res, pos)
			} else {
				// ~pos, Globally: treat as Eventually(~body)
				arg := p.Body
				vs := collectVarsSlice(arg)
				wNB := l2sW(vs, &lg.Not{Body: arg}, proofLabel)
				notWaiting := &lg.Or{Terms: []lg.Expr{
					&lg.Not{Body: l2sWaiting},
					&lg.Not{Body: applyNB(wNB, varsToNodes(vs)...)},
				}}
				*res = append(*res, &lg.Implies{T1: prop, T2: notWaiting})
				initNB := l2sInit(vs, &lg.Not{Body: prop}, proofLabel)
				*res = append(*res, applyNB(initNB, varsToNodes(vs)...))
			}
		case *lg.Eventually:
			knownInits[fmt.Sprint(prop)] = true
			if !pos {
				*res = append(*res, &lg.Not{Body: prop})
				initGlobally(p.Body, res, pos)
			} else {
				arg := p.Body
				vs := collectVarsSlice(arg)
				wNB := l2sW(vs, arg, proofLabel)
				notWaiting := &lg.Or{Terms: []lg.Expr{
					&lg.Not{Body: l2sWaiting},
					&lg.Not{Body: applyNB(wNB, varsToNodes(vs)...)},
				}}
				*res = append(*res, &lg.Implies{T1: &lg.Not{Body: prop}, T2: notWaiting})
				initNB := l2sInit(vs, prop, proofLabel)
				*res = append(*res, applyNB(initNB, varsToNodes(vs)...))
			}
		case *lg.Implies:
			if !pos {
				initGlobally(p.T1, res, !pos)
				initGlobally(p.T2, res, pos)
			}
		case *lg.And:
			if pos {
				for _, arg := range p.Terms {
					initGlobally(arg, res, pos)
				}
			}
		case *lg.Or:
			if !pos {
				for _, arg := range p.Terms {
					initGlobally(arg, res, pos)
				}
			}
		case *lg.ForAll:
			if pos {
				initGlobally(p.Body, res, pos)
			}
		case *lg.Exists:
			if !pos {
				initGlobally(p.Body, res, pos)
			}
		case *lg.Not:
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
		evf := &lg.Eventually{Body: arg}
		vs := eqLHSArgs(trigDef)
		vsNodes := varsToNodes(vs)
		initNB := l2sInit(vs, evf, proofLabel)
		initF := applyNB(initNB, vsNodes...)
		ninvs = append(ninvs, &lg.Or{Terms: []lg.Expr{initF, &lg.Not{Body: evf}}})
		wNB := l2sW(vs, arg, proofLabel)
		notWaiting := &lg.Or{Terms: []lg.Expr{
			&lg.Not{Body: l2sWaiting},
			&lg.Not{Body: applyNB(wNB, vsNodes...)},
		}}
		ninvs = append(ninvs, &lg.Implies{
			T1: makeAnd(initF, &lg.Not{Body: evf}),
			T2: notWaiting,
		})
		var tinvs []lg.Expr
		initGlobally(arg, &tinvs, true)
		for _, tinv := range tinvs {
			ninvs = append(ninvs, &lg.Implies{
				T1: makeAnd(initF, notWaiting),
				T2: tinv,
			})
		}
	}

	for i, ninv := range ninvs {
		invars = appendLF(autoAcfg, invars, fmt.Sprintf("l2s_globally_%d", i), ninv)
	}

	// --- convert_to_init: wrap temporal formula with l2s_init ---
	var iinvs []lg.Expr
	var convertToInit func(f lg.Expr) lg.Expr
	convertToInit = func(f lg.Expr) lg.Expr {
		switch n := f.(type) {
		case *lg.And:
			terms := make([]lg.Expr, len(n.Terms))
			for i, t := range n.Terms {
				terms[i] = convertToInit(t)
			}
			return &lg.And{Terms: terms}
		case *lg.Or:
			terms := make([]lg.Expr, len(n.Terms))
			for i, t := range n.Terms {
				terms[i] = convertToInit(t)
			}
			return &lg.Or{Terms: terms}
		case *lg.Not:
			return &lg.Not{Body: convertToInit(n.Body)}
		case *lg.Implies:
			return &lg.Implies{T1: convertToInit(n.T1), T2: convertToInit(n.T2)}
		case *lg.Iff:
			return &lg.Iff{T1: convertToInit(n.T1), T2: convertToInit(n.T2)}
		case *lg.ForAll:
			return &lg.ForAll{Variables: n.Variables, Body: convertToInit(n.Body)}
		case *lg.Exists:
			return &lg.Exists{Variables: n.Variables, Body: convertToInit(n.Body)}
		default:
			vs := collectVarsSlice(f)
			initNB := l2sInit(vs, f, proofLabel)
			ini := applyNB(initNB, varsToNodes(vs)...)
			key := fmt.Sprint(f)
			if _, ok := f.(*lg.Globally); ok && !knownInits[key] {
				iinvs = append(iinvs, &lg.Implies{T1: ini, T2: f})
				knownInits[key] = true
			}
			if _, ok := f.(*lg.Eventually); ok && !knownInits[key] {
				iinvs = append(iinvs, &lg.Implies{T1: f, T2: ini})
				knownInits[key] = true
			}
			return ini
		}
	}

	negPropInit := &lg.Not{Body: convertToInit(fmla)}
	for i, iinv := range iinvs {
		invars = appendLF(autoAcfg, invars, fmt.Sprintf("l2s_init_glob_%d", i), iinv)
	}
	invars = appendLF(autoAcfg, invars, "neg_prop_init", negPropInit)

	// --- l2s_status invariants ---
	invars = appendLF(autoAcfg, invars, "l2s_status_0",
		&lg.Or{Terms: []lg.Expr{l2sWaiting, L2SFrozen(), l2sSaved}})
	invars = appendLF(autoAcfg, invars, "l2s_status_1",
		&lg.Or{Terms: []lg.Expr{&lg.Not{Body: l2sWaiting}, &lg.Not{Body: L2SFrozen()}}})
	invars = appendLF(autoAcfg, invars, "l2s_status_2",
		&lg.Or{Terms: []lg.Expr{&lg.Not{Body: l2sWaiting}, &lg.Not{Body: l2sSaved}}})
	invars = appendLF(autoAcfg, invars, "l2s_status_3",
		&lg.Or{Terms: []lg.Expr{&lg.Not{Body: L2SFrozen()}, &lg.Not{Body: l2sSaved}}})

	// --- l2s_consts_d ---
	var constsDTerms []lg.Expr
	if m != nil && m.Sig != nil {
		for _, s := range uninterpretedSorts {
			sName := s.String()
			if finiteSorts[sName] {
				continue
			}
			for _, sym := range m.Sig.Symbols {
				if sym.Sort != nil && sym.Sort.String() == sName {
					d := L2SD(s)
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
		invars = appendLF(autoAcfg, invars, "l2s_consts_d", makeAnd(constsDTerms...))
	}

	return invars, nil
}

// appendLF appends a labeled formula to the invariant list.
func appendLF(cfg *ast.AstConfig, invars []*ast.LabeledFormula, name string, fmla lg.Expr) []*ast.LabeledFormula {
	lf := cfg.NewLabeledFormula(lg.NewConst(name, &lg.BooleanSort{}), fmla)
	return append(invars, lf)
}

// collectVarsSlice collects free variables from a node into a slice.
func collectVarsSlice(n lg.Expr) []*lg.Variable {
	vars := co.VariablesAST(n)
	return vars
}

// cloneLHS creates a new LHS with a different symbol name.
func cloneLHS(lhs lg.Expr, newName string) lg.Expr {
	switch l := lhs.(type) {
	case *lg.Apply:
		if c, ok := l.Func.(*lg.Const); ok {
			newC := lg.NewConst(newName, c.CSort)
			app, _ := lg.NewApply(newC, l.Terms...)
			return app
		}
	case *lg.Const:
		return lg.NewConst(newName, l.CSort)
	}
	return lhs
}
