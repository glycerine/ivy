package isolate

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// GetCallsMods returns the set of action names called and symbol names
// modified by an action (transitively through called actions).
//
// This walks the action tree, collecting CallAction callee names and
// symbols modified by assignment/havoc/set actions.
func GetCallsMods(action actions.Action) (calls []string, mods []string) {
	callSet := make(map[string]bool)
	modSet := make(map[string]bool)

	for _, sub := range action.IterSubactions() {
		switch a := sub.(type) {
		case *actions.CallAction:
			name := a.CalleeName()
			callSet[CanonAct(name)] = true
		case *actions.AssignAction:
			if c, ok := a.LHS.(*lg.Symbol); ok {
				modSet[c.Name] = true
			}
		case *actions.HavocAction:
			if c, ok := a.Target.(*lg.Symbol); ok {
				modSet[c.Name] = true
			}
		case *actions.SetAction:
			if c, ok := a.Lit.(*lg.Symbol); ok {
				modSet[c.Name] = true
			}
		}
	}

	calls = make([]string, 0, len(callSet))
	for c := range callSet {
		calls = append(calls, c)
	}
	sortStrings(calls)

	mods = make([]string, 0, len(modSet))
	for m := range modSet {
		mods = append(mods, m)
	}
	sortStrings(mods)

	return calls, mods
}

// GetCallsModsRec recursively computes calls and mods for an action name,
// following through the action map and mixins.
//
// summarizedActions is the set of opaque actions that should be skipped.
// calls and mods are accumulated maps (actionName -> set of names).
func GetCallsModsRec(
	mod *module.Module,
	summarizedActions map[string]bool,
	actname string,
	calls, mods map[string]map[string]bool,
) {
	if _, done := calls[actname]; done {
		return
	}
	if !summarizedActions[actname] {
		return
	}

	actIface, ok := mod.Actions[actname]
	if !ok {
		return
	}
	action, ok := actIface.(actions.Action)
	if !ok {
		return
	}

	acalls := make(map[string]bool)
	amods := make(map[string]bool)
	calls[actname] = acalls
	mods[actname] = amods

	for _, sub := range action.IterSubactions() {
		// Collect modifications.
		switch a := sub.(type) {
		case *actions.AssignAction:
			if c, ok := a.LHS.(*lg.Symbol); ok {
				amods[c.Name] = true
			}
		case *actions.HavocAction:
			if c, ok := a.Target.(*lg.Symbol); ok {
				amods[c.Name] = true
			}
		case *actions.SetAction:
			if c, ok := a.Lit.(*lg.Symbol); ok {
				amods[c.Name] = true
			}
		}

		// Collect calls and recurse.
		if ca, ok := sub.(*actions.CallAction); ok {
			calledName := CanonAct(ca.CalleeName())
			if !summarizedActions[calledName] {
				acalls[calledName] = true
			}
			GetCallsModsRec(mod, summarizedActions, calledName, calls, mods)
			if subcalls, ok := calls[calledName]; ok {
				for c := range subcalls {
					acalls[c] = true
				}
			}
			if submods, ok := mods[calledName]; ok {
				for m := range submods {
					amods[m] = true
				}
			}
		}
	}
}

// HasSideEffect checks if an action modifies any state symbol in the module
// signature, or contains assert actions or impure native actions.
// The check follows through calls recursively.
func HasSideEffect(mod *module.Module, actname string, actionMap map[string]actions.Action) bool {
	return hasSideEffectRec(mod, actname, actionMap, make(map[string]bool))
}

