package abc

// PI quantification from src/aig/gia/giaDup.c:Gia_ManDupExist/Gia_ManDupUniv.

// ExistentialInput returns f[x=0] OR f[x=1] for a single-output combinational
// graph. All CIs, including the unused quantified CI, remain in original order.
func (m *GiaManager) ExistentialInput(index int, options GiaCofactorOptions) (*GiaManager, error) {
	return m.quantifyInput(index, true, options)
}

// UniversalFirstOutput returns f0[x=0] AND f0[x=1] for output zero; every later
// output is its positive cofactor f_i[x=1], matching C's Gia_ManDupUniv.
// This deliberately does not universally quantify every output. Only
// combinational graphs without constraint roles are accepted. All CIs remain.
func (m *GiaManager) UniversalFirstOutput(index int, options GiaCofactorOptions) (*GiaManager, error) {
	return m.quantifyInput(index, false, options)
}

// ExistentialAllOutputs independently ORs each CO's two PI cofactors,
// including constraints and register inputs, as in Gia_ManDupExist2. Register
// and constraint roles survive. This is coordinate-wise abstraction: separate
// outputs may require different witnesses, so it is not existential elimination
// of the joint transition relation. The quantified PI remains as an unused CI.
func (m *GiaManager) ExistentialAllOutputs(index int, options GiaCofactorOptions) (*GiaManager, error) {
	return m.quantifyInputMode(index, true, true, options)
}

func (m *GiaManager) quantifyInput(index int, exists bool, options GiaCofactorOptions) (*GiaManager, error) {
	return m.quantifyInputMode(index, exists, false, options)
}
func (m *GiaManager) quantifyInputMode(index int, exists, all bool, options GiaCofactorOptions) (*GiaManager, error) {
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
	if index < 0 || index >= m.PrimaryInputCount() || (!all && (m.registers != 0 || m.constraints != 0 || len(m.cos) == 0 || (exists && len(m.cos) != 1))) {
		return nil, ErrDimension
	}
	if _, e := checkedSize(len(m.objects), 2, len(m.cos), options.MaxMappingEntries); e != nil {
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
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	var nodes []GiaID
	if all {
		nodes, err = m.CollectTransitiveFanins(m.cos, options.MaxMappingEntries)
		if err != nil {
			return nil, err
		}
	} else {
		for i, obj := range m.objects {
			if obj.IsAnd() {
				nodes = append(nodes, GiaID(i))
			}
		}
	}
	negative, positive := make([]GiaLit, len(m.cos)), make([]GiaLit, len(m.cos))
	for phase := 0; phase < 2; phase++ {
		mapping[m.cis[index]] = GiaLit(phase)
		for _, id := range nodes {
			if stopped() {
				return nil, ErrCancelled
			}
			mapping[id], err = hashGiaCopiedGate(m, g, id, mapping)
			if err != nil {
				return nil, err
			}
		}
		results := negative
		if phase == 1 {
			results = positive
		}
		for i, id := range m.cos {
			driver, e := m.objects[id].Fanin0(id)
			if e != nil {
				return nil, e
			}
			results[i], err = mappedGiaSignal(mapping, driver)
			if err != nil {
				return nil, err
			}
		}
	}
	for i, driver := range positive {
		if stopped() {
			return nil, ErrCancelled
		}
		if exists {
			driver, err = g.HashOr(driver, negative[i])
		} else if i == 0 {
			driver, err = g.HashAnd(driver, negative[i])
		}
		if err != nil {
			return nil, err
		}
		if _, err = g.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	if all {
		g.registers, g.constraints = m.registers, m.constraints
	}
	if stopped() {
		return nil, ErrCancelled
	}
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}
