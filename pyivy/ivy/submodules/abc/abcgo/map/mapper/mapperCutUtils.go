package mapper

import "fmt"

/**CFile****************************************************************

  FileName    [mapperCutUtils.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperCutUtils.h,v 1.0 2003/09/08 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Allocates the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutAlloc(p *Map_Man_t) *Map_Cut_t {
	cut := &Map_CutStruct_t_{}
	for i := range cut.M {
		cut.M[i].AreaFlow = MAP_FLOAT_LARGE
		cut.M[i].tArrive = Map_Time_t{Rise: MAP_FLOAT_LARGE, Fall: MAP_FLOAT_LARGE, Worst: MAP_FLOAT_LARGE}
	}
	return cut
}

/*
*Function*************************************************************

	Synopsis    [Deallocates the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutFree(_ *Map_Man_t, _ *Map_Cut_t) {}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutPrint(p *Map_Man_t, pRoot *Map_Node_t, pCut *Map_Cut_t, fPhase int32) {
	var i int32
	fmt.Printf("CUT:  Delay = (%4.2f, %4.2f). Area = %4.2f. Nodes = %d -> {", pCut.M[fPhase].tArrive.Rise, pCut.M[fPhase].tArrive.Fall, pCut.M[fPhase].AreaFlow, pRoot.Num)
	for i = 0; i < pCut.nLeaves; i++ {
		fmt.Printf(" %d", pCut.ppLeaves[i].Num)
	}
	fmt.Printf(" } \n")
}

/*
*function*************************************************************

	synopsis    [Computes the exact area associated with the cut.]

	description []

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_CutGetRootArea(pCut *Map_Cut_t, fPhase int32) float32 {
	assert(pCut.M[fPhase].pSuperBest != nil, "pCut.M[fPhase].pSuperBest")
	return pCut.M[fPhase].pSuperBest.Area
}

/*
*function*************************************************************

	synopsis    [Computes the exact area associated with the cut.]

	description []

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_CutGetLeafPhase(pCut *Map_Cut_t, fPhase int32, iLeaf int32) int32 {
	assert(pCut.M[fPhase].pSuperBest != nil, "pCut.M[fPhase].pSuperBest")
	return boolToInt(pCut.M[fPhase].uPhaseBest&(1<<iLeaf) == 0)
}

/*
*function*************************************************************

	synopsis    [Computes the exact area associated with the cut.]

	description []

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_NodeGetLeafPhase(pNode *Map_Node_t, fPhase int32, iLeaf int32) int32 {
	assert(pNode.pCutBest[fPhase] != nil && pNode.pCutBest[fPhase].M[fPhase].pSuperBest != nil, "pNode.pCutBest[fPhase].M[fPhase].pSuperBest")
	return boolToInt(pNode.pCutBest[fPhase].M[fPhase].uPhaseBest&(1<<iLeaf) == 0)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutListAppend(pSetAll *Map_Cut_t, pSets *Map_Cut_t) *Map_Cut_t {
	var pPrev *Map_Cut_t = nil // Suppress "might be used uninitialized"
	var pTemp *Map_Cut_t
	if pSetAll == nil {
		return pSets
	}
	if pSets == nil {
		return pSetAll
	}
	// find the last one
	for pTemp = pSets; pTemp != nil; pTemp = pTemp.pNext {
		pPrev = pTemp
	}
	// append all the end of the current set
	assert(pPrev.pNext == nil, "pPrev.pNext == nil")
	pPrev.pNext = pSetAll
	return pSets
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutListRecycle(p *Map_Man_t, pSetList *Map_Cut_t, pSave *Map_Cut_t) {
	var pNext, pTemp *Map_Cut_t
	for pTemp = pSetList; pTemp != nil; {
		pNext = pTemp.pNext
		if pTemp != pSave {
			Map_CutFree(p, pTemp)
		}
		pTemp = pNext
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
func Map_CutListCount(pSets *Map_Cut_t) int32 {
	var pTemp *Map_Cut_t
	var i int32
	for i, pTemp = 0, pSets; pTemp != nil; pTemp, i = pTemp.pNext, i+1 {
	}
	return i
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
