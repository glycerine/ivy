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

// Extra_ReorderCudd performs the original independent-manager comparison using
// CUDD symmetric sifting, then maps the optimized order back into dd. The inverse
// permutation restores the original function; neither dd's order nor root changes.
func Extra_ReorderCudd(dd *DdManager, root *DdNode, inverse []int) *DdNode {
	if dd == nil || root == nil {
		return nil
	}
	n := len(dd.Vars)
	if inverse != nil && len(inverse) < n {
		panic("reo: short permutation output")
	}
	scratch := cudd.NewBDDManager(n)
	defer cudd.Cudd_Quit(scratch)
	forward := make([]int, n)
	for level, index := range dd.InvPerm {
		forward[index] = level
	}
	temporary := dd.Transfer(root, scratch, forward)
	if temporary == nil {
		return nil
	}
	cudd.Cudd_Ref(temporary)
	defer cudd.Cudd_RecursiveDeref(scratch, temporary)
	if cudd.Cudd_ReduceHeap(scratch, cudd.CUDD_REORDER_SYMM_SIFT, 1) == 0 {
		return nil
	}
	backward := make([]int, n)
	for level, index := range scratch.InvPerm {
		backward[index] = dd.InvPerm[level]
	}
	result := scratch.Transfer(temporary, dd, backward)
	if result == nil {
		return nil
	}
	if inverse != nil {
		for level, index := range dd.InvPerm {
			inverse[index] = dd.InvPerm[scratch.InvPerm[level]]
		}
	}
	return result
}

// Extra_bddReorderTest returns the classic CUDD-sifting node count in a separate
// manager, preserving the source benchmark's transfer and reference ownership.
func Extra_bddReorderTest(dd *DdManager, root *DdNode) int {
	if dd == nil {
		return 0
	}
	scratch := cudd.NewBDDManager(len(dd.Vars))
	defer cudd.Cudd_Quit(scratch)
	transferred := cudd.Cudd_bddTransfer(dd, scratch, root)
	if transferred == nil {
		return 0
	}
	cudd.Cudd_Ref(transferred)
	defer cudd.Cudd_RecursiveDeref(scratch, transferred)
	if cudd.Cudd_ReduceHeap(scratch, cudd.CUDD_REORDER_SIFT, 1) == 0 {
		return 0
	}
	return cudd.Cudd_DagSize(transferred)
}

// Extra_addReorderTest follows the source's nonzero-pattern conversion before
// classical CUDD sifting. Numeric leaf distinctions are intentionally collapsed
// for this comparison only, never in the main REO reordering path.
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
	if dd == nil || root == nil || !cudd.Cudd_IsAdd(root) || cudd.Cudd_addPermute(dd, root, identity(len(dd.Vars))) == nil {
		return 0
	}
	scratch := cudd.NewBDDManager(len(dd.Vars))
	defer cudd.Cudd_Quit(scratch)
	transferred := dd.Transfer(pattern(root), scratch, nil)
	if transferred == nil {
		return 0
	}
	cudd.Cudd_Ref(transferred)
	defer cudd.Cudd_RecursiveDeref(scratch, transferred)
	if cudd.Cudd_ReduceHeap(scratch, cudd.CUDD_REORDER_SIFT, 1) == 0 {
		return 0
	}
	return cudd.Cudd_DagSize(transferred)
}
