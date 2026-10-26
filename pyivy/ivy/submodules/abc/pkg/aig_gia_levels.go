package abc

// Level assignment and traversal from src/aig/gia/giaUtil.c:Gia_ManLevelNum,
// Gia_ManBfsForCrossCut and src/aig/gia/giaDfs.c:Gia_ManLevelize.

// Levels returns owned object levels and maximum depth. CIs/constants start
// at zero, COs/buffers inherit their driver level, ANDs add one, and native
// XORs/muxes add two. Dangling logic participates, as in the default C mode.
func (m *GiaManager) Levels(maxObjects int) ([]uint32, uint32, error) {
	return m.LevelsWithInputs(nil, maxObjects)
}

// LevelsWithInputs ports Gia_ManSetLevels. Non-nil inputLevels must contain
// one level per CI, including register outputs; nil means all zero. Results
// are owned and arithmetic overflow returns ErrLimit.
func (m *GiaManager) LevelsWithInputs(inputLevels []uint32, maxObjects int) ([]uint32, uint32, error) {
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
	if inputLevels != nil && len(inputLevels) != len(m.cis) {
		return nil, 0, ErrDimension
	}
	levels := make([]uint32, len(m.objects))
	maximum := uint32(0)
	if inputLevels != nil {
		for i, id := range m.cis {
			levels[id] = inputLevels[i]
			maximum = max(maximum, inputLevels[i])
		}
	}
	for i, obj := range m.objects {
		id := GiaID(i)
		if !obj.IsAnd() && !obj.IsCO() {
			continue
		}
		a, err := obj.Fanin0(id)
		if err != nil {
			return nil, 0, err
		}
		level := levels[a.ID()]
		if obj.IsAnd() && !obj.IsBuffer() {
			b, err := obj.Fanin1(id)
			if err != nil {
				return nil, 0, err
			}
			level = max(level, levels[b.ID()])
			if control, yes := m.MuxControl(id); yes {
				level = max(level, levels[control.ID()])
				if level > ^uint32(0)-2 {
					return nil, 0, ErrLimit
				}
				level += 2
			} else if obj.IsXor() {
				if level > ^uint32(0)-2 {
					return nil, 0, ErrLimit
				}
				level += 2
			} else {
				if level == ^uint32(0) {
					return nil, 0, ErrLimit
				}
				level++
			}
		}
		levels[i] = level
		maximum = max(maximum, level)
	}
	return levels, maximum, nil
}

// Levelize groups AND-like objects (including buffers, native XORs and muxes)
// by level, in increasing object-ID order. Constants and terminals are omitted.
func (m *GiaManager) Levelize(maxObjects int) ([][]GiaID, error) {
	levels, maximum, err := m.Levels(maxObjects)
	if err != nil {
		return nil, err
	}
	result := make([][]GiaID, int(maximum)+1)
	for i, obj := range m.objects {
		if obj.IsAnd() {
			result[levels[i]] = append(result[levels[i]], GiaID(i))
		}
	}
	return result, nil
}

// BreadthFirstCrossCutOrder concatenates levelized AND-like nodes. It is the
// ordering used by Gia_ManBfsForCrossCut, not a computed frontier-size metric.
func (m *GiaManager) BreadthFirstCrossCutOrder(maxObjects int) ([]GiaID, error) {
	levels, err := m.Levelize(maxObjects)
	if err != nil {
		return nil, err
	}
	result := make([]GiaID, 0)
	for _, level := range levels {
		result = append(result, level...)
	}
	return result, nil
}

// AverageOutputLevel ports Gia_ManLevelAve using checked forward levels. All
// COs, including register inputs, participate. No COs returns ErrDimension
// rather than C's division by zero. inputLevels follows LevelsWithInputs.
func (m *GiaManager) AverageOutputLevel(inputLevels []uint32, maxObjects int) (float64, error) {
	levels, _, err := m.LevelsWithInputs(inputLevels, maxObjects)
	if err != nil {
		return 0, err
	}
	if len(m.cos) == 0 {
		return 0, ErrDimension
	}
	total := uint64(0)
	for _, id := range m.cos {
		total += uint64(levels[id])
	}
	return float64(total) / float64(len(m.cos)), nil
}
