package vec

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [vecWrd.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable arrays of long unsigned integers.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecWrd.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

         
type Vec_Wrd_t struct {
    nCap              int32;
    nSize              int32;
    pArray            *word;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Vec_WrdForEachEntry( vVec, Entry, i )                                               \
//     for ( i = 0; (i < Vec_WrdSize(vVec)) && (((Entry) = Vec_WrdEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_WrdForEachEntryStart( vVec, Entry, i, Start )                                   \
//     for ( i = Start; (i < Vec_WrdSize(vVec)) && (((Entry) = Vec_WrdEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_WrdForEachEntryStop( vVec, Entry, i, Stop )                                   \
//     for ( i = 0; (i < Stop) && (((Entry) = Vec_WrdEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_WrdForEachEntryStartStop( vVec, Entry, i, Start, Stop )                         \
//     for ( i = Start; (i < Stop) && (((Entry) = Vec_WrdEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_WrdForEachEntryReverse( vVec, pEntry, i )                                       \
//     for ( i = Vec_WrdSize(vVec) - 1; (i >= 0) && (((pEntry) = Vec_WrdEntry(vVec, i)), 1); i-- )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdAlloc(nCap int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    p = new(Vec_Wrd_t);
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( word, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}
func Vec_WrdAllocExact(nCap int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    assert(nCap >= 0, "nCap >= 0");
    p = new(Vec_Wrd_t);
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( word, p.nCap );
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
func Vec_WrdStart(nSize int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    p = Vec_WrdAlloc( nSize );
    p.nSize = nSize;
    memset( p.pArray, 0, sizeof(word) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given size and cleans it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdStartFull(nSize int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    p = Vec_WrdAlloc( nSize );
    p.nSize = nSize;
    memset( p.pArray, 0xff, sizeof(word) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given size and cleans it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdStartNatural(nSize int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    var i int32;
    p = Vec_WrdAlloc( nSize );
    p.nSize = nSize;
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = i;
    }
    return p;
}
func Vec_WrdStartRandom(nSize int32) *Vec_Wrd_t {
    var vSims *Vec_Wrd_t  = Vec_WrdStart( nSize ); var i int32;
    for  i = 0; i < nSize; i++  {
        vSims.pArray[i] = Abc_RandomW(0);
    }
    return vSims;
}
func Vec_WrdStartTruthTables(nVars int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    var Masks [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
    var i, k, nWords int32;
    if nVars <= 6 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 6));
    }
    p = Vec_WrdStart( nWords * nVars );
    for  i = 0; i < nVars; i++  {
        var pTruth *uint32  = (unsigned *)(p.pArray + nWords * i);
        if i < 5 {
            for  k = 0; k < 2*nWords; k++  {
                pTruth[k] = Masks[i];
            }
        
        } else {
            for  k = 0; k < 2*nWords; k++  {
                if k & (1 << (i-5)) {
                    pTruth[k] = ~uint32(0);
                } else {
                    pTruth[k] = 0; }
            }
        }
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdAllocArray(pArray *word, nSize int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    p = new(Vec_Wrd_t);
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
func Vec_WrdAllocArrayCopy(pArray *word, nSize int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    p = new(Vec_Wrd_t);
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = ABC_ALLOC( word, nSize );
    memcpy( p.pArray, pArray, sizeof(word) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdDup(pVec *Vec_Wrd_t) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    p = new(Vec_Wrd_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nSize;
    if p.nCap {
        p.pArray = ABC_ALLOC( word, p.nCap );
    } else {
        p.pArray = nil;
    }
    memcpy( p.pArray, pVec.pArray, sizeof(word) * (size_t)pVec.nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Transfers the array into another vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdDupArray(pVec *Vec_Wrd_t) *Vec_Wrd_t {
    var p *Vec_Wrd_t;
    p = new(Vec_Wrd_t);
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
func Vec_WrdErase(p *Vec_Wrd_t) {
    ABC_FREE( p.pArray );
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_WrdFree(p *Vec_Wrd_t) {
    ABC_FREE( p.pArray );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdFreeP(p **Vec_Wrd_t) {
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
func Vec_WrdReleaseArray(p *Vec_Wrd_t) *word {
    var pArray *word  = p.pArray;
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
func Vec_WrdArray(p *Vec_Wrd_t) *word {
    return p.pArray;
}
func Vec_WrdLimit(p *Vec_Wrd_t) *word {
    return p.pArray + p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdSize(p *Vec_Wrd_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdCap(p *Vec_Wrd_t) int32 {
    return p.nCap;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdMemory(p *Vec_Wrd_t) float64 {
    if !(p != nil) {
        return 0.0;
    } else {
        return 1.0 * sizeof(word) * (size_t)p.nCap + sizeof(Vec_Wrd_t);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdEntry(p *Vec_Wrd_t, i int32) word {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdEntryP(p *Vec_Wrd_t, i int32) *word {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray + i;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdWriteEntry(p *Vec_Wrd_t, i int32, Entry word) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdAddToEntry(p *Vec_Wrd_t, i int32, Addition word) word {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i] += Addition;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdEntryLast(p *Vec_Wrd_t) word {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[p.nSize-1];
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdGrow(p *Vec_Wrd_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pArray = ABC_REALLOC( word, p.pArray, nCapMin ); 
    assert(p.pArray, "p.pArray");
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdFill(p *Vec_Wrd_t, nSize int32, Fill word) {
    var i int32;
    Vec_WrdGrow( p, nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = Fill;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdFillExtra(p *Vec_Wrd_t, nSize int32, Fill word) {
    var i int32;
    if nSize <= p.nSize {
        return; }
    if nSize > 2 * p.nCap {
        Vec_WrdGrow( p, nSize );
    } else if nSize > p.nCap {
        Vec_WrdGrow( p, 2 * p.nCap ); }
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
func Vec_WrdGetEntry(p *Vec_Wrd_t, i int32) word {
    Vec_WrdFillExtra( p, i + 1, 0 );
    return Vec_WrdEntry( p, i );
}

/**Function*************************************************************

  Synopsis    [Returns the entry even if the place not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdGetEntryP(p *Vec_Wrd_t, i int32) *word {
    Vec_WrdFillExtra( p, i + 1, 0 );
    return Vec_WrdEntryP( p, i );
}

/**Function*************************************************************

  Synopsis    [Inserts the entry even if the place does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdSetEntry(p *Vec_Wrd_t, i int32, Entry word) {
    Vec_WrdFillExtra( p, i + 1, 0 );
    Vec_WrdWriteEntry( p, i, Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdShrink(p *Vec_Wrd_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdClear(p *Vec_Wrd_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdPush(p *Vec_Wrd_t, Entry word) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_WrdGrow( p, 16 );
        } else {
            Vec_WrdGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdPushFirst(p *Vec_Wrd_t, Entry word) {
    var i int32;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_WrdGrow( p, 16 );
        } else {
            Vec_WrdGrow( p, 2 * p.nCap ); }
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
func Vec_WrdPushOrder(p *Vec_Wrd_t, Entry word) {
    var i int32;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_WrdGrow( p, 16 );
        } else {
            Vec_WrdGrow( p, 2 * p.nCap ); }
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

  Synopsis    [Inserts the entry while preserving the increasing order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdPushUniqueOrder(p *Vec_Wrd_t, Entry word) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Vec_WrdPushOrder( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdPushUnique(p *Vec_Wrd_t, Entry word) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Vec_WrdPush( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the next nWords entries in the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdFetch(p *Vec_Wrd_t, nWords int32) *word {
    if nWords == 0 {
        return nil; }
    assert(nWords > 0, "nWords > 0");
    p.nSize += nWords;
    if p.nSize > p.nCap {
//         Vec_WrdGrow( p, 2 * p->nSize );
        return nil;
    }
    return p.pArray + p.nSize - nWords;
}

/**Function*************************************************************

  Synopsis    [Returns the last entry and removes it from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdPop(p *Vec_Wrd_t) word {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    [Find entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdFind(p *Vec_Wrd_t, Entry word) int32 {
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
func Vec_WrdRemove(p *Vec_Wrd_t, Entry word) int32 {
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

  Synopsis    [Interts entry at the index iHere. Shifts other entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdInsert(p *Vec_Wrd_t, iHere int32, Entry word) {
    var i int32;
    assert(iHere >= 0 && iHere < p.nSize, "iHere >= 0 && iHere < p.nSize");
    Vec_WrdPush( p, 0 );
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
func Vec_WrdFindMax(p *Vec_Wrd_t) word {
    var Best word;
    var i int32;
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
func Vec_WrdFindMin(p *Vec_Wrd_t) word {
    var Best word;
    var i int32;
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
func Vec_WrdReverseOrder(p *Vec_Wrd_t) {
    var Temp word;
    var i int32;
    for  i = 0; i < p.nSize/2; i++  {
        Temp = p.pArray[i];
        p.pArray[i] = p.pArray[p.nSize-1-i];
        p.pArray[p.nSize-1-i] = Temp;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdInvert(p *Vec_Wrd_t, Fill word) *Vec_Wrd_t {
    var i int32;
    var Entry word;
    var vRes *Vec_Wrd_t  = Vec_WrdAlloc( 0 );
    Vec_WrdFill( vRes, Vec_WrdFindMax(p) + 1, Fill );
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(p)) && (((Entry) = Vec_WrdEntry(p, i)), 1); i++  {
        if Entry != Fill {
            Vec_WrdWriteEntry( vRes, Entry, i ); }
    }
    return vRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdSum(p *Vec_Wrd_t) word {
    var Counter word;
    var i int32;
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
func Vec_WrdCountZero(p *Vec_Wrd_t) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] == 0);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Checks if two vectors are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdEqual(p1 *Vec_Wrd_t, p2 *Vec_Wrd_t) int32 {
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

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdCountCommon(p1 *Vec_Wrd_t, p2 *Vec_Wrd_t) int32 {
    var vTemp *Vec_Wrd_t;
    var Entry word;
    var i, Counter int32;
    if Vec_WrdSize(p1) < Vec_WrdSize(p2) {
        vTemp = p1;
p1 = p2;
p2 = vTemp; }
    assert(Vec_WrdSize(p1) >= Vec_WrdSize(p2), "Vec_WrdSize(p1) >= Vec_WrdSize(p2)");
    vTemp = Vec_WrdInvert( p2, -1 );
    Vec_WrdFillExtra( vTemp, Vec_WrdFindMax(p1) + 1, ~((word)0) );
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(p1)) && (((Entry) = Vec_WrdEntry(p1, i)), 1); i++  {
        if Vec_WrdEntry(vTemp, Entry) != ~((word)0) {
            Counter++; }
    }
    Vec_WrdFree( vTemp );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdSortCompare1(pp1 *word, pp2 *word) int32 {
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
func Vec_WrdSortCompare2(pp1 *word, pp2 *word) int32 {
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
func Vec_WrdSort(p *Vec_Wrd_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(word), 
                        (int (*)(void *, void *)) Vec_WrdSortCompare2 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(word), 
                        (int (*)(void *, void *)) Vec_WrdSortCompare1 ); }
}

/**Function*************************************************************

  Synopsis    [Leaves only unique entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdUniqify(p *Vec_Wrd_t) {
    var i, k int32;
    if p.nSize < 2 {
        return; }
    Vec_WrdSort( p, 0 );
    for  i = k = 1; i < p.nSize; i++  {
        if p.pArray[i] != p.pArray[i-1] {
            p.pArray[k++] = p.pArray[i]; }
    }
    p.nSize = k;
}
func Vec_WrdUniqueCount(vData *Vec_Wrd_t, nWordSize int32, pvMap **Vec_Int_t) int32 {
    var Result int32;
    var vDataInt *Vec_Int_t  = (Vec_Int_t *)vData;
    vDataInt.nSize *= 2;
    vDataInt.nCap *= 2;
    Result = Vec_IntUniqueCount( vDataInt, 2 * nWordSize, pvMap );
    vDataInt.nSize /= 2;
    vDataInt.nCap /= 2;
    return Result;
}
func Vec_WrdUniqifyHash(vData *Vec_Wrd_t, nWordSize int32) *Vec_Wrd_t {
    var vResInt *Vec_Int_t;
    var vDataInt *Vec_Int_t  = (Vec_Int_t *)vData;
    vDataInt.nSize *= 2;
    vDataInt.nCap *= 2;
    vResInt = Vec_IntUniqifyHash( vDataInt, 2 * nWordSize );
    vDataInt.nSize /= 2;
    vDataInt.nCap /= 2;
    vResInt.nSize /= 2;
    vResInt.nCap /= 2;
    return (Vec_Wrd_t *)vResInt;
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdSortCompareUnsigned(pp1 *word, pp2 *word) int32 {
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
func Vec_WrdSortUnsigned(p *Vec_Wrd_t) {
    qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(word), 
                (int (*)(void *, void *)) Vec_WrdSortCompareUnsigned );
}


/**Function*************************************************************

  Synopsis    [Appends the contents of the second vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdAppend(vVec1 *Vec_Wrd_t, vVec2 *Vec_Wrd_t) {
    var Entry word; var i int32;
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(vVec2)) && (((Entry) = Vec_WrdEntry(vVec2, i)), 1); i++  {
        Vec_WrdPush( vVec1, Entry );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimPatWriteOne(pFile *os.File, pSim *word, nWords int32) {
    var k int32;
    var Digit int32;
    var nDigits int32 = nWords*16;
    for  k = 0; k < nDigits; k++  {
        Digit = int32(((pSim[k/16] >> ((k%16) * 4)) & 15));
        if Digit < 10 {
            fmt.Fprintf(pFile, "%d", Digit);
        } else {
            fmt.Fprintf(pFile, "%c", 'A' + Digit-10); }
    }
    fmt.Fprintf(pFile, "\n");
}
func Vec_WrdPrintHex(p *Vec_Wrd_t, nWords int32) {
    var i int32;
    var nNodes int32 = Vec_WrdSize(p) / nWords;
    assert(Vec_WrdSize(p) % nWords == 0, "Vec_WrdSize(p) % nWords == 0");
    for  i = 0; i < nNodes; i++  {
        Gia_ManSimPatWriteOne( os.Stdout, Vec_WrdEntryP(p, i*nWords), nWords );
    }
}
func Vec_WrdDumpHex(pFileName string, p *Vec_Wrd_t, nWords int32, fVerbose int32) {
    var i int32;
    var nNodes int32 = Vec_WrdSize(p) / nWords;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    assert(Vec_WrdSize(p) % nWords == 0, "Vec_WrdSize(p) % nWords == 0");
    for  i = 0; i < nNodes; i++  {
        Gia_ManSimPatWriteOne( pFile, Vec_WrdEntryP(p, i*nWords), nWords );
    }
    _ = (pFile).Close();
    if fVerbose != 0 {
        fmt.Printf("Written %d words of simulation data for %d objects into file \"%s\".\n", nWords, Vec_WrdSize(p)/nWords, pFileName); }
}
func Vec_WrdReadHexOne(c byte) int32 {
    var Digit int32;
    if c >= '0' && c <= '9' {
        Digit = c - '0';
    } else if c >= 'A' && c <= 'F' {
        Digit = c - 'A' + 10;
    } else if c >= 'a' && c <= 'f' {
        Digit = c - 'a' + 10;
    } else { assert(false, "0"); }
    assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
    return Digit;
}
func Vec_WrdReadHex(pFileName string, pnWords *int32, fVerbose int32) *Vec_Wrd_t {
    var p *Vec_Wrd_t  = nil; 
    var c int32;
    var nWords int32 = -1;
    var nChars int32; var Num word;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName);
        return nil;
    }
    p = Vec_WrdAlloc( 1000 );
    for  (c = fgetc(pFile)) != -1  {
        if c == '\r' || c == '\t' || c == ' ' {
            continue; }
        if c == '\n' {
            if nChars > 0 {
                Vec_WrdPush( p, Num );
                nChars = 0; 
                Num = 0;
            }
            if nWords == -1 && Vec_WrdSize(p) > 0 {
                nWords = Vec_WrdSize(p); }
            continue;
        }
        Num |= (word)Vec_WrdReadHexOne(byte(c)) << (nChars * 4);
        if nChars++; nChars < 16 {
            continue; }
        Vec_WrdPush( p, Num );
        nChars = 0; 
        Num = 0;
    }
    assert(Vec_WrdSize(p) % nWords == 0, "Vec_WrdSize(p) % nWords == 0");
    _ = (pFile).Close();
    if pnWords != nil {
        *pnWords = nWords; }
    if fVerbose != 0 {
        fmt.Printf("Read %d words of simulation data for %d objects.\n", nWords, Vec_WrdSize(p)/nWords); }
    return p;
}


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

