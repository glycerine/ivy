package abc

// Object-order and zero-output copies from src/aig/gia/giaDup.c:
// Gia_ManDupOrderAiger and Gia_ManDupZero.

// NormalizeObjectOrder puts all CIs first, then gates in their original
// physical order, then COs. It retains dangling gates, raw extended node kinds,
// terminal/register/reset order and name/spec. Constraint roles clear as in C.
// The owned source-object map includes CO object literals. Source scratch is
// unchanged. Choice/equivalence metadata outside GiaManager is not copied.
func (m *GiaManager) NormalizeObjectOrder(options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, nil, ErrDimension
	}
	if options.Stop != nil && options.Stop() {
		return nil, nil, ErrCancelled
	}
	gates, e := m.AndObjects(options.MaxMappingEntries)
	if e != nil {
		return nil, nil, e
	}
	return m.DuplicateFromLists(m.cis, gates, m.cos, m.registers, options)
}

// ZeroOutputCopy retains CI/CO/register counts but replaces every CO driver,
// including RIs, with false. All original gates are omitted. Only Name survives;
// Spec and constraint roles clear, matching C. Terminal order and zero-based
// register positions stay unchanged. This creates a new circuit and makes no
// equivalence claim. Source scratch is unchanged.
func (m *GiaManager) ZeroOutputCopy(options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return nil, ErrLimit
	}
	capObjects := options.MaxObjects
	if capObjects == 0 {
		capObjects = 1 << 20
	}
	count, e := checkedSize(1, 1, len(m.cis), capObjects)
	if e != nil {
		return nil, e
	}
	if _, e = checkedSize(count, 1, len(m.cos), capObjects); e != nil {
		return nil, e
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	g, e := NewGiaManagerWithLimit(capObjects)
	if e != nil {
		return nil, e
	}
	g.Name = m.Name
	for range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		if _, e = g.AppendCI(); e != nil {
			return nil, e
		}
	}
	for range m.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		if _, e = g.AppendCO(GiaConst0); e != nil {
			return nil, e
		}
	}
	g.registers = m.registers
	return g, nil
}
