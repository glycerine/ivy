package module

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// Canon returns a single canonical s-expression of key module state.
// Used by SigCheck to Merkle-chain module state alongside Sig.
// Both Go and Python must produce identical output for identical state.
func (m *Module) Canon() iu.Canonical {
	var parts []string
	parts = append(parts, "axioms:"+canonLFSlice(m.LabeledAxioms))
	parts = append(parts, "defs:"+canonLFSlice(m.Definitions))
	parts = append(parts, "props:"+canonLFSlice(m.LabeledProps))
	parts = append(parts, "inits:"+canonLFSlice(m.LabeledInits))
	parts = append(parts, "conjs:"+canonLFSlice(m.LabeledConjs))
	if m.Sig != nil {
		parts = append(parts, "sig.sorts:"+canonSortMap(m.Sig.Sorts))
		parts = append(parts, "sig.interp:"+canonInterpMap(m.Sig.Interp))
	}
	parts = append(parts, "schemata:"+canonSchemaMap(m.Schemata))
	if m.InitCond != nil {
		parts = append(parts, "initCond:"+string(m.InitCond.Canon()))
	}
	if m.Theory != nil {
		parts = append(parts, "theory:"+string(m.Theory.Canon()))
	}
	parts = append(parts, "actions:"+canonActionMap(m.Actions))

	return iu.Canonical(fmt.Sprintf("(module %s)", strings.Join(parts, " ")))
}

