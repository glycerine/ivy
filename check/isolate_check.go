package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	"github.com/glycerine/goivy/compiler"
	ivyiso "github.com/glycerine/goivy/isolate"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
	iu "github.com/glycerine/goivy/ivyutils"
)

// CheckIsolate is the main isolate checking function.
// It orchestrates verification of properties, conjectures, and guarantees
// within a single isolate. This corresponds to Python's check_isolate.
//
// The traceHook parameter, if non-nil, is called to process trace output.
func CheckIsolate(mod *module.Module, traceHook func(interface{}) interface{}) error {
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

	// If there is an isolate proof, handle it via proof checker
	if mod.IsolateProof != nil {
		// Stub: requires proof checker, temporal model
		fmt.Println("    [isolate proof handling not yet implemented]")
		return nil
	}

	check := !OptSummary.GetBool()

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

	if (len(mod.LabeledProps) > 0 || len(schemaInstances) > 0) &&
		CheckedAction.GetString() == "" && check {
		fmt.Println("\n    The following properties are to be checked:")
		for _, lf := range schemaInstances {
			fmt.Println(PrettyLF(lf, 8) + " [proved by axiom schema]")
		}
		// Property checking: create AG with True pre-state, check properties
		// Python: ag = ivy_art.AnalysisGraph()
		//         pre = itp.State(value = true_clauses)
		//         check_fcs_in_state(mod, ag, pre, fcs)
		ag := art.NewAnalysisGraph(mod)
		pre := art.NewState(mod, clauseops.TrueClauses(actions.EmptyAnnotation{}))
		ag.Add(pre, nil)

		nonTemporal := make([]*ast.LabeledFormula, 0)
		for _, p := range mod.LabeledProps {
			if !p.Temporal {
				nonTemporal = append(nonTemporal, p)
			}
		}
		// Filter out explicitly proved subgoals
		nonTemporal = filterExplicitSubgoals(nonTemporal, subgoalMap)
		var checkers []Checker
		for _, prop := range nonTemporal {
			if prop.Assumed || subgoalMap[prop.ID] {
				checkers = append(checkers, NewConjAssumer(prop))
			} else {
				checkers = append(checkers, NewConjChecker(prop, 8))
			}
		}
		if len(checkers) > 0 {
			CheckFcsInStateWithAG(mod, ag, pre, checkers)
		}
	}

	// After checking properties, make non-temporal ones axioms
	for _, p := range mod.LabeledProps {
		if !p.Temporal {
			mod.LabeledAxioms = append(mod.LabeledAxioms, p)
		}
	}

	// Print initial properties
	if len(mod.LabeledInits) > 0 {
		fmt.Println("\n    The following properties are assumed initially:")
		for _, lf := range mod.LabeledInits {
			fmt.Println(PrettyLF(lf, 8))
		}
	}

	// Print checked invariants
	checkedInvariants := mod.LabeledConjs
	if len(checkedInvariants) > 0 {
		fmt.Println("\n    The inductive invariant consists of the following conjectures:")
		for _, lf := range checkedInvariants {
			fmt.Println(PrettyLF(lf, 8))
		}
	}

	// Apply conjecture proofs
	ApplyConjProofs(mod)

	// Print isolate implementations
	if mod.IsolateInfo != nil && len(mod.IsolateInfo.Implementations) > 0 {
		fmt.Println("\n    The following action implementations are present:")
		impls := make([]module.MixinTriple, len(mod.IsolateInfo.Implementations))
		copy(impls, mod.IsolateInfo.Implementations)
		sort.Slice(impls, func(i, j int) bool { return impls[i].Mixer < impls[j].Mixer })
		for _, impl := range impls {
			fmt.Printf("        implementation of %s\n", impl.Mixee)
		}
	}

	// Print monitors
	if mod.IsolateInfo != nil && len(mod.IsolateInfo.Monitors) > 0 {
		fmt.Println("\n    The following action monitors are present:")
		mons := make([]module.MixinTriple, len(mod.IsolateInfo.Monitors))
		copy(mons, mod.IsolateInfo.Monitors)
		sort.Slice(mons, func(i, j int) bool { return mons[i].Mixer < mons[j].Mixer })
		for _, mon := range mons {
			fmt.Printf("        monitor of %s\n", mon.Mixee)
		}
	}

	// Print initializers
	if len(mod.Initializers) > 0 {
		fmt.Println("\n    The following initializers are present:")
		inits := make([]module.NamedAction, len(mod.Initializers))
		copy(inits, mod.Initializers)
		sort.Slice(inits, func(i, j int) bool { return inits[i].Name < inits[j].Name })
		for _, na := range inits {
			fmt.Printf("        %s\n", na.Name)
		}
	}

	// Check initialization establishes invariant.
	// Python: ag = ivy_art.AnalysisGraph(initializer=lambda x:None)
	//         check_conjs_in_state(mod, ag, ag.states[0])
	// The initializer=lambda x:None means "use init_cond, no abstraction."
	// AddInitialState computes init state from mod.InitCond + initializer actions.
	if len(checkedInvariants) > 0 && CheckedAction.GetString() == "" && check {
		fmt.Println("\n    Initialization must establish the invariant")
		ag := art.NewAnalysisGraph(mod)
		ag.Initialize(func(s *art.State) {}) // no-op abstractor, matching Python
		if len(ag.States) > 0 {
			CheckConjsInStateWithAG(mod, ag, ag.States[0], 8, nil)
		}
	}

	// Check initializer assertions
	if len(mod.Initializers) > 0 {
		var guarantees []actions.Action
		for _, na := range mod.Initializers {
			if act, ok := na.Action.(actions.Action); ok {
				for _, sub := range act.IterSubactions() {
					if _, isAssert := sub.(*actions.AssertAction); isAssert {
						if IsGuaranteeModUnprovable(sub) {
							guarantees = append(guarantees, sub)
						}
					}
				}
			}
		}
		if len(guarantees) > 0 && check {
			fmt.Print("\n    Any assertions in initializers must be checked ")
			ag := art.NewAnalysisGraph(mod)
			ag.Initialize(func(s *art.State) {}) // no-op abstractor
			if len(ag.States) > 0 {
				CheckSafetyInStateWithAG(mod, ag, ag.States[0], true)
			}
		}
	}

	// Check that external actions preserve the invariant
	checkedActions := setFromSlice(GetCheckedActions(mod))
	prioritized := setFromSlice(GetPrioritizedActions())
	prioritizedChecked := intersect(prioritized, checkedActions)
	actionOrder := sortedUnion(prioritizedChecked, checkedActions)

	if len(checkedActions) > 0 && len(checkedInvariants) > 0 {
		fmt.Println("\n    The following set of external actions must preserve the invariant:")
		if PriorityActions.Get() != nil {
			var plist []string
			for k := range prioritizedChecked {
				plist = append(plist, k)
			}
			sort.Strings(plist)
			fmt.Printf("\n         (prioritized checking order: %v)\n", plist)
		}
		for _, actname := range actionOrder {
			// Build the env_action for this external action
			action := actions.BuildEnvAction(mod.PublicActions, mod.Actions, actname, "")
			fmt.Printf("        %s\n", actname)
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
				post := ag.Execute(action, pre, nil, actname)
				if post != nil {
					pcs := mod.Postconds[actname]
					CheckConjsInStateWithAG(mod, ag, post, 12, pcs)
				}
			}
		}
	}

	// Check guarantees (assert actions)
	// Python: iterates all actions, finds AssertActions, checks reachable
	// roots in checked_actions, and verifies safety for each.
	if !NoCheckGuarantees.GetBool() && check {
		// Build call graph
		callgraph := make(map[string][]string)
		for actname, action := range mod.Actions {
			if act, ok := action.(actions.Action); ok {
				for _, calledName := range act.IterCalls() {
					callgraph[calledName] = append(callgraph[calledName], actname)
				}
			}
		}

		// Print assumptions
		someAssumps := false
		for actname, action := range mod.Actions {
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
				if mod.PublicActions[actname] {
					callers = append(callers, "the environment")
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
					fmt.Printf("            %sassumption\n", prettyActionLineno(sub))
				}
			}
		}

		// Check guarantees
		tried := make(map[string]bool)
		someGuarants := false
		for actname, action := range mod.Actions {
			act, ok := action.(actions.Action)
			if !ok {
				continue
			}
			var guarantees []actions.Action
			for _, sub := range act.IterSubactions() {
				if _, isAssert := sub.(*actions.AssertAction); isAssert {
					if IsGuaranteeModUnprovable(sub) {
						guarantees = append(guarantees, sub)
					}
				}
			}
			if len(guarantees) > 0 {
				if !someGuarants {
					fmt.Println("\n    The following program assertions are treated as guarantees:")
					someGuarants = true
				}
				callers := callgraph[actname]
				if mod.PublicActions[actname] {
					callers = append(callers, "the environment")
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
					lineno := sub.GetLineno()
					fmt.Printf("            %sguarantee ", prettyActionLineno(sub))
					linenoKey := fmt.Sprintf("%d", lineno.Line)

					// Check if any root is in checked actions and hasn't been tried
					anyUntried := false
					for root := range checkedActions {
						if roots[root] && !tried[root+":"+linenoKey] {
							anyUntried = true
							break
						}
					}

					if anyUntried {
						fmt.Print("... ")
						someFailed := false
						for root := range checkedActions {
							if !roots[root] {
								continue
							}
							tried[root+":"+linenoKey] = true
							envAction := actions.BuildEnvAction(mod.PublicActions, mod.Actions, root, "")
							ag := art.NewAnalysisGraph(mod)
							pre := art.NewState(mod, GetConjs(mod))
							ag.Add(pre, nil)
							post := ag.Execute(envAction, pre, nil, root)
							if post != nil {
								if !CheckSafetyInStateWithAG(mod, ag, post, false) {
									someFailed = true
									break
								}
							}
						}
						if !someFailed {
							fmt.Println("PASS")
						}
					} else {
						fmt.Println("")
					}
				}
			}
		}
	}

	// Move conjectures to assumed invariants
	mod.AssumedInvs = append(mod.AssumedInvs, mod.LabeledConjs...)
	mod.LabeledConjs = nil

	// Check temporals
	if err := CheckTemporals(mod); err != nil {
		return err
	}

	return nil
}

