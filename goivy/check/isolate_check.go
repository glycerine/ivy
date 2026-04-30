package check

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/glycerine/ivy/goivy/acl"
	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/art"
	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/bmc"
	"github.com/glycerine/ivy/goivy/compiler"
	"github.com/glycerine/ivy/goivy/fragment"
	ivyiso "github.com/glycerine/ivy/goivy/isolate"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/mc"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/temporal"
	"github.com/glycerine/ivy/goivy/vmt"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// CheckIsolate is the main isolate checking function.
// It orchestrates verification of properties, conjectures, and guarantees
// within a single isolate. This corresponds to Python's check_isolate.
//
// The traceHook parameter, if non-nil, is called to process trace output.
func CheckIsolate(mod *module.Module, traceHook func(interface{}) interface{}) error {
	xtracer.Trace("check.CheckIsolate ENTER")
	// Print axioms
	fmt.Println("\n    The following properties are treated as axioms: ")
	for _, lf := range mod.LabeledAxioms {
		fmt.Println(PrettyLF(lf, 8))
	}

	// Print properties
	fmt.Println("\n    The following properties are treated as properties:")
	for _, lf := range mod.LabeledProps {
		fmt.Println(PrettyLF(lf, 8))
	}

	// Print conjectures
	fmt.Println("\n    The following properties are treated as conjectures: ")
	for _, lf := range mod.LabeledConjs {
		fmt.Println(PrettyLF(lf, 8))
	}

	// If there is an isolate proof, handle it via proof checker.
	// Python (ivy_check.py:506-516):
	//   pc = ivy_proof.ProofChecker(mod.labeled_axioms+mod.assumed_invariants, mod.definitions, mod.schemata)
	//   model = itmp.normal_program_from_module(im.module)
	//   prop = ivy_ast.LabeledFormula(ivy_ast.Atom('safety'), lg.And())
	//   subgoal = ivy_ast.LabeledFormula(ivy_ast.Atom('safety'), ivy_ast.TemporalModels(model, lg.And()))
	//   subgoal.lineno = mod.isolate_proof.lineno
	//   subgoals = pc.admit_proposition(prop, mod.isolate_proof, subgoals)
	//   check_subgoals(subgoals)
	if mod.IsolateProof != nil {
		pcAxioms := make([]*ast.LabeledFormula, 0, len(mod.LabeledAxioms)+len(mod.AssumedInvs))
		pcAxioms = append(pcAxioms, mod.LabeledAxioms...)
		pcAxioms = append(pcAxioms, mod.AssumedInvs...)
		pc := proof.NewProofChecker(mod.Cfg.ProofCfg, mod, pcAxioms, mod.Definitions, ModuleSchemataToAst(mod.Schemata))

		model := temporal.NormalProgramFromModule(mod)
		acfg := mod.Cfg.AstCfg
		safetyLabel := acfg.NewAtom("safety")
		prop := acfg.NewLabeledFormula(safetyLabel, &lg.And{})

		tm := acfg.NewTemporalModels(model, &lg.And{})
		subgoal := acfg.NewLabeledFormula(safetyLabel, tm)
		// Python: subgoal.lineno = mod.isolate_proof.lineno
		if pfNode, ok := mod.IsolateProof.(ast.Node); ok {
			subgoal.SetLineno(pfNode.GetLineno())
		}

		subgoals := []*ast.LabeledFormula{subgoal}
		pfNode, _ := mod.IsolateProof.(ast.Node)
		var err error
		subgoals, err = pc.AdmitProposition(prop, pfNode, subgoals...)
		if err != nil {
			return err
		}
		return CheckSubgoals(subgoals, nil, mod)
	}

	fcStartTm := time.Now()

	// Python: ifc.check_fragment()
	xtracer.Trace("check/isolate_check.go: CheckIsolate about to call fragment.CheckFragment(mod, false)")
	if err := fragment.CheckFragment(mod, false); err != nil {
		xtracer.Trace("check/isolate_check.go: CheckIsolate back from fragment.CheckFragment() but with error: %v", err)
		return err
	}
	xtracer.Trace("check/isolate_check.go: CheckIsolate back from fragment.CheckFragment().")

	fcElap := time.Since(fcStartTm)
	if mod.Cfg.OptIvyStats {
		fmt.Printf("\n\t IVY_STATS fragment checker elapsed time (s): %v\n", fcElap)
	}

	// Python: with im.module.theory_context():
	cleanupTheory := mod.TheoryContext()
	defer cleanupTheory()

	check := !mod.Cfg.OptSummary

	// Build subgoal map
	subgoalMap := make(map[int64]bool)
	for _, sg := range mod.Subgoals {
		subgoalMap[sg.Formula.ID] = true
	}

	// Print and check properties
	axioms := make([]*ast.LabeledFormula, 0)
	schemaInstances := make([]*ast.LabeledFormula, 0)
	for _, m := range mod.LabeledAxioms {
		if !subgoalMap[m.ID] {
			axioms = append(axioms, m)
		} else {
			schemaInstances = append(schemaInstances, m)
		}
	}

	if len(axioms) > 0 {
		fmt.Println("\n    The following properties are assumed as axioms:")
		for _, lf := range axioms {
			fmt.Println(PrettyLF(lf, 8))
		}
	}

	if len(mod.Definitions) > 0 {
		fmt.Println("\n    The following definitions are used:")
		for _, lf := range mod.Definitions {
			fmt.Println(PrettyLF(lf, 8))
		}
	}

	// Python: if (mod.labeled_props or schema_instances) and not checked_action.get() and not unprovable:
	//   print header always, then if check: do actual checking, else: print summaries
	if (len(mod.LabeledProps) > 0 || len(schemaInstances) > 0) &&
		mod.Cfg.CheckedAction == "" && !mod.Cfg.OnlyCheckUnprovable {
		fmt.Println("\n    The following properties are to be checked:")
		if check {
			for _, lf := range schemaInstances {
				fmt.Println(PrettyLF(lf, 8) + " [proved by axiom schema]")
			}
			// Property checking: create AG with True pre-state, check properties
			ag := art.NewAnalysisGraph(mod)
			pre := art.NewState(mod, module.TrueClauses(actions.EmptyAnnotation{}))
			ag.Add(pre, nil)

			nonTemporal := make([]*ast.LabeledFormula, 0)
			for _, p := range mod.LabeledProps {
				if !p.IsTemporal() {
					nonTemporal = append(nonTemporal, p)
				}
			}
			// Filter out explicitly proved subgoals
			nonTemporal = filterExplicitSubgoals(nonTemporal, subgoalMap)
			var checkers []Checker
			for _, prop := range nonTemporal {
				if prop.Assumed || subgoalMap[prop.ID] {
					checkers = append(checkers, NewConjAssumer(mod, prop))
				} else {
					checkers = append(checkers, NewConjChecker(mod, prop, 8))
				}
			}
			// Python ivy_check.py:572 calls check_fcs_in_state unconditionally,
			// even when fcs is empty.
			CheckFcsInStateWithAG(mod, ag, pre, checkers)
		} else {
			// Python: else: for lf in schema_instances + mod.labeled_props: print(pretty_lf(lf))
			for _, lf := range schemaInstances {
				fmt.Println(PrettyLF(lf, 8))
			}
			for _, lf := range mod.LabeledProps {
				fmt.Println(PrettyLF(lf, 8))
			}
		}
	}

	// After checking properties, make non-temporal ones axioms
	// Python: im.module.labeled_axioms.extend(p for p in im.module.labeled_props if not p.temporal)
	for _, p := range mod.LabeledProps {
		if !p.IsTemporal() {
			mod.LabeledAxioms = append(mod.LabeledAxioms, p)
		}
	}
	// Python: im.module.update_theory()
	mod.UpdateTheory()

	// Print initial properties
	if len(mod.LabeledInits) > 0 {
		fmt.Println("\n    The following properties are assumed initially:")
		for _, lf := range mod.LabeledInits {
			fmt.Println(PrettyLF(lf, 8))
		}
	}

	// Print checked invariants
	// Python: checked_invariants = [x for x in mod.labeled_conjs if is_check_mod_unprovable(x)]
	var checkedInvariants []*ast.LabeledFormula
	for _, c := range mod.LabeledConjs {
		if IsCheckModUnprovable(mod.Cfg, c) {
			checkedInvariants = append(checkedInvariants, c)
		}
	}
	if len(checkedInvariants) > 0 {
		fmt.Println("\n    The inductive invariant consists of the following conjectures:")
		for _, lf := range checkedInvariants {
			fmt.Println(PrettyLF(lf, 8))
		}
	}

	// Apply conjecture proofs
	ApplyConjProofs(mod)

	// Print isolate implementations
	// Python: "{}implementation of {}".format(pretty_lineno(action), mixee)
	if mod.IsolateInfo != nil && len(mod.IsolateInfo.Implementations) > 0 {
		fmt.Println("\n    The following action implementations are present:")
		impls := make([]module.MixinTriple, len(mod.IsolateInfo.Implementations))
		copy(impls, mod.IsolateInfo.Implementations)
		sort.Slice(impls, func(i, j int) bool { return impls[i].Mixer < impls[j].Mixer })
		for _, impl := range impls {
			fmt.Printf("        %simplementation of %s\n", PrettyActionLineno(impl.Action), impl.Mixee)
		}
	}

	// Print monitors
	// Python: "{}monitor of {}".format(pretty_lineno(action), mixee)
	if mod.IsolateInfo != nil && len(mod.IsolateInfo.Monitors) > 0 {
		fmt.Println("\n    The following action monitors are present:")
		mons := make([]module.MixinTriple, len(mod.IsolateInfo.Monitors))
		copy(mons, mod.IsolateInfo.Monitors)
		sort.Slice(mons, func(i, j int) bool { return mons[i].Mixer < mons[j].Mixer })
		for _, mon := range mons {
			fmt.Printf("        %smonitor of %s\n", PrettyActionLineno(mon.Action), mon.Mixee)
		}
	}

	// Print initializers
	// Python: "{}{}".format(pretty_lineno(action), actname)
	if len(mod.Initializers) > 0 {
		fmt.Println("\n    The following initializers are present:")
		inits := make([]module.NamedAction, len(mod.Initializers))
		copy(inits, mod.Initializers)
		sort.Slice(inits, func(i, j int) bool { return inits[i].Name < inits[j].Name })
		for _, na := range inits {
			fmt.Printf("initializer:        %s %s\n", PrettyActionLineno(na.Action), na.Name)
		}
	}

	// Check initialization establishes invariant.
	// Python: ag = ivy_art.AnalysisGraph(initializer=lambda x:None)
	//         check_conjs_in_state(mod, ag, ag.states[0])
	// The initializer=lambda x:None means "use init_cond, no abstraction."
	// AddInitialState computes init state from mod.InitCond + initializer actions.
	// Python: if checked_invariants and not checked_action.get() and not unprovable:
	//   print header always, then if check: do checking, else: print('')
	if len(checkedInvariants) > 0 && mod.Cfg.CheckedAction == "" && !mod.Cfg.OnlyCheckUnprovable {
		fmt.Println("\n    Initialization must establish the invariant")
		if check {
			ag := art.NewAnalysisGraph(mod)
			ag.Initialize(art.AbstractorFunc(func(s *art.State) {})) // no-op abstractor, matching Python
			if len(ag.States) > 0 {
				CheckConjsInStateWithAG(mod, ag, ag.States[0], 8, nil)
			}
		} else {
			fmt.Println("")
		}
	}

	// Check initializer assertions
	// Python: guarantees = [sub for sub in action.iter_subactions()
	//           if isinstance(sub, (act.AssertAction, act.Ranking))
	//           for action in mod.initializers]
	//
	// code review comments:
	// **A2. Python `check_isolate` initializer guarantee list comprehension
	// is a Python bug; Go behavior diverges**
	//
	// Python `ivy_check.py` lines ~635-645:
	// ```python
	// if mod.initializers:
	//     guarantees = [sub for sub in action.iter_subactions()
	//                   if isinstance(sub,(act.AssertAction,act.Ranking))
	//                   for action in mod.initializers]
	// ```
	// In Python 3, `action.iter_subactions()` at the START of the
	// comprehension uses `action` from the **outer scope** (the last
	// `env_action` from the preceding checked-actions loop), NOT from
	// `for action in mod.initializers`. This is a Python 3 scoping bug
	// in the original — the comprehension iterates the initializerslist
	// but applies `iter_subactions()` on the WRONG action.
	//
	// Go's `isolate_check.go:284-358` correctly iterates each initializer
	// action separately. **Go is more correct than Python here, but
	// it diverges.** The test oracle should know about this.
	//
	// *Files:* `isolate_check.go:310-324
	// *Severity:* MEDIUM — Go is BETTER than Python, but produces different output/behavior

	if len(mod.Initializers) > 0 {
		var guarantees []actions.Action
		for _, na := range mod.Initializers {
			if act, ok := na.Action.(actions.Action); ok {
				for _, sub := range act.IterSubactions() {
					isAssert := actions.IsAssertLike(sub)
					_, isRanking := sub.(*actions.Ranking)
					if isAssert || isRanking {
						if IsGuaranteeModUnprovable(mod.Cfg, sub) {
							guarantees = append(guarantees, sub)
						}
					}
				}
			}
		}
		// Python: if check_lineno is not None: guarantees = [sub for sub in guarantees if sub.lineno == check_lineno]
		if mod.Cfg.CheckLineno != "" {
			var filtered []actions.Action
			for _, sub := range guarantees {
				if sub.GetLineno().FileLineKey() == mod.Cfg.CheckLineno {
					filtered = append(filtered, sub)
				}
			}
			guarantees = filtered
		}
		// Python: guarantees = [x for x in guarantees if is_guarantee_mod_unprovable(x)]
		// (already filtered above)
		// Python: if guarantees and not unprovable: print header always; if check: verify
		if len(guarantees) > 0 && !mod.Cfg.OnlyCheckUnprovable {
			fmt.Print("\n    Any assertions in initializers must be checked ")
			if check {
				ag := art.NewAnalysisGraph(mod)
				ag.Initialize(art.AbstractorFunc(func(s *art.State) {})) // no-op abstractor
				if len(ag.States) > 0 {
					initState := ag.States[0]

					failState := art.NewState(mod, module.TrueClauses(actions.EmptyAnnotation{}))

					var innerAction actions.Action
					if aa, ok := initState.Prov.(*art.ActionApp); ok {
						switch rep := aa.Rep.(type) {
						case string:
							failState.Prov = art.NewActionApp("fail_"+rep, aa.Args...)
							if v, ok := mod.Actions.Get2(rep); ok {
								if a, ok := v.(actions.Action); ok {
									innerAction = a
								}
							}
						case actions.Action:
							failState.Prov = art.NewActionApp("fail_"+rep.Name(), aa.Args...)
							innerAction = rep
						}
					}
					if innerAction == nil {
						innerAction = initState.Action
					}

					if innerAction != nil {
						failState.Action = actions.NewFailAction(innerAction)
						if initState.Pred != nil {
							failState.Pred = initState.Pred
						} else if aa, ok := initState.Prov.(*art.ActionApp); ok && len(aa.Args) > 0 {
							failState.Pred = aa.Args[0]
						}
					}

					CheckSafetyInStateWithAG(mod, ag, failState, true)
				}
			}
		}
	}

	// Check that external actions preserve the invariant
	checkedActions := setFromSlice(GetCheckedActions(mod))
	prioritized := setFromSlice(GetPrioritizedActions(mod.Cfg))
	prioritizedChecked := intersect(prioritized, checkedActions)
	actionOrder := sortedUnion(prioritizedChecked, checkedActions)

	if len(checkedActions) > 0 && len(checkedInvariants) > 0 {
		fmt.Println("\n    The following set of external actions must preserve the invariant:")
		if mod.Cfg.PriorityActions != "" {
			var plist []string
			for k := range prioritizedChecked {
				plist = append(plist, k)
			}
			sort.Strings(plist)
			fmt.Printf("\n         (prioritized checking order: %v)\n", plist)
		}
		for _, actname := range actionOrder {
			// Build the env_action for this external action
			action := actions.BuildEnvAction(mod.Cfg.ActCfg, mod.PublicActions, mod.Actions, actname, "")
			// Python: print("        {}{}".format(pretty_lineno(action), actname))
			actionLineno := "(internal) "
			if action != nil {
				if loc := action.GetLineno(); loc.Line > 0 {
					actionLineno = fmt.Sprintf("%d", loc.Line)
				}
			}
			fmt.Printf("        %s%s\n", actionLineno, actname)
			if check {
				// Python: ag = ivy_art.AnalysisGraph()
				//         pre = itp.State()
				//         pre.clauses = get_conjs(mod)
				//         with EvalContext(check=False):
				//             post = ag.execute(action, pre)
				//         check_conjs_in_state(mod, ag, post, indent=12, pcs=...)
				ag := art.NewAnalysisGraph(mod)
				pre := art.NewState(mod, GetConjs(mod))
				ag.Add(pre, nil)
				// Execute action to get post-state
				// Python uses EvalContext(check=False) here, so checkPrecond=false.
				post, err := ag.Execute(false, action, pre, nil, actname)
				if err != nil {
					fmt.Printf("WARNING: Execute %s failed: %v\n", actname, err)
					continue
				}
				if post != nil {
					pcs := mod.Postconds[actname]
					CheckConjsInStateWithAG(mod, ag, post, 12, pcs)
				}
			}
		}
	}

	// Python: callgraph, assumptions, and guarantee printing all run unconditionally.
	// Only the actual SMT verification is gated on `check`. The guarantee
	// header/loop is gated on `!NoCheckGuarantees` but NOT on `check`.

	// Build call graph (always — used by assumptions and guarantees)
	callgraph := make(map[string][]string)
	for actname, action := range mod.Actions.All() {
		if act, ok := action.(actions.Action); ok {
			for _, calledName := range act.IterCalls() {
				callgraph[calledName] = append(callgraph[calledName], actname)
			}
		}
	}

	if actions.AssertLocEnabled {
		for actname, action := range mod.Actions.All() {
			if a, ok := action.(actions.Action); ok {
				actions.AssertEveryActionHasLoc(a, "check.isolate_check.print actname="+actname)
			}
		}
	}

	// Print assumptions (always — not gated on check or NoCheckGuarantees)
	someAssumps := false
	for actname, action := range mod.Actions.All() {
		act, ok := action.(actions.Action)
		if !ok {
			continue
		}
		var assumptions []actions.Action
		for _, sub := range act.IterSubactions() {
			if _, isAssume := sub.(*actions.AssumeAction); isAssume {
				if !IsUnprovableAssert(sub) {
					assumptions = append(assumptions, sub)
				}
			}
		}
		if len(assumptions) > 0 {
			if !someAssumps {
				fmt.Println("\n    The following program assertions are treated as assumptions:")
				someAssumps = true
			}
			callers := callgraph[actname]
			if mod.PublicActions.Get(actname) {
				callers = append(callers, "the environment")
				callgraph[actname] = callers
			}
			prettyname := actname
			if strings.HasPrefix(prettyname, "ext:") {
				prettyname = prettyname[4:]
			}
			var prettycallers []string
			for _, c := range callers {
				if strings.HasPrefix(c, "ext:") {
					prettycallers = append(prettycallers, c[4:])
				} else {
					prettycallers = append(prettycallers, c)
				}
			}
			fmt.Printf("        in action %s when called from %s:\n", prettyname, strings.Join(prettycallers, ","))
			for _, sub := range assumptions {
				fmt.Printf("            %sassumption\n", PrettyActionLineno(sub))
			}
		}
	}

	// Guarantee phase (always iterates; only SMT verification gated on check)
	xtracer.Trace("check.guarantee_phase ENTER")
	tried := make(map[string]bool)
	someGuarants := false
	for actname, action := range mod.Actions.All() {
		xtracer.Trace("check.guarantee_phase actname iter actname=%s", actname)
		act, ok := action.(actions.Action)
		if !ok {
			xtracer.Trace("check.guarantee_phase actname=%s skip:not-Action", actname)
			continue
		}
		xtracer.Trace("check.guarantee_phase HASH actname=%s canon=%s", actname, act.Canon())
		var guarantees []actions.Action
		subCount := 0
		for _, sub := range act.IterSubactions() {
			subCount++
			isAssert := actions.IsAssertLike(sub)
			_, isRanking := sub.(*actions.Ranking)
			xtracer.Trace("check.guarantee_phase sub HASH actname=%s kind=%s isAssert=%v isRanking=%v canon=%s", actname, actions.ActionTypeName(sub), isAssert, isRanking, sub.Canon())
			if isAssert || isRanking {
				guarantees = append(guarantees, sub)
			}
		}
		xtracer.Trace("check.guarantee_phase post iter_subactions actname=%s subCount=%d guarantees=%d", actname, subCount, len(guarantees))
		if mod.Cfg.CheckLineno != "" {
			var filtered []actions.Action
			for _, sub := range guarantees {
				if sub.GetLineno().FileLineKey() == mod.Cfg.CheckLineno {
					filtered = append(filtered, sub)
				}
			}
			guarantees = filtered
			xtracer.Trace("check.guarantee_phase post check_lineno_filter actname=%s guarantees=%d", actname, len(guarantees))
		}
		{
			var filtered []actions.Action
			for _, sub := range guarantees {
				pass := IsGuaranteeModUnprovable(mod.Cfg, sub)
				xtracer.Trace("check.guarantee_phase unprov_filter HASH actname=%s pass=%v canon=%s", actname, pass, sub.Canon())
				if pass {
					filtered = append(filtered, sub)
				}
			}
			guarantees = filtered
			xtracer.Trace("check.guarantee_phase post unprov_filter actname=%s guarantees=%d", actname, len(guarantees))
		}
		// Python: if guarantees and not(no_check_guarantees.get()):
		if len(guarantees) > 0 && !mod.Cfg.NoCheckGuarantees {
			if !someGuarants {
				fmt.Println("\n    The following program assertions are treated as guarantees:")
				someGuarants = true
			}
			callers := callgraph[actname]
			if mod.PublicActions.Get(actname) {
				callers = append(callers, "the environment")
				callgraph[actname] = callers
			}
			prettyname := actname
			if strings.HasPrefix(prettyname, "ext:") {
				prettyname = prettyname[4:]
			}
			var prettycallers []string
			for _, c := range callers {
				if strings.HasPrefix(c, "ext:") {
					prettycallers = append(prettycallers, c[4:])
				} else {
					prettycallers = append(prettycallers, c)
				}
			}
			fmt.Printf("        in action %s when called from %s:\n", prettyname, strings.Join(prettycallers, ","))

			roots := reachable([]string{actname}, func(x string) []string { return callgraph[x] })
			for _, sub := range guarantees {
				xtracer.Trace("check.guarantee_outer iter")
				lineno := sub.GetLineno()
				fmt.Printf("            %sguarantee ", PrettyActionLineno(sub))
				linenoKey := lineno.FileLineKey()

				anyUntried := false
				for root := range checkedActions {
					if roots[root] && !tried[root+":"+linenoKey] {
						anyUntried = true
						break
					}
				}

				if anyUntried && check {
					PrintDots()
					oldCheckedAssert := mod.Cfg.CheckLineno
					mod.Cfg.CheckLineno = linenoKey
					someFailed := false
					for root := range checkedActions {
						if !roots[root] {
							continue
						}
						tried[root+":"+linenoKey] = true
						xtracer.Trace("check.guarantee_loop pre BuildEnvAction")
						envAction := actions.BuildEnvAction(mod.Cfg.ActCfg, mod.PublicActions, mod.Actions, root, "")
						xtracer.Trace("check.guarantee_loop post BuildEnvAction")
						ag := art.NewAnalysisGraph(mod)
						xtracer.Trace("check.guarantee_loop post NewAnalysisGraph")
						pre := art.NewState(mod, GetConjs(mod))
						xtracer.Trace("check.guarantee_loop post NewState+GetConjs")
						ag.Add(pre, nil)
						xtracer.Trace("check.guarantee_loop post ag.Add")
						xtracer.Trace("check.guarantee_loop pre ag.Execute")
						post, execErr := ag.Execute(false, envAction, pre, nil, root)
						if execErr != nil {
							fmt.Printf("WARNING: Execute %s failed: %v\n", root, execErr)
							continue
						}
						if post != nil {
							failState := art.NewState(mod, module.TrueClauses(actions.EmptyAnnotation{}))
							if aa, ok := post.Prov.(*art.ActionApp); ok {
								if rep, ok := aa.Rep.(string); ok {
									failState.Prov = art.NewActionApp("fail_"+rep, aa.Args...)
								}
							}
							failState.Action = actions.NewFailAction(envAction)
							if post.Pred != nil {
								failState.Pred = post.Pred
							} else if aa, ok := post.Prov.(*art.ActionApp); ok && len(aa.Args) > 0 {
								failState.Pred = aa.Args[0]
							}

							if !CheckSafetyInStateWithAG(mod, ag, failState, false) {
								someFailed = true
								break
							}
						}
					}
					if !someFailed {
						fmt.Println("PASS")
					}
					mod.Cfg.CheckLineno = oldCheckedAssert
				} else {
					fmt.Println("")
				}
			}
		}
	}

	// Move conjectures to assumed invariants
	xtracer.Trace("check.CheckIsolate preMove nAssumedInvs=%d nLabeledConjs=%d", len(mod.AssumedInvs), len(mod.LabeledConjs))
	mod.AssumedInvs = append(mod.AssumedInvs, mod.LabeledConjs...)
	mod.LabeledConjs = nil

	// Check temporals
	// Python: if not unprovable: check_temporals()
	if !mod.Cfg.OnlyCheckUnprovable {
		if err := CheckTemporals(mod); err != nil {
			return err
		}
	}

	xtracer.Trace("check.CheckIsolate EXIT")
	return nil
}

