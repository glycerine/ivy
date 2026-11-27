// Go port of saigStrSim.c (structural matching using simulation) by Alan
// Mishchenko, UC Berkeley.
package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

const SAIG_WORDS = 16

// The simulation info of object Id is sims[Id*SAIG_WORDS:(Id+1)*SAIG_WORDS]
// (C keeps a pointer to it in pObj->pData). The hash table chains (ppNexts,
// ppCands) hold object IDs; -1 stands for NULL.

func strSimInfo(sims []uint32, e aig.Edge) []uint32 {
	id := e.NodeID() * SAIG_WORDS
	return sims[id : id+SAIG_WORDS : id+SAIG_WORDS]
}

var s_SPrimes = [128]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
	1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
	2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
	2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
	3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
	3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
	4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
	4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
	6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
	6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
	7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
	8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

// Saig_StrSimHash computes hash value of the node using its simulation info.
func Saig_StrSimHash(pSims []uint32) uint32 {
	var uHash uint32
	for i := 0; i < SAIG_WORDS; i++ {
		uHash ^= pSims[i] * s_SPrimes[i&0x7F]
	}
	return uHash
}

// Saig_StrSimIsEqual returns 1 if the simulation info is equal.
func Saig_StrSimIsEqual(pSims0, pSims1 []uint32) int32 {
	for i := 0; i < SAIG_WORDS; i++ {
		if pSims0[i] != pSims1[i] {
			return 0
		}
	}
	return 1
}

// Saig_StrSimIsZero returns 1 if simulation info is zero.
func Saig_StrSimIsZero(pSims []uint32) int32 {
	for i := 0; i < SAIG_WORDS; i++ {
		if pSims[i] != 0 {
			return 0
		}
	}
	return 1
}

// Saig_StrSimIsOne returns 1 if simulation info is one.
func Saig_StrSimIsOne(pSims []uint32) int32 {
	for i := 0; i < SAIG_WORDS; i++ {
		if pSims[i] != ^uint32(0) {
			return 0
		}
	}
	return 1
}

// Saig_StrSimAssignRandom assigns random simulation info.
func Saig_StrSimAssignRandom(pSims []uint32) {
	for i := 0; i < SAIG_WORDS; i++ {
		pSims[i] = aig.Aig_ManRandom(0)
	}
}

// Saig_StrSimAssignOne assigns constant 1 simulation info.
func Saig_StrSimAssignOne(pSims []uint32) {
	for i := 0; i < SAIG_WORDS; i++ {
		pSims[i] = ^uint32(0)
	}
}

// Saig_StrSimAssignZeroInit assigns constant 0 to the first frame.
func Saig_StrSimAssignZeroInit(pSims []uint32) {
	pSims[0] = 0
}

// Saig_StrSimulateNode simulates one node (as an AND gate) in frame i.
func Saig_StrSimulateNode(p *aig.Graph, sims []uint32, pObj aig.Edge, i int) {
	_, c0, c1 := p.KindFanins(pObj)
	pSims := strSimInfo(sims, pObj)
	pSims0 := strSimInfo(sims, c0)
	pSims1 := strSimInfo(sims, c1)
	fC0, fC1 := c0.Complemented(), c1.Complemented()
	if fC0 && fC1 {
		pSims[i] = ^(pSims0[i] | pSims1[i])
	} else if fC0 && !fC1 {
		pSims[i] = (^pSims0[i] & pSims1[i])
	} else if !fC0 && fC1 {
		pSims[i] = (pSims0[i] & ^pSims1[i])
	} else {
		pSims[i] = (pSims0[i] & pSims1[i])
	}
}

// Saig_StrSimSaveOutput saves output of one node.
func Saig_StrSimSaveOutput(p *aig.Graph, sims []uint32, pObj aig.Edge, i int) {
	_, c0, _ := p.KindFanins(pObj)
	pSims := strSimInfo(sims, pObj)
	pSims0 := strSimInfo(sims, c0)
	if c0.Complemented() {
		pSims[i] = ^pSims0[i]
	} else {
		pSims[i] = pSims0[i]
	}
}

