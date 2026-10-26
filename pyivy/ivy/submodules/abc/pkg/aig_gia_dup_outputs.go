package abc

// Internal-node outputs from src/aig/gia/giaDup.c:Gia_ManDupOuts.

// DuplicateWithGateOutputs preserves all logic and exposes each AND-like
// object's regular value as an additional ordinary PO, in source ID order.
// Output order is original ordinary POs, gate POs, original constraints, RIs.
// Moving constraints after gate POs preserves their suffix role (C drops the
// constraint count). Register layout, name/spec and raw node kinds survive.
// The owned map includes source CO-to-target-CO object literals.
func (m *GiaManager) DuplicateWithGateOutputs(options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	if len(m.objects) > options.MaxMappingEntries {
		return nil, nil, ErrLimit
	}
	if _, err := checkedSize(len(m.objects), 1, m.AndCount(), options.MaxObjects); err != nil {
		return nil, nil, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, nil, ErrCancelled
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
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, nil, ErrCancelled
		}
		mapping[i], err = appendGiaCopiedGate(m, g, GiaID(i), mapping)
		if err != nil {
			return nil, nil, err
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
		mapping[id], e = g.AppendCO(l)
		return e
	}
	ordinary := m.PrimaryOutputCount() - m.constraints
	for _, id := range m.cos[:ordinary] {
		if err = copyCO(id); err != nil {
			return nil, nil, err
		}
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, nil, ErrCancelled
		}
		if _, err = g.AppendCO(mapping[i]); err != nil {
			return nil, nil, err
		}
	}
	for _, id := range m.cos[ordinary:] {
		if err = copyCO(id); err != nil {
			return nil, nil, err
		}
	}
	g.registers, g.constraints = m.registers, m.constraints
	if err = g.Check(); err != nil {
		return nil, nil, err
	}
	return g, mapping, nil
}
