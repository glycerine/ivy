package abc

// Paired-output miters and selection from src/aig/gia/giaDup.c:
// Gia_ManTransformMiter/TransformMiter2 and Gia_ManDupTrimmedNonZero.

// TransformOutputPairs XORs POs into comparison outputs, retaining all RIs.
// With split false pairs are (0,1),(2,3),...; with split true the first and
// second PO halves pair by index. PO count must be even. All POs participate,
// including source constraints, whose roles are cleared as in C. CI/register
// order and name/spec survive; extended nodes are lowered and cleaned up.
func (m *GiaManager) TransformOutputPairs(split bool, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	if m.PrimaryOutputCount()%2 != 0 {
		return nil, ErrDimension
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[i], err = hashGiaCopiedGate(m, g, GiaID(i), mapping)
		if err != nil {
			return nil, err
		}
	}
	copyDriver := func(index int) (GiaLit, error) {
		id := m.cos[index]
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return InvalidGiaLit, e
		}
		return mappedGiaSignal(mapping, driver)
	}
	pairs := m.PrimaryOutputCount() / 2
	for i := 0; i < pairs; i++ {
		if stopped() {
			return nil, ErrCancelled
		}
		first, second := 2*i, 2*i+1
		if split {
			first, second = i, pairs+i
		}
		a, e := copyDriver(first)
		if e != nil {
			return nil, e
		}
		b, e := copyDriver(second)
		if e != nil {
			return nil, e
		}
		driver, e := g.HashXor(a, b)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	for index := m.PrimaryOutputCount(); index < len(m.cos); index++ {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, e := copyDriver(index)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	g.registers = m.registers
	if stopped() {
		return nil, ErrCancelled
	}
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// NonconstantOutputPairIndices returns adjacent PO pair indices whose XOR
// driver remains nonconstant after zero-initialized structural sweeping. ALL
// registers must initially be zero. Both constant-zero AND constant-one XORs
// are excluded, matching C despite its "NonZero" name. This is a structural
// trimming selector, not a list of inequivalent/equivalent output pairs.
func (m *GiaManager) NonconstantOutputPairIndices(options GiaStructuralSweepOptions) ([]int, error) {
	graph, err := m.TransformOutputPairs(false, options.GiaCofactorOptions)
	if err != nil {
		return nil, err
	}
	options.ReduceConstants, options.MergeEquivalent = true, true
	swept, err := graph.SweepZeroInitializedRegisters(options)
	if err != nil {
		return nil, err
	}
	result := make([]int, 0)
	for index, id := range swept.Graph.cos[:swept.Graph.PrimaryOutputCount()] {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		driver, e := swept.Graph.objects[id].Fanin0(id)
		if e != nil {
			return nil, e
		}
		if driver.ID() != 0 {
			result = append(result, index)
		}
	}
	return result, nil
}
