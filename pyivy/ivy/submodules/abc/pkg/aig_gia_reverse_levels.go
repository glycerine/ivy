package abc

// Reverse levels from src/aig/gia/giaUtil.c:Gia_ManLevelRNum and
// src/aig/gia/giaDfs.c:Gia_ManLevelizeR.

// ReverseLevels preserves ABC's reverse traversal metric: CO drivers begin at
// one, buffers add zero, and every other AND-like object adds one to each of
// its two packed fanins. This C metric ignores native mux control edges and
// does not use the forward XOR/mux delay of two. It is not a timing analysis.
// Dangling AND cones also contribute. maxObjects zero selects one million.
func (m *GiaManager) ReverseLevels(maxObjects int) ([]uint32, uint32, error) {
	if maxObjects < 0 {
		return nil, 0, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, 0, err
	}
	if len(m.objects) > maxObjects {
		return nil, 0, ErrLimit
	}
	levels := make([]uint32, len(m.objects))
	maximum := uint32(0)
	for i := len(m.objects) - 1; i >= 0; i-- {
		obj := m.objects[i]
		id := GiaID(i)
		if obj.IsAnd() || obj.IsCO() {
			a, err := obj.Fanin0(id)
			if err != nil {
				return nil, 0, err
			}
			switch {
			case obj.IsBuffer():
				levels[a.ID()] = max(levels[a.ID()], levels[i])
			case obj.IsAnd():
				b, err := obj.Fanin1(id)
				if err != nil {
					return nil, 0, err
				}
				levels[a.ID()] = max(levels[a.ID()], levels[i]+1)
				levels[b.ID()] = max(levels[b.ID()], levels[i]+1)
			case obj.IsCO():
				levels[a.ID()] = max(levels[a.ID()], 1)
			}
		} else {
			maximum = max(maximum, levels[i])
		}
	}
	return levels, maximum, nil
}

// ReverseLevelize groups all nonconstant objects with positive reverse level,
// plus COs at level zero. Within each level, IDs are increasing.
func (m *GiaManager) ReverseLevelize(maxObjects int) ([][]GiaID, error) {
	levels, maximum, err := m.ReverseLevels(maxObjects)
	if err != nil {
		return nil, err
	}
	groups := make([][]GiaID, int(maximum)+1)
	for i, obj := range m.objects {
		if i == 0 || (!obj.IsCO() && levels[i] == 0) {
			continue
		}
		if levels[i] > maximum {
			return nil, ErrInvariant
		}
		groups[levels[i]] = append(groups[levels[i]], GiaID(i))
	}
	return groups, nil
}
