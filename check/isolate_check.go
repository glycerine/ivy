package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/module"
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
	axioms := make([]*module.LabeledFormula, 0)
	schemaInstances := make([]*module.LabeledFormula, 0)
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
		// Stub: property checking requires analysis graph and solver
		nonTemporal := make([]*module.LabeledFormula, 0)
		for _, p := range mod.LabeledProps {
			if !p.Temporal {
				nonTemporal = append(nonTemporal, p)
			}
		}
		var checkers []Checker
		for _, prop := range nonTemporal {
			if prop.Assumed || subgoalMap[prop.ID] {
				checkers = append(checkers, NewConjAssumer(prop))
			} else {
				checkers = append(checkers, NewConjChecker(prop, 8))
			}
		}
		if len(checkers) > 0 {
			CheckFcsInState(mod, checkers)
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

	// Check initialization establishes invariant
	if len(checkedInvariants) > 0 && CheckedAction.GetString() == "" && check {
		fmt.Println("\n    Initialization must establish the invariant")
		// Stub: requires analysis graph initialization
	}

	// Check that external actions preserve the invariant
	checkedActions := setFromSlice(GetCheckedActions(mod))
	prioritized := setFromSlice(GetPrioritizedActions())
	prioritizedChecked := intersect(prioritized, checkedActions)
	actionOrder := sortedUnion(prioritizedChecked, checkedActions)

	if len(checkedActions) > 0 && len(checkedInvariants) > 0 {
		fmt.Println("\n    The following set of external actions must preserve the invariant:")
		for _, actname := range actionOrder {
			fmt.Printf("        %s\n", actname)
			if check {
				// Stub: requires analysis graph execution
			}
		}
	}

	// Check guarantees (assert actions)
	if !NoCheckGuarantees.GetBool() && check {
		// Stub: requires iterating subactions and checking asserts
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
// This corresponds to Python's check_subgoals.
func CheckSubgoals(goals []*ast.LabeledFormula, method func() error) error {
	// Stub: requires temporal models, proof goal manipulation
	for _, goal := range goals {
		_ = goal
		if method != nil {
			if err := method(); err != nil {
				Failures++
				fmt.Println("FAIL")
				return err
			}
			fmt.Println("PASS")
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

		// Stub: create_isolate and method dispatch
		if OptTrusted.GetBool() {
			continue
		}

		methodName := GetIsolateMethod(isolate, mod)
		switch {
		case methodName == "mc":
			if err := MCIsolate(isolate, mod, nil); err != nil {
				return err
			}
		case methodName == "vmt":
			if err := MCIsolate(isolate, mod, nil); err != nil {
				return err
			}
		case strings.HasPrefix(methodName, "bmc["):
			if err := MCIsolate(isolate, mod, nil); err != nil {
				return err
			}
		default:
			if err := CheckIsolate(mod, nil); err != nil {
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
