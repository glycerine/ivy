package saig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigGlaPba.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Gate level abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigGlaPba.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "saig.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Aig_Gla2Man_t struct {
    // user data
    pAig     *Aig_Man_t;
    nStart            int32;
    nFramesMax            int32;
    fVerbose            int32;
    // unrolling
    vObj2Vec     *Vec_Int_t;   // maps obj ID into its vec ID
    vVec2Var     *Vec_Int_t;   // maps vec ID into its sat Var (nFrames per vec ID)
    vVar2Inf     *Vec_Int_t;   // maps sat Var into its frame and obj ID
    // clause mapping
    vCla2Obj     *Vec_Int_t;   // maps clause into its root object
    vCla2Fra     *Vec_Int_t;   // maps clause into its frame
    vVec2Use     *Vec_Int_t;   // maps vec ID into its used frames (nFrames per vec ID)
    // SAT solver
    pSat    *sat_solver;
    // statistics
    timePre        int64;
    timeSat        int64;
    timeTotal        int64;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Adds constant to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2AddConst(pSat *sat_solver, iVar int32, fCompl int32) int32 {
    var Lit lit  = toLitCond( iVar, fCompl );
    if !sat_solver_addclause( pSat, &Lit, &Lit + 1 ) {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds buffer to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2AddBuffer(pSat *sat_solver, iVar0 int32, iVar1 int32, fCompl int32) int32 {
    var Lits [2]lit;

    Lits[0] = toLitCond( iVar0, 0 );
    Lits[1] = toLitCond( iVar1, !fCompl );
    if !sat_solver_addclause( pSat, Lits, Lits + 2 ) {
        return 0; }

    Lits[0] = toLitCond( iVar0, 1 );
    Lits[1] = toLitCond( iVar1, fCompl );
    if !sat_solver_addclause( pSat, Lits, Lits + 2 ) {
        return 0; }

    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds buffer to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2AddNode(pSat *sat_solver, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32) int32 {
    var Lits [3]lit;

    Lits[0] = toLitCond( iVar, 1 );
    Lits[1] = toLitCond( iVar0, fCompl0 );
    if !sat_solver_addclause( pSat, Lits, Lits + 2 ) {
        return 0; }

    Lits[0] = toLitCond( iVar, 1 );
    Lits[1] = toLitCond( iVar1, fCompl1 );
    if !sat_solver_addclause( pSat, Lits, Lits + 2 ) {
        return 0; }

    Lits[0] = toLitCond( iVar, 0 );
    Lits[1] = toLitCond( iVar0, !fCompl0 );
    Lits[2] = toLitCond( iVar1, !fCompl1 );
    if !sat_solver_addclause( pSat, Lits, Lits + 3 ) {
        return 0; }

    return 1;
}


/**Function*************************************************************

  Synopsis    [Finds existing SAT variable or creates a new one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2FetchVar(p *Aig_Gla2Man_t, pObj *Aig_Obj_t, k int32) int32 {
    var i, iVecId, iSatVar int32;
    assert(k < p.nFramesMax, "k < p.nFramesMax");
    iVecId = Vec_IntEntry( p.vObj2Vec, Aig_ObjId(pObj) );
    if iVecId == 0 {
        iVecId = Vec_IntSize( p.vVec2Var ) / p.nFramesMax;
        for  i = 0; i < p.nFramesMax; i++  {
            Vec_IntPush( p.vVec2Var, 0 );
        }
        Vec_IntWriteEntry( p.vObj2Vec, Aig_ObjId(pObj), iVecId );
    }
    iSatVar = Vec_IntEntry( p.vVec2Var, iVecId * p.nFramesMax + k );
    if iSatVar == 0 {
        iSatVar = Vec_IntSize( p.vVar2Inf ) / 2;
        Vec_IntPush( p.vVar2Inf, Aig_ObjId(pObj) );
        Vec_IntPush( p.vVar2Inf, k );
        Vec_IntWriteEntry( p.vVec2Var, iVecId * p.nFramesMax + k, iSatVar );
    }
    return iSatVar;
}

/**Function*************************************************************

  Synopsis    [Assigns variables to the AIG nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2AssignVars_rec(p *Aig_Gla2Man_t, pObj *Aig_Obj_t, f int32) {
    var nVars int32  = Vec_IntSize(p.vVar2Inf);
    Aig_Gla2FetchVar( p, pObj, f );
    if nVars == Vec_IntSize(p.vVar2Inf) {
        return; }
    if Aig_ObjIsConst1(pObj) {
        return; }
    if Saig_ObjIsPo( p.pAig, pObj ) {
        Aig_Gla2AssignVars_rec( p, Aig_ObjFanin0(pObj), f );
        return;
    }
    if Aig_ObjIsCi( pObj ) {
        if Saig_ObjIsLo(p.pAig, pObj) && f > 0 {
            Aig_Gla2AssignVars_rec( p, Aig_ObjFanin0( Saig_ObjLoToLi(p.pAig, pObj) ), f-1 ); }
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Aig_Gla2AssignVars_rec( p, Aig_ObjFanin0(pObj), f );
    Aig_Gla2AssignVars_rec( p, Aig_ObjFanin1(pObj), f );
}

/**Function*************************************************************

  Synopsis    [Creates SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2CreateSatSolver(p *Aig_Gla2Man_t) int32 {
    var vPoLits *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    var f int32;
    var ObjId int32;
    var nVars int32;
    var RetValue int32 = 1;

    // assign variables
    for  f = p.nFramesMax - 1; f >= 0; f--  {
//    for ( f = 0; f < p->nFramesMax; f++ )
        Aig_Gla2AssignVars_rec( p, Aig_ManCo(p.pAig, 0), f );
    }

    // create SAT solver
    p.pSat = sat_solver_new();
    sat_solver_store_alloc( p.pSat ); 
    sat_solver_setnvars( p.pSat, Vec_IntSize(p.vVar2Inf)/2 );

    // add clauses
    nVars = Vec_IntSize( p.vVar2Inf );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vVar2Inf)) && (((ObjId) = Vec_IntEntry(p.vVar2Inf, i)), 1) && (((f) = Vec_IntEntry(p.vVar2Inf, i+1)), 1); i += 2  {
        if ObjId == -1 {
            continue; }
        pObj = Aig_ManObj( p.pAig, ObjId );
        if Aig_ObjIsNode(pObj) {
            RetValue &= Aig_Gla2AddNode( p.pSat, Aig_Gla2FetchVar(p, pObj, f), 
                                                  Aig_Gla2FetchVar(p, Aig_ObjFanin0(pObj), f), 
                                                  Aig_Gla2FetchVar(p, Aig_ObjFanin1(pObj), f), 
                                                  Aig_ObjFaninC0(pObj), Aig_ObjFaninC1(pObj) );
            Vec_IntPush( p.vCla2Obj, ObjId );
            Vec_IntPush( p.vCla2Obj, ObjId );
            Vec_IntPush( p.vCla2Obj, ObjId );

            Vec_IntPush( p.vCla2Fra, f );
            Vec_IntPush( p.vCla2Fra, f );
            Vec_IntPush( p.vCla2Fra, f );
        
        } else if Saig_ObjIsLo(p.pAig, pObj) {
            if f == 0 {
                RetValue &= Aig_Gla2AddConst( p.pSat, Aig_Gla2FetchVar(p, pObj, f), 1 );
                Vec_IntPush( p.vCla2Obj, ObjId );

                Vec_IntPush( p.vCla2Fra, f );
            
            } else {
                var pObjLi *Aig_Obj_t  = Saig_ObjLoToLi(p.pAig, pObj);
                RetValue &= Aig_Gla2AddBuffer( p.pSat, Aig_Gla2FetchVar(p, pObj, f), 
                                                        Aig_Gla2FetchVar(p, Aig_ObjFanin0(pObjLi), f-1), 
                                                        Aig_ObjFaninC0(pObjLi) );
                Vec_IntPush( p.vCla2Obj, ObjId );
                Vec_IntPush( p.vCla2Obj, ObjId );

                Vec_IntPush( p.vCla2Fra, f );
                Vec_IntPush( p.vCla2Fra, f );
            }
        
        } else if Saig_ObjIsPo(p.pAig, pObj) {
            RetValue &= Aig_Gla2AddBuffer( p.pSat, Aig_Gla2FetchVar(p, pObj, f), 
                                                    Aig_Gla2FetchVar(p, Aig_ObjFanin0(pObj), f), 
                                                    Aig_ObjFaninC0(pObj) );
            Vec_IntPush( p.vCla2Obj, ObjId );
            Vec_IntPush( p.vCla2Obj, ObjId );

            Vec_IntPush( p.vCla2Fra, f );
            Vec_IntPush( p.vCla2Fra, f );
        
        } else if Aig_ObjIsConst1(pObj) {
            RetValue &= Aig_Gla2AddConst( p.pSat, Aig_Gla2FetchVar(p, pObj, f), 0 );
            Vec_IntPush( p.vCla2Obj, ObjId );

            Vec_IntPush( p.vCla2Fra, f );
        
        } else { assert(Saig_ObjIsPi(p.pAig, pObj), "Saig_ObjIsPi(p.pAig, pObj)"); }
    }

    // add output clause
    vPoLits = Vec_IntAlloc( p.nFramesMax );
    for  f = p.nStart; f < p.nFramesMax; f++  {
        Vec_IntPush( vPoLits, 2 * Aig_Gla2FetchVar(p, Aig_ManCo(p.pAig, 0), f) );
    }
    RetValue &= sat_solver_addclause( p.pSat, Vec_IntArray(vPoLits), Vec_IntArray(vPoLits) + Vec_IntSize(vPoLits) );
    Vec_IntFree( vPoLits );
    Vec_IntPush( p.vCla2Obj, 0 );
    Vec_IntPush( p.vCla2Fra, 0 );
    assert(Vec_IntSize(p.vCla2Fra) == Vec_IntSize(p.vCla2Obj), "Vec_IntSize(p.vCla2Fra) == Vec_IntSize(p.vCla2Obj)");

    assert(nVars == Vec_IntSize(p.vVar2Inf), "nVars == Vec_IntSize(p.vVar2Inf)");
    assert(((Sto_Man_t *)p.pSat.pStore).nClauses == Vec_IntSize(p.vCla2Obj), "((Sto_Man_t *)p.pSat.pStore).nClauses == Vec_IntSize(p.vCla2Obj)");
//    Sto_ManDumpClauses( ((Sto_Man_t *)p->pSat->pStore), "temp_sto.cnf" );
    sat_solver_store_mark_roots( p.pSat ); 

    if p.fVerbose {
        fmt.Printf("The resulting SAT problem contains %d variables and %d clauses.\n", p.pSat.size, p.pSat.stats.clauses); }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2ManStart(pAig *Aig_Man_t, nStart int32, nFramesMax int32, fVerbose int32) *Aig_Gla2Man_t {
    
    var i int32;

    // inlining of macro ABC_CALLOC
    p := &Aig_Gla2Man_t{};
    p.pAig       = pAig;

    p.vObj2Vec   = Vec_IntStart( Aig_ManObjNumMax(pAig) );
    p.vVec2Var   = Vec_IntAlloc( 1 << 20 );
    p.vVar2Inf   = Vec_IntAlloc( 1 << 20 );
    p.vCla2Obj   = Vec_IntAlloc( 1 << 20 );
    p.vCla2Fra   = Vec_IntAlloc( 1 << 20 );

    // skip first vector ID
    p.nStart     = nStart;
    p.nFramesMax = nFramesMax;
    p.fVerbose   = fVerbose;
    for  i = 0; i < p.nFramesMax; i++  {
        Vec_IntPush( p.vVec2Var, -1 );
    }

    // skip  first SAT variable
    Vec_IntPush( p.vVar2Inf, -1 );
    Vec_IntPush( p.vVar2Inf, -1 );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2ManStop(p *Aig_Gla2Man_t) {
    Vec_IntFreeP( &p.vObj2Vec );
    Vec_IntFreeP( &p.vVec2Var );
    Vec_IntFreeP( &p.vVar2Inf );
    Vec_IntFreeP( &p.vCla2Obj );
    Vec_IntFreeP( &p.vCla2Fra );
    Vec_IntFreeP( &p.vVec2Use );

    if p.pSat {
        sat_solver_delete( p.pSat ); }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Finds the set of clauses involved in the UNSAT core.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_AbsSolverUnsatCore(pSat *sat_solver, nConfMax int32, fVerbose int32, piRetValue *int32) *Vec_Int_t {
    var vCore *Vec_Int_t;
    var pSatCnf any; 
    var pManProof *Intp_Man_t;
    var RetValue int32;
    var clk int64  = time.Now().UnixNano();
    if piRetValue != nil {
        *piRetValue = -1; }
    // solve the problem
    RetValue = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfMax, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if RetValue == l_Undef {
        fmt.Printf("Conflict limit is reached.\n");
        return nil;
    }
    if RetValue == l_True {
        fmt.Printf("The BMC problem is SAT.\n");
        if piRetValue != nil {
            *piRetValue = 0; }
        return nil;
    }
    if fVerbose != 0 {
        fmt.Printf("SAT solver returned UNSAT after %7d conflicts.      ", int32(pSat.stats.conflicts));
        Abc_PrintTime( 1, "Time", time.Now().UnixNano() - clk );
    }
    assert(RetValue == l_False, "RetValue == l_False");
    pSatCnf = sat_solver_store_release( pSat ); 
    // derive the UNSAT core
    clk = time.Now().UnixNano();
    pManProof = Intp_ManAlloc();
    vCore = (Vec_Int_t *)Intp_ManUnsatCore( pManProof, (Sto_Man_t *)pSatCnf, 0, 0 );
    Intp_ManFree( pManProof );
    if fVerbose != 0 {
        fmt.Printf("SAT core contains %8d clauses (out of %8d).   ", Vec_IntSize(vCore), sat_solver_nclauses(pSat));
        Abc_PrintTime( 1, "Time", time.Now().UnixNano() - clk );
    }
    Sto_ManFree( (Sto_Man_t *)pSatCnf );
    return vCore;
}

/**Function*************************************************************

  Synopsis    [Collects abstracted objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2ManCollect(p *Aig_Gla2Man_t, vCore *Vec_Int_t) *Vec_Int_t {
    var vResult *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, ClaId, iVecId int32;
//    p->vVec2Use = Vec_IntStart( Vec_IntSize(p->vVec2Var) );

    vResult = Vec_IntStart( Aig_ManObjNumMax(p.pAig) );
    Vec_IntWriteEntry( vResult, 0, 1 ); // add const1
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCore)) && (((ClaId) = Vec_IntEntry(vCore, i)), 1); i++  {
        pObj = Aig_ManObj( p.pAig, Vec_IntEntry(p.vCla2Obj, ClaId) );
        if Saig_ObjIsPi(p.pAig, pObj) || Saig_ObjIsPo(p.pAig, pObj) || Aig_ObjIsConst1(pObj) {
            continue; }
        assert(Saig_ObjIsLo(p.pAig, pObj) || Aig_ObjIsNode(pObj), "Saig_ObjIsLo(p.pAig, pObj) || Aig_ObjIsNode(pObj)");
        Vec_IntWriteEntry( vResult, Aig_ObjId(pObj), 1 );
/*
        // add flop inputs with multiple fanouts
        if ( Saig_ObjIsLo(p->pAig, pObj) )
        {
            Aig_Obj_t * pObjLi = Saig_ObjLoToLi(p->pAig, pObj);
            if ( !Saig_ObjIsPi(p->pAig, Aig_ObjFanin0(pObjLi)) )
//            if ( Aig_ObjRefs( Aig_ObjFanin0(pObjLi) ) > 1 )
                Vec_IntWriteEntry( vResult, Aig_ObjFaninId0(pObjLi), 1 );
        }
        else
        {
            if ( !Saig_ObjIsPi(p->pAig, Aig_ObjFanin0(pObj)) )
                Vec_IntWriteEntry( vResult, Aig_ObjFaninId0(pObj), 1 );
            if ( !Saig_ObjIsPi(p->pAig, Aig_ObjFanin1(pObj)) )
                Vec_IntWriteEntry( vResult, Aig_ObjFaninId1(pObj), 1 );
        }
*/
        if p.vVec2Use {
            iVecId = Vec_IntEntry( p.vObj2Vec, Aig_ObjId(pObj) );
            Vec_IntWriteEntry( p.vVec2Use, iVecId * p.nFramesMax + Vec_IntEntry(p.vCla2Fra, ClaId), 1 );
        }
    }
//    printf( "Number of entries %d\n", Vec_IntCountPositive(vResult) );

    // count the number of objects in each frame
    if p.vVec2Use {
        var vCounts *Vec_Int_t  = Vec_IntStart( p.nFramesMax );
        var v int32;
        var f int32;
        var Entry int32;
        var nVecIds int32 = Vec_IntSize(p.vVec2Use) / p.nFramesMax;
        for  f = 0; f < p.nFramesMax; f++  {
        for  v = 0; v < nVecIds; v++  {
            if Vec_IntEntry( p.vVec2Use, v * p.nFramesMax + f ) {
                Vec_IntAddToEntry( vCounts, f, 1 ); }
        }
        }
        // inlining of macro Vec_IntForEachEntry
        for  f = 0; (f < Vec_IntSize(vCounts)) && (((Entry) = Vec_IntEntry(vCounts, f)), 1); f++  {
            fmt.Printf("%d ", Entry);
        }
        fmt.Printf("\n");
        Vec_IntFree( vCounts );
    }
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Performs gate-level localization abstraction.]

  Description [Returns array of objects included in the abstraction. This array
  may contain only const1, flop outputs, and internal nodes, that is, objects
  that should have clauses added to the SAT solver.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_Gla2ManPerform(pAig *Aig_Man_t, nStart int32, nFramesMax int32, nConfLimit int32, TimeLimit int32, fSkipRand int32, fVerbose int32) *Vec_Int_t {
    var p *Aig_Gla2Man_t;
    var vCore, vResult *Vec_Int_t;
    var nTimeToStop int64;
    if TimeLimit {
        nTimeToStop = TimeLimit * int64(time.Second) + time.Now().UnixNano();
    } else {
        nTimeToStop = 0;
    }
    var clk int64;
    var clk2 int64 = time.Now().UnixNano();
    assert(Saig_ManPoNum(pAig) == 1, "Saig_ManPoNum(pAig) == 1");

    if fVerbose != 0 {
        if TimeLimit {
            fmt.Printf("Abstracting from frame %d to frame %d with timeout %d sec.\n", nStart, nFramesMax, TimeLimit);
        } else {
            fmt.Printf("Abstracting from frame %d to frame %d with no timeout.\n", nStart, nFramesMax); }
    }

    // start the solver
    clk = time.Now().UnixNano();
    p = Aig_Gla2ManStart( pAig, nStart, nFramesMax, fVerbose );
    if !(Aig_Gla2CreateSatSolver( p ) != 0) {
        fmt.Printf("Error!  SAT solver became UNSAT.\n");
        Aig_Gla2ManStop( p );
        return nil;
    }
    sat_solver_set_random( p.pSat, fSkipRand );
    p.timePre += time.Now().UnixNano() - clk;

    // set runtime limit
    if TimeLimit {
        sat_solver_set_runtime_limit( p.pSat, nTimeToStop ); }

    // compute UNSAT core
    clk = time.Now().UnixNano();
    vCore = Saig_AbsSolverUnsatCore( p.pSat, nConfLimit, fVerbose, nil );
    if vCore == nil {
        Aig_Gla2ManStop( p );
        return nil;
    }
    p.timeSat += time.Now().UnixNano() - clk;
    p.timeTotal += time.Now().UnixNano() - clk2;

    // print stats
    if fVerbose != 0 {
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Pre   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timePre)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timePre)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Sat   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Total "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
    }

    // prepare return value
    vResult = Aig_Gla2ManCollect( p, vCore );
    Vec_IntFree( vCore );
    Aig_Gla2ManStop( p );
    return vResult;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