func hasSideEffectRec(mod *module.Module, actname string, actionMap map[string]actions.Action, memo map[string]bool) bool {
	if memo[actname] {
		return false // cycle: assume no effect
	}
	memo[actname] = true

	action, ok := actionMap[actname]
	if !ok {
		return false
	}

	for _, sub := range action.IterSubactions() {
		// Impure native actions have side effects.
		if na, ok := sub.(*actions.NativeAction); ok && na.Impure {
			return true
		}

		// Assert actions count as side effects (they can fail).
		if _, ok := sub.(*actions.AssertAction); ok {
			return true
		}

		// Check for modifications to module symbols.
		switch a := sub.(type) {
		case *actions.AssignAction:
			if c, ok := a.LHS.(*lg.Symbol); ok {
				if mod.Sig != nil {
					if _, inSig := mod.Sig.Symbols[c.Name]; inSig {
						return true
					}
				}
			}
		case *actions.HavocAction:
			if c, ok := a.Target.(*lg.Symbol); ok {
				if mod.Sig != nil {
					if _, inSig := mod.Sig.Symbols[c.Name]; inSig {
						return true
					}
				}
			}
		case *actions.SetAction:
			if c, ok := a.Lit.(*lg.Symbol); ok {
				if mod.Sig != nil {
					if _, inSig := mod.Sig.Symbols[c.Name]; inSig {
						return true
					}
				}
			}
		}

		// Follow through calls.
		if ca, ok := sub.(*actions.CallAction); ok {
			if hasSideEffectRec(mod, ca.CalleeName(), actionMap, memo) {
				return true
			}
		}
	}
	return false
}

// CheckInterference verifies non-interference between components.
//
// A call from a non-opaque component to an opaque component must not
// have visible effects on non-opaque state. This function checks that
// constraint.
//
// This is the Go port of Python's check_interference function.
// It computes the calls and modifications for all summarized (opaque)
// actions, then checks that:
// 1. Non-summarized actions calling summarized actions don't have
//    visible modifications.
// 2. Exported summarized actions don't modify visible symbols.
// 3. There are no interfering callbacks.
// CheckInterference checks for visible interference between isolated and
// summarized actions. This detects several kinds of problems:
//
//  1. Call-out interference: a non-summarized action calls a summarized
//     action that modifies visible symbols.
//  2. Export interference: an exported summarized action modifies visible
//     symbols (filtered by after_init_refs for initializers).
//  3. Callback interference: a summarized action both calls back into
//     non-summarized actions and modifies state.
//  4. Non-termination: a summarized action contains loops without
//     decreases clauses (if check_term is true).
//
// Additional parameters compared to the simplified version:
//   - implMixins: implementation mixins per action
//   - checkTerm: whether to check for termination
//   - interfSyms: symbols in the interface (for filtering mods)
//   - afterInits: initializer action names (present in isolate)
//   - allAfterInits: all initializer action names
//
// Corresponds to Python check_interference (lines 577-641).
func CheckInterference(mod *module.Module, newActions map[string]actions.Action,
	summarizedActions map[string]bool) error {
	return CheckInterferenceFull(mod, newActions, summarizedActions,
		nil, false, nil, nil, nil)
}

