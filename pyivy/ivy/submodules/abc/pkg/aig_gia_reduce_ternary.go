package abc

// History-derived register substitutions from src/aig/gia/giaTsim.c:
// Gia_ManTerTranspose, Gia_ManFindEqualFlop, Gia_ManTerCreateMap/Gia_ManReduceConst.

// ReduceTernaryZeroInitializedRegisters replaces all-history-zero ROs by zero
// and identical fully known register histories by their earliest source RO.
// ALL registers must initially be zero. Histories containing unknown values
// cannot justify equality; complementary histories are not merged. Simulation
// ignores constraints, so its substitutions do not depend on assumptions.
// Register slots remain for subsequent SequentialCleanup. Source CI substitutions
// and counters are returned; constraint roles/name/spec survive. History limits
// or cancellation return no result, never a speculative partial reduction.
func (m *GiaManager) ReduceTernaryZeroInitializedRegisters(options GiaTernaryHistoryOptions) (GiaRegisterReduction, error) {
	fail := func(e error) (GiaRegisterReduction, error) { return GiaRegisterReduction{}, e }
	history, err := m.SimulateTernaryHistory(options)
	if err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	replacements := make([]GiaLit, len(m.cis))
	for i := range replacements {
		replacements[i] = InvalidGiaLit
	}
	earliest := make(map[string]GiaLit)
	constants, equivalent := 0, 0
	for index := 0; index < m.registers; index++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		ci := m.PrimaryInputCount() + index
		if history.ZeroCounts[index] == len(history.States) {
			replacements[ci] = GiaConst0
			constants++
			continue
		}
		if history.UnknownCounts[index] != 0 {
			continue
		}
		column := make([]byte, len(history.States))
		for frame, state := range history.States {
			column[frame] = byte(state[index])
		}
		key := string(column)
		if representative, exists := earliest[key]; exists {
			replacements[ci] = representative
			equivalent++
		} else {
			earliest[key] = GiaLit(m.cis[ci]) << 1
		}
	}
	g, err := m.RemapInputs(replacements, nil, options.GiaCofactorOptions)
	if err != nil {
		return fail(err)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return GiaRegisterReduction{g, replacements, constants, equivalent}, nil
}
