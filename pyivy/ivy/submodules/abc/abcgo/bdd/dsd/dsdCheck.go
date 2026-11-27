package dsd

import (
	"sync"
	"unsafe"

	bdd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
)

// dsdCheck.c: procedures to check the identity of root functions.

type dsdEntry struct {
	bX  [4]*DdNode
	res int
}

type dsdCache struct {
	pTable     []dsdEntry
	nTableSize int
	nSuccess   int
	nFailure   int
}

// dsdCheckCacheAllocate is Dsd_CheckCacheAllocate for one manager: C keeps
// the cache in a static pointer that Dsd_ManagerStart reallocates; each Go
// manager owns its cache so independent managers may run concurrently.
func dsdCheckCacheAllocate(nEntries int) *dsdCache {
	p := &dsdCache{}
	nRequested := int(util.Abc_PrimeCudd(uint32(nEntries)))
	p.nTableSize = nRequested
	p.pTable = make([]dsdEntry, nRequested)
	return p
}

// the cache used by the manager-less API below
var pCache struct {
	sync.Mutex
	c *dsdCache
}

// Dsd_CheckCacheAllocate (re)allocates the local cache.
func Dsd_CheckCacheAllocate(nEntries int) {
	pCache.Lock()
	defer pCache.Unlock()
	pCache.c = dsdCheckCacheAllocate(nEntries)
}

// Dsd_CheckCacheDeallocate deallocates the local cache.
func Dsd_CheckCacheDeallocate() {
	pCache.Lock()
	defer pCache.Unlock()
	pCache.c = nil
}

// Dsd_CheckCacheClear clears the local cache. (C's loop clears entry 0
// nTableSize times; the table is freshly allocated, so this is moot.)
func Dsd_CheckCacheClear() {
	pCache.Lock()
	defer pCache.Unlock()
	if pCache.c != nil {
		clear(pCache.c.pTable)
	}
}

// Dsd_CheckRootFunctionIdentity checks whether it is true that
// bF1(bC1=0) == bF2(bC2=0).
func Dsd_CheckRootFunctionIdentity(dd *DdManager, bF1, bF2, bC1, bC2 *DdNode) int {
	pCache.Lock()
	defer pCache.Unlock()
	if pCache.c == nil {
		pCache.c = dsdCheckCacheAllocate(5000)
	}
	return dsdCheckRootFunctionIdentity_rec(dd, pCache.c, bF1, bF2, bC1, bC2)
}

func (m *Dsd_Manager_t) checkRootFunctionIdentity(bF1, bF2, bC1, bC2 *DdNode) int {
	return dsdCheckRootFunctionIdentity_rec(m.dd, m.cache, bF1, bF2, bC1, bC2)
}

func hashKey4(a, b, c, d *DdNode, tsize int) int {
	// extraBdd.h's hashKey4 on the node addresses (Go heap objects do not move)
	h := (uint64(uintptr(unsafe.Pointer(a))) + uint64(uintptr(unsafe.Pointer(b)))) * 12582917
	h = (h + uint64(uintptr(unsafe.Pointer(c)))) * 4256249
	h = (h + uint64(uintptr(unsafe.Pointer(d)))) * 741457
	return int(h % uint64(tsize))
}

// cuddI is the level of an index (CUDD_CONST_INDEX for constants).
func cuddI(dd *DdManager, index uint32) int {
	if index == bdd.CUDD_CONST_INDEX {
		return int(index)
	}
	return dd.Perm[index]
}

