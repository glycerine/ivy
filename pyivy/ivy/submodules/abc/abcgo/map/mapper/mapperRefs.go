package mapper

/**CFile****************************************************************

  FileName    [mapperRefs.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperRefs.h,v 1.0 2003/09/08 00:00:00 alanmi Exp $]

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

	Synopsis    [Reads the actual reference counter of a phase.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_NodeReadRefPhaseAct(pNode *Map_Node_t, fPhase int32) int32 {
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if pNode.pCutBest[0] != nil && pNode.pCutBest[1] != nil { // both assigned
		return pNode.nRefAct[fPhase]
	}
	assert(pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil, "pNode.pCutBest[0] != nil || pNode.pCutBest[1]") // at least one assigned
	return pNode.nRefAct[2]
}

/*
*Function*************************************************************

	Synopsis    [Reads the estimated reference counter of a phase.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_NodeReadRefPhaseEst(pNode *Map_Node_t, fPhase int32) float32 {
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if pNode.pCutBest[0] != nil && pNode.pCutBest[1] != nil { // both assigned
		return pNode.nRefEst[fPhase]
	}
	assert(pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil, "pNode.pCutBest[0] != nil || pNode.pCutBest[1]") // at least one assigned
	//    return pNode->nRefEst[0] + pNode->nRefEst[1];
	return pNode.nRefEst[2]
}

/*
*Function*************************************************************

	Synopsis    [Increments the actual reference counter of a phase.]

	Description [Returns the old reference counter.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_NodeIncRefPhaseAct(pNode *Map_Node_t, fPhase int32) int32 {
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if pNode.pCutBest[0] != nil && pNode.pCutBest[1] != nil { // both assigned
		return incInt(&pNode.nRefAct[fPhase])
	}
	assert(pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil, "pNode.pCutBest[0] != nil || pNode.pCutBest[1]") // at least one assigned
	return incInt(&pNode.nRefAct[2])
}

/*
*Function*************************************************************

	Synopsis    [Decrements the actual reference counter of a phase.]

	Description [Returns the new reference counter.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_NodeDecRefPhaseAct(pNode *Map_Node_t, fPhase int32) int32 {
	// inlining of macro Map_IsComplement
	assert(!Map_IsComplement(pNode), "node is regular")
	if pNode.pCutBest[0] != nil && pNode.pCutBest[1] != nil { // both assigned
		return decInt(&pNode.nRefAct[fPhase])
	}
	assert(pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil, "pNode.pCutBest[0] != nil || pNode.pCutBest[1]") // at least one assigned
	return decInt(&pNode.nRefAct[2])
}

/*
*Function*************************************************************

	Synopsis    [Sets the estimated reference counter for the PIs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingEstimateRefsInit(p *Map_Man_t) {
	var pNode *Map_Node_t
	var i int32
	for i = 0; i < p.vMapObjs.nSize; i++ {
		pNode = p.vMapObjs.pArray[i]
		//        pNode->nRefEst[0] = pNode->nRefEst[1] = ((float)pNode->nRefs)*(float)2.0;
		pNode.nRefEst[0], pNode.nRefEst[1], pNode.nRefEst[2] = float32(pNode.nRefs), float32(pNode.nRefs), float32(pNode.nRefs)
	}
}

/*
*Function*************************************************************

	Synopsis    [Sets the estimated reference counter.]

	Description [When this procedure is called for the first time,
	the reference counter is estimated from the AIG. Otherwise, it is
	a linear combination of reference counters in the last two iterations.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingEstimateRefs(p *Map_Man_t) {
	var pNode *Map_Node_t
	var i int32
	for i = 0; i < p.vMapObjs.nSize; i++ {
		pNode = p.vMapObjs.pArray[i]
		//        pNode->nRefEst[0] = (float)((2.0 * pNode->nRefEst[0] + 1.0 * pNode->nRefAct[0]) / 3.0);
		//        pNode->nRefEst[1] = (float)((2.0 * pNode->nRefEst[1] + 1.0 * pNode->nRefAct[1]) / 3.0);
		//        pNode->nRefEst[2] = (float)((2.0 * pNode->nRefEst[2] + 1.0 * pNode->nRefAct[2]) / 3.0);
		pNode.nRefEst[0] = float32(((3.0*pNode.nRefEst[0] + float32(pNode.nRefAct[0])) / 4.0))
		pNode.nRefEst[1] = float32(((3.0*pNode.nRefEst[1] + float32(pNode.nRefAct[1])) / 4.0))
		pNode.nRefEst[2] = float32(((3.0*pNode.nRefEst[2] + float32(pNode.nRefAct[2])) / 4.0))
	}
}

/*
*function*************************************************************

	synopsis    [Computes the area flow of the cut.]

	description [Computes the area flow of the cut if it is implemented using
	the best supergate with the best phase.]

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_CutGetAreaFlow(pCut *Map_Cut_t, fPhase int32) float32 {
	var pM *Map_Match_t = &pCut.M[fPhase]
	var pSuper *Map_Super_t = pM.pSuperBest
	var uPhaseTot uint32 = pM.uPhaseBest
	var pCutFanin *Map_Cut_t
	var aFlowRes, aFlowFanin, nRefs float32
	var i, fPinPhasePos int32

	// start the resulting area flow
	aFlowRes = pSuper.Area
	// iterate through the leaves
	for i = 0; i < pCut.nLeaves; i++ {
		// get the phase of this fanin
		fPinPhasePos = boolToInt((uPhaseTot & (1 << i)) == 0)
		// get the cut implementing this phase of the fanin
		pCutFanin = pCut.ppLeaves[i].pCutBest[fPinPhasePos]
		// if the cut is not available, we have to use the opposite phase
		if pCutFanin == nil {
			fPinPhasePos = 1 - fPinPhasePos
			pCutFanin = pCut.ppLeaves[i].pCutBest[fPinPhasePos]
		}
		aFlowFanin = pCutFanin.M[fPinPhasePos].AreaFlow // ignores the area of the interter
		// get the fanout count of the cut in the given phase
		nRefs = Map_NodeReadRefPhaseEst(pCut.ppLeaves[i], fPinPhasePos)
		// if the node does no fanout, assume fanout count equal to 1
		if nRefs == float32(0.0) {
			nRefs = float32(1.0)
		}
		// add the area flow due to the fanin
		aFlowRes += aFlowFanin / nRefs
	}
	pM.AreaFlow = aFlowRes
	return aFlowRes
}

/*
*function*************************************************************

	synopsis    [References or dereferences the cut.]

	description [This reference part is similar to Cudd_NodeReclaim().
	The dereference part is similar to Cudd_RecursiveDeref().]

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_CutRefDeref(pCut *Map_Cut_t, fPhase int32, fReference int32, fUpdateProf int32) float32 {
	var pNodeChild *Map_Node_t
	var pCutChild *Map_Cut_t
	var aArea float32
	var i, fPhaseChild int32
	//    int nRefs;

	// consider the elementary variable
	if pCut.nLeaves == 1 {
		return 0
	}
	// start the area of this cut
	aArea = Map_CutGetRootArea(pCut, fPhase)
	if fUpdateProf != 0 {
		if fReference != 0 {
			Mio_GateIncProfile2(pCut.M[fPhase].pSuperBest.pRoot)
		} else {
			Mio_GateDecProfile2(pCut.M[fPhase].pSuperBest.pRoot)
		}
	}
	// go through the children
	for i = 0; i < pCut.nLeaves; i++ {
		pNodeChild = pCut.ppLeaves[i]
		fPhaseChild = Map_CutGetLeafPhase(pCut, fPhase, i)
		// get the reference counter of the child
		/*
		   // this code does not take inverters into account
		   // the quality of area recovery seems to always be a little worse
		   if ( fReference )
		       nRefs = Map_NodeIncRefPhaseAct( pNodeChild, fPhaseChild );
		   else
		       nRefs = Map_NodeDecRefPhaseAct( pNodeChild, fPhaseChild );
		   assert( nRefs >= 0 );
		   // skip if the child was already reference before
		   if ( nRefs > 0 )
		       continue;
		*/

		if fReference != 0 {
			if pNodeChild.pCutBest[0] != nil && pNodeChild.pCutBest[1] != nil { // both phases are present

				// if this phase of the node is referenced, there is no recursive call
				pNodeChild.nRefAct[2]++
				if incInt(&pNodeChild.nRefAct[fPhaseChild]) > 0 {
					continue
				}

			} else { // only one phase is present

				// inverter should be added if the phase
				// (a) has no reference and (b) is implemented using other phase
				if incInt(&pNodeChild.nRefAct[fPhaseChild]) == 0 && pNodeChild.pCutBest[fPhaseChild] == nil {
					aArea += pNodeChild.p.pSuperLib.AreaInv
				}
				// if the node is referenced, there is no recursive call
				if incInt(&pNodeChild.nRefAct[2]) > 0 {
					continue
				}
			}

		} else {
			if pNodeChild.pCutBest[0] != nil && pNodeChild.pCutBest[1] != nil { // both phases are present

				// if this phase of the node is referenced, there is no recursive call
				decInt(&pNodeChild.nRefAct[2])
				if decInt(&pNodeChild.nRefAct[fPhaseChild]) > 0 {
					continue
				}

			} else { // only one phase is present

				// inverter should be added if the phase
				// (a) has no reference and (b) is implemented using other phase
				if decInt(&pNodeChild.nRefAct[fPhaseChild]) == 0 && pNodeChild.pCutBest[fPhaseChild] == nil {
					aArea += pNodeChild.p.pSuperLib.AreaInv
				}
				// if the node is referenced, there is no recursive call
				if decInt(&pNodeChild.nRefAct[2]) > 0 {
					continue
				}
			}
			assert(pNodeChild.nRefAct[fPhaseChild] >= 0, "pNodeChild.nRefAct[fPhaseChild] >= 0")
		}

		// get the child cut
		pCutChild = pNodeChild.pCutBest[fPhaseChild]
		// if the child does not have this phase mapped, take the opposite phase
		if pCutChild == nil {
			fPhaseChild = 1 - fPhaseChild
			pCutChild = pNodeChild.pCutBest[fPhaseChild]
		}
		// reference and compute area recursively
		aArea += Map_CutRefDeref(pCutChild, fPhaseChild, fReference, fUpdateProf)
	}
	return aArea
}

