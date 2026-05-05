// phase7.go implements Phase 7 helper functions for isolate,
// ported from Python ivy_isolate.py.
package goivy

import (
	"fmt"
	"strings"
)

// StartsWithSomeRec checks if name starts with any prefix recursively,
// respecting module privates.
// Corresponds to Python's startswith_some_rec (ivy_isolate.py).
func StartsWithSomeRec(name string, prefixes []string, mod *Module) bool {
	prefixSet := make(map[string]bool, len(prefixes))
	for _, p := range prefixes {
		prefixSet[p] = true
	}
	return startsWithSomeRec(name, prefixSet, mod)
}

// StartsWithEqSomeRec checks if name equals or starts with any prefix
// recursively, respecting module privates.
// Corresponds to Python's startswith_eq_some_rec (ivy_isolate.py).
func StartsWithEqSomeRec(name string, prefixes []string, mod *Module) bool {
	prefixSet := make(map[string]bool, len(prefixes))
	for _, p := range prefixes {
		prefixSet[p] = true
	}
	return startsWithEqSomeRec(name, prefixSet, mod)
}

// GetStripParams returns the strip parameters for a given symbol name
// and validates that the arguments match the expected strip binding.
// Corresponds to Python's get_strip_params (ivy_isolate.py lines 227-234).
func GetStripParams(mod *Module, name string, args []Expr, stripMap StripMap, stripBinding map[NodeKey]string) ([]string, error) {
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
		key := Key(ap)
		if existing, ok := stripBinding[key]; ok && existing != sp {
			return nil, fmt.Errorf("cannot strip parameter %v from %s", ap, Presentable(name))
		}
	}
	return stripParams, nil
}

// StripNative strips isolate parameters from a single native declaration.
// Corresponds to Python's strip_native (ivy_isolate.py lines 323-331).
func StripNative(native interface{}, stripMap StripMap, mod *Module) interface{} {
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
			if c, ok := args[i].(*Const); ok {
				sp := StripMapLookup(c.Name, stripMap, mod)
				if len(sp) > 0 {
					newSort := StripSort(c.CSort, len(sp))
					args[i] = NewConst(c.Name, newSort)
				}
			}
		}
	}
	return native
}

// StripNativesSlice strips isolate parameters from a slice of native
// declarations.
// Corresponds to Python's strip_natives (ivy_isolate.py lines 333-336).
func StripNativesSlice(natives []interface{}, stripMap StripMap, mod *Module) []interface{} {
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
func HasSideEffectRec(mod *Module, newActions *InsMap[string, ActionsAction], actname string, memo map[string]bool) bool {
	if memo[actname] {
		return false
	}
	memo[actname] = true

	action, ok := newActions.Get2(actname)
	if !ok {
		return false
	}

	for _, sub := range action.IterSubactions() {
		// Impure native actions have side effects.
		if na, ok := sub.(*NativeAction); ok && na.Impure {
			return true
		}
		// Modifications to signature symbols have side effects.
		for _, sym := range Modifies(sub) {
			if mod.Sig != nil {
				if _, inSig := mod.Sig.Symbols.Get2(sym.Name); inSig {
					return true
				}
			}
		}
		// Assert actions (and subclasses) have side effects.
		// Python: isinstance(sub, ia.AssertAction) — matches all subclasses.
		if IsAssertLike(sub) {
			return true
		}
		// Python line 472-473: Ranking has side effects.
		if _, isRanking := sub.(*Ranking); isRanking {
			return true
		}
		// Follow through calls.
		if ca, ok := sub.(*CallAction); ok {
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
func GetPropsProvedInIsolateOrig(mod *Module, iso interface{}) (proved, notProved []*LabeledFormula) {
	savePrivates := mod.Privates
	mod.Privates = make(map[string]bool)
	SetPrivatesFull(mod, iso, "spec")
	verified, _ := GetIsolateInfoFull(mod, iso, "spec", nil)

	checkPr := func(lf *LabeledFormula) bool {
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
func FollowDefinitionsRec(sym string, defs map[string]Expr, allSyms map[string]bool, memo map[string]bool) {
	allSyms[sym] = true
	if rhs, ok := defs[sym]; ok && !memo[sym] {
		memo[sym] = true
		for _, s := range isolateUsedSymbolNames(rhs) {
			FollowDefinitionsRec(s, defs, allSyms, memo)
		}
	}
}

// CollectRelevantDestructors adds destructor symbols that are relevant to
// the given symbol set. For each symbol whose sort has a range sort,
// it collects the sort's destructors.
// Corresponds to Python's collect_relevant_destructors (ivy_isolate.py lines 870-873).
func CollectRelevantDestructors(mod *Module, syms map[string]bool) map[string]bool {
	result := isolateCopyStringSet(syms)
	memo := make(map[string]bool)
	for sym := range syms {
		// Look up the symbol in the signature to get its sort.
		if mod.Sig == nil {
			continue
		}
		entry, ok := mod.Sig.Symbols.Get2(sym)
		if !ok {
			continue
		}
		// If the sort has a range (i.e., is a FunctionSort), collect
		// destructors for the range sort.
		if fs, ok := entry.Sort.(*FunctionSort); ok {
			rngName := isolateSortToName(fs.Range())
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

// ConjToAssume converts a labeled conjecture into an AssumeAction.
// Corresponds to Python's conj_to_assume (ivy_isolate.py lines 1517-1520).
func ConjToAssume(c *LabeledFormula) ActionsAction {
	fmla, ok := c.Formula.(Expr)
	if !ok {
		return NewSequence()
	}
	res := NewAssumeAction(fmla)
	res.SetLineno(c.GetLineno())
	return res
}

// FindSomeAssertion finds the first AssertAction reachable from the named
// action by iterating sub-actions.
// Corresponds to Python's find_some_assertion (ivy_isolate.py lines 1792-1796).
func FindSomeAssertion(mod *Module, actname string) ActionsAction {
	act, ok := mod.Actions.Get2(actname)
	if !ok {
		return nil
	}
	for _, sub := range act.IterSubactions() {
		// Python uses isinstance(action, kind if kind is not None else ia.AssertAction)
		// which matches AssertAction and all subclasses.
		if IsAssertLike(sub) {
			return sub
		}
	}
	return nil
}

// FindSomeCall finds the first CallAction that calls the given callee
// within the named action.
// Corresponds to Python's find_some_call (ivy_isolate.py lines 1798-1802).
func FindSomeCall(mod *Module, actname string, callee string) ActionsAction {
	act, ok := mod.Actions.Get2(actname)
	if !ok {
		return nil
	}
	for _, sub := range act.IterSubactions() {
		if ca, ok := sub.(*CallAction); ok {
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