// dsdCheckRootFunctionIdentity_rec performs the recursive step of
// Dsd_CheckRootFunctionIdentity().
func dsdCheckRootFunctionIdentity_rec(dd *DdManager, pCache *dsdCache, bF1, bF2, bC1, bC2 *DdNode) int {
	b0, b1 := dd.LogicZero, dd.One
	// if either bC1 or bC2 is zero, the test is true
	if bC1 == b0 || bC2 == b0 {
		panic("dsd: Dsd_CheckRootFunctionIdentity_rec: zero condition")
	}

	// if both bC1 and bC2 are one - perform comparison
	if bC1 == b1 && bC2 == b1 {
		return b2i(bF1 == bF2)
	}

	if bF1 == b0 {
		return b2i(bdd.Cudd_bddLeq(dd, bC2, bdd.Cudd_Not(bF2)))
	}

	if bF1 == b1 {
		return b2i(bdd.Cudd_bddLeq(dd, bC2, bF2))
	}

	if bF2 == b0 {
		return b2i(bdd.Cudd_bddLeq(dd, bC1, bdd.Cudd_Not(bF1)))
	}

	if bF2 == b1 {
		return b2i(bdd.Cudd_bddLeq(dd, bC1, bF1))
	}

	// otherwise, keep expanding

	// check cache
	HKey := hashKey4(bF1, bF2, bC1, bC2, pCache.nTableSize)
	e := &pCache.pTable[HKey]
	if e.bX[0] == bF1 &&
		e.bX[1] == bF2 &&
		e.bX[2] == bC1 &&
		e.bX[3] == bC2 {
		pCache.nSuccess++
		return e.res // the last bit records the result (yes/no)
	}

	// determine the top variables
	var RetValue int
	bA := [4]*DdNode{bF1, bF2, bC1, bC2}                                                                          // arguments
	bAR := [4]*DdNode{bdd.Cudd_Regular(bF1), bdd.Cudd_Regular(bF2), bdd.Cudd_Regular(bC1), bdd.Cudd_Regular(bC2)} // regular arguments
	CurLevel := [4]int{cuddI(dd, bAR[0].Index), cuddI(dd, bAR[1].Index), cuddI(dd, bAR[2].Index), cuddI(dd, bAR[3].Index)}
	TopLevel := int(bdd.CUDD_CONST_INDEX)
	var bE, bT [4]*DdNode
	var bF1next, bF2next, bC1next, bC2next *DdNode

	pCache.nFailure++

	// determine the top level
	for i := 0; i < 4; i++ {
		if TopLevel > CurLevel[i] {
			TopLevel = CurLevel[i]
		}
	}

	// compute the cofactors
	for i := 0; i < 4; i++ {
		if TopLevel == CurLevel[i] {
			if bA[i] != bAR[i] { // complemented
				bE[i] = bdd.Cudd_Not(bAR[i].E)
				bT[i] = bdd.Cudd_Not(bAR[i].T)
			} else {
				bE[i] = bAR[i].E
				bT[i] = bAR[i].T
			}
		} else {
			bE[i] = bA[i]
			bT[i] = bA[i]
		}
	}

	// solve subproblems
	// three cases are possible

	// (1) the top var belongs to both C1 and C2
	//     in this case, any cofactor of F1 and F2 will do,
	//     as long as the corresponding cofactor of C1 and C2 is not equal to 0
	if TopLevel == CurLevel[2] && TopLevel == CurLevel[3] {
		if bE[2] != b0 { // C1
			bF1next = bE[0]
			bC1next = bE[2]
		} else {
			bF1next = bT[0]
			bC1next = bT[2]
		}
		if bE[3] != b0 { // C2
			bF2next = bE[1]
			bC2next = bE[3]
		} else {
			bF2next = bT[1]
			bC2next = bT[3]
		}
		RetValue = dsdCheckRootFunctionIdentity_rec(dd, pCache, bF1next, bF2next, bC1next, bC2next)
	} else if TopLevel == CurLevel[2] && TopLevel != CurLevel[3] {
		// (2) the top var belongs to either C1 or C2
		//     in this case normal splitting of cofactors
		if bE[2] != b0 { // C1
			bF1next = bE[0]
			bC1next = bE[2]
		} else {
			bF1next = bT[0]
			bC1next = bT[2]
		}
		// split around this variable
		RetValue = dsdCheckRootFunctionIdentity_rec(dd, pCache, bF1next, bE[1], bC1next, bE[3])
		if RetValue == 1 { // test another branch; otherwise, there is no need to test
			RetValue = dsdCheckRootFunctionIdentity_rec(dd, pCache, bF1next, bT[1], bC1next, bT[3])
		}
	} else if TopLevel != CurLevel[2] && TopLevel == CurLevel[3] {
		if bE[3] != b0 { // C2
			bF2next = bE[1]
			bC2next = bE[3]
		} else {
			bF2next = bT[1]
			bC2next = bT[3]
		}
		// split around this variable
		RetValue = dsdCheckRootFunctionIdentity_rec(dd, pCache, bE[0], bF2next, bE[2], bC2next)
		if RetValue == 1 { // test another branch; otherwise, there is no need to test
			RetValue = dsdCheckRootFunctionIdentity_rec(dd, pCache, bT[0], bF2next, bT[2], bC2next)
		}
	} else { // if ( TopLevel != CurLevel[2] && TopLevel != CurLevel[3] )
		// (3) the top var does not belong to C1 and C2
		//     in this case normal splitting of cofactors
		// split around this variable
		RetValue = dsdCheckRootFunctionIdentity_rec(dd, pCache, bE[0], bE[1], bE[2], bE[3])
		if RetValue == 1 { // test another branch; otherwise, there is no need to test
			RetValue = dsdCheckRootFunctionIdentity_rec(dd, pCache, bT[0], bT[1], bT[2], bT[3])
		}
	}

	// set cache (the recursion may have reused the slot)
	e = &pCache.pTable[HKey]
	for i := 0; i < 4; i++ {
		e.bX[i] = bA[i]
	}
	e.res = RetValue

	return RetValue
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
