package fpga

/**CFile****************************************************************

  FileName    [fpgaVec.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Technology mapping for variable-size-LUT FPGAs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaVec.c,v 1.3 2005/01/23 06:59:42 alanmi Exp $]

***********************************************************************/

//#include "fpgaInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Fpga_NodeVecCompareLevels( Fpga_Node_t ** pp1, Fpga_Node_t ** pp2 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecAlloc(nCap int32) *Fpga_NodeVec_t {
    var p *Fpga_NodeVec_t;
    // inlining of macro ABC_ALLOC
    p = ((Fpga_NodeVec_t *) malloc(sizeof(Fpga_NodeVec_t) * (size_t)(1)));
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        // inlining of macro ABC_ALLOC
        p.pArray = ((Fpga_Node_t * *) malloc(sizeof(Fpga_Node_t *) * (size_t)(p.nCap)));
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
func Fpga_NodeVecFree(p *Fpga_NodeVec_t) {
    // inlining of macro ABC_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecReadArray(p *Fpga_NodeVec_t) **Fpga_Node_t {
    return p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecReadSize(p *Fpga_NodeVec_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecGrow(p *Fpga_NodeVec_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    // inlining of macro ABC_REALLOC
    
    if p.pArray {
        p.pArray = ((Fpga_Node_t * *) realloc((char *)(p.pArray), sizeof(Fpga_Node_t *) * (size_t)(nCapMin)));
    } else {
        p.pArray = ((Fpga_Node_t * *) malloc(sizeof(Fpga_Node_t *) * (size_t)(nCapMin)));
    } 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecShrink(p *Fpga_NodeVec_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecClear(p *Fpga_NodeVec_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecPush(p *Fpga_NodeVec_t, Entry *Fpga_Node_t) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Fpga_NodeVecGrow( p, 16 );
        } else {
            Fpga_NodeVecGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    [Add the element while ensuring uniqueness.]

  Description [Returns 1 if the element was found, and 0 if it was new. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecPushUnique(p *Fpga_NodeVec_t, Entry *Fpga_Node_t) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Fpga_NodeVecPush( p, Entry );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecPop(p *Fpga_NodeVec_t) *Fpga_Node_t {
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecWriteEntry(p *Fpga_NodeVec_t, i int32, Entry *Fpga_Node_t) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecReadEntry(p *Fpga_NodeVec_t, i int32) *Fpga_Node_t {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecCompareLevels(pp1 **Fpga_Node_t, pp2 **Fpga_Node_t) int32 {
    // inlining of macro Fpga_Regular
    var Level1 int32  = ((Fpga_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Level;
    // inlining of macro Fpga_Regular
    var Level2 int32  = ((Fpga_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Level;
    if Level1 < Level2 {
        return -1; }
    if Level1 > Level2 {
        return 1; }
    // inlining of macro Fpga_Regular
    if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Num < ((Fpga_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Num {
        return -1; }
    // inlining of macro Fpga_Regular
    if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(*pp1) & ~01)) .Num > ((Fpga_Node_t *)((ABC_PTRUINT_T)(*pp2) & ~01)) .Num {
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecSortByLevel(p *Fpga_NodeVec_t) {
    qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Fpga_Node_t *), 
                (int (*)(void *, void *)) Fpga_NodeVecCompareLevels );
}

/**Function*************************************************************

  Synopsis    [Compares the arrival times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecCompareArrivals(ppS1 **Fpga_Node_t, ppS2 **Fpga_Node_t) int32 {
    if (*ppS1).pCutBest.tArrival < (*ppS2).pCutBest.tArrival {
        return -1; }
    if (*ppS1).pCutBest.tArrival > (*ppS2).pCutBest.tArrival {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Orders the nodes in the increasing order of the arrival times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_SortNodesByArrivalTimes(p *Fpga_NodeVec_t) {
    qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Fpga_Node_t *), 
                (int (*)(void *, void *)) Fpga_NodeVecCompareArrivals );
//    assert( Fpga_CompareNodesByLevel( p->pArray, p->pArray + p->nSize - 1 ) <= 0 );
}


/**Function*************************************************************

  Synopsis    [Computes the union of nodes in two arrays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecUnion(p *Fpga_NodeVec_t, p1 *Fpga_NodeVec_t, p2 *Fpga_NodeVec_t) {
    var i int32;
    Fpga_NodeVecClear( p );
    for  i = 0; i < p1.nSize; i++  {
        Fpga_NodeVecPush( p, p1.pArray[i] );
    }
    for  i = 0; i < p2.nSize; i++  {
        Fpga_NodeVecPush( p, p2.pArray[i] );
    }
}

/**Function*************************************************************

  Synopsis    [Inserts a new node in the order by arrival times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecPushOrder(vNodes *Fpga_NodeVec_t, pNode *Fpga_Node_t, fIncreasing int32) {
    var pNode1, pNode2 *Fpga_Node_t;
    var i int32;
    Fpga_NodeVecPush( vNodes, pNode );
    // find the place of the node
    for  i = vNodes.nSize-1; i > 0; i--  {
        pNode1 = vNodes.pArray[i  ];
        pNode2 = vNodes.pArray[i-1];
        if ( (fIncreasing != 0) && pNode1.pCutBest.tArrival >= pNode2.pCutBest.tArrival) ||
            (!(fIncreasing != 0) && pNode1.pCutBest.tArrival <= pNode2.pCutBest.tArrival) {
            break; }
        vNodes.pArray[i  ] = pNode2;
        vNodes.pArray[i-1] = pNode1;
    }
}

/**Function*************************************************************

  Synopsis    [Inserts a new node in the order by arrival times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeVecReverse(vNodes *Fpga_NodeVec_t) {
    var pNode1, pNode2 *Fpga_Node_t;
    var i int32;
    for  i = 0; i < vNodes.nSize/2; i++  {
        pNode1 = vNodes.pArray[i];
        pNode2 = vNodes.pArray[vNodes.nSize-1-i];
        vNodes.pArray[i]                 = pNode2;
        vNodes.pArray[vNodes.nSize-1-i] = pNode1;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

