package abc

// Ordered six-input boundary evaluation from aig/gia/giaTruth.c:
// Gia_ObjComputeTruth6 and Gia_ObjComputeTruthTable6Lut_rec.

// BoundaryTruth6 evaluates a structural boundary in supplied input-slot order.
// Repeated leaves use their LAST slot, matching C's repeated scratch writes;
// other slots remain unused. This differs from CutTruth's first-slot policy.
// At most six nonconstant, non-CO leaves are accepted. Every reached CI must
// be covered. Constants within the cone remain fixed; phased/CI/CO roots,
// buffers and native XOR/mux are supported safely. Source scratch is unchanged.
func (m *GiaManager) BoundaryTruth6(root GiaLit, leaves []GiaID, options GiaCutTruthOptions) (uint64, error) {
	if len(leaves) > 6 {
		return 0, ErrDimension
	}
	unique := make([]GiaID, 0, len(leaves))
	targets := make([]int, 0, len(leaves))
	for slot, id := range leaves {
		found := -1
		for i, old := range unique {
			if old == id {
				found = i
				break
			}
		}
		if found < 0 {
			unique = append(unique, id)
			targets = append(targets, slot)
		} else {
			targets[found] = slot
		}
	}
	cut, err := m.CutTruth(root, unique, options)
	if err != nil {
		return 0, err
	}
	truth, err := cut.Table.ExpandVariables(len(leaves), targets)
	if err != nil {
		return 0, err
	}
	if options.Stop != nil && options.Stop() {
		return 0, ErrCancelled
	}
	return truth.words[0], nil
}
