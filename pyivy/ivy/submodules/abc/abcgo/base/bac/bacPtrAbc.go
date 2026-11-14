package bac

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	abcapi "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

/**CFile****************************************************************

  FileName    [bacPtrAbc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Simple interface with external tools.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacPtrAbc.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "base/abc/abc.h"
//#include "map/mio/mio.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Node type conversions.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ptr_HopToType(pObj *Abc_Obj_t) string {
	var uTruth uint64
	uTruths6 := [3]uint64{0xAAAAAAAAAAAAAAAA, 0xCCCCCCCCCCCCCCCC, 0xF0F0F0F0F0F0F0F0}
	assert(abcapi.Abc_ObjIsNode(pObj) != 0, "abcapi.Abc_ObjIsNode(pObj)")
	man := abcapi.Abc_ObjNtk(pObj).FunctionManager.(*hop.Hop_Man_t)
	uTruth = hop.Hop_ManComputeTruth6(man, pObj.Data.(*hop.Hop_Obj_t), abcapi.Abc_ObjFaninNum(pObj))
	/*
	   if ( uTruth ==  0 )                           return "BAC_BOX_C0";
	   if ( uTruth == ~(word)0 )                     return "BAC_BOX_C1";
	   if ( uTruth ==  uTruths6[0] )                 return "BAC_BOX_BUF";
	   if ( uTruth == ~uTruths6[0] )                 return "BAC_BOX_INV";
	   if ( uTruth == (uTruths6[0] & uTruths6[1]) )  return "BAC_BOX_AND";
	   if ( uTruth ==~(uTruths6[0] & uTruths6[1]) )  return "BAC_BOX_NAND";
	   if ( uTruth == (uTruths6[0] | uTruths6[1]) )  return "BAC_BOX_OR";
	   if ( uTruth ==~(uTruths6[0] | uTruths6[1]) )  return "BAC_BOX_NOR";
	   if ( uTruth == (uTruths6[0] ^ uTruths6[1]) )  return "BAC_BOX_XOR";
	   if ( uTruth ==~(uTruths6[0] ^ uTruths6[1]) )  return "BAC_BOX_XNOR";
	*/
	if uTruth == 0 {
		return "Const0T"
	}
	if uTruth == ^uint64(0) {
		return "Const1T"
	}
	if uTruth == uTruths6[0] {
		return "BufT"
	}
	if uTruth == ^uTruths6[0] {
		return "InvT"
	}
	if uTruth == (uTruths6[0] & uTruths6[1]) {
		return "AndT"
	}
	if uTruth == ^(uTruths6[0] & uTruths6[1]) {
		return "NandT"
	}
	if uTruth == (uTruths6[0] | uTruths6[1]) {
		return "OrT"
	}
	if uTruth == ^(uTruths6[0] | uTruths6[1]) {
		return "NorT"
	}
	if uTruth == (uTruths6[0] ^ uTruths6[1]) {
		return "XorT"
	}
	if uTruth == ^(uTruths6[0] ^ uTruths6[1]) {
		return "XnorT"
	}
	assert(false, "0")
	return ""
}

