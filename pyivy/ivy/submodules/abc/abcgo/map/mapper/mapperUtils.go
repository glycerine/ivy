package mapper

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/glycerine/abcgo/misc/st"
)

/**CFile****************************************************************

  FileName    [mapperUtils.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperUtils.c,v 1.8 2004/11/03 22:41:45 satrajit Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const MAP_CO_LIST_SIZE = 5

// static int   Map_MappingCountLevels_rec( Map_Node_t * pNode );

// static float Map_MappingSetRefsAndArea_rec( Map_Man_t * pMan, Map_Node_t * pNode );

// static float Map_MappingSetRefsAndSwitch_rec( Map_Man_t * pMan, Map_Node_t * pNode );

// static float Map_MappingSetRefsAndWire_rec( Map_Man_t * pMan, Map_Node_t * pNode );

// static void  Map_MappingDfsCuts_rec( Map_Node_t * pNode, Map_NodeVec_t * vNodes );

// static float Map_MappingArea_rec( Map_Man_t * pMan, Map_Node_t * pNode, Map_NodeVec_t * vNodes );

// static int   Map_MappingCompareOutputDelay( Map_Node_t ** ppNode1, Map_Node_t ** ppNode2 );

// static void  Map_MappingFindLatest( Map_Man_t * p, int * pNodes, int nNodesMax );

// static unsigned Map_MappingExpandTruth_rec( unsigned uTruth, int nVars );

// static int Map_MappingCountUsedNodes( Map_Man_t * pMan, int fChoices );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Computes the DFS ordering of the nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingDfs_rec(pNode *Map_Node_t, vNodes *Map_NodeVec_t, fCollectEquiv int32) {
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if pNode.fMark0 != 0 {
		return
	}
	// visit the transitive fanin
	if Map_NodeIsAnd(pNode) != 0 {
		// inlining of macro Map_Regular
		Map_MappingDfs_rec(Map_Regular(pNode.p1), vNodes, fCollectEquiv)
		// inlining of macro Map_Regular
		Map_MappingDfs_rec(Map_Regular(pNode.p2), vNodes, fCollectEquiv)
	}
	// visit the equivalent nodes
	if (fCollectEquiv != 0) && pNode.pNextE != nil {
		Map_MappingDfs_rec(pNode.pNextE, vNodes, fCollectEquiv)
	}
	// make sure the node is not visited through the equivalent nodes
	assert(pNode.fMark0 == 0, "pNode.fMark0 == 0")
	// mark the node as visited
	pNode.fMark0 = 1
	// add the node to the list
	Map_NodeVecPush(vNodes, pNode)
}
func Map_MappingDfs(pMan *Map_Man_t, fCollectEquiv int32) *Map_NodeVec_t {
	var vNodes *Map_NodeVec_t
	var i int32
	// perform the traversal
	vNodes = Map_NodeVecAlloc(100)
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_Regular
		Map_MappingDfs_rec(Map_Regular(pMan.pOutputs[i]), vNodes, fCollectEquiv)
	}
	for i = 0; i < vNodes.nSize; i++ {
		vNodes.pArray[i].fMark0 = 0
	}
	//    for ( i = 0; i < pMan->nOutputs; i++ )
	//        Map_MappingUnmark_rec( Map_Regular(pMan->pOutputs[i]) );
	return vNodes
}

/*
*Function*************************************************************

	Synopsis    [Computes the number of logic levels not counting PIs/POs.]

	Description []

	SideEffects [Note that this procedure will reassign the levels assigned
	originally by NodeCreate() because it counts the number of levels with
	choices differently!]

	SeeAlso     []

**********************************************************************
*/
func Map_MappingCountLevels(pMan *Map_Man_t) int32 {
	var i, LevelsMax, LevelsCur int32
	// perform the traversal
	LevelsMax = -1
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_Regular
		LevelsCur = Map_MappingCountLevels_rec(Map_Regular(pMan.pOutputs[i]))
		if LevelsMax < LevelsCur {
			LevelsMax = LevelsCur
		}
	}
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_Regular
		Map_MappingUnmark_rec(Map_Regular(pMan.pOutputs[i]))
	}
	return LevelsMax
}

