package abc

// Graph normalization from src/aig/gia/giaTim.c:Gia_ManDupNormalize,
// without timing-box/mapped-cell/equivalence metadata.

// NormalizeGraph orders all CIs, then physical-order gates, then COs, keeping
// CI/CO/register/reset order, name/spec and constraint roles. hashLogic merges
// and simplifies gates using hashing, lowering native XOR/mux nodes; buffers
// remain explicit objects as in C. Otherwise raw gate kinds survive. Neither
// mode removes dangling objects. Returned owned map includes all CO literals;
// source scratch/hash state stays unchanged. The result's hashing is stopped.
// Timing-box ordering and mapped-cell/equivalence metadata are not represented
// by GiaManager and are outside this operation.
func (m *GiaManager) NormalizeGraph(hashLogic bool, options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, nil, ErrDimension
	}
	if !hashLogic {
		g, mapping, e := m.NormalizeObjectOrder(options)
		if e != nil {
			return nil, nil, e
		}
		g.constraints = m.constraints
		return g, mapping, nil
	}
	if e := m.Check(); e != nil {
		return nil, nil, e
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return nil, nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, nil, ErrCancelled
	}
	capObjects := options.MaxObjects
	if capObjects == 0 {
		capObjects = 1 << 20
	}
	g, e := NewGiaManagerWithLimit(capObjects)
	if e != nil {
		return nil, nil, e
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
		mapping[id], e = g.AppendCI()
		if e != nil {
			return nil, nil, e
		}
	}
	if e = g.HashStart(); e != nil {
		return nil, nil, e
	}
	for i, o := range m.objects {
		if !o.IsAnd() {
			continue
		}
		if stopped() {
			return nil, nil, ErrCancelled
		}
		if o.IsBuffer() {
			mapping[i], e = appendGiaCopiedGate(m, g, GiaID(i), mapping)
		} else {
			mapping[i], e = hashGiaCopiedGate(m, g, GiaID(i), mapping)
		}
		if e != nil {
			return nil, nil, e
		}
	}
	g.HashStop()
	for _, id := range m.cos {
		if stopped() {
			return nil, nil, ErrCancelled
		}
		d, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, nil, err
		}
		v, err := mappedGiaSignal(mapping, d)
		if err != nil {
			return nil, nil, err
		}
		mapping[id], e = g.AppendCO(v)
		if e != nil {
			return nil, nil, e
		}
	}
	g.registers, g.constraints = m.registers, m.constraints
	if e = g.Check(); e != nil {
		return nil, nil, e
	}
	return g, mapping, nil
}
