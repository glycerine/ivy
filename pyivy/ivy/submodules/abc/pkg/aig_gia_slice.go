package abc

// Support-limited slicing from src/aig/gia/giaDup.c:Gia_ManDupSliced.

// GiaSlicing exposes source regular gate IDs in target output order. The
// graph retains every source CI, including register outputs as ordinary inputs.
// It is a combinational collection of slices, not a replacement state machine.
type GiaSlicing struct {
	Graph *GiaManager
	Roots []GiaID
}

// SliceBySupport retains every gate with structural CI support <= maximum.
// Retained gates feeding an omitted gate or source CO become outputs, in source
// ID order. Direct CI/constant COs do not become outputs, matching C. All retained
// gates survive, even dangling ones. Native mux controls participate in support
// and boundary marking. Result constraints/register count are zero; name/spec
// are retained (C keeps only name). MaxMappingEntries also bounds support sets.
func (m *GiaManager) SliceBySupport(maximum int, options GiaCofactorOptions) (GiaSlicing, error) {
	fail := func(e error) (GiaSlicing, error) { return GiaSlicing{}, e }
	if m == nil || maximum < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if len(m.objects) > options.MaxMappingEntries {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	gates := make([]GiaID, 0, m.AndCount())
	for i, obj := range m.objects {
		if obj.IsAnd() {
			gates = append(gates, GiaID(i))
		}
	}
	supports, err := m.NodeSupports(gates, options.MaxMappingEntries, options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	retained, boundary := make([]bool, len(m.objects)), make([]bool, len(m.objects))
	for _, id := range m.cis {
		if stopped() {
			return fail(ErrCancelled)
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	for i, id := range gates {
		if stopped() {
			return fail(ErrCancelled)
		}
		if len(supports[i]) <= maximum {
			mapping[id], err = appendGiaCopiedGate(m, g, id, mapping)
			if err != nil {
				return fail(err)
			}
			retained[id] = true
			continue
		}
		obj := m.objects[id]
		first, e := obj.Fanin0(id)
		if e != nil {
			return fail(e)
		}
		boundary[first.ID()] = true
		if !obj.IsBuffer() {
			second, e := obj.Fanin1(id)
			if e != nil {
				return fail(e)
			}
			boundary[second.ID()] = true
		}
		if control, yes := m.MuxControl(id); yes {
			boundary[control.ID()] = true
		}
	}
	for _, id := range m.cos {
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return fail(e)
		}
		boundary[driver.ID()] = true
	}
	roots := make([]GiaID, 0)
	for _, id := range gates {
		if !retained[id] || !boundary[id] {
			continue
		}
		if stopped() {
			return fail(ErrCancelled)
		}
		if _, err = g.AppendCO(mapping[id]); err != nil {
			return fail(err)
		}
		roots = append(roots, id)
	}
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaSlicing{g, roots}, nil
}
