package abc

// Ternary state history from src/aig/gia/giaTsim.c:Gia_ManTerSimulate,
// Gia_ManTerSimInfoInit/Transfer, Gia_ManTerStateCreate and Gia_ManTerRetire.

// GiaTernaryHistoryOptions bounds simulation from all-zero initial state.
// MaxStates defaults to 4096, MaxStateEntries to one million scalar values.
// History keys/storage are owned; lookup maps never determine iteration order.
type GiaTernaryHistoryOptions struct {
	GiaCofactorOptions
	MaxStates, MaxStateEntries int
}

// GiaTernaryHistory includes the initial and final repeated states, as in C.
// LoopStart identifies the earlier occurrence of the final state. Counts include
// both occurrences. Retirement broadens changing registers to unknown after
// rounds 301,311,..., before the next transition, matching C's default schedule.
// Inputs are unknown each round; no constraints restrict simulation.
type GiaTernaryHistory struct {
	States                    [][]GiaTernary
	LoopStart                 int
	ZeroCounts, UnknownCounts []int
}

// SimulateTernaryHistory runs bounded C-compatible convergence simulation.
// At least one register is required. Extended nodes use EvaluateTernary's
// conservative tables. Exhaustion/cancellation returns no partial history and
// must never be interpreted as convergence or evidence of constant registers.
func (m *GiaManager) SimulateTernaryHistory(options GiaTernaryHistoryOptions) (GiaTernaryHistory, error) {
	fail := func(e error) (GiaTernaryHistory, error) { return GiaTernaryHistory{}, e }
	if m == nil || options.MaxStates < 0 || options.MaxStateEntries < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if options.MaxStates == 0 {
		options.MaxStates = 4096
	}
	if options.MaxStateEntries == 0 {
		options.MaxStateEntries = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if m.registers == 0 {
		return fail(ErrDimension)
	}
	if len(m.objects) > options.MaxObjects || len(m.objects) > options.MaxMappingEntries {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	result := GiaTernaryHistory{LoopStart: -1, ZeroCounts: make([]int, m.registers), UnknownCounts: make([]int, m.registers)}
	seen := make(map[string]int)
	save := func(state []GiaTernary) (string, error) {
		if len(result.States) >= options.MaxStates {
			return "", ErrLimit
		}
		if _, e := checkedSize(0, m.registers, len(result.States)+1, options.MaxStateEntries); e != nil {
			return "", e
		}
		owned := append([]GiaTernary(nil), state...)
		key := make([]byte, len(state))
		for i, v := range state {
			key[i] = byte(v)
			if v == GiaTernaryZero {
				result.ZeroCounts[i]++
			}
			if v == GiaTernaryUnknown {
				result.UnknownCounts[i]++
			}
		}
		result.States = append(result.States, owned)
		return string(key), nil
	}
	state := make([]GiaTernary, m.registers)
	for i := range state {
		state[i] = GiaTernaryZero
	}
	key, err := save(state)
	if err != nil {
		return fail(err)
	}
	seen[key] = 0
	inputs := make([]GiaTernary, len(m.cis))
	for i := 0; i < m.PrimaryInputCount(); i++ {
		inputs[i] = GiaTernaryUnknown
	}
	for iteration := 0; ; iteration++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		copy(inputs[m.PrimaryInputCount():], state)
		outputs, e := m.EvaluateTernary(inputs, options.GiaCofactorOptions)
		if e != nil {
			return fail(e)
		}
		state = append([]GiaTernary(nil), outputs[m.PrimaryOutputCount():]...)
		key, e = save(state)
		if e != nil {
			return fail(e)
		}
		if earlier, exists := seen[key]; exists {
			result.LoopStart = earlier
			return result, nil
		}
		seen[key] = len(result.States) - 1
		if iteration >= 300 && iteration%10 == 0 {
			previous := result.States[len(result.States)-2]
			for i, v := range state {
				if v != previous[i] {
					state[i] = GiaTernaryUnknown
				}
			}
		}
	}
}
