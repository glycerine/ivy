package wlc

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [wlc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Parses several flavors of word-level Verilog.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlc.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


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
func Wlc_BlastFullAdderCtrlCnf(pSat *sat_solver, a int32, ac int32, b int32, c int32, pc *int32, ps *int32, piVars *int32) {
    int Cnf[12][6] = {            // xabc cs   // abc cs

        { -1, 0, 0, 0,  0, 0 },   // -000 00   // 000 00
        { -1, 0, 0, 1,  0, 1 },   // -001 01   // 001 01
        { -1, 0, 1, 0,  0, 1 },   // -010 01   // 010 01
        { -1, 0, 1, 1,  1, 0 },   // -011 10   // 011 10

        {  0,-1, 0, 0,  0, 0 },   // 0-00 00
        {  0,-1, 0, 1,  0, 1 },   // 0-01 01
        {  0,-1, 1, 0,  0, 1 },   // 0-10 01
        {  0,-1, 1, 1,  1, 0 },   // 0-11 10

        {  1, 1, 0, 0,  0, 1 },   // 1100 01   // 100 01
        {  1, 1, 0, 1,  1, 0 },   // 1101 10   // 101 10
        {  1, 1, 1, 0,  1, 0 },   // 1110 10   // 110 10
        {  1, 1, 1, 1,  1, 1 }    // 1111 11   // 111 11
    };

    var pVars [6]int32 = [6]int32{a, ac, b, c, *piVars, *piVars+1};
    var i int32;
    var v int32;
    var nLits int32;
    var pLits [6]int32;
    for  i = 0; i < 12; i++  {
        nLits = 0;
        for  v = 0; v < 6; v++  {
            if Cnf[i][v] == -1 {
                continue; }
            if pVars[v] == 0 { // const 0
            
                if Cnf[i][v] == 0 {
                    continue; }
                if Cnf[i][v] == 1 {
                    break; }
            }
            if pVars[v] == -1 { // const -1
            
                if Cnf[i][v] == 0 {
                    break; }
                if Cnf[i][v] == 1 {
                    continue; }
            }
            pLits[nLits++] = Abc_Var2Lit( pVars[v], Cnf[i][v] );
        }
        if v < 6 {
            continue; }
        assert(nLits > 2, "nLits > 2");
        sat_solver_addclause( pSat, pLits, pLits + nLits );
    }
    *pc = (*piVars)++;
    *ps = (*piVars)++;
}
func Wlc_BlastMultiplierCnf(pSat *sat_solver, pArgA *int32, pArgB *int32, nArgA int32, nArgB int32, vTemp *Vec_Int_t, vRes *Vec_Int_t, piVars *int32) {
    var pRes *int32;
    var pArgC *int32;
    var pArgS *int32;
    var a int32;
    var b int32;
    var Carry int32;
    assert(nArgA > 0 && nArgB > 0, "nArgA > 0 && nArgB > 0");
    // prepare result
    Vec_IntFill( vRes, nArgA + nArgB, 0 );
    pRes = Vec_IntArray( vRes );
    // prepare intermediate storage
    Vec_IntFill( vTemp, 2 * nArgA, 0 );
    pArgC = Vec_IntArray( vTemp );
    pArgS = pArgC + nArgA;
    // create matrix
    for  b = 0; b < nArgB; b++  {
        for  a = 0; a < nArgA; a++  {
            Wlc_BlastFullAdderCtrlCnf( pSat, pArgA[a], pArgB[b], pArgS[a], pArgC[a], &pArgC[a], a ? &pArgS[a-1] : &pRes[b], piVars );
        }
    }
    // final addition
    pArgS[nArgA-1] = 0;
    for  a = 0; a < nArgA; a++  {
        Wlc_BlastFullAdderCtrlCnf( pSat, -1, pArgC[a], pArgS[a], Carry, &Carry, &pRes[nArgB+a], piVars );
    }
}
func Wlc_BlastMultiplierCnfMain(nBits int32) *sat_solver {
    var vRes1 *Vec_Int_t  = Vec_IntAlloc( 2*nBits );
    var vRes2 *Vec_Int_t  = Vec_IntAlloc( 2*nBits );
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 2*nBits );

    var pArgA *int32  = ABC_ALLOC( int, nBits );
    var pArgB *int32  = ABC_ALLOC( int, nBits );
    var i int32;
    var Ent1 int32;
    var Ent2 int32;
    var nVars int32 = 1 + 2*nBits;
    var nVarsAll int32  = 1 + 4*nBits + 4*nBits*(nBits + 1);

    var pSat *sat_solver  = sat_solver_new();
    sat_solver_setnvars( pSat, nVarsAll );

    for  i = 0; i < nBits; i++  {
        pArgA[i] = 1 + i, pArgB[i] = 1 + nBits + i;
    }
    Wlc_BlastMultiplierCnf( pSat, pArgA, pArgB, nBits, nBits, vTemp, vRes1, &nVars );

    for  i = 0; i < nBits; i++  {
        pArgA[i] = 1 + nBits + i, pArgB[i] = 1 + i;
    }
    Wlc_BlastMultiplierCnf( pSat, pArgA, pArgB, nBits, nBits, vTemp, vRes2, &nVars );

    Vec_IntClear( vTemp );
    Vec_IntForEachEntryTwo( vRes1, vRes2, Ent1, Ent2, i )
    {
        Vec_IntPush( vTemp, Abc_Var2Lit(nVars, 0) );
        sat_solver_add_xor( pSat, Ent1, Ent2, nVars++, 0 );
    }
    assert(nVars == nVarsAll, "nVars == nVarsAll");
    sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );

    ABC_FREE( pArgA );
    ABC_FREE( pArgB );
    Vec_IntFree( vRes1 );
    Vec_IntFree( vRes2 );
    Vec_IntFree( vTemp );
    return pSat;
}
func Wlc_BlastMultiplierCnfTest(nBits int32) {
    var clk abctime  = Abc_Clock();
    var pSat *sat_solver  = Wlc_BlastMultiplierCnfMain( nBits );
    var i int32;
    var status int32 = sat_solver_solve( pSat, nil, nil, 0, 0, 0, 0 );
    Sat_SolverWriteDimacs( pSat, "test_mult.cnf", nil, nil, 0 );
    for  i = 0; i < sat_solver_nvars(pSat); i++  {
        fmt.Printf("%d=%d ", i, sat_solver_var_value(pSat, i));
    }
    fmt.Printf("\n");

    fmt.Printf("Verifying for %d bits:  %s   ", nBits, status == l_True ? "SAT" : "UNSAT");
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    sat_solver_delete( pSat );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

