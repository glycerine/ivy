package goivy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// GetCallsMods returns the set of action names called and symbol names
// modified by an action (transitively through called actions).
//
// This walks the action tree, collecting CallAction callee names and
// symbols modified by assignment/havoc/set actions.
func GetCallsMods(action ActionsAction) (calls []string, mods []string) {
	callSet := make(map[string]bool)
	modSet := make(map[string]bool)

	for _, sub := range action.IterSubactions() {
		switch a := sub.(type) {
		case *LogicCallAction:
			name := a.CalleeName()
			callSet[CanonAct(name)] = true
		case *LogicAssignAction:
			if c, ok := a.LHS.(*Const); ok {
				modSet[c.Name] = true
			}
		case *LogicHavocAction:
			if c, ok := a.Target.(*Const); ok {
				modSet[c.Name] = true
			}
		case *LogicSetAction:
			if c, ok := a.Lit.(*Const); ok {
				modSet[c.Name] = true
			}
		}
	}

	calls = make([]string, 0, len(callSet))
	for c := range callSet {
		calls = append(calls, c)
	}
	sort.Strings(calls)

	mods = make([]string, 0, len(modSet))
	for m := range modSet {
		mods = append(mods, m)
	}
	sort.Strings(mods)

	return calls, mods
}

// whileHasRanking checks if a WhileAction has a Ranking (decreases clause).
// In Go, Ranking is stored as a RankingWrapper in the last Invariants slot.
func whileHasRanking(w *LogicWhileAction) bool {
	if len(w.Invariants) == 0 {
		return false
	}
	lastInv := w.Invariants[len(w.Invariants)-1]
	// Ranking is stored as a RankingWrapper in the last Invariants slot.
	_, isRanking := lastInv.(*RankingWrapper)
	return isRanking
}

// GetCallsModsRec recursively computes calls and mods for an action name,
// following through the action map and mixins.
//
// summarizedActions is the set of opaque actions that should be skipped.
// calls and mods are accumulated maps (actionName -> set of names).
// loops collects WhileActions without Ranking (decreases) clauses per action.
// Pass nil for loops if loop detection is not needed.
func GetCallsModsRec(
	mod *Module,
	summarizedActions map[string]bool,
	actname string,
	calls, mods map[string]map[string]bool,
	loops ...map[string][]ActionsAction,
) {
	GetCallsModsRecFull(mod, summarizedActions, actname, calls, mods, nil, nil, loops...)
}

// GetCallsModsRecFull is the full version that also tracks mixin dependencies.
// Python: get_calls_mods(mod, summarized_actions, actname, calls, mods, mixins, loops, interf_syms)
// The mixins parameter tracks which callee actions are reached via mixin chains.
// interfSyms filters modifications at collection time (matching Python behavior).
func GetCallsModsRecFull(
	mod *Module,
	summarizedActions map[string]bool,
	actname string,
	calls, mods map[string]map[string]bool,
	mixins map[string]map[string]bool,
	interfSyms map[string]bool,
	loops ...map[string][]ActionsAction,
) {
	if _, done := calls[actname]; done {
		return
	}
	if !summarizedActions[actname] {
		return
	}

	action, ok := mod.Actions.Get2(actname)
	if !ok {
		return
	}

	acalls := make(map[string]bool)
	amods := make(map[string]bool)
	amixins := make(map[string]bool)
	calls[actname] = acalls
	mods[actname] = amods
	if mixins != nil {
		mixins[actname] = amixins
	}
	actCfg := &ActionsConfig{Context: NewActionContext(mod)}

	// Optional loop tracking
	var loopMap map[string][]ActionsAction
	if len(loops) > 0 && loops[0] != nil {
		loopMap = loops[0]
	}

	for _, sub := range action.IterSubactions() {
		// Collect modifications via sub.modifies() — matching Python exactly.
		// Python: for sym in sub.modifies(): if sym in interf_syms: amods.add(sym)
		for _, sym := range ModifiesSingle(sub, actCfg) {
			xtracer.Trace("isolate.GetCallsModsRecFull mod actname=%s sym=%s type=%s",
				actname, sym.Name, ActionTypeName(sub))
			if interfSyms == nil || interfSyms[sym.Name] {
				amods[sym.Name] = true
			}
		}

		// Python line 509: Detect WhileAction without Ranking.
		if wa, ok := sub.(*LogicWhileAction); ok && loopMap != nil {
			if !whileHasRanking(wa) {
				loopMap[actname] = append(loopMap[actname], wa)
			}
		}

		// Collect calls and recurse.
		if ca, ok := sub.(*LogicCallAction); ok {
			calledName := CanonAct(ca.CalleeName())
			if !summarizedActions[calledName] {
				acalls[calledName] = true
			}
			GetCallsModsRecFull(mod, summarizedActions, calledName, calls, mods, mixins, interfSyms, loops...)
			if subcalls, ok := calls[calledName]; ok {
				for c := range subcalls {
					acalls[c] = true
				}
			}
			// Python line 506: mixins of callees count as callees
			if mixins != nil {
				if submixins, ok := mixins[calledName]; ok {
					for c := range submixins {
						acalls[c] = true
					}
				}
			}
			if submods, ok := mods[calledName]; ok {
				for m := range submods {
					amods[m] = true
				}
			}
			if loopMap != nil {
				if calledLoops, ok := loopMap[calledName]; ok {
					loopMap[actname] = append(loopMap[actname], calledLoops...)
				}
			}
		}
	}

	// Python lines 511-520: Process mixins for this action.
	if mod.Mixins != nil {
		for _, mixin := range mixinsAutoVivify(mod, actname) {
			calledName := mixin.Mixer()
			if !summarizedActions[calledName] {
				if mixins != nil {
					amixins[calledName] = true
				}
			}
			GetCallsModsRecFull(mod, summarizedActions, calledName, calls, mods, mixins, interfSyms, loops...)
			if subcalls, ok := calls[calledName]; ok {
				for c := range subcalls {
					acalls[c] = true
				}
			}
			// Python line 518: mixins of mixins count as mixins
			if mixins != nil {
				if submixins, ok := mixins[calledName]; ok {
					for c := range submixins {
						amixins[c] = true
					}
				}
			}
			if submods, ok := mods[calledName]; ok {
				for m := range submods {
					amods[m] = true
				}
			}
			if loopMap != nil {
				if calledLoops, ok := loopMap[calledName]; ok {
					loopMap[actname] = append(loopMap[actname], calledLoops...)
				}
			}
		}
	}
}