// CheckSubgoals checks proof subgoals by constructing fake isolates.
// Corresponds to Python's check_subgoals (lines 726-803).
// For each goal:
//   - If conclusion is TemporalModels: build a fake module from the model
//     and call CheckIsolate recursively
//   - Otherwise: convert goal to property and check in a minimal module
func CheckSubgoals(goals []*ast.LabeledFormula, method func() error) error {
	mod := module.CurrentModule()
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
			_ = tm // model = conc.model; fmla = conc.fmla
			// Python: if not lg.is_true(fmla): raise error
			// Python: mod = im.module.copy(); set fields from model
			fakeMod := mod.Copy()
			fakeMod.IsolateProof = nil
			fakeMod.LabeledProps = nil
			fakeMod.ConceptSpaces = nil
			// Python: mod.labeled_conjs = model.invars
			// Python: mod.public_actions = set(model.calls)
			// Python: mod.actions = model.binding_map
			// Python: mod.initializers = [('init', model.init)]
			// Python: mod.assumed_invariants = model.asms
			// The model is stored in tm.Model but as ast.Node, not temporal.NormalProgram.
			// Add goal premises as axioms
			for _, premNode := range proof.GoalPrems(goal) {
				premLF, ok := premNode.(*ast.LabeledFormula)
				if !ok {
					continue
				}
				if proof.GoalIsProperty(premLF) {
					modLF := AstLFToModuleLF(premLF)
					if modLF != nil {
						fakeMod.LabeledAxioms = append(fakeMod.LabeledAxioms, modLF)
					}
				}
			}

			// Enter module context and check
			cleanup := fakeMod.TheoryContext()
			if method != nil {
				if CheckUnprovable.GetBool() {
					fmt.Println("SKIPPED")
					cleanup()
					continue
				}
				err := method()
				if err != nil {
					Failures++
					fmt.Println("FAIL")
					cleanup()
					return err
				}
				fmt.Println("PASS")
			} else {
				err := CheckIsolate(fakeMod, nil)
				if err != nil {
					cleanup()
					return err
				}
			}
			cleanup()

		} else {
			// Non-temporal branch (Python lines 765-776)
			pgoal := compiler.TheoremToProperty(AstLFToModuleLF(goal))
			fakeMod := mod.Copy()
			fakeMod.LabeledProps = []*ast.LabeledFormula{pgoal}
			fakeMod.ConceptSpaces = nil
			fakeMod.LabeledConjs = nil
			fakeMod.PublicActions = make(map[string]bool)
			fakeMod.Actions = make(map[string]interface{})
			fakeMod.Initializers = nil
			fakeMod.IsolateProof = nil
			fakeMod.IsolateInfo = nil

			// Enter module context and check
			cleanup := fakeMod.TheoryContext()
			if method != nil {
				if CheckUnprovable.GetBool() {
					fmt.Println("SKIPPED")
					cleanup()
					continue
				}
				err := method()
				if err != nil {
					Failures++
					fmt.Println("FAIL")
					cleanup()
					return err
				}
				fmt.Println("PASS")
			} else {
				err := CheckIsolate(fakeMod, nil)
				if err != nil {
					cleanup()
					return err
				}
			}
			cleanup()
		}
	}
	return nil
}

