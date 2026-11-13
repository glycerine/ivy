package mapper

import (
	"fmt"
	"os"
	"sort"
	"time"
)

import "github.com/glycerine/abcgo/misc/extra"
import abcutil "github.com/glycerine/abcgo/misc/util"

/**CFile****************************************************************

  FileName    [mapperCut.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperCut.c,v 1.12 2005/02/28 05:34:27 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the largest number of cuts considered
const MAP_CUTS_MAX_COMPUTE = 1000

// the largest number of cuts used
const MAP_CUTS_MAX_USE = 250

// temporary hash table to store the cuts

type Map_CutTable_t struct {
	pBins  []*Map_Cut_t // the table used for linear probing
	nBins  int32        // the size of the table
	pCuts  []int32      // the array of cuts currently stored
	nCuts  int32        // the number of cuts currently stored
	pArray []*Map_Cut_t // the temporary array of cuts
	pCuts1 []*Map_Cut_t // the temporary array of cuts
	pCuts2 []*Map_Cut_t // the temporary array of cuts
}

// primes used to compute the hash key
var s_HashPrimes [10]int32 = [10]int32{109, 499, 557, 619, 631, 709, 797, 881, 907, 991}

// static Map_Cut_t *      Map_CutCompute( Map_Man_t * p, Map_CutTable_t * pTable, Map_Node_t * pNode );

// static void             Map_CutFilter( Map_Man_t * p, Map_Node_t * pNode );

// static Map_Cut_t *      Map_CutMergeLists( Map_Man_t * p, Map_CutTable_t * pTable, Map_Cut_t * pList1, Map_Cut_t * pList2, int fComp1, int fComp2 );

// static int              Map_CutMergeTwo( Map_Cut_t * pCut1, Map_Cut_t * pCut2, Map_Node_t * ppNodes[], int nNodesMax );

// static Map_Cut_t *      Map_CutUnionLists( Map_Cut_t * pList1, Map_Cut_t * pList2 );

// static int              Map_CutBelongsToList( Map_Cut_t * pList, Map_Node_t * ppNodes[], int nNodes );

// static void             Map_CutListPrint( Map_Man_t * pMan, Map_Node_t * pRoot );

// static void             Map_CutListPrint2( Map_Man_t * pMan, Map_Node_t * pRoot );

// static void             Map_CutPrint_( Map_Man_t * pMan, Map_Cut_t * pCut, Map_Node_t * pRoot );

// static Map_CutTable_t * Map_CutTableStart( Map_Man_t * pMan );

// static void             Map_CutTableStop( Map_CutTable_t * p );

// static unsigned         Map_CutTableHash( Map_Node_t * ppNodes[], int nNodes );

// static int              Map_CutTableLookup( Map_CutTable_t * p, Map_Node_t * ppNodes[], int nNodes );

// static Map_Cut_t *      Map_CutTableConsider( Map_Man_t * pMan, Map_CutTable_t * p, Map_Node_t * ppNodes[], int nNodes );

// static void             Map_CutTableRestart( Map_CutTable_t * p );

// static Map_Cut_t *      Map_CutSortCuts( Map_Man_t * pMan, Map_CutTable_t * p, Map_Cut_t * pList );

// static int              Map_CutList2Array( Map_Cut_t ** pArray, Map_Cut_t * pList );

// static Map_Cut_t *      Map_CutArray2List( Map_Cut_t ** pArray, int nCuts );

// static unsigned         Map_CutComputeTruth( Map_Man_t * p, Map_Cut_t * pCut, Map_Cut_t * pTemp0, Map_Cut_t * pTemp1, int fComp0, int fComp1 );

// iterator through all the cuts of the list
// Original macro:
// #define Map_ListForEachCut( pList, pCut )                 \
//     for ( pCut = pList;                                   \
//           pCut;                                           \
//           pCut = pCut->pNext )

// Original macro:
// #define Map_ListForEachCutSafe( pList, pCut, pCut2 )      \
//     for ( pCut = pList,                                   \
//           pCut2 = pCut? pCut->pNext: NULL;                \
//           pCut;                                           \
//           pCut = pCut2,                                   \
//           pCut2 = pCut? pCut->pNext: NULL )

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Counts all the cuts.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingCountAllCuts(pMan *Map_Man_t) int32 {
	var pNode *Map_Node_t
	var pCut *Map_Cut_t
	var i, nCuts int32
	//    int nCuts55 = 0, nCuts5x = 0, nCuts4x = 0, nCuts3x = 0;
	//    int pCounts[7] = {0};
	nCuts = 0
	for i = 0; i < pMan.nBins; i++ {
		for pNode = pMan.pBins[i]; pNode != nil; pNode = pNode.pNext {
			for pCut = pNode.pCuts; pCut != nil; pCut = pCut.pNext {
				if pCut.nLeaves > 1 { // skip the elementary cuts

					nCuts++
					/*
					   if ( Map_CutRegular(pCut->pOne)->nLeaves == 5 && Map_CutRegular(pCut->pTwo)->nLeaves == 5 )
					       nCuts55++;
					   if ( Map_CutRegular(pCut->pOne)->nLeaves == 5 || Map_CutRegular(pCut->pTwo)->nLeaves == 5 )
					       nCuts5x++;
					   else if ( Map_CutRegular(pCut->pOne)->nLeaves == 4 || Map_CutRegular(pCut->pTwo)->nLeaves == 4 )
					       nCuts4x++;
					   else if ( Map_CutRegular(pCut->pOne)->nLeaves == 3 || Map_CutRegular(pCut->pTwo)->nLeaves == 3 )
					       nCuts3x++;
					*/
					//                    pCounts[ Map_CutRegular(pCut->pOne)->nLeaves ]++;
					//                    pCounts[ Map_CutRegular(pCut->pTwo)->nLeaves ]++;
				}
			}
		}
	}
	//    printf( "Total cuts = %6d. 55 = %6d. 5x = %6d. 4x = %6d. 3x = %6d.\n", nCuts, nCuts55, nCuts5x, nCuts4x, nCuts3x );

	//    printf( "Total cuts = %6d. 6= %6d. 5= %6d. 4= %6d. 3= %6d. 2= %6d. 1= %6d.\n",
	//        nCuts, pCounts[6], pCounts[5], pCounts[4], pCounts[3], pCounts[2], pCounts[1] );
	return nCuts
}

