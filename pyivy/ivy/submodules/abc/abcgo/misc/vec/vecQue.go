package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecQue.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Priority queue.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecQue.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

    
type Vec_Que_t struct {
    nCap             int32;
    nSize             int32;
    pHeap            *int32;
    pOrder            *int32;
    pCostsFlt         **float32;  // owned by the caller
}

func Vec_QuePrio(p *Vec_Que_t, v int32) float32 { if *p.pCostsFlt {
    return (*p.pCostsFlt)[v];
} else {
    return v;
} }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
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
func Vec_QueAlloc(nCap int32) *Vec_Que_t {
    var p *Vec_Que_t;
    p = new(Vec_Que_t);
    if nCap < 16 {
        nCap = 16; }
    p.nSize  = 1;
    p.nCap   = nCap + 1;
    p.pHeap  = ABC_FALLOC( int, p.nCap );
    p.pOrder = ABC_FALLOC( int, p.nCap );
    return p;
}
func Vec_QueFree(p *Vec_Que_t) {
    ABC_FREE( p.pOrder );
    ABC_FREE( p.pHeap );
    ABC_FREE( p );
}
func Vec_QueFreeP(p **Vec_Que_t) {
    if *p {
        Vec_QueFree( *p ); }
    *p = nil;
}
func Vec_QueSetPriority(p *Vec_Que_t, pCosts **float32) {
    assert(p.pCostsFlt == nil, "p.pCostsFlt == nil");
    p.pCostsFlt = pCosts;
}
func Vec_QueGrow(p *Vec_Que_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pHeap  = ABC_REALLOC( int, p.pHeap,  nCapMin );
    p.pOrder = ABC_REALLOC( int, p.pOrder, nCapMin ); 
    memset( p.pHeap  + p.nCap, 0xff, (size_t)(nCapMin - p.nCap) * sizeof(int) );
    memset( p.pOrder + p.nCap, 0xff, (size_t)(nCapMin - p.nCap) * sizeof(int) );
    p.nCap   = nCapMin;
}
func Vec_QueClear(p *Vec_Que_t) {
    var i int32;
    assert(p.pHeap[0] == -1, "p.pHeap[0] == -1");
    for  i = 1; i < p.nSize; i++  {
        assert(p.pHeap[i] >= 0 && p.pOrder[p.pHeap[i]] == i, "p.pHeap[i] >= 0 && p.pOrder[p.pHeap[i]] == i");
        p.pOrder[p.pHeap[i]] = -1;
        p.pHeap[i] = -1;
    }
    p.nSize = 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_QueSize(p *Vec_Que_t) int32 {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.nSize - 1;
}
func Vec_QueTop(p *Vec_Que_t) int32 {
    if Vec_QueSize(p) > 0 {
        return p.pHeap[1];
    } else {
        return -1;
    }
}
func Vec_QueTopPriority(p *Vec_Que_t) float32 {
    if Vec_QueSize(p) > 0 {
        return Vec_QuePrio(p, p.pHeap[1]);
    } else {
        return -ABC_INFINITY;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_QueMoveUp(p *Vec_Que_t, v int32) int32 {
    var Cost float32  = Vec_QuePrio(p, v);
    var i int32       = p.pOrder[v];
    var parent int32  = i >> 1;
    var fMoved int32;
    assert(p.pOrder[v] != -1, "p.pOrder[v] != -1");
    assert(p.pHeap[i] == v, "p.pHeap[i] == v");
    for  i > 1 && Cost > Vec_QuePrio(p, p.pHeap[parent])  {
        p.pHeap[i]            = p.pHeap[parent];
        p.pOrder[p.pHeap[i]] = i;
        i                      = parent;
        parent                 = i >> 1;
        fMoved                 = 1;
    }
    p.pHeap[i]  = v;
    p.pOrder[v] = i;
    return fMoved;
}
func Vec_QueMoveDown(p *Vec_Que_t, v int32) {
    var Cost float32  = Vec_QuePrio(p, v);
    var i int32       = p.pOrder[v];
    var child int32   = i << 1;
    for  child < p.nSize  {
        if child + 1 < p.nSize && Vec_QuePrio(p, p.pHeap[child]) < Vec_QuePrio(p, p.pHeap[child+1]) {
            child++; }
        assert(child < p.nSize, "child < p.nSize");
        if Cost >= Vec_QuePrio(p, p.pHeap[child]) {
            break; }
        p.pHeap[i]            = p.pHeap[child];
        p.pOrder[p.pHeap[i]] = i;
        i                      = child;
        child                  = child << 1;
    }
    p.pHeap[i]  = v;
    p.pOrder[v] = i;
}
func Vec_QueUpdate(p *Vec_Que_t, v int32) {
    if !(Vec_QueMoveUp( p, v ) != 0) {
        Vec_QueMoveDown( p, v ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_QueIsMember(p *Vec_Que_t, v int32) int32 {
    assert(v >= 0, "v >= 0");
    return int32(( v < p.nCap && p.pOrder[v] >= 0 ));
}
func Vec_QuePush(p *Vec_Que_t, v int32) {
    if p.nSize >= p.nCap {
        Vec_QueGrow( p, Abc_MaxInt(p.nSize+1, 2*p.nCap) ); }
    if v >= p.nCap {
        Vec_QueGrow( p, Abc_MaxInt(v+1, 2*p.nCap) ); }
    assert(p.nSize < p.nCap, "p.nSize < p.nCap");
    assert(p.pOrder[v] == -1, "p.pOrder[v] == -1");
    assert(p.pHeap[p.nSize] == -1, "p.pHeap[p.nSize] == -1");
    p.pOrder[v] = p.nSize;
    p.pHeap[p.nSize++] = v;
    Vec_QueMoveUp( p, v );
}
func Vec_QuePop(p *Vec_Que_t) int32 {
    var v, Res int32;
    assert(p.nSize > 1, "p.nSize > 1");
    Res = p.pHeap[1];      p.pOrder[Res] = -1;
    if --p.nSize == 1 {
        p.pHeap[1] = -1;
        return Res;
    }
    v = p.pHeap[p.nSize]; p.pHeap[p.nSize] = -1;
    p.pHeap[1] = v;        p.pOrder[v] = 1;
    Vec_QueMoveDown( p, v );
    return Res;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_QuePrint(p *Vec_Que_t) {
    var i, k, m int32;
    for  i = k = 1; i < p.nSize; i += k, k *= 2  {
        for  m = 0; m < k; m++  {
            if i+m < p.nSize {
                fmt.Printf("%-5d", p.pHeap[i+m]); }
        }
        fmt.Printf("\n");
        for  m = 0; m < k; m++  {
            if i+m < p.nSize {
                fmt.Printf("%-5.0f", Vec_QuePrio(p, p.pHeap[i+m])); }
        }
        fmt.Printf("\n");
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_QueCheck(p *Vec_Que_t) {
    var i, child int32;
    assert(p.nSize > 0, "p.nSize > 0");
    assert(p.nSize <= p.nCap, "p.nSize <= p.nCap");
    // check mapping
    for  i = 0; i < p.nSize-1; i++  {
        assert(p.pOrder[i] > 0, "p.pOrder[i] > 0");
    }
    for  ; i < p.nCap; i++  {
        assert(p.pOrder[i] == -1, "p.pOrder[i] == -1");
    }
    // check heap
    assert(p.pHeap[0] == -1, "p.pHeap[0] == -1");
    for  i = 0; i < p.nSize-1; i++  {
        assert(p.pHeap[p.pOrder[i]] == i, "p.pHeap[p.pOrder[i]] == i");
    }
    for  i++; i < p.nCap; i++  {
        assert(p.pHeap[i] == -1, "p.pHeap[i] == -1");
    }
    // check heap property
    for  i = 1; i < p.nSize; i++  {
        child = i << 1;
        if child < p.nSize {
            assert(Vec_QuePrio(p, p.pHeap[i]) >= Vec_QuePrio(p, p.pHeap[child]), "Vec_QuePrio(p, p.pHeap[i]) >= Vec_QuePrio(p, p.pHeap[child])"); }
        child++;
        if child < p.nSize {
            assert(Vec_QuePrio(p, p.pHeap[i]) >= Vec_QuePrio(p, p.pHeap[child]), "Vec_QuePrio(p, p.pHeap[i]) >= Vec_QuePrio(p, p.pHeap[child])"); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_QueTest(vCosts *Vec_Flt_t) {
    var vTop *Vec_Int_t;
    var p *Vec_Que_t;
    var i, Entry int32;

    // start the queue
    p = Vec_QueAlloc( Vec_FltSize(vCosts) );
    Vec_QueSetPriority( p, Vec_FltArrayP(vCosts) );
    for  i = 0; i < Vec_FltSize(vCosts); i++  {
        Vec_QuePush( p, i );
    }
//    Vec_QuePrint( p );
    Vec_QueCheck( p );

    // find the topmost 10%
    vTop = Vec_IntAlloc( Vec_FltSize(vCosts) / 10 );
    for  Vec_IntSize(vTop) < Vec_FltSize(vCosts) / 10  {
        Vec_IntPush( vTop, Vec_QuePop(p) );
    }
//    Vec_IntPrint( vTop );
//    Vec_QueCheck( p ); // queque is not ready at this point!!!

    // put them back
    Vec_IntForEachEntry( vTop, Entry, i )
        Vec_QuePush( p, Entry );
    Vec_IntFree( vTop );
    Vec_QueCheck( p );

    Vec_QueFree( p );
}

/*
    {
        extern void Vec_QueTest( Vec_Flt_t * p );
        Vec_QueTest( p->vTimesOut );
    }
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_END



