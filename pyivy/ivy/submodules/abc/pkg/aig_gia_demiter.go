package abc

// Combinational pair splitting from src/aig/gia/giaDup.c:
// Gia_ManCollectReach/CollectReach2 and Gia_ManDemiterDual/DemiterTwoWords.

// GiaDemiter contains two independently owned graphs and their target-CO to
// source-CO ordinal projections. Both graphs retain every original CI.
type GiaDemiter struct {
	Graphs  [2]*GiaManager
	Outputs [2][]int
}

// SplitPairedOutputs separates combinational paired outputs. With twoWords
// false it selects even/odd outputs; true selects first/second halves.
// Output count must be even. Shared gates copy independently into each side
// in selected-root DFS postorder, retaining raw extended node types. Source
// constraint roles clear and name/spec survive. No source scratch is changed;
// errors expose neither partial graph. Empty output pairs are supported.
func (m *GiaManager) SplitPairedOutputs(twoWords bool, options GiaCofactorOptions) (GiaDemiter, error) {
	fail := func(e error) (GiaDemiter, error) { return GiaDemiter{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if m.registers != 0 || len(m.cos)%2 != 0 {
		return fail(ErrDimension)
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	result := GiaDemiter{}
	for side := 0; side < 2; side++ {
		roots := make([]GiaID, 0, len(m.cos)/2)
		for i, id := range m.cos {
			selected := i%2 == side
			if twoWords {
				selected = (i < len(m.cos)/2) == (side == 0)
			}
			if selected {
				roots = append(roots, id)
				result.Outputs[side] = append(result.Outputs[side], i)
			}
		}
		if stopped() {
			return fail(ErrCancelled)
		}
		nodes, e := m.CollectTransitiveFanins(roots, limit)
		if e != nil {
			return fail(e)
		}
		result.Graphs[side], _, e = m.DuplicateFromLists(m.cis, nodes, roots, 0, options)
		if e != nil {
			return fail(e)
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}
