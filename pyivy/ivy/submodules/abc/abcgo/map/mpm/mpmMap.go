package mpm

import "fmt"

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

/**Function*************************************************************

  Synopsis    [Cut manipulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_CutAlloc(p *Mpm_Man_t, nLeaves int32, ppCut **Mpm_Cut_t) int32 { 
    var hHandle int32  = Mmr_StepFetch( p.pManCuts, Mpm_CutWordNum(nLeaves) );
    *ppCut      = (Mpm_Cut_t *)Mmr_StepEntry( p.pManCuts, hHandle );
    (*ppCut).nLeaves  = nLeaves;
    (*ppCut).hNext    = 0;
    (*ppCut).fUseless = 0;
    (*ppCut).fCompl   = 0;
    return hHandle;
}
func Mpm_CutCreateZero(p *Mpm_Man_t) int32 { 
    var pCut *Mpm_Cut_t;
    var hCut int32  = Mpm_CutAlloc( p, 0, &pCut );
    pCut.iFunc      = 0; // const0
    return hCut;
}
func Mpm_CutCreateUnit(p *Mpm_Man_t, Id int32) int32 { 
    var pCut *Mpm_Cut_t;
    var hCut int32  = Mpm_CutAlloc( p, 1, &pCut );
    pCut.iFunc      = Abc_Var2Lit( p.funcVar0, 0 ); // var
    pCut.pLeaves[0] = Abc_Var2Lit( Id, 0 );
    return hCut;
}
func Mpm_CutCreate(p *Mpm_Man_t, pUni *Mpm_Cut_t, ppCut **Mpm_Cut_t) int32 { 
    var hCutNew int32  = Mpm_CutAlloc( p, pUni.nLeaves, ppCut );
    (*ppCut).iFunc    = pUni.iFunc;
    (*ppCut).fCompl   = pUni.fCompl;
    (*ppCut).fUseless = pUni.fUseless;
    (*ppCut).nLeaves  = pUni.nLeaves;
    memcpy( (*ppCut).pLeaves, pUni.pLeaves, sizeof(int) * pUni.nLeaves );
    return hCutNew;
}
func Mpm_CutDup(p *Mpm_Man_t, pCut *Mpm_Cut_t, fCompl int32) int32 { 
    var pCutNew *Mpm_Cut_t;
    var hCutNew int32  = Mpm_CutAlloc( p, pCut.nLeaves, &pCutNew );
    pCutNew.iFunc    = Abc_LitNotCond( pCut.iFunc, fCompl );
    pCutNew.fUseless = pCut.fUseless;
    pCutNew.nLeaves  = pCut.nLeaves;
    memcpy( pCutNew.pLeaves, pCut.pLeaves, sizeof(int) * pCut.nLeaves );
    return hCutNew;
}
func Mpm_CutCopySet(p *Mpm_Man_t, pObj *Mig_Obj_t, fCompl int32) int32 {
    var pCut *Mpm_Cut_t;
    var hCut int32;
    var iList int32;
    var pList *int32 = &iList;
    // inlining of macro Mpm_ObjForEachCut
    for  hCut = Mpm_ObjCutList(p, pObj); hCut && (pCut = Mpm_CutFetch(p, hCut)); hCut = pCut.hNext  {
        *pList = Mpm_CutDup( p, pCut, fCompl );
        pList = &Mpm_CutFetch( p, *pList ).hNext;
    }
    *pList = 0;
    return iList;
}
func Mpm_CutPrint(pCut *Mpm_Cut_t) { 
    var i int32;
    fmt.Printf("%d : { ", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf("%d ", pCut.pLeaves[i]);
    }
    fmt.Printf("}\n");
}
func Mpm_CutPrintAll(p *Mpm_Man_t) { 
    var i int32;
    for  i = 0; i < p.nCutStore; i++  {
        fmt.Printf("%2d : ", i);
        Mpm_CutPrint( &p.pCutStore[i].pCut );
    }
}

static inline int Mpm_CutIsContained( Mpm_Man_t * p, Mpm_Cut_t * pBase, Mpm_Cut_t * pCut ) // check if pCut is contained pBase
{
    var i int32;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        if Mpm_CutFindLeaf( pBase, Abc_Lit2Var(pCut.pLeaves[i]) ) == int32(pBase.nLeaves) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Cut attibutes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_CutGetArea(p *Mpm_Man_t, pCut *Mpm_Cut_t) int32 {
    if p.pPars.fMap4Cnf {
        return MPM_UNIT_AREA * p.pDsd6[Abc_Lit2Var(pCut.iFunc)].nClauses; }
    if p.pPars.fMap4Aig {
        return MPM_UNIT_AREA * p.pDsd6[Abc_Lit2Var(pCut.iFunc)].nAnds; }
    if p.pPars.fMap4Gates {
        return MPM_UNIT_AREA * 1; }
    return p.pLibLut.pLutAreas[pCut.nLeaves];
}
func Mpm_CutGetSign(pCut *Mpm_Cut_t) word {
    var i, iLeaf int32;
    var uSign word;
    // inlining of macro Mpm_CutForEachLeafId
    for  i = 0; i < int32(pCut.nLeaves) && ((iLeaf = Abc_Lit2Var(pCut.pLeaves[i])), 1); i++  {
        uSign |= ((word)1 << (iLeaf & 0x3F));
    }
    return uSign;
}
func Mpm_CutGetArrTime(p *Mpm_Man_t, pCut *Mpm_Cut_t) int32 {
    var pmTimes *int32  = Vec_IntArray( &p.vTimes );
    var pDelays *int32  = p.pLibLut.pLutDelays[pCut.nLeaves];
    var i, iLeaf, ArrTime int32;
    // inlining of macro Mpm_CutForEachLeafId
    for  i = 0; i < int32(pCut.nLeaves) && ((iLeaf = Abc_Lit2Var(pCut.pLeaves[i])), 1); i++  {
        ArrTime = Abc_MaxInt( ArrTime, pmTimes[iLeaf] + pDelays[i] );
    }
    return ArrTime;
}
func Mpm_CutSetupInfo(p *Mpm_Man_t, pCut *Mpm_Cut_t, ArrTime int32) *Mpm_Uni_t {
    var pMigRefs *int32  = Vec_IntArray( &p.vMigRefs );
    var pMapRefs *int32  = Vec_IntArray( &p.vMapRefs );
    var pEstRefs *int32  = Vec_IntArray( &p.vEstRefs );
    var pmArea *int32    = Vec_IntArray( &p.vAreas );
    var pmEdge *int32    = Vec_IntArray( &p.vEdges );
    var i, iLeaf int32;
    var pUnit *Mpm_Uni_t  = (Mpm_Uni_t *)Vec_PtrEntryLast(&p.vFreeUnits);
    assert(&pUnit.pCut == pCut, "&pUnit.pCut == pCut");
    pUnit.mTime    = ArrTime;
    pUnit.mArea    = Mpm_CutGetArea( p, pCut );
    pUnit.mEdge    = MPM_UNIT_EDGE * pCut.nLeaves;
    pUnit.mAveRefs = 0;
    pUnit.Cost     = 0;
    pUnit.uSign    = 0;
    // inlining of macro Mpm_CutForEachLeafId
    for  i = 0; i < int32(pCut.nLeaves) && ((iLeaf = Abc_Lit2Var(pCut.pLeaves[i])), 1); i++  {
        if p.fMainRun && pMapRefs[iLeaf] == 0 { // not used in the mapping
        
            pUnit.mArea += pmArea[iLeaf];
            pUnit.mEdge += pmEdge[iLeaf];
        
        } else {
            assert(pEstRefs[iLeaf] > 0, "pEstRefs[iLeaf] > 0");
            pUnit.mArea += MPM_UNIT_REFS * pmArea[iLeaf] / pEstRefs[iLeaf];
            pUnit.mEdge += MPM_UNIT_REFS * pmEdge[iLeaf] / pEstRefs[iLeaf];
            pUnit.mAveRefs += p.fMainRun ? pMapRefs[iLeaf] : pMigRefs[iLeaf];
        }
        pUnit.uSign |= ((word)1 << (iLeaf & 0x3F));
    }
    pUnit.mAveRefs = pUnit.mAveRefs * MPM_UNIT_EDGE / Abc_MaxInt(pCut.nLeaves, 1);
    assert(pUnit.mTime <= 0x3FFFFFFF, "pUnit.mTime <= 0x3FFFFFFF");
    assert(pUnit.mArea <= 0x3FFFFFFF, "pUnit.mArea <= 0x3FFFFFFF");
    assert(pUnit.mEdge <= 0x3FFFFFFF, "pUnit.mEdge <= 0x3FFFFFFF");
    return pUnit;
}

/**Function*************************************************************

  Synopsis    [Compares cut against those present in the store.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ObjAddCutToStore(p *Mpm_Man_t, pCut *Mpm_Cut_t, ArrTime int32) int32 {
    var fEnableContainment int32  = 1;
    var pUnit, pUnitNew *Mpm_Uni_t;
    var k, iPivot, last int32;
    // create new unit
#ifdef MIG_RUNTIME
    abctime clk;
clk = Abc_Clock();
#endif
    pUnitNew = Mpm_CutSetupInfo( p, pCut, ArrTime );
#ifdef MIG_RUNTIME
p->timeEval += Abc_Clock() - clk;
#endif
    // special case when the cut store is empty
    if p.nCutStore == 0 {
        p.pCutStore[p.nCutStore++] = pUnitNew;
        Vec_PtrPop( &p.vFreeUnits );
        return 1;
    }
    // special case when the cut store is full and last cut is better than new cut
    if p.nCutStore == p.nNumCuts-1 && p.pCutCmp(pUnitNew, p.pCutStore[p.nCutStore-1]) > 0 {
        return 0; }

    // find place of the given cut in the store
    assert(p.nCutStore <= p.nNumCuts, "p.nCutStore <= p.nNumCuts");
    for  iPivot = p.nCutStore - 1; iPivot >= 0; iPivot--  {
        if p.pCutCmp(pUnitNew, p.pCutStore[iPivot]) > 0 { // iPivot-th cut is better than new cut
            break; }
    }

    if fEnableContainment != 0 {
#ifdef MIG_RUNTIME
clk = Abc_Clock();
#endif
        // filter this cut using other cuts
        for  k = 0; k <= iPivot; k++  {
            pUnit = p.pCutStore[k];
            if pUnitNew.pCut.nLeaves >= pUnit.pCut.nLeaves && 
                (pUnitNew.uSign & pUnit.uSign) == pUnit.uSign && 
                 Mpm_CutIsContained(p, &pUnitNew.pCut, &pUnit.pCut) {
#ifdef MIG_RUNTIME
p->timeCompare += Abc_Clock() - clk;
#endif
                return 0;
            }
        }
    }

    // special case when the best cut is useless while the new cut is not
    if p.pCutStore[0].pCut.fUseless && !pUnitNew.pCut.fUseless {
        iPivot = -1; }

    // add the cut to storage
    assert(pUnitNew == (Mpm_Uni_t *)Vec_PtrEntryLast(&p.vFreeUnits), "pUnitNew == (Mpm_Uni_t *)Vec_PtrEntryLast(&p.vFreeUnits)");
    Vec_PtrPop( &p.vFreeUnits );

    // insert this cut at location iPivot
    iPivot++;
    for  k = p.nCutStore++; k > iPivot; k--  {
        p.pCutStore[k] = p.pCutStore[k-1];
    }
    p.pCutStore[iPivot] = pUnitNew;

    if fEnableContainment != 0 {
        // filter other cuts using this cut
        for  k = last = iPivot+1; k < p.nCutStore; k++  {
            pUnit = p.pCutStore[k];
            if pUnitNew.pCut.nLeaves <= pUnit.pCut.nLeaves && 
                (pUnitNew.uSign & pUnit.uSign) == pUnitNew.uSign && 
                 Mpm_CutIsContained(p, &pUnit.pCut, &pUnitNew.pCut) {
                Vec_PtrPush( &p.vFreeUnits, pUnit );
                continue;
            }
            p.pCutStore[last++] = p.pCutStore[k];
        }
        p.nCutStore = last;
#ifdef MIG_RUNTIME
p->timeCompare += Abc_Clock() - clk;
#endif
    }

    // remove the last cut if too many
    if p.nCutStore == p.nNumCuts {
        Vec_PtrPush( &p.vFreeUnits, p.pCutStore[--p.nCutStore] ); }
    assert(p.nCutStore < p.nNumCuts, "p.nCutStore < p.nNumCuts");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Cut enumeration.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManExploreNewCut(p *Mpm_Man_t, pObj *Mig_Obj_t, pCut0 *Mpm_Cut_t, pCut1 *Mpm_Cut_t, pCut2 *Mpm_Cut_t, Required int32) int32 {
    var pCut *Mpm_Cut_t;
    var ArrTime int32;
#ifdef MIG_RUNTIME
abctime clk = clock();
#endif

    if pCut0.nLeaves >= pCut1.nLeaves {
        pCut = Mpm_ManMergeCuts( p, pCut0, pCut1, pCut2 );
#ifdef MIG_RUNTIME
p->timeMerge += clock() - clk;
#endif
        if pCut == nil {
            return 1; }
        if p.pPars.fUseTruth {
            Mpm_CutComputeTruth( p, pCut, pCut0, pCut1, pCut2, Mig_ObjFaninC0(pObj), Mig_ObjFaninC1(pObj), Mig_ObjFaninC2(pObj), Mig_ObjNodeType(pObj) ); 
        } else if p.pPars.fUseDsd {
            if !Mpm_CutComputeDsd6( p, pCut, pCut0, pCut1, pCut2, Mig_ObjFaninC0(pObj), Mig_ObjFaninC1(pObj), Mig_ObjFaninC2(pObj), Mig_ObjNodeType(pObj) ) {
                return 1; }
        }
    
    } else {
        pCut = Mpm_ManMergeCuts( p, pCut1, pCut0, pCut2 );
#ifdef MIG_RUNTIME
p->timeMerge += clock() - clk;
#endif
        if pCut == nil {
            return 1; }
        if p.pPars.fUseTruth {
            Mpm_CutComputeTruth( p, pCut, pCut1, pCut0, pCut2, Mig_ObjFaninC1(pObj), Mig_ObjFaninC0(pObj), 1 ^ Mig_ObjFaninC2(pObj), Mig_ObjNodeType(pObj) ); 
        } else if p.pPars.fUseDsd {
            if !Mpm_CutComputeDsd6( p, pCut, pCut1, pCut0, pCut2, Mig_ObjFaninC1(pObj), Mig_ObjFaninC0(pObj), 1 ^ Mig_ObjFaninC2(pObj), Mig_ObjNodeType(pObj) ) {
                return 1; }
        }
    }

#ifdef MIG_RUNTIME
clk = clock();
#endif
    ArrTime = Mpm_CutGetArrTime( p, pCut );
#ifdef MIG_RUNTIME
p->timeEval += clock() - clk;
#endif
    if p.fMainRun && ArrTime > Required {
        return 1; }

#ifdef MIG_RUNTIME
clk = Abc_Clock();
#endif
    Mpm_ObjAddCutToStore( p, pCut, ArrTime );
#ifdef MIG_RUNTIME
p->timeStore += Abc_Clock() - clk;
#endif

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
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ObjRecycleCuts(p *Mpm_Man_t, pObj *Mig_Obj_t) {
    var pCut *Mpm_Cut_t;
    var hCut, hNext int32;
    // inlining of macro Mpm_ObjForEachCutSafe
    for  hCut = Mpm_ObjCutList(p, pObj); hCut && (pCut = Mpm_CutFetch(p, hCut)) && ((hNext = pCut.hNext), 1); hCut = hNext  {
        Mmr_StepRecycle( p.pManCuts, hCut );
    }
    Mpm_ObjSetCutList( p, pObj, 0 );
}
func Mpm_ObjDerefFaninCuts(p *Mpm_Man_t, pObj *Mig_Obj_t) {
    var pFanin *Mig_Obj_t;
    var i int32;
    // inlining of macro Mig_ObjForEachFanin
    for  i = 0; Mig_ObjHasFanin(pObj, i) && ((pFanin) = Mig_ObjFanin(pObj, i)); i++  {
        if Mig_ObjIsNode(pFanin) && Mig_ObjMigRefDec(p, pFanin) == 0 {
            Mpm_ObjRecycleCuts( p, pFanin ); }
    }
    pFanin = Mig_ObjSibl(pObj);
    if (pFanin != nil) && Mig_ObjMigRefDec(p, pFanin) == 0 {
        Mpm_ObjRecycleCuts( p, pFanin ); }
    if Mig_ObjMigRefNum(p, pObj) == 0 {
        Mpm_ObjRecycleCuts( p, pObj ); }
}
func Mpm_ObjCollectFaninsAndSigns(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) {
    var pCut *Mpm_Cut_t;
    var hCut, nCuts int32;
    // inlining of macro Mpm_ObjForEachCut
    for  hCut = Mpm_ObjCutList(p, pObj); hCut && (pCut = Mpm_CutFetch(p, hCut)); hCut = pCut.hNext  {
        p.pCuts[i][nCuts] = pCut;
        p.pSigns[i][nCuts++] = Mpm_CutGetSign( pCut );
    }
    p.nCuts[i] = nCuts;
}
func Mpm_ObjPrepareFanins(p *Mpm_Man_t, pObj *Mig_Obj_t) {
    var pFanin *Mig_Obj_t;
    var i int32;
    // inlining of macro Mig_ObjForEachFanin
    for  i = 0; Mig_ObjHasFanin(pObj, i) && ((pFanin) = Mig_ObjFanin(pObj, i)); i++  {
        Mpm_ObjCollectFaninsAndSigns( p, pFanin, i );
    }
}
// create storage from cuts at the node
func Mpm_ObjAddChoiceCutsToStore(p *Mpm_Man_t, pRoot *Mig_Obj_t, pObj *Mig_Obj_t, ReqTime int32) {
    var pCut *Mpm_Cut_t;
    var hCut, hNext, ArrTime int32;
    var fCompl int32  = Mig_ObjPhase(pRoot) ^ Mig_ObjPhase(pObj);
    // inlining of macro Mpm_ObjForEachCutSafe
    for  hCut = Mpm_ObjCutList(p, pObj); hCut && (pCut = Mpm_CutFetch(p, hCut)) && ((hNext = pCut.hNext), 1); hCut = hNext  {
        if Abc_Lit2Var(pCut.pLeaves[0]) == Mig_ObjId(pObj) {
            continue; }
        ArrTime = Mpm_CutGetArrTime( p, pCut );
        if ArrTime > ReqTime {
            continue; }
        pCut.fCompl ^= fCompl;
        pCut = Mpm_ManMergeCuts( p, pCut, nil, nil );
        Mpm_ObjAddCutToStore( p, pCut, ArrTime );
    }
}
// create cuts at the node from storage
func Mpm_ObjTranslateCutsFromStore(p *Mpm_Man_t, pObj *Mig_Obj_t) {
    var pCut *Mpm_Cut_t  = nil;
    var pUnit *Mpm_Uni_t;
    var i int32;
    var pList *int32 = Mpm_ObjCutListP( p, pObj );
    assert(p.nCutStore > 0 && p.nCutStore <= p.nNumCuts, "p.nCutStore > 0 && p.nCutStore <= p.nNumCuts");
    assert(*pList == 0, "*pList == 0");
    // translate cuts
    for  i = 0; i < p.nCutStore; i++  {
        pUnit  = p.pCutStore[i];
        *pList = Mpm_CutCreate( p, &pUnit.pCut, &pCut );
        pList  = &pCut.hNext;
        Vec_PtrPush( &p.vFreeUnits, pUnit );
    }
    assert(Vec_PtrSize(&p.vFreeUnits) == p.nNumCuts + 1, "Vec_PtrSize(&p.vFreeUnits) == p.nNumCuts + 1");
    if p.nCutStore == 1 && pCut.nLeaves < 2 {
        *pList = 0;
    } else {
        *pList = Mpm_CutCreateUnit( p, Mig_ObjId(pObj) ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManDeriveCuts(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 {
    var pCut0, pCut1, pCut2 *Mpm_Cut_t;
    var Required int32  = Mpm_ObjRequired( p, pObj );
    var hCutBest int32  = Mpm_ObjCutBest( p, pObj );
    var c0, c1, c2 int32;
#ifdef MIG_RUNTIME
abctime clk;
#endif

    assert(Vec_PtrSize( &p.vFreeUnits ) == p.nNumCuts + 1, "Vec_PtrSize( &p.vFreeUnits ) == p.nNumCuts + 1");
    assert(Mpm_ObjCutList(p, pObj) == 0, "Mpm_ObjCutList(p, pObj) == 0");
    p.nCutStore = 0;
    if hCutBest > 0 { // cut list is assigned
    
        var pCut *Mpm_Cut_t  = Mpm_ObjCutBestP( p, pObj ); 
        var Times int32  = Mpm_CutGetArrTime( p, pCut );
        assert(pCut.hNext == 0, "pCut.hNext == 0");
        if Times > Required {
            fmt.Printf("Arrival time (%d) exceeds required time (%d) at object %d.\n", Times, Required, Mig_ObjId(pObj)); }
        if p.fMainRun {
            Mpm_ObjAddCutToStore( p, Mpm_ManMergeCuts(p, pCut, nil, nil), Times );
        } else {
            Mpm_ObjSetTime( p, pObj, Times ); }
    }
    // start storage with choice cuts
    if Mig_ManChoiceNum(p.pMig) && Mig_ObjSiblId(pObj) {
        Mpm_ObjAddChoiceCutsToStore( p, pObj, Mig_ObjSibl(pObj), Required ); }

#ifdef MIG_RUNTIME
clk = Abc_Clock();
#endif
    Mpm_ObjPrepareFanins( p, pObj );
    if Mig_ObjIsNode2(pObj) {
        // go through cut pairs
        for  c0 = 0; c0 < p.nCuts[0] && (pCut0 = p.pCuts[0][c0]); c0++  {
        for  c1 = 0; c1 < p.nCuts[1] && (pCut1 = p.pCuts[1][c1]); c1++  {
            if Abc_TtCountOnes(p.pSigns[0][c0] | p.pSigns[1][c1]) <= p.nLutSize {
                if !(Mpm_ManExploreNewCut( p, pObj, pCut0, pCut1, nil, Required ) != 0) {
                    goto finish; } }
        }
        }
    
    } else if Mig_ObjIsNode3(pObj) {
        // go through cut triples
        for  c0 = 0; c0 < p.nCuts[0] && (pCut0 = p.pCuts[0][c0]); c0++  {
        for  c1 = 0; c1 < p.nCuts[1] && (pCut1 = p.pCuts[1][c1]); c1++  {
        for  c2 = 0; c2 < p.nCuts[2] && (pCut2 = p.pCuts[2][c2]); c2++  {
            if Abc_TtCountOnes(p.pSigns[0][c0] | p.pSigns[1][c1] | p.pSigns[2][c2]) <= p.nLutSize {
                if !(Mpm_ManExploreNewCut( p, pObj, pCut0, pCut1, pCut2, Required ) != 0) {
                    goto finish; } }
        }
        }
        }
    
    } else { assert(false, "0"); }
#ifdef MIG_RUNTIME
p->timeDerive += Abc_Clock() - clk;
#endif

finish:
    // save best cut
    assert(p.nCutStore > 0, "p.nCutStore > 0");
    if p.pCutStore[0].mTime <= Required {
        var pCut *Mpm_Cut_t;
        if hCutBest != 0 {
            Mmr_StepRecycle( p.pManCuts, hCutBest ); }
        hCutBest = Mpm_CutCreate( p, &p.pCutStore[0].pCut, &pCut );
        Mpm_ObjSetCutBest( p, pObj, hCutBest );
        Mpm_ObjSetTime( p, pObj, p.pCutStore[0].mTime );
        Mpm_ObjSetArea( p, pObj, p.pCutStore[0].mArea );
        Mpm_ObjSetEdge( p, pObj, p.pCutStore[0].mEdge );
    
    } else { assert(!p.fMainRun, "!p.fMainRun"); }
    assert(hCutBest > 0, "hCutBest > 0");
    // transform internal storage into regular cuts
    Mpm_ObjTranslateCutsFromStore( p, pObj );
    // dereference fanin cuts and reference node
    Mpm_ObjDerefFaninCuts( p, pObj );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Required times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManFindArrivalMax(p *Mpm_Man_t) int32 {
    var pmTimes *int32  = Vec_IntArray( &p.vTimes );
    var pObj *Mig_Obj_t;
    var i, ArrMax int32;
    // inlining of macro Mig_ManForEachCo
    for  i = 0; (i < Vec_IntSize(&p.pMig.vCos)) && ((pObj) = Mig_ManCo(p.pMig, i)); i++  {
        ArrMax = Abc_MaxInt( ArrMax, pmTimes[ Mig_ObjFaninId0(pObj) ] );
    }
    return ArrMax;
}
func Mpm_ManFinalizeRound(p *Mpm_Man_t) {
    var pMapRefs *int32   = Vec_IntArray( &p.vMapRefs );
    var pRequired *int32  = Vec_IntArray( &p.vRequireds );
    var pObj *Mig_Obj_t;
    var pCut *Mpm_Cut_t;
    var pDelays *int32;
    var i, iLeaf int32;
    p.GloArea = 0;
    p.GloEdge = 0;
    p.GloRequired = Mpm_ManFindArrivalMax(p);
    if p.pPars.DelayTarget != -1 {
        p.GloRequired = Abc_MaxInt( p.GloRequired, p.pPars.DelayTarget ); }
    Mpm_ManCleanMapRefs( p );
    Mpm_ManCleanRequired( p );
    // inlining of macro Mig_ManForEachObjReverse
    for  p.pMig.iPage = Vec_PtrSize(&p.pMig.vPages) - 1; p.pMig.iPage >= 0 &&              ((p.pMig.pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p.pMig.vPages, p.pMig.iPage)); p.pMig.iPage--  {         for  pObj = (p.pMig.iPage == Vec_PtrSize(&p.pMig.vPages) - 1) ?                    Mig_ManObj(p.pMig, Mig_ManObjNum(p.pMig)-1) :  p.pMig.pPage + MIG_MASK;                   pObj - p.pMig.pPage >= 0; pObj--  {
        if Mig_ObjIsCo(pObj) {
            pRequired[Mig_ObjFaninId0(pObj)] = p.GloRequired;
            pMapRefs [Mig_ObjFaninId0(pObj)]++;
        
        } else if Mig_ObjIsNode(pObj) {
            var Required int32  = pRequired[Mig_ObjId(pObj)];
            assert(Required > 0, "Required > 0");
            if pMapRefs[Mig_ObjId(pObj)] > 0 {
                pCut = Mpm_ObjCutBestP( p, pObj );
                pDelays = p.pLibLut.pLutDelays[pCut.nLeaves];
                // inlining of macro Mpm_CutForEachLeafId
                for  i = 0; i < int32(pCut.nLeaves) && ((iLeaf = Abc_Lit2Var(pCut.pLeaves[i])), 1); i++  {
                    pRequired[iLeaf] = Abc_MinInt( pRequired[iLeaf], Required - pDelays[i] );
                    pMapRefs [iLeaf]++;
                }
                p.GloArea += Mpm_CutGetArea( p, pCut );
                p.GloEdge += pCut.nLeaves;
            }
        
        } else if Mig_ObjIsBuf(pObj) {
        }
    }
    }
    p.GloArea /= MPM_UNIT_AREA;
}
func Mpm_ManComputeEstRefs(p *Mpm_Man_t) {
    var pMapRefs *int32   = Vec_IntArray( &p.vMapRefs );
    var pEstRefs *int32   = Vec_IntArray( &p.vEstRefs );
    var i int32;
    assert(p.fMainRun, "p.fMainRun");
//  pObj->EstRefs = (float)((2.0 * pObj->EstRefs + pObj->nRefs) / 3.0);
    for  i = 0; i < Mig_ManObjNum(p.pMig); i++  {
        pEstRefs[i] = (1 * pEstRefs[i] + MPM_UNIT_REFS * pMapRefs[i]) / 2;
    }
}

/**Function*************************************************************

  Synopsis    [Cut comparison.]

  Description [Returns positive number if new one is better than old one.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_CutCompareDelay(pOld *Mpm_Uni_t, pNew *Mpm_Uni_t) int32 {
    if pOld.mTime        != pNew.mTime { return pOld.mTime        - pNew.mTime; }
    if pOld.pCut.nLeaves != pNew.pCut.nLeaves { return pOld.pCut.nLeaves - pNew.pCut.nLeaves; }
    if pOld.mArea        != pNew.mArea { return pOld.mArea        - pNew.mArea; }
    if pOld.mEdge        != pNew.mEdge { return pOld.mEdge        - pNew.mEdge; }
    return 0;
}
func Mpm_CutCompareDelay2(pOld *Mpm_Uni_t, pNew *Mpm_Uni_t) int32 {
    if pOld.mTime        != pNew.mTime { return pOld.mTime        - pNew.mTime; }
    if pOld.mArea        != pNew.mArea { return pOld.mArea        - pNew.mArea; }
    if pOld.mEdge        != pNew.mEdge { return pOld.mEdge        - pNew.mEdge; }
    if pOld.pCut.nLeaves != pNew.pCut.nLeaves { return pOld.pCut.nLeaves - pNew.pCut.nLeaves; }
    return 0;
}
func Mpm_CutCompareArea(pOld *Mpm_Uni_t, pNew *Mpm_Uni_t) int32 {
    if pOld.mArea        != pNew.mArea { return pOld.mArea        - pNew.mArea; }
    if pOld.pCut.nLeaves != pNew.pCut.nLeaves { return pOld.pCut.nLeaves - pNew.pCut.nLeaves; }
    if pOld.mEdge        != pNew.mEdge { return pOld.mEdge        - pNew.mEdge; }
    if pOld.mAveRefs     != pNew.mAveRefs { return pOld.mAveRefs     - pNew.mAveRefs; }
    if pOld.mTime        != pNew.mTime { return pOld.mTime        - pNew.mTime; }
    return 0;
}
func Mpm_CutCompareArea2(pOld *Mpm_Uni_t, pNew *Mpm_Uni_t) int32 {
    if pOld.mArea        != pNew.mArea { return pOld.mArea        - pNew.mArea; }
    if pOld.mEdge        != pNew.mEdge { return pOld.mEdge        - pNew.mEdge; }
    if pOld.mAveRefs     != pNew.mAveRefs { return pOld.mAveRefs     - pNew.mAveRefs; }
    if pOld.pCut.nLeaves != pNew.pCut.nLeaves { return pOld.pCut.nLeaves - pNew.pCut.nLeaves; }
    if pOld.mTime        != pNew.mTime { return pOld.mTime        - pNew.mTime; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Technology mapping experiment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManPrepare(p *Mpm_Man_t) {
    var pObj *Mig_Obj_t;
    var i, hCut int32;
    // inlining of macro Mig_ManForEachCi
    for  i = 0; (i < Vec_IntSize(&p.pMig.vCis)) && ((pObj) = Mig_ManCi(p.pMig, i)); i++  {
        hCut = Mpm_CutCreateUnit( p, Mig_ObjId(pObj) );
        Mpm_ObjSetCutBest( p, pObj, hCut );
        Mpm_ObjSetCutList( p, pObj, hCut );
    }
    // inlining of macro Mig_ManForEachCand
    // inlining of macro Mig_ManForEachObj
    for  p.pMig.iPage = 0; p.pMig.iPage < Vec_PtrSize(&p.pMig.vPages) &&                   ((p.pMig.pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p.pMig.vPages, p.pMig.iPage)); p.pMig.iPage++  {         for  pObj = p.pMig.pPage; !Mig_ObjIsNone(pObj); pObj++  { if !Mig_ObjIsCand(pObj) {} else {
        Mpm_ObjSetEstRef( p, pObj, MPM_UNIT_REFS * Mig_ObjRefNum(pObj) ); }
    }
    }
}
func Mpm_ManPerformRound(p *Mpm_Man_t) {
    var pObj *Mig_Obj_t;
    var clk abctime  = Abc_Clock();
    var i int32;
    // copy references
    assert(Vec_IntSize(&p.vMigRefs) == Vec_IntSize(&p.pMig.vRefs), "Vec_IntSize(&p.vMigRefs) == Vec_IntSize(&p.pMig.vRefs)");
    memcpy( Vec_IntArray(&p.vMigRefs), Vec_IntArray(&p.pMig.vRefs), sizeof(int) * Mig_ManObjNum(p.pMig) );
    // inlining of macro Mig_ManForEachCo
    for  i = 0; (i < Vec_IntSize(&p.pMig.vCos)) && ((pObj) = Mig_ManCo(p.pMig, i)); i++  {
        Mig_ObjMigRefDec( p, Mig_ObjFanin0(pObj) );
    }
    // derive cuts
    p.nCutsMerged = 0;
    // inlining of macro Mig_ManForEachNode
    // inlining of macro Mig_ManForEachObj
    for  p.pMig.iPage = 0; p.pMig.iPage < Vec_PtrSize(&p.pMig.vPages) &&                   ((p.pMig.pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p.pMig.vPages, p.pMig.iPage)); p.pMig.iPage++  {         for  pObj = p.pMig.pPage; !Mig_ObjIsNone(pObj); pObj++  { if !Mig_ObjIsNode(pObj) {} else {
        Mpm_ManDeriveCuts( p, pObj ); }
    }
    }
    assert(Mig_ManCandNum(p.pMig) == p.pManCuts.nEntries, "Mig_ManCandNum(p.pMig) == p.pManCuts.nEntries");
    Mpm_ManFinalizeRound( p );
    // report results
    if p.pPars.fVerbose {
    fmt.Printf("Del =%5d.  Ar =%8d.  Edge =%8d.  Cut =%10d. Max =%8d.  Tru =%8d. Small =%6d. ", p.GloRequired, int32(p.GloArea), int32(p.GloEdge), p.nCutsMerged, p.pManCuts.nEntriesMax, p.vTtMem ? p.vTtMem.nEntries : 0, p.nSmallSupp);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
}
func Mpm_ManPerform(p *Mpm_Man_t) {
    if p.pPars.fMap4Cnf {
        p.pCutCmp = Mpm_CutCompareArea;
        Mpm_ManPerformRound( p );   
    
    } else {
        p.pCutCmp = Mpm_CutCompareDelay;
        Mpm_ManPerformRound( p );
        if p.pPars.fOneRound {
            return; }
    
        p.pCutCmp = Mpm_CutCompareDelay2;
        Mpm_ManPerformRound( p );
    
        p.pCutCmp = Mpm_CutCompareArea;
        Mpm_ManPerformRound( p );   

        p.fMainRun = 1;

        p.pCutCmp = Mpm_CutCompareArea;
        Mpm_ManComputeEstRefs( p );
        Mpm_ManPerformRound( p );

        p.pCutCmp = Mpm_CutCompareArea2;
        Mpm_ManComputeEstRefs( p );
        Mpm_ManPerformRound( p );
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

