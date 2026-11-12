package ssc

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [sscCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sweeping under constraints.]

  Synopsis    [The core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: sscCore.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sscInt.h"
//#include "proof/cec/cec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_ManSetDefaultParams(p *Ssc_Pars_t) {
    memset( p, 0, sizeof(Ssc_Pars_t) );
    p.nWords         =     1;  // the number of simulation words
    p.nBTLimit       =  1000;  // conflict limit at a node
    p.nSatVarMax     =  5000;  // the max number of SAT variables
    p.nCallsRecycle  =   100;  // calls to perform before recycling SAT solver
    p.fVerbose       =     0;  // verbose stats
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_ManStop(p *Ssc_Man_t) {
    Vec_IntFreeP( &p.vFront );
    Vec_IntFreeP( &p.vFanins );
    Vec_IntFreeP( &p.vPattern );
    Vec_IntFreeP( &p.vDisPairs );
    Vec_IntFreeP( &p.vPivot );
    Vec_IntFreeP( &p.vId2Var );
    Vec_IntFreeP( &p.vVar2Id );
    if p.pSat { sat_solver_delete( p.pSat ); }
    Gia_ManStopP( &p.pFraig );
    p = nil;
}
func Ssc_ManStart(pAig *Gia_Man_t, pCare *Gia_Man_t, pPars *Ssc_Pars_t) *Ssc_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Ssc_Man_t{};
    p.pPars  = pPars;
    p.pAig   = pAig;
    p.pCare  = pCare;
    p.pFraig = Gia_ManDupDfs( p.pCare );
    assert(Vec_IntSize(&p.pFraig.vHTable) == 0, "Vec_IntSize(&p.pFraig.vHTable) == 0");
    assert(!Gia_ManHasDangling(p.pFraig), "!Gia_ManHasDangling(p.pFraig)");
    Gia_ManInvertPos( p.pFraig );
    Ssc_ManStartSolver( p );
    if p.pSat == nil {
        fmt.Printf("Constraints are UNSAT after propagation.\n");
        Ssc_ManStop( p );
        return (Ssc_Man_t *)(ABC_PTRINT_T)1;
    }
//    p->vPivot = Ssc_GiaFindPivotSim( p->pFraig );
//    Vec_IntFreeP( &p->vPivot  );
    p.vPivot = Ssc_ManFindPivotSat( p );
    if p.vPivot == (Vec_Int_t *)(ABC_PTRINT_T)1 {
        fmt.Printf("Constraints are UNSAT.\n");
        Ssc_ManStop( p );
        return (Ssc_Man_t *)(ABC_PTRINT_T)1;
    }
    if p.vPivot == nil {
        fmt.Printf("Conflict limit is reached while trying to find one SAT assignment.\n");
        Ssc_ManStop( p );
        return nil;
    }
    sat_solver_bookmark( p.pSat );
//    Vec_IntPrint( p->vPivot );
    Gia_ManSetPhasePattern( p.pAig, p.vPivot );
    Gia_ManSetPhasePattern( p.pCare, p.vPivot );
    if Gia_ManCheckCoPhase(p.pCare) {
        fmt.Printf("Computed reference pattern violates %d constraints (this is a bug!).\n", Gia_ManCheckCoPhase(p.pCare));
        Ssc_ManStop( p );
        return nil;
    }
    // other things
    p.vDisPairs = Vec_IntAlloc( 100 );
    p.vPattern = Vec_IntAlloc( 100 );
    p.vFanins = Vec_IntAlloc( 100 );
    p.vFront = Vec_IntAlloc( 100 );
    Ssc_GiaClassesInit( pAig );
    // now it is ready for refinement using simulation 
    return p;
}
func Ssc_ManPrintStats(p *Ssc_Man_t) {
    Abc_Print( 1, "Parameters: SimWords = %d. SatConfs = %d. SatVarMax = %d. CallsRec = %d. Verbose = %d.\n",
        p.pPars.nWords, p.pPars.nBTLimit, p.pPars.nSatVarMax, p.pPars.nCallsRecycle, p.pPars.fVerbose );
    Abc_Print( 1, "SAT calls : Total = %d. Proof = %d. Cex = %d. Undec = %d.\n",
        p.nSatCalls, p.nSatCallsUnsat, p.nSatCallsSat, p.nSatCallsUndec );
    Abc_Print( 1, "SAT solver: Vars = %d. Clauses = %d. Recycles = %d. Sim rounds = %d.\n",
        sat_solver_nvars(p.pSat), sat_solver_nclauses(p.pSat), p.nRecycles, p.nSimRounds );

    p.timeOther = p.timeTotal - p.timeSimInit - p.timeSimSat - p.timeCnfGen - p.timeSatSat - p.timeSatUnsat - p.timeSatUndec;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Initialization "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSimInit)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSimInit)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("SAT simulation "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSimSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSimSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("CNF generation "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSimSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSimSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("SAT solving    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat-p.timeCnfGen)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat-p.timeCnfGen)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  unsat        "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUnsat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  sat          "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  undecided    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUndec)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUndec)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Other          "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("TOTAL          "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
}

