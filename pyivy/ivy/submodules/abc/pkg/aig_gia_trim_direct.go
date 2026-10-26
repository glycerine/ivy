package abc

// Direct-output trimming from src/aig/gia/giaDup.c:Gia_ManDupTrimmed2.

// GiaTerminalTrimming maps target terminal indices to source indices. -1
// denotes an unused dummy PI or constant-zero dummy PO. Register order stays
// unchanged. Removed output properties/constraints are not preserved.
type GiaTerminalTrimming struct {
	Graph           *GiaManager
	Inputs, Outputs []int
}

// TrimDirectInputOutputs removes POs driven by constants or original PIs,
// ignoring phase, and PIs with no gate/RI references. Every RO/RI and raw gate
// survives, even dangling logic. At least one PI and one PO exist afterward;
// dummy terminals are reported as -1. Constraint roles are cleared as in C.
// Name/spec and extended gate kinds survive without source scratch mutation.
func (m *GiaManager) TrimDirectInputOutputs(options GiaCofactorOptions) (GiaTerminalTrimming, error) {
	fail := func(e error) (GiaTerminalTrimming, error) { return GiaTerminalTrimming{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	refs, err := m.ReferenceCounts(options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	for _, id := range m.cos[:m.PrimaryOutputCount()] {
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return fail(e)
		}
		refs[driver.ID()]--
	}
	inputs, outputs := make([]int, 0), make([]int, 0)
	for i, id := range m.cis[:m.PrimaryInputCount()] {
		if refs[id] > 0 {
			inputs = append(inputs, i)
		}
	}
	if len(inputs) == 0 {
		inputs = append(inputs, -1)
	}
	for i := m.PrimaryInputCount(); i < len(m.cis); i++ {
		inputs = append(inputs, i)
	}
	for i, id := range m.cos[:m.PrimaryOutputCount()] {
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return fail(e)
		}
		obj := m.objects[driver.ID()]
		isPI := false
		if obj.IsCI() {
			ordinal, e := obj.TerminalIndex()
			if e != nil {
				return fail(e)
			}
			isPI = ordinal < m.PrimaryInputCount()
		}
		if driver.ID() != 0 && !isPI {
			outputs = append(outputs, i)
		}
	}
	if len(outputs) == 0 {
		outputs = append(outputs, -1)
	}
	for i := m.PrimaryOutputCount(); i < len(m.cos); i++ {
		outputs = append(outputs, i)
	}
	count, err := checkedSize(1, 1, len(inputs), options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	count, err = checkedSize(count, 1, m.AndCount(), options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	if _, err = checkedSize(count, 1, len(outputs), options.MaxObjects); err != nil {
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
	for _, index := range inputs {
		if stopped() {
			return fail(ErrCancelled)
		}
		l, e := g.AppendCI()
		if e != nil {
			return fail(e)
		}
		if index >= 0 {
			mapping[m.cis[index]] = l
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
	for _, index := range outputs {
		if stopped() {
			return fail(ErrCancelled)
		}
		l := GiaConst0
		if index >= 0 {
			id := m.cos[index]
			driver, e := m.objects[id].Fanin0(id)
			if e != nil {
				return fail(e)
			}
			l, err = mappedGiaSignal(mapping, driver)
			if err != nil {
				return fail(err)
			}
		}
		if _, err = g.AppendCO(l); err != nil {
			return fail(err)
		}
	}
	g.registers = m.registers
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaTerminalTrimming{g, inputs, outputs}, nil
}
