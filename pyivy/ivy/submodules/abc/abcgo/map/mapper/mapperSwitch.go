package mapper

/**CFile****************************************************************

  FileName    [mapperSwitch.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: mapperSwitch.h,v 1.0 2003/09/08 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static float Map_SwitchCutRefDeref( Map_Node_t * pNode, Map_Cut_t * pCut, int fPhase, int fReference );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*function*************************************************************

	synopsis    [Computes the exact area associated with the cut.]

	description []

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_SwitchCutGetDerefed(pNode *Map_Node_t, pCut *Map_Cut_t, fPhase int32) float32 {
	var aResult float32
	//    assert( pNode->Switching > 0 );
	_ = Map_SwitchCutRefDeref(pNode, pCut, fPhase, 1)       // reference
	aResult = Map_SwitchCutRefDeref(pNode, pCut, fPhase, 0) // dereference
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
func Map_SwitchCutRef(pNode *Map_Node_t, pCut *Map_Cut_t, fPhase int32) float32 {
	return Map_SwitchCutRefDeref(pNode, pCut, fPhase, 1) // reference
}

/*
*function*************************************************************

	synopsis    [References the cut.]

	description []

	sideeffects []

	seealso     []

**********************************************************************
*/
func Map_SwitchCutDeref(pNode *Map_Node_t, pCut *Map_Cut_t, fPhase int32) float32 {
	return Map_SwitchCutRefDeref(pNode, pCut, fPhase, 0) // dereference
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
func Map_SwitchCutRefDeref(pNode *Map_Node_t, pCut *Map_Cut_t, fPhase int32, fReference int32) float32 {
	var pNodeChild *Map_Node_t
	var pCutChild *Map_Cut_t
	var aSwitchActivity float32
	var i, fPhaseChild int32

	// start switching activity for the node
	aSwitchActivity = pNode.Switching
	// consider the elementary variable
	if pCut.nLeaves == 1 {
		return aSwitchActivity
	}

	// go through the children
	assert(pCut.M[fPhase].pSuperBest != nil, "pCut.M[fPhase].pSuperBest")
	for i = 0; i < pCut.nLeaves; i++ {
		pNodeChild = pCut.ppLeaves[i]
		fPhaseChild = Map_CutGetLeafPhase(pCut, fPhase, i)
		// get the reference counter of the child

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
					aSwitchActivity += pNodeChild.Switching
				} // inverter switches the same as the node
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
					aSwitchActivity += pNodeChild.Switching
				} // inverter switches the same as the node
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
		aSwitchActivity += Map_SwitchCutRefDeref(pNodeChild, pCutChild, fPhaseChild, fReference)
	}
	return aSwitchActivity
}

/*
*Function*************************************************************

	Synopsis    [Computes the array of mapping.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_MappingGetSwitching(pMan *Map_Man_t) float32 {
	var pNode *Map_Node_t
	var Switch float32 = 0.0
	var i int32
	for i = 0; i < pMan.vMapObjs.nSize; i++ {
		pNode = pMan.vMapObjs.pArray[i]
		if pNode.nRefAct[2] == 0 {
			continue
		}
		// at least one phase has the best cut assigned
		assert(pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil, "pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil")
		// at least one phase is used in the mapping
		assert(pNode.nRefAct[0] > 0 || pNode.nRefAct[1] > 0, "pNode.nRefAct[0] > 0 || pNode.nRefAct[1] > 0")
		// compute the array due to the supergate
		if Map_NodeIsAnd(pNode) != 0 {
			// count switching of the negative phase
			if pNode.pCutBest[0] != nil && (pNode.nRefAct[0] > 0 || pNode.pCutBest[1] == nil) {
				Switch += pNode.Switching
			}
			// count switching of the positive phase
			if pNode.pCutBest[1] != nil && (pNode.nRefAct[1] > 0 || pNode.pCutBest[0] == nil) {
				Switch += pNode.Switching
			}
		}
		// count switching of the interver if we need to implement one phase with another phase
		if (pNode.pCutBest[0] == nil && pNode.nRefAct[0] > 0) ||
			(pNode.pCutBest[1] == nil && pNode.nRefAct[1] > 0) {
			Switch += pNode.Switching
		} // inverter switches the same as the node
	}
	// add buffers for each CO driven by a CI
	for i = 0; i < pMan.nOutputs; i++ {
		// inlining of macro Map_IsComplement
		if Map_NodeIsVar(Map_Regular(pMan.pOutputs[i])) != 0 && !Map_IsComplement(pMan.pOutputs[i]) {
			Switch += pMan.pOutputs[i].Switching
		}
	}
	return Switch
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
