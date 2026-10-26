package abc

// Guarded object cofactors from src/aig/gia/giaDup.c:Gia_ManDupCofactorObj.

// CofactorObjectGuarded fixes a combinational CI/gate pivot and guards every
// CO by the original pivot equalling value. Thus each output is equivalent to
// original_output AND (original_pivot == value), unlike an unguarded PI
// cofactor. All CIs and constraint metadata are retained. Source scratch is
// unchanged. COs preceding the pivot in object order are handled safely.
func (m *GiaManager) CofactorObjectGuarded(pivot GiaID, value bool, options GiaCofactorOptions) (*GiaManager, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	if m.registers != 0 {
		return nil, ErrDimension
	}
	if pivot == 0 || uint64(pivot) >= uint64(len(m.objects)) || (!m.objects[pivot].IsCI() && !m.objects[pivot].IsAnd()) {
		return nil, ErrInvalidLiteral
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	roots := make([]GiaID, 0, len(m.cos)+1)
	roots = append(roots, m.cos...)
	roots = append(roots, pivot)
	nodes, err := m.CollectConeAnds(roots, nil, options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	guard := InvalidGiaLit
	fixed := GiaConst0
	if value {
		fixed = GiaConst1
	}
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
		if id == pivot {
			guard = mapping[id].NotIf(!value)
			mapping[id] = fixed
		}
	}
	for _, id := range nodes {
		if stopped() {
			return nil, ErrCancelled
		}
		copied, err := hashGiaCopiedGate(m, g, id, mapping)
		if err != nil {
			return nil, err
		}
		mapping[id] = copied
		if id == pivot {
			guard = copied.NotIf(!value)
			mapping[id] = fixed
		}
	}
	if !guard.Valid() {
		return nil, ErrInvariant
	}
	for _, id := range m.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, err
		}
		copied, err := mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, err
		}
		guarded, err := g.HashAnd(copied, guard)
		if err != nil {
			return nil, err
		}
		if _, err = g.AppendCO(guarded); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	if err = g.SetConstraintCount(m.constraints); err != nil {
		return nil, err
	}
	if stopped() {
		return nil, ErrCancelled
	}
	if g.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	cleaned, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return cleaned, nil
}
