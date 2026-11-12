package vec

/**CFile****************************************************************

  FileName    [vecBit.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable arrays of bits.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecBit.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

         
type Vec_Bit_t struct {
    nCap              int32;
    nSize              int32;
    pArray             *int32;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Vec_BitForEachEntry( vVec, Entry, i )                                               \
//     for ( i = 0; (i < Vec_BitSize(vVec)) && (((Entry) = Vec_BitEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_BitForEachEntryStart( vVec, Entry, i, Start )                                   \
//     for ( i = Start; (i < Vec_BitSize(vVec)) && (((Entry) = Vec_BitEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_BitForEachEntryStop( vVec, Entry, i, Stop )                                   \
//     for ( i = 0; (i < Stop) && (((Entry) = Vec_BitEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_BitForEachEntryStartStop( vVec, Entry, i, Start, Stop )                         \
//     for ( i = Start; (i < Stop) && (((Entry) = Vec_BitEntry(vVec, i)), 1); i++ )

// Original macro:
// #define Vec_BitForEachEntryReverse( vVec, pEntry, i )                                       \
//     for ( i = Vec_BitSize(vVec) - 1; (i >= 0) && (((pEntry) = Vec_BitEntry(vVec, i)), 1); i-- )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitAlloc(nCap int32) *Vec_Bit_t {
    var p *Vec_Bit_t;
    nCap = (nCap >> 5) + ((nCap & 31) > 0);
    p = new(Vec_Bit_t);
    p.nSize  = 0;
    p.nCap   = nCap * 32;
    if nCap != 0 {
        p.pArray = ABC_ALLOC( int, nCap );
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
func Vec_BitStart(nSize int32) *Vec_Bit_t {
    var p *Vec_Bit_t;
    nSize = (nSize >> 5) + ((nSize & 31) > 0);
    p = Vec_BitAlloc( nSize * 32 );
    p.nSize = nSize * 32;
    memset( p.pArray, 0, sizeof(int) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given size and cleans it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitStartFull(nSize int32) *Vec_Bit_t {
    var p *Vec_Bit_t;
    nSize = (nSize >> 5) + ((nSize & 31) > 0);
    p = Vec_BitAlloc( nSize * 32 );
    p.nSize = nSize * 32;
    memset( p.pArray, 0xff, sizeof(int) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitDup(pVec *Vec_Bit_t) *Vec_Bit_t {
    var p *Vec_Bit_t;
    assert((pVec.nSize & 31) == 0, "(pVec.nSize & 31) == 0");
    p = new(Vec_Bit_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nSize;
    if p.nCap {
        p.pArray = ABC_ALLOC( int, p.nCap >> 5 );
    } else {
        p.pArray = nil;
    }
    memcpy( p.pArray, pVec.pArray, sizeof(int) * (size_t)(p.nCap >> 5) );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitFree(p *Vec_Bit_t) {
    ABC_FREE( p.pArray );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitFreeP(p **Vec_Bit_t) {
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
func Vec_BitReleaseArray(p *Vec_Bit_t) *int32 {
    var pArray *int32  = p.pArray;
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
func Vec_BitArray(p *Vec_Bit_t) *int32 {
    return p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitSize(p *Vec_Bit_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitCap(p *Vec_Bit_t) int32 {
    return p.nCap;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitMemory(p *Vec_Bit_t) float64 {
    if !(p != nil) {
        return 0.0;
    } else {
        return 1.0 * sizeof(int) * p.nCap + sizeof(Vec_Bit_t);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitEntry(p *Vec_Bit_t, i int32) int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return (p.pArray[i >> 5] >> (i & 31)) & 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitWriteEntry(p *Vec_Bit_t, i int32, Entry int32) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    if Entry == 1 {
        p.pArray[i >> 5] |=  (1 << (i & 31));
    } else if Entry == 0 {
        p.pArray[i >> 5] &= ~(1 << (i & 31));
    } else { assert(false, "0"); }
}
func Vec_BitAddEntry(p *Vec_Bit_t, i int32) int32 {
    if Vec_BitEntry(p, i) != 0 {
        return 1; }
    Vec_BitWriteEntry( p, i, 1 );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitEntryLast(p *Vec_Bit_t) int32 {
    assert(p.nSize > 0, "p.nSize > 0");
    return Vec_BitEntry( p, p.nSize-1 );
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitGrow(p *Vec_Bit_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    nCapMin = (nCapMin >> 5) + ((nCapMin & 31) > 0);
    p.pArray = ABC_REALLOC( int, p.pArray, nCapMin ); 
    assert(p.pArray, "p.pArray");
    p.nCap   = nCapMin * 32;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitFill(p *Vec_Bit_t, nSize int32, Fill int32) {
    var i int32;
    Vec_BitGrow( p, nSize );
    nSize = (nSize >> 5) + ((nSize & 31) > 0);
    if Fill == 0 {
        for  i = 0; i < nSize; i++  {
            p.pArray[i] = 0;
        }
    
    } else if Fill == 1 {
        for  i = 0; i < nSize; i++  {
            p.pArray[i] = ~0;
        }
    
    } else { assert(false, "0"); }
    p.nSize = nSize * 32;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitFillExtra(p *Vec_Bit_t, nSize int32, Fill int32) {
    var i int32;
    if nSize <= p.nSize {
        return; }
    if nSize > 2 * p.nCap {
        Vec_BitGrow( p, nSize );
    } else if nSize > p.nCap {
        Vec_BitGrow( p, 2 * p.nCap ); }

    assert(p.nSize < nSize, "p.nSize < nSize");
    if (p.nSize >> 5) == (nSize >> 5) {
        var Mask uint32  = (~(~0 << (nSize-p.nSize)) << p.nSize);
        if Fill == 1 {
            p.pArray[nSize >> 5] |= Mask;
        } else if Fill == 0 {
            p.pArray[nSize >> 5] &= ~Mask;
        } else { assert(false, "0"); }
    
    } else {
        var Mask1 uint32;
        if p.nSize & 31 {
            Mask1 = ~0 << (p.nSize & 31);
        } else {
            Mask1 = 0;
        }
        var Mask2 uint32;
        if nSize & 31 {
            Mask2 = ~(~0 << (nSize & 31));
        } else {
            Mask2 = 0;
        }
        var w1 int32  = (p.nSize >> 5);
        var w2 int32  = (nSize >> 5);
        if Fill == 1 {
            p.pArray[w1] |= Mask1;
            p.pArray[w2] |= Mask2;
            for  i = w1 + 1; i < w2; i++  {
                p.pArray[i] = ~0;
            }
        
        } else if Fill == 0 {
            p.pArray[w1] &= ~Mask1;
            p.pArray[w2] &= ~Mask2;
            for  i = w1 + 1; i < w2; i++  {
                p.pArray[i] = 0;
            }
        
        } else { assert(false, "0"); }
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Returns the entry even if the place not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitGetEntry(p *Vec_Bit_t, i int32) int32 {
    Vec_BitFillExtra( p, i + 1, 0 );
    return Vec_BitEntry( p, i );
}

/**Function*************************************************************

  Synopsis    [Inserts the entry even if the place does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitSetEntry(p *Vec_Bit_t, i int32, Entry int32) {
    Vec_BitFillExtra( p, i + 1, 0 );
    Vec_BitWriteEntry( p, i, Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitShrink(p *Vec_Bit_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitClear(p *Vec_Bit_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitPush(p *Vec_Bit_t, Entry int32) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_BitGrow( p, 16 );
        } else {
            Vec_BitGrow( p, 2 * p.nCap ); }
    }
    if Entry == 1 {
        p.pArray[p.nSize >> 5] |=  (1 << (p.nSize & 31));
    } else if Entry == 0 {
        p.pArray[p.nSize >> 5] &= ~(1 << (p.nSize & 31));
    } else { assert(false, "0"); }
    p.nSize++;
}

/**Function*************************************************************

  Synopsis    [Returns the last entry and removes it from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitPop(p *Vec_Bit_t) int32 {
    var Entry int32;
    assert(p.nSize > 0, "p.nSize > 0");
    Entry = Vec_BitEntryLast( p );
    p.nSize--;
    return Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitCountWord(uWord uint32) int32 {
    uWord = (uWord & 0x55555555) + ((uWord>>1) & 0x55555555);
    uWord = (uWord & 0x33333333) + ((uWord>>2) & 0x33333333);
    uWord = (uWord & 0x0F0F0F0F) + ((uWord>>4) & 0x0F0F0F0F);
    uWord = (uWord & 0x00FF00FF) + ((uWord>>8) & 0x00FF00FF);
    return  (uWord & 0x0000FFFF) + (uWord>>16);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitCount(p *Vec_Bit_t) int32 {
    var pArray *uint32  = (unsigned *)p.pArray;
    var nWords int32  = (p.nSize >> 5) + ((p.nSize & 31) > 0);
    var i, Counter int32;
    if p.nSize & 31 {
        assert(nWords > 0, "nWords > 0");
        for  i = 0; i < nWords-1; i++  {
            Counter += Vec_BitCountWord( pArray[i] );
        }
        Counter += Vec_BitCountWord( pArray[i] & ~(~0 << (p.nSize & 31)) );
    
    } else {
        for  i = 0; i < nWords; i++  {
            Counter += Vec_BitCountWord( pArray[i] );
        }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_BitReset(p *Vec_Bit_t) {
    var i int32;
    var nWords int32 = (p.nSize >> 5) + ((p.nSize & 31) > 0);
    for  i = 0; i < nWords; i++  {
        p.pArray[i] = 0;
    }
}

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

