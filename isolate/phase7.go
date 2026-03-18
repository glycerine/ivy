// phase7.go implements Phase 7 helper functions for isolate,
// ported from Python ivy_isolate.py.
package isolate

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// StartsWithSomeRec checks if name starts with any prefix recursively,
// respecting module privates.
// Corresponds to Python's startswith_some_rec (ivy_isolate.py).
func StartsWithSomeRec(name string, prefixes []string, mod *module.Module) bool {
	prefixSet := make(map[string]bool, len(prefixes))
	for _, p := range prefixes {
		prefixSet[p] = true
	}
	return startsWithSomeRec(name, prefixSet, mod)
}

// StartsWithEqSomeRec checks if name equals or starts with any prefix
// recursively, respecting module privates.
// Corresponds to Python's startswith_eq_some_rec (ivy_isolate.py).
func StartsWithEqSomeRec(name string, prefixes []string, mod *module.Module) bool {
	prefixSet := make(map[string]bool, len(prefixes))
	for _, p := range prefixes {
		prefixSet[p] = true
	}
	return startsWithEqSomeRec(name, prefixSet, mod)
}

// GetStripParams returns the strip parameters for a given symbol name
// and validates that the arguments match the expected strip binding.
// Corresponds to Python's get_strip_params (ivy_isolate.py lines 227-234).
func GetStripParams(mod *module.Module, name string, args []lg.Expr, stripMap StripMap, stripBinding map[lg.NodeKey]string) ([]string, error) {
	name = CanonAct(name)
	stripParams := StripMapLookup(name, stripMap, mod)
	if len(stripParams) == 0 {
		return nil, nil
	}
	if len(args) < len(stripParams) {
		return nil, fmt.Errorf("cannot strip isolate parameters from %s", Presentable(name))
	}
	for i, sp := range stripParams {
		ap := args[i]
		key := lg.Key(ap)
		if existing, ok := stripBinding[key]; ok && existing != sp {
			return nil, fmt.Errorf("cannot strip parameter %v from %s", ap, Presentable(name))
		}
	}
	return stripParams, nil
}

// StripNative strips isolate parameters from a single native declaration.
// Corresponds to Python's strip_native (ivy_isolate.py lines 323-331).
func StripNative(native interface{}, stripMap StripMap, mod *module.Module) interface{} {
	if len(stripMap) == 0 {
		return native
	}
	// Process the native's referenced symbols (args[2:]).
	type argsProvider interface {
		Args() []interface{}
	}
	if ap, ok := native.(argsProvider); ok {
		args := ap.Args()
		for i := 2; i < len(args); i++ {
			if c, ok := args[i].(*lg.Symbol); ok {
				sp := StripMapLookup(c.Name, stripMap, mod)
				if len(sp) > 0 {
					newSort := StripSort(c.CSort, len(sp))
					args[i] = lg.NewSymbol(c.Name, newSort)
				}
			}
		}
	}
	return native
}

// StripNativesSlice strips isolate parameters from a slice of native
// declarations.
// Corresponds to Python's strip_natives (ivy_isolate.py lines 333-336).
func StripNativesSlice(natives []interface{}, stripMap StripMap, mod *module.Module) []interface{} {
	if len(stripMap) == 0 {
		return natives
	}
	result := make([]interface{}, len(natives))
	for i, n := range natives {
		result[i] = StripNative(n, stripMap, mod)
	}
	return result
}

// HasSideEffectRec checks if a named action has side effects on the
// module signature. Follows through calls transitively using a memo set.
// Corresponds to Python's has_side_effect_rec (ivy_isolate.py lines 458-477).
func HasSideEffectRec(mod *module.Module, newActions map[string]actions.Action, actname string, memo map[string]bool) bool {
	if memo[actname] {
		return false
	}
	memo[actname] = true

	action, ok := newActions[actname]
	if !ok {
		return false
	}

	for _, sub := range action.IterSubactions() {
		// Impure native actions have side effects.
		if na, ok := sub.(*actions.NativeAction); ok && na.Impure {
			return true
		}
		// Modifications to signature symbols have side effects.
		for sym := range actions.Modifies(sub) {
			if mod.Sig != nil {
				if _, inSig := mod.Sig.Symbols[sym]; inSig {
					return true
				}
			}
		}
		// Assert actions have side effects.
		if _, ok := sub.(*actions.AssertAction); ok {
			return true
		}
		// SubgoalAction has side effects (similar to ranking).
		if _, ok := sub.(*actions.SubgoalAction); ok {
			return true
		}
		// Follow through calls.
		if ca, ok := sub.(*actions.CallAction); ok {
			calleeName := ca.CalleeName()
			if HasSideEffectRec(mod, newActions, calleeName, memo) {
				return true
			}
		}
	}
	return false
}

// GetPropsProvedInIsolateOrig returns the properties proved and not-proved
// in the given isolate using the original (v1.6) algorithm.
// Corresponds to Python's get_props_proved_in_isolate_orig (ivy_isolate.py lines 741-750).
func GetPropsProvedInIsolateOrig(mod *module.Module, iso interface{}) (proved, notProved []*ast.LabeledFormula) {
	savePrivates := mod.Privates
	mod.Privates = make(map[string]bool)
	SetPrivatesFull(mod, iso, "spec")
	verified, _ := GetIsolateInfoFull(mod, iso, "spec", nil)

	checkPr := func(lf *ast.LabeledFormula) bool {
		if lf.Label == nil {
			return true
		}
		name := lfLabelName(lf)
		return VStartsWithEqSome(name, verified, mod, nil)
	}

	for _, p := range mod.LabeledProps {
		if checkPr(p) {
			proved = append(proved, p)
		} else {
			notProved = append(notProved, p)
		}
	}
	mod.Privates = savePrivates
	return proved, notProved
}