/*
*Function*************************************************************

	Synopsis    [Computes the cuts for each node in the object graph.]

	Description [The cuts are computed in one sweep over the mapping graph.
	First, the elementary cuts, which include the node itself, are assigned
	to the PI nodes. The internal nodes are considered in the DFS order.
	Each node is two-input AND-gate. So to compute the cuts at a node, we
	need to merge the sets of cuts of its two predecessors. The merged set
	contains only unique cuts with the number of inputs equal to k or less.
	Finally, the elementary cut, composed of the node itself, is added to
	the set of cuts for the node.

	This procedure is pretty fast for 5-feasible cuts, but it dramatically
	slows down on some "dense" networks when computing 6-feasible cuts.
	The problem is that there are too many cuts in this case. We should
	think how to heuristically trim the number of cuts in such cases,
	to have reasonable runtime.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingCutsInput(p *Map_Man_t, pNode *Map_Node_t) {
	var pCut *Map_Cut_t
	assert(Map_NodeIsVar(pNode) != 0 || Map_NodeIsBuf(pNode) != 0, "Map_NodeIsVar(pNode) || Map_NodeIsBuf(pNode)")
	pCut = Map_CutAlloc(p)
	pCut.nLeaves = 1
	pCut.ppLeaves[0] = pNode
	pNode.pCuts = pCut
	pNode.pCutBest[0] = nil  // negative polarity is not mapped
	pNode.pCutBest[1] = pCut // positive polarity is a trivial cut
	pCut.uTruth = 0xAAAAAAAA // the first variable "1010"
	pCut.M[0].AreaFlow = 0.0
	pCut.M[1].AreaFlow = 0.0
}
func Map_MappingCuts(p *Map_Man_t) {
	var pProgress *ProgressBar
	var pTable *Map_CutTable_t
	var pNode *Map_Node_t
	var nCuts, nNodes, i int32
	var clk abctime = Abc_Clock()
	// set the elementary cuts for the PI variables
	assert(p.nVarsMax > 1 && p.nVarsMax < 7, "p.nVarsMax > 1 && p.nVarsMax < 7")
	for i = 0; i < p.nInputs; i++ {
		Map_MappingCutsInput(p, p.pInputs[i])
	}

	// compute the cuts for the internal nodes
	nNodes = p.vMapObjs.nSize
	pProgress = Extra_ProgressBarStart(os.Stdout, nNodes)
	pTable = Map_CutTableStart(p)
	for i = 0; i < nNodes; i++ {
		pNode = p.vMapObjs.pArray[i]
		if Map_NodeIsBuf(pNode) != 0 {
			Map_MappingCutsInput(p, pNode)
		} else if Map_NodeIsAnd(pNode) != 0 {
			Map_CutCompute(p, pTable, pNode)
		} else {
			continue
		}
		Extra_ProgressBarUpdate(pProgress, i, "Cuts ...")
	}
	Extra_ProgressBarStop(pProgress)
	Map_CutTableStop(pTable)

	// report the stats
	if p.fVerbose != 0 {
		nCuts = Map_MappingCountAllCuts(p)
		fmt.Printf("Nodes = %6d.  Total %d-feasible cuts = %10d.  Per node = %.1f. ", p.nNodes, p.nVarsMax, nCuts, float32(nCuts)/float32(p.nNodes))
		// inlining of macro ABC_PRT
		Abc_Print(1, "%s =", ("Time"))
		Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))))
	}

	// print the cuts for the first primary output
	//    Map_CutListPrint( p, Map_Regular(p->pOutputs[0]) );
}

/*
*Function*************************************************************

	Synopsis    [Computes the cuts for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutCompute(p *Map_Man_t, pTable *Map_CutTable_t, pNode *Map_Node_t) *Map_Cut_t {
	var pTemp *Map_Node_t
	var pList, pList1, pList2 *Map_Cut_t
	var pCut *Map_Cut_t

	// if the cuts are computed return them
	if pNode.pCuts != nil {
		return pNode.pCuts
	}

	// compute the cuts for the children
	// inlining of macro Map_Regular
	pList1 = Map_Regular(pNode.p1).pCuts
	// inlining of macro Map_Regular
	pList2 = Map_Regular(pNode.p2).pCuts
	// merge the lists
	pList = Map_CutMergeLists(p, pTable, pList1, pList2,
		// inlining of macro Map_IsComplement
		boolToInt(Map_IsComplement(pNode.p1)), boolToInt(Map_IsComplement(pNode.p2)))
	// if there are functionally equivalent nodes, union them with this list
	assert(pList != nil, "pList")
	// only add to the list of cuts if the node is a representative one
	if pNode.pRepr == nil {
		for pTemp = pNode.pNextE; pTemp != nil; pTemp = pTemp.pNextE {
			assert(pTemp.pCuts != nil, "pTemp.pCuts")
			pList = Map_CutUnionLists(pList, pTemp.pCuts)
			assert(pTemp.pCuts != nil, "pTemp.pCuts")
			pList = Map_CutSortCuts(p, pTable, pList)
		}
	}
	// add the new cut
	pCut = Map_CutAlloc(p)
	pCut.nLeaves = 1
	pCut.ppLeaves[0] = pNode
	pCut.uTruth = 0xAAAAAAAA
	// append (it is important that the elementary cut is appended first)
	pCut.pNext = pList
	// set at the node
	pNode.pCuts = pCut
	// remove the dominated cuts
	Map_CutFilter(p, pNode)
	// set the phase correctly
	if pNode.pRepr != nil && Map_NodeComparePhase(pNode, pNode.pRepr) != 0 {
		// inlining of macro Map_ListForEachCut
		for pCut = pNode.pCuts; pCut != nil; pCut = pCut.pNext {
			pCut.Phase = 1
		}
	}
	/*
	       {
	           int i, Counter = 0;;
	           for ( pCut = pNode->pCuts->pNext; pCut != nil; pCut = pCut->pNext )
	               for ( i = 0; i < pCut->nLeaves; i++ )
	                   Counter += (pCut->ppLeaves[i]->Level >= pNode->Level);
	   //        if ( Counter )
	   //            printf( " %d", Counter );
	       }
	*/
	return pCut
}