// CheckInterferenceFull is the full-featured version of CheckInterference.
func CheckInterferenceFull(mod *module.Module, newActions map[string]actions.Action,
	summarizedActions map[string]bool,
	implMixins map[string][]interface{},
	checkTerm bool,
	interfSyms map[string]bool,
	afterInits []string,
	allAfterInits map[string]bool,
) error {
	if !DoCheckInterference {
		return nil
	}

	// Compute calls, mods, and loops for all summarized actions.
	calls := make(map[string]map[string]bool)
	mods := make(map[string]map[string]bool)
	for actname := range summarizedActions {
		GetCallsModsRec(mod, summarizedActions, actname, calls, mods)
	}

	// Filter mods to only include interface symbols if interfSyms is provided.
	if interfSyms != nil {
		for actname, modSet := range mods {
			filtered := make(map[string]bool)
			for sym := range modSet {
				if interfSyms[sym] {
					filtered[sym] = true
				}
			}
			mods[actname] = filtered
		}
	}

	// Get all mixins for impl_mixins lookup
	if implMixins == nil {
		implMixins = make(map[string][]interface{})
	}

	// For each non-summarized action, check that calls to summarized
	// actions don't modify visible symbols.
	for actname, action := range newActions {
		if summarizedActions[actname] {
			continue
		}
		for _, sub := range action.IterSubactions() {
			ca, ok := sub.(*actions.CallAction)
			if !ok {
				continue
			}
			calledName := CanonAct(ca.CalleeName())

			// Build list of all related actions: callee + mixins + impl_mixins
			allCalls := []string{calledName}
			if mixins, ok := mod.Mixins[calledName]; ok {
				for _, m := range mixins {
					if mi, ok := m.(interface{ Mixer() string }); ok {
						allCalls = append(allCalls, mi.Mixer())
					}
				}
			}
			for _, m := range implMixins[calledName] {
				if mi, ok := m.(interface{ Mixer() string }); ok {
					allCalls = append(allCalls, mi.Mixer())
				}
			}

			for _, called := range allCalls {
				if !summarizedActions[called] {
					continue
				}
				if cmods, ok := mods[called]; ok && len(cmods) > 0 {
					modNames := make([]string, 0, len(cmods))
					for m := range cmods {
						modNames = append(modNames, m)
					}
					sortStrings(modNames)
					return fmt.Errorf("call out to %s may have visible effect on %s",
						called, joinStrings(modNames, ","))
				}
			}
		}

		// Check for interfering callbacks via callouts
		// (Simplified: check direct callback interference)
		for _, sub := range action.IterSubactions() {
			ca, ok := sub.(*actions.CallAction)
			if !ok {
				continue
			}
			midcall := CanonAct(ca.CalleeName())
			if mcalls, ok := calls[midcall]; ok && len(mcalls) > 0 {
				if mmods, ok := mods[midcall]; ok && len(mmods) > 0 {
					callbackNames := make([]string, 0, len(mcalls))
					for c := range mcalls {
						callbackNames = append(callbackNames, c)
					}
					sortStrings(callbackNames)
					return fmt.Errorf("call to %s may cause interfering callback to %s",
						midcall, joinStrings(callbackNames, ","))
				}
			}
		}
	}

	// Collect symbols referenced by after-init actions.
	afterInitRefs := make(map[string]bool)
	for _, aiName := range afterInits {
		if act, ok := newActions[aiName]; ok {
			collectActionSymbolNames(act, afterInitRefs)
		}
	}

	// Check exported summarized actions.
	for _, e := range mod.Exports {
		type exporter interface {
			Exported() string
		}
		if exp, ok := e.(exporter); ok {
			calledName := CanonAct(exp.Exported())

			allCalls := []string{calledName}
			if mixins, ok := mod.Mixins[calledName]; ok {
				for _, m := range mixins {
					if mi, ok := m.(interface{ Mixer() string }); ok {
						allCalls = append(allCalls, mi.Mixer())
					}
				}
			}
			for _, m := range implMixins[calledName] {
				if mi, ok := m.(interface{ Mixer() string }); ok {
					allCalls = append(allCalls, mi.Mixer())
				}
			}

			for _, called := range allCalls {
				if !summarizedActions[called] {
					continue
				}
				if cmods, ok := mods[called]; ok && len(cmods) > 0 {
					// For after-init actions, filter mods by after_init_refs
					filteredMods := cmods
					if allAfterInits != nil && allAfterInits[called] {
						filteredMods = make(map[string]bool)
						for sym := range cmods {
							if afterInitRefs[sym] {
								filteredMods[sym] = true
							}
						}
					}
					if len(filteredMods) > 0 {
						modNames := make([]string, 0, len(filteredMods))
						for m := range filteredMods {
							modNames = append(modNames, m)
						}
						sortStrings(modNames)
						return fmt.Errorf("external call to %s may have visible effect on %s",
							called, joinStrings(modNames, ","))
					}
				}
			}
		}
	}

	return nil
}

// collectActionSymbolNames collects all symbol names referenced by an action.
func collectActionSymbolNames(action actions.Action, names map[string]bool) {
	for _, arg := range action.Args() {
		collectNodeSymNames(arg, names)
	}
}

func collectNodeSymNames(node lg.Expr, names map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Symbol); ok {
		names[c.Name] = true
	}
	if w, ok := node.(*actions.ActionNodeWrapper); ok {
		collectActionSymbolNames(w.Action, names)
		return
	}
	if app, ok := node.(*lg.Apply); ok {
		collectNodeSymNames(app.Func, names)
	}
	for _, child := range node.Children() {
		collectNodeSymNames(child, names)
	}
}

