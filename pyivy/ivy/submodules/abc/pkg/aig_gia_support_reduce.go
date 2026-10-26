package abc

// Support-normalized strashing from src/aig/gia/giaDup.c:Gia_ManDupStrashReduce.

// GiaSupportReduction records per-source-output input projection and its
// reduced output index. Each cone uses its own sorted source CI support as
// the shared target CI positions. This is not whole-network equivalence under
// a single input assignment; evaluate each output with its own projection.
type GiaSupportReduction struct {
	Graph    *GiaManager
	Supports [][]int
	Outputs  []int
}

// ReduceBySupport hashes each combinational output cone with support inputs
// renamed to 0..n-1. mergeOutputs deduplicates equal phased target literals in
// first-encounter order. Unlike C's size-one shortcut, actual logic is copied
// even for single-input or constant structural cones, preserving buffer phases
// and extended gate semantics. Constraints require an explicit caller policy
// and are rejected. At least one target CI is retained, matching C.
func (m *GiaManager) ReduceBySupport(mergeOutputs bool, options GiaCofactorOptions) (GiaSupportReduction, error) {
	fail := func(e error) (GiaSupportReduction, error) { return GiaSupportReduction{}, e }
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
	if m.registers != 0 || m.constraints != 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	supports, err := m.OutputSupports(options.MaxMappingEntries, options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	maximum := 1
	for _, s := range supports {
		maximum = max(maximum, len(s))
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	g.Name, g.Spec = m.Name, m.Spec
	inputs := make([]GiaLit, maximum)
	for i := range inputs {
		if stopped() {
			return fail(ErrCancelled)
		}
		inputs[i], err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	if err = g.HashStart(); err != nil {
		return fail(err)
	}
	mapping := make([]GiaLit, len(m.objects))
	drivers := make([]GiaLit, len(m.cos))
	for output, co := range m.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		for i := range mapping {
			mapping[i] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
		for position, ci := range supports[output] {
			mapping[m.cis[ci]] = inputs[position]
		}
		nodes, e := m.CollectTransitiveFanins([]GiaID{co}, options.MaxMappingEntries)
		if e != nil {
			return fail(e)
		}
		for _, id := range nodes {
			if stopped() {
				return fail(ErrCancelled)
			}
			mapping[id], err = hashGiaCopiedGate(m, g, id, mapping)
			if err != nil {
				return fail(err)
			}
		}
		driver, e := m.objects[co].Fanin0(co)
		if e != nil {
			return fail(e)
		}
		drivers[output], err = mappedGiaSignal(mapping, driver)
		if err != nil {
			return fail(err)
		}
	}
	g.HashStop()
	outputs := make([]int, len(drivers))
	lookup := make(map[GiaLit]int)
	for i, driver := range drivers {
		if stopped() {
			return fail(ErrCancelled)
		}
		if index, exists := lookup[driver]; mergeOutputs && exists {
			outputs[i] = index
			continue
		}
		outputs[i] = g.OutputCount()
		lookup[driver] = outputs[i]
		if _, err = g.AppendCO(driver); err != nil {
			return fail(err)
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	cleaned, _, err := g.Cleanup()
	if err != nil {
		return fail(err)
	}
	return GiaSupportReduction{cleaned, supports, outputs}, nil
}
