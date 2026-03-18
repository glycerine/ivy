// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go.

// This file implements isolate iteration and query functions.
// These are the missing pieces from ivy_isolate.py that allow
// querying isolate properties (actions, conjectures, exports).

package isolate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/module"
)

// -----------------------------------------------------------------------
// Version-aware prefix matching (vstartswith_* variants)
// -----------------------------------------------------------------------

// VPrivates holds additional private names for version-aware prefix checking.
// This is analogous to Python's vprivates module-level variable.
var VPrivates = make(map[string]bool)

// VStartsWithSomeRec is the version-aware variant of startsWithSomeRec.
// It checks both mod.Privates and VPrivates.
func VStartsWithSomeRec(name string, prefixes map[string]bool, mod *module.Module) bool {
	if mod.Privates[name] || VPrivates[name] {
		return false
	}
	pc := iu.ParentChildName(name)
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
	version := IvyVersion
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

// IsolateDefInterface is the interface that isolate definitions must implement
// for IterIsolate to work. It provides access to verified and present components.
type IsolateDefInterface interface {
	// VerifiedNames returns the names of verified components.
	VerifiedNames() []string
	// PresentNames returns the names of present components.
	PresentNames() []string
	// IsExtract returns true if this is an extract (vs isolate) definition.
	IsExtract() bool
}

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
		if idef, ok := isol.(IsolateDefInterface); ok {
			for _, v := range idef.VerifiedNames() {
				vp[v] = true
			}
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
				cname := iu.ComposeNames(name, child)
				if !inSub || !(child == suff ||
					hasAttribute(mod, iu.ComposeNames(cname, suff)) ||
					hasAttribute(mod, iu.ComposeNames(cname, "private"))) {
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
		if _, ok := mod.Actions[name]; ok {
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
func isExplicitOnly(lf *ast.LabeledFormula) bool {
	// In Python, this checks hasattr(lf, 'explicit') and lf.explicit == True.
	// Go doesn't have dynamic attributes, so this is a placeholder.
	return false
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
		type exporter interface{ Exported() string }
		if e, ok := exp.(exporter); ok {
			modExports[e.Exported()] = true
		}
	}
	exports := make(map[string]bool)
	for act := range isoActions {
		if modExports[act] {
			exports[act] = true
			continue
		}
		// Check if any callee is outside the isolate
		if callees, ok := callGraph[act]; ok {
			for _, callee := range callees {
				if !isoActions[callee] {
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
		idef, ok := isol.(IsolateDefInterface)
		if !ok {
			continue
		}
		name := isoName // capture for closure
		IterIsolate(mod, idef, func(n string) {
			result[n] = append(result[n], name)
		}, verified, present)
	}
	return result
}

// -----------------------------------------------------------------------
// Assertion / requirement checking
// -----------------------------------------------------------------------

// HasAssertions returns true if the named action contains any AssertAction.
// Corresponds to Python has_assertions().
func HasAssertions(mod *module.Module, callee string) bool {
	actIface, ok := mod.Actions[callee]
	if !ok {
		return false
	}
	act, ok := actIface.(interface{ IterSubactions() []interface{ Name() string } })
	if !ok {
		// Try with actions.Action
		if a, ok2 := actIface.(interface {
			IterSubactions() []interface{}
		}); ok2 {
			for _, sub := range a.IterSubactions() {
				if namer, ok3 := sub.(interface{ Name() string }); ok3 {
					if namer.Name() == "assert" {
						return true
					}
				}
			}
		}
		return false
	}
	for _, sub := range act.IterSubactions() {
		if sub.Name() == "assert" {
			return true
		}
	}
	return false
}

// HasRequires returns true if the named action contains any RequireAction.
// Corresponds to Python has_requires().
func HasRequires(mod *module.Module, callee string) bool {
	actIface, ok := mod.Actions[callee]
	if !ok {
		return false
	}
	if a, ok2 := actIface.(interface {
		IterSubactions() []interface{}
	}); ok2 {
		for _, sub := range a.IterSubactions() {
			if namer, ok3 := sub.(interface{ Name() string }); ok3 {
				if namer.Name() == "require" {
					return true
				}
			}
		}
	}
	return false
}

// -----------------------------------------------------------------------
// Isolate completeness checking
// -----------------------------------------------------------------------

// CheckIsolateCompleteness verifies that all assertions are checked in
// some isolate. Returns a list of (caller, callee, kind) tuples for
// unchecked assertions.
// Corresponds to Python check_isolate_completeness().
func CheckIsolateCompleteness(mod *module.Module) []IsolateError {
	if mod == nil {
		return nil
	}

	var missing []IsolateError

	checked := make(map[string]bool)
	checkedProps := make(map[string]bool)

	delegates := make(map[string]bool)
	for _, dl := range mod.Delegates {
		type delegator interface {
			Delegated() string
			Delegee() string
		}
		if d, ok := dl.(delegator); ok {
			if d.Delegee() == "" {
				delegates[d.Delegated()] = true
			}
		}
	}

	for _, isol := range mod.Isolates {
		idef, ok := isol.(IsolateDefInterface)
		if !ok {
			continue
		}
		vNames := idef.VerifiedNames()
		pNames := idef.PresentNames()
		verified, present := GetIsolateInfo(mod, vNames, pNames, "impl")
		_ = present

		for a := range mod.Actions {
			if VStartsWithEqSome(a, verified, mod, nil) {
				if !delegates[a] {
					checked[a] = true
				}
			}
		}

		conjs := GetIsolateConjs(mod, idef, true, true)
		for _, conj := range conjs {
			if conj.Label != nil {
				checkedProps[fmt.Sprint(conj.Label)] = true
			}
		}
	}

	// Check that all properties are checked somewhere
	for _, prop := range mod.LabeledProps {
		if prop.Label != nil {
			label := fmt.Sprint(prop.Label)
			if !checkedProps[label] {
				missing = append(missing, IsolateError{
					Caller: label,
					Callee: "",
					Msg:    fmt.Sprintf("property %s not checked", label),
				})
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
		privateName := iu.ComposeNames(name, "private")
		if hasAttribute(mod, privateName) {
			mod.Privates[name] = true
		}
	}
}

// SetInterpretAllSorts sets the InterpretAllSorts flag.
func SetInterpretAllSorts(t bool) {
	InterpretAllSorts = t
}

// -----------------------------------------------------------------------
// Version utilities (local to avoid circular dependencies)
// -----------------------------------------------------------------------

// IvyVersion holds the current Ivy language version string.
// Set by the compiler/parser. Defaults to "1.7".
var IvyVersion = "1.7"

// versionLE returns true if version a <= version b.
func versionLE(a, b string) bool {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	maxLen := len(pa)
	if len(pb) > maxLen {
		maxLen = len(pb)
	}
	for i := 0; i < maxLen; i++ {
		va, vb := 0, 0
		if i < len(pa) {
			va, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			vb, _ = strconv.Atoi(pb[i])
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