// CheckSubgoals checks proof subgoals by constructing fake isolates.
// Corresponds to Python's check_subgoals (lines 726-803).
// For each goal:
//   - If conclusion is TemporalModels: build a fake module from the model
//     and call CheckIsolate recursively
//   - Otherwise: convert goal to property and check in a minimal module
func CheckSubgoals(goals []*ast.LabeledFormula, method func() error, mod *module.Module) error {
	if mod == nil {
		mod = module.New()
	}

	for _, goal := range goals {
		_ = proof.GoalConc(goal) // used for non-temporal branch via goal itself

		// Check for TemporalModels via the formula directly, since
		// GoalConc returns lg.Expr and TemporalModels is ast.Node.
		var tm *ast.TemporalModels
		if sb, ok := goal.Formula.(*ast.SchemaBody); ok {
			if c := sb.Conc(); c != nil {
				tm, _ = c.(*ast.TemporalModels)
			}
		} else if goal.Formula != nil {
			tm, _ = goal.Formula.(*ast.TemporalModels)
		}

		if tm != nil {
			// TemporalModels branch (Python lines 731-763)
			model := tm.Model
			// Python: if not lg.is_true(fmla): raise error
			if tm.Fmla != nil {
				if fmlaExpr, ok := tm.Fmla.(lg.Expr); ok && !lg.IsTrue(fmlaExpr) {
					return fmt.Errorf("the temporal subgoal %v has not been reduced to an invariance property. Try using a tactic such as l2s", goal)
				}
			}
			// Python: mod = im.module.copy(); set fields from model
			withLocalMod := mod.Copy()
			withLocalMod.IsolateProof = nil
			withLocalMod.LabeledProps = nil
			withLocalMod.ConceptSpaces = nil

			// Extract fields from NormalProgram if available.
			if np, ok := model.(*temporal.NormalProgram); ok {
				withLocalMod.LabeledConjs = np.Invars
				if np.Postconds != nil {
					withLocalMod.Postconds = np.Postconds
				}
				withLocalMod.PublicActions = iu.NewInsMap[string, bool]()
				for _, c := range np.Calls {
					withLocalMod.PublicActions.Set(c, true)
				}
				withLocalMod.Actions = iu.NewInsMap[string, module.Action]()
				for _, b := range np.Bindings {
					withLocalMod.Actions.Set(b.Name, b.Action.Stmt)
				}
				if np.Init != nil {
					withLocalMod.Initializers = []module.NamedAction{{Name: "init", Action: np.Init}}
				} else {
					withLocalMod.Initializers = nil
				}
				withLocalMod.AssumedInvs = np.Asms
				xtracer.Trace("check.CheckSubgoals TemporalModels applied mod.AssumedInvs nAsms=%d nLabeledConjs=%d nInvars_from_model=%d",
					len(withLocalMod.AssumedInvs), len(withLocalMod.LabeledConjs), len(np.Invars))
			}

			// Python: mod.labeled_axioms = list(mod.labeled_axioms)
			axiomsCopy := make([]*ast.LabeledFormula, len(withLocalMod.LabeledAxioms))
			copy(axiomsCopy, withLocalMod.LabeledAxioms)
			withLocalMod.LabeledAxioms = axiomsCopy

			// Python: mod.params = list(mod.params)
			// Python: mod.updates = list(mod.updates)
			// (params and updates are copied by mod.Copy() already)

			// Add goal premises as axioms
			// Python: for prem in ivy_proof.goal_prems(goal):
			//             if ivy_proof.goal_is_property(prem):
			//                 if prem.definition: mod.updates.append(act.DerivedUpdate(df))
			//                 mod.labeled_axioms.append(prem)
			//             elif ivy_proof.goal_is_defn(prem):
			//                 dfnd = ivy_proof.goal_defines(prem)
			//                 if lg.is_constant(dfnd): mod.params.append(dfnd)
			goalPrems := proof.GoalPrems(goal)
			if xtracer.Enabled {
				nLF, nProp := 0, 0
				for _, p := range goalPrems {
					if lf, ok := p.(*ast.LabeledFormula); ok {
						nLF++
						if proof.GoalIsProperty(lf) {
							nProp++
						}
					}
				}
				xtracer.Trace("check.CheckSubgoals temporal prems nTotal=%d nLF=%d nProp=%d", len(goalPrems), nLF, nProp)
			}
			for _, premNode := range goalPrems {
				// Python iterates ALL premise types, not just LabeledFormula.
				// GoalIsProperty requires LF; GoalIsDefn handles ConstantDecl
				// and UninterpretedSort which are NOT LabeledFormulas.
				premLF, isLF := premNode.(*ast.LabeledFormula)
				if isLF && proof.GoalIsProperty(premLF) {
					if premLF.IsDefinition {
						// Python: df = lg.drop_universals(prem.formula)
						//         mod.updates.append(act.DerivedUpdate(df))
						if fmla, ok := premLF.Formula.(lg.Expr); ok {
							df := il.DropUniversals(fmla)
							// Python DerivedUpdate(df) stores df and uses df.args[0] as symbol.
							// Go NewDerivedUpdate(sym, defn) takes both. Extract sym from df.
							var sym lg.Expr
							if children := df.Children(); len(children) > 0 {
								sym = children[0]
							}
							withLocalMod.Updates = append(withLocalMod.Updates, module.NewDerivedUpdate(sym, df))
						}
					}
					modLF := AstLFToModuleLF(premLF)
					if modLF != nil {
						withLocalMod.LabeledAxioms = append(withLocalMod.LabeledAxioms, modLF)
					}
				} else if proof.GoalIsDefn(premNode) {
					dfnd := proof.GoalDefines(premNode)
					if dfnd != nil && il.IsConstant(dfnd) {
						if sym, ok := dfnd.(*lg.Const); ok {
							withLocalMod.Params = append(withLocalMod.Params, sym)
						}
					}
				}
			}

			// Enter module context and check with vocab
			// Python: with lg.WithSymbols → with lg.WithSorts →
			//             if method is not None: with im.module.theory_context(): method()
			//             else: check_isolate()
			vocab := proof.GoalVocab(goal)
			ws := il.NewWithSymbols(withLocalMod.Sig, vocab.Symbols)
			ws.Enter()
			wsorts := il.NewWithSorts(withLocalMod.Sig, vocab.Sorts)
			wsorts.Enter()
			if method != nil {
				// Python: if act.check_unprovable.get(): print("SKIPPED\n"); return
				// Check BEFORE entering theory context (matching Python ordering).
				if mod.Cfg.OnlyCheckUnprovable {
					fmt.Println("SKIPPED")
					wsorts.Exit()
					ws.Exit()
					return nil
				}
				cleanup := withLocalMod.TheoryContext()
				err := method()
				if err != nil {
					mod.Cfg.Failures++
					fmt.Println("FAIL")
					// Python ivy_check.py:829-830:
					//   if hasattr(goal,"trace_hook"): foo = goal.trace_hook(foo)
					// The Go trace-hook propagation differs (it acts on a
					// MatchHandler via withLocalMod.TraceHook in the no-method
					// branch below, not as a transformer here). For now we
					// honor opt_trace and diagnose without re-routing the
					// failure value through goal.TraceHook — see followup
					// item 5 of plans/velvety-kindling-shamir.md.
					// Python: if opt_trace.get(): print(str(foo)); exit(0)
					if mod.Cfg.OptTrace {
						fmt.Println(err)
						cleanup()
						wsorts.Exit()
						ws.Exit()
						os.Exit(0)
					}
					// Python: if diagnose.get(): gui_art(foo)
					if mod.Cfg.Diagnose {
						if guiErr := GuiArt(mod, err, nil); guiErr != nil {
							fmt.Fprintf(os.Stderr, "GuiArt: %v\n", guiErr)
						}
					}
					cleanup()
					wsorts.Exit()
					ws.Exit()
					return err
				}
				fmt.Println("PASS")
				cleanup()
				wsorts.Exit()
				ws.Exit()
			} else {
				// C5 / Python ivy_check.py:843-846: propagate goal trace hook
				// to the module so check.go's trace formatter can use it.
				// No TheoryContext here — CheckIsolate handles its own.
				if goal.TraceHook != nil {
					withLocalMod.TraceHook = goal.TraceHook
				}
				failsBefore := withLocalMod.Cfg.Failures
				err := CheckIsolate(withLocalMod, nil)
				// Python's failures is a module-level global, so all
				// Checker.fail() calls aggregate regardless of which module
				// copy is active. Propagate the delta back to the original.
				mod.Cfg.Failures += withLocalMod.Cfg.Failures - failsBefore
				// Mirror Python's list aliasing (ivy_check.py:795 mod.assumed_invariants
				// = model.asms — same list object as im.module.assumed_invariants).
				// Inside the recursive check_isolate (line 761) extending that list
				// persists past the `with mod:` exit. In Go, slices are values, so
				// we propagate the accumulated AssumedInvs back to the outer mod
				// explicitly. Without this, the next iteration of check_temporals'
				// prop loop sees a stale mod.AssumedInvs.
				mod.AssumedInvs = withLocalMod.AssumedInvs
				if err != nil {
					wsorts.Exit()
					ws.Exit()
					return err
				}
				wsorts.Exit()
				ws.Exit()
			}

		} else {
			// Non-temporal branch (Python lines 765-776)
			pgoal := compiler.TheoremToProperty(AstLFToModuleLF(goal), mod)
			withLocalMod := mod.Copy()
			withLocalMod.LabeledProps = []*ast.LabeledFormula{pgoal}
			withLocalMod.ConceptSpaces = nil
			withLocalMod.LabeledConjs = nil
			withLocalMod.PublicActions = iu.NewInsMap[string, bool]()
			withLocalMod.Actions = iu.NewInsMap[string, module.Action]()
			withLocalMod.Initializers = nil
			withLocalMod.IsolateProof = nil
			withLocalMod.IsolateInfo = nil

			// Enter module context and check with vocab
			// Python: with lg.WithSymbols → with lg.WithSorts →
			//             if method is not None: with im.module.theory_context(): method()
			//             else: check_isolate()
			vocab := proof.GoalVocab(goal)
			ws := il.NewWithSymbols(withLocalMod.Sig, vocab.Symbols)
			ws.Enter()
			wsorts := il.NewWithSorts(withLocalMod.Sig, vocab.Sorts)
			wsorts.Enter()
			if method != nil {
				// Python: if act.check_unprovable.get(): print("SKIPPED\n"); return
				// Check BEFORE entering theory context (matching Python ordering).
				if mod.Cfg.OnlyCheckUnprovable {
					fmt.Println("SKIPPED")
					wsorts.Exit()
					ws.Exit()
					return nil
				}
				cleanup := withLocalMod.TheoryContext()
				err := method()
				if err != nil {
					mod.Cfg.Failures++
					fmt.Println("FAIL")
					// Python ivy_check.py:829-835. See the parallel block in
					// the temporal branch above for the trace-hook caveat.
					// Python: if opt_trace.get(): print(str(foo)); exit(0)
					if mod.Cfg.OptTrace {
						fmt.Println(err)
						cleanup()
						wsorts.Exit()
						ws.Exit()
						os.Exit(0)
					}
					// Python: if diagnose.get(): gui_art(foo)
					if mod.Cfg.Diagnose {
						if guiErr := GuiArt(mod, err, nil); guiErr != nil {
							fmt.Fprintf(os.Stderr, "GuiArt: %v\n", guiErr)
						}
					}
					cleanup()
					wsorts.Exit()
					ws.Exit()
					return err
				}
				fmt.Println("PASS")
				cleanup()
				wsorts.Exit()
				ws.Exit()
			} else {
				// C5 / Python ivy_check.py:843-846: propagate goal trace hook
				// to the module so check.go's trace formatter can use it.
				// No TheoryContext here — CheckIsolate handles its own.
				if goal.TraceHook != nil {
					withLocalMod.TraceHook = goal.TraceHook
				}
				failsBefore := withLocalMod.Cfg.Failures
				err := CheckIsolate(withLocalMod, nil)
				mod.Cfg.Failures += withLocalMod.Cfg.Failures - failsBefore
				if err != nil {
					wsorts.Exit()
					ws.Exit()
					return err
				}
				wsorts.Exit()
				ws.Exit()
			}
		}
	}
	return nil
}

