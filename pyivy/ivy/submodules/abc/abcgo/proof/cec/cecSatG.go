package cec

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [cecSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Detection of structural isomorphism.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "misc/util/utilTruth.h"
//#include "sat/glucose/AbcGlucose.h"
//#include "cec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// sweeping manager
   

// SAT solving manager
   
type Cec3_Man_t struct {
    pPars      *Cec3_Par_t;          // parameters
    pAig       *Gia_Man_t;           // user's AIG
    pNew       *Gia_Man_t;           // internal AIG
    // SAT solving
    pSat *bmcg_sat_solver;           // SAT solver
    vFrontier       *Vec_Ptr_t;      // CNF construction
    vFanins       *Vec_Ptr_t;        // CNF construction
    vSims       *Vec_Wrd_t;          // CI simulation info
    vNodesNew       *Vec_Int_t;      // nodes
    vSatVars       *Vec_Int_t;       // nodes
    vObjSatPairs       *Vec_Int_t;   // nodes
    vCexTriples       *Vec_Int_t;    // nodes
    // statistics
    nPatterns              int32;
    nSatSat              int32;
    nSatUnsat              int32;
    nSatUndec              int32;
    timeSatSat          abctime;
    timeSatUnsat          abctime;
    timeSatUndec          abctime;
    timeSim          abctime;
    timeRefine          abctime;
    timeExtra          abctime;
    timeStart          abctime;
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets parameter defaults.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_ManCreate(pAig *Gia_Man_t, pPars *Cec3_Par_t) *Cec3_Man_t {
    
    var pObj *Gia_Obj_t; var i int32;
    //satoko_opts_t Pars;
    //assert( Gia_ManRegNum(pAig) == 0 );
    // inlining of macro ABC_CALLOC
    p := &Cec3_Man_t{};
    memset( p, 0, sizeof(Cec3_Man_t) );
    p.timeStart    = Abc_Clock();
    p.pPars        = pPars;
    p.pAig         = pAig;
    // create new manager
    p.pNew         = Gia_ManStart( Gia_ManObjNum(pAig) );
    Gia_ManFillValue( pAig );
    Gia_ManConst0(pAig).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pAig.vCis)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
        pObj.Value = Gia_ManAppendCi( p.pNew );
    }
    Gia_ManHashAlloc( p.pNew );
    Vec_IntFill( &p.pNew.vCopies2, Gia_ManObjNum(p.pNew), -1 );
    // SAT solving
    //memset( &Pars, 0, sizeof(satoko_opts_t) );
    p.pSat         = bmcg_sat_solver_start();
    p.vFrontier    = Vec_PtrAlloc( 1000 );
    p.vFanins      = Vec_PtrAlloc( 100 );
    p.vNodesNew    = Vec_IntAlloc( 100 );
    p.vSatVars     = Vec_IntAlloc( 100 );
    p.vObjSatPairs = Vec_IntAlloc( 100 );
    p.vCexTriples  = Vec_IntAlloc( 100 );
    //Pars.conf_limit = pPars->nConfLimit;
    //satoko_configure(p->pSat, &Pars);
    // remember pointer to the solver in the AIG manager
    pAig.pData     = p.pSat;
    return p;
}
func Cec3_ManDestroy(p *Cec3_Man_t) {
    if p.pPars.fVerbose {
        var timeTotal abctime  = Abc_Clock() - p.timeStart;
        var timeSat abctime    = p.timeSatSat + p.timeSatUnsat + p.timeSatUndec;
        var timeOther abctime  = timeTotal - timeSat - p.timeSim - p.timeRefine - p.timeExtra;
//        Abc_Print( 1, "%d\n", p->Num );
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("SAT solving"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeSat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeSat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  sat      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  unsat    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUnsat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  fail     "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUndec)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUndec)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Simulation "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSim)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Refinement "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeRefine)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeRefine)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Rollback   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeExtra)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeExtra)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeOther)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeOther)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("TOTAL      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeTotal)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeTotal)))/(float64((timeTotal))) : 0.0);;
        fflush( os.Stdout );
    }

    Vec_WrdFreeP( &p.pAig.vSims );
    //Vec_WrdFreeP( &p->pAig->vSimsPi );
    Gia_ManCleanMark01( p.pAig );
    bmcg_sat_solver_stop( p.pSat );
    Gia_ManStopP( &p.pNew );
    Vec_PtrFreeP( &p.vFrontier );
    Vec_PtrFreeP( &p.vFanins );
    Vec_IntFreeP( &p.vNodesNew );
    Vec_IntFreeP( &p.vSatVars );
    Vec_IntFreeP( &p.vObjSatPairs );
    Vec_IntFreeP( &p.vCexTriples );
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Verify counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_ManVerify_rec(p *Gia_Man_t, iObj int32, pSat *bmcg_sat_solver) int32 {
    var Value0, Value1 int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if iObj == 0 { return 0; }
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return pObj.fMark1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if Gia_ObjIsCi(pObj) {
//        return pObj->fMark1 = satoko_var_polarity(pSat, Cec3_ObjSatId(p, pObj)) == SATOKO_LIT_TRUE;
        return pObj.fMark1 = bmcg_sat_solver_read_cex_varvalue(pSat, Cec3_ObjSatId(p, pObj)); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Cec3_ManVerify_rec( p, Gia_ObjFaninId0(pObj, iObj), pSat ) ^ Gia_ObjFaninC0(pObj);
    Value1 = Cec3_ManVerify_rec( p, Gia_ObjFaninId1(pObj, iObj), pSat ) ^ Gia_ObjFaninC1(pObj);
    return pObj.fMark1 = Value0 & Value1;
}
func Cec3_ManVerify(p *Gia_Man_t, iObj0 int32, iObj1 int32, fPhase int32, pSat *bmcg_sat_solver) {
//    int val0 = satoko_var_polarity(pSat, Cec3_ObjSatId(p, Gia_ManObj(p, iObj0))) == SATOKO_LIT_TRUE;
//    int val1 = satoko_var_polarity(pSat, Cec3_ObjSatId(p, Gia_ManObj(p, iObj1))) == SATOKO_LIT_TRUE;
    var Value0, Value1 int32;
    Gia_ManIncrementTravId( p );
    Value0 = Cec3_ManVerify_rec( p, iObj0, pSat );
    Value1 = Cec3_ManVerify_rec( p, iObj1, pSat );
    if (Value0 ^ Value1) == fPhase {
        fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", iObj0, iObj1); }
//    else
//        printf( "CEX verification succeeded for obj %d and obj %d.\n", iObj0, iObj1 );;
}

/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_ManCollect_rec(p *Cec3_Man_t, iObj int32) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p.pNew, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p.pNew, iObj);
    pObj = Gia_ManObj( p.pNew, iObj );
    if Cec3_ObjSatId(p.pNew, pObj) >= 0 {
        Vec_IntPush( p.vNodesNew, iObj );
        Vec_IntPush( p.vSatVars, Cec3_ObjSatId(p.pNew, pObj) );
    }
    if !(iObj != 0) {
        return; }
    if Gia_ObjIsAnd(pObj) {
        Cec3_ManCollect_rec( p, Gia_ObjFaninId0(pObj, iObj) );
        Cec3_ManCollect_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    
    } else {
        assert(Cec3_ObjSatId(p.pNew, pObj) >= 0, "Cec3_ObjSatId(p.pNew, pObj) >= 0");
        Vec_IntPushTwo( p.vObjSatPairs, Gia_ManCiIdToId(p.pAig, Gia_ObjCioId(pObj)), Cec3_ObjSatId(p.pNew, pObj) ); // SAT var
    }
}
func Cec3_ManSolveTwo(p *Cec3_Man_t, iObj0 int32, iObj1 int32, fPhase int32) int32 {
    var pObj *Gia_Obj_t;
    var status int32;
    var i int32;
    var iVar0 int32;
    var iVar1 int32;
    var Lits [2]int32;
    if iObj1 < iObj0 { 
        iObj1 ^= iObj0, iObj0 ^= iObj1, iObj1 ^= iObj0; }
    assert(iObj0 < iObj1, "iObj0 < iObj1");
    assert(p.pPars.fUseCones || bmcg_sat_solver_varnum(p.pSat) == 0, "p.pPars.fUseCones || bmcg_sat_solver_varnum(p.pSat) == 0");
    if !(iObj0 != 0) && Cec3_ObjSatId(p.pNew, Gia_ManConst0(p.pNew)) == -1 {
        Cec3_ObjSetSatId( p.pNew, Gia_ManConst0(p.pNew), bmcg_sat_solver_addvar(p.pSat) ); }
    iVar0 = Cec3_ObjGetCnfVar( p, iObj0 );
    iVar1 = Cec3_ObjGetCnfVar( p, iObj1 );
    // collect inputs and internal nodes
    Vec_IntClear( p.vNodesNew );
    Vec_IntClear( p.vSatVars );
    Vec_IntClear( p.vObjSatPairs );
    Gia_ManIncrementTravId( p.pNew );
    Cec3_ManCollect_rec( p, iObj0 );
    Cec3_ManCollect_rec( p, iObj1 );
//printf( "%d ", Vec_IntSize(p->vNodesNew) );
    // solve direct
    if p.pPars.fUseCones {  satoko_mark_cone( p.pSat, Vec_IntArray(p.vSatVars), Vec_IntSize(p.vSatVars) ); }
//    satoko_assump_push( p->pSat, Abc_Var2Lit(iVar0, 1) );
//    satoko_assump_push( p->pSat, Abc_Var2Lit(iVar1, fPhase) );
//    status = satoko_solve( p->pSat );
//    satoko_assump_pop( p->pSat );
//    satoko_assump_pop( p->pSat );
    Lits[0] = Abc_Var2Lit(iVar0, 1);
    Lits[1] = Abc_Var2Lit(iVar1, fPhase);
    bmcg_sat_solver_set_conflict_budget( p.pSat, p.pPars.nConfLimit );
    status = bmcg_sat_solver_solve( p.pSat, Lits, 2 );
    if status == GLUCOSE_UNSAT && iObj0 > 0 {
        // solve reverse
//        satoko_assump_push( p->pSat, Abc_Var2Lit(iVar0, 0) );
//        satoko_assump_push( p->pSat, Abc_Var2Lit(iVar1, !fPhase) );
//        status = satoko_solve( p->pSat );
//        satoko_assump_pop( p->pSat );
//        satoko_assump_pop( p->pSat );
        Lits[0] = Abc_Var2Lit(iVar0, 0);
        Lits[1] = Abc_Var2Lit(iVar1, !fPhase);
        bmcg_sat_solver_set_conflict_budget( p.pSat, p.pPars.nConfLimit );
        status = bmcg_sat_solver_solve( p.pSat, Lits, 2 );
    }
    if p.pPars.fUseCones {  satoko_unmark_cone( p.pSat, Vec_IntArray(p.vSatVars), Vec_IntSize(p.vSatVars) ); }
    //if ( status == SATOKO_SAT )
    //    Cec3_ManVerify( p->pNew, iObj0, iObj1, fPhase, p->pSat );
    if p.pPars.fUseCones {
        return status; }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vNodesNew)) && ((pObj) = Gia_ManObj(p.pNew, Vec_IntEntry(p.vNodesNew,i))); i++  {
        Cec3_ObjCleanSatId( p.pNew, pObj );
    }
    return status;
}

