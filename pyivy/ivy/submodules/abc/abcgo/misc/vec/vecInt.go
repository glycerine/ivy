package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable arrays of integers.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

         
type Vec_Int_t struct {
    nCap              int32;
    nSize              int32;
    pArray             *int32;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Vec_IntForEachEntry( vVec, Entry, i )                                               \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_IntForEachEntryStart( vVec, Entry, i, Start )                                   \
//     for ( i = Start; (i < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_IntForEachEntryStop( vVec, Entry, i, Stop )                                     \
//     for ( i = 0; (i < Stop) && (((Entry) = Vec_IntEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_IntForEachEntryStartStop( vVec, Entry, i, Start, Stop )                         \
//     for ( i = Start; (i < Stop) && (((Entry) = Vec_IntEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_IntForEachEntryReverse( vVec, pEntry, i )                                       \
//     for ( i = Vec_IntSize(vVec) - 1; (i >= 0) && (((pEntry) = Vec_IntEntry(vVec, i)), 1); i-- )

// Original macro:
// #define Vec_IntForEachEntryTwo( vVec1, vVec2, Entry1, Entry2, i )                           \
//     for ( i = 0; (i < Vec_IntSize(vVec1)) && (((Entry1) = Vec_IntEntry(vVec1, i)), 1) && (((Entry2) = Vec_IntEntry(vVec2, i)), 1); i++ )

// Original macro:
// #define Vec_IntForEachEntryTwoStart( vVec1, vVec2, Entry1, Entry2, i, Start )               \
//     for ( i = Start; (i < Vec_IntSize(vVec1)) && (((Entry1) = Vec_IntEntry(vVec1, i)), 1) && (((Entry2) = Vec_IntEntry(vVec2, i)), 1); i++ )

// Original macro:
// #define Vec_IntForEachEntryDouble( vVec, Entry1, Entry2, i )                                \
//     for ( i = 0; (i+1 < Vec_IntSize(vVec)) && (((Entry1) = Vec_IntEntry(vVec, i)), 1) && (((Entry2) = Vec_IntEntry(vVec, i+1)), 1); i += 2 )

// Original macro:
// #define Vec_IntForEachEntryDoubleStart( vVec, Entry1, Entry2, i, Start )                    \
//     for ( i = Start; (i+1 < Vec_IntSize(vVec)) && (((Entry1) = Vec_IntEntry(vVec, i)), 1) && (((Entry2) = Vec_IntEntry(vVec, i+1)), 1); i += 2 )

// Original macro:
// #define Vec_IntForEachEntryTriple( vVec, Entry1, Entry2, Entry3, i )                        \
//     for ( i = 0; (i+2 < Vec_IntSize(vVec)) && (((Entry1) = Vec_IntEntry(vVec, i)), 1) && (((Entry2) = Vec_IntEntry(vVec, i+1)), 1) && (((Entry3) = Vec_IntEntry(vVec, i+2)), 1); i += 3 )

// Original macro:
// #define Vec_IntForEachEntryThisNext( vVec, This, Next, i )                                  \
//     for ( i = 0, (This) = (Next) = (Vec_IntSize(vVec) ? Vec_IntEntry(vVec, 0) : -1); (i+1 < Vec_IntSize(vVec)) && (((Next) = Vec_IntEntry(vVec, i+1)), 1); i += 2, (This) = (Next) )

// Original macro:
// #define Vec_IntForEachEntryInVec( vVec2, vVec, Entry, i )                                   \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec2, Vec_IntEntry(vVec, i))), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAlloc(nCap int32) *Vec_Int_t {
    var p *Vec_Int_t;
    p = new(Vec_Int_t);
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( int, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}
func Vec_IntAllocExact(nCap int32) *Vec_Int_t {
    var p *Vec_Int_t;
    assert(nCap >= 0, "nCap >= 0");
    p = new(Vec_Int_t);
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( int, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given size and cleans it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntStart(nSize int32) *Vec_Int_t {
    var p *Vec_Int_t;
    p = Vec_IntAlloc( nSize );
    p.nSize = nSize;
    if p.pArray { memset( p.pArray, 0, sizeof(int) * (size_t)nSize ); }
    return p;
}
func Vec_IntStartFull(nSize int32) *Vec_Int_t {
    var p *Vec_Int_t;
    p = Vec_IntAlloc( nSize );
    p.nSize = nSize;
    if p.pArray { memset( p.pArray, 0xff, sizeof(int) * (size_t)nSize ); }
    return p;
}
func Vec_IntStartRange(First int32, Range int32) *Vec_Int_t {
    var p *Vec_Int_t;
    var i int32;
    p = Vec_IntAlloc( Range );
    p.nSize = Range;
    for  i = 0; i < Range; i++  {
        p.pArray[i] = First + i;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given size and cleans it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntStartNatural(nSize int32) *Vec_Int_t {
    var p *Vec_Int_t;
    var i int32;
    p = Vec_IntAlloc( nSize );
    p.nSize = nSize;
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = i;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAllocArray(pArray *int32, nSize int32) *Vec_Int_t {
    var p *Vec_Int_t;
    p = new(Vec_Int_t);
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = pArray;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAllocArrayCopy(pArray *int32, nSize int32) *Vec_Int_t {
    var p *Vec_Int_t;
    p = new(Vec_Int_t);
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = ABC_ALLOC( int, nSize );
    memcpy( p.pArray, pArray, sizeof(int) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntDup(pVec *Vec_Int_t) *Vec_Int_t {
    var p *Vec_Int_t;
    p = new(Vec_Int_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nSize;
    if p.nCap {
        p.pArray = ABC_ALLOC( int, p.nCap );
    } else {
        p.pArray = nil;
    }
    memcpy( p.pArray, pVec.pArray, sizeof(int) * (size_t)pVec.nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Transfers the array into another vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntDupArray(pVec *Vec_Int_t) *Vec_Int_t {
    var p *Vec_Int_t;
    p = new(Vec_Int_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    p.pArray = pVec.pArray;
    pVec.nSize  = 0;
    pVec.nCap   = 0;
    pVec.pArray = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntZero(p *Vec_Int_t) {
    p.pArray = nil;
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_IntErase(p *Vec_Int_t) {
    ABC_FREE( p.pArray );
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_IntFree(p *Vec_Int_t) {
    ABC_FREE( p.pArray );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFreeP(p **Vec_Int_t) {
    if *p == nil {
        return; }
    ABC_FREE( (*p).pArray );
    ABC_FREE( (*p) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntReleaseArray(p *Vec_Int_t) *int32 {
    var pArray *int32  = p.pArray;
    p.nCap = 0;
    p.nSize = 0;
    p.pArray = nil;
    return pArray;
}
func Vec_IntReleaseNewArray(p *Vec_Int_t) *int32 {
    var pArray *int32  = ABC_ALLOC( int, p.nSize+1 );
    pArray[0] = p.nSize+1;
    memcpy( pArray+1, p.pArray, sizeof(int)*(size_t)p.nSize );
    return pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntArray(p *Vec_Int_t) *int32 {
    return p.pArray;
}
func Vec_IntArrayP(p *Vec_Int_t) **int32 {
    return &p.pArray;
}
func Vec_IntLimit(p *Vec_Int_t) *int32 {
    return p.pArray + p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSize(p *Vec_Int_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntCap(p *Vec_Int_t) int32 {
    return p.nCap;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntMemory(p *Vec_Int_t) float64 {
    if !(p != nil) {
        return 0.0;
    } else {
        return 1.0 * sizeof(int) * (size_t)p.nCap + sizeof(Vec_Int_t);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntEntry(p *Vec_Int_t, i int32) int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntEntryP(p *Vec_Int_t, i int32) *int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray + i;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntWriteEntry(p *Vec_Int_t, i int32, Entry int32) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAddToEntry(p *Vec_Int_t, i int32, Addition int32) int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i] += Addition;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntUpdateEntry(p *Vec_Int_t, i int32, Value int32) {
    if Vec_IntEntry( p, i ) < Value {
        Vec_IntWriteEntry( p, i, Value ); }
}
func Vec_IntDowndateEntry(p *Vec_Int_t, i int32, Value int32) {
    if Vec_IntEntry( p, i ) > Value {
        Vec_IntWriteEntry( p, i, Value ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntEntryLast(p *Vec_Int_t) int32 {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[p.nSize-1];
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntGrow(p *Vec_Int_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pArray = ABC_REALLOC( int, p.pArray, nCapMin ); 
    assert(p.pArray, "p.pArray");
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntGrowResize(p *Vec_Int_t, nCapMin int32) {
    p.nSize  = nCapMin;
    if p.nCap >= nCapMin {
        return; }
    p.pArray = ABC_REALLOC( int, p.pArray, nCapMin ); 
    assert(p.pArray, "p.pArray");
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFill(p *Vec_Int_t, nSize int32, Fill int32) {
    var i int32;
    Vec_IntGrow( p, nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = Fill;
    }
    p.nSize = nSize;
}
func Vec_IntFillTwo(p *Vec_Int_t, nSize int32, FillEven int32, FillOdd int32) {
    var i int32;
    Vec_IntGrow( p, nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = (i & 1) ? FillOdd : FillEven;
    }
    p.nSize = nSize;
}
func Vec_IntFillNatural(p *Vec_Int_t, nSize int32) {
    var i int32;
    Vec_IntGrow( p, nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = i;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFillExtra(p *Vec_Int_t, nSize int32, Fill int32) {
    var i int32;
    if nSize <= p.nSize {
        return; }
    if nSize > 2 * p.nCap {
        Vec_IntGrow( p, nSize );
    } else if nSize > p.nCap {
        Vec_IntGrow( p, 2 * p.nCap ); }
    for  i = p.nSize; i < nSize; i++  {
        p.pArray[i] = Fill;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Returns the entry even if the place not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntGetEntry(p *Vec_Int_t, i int32) int32 {
    Vec_IntFillExtra( p, i + 1, 0 );
    return Vec_IntEntry( p, i );
}
func Vec_IntGetEntryFull(p *Vec_Int_t, i int32) int32 {
    Vec_IntFillExtra( p, i + 1, -1 );
    return Vec_IntEntry( p, i );
}

/**Function*************************************************************

  Synopsis    [Returns the entry even if the place not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntGetEntryP(p *Vec_Int_t, i int32) *int32 {
    Vec_IntFillExtra( p, i + 1, 0 );
    return Vec_IntEntryP( p, i );
}

/**Function*************************************************************

  Synopsis    [Inserts the entry even if the place does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSetEntry(p *Vec_Int_t, i int32, Entry int32) {
    Vec_IntFillExtra( p, i + 1, 0 );
    Vec_IntWriteEntry( p, i, Entry );
}
func Vec_IntSetEntryFull(p *Vec_Int_t, i int32, Entry int32) {
    Vec_IntFillExtra( p, i + 1, -1 );
    Vec_IntWriteEntry( p, i, Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntShrink(p *Vec_Int_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntClear(p *Vec_Int_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPush(p *Vec_Int_t, Entry int32) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_IntGrow( p, 16 );
        } else {
            Vec_IntGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}
func Vec_IntPushTwo(p *Vec_Int_t, Entry1 int32, Entry2 int32) {
    Vec_IntPush( p, Entry1 );
    Vec_IntPush( p, Entry2 );
}
func Vec_IntPushThree(p *Vec_Int_t, Entry1 int32, Entry2 int32, Entry3 int32) {
    Vec_IntPush( p, Entry1 );
    Vec_IntPush( p, Entry2 );
    Vec_IntPush( p, Entry3 );
}
func Vec_IntPushFour(p *Vec_Int_t, Entry1 int32, Entry2 int32, Entry3 int32, Entry4 int32) {
    Vec_IntPush( p, Entry1 );
    Vec_IntPush( p, Entry2 );
    Vec_IntPush( p, Entry3 );
    Vec_IntPush( p, Entry4 );
}
func Vec_IntPushArray(p *Vec_Int_t, pEntries *int32, nEntries int32) {
    var i int32;
    for  i = 0; i < nEntries; i++  {
        Vec_IntPush( p, pEntries[i] );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPushFirst(p *Vec_Int_t, Entry int32) {
    var i int32;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_IntGrow( p, 16 );
        } else {
            Vec_IntGrow( p, 2 * p.nCap ); }
    }
    p.nSize++;
    for  i = p.nSize - 1; i >= 1; i--  {
        p.pArray[i] = p.pArray[i-1];
    }
    p.pArray[0] = Entry;
}

/**Function*************************************************************

  Synopsis    [Inserts the entry while preserving the increasing order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPushOrder(p *Vec_Int_t, Entry int32) {
    var i int32;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_IntGrow( p, 16 );
        } else {
            Vec_IntGrow( p, 2 * p.nCap ); }
    }
    p.nSize++;
    for  i = p.nSize-2; i >= 0; i--  {
        if p.pArray[i] > Entry {
            p.pArray[i+1] = p.pArray[i];
        } else {
            break; }
    }
    p.pArray[i+1] = Entry;
}
func Vec_IntPushOrderCost(p *Vec_Int_t, Entry int32, vCost *Vec_Int_t) {
    var i int32;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_IntGrow( p, 16 );
        } else {
            Vec_IntGrow( p, 2 * p.nCap ); }
    }
    p.nSize++;
    for  i = p.nSize-2; i >= 0; i--  {
        if Vec_IntEntry(vCost, p.pArray[i]) > Vec_IntEntry(vCost, Entry) {
            p.pArray[i+1] = p.pArray[i];
        } else {
            break; }
    }
    p.pArray[i+1] = Entry;
}

/**Function*************************************************************

  Synopsis    [Inserts the entry while preserving the increasing order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPushOrderReverse(p *Vec_Int_t, Entry int32) {
    var i int32;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_IntGrow( p, 16 );
        } else {
            Vec_IntGrow( p, 2 * p.nCap ); }
    }
    p.nSize++;
    for  i = p.nSize-2; i >= 0; i--  {
        if p.pArray[i] < Entry {
            p.pArray[i+1] = p.pArray[i];
        } else {
            break; }
    }
    p.pArray[i+1] = Entry;
}

/**Function*************************************************************

  Synopsis    [Inserts the entry while preserving the increasing order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPushUniqueOrder(p *Vec_Int_t, Entry int32) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Vec_IntPushOrder( p, Entry );
    return 0;
}
func Vec_IntPushUniqueOrderCost(p *Vec_Int_t, Entry int32, vCost *Vec_Int_t) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Vec_IntPushOrderCost( p, Entry, vCost );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPushUnique(p *Vec_Int_t, Entry int32) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Vec_IntPush( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the next nWords entries in the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFetch(p *Vec_Int_t, nWords int32) *uint32 {
    if nWords == 0 {
        return nil; }
    assert(nWords > 0, "nWords > 0");
    p.nSize += nWords;
    if p.nSize > p.nCap {
//         Vec_IntGrow( p, 2 * p->nSize );
        return nil;
    }
    return ((unsigned *)p.pArray) + p.nSize - nWords;
}

/**Function*************************************************************

  Synopsis    [Returns the last entry and removes it from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPop(p *Vec_Int_t) int32 {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    [Find entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFind(p *Vec_Int_t, Entry int32) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntRemove(p *Vec_Int_t, Entry int32) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            break; }
    }
    if i == p.nSize {
        return 0; }
    assert(i < p.nSize, "i < p.nSize");
    for  i++; i < p.nSize; i++  {
        p.pArray[i-1] = p.pArray[i];
    }
    p.nSize--;
    return 1;
}
func Vec_IntRemove1(p *Vec_Int_t, Entry int32) int32 {
    var i int32;
    for  i = 1; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            break; }
    }
    if i >= p.nSize {
        return 0; }
    assert(i < p.nSize, "i < p.nSize");
    for  i++; i < p.nSize; i++  {
        p.pArray[i-1] = p.pArray[i];
    }
    p.nSize--;
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntDrop(p *Vec_Int_t, i int32) {
    var k int32;
    assert(i >= 0 && i < Vec_IntSize(p), "i >= 0 && i < Vec_IntSize(p)");
    p.nSize--;
    for  k = i; k < p.nSize; k++  {
        p.pArray[k] = p.pArray[k+1];
    }
}

/**Function*************************************************************

  Synopsis    [Interts entry at the index iHere. Shifts other entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntInsert(p *Vec_Int_t, iHere int32, Entry int32) {
    var i int32;
    assert(iHere >= 0 && iHere <= p.nSize, "iHere >= 0 && iHere <= p.nSize");
    Vec_IntPush( p, 0 );
    for  i = p.nSize - 1; i > iHere; i--  {
        p.pArray[i] = p.pArray[i-1];
    }
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    [Find entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFindMax(p *Vec_Int_t) int32 {
    var i, Best int32;
    if p.nSize == 0 {
        return 0; }
    Best = p.pArray[0];
    for  i = 1; i < p.nSize; i++  {
        if Best < p.pArray[i] {
            Best = p.pArray[i]; }
    }
    return Best;
}

/**Function*************************************************************

  Synopsis    [Find entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFindMin(p *Vec_Int_t) int32 {
    var i, Best int32;
    if p.nSize == 0 {
        return 0; }
    Best = p.pArray[0];
    for  i = 1; i < p.nSize; i++  {
        if Best > p.pArray[i] {
            Best = p.pArray[i]; }
    }
    return Best;
}

/**Function*************************************************************

  Synopsis    [Reverses the order of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntReverseOrder(p *Vec_Int_t) {
    var i, Temp int32;
    for  i = 0; i < p.nSize/2; i++  {
        Temp = p.pArray[i];
        p.pArray[i] = p.pArray[p.nSize-1-i];
        p.pArray[p.nSize-1-i] = Temp;
    }
}

/**Function*************************************************************

  Synopsis    [Removes odd entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntRemoveOdd(p *Vec_Int_t) {
    var i int32;
    assert((p.nSize & 1) == 0, "(p.nSize & 1) == 0");
    p.nSize >>= 1;
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] = p.pArray[2*i];
    }
}
func Vec_IntRemoveEven(p *Vec_Int_t) {
    var i int32;
    assert((p.nSize & 1) == 0, "(p.nSize & 1) == 0");
    p.nSize >>= 1;
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] = p.pArray[2*i+1];
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntInvert(p *Vec_Int_t, Fill int32) *Vec_Int_t {
    var Entry, i int32;
    var vRes *Vec_Int_t  = Vec_IntAlloc( 0 );
    if Vec_IntSize(p) == 0 {
        return vRes; }
    Vec_IntFill( vRes, Vec_IntFindMax(p) + 1, Fill );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p)) && (((Entry) = Vec_IntEntry(p, i)), 1); i++  {
        if Entry != Fill {
            Vec_IntWriteEntry( vRes, Entry, i ); }
    }
    return vRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntCondense(p *Vec_Int_t, Fill int32) *Vec_Int_t {
    var Entry, i int32;
    var vRes *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(p) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p)) && (((Entry) = Vec_IntEntry(p, i)), 1); i++  {
        if Entry != Fill {
            Vec_IntPush( vRes, Entry ); }
    }
    return vRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSum(p *Vec_Int_t) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += p.pArray[i];
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntCountEntry(p *Vec_Int_t, Entry int32) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] == Entry);
    }
    return Counter;
}
func Vec_IntCountLarger(p *Vec_Int_t, Entry int32) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] > Entry);
    }
    return Counter;
}
func Vec_IntCountSmaller(p *Vec_Int_t, Entry int32) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] < Entry);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntCountPositive(p *Vec_Int_t) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] > 0);
    }
    return Counter;
}
func Vec_IntCountZero(p *Vec_Int_t) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] == 0);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAddPositive(p *Vec_Int_t) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] > 0 {
            Counter += p.pArray[i]; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Checks if two vectors are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntEqual(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    var i int32;
    if p1.nSize != p2.nSize {
        return 0; }
    for  i = 0; i < p1.nSize; i++  {
        if p1.pArray[i] != p2.pArray[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Counts the number of common entries.]

  Description [Assumes that the entries are non-negative integers that
  are not very large, so inversion of the array can be performed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntCountCommon(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    var vTemp *Vec_Int_t;
    var Entry, i, Counter int32;
    if Vec_IntSize(p1) < Vec_IntSize(p2) {
        vTemp = p1;
p1 = p2;
p2 = vTemp; }
    assert(Vec_IntSize(p1) >= Vec_IntSize(p2), "Vec_IntSize(p1) >= Vec_IntSize(p2)");
    vTemp = Vec_IntInvert( p2, -1 );
    Vec_IntFillExtra( vTemp, Vec_IntFindMax(p1) + 1, -1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p1)) && (((Entry) = Vec_IntEntry(p1, i)), 1); i++  {
        if Vec_IntEntry(vTemp, Entry) >= 0 {
            Counter++; }
    }
    Vec_IntFree( vTemp );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSortCompare1(pp1 *int32, pp2 *int32) int32 {
    // for some reason commenting out lines (as shown) led to crashing of the release version
    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 { //
        return 1; }
    return 0; //
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSortCompare2(pp1 *int32, pp2 *int32) int32 {
    // for some reason commenting out lines (as shown) led to crashing of the release version
    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 { //
        return 1; }
    return 0; //
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSort(p *Vec_Int_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(int), 
                        (int (*)(void *, void *)) Vec_IntSortCompare2 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(int), 
                        (int (*)(void *, void *)) Vec_IntSortCompare1 ); }
}
func Vec_IntSortMulti(p *Vec_Int_t, nMulti int32, fReverse int32) {
    assert(Vec_IntSize(p) % nMulti == 0, "Vec_IntSize(p) % nMulti == 0");
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)(p.nSize/nMulti), nMulti*sizeof(int), 
                        (int (*)(void *, void *)) Vec_IntSortCompare2 );
    } else {
        qsort( (void *)p.pArray, (size_t)(p.nSize/nMulti), nMulti*sizeof(int), 
                        (int (*)(void *, void *)) Vec_IntSortCompare1 ); }
}
func Vec_IntIsSorted(p *Vec_Int_t, fReverse int32) int32 {
    var i int32;
    for  i = 1; i < p.nSize; i++  {
        if fReverse ? (p.pArray[i-1] < p.pArray[i]) : (p.pArray[i-1] > p.pArray[i]) {
            return 0; }
    }
    return 1;            
}

/**Function*************************************************************

  Synopsis    [Leaves only unique entries.]

  Description [Returns the number of duplicated entried found.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntUniqify(p *Vec_Int_t) int32 {
    var i, k, RetValue int32;
    if p.nSize < 2 {
        return 0; }
    Vec_IntSort( p, 0 );
    for  i = k = 1; i < p.nSize; i++  {
        if p.pArray[i] != p.pArray[i-1] {
            p.pArray[k++] = p.pArray[i]; }
    }
    RetValue = p.nSize - k;
    p.nSize = k;
    return RetValue;
}
func Vec_IntCountDuplicates(p *Vec_Int_t) int32 {
    var RetValue int32;
    var pDup *Vec_Int_t  = Vec_IntDup( p );
    Vec_IntUniqify( pDup );
    RetValue = Vec_IntSize(p) - Vec_IntSize(pDup);
    Vec_IntFree( pDup );
    return RetValue;
}
func Vec_IntCheckUniqueSmall(p *Vec_Int_t) int32 {
    var i, k int32;
    for  i = 0; i < p.nSize; i++  {
        for  k = i+1; k < p.nSize; k++  {
            if p.pArray[i] == p.pArray[k] {
                return 0; }
        }
    }
    return 1;
}
func Vec_IntCountUnique(p *Vec_Int_t) int32 {
    var i int32;
    var Count int32;
    var Max int32 = Vec_IntFindMax(p);
    unsigned char * pPres = ABC_CALLOC( unsigned char, Max+1 );
    for  i = 0; i < p.nSize; i++  {
        if pPres[p.pArray[i]] == 0 {
            pPres[p.pArray[i]] = 1, Count++; }
    }
    ABC_FREE( pPres );
    return Count;
}

/**Function*************************************************************

  Synopsis    [Counts the number of unique pairs.]

  Description []
               
  SideEffects [] 

  SeeAlso     []

***********************************************************************/
func Vec_IntUniqifyPairs(p *Vec_Int_t) int32 {
    var i, k, RetValue int32;
    assert(p.nSize % 2 == 0, "p.nSize % 2 == 0");
    if p.nSize < 4 {
        return 0; }
    Vec_IntSortMulti( p, 2, 0 );
    for  i = k = 1; i < p.nSize/2; i++  {
        if p.pArray[2*i] != p.pArray[2*(i-1)] || p.pArray[2*i+1] != p.pArray[2*(i-1)+1] {
            p.pArray[2*k]   = p.pArray[2*i];
            p.pArray[2*k+1] = p.pArray[2*i+1];
            k++;
        }
    }
    RetValue = p.nSize/2 - k;
    p.nSize = 2*k;
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Counts the number of unique entries.]

  Description []
               
  SideEffects [] 

  SeeAlso     []

***********************************************************************/
static inline unsigned Vec_IntUniqueHashKeyDebug( unsigned char * pStr, int nChars, int TableMask )
{
    static unsigned s_BigPrimes[4] = {12582917, 25165843, 50331653, 100663319};
    var Key uint32; var c int32;
    for  c = 0; c < nChars; c++  {
        Key += uint32(pStr[c]) * s_BigPrimes[c & 3];
        fmt.Printf("%d : ", c);
        fmt.Printf("%3d  ", pStr[c]);
        fmt.Printf("%12d ", Key);
        fmt.Printf("%12d ", Key&TableMask);
        fmt.Printf("\n");
    }
    return Key;
}
func Vec_IntUniqueProfile(vData *Vec_Int_t, pTable *int32, pNexts *int32, TableMask int32, nIntSize int32) {
    var i, Key, Counter int32;
    for  i = 0; i <= TableMask; i++  {
        Counter = 0;
        for  Key = pTable[i]; Key != -1; Key = pNexts[Key]  {
            Counter++;
        }
        if Counter < 7 {
            continue; }
        fmt.Printf("%d\n", Counter);
        for  Key = pTable[i]; Key != -1; Key = pNexts[Key]  {
//            Extra_PrintBinary( stdout, (unsigned *)Vec_IntEntryP(vData, Key*nIntSize), 40 ), printf( "\n" );
//            Vec_IntUniqueHashKeyDebug( (unsigned char *)Vec_IntEntryP(vData, Key*nIntSize), 4*nIntSize, TableMask );
        }
    }
    fmt.Printf("\n");
}

static inline unsigned Vec_IntUniqueHashKey2( unsigned char * pStr, int nChars )
{
    static unsigned s_BigPrimes[4] = {12582917, 25165843, 50331653, 100663319};
    var Key uint32; var c int32;
    for  c = 0; c < nChars; c++  {
        Key += uint32(pStr[c]) * s_BigPrimes[c & 3];
    }
    return Key;
}

static inline unsigned Vec_IntUniqueHashKey( unsigned char * pStr, int nChars )
{
    static unsigned s_BigPrimes[16] = 
    {
        0x984b6ad9,0x18a6eed3,0x950353e2,0x6222f6eb,0xdfbedd47,0xef0f9023,0xac932a26,0x590eaf55,
        0x97d0a034,0xdc36cd2e,0x22736b37,0xdc9066b0,0x2eb2f98b,0x5d9c7baf,0x85747c9e,0x8aca1055
    };
    static unsigned s_BigPrimes2[16] = 
    {
        0x8d8a5ebe,0x1e6a15dc,0x197d49db,0x5bab9c89,0x4b55dea7,0x55dede49,0x9a6a8080,0xe5e51035,
        0xe148d658,0x8a17eb3b,0xe22e4b38,0xe5be2a9a,0xbe938cbb,0x3b981069,0x7f9c0c8e,0xf756df10
    };
    var Key uint32; var c int32;
    for  c = 0; c < nChars; c++  {
        Key += s_BigPrimes2[(2*c)&15]   * s_BigPrimes[uint32(pStr[c]) & 15] +
               s_BigPrimes2[(2*c+1)&15] * s_BigPrimes[uint32(pStr[c]) >> 4];
    }
    return Key;
}
func Vec_IntUniqueLookup(vData *Vec_Int_t, i int32, nIntSize int32, pNexts *int32, pStart *int32) *int32 {
    var pData *int32  = Vec_IntEntryP( vData, i*nIntSize );
    for  ; *pStart != -1; pStart = pNexts + *pStart  {
        if !memcmp( pData, Vec_IntEntryP(vData, *pStart*nIntSize), sizeof(int) * (size_t)nIntSize ) {
            return pStart; }
    }
    return pStart;
}
func Vec_IntUniqueCount(vData *Vec_Int_t, nIntSize int32, pvMap **Vec_Int_t) int32 {
    var nEntries int32   = Vec_IntSize(vData) / nIntSize;
    var TableMask int32  = (1 << Abc_Base2Log(nEntries)) - 1;
    var pTable *int32   = ABC_FALLOC( int, TableMask+1 );
    var pNexts *int32   = ABC_FALLOC( int, TableMask+1 );
    var pClass *int32   = ABC_ALLOC( int, nEntries );
    var i int32;
    var Key int32;
    var pEnt *int32;
    var nUnique int32;
    assert(nEntries * nIntSize == Vec_IntSize(vData), "nEntries * nIntSize == Vec_IntSize(vData)");
    for  i = 0; i < nEntries; i++  {
        pEnt = Vec_IntEntryP( vData, i*nIntSize );
        Key  = TableMask & Vec_IntUniqueHashKey( (unsigned char *)pEnt, 4*nIntSize );
        pEnt = Vec_IntUniqueLookup( vData, i, nIntSize, pNexts, pTable+Key );
        if *pEnt == -1 {
            *pEnt = i, nUnique++; }
        pClass[i] = *pEnt;
    }
//    Vec_IntUniqueProfile( vData, pTable, pNexts, TableMask, nIntSize );
    ABC_FREE( pTable );
    ABC_FREE( pNexts );
    if pvMap {
        *pvMap = Vec_IntAllocArray( pClass, nEntries );
    } else {
        ABC_FREE( pClass ); }
    return nUnique;
}
func Vec_IntUniqifyHash(vData *Vec_Int_t, nIntSize int32) *Vec_Int_t {
    var vMap, vUnique *Vec_Int_t;
    var i int32;
    var Ent int32;
    var nUnique int32 = Vec_IntUniqueCount( vData, nIntSize, &vMap );
    vUnique = Vec_IntAlloc( nUnique * nIntSize );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vMap)) && (((Ent) = Vec_IntEntry(vMap, i)), 1); i++  {
        if Ent < i { continue; }
        assert(Ent == i, "Ent == i");
        Vec_IntPushArray( vUnique, Vec_IntEntryP(vData, i*nIntSize), nIntSize );
    }
    assert(Vec_IntSize(vUnique) == nUnique * nIntSize, "Vec_IntSize(vUnique) == nUnique * nIntSize");
    Vec_IntFree( vMap );
    return vUnique;
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSortCompareUnsigned(pp1 *uint32, pp2 *uint32) int32 {
    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSortUnsigned(p *Vec_Int_t) {
    qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(int), 
                (int (*)(void *, void *)) Vec_IntSortCompareUnsigned );
}

/**Function*************************************************************

  Synopsis    [Returns the number of common entries.]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntTwoCountCommon(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t) int32 {
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    var Counter int32;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            pBeg1++, pBeg2++, Counter++;
        } else if *pBeg1 < *pBeg2 {
            pBeg1++;
        } else { 
            pBeg2++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collects common entries.]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntTwoFindCommon(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, vArr *Vec_Int_t) int32 {
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    Vec_IntClear( vArr );
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            Vec_IntPush( vArr, *pBeg1 );
pBeg1++;
pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            pBeg1++;
        } else { 
            pBeg2++; }
    }
    return Vec_IntSize(vArr);
}
func Vec_IntTwoFindCommonReverse(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, vArr *Vec_Int_t) int32 {
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    Vec_IntClear( vArr );
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            Vec_IntPush( vArr, *pBeg1 );
pBeg1++;
pBeg2++;
        } else if *pBeg1 > *pBeg2 {
            pBeg1++;
        } else { 
            pBeg2++; }
    }
    return Vec_IntSize(vArr);
}

/**Function*************************************************************

  Synopsis    [Collects and removes common entries]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntTwoRemoveCommon(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, vArr *Vec_Int_t) int32 {
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    var pBeg1New *int32  = vArr1.pArray;
    var pBeg2New *int32  = vArr2.pArray;
    Vec_IntClear( vArr );
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            Vec_IntPush( vArr, *pBeg1 );
pBeg1++;
pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            *pBeg1New++ = *pBeg1++;
        } else { 
            *pBeg2New++ = *pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pBeg1New++ = *pBeg1++;
    }
    for  pBeg2 < pEnd2  {
        *pBeg2New++ = *pBeg2++;
    }
    Vec_IntShrink( vArr1, pBeg1New - vArr1.pArray );
    Vec_IntShrink( vArr2, pBeg2New - vArr2.pArray );
    return Vec_IntSize(vArr);
}

/**Function*************************************************************

  Synopsis    [Removes entries of the second one from the first one.]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntTwoRemove(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t) int32 {
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    var pBeg1New *int32  = vArr1.pArray;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            pBeg1++, pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            *pBeg1New++ = *pBeg1++;
        } else { 
            pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pBeg1New++ = *pBeg1++;
    }
    Vec_IntShrink( vArr1, pBeg1New - vArr1.pArray );
    return Vec_IntSize(vArr1);
}

/**Function*************************************************************

  Synopsis    [Returns the result of merging the two vectors.]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntTwoMerge2Int(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, vArr *Vec_Int_t) {
    var pBeg *int32   = vArr.pArray;
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            *pBeg++ = *pBeg1++, pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            *pBeg++ = *pBeg1++;
        } else { 
            *pBeg++ = *pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pBeg++ = *pBeg1++;
    }
    for  pBeg2 < pEnd2  {
        *pBeg++ = *pBeg2++;
    }
    vArr.nSize = pBeg - vArr.pArray;
    assert(vArr.nSize <= vArr.nCap, "vArr.nSize <= vArr.nCap");
    assert(vArr.nSize >= vArr1.nSize, "vArr.nSize >= vArr1.nSize");
    assert(vArr.nSize >= vArr2.nSize, "vArr.nSize >= vArr2.nSize");
}
func Vec_IntTwoMerge(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t) *Vec_Int_t {
    var vArr *Vec_Int_t  = Vec_IntAlloc( vArr1.nSize + vArr2.nSize ); 
    Vec_IntTwoMerge2Int( vArr1, vArr2, vArr );
    return vArr;
}
func Vec_IntTwoMerge2(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, vArr *Vec_Int_t) {
    Vec_IntGrow( vArr, Vec_IntSize(vArr1) + Vec_IntSize(vArr2) );
    Vec_IntTwoMerge2Int( vArr1, vArr2, vArr );
}

/**Function*************************************************************

  Synopsis    [Returns the result of splitting of the two vectors.]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntTwoSplit(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, vArr *Vec_Int_t, vArr1n *Vec_Int_t, vArr2n *Vec_Int_t) {
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            Vec_IntPush( vArr, *pBeg1++ );
pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            Vec_IntPush( vArr1n, *pBeg1++ );
        } else { 
            Vec_IntPush( vArr2n, *pBeg2++ ); }
    }
    for  pBeg1 < pEnd1  {
        Vec_IntPush( vArr1n, *pBeg1++ );
    }
    for  pBeg2 < pEnd2  {
        Vec_IntPush( vArr2n, *pBeg2++ );
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSelectSort(pArray *int32, nSize int32) {
    var temp, i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if pArray[j] < pArray[best_i] {
                best_i = j; }
        }
        temp = pArray[i]; 
        pArray[i] = pArray[best_i]; 
        pArray[best_i] = temp;
    }
}
func Vec_IntSelectSortReverse(pArray *int32, nSize int32) {
    var temp, i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if pArray[j] > pArray[best_i] {
                best_i = j; }
        }
        temp = pArray[i]; 
        pArray[i] = pArray[best_i]; 
        pArray[best_i] = temp;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSelectSortCost(pArray *int32, nSize int32, vCosts *Vec_Int_t) {
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if Vec_IntEntry(vCosts, pArray[j]) < Vec_IntEntry(vCosts, pArray[best_i]) {
                best_i = j; }
        }
        ABC_SWAP( int, pArray[i], pArray[best_i] );
    }
}
func Vec_IntSelectSortCostReverse(pArray *int32, nSize int32, vCosts *Vec_Int_t) {
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if Vec_IntEntry(vCosts, pArray[j]) > Vec_IntEntry(vCosts, pArray[best_i]) {
                best_i = j; }
        }
        ABC_SWAP( int, pArray[i], pArray[best_i] );
    }
}

func Vec_IntSelectSortCost2(pArray *int32, nSize int32, pCosts *int32) {
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if pCosts[j] < pCosts[best_i] {
                best_i = j; }
        }
        ABC_SWAP( int, pArray[i], pArray[best_i] );
        ABC_SWAP( int, pCosts[i], pCosts[best_i] );
    }
}
func Vec_IntSelectSortCost2Reverse(pArray *int32, nSize int32, pCosts *int32) {
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if pCosts[j] > pCosts[best_i] {
                best_i = j; }
        }
        ABC_SWAP( int, pArray[i], pArray[best_i] );
        ABC_SWAP( int, pCosts[i], pCosts[best_i] );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPrint(vVec *Vec_Int_t) {
    var i, Entry int32;
    fmt.Printf("Vector has %d entries: {", Vec_IntSize(vVec));
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, i)), 1); i++  {
        fmt.Printf(" %d", Entry);
    }
    fmt.Printf(" }\n");
}
func Vec_IntPrintBinary(vVec *Vec_Int_t) {
    var i, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, i)), 1); i++  {
        fmt.Printf("%d", int32((Entry != 0)));
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntCompareVec(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    if p1 == nil || p2 == nil {
        return (p1 != nil) - (p2 != nil); }
    if Vec_IntSize(p1) != Vec_IntSize(p2) {
        return Vec_IntSize(p1) - Vec_IntSize(p2); }
    return memcmp( Vec_IntArray(p1), Vec_IntArray(p2), sizeof(int)*(size_t)Vec_IntSize(p1) );
}

/**Function*************************************************************

  Synopsis    [Appends the contents of the second vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAppend(vVec1 *Vec_Int_t, vVec2 *Vec_Int_t) {
    var Entry, i int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVec2)) && (((Entry) = Vec_IntEntry(vVec2, i)), 1); i++  {
        Vec_IntPush( vVec1, Entry );
    }
}
func Vec_IntAppendSkip(vVec1 *Vec_Int_t, vVec2 *Vec_Int_t, iVar int32) {
    var Entry, i int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVec2)) && (((Entry) = Vec_IntEntry(vVec2, i)), 1); i++  {
        if i != iVar {
            Vec_IntPush( vVec1, Entry ); }
    }
}
func Vec_IntAppendMinus(vVec1 *Vec_Int_t, vVec2 *Vec_Int_t, fMinus int32) {
    var Entry, i int32;
    Vec_IntClear( vVec1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVec2)) && (((Entry) = Vec_IntEntry(vVec2, i)), 1); i++  {
        Vec_IntPush( vVec1, fMinus ? -Entry : Entry );
    }
}

/**Function*************************************************************

  Synopsis    [Remapping attributes after objects were duplicated.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntRemapArray(vOld2New *Vec_Int_t, vOld *Vec_Int_t, vNew *Vec_Int_t, nNew int32) {
    var iOld, iNew int32;
    if Vec_IntSize(vOld) == 0 {
        return; }
    Vec_IntFill( vNew, nNew, 0 );
    // inlining of macro Vec_IntForEachEntry
    for  iOld = 0; (iOld < Vec_IntSize(vOld2New)) && (((iNew) = Vec_IntEntry(vOld2New, iOld)), 1); iOld++  {
        if iNew > 0 && iNew < nNew && iOld < Vec_IntSize(vOld) && Vec_IntEntry(vOld, iOld) != 0 {
            Vec_IntWriteEntry( vNew, iNew, Vec_IntEntry(vOld, iOld) ); }
    }
}

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

