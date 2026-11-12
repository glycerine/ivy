package reo

// Verification entry points historically located in reoTest.c. These use the
// Go manager and canonical inverse-permutation equality, avoiding C random and
// static process-global managers. Reordering behavior lives in the main engine.
import cudd "github.com/glycerine/abcgo/bdd/cudd/core"

func Extra_ReorderTest(dd *DdManager, root *DdNode) {
	p := New(len(dd.Vars), cudd.Cudd_DagSize(root))
	defer p.Close()
	p.SetVerification(true)
	if _, _, err := p.Reorder(dd, []*DdNode{root}); err != nil {
		panic(err)
	}
}
func Extra_ReorderTestArray(dd *DdManager, roots []*DdNode, n int) {
	if n < 0 || n > len(roots) {
		panic("reo: invalid root count")
	}
	p := New(len(dd.Vars), 0)
	defer p.Close()
	p.SetVerification(true)
	if _, _, err := p.Reorder(dd, roots[:n]); err != nil {
		panic(err)
	}
}

// Extra_bddReorderTest returns the resulting node count, not a Boolean status.
// The Go sifting engine is used here; the original CUDD-reference benchmark
// still depends on dynamic heap-reordering support in the core.
func Extra_bddReorderTest(dd *DdManager, root *DdNode) int {
	p := New(len(dd.Vars), 0)
	defer p.Close()
	p.SetVerification(true)
	_, _, err := p.Reorder(dd, []*DdNode{root})
	if err != nil {
		panic(err)
	}
	return p.stats.After.Nodes
}

// Extra_addReorderTest follows the source's nonzero-pattern conversion before
// measuring the reordered ADD. Numeric leaf distinctions are intentionally
// collapsed for this benchmark only, never in the main reordering path.
func Extra_addReorderTest(dd *DdManager, root *DdNode) int {
	memo := map[*DdNode]*DdNode{}
	var pattern func(*DdNode) *DdNode
	pattern = func(node *DdNode) *DdNode {
		if old := memo[node]; old != nil {
			return old
		}
		if cudd.Cudd_IsConstant(node) {
			if node.Value == 0 {
				return cudd.Cudd_addConst(dd, 0)
			}
			return dd.One
		}
		result := cudd.Cudd_addIte(dd, cudd.Cudd_addIthVar(dd, int(node.Index)), pattern(node.T), pattern(node.E))
		memo[node] = result
		return result
	}
	if root == nil || !cudd.Cudd_IsAdd(root) || cudd.Cudd_addPermute(dd, root, identity(len(dd.Vars))) == nil {
		panic("reo: invalid ADD benchmark root")
	}
	p := New(len(dd.Vars), 0)
	defer p.Close()
	p.SetVerification(true)
	_, _, err := p.Reorder(dd, []*DdNode{pattern(root)})
	if err != nil {
		panic(err)
	}
	return p.stats.After.Nodes
}
