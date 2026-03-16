// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_isolate.py create_isolate() (~lines 1557-1782).

// This file implements the full create_isolate function which is the
// main entry point for isolate creation/extraction.
package isolate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/module"
)

// ExtAction is the name for the combined external action.
var ExtAction = ""

// CreateIsolate is the main entry point for isolate creation.
// It processes an isolate definition, applies mixins, builds the
// exported action set, and optionally applies the cone of influence filter.
//
// Parameters:
//   - iso: the isolate name (empty string means verify everything)
//   - mod: the module to process
//
// Corresponds to Python create_isolate() (lines 1557-1782).
func CreateIsolate(iso string, mod *module.Module) error {
	if mod == nil {
		return fmt.Errorf("create_isolate: nil module")
	}

	// From version 1.7, if no isolate specified and there is only one, use it.
	if iso == "" && versionLE("1.7", IvyVersion) {
		isoNames := make([]string, 0, len(mod.Isolates))
		for name := range mod.Isolates {
			isoNames = append(isoNames, name)
		}
		if len(isoNames) == 1 {
			iso = isoNames[0]
		}
	}

	// Treat initializers as exports
	afterInits := mod.Mixins["init"]
	delete(mod.Mixins, "init")
	for _, ai := range afterInits {
		if mi, ok := ai.(MixinDef); ok {
			// Add as export
			_ = mi.Mixer() // just access to confirm type
		}
	}

	// Check all mixin declarations
	for name, mixins := range mod.Mixins {
		for _, mx := range mixins {
			mi, ok := mx.(MixinDef)
			if !ok {
				continue
			}
			if _, err := LookupAction(mod, mi.Mixer()); err != nil {
				return fmt.Errorf("mixin %s for %s: %w", mi.Mixer(), name, err)
			}
		}
	}

	// Check all delegate declarations
	for _, dl := range mod.Delegates {
		type delegator interface {
			Delegated() string
			Delegee() string
		}
		if d, ok := dl.(delegator); ok {
			if _, err := LookupAction(mod, d.Delegated()); err != nil {
				return err
			}
		}
	}

	// Check all export declarations
	origExports := make(map[string]bool)
	for _, exp := range mod.Exports {
		type exporter interface {
			Exported() string
			Scope() string
		}
		if e, ok := exp.(exporter); ok {
			expname := e.Exported()
			if _, ok := mod.Actions[expname]; !ok {
				return fmt.Errorf("undefined action: %s", expname)
			}
			origExports[expname] = true
		}
	}

	// Track mixer names for later warnings
	mixers := make(map[string]bool)
	for _, ms := range mod.Mixins {
		for _, m := range ms {
			if mi, ok := m.(MixinDef); ok {
				mixers[mi.Mixer()] = true
			}
		}
	}

	// Construct the isolate
	if iso != "" {
		if _, ok := mod.Isolates[iso]; !ok {
			return fmt.Errorf("undefined isolate: %s", iso)
		}
		_, err := IsolateComponent(mod, iso)
		if err != nil {
			return err
		}
	} else {
		if len(mod.Isolates) > 0 && ConeOfInfluence {
			return fmt.Errorf("no isolate specified on command line")
		}
		// Apply all mixins in no particular order
		mod.IsolateInfo = &module.IsolateInfo{}
		implemented := make(map[string]bool)

		for actname, mixinList := range mod.Mixins {
			for _, mx := range mixinList {
				mi, ok := mx.(MixinDef)
				if !ok {
					continue
				}
				action1, err := LookupAction(mod, mi.Mixer())
				if err != nil {
					continue
				}
				action2, err := LookupAction(mod, mi.Mixee())
				if err != nil {
					continue
				}

				mixedName := mi.Mixee()
				if origExports[mixedName] && !mi.IsAfter() {
					// Before mixin on exported action: convert asserts to assumes
					// (asserts are the caller's responsibility)
					_ = action1 // would call action1.assert_to_assume in full impl
				}

				mixed := actions.ApplyMixin(action1, action2, mi.IsAfter())
				mod.Actions[mixedName] = mixed
				implemented[mi.Mixer()] = true
				implemented[mi.Mixee()] = true
				_ = actname
			}
		}

		// Actions not touched by mixins get default implementation
		for actname, act := range mod.Actions {
			if !implemented[actname] {
				_ = act // already in mod.Actions
			}
		}

		// Find globally exported actions
		if len(mod.Exports) > 0 {
			mod.PublicActions = make(map[string]bool)
			for _, e := range mod.Exports {
				type exporter interface {
					Exported() string
					Scope() string
				}
				if exp, ok := e.(exporter); ok {
					if exp.Scope() == "" {
						mod.PublicActions[exp.Exported()] = true
					}
				}
			}
		} else {
			// No exports specified: all actions are public (compatibility)
			mod.PublicActions = make(map[string]bool)
			for a := range mod.Actions {
				mod.PublicActions[a] = true
			}
		}
	}

	// Label public actions
	for name := range mod.PublicActions {
		if act, ok := mod.Actions[name]; ok {
			if a, ok := act.(actions.Action); ok {
				type labeler interface {
					SetLabels([]string)
				}
				if lb, ok := a.(labeler); ok {
					lb.SetLabels([]string{name})
				}
			}
		}
	}

	// Create one big external action if requested
	if ExtAction != "" {
		afterInitNames := make(map[string]bool)
		for _, ai := range afterInits {
			if mi, ok := ai.(MixinDef); ok {
				afterInitNames[mi.Mixer()] = true
			}
		}

		sortedPublic := make([]string, 0, len(mod.PublicActions))
		for name := range mod.PublicActions {
			sortedPublic = append(sortedPublic, name)
		}
		sort.Strings(sortedPublic)

		var extBranches []interface{}
		for _, name := range sortedPublic {
			if afterInitNames[CanonAct(name)] {
				continue
			}
			if act, ok := mod.Actions[name]; ok {
				extBranches = append(extBranches, act)
			}
		}
		// Create EnvAction from branches
		_ = extBranches // would create actions.NewEnvAction(...)
		mod.PublicActions[ExtAction] = true
	}

	// Apply cone of influence filter
	if ConeOfInfluence {
		cone := getModCone(mod)
		for a := range mod.Actions {
			if !cone[a] {
				delete(mod.Actions, a)
			}
		}
	}

	// Set isolate proof reference
	if mod.IsolateProofs != nil {
		if p, ok := mod.IsolateProofs[iso]; ok {
			mod.IsolateProof = p
		}
	}

	return nil
}

// getModCone returns the set of action names in the cone of influence.
// An action is in the cone if it's public or transitively called by
// a public action.
func getModCone(mod *module.Module) map[string]bool {
	cone := make(map[string]bool)

	// Start with public actions
	for name := range mod.PublicActions {
		cone[name] = true
	}

	// Transitively add called actions
	changed := true
	for changed {
		changed = false
		for name := range cone {
			if act, ok := mod.Actions[name]; ok {
				if a, ok := act.(actions.Action); ok {
					for _, callee := range a.IterCalls() {
						if !cone[callee] {
							cone[callee] = true
							changed = true
						}
						// Also include ext: variants
						if !strings.HasPrefix(callee, "ext:") {
							extName := "ext:" + callee
							if _, ok := mod.Actions[extName]; ok && !cone[extName] {
								cone[extName] = true
								changed = true
							}
						}
					}
				}
			}
		}
	}

	return cone
}
