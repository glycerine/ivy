package fraig

/**CFile****************************************************************

  FileName    [fraigVec.c]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Vector of FRAIG nodes.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigVec.c,v 1.7 2005/07/08 01:01:34 alanmi Exp $]

***********************************************************************/

//#include "fraigInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecAlloc(nCap int32) *Fraig_NodeVec_t {
    var p *Fraig_NodeVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Fraig_NodeVec_t *) malloc(sizeof(Fraig_NodeVec_t) * (size_t)(1)));
    if nCap > 0 && nCap < 8 {
        nCap = 8; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        // inlining of macro ABC_ALLOC
        p.pArray = ((Fraig_Node_t * *) malloc(sizeof(Fraig_Node_t *) * (size_t)(p.nCap)));
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
func Fraig_NodeVecFree(p *Fraig_NodeVec_t) {
    // inlining of macro ABC_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Duplicates the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecDup(pVec *Fraig_NodeVec_t) *Fraig_NodeVec_t {
    var p *Fraig_NodeVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Fraig_NodeVec_t *) malloc(sizeof(Fraig_NodeVec_t) * (size_t)(1)));
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    if p.nCap {
        // inlining of macro ABC_ALLOC
        p.pArray = ((Fraig_Node_t * *) malloc(sizeof(Fraig_Node_t *) * (size_t)(p.nCap)));
    } else {
        p.pArray = nil;
    }
    memcpy( p.pArray, pVec.pArray, sizeof(Fraig_Node_t *) * pVec.nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecReadArray(p *Fraig_NodeVec_t) **Fraig_Node_t {
    return p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecReadSize(p *Fraig_NodeVec_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecGrow(p *Fraig_NodeVec_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro ABC_REALLOC
    
    if p.pArray {
        p.pArray = ((Fraig_Node_t * *) realloc((char *)(p.pArray), sizeof(Fraig_Node_t *) * (size_t)(nCapMin)));
    } else {
        p.pArray = ((Fraig_Node_t * *) malloc(sizeof(Fraig_Node_t *) * (size_t)(nCapMin)));
    } 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecShrink(p *Fraig_NodeVec_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecClear(p *Fraig_NodeVec_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecPush(p *Fraig_NodeVec_t, Entry *Fraig_Node_t) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Fraig_NodeVecGrow( p, 16 );
        } else {
            Fraig_NodeVecGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    [Add the element while ensuring uniqueness.]

  Description [Returns 1 if the element was found, and 0 if it was new. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecPushUnique(p *Fraig_NodeVec_t, Entry *Fraig_Node_t) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Fraig_NodeVecPush( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Inserts a new node in the order by arrival times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecPushOrder(p *Fraig_NodeVec_t, pNode *Fraig_Node_t) {
    var pNode1, pNode2 *Fraig_Node_t;
    var i int32;
    Fraig_NodeVecPush( p, pNode );
    // find the p of the node
    for  i = p.nSize-1; i > 0; i--  {
        pNode1 = p.pArray[i  ];
        pNode2 = p.pArray[i-1];
        if pNode1 >= pNode2 {
            break; }
        p.pArray[i  ] = pNode2;
        p.pArray[i-1] = pNode1;
    }
}

/**Function*************************************************************

  Synopsis    [Add the element while ensuring uniqueness in the order.]

  Description [Returns 1 if the element was found, and 0 if it was new. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecPushUniqueOrder(p *Fraig_NodeVec_t, pNode *Fraig_Node_t) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == pNode {
            return 1; }
    }
    Fraig_NodeVecPushOrder( p, pNode );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Inserts a new node in the order by arrival times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecPushOrderByLevel(p *Fraig_NodeVec_t, pNode *Fraig_Node_t) {
    var pNode1, pNode2 *Fraig_Node_t;
    var i int32;
    Fraig_NodeVecPush( p, pNode );
    // find the p of the node
    for  i = p.nSize-1; i > 0; i--  {
        pNode1 = p.pArray[i  ];
        pNode2 = p.pArray[i-1];
        // inlining of macro Fraig_Regular
        if  ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1) & ~01)) .Level <= ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2) & ~01)) .Level {
            break; }
        p.pArray[i  ] = pNode2;
        p.pArray[i-1] = pNode1;
    }
}

/**Function*************************************************************

  Synopsis    [Add the element while ensuring uniqueness in the order.]

  Description [Returns 1 if the element was found, and 0 if it was new. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecPushUniqueOrderByLevel(p *Fraig_NodeVec_t, pNode *Fraig_Node_t) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == pNode {
            return 1; }
    }
    Fraig_NodeVecPushOrderByLevel( p, pNode );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecPop(p *Fraig_NodeVec_t) *Fraig_Node_t {
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecRemove(p *Fraig_NodeVec_t, Entry *Fraig_Node_t) {
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
func Fraig_NodeVecWriteEntry(p *Fraig_NodeVec_t, i int32, Entry *Fraig_Node_t) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecReadEntry(p *Fraig_NodeVec_t, i int32) *Fraig_Node_t {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecCompareLevelsIncreasing(pp1 **Fraig_Node_t, pp2 **Fraig_Node_t) int32 {
    // inlining of macro Fraig_Regular
    var Level1 int32  = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Level;
    // inlining of macro Fraig_Regular
    var Level2 int32  = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Level;
    if Level1 < Level2 {
        return -1; }
    if Level1 > Level2 {
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecCompareLevelsDecreasing(pp1 **Fraig_Node_t, pp2 **Fraig_Node_t) int32 {
    // inlining of macro Fraig_Regular
    var Level1 int32  = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Level;
    // inlining of macro Fraig_Regular
    var Level2 int32  = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Level;
    if Level1 > Level2 {
        return -1; }
    if Level1 < Level2 {
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecCompareNumbers(pp1 **Fraig_Node_t, pp2 **Fraig_Node_t) int32 {
    // inlining of macro Fraig_Regular
    var Num1 int32  = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Num;
    // inlining of macro Fraig_Regular
    var Num2 int32  = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Num;
    if Num1 < Num2 {
        return -1; }
    if Num1 > Num2 {
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecCompareRefCounts(pp1 **Fraig_Node_t, pp2 **Fraig_Node_t) int32 {
    // inlining of macro Fraig_Regular
    var nRefs1 int32  = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .nRefs;
    // inlining of macro Fraig_Regular
    var nRefs2 int32  = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .nRefs;

    if nRefs1 < nRefs2 {
        return -1; }
    if nRefs1 > nRefs2 {
        return 1; }

    // inlining of macro Fraig_Regular
    nRefs1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Level;
    // inlining of macro Fraig_Regular
    nRefs2 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Level;

    if nRefs1 < nRefs2 {
        return -1; }
    if nRefs1 > nRefs2 {
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecSortByLevel(p *Fraig_NodeVec_t, fIncreasing int32) {
    if fIncreasing != 0 {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Fraig_Node_t *), 
                        (int (*)(void *, void *)) Fraig_NodeVecCompareLevelsIncreasing );
    } else { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Fraig_Node_t *), 
                        (int (*)(void *, void *)) Fraig_NodeVecCompareLevelsDecreasing ); }
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecSortByNumber(p *Fraig_NodeVec_t) {
    qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Fraig_Node_t *), 
                (int (*)(void *, void *)) Fraig_NodeVecCompareNumbers );
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeVecSortByRefCount(p *Fraig_NodeVec_t) {
    qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Fraig_Node_t *), 
                (int (*)(void *, void *)) Fraig_NodeVecCompareRefCounts );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

