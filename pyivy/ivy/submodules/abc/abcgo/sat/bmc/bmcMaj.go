package bmc

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [bmcMaj.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Exact synthesis with majority gates.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 1, 2017.]

  Revision    [$Id: bmcMaj.c,v 1.00 2017/10/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilTruth.h"
//#include "sat/glucose/AbcGlucose.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const MAJ_NOBJS =  32 // Const0 + Const1 + nVars + nNodes

   
type Maj_Man_t struct {
    nVars               int32;     // inputs
    nNodes               int32;    // internal nodes
    nObjs               int32;     // total objects (2 consts, nVars inputs, nNodes internal nodes)
    nWords               int32;    // the truth table size in 64-bit words
    iVar               int32;      // the next available SAT variable
    fUseConst               int32; // use constant fanins
    fUseLine               int32;  // use cascade topology
    vInfo        *Vec_Wrd_t;     // Const0 + Const1 + nVars + nNodes + Maj(nVars)
    VarMarks [MAJ_NOBJS][3][MAJ_NOBJS]int32; // variable marks
    VarVals [MAJ_NOBJS+2]int32; // values of the first 2 + nVars variables
    vOutLits        *Vec_Wec_t;  // output vars
    pSat  *bmcg_sat_solver;      // SAT solver
}

func Maj_ManTruth(p *Maj_Man_t, v int32) *word { return Vec_WrdEntryP( p.vInfo, p.nWords * v ); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj_ManValue(iMint int32, nVars int32) int32 {
    var k, Count int32;
    for  k = 0; k < nVars; k++  {
        Count += (iMint >> k) & 1;
    }
    return int32((Count > nVars/2));
}
func Maj_ManTruthTables(p *Maj_Man_t) *Vec_Wrd_t {
    var vInfo *Vec_Wrd_t  = p.vInfo = Vec_WrdStart( p.nWords * (p.nObjs + 1) ); 
    var i int32;
    var nMints int32 = Abc_MaxInt( 64, 1 << p.nVars );
    Abc_TtFill( Maj_ManTruth(p, 1), p.nWords );
    for  i = 0; i < p.nVars; i++  {
        Abc_TtIthVar( Maj_ManTruth(p, i+2), i, p.nVars );
    }
    for  i = 0; i < nMints; i++  {
        if Maj_ManValue(i, p.nVars) != 0 {
            Abc_TtSetBit( Maj_ManTruth(p, p.nObjs), i ); }
    }
    //Dau_DsdPrintFromTruth( Maj_ManTruth(p, p->nObjs), p->nVars );
    return vInfo;
}
func Maj_ManMarkup(p *Maj_Man_t) int32 {
    var i, k, j int32;
    p.iVar = 1;
    assert(p.nObjs <= MAJ_NOBJS, "p.nObjs <= MAJ_NOBJS");
    // make exception for the first node
    i = p.nVars + 2;
    for  k = 0; k < 3; k++  {
        j = 4-k;
        Vec_WecPush( p.vOutLits, j, Abc_Var2Lit(p.iVar, 0) );
        p.VarMarks[i][k][j] = p.iVar++;
    }
    // assign variables for other nodes
    for  i = p.nVars + 3; i < p.nObjs; i++  {
        for  k = 0; k < 3; k++  {
            if p.fUseLine && k == 0 {
                j = i-1;
                Vec_WecPush( p.vOutLits, j, Abc_Var2Lit(p.iVar, 0) );
                p.VarMarks[i][k][j] = p.iVar++;
                continue;
            }
            for  j = (p.fUseConst && k == 2) ? 0 : 2; j < i - k; j++  {
                Vec_WecPush( p.vOutLits, j, Abc_Var2Lit(p.iVar, 0) );
                p.VarMarks[i][k][j] = p.iVar++;
            }
        }
    }
    fmt.Printf("The number of parameter variables = %d.\n", p.iVar);
    return p.iVar;
    // printout
    for  i = p.nVars + 2; i < p.nObjs; i++  {
        fmt.Printf("Node %d\n", i);
        for  j = 0; j < p.nObjs; j++  {
            for  k = 0; k < 3; k++  {
                fmt.Printf("%3d ", p.VarMarks[i][k][j]);
            }
            fmt.Printf("\n");
        }
    }
    return p.iVar;
}
func Maj_ManAlloc(nVars int32, nNodes int32, fUseConst int32, fUseLine int32) *Maj_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Maj_Man_t  = &Maj_Man_t{};
    p.nVars      = nVars;
    p.nNodes     = nNodes;
    p.nObjs      = 2 + nVars + nNodes;
    p.fUseConst  = fUseConst;
    p.fUseLine   = fUseLine;
    p.nWords     = Abc_TtWordNum(nVars);
    p.vOutLits   = Vec_WecStart( p.nObjs );
    p.iVar       = Maj_ManMarkup( p );
    p.VarVals[1] = 1;
    p.vInfo      = Maj_ManTruthTables( p );
    p.pSat       = bmcg_sat_solver_start();
    bmcg_sat_solver_set_nvars( p.pSat, p.iVar );
    return p;
}
func Maj_ManFree(p *Maj_Man_t) {
    bmcg_sat_solver_stop( p.pSat );
    Vec_WrdFree( p.vInfo );
    Vec_WecFree( p.vOutLits );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj_ManFindFanin(p *Maj_Man_t, i int32, k int32) int32 {
    var j int32;
    var Count int32;
    var iVar int32 = -1;
    for  j = 0; j < p.nObjs; j++  {
        if p.VarMarks[i][k][j] && bmcg_sat_solver_read_cex_varvalue(p.pSat, p.VarMarks[i][k][j]) {
            iVar = j;
            Count++;
        }
    }
    assert(Count == 1, "Count == 1");
    return iVar;
}
var spStatic_Maj_ManEval_Flag int32;
func Maj_ManEval(p *Maj_Man_t) int32 {
    var fUseMiddle int32  = 1;
    
    var i, k, iMint int32; var pFanins [3]*word;
    for  i = p.nVars + 2; i < p.nObjs; i++  {
        for  k = 0; k < 3; k++  {
            pFanins[k] = Maj_ManTruth( p, Maj_ManFindFanin(p, i, k) );
        }
        Abc_TtMaj( Maj_ManTruth(p, i), pFanins[0], pFanins[1], pFanins[2], p.nWords );
    }
    if fUseMiddle != 0 {
        iMint = -1;
        for  i = 0; i < (1 << p.nVars); i++  {
            var nOnes int32  = Abc_TtBitCount16(i);
            if nOnes < p.nVars/2 || nOnes > p.nVars/2+1 {
                continue; }
            if Abc_TtGetBit(Maj_ManTruth(p, p.nObjs), i) == Abc_TtGetBit(Maj_ManTruth(p, p.nObjs-1), i) {
                continue; }
            iMint = i;
            break;
        }
    
    } else {
        if (spStatic_Maj_ManEval_Flag != 0) && p.nVars >= 6 {
            iMint = Abc_TtFindLastDiffBit( Maj_ManTruth(p, p.nObjs-1), Maj_ManTruth(p, p.nObjs), p.nVars );
        } else {
            iMint = Abc_TtFindFirstDiffBit( Maj_ManTruth(p, p.nObjs-1), Maj_ManTruth(p, p.nObjs), p.nVars ); }
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
func Maj_ManPrintSolution(p *Maj_Man_t) {
    var i, k, iVar int32;
    fmt.Printf("Realization of %d-input majority using %d MAJ3 gates:\n", p.nVars, p.nNodes);
//    for ( i = p->nVars + 2; i < p->nObjs; i++ )
    for  i = p.nObjs - 1; i >= p.nVars + 2; i--  {
        fmt.Printf("%02d = MAJ(", i-2);
        for  k = 2; k >= 0; k--  {
            iVar = Maj_ManFindFanin( p, i, k );
            if iVar >= 2 && iVar < p.nVars + 2 {
                fmt.Printf(" %c", 'a'+iVar-2);
            } else if iVar < 2 {
                fmt.Printf(" %d", iVar);
            } else {
                fmt.Printf(" %02d", iVar-2); }
        }
        fmt.Printf(" )\n");
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Maj_ManAddCnfStart(p *Maj_Man_t) int32 {
    var pLits [MAJ_NOBJS]int32;
    var pLits2 [2]int32;
    var i int32;
    var j int32;
    var k int32;
    var n int32;
    var m int32;
    // input constraints
    for  i = p.nVars + 2; i < p.nObjs; i++  {
        for  k = 0; k < 3; k++  {
            var nLits int32;
            for  j = 0; j < p.nObjs; j++  {
                if p.VarMarks[i][k][j] {
                    pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k][j], 0 ); }
            }
            assert(nLits > 0, "nLits > 0");
            // input uniqueness
            if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                return 0; }
            for  n = 0;   n < nLits; n++  {
            for  m = n+1; m < nLits; m++  {
                pLits2[0] = Abc_LitNot(pLits[n]);
                pLits2[1] = Abc_LitNot(pLits[m]);
                if !bmcg_sat_solver_addclause( p.pSat, pLits2, 2 ) {
                    return 0; }
            }
            }
            if k == 2 {
                break; }
            // symmetry breaking
            for  j = 0; j < p.nObjs; j++  { if p.VarMarks[i][k][j] {
            for  n = j; n < p.nObjs; n++  { if p.VarMarks[i][k+1][n] {
                pLits2[0] = Abc_Var2Lit( p.VarMarks[i][k][j], 1 );
                pLits2[1] = Abc_Var2Lit( p.VarMarks[i][k+1][n], 1 );
                if !bmcg_sat_solver_addclause( p.pSat, pLits2, 2 ) {
                    return 0; }
            }
            } }
            }
        }
    }
    // outputs should be used
    for  i = 2; i < p.nObjs - 1; i++  {
        var vArray *Vec_Int_t  = Vec_WecEntry(p.vOutLits, i);
        assert(Vec_IntSize(vArray) > 0, "Vec_IntSize(vArray) > 0");
        if !bmcg_sat_solver_addclause( p.pSat, Vec_IntArray(vArray), Vec_IntSize(vArray) ) {
            return 0; }
    }
    return 1;
}
func Maj_ManAddCnf(p *Maj_Man_t, iMint int32) int32 {
    // save minterm values
    var i int32;
    var k int32;
    var n int32;
    var j int32;
    var Value int32 = Maj_ManValue(iMint, p.nVars);
    for  i = 0; i < p.nVars; i++  {
        p.VarVals[i+2] = (iMint >> i) & 1;
    }
    bmcg_sat_solver_set_nvars( p.pSat, p.iVar + 4*p.nNodes );
    //printf( "Adding clauses for minterm %d.\n", iMint );
    for  i = p.nVars + 2; i < p.nObjs; i++  {
        // fanin connectivity
        var iBaseSatVarI int32  = p.iVar + 4*(i - p.nVars - 2);
        for  k = 0; k < 3; k++  {
            for  j = 0; j < p.nObjs; j++  { if p.VarMarks[i][k][j] {
                var iBaseSatVarJ int32  = p.iVar + 4*(j - p.nVars - 2);
                for  n = 0; n < 2; n++  {
                    var pLits [3]int32;
                    var nLits int32;
                    pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k][j], 1 );
                    pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + k, n );
                    if j >= p.nVars + 2 {
                        pLits[nLits++] = Abc_Var2Lit( iBaseSatVarJ + 3, !n );
                    } else if p.VarVals[j] == n {
                        continue; }
                    if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                        return 0; }
                }
            }
            }
        }
        // node functionality
        for  n = 0; n < 2; n++  {
            if i == p.nObjs - 1 && n == Value {
                continue; }
            for  k = 0; k < 3; k++  {
                var pLits [3]int32;
                var nLits int32;
                if k != 0 { pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 0, n ); }
                if k != 1 { pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 1, n ); }
                if k != 2 { pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 2, n ); }
                if i != p.nObjs - 1 { pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 3, !n ); }
                assert(nLits <= 3, "nLits <= 3");
                if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                    return 0; }
            }
        }
    }
    p.iVar += 4*p.nNodes;
    return 1;
}
func Maj_ManExactSynthesis(nVars int32, nNodes int32, fUseConst int32, fUseLine int32, fVerbose int32) int32 {
    var i, iMint int32;
    var clkTotal abctime  = Abc_Clock();
    var p *Maj_Man_t  = Maj_ManAlloc( nVars, nNodes, fUseConst, fUseLine );
    var status int32  = Maj_ManAddCnfStart( p );
    assert((status != 0), "status");
    fmt.Printf("Running exact synthesis for %d-input majority with %d MAJ3 gates...\n", p.nVars, p.nNodes);
    for  i = 0; iMint != -1; i++  {
        var clk abctime  = Abc_Clock();
        if !(Maj_ManAddCnf( p, iMint ) != 0) {
            break; }
        status = bmcg_sat_solver_solve( p.pSat, nil, 0 );
        if fVerbose != 0 {
            fmt.Printf("Iter %3d : ", i);
            Extra_PrintBinary( os.Stdout, (unsigned *)&iMint, p.nVars );
            fmt.Printf("  Var =%5d  ", p.iVar);
            fmt.Printf("Cla =%6d  ", bmcg_sat_solver_clausenum(p.pSat));
            fmt.Printf("Conf =%9d  ", bmcg_sat_solver_conflictnum(p.pSat));
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        if status == GLUCOSE_UNSAT {
            fmt.Printf("The problem has no solution.\n");
            break;
        }
        iMint = Maj_ManEval( p );
    }
    if iMint == -1 {
        Maj_ManPrintSolution( p ); }
    Maj_ManFree( p );
    Abc_PrintTime( 1, "Total runtime", Abc_Clock() - clkTotal );
    return iMint == -1;
}








   
type Exa_Man_t struct {
    pPars      *Bmc_EsPar_t;     // parameters
    nVars               int32;     // inputs
    nNodes               int32;    // internal nodes
    nObjs               int32;     // total objects (nVars inputs + nNodes internal nodes)
    nWords               int32;    // the truth table size in 64-bit words
    iVar               int32;      // the next available SAT variable
    pTruth             *word;    // truth table
    vInfo        *Vec_Wrd_t;     // nVars + nNodes + 1
    VarMarks [MAJ_NOBJS][2][MAJ_NOBJS]int32; // variable marks
    VarVals [MAJ_NOBJS]int32; // values of the first nVars variables
    vOutLits        *Vec_Wec_t;  // output vars
    pSat  *bmcg_sat_solver;      // SAT solver
}

func Exa_ManTruth(p *Exa_Man_t, v int32) *word { return Vec_WrdEntryP( p.vInfo, p.nWords * v ); }


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Exa_ManTruthTables(p *Exa_Man_t) *Vec_Wrd_t {
    var vInfo *Vec_Wrd_t  = p.vInfo = Vec_WrdStart( p.nWords * (p.nObjs+1) ); var i int32;
    for  i = 0; i < p.nVars; i++  {
        Abc_TtIthVar( Exa_ManTruth(p, i), i, p.nVars );
    }
    //Dau_DsdPrintFromTruth( Exa_ManTruth(p, p->nObjs), p->nVars );
    return vInfo;
}
func Exa_ManMarkup(p *Exa_Man_t) int32 {
    var i, k, j int32;
    assert(p.nObjs <= MAJ_NOBJS, "p.nObjs <= MAJ_NOBJS");
    // assign functionality
    p.iVar = 1 + p.nNodes * 3;
    // assign connectivity variables
    for  i = p.nVars; i < p.nObjs; i++  {
        for  k = 0; k < 2; k++  {
            if p.pPars.fFewerVars && i == p.nObjs - 1 && k == 0 {
                j = p.nObjs - 2;
                Vec_WecPush( p.vOutLits, j, Abc_Var2Lit(p.iVar, 0) );
                p.VarMarks[i][k][j] = p.iVar++;
                continue;
            }
            for  j = p.pPars.fFewerVars ? 1 - k : 0; j < i - k; j++  {
                Vec_WecPush( p.vOutLits, j, Abc_Var2Lit(p.iVar, 0) );
                p.VarMarks[i][k][j] = p.iVar++;
            }
        }
    }
    fmt.Printf("The number of parameter variables = %d.\n", p.iVar);
    return p.iVar;
    // printout
    for  i = p.nVars; i < p.nObjs; i++  {
        fmt.Printf("Node %d\n", i);
        for  j = 0; j < p.nObjs; j++  {
            for  k = 0; k < 2; k++  {
                fmt.Printf("%3d ", p.VarMarks[i][k][j]);
            }
            fmt.Printf("\n");
        }
    }
    return p.iVar;
}
func Exa_ManAlloc(pPars *Bmc_EsPar_t, pTruth *word) *Exa_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Exa_Man_t  = &Exa_Man_t{};
    p.pPars      = pPars;
    p.nVars      = pPars.nVars;
    p.nNodes     = pPars.nNodes;
    p.nObjs      = pPars.nVars + pPars.nNodes;
    p.nWords     = Abc_TtWordNum(pPars.nVars);
    p.pTruth     = pTruth;
    p.vOutLits   = Vec_WecStart( p.nObjs );
    p.iVar       = Exa_ManMarkup( p );
    p.vInfo      = Exa_ManTruthTables( p );
    p.pSat       = bmcg_sat_solver_start();
    bmcg_sat_solver_set_nvars( p.pSat, p.iVar );
    if pPars.RuntimeLim {
        bmcg_sat_solver_set_runtime_limit( p.pSat, Abc_Clock() + pPars.RuntimeLim * int64(time.Second) ); }
    return p;
}
func Exa_ManFree(p *Exa_Man_t) {
    bmcg_sat_solver_stop( p.pSat );
    Vec_WrdFree( p.vInfo );
    Vec_WecFree( p.vOutLits );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Exa_ManFindFanin(p *Exa_Man_t, i int32, k int32) int32 {
    var j int32;
    var Count int32;
    var iVar int32 = -1;
    for  j = 0; j < p.nObjs; j++  {
        if p.VarMarks[i][k][j] && bmcg_sat_solver_read_cex_varvalue(p.pSat, p.VarMarks[i][k][j]) {
            iVar = j;
            Count++;
        }
    }
    assert(Count == 1, "Count == 1");
    return iVar;
}
var spStatic_Exa_ManEval_Flag int32;
func Exa_ManEval(p *Exa_Man_t) int32 {
    
    var i, k, iMint int32; var pFanins [2]*word;
    for  i = p.nVars; i < p.nObjs; i++  {
        var iVarStart int32  = 1 + 3*(i - p.nVars);
        for  k = 0; k < 2; k++  {
            pFanins[k] = Exa_ManTruth( p, Exa_ManFindFanin(p, i, k) );
        }
        Abc_TtConst0( Exa_ManTruth(p, i), p.nWords );
        for  k = 1; k < 4; k++  {
            if !bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart+k-1) {
                continue; }
            Abc_TtAndCompl( Exa_ManTruth(p, p.nObjs), pFanins[0], !(k&1), pFanins[1], !(k>>1), p.nWords );
            Abc_TtOr( Exa_ManTruth(p, i), Exa_ManTruth(p, i), Exa_ManTruth(p, p.nObjs), p.nWords );
        }
    }
    if (spStatic_Exa_ManEval_Flag != 0) && p.nVars >= 6 {
        iMint = Abc_TtFindLastDiffBit( Exa_ManTruth(p, p.nObjs-1), p.pTruth, p.nVars );
    } else {
        iMint = Abc_TtFindFirstDiffBit( Exa_ManTruth(p, p.nObjs-1), p.pTruth, p.nVars ); }
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
func Exa_ManDumpBlif(p *Exa_Man_t, fCompl int32) {
    var Buffer [1000]byte;
    var FileName [1100]byte;
    var pFile *os.File;
    var i, k, iVar int32;
    if fCompl != 0 {
        Abc_TtNot( p.pTruth, p.nWords ); }
    Extra_PrintHexadecimalString( Buffer, (unsigned *)p.pTruth, p.nVars );
    if fCompl != 0 {
        Abc_TtNot( p.pTruth, p.nWords ); }
    sprintf( FileName, "%s_%d_%d.blif", Buffer, 2, p.nNodes );
    pFile = fopen( FileName, "wb" );
    fprintf( pFile, "# Realization of the %d-input function %s using %d two-input gates:\n", p.nVars, Buffer, p.nNodes );
    fprintf( pFile, ".model %s_%d_%d\n", Buffer, 2, p.nNodes );
    fmt.Fprintf(pFile, ".inputs");
    for  i = 0; i < p.nVars; i++  {
        fmt.Fprintf(pFile, " %c", 'a'+i);
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".outputs F\n");
    for  i = p.nObjs - 1; i >= p.nVars; i--  {
        var iVarStart int32  = 1 + 3*(i - p.nVars);
        var Val1 int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart);
        var Val2 int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart+1);
        var Val3 int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart+2);

        fmt.Fprintf(pFile, ".names");
        for  k = 1; k >= 0; k--  {
            iVar = Exa_ManFindFanin( p, i, k );
            if iVar >= 0 && iVar < p.nVars {
                fmt.Fprintf(pFile, " %c", 'a'+iVar);
            } else {
                fmt.Fprintf(pFile, " %02d", iVar); }
        }
        if i == p.nObjs - 1 {
            fmt.Fprintf(pFile, " F\n");
        } else {
            fmt.Fprintf(pFile, " %02d\n", i); }
        if i == p.nObjs - 1 && (fCompl != 0) {
            fmt.Fprintf(pFile, "00 1\n"); }
        if (i == p.nObjs - 1 && fCompl) ^ Val1 {
            fmt.Fprintf(pFile, "01 1\n"); }
        if (i == p.nObjs - 1 && fCompl) ^ Val2 {
            fmt.Fprintf(pFile, "10 1\n"); }
        if (i == p.nObjs - 1 && fCompl) ^ Val3 {
            fmt.Fprintf(pFile, "11 1\n"); }
    }
    fmt.Fprintf(pFile, ".end\n\n");
    _ = (pFile).Close();
    printf( "Solution was dumped into file \"%s\".\n", FileName );
}
func Exa_ManPrintSolution(p *Exa_Man_t, fCompl int32) {
    var i, k, iVar int32;
    fmt.Printf("Realization of given %d-input function using %d two-input gates:\n", p.nVars, p.nNodes);
//    for ( i = p->nVars + 2; i < p->nObjs; i++ )
    for  i = p.nObjs - 1; i >= p.nVars; i--  {
        var iVarStart int32  = 1 + 3*(i - p.nVars);
        var Val1 int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart);
        var Val2 int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart+1);
        var Val3 int32  = bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart+2);
        if i == p.nObjs - 1 && (fCompl != 0) {
            fmt.Printf("%02d = 4\'b%d%d%d1(", i, !Val3, !Val2, !Val1);
        } else {
            fmt.Printf("%02d = 4\'b%d%d%d0(", i, Val3, Val2, Val1); }
        for  k = 1; k >= 0; k--  {
            iVar = Exa_ManFindFanin( p, i, k );
            if iVar >= 0 && iVar < p.nVars {
                fmt.Printf(" %c", 'a'+iVar);
            } else {
                fmt.Printf(" %02d", iVar); }
        }
        fmt.Printf(" )\n");
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Exa_ManAddCnfStart(p *Exa_Man_t, fOnlyAnd int32) int32 {
    var pLits [MAJ_NOBJS]int32;
    var pLits2 [2]int32;
    var i int32;
    var j int32;
    var k int32;
    var n int32;
    var m int32;
    // input constraints
    for  i = p.nVars; i < p.nObjs; i++  {
        var iVarStart int32  = 1 + 3*(i - p.nVars);
        for  k = 0; k < 2; k++  {
            var nLits int32;
            for  j = 0; j < p.nObjs; j++  {
                if p.VarMarks[i][k][j] {
                    pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k][j], 0 ); }
            }
            assert(nLits > 0, "nLits > 0");
            // input uniqueness
            if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                return 0; }
            for  n = 0;   n < nLits; n++  {
            for  m = n+1; m < nLits; m++  {
                pLits2[0] = Abc_LitNot(pLits[n]);
                pLits2[1] = Abc_LitNot(pLits[m]);
                if !bmcg_sat_solver_addclause( p.pSat, pLits2, 2 ) {
                    return 0; }
            }
            }
            if k == 1 {
                break; }
            // symmetry breaking
            for  j = 0; j < p.nObjs; j++  { if p.VarMarks[i][k][j] {
            for  n = j; n < p.nObjs; n++  { if p.VarMarks[i][k+1][n] {
                pLits2[0] = Abc_Var2Lit( p.VarMarks[i][k][j], 1 );
                pLits2[1] = Abc_Var2Lit( p.VarMarks[i][k+1][n], 1 );
                if !bmcg_sat_solver_addclause( p.pSat, pLits2, 2 ) {
                    return 0; }
            }
            } }
            }
        }
