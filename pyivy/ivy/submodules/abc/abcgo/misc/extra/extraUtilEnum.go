package extra

import "fmt"

/**CFile****************************************************************

  FileName    [extraUtilEnum.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Function enumeration.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilEnum.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
//#include "misc/vec/vec.h"
//#include "misc/vec/vecHsh.h"
//#include "bool/kit/kit.h"

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
func Abc_GetFirst(pnVars *int32, pnMints *int32, pnFuncs *int32, pVars *uint32, pMints *uint32, pFuncs *uint32) {
    var nVars int32   = 8;
    var nMints int32  = 16;
    var nFuncs int32  = 8;
    var pMintStrs [16]string = [16]string{ 
        "1-1-1-1-",
        "1-1--11-",
        "1-1-1--1",
        "1-1--1-1",

        "-11-1-1-",
        "-11--11-",
        "-11-1--1",
        "-11--1-1",

        "1--11-1-",
        "1--1-11-",
        "1--11--1",
        "1--1-1-1",

        "-1-11-1-",
        "-1-1-11-",
        "-1-11--1",
        "-1-1-1-1",
    };
    var pFuncStrs [8]string = [8]string{ 
        "1111101011111010",
        "0000010100000101",
        "1111110010101001",
        "0000001101010110",
        "1111111111001101",
        "0000000000110010",
        "1111111111111110",
        "0000000000000001",
    };
    var i, k int32;
    *pnVars  = nVars;
    *pnMints = nMints;
    *pnFuncs = nFuncs;
    // extract mints
    for  i = 0; i < nMints; i++  {
        for  k = 0; k < nVars; k++  {
            if pMintStrs[i][k] == '1' {
                pMints[i] |= (1 << k), pVars[k] |= (1 << i); }
        }
    }
    // extract funcs
    for  i = 0; i < nFuncs; i++  {
        for  k = 0; k < nMints; k++  {
            if pFuncStrs[i][k] == '1' {
                pFuncs[i] |= (1 << k); }
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_GetSecond(pnVars *int32, pnMints *int32, pnFuncs *int32, pVars *uint32, pMints *uint32, pFuncs *uint32) {
    var nVars int32   = 10;
    var nMints int32  = 32;
    var nFuncs int32  = 7;
    var pMintStrs [32]string = [32]string{ 
        "1-1---1---",
        "1-1----1--",
        "1-1-----1-",
        "1-1------1",

        "1--1--1---",
        "1--1---1--",
        "1--1----1-",
        "1--1-----1",

        "1---1-1---",
        "1---1--1--",
        "1---1---1-",
        "1---1----1",

        "1----11---",
        "1----1-1--",
        "1----1--1-",
        "1----1---1",


        "-11---1---",
        "-11----1--",
        "-11-----1-",
        "-11------1",

        "-1-1--1---",
        "-1-1---1--",
        "-1-1----1-",
        "-1-1-----1",

        "-1--1-1---",
        "-1--1--1--",
        "-1--1---1-",
        "-1--1----1",

        "-1---11---",
        "-1---1-1--",
        "-1---1--1-",
        "-1---1---1",
    };
    var pFuncStrs [7]string = [7]string{ 
        "11111110110010001110110010000000",
        "00000001001101110001001101111111",
        "10000001001001000001001001001000",
        "01001000000100101000000100100100",
        "00100100100000010100100000010010",
        "00010010010010000010010010000001",
        "11111111111111111111000000000000",
    };
    var i, k int32;
    *pnVars  = nVars;
    *pnMints = nMints;
    *pnFuncs = nFuncs;
    // extract mints
    for  i = 0; i < nMints; i++  {
        for  k = 0; k < nVars; k++  {
            if pMintStrs[i][k] == '1' {
                pMints[i] |= (1 << k), pVars[k] |= (1 << i); }
        }
    }
    // extract funcs
    for  i = 0; i < nFuncs; i++  {
        for  k = 0; k < nMints; k++  {
            if pFuncStrs[i][k] == '1' {
                pFuncs[i] |= (1 << k); }
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_GetThird(pnVars *int32, pnMints *int32, pnFuncs *int32, pVars *uint32, pMints *uint32, pFuncs *uint32) {
    var nVars int32   = 8;
    var nMints int32  = 16;
    var nFuncs int32  = 7;
    var pMintStrs [16]string = [16]string{ 
        "1---1---",
        "1----1--",
        "1-----1-",
        "1------1",

        "-1--1---",
        "-1---1--",
        "-1----1-",
        "-1-----1",

        "--1-1---",
        "--1--1--",
        "--1---1-",
        "--1----1",

        "---11---",
        "---1-1--",
        "---1--1-",
        "---1---1",
    };
    var pFuncStrs [7]string = [7]string{ 
        "1111111011001000",
        "0000000100110111",
        "1000000100100100",
        "0100100000010010",
        "0010010010000001",
        "0001001001001000",
        "1111111111111111",
    };
    var i, k int32;
    *pnVars  = nVars;
    *pnMints = nMints;
    *pnFuncs = nFuncs;
    // extract mints
    for  i = 0; i < nMints; i++  {
        for  k = 0; k < nVars; k++  {
            if pMintStrs[i][k] == '1' {
                pMints[i] |= (1 << k), pVars[k] |= (1 << i); }
        }
    }
    // extract funcs
    for  i = 0; i < nFuncs; i++  {
        for  k = 0; k < nMints; k++  {
            if pFuncStrs[i][k] == '1' {
                pFuncs[i] |= (1 << k); }
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_EnumPrint_rec(vGates *Vec_Int_t, i int32, nVars int32) {
    var Fan0 int32  = Vec_IntEntry(vGates, 2*i);
    var Fan1 int32  = Vec_IntEntry(vGates, 2*i+1);
    var pOper string  = (char*)(Fan0 < Fan1 ? "" : "+");
    if Fan0 > Fan1 {
        ABC_SWAP( int, Fan0, Fan1 ); }
    if Fan0 < nVars {
        fmt.Printf("%c", 'a'+Fan0);
    } else {
        fmt.Printf("(");
        Abc_EnumPrint_rec( vGates, Fan0, nVars );
        fmt.Printf(")");
    }
    fmt.Printf("%s", pOper);
    if Fan1 < nVars {
        fmt.Printf("%c", 'a'+Fan1);
    } else {
        fmt.Printf("(");
        Abc_EnumPrint_rec( vGates, Fan1, nVars );
        fmt.Printf(")");
    }
}
func Abc_EnumPrint(vGates *Vec_Int_t, i int32, nVars int32) {
    assert(2*i < Vec_IntSize(vGates), "2*i < Vec_IntSize(vGates)");
    Abc_EnumPrint_rec( vGates, i, nVars );
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DataHasBit(p *word, i word) int32 { return (p[(i)>>6] & (((word)1)<<((i) & 63))) > 0; }
func Abc_DataXorBit(p *word, i word) { p[(i)>>6] ^= (((word)1)<<((i) & 63));             }

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_EnumerateFunctions(nDecMax int32) {
    var nVars int32;
    var nMints int32;
    var nFuncs int32;
    var pVars [100]uint32 ;
    var pMints [100]uint32 ;
    var pFuncs [100]uint32 ;
    var Truth uint32;
    var FuncDone [100]int32 ;
    var nFuncDone int32;
    var GateCount [100]int32 ;
    var i, k, n, a, b, v int32;
    var clk abctime  = Abc_Clock();
    var vGates *Vec_Int_t  = Vec_IntAlloc( 100000 );
    var vTruths *Vec_Int_t  = Vec_IntAlloc( 100000 );
//    Vec_Int_t * vHash = Vec_IntStartFull( 1 << 16 );
    var pHash *word;

    // extract data
//    Abc_GetFirst( &nVars, &nMints, &nFuncs, pVars, pMints, pFuncs );
    Abc_GetSecond( &nVars, &nMints, &nFuncs, pVars, pMints, pFuncs );
//    Abc_GetThird( &nVars, &nMints, &nFuncs, pVars, pMints, pFuncs );

    // create hash table
    assert(nMints == 16 || nMints == 32, "nMints == 16 || nMints == 32");
    pHash = (word *)ABC_CALLOC( char, 1 << (nMints-3) );

    // create elementary gates
    for  k = 0; k < nVars; k++  {
//        Vec_IntWriteEntry( vHash, pVars[k], k );
        Abc_DataXorBit( pHash, pVars[k] );
        Vec_IntPush( vTruths, pVars[k] );
        Vec_IntPush( vGates, -1 );
        Vec_IntPush( vGates, -1 );
    }

    // go through different number of variables
    GateCount[0] = 0;
    GateCount[1] = nVars;
    assert(Vec_IntSize(vTruths) == nVars, "Vec_IntSize(vTruths) == nVars");
    for  n = 0; n < nDecMax && nFuncDone < nFuncs; n++  {
        for  a = 0; a <= n; a++  {
        for  b = a; b <= n; b++  {
        if a + b == n {
            fmt.Printf("Trying %d + %d + 1 = %d\n", a, b, n+1);
            for  i = GateCount[a]; i < GateCount[a+1]; i++  {
            for  k = GateCount[b]; k < GateCount[b+1]; k++  {
            if i < k {
                Truth = Vec_IntEntry(vTruths, i) & Vec_IntEntry(vTruths, k);
//                if ( Vec_IntEntry(vHash, Truth) == -1 )
                if !(Abc_DataHasBit(pHash, Truth) != 0) {
//                    Vec_IntWriteEntry( vHash, Truth, Vec_IntSize(vTruths) );
                    Abc_DataXorBit( pHash, Truth );
                    Vec_IntPush( vTruths, Truth );
                    Vec_IntPush( vGates, i );
                    Vec_IntPush( vGates, k );

                    for  v = 0; v < nFuncs; v++  {
                    if !FuncDone[v] && Truth == pFuncs[v] {
                        fmt.Printf("Found function %d with %d gates: ", v, n+1);
                        Abc_EnumPrint( vGates, Vec_IntSize(vTruths)-1, nVars );
                        FuncDone[v] = 1;
                        nFuncDone++;
                    }
                    }
                }
                Truth = Vec_IntEntry(vTruths, i) | Vec_IntEntry(vTruths, k);
//                if ( Vec_IntEntry(vHash, Truth) == -1 )
                if !(Abc_DataHasBit(pHash, Truth) != 0) {
//                    Vec_IntWriteEntry( vHash, Truth, Vec_IntSize(vTruths) );
                    Abc_DataXorBit( pHash, Truth );
                    Vec_IntPush( vTruths, Truth );
                    Vec_IntPush( vGates, k );
                    Vec_IntPush( vGates, i );

                    for  v = 0; v < nFuncs; v++  {
                    if !FuncDone[v] && Truth == pFuncs[v] {
                        fmt.Printf("Found function %d with %d gates: ", v, n+1);
                        Abc_EnumPrint( vGates, Vec_IntSize(vTruths)-1, nVars );
                        FuncDone[v] = 1;
                        nFuncDone++;
                    }
                    }
                }
            }
            }
            }
        }
        }
        }
        GateCount[n+2] = Vec_IntSize(vTruths);
        fmt.Printf("Finished %d gates.  Truths = %10d.  ", n+1, Vec_IntSize(vTruths));
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    ABC_FREE( pHash );
//    Vec_IntFree( vHash );
    Vec_IntFree( vGates );
    Vec_IntFree( vTruths );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
const ABC_ENUM_MAX = 16
var s_Truths6 [6]word = [6]word{
    ABC_CONST(0xAAAAAAAAAAAAAAAA),
    ABC_CONST(0xCCCCCCCCCCCCCCCC),
    ABC_CONST(0xF0F0F0F0F0F0F0F0),
    ABC_CONST(0xFF00FF00FF00FF00),
    ABC_CONST(0xFFFF0000FFFF0000),
    ABC_CONST(0xFFFFFFFF00000000),
};
   
type Abc_EnuMan_t struct {
    nVars              int32;        // support size
    nVarsFree              int32;    // number of PIs used
    fVerbose              int32;     // verbose flag
    fUseXor              int32;      // using XOR gate
    nNodeMax              int32;     // the max number of nodes
    nNodes              int32;       // current number of gates
    nTops              int32;        // the number of fanoutless gates
    pFans0 [ABC_ENUM_MAX]int32;   // fanins
    pFans1 [ABC_ENUM_MAX]int32;   // fanins 
    fCompl0 [ABC_ENUM_MAX]int32;  // complements
    fCompl1 [ABC_ENUM_MAX]int32;  // complements
    Polar [ABC_ENUM_MAX]int32;    // polarity
    pRefs [ABC_ENUM_MAX]int32;    // references
    pLevel [ABC_ENUM_MAX]int32;   // level 
    word             pTruths[ABC_ENUM_MAX];  // truth tables
    nTries             word;       // attempts to build a gate
    nBuilds             word;      // actually built gates
    nFinished             word;    // finished structures
}
func Abc_EnumRef(p *Abc_EnuMan_t, i int32) {
    assert(p.pRefs[i] >= 0, "p.pRefs[i] >= 0");
    if p.pRefs[i]++ == 0 {
        p.nTops--; }
}
func Abc_EnumDeref(p *Abc_EnuMan_t, i int32) {
    if --p.pRefs[i] == 0 {
        p.nTops++; }
    assert(p.pRefs[i] >= 0, "p.pRefs[i] >= 0");
}
func Abc_EnumRefNode(p *Abc_EnuMan_t, i int32) {
    Abc_EnumRef( p, p.pFans0[i] );
    Abc_EnumRef( p, p.pFans1[i] );
    p.nTops++;
    p.nNodes++; 
    assert(i < p.nNodes, "i < p.nNodes");
}
func Abc_EnumDerefNode(p *Abc_EnuMan_t, i int32) {
    assert(i < p.nNodes, "i < p.nNodes");
    Abc_EnumDeref( p, p.pFans0[i] );
    Abc_EnumDeref( p, p.pFans1[i] );
    p.nTops--;
    p.nNodes--; 
}
func Abc_EnumPrintOne(p *Abc_EnuMan_t) {
    var i int32;
    Kit_DsdPrintFromTruth( (unsigned *)(p.pTruths + p.nNodes - 1), p.nVars ); 
    for  i = p.nVars; i < p.nNodes; i++  {
        if p.Polar[i] == 4 {
            fmt.Printf("  %c=%c+%c", 'a'+i, 'a'+p.pFans0[i], 'a'+p.pFans1[i]);
        } else {
            fmt.Printf("  %c=%s%c%s%c", 'a'+i, p.fCompl0[i]?"!":"", 'a'+p.pFans0[i], p.fCompl1[i]?"!":"", 'a'+p.pFans1[i]); }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_EnumEquiv(a word, b word) int32 {
    return a == b || a == ~b;
}
func Abc_EnumerateFilter(p *Abc_EnuMan_t) int32 {
    var fUseFull int32  = 1;
    var n int32  = p.nNodes;
    var i int32  = p.pFans0[n];
    var k int32 = p.pFans1[n];
    var t int32;
    var pTruths *word  = p.pTruths;
    var uTruth word  = pTruths[n];
    assert(i < k, "i < k");
    // skip constants
    if Abc_EnumEquiv(uTruth, 0) != 0 {
        return 1; }
    // skip equal ones
    for  t = 0; t < n; t++  {
        if Abc_EnumEquiv(uTruth, pTruths[t]) != 0 {
            return 1; }
    }
    if fUseFull != 0 {
        // skip those that can be derived by any pair
        var a, b int32; 
        for  a = 0; a <= i; a++  {
        for  b = a + 1; b <= k; b++  {
            if a == i && b == k {
                continue; }
            if Abc_EnumEquiv(uTruth,  pTruths[a] &  pTruths[b]) != 0 {
                return 1; }
            if Abc_EnumEquiv(uTruth,  pTruths[a] & ~pTruths[b]) != 0 {
                return 1; }
            if Abc_EnumEquiv(uTruth, ~pTruths[a] &  pTruths[b]) != 0 {
                return 1; }
            if Abc_EnumEquiv(uTruth, ~pTruths[a] & ~pTruths[b]) != 0 {
                return 1; }
            if p.fUseXor && (Abc_EnumEquiv(uTruth,  pTruths[a] ^ pTruths[b]) != 0) {
                return 1; }
        }
        }
    
    } else {
        // skip those that can be derived by fanin and any other one in the cone
        var uTruthI int32;
        if p.fCompl0[n] {
            uTruthI = ~pTruths[i];
        } else {
            uTruthI = pTruths[i];
        }
        var uTruthK int32;
        if p.fCompl1[n] {
            uTruthK = ~pTruths[k];
        } else {
            uTruthK = pTruths[k];
        }
        assert(p.fUseXor == 0, "p.fUseXor == 0");
        for  t = 0; t < k; t++  {
            if (Abc_EnumEquiv(uTruth, pTruths[t] & uTruthI) != 0) || (Abc_EnumEquiv(uTruth, ~pTruths[t] & uTruthI) != 0) {
                return 1; }
        }
        for  t = 0; t < i; t++  {
            if (Abc_EnumEquiv(uTruth, pTruths[t] & uTruthK) != 0) || (Abc_EnumEquiv(uTruth, ~pTruths[t] & uTruthK) != 0) {
                return 1; }
        }
    }
    return 0;
}
void Abc_EnumerateFuncs_rec( Abc_EnuMan_t * p, int fNew, int iNode1st ) // the first node on the last level
{
    if p.nNodes == p.nNodeMax {
        assert(p.nTops == 1, "p.nTops == 1");
        if p.fVerbose {
            Abc_EnumPrintOne( p ); }
        p.nFinished++;
        return;
    }
    {
    var i int32;
    var k int32;
    var c int32;
    var cLim int32 = 4 + p.fUseXor;
    var n int32 = p.nNodes;
    var nRefedFans int32  = p.nNodeMax - n + 1 - p.nTops;
    var high0 int32;
    if fNew {
        high0 = iNode1st;
    } else {
        high0 = p.pFans1[n-1];
    }
    var high1 int32;
    if fNew {
        high1 = n;
    } else {
        high1 = iNode1st;
    }
    var low0 int32;
    if fNew {
        low0 = 0;
    } else {
        low0 = p.pFans0[n-1];
    }
    var c0 int32;
    if fNew {
        c0 = 0;
    } else {
        c0 = p.Polar[n-1];
    }
    var Level int32  = p.pLevel[high0];
    assert(p.nTops > 0 && p.nTops <= p.nNodeMax - n + 1, "p.nTops > 0 && p.nTops <= p.nNodeMax - n + 1");
    // go through nodes 
    for  k = high0; k < high1; k++  {
        if nRefedFans == 0 && p.pRefs[k] > 0 {
            continue; }
        if p.pRefs[k] > 0 {
            nRefedFans--; }
        assert(nRefedFans >= 0, "nRefedFans >= 0");
        // try second fanin
        for  i = (k == high0) ? low0 : 0; i < k; i++  {
            if nRefedFans == 0 && p.pRefs[i] > 0 {
                continue; }
            if Level == 0 && p.pRefs[i] == 0 && p.pRefs[k] == 0 && (i+1 != k || (i > 0 && p.pRefs[i-1] == 0)) { // NPN
                continue; }
            if p.pLevel[k] == 0 && p.pRefs[k] == 0 && p.pRefs[i] != 0 && k > 0 && p.pRefs[k-1] == 0 { // NPN
                continue; }
//            if ( p->pLevel[i] == 0 && p->pRefs[i] == 0 && p->pRefs[k] != 0 && i > 0 && p->pRefs[i-1] == 0 ) // NPN
//                continue;
            // try four polarities
            for  c = (k == high0 && i == low0 && !fNew) ? c0 + 1 : 0; c < cLim; c++  {
                if p.pLevel[i] == 0 && p.pRefs[i] == 0 && (c & 1) == 1 { // NPN
                    continue; }
                if p.pLevel[k] == 0 && p.pRefs[k] == 0 && (c & 2) == 2 { // NPN
                    continue; }
                p.nTries++;
                // create node
                assert(i < k, "i < k");
                p.pFans0[n]  = i;
                p.pFans1[n]  = k;
                p.fCompl0[n] = c & 1;
                p.fCompl1[n] = (c >> 1) & 1;
                p.Polar[n]   = c;
                if c == 4 {
                    p.pTruths[n] = p.pTruths[i] ^ p.pTruths[k];
                } else {
                    p.pTruths[n] = ((c & 1) ? ~p.pTruths[i] : p.pTruths[i]) & ((c & 2) ? ~p.pTruths[k] : p.pTruths[k]); }
                if Abc_EnumerateFilter(p) != 0 {
                    continue; }
                p.nBuilds++;
                assert(Level == Abc_MaxInt(p.pLevel[i], p.pLevel[k]), "Level == Abc_MaxInt(p.pLevel[i], p.pLevel[k])");
                p.pLevel[n]  = Level + 1;
                Abc_EnumRefNode( p, n );
                Abc_EnumerateFuncs_rec( p, 0, fNew ? n : iNode1st );
                Abc_EnumDerefNode( p, n );
                assert(n == p.nNodes, "n == p.nNodes");
            }
        }
        if p.pRefs[k] > 0 {
            nRefedFans++; }
    }
    if fNew {
        return; }
    // start a new level
    Abc_EnumerateFuncs_rec( p, 1, iNode1st );
    }
}
func Abc_EnumerateFuncs(nVars int32, nGates int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var P Abc_EnuMan_t;
    var p *Abc_EnuMan_t = &P; 
    var i int32;
    if nVars > nGates + 1 {
        fmt.Printf("The gate count %d is not enough to have functions with %d inputs.\n", nGates, nVars);
        return;
    }
    assert(nVars >= 2 && nVars <= 6, "nVars >= 2 && nVars <= 6");
    assert(nGates > 0 && nVars + nGates < ABC_ENUM_MAX, "nGates > 0 && nVars + nGates < ABC_ENUM_MAX");
    memset( p, 0, sizeof(Abc_EnuMan_t) );
    p.fVerbose  = fVerbose;
    p.fUseXor   = 0;
    p.nVars     = nVars;
    p.nNodeMax  = nVars + nGates;
    p.nNodes    = nVars;
    p.nTops     = nVars;
    for  i = 0; i < nVars; i++  {
        p.pTruths[i] = s_Truths6[i];
    }
    Abc_EnumerateFuncs_rec( p, 1, 0 );
    assert(p.nNodes == nVars, "p.nNodes == nVars");
    assert(p.nTops == nVars, "p.nTops == nVars");
    // report statistics
    fmt.Printf("Vars = %d.  Gates = %d.  Tries = %d. Builds = %d.  Finished = %d. ", nVars, nGates, uint32(p.nTries), uint32(p.nBuilds), uint32(p.nFinished));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

