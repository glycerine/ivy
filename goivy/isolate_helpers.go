// helpers.go implements helper functions for isolate_component, ported from
// Python ivy_isolate.py. These are the missing ~15 helper functions from §5.3.
package goivy

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// traceSymSet dumps the contents of a symbol set in insertion order using per-symbol
// traces. Matches Python: for x in syms: xtracer.trace(...)
// Caller MUST guard with `if xtracer.Enabled {}` so the compiler eliminates this
// when tracing is disabled.
func traceSymSet(label string, syms *InsMap[NodeKey, Expr]) {
	for _, v := range syms.All() {
		xtracer.Trace("%s.sym %s", label, PrettyFmla(v))
	}
	xtracer.Trace("%s n=%d", label, syms.Len())
}

func filterInterfSymsForSig(allSyms *InsMap[NodeKey, Expr], sig *Sig) *InsMap[NodeKey, Expr] {
	result := NewInsMap[NodeKey, Expr]()
	if allSyms == nil || sig == nil {
		return result
	}

	sigSymbols := make(map[NodeKey]bool)
	for _, sym := range sig.AllSymbols() {
		sigSymbols[ConstSymKey(sym)] = true
	}
	for key, sym := range allSyms.All() {
		if sigSymbols[key] {
			result.Set(key, sym)
		}
	}
	return result
}

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
	mod *Module,
	actname string,
	action2 ActionsAction,
	assertToAssume func(interface{}) map[string]bool,
	useMixin func(string) bool,
	modMixin func(interface{}, ActionsAction) ActionsAction,
) ActionsAction {
	res := action2
	if mod.Cfg.IsolateCfg.CreateImports {
		res = DropInvariants(res)
	}
	mixins := mixinsAutoVivify(mod, actname)
	if len(mixins) == 0 {
		return res
	}
	for _, mx := range mixins {
		mixerName := mx.Mixer()
		xtracer.Trace("isolate.add_mixins_ext actname=%s mixer=%s", actname, mixerName)
		action1, err := LookupAction(mod, mixerName)
		if err != nil {
			panic(err)
		}
		if useMixin != nil && !useMixin(mixerName) {
			xtracer.Trace("isolate.add_mixins_ext SKIP use_mixin=false mixer=%s", mixerName)
			continue
		}
		if assertToAssume != nil {
			ata := assertToAssume(mx)
			if ata != nil && len(ata) > 0 {
				action1 = AssertToAssume(action1, ata, mod.Cfg.IuCfg)
			}
		}
		if modMixin != nil {
			action1 = modMixin(mx, action1)
		}
		res = ApplyMixin(action1, res, mx.IsAfter())
		if seq, ok := res.(*LogicSequence); ok {
			xtracer.Trace("isolate.add_mixins_ext AFTER_MIXIN actname=%s mixer=%s res_nargs=%d action1_nargs=%d",
				actname, mixerName, len(seq.Elems), len(action1.Args()))
		} else {
			xtracer.Trace("isolate.add_mixins_ext AFTER_MIXIN actname=%s mixer=%s res_type=%s action1_nargs=%d",
				actname, mixerName, ShortTypeName(res), len(action1.Args()))
		}
	}
	return res
}

// -----------------------------------------------------------------------
// set_privates: full Python-compatible version
// -----------------------------------------------------------------------

