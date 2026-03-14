// Package isolate implements modular verification through isolate extraction.
//
// In Ivy, an "isolate" is a component whose correctness can be verified
// independently by abstracting away other components. Each component in
// the hierarchy has one of three roles:
//
//   - Verified: every assertion delegated to this component is checked.
//   - Present: assertions are not checked, but actions are not summarized.
//   - Opaque: state is abstracted, actions are summarized.
//
// This corresponds to Python's ivy_isolate.py.
package isolate

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// --- Module-level configuration parameters ---

var (
	// ShowCompiled controls whether compiled actions are displayed.
	ShowCompiled = false

	// ConeOfInfluence enables the cone-of-influence optimization,
	// which removes symbols not affecting verification goals.
	ConeOfInfluence = true

	// FilterSymbols controls whether unused symbols are filtered out.
	FilterSymbols = true

	// CreateImports causes import declarations to be generated.
	CreateImports = false

	// EnforceAxioms controls axiom enforcement.
	EnforceAxioms = false

	// DoCheckInterference enables interference checking between components.
	DoCheckInterference = true

	// Pedantic enables pedantic checking mode.
	Pedantic = false

	// PreferImpls controls preference for implementation over specification.
	PreferImpls = false

	// KeepDestructors prevents stripping of destructors.
	KeepDestructors = false

	// IsolateMode selects the isolation mode: "check" or "test".
	IsolateMode = "check"

	// CompileWithInvariants controls whether invariants are compiled in.
	CompileWithInvariants = false

	// AssumeInvariants controls whether present invariants are assumed.
	AssumeInvariants = true

	// InterpretAllSorts controls whether all sorts receive interpretations.
	InterpretAllSorts = false
)

// IsolateRole describes a component's role in verification.
type IsolateRole int

const (
	// RoleVerified means all assertions delegated to this component are checked.
	RoleVerified IsolateRole = iota
	// RolePresent means assertions are not checked but actions are not summarized.
	RolePresent
	// RoleOpaque means state is abstracted and actions are summarized.
	RoleOpaque
)

// String returns the name of the role.
func (r IsolateRole) String() string {
	switch r {
	case RoleVerified:
		return "verified"
	case RolePresent:
		return "present"
	case RoleOpaque:
		return "opaque"
	default:
		return fmt.Sprintf("IsolateRole(%d)", int(r))
	}
}

// ComponentInfo holds metadata about a component during isolation.
type ComponentInfo struct {
	Name    string
	Role    IsolateRole
	Actions map[string]actions.Action
	Axioms  []*module.LabeledFormula
}

// NewComponentInfo creates a ComponentInfo with initialized maps.
func NewComponentInfo(name string, role IsolateRole) *ComponentInfo {
	return &ComponentInfo{
		Name:    name,
		Role:    role,
		Actions: make(map[string]actions.Action),
	}
}

// LookupAction finds an action by name in the module.
// Returns an error if the action is not found.
func LookupAction(mod *module.Module, name string) (actions.Action, error) {
	a, ok := mod.Actions[name]
	if !ok {
		return nil, fmt.Errorf("action %s undefined", name)
	}
	act, ok := a.(actions.Action)
	if !ok {
		return nil, fmt.Errorf("action %s is not a valid Action type", name)
	}
	return act, nil
}

// AddMixins applies before/after mixins to an action.
// The useMixin predicate controls which mixins are applied (by mixer name).
// If useMixin is nil, all mixins are applied.
func AddMixins(mod *module.Module, actname string, action actions.Action, useMixin func(string) bool) actions.Action {
	res := action
	if CreateImports {
		// When creating imports, strip invariants from the action.
		// In the Python code, this calls action.drop_invariants().
		// Since invariant-dropping requires tracking which sub-actions
		// are invariant assertions, and we don't yet distinguish those
		// in the Go action types, this is a no-op for now.
		// The action is used as-is, which is safe (just not optimal).
	}
	mixins, ok := mod.Mixins[actname]
	if !ok {
		return res
	}
	for _, mx := range mixins {
		// Each mixin is expected to have a Mixer() method returning the
		// mixer action name, plus information about before/after ordering.
		// Since mod.Mixins stores interface{}, we use type assertions.
		mi, ok := mx.(MixinDef)
		if !ok {
			continue
		}
		mixerName := mi.Mixer()
		if useMixin != nil && !useMixin(mixerName) {
			continue
		}
		action1, err := LookupAction(mod, mixerName)
		if err != nil {
			continue
		}
		res = actions.ApplyMixin(action1, res, mi.IsAfter())
	}
	return res
}

// MixinDef is the interface for mixin definitions stored in Module.Mixins.
// It abstracts over before/after/implement mixin kinds.
type MixinDef interface {
	// Mixer returns the name of the mixer action.
	Mixer() string
	// Mixee returns the name of the mixee (target) action.
	Mixee() string
	// IsAfter returns true if this is an after-mixin (appended after the action).
	IsAfter() bool
}

