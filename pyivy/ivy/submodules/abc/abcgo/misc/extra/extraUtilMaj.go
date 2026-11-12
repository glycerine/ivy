package extra

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [extraUtilMaj.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Path enumeration.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilMaj.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/vec/vecMem.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Gem_Man_t =  Gem_Man_t_;
type Gem_Obj_t =  Gem_Obj_t_;

type Gem_Man_t_ struct {
    nVars          int32;         // max variable count
    nWords          int32;        // truth tabel word count
    nObjsAlloc          int32;    // allocated objects
    nObjs          int32;         // used objects
    pObjs   *Gem_Obj_t;         // function objects
    vTtMem   *Vec_Mem_t;        // truth table memory and hash table
    pTtElems       **word;      // elementary truth tables
    fVerbose          int32;
}

type Gem_Obj_t_ struct { // 8 bytes

    nVars uint32; /* C bitfield: 4 bits; review packing and truncation */  // variable count
    nNodes uint32; /* C bitfield: 4 bits; review packing and truncation */  // node count
    History uint32; /* C bitfield: 8 bits; review packing and truncation */  // (i < j) ? {vi, vj} : {vi, 0}
    Groups uint32; /* C bitfield: 16 bits; review packing and truncation */  // mask with last vars in each symmetric group
    Predec          int32;        // predecessor 
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gem_PrintNode(p *Gem_Man_t, f int32, pLabel string, fUpdate int32) {
    var pObj *Gem_Obj_t  = p.pObjs + f; 
    var GroupsMod int32  = pObj.Groups;
    if !p.fVerbose {
        return; }
    fmt.Printf("Node %6d : %s  Pred = %6d  Vars = %d  Nodes = %d  History = %d%d  Profile: ", f, pLabel, pObj.Predec, pObj.nVars, pObj.nNodes, pObj.History & 0xF, pObj.History >> 4);
    Extra_PrintBinary2( os.Stdout, (unsigned*)&GroupsMod, p.nVars ); fmt.Printf("%s\n", fUpdate?" *":"");
}
func Gem_ManAlloc(nVars int32, fVerbose int32) *Gem_Man_t {
    
    assert(nVars <= 16, "nVars <= 16");
    // inlining of macro ABC_CALLOC
    p := &Gem_Man_t{};
    p.nVars      = nVars;
    p.nWords     = Abc_TtWordNum( nVars );
    p.nObjsAlloc = 10000000;
    p.nObjs      = 2;
    // inlining of macro ABC_CALLOC
    p.pObjs      = ((Gem_Obj_t *) calloc((size_t)(p.nObjsAlloc), sizeof(Gem_Obj_t)));
    p.pObjs[1].nVars = p.pObjs[1].Groups = 1; // buffer
    p.vTtMem     = Vec_MemAllocForTT( nVars, 0 );
    p.pTtElems   = (word **)Extra_ArrayAlloc( nVars + 4, p.nWords, sizeof(word) );
    p.fVerbose   = fVerbose;
    Abc_TtElemInit( p.pTtElems, nVars );
    Gem_PrintNode( p, 1, "Original", 0 );
    return p;
}
func Gem_ManFree(p *Gem_Man_t) int32 {
    Vec_MemHashFree( p.vTtMem );
    Vec_MemFree( p.vTtMem );
    // inlining of macro ABC_FREE
    ((p.pTtElems) ? (free((char *) (p.pTtElems)), (p.pTtElems) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pObjs) ? (free((char *) (p.pObjs)), (p.pObjs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
    return 1;
}
func Gem_ManRealloc(p *Gem_Man_t) {
    var nObjNew int32  = Abc_MinInt( 2 * p.nObjsAlloc, 0x7FFFFFFF );
    assert(p.nObjs == p.nObjsAlloc, "p.nObjs == p.nObjsAlloc");
    if p.nObjs == 0x7FFFFFFF {
        fmt.Printf("Hard limit on the number of nodes (0x7FFFFFFF) is reached. Quitting...\n");
os.Exit(int(1)); }
    assert(p.nObjs < nObjNew, "p.nObjs < nObjNew");
    fmt.Printf("Extending object storage: %d -> %d.\n", p.nObjsAlloc, nObjNew);
    // inlining of macro ABC_REALLOC
    
    if p.pObjs {
        p.pObjs = ((Gem_Obj_t *) realloc((char *)(p.pObjs), sizeof(Gem_Obj_t) * (size_t)(nObjNew)));
    } else {
        p.pObjs = ((Gem_Obj_t *) malloc(sizeof(Gem_Obj_t) * (size_t)(nObjNew)));
    }
    memset( p.pObjs + p.nObjsAlloc, 0, sizeof(Gem_Obj_t) * (nObjNew - p.nObjsAlloc) );
    p.nObjsAlloc = nObjNew;
}

/**Function*************************************************************

  Synopsis    [Derive groups using symmetry info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gem_GroupsDerive(pTruth *word, nVars int32, pCof0 *word, pCof1 *word) int32 {
    var v int32;
    var Res int32 = 1 << (nVars-1);
    for  v = 0; v < nVars-1; v++  {
        if !Abc_TtVarsAreSymmetric(pTruth, nVars, v, v+1, pCof0, pCof1) {
            Res |= (1 << v); }
    }
    return Res;
}

/**Function*************************************************************

  Synopsis    [Extends function f by replacing var i with a new gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Gem_GroupVarRemove( int Groups, int i ) // remove i-th var
{
    var Mask int32;
    if i != 0 {
        Mask = Abc_InfoMask( i );
    } else {
        Mask = 0;
    }
    assert(i >= 0, "i >= 0");
    assert((Groups >> i) & 1, "(Groups >> i) & 1");
    return (Groups & Mask) | ((Groups & ~Mask) >> 1);
}
int Gem_GroupVarsInsert1( int Groups, int i, int fGroup ) // insert one bit after i
{
    var Mask int32;
    if i+1 {
        Mask = Abc_InfoMask( i+1 );
    } else {
        Mask = 0;
    }
    assert(i+1 >= 0, "i+1 >= 0");
    assert(!fGroup || i == -1 || ((Groups >> i) & 1), "!fGroup || i == -1 || ((Groups >> i) & 1)");
    assert(fGroup == 0 || fGroup == 1, "fGroup == 0 || fGroup == 1");
    return (Groups & Mask) | ((Groups & ~Mask) << 1) | (fGroup << (i+1));
}
int Gem_GroupVarsInsert3( int Groups, int i ) // insert group of 3 bits after i
{
    var Mask int32;
    if i+1 {
        Mask = Abc_InfoMask( i+1 );
    } else {
        Mask = 0;
    }
    assert(i+1 >= 0, "i+1 >= 0");
    assert(i == -1 || (Groups >> i) & 1, "i == -1 || (Groups >> i) & 1");
    return (Groups & Mask) | ((Groups & ~Mask) << 3) | (0x4 << (i+1));
}
func Gem_GroupUnpack(Groups int32, pVars *int32) int32 {
    var v, nGroups int32;
    for  v = 0; Groups; v++, Groups >>= 1  {
        if Groups & 1 {
            pVars[nGroups++] = v; }
    }
    return nGroups;
}
func Gem_FuncFindPlace(pTruth *word, nWords int32, Groups int32, pBest *word, fOneVar int32) int32 {
    var pLast [16]int32;
    var nGroups int32 = Gem_GroupUnpack( Groups, pLast );
    var g int32;
    var v int32;
    var Value int32;
    var BestPlace int32;
    if nGroups != 0 {
        BestPlace = pLast[nGroups - 1];
    } else {
        BestPlace = -1;
    }
    assert(nGroups >= 0, "nGroups >= 0");
    Abc_TtCopy( pBest, pTruth, nWords, 0 );
    for  g = nGroups - 1; g >= 0; g--  {
        var Limit int32;
        if g != 0 {
            Limit = pLast[g-1];
        } else {
            Limit = -1;
        }
        for  v = pLast[g]; v > Limit; v--  {
            Abc_TtSwapAdjacent( pTruth, nWords, v+0 );
            if fOneVar != 0 { 
                continue; }
            Abc_TtSwapAdjacent( pTruth, nWords, v+1 );
            Abc_TtSwapAdjacent( pTruth, nWords, v+2 );
        }
        Value = memcmp(pBest, pTruth, sizeof(word)*nWords);
        if Value < 0 {
            Abc_TtCopy( pBest, pTruth, nWords, 0 );
            BestPlace = Limit;
        }
    }
    return BestPlace;
}
func Gem_FuncExpand(p *Gem_Man_t, f int32, i int32) {
    var pNew *Gem_Obj_t = p.pObjs + p.nObjs;
    var pObj *Gem_Obj_t = p.pObjs + f; 
    var pTruth *word    = Vec_MemReadEntry( p.vTtMem, f );
    var pResult *word   = p.pTtElems[p.nVars];
    var pCofs [2]*word = [2]*word{ p.pTtElems[p.nVars+2], p.pTtElems[p.nVars+3] };
    var v, iFunc int32;
    var pCanonPermC [16]byte;
    assert(i < int32(pObj.nVars), "i < int32(pObj.nVars)");
    assert(int32(pObj.nVars) + 2 <= p.nVars, "int32(pObj.nVars) + 2 <= p.nVars");
    Abc_TtCopy( pResult, pTruth, p.nWords, 0 );
    // move i variable to the end
    for  v = i; v < int32(pObj.nVars)-1; v++  {
        Abc_TtSwapAdjacent( pResult, p.nWords, v );
    }
    // create new symmetric group
    assert(v == int32(pObj.nVars)-1, "v == int32(pObj.nVars)-1");
    Abc_TtCofactor0p( pCofs[0], pResult, p.nWords, v );
    Abc_TtCofactor1p( pCofs[1], pResult, p.nWords, v );
    Abc_TtMaj( pResult, p.pTtElems[v], p.pTtElems[v+1], p.pTtElems[v+2], p.nWords );
    Abc_TtMux( pResult, pResult, pCofs[1], pCofs[0], p.nWords );
    // canonicize
    //Extra_PrintHex( stdout, (unsigned*)pResult, pObj->nVars + 2 ); printf("\n");
    Abc_TtCanonicizePerm( pResult, pObj.nVars + 2, pCanonPermC );
    Abc_TtStretch6( pResult, Abc_MaxInt(6, pObj.nVars+2), p.nVars );
    //Extra_PrintHex( stdout, (unsigned*)pResult, pObj->nVars + 2 ); printf("\n\n");
    iFunc = Vec_MemHashInsert( p.vTtMem, pResult );
    if iFunc < p.nObjs {
        return; }
    assert(iFunc == p.nObjs, "iFunc == p.nObjs");
    pNew.nVars   = pObj.nVars + 2;
    pNew.nNodes  = pObj.nNodes + 1;
    pNew.Groups  = Gem_GroupsDerive( pResult, pNew.nVars, pCofs[0], pCofs[1] );
    pNew.Predec  = f;
    pNew.History = i; 
    Gem_PrintNode( p, iFunc, "Expand  ", 0 );

    assert(p.nObjs < p.nObjsAlloc, "p.nObjs < p.nObjsAlloc");
    if ++p.nObjs == p.nObjsAlloc {
        Gem_ManRealloc( p ); }
}

/**Function*************************************************************

  Synopsis    [Reduces function f by crossbaring variables i and j.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gem_FuncCheckMajority(p *Gem_Man_t, f int32) int32 {
    var pObj *Gem_Obj_t  = p.pObjs + f;
    var pTruth *word  = Vec_MemReadEntry( p.vTtMem, f );
    var Polar int32  = Abc_TtIsFullySymmetric( pTruth, pObj.nVars );
    if Polar != -1 {
        var nHalfVars int32  = (pObj.nVars+1) >> 1;
        var Mask int32  = Abc_Tt6Mask( nHalfVars );
        fmt.Printf("Found symmetric %d-variable function: ", pObj.nVars);
        Extra_PrintBinary2( os.Stdout, (unsigned *)&Polar, pObj.nVars + 1 );
        fmt.Printf("   ");
        if (pObj.nVars & 1) && Polar == (Mask << nHalfVars) {
            fmt.Printf("This is majority-%d.\n", pObj.nVars);
            return 0;
        }
        fmt.Printf("\n");
    }
    return 0;
}
func Gem_FuncReduce(p *Gem_Man_t, f int32, i int32, j int32) int32 {
    var pNew *Gem_Obj_t = p.pObjs + p.nObjs;
    var pObj *Gem_Obj_t = p.pObjs + f; 
    var pTruth *word    = Vec_MemReadEntry( p.vTtMem, f );
    var pResult *word   = p.pTtElems[p.nVars];
    var pCofs [2]*word = [2]*word{ p.pTtElems[p.nVars+2], p.pTtElems[p.nVars+3] };
    var v, iFunc int32;
    var pCanonPermC [16]byte;
    assert(i < j && j < 16, "i < j && j < 16");
    Abc_TtCopy( pResult, pTruth, p.nWords, 0 );
    // move j variable to the end
    for  v = j; v < int32(pObj.nVars)-1; v++  {
        Abc_TtSwapAdjacent( pResult, p.nWords, v );
    }
    assert(v == int32(pObj.nVars)-1, "v == int32(pObj.nVars)-1");
    // move i variable to the end
    for  v = i; v < int32(pObj.nVars)-2; v++  {
        Abc_TtSwapAdjacent( pResult, p.nWords, v );
    }
    assert(v == int32(pObj.nVars)-2, "v == int32(pObj.nVars)-2");
    // create new variable
    Abc_TtCofactor0p( pCofs[0], pResult, p.nWords, v+1 );
    Abc_TtCofactor1p( pCofs[1], pResult, p.nWords, v+1 );
    Abc_TtCofactor0( pCofs[0], p.nWords, v );
    Abc_TtCofactor1( pCofs[1], p.nWords, v );
    Abc_TtMux( pResult, p.pTtElems[v], pCofs[1], pCofs[0], p.nWords );
    // canonicize
    //Extra_PrintHex( stdout, (unsigned*)pResult, pObj->nVars - 1 ); printf("\n");
    Abc_TtCanonicizePerm( pResult, pObj.nVars - 1, pCanonPermC );
    Abc_TtStretch6( pResult, Abc_MaxInt(6, pObj.nVars-1), p.nVars );
    //Extra_PrintHex( stdout, (unsigned*)pResult, pObj->nVars - 1 ); printf("\n\n");
    iFunc = Vec_MemHashInsert( p.vTtMem, pResult );
    if iFunc < p.nObjs {
        return 0; }
    assert(iFunc == p.nObjs, "iFunc == p.nObjs");
    pNew.nVars   = pObj.nVars - 1;
    pNew.nNodes  = pObj.nNodes;
    pNew.Groups  = Gem_GroupsDerive( pResult, pNew.nVars, pCofs[0], pCofs[1] );
    pNew.Predec  = f;
    pNew.History = (j << 4) | i;
    Gem_PrintNode( p, iFunc, "Crossbar", 0 );
    Gem_FuncCheckMajority( p, iFunc );

    assert(p.nObjs < p.nObjsAlloc, "p.nObjs < p.nObjsAlloc");
    if ++p.nObjs == p.nObjsAlloc {
        Gem_ManRealloc( p ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gem_Enumerate(nVars int32, fDump int32, fVerbose int32) int32 {
    var clk abctime  = Abc_Clock();
    var p *Gem_Man_t  = Gem_ManAlloc( nVars, fVerbose );
    var v int32;
    var f int32;
    var i int32;
    var j int32;
    var nObjsStop int32 = 1;
    for  v = 1; v <= nVars-2; v++  {
        // expand functions by adding a gate
        var nObjsStopPrev int32  = nObjsStop;
        nObjsStop = p.nObjs;
        fmt.Printf("Expanding  var %2d (functions = %10d)  ", v, p.nObjs);
        Abc_PrintTime( 0, "Time", Abc_Clock() - clk );
        for  f = 0; f < nObjsStop; f++  {
            if v == int32(p.pObjs[f].nVars) || (v > int32(p.pObjs[f].nVars) && f >= nObjsStopPrev) {
                for  i = 0; i < v; i++  {
                    if int32(p.pObjs[f].Groups) & (1 << i) {
                        Gem_FuncExpand( p, f, i ); }
                } }
        }
        // reduce functions by adding a crossbar
        fmt.Printf("Connecting var %2d (functions = %10d)  ", v, p.nObjs);
        Abc_PrintTime( 0, "Time", Abc_Clock() - clk );
        for  f = nObjsStop; f < p.nObjs; f++  {
            for  i = 0; i < int32(p.pObjs[f].nVars); i++  {
                if int32(p.pObjs[f].Groups) & (1 << i) {
                    for  j = i+1; j < int32(p.pObjs[f].nVars); j++  {
                        if int32(p.pObjs[f].Groups) & (1 << j) {
                            if Gem_FuncReduce( p, f, i, j ) != 0 {
                                return Gem_ManFree( p ); } }
                    } }
            }
        }
    }
    fmt.Printf("Finished          (functions = %10d)  ", p.nObjs);
    Abc_PrintTime( 0, "Time", Abc_Clock() - clk );
    if fDump != 0 { Vec_MemDumpTruthTables( p.vTtMem, "enum", nVars ); }
    Gem_ManFree( p );
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

