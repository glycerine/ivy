package reo

import (
	"errors"
	"fmt"
	"sort"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func identity(n int) []int {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return p
}

// Reorder returns roots expressed on the selected support slots, and a map from
// each resulting variable to its original variable. Permuting the roots with
// that map restores the original functions. Inputs and manager ordering survive.
func (p *Manager) Reorder(dd *DdManager, roots []*DdNode) ([]*DdNode, []int, error) {
	if dd == nil {
		return nil, nil, errors.New("reo: nil diagram manager")
	}
	inverse := identity(len(dd.Vars))
	for _, root := range roots {
		if root == nil || permuteRoot(dd, root, inverse) == nil {
			return nil, nil, errors.New("reo: invalid or foreign root")
		}
	}
	// Transfer caches and arena belong to one reordering run. Reusing them
	// after a support remap would attach stale levels to otherwise identical nodes.
	p.stop()
	p.units = nil
	p.nodes = nil
	p.planes = nil
	p.dd = dd
	p.original = append(p.original[:0], roots...)
	p.stats = Statistics{}
	p.hasADD = diagramsHaveADD(dd, roots)

	// The source engine falls back to node minimization for BDD width: its
	// frontier-width heuristic is intended for ADDs, without complemented edges.
	if p.criterion == REO_MINIMIZE_WIDTH && !p.hasADD {
		p.criterion = REO_MINIMIZE_NODES
	}
	support := diagramSupport(dd, roots)
	p.support = p.support[:0]
	for v := range support {
		p.support = append(p.support, v)
	}
	sort.Slice(p.support, func(i, j int) bool { return dd.Perm[p.support[i]] < dd.Perm[p.support[j]] })
	if len(p.support) >= REO_CONST_LEVEL {
		return nil, nil, errors.New("reo: support exceeds unit level capacity")
	}
	p.slots = append(p.slots[:0], p.support...)
	if p.remapUp {
		copy(p.slots, dd.InvPerm[:len(p.support)])
	}
	p.order = append(p.order[:0], p.support...)
	var err error
	p.roots, p.current, err = p.evaluate(p.order)
	if err != nil {
		return nil, nil, err
	}
	p.stats.Before = cloneProfile(p.current)
	p.printProfile("INITIAL", p.current)
	for iteration := 0; iteration < p.iterations && len(p.order) > 1; iteration++ {
		before := p.cost(p.current)
		reoReorderSift(p)
		p.stats.Iterations++
		p.printProfile(fmt.Sprintf("ITER #%d", iteration+1), p.current)
		if p.cost(p.current) >= before-REO_COST_EPSILON {
			break
		}
	}
	p.stats.After = cloneProfile(p.current)
	for level, v := range p.order {
		inverse[p.slots[level]] = v
	}
	if p.verify {
		for i, root := range p.roots {
			if permuteRoot(dd, root, inverse) != p.original[i] {
				return nil, nil, errors.New("reo: inverse permutation verification failed")
			}
		}
	}
	return append([]*DdNode(nil), p.roots...), inverse, nil
}
func (p *Manager) evaluate(order []int) ([]*DdNode, Profile, error) {
	forward := identity(len(p.dd.Vars))
	for level, v := range order {
		forward[v] = p.slots[level]
	}
	roots := make([]*DdNode, len(p.original))
	for i, root := range p.original {
		roots[i] = permuteRoot(p.dd, root, forward)
		if roots[i] == nil {
			return nil, Profile{}, errors.New("reo: diagram rebuild failed")
		}
	}
	return roots, profileRoots(roots, p.slots), nil
}
func (p *Manager) applyOrder(order []int) error {
	roots, profile, err := p.evaluate(order)
	if err != nil {
		return err
	}
	p.order = append(p.order[:0], order...)
	p.roots = roots
	p.current = profile
	return nil
}
func (p *Manager) cost(profile Profile) float64 {
	switch p.criterion {
	case REO_MINIMIZE_APL:
		return profile.APL
	case REO_MINIMIZE_WIDTH:
		return float64(profile.Width)
	default:
		return float64(profile.Nodes)
	}
}
func reoResizeStructures(p *Manager, variables, nodes, functions int) {
	if variables < 0 || nodes < 0 || functions < 0 {
		panic("reo: negative capacity")
	}
	chunks := int32((10*nodes+1)/unitChunkSize + 1)
	if chunks > p.nMemChunksAlloc {
		p.nMemChunksAlloc = chunks
	}
}

func diagramsHaveADD(dd *DdManager, roots []*DdNode) bool {
	seen := map[*DdNode]bool{}
	var visit func(*DdNode) bool
	visit = func(edge *DdNode) bool {
		node := cudd.Cudd_Regular(edge)
		if seen[node] {
			return false
		}
		seen[node] = true
		if cudd.Cudd_IsConstant(node) {
			return node != dd.One
		}
		return visit(node.T) || visit(node.E)
	}
	for _, root := range roots {
		if visit(root) {
			return true
		}
	}
	return false
}

// Dispatch by diagram family, never by terminal truthiness. ADD terminals carry
// numeric values and must not pass through Boolean ITE/permutation.
func permuteRoot(dd *DdManager, root *DdNode, permutation []int) *DdNode {
	if cudd.Cudd_IsAdd(root) {
		return cudd.Cudd_addPermute(dd, root, permutation)
	}
	return cudd.Cudd_bddPermute(dd, root, permutation)
}

func diagramSupport(dd *DdManager, roots []*DdNode) map[int]bool {
	support := map[int]bool{}
	for _, root := range roots {
		for _, index := range dd.Support(root) {
			support[index] = true
		}
	}
	return support
}
