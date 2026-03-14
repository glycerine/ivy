package isolate

import (
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
// TODO: Full implementation requires:
//   - used_symbols_ast analysis
//   - Complete call graph with mixins
//   - Export/import analysis
//   - Currently stubbed.
func CheckInterference(mod *module.Module, newActions map[string]actions.Action, summarizedActions map[string]bool) error {
	if !DoCheckInterference {
		return nil
	}

	// TODO: implement full interference checking
	// This requires:
	// 1. Computing calls and mods for all summarized actions
	// 2. For each non-summarized action, checking that calls to summarized
	//    actions don't modify visible symbols
	// 3. Checking that exported summarized actions don't modify visible symbols
	// 4. Checking for interfering callbacks

	return nil
}

// ConeOfInfluenceFilter removes symbols not in the cone of influence
// of the verification goals.
//
// Starting from the symbols used in conjectures (goals), we transitively
// follow through definitions and action modifications to find all
// symbols that can influence the goals. Everything else is removed.
//
// TODO: Full implementation requires:
//   - used_symbols_ast / used_symbols_formula analysis
//   - Definition following
//   - Complete integration with module signature
//   - Currently stubbed.
func ConeOfInfluenceFilter(mod *module.Module, goals []*module.LabeledFormula) error {
	if !ConeOfInfluence {
		return nil
	}

	// TODO: implement cone of influence computation
	// Algorithm:
	// 1. Collect all symbols used in goals
	// 2. Follow through definitions to get transitive dependencies
	// 3. For each action, if it modifies a relevant symbol, add all symbols
	//    it reads to the relevant set
	// 4. Iterate to fixpoint
	// 5. Remove all symbols not in the relevant set

	return nil
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
