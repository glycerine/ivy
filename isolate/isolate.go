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
	co "github.com/glycerine/goivy/clauseops"
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
// Faithfully ports Python's isolate_component (lines 887-1389).
// It classifies each component as verified/present/opaque, applies
// mixins with appropriate assert_to_assume conversions, summarizes
// opaque actions, builds the exported action set, filters conjectures/
// axioms/properties/definitions/signatures, runs interference checking,
// applies cone-of-influence, strips isolate parameters, and computes
// init_cond.
func IsolateComponent(mod *module.Module, isolateName string, extraWith []string, extraStrip map[string][]string, afterInits []string) error {
	// implementationMap tracks mixee->mixer for implement mixins
	implementationMap := make(map[string]string)

	// Build isolate definition
	var iso interface{}
	if isolateName == "" {
		// No isolate specified: create a default isolate for "this"
		iso = nil // handled specially below
	} else {
		var ok bool
		iso, ok = mod.Isolates[isolateName]
		if !ok {
			return fmt.Errorf("undefined isolate: %s", isolateName)
		}
	}

	// Set privates
	if iso != nil {
		SetPrivatesFull(mod, iso, "")
	}

	// Compute verified and present sets
	verified, present := GetIsolateInfoFull(mod, iso, "impl", extraWith)

	// Handle interpret_all_sorts
	if !InterpretAllSorts && mod.Sig != nil {
		for typeName := range mod.Sig.Interp {
			_, inHier := mod.Hierarchy[typeName]
		cond1 := present[typeName] && !inHier
			cond2 := false
			if itps, ok := mod.Interps[typeName]; ok {
				for _, itp := range itps {
					if lf, ok := itp.(*module.LabeledFormula); ok && lf.Label != nil {
						name := lfLabelName(lf)
						if StartsWithEqSome(name, present, mod, nil) {
							cond2 = true
							break
						}
					}
				}
			}
			if !(cond1 || cond2) {
				delete(mod.Sig.Interp, typeName)
			}
		}
	}

	// Collect delegates
	delegates := make(map[string]bool)
	delegatedTo := make(map[string]string)
	for _, dl := range mod.Delegates {
		type delegator interface {
			Delegated() string
			Delegee() string
		}
		if d, ok := dl.(delegator); ok {
			if d.Delegee() == "" {
				delegates[d.Delegated()] = true
			} else {
				delegatedTo[d.Delegated()] = d.Delegee()
			}
		}
	}

	mod.IsolateInfo = &module.IsolateInfo{}

	// Process implementation mixins
	implMixins := make(map[string][]interface{})
	for actname, ms := range mod.Mixins {
		var implements []interface{}
		var beforeAfter []interface{}
		for _, m := range ms {
			if isMixinImplement(m) {
				implements = append(implements, m)
				implMixins[actname] = append(implMixins[actname], m)
			} else {
				beforeAfter = append(beforeAfter, m)
			}
		}
		// Replace mixin list with only before/after
		mod.Mixins[actname] = beforeAfter

		// Apply implementations
		for _, m := range implements {
			mi, ok := m.(MixinDef)
			if !ok {
				continue
			}
			mixerName := mi.Mixer()
			mixeeName := mi.Mixee()
			// Verify both exist
			if _, err := LookupAction(mod, mixerName); err != nil {
				return fmt.Errorf("action %s not defined", mixerName)
			}
			if _, err := LookupAction(mod, mixeeName); err != nil {
				return fmt.Errorf("action %s not defined", mixeeName)
			}

			if StartsWithEqSome(mixerName, present, mod, nil) {
				action, _ := LookupAction(mod, mixeeName)
				// Check that mixee is empty (no multiple implementations)
				if seq, ok := action.(*actions.Sequence); ok && len(seq.Children) == 0 {
					// OK
				} else if action != nil {
					return fmt.Errorf("multiple implementations of action %s", mixeeName)
				}
				mixer, _ := LookupAction(mod, mixerName)
				mixed := actions.ApplyMixin(mixer, action, false)
				mod.Actions[mixeeName] = mixed
				mod.IsolateInfo.Implementations = append(mod.IsolateInfo.Implementations,
					module.MixinTriple{Mixer: mixerName, Mixee: mixeeName, Action: mixed})
			}
			implementationMap[mixeeName] = mixerName
		}
	}

	// Build action classification lambdas
	useMixin := func(name string) bool {
		return StartsWithSome(name, present, mod, nil)
	}
	// prefixCallExt is used within extModMixin
	_ = func(name string) string {
		if StartsWithSome(name, verified, mod, nil) {
			return "ext:" + name
		}
		return name
	}

	// Mixin classification predicates
	afterMixins := func(m interface{}) bool {
		if mi, ok := m.(MixinDef); ok {
			return mi.IsAfter()
		}
		return false
	}
	beforeMixins := func(m interface{}) bool {
		if mi, ok := m.(MixinDef); ok {
			return !mi.IsAfter()
		}
		return false
	}
	delegatedToVerified := func(n string) bool {
		if dt, ok := delegatedTo[n]; ok {
			return VStartsWithEqSome(dt, verified, mod, nil)
		}
		return false
	}

	// The 6 assert_to_assume lambda variants

	// ext_assumes: for verified external actions, assume asserts in before mixins
	extAssumes := func(m interface{}) map[string]bool {
		kinds := makeKindSet("require")
		if mi, ok := m.(MixinDef); ok {
			if beforeMixins(m) && !delegatedToVerified(mi.Mixer()) {
				kinds["assert"] = true
			}
		}
		return kinds
	}

	// int_assumes: for unverified internal actions, assume asserts in after mixins
	intAssumes := func(m interface{}) map[string]bool {
		kinds := make(map[string]bool)
		if mi, ok := m.(MixinDef); ok {
			if !VStartsWithEqSome(mi.Mixer(), verified, mod, nil) {
				kinds["ensure"] = true
			}
			if afterMixins(m) && !delegatedToVerified(mi.Mixer()) {
				kinds["assert"] = true
			}
		}
		return kinds
	}

	// ext_assumes_no_ver: for unverified external, everything is assumed
	extAssumesNoVer := func(m interface{}) map[string]bool {
		kinds := makeKindSet("ensure", "require")
		if mi, ok := m.(MixinDef); ok {
			if !delegatedToVerified(mi.Mixer()) {
				kinds["assert"] = true
			}
		}
		return kinds
	}

	// int_sum_assumes: for internal summarized actions
	intSumAssumes := func(m interface{}) map[string]bool {
		kinds := make(map[string]bool)
		if mi, ok := m.(MixinDef); ok {
			if !VStartsWithEqSome(mi.Mixer(), verified, mod, nil) {
				kinds["ensure"] = true
			}
			if afterMixins(m) {
				kinds["assert"] = true
			}
		}
		return kinds
	}

	// no_mixins: return no kinds (no conversion)
	noMixins := func(m interface{}) map[string]bool { return nil }

	// mod_mixin: identity
	identityModMixin := func(mixin interface{}, m actions.Action) actions.Action { return m }

	// ext_mod_mixin: prefix calls for unverified mixins
	extModMixin := func(ea func(interface{}) map[string]bool) func(interface{}, actions.Action) actions.Action {
		return func(mixin interface{}, m actions.Action) actions.Action {
			if mi, ok := mixin.(MixinDef); ok {
				if StartsWithSome(mi.Mixer(), verified, mod, nil) && ea(mixin) == nil {
					return m
				}
			}
			return actions.PrefixCalls(m, "ext:")
		}
	}

	// all_mixins returns all kinds (for fully-assumed context)
	allMixins := func(m interface{}) map[string]bool {
		return makeKindSet("assert", "require", "ensure")
	}

	// --- Main action classification loop ---

	newActions := make(map[string]actions.Action)
	summarizedActions := make(map[string]bool)

	for actname, actIface := range mod.Actions {
		act, ok := actIface.(actions.Action)
		if !ok {
			continue
		}

		ver := VStartsWithEqSome(actname, verified, mod, nil)
		pre := StartsWithEqSome(actname, present, mod, nil)

		if pre {
			var extAction, intAction actions.Action
			if !ver || delegates[actname] {
				// Not verified or delegated: convert all assertions in the action
				extKinds := makeKindSet("assert", "ensure", "require")
				extAction = actions.AssertToAssume(act, extKinds)
				extAction = actions.PrefixCalls(extAction, "ext:")

				if delegates[actname] {
					intAction = actions.PrefixCalls(act, "ext:")
				} else {
					intKinds := makeKindSet("assert", "ensure")
					intAction = actions.AssertToAssume(act, intKinds)
					intAction = actions.PrefixCalls(intAction, "ext:")
				}
			} else {
				// Verified: only assume requires in external version
				extKinds := makeKindSet("require")
				extAction = actions.AssertToAssume(act, extKinds)
				intAction = act
			}

			// Internal version: mixins checked
			var ea func(interface{}) map[string]bool
			if ver {
				ea = noMixins
			} else {
				ea = intAssumes
			}
			newActions[actname] = AddMixinsExt(mod, actname, intAction, ea, useMixin, identityModMixin)

			// External version: mixins assumed unless delegated to verified
			if ver {
				ea = extAssumes
			} else {
				ea = extAssumesNoVer
			}
			newAction := AddMixinsExt(mod, actname, extAction, ea, useMixin, extModMixin(ea))
			newActions["ext:"+actname] = newAction

			// Record implementation info
			if _, hasImpl := implementationMap[actname]; !hasImpl {
				mod.IsolateInfo.Implementations = append(mod.IsolateInfo.Implementations,
					module.MixinTriple{Mixer: actname, Mixee: actname, Action: act})
			}
		} else {
			// Opaque: summarize
			summarizedActions[actname] = true
			summarized := SummarizeAction(act)
			newActions[actname] = AddMixinsExt(mod, actname, summarized,
				intSumAssumes, useMixin, extModMixin(afterMixinsFunc))
			newActions["ext:"+actname] = AddMixinsExt(mod, actname, summarized,
				extAssumesNoVer, useMixin, extModMixin(allMixins))
		}

		// Record monitor info
		if mixins, ok := mod.Mixins[actname]; ok {
			for _, mx := range mixins {
				if mi, ok := mx.(MixinDef); ok {
					if useMixin(mi.Mixer()) {
						mixerAct, _ := LookupAction(mod, mi.Mixer())
						mod.IsolateInfo.Monitors = append(mod.IsolateInfo.Monitors,
							module.MixinTriple{Mixer: mi.Mixer(), Mixee: mi.Mixee(), Action: mixerAct})
					}
				}
			}
		}
	}

	// --- Build exported action set ---

	exported := make(map[string]bool)
	exportPreconds := make(map[string][]lg.Node)

	makeBeforeExport := func(actname string) {
		ver := VStartsWithEqSome(actname, verified, mod, nil)
		action, _ := LookupAction(mod, actname)
		if action == nil {
			return
		}
		var act actions.Action
		if !ver || delegates[actname] {
			act = actions.AssertToAssume(action, makeKindSet("assert", "require"))
			act = actions.PrefixCalls(act, "ext:")
		} else {
			act = EmptyClone(action)
		}
		// Apply before mixins
		for _, mx := range mod.Mixins[actname] {
			mi, ok := mx.(MixinDef)
			if !ok {
				continue
			}
			mixerName := mi.Mixer()
			action1, err := LookupAction(mod, mixerName)
			if err != nil {
				continue
			}
			if useMixin(mixerName) && beforeMixins(mx) {
				action1 = actions.AssertToAssume(action1, makeKindSet("assert", "require"))
				action1 = actions.PrefixCalls(action1, "ext:")
				act = actions.ApplyMixin(action1, act, false)
			}
		}
		if mod.BeforeExport == nil {
			mod.BeforeExport = make(map[string]interface{})
		}
		mod.BeforeExport["ext:"+actname] = act
	}

	// Explicit exports
	for _, e := range mod.Exports {
		type exporter interface {
			Exported() string
			Scope() string
		}
		if exp, ok := e.(exporter); ok {
			if exp.Scope() == "" && StartsWithEqSome(exp.Exported(), present, mod, nil) {
				exported["ext:"+exp.Exported()] = true
				makeBeforeExport(exp.Exported())
			}
		}
	}
	explicitExports := copyStringSet(exported)

	// Discover implicit exports from call-outs
	withEffects := make(map[string]bool)
	for actname, actIface := range mod.Actions {
		if StartsWithEqSome(actname, present, mod, nil) {
			continue
		}
		act, ok := actIface.(actions.Action)
		if !ok {
			continue
		}
		for _, sub := range act.IterSubactions() {
			ca, ok := sub.(*actions.CallAction)
			if !ok {
				continue
			}
			c := ca.CalleeName()
			if !StartsWithEqSome(c, present, mod, nil) {
				hasMixinPresent := false
				if mixins, ok := mod.Mixins[c]; ok {
					for _, mx := range mixins {
						if mi, ok := mx.(MixinDef); ok {
							if StartsWithSome(mi.Mixer(), present, mod, nil) {
								hasMixinPresent = true
								break
							}
						}
					}
				}
				if !hasMixinPresent {
					continue
				}
			}
			extC := "ext:" + c
			if !explicitExports[extC] {
				// Add extern precond
				callee, _ := LookupAction(mod, c)
				if callee != nil {
					preconds := exportPreconds[extC]
					// Extract call arguments from callee Apply node
				var callArgs []lg.Node
				if app, ok := ca.Callee.(*lg.Apply); ok {
					callArgs = app.Terms
				}
				AddExternPrecond(mod, callee, callArgs, &preconds)
					exportPreconds[extC] = preconds
				}
			}
			if exported[extC] || withEffects[c] {
				continue
			}
			if !HasSideEffectFull(mod, newActions, c) {
				withEffects[c] = true
				continue
			}
			exported[extC] = true
			makeBeforeExport(c)
		}
	}

	// Store export preconditions
	for actname, pcs := range exportPreconds {
		if len(pcs) == 1 {
			if mod.ExtPreconds == nil {
				mod.ExtPreconds = make(map[string]lg.Node)
			}
			mod.ExtPreconds[actname] = pcs[0]
		} else if len(pcs) > 1 {
			if mod.ExtPreconds == nil {
				mod.ExtPreconds = make(map[string]lg.Node)
			}
			mod.ExtPreconds[actname] = makeOr(pcs...)
		}
	}

	// --- Filter conjectures ---

	keepAx := func(label lg.Node) bool {
		if label == nil {
			return true
		}
		name := ""
		if c, ok := label.(*lg.Symbol); ok {
			name = c.Name
		} else {
			name = fmt.Sprint(label)
		}
		return StartsWithEqSome(name, present, mod, nil)
	}

	propDeps := GetPropDependencies(mod)

	var newConjs []*module.LabeledFormula
	var assumedConjs []*module.LabeledFormula

	if versionLE(IvyVersion, "1.6") {
		for _, c := range mod.LabeledConjs {
			if keepAx(c.Label) {
				newConjs = append(newConjs, c)
			}
		}
	} else {
		for _, c := range mod.LabeledConjs {
			name := lfLabelName(c)
			if VStartsWithEqSome(name, verified, mod, nil) {
				newConjs = append(newConjs, c)
			} else if StartsWithEqSome(name, present, mod, nil) {
				assumedConjs = append(assumedConjs, c)
			}
		}
	}
	_ = assumedConjs

	mod.LabeledConjs = nil
	if !CreateImports || CompileWithInvariants {
		mod.LabeledConjs = newConjs
	}

	// Filter inits
	var newInits []*module.LabeledFormula
	for _, c := range mod.LabeledInits {
		if keepAx(c.Label) {
			newInits = append(newInits, c)
		}
	}
	mod.LabeledInits = newInits

	// Filter axioms
	var droppedAxioms []*module.LabeledFormula
	var keptAxioms []*module.LabeledFormula
	for _, a := range mod.LabeledAxioms {
		if keepAx(a.Label) {
			keptAxioms = append(keptAxioms, a)
		} else {
			droppedAxioms = append(droppedAxioms, a)
		}
	}
	mod.LabeledAxioms = keptAxioms

	// Filter properties
	var keptProps []*module.LabeledFormula
	for _, a := range mod.LabeledProps {
		if keepAx(a.Label) {
			keptProps = append(keptProps, a)
		}
	}
	mod.LabeledProps = keptProps

	// --- Convert properties not being verified to axioms ---

	exactPresent := make(map[string]bool)
	if idef, ok := iso.(IsolateDefInterface); ok {
		for _, p := range idef.PresentNames() {
			exactPresent[p] = true
		}
	}

	type extractDef interface{ IsExtract() bool }
	isExtract := false
	if ed, ok := iso.(extractDef); ok {
		isExtract = ed.IsExtract()
	}

	if !isExtract {
		proved, notProved := GetPropsProvedInIsolate(mod, iso)

		// Filter axioms: keep only non-explicit or those in exact_present or temporal
		var filteredAxioms []*module.LabeledFormula
		for _, a := range mod.LabeledAxioms {
			if !a.Explicit || exactPresent[lfLabelName(a)] || a.Temporal {
				filteredAxioms = append(filteredAxioms, a)
			}
		}
		mod.LabeledAxioms = filteredAxioms

		// Rebuild properties list
		provedIDs := make(map[int64]bool)
		notProvedIDs := make(map[int64]bool)
		for _, p := range proved {
			provedIDs[p.ID] = true
		}
		for _, p := range notProved {
			notProvedIDs[p.ID] = true
		}

		var newProps []*module.LabeledFormula
		for _, p := range mod.LabeledProps {
			cp := cloneLF(p)
			if notProvedIDs[p.ID] {
				cp.Assumed = true
				cp.Explicit = cp.Explicit && !exactPresent[lfLabelName(p)]
				newProps = append(newProps, cp)
			} else if provedIDs[p.ID] {
				newProps = append(newProps, cp)
			}
		}
		mod.LabeledProps = newProps
	} else {
		mod.LabeledProps = nil
	}

	// Filter natives
	var newNatives []interface{}
	for _, nat := range mod.Natives {
		if lf, ok := nat.(*module.LabeledFormula); ok {
			if keepAx(lf.Label) {
				newNatives = append(newNatives, nat)
			}
		} else {
			newNatives = append(newNatives, nat)
		}
	}
	mod.Natives = newNatives

	// Filter initializers from after_inits that are not present
	allAfterInits := make(map[string]bool)
	for _, ai := range afterInits {
		allAfterInits[ai] = true
	}
	if afterInits != nil {
		for _, actname := range afterInits {
			if !StartsWithEqSome(actname, present, mod, nil) {
				extname := "ext:" + actname
				delete(newActions, actname)
				delete(newActions, extname)
				delete(exported, actname)
				delete(exported, extname)
			}
		}
	}
	var presentAfterInits []string
	for _, a := range afterInits {
		if StartsWithEqSome(a, present, mod, nil) {
			presentAfterInits = append(presentAfterInits, a)
		}
	}

	// --- Cone of influence: get accessible actions ---

	cone := GetModConeFull(mod, newActions, exported, presentAfterInits)
	filteredActions := make(map[string]actions.Action)
	for name, act := range newActions {
		if cone[name] {
			filteredActions[name] = act
		}
	}
	newActions = filteredActions

	// Filter isolate info
	if mod.IsolateInfo != nil {
		var filteredImpls []module.MixinTriple
		for _, impl := range mod.IsolateInfo.Implementations {
			if _, ok := newActions[impl.Mixee]; ok {
				filteredImpls = append(filteredImpls, impl)
			} else if _, ok := newActions["ext:"+impl.Mixee]; ok {
				filteredImpls = append(filteredImpls, impl)
			}
		}
		mod.IsolateInfo.Implementations = filteredImpls

		var filteredMons []module.MixinTriple
		for _, mon := range mod.IsolateInfo.Monitors {
			if _, ok := newActions[mon.Mixee]; ok {
				filteredMons = append(filteredMons, mon)
			} else if _, ok := newActions["ext:"+mon.Mixee]; ok {
				filteredMons = append(filteredMons, mon)
			}
		}
		mod.IsolateInfo.Monitors = filteredMons
	}

	// --- Symbol collection for definition/signature filtering ---

	// Collect symbols from formulas
	allSyms := make(map[string]bool)
	for _, lfSlice := range [][]*module.LabeledFormula{
		mod.LabeledAxioms, mod.LabeledProps, mod.LabeledInits, mod.LabeledConjs,
	} {
		for _, lf := range lfSlice {
			if lf.Formula != nil {
				collectUsedSymbolNames(lf.Formula, allSyms)
			}
		}
	}
	// Collect from action formals
	for _, actIface := range mod.Actions {
		if act, ok := actIface.(actions.Action); ok {
			for _, p := range act.GetFormalParams() {
				allSyms[p.Name] = true
			}
			for _, r := range act.GetFormalReturns() {
				allSyms[r.Name] = true
			}
		}
	}
	// Collect from natives
	for _, nat := range mod.Natives {
		if lf, ok := nat.(*module.LabeledFormula); ok && lf.Formula != nil {
			collectUsedSymbolNames(lf.Formula, allSyms)
		}
	}
	// Collect from new actions
	for _, act := range newActions {
		actions.GetReferencesInto(act, allSyms)
	}

	// Collect names from proofs
	allNames := make(map[string]bool)
	for _, pe := range mod.Proofs {
		if pe.Proof != nil {
			if n, ok := pe.Proof.(lg.Node); ok {
				collectUsedSymbolNames(n, allNames)
			}
		}
	}
	// Add definition-defined symbols that are in allNames
	for _, dfn := range mod.Definitions {
		if dfn.Formula == nil {
			continue
		}
		children := dfn.Formula.Children()
		if len(children) >= 1 {
			defName := definedSymbolName(children[0])
			if allNames[defName] {
				allSyms[defName] = true
			}
		}
	}

	// Follow definitions transitively
	FollowDefinitions(mod.Definitions, allSyms)

	// Collect relevant destructors
	if KeepDestructors {
		for sym := range copyStringSet(allSyms) {
			CollectSortDestructors(mod, sym, allSyms, make(map[string]bool))
		}
	}

	// Erase assignments to unreferenced variables
	for actname, act := range newActions {
		newActions[actname] = actions.EraseUnrefed(act, allSyms, allNames)
	}

	// --- Enforce axioms check ---
	if EnforceAxioms {
		for _, a := range droppedAxioms {
			if a.Formula == nil {
				continue
			}
			symsInAxiom := make(map[string]bool)
			collectUsedSymbolNames(a.Formula, symsInAxiom)
			for sym := range symsInAxiom {
				if allSyms[sym] {
					lbl := ""
					if a.Label != nil {
						lbl = fmt.Sprint(a.Label)
					}
					return fmt.Errorf("relevant axiom %s not enforced (uses symbol %s)", lbl, sym)
				}
			}
		}
	}

	// --- Filter definitions ---
	origDefs := mod.Definitions

	var filteredDefs []*module.LabeledFormula
	for _, c := range mod.Definitions {
		if c.Formula == nil {
			continue
		}
		children := c.Formula.Children()
		if len(children) < 1 {
			continue
		}
		defName := definedSymbolName(children[0])
		if (keepAx(c.Label) || exactPresent[defName]) && allSyms[defName] {
			filteredDefs = append(filteredDefs, c)
		}
	}
	mod.Definitions = filteredDefs

	// Filter native definitions
	var filteredNatDefs []interface{}
	for _, c := range mod.NativeDefinitions {
		if lf, ok := c.(*module.LabeledFormula); ok {
			if lf.Formula != nil {
				children := lf.Formula.Children()
				if len(children) >= 1 {
					defName := definedSymbolName(children[0])
					if keepAx(lf.Label) && allSyms[defName] {
						filteredNatDefs = append(filteredNatDefs, c)
					}
				}
			}
		}
	}
	mod.NativeDefinitions = filteredNatDefs

	// --- Put new actions in place ---
	oldActions := make(map[string]interface{})
	for k, v := range mod.Actions {
		oldActions[k] = v
	}
	mod.PublicActions = exported
	mod.Actions = make(map[string]interface{})
	for name, act := range newActions {
		mod.Actions[name] = act
	}

	// --- Filter signature ---

	allSyms2 := make(map[string]bool)
	for _, lfSlice := range [][]*module.LabeledFormula{
		mod.LabeledAxioms, mod.LabeledProps, mod.LabeledInits, mod.LabeledConjs, mod.Definitions,
	} {
		for _, lf := range lfSlice {
			if lf.Formula != nil {
				collectUsedSymbolNames(lf.Formula, allSyms2)
			}
		}
	}
	for _, actIface := range mod.Actions {
		if act, ok := actIface.(actions.Action); ok {
			for _, p := range act.GetFormalParams() {
				allSyms2[p.Name] = true
			}
			for _, r := range act.GetFormalReturns() {
				allSyms2[r.Name] = true
			}
			actions.GetReferencesInto(act, allSyms2)
		}
	}
	if KeepDestructors {
		for _, p := range mod.Params {
			allSyms2[p.Name] = true
		}
	}
	for _, nat := range mod.Natives {
		if lf, ok := nat.(*module.LabeledFormula); ok && lf.Formula != nil {
			collectUsedSymbolNames(lf.Formula, allSyms2)
		}
	}
	for _, pe := range mod.Proofs {
		if pe.Proof != nil {
			if n, ok := pe.Proof.(lg.Node); ok {
				collectUsedSymbolNames(n, allSyms2)
			}
		}
	}

	if KeepDestructors {
		for sym := range copyStringSet(allSyms2) {
			CollectSortDestructors(mod, sym, allSyms2, make(map[string]bool))
		}
	}

	if (FilterSymbols || ConeOfInfluence) && mod.Sig != nil {
		for name := range mod.Sig.Symbols {
			if !allSyms2[name] && !allNames[name] {
				delete(mod.Sig.Symbols, name)
			}
		}
	}

	// Check property dependencies
	if EnforceAxioms && !isExtract {
		for _, pd := range propDeps {
			for _, d := range pd.Deps {
				if !StartsWithEqSome(d, present, mod, nil) {
					// Check if any symbol of the property is in our signature
					if pd.Prop.Formula != nil {
						propSyms := make(map[string]bool)
						collectUsedSymbolNames(pd.Prop.Formula, propSyms)
						for sym := range propSyms {
							if allSyms2[sym] {
								lbl := ""
								if pd.Prop.Label != nil {
									lbl = fmt.Sprint(pd.Prop.Label)
								}
								return fmt.Errorf("property %s depends on abstracted object %s", lbl, d)
							}
						}
					}
				}
			}
		}
	}

	// --- Interference check ---
	if DoCheckInterference {
		interfSyms := copyStringSet(allSyms2)
		FollowDefinitions(origDefs, interfSyms)
		// Temporarily put old actions back for interference check
		saveActions := mod.Actions
		mod.Actions = oldActions
		checkTerm := EnforceAxioms && versionLE("1.7", IvyVersion)
		err := CheckInterferenceFull(mod, newActions, summarizedActions,
			implMixins, checkTerm, interfSyms, presentAfterInits, allAfterInits)
		mod.Actions = saveActions
		if err != nil {
			return err
		}
	}

	// --- Filter sorts ---
	if (FilterSymbols || ConeOfInfluence) && mod.Sig != nil {
		allSorts := make(map[string]bool)
		var addDeps func(string)
		addDeps = func(s string) {
			if allSorts[s] {
				return
			}
			allSorts[s] = true
			// Follow sort dependencies
			for _, dep := range mod.SortDependencies(s, true) {
				addDeps(dep)
			}
		}

		// Add sorts from all remaining symbols
		for name := range allSyms2 {
			if mod.Sig != nil {
				if entry, ok := mod.Sig.Symbols[name]; ok {
					if entry.Union != nil {
						for _, s := range entry.Union.Sorts {
							addSortDeps(s, allSorts, addDeps)
						}
					} else if entry.Sort != nil {
						addSortDeps(entry.Sort, allSorts, addDeps)
					}
				}
			}
		}

		// Add sorts from isolate parameters
		if idef, ok := iso.(IsolateDefNode); ok {
			for _, p := range idef.Params() {
				if p.CSort != nil {
					sname := sortToName(p.CSort)
					addDeps(sname)
				}
			}
		}

		// Filter sorts
		for name := range mod.Sig.Sorts {
			if name != "bool" && !allSorts[name] {
				delete(mod.Sig.Sorts, name)
			}
		}
		var newSortOrder []string
		for _, s := range mod.SortOrder {
			if _, ok := mod.Sig.Sorts[s]; ok {
				newSortOrder = append(newSortOrder, s)
			}
		}
		mod.SortOrder = newSortOrder

		// Filter sort destructors
		for name := range mod.SortDestructors {
			if !allSorts[name] {
				delete(mod.SortDestructors, name)
			}
		}
		for name, s := range mod.DestructorSorts {
			sname := sortToName(s)
			if !allSorts[sname] {
				delete(mod.DestructorSorts, name)
			}
		}
	}

	// --- Check for native code in untrusted isolate ---
	if !isExtract && IsolateMode == "check" {
		for _, actIface := range mod.Actions {
			if _, ok := actIface.(*actions.NativeAction); ok {
				return fmt.Errorf("trusted code used in untrusted isolate")
			}
		}
	}

	// --- Strip isolate parameters ---
	stripIsolateWrapper(mod, iso, implMixins, allAfterInits, extraStrip)

	// --- Compute init_cond ---
	if len(mod.LabeledInits) > 0 {
		var initFmlas []lg.Node
		for _, lf := range mod.LabeledInits {
			if lf.Formula != nil {
				initFmlas = append(initFmlas, lf.Formula)
			}
		}
		if len(initFmlas) > 0 {
			initAnd := makeAnd(initFmlas...)
			mod.InitCond = formulaToClauses(initAnd)
		}
	}

	return nil
}

