package wlc

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [wlcMem.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Memory abstraction.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcMem.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "sat/bsat/satStore.h"
//#include "proof/pdr/pdr.h"
//#include "proof/pdr/pdrInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Memory blasting.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkMemBlast_rec(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, iObj int32, vFanins *Vec_Int_t) {
	var pObj *Wlc_Obj_t
	var i, iFanin int32
	if Wlc_ObjCopy(p, iObj) {
		return
	}
	pObj = Wlc_NtkObj(p, iObj)
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		iFanin = Wlc_ObjFaninId(pObj, i)
		Wlc_NtkMemBlast_rec(pNew, p, iFanin, vFanins)
	}
	if pObj.Type == WLC_OBJ_WRITE {
		var vTemp *Vec_Int_t = Vec_IntAlloc(1)
		var vBits *Vec_Int_t = Vec_IntAlloc(100)
		var pMem *Wlc_Obj_t = Wlc_ObjFanin0(p, pObj)
		var pAddr *Wlc_Obj_t = Wlc_ObjFanin1(p, pObj)
		var pData *Wlc_Obj_t = Wlc_ObjFanin2(p, pObj)
		var DataW int32 = Wlc_ObjRange(pData)
		var AddrW int32 = Wlc_ObjRange(pAddr)
		var nRegs int32 = 1 << AddrW
		var iObjNew int32
		var iObjDec int32
		assert(nRegs*DataW == Wlc_ObjRange(pMem), "nRegs * DataW == Wlc_ObjRange(pMem)")
		assert(Wlc_ObjRange(pObj) == Wlc_ObjRange(pMem), "Wlc_ObjRange(pObj) == Wlc_ObjRange(pMem)")
		// create decoder
		iObjDec = Wlc_ObjAlloc(pNew, WLC_OBJ_DEC, 0, nRegs-1, 0)
		Vec_IntFill(vTemp, 1, Wlc_ObjCopy(p, Wlc_ObjId(p, pAddr)))
		Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObjDec), vTemp)
		// create decoder bits
		for i = 0; i < nRegs; i++ {
			var iObj2 int32 = Wlc_ObjAlloc(pNew, WLC_OBJ_BIT_SELECT, 0, i, i)
			Vec_IntFill(vTemp, 1, iObjDec)
			Vec_IntPushTwo(vTemp, i, i)
			Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObj2), vTemp)
			Vec_IntPush(vBits, iObj2)
		}
		// create data words
		Vec_IntClear(vFanins)
		for i = 0; i < nRegs; i++ {
			var iObj2 int32 = Wlc_ObjAlloc(pNew, WLC_OBJ_BIT_SELECT, 0, i*DataW+DataW-1, i*DataW)
			Vec_IntFill(vTemp, 1, Wlc_ObjCopy(p, Wlc_ObjId(p, pMem)))
			Vec_IntPushTwo(vTemp, i*DataW+DataW-1, i*DataW)
			Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObj2), vTemp)
			Vec_IntPush(vFanins, iObj2)
		}
		// create MUXes of data words controlled by decoder bits
		for i = 0; i < nRegs; i++ {
			var iObj2 int32 = Wlc_ObjAlloc(pNew, WLC_OBJ_MUX, 0, DataW-1, 0)
			Vec_IntFill(vTemp, 1, Vec_IntEntry(vBits, i))
			Vec_IntPushTwo(vTemp, Wlc_ObjCopy(p, Wlc_ObjId(p, pData)), Vec_IntEntry(vFanins, i))
			Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObj2), vTemp)
			Vec_IntWriteEntry(vFanins, i, iObj2)
		}
		// concatenate the results
		iObjNew = Wlc_ObjAlloc(pNew, WLC_OBJ_BIT_CONCAT, 0, nRegs*DataW-1, 0)
		Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObjNew), vFanins)
		Wlc_ObjSetCopy(p, iObj, iObjNew)
		Vec_IntFree(vTemp)
		Vec_IntFree(vBits)

	} else if pObj.Type == WLC_OBJ_READ {
		var vTemp *Vec_Int_t = Vec_IntAlloc(1)
		var pMem *Wlc_Obj_t = Wlc_ObjFanin0(p, pObj)
		var pAddr *Wlc_Obj_t = Wlc_ObjFanin1(p, pObj)
		var DataW int32 = Wlc_ObjRange(pObj)
		var AddrW int32 = Wlc_ObjRange(pAddr)
		var nRegs int32 = 1 << AddrW
		var iObjNew int32
		assert(nRegs*DataW == Wlc_ObjRange(pMem), "nRegs * DataW == Wlc_ObjRange(pMem)")
		Vec_IntClear(vFanins)
		Vec_IntPush(vFanins, Wlc_ObjCopy(p, Wlc_ObjId(p, pAddr)))
		for i = 0; i < nRegs; i++ {
			var iObj2 int32 = Wlc_ObjAlloc(pNew, WLC_OBJ_BIT_SELECT, 0, i*DataW+DataW-1, i*DataW)
			Vec_IntFill(vTemp, 1, Wlc_ObjCopy(p, Wlc_ObjId(p, pMem)))
			Vec_IntPushTwo(vTemp, i*DataW+DataW-1, i*DataW)
			Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObj2), vTemp)
			Vec_IntPush(vFanins, iObj2)
		}
		iObjNew = Wlc_ObjAlloc(pNew, WLC_OBJ_MUX, 0, DataW-1, 0)
		Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObjNew), vFanins)
		Wlc_ObjSetCopy(p, iObj, iObjNew)
		Vec_IntFree(vTemp)

	} else {
		Wlc_ObjDup(pNew, p, iObj, vFanins)
	}
}
func Wlc_NtkMemBlast(p *Wlc_Ntk_t) *Wlc_Ntk_t {
	var pNew *Wlc_Ntk_t
	var pObj *Wlc_Obj_t
	var vFanins *Vec_Int_t
	var i int32
	Wlc_NtkCleanCopy(p)
	vFanins = Vec_IntAlloc(100)
	pNew = Wlc_NtkAlloc(p.pName, p.nObjsAlloc)
	pNew.fSmtLib = p.fSmtLib
	pNew.fAsyncRst = p.fAsyncRst
	pNew.fMemPorts = p.fMemPorts
	pNew.fEasyFfs = p.fEasyFfs
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		Wlc_NtkMemBlast_rec(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), pObj.fIsFi)
	}
	if p.vInits {
		pNew.vInits = Vec_IntDup(p.vInits)
	}
	if p.pInits {
		pNew.pInits = Abc_UtilStrsav(p.pInits)
	}
	Vec_IntFree(vFanins)
	if p.pSpec {
		pNew.pSpec = Abc_UtilStrsav(p.pSpec)
	}
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Collect memory nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkCollectMemSizes(p *Wlc_Ntk_t) *Vec_Int_t {
	var pObj *Wlc_Obj_t
	var i int32
	var vMemSizes *Vec_Int_t = Vec_IntAlloc(10)
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjType(pObj) != WLC_OBJ_WRITE && Wlc_ObjType(pObj) != WLC_OBJ_READ {
			continue
		}
		pObj = Wlc_ObjFanin(p, pObj, 0)
		Vec_IntPushUnique(vMemSizes, Wlc_ObjRange(pObj))
	}
	return vMemSizes
}

// remove PIs without fanout, non-driven object and their fanouts
func Wlc_ObjCheckIsEmpty_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	var k, iFanin int32
	if Wlc_ObjType(pObj) == 0 {
		return 1
	}
	if Wlc_ObjIsPi(pObj) {
		return Vec_IntEntry(&p.vRefs, Wlc_ObjId(p, pObj)) == 0
	} else if Wlc_ObjIsCi(pObj) {
		return 0
	}
	for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
		iFanin = Wlc_ObjFaninId(pObj, k)
		if !Wlc_ObjCheckIsEmpty_rec(p, Wlc_NtkObj(p, iFanin)) {
			return 0
		}
	}
	return 1
}
func Wlc_NtkCleanObjects(p *Wlc_Ntk_t, vObjs *Vec_Int_t) *Vec_Int_t {
	var pObj *Wlc_Obj_t
	var i int32
	var vMemObjs *Vec_Int_t = Vec_IntAlloc(10)
	Wlc_NtkSetRefs(p)
	for i = 0; i < Vec_IntSize(vObjs); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vObjs, i))
		//printf( "Considering %d (%s)\n", Wlc_ObjId(p, pObj), Wlc_ObjName(p, Wlc_ObjId(p, pObj)) );
		if !(Wlc_ObjCheckIsEmpty_rec(p, pObj) != 0) {
			Vec_IntPush(vMemObjs, Wlc_ObjId(p, pObj))
		}
	}
	return vMemObjs
}
func Wlc_NtkCollectMemory(p *Wlc_Ntk_t, fClean int32) *Vec_Int_t {
	var pObj *Wlc_Obj_t
	var i int32
	var vMemSizes *Vec_Int_t = Wlc_NtkCollectMemSizes(p)
	var vMemObjs *Vec_Int_t = Vec_IntAlloc(10)
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjType(pObj) == WLC_OBJ_WRITE || Wlc_ObjType(pObj) == WLC_OBJ_READ {
			Vec_IntPush(vMemObjs, i)
		} else if Vec_IntFind(vMemSizes, Wlc_ObjRange(pObj)) >= 0 {
			Vec_IntPush(vMemObjs, i)
		}
	}
	Vec_IntFree(vMemSizes)
	Vec_IntSort(vMemObjs, 0)
	//Wlc_NtkPrintNodeArray( p, vMemObjs );
	if fClean != 0 {
		var vTemp *Vec_Int_t
		vTemp = vMemObjs
		vMemObjs = Wlc_NtkCleanObjects(p, vTemp)
		Vec_IntFree(vTemp)
	}
	return vMemObjs
}
func Wlc_NtkPrintMemory(p *Wlc_Ntk_t) {
	var vMemory *Vec_Int_t
	vMemory = Wlc_NtkCollectMemory(p, 1)
	fmt.Printf("Memory subsystem is composed of the following objects:\n")
	Wlc_NtkPrintNodeArray(p, vMemory)
	Vec_IntFree(vMemory)
}