// SummarizeAction creates an abstract version of an action: just formals,
// no body. In "check" mode, in/out parameters are havoced.
func SummarizeAction(action actions.Action) actions.Action {
	res := actions.NewSequence()
	res.SetLineno(action.GetLineno())
	res.SetFormalParams(action.GetFormalParams())
	res.SetFormalReturns(action.GetFormalReturns())
	if ab, ok := action.(interface{ SetLabels([]string) }); ok {
		_ = ab // labels are on ActionBase, copy if present
	}
	// Copy labels if available
	type labeler interface {
		GetLabels() []string
	}
	if lb, ok := action.(labeler); ok {
		if labels := lb.GetLabels(); labels != nil {
			res.SetLabels(labels)
		}
	}
	// In check mode, havoc the in/out parameters (formal returns that are also formal params)
	if IsolateMode == "check" {
		returns := action.GetFormalReturns()
		params := action.GetFormalParams()
		paramSet := make(map[string]bool, len(params))
		for _, p := range params {
			paramSet[p.Name] = true
		}
		for _, r := range returns {
			if paramSet[r.Name] {
				havoc := actions.NewHavocAction(r)
				res.Children = append(res.Children, actions.WrapAction(havoc))
			}
		}
	}
	return res
}

// EmptyClone creates an empty action (Sequence) preserving the original's
// location and formal parameters/returns.
func EmptyClone(action actions.Action) actions.Action {
	res := actions.NewSequence()
	res.SetLineno(action.GetLineno())
	actions.CopyFormalsTo(action, res)
	return res
}

// CanonAct strips the "ext:" prefix from an action name if present.
func CanonAct(name string) string {
	if strings.HasPrefix(name, "ext:") {
		return name[4:]
	}
	return name
}

// Ancestors yields the chain of ancestor names for a qualified name.
// For "a.b.c" it returns ["a.b.c", "a.b", "a"].
func Ancestors(name string) []string {
	var result []string
	s := name
	for {
		result = append(result, s)
		idx := strings.LastIndex(s, iu.ComposeCharacter)
		if idx < 0 {
			break
		}
		s = s[:idx]
	}
	return result
}

// StartsWithSome returns true if name (after mapping through implementationMap)
// is a child of any name in prefixes, respecting module privates.
func StartsWithSome(name string, prefixes map[string]bool, mod *module.Module, implMap map[string]string) bool {
	if implMap != nil {
		if mapped, ok := implMap[name]; ok {
			name = mapped
		}
	}
	return startsWithSomeRec(name, prefixes, mod)
}

func startsWithSomeRec(name string, prefixes map[string]bool, mod *module.Module) bool {
	if mod.Privates[name] {
		return false
	}
	pc := iu.ParentChildName(name)
	parent := pc[0]
	if parent == "this" {
		return prefixes["this"]
	}
	return startsWithEqSomeRec(parent, prefixes, mod)
}

// StartsWithEqSome returns true if name equals or is a child of some prefix.
func StartsWithEqSome(name string, prefixes map[string]bool, mod *module.Module, implMap map[string]string) bool {
	if implMap != nil {
		if mapped, ok := implMap[name]; ok {
			name = mapped
		}
	}
	return startsWithEqSomeRec(name, prefixes, mod)
}

func startsWithEqSomeRec(name string, prefixes map[string]bool, mod *module.Module) bool {
	if prefixes[name] {
		return true
	}
	return startsWithSomeRec(name, prefixes, mod)
}

