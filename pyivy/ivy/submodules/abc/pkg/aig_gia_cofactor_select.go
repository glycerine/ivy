package abc

// Fanout pivot selection from src/aig/gia/giaCof.c:
// Gia_ManCofVars and Gia_ManDupCofAll.

// CofactorCandidates returns CI/gate IDs whose incoming reference count is
// at least minimumFanout. Order is increasing ID, matching the C loop despite
// its reverse-topological-order comment. A zero threshold includes unconsumed
// gates and unused CIs. Constants and COs are excluded. Zero object cap selects
// one million objects. Counts include native mux control edges.
func (m *GiaManager) CofactorCandidates(minimumFanout, maxObjects int) ([]GiaID, error) {
	if minimumFanout < 0 {
		return nil, ErrDimension
	}
	refs, err := m.ReferenceCounts(maxObjects)
	if err != nil {
		return nil, err
	}
	result := make([]GiaID, 0)
	for i, obj := range m.objects {
		if (obj.IsCI() || obj.IsAnd()) && refs[i] >= uint64(minimumFanout) {
			result = append(result, GiaID(i))
		}
	}
	return result, nil
}

// CofactorRebuildByFanout selects pivots inclusively by reference count, then
// runs the bounded multi-pivot reconstruction. The resulting graph remains
// equivalent to source; selecting no pivots returns an independent clone.
func (m *GiaManager) CofactorRebuildByFanout(minimumFanout int, options GiaMultiCofactorOptions) (*GiaManager, error) {
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	signals, err := m.CofactorCandidates(minimumFanout, options.MaxObjects)
	if err != nil {
		return nil, err
	}
	return m.CofactorRebuildAll(signals, options)
}
