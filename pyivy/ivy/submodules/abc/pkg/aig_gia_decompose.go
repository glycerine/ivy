package abc

import "sort"

// Output decomposition from src/aig/gia/giaDup.c:Gia_ManDupDemiter.

// GiaOutputDecomposition contains source literals for the new property outputs.
// OR says their disjunction reconstructs the source PO; otherwise use AND.
// Original RIs follow those outputs, with register/reset order unchanged.
type GiaOutputDecomposition struct {
	Graph *GiaManager
	Parts []GiaLit
	OR    bool
}

// DecomposeOutput splits one PO's ordinary-AND tree at complemented or nongate
// boundaries. A complemented root applies De Morgan to every boundary. Parts
// sort by ascending source level, preserving DFS encounter order on ties.
// At least two distinct phased boundaries are required. Extended nodes are
// opaque boundaries; all copied logic lowers to hashed ANDs. All CIs and RIs
// remain, name/spec survive, and constraint roles clear as in C. Unsupported
// root forms return ErrUnsupported instead of invoking C's unsafe traversal.
func (m *GiaManager) DecomposeOutput(options GiaCofactorOptions) (GiaOutputDecomposition, error) {
	fail := func(e error) (GiaOutputDecomposition, error) { return GiaOutputDecomposition{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if m.PrimaryOutputCount() != 1 {
		return fail(ErrDimension)
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	root, e := m.objects[m.cos[0]].Fanin0(m.cos[0])
	if e != nil {
		return fail(e)
	}
	ordinary := func(l GiaLit) bool {
		o := m.objects[l.ID()]
		_, mux := m.MuxControl(l.ID())
		return !l.Complemented() && o.IsAnd() && !o.IsBuffer() && !o.IsXor() && !mux
	}
	regular := root.NotIf(root.Complemented())
	if !ordinary(regular) {
		return fail(ErrUnsupported)
	}
	stack := []GiaLit{regular}
	expanded := make([]bool, len(m.objects))
	seen := make(map[GiaLit]bool)
	leaves := []GiaLit{}
	for len(stack) > 0 {
		if stopped() {
			return fail(ErrCancelled)
		}
		l := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !ordinary(l) {
			if !seen[l] {
				seen[l] = true
				leaves = append(leaves, l)
			}
			continue
		}
		if expanded[l.ID()] {
			continue
		}
		expanded[l.ID()] = true
		a, err := m.objects[l.ID()].Fanin0(l.ID())
		if err != nil {
			return fail(err)
		}
		b, err := m.objects[l.ID()].Fanin1(l.ID())
		if err != nil {
			return fail(err)
		}
		stack = append(stack, b, a)
	}
	if len(leaves) < 2 {
		return fail(ErrUnsupported)
	}
	levels, _, e := m.Levels(limit)
	if e != nil {
		return fail(e)
	}
	sort.SliceStable(leaves, func(i, j int) bool { return levels[leaves[i].ID()] < levels[leaves[j].ID()] })
	capObjects := options.MaxObjects
	if capObjects == 0 {
		capObjects = 1 << 20
	}
	g, e := NewGiaManagerWithLimit(capObjects)
	if e != nil {
		return fail(e)
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	if e = g.HashStart(); e != nil {
		return fail(e)
	}
	for i, o := range m.objects {
		if stopped() {
			return fail(ErrCancelled)
		}
		if o.IsCI() {
			mapping[i], e = g.AppendCI()
		} else if o.IsAnd() {
			mapping[i], e = hashGiaCopiedGate(m, g, GiaID(i), mapping)
		}
		if e != nil {
			return fail(e)
		}
	}
	parts := make([]GiaLit, len(leaves))
	for i, l := range leaves {
		parts[i] = l.NotIf(root.Complemented())
	}
	for _, l := range parts {
		if stopped() {
			return fail(ErrCancelled)
		}
		v, err := mappedGiaSignal(mapping, l)
		if err != nil {
			return fail(err)
		}
		if _, e = g.AppendCO(v); e != nil {
			return fail(e)
		}
	}
	for _, id := range m.cos[1:] {
		if stopped() {
			return fail(ErrCancelled)
		}
		d, err := m.objects[id].Fanin0(id)
		if err != nil {
			return fail(err)
		}
		v, err := mappedGiaSignal(mapping, d)
		if err != nil {
			return fail(err)
		}
		if _, e = g.AppendCO(v); e != nil {
			return fail(e)
		}
	}
	g.registers = m.registers
	g.HashStop()
	if stopped() {
		return fail(ErrCancelled)
	}
	cleaned, _, e := g.Cleanup()
	if e != nil {
		return fail(e)
	}
	return GiaOutputDecomposition{cleaned, parts, root.Complemented()}, nil
}