// Saig_StrSimTransfer transfers simulation output to another node.
func Saig_StrSimTransfer(pSims0, pSims1 []uint32) {
	copy(pSims1[:SAIG_WORDS], pSims0[:SAIG_WORDS])
}

// Saig_StrSimTransferNext transfers frame i of one node to frame i+1 of another.
func Saig_StrSimTransferNext(pSims0, pSims1 []uint32, i int) {
	if i >= SAIG_WORDS-1 {
		panic("saig: Saig_StrSimTransferNext frame out of range")
	}
	pSims1[i+1] = pSims0[i]
}

func strSimIsCiOrNode(p *aig.Graph, e aig.Edge) bool {
	k, _, _ := p.KindFanins(e)
	return k == aig.Input || k == aig.And || k == aig.ExclusiveOr
}

func strSimIsNode(p *aig.Graph, e aig.Edge) bool {
	k, _, _ := p.KindFanins(e)
	return k == aig.And || k == aig.ExclusiveOr
}

// strSimLiveObjs is Aig_ManForEachObj: the live objects in ID order.
func strSimLiveObjs(p *aig.Graph) []aig.Edge { return p.LiveEdges() }

func strSimSimulateMan(p *aig.Graph, sims []uint32, objs []aig.Edge, f int) {
	for _, pObj := range objs {
		if strSimIsNode(p, pObj) && aig.Aig_ObjRepr(p, pObj) == aig.InvalidEdge {
			Saig_StrSimulateNode(p, sims, pObj, f)
		}
	}
	nRegs := Saig_ManRegNum(p)
	for i := int32(0); i < nRegs; i++ {
		Saig_StrSimSaveOutput(p, sims, Saig_ManLi(p, i), f)
	}
	if f < SAIG_WORDS-1 {
		for i := int32(0); i < nRegs; i++ {
			Saig_StrSimTransferNext(strSimInfo(sims, Saig_ManLi(p, i)), strSimInfo(sims, Saig_ManLo(p, i)), f)
		}
	}
}

// Saig_StrSimulateRound performs one round of simulation.
func Saig_StrSimulateRound(p0, p1 *aig.Graph, sims0, sims1 []uint32) {
	objs0, objs1 := strSimLiveObjs(p0), strSimLiveObjs(p1)
	// simulate the nodes
	for _, pObj0 := range objs0 {
		if !strSimIsCiOrNode(p0, pObj0) {
			continue
		}
		pObj1 := aig.Aig_ObjRepr(p0, pObj0)
		if pObj1 == aig.InvalidEdge {
			continue
		}
		if aig.Aig_ObjRepr(p1, pObj1) != pObj0 {
			panic("saig: Aig_ObjRepr(p1, pObj1) == pObj0")
		}
		Saig_StrSimAssignRandom(strSimInfo(sims0, pObj0))
		Saig_StrSimTransfer(strSimInfo(sims0, pObj0), strSimInfo(sims1, pObj1))
	}
	// simulate the timeframes
	for f := 0; f < SAIG_WORDS; f++ {
		// simulate the first AIG
		strSimSimulateMan(p0, sims0, objs0, f)
		// simulate the second AIG
		strSimSimulateMan(p1, sims1, objs1, f)
	}
}

// Saig_StrSimTableLookup checks if the entry exists in the table.
func Saig_StrSimTableLookup(ppTable, ppNexts []int32, nTableSize int32, sims []uint32, pSims []uint32) int32 {
	// find the hash entry
	iEntry := Saig_StrSimHash(pSims) % uint32(nTableSize)
	// check if there are nodes with this signatures
	for pEntry := ppTable[iEntry]; pEntry >= 0; pEntry = ppNexts[pEntry] {
		if Saig_StrSimIsEqual(sims[int(pEntry)*SAIG_WORDS:], pSims) != 0 {
			return pEntry
		}
	}
	return -1
}