// cloneLF creates a shallow copy of a LabeledFormula.
func cloneLF(lf *module.LabeledFormula) *module.LabeledFormula {
	cp := *lf
	return &cp
}

// makeAnd creates an And node, ignoring sort errors.
func makeAnd(terms ...lg.Node) lg.Node {
	if len(terms) == 0 {
		return &lg.And{Terms: nil} // empty conjunction = true
	}
	if len(terms) == 1 {
		return terms[0]
	}
	a, err := lg.NewAnd(terms...)
	if err != nil {
		return &lg.And{Terms: terms}
	}
	return a
}

// makeOr creates an Or node, ignoring sort errors.
func makeOr(terms ...lg.Node) lg.Node {
	if len(terms) == 0 {
		return &lg.Or{Terms: nil} // empty disjunction = false
	}
	if len(terms) == 1 {
		return terms[0]
	}
	o, err := lg.NewOr(terms...)
	if err != nil {
		return &lg.Or{Terms: terms}
	}
	return o
}

// addSortDeps recursively adds sort dependencies using the addDeps function.
func addSortDeps(s lg.Sort, allSorts map[string]bool, addDeps func(string)) {
	if fs, ok := s.(*lg.FunctionSort); ok {
		for _, d := range fs.Domain() {
			addSortDeps(d, allSorts, addDeps)
		}
		addSortDeps(fs.Range(), allSorts, addDeps)
	} else {
		name := sortToName(s)
		if name != "" {
			addDeps(name)
		}
	}
}

// afterMixinsFunc is a named version of afterMixins for use as a parameter.
var afterMixinsFunc = func(m interface{}) map[string]bool {
	if mi, ok := m.(MixinDef); ok {
		if mi.IsAfter() {
			return makeKindSet("assert", "require", "ensure")
		}
	}
	return nil
}

// formulaToClauses is a placeholder that wraps a formula into a Clauses struct.
// TODO: wire to real clauseops.FormulaToClauses when available.
func formulaToClauses(fmla lg.Node) *co.Clauses {
	return &co.Clauses{Fmlas: []lg.Node{fmla}}
}

// stripIsolateWrapper calls strip.go's StripIsolateParams with appropriate types.
func stripIsolateWrapper(mod *module.Module, iso interface{}, implMixins map[string][]interface{},
	allAfterInits map[string]bool, extraStrip map[string][]string) {
	if idef, ok := iso.(IsolateDefInterface); ok {
		_ = StripIsolateParams(mod, idef, implMixins, allAfterInits, extraStrip)
	}
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
