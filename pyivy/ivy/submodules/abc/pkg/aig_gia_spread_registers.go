package abc

// Register gaps from src/aig/gia/giaDup.c:Gia_ManDupSpreadFlop/PermFlopGap.

// SpreadRegisters inserts dummy registers at mask entries -1. Every other
// entry consumes the next source register, ignoring the entry value as in C.
// Exactly RegisterCount non-gap entries are required. Dummy RIs are constant
// zero; dummy ROs are unused. Callers project resets into the non-gap slots.
func (m *GiaManager) SpreadRegisters(mask []int, options GiaCofactorOptions) (*GiaManager, error) {
	return m.spreadRegisters(mask, false, options)
}

// PermuteRegistersWithGaps interprets non-gap entries as source register
// indices in new-to-old order. They must form a bijection; -1 inserts a dummy
// register. Both APIs preserve PO/constraint roles, raw gate kinds and metadata.
// Dummy resets are caller-selected; zero initialization matches their zero RIs.
func (m *GiaManager) PermuteRegistersWithGaps(mask []int, options GiaCofactorOptions) (*GiaManager, error) {
	return m.spreadRegisters(mask, true, options)
}

func (m *GiaManager) spreadRegisters(mask []int, permute bool, options GiaCofactorOptions) (*GiaManager, error) {
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
	if len(mask) < m.registers {
		return nil, ErrDimension
	}
	if _, e := checkedSize(len(m.objects), 1, len(mask), options.MaxMappingEntries); e != nil {
		return nil, e
	}
	if _, e := checkedSize(len(m.objects), 2, len(mask)-m.registers, options.MaxObjects); e != nil {
		return nil, e
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	seen := make([]bool, m.registers)
	count := 0
	for _, entry := range mask {
		if entry == -1 {
			continue
		}
		index := count
		if permute {
			index = entry
		}
		if index < 0 || index >= m.registers || seen[index] {
			return nil, ErrDimension
		}
		seen[index] = true
		count++
	}
	if count != m.registers {
		return nil, ErrDimension
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
	for _, id := range m.cis[:m.PrimaryInputCount()] {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	count = 0
	for _, entry := range mask {
		if stopped() {
			return nil, ErrCancelled
		}
		l, e := g.AppendCI()
		if e != nil {
			return nil, e
		}
		if entry != -1 {
			index := count
			if permute {
				index = entry
			}
			mapping[m.cis[m.PrimaryInputCount()+index]] = l
			count++
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
	copyDriver := func(id GiaID) (GiaLit, error) {
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return InvalidGiaLit, e
		}
		return mappedGiaSignal(mapping, driver)
	}
	for _, id := range m.cos[:m.PrimaryOutputCount()] {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, e := copyDriver(id)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	count = 0
	for _, entry := range mask {
		if stopped() {
			return nil, ErrCancelled
		}
		driver := GiaConst0
		if entry != -1 {
			index := count
			if permute {
				index = entry
			}
			driver, err = copyDriver(m.cos[m.PrimaryOutputCount()+index])
			if err != nil {
				return nil, err
			}
			count++
		}
		if _, err = g.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	g.registers, g.constraints = len(mask), m.constraints
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