// CanonSnapshot emits a canonical s-expression snapshot of ALL module state
// via xtracer. The label identifies the compilation pass (e.g. "after-domain-setup").
// Both Go and Python emit identical snapshots for identical compiled state.
func (m *Module) CanonSnapshot(label string) {
	xtracer.Trace("module.CanonSnapshot ENTER label=%s", label)

	// Group 1: Declarations (labeled formula slices)
	xtracer.Trace("module.CanonSnapshot %s labeledAxioms=%s", label, canonLFSlice(m.LabeledAxioms))
	xtracer.Trace("module.CanonSnapshot %s definitions=%s", label, canonLFSlice(m.Definitions))
	xtracer.Trace("module.CanonSnapshot %s labeledProps=%s", label, canonLFSlice(m.LabeledProps))
	xtracer.Trace("module.CanonSnapshot %s labeledInits=%s", label, canonLFSlice(m.LabeledInits))
	xtracer.Trace("module.CanonSnapshot %s labeledConjs=%s", label, canonLFSlice(m.LabeledConjs))
	xtracer.Trace("module.CanonSnapshot %s assertions=%s", label, canonLFSlice(m.Assertions))
	xtracer.Trace("module.CanonSnapshot %s assumedInvs=%s", label, canonLFSlice(m.AssumedInvs))
	xtracer.Trace("module.CanonSnapshot %s nativeDefinitions=%s", label, canonLFSlice(m.NativeDefinitions))
	if m.ConjSubgoals != nil {
		xtracer.Trace("module.CanonSnapshot %s conjSubgoals=%s", label, canonLFSlice(m.ConjSubgoals))
	}

	// Group 2: All relations
	xtracer.Trace("module.CanonSnapshot %s allRelations=%s", label, canonExprSlice(m.AllRelations))

	// Group 3: Signature
	if m.Sig != nil {
		xtracer.Trace("module.CanonSnapshot %s sig.sorts=%s", label, canonSortMap(m.Sig.Sorts))
		xtracer.Trace("module.CanonSnapshot %s sig.interp=%s", label, canonInterpMap(m.Sig.Interp))
	}

	// Group 4: Relations and functions
	xtracer.Trace("module.CanonSnapshot %s relations=%s", label, canonInsMapSort(m.Relations))
	xtracer.Trace("module.CanonSnapshot %s functions=%s", label, canonInsMapSort(m.Functions))

	// Group 5: Schemata, theorems, predicates
	xtracer.Trace("module.CanonSnapshot %s schemata=%s", label, canonSchemaMap(m.Schemata))
	xtracer.Trace("module.CanonSnapshot %s theorems=%s", label, canonSchemaMap(m.Theorems))
	xtracer.Trace("module.CanonSnapshot %s predicates=%s", label, canonSchemaMap(m.Predicates))

	// Group 6: InitCond and Theory
	if m.InitCond != nil {
		xtracer.Trace("module.CanonSnapshot %s initCond=%s", label, string(m.InitCond.Canon()))
	}
	if m.Theory != nil {
		xtracer.Trace("module.CanonSnapshot %s theory=%s", label, string(m.Theory.Canon()))
	}

	// Group 7: Actions
	if m.Actions != nil && m.Actions.Len() > 0 {
		var actKeys []string
		for name := range m.Actions.All() {
			actKeys = append(actKeys, name)
		}
		xtracer.Trace("module.CanonSnapshot %s actions.keys=%d keys=%s", label, len(actKeys), strings.Join(actKeys, ","))
		for name, action := range m.Actions.All() {
			xtracer.Trace("module.CanonSnapshot %s action[%s]=%s", label, name, action.Sexp())
		}
	}
	xtracer.Trace("module.CanonSnapshot %s beforeExport=%s", label, canonActionMap(m.BeforeExport))

	// Group 8: Mixins, PublicActions
	xtracer.Trace("module.CanonSnapshot %s mixins=%s", label, canonInsMapMixins(m.Mixins))
	xtracer.Trace("module.CanonSnapshot %s publicActions=%s", label, canonInsMapBool(m.PublicActions))

	// Group 9: Postconds
	xtracer.Trace("module.CanonSnapshot %s postconds=%s", label, canonPostcondsMap(m.Postconds))

	// Group 10: Initializers
	xtracer.Trace("module.CanonSnapshot %s initializers=%s", label, canonNamedActionSlice(m.Initializers))
	xtracer.Trace("module.CanonSnapshot %s initialActions=%s", label, canonActionSlice(m.InitialActions))

	// Group 11: Hierarchy
	xtracer.Trace("module.CanonSnapshot %s hierarchy=%s", label, canonInsMapHierarchy(m.Hierarchy))

	// Group 12: Updates, Instantiations
	xtracer.Trace("module.CanonSnapshot %s updates=%s", label, canonInterfaceSlice(m.Updates))
	xtracer.Trace("module.CanonSnapshot %s instantiations=%s", label, canonInstantiationSlice(m.Instantiations))

	// Group 13: Isolates
	xtracer.Trace("module.CanonSnapshot %s isolates=%s", label, canonIsolateMap(m.Isolates))
	if m.IsolateInfo != nil {
		xtracer.Trace("module.CanonSnapshot %s isolateInfo=%s", label, canonIsolateInfo(m.IsolateInfo))
	}
	xtracer.Trace("module.CanonSnapshot %s isolateProofs=%s", label, canonSchemaMap(m.IsolateProofs))
	if m.IsolateProof != nil {
		xtracer.Trace("module.CanonSnapshot %s isolateProof=%s", label, canonNode(m.IsolateProof))
	}

	// Group 14: Exports, imports, delegates
	xtracer.Trace("module.CanonSnapshot %s exports=%s", label, canonExporterSlice(m.Exports))
	xtracer.Trace("module.CanonSnapshot %s imports=%s", label, canonNodeSlice(m.Imports))
	xtracer.Trace("module.CanonSnapshot %s delegates=%s", label, canonDelegatorSlice(m.Delegates))

	// Group 15: Sorts and destructors
	xtracer.Trace("module.CanonSnapshot %s destructorSorts=%s", label, canonSortMap(m.DestructorSorts))
	xtracer.Trace("module.CanonSnapshot %s sortDestructors=%s", label, canonConstSliceMap(m.SortDestructors))
	xtracer.Trace("module.CanonSnapshot %s constructorSorts=%s", label, canonSortMap(m.ConstructorSorts))
	xtracer.Trace("module.CanonSnapshot %s sortConstructors=%s", label, canonConstSliceMap(m.SortConstructors))
	xtracer.Trace("module.CanonSnapshot %s ghostSorts=%s", label, canonBoolMap(m.GhostSorts))
	xtracer.Trace("module.CanonSnapshot %s sortOrder=%s", label, canonStringSlice(m.SortOrder))
	xtracer.Trace("module.CanonSnapshot %s symbolOrder=%s", label, canonConstSlice(m.SymbolOrder))
	xtracer.Trace("module.CanonSnapshot %s variants=%s", label, canonSortSliceMap(m.Variants))
	xtracer.Trace("module.CanonSnapshot %s supertypes=%s", label, canonSortSliceMap(m.Supertypes))
	xtracer.Trace("module.CanonSnapshot %s finiteSorts=%s", label, canonBoolMap(m.FiniteSorts))

	// Group 16: Interpretations and natives
	xtracer.Trace("module.CanonSnapshot %s interps=%s", label, canonNodeMapSlice(m.Interps))
	xtracer.Trace("module.CanonSnapshot %s natives=%s", label, canonNodeSlice(m.Natives))
	xtracer.Trace("module.CanonSnapshot %s nativeTypes=%s", label, canonNativeTypeMap(m.NativeTypes))

	// Group 17: Properties and proofs
	xtracer.Trace("module.CanonSnapshot %s progress=%s", label, canonInterfaceSlice(m.Progress))
	xtracer.Trace("module.CanonSnapshot %s rely=%s", label, canonExprSlice(m.Rely))
	xtracer.Trace("module.CanonSnapshot %s mixOrd=%s", label, canonNodeSlice(m.MixOrd))
	xtracer.Trace("module.CanonSnapshot %s privates=%s", label, canonBoolMap(m.Privates))
	xtracer.Trace("module.CanonSnapshot %s proofs=%s", label, canonProofEntrySlice(m.Proofs))
	xtracer.Trace("module.CanonSnapshot %s named=%s", label, canonNamedEntrySlice(m.Named))
	xtracer.Trace("module.CanonSnapshot %s subgoals=%s", label, canonSubgoalEntrySlice(m.Subgoals))
	xtracer.Trace("module.CanonSnapshot %s conjActions=%s", label, canonStringSliceMap(m.ConjActions))

	// Group 18: Parameters
	xtracer.Trace("module.CanonSnapshot %s params=%s", label, canonConstSlice(m.Params))
	xtracer.Trace("module.CanonSnapshot %s paramDefaults=%s", label, canonNodeSlice(m.ParamDefaults))

	// Group 19: Other
	xtracer.Trace("module.CanonSnapshot %s aliases=%s", label, canonStringMap(m.Aliases))
	xtracer.Trace("module.CanonSnapshot %s attributes=%s", label, canonInterpMap(m.Attributes))
	xtracer.Trace("module.CanonSnapshot %s extPreconds=%s", label, canonExprMap(m.ExtPreconds))
	xtracer.Trace("module.CanonSnapshot %s conceptSpaces=%s", label, canonConceptSpaceSlice(m.ConceptSpaces))
	xtracer.Trace("module.CanonSnapshot %s logics=%s", label, canonStringSlice(m.Logics))

	xtracer.Trace("module.CanonSnapshot EXIT label=%s", label)
}

