package fragment

import (
	"fmt"
	"sort"
	"strings"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	uf "github.com/glycerine/ivy/goivy/unionfind"
)

// --- helpers for UFNode serialization ---

func ufNodeSexp(n *uf.UFNode) string {
	if n == nil {
		return "nil"
	}
	return fmt.Sprintf("(ufNode id:%d)", n.ID)
}

func ufNodeSetSexp(m map[*uf.UFNode]bool) string {
	if m == nil {
		return "nil"
	}
	ids := make([]int, 0, len(m))
	for n := range m {
		ids = append(ids, int(n.ID))
	}
	sort.Ints(ids)
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func exprSexp(e lg.Expr) string {
	if e == nil {
		return "nil"
	}
	return string(lg.Key(e))
}

func sortSexp(s lg.Sort) string {
	if s == nil {
		return "nil"
	}
	return string(lg.SortKey(s))
}

// --- FragmentError ---

func (e *FragmentError) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf(`(fragmentError message:"%s")`, e.Message))
}

func (e *FragmentError) Canon() iu.Canonical { return iu.Canonical(e.Sexp()) }

// --- stratEntry ---

func (s *stratEntry) Sexp() lg.NodeKey {
	symStr := "nil"
	if s.sym != nil {
		symStr = string(s.sym.Sexp())
	}
	vStr := "nil"
	if s.v != nil {
		vStr = string(s.v.Sexp())
	}
	return lg.NodeKey(fmt.Sprintf("(stratEntry sym:%s idx:%d v:%s isSort:%v)",
		symStr, s.idx, vStr, s.isSort))
}

func (s *stratEntry) Canon() iu.Canonical { return iu.Canonical(s.Sexp()) }

// --- arc ---

func (a *arc) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(arc from:%s to:%s fmla:%s lineno:%d argIdx:%d hasIdx:%v)",
		ufNodeSexp(a.from), ufNodeSexp(a.to), exprSexp(a.fmla),
		a.lineno, a.argIdx, a.hasIdx))
}

func (a *arc) Canon() iu.Canonical { return iu.Canonical(a.Sexp()) }

// --- varID ---

func (v varID) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf(`(varID name:"%s" sort:"%s")`, v.name, v.sort))
}

func (v varID) Canon() iu.Canonical { return iu.Canonical(v.Sexp()) }

// --- macroDef ---

func (m *macroDef) Sexp() lg.NodeKey {
	defStr := "nil"
	if m.def != nil {
		defStr = string(m.def.Sexp())
	}
	lfStr := "nil"
	if m.lf != nil {
		lfStr = string(m.lf.Canon())
	}
	return lg.NodeKey(fmt.Sprintf("(macroDef def:%s lf:%s)", defStr, lfStr))
}

func (m *macroDef) Canon() iu.Canonical { return iu.Canonical(m.Sexp()) }

// --- mapFmlaRes ---

func (r *mapFmlaRes) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(mapFmlaRes node:%s uvs:%s)",
		ufNodeSexp(r.node), ufNodeSetSexp(r.uvs)))
}

func (r *mapFmlaRes) Canon() iu.Canonical { return iu.Canonical(r.Sexp()) }

// --- skolemEntry ---

func (s *skolemEntry) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(skolemEntry fmla:%s ast:%s)",
		exprSexp(s.fmla), exprSexp(s.ast)))
}

func (s *skolemEntry) Canon() iu.Canonical { return iu.Canonical(s.Sexp()) }

// --- fmlaPair ---

func (f *fmlaPair) Sexp() lg.NodeKey {
	sourceStr := "nil"
	if f.source != nil {
		if c, ok := f.source.(iu.Canonizer); ok {
			sourceStr = string(c.Canon())
		} else {
			sourceStr = fmt.Sprintf("%v", f.source)
		}
	}
	return lg.NodeKey(fmt.Sprintf("(fmlaPair fmla:%s source:%s lineno:%d)",
		exprSexp(f.fmla), sourceStr, f.lineno))
}

func (f *fmlaPair) Canon() iu.Canonical { return iu.Canonical(f.Sexp()) }

// --- checker ---

