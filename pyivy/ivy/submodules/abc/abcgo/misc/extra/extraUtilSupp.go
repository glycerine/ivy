package extra

import (
	"fmt"
	"strconv"
)

/**CFile****************************************************************

  FileName    [extraUtilSupp.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Support minimization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilSupp.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
//#include "misc/vec/vec.h"
//#include "misc/vec/vecWec.h"
//#include "extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern void         Extra_PrintBinary( FILE * pFile, unsigned Sign[], int nBits );



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Generate m-out-of-n vectors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppCountOnes(i uint32) int32 {
    i = i - ((i >> 1) & 0x55555555);
    i = (i & 0x33333333) + ((i >> 2) & 0x33333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F);
    return (i*(0x01010101))>>24;
}
func Abc_SuppGen(m int32, n int32) *Vec_Wrd_t {
    var vRes *Vec_Wrd_t  = Vec_WrdAlloc( 1000 );
    var i int32;
    var Size int32 = (1 << n);
    for  i = 0; i < Size; i++  {
        if Abc_SuppCountOnes(i) == m {
            Vec_WrdPush( vRes, i ); }
    }
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Perform verification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppVerify(p *Vec_Wrd_t, pMatrix *word, nVars int32, nVarsMin int32) {
    var pNew *Vec_Wrd_t;
    var pLimit, pEntry1, pEntry2 *word;
    var Entry, EntryNew word;
    var i, k, v, Value, Counter int32;
    pNew = Vec_WrdAlloc( Vec_WrdSize(p) );
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(p)) && (((Entry) = Vec_WrdEntry(p, i)), 1); i++  {
        EntryNew = 0;
        for  v = 0; v < nVarsMin; v++  {
            Value = 0;
            for  k = 0; k < nVars; k++  {
                if ((pMatrix[v] >> k) & 1) && ((Entry >> k) & 1) {
                    Value ^= 1; }
            }
            if Value != 0 {
                EntryNew |= (((word)1) << v); }            
        }
        Vec_WrdPush( pNew, EntryNew );
    }
    // check that they are disjoint
    pLimit  = Vec_WrdLimit(pNew);
    pEntry1 = Vec_WrdArray(pNew);
    for  ; pEntry1 < pLimit; pEntry1++  {
    for  pEntry2 = pEntry1 + 1; pEntry2 < pLimit; pEntry2++  {
        if *pEntry1 == *pEntry2 {
            Counter++; }
    }
    }
    if Counter != 0 {
        fmt.Printf("The total of %d pairs fail verification.\n", Counter);
    } else {
        fmt.Printf("Verification successful.\n"); }
    Vec_WrdFree( pNew );
}

/**Function*************************************************************

  Synopsis    [Generate pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppGenPairs(p *Vec_Wrd_t, nBits int32) *Vec_Wrd_t {
    var vRes *Vec_Wrd_t  = Vec_WrdAlloc( 1000 );
    // inlining of macro ABC_CALLOC
    var pMap *uint32  = ((unsigned *) calloc((size_t)(1 << Abc_MaxInt(0,nBits-5)), sizeof(unsigned))); 
    var pLimit *word  = Vec_WrdLimit(p);
    var pEntry1 *word  = Vec_WrdArray(p);
    var pEntry2 *word;
    var Value word;
    for  ; pEntry1 < pLimit; pEntry1++  {
    for  pEntry2 = pEntry1 + 1; pEntry2 < pLimit; pEntry2++  {
        Value = *pEntry1 ^ *pEntry2;
        if Abc_InfoHasBit(pMap, int32(Value)) {
            continue; }
        Abc_InfoXorBit( pMap, int32(Value) );
        Vec_WrdPush( vRes, Value );
    }
    }
    pMap = nil;
    return vRes;
}
func Abc_SuppGenPairs2(nOnes int32, nBits int32) *Vec_Wrd_t {
    var vRes *Vec_Wrd_t  = Vec_WrdAlloc( 1000 );
    var i int32;
    var k int32;
    var Size int32 = (1 << nBits);
    var Value int32;
    for  i = 0; i < Size; i++  {
        Value = Abc_SuppCountOnes(i);
        for  k = 1; k <= nOnes; k++  {
            if Value == 2*k {
                break; }
        }
        if k <= nOnes {
            Vec_WrdPush( vRes, i ); }
    }
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Select variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppPrintMask(uMask word, nBits int32) {
    var i int32;
    for  i = 0; i < nBits; i++  {
        fmt.Printf("%d", int32(((uMask >> i) & 1)));
    }
    fmt.Printf("\n");
}
func Abc_SuppGenProfile(p *Vec_Wrd_t, nBits int32, pCounts *int32) {
    var Ent word;
    var i, k, b int32;
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(p)) && (((Ent) = Vec_WrdEntry(p, i)), 1); i++  {
        for  b = ((Ent >> nBits) & 1), k = 0; k < nBits; k++  {
            pCounts[k] += ((Ent >> k) & 1) ^ b;
        }
    }
}
func Abc_SuppPrintProfile(p *Vec_Wrd_t, nBits int32) {
    var k int32;
    var Counts [64]int32 ;
    Abc_SuppGenProfile( p, nBits, Counts );
    for  k = 0; k < nBits; k++  {
        fmt.Printf("%2d : %6d  %6.2f %%\n", k, Counts[k], 100.0 * Counts[k] / Vec_WrdSize(p));
    }
}
func Abc_SuppGenFindBest(p *Vec_Wrd_t, nBits int32, pMerit *int32) int32 {
    var k int32;
    var kBest int32;
    var Counts [64]int32 ;
    Abc_SuppGenProfile( p, nBits, Counts );
    for  k = 1; k < nBits; k++  {
        if Counts[kBest] < Counts[k] {
            kBest = k; }
    }
    *pMerit = Counts[kBest];
    return kBest;
}
func Abc_SuppGenSelectVar(p *Vec_Wrd_t, nBits int32, iVar int32) {
    var pEntry *word  = Vec_WrdArray(p);
    var pLimit *word  = Vec_WrdLimit(p);
    for  ; pEntry < pLimit; pEntry++  {
        if (*pEntry >> iVar) & 1 {
            *pEntry ^= (((word)1) << nBits); }
    }
}
func Abc_SuppGenFilter(p *Vec_Wrd_t, nBits int32) {
    var Ent word;
    var i, k int32;
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(p)) && (((Ent) = Vec_WrdEntry(p, i)), 1); i++  {
        if ((Ent >> nBits) & 1) == 0 {
            Vec_WrdWriteEntry( p, k++, Ent ); }
    }
    Vec_WrdShrink( p, k );
}
func Abc_SuppFindOne(p *Vec_Wrd_t, nBits int32) word {
    var uMask word;
    var Prev int32 = -1;
    var This int32;
    var Var int32;
    for {
        Var = Abc_SuppGenFindBest( p, nBits, &This );
        if Prev >= This {
            break; }
        Prev = This;
        Abc_SuppGenSelectVar( p, nBits, Var );
        uMask |= (((word)1) << Var);
    }
    return uMask;
}
func Abc_SuppMinimize(pMatrix *word, p *Vec_Wrd_t, nBits int32, fVerbose int32) int32 {
    var i int32;
    for  i = 0; Vec_WrdSize(p) > 0; i++  {
//        Abc_SuppPrintProfile( p, nBits );
        pMatrix[i] = Abc_SuppFindOne( p, nBits );
        Abc_SuppGenFilter( p, nBits );   
        if !(fVerbose != 0) {
            continue; }
        // print stats
        fmt.Printf("%2d : ", i);
        fmt.Printf("%6d  ", Vec_WrdSize(p));
        Abc_SuppPrintMask( pMatrix[i], nBits );
//        printf( "\n" );
    }
    return i;
}

/**Function*************************************************************

  Synopsis    [Create representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppTest(nOnes int32, nVars int32, fUseSimple int32, fCheck int32, fVerbose int32) {
    var nVarsMin int32;
    var Matrix [64]word;
    var clk abctime  = Abc_Clock();
    // create the problem
    var vRes *Vec_Wrd_t  = Abc_SuppGen( nOnes, nVars );
    var vPairs *Vec_Wrd_t;
    if fUseSimple != 0 {
        vPairs = Abc_SuppGenPairs2( nOnes, nVars );
    } else {
        vPairs = Abc_SuppGenPairs( vRes, nVars );
    }
    assert(nVars < 100, "nVars < 100");
    fmt.Printf("M = %2d  N = %2d : ", nOnes, nVars);
    fmt.Printf("K = %6d   ", Vec_WrdSize(vRes));
    fmt.Printf("Total = %12.0f  ", 0.5 * Vec_WrdSize(vRes) * (Vec_WrdSize(vRes) - 1));
    fmt.Printf("Distinct = %8d  ", Vec_WrdSize(vPairs));
    Abc_PrintTime( 1, "Reduction time", Abc_Clock() - clk );
    // solve the problem
    clk = Abc_Clock();
    nVarsMin = Abc_SuppMinimize( Matrix, vPairs, nVars, fVerbose );
    fmt.Printf("Solution with %d variables found.  ", nVarsMin);
    Abc_PrintTime( 1, "Covering time", Abc_Clock() - clk );
    if fCheck != 0 {
        Abc_SuppVerify( vRes, Matrix, nVars, nVarsMin ); }
    Vec_WrdFree( vPairs );
    Vec_WrdFree( vRes );
}


/**Function*************************************************************

  Synopsis    [Reads the input part of the cubes specified in MIN format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppReadMin(pFileName string, pnVars *int32) *Vec_Wrd_t {
    var vRes *Vec_Wrd_t; var uCube word;
    var nCubes int32;
    var nVars int32 = -1;
    var iVar int32;
    var pCur string;
    var pToken string;
    var pStart string = "INPUT F-COVER";
    var pStr string  = Extra_FileReadContents( pFileName );
    if pStr == nil { fmt.Printf("Cannot open input file (%s).\n", pFileName); return nil; }
    pCur = strstr( pStr, pStart );
    if pCur == nil { fmt.Printf("Cannot find beginning of cube cover (%s).\n", pStart); return nil; }
    pToken = strtok( pCur + len(pStart), " \t\r\n," );
    nCubes = strconv.Atoi(pToken);
    if nCubes < 1 || nCubes > 1000000 { fmt.Printf("The number of cubes in not in the range [1; 1000000].\n"); return nil; }
    vRes = Vec_WrdAlloc( 1000 );
    uCube = 0; iVar = 0;
    for  (pToken = strtok(nil, " \t\r\n,")) != nil  {
        if len(pToken) > 2 {
            if !strncmp(pToken, "INPUT", len("INPUT")) {
                break; }
            if iVar > 64 { fmt.Printf("The number of inputs (%d) is too high.\n", iVar); Vec_WrdFree(vRes); return nil; }
            if nVars == -1 {
                nVars = iVar;
            } else if nVars != iVar { fmt.Printf("The number of inputs (%d) does not match declaration (%d).\n", nVars, iVar); Vec_WrdFree(vRes); return nil; }
            Vec_WrdPush( vRes, uCube );
            uCube = 0; iVar = 0;
            continue;
        }
        if pToken[1] == '0' && pToken[0] == '1' { // value 1
            uCube |= (((word)1) << iVar);
        } else if pToken[1] != '1' || pToken[0] != '0' { // value 0
             fmt.Printf("Strange literal representation (%s) of cube %d.\n", pToken, nCubes); Vec_WrdFree(vRes); return nil; }
        iVar++;
    }
    pStr = nil;
    if Vec_WrdSize(vRes) != nCubes { fmt.Printf("The number of cubes (%d) does not match declaration (%d).\n", Vec_WrdSize(vRes), nCubes); Vec_WrdFree(vRes); return nil; 
    } else {
        fmt.Printf("Successfully parsed function with %d inputs and %d cubes.\n", nVars, nCubes); }
    *pnVars = nVars;
    return vRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppDiffMatrix(vCubes *Vec_Wrd_t) *Vec_Wrd_t {
    var clk abctime  = Abc_Clock();
    var pEnt2 *word;
    var pEnt *word = Vec_WrdArray( vCubes );
    var pLimit *word  = Vec_WrdLimit( vCubes );
    var vRes *Vec_Wrd_t;
    var vPairs *Vec_Wrd_t = Vec_WrdAlloc( Vec_WrdSize(vCubes) * (Vec_WrdSize(vCubes) - 1) / 2 );
    var pStore *word  = Vec_WrdArray( vPairs );
    for  ; pEnt < pLimit; pEnt++  {
    for  pEnt2 = pEnt+1; pEnt2 < pLimit; pEnt2++  {
        *pStore++ = *pEnt ^ *pEnt2;
    }
    }
    vPairs.nSize = Vec_WrdCap(vPairs);
    assert(pStore == Vec_WrdLimit(vPairs), "pStore == Vec_WrdLimit(vPairs)");
//    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
//    vRes = Vec_WrdUniqifyHash( vPairs, 1 );
    vRes = Vec_WrdDup( vPairs );
    fmt.Printf("Successfully generated diff matrix with %10d rows (%6.2f %%).  ", Vec_WrdSize(vRes), 100.0 * Vec_WrdSize(vRes) / Vec_WrdSize(vPairs));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Vec_WrdFree( vPairs );
    return vRes;    
}

/**Function*************************************************************

  Synopsis    [Solve difference matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppCountOnes64(i word) int32 {
    i = i - ((i >> 1) & 0x5555555555555555);
    i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F0F0F0F0F);
    return (i*(0x0101010101010101))>>56;
}
func Abc_SuppFindVar(pS *Vec_Wec_t, pD *Vec_Wec_t, nVars int32) int32 {
    var v int32;
    var vBest int32 = -1;
    var dBest int32 = -1;
    for  v = 0; v < nVars; v++  {
        if Vec_WecLevelSize(pS, v) {
            continue; }
        if vBest == -1 || dBest > Vec_WecLevelSize(pD, v) {
            vBest = v;
dBest = Vec_WecLevelSize(pD, v); }
    }
    return vBest;
}
func Abc_SuppRemove(p *Vec_Wrd_t, pCounts *int32, pS *Vec_Wec_t, pD *Vec_Wec_t, iVar int32, nVars int32) {
    var Entry word; var i, v int32;
    assert(Vec_WecLevelSize(pS, iVar) == 0, "Vec_WecLevelSize(pS, iVar) == 0");
    Vec_IntClear( Vec_WecEntry(pD, iVar) );
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(p)) && (((Entry) = Vec_WrdEntry(p, i)), 1); i++  {
        if ((Entry >> iVar) & 1) == 0 {
            continue; }
        pCounts[i]--;
        if pCounts[i] == 1 {
            for  v = 0; v < nVars; v++  {
                if (Entry >> v) & 1 {
                    Vec_IntRemove( Vec_WecEntry(pD, v), i );
                    Vec_WecPush( pS, v, i );
                }
            }
        
        } else if pCounts[i] == 2 {
            for  v = 0; v < nVars; v++  {
                if (Entry >> v) & 1 {
                    Vec_WecPush( pD, v, i ); }
            }
        }        
    }
}
func Abc_SuppProfile(pS *Vec_Wec_t, pD *Vec_Wec_t, nVars int32) {
    var v int32;
    for  v = 0; v < nVars; v++  {
        fmt.Printf("%2d : S = %3d  D = %3d\n", v, Vec_WecLevelSize(pS, v), Vec_WecLevelSize(pD, v));
    }
}
func Abc_SuppSolve(p *Vec_Wrd_t, nVars int32) int32 {
    var clk abctime  = Abc_Clock();
    var pNew *Vec_Wrd_t  = Vec_WrdDup( p );
    var vSingles *Vec_Wec_t  = Vec_WecStart( 64 );
    var vDoubles *Vec_Wec_t  = Vec_WecStart( 64 );
    var Entry word; var i int32;
var v int32;
var iVar int32;
var nVarsNew int32 = nVars;
    // inlining of macro ABC_ALLOC
    var pCounts *int32  = ((int *) malloc(sizeof(int) * (size_t)(Vec_WrdSize(p))));
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(p)) && (((Entry) = Vec_WrdEntry(p, i)), 1); i++  {
        pCounts[i] = Abc_SuppCountOnes64( Entry );
        if pCounts[i] == 1 {
            for  v = 0; v < nVars; v++  {
                if (Entry >> v) & 1 {
                    Vec_WecPush( vSingles, v, i ); }
            }
        
        } else if pCounts[i] == 2 {
            for  v = 0; v < nVars; v++  {
                if (Entry >> v) & 1 {
                    Vec_WecPush( vDoubles, v, i ); }
            }
        }
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
//    Abc_SuppProfile( vSingles, vDoubles, nVars );
    // find variable in 0 singles and the smallest doubles
    for {
        iVar = Abc_SuppFindVar( vSingles, vDoubles, nVars );
        if iVar == -1 {
            break; }
//        printf( "Selected variable %d.\n", iVar );
        Abc_SuppRemove( pNew, pCounts, vSingles, vDoubles, iVar, nVars );
//        Abc_SuppProfile( vSingles, vDoubles, nVars );
        nVarsNew--;
    }
//    printf( "Result = %d (out of %d)\n", nVarsNew, nVars );
    Vec_WecFree( vSingles );
    Vec_WecFree( vDoubles );
    Vec_WrdFree( pNew );
    pCounts = nil;
    return nVarsNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SuppReadMinTest(pFileName string) {
//    int fVerbose = 0;
    var clk abctime  = Abc_Clock();
//    word Matrix[64];
    var nVars, nVarsMin int32;
    var vPairs, vCubes *Vec_Wrd_t;
    vCubes = Abc_SuppReadMin( pFileName, &nVars );
    if vCubes == nil {
        return; }
    vPairs = Abc_SuppDiffMatrix( vCubes );
    Vec_WrdFreeP( &vCubes );
    // solve the problem
    clk = Abc_Clock();
//    nVarsMin = Abc_SuppMinimize( Matrix, vPairs, nVars, fVerbose );
    nVarsMin = Abc_SuppSolve( vPairs, nVars );
    fmt.Printf("Solution with %d variables found.  ", nVarsMin);
    Abc_PrintTime( 1, "Covering time", Abc_Clock() - clk );

    Vec_WrdFreeP( &vPairs );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

