package abc

// General terminal trimming from src/aig/gia/giaDup.c:Gia_ManDupTrimmed
// and Gia_ManPoIsToRemove.

// GiaTrimOptions configures terminal trimming. RemoveConstant is 0 or 1 to
// remove that constant phase, or -1 for both; it matters only for non-dual
// output trimming. DualOutputs+TrimOutputs uses zero-initialized pair selection
// and sweeping, so ALL registers must initially be zero in that mode.
type GiaTrimOptions struct {
	GiaStructuralSweepOptions
	TrimInputs, TrimOutputs, DualOutputs bool
	RemoveConstant                       int
}

// GiaTrimming exposes source CI/CO/register projections. -1 terminal entries
// are dummy PI/zero PO placeholders. Register indices project resets/state;
// deleted-state witness reconstruction is a separate responsibility.
type GiaTrimming struct {
	Graph                      *GiaManager
	Inputs, Outputs, Registers []int
}

// Trim removes unreferenced PIs and selected constant POs when requested.
// PIs referenced only by dangling logic are retained; all raw gates are copied
// before optional dual-mode sweeping. Dummy PI/PO fallbacks match C. Constraint
// roles are cleared, as C may remove/reinterpret those POs. Name/spec survive.
// Dual mode excludes both constant-zero and constant-one pair XORs after
// sweeping, matching C; it must not be interpreted as an equivalence verdict.
func (m *GiaManager) Trim(options GiaTrimOptions) (GiaTrimming, error) {
	fail := func(e error) (GiaTrimming, error) { return GiaTrimming{}, e }
	if m == nil || options.RemoveConstant < -1 || options.RemoveConstant > 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxStates < 0 || options.MaxStateEntries < 0 || options.MaxIterations < 0 {
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
	dual := options.DualOutputs && options.TrimOutputs
	var pairs []int
	if dual {
		var e error
		pairs, e = m.NonconstantOutputPairIndices(options.GiaStructuralSweepOptions)
		if e != nil {
			return fail(e)
		}
	}
	refs, err := m.ReferenceCounts(options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	inputs, outputs, registers := make([]int, 0), make([]int, 0), make([]int, m.registers)
	for i, id := range m.cis[:m.PrimaryInputCount()] {
		if !options.TrimInputs || refs[id] > 0 {
			inputs = append(inputs, i)
		}
	}
	if len(inputs) == 0 {
		inputs = append(inputs, -1)
	}
	for i := m.PrimaryInputCount(); i < len(m.cis); i++ {
		inputs = append(inputs, i)
	}
	if dual {
		for _, index := range pairs {
			outputs = append(outputs, 2*index, 2*index+1)
		}
	} else {
		for i, id := range m.cos[:m.PrimaryOutputCount()] {
			driver, e := m.objects[id].Fanin0(id)
			if e != nil {
				return fail(e)
			}
			remove := options.TrimOutputs && driver.ID() == 0 && (options.RemoveConstant == -1 || driver.Complemented() == (options.RemoveConstant == 1))
			if !remove {
				outputs = append(outputs, i)
			}
		}
	}
	if len(outputs) == 0 {
		outputs = append(outputs, -1)
		if dual {
			outputs = append(outputs, -1)
		}
	}
	poCount := len(outputs)
	for i := m.PrimaryOutputCount(); i < len(m.cos); i++ {
		outputs = append(outputs, i)
	}
	for i := range registers {
		registers[i] = i
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
		driver := GiaConst0
		if index >= 0 {
			id := m.cos[index]
			l, e := m.objects[id].Fanin0(id)
			if e != nil {
				return fail(e)
			}
			driver, err = mappedGiaSignal(mapping, l)
			if err != nil {
				return fail(err)
			}
		}
		if _, err = g.AppendCO(driver); err != nil {
			return fail(err)
		}
	}
	g.registers = m.registers
	if dual {
		sweepOptions := options.GiaStructuralSweepOptions
		sweepOptions.ReduceConstants, sweepOptions.MergeEquivalent = true, true
		swept, e := g.SweepZeroInitializedRegisters(sweepOptions)
		if e != nil {
			return fail(e)
		}
		projectedInputs := make([]int, 0, len(swept.Graph.cis))
		for _, index := range swept.PrimaryInputs {
			projectedInputs = append(projectedInputs, inputs[index])
		}
		projectedOutputs := append([]int(nil), outputs[:poCount]...)
		for _, index := range swept.Registers {
			projectedInputs = append(projectedInputs, inputs[len(inputs)-m.registers+index])
			projectedOutputs = append(projectedOutputs, outputs[poCount+index])
		}
		inputs, outputs, registers = projectedInputs, projectedOutputs, swept.Registers
		g = swept.Graph
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaTrimming{g, inputs, outputs, registers}, nil
}
