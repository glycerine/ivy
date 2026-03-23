// helpers.go implements helper functions for isolate_component, ported from
// Python ivy_isolate.py. These are the missing ~15 helper functions from §5.3.
package isolate

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// -----------------------------------------------------------------------
// AddMixinsExt: extended version of AddMixins with assert_to_assume and
// mod_mixin callbacks. Corresponds to Python add_mixins().
// -----------------------------------------------------------------------

// AddMixinsExt applies before/after mixins to an action.
//
// assertToAssume is called with each mixin to determine which assertion
// kinds should be converted to assumptions. It returns nil for no conversion.
// useMixin controls which mixins are applied (by mixer name).
// modMixin is called to transform mixer actions (e.g., prefix_calls).
//
// Corresponds to Python add_mixins (lines 40-54).
func AddMixinsExt(
	mod *module.Module,
	actname string,
	action2 actions.Action,
	assertToAssume func(interface{}) map[string]bool,
	useMixin func(string) bool,
	modMixin func(interface{}, actions.Action) actions.Action,
) actions.Action {
	res := action2
	if CreateImports {
		res = actions.DropInvariants(res)
	}
	mixins, ok := mod.Mixins[actname]
	if !ok {
		return res
	}
	for _, mx := range mixins {
		mixerName := mx.Mixer()
		action1, err := LookupAction(mod, mixerName)
		if err != nil {
			continue
		}
		if useMixin != nil && !useMixin(mixerName) {
			continue
		}
		if assertToAssume != nil {
			ata := assertToAssume(mx)
			if ata != nil && len(ata) > 0 {
				action1 = actions.AssertToAssume(action1, ata)
			}
		}
		if modMixin != nil {
			action1 = modMixin(mx, action1)
		}
		// Defensive: skip mixin if param counts don't match.
		// Python: apply_mixin raises IvyError (caught upstream). Go must not panic.
		// This can happen when CompileAction fails and registers a fallback with
		// placeholder params that don't match the mixin's expected signature.
		if len(action1.GetFormalParams()) != len(res.GetFormalParams()) ||
			len(action1.GetFormalReturns()) != len(res.GetFormalReturns()) {
			fmt.Fprintf(os.Stderr, "warning: skipping mixin %s for %s: param count mismatch (%d vs %d)\n",
				mixerName, actname,
				len(action1.GetFormalParams()), len(res.GetFormalParams()))
			continue
		}
		res = actions.ApplyMixin(action1, res, mx.IsAfter())
	}
	return res
}

// -----------------------------------------------------------------------
// set_privates: full Python-compatible version
// -----------------------------------------------------------------------

// SetPrivatesFull sets mod.Privates based on isolate definition and
// hierarchy attributes. This is the full version matching Python's
// set_privates() (lines 705-738).
func SetPrivatesFull(mod *module.Module, iso interface{}, suff string) {
	if mod.Privates == nil {
		mod.Privates = make(map[string]bool)
	}

	// Check for prefer_impls
	if suff == "" && PreferImpls {
		setPrivatesPrefer(mod, iso, "impl")
		return
	}

	// Determine suffix
	if suff == "" {
		// Check if this is an ExtractDef
		type extractor interface{ IsExtract() bool }
		if ex, ok := iso.(extractor); ok && ex.IsExtract() {
			suff = "spec"
		} else {
			suff = "impl"
		}
	}

	// Mark top-level suffix as private
	if _, ok := mod.Hierarchy[suff]; ok {
		mod.Privates[suff] = true
	}

	// Walk hierarchy
	for n, children := range mod.Hierarchy {
		nsuff := getPrivateFromAttributes(mod, n, suff)
		nsList := []string{nsuff}
		if nsuff == "priv" {
			nsList = []string{"impl", "spec"}
		}
		for _, ns := range nsList {
			if children[ns] {
				pname := iu.ComposeNames(n, ns)
				mod.Privates[pname] = true
			}
		}
	}

	// Handle explicit private attributes
	for name := range mod.Attributes {
		pc := iu.ParentChildName(name)
		p, c := pc[0], pc[1]
		if c == "spec" || c == "impl" || c == "private" {
			ppc := iu.ParentChildName(p)
			pp := ppc[0]
			nsuff := getPrivateFromAttributes(mod, pp, suff)
			if c == nsuff || nsuff == "priv" || c == "private" {
				mod.Privates[p] = true
			}
		}
	}

	// Set vprivates
	VPrivates = make(map[string]bool)
	for _, isol := range mod.Isolates {
		for _, v := range isol.VerifiedNames() {
			VPrivates[v] = true
		}
	}

	// Handle ProcessDef
	type processDef interface {
		IsProcess() bool
		ProcessName() string
	}
	if pd, ok := iso.(processDef); ok && pd.IsProcess() {
		for _, isol := range mod.Isolates {
			if pd2, ok2 := interface{}(isol).(processDef); ok2 && pd2.IsProcess() {
				if pd.ProcessName() != pd2.ProcessName() {
					mod.Privates[pd2.ProcessName()] = true
				}
			}
			for _, v := range isol.VerifiedNames() {
				VPrivates[v] = true
			}
		}
	}
}

