package if_

/**CFile****************************************************************

  FileName    [ifSeq.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Sequential mapping.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifSeq.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Prepares for sequential mapping by linking the latches.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManPrepareMappingSeq(p *If_Man_t) {
	var pObjLi, pObjLo *If_Obj_t

	// link the latch outputs (CIs) directly to the drivers of latch inputs (COs)
	for i := int32(0); i < p.pPars.nLatchesCi; i++ {
		pObjLi = If_ManLi(p, i)
		pObjLo = If_ManLo(p, i)
		pObjLo.pFanin0 = If_ObjFanin0(pObjLi)
		pObjLo.fCompl0 = uint32(If_ObjFaninC0(pObjLi))
	}
}

/*
*Function*************************************************************

	Synopsis    [Collects latches in the topological order.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManCollectLatches_rec(pObj *If_Obj_t, vLatches *Vec_Ptr_t) {
	if If_ObjIsLatch(pObj) == 0 {
		return
	}
	if pObj.fMark != 0 {
		return
	}
	pObj.fMark = 1
	If_ManCollectLatches_rec(pObj.pFanin0, vLatches)
	Vec_PtrPush(vLatches, pObj)
}

/*
*Function*************************************************************

	Synopsis    [Collects latches in the topological order.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManCollectLatches(p *If_Man_t) *Vec_Ptr_t {
	var vLatches *Vec_Ptr_t
	var pObj *If_Obj_t
	// collect latches
	vLatches = Vec_PtrAlloc(int(p.pPars.nLatchesCi))
	// inlining of macro If_ManForEachLatchOutput
	// inlining of macro Vec_PtrForEachEntryStartStop
	for i := If_ManCiNum(p) - p.pPars.nLatchesCi - p.pPars.nLatchesCiBox; i < If_ManCiNum(p)-p.pPars.nLatchesCiBox; i++ {
		pObj = Vec_PtrEntry(p.vCis, int(i)).(*If_Obj_t)
		If_ManCollectLatches_rec(pObj, vLatches)
	}
	// clean marks
	// inlining of macro Vec_PtrForEachEntry
	for i := int32(0); i < int32(Vec_PtrSize(vLatches)); i++ {
		pObj = Vec_PtrEntry(vLatches, int(i)).(*If_Obj_t)
		pObj.fMark = 0
	}
	assert(int32(Vec_PtrSize(vLatches)) == p.pPars.nLatchesCi, "Vec_PtrSize(vLatches) == p.pPars.nLatchesCi")
	return vLatches
}

/*
*Function*************************************************************

	Synopsis    [Performs one pass of l-value computation over all nodes.]

	Description [Experimentally it was found that checking POs changes
	is not enough to detect the convergence of l-values in the network.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManPerformMappingRoundSeq(p *If_Man_t, nIter int32) int32 {
	var pObj *If_Obj_t
	var clk abctime = Abc_Clock()
	var fVeryVerbose int32
	var fChange int32

	if nIter == 1 {
		// if some latches depend on PIs, update their values
		// inlining of macro Vec_PtrForEachEntry
		for i := int32(0); i < int32(Vec_PtrSize(p.vLatchOrder)); i++ {
			pObj = Vec_PtrEntry(p.vLatchOrder, int(i)).(*If_Obj_t)
			If_ObjSetLValue(pObj, If_ObjLValue(If_ObjFanin0(pObj))-float32(p.Period))
			If_ObjSetArrTime(pObj, If_ObjLValue(pObj))
		}
	}

	// map the internal nodes
	p.nCutsMerged = 0
	// inlining of macro If_ManForEachNode
	// inlining of macro If_ManForEachObj
	// inlining of macro Vec_PtrForEachEntry
	for i := int32(0); i < int32(Vec_PtrSize(p.vObjs)); i++ {
		pObj = Vec_PtrEntry(p.vObjs, int(i)).(*If_Obj_t)
		if pObj.Type == IF_AND {
			If_ObjPerformMappingAnd(p, pObj, 0, 0, 0)
			if pObj.fRepr != 0 {
				If_ObjPerformMappingChoice(p, pObj, 0, 0)
			}
		}
	}

	// postprocess the mapping
	//Abc_Print( 1, "Itereation %d: \n", nIter );
	// inlining of macro If_ManForEachNode
	// inlining of macro If_ManForEachObj
	// inlining of macro Vec_PtrForEachEntry
	for i := int32(0); i < int32(Vec_PtrSize(p.vObjs)); i++ {
		pObj = Vec_PtrEntry(p.vObjs, int(i)).(*If_Obj_t)
		if pObj.Type == IF_AND {
			// update the LValues stored separately
			if If_ObjLValue(pObj) < If_ObjCutBest(pObj).Delay-p.fEpsilon {
				If_ObjSetLValue(pObj, If_ObjCutBest(pObj).Delay)
				fChange = 1
			}
			//Abc_Print( 1, "%d ", (int)If_ObjLValue(pObj) );
			// reset the visit counters
			assert(pObj.nVisits == 0, "pObj.nVisits == 0")
			pObj.nVisits = pObj.nVisitsCopy
		}
	}
	//Abc_Print( 1, "\n" );

	// propagate LValues over the registers
	// inlining of macro Vec_PtrForEachEntry
	for i := int32(0); i < int32(Vec_PtrSize(p.vLatchOrder)); i++ {
		pObj = Vec_PtrEntry(p.vLatchOrder, int(i)).(*If_Obj_t)
		If_ObjSetLValue(pObj, If_ObjLValue(If_ObjFanin0(pObj))-float32(p.Period))
		If_ObjSetArrTime(pObj, If_ObjLValue(pObj))
	}

	// compute area and delay
	If_ManMarkMapping(p)
	if fVeryVerbose != 0 {
		p.RequiredGlo = If_ManDelayMax(p, 1)
		//        p->AreaGlo = If_ManScanMapping(p);
		Abc_Print(1, "S%d:  Fi = %6.2f. Del = %6.2f. Area = %8.2f. Cuts = %8d. ",
			nIter, float32(p.Period), p.RequiredGlo, p.AreaGlo, p.nCutsMerged)
		Abc_PrintTime(1, "T", Abc_Clock()-clk)
	}
	return fChange
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if retiming with this clock period is feasible.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManBinarySearchPeriod(p *If_Man_t) int32 {
	var pObj *If_Obj_t
	var c, fConverged int32
	//    int fResetRefs = 0;
	p.nAttempts++

	// reset initial LValues (PIs to 0; others to -inf)
	// inlining of macro If_ManForEachObj
	// inlining of macro Vec_PtrForEachEntry
	for i := int32(0); i < int32(Vec_PtrSize(p.vObjs)); i++ {
		pObj = Vec_PtrEntry(p.vObjs, int(i)).(*If_Obj_t)
		If_ObjSetLValue(pObj, float32(-IF_INFINITY))
		If_ObjSetArrTime(pObj, float32(-IF_INFINITY))
		// undo any previous mapping, except for CIs
		if If_ObjIsAnd(pObj) != 0 {
			If_ObjCutBest(pObj).nLeaves = 0
		}
	}
	pObj = If_ManConst1(p)
	If_ObjSetLValue(pObj, float32(0.0))
	If_ObjSetArrTime(pObj, float32(0.0))
	// inlining of macro If_ManForEachPi
	// inlining of macro Vec_PtrForEachEntryStop
	for i := int32(0); i < int32(If_ManCiNum(p))-p.pPars.nLatchesCi-p.pPars.nLatchesCiBox; i++ {
		pObj = Vec_PtrEntry(p.vCis, int(i)).(*If_Obj_t)
		pObj = If_ManCi(p, i)
		If_ObjSetLValue(pObj, float32(0.0))
		If_ObjSetArrTime(pObj, float32(0.0))
	}

	// update all values iteratively
	fConverged = 0
	for c = 1; c <= p.nMaxIters; c++ {
		if !(If_ManPerformMappingRoundSeq(p, c) != 0) {
			p.RequiredGlo = If_ManDelayMax(p, 1)
			fConverged = 1
			break
		}
		p.RequiredGlo = If_ManDelayMax(p, 1)
		//Abc_Print( 1, "Global = %d \n", (int)p->RequiredGlo );
		if p.RequiredGlo > float32(p.Period)+p.fEpsilon {
			break
		}
	}

	// report the results
	If_ManMarkMapping(p)
	if p.pPars.FVerbose != 0 {
		//        p->AreaGlo = If_ManScanMapping(p);
		Abc_Print(1, "Attempt = %2d.  Iters = %3d.  Area = %10.2f.  Fi = %6.2f.  ", p.nAttempts, c, p.AreaGlo, float32(p.Period))
		if fConverged != 0 {
			Abc_Print(1, "  Feasible")
		} else if c > p.nMaxIters {
			Abc_Print(1, "Infeasible (timeout)")
		} else {
			Abc_Print(1, "Infeasible")
		}
		Abc_Print(1, "\n")
	}
	return fConverged
}

/*
*Function*************************************************************

	Synopsis    [Performs binary search for the optimal clock period.]

	Description [Assumes that FiMin is infeasible while FiMax is feasible.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManBinarySearch_rec(p *If_Man_t, FiMin int32, FiMax int32) int32 {
	assert(FiMin < FiMax, "FiMin < FiMax")
	if FiMin+1 == FiMax {
		return FiMax
	}
	// compute the median
	p.Period = FiMin + (FiMax-FiMin)/2
	if If_ManBinarySearchPeriod(p) != 0 {
		return If_ManBinarySearch_rec(p, FiMin, p.Period) // Median is feasible
	} else {
		return If_ManBinarySearch_rec(p, p.Period, FiMax)
	} // Median is infeasible
}

/*
*Function*************************************************************

	Synopsis    [Performs sequential mapping.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManPerformMappingSeqPost(p *If_Man_t) {
	var pObjLi, pObjLo, pObj *If_Obj_t
	assert(false, "0")

	// set arrival times
	assert(p.pPars.pTimesArr != nil, "p.pPars.pTimesArr != nil")
	// inlining of macro If_ManForEachLatchOutput
	// inlining of macro Vec_PtrForEachEntryStartStop
	for i := int32(If_ManCiNum(p)) - p.pPars.nLatchesCi - p.pPars.nLatchesCiBox; i < int32(If_ManCiNum(p))-p.pPars.nLatchesCiBox; i++ {
		pObjLo = Vec_PtrEntry(p.vCis, int(i)).(*If_Obj_t)
		p.pPars.pTimesArr[i] = If_ObjLValue(pObjLo)
	}

	// set the required times
	assert(p.pPars.pTimesReq == nil, "p.pPars.pTimesReq == nil")
	// inlining of macro ABC_ALLOC
	p.pPars.pTimesReq = make([]float32, If_ManCoNum(p))
	// inlining of macro If_ManForEachPo
	// inlining of macro Vec_PtrForEachEntryStartStop
	for i := p.pPars.nLatchesCoBox; i < int32(If_ManCoNum(p))-p.pPars.nLatchesCo; i++ {
		pObj = Vec_PtrEntry(p.vCos, int(i)).(*If_Obj_t)
		p.pPars.pTimesReq[i] = p.RequiredGlo2
	}
	// inlining of macro If_ManForEachLatchInput
	// inlining of macro Vec_PtrForEachEntryStart
	for i := int32(If_ManCoNum(p)) - p.pPars.nLatchesCo; i < int32(Vec_PtrSize(p.vCos)); i++ {
		pObjLi = Vec_PtrEntry(p.vCos, int(i)).(*If_Obj_t)
		p.pPars.pTimesReq[i] = If_ObjLValue(If_ObjFanin0(pObjLi))
	}

	// undo previous mapping
	// inlining of macro If_ManForEachObj
	// inlining of macro Vec_PtrForEachEntry
	for i := 0; i < Vec_PtrSize(p.vObjs); i++ {
		pObj = Vec_PtrEntry(p.vObjs, int(i)).(*If_Obj_t)
		if If_ObjIsAnd(pObj) != 0 {
			If_ObjCutBest(pObj).nLeaves = 0
		}
	}

	// map again combinationally
	//    p->pPars->fSeqMap = 0;
	If_ManPerformMappingComb(p)
	// p->pPars->fSeqMap = 1;
}

/*
*Function*************************************************************

	Synopsis    [Performs sequential mapping.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManPerformMappingSeq(p *If_Man_t) int32 {
	var clkTotal abctime = Abc_Clock()
	var PeriodBest int32

	p.SortMode = 0

	// perform combinational mapping to get the upper bound on the clock period
	If_ManPerformMappingRound(p, 1, 0, 0, 1, "")
	p.RequiredGlo = If_ManDelayMax(p, 0)
	p.RequiredGlo2 = p.RequiredGlo

	// set direct linking of latches with their inputs
	If_ManPrepareMappingSeq(p)

	// collect latches
	p.vLatchOrder = If_ManCollectLatches(p)

	// set parameters
	p.nCutsUsed = p.pPars.NCutsMax
	p.nAttempts = 0
	p.nMaxIters = 50
	p.Period = int32(p.RequiredGlo)

	// make sure the clock period works
	if !(If_ManBinarySearchPeriod(p) != 0) {
		Abc_Print(1, "If_ManPerformMappingSeq(): The upper bound on the clock period cannot be computed.\n")
		return 0
	}

	// perform binary search
	PeriodBest = If_ManBinarySearch_rec(p, 0, p.Period)

	// recompute the best l-values
	if p.Period != PeriodBest {
		p.Period = PeriodBest
		if !(If_ManBinarySearchPeriod(p) != 0) {
			Abc_Print(1, "If_ManPerformMappingSeq(): The final clock period cannot be confirmed.\n")
			return 0
		}
	}
	//    if ( p->pPars->fVerbose )
	{
		Abc_Print(1, "The best clock period is %3d.  ", p.Period)
		Abc_PrintTime(1, "Time", Abc_Clock()-clkTotal)
	}
	p.RequiredGlo = float32((PeriodBest))

	// postprocess it using combinational mapping
	If_ManPerformMappingSeqPost(p)
	s_MappingTime = Abc_Clock() - clkTotal
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
