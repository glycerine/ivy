package abc

// Reset normalization from src/aig/gia/giaDup.c:Gia_ManDupZeroUndc.

// GiaInitialNormalization reports original unknown registers in added-PI order.
// ResetRegister is the added selector register index, or -1 without unknowns.
// Initial contains all-zero target resets. PIs are original, unknown-value PIs,
// then caller-requested unused extras. Original register order is retained.
type GiaInitialNormalization struct {
	Graph            *GiaManager
	Initial          []LatchInit
	UnknownRegisters []int
	ResetRegister    int
}

// NormalizeInitialState encodes zero/one/unknown source resets using only zero
// target resets. Known ones invert state reads and corresponding RIs. Unknown
// state reads select an added PI at frame zero and the real RO thereafter,
// controlled by an added zero-initialized selector with constant-one RI.
// Later values of added PIs have no effect. Extra PIs are unused. Raw extended
// nodes, constraint roles and name/spec survive; source scratch is unchanged.
// This implements C's default nonsimplifying mode, with mux-control fanouts
// included when deciding whether an unknown RO needs a selector mux.
func (m *GiaManager) NormalizeInitialState(initial []LatchInit, extraInputs int, options GiaCofactorOptions) (GiaInitialNormalization, error) {
	fail := func(e error) (GiaInitialNormalization, error) { return GiaInitialNormalization{}, e }
	if m == nil || extraInputs < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if len(initial) != m.registers {
		return fail(ErrDimension)
	}
	if len(m.objects) > options.MaxMappingEntries {
		return fail(ErrLimit)
	}
	unknown := make([]int, 0)
	for i, value := range initial {
		if value > LatchInitUnknown {
			return fail(ErrDimension)
		}
		if value == LatchInitUnknown {
			unknown = append(unknown, i)
		}
	}
	base, err := checkedSize(1, 1, len(m.cis), options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	base, err = checkedSize(base, 1, len(unknown), options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	if _, err = checkedSize(base, 1, extraInputs, options.MaxObjects); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	refs, err := m.ReferenceCounts(options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, id := range m.cis[:m.PrimaryInputCount()] {
		if stopped() {
			return fail(ErrCancelled)
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	unknownInputs := make([]GiaLit, m.registers)
	for _, index := range unknown {
		if stopped() {
			return fail(ErrCancelled)
		}
		unknownInputs[index], err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	for i := 0; i < extraInputs; i++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		if _, err = g.AppendCI(); err != nil {
			return fail(err)
		}
	}
	for _, id := range m.cis[m.PrimaryInputCount():] {
		if stopped() {
			return fail(ErrCancelled)
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	resetRegister := -1
	reset := InvalidGiaLit
	if len(unknown) > 0 {
		resetRegister = m.registers
		reset, err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	for index, id := range m.cis[m.PrimaryInputCount():] {
		if stopped() {
			return fail(ErrCancelled)
		}
		if initial[index] == LatchInitOne {
			mapping[id] = mapping[id].Not()
		} else if initial[index] == LatchInitUnknown && refs[id] > 0 {
			low, e := g.AppendAnd(reset.Not(), unknownInputs[index])
			if e != nil {
				return fail(e)
			}
			high, e := g.AppendAnd(reset, mapping[id])
			if e != nil {
				return fail(e)
			}
			joined, e := g.AppendAnd(low.Not(), high.Not())
			if e != nil {
				return fail(e)
			}
			mapping[id] = joined.Not()
		}
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return fail(ErrCancelled)
		}
		mapping[i], err = appendGiaCopiedGate(m, g, GiaID(i), mapping)
		if err != nil {
			return fail(err)
		}
	}
	for index, id := range m.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return fail(e)
		}
		l, e := mappedGiaSignal(mapping, driver)
		if e != nil {
			return fail(e)
		}
		if index >= m.PrimaryOutputCount() {
			l = l.NotIf(initial[index-m.PrimaryOutputCount()] == LatchInitOne)
		}
		if _, err = g.AppendCO(l); err != nil {
			return fail(err)
		}
	}
	if resetRegister >= 0 {
		if _, err = g.AppendCO(GiaConst1); err != nil {
			return fail(err)
		}
	}
	g.registers = m.registers
	if resetRegister >= 0 {
		g.registers++
	}
	g.constraints = m.constraints
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaInitialNormalization{g, make([]LatchInit, g.registers), unknown, resetRegister}, nil
}
