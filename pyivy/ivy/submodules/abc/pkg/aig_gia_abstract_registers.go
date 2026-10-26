package abc

// Register abstraction from src/aig/gia/giaDup.c:Gia_ManDupAbs.

// AbstractRegisters externalizes the listed source registers as independent
// pseudo-PIs, in requested order. Their RIs are dropped. Other registers remain
// in original order. Only PO and retained-RI cones are copied, in DFS order.
// GiaFlopClass's source CI/CO/register projections describe the result and reset
// projection. This overapproximates state behavior; it is not an equivalence
// transformation. Constraints/name/spec and extended gate semantics survive.
func (m *GiaManager) AbstractRegisters(external []int, options GiaCofactorOptions) (GiaFlopClass, error) {
	fail := func(e error) (GiaFlopClass, error) { return GiaFlopClass{}, e }
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
	if len(external) > m.registers {
		return fail(ErrDimension)
	}
	if options.Stop != nil && options.Stop() {
		return fail(ErrCancelled)
	}
	selected := make([]bool, m.registers)
	for _, index := range external {
		if index < 0 || index >= m.registers || selected[index] {
			return fail(ErrDimension)
		}
		selected[index] = true
	}
	inputs := make([]int, 0, len(m.cis))
	outputs := make([]int, 0, len(m.cos))
	registers := make([]int, 0, m.registers-len(external))
	for i := 0; i < m.PrimaryInputCount(); i++ {
		inputs = append(inputs, i)
	}
	for _, index := range external {
		inputs = append(inputs, m.PrimaryInputCount()+index)
	}
	for i := 0; i < m.PrimaryOutputCount(); i++ {
		outputs = append(outputs, i)
	}
	for index := range selected {
		if selected[index] {
			continue
		}
		inputs = append(inputs, m.PrimaryInputCount()+index)
		outputs = append(outputs, m.PrimaryOutputCount()+index)
		registers = append(registers, index)
	}
	cis := make([]GiaID, len(inputs))
	for i, index := range inputs {
		cis[i] = m.cis[index]
	}
	cos := make([]GiaID, len(outputs))
	for i, index := range outputs {
		cos[i] = m.cos[index]
	}
	gates, err := m.CollectTransitiveFanins(cos, limit)
	if err != nil {
		return fail(err)
	}
	g, _, err := m.DuplicateFromLists(cis, gates, cos, len(registers), options)
	if err != nil {
		return fail(err)
	}
	g.constraints = m.constraints
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaFlopClass{g, inputs, outputs, registers}, nil
}
