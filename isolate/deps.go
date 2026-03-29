package isolate

import (
	"fmt"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
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
			if c, ok := a.LHS.(*lg.Const); ok {
				modSet[c.Name] = true
			}
		case *actions.HavocAction:
			if c, ok := a.Target.(*lg.Const); ok {
				modSet[c.Name] = true
			}
		case *actions.SetAction:
			if c, ok := a.Lit.(*lg.Const); ok {
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

// whileHasRanking checks if a WhileAction has a Ranking (decreases clause).
// In Go, Ranking is stored as a RankingWrapper in the last Invariants slot.
func whileHasRanking(w *actions.WhileAction) bool {
	if len(w.Invariants) == 0 {
		return false
	}
	lastInv := w.Invariants[len(w.Invariants)-1]
	// Ranking is stored as a RankingWrapper in the last Invariants slot.
	_, isRanking := lastInv.(*actions.RankingWrapper)
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
	mod *module.Module,
	summarizedActions map[string]bool,
	actname string,
	calls, mods map[string]map[string]bool,
	loops ...map[string][]actions.Action,
) {
	GetCallsModsRecFull(mod, summarizedActions, actname, calls, mods, nil, loops...)
}

// GetCallsModsRecFull is the full version that also tracks mixin dependencies.
// Python: get_calls_mods(mod, summarized_actions, actname, calls, mods, mixins, loops, interf_syms)
// The mixins parameter tracks which callee actions are reached via mixin chains.
func GetCallsModsRecFull(
	mod *module.Module,
	summarizedActions map[string]bool,
	actname string,
	calls, mods map[string]map[string]bool,
	mixins map[string]map[string]bool,
	loops ...map[string][]actions.Action,
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

	// Optional loop tracking
	var loopMap map[string][]actions.Action
	if len(loops) > 0 && loops[0] != nil {
		loopMap = loops[0]
	}

	for _, sub := range action.IterSubactions() {
		// Collect modifications.
		switch a := sub.(type) {
		case *actions.AssignAction:
			if c, ok := a.LHS.(*lg.Const); ok {
				amods[c.Name] = true
			}
		case *actions.HavocAction:
			if c, ok := a.Target.(*lg.Const); ok {
				amods[c.Name] = true
			}
		case *actions.SetAction:
			if c, ok := a.Lit.(*lg.Const); ok {
				amods[c.Name] = true
			}
		}

		// Python line 509: Detect WhileAction without Ranking.
		if wa, ok := sub.(*actions.WhileAction); ok && loopMap != nil {
			if !whileHasRanking(wa) {
				loopMap[actname] = append(loopMap[actname], wa)
			}
		}

		// Collect calls and recurse.
		if ca, ok := sub.(*actions.CallAction); ok {
			calledName := CanonAct(ca.CalleeName())
			if !summarizedActions[calledName] {
				acalls[calledName] = true
			}
			GetCallsModsRecFull(mod, summarizedActions, calledName, calls, mods, mixins, loops...)
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
		for _, mixin := range mod.Mixins[actname] {
			calledName := mixin.Mixer()
			if !summarizedActions[calledName] {
				if mixins != nil {
					amixins[calledName] = true
				}
			}
			GetCallsModsRecFull(mod, summarizedActions, calledName, calls, mods, mixins, loops...)
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
func HasSideEffect(mod *module.Module, actname string, actionMap *iu.InsMap[string, actions.Action]) bool {
	return hasSideEffectRec(mod, actname, actionMap, make(map[string]bool))
}

func hasSideEffectRec(mod *module.Module, actname string, actionMap *iu.InsMap[string, actions.Action], memo map[string]bool) bool {
	if memo[actname] {
		return false // cycle: assume no effect
	}
	memo[actname] = true

	action, ok := actionMap.Get2(actname)
	if !ok {
		return false
	}

	for _, sub := range action.IterSubactions() {
		// Impure native actions have side effects.
		if na, ok := sub.(*actions.NativeAction); ok && na.Impure {
			return true
		}

		// Assert actions count as side effects (they can fail).
		// Python: isinstance(sub, ia.AssertAction) — matches all subclasses.
		if actions.IsAssertLike(sub) {
			return true
		}

		// Python line 472-473: Ranking has side effects.
		if _, isRanking := sub.(*actions.Ranking); isRanking {
			return true
		}

		// Check for modifications to module symbols.
		switch a := sub.(type) {
		case *actions.AssignAction:
			if c, ok := a.LHS.(*lg.Const); ok {
				if mod.Sig != nil {
					if _, inSig := mod.Sig.Symbols[c.Name]; inSig {
						return true
					}
				}
			}
		case *actions.HavocAction:
			if c, ok := a.Target.(*lg.Const); ok {
				if mod.Sig != nil {
					if _, inSig := mod.Sig.Symbols[c.Name]; inSig {
						return true
					}
				}
			}
		case *actions.SetAction:
			if c, ok := a.Lit.(*lg.Const); ok {
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
func CheckInterference(mod *module.Module, newActions *iu.InsMap[string, actions.Action],
	summarizedActions map[string]bool) error {
	return CheckInterferenceFull(mod, newActions, summarizedActions,
		nil, false, nil, nil, nil)
}

// CheckInterferenceFull is the full-featured version of CheckInterference.
// Python: check_interference (lines 577-641).
func CheckInterferenceFull(mod *module.Module, newActions *iu.InsMap[string, actions.Action],
	summarizedActions map[string]bool,
	implMixins map[string][]module.MixinDef,
	checkTerm bool,
	interfSyms map[string]bool,
	afterInits []string,
	allAfterInits map[string]bool,
) error {
	if !mod.Cfg.IsolateCfg.DoCheckInterference {
		return nil
	}

	// Compute calls, mods, mixins, and loops for all summarized actions.
	calls := make(map[string]map[string]bool)
	mods := make(map[string]map[string]bool)
	mixinDeps := make(map[string]map[string]bool)
	loops := make(map[string][]actions.Action)
	locmods := make(map[string]map[string]bool) // Python line 582
	for actname := range summarizedActions {
		GetCallsModsRecFull(mod, summarizedActions, actname, calls, mods, mixinDeps, loops)
		// Python line 586: locmods[actname] = get_loc_mods(mod, actname)
		locmods[actname] = make(map[string]bool)
		for _, s := range GetLocMods(mod, actname) {
			locmods[actname][s] = true
		}
	}

	// Python line 587: compute callouts for all actions
	callouts := make(map[string]Callouts)
	for actname := range newActions.All() {
		GetCallouts(mod, newActions, summarizedActions, actname, callouts)
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
		implMixins = make(map[string][]module.MixinDef)
	}

	// Python lines 590-622: For each non-summarized action, check interference.
	for actname, action := range newActions.All() {
		if summarizedActions[actname] {
			continue
		}
		for _, sub := range action.IterSubactions() {
			ca, ok := sub.(*actions.CallAction)
			if !ok {
				continue
			}
			calledName := CanonAct(ca.CalleeName())

			// Python lines 594-597: Compute pre_refed — symbols used in
			// unsummarized before-mixins of the called action.
			preRefed := make(map[string]bool)
			if mod.Mixins != nil {
				for _, m := range mod.Mixins[calledName] {
					if !m.IsAfter() && !summarizedActions[m.Mixer()] {
						if mixerAct, ok := mod.Actions.Get2(m.Mixer()); ok {
							collectActionSymbolNames(mixerAct, preRefed)
						}
					}
				}
			}

			// Build list of all related actions: callee + mixins + impl_mixins
			allCalls := []string{calledName}
			if modMixins, ok := mod.Mixins[calledName]; ok {
				for _, m := range modMixins {
					allCalls = append(allCalls, m.Mixer())
				}
			}
			for _, m := range implMixins[calledName] {
				allCalls = append(allCalls, m.Mixer())
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
					sortStrings(modNames)
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
						sortStrings(callbackNames)
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
		if modMixins, ok := mod.Mixins[calledName]; ok {
			for _, m := range modMixins {
				allCalls = append(allCalls, m.Mixer())
			}
		}
		for _, m := range implMixins[calledName] {
			allCalls = append(allCalls, m.Mixer())
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
func collectActionSymbolNames(action actions.Action, names map[string]bool) {
	for _, arg := range action.ActionArgs() {
		collectNodeSymNames(arg, names)
	}
}

func collectNodeSymNames(node lg.Expr, names map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Const); ok {
		names[c.Name] = true
	}
	if act, ok := node.(actions.Action); ok {
		collectActionSymbolNames(act, names)
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
func ConeOfInfluenceFilter(mod *module.Module, goals []*ast.LabeledFormula) error {
	isoCfg := mod.Cfg.IsolateCfg
	if !isoCfg.ConeOfInfluence {
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
		case *lg.Const:
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
	for _, lfSlice := range [][]*ast.LabeledFormula{
		goals,
		mod.LabeledAxioms,
		mod.LabeledProps,
		mod.LabeledInits,
		mod.LabeledConjs,
		mod.Definitions,
	} {
		for _, lf := range lfSlice {
			collectSymbols(lf.Formula.(lg.Expr))
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
			if apply, ok := dfn.Formula.(*lg.Apply); ok && len(apply.Terms) > 0 {
				if c, ok := apply.Func.(*lg.Const); ok {
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
		for name := range mod.Sig.Symbols {
			if !allSyms[name] {
				delete(mod.Sig.Symbols, name)
			}
		}
	}

	// Filter sorts: collect all sorts referenced by relevant symbols,
	// then remove unreferenced sorts.
	if isoCfg.FilterSymbols && mod.Sig != nil {
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
	for name, act := range mod.Actions.All() {
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

