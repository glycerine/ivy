package abc

// Primary-input cofactors from src/aig/gia/giaDup.c:Gia_ManDupCofactorVar.

// CofactorPrimaryInput fixes one PI to value while retaining every CI,
// including the now-unused pivot CI. POs, register inputs, constraints and
// metadata retain their order. Register outputs cannot be selected. This is
// a fixed cofactor, unlike CofactorRebuild's equivalent reconstruction.
func (m *GiaManager) CofactorPrimaryInput(index int, value bool, options GiaCofactorOptions) (*GiaManager, error) {
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
	if index < 0 || index >= m.PrimaryInputCount() {
		return nil, ErrDimension
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	nodes, err := m.CollectConeAnds(m.cos, nil, options.MaxMappingEntries)
	if err != nil {
		return nil, err
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
	mapping[m.cis[index]] = GiaConst0
	if value {
		mapping[m.cis[index]] = GiaConst1
	}
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	translate := func(l GiaLit) (GiaLit, error) {
		if !mapping[l.ID()].Valid() {
			return InvalidGiaLit, ErrInvariant
		}
		return mapping[l.ID()].NotIf(l.Complemented()), nil
	}
	for _, id := range nodes {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = hashGiaCopiedGate(m, g, id, mapping)
		if err != nil {
			return nil, err
		}
	}
	for _, id := range m.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, err
		}
		copied, err := translate(driver)
		if err != nil {
			return nil, err
		}
		if _, err = g.AppendCO(copied); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	if err = g.SetRegisterCount(m.registers); err != nil {
		return nil, err
	}
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
