package res

/**CFile****************************************************************

  FileName    [resDivs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Collect divisors for the given window.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resDivs.c,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "resInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Res_WinMarkTfi( Res_Win_t * p );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Adds candidate divisors of the node to its window.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinDivisors(p *Res_Win_t, nLevDivMax int32) {
	var pObj, pFanout, pFanin *Abc_Obj_t
	var k, f, m int32

	// set the maximum level of the divisors
	p.nLevDivMax = nLevDivMax

	// mark the TFI with the current trav ID
	Abc_NtkIncrementTravId(p.pNode.pNtk)
	Res_WinMarkTfi(p)

	// mark with the current trav ID those nodes that should not be divisors:
	// (1) the node and its TFO
	// (2) the MFFC of the node
	// (3) the node's fanins (these are treated as a special case)
	Abc_NtkIncrementTravId(p.pNode.pNtk)
	Res_WinSweepLeafTfo_rec(p.pNode, p.nLevDivMax)
	Res_WinVisitMffc(p.pNode)
	// inlining of macro Abc_ObjForEachFanin
	for k = 0; k < Abc_ObjFaninNum(p.pNode); k++ {
		pObj = Abc_ObjFanin(p.pNode, k)
		Abc_NodeSetTravIdCurrent(pObj)
	}

	// at this point the nodes are marked with two trav IDs:
	// nodes to be collected as divisors are marked with previous trav ID
	// nodes to be avoided as divisors are marked with current trav ID

	// start collecting the divisors
	Vec_PtrClear(p.vDivs)
	// inlining of macro Vec_PtrForEachEntry
	for k = 0; k < Vec_PtrSize(p.vLeaves); k++ {
		pObj = Vec_PtrEntry(p.vLeaves, k)
		assert(int32(pObj.Level) >= p.nLevLeafMin, "int32(pObj.Level) >= p.nLevLeafMin")
		if !Abc_NodeIsTravIdPrevious(pObj) {
			continue
		}
		if int32(pObj.Level) > p.nLevDivMax {
			continue
		}
		Vec_PtrPush(p.vDivs, pObj)
	}
	// add the internal nodes to the data structure
	// inlining of macro Vec_PtrForEachEntry
	for k = 0; k < Vec_PtrSize(p.vNodes); k++ {
		pObj = Vec_PtrEntry(p.vNodes, k)
		if !Abc_NodeIsTravIdPrevious(pObj) {
			continue
		}
		if int32(pObj.Level) > p.nLevDivMax {
			continue
		}
		Vec_PtrPush(p.vDivs, pObj)
	}

	// explore the fanouts of already collected divisors
	p.nDivsPlus = 0
	// inlining of macro Vec_PtrForEachEntry
	for k = 0; k < Vec_PtrSize(p.vDivs); k++ {
		pObj = Vec_PtrEntry(p.vDivs, k)
		// consider fanouts of this node
		// inlining of macro Abc_ObjForEachFanout
		for f = 0; f < Abc_ObjFanoutNum(pObj); f++ {
			pFanout = Abc_ObjFanout(pObj, f)
			// stop if there are too many fanouts
			if f > 20 {
				break
			}
			// skip nodes that are already added
			if Abc_NodeIsTravIdPrevious(pFanout) {
				continue
			}
			// skip nodes in the TFO or in the MFFC of node
			if Abc_NodeIsTravIdCurrent(pFanout) {
				continue
			}
			// skip COs
			if !Abc_ObjIsNode(pFanout) {
				continue
			}
			// skip nodes with large level
			if int32(pFanout.Level) > p.nLevDivMax {
				continue
			}
			// skip nodes whose fanins are not divisors
			// inlining of macro Abc_ObjForEachFanin
			for m = 0; m < Abc_ObjFaninNum(pFanout); m++ {
				pFanin = Abc_ObjFanin(pFanout, m)
				if !Abc_NodeIsTravIdPrevious(pFanin) {
					break
				}
			}
			if m < Abc_ObjFaninNum(pFanout) {
				continue
			}
			// add the node to the divisors
			Vec_PtrPush(p.vDivs, pFanout)
			Vec_PtrPush(p.vNodes, pFanout)
			Abc_NodeSetTravIdPrevious(pFanout)
			p.nDivsPlus++
		}
	}
	/*
	   printf( "Node level = %d.  ", Abc_ObjLevel(p->pNode) );
	   Vec_PtrForEachEntryStart( Abc_Obj_t *, p->vDivs, pObj, k, Vec_PtrSize(p->vDivs)-p->nDivsPlus )
	       printf( "%d ", Abc_ObjLevel(pObj) );
	   printf( "\n" );
	*/
	//printf( "%d ", p->nDivsPlus );
}