/*
*Function*************************************************************

	Synopsis    [Create Ptr from Abc_Ntk_t.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ptr_AbcObjName(pObj *Abc_Obj_t) string {
	if abcapi.Abc_ObjIsNet(pObj) != 0 || abcapi.Abc_ObjIsBox(pObj) != 0 {
		return abcapi.Abc_ObjName(pObj)
	}
	if abcapi.Abc_ObjIsCi(pObj) != 0 || abcapi.Abc_ObjIsNode(pObj) != 0 {
		return Ptr_AbcObjName(abcapi.Abc_ObjFanout0(pObj))
	}
	if abcapi.Abc_ObjIsCo(pObj) != 0 {
		return Ptr_AbcObjName(abcapi.Abc_ObjFanin0(pObj))
	}
	assert(false, "0")
	return ""
}
func Ptr_CheckArray(vArray *vec.Vec_Ptr_t) int32 {
	assert(vec.Vec_PtrSize(vArray) == vec.Vec_PtrCap(vArray), "vec.Vec_PtrSize(vArray) == vec.Vec_PtrCap(vArray)")
	return 1
}
func Ptr_AbcDeriveNode(pObj *Abc_Obj_t) *vec.Vec_Ptr_t {
	var vNode *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(int(2 + 2*(1+abcapi.Abc_ObjFaninNum(pObj))))
	assert(abcapi.Abc_ObjIsNode(pObj) != 0, "abcapi.Abc_ObjIsNode(pObj)")
	ntk := abcapi.Abc_ObjNtk(pObj)
	if abcapi.Abc_NtkHasAig(ntk) != 0 {
		vec.Vec_PtrPush(vNode, Ptr_HopToType(pObj))
	} else if abcapi.Abc_NtkHasSop(ntk) != 0 {
		vec.Vec_PtrPush(vNode, Ptr_SopToTypeName(pObj.Data.(string)))
	} else {
		assert(false, "0")
	}
	vec.Vec_PtrPush(vNode, Ptr_AbcObjName(pObj))
	assert(abcapi.Abc_ObjFaninNum(pObj) <= 2, "abcapi.Abc_ObjFaninNum(pObj) <= 2")
	for i := int32(0); i < abcapi.Abc_ObjFaninNum(pObj); i++ {
		pFanin := abcapi.Abc_ObjFanin(pObj, i)
		pin := "l"
		if i != 0 {
			pin = "r"
		}
		vec.Vec_PtrPush(vNode, pin)
		vec.Vec_PtrPush(vNode, Ptr_AbcObjName(pFanin))
	}
	vec.Vec_PtrPush(vNode, "o")
	vec.Vec_PtrPush(vNode, Ptr_AbcObjName(pObj))
	assert((Ptr_CheckArray(vNode) != 0), "Ptr_CheckArray(vNode)")
	return vNode
}
func Ptr_AbcDeriveBox(pObj *Abc_Obj_t) *vec.Vec_Ptr_t {
	pNtk := abcapi.Abc_ObjModel(pObj)
	vBox := vec.Vec_PtrAllocExact(int(2 + 2*abcapi.Abc_ObjFaninNum(pObj) + 2*abcapi.Abc_ObjFanoutNum(pObj)))
	assert(abcapi.Abc_ObjIsBox(pObj) != 0, "abcapi.Abc_ObjIsBox(pObj)")
	vec.Vec_PtrPush(vBox, abcapi.Abc_NtkName(pNtk))
	vec.Vec_PtrPush(vBox, Ptr_AbcObjName(pObj))
	for i := int32(0); i < abcapi.Abc_ObjFaninNum(pObj); i++ {
		pNext := abcapi.Abc_ObjFanin(pObj, i)
		vec.Vec_PtrPush(vBox, Ptr_AbcObjName(abcapi.Abc_NtkPi(pNtk, i)))
		vec.Vec_PtrPush(vBox, Ptr_AbcObjName(pNext))
	}
	for i := int32(0); i < abcapi.Abc_ObjFanoutNum(pObj); i++ {
		pNext := abcapi.Abc_ObjFanout(pObj, i)
		vec.Vec_PtrPush(vBox, Ptr_AbcObjName(abcapi.Abc_NtkPo(pNtk, i)))
		vec.Vec_PtrPush(vBox, Ptr_AbcObjName(pNext))
	}
	assert((Ptr_CheckArray(vBox) != 0), "Ptr_CheckArray(vBox)")
	return vBox
}
func Ptr_AbcDeriveBoxes(pNtk *Abc_Ntk_t) *vec.Vec_Ptr_t {
	vBoxes := vec.Vec_PtrAllocExact(int(abcapi.Abc_NtkBoxNum(pNtk) + abcapi.Abc_NtkNodeNum(pNtk)))
	for _, pObj := range abcapi.Abc_NtkBoxes(pNtk) {
		vec.Vec_PtrPush(vBoxes, Ptr_AbcDeriveBox(pObj))
	}
	for _, pObj := range abcapi.Abc_NtkNodes(pNtk) {
		vec.Vec_PtrPush(vBoxes, Ptr_AbcDeriveNode(pObj))
	}
	assert((Ptr_CheckArray(vBoxes) != 0), "Ptr_CheckArray(vBoxes)")
	return vBoxes
}

func Ptr_AbcDeriveInputs(pNtk *Abc_Ntk_t) *vec.Vec_Ptr_t {
	vSigs := vec.Vec_PtrAllocExact(int(abcapi.Abc_NtkPiNum(pNtk)))
	for i := int32(0); i < abcapi.Abc_NtkPiNum(pNtk); i++ {
		vec.Vec_PtrPush(vSigs, Ptr_AbcObjName(abcapi.Abc_NtkPi(pNtk, i)))
	}
	assert((Ptr_CheckArray(vSigs) != 0), "Ptr_CheckArray(vSigs)")
	return vSigs
}
func Ptr_AbcDeriveOutputs(pNtk *Abc_Ntk_t) *vec.Vec_Ptr_t {
	vSigs := vec.Vec_PtrAllocExact(int(abcapi.Abc_NtkPoNum(pNtk)))
	for i := int32(0); i < abcapi.Abc_NtkPoNum(pNtk); i++ {
		vec.Vec_PtrPush(vSigs, Ptr_AbcObjName(abcapi.Abc_NtkPo(pNtk, i)))
	}
	assert((Ptr_CheckArray(vSigs) != 0), "Ptr_CheckArray(vSigs)")
	return vSigs
}
func Ptr_AbcDeriveNtk(pNtk *Abc_Ntk_t) *vec.Vec_Ptr_t {
	var vNtk *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(5)
	vec.Vec_PtrPush(vNtk, abcapi.Abc_NtkName(pNtk))
	vec.Vec_PtrPush(vNtk, Ptr_AbcDeriveInputs(pNtk))
	vec.Vec_PtrPush(vNtk, Ptr_AbcDeriveOutputs(pNtk))
	vec.Vec_PtrPush(vNtk, vec.Vec_PtrAllocExact(0))
	vec.Vec_PtrPush(vNtk, Ptr_AbcDeriveBoxes(pNtk))
	assert((Ptr_CheckArray(vNtk) != 0), "Ptr_CheckArray(vNtk)")
	return vNtk
}
func Ptr_AbcDeriveDes(pNtk *Abc_Ntk_t) *vec.Vec_Ptr_t {
	var vDes *vec.Vec_Ptr_t
	design := pNtk.Design
	vDes = vec.Vec_PtrAllocExact(1 + len(design.Modules))
	vec.Vec_PtrPush(vDes, design.Name)
	for _, pTemp := range design.Modules {
		vec.Vec_PtrPush(vDes, Ptr_AbcDeriveNtk(pTemp))
	}
	assert((Ptr_CheckArray(vDes) != 0), "Ptr_CheckArray(vDes)")
	return vDes
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ptr_ManExperiment(pNtk *Abc_Ntk_t) {
	var clk int64 = util.Abc_Clock()
	var pFileName string = Extra_FileNameGenericAppend(pNtk.Design.Name, "_out.blif")
	var vDes *vec.Vec_Ptr_t = Ptr_AbcDeriveDes(pNtk)
	fmt.Printf("Converting to Ptr:  Memory = %6.3f MB  ", float64(Bac_PtrMemory(vDes))/(1<<20))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	Bac_PtrDumpBlif(pFileName, vDes)
	fmt.Printf("Finished writing output file \"%s\".  ", pFileName)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	Bac_PtrFree(vDes)
}

/*
*Function*************************************************************

	Synopsis    [Create Bac_Man_t from tech-ind Ptr.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ptr_NameToType(pSop string) int32 {
	if !(int32(strings.Compare(pSop, "Const0T")) != 0) {
		return int32(BAC_BOX_CF)
	}
	if !(int32(strings.Compare(pSop, "Const1T")) != 0) {
		return int32(BAC_BOX_CT)
	}
	if !(int32(strings.Compare(pSop, "BufT")) != 0) {
		return int32(BAC_BOX_BUF)
	}
	if !(int32(strings.Compare(pSop, "InvT")) != 0) {
		return int32(BAC_BOX_INV)
	}
	if !(int32(strings.Compare(pSop, "AndT")) != 0) {
		return int32(BAC_BOX_AND)
	}
	if !(int32(strings.Compare(pSop, "NandT")) != 0) {
		return int32(BAC_BOX_NAND)
	}
	if !(int32(strings.Compare(pSop, "OrT")) != 0) {
		return int32(BAC_BOX_OR)
	}
	if !(int32(strings.Compare(pSop, "NorT")) != 0) {
		return int32(BAC_BOX_NOR)
	}
	if !(int32(strings.Compare(pSop, "XorT")) != 0) {
		return int32(BAC_BOX_XOR)
	}
	if !(int32(strings.Compare(pSop, "XnorT")) != 0) {
		return int32(BAC_BOX_XNOR)
	}
	return int32(BAC_OBJ_BOX)
}
func Ptr_ManCountNtk(vNtk *vec.Vec_Ptr_t) int32 {
	var vInputs *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t)
	var vOutputs *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t)
	var vNodes *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 3).(*vec.Vec_Ptr_t)
	var vBoxes *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 4).(*vec.Vec_Ptr_t)
	var vBox *vec.Vec_Ptr_t
	var i int
	var Counter int32
	assert(vec.Vec_PtrSize(vNodes) == 0, "vec.Vec_PtrSize(vNodes) == 0")
	Counter += int32(vec.Vec_PtrSize(vInputs))
	Counter += int32(vec.Vec_PtrSize(vOutputs))
	for i = 0; i < vec.Vec_PtrSize(vBoxes); i++ {
		vBox = vec.Vec_PtrEntry(vBoxes, i).(*vec.Vec_Ptr_t)
		Counter += int32(vec.Vec_PtrSize(vBox) / 2)
	}
	return Counter
}
func Bac_BoxCountOutputs(pNtk *Bac_Ntk_t, pBoxNtk string) int32 {
	var ModuleId int32 = Bac_ManNtkFindId(pNtk.pDesign, pBoxNtk)
	if ModuleId == 0 {
		return 1
	}
	return Bac_NtkPoNumAlloc(Bac_ManNtk(pNtk.pDesign, ModuleId))
}
func Bac_NtkDeriveFromPtr(pNtk *Bac_Ntk_t, vNtk *vec.Vec_Ptr_t, vMap *vec.Vec_Int_t, vBox2Id *vec.Vec_Int_t) int32 {
	var pName string
	var pModuleName string = vec.Vec_PtrEntry(vNtk, 0).(string)
	var vInputs *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t)
	var vOutputs *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t)
	var vBoxes *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 4).(*vec.Vec_Ptr_t)
	var vBox *vec.Vec_Ptr_t
	var i int
	var k, iObj, iTerm, NameId int32
	// start network with the given name
	NameId = Abc_NamStrFindOrAdd(pNtk.pDesign.pStrs, pModuleName, nil)
	assert(Bac_NtkNameId(pNtk) == NameId, "Bac_NtkNameId(pNtk) == NameId")
	// map driven NameIds into their ObjIds for PIs
	for i = 0; i < vec.Vec_PtrSize(vInputs); i++ {
		pName = vec.Vec_PtrEntry(vInputs, i).(string)
		NameId = Abc_NamStrFindOrAdd(pNtk.pDesign.pStrs, pName, nil)
		if vec.Vec_IntGetEntryFull(vMap, int(NameId)) != -1 {
			fmt.Printf("PI with name \"%s\" is not unique module \"%s\".\n", pName, pModuleName)
			return 0
		}
		iObj = Bac_ObjAlloc(pNtk, BAC_OBJ_PI, -1)
		Bac_ObjSetName(pNtk, iObj, Abc_Var2Lit2(NameId, int32(BAC_NAME_BIN)))
		vec.Vec_IntSetEntryFull(vMap, int(NameId), iObj)
		Bac_NtkAddInfo(pNtk, Abc_Var2Lit2(NameId, 1), -1, -1)
	}
	// map driven NameIds into their ObjIds for BOs
	vec.Vec_IntClear(vBox2Id)
	for i = 0; i < vec.Vec_PtrSize(vBoxes); i++ {
		vBox = vec.Vec_PtrEntry(vBoxes, i).(*vec.Vec_Ptr_t)
		var pBoxNtk string = vec.Vec_PtrEntry(vBox, 0).(string)
		var pBoxName string = vec.Vec_PtrEntry(vBox, 1).(string)
		var nOutputs int32 = Bac_BoxCountOutputs(pNtk, pBoxNtk)
		var nInputs int32 = int32(vec.Vec_PtrSize(vBox))/2 - nOutputs - 1
		var NtkId int32 = Bac_ManNtkFindId(pNtk.pDesign, pBoxNtk)
		assert(vec.Vec_PtrSize(vBox)%2 == 0, "vec.Vec_PtrSize(vBox) % 2 == 0")
		assert(nOutputs > 0 && 2*(nOutputs+1) <= int32(vec.Vec_PtrSize(vBox)), "nOutputs > 0 && 2*(nOutputs + 1) <= vec.Vec_PtrSize(vBox)")
		iObj = Bac_BoxAlloc(pNtk, Bac_ObjType_t(Ptr_NameToType(pBoxNtk)), nInputs, nOutputs, NtkId)
		if NtkId > 0 {
			Bac_NtkSetHost(Bac_ManNtk(pNtk.pDesign, NtkId), Bac_NtkId(pNtk), iObj)
		}
		Bac_ObjSetName(pNtk, iObj, Abc_Var2Lit2(Abc_NamStrFindOrAdd(pNtk.pDesign.pStrs, pBoxName, nil), int32(BAC_NAME_BIN)))
		for k = 0; k < Bac_BoxBoNum(pNtk, iObj); k++ {
			iTerm = Bac_BoxBo(pNtk, iObj, k)
			pName = vec.Vec_PtrEntry(vBox, vec.Vec_PtrSize(vBox)-2*int(nOutputs-k)+1).(string)
			NameId = Abc_NamStrFindOrAdd(pNtk.pDesign.pStrs, pName, nil)
			if vec.Vec_IntGetEntryFull(vMap, int(NameId)) != -1 {
				fmt.Printf("Signal \"%s\" has multiple drivers in module \"%s\".\n", pName, pModuleName)
				return 0
			}
			Bac_ObjSetName(pNtk, iTerm, Abc_Var2Lit2(NameId, int32(BAC_NAME_BIN)))
			vec.Vec_IntSetEntryFull(vMap, int(NameId), iTerm)
		}
		vec.Vec_IntPush(vBox2Id, iObj)
	}
	assert(vec.Vec_IntSize(vBox2Id) == vec.Vec_PtrSize(vBoxes), "vec.Vec_IntSize(vBox2Id) == vec.Vec_PtrSize(vBoxes)")
	// connect BIs
	for i = 0; i < vec.Vec_PtrSize(vBoxes); i++ {
		vBox = vec.Vec_PtrEntry(vBoxes, i).(*vec.Vec_Ptr_t)
		iObj = vec.Vec_IntEntry(vBox2Id, i)
		for k = 0; k < Bac_BoxBiNum(pNtk, iObj); k++ {
			iTerm = Bac_BoxBi(pNtk, iObj, k)
			pName = vec.Vec_PtrEntry(vBox, 2*int(k+1)+1).(string)
			NameId = Abc_NamStrFindOrAdd(pNtk.pDesign.pStrs, pName, nil)
			if vec.Vec_IntGetEntryFull(vMap, int(NameId)) == -1 {
				fmt.Printf("Signal \"%s\" in not driven in module \"%s\".\n", pName, pModuleName)
			}
			Bac_ObjSetFanin(pNtk, iTerm, vec.Vec_IntGetEntryFull(vMap, int(NameId)))
		}
	}
	// connect POs
	for i = 0; i < vec.Vec_PtrSize(vOutputs); i++ {
		pName = vec.Vec_PtrEntry(vOutputs, i).(string)
		NameId = Abc_NamStrFindOrAdd(pNtk.pDesign.pStrs, pName, nil)
		if vec.Vec_IntGetEntryFull(vMap, int(NameId)) == -1 {
			fmt.Printf("PO with name \"%s\" in not driven in module \"%s\".\n", pName, pModuleName)
		}
		iObj = Bac_ObjAlloc(pNtk, BAC_OBJ_PO, vec.Vec_IntGetEntryFull(vMap, int(NameId)))
		Bac_NtkAddInfo(pNtk, Abc_Var2Lit2(NameId, 2), -1, -1)
	}
	// update map
	for iObj = 0; iObj < Bac_NtkObjNum(pNtk); iObj++ {
		if Bac_ObjIsCi(pNtk, iObj) != 0 {
			vec.Vec_IntSetEntryFull(vMap, int(Bac_ObjNameId(pNtk, iObj)), -1)
		}
	}
	// double check
	for i := 0; i < vec.Vec_IntSize(vMap); i++ {
		assert(vec.Vec_IntEntry(vMap, i) == -1, "iObj == -1")
	}
	assert(Bac_NtkObjNum(pNtk) == int32(vec.Vec_StrCap(&pNtk.vType)), "Bac_NtkObjNum(pNtk) == vec.Vec_StrCap(&pNtk.vType)")
	return 1
}
func Bac_PtrTransformToCba(vDes *vec.Vec_Ptr_t) *Bac_Man_t {
	var pName string = vec.Vec_PtrEntry(vDes, 0).(string)
	var pNew *Bac_Man_t = Bac_ManAlloc(pName, int32(vec.Vec_PtrSize(vDes)-1))
	var vMap *vec.Vec_Int_t = vec.Vec_IntStartFull(1000)
	var vBox2Id *vec.Vec_Int_t = vec.Vec_IntAlloc(1000)
	// create interfaces
	var pNtk *Bac_Ntk_t
	var i int32
	for i = 1; i <= Bac_ManNtkNum(pNew); i++ {
		pNtk = Bac_ManNtk(pNew, i)
		var vNtk *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vDes, int(i)).(*vec.Vec_Ptr_t)
		var vInputs *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 1).(*vec.Vec_Ptr_t)
		var vOutputs *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vNtk, 2).(*vec.Vec_Ptr_t)
		var NameId int32 = Abc_NamStrFindOrAdd(pNew.pStrs, vec.Vec_PtrEntry(vNtk, 0).(string), nil)
		Bac_NtkAlloc(pNtk, NameId, int32(vec.Vec_PtrSize(vInputs)), int32(vec.Vec_PtrSize(vOutputs)), Ptr_ManCountNtk(vNtk))
		Bac_NtkStartNames(pNtk)
	}
	// parse the networks
	for i = 1; i <= Bac_ManNtkNum(pNew); i++ {
		pNtk = Bac_ManNtk(pNew, i)
		var vNtk *vec.Vec_Ptr_t = vec.Vec_PtrEntry(vDes, int(i)).(*vec.Vec_Ptr_t)
		if !(Bac_NtkDeriveFromPtr(pNtk, vNtk, vMap, vBox2Id) != 0) {
			break
		}
	}
	if i <= Bac_ManNtkNum(pNew) {
		Bac_ManFree(pNew)
		pNew = nil
	}
	vec.Vec_IntFree(vBox2Id)
	vec.Vec_IntFree(vMap)
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Create Ptr from mapped Bac_Man_t.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_NtkTransformToPtrBox(p *Bac_Ntk_t, iBox int32) *vec.Vec_Ptr_t {
	var i int32
	var iTerm int32
	var fUser int32 = Bac_ObjIsBoxUser(p, iBox)
	var pBoxNtk *Bac_Ntk_t = Bac_BoxNtk(p, iBox)
	pLib, _ := p.pDesign.pMioLib.(*Mio_Library_t)
	var pGate *Mio_Gate_t
	if pLib != nil {
		pGate = Mio_LibraryReadGateByName(pLib, Bac_BoxNtkName(p, iBox), "")
	} else {
		pGate = nil
	}
	var vBox *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(int(2 * Bac_BoxSize(p, iBox)))
	vec.Vec_PtrPush(vBox, Bac_BoxNtkName(p, iBox))
	vec.Vec_PtrPush(vBox, Bac_ObjNameStr(p, iBox))
	for i = 0; i < Bac_BoxBiNum(p, iBox); i++ {
		iTerm = Bac_BoxBi(p, iBox, i)
		pinName := Mio_GateReadPinName(pGate, i)
		if fUser != 0 {
			pinName = Bac_ObjNameStr(pBoxNtk, Bac_NtkPi(pBoxNtk, i))
		}
		vec.Vec_PtrPush(vBox, pinName)
		vec.Vec_PtrPush(vBox, Bac_ObjNameStr(p, iTerm))
	}
	for i = 0; i < Bac_BoxBoNum(p, iBox); i++ {
		iTerm = Bac_BoxBo(p, iBox, i)
		pinName := Mio_GateReadOutName(pGate)
		if fUser != 0 {
			pinName = Bac_ObjNameStr(pBoxNtk, Bac_NtkPo(pBoxNtk, i))
		}
		vec.Vec_PtrPush(vBox, pinName)
		vec.Vec_PtrPush(vBox, Bac_ObjNameStr(p, iTerm))
	}
	assert((Ptr_CheckArray(vBox) != 0), "Ptr_CheckArray(vBox)")
	return vBox
}
func Bac_NtkTransformToPtrBoxes(p *Bac_Ntk_t) *vec.Vec_Ptr_t {
	var vBoxes *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(int(Bac_NtkBoxNum(p)))
	for iBox := int32(0); iBox < Bac_NtkObjNum(p); iBox++ {
		if Bac_ObjIsBox(p, iBox) != 0 {
			vec.Vec_PtrPush(vBoxes, Bac_NtkTransformToPtrBox(p, iBox))
		}
	}
	assert((Ptr_CheckArray(vBoxes) != 0), "Ptr_CheckArray(vBoxes)")
	return vBoxes
}

func Bac_NtkTransformToPtrInputs(p *Bac_Ntk_t) *vec.Vec_Ptr_t {
	vSigs := vec.Vec_PtrAllocExact(int(Bac_NtkPiNum(p)))
	for i := int32(0); i < Bac_NtkPiNum(p); i++ {
		vec.Vec_PtrPush(vSigs, Bac_ObjNameStr(p, Bac_NtkPi(p, i)))
	}
	assert((Ptr_CheckArray(vSigs) != 0), "Ptr_CheckArray(vSigs)")
	return vSigs
}
func Bac_NtkTransformToPtrOutputs(p *Bac_Ntk_t) *vec.Vec_Ptr_t {
	vSigs := vec.Vec_PtrAllocExact(int(Bac_NtkPoNum(p)))
	for i := int32(0); i < Bac_NtkPoNum(p); i++ {
		vec.Vec_PtrPush(vSigs, Bac_ObjNameStr(p, Bac_NtkPo(p, i)))
	}
	assert((Ptr_CheckArray(vSigs) != 0), "Ptr_CheckArray(vSigs)")
	return vSigs
}
func Bac_NtkTransformToPtr(p *Bac_Ntk_t) *vec.Vec_Ptr_t {
	var vNtk *vec.Vec_Ptr_t = vec.Vec_PtrAllocExact(5)
	vec.Vec_PtrPush(vNtk, Bac_NtkName(p))
	vec.Vec_PtrPush(vNtk, Bac_NtkTransformToPtrInputs(p))
	vec.Vec_PtrPush(vNtk, Bac_NtkTransformToPtrOutputs(p))
	vec.Vec_PtrPush(vNtk, vec.Vec_PtrAllocExact(0))
	vec.Vec_PtrPush(vNtk, Bac_NtkTransformToPtrBoxes(p))
	assert((Ptr_CheckArray(vNtk) != 0), "Ptr_CheckArray(vNtk)")
	return vNtk
}
func Bac_PtrDeriveFromCba(p *Bac_Man_t) *vec.Vec_Ptr_t {
	var vDes *vec.Vec_Ptr_t
	if p == nil {
		return nil
	}
	if p.pMioLib == nil {
		fmt.Printf("Cannot transform CBA network into Ptr because it is not mapped.\n")
		return nil
	}
	Bac_ManAssignInternWordNames(p)
	vDes = vec.Vec_PtrAllocExact(int(1 + Bac_ManNtkNum(p)))
	vec.Vec_PtrPush(vDes, p.pName)
	for i := int32(1); i <= Bac_ManNtkNum(p); i++ {
		vec.Vec_PtrPush(vDes, Bac_NtkTransformToPtr(Bac_ManNtk(p, i)))
	}
	assert((Ptr_CheckArray(vDes) != 0), "Ptr_CheckArray(vDes)")
	return vDes
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
