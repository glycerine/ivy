package abc

// Sequential reachability cleanup from src/aig/gia/giaScl.c:
// Gia_ManSeqMarkUsed_rec, Gia_ManSeqMarkUsed and Gia_ManSeqCleanup.

// SequentialCleanup retains all PIs/POs and only registers/logic transitively
// reachable from any PO, following each reached RO to its RI until closure.
// Constraints are POs and therefore participate. Retained registers stay in
// source order, unlike ExtractOutputCones' discovery order. Returned register
// indices project source resets/state. Native mux controls and buffers are
// traversed explicitly; source scratch, raw node kinds and name/spec survive.
func (m *GiaManager) SequentialCleanup(options GiaCofactorOptions) (GiaConeExtraction, error) {
	fail := func(e error) (GiaConeExtraction, error) { return GiaConeExtraction{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
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
	outputs := make([]int, m.PrimaryOutputCount())
	for i := range outputs {
		outputs[i] = i
	}
	nodes, err := m.CollectSequential(outputs, options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	used := make([]bool, len(m.objects))
	for _, id := range nodes {
		used[id] = true
	}
	cis := make([]GiaID, 0, len(m.cis))
	cos := make([]GiaID, 0, len(m.cos))
	gates := make([]GiaID, 0)
	primary := make([]int, m.PrimaryInputCount())
	for i := range primary {
		primary[i] = i
		cis = append(cis, m.cis[i])
	}
	registers := make([]int, 0, m.registers)
	for index, id := range m.cis[m.PrimaryInputCount():] {
		if used[id] {
			registers = append(registers, index)
			cis = append(cis, id)
		}
	}
	for i, obj := range m.objects {
		if obj.IsAnd() && used[i] {
			gates = append(gates, GiaID(i))
		}
	}
	cos = append(cos, m.cos[:m.PrimaryOutputCount()]...)
	for _, index := range registers {
		cos = append(cos, m.cos[m.PrimaryOutputCount()+index])
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	g, _, err := m.DuplicateFromLists(cis, gates, cos, len(registers), options)
	if err != nil {
		return fail(err)
	}
	g.constraints = m.constraints
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaConeExtraction{Graph: g, PrimaryInputs: primary, Registers: registers}, nil
}