/*
*Function*************************************************************

	Synopsis    [Recursively computes the number of logic levels.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingCountLevels_rec(pNode *Map_Node_t) int32 {
	var Level1, Level2 int32
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if Map_NodeIsAnd(pNode) == 0 {
		pNode.Level = 0
		return 0
	}
	if pNode.fMark0 != 0 {
		return int32(pNode.Level)
	}
	pNode.fMark0 = 1
	// visit the transitive fanin
	// inlining of macro Map_Regular
	Level1 = Map_MappingCountLevels_rec(Map_Regular(pNode.p1))
	// inlining of macro Map_Regular
	Level2 = Map_MappingCountLevels_rec(Map_Regular(pNode.p2))
	// set the number of levels
	pNode.Level = uint32(1 + max(Level1, Level2))
	return int32(pNode.Level)
}

/*
*Function*************************************************************

	Synopsis    [Unmarks the nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingUnmark(pMan *Map_Man_t) {
	var i int32
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_Regular
		Map_MappingUnmark_rec(Map_Regular(pMan.pOutputs[i]))
	}
}

/*
*Function*************************************************************

	Synopsis    [Recursively unmarks the nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingUnmark_rec(pNode *Map_Node_t) {
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if pNode.fMark0 == 0 {
		return
	}
	pNode.fMark0 = 0
	if Map_NodeIsAnd(pNode) == 0 {
		return
	}
	// inlining of macro Map_Regular
	Map_MappingUnmark_rec(Map_Regular(pNode.p1))
	// inlining of macro Map_Regular
	Map_MappingUnmark_rec(Map_Regular(pNode.p2))
	// visit the equivalent nodes
	if pNode.pNextE != nil {
		Map_MappingUnmark_rec(pNode.pNextE)
	}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingMark_rec(pNode *Map_Node_t) {
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if pNode.fMark0 == 1 {
		return
	}
	pNode.fMark0 = 1
	if Map_NodeIsAnd(pNode) == 0 {
		return
	}
	// visit the transitive fanin of the selected cut
	// inlining of macro Map_Regular
	Map_MappingMark_rec(Map_Regular(pNode.p1))
	// inlining of macro Map_Regular
	Map_MappingMark_rec(Map_Regular(pNode.p2))
}

/*
*Function*************************************************************

	Synopsis    [Compares the outputs by their arrival times.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingCompareOutputDelay(ppNode1 **Map_Node_t, ppNode2 **Map_Node_t) int32 {
	// inlining of macro Map_Regular
	var pNode1 *Map_Node_t = Map_Regular(*ppNode1)
	// inlining of macro Map_Regular
	var pNode2 *Map_Node_t = Map_Regular(*ppNode2)
	// inlining of macro Map_IsComplement
	var fPhase1 int32 = boolToInt(!Map_IsComplement(*ppNode1))
	// inlining of macro Map_IsComplement
	var fPhase2 int32 = boolToInt(!Map_IsComplement(*ppNode2))
	var Arrival1 float32 = pNode1.tArrival[fPhase1].Worst
	var Arrival2 float32 = pNode2.tArrival[fPhase2].Worst
	if Arrival1 < Arrival2 {
		return -1
	}
	if Arrival1 > Arrival2 {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Finds given number of latest arriving COs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingFindLatest(p *Map_Man_t, pNodes []int32, nNodesMax int32) {
	var nNodes, i, k, v int32
	assert(p.nOutputs >= nNodesMax, "p.nOutputs >= nNodesMax")
	pNodes[0] = 0
	nNodes = 1
	for i = 1; i < p.nOutputs; i++ {
		for k = nNodes - 1; k >= 0; k-- {
			if Map_MappingCompareOutputDelay(&p.pOutputs[pNodes[k]], &p.pOutputs[i]) >= 0 {
				break
			}
		}
		if k == nNodesMax-1 {
			continue
		}
		if nNodes < nNodesMax {
			nNodes++
		}
		for v = nNodes - 1; v > k+1; v-- {
			pNodes[v] = pNodes[v-1]
		}
		pNodes[k+1] = i
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints a bunch of latest arriving outputs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingPrintOutputArrivals(p *Map_Man_t) {
	var pSorted [MAP_CO_LIST_SIZE]int32
	var pTimes *Map_Time_t
	var pNode *Map_Node_t
	var fPhase, Limit, i int32
	var MaxNameSize int32

	// determine the number of nodes to print
	if p.nOutputs > MAP_CO_LIST_SIZE {
		Limit = MAP_CO_LIST_SIZE
	} else {
		Limit = p.nOutputs
	}

	// determine the order
	Map_MappingFindLatest(p, pSorted[:], Limit)

	// determine max size of the node's name
	MaxNameSize = 0
	for i = 0; i < Limit; i++ {
		if MaxNameSize < int32(len(p.ppOutputNames[pSorted[i]])) {
			MaxNameSize = int32(len(p.ppOutputNames[pSorted[i]]))
		}
	}

	// print the latest outputs
	for i = 0; i < Limit; i++ {
		// get the i-th latest output
		// inlining of macro Map_Regular
		pNode = Map_Regular(p.pOutputs[pSorted[i]])
		// inlining of macro Map_IsComplement
		fPhase = boolToInt(!Map_IsComplement(p.pOutputs[pSorted[i]]))
		pTimes = &pNode.tArrival[fPhase]
		// print out the best arrival time
		fmt.Printf("Output  %-*s : ", MaxNameSize+3, p.ppOutputNames[pSorted[i]])
		fmt.Printf("Delay = (%5.2f, %5.2f)  ", float64(pTimes.Rise), float64(pTimes.Fall))
		if fPhase != 0 {
			fmt.Print("POS")
		} else {
			fmt.Print("NEG")
		}
		fmt.Printf("\n")
	}
}

/*
*Function*************************************************************

	Synopsis    [Sets up the truth tables.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingSetupTruthTables(uTruths [][2]uint32) {
	var m, v int32
	// set up the truth tables
	for m = 0; m < 32; m++ {
		for v = 0; v < 5; v++ {
			if m&(1<<v) != 0 {
				uTruths[v][0] |= (1 << m)
			}
		}
	}
	// make adjustments for the case of 6 variables
	for v = 0; v < 5; v++ {
		uTruths[v][1] = uTruths[v][0]
	}
	uTruths[5][0] = 0
	uTruths[5][1] = MAP_FULL
}

/*
*Function*************************************************************

	Synopsis    [Sets up the truth tables.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingSetupTruthTablesLarge(uTruths [][32]uint32) {
	var m, v int32
	// clean everything
	for m = 0; m < 32; m++ {
		for v = 0; v < 10; v++ {
			uTruths[v][m] = 0
		}
	}
	// set up the truth tables
	for m = 0; m < 32; m++ {
		for v = 0; v < 5; v++ {
			if m&(1<<v) != 0 {
				uTruths[v][0] |= (1 << m)
				uTruths[v+5][m] = MAP_FULL
			}
		}
	}
	// extend this info for the rest of the first 5 variables
	for m = 0; m < 32; m++ {
		for v = 0; v < 5; v++ {
			uTruths[v][m] = uTruths[v][0]
		}
	}
	/*
	   // verify
	   for ( m = 0; m < 1024; m++, printf("\n") )

	   	for ( v = 0; v < 10; v++ )
	   	    if ( Map_InfoReadVar( uTruths[v], m ) )
	   	        printf( "1" );
	   	    else
	   	        printf( "0" );
	*/
}

