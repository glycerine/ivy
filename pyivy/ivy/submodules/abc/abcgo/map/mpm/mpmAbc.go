package mpm

import (
	gia "github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	. "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [mpmAbc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [Interface with ABC data structures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmAbc.c,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "mpmInt.h"

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
func Mig_ManCreateChoices(pMig *Mig_Man_t, p *gia.Gia_Man_t) {
	var pObj *gia.Gia_Obj_t
	var i int32
	assert(Mig_ManObjNum(pMig) == gia.Gia_ManObjNum(p), "Mig_ManObjNum(pMig) == gia.Gia_ManObjNum(p)")
	assert(vec.Vec_IntSize(&pMig.vSibls) == 0, "vec.Vec_IntSize(&pMig.vSibls) == 0")
	vec.Vec_IntFill(&pMig.vSibls, int(gia.Gia_ManObjNum(p)), 0)
	gia.Gia_ManMarkFanoutDrivers(p)
	// inlining of macro gia.Gia_ManForEachObj
	for i = 0; i < gia.Gia_ManObjNum(p); i++ {
		pObj = gia.Gia_ManObj(p, i)
		gia.Gia_ObjSetPhase(p, pObj)
		assert(Abc_Lit2Var(int32(pObj.Value)) == i, "Abc_Lit2Var(int32(pObj.Value)) == i")
		Mig_ObjSetPhase(Mig_ManObj(pMig, i), gia.Gia_ObjPhase(pObj))
		if gia.Gia_ObjSibl(p, i) != 0 && gia.Gia_ObjMark0(pObj) != 0 {
			var pSibl, pPrev *gia.Gia_Obj_t
			pPrev = pObj
			pSibl = gia.Gia_ObjSiblObj(p, i)
			for pSibl != nil {
				Mig_ObjSetSiblId(Mig_ManObj(pMig, Abc_Lit2Var(int32(pPrev.Value))), Abc_Lit2Var(int32(pSibl.Value)))
				pPrev = pSibl
				pSibl = gia.Gia_ObjSiblObj(p, gia.Gia_ObjId(p, pSibl))
			}
			pMig.nChoices++
		}
	}
	gia.Gia_ManCleanMark0(p)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mig_ObjFanin0Copy(pObj *gia.Gia_Obj_t) int32 {
	return Abc_LitNotCond(int32(gia.Gia_ObjFanin0(pObj).Value), gia.Gia_ObjFaninC0(pObj))
}
func Mig_ObjFanin1Copy(pObj *gia.Gia_Obj_t) int32 {
	return Abc_LitNotCond(int32(gia.Gia_ObjFanin1(pObj).Value), gia.Gia_ObjFaninC1(pObj))
}
func Mig_ManCreate(pGia any) *Mig_Man_t {
	p, ok := pGia.(*gia.Gia_Man_t)
	assert(ok, "pGia is *gia.Gia_Man_t")
	var pNew *Mig_Man_t
	var pObj *gia.Gia_Obj_t
	var i int32
	pNew = Mig_ManStart()
	pNew.pName = Abc_UtilStrsav(gia.Gia_ManName(p))
	gia.Gia_ManConst0(p).Value = 0
	// inlining of macro gia.Gia_ManForEachObj1
	for i = 1; i < gia.Gia_ManObjNum(p); i++ {
		pObj = gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsMuxId(p, i) != 0 {
			pObj.Value = uint32(Mig_ManAppendMux(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj), gia.Gia_ObjFanin2Copy(p, pObj)))
		} else if gia.Gia_ObjIsXor(pObj) != 0 {
			pObj.Value = uint32(Mig_ManAppendXor(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))
		} else if gia.Gia_ObjIsAnd(pObj) != 0 {
			pObj.Value = uint32(Mig_ManAppendAnd(pNew, Mig_ObjFanin0Copy(pObj), Mig_ObjFanin1Copy(pObj)))
		} else if gia.Gia_ObjIsCi(pObj) != 0 {
			pObj.Value = uint32(Mig_ManAppendCi(pNew))
		} else if gia.Gia_ObjIsCo(pObj) != 0 {
			pObj.Value = uint32(Mig_ManAppendCo(pNew, Mig_ObjFanin0Copy(pObj)))
		} else {
			assert(false, "0")
		}
	}
	Mig_ManSetRegNum(pNew, gia.Gia_ManRegNum(p))
	if gia.Gia_ManHasChoices(p) != 0 {
		Mig_ManCreateChoices(pNew, p)
	}
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Recursively derives the local AIG for the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_CutDataInt(pCut *Mpm_Cut_t) int32          { return pCut.hNext }
func Mpm_CutSetDataInt(pCut *Mpm_Cut_t, Data int32) { pCut.hNext = Data }
func Mpm_ManNodeIfToGia_rec(pNew *gia.Gia_Man_t, pMan *Mpm_Man_t, pObj *Mig_Obj_t, vVisited *vec.Vec_Ptr_t, fHash int32) int32 {
	var pTemp *Mig_Obj_t
	var pCut *Mpm_Cut_t
	var iFunc, iFunc0, iFunc1, iFunc2 int32
	assert(fHash == 0, "fHash == 0")
	// get the best cut
	pCut = Mpm_ObjCutBestP(pMan, pObj)
	// if the cut is visited, return the result
	if Mpm_CutDataInt(pCut) != 0 {
		return Mpm_CutDataInt(pCut)
	}
	// mark the node as visited
	vec.Vec_PtrPush(vVisited, pCut)
	// insert the worst case
	Mpm_CutSetDataInt(pCut, ^int32(0))
	// skip in case of primary input
	if Mig_ObjIsCi(pObj) != 0 {
		return Mpm_CutDataInt(pCut)
	}
	// compute the functions of the children
	for pTemp = pObj; pTemp != nil; pTemp = Mig_ObjSibl(pTemp) {
		iFunc0 = Mpm_ManNodeIfToGia_rec(pNew, pMan, Mig_ObjFanin0(pTemp), vVisited, fHash)
		if iFunc0 == ^int32(0) {
			continue
		}
		iFunc1 = Mpm_ManNodeIfToGia_rec(pNew, pMan, Mig_ObjFanin1(pTemp), vVisited, fHash)
		if iFunc1 == ^int32(0) {
			continue
		}
		if Mig_ObjIsNode3(pTemp) != 0 {
			iFunc2 = Mpm_ManNodeIfToGia_rec(pNew, pMan, Mig_ObjFanin2(pTemp), vVisited, fHash)
			if iFunc2 == ^int32(0) {
				continue
			}
			iFunc2 = Abc_LitNotCond(iFunc2, Mig_ObjFaninC2(pTemp))
		}
		iFunc0 = Abc_LitNotCond(iFunc0, Mig_ObjFaninC0(pTemp))
		iFunc1 = Abc_LitNotCond(iFunc1, Mig_ObjFaninC1(pTemp))
		// both branches are solved
		if fHash != 0 {
			if Mig_ObjIsMux(pTemp) != 0 {
				iFunc = gia.Gia_ManHashMux(pNew, iFunc2, iFunc1, iFunc0)
			} else if Mig_ObjIsXor(pTemp) != 0 {
				iFunc = gia.Gia_ManHashXor(pNew, iFunc0, iFunc1)
			} else {
				iFunc = gia.Gia_ManHashAnd(pNew, iFunc0, iFunc1)
			}

		} else {
			if Mig_ObjIsMux(pTemp) != 0 {
				iFunc = gia.Gia_ManAppendMux(pNew, iFunc2, iFunc1, iFunc0)
			} else if Mig_ObjIsXor(pTemp) != 0 {
				iFunc = gia.Gia_ManAppendXor(pNew, iFunc0, iFunc1)
			} else {
				iFunc = gia.Gia_ManAppendAnd(pNew, iFunc0, iFunc1)
			}
		}
		if Mig_ObjPhase(pTemp) != Mig_ObjPhase(pObj) {
			iFunc = Abc_LitNot(iFunc)
		}
		Mpm_CutSetDataInt(pCut, iFunc)
		break
	}
	return Mpm_CutDataInt(pCut)
}
func Mpm_ManNodeIfToGia(pNew *gia.Gia_Man_t, pMan *Mpm_Man_t, pObj *Mig_Obj_t, vLeaves *vec.Vec_Int_t, fHash int32) int32 {
	var pCut *Mpm_Cut_t
	var pFanin *Mig_Obj_t
	var i, iRes int32
	// get the best cut
	pCut = Mpm_ObjCutBestP(pMan, pObj)
	assert(pCut.nLeaves > 1, "pCut.nLeaves > 1")
	// set the leaf variables
	// inlining of macro Mpm_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pFanin = Mig_ManObj(pMan.pMig, Abc_Lit2Var(pCut.pLeaves[i]))
		Mpm_CutSetDataInt(Mpm_ObjCutBestP(pMan, pFanin), vec.Vec_IntEntry(vLeaves, int(i)))
	}
	// recursively compute the function while collecting visited cuts
	vec.Vec_PtrClear(pMan.vTemp)
	iRes = Mpm_ManNodeIfToGia_rec(pNew, pMan, pObj, pMan.vTemp, fHash)
	if iRes == ^int32(0) {
		Abc_Print(-1, "Mpm_ManNodeIfToGia(): Computing local AIG has failed.\n")
		return ^int32(0)
	}
	// clean the cuts
	// inlining of macro Mpm_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pFanin = Mig_ManObj(pMan.pMig, Abc_Lit2Var(pCut.pLeaves[i]))
		Mpm_CutSetDataInt(Mpm_ObjCutBestP(pMan, pFanin), 0)
	}
	// inlining of macro vec.Vec_PtrForEachEntry
	for i = 0; int(i) < vec.Vec_PtrSize(pMan.vTemp); i++ {
		pCut = vec.Vec_PtrEntry(pMan.vTemp, int(i)).(*Mpm_Cut_t)
		Mpm_CutSetDataInt(pCut, 0)
	}
	return iRes
}
func Mpm_ManFromIfLogic(pMan *Mpm_Man_t) any {
	var pNew *gia.Gia_Man_t
	var pCutBest *Mpm_Cut_t
	var pObj, pFanin *Mig_Obj_t
	var vMapping *vec.Vec_Int_t
	var vMapping2 *vec.Vec_Int_t
	var vPacking *vec.Vec_Int_t = nil
	var vLeaves, vLeaves2, vCover *vec.Vec_Int_t
	var uTruth word
	var pTruth []word
	var i, k, Entry, iLitNew int32
	//    assert( !pMan->pPars->fDeriveLuts || pMan->pPars->fTruth );
	// start mapping and packing
	vMapping = vec.Vec_IntStart(int(Mig_ManObjNum(pMan.pMig)))
	vMapping2 = vec.Vec_IntStart(1)
	if false { // pMan->pPars->fDeriveLuts && pMan->pPars->pLutStruct )

		vPacking = vec.Vec_IntAlloc(1000)
		vec.Vec_IntPush(vPacking, 0)
	}
	// create new manager
	pNew = gia.Gia_ManStart(Mig_ManObjNum(pMan.pMig))
	// iterate through nodes used in the mapping
	vCover = vec.Vec_IntAlloc(1 << 16)
	vLeaves = vec.Vec_IntAlloc(16)
	vLeaves2 = vec.Vec_IntAlloc(16)
	Mig_ManCleanCopy(pMan.pMig)
	// inlining of macro Mig_ManForEachObj
	for id := int32(0); id < Mig_ManObjNum(pMan.pMig); id++ {
		pObj = Mig_ManObj(pMan.pMig, id)
		if Mpm_ObjMapRef(pMan, pObj) == 0 && Mig_ObjIsTerm(pObj) == 0 {
			continue
		}
		if Mig_ObjIsNode(pObj) != 0 {
			// collect leaves of the best cut
			vec.Vec_IntClear(vLeaves)
			pCutBest = Mpm_ObjCutBestP(pMan, pObj)
			// inlining of macro Mpm_CutForEachLeaf
			for k = 0; k < int32(pCutBest.nLeaves); k++ {
				pFanin = Mig_ManObj(pMan.pMig, Abc_Lit2Var(pCutBest.pLeaves[k]))
				vec.Vec_IntPush(vLeaves, Mig_ObjCopy(pFanin))
			}
			if pMan.pPars.FDeriveLuts != 0 && (pMan.pPars.FUseTruth != 0 || pMan.pPars.FUseDsd != 0) {
				if pMan.pPars.FUseTruth != 0 {
					pTruth = Mpm_CutTruth(pMan, Abc_Lit2Var(int32(pCutBest.iFunc)))
				} else {
					uTruth = Mpm_CutTruthFromDsd(pMan, pCutBest, Abc_Lit2Var(int32(pCutBest.iFunc)))
					pTruth = []word{uTruth}
				}
				//                Kit_DsdPrintFromTruth( pTruth, vec.Vec_IntSize(vLeaves) );  printf( "\n" );
				// perform decomposition of the cut
				iLitNew = if_.Gia_ManFromIfLogicNode(nil, pNew, Mig_ObjId(pObj), vLeaves, vLeaves2, pTruth, "", vCover, vMapping, vMapping2, vPacking, 0, 0)
				iLitNew = Abc_LitNotCond(iLitNew, int32(pCutBest.fCompl)^Abc_LitIsCompl(int32(pCutBest.iFunc)))

			} else {
				// perform one of the two types of mapping: with and without structures
				iLitNew = Mpm_ManNodeIfToGia(pNew, pMan, pObj, vLeaves, 0)
				// write mapping
				vec.Vec_IntSetEntry(vMapping, int(Abc_Lit2Var(iLitNew)), int32(vec.Vec_IntSize(vMapping2)))
				vec.Vec_IntPush(vMapping2, int32(vec.Vec_IntSize(vLeaves)))
				// inlining of macro vec.Vec_IntForEachEntry
				for k = 0; int(k) < vec.Vec_IntSize(vLeaves); k++ {
					Entry = vec.Vec_IntEntry(vLeaves, int(k))
					assert(Abc_Lit2Var(Entry) < Abc_Lit2Var(iLitNew), "Abc_Lit2Var(Entry) < Abc_Lit2Var(iLitNew)")
				}
				// inlining of macro vec.Vec_IntForEachEntry
				for k = 0; int(k) < vec.Vec_IntSize(vLeaves); k++ {
					Entry = vec.Vec_IntEntry(vLeaves, int(k))
					vec.Vec_IntPush(vMapping2, Abc_Lit2Var(Entry))
				}
				vec.Vec_IntPush(vMapping2, Abc_Lit2Var(iLitNew))
			}

		} else if Mig_ObjIsCi(pObj) != 0 {
			iLitNew = gia.Gia_ManAppendCi(pNew)
		} else if Mig_ObjIsCo(pObj) != 0 {
			iLitNew = gia.Gia_ManAppendCo(pNew, Abc_LitNotCond(Mig_ObjCopy(Mig_ObjFanin0(pObj)), Mig_ObjFaninC0(pObj)))
		} else if Mig_ObjIsConst0(pObj) != 0 {
			iLitNew = 0
			// create const LUT
			vec.Vec_IntWriteEntry(vMapping, 0, int32(vec.Vec_IntSize(vMapping2)))
			vec.Vec_IntPush(vMapping2, 0)
			vec.Vec_IntPush(vMapping2, 0)

		} else {
			assert(false, "0")
		}
		Mig_ObjSetCopy(pObj, iLitNew)
	}
	vec.Vec_IntFree(vCover)
	vec.Vec_IntFree(vLeaves)
	vec.Vec_IntFree(vLeaves2)
	//    printf( "Mapping array size:  IfMan = %d. Gia = %d. Increase = %.2f\n",
	//        Mig_ManObjNum(pMan), gia.Gia_ManObjNum(pNew), 1.0 * gia.Gia_ManObjNum(pNew) / Mig_ManObjNum(pMan) );
	// finish mapping
	if vec.Vec_IntSize(vMapping) > int(gia.Gia_ManObjNum(pNew)) {
		vec.Vec_IntShrink(vMapping, int(gia.Gia_ManObjNum(pNew)))
	} else {
		vec.Vec_IntFillExtra(vMapping, int(gia.Gia_ManObjNum(pNew)), 0)
	}
	assert(vec.Vec_IntSize(vMapping) == int(gia.Gia_ManObjNum(pNew)), "vec.Vec_IntSize(vMapping) == int(gia.Gia_ManObjNum(pNew))")
	// inlining of macro vec.Vec_IntForEachEntry
	for i = 0; int(i) < vec.Vec_IntSize(vMapping); i++ {
		Entry = vec.Vec_IntEntry(vMapping, int(i))
		if Entry > 0 {
			vec.Vec_IntAddToEntry(vMapping, int(i), gia.Gia_ManObjNum(pNew))
		}
	}
	vec.Vec_IntAppend(vMapping, vMapping2)
	vec.Vec_IntFree(vMapping2)
	// attach mapping and packing
	assert(gia.Gia_ManMapping(pNew) == nil, "Gia_ManMapping(pNew) == nil")
	assert(gia.Gia_ManPacking(pNew) == nil, "Gia_ManPacking(pNew) == nil")
	gia.Gia_ManSetMapping(pNew, vMapping)
	assert(pNew.ReplacePacking(vPacking) == nil, "replace Gia packing")
	// verify that COs have mapping
	{
		var pObj *gia.Gia_Obj_t
		// inlining of macro gia.Gia_ManForEachCo
		for i = 0; int(i) < int(gia.Gia_ManCoNum(pNew)); i++ {
			pObj = gia.Gia_ManCo(pNew, i)
			assert(gia.Gia_ObjIsAnd(gia.Gia_ObjFanin0(pObj)) == 0 || gia.Gia_ObjIsLut(pNew, gia.Gia_ObjFaninId0p(pNew, pObj)) != 0, "!gia.Gia_ObjIsAnd(gia.Gia_ObjFanin0(pObj)) || gia.Gia_ObjIsLut(pNew, gia.Gia_ObjFaninId0p(pNew, pObj))")
		}
	}
	return pNew
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
void Mig_ManTest2( gia.Gia_Man_t * pGia )
{
    extern int gia.Gia_ManSuppSizeTest( gia.Gia_Man_t * p );
    Mig_Man_t * p;
    gia.Gia_ManSuppSizeTest( pGia );
    p = Mig_ManCreate( pGia );
    Mig_ManSuppSizeTest( p );
    Mig_ManStop( p );
}
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
