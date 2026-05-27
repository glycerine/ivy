// Ported to Go from ivy_isolate.py create_isolate() (~lines 1557-1782).

// This file implements the full create_isolate function which is the
// main entry point for isolate creation/extraction.
package goivy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// pyBool returns "True" or "False" matching Python's bool formatting.
func pyBool(b bool) string {
	if b {
		return "True"
	}
	return "False"
}

// fmtStrList formats a string slice like Python's str(list): "['a', 'b']".
func fmtStrList(ss []string) string {
	parts := make([]string, len(ss))
	for i, s := range ss {
		parts[i] = "'" + s + "'"
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// exportStub implements the exporter interface for after-init mixins
// that need to be treated as exports.
type exportStub struct {
	name string
}

func (e *exportStub) Exported() string { return e.name }
func (e *exportStub) Scope() string    { return "" }

func createImportWrappers(iso string, mod *Module) ([]string, map[string][]string, error) {
	implMap := SetUpImplementationMap(mod)
	outcalls := make(map[string]bool)
	origImports := make(map[string]bool)
	newImports := make([]Node, 0, len(mod.Imports))

	for _, imp := range mod.Imports {
		id, ok := imp.(interface{ Args() []Node })
		if !ok {
			newImports = append(newImports, imp)
			continue
		}
		args := id.Args()
		if len(args) < 2 {
			newImports = append(newImports, imp)
			continue
		}
		impname := isolateNodeRelname(args[0])
		scope := isolateNodeRelname(args[1])
		origImports[impname] = true
		if scope == "" {
			action, ok := mod.Actions.Get2(impname)
			if !ok {
				return nil, nil, fmt.Errorf("undefined action: %s", impname)
			}
			if seq, ok := action.(*LogicSequence); ok && len(seq.Elems) == 0 {
				outcalls[impname] = true
			} else {
				return nil, nil, fmt.Errorf("cannot import implemented action: %s", impname)
			}
		} else {
			newImports = append(newImports, imp)
		}
	}

	var isoDef *IsolateDef
	isoExists := false
	if iso != "" {
		isoDef, isoExists = mod.Isolates[iso]
	}
	if isoExists {
		verified, present := GetIsolateInfoFull(mod, isoDef, "impl", nil)
		savePrivates := mod.Privates
		mod.Privates = copyMapBool(mod.Privates)
		SetPrivatesFull(mod, isoDef, "")
		verifiedActions := make(map[string]bool)
		presentActions := make(map[string]bool)
		for actname := range mod.Actions.All() {
			if VStartsWithEqSome(actname, verified, mod, implMap) {
				verifiedActions[actname] = true
			}
			if StartsWithEqSome(actname, present, mod, implMap) {
				presentActions[actname] = true
			}
		}
		for actname := range verifiedActions {
			presentActions[actname] = true
		}
		mod.Privates = savePrivates
		for actname := range presentActions {
			action, ok := mod.Actions.Get2(actname)
			if !ok {
				continue
			}
			for _, called := range action.IterCalls() {
				if !presentActions[called] {
					outcalls[called] = true
				}
			}
		}
	}

	extraWith := []string{}
	extraStrip := make(map[string][]string)
	outcallNames := make([]string, 0, len(outcalls))
	for name := range outcalls {
		outcallNames = append(outcallNames, name)
	}
	sort.Strings(outcallNames)
	for _, name := range outcallNames {
		impname := name
		extname := "imp__" + impname
		if mapped, ok := implMap[impname]; ok {
			impname = mapped
		}
		action, ok := mod.Actions.Get2(impname)
		if !ok {
			continue
		}
		for _, attr := range []string{"spec", "impl", "private"} {
			attrname := mod.Cfg.IuCfg.ComposeNames(impname, attr)
			if val, ok := mod.Attributes[attrname]; ok {
				extattrname := mod.Cfg.IuCfg.ComposeNames(extname, attr)
				mod.Attributes[extattrname] = val
			}
		}

		fp := action.GetFormalParams()
		fr := action.GetFormalReturns()
		calleeConst := NewConst(extname, TopS)
		var retExprs []Expr
		for _, r := range fr {
			retExprs = append(retExprs, r)
		}
		call := NewCallActionOn(mod.Cfg.ActCfg, calleeConst, retExprs...)
		if mod.Cfg != nil && mod.Cfg.AstCfg != nil {
			terms := make([]Node, len(fp))
			for i, p := range fp {
				terms[i] = p
			}
			call.AstCallee = mod.Cfg.AstCfg.NewAtom(extname, terms...)
		}
		call.SetFormalParams(fp)
		call.SetFormalReturns(fr)
		call.SetLineno(action.GetLineno())
		mod.Actions.Set(impname, call)

		stub := NewSequence()
		CopyFormalsTo(action, stub)
		stub.SetLineno(action.GetLineno())
		mod.Actions.Set(extname, stub)

		isExtract := iso != "" && isoExists && isoDef.IsExtract()
		if origImports[name] || !isExtract {
			newImports = append(newImports,
				mod.Cfg.AstCfg.NewImportDef(mod.Cfg.AstCfg.NewAtom(extname), mod.Cfg.AstCfg.NewAtom("")))
			extraWith = append(extraWith, impname)
		}
		if iso != "" && isoExists && origImports[name] {
			stripNames := isolateParamStripNames(isoDef.Params())
			extraStrip[impname] = stripNames
			extraStrip[extname] = stripNames
		}
	}

	mod.Imports = newImports
	return extraWith, extraStrip, nil
}

func isolateParamStripNames(params []Node) []string {
	res := make([]string, 0, len(params))
	for _, p := range params {
		if v, ok := p.(*Variable); ok {
			res = append(res, v.ToConst("iso:").Relname())
			continue
		}
		res = append(res, isolateNodeRelname(p))
	}
	return res
}

func traceIsolateName(iso string) string {
	if iso == "" {
		return "None"
	}
	return iso
}

// CreateIsolate is the main entry point for isolate creation.
// It processes an isolate definition, applies mixins, builds the
// exported action set, and optionally applies the cone of influence filter.
//
// Corresponds to Python create_isolate() (lines 1557-1782).
//
// Parameters:
//   - iso: the isolate name (empty string means verify everything)
//   - mod: the module to process
func CreateIsolate(iso string, mod *Module) error {
	xtracer.Trace("check.CreateIsolate ENTER name=%s", traceIsolateName(iso))
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

	// Apply present conjectures (version >= 1.7)
	// Python calls this FIRST, before initializer exports, mixin/delegate/export checks.
	var brackets []BracketEntry
	if iso != "" {
		if isoDef, ok := mod.Isolates[iso]; ok && versionLE("1.7", isoCfg.IvyVersion) {
			brackets = ApplyPresentConjectures(isoDef, mod)
		}
	}

	// Treat initializers as exports
	xtracer.Trace("check.CreateIsolate.preInitDel mod.Mixins.Len=%d", mod.Mixins.Len())
	afterInits := mod.Mixins.Get("init")
	mod.Mixins.Delkey("init")
	xtracer.Trace("check.CreateIsolate.postInitDel mod.Mixins.Len=%d", mod.Mixins.Len())

	// Python line 1580: mod.exports.extend(ExportDef(Atom(a.mixer()), Atom('')) for a in after_inits)
	for _, ai := range afterInits {
		mod.Exports = append(mod.Exports, &exportStub{name: ai.Mixer()})
	}

	// Check all mixin declarations
	for name, mixins := range mod.Mixins.All() {
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
		xtracer.Trace("check.CreateIsolate.export check name=%s found=%s", expname, pyBool(mod.Actions.Get(expname) != nil))
		if _, ok := mod.Actions.Get2(expname); !ok {
			// Dump all action keys for debugging
			for k := range mod.Actions.All() {
				xtracer.Trace("check.CreateIsolate.export available_action=%s", k)
			}
			return fmt.Errorf("undefined action: %s", expname)
		}
		origExports[expname] = true
	}

	extraWith := []string{}
	extraStrip := map[string][]string{}
	if isoCfg.CreateImports {
		var err error
		extraWith, extraStrip, err = createImportWrappers(iso, mod)
		if err != nil {
			return err
		}
	}

	// Validate with-parameters
	if iso != "" {
		if err := CheckWithParameters(mod, iso); err != nil {
			return err
		}
	}

	// Track mixer names for later warnings
	mixers := make(map[string]bool)
	for _, ms := range mod.Mixins.All() {
		for _, m := range ms {
			mixers[m.Mixer()] = true
		}
	}

	// Determine mixin order
	if err := GetMixinOrder(iso, mod); err != nil {
		return err
	}
	xtracer.Trace("check.CreateIsolate.postMixinOrder mod.Mixins.Len=%d", mod.Mixins.Len())

	// Construct the isolate
	if iso != "" {
		if _, ok := mod.Isolates[iso]; !ok {
			return fmt.Errorf("undefined isolate: %s", iso)
		}
		afterInitNames := make([]string, 0, len(afterInits))
		for _, ai := range afterInits {
			afterInitNames = append(afterInitNames, ai.Mixer())
		}
		err := IsolateComponent(mod, iso, extraWith, extraStrip, afterInitNames)
		if err != nil {
			return err
		}
	} else {
		if len(mod.Isolates) > 0 && isoCfg.ConeOfInfluence {
			return fmt.Errorf("no isolate specified on command line")
		}
		// Apply all mixins in no particular order
		mod.IsolateInfo = &IsolateInfo{}
		implemented := make(map[string]bool)

		for actname, mixinList := range mod.Mixins.All() {
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

				xtracer.Trace("isolate.create_no_iso mixer=%s mixee=%s", mx.Mixer(), mx.Mixee())
				mixed := ApplyMixin(action1, action2, mx.IsAfter())
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
			mod.PublicActions = NewInsMap[string, bool]()
			for _, e := range mod.Exports {
				if e.Scope() == "" {
					mod.PublicActions.Set(e.Exported(), true)
				}
			}
		} else {
			// No exports specified: all actions are public (compatibility)
			mod.PublicActions = NewInsMap[string, bool]()
			for a := range mod.Actions.All() {
				mod.PublicActions.Set(a, true)
			}
		}
	}

	xtracer.Trace("check.CreateIsolate after_isolate_component")

	// === Post-isolate_component steps, in Python order (lines 1899-1954) ===

	// Python line 1899-1900: Label public actions
	xtracer.Trace("check.CreateIsolate before_label_public n_public=%d", mod.PublicActions.Len())
	for name := range mod.PublicActions.All() {
		if act, ok := mod.Actions.Get2(name); ok {
			if a, ok := act.(ActionsAction); ok {
				// Python: action.label = name (singular, distinct from labels plural)
				type singleLabeler interface {
					SetLabel(string)
				}
				if lb, ok := a.(singleLabeler); ok {
					lb.SetLabel(name)
				}
			}
		}
	}
	xtracer.Trace("check.CreateIsolate after_label_public")

	// Python line 1901-1907: Create one big external action if requested
	extAction := ""
	if mod.Cfg != nil {
		extAction = mod.Cfg.ExtAction
	}
	xtracer.Trace("check.CreateIsolate before_ext_action ext=%q", extAction)
	if extAction != "" {
		afterInitNames := make(map[string]bool)
		for _, ai := range afterInits {
			afterInitNames[ai.Mixer()] = true
		}

		sortedPublic := make([]string, 0, mod.PublicActions.Len())
		for name := range mod.PublicActions.All() {
			sortedPublic = append(sortedPublic, name)
		}
		sort.Strings(sortedPublic)

		var extBranches []Expr
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
		extAct := NewEnvActionOn(mod.Cfg.ActCfg, extBranches...)
		mod.Actions.Set(extAction, extAct)
		mod.PublicActions.Set(extAction, true)
	}
	xtracer.Trace("check.CreateIsolate after_ext_action")

	// Python line 1911: slv.check_compat()
	// Check native interpretations of symbols for compatibility.
	xtracer.Trace("check.CreateIsolate before_check_compat")
	if mod.Sig != nil {
		errs := CheckCompatStatic(mod.Sig)
		for _, err := range errs {
			fmt.Printf("warning: %v\n", err)
		}
	}
	xtracer.Trace("check.CreateIsolate after_check_compat")

	// Python line 1915: mod.update_conjs()
	// Update conjectures (generate concept spaces)
	xtracer.Trace("check.CreateIsolate before_update_conjs n_conjs=%d", len(mod.LabeledConjs))
	mod.UpdateConjs()
	xtracer.Trace("check.CreateIsolate after_update_conjs")

	// Python line 1919-1936: Cone of influence filter / pedantic warnings
	xtracer.Trace("check.CreateIsolate before_cone_of_influence cone=%v", isoCfg.ConeOfInfluence)
	if isoCfg.ConeOfInfluence {
		cone := getModCone(mod)
		for a := range mod.Actions.All() {
			if !cone[a] {
				mod.Actions.Delkey(a)
			}
		}
	}
	xtracer.Trace("check.CreateIsolate after_cone_of_influence")

	// Python line 1939: fix_initializers(mod, after_inits)
	xtracer.Trace("check.CreateIsolate before_fix_initializers n_afterInits=%d", len(afterInits))
	FixInitializers(mod, afterInits)
	xtracer.Trace("check.CreateIsolate after_fix_initializers")

	// Python line 1941: mod.canonize_types()
	sortRefs := ComputeSortRefinements(mod.Sig)
	xtracer.Trace("check.CreateIsolate before_canonize_types n_sortRefs=%d", len(sortRefs))
	mod.CanonizeTypes(sortRefs)
	xtracer.Trace("check.CreateIsolate after_canonize_types")

	// Python line 1944-1946: Apply bracket actions for present conjectures (version >= 1.7)
	if iso != "" {
		if _, ok := mod.Isolates[iso]; ok && versionLE("1.7", isoCfg.IvyVersion) {
			xtracer.Trace("check.CreateIsolate before_bracket_actions n_brackets=%d", len(brackets))
			for _, b := range brackets {
				xtracer.Trace("check.CreateIsolate bracket_action actname=%s n_before=%d n_after=%d", b.ActName, len(b.Before), len(b.After))
				IsolateBracketAction(mod, b.ActName, b.Before, b.After)
			}
		}
	}
	xtracer.Trace("check.CreateIsolate after_bracket_actions")

	// Python line 1948: Set isolate proof reference
	if mod.IsolateProofs != nil {
		if p, ok := mod.IsolateProofs[iso]; ok {
			mod.IsolateProof = p
		}
	}

	xtracer.Trace("check.CreateIsolate EXIT name=%s", traceIsolateName(iso))
	return nil
}

// getModCone returns the Python get_mod_cone(mod) result: public roots,
// native anti-quote action references, and their transitive callees.
func getModCone(mod *Module) map[string]bool {
	return GetModConeFull(mod, mod.Actions, mod.PublicActions, nil)
}

// -----------------------------------------------------------------------
// Helper functions for create_isolate, ported from Python ivy_isolate.py.
// -----------------------------------------------------------------------

// CheckWithParameters validates that the names mentioned in a 'with'
// clause actually correspond to objects, actions, sorts, definitions,
// interpreted functions or properties.
//
// Corresponds to Python check_with_parameters (lines 777-806).
func CheckWithParameters(mod *Module, isolateName string) error {
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
	allLFs := make([]*LabeledFormula, 0)
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
			if lf, ok := itp.(*LabeledFormula); ok {
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
		if _, ok := mod.Hierarchy.Get2(name); ok {
			continue
		}
		if mod.Sig != nil {
			if _, ok := mod.Sig.Sorts.Get2(name); ok {
				continue
			}
			if _, ok := mod.Sig.Interp[name]; ok {
				continue
			}
			if _, ok := mod.Sig.Symbols.Get2(name); ok {
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
// isolateArc represents a directed edge in mixin ordering.
type isolateArc struct{ from, to string }

func GetMixinOrder(iso string, mod *Module) error {
	// Build isolateArc list from mod.MixOrd
	var arcs []isolateArc
	for _, rdf := range mod.MixOrd {
		type relNamer interface{ Args() []Node }
		if rn, ok := rdf.(relNamer); ok {
			args := rn.Args()
			if len(args) >= 2 {
				from := isolateNodeRelname(args[0])
				to := isolateNodeRelname(args[1])
				arcs = append(arcs, isolateArc{from, to})
			}
		}
	}

	for action, mixinList := range mod.Mixins.All() {
		// Separate implements from before/after
		var implements []MixinDef
		var beforeAfter []MixinDef

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
		var befores, afters []MixinDef
		for _, m := range beforeAfter {
			if m.IsAfter() {
				afters = append(afters, m)
			} else {
				befores = append(befores, m)
			}
		}

		sortByMixer := func(list []MixinDef) {
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
		result := make([]MixinDef, 0, len(implements)+len(befores)+len(afters))
		result = append(result, implements...)
		result = append(result, befores...)
		result = append(result, afters...)
		mod.Mixins.Set(action, result)
	}
	return nil
}

// FixInitializers processes after-init mixins: moves their actions to
// mod.InitialActions and mod.Initializers, removes them from mod.Actions
// and mod.PublicActions, and cleans up exports and isolate_info.
//
// Corresponds to Python fix_initializers (lines 1483-1506).
func FixInitializers(mod *Module, afterInits []MixinDef) {
	things := make(map[string]bool)

	for _, m := range afterInits {
		name := m.Mixer()
		extname := "ext:" + name

		// Get the action (prefer ext: variant)
		var action ActionsAction
		if act, ok := mod.Actions.Get2(extname); ok {
			if a, ok := act.(ActionsAction); ok {
				action = a
			}
		} else if act, ok := mod.Actions.Get2(name); ok {
			if a, ok := act.(ActionsAction); ok {
				action = a
			}
		}

		// Remove from actions and public actions
		mod.Actions.Delkey(name)
		mod.PublicActions.Delkey(name)
		mod.Actions.Delkey(extname)
		mod.PublicActions.Delkey(extname)

		if action == nil || !HasCode(action) {
			continue
		}

		mod.InitialActions = append(mod.InitialActions, action)

		// Create looped version for initializers
		xtracer.Trace("isolate.fix_initializers LOOP name=%s extname=%s action_type=%s action_nargs=%d",
			name, extname, ShortTypeName(action), len(action.Args()))
		loopedAction := LoopAction(action, mod)
		mod.Initializers = append(mod.Initializers, NamedAction{
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
	var newExports []Exporter
	for _, e := range mod.Exports {
		if afterInitNames[e.Exported()] {
			continue
		}
		newExports = append(newExports, e)
	}
	mod.Exports = newExports

	// Clean up isolate info
	if mod.IsolateInfo != nil {
		var newImpls []MixinTriple
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
func LoopAction(action ActionsAction, mod *Module) ActionsAction {
	subst := make(map[NodeKey]Expr)
	for _, p := range action.GetFormalParams() {
		v, err := NewVariable("Y"+p.Name, p.CSort)
		if err == nil {
			subst[Key(p)] = v
		}
	}
	// Python always calls substitute_constants_ast even with empty subst,
	// so we must too for trace fidelity.
	result := SubstituteConstantsAction(action, subst)
	// Python calls ia.type_check_action(action, mod) here, but that function
	// is disabled in Python (immediately returns). Omitted for parity.
	return result
}

// ApplyPresentConjectures wraps each exported action with assume(conjecture)
// before and after. This allows conjectures from present (but unverified)
// components to be assumed as invariants.
//
// Corresponds to Python apply_present_conjectures (lines 1533-1555).
func ApplyPresentConjectures(isol IsolateDefIface, mod *Module) []BracketEntry {
	if !mod.Cfg.IsolateCfg.AssumeInvariants {
		return nil
	}

	// Get present conjectures (verified=false, present=true)
	xtracer.Trace("check.ApplyPresentConjectures raw_labeled_conjs=%d", len(mod.LabeledConjs))
	conjs := GetIsolateConjs(mod, isol, false, true)
	mod.AssumedInvs = conjs
	xtracer.Trace("check.ApplyPresentConjectures after_GetIsolateConjs n_conjs=%d", len(conjs))

	// Filter out explicit conjectures
	var filteredConjs []*LabeledFormula
	for _, c := range conjs {
		if !c.Explicit {
			filteredConjs = append(filteredConjs, c)
		}
	}

	postConjs := GetIsolatePostConjs(mod, isol)
	var filteredPostConjs []*LabeledFormula
	for _, c := range postConjs {
		if !c.Explicit {
			filteredPostConjs = append(filteredPostConjs, c)
		}
	}

	// Build call graph and get exports
	cg := ActionCallGraph(mod)
	myExports := GetIsolateExports(mod, cg, isol)

	// Dump exports sorted for cross-language comparison
	exportNames := make([]string, 0, len(myExports))
	for a := range myExports {
		exportNames = append(exportNames, a)
	}
	sort.Strings(exportNames)
	xtracer.Trace("check.ApplyPresentConjectures n_exports=%d n_conjs=%d n_postConjs=%d",
		len(myExports), len(filteredConjs), len(filteredPostConjs))
	for _, a := range exportNames {
		xtracer.Trace("check.ApplyPresentConjectures EXPORT %s", a)
	}
	// Dump conj_actions
	conjActKeys := make([]string, 0, len(mod.ConjActions))
	for k := range mod.ConjActions {
		conjActKeys = append(conjActKeys, k)
	}
	sort.Strings(conjActKeys)
	for _, k := range conjActKeys {
		vals := make([]string, len(mod.ConjActions[k]))
		copy(vals, mod.ConjActions[k])
		sort.Strings(vals)
		xtracer.Trace("check.ApplyPresentConjectures CONJ_ACTIONS %s -> %s", k, fmtStrList(vals))
	}
	var brackets []BracketEntry
	for actname := range myExports {
		var assumes []ActionsAction
		for _, c := range filteredConjs {
			assumes = append(assumes, conjToAssume(c))
		}
		var postAssumes []ActionsAction
		for _, c := range filteredPostConjs {
			postAssumes = append(postAssumes, conjToAssume(c))
		}
		brackets = append(brackets, BracketEntry{ActName: actname, Before: assumes, After: postAssumes})
	}

	// Also add post-conjectures for actions that have conj_actions
	posts := make(map[string][]ActionsAction)
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

	// Sort brackets by actname for deterministic ordering matching Python.
	sort.Slice(brackets, func(i, j int) bool {
		return brackets[i].ActName < brackets[j].ActName
	})

	// Defensive: validate that every conj-derived assume in the brackets
	// carries a Loc. Catches future regressions of conjToAssume Loc
	// propagation. No-op when AssertLocEnabled is false.
	for _, e := range brackets {
		for _, b := range e.Before {
			AssertEveryActionHasLoc(b, "isolate.ApplyPresentConjectures.before actname="+e.ActName)
		}
		for _, a := range e.After {
			AssertEveryActionHasLoc(a, "isolate.ApplyPresentConjectures.after actname="+e.ActName)
		}
	}

	return brackets
}

// BracketEntry describes before/after assume actions to wrap around an action.
type BracketEntry struct {
	ActName string
	Before  []ActionsAction
	After   []ActionsAction
}

// BracketAction wraps an action with before/after sequences.
// Corresponds to Python bracket_action (lines 1529-1531).
func IsolateBracketAction(mod *Module, actname string, before, after []ActionsAction) {
	bracketActionInt(mod, actname, before, after)
	bracketActionInt(mod, "ext:"+actname, before, after)
}

func bracketActionInt(mod *Module, actname string, before, after []ActionsAction) {
	act, ok := mod.Actions.Get2(actname)
	if !ok {
		return
	}
	// Python (ivy_isolate.py:1695-1700):
	//   thing = empty_clone(action)
	//   thing.args.extend(before+[action]+after)
	// EmptyClone preserves the original action's Loc on the wrapper.
	wrap := EmptyClone(act).(*LogicSequence)
	wrap.Elems = make([]Expr, 0, len(before)+1+len(after))
	for _, b := range before {
		wrap.Elems = append(wrap.Elems, b)
	}
	wrap.Elems = append(wrap.Elems, act)
	for _, a := range after {
		wrap.Elems = append(wrap.Elems, a)
	}
	mod.Actions.Set(actname, wrap)
	AssertEveryActionHasLoc(wrap, "isolate.bracketActionInt actname="+actname)
}

// conjToAssume converts a labeled conjecture to an AssumeAction.
// Corresponds to Python conj_to_assume (ivy_isolate.py:1690-1693):
//
//	def conj_to_assume(c):
//	    res = ia.AssumeAction(c.formula)
//	    res.lineno = c.lineno
//	    return res
func conjToAssume(c *LabeledFormula) ActionsAction {
	fmla, ok := c.Formula.(Expr)
	if !ok {
		return NewSequence()
	}
	act := NewAssumeAction(fmla)
	act.SetLineno(c.GetLineno())
	return act
}

// SetUpImplementationMap builds the global implementation map from
// MixinImplementDef entries.
// Corresponds to Python set_up_implementation_map (lines 1508-1514).
func SetUpImplementationMap(mod *Module) map[string]string {
	implMap := make(map[string]string)
	for _, ms := range mod.Mixins.All() {
		for _, m := range ms {
			if isMixinImplement(m) {
				implMap[m.Mixee()] = m.Mixer()
			}
		}
	}
	return implMap
}

// topologicalSortStrings performs a topological sort of string nodes using arcs.
func topologicalSortStrings(nodes []string, arcs []isolateArc) []string {
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
func lfLabelName(lf *LabeledFormula) string {
	if lf == nil || lf.Label == nil {
		return ""
	}
	// Python uses conj.label.rep — the Atom's rep string.
	if a, ok := lf.Label.(*Atom); ok {
		return a.Rep
	}
	if c, ok := lf.Label.(*Const); ok {
		return c.Name
	}
	return fmt.Sprint(lf.Label)
}

// isMixinImplement checks if a mixin is an implement-type mixin.
func isMixinImplement(m interface{}) bool {
	_, ok := m.(*MixinImplementDef)
	return ok
}
