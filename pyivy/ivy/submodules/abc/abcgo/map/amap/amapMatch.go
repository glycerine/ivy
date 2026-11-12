package amap

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [amapMatch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapMatch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "amapInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Duplicates the cut using new memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManDupCut(p *Amap_Man_t, pCut *Amap_Cut_t) *Amap_Cut_t {
    var pNew *Amap_Cut_t;
    var nBytes int32  = sizeof(Amap_Cut_t) + sizeof(int) * pCut.nFans;
    pNew = (Amap_Cut_t *)Aig_MmFlexEntryFetch( p.pMemCutBest, nBytes );
    memcpy( pNew, pCut, (size_t)nBytes );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Starts the match with cut and set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMatchStart(p *Amap_Mat_t, pCut *Amap_Cut_t, pSet *Amap_Set_t) {
    memset( p, 0, sizeof(Amap_Mat_t) );
    p.pCut = pCut;
    p.pSet = pSet;
}

/**Function*************************************************************

  Synopsis    [Cleans reference counters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManCleanRefs(p *Amap_Man_t) {
    var pObj *Amap_Obj_t;
    var i int32;
    // inlining of macro Amap_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.nFouts[0] = pObj.nFouts[1] = 0; }
    }
}

/**Function*************************************************************

  Synopsis    [Computes delay.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMaxDelay(p *Amap_Man_t) float32 {
    var pObj *Amap_Obj_t;
    var Delay float32  = 0.0;
    var i int32;
    // inlining of macro Amap_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Delay = Abc_MaxInt( Delay, Amap_ObjFanin0(p,pObj).Best.Delay );
    }
    return Delay;
}

/**Function*************************************************************

  Synopsis    [Cleans reference counters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManCleanData(p *Amap_Man_t) {
    var pObj *Amap_Obj_t;
    var i int32;
//    Amap_ManForEachNode( p, pObj, i )
//        ABC_FREE( pObj->pData );
    // inlining of macro Amap_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = nil; }
    }
}

/**Function*************************************************************

  Synopsis    [Compute nodes used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManComputeMapping_rec(p *Amap_Man_t, pObj *Amap_Obj_t, fCompl int32) float32 {
    var pM *Amap_Mat_t  = &pObj.Best;
    var pFanin *Amap_Obj_t;
    var pGate *Amap_Gat_t;
    var i, iFanin, fComplFanin int32;
    var Area float32;
    if pObj.nFouts[fCompl]++ + pObj.nFouts[!fCompl] > 0 {
        return 0.0; }
    if Amap_ObjIsPi(pObj) || Amap_ObjIsConst1(pObj) {
        return 0.0; }
    pGate = Amap_LibGate( p.pLib, pM.pSet.iGate );
    assert(pGate.nPins == pM.pCut.nFans, "pGate.nPins == pM.pCut.nFans");
    Area = pGate.dArea;
    for  i = 0; i < int32(pGate.nPins); i++  {
        iFanin = Abc_Lit2Var( pM.pSet.Ins[i] );
        pFanin = Amap_ManObj( p, Abc_Lit2Var(pM.pCut.Fans[iFanin]) );
        fComplFanin = Abc_LitIsCompl( pM.pSet.Ins[i] ) ^ Abc_LitIsCompl( pM.pCut.Fans[iFanin] );
        Area += Amap_ManComputeMapping_rec( p, pFanin, fComplFanin );
    }
    return Area;
}

/**Function*************************************************************

  Synopsis    [Compute nodes used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManComputeMapping(p *Amap_Man_t) float32 {
    var pObj *Amap_Obj_t;
    var Area float32  = 0.0;
    var i int32;
    Amap_ManCleanRefs( p );
    // inlining of macro Amap_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Area += Amap_ManComputeMapping_rec( p, Amap_ObjFanin0(p, pObj), Amap_ObjFaninC0(pObj) );
    }
    return Area;
}

/**Function*************************************************************

  Synopsis    [Counts the number of inverters to be added.]

  Description [Should be called after mapping has been set.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManCountInverters(p *Amap_Man_t) int32 {
    var pObj *Amap_Obj_t;
    var i, Counter int32;
    // inlining of macro Amap_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        Counter += int32((pObj.nFouts[!pObj.fPolar] > 0)); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Compare two matches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_CutCompareDelay(p *Amap_Man_t, pM0 *Amap_Mat_t, pM1 *Amap_Mat_t) int32 {
    // compare delay
    if pM0.Delay < pM1.Delay - p.pPars.fEpsilon {
        return -1; }
    if pM0.Delay > pM1.Delay + p.pPars.fEpsilon {
        return 1; }

    // compare area flows
    if pM0.Area < pM1.Area - p.pPars.fEpsilon {
        return -1; }
    if pM0.Area > pM1.Area + p.pPars.fEpsilon {
        return 1; }

    // compare average fanouts
    if pM0.AveFan > pM1.AveFan - p.pPars.fEpsilon {
        return -1; }
    if pM0.AveFan < pM1.AveFan + p.pPars.fEpsilon {
        return 1; }
    return 1;
}
func Amap_CutCompareArea(p *Amap_Man_t, pM0 *Amap_Mat_t, pM1 *Amap_Mat_t) int32 {
    // compare area flows
    if pM0.Area < pM1.Area - p.pPars.fEpsilon {
        return -1; }
    if pM0.Area > pM1.Area + p.pPars.fEpsilon {
        return 1; }

    // compare average fanouts
    if pM0.AveFan > pM1.AveFan - p.pPars.fEpsilon {
        return -1; }
    if pM0.AveFan < pM1.AveFan + p.pPars.fEpsilon {
        return 1; }

    // compare delay
    if pM0.Delay < pM1.Delay - p.pPars.fEpsilon {
        return -1; }
    if pM0.Delay > pM1.Delay + p.pPars.fEpsilon {
        return 1; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Counts area while dereferencing the match.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_CutAreaDeref(p *Amap_Man_t, pM *Amap_Mat_t) float32 {
    var pFanin *Amap_Obj_t;
    var i, fCompl int32;
    var Area float32  = Amap_LibGate( p.pLib, pM.pSet.iGate ).dArea;
    // inlining of macro Amap_MatchForEachFaninCompl
    for  i = 0; i < int32((pM).pCut.nFans) &&                                     ((pFanin = Amap_ManObj((p), Abc_Lit2Var((pM).pCut.Fans[Abc_Lit2Var((pM).pSet.Ins[i])]))), 1) &&         ((fCompl = Abc_LitIsCompl((pM).pSet.Ins[i]) ^ Abc_LitIsCompl((pM).pCut.Fans[Abc_Lit2Var((pM).pSet.Ins[i])])), 1);         i++  {
        assert(Amap_ObjRefsTotal(pFanin) > 0, "Amap_ObjRefsTotal(pFanin) > 0");
        if int32(pFanin.fPolar) != fCompl && pFanin.nFouts[fCompl] == 1 {
            Area += p.fAreaInv; }
        if --pFanin.nFouts[fCompl] + pFanin.nFouts[!fCompl] == 0 && Amap_ObjIsNode(pFanin) {
            Area += Amap_CutAreaDeref( p, &pFanin.Best ); }
    }
    return Area;
}

/**Function*************************************************************

  Synopsis    [Counts area while referencing the match.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_CutAreaRef2(p *Amap_Man_t, pM *Amap_Mat_t, vTemp *Vec_Ptr_t, Limit int32) float32 {
    var pFanin *Amap_Obj_t;
    var i, fCompl int32;
    var Area float32  = Amap_LibGate( p.pLib, pM.pSet.iGate ).dArea;
    if Limit == 0 { return Area; }
    // inlining of macro Amap_MatchForEachFaninCompl
    for  i = 0; i < int32((pM).pCut.nFans) &&                                     ((pFanin = Amap_ManObj((p), Abc_Lit2Var((pM).pCut.Fans[Abc_Lit2Var((pM).pSet.Ins[i])]))), 1) &&         ((fCompl = Abc_LitIsCompl((pM).pSet.Ins[i]) ^ Abc_LitIsCompl((pM).pCut.Fans[Abc_Lit2Var((pM).pSet.Ins[i])])), 1);         i++  {
        Vec_PtrPush( vTemp, pFanin.nFouts + fCompl );
        assert(Amap_ObjRefsTotal(pFanin) >= 0, "Amap_ObjRefsTotal(pFanin) >= 0");
        if int32(pFanin.fPolar) != fCompl && pFanin.nFouts[fCompl] == 0 {
            Area += p.fAreaInv; }
        if pFanin.nFouts[fCompl]++ + pFanin.nFouts[!fCompl] == 0 && Amap_ObjIsNode(pFanin) {
            Area += Amap_CutAreaRef2( p, &pFanin.Best, vTemp, Limit-1 ); }
    }
    return Area;
}

/**Function*************************************************************

  Synopsis    [Derives area of the match for a non-referenced node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_CutAreaDerefed2(p *Amap_Man_t, pNode *Amap_Obj_t, pM *Amap_Mat_t) float32 {
    var nRecurLevels int32  = 8;
    var fComplNew int32;
    var i int32;
    var pInt *int32;
    var aResult float32;
    Vec_PtrClear( p.vTempP );
    aResult = Amap_CutAreaRef2( p, pM, p.vTempP, nRecurLevels );
    //Amap_CutAreaDeref( p, pM );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTempP)) && (((pInt) = (int *)Vec_PtrEntry(p.vTempP, i)), 1); i++  {
        (*pInt)--;
    }
    // if node is needed in another polarity, add inverter
    fComplNew = pM.pCut.fInv ^ pM.pSet.fInv;
    if pNode.nFouts[fComplNew] == 0 && pNode.nFouts[!fComplNew] > 0 {
        aResult += p.fAreaInv; }
    return aResult;
}

/**Function*************************************************************

  Synopsis    [Counts area while referencing the match.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_CutAreaRef(p *Amap_Man_t, pM *Amap_Mat_t) float32 {
    var pFanin *Amap_Obj_t;
    var i, fCompl int32;
    var Area float32  = Amap_LibGate( p.pLib, pM.pSet.iGate ).dArea;
    // inlining of macro Amap_MatchForEachFaninCompl
    for  i = 0; i < int32((pM).pCut.nFans) &&                                     ((pFanin = Amap_ManObj((p), Abc_Lit2Var((pM).pCut.Fans[Abc_Lit2Var((pM).pSet.Ins[i])]))), 1) &&         ((fCompl = Abc_LitIsCompl((pM).pSet.Ins[i]) ^ Abc_LitIsCompl((pM).pCut.Fans[Abc_Lit2Var((pM).pSet.Ins[i])])), 1);         i++  {
        assert(Amap_ObjRefsTotal(pFanin) >= 0, "Amap_ObjRefsTotal(pFanin) >= 0");
        if int32(pFanin.fPolar) != fCompl && pFanin.nFouts[fCompl] == 0 {
            Area += p.fAreaInv; }
        if pFanin.nFouts[fCompl]++ + pFanin.nFouts[!fCompl] == 0 && Amap_ObjIsNode(pFanin) {
            Area += Amap_CutAreaRef( p, &pFanin.Best ); }
    }
    return Area;
}

/**Function*************************************************************

  Synopsis    [Derives area of the match for a non-referenced node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_CutAreaDerefed(p *Amap_Man_t, pNode *Amap_Obj_t, pM *Amap_Mat_t) float32 {
    var aResult, aResult2 float32;
    var fComplNew int32;
    aResult2 = Amap_CutAreaRef( p, pM );
    aResult  = Amap_CutAreaDeref( p, pM );
    assert(aResult > aResult2 - p.fEpsilonInternal, "aResult > aResult2 - p.fEpsilonInternal");
    assert(aResult < aResult2 + p.fEpsilonInternal, "aResult < aResult2 + p.fEpsilonInternal");
    // if node is needed in another polarity, add inverter
    fComplNew = pM.pCut.fInv ^ pM.pSet.fInv;
    if pNode.nFouts[fComplNew] == 0 && pNode.nFouts[!fComplNew] > 0 {
        aResult += p.fAreaInv; }
    return aResult;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_CutAreaTest(p *Amap_Man_t, pNode *Amap_Obj_t) {
    var aResult, aResult2 float32;
    if Amap_ObjRefsTotal(pNode) == 0 {
        aResult2 = Amap_CutAreaRef( p, &pNode.Best );
        aResult  = Amap_CutAreaDeref( p, &pNode.Best );
        assert(aResult > aResult2 - p.fEpsilonInternal, "aResult > aResult2 - p.fEpsilonInternal");
        assert(aResult < aResult2 + p.fEpsilonInternal, "aResult < aResult2 + p.fEpsilonInternal");
    
    } else {
        aResult  = Amap_CutAreaDeref( p, &pNode.Best );
        aResult2 = Amap_CutAreaRef( p, &pNode.Best );
        assert(aResult > aResult2 - p.fEpsilonInternal, "aResult > aResult2 - p.fEpsilonInternal");
        assert(aResult < aResult2 + p.fEpsilonInternal, "aResult < aResult2 + p.fEpsilonInternal");
    }
}

/**Function*************************************************************

  Synopsis    [Derives parameters for the match.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMatchGetFlows(p *Amap_Man_t, pM *Amap_Mat_t) {
    var pMFanin *Amap_Mat_t;
    var pFanin *Amap_Obj_t;
    var pGate *Amap_Gat_t;
    var AddOn float32; 
    var i int32;
    pGate = Amap_LibGate( p.pLib, pM.pSet.iGate );
    assert(pGate.nPins == pM.pCut.nFans, "pGate.nPins == pM.pCut.nFans");
    assert(pM.Area == 0.0, "pM.Area == 0.0");
    pM.Area = pGate.dArea;
    pM.AveFan = 0.0;
    pM.Delay = 0.0;
    // inlining of macro Amap_MatchForEachFanin
    for  i = 0; i < int32((pM).pCut.nFans) &&                                     ((pFanin = Amap_ManObj((p), Abc_Lit2Var((pM).pCut.Fans[Abc_Lit2Var((pM).pSet.Ins[i])]))), 1);         i++  {
        pMFanin = &pFanin.Best;
        pM.Delay = Abc_MaxInt( pM.Delay, pMFanin.Delay );
        pM.AveFan += Amap_ObjRefsTotal(pFanin);
//        if ( Amap_ObjRefsTotal(pFanin) == 0 )
//            pM->Area += pMFanin->Area;
//        else
//            pM->Area += pMFanin->Area / pFanin->EstRefs;
        if Amap_ObjRefsTotal(pFanin) == 0 {
            AddOn = pMFanin.Area;
        } else {
            AddOn = pMFanin.Area / pFanin.EstRefs;
        } 
        if pM.Area >= float32(1e32) || AddOn >= float32(1e32) {
            pM.Area = float32(1e32);
        } else { 
            pM.Area += AddOn; }
    }
    pM.AveFan /= pGate.nPins;
    pM.Delay += 1.0;
}

/**Function*************************************************************

  Synopsis    [Derives parameters for the match.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMatchGetExacts(p *Amap_Man_t, pNode *Amap_Obj_t, pM *Amap_Mat_t) {
    var pMFanin *Amap_Mat_t;
    var pFanin *Amap_Obj_t;
    var pGate *Amap_Gat_t;
    var i int32;
    pGate = Amap_LibGate( p.pLib, pM.pSet.iGate );
    assert(pGate.nPins == pM.pCut.nFans, "pGate.nPins == pM.pCut.nFans");
    assert(pM.Area == 0.0, "pM.Area == 0.0");
    pM.AveFan = 0.0;
    pM.Delay = 0.0;
    // inlining of macro Amap_MatchForEachFanin
    for  i = 0; i < int32((pM).pCut.nFans) &&                                     ((pFanin = Amap_ManObj((p), Abc_Lit2Var((pM).pCut.Fans[Abc_Lit2Var((pM).pSet.Ins[i])]))), 1);         i++  {
        pMFanin = &pFanin.Best;
        pM.Delay = Abc_MaxInt( pM.Delay, pMFanin.Delay );
        pM.AveFan += Amap_ObjRefsTotal(pFanin);
    }
    pM.AveFan /= pGate.nPins;
    pM.Delay += 1.0;
    //pM->Area = Amap_CutAreaDerefed( p, pNode, pM );
    pM.Area = Amap_CutAreaDerefed2( p, pNode, pM );
}

/**Function*************************************************************

  Synopsis    [Computes the best match at each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMatchNode(p *Amap_Man_t, pNode *Amap_Obj_t, fFlow int32, fRefs int32) {
    var fVerbose int32; //(pNode->Level == 2 || pNode->Level == 4);
    var fVeryVerbose int32  = fVerbose;

    var MA Amap_Mat_t ;
    var MD Amap_Mat_t ;
    var M Amap_Mat_t ;
    var pMBestA *Amap_Mat_t = &MA;
    var pMBestD *Amap_Mat_t = &MD;
    var pMThis *Amap_Mat_t = &M;
    var pMBest *Amap_Mat_t;
    var pCut *Amap_Cut_t;
    var pSet *Amap_Set_t;
    var pNod *Amap_Nod_t;
    var i int32;

    if fRefs != 0 {
        pNode.EstRefs = float32(((2.0 * pNode.EstRefs + Amap_ObjRefsTotal(pNode)) / 3.0));
    } else {
        pNode.EstRefs = float32(pNode.nRefs); }
    if (fRefs != 0) && Amap_ObjRefsTotal(pNode) > 0 {
        Amap_CutAreaDeref( p, &pNode.Best ); }

    if fVerbose != 0 {
        fmt.Printf("\nNode %d (%d)\n", pNode.Id, pNode.Level); }

    pMBestA.pCut = pMBestD.pCut = nil;
    // inlining of macro Amap_NodeForEachCut
    for  i = 0, pCut = (Amap_Cut_t *)pNode.pData; i < int32(pNode.nCuts);         i++, pCut = Amap_ManCutNext(pCut)  {
        if pCut.iMat == 0 {
            continue; }
        pNod = Amap_LibNod( p.pLib, pCut.iMat );
        // inlining of macro Amap_LibNodeForEachSet
        for  pSet = pNod.pSets; (pSet != nil); pSet = pSet.pNext  {
            Amap_ManMatchStart( pMThis, pCut, pSet );
            if fFlow != 0 {
                Amap_ManMatchGetFlows( p, pMThis );
            } else {
                Amap_ManMatchGetExacts( p, pNode, pMThis ); }
            if pMBestD.pCut == nil || Amap_CutCompareDelay(p, pMBestD, pMThis) == 1 {
                *pMBestD = *pMThis; }
            if pMBestA.pCut == nil || Amap_CutCompareArea(p, pMBestA, pMThis) == 1 {
                *pMBestA = *pMThis; }

            if fVeryVerbose != 0 {
                fmt.Printf("Cut %2d (%d) :  ", i, pCut.nFans);
                fmt.Printf("Gate %10s  ", Amap_LibGate(p.pLib, pMThis.pSet.iGate).pName);
                fmt.Printf("%s  ", pMThis.pSet.fInv ? "inv" : "   ");
                fmt.Printf("Delay %5.2f  ", pMThis.Delay);
                fmt.Printf("Area %5.2f  ", pMThis.Area);
                fmt.Printf("\n");
            }
        }
    }

    if Abc_AbsFloat(pMBestA.Area - pMBestD.Area) / pMBestD.Area >= p.pPars.fADratio * Abc_AbsFloat(pMBestA.Delay - pMBestD.Delay) / pMBestA.Delay {
        pMBest = pMBestA;
    } else {
        pMBest = pMBestD; }

    if fVerbose != 0 {
        fmt.Printf("BEST MATCHA:  ");
        fmt.Printf("Gate %10s  ", Amap_LibGate(p.pLib, pMBestA.pSet.iGate).pName);
        fmt.Printf("%s  ", pMBestA.pSet.fInv ? "inv" : "   ");
        fmt.Printf("Delay %5.2f  ", pMBestA.Delay);
        fmt.Printf("Area %5.2f  ", pMBestA.Area);
        fmt.Printf("\n"); 

        fmt.Printf("BEST MATCHD:  ");
        fmt.Printf("Gate %10s  ", Amap_LibGate(p.pLib, pMBestD.pSet.iGate).pName);
        fmt.Printf("%s  ", pMBestD.pSet.fInv ? "inv" : "   ");
        fmt.Printf("Delay %5.2f  ", pMBestD.Delay);
        fmt.Printf("Area %5.2f  ", pMBestD.Area);
        fmt.Printf("\n"); 

        fmt.Printf("BEST MATCH :  ");
        fmt.Printf("Gate %10s  ", Amap_LibGate(p.pLib, pMBest.pSet.iGate).pName);
        fmt.Printf("%s  ", pMBest.pSet.fInv ? "inv" : "   ");
        fmt.Printf("Delay %5.2f  ", pMBest.Delay);
        fmt.Printf("Area %5.2f  ", pMBest.Area);
        fmt.Printf("\n"); 
    }

    pNode.fPolar = pMBest.pCut.fInv ^ pMBest.pSet.fInv;
    pNode.Best = *pMBest;
    pNode.Best.pCut = Amap_ManDupCut( p, pNode.Best.pCut );
    if (fRefs != 0) && Amap_ObjRefsTotal(pNode) > 0 {
        Amap_CutAreaRef( p, &pNode.Best ); }
}

/**Function*************************************************************

  Synopsis    [Performs one round of mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMatch(p *Amap_Man_t, fFlow int32, fRefs int32) {
    var pMemOld *Aig_MmFlex_t;
    var pObj *Amap_Obj_t;
    var Area float32;
    var i, nInvs int32;
    var clk abctime  = Abc_Clock();
    pMemOld = p.pMemCutBest;
    p.pMemCutBest = Aig_MmFlexStart();
    // inlining of macro Amap_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Amap_ObjIsNode(pObj) {} else
        if pObj.pData {
            Amap_ManMatchNode( p, pObj, fFlow, fRefs ); }
    }
    Aig_MmFlexStop( pMemOld, 0 );
    Area = Amap_ManComputeMapping( p );
    nInvs = Amap_ManCountInverters( p );
if p.pPars.fVerbose {
    fmt.Printf("Area =%9.2f. Gate =%9.2f. Inv =%9.2f. (%6d.) Delay =%6.2f. ", Area + nInvs * p.fAreaInv, Area, nInvs * p.fAreaInv, nInvs, Amap_ManMaxDelay(p));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
    // test procedures
//    Amap_ManForEachNode( p, pObj, i )
//        Amap_CutAreaTest( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Performs mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMap(p *Amap_Man_t) {
    var i int32;
    Amap_ManMerge( p );
    for  i = 0; i < p.pPars.nIterFlow; i++  {
        Amap_ManMatch( p, 1, i>0 );
    }
    for  i = 0; i < p.pPars.nIterArea; i++  {
        Amap_ManMatch( p, 0, p.pPars.nIterFlow>0||i>0 );
    }
/*
    for ( i = 0; i < p->pPars->nIterFlow; i++ )
        Amap_ManMatch( p, 1, 1 );
    for ( i = 0; i < p->pPars->nIterArea; i++ )
        Amap_ManMatch( p, 0, 1 );
*/
    Amap_ManCleanData( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

