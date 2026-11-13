package mapper

import "fmt"

/**CFile****************************************************************

  FileName    [mapperTime.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperTime.c,v 1.3 2005/03/02 02:35:54 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Computes the maximum arrival times.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_TimeComputeArrivalMax(p *Map_Man_t) float32 {
	var tReqMax, tReq float32
	var i, fPhase int32
	// get the critical PO arrival time
	tReqMax = -MAP_FLOAT_LARGE
	for i = 0; i < p.nOutputs; i++ {
		if Map_NodeIsConst(p.pOutputs[i]) != 0 {
			continue
		}
		// inlining of macro Map_IsComplement
		fPhase = 1 - boolToInt(Map_IsComplement(p.pOutputs[i]))
		// inlining of macro Map_Regular
		tReq = Map_Regular(p.pOutputs[i]).tArrival[fPhase].Worst
		// inlining of macro MAP_MAX

		tReqMax = max(tReqMax, tReq)
	}
	return tReqMax
}

/*
*Function*************************************************************

	Synopsis    [Computes the arrival times of the cut.]

	Description [Computes the arrival times of the cut if it is implemented using
	the given supergate with the given phase. Uses the constraint-type specification
	of rise/fall arrival times.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_TimeCutComputeArrival(pNode *Map_Node_t, pCut *Map_Cut_t, fPhase int32, tWorstLimit float32) float32 {
	var pM *Map_Match_t = &pCut.M[fPhase]
	var pSuper *Map_Super_t = pM.pSuperBest
	var uPhaseTot uint32 = pM.uPhaseBest
	var ptArrRes *Map_Time_t = &pM.tArrive
	var ptArrIn *Map_Time_t
	var fPinPhase int32
	var tDelay, tExtra float32
	var i int32

	tExtra = nodeDelay(pNode.p, pNode)
	ptArrRes.Rise, ptArrRes.Fall = 0, 0
	ptArrRes.Worst = MAP_FLOAT_LARGE
	for i = pCut.nLeaves - 1; i >= 0; i-- {
		// get the phase of the given pin
		fPinPhase = boolToInt((uPhaseTot & (1 << i)) == 0)
		ptArrIn = &pCut.ppLeaves[i].tArrival[fPinPhase]

		// get the rise of the output due to rise of the inputs
		if pSuper.tDelaysR[i].Rise > 0 {
			tDelay = ptArrIn.Rise + pSuper.tDelaysR[i].Rise + tExtra
			if tDelay > tWorstLimit {
				return MAP_FLOAT_LARGE
			}
			if ptArrRes.Rise < tDelay {
				ptArrRes.Rise = tDelay
			}
		}

		// get the rise of the output due to fall of the inputs
		if pSuper.tDelaysR[i].Fall > 0 {
			tDelay = ptArrIn.Fall + pSuper.tDelaysR[i].Fall + tExtra
			if tDelay > tWorstLimit {
				return MAP_FLOAT_LARGE
			}
			if ptArrRes.Rise < tDelay {
				ptArrRes.Rise = tDelay
			}
		}

		// get the fall of the output due to rise of the inputs
		if pSuper.tDelaysF[i].Rise > 0 {
			tDelay = ptArrIn.Rise + pSuper.tDelaysF[i].Rise + tExtra
			if tDelay > tWorstLimit {
				return MAP_FLOAT_LARGE
			}
			if ptArrRes.Fall < tDelay {
				ptArrRes.Fall = tDelay
			}
		}

		// get the fall of the output due to fall of the inputs
		if pSuper.tDelaysF[i].Fall > 0 {
			tDelay = ptArrIn.Fall + pSuper.tDelaysF[i].Fall + tExtra
			if tDelay > tWorstLimit {
				return MAP_FLOAT_LARGE
			}
			if ptArrRes.Fall < tDelay {
				ptArrRes.Fall = tDelay
			}
		}
	}
	// return the worst-case of rise/fall arrival times
	// inlining of macro MAP_MAX

	if (ptArrRes.Rise) > (ptArrRes.Fall) {
		ptArrRes.Worst = (ptArrRes.Rise)
	} else {
		ptArrRes.Worst = (ptArrRes.Fall)
	}
	return ptArrRes.Worst
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_TimePropagateRequiredPhase(p *Map_Man_t, pNode *Map_Node_t, fPhase int32) {
	var ptReqIn, ptReqOut *Map_Time_t
	var pCut *Map_Cut_t
	var pSuper *Map_Super_t
	var tNewReqTime, tExtra float32
	var uPhase uint32
	var fPinPhase, i int32

	tExtra = nodeDelay(pNode.p, pNode)
	// get the cut to be propagated
	pCut = pNode.pCutBest[fPhase]
	assert(pCut != nil, "pCut != nil")
	// get the supergate and its polarity
	pSuper = pCut.M[fPhase].pSuperBest
	uPhase = pCut.M[fPhase].uPhaseBest
	// get the required time of the output of the supergate
	ptReqOut = &pNode.tRequired[fPhase]
	// set the required time of the children
	for i = 0; i < pCut.nLeaves; i++ {
		// get the phase of the given pin of the supergate
		fPinPhase = boolToInt((uPhase & (1 << i)) == 0)
		ptReqIn = &pCut.ppLeaves[i].tRequired[fPinPhase]
		assert(pCut.ppLeaves[i].nRefAct[2] > 0, "pCut.ppLeaves[i].nRefAct[2] > 0")

		// get the rise of the output due to rise of the inputs
		//            if ( ptArrOut->Rise < ptArrIn->Rise + pSuper->tDelaysR[i].Rise )
		//                ptArrOut->Rise = ptArrIn->Rise + pSuper->tDelaysR[i].Rise;
		if pSuper.tDelaysR[i].Rise > 0 {
			tNewReqTime = ptReqOut.Rise - pSuper.tDelaysR[i].Rise - tExtra
			// inlining of macro MAP_MIN

			if (ptReqIn.Rise) < (tNewReqTime) {
				ptReqIn.Rise = (ptReqIn.Rise)
			} else {
				ptReqIn.Rise = (tNewReqTime)
			}
		}

		// get the rise of the output due to fall of the inputs
		//            if ( ptArrOut->Rise < ptArrIn->Fall + pSuper->tDelaysR[i].Fall )
		//                ptArrOut->Rise = ptArrIn->Fall + pSuper->tDelaysR[i].Fall;
		if pSuper.tDelaysR[i].Fall > 0 {
			tNewReqTime = ptReqOut.Rise - pSuper.tDelaysR[i].Fall - tExtra
			// inlining of macro MAP_MIN

			if (ptReqIn.Fall) < (tNewReqTime) {
				ptReqIn.Fall = (ptReqIn.Fall)
			} else {
				ptReqIn.Fall = (tNewReqTime)
			}
		}

		// get the fall of the output due to rise of the inputs
		//            if ( ptArrOut->Fall < ptArrIn->Rise + pSuper->tDelaysF[i].Rise )
		//                ptArrOut->Fall = ptArrIn->Rise + pSuper->tDelaysF[i].Rise;
		if pSuper.tDelaysF[i].Rise > 0 {
			tNewReqTime = ptReqOut.Fall - pSuper.tDelaysF[i].Rise - tExtra
			// inlining of macro MAP_MIN

			if (ptReqIn.Rise) < (tNewReqTime) {
				ptReqIn.Rise = (ptReqIn.Rise)
			} else {
				ptReqIn.Rise = (tNewReqTime)
			}
		}

		// get the fall of the output due to fall of the inputs
		//            if ( ptArrOut->Fall < ptArrIn->Fall + pSuper->tDelaysF[i].Fall )
		//                ptArrOut->Fall = ptArrIn->Fall + pSuper->tDelaysF[i].Fall;
		if pSuper.tDelaysF[i].Fall > 0 {
			tNewReqTime = ptReqOut.Fall - pSuper.tDelaysF[i].Fall - tExtra
			// inlining of macro MAP_MIN

			if (ptReqIn.Fall) < (tNewReqTime) {
				ptReqIn.Fall = (ptReqIn.Fall)
			} else {
				ptReqIn.Fall = (tNewReqTime)
			}
		}
	}

	// compare the required times with the arrival times
	//    assert( pNode->tArrival[fPhase].Rise < ptReqOut->Rise + p->fEpsilon );
	//    assert( pNode->tArrival[fPhase].Fall < ptReqOut->Fall + p->fEpsilon );
}
func Map_MatchComputeReqTimes(pCut *Map_Cut_t, fPhase int32, ptArrRes *Map_Time_t) float32 {
	var ptArrIn *Map_Time_t
	var pSuper *Map_Super_t
	var uPhaseTot uint32
	var fPinPhase, i int32
	var tDelay float32

	// get the supergate and the phase
	pSuper = pCut.M[fPhase].pSuperBest
	uPhaseTot = pCut.M[fPhase].uPhaseBest

	// propagate the arrival times
	ptArrRes.Rise, ptArrRes.Fall = -MAP_FLOAT_LARGE, -MAP_FLOAT_LARGE
	for i = 0; i < pCut.nLeaves; i++ {
		// get the phase of the given pin
		fPinPhase = boolToInt((uPhaseTot & (1 << i)) == 0)
		ptArrIn = &pCut.ppLeaves[i].tRequired[fPinPhase]
		//        assert( ptArrIn->Worst < MAP_FLOAT_LARGE );

		// get the rise of the output due to rise of the inputs
		if pSuper.tDelaysR[i].Rise > 0 {
			tDelay = ptArrIn.Rise + pSuper.tDelaysR[i].Rise
			if ptArrRes.Rise < tDelay {
				ptArrRes.Rise = tDelay
			}
		}

		// get the rise of the output due to fall of the inputs
		if pSuper.tDelaysR[i].Fall > 0 {
			tDelay = ptArrIn.Fall + pSuper.tDelaysR[i].Fall
			if ptArrRes.Rise < tDelay {
				ptArrRes.Rise = tDelay
			}
		}

		// get the fall of the output due to rise of the inputs
		if pSuper.tDelaysF[i].Rise > 0 {
			tDelay = ptArrIn.Rise + pSuper.tDelaysF[i].Rise
			if ptArrRes.Fall < tDelay {
				ptArrRes.Fall = tDelay
			}
		}

		// get the fall of the output due to fall of the inputs
		if pSuper.tDelaysF[i].Fall > 0 {
			tDelay = ptArrIn.Fall + pSuper.tDelaysF[i].Fall
			if ptArrRes.Fall < tDelay {
				ptArrRes.Fall = tDelay
			}
		}
	}
	// return the worst-case of rise/fall arrival times
	// inlining of macro MAP_MAX
	return max(ptArrRes.Rise, ptArrRes.Fall)
}

/*
*Function*************************************************************

	Synopsis    [Computes the required times of all nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_TimePropagateRequired(p *Map_Man_t) {
	var pNode *Map_Node_t
	//Map_Time_t tReqOutTest, * ptReqOutTest = &tReqOutTest;
	var ptReqIn, ptReqOut *Map_Time_t
	var fPhase, k int32

	// go through the nodes in the reverse topological order
	for k = p.vMapObjs.nSize - 1; k >= 0; k-- {
		pNode = p.vMapObjs.pArray[k]
		if pNode.nRefAct[2] == 0 {
			continue
		}

		// propagate required times through the buffer
		if Map_NodeIsBuf(pNode) != 0 {
			assert(pNode.p2 == nil, "pNode.p2 == nil")
			// inlining of macro Map_Regular
			// inlining of macro Map_IsComplement
			phase := boolToInt(Map_IsComplement(pNode.p1))
			Map_Regular(pNode.p1).tRequired[phase] = pNode.tRequired[0]
			// inlining of macro Map_Regular
			// inlining of macro Map_IsComplement
			Map_Regular(pNode.p1).tRequired[1-phase] = pNode.tRequired[1]
			continue
		}

		// this computation works for regular nodes only
		// inlining of macro Map_IsComplement
		assert(!Map_IsComplement(pNode), "node is regular")
		// at least one phase should be mapped
		assert(pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil, "pNode.pCutBest[0] != nil || pNode.pCutBest[1] != nil")
		// the node should be used in the currently assigned mapping
		assert(pNode.nRefAct[0] > 0 || pNode.nRefAct[1] > 0, "pNode.nRefAct[0] > 0 || pNode.nRefAct[1] > 0")

		// if one of the cuts is not given, project the required times from the other cut
		if pNode.pCutBest[0] == nil || pNode.pCutBest[1] == nil {
			//            assert( 0 );
			// get the missing phase
			fPhase = boolToInt(pNode.pCutBest[1] == nil)
			// check if the missing phase is needed in the mapping
			if pNode.nRefAct[fPhase] > 0 {
				// get the pointers to the required times of the missing phase
				ptReqOut = &pNode.tRequired[fPhase]
				//                assert( ptReqOut->Fall < MAP_FLOAT_LARGE );
				// get the pointers to the required times of the present phase
				ptReqIn = &pNode.tRequired[1-fPhase]
				// propagate the required times from the missing phase to the present phase
				//            tArrInv.Fall  = pMatch->tArrive.Rise + p->pSuperLib->tDelayInv.Fall;
				//            tArrInv.Rise  = pMatch->tArrive.Fall + p->pSuperLib->tDelayInv.Rise;
				// inlining of macro MAP_MIN

				ptReqIn.Fall = min(ptReqIn.Fall, ptReqOut.Rise-p.pSuperLib.tDelayInv.Rise)
				// inlining of macro MAP_MIN

				ptReqIn.Rise = min(ptReqIn.Rise, ptReqOut.Fall-p.pSuperLib.tDelayInv.Fall)
			}
		}

		// finalize the worst case computation
		// inlining of macro MAP_MIN
		pNode.tRequired[0].Worst = min(pNode.tRequired[0].Fall, pNode.tRequired[0].Rise)
		// inlining of macro MAP_MIN
		pNode.tRequired[1].Worst = min(pNode.tRequired[1].Fall, pNode.tRequired[1].Rise)

		// skip the PIs
		if Map_NodeIsAnd(pNode) == 0 {
			continue
		}

		// propagate required times of different phases of the node
		// the ordering of phases does not matter since they are mapped independently
		if pNode.pCutBest[0] != nil && pNode.tRequired[0].Worst < MAP_FLOAT_LARGE {
			Map_TimePropagateRequiredPhase(p, pNode, 0)
		}
		if pNode.pCutBest[1] != nil && pNode.tRequired[1].Worst < MAP_FLOAT_LARGE {
			Map_TimePropagateRequiredPhase(p, pNode, 1)
		}
	}
	/*
	   	// in the end, we verify the required times
	   	// for this, we compute the arrival times of the outputs of each phase
	   	// of the supergates using the fanins' required times as the fanins' arrival times
	   	// the resulting arrival time of the supergate should be less than the actual required time
	   	for ( k = p->vMapObjs->nSize - 1; k >= 0; k-- )
	   	{
	   	    pNode = p->vMapObjs->pArray[k];
	   	    if ( pNode->nRefAct[2] == 0 )
	   	        continue;
	   	    if ( Map_NodeIsAnd(pNode) == 0 )
	   	        continue;
	   	    // verify that the required times are propagated correctly

	   //        if ( pNode->pCutBest[0] && (pNode->nRefAct[0] > 0 || pNode->pCutBest[1] == NULL) )

	   	if ( pNode->pCutBest[0] && pNode->tRequired[0].Worst < MAP_FLOAT_LARGE/2 )
	   	{
	   	    Map_MatchComputeReqTimes( pNode->pCutBest[0], 0, ptReqOutTest );

	   //            assert( ptReqOutTest->Rise < pNode->tRequired[0].Rise + p->fEpsilon );
	   //            assert( ptReqOutTest->Fall < pNode->tRequired[0].Fall + p->fEpsilon );

	   	}

	   //        if ( pNode->pCutBest[1] && (pNode->nRefAct[1] > 0 || pNode->pCutBest[0] == NULL) )

	   	if ( pNode->pCutBest[1] && pNode->tRequired[1].Worst < MAP_FLOAT_LARGE/2 )
	   	{
	   	    Map_MatchComputeReqTimes( pNode->pCutBest[1], 1, ptReqOutTest );

	   //            assert( ptReqOutTest->Rise < pNode->tRequired[1].Rise + p->fEpsilon );
	   //            assert( ptReqOutTest->Fall < pNode->tRequired[1].Fall + p->fEpsilon );

	   	    }
	   	}
	*/
}
func Map_TimeComputeRequiredGlobal(p *Map_Man_t) {
	var fUseConMan int32 = boolToInt(Scl_ConIsRunning() && Scl_ConHasOutReqs())
	var ptTime, ptTimeA *Map_Time_t
	var fPhase, i int32
	// update the required times according to the target
	p.fRequiredGlo = Map_TimeComputeArrivalMax(p)
	if p.DelayTarget != -1 {
		if p.fRequiredGlo > p.DelayTarget+p.fEpsilon {
			if p.fMappingMode == 1 {
				fmt.Printf("Cannot meet the target required times (%4.2f). Continue anyway.\n", p.DelayTarget)
			}

		} else if p.fRequiredGlo < p.DelayTarget-p.fEpsilon {
			if p.fMappingMode == 1 && p.fVerbose != 0 {
				fmt.Printf("Relaxing the required times from (%4.2f) to the target (%4.2f).\n", p.fRequiredGlo, p.DelayTarget)
			}
			p.fRequiredGlo = p.DelayTarget
		}
	}
	// clean the required times
	for i = 0; i < p.vMapObjs.nSize; i++ {
		p.vMapObjs.pArray[i].tRequired[0].Rise = MAP_FLOAT_LARGE
		p.vMapObjs.pArray[i].tRequired[0].Fall = MAP_FLOAT_LARGE
		p.vMapObjs.pArray[i].tRequired[0].Worst = MAP_FLOAT_LARGE
		p.vMapObjs.pArray[i].tRequired[1].Rise = MAP_FLOAT_LARGE
		p.vMapObjs.pArray[i].tRequired[1].Fall = MAP_FLOAT_LARGE
		p.vMapObjs.pArray[i].tRequired[1].Worst = MAP_FLOAT_LARGE
	}
	// set the required times for the POs
	for i = 0; i < p.nOutputs; i++ {
		// inlining of macro Map_IsComplement
		fPhase = 1 - boolToInt(Map_IsComplement(p.pOutputs[i]))
		// inlining of macro Map_Regular
		ptTime = &Map_Regular(p.pOutputs[i]).tRequired[fPhase]
		// inlining of macro Map_Regular
		ptTimeA = &Map_Regular(p.pOutputs[i]).tArrival[fPhase]

		if fUseConMan != 0 {
			var Value float32 = Scl_ConGetOutReqFloat(i)
			// if external required time can be achieved, use it
			if Value > 0 && ptTimeA.Worst <= Value { //&& Value <= p->fRequiredGlo )
				ptTime.Rise, ptTime.Fall, ptTime.Worst = Value, Value, Value
				// if external required cannot be achieved, set the earliest possible arrival time
			} else if Value > 0 && ptTimeA.Worst > Value {
				ptTime.Rise, ptTime.Fall, ptTime.Worst = ptTimeA.Worst, ptTimeA.Worst, ptTimeA.Worst
				// otherwise, set the global required time
			} else {
				ptTime.Rise, ptTime.Fall, ptTime.Worst = p.fRequiredGlo, p.fRequiredGlo, p.fRequiredGlo
			}

		} else {
			// if external required time can be achieved, use it
			if p.pOutputRequireds != nil && p.pOutputRequireds[i].Worst > 0 && ptTimeA.Worst <= p.pOutputRequireds[i].Worst { //&& p->pOutputRequireds[i].Worst <= p->fRequiredGlo )
				ptTime.Rise, ptTime.Fall, ptTime.Worst = p.pOutputRequireds[i].Worst, p.pOutputRequireds[i].Worst, p.pOutputRequireds[i].Worst
				// if external required cannot be achieved, set the earliest possible arrival time
			} else if p.pOutputRequireds != nil && p.pOutputRequireds[i].Worst > 0 && ptTimeA.Worst > p.pOutputRequireds[i].Worst {
				ptTime.Rise, ptTime.Fall, ptTime.Worst = ptTimeA.Worst, ptTimeA.Worst, ptTimeA.Worst
				// otherwise, set the global required time
			} else {
				ptTime.Rise, ptTime.Fall, ptTime.Worst = p.fRequiredGlo, p.fRequiredGlo, p.fRequiredGlo
			}
		}
	}
	// visit nodes in the reverse topological order
	Map_TimePropagateRequired(p)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