func setPrivatesPrefer(mod *module.Module, iso interface{}, preferred string) {
	idef, ok := iso.(IsolateDefInterface)
	if !ok {
		return
	}
	verified := make(map[string]bool)
	for _, v := range idef.VerifiedNames() {

		verified[v] = true
	}
	suff := "impl"
	if preferred == "spec" {
		suff = "spec"
	} else {
		suff = "spec" // suff is the non-preferred
	}

	if !verified["this"] {
		if mod.Hierarchy[suff] != nil && mod.Hierarchy[preferred] != nil {
			mod.Privates[suff] = true
		}
	}
	for n, children := range mod.Hierarchy {
		if !verified[n] {
			if children[suff] && children[preferred] {
				mod.Privates[iu.ComposeNames(n, suff)] = true
			}
		}
	}
}

func getPrivateFromAttributes(mod *module.Module, name string, suff string) string {
	attrname := iu.ComposeNames(name, IsolateMode)
	if val, ok := mod.Attributes[attrname]; ok {
		aval := fmt.Sprint(val)
		switch aval {
		case "priv":
			return "priv"
		case "impl":
			return "spec"
		case "spec":
			return "impl"
		default:
			// Python line 700-701: raises error for invalid attribute values
			fmt.Fprintf(os.Stderr, "error: invalid isolate attribute value %q for %s\n", aval, attrname)
		}
	}
	return suff
}

// -----------------------------------------------------------------------
// get_isolate_info: full Python-compatible version
// -----------------------------------------------------------------------

// GetIsolateInfoFull computes verified and present sets from an isolate
// definition. This is the full version matching Python get_isolate_info
// (lines 809-837).
func GetIsolateInfoFull(mod *module.Module, iso interface{}, kind string, extraWith []string) (verified, present map[string]bool) {
	idef, ok := iso.(IsolateDefInterface)
	if !ok {
		return make(map[string]bool), make(map[string]bool)
	}

	verified = make(map[string]bool)
	present = make(map[string]bool)

	// Add verified names + extra_with
	for _, name := range idef.VerifiedNames() {
		verified[name] = true
		present[name] = true
	}
	for _, name := range extraWith {
		verified[name] = true
		present[name] = true
	}

	// Add present names
	for _, name := range idef.PresentNames() {
		present[name] = true
	}

	// Add kind suffixes for verified names
	for _, name := range idef.VerifiedNames() {
		kindName := iu.ComposeNames(name, kind)
		verified[kindName] = true
		present[kindName] = true
	}

	// Handle attributes (kind or "private")
	vp := make(map[string]bool)
	for _, isol := range mod.Isolates {
		for _, v := range isol.VerifiedNames() {
			vp[v] = true
		}
	}

	for name := range mod.Attributes {
		pc := iu.ParentChildName(name)
		pName, c := pc[0], pc[1]
		if c == kind || c == "private" {
			isIso := vp[pName]
			var recur func(string)
			recur = func(p1 string) {
				p1parts := iu.ParentChildName(p1)
				parent := p1parts[0]
				if verified[parent] {
					if !isIso {
						verified[pName] = true
					}
					present[pName] = true
				} else if parent != "this" && !vp[parent] {
					recur(parent)
				}
			}
			recur(pName)
		}
	}

	return verified, present
}