// Saig_StrSimTableInsert inserts the entry into the table.
func Saig_StrSimTableInsert(ppTable, ppNexts []int32, nTableSize int32, sims []uint32, pObj int32) {
	// find the hash entry
	iEntry := Saig_StrSimHash(sims[int(pObj)*SAIG_WORDS:]) % uint32(nTableSize)
	// check if there are nodes with this signatures
	if ppTable[iEntry] < 0 {
		ppTable[iEntry] = pObj
	} else {
		ppNexts[pObj] = ppNexts[ppTable[iEntry]]
		ppNexts[ppTable[iEntry]] = pObj
	}
}

func strSimMinusOnes(n int) []int32 {
	s := make([]int32, n)
	for i := range s {
		s[i] = -1
	}
	return s
}

// Saig_StrSimDetectUnique performs one round of matching.
func Saig_StrSimDetectUnique(p0, p1 *aig.Graph, sims0, sims1 []uint32) int32 {
	// allocate the hash table hashing simulation info into nodes
	nTableSize := util.Abc_PrimeCudd(uint32(aig.Aig_ManObjNum(p0) / 2))
	ppTable := strSimMinusOnes(int(nTableSize))
	ppNexts := strSimMinusOnes(int(aig.Aig_ManObjNumMax(p0)))
	ppCands := strSimMinusOnes(int(aig.Aig_ManObjNumMax(p0)))
	fMarkA := make([]bool, aig.Aig_ManObjNumMax(p0))

	// hash nodes of the first AIG
	for _, pObj := range strSimLiveObjs(p0) {
		if !strSimIsCiOrNode(p0, pObj) {
			continue
		}
		if aig.Aig_ObjRepr(p0, pObj) != aig.InvalidEdge {
			continue
		}
		pSims := strSimInfo(sims0, pObj)
		if Saig_StrSimIsZero(pSims) != 0 || Saig_StrSimIsOne(pSims) != 0 {
			continue
		}
		// check if the entry exists
		pEntry := Saig_StrSimTableLookup(ppTable, ppNexts, nTableSize, sims0, pSims)
		if pEntry < 0 { // insert
			Saig_StrSimTableInsert(ppTable, ppNexts, nTableSize, sims0, int32(pObj.NodeID()))
		} else { // mark the entry as not unique
			fMarkA[pEntry] = true
		}
	}

	// hash nodes from the second AIG
	for _, pObj := range strSimLiveObjs(p1) {
		if !strSimIsCiOrNode(p1, pObj) {
			continue
		}
		if aig.Aig_ObjRepr(p1, pObj) != aig.InvalidEdge {
			continue
		}
		pSims := strSimInfo(sims1, pObj)
		if Saig_StrSimIsZero(pSims) != 0 || Saig_StrSimIsOne(pSims) != 0 {
			continue
		}
		// check if the entry exists
		pEntry := Saig_StrSimTableLookup(ppTable, ppNexts, nTableSize, sims0, pSims)
		if pEntry < 0 { // skip
			continue
		}
		// if there is no candidate, label it
		if ppCands[pEntry] < 0 {
			ppCands[pEntry] = int32(pObj.NodeID())
		} else { // mark the entry as not unique
			fMarkA[pEntry] = true
		}
	}

	// create representatives for the unique entries
	var Counter int32
	for i := int32(0); i < nTableSize; i++ {
		for pEntry := ppTable[i]; pEntry >= 0; pEntry = ppNexts[pEntry] {
			if pObj := ppCands[pEntry]; !fMarkA[pEntry] && pObj >= 0 {
				e0, e1 := aig.Aig_ManObj(p0, pEntry), aig.Aig_ManObj(p1, pObj)
				if aig.Aig_ObjType(p0, e0) != aig.Aig_ObjType(p1, e1) {
					continue
				}
				aig.Aig_ObjSetRepr(p0, e0, e1)
				aig.Aig_ObjSetRepr(p1, e1, e0)
				Counter++
			}
		}
	}
	return Counter
}

