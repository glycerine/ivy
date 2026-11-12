package hash

/**CFile****************************************************************

  FileName    [hashFlt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hash maps.]

  Synopsis    [Hash maps.]

  Author      [Aaron P. Hurst]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 16, 2006.]

  Revision    [$Id: vecInt.h,v 1.00 2005/06/20 00:00:00 ahurst Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_HEADER_START


// extern int Hash_DefaultHashFunc(int key, int nBins);


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Hash_Ptr_t =       Hash_Ptr_t_;
   

type Hash_Ptr_Entry_t struct {
  key                              int32;
  data                            any;
  pNext         *Hash_Ptr_Entry_t;
}

type Hash_Ptr_t_ struct {
  nSize                             int32;
  nBins                             int32;
  int (* fHash)(int key, int nBins);
  pArray              **Hash_Ptr_Entry_t;
}



////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Hash_PtrForEachEntry( pHash, pEntry, bin )   \
//   for(bin=-1, pEntry=NULL; bin < pHash->nBins; (!pEntry)?(pEntry=pHash->pArray[++bin]):(pEntry=pEntry->pNext)) \
//     if (pEntry)


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a hash map with the given number of bins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_PtrAlloc(nBins int32) *Hash_Ptr_t {
  var p *Hash_Ptr_t;
  var i int32;
  assert(nBins > 0, "nBins > 0");
  // inlining of macro ABC_ALLOC
  p = ((Hash_Ptr_t *) malloc(sizeof(Hash_Ptr_t) * (size_t)(1)));
  p.nBins = nBins;
  p.fHash = Hash_DefaultHashFunc;
  p.nSize  = 0;
  // inlining of macro ABC_ALLOC
  p.pArray = ((Hash_Ptr_Entry_t * *) malloc(sizeof(Hash_Ptr_Entry_t *) * (size_t)(nBins)));
  for i=0; i<nBins; i++ {
    p.pArray[i] = nil;
  }

  return p;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if a key already exists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_PtrExists(p *Hash_Ptr_t, key int32) int32 {
  var bin int32;
  var pEntry *Hash_Ptr_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if pEntry.key == key {
      return 1;
    }
    pEntry = pEntry.pNext;
  }

  return 0;
}

/**Function*************************************************************

  Synopsis    [Finds or creates an entry with a key and writes value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_PtrWriteEntry(p *Hash_Ptr_t, key int32, data any) {
  var bin int32;
  var pEntry *Hash_Ptr_Entry_t;
  var pLast **Hash_Ptr_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pLast = &(p.pArray[bin]);
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if pEntry.key == key {
      pEntry.data = data;
      return;
    }
    pLast = &(pEntry.pNext);
    pEntry = pEntry.pNext;
  }

  // this key does not currently exist
  // create a new entry and add to bin
  p.nSize++;
  // inlining of macro ABC_ALLOC
  (*pLast) = pEntry = ((Hash_Ptr_Entry_t *) malloc(sizeof(Hash_Ptr_Entry_t) * (size_t)(1)));
  pEntry.pNext = nil;
  pEntry.key = key;
  pEntry.data = data;

  return;
}


/**Function*************************************************************

  Synopsis    [Finds or creates an entry with a key.]

  Description [fCreate specifies whether a new entry should be created.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_PtrEntry(p *Hash_Ptr_t, key int32, fCreate int32) any {
  var bin int32;
  var pEntry *Hash_Ptr_Entry_t;
  var pLast **Hash_Ptr_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pLast = &(p.pArray[bin]);
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if pEntry.key == key {
      return pEntry.data; }
    pLast = &(pEntry.pNext);
    pEntry = pEntry.pNext;
  }

  // this key does not currently exist
  if fCreate != 0 {
    // create a new entry and add to bin
    p.nSize++;
    // inlining of macro ABC_ALLOC
    (*pLast) = pEntry = ((Hash_Ptr_Entry_t *) malloc(sizeof(Hash_Ptr_Entry_t) * (size_t)(1)));
    pEntry.pNext = nil;
    pEntry.key = key;
    pEntry.data = nil;
    return pEntry.data;
  }

  return nil;
}


/**Function*************************************************************

  Synopsis    [Finds or creates an entry with a key and returns the pointer to it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_PtrEntryPtr(p *Hash_Ptr_t, key int32) *any {
  var bin int32;
  var pEntry *Hash_Ptr_Entry_t;
  var pLast **Hash_Ptr_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pLast = &(p.pArray[bin]);
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if pEntry.key == key {
      return &(pEntry.data); }
    pLast = &(pEntry.pNext);
    pEntry = pEntry.pNext;
  }

  // this key does not currently exist
  // create a new entry and add to bin
  p.nSize++;
  // inlining of macro ABC_ALLOC
  (*pLast) = pEntry = ((Hash_Ptr_Entry_t *) malloc(sizeof(Hash_Ptr_Entry_t) * (size_t)(1)));
  pEntry.pNext = nil;
  pEntry.key = key;
  pEntry.data = nil;

  return &(pEntry.data);
}

/**Function*************************************************************

  Synopsis    [Deletes an entry.]

  Description [Returns data, if there was any.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_PtrRemove(p *Hash_Ptr_t, key int32) any {
  var bin int32;
  var data any;
  var pEntry *Hash_Ptr_Entry_t;
  var pLast **Hash_Ptr_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pLast = &(p.pArray[bin]);
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if pEntry.key == key {
      p.nSize--;
      data = pEntry.data;
      *pLast = pEntry.pNext;
      return data;
    }
    pLast = &(pEntry.pNext);
    pEntry = pEntry.pNext;
  }
    
  // could not find key
  return nil;
}

/**Function*************************************************************

  Synopsis    [Frees the hash.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_PtrFree(p *Hash_Ptr_t) {
  var bin int32;
  var pEntry, pTemp *Hash_Ptr_Entry_t;

  // free bins
  for bin = 0; bin < p.nBins; bin++ {
    pEntry = p.pArray[bin];
    for pEntry != nil {
      pTemp = pEntry;
      pEntry = pEntry.pNext;
      // inlining of macro ABC_FREE
      ((pTemp) ? (free((char *) (pTemp)), (pTemp) = 0) : 0);
    }
  }
 
  // free hash
  // inlining of macro ABC_FREE
  ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
  // inlining of macro ABC_FREE
  ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_END


