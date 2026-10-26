package abc

// Level-ordered duplication from src/aig/gia/giaDup.c:Gia_ManDupLevelized.

// DuplicateLevelized copies all objects, including dangling gates, with CIs
// first, gates in increasing forward level then source ID, and COs last.
// The owned map translates regular source objects to target literals (COs map
// to CO objects). Raw buffers/XORs/muxes and register/constraint roles survive.
// Unlike the C routine's ordinary-AND path, zero-delay buffers are supported;
// stable source-ID order keeps same-level buffer dependencies topological.
func (m *GiaManager) DuplicateLevelized(options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, nil, err
	}
	if len(m.objects) > options.MaxMappingEntries || len(m.objects) > options.MaxObjects {
		return nil, nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, nil, ErrCancelled
	}
	groups, err := m.Levelize(options.MaxMappingEntries)
	if err != nil {
		return nil, nil, err
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, id := range m.cis {
		if stopped() {
			return nil, nil, ErrCancelled
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return nil, nil, err
		}
	}
	for _, group := range groups {
		for _, id := range group {
			if stopped() {
				return nil, nil, ErrCancelled
			}
			mapping[id], err = appendGiaCopiedGate(m, g, id, mapping)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	for _, id := range m.cos {
		if stopped() {
			return nil, nil, ErrCancelled
		}
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return nil, nil, e
		}
		copied, e := mappedGiaSignal(mapping, driver)
		if e != nil {
			return nil, nil, e
		}
		mapping[id], err = g.AppendCO(copied)
		if err != nil {
			return nil, nil, err
		}
	}
	g.registers, g.constraints = m.registers, m.constraints
	if err = g.Check(); err != nil {
		return nil, nil, err
	}
	return g, mapping, nil
}