// Note: GetLocMods is defined in helpers.go.

// HasSideEffect checks if an action modifies any state symbol in the module
// signature, or contains assert actions or impure native actions.
// The check follows through calls recursively.
func HasSideEffect(mod *Module, actname string, actionMap *InsMap[string, ActionsAction]) bool {
	return HasSideEffectRec(mod, actionMap, actname, make(map[string]bool))
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
//  1. Non-summarized actions calling summarized actions don't have
//     visible modifications.
//  2. Exported summarized actions don't modify visible symbols.
//  3. There are no interfering callbacks.
//
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
func CheckInterference(mod *Module, newActions *InsMap[string, ActionsAction],
	summarizedActions map[string]bool) error {
	return CheckInterferenceFull(mod, newActions, summarizedActions,
		nil, false, nil, nil, nil)
}

// CheckInterferenceFull is the full-featured version of CheckInterference.
// Python: check_interference (lines 577-641).
func CheckInterferenceFull(mod *Module, newActions *InsMap[string, ActionsAction],
	summarizedActions map[string]bool,
	implMixins *InsMap[string, []MixinDef],
	checkTerm bool,
	interfSymsInsMap *InsMap[NodeKey, Expr],
	afterInits []string,
	allAfterInits map[string]bool,
) error {
	if !mod.Cfg.IsolateCfg.DoCheckInterference {
		return nil
	}

	// Convert InsMap to name-based set (collapses entries with same name but different sorts)
	var interfSyms map[string]bool
	if interfSymsInsMap != nil {
		interfSyms = make(map[string]bool, interfSymsInsMap.Len())
		for _, v := range interfSymsInsMap.All() {
			if c, ok := v.(*Const); ok {
				interfSyms[c.Name] = true
			}
		}
	}

	// Use InsMap length for trace to match Python (which counts symbol objects, not unique names)
	interfSymsLen := 0
	if interfSymsInsMap != nil {
		interfSymsLen = interfSymsInsMap.Len()
	}
	xtracer.Trace("isolate.CheckInterferenceFull ENTER n_summarized=%d n_interfSyms=%d n_afterInits=%d",
		len(summarizedActions), interfSymsLen, len(afterInits))
	xtracer.Trace("isolate.CheckInterferenceFull summarized=%s", strings.Join(isolateSortedKeys(summarizedActions), ","))
	// Trace interfSyms using PrettyFmla representation to match Python's str(x)
	{
		var symStrs []string
		if interfSymsInsMap != nil {
			for _, v := range interfSymsInsMap.All() {
				symStrs = append(symStrs, PrettyFmla(v))
			}
		}
		sort.Strings(symStrs)
		xtracer.Trace("isolate.CheckInterferenceFull interfSyms=%s", strings.Join(symStrs, ","))
	}

	// Compute calls, mods, mixins, and loops for all summarized actions.
	calls := make(map[string]map[string]bool)
	mods := make(map[string]map[string]bool)
	mixinDeps := make(map[string]map[string]bool)
	loops := make(map[string][]ActionsAction)
	locmods := make(map[string]map[string]bool) // Python line 582
	// Sort summarizedActions keys to get deterministic iteration order matching Python
	sortedSummarized := isolateSortedKeys(summarizedActions)
	for _, actname := range sortedSummarized {
		GetCallsModsRecFull(mod, summarizedActions, actname, calls, mods, mixinDeps, interfSyms, loops)
		// Python line 586: locmods[actname] = get_loc_mods(mod, actname)
		locmods[actname] = make(map[string]bool)
		for _, s := range GetLocMods(mod, actname) {
			locmods[actname][s] = true
		}
		xtracer.Trace("isolate.GetCallsModsRecFull actname=%s calls=%s mods=%s",
			actname, strings.Join(isolateSortedKeys(calls[actname]), ","), strings.Join(isolateSortedKeys(mods[actname]), ","))
		xtracer.Trace("isolate.GetLocMods actname=%s locmods=%s",
			actname, strings.Join(isolateSortedKeys(locmods[actname]), ","))
	}

	// Python line 587: compute callouts for all actions
	callouts := make(map[string]Callouts)
	for actname := range newActions.All() {
		GetCallouts(mod, newActions, summarizedActions, actname, callouts)
	}

	// interfSyms filtering now happens at collection time inside GetCallsModsRecFull,
	// matching Python's inline `if sym in interf_syms` check.

	// Get all mixins for impl_mixins lookup
	if implMixins == nil {
		implMixins = NewInsMap[string, []MixinDef]()
	}

	// Python lines 590-622: For each non-summarized action, check interference.
	for actname, action := range newActions.All() {
		if summarizedActions[actname] {
			continue
		}
		xtracer.Trace("isolate.CheckInterferenceFull non_summarized actname=%s", actname)
		for _, sub := range action.IterSubactions() {
			ca, ok := sub.(*LogicCallAction)
			if !ok {
				continue
			}
			calledName := CanonAct(ca.CalleeName())
			xtracer.Trace("isolate.CheckInterferenceFull call_check actname=%s calledName=%s", actname, calledName)

			// Python lines 594-597: Compute pre_refed — symbols used in
			// unsummarized before-mixins of the called action.
			preRefed := make(map[string]bool)
			if mod.Mixins != nil {
				for _, m := range mixinsAutoVivify(mod, calledName) {
					if !m.IsAfter() && !summarizedActions[m.Mixer()] {
						if mixerAct, ok := mod.Actions.Get2(m.Mixer()); ok {
							collectActionSymbolNames(mixerAct, preRefed)
						}
					}
				}
			}

			// Build list of all related actions: callee + mixins + impl_mixins
			allCalls := []string{calledName}
			for _, m := range mixinsAutoVivify(mod, calledName) {
				allCalls = append(allCalls, m.Mixer())
			}
			if ims, ok := implMixins.Get2(calledName); ok {
				for _, m := range ims {
					allCalls = append(allCalls, m.Mixer())
				}
			}

			for _, called := range allCalls {
				if !summarizedActions[called] {
					continue
				}
				// Python lines 602-605: cmods = set(mods[called])
				// then add locmods that are in pre_refed
				cmods := make(map[string]bool)
				if m, ok := mods[called]; ok {
					for sym := range m {
						cmods[sym] = true
					}
				}
				if lm, ok := locmods[called]; ok {
					for loc := range lm {
						if preRefed[loc] {
							cmods[loc] = true
						}
					}
				}
				if len(cmods) > 0 {
					modNames := make([]string, 0, len(cmods))
					for m := range cmods {
						modNames = append(modNames, m)
					}
					sort.Strings(modNames)
					xtracer.Trace("isolate.CheckInterferenceFull ERROR_CALLOUT actname=%s called=%s cmods=%s",
						actname, called, strings.Join(modNames, ","))
					refs := FindReferences(mod, cmods, newActions)
					refStr := ""
					for ln := range refs {
						refStr += fmt.Sprintf("\n%d referenced here", ln)
					}
					return fmt.Errorf("call out to %s may have visible effect on %s%s",
						called, joinStrings(modNames, ","), refStr)
				}

				// Python lines 611-615: Check termination per called action.
				if checkTerm {
					if calledLoops, ok := loops[called]; ok && len(calledLoops) > 0 {
						return fmt.Errorf("call out to %s may not terminate (needs a decreases clause)", called)
					}
				}
			}
		}

		// Python lines 616-622: Check for interfering callbacks via callouts.
		if co, ok := callouts[actname]; ok {
			midcalls := co[0] // index 0 = middle position (not head, not tail)
			for midcall := range midcalls {
				if mcalls, ok := calls[midcall]; ok && len(mcalls) > 0 {
					if mmods, ok := mods[midcall]; ok && len(mmods) > 0 {
						callbackNames := make([]string, 0, len(mcalls))
						for c := range mcalls {
							callbackNames = append(callbackNames, c)
						}
						sort.Strings(callbackNames)
						return fmt.Errorf("call to %s may cause interfering callback to %s",
							midcall, joinStrings(callbackNames, ","))
					}
				}
			}
		}
	}

	// Also do a standalone termination check over all loops for summarized actions,
	// catching cases where a summarized action has loops but isn't called by
	// any non-summarized action.
	if checkTerm {
		for actname, actLoops := range loops {
			if len(actLoops) > 0 {
				return fmt.Errorf("action %s contains a loop without a decreases clause", actname)
			}
		}
	}

	// Collect symbols referenced by after-init actions.
	afterInitRefs := make(map[string]bool)
	for _, aiName := range afterInits {
		if act, ok := newActions.Get2(aiName); ok {
			collectActionSymbolNames(act, afterInitRefs)
		}
	}

	// Check exported summarized actions.
	for _, exp := range mod.Exports {
		calledName := CanonAct(exp.Exported())

		allCalls := []string{calledName}
		for _, m := range mixinsAutoVivify(mod, calledName) {
			allCalls = append(allCalls, m.Mixer())
		}
		if ims, ok := implMixins.Get2(calledName); ok {
			for _, m := range ims {
				allCalls = append(allCalls, m.Mixer())
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
					sort.Strings(modNames)
					refs := FindReferences(mod, filteredMods, newActions)
					refStr := ""
					for ln := range refs {
						refStr += fmt.Sprintf("\n%d referenced here", ln)
					}
					return fmt.Errorf("external call to %s may have visible effect on %s%s",
						called, joinStrings(modNames, ","), refStr)
				}
			}
		}
	}

	return nil
}

