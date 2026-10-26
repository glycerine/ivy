package abc

// Deterministic result-data replacement for
// src/aig/gia/giaMffc.c:Gia_ManComputeDomsTry.

type GiaDominatedMffcOptions struct {
	MinimumSize, MaxObjects, MaxCandidates int
	Stop                                   func() bool
}

type GiaDominatedMffcCandidate struct {
	// Object is the self-dominated AND/CO selected by the source loop. Root is
	// Object for an AND and its driver for a CO, so roots may occur more than
	// once exactly as they can in the diagnostic driver.
	Object, Root GiaID
	Mffc         GiaMffc
	Cone         *GiaManager
}

// DominatedMffcCandidates replaces the source's stdout, fixed-name AIGER dump
// and external collapse-test call with owned deterministic records. Zero
// MinimumSize preserves the source threshold of ten nodes; zero caps select
// one million. Boundary/reference pairs use the source selection-sort order.
// Cone is the driver's k=0 dominated-cone derivation.
func (m *GiaManager) DominatedMffcCandidates(options GiaDominatedMffcOptions) ([]GiaDominatedMffcCandidate, error) {
	if m == nil || options.MinimumSize < 0 || options.MaxObjects < 0 || options.MaxCandidates < 0 {
		return nil, ErrDimension
	}
	if options.MinimumSize == 0 {
		options.MinimumSize = 10
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxCandidates == 0 {
		options.MaxCandidates = 1 << 20
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	dominators, err := m.ClosestDominators(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	result := make([]GiaDominatedMffcCandidate, 0)
	for i := 1; i < len(m.objects); i++ {
		if stopped() {
			return nil, ErrCancelled
		}
		object := m.objects[i]
		if (!object.IsAnd() && !object.IsCO()) || dominators[i] != GiaID(i) {
			continue
		}
		root := GiaID(i)
		if object.IsCO() {
			driver, e := object.Fanin0(root)
			if e != nil {
				return nil, e
			}
			root = driver.ID()
		}
		if !m.objects[root].IsAnd() {
			continue
		}
		cone, e := m.Mffc(root, options.MaxObjects)
		if e != nil {
			return nil, e
		}
		if cone.Size < options.MinimumSize {
			continue
		}
		if len(result) >= options.MaxCandidates {
			return nil, ErrLimit
		}
		for left := 0; left+1 < len(cone.Boundary); left++ {
			best := left
			for right := left + 1; right < len(cone.Boundary); right++ {
				if cone.BoundaryReferenceDeltas[right] < cone.BoundaryReferenceDeltas[best] {
					best = right
				}
			}
			cone.Boundary[left], cone.Boundary[best] = cone.Boundary[best], cone.Boundary[left]
			cone.BoundaryReferenceDeltas[left], cone.BoundaryReferenceDeltas[best] = cone.BoundaryReferenceDeltas[best], cone.BoundaryReferenceDeltas[left]
		}
		derived, e := m.DeriveDominatedCone(root, cone.Boundary, 0, options.MaxObjects)
		if e != nil {
			return nil, e
		}
		result = append(result, GiaDominatedMffcCandidate{Object: GiaID(i), Root: root, Mffc: cone, Cone: derived})
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return result, nil
}
