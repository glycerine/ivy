package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigInter.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Interpolate two AIGs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigInter.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var timeCnf abctime;
var timeSat abctime;
var timeInt abctime;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManInterFast(pManOn *Aig_Man_t, pManOff *Aig_Man_t, fVerbose int32) {
    var pSat *sat_solver;
    var pCnfOn, pCnfOff *Cnf_Dat_t;
    var pObj, pObj2 *Aig_Obj_t;
    var Lits [3]int32;
    var status int32;
    var i int32;
//    abctime clk = Abc_Clock();

    assert(Aig_ManCiNum(pManOn) == Aig_ManCiNum(pManOff), "Aig_ManCiNum(pManOn) == Aig_ManCiNum(pManOff)");
    assert(Aig_ManCoNum(pManOn) == Aig_ManCoNum(pManOff), "Aig_ManCoNum(pManOn) == Aig_ManCoNum(pManOff)");

    // derive CNFs
    pManOn.nRegs = Aig_ManCoNum(pManOn);
    pCnfOn  = Cnf_Derive( pManOn, Aig_ManCoNum(pManOn) );
    pManOn.nRegs = 0;

    pManOff.nRegs = Aig_ManCoNum(pManOn);
    pCnfOff = Cnf_Derive( pManOff, Aig_ManCoNum(pManOff) );
    pManOff.nRegs = 0;

//    pCnfOn  = Cnf_DeriveSimple( pManOn, Aig_ManCoNum(pManOn) );
//    pCnfOff = Cnf_DeriveSimple( pManOff, Aig_ManCoNum(pManOn) );
    Cnf_DataLift( pCnfOff, pCnfOn.nVars );

    // start the solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, pCnfOn.nVars + pCnfOff.nVars );

    // add clauses of A
    for  i = 0; i < pCnfOn.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnfOn.pClauses[i], pCnfOn.pClauses[i+1] ) {
            Cnf_DataFree( pCnfOn );
            Cnf_DataFree( pCnfOff );
            sat_solver_delete( pSat );
            return;
        }
    }

    // add clauses of B
    for  i = 0; i < pCnfOff.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnfOff.pClauses[i], pCnfOff.pClauses[i+1] ) {
            Cnf_DataFree( pCnfOn );
            Cnf_DataFree( pCnfOff );
            sat_solver_delete( pSat );
            return;
        }
    }

    // add PI clauses
    // collect the common variables
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pManOn.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pManOn.vCis, i)), 1); i++  {
        pObj2 = Aig_ManCi( pManOff, i );

        Lits[0] = toLitCond( pCnfOn.pVarNums[pObj.Id], 0 );
        Lits[1] = toLitCond( pCnfOff.pVarNums[pObj2.Id], 1 );
        if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
            assert(false, "0"); }
        Lits[0] = toLitCond( pCnfOn.pVarNums[pObj.Id], 1 );
        Lits[1] = toLitCond( pCnfOff.pVarNums[pObj2.Id], 0 );
        if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
            assert(false, "0"); }
    }
    status = sat_solver_simplify( pSat );
    assert(status != 0, "status != 0");

    // solve incremental SAT problems
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pManOn.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pManOn.vCos, i)), 1); i++  {
        pObj2 = Aig_ManCo( pManOff, i );

        Lits[0] = toLitCond( pCnfOn.pVarNums[pObj.Id], 0 );
        Lits[1] = toLitCond( pCnfOff.pVarNums[pObj2.Id], 0 );
        status = sat_solver_solve( pSat, Lits, Lits+2, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if status != l_False {
            fmt.Printf("The incremental SAT problem is not UNSAT.\n"); }
    }
    Cnf_DataFree( pCnfOn );
    Cnf_DataFree( pCnfOff );
    sat_solver_delete( pSat );
//    ABC_PRT( "Fast interpolation time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManInter(pManOn *Aig_Man_t, pManOff *Aig_Man_t, fRelation int32, fVerbose int32) *Aig_Man_t {
    var pSatCnf any  = nil;
    var pManInter *Inta_Man_t; 
    var pRes *Aig_Man_t;
    var pSat *sat_solver;
    var pCnfOn, pCnfOff *Cnf_Dat_t;
    var vVarsAB *Vec_Int_t;
    var pObj, pObj2 *Aig_Obj_t;
    var Lits [3]int32;
    var status int32;
    var i int32;
    var clk abctime;
    var iLast int32  = -1; // Suppress "might be used uninitialized"

    assert(Aig_ManCiNum(pManOn) == Aig_ManCiNum(pManOff), "Aig_ManCiNum(pManOn) == Aig_ManCiNum(pManOff)");

clk = Abc_Clock();
    // derive CNFs
//    pCnfOn  = Cnf_Derive( pManOn, 0 );
//    pCnfOff = Cnf_Derive( pManOff, 0 );
    pCnfOn  = Cnf_DeriveSimple( pManOn, 0 );
    pCnfOff = Cnf_DeriveSimple( pManOff, 0 );
    Cnf_DataLift( pCnfOff, pCnfOn.nVars );
timeCnf += Abc_Clock() - clk;

clk = Abc_Clock();
    // start the solver
    pSat = sat_solver_new();
    sat_solver_store_alloc( pSat );
    sat_solver_setnvars( pSat, pCnfOn.nVars + pCnfOff.nVars );

    // add clauses of A
    for  i = 0; i < pCnfOn.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnfOn.pClauses[i], pCnfOn.pClauses[i+1] ) {
            Cnf_DataFree( pCnfOn );
            Cnf_DataFree( pCnfOff );
            sat_solver_delete( pSat );
            return nil;
        }
    }
    sat_solver_store_mark_clauses_a( pSat );

    // update the last clause
    if fRelation != 0 {
        extern int sat_solver_store_change_last( sat_solver * pSat );
        iLast = sat_solver_store_change_last( pSat );
    }

    // add clauses of B
    for  i = 0; i < pCnfOff.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnfOff.pClauses[i], pCnfOff.pClauses[i+1] ) {
            Cnf_DataFree( pCnfOn );
            Cnf_DataFree( pCnfOff );
            sat_solver_delete( pSat );
            return nil;
        }
    }

    // add PI clauses
    // collect the common variables
    vVarsAB = Vec_IntAlloc( Aig_ManCiNum(pManOn) );
    if fRelation != 0 {
        Vec_IntPush( vVarsAB, iLast ); }

    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pManOn.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pManOn.vCis, i)), 1); i++  {
        Vec_IntPush( vVarsAB, pCnfOn.pVarNums[pObj.Id] );
        pObj2 = Aig_ManCi( pManOff, i );

        Lits[0] = toLitCond( pCnfOn.pVarNums[pObj.Id], 0 );
        Lits[1] = toLitCond( pCnfOff.pVarNums[pObj2.Id], 1 );
        if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
            assert(false, "0"); }
        Lits[0] = toLitCond( pCnfOn.pVarNums[pObj.Id], 1 );
        Lits[1] = toLitCond( pCnfOff.pVarNums[pObj2.Id], 0 );
        if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
            assert(false, "0"); }
    }
    Cnf_DataFree( pCnfOn );
    Cnf_DataFree( pCnfOff );
    sat_solver_store_mark_roots( pSat );

