package mpm

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
	. "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [mpmMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [Mapping algorithm.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmMap.c,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mpmInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//#define MIG_RUNTIME

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Cut manipulation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ObjAddCutToStore(p *Mpm_Man_t, pCut *Mpm_Cut_t, ArrTime int32) int32 {
	var fEnableContainment int32 = 1
	var pUnit, pUnitNew *Mpm_Uni_t
	var k, iPivot, last int32
	// create new unit
	pUnitNew = Mpm_CutSetupInfo(p, pCut, ArrTime)
	// special case when the cut store is empty
	if p.nCutStore == 0 {
		p.pCutStore[p.nCutStore] = pUnitNew
		p.nCutStore++
		Vec_PtrPop(&p.vFreeUnits)
		return 1
	}
	// special case when the cut store is full and last cut is better than new cut
	if p.nCutStore == p.nNumCuts-1 && p.pCutCmp(pUnitNew, p.pCutStore[p.nCutStore-1]) > 0 {
		return 0
	}

	// find place of the given cut in the store
	assert(p.nCutStore <= p.nNumCuts, "p.nCutStore <= p.nNumCuts")
	for iPivot = p.nCutStore - 1; iPivot >= 0; iPivot-- {
		if p.pCutCmp(pUnitNew, p.pCutStore[iPivot]) > 0 { // iPivot-th cut is better than new cut
			break
		}
	}

	if fEnableContainment != 0 {
		// filter this cut using other cuts
		for k = 0; k <= iPivot; k++ {
			pUnit = p.pCutStore[k]
			if pUnitNew.pCut.nLeaves >= pUnit.pCut.nLeaves &&
				(pUnitNew.uSign&pUnit.uSign) == pUnit.uSign &&
				Mpm_CutIsContained(p, &pUnitNew.pCut, &pUnit.pCut) != 0 {
				return 0
			}
		}
	}

	// special case when the best cut is useless while the new cut is not
	if p.pCutStore[0].pCut.fUseless != 0 && pUnitNew.pCut.fUseless == 0 {
		iPivot = -1
	}

	// add the cut to storage
	assert(pUnitNew == Vec_PtrEntryLast(&p.vFreeUnits).(*Mpm_Uni_t), "pUnitNew == Vec_PtrEntryLast(&p.vFreeUnits)")
	Vec_PtrPop(&p.vFreeUnits)

	// insert this cut at location iPivot
	iPivot++
	k = p.nCutStore
	p.nCutStore++
	for k > iPivot {
		p.pCutStore[k] = p.pCutStore[k-1]
		k--
	}
	p.pCutStore[iPivot] = pUnitNew

	if fEnableContainment != 0 {
		// filter other cuts using this cut
		k = iPivot + 1
		last = k
		for k < p.nCutStore {
			pUnit = p.pCutStore[k]
			if pUnitNew.pCut.nLeaves <= pUnit.pCut.nLeaves &&
				(pUnitNew.uSign&pUnit.uSign) == pUnitNew.uSign &&
				Mpm_CutIsContained(p, &pUnit.pCut, &pUnitNew.pCut) != 0 {
				Vec_PtrPush(&p.vFreeUnits, pUnit)
				continue
			}
			p.pCutStore[last] = p.pCutStore[k]
			last++
			k++
		}
		p.nCutStore = last
	}

	// remove the last cut if too many
	if p.nCutStore == p.nNumCuts {
		p.nCutStore--
		Vec_PtrPush(&p.vFreeUnits, p.pCutStore[p.nCutStore])
	}
	assert(p.nCutStore < p.nNumCuts, "p.nCutStore < p.nNumCuts")
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Cut enumeration.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ManExploreNewCut(p *Mpm_Man_t, pObj *Mig_Obj_t, pCut0 *Mpm_Cut_t, pCut1 *Mpm_Cut_t, pCut2 *Mpm_Cut_t, Required int32) int32 {
	var pCut *Mpm_Cut_t
	var ArrTime int32

	if pCut0.nLeaves >= pCut1.nLeaves {
		pCut = Mpm_ManMergeCuts(p, pCut0, pCut1, pCut2)
		if pCut == nil {
			return 1
		}
		if p.pPars.fUseTruth != 0 {
			Mpm_CutComputeTruth(p, pCut, pCut0, pCut1, pCut2, Mig_ObjFaninC0(pObj), Mig_ObjFaninC1(pObj), Mig_ObjFaninC2(pObj), Mig_ObjNodeType(pObj))
		} else if p.pPars.fUseDsd != 0 {
			if Mpm_CutComputeDsd6(p, pCut, pCut0, pCut1, pCut2, Mig_ObjFaninC0(pObj), Mig_ObjFaninC1(pObj), Mig_ObjFaninC2(pObj), Mig_ObjNodeType(pObj)) == 0 {
				return 1
			}
		}

	} else {
		pCut = Mpm_ManMergeCuts(p, pCut1, pCut0, pCut2)
		if pCut == nil {
			return 1
		}
		if p.pPars.fUseTruth != 0 {
			Mpm_CutComputeTruth(p, pCut, pCut1, pCut0, pCut2, Mig_ObjFaninC1(pObj), Mig_ObjFaninC0(pObj), 1^Mig_ObjFaninC2(pObj), Mig_ObjNodeType(pObj))
		} else if p.pPars.fUseDsd != 0 {
			if Mpm_CutComputeDsd6(p, pCut, pCut1, pCut0, pCut2, Mig_ObjFaninC1(pObj), Mig_ObjFaninC0(pObj), 1^Mig_ObjFaninC2(pObj), Mig_ObjNodeType(pObj)) == 0 {
				return 1
			}
		}
	}

	ArrTime = Mpm_CutGetArrTime(p, pCut)
	if p.fMainRun != 0 && ArrTime > Required {
		return 1
	}

	Mpm_ObjAddCutToStore(p, pCut, ArrTime)

	/*
	       // return 0 if const or buffer cut is derived - reset all cuts to contain only one --- does not work
	   //    if ( pCut->nLeaves < 2 && p->nCutStore == 1 )
	   //        return 0;
	       if ( pCut->nLeaves < 2 )
	       {
	           int i;
	           assert( p->nCutStore >= 1 );
	           for ( i = 1; i < p->nCutStore; i++ )
	               Vec_PtrPush( &p->vFreeUnits, p->pCutStore[i] );
	           p->nCutStore = 1;
	           return 0;
	       }
	*/
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ObjRecycleCuts(p *Mpm_Man_t, pObj *Mig_Obj_t) {
	var pCut *Mpm_Cut_t
	var hCut, hNext int32
	// inlining of macro Mpm_ObjForEachCutSafe
	for hCut = Mpm_ObjCutList(p, pObj); hCut != 0; hCut = hNext {
		pCut = Mpm_CutFetch(p, hCut)
		if pCut == nil {
			break
		}
		hNext = pCut.hNext
		mem.Mmr_StepRecycle(p.pManCuts, hCut)
	}
	Mpm_ObjSetCutList(p, pObj, 0)
}
func Mpm_ObjDerefFaninCuts(p *Mpm_Man_t, pObj *Mig_Obj_t) {
	var pFanin *Mig_Obj_t
	var i int32
	// inlining of macro Mig_ObjForEachFanin
	for i = 0; Mig_ObjHasFanin(pObj, i) != 0; i++ {
		pFanin = Mig_ObjFanin(pObj, i)
		if Mig_ObjIsNode(pFanin) != 0 && Mig_ObjMigRefDec(p, pFanin) == 0 {
			Mpm_ObjRecycleCuts(p, pFanin)
		}
	}
	pFanin = Mig_ObjSibl(pObj)
	if pFanin != nil && Mig_ObjMigRefDec(p, pFanin) == 0 {
		Mpm_ObjRecycleCuts(p, pFanin)
	}
	if Mig_ObjMigRefNum(p, pObj) == 0 {
		Mpm_ObjRecycleCuts(p, pObj)
	}
}
func Mpm_ObjCollectFaninsAndSigns(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) {
	var pCut *Mpm_Cut_t
	var hCut, nCuts int32
	// inlining of macro Mpm_ObjForEachCut
	for hCut = Mpm_ObjCutList(p, pObj); hCut != 0; hCut = pCut.hNext {
		pCut = Mpm_CutFetch(p, hCut)
		if pCut == nil {
			break
		}
		p.pCuts[i][nCuts] = pCut
		p.pSigns[i][nCuts] = Mpm_CutGetSign(pCut)
		nCuts++
	}
	p.nCuts[i] = nCuts
}
func Mpm_ObjPrepareFanins(p *Mpm_Man_t, pObj *Mig_Obj_t) {
	var pFanin *Mig_Obj_t
	var i int32
	// inlining of macro Mig_ObjForEachFanin
	for i = 0; Mig_ObjHasFanin(pObj, i) != 0; i++ {
		pFanin = Mig_ObjFanin(pObj, i)
		Mpm_ObjCollectFaninsAndSigns(p, pFanin, i)
	}
}

// create storage from cuts at the node
func Mpm_ObjAddChoiceCutsToStore(p *Mpm_Man_t, pRoot *Mig_Obj_t, pObj *Mig_Obj_t, ReqTime int32) {
	var pCut *Mpm_Cut_t
	var hCut, hNext, ArrTime int32
	var fCompl int32 = Mig_ObjPhase(pRoot) ^ Mig_ObjPhase(pObj)
	// inlining of macro Mpm_ObjForEachCutSafe
	for hCut = Mpm_ObjCutList(p, pObj); hCut != 0; hCut = hNext {
		pCut = Mpm_CutFetch(p, hCut)
		if pCut == nil {
			break
		}
		hNext = pCut.hNext
		if util.Abc_Lit2Var(pCut.pLeaves[0]) == Mig_ObjId(pObj) {
			continue
		}
		ArrTime = Mpm_CutGetArrTime(p, pCut)
		if ArrTime > ReqTime {
			continue
		}
		pCut.fCompl ^= uint32(fCompl)
		pCut = Mpm_ManMergeCuts(p, pCut, nil, nil)
		Mpm_ObjAddCutToStore(p, pCut, ArrTime)
	}
}

// create cuts at the node from storage
func Mpm_ObjTranslateCutsFromStore(p *Mpm_Man_t, pObj *Mig_Obj_t) {
	var pCut *Mpm_Cut_t = nil
	var pUnit *Mpm_Uni_t
	var i int32
	var pList *int32 = Mpm_ObjCutListP(p, pObj)
	assert(p.nCutStore > 0 && p.nCutStore <= p.nNumCuts, "p.nCutStore > 0 && p.nCutStore <= p.nNumCuts")
	assert(*pList == 0, "*pList == 0")
	// translate cuts
	for i = 0; i < p.nCutStore; i++ {
		pUnit = p.pCutStore[i]
		*pList = Mpm_CutCreate(p, &pUnit.pCut, &pCut)
		pList = &pCut.hNext
		Vec_PtrPush(&p.vFreeUnits, pUnit)
	}
	assert(int32(Vec_PtrSize(&p.vFreeUnits)) == p.nNumCuts+1, "int32(Vec_PtrSize(&p.vFreeUnits)) == p.nNumCuts + 1")
	if p.nCutStore == 1 && pCut.nLeaves < 2 {
		*pList = 0
	} else {
		*pList = Mpm_CutCreateUnit(p, Mig_ObjId(pObj))
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
func Mpm_ManDeriveCuts(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 {
	var pCut0, pCut1, pCut2 *Mpm_Cut_t
	var Required int32 = Mpm_ObjRequired(p, pObj)
	var hCutBest int32 = Mpm_ObjCutBest(p, pObj)
	var c0, c1, c2 int32

	assert(int32(Vec_PtrSize(&p.vFreeUnits)) == p.nNumCuts+1, "Vec_PtrSize( &p.vFreeUnits ) == p.nNumCuts + 1")
	assert(Mpm_ObjCutList(p, pObj) == 0, "Mpm_ObjCutList(p, pObj) == 0")
	p.nCutStore = 0
	if hCutBest > 0 { // cut list is assigned

		var pCut *Mpm_Cut_t = Mpm_ObjCutBestP(p, pObj)
		var Times int32 = Mpm_CutGetArrTime(p, pCut)
		assert(pCut.hNext == 0, "pCut.hNext == 0")
		if Times > Required {
			fmt.Printf("Arrival time (%d) exceeds required time (%d) at object %d.\n", Times, Required, Mig_ObjId(pObj))
		}
		if p.fMainRun != 0 {
			Mpm_ObjAddCutToStore(p, Mpm_ManMergeCuts(p, pCut, nil, nil), Times)
		} else {
			Mpm_ObjSetTime(p, pObj, Times)
		}
	}
	// start storage with choice cuts
	if Mig_ManChoiceNum(p.pMig) != 0 && Mig_ObjSiblId(pObj) != 0 {
		Mpm_ObjAddChoiceCutsToStore(p, pObj, Mig_ObjSibl(pObj), Required)
	}

	Mpm_ObjPrepareFanins(p, pObj)
	if Mig_ObjIsNode2(pObj) != 0 {
		// go through cut pairs
		for c0 = 0; c0 < p.nCuts[0]; c0++ {
			pCut0 = p.pCuts[0][c0]
			for c1 = 0; c1 < p.nCuts[1]; c1++ {
				pCut1 = p.pCuts[1][c1]
				if util.Abc_TtCountOnes(p.pSigns[0][c0]|p.pSigns[1][c1]) <= p.nLutSize {
					if !(Mpm_ManExploreNewCut(p, pObj, pCut0, pCut1, nil, Required) != 0) {
						goto finish
					}
				}
			}
		}

	} else if Mig_ObjIsNode3(pObj) != 0 {
		// go through cut triples
		for c0 = 0; c0 < p.nCuts[0]; c0++ {
			pCut0 = p.pCuts[0][c0]
			for c1 = 0; c1 < p.nCuts[1]; c1++ {
				pCut1 = p.pCuts[1][c1]
				for c2 = 0; c2 < p.nCuts[2]; c2++ {
					pCut2 = p.pCuts[2][c2]
					if util.Abc_TtCountOnes(p.pSigns[0][c0]|p.pSigns[1][c1]|p.pSigns[2][c2]) <= p.nLutSize {
						if !(Mpm_ManExploreNewCut(p, pObj, pCut0, pCut1, pCut2, Required) != 0) {
							goto finish
						}
					}
				}
			}
		}

	} else {
		assert(false, "0")
	}

finish:
	// save best cut
	assert(p.nCutStore > 0, "p.nCutStore > 0")
	if p.pCutStore[0].mTime <= Required {
		var pCut *Mpm_Cut_t
		if hCutBest != 0 {
			mem.Mmr_StepRecycle(p.pManCuts, hCutBest)
		}
		hCutBest = Mpm_CutCreate(p, &p.pCutStore[0].pCut, &pCut)
		Mpm_ObjSetCutBest(p, pObj, hCutBest)
		Mpm_ObjSetTime(p, pObj, p.pCutStore[0].mTime)
		Mpm_ObjSetArea(p, pObj, p.pCutStore[0].mArea)
		Mpm_ObjSetEdge(p, pObj, p.pCutStore[0].mEdge)

	} else {
		assert(p.fMainRun == 0, "p.fMainRun == 0")
	}
	assert(hCutBest > 0, "hCutBest > 0")
	// transform internal storage into regular cuts
	Mpm_ObjTranslateCutsFromStore(p, pObj)
	// dereference fanin cuts and reference node
	Mpm_ObjDerefFaninCuts(p, pObj)
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Required times.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ManFindArrivalMax(p *Mpm_Man_t) int32 {
	pmTimes := Vec_IntArray(&p.vTimes)
	var pObj *Mig_Obj_t
	var i, ArrMax int32
	// inlining of macro Mig_ManForEachCo
	for i = 0; int(i) < Vec_IntSize(&p.pMig.vCos); i++ {
		pObj = Mig_ManCo(p.pMig, i)
		ArrMax = util.Abc_MaxInt(ArrMax, pmTimes[Mig_ObjFaninId0(pObj)])
	}
	return ArrMax
}
func Mpm_ManFinalizeRound(p *Mpm_Man_t) {
	pMapRefs := Vec_IntArray(&p.vMapRefs)
	pRequired := Vec_IntArray(&p.vRequireds)
	var pObj *Mig_Obj_t
	var pCut *Mpm_Cut_t
	var pDelays []int32
	var i, iLeaf int32
	p.GloArea = 0
	p.GloEdge = 0
	p.GloRequired = Mpm_ManFindArrivalMax(p)
	if p.pPars.DelayTarget != -1 {
		p.GloRequired = util.Abc_MaxInt(p.GloRequired, p.pPars.DelayTarget)
	}
	Mpm_ManCleanMapRefs(p)
	Mpm_ManCleanRequired(p)
	// inlining of macro Mig_ManForEachObjReverse
	for i = Mig_ManObjNum(p.pMig) - 1; i >= 0; i-- {
		pObj = Mig_ManObj(p.pMig, i)
		if Mig_ObjIsCo(pObj) != 0 {
			pRequired[Mig_ObjFaninId0(pObj)] = p.GloRequired
			pMapRefs[Mig_ObjFaninId0(pObj)]++

		} else if Mig_ObjIsNode(pObj) != 0 {
			var Required int32 = pRequired[Mig_ObjId(pObj)]
			assert(Required > 0, "Required > 0")
			if pMapRefs[Mig_ObjId(pObj)] > 0 {
				pCut = Mpm_ObjCutBestP(p, pObj)
				pDelays = p.pLibLut.pLutDelays[pCut.nLeaves][:]
				// inlining of macro Mpm_CutForEachLeafId
				for i = 0; i < int32(pCut.nLeaves); i++ {
					iLeaf = util.Abc_Lit2Var(pCut.pLeaves[i])
					pRequired[iLeaf] = util.Abc_MinInt(pRequired[iLeaf], Required-pDelays[i])
					pMapRefs[iLeaf]++
				}
				p.GloArea += word(Mpm_CutGetArea(p, pCut))
				p.GloEdge += word(pCut.nLeaves)
			}

		} else if Mig_ObjIsBuf(pObj) != 0 {
		}
	}
	p.GloArea /= MPM_UNIT_AREA
}
func Mpm_ManComputeEstRefs(p *Mpm_Man_t) {
	pMapRefs := Vec_IntArray(&p.vMapRefs)
	pEstRefs := Vec_IntArray(&p.vEstRefs)
	var i int32
	assert(p.fMainRun != 0, "p.fMainRun != 0")
	//  pObj->EstRefs = (float)((2.0 * pObj->EstRefs + pObj->nRefs) / 3.0);
	for i = 0; i < Mig_ManObjNum(p.pMig); i++ {
		pEstRefs[i] = (1*pEstRefs[i] + MPM_UNIT_REFS*pMapRefs[i]) / 2
	}
}

/*
*Function*************************************************************

	Synopsis    [Cut comparison.]

	Description [Returns positive number if new one is better than old one.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_CutCompareDelay(pOld *Mpm_Uni_t, pNew *Mpm_Uni_t) int32 {
	if pOld.mTime != pNew.mTime {
		return pOld.mTime - pNew.mTime
	}
	if pOld.pCut.nLeaves != pNew.pCut.nLeaves {
		return int32(pOld.pCut.nLeaves) - int32(pNew.pCut.nLeaves)
	}
	if pOld.mArea != pNew.mArea {
		return pOld.mArea - pNew.mArea
	}
	if pOld.mEdge != pNew.mEdge {
		return pOld.mEdge - pNew.mEdge
	}
	return 0
}
func Mpm_CutCompareDelay2(pOld *Mpm_Uni_t, pNew *Mpm_Uni_t) int32 {
	if pOld.mTime != pNew.mTime {
		return pOld.mTime - pNew.mTime
	}
	if pOld.mArea != pNew.mArea {
		return pOld.mArea - pNew.mArea
	}
	if pOld.mEdge != pNew.mEdge {
		return pOld.mEdge - pNew.mEdge
	}
	if pOld.pCut.nLeaves != pNew.pCut.nLeaves {
		return int32(pOld.pCut.nLeaves) - int32(pNew.pCut.nLeaves)
	}
	return 0
}
func Mpm_CutCompareArea(pOld *Mpm_Uni_t, pNew *Mpm_Uni_t) int32 {
	if pOld.mArea != pNew.mArea {
		return pOld.mArea - pNew.mArea
	}
	if pOld.pCut.nLeaves != pNew.pCut.nLeaves {
		return int32(pOld.pCut.nLeaves) - int32(pNew.pCut.nLeaves)
	}
	if pOld.mEdge != pNew.mEdge {
		return pOld.mEdge - pNew.mEdge
	}
	if pOld.mAveRefs != pNew.mAveRefs {
		return pOld.mAveRefs - pNew.mAveRefs
	}
	if pOld.mTime != pNew.mTime {
		return pOld.mTime - pNew.mTime
	}
	return 0
}
func Mpm_CutCompareArea2(pOld *Mpm_Uni_t, pNew *Mpm_Uni_t) int32 {
	if pOld.mArea != pNew.mArea {
		return pOld.mArea - pNew.mArea
	}
	if pOld.mEdge != pNew.mEdge {
		return pOld.mEdge - pNew.mEdge
	}
	if pOld.mAveRefs != pNew.mAveRefs {
		return pOld.mAveRefs - pNew.mAveRefs
	}
	if pOld.pCut.nLeaves != pNew.pCut.nLeaves {
		return int32(pOld.pCut.nLeaves) - int32(pNew.pCut.nLeaves)
	}
	if pOld.mTime != pNew.mTime {
		return pOld.mTime - pNew.mTime
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Technology mapping experiment.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ManPrepare(p *Mpm_Man_t) {
	var pObj *Mig_Obj_t
	var i, hCut int32
	// inlining of macro Mig_ManForEachCi
	for i = 0; int(i) < Vec_IntSize(&p.pMig.vCis); i++ {
		pObj = Mig_ManCi(p.pMig, i)
		hCut = Mpm_CutCreateUnit(p, Mig_ObjId(pObj))
		Mpm_ObjSetCutBest(p, pObj, hCut)
		Mpm_ObjSetCutList(p, pObj, hCut)
	}
	// inlining of macro Mig_ManForEachCand
	// inlining of macro Mig_ManForEachObj
	for i = 0; i < Mig_ManObjNum(p.pMig); i++ {
		pObj = Mig_ManObj(p.pMig, i)
		if Mig_ObjIsCand(pObj) == 0 {
		} else {
			Mpm_ObjSetEstRef(p, pObj, MPM_UNIT_REFS*Mig_ObjRefNum(pObj))
		}
	}
}
func Mpm_ManPerformRound(p *Mpm_Man_t) {
	var pObj *Mig_Obj_t
	var clk abctime = util.Abc_Clock()
	var i int32
	// copy references
	assert(Vec_IntSize(&p.vMigRefs) == Vec_IntSize(&p.pMig.vRefs), "Vec_IntSize(&p.vMigRefs) == Vec_IntSize(&p.pMig.vRefs)")
	copy(Vec_IntArray(&p.vMigRefs), Vec_IntArray(&p.pMig.vRefs))
	// inlining of macro Mig_ManForEachCo
	for i = 0; int(i) < Vec_IntSize(&p.pMig.vCos); i++ {
		pObj = Mig_ManCo(p.pMig, i)
		Mig_ObjMigRefDec(p, Mig_ObjFanin0(pObj))
	}
	// derive cuts
	p.nCutsMerged = 0
	// inlining of macro Mig_ManForEachNode
	// inlining of macro Mig_ManForEachObj
	for i = 0; i < Mig_ManObjNum(p.pMig); i++ {
		pObj = Mig_ManObj(p.pMig, i)
		if Mig_ObjIsNode(pObj) == 0 {
		} else {
			Mpm_ManDeriveCuts(p, pObj)
		}
	}
	assert(Mig_ManCandNum(p.pMig) == mem.Mmr_StepEntryNum(p.pManCuts), "Mig_ManCandNum(p.pMig) == mem.Mmr_StepEntryNum(p.pManCuts)")
	Mpm_ManFinalizeRound(p)
	// report results
	if p.pPars.fVerbose != 0 {
		ttEntries := int32(0)
		if p.vTtMem != nil {
			ttEntries = int32(Vec_MemEntryNum(p.vTtMem))
		}
		fmt.Printf("Del =%5d.  Ar =%8d.  Edge =%8d.  Cut =%10d. Max =%8d.  Tru =%8d. Small =%6d. ", p.GloRequired, int32(p.GloArea), int32(p.GloEdge), p.nCutsMerged, mem.Mmr_StepEntryNumMax(p.pManCuts), ttEntries, p.nSmallSupp)
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	}
}
func Mpm_ManPerform(p *Mpm_Man_t) {
	if p.pPars.fMap4Cnf != 0 {
		p.pCutCmp = Mpm_CutCompareArea
		Mpm_ManPerformRound(p)

	} else {
		p.pCutCmp = Mpm_CutCompareDelay
		Mpm_ManPerformRound(p)
		if p.pPars.fOneRound != 0 {
			return
		}

		p.pCutCmp = Mpm_CutCompareDelay2
		Mpm_ManPerformRound(p)

		p.pCutCmp = Mpm_CutCompareArea
		Mpm_ManPerformRound(p)

		p.fMainRun = 1

		p.pCutCmp = Mpm_CutCompareArea
		Mpm_ManComputeEstRefs(p)
		Mpm_ManPerformRound(p)

		p.pCutCmp = Mpm_CutCompareArea2
		Mpm_ManComputeEstRefs(p)
		Mpm_ManPerformRound(p)
	}
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
