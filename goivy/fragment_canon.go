package goivy

import (
	"fmt"
	"sort"
	"strings"
)

// --- helpers for UFNode serialization ---

func ufNodeSexp(n *UFNode) string {
	if n == nil {
		return "nil"
	}
	return fmt.Sprintf("(ufNode id:%d)", n.ID)
}

func ufNodeSetSexp(m map[*UFNode]bool) string {
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

func fragmentExprSexp(e Node) string {
	if e == nil {
		return "nil"
	}
	return string(e.Canon()) // lg.Key(e))
}

func sortSexp(s Sort) string {
	if s == nil {
		return "nil"
	}
	return string(SortKey(s))
}

// --- FragmentError ---

func (e *FragmentError) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf(`(fragmentError message:"%s")`, e.Message))
}

func (e *FragmentError) Canon() Canonical { return Canonical(e.Sexp()) }

// --- stratEntry ---

func (s *stratEntry) Sexp() NodeKey {
	symStr := "nil"
	if s.isSort && s.eqExpr != nil {
		// Python stores il.Symbol('=', expr) where expr is the expression itself
		// (not its sort). Match that: serialize as (Symbol name:= sort:<expr.Sexp()>).
		symStr = fmt.Sprintf("(Symbol name:= sort:%s)", Key(s.eqExpr))
	} else if s.sym != nil {
		symStr = string(s.sym.Sexp())
	}
	vStr := "nil"
	if s.v != nil {
		vStr = string(s.v.Sexp())
	}
	return NodeKey(fmt.Sprintf("(stratEntry sym:%s idx:%d v:%s isSort:%v)",
		symStr, s.idx, vStr, s.isSort))
}

func (s *stratEntry) Canon() Canonical { return Canonical(s.Sexp()) }

// --- arc ---

func (a *arc) Sexp() NodeKey {
	lineno := a.lineno
	// TODO remove?
	lineno = 0
	return NodeKey(fmt.Sprintf("(arc from:%s to:%s fmla:%s lineno:%d argIdx:%d hasIdx:%v)",
		ufNodeSexp(a.from), ufNodeSexp(a.to), fragmentExprSexp(a.fmla),
		lineno, a.argIdx, a.hasIdx))
}

func (a *arc) Canon() Canonical { return Canonical(a.Sexp()) }

// --- varID ---

func (v varID) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf(`(varID name:"%s" sort:"%s")`, v.name, v.sort))
}

func (v varID) Canon() Canonical { return Canonical(v.Sexp()) }

// --- macroDef ---

func (m *macroDef) Sexp() NodeKey {
	defStr := "nil"
	if m.def != nil {
		defStr = string(m.def.Sexp())
	}
	lfStr := "nil"
	if m.lf != nil {
		lfStr = string(m.lf.Canon())
	}
	return NodeKey(fmt.Sprintf("(macroDef def:%s lf:%s)", defStr, lfStr))
}

func (m *macroDef) Canon() Canonical { return Canonical(m.Sexp()) }

// --- mapFmlaRes ---

func (r *mapFmlaRes) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(mapFmlaRes node:%s uvs:%s)",
		ufNodeSexp(r.node), ufNodeSetSexp(r.uvs)))
}

func (r *mapFmlaRes) Canon() Canonical { return Canonical(r.Sexp()) }

// --- skolemEntry ---

func (s *skolemEntry) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(skolemEntry fmla:%s ast:%s)",
		fragmentExprSexp(s.fmla), fragmentExprSexp(s.ast)))
}

func (s *skolemEntry) Canon() Canonical { return Canonical(s.Sexp()) }

// --- fmlaPair ---

func (f *fmlaPair) Sexp() NodeKey {
	sourceStr := "nil"
	if f.source != nil {
		sourceStr = string(f.source.Canon())
	}
	lineno := f.lineno
	// TODO revert once golden working? we avoid spurious Sexp diffs with
	// this because the python line numbering is broken. Arguably we
	// should fix it, but we did not want to risk messing up the python
	// accidentally until we have very good confidence the Go is matching
	// it in all ways. So for now we just report lineno of 0 on both sides.
	lineno = 0
	return NodeKey(fmt.Sprintf("(fmlaPair fmla:%s source:%s lineno:%d)",
		fragmentExprSexp(f.fmla), sourceStr, lineno))
}

func (f *fmlaPair) Canon() Canonical { return Canonical(f.Sexp()) }

// --- checker ---

// sorted-map helper: map[varID]*lg.Variable
func varIDVarMapSexp(m map[varID]*LogicVariable) string {
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
func nodeKeyUFMapSexp(m map[NodeKey]*UFNode) string {
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
		parts[i] = fmt.Sprintf("%s:%s", k, ufNodeSexp(m[NodeKey(k)]))
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[lg.NodeKey]stratEntry
func nodeKeyStratEntryMapSexp(m map[NodeKey]stratEntry) string {
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
		e := m[NodeKey(k)]
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

// sorted-map helper: map[lg.NodeKey]mapFmlaRes
func nodeKeyMapFmlaResMapSexp(m map[NodeKey]mapFmlaRes) string {
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
		r := m[NodeKey(k)]
		parts[i] = fmt.Sprintf(`"%s":%s`, k, string(r.Sexp()))
	}
	return "(hash " + strings.Join(parts, " ") + ")"
}

// sorted-map helper: map[varID]*uf.UFNode
func varIDUFNodeMapSexp(m map[varID]*UFNode) string {
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
func varIDDepMapSexp(m map[varID]map[*UFNode]bool) string {
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

func (c *checker) Sexp() NodeKey {
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
	if len(c.interp) == 0 {
		b.WriteString("nil")
	} else {
		b.WriteString(stringInterfaceMapSexp(c.interp))
	}

	// universallyQuantifiedVars - can infinitely loop, so skip.
	//b.WriteString(" universallyQuantifiedVars:")
	//b.WriteString(varIDVarMapSexp(c.universallyQuantifiedVars))

	// universalVarLineno
	//b.WriteString(" universalVarLineno:")
	//b.WriteString(varIDIntMapSexp(c.universalVarLineno))

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

	// macroMap - can inf loop, skip.
	//b.WriteString(" macroMap:")
	//b.WriteString(stringMacroDefMapSexp(c.macroMap))

	// macroValueMap
	b.WriteString(" macroValueMap:")
	b.WriteString(nodeKeyMapFmlaResMapSexp(c.macroValueMap))

	// macroVarMap
	b.WriteString(" macroVarMap:")
	b.WriteString(varIDUFNodeMapSexp(c.macroVarMap))

	// macroDepMap
	b.WriteString(" macroDepMap:")
	b.WriteString(varIDDepMapSexp(c.macroDepMap))

	// skolemMap - can inf loop, skip.
	//b.WriteString(" skolemMap:")
	//b.WriteString(varIDSkolemMapSexp(c.skolemMap))

	// varUniq — omit internals, just indicate presence
	b.WriteString(" varUniq:")
	if c.varUniq != nil {
		b.WriteString("(variableUniqifier)")
	} else {
		b.WriteString("nil")
	}

	b.WriteString(")")
	return NodeKey(b.String())
}

func (c *checker) Canon() Canonical { return Canonical(c.Sexp()) }
