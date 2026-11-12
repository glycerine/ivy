package sbd

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [sbd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const MAX_M =  8 // max inputs
const MAX_N = 30 // max nodes
const MAX_K =  6 // max lutsize
const MAX_D =  8 // max delays

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// new AIG manager
   
type Sbd_Pro_t struct {
    nLuts             int32;  // LUT count
    nSize             int32;  // LUT size
    nDivs             int32;  // divisor count
    nVars             int32;  // intermediate variables (nLuts * nSize)
    nPars             int32;  // total parameter count (nLuts * (1 << nSize) + nLuts * nSize * nDivs) 
    pPars1 [SBD_LUTS_MAX][1<<SBD_SIZE_MAX]int32;            // lut parameters
    pPars2 [SBD_LUTS_MAX][SBD_SIZE_MAX][SBD_DIV_MAX]int32;  // topo parameters
    pVars [SBD_LUTS_MAX][SBD_SIZE_MAX+1]int32;              // internal variables
    pDivs [SBD_DIV_MAX]int32;                               // divisor variables (external)
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_ProblemSetup(p *Sbd_Pro_t, nLuts int32, nSize int32, nDivs int32) *Vec_Int_t {
    var vCnf *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var i, k, d, v, n, iVar int32;
    assert(nLuts >= 1 && nLuts <= 2, "nLuts >= 1 && nLuts <= 2");
    *p = Sbd_Pro_t{};
    p.nLuts = nLuts;
    p.nSize = nSize;
    p.nDivs = nDivs;
    p.nVars = nLuts * nSize;
    p.nPars = nLuts * (1 << nSize) + nLuts * nSize * nDivs;
    // set parameters
    for  i = 0; i < nLuts; i++  {
    for  k = 0; k < (1 << nSize); k++  {
        p.pPars1[i][k] = iVar++;
    }
    }
    for  i = 0; i < nLuts; i++  {
    for  k = 0; k < nSize; k++  {
    for  d = 0; d < nDivs; d++  {
        p.pPars2[i][k][d] = iVar++;
    }
    }
    }
    // set intermediate variables
    for  i = 0; i < nLuts; i++  {
    for  k = 0; k < nSize; k++  {
        p.pVars[i][k] = iVar++;
    }
    }
    // set top variables 
    for  i = 1; i < nLuts; i++  {
        p.pVars[i][nSize] = p.pVars[i-1][0];
    }
    // set divisor variables
    for  d = 0; d < nDivs; d++  {
        p.pDivs[d] = iVar++;
    }
    assert(iVar == p.nPars + p.nVars + p.nDivs, "iVar == p.nPars + p.nVars + p.nDivs");

    // input compatiblity clauses
    for  i = 0; i < nLuts; i++  {
    for  k = (i > 0); k < nSize; k++  {
    for  d = 0; d < nDivs; d++  {
    for  n = 0; n < nDivs; n++  {
        if n < d {
            Vec_IntPush( vCnf, Abc_Var2Lit(p.pPars2[i][k][d], 0) );
            Vec_IntPush( vCnf, Abc_Var2Lit(p.pPars2[i][k][n], 0) );
            Vec_IntPush( vCnf, -1 );
        
        } else if k < nSize-1 {
            Vec_IntPush( vCnf, Abc_Var2Lit(p.pPars2[i][k][d], 0) );
            Vec_IntPush( vCnf, Abc_Var2Lit(p.pPars2[i][k+1][n], 0) );
            Vec_IntPush( vCnf, -1 );
        }
    }
    }
    }
    }

    // create LUT clauses
    for  i = 0; i < nLuts; i++  {
    for  k = 0; k < (1 << nSize); k++  {
    for  n = 0; n < 2; n++  {
        for  v = 0; v < nSize; v++  {
            Vec_IntPush( vCnf, Abc_Var2Lit(p.pPars1[i][v], (k >> v) & 1) );
        }
        Vec_IntPush( vCnf, Abc_Var2Lit(p.pVars[i][nSize], n) );
        Vec_IntPush( vCnf, Abc_Var2Lit(p.pPars1[i][k], !n) );
        Vec_IntPush( vCnf, -1 );
    }
    }
    }

    // create input clauses
    for  i = 0; i < nLuts; i++  {
    for  k = (i > 0); k < nSize; k++  {
    for  d = 0; d < nDivs; d++  {
    for  n = 0; n < 2; n++  {
        Vec_IntPush( vCnf, Abc_Var2Lit(p.pPars2[i][k][d], 0) );
        Vec_IntPush( vCnf, Abc_Var2Lit(p.pPars1[i][k], n) );
        Vec_IntPush( vCnf, Abc_Var2Lit(p.pDivs[d], !n) );
        Vec_IntPush( vCnf, -1 );
    }
    }
    }
    }

    return vCnf;
}
// add clauses to the don't-care computation solver
func Sbd_ProblemLoad1(p *Sbd_Pro_t, vCnf *Vec_Int_t, iStartVar int32, pDivVars *int32, iTopVar int32, pSat *sat_solver) {
    var pLits [8]int32;
    var nLits int32;
    var i int32;
    var k int32;
    var iLit int32;
    var RetValue int32;
    var ThisTopVar int32  = p.pVars[0][p.nSize];
    var FirstDivVar int32  = p.nPars + p.nVars;
    // add clauses
    for  i = 0; i < Vec_IntSize(vCnf); i++  {
        assert(Vec_IntEntry(vCnf, i) != -1, "Vec_IntEntry(vCnf, i) != -1");
        for  k = i+1; k < Vec_IntSize(vCnf); k++  {
            if Vec_IntEntry(vCnf, i) == -1 {
                break; }
        }
        nLits = 0;
        Vec_IntForEachEntryStartStop( vCnf, iLit, i, i, k ) {
            if Abc_Lit2Var(iLit) == ThisTopVar {
                pLits[nLits++] = Abc_Var2Lit( ThisTopVar, Abc_LitIsCompl(iLit) );
            } else if Abc_Lit2Var(iLit) >= FirstDivVar {
                pLits[nLits++] = Abc_Var2Lit( pDivVars[Abc_Lit2Var(iLit)-FirstDivVar], Abc_LitIsCompl(iLit) );
            } else {
                pLits[nLits++] = iLit + 2 * iStartVar; }
        }
        assert(nLits <= 8, "nLits <= 8");
        RetValue = sat_solver_addclause( pSat, pLits, pLits + nLits );
        assert((RetValue != 0), "RetValue");
    }
}
// add clauses to the functionality evaluation solver
func Sbd_ProblemLoad2(p *Sbd_Pro_t, vCnf *Vec_Wec_t, iStartVar int32, pDivVarValues *int32, iTopVarValue int32, pSat *sat_solver) {
    var vLevel *Vec_Int_t;
    var pLits [8]int32;
    var nLits int32;
    var i int32;
    var k int32;
    var iLit int32;
    var RetValue int32;
    var ThisTopVar int32  = p.pVars[0][p.nSize];
    var FirstDivVar int32  = p.nPars + p.nVars;
    var FirstIntVar int32  = p.nPars;
    // add clauses
    Vec_WecForEachLevel( vCnf, vLevel, i )
    {
        nLits = 0;
        Vec_IntForEachEntry( vLevel, iLit, k ) {
            if Abc_Lit2Var(iLit) == ThisTopVar {
                if Abc_LitIsCompl(iLit) == iTopVarValue {
                    break; }
                continue;
            
            } else if Abc_Lit2Var(iLit) >= FirstDivVar {
                if Abc_LitIsCompl(iLit) == pDivVarValues[Abc_Lit2Var(iLit)-FirstDivVar] {
                    break; }
                continue;
            
            } else if Abc_Lit2Var(iLit) >= FirstIntVar {
                pLits[nLits++] = iLit + 2 * iStartVar;
            } else {
                pLits[nLits++] = iLit; }
        }
        if k < Vec_IntSize(vLevel) {
            continue; }
        assert(nLits <= 8, "nLits <= 8");
        RetValue = sat_solver_addclause( pSat, pLits, pLits + nLits );
        assert((RetValue != 0), "RetValue");
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
sat_solver * Sbd_SolverTopo( int M, int N, int K, int pVars[MAX_N][MAX_M+MAX_N][MAX_K], int pVars2[MAX_M+MAX_N][MAX_D], int pDelays[], int Req, int * pnVars ) // inputs, nodes, lutsize
{
    var pSat *sat_solver  = nil;
    var vTemp *Vec_Int_t  = Vec_IntAlloc(100);
    // assign vars
    var RetValue, n, i, j, j2, k, k2, d, Count, nVars int32;
    for  n = 0; n < N;   n++  {
    for  i = 0; i < M+N; i++  {
    for  k = 0; k < K;   k++  {
        pVars[n][i][k] = -1;
    }
    }
    }
    for  n = 0; n < N;   n++  {
    for  i = 0; i < M+n; i++  {
    for  k = 0; k < K;   k++  {
        pVars[n][i][k] = nVars++;
    }
    }
    }
    fmt.Printf("Number of topo vars = %d.\n", nVars);
    *pnVars = nVars;
    // add constraints
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, nVars );
    // each node is used
    for  i = 0; i < M+N-1; i++  {
        Vec_IntClear( vTemp );
        for  n = 0; n < N; n++  {
        for  k = 0; k < K; k++  {
            if pVars[n][i][k] >= 0 {
                Vec_IntPush( vTemp, Abc_Var2Lit(pVars[n][i][k], 0) ); }
        }
        }
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
        assert((RetValue != 0), "RetValue");
    }
    fmt.Printf("Added %d node connectivity constraints.\n", i);
    // each fanin of each node is connected exactly once
    Count = 0;
    for  n = 0; n < N; n++  {
    for  k = 0; k < K; k++  {
        // connected
        Vec_IntClear( vTemp );
        for  i = 0; i < M+n; i++  {
            Vec_IntPush( vTemp, Abc_Var2Lit(pVars[n][i][k], 0) );
        }
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
        assert((RetValue != 0), "RetValue");
        // exactly once
        for  i = 0;   i < M+n; i++  {
        for  j = i+1; j < M+n; j++  {
            Vec_IntFillTwo( vTemp, 2, Abc_Var2Lit(pVars[n][i][k], 1), Abc_Var2Lit(pVars[n][j][k], 1) );
            RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
            assert((RetValue != 0), "RetValue");
            Count++;
        }
        }
    }
    }
    fmt.Printf("Added %d fanin connectivity constraints.\n", Count);
    // node fanins are unique
    Count = 0;
    for  n = 0; n < N;   n++  {
    for  i = 0; i < M+n; i++  {
    for  k = 0; k < K;   k++  {
    for  j = i; j < M+n; j++  {
    for  k2 = k+1; k2 < K; k2++  {
        Vec_IntFillTwo( vTemp, 2, Abc_Var2Lit(pVars[n][i][k], 1), Abc_Var2Lit(pVars[n][j][k2], 1) );
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
        assert((RetValue != 0), "RetValue");
        Count++;
    }
    }
    }
    }
    }
    fmt.Printf("Added %d fanin exclusivity constraints.\n", Count);
    // nodes are ordered
    Count = 0;
    for  n = 1; n < N;     n++  {
    for  i = 0; i < M+n-1; i++  {
        // first of n cannot be smaller than first of n-1 (but can be equal)
        for  j = i+1; j < M+n-1; j++  {
            Vec_IntFillTwo( vTemp, 2, Abc_Var2Lit(pVars[n][i][0], 1), Abc_Var2Lit(pVars[n-1][j][0], 1) );
            RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
            assert((RetValue != 0), "RetValue");
            Count++;
        }
        // if first nodes of n and n-1 are equal, second nodes are ordered
        Vec_IntFillTwo( vTemp, 2, Abc_Var2Lit(pVars[n][i][0], 1), Abc_Var2Lit(pVars[n-1][i][0], 1) );
        for  j = 0;    j < i;  j++  {
        for  j2 = j+1; j2 < i; j2++  {
            Vec_IntPushTwo( vTemp, Abc_Var2Lit(pVars[n][j][1], 1), Abc_Var2Lit(pVars[n-1][j2][1], 1) );
            RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
            assert((RetValue != 0), "RetValue");
            Vec_IntShrink( vTemp, 2 );
            Count++;
        }
        }
    }
    }
    fmt.Printf("Added %d node ordering constraints.\n", Count);
    // exclude fanins of two-input nodes
    Count = 0;
    if K == 2 {
    for  n = 1; n < N;   n++  {
    for  i = M; i < M+n; i++  {
    for  j = 0; j < i;   j++  {
    for  k = 0; k < K;   k++  {
        Vec_IntClear( vTemp );
        Vec_IntPush( vTemp, Abc_Var2Lit(pVars[n][i][0], 1) );
        Vec_IntPush( vTemp, Abc_Var2Lit(pVars[n][j][1], 1) );
        Vec_IntPush( vTemp, Abc_Var2Lit(pVars[i-M][j][k], 1) );
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
        assert((RetValue != 0), "RetValue");
        Count++;
    }
    }
    }
    } }
    fmt.Printf("Added %d two-node non-triviality constraints.\n", Count);


    // assign delay vars
    assert(Req < MAX_D-1, "Req < MAX_D-1");
    for  i = 0; i < M+N;   i++  {
    for  d = 0; d < MAX_D; d++  {
        pVars2[i][d] = nVars++;
    }
    }
    fmt.Printf("Number of total vars = %d.\n", nVars);
    // set input delays
    for  i = 0; i < M; i++  {
        assert(pDelays[i] < MAX_D-2, "pDelays[i] < MAX_D-2");
        Vec_IntFill( vTemp, 1, Abc_Var2Lit(pVars2[i][pDelays[i]], 0) );
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
        assert((RetValue != 0), "RetValue");
    }
    // set output delay
    for  k = Req; k < MAX_D; k++  {
        Vec_IntFill( vTemp, 1, Abc_Var2Lit(pVars2[M+N-1][Req+1], 1) );
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
        assert((RetValue != 0), "RetValue");
    }
    // set internal nodes
    for  n = 0; n < N;   n++  {
    for  i = 0; i < M+n; i++  {
    for  k = 0; k < K;   k++  {
    for  d = 0; d < MAX_D-1; d++  {
        Vec_IntFill( vTemp, 1, Abc_Var2Lit(pVars[n][i][k],   1) );
        Vec_IntPush( vTemp,    Abc_Var2Lit(pVars2[i][d],     1) );
        Vec_IntPush( vTemp,    Abc_Var2Lit(pVars2[M+n][d+1], 0) );
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vTemp), Vec_IntLimit(vTemp) );
        assert((RetValue != 0), "RetValue");
    }
    }
    }
    }

    
    Vec_IntFree( vTemp );
    return pSat;
}
func Sbd_SolverTopoPrint(pSat *sat_solver, M int32, N int32, K int32, pVars [][MAX_M+MAX_N][MAX_K]int32) {
    var n, i, k int32;
    fmt.Printf("Solution:\n");
    fmt.Printf("     | ");
    for  n = 0; n < N; n++  {
        fmt.Printf("%2d  ", M+n);
    }
    fmt.Printf("\n");
    for  i = M+N-2; i >= 0; i--  {
        fmt.Printf("%2d %c | ", i, i < M ? 'i' : ' ');
        for  n = 0; n < N; n++  {
            for  k = K-1; k >= 0; k--  {
                if pVars[n][i][k] == -1 {
                    fmt.Printf(" ");
                } else {                    
                    fmt.Printf("%c", sat_solver_var_value(pSat, pVars[n][i][k]) ? '*' : '.'); }
            }
            fmt.Printf("  ");
        }
        fmt.Printf("\n");
    }
}
func Sbd_SolverTopoTest() {
    var M int32  = 8;  //  6;  // inputs
    var N int32  = 3;  // 16;  // nodes
    var K int32  = 4;  //  2;  // lutsize
    var status, v, nVars, nIter, nSols int32;
    var pVars [MAX_N][MAX_M+MAX_N][MAX_K]int32; // 20 x 32 x 6 = 3840
    var pVars2 [MAX_M+MAX_N][MAX_D]int32;       // 20 x 32 x 6 = 3840
    var pDelays [MAX_M]int32 = [MAX_M]int32{1,0,0,0,1};
    var clk abctime  = Abc_Clock();
    var vLits *Vec_Int_t  = Vec_IntAlloc(100);
    var pSat *sat_solver  = Sbd_SolverTopo( M, N, K, pVars, pVars2, pDelays, 2, &nVars );
    for  nIter = 0; nIter < 1000000; nIter++  {
        // find onset minterm
        status = sat_solver_solve( pSat, nil, nil, 0, 0, 0, 0 );
        if status == l_Undef {
            break; }
        if status == l_False {
            break; }
        assert(status == l_True, "status == l_True");
        nSols++;
        // print solution
        if nIter < 5 {
            Sbd_SolverTopoPrint( pSat, M, N, K, pVars ); }
        // remember variable values
        Vec_IntClear( vLits );
        for  v = 0; v < nVars; v++  {
            if sat_solver_var_value(pSat, v) {
                Vec_IntPush( vLits, Abc_Var2Lit(v, 1) ); }
        }
        // add breaking clause
        status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits) );
        if status == 0 {
            break; }
        //if ( nIter == 5 )
        //    break;
    }
    sat_solver_delete( pSat );
    Vec_IntFree( vLits );
    fmt.Printf("Found %d solutions. ", nSols);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}