// Saig_StrSimCountMatchedFlops counts the number of matched flops.
func Saig_StrSimCountMatchedFlops(p *aig.Graph) int32 {
	var Counter int32
	for i := int32(0); i < Saig_ManRegNum(p); i++ {
		if aig.Aig_ObjRepr(p, Saig_ManLo(p, i)) != aig.InvalidEdge {
			Counter++
		}
	}
	return Counter
}

// Saig_StrSimCountMatchedNodes counts the number of matched nodes.
func Saig_StrSimCountMatchedNodes(p *aig.Graph) int32 {
	var Counter int32
	for _, pObj := range strSimLiveObjs(p) {
		if strSimIsNode(p, pObj) && aig.Aig_ObjRepr(p, pObj) != aig.InvalidEdge {
			Counter++
		}
	}
	return Counter
}

// Saig_StrSimPrepareAig starts the representatives and allocates the
// simulation info, returned in place of p->pData2.
func Saig_StrSimPrepareAig(p *aig.Graph) []uint32 {
	aig.Aig_ManReprStart(p, aig.Aig_ManObjNumMax(p))
	// allocate simulation info
	sims := make([]uint32, int(aig.Aig_ManObjNumMax(p))*SAIG_WORDS)
	// set simulation info for constant1 and register outputs
	Saig_StrSimAssignOne(strSimInfo(sims, aig.Aig_ManConst1(p)))
	for i := int32(0); i < Saig_ManRegNum(p); i++ {
		Saig_StrSimAssignZeroInit(strSimInfo(sims, Saig_ManLo(p, i)))
	}
	return sims
}

// Saig_StrSimSetInitMatching matches the constants and the primary inputs.
func Saig_StrSimSetInitMatching(p0, p1 *aig.Graph) {
	pObj0 := aig.Aig_ManConst1(p0)
	pObj1 := aig.Aig_ManConst1(p1)
	aig.Aig_ObjSetRepr(p0, pObj0, pObj1)
	aig.Aig_ObjSetRepr(p1, pObj1, pObj0)
	for i := int32(0); i < Saig_ManPiNum(p0); i++ {
		pObj0 = aig.Aig_ManCi(p0, i)
		pObj1 = aig.Aig_ManCi(p1, i)
		aig.Aig_ObjSetRepr(p0, pObj0, pObj1)
		aig.Aig_ObjSetRepr(p1, pObj1, pObj0)
	}
}

func strSimIsLo(p *aig.Graph, e aig.Edge) bool {
	k, _, _ := p.KindFanins(e)
	return k == aig.Input && aig.Aig_ObjCioId(p, e) >= Saig_ManPiNum(p)
}

func strSimIsPi(p *aig.Graph, e aig.Edge) bool {
	k, _, _ := p.KindFanins(e)
	return k == aig.Input && aig.Aig_ObjCioId(p, e) < Saig_ManPiNum(p)
}

func strSimIsPo(p *aig.Graph, e aig.Edge) bool {
	k, _, _ := p.KindFanins(e)
	return k == aig.Output && aig.Aig_ObjCioId(p, e) < Saig_ManPoNum(p)
}

func strSimIsLi(p *aig.Graph, e aig.Edge) bool {
	k, _, _ := p.KindFanins(e)
	return k == aig.Output && aig.Aig_ObjCioId(p, e) >= Saig_ManPoNum(p)
}

func strSimLoToLi(p *aig.Graph, e aig.Edge) aig.Edge {
	return Saig_ManLi(p, aig.Aig_ObjCioId(p, e)-Saig_ManPiNum(p))
}

func strSimLiToLo(p *aig.Graph, e aig.Edge) aig.Edge {
	return Saig_ManLo(p, aig.Aig_ObjCioId(p, e)-Saig_ManPoNum(p))
}

func strSimPercent(a, b int32) float64 { return 100.0 * float64(a) / float64(b) }

