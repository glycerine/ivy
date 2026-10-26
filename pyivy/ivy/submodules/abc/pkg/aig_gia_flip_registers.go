package abc

// Register phase recoding from src/aig/gia/giaDup.c:Gia_ManDupFlip.

// FlipRegisters changes state encoding to newState = oldState XOR phases.
// POs/constraints keep their functions under that state mapping; each RI is
// recoded by the same phase. CI/CO order, raw extended gate kinds, dangling
// logic and name/spec survive. Graphs have no reset list: callers must XOR
// known initial values by phases (unknown initial values remain unknown).
func (m *GiaManager) FlipRegisters(phases []bool, options GiaCofactorOptions) (*GiaManager, error) {
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
	if len(phases) != m.registers {
		return nil, ErrDimension
	}
	if len(m.objects) > options.MaxMappingEntries || len(m.objects) > options.MaxObjects {
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
	for i, obj := range m.objects {
		if stopped() {
			return nil, ErrCancelled
		}
		id := GiaID(i)
		if obj.IsCI() {
			mapping[id], err = g.AppendCI()
			if err != nil {
				return nil, err
			}
			index, e := obj.TerminalIndex()
			if e != nil {
				return nil, e
			}
			if index >= m.PrimaryInputCount() {
				mapping[id] = mapping[id].NotIf(phases[index-m.PrimaryInputCount()])
			}
		} else if obj.IsAnd() {
			mapping[id], err = appendGiaCopiedGate(m, g, id, mapping)
		} else if obj.IsCO() {
			driver, e := obj.Fanin0(id)
			if e != nil {
				return nil, e
			}
			copied, e := mappedGiaSignal(mapping, driver)
			if e != nil {
				return nil, e
			}
			index, e := obj.TerminalIndex()
			if e != nil {
				return nil, e
			}
			if index >= m.PrimaryOutputCount() {
				copied = copied.NotIf(phases[index-m.PrimaryOutputCount()])
			}
			mapping[id], err = g.AppendCO(copied)
		}
		if err != nil {
			return nil, err
		}
	}
	g.registers, g.constraints = m.registers, m.constraints
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
