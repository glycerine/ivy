package module

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/xtracer"
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

func canonActionMap(actions *iu.InsMap[string, Action]) string {
	if actions == nil || actions.Len() == 0 {
		return "(insMap)"
	}
	var b strings.Builder
	b.WriteString("(insMap")
	// Actions map: dump keys in insertion order
	for name, action := range actions.All() {
		fmt.Fprintf(&b, " %v:%s", name, action.Sexp())
	}
	b.WriteString(")")
	return b.String()
}

// CanonSnapshot emits a canonical s-expression snapshot of key module state
// via xtracer. The label identifies the compilation pass (e.g. "after-domain-setup").
// Both Go and Python emit identical snapshots for identical compiled state.
func (m *Module) CanonSnapshot(label string) {
	xtracer.Trace("module.CanonSnapshot ENTER label=%s", label)

	// Labeled formula slices
	xtracer.Trace("module.CanonSnapshot %s labeledAxioms=%s", label, canonLFSlice(m.LabeledAxioms))
	xtracer.Trace("module.CanonSnapshot %s definitions=%s", label, canonLFSlice(m.Definitions))
	xtracer.Trace("module.CanonSnapshot %s labeledProps=%s", label, canonLFSlice(m.LabeledProps))
	xtracer.Trace("module.CanonSnapshot %s labeledInits=%s", label, canonLFSlice(m.LabeledInits))
	xtracer.Trace("module.CanonSnapshot %s labeledConjs=%s", label, canonLFSlice(m.LabeledConjs))

	// Signature sorts (sorted by name)
	if m.Sig != nil {
		xtracer.Trace("module.CanonSnapshot %s sig.sorts=%s", label, canonSortMap(m.Sig.Sorts))
		xtracer.Trace("module.CanonSnapshot %s sig.interp=%s", label, canonInterpMap(m.Sig.Interp))
	}

	// Schemata (sorted by name)
	xtracer.Trace("module.CanonSnapshot %s schemata=%s", label, canonSchemaMap(m.Schemata))

	// InitCond and Theory
	if m.InitCond != nil {
		xtracer.Trace("module.CanonSnapshot %s initCond=%s", label, string(m.InitCond.Canon()))
	}
	if m.Theory != nil {
		xtracer.Trace("module.CanonSnapshot %s theory=%s", label, string(m.Theory.Canon()))
	}

	// Actions map: dump keys summary then per-action bodies
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

	xtracer.Trace("module.CanonSnapshot EXIT label=%s", label)
}

// canonLFSlice returns a canonical s-expression for a slice of LabeledFormulas.
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

// canonSortMap returns a sorted canonical representation of a sort map.
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

// canonInterpMap returns a sorted canonical representation of the interp map.
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

// canonSchemaMap returns a sorted canonical representation of the schemata map.
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