// CheckModule is the top-level module checker. It determines which
// isolates to check and dispatches to the appropriate checking method.
// This corresponds to Python's check_module.
func CheckModule(mod *module.Module) error {
	var isolates []string

	// Determine which isolates to check
	// Stub: the compiler.isolate parameter is not yet accessible
	if len(mod.Isolates) > 0 {
		for name := range mod.Isolates {
			isolates = append(isolates, name)
		}
		sort.Strings(isolates)
		if Coverage.GetBool() {
			// Stub: check_isolate_completeness
		}
	} else {
		isolates = []string{""}
	}

	if OptIvyStats.GetBool() {
		fmt.Printf(" +++ IVY_STATS starting checking module. Num isolates = %d\n", len(isolates))
	}

	for _, isolate := range isolates {
		if OptIvyStats.GetBool() {
			fmt.Printf("\n\tIVY_STATS checking isolate %s\n", isolate)
		}

		if isolate != "" {
			// Check if isolate has anything to verify
			idef, exists := mod.Isolates[isolate]
			if exists {
				if isoDef, ok := idef.(*ast.IsolateDef); ok {
					if len(isoDef.Verified()) == 0 {
						continue
					}
				}
				if _, ok := idef.(*ast.TrustedIsolateDef); ok {
					continue
				}
			}
			fmt.Printf("\nIsolate %s:\n", isolate)
		}

		// Create the isolate (flattens module hierarchy, resolves mixins, etc.)
		// Python: with im.module.copy(): ivy_isolate.create_isolate(isolate)
		// We copy the module so modifications don't leak.
		isoMod := mod.Copy()
		if err := ivyiso.CreateIsolate(isolate, isoMod); err != nil {
			return fmt.Errorf("create_isolate(%s): %w", isolate, err)
		}

		if OptTrusted.GetBool() {
			continue
		}

		// Preprocess assumed/ignored properties if ACL file is specified
		if OptUncheckedProps.Get() != nil {
			PreprocessAssumedIgnoredProperties(isoMod)
		}

		methodName := GetIsolateMethod(isolate, isoMod)
		switch {
		case methodName == "mc":
			if err := MCIsolate(isolate, isoMod, nil); err != nil {
				return err
			}
		case methodName == "vmt":
			if err := MCIsolate(isolate, isoMod, nil); err != nil {
				return err
			}
		case strings.HasPrefix(methodName, "bmc["):
			if err := MCIsolate(isolate, isoMod, nil); err != nil {
				return err
			}
		default:
			// Set up theory context for the isolated module
			cleanup := isoMod.TheoryContext()
			err := CheckIsolate(isoMod, nil)
			cleanup()
			if err != nil {
				return err
			}
		}
	}

	fmt.Println()
	if Failures > 0 {
		return fmt.Errorf("failed checks: %d", Failures)
	}
	if CheckedAction.GetString() != "" && !CheckedActionFound {
		return fmt.Errorf("%s is not an exported action of any isolate",
			CheckedAction.GetString())
	}
	return nil
}