// -----------------------------------------------------------------------
// follow_definitions: transitively follow definition dependencies
// -----------------------------------------------------------------------

// FollowDefinitions transitively adds all symbols referenced by definitions
// of symbols already in allSyms. Corresponds to Python follow_definitions
// (lines 847-850).
func FollowDefinitions(ldfs []*ast.LabeledFormula, allSyms map[string]bool) {
	// Build map from defined symbol name to RHS
	dmap := make(map[string]lg.Expr)
	for _, ldf := range ldfs {
		if ldf.Formula == nil {
			continue
		}
		// Definition: lhs = rhs, where lhs is an Apply or Const
		children := ldf.Formula.(lg.Expr).Children()
		if len(children) < 2 {
			continue
		}
		defSym := definedSymbolName(children[0])
		if defSym != "" {
			dmap[defSym] = children[1]
		}
	}
	// For each symbol already in allSyms, follow its definition
	for sym := range copyStringSet(allSyms) {
		followDefinitionsRec(sym, dmap, allSyms, make(map[string]bool))
	}
}

func followDefinitionsRec(sym string, dmap map[string]lg.Expr, allSyms, memo map[string]bool) {
	allSyms[sym] = true
	if rhs, ok := dmap[sym]; ok && !memo[sym] {
		memo[sym] = true
		for _, s := range usedSymbolNames(rhs) {
			followDefinitionsRec(s, dmap, allSyms, memo)
		}
	}
}

func definedSymbolName(node lg.Expr) string {
	if c, ok := node.(*lg.Symbol); ok {
		return c.Name
	}
	if app, ok := node.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Symbol); ok {
			return c.Name
		}
	}
	return ""
}

func usedSymbolNames(node lg.Expr) []string {
	syms := make(map[string]bool)
	collectUsedSymbolNames(node, syms)
	result := make([]string, 0, len(syms))
	for s := range syms {
		result = append(result, s)
	}
	return result
}

func collectUsedSymbolNames(node lg.Expr, syms map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Symbol); ok {
		syms[c.Name] = true
	}
	if app, ok := node.(*lg.Apply); ok {
		collectUsedSymbolNames(app.Func, syms)
	}
	for _, child := range node.Children() {
		collectUsedSymbolNames(child, syms)
	}
}

func copyStringSet(s map[string]bool) map[string]bool {
	c := make(map[string]bool, len(s))
	for k, v := range s {
		c[k] = v
	}
	return c
}

// -----------------------------------------------------------------------
// get_prop_dependencies: get property-to-object dependencies
// -----------------------------------------------------------------------

// GetPropDependencies returns a list of (property, dependency names) pairs.
// Each property's proof depends on the objects listed.
// Corresponds to Python get_prop_dependencies (lines 659-683).
func GetPropDependencies(mod *module.Module) []PropDep {
	// Build depmap: for each verified name, map to all verified+present names
	depmap := make(map[string][]string)
	for _, isol := range mod.Isolates {
		allNames := append(isol.VerifiedNames(), isol.PresentNames()...)
		for _, v := range isol.VerifiedNames() {
			depmap[v] = append(depmap[v], allNames...)
		}
	}

	// Collect all object names from axioms and interpretations
	objs := make(map[string]bool)
	for _, ax := range mod.LabeledAxioms {
		if ax.Label != nil {
			name := lfLabelName(ax)
			for _, anc := range Ancestors(name) {
				objs[anc] = true
			}
		}
	}
	for _, itps := range mod.Interps {
		for _, itp := range itps {
			if lf, ok := itp.(*ast.LabeledFormula); ok {
				name := lfLabelName(lf)
				if name != "" {
					for _, anc := range Ancestors(name) {
						objs[anc] = true
					}
				}
			}
		}
	}

	// For each property, collect dependencies
	var result []PropDep
	for _, prop := range mod.LabeledProps {
		if prop.Label == nil {
			continue
		}
		name := lfLabelName(prop)
		var ds []string
		for _, anc := range specAncestors(name) {
			for _, d := range depmap[anc] {
				if objs[d] {
					ds = append(ds, d)
				}
			}
		}
		result = append(result, PropDep{Prop: prop, Deps: ds})
	}
	return result
}

