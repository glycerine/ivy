package acb

import "fmt"

/**CFile****************************************************************

  FileName    [acbAbc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Bridge.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: acbAbc.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acb.h"
//#include "base/abc/abc.h"
//#include "aig/miniaig/ndr.h"
//#include "acbPar.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

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
func Acb_NtkFromAbc2(p *Abc_Ntk_t) *Acb_Ntk_t {
	var pMan *Acb_Man_t = Acb_ManAlloc(Abc_NtkSpec(p), 1, nil, nil, nil, nil)
	var i int32
	var k int32
	var NameId int32 = Abc_NamStrFindOrAdd(pMan.pStrs, Abc_NtkName(p), nil)
	var pNtk *Acb_Ntk_t = Acb_NtkAlloc(pMan, NameId, Abc_NtkCiNum(p), Abc_NtkCoNum(p), Abc_NtkObjNum(p))
	var pObj, pFanin *Abc_Obj_t
	assert(Abc_NtkIsSopLogic(p) != 0, "Abc_NtkIsSopLogic(p)")
	pNtk.nFaninMax = 6
	// inlining of macro Abc_NtkForEachCi
	for i = 0; i < Abc_NtkCiNum(p); i++ {
		pObj = Abc_NtkCi(p, i)
		pObj.Temp = int(Acb_ObjAlloc(pNtk, ABC_OPER_CI, 0, 0))
	}
	// inlining of macro Abc_NtkForEachNode
	for i = 0; i < Abc_NtkObjNumMax(p); i++ {
		pObj = Abc_NtkObj(p, i)
		if (pObj) == nil || Abc_ObjIsNode(pObj) == 0 {
		} else {
			pObj.Temp = int(Acb_ObjAlloc(pNtk, ABC_OPER_LUT, Abc_ObjFaninNum(pObj), 0))
		}
	}
	// inlining of macro Abc_NtkForEachCo
	for i = 0; i < Abc_NtkCoNum(p); i++ {
		pObj = Abc_NtkCo(p, i)
		pObj.Temp = int(Acb_ObjAlloc(pNtk, ABC_OPER_CO, 1, 0))
	}
	// inlining of macro Abc_NtkForEachNode
	for i = 0; i < Abc_NtkObjNumMax(p); i++ {
		pObj = Abc_NtkObj(p, i)
		if (pObj) == nil || Abc_ObjIsNode(pObj) == 0 {
		} else {
			// inlining of macro Abc_ObjForEachFanin
			for k = 0; k < Abc_ObjFaninNum(pObj); k++ {
				pFanin = Abc_ObjFanin(pObj, k)
				Acb_ObjAddFanin(pNtk, int32(pObj.Temp), int32(pFanin.Temp))
			}
		}
	}
	// inlining of macro Abc_NtkForEachCo
	for i = 0; i < Abc_NtkCoNum(p); i++ {
		pObj = Abc_NtkCo(p, i)
		Acb_ObjAddFanin(pNtk, int32(pObj.Temp), int32(Abc_ObjFanin(pObj, 0).Temp))
	}
	Acb_NtkCleanObjTruths(pNtk)
	// inlining of macro Abc_NtkForEachNode
	for i = 0; i < Abc_NtkObjNumMax(p); i++ {
		pObj = Abc_NtkObj(p, i)
		if (pObj) == nil || Abc_ObjIsNode(pObj) == 0 {
		} else {
			Acb_ObjSetTruth(pNtk, int32(pObj.Temp), Abc_SopToTruth(pObj.Data.(string), Abc_ObjFaninNum(pObj)))
		}
	}
	Acb_NtkSetRegNum(pNtk, Abc_NtkLatchNum(p))
	Acb_NtkAdd(pMan, pNtk)
	return pNtk
}
func Acb_NtkFromAbc(p *Abc_Ntk_t) *Acb_Ntk_t {
	var pMan *Acb_Man_t = Acb_ManAlloc(Abc_NtkSpec(p), 1, nil, nil, nil, nil)
	var i int32
	var k int32
	var NameId int32 = Abc_NamStrFindOrAdd(pMan.pStrs, Abc_NtkName(p), nil)
	var pNtk *Acb_Ntk_t = Acb_NtkAlloc(pMan, NameId, Abc_NtkCiNum(p), Abc_NtkCoNum(p), Abc_NtkObjNumMax(p)-1)
	var pObj, pFanin *Abc_Obj_t
	assert(Abc_NtkIsSopLogic(p) != 0, "Abc_NtkIsSopLogic(p)")
	pNtk.nFaninMax = 6
	for i = 1; i < Abc_NtkObjNumMax(p); i++ {
		pObj = Abc_NtkObj(p, i)
		if pObj == nil {
			Acb_ObjAlloc(pNtk, ABC_OPER_NONE, 0, 0)
		} else if Abc_ObjIsCi(pObj) != 0 {
			pObj.Temp = int(Acb_ObjAlloc(pNtk, ABC_OPER_CI, 0, 0))
		} else if Abc_ObjIsCo(pObj) != 0 {
			pObj.Temp = int(Acb_ObjAlloc(pNtk, ABC_OPER_CO, 1, 0))
		} else if Abc_ObjIsNode(pObj) != 0 {
			pObj.Temp = int(Acb_ObjAlloc(pNtk, ABC_OPER_LUT, Abc_ObjFaninNum(pObj), 0))
		} else {
			assert(false, "0")
		}
		assert(pObj == nil || int32(pObj.Temp) == int32(Abc_ObjId(pObj)), "pObj == nil || int32(pObj.Temp) == int32(Abc_ObjId(pObj))")
	}
	// inlining of macro Abc_NtkForEachNode
	for i = 0; i < Abc_NtkObjNumMax(p); i++ {
		pObj = Abc_NtkObj(p, i)
		if (pObj) == nil || Abc_ObjIsNode(pObj) == 0 {
		} else {
			// inlining of macro Abc_ObjForEachFanin
			for k = 0; k < Abc_ObjFaninNum(pObj); k++ {
				pFanin = Abc_ObjFanin(pObj, k)
				Acb_ObjAddFanin(pNtk, int32(pObj.Temp), int32(pFanin.Temp))
			}
		}
	}
	// inlining of macro Abc_NtkForEachCo
	for i = 0; i < Abc_NtkCoNum(p); i++ {
		pObj = Abc_NtkCo(p, i)
		Acb_ObjAddFanin(pNtk, int32(pObj.Temp), int32(Abc_ObjFanin(pObj, 0).Temp))
	}
	Acb_NtkCleanObjTruths(pNtk)
	// inlining of macro Abc_NtkForEachNode
	for i = 0; i < Abc_NtkObjNumMax(p); i++ {
		pObj = Abc_NtkObj(p, i)
		if (pObj) == nil || Abc_ObjIsNode(pObj) == 0 {
		} else {
			Acb_ObjSetTruth(pNtk, int32(pObj.Temp), Abc_SopToTruth(pObj.Data.(string), Abc_ObjFaninNum(pObj)))
		}
	}
	Acb_NtkSetRegNum(pNtk, Abc_NtkLatchNum(p))
	Acb_NtkAdd(pMan, pNtk)
	return pNtk
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkToAbc(pNtk *Abc_Ntk_t, p *Acb_Ntk_t) *Abc_Ntk_t {
	var i, k, iObj, iFanin int32
	var pNtkNew *Abc_Ntk_t = Abc_NtkStartFrom(pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP)
	var pMan *Mem_Flex_t = Abc_NtkFuncManager(pNtkNew)
	var vCover *Vec_Int_t = Vec_IntAlloc(1000)
	Acb_NtkCleanObjCopies(p)
	// inlining of macro Acb_NtkForEachCi
	for i = 0; i < Acb_NtkCiNum(p); i++ {
		iObj = Acb_NtkCi(p, i)
		Acb_ObjSetCopy(p, iObj, Abc_ObjId(Abc_NtkCi(pNtkNew, i)))
	}
	// inlining of macro Acb_NtkForEachNode
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 {
		} else {
			var pObjNew *Abc_Obj_t = Abc_NtkCreateNode(pNtkNew)
			pObjNew.Data = Abc_SopCreateFromTruthIsop(pMan, Acb_ObjFaninNum(p, iObj), truthWords(Acb_ObjTruthP(p, iObj), Acb_ObjFaninNum(p, iObj)), vCover)
			Acb_ObjSetCopy(p, iObj, Abc_ObjId(pObjNew))
		}
	}
	Vec_IntFree(vCover)
	// inlining of macro Acb_NtkForEachNode
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 {
		} else {
			var pObjNew *Abc_Obj_t = Abc_NtkObj(pNtkNew, Acb_ObjCopy(p, iObj))
			// inlining of macro Acb_ObjForEachFanin
			for k = 0; k < Acb_ObjFaninNum(p, iObj); k++ {
				iFanin = Acb_ObjFanin(p, iObj, k)
				Abc_ObjAddFanin(pObjNew, Abc_NtkObj(pNtkNew, Acb_ObjCopy(p, iFanin)))
			}
			assert(Abc_SopGetVarNum(pObjNew.Data.(string)) == Abc_ObjFaninNum(pObjNew), "Abc_SopGetVarNum(pObjNew.Data.(string)) == Abc_ObjFaninNum(pObjNew)")
		}
	}
	// inlining of macro Acb_NtkForEachCoDriver
	for i = 0; i < Acb_NtkCoNum(p); i++ {
		iFanin = Acb_ObjFanin(p, Acb_NtkCo(p, i), 0)
		Abc_ObjAddFanin(Abc_NtkCo(pNtkNew, i), Abc_NtkObj(pNtkNew, Acb_ObjCopy(p, iFanin)))
	}
	if Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Printf("Acb_NtkToAbc: The network check has failed.\n")
		Abc_NtkDelete(pNtkNew)
		return nil
	}
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkFromNdr(pFileName string, pModule any, pNames *Abc_Nam_t, vWeights *Vec_Int_t, nNameIdMax int32) *Acb_Ntk_t {
	var p *Ndr_Data_t = pModule.(*Ndr_Data_t)
	var pMan *Acb_Man_t = Acb_ManAlloc(pFileName, 1, Abc_NamRef(pNames), nil, nil, nil)
	var k int32
	var NameId int32 = Abc_NamStrFindOrAdd(pMan.pStrs, pMan.pName, nil)
	var Mod int32 = 2
	var Obj int32
	var Type int32
	var nArray int32
	var pArray []int32
	var ObjId int32
	var Token0 int32 = Abc_NamStrFind(pMan.pStrs, "1'b0")
	var pNtk *Acb_Ntk_t = Acb_NtkAlloc(pMan, NameId, Ndr_DataCiNum(p, Mod), Ndr_DataCoNum(p, Mod), Ndr_DataObjNum(p, Mod))
	var vMap *Vec_Int_t = Vec_IntStart(nNameIdMax)
	Acb_NtkCleanObjWeights(pNtk)
	Acb_NtkCleanObjNames(pNtk)
	// inlining of macro Ndr_ModForEachPi
	// inlining of macro Ndr_ModForEachObj
	for Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj) {
		if Ndr_DataType(p, Obj) != NDR_OBJECT {
		} else if !Ndr_ObjIsType(p, Obj, ABC_OPER_CI) {
		} else {
			NameId = Ndr_ObjReadBody(p, Obj, NDR_OUTPUT)
			ObjId = Acb_ObjAlloc(pNtk, ABC_OPER_CI, 0, 0)
			Vec_IntWriteEntry(vMap, NameId, ObjId)
			Acb_ObjSetName(pNtk, ObjId, NameId)
			weight := int32(1)
			if vWeights != nil {
				weight = Vec_IntEntry(vWeights, NameId)
			}
			Acb_ObjSetWeight(pNtk, ObjId, weight)
		}
	}
	// inlining of macro Ndr_ModForEachTarget
	for Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj) {
		if Ndr_DataType(p, Obj) != NDR_TARGET {
		} else {
			NameId = Ndr_DataEntry(p, Obj)
			ObjId = Acb_ObjAlloc(pNtk, ABC_OPER_CONST_F, 0, 0)
			Vec_IntWriteEntry(vMap, NameId, ObjId)
			Acb_ObjSetName(pNtk, ObjId, NameId)
			Vec_IntPush(&pNtk.vTargets, ObjId)
		}
	}
	// inlining of macro Ndr_ModForEachNode
	// inlining of macro Ndr_ModForEachObj
	for Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj) {
		if Ndr_DataType(p, Obj) != NDR_OBJECT {
		} else if Ndr_ObjIsType(p, Obj, ABC_OPER_CI) || Ndr_ObjIsType(p, Obj, ABC_OPER_CO) {
		} else {
			NameId = Ndr_ObjReadBody(p, Obj, NDR_OUTPUT)
			nArray = Ndr_ObjReadArray(p, Obj, NDR_INPUT, &pArray)
			Type = Ndr_ObjReadBody(p, Obj, NDR_OPERTYPE)
			ObjId = Acb_ObjAlloc(pNtk, Acb_ObjType_t(Type), nArray, 0)
			Vec_IntWriteEntry(vMap, NameId, ObjId)
			Acb_ObjSetName(pNtk, ObjId, NameId)
		}
	}
	// inlining of macro Ndr_ModForEachNode
	// inlining of macro Ndr_ModForEachObj
	for Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj) {
		if Ndr_DataType(p, Obj) != NDR_OBJECT {
		} else if Ndr_ObjIsType(p, Obj, ABC_OPER_CI) || Ndr_ObjIsType(p, Obj, ABC_OPER_CO) {
		} else {
			var NameId int32 = Ndr_ObjReadBody(p, Obj, NDR_OUTPUT)
			var pName string = Abc_NamStr(pMan.pStrs, NameId)
			ObjId = Vec_IntEntry(vMap, NameId)
			nArray = Ndr_ObjReadArray(p, Obj, NDR_INPUT, &pArray)
			for k = 0; k < nArray; k++ {
				if Vec_IntEntry(vMap, pArray[k]) == 0 {
					fmt.Printf("Cannot find fanin %d of node \"%s\".\n", k, pName)
				}
				Acb_ObjAddFanin(pNtk, ObjId, Vec_IntEntry(vMap, pArray[k]))
			}
			weight := int32(1)
			if vWeights != nil {
				weight = Vec_IntEntry(vWeights, NameId)
			}
			Acb_ObjSetWeight(pNtk, ObjId, weight)
		}
	}
	// inlining of macro Ndr_ModForEachPo
	// inlining of macro Ndr_ModForEachObj
	for Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj) {
		if Ndr_DataType(p, Obj) != NDR_OBJECT {
		} else if !Ndr_ObjIsType(p, Obj, ABC_OPER_CO) {
		} else {
			var NameId int32 = Ndr_ObjReadBody(p, Obj, NDR_OUTPUT)
			var pName string = Abc_NamStr(pMan.pStrs, NameId)
			nArray = Ndr_ObjReadArray(p, Obj, NDR_INPUT, &pArray)
			assert(nArray == 1, "nArray == 1")
			ObjId = Acb_ObjAlloc(pNtk, ABC_OPER_CO, 1, 0)
			if Vec_IntEntry(vMap, pArray[0]) == 0 {
				fmt.Printf("Adding constant 0 driver to non-driven PO \"%s\".\n", pName)
				Vec_IntWriteEntry(vMap, pArray[0], Token0)
			}
			Acb_ObjAddFanin(pNtk, ObjId, Vec_IntEntry(vMap, pArray[0]))
			Acb_ObjSetName(pNtk, ObjId, pArray[0])
		}
	}
	Vec_IntFree(vMap)
	Acb_NtkSetRegNum(pNtk, 0)
	Acb_NtkAdd(pMan, pNtk)
	return pNtk
}

/**Function*************************************************************

  Synopsis    [Setup parameter structure.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkOptMfse(pNtk *Abc_Ntk_t, pPars *Acb_Par_t) *Abc_Ntk_t {

	var pNtkNew *Abc_Ntk_t
	var p *Acb_Ntk_t = Acb_NtkFromAbc(pNtk)
	Acb_NtkOpt(p, pPars)
	pNtkNew = Acb_NtkToAbc(pNtk, p)
	Acb_ManFree(p.pDesign)
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkOptPush(pNtk *Abc_Ntk_t, nLutSize int32, fVerbose int32) *Abc_Ntk_t {

	var pNtkNew *Abc_Ntk_t
	var p *Acb_Ntk_t = Acb_NtkFromAbc(pNtk)
	Acb_NtkPushLogic(p, nLutSize, fVerbose)
	pNtkNew = Acb_NtkToAbc(pNtk, p)
	Acb_ManFree(p.pDesign)
	return pNtkNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
