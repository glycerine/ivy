package abc

// Dual-output layouts from src/aig/gia/giaDup.c:Gia_ManTransformToDual
// and Gia_ManTransformTwoWord2DualOutput.

// ToDualOutputs replaces every PO f by adjacent (f,0), leaving RIs unchanged.
// Every source PO participates, and constraint roles are cleared as in C.
// CI/register order, name/spec and hashed dangling logic survive; extended
// gates are lowered. XORing adjacent pairs recovers source PO functions.
func (m *GiaManager) ToDualOutputs(options GiaCofactorOptions) (*GiaManager, error) {
	return m.dualOutputLayout(false, options)
}

// InterleaveOutputHalves turns [a0,a1,...,b0,b1,...] POs into [a0,b0,a1,b1,...].
// An even PO count is required. RIs are unchanged, constraint roles are cleared,
// and cleanup removes unused gates. Inputs/registers/name/spec retain order.
func (m *GiaManager) InterleaveOutputHalves(options GiaCofactorOptions) (*GiaManager, error) {
	return m.dualOutputLayout(true, options)
}

func (m *GiaManager) dualOutputLayout(interleave bool, options GiaCofactorOptions) (*GiaManager, error) {
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
	if interleave && m.PrimaryOutputCount()%2 != 0 {
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
	if err = g.HashStart(); err != nil {
		return nil, err
	}
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
		mapping[i], err = hashGiaCopiedGate(m, g, GiaID(i), mapping)
		if err != nil {
			return nil, err
		}
	}
	copyCO := func(index int) error {
		if stopped() {
			return ErrCancelled
		}
		id := m.cos[index]
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return e
		}
		l, e := mappedGiaSignal(mapping, driver)
		if e != nil {
			return e
		}
		_, e = g.AppendCO(l)
		return e
	}
	if interleave {
		half := m.PrimaryOutputCount() / 2
		for i := 0; i < half; i++ {
			if err = copyCO(i); err != nil {
				return nil, err
			}
			if err = copyCO(half + i); err != nil {
				return nil, err
			}
		}
	} else {
		for i := 0; i < m.PrimaryOutputCount(); i++ {
			if err = copyCO(i); err != nil {
				return nil, err
			}
			if _, err = g.AppendCO(GiaConst0); err != nil {
				return nil, err
			}
		}
	}
	for i := m.PrimaryOutputCount(); i < len(m.cos); i++ {
		if err = copyCO(i); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	g.registers = m.registers
	if stopped() {
		return nil, ErrCancelled
	}
	if interleave {
		result, _, e := g.Cleanup()
		if e != nil {
			return nil, e
		}
		return result, nil
	}
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