// ---------------------------------------------------------------------------
// Existing helpers
// ---------------------------------------------------------------------------

func canonActionMap(actions *iu.InsMap[string, Action]) string {
	if actions == nil || actions.Len() == 0 {
		return "(insMap)"
	}
	var b strings.Builder
	b.WriteString("(insMap")
	for name, action := range actions.All() {
		fmt.Fprintf(&b, " %v:%s", name, action.Sexp())
	}
	b.WriteString(")")
	return b.String()
}

func canonLFSlice(lfs []*ast.LabeledFormula) string {
	if len(lfs) == 0 {
		return "[]"
	}
	var parts []string
	for _, lf := range lfs {
		parts = append(parts, string(lf.Canon()))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

func canonSortMap(sorts map[string]lg.Sort) string {
	if len(sorts) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(sorts))
	for k := range sorts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, string(sorts[k].Sexp())))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

func canonInterpMap(interp map[string]interface{}) string {
	if len(interp) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(interp))
	for k := range interp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		v := interp[k]
		var vs string
		switch val := v.(type) {
		case string:
			vs = fmt.Sprintf("%q", val)
		case fmt.Stringer:
			vs = val.String()
		case iu.Canonizer:
			vs = string(val.Canon())
		default:
			vs = fmt.Sprintf("%v", val)
		}
		parts = append(parts, fmt.Sprintf("%s:%s", k, vs))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