// CheckModule is the top-level module checker. It determines which
// isolates to check and dispatches to the appropriate checking method.
// This corresponds to Python's check_module.
func CheckModule(mod *module.Module) error {
	if xtracer.Enabled {
		xtracer.Trace("check.CheckModule ENTER isolates=%d", len(mod.Isolates))
	}
	var isolates []string

	// xtrace all the module data at this point.
	mod.CanonSnapshot("top-of-CheckModule")

	// Determine which isolates to check
	// Python: isolate = ivy_compiler.isolate.get()
	//         if isolate != None: isolates = [isolate]
	//         else: isolates = sorted(list(im.module.isolates))
	if mod.Cfg.Isolate != "" {
		isolates = []string{mod.Cfg.Isolate}
	} else if len(mod.Isolates) > 0 {
		for name := range mod.Isolates {
			isolates = append(isolates, name)
		}
		sort.Strings(isolates)
		if mod.Cfg.Coverage {
			if missing := ivyiso.CheckIsolateCompleteness(mod); len(missing) > 0 {
				return fmt.Errorf("some assertions are not checked")
			}
		}
	} else {
		isolates = []string{""}
	}

	if mod.Cfg.OptIvyStats {
		fmt.Printf(" +++ IVY_STATS starting checking module. Num isolates = %d\n", len(isolates))
	}

	// Python: if opt_unchecked_properties.get() != None:
	//             ivy_acl.register_from_file(opt_unchecked_properties.get())
	// Load ACL rules ONCE before the isolate loop (matching Python line 982-983).
	var aclCfg *acl.Config
	if mod.Cfg.OptUncheckedProps != "" {
		var err error
		aclCfg, err = acl.RegisterFromFile(mod.Cfg.OptUncheckedProps)
		if err != nil {
			return err
		}
	}

	for _, isolate := range isolates {
		if mod.Cfg.OptIvyStats {
			fmt.Printf("\n\tIVY_STATS checking isolate %s\n", isolate)
		}

		if isolate != "" {
			// Check if isolate has anything to verify or is trusted
			// Python: if len(idef.verified()) == 0 or isinstance(idef, ivy_ast.TrustedIsolateDef): continue
			if idef, exists := mod.Isolates[isolate]; exists {
				if len(idef.Verified()) == 0 || idef.Trusted {
					continue
				}
			}
			fmt.Printf("\nIsolate %s:\n", isolate)
		}

		// Python: macro_finder save/restore around isolate check
		// if isolate is not None and compose_names(isolate,'macro_finder') in mod.attributes:
		//     save_macro_finder = islv.opt_macro_finder.get()
		//     if save_macro_finder: print("Turning off macro_finder"); islv.set_macro_finder(False)
		hasMFAttr := false
		saveMacroFinder := false
		if isolate != "" {
			attrKey := mod.Cfg.IuCfg.ComposeNames(isolate, "macro_finder")
			if _, ok := mod.Attributes[attrKey]; ok {
				hasMFAttr = true
				saveMacroFinder = mod.Cfg.SolverOpts.MacroFinder
				if saveMacroFinder {
					fmt.Println("Turning off macro_finder")
					mod.Cfg.SolverOpts.MacroFinder = false
					xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=False")
				}
			}
		}

		// Create the isolate (flattens module hierarchy, resolves mixins, etc.)
		// Python: with im.module.copy(): ivy_isolate.create_isolate(isolate)
		// We copy the module so modifications don't leak.
		isoMod := mod.Copy()
		if err := ivyiso.CreateIsolate(isolate, isoMod); err != nil {
			return fmt.Errorf("create_isolate(%s): %w", isolate, err)
		}

		if mod.Cfg.OptTrusted {
			continue
		}

		// Preprocess assumed/ignored properties using the pre-loaded ACL config
		if aclCfg != nil {
			isoMod.AclCfg = aclCfg
			PreprocessAssumedIgnoredProperties(isoMod, aclCfg)
		}

		methodName := GetIsolateMethod(isolate, isoMod)
		switch {
		case methodName == "mc":
			mcMethod := func(m *module.Module) error {
				res, err := mc.CheckIsolate(m, "mc")
				if err != nil {
					return err
				}
				if res != nil && !res.Proved {
					if res.Error != nil {
						return res.Error
					}
					return fmt.Errorf("model checking failed")
				}
				return nil
			}
			if err := MCIsolate(isolate, isoMod, mcMethod); err != nil {
				return err
			}
		case methodName == "vmt":
			vmtMethod := func(m *module.Module) error {
				return vmt.CheckIsolate("mc", m)
			}
			if err := MCIsolate(isolate, isoMod, vmtMethod); err != nil {
				return err
			}
		case strings.HasPrefix(methodName, "bmc["):
			mod.Cfg.SomeBounded = true
			nSteps, nUnroll, err := parseBMCParams(methodName)
			if err != nil {
				return err
			}
			bmcMethod := func(m *module.Module) error {
				cfg := &bmc.Config{
					NSteps: nSteps,
					Module: m,
				}
				if nUnroll >= 0 {
					nu := nUnroll
					cfg.NUnroll = &nu
				}
				res := bmc.CheckIsolate(cfg)
				if res != nil && res.Found {
					return fmt.Errorf("%s", res.Message)
				}
				return nil
			}
			if err := MCIsolate(isolate, isoMod, bmcMethod); err != nil {
				return err
			}
		default:
			// Python: logic = get_isolate_attr(isolate, 'complete', None)
			//         if logic is not None: im.module.logics = [logic]
			logic := GetIsolateAttr(isolate, "complete", "", isoMod)
			if logic != "" {
				isoMod.Logics = []string{logic}
			}
			// Python check_module (ivy_check.py:974) calls check_isolate()
			// without a theory_context() wrapper. CheckIsolate handles
			// TheoryContext internally (after CheckFragment).
			failsBefore := isoMod.Cfg.Failures
			err := CheckIsolate(isoMod, nil)
			// Python's failures is a module-level global; propagate
			// delta back from the copy to the original mod.
			mod.Cfg.Failures += isoMod.Cfg.Failures - failsBefore
			if err != nil {
				return err
			}
		}

		// Python: macro_finder restore
		// if isolate is not None and compose_names(isolate,'macro_finder') in mod.attributes:
		//     if save_macro_finder: print("Turning on macro_finder"); islv.set_macro_finder(True)
		if hasMFAttr && saveMacroFinder {
			fmt.Println("Turning on macro_finder")
			mod.Cfg.SolverOpts.MacroFinder = true
			xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=True")
		}
	}

	fmt.Println()
	if mod.Cfg.Failures > 0 {
		return fmt.Errorf("check/isolate_check.go:1126: failed checks: %d", mod.Cfg.Failures)
	}
	if mod.Cfg.CheckedAction != "" && !mod.Cfg.CheckedActionFound {
		return fmt.Errorf("%s is not an exported action of any isolate",
			mod.Cfg.CheckedAction)
	}
	xtracer.Trace("check.CheckModule EXIT")
	return nil
}

