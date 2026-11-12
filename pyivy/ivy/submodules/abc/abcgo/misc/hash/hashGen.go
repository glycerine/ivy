package hash

import "strings"

/**CFile****************************************************************

  FileName    [vecGen.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hash maps.]

  Synopsis    [Hash maps.]

  Author      [Aaron P. Hurst, Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Jan 26, 2011.]

  Revision    [$Id: vecGen.h,v 1.00 2005/06/20 00:00:00 ahurst Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Hash_Gen_t =       Hash_Gen_t_;
   

type Hash_Gen_Entry_t struct {
  key                           string;
  data                           any;
  pNext       *Hash_Gen_Entry_t;
}

typedef int (*Hash_GenHashFunction_t)(void* key, int nBins);
typedef int (*Hash_GenCompFunction_t)(void* key, void* data);

type Hash_Gen_t_ struct {
  nSize                             int32;
  nBins                             int32;
  fHash          Hash_GenHashFunction_t;
  fComp          Hash_GenCompFunction_t;
  fFreeKey                             int32;
  pArray              **Hash_Gen_Entry_t;
}



////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Hash_GenForEachEntry( pHash, pEntry, bin )   \
//   for(bin=-1, pEntry=NULL; bin < pHash->nBins; (!pEntry)?(pEntry=pHash->pArray[++bin]):(pEntry=pEntry->pNext)) \
//     if (pEntry)


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Default hash function for strings.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_DefaultHashFuncStr(key any, nBins int32) int32 {
    char* p = (char*)key;
    var h int32;

    for  ; *p ; ++p  {
        h += h*5 + *p;
    }
  
    return uint32(h) % nBins; 
}

func Hash_DefaultCmpFuncStr(key1 any, key2 any) int32 {
    return int32(strings.Compare((char*)key1, (char*) key2));
}

/**Function*************************************************************

  Synopsis    [Default hash function for (long) integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_DefaultHashFuncInt(key any, nBins int32) int32 {
    return int64(key) % nBins;
}

/**Function*************************************************************

  Synopsis    [Default comparison function for (long) integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_DefaultCmpFuncInt(key1 any, key2 any) int32 {
    return int64(key1) - int64(key2);
}

/**Function*************************************************************

  Synopsis    [Allocates a hash map with the given number of bins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_GenAlloc(nBins int32, Hash_FuncHash func(any, int32) int32, Hash_FuncComp func(any, any) int32, fFreeKey int32) *Hash_Gen_t {
  
  assert(nBins > 0, "nBins > 0");
  // inlining of macro ABC_CALLOC
  p := &Hash_Gen_t{};
  p.nBins  = nBins;
  if Hash_FuncHash {
      p.fHash = Hash_FuncHash;
  } else {
      p.fHash = (int (*)(void *, int))Hash_DefaultHashFuncStr;
  }
  if Hash_FuncComp {
      p.fComp = Hash_FuncComp;
  } else {
      p.fComp = (int (*)(void *, void *))Hash_DefaultCmpFuncStr;
  }
  p.fFreeKey = fFreeKey;
  p.nSize  = 0;
  // inlining of macro ABC_CALLOC
  p.pArray = ((Hash_Gen_Entry_t * *) calloc((size_t)(nBins), sizeof(Hash_Gen_Entry_t *)));
  return p;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if a key already exists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_GenExists(p *Hash_Gen_t, key any) int32 {
  var bin int32;
  var pEntry *Hash_Gen_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if !p.fComp(pEntry.key,key) {
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
func Hash_GenWriteEntry(p *Hash_Gen_t, key any, data any) {
  var bin int32;
  var pEntry *Hash_Gen_Entry_t;
  var pLast **Hash_Gen_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pLast = &(p.pArray[bin]);
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if !p.fComp(pEntry.key,key) {
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
  (*pLast) = pEntry = ((Hash_Gen_Entry_t *) malloc(sizeof(Hash_Gen_Entry_t) * (size_t)(1)));
  pEntry.pNext = nil;
  pEntry.key  = (char *)key;
  pEntry.data = data;

  return;
}


/**Function*************************************************************

  Synopsis    [Finds or creates an entry with a key.]

  Description [fCreate specifies whether a new entry should be created.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_GenEntry(p *Hash_Gen_t, key any, fCreate int32) *Hash_Gen_Entry_t {
  var bin int32;
  var pEntry *Hash_Gen_Entry_t;
  var pLast **Hash_Gen_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pLast = &(p.pArray[bin]);
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if !p.fComp(pEntry.key,key) {
      return pEntry; }
    pLast = &(pEntry.pNext);
    pEntry = pEntry.pNext;
  }

  // this key does not currently exist
  if fCreate != 0 {
    // create a new entry and add to bin
    p.nSize++;
    // inlining of macro ABC_ALLOC
    (*pLast) = pEntry = ((Hash_Gen_Entry_t *) malloc(sizeof(Hash_Gen_Entry_t) * (size_t)(1)));
    pEntry.pNext = nil;
    pEntry.key  = (char *)key;
    pEntry.data = nil;
    return pEntry;
  }

  return nil;
}

/**Function*************************************************************

  Synopsis    [Deletes an entry.]

  Description [Returns data, if there was any.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_GenRemove(p *Hash_Gen_t, key any) any {
  var bin int32;
  var data any;
  var pEntry *Hash_Gen_Entry_t;
  var pLast **Hash_Gen_Entry_t;

  // find the bin where this key would live
  bin = (*(p.fHash))(key, p.nBins);

  // search for key
  pLast = &(p.pArray[bin]);
  pEntry = p.pArray[bin];
  for pEntry != nil {
    if !p.fComp(pEntry.key,key) {
      p.nSize--;
      data = pEntry.data;
      *pLast = pEntry.pNext;
      if p.fFreeKey {
        // inlining of macro ABC_FREE
        ((pEntry.key) ? (free((char *) (pEntry.key)), (pEntry.key) = 0) : 0); }
      // inlining of macro ABC_FREE
      ((pEntry) ? (free((char *) (pEntry)), (pEntry) = 0) : 0);
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
func Hash_GenFree(p *Hash_Gen_t) {
  var bin int32;
  var pEntry, pTemp *Hash_Gen_Entry_t;

  // free bins
  for bin = 0; bin < p.nBins; bin++ {
    pEntry = p.pArray[bin];
    for pEntry != nil {
      pTemp = pEntry;
      if p.fFreeKey {
        // inlining of macro ABC_FREE
        ((pTemp.key) ? (free((char *) (pTemp.key)), (pTemp.key) = 0) : 0); }
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


