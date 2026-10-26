package abc

// Controlled register self-loops from src/aig/gia/giaDup.c:Gia_ManDupSelf.

// WithRegisterHoldControl adds the first PI as a shared register update enable.
// True selects original next state; false holds each current register value.
// Original PIs shift by one, RO order and all PO functions remain unchanged.
// Constraint roles/name/spec survive; extended gates are lowered by hashing.
// Even a combinational source receives the unused control PI, matching C.
// Graphs carry no latch initialization; callers retain the original reset list.
func (m *GiaManager) WithRegisterHoldControl(options GiaCofactorOptions) (*GiaManager, error) {
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
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
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
	control, err := g.AppendCI()
	if err != nil {
		return nil, err
	}
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[i], err = hashGiaCopiedGate(m, g, GiaID(i), mapping)
		if err != nil {
			return nil, err
		}
	}
	drivers := make([]GiaLit, len(m.cos))
	for index, id := range m.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return nil, e
		}
		drivers[index], err = mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, err
		}
		if index >= m.PrimaryOutputCount() {
			state := mapping[m.cis[m.PrimaryInputCount()+index-m.PrimaryOutputCount()]]
			drivers[index], err = g.HashMux(control, drivers[index], state)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, driver := range drivers {
		if stopped() {
			return nil, ErrCancelled
		}
		if _, err = g.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	g.registers, g.constraints = m.registers, m.constraints
	if stopped() {
		return nil, ErrCancelled
	}
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}
