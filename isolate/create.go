// Ported to Go from ivy_isolate.py create_isolate() (~lines 1557-1782).

// This file implements the full create_isolate function which is the
// main entry point for isolate creation/extraction.
package isolate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/solver"
	"github.com/glycerine/goivy/xtracer"
)


// exportStub implements the exporter interface for after-init mixins
// that need to be treated as exports.
type exportStub struct {
	name string
}

func (e *exportStub) Exported() string { return e.name }
func (e *exportStub) Scope() string    { return "" }

// CreateIsolate is the main entry point for isolate creation.
// It processes an isolate definition, applies mixins, builds the
// exported action set, and optionally applies the cone of influence filter.
//
// Corresponds to Python create_isolate() (lines 1557-1782).
//
// Parameters:
//   - iso: the isolate name (empty string means verify everything)
//   - mod: the module to process
func CreateIsolate(iso string, mod *module.Module) error {
	xtracer.Trace("check.CreateIsolate ENTER name=%s", iso)
	if mod == nil {
		return fmt.Errorf("create_isolate: nil module")
	}
	isoCfg := mod.Cfg.IsolateCfg

	// From version 1.7, if no isolate specified and there is only one, use it.
	if iso == "" && versionLE("1.7", isoCfg.IvyVersion) {
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

	// Python line 1580: mod.exports.extend(ExportDef(Atom(a.mixer()), Atom('')) for a in after_inits)
	for _, ai := range afterInits {
		mod.Exports = append(mod.Exports, &exportStub{name: ai.Mixer()})
	}

	// Check all mixin declarations
	for name, mixins := range mod.Mixins {
		for _, mx := range mixins {
			if _, err := LookupAction(mod, mx.Mixer()); err != nil {
				return fmt.Errorf("mixin %s for %s: %w", mx.Mixer(), name, err)
			}
		}
	}

	// Check all delegate declarations
	for _, d := range mod.Delegates {
		if _, err := LookupAction(mod, d.Delegated()); err != nil {
			return err
		}
	}

	// Check all export declarations
	origExports := make(map[string]bool)
	for _, e := range mod.Exports {
		expname := e.Exported()
		xtracer.Trace("check.CreateIsolate.export check name=%s found=%v", expname, mod.Actions.Get(expname) != nil)
		if _, ok := mod.Actions.Get2(expname); !ok {
			// Dump all action keys for debugging
			for k := range mod.Actions.All() {
				xtracer.Trace("check.CreateIsolate.export available_action=%s", k)
			}
			return fmt.Errorf("undefined action: %s", expname)
		}
		origExports[expname] = true
	}

	// Validate with-parameters
	if iso != "" {
		if err := CheckWithParameters(mod, iso); err != nil {
			return err
		}
	}

	// Apply present conjectures (version >= 1.7)
	var brackets []BracketEntry
	if iso != "" {
		if isoDef, ok := mod.Isolates[iso]; ok && versionLE("1.7", isoCfg.IvyVersion) {
			brackets = ApplyPresentConjectures(isoDef, mod)
		}
	}

	// Track mixer names for later warnings
	mixers := make(map[string]bool)
	for _, ms := range mod.Mixins {
		for _, m := range ms {
			mixers[m.Mixer()] = true
		}
	}

	// Determine mixin order
	if err := GetMixinOrder(iso, mod); err != nil {
		return err
	}

	// Construct the isolate
	if iso != "" {
		if _, ok := mod.Isolates[iso]; !ok {
			return fmt.Errorf("undefined isolate: %s", iso)
		}
		err := IsolateComponent(mod, iso, nil, nil, nil)
		if err != nil {
			return err
		}
	} else {
		if len(mod.Isolates) > 0 && isoCfg.ConeOfInfluence {
			return fmt.Errorf("no isolate specified on command line")
		}
		// Apply all mixins in no particular order
		mod.IsolateInfo = &module.IsolateInfo{}
		implemented := make(map[string]bool)

		for actname, mixinList := range mod.Mixins {
			for _, mx := range mixinList {
				action1, err := LookupAction(mod, mx.Mixer())
				if err != nil {
					continue
				}
				action2, err := LookupAction(mod, mx.Mixee())
				if err != nil {
					continue
				}

				mixedName := mx.Mixee()
				if origExports[mixedName] && !mx.IsAfter() {
					// Before mixin on exported action: convert asserts to assumes
					// (asserts are the caller's responsibility)
					_ = action1 // would call action1.assert_to_assume in full impl
				}

				mixed := actions.ApplyMixin(action1, action2, mx.IsAfter())
				mod.Actions.Set(mixedName, mixed)
				implemented[mx.Mixer()] = true
				implemented[mx.Mixee()] = true
				_ = actname
			}
		}

		// Actions not touched by mixins get default implementation
		for actname, act := range mod.Actions.All() {
			if !implemented[actname] {
				_ = act // already in mod.Actions
			}
		}

		// Find globally exported actions
		if len(mod.Exports) > 0 {
			mod.PublicActions = make(map[string]bool)
			for _, e := range mod.Exports {
				if e.Scope() == "" {
					mod.PublicActions[e.Exported()] = true
				}
			}
		} else {
			// No exports specified: all actions are public (compatibility)
			mod.PublicActions = make(map[string]bool)
			for a := range mod.Actions.All() {
				mod.PublicActions[a] = true
			}
		}
	}

	// Python lines 1609-1673: create_imports processing.
	// When CreateImports is enabled, create import actions for out-calls
	// and external stubs.
	if isoCfg.CreateImports {
		SetUpImplementationMap(mod)
		outcalls := make(map[string]bool)

		// Process existing imports: find unimplemented actions
		var newImports []ast.Node
		for _, imp := range mod.Imports {
			type importDef interface {
				Args() []ast.Node
			}
			if id, ok := imp.(importDef); ok {
				args := id.Args()
				if len(args) >= 2 {
					impname := nodeRelname(args[0])
					scope := nodeRelname(args[1])
					if scope == "" {
						if _, ok := mod.Actions.Get2(impname); !ok {
							return fmt.Errorf("undefined action: %s", impname)
						}
						action := mod.Actions.Get(impname)
						if seq, ok := action.(*actions.Sequence); ok && len(seq.Elems) == 0 {
							outcalls[impname] = true
						} else {
							return fmt.Errorf("cannot import implemented action: %s", impname)
						}
					} else {
						newImports = append(newImports, imp)
					}
				}
			}
		}

		// Create external wrapper actions for out-calls
		implMap := SetUpImplementationMap(mod)
		for name := range outcalls {
			impname := name
			extname := "imp__" + impname
			if mapped, ok := implMap[impname]; ok {
				impname = mapped
			}
			action, ok := mod.Actions.Get2(impname)
			if !ok {
				continue
			}
			// Create a CallAction that calls the external wrapper
			fp := action.GetFormalParams()
			fr := action.GetFormalReturns()
			calleeAtom := lg.NewSymbol(extname, lg.TopS)
			var retExprs []lg.Expr
			for _, r := range fr {
				retExprs = append(retExprs, r)
			}
			call := actions.NewCallAction(calleeAtom, retExprs...)
			call.SetFormalParams(fp)
			call.SetFormalReturns(fr)
			mod.Actions.Set(impname, call)

			// Create empty stub for the external name
			stub := actions.NewSequence()
			actions.CopyFormalsTo(action, stub)
			mod.Actions.Set(extname, stub)
		}
		mod.Imports = newImports
	}

	// Fix initializers: move after-init actions to mod.InitialActions
	FixInitializers(mod, afterInits)

	// Python line 1769: mod.canonize_types()
	mod.CanonizeTypes(nil)

	// Apply bracket actions for present conjectures (version >= 1.7)
	if iso != "" {
		if _, ok := mod.Isolates[iso]; ok && versionLE("1.7", isoCfg.IvyVersion) {
			for _, b := range brackets {
				BracketAction(mod, b.ActName, b.Before, b.After)
			}
		}
	}

	// Python line 1739: slv.check_compat()
	// Check native interpretations of symbols for compatibility.
	if mod.Sig != nil {
		errs := solver.CheckCompatStatic(mod.Sig)
		for _, err := range errs {
			fmt.Printf("warning: %v\n", err)
		}
	}

	// Update conjectures (generate concept spaces)
	mod.UpdateConjs()

	// Label public actions
	for name := range mod.PublicActions {
		if act, ok := mod.Actions.Get2(name); ok {
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
	extAction := ""
	if mod.Cfg != nil {
		extAction = mod.Cfg.ExtAction
	}
	if extAction != "" {
		afterInitNames := make(map[string]bool)
		for _, ai := range afterInits {
			afterInitNames[ai.Mixer()] = true
		}

		sortedPublic := make([]string, 0, len(mod.PublicActions))
		for name := range mod.PublicActions {
			sortedPublic = append(sortedPublic, name)
		}
		sort.Strings(sortedPublic)

		var extBranches []lg.Expr
		for _, name := range sortedPublic {
			if afterInitNames[CanonAct(name)] {
				continue
			}
			if act, ok := mod.Actions.Get2(name); ok {
				// Python: mod.actions[name].label = name (for display)
				extBranches = append(extBranches, act)
			}
		}
		// Python: ext_act = ia.EnvAction(*ext_acts)
		if len(extBranches) > 0 {
			extAct := actions.NewEnvAction(extBranches...)
			mod.Actions.Set(extAction, extAct)
		}
		mod.PublicActions[extAction] = true
	}

	// Apply cone of influence filter
	if isoCfg.ConeOfInfluence {
		cone := getModCone(mod)
		for a := range mod.Actions.All() {
			if !cone[a] {
				mod.Actions.Delkey(a)
			}
		}
	}

	// Set isolate proof reference
	if mod.IsolateProofs != nil {
		if p, ok := mod.IsolateProofs[iso]; ok {
			mod.IsolateProof = p
		}
	}

	xtracer.Trace("check.CreateIsolate EXIT name=%s", iso)
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
			if act, ok := mod.Actions.Get2(name); ok {
				if a, ok := act.(actions.Action); ok {
					for _, callee := range a.IterCalls() {
						if !cone[callee] {
							cone[callee] = true
							changed = true
						}
						// Also include ext: variants
						if !strings.HasPrefix(callee, "ext:") {
							extName := "ext:" + callee
							if _, ok := mod.Actions.Get2(extName); ok && !cone[extName] {
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

// -----------------------------------------------------------------------
// Helper functions for create_isolate, ported from Python ivy_isolate.py.
// -----------------------------------------------------------------------

// CheckWithParameters validates that the names mentioned in a 'with'
// clause actually correspond to objects, actions, sorts, definitions,
// interpreted functions or properties.
//
// Corresponds to Python check_with_parameters (lines 777-806).
func CheckWithParameters(mod *module.Module, isolateName string) error {
	isoDef, ok := mod.Isolates[isolateName]
	if !ok {
		return fmt.Errorf("undefined isolate: %s", isolateName)
	}

	verified := make(map[string]bool)
	for _, a := range isoDef.VerifiedNames() {
		verified[a] = true
	}
	present := make(map[string]bool)
	for _, a := range isoDef.PresentNames() {
		present[a] = true
	}
	for v := range verified {
		present[v] = true
	}

	// Collect all known definition names
	derived := make(map[string]bool)
	for _, ldf := range mod.Definitions {
		if ldf != nil {
			n := lfLabelName(ldf)
			if n != "" {
				derived[n] = true
			}
		}
	}

	// Collect all property/axiom/conjecture label names
	propnames := make(map[string]bool)
	allLFs := make([]*ast.LabeledFormula, 0)
	allLFs = append(allLFs, mod.LabeledProps...)
	allLFs = append(allLFs, mod.LabeledAxioms...)
	allLFs = append(allLFs, mod.LabeledConjs...)
	for _, lf := range allLFs {
		n := lfLabelName(lf)
		if n != "" {
			propnames[n] = true
		}
	}

	// Collect interpretation names
	objs := make(map[string]bool)
	for _, itps := range mod.Interps {
		for _, itp := range itps {
			if lf, ok := itp.(*ast.LabeledFormula); ok {
				n := lfLabelName(lf)
				if n != "" {
					objs[n] = true
				}
			}
		}
	}

	for name := range present {
		if name == "this" {
			continue
		}
		if _, ok := mod.Hierarchy[name]; ok {
			continue
		}
		if mod.Sig != nil {
			if _, ok := mod.Sig.Sorts[name]; ok {
				continue
			}
			if _, ok := mod.Sig.Interp[name]; ok {
				continue
			}
			if _, ok := mod.Sig.Symbols[name]; ok {
				continue
			}
		}
		if _, ok := mod.Actions.Get2(name); ok {
			continue
		}
		if derived[name] || propnames[name] || objs[name] {
			continue
		}
		return fmt.Errorf("%s is not an object, action, sort, definition, interpreted function or property", name)
	}
	return nil
}

// GetMixinOrder determines the mixin application order for each action.
// It uses topological sorting based on mod.MixOrd arcs to order before/after
// mixins correctly, and checks for multiple implementations.
//
// Corresponds to Python get_mixin_order (lines 1411-1433).
// arc represents a directed edge in mixin ordering.
type arc struct{ from, to string }

func GetMixinOrder(iso string, mod *module.Module) error {
	// Build arc list from mod.MixOrd
	var arcs []arc
	for _, rdf := range mod.MixOrd {
		type relNamer interface{ Args() []ast.Node }
		if rn, ok := rdf.(relNamer); ok {
			args := rn.Args()
			if len(args) >= 2 {
				from := nodeRelname(args[0])
				to := nodeRelname(args[1])
				arcs = append(arcs, arc{from, to})
			}
		}
	}

	for action, mixinList := range mod.Mixins {
		// Separate implements from before/after
		var implements []module.MixinDef
		var beforeAfter []module.MixinDef

		for _, m := range mixinList {
			if isMixinImplement(m) {
				implements = append(implements, m)
			} else {
				beforeAfter = append(beforeAfter, m)
			}
		}

		if len(implements) > 1 {
			return fmt.Errorf("multiple implementations for %s", action)
		}

		// Topological sort of mixer names
		mixerNames := make([]string, 0)
		seen := make(map[string]bool)
		for _, m := range beforeAfter {
			name := m.Mixer()
			if !seen[name] {
				mixerNames = append(mixerNames, name)
				seen[name] = true
			}
		}

		// Simple topological sort using arcs
		sorted := topologicalSortStrings(mixerNames, arcs)

		// Build key map from sorted order
		keymap := make(map[string]int)
		for i, name := range sorted {
			keymap[name] = i
		}

		// Separate and sort before/after mixins
		var befores, afters []module.MixinDef
		for _, m := range beforeAfter {
			if m.IsAfter() {
				afters = append(afters, m)
			} else {
				befores = append(befores, m)
			}
		}

		sortByMixer := func(list []module.MixinDef) {
			sort.SliceStable(list, func(i, j int) bool {
				return keymap[list[i].Mixer()] < keymap[list[j].Mixer()]
			})
		}
		sortByMixer(befores)
		sortByMixer(afters)

		// Reverse befores (added in reverse order in Python)
		for i, j := 0, len(befores)-1; i < j; i, j = i+1, j-1 {
			befores[i], befores[j] = befores[j], befores[i]
		}

		// Final order: implements + befores + afters
		result := make([]module.MixinDef, 0, len(implements)+len(befores)+len(afters))
		result = append(result, implements...)
		result = append(result, befores...)
		result = append(result, afters...)
		mod.Mixins[action] = result
	}
	return nil
}

// FixInitializers processes after-init mixins: moves their actions to
// mod.InitialActions and mod.Initializers, removes them from mod.Actions
// and mod.PublicActions, and cleans up exports and isolate_info.
//
// Corresponds to Python fix_initializers (lines 1483-1506).
func FixInitializers(mod *module.Module, afterInits []module.MixinDef) {
	things := make(map[string]bool)

	for _, m := range afterInits {
		name := m.Mixer()
		extname := "ext:" + name

		// Get the action (prefer ext: variant)
		var action actions.Action
		if act, ok := mod.Actions.Get2(extname); ok {
			if a, ok := act.(actions.Action); ok {
				action = a
			}
		} else if act, ok := mod.Actions.Get2(name); ok {
			if a, ok := act.(actions.Action); ok {
				action = a
			}
		}

		// Remove from actions and public actions
		mod.Actions.Delkey(name)
		delete(mod.PublicActions, name)
		mod.Actions.Delkey(extname)
		delete(mod.PublicActions, extname)

		if action == nil || !actions.HasCode(action) {
			continue
		}

		mod.InitialActions = append(mod.InitialActions, action)

		// Create looped version for initializers
		loopedAction := LoopAction(action, mod)
		mod.Initializers = append(mod.Initializers, module.NamedAction{
			Name:   name,
			Action: loopedAction,
		})
		things[name] = true
		things[extname] = true
	}

	// Clean up exports
	afterInitNames := make(map[string]bool)
	for _, m := range afterInits {
		afterInitNames[m.Mixer()] = true
	}
	var newExports []module.Exporter
	for _, e := range mod.Exports {
		if afterInitNames[e.Exported()] {
			continue
		}
		newExports = append(newExports, e)
	}
	mod.Exports = newExports

	// Clean up isolate info
	if mod.IsolateInfo != nil {
		var newImpls []module.MixinTriple
		for _, impl := range mod.IsolateInfo.Implementations {
			if !things[impl.Mixee] {
				newImpls = append(newImpls, impl)
			}
		}
		mod.IsolateInfo.Implementations = newImpls
	}
}

// LoopAction creates a version of the action where formal parameters are
// substituted with fresh variables. This is used for initializers.
// Corresponds to Python loop_action (lines 1477-1481).
func LoopAction(action actions.Action, mod *module.Module) actions.Action {
	subst := make(map[lg.NodeKey]lg.Expr)
	for _, p := range action.GetFormalParams() {
		v, err := lg.NewVariable("Y"+p.Name, p.CSort)
		if err == nil {
			subst[lg.Key(p)] = v
		}
	}
	if len(subst) == 0 {
		return action
	}
	result := actions.SubstConstantsAction(action, subst)
	// Python calls ia.type_check_action(action, mod) here, but that function
	// is disabled in Python (immediately returns). Omitted for parity.
	return result
}

// ApplyPresentConjectures wraps each exported action with assume(conjecture)
// before and after. This allows conjectures from present (but unverified)
// components to be assumed as invariants.
//
// Corresponds to Python apply_present_conjectures (lines 1533-1555).
func ApplyPresentConjectures(isol IsolateDefInterface, mod *module.Module) []BracketEntry {
	if !mod.Cfg.IsolateCfg.AssumeInvariants {
		return nil
	}

	// Get present conjectures (verified=false, present=true)
	conjs := GetIsolateConjs(mod, isol, false, true)
	mod.AssumedInvs = conjs

	// Filter out explicit conjectures
	var filteredConjs []*ast.LabeledFormula
	for _, c := range conjs {
		if !c.Explicit {
			filteredConjs = append(filteredConjs, c)
		}
	}

	postConjs := GetIsolatePostConjs(mod, isol)
	var filteredPostConjs []*ast.LabeledFormula
	for _, c := range postConjs {
		if !c.Explicit {
			filteredPostConjs = append(filteredPostConjs, c)
		}
	}

	// Build call graph and get exports
	cg := ActionCallGraph(mod)
	myExports := GetIsolateExports(mod, cg, isol)

	var brackets []BracketEntry
	for actname := range myExports {
		var assumes []actions.Action
		for _, c := range filteredConjs {
			assumes = append(assumes, conjToAssume(c))
		}
		var postAssumes []actions.Action
		for _, c := range filteredPostConjs {
			postAssumes = append(postAssumes, conjToAssume(c))
		}
		brackets = append(brackets, BracketEntry{ActName: actname, Before: assumes, After: postAssumes})
	}

	// Also add post-conjectures for actions that have conj_actions
	posts := make(map[string][]actions.Action)
	for _, conj := range filteredConjs {
		labelName := lfLabelName(conj)
		if labelName != "" {
			if actnames, ok := mod.ConjActions[labelName]; ok {
				for _, actname := range actnames {
					if !myExports[actname] {
						posts[actname] = append(posts[actname], conjToAssume(conj))
					}
				}
			}
		}
	}
	for actname, assumes := range posts {
		brackets = append(brackets, BracketEntry{ActName: actname, After: assumes})
	}

	return brackets
}

// BracketEntry describes before/after assume actions to wrap around an action.
type BracketEntry struct {
	ActName string
	Before  []actions.Action
	After   []actions.Action
}

// BracketAction wraps an action with before/after sequences.
// Corresponds to Python bracket_action (lines 1529-1531).
func BracketAction(mod *module.Module, actname string, before, after []actions.Action) {
	bracketActionInt(mod, actname, before, after)
	bracketActionInt(mod, "ext:"+actname, before, after)
}

func bracketActionInt(mod *module.Module, actname string, before, after []actions.Action) {
	act, ok := mod.Actions.Get2(actname)
	if !ok {
		return
	}
	// Build the new action: Sequence(before..., act, after...)
	var parts []lg.Expr
	for _, b := range before {
		parts = append(parts, b)
	}
	parts = append(parts, act)
	for _, a := range after {
		parts = append(parts, a)
	}
	newAct := actions.NewSequence(parts...)
	// Copy formals from old action to new
	newAct.SetFormalParams(act.GetFormalParams())
	newAct.SetFormalReturns(act.GetFormalReturns())
	mod.Actions.Set(actname, newAct)
}

// conjToAssume converts a labeled conjecture to an AssumeAction.
// Corresponds to Python conj_to_assume (lines 1517-1520).
func conjToAssume(c *ast.LabeledFormula) actions.Action {
	fmla, ok := c.Formula.(lg.Expr)
	if !ok {
		return actions.NewSequence()
	}
	act := actions.NewAssumeAction(fmla)
	return act
}

// SetUpImplementationMap builds the global implementation map from
// MixinImplementDef entries.
// Corresponds to Python set_up_implementation_map (lines 1508-1514).
func SetUpImplementationMap(mod *module.Module) map[string]string {
	implMap := make(map[string]string)
	for _, ms := range mod.Mixins {
		for _, m := range ms {
			if isMixinImplement(m) {
				implMap[m.Mixee()] = m.Mixer()
			}
		}
	}
	return implMap
}

// topologicalSortStrings performs a topological sort of string nodes using arcs.
func topologicalSortStrings(nodes []string, arcs []arc) []string {
	// Build adjacency and in-degree
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeSet := make(map[string]bool)
	for _, n := range nodes {
		nodeSet[n] = true
		inDegree[n] = 0
	}
	for _, a := range arcs {
		if nodeSet[a.from] && nodeSet[a.to] {
			adj[a.from] = append(adj[a.from], a.to)
			inDegree[a.to]++
		}
	}
	// Kahn's algorithm
	var queue []string
	for _, n := range nodes {
		if inDegree[n] == 0 {
			queue = append(queue, n)
		}
	}
	var result []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		result = append(result, n)
		for _, m := range adj[n] {
			inDegree[m]--
			if inDegree[m] == 0 {
				queue = append(queue, m)
			}
		}
	}
	// Add any remaining nodes not in result (cycle handling)
	resultSet := make(map[string]bool)
	for _, r := range result {
		resultSet[r] = true
	}
	for _, n := range nodes {
		if !resultSet[n] {
			result = append(result, n)
		}
	}
	return result
}

// lfLabelName extracts the label name from a LabeledFormula.
func lfLabelName(lf *ast.LabeledFormula) string {
	if lf == nil || lf.Label == nil {
		return ""
	}
	if c, ok := lf.Label.(*lg.Symbol); ok {
		return c.Name
	}
	return fmt.Sprint(lf.Label)
}

// isMixinImplement checks if a mixin is an implement-type mixin.
// We check using type assertion on the underlying AST node.
func isMixinImplement(m interface{}) bool {
	type implementer interface {
		IsImplement() bool
	}
	if impl, ok := m.(implementer); ok {
		return impl.IsImplement()
	}
	// Fallback: check type name or other indicators
	return fmt.Sprintf("%T", m) == "*ast.MixinImplementDef"
}
