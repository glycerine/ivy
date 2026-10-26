package abc

// Structural sequential sweeping from src/aig/gia/giaScl.c:Gia_ManSeqStructSweep.

// GiaStructuralSweepOptions selects C's constant/equivalent-register passes.
// MaxIterations bounds the cleanup/equivalence loop; zero uses source registers
// plus one. Ternary-history budgets apply to the optional constant pass.
type GiaStructuralSweepOptions struct {
	GiaTernaryHistoryOptions
	ReduceConstants, MergeEquivalent bool
	MaxIterations                    int
}

// SweepZeroInitializedRegisters performs sequential cleanup, optional ternary
// reduction, then optional repeated cleanup/structural-equivalence reduction.
// ALL registers must initially be zero. Returned register indices identify
// surviving source state/reset coordinates; deleted state may require additional
// reconstruction for witness lifting. All PIs/POs and constraint roles survive.
// With MergeEquivalent false, the constant pass retains substituted slots just
// as C does. Combinational inputs take C's ordinary cleanup path. Source graphs
// remain unchanged and no partial result is returned on any budget failure.
func (m *GiaManager) SweepZeroInitializedRegisters(options GiaStructuralSweepOptions) (GiaConeExtraction, error) {
	fail := func(e error) (GiaConeExtraction, error) { return GiaConeExtraction{}, e }
	if m == nil || options.MaxIterations < 0 || options.MaxStates < 0 || options.MaxStateEntries < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
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
	if options.MaxIterations == 0 {
		options.MaxIterations = m.registers + 1
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	primary := make([]int, m.PrimaryInputCount())
	for i := range primary {
		primary[i] = i
	}
	if m.registers == 0 {
		g, _, err := m.Cleanup()
		if err != nil {
			return fail(err)
		}
		if g.ObjectCount() > options.MaxObjects {
			return fail(ErrLimit)
		}
		g.objectLimit = options.MaxObjects
		if stopped() {
			return fail(ErrCancelled)
		}
		return GiaConeExtraction{Graph: g, PrimaryInputs: primary, Registers: []int{}}, nil
	}
	initial, err := m.SequentialCleanup(options.GiaCofactorOptions)
	if err != nil {
		return fail(err)
	}
	current := initial.Graph
	registers := initial.Registers
	if options.ReduceConstants && current.registers > 0 {
		reduced, e := current.ReduceTernaryZeroInitializedRegisters(options.GiaTernaryHistoryOptions)
		if e != nil {
			return fail(e)
		}
		current = reduced.Graph
	}
	if options.MergeEquivalent {
		for iteration := 0; ; iteration++ {
			if iteration >= options.MaxIterations {
				return fail(ErrLimit)
			}
			if stopped() {
				return fail(ErrCancelled)
			}
			cleaned, e := current.SequentialCleanup(options.GiaCofactorOptions)
			if e != nil {
				return fail(e)
			}
			projected := make([]int, len(cleaned.Registers))
			for i, index := range cleaned.Registers {
				projected[i] = registers[index]
			}
			registers = projected
			current = cleaned.Graph
			if current.registers == 0 {
				break
			}
			reduced, e := current.ReduceZeroInitializedRegisterInputs(options.GiaCofactorOptions)
			if e != nil {
				return fail(e)
			}
			current = reduced.Graph
			if reduced.Constants+reduced.Equivalent == 0 {
				break
			}
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return GiaConeExtraction{Graph: current, PrimaryInputs: primary, Registers: registers}, nil
}
