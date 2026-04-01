// This file implements isolate iteration and query functions.
// These are the missing pieces from ivy_isolate.py that allow
// querying isolate properties (actions, conjectures, exports).

package isolate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/module"
)

// -----------------------------------------------------------------------
// Version-aware prefix matching (vstartswith_* variants)
// -----------------------------------------------------------------------

// VStartsWithSomeRec is the version-aware variant of startsWithSomeRec.
// It checks both mod.Privates and mod.VPrivates.
func VStartsWithSomeRec(name string, prefixes map[string]bool, mod *module.Module) bool {
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
func VStartsWithEqSomeRec(name string, prefixes map[string]bool, mod *module.Module) bool {
	if prefixes[name] {
		return true
	}
	return VStartsWithSomeRec(name, prefixes, mod)
}

// VStartsWithEqSome is the version-aware variant of StartsWithEqSome.
// For version <= 1.6, it falls back to StartsWithEqSome.
// For version >= 1.7, it uses VStartsWithEqSomeRec with VPrivates.
func VStartsWithEqSome(name string, prefixes map[string]bool, mod *module.Module, implMap map[string]string) bool {
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

// IsolateDefInterface is an alias for module.IsolateDefInterface, kept for
// convenience within the isolate package.
type IsolateDefInterface = module.IsolateDefInterface

// IterIsolate iterates over all components of an isolate, applying fun
// to each component name. If verified is true, verified components are
// included. If present is true, present components are included.
//
// This handles sub-isolates: a sub-isolate is treated as 'present',
// not 'verified'. A component may be visited more than once if there
// are redundant entries.
//
// Corresponds to Python ivy_isolate.py iter_isolate().
func IterIsolate(mod *module.Module, iso IsolateDefInterface, fun func(string), verified, present bool) {
	suff := "impl"
	if iso.IsExtract() {
		suff = "spec"
	}

	// Record all isolate roots to detect sub-isolates
	vp := make(map[string]bool)
	for _, isol := range mod.Isolates {
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

		if children, ok := mod.Hierarchy[name]; ok {
			for child := range children {
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
func hasAttribute(mod *module.Module, name string) bool {
	if mod.Attributes == nil {
		return false
	}
	_, ok := mod.Attributes[name]
	return ok
}

// -----------------------------------------------------------------------
// Isolate query functions
// -----------------------------------------------------------------------

// GetIsolateActions returns the set of action names present in an isolate.
// Corresponds to Python get_isolate_actions().
func GetIsolateActions(mod *module.Module, iso IsolateDefInterface) map[string]bool {
	result := make(map[string]bool)
	IterIsolate(mod, iso, func(name string) {
		if _, ok := mod.Actions.Get2(name); ok {
			result[name] = true
		}
	}, true, true)
	return result
}

// GetIsolateLFs returns labeled formulas from the isolate that match
// names in the given lfs slice. Respects verified/present flags.
// Corresponds to Python get_isolate_lfs().
func GetIsolateLFs(mod *module.Module, iso IsolateDefInterface, lfs []*ast.LabeledFormula, verified, present bool) []*ast.LabeledFormula {
	// Build map from label name to labeled formula
	lfMap := make(map[string]*ast.LabeledFormula)
	for _, lf := range lfs {
		if lf.Label != nil {
			key := fmt.Sprint(lf.Label)
			lfMap[key] = lf
		}
	}

	// Track explicit present names
	explicit := make(map[string]bool)
	for _, pres := range iso.PresentNames() {
		explicit[pres] = true
	}

	memo := make(map[string]bool)
	var result []*ast.LabeledFormula

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
func isExplicitOnly(lf *ast.LabeledFormula) bool {
	return lf.Explicit
}

// GetIsolateConjs returns conjectures present in an isolate.
// Corresponds to Python get_isolate_conjs().
func GetIsolateConjs(mod *module.Module, iso IsolateDefInterface, verified, present bool) []*ast.LabeledFormula {
	return GetIsolateLFs(mod, iso, mod.LabeledConjs, verified, present)
}

// GetIsolatePostConjs returns conjectures that appear before the first
// verified conjecture. Corresponds to Python get_isolate_post_conjs().
func GetIsolatePostConjs(mod *module.Module, iso IsolateDefInterface) []*ast.LabeledFormula {
	verConjs := GetIsolateConjs(mod, iso, true, false)
	verSet := make(map[string]bool)
	for _, lf := range verConjs {
		if lf.Label != nil {
			verSet[fmt.Sprint(lf.Label)] = true
		}
	}
	var postConjs []*ast.LabeledFormula
	for _, ver := range mod.LabeledConjs {
		if ver.Label != nil && verSet[fmt.Sprint(ver.Label)] {
			break
		}
		postConjs = append(postConjs, ver)
	}
	return GetIsolateLFs(mod, iso, postConjs, false, true)
}

// GetIsolateExports returns the set of exported action names for an isolate.
// An action is exported if it's in the module's exports and present in the
// isolate, or if it calls an action outside the isolate.
// Corresponds to Python get_isolate_exports().
func GetIsolateExports(mod *module.Module, callGraph map[string][]string, iso IsolateDefInterface) map[string]bool {
	isoActions := GetIsolateActions(mod, iso)
	modExports := make(map[string]bool)
	for _, exp := range mod.Exports {
		modExports[exp.Exported()] = true
	}
	exports := make(map[string]bool)
	for act := range isoActions {
		if modExports[act] {
			exports[act] = true
			continue
		}
		// Check if any caller of this action is outside the isolate
		if callers, ok := callGraph[act]; ok {
			for _, caller := range callers {
				if !isoActions[caller] {
					exports[act] = true
					break
				}
			}
		}
	}
	return exports
}

// GetIsolateMap creates a map from hierarchy names to the list of
// isolate names in which they are present/verified.
// Corresponds to Python get_isolate_map().
func GetIsolateMap(mod *module.Module, verified, present bool) map[string][]string {
	result := make(map[string][]string)
	for isoName, isol := range mod.Isolates {
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
func HasAssertions(mod *module.Module, callee string) bool {
	act, ok := mod.Actions.Get2(callee)
	if !ok {
		return false
	}
	for _, sub := range act.IterSubactions() {
		if actions.IsAssertLike(sub) {
			return true
		}
	}
	return false
}

// HasRequires returns true if the named action contains any RequiresAction.
// Corresponds to Python has_requires() which uses isinstance(action, ia.RequiresAction).
func HasRequires(mod *module.Module, callee string) bool {
	act, ok := mod.Actions.Get2(callee)
	if !ok {
		return false
	}
	for _, sub := range act.IterSubactions() {
		if _, ok := sub.(*actions.RequiresAction); ok {
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
func CheckIsolateCompleteness(mod *module.Module) []IsolateError {
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

	// Build implementation map from implement mixins
	implementationMap := make(map[string]string)
	for _, ms := range mod.Mixins.All() {
		for _, m := range ms {
			if isMixinImplement(m) {
				implementationMap[m.Mixee()] = m.Mixer()
			}
		}
	}

	// Process each isolate
	for _, isol := range mod.Isolates {
		vNames := isol.VerifiedNames()
		pNames := isol.PresentNames()
		verified, _ := GetIsolateInfo(mod, vNames, pNames, "impl")

		// Compute verified and present action sets
		verifiedActions := make(map[string]bool)
		presentActions := make(map[string]bool)
		for a := range mod.Actions.All() {
			if VStartsWithEqSome(a, verified, mod, nil) {
				verifiedActions[a] = true
			}
			if StartsWithEqSome(a, verified, mod, nil) {
				presentActions[a] = true
			}
		}

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
		if lf, ok := n.(*ast.LabeledFormula); ok && lf.Label != nil {
			trusted[fmt.Sprint(lf.Label)] = true
		}
	}

	var missing []IsolateError

	// Check all action calls
	for actname, action := range mod.Actions.All() {
		if StartsWithEqSome(actname, trusted, mod, nil) {
			continue
		}
		for _, callee := range action.IterCalls() {
			// Check assertions
			if !(checked[callee] || !HasAssertions(mod, callee) ||
				(delegates[callee] && checkedContext[callee] != nil && checkedContext[callee][actname])) {
				missing = append(missing, IsolateError{
					Caller: actname,
					Callee: callee,
					Msg:    "assertion is not checked",
				})
			}

			// Check requires
			if HasRequires(mod, callee) {
				if checkedContext[callee] == nil || !checkedContext[callee][actname] {
					missing = append(missing, IsolateError{
						Caller: actname,
						Callee: callee,
						Kind:   "require",
						Msg:    "requires assertion is not checked",
					})
				}
			}

			// Check mixin assertions
			if mixins, ok := mod.Mixins.Get2(callee); ok {
				for _, mixin := range mixins {
					mixed := mixin.Mixer()

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
								Kind:   "require",
								Msg:    "requires assertion in mixin is not checked",
							})
						}
					}

					if !HasAssertions(mod, mixed) || isMixinImplement(mixin) {
						continue
					}
					if mixin.IsAfter() && StartsWithEqSome(callee, trusted, mod, nil) {
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
				Msg:    "assertion is not checked when called from the environment",
			})
		}
		if mixins, ok := mod.Mixins.Get2(callee); ok {
			for _, mixin := range mixins {
				mixed := mixin.Mixer()
				if HasAssertions(mod, mixed) && mixin.IsAfter() {
					if checkedContext[mixed] == nil || !checkedContext[mixed][callee] {
						missing = append(missing, IsolateError{
							Caller: "external",
							Callee: mixed,
							Msg:    "assertion in mixin is not checked when called from environment",
						})
					}
				}
			}
		}
	}

	// Check properties
	done := make(map[string]bool)
	for _, prop := range mod.LabeledProps {
		if prop.Label != nil {
			label := lfLabelName(prop)
			if label != "" && !checkedProps[label] && !done[label] {
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
func SetPrivates(mod *module.Module, iso IsolateDefInterface) {
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
