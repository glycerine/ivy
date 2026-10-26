package abc

// Weighted reverse/required levels and slack from src/aig/gia/giaUtil.c:
// Gia_ManReverseLevel, Gia_ManRequiredLevel and Gia_ManComputeSlacks.

// WeightedReverseLevels propagates from all AND-like nodes, including dangling
// ones, with no extra CO distance. ANDs add one, XORs/muxes add two and buffers
// add zero. Native mux controls participate. This differs from ReverseLevels.
func (m *GiaManager) WeightedReverseLevels(maxObjects int) ([]uint32, error) {
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
	levels := make([]uint32, len(m.objects))
	for i := len(m.objects) - 1; i >= 0; i-- {
		obj := m.objects[i]
		if !obj.IsAnd() {
			continue
		}
		delay := uint32(1)
		if _, mux := m.MuxControl(GiaID(i)); mux || obj.IsXor() {
			delay = 2
		} else if obj.IsBuffer() {
			delay = 0
		}
		if levels[i] > ^uint32(0)-delay {
			return nil, ErrLimit
		}
		children, n, err := m.crossCutFanins(GiaID(i))
		if err != nil {
			return nil, err
		}
		for _, child := range children[:n] {
			levels[child] = max(levels[child], levels[i]+delay)
		}
	}
	return levels, nil
}

// RequiredLevels subtracts weighted reverse distance from the maximum reverse
// distance at CIs, matching ABC. Signed results preserve negative values for
// unusual constant-only cones. Results are owned; source scratch is unchanged.
func (m *GiaManager) RequiredLevels(maxObjects int) ([]int64, error) {
	reverse, err := m.WeightedReverseLevels(maxObjects)
	if err != nil {
		return nil, err
	}
	maximum := uint32(0)
	for _, id := range m.cis {
		maximum = max(maximum, reverse[id])
	}
	result := make([]int64, len(reverse))
	for i, level := range reverse {
		result[i] = int64(maximum) - int64(level)
	}
	return result, nil
}

// LevelSlacks returns maximum forward depth minus forward and weighted reverse
// distance for each object. This is ABC's structural level metric, not timing.
func (m *GiaManager) LevelSlacks(maxObjects int) ([]int64, error) {
	forward, maximum, err := m.Levels(maxObjects)
	if err != nil {
		return nil, err
	}
	reverse, err := m.WeightedReverseLevels(maxObjects)
	if err != nil {
		return nil, err
	}
	result := make([]int64, len(forward))
	for i := range result {
		result[i] = int64(maximum) - int64(forward[i]) - int64(reverse[i])
	}
	return result, nil
}
