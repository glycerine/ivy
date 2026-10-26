package abc

// Closest fanout dominators from src/aig/gia/giaMffc.c:
// Gia_ManAddDom and Gia_ManComputeDoms.

// ClosestDominators returns an owned closest-dominator forest. Outputs point
// to themselves; a node with incompatible fanout roots points to itself;
// objects with no consumers remain InvalidGiaID. Native mux controls and
// buffers use actual edges. Unlike C, dangling-chain intersection terminates
// safely at missing parents. maxObjects zero selects one million objects.
func (m *GiaManager) ClosestDominators(maxObjects int) ([]GiaID, error) {
	if maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > maxObjects {
		return nil, ErrLimit
	}
	dom := make([]GiaID, len(m.objects))
	for i := range dom {
		dom[i] = InvalidGiaID
	}
	add := func(node, candidate GiaID) error {
		if dom[node] == InvalidGiaID {
			dom[node] = candidate
			return nil
		}
		other := dom[node]
		for candidate != other {
			if candidate > other {
				next := dom[other]
				if next == other || next == InvalidGiaID {
					dom[node] = node
					return nil
				}
				if next < other {
					return ErrInvariant
				}
				other = next
			} else {
				next := dom[candidate]
				if next == candidate || next == InvalidGiaID {
					dom[node] = node
					return nil
				}
				if next < candidate {
					return ErrInvariant
				}
				candidate = next
			}
		}
		dom[node] = candidate
		return nil
	}
	for i := len(m.objects) - 1; i >= 1; i-- {
		obj := m.objects[i]
		id := GiaID(i)
		if obj.IsCI() {
			continue
		}
		if obj.IsCO() {
			dom[i] = id
		}
		children, n, err := m.crossCutFanins(id)
		if err != nil {
			return nil, err
		}
		for _, child := range children[:n] {
			if err = add(child, id); err != nil {
				return nil, err
			}
		}
	}
	return dom, nil
}
