package reo

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"math/rand"
)

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
	support := dd.Support(root)
	if len(support) < 2 {
		return
	}
	permutation := identity(len(support))
	random := rand.New(rand.NewSource(0x12341234))
	for i := 0; i < 120; i++ {
		offset := random.Intn(len(support) - 1)
		permutation[offset], permutation[offset+1] = permutation[offset+1], permutation[offset]
	}
	// The source test first maps support to the topmost manager levels.
	forward := identity(len(dd.Vars))
	for level, index := range support {
		forward[index] = dd.InvPerm[level]
	}
	remapped := permuteRoot(dd, root, forward)
	if remapped == nil {
		panic("reo: shuffle support remapping failed")
	}
	cudd.Cudd_Ref(remapped)
	defer cudd.Cudd_RecursiveDeref(dd, remapped)
	result, inverse := p.Shuffle(dd, remapped, permutation)
	if result == nil || permuteRoot(dd, result, inverse) != remapped {
		panic("reo: shuffle verification failed")
	}
}
