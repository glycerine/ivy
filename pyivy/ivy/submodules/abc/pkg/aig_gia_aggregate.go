package abc

// Output aggregation from src/aig/gia/giaDup.c:Gia_ManDupAndOr.

// AggregateOutputs folds all combinational outputs in order using AND (useOr
// false) or OR, optionally complements the result, and repeats it outputCount
// times. Empty AND is true; empty OR is false. All CIs survive, including
// unused ones. Source constraints participate as data and result roles clear.
// Name survives, Spec clears as in C. Extended nodes lower to hashed ANDs;
// dangling gates are removed. Source scratch is unchanged.
func (m *GiaManager) AggregateOutputs(outputCount int, useOr, complement bool, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || outputCount < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	if m.registers != 0 {
		return nil, ErrDimension
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return nil, ErrLimit
	}
	objectLimit := options.MaxObjects
	if objectLimit == 0 {
		objectLimit = 1 << 20
	}
	if _, e := checkedSize(1, 1, outputCount, objectLimit); e != nil {
		return nil, e
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	g, e := NewGiaManagerWithLimit(objectLimit)
	if e != nil {
		return nil, e
	}
	g.Name = m.Name
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], e = g.AppendCI()
		if e != nil {
			return nil, e
		}
	}
	if e = g.HashStart(); e != nil {
		return nil, e
	}
	for i, obj := range m.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[i], e = hashGiaCopiedGate(m, g, GiaID(i), mapping)
		if e != nil {
			return nil, e
		}
	}
	result := GiaConst0.Not()
	if useOr {
		result = GiaConst0
	}
	for _, id := range m.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, err
		}
		signal, err := mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, err
		}
		if useOr {
			result, e = g.HashOr(result, signal)
		} else {
			result, e = g.HashAnd(result, signal)
		}
		if e != nil {
			return nil, e
		}
	}
	result = result.NotIf(complement)
	for i := 0; i < outputCount; i++ {
		if stopped() {
			return nil, ErrCancelled
		}
		if _, e = g.AppendCO(result); e != nil {
			return nil, e
		}
	}
	g.HashStop()
	if stopped() {
		return nil, ErrCancelled
	}
	cleaned, _, e := g.Cleanup()
	return cleaned, e
}