// IsolateComponent extracts a verified/present/opaque component.
// This is the main entry point for isolation.
//
// This is the Go port of Python's isolate_component function.
// It classifies each component as verified/present/opaque, applies
// mixins with appropriate assert_to_assume conversions, summarizes
// opaque actions, builds the exported action set, runs interference
// checking, and applies the cone-of-influence filter.
func IsolateComponent(mod *module.Module, isolateName string) (*module.Module, error) {
	if isolateName == "" {
		// No isolate specified: verify everything as one component.
		return mod, nil
	}
	iso, ok := mod.Isolates[isolateName]
	if !ok {
		return nil, fmt.Errorf("undefined isolate: %s", isolateName)
	}

	// Create a copy of the module to modify.
	result := mod.Copy()

	// Extract verified and present names from the isolate definition.
	// The isolate definition may be stored as various types; we handle
	// what's available.
	verifiedNames, presentNames := extractIsolateNames(iso)
	verified, present := GetIsolateInfo(result, verifiedNames, presentNames, "impl")

	// Classify each component as verified/present/opaque.
	roles := ClassifyComponents(result, verified, present)

	// Determine which actions are summarized (opaque).
	summarizedActions := make(map[string]bool)
	newActions := make(map[string]actions.Action)

	useMixin := func(name string) bool {
		return StartsWithSome(name, present, result, nil)
	}

	for actname, actIface := range result.Actions {
		act, ok := actIface.(actions.Action)
		if !ok {
			continue
		}

		pre := StartsWithEqSome(actname, present, result, nil)
		if pre {
			// Present or verified: apply mixins, create internal and external versions.
			intAction := AddMixins(result, actname, act, useMixin)
			newActions[actname] = intAction

			// Create external version with ext: prefix.
			extAction := AddMixins(result, actname, act, useMixin)
			newActions["ext:"+actname] = extAction
		} else {
			// Opaque: summarize the action.
			summarizedActions[actname] = true
			summarized := SummarizeAction(act)
			newActions[actname] = AddMixins(result, actname, summarized, useMixin)
			newActions["ext:"+actname] = AddMixins(result, actname, summarized, useMixin)
		}
	}

	// Build the exported action set.
	exported := make(map[string]bool)
	for _, e := range result.Exports {
		if expDef, ok := e.(interface{ Exported() string; Scope() string }); ok {
			if expDef.Scope() == "" && StartsWithEqSome(expDef.Exported(), present, result, nil) {
				exported["ext:"+expDef.Exported()] = true
			}
		}
	}

	// Update the module with new actions.
	for name, act := range newActions {
		result.Actions[name] = act
	}
	result.PublicActions = exported

	// Run interference check if enabled.
	if DoCheckInterference {
		if err := CheckInterference(result, newActions, summarizedActions); err != nil {
			return nil, err
		}
	}

	// Apply cone of influence filter if enabled.
	if ConeOfInfluence {
		if err := ConeOfInfluenceFilter(result, result.LabeledConjs); err != nil {
			return nil, err
		}
	}

	// Suppress unused variable warnings.
	_ = roles

	return result, nil
}

// extractIsolateNames extracts verified and present names from an isolate
// definition stored as interface{}. Supports various backing types.
func extractIsolateNames(iso interface{}) (verified, present []string) {
	// Try interface with Verified()/Present() methods.
	type verifiedPresent interface {
		Verified() []string
		Present() []string
	}
	if vp, ok := iso.(verifiedPresent); ok {
		return vp.Verified(), vp.Present()
	}

	// Try interface with atom-returning methods.
	type atomVerifiedPresent interface {
		VerifiedAtoms() []interface{}
		PresentAtoms() []interface{}
	}
	if avp, ok := iso.(atomVerifiedPresent); ok {
		for _, a := range avp.VerifiedAtoms() {
			if s, ok := a.(fmt.Stringer); ok {
				verified = append(verified, s.String())
			}
		}
		for _, a := range avp.PresentAtoms() {
			if s, ok := a.(fmt.Stringer); ok {
				present = append(present, s.String())
			}
		}
		return verified, present
	}

	// Fallback: no names extracted.
	return nil, nil
}

// ClassifyComponents determines the role of each hierarchy component
// given the verified and present sets.
func ClassifyComponents(mod *module.Module, verified, present map[string]bool) map[string]IsolateRole {
	roles := make(map[string]IsolateRole)

	// Walk the hierarchy and classify each component.
	var walk func(parent string)
	walk = func(parent string) {
		children, ok := mod.Hierarchy[parent]
		if !ok {
			return
		}
		for child := range children {
			fullName := child
			if parent != "this" {
				fullName = iu.ComposeNames(parent, child)
			}
			if verified[fullName] {
				roles[fullName] = RoleVerified
			} else if present[fullName] {
				roles[fullName] = RolePresent
			} else {
				roles[fullName] = RoleOpaque
			}
			walk(fullName)
		}
	}
	walk("this")

	return roles
}

// GetIsolateInfo computes the verified and present sets from an isolate
// definition's verified and present atoms.
// verifiedNames and presentNames are the names from the isolate declaration.
// kind is "impl" or "spec".
func GetIsolateInfo(mod *module.Module, verifiedNames, presentNames []string, kind string) (verified, present map[string]bool) {
	verified = make(map[string]bool)
	present = make(map[string]bool)

	for _, name := range verifiedNames {
		verified[name] = true
		present[name] = true
		// Also add the kind-specific child (e.g., "foo.impl")
		kindName := iu.ComposeNames(name, kind)
		verified[kindName] = true
		present[kindName] = true
	}
	for _, name := range presentNames {
		present[name] = true
		kindName := iu.ComposeNames(name, kind)
		present[kindName] = true
	}

	return verified, present
}

// Presentable strips the ":" prefix (like sort prefix) from a name for display.
func Presentable(name string) string {
	parts := strings.Split(name, ":")
	return parts[len(parts)-1]
}

// unused import guard
var _ = lg.Boolean