/*
*Function*************************************************************

	Synopsis    [Collect fanins of memory subsystem.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkCollectMemFanins(p *Wlc_Ntk_t, vMemObjs *Vec_Int_t) *Vec_Int_t {
	var pObj *Wlc_Obj_t
	var i, k, iFanin int32
	var vMemFanins *Vec_Int_t = Vec_IntAlloc(100)
	for i = 0; i < Vec_IntSize(vMemObjs); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjs, i))
		if Wlc_ObjType(pObj) == WLC_OBJ_MUX {
			Vec_IntPush(vMemFanins, Wlc_ObjFaninId0(pObj))
		} else if Wlc_ObjType(pObj) == WLC_OBJ_READ || Wlc_ObjType(pObj) == WLC_OBJ_WRITE {
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				if k > 0 {
					Vec_IntPush(vMemFanins, iFanin)
				}
			}
		}
	}
	return vMemFanins
}

/*
*Function*************************************************************

	Synopsis    [Abstract memory nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_CountDcs(pInit string) int32 {
	var Count int32
	for _, c := range pInit {
		if c == 'x' || c == 'X' {
			Count++
		}
	}
	return Count
}
func Wlc_NtkDupOneObject(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, pObj *Wlc_Obj_t, TypeNew int32, vFanins *Vec_Int_t) int32 {
	var iObj int32 = Wlc_ObjId(p, pObj)
	var Type uint32 = pObj.Type
	var iObjNew int32
	var nFanins int32 = Wlc_ObjFaninNum(pObj)
	var iObjCopy int32 = Wlc_ObjCopy(p, iObj)
	pObj.Type = TypeNew
	pObj.nFanins = 0
	iObjNew = Wlc_ObjDup(pNew, p, iObj, vFanins)
	pObj.Type = Type
	pObj.nFanins = uint32(nFanins)
	if TypeNew == WLC_OBJ_FO {
		Vec_IntPush(pNew.vInits, -Wlc_ObjRange(pObj))
		Wlc_ObjSetCopy(p, iObj, iObjCopy)
	}
	return iObjNew
}
func Wlc_NtkDupOneBuffer(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, pObj *Wlc_Obj_t, iFanin int32, vFanins *Vec_Int_t, fIsFi int32) {
	var iObj int32 = Wlc_ObjAlloc(pNew, WLC_OBJ_BUF, pObj.Signed, pObj.End, pObj.Beg)
	var pObjNew *Wlc_Obj_t = Wlc_NtkObj(pNew, iObj)
	Vec_IntFill(vFanins, 1, iFanin)
	Wlc_ObjAddFanins(pNew, pObjNew, vFanins)
	Wlc_ObjSetCo(pNew, pObjNew, fIsFi)
}

func Wlc_NtkAbsAddToNodeFrames(vNodeFrames *Vec_Int_t, vLevel *Vec_Int_t) {
	var i, Entry int32
	Vec_IntForEachEntry(vLevel, Entry, i)
	Vec_IntPushUnique(vNodeFrames, Entry)
	Vec_IntSort(vNodeFrames, 0)
}
func Wlc_NtkAbsCreateFlopOutputs(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, vNodeFrames *Vec_Int_t, vFanins *Vec_Int_t) *Vec_Int_t {
	var vNewObjs *Vec_Int_t = Vec_IntAlloc(2 * Vec_IntSize(vNodeFrames))
	var pObj *Wlc_Obj_t
	var pFanin *Wlc_Obj_t = nil
	var i, Entry, iObj, iFrame int32
	for i = 0; i < Vec_IntSize(vNodeFrames); i++ {
		Entry = Vec_IntEntry(vNodeFrames, i)
		iObj = Entry >> 11
		iFrame = (Entry >> 1) & 0x3FF
		pObj = Wlc_NtkObj(p, iObj)
		// address
		if Wlc_ObjType(pObj) == WLC_OBJ_MUX {
			pFanin = Wlc_ObjFanin0(p, pObj)
		} else if Wlc_ObjType(pObj) == WLC_OBJ_READ || Wlc_ObjType(pObj) == WLC_OBJ_WRITE {
			pFanin = Wlc_ObjFanin1(p, pObj)
		} else {
			assert(false, "0")
		}
		Vec_IntPush(vNewObjs, Wlc_NtkDupOneObject(pNew, p, pFanin, WLC_OBJ_FO, vFanins))
		// data
		if Wlc_ObjType(pObj) == WLC_OBJ_MUX {
			pFanin = nil
		} else if Wlc_ObjType(pObj) == WLC_OBJ_READ {
			pFanin = pObj
		} else if Wlc_ObjType(pObj) == WLC_OBJ_WRITE {
			pFanin = Wlc_ObjFanin2(p, pObj)
		} else {
			assert(false, "0")
		}
		if pFanin != nil {
			Vec_IntPush(vNewObjs, Wlc_NtkDupOneObject(pNew, p, pFanin, WLC_OBJ_FO, vFanins))
		} else {
			Vec_IntPush(vNewObjs, 0)
		}
	}
	assert(Vec_IntSize(vNewObjs) == 2*Vec_IntSize(vNodeFrames), "Vec_IntSize(vNewObjs) == 2 * Vec_IntSize(vNodeFrames)")
	return vNewObjs
}
func Wlc_NtkAbsCreateFlopInputs(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, vNodeFrames *Vec_Int_t, vFanins *Vec_Int_t, vNewObjs *Vec_Int_t, pCounter *Wlc_Obj_t, AdderBits int32) {
	var pObj, pFanin, pFlop, pCond, pMux, pConst *Wlc_Obj_t
	var i, n, Entry, Value, iObj, iFrame int32
	for i = 0; i < Vec_IntSize(vNodeFrames); i++ {
		Entry = Vec_IntEntry(vNodeFrames, i)
		Value = Entry & 1
		iObj = Entry >> 11
		iFrame = (Entry >> 1) & 0x3FF
		pObj = Wlc_NtkObj(p, iObj)
		for n = 0; n < 2; n++ { // n=0 -- address   n=1 -- data

			pFlop = Wlc_NtkObj(pNew, Vec_IntEntry(vNewObjs, 2*i+n))
			if Wlc_ObjType(pObj) == WLC_OBJ_WRITE {
				if n != 0 {
					pFanin = Wlc_ObjCopyObj(pNew, p, Wlc_ObjFanin2(p, pObj))
				} else {
					pFanin = Wlc_ObjCopyObj(pNew, p, Wlc_ObjFanin1(p, pObj))
				}
			} else if Wlc_ObjType(pObj) == WLC_OBJ_READ {
				if n {
					pFanin = Wlc_NtkObj(pNew, Wlc_ObjCopy(p, iObj))
				} else {
					pFanin = Wlc_ObjCopyObj(pNew, p, Wlc_ObjFanin1(p, pObj))
				}
			} else if Wlc_ObjType(pObj) == WLC_OBJ_MUX {
				if n {
					continue
				}
				pFanin = Wlc_ObjCopyObj(pNew, p, Wlc_ObjFanin0(p, pObj))
				if Value != 0 {
					Vec_IntFill(vFanins, 1, Wlc_ObjId(pNew, pFanin))
					pFanin = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_BIT_NOT, 0, 0, 0))
					Wlc_ObjAddFanins(pNew, pFanin, vFanins)
				}

			} else {
				assert(false, "0")
			}
			assert(Wlc_ObjRange(pFlop) == Wlc_ObjRange(pFanin), "Wlc_ObjRange(pFlop) == Wlc_ObjRange(pFanin)")
			// create constant
			pConst = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_CONST, 0, AdderBits-1, 0))
			Vec_IntFill(vFanins, 1, iFrame)
			Wlc_ObjAddFanins(pNew, pConst, vFanins)
			// create equality
			pCond = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_COMP_EQU, 0, 0, 0))
			Vec_IntFillTwo(vFanins, 2, Wlc_ObjId(pNew, pConst), Wlc_ObjId(pNew, pCounter))
			Wlc_ObjAddFanins(pNew, pCond, vFanins)
			// create MUX
			pMux = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_MUX, 0, Wlc_ObjRange(pFlop)-1, 0))
			Vec_IntClear(vFanins)
			Vec_IntPush(vFanins, Wlc_ObjId(pNew, pCond))
			Vec_IntPush(vFanins, Wlc_ObjId(pNew, pFlop))
			Vec_IntPush(vFanins, Wlc_ObjId(pNew, pFanin))
			Wlc_ObjAddFanins(pNew, pMux, vFanins)
			Wlc_ObjSetCo(pNew, pMux, 1)
		}
	}
}
func Wlc_NtkAbsCreateLogic(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, vNodeFrames *Vec_Int_t, vFanins *Vec_Int_t, vNewObjs *Vec_Int_t, vConstrs *Vec_Wec_t, pConstrOut *Wlc_Obj_t) {
	var vTrace *Vec_Int_t
	var vBitConstr *Vec_Int_t = Vec_IntAlloc(100)
	var vLevConstr *Vec_Int_t = Vec_IntAlloc(100)
	var pAddrs [2]*Wlc_Obj_t
	var pDatas [2]*Wlc_Obj_t
	var pCond *Wlc_Obj_t
	var pConstr *Wlc_Obj_t = nil
	var i int32
	var k int32
	var Entry int32
	var Index [2]int32
	var iFrameLast int32
	var iFrameThis int32
	assert(Vec_IntSize(vNewObjs) == 2*Vec_IntSize(vNodeFrames), "Vec_IntSize(vNewObjs) == 2 * Vec_IntSize(vNodeFrames)")
	for i = 0; i < Vec_WecSize(vConstrs); i++ {
		vTrace = Vec_WecEntry(vConstrs, i)
		if Vec_IntSize(vTrace) == 0 {
			continue
		}
		assert(Vec_IntSize(vTrace) >= 2, "Vec_IntSize(vTrace) >= 2")
		Vec_IntClear(vLevConstr)

		iFrameThis = (Vec_IntEntry(vTrace, 0) >> 1) & 0x3FF
		iFrameLast = (Vec_IntEntryLast(vTrace) >> 1) & 0x3FF

		Index[0] = Vec_IntFind(vNodeFrames, Vec_IntEntry(vTrace, 0))
		Index[1] = Vec_IntFind(vNodeFrames, Vec_IntEntryLast(vTrace))
		assert(Index[0] >= 0 && Index[1] >= 0, "Index[0] >= 0 && Index[1] >= 0")

		pAddrs[0] = Wlc_NtkObj(pNew, Vec_IntEntry(vNewObjs, 2*Index[0]))
		pAddrs[1] = Wlc_NtkObj(pNew, Vec_IntEntry(vNewObjs, 2*Index[1]))

		pDatas[0] = Wlc_NtkObj(pNew, Vec_IntEntry(vNewObjs, 2*Index[0]+1))
		pDatas[1] = Wlc_NtkObj(pNew, Vec_IntEntry(vNewObjs, 2*Index[1]+1))

		// redirect the source in the last time frame to point to FOs
		if iFrameThis == iFrameLast {
			pAddrs[0] = Wlc_ObjFo2Fi(pNew, pAddrs[0])
			pDatas[0] = Wlc_ObjFo2Fi(pNew, pDatas[0])
		}

		// redirect the sink in the last time frame to point to FOs
		pAddrs[1] = Wlc_ObjFo2Fi(pNew, pAddrs[1])
		pDatas[1] = Wlc_ObjFo2Fi(pNew, pDatas[1])

		// equal addresses
		pCond = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_COMP_EQU, 0, 0, 0))
		Vec_IntFillTwo(vFanins, 2, Wlc_ObjId(pNew, pAddrs[0]), Wlc_ObjId(pNew, pAddrs[1]))
		Wlc_ObjAddFanins(pNew, pCond, vFanins)
		Vec_IntPush(vLevConstr, Wlc_ObjId(pNew, pCond))

		// different datas
		pCond = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_COMP_NOTEQU, 0, 0, 0))
		Vec_IntFillTwo(vFanins, 2, Wlc_ObjId(pNew, pDatas[0]), Wlc_ObjId(pNew, pDatas[1]))
		Wlc_ObjAddFanins(pNew, pCond, vFanins)
		Vec_IntPush(vLevConstr, Wlc_ObjId(pNew, pCond))

		Vec_IntForEachEntryStartStop(vTrace, Entry, k, 1, Vec_IntSize(vTrace)-1)
		{
			iFrameThis = (Entry >> 1) & 0x3FF
			Index[0] = Vec_IntFind(vNodeFrames, Entry)
			assert(Index[0] >= 0, "Index[0] >= 0")
			pAddrs[0] = Wlc_NtkObj(pNew, Vec_IntEntry(vNewObjs, 2*Index[0]))
			// redirect the middle in the last time frame to point to FOs
			if iFrameThis == iFrameLast {
				pAddrs[0] = Wlc_ObjFo2Fi(pNew, pAddrs[0])
			}
			if Vec_IntEntry(vNewObjs, 2*Index[0]+1) == 0 { // MUX

				assert(Wlc_ObjType(Wlc_NtkObj(p, Entry>>11)) == WLC_OBJ_MUX, "Wlc_ObjType(Wlc_NtkObj(p, Entry >> 11)) == WLC_OBJ_MUX")
				Vec_IntPush(vLevConstr, Wlc_ObjId(pNew, pAddrs[0]))

			} else { // different addresses

				assert(Wlc_ObjRange(pAddrs[0]) == Wlc_ObjRange(pAddrs[1]), "Wlc_ObjRange(pAddrs[0]) == Wlc_ObjRange(pAddrs[1])")
				pCond = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_COMP_NOTEQU, 0, 0, 0))
				Vec_IntFillTwo(vFanins, 2, Wlc_ObjId(pNew, pAddrs[0]), Wlc_ObjId(pNew, pAddrs[1]))
				Wlc_ObjAddFanins(pNew, pCond, vFanins)
				Vec_IntPush(vLevConstr, Wlc_ObjId(pNew, pCond))
			}
		}

		// create last output
		pConstr = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_BIT_CONCAT, 0, Vec_IntSize(vLevConstr)-1, 0))
		Wlc_ObjAddFanins(pNew, pConstr, vLevConstr)
		Vec_IntFill(vFanins, 1, Wlc_ObjId(pNew, pConstr))
		pConstr = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_REDUCT_AND, 0, 0, 0))
		Wlc_ObjAddFanins(pNew, pConstr, vFanins)

		// save the result
		Vec_IntPush(vBitConstr, Wlc_ObjId(pNew, pConstr))
	}
	if Vec_IntSize(vBitConstr) > 0 {
		// create last output
		pConstr = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_BIT_CONCAT, 0, Vec_IntSize(vBitConstr)-1, 0))
		Wlc_ObjAddFanins(pNew, pConstr, vBitConstr)
		// create last output
		Vec_IntFill(vFanins, 1, Wlc_ObjId(pNew, pConstr))
		pConstr = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_REDUCT_OR, 0, 0, 0))
		Wlc_ObjAddFanins(pNew, pConstr, vFanins)

	} else {
		pConstr = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_CONST, 0, 0, 0))
		Vec_IntFill(vFanins, 1, 0) // const0 - always true
		Wlc_ObjAddFanins(pNew, pConstr, vFanins)
	}
	// cleanup
	Vec_IntFree(vBitConstr)
	Vec_IntFree(vLevConstr)
	// add the constraint as fanin to the output
	Vec_IntFill(vFanins, 1, Wlc_ObjId(pNew, pConstr))
	Wlc_ObjAddFanins(pNew, pConstrOut, vFanins)
}
func Wlc_NtkAbstractMemory(p *Wlc_Ntk_t, vMemObjs *Vec_Int_t, vMemFanins *Vec_Int_t, piFirstMemPi *int32, piFirstCi *int32, piFirstMemCi *int32, vConstrs *Vec_Wec_t, vNodeFrames *Vec_Int_t) *Wlc_Ntk_t {
	var pNew, pTemp *Wlc_Ntk_t
	var pObj *Wlc_Obj_t
	var pCounter *Wlc_Obj_t
	var pConst *Wlc_Obj_t
	var pAdder *Wlc_Obj_t
	var pConstr *Wlc_Obj_t = nil
	var vNewObjs *Vec_Int_t = nil
	var vFanins *Vec_Int_t = Vec_IntAlloc(100)
	var i int32
	var Po0 int32
	var Po1 int32
	var AdderBits int32 = 16
	var nBits int32

	// mark memory nodes
	Wlc_NtkCleanMarks(p)
	Wlc_NtkForEachObjVec(vMemObjs, p, pObj, i)
	pObj.Mark = 1

	// start new network
	Wlc_NtkCleanCopy(p)
	pNew = Wlc_NtkAlloc(p.pName, p.nObjsAlloc+1000)
	pNew.fSmtLib = p.fSmtLib
	pNew.fAsyncRst = p.fAsyncRst
	pNew.fMemPorts = p.fMemPorts
	pNew.fEasyFfs = p.fEasyFfs
	pNew.vInits = Vec_IntAlloc(100)

	// duplicate PIs
	Wlc_NtkForEachPi(p, pObj, i)
	if !pObj.Mark {
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
		nBits += Wlc_ObjRange(pObj)
	}

	// create new PIs
	*piFirstMemPi = nBits
	Wlc_NtkForEachObjVec(vMemObjs, p, pObj, i)
	if Wlc_ObjType(pObj) == WLC_OBJ_READ {
		Wlc_NtkDupOneObject(pNew, p, pObj, WLC_OBJ_PI, vFanins)
		nBits += Wlc_ObjRange(pObj)
	}

	// duplicate flop outputs
	*piFirstCi = nBits
	Wlc_NtkForEachCi(p, pObj, i)
	if !Wlc_ObjIsPi(pObj) && !pObj.Mark {
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
		Vec_IntPush(pNew.vInits, Vec_IntEntry(p.vInits, i-Wlc_NtkPiNum(p)))
		nBits += Wlc_ObjRange(pObj)
	}

	// create flop for time-frame counter
	pCounter = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_FO, 0, AdderBits-1, 0))
	Vec_IntPush(pNew.vInits, -AdderBits)
	nBits += AdderBits

	// create flops for memory objects
	*piFirstMemCi = nBits
	if vMemFanins != nil {
		Wlc_NtkForEachObjVec(vMemFanins, p, pObj, i)
		Wlc_NtkDupOneObject(pNew, p, pObj, WLC_OBJ_FO, vFanins)
	}

	// create flops for constraint objects
	if vConstrs != nil {
		vNewObjs = Wlc_NtkAbsCreateFlopOutputs(pNew, p, vNodeFrames, vFanins)
	}

	// duplicate objects
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if !Wlc_ObjIsCi(pObj) && !pObj.Mark {
			Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
		}
	}

	if Vec_IntSize(&p.vPoPairs) != 0 {
		// create miter for the PO pairs
		Vec_IntForEachEntryDouble(&p.vPoPairs, Po0, Po1, i)
		{
			var iCopy0 int32 = Wlc_ObjCopy(p, Wlc_ObjId(p, Wlc_NtkPo(p, Po0)))
			var iCopy1 int32 = Wlc_ObjCopy(p, Wlc_ObjId(p, Wlc_NtkPo(p, Po1)))
			var iObj int32 = Wlc_ObjAlloc(pNew, WLC_OBJ_COMP_NOTEQU, 0, 0, 0)
			var pObjNew *Wlc_Obj_t = Wlc_NtkObj(pNew, iObj)
			Vec_IntFillTwo(vFanins, 1, iCopy0, iCopy1)
			Wlc_ObjAddFanins(pNew, pObjNew, vFanins)
			Wlc_ObjSetCo(pNew, pObjNew, 0)
		}
		fmt.Printf("Memory abstraction created %d miter outputs.\n", Wlc_NtkPoNum(pNew))
		for i = 0; i < Wlc_NtkCoNum(p); i++ {
			pObj = Wlc_NtkCo(p, i)
			if pObj.fIsFi && !pObj.Mark {
				Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), pObj.fIsFi)
			}
		}

		// create constraint output
		if vConstrs != nil {
			pConstr = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_BUF, 0, 0, 0))
			Wlc_ObjSetCo(pNew, pConstr, 0)
		}

	} else {
		// duplicate POs
		for i = 0; i < Wlc_NtkPoNum(p); i++ {
			pObj = Wlc_NtkPo(p, i)
			if !pObj.Mark {
				Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), pObj.fIsFi)
			}
		}
		// create constraint output
		if vConstrs != nil {
			pConstr = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_BUF, 0, 0, 0))
			Wlc_ObjSetCo(pNew, pConstr, 0)
		}
		// duplicate FIs
		for i = 0; i < Wlc_NtkCoNum(p); i++ {
			pObj = Wlc_NtkCo(p, i)
			if !Wlc_ObjIsPo(pObj) && !pObj.Mark {
				Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), pObj.fIsFi)
			}
		}
	}

	// create time-frame counter
	pConst = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_CONST, 0, AdderBits-1, 0))
	Vec_IntFill(vFanins, 1, 1)
	Wlc_ObjAddFanins(pNew, pConst, vFanins)

	pAdder = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_ARI_ADD, 0, AdderBits-1, 0))
	Vec_IntFillTwo(vFanins, 2, Wlc_ObjId(pNew, pCounter), Wlc_ObjId(pNew, pConst))
	Wlc_ObjAddFanins(pNew, pAdder, vFanins)
	Wlc_ObjSetCo(pNew, pAdder, 1)

	// create new flop inputs
	if vMemFanins != nil {
		Wlc_NtkForEachObjVec(vMemFanins, p, pObj, i)
		Wlc_NtkDupOneBuffer(pNew, p, pObj, Wlc_ObjCopy(p, Wlc_ObjId(p, pObj)), vFanins, 1)
	}

	// create flop inputs for constraint objects
	if vConstrs != nil {
		Wlc_NtkAbsCreateFlopInputs(pNew, p, vNodeFrames, vFanins, vNewObjs, pCounter, AdderBits)
	}

	// create constraint logic nodes
	if vConstrs != nil {
		Wlc_NtkAbsCreateLogic(pNew, p, vNodeFrames, vFanins, vNewObjs, vConstrs, pConstr)
	}

	// append init states
	pNew.pInits = Wlc_PrsConvertInitValues(pNew)
	if p.pSpec {
		pNew.pSpec = Abc_UtilStrsav(p.pSpec)
	}
	//Wlc_NtkTransferNames( pNew, p );
	Vec_IntFree(vFanins)
	Vec_IntFreeP(&vNewObjs)
	Wlc_NtkCleanMarks(p)
	// derive topological order
	pTemp = pNew
	pNew = Wlc_NtkDupDfs(pTemp, 0, 1)
	Wlc_NtkFree(pTemp)
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Derives values of memory subsystem objects under the CEX.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkDeriveFirstTotal(p *Wlc_Ntk_t, vMemObjs *Vec_Int_t, vMemFanins *Vec_Int_t, iFirstMemPi int32, iFirstMemCi int32, fVerbose int32) *Vec_Int_t {
	var vFirstTotal *Vec_Int_t = Vec_IntStart(3 * Vec_IntSize(vMemObjs)) // contains pairs of (first CO bit: range) for each input/output of a node
	var pObj, pFanin *Wlc_Obj_t
	var i, k, iFanin, nMemFanins int32
	for i = 0; i < Vec_IntSize(vMemObjs); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjs, i))
		if Wlc_ObjType(pObj) == WLC_OBJ_MUX {
			//Vec_IntPush( vMemFanins, Wlc_ObjFaninId0(pObj) );
			pFanin = Wlc_ObjFanin0(p, pObj)
			assert(Wlc_ObjRange(pFanin) == 1, "Wlc_ObjRange(pFanin) == 1")
			Vec_IntWriteEntry(vFirstTotal, 3*i, (iFirstMemCi<<10)|Wlc_ObjRange(pFanin))
			iFirstMemCi += Wlc_ObjRange(pFanin)
			nMemFanins++

		} else if Wlc_ObjType(pObj) == WLC_OBJ_READ || Wlc_ObjType(pObj) == WLC_OBJ_WRITE {
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				if k > 0 {
					//Vec_IntPush( vMemFanins, iFanin );
					pFanin = Wlc_NtkObj(p, iFanin)
					Vec_IntWriteEntry(vFirstTotal, 3*i+k, (iFirstMemCi<<10)|Wlc_ObjRange(pFanin))
					iFirstMemCi += Wlc_ObjRange(pFanin)
					nMemFanins++
				}
			}
			if Wlc_ObjType(pObj) == WLC_OBJ_READ {
				Vec_IntWriteEntry(vFirstTotal, 3*i+2, (iFirstMemPi<<10)|Wlc_ObjRange(pObj))
				iFirstMemPi += Wlc_ObjRange(pObj)
			}
		}
	}
	assert(nMemFanins == Vec_IntSize(vMemFanins), "nMemFanins == Vec_IntSize(vMemFanins)")
	if fVerbose != 0 {
		for i = 0; i < Vec_IntSize(vFirstTotal); i++ {
			iFanin = Vec_IntEntry(vFirstTotal, i)
			fmt.Printf("Obj %5d  Fanin %5d : ", i/3, i%3)
			fmt.Printf("%16s : %d(%d)\n", Wlc_ObjName(p, Vec_IntEntry(vMemObjs, i/3)), iFanin>>10, iFanin&0x3FF)
		}
	}
	return vFirstTotal
}
func Wlc_NtkCexResim(pAbs *Gia_Man_t, p *Abc_Cex_t, vFirstTotal *Vec_Int_t, iBit int32, vRes *Vec_Wrd_t, iFrame int32) int32 {
	var pObj, pObjRi, pObjRo *Gia_Obj_t
	var k, b, Entry, First, nBits int32
	var Value word
	// assuming that flop outputs have been assigned in the previous timeframe
	// assign primary inputs
	for k = 0; k < Gia_ManPiNum(pAbs); k++ {
		pObj = Gia_ManPi(pAbs, k)
		pObj.fMark0 = Abc_InfoHasBit(p.pData, iBit)
		iBit++
	}
	// simulate
	for k = 0; k < Gia_ManObjNum(pAbs); k++ {
		pObj = Gia_ManObj(pAbs, k)
		if Gia_ObjIsAnd(pObj) {
			pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj))
		}
	}
	// transfer to combinational outputs
	for k = 0; k < Gia_ManCoNum(pAbs); k++ {
		pObj = Gia_ManCo(pAbs, k)
		pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)
	}
	// transfer values to flop outputs
	for k = 0; k < Gia_ManRegNum(pAbs); k++ {
		pObjRi = Gia_ManRi(pAbs, k)
		pObjRo = Gia_ManRo(pAbs, k)
		pObjRo.fMark0 = pObjRi.fMark0
	}
	// at this point PIs and FOs area assigned according to this time-frame
	// collect values
	for k = 0; k < Vec_IntSize(vFirstTotal); k++ {
		Entry = Vec_IntEntry(vFirstTotal, k)
		if Entry == 0 {
			Vec_WrdWriteEntry(vRes, Vec_IntSize(vFirstTotal)*iFrame+k, ^word(0))
			continue
		}
		First = Entry >> 10
		assert(First+nBits <= Gia_ManCiNum(pAbs), "First+nBits <= Gia_ManCiNum(pAbs)")
		nBits = Entry & 0x3FF
		assert(nBits <= 64, "nBits <= 64")
		Value = 0
		for b = 0; b < nBits; b++ {
			if Gia_ManCi(pAbs, First+b).fMark0 {
				Value |= word(1) << uint(b)
			}
		}
		Vec_WrdWriteEntry(vRes, Vec_IntSize(vFirstTotal)*iFrame+k, Value)
	}
	return iBit
}
func Wlc_NtkConvertCex(vFirstTotal *Vec_Int_t, pAbs *Gia_Man_t, pCex *Abc_Cex_t, fVerbose int32) *Vec_Wrd_t {
	var vValues *Vec_Wrd_t = Vec_WrdStartFull(Vec_IntSize(vFirstTotal) * (pCex.iFrame + 1))
	var k int32
	var f int32
	var nBits int32 = pCex.nRegs
	var Value word
	Gia_ManCleanMark0(pAbs) // init FOs to zero
	for f = 0; f <= pCex.iFrame; f++ {
		nBits = Wlc_NtkCexResim(pAbs, pCex, vFirstTotal, nBits, vValues, f)
	}
	if fVerbose != 0 {
		for k = 0; k < Vec_WrdSize(vValues); k++ {
			Value = Vec_WrdEntry(vValues, k)
			if k%Vec_IntSize(vFirstTotal) == 0 {
				fmt.Printf("Frame %d:\n", k/Vec_IntSize(vFirstTotal))
			}
			fmt.Printf("Obj %5d  Fanin %5d : ", k/3, k%3)
			Extra_PrintBinary(os.Stdout, &Value, 32)
			fmt.Printf("\n")
		}
	}
	return vValues
}

/*
*Function*************************************************************

	Synopsis    [Derives the reason for failure in terms of memory values.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkTrace_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, iFrame int32, vMemObjs *Vec_Int_t, vValues *Vec_Wrd_t, ValueA word, vRes *Vec_Int_t) {
	var pFanin *Wlc_Obj_t
	var iObj int32 = Wlc_ObjId(p, pObj)
	var iNum int32 = Wlc_ObjCopy(p, iObj)
	assert(iObj == Vec_IntEntry(vMemObjs, iNum), "iObj == Vec_IntEntry(vMemObjs, iNum)")
	assert(iFrame >= 0, "iFrame >= 0")
	if Wlc_ObjIsPi(pObj) {
		Vec_IntPush(vRes, (iObj<<11)|(iFrame<<1))
	} else if Wlc_ObjIsCi(pObj) && iFrame == 0 {
		var PiId int32 = Vec_IntEntry(p.vInits, Wlc_ObjCiId(pObj)-Wlc_NtkPiNum(p))
		var iPiObj int32 = Wlc_ObjId(p, Wlc_NtkPi(p, PiId))
		Vec_IntPush(vRes, (iPiObj<<11)|(iFrame<<1))

	} else if Wlc_ObjIsCi(pObj) {
		Wlc_NtkTrace_rec(p, Wlc_ObjFo2Fi(p, pObj), iFrame-1, vMemObjs, vValues, ValueA, vRes)
	} else if Wlc_ObjType(pObj) == WLC_OBJ_BUF {
		Wlc_NtkTrace_rec(p, Wlc_ObjFanin0(p, pObj), iFrame, vMemObjs, vValues, ValueA, vRes)
	} else if Wlc_ObjType(pObj) == WLC_OBJ_MUX {
		var Index int32 = 3 * (iFrame*Vec_IntSize(vMemObjs) + iNum)
		var Value int32 = int32(Vec_WrdEntry(vValues, Index))
		assert(Value == 0 || Value == 1, "Value == 0 || Value == 1")
		pFanin = Wlc_ObjFanin1(p, pObj)
		if Value != 0 {
			pFanin = Wlc_ObjFanin2(p, pObj)
		}
		Wlc_NtkTrace_rec(p, pFanin, iFrame, vMemObjs, vValues, ValueA, vRes)
		Vec_IntPush(vRes, (iObj<<11)|(iFrame<<1)|Value)

	} else if Wlc_ObjType(pObj) == WLC_OBJ_WRITE {
		var Index int32 = 3 * (iFrame*Vec_IntSize(vMemObjs) + iNum)
		if Vec_WrdEntry(vValues, Index+1) != ValueA { // address
			Wlc_NtkTrace_rec(p, Wlc_ObjFanin0(p, pObj), iFrame, vMemObjs, vValues, ValueA, vRes)
		}
		Vec_IntPush(vRes, (iObj<<11)|(iFrame<<1))

	} else {
		assert(false, "0")
	}
}
func Wlc_NtkTrace(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, iFrame int32, vMemObjs *Vec_Int_t, vValues *Vec_Wrd_t) *Vec_Int_t {
	var iObj int32 = Wlc_ObjId(p, pObj)
	var iNum int32 = Wlc_ObjCopy(p, iObj)
	var vTrace *Vec_Int_t = Vec_IntAlloc(10)
	assert(Wlc_ObjType(pObj) == WLC_OBJ_READ, "Wlc_ObjType(pObj) == WLC_OBJ_READ")
	assert(iObj == Vec_IntEntry(vMemObjs, iNum), "iObj == Vec_IntEntry(vMemObjs, iNum)")
	Wlc_NtkTrace_rec(p, Wlc_ObjFanin0(p, pObj), iFrame, vMemObjs, vValues, Vec_WrdEntry(vValues, 3*(iFrame*Vec_IntSize(vMemObjs)+iNum)+1), vTrace)
	Vec_IntPush(vTrace, (iObj<<11)|(iFrame<<1))
	return vTrace
}
func Wlc_NtkTraceCheckConfict(p *Wlc_Ntk_t, vTrace *Vec_Int_t, vMemObjs *Vec_Int_t, vValues *Vec_Wrd_t) int32 {
	var pObjLast, pObjFirst *Wlc_Obj_t
	var iObjLast int32 = Vec_IntEntryLast(vTrace) >> 11
	var iFrmLast int32 = (Vec_IntEntryLast(vTrace) >> 1) & 0x3FF
	var iNumLast int32 = Wlc_ObjCopy(p, iObjLast)
	var iIndLast int32 = 3 * (iFrmLast*Vec_IntSize(vMemObjs) + iNumLast)
	var iObjFirst int32 = Vec_IntEntry(vTrace, 0) >> 11
	var iFrmFirst int32 = (Vec_IntEntry(vTrace, 0) >> 1) & 0x3FF
	var iNumFirst int32 = Wlc_ObjCopy(p, iObjFirst)
	var iIndFirst int32 = 3 * (iFrmFirst*Vec_IntSize(vMemObjs) + iNumFirst)
	assert(Vec_IntSize(vTrace) >= 2, "Vec_IntSize(vTrace) >= 2")
	assert(iObjLast == Vec_IntEntry(vMemObjs, iNumLast), "iObjLast  == Vec_IntEntry(vMemObjs, iNumLast)")
	assert(iObjFirst == Vec_IntEntry(vMemObjs, iNumFirst), "iObjFirst == Vec_IntEntry(vMemObjs, iNumFirst)")
	pObjLast = Wlc_NtkObj(p, iObjLast)
	pObjFirst = Wlc_NtkObj(p, iObjFirst)
	assert(Wlc_ObjType(pObjLast) == WLC_OBJ_READ, "Wlc_ObjType(pObjLast) == WLC_OBJ_READ")
	assert(Wlc_ObjType(pObjFirst) == WLC_OBJ_WRITE || Wlc_ObjIsPi(pObjFirst), "Wlc_ObjType(pObjFirst) == WLC_OBJ_WRITE || Wlc_ObjIsPi(pObjFirst)")
	if Wlc_ObjIsPi(pObjFirst) {
		return 0
	}
	assert(Vec_WrdEntry(vValues, iIndLast+1) == Vec_WrdEntry(vValues, iIndFirst+1), "Vec_WrdEntry(vValues, iIndLast + 1) == Vec_WrdEntry(vValues, iIndFirst + 1)") // equal addr
	return Vec_WrdEntry(vValues, iIndLast+2) != Vec_WrdEntry(vValues, iIndFirst+2)                                                                                 // diff data
}
func Wlc_NtkFindConflict(p *Wlc_Ntk_t, vMemObjs *Vec_Int_t, vValues *Vec_Wrd_t, nFrames int32) *Vec_Int_t {
	var pObj *Wlc_Obj_t
	var i, f, Entry int32
	var vTraces *Vec_Wec_t = Vec_WecAlloc(100)
	var vTrace, vTrace1, vTrace2 *Vec_Int_t
	assert(3*nFrames*Vec_IntSize(vMemObjs) == Vec_WrdSize(vValues), "3 * nFrames * Vec_IntSize(vMemObjs) == Vec_WrdSize(vValues)")
	Wlc_NtkCleanCopy(p)
	Vec_IntForEachEntry(vMemObjs, Entry, i)
	Wlc_ObjSetCopy(p, Entry, i)
	for f = 0; f < nFrames; f++ {
		for i = 0; i < Vec_IntSize(vMemObjs); i++ {
			pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjs, i))
			if Wlc_ObjType(pObj) != WLC_OBJ_READ {
				continue
			}
			vTrace = Wlc_NtkTrace(p, pObj, f, vMemObjs, vValues)
			if Wlc_NtkTraceCheckConfict(p, vTrace, vMemObjs, vValues) != 0 {
				Vec_WecFree(vTraces)
				return vTrace
			}
			Vec_WecPushLevel(vTraces)
			Vec_IntAppend(Vec_WecEntryLast(vTraces), vTrace)
			Vec_IntFree(vTrace)
		}
	}
	// check if there are any common read addresses
	Vec_WecForEachLevel(vTraces, vTrace1, i)
	Vec_WecForEachLevelStop(vTraces, vTrace2, f, i)
	if Vec_IntEntry(vTrace1, 0) == Vec_IntEntry(vTrace2, 0) {
		var iObj1 int32 = Vec_IntEntryLast(vTrace1) >> 11
		var iFrm1 int32 = (Vec_IntEntryLast(vTrace1) >> 1) & 0x3FF
		var iNum1 int32 = Wlc_ObjCopy(p, iObj1)
		var iInd1 int32 = 3 * (iFrm1*Vec_IntSize(vMemObjs) + iNum1)

		var iObj2 int32 = Vec_IntEntryLast(vTrace2) >> 11
		var iFrm2 int32 = (Vec_IntEntryLast(vTrace2) >> 1) & 0x3FF
		var iNum2 int32 = Wlc_ObjCopy(p, iObj2)
		var iInd2 int32 = 3 * (iFrm2*Vec_IntSize(vMemObjs) + iNum2)

		assert(iObj1 == Vec_IntEntry(vMemObjs, iNum1), "iObj1 == Vec_IntEntry(vMemObjs, iNum1)")
		assert(iObj2 == Vec_IntEntry(vMemObjs, iNum2), "iObj2 == Vec_IntEntry(vMemObjs, iNum2)")
		if Vec_WrdEntry(vValues, iInd1+1) == Vec_WrdEntry(vValues, iInd2+1) && // equal address
			Vec_WrdEntry(vValues, iInd1+2) != Vec_WrdEntry(vValues, iInd2+2) { // diff data

			vTrace = Vec_IntAlloc(100)
			Vec_IntPush(vTrace, Vec_IntPop(vTrace1))
			Vec_IntForEachEntryStart(vTrace1, Entry, i, 1)
			Vec_IntPushUnique(vTrace, Entry)
			Vec_IntForEachEntryStart(vTrace2, Entry, i, 1)
			Vec_IntPushUnique(vTrace, Entry)
			Vec_WecFree(vTraces)
			return vTrace
		}
	}
	Vec_WecFree(vTraces)
	return nil
}
func Wlc_NtkPrintConflict(p *Wlc_Ntk_t, vTrace *Vec_Int_t) {
	var Entry, i int32
	fmt.Printf("Memory semantics failure trace:\n")
	Vec_IntForEachEntry(vTrace, Entry, i)
	fmt.Printf("%3d: entry %9d : obj %5d with name %16s in frame %d\n", i, Entry, Entry>>11, Wlc_ObjName(p, Entry>>11), (Entry>>1)&0x3FF)
}
func Wlc_NtkPrintCex(p *Wlc_Ntk_t, pAbs *Wlc_Ntk_t, pCex *Abc_Cex_t) {
	var pObj *Wlc_Obj_t
	var i int32
	var k int32
	var f int32
	var nBits int32
	if pCex != nil {
		nBits = pCex.nRegs
	} else {
		nBits = 0
	}
	if pCex == nil {
		fmt.Printf("The CEX is NULL.\n")
		return
	}
	for f = 0; f <= pCex.iFrame; f++ {
		fmt.Printf("Frame%02d ", f)
		for i = 0; i < Wlc_NtkPiNum(pAbs); i++ {
			pObj = Wlc_NtkPi(pAbs, i)
			fmt.Printf("PI%d:", i)
			//printf( "%d(%d):", nBits, Wlc_ObjRange(pObj) );
			for k = 0; k < Wlc_ObjRange(pObj); k++ {
				fmt.Printf("%d", Abc_InfoHasBit(pCex.pData, nBits))
				nBits++
			}
			fmt.Printf(" ")
		}
		fmt.Printf("FF:")
		for k = 0; nBits < pCex.nPis; k++ {
			fmt.Printf("%d", Abc_InfoHasBit(pCex.pData, nBits))
			nBits++
		}
		fmt.Printf("\n")
	}

}

/*
*Function*************************************************************

	Synopsis    [Perform abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkMemAbstractTest(p *Wlc_Ntk_t) *Wlc_Ntk_t {
	var iFirstMemPi, iFirstCi, iFirstMemCi, nDcBits int32
	var vRefine *Vec_Int_t
	var vMemObjs *Vec_Int_t = Wlc_NtkCollectMemory(p, 0)
	var vMemFanins *Vec_Int_t = Wlc_NtkCollectMemFanins(p, vMemObjs)

	var vRefines *Vec_Wec_t = Vec_WecAlloc(100)
	var vNodeFrames *Vec_Int_t = Vec_IntAlloc(100)
	var pNewFull *Wlc_Ntk_t

	vRefine = Vec_WecPushLevel(vRefines)
	Vec_IntPush(vRefine, (11<<11)|0)
	Vec_IntPush(vRefine, (10<<11)|0)
	Vec_IntPush(vRefine, (8<<11)|0)
	Vec_IntPush(vRefine, (9<<11)|0)
	Wlc_NtkAbsAddToNodeFrames(vNodeFrames, vRefine)

	//    pNewFull    = Wlc_NtkAbstractMemory( p, vMemObjs, vMemFanins, &iFirstMemPi, &iFirstCi, &iFirstMemCi, NULL, NULL );
	pNewFull = Wlc_NtkAbstractMemory(p, vMemObjs, nil, &iFirstMemPi, &iFirstCi, &iFirstMemCi, vRefines, vNodeFrames)

	Vec_WecFree(vRefines)
	Vec_IntFree(vNodeFrames)

	nDcBits = Wlc_CountDcs(pNewFull.pInits)
	fmt.Printf("iFirstMemPi = %d  iFirstCi = %d  iFirstMemCi = %d  nDcBits = %d\n", iFirstMemPi, iFirstCi, iFirstMemCi, nDcBits)

	Vec_IntFreeP(&vMemObjs)
	Vec_IntFreeP(&vMemFanins)
	return pNewFull
}

func Wlc_NtkMemAbstract(p *Wlc_Ntk_t, nIterMax int32, fDumpAbs int32, fPdrVerbose int32, fVerbose int32) int32 {
	var clk abctime = Abc_Clock()
	var pNewFull, pNew *Wlc_Ntk_t
	var pAig, pTempAig *Aig_Man_t
	var pAbsFull, pAbs *Gia_Man_t
	var pCex *Abc_Cex_t = nil
	var vValues *Vec_Wrd_t = nil
	var vRefines *Vec_Wec_t = Vec_WecAlloc(100)
	var vNodeFrames *Vec_Int_t = Vec_IntAlloc(100)
	var vMemObjs, vMemFanins, vFirstTotal, vRefine *Vec_Int_t
	var RetValue int32 = -1
	var iFirstMemPi int32
	var iFirstCi int32
	var iFirstMemCi int32
	var nDcBits int32
	var nIters int32

	vMemObjs = Wlc_NtkCollectMemory(p, 0)
	vMemFanins = Wlc_NtkCollectMemFanins(p, vMemObjs)
	pNewFull = Wlc_NtkAbstractMemory(p, vMemObjs, vMemFanins, &iFirstMemPi, &iFirstCi, &iFirstMemCi, nil, nil)
	nDcBits = Wlc_CountDcs(pNewFull.pInits)
	vFirstTotal = Wlc_NtkDeriveFirstTotal(p, vMemObjs, vMemFanins, iFirstMemPi, nDcBits+iFirstMemCi, fVerbose)

	pAbsFull = Wlc_NtkBitBlast(pNewFull, nil)
	assert(Gia_ManPiNum(pAbsFull) == iFirstCi+nDcBits, "Gia_ManPiNum(pAbsFull) == iFirstCi + nDcBits")
	Wlc_NtkFree(pNewFull)

	// perform abstraction
	for nIters = 0; nIters < nIterMax; nIters++ {
		// set up parameters to run PDR
		var PdrPars Pdr_Par_t
		var pPdrPars *Pdr_Par_t = &PdrPars
		Pdr_ManSetDefaultParams(pPdrPars)
		pPdrPars.fUseAbs = 0 // use 'pdr -t'  (on-the-fly abstraction)
		pPdrPars.fVerbose = fVerbose

		// derive specific abstraction
		pNew = Wlc_NtkAbstractMemory(p, vMemObjs, nil, &iFirstMemPi, &iFirstCi, &iFirstMemCi, vRefines, vNodeFrames)
		pAbs = Wlc_NtkBitBlast(pNew, nil)
		// simplify the AIG
		//pAbs = Gia_ManSeqStructSweep( pGiaTemp = pAbs, 1, 1, 0 );
		//Gia_ManStop( pGiaTemp );

		// roll in the constraints
		pAig = Gia_ManToAigSimple(pAbs)
		Gia_ManStop(pAbs)
		pAig.nConstrs = 1
		pTempAig = pAig
		pAig = Saig_ManDupFoldConstrsFunc(pTempAig, 0, 0)
		Aig_ManStop(pTempAig)
		pAbs = Gia_ManFromAigSimple(pAig)
		Aig_ManStop(pAig)

		// try to prove abstracted GIA by converting it to AIG and calling PDR
		pAig = Gia_ManToAigSimple(pAbs)
		RetValue = Pdr_ManSolve(pAig, pPdrPars)
		pCex = pAig.pSeqModel
		pAig.pSeqModel = nil
		Aig_ManStop(pAig)

		if fVerbose != 0 {
			fmt.Printf("\nITERATIONS %d:\n", nIters)
		}
		if fVerbose != 0 {
			Wlc_NtkPrintCex(p, pNew, pCex)
		}
		Wlc_NtkFree(pNew)

		if fDumpAbs != 0 {
			var pFileName string = "mem_abs.aig"
			Gia_AigerWrite(pAbs, pFileName, 0, 0, 0)
			iFrame := int32(-1)
			if pCex != nil {
				iFrame = pCex.iFrame
			}
			fmt.Printf("Iteration %3d: Dumped abstraction in file \"%s\" after finding CEX in frame %d.\n", nIters, pFileName, iFrame)
		}

		// check if proved or undecided
		if pCex == nil {
			assert((RetValue != 0), "RetValue")
			Gia_ManStop(pAbs)
			break
		}

		// analyze counter-example
		vValues = Wlc_NtkConvertCex(vFirstTotal, pAbsFull, pCex, fVerbose)
		Gia_ManStop(pAbs)
		vRefine = Wlc_NtkFindConflict(p, vMemObjs, vValues, pCex.iFrame+1)
		Vec_WrdFree(vValues)
		if vRefine == nil { // cannot refine
			break
		}
		Abc_CexFreeP(&pCex)

		// save refinement for the future
		if fVerbose != 0 {
			Wlc_NtkPrintConflict(p, vRefine)
		}
		Vec_WecPushLevel(vRefines)
		Vec_IntAppend(Vec_WecEntryLast(vRefines), vRefine)
		Wlc_NtkAbsAddToNodeFrames(vNodeFrames, vRefine)
		Vec_IntFree(vRefine)
	}
	// cleanup
	Gia_ManStop(pAbsFull)
	Vec_WecFree(vRefines)
	Vec_IntFreeP(&vMemObjs)
	Vec_IntFreeP(&vMemFanins)
	Vec_IntFreeP(&vFirstTotal)
	Vec_IntFreeP(&vNodeFrames)

	// report the result
	if fVerbose != 0 {
		fmt.Printf("\n")
	}
	fmt.Printf("Abstraction ")
	if RetValue == 0 && (pCex != nil) {
		fmt.Printf("resulted in a real CEX in frame %d", pCex.iFrame)
	} else if RetValue == 1 {
		fmt.Printf("is successfully proved")
	} else {
		fmt.Printf("timed out")
	}
	fmt.Printf(" after %d iterations. ", nIters)
	Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	Abc_CexFreeP(&pCex) // return CEX in the future
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Collect visited objects.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkExploreMem2_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vCollect *Vec_Int_t, nFrames int32) int32 {
	var k, iFanin, nVisited int32
	if pObj.Mark == 0 {
		return 0
	}
	if Wlc_ObjType(pObj) == WLC_OBJ_PI || (Wlc_ObjType(pObj) == WLC_OBJ_FO && nFrames == 0) {
		Vec_IntPushTwo(vCollect, Wlc_ObjId(p, pObj), nFrames)
		return 1
	}
	if Wlc_ObjType(pObj) == WLC_OBJ_FO {
		return Wlc_NtkExploreMem2_rec(p, Wlc_ObjFo2Fi(p, pObj), vCollect, nFrames-1)
	}
	Wlc_ObjForEachFanin(pObj, iFanin, k)
	nVisited += Wlc_NtkExploreMem2_rec(p, Wlc_NtkObj(p, iFanin), vCollect, nFrames)
	Vec_IntPushTwo(vCollect, Wlc_ObjId(p, pObj), nFrames)
	return nVisited + 1
}
func Wlc_NtkExploreMem2(p *Wlc_Ntk_t, nFrames int32) {
	var pObj *Wlc_Obj_t
	var i, k, Entry, Frame, nVisited int32
	var vCollect *Vec_Int_t = Vec_IntStart(1000)
	var vMemObjsClean *Vec_Int_t = Wlc_NtkCollectMemory(p, 1)
	Wlc_NtkCleanMarks(p)
	Wlc_NtkForEachObjVec(vMemObjsClean, p, pObj, i)
	pObj.Mark = 1
	for i = 0; i < Vec_IntSize(vMemObjsClean); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, i))
		if Wlc_ObjType(pObj) != WLC_OBJ_READ {
			continue
		}
		Vec_IntClear(vCollect)
		nVisited = Wlc_NtkExploreMem2_rec(p, pObj, vCollect, nFrames)
		fmt.Printf("Obj %6d : ", Wlc_ObjId(p, pObj))
		fmt.Printf("Visit = %6d  ", nVisited)
		fmt.Printf("Pair = %6d  ", Vec_IntSize(vCollect)/2)
		if Vec_IntSize(vCollect)/2 < 10 {
			Vec_IntForEachEntryDouble(vCollect, Entry, Frame, k)
			fmt.Printf("%d(%d) ", Entry, Frame)
		}
		fmt.Printf("\n")
	}
	Vec_IntFree(vMemObjsClean)
	Vec_IntFree(vCollect)
	Wlc_NtkCleanMarks(p)
}

/*
*Function*************************************************************

	Synopsis    [Collect memory flops reachable.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkExploreMem_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vCollect *Vec_Int_t, nFrames int32) {
	var k, iFanin int32
	if pObj.Mark == 0 {
		return
	}
	if Wlc_ObjType(pObj) == WLC_OBJ_PI || (Wlc_ObjType(pObj) == WLC_OBJ_FO && nFrames == 0) {
		Vec_IntPushUnique(vCollect, Wlc_ObjId(p, pObj))
		return
	}
	if Wlc_ObjType(pObj) == WLC_OBJ_FO {
		Wlc_NtkExploreMem_rec(p, Wlc_ObjFo2Fi(p, pObj), vCollect, nFrames-1)
		return
	}
	Wlc_ObjForEachFanin(pObj, iFanin, k)
	Wlc_NtkExploreMem_rec(p, Wlc_NtkObj(p, iFanin), vCollect, nFrames)
}
func Wlc_NtkExploreMem(p *Wlc_Ntk_t, nFrames int32) {
	var pObj *Wlc_Obj_t
	var i, k, iObj int32
	var vCollect *Vec_Int_t = Vec_IntStart(1000)
	var vMemObjsClean *Vec_Int_t = Wlc_NtkCollectMemory(p, 1)
	Wlc_NtkCleanMarks(p)
	Wlc_NtkForEachObjVec(vMemObjsClean, p, pObj, i)
	pObj.Mark = 1
	for i = 0; i < Vec_IntSize(vMemObjsClean); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, i))
		if Wlc_ObjType(pObj) != WLC_OBJ_READ {
			continue
		}
		Vec_IntClear(vCollect)
		Wlc_NtkExploreMem_rec(p, pObj, vCollect, nFrames)
		fmt.Printf("Read port %6d : ", Wlc_ObjId(p, pObj))
		fmt.Printf("Inputs = %6d  ", Vec_IntSize(vCollect))
		Vec_IntForEachEntry(vCollect, iObj, k)
		fmt.Printf("%d(%s) ", iObj, Wlc_ObjName(p, iObj))
		fmt.Printf("\n")
	}
	Vec_IntFree(vMemObjsClean)
	Vec_IntFree(vCollect)
	Wlc_NtkCleanMarks(p)
}

/*
*Function*************************************************************

	Synopsis    [For each READ find reachable CIs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkFindReachablePiFo(p *Wlc_Ntk_t, vMemObjsClean *Vec_Int_t, nFrames int32) *Vec_Int_t {
	var vRes *Vec_Int_t = Vec_IntAlloc(100)
	var pObjR, pObjI, pObj *Wlc_Obj_t
	var i, j, k, f, fFanin int32
	Wlc_NtkForEachObj(p, pObj, i)
	pObj.Mark2 = 0
	// go through read ports
	for i = 0; i < Vec_IntSize(vMemObjsClean); i++ {
		pObjR = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, i))
		if !Wlc_ObjIsRead(pObjR) {
			continue
		}
		// go through memory CIs
		for j = 0; j < Vec_IntSize(vMemObjsClean); j++ {
			pObjI = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, j))
			if !Wlc_ObjIsCi(pObjI) {
				continue
			}
			// propagate bit from CI to READ
			pObjI.Mark2 = 1
			for k = 0; k < Vec_IntSize(vMemObjsClean); k++ {
				pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, k))
				if pObj == pObjI {
					continue
				}
				assert(pObj.Mark2 == 0, "pObj.Mark2 == 0")
				Wlc_ObjForEachFanin(pObj, fFanin, f)
				pObj.Mark2 |= Wlc_NtkObj(p, fFanin).Mark2
			}
			if pObjR.Mark2 {
				Vec_IntPushThree(vRes, Wlc_ObjId(p, pObjR), Wlc_ObjId(p, pObjI), -1)
			}
			for k = 0; k < Vec_IntSize(vMemObjsClean); k++ {
				pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, k))
				pObj.Mark2 = 0
			}
		}
	}
	Wlc_NtkForEachObj(p, pObj, i)
	assert(pObj.Mark2 == 0, "pObj.Mark2 == 0")
	return vRes
}
func Wlc_NtkExtractCisForThisRead(vReachReadCi *Vec_Int_t, iRead int32) *Vec_Int_t {
	var k int32
	var vRes *Vec_Int_t = Vec_IntAlloc(100)
	for k = 0; k < Vec_IntSize(vReachReadCi)/3; k++ {
		if iRead != Vec_IntEntry(vReachReadCi, 3*k) {
			continue
		}
		Vec_IntPush(vRes, Vec_IntEntry(vReachReadCi, 3*k+1))
		Vec_IntPush(vRes, Vec_IntEntry(vReachReadCi, 3*k+2))
	}
	return vRes
}
func Wlc_NtkCollectOneType(p *Wlc_Ntk_t, vMemObjsClean *Vec_Int_t, Type1 int32, Type2 int32) *Vec_Int_t {
	var pObj *Wlc_Obj_t
	var i int32
	var vRes *Vec_Int_t = Vec_IntAlloc(100)
	for i = 0; i < Vec_IntSize(vMemObjsClean); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, i))
		if Wlc_ObjType(pObj) == Type1 || Wlc_ObjType(pObj) == Type2 {
			Vec_IntPush(vRes, Wlc_ObjId(p, pObj))
		}
	}
	return vRes
}

/*
*Function*************************************************************

	Synopsis    [Create memory constraints.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkCreateMemoryConstr(pNew *Wlc_Ntk_t, p *Wlc_Ntk_t, vMemObjsClean *Vec_Int_t, vReachReadCi *Vec_Int_t) {
	var vReads *Vec_Int_t = Wlc_NtkCollectOneType(p, vMemObjsClean, WLC_OBJ_READ, -1)
	var vObjCis *Vec_Int_t = Wlc_NtkCollectOneType(p, vMemObjsClean, WLC_OBJ_PI, WLC_OBJ_FO)
	var vFanins *Vec_Int_t = Vec_IntAlloc(10)
	var pObjR, pObj, pCond *Wlc_Obj_t
	var i int32
	var k int32
	var iObjCi int32
	var iObjNew int32 = -1
	// go through read ports
	for i = 0; i < Vec_IntSize(vReads); i++ {
		pObjR = Wlc_NtkObj(p, Vec_IntEntry(vReads, i))
		// ABC_READ  ( .mem_in(mem_fi1), .addr(raddr), .data(read1) ) ;
		// ABC_WRITE ( .mem_in(mem_fo1), .addr(waddr), .data(data), .mem_out(mem_fi1) ) ;
		// initialize CI related to the read port
		var vLocalCis *Vec_Int_t = Wlc_NtkExtractCisForThisRead(vReachReadCi, Wlc_ObjId(p, pObjR))
		Wlc_NtkForEachObjVec(vObjCis, p, pObj, k)
		Wlc_ObjSetCopy(p, Wlc_ObjId(p, pObj), -1)
		Vec_IntForEachEntryDouble(vLocalCis, iObjCi, iObjNew, k)
		Wlc_ObjSetCopy(p, iObjCi, iObjNew)
		Vec_IntFree(vLocalCis)
		// implement the nodes
		for k = 0; k < Vec_IntSize(vMemObjsClean); k++ {
			pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, k))
			if Wlc_ObjIsRead(pObj) || Wlc_ObjIsCi(pObj) {
				continue
			}
			Wlc_ObjSetCopy(p, Wlc_ObjId(p, pObj), -1)
			if Wlc_ObjIsWrite(pObj) {
				if Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObj)) == -1 {
					continue
				}
				if Wlc_ObjRange(pObjR) != Wlc_ObjRange(Wlc_ObjFanin2(p, pObj)) {
					continue
				}
				// create equality
				pCond = Wlc_NtkObj(pNew, Wlc_ObjAlloc(pNew, WLC_OBJ_COMP_EQU, 0, 0, 0))
				Vec_IntFillTwo(vFanins, 2, Wlc_ObjCopy(p, Wlc_ObjFaninId1(pObjR)), Wlc_ObjCopy(p, Wlc_ObjFaninId1(pObj)))
				//printf( "%d : ", Wlc_ObjId(p,pObj) ), Vec_IntPrint( vFanins );
				Wlc_ObjAddFanins(pNew, pCond, vFanins)
				// create MUX
				iObjNew = Wlc_ObjAlloc(pNew, WLC_OBJ_MUX, 0, Wlc_ObjRange(pObjR)-1, 0)
				Vec_IntFill(vFanins, 1, Wlc_ObjId(pNew, pCond))
				Vec_IntPush(vFanins, Wlc_ObjCopy(p, Wlc_ObjFaninId2(pObj)))
				Vec_IntPush(vFanins, Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObj)))
				Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObjNew), vFanins)

			} else if Wlc_ObjIsMux(pObj) {
				var iFanin0New int32 = Wlc_ObjCopy(p, Wlc_ObjFaninId1(pObj))
				var iFanin1New int32 = Wlc_ObjCopy(p, Wlc_ObjFaninId2(pObj))
				if iFanin0New == -1 || iFanin1New == -1 {
					continue
				}
				//Wlc_NtkPrintNode( pNew, Wlc_NtkObj(pNew, iFanin0New) );
				assert(Wlc_ObjRange(pObjR) == Wlc_ObjRange(Wlc_NtkObj(pNew, iFanin0New)), "Wlc_ObjRange(pObjR) == Wlc_ObjRange(Wlc_NtkObj(pNew, iFanin0New))")
				assert(Wlc_ObjRange(pObjR) == Wlc_ObjRange(Wlc_NtkObj(pNew, iFanin1New)), "Wlc_ObjRange(pObjR) == Wlc_ObjRange(Wlc_NtkObj(pNew, iFanin1New))")
				iObjNew = Wlc_ObjAlloc(pNew, WLC_OBJ_MUX, 0, Wlc_ObjRange(pObjR)-1, 0)
				Vec_IntFill(vFanins, 1, Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObj)))
				Vec_IntPush(vFanins, iFanin0New)
				Vec_IntPush(vFanins, iFanin1New)
				Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObjNew), vFanins)

			} else if Wlc_ObjIsBuf(pObj) {
				iObjNew = Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObj))
				if iObjNew == -1 {
					continue
				}

			} else {
				assert(false, "0")
			}
			Wlc_ObjSetCopy(p, Wlc_ObjId(p, pObj), iObjNew)
		}
		// extract fanin
		iObjNew = Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObjR))
		assert(iObjNew != -1, "iObjNew != -1")
		Vec_IntFill(vFanins, 1, iObjNew)
		// add it as buffer fanin
		iObjNew = Wlc_ObjCopy(p, Wlc_ObjId(p, pObjR))
		assert(Wlc_NtkObj(pNew, iObjNew).Type == WLC_OBJ_BUF, "Wlc_NtkObj(pNew, iObjNew).Type == WLC_OBJ_BUF")
		Wlc_ObjAddFanins(pNew, Wlc_NtkObj(pNew, iObjNew), vFanins)
	}
	Wlc_NtkForEachObjVec(vObjCis, p, pObj, k)
	Wlc_ObjSetCopy(p, Wlc_ObjId(p, pObj), -1)
	Vec_IntFree(vFanins)
	Vec_IntFree(vReads)
	Vec_IntFree(vObjCis)
}

/*
*Function*************************************************************

	Synopsis    [Perform abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbstractMem(p *Wlc_Ntk_t, nFrames int32, fVerbose int32) *Wlc_Ntk_t {
	var vMemObjsAll *Vec_Int_t = Wlc_NtkCollectMemory(p, 0)
	var vMemObjsClean *Vec_Int_t = Wlc_NtkCollectMemory(p, 1)
	var vReachReadCi *Vec_Int_t = Wlc_NtkFindReachablePiFo(p, vMemObjsClean, nFrames)
	var pNew, pTemp *Wlc_Ntk_t
	var pObj *Wlc_Obj_t
	var vFanins *Vec_Int_t = Vec_IntAlloc(100)
	var i, Po0, Po1, iObjNew int32

	assert(nFrames == 0, "nFrames == 0")

	// mark memory nodes
	Wlc_NtkCleanMarks(p)
	Wlc_NtkForEachObjVec(vMemObjsAll, p, pObj, i)
	pObj.Mark = 1

	// start new network
	Wlc_NtkCleanCopy(p)
	pNew = Wlc_NtkAlloc(p.pName, p.nObjsAlloc+Vec_IntSize(vMemObjsClean)*nFrames*10)
	pNew.fSmtLib = p.fSmtLib
	pNew.fAsyncRst = p.fAsyncRst
	pNew.fMemPorts = p.fMemPorts
	pNew.fEasyFfs = p.fEasyFfs
	pNew.vInits = Vec_IntAlloc(100)

	// duplicate PIs
	Wlc_NtkForEachPi(p, pObj, i)
	if !pObj.Mark {
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}

	// create new PIs
	assert(Vec_IntSize(vReachReadCi)%3 == 0, "Vec_IntSize(vReachReadCi) % 3 == 0")
	for i = 0; i < Vec_IntSize(vReachReadCi)/3; i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vReachReadCi, 3*i))
		assert(Wlc_ObjIsRead(pObj), "Wlc_ObjIsRead(pObj)")
		iObjNew = Wlc_NtkDupOneObject(pNew, p, pObj, WLC_OBJ_PI, vFanins)
		Vec_IntWriteEntry(vReachReadCi, 3*i+2, iObjNew)
	}
	//Vec_IntPrint( vReachReadCi );

	// duplicate flop outputs
	Wlc_NtkForEachCi(p, pObj, i)
	if !Wlc_ObjIsPi(pObj) && !pObj.Mark {
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
		Vec_IntPush(pNew.vInits, Vec_IntEntry(p.vInits, i-Wlc_NtkPiNum(p)))
	}

	/*
	   // create flops for memory objects
	   Wlc_NtkForEachObjVec( vMemFanins, p, pObj, i )
	       for ( k = 0; k < nFrames; k++ )
	           Wlc_NtkDupOneObject( pNew, p, pObj, WLC_OBJ_FO, vFanins );
	*/

	// create buffers for read ports
	for i = 0; i < Vec_IntSize(vMemObjsClean); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vMemObjsClean, i))
		if !Wlc_ObjIsRead(pObj) {
			continue
		}
		iObjNew = Wlc_ObjAlloc(pNew, WLC_OBJ_BUF, pObj.Signed, pObj.End, pObj.Beg)
		Wlc_ObjSetCopy(p, Wlc_ObjId(p, pObj), iObjNew)
	}

	// duplicate objects
	Wlc_NtkForEachObj(p, pObj, i)
	if !Wlc_ObjIsCi(pObj) && !pObj.Mark {
		Wlc_ObjDup(pNew, p, Wlc_ObjId(p, pObj), vFanins)
	}

	// create memory constraints
	Wlc_NtkCreateMemoryConstr(pNew, p, vMemObjsClean, vReachReadCi)

	// create outpus
	if Vec_IntSize(&p.vPoPairs) {
		// create miter for the PO pairs
		Vec_IntForEachEntryDouble(&p.vPoPairs, Po0, Po1, i)
		{
			var iCopy0 int32 = Wlc_ObjCopy(p, Wlc_ObjId(p, Wlc_NtkPo(p, Po0)))
			var iCopy1 int32 = Wlc_ObjCopy(p, Wlc_ObjId(p, Wlc_NtkPo(p, Po1)))
			var iObj int32 = Wlc_ObjAlloc(pNew, WLC_OBJ_COMP_NOTEQU, 0, 0, 0)
			var pObjNew *Wlc_Obj_t = Wlc_NtkObj(pNew, iObj)
			Vec_IntFillTwo(vFanins, 1, iCopy0, iCopy1)
			Wlc_ObjAddFanins(pNew, pObjNew, vFanins)
			Wlc_ObjSetCo(pNew, pObjNew, 0)
		}
		fmt.Printf("Memory abstraction created %d miter outputs.\n", Wlc_NtkPoNum(pNew))
		Wlc_NtkForEachCo(p, pObj, i)
		if pObj.fIsFi && !pObj.Mark {
			Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), pObj.fIsFi)
		}

	} else {
		// duplicate POs
		Wlc_NtkForEachPo(p, pObj, i)
		if !pObj.Mark {
			Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), pObj.fIsFi)
		}
		// duplicate FIs
		Wlc_NtkForEachCo(p, pObj, i)
		if !Wlc_ObjIsPo(pObj) && !pObj.Mark {
			Wlc_ObjSetCo(pNew, Wlc_ObjCopyObj(pNew, p, pObj), pObj.fIsFi)
		}
	}

	/*
	   // create new flop inputs
	   Wlc_NtkForEachObjVec( vMemFanins, p, pObj, i )
	       for ( k = 0; k < nFrames; k++ )
	           Wlc_NtkDupOneBuffer( pNew, p, pObj, Wlc_ObjCopy(p, Wlc_ObjId(p, pObj))-(nFrames-1-k), vFanins, 1 );
	*/

	// append init states
	pNew.pInits = Wlc_PrsConvertInitValues(pNew)
	if p.pSpec {
		pNew.pSpec = Abc_UtilStrsav(p.pSpec)
	}
	if Wlc_NtkHasNameId(p) {
		Wlc_NtkTransferNames(pNew, p)
	}
	Vec_IntFree(vFanins)
	Vec_IntFree(vMemObjsAll)
	Vec_IntFree(vMemObjsClean)
	Vec_IntFree(vReachReadCi)
	Wlc_NtkCleanMarks(p)
	// derive topological order
	//printf( "Objects before = %d.\n", Wlc_NtkObjNum(pNew) );
	pTemp = pNew
	pNew = Wlc_NtkDupDfs(pTemp, 0, 1)
	Wlc_NtkFree(pTemp)
	//printf( "Objects after = %d.\n", Wlc_NtkObjNum(pNew) );
	return pNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
