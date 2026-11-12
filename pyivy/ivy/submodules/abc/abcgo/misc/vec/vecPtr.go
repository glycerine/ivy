package vec

import (
	"fmt"
	"strings"
)

/**CFile****************************************************************

  FileName    [vecPtr.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable arrays of generic pointers.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecPtr.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

         
type Vec_Ptr_t struct {
    nCap              int32;
    nSize              int32;
    pArray           *any;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// iterators through entries
// Original macro:
// #define Vec_PtrForEachEntry( Type, vVec, pEntry, i )                                               \
//     for ( i = 0; (i < Vec_PtrSize(vVec)) && (((pEntry) = (Type)Vec_PtrEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_PtrForEachEntryStart( Type, vVec, pEntry, i, Start )                                   \
//     for ( i = Start; (i < Vec_PtrSize(vVec)) && (((pEntry) = (Type)Vec_PtrEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_PtrForEachEntryStop( Type, vVec, pEntry, i, Stop )                                     \
//     for ( i = 0; (i < Stop) && (((pEntry) = (Type)Vec_PtrEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_PtrForEachEntryStartStop( Type, vVec, pEntry, i, Start, Stop )                         \
//     for ( i = Start; (i < Stop) && (((pEntry) = (Type)Vec_PtrEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_PtrForEachEntryReverse( Type, vVec, pEntry, i )                                        \
//     for ( i = Vec_PtrSize(vVec) - 1; (i >= 0) && (((pEntry) = (Type)Vec_PtrEntry(vVec, i)), 1); i-- )

// Original macro:
// #define Vec_PtrForEachEntryTwo( Type1, vVec1, Type2, vVec2, pEntry1, pEntry2, i )                  \
//     for ( i = 0; (i < Vec_PtrSize(vVec1)) && (((pEntry1) = (Type1)Vec_PtrEntry(vVec1, i)), 1) && (((pEntry2) = (Type2)Vec_PtrEntry(vVec2, i)), 1); i++ )

// Original macro:
// #define Vec_PtrForEachEntryDouble( Type1, Type2, vVec, Entry1, Entry2, i )                                \
//     for ( i = 0; (i+1 < Vec_PtrSize(vVec)) && (((Entry1) = (Type1)Vec_PtrEntry(vVec, i)), 1) && (((Entry2) = (Type2)Vec_PtrEntry(vVec, i+1)), 1); i += 2 )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrAlloc(nCap int32) *Vec_Ptr_t {
    var p *Vec_Ptr_t;
    p = new(Vec_Ptr_t);
    if nCap > 0 && nCap < 8 {
        nCap = 8; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( void *, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}
func Vec_PtrAllocExact(nCap int32) *Vec_Ptr_t {
    var p *Vec_Ptr_t;
    assert(nCap >= 0, "nCap >= 0");
    p = new(Vec_Ptr_t);
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( void *, p.nCap );
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
func Vec_PtrStart(nSize int32) *Vec_Ptr_t {
    var p *Vec_Ptr_t;
    p = Vec_PtrAlloc( nSize );
    p.nSize = nSize;
    memset( p.pArray, 0, sizeof(void *) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrAllocArray(pArray *any, nSize int32) *Vec_Ptr_t {
    var p *Vec_Ptr_t;
    p = new(Vec_Ptr_t);
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
func Vec_PtrAllocArrayCopy(pArray *any, nSize int32) *Vec_Ptr_t {
    var p *Vec_Ptr_t;
    p = new(Vec_Ptr_t);
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = ABC_ALLOC( void *, nSize );
    memcpy( p.pArray, pArray, sizeof(void *) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrDup(pVec *Vec_Ptr_t) *Vec_Ptr_t {
    var p *Vec_Ptr_t;
    p = new(Vec_Ptr_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( void *, p.nCap );
    } else {
        p.pArray = nil;
    }
    memcpy( p.pArray, pVec.pArray, sizeof(void *) * (size_t)pVec.nSize );
    return p;
}
func Vec_PtrDupStr(pVec *Vec_Ptr_t) *Vec_Ptr_t {
    var i int32;
    var p *Vec_Ptr_t  = Vec_PtrDup( pVec );
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] = Abc_UtilStrsav( (char *)p.pArray[i] );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Transfers the array into another vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrDupArray(pVec *Vec_Ptr_t) *Vec_Ptr_t {
    var p *Vec_Ptr_t;
    p = new(Vec_Ptr_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    p.pArray = pVec.pArray;
    pVec.nSize  = 0;
    pVec.nCap   = 0;
    pVec.pArray = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    [Frees the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrZero(p *Vec_Ptr_t) {
    p.pArray = nil;
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_PtrErase(p *Vec_Ptr_t) {
    ABC_FREE( p.pArray );
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_PtrFree(p *Vec_Ptr_t) {
    ABC_FREE( p.pArray );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrFreeP(p **Vec_Ptr_t) {
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
func Vec_PtrReleaseArray(p *Vec_Ptr_t) *any {
    var pArray *any  = p.pArray;
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
func Vec_PtrArray(p *Vec_Ptr_t) *any {
    return p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrSize(p *Vec_Ptr_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrCap(p *Vec_Ptr_t) int32 {
    return p.nCap;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrMemory(p *Vec_Ptr_t) float64 {
    if !(p != nil) {
        return 0.0;
    } else {
        return 1.0 * sizeof(void *) * (size_t)p.nCap + sizeof(Vec_Ptr_t);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrCountZero(p *Vec_Ptr_t) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] == nil);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrEntry(p *Vec_Ptr_t, i int32) any {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrEntryP(p *Vec_Ptr_t, i int32) *any {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray + i;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrWriteEntry(p *Vec_Ptr_t, i int32, Entry any) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrEntryLast(p *Vec_Ptr_t) any {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[p.nSize-1];
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrGrow(p *Vec_Ptr_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pArray = ABC_REALLOC( void *, p.pArray, nCapMin ); 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrFill(p *Vec_Ptr_t, nSize int32, Entry any) {
    var i int32;
    Vec_PtrGrow( p, nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = Entry;
    }
    p.nSize = nSize;
}
func Vec_PtrFillTwo(p *Vec_Ptr_t, nSize int32, EntryEven any, EntryOdd any) {
    var i int32;
    Vec_PtrGrow( p, nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = (i & 1) ? EntryOdd : EntryEven;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrFillExtra(p *Vec_Ptr_t, nSize int32, Fill any) {
    var i int32;
    if nSize <= p.nSize {
        return; }
    if nSize > 2 * p.nCap {
        Vec_PtrGrow( p, nSize );
    } else if nSize > p.nCap {
        Vec_PtrGrow( p, 2 * p.nCap ); }
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
func Vec_PtrGetEntry(p *Vec_Ptr_t, i int32) any {
    Vec_PtrFillExtra( p, i + 1, nil );
    return Vec_PtrEntry( p, i );
}

/**Function*************************************************************

  Synopsis    [Inserts the entry even if the place does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrSetEntry(p *Vec_Ptr_t, i int32, Entry any) {
    Vec_PtrFillExtra( p, i + 1, nil );
    Vec_PtrWriteEntry( p, i, Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrShrink(p *Vec_Ptr_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrClear(p *Vec_Ptr_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    [Deallocates array of memory pointers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrFreeData(p *Vec_Ptr_t) {
    var pTemp any; var i int32;
    if p == nil { return; }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p)) && (((pTemp) = (void *)Vec_PtrEntry(p, i)), 1); i++  {
        if pTemp != (void *)(ABC_PTRINT_T)1 && pTemp != (void *)(ABC_PTRINT_T)2 {
            ABC_FREE( pTemp ); }
    }
}
func Vec_PtrFreeFree(p *Vec_Ptr_t) {
    if p == nil { return; }
    Vec_PtrFreeData( p );
    Vec_PtrFree( p );
}

/**Function*************************************************************

  Synopsis    [Copies the interger array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrCopy(pDest *Vec_Ptr_t, pSour *Vec_Ptr_t) {
    pDest.nSize = 0;
    Vec_PtrGrow( pDest, pSour.nSize );
    memcpy( pDest.pArray, pSour.pArray, sizeof(void *) * (size_t)pSour.nSize );
    pDest.nSize = pSour.nSize;
}

/**Function*************************************************************

  Synopsis    [Print names stored in the array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrPrintNames(p *Vec_Ptr_t) {
    var pName string; var i int32;
    fmt.Printf("Vector has %d entries: {", Vec_PtrSize(p));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p)) && (((pName) = (char *)Vec_PtrEntry(p, i)), 1); i++  {
        fmt.Printf("%s ", pName);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrPush(p *Vec_Ptr_t, Entry any) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_PtrGrow( p, 16 );
        } else {
            Vec_PtrGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}
func Vec_PtrPushTwo(p *Vec_Ptr_t, Entry1 any, Entry2 any) {
    Vec_PtrPush( p, Entry1 );
    Vec_PtrPush( p, Entry2 );
}
func Vec_PtrAppend(vVec1 *Vec_Ptr_t, vVec2 *Vec_Ptr_t) {
    var Entry any; var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vVec2)) && (((Entry) = (void *)Vec_PtrEntry(vVec2, i)), 1); i++  {
        Vec_PtrPush( vVec1, Entry );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrPushFirst(p *Vec_Ptr_t, Entry any) {
    var i int32;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_PtrGrow( p, 16 );
        } else {
            Vec_PtrGrow( p, 2 * p.nCap ); }
    }
    p.nSize++;
    for  i = p.nSize - 1; i >= 1; i--  {
        p.pArray[i] = p.pArray[i-1];
    }
    p.pArray[0] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrPushUnique(p *Vec_Ptr_t, Entry any) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Vec_PtrPush( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the last entry and removes it from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrPop(p *Vec_Ptr_t) any {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    [Find entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrFind(p *Vec_Ptr_t, Entry any) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return i; }
    }
    return -1;
}
func Vec_PtrFindStr(p *Vec_Ptr_t, Entry string) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] && !int32(strings.Compare((char *)p.pArray[i], Entry)) {
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
func Vec_PtrRemove(p *Vec_Ptr_t, Entry any) {
    var i int32;
    // delete assuming that it is closer to the end
    for  i = p.nSize - 1; i >= 0; i--  {
        if p.pArray[i] == Entry {
            break; }
    }
    assert(i >= 0, "i >= 0");
/*
    // delete assuming that it is closer to the beginning
    for ( i = 0; i < p->nSize; i++ )
        if ( p->pArray[i] == Entry )
            break;
    assert( i < p->nSize );
*/
    for  i++; i < p.nSize; i++  {
        p.pArray[i-1] = p.pArray[i];
    }
    p.nSize--;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrDrop(p *Vec_Ptr_t, i int32) {
    var k int32;
    assert(i >= 0 && i < Vec_PtrSize(p), "i >= 0 && i < Vec_PtrSize(p)");
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
func Vec_PtrInsert(p *Vec_Ptr_t, iHere int32, Entry any) {
    var i int32;
    assert(iHere >= 0 && iHere < p.nSize, "iHere >= 0 && iHere < p.nSize");
    Vec_PtrPush( p, 0 );
    for  i = p.nSize - 1; i > iHere; i--  {
        p.pArray[i] = p.pArray[i-1];
    }
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    [Moves the first nItems to the end.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrReorder(p *Vec_Ptr_t, nItems int32) {
    assert(nItems < p.nSize, "nItems < p.nSize");
    Vec_PtrGrow( p, nItems + p.nSize );
    memmove( (char **)p.pArray + p.nSize, p.pArray, (size_t)nItems * sizeof(void*) );
    memmove( p.pArray, (char **)p.pArray + nItems, (size_t)p.nSize * sizeof(void*) );
}

/**Function*************************************************************

  Synopsis    [Reverses the order of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrReverseOrder(p *Vec_Ptr_t) {
    var Temp any;
    var i int32;
    for  i = 0; i < p.nSize/2; i++  {
        Temp = p.pArray[i];
        p.pArray[i] = p.pArray[p.nSize-1-i];
        p.pArray[p.nSize-1-i] = Temp;
    }
}

/**Function*************************************************************

  Synopsis    [Checks if two vectors are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrEqual(p1 *Vec_Ptr_t, p2 *Vec_Ptr_t) int32 {
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

  Synopsis    [Comparison procedure for two integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrSortComparePtr(pp1 *any, pp2 *any) int32 {
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
static void Vec_PtrSort( Vec_Ptr_t * p, int (*Vec_PtrSortCompare)() ) ___unused;
func Vec_PtrSort(p *Vec_Ptr_t, Vec_PtrSortCompare func() int32) {
    if p.nSize < 2 {
        return; }
    if Vec_PtrSortCompare == nil {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(void *), 
                        (int (*)(void *, void *)) Vec_PtrSortComparePtr );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(void *), 
                        (int (*)(void *, void *)) Vec_PtrSortCompare ); }
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static void Vec_PtrUniqify( Vec_Ptr_t * p, int (*Vec_PtrSortCompare)() ) ___unused;
func Vec_PtrUniqify(p *Vec_Ptr_t, Vec_PtrSortCompare func() int32) {
    var i, k int32;
    if p.nSize < 2 {
        return; }
    Vec_PtrSort( p, Vec_PtrSortCompare );
    for  i = k = 1; i < p.nSize; i++  {
        if p.pArray[i] != p.pArray[i-1] {
            p.pArray[k++] = p.pArray[i]; }
    }
    p.nSize = k;
}
static void Vec_PtrUniqify2( Vec_Ptr_t * p, int (*Vec_PtrSortCompare)(void**, void**), void (*Vec_PtrObjFree)(void*), Vec_Int_t * vCounts ) ___unused;
func Vec_PtrUniqify2(p *Vec_Ptr_t, Vec_PtrSortCompare func(*any, *any) int32, Vec_PtrObjFree func(any), vCounts *Vec_Int_t) {
    var i, k int32;
    if vCounts != nil {
        Vec_IntFill( vCounts, 1, 1 ); }
    if p.nSize < 2 {
        return; }
    Vec_PtrSort( p, (int (*)())Vec_PtrSortCompare );
    for  i = k = 1; i < p.nSize; i++  {
        if Vec_PtrSortCompare(p.pArray+i, p.pArray+k-1) != 0 {
            p.pArray[k++] = p.pArray[i];
            if vCounts != nil {
                Vec_IntPush( vCounts, 1 ); }
        
        } else {
            if Vec_PtrObjFree {
                Vec_PtrObjFree( p.pArray[i] ); }
            if vCounts != nil {
                Vec_IntAddToEntry( vCounts, Vec_IntSize(vCounts)-1, 1 ); }
        }
    }
    p.nSize = k;
    assert(vCounts == nil || Vec_IntSize(vCounts) == Vec_PtrSize(p), "vCounts == nil || Vec_IntSize(vCounts) == Vec_PtrSize(p)");
}



/**Function*************************************************************

  Synopsis    [Allocates the array of simulation info.]

  Description [Allocates the array containing given number of entries, 
  each of which contains given number of unsigned words of simulation data.
  The resulting array can be freed using regular procedure Vec_PtrFree().
  It is the responsibility of the user to ensure this array is never grown.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrAllocSimInfo(nEntries int32, nWords int32) *Vec_Ptr_t {
    var pMemory *any;
    var pInfo *uint32;
    var i int32;
    pMemory = (void **)ABC_ALLOC( char, (sizeof(void *) + sizeof(unsigned) * (size_t)nWords) * nEntries );
    pInfo = (unsigned *)(pMemory + nEntries);
    for  i = 0; i < nEntries; i++  {
        pMemory[i] = pInfo + i * nWords;
    }
    return Vec_PtrAllocArray( pMemory, nEntries );
}

/**Function*************************************************************

  Synopsis    [Cleans simulation info of each entry beginning with iWord.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrReadWordsSimInfo(p *Vec_Ptr_t) int32 {
    return (unsigned *)Vec_PtrEntry(p,1) - (unsigned *)Vec_PtrEntry(p,0);
}

/**Function*************************************************************

  Synopsis    [Cleans simulation info of each entry beginning with iWord.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrCleanSimInfo(vInfo *Vec_Ptr_t, iWord int32, nWords int32) {
    var i int32;
    for  i = 0; i < vInfo.nSize; i++  {
        memset( (char*)Vec_PtrEntry(vInfo,i) + 4*iWord, 0, (size_t)(4*(nWords-iWord)) );
    }
}

/**Function*************************************************************

  Synopsis    [Cleans simulation info of each entry beginning with iWord.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrFillSimInfo(vInfo *Vec_Ptr_t, iWord int32, nWords int32) {
    var i int32;
    for  i = 0; i < vInfo.nSize; i++  {
        memset( (char*)Vec_PtrEntry(vInfo,i) + 4*iWord, 0xFF, (size_t)(4*(nWords-iWord)) );
    }
}

/**Function*************************************************************

  Synopsis    [Resizes the array of simulation info.]

  Description [Doubles the number of objects for which siminfo is allocated.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrDoubleSimInfo(vInfo *Vec_Ptr_t) {
    var vInfoNew *Vec_Ptr_t;
    var nWords int32;
    assert(Vec_PtrSize(vInfo) > 1, "Vec_PtrSize(vInfo) > 1");
    // get the new array
    nWords = (unsigned *)Vec_PtrEntry(vInfo,1) - (unsigned *)Vec_PtrEntry(vInfo,0);
    vInfoNew = Vec_PtrAllocSimInfo( 2*Vec_PtrSize(vInfo), nWords );
    // copy the simulation info
    memcpy( Vec_PtrEntry(vInfoNew,0), Vec_PtrEntry(vInfo,0), (size_t)(Vec_PtrSize(vInfo) * nWords * 4) );
    // replace the array
    ABC_FREE( vInfo.pArray );
    vInfo.pArray = vInfoNew.pArray;
    vInfo.nSize *= 2;
    vInfo.nCap *= 2;
    // free the old array
    vInfoNew.pArray = nil;
    ABC_FREE( vInfoNew );
}

/**Function*************************************************************

  Synopsis    [Resizes the array of simulation info.]

  Description [Doubles the number of simulation patterns stored for each object.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrReallocSimInfo(vInfo *Vec_Ptr_t) {
    var vInfoNew *Vec_Ptr_t;
    var nWords, i int32;
    assert(Vec_PtrSize(vInfo) > 1, "Vec_PtrSize(vInfo) > 1");
    // get the new array
    nWords = (unsigned *)Vec_PtrEntry(vInfo,1) - (unsigned *)Vec_PtrEntry(vInfo,0);
    vInfoNew = Vec_PtrAllocSimInfo( Vec_PtrSize(vInfo), 2*nWords );
    // copy the simulation info
    for  i = 0; i < vInfo.nSize; i++  {
        memcpy( Vec_PtrEntry(vInfoNew,i), Vec_PtrEntry(vInfo,i), (size_t)(nWords * 4) );
    }
    // replace the array
    ABC_FREE( vInfo.pArray );
    vInfo.pArray = vInfoNew.pArray;
    // free the old array
    vInfoNew.pArray = nil;
    ABC_FREE( vInfoNew );
}

/**Function*************************************************************

  Synopsis    [Allocates the array of truth tables for the given number of vars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_PtrAllocTruthTables(nVars int32) *Vec_Ptr_t {
    var p *Vec_Ptr_t;
    var Masks [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
    var pTruth *uint32;
    var i, k, nWords int32;
    if nVars <= 5 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 5));
    }
    p = Vec_PtrAllocSimInfo( nVars, nWords );
    for  i = 0; i < nVars; i++  {
        pTruth = (unsigned *)p.pArray[i];
        if i < 5 {
            for  k = 0; k < nWords; k++  {
                pTruth[k] = Masks[i];
            }
        
        } else {
            for  k = 0; k < nWords; k++  {
                if k & (1 << (i-5)) {
                    pTruth[k] = ~uint32(0);
                } else {
                    pTruth[k] = 0; }
            }
        }
    }
    return p;
}



//ABC_NAMESPACE_HEADER_END




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