func canonSchemaMap(schemata map[string]ast.Node) string {
	if len(schemata) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(schemata))
	for k := range schemata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		v := schemata[k]
		var vs string
		if c, ok := v.(iu.Canonizer); ok {
			vs = string(c.Canon())
		} else {
			vs = fmt.Sprintf("%v", v)
		}
		parts = append(parts, fmt.Sprintf("%s:%s", k, vs))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// ---------------------------------------------------------------------------
// New generic helpers
// ---------------------------------------------------------------------------

// canonExprSlice returns canonical s-expression for a slice of lg.Expr.
func canonExprSlice(exprs []lg.Expr) string {
	if len(exprs) == 0 {
		return "[]"
	}
	var parts []string
	for _, e := range exprs {
		parts = append(parts, string(e.Sexp()))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonActionSlice returns canonical s-expression for a slice of Action.
func canonActionSlice(actions []Action) string {
	if len(actions) == 0 {
		return "[]"
	}
	var parts []string
	for _, a := range actions {
		parts = append(parts, string(a.Sexp()))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonBoolMap returns canonical representation of a map[string]bool as a sorted set.
// Python equivalents are sets; we emit sorted keys only.
func canonBoolMap(m map[string]bool) string {
	if len(m) == 0 {
		return "(set)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return fmt.Sprintf("(set %s)", strings.Join(keys, " "))
}

// canonStringSlice returns canonical form of a []string.
func canonStringSlice(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	var parts []string
	for _, s := range ss {
		parts = append(parts, fmt.Sprintf("%q", s))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonConstSlice returns canonical form of a []*lg.Const.
func canonConstSlice(cs []*lg.Const) string {
	if len(cs) == 0 {
		return "[]"
	}
	var parts []string
	for _, c := range cs {
		parts = append(parts, string(c.Sexp()))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonNodeSlice returns canonical form of a []ast.Node.
// nil elements produce "nil".
func canonNodeSlice(nodes []ast.Node) string {
	if len(nodes) == 0 {
		return "[]"
	}
	var parts []string
	for _, n := range nodes {
		parts = append(parts, canonNode(n))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonNode returns canonical form of a single ast.Node.
func canonNode(n ast.Node) string {
	if n == nil {
		return "nil"
	}
	if c, ok := n.(iu.Canonizer); ok {
		return string(c.Canon())
	}
	return fmt.Sprintf("%v", n)
}

// canonInterfaceSlice returns canonical form of a []interface{}.
func canonInterfaceSlice(items []interface{}) string {
	if len(items) == 0 {
		return "[]"
	}
	var parts []string
	for _, item := range items {
		if item == nil {
			parts = append(parts, "nil")
		} else if c, ok := item.(iu.Canonizer); ok {
			parts = append(parts, string(c.Canon()))
		} else if s, ok := item.(fmt.Stringer); ok {
			parts = append(parts, s.String())
		} else {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// ---------------------------------------------------------------------------
// New InsMap helpers
// ---------------------------------------------------------------------------

// canonInsMapSort returns canonical form of an InsMap[string, lg.Sort].
// Python stores arity integers in relations/functions dicts, so we emit
// the arity (number of domain parameters) to match Python's format.
func canonInsMapSort(m *iu.InsMap[string, lg.Sort]) string {
	if m == nil || m.Len() == 0 {
		return "(insMap)"
	}
	var b strings.Builder
	b.WriteString("(insMap")
	for name, s := range m.All() {
		arity := sortArity(s)
		fmt.Fprintf(&b, " %s:%d", name, arity)
	}
	b.WriteString(")")
	return b.String()
}

// sortArity returns the arity (number of domain parameters) of a sort.
// For FunctionSort, arity = len(sorts) - 1 (all args minus the result type).
// For non-function sorts, arity = 0.
func sortArity(s lg.Sort) int {
	if fs, ok := s.(*lg.FunctionSort); ok && len(fs.Sorts) > 0 {
		return len(fs.Sorts) - 1
	}
	return 0
}

// canonInsMapBool returns canonical form of an InsMap[string, bool].
// Only keys are emitted (insertion order).
func canonInsMapBool(m *iu.InsMap[string, bool]) string {
	if m == nil || m.Len() == 0 {
		return "(insMap)"
	}
	var b strings.Builder
	b.WriteString("(insMap")
	for name := range m.All() {
		fmt.Fprintf(&b, " %s", name)
	}
	b.WriteString(")")
	return b.String()
}

// canonInsMapMixins returns canonical form of an InsMap[string, []MixinDef].
func canonInsMapMixins(m *iu.InsMap[string, []MixinDef]) string {
	if m == nil || m.Len() == 0 {
		return "(insMap)"
	}
	var b strings.Builder
	b.WriteString("(insMap")
	for name, defs := range m.All() {
		var parts []string
		for _, d := range defs {
			if c, ok := d.(iu.Canonizer); ok {
				parts = append(parts, string(c.Canon()))
			} else {
				parts = append(parts, fmt.Sprintf("%v", d))
			}
		}
		fmt.Fprintf(&b, " %s:[%s]", name, strings.Join(parts, " "))
	}
	b.WriteString(")")
	return b.String()
}

// canonInsMapHierarchy returns canonical form of nested InsMap[string, InsMap[string,bool]].
func canonInsMapHierarchy(m *iu.InsMap[string, *iu.InsMap[string, bool]]) string {
	if m == nil || m.Len() == 0 {
		return "(insMap)"
	}
	var b strings.Builder
	b.WriteString("(insMap")
	for parent, children := range m.All() {
		fmt.Fprintf(&b, " %s:%s", parent, canonInsMapBool(children))
	}
	b.WriteString(")")
	return b.String()
}

// ---------------------------------------------------------------------------
// New sorted-map helpers
// ---------------------------------------------------------------------------

// canonPostcondsMap returns canonical form of map[string][]*ast.LabeledFormula.
func canonPostcondsMap(m map[string][]*ast.LabeledFormula) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, canonLFSlice(m[k])))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// canonSortSliceMap returns canonical form of map[string][]lg.Sort.
func canonSortSliceMap(m map[string][]lg.Sort) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		var ss []string
		for _, s := range m[k] {
			ss = append(ss, string(s.Sexp()))
		}
		parts = append(parts, fmt.Sprintf("%s:[%s]", k, strings.Join(ss, " ")))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// canonConstSliceMap returns canonical form of map[string][]*lg.Const.
func canonConstSliceMap(m map[string][]*lg.Const) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		var cs []string
		for _, c := range m[k] {
			cs = append(cs, string(c.Sexp()))
		}
		parts = append(parts, fmt.Sprintf("%s:[%s]", k, strings.Join(cs, " ")))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// canonStringMap returns canonical form of map[string]string.
func canonStringMap(m map[string]string) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%q", k, m[k]))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// canonStringSliceMap returns canonical form of map[string][]string.
func canonStringSliceMap(m map[string][]string) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		var ss []string
		for _, s := range m[k] {
			ss = append(ss, fmt.Sprintf("%q", s))
		}
		parts = append(parts, fmt.Sprintf("%s:[%s]", k, strings.Join(ss, " ")))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// canonNodeMapSlice returns canonical form of map[string][]ast.Node.
func canonNodeMapSlice(m map[string][]ast.Node) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, canonNodeSlice(m[k])))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// canonExprMap returns canonical form of map[string]lg.Expr.
func canonExprMap(m map[string]lg.Expr) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, string(m[k].Sexp())))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// canonIsolateMap returns canonical form of map[string]*ast.IsolateDef.
func canonIsolateMap(m map[string]*ast.IsolateDef) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, string(m[k].Canon())))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// canonNativeTypeMap returns canonical form of map[string]*ast.NativeType.
func canonNativeTypeMap(m map[string]*ast.NativeType) string {
	if len(m) == 0 {
		return "(hash)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, string(m[k].Canon())))
	}
	return fmt.Sprintf("(hash %s)", strings.Join(parts, " "))
}

// ---------------------------------------------------------------------------
// New struct-specific helpers
// ---------------------------------------------------------------------------

// canonNamedActionSlice returns canonical form of []NamedAction.
func canonNamedActionSlice(nas []NamedAction) string {
	if len(nas) == 0 {
		return "[]"
	}
	var parts []string
	for _, na := range nas {
		parts = append(parts, fmt.Sprintf("(%s:%s)", na.Name, string(na.Action.Sexp())))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonInstantiationSlice returns canonical form of []Instantiation.
func canonInstantiationSlice(insts []Instantiation) string {
	if len(insts) == 0 {
		return "[]"
	}
	var parts []string
	for _, inst := range insts {
		parts = append(parts, fmt.Sprintf("(schema:%s inst:%s)", canonNode(inst.Schema), canonNode(inst.Inst)))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonIsolateInfo returns canonical form of *IsolateInfo.
func canonIsolateInfo(info *IsolateInfo) string {
	if info == nil {
		return "nil"
	}
	return fmt.Sprintf("(isolateInfo impls:%s monitors:%s)",
		canonMixinTripleSlice(info.Implementations),
		canonMixinTripleSlice(info.Monitors))
}

// canonMixinTripleSlice returns canonical form of []MixinTriple.
func canonMixinTripleSlice(triples []MixinTriple) string {
	if len(triples) == 0 {
		return "[]"
	}
	var parts []string
	for _, t := range triples {
		parts = append(parts, fmt.Sprintf("(mixer:%s mixee:%s action:%s)",
			t.Mixer, t.Mixee, string(t.Action.Sexp())))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonExporterSlice returns canonical form of []Exporter.
func canonExporterSlice(exports []Exporter) string {
	if len(exports) == 0 {
		return "[]"
	}
	var parts []string
	for _, e := range exports {
		if c, ok := e.(iu.Canonizer); ok {
			parts = append(parts, string(c.Canon()))
		} else {
			parts = append(parts, fmt.Sprintf("(export %s %s)", e.Exported(), e.Scope()))
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonDelegatorSlice returns canonical form of []Delegator.
func canonDelegatorSlice(delegates []Delegator) string {
	if len(delegates) == 0 {
		return "[]"
	}
	var parts []string
	for _, d := range delegates {
		if c, ok := d.(iu.Canonizer); ok {
			parts = append(parts, string(c.Canon()))
		} else {
			parts = append(parts, fmt.Sprintf("(delegate %s %s)", d.Delegated(), d.Delegee()))
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonConceptSpaceSlice returns canonical form of []ConceptSpace.
func canonConceptSpaceSlice(css []ConceptSpace) string {
	if len(css) == 0 {
		return "[]"
	}
	var parts []string
	for _, cs := range css {
		parts = append(parts, fmt.Sprintf("(label:%s body:%s)",
			string(cs.Label.Sexp()), string(cs.Body.Sexp())))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonProofEntrySlice returns canonical form of []ProofEntry.
func canonProofEntrySlice(proofs []ProofEntry) string {
	if len(proofs) == 0 {
		return "[]"
	}
	var parts []string
	for _, pe := range proofs {
		parts = append(parts, fmt.Sprintf("(formula:%s proof:%s)",
			string(pe.Formula.Canon()), canonNode(pe.Proof)))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonNamedEntrySlice returns canonical form of []NamedEntry.
func canonNamedEntrySlice(named []NamedEntry) string {
	if len(named) == 0 {
		return "[]"
	}
	var parts []string
	for _, ne := range named {
		parts = append(parts, fmt.Sprintf("(formula:%s name:%s)",
			string(ne.Formula.Canon()), string(ne.Name.Sexp())))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// canonSubgoalEntrySlice returns canonical form of []SubgoalEntry.
func canonSubgoalEntrySlice(subs []SubgoalEntry) string {
	if len(subs) == 0 {
		return "[]"
	}
	var parts []string
	for _, se := range subs {
		parts = append(parts, fmt.Sprintf("(formula:%s subgoals:%s)",
			string(se.Formula.Canon()), canonLFSlice(se.Subgoals)))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}