// FollowDefinitionsRec transitively follows definition dependencies for a
// single symbol, adding all referenced symbols to allSyms.
// Corresponds to Python's follow_definitions_rec (ivy_isolate.py lines 840-845).
func FollowDefinitionsRec(sym string, defs map[string]lg.Expr, allSyms map[string]bool, memo map[string]bool) {
	allSyms[sym] = true
	if rhs, ok := defs[sym]; ok && !memo[sym] {
		memo[sym] = true
		for _, s := range usedSymbolNames(rhs) {
			FollowDefinitionsRec(s, defs, allSyms, memo)
		}
	}
}

// CollectRelevantDestructors adds destructor symbols that are relevant to
// the given symbol set. For each symbol whose sort has a range sort,
// it collects the sort's destructors.
// Corresponds to Python's collect_relevant_destructors (ivy_isolate.py lines 870-873).
func CollectRelevantDestructors(mod *module.Module, syms map[string]bool) map[string]bool {
	result := copyStringSet(syms)
	memo := make(map[string]bool)
	for sym := range syms {
		// Look up the symbol in the signature to get its sort.
		if mod.Sig == nil {
			continue
		}
		entry, ok := mod.Sig.Symbols[sym]
		if !ok {
			continue
		}
		// If the sort has a range (i.e., is a FunctionSort), collect
		// destructors for the range sort.
		if fs, ok := entry.Sort.(*lg.FunctionSort); ok {
			rngName := sortToName(fs.Range())
			if rngName != "" {
				CollectSortDestructors(mod, rngName, result, memo)
			}
		}
	}
	return result
}

// SortOrder holds a sort ordering used by the cone-of-influence computation.
// Corresponds to Python's sort_order concept in ivy_isolate.py.
type SortOrder struct {
	Sorts []string
	Less  func(a, b string) bool
}

// GetCone computes the cone of influence: the set of action names reachable
// from the given action name by following calls and native references.
// Corresponds to Python's get_cone (ivy_isolate.py lines 1443-1456).
func GetCone(actionsMap map[string]actions.Action, actionName string, cone map[string]bool) {
	if cone[actionName] {
		return
	}
	cone[actionName] = true
	action, ok := actionsMap[actionName]
	if !ok {
		return
	}
	for _, sub := range action.IterSubactions() {
		if ca, ok := sub.(*actions.CallAction); ok {
			calleeName := ca.CalleeName()
			GetCone(actionsMap, calleeName, cone)
		}
		if na, ok := sub.(*actions.NativeAction); ok {
			// Native actions may reference other actions by name in args[1:]
			for _, arg := range na.Args() {
				if sym, ok := arg.(*lg.Symbol); ok {
					if _, exists := actionsMap[sym.Name]; exists {
						GetCone(actionsMap, sym.Name, cone)
					}
				}
			}
		}
	}
}

// GetModCone returns the cone of action names reachable from the given
// roots (normally the exported actions). An action is accessible if it
// is a root, is referenced from native code, or is called in an initializer.
// Corresponds to Python's get_mod_cone (ivy_isolate.py lines 1463-1475).
func GetModCone(mod *module.Module, actionsMap map[string]actions.Action, roots map[string]bool, afterInits []string) map[string]bool {
	if actionsMap == nil {
		actionsMap = make(map[string]actions.Action)
		for name, a := range mod.Actions {
			if act, ok := a.(actions.Action); ok {
				actionsMap[name] = act
			}
		}
	}
	if roots == nil {
		roots = mod.PublicActions
	}
	cone := make(map[string]bool)
	for a := range roots {
		GetCone(actionsMap, a, cone)
	}
	// Add actions referenced by natives.
	for _, n := range mod.Natives {
		if lf, ok := n.(*ast.LabeledFormula); ok {
			name := lfLabelName(lf)
			if name != "" {
				if _, exists := actionsMap[name]; exists {
					GetCone(actionsMap, name, cone)
				}
			}
		}
	}
	// Add after-init actions.
	for _, ai := range afterInits {
		GetCone(actionsMap, ai, cone)
	}
	return cone
}

// ConjToAssume converts a labeled conjecture into an AssumeAction.
// Corresponds to Python's conj_to_assume (ivy_isolate.py lines 1517-1520).
func ConjToAssume(c *ast.LabeledFormula) actions.Action {
	res := actions.NewAssumeAction(c.Formula)
	res.SetLineno(ast.Location{Line: c.Lineno})
	return res
}

// FindSomeAssertion finds the first AssertAction reachable from the named
// action by iterating sub-actions.
// Corresponds to Python's find_some_assertion (ivy_isolate.py lines 1792-1796).
func FindSomeAssertion(mod *module.Module, actname string) actions.Action {
	actIface, ok := mod.Actions[actname]
	if !ok {
		return nil
	}
	act, ok := actIface.(actions.Action)
	if !ok {
		return nil
	}
	for _, sub := range act.IterSubactions() {
		if _, ok := sub.(*actions.AssertAction); ok {
			return sub
		}
	}
	return nil
}

// FindSomeCall finds the first CallAction that calls the given callee
// within the named action.
// Corresponds to Python's find_some_call (ivy_isolate.py lines 1798-1802).
func FindSomeCall(mod *module.Module, actname string, callee string) actions.Action {
	actIface, ok := mod.Actions[actname]
	if !ok {
		return nil
	}
	act, ok := actIface.(actions.Action)
	if !ok {
		return nil
	}
	for _, sub := range act.IterSubactions() {
		if ca, ok := sub.(*actions.CallAction); ok {
			if ca.CalleeName() == callee {
				return sub
			}
		}
	}
	return nil
}

// ensure imports are used
var (
	_ = strings.HasPrefix
)