// MCIsolate model-checks an isolate using the given method.
// This is a stub corresponding to Python's mc_isolate.
func MCIsolate(isolate string, mod *module.Module, method func() error) error {
	// Check that all properties are temporal
	for _, p := range mod.LabeledProps {
		if !p.Temporal {
			return fmt.Errorf("model checking not supported for non-temporal property yet")
		}
	}

	if !CheckSeparately(isolate, mod) {
		if method != nil {
			if err := method(); err != nil {
				fmt.Println(err)
				fmt.Println("FAIL")
				return err
			}
		}
		return nil
	}

	// Check separately per assertion line
	for _, lineno := range AllAssertLinenos(mod) {
		_ = lineno
		if method != nil {
			if err := method(); err != nil {
				fmt.Println(err)
				fmt.Println("FAIL")
				return err
			}
		}
	}
	return nil
}

// GetIsolateMethod returns the verification method for an isolate.
// Returns "mc", "vmt", "bmc[...]", or "ic" (default).
func GetIsolateMethod(isolate string, mod *module.Module) string {
	if OptMC.GetBool() {
		return "mc"
	}
	return GetIsolateAttr(isolate, "method", "ic", mod)
}

// GetIsolateAttr retrieves an attribute for an isolate from the module.
func GetIsolateAttr(isolate, attrName, defaultVal string, mod *module.Module) string {
	if isolate == "" {
		return defaultVal
	}
	attr := iu.ComposeNames(isolate, attrName)
	if _, ok := mod.Attributes[attr]; !ok {
		pc := iu.ParentChildName(isolate)
		if pc[1] == "iso" {
			attr = iu.ComposeNames(pc[0], attrName)
		}
		if _, ok := mod.Attributes[attr]; !ok {
			return defaultVal
		}
	}
	// Stub: would extract .rep from the attribute value
	return defaultVal
}