// SetPrivatesFull sets mod.Privates based on isolate definition and
// hierarchy attributes. This is the full version matching Python's
// set_privates() (lines 705-738).
func SetPrivatesFull(mod *Module, iso interface{}, suff string) {
	if mod.Privates == nil {
		mod.Privates = make(map[string]bool)
	}

	// Check for prefer_impls
	if suff == "" && mod.Cfg.IsolateCfg.PreferImpls {
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
	if _, ok := mod.Hierarchy.Get2(suff); ok {
		mod.Privates[suff] = true
	}

	// Walk hierarchy
	for n, children := range mod.Hierarchy.All() {
		nsuff := getPrivateFromAttributes(mod, n, suff)
		nsList := []string{nsuff}
		if nsuff == "priv" {
			nsList = []string{"impl", "spec"}
		}
		for _, ns := range nsList {
			if children.Get(ns) {
				pname := mod.Cfg.IuCfg.ComposeNames(n, ns)
				mod.Privates[pname] = true
			}
		}
	}

	// Handle explicit private attributes
	for name := range mod.Attributes.All() {
		pc := mod.Cfg.IuCfg.ParentChildName(name)
		p, c := pc[0], pc[1]
		if c == "spec" || c == "impl" || c == "private" {
			ppc := mod.Cfg.IuCfg.ParentChildName(p)
			pp := ppc[0]
			nsuff := getPrivateFromAttributes(mod, pp, suff)
			if c == nsuff || nsuff == "priv" || c == "private" {
				mod.Privates[p] = true
			}
		}
	}

	// Set vprivates on module
	mod.VPrivates = make(map[string]bool)
	for _, isol := range mod.Isolates {
		for _, v := range isol.VerifiedNames() {
			mod.VPrivates[v] = true
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
				mod.VPrivates[v] = true
			}
		}
	}
}

func setPrivatesPrefer(mod *Module, iso interface{}, preferred string) {
	idef, ok := iso.(IsolateDefIface)
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
		suffVal, _ := mod.Hierarchy.Get2(suff)
		prefVal, _ := mod.Hierarchy.Get2(preferred)
		if suffVal != nil && prefVal != nil {
			mod.Privates[suff] = true
		}
	}
	for n, children := range mod.Hierarchy.All() {
		if !verified[n] {
			if children.Get(suff) && children.Get(preferred) {
				mod.Privates[mod.Cfg.IuCfg.ComposeNames(n, suff)] = true
			}
		}
	}
}

