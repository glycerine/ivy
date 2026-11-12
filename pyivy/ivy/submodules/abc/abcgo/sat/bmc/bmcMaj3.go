package bmc

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [bmcMaj3.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Majority gate synthesis.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcMaj3.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilTruth.h"
//#include "sat/glucose/AbcGlucose.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const MAJ3_OBJS = 32 // nVars + nNodes

   
type Maj3_Man_t struct {
    nVars               int32;     // inputs
    nNodes               int32;    // internal nodes
    nObjs               int32;     // total objects (nVars inputs, nNodes internal nodes)
    nWords               int32;    // the truth table size in 64-bit words
    iVar               int32;      // the next available SAT variable
    vInfo        *Vec_Wrd_t;     // nVars + nNodes + 1
    vLevels        *Vec_Int_t;   // distriction of nodes by levels
    VarMarks [MAJ3_OBJS][MAJ3_OBJS]int32; // variable marks
    ObjVals [MAJ3_OBJS]int32;             // object values
    pLits [2][MAJ3_OBJS]int32; // neg, pos literals
    nLits [3]int32;  // neg, pos, fixed literal
    pSat  *bmcg_sat_solver;      // SAT solver
}

func Maj3_ManTruth(p *Maj3_Man_t, v int32) *word { return Vec_WrdEntryP( p.vInfo, p.nWords * v ); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj3_ManValue(iMint int32, nVars int32) int32 {
    var k, Count int32;
    for  k = 0; k < nVars; k++  {
        Count += (iMint >> k) & 1;
    }
    return int32((Count > nVars/2));
}
func Maj3_ManTruthTables(p *Maj3_Man_t) *Vec_Wrd_t {
    var vInfo *Vec_Wrd_t  = p.vInfo = Vec_WrdStart( p.nWords * (p.nObjs + 1) ); 
    var i int32;
    var nMints int32 = Abc_MaxInt( 64, 1 << p.nVars );
    for  i = 0; i < p.nVars; i++  {
        Abc_TtIthVar( Maj3_ManTruth(p, i), i, p.nVars );
    }
    for  i = 0; i < nMints; i++  {
        if Maj3_ManValue(i, p.nVars) != 0 {
            Abc_TtSetBit( Maj3_ManTruth(p, p.nObjs), i ); }
    }
    //Dau_DsdPrintFromTruth( Maj3_ManTruth(p, p->nObjs), p->nVars );
    return vInfo;
}
func Maj3_ManFirstAndLevel(vLevels *Vec_Int_t, Firsts []int32, Levels []int32, nVars int32, nObjs int32) {
    var i, k, Entry, Obj int32;
    Firsts[0] = Obj;
    for  k = 0; k < nVars; k++  {
        Levels[Obj++] = 0;
    }
    Vec_IntReverseOrder( vLevels );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLevels)) && (((Entry) = Vec_IntEntry(vLevels, i)), 1); i++  {
        Firsts[i+1] = Obj;
        for  k = 0; k < Entry; k++  {
            Levels[Obj++] = i+1;
        }
    }
    Vec_IntReverseOrder( vLevels );
    assert(Obj == nObjs, "Obj == nObjs");
}
func Maj3_ManMarkup(p *Maj3_Man_t) int32 {
    var nSatVars int32  = 2; // SAT variable counter
    var nLevels int32  = Vec_IntSize(p.vLevels);
    var nSecond int32  = Vec_IntEntry(p.vLevels, 1);
    var i int32;
    var k int32;
    var Firsts [MAJ3_OBJS]int32;
    var Levels [MAJ3_OBJS]int32;
    assert(Vec_IntEntry(p.vLevels, 0) == 1, "Vec_IntEntry(p.vLevels, 0) == 1");
    assert(p.nObjs <= MAJ3_OBJS, "p.nObjs <= MAJ3_OBJS");
    assert(p.nNodes == Vec_IntSum(p.vLevels), "p.nNodes == Vec_IntSum(p.vLevels)");
    Maj3_ManFirstAndLevel( p.vLevels, Firsts, Levels, p.nVars, p.nObjs );
    // create map 
    for  i = 0; i < p.nObjs; i++  {
    for  k = 0; k < p.nObjs; k++  {
        p.VarMarks[i][k] = -1;
    }
    }
    for  k = 0; k < 3; k++  { // first node
        p.VarMarks[p.nVars][k] = 1;
    }
    for  k = 0; k < nSecond; k++  { // top node
        p.VarMarks[p.nObjs-1][p.nObjs-2-k] = 1;
    }
    for  k = 2; k < nLevels; k++  { // cascade
        p.VarMarks[Firsts[k]][Firsts[k-1]] = 1;
    }
    for  i = p.nVars+1; i < (nSecond == 3 ? p.nObjs-1 : p.nObjs); i++  {
        for  k = 0; k < Firsts[Levels[i]]; k++  {
            if p.VarMarks[i][k] == -1 {
                p.VarMarks[i][k] = nSatVars++; }
        }
    }
    //printf( "The number of variables is %d.\n", nSatVars );
    return nSatVars;
}
func Maj3_ManVarMapPrint(p *Maj3_Man_t) {
    var i int32;
    var k int32;
    var Firsts [MAJ3_OBJS]int32;
    var Levels [MAJ3_OBJS]int32;
    Maj3_ManFirstAndLevel( p.vLevels, Firsts, Levels, p.nVars, p.nObjs );
    // print
    fmt.Printf("Variable map for problem with %d inputs, %d nodes and %d levels: ", p.nVars, p.nNodes, Vec_IntSize(p.vLevels));
    Vec_IntPrint( p.vLevels );
    fmt.Printf("    ");
    fmt.Printf("      ");
    for  i = 0; i < p.nObjs; i++  {
        fmt.Printf("%3d  ", i);
    }
    fmt.Printf("\n");
    for  i = p.nObjs-1; i >= p.nVars; i--  {
        fmt.Printf(" %2d ", i);
        fmt.Printf(" %2d   ", Levels[i]);
        for  k = 0; k < p.nObjs; k++  {
            if p.VarMarks[i][k] == -1 {
                fmt.Printf("  .  ");
            } else if p.VarMarks[i][k] == 1 {
                fmt.Printf("  +  ");
            } else { 
                fmt.Printf("%3d%c ", p.VarMarks[i][k], bmcg_sat_solver_read_cex_varvalue(p.pSat, p.VarMarks[i][k]) ? '+' : ' '); }
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj3_ManFindFanin(p *Maj3_Man_t, i int32, pFanins *int32) int32 {
    var f, nFanins int32;
    p.nLits[0] = p.nLits[1] = p.nLits[2] = 0;
    for  f = 0; f < i; f++  {
        if p.VarMarks[i][f] < 0 {
            continue; }
        assert(p.VarMarks[i][f] > 0, "p.VarMarks[i][f] > 0");
        if p.VarMarks[i][f] == 1 { // fixed
        
            p.nLits[2]++;
            pFanins[nFanins++] = f;
        
        } else if bmcg_sat_solver_read_cex_varvalue(p.pSat, p.VarMarks[i][f]) { // pos
        
            p.pLits[1][p.nLits[1]++] = Abc_Var2Lit(p.VarMarks[i][f], 1);
            pFanins[nFanins++] = f;
        
        } else { // neg
            p.pLits[0][p.nLits[0]++] = Abc_Var2Lit(p.VarMarks[i][f], 0); }
    }
    return nFanins;
}
func Maj3_ManPrintSolution(p *Maj3_Man_t) {
    var i int32;
    var k int32;
    var iVar int32;
    var nFanins int32;
    var pFanins [MAJ3_OBJS]int32;
    fmt.Printf("Realization of %d-input majority using %d MAJ3 gates:\n", p.nVars, p.nNodes);
//    for ( i = p->nVars; i < p->nObjs; i++ )
    for  i = p.nObjs - 1; i >= p.nVars; i--  {
        fmt.Printf("%02d = MAJ(", i);
        nFanins = Maj3_ManFindFanin( p, i, pFanins );
        assert(nFanins == 3, "nFanins == 3");
        for  k = 0; k < 3; k++  {
            iVar = pFanins[k];
            if iVar >= 0 && iVar < p.nVars {
                fmt.Printf(" %c", 'a'+iVar);
            } else {
                fmt.Printf(" %02d", iVar); }
        }
        fmt.Printf(" )\n");
    }
}
var spStatic_Maj3_ManEval_Flag int32;
func Maj3_ManEval(p *Maj3_Man_t) int32 {
    var fUseMiddle int32  = 1;
    
    var i int32;
    var k int32;
    var iMint int32;
    var pFanins [3]int32; var pFaninsW [3]*word;
    for  i = p.nVars; i < p.nObjs; i++  {
        var nFanins int32  = Maj3_ManFindFanin( p, i, pFanins );
        assert(nFanins == 3, "nFanins == 3");
        for  k = 0; k < 3; k++  {
            pFaninsW[k] = Maj3_ManTruth( p, pFanins[k] );
        }
        Abc_TtMaj( Maj3_ManTruth(p, i), pFaninsW[0], pFaninsW[1], pFaninsW[2], p.nWords );
    }
    if fUseMiddle != 0 {
        iMint = -1;
        for  i = 0; i < (1 << p.nVars); i++  {
            var nOnes int32  = Abc_TtBitCount16(i);
            if nOnes < p.nVars/2 || nOnes > p.nVars/2+1 {
                continue; }
            if Abc_TtGetBit(Maj3_ManTruth(p, p.nObjs), i) == Abc_TtGetBit(Maj3_ManTruth(p, p.nObjs-1), i) {
                continue; }
            iMint = i;
            break;
        }
    
    } else {
        if (spStatic_Maj3_ManEval_Flag != 0) && p.nVars >= 6 {
            iMint = Abc_TtFindLastDiffBit( Maj3_ManTruth(p, p.nObjs-1), Maj3_ManTruth(p, p.nObjs), p.nVars );
        } else {
            iMint = Abc_TtFindFirstDiffBit( Maj3_ManTruth(p, p.nObjs-1), Maj3_ManTruth(p, p.nObjs), p.nVars ); }
    }
    //Flag ^= 1;
    assert(iMint < (1 << p.nVars), "iMint < (1 << p.nVars)");
    return iMint;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj3_PrintClause(pLits *int32, nLits int32) {
    var i int32;
    for  i = 0; i < nLits; i++  {
        fmt.Printf("%c%d ", Abc_LitIsCompl(pLits[i]) ? '-' : '+', Abc_Lit2Var(pLits[i]));
    }
    fmt.Printf("\n");
}
func Maj3_ManAddCnfStart(p *Maj3_Man_t) int32 {
    var i int32;
    var k int32;
    var status int32;
    var nLits int32;
    var pLits [MAJ3_OBJS]int32;
    // nodes have at least one fanin
    //printf( "Fanin clauses:\n" );
    for  i = p.nVars; i < p.nObjs; i++  {
        // check if it is already connected
        var Count int32;
        for  k = 0; k < p.nObjs; k++  { 
            Count += p.VarMarks[i][k] == 1;
        }
        assert(Count <= 3, "Count <= 3");
        if Count == 3 {
            continue; }
        // collect connections
        nLits = 0;
        for  k = 0; k < p.nObjs; k++  { 
            if p.VarMarks[i][k] > 1 {
                pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k], 0 ); }
        }
        //Maj3_PrintClause( pLits, nLits );
        assert(nLits > 0, "nLits > 0");
        if nLits > 0 && !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
            assert(false, "0"); }
    }
    // nodes have at least one fanout
    //printf( "Fanout clauses:\n" );
    for  k = 0; k < p.nObjs-1; k++  {
        // check if it is already connected
        var Count int32;
        for  i = 0; i < p.nObjs; i++  { 
            Count += p.VarMarks[i][k] == 1;
        }
        assert(Count <= 3, "Count <= 3");
        if Count > 0 {
            continue; }
        // collect connections
        nLits = 0;
        for  i = 0; i < p.nObjs; i++  { 
            if p.VarMarks[i][k] > 1 {
                pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k], 0 ); }
        }
        //Maj3_PrintClause( pLits, nLits );
        //assert( nLits > 0 );
        if nLits > 0 && !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
            assert(false, "0"); }
    }
    status = bmcg_sat_solver_solve( p.pSat, nil, 0 );
    assert(status == GLUCOSE_SAT, "status == GLUCOSE_SAT");
    Maj3_ManVarMapPrint( p );
    return 1;
}
func Maj3_ManAddCnf(p *Maj3_Man_t, iMint int32) int32 {
    // input values
    var i int32;
    var k int32;
    var j int32;
    var n int32;
    var pVals *int32 = p.ObjVals;
    for  i = 0; i < p.nVars; i++  {
        pVals[i] = (iMint >> i) & 1;
    }
    // first node value
    pVals[p.nVars] = (pVals[0] && pVals[1]) || (pVals[0] && pVals[2]) || (pVals[1] && pVals[2]);
    // last node value
    pVals[p.nObjs-1] = Maj3_ManValue(iMint, p.nVars);
    // intermediate node values
    for  i = p.nVars + 1; i < p.nObjs - 1; i++  {
        pVals[i] = p.iVar++;
    }
    //printf( "Adding clauses for minterm %d.\n", iMint );
    bmcg_sat_solver_set_nvars( p.pSat, p.iVar );
    for  n = 0; n < 2; n++  {
    for  i = p.nVars + 1; i < p.nObjs; i++  {
        //printf( "\nNode %d. Phase %d.\n", i, n );
        for  k = 0; k < p.nObjs; k++  { if p.VarMarks[i][k] >= 1 {
            // add first input
            var pLits [5]int32;
            var nLits int32;
            if pVals[k] == !n {
                continue; }
            if pVals[k] > 1 {
                pLits[nLits++] = Abc_Var2Lit( pVals[k], n ); }
            if p.VarMarks[i][k] > 1 {
                pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k], 1 ); }
            for  j = k+1; j < p.nObjs; j++  { if p.VarMarks[i][j] >= 1 {
                // add second input
                var nLits2 int32  = nLits;
                if pVals[j] == !n {
                    continue; }
                if pVals[j] > 1 {
                    pLits[nLits2++] = Abc_Var2Lit( pVals[j], n ); }
                if p.VarMarks[i][j] > 1 {
                    pLits[nLits2++] = Abc_Var2Lit( p.VarMarks[i][j], 1 ); }
                // add output
                if pVals[i] == n {
                    continue; }
                if pVals[i] > 1 {
                    pLits[nLits2++] = Abc_Var2Lit( pVals[i], !n ); }
                assert(nLits2 > 0 && nLits2 <= 5, "nLits2 > 0 && nLits2 <= 5");
                //Maj3_PrintClause( pLits, nLits2 );
                if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits2 ) {
                    return 0; }
            }
            }
        }
        }
    }
    }
    return 1;
}
func Maj3_ManAddConstraintsLazy(p *Maj3_Man_t) int32 {
    var i int32;
    var pFanins [MAJ3_OBJS]int32;
    var nConstr int32;
    //Maj3_ManVarMapPrint( p );
    for  i = p.nVars+1; i < p.nObjs; i++  {
        var nFanins int32  = Maj3_ManFindFanin( p, i, pFanins );
        if nFanins == 3 {
            continue; }
        //printf( "Node %d has %d fanins.\n", i, nFanins );
        nConstr++;
        if nFanins < 3 {
            assert(p.nLits[0] > 0, "p.nLits[0] > 0");
            //Maj3_PrintClause( p->pLits[0], p->nLits[0] );
            if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
                return -1; }
        
        } else if nFanins > 3 {
            var nLits int32  = Abc_MinInt(4 - p.nLits[2], p.nLits[1]);
            assert(nLits > 0, "nLits > 0");
            //Maj3_PrintClause( p->pLits[1], nLits );
            if !bmcg_sat_solver_addclause( p.pSat, p.pLits[1], nLits ) {
                return -1; }
        }
    }
    return nConstr;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj3_ManAlloc(nVars int32, nNodes int32, vLevels *Vec_Int_t) *Maj3_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Maj3_Man_t  = &Maj3_Man_t{};
    p.vLevels    = vLevels;
    p.nVars      = nVars;
    p.nNodes     = nNodes;
    p.nObjs      = nVars + nNodes;
    p.nWords     = Abc_TtWordNum(nVars);
    p.iVar       = Maj3_ManMarkup( p );
    p.vInfo      = Maj3_ManTruthTables( p );
    p.pSat       = bmcg_sat_solver_start();
    bmcg_sat_solver_set_nvars( p.pSat, p.iVar );
    Maj3_ManAddCnfStart( p );
    return p;
}
func Maj3_ManFree(p *Maj3_Man_t) {
    bmcg_sat_solver_stop( p.pSat );
    Vec_WrdFree( p.vInfo );
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj3_ManExactSynthesis(nVars int32, nNodes int32, fVerbose int32, vLevels *Vec_Int_t) int32 {
    var p *Maj3_Man_t;  var clkTotal abctime  = Abc_Clock();
    var i, status, nLazy, nLazyAll, iMint int32;
    fmt.Printf("Running exact synthesis for %d-input majority with %d MAJ3 gates...\n", nVars, nNodes);
    p = Maj3_ManAlloc( nVars, nNodes, vLevels );
    for  i = 0; iMint != -1; i++  {
        var clk abctime  = Abc_Clock();
        if !(Maj3_ManAddCnf( p, iMint ) != 0) {
            break; }
        for  (status = bmcg_sat_solver_solve(p.pSat, nil, 0)) == GLUCOSE_SAT  {
            nLazy = Maj3_ManAddConstraintsLazy( p );
            if nLazy == -1 {
                fmt.Printf("Became UNSAT after adding lazy constraints.\n");
                status = GLUCOSE_UNSAT;
                break;
            }
            //printf( "Added %d lazy constraints.\n\n", nLazy );
            if nLazy == 0 {
                break; }
            nLazyAll += nLazy;
        }
        if fVerbose != 0 {
            fmt.Printf("Iter %3d : ", i);
            Extra_PrintBinary( os.Stdout, (unsigned *)&iMint, p.nVars );
            fmt.Printf("  Var =%5d  ", p.iVar);
            fmt.Printf("Cla =%6d  ", bmcg_sat_solver_clausenum(p.pSat));
            fmt.Printf("Conf =%9d  ", bmcg_sat_solver_conflictnum(p.pSat));
            fmt.Printf("Lazy =%9d  ", nLazyAll);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        if status == GLUCOSE_UNSAT {
            fmt.Printf("The problem has no solution.\n");
            break;
        }
        iMint = Maj3_ManEval( p );
    }
    if iMint == -1 {
        Maj3_ManPrintSolution( p ); }
    Maj3_ManFree( p );
    Abc_PrintTime( 1, "Total runtime", Abc_Clock() - clkTotal );
    return iMint == -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj3_ManTest() int32 {
/*
    int nVars    =  5;
    int nNodes   =  4;
    int fVerbose =  1;
    int Levels[MAJ3_OBJS] = { 1, 2, 1 };
    Vec_Int_t vLevels = { 3, 3, Levels };
*/

    var nVars int32     =  7;
    var nNodes int32    =  7;
    var fVerbose int32  =  1;
    var Levels [MAJ3_OBJS]int32 = [MAJ3_OBJS]int32{ 1, 2, 2, 2 };
    var vLevels Vec_Int_t  = Vec_Int_t{ 4, 4, Levels };

/*
    int nVars    =  9;
    int nNodes   = 10;
    int fVerbose =  1;
    int Levels[MAJ3_OBJS] = { 1, 2, 3, 2, 2 };
    Vec_Int_t vLevels = { 5, 5, Levels };
*/
/*
    int nVars    =  9;
    int nNodes   = 10;
    int fVerbose =  1;
    int Levels[MAJ3_OBJS] = { 1, 1, 1, 1, 1, 1, 1, 1, 1, 1 };
    Vec_Int_t vLevels = { 10, 10, Levels };
*/

//    Maj3_Man_t * p = Maj3_ManAlloc( nVars, nNodes, &vLevels );
//    Maj3_ManFree( p );
    Maj3_ManExactSynthesis( nVars, nNodes, fVerbose, &vLevels );
    return 0;
}





   
type Zyx_Man_t struct {
    pPars      *Bmc_EsPar_t;      // parameters
    pTruth             *word;     // truth table
    nObjs               int32;      // total objects (nVars inputs, nNodes internal nodes)
    nWords               int32;     // the truth table size in 64-bit words
    LutMask               int32;    // 1 << nLutSize
    TopoBase               int32;   // (1 << nLutSize) * nObjs
    MintBase               int32;   // TopoBase + nObjs * nObjs
    vInfo        *Vec_Wrd_t;      // nVars + nNodes + 1
    vVarValues        *Vec_Int_t; // variable values
    vMidMints        *Vec_Int_t;  // middle minterms
    vUsed2        *Vec_Bit_t;     // bit masks
    vUsed3        *Vec_Bit_t;     // bit masks
    vPairs        *Vec_Int_t;     // sym var pairs
    nUsed [2]int32;
    pFanins [MAJ3_OBJS][MAJ3_OBJS]int32;  // fanins
    pLits [2][2*MAJ3_OBJS]int32; // neg/pos literals
    nLits [2]int32;   // neg/pos literal
    Counts [1024]int32;
    pSat  *bmcg_sat_solver;       // SAT solver
    clkEval           abctime;
}

func Zyx_ManTruth(p *Zyx_Man_t, v int32) *word { return Vec_WrdEntryP( p.vInfo, p.nWords * v ); }

func Zyx_FuncVar(p *Zyx_Man_t, i int32, m int32) int32 { return (p.LutMask + 1) * (i - p.pPars.nVars) + m;         }
func Zyx_TopoVar(p *Zyx_Man_t, i int32, f int32) int32 { return p.TopoBase + p.nObjs * (i - p.pPars.nVars) + f;   }
func Zyx_MintVar(p *Zyx_Man_t, m int32, i int32) int32 { return p.MintBase + p.nObjs * m + i;                       }

//static inline int     Zyx_MintVar2( Zyx_Man_t * p, int m, int i, int f ) { assert(i >= p->pPars->nVars); return p->MintBase + m * p->pPars->nNodes * (p->pPars->nLutSize + 1) + (i - p->pPars->nVars) * (p->pPars->nLutSize + 1) + f; }

func Zyx_ManIsUsed2(p *Zyx_Man_t, m int32, n int32, i int32, j int32) int32 {
    var Pos int32  = ((m * p.pPars.nNodes + n - p.pPars.nVars) * p.nObjs + i) * p.nObjs + j;
    p.nUsed[0]++;
    assert(i < n && j < n && i < j, "i < n && j < n && i < j");
    if Vec_BitEntry(p.vUsed2, Pos) {
        return 1; }
    p.nUsed[1]++;
    Vec_BitWriteEntry( p.vUsed2, Pos, 1 );
    return 0;
}

func Zyx_ManIsUsed3(p *Zyx_Man_t, m int32, n int32, i int32, j int32, k int32) int32 {
    var Pos int32  = (((m * p.pPars.nNodes + n - p.pPars.nVars) * p.nObjs + i) * p.nObjs + j) * p.nObjs + k;
    p.nUsed[0]++;
    assert(i < n && j < n && k < n && i < j && j < k, "i < n && j < n && k < n && i < j && j < k");
    if Vec_BitEntry(p.vUsed3, Pos) {
        return 1; }
    p.nUsed[1]++;
    Vec_BitWriteEntry( p.vUsed3, Pos, 1 );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Zyx_SetConstVar(p *Zyx_Man_t, Var int32, Value int32) {
    var iLit int32  = Abc_Var2Lit( Var, !Value );
    var status int32  = bmcg_sat_solver_addclause( p.pSat, &iLit, 1 );
    assert((status != 0), "status");
    assert(Vec_IntEntry(p.vVarValues, Var) == -1, "Vec_IntEntry(p.vVarValues, Var) == -1");
    Vec_IntWriteEntry( p.vVarValues, Var, Value );
    //printf( "Setting var %d to value %d.\n", Var, Value );
}
func Zyx_ManSetupVars(p *Zyx_Man_t) {
    var i, k, m int32;
    var pSpec *word;
    if p.pPars.fMajority {
        pSpec = Zyx_ManTruth(p, p.nObjs);
    } else {
        pSpec = p.pTruth;
    }
    // set unused functionality vars to 0
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        Zyx_SetConstVar( p, Zyx_FuncVar(p, i, 0), 0 );
    }
    // set unused topology vars
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        for  k = i; k < p.nObjs; k++  {
            Zyx_SetConstVar( p, Zyx_TopoVar(p, i, k), 0 );
        }
    }
    // connect topmost node
    Zyx_SetConstVar( p, Zyx_TopoVar(p, p.nObjs-1, p.nObjs-2), 1 );
    // connect first node
    if p.pPars.fMajority {
        for  k = 0; k < p.pPars.nVars; k++  {
            Zyx_SetConstVar( p, Zyx_TopoVar(p, p.pPars.nVars, k), k < 3 );
        } }
    // set minterm vars
    for  m = 0; m < (1 << p.pPars.nVars); m++  {
        for  i = 0; i < p.pPars.nVars; i++  {
            Zyx_SetConstVar( p, Zyx_MintVar(p, m, i), (m >> i) & 1 );
        }
        Zyx_SetConstVar( p, Zyx_MintVar(p, m, p.nObjs-1), Abc_TtGetBit(pSpec, m) );
    }
}
func Zyx_ManAddCnfStart(p *Zyx_Man_t) int32 {
    // nodes have fanins
    var i int32;
    var k int32;
    var pLits [MAJ3_OBJS]int32;
    //printf( "Adding initial clauses:\n" );
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        var nLits int32;
        for  k = 0; k < i; k++  {
            pLits[nLits++] = Abc_Var2Lit( Zyx_TopoVar(p, i, k), 0 );
        }
        assert(nLits > 0, "nLits > 0");
        if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
            return 0; }
    }
 
    // nodes have fanouts
    for  k = 0; k < p.nObjs-1; k++  {
        var nLits int32;
        for  i = p.pPars.nVars; i < p.nObjs; i++  {
            pLits[nLits++] = Abc_Var2Lit( Zyx_TopoVar(p, i, k), 0 );
        }
        assert(nLits > 0, "nLits > 0");
        if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
            return 0; }
    }
 
    // two input functions
    if p.pPars.nLutSize != 2 {
        return 1; }
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        for  k = 0; k < 3; k++  {
            pLits[0] = Abc_Var2Lit( Zyx_FuncVar(p, i, 1), k==1 );
            pLits[1] = Abc_Var2Lit( Zyx_FuncVar(p, i, 2), k==2 );
            pLits[2] = Abc_Var2Lit( Zyx_FuncVar(p, i, 3), k!=0 );
            if !bmcg_sat_solver_addclause( p.pSat, pLits, 3 ) {
                return 0; }
        }
        if p.pPars.fOnlyAnd {
            pLits[0] = Abc_Var2Lit( Zyx_FuncVar(p, i, 1), 1 );
            pLits[1] = Abc_Var2Lit( Zyx_FuncVar(p, i, 2), 1 );
            pLits[2] = Abc_Var2Lit( Zyx_FuncVar(p, i, 3), 0 );
            if !bmcg_sat_solver_addclause( p.pSat, pLits, 3 ) {
                return 0; }
        }
    }
    return 1;
}
func Zyx_ManPrintVarMap(p *Zyx_Man_t, fSat int32) {
    var i, k, nTopoVars int32;
    fmt.Printf("      ");
    for  k = 0; k < p.nObjs-1; k++  {
        fmt.Printf("%3d  ", k);
    }
    fmt.Printf("\n");
    for  i = p.nObjs-1; i >= p.pPars.nVars; i--  {
        fmt.Printf("%3d   ", i);
        for  k = 0; k < p.nObjs-1; k++  {
            var iVar int32  = Zyx_TopoVar(p, i, k);
            if Vec_IntEntry(p.vVarValues, iVar) == -1 {
                fmt.Printf("%3d%c ", iVar, (fSat && bmcg_sat_solver_read_cex_varvalue(p.pSat, iVar)) ? '*' : ' ');
nTopoVars++;
            } else {
                fmt.Printf("%3d  ", Vec_IntEntry(p.vVarValues, iVar)); }
        }
        fmt.Printf("\n");
    }
    if fSat != 0 { return; }
    fmt.Printf("Using %d active functionality vars and %d active topology vars (out of %d SAT vars).\n", p.pPars.fMajority ? 0 : p.pPars.nNodes * p.LutMask, nTopoVars, bmcg_sat_solver_varnum(p.pSat));
}
func Zyx_PrintClause(pLits *int32, nLits int32) {
    var i int32;
    for  i = 0; i < nLits; i++  {
        fmt.Printf("%c%d ", Abc_LitIsCompl(pLits[i]) ? '-' : '+', Abc_Lit2Var(pLits[i]));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Zyx_ManValue(iMint int32, nVars int32) int32 {
    var k, Count int32;
    for  k = 0; k < nVars; k++  {
        Count += (iMint >> k) & 1;
    }
    return int32((Count > nVars/2));
}
func Zyx_ManTruthTables(p *Zyx_Man_t, pTruth *word) *Vec_Wrd_t {
    var vInfo *Vec_Wrd_t  = p.vInfo = Vec_WrdStart( p.nWords * (p.nObjs + 1) ); 
    var i int32;
    var nMints int32 = Abc_MaxInt( 64, 1 << p.pPars.nVars );
    assert(p.pPars.fMajority == (pTruth == nil), "p.pPars.fMajority == (pTruth == nil)");
    for  i = 0; i < p.pPars.nVars; i++  {
        Abc_TtIthVar( Zyx_ManTruth(p, i), i, p.pPars.nVars );
    }
    if p.pPars.fMajority {
        for  i = 0; i < nMints; i++  {
            if Zyx_ManValue(i, p.pPars.nVars) != 0 {
                Abc_TtSetBit( Zyx_ManTruth(p, p.nObjs), i ); }
        }
        for  i = 0; i < nMints; i++  {
            if Abc_TtBitCount16(i) == p.pPars.nVars/2 || Abc_TtBitCount16(i) == p.pPars.nVars/2+1 {
                Vec_IntPush( p.vMidMints, i ); }
        }
    }
    //Dau_DsdPrintFromTruth( Zyx_ManTruth(p, p->nObjs), p->pPars->nVars );
    return vInfo;
}
func Zyx_ManCreateSymVarPairs(pTruth *word, nVars int32) *Vec_Int_t {
    var vPairs *Vec_Int_t  = Vec_IntAlloc( 100 ); 
    var i int32;
    var k int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    var Cof0, Cof1 [64]word; 
    var Cof01, Cof10 [64]word;
    assert(nVars <= 12, "nVars <= 12");
    for  i = 0; i < nVars; i++  {
        Abc_TtCofactor0p( Cof0, pTruth, nWords, i );
        Abc_TtCofactor1p( Cof1, pTruth, nWords, i );
        for  k = i+1; k < nVars; k++  {
            Abc_TtCofactor1p( Cof01, Cof0, nWords, k );
            Abc_TtCofactor0p( Cof10, Cof1, nWords, k );
            if Abc_TtEqual( Cof01, Cof10, nWords ) {
                 Vec_IntPushTwo( vPairs, i, k ); }
        }
    }
    return vPairs;
}
func Zyx_ManAlloc(pPars *Bmc_EsPar_t, pTruth *word) *Zyx_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Zyx_Man_t  = &Zyx_Man_t{};
    p.pPars      = pPars;
    p.pTruth     = pTruth;
    p.nObjs      = p.pPars.nVars + p.pPars.nNodes;
    p.nWords     = Abc_TtWordNum(p.pPars.nVars);
    p.LutMask    = (1 << p.pPars.nLutSize) - 1;
    p.TopoBase   = (1 << p.pPars.nLutSize) * p.pPars.nNodes;
    p.MintBase   = p.TopoBase + p.pPars.nNodes * p.nObjs;
    p.vVarValues = Vec_IntStartFull( p.MintBase + (1 << p.pPars.nVars) * p.nObjs );
    p.vMidMints  = Vec_IntAlloc( 1 << p.pPars.nVars );
    p.vInfo      = Zyx_ManTruthTables( p, pTruth );
    p.vPairs     = Zyx_ManCreateSymVarPairs( p.pPars.fMajority ? Zyx_ManTruth(p, p.nObjs) : pTruth, p.pPars.nVars );
    p.pSat       = bmcg_sat_solver_start();
    if pPars.fUseIncr {
        if p.pPars.nLutSize == 2 || p.pPars.fMajority {
            p.vUsed2     = Vec_BitStart( (1 << p.pPars.nVars) * p.pPars.nNodes * p.nObjs * p.nObjs );
        } else if p.pPars.nLutSize == 3 {
            p.vUsed3     = Vec_BitStart( (1 << p.pPars.nVars) * p.pPars.nNodes * p.nObjs * p.nObjs * p.nObjs ); }
    }
    bmcg_sat_solver_set_nvars( p.pSat, p.MintBase + (1 << p.pPars.nVars) * p.nObjs );
    Zyx_ManSetupVars( p );
    Zyx_ManAddCnfStart( p );
    Zyx_ManPrintVarMap( p, 0 );
    return p;
}
func Zyx_ManFree(p *Zyx_Man_t) {
    bmcg_sat_solver_stop( p.pSat );
    Vec_WrdFree( p.vInfo );
    Vec_BitFreeP( &p.vUsed2 );
    Vec_BitFreeP( &p.vUsed3 );
    Vec_IntFree( p.vPairs );
    Vec_IntFree( p.vMidMints );
    Vec_IntFree( p.vVarValues );
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Zyx_ManCollectFanins(p *Zyx_Man_t, i int32) int32 {
    var k, Val int32;
    assert(i >= p.pPars.nVars && i < p.nObjs, "i >= p.pPars.nVars && i < p.nObjs");
    p.nLits[0] = p.nLits[1] = 0;
    for  k = 0; k < i; k++  {
        Val = bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i, k));
        p.pFanins[i][p.nLits[1]] = k;
        p.pLits[Val][p.nLits[Val]++] = Abc_Var2Lit(Zyx_TopoVar(p, i, k), Val);
    }
    return p.nLits[1];
}
func Zyx_ManAddCnfLazyTopo(p *Zyx_Man_t) int32 {
    var i int32;
    var k int32;
    var j int32;
    var Entry [2]int32;
    var Node [2]int32;
    var nLazy int32;
    // fanin count
    //printf( "Adding topology clauses.\n" );
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        var nFanins int32  = Zyx_ManCollectFanins( p, i );
        if nFanins == p.pPars.nLutSize {
            continue; }
        nLazy++;
        assert(nFanins == p.nLits[1], "nFanins == p.nLits[1]");
        if p.nLits[1] > p.pPars.nLutSize {
            p.nLits[1] = p.pPars.nLutSize + 1;
            //Zyx_PrintClause( p->pLits[1], p->nLits[1] );
            if !bmcg_sat_solver_addclause( p.pSat, p.pLits[1], p.nLits[1] ) {
                return -1; }
        
        } else { // if ( p->nLits[1] < p->pPars->nLutSize )
        
            //Zyx_PrintClause( p->pLits[0], p->nLits[0] );
            if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
                return -1; }
        }
    }
    if (nLazy != 0) || !p.pPars.fOrderNodes {
        return nLazy; }
    // ordering
    for  i = p.pPars.nVars + 1; i < p.nObjs; i++  {
        for  k = p.pPars.nLutSize - 1; k >= 0; k--  {
            if p.pFanins[i-1][k] != p.pFanins[i][k] {
                break; }
        }
        if k == -1 { // fanins are equal
        
            if p.pPars.fMajority {
                continue; }
            // compare by LUT functions
            for  k = p.LutMask; k >= 0; k--  {
                if bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i-1, k)) != 
                     bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i  , k)) {
                    break; }
            }
            if k == -1 { // truth tables cannot be equal
                continue; }
            // rule out these truth tables
            if bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i-1, k)) == 0 && 
                 bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i  , k)) == 1 {
                continue;
            }
            nLazy++;
            assert(bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i-1, k)) == 1, "bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i-1, k)) == 1");
            assert(bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i  , k)) == 0, "bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i  , k)) == 0");
            // rule out this order
            p.nLits[0] = 0;
            for  j = p.LutMask; j >= k; j--  {
                var ValA int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i-1, j));
                var ValB int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i,   j));
                p.pLits[0][p.nLits[0]++] = Abc_Var2Lit(Zyx_FuncVar(p, i-1, j), ValA );
                p.pLits[0][p.nLits[0]++] = Abc_Var2Lit(Zyx_FuncVar(p, i,   j), ValB );
            }
            if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
                return -1; }
            continue;
        }
        if p.pFanins[i-1][k] < p.pFanins[i][k] {
            assert(bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i-1, p.pFanins[i][k])) == 0, "bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i-1, p.pFanins[i][k])) == 0");
            assert(bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i  , p.pFanins[i][k])) == 1, "bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i  , p.pFanins[i][k])) == 1");
            continue;
        }
        nLazy++;
        assert(bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i-1, p.pFanins[i-1][k])) == 1, "bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i-1, p.pFanins[i-1][k])) == 1");
        assert(bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i  , p.pFanins[i-1][k])) == 0, "bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i  , p.pFanins[i-1][k])) == 0");
        // rule out this order
        p.nLits[0] = 0;
        for  j = p.pFanins[i-1][k]; j < p.nObjs-1; j++  {
            var ValA int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i-1, j));
            var ValB int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i,   j));
            p.pLits[0][p.nLits[0]++] = Abc_Var2Lit(Zyx_TopoVar(p, i-1, j), ValA );
            p.pLits[0][p.nLits[0]++] = Abc_Var2Lit(Zyx_TopoVar(p, i,   j), ValB );
        }
        //printf( "\n" );
        //Zyx_ManPrintVarMap( p, 1 );
        //Zyx_PrintClause( p->pLits[0], p->nLits[0] );
        if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
            return -1; }
        //break;
    }
    // check symmetric variables
    // inlining of macro Vec_IntForEachEntryDouble
    for  k = 0; (k+1 < Vec_IntSize(p.vPairs)) && (((Entry[0]) = Vec_IntEntry(p.vPairs, k)), 1) && (((Entry[1]) = Vec_IntEntry(p.vPairs, k+1)), 1); k += 2  {
        assert(Entry[0] < Entry[1], "Entry[0] < Entry[1]");
        for  j = 0; j < 2; j++  {
            Node[j] = -1;
            for  i = p.pPars.nVars; i < p.nObjs; i++  {
                if bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i, Entry[j])) {
                    Node[j] = i;
                    break;
                }
            }
            assert(Node[j] >= p.pPars.nVars, "Node[j] >= p.pPars.nVars");
        }
        // compare the nodes
        if Node[0] <= Node[1] {
            continue; }
        assert(Node[0] > Node[1], "Node[0] > Node[1]");
        // create blocking clause
        nLazy++;
        assert(bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, Node[1], Entry[0])) == 0, "bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, Node[1], Entry[0])) == 0");
        assert(bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, Node[1], Entry[1])) == 1, "bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, Node[1], Entry[1])) == 1");
        // rule out this order
        p.nLits[0] = 0;
        for  j = p.pPars.nVars; j <= Node[1]; j++  {
            var ValA int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, j, Entry[0]));
            var ValB int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, j, Entry[1]));
            p.pLits[0][p.nLits[0]++] = Abc_Var2Lit(Zyx_TopoVar(p, j, Entry[0]), ValA );
            p.pLits[0][p.nLits[0]++] = Abc_Var2Lit(Zyx_TopoVar(p, j, Entry[1]), ValB );
        }
        if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
            return -1; }
    }
    return nLazy;
}
func Zyx_ManAddCnfBlockSolution(p *Zyx_Man_t) int32 {
    var vLits *Vec_Int_t  = Vec_IntAlloc( 100 ); var i, k int32;
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        var nFanins int32  = Zyx_ManCollectFanins( p, i );
        assert(nFanins == p.pPars.nLutSize, "nFanins == p.pPars.nLutSize");
        for  k = 0; k < p.pPars.nLutSize; k++  {
            Vec_IntPush( vLits, Abc_Var2Lit(Zyx_TopoVar(p, i, p.pFanins[i][k]), 1) );
        }
    }
    //Zyx_ManPrintVarMap( p, 1 );
    //Zyx_PrintClause( Vec_IntArray(vLits), Vec_IntSize(vLits) );
    if !bmcg_sat_solver_addclause( p.pSat, Vec_IntArray(vLits), Vec_IntSize(vLits) ) {
        return 0; }
    Vec_IntFree( vLits );
    return 1;
}
func Zyx_ManAddCnfLazyFunc2(p *Zyx_Man_t, iMint int32) int32 {
    var i, k, n, j, s int32;
    assert(!p.pPars.fMajority || p.pPars.nLutSize == 3, "!p.pPars.fMajority || p.pPars.nLutSize == 3");
    //printf( "Adding functionality clauses for minterm %d.\n", iMint );
    p.Counts[iMint]++;
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        var nFanins int32  = Zyx_ManCollectFanins( p, i );
        assert(nFanins == p.pPars.nLutSize, "nFanins == p.pPars.nLutSize");
        if p.pPars.fMajority {
            var Sets [3][2]int32 = [3][2]int32{{0, 1}, {0, 2}, {1, 2}};
            for  k = 0; k < 3; k++  {
                if Zyx_ManIsUsed2(p, iMint, i, p.pFanins[i][Sets[k][0]], p.pFanins[i][Sets[k][1]]) != 0 {
                    continue; }
                for  n = 0; n < 2; n++  {
                    p.nLits[0] = 0;
                    for  s = 0; s < 2; s++  {
                        p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_TopoVar(p, i, p.pFanins[i][Sets[k][s]]), 1 );
                        p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_MintVar(p, iMint, p.pFanins[i][Sets[k][s]]), n );
                    }
                    p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_MintVar(p, iMint, i), !n );
                    if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
                        return 0; }
                }
            }
        
        } else {
            if p.pPars.nLutSize == 2 && (Zyx_ManIsUsed2(p, iMint, i, p.pFanins[i][0], p.pFanins[i][1]) != 0) {
                continue; }
            if p.pPars.nLutSize == 3 && (Zyx_ManIsUsed3(p, iMint, i, p.pFanins[i][0], p.pFanins[i][1], p.pFanins[i][2]) != 0) {
                continue; }
            for  k = 0; k <= p.LutMask; k++  {
            for  n = 0; n < 2; n++  {
                p.nLits[0] = 0;
                p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_FuncVar(p, i, k), n );
                for  j = 0; j < p.pPars.nLutSize; j++  {
                    p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_TopoVar(p, i, p.pFanins[i][j]), 1 );
                    p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_MintVar(p, iMint, p.pFanins[i][j]), (k >> j) & 1 );
                }
                p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_MintVar(p, iMint, i), !n );
                if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
                    return 0; }
            }
            }
        }
    }
    return 1;
}

