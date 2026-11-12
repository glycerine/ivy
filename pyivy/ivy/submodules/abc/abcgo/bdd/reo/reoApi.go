package reo

import (
	"os"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

// Compatibility names retain the original REO interface; pointer arrays and
// permutation arrays become slices. Reorder returns an unreferenced root;
// ReorderArray references each output exactly once, as the C API requires.
func Extra_ReorderInit(variables, nodes int) *Manager            { return New(variables, nodes) }
func Extra_ReorderQuit(p *Manager)                               { p.Close() }
func Extra_ReorderSetMinimizationType(p *Manager, kind Minimize) { p.SetMinimization(kind) }
func Extra_ReorderSetRemapping(p *Manager, enabled int)          { p.SetRemapping(enabled != 0) }
func Extra_ReorderSetIterations(p *Manager, n int)               { p.SetIterations(n) }
func Extra_ReorderSetVerification(p *Manager, enabled int)       { p.SetVerification(enabled != 0) }
func Extra_ReorderSetVerbosity(p *Manager, enabled int) {
	if enabled != 0 {
		p.SetOutput(os.Stdout)
	} else {
		p.SetOutput(nil)
	}
}
func Extra_Reorder(p *Manager, dd *DdManager, root *DdNode, order []int) *DdNode {
	result, permutation, err := p.Reorder(dd, []*DdNode{root})
	if err != nil {
		return nil
	}
	if order != nil {
		if len(order) < len(permutation) {
			panic("reo: short permutation output")
		}
		copy(order, permutation)
	}
	return result[0]
}
func Extra_ReorderArray(p *Manager, dd *DdManager, roots, result []*DdNode, n int, order []int) {
	if n < 0 || n > len(roots) || n > len(result) {
		panic("reo: invalid root count")
	}
	reordered, permutation, err := p.Reorder(dd, roots[:n])
	if err != nil {
		panic(err)
	}
	if order != nil {
		if len(order) < len(permutation) {
			panic("reo: short permutation output")
		}
		copy(order, permutation)
	}
	copy(result, reordered)
	for _, root := range reordered {
		cudd.Cudd_Ref(root)
	}
}
func reoReorderArray(p *Manager, dd *DdManager, roots, result []*DdNode, n int, order []int) {
	Extra_ReorderArray(p, dd, roots, result, n, order)
}
