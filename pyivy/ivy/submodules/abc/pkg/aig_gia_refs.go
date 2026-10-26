package abc

// Reference counting from src/aig/gia/giaUtil.c:Gia_ManCreateRefs.

// ReferenceCounts returns incoming graph-edge counts for every object,
// including constant zero and dangling cones. Buffers contribute one edge,
// XORs two and native muxes three. CO edges are included. The result is owned
// by the caller; manager scratch fields are not modified. maxObjects zero
// selects one million objects.
func (m *GiaManager) ReferenceCounts(maxObjects int) ([]uint64, error) {
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
	counts := make([]uint64, len(m.objects))
	for i, obj := range m.objects {
		id := GiaID(i)
		if obj.IsAnd() || obj.IsCO() {
			a, err := obj.Fanin0(id)
			if err != nil {
				return nil, err
			}
			counts[a.ID()]++
			if obj.IsAnd() && !obj.IsBuffer() {
				b, err := obj.Fanin1(id)
				if err != nil {
					return nil, err
				}
				counts[b.ID()]++
			}
			if control, yes := m.MuxControl(id); yes {
				counts[control.ID()]++
			}
		}
	}
	return counts, nil
}
