package cgt

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [cgtMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Clock gating package.]

  Synopsis    [Decide what gate to use for what flop.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cgtMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cgtInt.h"
//#include "proof/ssw/sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Ssw_SmlCheckXorImplication( Ssw_Sml_t * p, Aig_Obj_t * pObjLi, Aig_Obj_t * pObjLo, Aig_Obj_t * pCand );

// extern int Ssw_SmlCountXorImplication( Ssw_Sml_t * p, Aig_Obj_t * pObjLi, Aig_Obj_t * pObjLo, Aig_Obj_t * pCand );

// extern int Ssw_SmlCountEqual( Ssw_Sml_t * p, Aig_Obj_t * pObjLi, Aig_Obj_t * pObjLo );

// extern int Ssw_SmlNodeCountOnesReal( Ssw_Sml_t * p, Aig_Obj_t * pObj );

// extern int Ssw_SmlNodeCountOnesRealVec( Ssw_Sml_t * p, Vec_Ptr_t * vObjs );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collects POs in the transitive fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManCollectFanoutPos_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, vFanout *Vec_Ptr_t) {
    var pFanout *Aig_Obj_t;
    var f int32;
    var iFanout int32 = -1;
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    if Aig_ObjIsCo(pObj) {
        Vec_PtrPush( vFanout, pObj );
        return;
    }
    // inlining of macro Aig_ObjForEachFanout
    for  assert(pAig.pFanData, "pAig.pFanData"), f = 0; (f < int32((pObj).nRefs)) &&                         (((iFanout) = f? Aig_ObjFanoutNext(pAig, iFanout) : Aig_ObjFanout0Int(pAig, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(pAig, iFanout>>1)), 1); f++  {
        Cgt_ManCollectFanoutPos_rec( pAig, pFanout, vFanout );
    }
}

/**Function*************************************************************

  Synopsis    [Collects POs in the transitive fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManCollectFanoutPos(pAig *Aig_Man_t, pObj *Aig_Obj_t, vFanout *Vec_Ptr_t) {
    Vec_PtrClear( vFanout );
    Aig_ManIncrementTravId( pAig );
    Cgt_ManCollectFanoutPos_rec( pAig, pObj, vFanout );
}

/**Function*************************************************************

  Synopsis    [Checks if all PO fanouts can be gated by this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManCheckGateComplete(pAig *Aig_Man_t, vGatesAll *Vec_Vec_t, pGate *Aig_Obj_t, vFanout *Vec_Ptr_t) int32 {
    var vGates *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanout)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vFanout, i)), 1); i++  {
        if Saig_ObjIsPo(pAig, pObj) {
            return 0; }
        vGates = Vec_VecEntry( vGatesAll, Aig_ObjCioId(pObj) - Saig_ManPoNum(pAig) );
        if Vec_PtrFind( vGates, pGate ) == -1 {
            return 0; }            
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes the set of complete clock gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManCompleteGates(pAig *Aig_Man_t, vGatesAll *Vec_Vec_t, nOdcMax int32, fVerbose int32) *Vec_Ptr_t {
    var vFanout, vGatesFull *Vec_Ptr_t;
    var pGate, pGateR *Aig_Obj_t;
    var i, k int32;
    vFanout    = Vec_PtrAlloc( 100 );
    vGatesFull = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_VecForEachEntry
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; i < Vec_VecSize(vGatesAll); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(vGatesAll, i))) && (((pGate) = (Aig_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vGatesAll, i), k)), 1); k++  {
        pGateR = Aig_Regular(pGate);
        if pGateR.fMarkA {
            continue; }
        pGateR.fMarkA = 1;
        Cgt_ManCollectFanoutPos( pAig, pGateR, vFanout );
        if Cgt_ManCheckGateComplete( pAig, vGatesAll, pGate, vFanout ) != 0 {
            Vec_PtrPush( vGatesFull, pGate ); }
    }
    }
    Vec_PtrFree( vFanout );
    // inlining of macro Vec_VecForEachEntry
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; i < Vec_VecSize(vGatesAll); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(vGatesAll, i))) && (((pGate) = (Aig_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vGatesAll, i), k)), 1); k++  { 
        Aig_Regular(pGate).fMarkA = 0;
    }
    }
    return vGatesFull;
}

/**Function*************************************************************

  Synopsis    [Calculates coverage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManComputeCoverage(pAig *Aig_Man_t, vGates *Vec_Vec_t) float32 {
    var nFrames int32  = 32;
    var nWords int32   =  1;
    var pSml *Ssw_Sml_t;
    var vOne *Vec_Ptr_t;
    var i, nTransSaved int32;
    pSml = Ssw_SmlSimulateSeq( pAig, 0, nFrames, nWords );
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(vGates)) && (((vOne) = Vec_VecEntry(vGates, i)), 1); i++  {
        nTransSaved += Ssw_SmlNodeCountOnesRealVec( pSml, vOne );
    }
    Ssw_SmlStop( pSml );
    return float32(100.0)*nTransSaved/32/nFrames/nWords/Vec_VecSize(vGates);
}

/**Function*************************************************************

  Synopsis    [Chooses what clock-gate to use for this register.]

  Description [Currently uses the naive approach: For each register, 
  choose the clock gate, which covers most of the transitions.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDecideSimple(pAig *Aig_Man_t, vGatesAll *Vec_Vec_t, nOdcMax int32, fVerbose int32) *Vec_Vec_t {
    var nFrames int32  = 32;
    var nWords int32   =  1;
    var pSml *Ssw_Sml_t;
    var vGates *Vec_Vec_t;
    var vCands *Vec_Ptr_t;
    var pObjLi, pObjLo, pCand, pCandBest *Aig_Obj_t;
    var i, k, nHitsCur, nHitsMax, Counter int32;
    var clk abctime  = Abc_Clock();
    var nTransTotal, nTransSaved int32;
    vGates = Vec_VecStart( Saig_ManRegNum(pAig) );
    pSml = Ssw_SmlSimulateSeq( pAig, 0, nFrames, nWords );
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
        nHitsMax = 0;
        pCandBest = nil;
        vCands = Vec_VecEntry( vGatesAll, i );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vCands)) && (((pCand) = (Aig_Obj_t *)Vec_PtrEntry(vCands, k)), 1); k++  {
            // check if this is indeed a clock-gate
            if nOdcMax == 0 && !Ssw_SmlCheckXorImplication( pSml, pObjLi, pObjLo, pCand ) {
                fmt.Printf("Clock gate candidate is invalid!\n"); }
            // find its characteristic number
            nHitsCur = Ssw_SmlNodeCountOnesReal( pSml, pCand );
            if nHitsMax < nHitsCur {
                nHitsMax = nHitsCur;
                pCandBest = pCand;
            }
        }
        if pCandBest != nil {
            Vec_VecPush( vGates, i, pCandBest );
            Counter++;
            nTransSaved += nHitsMax;
        }
        nTransTotal += 32 * nFrames * nWords;
    }
    Ssw_SmlStop( pSml );
    if fVerbose != 0 {
        fmt.Printf("Gating signals = %6d. Gated flops = %6d. (Total flops = %6d.)\n", Vec_VecSizeSize(vGatesAll), Counter, Saig_ManRegNum(pAig));
//        printf( "Gated transitions = %5.2f %%. (%5.2f %%.) ", 
//            100.0*nTransSaved/nTransTotal, Cgt_ManComputeCoverage(pAig, vGates) );
        fmt.Printf("Gated transitions = %5.2f %%. ", Cgt_ManComputeCoverage(pAig, vGates));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
/*
    {
        Vec_Ptr_t * vCompletes;
        vCompletes = Cgt_ManCompleteGates( pAig, vGatesAll, nOdcMax, fVerbose );
        printf( "Complete gates = %d. \n", Vec_PtrSize(vCompletes) );
        Vec_PtrFree( vCompletes );
    }
*/
    return vGates;
}

