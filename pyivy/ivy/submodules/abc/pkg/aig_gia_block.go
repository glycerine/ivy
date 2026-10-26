package abc

// Interleaved block extraction from src/aig/gia/giaDup.c:Gia_ManDupBlock.

// ExtractInterleavedBlock keeps source CI/CO indices 0,stride,2*stride,...,
// substituting zero for other CIs. Both terminal counts must divide evenly.
// Target register count is floor(source registers/stride); constraint roles
// are discarded as in C. This is a structural selection/substitution, not a
// whole-machine equivalence claim. Name/spec survive; extended gates are
// lowered through hashing and unreachable logic is cleaned up.
func (m *GiaManager) ExtractInterleavedBlock(stride int, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || stride < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	if len(m.cis)%stride != 0 || len(m.cos)%stride != 0 {
		return nil, ErrDimension
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
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
	for i, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id] = GiaConst0
		if i%stride == 0 {
			mapping[id], err = g.AppendCI()
			if err != nil {
				return nil, err
			}
		}
	}
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[i], err = hashGiaCopiedGate(m, g, GiaID(i), mapping)
		if err != nil {
			return nil, err
		}
	}
	for i, id := range m.cos {
		if i%stride != 0 {
			continue
		}
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
	}
	g.HashStop()
	g.registers = m.registers / stride
	if stopped() {
		return nil, ErrCancelled
	}
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}
