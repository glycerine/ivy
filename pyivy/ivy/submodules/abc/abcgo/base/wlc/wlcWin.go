package wlc

import "fmt"

// wlcBoolInt preserves the C helpers' 0/1 return convention.
func wlcBoolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

/**CFile****************************************************************

  FileName    [wlcWin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Parses several flavors of word-level Verilog.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcWin.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "base/abc/abc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Collect arithmetic nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_ObjIsArithm(pObj *Wlc_Obj_t) int32 {
	return wlcBoolInt(pObj.Type == uint32(WLC_OBJ_CONST) ||
		pObj.Type == uint32(WLC_OBJ_BUF) || pObj.Type == uint32(WLC_OBJ_BIT_NOT) ||
		pObj.Type == uint32(WLC_OBJ_BIT_ZEROPAD) || pObj.Type == uint32(WLC_OBJ_BIT_SIGNEXT) ||
		//           pObj->Type == WLC_OBJ_BIT_SELECT  || pObj->Type == WLC_OBJ_BIT_CONCAT  ||
		pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) ||
		pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS))
}
func Wlc_ObjIsArithmReal(pObj *Wlc_Obj_t) int32 {
	return wlcBoolInt(pObj.Type == uint32(WLC_OBJ_BIT_NOT) ||
		pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) ||
		pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS))
}
func Wlc_ManCountArithmReal(p *Wlc_Ntk_t, vNodes *Vec_Int_t) int32 {
	var pObj *Wlc_Obj_t
	var i, Counter int32
	// inlining of macro Wlc_NtkForEachObjVec
	for i = 0; i < Vec_IntSize(vNodes); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vNodes, i))
		Counter += Wlc_ObjIsArithmReal(pObj)
	}
	return Counter
}
func Wlc_ObjHasArithm_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	if pObj.Type == uint32(WLC_OBJ_CONST) {
		return 0
	}
	if pObj.Type == uint32(WLC_OBJ_BUF) || pObj.Type == uint32(WLC_OBJ_BIT_NOT) ||
		pObj.Type == uint32(WLC_OBJ_BIT_ZEROPAD) || pObj.Type == uint32(WLC_OBJ_BIT_SIGNEXT) {
		return Wlc_ObjHasArithm_rec(p, Wlc_ObjFanin0(p, pObj))
	}
	return wlcBoolInt(pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) ||
		pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS))
}
func Wlc_ObjHasArithmFanins(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	var pFanin *Wlc_Obj_t
	var i int32
	assert(!(Wlc_ObjHasArithm_rec(p, pObj) != 0), "!Wlc_ObjHasArithm_rec(p, pObj)")
	// inlining of macro Wlc_ObjForEachFaninObj
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		pFanin = Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, i))
		if Wlc_ObjHasArithm_rec(p, pFanin) != 0 {
			return 1
		}
	}
	return 0
}
func Wlc_WinCompute_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t) {
	var pFanin *Wlc_Obj_t
	var i int32
	if pObj.Mark != 0 {
		return
	}
	pObj.Mark = 1
	if !(Wlc_ObjIsArithm(pObj) != 0) {
		Vec_IntPush(vLeaves, Wlc_ObjId(p, pObj))
		return
	}
	// inlining of macro Wlc_ObjForEachFaninObj
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		pFanin = Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, i))
		Wlc_WinCompute_rec(p, pFanin, vLeaves, vNodes)
	}
	Vec_IntPush(vNodes, Wlc_ObjId(p, pObj))
}
func Wlc_WinCleanMark_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) {
	var pFanin *Wlc_Obj_t
	var i int32
	if pObj.Mark == 0 {
		return
	}
	pObj.Mark = 0
	// inlining of macro Wlc_ObjForEachFaninObj
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		pFanin = Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, i))
		Wlc_WinCleanMark_rec(p, pFanin)
	}
}
func Wlc_WinCompute(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t) {
	Vec_IntClear(vLeaves)
	Vec_IntClear(vNodes)
	if Wlc_ObjHasArithm_rec(p, pObj) != 0 {
		Wlc_WinCompute_rec(p, pObj, vLeaves, vNodes)
		Wlc_WinCleanMark_rec(p, pObj)

	} else if Wlc_ObjHasArithmFanins(p, pObj) != 0 {
		var pFanin *Wlc_Obj_t
		var i int32
		// inlining of macro Wlc_ObjForEachFaninObj
		for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
			pFanin = Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, i))
			if Wlc_ObjHasArithm_rec(p, pFanin) != 0 {
				Wlc_WinCompute_rec(p, pFanin, vLeaves, vNodes)
			}
		}
		// inlining of macro Wlc_ObjForEachFaninObj
		for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
			pFanin = Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, i))
			if Wlc_ObjHasArithm_rec(p, pFanin) != 0 {
				Wlc_WinCleanMark_rec(p, pFanin)
			}
		}

	} else {
		assert(false, "0")
	}
}
func Wlc_WinProfileArith(p *Wlc_Ntk_t) {
	var vLeaves *Vec_Int_t = Vec_IntAlloc(1000)
	var vNodes *Vec_Int_t = Vec_IntAlloc(1000)
	var pObj *Wlc_Obj_t
	var i, Count int32
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		pObj.Mark = 0
	}
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if (Wlc_ObjHasArithm_rec(p, pObj) != 0 && Wlc_ObjIsCo(pObj) != 0) ||
			(Wlc_ObjHasArithm_rec(p, pObj) == 0 && Wlc_ObjHasArithmFanins(p, pObj) != 0) {
			Wlc_WinCompute(p, pObj, vLeaves, vNodes)
			if Wlc_ManCountArithmReal(p, vNodes) < 2 {
				continue
			}

			fmt.Printf("Arithmetic cone of node %d (%s):\n", Wlc_ObjId(p, pObj), Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
			Wlc_NtkPrintNode(p, pObj)
			Vec_IntReverseOrder(vNodes)
			Wlc_NtkPrintNodeArray(p, vNodes)
			fmt.Printf("\n")
			Count++
		}
	}
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		assert(pObj.Mark == 0, "pObj.Mark == 0")
	}
	fmt.Printf("Finished printing %d arithmetic cones.\n", Count)
	Vec_IntFree(vLeaves)
	Vec_IntFree(vNodes)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