// sorted-map helper: map[varID]*lg.Variable
func varIDVarMapSexp(m map[varID]*lg.Variable) string {
	if m == nil {
		return "nil"
	}
	type kv struct {
		k string
		v string
	}
	pairs := make([]kv, 0, len(m))
	for vid, v := range m {
		pairs = append(pairs, kv{k: string(vid.Sexp()), v: string(v.Sexp())})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = p.k + ":" + p.v
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[varID]int
func varIDIntMapSexp(m map[varID]int) string {
	if m == nil {
		return "nil"
	}
	type kv struct {
		k string
		v int
	}
	pairs := make([]kv, 0, len(m))
	for vid, n := range m {
		pairs = append(pairs, kv{k: string(vid.Sexp()), v: n})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = fmt.Sprintf("%s:%d", p.k, p.v)
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[lg.NodeKey]*uf.UFNode
func nodeKeyUFMapSexp(m map[lg.NodeKey]*uf.UFNode) string {
	if m == nil {
		return "nil"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s:%s", k, ufNodeSexp(m[lg.NodeKey(k)]))
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[lg.NodeKey]stratEntry
func nodeKeyStratEntryMapSexp(m map[lg.NodeKey]stratEntry) string {
	if m == nil {
		return "nil"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		e := m[lg.NodeKey(k)]
		parts[i] = fmt.Sprintf("%s:%s", k, string(e.Sexp()))
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[string]macroDef
func stringMacroDefMapSexp(m map[string]macroDef) string {
	if m == nil {
		return "nil"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		md := m[k]
		parts[i] = fmt.Sprintf(`"%s":%s`, k, string(md.Sexp()))
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[string]mapFmlaRes
func stringMapFmlaResMapSexp(m map[string]mapFmlaRes) string {
	if m == nil {
		return "nil"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		r := m[k]
		parts[i] = fmt.Sprintf(`"%s":%s`, k, string(r.Sexp()))
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[varID]*uf.UFNode
func varIDUFNodeMapSexp(m map[varID]*uf.UFNode) string {
	if m == nil {
		return "nil"
	}
	type kv struct {
		k string
		v string
	}
	pairs := make([]kv, 0, len(m))
	for vid, n := range m {
		pairs = append(pairs, kv{k: string(vid.Sexp()), v: ufNodeSexp(n)})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = p.k + ":" + p.v
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[varID]map[*uf.UFNode]bool
func varIDDepMapSexp(m map[varID]map[*uf.UFNode]bool) string {
	if m == nil {
		return "nil"
	}
	type kv struct {
		k string
		v string
	}
	pairs := make([]kv, 0, len(m))
	for vid, deps := range m {
		pairs = append(pairs, kv{k: string(vid.Sexp()), v: ufNodeSetSexp(deps)})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = p.k + ":" + p.v
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[varID]skolemEntry
func varIDSkolemMapSexp(m map[varID]skolemEntry) string {
	if m == nil {
		return "nil"
	}
	type kv struct {
		k string
		v string
	}
	pairs := make([]kv, 0, len(m))
	for vid, se := range m {
		pairs = append(pairs, kv{k: string(vid.Sexp()), v: string(se.Sexp())})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = p.k + ":" + p.v
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[string]interface{}
func stringInterfaceMapSexp(m map[string]interface{}) string {
	if m == nil {
		return "nil"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf(`"%s":%v`, k, m[k])
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

func (c *checker) Sexp() lg.NodeKey {
	var b strings.Builder
	b.WriteString("(checker")

	// sig
	b.WriteString(" sig:")
	if c.sig != nil {
		b.WriteString(string(c.sig.Canon()))
	} else {
		b.WriteString("nil")
	}

	// interp
	b.WriteString(" interp:")
	b.WriteString(stringInterfaceMapSexp(c.interp))

	// universallyQuantifiedVars
	b.WriteString(" universallyQuantifiedVars:")
	b.WriteString(varIDVarMapSexp(c.universallyQuantifiedVars))

	// universalVarLineno
	b.WriteString(" universalVarLineno:")
	b.WriteString(varIDIntMapSexp(c.universalVarLineno))

	// stratMap
	b.WriteString(" stratMap:")
	b.WriteString(nodeKeyUFMapSexp(c.stratMap))

	// stratInfo
	b.WriteString(" stratInfo:")
	b.WriteString(nodeKeyStratEntryMapSexp(c.stratInfo))

	// arcs
	b.WriteString(" arcs:[")
	for i, a := range c.arcs {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(string(a.Sexp()))
	}
	b.WriteString("]")

	// macroMap
	b.WriteString(" macroMap:")
	b.WriteString(stringMacroDefMapSexp(c.macroMap))

	// macroValueMap
	b.WriteString(" macroValueMap:")
	b.WriteString(stringMapFmlaResMapSexp(c.macroValueMap))

	// macroVarMap
	b.WriteString(" macroVarMap:")
	b.WriteString(varIDUFNodeMapSexp(c.macroVarMap))

	// macroDepMap
	b.WriteString(" macroDepMap:")
	b.WriteString(varIDDepMapSexp(c.macroDepMap))

	// skolemMap
	b.WriteString(" skolemMap:")
	b.WriteString(varIDSkolemMapSexp(c.skolemMap))

	// varUniq — omit internals, just indicate presence
	b.WriteString(" varUniq:")
	if c.varUniq != nil {
		b.WriteString("(variableUniqifier)")
	} else {
		b.WriteString("nil")
	}

	b.WriteString(")")
	return lg.NodeKey(b.String())
}

func (c *checker) Canon() iu.Canonical { return iu.Canonical(c.Sexp()) }
