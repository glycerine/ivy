package abc

// Contiguous CO extraction from src/aig/gia/giaDup.c:Gia_ManDupOutputGroup.

// ExtractOutputGroup copies CO ordinals [start, stop) as a combinational graph.
// Inputs appear at first DFS encounter, including source register outputs;
// returned Inputs maps target CI ordinals to source CI ordinals. Unused inputs
// and unreachable gates disappear. Raw extended kinds, name/spec and physical
// per-root gate/terminal interleaving survive. State/constraint roles clear.
// An empty range returns a constant-only graph. Source scratch stays unchanged.
func (m *GiaManager) ExtractOutputGroup(start, stop int, options GiaCofactorOptions) (GiaEncounterCone, error) {
	fail := func(e error) (GiaEncounterCone, error) { return GiaEncounterCone{}, e }
	if m == nil || start < 0 || stop < start || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if stop > len(m.cos) {
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
	inputs := []int{}
	for _, co := range m.cos[start:stop] {
		if stopped() {
			return fail(ErrCancelled)
		}
		nodes, err := m.CollectNodesAndInputs([]GiaID{co}, limit)
		if err != nil {
			return fail(err)
		}
		for _, id := range nodes {
			if mapping[id].Valid() {
				continue
			}
			if stopped() {
				return fail(ErrCancelled)
			}
			if m.objects[id].IsCI() {
				index, err := m.objects[id].TerminalIndex()
				if err != nil {
					return fail(err)
				}
				inputs = append(inputs, index)
				mapping[id], e = g.AppendCI()
			} else {
				mapping[id], e = appendGiaCopiedGate(m, g, id, mapping)
			}
			if e != nil {
				return fail(e)
			}
		}
		d, err := m.objects[co].Fanin0(co)
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
	if e = g.Check(); e != nil {
		return fail(e)
	}
	return GiaEncounterCone{g, inputs}, nil
}