// PropDep represents a property and its proof dependencies.
type PropDep struct {
	Prop *ast.LabeledFormula
	Deps []string
}

// specAncestors returns ancestors, skipping "spec" children.
// Corresponds to Python spec_ancestors().
func specAncestors(name string) []string {
	var result []string
	s := name
	for {
		result = append(result, s)
		idx := strings.LastIndex(s, iu.ComposeCharacter)
		if idx < 0 {
			break
		}
		child := s[idx+len(iu.ComposeCharacter):]
		s = s[:idx]
		if child == "spec" {
			break
		}
	}
	return result
}

// -----------------------------------------------------------------------
// get_props_proved_in_isolate
// -----------------------------------------------------------------------

// GetPropsProvedInIsolate classifies properties as proved or not-proved
// in the given isolate. Corresponds to Python get_props_proved_in_isolate
// (lines 752-774).
func GetPropsProvedInIsolate(mod *module.Module, iso interface{}) (proved, notProved []*ast.LabeledFormula) {
	if versionLE(IvyVersion, "1.6") {
		return getPropsProvedInIsolateOrig(mod, iso)
	}

	// Save and temporarily modify privates
	savePrivates := mod.Privates
	mod.Privates = make(map[string]bool)
	SetPrivatesFull(mod, iso, "impl")
	verified, _ := GetIsolateInfoFull(mod, iso, "impl", nil)

	// For version > 1.6: mark verified names from other isolates as private
	idef, _ := iso.(IsolateDefInterface)
	for _, otherIso := range mod.Isolates {
		if interface{}(otherIso) == iso {
			continue
		}
		for _, ovn := range otherIso.VerifiedNames() {
			if StartsWithSome(ovn, verified, mod, nil) {
				mod.Privates[ovn] = true
			}
		}
	}

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

	// Remove subgoals from not_proved
	subs := make(map[int64]bool)
	for _, sg := range mod.Subgoals {
		for _, sub := range sg.Subgoals {
			subs[sub.ID] = true
		}
	}
	var filteredNotProved []*ast.LabeledFormula
	for _, p := range notProved {
		if !subs[p.ID] {
			filteredNotProved = append(filteredNotProved, p)
		}
	}
	notProved = filteredNotProved

	// Also get names from idef for sub-isolate filtering
	_ = idef
	return proved, notProved
}

