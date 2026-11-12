package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigMem.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Memory managers.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigMem.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Aig_MmFixed_t_ struct {
    // information about individual entries
    nEntrySize           int32;    // the size of one entry
    nEntriesAlloc           int32; // the total number of entries allocated
    nEntriesUsed           int32;  // the number of entries in use
    nEntriesMax           int32;   // the max number of entries in use
    pEntriesFree         string;  // the linked list of free entries

    // this is where the memory is stored
    nChunkSize           int32;    // the size of one chunk
    nChunksAlloc           int32;  // the maximum number of memory chunks 
    nChunks           int32;       // the current number of memory chunks 
    pChunks        *string;       // the allocated memory

    // statistics
    nMemoryUsed           int32;   // memory used in the allocated entries
    nMemoryAlloc           int32;  // memory allocated
}

type Aig_MmFlex_t_ struct {
    // information about individual entries
    nEntriesUsed           int32;  // the number of entries allocated
    pCurrent         string;      // the current pointer to free memory
    pEnd         string;          // the first entry outside the free memory

    // this is where the memory is stored
    nChunkSize           int32;    // the size of one chunk
    nChunksAlloc           int32;  // the maximum number of memory chunks 
    nChunks           int32;       // the current number of memory chunks 
    pChunks        *string;       // the allocated memory

    // statistics
    nMemoryUsed           int32;   // memory used in the allocated entries
    nMemoryAlloc           int32;  // memory allocated
}

