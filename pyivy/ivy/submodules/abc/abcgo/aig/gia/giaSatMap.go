package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaSatMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatMap.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "sat/bsat/satStore.h"
//#include "misc/vec/vecWec.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


// operation manager
   
/**Function*************************************************************

  Synopsis    []

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
func sat_solver_add_and1(pSat *sat_solver, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fCompl int32) int32 {
    var Lits [3]lit;
    var Cid int32;

    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar0, fCompl0 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar1, fCompl1 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");
/*
    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar0, !fCompl0 );
    Lits[2] = toLitCond( iVar1, !fCompl1 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 3 );
    assert( Cid );
*/
    return 3;
}

func sat_solver_add_and2(pSat *sat_solver, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fCompl int32) int32 {
    var Lits [3]lit;
    var Cid int32;
/*
    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar0, fCompl0 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 2 );
    assert( Cid );

    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar1, fCompl1 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 2 );
    assert( Cid );
*/
    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar0, !fCompl0 );
    Lits[2] = toLitCond( iVar1, !fCompl1 );
    Cid = sat_solver_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");
    return 3;
}

/**Function*************************************************************

  Synopsis    [Adds a general cardinality constraint in terms of vVars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbm_AddCardinConstrTest() {
    var LogN int32 = 3;
    var nVars int32 = 1 << LogN;
    var K int32 = 2;
    var Count int32 = 1;
    var vVars *Vec_Int_t;
    var vLits *Vec_Int_t = Vec_IntAlloc( nVars );
    var pSat *sat_solver  = Sbm_AddCardinSolver( LogN, &vVars );
    var nVarsReal int32  = sat_solver_nvars( pSat );

    var Lit int32  = Abc_Var2Lit( Vec_IntEntry(vVars, K), 1 );
    fmt.Printf("LogN = %d. N = %3d.   Vars = %5d. Clauses = %6d.  Comb = %d.\n", LogN, nVars, nVarsReal, sat_solver_nclauses(pSat), nVars * (nVars-1)/2 + nVars + 1);
    for {
        var i int32;
        var status int32 = sat_solver_solve( pSat, &Lit, &Lit+1, 0, 0, 0, 0 );
        if status != l_True {
            break; }
        Vec_IntClear( vLits );
        fmt.Printf("%3d : ", Count++);
        for  i = 0; i < nVars; i++  {
            Vec_IntPush( vLits, Abc_Var2Lit(i, sat_solver_var_value(pSat, i)) );
            fmt.Printf("%d", sat_solver_var_value(pSat, i));
        }
        fmt.Printf("\n");
        status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + nVars );
        if status == 0 {
            break; }
    }

    sat_solver_delete( pSat );
    Vec_IntFree( vVars );
    Vec_IntFree( vLits );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbm_ManTestSat(pMan any) int32 {
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var i int32;
    var k int32;
    var Lit int32;
    var LogN int32 = 7;
    var nVars int32 = 1 << LogN;
    var status int32;
    var Root int32;
    var p *Sbm_Man_t  = Sbm_ManAlloc( LogN );
    var InvArea word;
    var fKeepTrying int32  = 1;
    var StartSol int32;
    // derive window
    extern int Nf_ManExtractWindow( void * pMan, Vec_Int_t * vRoots, Vec_Wec_t * vCuts, Vec_Wec_t * vObjCuts, Vec_Int_t * vSolCuts, Vec_Int_t * vCutGates, Vec_Wrd_t * vCutAreas, word * pInvArea, int StartVar, int nVars );
    p.nInputs = Nf_ManExtractWindow( pMan, p.vRoots, p.vCuts, p.vObjCuts, p.vSolCuts, p.vCutGates, p.vCutAreas, &InvArea, p.FirstVar, nVars );
    p.LitShift = 2*p.nInputs+2;
    assert(Vec_WecSize(p.vObjCuts) + p.nInputs <= nVars, "Vec_WecSize(p.vObjCuts) + p.nInputs <= nVars");

    // print-out
//    Vec_WecPrint( p->vCuts, 0 );
//    printf( "\n" );
//    Vec_WecPrint( p->vObjCuts, 0 );
//    printf( "\n" );
    Vec_IntPrint( p.vSolCuts );
    fmt.Printf("All clauses = %d.  Multi clauses = %d.  Binary clauses = %d.  Other clauses = %d.\n", sat_solver_nclauses(p.pSat), Vec_WecSize(p.vObjCuts), Vec_WecSizeSize(p.vCuts), sat_solver_nclauses(p.pSat) - Vec_WecSize(p.vObjCuts) - Vec_WecSizeSize(p.vCuts));

    // creating CNF
    if !(Sbm_ManCreateCnf(p) != 0) {
        return 0; }

    // create assumptions
    // cardinality
    Vec_IntClear( p.vAssump );
    Vec_IntPush( p.vAssump, -1 );
    // unused variables
    for  i = Vec_WecSize(p.vObjCuts) + p.nInputs; i < nVars; i++  {
        Vec_IntPush( p.vAssump, Abc_Var2Lit(i, 1) );
    }
    // root variables
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vRoots)) && (((Root) = Vec_IntEntry(p.vRoots, i)), 1); i++  {
        Vec_IntPush( p.vAssump, Abc_Var2Lit(Root, 0) );
    }
//    Vec_IntPrint( p->vAssump );

    StartSol = Vec_IntSize(p.vSolCuts);
//    StartSol = 30;
    for  (fKeepTrying != 0) && StartSol-fKeepTrying > 0  {
        fmt.Printf("Trying to find mapping with %d gates.\n", StartSol-fKeepTrying);
    //    for ( i = Vec_IntSize(p->vSolCuts)-5; i < nVars; i++ )
    //        Vec_IntPush( p->vAssump, Abc_Var2Lit(Vec_IntEntry(p->vCardVars, i), 1) );
        Vec_IntWriteEntry( p.vAssump, 0, Abc_Var2Lit(Vec_IntEntry(p.vCardVars, StartSol-fKeepTrying), 1) );
        // solve the problem
        clk2 = Abc_Clock();
        status = sat_solver_solve( p.pSat, Vec_IntArray(p.vAssump), Vec_IntLimit(p.vAssump), 0, 0, 0, 0 );
        if status == l_True {
            var AreaNew word;
            var Count int32;
            fmt.Printf("AND Lits = %d.  Inputs = %d.  Vars = %d.  All vars = %d.\n", Vec_WecSize(p.vObjCuts), p.nInputs, Vec_WecSize(p.vObjCuts) + p.nInputs, nVars);
//            for ( i = 0; i < nVars; i++ )
//                printf( "%d", sat_solver_var_value(p->pSat, i) );
//            printf( "\n" );
            for  i = 0; i < nVars; i++  {
                if sat_solver_var_value(p.pSat, i) {
                    fmt.Printf("%d=%d ", i, sat_solver_var_value(p.pSat, i));
                    Count++;
                    if i >= Vec_WecSize(p.vObjCuts) {
                        AreaNew += InvArea; }
                }
            }
            fmt.Printf("Count = %d\n", Count);
//            for ( i = p->FirstVar; i < sat_solver_nvars(p->pSat); i++ )
//                printf( "%d", sat_solver_var_value(p->pSat, i) );
//            printf( "\n" );
            Count = 1;
            for  i = p.FirstVar; i < sat_solver_nvars(p.pSat); i++  {
                if sat_solver_var_value(p.pSat, i) {
                    var vCutOne *Vec_Int_t  = Vec_WecEntry(p.vCuts, i-p.FirstVar);
                    fmt.Printf("%2d : Cut %3d  (Gate %2d)  ", Count, i-p.FirstVar, Vec_IntEntry(p.vCutGates, i-p.FirstVar));
                    // inlining of macro Vec_IntForEachEntry
                    for  k = 0; (k < Vec_IntSize(vCutOne)) && (((Lit) = Vec_IntEntry(vCutOne, k)), 1); k++  {
                        fmt.Printf("%d(%d) ", Lit - 2*(p.nInputs+1), Abc_Lit2Var(Lit));
                    }
                    fmt.Printf("\n");
                    Count++;
                    AreaNew += Vec_WrdEntry(p.vCutAreas, i-p.FirstVar);
                }
            }
            fmt.Printf("Area = %7.2f\n", Scl_Int2Flt(int32(AreaNew)));
        }
        if status == l_False {
            fmt.Printf("UNSAT ");
fKeepTrying = 0;
        } else if status == l_True {
            fmt.Printf("SAT   ");
fKeepTrying++; }
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk2 );
        Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
        fmt.Printf("\n");
    }
    Sbm_ManStop( p );
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

