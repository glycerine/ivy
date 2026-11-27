package reo

// Port of src/bdd/reo/reoShuffle.c: implementation of the permutation of
// variables in the BDD by adjacent swaps.

import (
	"fmt"
	"os"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
	"github.com/glycerine/abcgo/misc/util"
)

// reoShuffle is similar to Cudd_ShuffleHeap() and Cudd_bddPermute(). The
// i-th entry of pPerm contains the index of the variable that should be
// brought to the i-th level. The resulting BDD is not referenced.
func reoShuffle(p *reo_man, dd *DdManager, bFunc *DdNode, pPerm, pPermInv []int) *DdNode {
	var bFuncRes *DdNode
	var i, k, v int

	if cudd.Cudd_IsConstant(bFunc) {
		return bFunc
	}

	// set the initial parameters
	p.dd = dd
	p.nSupp = cudd.Cudd_SupportSize(dd, bFunc)
	p.nTops = 1

	// set the starting permutation
	for i = 0; i < p.nSupp; i++ {
		p.pOrderInt[i] = i
		p.pMapToPlanes[dd.InvPerm[i]] = i
		p.pMapToDdVarsFinal[i] = dd.InvPerm[i]
	}

	// set the initial parameters
	p.nUnitsUsed = 0
	p.nNodesCur = 0
	p.fThisIsAdd = 0
	p.Signature++

	// transfer the function from the CUDD package into REO's internal data structure
	p.pTops[0] = reoTransferNodesToUnits_rec(p, bFunc)

	// reorder one variable at a time
	for i = 0; i < p.nSupp; i++ {
		if p.pOrderInt[i] == pPerm[i] {
			continue
		}
		// find where is variable number pPerm[i]
		for k = i + 1; k < p.nSupp; k++ {
			if pPerm[i] == p.pOrderInt[k] {
				break
			}
		}
		if k == p.nSupp {
			fmt.Printf("reoShuffle() Error: Cannot find a variable.\n")
			goto finish
		}
		// move the variable up
		for v = k - 1; v >= i; v-- {
			reoReorderSwapAdjacentVars(p, v, 1)
			// check if the number of nodes is not too large
			if p.nNodesCur > 10000 {
				fmt.Printf("reoShuffle() Error: BDD size is too large.\n")
				goto finish
			}
		}
		assert(p.pOrderInt[i] == pPerm[i], "p->pOrderInt[i] == pPerm[i]")
	}

	// set the initial parameters
	p.nRefNodes = 0
	p.nNodesCur = 0
	p.Signature++
	// transfer the BDDs from REO's internal data structure to CUDD
	bFuncRes = reoTransferUnitsToNodes_rec(p, p.pTops[0])
	cudd.Cudd_Ref(bFuncRes)
	// undo the DDs referenced for storing in the cache
	for i = 0; i < p.nRefNodes; i++ {
		cudd.Cudd_RecursiveDeref(dd, p.pRefNodes[i])
		p.pRefNodes[i] = nil
	}

	// perform verification
	if p.fVerify != 0 {
		bFuncPerm := cudd.Cudd_bddPermute(dd, bFunc, pPermInv)
		cudd.Cudd_Ref(bFuncPerm)
		if bFuncPerm != bFuncRes {
			fmt.Printf("REO: Internal verification has failed!\n")
			os.Stdout.Sync()
		}
		cudd.Cudd_RecursiveDeref(dd, bFuncPerm)
	}

	// recycle the data structure
	for i = 0; i <= p.nSupp; i++ {
		reoUnitsRecycleUnitList(p, &p.pPlanes[i])
	}

finish:
	if bFuncRes != nil {
		cudd.Cudd_Deref(bFuncRes)
	}
	return bFuncRes
}

// Extra_ShuffleTest reorders the DD using REO and CUDD; useful to test the
// performance of the reordering package.
func Extra_ShuffleTest(pReo *reo_man, dd *DdManager, Func *DdNode) {
	var pOrder, pOrderInv [1000]int
	assert(len(dd.Vars) < 1000, "dd->size < 1000")

	util.C_Srand(0x12341234)

	nSuppSize := cudd.Cudd_SupportSize(dd, Func)
	if nSuppSize < 2 {
		return
	}

	for i := 0; i < nSuppSize; i++ {
		pOrder[i] = i
	}
	for i := 0; i < 120; i++ {
		OffSet := int(util.C_Rand() % uint32(nSuppSize-1))
		Num := pOrder[OffSet]
		pOrder[OffSet] = pOrder[OffSet+1]
		pOrder[OffSet+1] = Num
	}
	for i := 0; i < nSuppSize; i++ {
		pOrderInv[pOrder[i]] = i
	}

	// create permutation
	bRemap := extrab.Extra_bddRemapUp(dd, Func)
	cudd.Cudd_Ref(bRemap)
	Temp := reoShuffle(pReo, dd, bRemap, pOrder[:], pOrderInv[:])
	cudd.Cudd_Ref(Temp)
	{
		bFuncPerm := cudd.Cudd_bddPermute(dd, bRemap, pOrderInv[:])
		cudd.Cudd_Ref(bFuncPerm)
		if bFuncPerm != Temp {
			fmt.Printf("REO: Internal verification has failed!\n")
			os.Stdout.Sync()
		}
		cudd.Cudd_RecursiveDeref(dd, bFuncPerm)
	}
	cudd.Cudd_RecursiveDeref(dd, Temp)
	cudd.Cudd_RecursiveDeref(dd, bRemap)
}
