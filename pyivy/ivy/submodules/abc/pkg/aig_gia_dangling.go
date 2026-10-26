package abc

// Dangling-root analysis from src/aig/gia/giaUtil.c:
// Gia_ManHasDangling, Gia_ManMarkDangling and Gia_ManGetDangling.

// DanglingRoots returns AND-like objects with no incoming graph references,
// in object-ID order. This is the set of unconsumed roots, not every object
// outside the output cones. Buffers/XORs/muxes participate, and native mux
// control edges are included. Source mark bits are preserved. Zero cap selects
// one million objects.
func (m *GiaManager) DanglingRoots(maxObjects int) ([]GiaID, error) {
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
	consumed := make([]bool, len(m.objects))
	for i := range m.objects {
		children, n, err := m.crossCutFanins(GiaID(i))
		if err != nil {
			return nil, err
		}
		for _, child := range children[:n] {
			consumed[child] = true
		}
	}
	roots := make([]GiaID, 0)
	for i, obj := range m.objects {
		if obj.IsAnd() && !consumed[i] {
			roots = append(roots, GiaID(i))
		}
	}
	return roots, nil
}

// DanglingRootCount corresponds to Gia_ManHasDangling's actual return value:
// it is a count, despite that C function's Boolean-looking name.
func (m *GiaManager) DanglingRootCount(maxObjects int) (int, error) {
	roots, err := m.DanglingRoots(maxObjects)
	return len(roots), err
}
