package abc

// Trailing-PI retention from src/aig/gia/giaDup.c:Gia_ManDupLastPis.

// RetainLastPrimaryInputs keeps the last count CIs of a combinational graph.
// Every gate, including dangling logic, and every CO is raw-copied. Any
// reference to an omitted input is an error: this operation does not fix
// input values or trim dependent logic. Zero/all counts are allowed. Constraint
// roles, extended gate kinds and name/spec survive; source scratch is unchanged.
func (m *GiaManager) RetainLastPrimaryInputs(count int, options GiaCofactorOptions) (*GiaManager, error) {
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
	if m.registers != 0 || count < 0 || count > len(m.cis) {
		return nil, ErrDimension
	}
	if len(m.objects) > options.MaxMappingEntries || len(m.objects)-(len(m.cis)-count) > options.MaxObjects {
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
	for _, id := range m.cis[len(m.cis)-count:] {
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
	}
	g.constraints = m.constraints
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