// Saig_StrSimSetFinalMatching removes the matches whose fanins do not match.
func Saig_StrSimSetFinalMatching(p0, p1 *aig.Graph) {
	var CountAll, CountNot int32
	aig.Aig_ManIncrementTravId(p0)
	objs0 := strSimLiveObjs(p0)
	for _, pObj0 := range objs0 {
		pObj1 := aig.Aig_ObjRepr(p0, pObj0)
		if pObj1 == aig.InvalidEdge {
			continue
		}
		CountAll++
		if pObj0 != aig.Aig_ObjRepr(p1, pObj1) {
			panic("saig: pObj0 == Aig_ObjRepr( p1, pObj1 )")
		}
		if strSimIsNode(p0, pObj0) {
			if !strSimIsNode(p1, pObj1) {
				panic("saig: Aig_ObjIsNode(pObj1)")
			}
			pFanin00 := aig.Aig_ObjFanin0(p0, pObj0)
			pFanin01 := aig.Aig_ObjFanin1(p0, pObj0)
			pFanin10 := aig.Aig_ObjFanin0(p1, pObj1)
			pFanin11 := aig.Aig_ObjFanin1(p1, pObj1)
			if aig.Aig_ObjRepr(p0, pFanin00) != pFanin10 ||
				aig.Aig_ObjRepr(p0, pFanin01) != pFanin11 {
				aig.Aig_ObjSetTravIdCurrent(p0, pObj0)
				CountNot++
			}
		} else if strSimIsLo(p0, pObj0) {
			if !strSimIsLo(p1, pObj1) {
				panic("saig: Saig_ObjIsLo(p1, pObj1)")
			}
			pFanin00 := aig.Aig_ObjFanin0(p0, strSimLoToLi(p0, pObj0))
			pFanin10 := aig.Aig_ObjFanin0(p1, strSimLoToLi(p1, pObj1))
			if aig.Aig_ObjRepr(p0, pFanin00) != pFanin10 {
				aig.Aig_ObjSetTravIdCurrent(p0, pObj0)
				CountNot++
			}
		}
	}
	// remove irrelevant matches
	for _, pObj0 := range objs0 {
		pObj1 := aig.Aig_ObjRepr(p0, pObj0)
		if pObj1 == aig.InvalidEdge {
			continue
		}
		if aig.Aig_ObjIsTravIdCurrent(p0, pObj0) {
			aig.Aig_ObjSetRepr(p0, pObj0, aig.InvalidEdge)
			aig.Aig_ObjSetRepr(p1, pObj1, aig.InvalidEdge)
		}
	}
	util.Abc_Print(1, "Total matches = %6d.  Wrong matches = %6d.  Ratio = %5.2f %%\n",
		CountAll, CountNot, strSimPercent(CountNot, CountAll))
}

// strSimForEachFanout is Aig_ObjForEachFanout.
func strSimForEachFanout(p *aig.Graph, pObj aig.Edge, f func(aig.Edge)) {
	nRefs := aig.Aig_ObjRefs(p, pObj)
	iFan := int32(-1)
	for i := int32(0); i < nRefs; i++ {
		if i == 0 {
			iFan = aig.Aig_ObjFanout0Int(p, int32(pObj.NodeID()))
		} else {
			iFan = aig.Aig_ObjFanoutNext(p, iFan)
		}
		f(aig.Aig_ManObj(p, iFan>>1))
	}
}