type Aig_MmStep_t_ struct {
    nMems               int32;    // the number of fixed memory managers employed
    pMems   **Aig_MmFixed_t;    // memory managers: 2^1 words, 2^2 words, etc
    nMapSize               int32; // the size of the memory array
    pMap   **Aig_MmFixed_t;     // maps the number of bytes into its memory manager
    // additional memory chunks
    nChunksAlloc           int32;  // the maximum number of memory chunks 
    nChunks           int32;       // the current number of memory chunks 
    pChunks        *string;       // the allocated memory
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates memory pieces of fixed size.]

  Description [The size of the chunk is computed as the minimum of
  1024 entries and 64K. Can only work with entry size at least 4 byte long.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFixedStart(nEntrySize int32, nEntriesMax int32) *Aig_MmFixed_t {
    var p *Aig_MmFixed_t;

    // inlining of macro ABC_ALLOC
    p = ((Aig_MmFixed_t *) malloc(sizeof(Aig_MmFixed_t) * (size_t)(1)));
    memset( p, 0, sizeof(Aig_MmFixed_t) );

    p.nEntrySize    = nEntrySize;
    p.nEntriesAlloc = 0;
    p.nEntriesUsed  = 0;
    p.pEntriesFree  = nil;

    p.nChunkSize = nEntriesMax / 8;
    if p.nChunkSize < 8 {
        p.nChunkSize = 8; }

    p.nChunksAlloc  = 64;
    p.nChunks       = 0;
    // inlining of macro ABC_ALLOC
    p.pChunks       = ((char * *) malloc(sizeof(char *) * (size_t)(p.nChunksAlloc)));

    p.nMemoryUsed   = 0;
    p.nMemoryAlloc  = 0;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFixedStop(p *Aig_MmFixed_t, fVerbose int32) {
    var i int32;
    if p == nil {
        return; }
    if fVerbose != 0 {
        fmt.Printf("Fixed memory manager: Entry = %5d. Chunk = %5d. Chunks used = %5d.\n", p.nEntrySize, p.nChunkSize, p.nChunks);
        fmt.Printf("   Entries used = %8d. Entries peak = %8d. Memory used = %8d. Memory alloc = %8d.\n", p.nEntriesUsed, p.nEntriesMax, p.nEntrySize * p.nEntriesUsed, p.nMemoryAlloc);
    }
    for  i = 0; i < p.nChunks; i++  {
        // inlining of macro ABC_FREE
        ((p.pChunks[i]) ? (free((char *) (p.pChunks[i])), (p.pChunks[i]) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((p.pChunks) ? (free((char *) (p.pChunks)), (p.pChunks) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFixedEntryFetch(p *Aig_MmFixed_t) string {
    var pTemp string;
    var i int32;

    // check if there are still free entries
    if p.nEntriesUsed == p.nEntriesAlloc { // need to allocate more entries
        assert(p.pEntriesFree == nil, "p.pEntriesFree == nil");
        if p.nChunks == p.nChunksAlloc {
            p.nChunksAlloc *= 2;
            // inlining of macro ABC_REALLOC
            
            if p.pChunks {
                p.pChunks = ((char * *) realloc((char *)(p.pChunks), sizeof(char *) * (size_t)(p.nChunksAlloc)));
            } else {
                p.pChunks = ((char * *) malloc(sizeof(char *) * (size_t)(p.nChunksAlloc)));
            } 
        }
        // inlining of macro ABC_ALLOC
        p.pEntriesFree = ((char *) malloc(sizeof(char) * (size_t)(p.nEntrySize * p.nChunkSize)));
        p.nMemoryAlloc += p.nEntrySize * p.nChunkSize;
        // transform these entries into a linked list
        pTemp = p.pEntriesFree;
        for  i = 1; i < p.nChunkSize; i++  {
            *((char **)pTemp) = pTemp + p.nEntrySize;
            pTemp += p.nEntrySize;
        }
        // set the last link
        *((char **)pTemp) = nil;
        // add the chunk to the chunk storage
        p.pChunks[ p.nChunks++ ] = p.pEntriesFree;
        // add to the number of entries allocated
        p.nEntriesAlloc += p.nChunkSize;
    }
    // incrememt the counter of used entries
    p.nEntriesUsed++;
    if p.nEntriesMax < p.nEntriesUsed {
        p.nEntriesMax = p.nEntriesUsed; }
    // return the first entry in the free entry list
    pTemp = p.pEntriesFree;
    p.pEntriesFree = *((char **)pTemp);
    return pTemp;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFixedEntryRecycle(p *Aig_MmFixed_t, pEntry string) {
    // decrement the counter of used entries
    p.nEntriesUsed--;
    // add the entry to the linked list of free entries
    *((char **)pEntry) = p.pEntriesFree;
    p.pEntriesFree = pEntry;
}

/**Function*************************************************************

  Synopsis    []

  Description [Relocates all the memory except the first chunk.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFixedRestart(p *Aig_MmFixed_t) {
    var i int32;
    var pTemp string;
    if p.nChunks == 0 {
        return; }
    // deallocate all chunks except the first one
    for  i = 1; i < p.nChunks; i++  {
        // inlining of macro ABC_FREE
        ((p.pChunks[i]) ? (free((char *) (p.pChunks[i])), (p.pChunks[i]) = 0) : 0);
    }
    p.nChunks = 1;
    // transform these entries into a linked list
    pTemp = p.pChunks[0];
    for  i = 1; i < p.nChunkSize; i++  {
        *((char **)pTemp) = pTemp + p.nEntrySize;
        pTemp += p.nEntrySize;
    }
    // set the last link
    *((char **)pTemp) = nil;
    // set the free entry list
    p.pEntriesFree  = p.pChunks[0];
    // set the correct statistics
    p.nMemoryAlloc  = p.nEntrySize * p.nChunkSize;
    p.nMemoryUsed   = 0;
    p.nEntriesAlloc = p.nChunkSize;
    p.nEntriesUsed  = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFixedReadMemUsage(p *Aig_MmFixed_t) int32 {
    return p.nMemoryAlloc;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFixedReadMaxEntriesUsed(p *Aig_MmFixed_t) int32 {
    return p.nEntriesMax;
}



/**Function*************************************************************

  Synopsis    [Allocates entries of flexible size.]

  Description [Can only work with entry size at least 4 byte long.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFlexStart() *Aig_MmFlex_t {
    var p *Aig_MmFlex_t;

    // inlining of macro ABC_ALLOC
    p = ((Aig_MmFlex_t *) malloc(sizeof(Aig_MmFlex_t) * (size_t)(1)));
    memset( p, 0, sizeof(Aig_MmFlex_t) );

    p.nEntriesUsed  = 0;
    p.pCurrent      = nil;
    p.pEnd          = nil;

    p.nChunkSize    = (1 << 18);
    p.nChunksAlloc  = 64;
    p.nChunks       = 0;
    // inlining of macro ABC_ALLOC
    p.pChunks       = ((char * *) malloc(sizeof(char *) * (size_t)(p.nChunksAlloc)));

    p.nMemoryUsed   = 0;
    p.nMemoryAlloc  = 0;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFlexStop(p *Aig_MmFlex_t, fVerbose int32) {
    var i int32;
    if p == nil {
        return; }
    if fVerbose != 0 {
        fmt.Printf("Flexible memory manager: Chunk size = %d. Chunks used = %d.\n", p.nChunkSize, p.nChunks);
        fmt.Printf("   Entries used = %d. Memory used = %d. Memory alloc = %d.\n", p.nEntriesUsed, p.nMemoryUsed, p.nMemoryAlloc);
    }
    for  i = 0; i < p.nChunks; i++  {
        // inlining of macro ABC_FREE
        ((p.pChunks[i]) ? (free((char *) (p.pChunks[i])), (p.pChunks[i]) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((p.pChunks) ? (free((char *) (p.pChunks)), (p.pChunks) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFlexEntryFetch(p *Aig_MmFlex_t, nBytes int32) string {
    var pTemp string;
#ifdef ABC_MEMALIGN
    // extend size to max alignment
    nBytes += (ABC_MEMALIGN - nBytes % ABC_MEMALIGN) % ABC_MEMALIGN;
#endif
    // check if there are still free entries
    if p.pCurrent == nil || p.pCurrent + nBytes > p.pEnd { // need to allocate more entries
        if p.nChunks == p.nChunksAlloc {
            p.nChunksAlloc *= 2;
            // inlining of macro ABC_REALLOC
            
            if p.pChunks {
                p.pChunks = ((char * *) realloc((char *)(p.pChunks), sizeof(char *) * (size_t)(p.nChunksAlloc)));
            } else {
                p.pChunks = ((char * *) malloc(sizeof(char *) * (size_t)(p.nChunksAlloc)));
            } 
        }
        if nBytes > p.nChunkSize {
            // resize the chunk size if more memory is requested than it can give
            // (ideally, this should never happen)
            p.nChunkSize = 2 * nBytes;
        }
        // inlining of macro ABC_ALLOC
        p.pCurrent = ((char *) malloc(sizeof(char) * (size_t)(p.nChunkSize)));
        p.pEnd     = p.pCurrent + p.nChunkSize;
        p.nMemoryAlloc += p.nChunkSize;
        // add the chunk to the chunk storage
        p.pChunks[ p.nChunks++ ] = p.pCurrent;
    }
    assert(p.pCurrent + nBytes <= p.pEnd, "p.pCurrent + nBytes <= p.pEnd");
    // increment the counter of used entries
    p.nEntriesUsed++;
    // keep track of the memory used
    p.nMemoryUsed += nBytes;
    // return the next entry
    pTemp = p.pCurrent;
    p.pCurrent += nBytes;
    return pTemp;
}

/**Function*************************************************************

  Synopsis    []

  Description [Relocates all the memory except the first chunk.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFlexRestart(p *Aig_MmFlex_t) {
    var i int32;
    if p.nChunks == 0 {
        return; }
    // deallocate all chunks except the first one
    for  i = 1; i < p.nChunks; i++  {
        // inlining of macro ABC_FREE
        ((p.pChunks[i]) ? (free((char *) (p.pChunks[i])), (p.pChunks[i]) = 0) : 0);
    }
    p.nChunks  = 1;
    p.nMemoryAlloc = p.nChunkSize;
    // transform these entries into a linked list
    p.pCurrent = p.pChunks[0];
    p.pEnd     = p.pCurrent + p.nChunkSize;
    p.nEntriesUsed = 0;
    p.nMemoryUsed = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmFlexReadMemUsage(p *Aig_MmFlex_t) int32 {
    return p.nMemoryUsed;
}





/**Function*************************************************************

  Synopsis    [Starts the hierarchical memory manager.]

  Description [This manager can allocate entries of any size.
  Iternally they are mapped into the entries with the number of bytes
  equal to the power of 2. The smallest entry size is 8 bytes. The
  next one is 16 bytes etc. So, if the user requests 6 bytes, he gets 
  8 byte entry. If we asks for 25 bytes, he gets 32 byte entry etc.
  The input parameters "nSteps" says how many fixed memory managers
  are employed internally. Calling this procedure with nSteps equal
  to 10 results in 10 hierarchically arranged internal memory managers, 
  which can allocate up to 4096 (1Kb) entries. Requests for larger 
  entries are handed over to malloc() and then ABC_FREE()ed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmStepStart(nSteps int32) *Aig_MmStep_t {
    var p *Aig_MmStep_t;
    var i, k int32;
    // inlining of macro ABC_ALLOC
    p = ((Aig_MmStep_t *) malloc(sizeof(Aig_MmStep_t) * (size_t)(1)));
    memset( p, 0, sizeof(Aig_MmStep_t) );
    p.nMems = nSteps;
    // start the fixed memory managers
    // inlining of macro ABC_ALLOC
    p.pMems = ((Aig_MmFixed_t * *) malloc(sizeof(Aig_MmFixed_t *) * (size_t)(p.nMems)));
    for  i = 0; i < p.nMems; i++  {
        p.pMems[i] = Aig_MmFixedStart( (8<<i), (1<<13) );
    }
    // set up the mapping of the required memory size into the corresponding manager
    p.nMapSize = (4<<p.nMems);
    // inlining of macro ABC_ALLOC
    p.pMap = ((Aig_MmFixed_t * *) malloc(sizeof(Aig_MmFixed_t *) * (size_t)(p.nMapSize+1)));
    p.pMap[0] = nil;
    for  k = 1; k <= 4; k++  {
        p.pMap[k] = p.pMems[0];
    }
    for  i = 0; i < p.nMems; i++  {
        for  k = (4<<i)+1; k <= (8<<i); k++  {
            p.pMap[k] = p.pMems[i];
        }
    }
//for ( i = 1; i < 100; i ++ )
//printf( "%10d: size = %10d\n", i, p->pMap[i]->nEntrySize );
    p.nChunksAlloc  = 64;
    p.nChunks       = 0;
    // inlining of macro ABC_ALLOC
    p.pChunks       = ((char * *) malloc(sizeof(char *) * (size_t)(p.nChunksAlloc)));
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmStepStop(p *Aig_MmStep_t, fVerbose int32) {
    var i int32;
    for  i = 0; i < p.nMems; i++  {
        Aig_MmFixedStop( p.pMems[i], fVerbose );
    }
    if p.nChunksAlloc {
        for  i = 0; i < p.nChunks; i++  {
            // inlining of macro ABC_FREE
            ((p.pChunks[i]) ? (free((char *) (p.pChunks[i])), (p.pChunks[i]) = 0) : 0);
        }
        // inlining of macro ABC_FREE
        ((p.pChunks) ? (free((char *) (p.pChunks)), (p.pChunks) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((p.pMems) ? (free((char *) (p.pMems)), (p.pMems) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMap) ? (free((char *) (p.pMap)), (p.pMap) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Creates the entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmStepEntryFetch(p *Aig_MmStep_t, nBytes int32) string {
    if nBytes == 0 {
        return nil; }
#ifdef ABC_MEMALIGN
    // extend size to max alignment
    nBytes += (ABC_MEMALIGN - nBytes % ABC_MEMALIGN) % ABC_MEMALIGN;
#endif
    if nBytes > p.nMapSize {
        if p.nChunks == p.nChunksAlloc {
            p.nChunksAlloc *= 2;
            // inlining of macro ABC_REALLOC
            
            if p.pChunks {
                p.pChunks = ((char * *) realloc((char *)(p.pChunks), sizeof(char *) * (size_t)(p.nChunksAlloc)));
            } else {
                p.pChunks = ((char * *) malloc(sizeof(char *) * (size_t)(p.nChunksAlloc)));
            } 
        }
        // inlining of macro ABC_ALLOC
        p.pChunks[ p.nChunks++ ] = ((char *) malloc(sizeof(char) * (size_t)(nBytes)));
        return p.pChunks[p.nChunks-1];
    }
    return Aig_MmFixedEntryFetch( p.pMap[nBytes] );
}


/**Function*************************************************************

  Synopsis    [Recycles the entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmStepEntryRecycle(p *Aig_MmStep_t, pEntry string, nBytes int32) {
    if nBytes == 0 {
        return; }
#ifdef ABC_MEMALIGN
    // extend size to max alignment
    nBytes += (ABC_MEMALIGN - nBytes % ABC_MEMALIGN) % ABC_MEMALIGN;
#endif
    if nBytes > p.nMapSize {
//        ABC_FREE( pEntry );
        return;
    }
    Aig_MmFixedEntryRecycle( p.pMap[nBytes], pEntry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_MmStepReadMemUsage(p *Aig_MmStep_t) int32 {
    var i, nMemTotal int32;
    for  i = 0; i < p.nMems; i++  {
        nMemTotal += p.pMems[i].nMemoryAlloc;
    }
    return nMemTotal;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