#ifdef USE_NODE_ORDER
        // node ordering
        for ( j = p->nVars; j < i; j++ )
        for ( n = 0;   n < p->nObjs; n++ ) if ( p->VarMarks[i][0][n] )
        for ( m = n+1; m < p->nObjs; m++ ) if ( p->VarMarks[j][0][m] )
        {
            pLits2[0] = Abc_Var2Lit( p->VarMarks[i][0][n], 1 );
            pLits2[1] = Abc_Var2Lit( p->VarMarks[j][0][m], 1 );
            if ( !bmcg_sat_solver_addclause( p->pSat, pLits2, 2 ) )
                return 0;
        }
#endif
        // two input functions
        for  k = 0; k < 3; k++  {
            pLits[0] = Abc_Var2Lit( iVarStart,   k==1 );
            pLits[1] = Abc_Var2Lit( iVarStart+1, k==2 );
            pLits[2] = Abc_Var2Lit( iVarStart+2, k!=0 );
            if !bmcg_sat_solver_addclause( p.pSat, pLits, 3 ) {
                return 0; }
        }
        if fOnlyAnd != 0 {
            pLits[0] = Abc_Var2Lit( iVarStart,   1 );
            pLits[1] = Abc_Var2Lit( iVarStart+1, 1 );
            pLits[2] = Abc_Var2Lit( iVarStart+2, 0 );
            if !bmcg_sat_solver_addclause( p.pSat, pLits, 3 ) {
                return 0; }
        }
    }
    // outputs should be used
    for  i = 0; i < p.nObjs - 1; i++  {
        var vArray *Vec_Int_t  = Vec_WecEntry(p.vOutLits, i);
        assert(Vec_IntSize(vArray) > 0, "Vec_IntSize(vArray) > 0");
        if !bmcg_sat_solver_addclause( p.pSat, Vec_IntArray(vArray), Vec_IntSize(vArray) ) {
            return 0; }
    }
    return 1;
}
func Exa_ManAddCnf(p *Exa_Man_t, iMint int32) int32 {
    // save minterm values
    var i int32;
    var k int32;
    var n int32;
    var j int32;
    var Value int32 = Abc_TtGetBit(p.pTruth, iMint);
    for  i = 0; i < p.nVars; i++  {
        p.VarVals[i] = (iMint >> i) & 1;
    }
    bmcg_sat_solver_set_nvars( p.pSat, p.iVar + 3*p.nNodes );
    //printf( "Adding clauses for minterm %d with value %d.\n", iMint, Value );
    for  i = p.nVars; i < p.nObjs; i++  {
        // fanin connectivity
        var iVarStart int32  = 1 + 3*(i - p.nVars);
        var iBaseSatVarI int32  = p.iVar + 3*(i - p.nVars);
        for  k = 0; k < 2; k++  {
            for  j = 0; j < p.nObjs; j++  { if p.VarMarks[i][k][j] {
                var iBaseSatVarJ int32  = p.iVar + 3*(j - p.nVars);
                for  n = 0; n < 2; n++  {
                    var pLits [3]int32;
                    var nLits int32;
                    pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k][j], 1 );
                    pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + k, n );
                    if j >= p.nVars {
                        pLits[nLits++] = Abc_Var2Lit( iBaseSatVarJ + 2, !n );
                    } else if p.VarVals[j] == n {
                        continue; }
                    if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                        return 0; }
                }
            }
            }
        }
        // node functionality
        for  n = 0; n < 2; n++  {
            if i == p.nObjs - 1 && n == Value {
                continue; }
            for  k = 0; k < 4; k++  {
                var pLits [4]int32;
                var nLits int32;
                if k == 0 && n == 1 {
                    continue; }
                pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 0, (k&1)  );
                pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 1, (k>>1) );
                if i != p.nObjs - 1 { pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 2, !n ); }
                if k > 0 {             pLits[nLits++] = Abc_Var2Lit( iVarStart +  k-1,  n ); }
                assert(nLits <= 4, "nLits <= 4");
                if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                    return 0; }
            }
        }
    }
    p.iVar += 3*p.nNodes;
    return 1;
}
func Exa_ManExactSynthesis(pPars *Bmc_EsPar_t) {
    var i int32;
    var status int32;
    var iMint int32 = 1;
    var clkTotal abctime  = Abc_Clock();
    var p *Exa_Man_t; var fCompl int32;
    var pTruth [16]word; Abc_TtReadHex( pTruth, pPars.pTtStr );
    assert(pPars.nVars <= 10, "pPars.nVars <= 10");
    p = Exa_ManAlloc( pPars, pTruth );
    if pTruth[0] & 1 { fCompl = 1; Abc_TtNot( pTruth, p.nWords ); }
    status = Exa_ManAddCnfStart( p, pPars.fOnlyAnd );
    assert((status != 0), "status");
    fmt.Printf("Running exact synthesis for %d-input function with %d two-input gates...\n", p.nVars, p.nNodes);
    for  i = 0; iMint != -1; i++  {
        var clk abctime  = Abc_Clock();
        if !(Exa_ManAddCnf( p, iMint ) != 0) {
            break; }
        status = bmcg_sat_solver_solve( p.pSat, nil, 0 );
        if pPars.fVerbose {
            fmt.Printf("Iter %3d : ", i);
            Extra_PrintBinary( os.Stdout, (unsigned *)&iMint, p.nVars );
            fmt.Printf("  Var =%5d  ", p.iVar);
            fmt.Printf("Cla =%6d  ", bmcg_sat_solver_clausenum(p.pSat));
            fmt.Printf("Conf =%9d  ", bmcg_sat_solver_conflictnum(p.pSat));
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        if status == GLUCOSE_UNSAT {
            fmt.Printf("The problem has no solution.\n");
            break;
        }
        if status == GLUCOSE_UNDEC {
            fmt.Printf("The problem timed out after %d sec.\n", pPars.RuntimeLim);
            break;
        }
        iMint = Exa_ManEval( p );
    }
    if iMint == -1 {
        Exa_ManPrintSolution( p, fCompl );
        Exa_ManDumpBlif( p, fCompl );
    }
    Exa_ManFree( p );
    Abc_PrintTime( 1, "Total runtime", Abc_Clock() - clkTotal );
}




   
type Exa3_Man_t struct {
    pPars      *Bmc_EsPar_t;     // parameters
    nVars               int32;     // inputs
    nNodes               int32;    // internal nodes
    nLutSize               int32;  // lut size
    LutMask               int32;   // lut mask
    nObjs               int32;     // total objects (nVars inputs + nNodes internal nodes)
    nWords               int32;    // the truth table size in 64-bit words
    iVar               int32;      // the next available SAT variable
    pTruth             *word;    // truth table
    vInfo        *Vec_Wrd_t;     // nVars + nNodes + 1
    vUsed2        *Vec_Bit_t;    // bit masks
    vUsed3        *Vec_Bit_t;    // bit masks
    VarMarks [MAJ_NOBJS][6][MAJ_NOBJS]int32; // variable marks
    VarVals [MAJ_NOBJS]int32; // values of the first nVars variables
    vOutLits        *Vec_Wec_t;  // output vars
    pSat  *bmcg_sat_solver;      // SAT solver
    nUsed [2]int32;
}

func Exa3_ManTruth(p *Exa3_Man_t, v int32) *word { return Vec_WrdEntryP( p.vInfo, p.nWords * v ); }

func Exa3_ManIsUsed2(p *Exa3_Man_t, m int32, n int32, i int32, j int32) int32 {
    var Pos int32  = ((m * p.pPars.nNodes + n - p.pPars.nVars) * p.nObjs + i) * p.nObjs + j;
    p.nUsed[0]++;
    assert(i < n && j < n && i < j, "i < n && j < n && i < j");
    if Vec_BitEntry(p.vUsed2, Pos) {
        return 1; }
    p.nUsed[1]++;
    Vec_BitWriteEntry( p.vUsed2, Pos, 1 );
    return 0;
}

func Exa3_ManIsUsed3(p *Exa3_Man_t, m int32, n int32, i int32, j int32, k int32) int32 {
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
func Exa3_ManTruthTables(p *Exa3_Man_t) *Vec_Wrd_t {
    var vInfo *Vec_Wrd_t  = p.vInfo = Vec_WrdStart( p.nWords * (p.nObjs+1) ); var i int32;
    for  i = 0; i < p.nVars; i++  {
        Abc_TtIthVar( Exa3_ManTruth(p, i), i, p.nVars );
    }
    //Dau_DsdPrintFromTruth( Exa3_ManTruth(p, p->nObjs), p->nVars );
    return vInfo;
}
func Exa3_ManMarkup(p *Exa3_Man_t) int32 {
    var i, k, j int32;
    assert(p.nObjs <= MAJ_NOBJS, "p.nObjs <= MAJ_NOBJS");
    // assign functionality variables
    p.iVar = 1 + p.LutMask * p.nNodes;
    // assign connectivity variables
    for  i = p.nVars; i < p.nObjs; i++  {
        for  k = 0; k < p.nLutSize; k++  {
            if p.pPars.fFewerVars && i == p.nObjs - 1 && k == 0 {
                j = p.nObjs - 2;
                Vec_WecPush( p.vOutLits, j, Abc_Var2Lit(p.iVar, 0) );
                p.VarMarks[i][k][j] = p.iVar++;
                continue;
            }
            for  j = p.pPars.fFewerVars ? p.nLutSize - 1 - k : 0; j < i - k; j++  {
                Vec_WecPush( p.vOutLits, j, Abc_Var2Lit(p.iVar, 0) );
                p.VarMarks[i][k][j] = p.iVar++;
            }
        }
    }
    fmt.Printf("The number of parameter variables = %d.\n", p.iVar);
    return p.iVar;
    // printout
    for  i = p.nObjs - 1; i >= p.nVars; i--  {
        fmt.Printf("       Node %2d\n", i);
        for  j = 0; j < p.nObjs; j++  {
            fmt.Printf("Fanin %2d : ", j);
            for  k = 0; k < p.nLutSize; k++  {
                fmt.Printf("%3d ", p.VarMarks[i][k][j]);
            }
            fmt.Printf("\n");
        }
    }
    return p.iVar;
}
func Exa3_ManAlloc(pPars *Bmc_EsPar_t, pTruth *word) *Exa3_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Exa3_Man_t  = &Exa3_Man_t{};
    p.pPars      = pPars;
    p.nVars      = pPars.nVars;
    p.nNodes     = pPars.nNodes;
    p.nLutSize   = pPars.nLutSize;
    p.LutMask    = (1 << pPars.nLutSize) - 1;
    p.nObjs      = pPars.nVars + pPars.nNodes;
    p.nWords     = Abc_TtWordNum(pPars.nVars);
    p.pTruth     = pTruth;
    p.vOutLits   = Vec_WecStart( p.nObjs );
    p.iVar       = Exa3_ManMarkup( p );
    p.vInfo      = Exa3_ManTruthTables( p );
    if p.pPars.nLutSize == 2 {
    p.vUsed2     = Vec_BitStart( (1 << p.pPars.nVars) * p.pPars.nNodes * p.nObjs * p.nObjs );
    } else if p.pPars.nLutSize == 3 {
    p.vUsed3     = Vec_BitStart( (1 << p.pPars.nVars) * p.pPars.nNodes * p.nObjs * p.nObjs * p.nObjs ); }
    p.pSat       = bmcg_sat_solver_start();
    bmcg_sat_solver_set_nvars( p.pSat, p.iVar );
    if pPars.RuntimeLim {
        bmcg_sat_solver_set_runtime_limit( p.pSat, Abc_Clock() + pPars.RuntimeLim * int64(time.Second) ); }
    return p;
}
func Exa3_ManFree(p *Exa3_Man_t) {
    bmcg_sat_solver_stop( p.pSat );
    Vec_WrdFree( p.vInfo );
    Vec_BitFreeP( &p.vUsed2 );
    Vec_BitFreeP( &p.vUsed3 );
    Vec_WecFree( p.vOutLits );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Exa3_ManFindFanin(p *Exa3_Man_t, i int32, k int32) int32 {
    var j int32;
    var Count int32;
    var iVar int32 = -1;
    for  j = 0; j < p.nObjs; j++  {
        if p.VarMarks[i][k][j] && bmcg_sat_solver_read_cex_varvalue(p.pSat, p.VarMarks[i][k][j]) {
            iVar = j;
            Count++;
        }
    }
    assert(Count == 1, "Count == 1");
    return iVar;
}
var spStatic_Exa3_ManEval_Flag int32;
func Exa3_ManEval(p *Exa3_Man_t) int32 {
    
    var i, k, j, iMint int32; var pFanins [6]*word;
    for  i = p.nVars; i < p.nObjs; i++  {
        var iVarStart int32  = 1 + p.LutMask*(i - p.nVars);
        for  k = 0; k < p.nLutSize; k++  {
            pFanins[k] = Exa3_ManTruth( p, Exa3_ManFindFanin(p, i, k) );
        }
        Abc_TtConst0( Exa3_ManTruth(p, i),        p.nWords );
        for  k = 1; k <= p.LutMask; k++  {
            if !bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart+k-1) {
                continue; }
//            Abc_TtAndCompl( Exa3_ManTruth(p, p->nObjs), pFanins[0], !(k&1), pFanins[1], !(k>>1), p->nWords );
            Abc_TtConst1( Exa3_ManTruth(p, p.nObjs), p.nWords );
            for  j = 0; j < p.nLutSize; j++  {
                Abc_TtAndCompl( Exa3_ManTruth(p, p.nObjs), Exa3_ManTruth(p, p.nObjs), 0, pFanins[j], !((k >> j) & 1), p.nWords );
            }
            Abc_TtOr( Exa3_ManTruth(p, i), Exa3_ManTruth(p, i), Exa3_ManTruth(p, p.nObjs), p.nWords );
        }
    }
    if (spStatic_Exa3_ManEval_Flag != 0) && p.nVars >= 6 {
        iMint = Abc_TtFindLastDiffBit( Exa3_ManTruth(p, p.nObjs-1), p.pTruth, p.nVars );
    } else {
        iMint = Abc_TtFindFirstDiffBit( Exa3_ManTruth(p, p.nObjs-1), p.pTruth, p.nVars ); }
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
func Exa3_ManPrintSolution(p *Exa3_Man_t, fCompl int32) {
    var i, k, iVar int32;
    fmt.Printf("Realization of given %d-input function using %d %d-input LUTs:\n", p.nVars, p.nNodes, p.nLutSize);
    for  i = p.nObjs - 1; i >= p.nVars; i--  {
        var Val int32;
        var iVarStart int32 = 1 + p.LutMask*(i - p.nVars);
        fmt.Printf("%02d = %d\'b", i, 1 << p.nLutSize);
        for  k = p.LutMask - 1; k >= 0; k--  {
            Val = bmcg_sat_solver_read_cex_varvalue(p.pSat, iVarStart+k); 
            if i == p.nObjs - 1 && (fCompl != 0) {
                fmt.Printf("%d", !Val);
            } else {
                fmt.Printf("%d", Val); }
        }
        if i == p.nObjs - 1 && (fCompl != 0) {
            fmt.Printf("1(");
        } else {
            fmt.Printf("0("); }

        for  k = p.nLutSize - 1; k >= 0; k--  {
            iVar = Exa3_ManFindFanin( p, i, k );
            if iVar >= 0 && iVar < p.nVars {
                fmt.Printf(" %c", 'a'+iVar);
            } else {
                fmt.Printf(" %02d", iVar); }
        }
        fmt.Printf(" )\n");
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Exa3_ManAddCnfStart(p *Exa3_Man_t, fOnlyAnd int32) int32 {
    var pLits [MAJ_NOBJS]int32;
    var pLits2 [2]int32;
    var i int32;
    var j int32;
    var k int32;
    var n int32;
    var m int32;
    // input constraints
    for  i = p.nVars; i < p.nObjs; i++  {
        var iVarStart int32  = 1 + p.LutMask*(i - p.nVars);
        for  k = 0; k < p.nLutSize; k++  {
            var nLits int32;
            for  j = 0; j < p.nObjs; j++  {
                if p.VarMarks[i][k][j] {
                    pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k][j], 0 ); }
            }
            assert(nLits > 0, "nLits > 0");
            // input uniqueness
            if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                return 0; }
            for  n = 0;   n < nLits; n++  {
            for  m = n+1; m < nLits; m++  {
                pLits2[0] = Abc_LitNot(pLits[n]);
                pLits2[1] = Abc_LitNot(pLits[m]);
                if !bmcg_sat_solver_addclause( p.pSat, pLits2, 2 ) {
                    return 0; }
            }
            }
            if k == p.nLutSize - 1 {
                break; }
            // symmetry breaking
            for  j = 0; j < p.nObjs; j++  { if p.VarMarks[i][k][j] {
            for  n = j; n < p.nObjs; n++  { if p.VarMarks[i][k+1][n] {
                pLits2[0] = Abc_Var2Lit( p.VarMarks[i][k][j], 1 );
                pLits2[1] = Abc_Var2Lit( p.VarMarks[i][k+1][n], 1 );
                if !bmcg_sat_solver_addclause( p.pSat, pLits2, 2 ) {
                    return 0; }
            }
            } }
            }
        }
//#ifdef USE_NODE_ORDER
        // node ordering
        if p.pPars.fUseIncr {
            for  j = p.nVars; j < i; j++  {
            for  n = 0;   n < p.nObjs; n++  { if p.VarMarks[i][0][n] {
            for  m = n+1; m < p.nObjs; m++  { if p.VarMarks[j][0][m] {
                pLits2[0] = Abc_Var2Lit( p.VarMarks[i][0][n], 1 );
                pLits2[1] = Abc_Var2Lit( p.VarMarks[j][0][m], 1 );
                if !bmcg_sat_solver_addclause( p.pSat, pLits2, 2 ) {
                    return 0; }
            }
            } }
            }
            }
        }
//#endif
        if p.nLutSize != 2 {
            continue; }
        // two-input functions
        for  k = 0; k < 3; k++  {
            pLits[0] = Abc_Var2Lit( iVarStart,   k==1 );
            pLits[1] = Abc_Var2Lit( iVarStart+1, k==2 );
            pLits[2] = Abc_Var2Lit( iVarStart+2, k!=0 );
            if !bmcg_sat_solver_addclause( p.pSat, pLits, 3 ) {
                return 0; }
        }
        if fOnlyAnd != 0 {
            pLits[0] = Abc_Var2Lit( iVarStart,   1 );
            pLits[1] = Abc_Var2Lit( iVarStart+1, 1 );
            pLits[2] = Abc_Var2Lit( iVarStart+2, 0 );
            if !bmcg_sat_solver_addclause( p.pSat, pLits, 3 ) {
                return 0; }
        }
    }
    // outputs should be used
    for  i = 0; i < p.nObjs - 1; i++  {
        var vArray *Vec_Int_t  = Vec_WecEntry(p.vOutLits, i);
        assert(Vec_IntSize(vArray) > 0, "Vec_IntSize(vArray) > 0");
        if !bmcg_sat_solver_addclause( p.pSat, Vec_IntArray(vArray), Vec_IntSize(vArray) ) {
            return 0; }
    }
    return 1;
}
func Exa3_ManAddCnf(p *Exa3_Man_t, iMint int32) int32 {
    // save minterm values
    var i int32;
    var k int32;
    var n int32;
    var j int32;
    var Value int32 = Abc_TtGetBit(p.pTruth, iMint);
    for  i = 0; i < p.nVars; i++  {
        p.VarVals[i] = (iMint >> i) & 1;
    }
//    sat_solver_setnvars( p->pSat, p->iVar + (p->nLutSize+1)*p->nNodes );
    bmcg_sat_solver_set_nvars( p.pSat, p.iVar + (p.nLutSize+1)*p.nNodes );
    //printf( "Adding clauses for minterm %d with value %d.\n", iMint, Value );
    for  i = p.nVars; i < p.nObjs; i++  {
        // fanin connectivity
        var iVarStart int32  = 1 + p.LutMask*(i - p.nVars);
        var iBaseSatVarI int32  = p.iVar + (p.nLutSize+1)*(i - p.nVars);
        for  k = 0; k < p.nLutSize; k++  {
            for  j = 0; j < p.nObjs; j++  { if p.VarMarks[i][k][j] {
                var iBaseSatVarJ int32  = p.iVar + (p.nLutSize+1)*(j - p.nVars);
                for  n = 0; n < 2; n++  {
                    var pLits [3]int32;
                    var nLits int32;
                    pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][k][j], 1 );
                    pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + k, n );
                    if j >= p.nVars {
                        pLits[nLits++] = Abc_Var2Lit( iBaseSatVarJ + p.nLutSize, !n );
                    } else if p.VarVals[j] == n {
                        continue; }
                    if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                        return 0; }
                }
            }
            }
        }
        // node functionality
        for  n = 0; n < 2; n++  {
            if i == p.nObjs - 1 && n == Value {
                continue; }
            for  k = 0; k <= p.LutMask; k++  {
                var pLits [16]int32;
                var nLits int32;
                if k == 0 && n == 1 {
                    continue; }
                //pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 0, (k&1)  );
                //pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 1, (k>>1) );
                //if ( i != p->nObjs - 1 ) pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + 2, !n );
                for  j = 0; j < p.nLutSize; j++  {
                    pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + j, (k >> j) & 1 );
                }
                if i != p.nObjs - 1 { pLits[nLits++] = Abc_Var2Lit( iBaseSatVarI + j, !n ); }
                if k > 0 {             pLits[nLits++] = Abc_Var2Lit( iVarStart +  k-1,  n ); }
                assert(nLits <= p.nLutSize+2, "nLits <= p.nLutSize+2");
                if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                    return 0; }
            }
        }
    }
    p.iVar += (p.nLutSize+1)*p.nNodes;
    return 1;
}