// MCIsolate model-checks an isolate using the given method.
// Corresponds to Python's mc_isolate (ivy_check.py:871-892).
//
// The method parameter is the model checking backend to call:
//   - mc.CheckIsolate (default)
//   - bmc.CheckIsolate
//   - vmt.CheckIsolate
//
// If method is nil, this is a no-op (the caller should have provided one).
func MCIsolate(isolate string, mod *module.Module, method func(*module.Module) error) error {
	for _, p := range mod.LabeledProps {
		if !p.IsTemporal() {
			return fmt.Errorf("model checking not supported for non-temporal property yet")
		}
	}

	if method == nil {
		return nil
	}

	if !CheckSeparately(isolate, mod) {
		cleanup := mod.TheoryContext()
		err := method(mod)
		cleanup()
		if err != nil {
			fmt.Println(err)
			fmt.Println("FAIL")
			return err
		}
		return nil
	}

	// Python: for lineno in all_assert_linenos():
	//             with im.module.copy():
	//                 act.checked_assert.value = lineno
	//                 with im.module.theory_context(): res = meth()
	//                 act.checked_assert.value = old_checked_assert
	linenos, err := AllAssertLinenos(mod)
	if err != nil {
		return err
	}
	for _, lineno := range linenos {
		modCopy := mod.Copy()
		oldCheckedAssert := mod.Cfg.CheckLineno
		modCopy.Cfg.CheckLineno = lineno.FileLineKey()
		cleanup := modCopy.TheoryContext()
		err := method(modCopy)
		cleanup()
		mod.Cfg.CheckLineno = oldCheckedAssert
		if err != nil {
			fmt.Println(err)
			fmt.Println("FAIL")
			return err
		}
	}
	return nil
}

