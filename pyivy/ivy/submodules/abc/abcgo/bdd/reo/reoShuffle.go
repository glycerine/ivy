package reo

// Shuffle maps support-level positions according to permutation and optionally
// returns the inverse variable map. Like the original helper, the result is not
// referenced; malformed permutations return nil without changing the input.
func (p *Manager) Shuffle(dd *DdManager, root *DdNode, permutation []int) (*DdNode, []int) {
	if dd == nil || root == nil {
		return nil, nil
	}
	oldIterations := p.iterations
	p.iterations = 0
	_, _, err := p.Reorder(dd, []*DdNode{root})
	p.iterations = oldIterations
	if err != nil || len(permutation) != len(p.support) {
		return nil, nil
	}
	seen := make([]bool, len(permutation))
	order := make([]int, len(permutation))
	for i, v := range permutation {
		if v < 0 || v >= len(permutation) || seen[v] {
			return nil, nil
		}
		seen[v] = true
		order[i] = p.support[v]
	}
	if err := p.applyOrder(order); err != nil {
		return nil, nil
	}
	inverse := identity(len(dd.Vars))
	for level, v := range p.order {
		inverse[p.slots[level]] = v
	}
	p.stats.After = cloneProfile(p.current)
	return p.roots[0], inverse
}
func reoShuffle(p *Manager, dd *DdManager, root *DdNode, permutation, inverse []int) *DdNode {
	result, mapping := p.Shuffle(dd, root, permutation)
	if result != nil && inverse != nil {
		if len(inverse) < len(mapping) {
			panic("reo: short inverse permutation")
		}
		copy(inverse, mapping)
	}
	return result
}
func Extra_ShuffleTest(p *Manager, dd *DdManager, root *DdNode) {
	support := diagramSupport(dd, []*DdNode{root})
	permutation := identity(len(support))
	for i, j := 0, len(permutation)-1; i < j; i, j = i+1, j-1 {
		permutation[i], permutation[j] = permutation[j], permutation[i]
	}
	result, inverse := p.Shuffle(dd, root, permutation)
	if result == nil || permuteRoot(dd, result, inverse) != root {
		panic("reo: shuffle verification failed")
	}
}