func Zyx_ManAddCnfLazyFunc(p *Zyx_Man_t, iMint int32) int32 {
    var i, k, n, j, s, t, u int32;
    //printf( "Adding clauses for minterm %d with value %d.\n", iMint, Value );
    assert(!p.pPars.fMajority && p.pPars.nLutSize < 4, "!p.pPars.fMajority && p.pPars.nLutSize < 4");
    p.Counts[iMint]++;
    if p.pPars.nLutSize == 2 {
        for  i = p.pPars.nVars; i < p.nObjs; i++  {
        for  s = 0;               s < i;        s++  {
        for  t = s+1;             t < i;        t++  {
            p.pFanins[i][0] = s;
            p.pFanins[i][1] = t;
            for  k = 0; k <= p.LutMask; k++  {
            for  n = 0; n < 2; n++  {
                p.nLits[0] = 0;
                p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_FuncVar(p, i, k), n );
                for  j = 0; j < p.pPars.nLutSize; j++  {
                    p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_TopoVar(p, i, p.pFanins[i][j]), 1 );
                    p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_MintVar(p, iMint, p.pFanins[i][j]), (k >> j) & 1 );
                }
                p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_MintVar(p, iMint, i), !n );
                if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
                    return 0; }
            }
            }
        }
        }
        }
    
    } else if p.pPars.nLutSize == 3 {
        for  i = p.pPars.nVars; i < p.nObjs; i++  {
        for  s = 0;               s < i;        s++  {
        for  t = s+1;             t < i;        t++  {
        for  u = t+1;             u < i;        u++  {
            p.pFanins[i][0] = s;
            p.pFanins[i][1] = t;
            p.pFanins[i][2] = u;
            for  k = 0; k <= p.LutMask; k++  {
            for  n = 0; n < 2; n++  {
                p.nLits[0] = 0;
                p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_FuncVar(p, i, k), n );
                for  j = 0; j < p.pPars.nLutSize; j++  {
                    p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_TopoVar(p, i, p.pFanins[i][j]), 1 );
                    p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_MintVar(p, iMint, p.pFanins[i][j]), (k >> j) & 1 );
                }
                p.pLits[0][p.nLits[0]++] = Abc_Var2Lit( Zyx_MintVar(p, iMint, i), !n );
                if !bmcg_sat_solver_addclause( p.pSat, p.pLits[0], p.nLits[0] ) {
                    return 0; }
            }
            }
        }
        }
        }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Zyx_ManPrintSolutionFile(p *Zyx_Man_t, fCompl int32, fFirst int32) {
    var pFile *os.File;
    var FileName [1000]byte;  var i, k int32;
    if fCompl != 0 { Abc_TtNot( p.pTruth, Abc_TtWordNum(p.pPars.nVars) ); }
    Abc_TtWriteHexRev( FileName, p.pTruth, p.pPars.nVars );
    if fCompl != 0 { Abc_TtNot( p.pTruth, Abc_TtWordNum(p.pPars.nVars) ); }
    sprintf( FileName + (1 << (p.pPars.nVars-2)), "-%d-%d.bool", p.pPars.nLutSize, p.pPars.nNodes );
    pFile = fopen( FileName, fFirst ? "wb" : "ab" );
    if pFile == nil {
        printf( "Cannot open input file \"%s\".\n", FileName );
        return;
    }
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        fmt.Fprintf(pFile, "%c", 'A' + i);
        if p.pPars.fMajority {
            fmt.Fprintf(pFile, "maj3");
        } else {
            for  k = p.LutMask; k >= 0; k--  {
                fmt.Fprintf(pFile, "%d", bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i, k)) ^ (i == p.nObjs - 1 && fCompl));
            }
            for  k = 0; k < i; k++  {
                if bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i, k)) {
                    if k >= 0 && k < p.pPars.nVars {
                        fmt.Fprintf(pFile, "%c", 'a' + k);
                    } else {
                        fmt.Fprintf(pFile, "%c", 'A' + k); }
                }
            }
        }
        fmt.Fprintf(pFile, "\n");
    }
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
    printf( "Dumped solution into file \"%s\".\n", FileName );
}
func Zyx_ManPrintSolution(p *Zyx_Man_t, fCompl int32, fFirst int32) {
    var i, k int32;
    fmt.Printf("Realization of given %d-input function using %d %d-input %s:\n", p.pPars.nVars, p.pPars.nNodes, p.pPars.nLutSize, p.pPars.fMajority ? "MAJ-gates" : "LUTs");
    for  i = p.nObjs - 1; i >= p.pPars.nVars; i--  {
        fmt.Printf("%02d = ", i);
        if p.pPars.fMajority {
            fmt.Printf("MAJ3");
        } else {
            fmt.Printf("%d\'b", 1 << p.pPars.nLutSize);
            for  k = p.LutMask; k >= 0; k--  {
                fmt.Printf("%d", bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i, k)) ^ (i == p.nObjs - 1 && fCompl));
            }
        }
        fmt.Printf("(");
        for  k = 0; k < i; k++  {
            if bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_TopoVar(p, i, k)) {
                if k >= 0 && k < p.pPars.nVars {
                    fmt.Printf(" %c", 'a'+k);
                } else {
                    fmt.Printf(" %02d", k); }
            }
        }
        fmt.Printf(" )\n");
    }
    if !p.pPars.fMajority {
        Zyx_ManPrintSolutionFile( p, fCompl, fFirst ); }
}
var spStatic_Zyx_ManEval_Flag int32;
func Zyx_ManEval(p *Zyx_Man_t) int32 {
    
    //abctime clk = Abc_Clock();
    var i, k, j, iMint int32; var pFaninsW [6]*word;
var pSpec *word;
    for  i = p.pPars.nVars; i < p.nObjs; i++  {
        var nFanins int32  = Zyx_ManCollectFanins( p, i );
        assert(nFanins == p.pPars.nLutSize, "nFanins == p.pPars.nLutSize");
        for  k = 0; k < p.pPars.nLutSize; k++  {
            pFaninsW[k] = Zyx_ManTruth( p, p.pFanins[i][k] );
        }
        if p.pPars.fMajority {
            Abc_TtMaj( Zyx_ManTruth(p, i), pFaninsW[0], pFaninsW[1], pFaninsW[2], p.nWords );
        } else {
            Abc_TtConst0( Zyx_ManTruth(p, i), p.nWords );
            for  k = 1; k <= p.LutMask; k++  {
                if bmcg_sat_solver_read_cex_varvalue(p.pSat, Zyx_FuncVar(p, i, k)) {
                    Abc_TtConst1( Zyx_ManTruth(p, p.nObjs), p.nWords );
                    for  j = 0; j < p.pPars.nLutSize; j++  {
                        Abc_TtAndCompl( Zyx_ManTruth(p, p.nObjs), Zyx_ManTruth(p, p.nObjs), 0, pFaninsW[j], !((k >> j) & 1), p.nWords );
                    }
                    Abc_TtOr( Zyx_ManTruth(p, i), Zyx_ManTruth(p, i), Zyx_ManTruth(p, p.nObjs), p.nWords );
                }
            }
        }
    }
    if p.pPars.fMajority {
        pSpec = Zyx_ManTruth(p, p.nObjs);
    } else {
        pSpec = p.pTruth;
    }
    if p.pPars.fMajority {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vMidMints)) && (((iMint) = Vec_IntEntry(p.vMidMints, i)), 1); i++  {
            if Abc_TtGetBit(pSpec, iMint) != Abc_TtGetBit(Zyx_ManTruth(p, p.nObjs-1), iMint) {
                return iMint; }
        }
        return -1;
    
    } else {
        if (spStatic_Zyx_ManEval_Flag != 0) && p.pPars.nVars >= 6 {
            iMint = Abc_TtFindLastDiffBit( Zyx_ManTruth(p, p.nObjs-1), pSpec, p.pPars.nVars );
        } else {
            iMint = Abc_TtFindFirstDiffBit( Zyx_ManTruth(p, p.nObjs-1), pSpec, p.pPars.nVars ); }
    }
    //Flag ^= 1;
    assert(iMint < (1 << p.pPars.nVars), "iMint < (1 << p.pPars.nVars)");
    //p->clkEval += Abc_Clock() - clk;
    return iMint;
}
func Zyx_ManEvalStats(p *Zyx_Man_t) {
    var i int32;
    for  i = 0; i < (1 << p.pPars.nVars); i++  {
        fmt.Printf("%d=%d ", i, p.Counts[i]);
    }
    fmt.Printf("\n");
}
func Zyx_ManPrint(p *Zyx_Man_t, Iter int32, iMint int32, nLazyAll int32, clk abctime) {
    fmt.Printf("Iter %6d : ", Iter);
    Extra_PrintBinary( os.Stdout, (unsigned *)&iMint, p.pPars.nVars );
    fmt.Printf("  ");
    fmt.Printf("Cla =%9d  ", bmcg_sat_solver_clausenum(p.pSat));
    fmt.Printf("Lazy =%6d  ", nLazyAll);
    fmt.Printf("Conf =%9d  ", bmcg_sat_solver_conflictnum(p.pSat));
    Abc_PrintTime( 1, "Time", clk );
    //Zyx_ManEvalStats( p );
    //Abc_PrintTime( 1, "Eval", p->clkEval );
}
func Zyx_ManExactSynthesis(pPars *Bmc_EsPar_t) {
    var status, Iter, iMint, fCompl, nLazyAll, nSols int32;
    var clkTotal abctime = Abc_Clock();
    var clk abctime = Abc_Clock();  var p *Zyx_Man_t; 
    var pTruth [16]word; 
    if !pPars.fMajority {
        Abc_TtReadHex( pTruth, pPars.pTtStr );
        if pTruth[0] & 1 { fCompl = 1; Abc_TtNot( pTruth, Abc_TtWordNum(pPars.nVars) ); }
    }
    assert(pPars.nVars <= 10, "pPars.nVars <= 10");
    assert(pPars.nLutSize <= 6, "pPars.nLutSize <= 6");
    p = Zyx_ManAlloc( pPars, pPars.fMajority ? nil : pTruth );
    fmt.Printf("Running exact synthesis for %d-input function with %d %d-input %s...\n", p.pPars.nVars, p.pPars.nNodes, p.pPars.nLutSize, p.pPars.fMajority ? "MAJ-gates" : "LUTs");
    for  Iter = 0 ; ; Iter++  {
        for  (status = bmcg_sat_solver_solve(p.pSat, nil, 0)) == GLUCOSE_SAT  {
            var nLazy int32  = Zyx_ManAddCnfLazyTopo( p );
            if nLazy == -1 {
                fmt.Printf("Became UNSAT after adding lazy constraints.\n");
                status = GLUCOSE_UNSAT;
                break;
            }
            //printf( "Added %d lazy constraints.\n\n", nLazy );
            if nLazy == 0 {
                break; }
            nLazyAll += nLazy;
        }
        if status == GLUCOSE_UNSAT {
            break; }
        // find mismatch
        iMint = Zyx_ManEval( p );
        if iMint == -1 {
            if pPars.fEnumSols {
                nSols++;
                if pPars.fVerbose {
                    Zyx_ManPrint( p, Iter, iMint, nLazyAll, Abc_Clock() - clkTotal );
                    clk = Abc_Clock();
                }
                Zyx_ManPrintSolution( p, fCompl, nSols==1 );
                if !(Zyx_ManAddCnfBlockSolution( p ) != 0) {
                    status = GLUCOSE_UNSAT;
                    break;
                }
                continue;
            
            } else {
                break; }
        }
        if pPars.fUseIncr ? !Zyx_ManAddCnfLazyFunc2(p, iMint) : !Zyx_ManAddCnfLazyFunc(p, iMint) {
            fmt.Printf("Became UNSAT after adding constraints for minterm %d\n", iMint);
            status = GLUCOSE_UNSAT;
            break;
        }
        status = bmcg_sat_solver_solve( p.pSat, nil, 0 );
        if pPars.fVerbose && (!pPars.fUseIncr || Iter % 100 == 0) {
            Zyx_ManPrint( p, Iter, iMint, nLazyAll, Abc_Clock() - clk );
            clk = Abc_Clock();
        }
        if status == GLUCOSE_UNSAT {
            break; }
    }
    if pPars.fVerbose {
        Zyx_ManPrint( p, Iter, iMint, nLazyAll, Abc_Clock() - clkTotal ); }
    if pPars.fEnumSols {
        fmt.Printf("Finished enumerating %d solutions.\n", nSols);
    } else if iMint == -1 {
        Zyx_ManPrintSolution( p, fCompl, 1 );
    } else {
        fmt.Printf("The problem has no solution.\n"); }
    //Zyx_ManEvalStats( p );
    fmt.Printf("Added = %d.  Tried = %d.  ", p.nUsed[1], p.nUsed[0]);
    Abc_PrintTime( 1, "Total runtime", Abc_Clock() - clkTotal );
    Zyx_ManFree( p );
}