// GetIsolateMethod returns the verification method for an isolate.
// Returns "mc", "vmt", "bmc[...]", or "ic" (default).
func GetIsolateMethod(isolate string, mod *module.Module) string {
	if mod.Cfg.OptMC {
		return "mc"
	}
	return GetIsolateAttr(isolate, "method", "ic", mod)
}

// GetIsolateAttr retrieves an attribute for an isolate from the module.
// Corresponds to Python's get_isolate_attr (ivy_check.py:854-864).
// Python returns im.module.attributes[attr].rep (the string representation
// of the attribute AST node). In Go, attributes are interface{} — we use
// fmt.Sprint to get the string value, matching how other Go code handles
// attribute values (e.g., isolate/helpers.go:199).
func GetIsolateAttr(isolate, attrName, defaultVal string, mod *module.Module) string {
	if isolate == "" {
		return defaultVal
	}
	attr := mod.Cfg.IuCfg.ComposeNames(isolate, attrName)
	val, ok := mod.Attributes[attr]
	if !ok {
		pc := mod.Cfg.IuCfg.ParentChildName(isolate)
		if pc[1] == "iso" {
			attr = mod.Cfg.IuCfg.ComposeNames(pc[0], attrName)
		}
		val, ok = mod.Attributes[attr]
		if !ok {
			return defaultVal
		}
	}
	// Extract string representation, matching Python's .rep access.
	if s, ok := val.(string); ok {
		return s
	}
	if s, ok := val.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprint(val)
}

