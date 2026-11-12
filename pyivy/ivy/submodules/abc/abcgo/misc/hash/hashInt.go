package hash

/**CFile****************************************************************

  FileName    [hashInt.h]

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

type Hash_Int_t =       Hash_Int_t_;
   

type Hash_Int_Entry_t struct {
  key                              int32;
  data                            int32;
  pNext         *Hash_Int_Entry_t;
}

type Hash_Int_t_ struct {
  nSize                             int32;
  nBins                             int32;
  int (* fHash)(int key, int nBins);
  pArray              **Hash_Int_Entry_t;
}



////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Hash_IntForEachEntry( pHash, pEntry, bin)   \
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
func Hash_IntAlloc(nBins int32) *Hash_Int_t {
  var p *Hash_Int_t;
  var i int32;
  assert(nBins > 0, "nBins > 0");
  // inlining of macro ABC_ALLOC
  p = ((Hash_Int_t *) malloc(sizeof(Hash_Int_t) * (size_t)(1)));
  p.nBins = nBins;
  p.fHash = Hash_DefaultHashFunc;
  p.nSize  = 0;
  // inlining of macro ABC_ALLOC
  p.pArray = ((Hash_Int_Entry_t * *) malloc(sizeof(Hash_Int_Entry_t *) * (size_t)(nBins+1)));
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
func Hash_IntExists(p *Hash_Int_t, key int32) int32 {
  var bin int32;
  var pEntry *Hash_Int_Entry_t;
  var pLast **Hash_Int_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pLast = &(p.pArray[bin]);
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if pEntry.key == key {
      return 1;
    }
    pLast = &(pEntry.pNext);
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
func Hash_IntWriteEntry(p *Hash_Int_t, key int32, data int32) {
  var bin int32;
  var pEntry *Hash_Int_Entry_t;
  var pLast **Hash_Int_Entry_t;

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
  (*pLast) = pEntry = ((Hash_Int_Entry_t *) malloc(sizeof(Hash_Int_Entry_t) * (size_t)(1)));
  pEntry.pNext = nil;
  pEntry.key = key;
  pEntry.data = data;

  return;
}


/**Function*************************************************************

  Synopsis    [Finds or creates an entry with a key.]

  Description [fCreate specifies whether new entries will be created.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_IntEntry(p *Hash_Int_t, key int32, fCreate int32) int32 {
  var bin int32;
  var pEntry *Hash_Int_Entry_t;
  var pLast **Hash_Int_Entry_t;

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
    (*pLast) = pEntry = ((Hash_Int_Entry_t *) malloc(sizeof(Hash_Int_Entry_t) * (size_t)(1)));
    pEntry.pNext = nil;
    pEntry.key = key;
    pEntry.data = 0;
    return pEntry.data;
  }

  return 0;
}


/**Function*************************************************************

  Synopsis    [Finds or creates an entry with a key and returns the pointer to it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_IntEntryPtr(p *Hash_Int_t, key int32) *int32 {
  var bin int32;
  var pEntry *Hash_Int_Entry_t;
  var pLast **Hash_Int_Entry_t;

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
  (*pLast) = pEntry = ((Hash_Int_Entry_t *) malloc(sizeof(Hash_Int_Entry_t) * (size_t)(1)));
  pEntry.pNext = nil;
  pEntry.key = key;
  pEntry.data = 0;

  return &(pEntry.data);
}

/**Function*************************************************************

  Synopsis    [Frees the hash.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_IntFree(p *Hash_Int_t) {
  var bin int32;
  var pEntry, pTemp *Hash_Int_Entry_t;

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


