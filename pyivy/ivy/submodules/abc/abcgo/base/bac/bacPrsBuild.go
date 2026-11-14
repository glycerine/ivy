package bac

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [bacPrsBuild.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Parse tree to netlist transformation.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacPrsBuild.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "bacPrs.h"
//#include "map/mio/mio.h"
//#include "base/main/main.h"

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
func Psr_ManIsMapped(pNtk *Psr_Ntk_t) int32 {
	var iBox int32
	var pLib *Mio_Library_t = Abc_FrameReadLibGen()
	if pLib == nil {
		return 0
	}
	// inlining of macro Psr_NtkForEachBox
	for iBox = 0; iBox < Psr_NtkBoxNum(pNtk); iBox++ {
		if Psr_BoxIsNode(pNtk, iBox) == 0 {
			var NtkId int32 = Psr_BoxNtk(pNtk, iBox)
			if Mio_LibraryReadGateByName(pLib, Psr_NtkStr(pNtk, NtkId), "") != nil {
				return 1
			}
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_NtkCountObjects(pNtk *Psr_Ntk_t) int32 {
	var i int32
	var Count int32 = Psr_NtkObjNum(pNtk)
	// inlining of macro Psr_NtkForEachBox
	for i = 0; i < Psr_NtkBoxNum(pNtk); i++ {
		Count += Psr_BoxIONum(pNtk, i)
	}
	return Count
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// replaces NameIds of formal names by their index in the box model
func Psr_ManRemapOne(vSigs *vec.Vec_Int_t, pNtkBox *Psr_Ntk_t, vMap *vec.Vec_Int_t) {
	var i, NameId int32
	// map formal names into I/O indexes
	// inlining of macro Psr_NtkForEachPi
	for i = 0; i < Psr_NtkPiNum(pNtkBox); i++ {
		NameId = Vec_IntEntry(&pNtkBox.vInputs, i)
		assert(Vec_IntEntry(vMap, NameId) == -1, "Vec_IntEntry(vMap, NameId) == -1")
		Vec_IntWriteEntry(vMap, NameId, i+1) // +1 to keep 1st form input non-zero
	}
	// inlining of macro Psr_NtkForEachPo
	for i = 0; i < Psr_NtkPoNum(pNtkBox); i++ {
		NameId = Vec_IntEntry(&pNtkBox.vOutputs, i)
		assert(Vec_IntEntry(vMap, NameId) == -1, "Vec_IntEntry(vMap, NameId) == -1")
		Vec_IntWriteEntry(vMap, NameId, Psr_NtkPiNum(pNtkBox)+i+1) // +1 to keep 1st form input non-zero
	}
	// remap box
	assert(Vec_IntSize(vSigs)%2 == 0, "Vec_IntSize(vSigs) % 2 == 0")
	// inlining of macro vec.Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vSigs); i += 2 {
		NameId = Vec_IntEntry(vSigs, i)
		assert(Vec_IntEntry(vMap, NameId) != -1, "Vec_IntEntry(vMap, NameId) != -1")
		Vec_IntWriteEntry(vSigs, i, Vec_IntEntry(vMap, NameId))
	}
	// unmap formal inputs
	// inlining of macro Psr_NtkForEachPi
	for i = 0; i < Psr_NtkPiNum(pNtkBox); i++ {
		NameId = Vec_IntEntry(&pNtkBox.vInputs, i)
		Vec_IntWriteEntry(vMap, NameId, -1)
	}
	// inlining of macro Psr_NtkForEachPo
	for i = 0; i < Psr_NtkPoNum(pNtkBox); i++ {
		NameId = Vec_IntEntry(&pNtkBox.vOutputs, i)
		Vec_IntWriteEntry(vMap, NameId, -1)
	}
}
func Psr_ManRemapGate(vSigs *vec.Vec_Int_t) {
	var i int32
	// inlining of macro vec.Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vSigs); i++ {
		_ = Vec_IntEntry(vSigs, i)
		Vec_IntWriteEntry(vSigs, i, i/2+1)
		i++
	}
}
func Psr_ManRemapBoxes(pNew *Bac_Man_t, vDes *vec.Vec_Ptr_t, pNtk *Psr_Ntk_t, vMap *vec.Vec_Int_t) {
	var vSigs *vec.Vec_Int_t
	var iBox int32
	// inlining of macro Psr_NtkForEachBox
	for iBox = 0; iBox < Psr_NtkBoxNum(pNtk); iBox++ {
		vSigs = Psr_BoxSignals(pNtk, iBox)
		if Psr_BoxIsNode(pNtk, iBox) == 0 {
			var NtkId int32 = Psr_BoxNtk(pNtk, iBox)
			var NtkIdNew int32 = Bac_ManNtkFindId(pNew, Psr_NtkStr(pNtk, NtkId))
			assert(NtkIdNew > 0, "NtkIdNew > 0")
			Psr_BoxSetNtk(pNtk, iBox, NtkIdNew)
			if NtkIdNew <= Bac_ManNtkNum(pNew) {
				Psr_ManRemapOne(vSigs, Psr_ManNtk(vDes, NtkIdNew-1), vMap)
			}
			//else
			//    Psr_ManRemapGate( vSigs );
		}
	}
}
func Psr_ManCleanMap(pNtk *Psr_Ntk_t, vMap *vec.Vec_Int_t) {
	var vSigs *vec.Vec_Int_t
	var i, k, NameId, Sig int32
	// inlining of macro Psr_NtkForEachPi
	for i = 0; i < Psr_NtkPiNum(pNtk); i++ {
		NameId = Vec_IntEntry(&pNtk.vInputs, i)
		Vec_IntWriteEntry(vMap, NameId, -1)
	}
	// inlining of macro Psr_NtkForEachBox
	for i = 0; i < Psr_NtkBoxNum(pNtk); i++ {
		vSigs = Psr_BoxSignals(pNtk, i)
		// inlining of macro vec.Vec_IntForEachEntryDouble
		for k = 0; k+1 < Vec_IntSize(vSigs); k += 2 {
			NameId = Vec_IntEntry(vSigs, k)
			Sig = Vec_IntEntry(vSigs, k+1)
			Vec_IntWriteEntry(vMap, Psr_NtkSigName(pNtk, Sig), -1)
		}
	}
	// inlining of macro Psr_NtkForEachPo
	for i = 0; i < Psr_NtkPoNum(pNtk); i++ {
		NameId = Vec_IntEntry(&pNtk.vOutputs, i)
		Vec_IntWriteEntry(vMap, NameId, -1)
	}
}

// create maps of NameId and boxes
func Psr_ManBuildNtk(pNew *Bac_Ntk_t, vDes *vec.Vec_Ptr_t, pNtk *Psr_Ntk_t, vMap *vec.Vec_Int_t, vBoxes *vec.Vec_Int_t) {
	var pNtkBox *Psr_Ntk_t
	var vSigs *vec.Vec_Int_t
	var iBox int32
	var i, Index, NameId, iObj, iConst0, iTerm int32
	var iNonDriven int32 = -1
	var nNonDriven int32
	assert(Psr_NtkPioNum(pNtk) == 0, "Psr_NtkPioNum(pNtk) == 0")
	Psr_ManRemapBoxes(pNew.pDesign, vDes, pNtk, vMap)
	Bac_NtkStartNames(pNew)
	// create primary inputs
	// inlining of macro Psr_NtkForEachPi
	for i = 0; i < Psr_NtkPiNum(pNtk); i++ {
		NameId = Vec_IntEntry(&pNtk.vInputs, i)
		if Vec_IntEntry(vMap, NameId) != -1 {
			fmt.Printf("Primary inputs %d and %d have the same name.\n", Vec_IntEntry(vMap, NameId), i)
		}
		iObj = Bac_ObjAlloc(pNew, BAC_OBJ_PI, -1)
		Bac_ObjSetName(pNew, iObj, Abc_Var2Lit2(NameId, int32(BAC_NAME_BIN)))
		Vec_IntWriteEntry(vMap, NameId, iObj)
	}
	// create box outputs
	vec.Vec_IntClear(vBoxes)
	// inlining of macro Psr_NtkForEachBox
	for iBox = 0; iBox < Psr_NtkBoxNum(pNtk); iBox++ {
		vSigs = Psr_BoxSignals(pNtk, iBox)
		if Psr_BoxIsNode(pNtk, iBox) == 0 {
			pNtkBox = Psr_ManNtk(vDes, Psr_BoxNtk(pNtk, iBox)-1)
			if pNtkBox == nil {
				iObj = Bac_BoxAlloc(pNew, BAC_BOX_GATE, Vec_IntSize(vSigs)/2-1, 1, Psr_BoxNtk(pNtk, iBox))
				Bac_ObjSetName(pNew, iObj, Abc_Var2Lit2(Psr_BoxName(pNtk, iBox), int32(BAC_NAME_BIN)))
				// consider box output
				NameId = Vec_IntEntryLast(vSigs)
				NameId = Psr_NtkSigName(pNtk, NameId)
				if Vec_IntEntry(vMap, NameId) != -1 {
					fmt.Printf("Box output name %d is already driven.\n", NameId)
				}
				iTerm = Bac_BoxBo(pNew, iObj, 0)
				Bac_ObjSetName(pNew, iTerm, Abc_Var2Lit2(NameId, int32(BAC_NAME_BIN)))
				Vec_IntWriteEntry(vMap, NameId, iTerm)

			} else {
				iObj = Bac_BoxAlloc(pNew, BAC_OBJ_BOX, Psr_NtkPiNum(pNtkBox), Psr_NtkPoNum(pNtkBox), Psr_BoxNtk(pNtk, iBox))
				Bac_ObjSetName(pNew, iObj, Abc_Var2Lit2(Psr_BoxName(pNtk, iBox), int32(BAC_NAME_BIN)))
				Bac_NtkSetHost(Bac_ManNtk(pNew.pDesign, Psr_BoxNtk(pNtk, iBox)), Bac_NtkId(pNew), iObj)
				// inlining of macro vec.Vec_IntForEachEntry
				for i = 0; i < Vec_IntSize(vSigs); i++ {
					Index = Vec_IntEntry(vSigs, i)
					i++
					if Index--; Index < Psr_NtkPiNum(pNtkBox) {
						continue
					}
					assert(Index-Psr_NtkPiNum(pNtkBox) < Psr_NtkPoNum(pNtkBox), "Index - Psr_NtkPiNum(pNtkBox) < Psr_NtkPoNum(pNtkBox)")
					// consider box output
					NameId = Vec_IntEntry(vSigs, i)
					NameId = Psr_NtkSigName(pNtk, NameId)
					if Vec_IntEntry(vMap, NameId) != -1 {
						fmt.Printf("Box output name %d is already driven.\n", NameId)
					}
					iTerm = Bac_BoxBo(pNew, iObj, Index-Psr_NtkPiNum(pNtkBox))
					Bac_ObjSetName(pNew, iTerm, Abc_Var2Lit2(NameId, int32(BAC_NAME_BIN)))
					Vec_IntWriteEntry(vMap, NameId, iTerm)
				}
			}
			// remember box
			vec.Vec_IntPush(vBoxes, iObj)

		} else {
			iObj = Bac_BoxAlloc(pNew, Bac_ObjType_t(Psr_BoxNtk(pNtk, iBox)), Psr_BoxIONum(pNtk, iBox)-1, 1, -1)
			// consider box output
			NameId = Vec_IntEntryLast(vSigs)
			NameId = Psr_NtkSigName(pNtk, NameId)
			if Vec_IntEntry(vMap, NameId) != -1 {
				fmt.Printf("Node output name %d is already driven.\n", NameId)
			}
			iTerm = Bac_BoxBo(pNew, iObj, 0)
			Bac_ObjSetName(pNew, iTerm, Abc_Var2Lit2(NameId, int32(BAC_NAME_BIN)))
			Vec_IntWriteEntry(vMap, NameId, iTerm)
			// remember box
			vec.Vec_IntPush(vBoxes, iObj)
		}
	}
	// add fanins for box inputs
	// inlining of macro Psr_NtkForEachBox
	for iBox = 0; iBox < Psr_NtkBoxNum(pNtk); iBox++ {
		vSigs = Psr_BoxSignals(pNtk, iBox)
		if Psr_BoxIsNode(pNtk, iBox) == 0 {
			pNtkBox = Psr_ManNtk(vDes, Psr_BoxNtk(pNtk, iBox)-1)
			iObj = Vec_IntEntry(vBoxes, iBox)
			if pNtkBox == nil {
				// inlining of macro vec.Vec_IntForEachEntryStop
				for i = 0; i < Vec_IntSize(vSigs)-2; i++ {
					Index = Vec_IntEntry(vSigs, i)
					i++
					NameId = Vec_IntEntry(vSigs, i)
					NameId = Psr_NtkSigName(pNtk, NameId)
					iTerm = Bac_BoxBi(pNew, iObj, i/2)
					if Vec_IntEntry(vMap, NameId) == -1 {
						iConst0 = Bac_BoxAlloc(pNew, BAC_BOX_CF, 0, 1, -1)
						Vec_IntWriteEntry(vMap, NameId, iConst0+1)
						if iNonDriven == -1 {
							iNonDriven = NameId
						}
						nNonDriven++
					}
					Bac_ObjSetFanin(pNew, iTerm, Vec_IntEntry(vMap, NameId))
				}

			} else {
				// inlining of macro vec.Vec_IntForEachEntry
				for i = 0; i < Vec_IntSize(vSigs); i++ {
					Index = Vec_IntEntry(vSigs, i)
					i++
					if Index--; Index >= Psr_NtkPiNum(pNtkBox) {
						continue
					}
					NameId = Vec_IntEntry(vSigs, i)
					NameId = Psr_NtkSigName(pNtk, NameId)
					iTerm = Bac_BoxBi(pNew, iObj, Index)
					if Vec_IntEntry(vMap, NameId) == -1 {
						iConst0 = Bac_BoxAlloc(pNew, BAC_BOX_CF, 0, 1, -1)
						Vec_IntWriteEntry(vMap, NameId, iConst0+1)
						if iNonDriven == -1 {
							iNonDriven = NameId
						}
						nNonDriven++
					}
					Bac_ObjSetFanin(pNew, iTerm, Vec_IntEntry(vMap, NameId))
				}
			}

		} else {
			iObj = Vec_IntEntry(vBoxes, iBox)
			// inlining of macro vec.Vec_IntForEachEntryStop
			for i = 0; i < Vec_IntSize(vSigs)-2; i += 2 {
				Index = Vec_IntEntry(vSigs, i)
				NameId = Vec_IntEntry(vSigs, i+1)
				NameId = Psr_NtkSigName(pNtk, NameId)
				iTerm = Bac_BoxBi(pNew, iObj, i/2)
				if Vec_IntEntry(vMap, NameId) == -1 {
					iConst0 = Bac_BoxAlloc(pNew, BAC_BOX_CF, 0, 1, -1)
					Vec_IntWriteEntry(vMap, NameId, iConst0+1)
					if iNonDriven == -1 {
						iNonDriven = NameId
					}
					nNonDriven++
				}
				Bac_ObjSetFanin(pNew, iTerm, Vec_IntEntry(vMap, NameId))
			}
		}
	}
	// add fanins for primary outputs
	// inlining of macro Psr_NtkForEachPo
	for i = 0; i < Psr_NtkPoNum(pNtk); i++ {
		NameId = Vec_IntEntry(&pNtk.vOutputs, i)
		if Vec_IntEntry(vMap, NameId) == -1 {
			iConst0 = Bac_BoxAlloc(pNew, BAC_BOX_CF, 0, 1, -1)
			Vec_IntWriteEntry(vMap, NameId, iConst0+1)
			if iNonDriven == -1 {
				iNonDriven = NameId
			}
			nNonDriven++
		}
	}
	// inlining of macro Psr_NtkForEachPo
	for i = 0; i < Psr_NtkPoNum(pNtk); i++ {
		NameId = Vec_IntEntry(&pNtk.vOutputs, i)
		iObj = Bac_ObjAlloc(pNew, BAC_OBJ_PO, Vec_IntEntry(vMap, NameId))
	}
	if nNonDriven != 0 {
		fmt.Printf("Module %s has %d non-driven nets (for example, %s).\n", Psr_NtkName(pNtk), nNonDriven, Psr_NtkStr(pNtk, iNonDriven))
	}
	Psr_ManCleanMap(pNtk, vMap)
	// setup info
	// inlining of macro vec.Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(&pNtk.vOrder); i++ {
		NameId = Vec_IntEntry(&pNtk.vOrder, i)
		Bac_NtkAddInfo(pNew, NameId, -1, -1)
	}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Psr_ManBuildCba(pFileName string, vDes *vec.Vec_Ptr_t) *Bac_Man_t {
	var pNtk *Psr_Ntk_t = Psr_ManRoot(vDes)
	var i int32
	var pNew *Bac_Man_t = Bac_ManAlloc(pFileName, int32(vec.Vec_PtrSize(vDes)))
	var vMap *vec.Vec_Int_t = vec.Vec_IntStartFull(int(Abc_NamObjNumMax(pNtk.pStrs) + 1))
	var vTmp *vec.Vec_Int_t = vec.Vec_IntAlloc(int(Psr_NtkBoxNum(pNtk)))
	Abc_NamDeref(pNew.pStrs)
	pNew.pStrs = Abc_NamRef(pNtk.pStrs)
	// inlining of macro vec.Vec_PtrForEachEntry
	for i = 0; int(i) < vec.Vec_PtrSize(vDes); i++ {
		pNtk = vec.Vec_PtrEntry(vDes, int(i)).(*Psr_Ntk_t)
		Bac_NtkAlloc(Bac_ManNtk(pNew, i+1), Psr_NtkId(pNtk), Psr_NtkPiNum(pNtk), Psr_NtkPoNum(pNtk), Psr_NtkCountObjects(pNtk))
	}
	if (pNtk.fMapped != 0 || (pNtk.fSlices != 0 && (Psr_ManIsMapped(pNtk) != 0))) && Bac_NtkBuildLibrary(pNew) == 0 {
		Bac_ManFree(pNew)
		pNew = nil
	} else {
		// inlining of macro vec.Vec_PtrForEachEntry
		for i = 0; int(i) < vec.Vec_PtrSize(vDes); i++ {
			pNtk = vec.Vec_PtrEntry(vDes, int(i)).(*Psr_Ntk_t)
			Psr_ManBuildNtk(Bac_ManNtk(pNew, i+1), vDes, pNtk, vMap, vTmp)
		}
	}
	assert(int32(vec.Vec_IntCountEntry(vMap, -1)) == Vec_IntSize(vMap), "int32(vec.Vec_IntCountEntry(vMap, -1)) == Vec_IntSize(vMap)")
	vec.Vec_IntFree(vMap)
	vec.Vec_IntFree(vTmp)
	//    vec.Vec_StrPrint( &Bac_ManNtk(pNew, 1)->vType, 1 );
	return pNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
