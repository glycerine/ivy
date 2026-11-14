package wlc

import "fmt"

/**CFile****************************************************************

  FileName    [wlcAbs2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Abstraction for word-level networks.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcAbs2.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "proof/pdr/pdr.h"
//#include "aig/gia/giaAig.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Mark operators that meet the abstraction criteria.]

	Description [This procedure returns the array of objects (vLeaves) that
	should be abstracted because of their high bit-width. It uses input array (vUnmark)
	to not abstract those objects that have been refined in the previous rounds.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbs2MarkOpers2(p *Wlc_Ntk_t, pPars *Wlc_Par_t, vUnmark *Vec_Bit_t, fVerbose int32) *Vec_Bit_t {
	var vLeaves *Vec_Bit_t = Vec_BitStart(Wlc_NtkObjNumMax(p))
	var pObj *Wlc_Obj_t
	var i int32
	var Count [4]int32
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if (vUnmark != nil) && Vec_BitEntry(vUnmark, i) { // not allow this object to be abstracted away
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsAdd {
				Vec_BitWriteEntry(vLeaves, Wlc_ObjId(p, pObj), 1)
				Count[0]++
			}
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE) || pObj.Type == uint32(WLC_OBJ_ARI_REM) || pObj.Type == uint32(WLC_OBJ_ARI_MODULUS) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsMul {
				Vec_BitWriteEntry(vLeaves, Wlc_ObjId(p, pObj), 1)
				Count[1]++
			}
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_MUX) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsMux {
				Vec_BitWriteEntry(vLeaves, Wlc_ObjId(p, pObj), 1)
				Count[2]++
			}
			continue
		}
		if Wlc_ObjIsCi(pObj) != 0 && Wlc_ObjIsPi(pObj) == 0 {
			if Wlc_ObjRange(pObj) >= pPars.nBitsFlop {
				Vec_BitWriteEntry(vLeaves, Wlc_ObjId(p, pObj), 1)
				Count[3]++
			}
			continue
		}
	}
	if fVerbose != 0 {
		fmt.Printf("Abstraction engine marked %d adds/subs, %d muls/divs, %d muxes, and %d flops to be abstracted away.\n", Count[0], Count[1], Count[2], Count[3])
	}
	return vLeaves
}

/*
*Function*************************************************************

	Synopsis    [Marks nodes to be included in the abstracted network.]

	Description [Marks all objects that will be included in the abstracted model.
	Stops at the objects (vLeaves) that are abstracted away. Returns three arrays:
	a subset of original PIs (vPisOld), a subset of pseudo-PIs (vPisNew) and the
	set of flops present as flops in the abstracted network.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbs2MarkNodes2_rec2(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vLeaves *Vec_Bit_t, vPisOld *Vec_Int_t, vPisNew *Vec_Int_t, vFlops *Vec_Int_t) {
	var i, iFanin int32
	if pObj.Mark != 0 {
		return
	}
	pObj.Mark = 1
	if Vec_BitEntry(vLeaves, Wlc_ObjId(p, pObj)) {
		assert(Wlc_ObjIsPi(pObj) == 0, "Wlc_ObjIsPi(pObj) == 0")
		Vec_IntPush(vPisNew, Wlc_ObjId(p, pObj))
		return
	}
	if Wlc_ObjIsCi(pObj) != 0 {
		if Wlc_ObjIsPi(pObj) != 0 {
			Vec_IntPush(vPisOld, Wlc_ObjId(p, pObj))
		} else {
			Vec_IntPush(vFlops, Wlc_ObjId(p, pObj))
		}
		return
	}
	// inlining of macro Wlc_ObjForEachFanin
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		iFanin = Wlc_ObjFaninId(pObj, i)
		Wlc_NtkAbs2MarkNodes2_rec2(p, Wlc_NtkObj(p, iFanin), vLeaves, vPisOld, vPisNew, vFlops)
	}
}
func Wlc_NtkAbs2MarkNodes2(p *Wlc_Ntk_t, vLeaves *Vec_Bit_t, vPisOld *Vec_Int_t, vPisNew *Vec_Int_t, vFlops *Vec_Int_t) {
	var pObj *Wlc_Obj_t
	var i, Count int32
	Wlc_NtkCleanMarks(p)
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		Wlc_NtkAbs2MarkNodes2_rec2(p, pObj, vLeaves, vPisOld, vPisNew, vFlops)
	}
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vFlops); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vFlops, i))
		Wlc_NtkAbs2MarkNodes2_rec2(p, Wlc_ObjFo2Fi(p, pObj), vLeaves, vPisOld, vPisNew, vFlops)
	}
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		Count += int32(pObj.Mark)
	}
	//    printf( "Collected %d old PIs, %d new PIs, %d flops, and %d other objects.\n",
	//        Vec_IntSize(vPisOld), Vec_IntSize(vPisNew), Vec_IntSize(vFlops),
	//        Count - Vec_IntSize(vPisOld) - Vec_IntSize(vPisNew) - Vec_IntSize(vFlops) );
	Vec_IntSort(vPisOld, 0)
	Vec_IntSort(vPisNew, 0)
	Vec_IntSort(vFlops, 0)
	Wlc_NtkCleanMarks(p)
}

/*
*Function*************************************************************

	Synopsis    [Derive word-level abstracted model based on the parameter values.]

	Description [Retuns the word-level abstracted network and the set of pseudo-PIs
	(vPisNew), which were created during abstraction. If the abstraction is
	satisfiable, some of the pseudo-PIs will be un-abstracted. These pseudo-PIs
	and their MFFC cones will be listed in the array (vUnmark), which will
	force the abstraction to not stop at these pseudo-PIs in the future.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbsAbstract2(p *Wlc_Ntk_t, pPars *Wlc_Par_t, vUnmark *Vec_Bit_t, pvPisNew **Vec_Int_t, fVerbose int32) *Wlc_Ntk_t {
	var pNtkNew *Wlc_Ntk_t = nil
	var vPisOld *Vec_Int_t = Vec_IntAlloc(100)
	var vPisNew *Vec_Int_t = Vec_IntAlloc(100)
	var vFlops *Vec_Int_t = Vec_IntAlloc(100)
	var vLeaves *Vec_Bit_t = Wlc_NtkAbs2MarkOpers2(p, pPars, vUnmark, fVerbose)
	Wlc_NtkAbs2MarkNodes2(p, vLeaves, vPisOld, vPisNew, vFlops)
	Vec_BitFree(vLeaves)
	pNtkNew = Wlc_NtkDupDfsAbs(p, vPisOld, vPisNew, vFlops)
	Vec_IntFree(vPisOld)
	Vec_IntFree(vFlops)
	if pvPisNew != nil {
		*pvPisNew = vPisNew
	} else {
		Vec_IntFree(vPisNew)
	}
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Find what objects need to be un-abstracted.]

	Description [Returns a subset of pseudo-PIs (vPisNew), which will be
	prevented from being abstracted in the future rounds of abstraction.
	The AIG manager (pGia) is a bit-level view of the abstracted model.
	The counter-example (pCex) is used to find waht PPIs to refine.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbs2Refinement(p *Wlc_Ntk_t, pGia *Gia_Man_t, pCex *Abc_Cex_t, vPisNew *Vec_Int_t) *Vec_Int_t {
	var vRefine *Vec_Int_t = Vec_IntAlloc(100)
	var pCexCare *Abc_Cex_t
	var pObj *Wlc_Obj_t
	// count the number of bit-level PPIs and map them into word-level objects they were derived from
	var f, i, b, nRealPis, nPpiBits int32
	var vMap *Vec_Int_t = Vec_IntStartFull(pCex.NPis)
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vPisNew); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vPisNew, i))
		for b = 0; b < Wlc_ObjRange(pObj); b++ {
			Vec_IntWriteEntry(vMap, nPpiBits, Wlc_ObjId(p, pObj))
			nPpiBits++
		}
	}
	// since PPIs are ordered last, the previous bits are real PIs
	nRealPis = pCex.NPis - nPpiBits
	// find the care-set
	pCexCare = Bmc_CexCareMinimizeAig(pGia, nRealPis, pCex, 1, 0, 0)
	assert(pCexCare.NPis == pCex.NPis, "pCexCare.NPis == pCex.NPis")
	// detect care PPIs
	for f = 0; f <= pCexCare.IFrame; f++ {
		for i = nRealPis; i < pCexCare.NPis; i++ {
			if Abc_InfoHasBit(pCexCare.PData, pCexCare.NRegs+pCexCare.NPis*f+i) {
				Vec_IntPushUniqueOrder(vRefine, Vec_IntEntry(vMap, i-nRealPis))
			}
		}
	}
	Abc_CexFree(pCexCare)
	Vec_IntFree(vMap)
	if Vec_IntSize(vRefine) == 0 { // real CEX
		Vec_IntFreeP(&vRefine)
	}
	return vRefine
}

/*
*Function*************************************************************

	Synopsis    [Mark MFFC cones of the un-abstracted objects.]

	Description [The MFFC cones of the objects in vRefine are traversed
	and all their nodes are marked in vUnmark.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkNodeDeref_rec2(p *Wlc_Ntk_t, pNode *Wlc_Obj_t, vUnmark *Vec_Bit_t) int32 {
	var i int32
	var Fanin int32
	var Counter int32 = 1
	if Wlc_ObjIsCi(pNode) != 0 {
		return 0
	}
	Vec_BitWriteEntry(vUnmark, Wlc_ObjId(p, pNode), 1)
	// inlining of macro Wlc_ObjForEachFanin
	for i = 0; i < Wlc_ObjFaninNum(pNode); i++ {
		Fanin = Wlc_ObjFaninId(pNode, i)
		Vec_IntAddToEntry(&p.vRefs, Fanin, -1)
		if Vec_IntEntry(&p.vRefs, Fanin) == 0 {
			Counter += Wlc_NtkNodeDeref_rec2(p, Wlc_NtkObj(p, Fanin), vUnmark)
		}
	}
	return Counter
}
func Wlc_NtkNodeRef_rec2(p *Wlc_Ntk_t, pNode *Wlc_Obj_t) int32 {
	var i int32
	var Fanin int32
	var Counter int32 = 1
	if Wlc_ObjIsCi(pNode) != 0 {
		return 0
	}
	// inlining of macro Wlc_ObjForEachFanin
	for i = 0; i < Wlc_ObjFaninNum(pNode); i++ {
		Fanin = Wlc_ObjFaninId(pNode, i)
		if Vec_IntEntry(&p.vRefs, Fanin) == 0 {
			Counter += Wlc_NtkNodeRef_rec2(p, Wlc_NtkObj(p, Fanin))
		}
		Vec_IntAddToEntry(&p.vRefs, Fanin, 1)
	}
	return Counter
}
func Wlc_NtkMarkMffc2(p *Wlc_Ntk_t, pNode *Wlc_Obj_t, vUnmark *Vec_Bit_t) int32 {
	var Count1, Count2 int32
	// if this is a flop output, compute MFFC of the corresponding flop input
	for Wlc_ObjIsCi(pNode) != 0 {
		Vec_BitWriteEntry(vUnmark, Wlc_ObjId(p, pNode), 1)
		pNode = Wlc_ObjFo2Fi(p, pNode)
	}
	assert(Wlc_ObjIsCi(pNode) == 0, "Wlc_ObjIsCi(pNode) == 0")
	// dereference the node (and set the bits in vUnmark)
	Count1 = Wlc_NtkNodeDeref_rec2(p, pNode, vUnmark)
	// reference it back
	Count2 = Wlc_NtkNodeRef_rec2(p, pNode)
	assert(Count1 == Count2, "Count1 == Count2")
	return Count1
}
func Wlc_NtkRemoveFromAbstraction2(p *Wlc_Ntk_t, vRefine *Vec_Int_t, vUnmark *Vec_Bit_t) int32 {
	var pObj *Wlc_Obj_t
	var i, nNodes int32
	if Vec_IntSize(&p.vRefs) == 0 {
		Wlc_NtkSetRefs(p)
	}
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vRefine); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vRefine, i))
		nNodes += Wlc_NtkMarkMffc2(p, pObj, vUnmark)
	}
	return nNodes
}

/*
*Function*************************************************************

	Synopsis    [Performs abstraction.]

	Description [Derives initial abstraction based on user-specified
	parameter values, which tell what is the smallest bit-width of a
	primitive that is being abstracted away.  Currently only add/sub,
	mul/div, mux, and flop are supported with individual parameters.
	The second step is to refine the initial abstraction until the
	point when the property is proved.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbs2Core2(p *Wlc_Ntk_t, pPars *Wlc_Par_t) int32 {
	var clk abctime = Abc_Clock()
	var nIters int32
	var nNodes int32
	var nDcFlops int32
	var RetValue int32 = -1
	// start the bitmap to mark objects that cannot be abstracted because of refinement
	// currently, this bitmap is empty because abstraction begins without refinement
	var vUnmark *Vec_Bit_t = Vec_BitStart(Wlc_NtkObjNumMax(p))
	// set up parameters to run PDR
	var PdrPars Pdr_Par_t
	var pPdrPars *Pdr_Par_t = &PdrPars
	Pdr_ManSetDefaultParams(pPdrPars)
	pPdrPars.SetUseAbs(1)   // use 'pdr -t'  (on-the-fly abstraction)
	pPdrPars.SetCtgs(1)     // use 'pdr -nc' (improved generalization)
	pPdrPars.SetSkipDown(0) // use 'pdr -nc' (improved generalization)
	//pPdrPars->nRestLimit = 500; // reset queue or proof-obligations when it gets larger than this
	pPdrPars.SetVerbose(pPars.fPdrVerbose)
	// perform refinement iterations
	for nIters = 1; nIters < pPars.nIterMax; nIters++ {
		var pAig *Aig_Man_t
		var pCex *Abc_Cex_t
		var vPisNew, vRefine *Vec_Int_t
		var pGia, pTemp *Gia_Man_t
		var pAbs *Wlc_Ntk_t

		if pPars.fVerbose != 0 {
			fmt.Printf("\nIteration %d:\n", nIters)
		}

		// get abstracted GIA and the set of pseudo-PIs (vPisNew)
		pAbs = Wlc_NtkAbsAbstract2(p, pPars, vUnmark, &vPisNew, pPars.fVerbose)
		pGia = Wlc_NtkBitBlast(pAbs, nil)

		// if the abstraction has flops with DC-init state,
		// new PIs were introduced by bit-blasting at the end of the PI list
		// here we move these variables to be *before* PPIs, because
		// PPIs are supposed to be at the end of the PI list for refinement
		nDcFlops = Wlc_NtkDcFlopNum(pAbs)
		if nDcFlops > 0 { // DC-init flops are present

			pTemp = pGia
			pGia = Gia_ManPermuteInputs(pTemp, Wlc_NtkCountObjBits(p, vPisNew), nDcFlops)
			Gia_ManStop(pTemp)
		}
		// if the word-level outputs have to be XORs, this is a place to do it
		if pPars.fXorOutput != 0 {
			pTemp = pGia
			pGia = Gia_ManTransformMiter2(pTemp)
			Gia_ManStop(pTemp)
		}
		if pPars.fVerbose != 0 {
			fmt.Printf("Derived abstraction with %d objects and %d PPIs. Bit-blasted AIG stats are:\n", Wlc_NtkObjNum(pAbs), Vec_IntSize(vPisNew))
			Gia_ManPrintStats(pGia, nil)
		}
		Wlc_NtkFree(pAbs)

		// try to prove abstracted GIA by converting it to AIG and calling PDR
		pAig = Gia_ManToAigSimple(pGia)
		RetValue = Pdr_ManSolve(pAig, pPdrPars)
		pCex = pAig.TakeSequentialCounterexample()
		Aig_ManStop(pAig)

		// consider outcomes
		if pCex == nil {
			assert((RetValue != 0), "RetValue") // proved or undecided
			Gia_ManStop(pGia)
			Vec_IntFree(vPisNew)
			break
		}

		// perform refinement
		vRefine = Wlc_NtkAbs2Refinement(p, pGia, pCex, vPisNew)
		Gia_ManStop(pGia)
		Vec_IntFree(vPisNew)
		if vRefine == nil { // real CEX

			Abc_CexFree(pCex) // return CEX in the future
			break
		}

		// update the set of objects to be un-abstracted
		nNodes = Wlc_NtkRemoveFromAbstraction2(p, vRefine, vUnmark)
		if pPars.fVerbose != 0 {
			fmt.Printf("Refinement of CEX in frame %d came up with %d un-abstacted PPIs, whose MFFCs include %d objects.\n", pCex.IFrame, Vec_IntSize(vRefine), nNodes)
		}
		Vec_IntFree(vRefine)
		Abc_CexFree(pCex)
	}
	Vec_BitFree(vUnmark)
	// report the result
	if pPars.fVerbose != 0 {
		fmt.Printf("\n")
	}
	fmt.Printf("Abstraction ")
	if RetValue == 0 {
		fmt.Printf("resulted in a real CEX")
	} else if RetValue == 1 {
		fmt.Printf("is successfully proved")
	} else {
		fmt.Printf("timed out")
	}
	fmt.Printf(" after %d iterations. ", nIters)
	Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	return RetValue
}

// Wlc_NtkAbsCore2 retains the public C entry-point spelling used by the
// command layer. The implementation is the second abstraction core.
func Wlc_NtkAbsCore2(p *Wlc_Ntk_t, pPars *Wlc_Par_t) int32 {
	return Wlc_NtkAbs2Core2(p, pPars)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
