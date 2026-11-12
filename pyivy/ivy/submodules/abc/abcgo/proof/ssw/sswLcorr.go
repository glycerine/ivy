package ssw

/**CFile****************************************************************

  FileName    [sswLcorr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Latch correspondence.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswLcorr.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"
//#include "bar.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Tranfers simulation information from FRAIG to AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepTransfer(p *Ssw_Man_t) {
    var pObj, pObjFraig *Aig_Obj_t;
    var pInfo *uint32;
    var i int32;
    // transfer simulation information
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        pObjFraig = Ssw_ObjFrame( p, pObj, 0 );
        if pObjFraig == Aig_ManConst0(p.pFrames) {
            Ssw_SmlObjAssignConst( p.pSml, pObj, 0, 0 );
            continue;
        }
        assert(!Aig_IsComplement(pObjFraig), "!Aig_IsComplement(pObjFraig)");
        assert(Aig_ObjIsCi(pObjFraig), "Aig_ObjIsCi(pObjFraig)");
        pInfo = (unsigned *)Vec_PtrEntry( p.vSimInfo, Aig_ObjCioId(pObjFraig) );
        Ssw_SmlObjSetWord( p.pSml, pObj, pInfo[0], 0, 0 );
    }
}

/**Function*************************************************************

  Synopsis    [Performs one round of simulation with counter-examples.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepResimulate(p *Ssw_Man_t) int32 {
    var RetValue1, RetValue2 int32;
    var clk abctime  = Abc_Clock();
    // transfer PI simulation information from storage
    Ssw_ManSweepTransfer( p );
    // simulate internal nodes
    Ssw_SmlSimulateOneFrame( p.pSml );
    // check equivalence classes
    RetValue1 = Ssw_ClassesRefineConst1( p.ppClasses, 1 );
    RetValue2 = Ssw_ClassesRefine( p.ppClasses, 1 );
    // prepare simulation info for the next round
    Vec_PtrCleanSimInfo( p.vSimInfo, 0, 1 );
    p.nPatterns = 0;
    p.nSimRounds++;
p.timeSimSat += Abc_Clock() - clk;
    return RetValue1 > 0 || RetValue2 > 0;
}

/**Function*************************************************************

  Synopsis    [Saves one counter-example into internal storage.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SmlAddPattern(p *Ssw_Man_t, pRepr *Aig_Obj_t, pCand *Aig_Obj_t) {
    var pObj *Aig_Obj_t;
    var pInfo *uint32;
    var i, nVarNum, Value int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pMSat.vUsedPis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pMSat.vUsedPis, i)), 1); i++  {
        nVarNum = Ssw_ObjSatNum( p.pMSat, pObj );
        assert(nVarNum > 0, "nVarNum > 0");
        Value = sat_solver_var_value( p.pMSat.pSat, nVarNum );
        if Value == 0 {
            continue; }
        pInfo = (unsigned *)Vec_PtrEntry( p.vSimInfo, Aig_ObjCioId(pObj) );
        Abc_InfoSetBit( pInfo, p.nPatterns );
    }
}

/**Function*************************************************************

  Synopsis    [Builds fraiged logic cone of the node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManBuildCone_rec(p *Ssw_Man_t, pObj *Aig_Obj_t) {
    var pObjNew *Aig_Obj_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Ssw_ObjFrame( p, pObj, 0 ) {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Ssw_ManBuildCone_rec( p, Aig_ObjFanin0(pObj) );
    Ssw_ManBuildCone_rec( p, Aig_ObjFanin1(pObj) );
    pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, 0), Ssw_ObjChild1Fra(p, pObj, 0) );
    Ssw_ObjSetFrame( p, pObj, 0, pObjNew );
}

/**Function*************************************************************

  Synopsis    [Recycles the SAT solver.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepLatchOne(p *Ssw_Man_t, pObjRepr *Aig_Obj_t, pObj *Aig_Obj_t) {
    var pObjFraig, pObjReprFraig, pObjLi *Aig_Obj_t;
    var RetValue int32;
    var clk abctime;
    assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
    assert(Aig_ObjIsCi(pObjRepr) || Aig_ObjIsConst1(pObjRepr), "Aig_ObjIsCi(pObjRepr) || Aig_ObjIsConst1(pObjRepr)");
    // check if it makes sense to skip some calls
    if p.nCallsCount > 100 && p.nCallsUnsat < p.nCallsSat {
        if ++p.nCallsDelta < 0 {
            return; }
    }
    p.nCallsDelta = 0;
clk = Abc_Clock();
    // get the fraiged node
    pObjLi = Saig_ObjLoToLi( p.pAig, pObj );
    Ssw_ManBuildCone_rec( p, Aig_ObjFanin0(pObjLi) );
    pObjFraig = Ssw_ObjChild0Fra( p, pObjLi, 0 );
    // get the fraiged representative
    if Aig_ObjIsCi(pObjRepr) {
        pObjLi = Saig_ObjLoToLi( p.pAig, pObjRepr );
        Ssw_ManBuildCone_rec( p, Aig_ObjFanin0(pObjLi) );
        pObjReprFraig = Ssw_ObjChild0Fra( p, pObjLi, 0 );
    
    } else {
        pObjReprFraig = Ssw_ObjFrame( p, pObjRepr, 0 ); }
p.timeReduce += Abc_Clock() - clk;
    // if the fraiged nodes are the same, return
    if Aig_Regular(pObjFraig) == Aig_Regular(pObjReprFraig) {
        return; }
    p.nRecycleCalls++;
    p.nCallsCount++;

    // check equivalence of the two nodes
    if (pObj.fPhase == pObjRepr.fPhase) != (Aig_ObjPhaseReal(pObjFraig) == Aig_ObjPhaseReal(pObjReprFraig)) {
        p.nPatterns++;
        p.nStrangers++;
        p.fRefined = 1;
    
    } else {
        RetValue = Ssw_NodesAreEquiv( p, Aig_Regular(pObjReprFraig), Aig_Regular(pObjFraig) );
        if RetValue == 1 {  // proved equivalence
        
            p.nCallsUnsat++;
            return;
        }
        if RetValue == -1 { // timed out
        
            Ssw_ClassesRemoveNode( p.ppClasses, pObj );
            p.nCallsUnsat++;
            p.fRefined = 1;
            return;
        
        } else { // disproved equivalence
        
            Ssw_SmlAddPattern( p, pObjRepr, pObj );
            p.nPatterns++;
            p.nCallsSat++;
            p.fRefined = 1;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Performs one iteration of sweeping latches.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManSweepLatch(p *Ssw_Man_t) int32 {
//    Bar_Progress_t * pProgress = NULL;
    var vClass *Vec_Ptr_t;
    var pObj, pRepr, pTemp *Aig_Obj_t;
    var i, k int32;

    // start the timeframe
    p.pFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) );
    // map constants and PIs
    Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), 0, Aig_ManConst1(p.pFrames) );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Ssw_ObjSetFrame( p, pObj, 0, Aig_ObjCreateCi(p.pFrames) );
    }

    // implement equivalence classes
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        pRepr = Aig_ObjRepr( p.pAig, pObj );
        if pRepr == nil {
            pTemp = Aig_ObjCreateCi(p.pFrames);
            pTemp.pData = pObj;
        
        } else {
            pTemp = Aig_NotCond( Ssw_ObjFrame(p, pRepr, 0), pRepr.fPhase ^ pObj.fPhase ); }
        Ssw_ObjSetFrame( p, pObj, 0, pTemp );
    }
    Aig_ManSetCioIds( p.pFrames );

    // prepare simulation info
    assert(p.vSimInfo == nil, "p.vSimInfo == nil");
    p.vSimInfo = Vec_PtrAllocSimInfo( Aig_ManCiNum(p.pFrames), 1 );
    Vec_PtrCleanSimInfo( p.vSimInfo, 0, 1 );

    // go through the registers
//    if ( p->pPars->fVerbose )
//        pProgress = Bar_ProgressStart( stdout, Aig_ManRegNum(p->pAig) );
    vClass = Vec_PtrAlloc( 100 );
    p.fRefined = 0;
    p.nCallsCount = p.nCallsSat = p.nCallsUnsat = 0;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
//        if ( p->pPars->fVerbose )
//            Bar_ProgressUpdate( pProgress, i, NULL );
        // consider the case of constant candidate
        if Ssw_ObjIsConst1Cand( p.pAig, pObj ) {
            Ssw_ManSweepLatchOne( p, Aig_ManConst1(p.pAig), pObj );
        } else {
            // consider the case of equivalence class
            Ssw_ClassesCollectClass( p.ppClasses, pObj, vClass );
            if Vec_PtrSize(vClass) == 0 {
                continue; }
            // try to prove equivalences in this class
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vClass)) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(vClass, k)), 1); k++  {
                if Aig_ObjRepr(p.pAig, pTemp) == pObj {
                    Ssw_ManSweepLatchOne( p, pObj, pTemp );
                    if p.nPatterns == 32 {
                        break; }
                }
            }
        }
        // resimulate
        if p.nPatterns == 32 {
            Ssw_ManSweepResimulate( p ); }
        // attempt recycling the SAT solver
        if p.pPars.nSatVarMax &&
             p.pMSat.nSatVars > p.pPars.nSatVarMax &&
             p.nRecycleCalls > p.pPars.nRecycleCalls {
            p.nVarsMax  = Abc_MaxInt( p.nVarsMax,  p.pMSat.nSatVars );
            p.nCallsMax = Abc_MaxInt( p.nCallsMax, p.pMSat.nSolverCalls );
            Ssw_SatStop( p.pMSat );
            p.pMSat = Ssw_SatStart( 0 );
            p.nRecycles++;
            p.nRecycleCalls = 0;
        }
    }
//    ABC_PRT( "reduce", p->timeReduce );
//    Aig_TableProfile( p->pFrames );
//    Abc_Print( 1, "And gates = %d\n", Aig_ManNodeNum(p->pFrames) );
    // resimulate
    if p.nPatterns > 0 {
        Ssw_ManSweepResimulate( p ); }
    // cleanup
    Vec_PtrFree( vClass );
//    if ( p->pPars->fVerbose )
//        Bar_ProgressStop( pProgress );

    // cleanup
//    Ssw_ClassesCheck( p->ppClasses );
    return p.fRefined;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