/**Function*************************************************************

  Synopsis    [Refine one class by resimulating one pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaSimulatePatternFraig_rec(p *Ssc_Man_t, iFraigObj int32) int32 {
    var pObj *Gia_Obj_t;
    var Res0, Res1 int32;
    if Ssc_ObjSatVar(p, iFraigObj) {
        return sat_solver_var_value( p.pSat, Ssc_ObjSatVar(p, iFraigObj) ); }
    pObj = Gia_ManObj( p.pFraig, iFraigObj );
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Res0 = Ssc_GiaSimulatePatternFraig_rec( p, Gia_ObjFaninId0(pObj, iFraigObj) );
    Res1 = Ssc_GiaSimulatePatternFraig_rec( p, Gia_ObjFaninId1(pObj, iFraigObj) );
    pObj.fMark0 = (Res0 ^ Gia_ObjFaninC0(pObj)) & (Res1 ^ Gia_ObjFaninC1(pObj));
    return pObj.fMark0;
}
func Ssc_GiaSimulatePattern_rec(p *Ssc_Man_t, pObj *Gia_Obj_t) int32 {
    var Res0, Res1 int32;
    if Gia_ObjIsTravIdCurrent(p.pAig, pObj) {
        return pObj.fMark0; }
    Gia_ObjSetTravIdCurrent(p.pAig, pObj);
    if ~pObj.Value {  // mapping into FRAIG exists - simulate FRAIG
     
        Res0 = Ssc_GiaSimulatePatternFraig_rec( p, Abc_Lit2Var(pObj.Value) );
        pObj.fMark0 = Res0 ^ Abc_LitIsCompl(pObj.Value);
    
    } else { // mapping into FRAIG does not exist - simulate AIG
    
        Res0 = Ssc_GiaSimulatePattern_rec( p, Gia_ObjFanin0(pObj) );
        Res1 = Ssc_GiaSimulatePattern_rec( p, Gia_ObjFanin1(pObj) );
        pObj.fMark0 = (Res0 ^ Gia_ObjFaninC0(pObj)) & (Res1 ^ Gia_ObjFaninC1(pObj));
    }
    return pObj.fMark0;
}
func Ssc_GiaResimulateOneClass(p *Ssc_Man_t, iRepr int32, iObj int32) int32 {
    var Ent, RetValue int32;
    assert(iRepr == Gia_ObjRepr(p.pAig, iObj), "iRepr == Gia_ObjRepr(p.pAig, iObj)");
    assert(Gia_ObjIsHead( p.pAig, iRepr ), "Gia_ObjIsHead( p.pAig, iRepr )");
    // set bit-values at the nodes according to the counter-example
    Gia_ManIncrementTravId( p.pAig );
    // inlining of macro Gia_ClassForEachObj
    for  assert(Gia_ObjIsHead(p.pAig, iRepr) && (iRepr != 0), "Gia_ObjIsHead(p.pAig, iRepr) && iRepr"), Ent = iRepr; Ent > 0; Ent = Gia_ObjNext(p.pAig, Ent)  {
        Ssc_GiaSimulatePattern_rec( p, Gia_ManObj(p.pAig, Ent) );
    }
    // refine one class using these bit-values
    RetValue = Ssc_GiaSimClassRefineOneBit( p.pAig, iRepr );
    // check that the candidate equivalence is indeed refined
    assert(iRepr != Gia_ObjRepr(p.pAig, iObj), "iRepr != Gia_ObjRepr(p.pAig, iObj)");
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Perform verification of conditional sweeping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_PerformVerification(p0 *Gia_Man_t, p1 *Gia_Man_t, pC *Gia_Man_t) int32 {
    var Status int32;
    var ParsCec Cec_ParCec_t;
    var pPars *Cec_ParCec_t = &ParsCec;
    // derive the OR of constraint outputs
    var pCond *Gia_Man_t  = Gia_ManDupAndOr( pC, Gia_ManPoNum(p0), 1, 0 );
    // derive F = F & !OR(c0, c1, c2, ...)
    var p0c *Gia_Man_t  = Gia_ManMiter( p0, pCond, 0, 0, 0, 1, 0 );
    // derive F = F & !OR(c0, c1, c2, ...)
    var p1c *Gia_Man_t  = Gia_ManMiter( p1, pCond, 0, 0, 0, 1, 0 );
    // derive dual-output miter
    var pMiter *Gia_Man_t  = Gia_ManMiter( p0c, p1c, 0, 1, 0, 0, 0 );
    Gia_ManStop( p0c );
    Gia_ManStop( p1c );
    Gia_ManStop( pCond );
    // run equivalence checking
    Cec_ManCecSetDefaultParams( pPars );
    Status = Cec_ManVerify( pMiter, pPars );
    Gia_ManStop( pMiter );
    // report the results
    if Status == 1 {
        fmt.Printf("Verification succeeded.\n");
    } else if Status == 0 {
        fmt.Printf("Verification failed.\n");
    } else if Status == -1 {
        fmt.Printf("Verification undecided.\n");
    } else { assert(false, "0"); }
    return Status;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_PerformSweepingInt(pAig *Gia_Man_t, pCare *Gia_Man_t, pPars *Ssc_Pars_t) *Gia_Man_t {
    var p *Ssc_Man_t;
    var pResult, pTemp *Gia_Man_t;
    var pObj, pRepr *Gia_Obj_t;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    var i, fCompl, nRefined, status int32;
clk = Abc_Clock();
    assert(Gia_ManRegNum(pCare) == 0, "Gia_ManRegNum(pCare) == 0");
    assert(Gia_ManCiNum(pAig) == Gia_ManCiNum(pCare), "Gia_ManCiNum(pAig) == Gia_ManCiNum(pCare)");
    assert(Gia_ManIsNormalized(pAig), "Gia_ManIsNormalized(pAig)");
    assert(Gia_ManIsNormalized(pCare), "Gia_ManIsNormalized(pCare)");
    // reset random numbers
    Gia_ManRandom( 1 );
    // sweeping manager
    p = Ssc_ManStart( pAig, pCare, pPars );
    if p == (Ssc_Man_t *)(ABC_PTRINT_T)1 { // UNSAT
        return Gia_ManDupZero( pAig ); }
    if p == nil { // timeout
        return Gia_ManDup( pAig ); }
    if p.pPars.fVerbose {
        fmt.Printf("Care set produced %d hits out of %d.\n", Ssc_GiaEstimateCare(p.pFraig, 5), 640); }
    // perform simulation 
    for {
        // simulate care set
        Ssc_GiaRandomPiPattern( p.pFraig, 5, nil );
        Ssc_GiaSimRound( p.pFraig );
        // transfer care patterns to user's AIG
        if !Ssc_GiaTransferPiPattern( pAig, p.pFraig, p.vPivot ) {
            break; }
        // simulate the main AIG
        Ssc_GiaSimRound( pAig );
        nRefined = Ssc_GiaClassesRefine( pAig );
        if pPars.fVerbose { 
            Gia_ManEquivPrintClasses( pAig, 0, 0 ); }
        if nRefined <= Gia_ManCandNum(pAig) / 100 {
            break; }
    }
p.timeSimInit += Abc_Clock() - clk;

    // prepare user's AIG
    Gia_ManFillValue(pAig);
    Gia_ManConst0(pAig).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pAig.vCis)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
        pObj.Value = Gia_Obj2Lit( p.pFraig, Gia_ManCi(p.pFraig, i) );
    }
//    Gia_ManLevelNum(pAig);
    // prepare swept AIG (should be done after starting SAT solver in Ssc_ManStart)
    Gia_ManHashStart( p.pFraig );
    // perform sweeping
    Ssc_GiaResetPiPattern( pAig, pPars.nWords );
    Ssc_GiaSavePiPattern( pAig, p.vPivot );
    // inlining of macro Gia_ManForEachCand
    for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {      if !Gia_ObjIsCand(pObj) {} else {
        if pAig.iPatsPi == 64 * pPars.nWords {
clk = Abc_Clock();
            Ssc_GiaSimRound( pAig );
            Ssc_GiaClassesRefine( pAig );
            if pPars.fVerbose { 
                Gia_ManEquivPrintClasses( pAig, 0, 0 ); }
            Ssc_GiaClassesCheckPairs( pAig, p.vDisPairs );
            Vec_IntClear( p.vDisPairs );
            // prepare next patterns
            Ssc_GiaResetPiPattern( pAig, pPars.nWords );
            Ssc_GiaSavePiPattern( pAig, p.vPivot );
p.timeSimSat += Abc_Clock() - clk;
//printf( "\n" );
        }
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( p.pFraig, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        if !Gia_ObjHasRepr(pAig, i) {
            continue; }
        pRepr = Gia_ObjReprObj(pAig, i);
        if int32(pObj.Value) == Abc_LitNotCond( pRepr.Value, pRepr.fPhase ^ pObj.fPhase ) {
            Gia_ObjSetProved( pAig, i );
            continue;
        }
        assert(Abc_Lit2Var(pRepr.Value) != Abc_Lit2Var(pObj.Value), "Abc_Lit2Var(pRepr.Value) != Abc_Lit2Var(pObj.Value)");
        fCompl = pRepr.fPhase ^ pObj.fPhase ^ Abc_LitIsCompl(pRepr.Value) ^ Abc_LitIsCompl(pObj.Value);

        // perform SAT call
clk = Abc_Clock();
        p.nSatCalls++;
        status = Ssc_ManCheckEquivalence( p, Abc_Lit2Var(pRepr.Value), Abc_Lit2Var(pObj.Value), fCompl );
        if status == l_False {
            p.nSatCallsUnsat++;
            pObj.Value = Abc_LitNotCond( pRepr.Value, pRepr.fPhase ^ pObj.fPhase );
            Gia_ObjSetProved( pAig, i );
        
        } else if status == l_True {
            p.nSatCallsSat++;
            Ssc_GiaSavePiPattern( pAig, p.vPattern );
            Vec_IntPush( p.vDisPairs, Gia_ObjRepr(p.pAig, i) );
            Vec_IntPush( p.vDisPairs, i );
//            printf( "Try %2d and %2d: ", Gia_ObjRepr(p->pAig, i), i );
//            Vec_IntPrint( p->vPattern );
            if Gia_ObjRepr(p.pAig, i) > 0 {
                Ssc_GiaResimulateOneClass( p, Gia_ObjRepr(p.pAig, i), i ); }
        
        } else if status == l_Undef {
            p.nSatCallsUndec++;
        } else { assert(false, "0"); }
p.timeSat += Abc_Clock() - clk;
    }
    }
    if pAig.iPatsPi > 1 {
clk = Abc_Clock();
        for  pAig.iPatsPi < 64 * pPars.nWords  {
            Ssc_GiaSavePiPattern( pAig, p.vPivot );
        }
        Ssc_GiaSimRound( pAig );
        Ssc_GiaClassesRefine( pAig );
        if pPars.fVerbose { 
            Gia_ManEquivPrintClasses( pAig, 0, 0 ); }
        Ssc_GiaClassesCheckPairs( pAig, p.vDisPairs );
        Vec_IntClear( p.vDisPairs );
p.timeSimSat += Abc_Clock() - clk;
    }
//    Gia_ManEquivPrintClasses( pAig, 1, 0 );
//    Gia_ManPrint( pAig );

    // generate the resulting AIG
    pResult = Gia_ManEquivReduce( pAig, 0, 0, 1, 0 );
    if pResult == nil {
        fmt.Printf("There is no equivalences.\n");
        pAig.pReprs = nil;
        pAig.pNexts = nil;
        pResult = Gia_ManDup( pAig );
    }
    pResult = Gia_ManCleanup( pTemp = pResult );
    Gia_ManStop( pTemp );
    p.timeTotal = Abc_Clock() - clkTotal;
    if pPars.fVerbose {
        Ssc_ManPrintStats( p ); }
    Ssc_ManStop( p );
    // count the number of representatives
    if pPars.fVerbose {
        Abc_Print( 1, "Reduction in AIG nodes:%8d  ->%8d (%6.2f %%).  ", 
            Gia_ManAndNum(pAig), Gia_ManAndNum(pResult), 
            100.0 - 100.0 * Gia_ManAndNum(pResult) / Gia_ManAndNum(pAig) );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
    }
    return pResult;
}
func Ssc_PerformSweeping(pAig *Gia_Man_t, pCare *Gia_Man_t, pPars *Ssc_Pars_t) *Gia_Man_t {
    var pResult *Gia_Man_t  = Ssc_PerformSweepingInt( pAig, pCare, pPars );
    if pPars.fVerify {
        Ssc_PerformVerification( pAig, pResult, pCare ); }
    return pResult;
}
func Ssc_PerformSweepingConstr(p *Gia_Man_t, pPars *Ssc_Pars_t) *Gia_Man_t {
    var pAig, pCare, pResult *Gia_Man_t;
    var i int32;
    if pPars.fVerbose {
        Abc_Print( 0, "SAT sweeping AIG with %d constraints.\n", p.nConstrs ); }
    if p.nConstrs == 0 {
        pAig = Gia_ManDup( p );
        pCare = Gia_ManStart( Gia_ManCiNum(p) + 2 );
        pCare.pName = Abc_UtilStrsav( "care" );
        for  i = 0; i < Gia_ManCiNum(p); i++  {
            Gia_ManAppendCi( pCare );
        }
        Gia_ManAppendCo( pCare, 0 );
    
    } else {
        var vOuts *Vec_Int_t  = Vec_IntStartNatural( Gia_ManPoNum(p) );
        pAig = Gia_ManDupCones( p, Vec_IntArray(vOuts), Gia_ManPoNum(p) - p.nConstrs, 0 );
        pCare = Gia_ManDupCones( p, Vec_IntArray(vOuts) + Gia_ManPoNum(p) - p.nConstrs, p.nConstrs, 0 );
        Vec_IntFree( vOuts );
    }
    if pPars.fVerbose {
        fmt.Printf("User AIG: ");
        Gia_ManPrintStats( pAig, nil );
        fmt.Printf("Care AIG: ");
        Gia_ManPrintStats( pCare, nil );
    }

    pAig = Gia_ManDupLevelized( pResult = pAig );
    Gia_ManStop( pResult );
    pResult = Ssc_PerformSweeping( pAig, pCare, pPars );
    if pPars.fAppend {
        Gia_ManDupAppendShare( pResult, pCare );
        pResult.nConstrs = Gia_ManPoNum(pCare);
    }
    Gia_ManStop( pAig );
    Gia_ManStop( pCare );
    return pResult;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