/**Function*************************************************************

  Synopsis    [Tests solution to the exact synthesis problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Zyx_TestGetTruthTablePars(pFileName string, pTruth *word, nVars *int32, nLutSize *int32, nNodes *int32) int32 {
    var Symb byte;
    var pCur string;
    var pBuffer string = Abc_UtilStrsav( pFileName );
    var nLength int32;
    for  pCur = pBuffer; *pCur; pCur++  {
        if !Abc_TtIsHexDigit(*pCur) {
            break; }
    }
    Symb = *pCur; *pCur = 0;
    nLength = int32(len(pBuffer));
    if nLength == 1 {
        *nVars = 2;
    } else if nLength == 2 {
        *nVars = 3;
    } else if nLength == 4 {
        *nVars = 4;
    } else if nLength == 8 {
        *nVars = 5;
    } else if nLength == 16 {
        *nVars = 6;
    } else if nLength == 32 {
        *nVars = 7;
    } else if nLength == 64 {
        *nVars = 8;
    } else {
        pBuffer = nil;
        fmt.Printf("Invalid truth table size.\n");
        return 0;
    }
    Abc_TtReadHex( pTruth, pBuffer );
    *pCur = Symb;
    // read LUT size
    for  *pCur && *pCur++ != '-'  {; }
    if *pCur == 0 {
        pBuffer = nil;
        fmt.Printf("Expecting \'-\' after truth table before LUT size.\n");
        return 0;
    }
    // read node count
    *nLutSize = strconv.Atoi(pCur);
    for  *pCur && *pCur++ != '-'  {; }
    if *pCur == 0 {
        pBuffer = nil;
        fmt.Printf("Expecting \'-\' after LUT size before node count.\n");
        return 0;
    }
    *nNodes = strconv.Atoi(pCur);
    pBuffer = nil;
    return 1;
}

func Zyx_TestTruth(vInfo *Vec_Wrd_t, i int32, nWords int32) *word { return Vec_WrdEntryP(vInfo, nWords * i); }

func Zyx_TestCreateTruthTables(nVars int32, nNodes int32) *Vec_Wrd_t {
    var i int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    var vInfo *Vec_Wrd_t  = Vec_WrdStart( nWords * (nVars + nNodes + 1) ); 
    for  i = 0; i < nVars; i++  {
        Abc_TtIthVar( Zyx_TestTruth(vInfo, i, nWords), i, nVars );
    }
    //Dau_DsdPrintFromTruth( Maj3_ManTruth(p, p->nObjs), p->nVars );
    return vInfo;
}
func Zyx_TestReadNode(pLine string, vTruths *Vec_Wrd_t, nVars int32, nLutSize int32, iObj int32) int32 {
    var k int32;
    var j int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    var pFaninsW [6]*word; var pTruth, pFanins string;
    var pThis *word;
    var pLast *word = Zyx_TestTruth( vTruths, Vec_WrdSize(vTruths)/nWords - 1, nWords );
    if pLine[len(pLine)-1] == '\n' { pLine[len(pLine)-1] = 0; }
    if pLine[len(pLine)-1] == '\r' { pLine[len(pLine)-1] = 0; }
    if pLine[0] == 0 {
        return 0; }
    if int32(len(pLine)) != 1 + nLutSize + (1 << nLutSize) {
        fmt.Printf("Node representation has %d chars (expecting %d chars).\n", int32(len(pLine)), 1 + nLutSize + (1 << nLutSize));
        return 0;
    }
    if pLine[0] != 'A' + iObj {
        fmt.Printf("The output node in line %s is not correct.\n", pLine);
        return 0;
    }
    pTruth  = pLine + 1;
    pFanins = pTruth + (1 << nLutSize);
    for  k = nLutSize - 1; k >= 0; k--  {
        pFaninsW[k] = Zyx_TestTruth( vTruths, pFanins[k] >= 'a' ? pFanins[k] - 'a' : pFanins[k] - 'A', nWords );
    }
    pThis = Zyx_TestTruth(vTruths, iObj, nWords);
    Abc_TtConst0( pThis, nWords );
    for  k = 0; k < (1 << nLutSize); k++  {
        if pTruth[(1 << nLutSize) - 1 - k] == '0' {
            continue; }
        Abc_TtConst1( pLast, nWords );
        for  j = 0; j < nLutSize; j++  {
            Abc_TtAndCompl( pLast, pLast, 0, pFaninsW[j], !((k >> j) & 1), nWords );
        }
        Abc_TtOr( pThis, pThis, pLast, nWords );
    }
    //Dau_DsdPrintFromTruth( pThis, nVars );
    return 1;
}
func Zyx_TestExact(pFileName string) {
    var iObj int32;
    var nStrs int32;
    var nVars int32 = -1;
    var nLutSize int32 = -1;
    var nNodes int32 = -1;
    var pImpl *word;
    var pSpec [4]word; var vInfo *Vec_Wrd_t; var Line [1000]byte;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open input file \"%s\".\n", pFileName);
        return;
    }
    if !(Zyx_TestGetTruthTablePars( pFileName, pSpec, &nVars, &nLutSize, &nNodes ) != 0) {
        return; }
    if nVars > 8 {
        fmt.Printf("This tester does not support functions with more than 8 inputs.\n");
        return;
    }
    if nLutSize > 6 {
        fmt.Printf("This tester does not support nodes with more than 6 inputs.\n");
        return;
    }
    if nNodes > 16 {
        fmt.Printf("This tester does not support structures with more than 16 inputs.\n");
        return;
    }
    vInfo = Zyx_TestCreateTruthTables( nVars, nNodes );
    for  iObj = nVars; fgets(Line, 1000, pFile) != nil; iObj++  {
        if Zyx_TestReadNode( Line, vInfo, nVars, nLutSize, iObj ) != 0 {
            continue; }
        if iObj != nVars + nNodes {
            fmt.Printf("The number of nodes in the structure is not correct.\n");
            break;
        }
        nStrs++;
        pImpl = Zyx_TestTruth( vInfo, iObj-1, Abc_TtWordNum(nVars) );
        if Abc_TtEqual( pImpl, pSpec, Abc_TtWordNum(nVars) ) {
            fmt.Printf("Structure %3d : Verification successful.\n", nStrs);
        } else {
            fmt.Printf("Structure %3d : Verification FAILED.\n", nStrs);
            fmt.Printf("Implementation: "); Dau_DsdPrintFromTruth( pImpl, nVars );
            fmt.Printf("Specification:  "); Dau_DsdPrintFromTruth( pSpec, nVars );
        }
        iObj = nVars - 1;
    }
    Vec_WrdFree( vInfo );
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

