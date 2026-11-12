package util

/**CFile****************************************************************

  FileName    [utilMem.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Memory recycling utilities.]

  Synopsis    [Memory recycling utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: utilMem.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//#include <assert.h>

//#include "abc_global.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 

         
type Vec_Mem_t struct {
    nCap              int32;
    nSize              int32;
    pArray           *any;
}

var s_vAllocs any = nil;
var s_vFrees any = nil;
var s_fInterrupt int32;

// Original macro:
// #define ABC_MEM_ALLOC(type, num)     ((type *) malloc(sizeof(type) * (num)))

// Original macro:
// #define ABC_MEM_CALLOC(type, num)     ((type *) calloc((num), sizeof(type)))

// Original macro:
// #define ABC_MEM_FALLOC(type, num)     ((type *) memset(malloc(sizeof(type) * (num)), 0xff, sizeof(type) * (num)))

// Original macro:
// #define ABC_MEM_FREE(obj)             ((obj) ? (free((char *) (obj)), (obj) = 0) : 0)

// Original macro:
// #define ABC_MEM_REALLOC(type, obj, num)    \
//         ((obj) ? ((type *) realloc((char *)(obj), sizeof(type) * (num))) : \
//          ((type *) malloc(sizeof(type) * (num))))


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemAlloc(nCap int32) *Vec_Mem_t {
    var p *Vec_Mem_t;
    // inlining of macro ABC_MEM_ALLOC
    p = ((Vec_Mem_t *) malloc(sizeof(Vec_Mem_t) * (1)));
    if nCap > 0 && nCap < 8 {
        nCap = 8; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        // inlining of macro ABC_MEM_ALLOC
        p.pArray = ((void * *) malloc(sizeof(void *) * (p.nCap)));
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Frees the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemFree(p *Vec_Mem_t) {
    // inlining of macro ABC_MEM_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro ABC_MEM_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemGrow(p *Vec_Mem_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro ABC_MEM_REALLOC
    
    if p.pArray {
        p.pArray = ((void * *) realloc((char *)(p.pArray), sizeof(void *) * (nCapMin)));
    } else {
        p.pArray = ((void * *) malloc(sizeof(void *) * (nCapMin)));
    } 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemPush(p *Vec_Mem_t, Entry any) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_MemGrow( p, 16 );
        } else {
            Vec_MemGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemSort(p *Vec_Mem_t, Vec_MemSortCompare func() int32) {
    if p.nSize < 2 {
        return; }
    qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(void *), 
                (int (*)(void *, void *)) Vec_MemSortCompare );
}


/**Function*************************************************************

  Synopsis    [Remembers an allocated pointer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Util_MemRecAlloc(pMem any) any {
    if s_vAllocs {
        Vec_MemPush( (Vec_Mem_t *)s_vAllocs, pMem ); }
    return pMem;
}

/**Function*************************************************************

  Synopsis    [Remembers a deallocated pointer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Util_MemRecFree(pMem any) any {
    if s_vFrees {
        Vec_MemPush( (Vec_Mem_t *)s_vFrees, pMem ); }
    return pMem;
}

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in decreasing order of levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Util_ComparePointers(pp1 *any, pp2 *any) int32 {
    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Finds entries that do not appear in both lists.]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_MemTwoMerge(vArr1 *Vec_Mem_t, vArr2 *Vec_Mem_t) *Vec_Mem_t {
    var vArr *Vec_Mem_t  = Vec_MemAlloc( vArr1.nSize + vArr2.nSize ); 
    var pBeg *any   = vArr.pArray;
    var pBeg1 *any  = vArr1.pArray;
    var pBeg2 *any  = vArr2.pArray;
    var pEnd1 *any  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *any  = vArr2.pArray + vArr2.nSize;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            pBeg1++, pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            free( *pBeg1 );
            *pBeg++ = *pBeg1++;
        
        } else { 
            assert(false, "0"); }
//            *pBeg++ = *pBeg2++;
    }
    for  pBeg1 < pEnd1  {
        *pBeg++ = *pBeg1++;
    }
//    while ( pBeg2 < pEnd2 )
//        *pBeg++ = *pBeg2++;
    assert(pBeg2 >= pEnd2, "pBeg2 >= pEnd2");
    vArr.nSize = pBeg - vArr.pArray;
    assert(vArr.nSize <= vArr.nCap, "vArr.nSize <= vArr.nCap");
    assert(vArr.nSize >= vArr1.nSize, "vArr.nSize >= vArr1.nSize");
    assert(vArr.nSize >= vArr2.nSize, "vArr.nSize >= vArr2.nSize");
    return vArr;
}

/**Function*************************************************************

  Synopsis    [Recycles the accumulated memory.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Util_MemRecRecycle() {
    var vMerge *Vec_Mem_t;
    assert(s_vAllocs == nil, "s_vAllocs == nil");
    assert(s_vFrees == nil, "s_vFrees == nil");
    Vec_MemSort( (Vec_Mem_t *)s_vAllocs, (int (*)())Util_ComparePointers );
    Vec_MemSort( (Vec_Mem_t *)s_vFrees, (int (*)())Util_ComparePointers );
    vMerge = (Vec_Mem_t *)Vec_MemTwoMerge( (Vec_Mem_t *)s_vAllocs, (Vec_Mem_t *)s_vFrees );
    Vec_MemFree( vMerge );
}

/**Function*************************************************************

  Synopsis    [Starts memory structures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Util_MemRecStart() {
    assert(s_vAllocs == nil && s_vFrees == nil, "s_vAllocs == nil && s_vFrees == nil");
    s_vAllocs = Vec_MemAlloc( 1000 );
    s_vFrees = Vec_MemAlloc( 1000 );
}

/**Function*************************************************************

  Synopsis    [Quits memory structures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Util_MemRecQuit() {
    assert(s_vAllocs != nil && s_vFrees != nil, "s_vAllocs != nil && s_vFrees != nil");
    Vec_MemFree( (Vec_Mem_t *)s_vAllocs );
    Vec_MemFree( (Vec_Mem_t *)s_vFrees );
}

/**Function*************************************************************

  Synopsis    [Starts memory structures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Util_MemRecIsSet() int32 {
    return s_vAllocs != nil && s_vFrees != nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