func getPropsProvedInIsolateOrig(mod *module.Module, iso interface{}) (proved, notProved []*ast.LabeledFormula) {
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

// -----------------------------------------------------------------------
// add_extern_precond
// -----------------------------------------------------------------------

// AddExternPrecond adds preconditions from call arguments to the
// preconds list. Corresponds to Python add_extern_precond (lines 876-884).
func AddExternPrecond(mod *module.Module, callee actions.Action, callArgs []lg.Expr, preconds *[]lg.Expr) {
	calleeAct, ok := callee.(actions.Action)
	if !ok {
		return
	}
	formalParams := calleeAct.GetFormalParams()
	var conjs []lg.Expr
	for i, fml := range formalParams {
		if i >= len(callArgs) {
			break
		}
		act := callArgs[i]
		if isNumeralOrConstructor(act, mod) {
			conjs = append(conjs, &lg.Eq{T1: fml, T2: act})
		}
	}
	if len(conjs) == 0 {
		// Anything or true is true: clear preconds
		*preconds = nil
	} else {
		and := makeAndH(conjs...)
		*preconds = append(*preconds, and)
	}
}

func isNumeralOrConstructor(node lg.Expr, mod *module.Module) bool {
	if c, ok := node.(*lg.Symbol); ok {
		// Check if it's a constructor
		if _, ok := mod.ConstructorSorts[c.Name]; ok {
			return true
		}
		// Check if it's a numeral (starts with digit or is 0)
		if len(c.Name) > 0 && c.Name[0] >= '0' && c.Name[0] <= '9' {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------
// get_mod_cone: full version with roots and after_inits
// -----------------------------------------------------------------------

// GetModConeFull returns the cone of action names reachable from roots.
// Actions referenced by natives and initializers are also included.
// Corresponds to Python get_mod_cone (lines 1463-1475).
func GetModConeFull(mod *module.Module, actionsMap map[string]actions.Action,
	roots map[string]bool, afterInits []string) map[string]bool {

	cone := make(map[string]bool)

	// Start with roots
	for name := range roots {
		cone[name] = true
	}

	// Add after-init actions
	for _, ai := range afterInits {
		cone[ai] = true
		cone["ext:"+ai] = true
	}

	// Add actions referenced by natives
	for _, nat := range mod.Natives {
		if lf, ok := nat.(*ast.LabeledFormula); ok {
			n := lfLabelName(lf)
			if n != "" {
				cone[n] = true
			}
		}
	}

	// Transitively follow calls
	changed := true
	for changed {
		changed = false
		for name := range copyStringSet(cone) {
			act, ok := actionsMap[name]
			if !ok {
				continue
			}
			for _, callee := range act.IterCalls() {
				if !cone[callee] {
					cone[callee] = true
					changed = true
				}
				extName := "ext:" + callee
				if _, ok := actionsMap[extName]; ok && !cone[extName] {
					cone[extName] = true
					changed = true
				}
			}
		}
	}

	return cone
}

// -----------------------------------------------------------------------
// has_side_effect: Python-compatible version
// -----------------------------------------------------------------------

// HasSideEffectFull checks if an action has side effects on the module
// signature. Follows through calls transitively.
// Corresponds to Python has_side_effect (lines 458-479).
func HasSideEffectFull(mod *module.Module, newActions map[string]actions.Action, actname string) bool {
	return HasSideEffect(mod, actname, newActions)
}

// -----------------------------------------------------------------------
// Utility: IsolateDefNode interface for the actual isolate AST node
// -----------------------------------------------------------------------

// IsolateDefNode extends IsolateDefInterface with additional methods
// needed by isolate_component.
type IsolateDefNode interface {
	IsolateDefInterface
	// Params returns the isolate parameters.
	Params() []*lg.Symbol
	// WithArgs returns the number of with-clause arguments.
	WithArgs() int
}

// makeAndH creates an And node, ignoring sort errors.
func makeAndH(terms ...lg.Expr) lg.Expr {
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

// makeKindSet creates a map[string]bool from action type names.
// Convenience for building assert_to_assume kind sets.
func makeKindSet(names ...string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}

// -----------------------------------------------------------------------
// get_strip_binding: collect strip parameter bindings from an AST node
// -----------------------------------------------------------------------

// GetStripBinding walks an AST node and collects strip parameter bindings
// into the stripBinding map. For each function application or atom with a
// name in the strip map, it maps the actual arguments to the corresponding
// strip parameters.
// Corresponds to Python get_strip_binding (lines 291-301).
func GetStripBinding(node lg.Expr, stripMap StripMap, stripBinding map[lg.NodeKey]string, mod *module.Module) error {
	if node == nil {
		return nil
	}
	// Recurse into children first
	for _, child := range node.Children() {
		if child == nil {
			continue
		}
		if err := GetStripBinding(child, stripMap, stripBinding, mod); err != nil {
			return err
		}
	}
	// Get the name of this node (if it's an application or constant)
	name := ""
	var args []lg.Expr
	switch n := node.(type) {
	case *lg.Apply:
		if c, ok := n.Func.(*lg.Symbol); ok {
			name = c.Name
		}
		args = n.Terms
	case *lg.Symbol:
		name = n.Name
		args = nil
	}
	if name == "" {
		return nil
	}
	stripParams := StripMapLookup(name, stripMap, mod)
	if len(stripParams) == 0 {
		return nil
	}
	if len(args) < len(stripParams) {
		return fmt.Errorf("cannot strip isolate parameters from %s", name)
	}
	for i, sp := range stripParams {
		ap := args[i]
		if existing, ok := stripBinding[lg.Key(ap)]; ok && existing != sp {
			return fmt.Errorf("cannot strip parameter %v from %s", ap, name)
		}
		stripBinding[lg.Key(ap)] = sp
	}
	return nil
}

// -----------------------------------------------------------------------
// has_unsummarized_mixins: check if action has unsummarized mixins of given kind
// -----------------------------------------------------------------------

// MixinKind represents the type of mixin (before or after).
type MixinKind int

const (
	MixinKindBefore MixinKind = iota
	MixinKindAfter
)

// HasUnsummarizedMixins checks whether any mixin of the given kind for
// actname is not in summarized_actions.
// Corresponds to Python has_unsummarized_mixins (lines 523-525).
func HasUnsummarizedMixins(mod *module.Module, actname string, summarizedActions map[string]bool, kind MixinKind) bool {
	mixins, ok := mod.Mixins[actname]
	if !ok {
		return false
	}
	for _, mx := range mixins {
		// Check if this mixin is of the right kind
		isAfter := mx.IsAfter()
		if kind == MixinKindBefore && isAfter {
			continue
		}
		if kind == MixinKindAfter && !isAfter {
			continue
		}
		// Check if the mixer is NOT summarized
		if !summarizedActions[mx.Mixer()] {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------
// get_callouts_action / get_callouts: collect callout information
// -----------------------------------------------------------------------

// Callouts represents the 4-tuple of callout sets for an action.
// Index 0: !head && !tail, 1: head && !tail, 2: !head && tail, 3: head && tail
type Callouts [4]map[string]bool

// NewCallouts creates a new Callouts with initialized sets.
func NewCallouts() Callouts {
	return Callouts{
		make(map[string]bool),
		make(map[string]bool),
		make(map[string]bool),
		make(map[string]bool),
	}
}

// GetCalloutsAction recursively collects callout information from an action.
// head indicates whether the action is at the beginning of its parent sequence.
// tail indicates whether it's at the end.
// Corresponds to Python get_callouts_action (lines 527-548).
func GetCalloutsAction(
	mod *module.Module,
	newActions map[string]actions.Action,
	summarizedActions map[string]bool,
	callouts map[string]Callouts,
	action actions.Action,
	acallouts *Callouts,
	head, tail bool,
) {
	switch a := action.(type) {
	case *actions.Sequence:
		for idx, child := range a.Children {
			subAct := actions.UnwrapAction(child)
			if subAct == nil {
				if act, ok := child.(actions.Action); ok {
					subAct = act
				}
			}
			if subAct != nil {
				GetCalloutsAction(mod, newActions, summarizedActions, callouts, subAct, acallouts,
					head && idx == 0, tail && idx == len(a.Children)-1)
			}
		}
	case *actions.CallAction:
		calledName := a.CalleeName()
		if summarizedActions[calledName] {
			h := head
			t := tail
			if HasUnsummarizedMixins(mod, calledName, summarizedActions, MixinKindBefore) {
				h = false
			}
			if HasUnsummarizedMixins(mod, calledName, summarizedActions, MixinKindAfter) {
				t = false
			}
			// Compute index: h=1bit, t=1bit → 0..3
			// Python: (3 if tail else 1) if head else (2 if tail else 0)
			// Must use mutated h and t, not the original head/tail parameters.
			var idx int
			if h {
				if t {
					idx = 3
				} else {
					idx = 1
				}
			} else {
				if t {
					idx = 2
				} else {
					idx = 0
				}
			}
			acallouts[idx][calledName] = true
		} else {
			GetCallouts(mod, newActions, summarizedActions, calledName, callouts)
			// Merge callee's callouts into ours
			if calleeCO, ok := callouts[calledName]; ok {
				for i := 0; i < 4; i++ {
					for k := range calleeCO[i] {
						acallouts[i][k] = true
					}
				}
			}
		}
	default:
		// For other action types, recurse into sub-actions
		for _, arg := range action.ActionArgs() {
			if subAct, ok := arg.(actions.Action); ok {
				GetCalloutsAction(mod, newActions, summarizedActions, callouts, subAct, acallouts, head, tail)
			} else if w := actions.UnwrapAction(arg); w != nil {
				GetCalloutsAction(mod, newActions, summarizedActions, callouts, w, acallouts, head, tail)
			}
		}
	}
}

// GetCallouts computes callout information for a named action.
// Corresponds to Python get_callouts (lines 551-557).
func GetCallouts(
	mod *module.Module,
	newActions map[string]actions.Action,
	summarizedActions map[string]bool,
	actname string,
	callouts map[string]Callouts,
) {
	if _, ok := callouts[actname]; ok {
		return // already computed
	}
	if summarizedActions[actname] {
		return
	}
	acallouts := NewCallouts()
	callouts[actname] = acallouts
	action, ok := newActions[actname]
	if !ok {
		return
	}
	GetCalloutsAction(mod, newActions, summarizedActions, callouts, action, &acallouts, true, true)
	callouts[actname] = acallouts
}

// -----------------------------------------------------------------------
// get_loc_mods: get locally modified symbols
// -----------------------------------------------------------------------

// GetLocMods returns the symbols modified by an action whose names start
// with 'fml:' (i.e., formal/local symbols).
// Corresponds to Python get_loc_mods (lines 560-563).
func GetLocMods(mod *module.Module, actname string) []string {
	act, ok := mod.Actions[actname]
	if !ok {
		return nil
	}
	modSet := actions.Modifies(act)
	var result []string
	for _, sym := range modSet {
		if strings.HasPrefix(sym.Name, "fml:") {
			result = append(result, sym.Name)
		}
	}
	return result
}

// -----------------------------------------------------------------------
// find_references: find line numbers referencing given symbols
// -----------------------------------------------------------------------

// FindReferences returns the set of line numbers in the module's axioms,
// properties, inits, conjectures, definitions, and actions that reference
// any of the given symbol names.
// Corresponds to Python find_references (lines 565-573).
func FindReferences(mod *module.Module, syms map[string]bool, newActions map[string]actions.Action) map[int]bool {
	refs := make(map[int]bool)

	// Check labeled formulas
	allFormulas := make([]*ast.LabeledFormula, 0)
	allFormulas = append(allFormulas, mod.LabeledAxioms...)
	allFormulas = append(allFormulas, mod.LabeledProps...)
	allFormulas = append(allFormulas, mod.LabeledInits...)
	allFormulas = append(allFormulas, mod.LabeledConjs...)
	allFormulas = append(allFormulas, mod.Definitions...)

	for _, lf := range allFormulas {
		if lf.Formula == nil {
			continue
		}
		fSyms := usedSymbolNames(lf.Formula.(lg.Expr))
		for _, s := range fSyms {
			if syms[s] {
				refs[lf.Lineno] = true
				break
			}
		}
	}

	// Check actions
	for _, act := range newActions {
		actSyms := collectActionSymNames(act)
		for s := range actSyms {
			if syms[s] {
				loc := act.GetLineno()
				refs[loc.Line] = true
				break
			}
		}
	}

	return refs
}

// collectActionSymNames collects all constant symbol names referenced by an action.
func collectActionSymNames(act actions.Action) map[string]bool {
	syms := make(map[string]bool)
	for _, arg := range act.ActionArgs() {
		collectUsedSymbolNames(arg, syms)
	}
	// Also recurse into sub-actions
	for _, sub := range act.IterSubactions() {
		if sub == act {
			continue // skip self to avoid infinite loop
		}
		for _, arg := range sub.ActionArgs() {
			collectUsedSymbolNames(arg, syms)
		}
	}
	return syms
}

// -----------------------------------------------------------------------
// hide_action_params: wrap action with LocalAction hiding formals
// -----------------------------------------------------------------------

// HideActionParams wraps an action in a LocalAction that hides its
// formal parameters and returns.
// Corresponds to Python hide_action_params (lines 1438-1441).
func HideActionParams(action actions.Action) actions.Action {
	params := action.GetFormalParams()
	returns := action.GetFormalReturns()

	// Build locals list: params + returns
	var locals []lg.Expr
	for _, p := range params {
		locals = append(locals, p)
	}
	for _, r := range returns {
		locals = append(locals, r)
	}

	// Create LocalAction with locals + body (action wrapped as node)
	args := make([]lg.Expr, 0, len(locals)+1)
	args = append(args, locals...)
	args = append(args, actions.WrapAction(action))
	return actions.NewLocalAction(args...)
}