func Exa3_ManAddCnf2(p *Exa3_Man_t, iMint int32) int32 {
    var iBaseSatVar int32  = p.iVar + p.nNodes*iMint;
    // save minterm values
    var i int32;
    var k int32;
    var n int32;
    var j int32;
    var Value int32 = Abc_TtGetBit(p.pTruth, iMint);
    for  i = 0; i < p.nVars; i++  {
        p.VarVals[i] = (iMint >> i) & 1;
    }
    //printf( "Adding clauses for minterm %d with value %d.\n", iMint, Value );
    for  i = p.nVars; i < p.nObjs; i++  {
//        int iBaseSatVarI = p->iVar + (p->nLutSize+1)*(i - p->nVars);
        var iVarStart int32  = 1 + p.LutMask*(i - p.nVars);
        // collect fanin variables
        var pFanins [16]int32;
        for  j = 0; j < p.nLutSize; j++  {
            pFanins[j] = Exa3_ManFindFanin(p, i, j);
        }
        // check cache
        if p.pPars.nLutSize == 2 && (Exa3_ManIsUsed2(p, iMint, i, pFanins[1], pFanins[0]) != 0) {
            continue; }
        if p.pPars.nLutSize == 3 && (Exa3_ManIsUsed3(p, iMint, i, pFanins[2], pFanins[1], pFanins[0]) != 0) {
            continue; }
        // node functionality
        for  n = 0; n < 2; n++  {
            if i == p.nObjs - 1 && n == Value {
                continue; }
            for  k = 0; k <= p.LutMask; k++  {
                var pLits [16]int32;
                var nLits int32;
                if k == 0 && n == 1 {
                    continue; }
                for  j = 0; j < p.nLutSize; j++  {
//                    pLits[nLits++] = Abc_Var2Lit( iBaseSatVar + j, (k >> j) & 1 );
                    pLits[nLits++] = Abc_Var2Lit( p.VarMarks[i][j][pFanins[j]], 1 );
                    if pFanins[j] >= p.nVars {
                        pLits[nLits++] = Abc_Var2Lit( iBaseSatVar + pFanins[j] - p.nVars, (k >> j) & 1 );
                    } else if p.VarVals[pFanins[j]] != ((k >> j) & 1) {
                        break; }
                }
                if j < p.nLutSize {
                    continue; }
//                if ( i != p->nObjs - 1 ) pLits[nLits++] = Abc_Var2Lit( iBaseSatVar + j, !n );
                if i != p.nObjs - 1 { pLits[nLits++] = Abc_Var2Lit( iBaseSatVar + i - p.nVars, !n ); }
                if k > 0 {             pLits[nLits++] = Abc_Var2Lit( iVarStart + k-1,             n ); }
                assert(nLits <= 2*p.nLutSize+2, "nLits <= 2*p.nLutSize+2");
                if !bmcg_sat_solver_addclause( p.pSat, pLits, nLits ) {
                    return 0; }
            }
        }
    }
    return 1;
}
func Exa3_ManPrint(p *Exa3_Man_t, i int32, iMint int32, clk abctime) {
    fmt.Printf("Iter%6d : ", i);
    Extra_PrintBinary( os.Stdout, (unsigned *)&iMint, p.nVars );
    fmt.Printf("  Var =%5d  ", bmcg_sat_solver_varnum(p.pSat));
    fmt.Printf("Cla =%6d  ", bmcg_sat_solver_clausenum(p.pSat));
    fmt.Printf("Conf =%9d  ", bmcg_sat_solver_conflictnum(p.pSat));
    Abc_PrintTime( 1, "Time", clk );
}
func Exa3_ManExactSynthesis(pPars *Bmc_EsPar_t) {
    var i int32;
    var status int32;
    var iMint int32 = 1;
    var clkTotal abctime  = Abc_Clock();
    var p *Exa3_Man_t; var fCompl int32;
    var pTruth [16]word; Abc_TtReadHex( pTruth, pPars.pTtStr );
    assert(pPars.nVars <= 10, "pPars.nVars <= 10");
    assert(pPars.nLutSize <= 6, "pPars.nLutSize <= 6");
    p = Exa3_ManAlloc( pPars, pTruth );
    if pTruth[0] & 1 { fCompl = 1; Abc_TtNot( pTruth, p.nWords ); }
    status = Exa3_ManAddCnfStart( p, pPars.fOnlyAnd );
    assert((status != 0), "status");
    fmt.Printf("Running exact synthesis for %d-input function with %d %d-input LUTs...\n", p.nVars, p.nNodes, p.nLutSize);
    if pPars.fUseIncr {
        bmcg_sat_solver_set_nvars( p.pSat, p.iVar + p.nNodes*(1 << p.nVars) );
        status = bmcg_sat_solver_solve( p.pSat, nil, 0 );
        assert(status == GLUCOSE_SAT, "status == GLUCOSE_SAT");
    }
    for  i = 0; iMint != -1; i++  {
        if pPars.fUseIncr ? !Exa3_ManAddCnf2( p, iMint ) : !Exa3_ManAddCnf( p, iMint ) {
            break; }
        status = bmcg_sat_solver_solve( p.pSat, nil, 0 );
        if pPars.fVerbose && (!pPars.fUseIncr || i % 100 == 0) {
            Exa3_ManPrint( p, i, iMint, Abc_Clock() - clkTotal ); }
        if status == GLUCOSE_UNSAT || status == GLUCOSE_UNDEC {
            break; }
        iMint = Exa3_ManEval( p );
    }
    if pPars.fVerbose && status != GLUCOSE_UNDEC {
        Exa3_ManPrint( p, i, iMint, Abc_Clock() - clkTotal ); }
    if iMint == -1 {
        Exa3_ManPrintSolution( p, fCompl );
    } else if status == GLUCOSE_UNDEC {
        fmt.Printf("The problem timed out after %d sec.\n", pPars.RuntimeLim);
    } else { 
        fmt.Printf("The problem has no solution.\n"); }
    fmt.Printf("Added = %d.  Tried = %d.  ", p.nUsed[1], p.nUsed[0]);
    Exa3_ManFree( p );
    Abc_PrintTime( 1, "Total runtime", Abc_Clock() - clkTotal );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