// Saig_StrSimSetContiguousMatching_rec marks the matched objects reachable
// from pObj through matched objects.
func Saig_StrSimSetContiguousMatching_rec(p *aig.Graph, pObj aig.Edge) {
	if aig.Aig_ObjIsTravIdCurrent(p, pObj) {
		return
	}
	aig.Aig_ObjSetTravIdCurrent(p, pObj)
	if strSimIsPo(p, pObj) {
		return
	}
	if strSimIsLi(p, pObj) {
		Saig_StrSimSetContiguousMatching_rec(p, strSimLiToLo(p, pObj))
		return
	}
	if !strSimIsCiOrNode(p, pObj) {
		panic("saig: Aig_ObjIsCi(pObj) || Aig_ObjIsNode(pObj)")
	}
	if aig.Aig_ObjRepr(p, pObj) == aig.InvalidEdge {
		return
	}
	// go through the fanouts
	strSimForEachFanout(p, pObj, func(pFanout aig.Edge) {
		Saig_StrSimSetContiguousMatching_rec(p, pFanout)
	})
	// go through the fanins
	if !aig.Aig_ObjIsCi(p, pObj) {
		Saig_StrSimSetContiguousMatching_rec(p, aig.Aig_ObjFanin0(p, pObj))
		Saig_StrSimSetContiguousMatching_rec(p, aig.Aig_ObjFanin1(p, pObj))
	}
}

// Saig_StrSimSetContiguousMatching keeps the matches reachable from the PIs
// (unused by Saig_StrSimPerformMatching, as in ABC).
func Saig_StrSimSetContiguousMatching(p0, p1 *aig.Graph) {
	var CountAll, CountNot int32
	// mark nodes reachable through the PIs
	aig.Aig_ManIncrementTravId(p0)
	aig.Aig_ObjSetTravIdCurrent(p0, aig.Aig_ManConst1(p0))
	for i := int32(0); i < Saig_ManPiNum(p0); i++ {
		Saig_StrSimSetContiguousMatching_rec(p0, aig.Aig_ManCi(p0, i))
	}
	// remove irrelevant matches
	for _, pObj0 := range strSimLiveObjs(p0) {
		pObj1 := aig.Aig_ObjRepr(p0, pObj0)
		if pObj1 == aig.InvalidEdge {
			continue
		}
		CountAll++
		if !aig.Aig_ObjIsTravIdCurrent(p0, pObj0) {
			aig.Aig_ObjSetRepr(p0, pObj0, aig.InvalidEdge)
			aig.Aig_ObjSetRepr(p1, pObj1, aig.InvalidEdge)
			CountNot++
		}
	}
	util.Abc_Print(1, "Total matches = %6d.  Wrong matches = %6d.  Ratio = %5.2f %%\n",
		CountAll, CountNot, strSimPercent(CountNot, CountAll))
}

// Ssw_StrSimMatchingExtendOne collects the matched neighbors of the
// unmatched objects.
func Ssw_StrSimMatchingExtendOne(p *aig.Graph, vNodes *vec.Vec_Ptr_t) {
	vec.Vec_PtrClear(vNodes)
	aig.Aig_ManIncrementTravId(p)
	visit := func(pNext aig.Edge) {
		if aig.Aig_ObjRepr(p, pNext) != aig.InvalidEdge && !aig.Aig_ObjIsTravIdCurrent(p, pNext) {
			aig.Aig_ObjSetTravIdCurrent(p, pNext)
			vec.Vec_PtrPush(vNodes, pNext)
		}
	}
	for _, pObj := range strSimLiveObjs(p) {
		if !strSimIsCiOrNode(p, pObj) {
			continue
		}
		if aig.Aig_ObjRepr(p, pObj) != aig.InvalidEdge {
			continue
		}
		if strSimIsLo(p, pObj) {
			pNext := aig.Aig_ObjFanin0(p, strSimLoToLi(p, pObj))
			if aig.Aig_ObjRepr(p, pNext) != aig.InvalidEdge && !aig.Aig_ObjIsTravIdCurrent(p, pNext) && !aig.Aig_ObjIsConst1(p, pNext) {
				aig.Aig_ObjSetTravIdCurrent(p, pNext)
				vec.Vec_PtrPush(vNodes, pNext)
			}
		}
		if strSimIsNode(p, pObj) {
			visit(aig.Aig_ObjFanin0(p, pObj))
			visit(aig.Aig_ObjFanin1(p, pObj))
		}
		strSimForEachFanout(p, pObj, func(pNext aig.Edge) {
			if strSimIsPo(p, pNext) {
				return
			}
			if strSimIsLi(p, pNext) {
				pNext = strSimLiToLo(p, pNext)
			}
			visit(pNext)
		})
	}
}

