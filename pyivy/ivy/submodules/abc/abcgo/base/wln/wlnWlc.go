package wln

import (
	"fmt"
	wlc "github.com/glycerine/abcgo/base/wlc"
)

/**CFile****************************************************************

  FileName    [wlnWlc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Word-level network.]

  Synopsis    [Network transformation.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 23, 2018.]

  Revision    [$Id: wlnWlc.c,v 1.00 2018/09/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wln.h"
//#include "base/wlc/wlc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Ndr_TypeWlc2Ndr( int Type );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wln_TrasformNameId(pNew *Wln_Ntk_t, p *wlc.Wlc_Ntk_t, pObj *wlc.Wlc_Obj_t) int32 {
	return Abc_NamStrFindOrAdd(pNew.pManName, wlc.Wlc_ObjName(p, wlc.Wlc_ObjId(p, pObj)), nil)
}
func Wln_NtkFromWlc(p *wlc.Wlc_Ntk_t) *Wln_Ntk_t {
	var pObj *wlc.Wlc_Obj_t
	var Buffer string // was [1000]byte in C

	var i, j, n, Type, iFanin, iOutId, iBit int32
	var vFanins *Vec_Int_t = Vec_IntAlloc(10)
	var vInits *Vec_Int_t = Vec_IntAlloc(wlc.Wlc_NtkFfNum(p))
	var pNew *Wln_Ntk_t = Wln_NtkAlloc(wlc.Wlc_NtkName(p), wlc.Wlc_NtkObjNum(p)+wlc.Wlc_NtkCoNum(p)+wlc.Wlc_NtkFfNum(p))
	pNew.pManName = Abc_NamStart(Abc_NamObjNumMax(wlc.Wlc_NtkNameMan(p)), 10)
	if wlc.Wlc_NtkSpec(p) != "" {
		pNew.pSpec = Abc_UtilStrsav(wlc.Wlc_NtkSpec(p))
	}
	pNew.fSmtLib = wlnBool(wlc.Wlc_NtkSmtLib(p))
	wlc.Wlc_NtkCleanCopy(p)
	Wln_NtkCleanNameId(pNew)
	// add primary inputs
	// inlining of macro wlc.Wlc_NtkForEachPi
	for i = 0; i < wlc.Wlc_NtkPiNum(p); i++ {
		pObj = wlc.Wlc_NtkPi(p, i)
		iOutId = Wln_ObjAlloc(pNew, ABC_OPER_CI, wlnBool(wlc.Wlc_ObjSigned(pObj)), wlc.Wlc_ObjEnd(pObj), wlc.Wlc_ObjBeg(pObj))
		Wln_ObjSetNameId(pNew, iOutId, Wln_TrasformNameId(pNew, p, pObj))
		wlc.Wlc_ObjSetCopy(p, wlc.Wlc_ObjId(p, pObj), iOutId)
	}
	// create initial state of the flops
	// inlining of macro wlc.Wlc_NtkForEachCi
	for i = 0; i < wlc.Wlc_NtkCiNum(p); i++ {
		pObj = wlc.Wlc_NtkCi(p, i)
		assert(i == wlc.Wlc_ObjCiId(pObj), "i == wlc.Wlc_ObjCiId(pObj)")
		if pObj.Type == uint32(wlc.WLC_OBJ_PI) {
			continue
		}
		for j = 0; j < wlc.Wlc_ObjRange(pObj); j++ {
			if wlc.Wlc_NtkInits(p)[iBit+j] == 'x' {
				break
			}
		}
		// print flop init state
		if true {
			fmt.Printf("Flop %3d init state: %d'b", i-wlc.Wlc_NtkPiNum(p), wlc.Wlc_ObjRange(pObj))
			if j == wlc.Wlc_ObjRange(pObj) {
				var Count int32
				for n = 0; n < wlc.Wlc_ObjRange(pObj); n++ {
					Count += wlnBool(wlc.Wlc_NtkInits(p)[iBit+n] == '0')
				}
				if Count == wlc.Wlc_ObjRange(pObj) {
					fmt.Printf("0")
				} else {
					for n = 0; n < wlc.Wlc_ObjRange(pObj); n++ {
						fmt.Printf("%c", wlc.Wlc_NtkInits(p)[iBit+n])
					}
				}

			} else {
				var Count int32
				for n = 0; n < wlc.Wlc_ObjRange(pObj); n++ {
					Count += wlnBool(wlc.Wlc_NtkInits(p)[iBit+n] == 'x')
				}
				fmt.Printf("x")
				if Count != wlc.Wlc_ObjRange(pObj) {
					fmt.Printf(" (range %d)", wlc.Wlc_ObjRange(pObj))
				}
			}
			fmt.Printf("\n")
		}
		if j == wlc.Wlc_ObjRange(pObj) {
			Type = ABC_OPER_CONST
		} else {
			Type = ABC_OPER_CI
		}
		iOutId = Wln_ObjAlloc(pNew, Type, wlnBool(wlc.Wlc_ObjSigned(pObj)), wlc.Wlc_ObjEnd(pObj), wlc.Wlc_ObjBeg(pObj))
		if j == wlc.Wlc_ObjRange(pObj) { // constant

			var pString string = Wln_ConstFromStr(wlc.Wlc_NtkInits(p)[iBit:], wlc.Wlc_ObjRange(pObj))
			Wln_ObjSetConst(pNew, iOutId, Abc_NamStrFindOrAdd(pNew.pManName, pString, nil))
			ABC_FREE(pString)
		}
		Buffer = fmt.Sprintf("ff_init_%d", Vec_IntSize(vInits))
		Wln_ObjSetNameId(pNew, iOutId, Abc_NamStrFindOrAdd(pNew.pManName, Buffer, nil))
		Vec_IntPush(vInits, iOutId)
		iBit += wlc.Wlc_ObjRange(pObj)
	}
	assert(wlc.Wlc_NtkInits(p) == "" || iBit == int32(len(wlc.Wlc_NtkInits(p))), "initial bits match input length")
	// add flop outputs
	// inlining of macro wlc.Wlc_NtkForEachCi
	for i = 0; i < wlc.Wlc_NtkCiNum(p); i++ {
		pObj = wlc.Wlc_NtkCi(p, i)
		assert(i == wlc.Wlc_ObjCiId(pObj), "i == wlc.Wlc_ObjCiId(pObj)")
		if pObj.Type == uint32(wlc.WLC_OBJ_PI) {
			continue
		}
		iOutId = Wln_ObjAlloc(pNew, ABC_OPER_DFFRSE, wlnBool(wlc.Wlc_ObjSigned(pObj)), wlc.Wlc_ObjEnd(pObj), wlc.Wlc_ObjBeg(pObj))
		Wln_ObjSetNameId(pNew, iOutId, Wln_TrasformNameId(pNew, p, pObj))
		wlc.Wlc_ObjSetCopy(p, wlc.Wlc_ObjId(p, pObj), iOutId)
	}
	// add internal nodes
	// inlining of macro wlc.Wlc_NtkForEachObj
	for i = 1; i < wlc.Wlc_NtkObjNumMax(p); i++ {
		pObj = wlc.Wlc_NtkObj(p, i)
		if wlc.Wlc_ObjIsCi(pObj) != 0 || pObj.Type == 0 {
			continue
		}
		iOutId = Wln_ObjAlloc(pNew, wlc.Ndr_TypeWlc2Ndr(int32(pObj.Type)), wlnBool(wlc.Wlc_ObjSigned(pObj)), wlc.Wlc_ObjEnd(pObj), wlc.Wlc_ObjBeg(pObj))
		Vec_IntClear(vFanins)
		// inlining of macro wlc.Wlc_ObjForEachFanin
		for n = 0; n < wlc.Wlc_ObjFaninNum(pObj); n++ {
			iFanin = wlc.Wlc_ObjFaninId(pObj, n)
			Vec_IntPush(vFanins, wlc.Wlc_ObjCopy(p, iFanin))
		}
		Wln_ObjAddFanins(pNew, iOutId, vFanins)
		if pObj.Type == uint32(wlc.WLC_OBJ_BIT_SELECT) {
			Wln_ObjSetSlice(pNew, iOutId, Hash_Int2ManInsert(pNew.pRanges, wlc.Wlc_ObjEnd(pObj), wlc.Wlc_ObjBeg(pObj), 0))
		} else if pObj.Type == uint32(wlc.WLC_OBJ_CONST) {
			var pString string = Wln_ConstFromBits(wlc.Wlc_ObjConstValue(pObj), wlc.Wlc_ObjRange(pObj))
			Wln_ObjSetConst(pNew, iOutId, Abc_NamStrFindOrAdd(pNew.pManName, pString, nil))
			ABC_FREE(pString)
		}
		//        else if ( Type == ABC_OPER_BIT_MUX && Vec_IntSize(vFanins) == 3 )
		//            ABC_SWAP( int, Wln_ObjFanins(p, iObj)[1], Wln_ObjFanins(p, iObj)[2] );
		Wln_ObjSetNameId(pNew, iOutId, Wln_TrasformNameId(pNew, p, pObj))
		wlc.Wlc_ObjSetCopy(p, i, iOutId)
	}
	// inlining of macro wlc.Wlc_NtkForEachPo
	for i = 0; i < wlc.Wlc_NtkPoNum(p); i++ {
		pObj = wlc.Wlc_NtkPo(p, i)
		iOutId = Wln_ObjAlloc(pNew, ABC_OPER_CO, wlnBool(wlc.Wlc_ObjSigned(pObj)), wlc.Wlc_ObjEnd(pObj), wlc.Wlc_ObjBeg(pObj))
		Wln_ObjAddFanin(pNew, iOutId, wlc.Wlc_ObjCopy(p, wlc.Wlc_ObjId(p, pObj)))
		//Wln_ObjSetNameId( pNew, iOutId, Wln_TrasformNameId(pNew, p, pObj) );
	}
	assert(Vec_IntSize(vInits) == wlc.Wlc_NtkCoNum(p)-wlc.Wlc_NtkPoNum(p), "Vec_IntSize(vInits) == wlc.Wlc_NtkCoNum(p) - wlc.Wlc_NtkPoNum(p)")
	// inlining of macro wlc.Wlc_NtkForEachCo
	for i = 0; i < wlc.Wlc_NtkCoNum(p); i++ {
		pObj = wlc.Wlc_NtkCo(p, i)
		if i < wlc.Wlc_NtkPoNum(p) {
			continue
		}
		//char * pInNames[8] = {"d", "clk", "reset", "set", "enable", "async", "sre", "init"};
		Vec_IntClear(vFanins)
		Vec_IntPush(vFanins, wlc.Wlc_ObjCopy(p, wlc.Wlc_ObjFaninId0(pObj)))
		for n = 0; n < 6; n++ {
			Vec_IntPush(vFanins, 0)
		}
		Vec_IntPush(vFanins, Vec_IntEntry(vInits, i-wlc.Wlc_NtkPoNum(p)))
		Wln_ObjAddFanins(pNew, Vec_IntEntry(&pNew.vFfs, i-wlc.Wlc_NtkPoNum(p)), vFanins)
	}
	Vec_IntFree(vFanins)
	Vec_IntFree(vInits)
	return pNew
}
func Wln_NtkFromWlcTest(p *wlc.Wlc_Ntk_t) {
	var pNew *Wln_Ntk_t = Wln_NtkFromWlc(p)
	Wln_WriteVer(pNew, "test_wlc2wln.v")
	Wln_NtkFree(pNew)
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
