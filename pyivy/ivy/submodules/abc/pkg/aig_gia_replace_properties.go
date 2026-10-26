package abc

// Transition/property composition from src/aig/gia/giaDup.c:Gia_ManDupWithNewPo.

// WithStateProperties replaces ordinary POs with a combinational property's
// outputs over register state. Property CI i maps to source RO i. Original RIs
// and CI/reset ordering stay intact. Output order is new ordinary properties,
// property constraints, original constraints, RIs. Both constraint sets survive
// (C drops original POs and constraint roles). Name/spec come from transition.
// Extended gates are hashed/lowered and unused logic is cleaned up.
func (m *GiaManager) WithStateProperties(property *GiaManager, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || property == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	for _, source := range []*GiaManager{m, property} {
		if e := source.Check(); e != nil {
			return nil, e
		}
	}
	if property.registers != 0 || property.InputCount() != m.registers {
		return nil, ErrDimension
	}
	if _, e := checkedSize(len(m.objects), 1, len(property.objects), options.MaxMappingEntries); e != nil {
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
	lower := make([]GiaLit, len(m.objects))
	upper := make([]GiaLit, len(property.objects))
	for _, mapping := range [][]GiaLit{lower, upper} {
		for i := range mapping {
			mapping[i] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
	}
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		lower[id], err = g.AppendCI()
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
		lower[i], err = hashGiaCopiedGate(m, g, GiaID(i), lower)
		if err != nil {
			return nil, err
		}
	}
	for i, id := range property.cis {
		upper[id] = lower[m.cis[m.PrimaryInputCount()+i]]
	}
	for i, obj := range property.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		upper[i], err = hashGiaCopiedGate(property, g, GiaID(i), upper)
		if err != nil {
			return nil, err
		}
	}
	copyCO := func(source *GiaManager, mapping []GiaLit, id GiaID) error {
		if stopped() {
			return ErrCancelled
		}
		driver, e := source.objects[id].Fanin0(id)
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
	for _, id := range property.cos {
		if err = copyCO(property, upper, id); err != nil {
			return nil, err
		}
	}
	for _, id := range m.cos[m.PrimaryOutputCount()-m.constraints:] {
		if err = copyCO(m, lower, id); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	g.registers, g.constraints = m.registers, m.constraints+property.constraints
	if stopped() {
		return nil, ErrCancelled
	}
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}
