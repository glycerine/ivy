package if_

/**CFile****************************************************************

  FileName    [ifExpand.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Incremental improvement of current mapping.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifExpand.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void If_ManImproveExpand( If_Man_t * p, int nLimit );

// static void If_ManImproveNodeExpand( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vFrontOld, Vec_Ptr_t * vVisited );

// static void If_ManImproveNodePrepare( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vFrontOld, Vec_Ptr_t * vVisited );

// static void If_ManImproveNodeUpdate( If_Man_t * p, If_Obj_t * pObj, Vec_Ptr_t * vFront );

// static void If_ManImproveNodeFaninCompact( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Improves current mapping using expand/Expand of one cut.]

	Description [Assumes current mapping assigned and required times computed.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveMapping(p *If_Man_t) {
	var clk abctime

	clk = Abc_Clock()
	If_ManImproveExpand(p, p.pPars.NLutSize)
	If_ManComputeRequired(p)
	if p.pPars.FVerbose != 0 {
		Abc_Print(1, "E:  Del = %7.2f.  Ar = %9.1f.  Edge = %8d.  ",
			p.RequiredGlo, p.AreaGlo, p.nNets)
		if p.dPower != 0 {
			Abc_Print(1, "Switch = %7.2f.  ", p.dPower)
		}
		Abc_Print(1, "Cut = %8d.  ", p.nCutsMerged)
		Abc_PrintTime(1, "T", Abc_Clock()-clk)
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs area recovery for each node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveExpand(p *If_Man_t, nLimit int32) {
	var vFront, vFrontOld, vVisited *Vec_Ptr_t
	var pObj *If_Obj_t
	vFront = Vec_PtrAlloc(int(nLimit))
	vFrontOld = Vec_PtrAlloc(int(nLimit))
	vVisited = Vec_PtrAlloc(100)
	// iterate through all nodes in the topological order
	// inlining of macro If_ManForEachNode
	// inlining of macro If_ManForEachObj
	// inlining of macro Vec_PtrForEachEntry
	for _, entry := range Vec_PtrArray(p.vObjs) {
		pObj = entry.(*If_Obj_t)
		if pObj.Type == IF_AND {
			If_ManImproveNodeExpand(p, pObj, nLimit, vFront, vFrontOld, vVisited)
		}
	}
	Vec_PtrFree(vFront)
	Vec_PtrFree(vFrontOld)
	Vec_PtrFree(vVisited)
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of nodes with no external fanout.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveCutCost(p *If_Man_t, vFront *Vec_Ptr_t) int32 {
	var pFanin *If_Obj_t
	var Counter int32
	// inlining of macro Vec_PtrForEachEntry
	for _, entry := range Vec_PtrArray(vFront) {
		pFanin = entry.(*If_Obj_t)
		if pFanin.nRefs == 0 {
			Counter++
		}
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Performs area recovery for each node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeExpand(p *If_Man_t, pObj *If_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vFrontOld *Vec_Ptr_t, vVisited *Vec_Ptr_t) {
	var pFanin *If_Obj_t
	var pCut *If_Cut_t
	var CostBef, CostAft int32
	var DelayOld, AreaBef, AreaAft float32
	pCut = If_ObjCutBest(pObj)
	pCut.Delay = If_CutDelay(p, pObj, pCut)
	assert(pCut.Delay <= pObj.Required+p.fEpsilon, "pCut.Delay <= pObj.Required + p.fEpsilon")
	if pObj.nRefs == 0 {
		return
	}
	// get the delay
	DelayOld = pCut.Delay
	// get the area
	AreaBef = If_CutAreaRefed(p, pCut)
	//    if ( AreaBef == 1 )
	//        return;
	// the cut is non-trivial
	If_ManImproveNodePrepare(p, pObj, nLimit, vFront, vFrontOld, vVisited)
	// iteratively modify the cut
	If_CutAreaDeref(p, pCut)
	CostBef = If_ManImproveCutCost(p, vFront)
	If_ManImproveNodeFaninCompact(p, pObj, nLimit, vFront, vVisited)
	CostAft = If_ManImproveCutCost(p, vFront)
	If_CutAreaRef(p, pCut)
	assert(CostBef >= CostAft, "CostBef >= CostAft")
	// clean up
	// inlining of macro Vec_PtrForEachEntry
	for _, entry := range Vec_PtrArray(vVisited) {
		pFanin = entry.(*If_Obj_t)
		pFanin.fMark = 0
	}
	// update the node
	If_ManImproveNodeUpdate(p, pObj, vFront)
	pCut.Delay = If_CutDelay(p, pObj, pCut)
	// get the new area
	AreaAft = If_CutAreaRefed(p, pCut)
	if AreaAft > AreaBef || pCut.Delay > pObj.Required+p.fEpsilon {
		If_ManImproveNodeUpdate(p, pObj, vFrontOld)
		AreaAft = If_CutAreaRefed(p, pCut)
		assert(AreaAft == AreaBef, "AreaAft == AreaBef")
		pCut.Delay = DelayOld
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs area recovery for each node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveMark_rec(p *If_Man_t, pObj *If_Obj_t, vVisited *Vec_Ptr_t) {
	if pObj.fMark != 0 {
		return
	}
	assert(If_ObjIsAnd(pObj) != 0, "If_ObjIsAnd(pObj)")
	If_ManImproveMark_rec(p, If_ObjFanin0(pObj), vVisited)
	If_ManImproveMark_rec(p, If_ObjFanin1(pObj), vVisited)
	Vec_PtrPush(vVisited, pObj)
	pObj.fMark = 1
}

/*
*Function*************************************************************

	Synopsis    [Prepares node mapping.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodePrepare(p *If_Man_t, pObj *If_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vFrontOld *Vec_Ptr_t, vVisited *Vec_Ptr_t) {
	var pCut *If_Cut_t
	var pLeaf *If_Obj_t
	Vec_PtrClear(vFront)
	Vec_PtrClear(vFrontOld)
	Vec_PtrClear(vVisited)
	// expand the cut downwards from the given place
	pCut = If_ObjCutBest(pObj)
	// inlining of macro If_CutForEachLeaf
	for i := int32(0); i < int32(pCut.nLeaves); i++ {
		pLeaf = If_CutLeaf(p, pCut, i)
		Vec_PtrPush(vFront, pLeaf)
		Vec_PtrPush(vFrontOld, pLeaf)
		Vec_PtrPush(vVisited, pLeaf)
		pLeaf.fMark = 1
	}
	// mark the nodes in the cone
	If_ManImproveMark_rec(p, pObj, vVisited)
}

/*
*Function*************************************************************

	Synopsis    [Updates the frontier.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeUpdate(p *If_Man_t, pObj *If_Obj_t, vFront *Vec_Ptr_t) {
	var pCut *If_Cut_t
	var pFanin *If_Obj_t
	pCut = If_ObjCutBest(pObj)
	// deref node's cut
	If_CutAreaDeref(p, pCut)
	// update the node's cut
	pCut.nLeaves = uint32(Vec_PtrSize(vFront))
	// inlining of macro Vec_PtrForEachEntry
	for i, entry := range Vec_PtrArray(vFront) {
		pFanin = entry.(*If_Obj_t)
		pCut.pLeaves[i] = pFanin.Id
	}
	If_CutOrder(pCut)
	pCut.uSign = If_ObjCutSignCompute(pCut)
	// ref the new cut
	If_CutAreaRef(p, pCut)
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the number of fanins will grow.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeWillGrow(p *If_Man_t, pObj *If_Obj_t) int32 {
	var pFanin0, pFanin1 *If_Obj_t
	assert(If_ObjIsAnd(pObj) != 0, "If_ObjIsAnd(pObj)")
	pFanin0 = If_ObjFanin0(pObj)
	pFanin1 = If_ObjFanin1(pObj)
	return ifMapFlag(pFanin0.fMark == 0 && pFanin1.fMark == 0)
}

/*
*Function*************************************************************

	Synopsis    [Returns the increase in the number of fanins with no external refs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeFaninCost(p *If_Man_t, pObj *If_Obj_t) int32 {
	var Counter int32
	assert(If_ObjIsAnd(pObj) != 0, "If_ObjIsAnd(pObj)")
	// check if the node has external refs
	if pObj.nRefs == 0 {
		Counter--
	}
	// increment the number of fanins without external refs
	if If_ObjFanin0(pObj).fMark == 0 && If_ObjFanin0(pObj).nRefs == 0 {
		Counter++
	}
	// increment the number of fanins without external refs
	if If_ObjFanin1(pObj).fMark == 0 && If_ObjFanin1(pObj).nRefs == 0 {
		Counter++
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Updates the frontier.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeFaninUpdate(p *If_Man_t, pObj *If_Obj_t, vFront *Vec_Ptr_t, vVisited *Vec_Ptr_t) {
	var pFanin *If_Obj_t
	assert(If_ObjIsAnd(pObj) != 0, "If_ObjIsAnd(pObj)")
	Vec_PtrRemove(vFront, pObj)
	pFanin = If_ObjFanin0(pObj)
	if pFanin.fMark == 0 {
		Vec_PtrPush(vFront, pFanin)
		Vec_PtrPush(vVisited, pFanin)
		pFanin.fMark = 1
	}
	pFanin = If_ObjFanin1(pObj)
	if pFanin.fMark == 0 {
		Vec_PtrPush(vFront, pFanin)
		Vec_PtrPush(vVisited, pFanin)
		pFanin.fMark = 1
	}
}

/*
*Function*************************************************************

	Synopsis    [Compacts the number of external refs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeFaninCompact0(p *If_Man_t, pObj *If_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vVisited *Vec_Ptr_t) int32 {
	var pFanin *If_Obj_t
	// inlining of macro Vec_PtrForEachEntry
	for _, entry := range Vec_PtrArray(vFront) {
		pFanin = entry.(*If_Obj_t)
		if If_ObjIsCi(pFanin) != 0 {
			continue
		}
		if If_ManImproveNodeWillGrow(p, pFanin) != 0 {
			continue
		}
		if If_ManImproveNodeFaninCost(p, pFanin) <= 0 {
			If_ManImproveNodeFaninUpdate(p, pFanin, vFront, vVisited)
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Compacts the number of external refs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeFaninCompact1(p *If_Man_t, pObj *If_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vVisited *Vec_Ptr_t) int32 {
	var pFanin *If_Obj_t
	// inlining of macro Vec_PtrForEachEntry
	for _, entry := range Vec_PtrArray(vFront) {
		pFanin = entry.(*If_Obj_t)
		if If_ObjIsCi(pFanin) != 0 {
			continue
		}
		if If_ManImproveNodeFaninCost(p, pFanin) < 0 {
			If_ManImproveNodeFaninUpdate(p, pFanin, vFront, vVisited)
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Compacts the number of external refs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeFaninCompact2(p *If_Man_t, pObj *If_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vVisited *Vec_Ptr_t) int32 {
	var pFanin *If_Obj_t
	// inlining of macro Vec_PtrForEachEntry
	for _, entry := range Vec_PtrArray(vFront) {
		pFanin = entry.(*If_Obj_t)
		if If_ObjIsCi(pFanin) != 0 {
			continue
		}
		if If_ManImproveNodeFaninCost(p, pFanin) <= 0 {
			If_ManImproveNodeFaninUpdate(p, pFanin, vFront, vVisited)
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Compacts the number of external refs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeFaninCompact_int(p *If_Man_t, pObj *If_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vVisited *Vec_Ptr_t) int32 {
	if If_ManImproveNodeFaninCompact0(p, pObj, nLimit, vFront, vVisited) != 0 {
		return 1
	}
	if int32(Vec_PtrSize(vFront)) < nLimit && (If_ManImproveNodeFaninCompact1(p, pObj, nLimit, vFront, vVisited) != 0) {
		return 1
	}
	//    if ( int32(Vec_PtrSize(vFront)) < nLimit && If_ManImproveNodeFaninCompact2(p, pObj, nLimit, vFront, vVisited) )
	//        return 1;
	assert(int32(Vec_PtrSize(vFront)) <= nLimit, "int32(Vec_PtrSize(vFront)) <= nLimit")
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Compacts the number of external refs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManImproveNodeFaninCompact(p *If_Man_t, pObj *If_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vVisited *Vec_Ptr_t) {
	for If_ManImproveNodeFaninCompact_int(p, pObj, nLimit, vFront, vVisited) != 0 {
	}
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
