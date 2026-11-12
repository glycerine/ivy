package vec

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [vecMem.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable array of memory pieces.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 20, 2012.]

  Revision    [$Id: vecMem.h,v 1.00 2012/07/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>

//ABC_NAMESPACE_HEADER_START

/* 
   This vector stores pieces of memory of the given size.
   It is useful for representing truth tables and any other objects
   of the fixed size.  It is better that Extra_MmFixed because the
   entry IDs can be used as handles to retrieve memory pieces without 
   the need for an array of pointers from entry IDs into memory pieces
   (this can save 8(4) bytes per object on a 64(32)-bit platform).
*/

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

         
type Vec_Mem_t struct {
    nEntrySize              int32;  // entry size (in terms of 8-byte words)
    nEntries              int32;    // number of entries currently used
    LogPageSze              int32;  // log2 of page size (in terms of entries)
    PageMask              int32;    // page mask
    nPageAlloc              int32;  // number of pages currently allocated
    iPage              int32;       // the number of a page currently used   
    ppPages           **word;     // memory pages
    vTable       *Vec_Int_t;      // hash table
    vNexts       *Vec_Int_t;      // next pointers
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Vec_MemForEachEntry( p, pEntry, i )                                              \
//     for ( i = 0; (i < Vec_MemEntryNum(p)) && ((pEntry) = Vec_MemReadEntry(p, i)); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a memory vector.]

  Description [Entry size is in terms of 8-byte words. Page size is log2
  of the number of entries on one page.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemAlloc_(p *Vec_Mem_t, nEntrySize int32, LogPageSze int32) {
    *p = Vec_Mem_t{};
    p.nEntrySize = nEntrySize;
    p.LogPageSze = LogPageSze;
    p.PageMask   = (1 << p.LogPageSze) - 1;
    p.iPage      = -1;
}
func Vec_MemAlloc(nEntrySize int32, LogPageSze int32) *Vec_Mem_t {
    var p *Vec_Mem_t;
    p = new(Vec_Mem_t);
    p.nEntrySize = nEntrySize;
    p.LogPageSze = LogPageSze;
    p.PageMask   = (1 << p.LogPageSze) - 1;
    p.iPage      = -1;
    return p;
}
func Vec_MemFree(p *Vec_Mem_t) {
    var i int32;
    for  i = 0; i <= p.iPage; i++  {
        ABC_FREE( p.ppPages[i] );
    }
    ABC_FREE( p.ppPages );
    ABC_FREE( p );
}
func Vec_MemFreeP(p **Vec_Mem_t) {
    if *p == nil {
        return; }
    Vec_MemFree( *p );
    *p = nil;
}
func Vec_MemDup(pVec *Vec_Mem_t) *Vec_Mem_t {
    var p *Vec_Mem_t  = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemFill(pVec *Vec_Mem_t, nEntries int32) {
}
func Vec_MemClean(pVec *Vec_Mem_t, nEntries int32) {
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemEntrySize(p *Vec_Mem_t) int32 {
    return p.nEntrySize;
}
func Vec_MemEntryNum(p *Vec_Mem_t) int32 {
    return p.nEntries;
}
func Vec_MemPageSize(p *Vec_Mem_t) int32 {
    return p.LogPageSze;
}
func Vec_MemPageNum(p *Vec_Mem_t) int32 {
    return p.iPage+1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemMemory(p *Vec_Mem_t) float64 {
    return float64(sizeof(word)) * p.nEntrySize * (1 << p.LogPageSze) * (p.iPage + 1) + float64(sizeof(word *)) * p.nPageAlloc + float64(sizeof(Vec_Mem_t));
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemReadEntry(p *Vec_Mem_t, i int32) *word {
    assert(i >= 0 && i < p.nEntries, "i >= 0 && i < p.nEntries");
    return p.ppPages[i >> p.LogPageSze] + p.nEntrySize * (i & p.PageMask);
}
func Vec_MemReadEntryLast(p *Vec_Mem_t) *word {
    assert(p.nEntries > 0, "p.nEntries > 0");
    return Vec_MemReadEntry( p, p.nEntries-1 );
}
func Vec_MemWriteEntry(p *Vec_Mem_t, i int32, pEntry *word) {
    var pPlace *word  = Vec_MemReadEntry( p, i );
    memmove( pPlace, pEntry, sizeof(word) * p.nEntrySize );
}
func Vec_MemGetEntry(p *Vec_Mem_t, i int32) *word {
    assert(i >= 0, "i >= 0");
    if i >= p.nEntries {
        var k int32;
        var iPageNew int32 = (i >> p.LogPageSze);
        if p.iPage < iPageNew {
            // realloc page pointers if needed
            if iPageNew >= p.nPageAlloc {
                p.ppPages = ABC_REALLOC( word *, p.ppPages, (p.nPageAlloc = p.nPageAlloc ? 2 * p.nPageAlloc : iPageNew + 32) ); }
            // allocate new pages if needed
            for  k = p.iPage + 1; k <= iPageNew; k++  {
                p.ppPages[k] = ABC_ALLOC( word, p.nEntrySize * (1 << p.LogPageSze) );
            }
            // update page counter
            p.iPage = iPageNew;
        }
        // update entry counter
        p.nEntries = i + 1;
    }
    return Vec_MemReadEntry( p, i );
}
func Vec_MemSetEntry(p *Vec_Mem_t, i int32, pEntry *word) {
    var pPlace *word  = Vec_MemGetEntry( p, i );
    memmove( pPlace, pEntry, sizeof(word) * (size_t)p.nEntrySize );
}
func Vec_MemPush(p *Vec_Mem_t, pEntry *word) {
    var pPlace *word  = Vec_MemGetEntry( p, p.nEntries );
    memmove( pPlace, pEntry, sizeof(word) * (size_t)p.nEntrySize );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemShrink(p *Vec_Mem_t, nEntriesNew int32) {
    var i int32;
    var iPageOld int32 = p.iPage;
    assert(nEntriesNew <= p.nEntries, "nEntriesNew <= p.nEntries");
    p.nEntries = nEntriesNew;
    p.iPage = (nEntriesNew >> p.LogPageSze);
    for  i = p.iPage + 1; i <= iPageOld; i++  {
        ABC_FREE( p.ppPages[i] );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemDumpDigit(pFile *os.File, HexDigit int32) {
    assert(HexDigit >= 0 && HexDigit < 16, "HexDigit >= 0 && HexDigit < 16");
    if HexDigit < 10 {
        fmt.Fprintf(pFile, "%d", HexDigit);
    } else {
        fmt.Fprintf(pFile, "%c", 'A' + HexDigit-10); }
}
func Vec_MemDump(pFile *os.File, pVec *Vec_Mem_t) {
    var pEntry *word;
    var i, w, d int32;
    if pFile == os.Stdout {
        fmt.Printf("Memory vector has %d entries: \n", Vec_MemEntryNum(pVec)); }
    // inlining of macro Vec_MemForEachEntry
    for  i = 0; (i < Vec_MemEntryNum(pVec)) && ((pEntry) = Vec_MemReadEntry(pVec, i)); i++  {
        for  w = pVec.nEntrySize - 1; w >= 0; w--  {
            for  d = 15; d >= 0; d--  {
                Vec_MemDumpDigit( pFile, int32((pEntry[w] >> (d<<2))) & 15 );
            }
        }
        fmt.Fprintf(pFile, "\n");
    }
}

/**Function*************************************************************

  Synopsis    [Hashing entries in the memory vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemHashAlloc(p *Vec_Mem_t, nTableSize int32) {
    assert(p.vTable == nil && p.vNexts == nil, "p.vTable == nil && p.vNexts == nil");
    p.vTable = Vec_IntStartFull( Abc_PrimeCudd(nTableSize) );
    p.vNexts = Vec_IntAlloc( nTableSize );
}
func Vec_MemHashFree(p *Vec_Mem_t) {
    if p == nil {
        return; }
    Vec_IntFreeP( &p.vTable );
    Vec_IntFreeP( &p.vNexts );
}
var spStatic_Vec_MemHashKey_s_Primes [8]int32 = [8]int32{ 1699, 4177, 5147, 5647, 6343, 7103, 7873, 8147 };
func Vec_MemHashKey(p *Vec_Mem_t, pEntry *word) uint32 {
    
    var i int32;
    var nData int32 = 2 * p.nEntrySize;
    var pData *uint32  = (unsigned *)pEntry;
    var uHash uint32;
    for  i = 0; i < nData; i++  {
        uHash += pData[i] * spStatic_Vec_MemHashKey_s_Primes[i & 0x7];
    }
    return uHash % Vec_IntSize(p.vTable);
}
func Vec_MemHashLookup(p *Vec_Mem_t, pEntry *word) *int32 {
    var pSpot *int32  = Vec_IntEntryP( p.vTable, Vec_MemHashKey(p, pEntry) );
    for  ; *pSpot != -1; pSpot = Vec_IntEntryP(p.vNexts, *pSpot)  {
        if !memcmp( Vec_MemReadEntry(p, *pSpot), pEntry, sizeof(word) * p.nEntrySize ) { // equal
            return pSpot; }
    }
    return pSpot;
}
func Vec_MemHashProfile(p *Vec_Mem_t) {
    var e int32;
    for  e = 0; e < 1000; e++  {
        var Count int32;
        var pSpot *int32  = Vec_IntEntryP( p.vTable, e );
        for  ; *pSpot != -1; pSpot = Vec_IntEntryP(p.vNexts, *pSpot)  {
            Count++;
        }
        fmt.Printf("%d ", Count);
    }
    fmt.Printf("\n");
}
func Vec_MemHashResize(p *Vec_Mem_t) {
    var pEntry *word;
    var i int32;
    var pSpot *int32;
    //Vec_MemHashProfile( p );
    Vec_IntFill( p.vTable, Abc_PrimeCudd(2 * Vec_IntSize(p.vTable)), -1 );
    Vec_IntClear( p.vNexts );
    // inlining of macro Vec_MemForEachEntry
    for  i = 0; (i < Vec_MemEntryNum(p)) && ((pEntry) = Vec_MemReadEntry(p, i)); i++  {
        pSpot = Vec_MemHashLookup( p, pEntry );
        assert(*pSpot == -1, "*pSpot == -1");
        *pSpot = Vec_IntSize(p.vNexts);
        Vec_IntPush( p.vNexts, -1 );
    }
    assert(p.nEntries == Vec_IntSize(p.vNexts), "p.nEntries == Vec_IntSize(p.vNexts)");
}
func Vec_MemHashInsert(p *Vec_Mem_t, pEntry *word) int32 {
    var pSpot *int32;
    if p.nEntries > Vec_IntSize(p.vTable) {
        Vec_MemHashResize( p ); }
    pSpot = Vec_MemHashLookup( p, pEntry );
    if *pSpot != -1 {
        return *pSpot; }
    *pSpot = Vec_IntSize(p.vNexts);
    Vec_IntPush( p.vNexts, -1 );
    Vec_MemPush( p, pEntry );
    assert(p.nEntries == Vec_IntSize(p.vNexts), "p.nEntries == Vec_IntSize(p.vNexts)");
    return Vec_IntSize(p.vNexts) - 1;
}


/**Function*************************************************************

  Synopsis    [Allocates memory vector for storing truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemAllocForTTSimple(nVars int32) *Vec_Mem_t {
    var nWords int32;
    if nVars <= 6 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 6));
    }
    var vTtMem *Vec_Mem_t  = Vec_MemAlloc( nWords, 12 );
    Vec_MemHashAlloc( vTtMem, 10000 );
    return vTtMem;
}
func Vec_MemAllocForTT(nVars int32, fCompl int32) *Vec_Mem_t {
    var Value int32;
    var nWords int32;
    if nVars <= 6 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 6));
    }
    var uTruth *word  = ABC_ALLOC( word, nWords ); 
    var vTtMem *Vec_Mem_t  = Vec_MemAlloc( nWords, 12 );
    Vec_MemHashAlloc( vTtMem, 10000 );
    memset( uTruth, 0x00, sizeof(word) * nWords );
    Value = Vec_MemHashInsert( vTtMem, uTruth ); assert(Value == 0, "Value == 0");
    if fCompl != 0 {
        memset( uTruth, 0x55, sizeof(word) * nWords );
    } else {
        memset( uTruth, 0xAA, sizeof(word) * nWords ); }
    Value = Vec_MemHashInsert( vTtMem, uTruth ); assert(Value == 1, "Value == 1");
    ABC_FREE( uTruth );
    return vTtMem;
}
func Vec_MemAddMuxTT(p *Vec_Mem_t, nVars int32) {
    var Value int32;
    var nWords int32;
    if nVars <= 6 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 6));
    }
    var uTruth *word  = ABC_ALLOC( word, nWords ); 
    memset( uTruth, 0xCA, sizeof(word) * nWords );
    Value = Vec_MemHashInsert( p, uTruth ); assert(Value == 2, "Value == 2");
    ABC_FREE( uTruth );
}
func Vec_MemDumpTruthTables(p *Vec_Mem_t, pName string, nLutSize int32) {
    var pFile *os.File;
    var pFileName string // was [1000]byte in C

    pFileName = fmt.Sprintf("tt_%s_%02d.txt", pName ? pName : nil, nLutSize);
    if pName {
        pFile = fopen( pFileName, "wb" );
    } else {
        pFile = os.Stdout;
    }
    Vec_MemDump( pFile, p );
    if pFile != os.Stdout {
        _ = (pFile).Close(); }
    fmt.Printf("Dumped %d %d-var truth tables into file \"%s\" (%.2f MB).\n", Vec_MemEntryNum(p), nLutSize, pName ? pFileName : "stdout", 8.0 * Vec_MemEntryNum(p) * Vec_MemEntrySize(p) / (1 << 20));
}

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