/*
*Function*************************************************************

	Synopsis    [Filter the cuts using dominance.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutFilter(p *Map_Man_t, pNode *Map_Node_t) {
	var pTemp, pPrev, pCut, pCut2 *Map_Cut_t
	var i, k, Counter int32

	Counter = 0
	pPrev = pNode.pCuts
	// inlining of macro Map_ListForEachCutSafe
	for pCut = pNode.pCuts.pNext; pCut != nil; {
		pCut2 = pCut.pNext
		// go through all the previous cuts up to pCut
		for pTemp = pNode.pCuts.pNext; pTemp != pCut; pTemp = pTemp.pNext {
			// check if every node in pTemp is contained in pCut
			for i = 0; i < pTemp.nLeaves; i++ {
				for k = 0; k < pCut.nLeaves; k++ {
					if pTemp.ppLeaves[i] == pCut.ppLeaves[k] {
						break
					}
				}
				if k == pCut.nLeaves { // node i in pTemp is not contained in pCut
					break
				}
			}
			if i == pTemp.nLeaves { // every node in pTemp is contained in pCut

				Counter++
				break
			}
		}
		if pTemp != pCut { // pTemp contain pCut

			pPrev.pNext = pCut.pNext // skip pCut
			// recycle pCut
			Map_CutFree(p, pCut)

		} else {
			pPrev = pCut
		}
		pCut = pCut2
	}
	// printf( "Dominated = %3d. \n", Counter );
}

/*
*Function*************************************************************

	Synopsis    [Merges two lists of cuts.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutMergeLists(p *Map_Man_t, pTable *Map_CutTable_t, pList1 *Map_Cut_t, pList2 *Map_Cut_t, fComp1 int32, fComp2 int32) *Map_Cut_t {
	var ppNodes [6]*Map_Node_t
	var pListNew *Map_Cut_t
	var ppListNew **Map_Cut_t
	var pLists [7]*Map_Cut_t = [7]*Map_Cut_t{nil}
	var pCut, pPrev, pTemp1, pTemp2 *Map_Cut_t
	var nNodes, Counter, i int32
	var ppArray1, ppArray2, ppArray3 []*Map_Cut_t
	var nCuts1, nCuts2, nCuts3, k, fComp3 int32

	ppArray1 = pTable.pCuts1
	ppArray2 = pTable.pCuts2
	nCuts1 = Map_CutList2Array(ppArray1, pList1)
	nCuts2 = Map_CutList2Array(ppArray2, pList2)
	// swap the lists based on their length
	if nCuts1 > nCuts2 {
		ppArray3 = ppArray1
		ppArray1 = ppArray2
		ppArray2 = ppArray3

		nCuts3 = nCuts1
		nCuts1 = nCuts2
		nCuts2 = nCuts3

		fComp3 = fComp1
		fComp1 = fComp2
		fComp2 = fComp3
	}
	// pList1 is shorter or equal length compared to pList2

	// prepare the manager for the cut computation
	Map_CutTableRestart(pTable)
	// go through the cut pairs
	Counter = 0
	//    for ( pTemp1 = pList1; pTemp1; pTemp1 = fPivot1? NULL: pTemp1->pNext )
	//        for ( pTemp2 = pList2; pTemp2; pTemp2 = fPivot2? NULL: pTemp2->pNext )
	for i = 0; i < nCuts1; i++ {
		for k = 0; k <= i; k++ {
			pTemp1 = ppArray1[i]
			pTemp2 = ppArray2[k]

			if pTemp1.nLeaves == p.nVarsMax && pTemp2.nLeaves == p.nVarsMax {
				if pTemp1.ppLeaves[0] != pTemp2.ppLeaves[0] {
					continue
				}
				if pTemp1.ppLeaves[1] != pTemp2.ppLeaves[1] {
					continue
				}
			}

			// check if k-feasible cut exists
			nNodes = Map_CutMergeTwo(pTemp1, pTemp2, ppNodes[:], p.nVarsMax)
			if nNodes == 0 {
				continue
			}
			// consider the cut for possible addition to the set of new cuts
			pCut = Map_CutTableConsider(p, pTable, ppNodes[:], nNodes)
			if pCut == nil {
				continue
			}
			// add data to the cut
			// inlining of macro Map_CutNotCond
			pCut.pOne = Map_CutNotCond(pTemp1, fComp1 != 0)
			// inlining of macro Map_CutNotCond
			pCut.pTwo = Map_CutNotCond(pTemp2, fComp2 != 0)
			//            if ( p->nVarsMax == 5 )
			//            pCut->uTruth = Map_CutComputeTruth( p, pCut, pTemp1, pTemp2, fComp1, fComp2 );
			// add it to the corresponding list
			pCut.pNext = pLists[int32(pCut.nLeaves)]
			pLists[int32(pCut.nLeaves)] = pCut
			// count this cut and quit if limit is reached
			Counter++
			if Counter == MAP_CUTS_MAX_COMPUTE {
				goto QUITS
			}
		}
		for k = 0; k < i; k++ {
			pTemp1 = ppArray1[k]
			pTemp2 = ppArray2[i]

			if pTemp1.nLeaves == p.nVarsMax && pTemp2.nLeaves == p.nVarsMax {
				if pTemp1.ppLeaves[0] != pTemp2.ppLeaves[0] {
					continue
				}
				if pTemp1.ppLeaves[1] != pTemp2.ppLeaves[1] {
					continue
				}
			}

			// check if k-feasible cut exists
			nNodes = Map_CutMergeTwo(pTemp1, pTemp2, ppNodes[:], p.nVarsMax)
			if nNodes == 0 {
				continue
			}
			// consider the cut for possible addition to the set of new cuts
			pCut = Map_CutTableConsider(p, pTable, ppNodes[:], nNodes)
			if pCut == nil {
				continue
			}
			// add data to the cut
			// inlining of macro Map_CutNotCond
			pCut.pOne = Map_CutNotCond(pTemp1, fComp1 != 0)
			// inlining of macro Map_CutNotCond
			pCut.pTwo = Map_CutNotCond(pTemp2, fComp2 != 0)
			//            if ( p->nVarsMax == 5 )
			//            pCut->uTruth = Map_CutComputeTruth( p, pCut, pTemp1, pTemp2, fComp1, fComp2 );
			// add it to the corresponding list
			pCut.pNext = pLists[int32(pCut.nLeaves)]
			pLists[int32(pCut.nLeaves)] = pCut
			// count this cut and quit if limit is reached
			Counter++
			if Counter == MAP_CUTS_MAX_COMPUTE {
				goto QUITS
			}
		}
	}
	// consider the rest of them
	for i = nCuts1; i < nCuts2; i++ {
		for k = 0; k < nCuts1; k++ {
			pTemp1 = ppArray1[k]
			pTemp2 = ppArray2[i]

			if pTemp1.nLeaves == p.nVarsMax && pTemp2.nLeaves == p.nVarsMax {
				if pTemp1.ppLeaves[0] != pTemp2.ppLeaves[0] {
					continue
				}
				if pTemp1.ppLeaves[1] != pTemp2.ppLeaves[1] {
					continue
				}
			}

			// check if k-feasible cut exists
			nNodes = Map_CutMergeTwo(pTemp1, pTemp2, ppNodes[:], p.nVarsMax)
			if nNodes == 0 {
				continue
			}
			// consider the cut for possible addition to the set of new cuts
			pCut = Map_CutTableConsider(p, pTable, ppNodes[:], nNodes)
			if pCut == nil {
				continue
			}
			// add data to the cut
			// inlining of macro Map_CutNotCond
			pCut.pOne = Map_CutNotCond(pTemp1, fComp1 != 0)
			// inlining of macro Map_CutNotCond
			pCut.pTwo = Map_CutNotCond(pTemp2, fComp2 != 0)
			//            if ( p->nVarsMax == 5 )
			//            pCut->uTruth = Map_CutComputeTruth( p, pCut, pTemp1, pTemp2, fComp1, fComp2 );
			// add it to the corresponding list
			pCut.pNext = pLists[int32(pCut.nLeaves)]
			pLists[int32(pCut.nLeaves)] = pCut
			// count this cut and quit if limit is reached
			Counter++
			if Counter == MAP_CUTS_MAX_COMPUTE {
				goto QUITS
			}
		}
	}
QUITS:
	// combine all the lists into one
	pListNew = nil
	ppListNew = &pListNew
	for i = 1; i <= p.nVarsMax; i++ {
		if pLists[i] == nil {
			continue
		}
		// find the last entry
		pPrev = pLists[i]
		for pCut = pPrev.pNext; pCut != nil; pCut = pCut.pNext {
			pPrev = pCut
		}
		// connect these lists
		*ppListNew = pLists[i]
		ppListNew = &pPrev.pNext
	}
	*ppListNew = nil
	// soft the cuts by arrival times and use only the first MAP_CUTS_MAX_USE
	pListNew = Map_CutSortCuts(p, pTable, pListNew)
	return pListNew
}

/*
*Function*************************************************************

	Synopsis    [Merges two lists of cuts.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutMergeLists2(p *Map_Man_t, pTable *Map_CutTable_t, pList1 *Map_Cut_t, pList2 *Map_Cut_t, fComp1 int32, fComp2 int32) *Map_Cut_t {
	var ppNodes [6]*Map_Node_t
	var pListNew *Map_Cut_t
	var ppListNew **Map_Cut_t
	var pLists [7]*Map_Cut_t = [7]*Map_Cut_t{nil}
	var pCut, pPrev, pTemp1, pTemp2 *Map_Cut_t
	var nNodes, Counter, i int32

	// prepare the manager for the cut computation
	Map_CutTableRestart(pTable)
	// go through the cut pairs
	Counter = 0
	for pTemp1 = pList1; pTemp1 != nil; pTemp1 = pTemp1.pNext {
		for pTemp2 = pList2; pTemp2 != nil; pTemp2 = pTemp2.pNext {
			// check if k-feasible cut exists
			nNodes = Map_CutMergeTwo(pTemp1, pTemp2, ppNodes[:], p.nVarsMax)
			if nNodes == 0 {
				continue
			}
			// consider the cut for possible addition to the set of new cuts
			pCut = Map_CutTableConsider(p, pTable, ppNodes[:], nNodes)
			if pCut == nil {
				continue
			}
			// add data to the cut
			// inlining of macro Map_CutNotCond
			pCut.pOne = Map_CutNotCond(pTemp1, fComp1 != 0)
			// inlining of macro Map_CutNotCond
			pCut.pTwo = Map_CutNotCond(pTemp2, fComp2 != 0)
			// add it to the corresponding list
			pCut.pNext = pLists[int32(pCut.nLeaves)]
			pLists[int32(pCut.nLeaves)] = pCut
			// count this cut and quit if limit is reached
			Counter++
			if Counter == MAP_CUTS_MAX_COMPUTE {
				goto QUITS
			}
		}
	}
QUITS:
	// combine all the lists into one
	pListNew = nil
	ppListNew = &pListNew
	for i = 1; i <= p.nVarsMax; i++ {
		if pLists[i] == nil {
			continue
		}
		// find the last entry
		pPrev = pLists[i]
		for pCut = pPrev.pNext; pCut != nil; pCut = pCut.pNext {
			pPrev = pCut
		}
		// connect these lists
		*ppListNew = pLists[i]
		ppListNew = &pPrev.pNext
	}
	*ppListNew = nil
	// soft the cuts by arrival times and use only the first MAP_CUTS_MAX_USE
	pListNew = Map_CutSortCuts(p, pTable, pListNew)
	return pListNew
}

/*
*Function*************************************************************

	Synopsis    [Merges two cuts.]

	Description [Returns the number of nodes in the resulting cut, or 0 if the
	cut is infeasible. Returns the resulting nodes in the array ppNodes[].]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutMergeTwo(pCut1 *Map_Cut_t, pCut2 *Map_Cut_t, ppNodes []*Map_Node_t, nNodesMax int32) int32 {
	var pNodeTemp *Map_Node_t
	var nTotal, i, k, min, fMismatch int32

	// check the special case when at least of the cuts is the largest
	if pCut1.nLeaves == nNodesMax {
		if pCut2.nLeaves == nNodesMax {
			// return 0 if the cuts are different
			for i = 0; i < nNodesMax; i++ {
				if pCut1.ppLeaves[i] != pCut2.ppLeaves[i] {
					return 0
				}
			}
			// return nNodesMax if they are the same
			for i = 0; i < nNodesMax; i++ {
				ppNodes[i] = pCut1.ppLeaves[i]
			}
			return nNodesMax

		} else if pCut2.nLeaves == nNodesMax-1 {
			// return 0 if the cuts are different
			fMismatch = 0
			for i = 0; i < nNodesMax; i++ {
				if pCut1.ppLeaves[i] != pCut2.ppLeaves[i-fMismatch] {
					if fMismatch == 1 {
						return 0
					}
					fMismatch = 1
				}
			}
			// return nNodesMax if they are the same
			for i = 0; i < nNodesMax; i++ {
				ppNodes[i] = pCut1.ppLeaves[i]
			}
			return nNodesMax
		}

	} else if pCut1.nLeaves == nNodesMax-1 && pCut2.nLeaves == nNodesMax {
		// return 0 if the cuts are different
		fMismatch = 0
		for i = 0; i < nNodesMax; i++ {
			if pCut1.ppLeaves[i-fMismatch] != pCut2.ppLeaves[i] {
				if fMismatch == 1 {
					return 0
				}
				fMismatch = 1
			}
		}
		// return nNodesMax if they are the same
		for i = 0; i < nNodesMax; i++ {
			ppNodes[i] = pCut2.ppLeaves[i]
		}
		return nNodesMax
	}

	// count the number of unique entries in pCut2
	nTotal = pCut1.nLeaves
	for i = 0; i < pCut2.nLeaves; i++ {
		// try to find this entry among the leaves of pCut1
		for k = 0; k < pCut1.nLeaves; k++ {
			if pCut2.ppLeaves[i] == pCut1.ppLeaves[k] {
				break
			}
		}
		if k < pCut1.nLeaves { // found
			continue
		}
		// we found a new entry to add
		if nTotal == nNodesMax {
			return 0
		}
		ppNodes[nTotal] = pCut2.ppLeaves[i]
		nTotal++
	}
	// we know that the feasible cut exists

	// add the starting entries
	for k = 0; k < pCut1.nLeaves; k++ {
		ppNodes[k] = pCut1.ppLeaves[k]
	}

	// selection-sort the entries
	for i = 0; i < nTotal-1; i++ {
		min = i
		for k = i + 1; k < nTotal; k++ {
			//            if ( ppNodes[k] < ppNodes[min] ) // reported bug fix (non-determinism!)
			if ppNodes[k].Num < ppNodes[min].Num {
				min = k
			}
		}
		pNodeTemp = ppNodes[i]
		ppNodes[i] = ppNodes[min]
		ppNodes[min] = pNodeTemp
	}

	return nTotal
}

/*
*Function*************************************************************

	Synopsis    [Computes the union of the two lists of cuts.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutUnionLists(pList1 *Map_Cut_t, pList2 *Map_Cut_t) *Map_Cut_t {
	var pTemp, pRoot *Map_Cut_t
	// find the last cut in the first list
	pRoot = pList1
	// inlining of macro Map_ListForEachCut
	for pTemp = pList1; pTemp != nil; pTemp = pTemp.pNext {
		pRoot = pTemp
	}
	// attach the non-trival part of the second cut to the end of the first
	assert(pRoot.pNext == nil, "pRoot.pNext == nil")
	pRoot.pNext = pList2.pNext
	pList2.pNext = nil
	return pList1
}

/*
*Function*************************************************************

	Synopsis    [Checks whether the given cut belongs to the list.]

	Description [This procedure takes most of the runtime in the cut
	computation.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutBelongsToList(pList *Map_Cut_t, ppNodes []*Map_Node_t, nNodes int32) int32 {
	var pTemp *Map_Cut_t
	var i int32
	for pTemp = pList; pTemp != nil; pTemp = pTemp.pNext {
		for i = 0; i < nNodes; i++ {
			if pTemp.ppLeaves[i] != ppNodes[i] {
				break
			}
		}
		if i == nNodes {
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Prints the cuts in the list.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutListPrint(pMan *Map_Man_t, pRoot *Map_Node_t) {
	var pTemp *Map_Cut_t
	var Counter int32
	for Counter, pTemp = 0, pRoot.pCuts; pTemp != nil; pTemp, Counter = pTemp.pNext, Counter+1 {
		fmt.Printf("%2d : ", Counter+1)
		Map_CutPrint_(pMan, pTemp, pRoot)
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints the cuts in the list.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutListPrint2(pMan *Map_Man_t, pRoot *Map_Node_t) {
	var pTemp *Map_Cut_t
	var Counter int32
	for Counter, pTemp = 0, pRoot.pCuts; pTemp != nil; pTemp, Counter = pTemp.pNext, Counter+1 {
		fmt.Printf("%2d : ", Counter+1)
		Map_CutPrint_(pMan, pTemp, pRoot)
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutPrint_(pMan *Map_Man_t, pCut *Map_Cut_t, pRoot *Map_Node_t) {
	var i int32
	fmt.Printf("(%3d)  {", pRoot.Num)
	for i = 0; i < pMan.nVarsMax; i++ {
		if pCut.ppLeaves[i] != nil {
			fmt.Printf(" %3d", pCut.ppLeaves[i].Num)
		}
	}
	fmt.Printf(" }\n")
}

/*
*Function*************************************************************

	Synopsis    [Starts the hash table to canonicize cuts.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutTableStart(pMan *Map_Man_t) *Map_CutTable_t {
	bins := abcutil.Abc_PrimeCudd(uint32(10 * MAP_CUTS_MAX_COMPUTE))
	capacity := 2 * MAP_CUTS_MAX_COMPUTE
	return &Map_CutTable_t{nBins: bins, pBins: make([]*Map_Cut_t, bins), pCuts: make([]int32, capacity), pArray: make([]*Map_Cut_t, capacity), pCuts1: make([]*Map_Cut_t, capacity), pCuts2: make([]*Map_Cut_t, capacity)}
}

/*
*Function*************************************************************

	Synopsis    [Stops the hash table.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutTableStop(p *Map_CutTable_t) {
	p.pCuts1 = nil
	p.pCuts2 = nil
	p.pArray = nil
	p.pBins = nil
	p.pCuts = nil
	p = nil
}

/*
*Function*************************************************************

	Synopsis    [Computes the hash value of the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutTableHash(ppNodes []*Map_Node_t, nNodes int32) uint32 {
	var uRes uint32
	var i int32
	uRes = 0
	for i = 0; i < nNodes; i++ {
		uRes += uint32(s_HashPrimes[i]) * uint32(ppNodes[i].Num)
	}
	return uRes
}

/*
*Function*************************************************************

	Synopsis    [Looks up the table for the available cut.]

	Description [Returns -1 if the same cut is found. Returns the index
	of the cell where the cut should be added, if it does not exist.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutTableLookup(p *Map_CutTable_t, ppNodes []*Map_Node_t, nNodes int32) int32 {
	var pCut *Map_Cut_t
	var Key uint32
	var b, i int32

	Key = Map_CutTableHash(ppNodes, nNodes) % uint32(p.nBins)
	for b = int32(Key); p.pBins[b] != nil; b = (b + 1) % p.nBins {
		pCut = p.pBins[b]
		if pCut.nLeaves != nNodes {
			continue
		}
		for i = 0; i < nNodes; i++ {
			if pCut.ppLeaves[i] != ppNodes[i] {
				break
			}
		}
		if i == nNodes {
			return -1
		}
	}
	return b
}

/*
*Function*************************************************************

	Synopsis    [Starts the hash table to canonicize cuts.]

	Description [Considers addition of the cut to the hash table.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutTableConsider(pMan *Map_Man_t, p *Map_CutTable_t, ppNodes []*Map_Node_t, nNodes int32) *Map_Cut_t {
	var pCut *Map_Cut_t
	var Place, i int32
	//    abctime clk;
	// check the cut
	Place = Map_CutTableLookup(p, ppNodes, nNodes)
	if Place == -1 {
		return nil
	}
	assert(nNodes > 0, "nNodes > 0")
	// create the new cut
	//clk = Abc_Clock();
	pCut = Map_CutAlloc(pMan)
	//pMan->time1 += Abc_Clock() - clk;
	pCut.nLeaves = nNodes
	for i = 0; i < nNodes; i++ {
		pCut.ppLeaves[i] = ppNodes[i]
	}
	// add the cut to the table
	assert(p.pBins[Place] == nil, "p.pBins[Place] == nil")
	p.pBins[Place] = pCut
	// add the cut to the new list
	p.pCuts[p.nCuts] = Place
	p.nCuts++
	return pCut
}

/*
*Function*************************************************************

	Synopsis    [Prepares the table to be used with other cuts.]

	Description [Restarts the table by cleaning the info about cuts stored
	when the previous node was considered.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutTableRestart(p *Map_CutTable_t) {
	var i int32
	for i = 0; i < p.nCuts; i++ {
		assert(p.pBins[p.pCuts[i]] != nil, "p.pBins[ p.pCuts[i] ]")
		p.pBins[p.pCuts[i]] = nil
	}
	p.nCuts = 0
}

/*
*Function*************************************************************

	Synopsis    [Compares the cuts by the number of leaves and then by delay.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutSortCutsCompare(pC1, pC2 *Map_Cut_t) int {
	if pC1.nLeaves < pC2.nLeaves {
		return -1
	}
	if pC1.nLeaves > pC2.nLeaves {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Sorts the cuts by average arrival time.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutSortCuts(pMan *Map_Man_t, p *Map_CutTable_t, pList *Map_Cut_t) *Map_Cut_t {
	var pListNew *Map_Cut_t
	var nCuts, i int32
	//    abctime clk;
	// move the cuts from the list into the array
	nCuts = Map_CutList2Array(p.pCuts1, pList)
	assert(nCuts <= MAP_CUTS_MAX_COMPUTE, "nCuts <= MAP_CUTS_MAX_COMPUTE")
	// sort the cuts
	//clk = Abc_Clock();
	sort.SliceStable(p.pCuts1[:nCuts], func(i, j int) bool { return Map_CutSortCutsCompare(p.pCuts1[i], p.pCuts1[j]) < 0 })
	//pMan->time2 += Abc_Clock() - clk;
	// move them back into the list
	if nCuts > MAP_CUTS_MAX_USE-1 {
		// free the remaining cuts
		for i = MAP_CUTS_MAX_USE - 1; i < nCuts; i++ {
			p.pCuts1[i] = nil
		}
		// update the number of cuts
		nCuts = MAP_CUTS_MAX_USE - 1
	}
	pListNew = Map_CutArray2List(p.pCuts1, nCuts)
	return pListNew
}

/*
*Function*************************************************************

	Synopsis    [Moves the nodes from the list into the array.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutList2Array(pArray []*Map_Cut_t, pList *Map_Cut_t) int32 {
	var i int32
	for i = 0; pList != nil; pList, i = pList.pNext, i+1 {
		pArray[i] = pList
	}
	return i
}

/*
*Function*************************************************************

	Synopsis    [Moves the nodes from the array into the list.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutArray2List(pArray []*Map_Cut_t, nCuts int32) *Map_Cut_t {
	var pListNew *Map_Cut_t
	var ppListNew **Map_Cut_t
	var i int32
	pListNew = nil
	ppListNew = &pListNew
	for i = 0; i < nCuts; i++ {
		// connect these lists
		*ppListNew = pArray[i]
		ppListNew = &pArray[i].pNext
	}
	//printf( "\n" );

	*ppListNew = nil
	return pListNew
}

/*
*Function*************************************************************

	Synopsis    [Computes the truth table of the 5-input cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Map_CutComputeTruth_pPerms53 [][]uint32
var spStatic_Map_CutComputeTruth_pPerms54 [][]uint32

func Map_CutComputeTruth(p *Map_Man_t, pCut *Map_Cut_t, pTemp0 *Map_Cut_t, pTemp1 *Map_Cut_t, fComp0 int32, fComp1 int32) uint32 {

	var uPhase, uTruth, uTruth0, uTruth1 uint32
	var i, k int32

	if spStatic_Map_CutComputeTruth_pPerms53 == nil {
		spStatic_Map_CutComputeTruth_pPerms53 = extra.Extra_TruthPerm53()
		spStatic_Map_CutComputeTruth_pPerms54 = extra.Extra_TruthPerm54()
	}

	// find the mapping from the old nodes to the new
	if pTemp0.nLeaves == pCut.nLeaves {
		uTruth0 = pTemp0.uTruth
	} else {
		assert(pTemp0.nLeaves < pCut.nLeaves, "pTemp0.nLeaves < pCut.nLeaves")
		uPhase = 0
		for i = 0; i < int32(pTemp0.nLeaves); i++ {
			for k = 0; k < pCut.nLeaves; k++ {
				if pTemp0.ppLeaves[i] == pCut.ppLeaves[k] {
					break
				}
			}
			uPhase |= (1 << k)
		}
		assert(uPhase < 32, "uPhase < 32")
		if pTemp0.nLeaves == 4 {
			if uPhase == 31-16 { // 01111
				uTruth0 = pTemp0.uTruth
			} else if uPhase == 31-8 { // 10111
				uTruth0 = spStatic_Map_CutComputeTruth_pPerms54[pTemp0.uTruth&0xFFFF][0]
			} else if uPhase == 31-4 { // 11011
				uTruth0 = spStatic_Map_CutComputeTruth_pPerms54[pTemp0.uTruth&0xFFFF][1]
			} else if uPhase == 31-2 { // 11101
				uTruth0 = spStatic_Map_CutComputeTruth_pPerms54[pTemp0.uTruth&0xFFFF][2]
			} else if uPhase == 31-1 { // 11110
				uTruth0 = spStatic_Map_CutComputeTruth_pPerms54[pTemp0.uTruth&0xFFFF][3]
			} else {
				assert(false, "0")
			}

		} else {
			uTruth0 = spStatic_Map_CutComputeTruth_pPerms53[pTemp0.uTruth&0xFF][uPhase]
		}
	}
	if fComp0 != 0 {
		uTruth0 = ^uTruth0
	}

	// find the mapping from the old nodes to the new
	if pTemp1.nLeaves == pCut.nLeaves {
		uTruth1 = pTemp1.uTruth
	} else {
		assert(pTemp1.nLeaves < pCut.nLeaves, "pTemp1.nLeaves < pCut.nLeaves")
		uPhase = 0
		for i = 0; i < int32(pTemp1.nLeaves); i++ {
			for k = 0; k < pCut.nLeaves; k++ {
				if pTemp1.ppLeaves[i] == pCut.ppLeaves[k] {
					break
				}
			}
			uPhase |= (1 << k)
		}
		assert(uPhase < 32, "uPhase < 32")
		if pTemp1.nLeaves == 4 {
			if uPhase == 31-16 { // 01111
				uTruth1 = pTemp1.uTruth
			} else if uPhase == 31-8 { // 10111
				uTruth1 = spStatic_Map_CutComputeTruth_pPerms54[pTemp1.uTruth&0xFFFF][0]
			} else if uPhase == 31-4 { // 11011
				uTruth1 = spStatic_Map_CutComputeTruth_pPerms54[pTemp1.uTruth&0xFFFF][1]
			} else if uPhase == 31-2 { // 11101
				uTruth1 = spStatic_Map_CutComputeTruth_pPerms54[pTemp1.uTruth&0xFFFF][2]
			} else if uPhase == 31-1 { // 11110
				uTruth1 = spStatic_Map_CutComputeTruth_pPerms54[pTemp1.uTruth&0xFFFF][3]
			} else {
				assert(false, "0")
			}

		} else {
			uTruth1 = spStatic_Map_CutComputeTruth_pPerms53[pTemp1.uTruth&0xFF][uPhase]
		}
	}
	if fComp1 != 0 {
		uTruth1 = ^uTruth1
	}
	uTruth = uTruth0 & uTruth1
	return uTruth
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
