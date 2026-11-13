package res

import "github.com/glycerine/abcgo/base/abc"

/**CFile****************************************************************

  FileName    [resWin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Windowing algorithm.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resWin.c,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "resInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Allocates the window.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinAlloc() *Res_Win_t {
	p := &Res_Win_t{}
	// start the manager
	// set internal parameters
	p.nFanoutLimit = 10
	p.nLevTfiMinus = 3
	// allocate storage
	p.vRoots = Vec_PtrAlloc(256)
	p.vLeaves = Vec_PtrAlloc(256)
	p.vBranches = Vec_PtrAlloc(256)
	p.vNodes = Vec_PtrAlloc(256)
	p.vDivs = Vec_PtrAlloc(256)
	p.vMatrix = Vec_VecStart(128)
	return p
}

/*
*Function*************************************************************

	Synopsis    [Frees the window.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinFree(p *Res_Win_t) {
	if p == nil {
		return
	}
	Vec_PtrFree(p.vRoots)
	p.vRoots = nil
	Vec_PtrFree(p.vLeaves)
	p.vLeaves = nil
	Vec_PtrFree(p.vBranches)
	p.vBranches = nil
	Vec_PtrFree(p.vNodes)
	p.vNodes = nil
	Vec_PtrFree(p.vDivs)
	p.vDivs = nil
	Vec_VecFree(p.vMatrix)
	p.vMatrix = nil
	p.pNode = nil
}

/*
*Function*************************************************************

	Synopsis    [Collect the limited TFI cone of the node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinCollectLeavesAndNodes(p *Res_Win_t) int32 {
	var vFront *Vec_Ptr_t
	var pObj, pTemp *Abc_Obj_t
	var i, k, m int32

	assert(p.nWinTfiMax > 0, "p.nWinTfiMax > 0")
	assert(Vec_VecSize(p.vMatrix) > p.nWinTfiMax, "Vec_VecSize(p.vMatrix) > p.nWinTfiMax")

	// start matrix with the node
	Vec_VecClear(p.vMatrix)
	Vec_VecPush(p.vMatrix, 0, p.pNode)
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(p.pNode))
	abc.Abc_NodeSetTravIdCurrent(p.pNode)

	// collect the leaves (nodes pTemp such that "p->pNode->Level - pTemp->Level > p->nWinTfiMax")
	Vec_PtrClear(p.vLeaves)
	// inlining of macro Vec_VecForEachLevelStartStop
	for i = 0; i < p.nWinTfiMax+1; i++ {
		vFront = Vec_VecEntry(p.vMatrix, i)
		// inlining of macro Vec_PtrForEachEntry
		for k = 0; k < Vec_PtrSize(vFront); k++ {
			pObj = Vec_PtrEntry(vFront, k)
			// inlining of macro Abc_ObjForEachFanin
			for m = 0; m < abc.Abc_ObjFaninNum(pObj); m++ {
				pTemp = abc.Abc_ObjFanin(pObj, m)
				if abc.Abc_NodeIsTravIdCurrent(pTemp) != 0 {
					continue
				}
				abc.Abc_NodeSetTravIdCurrent(pTemp)
				if abc.Abc_ObjIsCi(pTemp) != 0 || abc.Abc_ObjLevel(p.pNode)-abc.Abc_ObjLevel(pTemp) > p.nWinTfiMax {
					Vec_PtrPush(p.vLeaves, pTemp)
				} else {
					Vec_VecPush(p.vMatrix, abc.Abc_ObjLevel(p.pNode)-abc.Abc_ObjLevel(pTemp), pTemp)
				}
			}
		}
	}
	if Vec_PtrSize(p.vLeaves) == 0 {
		return 0
	}

	// collect the nodes in the reverse order
	Vec_PtrClear(p.vNodes)
	// inlining of macro Vec_VecForEachLevelReverseStartStop
	for i = p.nWinTfiMax + 1 - 1; i >= 0; i-- {
		vFront = Vec_VecEntry(p.vMatrix, i)
		// inlining of macro Vec_PtrForEachEntry
		for k = 0; k < Vec_PtrSize(vFront); k++ {
			pObj = Vec_PtrEntry(vFront, k)
			Vec_PtrPush(p.vNodes, pObj)
		}
		Vec_PtrClear(vFront)
	}

	// get the lowest leaf level
	p.nLevLeafMin = 1000000000
	// inlining of macro Vec_PtrForEachEntry
	for k = 0; k < Vec_PtrSize(p.vLeaves); k++ {
		pObj = Vec_PtrEntry(p.vLeaves, k)
		if level := abc.Abc_ObjLevel(pObj); level < p.nLevLeafMin {
			p.nLevLeafMin = level
		}
	}

	// set minimum traversal level
	p.nLevTravMin = abc.Abc_ObjLevel(p.pNode) - p.nWinTfiMax - p.nLevTfiMinus
	if p.nLevTravMin < p.nLevLeafMin {
		p.nLevTravMin = p.nLevLeafMin
	}
	assert(p.nLevTravMin >= 0, "p.nLevTravMin >= 0")
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the node should be a root.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinComputeRootsCheck(pNode *Abc_Obj_t, nLevelMax int32, nFanoutLimit int32) int32 {
	var pFanout *Abc_Obj_t
	var i int32
	// the node is the root if one of the following is true:
	// (1) the node has more than fanouts than the limit
	if abc.Abc_ObjFanoutNum(pNode) > nFanoutLimit {
		return 1
	}
	// (2) the node has CO fanouts
	// (3) the node has fanouts above the cutoff level
	// inlining of macro Abc_ObjForEachFanout
	for i = 0; i < abc.Abc_ObjFanoutNum(pNode); i++ {
		pFanout = abc.Abc_ObjFanout(pNode, i)
		if abc.Abc_ObjIsCo(pFanout) != 0 || abc.Abc_ObjLevel(pFanout) > nLevelMax {
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Recursively collects the root candidates.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinComputeRoots_rec(pNode *Abc_Obj_t, nLevelMax int32, nFanoutLimit int32, vRoots *Vec_Ptr_t) {
	var pFanout *Abc_Obj_t
	var i int32
	assert(abc.Abc_ObjIsNode(pNode) != 0, "abc.Abc_ObjIsNode(pNode)")
	if abc.Abc_NodeIsTravIdCurrent(pNode) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(pNode)
	// check if the node should be the root
	if Res_WinComputeRootsCheck(pNode, nLevelMax, nFanoutLimit) != 0 {
		Vec_PtrPush(vRoots, pNode)
	} else { // if not, explore its fanouts
		// inlining of macro Abc_ObjForEachFanout
		for i = 0; i < abc.Abc_ObjFanoutNum(pNode); i++ {
			pFanout = abc.Abc_ObjFanout(pNode, i)
			Res_WinComputeRoots_rec(pFanout, nLevelMax, nFanoutLimit, vRoots)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Recursively collects the root candidates.]

	Description [Returns 1 if the only root is this node.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinComputeRoots(p *Res_Win_t) int32 {
	Vec_PtrClear(p.vRoots)
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(p.pNode))
	Res_WinComputeRoots_rec(p.pNode, abc.Abc_ObjLevel(p.pNode)+p.nWinTfoMax, p.nFanoutLimit, p.vRoots)
	assert(Vec_PtrSize(p.vRoots) > 0, "Vec_PtrSize(p.vRoots) > 0")
	if Vec_PtrSize(p.vRoots) == 1 && Vec_PtrEntry(p.vRoots, 0) == p.pNode {
		return 0
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Marks the paths from the roots to the leaves.]

	Description [Returns 1 if the the node can reach a leaf.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinMarkPaths_rec(pNode *Abc_Obj_t, pPivot *Abc_Obj_t, nLevelMin int32) int32 {
	var pFanin *Abc_Obj_t
	var i, RetValue int32
	// skip visited nodes
	if abc.Abc_NodeIsTravIdCurrent(pNode) != 0 {
		return 1
	}
	if abc.Abc_NodeIsTravIdPrevious(pNode) != 0 {
		return 0
	}
	// assume that the node does not have access to the leaves
	abc.Abc_NodeSetTravIdPrevious(pNode)
	// skip nodes below the given level
	if pNode == pPivot || abc.Abc_ObjLevel(pNode) <= nLevelMin {
		return 0
	}
	assert(abc.Abc_ObjIsNode(pNode) != 0, "abc.Abc_ObjIsNode(pNode)")
	// check if the fanins have access to the leaves
	RetValue = 0
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < abc.Abc_ObjFaninNum(pNode); i++ {
		pFanin = abc.Abc_ObjFanin(pNode, i)
		RetValue |= Res_WinMarkPaths_rec(pFanin, pPivot, nLevelMin)
	}
	// relabel the node if it has access to the leaves
	if RetValue != 0 {
		abc.Abc_NodeSetTravIdCurrent(pNode)
	}
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Marks the paths from the roots to the leaves.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinMarkPaths(p *Res_Win_t) {
	var pObj *Abc_Obj_t
	var i int32
	// mark the leaves
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(p.pNode))
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(p.pNode))
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(p.vLeaves); i++ {
		pObj = Vec_PtrEntry(p.vLeaves, i)
		abc.Abc_NodeSetTravIdCurrent(pObj)
	}
	// traverse from the roots and mark the nodes that can reach leaves
	// the nodes that do not reach leaves have previous trav ID
	// the nodes that reach leaves have current trav ID
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(p.vRoots); i++ {
		pObj = Vec_PtrEntry(p.vRoots, i)
		Res_WinMarkPaths_rec(pObj, p.pNode, p.nLevTravMin)
	}
}

/*
*Function*************************************************************

	Synopsis    [Recursively collects the roots.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinFinalizeRoots_rec(pObj *Abc_Obj_t, vRoots *Vec_Ptr_t) {
	var pFanout *Abc_Obj_t
	var i int32
	assert(abc.Abc_ObjIsNode(pObj) != 0, "abc.Abc_ObjIsNode(pObj)")
	assert(abc.Abc_NodeIsTravIdCurrent(pObj) != 0, "abc.Abc_NodeIsTravIdCurrent(pObj)")
	// check if the node has all fanouts marked
	// inlining of macro Abc_ObjForEachFanout
	for i = 0; i < abc.Abc_ObjFanoutNum(pObj); i++ {
		pFanout = abc.Abc_ObjFanout(pObj, i)
		if abc.Abc_NodeIsTravIdCurrent(pFanout) == 0 {
			break
		}
	}
	// if some of the fanouts are unmarked, add the node to the roots
	if i < abc.Abc_ObjFanoutNum(pObj) {
		Vec_PtrPushUnique(vRoots, pObj)
	} else { // otherwise, call recursively
		// inlining of macro Abc_ObjForEachFanout
		for i = 0; i < abc.Abc_ObjFanoutNum(pObj); i++ {
			pFanout = abc.Abc_ObjFanout(pObj, i)
			Res_WinFinalizeRoots_rec(pFanout, vRoots)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Finalizes the roots of the window.]

	Description [Roots of the window are the nodes that have at least
	one fanout that it not in the TFO of the leaves.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinFinalizeRoots(p *Res_Win_t) int32 {
	assert(abc.Abc_NodeIsTravIdCurrent(p.pNode) == 0, "!abc.Abc_NodeIsTravIdCurrent(p.pNode)")
	// mark the node with the old traversal ID
	abc.Abc_NodeSetTravIdCurrent(p.pNode)
	// recollect the roots
	Vec_PtrClear(p.vRoots)
	Res_WinFinalizeRoots_rec(p.pNode, p.vRoots)
	assert(Vec_PtrSize(p.vRoots) > 0, "Vec_PtrSize(p.vRoots) > 0")
	if Vec_PtrSize(p.vRoots) == 1 && Vec_PtrEntry(p.vRoots, 0) == p.pNode {
		return 0
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Recursively adds missing nodes and leaves.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinAddMissing_rec(p *Res_Win_t, pObj *Abc_Obj_t, nLevTravMin int32) {
	var pFanin *Abc_Obj_t
	var i int32
	// skip the already collected leaves, nodes, and branches
	if abc.Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return
	}
	// if this is not an internal node - make it a new branch
	if abc.Abc_NodeIsTravIdPrevious(pObj) == 0 {
		assert(Vec_PtrFind(p.vLeaves, pObj) == -1, "Vec_PtrFind(p.vLeaves, pObj) == -1")
		abc.Abc_NodeSetTravIdCurrent(pObj)
		Vec_PtrPush(p.vBranches, pObj)
		return
	}
	assert(abc.Abc_ObjIsNode(pObj) != 0, "abc.Abc_ObjIsNode(pObj)") // if this is a CI, then the window is incorrect!
	abc.Abc_NodeSetTravIdCurrent(pObj)
	// visit the fanins of the node
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < abc.Abc_ObjFaninNum(pObj); i++ {
		pFanin = abc.Abc_ObjFanin(pObj, i)
		Res_WinAddMissing_rec(p, pFanin, nLevTravMin)
	}
	// collect the node
	Vec_PtrPush(p.vNodes, pObj)
}

/*
*Function*************************************************************

	Synopsis    [Adds to the window nodes and leaves in the TFI of the roots.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinAddMissing(p *Res_Win_t) {
	var pObj *Abc_Obj_t
	var i int32
	// mark the leaves
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(p.pNode))
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(p.vLeaves); i++ {
		pObj = Vec_PtrEntry(p.vLeaves, i)
		abc.Abc_NodeSetTravIdCurrent(pObj)
	}
	// mark the already collected nodes
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(p.vNodes); i++ {
		pObj = Vec_PtrEntry(p.vNodes, i)
		abc.Abc_NodeSetTravIdCurrent(pObj)
	}
	// explore from the roots
	Vec_PtrClear(p.vBranches)
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(p.vRoots); i++ {
		pObj = Vec_PtrEntry(p.vRoots, i)
		Res_WinAddMissing_rec(p, pObj, p.nLevTravMin)
	}
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the window is trivial (without TFO).]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinIsTrivial(p *Res_Win_t) int32 {
	if Vec_PtrSize(p.vRoots) == 1 && Vec_PtrEntry(p.vRoots, 0) == p.pNode {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Computes the window.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinCompute(pNode *Abc_Obj_t, nWinTfiMax int32, nWinTfoMax int32, p *Res_Win_t) int32 {
	assert(abc.Abc_ObjIsNode(pNode) != 0, "abc.Abc_ObjIsNode(pNode)")
	assert(nWinTfiMax > 0 && nWinTfiMax < 10, "nWinTfiMax > 0 && nWinTfiMax < 10")
	assert(nWinTfoMax >= 0 && nWinTfoMax < 10, "nWinTfoMax >= 0 && nWinTfoMax < 10")

	// initialize the window
	p.pNode = pNode
	p.nWinTfiMax = nWinTfiMax
	p.nWinTfoMax = nWinTfoMax

	Vec_PtrClear(p.vBranches)
	Vec_PtrClear(p.vDivs)
	Vec_PtrClear(p.vRoots)
	Vec_PtrPush(p.vRoots, pNode)

	// compute the leaves
	if !(Res_WinCollectLeavesAndNodes(p) != 0) {
		return 0
	}

	// compute the candidate roots
	if p.nWinTfoMax > 0 && (Res_WinComputeRoots(p) != 0) {
		// mark the paths from the roots to the leaves
		Res_WinMarkPaths(p)
		// refine the roots and add branches and missing nodes
		if Res_WinFinalizeRoots(p) != 0 {
			Res_WinAddMissing(p)
		}
	}

	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
