package reo

// Port of src/bdd/reo/reoTest.c: testing the reordering against CUDD.
import (
	"fmt"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"time"
)

func Extra_ReorderTest(dd *DdManager, Func *DdNode) {
	var pOrder [1000]int
	pReo := Extra_ReorderInit(100, 100)
	Temp := Extra_Reorder(pReo, dd, Func, pOrder[:])
	cudd.Cudd_Ref(Temp)
	Temp1 := Extra_ReorderCudd(dd, Func, nil)
	cudd.Cudd_Ref(Temp1)
	fmt.Printf("Initial = %d. Final = %d. Cudd = %d.\n", cudd.Cudd_DagSize(Func), cudd.Cudd_DagSize(Temp), cudd.Cudd_DagSize(Temp1))
	cudd.Cudd_RecursiveDeref(dd, Temp1)
	cudd.Cudd_RecursiveDeref(dd, Temp)
	Extra_ReorderQuit(pReo)
}

func Extra_ReorderTestArray(dd *DdManager, Funcs []*DdNode, nFuncs int) {
	var FuncsRes [1000]*DdNode
	var pOrder [1000]int
	pReo := Extra_ReorderInit(100, 100)
	Extra_ReorderArray(pReo, dd, Funcs, FuncsRes[:], nFuncs, pOrder[:])
	Extra_ReorderQuit(pReo)
	fmt.Printf("Initial = %d. Final = %d.\n", sharingSize(Funcs[:nFuncs]), sharingSize(FuncsRes[:nFuncs]))
	for i := 0; i < nFuncs; i++ {
		cudd.Cudd_RecursiveDeref(dd, FuncsRes[i])
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
	fmt.Printf("Nodes before = %d.\n", cudd.Cudd_DagSize(temporary))
	if cudd.Cudd_ReduceHeap(scratch, cudd.CUDD_REORDER_SYMM_SIFT, 1) == 0 {
		return nil
	}
	fmt.Printf("Nodes before = %d.\n", cudd.Cudd_DagSize(temporary))
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
	started := time.Now()
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
	before := cudd.Cudd_DagSize(transferred)
	if cudd.Cudd_ReduceHeap(scratch, cudd.CUDD_REORDER_SIFT, 1) == 0 {
		return 0
	}
	after := cudd.Cudd_DagSize(transferred)
	fmt.Printf("Classical reordering of ADDs: Before = %d. After = %d.\n", before, after)
	fmt.Printf("Classical variable reordering time = %.2f sec\n", time.Since(started).Seconds())
	return after
}

func identity(n int) []int {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return p
}

// sharingSize is Cudd_SharingSize.
func sharingSize(Funcs []*DdNode) int {
	var s nodeSet
	s.clear()
	count := 0
	var visit func(n *DdNode)
	visit = func(n *DdNode) {
		n = cudd.Cudd_Regular(n)
		if !s.insert(n) {
			return
		}
		count++
		if !cudd.Cudd_IsConstant(n) {
			visit(n.T)
			visit(n.E)
		}
	}
	for _, f := range Funcs {
		visit(f)
	}
	return count
}