// Ssw_StrSimMatchingCountUnmached counts the unmatched CIs and nodes.
func Ssw_StrSimMatchingCountUnmached(p *aig.Graph) int32 {
	var Counter int32
	for _, pObj := range strSimLiveObjs(p) {
		if !strSimIsCiOrNode(p, pObj) {
			continue
		}
		if aig.Aig_ObjRepr(p, pObj) != aig.InvalidEdge {
			continue
		}
		Counter++
	}
	return Counter
}

func strSimPrintUnmatched(p0 *aig.Graph, d int32) {
	nUnmached := Ssw_StrSimMatchingCountUnmached(p0)
	nTotal := aig.Aig_ManCiNum(p0) + aig.Aig_ManNodeNum(p0)
	util.Abc_Print(1, "%2d : Total = %6d. Unmatched = %6d.  Ratio = %6.2f %%\n",
		d, nTotal, nUnmached, strSimPercent(nUnmached, nTotal))
}

// Ssw_StrSimMatchingExtend removes the matches next to the unmatched
// objects nDist times.
func Ssw_StrSimMatchingExtend(p0, p1 *aig.Graph, nDist, fVerbose int32) {
	vNodes0 := vec.Vec_PtrAlloc(1000)
	vNodes1 := vec.Vec_PtrAlloc(1000)
	if fVerbose != 0 {
		util.Abc_Print(1, "Extending islands by %d steps:\n", nDist)
		strSimPrintUnmatched(p0, 0)
	}
	for d := int32(0); d < nDist; d++ {
		Ssw_StrSimMatchingExtendOne(p0, vNodes0)
		Ssw_StrSimMatchingExtendOne(p1, vNodes1)
		for k := 0; k < vec.Vec_PtrSize(vNodes0); k++ {
			pNext0 := vec.Vec_PtrEntry(vNodes0, k).(aig.Edge)
			pNext1 := aig.Aig_ObjRepr(p0, pNext0)
			if pNext1 == aig.InvalidEdge {
				continue
			}
			if pNext0 != aig.Aig_ObjRepr(p1, pNext1) {
				panic("saig: pNext0 == Aig_ObjRepr( p1, pNext1 )")
			}
			if strSimIsPi(p1, pNext1) {
				continue
			}
			aig.Aig_ObjSetRepr(p0, pNext0, aig.InvalidEdge)
			aig.Aig_ObjSetRepr(p1, pNext1, aig.InvalidEdge)
		}
		for k := 0; k < vec.Vec_PtrSize(vNodes1); k++ {
			pNext1 := vec.Vec_PtrEntry(vNodes1, k).(aig.Edge)
			pNext0 := aig.Aig_ObjRepr(p1, pNext1)
			if pNext0 == aig.InvalidEdge {
				continue
			}
			if pNext1 != aig.Aig_ObjRepr(p0, pNext0) {
				panic("saig: pNext1 == Aig_ObjRepr( p0, pNext0 )")
			}
			if strSimIsPi(p0, pNext0) {
				continue
			}
			aig.Aig_ObjSetRepr(p0, pNext0, aig.InvalidEdge)
			aig.Aig_ObjSetRepr(p1, pNext1, aig.InvalidEdge)
		}
		if fVerbose != 0 {
			strSimPrintUnmatched(p0, d+1)
		}
	}
}