/**Function*************************************************************

  Synopsis    [Synthesize random topology.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_SolverSynth(M int32, N int32, K int32, pLuts [][MAX_K]int32) {
    var Used [MAX_M+MAX_N]int32 ;
    var nUnused int32  = M;
    var n, iFan0, iFan1 int32;
    srand( time.Now().Unix() );
    for  n = 0; nUnused < N - n; n++  {
        iFan0 = iFan1 = 0;
        for  (iFan0 = rand() % (M + n)) == (iFan1 = rand() % (M + n))  {
            ;
        }
        pLuts[n][0] = iFan0;
        pLuts[n][1] = iFan1;
        if Used[iFan0] == 0 {
            Used[iFan0] = 1;
            nUnused--;
        }
        if Used[iFan1] == 0 {
            Used[iFan1] = 1;
            nUnused--;
        }
        nUnused++;
    }
    if nUnused == N - n {
        // undo the first one
        for  iFan0 = 0; iFan0 < M+n; iFan0++  {
            if Used[iFan0] {
                Used[iFan0] = 0;
                nUnused++;
                break;
            }
        }

    }
    assert(nUnused == N - n + 1, "nUnused == N - n + 1");
    for  ; n < N; n++  {
        for  iFan0 = 0; iFan0 < M+n; iFan0++  {
            if Used[iFan0] == 0 {
                Used[iFan0] = 1;
                break;
            }
        }
        assert(iFan0 < M+n, "iFan0 < M+n");
        for  iFan1 = 0; iFan1 < M+n; iFan1++  {
            if Used[iFan1] == 0 {
                Used[iFan1] = 1;
                break;
            }
        }
        assert(iFan1 < M+n, "iFan1 < M+n");
        pLuts[n][0] = iFan0;
        pLuts[n][1] = iFan1;
    }

    fmt.Printf("{\n");
    for  n = 0; n < N; n++  {
        fmt.Printf("    {%d, %d}%s // %d\n", pLuts[n][0], pLuts[n][1], n==N-1 ? "" :",", M+n);
    }
    fmt.Printf("};\n");
}


/**Function*************************************************************

  Synopsis    [Compute truth table for the given parameter settings.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_SolverTruth(M int32, N int32, K int32, pLuts [][MAX_K]int32, pValues []int32) word {
    var i int32;
    var k int32;
    var v int32;
    var nLutPars int32 = (1 << K) - 1;
    var Truths [MAX_M+MAX_N]word;
    assert(M <= 6 && N <= MAX_N, "M <= 6 && N <= MAX_N");
    for  i = 0; i < M; i++  {
        Truths[i] = s_Truths6[i];
    }
    for  i = 0; i < N; i++  {
        var Truth, Mint word;
        for  k = 1; k <= nLutPars; k++  {            
            if !pValues[i*nLutPars+k-1] {
                continue; }
            Mint = ~(word)0;
            for  v = 0; v < K; v++  {
                Mint &= ((k >> v) & 1) ? Truths[pLuts[i][v]] :  ~Truths[pLuts[i][v]];
            }
            Truth |= Mint;
        }
        Truths[M+i] = Truth;
    }
    return Truths[M+N-1];
}
func Sbd_SolverTruthWord(M int32, N int32, K int32, pLuts [][MAX_K]int32, pValues []int32, pTruthsElem *word, fCompl int32) *word {
    var i int32;
    var k int32;
    var v int32;
    var nLutPars int32 = (1 << K) - 1;
    var nWords int32  = Abc_TtWordNum( M );
    var pRes *word  = pTruthsElem + (M+N-1)*nWords;
    assert(M <= MAX_M && N <= MAX_N, "M <= MAX_M && N <= MAX_N");
    for  i = 0; i < N; i++  {
        var pMint *word;
        var pTruth *word = pTruthsElem + (M+i)*nWords;
        Abc_TtClear( pTruth, nWords );
        for  k = 1; k <= nLutPars; k++  {            
            if !pValues[i*nLutPars+k-1] {
                continue; }
            pMint = pTruthsElem + (M+N)*nWords;
            Abc_TtFill( pMint, nWords );
            for  v = 0; v < K; v++  {
                var pFanin *word  = pTruthsElem + pLuts[i][v]*nWords;
                Abc_TtAndSharp( pMint, pMint, pFanin, nWords, ((k >> v) & 1) == 0 );
            }
            Abc_TtOr( pTruth, pTruth, pMint, nWords );
        }
    }
    if fCompl != 0 {
        Abc_TtNot( pRes, nWords ); }
    return pRes;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_SolverFunc(M int32, N int32, K int32, pLuts [][MAX_K]int32, pTruthInit *word, pValues *int32) int32 {
    var fVerbose int32;
    var clk abctime  = Abc_Clock();
    var clk2, clkOther abctime;
    var pSat *sat_solver  = nil;
    var nWords int32  = Abc_TtWordNum(M);
    var pLits [MAX_K+2]int32;
    var pLits2 [MAX_K+2]int32;
    var nLits int32;
    var nLutPars int32 = (1 << K) - 1;
    var nVars int32 = N * nLutPars;
    var i int32;
    var k int32;
    var m int32;
    var status int32;
    var iMint int32;
    var Iter int32;
    var fCompl int32 = int32((pTruthInit[0] & 1));
    // create truth tables
    var pTruthNew *word;
    var pTruths *word = ABC_ALLOC( word, Abc_TtWordNum(MAX_N) * (MAX_M + MAX_N + 1) );
    Abc_TtElemInit2( pTruths, M );
    // create solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, nVars );
    fmt.Printf("Number of parameters %d x %d = %d.\n", N, nLutPars, nVars);
    // start with the last minterm
//    iMint = (1 << M) - 1;
    iMint = 1;
    for  Iter = 0; Iter < (1 << M); Iter++  {
        // assign the first intermediate variable
        var nVarStart int32  = sat_solver_nvars(pSat);
        sat_solver_setnvars( pSat, nVarStart + N - 1 );
        // add clauses for nodes
        //if ( fVerbose )
            fmt.Printf("Iter %3d : Mint = %3d. Conflicts =%8d.\n", Iter, iMint, sat_solver_nconflicts(pSat));
        for  i = 0; i < N; i++  {
        for  m = 0; m <= nLutPars; m++  {
            if fVerbose != 0 {
                fmt.Printf("i = %d.  m = %d.\n", i, m); }
            // selector variables
            nLits = 0;
            for  k = 0; k < K; k++  {
                if pLuts[i][k] >= M {
                    assert(pLuts[i][k] - M < N - 1, "pLuts[i][k] - M < N - 1");
                    pLits[nLits] = pLits2[nLits] = Abc_Var2Lit( nVarStart + pLuts[i][k] - M, (m >> k) & 1 ); 
                    nLits++;
                
                } else if ((iMint >> pLuts[i][k]) & 1) != ((m >> k) & 1) {
                    break; }
            }
            if k < K {
                continue; }
            // add parameter
            if m != 0 {
                pLits[nLits]  = Abc_Var2Lit( i*nLutPars + m-1, 1 );
                pLits2[nLits] = Abc_Var2Lit( i*nLutPars + m-1, 0 );
                nLits++;
            }
            // node variable
            if i != N - 1 {
                pLits[nLits]  = Abc_Var2Lit( nVarStart + i, 0 );
                pLits2[nLits] = Abc_Var2Lit( nVarStart + i, 1 );
                nLits++;
            }
            // add clauses
            if i != N - 1 || Abc_TtGetBit(pTruthInit, iMint) != fCompl {
                status = sat_solver_addclause( pSat, pLits2, pLits2 + nLits );
                if status == 0 {
                    fCompl = -1;
                    goto finish;
                }
            }
            if (i != N - 1 || Abc_TtGetBit(pTruthInit, iMint) == fCompl) && m > 0 {
                status = sat_solver_addclause( pSat, pLits, pLits + nLits );
                if status == 0 {
                    fCompl = -1;
                    goto finish;
                }
            }
        }
        }

        // run SAT solver
        status = sat_solver_solve( pSat, nil, nil, 0, 0, 0, 0 );
        if status == l_Undef {
            break; }
        if status == l_False {
            fCompl = -1;
            goto finish;
        }
        assert(status == l_True, "status == l_True");

        // collect values
        for  i = 0; i < nVars; i++  {
            pValues[i] = sat_solver_var_value(pSat, i);
        }

        clk2 = Abc_Clock();
        pTruthNew = Sbd_SolverTruthWord( M, N, K, pLuts, pValues, pTruths, fCompl );
        clkOther += Abc_Clock() - clk2;

        if fVerbose != 0 {
            for  i = 0; i < nVars; i++  {
                fmt.Printf("%d=%d ", i, pValues[i]);
            }
            fmt.Printf("  ");
            for  i = nVars; i < sat_solver_nvars(pSat); i++  {
                fmt.Printf("%d=%d ", i, sat_solver_var_value(pSat, i));
            }
            fmt.Printf("\n");
            Extra_PrintBinary( os.Stdout, (unsigned *)pTruthInit, (1 << M) );  fmt.Printf("\n");
            Extra_PrintBinary( os.Stdout, (unsigned *)pTruthNew,  (1 << M) );  fmt.Printf("\n");
        }
        if Abc_TtEqual(pTruthInit, pTruthNew, nWords) {
            break; }

        // get new minterm
        iMint = Abc_TtFindFirstDiffBit( pTruthInit, pTruthNew, M );
    }
finish:
    fmt.Printf("Finished after %d iterations and %d conflicts.  ", Iter, sat_solver_nconflicts(pSat));
    sat_solver_delete( pSat );
    ABC_FREE( pTruths );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Abc_PrintTime( 1, "Time", clkOther );
    return fCompl;
}
func Sbd_SolverFuncTest() {
//    int M = 4;  //  6;  // inputs
//    int N = 3;  // 16;  // nodes
//    int K = 2;  //  2;  // lutsize
//    word Truth = ~((word)3 << 8);
//    int pLuts[MAX_N][MAX_K] = { {0,1}, {2,3}, {4,5}, {6,7}, {8,9} };

/*
    int M =  6;  //  6;  // inputs
    int N = 19;  // 16;  // nodes
    int K =  2;  //  2;  // lutsize
    word pTruth[4] = { ABC_CONST(0x9ef7a8d9c7193a0f), 0, 0, 0 };
    int pLuts[MAX_N][MAX_K] = { 
        {3, 5}, {1, 6}, {0, 5}, {8, 2}, {7, 9},
        {0, 1}, {2, 11}, {5, 12}, {3, 13}, {1, 14},
        {10, 15}, {11, 2}, {3, 17}, {9, 18}, {0, 13},
        {20, 7}, {19, 21}, {4, 16}, {23, 22} 
    };
*/

