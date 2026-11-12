package mapper

/**CFile****************************************************************

  FileName    [mapperVec.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperVec.c,v 1.3 2005/01/23 06:59:45 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Map_NodeVecCompareLevels( Map_Node_t ** pp1, Map_Node_t ** pp2 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecAlloc(nCap int32) *Map_NodeVec_t {
    var p *Map_NodeVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Map_NodeVec_t *) malloc(sizeof(Map_NodeVec_t) * (size_t)(1)));
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        // inlining of macro ABC_ALLOC
        p.pArray = ((Map_Node_t * *) malloc(sizeof(Map_Node_t *) * (size_t)(p.nCap)));
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecFree(p *Map_NodeVec_t) {
    if p == nil {
        return; }
    p.pArray = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecDup(p *Map_NodeVec_t) *Map_NodeVec_t {
    var pNew *Map_NodeVec_t  = Map_NodeVecAlloc( p.nSize );
    memcpy( pNew.pArray, p.pArray, sizeof(int) * p.nSize );
    pNew.nSize = p.nSize;
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecReadArray(p *Map_NodeVec_t) **Map_Node_t {
    return p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecReadSize(p *Map_NodeVec_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecGrow(p *Map_NodeVec_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro ABC_REALLOC
    
    if p.pArray {
        p.pArray = ((Map_Node_t * *) realloc((char *)(p.pArray), sizeof(Map_Node_t *) * (size_t)(nCapMin)));
    } else {
        p.pArray = ((Map_Node_t * *) malloc(sizeof(Map_Node_t *) * (size_t)(nCapMin)));
    } 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecShrink(p *Map_NodeVec_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecClear(p *Map_NodeVec_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecPush(p *Map_NodeVec_t, Entry *Map_Node_t) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Map_NodeVecGrow( p, 16 );
        } else {
            Map_NodeVecGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    [Add the element while ensuring uniqueness.]

  Description [Returns 1 if the element was found, and 0 if it was new. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecPushUnique(p *Map_NodeVec_t, Entry *Map_Node_t) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Map_NodeVecPush( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecPop(p *Map_NodeVec_t) *Map_Node_t {
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecRemove(p *Map_NodeVec_t, Entry *Map_Node_t) {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            break; }
    }
    assert(i < p.nSize, "i < p.nSize");
    for  i++; i < p.nSize; i++  {
        p.pArray[i-1] = p.pArray[i];
    }
    p.nSize--;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecWriteEntry(p *Map_NodeVec_t, i int32, Entry *Map_Node_t) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecReadEntry(p *Map_NodeVec_t, i int32) *Map_Node_t {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecSortByLevel(p *Map_NodeVec_t) {
    qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Map_Node_t *), 
                (int (*)(void *, void *)) Map_NodeVecCompareLevels );
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeVecCompareLevels(pp1 **Map_Node_t, pp2 **Map_Node_t) int32 {
    // inlining of macro Map_Regular
    var Level1 int32  = ((Map_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Level;
    // inlining of macro Map_Regular
    var Level2 int32  = ((Map_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Level;
    if Level1 < Level2 {
        return -1; }
    if Level1 > Level2 {
        return 1; }
    // inlining of macro Map_Regular
    if  ((Map_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Num < ((Map_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Num {
        return -1; }
    // inlining of macro Map_Regular
    if  ((Map_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Num > ((Map_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Num {
        return 1; }
    return 0; 
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