/**Function*************************************************************

  Synopsis    [Computes the set of complete clock gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDecideArea(pAig *Aig_Man_t, vGatesAll *Vec_Vec_t, nOdcMax int32, fVerbose int32) *Vec_Vec_t {
    var vGates *Vec_Vec_t;
    var vCompletes, vOne *Vec_Ptr_t;
    var pGate *Aig_Obj_t;
    var i, k, Counter int32;
    var clk abctime  = Abc_Clock();
    // derive and label complete gates
    vCompletes = Cgt_ManCompleteGates( pAig, vGatesAll, nOdcMax, fVerbose );
    // label complete gates
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCompletes)) && (((pGate) = (Aig_Obj_t *)Vec_PtrEntry(vCompletes, i)), 1); i++  {
        Aig_Regular(pGate).fMarkA = 1;
    }
    // select only complete gates
    vGates = Vec_VecStart( Saig_ManRegNum(pAig) );
    // inlining of macro Vec_VecForEachEntry
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; i < Vec_VecSize(vGatesAll); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(vGatesAll, i))) && (((pGate) = (Aig_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vGatesAll, i), k)), 1); k++  { 
        if Aig_Regular(pGate).fMarkA {
            Vec_VecPush( vGates, i, pGate ); }
    }
    }
    // unlabel complete gates
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCompletes)) && (((pGate) = (Aig_Obj_t *)Vec_PtrEntry(vCompletes, i)), 1); i++  {
        Aig_Regular(pGate).fMarkA = 0;
    }
    // count the number of gated flops
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(vGates)) && (((vOne) = Vec_VecEntry(vGates, i)), 1); i++  {
        Counter += int32((Vec_PtrSize(vOne) > 0));
//        printf( "%d ", Vec_PtrSize(vOne) );
    }
//    printf( "\n" );
    if fVerbose != 0 {
        fmt.Printf("Gating signals = %6d. Gated flops = %6d. (Total flops = %6d.)\n", Vec_VecSizeSize(vGatesAll), Counter, Saig_ManRegNum(pAig));
        fmt.Printf("Complete gates = %6d. Gated transitions = %5.2f %%. ", Vec_PtrSize(vCompletes), Cgt_ManComputeCoverage(pAig, vGates));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    Vec_PtrFree( vCompletes );
    return vGates;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