/*
*Function*************************************************************

	Synopsis    [Sets up the mask.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingSetupMask(uMask []uint32, nVarsMax int32) {
	if nVarsMax == 6 {
		uMask[0], uMask[1] = MAP_FULL, MAP_FULL
	} else {
		// inlining of macro MAP_MASK
		uMask[0] = ((^(uint32(0))) >> (32 - (1 << nVarsMax)))
		uMask[1] = 0
	}
}

/*
*Function*************************************************************

	Synopsis    [Verify one useful property.]

	Description [This procedure verifies one useful property. After
	the FRAIG construction with choice nodes is over, each primary node
	should have fanins that are primary nodes. The primary nodes is the
	one that does not have pNode->pRepr set to point to another node.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_ManCheckConsistency(p *Map_Man_t) int32 {
	var pNode *Map_Node_t
	var pVec *Map_NodeVec_t
	var i int32
	pVec = Map_MappingDfs(p, 0)
	for i = 0; i < pVec.nSize; i++ {
		pNode = pVec.pArray[i]
		if Map_NodeIsVar(pNode) != 0 {
			if pNode.pRepr != nil {
				fmt.Printf("Primary input %d is a secondary node.\n", pNode.Num)
			}

		} else if Map_NodeIsConst(pNode) != 0 {
			if pNode.pRepr != nil {
				fmt.Printf("Constant 1 %d is a secondary node.\n", pNode.Num)
			}

		} else {
			if pNode.pRepr != nil {
				fmt.Printf("Internal node %d is a secondary node.\n", pNode.Num)
			}
			// inlining of macro Map_Regular
			if Map_Regular(pNode.p1).pRepr != nil {
				fmt.Printf("Internal node %d has first fanin that is a secondary node.\n", pNode.Num)
			}
			// inlining of macro Map_Regular
			if Map_Regular(pNode.p2).pRepr != nil {
				fmt.Printf("Internal node %d has second fanin that is a secondary node.\n", pNode.Num)
			}
		}
	}
	Map_NodeVecFree(pVec)
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if current mapping of the node violates fanout limits.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingNodeIsViolator(pNode *Map_Node_t, pCut *Map_Cut_t, fPosPol int32) int32 {
	return boolToInt(pNode.nRefAct[fPosPol] > int32(pCut.M[fPosPol].pSuperBest.nFanLimit))
}

/*
*Function*************************************************************

	Synopsis    [Computes the total are flow of the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingGetAreaFlow(p *Map_Man_t) float32 {
	var pNode *Map_Node_t
	var pCut *Map_Cut_t
	var aFlowFlowTotal float32
	var fPosPol, i int32
	for i = 0; i < p.nOutputs; i++ {
		// inlining of macro Map_Regular
		pNode = Map_Regular(p.pOutputs[i])
		if Map_NodeIsAnd(pNode) == 0 {
			continue
		}
		// inlining of macro Map_IsComplement
		fPosPol = boolToInt(!Map_IsComplement(p.pOutputs[i]))
		pCut = pNode.pCutBest[fPosPol]
		if pCut == nil {
			fPosPol = 1 - fPosPol
			pCut = pNode.pCutBest[fPosPol]
		}
		aFlowFlowTotal += pNode.pCutBest[fPosPol].M[fPosPol].AreaFlow
	}
	return aFlowFlowTotal
}

/*
*Function*************************************************************

	Synopsis    [Compares the supergates by their level.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CompareNodesByLevel(ppS1 **Map_Node_t, ppS2 **Map_Node_t) int32 {
	// inlining of macro Map_Regular
	var pN1 *Map_Node_t = Map_Regular(*ppS1)
	// inlining of macro Map_Regular
	var pN2 *Map_Node_t = Map_Regular(*ppS2)
	if pN1.Level > pN2.Level {
		return -1
	}
	if pN1.Level < pN2.Level {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Orders the nodes in the decreasing order of levels.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingSortByLevel(pMan *Map_Man_t, vNodes *Map_NodeVec_t) {
	sort.Slice(vNodes.pArray[:vNodes.nSize], func(i, j int) bool { return Map_Regular(vNodes.pArray[i]).Level > Map_Regular(vNodes.pArray[j]).Level })
	// assert( Map_CompareNodesByLevel( vNodes->pArray, vNodes->pArray + vNodes->nSize - 1 ) <= 0 );
}

/*
*Function*************************************************************

	Synopsis    [Compares the supergates by their pointer.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CompareNodesByPointer(ppS1 **Map_Node_t, ppS2 **Map_Node_t) int32 {
	a, b := reflect.ValueOf(*ppS1).Pointer(), reflect.ValueOf(*ppS2).Pointer()
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Counts how many AIG nodes are mapped in both polarities.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingCountDoubles(pMan *Map_Man_t, vNodes *Map_NodeVec_t) int32 {
	var pNode *Map_Node_t
	var Counter, i int32
	// count the number of equal adjacent nodes
	Counter = 0
	for i = 0; i < vNodes.nSize; i++ {
		pNode = vNodes.pArray[i]
		if Map_NodeIsAnd(pNode) == 0 {
			continue
		}
		if (pNode.nRefAct[0] != 0 && pNode.pCutBest[0] != nil) &&
			(pNode.nRefAct[1] != 0 && pNode.pCutBest[1] != nil) {
			Counter++
		}
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
// The C table uses st__strcmp/st__strhash on a Mio_Gate_t pointer. Its first
// field is the name pointer, so the key consists of the in-memory bytes of
// that pointer through the first zero byte, rather than the gate's name.
func mapGatePointerKey(key any) string {
	name := Mio_GateReadName(key.(*Mio_Gate_t))
	ptr := reflect.ValueOf(name).Pointer()
	var data [8]byte
	if strconv.IntSize == 32 {
		binary.NativeEndian.PutUint32(data[:], uint32(ptr))
	} else {
		binary.NativeEndian.PutUint64(data[:], uint64(ptr))
	}
	end := 0
	for end < strconv.IntSize/8 && data[end] != 0 {
		end++
	}
	return string(data[:end])
}

func Map_CreateTableGate2Super(pMan *Map_Man_t) *st__table {
	var pSuper *Map_Super_t
	var tTable *st__table
	var i, nInputs, v int32
	tTable = st.New(func(a, b any) int { return st.StringCompare(mapGatePointerKey(a), mapGatePointerKey(b)) }, func(a any, n int) int { return st.StringHash(mapGatePointerKey(a), n) })
	for i = 0; i < pMan.pSuperLib.nSupersAll; i++ {
		pSuper = pMan.pSuperLib.ppSupers[i]
		if pSuper.nGates == 1 {
			// skip different versions of the same root gate
			nInputs = Mio_GateReadPinNum(pSuper.pRoot)
			for v = 0; v < nInputs; v++ {
				if pSuper.pFanins[v].Num != nInputs-1-v {
					break
				}
			}
			if v != nInputs {
				continue
			}
			//            printf( "%s\n", Mio_GateReadName(pSuper->pRoot) );
			if tTable.Insert(pSuper.pRoot, pSuper) {
				assert(false, "0")
			}
		}
	}
	return tTable
}

/*
*Function*************************************************************

	Synopsis    [Get the FRAIG node with phase.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_ManCleanData(p *Map_Man_t) {
	var i int32
	for i = 0; i < p.vMapObjs.nSize; i++ {
		p.vMapObjs.pArray[i].pData0, p.vMapObjs.pArray[i].pData1 = nil, nil
	}
}

/*
*Function*************************************************************

	Synopsis    [Expand the truth table]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingExpandTruth(uTruth []uint32, nVars int32) {
	assert(nVars < 7, "nVars < 7")
	if nVars == 6 {
		return
	}
	if nVars < 5 {
		// inlining of macro MAP_MASK
		uTruth[0] &= ((^(uint32(0))) >> (32 - (1 << nVars)))
		uTruth[0] = Map_MappingExpandTruth_rec(uTruth[0], nVars)
	}
	uTruth[1] = uTruth[0]
}

/*
*Function*************************************************************

	Synopsis    [Expand the truth table]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingExpandTruth_rec(uTruth uint32, nVars int32) uint32 {
	assert(nVars < 6, "nVars < 6")
	if nVars == 5 {
		return uTruth
	}
	return Map_MappingExpandTruth_rec(uTruth|(uTruth<<(1<<nVars)), nVars+1)
}

/*
*Function*************************************************************

	Synopsis    [Compute the arrival times.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingComputeDelayWithFanouts(p *Map_Man_t) float32 {
	var pNode *Map_Node_t
	var Result float32
	var i int32
	for i = 0; i < p.vMapObjs.nSize; i++ {
		// skip primary inputs
		pNode = p.vMapObjs.pArray[i]
		if Map_NodeIsAnd(pNode) == 0 {
			continue
		}
		// skip a secondary node
		if pNode.pRepr != nil {
			continue
		}
		// count the switching nodes
		if pNode.nRefAct[0] > 0 {
			Map_TimeCutComputeArrival(pNode, pNode.pCutBest[0], 0, MAP_FLOAT_LARGE)
		}
		if pNode.nRefAct[1] > 0 {
			Map_TimeCutComputeArrival(pNode, pNode.pCutBest[1], 1, MAP_FLOAT_LARGE)
		}
	}
	Result = Map_TimeComputeArrivalMax(p)
	fmt.Printf("Max arrival times with fanouts = %10.2f.\n", Result)
	return Result
}

/*
*Function*************************************************************

	Synopsis    [Sets up the mask.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingGetMaxLevel(pMan *Map_Man_t) int32 {
	var nLevelMax, i int32
	nLevelMax = 0
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_Regular
		if uint32(nLevelMax) <= Map_Regular(pMan.pOutputs[i]).Level {
			// inlining of macro Map_Regular
			nLevelMax = int32(Map_Regular(pMan.pOutputs[i]).Level)
		}
	}
	return nLevelMax
}

/*
*Function*************************************************************

	Synopsis    [Analyses choice nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingUpdateLevel_rec(pMan *Map_Man_t, pNode *Map_Node_t, fMaximum int32) int32 {
	var pTemp *Map_Node_t
	var Level1, Level2, LevelE int32
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if Map_NodeIsAnd(pNode) == 0 {
		return int32(pNode.Level)
	}
	// skip the visited node
	if pNode.TravId == pMan.nTravIds {
		return int32(pNode.Level)
	}
	pNode.TravId = pMan.nTravIds
	// compute levels of the children nodes
	// inlining of macro Map_Regular
	Level1 = Map_MappingUpdateLevel_rec(pMan, Map_Regular(pNode.p1), fMaximum)
	// inlining of macro Map_Regular
	Level2 = Map_MappingUpdateLevel_rec(pMan, Map_Regular(pNode.p2), fMaximum)
	// inlining of macro MAP_MAX
	pNode.Level = uint32(1 + max(Level1, Level2))
	if pNode.pNextE != nil {
		LevelE = Map_MappingUpdateLevel_rec(pMan, pNode.pNextE, fMaximum)
		if fMaximum != 0 {
			if pNode.Level < uint32(LevelE) {
				pNode.Level = uint32(LevelE)
			}

		} else {
			if pNode.Level > uint32(LevelE) {
				pNode.Level = uint32(LevelE)
			}
		}
		// set the level of all equivalent nodes to be the same minimum
		if pNode.pRepr == nil { // the primary node
			for pTemp = pNode.pNextE; pTemp != nil; pTemp = pTemp.pNextE {
				pTemp.Level = pNode.Level
			}
		}
	}
	return int32(pNode.Level)
}

/*
*Function*************************************************************

	Synopsis    [Resets the levels of the nodes in the choice graph.]

	Description [Makes the level of the choice nodes to be equal to the
	maximum of the level of the nodes in the equivalence class. This way
	sorting by level leads to the reverse topological order, which is
	needed for the required time computation.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingSetChoiceLevels(pMan *Map_Man_t) {
	var i int32
	pMan.nTravIds++
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_Regular
		Map_MappingUpdateLevel_rec(pMan, Map_Regular(pMan.pOutputs[i]), 1)
	}
}

/*
*Function*************************************************************

	Synopsis    [Reports statistics on choice nodes.]

	Description [The number of choice nodes is the number of primary nodes,
	which has pNextE set to a pointer. The number of choices is the number
	of entries in the equivalent-node lists of the primary nodes.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingReportChoices(pMan *Map_Man_t) {
	var pNode, pTemp *Map_Node_t
	var nChoiceNodes, nChoices int32
	var i, LevelMax1, LevelMax2 int32

	// report the number of levels
	LevelMax1 = Map_MappingGetMaxLevel(pMan)
	pMan.nTravIds++
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_Regular
		Map_MappingUpdateLevel_rec(pMan, Map_Regular(pMan.pOutputs[i]), 0)
	}
	LevelMax2 = Map_MappingGetMaxLevel(pMan)

	// report statistics about choices
	nChoiceNodes, nChoices = 0, 0
	for i = 0; i < pMan.vMapObjs.nSize; i++ {
		pNode = pMan.vMapObjs.pArray[i]
		if pNode.pRepr == nil && pNode.pNextE != nil { // this is a choice node = the primary node that has equivalent nodes
			nChoiceNodes++
			for pTemp = pNode; pTemp != nil; pTemp = pTemp.pNextE {
				nChoices++
			}
		}
	}
	fmt.Printf("Maximum level: Original = %d. Reduced due to choices = %d.\n", LevelMax1, LevelMax2)
	fmt.Printf("Choice stats:  Choice nodes = %d. Total choices = %d.\n", nChoiceNodes, nChoices)
}

/*
*Function*************************************************************

	Synopsis    [Computes the maximum and minimum levels of the choice nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingCountUsedNodes(pMan *Map_Man_t, fChoices int32) int32 {
	var vNodes *Map_NodeVec_t
	var Result int32
	vNodes = Map_MappingDfs(pMan, fChoices)
	Result = vNodes.nSize
	Map_NodeVecFree(vNodes)
	return Result
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
