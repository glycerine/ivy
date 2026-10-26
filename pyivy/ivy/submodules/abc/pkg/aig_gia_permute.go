package abc

// Terminal permutations from src/aig/gia/giaDup.c:Gia_ManDupPerm/PermFlop.

// PermutePrimaryInputs creates target PI i from source PI order[i]. Register
// order stays unchanged. order must be a bijection over all source PIs.
func (m *GiaManager) PermutePrimaryInputs(order []int, options GiaCofactorOptions) (*GiaManager, error) {
	return m.permuteTerminals(order, false, options)
}

// PermuteRegisters creates target register i from source register order[i],
// reordering both ROs and RIs. Reset lists must use the same new-to-old order.
// Both permutation APIs preserve constraints, raw gate kinds and name/spec.
func (m *GiaManager) PermuteRegisters(order []int, options GiaCofactorOptions) (*GiaManager, error) {
	return m.permuteTerminals(order, true, options)
}

func (m *GiaManager) permuteTerminals(order []int, registers bool, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	size := m.PrimaryInputCount()
	if registers {
		size = m.registers
	}
	if len(order) != size {
		return nil, ErrDimension
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return nil, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	seen := make([]bool, size)
	for _, index := range order {
		if index < 0 || index >= size || seen[index] {
			return nil, ErrDimension
		}
		seen[index] = true
	}
	cis := make([]GiaID, len(m.cis))
	copy(cis, m.cis)
	cos := make([]GiaID, len(m.cos))
	copy(cos, m.cos)
	for i, index := range order {
		if registers {
			cis[m.PrimaryInputCount()+i] = m.cis[m.PrimaryInputCount()+index]
			cos[m.PrimaryOutputCount()+i] = m.cos[m.PrimaryOutputCount()+index]
		} else {
			cis[i] = m.cis[index]
		}
	}
	gates := make([]GiaID, 0, m.AndCount())
	for i, obj := range m.objects {
		if obj.IsAnd() {
			gates = append(gates, GiaID(i))
		}
	}
	g, _, err := m.DuplicateFromLists(cis, gates, cos, m.registers, options)
	if err != nil {
		return nil, err
	}
	g.constraints = m.constraints
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