// CheckSeparately returns whether to check assertions separately.
func CheckSeparately(isolate string, mod *module.Module) bool {
	if OptSeparate.Get() != nil {
		if b, ok := OptSeparate.Get().(bool); ok {
			return b
		}
	}
	return GetIsolateAttr(isolate, "separate", "false", mod) == "true"
}

// AllAssertLinenos returns all unique assertion line numbers in the module.
func AllAssertLinenos(mod *module.Module) []int {
	seen := make(map[int]bool)
	var result []int

	for _, action := range mod.Actions {
		if act, ok := action.(interface{ IterSubactions() []actions.Action }); ok {
			for _, sub := range act.IterSubactions() {
				if _, isAssert := sub.(*actions.AssertAction); isAssert {
					loc := sub.GetLineno()
					if !seen[loc.Line] {
						seen[loc.Line] = true
						result = append(result, loc.Line)
					}
				}
			}
		}
	}

	for _, lf := range mod.LabeledConjs {
		if !seen[lf.Lineno] {
			seen[lf.Lineno] = true
			result = append(result, lf.Lineno)
		}
	}

	return result
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

	var checkable []*ast.LabeledFormula
	for _, c := range conjs {
		if !c.Unprovable {
			checkable = append(checkable, c)
		}
	}

	if len(pcs) > 0 {
		converted := ConvertPostconds(pcs)
		checkable = append(checkable, converted...)
	}

	var checkers []Checker
	for _, c := range checkable {
		checkers = append(checkers, NewConjChecker(c, indent))
	}

	if CheckLineno != "" {
		checkers = FilterCheckers(checkers, CheckLineno)
	}

	return CheckFcsInStateWithAG(mod, ag, post, checkers)
}

// CheckSafetyInStateWithAG checks safety in a state using the analysis graph.
func CheckSafetyInStateWithAG(mod *module.Module, ag *art.AnalysisGraph, post *art.State, reportPass bool) bool {
	checker := NewBaseChecker(&lg.Or{}, reportPass, true)
	return CheckFcsInStateWithAG(mod, ag, post, []Checker{checker})
}

// --- Helper functions ---

// prettyActionLineno formats a line number from an action.
func prettyActionLineno(act actions.Action) string {
	if act == nil {
		return "(internal) "
	}
	loc := act.GetLineno()
	if loc.Line > 0 {
		return fmt.Sprintf("line %d: ", loc.Line)
	}
	return "(internal) "
}

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
