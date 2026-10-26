package abc

// Structural register equivalence from src/aig/gia/giaScl.c:Gia_ManReduceEquiv.

// GiaRegisterReduction reports source-CI substitutions (InvalidGiaLit means
// identity). Register slots are not removed; SequentialCleanup can remove
// newly unobservable slots. Graph always has independent ownership.
type GiaRegisterReduction struct {
	Graph                 *GiaManager
	Replacements          []GiaLit
	Constants, Equivalent int
}

// ReduceZeroInitializedRegisterInputs replaces reads of constant-zero-next
// registers by zero and reads of registers with identical phased RI drivers
// by the earliest such RO. ALL registers must initially be zero. Arbitrary,
// one or independent unknown initial values do not satisfy this contract.
// Equality is structural, not SAT-based. Complementary drivers are not merged.
// Constraint roles, CI/CO/register order and name/spec survive; changed graphs
// hash/lower extended nodes. Unchanged graphs are cloned, unlike C's alias.
func (m *GiaManager) ReduceZeroInitializedRegisterInputs(options GiaCofactorOptions) (GiaRegisterReduction, error) {
	fail := func(e error) (GiaRegisterReduction, error) { return GiaRegisterReduction{}, e }
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
	if len(m.objects) > options.MaxMappingEntries {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	replacements := make([]GiaLit, len(m.cis))
	for i := range replacements {
		replacements[i] = InvalidGiaLit
	}
	earliest := make(map[GiaLit]GiaLit)
	constants, equivalent := 0, 0
	for index, id := range m.cos[m.PrimaryOutputCount():] {
		if stopped() {
			return fail(ErrCancelled)
		}
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return fail(e)
		}
		ci := m.PrimaryInputCount() + index
		if driver == GiaConst0 {
			replacements[ci] = GiaConst0
			constants++
		} else if representative, exists := earliest[driver]; exists {
			replacements[ci] = representative
			equivalent++
		} else {
			earliest[driver] = GiaLit(m.cis[ci]) << 1
		}
	}
	var g *GiaManager
	if constants+equivalent == 0 {
		if len(m.objects) > options.MaxObjects {
			return fail(ErrLimit)
		}
		g = m.Clone()
	} else {
		var err error
		g, err = m.RemapInputs(replacements, nil, options)
		if err != nil {
			return fail(err)
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return GiaRegisterReduction{g, replacements, constants, equivalent}, nil
}
