package abc

// Paired CO phases from src/aig/gia/giaDup.c:Gia_ManDupOnsetOffset.

// DuplicateAllOutputPhases emits each source CO driver followed by its
// complement, including RI drivers, in source CO order. It retains the source
// register count: only the final RegisterCount paired outputs become RIs.
// This exactly preserves C's layout, NOT the original transition functions.
// Source constraint roles are intentionally not inherited, also matching C.
// All raw gate kinds/dangling logic and CI/name/spec metadata are retained.
func (m *GiaManager) DuplicateAllOutputPhases(options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	if _, e := checkedSize(len(m.objects), 1, len(m.cos), options.MaxObjects); e != nil {
		return nil, e
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[i], err = appendGiaCopiedGate(m, g, GiaID(i), mapping)
		if err != nil {
			return nil, err
		}
	}
	for _, id := range m.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return nil, e
		}
		l, e := mappedGiaSignal(mapping, driver)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(l); err != nil {
			return nil, err
		}
		if _, err = g.AppendCO(l.Not()); err != nil {
			return nil, err
		}
	}
	g.registers = m.registers
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
