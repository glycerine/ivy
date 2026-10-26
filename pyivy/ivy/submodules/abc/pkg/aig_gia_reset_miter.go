package abc

// Reset-aware two-copy comparison from src/aig/gia/giaDup.c:Gia_ManMiter2.

// GiaResetMiter describes two normalized zero-initialized state banks.
// FirstInputs/SecondInputs map normalized PI indices to target PI indices:
// original PIs first, followed by UnknownRegisters' initial-value inputs.
// RegistersPerSide includes the optional selector register. Original source
// constraint POs become comparisons, not assumptions, exactly as in C.
type GiaResetMiter struct {
	Graph                                       *GiaManager
	Initial                                     []LatchInit
	FirstInputs, SecondInputs, UnknownRegisters []int
	RegistersPerSide                            int
}

// ResetComparisonMiter compares two source copies. specification has one byte
// per source CI: PI bytes are x (shared) or X (independent); register bytes
// are 0/1 (known reset), x (shared unknown reset), X (independent unknown reset).
// Normalized initial-value PIs affect frame zero only. Comparison POs are XORs
// of all source POs, followed by first then second RI banks. All target resets
// are zero. This constructs a query, not an equivalence verdict. Name/spec and
// source scratch survive; extended gates are lowered and unused logic cleaned.
func (m *GiaManager) ResetComparisonMiter(specification string, options GiaCofactorOptions) (GiaResetMiter, error) {
	fail := func(e error) (GiaResetMiter, error) { return GiaResetMiter{}, e }
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
	if len(specification) != len(m.cis) {
		return fail(ErrDimension)
	}
	initial := make([]LatchInit, m.registers)
	shared := make([]bool, 0, len(m.cis))
	for i := 0; i < m.PrimaryInputCount(); i++ {
		ch := specification[i]
		if ch != 'x' && ch != 'X' {
			return fail(ErrDimension)
		}
		shared = append(shared, ch == 'x')
	}
	for i := range initial {
		switch specification[m.PrimaryInputCount()+i] {
		case '0':
			initial[i] = LatchInitZero
		case '1':
			initial[i] = LatchInitOne
		case 'x', 'X':
			initial[i] = LatchInitUnknown
			shared = append(shared, specification[m.PrimaryInputCount()+i] == 'x')
		default:
			return fail(ErrDimension)
		}
	}
	normalized, err := m.NormalizeInitialState(initial, 0, options)
	if err != nil {
		return fail(err)
	}
	source := normalized.Graph
	if _, err = checkedSize(len(source.objects), 2, len(source.cos), options.MaxMappingEntries); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	g.Name, g.Spec = m.Name, m.Spec
	firstInputs, secondInputs := make([]int, len(shared)), make([]int, len(shared))
	inputLits := make([]GiaLit, len(shared))
	for i := range shared {
		if stopped() {
			return fail(ErrCancelled)
		}
		firstInputs[i] = g.InputCount()
		inputLits[i], err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	secondLits := make([]GiaLit, len(shared))
	for i, isShared := range shared {
		if stopped() {
			return fail(ErrCancelled)
		}
		if isShared {
			secondInputs[i] = firstInputs[i]
			secondLits[i] = inputLits[i]
		} else {
			secondInputs[i] = g.InputCount()
			secondLits[i], err = g.AppendCI()
			if err != nil {
				return fail(err)
			}
		}
	}
	if err = g.HashStart(); err != nil {
		return fail(err)
	}
	values := [2][]GiaLit{}
	for side := 0; side < 2; side++ {
		mapping := make([]GiaLit, len(source.objects))
		for i := range mapping {
			mapping[i] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
		pis := inputLits
		if side == 1 {
			pis = secondLits
		}
		for i, l := range pis {
			mapping[source.cis[i]] = l
		}
		for _, id := range source.cis[source.PrimaryInputCount():] {
			if stopped() {
				return fail(ErrCancelled)
			}
			mapping[id], err = g.AppendCI()
			if err != nil {
				return fail(err)
			}
		}
		for i, obj := range source.objects {
			if !obj.IsAnd() {
				continue
			}
			if stopped() {
				return fail(ErrCancelled)
			}
			mapping[i], err = hashGiaCopiedGate(source, g, GiaID(i), mapping)
			if err != nil {
				return fail(err)
			}
		}
		values[side] = make([]GiaLit, len(source.cos))
		for i, id := range source.cos {
			driver, e := source.objects[id].Fanin0(id)
			if e != nil {
				return fail(e)
			}
			values[side][i], err = mappedGiaSignal(mapping, driver)
			if err != nil {
				return fail(err)
			}
		}
	}
	for i := 0; i < source.PrimaryOutputCount(); i++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		l, e := g.HashXor(values[0][i], values[1][i])
		if e != nil {
			return fail(e)
		}
		if _, err = g.AppendCO(l); err != nil {
			return fail(err)
		}
	}
	for side := 0; side < 2; side++ {
		for _, l := range values[side][source.PrimaryOutputCount():] {
			if stopped() {
				return fail(ErrCancelled)
			}
			if _, err = g.AppendCO(l); err != nil {
				return fail(err)
			}
		}
	}
	g.HashStop()
	g.registers = 2 * source.registers
	if stopped() {
		return fail(ErrCancelled)
	}
	cleaned, _, err := g.Cleanup()
	if err != nil {
		return fail(err)
	}
	return GiaResetMiter{cleaned, make([]LatchInit, cleaned.registers), firstInputs, secondInputs, normalized.UnknownRegisters, source.registers}, nil
}
