package abc

// Support counting and compression from src/proof/pdr/pdrInv.c:
// Pdr_InvCounts, Pdr_InvMap, Pdr_InvUsedFlopNum and Pdr_InvPrintStr.

// RegisterCounts counts literal occurrences in invariant cubes, including
// repeated cubes. maxRegisters zero selects one million registers.
func (invariant *PdrInvariant) RegisterCounts(registers, maxRegisters int) ([]uint64, error) {
	if invariant == nil || registers < 0 || maxRegisters < 0 {
		return nil, ErrDimension
	}
	if maxRegisters == 0 {
		maxRegisters = 1 << 20
	}
	if registers > maxRegisters {
		return nil, ErrLimit
	}
	for _, cube := range invariant.clauses {
		if err := cube.Check(); err != nil {
			return nil, err
		}
		for _, l := range cube.state {
			if uint64(l.Index()) >= uint64(registers) {
				return nil, ErrDimension
			}
		}
	}
	counts := make([]uint64, registers)
	for _, cube := range invariant.clauses {
		for _, l := range cube.state {
			counts[l.Index()]++
		}
	}
	return counts, nil
}

// MarshalSupportPLA omits unused registers and returns the original register
// index for each retained column, in increasing order. Retain this mapping
// with the PLA: columns no longer correspond to the full model's registers.
// Output one denotes an excluded state. Export is not proof verification.
func (invariant *PdrInvariant) MarshalSupportPLA(registers, maxRegisters, maxBytes int) ([]byte, []int, error) {
	if maxBytes < 0 {
		return nil, nil, ErrDimension
	}
	counts, err := invariant.RegisterCounts(registers, maxRegisters)
	if err != nil {
		return nil, nil, err
	}
	columns := make([]int, 0)
	mapping := make([]int, len(counts))
	for reg, count := range counts {
		mapping[reg] = -1
		if count != 0 {
			mapping[reg] = len(columns)
			columns = append(columns, reg)
		}
	}
	if maxBytes == 0 {
		maxBytes = 64 << 20
	}
	if _, err := checkedSize(0, len(invariant.clauses), len(columns)+3, maxBytes); err != nil {
		return nil, nil, err
	}
	cubes := make([]*PdrCube, len(invariant.clauses))
	for i, cube := range invariant.clauses {
		literals := make([]PdrLit, len(cube.state))
		for j, l := range cube.state {
			index := mapping[l.Index()]
			if uint64(index) > uint64(1<<30-1) {
				return nil, nil, ErrLimit
			}
			literals[j] = PdrLit(2 * index)
			if l.Negative() {
				literals[j] = literals[j].Not()
			}
		}
		cubes[i], err = NewPdrCube(literals, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	compressed := &PdrInvariant{clauses: cubes}
	data, err := compressed.MarshalPLA(len(columns), maxBytes)
	if err != nil {
		return nil, nil, err
	}
	return data, columns, nil
}