/*
    int M = 6;  //  6;  // inputs
    int N = 5;  // 16;  // nodes
    int K = 4;  //  2;  // lutsize
    word Truth = ABC_CONST(0x9ef7a8d9c7193a0f);
    int pLuts[MAX_N][MAX_K] = { 
        {0, 1, 2, 3}, // 6
        {1, 2, 3, 4}, // 7
        {2, 3, 4, 5}, // 8
        {0, 1, 4, 5}, // 9
        {6, 7, 8, 9}  // 10
    };
*/

/*
    int M =  8;  //  6;  // inputs
    int N =  7;  // 16;  // nodes
    int K =  2;  //  2;  // lutsize
//    word pTruth[4] = { 0, 0, 0, ABC_CONST(0x8000000000000000) };
//    word pTruth[4] = { ABC_CONST(0x0000000000000001), 0, 0, 0 };
    word pTruth[4] = { 0, 0, 0, ABC_CONST(0x0000000000020000) };
    int pLuts[MAX_N][MAX_K] = { {0,1}, {2,3}, {4,5}, {6,7}, {8,9}, {10,11}, {12,13} };
*/

    var M int32  =  8;  //  6;  // inputs
    var N int32  =  7;  // 16;  // nodes
    var K int32  =  2;  //  2;  // lutsize
    var pTruth [4]word = [4]word{ ABC_CONST(0x0000080000020000), ABC_CONST(0x0000000000020000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000020000) };
    var pLuts [MAX_N][MAX_K]int32 = [MAX_N][MAX_K]int32{ {0,1}, {2,3}, {4,5}, {6,7}, {8,9}, {10,11}, {12,13} };

    var pValues [MAX_N*((1<<MAX_K)-1)]int32;
    var Res int32;
    var i int32;
    var k int32;
    var nLutPars int32 = (1 << K) - 1;

    //Sbd_SolverSynth( M, N, K, pLuts );

    Res = Sbd_SolverFunc( M, N, K, pLuts, pTruth, pValues );
    if Res == -1 {
        fmt.Printf("Solution does not exist.\n");
        return;
    }
    fmt.Printf("Result (compl = %d):\n", Res);
    for  i = 0; i < N; i++  {
        for  k = nLutPars-1; k >= 0; k--  {
            fmt.Printf("%d", pValues[i*nLutPars+k]);
        }
        fmt.Printf("0\n");
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

