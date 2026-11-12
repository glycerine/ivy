package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcCexMin1.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [CEX minimization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcCexMin1.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/ioa/ioa.h"
//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Find the roots to begin traversal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinGetCos(pAig *Aig_Man_t, pCex *Abc_Cex_t, vLeaves *Vec_Int_t, vRoots *Vec_Int_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Vec_IntClear( vRoots );
    if vLeaves == nil {
        pObj = Aig_ManCo( pAig, pCex.iPo );
        Vec_IntPush( vRoots, Aig_ObjId(pObj) );
        return;
    }
    // inlining of macro Aig_ManForEachObjVec
    for  i = 0; i < Vec_IntSize(vLeaves) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vLeaves,i))), 1); i++  {
        if Saig_ObjIsLo(pAig, pObj) {
            Vec_IntPush( vRoots, Aig_ObjId( Saig_ObjLoToLi(pAig, pObj) ) ); }
    }
}

/**Function*************************************************************

  Synopsis    [Collects CI of the given timeframe.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinCollectFrameTerms_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, vFrameCisOne *Vec_Int_t) {
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    if Aig_ObjIsCo(pObj) {
        Saig_ManCexMinCollectFrameTerms_rec( pAig, Aig_ObjFanin0(pObj), vFrameCisOne );
    } else if Aig_ObjIsNode(pObj) {
        Saig_ManCexMinCollectFrameTerms_rec( pAig, Aig_ObjFanin0(pObj), vFrameCisOne );
        Saig_ManCexMinCollectFrameTerms_rec( pAig, Aig_ObjFanin1(pObj), vFrameCisOne );
    
    } else if Aig_ObjIsCi(pObj) {
        Vec_IntPush( vFrameCisOne, Aig_ObjId(pObj) ); }
}

/**Function*************************************************************

  Synopsis    [Collects CIs of all timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinCollectFrameTerms(pAig *Aig_Man_t, pCex *Abc_Cex_t) *Vec_Vec_t {
    var vFrameCis *Vec_Vec_t;
    var vRoots, vLeaves *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, f int32;
    // create terminals
    vRoots = Vec_IntAlloc( 1000 );
    vFrameCis = Vec_VecStart( pCex.iFrame+1 );
    for  f = pCex.iFrame; f >= 0; f--  {
        // create roots
        if f == pCex.iFrame {
            vLeaves = nil;
        } else {
            vLeaves = Vec_VecEntryInt(vFrameCis, f+1);
        }
        Saig_ManCexMinGetCos( pAig, pCex, vLeaves, vRoots );
        // collect nodes starting from the roots
        Aig_ManIncrementTravId( pAig );
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vRoots,i))), 1); i++  {
            Saig_ManCexMinCollectFrameTerms_rec( pAig, pObj, Vec_VecEntryInt(vFrameCis, f) );
        }
    }
    Vec_IntFree( vRoots );
    return vFrameCis;
}



/**Function*************************************************************

  Synopsis    [Recursively sets phase and priority.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinDerivePhasePriority_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t) {
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    if Aig_ObjIsCo(pObj) {
        Saig_ManCexMinDerivePhasePriority_rec( pAig, Aig_ObjFanin0(pObj) );
        assert(Aig_ObjFanin0(pObj).iData >= 0, "Aig_ObjFanin0(pObj).iData >= 0");
        pObj.iData = Aig_ObjFanin0(pObj).iData ^ Aig_ObjFaninC0(pObj);
    
    } else if Aig_ObjIsNode(pObj) {
        var fPhase0, fPhase1, iPrio0, iPrio1 int32;
        Saig_ManCexMinDerivePhasePriority_rec( pAig, Aig_ObjFanin0(pObj) );
        Saig_ManCexMinDerivePhasePriority_rec( pAig, Aig_ObjFanin1(pObj) );
        assert(Aig_ObjFanin0(pObj).iData >= 0, "Aig_ObjFanin0(pObj).iData >= 0");
        assert(Aig_ObjFanin1(pObj).iData >= 0, "Aig_ObjFanin1(pObj).iData >= 0");
        fPhase0 = Abc_LitIsCompl( Aig_ObjFanin0(pObj).iData ) ^ Aig_ObjFaninC0(pObj);
        fPhase1 = Abc_LitIsCompl( Aig_ObjFanin1(pObj).iData ) ^ Aig_ObjFaninC1(pObj);
        iPrio0  = Abc_Lit2Var( Aig_ObjFanin0(pObj).iData );
        iPrio1  = Abc_Lit2Var( Aig_ObjFanin1(pObj).iData );
        if (fPhase0 != 0) && (fPhase1 != 0) { // both are one
            pObj.iData = Abc_Var2Lit( Abc_MinInt(iPrio0, iPrio1), 1 );
        } else if !(fPhase0 != 0) && (fPhase1 != 0) { 
            pObj.iData = Abc_Var2Lit( iPrio0, 0 );
        } else if (fPhase0 != 0) && !(fPhase1 != 0) {
            pObj.iData = Abc_Var2Lit( iPrio1, 0 );
        } else { // both are zero
            pObj.iData = Abc_Var2Lit( Abc_MaxInt(iPrio0, iPrio1), 0 ); }
    }
}

/**Function*************************************************************

  Synopsis    [Verify phase.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinVerifyPhase(pAig *Aig_Man_t, pCex *Abc_Cex_t, f int32) {
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManConst1(pAig).fPhase = 1;
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.fPhase = Abc_InfoHasBit(pCex.pData, pCex.nRegs + f * pCex.nPis + i);
    }
    if f == 0 {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
            pObj.fPhase = 0;
        }
    
    } else {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
            pObj.fPhase = Saig_ObjLoToLi(pAig, pObj).fPhase;
        }
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.fPhase = (Aig_ObjFaninC0(pObj) ^ Aig_ObjFanin0(pObj).fPhase) & 
                       (Aig_ObjFaninC1(pObj) ^ Aig_ObjFanin1(pObj).fPhase); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        pObj.fPhase = (Aig_ObjFaninC0(pObj) ^ Aig_ObjFanin0(pObj).fPhase);
    }
}

/**Function*************************************************************

  Synopsis    [Collects phase and priority of all timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinDerivePhasePriority(pAig *Aig_Man_t, pCex *Abc_Cex_t, vFrameCis *Vec_Vec_t, vFramePPs *Vec_Vec_t, f int32, vRoots *Vec_Int_t) {
    var vFramePPsOne, vFrameCisOne, vLeaves *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // set PP for the CIs
    vFrameCisOne = Vec_VecEntryInt( vFrameCis, f );
    vFramePPsOne = Vec_VecEntryInt( vFramePPs, f );
    // inlining of macro Aig_ManForEachObjVec
    for  i = 0; i < Vec_IntSize(vFrameCisOne) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vFrameCisOne,i))), 1); i++  {
        pObj.iData = Vec_IntEntry( vFramePPsOne, i );
        assert(pObj.iData >= 0, "pObj.iData >= 0");
    }
//    if ( f == 0 )
//        Saig_ManCexMinVerifyPhase( pAig, pCex, f );

    // create roots
    if f == pCex.iFrame {
        vLeaves = nil;
    } else {
        vLeaves = Vec_VecEntryInt(vFrameCis, f+1);
    }
    Saig_ManCexMinGetCos( pAig, pCex, vLeaves, vRoots );
    // derive for the nodes starting from the roots
    Aig_ManIncrementTravId( pAig );
    // inlining of macro Aig_ManForEachObjVec
    for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vRoots,i))), 1); i++  {
        Saig_ManCexMinDerivePhasePriority_rec( pAig, pObj );
//        if ( f == 0 )
//            assert( (pObj->iData & 1) == pObj->fPhase );
    }
}

/**Function*************************************************************

  Synopsis    [Collects phase and priority of all timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Saig_ManCexMinCollectPhasePriority_(pAig *Aig_Man_t, pCex *Abc_Cex_t, vFrameCis *Vec_Vec_t) *Vec_Vec_t {
    var vFramePPs *Vec_Vec_t;
    var vRoots, vFramePPsOne, vFrameCisOne *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, f, nPrioOffset int32;

    // initialize phase and priority
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.iData = -1; }
    }

    // set the constant node to higher priority than the flops
    vFramePPs = Vec_VecStart( pCex.iFrame+1 );
    nPrioOffset = (pCex.iFrame + 1) * pCex.nPis;
    Aig_ManConst1(pAig).iData = Abc_Var2Lit( nPrioOffset + pCex.nRegs, 1 );
    vRoots = Vec_IntAlloc( 1000 );
    for  f = 0; f <= pCex.iFrame; f++  {
        var nPiCount int32;
        // fill in PP for the CIs
        vFrameCisOne = Vec_VecEntryInt( vFrameCis, f );
        vFramePPsOne = Vec_VecEntryInt( vFramePPs, f );
        assert(Vec_IntSize(vFramePPsOne) == 0, "Vec_IntSize(vFramePPsOne) == 0");
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vFrameCisOne) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vFrameCisOne,i))), 1); i++  {
            assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
            if Saig_ObjIsPi(pAig, pObj) {
                Vec_IntPush( vFramePPsOne, Abc_Var2Lit( (f+1) * pCex.nPis - nPiCount++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + f * pCex.nPis + Aig_ObjCioId(pObj)) ) );
            } else if f == 0 {
                Vec_IntPush( vFramePPsOne, Abc_Var2Lit( nPrioOffset + Saig_ObjRegId(pAig, pObj), 0 ) );
            } else {
                Vec_IntPush( vFramePPsOne, Saig_ObjLoToLi(pAig, pObj).iData ); }
        }
        // compute the PP info
        Saig_ManCexMinDerivePhasePriority( pAig, pCex, vFrameCis, vFramePPs, f, vRoots );
    }
    Vec_IntFree( vRoots );
    // check the output
    pObj = Aig_ManCo( pAig, pCex.iPo );
    assert(Abc_LitIsCompl(pObj.iData), "Abc_LitIsCompl(pObj.iData)");
    return vFramePPs;
}

/**Function*************************************************************

  Synopsis    [Collects phase and priority of all timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Saig_ManCexMinCollectPhasePriority(pAig *Aig_Man_t, pCex *Abc_Cex_t, vFrameCis *Vec_Vec_t) *Vec_Vec_t {
    var vFramePPs *Vec_Vec_t;
    var vRoots, vFramePPsOne, vFrameCisOne *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, f, nPrioOffset int32;

    // initialize phase and priority
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.iData = -1; }
    }

    // set the constant node to higher priority than the flops
    vFramePPs = Vec_VecStart( pCex.iFrame+1 );
    nPrioOffset = pCex.nRegs;
    Aig_ManConst1(pAig).iData = Abc_Var2Lit( nPrioOffset + (pCex.iFrame + 1) * pCex.nPis, 1 );
    vRoots = Vec_IntAlloc( 1000 );
    for  f = 0; f <= pCex.iFrame; f++  {
        var nPiCount int32;
        // fill in PP for the CIs
        vFrameCisOne = Vec_VecEntryInt( vFrameCis, f );
        vFramePPsOne = Vec_VecEntryInt( vFramePPs, f );
        assert(Vec_IntSize(vFramePPsOne) == 0, "Vec_IntSize(vFramePPsOne) == 0");
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vFrameCisOne) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vFrameCisOne,i))), 1); i++  {
            assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
            if Saig_ObjIsPi(pAig, pObj) {
                Vec_IntPush( vFramePPsOne, Abc_Var2Lit( nPrioOffset + (f+1) * pCex.nPis - 1 - nPiCount++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + f * pCex.nPis + Aig_ObjCioId(pObj)) ) );
            } else if f == 0 {
                Vec_IntPush( vFramePPsOne, Abc_Var2Lit( Saig_ObjRegId(pAig, pObj), 0 ) );
            } else {
                Vec_IntPush( vFramePPsOne, Saig_ObjLoToLi(pAig, pObj).iData ); }
        }
        // compute the PP info
        Saig_ManCexMinDerivePhasePriority( pAig, pCex, vFrameCis, vFramePPs, f, vRoots );
    }
    Vec_IntFree( vRoots );
    // check the output
    pObj = Aig_ManCo( pAig, pCex.iPo );
    assert(Abc_LitIsCompl(pObj.iData), "Abc_LitIsCompl(pObj.iData)");
    return vFramePPs;
}


/**Function*************************************************************

  Synopsis    [Returns reasons for the property to fail.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinCollectReason_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vReason *Vec_Int_t, fPiReason int32) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsCi(pObj) {
        if (fPiReason != 0) && Saig_ObjIsPi(p, pObj) {
            Vec_IntPush( vReason, Abc_Var2Lit( Aig_ObjCioId(pObj), !Abc_LitIsCompl(pObj.iData) ) );
        } else if !(fPiReason != 0) && Saig_ObjIsLo(p, pObj) {
            Vec_IntPush( vReason, Abc_Var2Lit( Saig_ObjRegId(p, pObj), !Abc_LitIsCompl(pObj.iData) ) ); }
        return;
    }
    if Aig_ObjIsCo(pObj) {
        Saig_ManCexMinCollectReason_rec( p, Aig_ObjFanin0(pObj), vReason, fPiReason );
        return;
    }
    if Aig_ObjIsConst1(pObj) {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    if Abc_LitIsCompl(pObj.iData) { // value 1
    
        var fPhase0 int32  = Abc_LitIsCompl( Aig_ObjFanin0(pObj).iData ) ^ Aig_ObjFaninC0(pObj);
        var fPhase1 int32  = Abc_LitIsCompl( Aig_ObjFanin1(pObj).iData ) ^ Aig_ObjFaninC1(pObj);
        assert((fPhase0 != 0) && (fPhase1 != 0), "fPhase0 && fPhase1");
        Saig_ManCexMinCollectReason_rec( p, Aig_ObjFanin0(pObj), vReason, fPiReason );
        Saig_ManCexMinCollectReason_rec( p, Aig_ObjFanin1(pObj), vReason, fPiReason );
    
    } else {
        var fPhase0 int32  = Abc_LitIsCompl( Aig_ObjFanin0(pObj).iData ) ^ Aig_ObjFaninC0(pObj);
        var fPhase1 int32  = Abc_LitIsCompl( Aig_ObjFanin1(pObj).iData ) ^ Aig_ObjFaninC1(pObj);
        assert(!(fPhase0 != 0) || !(fPhase1 != 0), "!fPhase0 || !fPhase1");
        if !(fPhase0 != 0) && (fPhase1 != 0) {
            Saig_ManCexMinCollectReason_rec( p, Aig_ObjFanin0(pObj), vReason, fPiReason );
        } else if (fPhase0 != 0) && !(fPhase1 != 0) {
            Saig_ManCexMinCollectReason_rec( p, Aig_ObjFanin1(pObj), vReason, fPiReason );
        } else {
            var iPrio0 int32  = Abc_Lit2Var( Aig_ObjFanin0(pObj).iData );
            var iPrio1 int32  = Abc_Lit2Var( Aig_ObjFanin1(pObj).iData );
            if iPrio0 >= iPrio1 {
                Saig_ManCexMinCollectReason_rec( p, Aig_ObjFanin0(pObj), vReason, fPiReason );
            } else {
                Saig_ManCexMinCollectReason_rec( p, Aig_ObjFanin1(pObj), vReason, fPiReason ); }
        }
    }
}


/**Function*************************************************************

  Synopsis    [Collects phase and priority of all timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinCollectReason(pAig *Aig_Man_t, pCex *Abc_Cex_t, vFrameCis *Vec_Vec_t, vFramePPs *Vec_Vec_t, fPiReason int32) *Vec_Vec_t {
    var vFrameReas *Vec_Vec_t;
    var vRoots, vLeaves *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, f int32;
    // select reason for the property to fail
    vFrameReas = Vec_VecStart( pCex.iFrame+1 );
    vRoots = Vec_IntAlloc( 1000 );
    for  f = pCex.iFrame; f >= 0; f--  {
        // set phase and polarity
        Saig_ManCexMinDerivePhasePriority( pAig, pCex, vFrameCis, vFramePPs, f, vRoots );
        // create roots
        if f == pCex.iFrame {
            vLeaves = nil;
        } else {
            vLeaves = Vec_VecEntryInt(vFrameCis, f+1);
        }
        Saig_ManCexMinGetCos( pAig, pCex, vLeaves, vRoots );
        // collect nodes starting from the roots
        Aig_ManIncrementTravId( pAig );
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vRoots,i))), 1); i++  {
            Saig_ManCexMinCollectReason_rec( pAig, pObj, Vec_VecEntryInt(vFrameReas, f), fPiReason );
        }
//printf( "%d(%d) ", Vec_VecLevelSize(vFrameCis, f), Vec_VecLevelSize(vFrameReas, f) ); 
    }
//printf( "\n" );
    Vec_IntFree( vRoots );
    return vFrameReas;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinComputeReason(pAig *Aig_Man_t, pCex *Abc_Cex_t, fPiReason int32) *Vec_Vec_t {
    var vFrameCis, vFramePPs, vFrameReas *Vec_Vec_t;
    // sanity checks
    assert(pCex.nPis == Saig_ManPiNum(pAig), "pCex.nPis == Saig_ManPiNum(pAig)");
    assert(pCex.nRegs == Saig_ManRegNum(pAig), "pCex.nRegs == Saig_ManRegNum(pAig)");
    assert(pCex.iPo >= 0 && pCex.iPo < Saig_ManPoNum(pAig), "pCex.iPo >= 0 && pCex.iPo < Saig_ManPoNum(pAig)");
    // derive frame terms
    vFrameCis = Saig_ManCexMinCollectFrameTerms( pAig, pCex );
    // derive phase and priority
    vFramePPs = Saig_ManCexMinCollectPhasePriority( pAig, pCex, vFrameCis );
    // derive reasons for property failure
    vFrameReas = Saig_ManCexMinCollectReason( pAig, pCex, vFrameCis, vFramePPs, fPiReason );
    Vec_VecFree( vFramePPs );
    Vec_VecFree( vFrameCis );
    return vFrameReas;
}


/**Function*************************************************************

  Synopsis    [Duplicate with literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinDupWithCubes(pAig *Aig_Man_t, vReg2Value *Vec_Vec_t) *Aig_Man_t {
    var vLevel *Vec_Int_t;
    var pAigNew *Aig_Man_t;
    var pObj, pMiter *Aig_Obj_t;
    var i, k, Lit int32;
    assert(pAig.nConstrs == 0, "pAig.nConstrs == 0");
    // start the new manager
    pAigNew = Aig_ManStart( Aig_ManNodeNum(pAig) + Vec_VecSizeSize(vReg2Value) + Vec_VecSize(vReg2Value) );
    pAigNew.pName = Abc_UtilStrsav( pAig.pName );
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pAigNew );
    // create variables for PIs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pAigNew );
    }
    // add internal nodes of this frame
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create POs for cubes
    // inlining of macro Vec_VecForEachLevelInt
    for  i = 0; (i < Vec_VecSize(vReg2Value)) && (((vLevel) = Vec_VecEntryInt(vReg2Value, i)), 1); i++  {
        if i == 0 {
            continue; }
        pMiter = Aig_ManConst1( pAigNew );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((Lit) = Vec_IntEntry(vLevel, k)), 1); k++  {
            assert(Lit >= 0 && Lit < 2 * Aig_ManRegNum(pAig), "Lit >= 0 && Lit < 2 * Aig_ManRegNum(pAig)");
            pObj = Saig_ManLi( pAig, Abc_Lit2Var(Lit) );
            pMiter = Aig_And( pAigNew, pMiter, Aig_NotCond(Aig_ObjChild0Copy(pObj), Abc_LitIsCompl(Lit)) );
        }
        Aig_ObjCreateCo( pAigNew, pMiter );
    }
    // transfer to register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }
    // finalize
    Aig_ManCleanup( pAigNew );
    Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAig) );
    return pAigNew;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinPerform(pAig *Aig_Man_t, pCex *Abc_Cex_t) *Abc_Cex_t {
    var fReasonPi int32;

    var pCexMin *Abc_Cex_t  = nil;
    var pManNew *Aig_Man_t  = nil;
    var vFrameReas *Vec_Vec_t;
    vFrameReas = Saig_ManCexMinComputeReason( pAig, pCex, fReasonPi );
    fmt.Printf("Reason size = %d.  Ave = %d.\n", Vec_VecSizeSize(vFrameReas), Vec_VecSizeSize(vFrameReas)/(pCex.iFrame + 1));

    // try reducing the frames
    if !(fReasonPi != 0) {
        var pFileName string  = "aigcube.aig";
        pManNew = Saig_ManCexMinDupWithCubes( pAig, vFrameReas );
        Ioa_WriteAiger( pManNew, pFileName, 0, 0 );
        Aig_ManStop( pManNew );
        fmt.Printf("Intermediate AIG is written into file \"%s\".\n", pFileName);
    }

    // find reduced counter-example
    Vec_VecFree( vFrameReas );
    return pCexMin;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

