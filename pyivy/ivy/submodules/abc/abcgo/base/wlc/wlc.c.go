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
func Wlc_GenerateCodeMax4(nBits int32) {
    var nWidth, nSteps, i int32;
    var pFile *os.File  = os.OpenFile("max4.v", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        return; }

    for  nSteps, nWidth = 0, 1; nWidth < nBits; nWidth *= 3, nSteps++  {}

    fmt.Fprintf(pFile, "module max4 ( a, b, c, d, res, addr );\n\n");
    fmt.Fprintf(pFile, "  input [%d:0] a, b, c, d;\n", nBits-1);
    fmt.Fprintf(pFile, "  output [%d:0] res;\n", nBits-1);
    fmt.Fprintf(pFile, "  output [1:0] addr;\n\n");

    fmt.Fprintf(pFile, "  wire [%d:0] A = a;\n", nWidth-1);
    fmt.Fprintf(pFile, "  wire [%d:0] B = b;\n", nWidth-1);
    fmt.Fprintf(pFile, "  wire [%d:0] C = c;\n", nWidth-1);
    fmt.Fprintf(pFile, "  wire [%d:0] D = d;\n\n", nWidth-1);

    fmt.Fprintf(pFile, "  wire AB, AC, AD, BC, BD, CD;\n\n");

    fmt.Fprintf(pFile, "  comp( A, B, AB );\n");
    fmt.Fprintf(pFile, "  comp( A, C, AC );\n");
    fmt.Fprintf(pFile, "  comp( A, D, AD );\n");
    fmt.Fprintf(pFile, "  comp( B, C, BC );\n");
    fmt.Fprintf(pFile, "  comp( B, D, BD );\n");
    fmt.Fprintf(pFile, "  comp( C, D, CD );\n\n");

    fmt.Fprintf(pFile, "  assign addr = AB ?  (CD ? (AC ? 2\'b00 : 2\'b10) : (AD ? 2\'b00 : 2\'b11))  :  (CD ? (BC ? 2\'b01 : 2\'b10) : (BD ? 2\'b01 : 2\'b11));\n\n");
    fmt.Fprintf(pFile, "  assign res = addr[1] ? (addr[1] ? d : c) : (addr[0] ? b : a);\n\n");
    fmt.Fprintf(pFile, "endmodule\n\n\n");


    fmt.Fprintf(pFile, "module comp ( a, b, res );\n\n");
    fmt.Fprintf(pFile, "  input [%d:0] a, b;\n", nWidth-1);
    fmt.Fprintf(pFile, "  output res;\n");
    fmt.Fprintf(pFile, "  wire res2;\n\n");

    fmt.Fprintf(pFile, "  wire [%d:0] A =  a & ~b;\n", nWidth-1);
    fmt.Fprintf(pFile, "  wire [%d:0] B = ~a &  b;\n\n", nWidth-1);

    fmt.Fprintf(pFile, "  comp0( A, B, res, res2 );\n\n");

    fmt.Fprintf(pFile, "endmodule\n\n\n");


    for  i = 0; i < nSteps; i++  {
        fmt.Fprintf(pFile, "module comp%d ( a, b, yes, no );\n\n", i);
        fmt.Fprintf(pFile, "  input [%d:0] a, b;\n", nWidth-1);
        fmt.Fprintf(pFile, "  output yes, no;\n\n", nWidth/3-1);

        fmt.Fprintf(pFile, "  wire [2:0] y, n;\n\n");

        if i == nSteps - 1 {
            fmt.Fprintf(pFile, "  assign y = a;\n");
            fmt.Fprintf(pFile, "  assign n = b;\n\n");
        
        } else {
            fmt.Fprintf(pFile, "  wire [%d:0] A0 = a[%d:%d];\n", nWidth/3-1, nWidth/3-1, 0);
            fmt.Fprintf(pFile, "  wire [%d:0] A1 = a[%d:%d];\n", nWidth/3-1, 2*nWidth/3-1, nWidth/3);
            fmt.Fprintf(pFile, "  wire [%d:0] A2 = a[%d:%d];\n\n", nWidth/3-1, nWidth-1, 2*nWidth/3);

            fmt.Fprintf(pFile, "  wire [%d:0] B0 = b[%d:%d];\n", nWidth/3-1, nWidth/3-1, 0);
            fmt.Fprintf(pFile, "  wire [%d:0] B1 = b[%d:%d];\n", nWidth/3-1, 2*nWidth/3-1, nWidth/3);
            fmt.Fprintf(pFile, "  wire [%d:0] B2 = b[%d:%d];\n\n", nWidth/3-1, nWidth-1, 2*nWidth/3);

            fmt.Fprintf(pFile, "  comp%d( A0, B0, y[0], n[0] );\n", i+1);
            fmt.Fprintf(pFile, "  comp%d( A1, B1, y[1], n[1] );\n", i+1);
            fmt.Fprintf(pFile, "  comp%d( A2, B2, y[2], n[2] );\n\n", i+1);
        }

        fmt.Fprintf(pFile, "  assign yes = y[0] | (~y[0] & ~n[0] & y[1]) | (~y[0] & ~n[0] & ~y[1] & ~n[1] & y[2]);\n");
        fmt.Fprintf(pFile, "  assign no  = n[0] | (~y[0] & ~n[0] & n[1]) | (~y[0] & ~n[0] & ~y[1] & ~n[1] & n[2]);\n\n");

        fmt.Fprintf(pFile, "endmodule\n\n\n");

        nWidth /= 3;
    }
    _ = (pFile).Close();
}


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