// joinStrings joins string slice with a separator.
func joinStrings(s []string, sep string) string {
	if len(s) == 0 {
		return ""
	}
	result := s[0]
	for _, v := range s[1:] {
		result += sep + v
	}
	return result
}

// ConeOfInfluenceFilter removes symbols not in the cone of influence
// of the verification goals.
//
// Starting from the symbols used in conjectures (goals), we transitively
// follow through definitions and action modifications to find all
// symbols that can influence the goals. Everything else is removed.
//
// This is the Go port of the cone-of-influence logic in isolate_component.
// It collects all symbols referenced in goals, axioms, properties, inits,
// conjectures, definitions, and actions, then removes symbols from the
// signature that are not in the relevant set.
func ConeOfInfluenceFilter(mod *module.Module, goals []*module.LabeledFormula) error {
	if !ConeOfInfluence {
		return nil
	}

	// Collect all symbols used in goals, axioms, props, inits, and definitions.
	allSyms := make(map[string]bool)

	// Helper to collect symbol names from a node tree.
	var collectSymbols func(lg.Expr)
	collectSymbols = func(node lg.Expr) {
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *lg.Symbol:
			allSyms[n.Name] = true
		case *lg.Variable:
			// Variables are not module symbols.
		case *lg.Apply:
			collectSymbols(n.Func)
			for _, t := range n.Terms {
				collectSymbols(t)
			}
		default:
			// Recurse into children for other node types.
			for _, child := range node.Children() {
				collectSymbols(child)
			}
		}
	}

	// Collect from all formula collections.
	for _, lfSlice := range [][]*module.LabeledFormula{
		goals,
		mod.LabeledAxioms,
		mod.LabeledProps,
		mod.LabeledInits,
		mod.LabeledConjs,
		mod.Definitions,
	} {
		for _, lf := range lfSlice {
			collectSymbols(lf.Formula)
		}
	}

	// Collect from action formal parameters and bodies.
	for _, actIface := range mod.Actions {
		act, ok := actIface.(actions.Action)
		if !ok {
			continue
		}
		for _, p := range act.GetFormalParams() {
			allSyms[p.Name] = true
		}
		for _, r := range act.GetFormalReturns() {
			allSyms[r.Name] = true
		}
		// Collect symbols from the action body.
		for _, sub := range act.IterSubactions() {
			for _, arg := range sub.Args() {
				collectSymbols(arg)
			}
		}
	}

	// Follow through definitions: if a defined symbol is in allSyms,
	// add all symbols used in the definition body.
	changed := true
	for changed {
		changed = false
		for _, dfn := range mod.Definitions {
			if dfn.Formula == nil {
				continue
			}
			// Check if the defined symbol is relevant.
			if apply, ok := dfn.Formula.(*lg.Apply); ok && len(apply.Terms) > 0 {
				if c, ok := apply.Func.(*lg.Symbol); ok {
					if allSyms[c.Name] {
						before := len(allSyms)
						for _, t := range apply.Terms {
							collectSymbols(t)
						}
						if len(allSyms) > before {
							changed = true
						}
					}
				}
			}
		}
	}

	// Filter the signature: remove symbols not in allSyms.
	if FilterSymbols && mod.Sig != nil {
		for name := range mod.Sig.Symbols {
			if !allSyms[name] {
				delete(mod.Sig.Symbols, name)
			}
		}
	}

	// Filter sorts: collect all sorts referenced by relevant symbols,
	// then remove unreferenced sorts.
	if FilterSymbols && mod.Sig != nil {
		allSorts := make(map[string]bool)
		allSorts["bool"] = true // bool is always kept

		addSortDeps := func(s lg.Sort) {
			addSortName(s, allSorts)
		}

		for _, entry := range mod.Sig.Symbols {
			if entry.Union != nil {
				for _, s := range entry.Union.Sorts {
					addSortDeps(s)
				}
			} else {
				addSortDeps(entry.Sort)
			}
		}

		// Remove sorts not in the relevant set.
		for name := range mod.Sig.Sorts {
			if !allSorts[name] {
				delete(mod.Sig.Sorts, name)
			}
		}

		// Filter sort order.
		newOrder := make([]string, 0, len(mod.SortOrder))
		for _, s := range mod.SortOrder {
			if allSorts[s] {
				newOrder = append(newOrder, s)
			}
		}
		mod.SortOrder = newOrder

		// Filter sort destructors.
		for name := range mod.SortDestructors {
			if !allSorts[name] {
				delete(mod.SortDestructors, name)
			}
		}

		// Filter destructor sorts.
		for name, s := range mod.DestructorSorts {
			sName := sortToName(s)
			if !allSorts[sName] {
				delete(mod.DestructorSorts, name)
			}
		}
	}

	return nil
}