// Saig_StrSimPerformMatching performs structural matching of two AIGs using
// simulation. With p1 == nil, p0 is a miter to be demitered. The result is
// the array of matched object ID pairs; ppMiter, if given, receives the
// miter of the unmatched windows.
func Saig_StrSimPerformMatching(p0, p1 *aig.Graph, nDist, fVerbose int32, ppMiter **aig.Graph) *vec.Vec_Int_t {
	var pPart0, pPart1 *aig.Graph
	clkTotal := util.Abc_Clock()
	aig.Aig_ManRandom(1)
	// consider the case when a miter is given
	if p1 == nil {
		if fVerbose != 0 {
			aig.Aig_ManPrintStats(p0)
		}
		// demiter the miter
		if Saig_ManDemiterSimpleDiff(p0, &pPart0, &pPart1) == 0 {
			util.Abc_Print(1, "Demitering has failed.\n")
			return nil
		}
	} else {
		pPart0 = aig.Aig_ManDupSimple(p0)
		pPart1 = aig.Aig_ManDupSimple(p1)
	}
	if fVerbose != 0 {
		aig.Aig_ManPrintStats(pPart0)
		aig.Aig_ManPrintStats(pPart1)
	}
	// start simulation
	sims0 := Saig_StrSimPrepareAig(pPart0)
	sims1 := Saig_StrSimPrepareAig(pPart1)
	Saig_StrSimSetInitMatching(pPart0, pPart1)
	if fVerbose != 0 {
		util.Abc_Print(1, "Allocated %6.2f MB to simulate the first AIG.\n",
			1.0*float64(aig.Aig_ManObjNumMax(pPart0))*SAIG_WORDS*4/(1<<20))
		util.Abc_Print(1, "Allocated %6.2f MB to simulate the second AIG.\n",
			1.0*float64(aig.Aig_ManObjNumMax(pPart1))*SAIG_WORDS*4/(1<<20))
	}
	// iterate matching
	nMatches := int32(1)
	for i := int32(0); nMatches > 0; i++ {
		clk := util.Abc_Clock()
		Saig_StrSimulateRound(pPart0, pPart1, sims0, sims1)
		nMatches = Saig_StrSimDetectUnique(pPart0, pPart1, sims0, sims1)
		if fVerbose != 0 {
			nFlops := Saig_StrSimCountMatchedFlops(pPart0)
			nNodes := Saig_StrSimCountMatchedNodes(pPart0)
			util.Abc_Print(1, "%3d : Match =%6d.  FF =%6d. (%6.2f %%)  Node =%6d. (%6.2f %%)  ",
				i, nMatches,
				nFlops, strSimPercent(nFlops, Saig_ManRegNum(pPart0)),
				nNodes, strSimPercent(nNodes, aig.Aig_ManNodeNum(pPart0)))
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
		}
		if i == 20 {
			break
		}
	}
	// extend the islands
	aig.Aig_ManFanoutStart(pPart0)
	aig.Aig_ManFanoutStart(pPart1)
	if nDist != 0 {
		Ssw_StrSimMatchingExtend(pPart0, pPart1, nDist, fVerbose)
	}
	Saig_StrSimSetFinalMatching(pPart0, pPart1)
	//    Saig_StrSimSetContiguousMatching( pPart0, pPart1 );
	// copy the results into array
	vPairs := vec.Vec_IntAlloc(int(2 * aig.Aig_ManObjNumMax(pPart0)))
	for _, pObj0 := range strSimLiveObjs(pPart0) {
		pObj1 := aig.Aig_ObjRepr(pPart0, pObj0)
		if pObj1 == aig.InvalidEdge {
			continue
		}
		if pObj0 != aig.Aig_ObjRepr(pPart1, pObj1) {
			panic("saig: pObj0 == Aig_ObjRepr(pPart1, pObj1)")
		}
		vec.Vec_IntPush(vPairs, int32(pObj0.NodeID()))
		vec.Vec_IntPush(vPairs, int32(pObj1.NodeID()))
	}
	// this procedure adds matching of PO and LI
	if ppMiter != nil {
		*ppMiter = Saig_ManWindowExtractMiter(pPart0, pPart1)
	}
	aig.Aig_ManFanoutStop(pPart0)
	aig.Aig_ManFanoutStop(pPart1)
	aig.Aig_ManStop(pPart0)
	aig.Aig_ManStop(pPart1)
	util.Abc_PrintTime(1, "Total runtime", util.Abc_Clock()-clkTotal)
	return vPairs
}
