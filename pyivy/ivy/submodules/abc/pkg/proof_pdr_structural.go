package abc

// Port of src/proof/pdr/pdrMan.c:Pdr_ManDeriveFlopPriorities2/3.

import "slices"

// StructuralFlopPriorities ranks registers by fanout cost. Recognized mux
// controls optionally receive a bonus greater than any plain reference count.
// Rank zero has lowest priority; equal-cost ranking follows ABC's cost sort.
func (m *GiaManager) StructuralFlopPriorities(muxControls bool, maxObjects int) ([]uint64, error) {
	costs, err := m.structuralFlopCosts(muxControls, maxObjects)
	if err != nil {
		return nil, err
	}
	order := CostPermutation(costs)
	ranks := make([]uint64, len(order))
	for rank, reg := range order {
		ranks[reg] = uint64(rank)
	}
	return ranks, nil
}

// StructuralFlopPrioritiesStable implements the alternate bucket-ranking
// algorithm: equal-cost registers retain increasing register-index order.
// Sorting only the register indices avoids allocating a sparse cost domain.
func (m *GiaManager) StructuralFlopPrioritiesStable(muxControls bool, maxObjects int) ([]uint64, error) {
	costs, err := m.structuralFlopCosts(muxControls, maxObjects)
	if err != nil {
		return nil, err
	}
	order := make([]int, len(costs))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		if costs[a] < costs[b] {
			return -1
		}
		if costs[a] > costs[b] {
			return 1
		}
		return 0
	})
	ranks := make([]uint64, len(order))
	for rank, reg := range order {
		ranks[reg] = uint64(rank)
	}
	return ranks, nil
}

func (m *GiaManager) structuralFlopCosts(muxControls bool, maxObjects int) ([]uint64, error) {
	refs, err := m.ReferenceCounts(maxObjects)
	if err != nil {
		return nil, err
	}
	costs := make([]uint64, m.registers)
	maximum := uint64(0)
	for r, id := range m.cis[len(m.cis)-m.registers:] {
		costs[r] = refs[id]
		if refs[id] > maximum {
			maximum = refs[id]
		}
	}
	if muxControls {
		controls := make([]bool, len(m.objects))
		for i := range m.objects {
			c, _, _, ok, err := m.RecognizeMux(GiaID(i))
			if err != nil {
				return nil, err
			}
			if ok {
				controls[c.ID()] = true
			}
		}
		for r, id := range m.cis[len(m.cis)-m.registers:] {
			if controls[id] {
				costs[r] += maximum + 1
			}
		}
	}
	return costs, nil
}
