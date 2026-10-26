package abc

// Register-class partitioning from src/aig/gia/giaDup.c:Gia_ManDupFlopClass.

// GiaFlopClass gives source terminal indices in target CI/CO order, plus
// source register indices retained as state (for projecting reset values).
// Externalized registers become independent PIs and ordinary POs.
type GiaFlopClass struct {
	Graph                      *GiaManager
	Inputs, Outputs, Registers []int
}

// ExtractFlopClass retains registers whose explicit class equals selected.
// CI order is original PIs, other-class ROs, selected ROs. CO order is
// original ordinary POs, other-class RIs, original constraints, selected RIs.
// Moving constraints preserves their suffix role, which C drops. Gate kinds,
// dangling logic, name/spec and source scratch are preserved. This abstracts
// externalized state; callers must interpret the returned projections.
func (m *GiaManager) ExtractFlopClass(classes []int, selected int, options GiaCofactorOptions) (GiaFlopClass, error) {
	fail := func(e error) (GiaFlopClass, error) { return GiaFlopClass{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if len(classes) != m.registers {
		return fail(ErrDimension)
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return fail(ErrLimit)
	}
	inputs := make([]int, 0, len(m.cis))
	outputs := make([]int, 0, len(m.cos))
	registers := make([]int, 0, m.registers)
	for i := 0; i < m.PrimaryInputCount(); i++ {
		inputs = append(inputs, i)
	}
	for _, match := range []bool{false, true} {
		for i, class := range classes {
			if (class == selected) == match {
				inputs = append(inputs, m.PrimaryInputCount()+i)
				if match {
					registers = append(registers, i)
				}
			}
		}
	}
	ordinary := m.PrimaryOutputCount() - m.constraints
	for i := 0; i < ordinary; i++ {
		outputs = append(outputs, i)
	}
	for i, class := range classes {
		if class != selected {
			outputs = append(outputs, m.PrimaryOutputCount()+i)
		}
	}
	for i := ordinary; i < m.PrimaryOutputCount(); i++ {
		outputs = append(outputs, i)
	}
	for i, class := range classes {
		if class == selected {
			outputs = append(outputs, m.PrimaryOutputCount()+i)
		}
	}
	cis := make([]GiaID, len(inputs))
	for i, index := range inputs {
		cis[i] = m.cis[index]
	}
	cos := make([]GiaID, len(outputs))
	for i, index := range outputs {
		cos[i] = m.cos[index]
	}
	gates := make([]GiaID, 0, m.AndCount())
	for i, obj := range m.objects {
		if obj.IsAnd() {
			gates = append(gates, GiaID(i))
		}
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
