package abc

// Level-based reverse ordering from src/aig/gia/giaDfs.c:Gia_ManOrderReverse.

// ReverseLevelOrder returns every object in descending forward-level order.
// COs precede other objects at the same level; remaining ties use increasing
// object ID. This is ABC's ordering, not a strict reverse topological order
// when zero-delay buffers share a level with their drivers. maxObjects zero
// selects one million objects. Levels and traversal scratch are not mutated.
func (m *GiaManager) ReverseLevelOrder(maxObjects int) ([]GiaID, error) {
	levels, maximum, err := m.Levels(maxObjects)
	if err != nil {
		return nil, err
	}
	groups := make([][]GiaID, int(maximum)+1)
	for _, id := range m.cos {
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, err
		}
		level := levels[driver.ID()]
		groups[level] = append(groups[level], id)
	}
	for i, obj := range m.objects {
		if !obj.IsCO() {
			groups[levels[i]] = append(groups[levels[i]], GiaID(i))
		}
	}
	order := make([]GiaID, 0, len(m.objects))
	for i := len(groups) - 1; i >= 0; i-- {
		order = append(order, groups[i]...)
	}
	return order, nil
}
