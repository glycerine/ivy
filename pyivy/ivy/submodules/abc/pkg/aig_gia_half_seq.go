package abc

// Sequential state partition from src/aig/gia/giaDup.c:Gia_ManDupHalfSeq.

// GiaHalfSequence records target-CI to source-CI ordinals and target-register
// to source-register ordinals. Omitted state transitions become unconstrained
// primary inputs; this operation alone does not prove a property of the source.
type GiaHalfSequence struct {
	Graph             *GiaManager
	Inputs, Registers []int
}

// SplitRegisterHalf keeps every PO and the selected half of RIs. The first
// floor(R/2) registers form the first half, the remainder the second. Other
// ROs become PIs before retained ROs, matching C, including odd/zero R counts.
// Constraint roles clear; raw gate kinds and name/spec survive. Reachable gates
// copy in per-output DFS order; all CIs remain. Sources stay unchanged.
func (m *GiaManager) SplitRegisterHalf(second bool, options GiaCofactorOptions) (GiaHalfSequence, error) {
	fail := func(e error) (GiaHalfSequence, error) { return GiaHalfSequence{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	capObjects := options.MaxObjects
	if capObjects == 0 {
		capObjects = 1 << 20
	}
	g, e := NewGiaManagerWithLimit(capObjects)
	if e != nil {
		return fail(e)
	}
	g.Name, g.Spec = m.Name, m.Spec
	pi, half := m.PrimaryInputCount(), m.registers/2
	begin, end := 0, half
	if second {
		begin, end = half, m.registers
	}
	inputs := make([]int, 0, len(m.cis))
	registers := make([]int, 0, end-begin)
	for i := 0; i < pi; i++ {
		inputs = append(inputs, i)
	}
	for i := 0; i < m.registers; i++ {
		if i < begin || i >= end {
			inputs = append(inputs, pi+i)
		}
	}
	for i := begin; i < end; i++ {
		inputs = append(inputs, pi+i)
		registers = append(registers, i)
	}
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, i := range inputs {
		if stopped() {
			return fail(ErrCancelled)
		}
		mapping[m.cis[i]], e = g.AppendCI()
		if e != nil {
			return fail(e)
		}
	}
	roots := append([]GiaID(nil), m.cos[:m.PrimaryOutputCount()]...)
	roots = append(roots, m.cos[m.PrimaryOutputCount()+begin:m.PrimaryOutputCount()+end]...)
	for _, co := range roots {
		if stopped() {
			return fail(ErrCancelled)
		}
		nodes, err := m.CollectTransitiveFanins([]GiaID{co}, limit)
		if err != nil {
			return fail(err)
		}
		for _, id := range nodes {
			if mapping[id].Valid() {
				continue
			}
			if stopped() {
				return fail(ErrCancelled)
			}
			mapping[id], e = appendGiaCopiedGate(m, g, id, mapping)
			if e != nil {
				return fail(e)
			}
		}
		driver, err := m.objects[co].Fanin0(co)
		if err != nil {
			return fail(err)
		}
		signal, err := mappedGiaSignal(mapping, driver)
		if err != nil {
			return fail(err)
		}
		mapping[co], e = g.AppendCO(signal)
		if e != nil {
			return fail(e)
		}
	}
	g.registers = end - begin
	if e = g.Check(); e != nil {
		return fail(e)
	}
	return GiaHalfSequence{g, inputs, registers}, nil
}
