package abc

// Shannon reconstruction from src/aig/gia/giaCof.c:Gia_ManDupCofInt.

type GiaCofactorOptions struct {
	MaxObjects, MaxMappingEntries int
	Stop                          func() bool
}

// CofactorRebuild duplicates both cofactors of a CI or gate pivot and combines
// each CO using the original pivot function. The graph is functionally
// equivalent, not a graph with the pivot fixed. Intermediate nodes are retained
// as in C. The returned source-object map contains positive-cofactor signals;
// COs map to new CO objects. Source scratch remains unchanged. Zero caps select
// one million objects/mapping entries. Register/constraint metadata is retained.
func (m *GiaManager) CofactorRebuild(pivot GiaID, options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	if pivot == 0 || uint64(pivot) >= uint64(len(m.objects)) || (!m.objects[pivot].IsCI() && !m.objects[pivot].IsAnd()) {
		return nil, nil, ErrInvalidLiteral
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, nil, ErrCancelled
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, nil, err
	}
	target.Name, target.Spec = m.Name, m.Spec
	if err = target.HashStart(); err != nil {
		return nil, nil, err
	}
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	inputs := make([]GiaLit, len(m.cis))
	for i := range inputs {
		if stopped() {
			return nil, nil, ErrCancelled
		}
		inputs[i], err = target.AppendCI()
		if err != nil {
			return nil, nil, err
		}
	}
	translate := func(l GiaLit) (GiaLit, error) {
		if !mapping[l.ID()].Valid() {
			return InvalidGiaLit, ErrInvariant
		}
		return mapping[l.ID()].NotIf(l.Complemented()), nil
	}
	control := InvalidGiaLit
	negative := make([]GiaLit, len(m.cos))
	for phase := 0; phase < 2; phase++ {
		for i, id := range m.cis {
			mapping[id] = inputs[i]
			if id == pivot {
				control = inputs[i]
				mapping[id] = GiaLit(phase)
			}
		}
		for i, obj := range m.objects {
			if stopped() {
				return nil, nil, ErrCancelled
			}
			if !obj.IsAnd() {
				continue
			}
			id := GiaID(i)
			first, err := obj.Fanin0(id)
			if err != nil {
				return nil, nil, err
			}
			a, err := translate(first)
			if err != nil {
				return nil, nil, err
			}
			value := a
			if !obj.IsBuffer() {
				second, err := obj.Fanin1(id)
				if err != nil {
					return nil, nil, err
				}
				b, err := translate(second)
				if err != nil {
					return nil, nil, err
				}
				if c, yes := m.MuxControl(id); yes {
					condition, err := translate(c)
					if err != nil {
						return nil, nil, err
					}
					value, err = target.HashMux(condition, b, a)
					if err != nil {
						return nil, nil, err
					}
				} else if obj.IsXor() {
					value, err = target.HashXor(a, b)
				} else {
					value, err = target.HashAnd(a, b)
				}
				if err != nil {
					return nil, nil, err
				}
			}
			mapping[id] = value
			if id == pivot {
				if phase == 0 {
					control = value
				}
				mapping[id] = GiaLit(phase)
			}
		}
		if phase == 0 {
			for i, id := range m.cos {
				driver, err := m.objects[id].Fanin0(id)
				if err != nil {
					return nil, nil, err
				}
				negative[i], err = translate(driver)
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}
	if !control.Valid() {
		return nil, nil, ErrInvariant
	}
	for i, id := range m.cos {
		if stopped() {
			return nil, nil, ErrCancelled
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, nil, err
		}
		positive, err := translate(driver)
		if err != nil {
			return nil, nil, err
		}
		value, err := target.HashMux(control, positive, negative[i])
		if err != nil {
			return nil, nil, err
		}
		mapping[id], err = target.AppendCO(value)
		if err != nil {
			return nil, nil, err
		}
	}
	target.HashStop()
	if err = target.SetRegisterCount(m.registers); err != nil {
		return nil, nil, err
	}
	if err = target.SetConstraintCount(m.constraints); err != nil {
		return nil, nil, err
	}
	if err = target.Check(); err != nil {
		return nil, nil, err
	}
	return target, mapping, nil
}
