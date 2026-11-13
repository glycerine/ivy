package wln

import "fmt"

/**CFile****************************************************************

  FileName    [wlnObj.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Word-level network.]

  Synopsis    [Object construction procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 23, 2018.]

  Revision    [$Id: wlnObj.c,v 1.00 2018/09/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wln.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Creating objects.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Wln_ObjName_Buffer string // was [100]byte in C

func Wln_ObjName(p *Wln_Ntk_t, iObj int32) string {

	if Wln_NtkHasNameId(p) != 0 && Wln_ObjNameId(p, iObj) != 0 {
		return Abc_NamStr(p.pManName, Wln_ObjNameId(p, iObj))
	}
	spStatic_Wln_ObjName_Buffer = fmt.Sprintf("n%d", iObj)
	return spStatic_Wln_ObjName_Buffer
}
func Wln_ObjConstString(p *Wln_Ntk_t, iObj int32) string {
	assert(Wln_ObjIsConst(p, iObj) != 0, "Wln_ObjIsConst(p, iObj)")
	return Abc_NamStr(p.pManName, Wln_ObjFanin0(p, iObj))
}
func Wln_ObjUpdateType(p *Wln_Ntk_t, iObj int32, Type int32) {
	assert(Wln_ObjIsNone(p, iObj) != 0, "Wln_ObjIsNone(p, iObj)")
	p.nObjs[Wln_ObjType(p, iObj)]--
	Vec_IntWriteEntry(&p.vTypes, iObj, Type)
	p.nObjs[Wln_ObjType(p, iObj)]++
}
func Wln_ObjSetConst(p *Wln_Ntk_t, iObj int32, NameId int32) {
	assert(Wln_ObjIsConst(p, iObj) != 0, "Wln_ObjIsConst(p, iObj)")
	Wln_ObjSetFanin(p, iObj, 0, NameId)
}
func Wln_ObjSetSlice(p *Wln_Ntk_t, iObj int32, SliceId int32) {
	assert(Wln_ObjIsSlice(p, iObj) != 0, "Wln_ObjIsSlice(p, iObj)")
	Wln_ObjSetFanin(p, iObj, 1, SliceId)
}
func Wln_ObjAddFanin(p *Wln_Ntk_t, iObj, i int32) {
	p.vFanins[iObj].appendFanin(i)
}
func Wln_ObjAddFanins(p *Wln_Ntk_t, iObj int32, vFanins *Vec_Int_t) int32 {
	for i := int32(0); i < Vec_IntSize(vFanins); i++ {
		Wln_ObjAddFanin(p, iObj, Vec_IntEntry(vFanins, i))
	}
	return iObj
}
func Wln_ObjAlloc(p *Wln_Ntk_t, Type, Signed, End, Beg int32) int32 {
	iObj := Vec_IntSize(&p.vTypes)
	if iObj == Vec_IntCap(&p.vTypes) {
		old := len(p.vFanins)
		p.vFanins = append(p.vFanins, make([]Wln_Vec_t, old)...)
		Vec_IntGrow(&p.vTypes, 2*iObj)
	}
	assert(iObj == Vec_StrSize(&p.vSigns), "iObj == Vec_StrSize(&p.vSigns)")
	assert(iObj == Vec_IntSize(&p.vRanges), "iObj == Vec_IntSize(&p.vRanges)")
	Vec_IntPush(&p.vTypes, Type)
	Vec_StrPush(&p.vSigns, byte(Signed))
	Vec_IntPush(&p.vRanges, Hash_Int2ManInsert(p.pRanges, End, Beg, 0))
	if Wln_ObjIsCi(p, iObj) != 0 {
		Wln_ObjSetFanin(p, iObj, 1, Vec_IntSize(&p.vCis))
		Vec_IntPush(&p.vCis, iObj)
	}
	if Wln_ObjIsCo(p, iObj) != 0 {
		Wln_ObjSetFanin(p, iObj, 1, Vec_IntSize(&p.vCos))
		Vec_IntPush(&p.vCos, iObj)
	}
	if Wln_ObjIsFf(p, iObj) != 0 {
		Vec_IntPush(&p.vFfs, iObj)
	}
	p.nObjs[Type]++
	return iObj
}
func Wln_ObjClone(pNew *Wln_Ntk_t, p *Wln_Ntk_t, iObj int32) int32 {
	return Wln_ObjAlloc(pNew, Wln_ObjType(p, iObj), Wln_ObjIsSigned(p, iObj), Wln_ObjRangeEnd(p, iObj), Wln_ObjRangeBeg(p, iObj))
}
func Wln_ObjCreateCo(p *Wln_Ntk_t, iFanin int32) int32 {
	var iCo int32 = Wln_ObjClone(p, p, iFanin)
	Wln_ObjUpdateType(p, iCo, ABC_OPER_CO)
	Wln_ObjAddFanin(p, iCo, iFanin)
	return iCo
}
func Wln_ObjPrint(p *Wln_Ntk_t, iObj int32) {
	fmt.Printf("Obj %6d : Type = %6s  Fanins = %d : ", iObj, Abc_OperName(Wln_ObjType(p, iObj)), Wln_ObjFaninNum(p, iObj))
	for _, iFanin := range Wln_ObjFanins(p, iObj)[:Wln_ObjFaninNum(p, iObj)] {
		if iFanin != 0 {
			fmt.Printf("%5d ", iFanin)
		}
	}
	fmt.Println()
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