// CheckSeparately returns whether to check assertions separately.
// Python: opt_separate is BooleanParameter("separate", None) — tri-state.
//
//	if opt_separate.get() is not None: return opt_separate.get()
//	return get_isolate_attr(isolate,'separate','false') == 'true'
func CheckSeparately(isolate string, mod *module.Module) bool {
	if mod.Cfg.OptSeparateSet {
		return mod.Cfg.OptSeparate
	}
	return GetIsolateAttr(isolate, "separate", "false", mod) == "true"
}

// AllAssertLinenos returns all unique assertion locations in the module.
// Corresponds to Python all_assert_linenos (ivy_check.py:833-852).
// Returns full ast.Location values so callers can produce "file:line" keys.
// Returns an error if a checked_assert line is specified but not found.
func AllAssertLinenos(mod *module.Module) ([]ast.Location, error) {
	seen := make(map[string]bool)
	var result []ast.Location

	for _, action := range mod.Actions.All() {
		if act, ok := action.(interface{ IterSubactions() []actions.Action }); ok {
			for _, sub := range act.IterSubactions() {
				isAssert := actions.IsAssertLike(sub)
				_, isRanking := sub.(*actions.Ranking)
				if isAssert || isRanking {
					loc := sub.GetLineno()
					key := loc.FileLineKey()
					if !seen[key] {
						seen[key] = true
						result = append(result, loc)
					}
				}
			}
		}
	}

	for _, lf := range mod.LabeledConjs {
		loc := lf.GetLineno()
		key := loc.FileLineKey()
		if !seen[key] {
			seen[key] = true
			result = append(result, loc)
		}
	}

	if mod.Cfg.CheckLineno != "" {
		if seen[mod.Cfg.CheckLineno] {
			for _, loc := range result {
				if loc.FileLineKey() == mod.Cfg.CheckLineno {
					return []ast.Location{loc}, nil
				}
			}
		}
		return nil, fmt.Errorf("there is no assertion at the specified line")
	}
	return result, nil
}