/*
*function*************************************************************

	synopsis    [Computes the exact area associated with the cut.]

	description [Assumes that the cut is referenced.]

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_CutGetAreaRefed(pCut *Map_Cut_t, fPhase int32) float32 {
	var aResult float32
	_ = Map_CutRefDeref(pCut, fPhase, 0, 0)       // dereference
	aResult = Map_CutRefDeref(pCut, fPhase, 1, 0) // reference
	//    assert( aResult == aResult2 );
	return aResult
}

/*
*function*************************************************************

	synopsis    [Computes the exact area associated with the cut.]

	description []

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_CutGetAreaDerefed(pCut *Map_Cut_t, fPhase int32) float32 {
	var aResult float32
	_ = Map_CutRefDeref(pCut, fPhase, 1, 0)       // reference
	aResult = Map_CutRefDeref(pCut, fPhase, 0, 0) // dereference
	//    assert( aResult == aResult2 );
	return aResult
}

/*
*function*************************************************************

	synopsis    [References the cut.]

	description []

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_CutRef(pCut *Map_Cut_t, fPhase int32, fProfile int32) float32 {
	return Map_CutRefDeref(pCut, fPhase, 1, fProfile) // reference
}

/*
*function*************************************************************

	synopsis    [Dereferences the cut.]

	description []

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_CutDeref(pCut *Map_Cut_t, fPhase int32, fProfile int32) float32 {
	return Map_CutRefDeref(pCut, fPhase, 0, fProfile) // dereference
}

/*
*Function*************************************************************

	Synopsis    [Computes actual reference counters.]

	Description [Collects the nodes used in the mapping in array pMan->vMapping.
	Nodes are collected in reverse topological order to facilitate the
	computation of required times.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingSetRefs_rec(pMan *Map_Man_t, pNode *Map_Node_t) {
	var pCut *Map_Cut_t
	var pNodeR *Map_Node_t
	var uPhase uint32
	var i, fPhase, fInvPin int32
	// get the regular node and its phase
	// inlining of macro Map_Regular
	pNodeR = Map_Regular(pNode)
	// inlining of macro Map_IsComplement
	fPhase = 1 - boolToInt(Map_IsComplement(pNode))
	pNodeR.nRefAct[2]++
	// quit if the node was already visited in this phase
	if incInt(&pNodeR.nRefAct[fPhase]) != 0 {
		return
	}
	// quit if this is a PI node
	if Map_NodeIsVar(pNodeR) != 0 {
		return
	}
	// propagate through buffer
	if Map_NodeIsBuf(pNodeR) != 0 {
		// inlining of macro Map_NotCond
		// inlining of macro Map_IsComplement
		Map_MappingSetRefs_rec(pMan, Map_NotCond(pNodeR.p1, Map_IsComplement(pNode)))
		return
	}
	assert(Map_NodeIsAnd(pNode) != 0, "Map_NodeIsAnd(pNode)")
	// get the cut implementing this or opposite polarity
	pCut = pNodeR.pCutBest[fPhase]
	if pCut == nil {
		fPhase = 1 - fPhase
		pCut = pNodeR.pCutBest[fPhase]
	}
	if pMan.fUseProfile != 0 {
		Mio_GateIncProfile2(pCut.M[fPhase].pSuperBest.pRoot)
	}
	// visit the transitive fanin
	uPhase = pCut.M[fPhase].uPhaseBest
	for i = 0; i < pCut.nLeaves; i++ {
		fInvPin = boolToInt((uPhase & (1 << i)) > 0)
		// inlining of macro Map_NotCond
		Map_MappingSetRefs_rec(pMan, Map_NotCond(pCut.ppLeaves[i], fInvPin != 0))
	}
}
func Map_MappingSetRefs(pMan *Map_Man_t) {
	var pNode *Map_Node_t
	var i int32
	if pMan.fUseProfile != 0 {
		Mio_LibraryCleanProfile2(pMan.pSuperLib.pGenlib)
	}
	// clean all references
	for i = 0; i < pMan.vMapObjs.nSize; i++ {
		pNode = pMan.vMapObjs.pArray[i]
		pNode.nRefAct[0] = 0
		pNode.nRefAct[1] = 0
		pNode.nRefAct[2] = 0
	}
	// visit nodes reachable from POs in the DFS order through the best cuts
	for i = 0; i < pMan.nOutputs; i++ {
		pNode = pMan.pOutputs[i]
		if Map_NodeIsConst(pNode) == 0 {
			Map_MappingSetRefs_rec(pMan, pNode)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Computes the array of mapping.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingGetArea(pMan *Map_Man_t) float32 {
	var pNode *Map_Node_t
	var Area float32 = 0.0
	var i int32
	if pMan.fUseProfile != 0 {
		Mio_LibraryCleanProfile2(pMan.pSuperLib.pGenlib)
	}
	for i = 0; i < pMan.vMapObjs.nSize; i++ {
		pNode = pMan.vMapObjs.pArray[i]
		if pNode.nRefAct[2] == 0 {
			continue
		}
		if Map_NodeIsBuf(pNode) != 0 {
			continue
		}
		// at least one phase has the best cut assigned
		assert(pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil, "pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil")
		// at least one phase is used in the mapping
		assert(pNode.nRefAct[0] > 0 || pNode.nRefAct[1] > 0, "pNode.nRefAct[0] > 0 || pNode.nRefAct[1] > 0")
		// compute the array due to the supergate
		if Map_NodeIsAnd(pNode) != 0 {
			// count area of the negative phase
			if pNode.pCutBest[0] != nil && (pNode.nRefAct[0] > 0 || pNode.pCutBest[1] == nil) {
				Area += pNode.pCutBest[0].M[0].pSuperBest.Area
				if pMan.fUseProfile != 0 {
					Mio_GateIncProfile2(pNode.pCutBest[0].M[0].pSuperBest.pRoot)
				}
			}
			// count area of the positive phase
			if pNode.pCutBest[1] != nil && (pNode.nRefAct[1] > 0 || pNode.pCutBest[0] == nil) {
				Area += pNode.pCutBest[1].M[1].pSuperBest.Area
				if pMan.fUseProfile != 0 {
					Mio_GateIncProfile2(pNode.pCutBest[1].M[1].pSuperBest.pRoot)
				}
			}
		}
		// count area of the interver if we need to implement one phase with another phase
		if (pNode.pCutBest[0] == nil && pNode.nRefAct[0] > 0) ||
			(pNode.pCutBest[1] == nil && pNode.nRefAct[1] > 0) {
			Area += pMan.pSuperLib.AreaInv
		}
	}
	// add buffers for each CO driven by a CI
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_IsComplement
		if Map_NodeIsVar(Map_Regular(pMan.pOutputs[i])) != 0 && !Map_IsComplement(pMan.pOutputs[i]) {
			Area += pMan.pSuperLib.AreaBuf
		}
	}
	return Area
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