// collectActionSymbolNames collects all symbol names referenced by an action.
func collectActionSymbolNames(action ActionsAction, names map[string]bool) {
	for _, arg := range action.ActionArgs() {
		collectNodeSymNames(arg, names)
	}
}

func collectNodeSymNames(node Expr, names map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*Const); ok {
		names[c.Name] = true
	}
	if act, ok := node.(ActionsAction); ok {
		collectActionSymbolNames(act, names)
		return
	}
	if app, ok := node.(*Apply); ok {
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
func ConeOfInfluenceFilter(mod *Module, goals []*LabeledFormula) error {
	isoCfg := mod.Cfg.IsolateCfg
	if !isoCfg.ConeOfInfluence {
		return nil
	}

	// Collect all symbols used in goals, axioms, props, inits, and definitions.
	allSyms := make(map[string]bool)

	// Helper to collect symbol names from a node tree.
	var collectSymbols func(Expr)
	collectSymbols = func(node Expr) {
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *Const:
			allSyms[n.Name] = true
		case *LogicVariable:
			// Variables are not module symbols.
		case *Apply:
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
	for _, lfSlice := range [][]*LabeledFormula{
		goals,
		mod.LabeledAxioms,
		mod.LabeledProps,
		mod.LabeledInits,
		mod.LabeledConjs,
		mod.Definitions,
	} {
		for _, lf := range lfSlice {
			collectSymbols(lf.Formula.(Expr))
		}
	}

	// Collect from action formal parameters and bodies.
	for _, act := range mod.Actions.All() {
		for _, p := range act.GetFormalParams() {
			allSyms[p.Name] = true
		}
		for _, r := range act.GetFormalReturns() {
			allSyms[r.Name] = true
		}
		// Collect symbols from the action body.
		for _, sub := range act.IterSubactions() {
			for _, arg := range sub.ActionArgs() {
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
			if apply, ok := dfn.Formula.(*Apply); ok && len(apply.Terms) > 0 {
				if c, ok := apply.Func.(*Const); ok {
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
	if isoCfg.FilterSymbols && mod.Sig != nil {
		for name := range mod.Sig.Symbols.All() {
			if !allSyms[name] {
				mod.Sig.Symbols.Delkey(name)
			}
		}
	}

	// Filter sorts: collect all sorts referenced by relevant symbols,
	// then remove unreferenced sorts.
	if isoCfg.FilterSymbols && mod.Sig != nil {
		allSorts := make(map[string]bool)
		allSorts["bool"] = true // bool is always kept

		addSortDeps := func(s Sort) {
			addSortName(s, allSorts)
		}

		for _, entry := range mod.Sig.Symbols.All() {
			if entry.Union != nil {
				for _, s := range entry.Union.Sorts {
					addSortDeps(s)
				}
			} else {
				addSortDeps(entry.Sort)
			}
		}

		// Remove sorts not in the relevant set.
		var sortKeysToDelete []string
		for name, _ := range mod.Sig.Sorts.All() {
			if !allSorts[name] {
				sortKeysToDelete = append(sortKeysToDelete, name)
			}
		}
		for _, name := range sortKeysToDelete {
			mod.Sig.Sorts.Delkey(name)
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
		for name := range mod.SortDestructors.All() {
			if !allSorts[name] {
				mod.SortDestructors.Delkey(name)
			}
		}

		// Filter destructor sorts.
		for name, s := range mod.DestructorSorts {
			sName := isolateSortToName(s)
			if !allSorts[sName] {
				delete(mod.DestructorSorts, name)
			}
		}
	}

	return nil
}

// addSortName recursively adds sort names from a sort to the set.
func addSortName(s Sort, set map[string]bool) {
	switch t := s.(type) {
	case *UninterpretedSort:
		set[t.Name] = true
	case *LogicEnumeratedSort:
		set[t.Name] = true
	case *RangeSort:
		set[t.Name] = true
	case *BooleanSort:
		set["bool"] = true
	case *LogicFunctionSort:
		for _, d := range t.Domain() {
			addSortName(d, set)
		}
		addSortName(t.Range(), set)
	}
}

// sortToName extracts the name from a sort.
func isolateSortToName(s Sort) string {
	switch t := s.(type) {
	case *UninterpretedSort:
		return t.Name
	case *LogicEnumeratedSort:
		return t.Name
	case *RangeSort:
		return t.Name
	case *BooleanSort:
		return "bool"
	default:
		return s.String()
	}
}

// collectRelevantDestructorsForSym mirrors Python collect_relevant_destructors.
// It extracts the range sort from the symbol's sort and collects destructors for it.
func collectRelevantDestructorsForSym(mod *Module, sym Expr, result map[string]bool, memo map[string]bool) {
	c, ok := sym.(*Const)
	if !ok || c.CSort == nil {
		return
	}
	fs, ok := c.CSort.(*LogicFunctionSort)
	if !ok {
		return // sym.sort has no .rng — matches Python's hasattr(sym.sort, 'rng') check
	}
	rng := fs.Range()
	if us, ok := rng.(*UninterpretedSort); ok {
		CollectSortDestructors(mod, us.Name, result, memo)
	}
}

// CollectSortDestructors collects all destructor symbols for a sort
// and its variants, recursively.
func CollectSortDestructors(mod *Module, sortName string, result map[string]bool, memo map[string]bool) {
	if memo[sortName] {
		return
	}
	memo[sortName] = true

	// Add destructors for this sort.
	if destrs, ok := mod.SortDestructors.Get2(sortName); ok {
		for _, d := range destrs {
			result[d.Name] = true
			// Recursively collect destructors of the range sort.
			if fs, ok := d.CSort.(*LogicFunctionSort); ok {
				rng := fs.Range()
				if us, ok := rng.(*UninterpretedSort); ok {
					CollectSortDestructors(mod, us.Name, result, memo)
				}
			}
		}
	}

	// If this sort has variants, collect destructors for each variant.
	if variants, ok := mod.Variants[sortName]; ok {
		for _, v := range variants {
			if us, ok := v.(*UninterpretedSort); ok {
				CollectSortDestructors(mod, us.Name, result, memo)
			}
		}
	}
}

// ActionCallGraph builds a REVERSE call graph: for each action name that
// is called, maps it to the list of action names that call it.
// Matches Python's Module.call_graph() (ivy_module.py:282-287).
func ActionCallGraph(mod *Module) map[string][]string {
	graph := make(map[string][]string)
	for actname, act := range mod.Actions.All() {
		for _, sub := range act.IterSubactions() {
			if ca, ok := sub.(*LogicCallAction); ok {
				calledName := CanonAct(ca.CalleeName())
				graph[calledName] = append(graph[calledName], actname)
			}
		}
	}
	// Sort each caller list for determinism.
	for k := range graph {
		sort.Strings(graph[k])
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

// sortedKeys returns sorted keys from a map[string]bool.
func isolateSortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
