// This file implements isolate iteration and query functions.
// These are the missing pieces from ivy_isolate.py that allow
// querying isolate properties (actions, conjectures, exports).

package goivy

import (
	"fmt"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------
// Version-aware prefix matching (vstartswith_* variants)
// -----------------------------------------------------------------------

// VStartsWithSomeRec is the version-aware variant of startsWithSomeRec.
// It checks both mod.Privates and mod.VPrivates.
func VStartsWithSomeRec(name string, prefixes map[string]bool, mod *Module) bool {
	if mod.Privates[name] || mod.VPrivates[name] {
		return false
	}
	pc := mod.Cfg.IuCfg.ParentChildName(name)
	parent := pc[0]
	if parent == "this" {
		return prefixes["this"]
	}
	return VStartsWithEqSomeRec(parent, prefixes, mod)
}

// VStartsWithEqSomeRec is the version-aware variant of startsWithEqSomeRec.
func VStartsWithEqSomeRec(name string, prefixes map[string]bool, mod *Module) bool {
	if prefixes[name] {
		return true
	}
	return VStartsWithSomeRec(name, prefixes, mod)
}

// VStartsWithEqSome is the version-aware variant of StartsWithEqSome.
// For version <= 1.6, it falls back to StartsWithEqSome.
// For version >= 1.7, it uses VStartsWithEqSomeRec with VPrivates.
func VStartsWithEqSome(name string, prefixes map[string]bool, mod *Module, implMap map[string]string) bool {
	version := mod.Cfg.IsolateCfg.IvyVersion
	if versionLE(version, "1.6") {
		return StartsWithEqSome(name, prefixes, mod, implMap)
	}
	if implMap != nil {
		if mapped, ok := implMap[name]; ok {
			name = mapped
		}
	}
	return VStartsWithEqSomeRec(name, prefixes, mod)
}

// -----------------------------------------------------------------------
// IterIsolate: iterate over isolate components
// -----------------------------------------------------------------------

// IsolateDefIface is an alias for module.IsolateDefInterface, kept for
// convenience within the isolate package.
type IsolateDefIface = IsolateDefInterface

// IterIsolate iterates over all components of an isolate, applying fun
// to each component name. If verified is true, verified components are
// included. If present is true, present components are included.
//
// This handles sub-isolates: a sub-isolate is treated as 'present',
// not 'verified'. A component may be visited more than once if there
// are redundant entries.
//
// Corresponds to Python ivy_isolate.py iter_isolate().
func IterIsolate(mod *Module, iso IsolateDefIface, fun func(string), verified, present bool) {
	suff := "impl"
	if iso.IsExtract() {
		suff = "spec"
	}

	// Record all isolate roots to detect sub-isolates
	vp := make(map[string]bool)
	for _, isol := range mod.Isolates.All() {
		for _, v := range isol.VerifiedNames() {
			vp[v] = true
		}
	}

	// Recursive descent through hierarchy
	var recur func(name string, inSub bool)
	recur = func(name string, inSub bool) {
		if inSub {
			if present {
				fun(name)
			}
		} else {
			if verified {
				fun(name)
			}
		}

		if children, ok := mod.Hierarchy.Get2(name); ok {
			for child := range children.All() {
				cname := mod.Cfg.IuCfg.ComposeNames(name, child)
				if !inSub || !(child == suff ||
					hasAttribute(mod, mod.Cfg.IuCfg.ComposeNames(cname, suff)) ||
					hasAttribute(mod, mod.Cfg.IuCfg.ComposeNames(cname, "private"))) {
					recur(cname, inSub || vp[cname])
				}
			}
		}
	}

	for _, ver := range iso.VerifiedNames() {
		recur(ver, false)
	}

	if present {
		for _, pres := range iso.PresentNames() {
			recur(pres, true)
		}
	}
}

// hasAttribute checks if a name is in the module's attributes.
func hasAttribute(mod *Module, name string) bool {
	if mod.Attributes == nil {
		return false
	}
	_, ok := mod.Attributes.Get2(name)
	return ok
}

// -----------------------------------------------------------------------
// Isolate query functions
// -----------------------------------------------------------------------

// GetIsolateActions returns the set of action names present in an isolate.
// Corresponds to Python get_isolate_actions().
//
// Python returns an insertion-ordered dict (iter_isolate order); so does
// this (an InsMap, not a randomly ordered Go map).
func GetIsolateActions(mod *Module, iso IsolateDefIface) *InsMap[string, bool] {
	return getIsolateActions(mod, iso, true)
}

func getIsolateActions(mod *Module, iso IsolateDefIface, includePresent bool) *InsMap[string, bool] {
	result := NewInsMap[string, bool]()
	IterIsolate(mod, iso, func(name string) {
		if _, ok := mod.Actions.Get2(name); ok {
			if _, seen := result.Get2(name); !seen {
				result.Set(name, true)
			}
		}
	}, true, includePresent)
	return result
}

// GetIsolateLFs returns labeled formulas from the isolate that match
// names in the given lfs slice. Respects verified/present flags.
// Corresponds to Python get_isolate_lfs().
func GetIsolateLFs(mod *Module, iso IsolateDefIface, lfs []*LabeledFormula, verified, present bool) []*LabeledFormula {
	// Build map from label name to labeled formula.
	// Python uses lf.label.rep (the Atom's rep string), so we use the Atom's Rep.
	lfMap := make(map[string]*LabeledFormula)
	for _, lf := range lfs {
		if lf.Label != nil {
			if atom, ok := lf.Label.(*Atom); ok {
				lfMap[atom.Rep] = lf
			}
		}
	}

	// Track explicit present names
	explicit := make(map[string]bool)
	for _, pres := range iso.PresentNames() {
		explicit[pres] = true
	}

	memo := make(map[string]bool)
	var result []*LabeledFormula

	IterIsolate(mod, iso, func(name string) {
		if lf, ok := lfMap[name]; ok {
			if !memo[name] {
				// Include if explicitly present or not marked as explicit-only
				if explicit[name] || !isExplicitOnly(lf) {
					result = append(result, lf)
				}
			}
			memo[name] = true
		}
	}, verified, present)

	return result
}

// isExplicitOnly checks if a labeled formula is marked as explicit-only.
// Corresponds to Python: hasattr(lf, 'explicit') and lf.explicit == True.
func isExplicitOnly(lf *LabeledFormula) bool {
	return lf.Explicit
}

// GetIsolateConjs returns conjectures present in an isolate.
// Corresponds to Python get_isolate_conjs().
func GetIsolateConjs(mod *Module, iso IsolateDefIface, verified, present bool) []*LabeledFormula {
	result := GetIsolateLFs(mod, iso, mod.LabeledConjs, verified, present)
	if present && !verified {
		for _, lf := range GetIsolateLFs(mod, iso, mod.LabeledConjs, true, false) {
			if requiresInvariant(mod, lfLabelName(lf)) {
				result = append(result, lf)
			}
		}
	}
	if len(mod.InvariantProviders) == 0 && present {
		return result
	}
	owners := make(map[string]bool)
	IterIsolate(mod, iso, func(name string) { owners[name] = true }, true, false)
	seen := make(map[string]bool)
	filtered := result[:0]
	for _, lf := range result {
		label := lfLabelName(lf)
		providesHere := owners[mod.InvariantProviders[label].owner]
		if (!verified && providesHere) || (verified && !present && requiresInvariant(mod, label) && !providesHere) {
			continue
		}
		filtered = append(filtered, lf)
		seen[label] = true
	}
	if verified {
		for _, lf := range mod.LabeledConjs {
			label := lfLabelName(lf)
			if !seen[label] && owners[mod.InvariantProviders[label].owner] {
				filtered = append(filtered, lf)
			}
		}
	}
	return filtered
}

// GetIsolatePostConjs returns conjectures that appear before the first
// verified conjecture. Corresponds to Python get_isolate_post_conjs().
func GetIsolatePostConjs(mod *Module, iso IsolateDefIface) []*LabeledFormula {
	verConjs := GetIsolateConjs(mod, iso, true, false)
	verSet := make(map[string]bool)
	for _, lf := range verConjs {
		if atom, ok := lf.Label.(*Atom); ok {
			verSet[atom.Rep] = true
		}
	}
	var postConjs []*LabeledFormula
	for _, ver := range mod.LabeledConjs {
		if atom, ok := ver.Label.(*Atom); ok && verSet[atom.Rep] {
			break
		}
		postConjs = append(postConjs, ver)
	}
	return GetIsolateLFs(mod, iso, postConjs, false, true)
}

// GetIsolateExports returns the set of exported action names for an isolate.
// A present action contributes an invariant obligation if it has side effects
// and is exported or can be called from outside the isolate.
// Corresponds to Python get_isolate_exports().
//
// Python: dict.fromkeys(act for act in actions if ...), in the order of
// get_isolate_actions; so this returns an InsMap in that order.
func GetIsolateExports(mod *Module, callGraph map[string][]string, iso IsolateDefIface) *InsMap[string, bool] {
	isoActions := GetIsolateActions(mod, iso)
	verified := getIsolateActions(mod, iso, false)
	mixees := make(map[string][]string)
	for _, mixins := range mod.Mixins.All() {
		for _, mixin := range mixins {
			mixees[mixin.Mixer()] = append(mixees[mixin.Mixer()], mixin.Mixee())
		}
	}
	for name := range isoActions.Clone().All() {
		for _, mixee := range mixees[name] {
			isoActions.Set(mixee, true)
			if verified.Get(name) {
				verified.Set(mixee, true)
			}
		}
	}
	modExports := make(map[string]bool)
	for _, exp := range mod.Exports {
		modExports[exp.Exported()] = true
	}
	// Resolve writes through destructors against the module being queried.
	actCfg := &ActionsConfig{Context: NewActionContext(mod)}
	exports := NewInsMap[string, bool]()
	for act := range isoActions.All() {
		external := modExports[act]
		for _, caller := range callGraph[act] {
			if !isoActions.Get(caller) {
				external = true
				break
			}
		}
		if external && hasIsolateSideEffect(mod, actCfg, act, isoActions, verified, make(map[string]bool)) {
			exports.Set(act, true)
		}
	}
	return exports
}

func hasIsolateSideEffect(mod *Module, actCfg *ActionsConfig, name string, present, verified *InsMap[string, bool], seen map[string]bool) bool {
	name = CanonAct(name)
	if seen[name] || !present.Get(name) {
		return false
	}
	seen[name] = true
	action := mod.Actions.Get(name)
	if action == nil {
		return false
	}
	for _, sub := range action.IterSubactions() {
		if native, ok := sub.(*LogicNativeAction); ok && native.Impure {
			return true
		}
		for _, sym := range ModifiesSingle(sub, actCfg) {
			if _, inSignature := mod.Sig.Symbols.Get2(sym.Name); inSignature {
				return true
			}
		}
		if verified.Get(name) {
			if _, ranking := sub.(*LogicRanking); ranking || IsAssertLike(sub) {
				return true
			}
		}
		if call, ok := sub.(*LogicCallAction); ok {
			for _, target := range call.ActualReturns {
				for _, sym := range ModifiesSingle(NewHavocAction(target), actCfg) {
					if _, inSignature := mod.Sig.Symbols.Get2(sym.Name); inSignature {
						return true
					}
				}
			}
			if hasIsolateSideEffect(mod, actCfg, call.CalleeName(), present, verified, seen) {
				return true
			}
		}
	}
	for _, mixin := range mod.Mixins.Get(name) {
		if hasIsolateSideEffect(mod, actCfg, mixin.Mixer(), present, verified, seen) {
			return true
		}
	}
	return false
}

// GetIsolateMap creates a map from hierarchy names to the list of
// isolate names in which they are present/verified.
// Corresponds to Python get_isolate_map().
func GetIsolateMap(mod *Module, verified, present bool) map[string][]string {
	result := make(map[string][]string)
	for isoName, isol := range mod.Isolates.All() {
		name := isoName // capture for closure
		IterIsolate(mod, isol, func(n string) {
			result[n] = append(result[n], name)
		}, verified, present)
	}
	return result
}

// -----------------------------------------------------------------------
// Assertion / requirement checking
// -----------------------------------------------------------------------

// HasAssertions returns true if the named action contains any AssertAction
// (including subclasses RequiresAction, EnsuresAction, SubgoalAction).
// Corresponds to Python has_assertions() which uses isinstance(action, ia.AssertAction).
func HasAssertions(mod *Module, callee string) bool {
	act, ok := mod.Actions.Get2(callee)
	if !ok {
		return false
	}
	for _, sub := range act.IterSubactions() {
		if IsAssertLike(sub) {
			return true
		}
	}
	return false
}

// HasRequires returns true if the named action contains any RequiresAction.
// Corresponds to Python has_requires() which uses isinstance(action, ia.RequiresAction).
func HasRequires(mod *Module, callee string) bool {
	act, ok := mod.Actions.Get2(callee)
	if !ok {
		return false
	}
	for _, sub := range act.IterSubactions() {
		if _, ok := sub.(*LogicRequiresAction); ok {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------
// Isolate completeness checking
// -----------------------------------------------------------------------

// CheckIsolateCompleteness verifies that all assertions are checked in
// some isolate. Returns a list of errors for unchecked assertions.
// Corresponds to Python check_isolate_completeness (lines 1804-1916).
func CheckIsolateCompleteness(mod *Module) []IsolateError {
	if mod == nil {
		return nil
	}

	checked := make(map[string]bool)
	checkedProps := make(map[string]bool)
	checkedContext := make(map[string]map[string]bool)  // action -> set of verified actions
	verifiedContext := make(map[string]map[string]bool) // action -> set of verified actions

	delegates := make(map[string]bool)
	delegatedTo := make(map[string]string)
	for _, d := range mod.Delegates {
		if d.Delegee() == "" {
			delegates[d.Delegated()] = true
		} else {
			delegatedTo[d.Delegated()] = d.Delegee()
		}
	}
	_ = delegatedTo

	// Build implementation map from implement mixins. Python fills the
	// global implementation_map here, and startswith_eq_some and
	// vstartswith_eq_some map each name through it, so pass it to them.
	implementationMap := make(map[string]string)
	for _, ms := range mod.Mixins.All() {
		for _, m := range ms {
			if isMixinImplement(m) {
				implementationMap[m.Mixee()] = m.Mixer()
			}
		}
	}

	// Process each isolate
	for _, isol := range mod.Isolates.All() {
		// Python:
		//   verified,present = get_isolate_info(mod,isolate,'impl')
		//   save_privates = mod.privates
		//   set_privates(mod,isolate)
		//   verified_actions = ... vstartswith_eq_some(a,verified,mod)
		//   present_actions = ... startswith_eq_some(a,present,mod)
		//   mod.privates = save_privates
		// set_privates rebinds mod.privates to a new set; goivy's
		// SetPrivatesFull mutates the map, so save a copy.
		verified, present := GetIsolateInfoFull(mod, isol, "impl", nil)
		savePrivates := make(map[string]bool, len(mod.Privates))
		for k, v := range mod.Privates {
			savePrivates[k] = v
		}
		SetPrivatesFull(mod, isol, "")
		verifiedActions := make(map[string]bool)
		presentActions := make(map[string]bool)
		for a := range mod.Actions.All() {
			if VStartsWithEqSome(a, verified, mod, implementationMap) {
				verifiedActions[a] = true
			}
			if StartsWithEqSome(a, present, mod, implementationMap) {
				presentActions[a] = true
			}
		}
		mod.Privates = savePrivates

		// Track checked and context
		for a := range verifiedActions {
			if !delegates[a] {
				checked[a] = true
				if verifiedContext[a] == nil {
					verifiedContext[a] = make(map[string]bool)
				}
				for va := range verifiedActions {
					verifiedContext[a][va] = true
				}
			}
		}
		for a := range presentActions {
			if checkedContext[a] == nil {
				checkedContext[a] = make(map[string]bool)
			}
			for va := range verifiedActions {
				checkedContext[a][va] = true
			}
		}

		// Track proved properties
		proved, _ := GetPropsProvedInIsolate(mod, isol)
		for _, prop := range proved {
			if prop.Label != nil {
				label := lfLabelName(prop)
				if label != "" {
					checkedProps[label] = true
				}
			}
		}
	}

	// Build trusted set from native declarations
	trusted := make(map[string]bool)
	for _, n := range mod.Natives {
		if lf, ok := n.(*LabeledFormula); ok && lf.Label != nil {
			trusted[fmt.Sprint(lf.Label)] = true
		}
	}

	var missing []IsolateError

	// Check all action calls
	for actname, action := range mod.Actions.All() {
		if StartsWithEqSome(actname, trusted, mod, implementationMap) {
			continue
		}
		for _, callee := range action.IterCalls() {
			// Match Python defaultdict(set) auto-vivification
			if checkedContext[callee] == nil {
				checkedContext[callee] = make(map[string]bool)
			}
			// Check assertions
			if !(checked[callee] || !HasAssertions(mod, callee) ||
				(delegates[callee] && checkedContext[callee] != nil && checkedContext[callee][actname])) {
				missing = append(missing, IsolateError{
					Caller: actname,
					Callee: callee,
					Mixer:  callee,
					Mixee:  callee,
					Msg:    "assertion is not checked",
				})
			}

			// Check requires
			if HasRequires(mod, callee) {
				if checkedContext[callee] == nil || !checkedContext[callee][actname] {
					missing = append(missing, IsolateError{
						Caller: actname,
						Callee: callee,
						Mixer:  callee,
						Mixee:  callee,
						Kind:   "require",
						Msg:    "requires assertion is not checked",
					})
				}
			}

			// Check mixin assertions
			// Python: mod.mixins[callee] — auto-vivifies
			if mixins := mixinsAutoVivify(mod, callee); len(mixins) > 0 {
				for _, mixin := range mixins {
					mixed := mixin.Mixer()
					// Match Python defaultdict(set) auto-vivification
					if checkedContext[mixed] == nil {
						checkedContext[mixed] = make(map[string]bool)
					}
					if verifiedContext[mixed] == nil {
						verifiedContext[mixed] = make(map[string]bool)
					}

					// Check requires on mixin
					if HasRequires(mod, mixed) {
						verifier := actname
						if mapped, ok := implementationMap[actname]; ok {
							verifier = mapped
						}
						if checkedContext[mixed] == nil || !checkedContext[mixed][verifier] {
							missing = append(missing, IsolateError{
								Caller: actname,
								Callee: mixed,
								Mixer:  mixed,
								Mixee:  mixin.Mixee(),
								Kind:   "require",
								Msg:    "requires assertion in mixin is not checked",
							})
						}
					}

					if !HasAssertions(mod, mixed) || isMixinImplement(mixin) {
						continue
					}
					if mixin.IsAfter() && StartsWithEqSome(callee, trusted, mod, implementationMap) {
						continue
					}

					// Determine verifier
					verifier := actname
					if mixin.IsAfter() {
						verifier = callee
					}
					if mapped, ok := implementationMap[verifier]; ok {
						verifier = mapped
					}
					if checkedContext[mixed] == nil || !checkedContext[mixed][verifier] {
						if verifiedContext[mixed] == nil || !verifiedContext[mixed][actname] {
							missing = append(missing, IsolateError{
								Caller: actname,
								Callee: mixed,
								Mixer:  mixed,
								Mixee:  mixin.Mixee(),
								Msg:    "assertion in mixin is not checked",
							})
						}
					}
				}
			}
		}
	}

	// Check exports
	for _, e := range mod.Exports {
		if e.Scope() != "" { // skip scoped exports
			continue
		}
		callee := e.Exported()
		if !(checked[callee] || !HasAssertions(mod, callee) || delegates[callee]) {
			missing = append(missing, IsolateError{
				Caller: "external",
				Callee: callee,
				Mixer:  callee,
				Mixee:  callee,
				Msg:    "assertion is not checked when called from the environment",
			})
		}
		// Python: mod.mixins[callee] — auto-vivifies
		if mixins := mixinsAutoVivify(mod, callee); len(mixins) > 0 {
			for _, mixin := range mixins {
				mixed := mixin.Mixer()
				// Match Python defaultdict(set) auto-vivification
				if checkedContext[mixed] == nil {
					checkedContext[mixed] = make(map[string]bool)
				}
				if HasAssertions(mod, mixed) && mixin.IsAfter() {
					if checkedContext[mixed] == nil || !checkedContext[mixed][callee] {
						missing = append(missing, IsolateError{
							Caller: "external",
							Callee: mixed,
							Mixer:  mixed,
							Mixee:  mixin.Mixee(),
							Msg:    "assertion in mixin is not checked when called from environment",
						})
					}
				}
			}
		}
	}

	// Python:
	//   if missing:
	//       for x,y,z in missing:
	//           mixer = y.mixer() if isinstance(y,ivy_ast.MixinDef) else y
	//           mixee = y.mixee() if isinstance(y,ivy_ast.MixinDef) else y
	//           print(iu.IvyError(find_some_assertion(mod,mixer,kind=z),"assertion is not checked"))
	//           if mixee != mixer:
	//               print(iu.IvyError(mod.actions[mixee],"...in action {}".format(mixee)))
	//           if x == "external":
	//               print("error: ...when called from the environment")
	//           else:
	//               print(iu.IvyError(find_some_call(mod,x,mixee),"...when called from {}".format(x)))
	for _, m := range missing {
		fmt.Println(pyErrorString(NewIvyError(actionNode(findSomeAssertionKind(mod, m.Mixer, m.Kind)), "assertion is not checked")))
		if m.Mixee != m.Mixer {
			var mixeeAct Node
			if act, ok := mod.Actions.Get2(m.Mixee); ok {
				mixeeAct = act
			}
			fmt.Println(pyErrorString(NewIvyError(mixeeAct, fmt.Sprintf("...in action %s", m.Mixee))))
		}
		if m.Caller == "external" {
			fmt.Println("error: ...when called from the environment")
		} else {
			fmt.Println(pyErrorString(NewIvyError(actionNode(FindSomeCall(mod, m.Caller, m.Mixee)), fmt.Sprintf("...when called from %s", m.Caller))))
		}
	}

	// Check properties
	done := make(map[string]bool)
	for _, prop := range mod.LabeledProps {
		if prop.Label != nil {
			label := lfLabelName(prop)
			if label != "" && !checkedProps[label] && !done[label] {
				// Python: print(iu.IvyError(prop,"property {} not checked".format(label)))
				fmt.Println(pyErrorString(NewIvyError(prop, fmt.Sprintf("property %s not checked", label))))
				missing = append(missing, IsolateError{
					Caller: label,
					Callee: "",
					Msg:    fmt.Sprintf("property %s not checked", label),
				})
				done[label] = true
			}
		}
	}

	return missing
}

// IsolateError represents an error found during isolate completeness checking.
type IsolateError struct {
	Caller string
	Callee string
	Mixer  string // Python: y.mixer() for a mixin entry, else the callee
	Mixee  string // Python: y.mixee() for a mixin entry, else the callee
	Kind   string // "assert", "require", or ""
	Msg    string
}

func (e IsolateError) Error() string {
	return e.Msg
}

// -----------------------------------------------------------------------
// SetPrivates: set private attributes on isolate components
// -----------------------------------------------------------------------

// SetPrivates sets mod.Privates based on the isolate definition.
// Components with "private" attributes are marked as private.
// Corresponds to Python set_privates().
func SetPrivates(mod *Module, iso IsolateDefIface) {
	if mod.Privates == nil {
		mod.Privates = make(map[string]bool)
	}
	// Walk verified and present components looking for private attributes
	allNames := append(iso.VerifiedNames(), iso.PresentNames()...)
	for _, name := range allNames {
		privateName := mod.Cfg.IuCfg.ComposeNames(name, "private")
		if hasAttribute(mod, privateName) {
			mod.Privates[name] = true
		}
	}
}

// -----------------------------------------------------------------------
// Version utilities (local to avoid circular dependencies)
// -----------------------------------------------------------------------

func trimLeftZerosToInteger(n string) (int, error) {
	if n == "" {
		return 0, fmt.Errorf("empty version part")
	}
	for i, c := range n {
		if c != '0' {
			k, err := strconv.Atoi(n[i:])
			if err != nil {
				return 0, err
			}
			return k, nil
		}
	}
	return 0, nil
}

// versionLE returns true if version a <= version b.
func versionLE(a, b string) bool {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	if len(pa) < 2 {
		panic(fmt.Sprintf("need at least version x.y, not: a='%v'", a))
	}
	if len(pb) < 2 {
		panic(fmt.Sprintf("need at least version x.y, not: b='%v'", b))
	}
	maxLen := len(pa)
	if len(pb) > maxLen {
		maxLen = len(pb)
	}
	for i := 0; i < maxLen; i++ {
		var err error
		va, vb := 0, 0
		if i < len(pa) {
			va, err = trimLeftZerosToInteger(pa[i])
			if err != nil {
				panic(err)
			}
		}
		if i < len(pb) {
			vb, err = trimLeftZerosToInteger(pb[i])
			if err != nil {
				panic(err)
			}
		}
		if va < vb {
			return true
		}
		if va > vb {
			return false
		}
	}
	return true // equal
}

// findSomeAssertionKind is Python's find_some_assertion(mod,actname,kind):
// the first AssertAction (or subclass) in the action, or with kind
// "require" the first RequiresAction.
func findSomeAssertionKind(mod *Module, actname string, kind string) ActionsAction {
	if kind != "require" {
		return FindSomeAssertion(mod, actname)
	}
	act, ok := mod.Actions.Get2(actname)
	if !ok {
		return nil
	}
	for _, sub := range act.IterSubactions() {
		if _, ok := sub.(*LogicRequiresAction); ok {
			return sub
		}
	}
	return nil
}

// actionNode converts a possibly-nil action to a Node without a typed nil.
func actionNode(a ActionsAction) Node {
	if a == nil {
		return nil
	}
	return a
}
