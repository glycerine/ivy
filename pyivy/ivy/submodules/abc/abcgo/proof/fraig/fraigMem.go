package fraig

import "fmt"

/**CFile****************************************************************

  FileName    [fraigMem.c]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Fixed-size-entry memory manager for the FRAIG package.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigMem.c,v 1.4 2005/07/08 01:01:31 alanmi Exp $]

***********************************************************************/

//#include "fraigInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Fraig_MemFixed_t_ struct {
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

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the internal memory manager.]

  Description [Can only work with entry size at least 4 byte long.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MemFixedStart(nEntrySize int32) *Fraig_MemFixed_t {
    var p *Fraig_MemFixed_t;

    // inlining of macro ABC_ALLOC
    p = ((Fraig_MemFixed_t *) malloc(sizeof(Fraig_MemFixed_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fraig_MemFixed_t) );

    p.nEntrySize    = nEntrySize;
    p.nEntriesAlloc = 0;
    p.nEntriesUsed  = 0;
    p.pEntriesFree  = nil;

    if nEntrySize * (1 << 10) < (1<<16) {
        p.nChunkSize = (1 << 10);
    } else {
        p.nChunkSize = (1<<16) / nEntrySize; }
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

  Synopsis    [Stops the internal memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MemFixedStop(p *Fraig_MemFixed_t, fVerbose int32) {
    var i int32;
    if p == nil {
        return; }
    if fVerbose != 0 {
        fmt.Printf("Fixed memory manager: Entry = %5d. Chunk = %5d. Chunks used = %5d.\n", p.nEntrySize, p.nChunkSize, p.nChunks);
        fmt.Printf("   Entries used = %8d. Entries peak = %8d. Memory used = %8d. Memory alloc = %8d.\n", p.nEntriesUsed, p.nEntriesMax, p.nEntrySize * p.nEntriesUsed, p.nMemoryAlloc);
    }
    for  i = 0; i < p.nChunks; i++  {
        p.pChunks[i] = nil;
    }
    p.pChunks = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Extracts one entry from the memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MemFixedEntryFetch(p *Fraig_MemFixed_t) string {
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

  Synopsis    [Returns one entry into the memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MemFixedEntryRecycle(p *Fraig_MemFixed_t, pEntry string) {
    // decrement the counter of used entries
    p.nEntriesUsed--;
    // add the entry to the linked list of free entries
    *((char **)pEntry) = p.pEntriesFree;
    p.pEntriesFree = pEntry;
}

/**Function*************************************************************

  Synopsis    [Frees all associated memory and resets the manager.]

  Description [Relocates all the memory except the first chunk.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MemFixedRestart(p *Fraig_MemFixed_t) {
    var i int32;
    var pTemp string;

    // deallocate all chunks except the first one
    for  i = 1; i < p.nChunks; i++  {
        p.pChunks[i] = nil;
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

  Synopsis    [Reports the memory usage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MemFixedReadMemUsage(p *Fraig_MemFixed_t) int32 {
    return p.nMemoryAlloc;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