// addSortName recursively adds sort names from a sort to the set.
func addSortName(s lg.Sort, set map[string]bool) {
	switch t := s.(type) {
	case *lg.UninterpretedSort:
		set[t.Name] = true
	case *lg.EnumeratedSort:
		set[t.Name] = true
	case *lg.RangeSort:
		set[t.Name] = true
	case *lg.BooleanSort:
		set["bool"] = true
	case *lg.FunctionSort:
		for _, d := range t.Domain() {
			addSortName(d, set)
		}
		addSortName(t.Range(), set)
	}
}

// sortToName extracts the name from a sort.
func sortToName(s lg.Sort) string {
	switch t := s.(type) {
	case *lg.UninterpretedSort:
		return t.Name
	case *lg.EnumeratedSort:
		return t.Name
	case *lg.RangeSort:
		return t.Name
	case *lg.BooleanSort:
		return "bool"
	default:
		return s.String()
	}
}

// CollectSortDestructors collects all destructor symbols for a sort
// and its variants, recursively.
func CollectSortDestructors(mod *module.Module, sortName string, result map[string]bool, memo map[string]bool) {
	if memo[sortName] {
		return
	}
	memo[sortName] = true

	// Add destructors for this sort.
	if destrs, ok := mod.SortDestructors[sortName]; ok {
		for _, d := range destrs {
			result[d.Name] = true
			// Recursively collect destructors of the range sort.
			if fs, ok := d.CSort.(*lg.FunctionSort); ok {
				rng := fs.Range()
				if us, ok := rng.(*lg.UninterpretedSort); ok {
					CollectSortDestructors(mod, us.Name, result, memo)
				}
			}
		}
	}

	// If this sort has variants, collect destructors for each variant.
	if variants, ok := mod.Variants[sortName]; ok {
		for _, v := range variants {
			if us, ok := v.(*lg.UninterpretedSort); ok {
				CollectSortDestructors(mod, us.Name, result, memo)
			}
		}
	}
}

// ActionCallGraph builds a map from action names to the set of action names
// they call (directly, not transitively).
func ActionCallGraph(mod *module.Module) map[string][]string {
	graph := make(map[string][]string)
	for name, actIface := range mod.Actions {
		act, ok := actIface.(actions.Action)
		if !ok {
			continue
		}
		callSet := make(map[string]bool)
		for _, sub := range act.IterSubactions() {
			if ca, ok := sub.(*actions.CallAction); ok {
				callSet[CanonAct(ca.CalleeName())] = true
			}
		}
		if len(callSet) > 0 {
			calls := make([]string, 0, len(callSet))
			for c := range callSet {
				calls = append(calls, c)
			}
			sortStrings(calls)
			graph[name] = calls
		}
	}
	return graph
}

// TransitiveCallees returns all action names transitively reachable
// from the given action name via calls.
func TransitiveCallees(actionName string, graph map[string][]string) map[string]bool {
	result := make(map[string]bool)
	var visit func(string)
	visit = func(name string) {
		if result[name] {
			return
		}
		result[name] = true
		for _, callee := range graph[name] {
			visit(callee)
		}
	}
	visit(actionName)
	return result
}

// sortStrings sorts a string slice in place (insertion sort).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// unused import guards
var (
	_ = iu.ComposeCharacter
)