func Cec3_ManSweepNode(p *Cec3_Man_t, iObj int32) int32 {
    var clk abctime  = Abc_Clock();
    var i int32;
    var IdAig int32;
    var IdSat int32;
    var status int32;
    var RetValue int32 = 1;
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pAig, iObj );
    var pRepr *Gia_Obj_t  = Gia_ObjReprObj( p.pAig, iObj );
    var fCompl int32  = Abc_LitIsCompl(pObj.Value) ^ Abc_LitIsCompl(pRepr.Value) ^ pObj.fPhase ^ pRepr.fPhase;
    status = Cec3_ManSolveTwo( p, Abc_Lit2Var(pRepr.Value), Abc_Lit2Var(pObj.Value), fCompl );
    if status == GLUCOSE_SAT {
        p.nSatSat++;
        p.nPatterns++;
        if p.pAig.iPatsPi == 64 * p.pAig.nSimWords - 1 {
            p.pAig.iPatsPi = 1;
        } else {
            p.pAig.iPatsPi = p.pAig.iPatsPi + 1;
        }
        assert(p.pAig.iPatsPi > 0 && p.pAig.iPatsPi < 64 * p.pAig.nSimWords, "p.pAig.iPatsPi > 0 && p.pAig.iPatsPi < 64 * p.pAig.nSimWords");
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(p.vObjSatPairs)) && (((IdAig) = Vec_IntEntry(p.vObjSatPairs, i)), 1) && (((IdSat) = Vec_IntEntry(p.vObjSatPairs, i+1)), 1); i += 2  {
//            Cec3_ObjSimSetInputBit( p->pAig, IdAig, satoko_var_polarity(p->pSat, IdSat) == SATOKO_LIT_TRUE );
            Cec3_ObjSimSetInputBit( p.pAig, IdAig, bmcg_sat_solver_read_cex_varvalue(p.pSat, IdSat) );
        }
        p.timeSatSat += Abc_Clock() - clk;
        RetValue = 0;
    
    } else if status == GLUCOSE_UNSAT {
        p.nSatUnsat++;
        pObj.Value = Abc_LitNotCond( pRepr.Value, fCompl );
        Gia_ObjSetProved( p.pAig, iObj );
        p.timeSatUnsat += Abc_Clock() - clk;
        RetValue = 1;
    
    } else {
        p.nSatUndec++;
        assert(status == GLUCOSE_UNDEC, "status == GLUCOSE_UNDEC");
        Gia_ObjSetFailed( p.pAig, iObj );
        p.timeSatUndec += Abc_Clock() - clk;
        RetValue = 2;
    }
    if p.pPars.fUseCones {
        return RetValue; }
    clk = Abc_Clock();
    bmcg_sat_solver_reset( p.pSat );
    p.timeExtra += Abc_Clock() - clk;
