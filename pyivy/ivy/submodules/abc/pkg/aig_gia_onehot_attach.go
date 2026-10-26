package abc

// One-hot integration from src/aig/gia/giaDup.c:Gia_ManDupOneHot.

// WithOneHotAssumptions copies the graph and constrains exactly one CI true:
// all PIs for a combinational graph, or register outputs for a sequential one.
// Original POs/constraints keep their order, new positive assumptions follow,
// and RIs remain last. C appends violation outputs after RIs; Go adapts both
// polarity and register placement. Original raw node kinds and metadata survive.
func (m *GiaManager) WithOneHotAssumptions(options GiaCofactorOptions) (*GiaManager, error) {
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
	skips, variables := 0, m.InputCount()
	if m.registers > 0 {
		skips, variables = m.PrimaryInputCount(), m.registers
	}
	generator, err := NewGiaOneHot(skips, variables, options)
	if err != nil {
		return nil, err
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
	if len(generator.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	generated := make([]GiaLit, len(generator.objects))
	for i := range generated {
		generated[i] = InvalidGiaLit
	}
	generated[0] = GiaConst0
	for index, id := range generator.cis {
		generated[id] = mapping[m.cis[index]]
	}
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for i, obj := range generator.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		generated[i], err = hashGiaCopiedGate(generator, g, GiaID(i), generated)
		if err != nil {
			return nil, err
		}
	}
	g.HashStop()
	copyCO := func(id GiaID) error {
		if stopped() {
			return ErrCancelled
		}
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return e
		}
		copied, e := mappedGiaSignal(mapping, driver)
		if e != nil {
			return e
		}
		_, e = g.AppendCO(copied)
		return e
	}
	for _, id := range m.cos[:m.PrimaryOutputCount()] {
		if err = copyCO(id); err != nil {
			return nil, err
		}
	}
	for _, id := range generator.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, e := generator.objects[id].Fanin0(id)
		if e != nil {
			return nil, e
		}
		copied, e := mappedGiaSignal(generated, driver)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(copied.Not()); err != nil {
			return nil, err
		}
	}
	for _, id := range m.cos[m.PrimaryOutputCount():] {
		if err = copyCO(id); err != nil {
			return nil, err
		}
	}
	g.registers, g.constraints = m.registers, m.constraints+generator.OutputCount()
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
