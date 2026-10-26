package abc

// Explicit-list duplication from src/aig/gia/giaDup.c:Gia_ManDupFromVecs.

// DuplicateFromLists copies unique source CIs, gates and COs in the supplied
// order. Gates must follow all their dependencies; omitted dependencies and
// duplicates are errors. Unselected objects map to InvalidGiaLit. The final
// registers CIs/COs become register outputs/inputs, irrespective of their
// source roles. As in C, constraint roles are not inherited: this API builds
// an explicitly specified graph. Name/spec and raw node kinds are retained.
func (m *GiaManager) DuplicateFromLists(cis, gates, cos []GiaID, registers int, options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if registers < 0 || registers > len(cis) || registers > len(cos) || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	count, err := checkedSize(1, 1, len(cis), options.MaxObjects)
	if err != nil {
		return nil, nil, err
	}
	count, err = checkedSize(count, 1, len(gates), options.MaxObjects)
	if err != nil {
		return nil, nil, err
	}
	if _, err = checkedSize(count, 1, len(cos), options.MaxObjects); err != nil {
		return nil, nil, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, nil, ErrCancelled
	}
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	for kind, list := range [][]GiaID{cis, gates, cos} {
		for _, id := range list {
			if stopped() {
				return nil, nil, ErrCancelled
			}
			obj, e := m.Object(id)
			if e != nil {
				return nil, nil, e
			}
			if mapping[id].Valid() {
				return nil, nil, ErrDimension
			}
			switch kind {
			case 0:
				if !obj.IsCI() {
					return nil, nil, ErrDimension
				}
				mapping[id], err = g.AppendCI()
			case 1:
				if !obj.IsAnd() {
					return nil, nil, ErrDimension
				}
				mapping[id], err = appendGiaCopiedGate(m, g, id, mapping)
			case 2:
				if !obj.IsCO() {
					return nil, nil, ErrDimension
				}
				driver, e := obj.Fanin0(id)
				if e != nil {
					return nil, nil, e
				}
				copied, e := mappedGiaSignal(mapping, driver)
				if e != nil {
					return nil, nil, e
				}
				mapping[id], err = g.AppendCO(copied)
			}
			if err != nil {
				return nil, nil, err
			}
		}
	}
	g.registers = registers
	if err = g.Check(); err != nil {
		return nil, nil, err
	}
	return g, mapping, nil
}
