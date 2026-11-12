package if_

import "fmt"

/**CFile****************************************************************

  FileName    [ifCheck.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Sequential mapping.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifCheck.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// elementary truth tables
var Truths6 [6]word = [6]word{
	0xAAAAAAAAAAAAAAAA,
	0xCCCCCCCCCCCCCCCC,
	0xF0F0F0F0F0F0F0F0,
	0xFF00FF00FF00FF00,
	0xFFFF0000FFFF0000,
	0xFFFFFFFF00000000,
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the node Leaf is reachable on the path.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManCutReach_rec(pPath *If_Obj_t, pLeaf *If_Obj_t) int32 {
	if pPath == pLeaf {
		return 1
	}
	if pPath.fMark != 0 {
		return 0
	}
	assert(If_ObjIsAnd(pPath), "If_ObjIsAnd(pPath)")
	if If_ManCutReach_rec(If_ObjFanin0(pPath), pLeaf) != 0 {
		return 1
	}
	if If_ManCutReach_rec(If_ObjFanin1(pPath), pLeaf) != 0 {
		return 1
	}
	return 0
}
func If_ManCutReach(p *If_Man_t, pCut *If_Cut_t, pPath *If_Obj_t, pLeaf *If_Obj_t) int32 {
	var pTemp *If_Obj_t
	RetValue := int32(0)
	for _, leafID := range pCut.pLeaves {
		pTemp = If_ManObj(p, leafID)
		pTemp.fMark = 1
	}
	RetValue = If_ManCutReach_rec(pPath, pLeaf)
	for _, leafID := range pCut.pLeaves {
		pTemp = If_ManObj(p, leafID)
		pTemp.fMark = 0
	}
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Derive truth table for each cofactor.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManCutTruthCheck_rec(pObj *If_Obj_t, pTruths []word) word {
	var T0, T1 word
	if pObj.fMark != 0 {
		return pTruths[If_ObjId(pObj)]
	}
	assert(If_ObjIsAnd(pObj), "If_ObjIsAnd(pObj)")
	T0 = If_ManCutTruthCheck_rec(If_ObjFanin0(pObj), pTruths)
	T1 = If_ManCutTruthCheck_rec(If_ObjFanin1(pObj), pTruths)
	if If_ObjFaninC0(pObj) {
		T0 = ~T0
	} else {
		T0 = T0
	}
	if If_ObjFaninC1(pObj) {
		T1 = ~T1
	} else {
		T1 = T1
	}
	return T0 & T1
}
func If_ManCutTruthCheck(p *If_Man_t, pObj *If_Obj_t, pCut *If_Cut_t, pLeaf *If_Obj_t, Cof int32, pTruths []word) int32 {
	var Truth word
	var pTemp *If_Obj_t
	var i, k int32
	assert(Cof == 0 || Cof == 1, "Cof == 0 || Cof == 1")
	// inlining of macro If_CutForEachLeaf
	for _, leafID := range pCut.pLeaves {
		pTemp = If_ManObj(p, leafID)
		assert(pTemp.fMark == 0, "pTemp.fMark == 0")
		pTemp.fMark = 1
		if pLeaf == pTemp {
			if Cof != 0 {
				pTruths[If_ObjId(pTemp)] = ^word(0)
			} else {
				pTruths[If_ObjId(pTemp)] = 0
			}
		} else {
			pTruths[If_ObjId(pTemp)] = Truths6[k]
			k++
		}
	}
	assert(k+1 == If_CutLeaveNum(pCut), "k + 1 == If_CutLeaveNum(pCut)")
	// compute truth table
	Truth = If_ManCutTruthCheck_rec(pObj, pTruths)
	// inlining of macro If_CutForEachLeaf
	for _, leafID := range pCut.pLeaves {
		If_ManObj(p, leafID).fMark = 0
	}
	if Truth == 0 || Truth == ^word(0) {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Checks if cut can be structurally/functionally decomposed.]

	Description [The decomposition is Fn(a,b,c,...) = F2(a, Fn-1(b,c,...)).]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_If_ManCutCheck_nDecCalls int32
var spStatic_If_ManCutCheck_nDecStruct int32
var spStatic_If_ManCutCheck_nDecStruct2 int32
var spStatic_If_ManCutCheck_nDecFunction int32

func If_ManCutCheck(p *If_Man_t, pObj *If_Obj_t, pCut *If_Cut_t) {

	var pLeaf, pPath *If_Obj_t
	if pCut == nil {
		fmt.Printf("DecStr  = %9d (%6.2f %%).\n", spStatic_If_ManCutCheck_nDecStruct, 100.0*spStatic_If_ManCutCheck_nDecStruct/spStatic_If_ManCutCheck_nDecCalls)
		fmt.Printf("DecStr2 = %9d (%6.2f %%).\n", spStatic_If_ManCutCheck_nDecStruct2, 100.0*spStatic_If_ManCutCheck_nDecStruct2/spStatic_If_ManCutCheck_nDecCalls)
		fmt.Printf("DecFunc = %9d (%6.2f %%).\n", spStatic_If_ManCutCheck_nDecFunction, 100.0*spStatic_If_ManCutCheck_nDecFunction/spStatic_If_ManCutCheck_nDecCalls)
		fmt.Printf("Total   = %9d (%6.2f %%).\n", spStatic_If_ManCutCheck_nDecCalls, 100.0*spStatic_If_ManCutCheck_nDecCalls/spStatic_If_ManCutCheck_nDecCalls)
		return
	}
	assert(If_ObjIsAnd(pObj), "If_ObjIsAnd(pObj)")
	assert(pCut.nLeaves <= 7, "pCut.nLeaves <= 7")
	spStatic_If_ManCutCheck_nDecCalls++
	// check structural decomposition
	// inlining of macro If_CutForEachLeaf
	found := false
	for _, leafID := range pCut.pLeaves {
		pLeaf = If_ManObj(p, leafID)
		if pLeaf == If_ObjFanin0(pObj) || pLeaf == If_ObjFanin1(pObj) {
			found = true
			break
		}
	}
	if found {
		if pLeaf == If_ObjFanin0(pObj) {
			pPath = If_ObjFanin1(pObj)
		} else {
			pPath = If_ObjFanin0(pObj)
		}
		if !(If_ManCutReach(p, pCut, pPath, pLeaf) != 0) {
			spStatic_If_ManCutCheck_nDecStruct++
			//            nDecFunction++;
			//            return;

		} else {
			spStatic_If_ManCutCheck_nDecStruct2++
		}
	}
	// check functional decomposition
	pTruths := make([]word, Vec_PtrSize(p.vObjs))
	for _, leafID := range pCut.pLeaves {
		pLeaf = If_ManObj(p, leafID)
		if If_ManCutTruthCheck(p, pObj, pCut, pLeaf, 0, pTruths) != 0 {
			spStatic_If_ManCutCheck_nDecFunction++
			break
		}
		if If_ManCutTruthCheck(p, pObj, pCut, pLeaf, 1, pTruths) != 0 {
			spStatic_If_ManCutCheck_nDecFunction++
			break
		}
	}

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
