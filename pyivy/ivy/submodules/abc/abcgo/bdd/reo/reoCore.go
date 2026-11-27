package reo

// Port of src/bdd/reo/reoCore.c: implementation of the core reordering procedure.

import (
	"fmt"
	"os"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

var s_AplBefore, s_AplAfter float64

func reoReorderArray(p *reo_man, dd *DdManager, Funcs []*DdNode, FuncsRes []*DdNode, nFuncs int, pOrder []int) {
	var Counter, i int

	// set the initial parameters
	p.dd = dd
	p.pOrder = pOrder
	p.nTops = nFuncs
	// get the initial number of nodes (Cudd_SharingSize) and
	// compute the support (Extra_VectorSupportArray) in one traversal;
	// the support array is sized for the manager first
	reoResizeStructures(p, max(len(dd.Vars), dd.ZddSize), 0, nFuncs)
	p.nNodesBeg = reoSharingSupport(p, dd, Funcs[:nFuncs])
	// resize the internal data structures of the manager if necessary
	reoResizeStructures(p, max(len(dd.Vars), dd.ZddSize), p.nNodesBeg, nFuncs)
	// get the number of support variables
	p.nSupp = 0
	for i = 0; i < len(dd.Vars); i++ {
		p.nSupp += p.pSupp[i]
	}

	// if it is the constant function, no need to reorder
	if p.nSupp == 0 {
		for i = 0; i < nFuncs; i++ {
			FuncsRes[i] = Funcs[i]
			cudd.Cudd_Ref(FuncsRes[i])
		}
		return
	}

	// create the internal variable maps
	// go through variable levels in the manager
	Counter = 0
	for i = 0; i < len(dd.Vars); i++ {
		if p.pSupp[dd.InvPerm[i]] != 0 {
			p.pMapToPlanes[dd.InvPerm[i]] = Counter
			p.pMapToDdVarsOrig[Counter] = dd.InvPerm[i]
			if p.fRemapUp == 0 {
				p.pMapToDdVarsFinal[Counter] = dd.InvPerm[i]
			} else {
				p.pMapToDdVarsFinal[Counter] = dd.InvPerm[Counter]
			}
			p.pOrderInt[Counter] = Counter
			Counter++
		}
	}

	// set the initial parameters
	p.nUnitsUsed = 0
	p.nNodesCur = 0
	p.fThisIsAdd = 0
	p.Signature++
	// transfer the function from the CUDD package into REO"s internal data structure
	for i = 0; i < nFuncs; i++ {
		p.pTops[i] = reoTransferNodesToUnits_rec(p, Funcs[i])
	}
	assert(p.nNodesBeg == p.nNodesCur, "p->nNodesBeg == p->nNodesCur")

	if p.fThisIsAdd == 0 && p.fMinWidth != 0 {
		fmt.Printf("An important message from the REO reordering engine:\n")
		fmt.Printf("The BDD given to the engine for reordering contains complemented edges.\n")
		fmt.Printf("Currently, such BDDs cannot be reordered for the minimum width.\n")
		fmt.Printf("Therefore, minimization for the number of BDD nodes is performed.\n")
		os.Stdout.Sync()
		p.fMinApl = 0
		p.fMinWidth = 0
	}

	if p.fMinWidth != 0 {
		reoProfileWidthStart(p)
	} else if p.fMinApl != 0 {
		reoProfileAplStart(p)
	} else {
		reoProfileNodesStart(p)
	}

	if p.fVerbose != 0 {
		fmt.Printf("INITIAL:\n")
		if p.fMinWidth != 0 {
			reoProfileWidthPrint(p)
		} else if p.fMinApl != 0 {
			reoProfileAplPrint(p)
		} else {
			reoProfileNodesPrint(p)
		}
	}

	///////////////////////////////////////////////////////////////////
	// performs the reordering
	p.nSwaps = 0
	p.nNISwaps = 0
	for i = 0; i < p.nIters; i++ {
		reoReorderSift(p)
		// print statistics after each iteration
		if p.fVerbose != 0 {
			fmt.Printf("ITER #%d:\n", i+1)
			if p.fMinWidth != 0 {
				reoProfileWidthPrint(p)
			} else if p.fMinApl != 0 {
				reoProfileAplPrint(p)
			} else {
				reoProfileNodesPrint(p)
			}
		}
		// if the cost function did not change, stop iterating
		if p.fMinWidth != 0 {
			p.nWidthEnd = p.nWidthCur
			if reoDebug {
				assert(p.nWidthEnd <= p.nWidthBeg, "p->nWidthEnd <= p->nWidthBeg")
			}
			if p.nWidthEnd == p.nWidthBeg {
				break
			}
		} else if p.fMinApl != 0 {
			p.nAplEnd = p.nAplCur
			if reoDebug {
				assert(p.nAplEnd <= p.nAplBeg, "p->nAplEnd <= p->nAplBeg")
			}
			if p.nAplEnd == p.nAplBeg {
				break
			}
		} else {
			p.nNodesEnd = p.nNodesCur
			if reoDebug {
				assert(p.nNodesEnd <= p.nNodesBeg, "p->nNodesEnd <= p->nNodesBeg")
			}
			if p.nNodesEnd == p.nNodesBeg {
				break
			}
		}
	}
	if reoDebug { // C: assert( reoCheckLevels( p ) ), compiled out under NDEBUG
		assert(reoCheckLevels(p), "reoCheckLevels( p )")
	}
	///////////////////////////////////////////////////////////////////

	s_AplBefore = p.nAplBeg
	s_AplAfter = p.nAplEnd

	// set the initial parameters
	p.nRefNodes = 0
	p.nNodesCur = 0
	p.Signature++
	// transfer the BDDs from REO's internal data structure to CUDD
	for i = 0; i < nFuncs; i++ {
		FuncsRes[i] = reoTransferUnitsToNodes_rec(p, p.pTops[i])
		cudd.Cudd_Ref(FuncsRes[i])
	}
	// undo the DDs referenced for storing in the cache
	for i = 0; i < p.nRefNodes; i++ {
		cudd.Cudd_RecursiveDeref(dd, p.pRefNodes[i])
		p.pRefNodes[i] = nil
	}
	// verify zero refs of the terminal nodes
	if reoDebug { // C asserts these, compiled out under NDEBUG
		for i = 0; i < nFuncs; i++ {
			assert(reoRecursiveDeref(p, p.pTops[i]), "reoRecursiveDeref( p->pTops[i] )")
		}
		assert(reoCheckZeroRefs(p, &p.pPlanes[p.nSupp]), "reoCheckZeroRefs( &(p->pPlanes[p->nSupp]) )")
	}

	// prepare the variable map to return to the user
	if p.pOrder != nil {
		// i is the current level in the planes data structure
		// p->pOrderInt[i] is the original level in the planes data structure
		// p->pMapToDdVarsOrig[i] is the variable, into which we remap when we construct the BDD from planes
		// p->pMapToDdVarsOrig[ p->pOrderInt[i] ] is the original BDD variable corresponding to this level
		// Therefore, p->pOrder[ p->pMapToDdVarsFinal[i] ] = p->pMapToDdVarsOrig[ p->pOrderInt[i] ]
		// creates the permutation, which remaps the resulting BDD variable into the original BDD variable
		for i = 0; i < p.nSupp; i++ {
			p.pOrder[p.pMapToDdVarsFinal[i]] = p.pMapToDdVarsOrig[p.pOrderInt[i]]
		}
	}

	if p.fVerify != 0 {
		var fVerification int
		var FuncRemapped *DdNode
		// C indexes pOrder by variable but allocates nSupp entries when the
		// caller passed none; here the array covers every variable, and the
		// entries outside the support (not read by the permutation) are identity.
		pOrder := make([]int, len(dd.Vars))
		for i = range pOrder {
			pOrder[i] = i
		}
		for i = 0; i < p.nSupp; i++ {
			pOrder[p.pMapToDdVarsFinal[i]] = p.pMapToDdVarsOrig[p.pOrderInt[i]]
		}

		fVerification = 1
		for i = 0; i < nFuncs; i++ {
			// verify the result
			if p.fThisIsAdd != 0 && cudd.Cudd_IsAdd(FuncsRes[i]) {
				FuncRemapped = cudd.Cudd_addPermute(dd, FuncsRes[i], pOrder)
			} else {
				FuncRemapped = cudd.Cudd_bddPermute(dd, FuncsRes[i], pOrder)
			}
			cudd.Cudd_Ref(FuncRemapped)

			if FuncRemapped != Funcs[i] {
				fVerification = 0
				fmt.Printf("REO: Internal verification has failed!\n")
				os.Stdout.Sync()
			}
			cudd.Cudd_RecursiveDeref(dd, FuncRemapped)
		}
		if fVerification != 0 {
			fmt.Printf("REO: Internal verification is okay!\n")
		}
	}

	// recycle the data structure
	for i = 0; i <= p.nSupp; i++ {
		reoUnitsRecycleUnitList(p, &p.pPlanes[i])
	}
}

// reoSharingSupport returns Cudd_SharingSize(Funcs), the number of
// distinct nodes (constants included), and marks in p.pSupp the variables
// on which Funcs depend, as Extra_VectorSupportArray.
func reoSharingSupport(p *reo_man, dd *DdManager, Funcs []*DdNode) int {
	size := max(len(dd.Vars), dd.ZddSize)
	for i := 0; i < size; i++ {
		p.pSupp[i] = 0
	}
	p.visited.clear()
	count := 0
	var visit func(n *DdNode)
	visit = func(n *DdNode) {
		for {
			n = cudd.Cudd_Regular(n)
			if !p.visited.insert(n) {
				return
			}
			count++
			if cudd.Cudd_IsConstant(n) {
				return
			}
			p.pSupp[n.Index] = 1
			visit(n.T)
			n = n.E
		}
	}
	for _, f := range Funcs {
		visit(f)
	}
	return count
}

// nodeSet is the traversal mark of ddSupportStep2/ddClearFlag2: an open
// addressing set of nodes, emptied by advancing the stamp.
type nodeSet struct {
	stamp uint32
	count int
	slots []nodeSlot
}
type nodeSlot struct {
	stamp uint32
	node  *DdNode
}

func (s *nodeSet) clear() {
	s.stamp++
	s.count = 0
	if s.stamp == 0 || s.slots == nil {
		s.slots = make([]nodeSlot, max(len(s.slots), 256))
		s.stamp = 1
	}
}

// insert adds n and reports whether it was absent.
func (s *nodeSet) insert(n *DdNode) bool {
	if 2*(s.count+1) > len(s.slots) {
		old := s.slots
		s.slots = make([]nodeSlot, 2*len(old))
		for _, e := range old {
			if e.stamp == s.stamp {
				s.place(e.node)
			}
		}
	}
	mask := uint64(len(s.slots) - 1)
	for h := (n.ID * 0x9E3779B97F4A7C15) >> 20 & mask; ; h = (h + 1) & mask {
		e := &s.slots[h]
		if e.stamp != s.stamp {
			e.stamp, e.node = s.stamp, n
			s.count++
			return true
		}
		if e.node == n {
			return false
		}
	}
}

func (s *nodeSet) place(n *DdNode) {
	mask := uint64(len(s.slots) - 1)
	for h := (n.ID * 0x9E3779B97F4A7C15) >> 20 & mask; ; h = (h + 1) & mask {
		if s.slots[h].stamp != s.stamp {
			s.slots[h] = nodeSlot{s.stamp, n}
			return
		}
	}
}

// reoResizeStructures resizes the internal manager data structures.
func reoResizeStructures(p *reo_man, nDdVarsMax, nNodesMax, nFuncs int) {
	// resize data structures depending on the number of variables in the DD manager
	if p.nSuppAlloc == 0 || p.nSuppAlloc < nDdVarsMax {
		p.pSupp = make([]int, nDdVarsMax+1)
		p.pOrderInt = make([]int, nDdVarsMax+1)
		p.pMapToPlanes = make([]int, nDdVarsMax+1)
		p.pMapToDdVarsOrig = make([]int, nDdVarsMax+1)
		p.pMapToDdVarsFinal = make([]int, nDdVarsMax+1)
		p.pPlanes = make([]reo_plane, nDdVarsMax+1)
		p.pVarCosts = make([]float64, nDdVarsMax+1)
		p.pLevelOrder = make([]int, nDdVarsMax+1)
		p.nSuppAlloc = nDdVarsMax + 1
	}

	// resize the data structures depending on the number of nodes
	if p.nRefNodesAlloc == 0 || p.nNodesMaxAlloc < nNodesMax {
		p.nNodesMaxAlloc = nNodesMax
		p.nTableSize = 1
		p.tableShift = 64
		for p.nTableSize < 3*nNodesMax+1 {
			p.nTableSize <<= 1
			p.tableShift--
		}
		p.tableMask = p.nTableSize - 1
		p.nRefNodesAlloc = 3*nNodesMax + 1
		p.nMemChunksAlloc = (10*nNodesMax+1)/REO_CHUNK_SIZE + 1

		p.HTable = make([]reo_hash, p.nTableSize)
		p.pRefNodes = make([]*DdNode, p.nRefNodesAlloc)
		p.pWidthCofs = make([]uref, p.nRefNodesAlloc)
		// the unit chunks (p->pMemChunks) hold units currently in use and stay
	}

	// resize the data structures depending on the number of functions
	if p.nTopsAlloc == 0 || p.nTopsAlloc < nFuncs {
		p.pTops = make([]uref, nFuncs)
		p.nTopsAlloc = nFuncs
	}
}

// reoRecursiveDeref dereferences units the data structure after reordering.
func reoRecursiveDeref(p *reo_man, pUnit uref) bool {
	pUnitR := p.U(Unit_Regular(pUnit))
	pUnitR.n--
	if Unit_IsConstant(pUnitR) {
		return true
	}
	if pUnitR.n == 0 {
		reoRecursiveDeref(p, pUnitR.pE)
		reoRecursiveDeref(p, pUnitR.pT)
	}
	return true
}

// reoCheckZeroRefs checks the zero references for the given plane.
func reoCheckZeroRefs(p *reo_man, pPlane *reo_plane) bool {
	for pUnit := pPlane.pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
		if p.U(pUnit).n != 0 {
			assert(false, "pUnit->n == 0")
		}
	}
	return true
}

// reoCheckLevels checks that each level has nodes with the proper level.
func reoCheckLevels(p *reo_man) bool {
	for i := 0; i < p.nSupp; i++ {
		// there are some nodes left on each level
		assert(p.pPlanes[i].statsNodes != 0, "p->pPlanes[i].statsNodes")
		for pUnit := p.pPlanes[i].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
			// the level is properly set
			assert(int(p.U(pUnit).lev) == i, "pUnit->lev == i")
		}
	}
	return true
}
