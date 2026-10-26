package abc

// Independent graph copies from src/aig/gia/giaDup.c:Gia_ManDupTimes.

// Replicate builds copies independent copies sharing only constant zero.
// Logical CI order groups all copy PIs before all copy ROs. CO order groups
// ordinary POs, constraints, then RIs; each group is copy-major. Register/reset
// order is copy-major. Raw extended gates, dangling logic and name/spec survive.
// Go normalizes terminal placement instead of C's reordered physical tables.
// Source scratch is unchanged.
func (m *GiaManager) Replicate(copies int, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || copies < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	if _, e := checkedSize(1, len(m.objects)-1, copies, options.MaxObjects); e != nil {
		return nil, e
	}
	if copies > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	if _, e := checkedSize(len(m.objects), len(m.cis)+len(m.cos), copies, options.MaxMappingEntries); e != nil {
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
	pis := make([]GiaLit, copies*m.PrimaryInputCount())
	ros := make([]GiaLit, copies*m.registers)
	for _, list := range [][]GiaLit{pis, ros} {
		for i := range list {
			if stopped() {
				return nil, ErrCancelled
			}
			list[i], err = g.AppendCI()
			if err != nil {
				return nil, err
			}
		}
	}
	mapping := make([]GiaLit, len(m.objects))
	drivers := make([]GiaLit, copies*len(m.cos))
	for copyIndex := 0; copyIndex < copies; copyIndex++ {
		if stopped() {
			return nil, ErrCancelled
		}
		for i := range mapping {
			mapping[i] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
		for i, id := range m.cis {
			if i < m.PrimaryInputCount() {
				mapping[id] = pis[copyIndex*m.PrimaryInputCount()+i]
			} else {
				mapping[id] = ros[copyIndex*m.registers+i-m.PrimaryInputCount()]
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
		for i, id := range m.cos {
			driver, e := m.objects[id].Fanin0(id)
			if e != nil {
				return nil, e
			}
			drivers[copyIndex*len(m.cos)+i], err = mappedGiaSignal(mapping, driver)
			if err != nil {
				return nil, err
			}
		}
	}
	for role := 0; role < 3; role++ {
		lo, hi := 0, m.PrimaryOutputCount()-m.constraints
		if role == 1 {
			lo, hi = hi, m.PrimaryOutputCount()
		} else if role == 2 {
			lo, hi = m.PrimaryOutputCount(), len(m.cos)
		}
		for copyIndex := 0; copyIndex < copies; copyIndex++ {
			for index := lo; index < hi; index++ {
				if stopped() {
					return nil, ErrCancelled
				}
				if _, err = g.AppendCO(drivers[copyIndex*len(m.cos)+index]); err != nil {
					return nil, err
				}
			}
		}
	}
	g.registers, g.constraints = copies*m.registers, copies*m.constraints
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