// parseBMCParams parses "bmc[N]" or "bmc[N][M]" into (nSteps, nUnroll, err).
// nUnroll is -1 if not specified.
// Corresponds to Python: iu.parse_int_subscripts(method_name)
func parseBMCParams(methodName string) (nSteps, nUnroll int, err error) {
	nUnroll = -1
	if !strings.HasPrefix(methodName, "bmc[") {
		return 0, -1, fmt.Errorf("invalid BMC method specifier: %q", methodName)
	}
	rest := methodName[3:] // "bmc" prefix removed, rest starts with "["
	var params []int
	for len(rest) > 0 && rest[0] == '[' {
		end := strings.Index(rest, "]")
		if end < 0 {
			return 0, -1, fmt.Errorf("invalid BMC method specifier: %q (missing ']')", methodName)
		}
		val, err := strconv.Atoi(rest[1:end])
		if err != nil {
			return 0, -1, fmt.Errorf("invalid BMC method specifier: %q (%w)", methodName, err)
		}
		params = append(params, val)
		rest = rest[end+1:]
	}
	if len(params) < 1 || len(params) > 2 {
		return 0, -1, fmt.Errorf("BMC method specifier should be bmc[<steps>] or bmc[<steps>][<unroll>]. Got %q", methodName)
	}
	nSteps = params[0]
	if len(params) >= 2 {
		nUnroll = params[1]
	}
	return nSteps, nUnroll, nil
}

