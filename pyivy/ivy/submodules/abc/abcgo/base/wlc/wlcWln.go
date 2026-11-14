package wlc

import (
	"fmt"
	wln "github.com/glycerine/abcgo/base/wln"
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
//#include "base/wlc/h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Ndr_TypeWlc2Ndr( int Type );

func wlnBool(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

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
func Wlc_TransformNameId(pNew *wln.Wln_Ntk_t, p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
	return Abc_NamStrFindOrAdd(wln.Wln_NtkNameMan(pNew), Wlc_ObjName(p, Wlc_ObjId(p, pObj)), nil)
}
func Wlc_NtkToWln(p *Wlc_Ntk_t) *wln.Wln_Ntk_t {
	var pObj *Wlc_Obj_t
	var Buffer string // was [1000]byte in C

	var i, j, n, Type, iFanin, iOutId, iBit int32
	var vFanins *Vec_Int_t = Vec_IntAlloc(10)
	var vInits *Vec_Int_t = Vec_IntAlloc(Wlc_NtkFfNum(p))
	var pNew *wln.Wln_Ntk_t = wln.Wln_NtkAlloc(Wlc_NtkName(p), Wlc_NtkObjNum(p)+Wlc_NtkCoNum(p)+Wlc_NtkFfNum(p))
	wln.Wln_NtkSetNameMan(pNew, Abc_NamStart(Abc_NamObjNumMax(Wlc_NtkNameMan(p)), 10))
	if Wlc_NtkSpec(p) != "" {
		wln.Wln_NtkSetMetadata(pNew, Wlc_NtkSpec(p), Wlc_NtkSmtLib(p))
	}

	Wlc_NtkCleanCopy(p)
	wln.Wln_NtkCleanNameId(pNew)
	// add primary inputs
	// inlining of macro Wlc_NtkForEachPi
	for i = 0; i < Wlc_NtkPiNum(p); i++ {
		pObj = Wlc_NtkPi(p, i)
		iOutId = wln.Wln_ObjAlloc(pNew, wln.ABC_OPER_CI, wlnBool(Wlc_ObjSigned(pObj)), Wlc_ObjEnd(pObj), Wlc_ObjBeg(pObj))
		wln.Wln_ObjSetNameId(pNew, iOutId, Wlc_TransformNameId(pNew, p, pObj))
		Wlc_ObjSetCopy(p, Wlc_ObjId(p, pObj), iOutId)
	}
	// create initial state of the flops
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		assert(i == Wlc_ObjCiId(pObj), "i == Wlc_ObjCiId(pObj)")
		if pObj.Type == uint32(WLC_OBJ_PI) {
			continue
		}
		for j = 0; j < Wlc_ObjRange(pObj); j++ {
			if Wlc_NtkInits(p)[iBit+j] == 'x' {
				break
			}
		}
		// print flop init state
		if true {
			fmt.Printf("Flop %3d init state: %d'b", i-Wlc_NtkPiNum(p), Wlc_ObjRange(pObj))
			if j == Wlc_ObjRange(pObj) {
				var Count int32
				for n = 0; n < Wlc_ObjRange(pObj); n++ {
					Count += wlnBool(Wlc_NtkInits(p)[iBit+n] == '0')
				}
				if Count == Wlc_ObjRange(pObj) {
					fmt.Printf("0")
				} else {
					for n = 0; n < Wlc_ObjRange(pObj); n++ {
						fmt.Printf("%c", Wlc_NtkInits(p)[iBit+n])
					}
				}

			} else {
				var Count int32
				for n = 0; n < Wlc_ObjRange(pObj); n++ {
					Count += wlnBool(Wlc_NtkInits(p)[iBit+n] == 'x')
				}
				fmt.Printf("x")
				if Count != Wlc_ObjRange(pObj) {
					fmt.Printf(" (range %d)", Wlc_ObjRange(pObj))
				}
			}
			fmt.Printf("\n")
		}
		if j == Wlc_ObjRange(pObj) {
			Type = wln.ABC_OPER_CONST
		} else {
			Type = wln.ABC_OPER_CI
		}
		iOutId = wln.Wln_ObjAlloc(pNew, Type, wlnBool(Wlc_ObjSigned(pObj)), Wlc_ObjEnd(pObj), Wlc_ObjBeg(pObj))
		if j == Wlc_ObjRange(pObj) { // constant

			var pString string = wln.Wln_ConstFromStr(Wlc_NtkInits(p)[iBit:], Wlc_ObjRange(pObj))
			wln.Wln_ObjSetConst(pNew, iOutId, Abc_NamStrFindOrAdd(wln.Wln_NtkNameMan(pNew), pString, nil))
			pString = ""
		}
		Buffer = fmt.Sprintf("ff_init_%d", Vec_IntSize(vInits))
		wln.Wln_ObjSetNameId(pNew, iOutId, Abc_NamStrFindOrAdd(wln.Wln_NtkNameMan(pNew), Buffer, nil))
		Vec_IntPush(vInits, iOutId)
		iBit += Wlc_ObjRange(pObj)
	}
	assert(Wlc_NtkInits(p) == "" || iBit == int32(len(Wlc_NtkInits(p))), "initial bits match input length")
	// add flop outputs
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		assert(i == Wlc_ObjCiId(pObj), "i == Wlc_ObjCiId(pObj)")
		if pObj.Type == uint32(WLC_OBJ_PI) {
			continue
		}
		iOutId = wln.Wln_ObjAlloc(pNew, wln.ABC_OPER_DFFRSE, wlnBool(Wlc_ObjSigned(pObj)), Wlc_ObjEnd(pObj), Wlc_ObjBeg(pObj))
		wln.Wln_ObjSetNameId(pNew, iOutId, Wlc_TransformNameId(pNew, p, pObj))
		Wlc_ObjSetCopy(p, Wlc_ObjId(p, pObj), iOutId)
	}
	// add internal nodes
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjIsCi(pObj) != 0 || pObj.Type == 0 {
			continue
		}
		iOutId = wln.Wln_ObjAlloc(pNew, Ndr_TypeWlc2Ndr(int32(pObj.Type)), wlnBool(Wlc_ObjSigned(pObj)), Wlc_ObjEnd(pObj), Wlc_ObjBeg(pObj))
		Vec_IntClear(vFanins)
		// inlining of macro Wlc_ObjForEachFanin
		for n = 0; n < Wlc_ObjFaninNum(pObj); n++ {
			iFanin = Wlc_ObjFaninId(pObj, n)
			Vec_IntPush(vFanins, Wlc_ObjCopy(p, iFanin))
		}
		wln.Wln_ObjAddFanins(pNew, iOutId, vFanins)
		if pObj.Type == uint32(WLC_OBJ_BIT_SELECT) {
			wln.Wln_ObjSetSlice(pNew, iOutId, Hash_Int2ManInsert(wln.Wln_NtkRangeMan(pNew), Wlc_ObjEnd(pObj), Wlc_ObjBeg(pObj), 0))
		} else if pObj.Type == uint32(WLC_OBJ_CONST) {
			var pString string = wln.Wln_ConstFromBits(ptrSlice(Wlc_ObjConstValue(pObj), (Wlc_ObjRange(pObj)+31)/32), Wlc_ObjRange(pObj))
			wln.Wln_ObjSetConst(pNew, iOutId, Abc_NamStrFindOrAdd(wln.Wln_NtkNameMan(pNew), pString, nil))
			pString = ""
		}
		//        else if ( Type == ABC_OPER_BIT_MUX && Vec_IntSize(vFanins) == 3 )
		//            ABC_SWAP( int, wln.Wln_ObjFanins(p, iObj)[1], wln.Wln_ObjFanins(p, iObj)[2] );
		wln.Wln_ObjSetNameId(pNew, iOutId, Wlc_TransformNameId(pNew, p, pObj))
		Wlc_ObjSetCopy(p, i, iOutId)
	}
	// inlining of macro Wlc_NtkForEachPo
	for i = 0; i < Wlc_NtkPoNum(p); i++ {
		pObj = Wlc_NtkPo(p, i)
		iOutId = wln.Wln_ObjAlloc(pNew, wln.ABC_OPER_CO, wlnBool(Wlc_ObjSigned(pObj)), Wlc_ObjEnd(pObj), Wlc_ObjBeg(pObj))
		wln.Wln_ObjAddFanin(pNew, iOutId, Wlc_ObjCopy(p, Wlc_ObjId(p, pObj)))
		//wln.Wln_ObjSetNameId( pNew, iOutId, Wlc_TransformNameId(pNew, p, pObj) );
	}
	assert(Vec_IntSize(vInits) == Wlc_NtkCoNum(p)-Wlc_NtkPoNum(p), "Vec_IntSize(vInits) == Wlc_NtkCoNum(p) - Wlc_NtkPoNum(p)")
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		if i < Wlc_NtkPoNum(p) {
			continue
		}
		//char * pInNames[8] = {"d", "clk", "reset", "set", "enable", "async", "sre", "init"};
		Vec_IntClear(vFanins)
		Vec_IntPush(vFanins, Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObj)))
		for n = 0; n < 6; n++ {
			Vec_IntPush(vFanins, 0)
		}
		Vec_IntPush(vFanins, Vec_IntEntry(vInits, i-Wlc_NtkPoNum(p)))
		wln.Wln_ObjAddFanins(pNew, Vec_IntEntry(wln.Wln_NtkFfs(pNew), i-Wlc_NtkPoNum(p)), vFanins)
	}
	Vec_IntFree(vFanins)
	Vec_IntFree(vInits)
	return pNew
}
func Wlc_NtkToWlnTest(p *Wlc_Ntk_t) {
	var pNew *wln.Wln_Ntk_t = Wlc_NtkToWln(p)
	wln.Wln_WriteVer(pNew, "test_wlc2wln.v")
	wln.Wln_NtkFree(pNew)
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
