package abc

// Graph append from src/aig/gia/giaDup.c:Gia_ManDupAppend/AppendShare.

// WithAppendedGraph returns an owned clone with every source object appended.
// In shared mode source CIs bind by ordinal to existing CIs (counts must match)
// and gates use hashing, lowering extended nodes. In independent mode all CIs
// append and raw gate types survive. Neither mode removes dangling logic.
// The source-object map includes appended CO literals; base IDs stay unchanged.
//
// Role treatment matches C's positional operations, not sequential composition:
// independent mode clears register count; shared mode retains the base count,
// so the final appended COs become RIs and former RIs may become POs. Existing
// constraint count remains positional; source roles are not imported. Name/spec
// come from base. Hashing remains active as in C. Both inputs stay unchanged,
// including when source==base; failures expose no partial graph or map.
func (m *GiaManager) WithAppendedGraph(source *GiaManager, shared bool, options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if m == nil || source == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, nil, ErrDimension
	}
	for _, g := range []*GiaManager{m, source} {
		if e := g.Check(); e != nil {
			return nil, nil, e
		}
	}
	if shared && len(m.cis) != len(source.cis) {
		return nil, nil, ErrDimension
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if _, e := checkedSize(len(m.objects), 1, len(source.objects), limit); e != nil {
		return nil, nil, e
	}
	capObjects := options.MaxObjects
	if capObjects == 0 {
		capObjects = 1 << 20
	}
	if len(m.objects) > capObjects {
		return nil, nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, nil, ErrCancelled
	}
	g := m.Clone()
	g.objectLimit = capObjects
	if !shared {
		g.registers = 0
	}
	if g.hash == nil {
		if e := g.HashStart(); e != nil {
			return nil, nil, e
		}
	}
	mapping := make([]GiaLit, len(source.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for i, o := range source.objects {
		if stopped() {
			return nil, nil, ErrCancelled
		}
		id := GiaID(i)
		var e error
		if o.IsCI() {
			if shared {
				index, err := o.TerminalIndex()
				if err != nil {
					return nil, nil, err
				}
				mapping[id] = GiaLit(g.cis[index]) << 1
			} else {
				mapping[id], e = g.AppendCI()
			}
		}
		if o.IsAnd() {
			if shared {
				mapping[id], e = hashGiaCopiedGate(source, g, id, mapping)
			} else {
				mapping[id], e = appendGiaCopiedGate(source, g, id, mapping)
			}
		}
		if o.IsCO() {
			d, err := o.Fanin0(id)
			if err != nil {
				return nil, nil, err
			}
			v, err := mappedGiaSignal(mapping, d)
			if err != nil {
				return nil, nil, err
			}
			mapping[id], e = g.AppendCO(v)
		}
		if e != nil {
			return nil, nil, e
		}
	}
	if stopped() {
		return nil, nil, ErrCancelled
	}
	if e := g.Check(); e != nil {
		return nil, nil, e
	}
	return g, mapping, nil
}