//    satoko_stats(p->pSat)->n_conflicts = 0;
    return RetValue;
}
func Cec3_ManPrintStats(p *Gia_Man_t, pPars *Cec3_Par_t, pMan *Cec3_Man_t) {
    if !pPars.fVerbose {
        return; }
    fmt.Printf("S =%5d ", pMan ? pMan.nSatSat   : 0);
    fmt.Printf("U =%5d ", pMan ? pMan.nSatUnsat : 0);
    fmt.Printf("F =%5d ", pMan ? pMan.nSatUndec : 0);
    Gia_ManEquivPrintClasses( p, pPars.fVeryVerbose, 0 );
}
func Cec3_ManPerformSweeping(p *Gia_Man_t, pPars *Cec3_Par_t, ppNew **Gia_Man_t) int32 {
    var pMan *Cec3_Man_t  = Cec3_ManCreate( p, pPars ); 
    var pObj, pRepr, pObjNew *Gia_Obj_t; 
    var i int32;
    var Iter int32;
    var fDisproved int32 = 1;

    // check if any output trivially fails under all-0 pattern
    Gia_ManRandomW( 1 );
    Gia_ManSetPhase( p );
    if pPars.fIsMiter {
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if pObj.fPhase {
                p.pCexSeq = Cec3_ManDeriveCex( p, i, -1 );
                goto finalize;
            }
        }
    }

    // simulate one round and create classes
    Cec3_ManSimAlloc( p, pPars.nSimWords );
    Cec3_ManSimulateCis( p );
    Cec3_ManSimulate( p, nil, pMan );
    if pPars.fIsMiter && !(Cec3_ManSimulateCos(p) != 0) { // cex detected
        goto finalize; }
    Cec3_ManCreateClasses( p, pMan );
    Cec3_ManPrintStats( p, pPars, pMan );

    // perform additinal simulation
    for  i = 0; i < pPars.nSimRounds; i++  {
        Cec3_ManSimulateCis( p );
        Cec3_ManSimulate( p, nil, pMan );
        if pPars.fIsMiter && !(Cec3_ManSimulateCos(p) != 0) { // cex detected
            goto finalize; }
        Cec3_ManPrintStats( p, pPars, pMan );
    }
    // perform sweeping
    //pMan = Cec3_ManCreate( p, pPars );
    for  Iter = 0; (fDisproved != 0) && Iter < pPars.nItersMax; Iter++  {
        fDisproved = 0;
        pMan.nPatterns = 0;
        Cec3_ManSimulateCis( p );
        Vec_IntClear( pMan.vCexTriples );
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            if ~pObj.Value || Gia_ObjFailed(p, i) { // skip swept nodes and failed nodes
                continue; }
            if !~Gia_ObjFanin0(pObj).Value || !~Gia_ObjFanin1(pObj).Value { // skip fanouts of non-swept nodes
                continue; }
            assert(!Gia_ObjProved(p, i) && !Gia_ObjFailed(p, i), "!Gia_ObjProved(p, i) && !Gia_ObjFailed(p, i)");
            // duplicate the node
            pObj.Value = Gia_ManHashAnd( pMan.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            if Vec_IntSize(&pMan.pNew.vCopies2) == Abc_Lit2Var(pObj.Value) {
                pObjNew = Gia_ManObj( pMan.pNew, Abc_Lit2Var(pObj.Value) );
                pObjNew.fMark0 = Gia_ObjIsMuxType( pObjNew );
                Gia_ObjSetPhase( pMan.pNew, pObjNew );
                Vec_IntPush( &pMan.pNew.vCopies2, -1 );
            }
            assert(Vec_IntSize(&pMan.pNew.vCopies2) == Gia_ManObjNum(pMan.pNew), "Vec_IntSize(&pMan.pNew.vCopies2) == Gia_ManObjNum(pMan.pNew)");
            pRepr = Gia_ObjReprObj( p, i );
            if pRepr == nil || !~pRepr.Value {
                continue; }
            if Abc_Lit2Var(pObj.Value) == Abc_Lit2Var(pRepr.Value) {
                assert((pObj.Value ^ pRepr.Value) == (pObj.fPhase ^ pRepr.fPhase), "(pObj.Value ^ pRepr.Value) == (pObj.fPhase ^ pRepr.fPhase)");
                Gia_ObjSetProved( p, i );
                continue;
            }
            if Cec3_ManSweepNode(pMan, i) != 0 {
                if Gia_ObjProved(p, i) {
                    pObj.Value = Abc_LitNotCond( pRepr.Value, pObj.fPhase ^ pRepr.fPhase ); }
                continue;
            }
            pObj.Value = ~0;
            Vec_IntPushThree( pMan.vCexTriples, Gia_ObjId(p, pRepr), i, Abc_Var2Lit(p.iPatsPi, pObj.fPhase ^ pRepr.fPhase) );
            fDisproved = 1;
        }
        }
        if fDisproved != 0 {
            var Fails int32  = Cec3_ManSimulate( p, pMan.vCexTriples, pMan );
            if (Fails != 0) && pPars.fVerbose {
                fmt.Printf("Failed to resimulate %d times with pattern = %d  (total = %d).\n", Fails, pMan.nPatterns, pPars.nSimWords * 64); }
            if pPars.fIsMiter && !(Cec3_ManSimulateCos(p) != 0) { // cex detected
                break; }
        }
        Cec3_ManPrintStats( p, pPars, pMan );
    }
    // finish the AIG, if it is not finished
    if ppNew {
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
            if !~pObj.Value { 
                pObj.Value = Gia_ManHashAnd( pMan.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            pObj.Value = Gia_ManAppendCo( pMan.pNew, Gia_ObjFanin0Copy(pObj) );
        }
        *ppNew = Gia_ManCleanup( pMan.pNew );
        (*ppNew).pName = Abc_UtilStrsav( p.pName );
        (*ppNew).pSpec = Abc_UtilStrsav( p.pSpec );
    }
finalize:
    Cec3_ManDestroy( pMan );
    //Gia_ManEquivPrintClasses( p, 1, 0 );
    if p.pCexSeq {
        return 0;
    } else {
        return 1;
    }
}
func Cec3_ManSimulateTest(p *Gia_Man_t, pPars0 *Cec_ParFra_t) *Gia_Man_t {
    var pNew *Gia_Man_t  = nil;
    //abctime clk = Abc_Clock();
    var Pars Cec3_Par_t;
    var pPars *Cec3_Par_t = &Pars;
    Cec3_SetDefaultParams( pPars );
    // set resource limits
//    pPars->nSimWords  = pPars0->nWords;     // simulation words
//    pPars->nSimRounds = pPars0->nRounds;    // simulation rounds
//    pPars->nItersMax  = pPars0->nItersMax;  // max number of iterations
    pPars.nConfLimit = pPars0.nBTLimit;   // conflict limit at a node
    pPars.fUseCones  = pPars0.fUseCones;
    pPars.fVerbose   = pPars0.fVerbose;
//    Gia_ManComputeGiaEquivs( p, 100000, 0 );
//    Gia_ManEquivPrintClasses( p, 1, 0 );
    Cec3_ManPerformSweeping( p, pPars, &pNew );
    //Abc_PrintTime( 1, "SAT sweeping time", Abc_Clock() - clk );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