func getPrivateFromAttributes(mod *Module, name string, suff string) string {
	attrname := mod.Cfg.IuCfg.ComposeNames(name, mod.Cfg.IsolateCfg.IsolateMode)
	if val, ok := mod.Attributes.Get2(attrname); ok {
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
func GetIsolateInfoFull(mod *Module, iso interface{}, kind string, extraWith []string) (verified, present map[string]bool) {
	idef, ok := iso.(IsolateDefIface)
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
		kindName := mod.Cfg.IuCfg.ComposeNames(name, kind)
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

	for name := range mod.Attributes.All() {
		pc := mod.Cfg.IuCfg.ParentChildName(name)
		pName, c := pc[0], pc[1]
		if c == kind || c == "private" {
			isIso := vp[pName]
			var recur func(string)
			recur = func(p1 string) {
				p1parts := mod.Cfg.IuCfg.ParentChildName(p1)
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
func FollowDefinitions(ldfs []*LabeledFormula, allSyms *InsMap[NodeKey, Expr]) {
	FollowDefinitionsLabeled("", ldfs, allSyms)
}

func FollowDefinitionsLabeled(label string, ldfs []*LabeledFormula, allSyms *InsMap[NodeKey, Expr]) {
	before := allSyms.Len()
	// Build map from defined symbol Sexp key to RHS.
	// Python: dmap = dict((ldf.formula.args[0].rep, ldf.formula.args[1]) for ldf in ldfs)
	// Python uses full Const object as key; we use Sexp() for equivalent structural identity.
	dmap := make(map[NodeKey]Expr)
	for _, ldf := range ldfs {
		if ldf.Formula == nil {
			continue
		}
		fmla, ok := ldf.Formula.(Expr)
		if !ok {
			continue
		}
		children := fmla.Children()
		if len(children) < 2 {
			continue
		}
		if c := definedSymbolConst(children[0]); c != nil {
			dmap[ConstSymKey(c)] = children[1]
		}
	}
	// For each symbol already in allSyms, follow its definition
	for k, v := range copySymSet(allSyms).All() {
		followDefinitionsRec(k, v, dmap, allSyms, make(map[NodeKey]bool))
	}
	if label != "" {
		xtracer.Trace("isolate.FollowDefinitions.%s before=%d after=%d", label, before, allSyms.Len())
	} else {
		xtracer.Trace("isolate.FollowDefinitions before=%d after=%d", before, allSyms.Len())
	}
}

func followDefinitionsRec(key NodeKey, expr Expr, dmap map[NodeKey]Expr, allSyms *InsMap[NodeKey, Expr], memo map[NodeKey]bool) {
	allSyms.Set(key, expr)
	if rhs, ok := dmap[key]; ok && !memo[key] {
		memo[key] = true
		// Use SymbolsIluAst directly (no traces) matching Python's:
		//   for s in lu.used_symbols_ast(dmap[sym]): follow_definitions_rec(s,...)
		for sym := range SymbolsIluAst(rhs) {
			followDefinitionsRec(Key(sym), sym, dmap, allSyms, memo)
		}
	}
}

// definedSymbolConst returns the defining Const from a definition LHS.
func definedSymbolConst(node Expr) *Const {
	if c, ok := node.(*Const); ok {
		return c
	}
	if app, ok := node.(*Apply); ok {
		if c, ok := app.Func.(*Const); ok {
			return c
		}
	}
	return nil
}

// definedSymbolName returns just the name of the defined symbol (no sort info).
// Used where only the name is needed (e.g., allNames checks, definition filtering).
func definedSymbolName(node Expr) string {
	if c := definedSymbolConst(node); c != nil {
		return c.Name
	}
	return ""
}

func usedSymbolExprs(node Expr) *InsMap[NodeKey, Expr] {
	syms := NewInsMap[NodeKey, Expr]()
	collectSymbolsInto("isolate.usedSymbolExprs", node, syms)
	return syms
}

// usedSymbolNames returns a list of plain symbol names (no sort info)
// from the given expression. Used by FindReferences and other name-based lookups.
func isolateUsedSymbolNames(node Expr) []string {
	exprs := usedSymbolExprs(node)
	result := make([]string, 0, exprs.Len())
	seen := make(map[string]bool)
	for _, v := range exprs.All() {
		if c, ok := v.(*Const); ok {
			if !seen[c.Name] {
				seen[c.Name] = true
				result = append(result, c.Name)
			}
		}
	}
	return result
}

// collectSymbolsInto walks node with il.SymbolsIluAst and adds all
// yielded symbols into the target InsMap. This matches Python's
// lu.symbols_ilu_ast behavior, including binder expansion, in AST traversal order.
// The label parameter is used for online per-symbol xtracer tracing:
// each NEW symbol addition emits an xtracer.Trace line immediately.
func collectSymbolsInto(label string, node Expr, syms *InsMap[NodeKey, Expr]) {
	if node == nil {
		return
	}
	for sym := range SymbolsIluAst(node) {
		key := Key(sym)
		if xtracer.Enabled {
			if _, exists := syms.Get2(key); !exists {
				xtracer.Trace("%s.add %s", label, PrettyFmla(sym))
			}
		}
		syms.Set(key, sym)
	}
}

// normalizeSymbolKeys applies normalize_symbol to the entries of a symbol set,
// returning a NEW insertion-ordered map. Matches Python's:
//
//	all_syms = OrderedSymSet()
//	for sym in all_syms_raw:
//	    nsym = ivy_logic.normalize_symbol(sym)
//	    if nsym not in all_syms: xtracer.trace("label.add_normalized nsym")
//	    all_syms.add(nsym)
//
// Polymorphic macros (<=, >, >=) are mapped to canonical form (<).
// Non-polymorphic symbols pass through unchanged.
// The label parameter is used for inline add_normalized traces.
func normalizeSymbolKeys(label string, syms *InsMap[NodeKey, Expr], usePolymorphicMacros bool, iuCfg *IvyUtilsConfig) *InsMap[NodeKey, Expr] {
	result := NewInsMap[NodeKey, Expr]()
	for _, expr := range syms.All() {
		var norm Expr = expr
		if c, ok := expr.(*Const); ok && usePolymorphicMacros {
			if _, isPolyMacro := PolymorphicMacrosMap[c.Name]; isPolyMacro {
				nc := NormalizeSymbol(c, iuCfg)
				// NormalizeSymbol may not work if iuCfg flag is wrong;
				// construct manually if needed
				if nc.Name == c.Name {
					canonical := PolymorphicMacrosMap[c.Name]
					nc = NewConst(canonical, c.CSort)
				}
				norm = nc
			}
		}
		var newKey NodeKey
		if c, ok := norm.(*Const); ok {
			newKey = ConstSymKey(c)
		} else {
			newKey = Key(norm)
		}
		if xtracer.Enabled {
			if _, exists := result.Get2(newKey); !exists {
				xtracer.Trace("%s.add_normalized %s", label, PrettyFmla(norm))
			}
		}
		result.Set(newKey, norm)
	}
	return result
}

func isolateCopyStringSet(s map[string]bool) map[string]bool {
	c := make(map[string]bool, len(s))
	for k, v := range s {
		c[k] = v
	}
	return c
}

// allSymsNameSet extracts a name-only set from a NodeKey→Expr symbol InsMap.
// Used where downstream code needs name-based lookup (e.g., CollectSortDestructors).
func allSymsNameSet(syms *InsMap[NodeKey, Expr]) map[string]bool {
	names := make(map[string]bool, syms.Len())
	for _, v := range syms.All() {
		if c, ok := v.(*Const); ok {
			names[c.Name] = true
		}
	}
	return names
}

// symSetContainsName checks if any entry in a symbol set has the given name.
func symSetContainsName(syms *InsMap[NodeKey, Expr], name string) bool {
	for _, v := range syms.All() {
		if c, ok := v.(*Const); ok && c.Name == name {
			return true
		}
	}
	return false
}

func copySymSet(s *InsMap[NodeKey, Expr]) *InsMap[NodeKey, Expr] {
	c := NewInsMap[NodeKey, Expr]()
	for k, v := range s.All() {
		c.Set(k, v)
	}
	return c
}

// -----------------------------------------------------------------------
// get_prop_dependencies: get property-to-object dependencies
// -----------------------------------------------------------------------

// GetPropDependencies returns a list of (property, dependency names) pairs.
// Each property's proof depends on the objects listed.
// Corresponds to Python get_prop_dependencies (lines 659-683).
func GetPropDependencies(mod *Module) []PropDep {
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
			for _, anc := range Ancestors(name, mod.Cfg.IuCfg.ComposeCharacter) {
				objs[anc] = true
			}
		}
	}
	for _, itps := range mod.Interps {
		for _, itp := range itps {
			if lf, ok := itp.(*LabeledFormula); ok {
				name := lfLabelName(lf)
				if name != "" {
					for _, anc := range Ancestors(name, mod.Cfg.IuCfg.ComposeCharacter) {
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
		for _, anc := range specAncestors(name, mod.Cfg.IuCfg.ComposeCharacter) {
			// Match Python defaultdict auto-vivification
			if _, ok := depmap[anc]; !ok {
				depmap[anc] = nil
			}
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
	Prop *LabeledFormula
	Deps []string
}

// specAncestors returns ancestors, skipping "spec" children.
// Corresponds to Python spec_ancestors().
func specAncestors(name string, cc string) []string {
	var result []string
	s := name
	for {
		result = append(result, s)
		idx := strings.LastIndex(s, cc)
		if idx < 0 {
			break
		}
		child := s[idx+len(cc):]
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
func GetPropsProvedInIsolate(mod *Module, iso interface{}) (proved, notProved []*LabeledFormula) {
	if versionLE(mod.Cfg.IsolateCfg.IvyVersion, "1.6") {
		return GetPropsProvedInIsolateOrig(mod, iso)
	}

	// Save and temporarily modify privates
	savePrivates := mod.Privates
	mod.Privates = make(map[string]bool)
	SetPrivatesFull(mod, iso, "impl")
	verified, _ := GetIsolateInfoFull(mod, iso, "impl", nil)

	// For version > 1.6: mark verified names from other isolates as private
	idef, _ := iso.(IsolateDefIface)
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

	// Remove subgoals from not_proved
	subs := make(map[int64]bool)
	for _, sg := range mod.Subgoals {
		for _, sub := range sg.Subgoals {
			subs[sub.ID] = true
		}
	}
	var filteredNotProved []*LabeledFormula
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

// -----------------------------------------------------------------------
// add_extern_precond
// -----------------------------------------------------------------------

// AddExternPrecond adds preconditions from call arguments to the
// preconds list. Corresponds to Python add_extern_precond (lines 876-884).
func AddExternPrecond(mod *Module, callee ActionsAction, callArgs []Expr, preconds *[]Expr) {
	calleeAct, ok := callee.(ActionsAction)
	if !ok {
		return
	}
	formalParams := calleeAct.GetFormalParams()
	var conjs []Expr
	for i, fml := range formalParams {
		if i >= len(callArgs) {
			break
		}
		act := callArgs[i]
		if isNumeralOrConstructor(act, mod) {
			conjs = append(conjs, &Eq{T1: fml, T2: act})
		}
	}
	if len(conjs) == 0 {
		// Anything or true is true: clear preconds
		// Python: del preconds[:]
		*preconds = (*preconds)[:0]
	}
	// Python line 936 runs unconditionally: preconds.append(And(*conjs))
	*preconds = append(*preconds, isolateMakeAnd(conjs...))
}

func isNumeralOrConstructor(node Expr, mod *Module) bool {
	if c, ok := node.(*Const); ok {
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

// getCone recursively adds action_name and its transitive callees to cone.
// Matches Python's get_cone (ivy_isolate.py:1462-1475).
func GetCone(actionsMap *InsMap[string, ActionsAction], actionName string, cone map[string]bool) {
	if cone[actionName] {
		return
	}
	cone[actionName] = true
	act, ok := actionsMap.Get2(actionName)
	if !ok {
		return
	}
	for _, sub := range act.IterSubactions() {
		switch a := sub.(type) {
		case *LogicCallAction:
			GetCone(actionsMap, a.CalleeName(), cone)
		case *LogicNativeAction:
			// Python: for arg in a.args[1:]: if isinstance(arg,ivy_ast.Atom) and a.rep in actions
			// In Go, native params are lg.Expr — Atoms become *lg.Const after compilation
			for _, arg := range a.Params {
				if c, ok := arg.(*Const); ok {
					if _, exists := actionsMap.Get2(c.Name); exists {
						GetCone(actionsMap, c.Name, cone)
					}
				}
			}
		}
	}
}

func nativeConeActionName(node Node) string {
	switch n := node.(type) {
	case *Atom:
		return n.Rep
	case *nativeAtomExpr:
		return n.Rep
	case *Const:
		return n.Name
	case *CompiledNode:
		switch e := n.Node.(type) {
		case *nativeAtomExpr:
			return e.Rep
		case *Atom:
			return e.Rep
		case *Const:
			return e.Name
		case *Apply:
			if c, ok := e.Func.(*Const); ok {
				return c.Name
			}
		case interface{ Relname() string }:
			return e.Relname()
		}
	case interface{ Relname() string }:
		return n.Relname()
	}
	return ""
}

// GetModConeFull returns the cone of action names reachable from roots.
// Actions referenced by natives and initializers are also included.
// Matches Python get_mod_cone (ivy_isolate.py:1482-1494).
func GetModConeFull(mod *Module, actionsMap *InsMap[string, ActionsAction],
	roots *InsMap[string, bool], afterInits []string) map[string]bool {

	cone := make(map[string]bool)

	// Start with roots — Python: for a in roots: get_cone(actions, a, cone)
	for name := range roots.All() {
		GetCone(actionsMap, name, cone)
	}

	// Add actions referenced by natives
	// Python: for n in mod.natives: for a in n.args[2:]: if isinstance(a,ivy_ast.Atom) and a.rep in mod.actions
	// Natives are *ast.NativeDef with Args() = [name, code_template, ref1, ref2, ...].
	// After compilation, refs are *ast.CompiledNode wrapping lg.Expr.
	for _, nat := range mod.Natives {
		args := nat.Args()
		for i := 2; i < len(args); i++ {
			name := nativeConeActionName(args[i])
			if name == "" {
				continue
			}
			if _, exists := mod.Actions.Get2(name); exists {
				GetCone(actionsMap, name, cone)
			}
		}
	}

	// Add after-init actions — Python: for ai in after_inits: get_cone(actions, ai, cone)
	for _, ai := range afterInits {
		GetCone(actionsMap, ai, cone)
	}

	return cone
}

// -----------------------------------------------------------------------
// has_side_effect: Python-compatible version
// -----------------------------------------------------------------------

// HasSideEffectFull checks if an action has side effects on the module
// signature. Follows through calls transitively.
// Corresponds to Python has_side_effect (lines 458-479).
func HasSideEffectFull(mod *Module, newActions *InsMap[string, ActionsAction], actname string) bool {
	return HasSideEffect(mod, actname, newActions)
}

// -----------------------------------------------------------------------
// Utility: IsolateDefNode interface for the actual isolate AST node
// -----------------------------------------------------------------------

// IsolateDefNode extends IsolateDefIface with additional methods
// needed by isolate_component.
type IsolateDefNode interface {
	IsolateDefIface
	// Params returns the isolate parameters.
	Params() []*Const
	// WithArgs returns the number of with-clause arguments.
	WithArgs() int
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
func GetStripBinding(node Expr, stripMap StripMap, stripBinding map[NodeKey]string, mod *Module) error {
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
	var args []Expr
	switch n := node.(type) {
	case *Apply:
		if c, ok := n.Func.(*Const); ok {
			name = c.Name
		}
		args = n.Terms
	case *Const:
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
		if existing, ok := stripBinding[Key(ap)]; ok && existing != sp {
			return fmt.Errorf("cannot strip parameter %v from %s", ap, name)
		}
		stripBinding[Key(ap)] = sp
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
func HasUnsummarizedMixins(mod *Module, actname string, summarizedActions map[string]bool, kind MixinKind) bool {
	// Python: mod.mixins[actname] — auto-vivifies
	mixins := mixinsAutoVivify(mod, actname)
	if len(mixins) == 0 {
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
	mod *Module,
	newActions *InsMap[string, ActionsAction],
	summarizedActions map[string]bool,
	callouts map[string]Callouts,
	action ActionsAction,
	acallouts *Callouts,
	head, tail bool,
) {
	switch a := action.(type) {
	case *LogicSequence:
		for idx, child := range a.Elems {
			subAct, _ := child.(ActionsAction)
			if subAct == nil {
				if act, ok := child.(ActionsAction); ok {
					subAct = act
				}
			}
			if subAct != nil {
				GetCalloutsAction(mod, newActions, summarizedActions, callouts, subAct, acallouts,
					head && idx == 0, tail && idx == len(a.Elems)-1)
			}
		}
	case *LogicCallAction:
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
			if subAct, ok := arg.(ActionsAction); ok {
				GetCalloutsAction(mod, newActions, summarizedActions, callouts, subAct, acallouts, head, tail)
			} else if w, ok := arg.(ActionsAction); ok {
				GetCalloutsAction(mod, newActions, summarizedActions, callouts, w, acallouts, head, tail)
			}
		}
	}
}

// GetCallouts computes callout information for a named action.
// Corresponds to Python get_callouts (lines 551-557).
func GetCallouts(
	mod *Module,
	newActions *InsMap[string, ActionsAction],
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
	xtracer.Trace("isolate.GetCallouts actname=%s", actname)
	acallouts := NewCallouts()
	callouts[actname] = acallouts
	action, ok := newActions.Get2(actname)
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
func GetLocMods(mod *Module, actname string) []string {
	xtracer.Trace("isolate.GetLocMods ENTER actname=%s", actname)
	act, ok := mod.Actions.Get2(actname)
	if !ok {
		return nil
	}
	// Python: action.modifies() — non-recursive, returns [] for Sequence.
	modSet := ModifiesSingle(act, &ActionsConfig{Context: NewActionContext(mod)})
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
func FindReferences(mod *Module, syms map[string]bool, newActions *InsMap[string, ActionsAction]) map[int]bool {
	refs := make(map[int]bool)

	// Check labeled formulas
	allFormulas := make([]*LabeledFormula, 0)
	allFormulas = append(allFormulas, mod.LabeledAxioms...)
	allFormulas = append(allFormulas, mod.LabeledProps...)
	allFormulas = append(allFormulas, mod.LabeledInits...)
	allFormulas = append(allFormulas, mod.LabeledConjs...)
	allFormulas = append(allFormulas, mod.Definitions...)

	for _, lf := range allFormulas {
		if lf.Formula == nil {
			continue
		}
		fmla, ok := lf.Formula.(Expr)
		if !ok {
			continue
		}
		fSyms := isolateUsedSymbolNames(fmla)
		for _, s := range fSyms {
			if syms[s] {
				refs[lf.Lineno()] = true
				break
			}
		}
	}

	// Check actions
	for _, act := range newActions.All() {
		actSyms := collectActionSymNames(act)
		for s := range actSyms {
			if syms[s] {
				loc := act.GetLineno()
				refs[loc.Line] = true
				break
			}
		}
	}

	xtracer.Trace("isolate.FindReferences n_syms=%d n_refs=%d", len(syms), len(refs))
	return refs
}

// collectActionSymNames collects all constant symbol names referenced by an action.
// Returns a name-only set for use in FindReferences.
func collectActionSymNames(act ActionsAction) map[string]bool {
	exprs := NewInsMap[NodeKey, Expr]()
	for _, arg := range actionSymbolArgs(act) {
		collectSymbolsInto("isolate.collectActionSymNames", arg, exprs)
	}
	// Also recurse into sub-actions
	for _, sub := range act.IterSubactions() {
		if sub == act {
			continue // skip self to avoid infinite loop
		}
		for _, arg := range actionSymbolArgs(sub) {
			collectSymbolsInto("isolate.collectActionSymNames", arg, exprs)
		}
	}
	names := make(map[string]bool, exprs.Len())
	for _, v := range exprs.All() {
		if c, ok := v.(*Const); ok {
			names[c.Name] = true
		}
	}
	return names
}

func actionSymbolArgs(act ActionsAction) []Expr {
	if call, ok := act.(*LogicCallAction); ok {
		return call.ActualReturns
	}
	return act.ActionArgs()
}

// -----------------------------------------------------------------------
// hide_action_params: wrap action with LocalAction hiding formals
// -----------------------------------------------------------------------

// HideActionParams wraps an action in a LocalAction that hides its
// formal parameters and returns.
// Corresponds to Python hide_action_params (lines 1438-1441).
func HideActionParams(action ActionsAction, mod *Module) ActionsAction {
	params := action.GetFormalParams()
	returns := action.GetFormalReturns()

	// Build locals list: params + returns
	var locals []Expr
	for _, p := range params {
		locals = append(locals, p)
	}
	for _, r := range returns {
		locals = append(locals, r)
	}

	// Create LocalAction with locals + body (action wrapped as node)
	args := make([]Expr, 0, len(locals)+1)
	args = append(args, locals...)
	args = append(args, action)
	actCfg := mod.Cfg.ActCfg
	return NewLocalActionOn(actCfg, "isolate.hide_action_params", args...)
}
