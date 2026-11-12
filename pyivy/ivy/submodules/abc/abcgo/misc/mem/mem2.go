package mem

import "fmt"

/**CFile****************************************************************

  FileName    [mem2.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Memory management.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mem2.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/




//#include "misc/vec/vec.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Mmr_Flex_t =     Mmr_Flex_t_;     
type Mmr_Fixed_t =    Mmr_Fixed_t_;    
type Mmr_Step_t =     Mmr_Step_t_;     

type Mmr_Flex_t_ struct {
    nPageBase           int32;     // log2 page size in words
    PageMask           int32;      // page mask
    nEntries           int32;      // entries allocated
    nEntriesMax           int32;   // max number of enries used
    iNext           int32;         // next word to be used
    vPages     Vec_Ptr_t;        // memory pages
}

type Mmr_Fixed_t_ struct {
    nPageBase           int32;     // log2 page size in words
    PageMask           int32;      // page mask
    nEntryWords           int32;   // entry size in words
    nEntries           int32;      // entries allocated
    nEntriesMax           int32;   // max number of enries used
    vPages     Vec_Ptr_t;        // memory pages
    vFrees     Vec_Int_t;        // free entries
}

type Mmr_Step_t_ struct {
    nBits           int32;         // the number of bits
    uMask           int32;         // the number of managers minus 1
    nEntries           int32;      // the number of entries
    nEntriesMax           int32;   // the max number of entries
    nEntriesAll           int32;   // the total number of entries
    pMems [0]Mmr_Fixed_t;      // memory managers: 2^0 words, 2^1 words, etc
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
func Mmr_FlexStart(nPageBase int32) *Mmr_Flex_t {
    
    // inlining of macro ABC_CALLOC
    p := &Mmr_Flex_t{};
    p.nPageBase = nPageBase;
    p.PageMask  = (1 << nPageBase) - 1;
    p.iNext     = (1 << nPageBase);
    return p;
}
func Mmr_FlexStop(p *Mmr_Flex_t) {
    var pPage *word;
    var i int32;
    if false && Vec_PtrSize(&p.vPages) {
        fmt.Printf("Using %3d pages of %6d words each with %6d entries (max = %6d). Total memory = %5.2f MB.\n", Vec_PtrSize(&p.vPages), p.nPageBase ? 1 << p.nPageBase : 0, p.nEntries, p.nEntriesMax, 1.0 * Vec_PtrSize(&p.vPages) * (1 << p.nPageBase) * 8 / (1 << 20)); }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vPages)) && (((pPage) = (word *)Vec_PtrEntry(&p.vPages, i)), 1); i++  {
        // inlining of macro ABC_FREE
        ((pPage) ? (free((char *) (pPage)), (pPage) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((p.vPages.pArray) ? (free((char *) (p.vPages.pArray)), (p.vPages.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Mmr_FlexEntry(p *Mmr_Flex_t, h int32) *word {
    assert(h > 0 && h < p.iNext, "h > 0 && h < p.iNext");
    return (word *)Vec_PtrEntry(&p.vPages, (h >> p.nPageBase)) + (h & p.PageMask);
}
func Mmr_FlexFetch(p *Mmr_Flex_t, nWords int32) int32 {
    var hEntry int32;
    assert(nWords > 0 && nWords < p.PageMask, "nWords > 0 && nWords < p.PageMask");
    if p.iNext + nWords >= p.PageMask {
        // inlining of macro ABC_FALLOC
        Vec_PtrPush( &p.vPages, ((word *) memset(malloc(sizeof(word) * (size_t)(p.PageMask + 1)), 0xff, sizeof(word) * (size_t)(p.PageMask + 1))) );
        p.iNext = 1;
    }
    hEntry = ((Vec_PtrSize(&p.vPages) - 1) << p.nPageBase) | p.iNext;
    p.iNext += nWords;
    p.nEntries++;
    p.nEntriesMax = Abc_MaxInt( p.nEntriesMax, p.nEntries );
    return hEntry;
}
func Mmr_FlexRelease(p *Mmr_Flex_t, h int32) {
    assert(h > 0 && h < p.iNext, "h > 0 && h < p.iNext");
    if (h >> p.nPageBase) && Vec_PtrEntry(&p.vPages, (h >> p.nPageBase) - 1) {
        var pPage *word  = (word *)Vec_PtrEntry(&p.vPages, (h >> p.nPageBase) - 1);
        Vec_PtrWriteEntry( &p.vPages, (h >> p.nPageBase) - 1, nil );
        // inlining of macro ABC_FREE
        ((pPage) ? (free((char *) (pPage)), (pPage) = 0) : 0);
    }
}

 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mmr_FixedCreate(p *Mmr_Fixed_t, nPageBase int32, nEntryWords int32) {
    assert(nEntryWords > 0 && nEntryWords < (1 << nPageBase), "nEntryWords > 0 && nEntryWords < (1 << nPageBase)");
    p.nPageBase   = nPageBase;
    p.PageMask    = (1 << nPageBase) - 1;
    p.nEntryWords = nEntryWords;
}
func Mmr_FixedStart(nPageBase int32, nEntryWords int32) *Mmr_Fixed_t {
    // inlining of macro ABC_CALLOC
    var p *Mmr_Fixed_t  = &Mmr_Fixed_t{};
    Mmr_FixedCreate( p, nPageBase, nEntryWords );
    return p;
}
func Mmr_FixedStop(p *Mmr_Fixed_t, fFreeLast int32) {
    var pPage *word;
    var i int32;
    if false && Vec_PtrSize(&p.vPages) {
        fmt.Printf("Using %3d pages of %6d words each with %6d entries (max = %6d) of size %d. Total memory = %5.2f MB.\n", Vec_PtrSize(&p.vPages), p.nPageBase ? 1 << p.nPageBase : 0, p.nEntries, p.nEntriesMax, p.nEntryWords, 1.0 * Vec_PtrSize(&p.vPages) * (1 << p.nPageBase) * 8 / (1 << 20)); }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vPages)) && (((pPage) = (word *)Vec_PtrEntry(&p.vPages, i)), 1); i++  {
        // inlining of macro ABC_FREE
        ((pPage) ? (free((char *) (pPage)), (pPage) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((p.vPages.pArray) ? (free((char *) (p.vPages.pArray)), (p.vPages.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vFrees.pArray) ? (free((char *) (p.vFrees.pArray)), (p.vFrees.pArray) = 0) : 0);
    if fFreeLast != 0 {
        // inlining of macro ABC_FREE
        ((p) ? (free((char *) (p)), (p) = 0) : 0); }
}
func Mmr_FixedEntry(p *Mmr_Fixed_t, h int32) *word {
    assert(h > 0 && h < (Vec_PtrSize(&p.vPages) << p.nPageBase), "h > 0 && h < (Vec_PtrSize(&p.vPages) << p.nPageBase)");
    return (word *)Vec_PtrEntry(&p.vPages, (h >> p.nPageBase)) + (h & p.PageMask);
}
func Mmr_FixedFetch(p *Mmr_Fixed_t) int32 {
    if Vec_IntSize(&p.vFrees) == 0 {
        var i int32;
        var hEntry int32 = Vec_PtrSize(&p.vPages) << p.nPageBase;
        // inlining of macro ABC_FALLOC
        Vec_PtrPush( &p.vPages, ((word *) memset(malloc(sizeof(word) * (size_t)(p.PageMask + 1)), 0xff, sizeof(word) * (size_t)(p.PageMask + 1))) );
        for  i = 1; i + p.nEntryWords <= p.PageMask; i += p.nEntryWords  {
            Vec_IntPush( &p.vFrees, hEntry | i );
        }
        Vec_IntReverseOrder( &p.vFrees );
    }
    p.nEntries++;
    p.nEntriesMax = Abc_MaxInt( p.nEntriesMax, p.nEntries );
    return Vec_IntPop( &p.vFrees );
}
func Mmr_FixedRecycle(p *Mmr_Fixed_t, h int32) {
    p.nEntries--;
    memset( Mmr_FixedEntry(p, h), 0xFF, sizeof(word) * p.nEntryWords );
    Vec_IntPush( &p.vFrees, h );
}
func Mmr_FixedMemory(p *Mmr_Fixed_t) int32 {
    return Vec_PtrSize(&p.vPages) * (p.PageMask + 1);
}

 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mmr_StepStart(nPageBase int32, nWordBase int32) *Mmr_Step_t {
    // inlining of macro ABC_CALLOC
    var pMemory string  = ((char *) calloc((size_t)(sizeof(Mmr_Step_t) + sizeof(Mmr_Fixed_t) * (1 << nWordBase)), sizeof(char)));
    var p *Mmr_Step_t  = (Mmr_Step_t *)pMemory;
    var i int32;
    p.nBits = nWordBase;
    p.uMask = (1 << nWordBase) - 1;
    for  i = 1; i <= p.uMask; i++  {
        Mmr_FixedCreate( p.pMems + i, nPageBase, i );
    }
    return p;
}
func Mmr_StepStop(p *Mmr_Step_t) {
    var i int32;
    for  i = 0; i <= p.uMask; i++  {
        Mmr_FixedStop( p.pMems + i, 0 );
    }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Mmr_StepEntry(p *Mmr_Step_t, h int32) *word {
    assert((h & p.uMask) > 0, "(h & p.uMask) > 0");
    return Mmr_FixedEntry( p.pMems + (h & p.uMask), (h >> p.nBits) );
}
func Mmr_StepFetch(p *Mmr_Step_t, nWords int32) int32 {
    assert(nWords > 0 && nWords <= p.uMask, "nWords > 0 && nWords <= p.uMask");
    p.nEntries++;
    p.nEntriesAll++;
    p.nEntriesMax = Abc_MaxInt( p.nEntriesMax, p.nEntries );
    return (Mmr_FixedFetch(p.pMems + nWords) << p.nBits) | nWords;
}
func Mmr_StepRecycle(p *Mmr_Step_t, h int32) {
    p.nEntries--;
    Mmr_FixedRecycle( p.pMems + (h & p.uMask), (h >> p.nBits) );
}
func Mmr_StepMemory(p *Mmr_Step_t) int32 {
    var i, Mem int32;
    for  i = 1; i <= p.uMask; i++  {
        Mem += Mmr_FixedMemory( p.pMems + i );
    }
    return Mem;
}


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


