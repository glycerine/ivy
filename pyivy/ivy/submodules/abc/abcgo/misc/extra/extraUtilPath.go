package extra

import (
	"fmt"
	"math"
)

/**CFile****************************************************************

  FileName    [extraUtilPath.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Path enumeration.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilPath.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
//#include "aig/gia/gia.h"
//#include "misc/vec/vecHsh.h"

//#include <math.h>
//#include "sat/bmc/bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "misc/extra/extra.h"


//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes AIG representing the set of all paths in the graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeVarX(nSize int32, y int32, x int32) int32 {
    return Abc_Var2Lit( nSize * y + x, 0 );
}
func Abc_NodeVarY(nSize int32, y int32, x int32) int32 {
    return Abc_Var2Lit( nSize * (nSize + 1) + nSize * x + y, 0 );
}

func Abc_EnumeratePaths(nSize int32) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    var pGia *Gia_Man_t = Gia_ManStart( 10000 );
    // inlining of macro ABC_CALLOC
    var pNodes *int32  = ((int *) calloc((size_t)(nSize+1), sizeof(int)));
    var x int32;
    var y int32;
    var nVars int32 = 2*nSize*(nSize+1);
    for  x = 0; x < nVars; x++  {
        Gia_ManAppendCi( pGia );
    }
    Gia_ManHashAlloc( pGia );
    // y = 0; x = 0;
    pNodes[0] = 1;
    // y = 0; x > 0 
    for  x = 1; x <= nSize; x++  {
        pNodes[x] = Gia_ManHashAnd( pGia, pNodes[x-1], Abc_NodeVarX(nSize, 0, x) );
    }
    // y > 0; x >= 0
    for  y = 1; y <= nSize; y++  {
        // y > 0; x = 0
        pNodes[0] = Gia_ManHashAnd( pGia, pNodes[0], Abc_NodeVarY(nSize, y, 0) );
        // y > 0; x > 0
        for  x = 1; x <= nSize; x++  {
            var iHor int32   = Gia_ManHashAnd( pGia, pNodes[x-1], Abc_NodeVarX(nSize, y, x) );
            var iVer int32   = Gia_ManHashAnd( pGia, pNodes[x],   Abc_NodeVarY(nSize, y, x) );
            pNodes[x] = Gia_ManHashOr( pGia, iHor, iVer );
        }
    }
    Gia_ManAppendCo( pGia, pNodes[nSize] );
    pGia = Gia_ManCleanup( pTemp = pGia );
    Gia_ManStop( pTemp );
    // inlining of macro ABC_FREE
    ((pNodes) ? (free((char *) (pNodes)), (pNodes) = 0) : 0);
    return pGia;
}
func Abc_EnumeratePathsTest() {
    var nSize int32  = 2;
    var pGia *Gia_Man_t  = Abc_EnumeratePaths( nSize );
    Gia_AigerWrite( pGia, "testpath.aig", 0, 0, 0 );
    Gia_ManStop( pGia );
}


/**Function*************************************************************

  Synopsis    [Generate NxN grid.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_GraphGrid(n int32) *Vec_Int_t {
    var vEdges *Vec_Int_t  = Vec_IntAlloc( 4*n*(n-1) ); // two nodes per edge
    var i, k int32;
    for  i = 0; i < n; i++  {
        for  k = 0; k < n-1; k++  {
            Vec_IntPushTwo( vEdges, i*n+k, i*n+k+1 );
        }
        if i == n-1 { break; }
        for  k = 0; k < n; k++  {
            Vec_IntPushTwo( vEdges, i*n+k, i*n+k+n );
        }
    }
    //Vec_IntPrint( vEdges );
    return vEdges;
}
func Abc_GraphNodeLife(vEdges *Vec_Int_t, n int32) *Vec_Int_t {
    var vLife *Vec_Int_t  = Vec_IntStartFull( 2*n*n ); // start/stop per node
    var One, Two, i int32;
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vEdges)) && (((One) = Vec_IntEntry(vEdges, i)), 1) && (((Two) = Vec_IntEntry(vEdges, i+1)), 1); i += 2  {
        if Vec_IntEntry(vLife, 2*One) == -1 {
            Vec_IntWriteEntry(vLife, 2*One, i/2); }
        if Vec_IntEntry(vLife, 2*Two) == -1 {
            Vec_IntWriteEntry(vLife, 2*Two, i/2); }
        Vec_IntWriteEntry(vLife, 2*One+1, i/2);
        Vec_IntWriteEntry(vLife, 2*Two+1, i/2);
    }
    //Vec_IntPrint( vLife );
    return vLife;
}
func Abc_GraphFrontiers(vEdges *Vec_Int_t, vLife *Vec_Int_t) *Vec_Wec_t {
    var vFronts *Vec_Wec_t  = Vec_WecAlloc( Vec_IntSize(vEdges)/2 ); // front for each edge
    var vTemp *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(vLife)/2 );
    var e, n int32;    
    Vec_WecPushLevel(vFronts);
    for  e = 0; e < Vec_IntSize(vEdges)/2; e++  {
        var pNodes *int32  = Vec_IntEntryP(vEdges, 2*e);
        for  n = 0; n < 2; n++  {
            if Vec_IntEntry(vLife, 2*pNodes[n]) == e { // first time
                Vec_IntPush( vTemp, pNodes[n] );
            } else if Vec_IntEntry(vLife, 2*pNodes[n]+1) == e { // last time
                Vec_IntRemove( vTemp, pNodes[n] ); }
        }
        Vec_IntAppend( Vec_WecPushLevel(vFronts), vTemp );
    }
    //Vec_WecPrint( vFronts, 0 );
    Vec_IntFree( vTemp );
    return vFronts;
}

/**Function*************************************************************

  Synopsis    [Print grid.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_GraphPathPrint4(pBuffer *int32, vEdges *Vec_Int_t) {
    var Box [13][13]byte;
    var x, y int32;
    var e int32;
    var nEdges int32 = Vec_IntSize(vEdges)/2;
    for  y = 0; y < 13; y++  {
    for  x = 0; x < 13; x++  {
        if y % 4 == 0 && x % 4 == 0 {
            Box[y][x] = '*';
        } else {
            Box[y][x] = ' '; }
    }
    }
    for  e = 0; e < nEdges; e++  {
        var pNodes *int32  = Vec_IntEntryP(vEdges, 2*e);
        var y0 int32  = 4*(pNodes[0]/4);
        var x0 int32  = 4*(pNodes[0]%4);
        var y1 int32  = 4*(pNodes[1]/4);
        var x1 int32  = 4*(pNodes[1]%4);
        if !pBuffer[e] {
            continue; }
        if y0 == y1 {
            for  x = x0+1; x < x1; x++  {
                Box[y0][x] = '-';
            }
        
        } else if x0 == x1 {
            for  y = y0+1; y < y1; y++  {
                Box[y][x0] = '|';
            }
        
        } else { assert(false, "0"); }
    }
    for  y = 0; y < 13; y++, printf("\n")  {
    for  x = 0; x < 13; x++  {
        fmt.Printf("%c", Box[y][x]);
    }
    }
    fmt.Printf("\n\n=================================\n\n");
}
func Abc_GraphPathPrint5(pBuffer *int32, vEdges *Vec_Int_t) {
    var Box [17][17]byte;
    var x, y int32;
    var e int32;
    var nEdges int32 = Vec_IntSize(vEdges)/2;
    for  y = 0; y < 17; y++  {
    for  x = 0; x < 17; x++  {
        if y % 4 == 0 && x % 4 == 0 {
            Box[y][x] = '*';
        } else {
            Box[y][x] = ' '; }
    }
    }
    for  e = 0; e < nEdges; e++  {
        var pNodes *int32  = Vec_IntEntryP(vEdges, 2*e);
        var y0 int32  = 4*(pNodes[0]/5);
        var x0 int32  = 4*(pNodes[0]%5);
        var y1 int32  = 4*(pNodes[1]/5);
        var x1 int32  = 4*(pNodes[1]%5);
        if !pBuffer[e] {
            continue; }
        if y0 == y1 {
            for  x = x0+1; x < x1; x++  {
                Box[y0][x] = '-';
            }
        
        } else if x0 == x1 {
            for  y = y0+1; y < y1; y++  {
                Box[y][x0] = '|';
            }
        
        } else { assert(false, "0"); }
    }
    for  y = 0; y < 17; y++, printf("\n")  {
    for  x = 0; x < 17; x++  {
        fmt.Printf("%c", Box[y][x]);
    }
    }
    fmt.Printf("\n\n=================================\n\n");
}

/**Function*************************************************************

  Synopsis    [Count paths.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_GraphCountPaths_rec(Lev int32, Node int32, vNodes *Vec_Wec_t, pCache **float64, pBuffer *int32, vEdges *Vec_Int_t) float64 {
    var Res0, Res1 float64;
//    if ( Node == -2 ) Abc_GraphPathPrint4( pBuffer, vEdges );
    if Node == -2 {
        return 1; }
    if Node == -1 {
        return 0; }
    if pCache[Lev][Node] != -1.0 {
        return pCache[Lev][Node]; }
    pBuffer[Lev] = 0;
    Res0 = Abc_GraphCountPaths_rec( Lev+1, Vec_IntEntry( Vec_WecEntry(vNodes, Lev), 2*Node   ), vNodes, pCache, pBuffer, vEdges );
    pBuffer[Lev] = 1;
    Res1 = Abc_GraphCountPaths_rec( Lev+1, Vec_IntEntry( Vec_WecEntry(vNodes, Lev), 2*Node+1 ), vNodes, pCache, pBuffer, vEdges );
    return (pCache[Lev][Node] = Res0 + Res1);
}
func Abc_GraphCountPaths(vNodes *Vec_Wec_t, vEdges *Vec_Int_t) float64 {
    var i int32;
    var k int32;
    var pBuffer [1000]int32 ;
    // inlining of macro ABC_ALLOC
    var pCache **float64  = ((double * *) malloc(sizeof(double *) * (size_t)(Vec_WecSize(vNodes))));
    var vLevel *Vec_Int_t;   var Value float64;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vNodes)) && (((vLevel) = Vec_WecEntry(vNodes, i)), 1); i++  {
        // inlining of macro ABC_ALLOC
        pCache[i] = ((double *) malloc(sizeof(double) * (size_t)(Vec_IntSize(vLevel))));
        for  k = 0; k < Vec_IntSize(vLevel); k++  {
            pCache[i][k] = -1.0;
        }
    }
    Value = Abc_GraphCountPaths_rec( 0, 0, vNodes, pCache, pBuffer, vEdges );
    for  i = 0; i < Vec_WecSize(vNodes); i++  {
        // inlining of macro ABC_FREE
        ((pCache[i]) ? (free((char *) (pCache[i])), (pCache[i]) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((pCache) ? (free((char *) (pCache)), (pCache) = 0) : 0);
    return Value;
}

/**Function*************************************************************

  Synopsis    [Build AIG for paths.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_GraphDeriveGia_rec(p *Gia_Man_t, Lev int32, Node int32, vNodes *Vec_Wec_t, pCache **int32, pBuffer *int32, vEdges *Vec_Int_t) int32 {
    var Res0, Res1 int32;
    if Node == -2 {
        return 1; }
    if Node == -1 {
        return 0; }
    if pCache[Lev][Node] != -1 {
        return pCache[Lev][Node]; }
    pBuffer[Lev] = 0;
    Res0 = Abc_GraphDeriveGia_rec( p, Lev+1, Vec_IntEntry( Vec_WecEntry(vNodes, Lev), 2*Node   ), vNodes, pCache, pBuffer, vEdges );
    pBuffer[Lev] = 1;
    Res1 = Abc_GraphDeriveGia_rec( p, Lev+1, Vec_IntEntry( Vec_WecEntry(vNodes, Lev), 2*Node+1 ), vNodes, pCache, pBuffer, vEdges );
    return ( pCache[Lev][Node] = Gia_ManHashMux(p, Gia_Obj2Lit(p, Gia_ManCi(p, Lev)), Res1, Res0) );
}
func Abc_GraphDeriveGia(vNodes *Vec_Wec_t, vEdges *Vec_Int_t) *Gia_Man_t {
    var pCache **int32;
    var i int32;
    var Value int32;
    var pBuffer [1000]int32 ;  
    var vLevel *Vec_Int_t;
    // start AIG
    var pTemp *Gia_Man_t;
    var p *Gia_Man_t = Gia_ManStart( 1000 );
    p.pName = Abc_UtilStrsav("paths");
    for  i = 0; i < Vec_IntSize(vEdges)/2; i++  {
        Gia_ManAppendCi(p);
    }
    Gia_ManHashAlloc(p);
    // alloc cache
    // inlining of macro ABC_ALLOC
    pCache = ((int * *) malloc(sizeof(int *) * (size_t)(Vec_WecSize(vNodes))));
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vNodes)) && (((vLevel) = Vec_WecEntry(vNodes, i)), 1); i++  {
        // inlining of macro ABC_FALLOC
        pCache[i] = ((int *) memset(malloc(sizeof(int) * (size_t)(Vec_IntSize(vLevel))), 0xff, sizeof(int) * (size_t)(Vec_IntSize(vLevel))));
    }
    Value = Abc_GraphDeriveGia_rec( p, 0, 0, vNodes, pCache, pBuffer, vEdges );
    for  i = 0; i < Vec_WecSize(vNodes); i++  {
        // inlining of macro ABC_FREE
        ((pCache[i]) ? (free((char *) (pCache[i])), (pCache[i]) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((pCache) ? (free((char *) (pCache)), (pCache) = 0) : 0);
    // cleanup
    Gia_ManAppendCo( p, Value );
    p = Gia_ManCleanup( pTemp = p );
    Gia_ManStop( pTemp );
    return p;
}
func Abc_GraphDeriveGiaDump(vNodes *Vec_Wec_t, vEdges *Vec_Int_t, Size int32) {
    var pFileName string // was [100]byte in C

    var pGia *Gia_Man_t  = Abc_GraphDeriveGia( vNodes, vEdges );
    pFileName = fmt.Sprintf("grid_%dx%d_e%03d.aig", Size, Size, Vec_IntSize(vEdges)/2);
    Gia_AigerWrite( pGia, pFileName, 0, 0, 0 );
    Gia_ManStop( pGia );
    printf( "Finished dumping AIG into file \"%s\".\n", pFileName );
}

/**Function*************************************************************

  Synopsis    [Build frontier.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_GraphBuildState(vState *Vec_Int_t, e int32, x int32, vEdges *Vec_Int_t, vLife *Vec_Int_t, vFronts *Vec_Wec_t, pFront *int32, vStateNew *Vec_Int_t, fVerbose int32) int32 {
    var vFront *Vec_Int_t  = Vec_WecEntry( vFronts, e );
    var vFront2 *Vec_Int_t  = Vec_WecEntry( vFronts, e+1 );
    var pNodes *int32  = Vec_IntEntryP(vEdges, 2*e);
    var nNodes int32  = Vec_IntSize(vLife)/2;
    var i int32;
    var n int32;
    var Node int32;
    var First int32;
    var pEquivs [2]int32;
    assert(pNodes[0] < pNodes[1], "pNodes[0] < pNodes[1]");
    if fVerbose != 0 { fmt.Printf("Edge = %d. Arc = %d.\nCurrent state: ", e, x); }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFront)) && (((Node) = Vec_IntEntry(vFront, i)), 1); i++  {
        pFront[Node] = Vec_IntEntry(vState, i);
        if fVerbose != 0 { fmt.Printf("%d(%d) ", pFront[Node] & 0xFFFF, pFront[Node] >> 16); }
    }
    if fVerbose != 0 { fmt.Printf("\n"); }
    for  n = 0; n < 2; n++  {
        if Vec_IntEntry(vLife, 2*pNodes[n]) == e { // first time
            pFront[pNodes[n]] = pNodes[n]; }
    } // degree = 0; comp = singleton
    if x != 0 {
        if (pFront[pNodes[0]] & 0xFFFF) == (pFront[pNodes[1]] & 0xFFFF) { // the same comp
            return -1; } // const 0
        for  n = 0; n < 2; n++  {
            var Degree int32  = pFront[pNodes[n]] >> 16;
            if (pNodes[n] == 0 || pNodes[n] == nNodes-1) ? Degree >= 1 : Degree >= 2 {
                return -1; } // const 0
            pFront[pNodes[n]] += (1 << 16); // degree++
        }
    }
    // remember equivalence classes
    pEquivs[0] = pFront[pNodes[0]] & 0xFFFF;
    pEquivs[1] = pFront[pNodes[1]] & 0xFFFF;
    // remove some nodes from the frontier
    for  n = 0; n < 2; n++  {
        if Vec_IntEntry(vLife, 2*pNodes[n]+1) == e { // last time
        
            var Degree int32  = pFront[pNodes[n]] >> 16;
            if (pNodes[n] == 0 || pNodes[n] == nNodes-1) ? Degree != 1 : Degree != 0 && Degree != 2 {
                return -1; } // const 0
            // if it is part of the comp, update
            First = -1;
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vFront2)) && (((Node) = Vec_IntEntry(vFront2, i)), 1); i++  {
                assert(Node != pNodes[n], "Node != pNodes[n]");
                if (pFront[Node] & 0xFFFF) == pEquivs[n] {
                    if First == -1 {
                        First = Node; }
                    pFront[Node] = (pFront[Node] & 0xFFFF0000) | First;
                }
            }
            if First != -1 {
                pEquivs[n] = First; }
        }
    }
    if x != 0 {
        // union comp
        var First int32  = -1;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vFront2)) && (((Node) = Vec_IntEntry(vFront2, i)), 1); i++  {
            if (pFront[Node] & 0xFFFF) == pEquivs[0] || (pFront[Node] & 0xFFFF) == pEquivs[1] {
                if First == -1 {
                    First = Node; }
                pFront[Node] = (pFront[Node] & 0xFFFF0000) | First;
            }
        }
    }
    // create next state
    Vec_IntClear( vStateNew );
    if fVerbose != 0 { fmt.Printf("Next state: "); }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFront2)) && (((Node) = Vec_IntEntry(vFront2, i)), 1); i++  {
        Vec_IntPush( vStateNew, pFront[Node] );
        if fVerbose != 0 { fmt.Printf("%d(%d) ", pFront[Node] & 0xFFFF, pFront[Node] >> 16); }
    }
    if fVerbose != 0 { fmt.Printf("\n\n"); }
    return 1;
}
func Abc_GraphBuildFrontier(nSize int32, vEdges *Vec_Int_t, vLife *Vec_Int_t, vFronts *Vec_Wec_t, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var nPaths float64;
    var nEdges int32  = Vec_IntSize(vEdges)/2;
    var nNodes int32  = Vec_IntSize(vLife)/2;
    var vNodes *Vec_Wec_t  = Vec_WecAlloc( nEdges );
    var vStateNew *Vec_Int_t  = Vec_IntAlloc( nNodes );
    var vStateCount *Vec_Int_t  = Vec_IntAlloc( nEdges );
    // inlining of macro ABC_CALLOC
    int e, s, x, Next, * pFront = ((int *) calloc((size_t)(nNodes), sizeof(int)));
    var pThis *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );
    var pNext *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );
    Hsh_VecManAdd( pThis, vStateNew );
    for  e = 0; e < nEdges; e++  {
        var vNode *Vec_Int_t  = Vec_WecPushLevel(vNodes);
        var nStates int32  = Hsh_VecSize( pThis );
        Vec_IntPush( vStateCount, nStates );
        if fVerbose != 0 {
            fmt.Printf("\n");
            fmt.Printf("Processing edge %d = {%d %d}\n", e, Vec_IntEntry(vEdges, 2*e), Vec_IntEntry(vEdges, 2*e+1));
            fmt.Printf("Frontier: ");  Vec_IntPrint( Vec_WecEntry(vFronts, e) );
            fmt.Printf("\n");
        }
        for  s = 0; s < nStates; s++  {
            var vState *Vec_Int_t  = Hsh_VecReadEntry(pThis, s);
            for  x = 0; x < 2; x++  {
                Next = Abc_GraphBuildState(vState, e, x, vEdges, vLife, vFronts, pFront, vStateNew, fVerbose);
                if Next == 1 {
                    if e == nEdges - 1 { // last edge
                        Next = -2; // const1
                    } else {
                        Next = Hsh_VecManAdd( pNext, vStateNew ); }
                }
                if fVerbose != 0 { fmt.Printf("Return value = %d\n", Next); }
                Vec_IntPush( vNode, Next );
            }
        }
        Hsh_VecManStop( pThis );
        pThis = pNext;
        pNext = Hsh_VecManStart( 1000 );
    }
    nPaths = Abc_GraphCountPaths(vNodes, vEdges);
    fmt.Printf("States = %8d   Paths = %24.0f  ", Vec_IntSum(vStateCount), nPaths); 
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if fVerbose != 0 {
        Vec_IntPrint( vStateCount ); }
    Abc_GraphDeriveGiaDump( vNodes, vEdges, nSize );
    // inlining of macro ABC_FREE
    ((pFront) ? (free((char *) (pFront)), (pFront) = 0) : 0);
    Vec_WecFree( vNodes );
    Vec_IntFree( vStateNew );
    Vec_IntFree( vStateCount );
    Hsh_VecManStop( pThis );
    Hsh_VecManStop( pNext );
}
func Abc_EnumerateFrontierTest(nSize int32) {
    //int nSize    = 3;
    var fVerbose int32;
    var vEdges *Vec_Int_t  = Abc_GraphGrid( nSize );
    var vLife *Vec_Int_t  = Abc_GraphNodeLife( vEdges, nSize );
    var vFronts *Vec_Wec_t  = Abc_GraphFrontiers( vEdges, vLife );

    Abc_GraphBuildFrontier( nSize, vEdges, vLife, vFronts, fVerbose );

    Vec_WecFree( vFronts );
    Vec_IntFree( vLife );
    Vec_IntFree( vEdges );
}


/**Function*************************************************************

  Synopsis    [Performs SAT-based path enumeration.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Word2Double(w word) float64 {
    var Res float64;  var i int32;
    for  i = 0; i < 64; i++  {
        if (w >> i) & 1 {
            Res += math.Pow(float64(2), float64(i)); }
    }
    return Res;
}
func Abc_GraphSolve(pGia *Gia_Man_t) {
    var nIters int32  = 1000;
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( pGia, 8, 0, 1, 0, 0 );
    var pSat *sat_solver; var vLits *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32;
    var k int32;
    var iLit int32;
    var nVars int32 = Gia_ManCiNum(pGia);
    var iCiVarBeg int32  = pCnf.nVars - nVars;
    var Total word;
    var Mint1 word;
    var Mint2 word;

    // restart the SAT solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, pCnf.nVars );
    // add timeframe clauses
    for  i = 0; i < pCnf.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
            assert(false, "0"); }
    }
    // create trivial assignment
    Vec_IntClear( vLits );
    for  k = 0; k < nVars; k++  {
        Vec_IntPush( vLits, Abc_Var2Lit(iCiVarBeg+k, 1) );
    }
    // generate random assignment
    for  i = 0; i < nIters; i++  {
        var Status int32  = sat_solver_solve_lexsat( pSat, Vec_IntArray(vLits), Vec_IntSize(vLits) );
        if Status != l_True {
            break; }
        assert(Status == l_True, "Status == l_True");
        // block this assignment
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, k)), 1); k++  {
            Vec_IntWriteEntry( vLits, k, Abc_LitNot(iLit) );
        }
        if !sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits) ) {
            break; }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, k)), 1); k++  {
            Vec_IntWriteEntry( vLits, k, Abc_LitNot(iLit) );
        }
        // collect new minterm
        Mint2 = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, k)), 1); k++  {
            if !Abc_LitIsCompl(iLit) {
                Mint2 |= ((word)1) << (nVars-1-k); }
        }
        if Mint1 == 0 {
            Mint1 = Mint2; }
        // report
        //printf( "Iter %3d : ", i );
        //Extra_PrintBinary( stdout, (unsigned *)&Mint2, Abc_MinInt(64, nVars) ); printf( "\n" );
    }
    //Mint1 = 0;
    Total = (Mint2-Mint1)/nIters;
    fmt.Printf("Vars = %d   Iters = %d   Ave = %.0f   Total = %.0f  ", nVars, nIters, Abc_Word2Double(Mint2-Mint1), Abc_Word2Double(Total));
    fmt.Printf("Estimate = %.0f\n", (math.Pow(float64(2), float64(nVars))-Abc_Word2Double(Mint1))/Abc_Word2Double((Mint2-Mint1)/nIters));

    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Vec_IntFree( vLits );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