// --- Helper set functions ---

func setFromSlice(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, v := range s {
		m[v] = true
	}
	return m
}

func intersect(a, b map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for k := range a {
		if b[k] {
			result[k] = true
		}
	}
	return result
}

func sortedUnion(priority, all map[string]bool) []string {
	var plist []string
	for k := range priority {
		plist = append(plist, k)
	}
	sort.Strings(plist)

	remaining := make(map[string]bool)
	for k := range all {
		if !priority[k] {
			remaining[k] = true
		}
	}
	var rlist []string
	for k := range remaining {
		rlist = append(rlist, k)
	}
	sort.Strings(rlist)

	return append(plist, rlist...)
}

// --- AG-aware check functions ---

// CheckConjsInStateWithAG checks conjectures against a state using the
// analysis graph. This is the full version of CheckConjsInState.
func CheckConjsInStateWithAG(mod *module.Module, ag *art.AnalysisGraph, post *art.State, indent int, pcs []*ast.LabeledFormula) bool {
	conjs := mod.ConjSubgoals
	if conjs == nil {
		conjs = mod.LabeledConjs
	}

	// Filter for checkable conjectures using is_check_mod_unprovable.
	// Python: conjs = [x for x in conjs if is_check_mod_unprovable(x)]
	var checkable []*ast.LabeledFormula
	for _, c := range conjs {
		if IsCheckModUnprovable(mod.Cfg, c) {
			checkable = append(checkable, c)
		}
	}

	// Append converted postconditions using the post state's update.
	// Python: conjs += convert_postconds(post, pcs)
	if len(pcs) > 0 {
		var update *actions.Update
		if post != nil {
			update = post.Update
		}
		converted := ConvertPostcondsWithUpdate(update, pcs)
		checkable = append(checkable, converted...)
	}

	// Apply line-number filter if set (file:line format).
	checkLineno := mod.Cfg.CheckLineno
	if checkLineno != "" {
		var filtered []*ast.LabeledFormula
		for _, c := range checkable {
			if c.GetLineno().FileLineKey() == checkLineno {
				filtered = append(filtered, c)
			}
		}
		checkable = filtered
	}

	var checkers []Checker
	for _, c := range checkable {
		checkers = append(checkers, NewConjChecker(mod, c, indent))
	}

	return CheckFcsInStateWithAG(mod, ag, post, checkers)
}

// CheckSafetyInStateWithAG checks safety in a state using the analysis graph.
func CheckSafetyInStateWithAG(mod *module.Module, ag *art.AnalysisGraph, post *art.State, reportPass bool) bool {
	checker := NewBaseChecker(mod, &lg.Or{}, reportPass, true)
	return CheckFcsInStateWithAG(mod, ag, post, []Checker{checker})
}

// --- Helper functions ---

// filterExplicitSubgoals filters out explicit subgoals from the list.
func filterExplicitSubgoals(props []*ast.LabeledFormula, subgoalMap map[int64]bool) []*ast.LabeledFormula {
	var result []*ast.LabeledFormula
	for _, p := range props {
		if !(subgoalMap[p.ID] && p.Explicit) {
			result = append(result, p)
		}
	}
	return result
}

// reachable computes the set of nodes reachable from the starting nodes
// via the successors function. Corresponds to Python's iu.reachable.
func reachable(starts []string, successors func(string) []string) map[string]bool {
	visited := make(map[string]bool)
	var queue []string
	queue = append(queue, starts...)
	for _, s := range starts {
		visited[s] = true
	}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		for _, succ := range successors(node) {
			if !visited[succ] {
				visited[succ] = true
				queue = append(queue, succ)
			}
		}
	}
	return visited
}
