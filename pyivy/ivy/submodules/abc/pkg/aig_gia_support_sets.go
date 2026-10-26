package abc

// Structural CI support sets from src/aig/gia/giaDup.c:
// Gia_ManCreateNodeSupps and Gia_ManCreateCoSupps.

// NodeSupports returns sorted CI terminal indices for each requested object.
// Supports are structural, including logically redundant fanins. COs use their
// driver; native mux controls participate. maxEntries bounds all retained
// support entries, including intermediate sets and returned copies; zero uses
// one million. Requested IDs may repeat and each result is independently owned.
func (m *GiaManager) NodeSupports(nodes []GiaID, maxObjects, maxEntries int) ([][]int, error) {
	if m == nil || maxObjects < 0 || maxEntries < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > maxObjects || len(nodes) > maxObjects {
		return nil, ErrLimit
	}
	for _, id := range nodes {
		if uint64(id) >= uint64(len(m.objects)) {
			return nil, ErrInvalidLiteral
		}
	}
	supports := make([][]int, len(m.objects))
	used := 0
	reserve := func(n int) bool {
		if n > maxEntries-used {
			return false
		}
		used += n
		return true
	}
	for index, id := range m.cis {
		if !reserve(1) {
			return nil, ErrLimit
		}
		supports[id] = []int{index}
	}
	// Merge directly into the final set, avoiding unbounded temporary unions.
	merge := func(a, b []int) ([]int, error) {
		result := make([]int, 0)
		for i, j := 0, 0; i < len(a) || j < len(b); {
			var v int
			if j == len(b) || (i < len(a) && a[i] < b[j]) {
				v = a[i]
				i++
			} else if i == len(a) || b[j] < a[i] {
				v = b[j]
				j++
			} else {
				v = a[i]
				i++
				j++
			}
			if !reserve(1) {
				return nil, ErrLimit
			}
			result = append(result, v)
		}
		return result, nil
	}
	for i, obj := range m.objects {
		id := GiaID(i)
		if !obj.IsAnd() && !obj.IsCO() {
			continue
		}
		first, err := obj.Fanin0(id)
		if err != nil {
			return nil, err
		}
		// Buffers and COs can share immutable internal support storage.
		if obj.IsCO() || obj.IsBuffer() {
			supports[id] = supports[first.ID()]
			continue
		}
		second, err := obj.Fanin1(id)
		if err != nil {
			return nil, err
		}
		supports[id], err = merge(supports[first.ID()], supports[second.ID()])
		if err != nil {
			return nil, err
		}
		if control, yes := m.MuxControl(id); yes {
			// Release the previous union's accounting before replacing it.
			old := supports[id]
			used -= len(old)
			supports[id], err = merge(old, supports[control.ID()])
			if err != nil {
				return nil, err
			}
		}
	}
	result := make([][]int, len(nodes))
	for i, id := range nodes {
		if !reserve(len(supports[id])) {
			return nil, ErrLimit
		}
		result[i] = append([]int(nil), supports[id]...)
	}
	return result, nil
}

// OutputSupports returns structural CI supports in CO order, including RIs.
func (m *GiaManager) OutputSupports(maxObjects, maxEntries int) ([][]int, error) {
	if m == nil {
		return nil, ErrDimension
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	return m.NodeSupports(m.cos, maxObjects, maxEntries)
}