/*
*Function*************************************************************

	Synopsis    [Marks the TFI cone of the node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinMarkTfi_rec(p *Res_Win_t, pObj *Abc_Obj_t) {
	var pFanin *Abc_Obj_t
	var i int32
	if Abc_NodeIsTravIdCurrent(pObj) {
		return
	}
	Abc_NodeSetTravIdCurrent(pObj)
	assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)")
	// visit the fanins of the node
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < Abc_ObjFaninNum(pObj); i++ {
		pFanin = Abc_ObjFanin(pObj, i)
		Res_WinMarkTfi_rec(p, pFanin)
	}
}

/*
*Function*************************************************************

	Synopsis    [Marks the TFI cone of the node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinMarkTfi(p *Res_Win_t) {
	var pObj *Abc_Obj_t
	var i int32
	// mark the leaves
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(p.vLeaves); i++ {
		pObj = Vec_PtrEntry(p.vLeaves, i)
		Abc_NodeSetTravIdCurrent(pObj)
	}
	// start from the node
	Res_WinMarkTfi_rec(p, p.pNode)
}

/*
*Function*************************************************************

	Synopsis    [Marks the TFO of the collected nodes up to the given level.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinSweepLeafTfo_rec(pObj *Abc_Obj_t, nLevelLimit int32) {
	var pFanout *Abc_Obj_t
	var i int32
	if Abc_ObjIsCo(pObj) || int32(pObj.Level) > nLevelLimit {
		return
	}
	if Abc_NodeIsTravIdCurrent(pObj) {
		return
	}
	Abc_NodeSetTravIdCurrent(pObj)
	// inlining of macro Abc_ObjForEachFanout
	for i = 0; i < Abc_ObjFanoutNum(pObj); i++ {
		pFanout = Abc_ObjFanout(pObj, i)
		Res_WinSweepLeafTfo_rec(pFanout, nLevelLimit)
	}
}

/*
*Function*************************************************************

	Synopsis    [Dereferences the node's MFFC.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_NodeDeref_rec(pNode *Abc_Obj_t) int32 {
	var pFanin *Abc_Obj_t
	var i int32
	var Counter int32 = 1
	if Abc_ObjIsCi(pNode) {
		return 0
	}
	Abc_NodeSetTravIdCurrent(pNode)
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < Abc_ObjFaninNum(pNode); i++ {
		pFanin = Abc_ObjFanin(pNode, i)
		assert(pFanin.vFanouts.nSize > 0, "pFanin.vFanouts.nSize > 0")
		pFanin.vFanouts.nSize--
		if pFanin.vFanouts.nSize == 0 {
			Counter += Res_NodeDeref_rec(pFanin)
		}
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [References the node's MFFC.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_NodeRef_rec(pNode *Abc_Obj_t) int32 {
	var pFanin *Abc_Obj_t
	var i int32
	var Counter int32 = 1
	if Abc_ObjIsCi(pNode) {
		return 0
	}
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < Abc_ObjFaninNum(pNode); i++ {
		pFanin = Abc_ObjFanin(pNode, i)
		wasUnreferenced := pFanin.vFanouts.nSize == 0
		pFanin.vFanouts.nSize++
		if wasUnreferenced {
			Counter += Res_NodeRef_rec(pFanin)
		}
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Labels MFFC of the node with the current trav ID.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_WinVisitMffc(pNode *Abc_Obj_t) int32 {
	var Count1, Count2 int32
	assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)")
	// dereference the node (mark with the current trav ID)
	Count1 = Res_NodeDeref_rec(pNode)
	// reference it back
	Count2 = Res_NodeRef_rec(pNode)
	assert(Count1 == Count2, "Count1 == Count2")
	return Count1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
