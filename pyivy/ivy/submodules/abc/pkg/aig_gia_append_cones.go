package abc

// Additional property cones from src/aig/gia/giaDup.c:Gia_ManDupAppendCones.

// AppendPropertyCones adds single-output combinational cones as ordinary POs.
// Cone inputs map to all source CIs, or just ROs when onlyRegisters is true.
// Output order is original ordinary POs, cones, original constraints, RIs.
// Constraint roles are preserved (C drops them). Name/spec and CI/register
// order survive cleanup. Extended gates are lowered through hashing; source
// graphs are unchanged. MaxMappingEntries bounds summed source object spans.
func (m *GiaManager) AppendPropertyCones(cones []*GiaManager, onlyRegisters bool, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	entries := len(m.objects)
	if entries > options.MaxMappingEntries || len(cones) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	expected := m.InputCount()
	offset := 0
	if onlyRegisters {
		expected = m.registers
		offset = m.PrimaryInputCount()
	}
	for _, cone := range cones {
		if cone == nil {
			return nil, ErrDimension
		}
		if err := cone.Check(); err != nil {
			return nil, err
		}
		if cone.registers != 0 || cone.constraints != 0 || len(cone.cos) != 1 || cone.InputCount() != expected {
			return nil, ErrDimension
		}
		var err error
		entries, err = checkedSize(entries, 1, len(cone.objects), options.MaxMappingEntries)
		if err != nil {
			return nil, err
		}
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
	copyCO := func(id GiaID) error {
		if stopped() {
			return ErrCancelled
		}
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
	ordinary := m.PrimaryOutputCount() - m.constraints
	for _, id := range m.cos[:ordinary] {
		if err = copyCO(id); err != nil {
			return nil, err
		}
	}
	for _, cone := range cones {
		if stopped() {
			return nil, ErrCancelled
		}
		copies := make([]GiaLit, len(cone.objects))
		for i := range copies {
			copies[i] = InvalidGiaLit
		}
		copies[0] = GiaConst0
		for i, id := range cone.cis {
			copies[id] = mapping[m.cis[offset+i]]
		}
		for i, obj := range cone.objects {
			if !obj.IsAnd() {
				continue
			}
			if stopped() {
				return nil, ErrCancelled
			}
			copies[i], err = hashGiaCopiedGate(cone, g, GiaID(i), copies)
			if err != nil {
				return nil, err
			}
		}
		co := cone.cos[0]
		driver, e := cone.objects[co].Fanin0(co)
		if e != nil {
			return nil, e
		}
		l, e := mappedGiaSignal(copies, driver)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(l); err != nil {
			return nil, err
		}
	}
	for _, id := range m.cos[ordinary:] {
		if err = copyCO(id); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	g.registers, g.constraints = m.registers, m.constraints
	if stopped() {
		return nil, ErrCancelled
	}
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}