/*
    status = sat_solver_simplify(pSat);
    if ( status == 0 )
    {
        Vec_IntFree( vVarsAB );
        Cnf_DataFree( pCnfOn );
        Cnf_DataFree( pCnfOff );
        sat_solver_delete( pSat );
        return NULL;
    }
*/

    // solve the problem
    status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
timeSat += Abc_Clock() - clk;
    if status == l_False {
        pSatCnf = sat_solver_store_release( pSat );
//        printf( "unsat\n" );
    
    } else if status == l_True {
//        printf( "sat\n" );
    
    } else {
//        printf( "undef\n" );
    }
    sat_solver_delete( pSat );
    if pSatCnf == nil {
        fmt.Printf("The SAT problem is not unsat.\n");
        Vec_IntFree( vVarsAB );
        return nil;
    }

    // create the resulting manager
clk = Abc_Clock();
    pManInter = Inta_ManAlloc();
    pRes = (Aig_Man_t *)Inta_ManInterpolate( pManInter, (Sto_Man_t *)pSatCnf, 0, vVarsAB, fVerbose );
    Inta_ManFree( pManInter );
timeInt += Abc_Clock() - clk;
/*
    // test UNSAT core computation
    {
        Intp_Man_t * pManProof;
        Vec_Int_t * vCore;
        pManProof = Intp_ManAlloc();
        vCore = Intp_ManUnsatCore( pManProof, pSatCnf, 0 );
        Intp_ManFree( pManProof );
        Vec_IntFree( vCore );
    }
*/
    Vec_IntFree( vVarsAB );
    Sto_ManFree( (Sto_Man_t *)pSatCnf );

//    Ioa_WriteAiger( pRes, "inter2.aig", 0, 0 );
    return pRes;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

