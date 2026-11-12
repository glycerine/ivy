package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecFlt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable arrays of floats.]

  Author      [Aaron P. Hurst]
  
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

         
type Vec_Flt_t struct {
    nCap              int32;
    nSize              int32;
    pArray           *float32;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Vec_FltForEachEntry( vVec, Entry, i )                                               \
//     for ( i = 0; (i < Vec_FltSize(vVec)) && (((Entry) = Vec_FltEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_FltForEachEntryStart( vVec, Entry, i, Start )                                   \
//     for ( i = Start; (i < Vec_FltSize(vVec)) && (((Entry) = Vec_FltEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_FltForEachEntryStartStop( vVec, Entry, i, Start, Stop )                         \
//     for ( i = Start; (i < Stop) && (((Entry) = Vec_FltEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_FltForEachEntryReverse( vVec, pEntry, i )                                       \
//     for ( i = Vec_FltSize(vVec) - 1; (i >= 0) && (((pEntry) = Vec_FltEntry(vVec, i)), 1); i-- )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltAlloc(nCap int32) *Vec_Flt_t {
    var p *Vec_Flt_t;
    p = new(Vec_Flt_t);
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( float, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}
func Vec_FltAllocExact(nCap int32) *Vec_Flt_t {
    var p *Vec_Flt_t;
    assert(nCap >= 0, "nCap >= 0");
    p = new(Vec_Flt_t);
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( float, p.nCap );
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
func Vec_FltStart(nSize int32) *Vec_Flt_t {
    var p *Vec_Flt_t;
    p = Vec_FltAlloc( nSize );
    p.nSize = nSize;
    memset( p.pArray, 0, sizeof(float) * (size_t)nSize );
    return p;
}
func Vec_FltStartFull(nSize int32) *Vec_Flt_t {
    var p *Vec_Flt_t;
    p = Vec_FltAlloc( nSize );
    p.nSize = nSize;
    memset( p.pArray, 0xFF, sizeof(float) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from a float array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltAllocArray(pArray *float32, nSize int32) *Vec_Flt_t {
    var p *Vec_Flt_t;
    p = new(Vec_Flt_t);
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = pArray;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from a float array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltAllocArrayCopy(pArray *float32, nSize int32) *Vec_Flt_t {
    var p *Vec_Flt_t;
    p = new(Vec_Flt_t);
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = ABC_ALLOC( float, nSize );
    memcpy( p.pArray, pArray, sizeof(float) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the float array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltDup(pVec *Vec_Flt_t) *Vec_Flt_t {
    var p *Vec_Flt_t;
    p = new(Vec_Flt_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( float, p.nCap );
    } else {
        p.pArray = nil;
    }
    memcpy( p.pArray, pVec.pArray, sizeof(float) * (size_t)pVec.nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Transfers the array into another vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltDupArray(pVec *Vec_Flt_t) *Vec_Flt_t {
    var p *Vec_Flt_t;
    p = new(Vec_Flt_t);
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
func Vec_FltZero(p *Vec_Flt_t) {
    p.pArray = nil;
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_FltErase(p *Vec_Flt_t) {
    ABC_FREE( p.pArray );
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_FltFree(p *Vec_Flt_t) {
    ABC_FREE( p.pArray );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltFreeP(p **Vec_Flt_t) {
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
func Vec_FltReleaseArray(p *Vec_Flt_t) *float32 {
    var pArray *float32  = p.pArray;
    p.nCap = 0;
    p.nSize = 0;
    p.pArray = nil;
    return pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltArray(p *Vec_Flt_t) *float32 {
    return p.pArray;
}
func Vec_FltArrayP(p *Vec_Flt_t) **float32 {
    return &p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltSize(p *Vec_Flt_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltCap(p *Vec_Flt_t) int32 {
    return p.nCap;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltMemory(p *Vec_Flt_t) float64 {
    if !(p != nil) {
        return 0.0;
    } else {
        return 1.0 * sizeof(float) * (size_t)p.nCap + sizeof(Vec_Flt_t);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltEntry(p *Vec_Flt_t, i int32) float32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}
func Vec_FltEntryP(p *Vec_Flt_t, i int32) *float32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray + i;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltWriteEntry(p *Vec_Flt_t, i int32, Entry float32) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltAddToEntry(p *Vec_Flt_t, i int32, Addition float32) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] += Addition;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltUpdateEntry(p *Vec_Flt_t, i int32, Value float32) {
    if Vec_FltEntry( p, i ) < Value {
        Vec_FltWriteEntry( p, i, Value ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltEntryLast(p *Vec_Flt_t) float32 {
    return p.pArray[p.nSize-1];
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltGrow(p *Vec_Flt_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pArray = ABC_REALLOC( float, p.pArray, nCapMin ); 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltFill(p *Vec_Flt_t, nSize int32, Entry float32) {
    var i int32;
    Vec_FltGrow( p, nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = Entry;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltFillExtra(p *Vec_Flt_t, nSize int32, Fill float32) {
    var i int32;
    if nSize <= p.nSize {
        return; }
    if nSize > 2 * p.nCap {
        Vec_FltGrow( p, nSize );
    } else if nSize > p.nCap {
        Vec_FltGrow( p, 2 * p.nCap ); }
    for  i = p.nSize; i < nSize; i++  {
        p.pArray[i] = Fill;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltShrink(p *Vec_Flt_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltClear(p *Vec_Flt_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltPush(p *Vec_Flt_t, Entry float32) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_FltGrow( p, 16 );
        } else {
            Vec_FltGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltPushOrder(p *Vec_Flt_t, Entry float32) {
    var i int32;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_FltGrow( p, 16 );
        } else {
            Vec_FltGrow( p, 2 * p.nCap ); }
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

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltPushUnique(p *Vec_Flt_t, Entry float32) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Vec_FltPush( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the last entry and removes it from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltPop(p *Vec_Flt_t) float32 {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    [Find entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltFind(p *Vec_Flt_t, Entry float32) int32 {
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
func Vec_FltRemove(p *Vec_Flt_t, Entry float32) int32 {
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

/**Function*************************************************************

  Synopsis    [Find entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltFindMax(p *Vec_Flt_t) float32 {
    var i int32;
    var Best float32;
    if p.nSize == 0 {
        return 0; }
    Best = p.pArray[0];
    for  i = 1; i < p.nSize; i++  {
        if Best < p.pArray[i] {
            Best = p.pArray[i]; }
    }
    return Best;
}
func Vec_FltFindMin(p *Vec_Flt_t) float32 {
    var i int32;
    var Best float32;
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

  Synopsis    [Checks if two vectors are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltEqual(p1 *Vec_Flt_t, p2 *Vec_Flt_t) int32 {
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

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltPrint(vVec *Vec_Flt_t) {
    var i int32; var Entry float32;
    fmt.Printf("Vector has %d entries: {", Vec_FltSize(vVec));
    // inlining of macro Vec_FltForEachEntry
    for  i = 0; (i < Vec_FltSize(vVec)) && (((Entry) = Vec_FltEntry(vVec, i)), 1); i++  {
        fmt.Printf(" %f", Entry);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two floats.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltSortCompare1(pp1 *float32, pp2 *float32) int32 {
    // for some reason commenting out lines (as shown) led to crashing of the release version
    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 { //
        return 1; }
    return 0; //
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two floats.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltSortCompare2(pp1 *float32, pp2 *float32) int32 {
    // for some reason commenting out lines (as shown) led to crashing of the release version
    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 { //
        return 1; }
    return 0; //
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_FltSort(p *Vec_Flt_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(float), 
                        (int (*)(void *, void *)) Vec_FltSortCompare2 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(float), 
                        (int (*)(void *, void *)) Vec_FltSortCompare1 ); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_END



