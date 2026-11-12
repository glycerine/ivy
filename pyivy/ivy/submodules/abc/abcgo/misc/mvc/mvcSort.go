package mvc

/**CFile****************************************************************

  FileName    [mvcSort.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Sorting cubes in the cover with the mask.]

  Author      [MVSIS Group]
  
  Affiliation [uC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcSort.c,v 1.5 2003/04/27 01:03:45 wjiang Exp $]

***********************************************************************/

//#include "mvc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


// Mvc_Cube_t * Mvc_CoverSort_rec( Mvc_Cube_t * pList, int nItems, Mvc_Cube_t * pMask, int (* pCompareFunc)(Mvc_Cube_t *, Mvc_Cube_t *, Mvc_Cube_t *) );

// Mvc_Cube_t * Mvc_CoverSortMerge( Mvc_Cube_t * pList1, Mvc_Cube_t * pList2, Mvc_Cube_t * pMask, int (* pCompareFunc)(Mvc_Cube_t *, Mvc_Cube_t *, Mvc_Cube_t *) );


////////////////////////////////////////////////////////////////////////
///                     FuNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sorts cubes using the given cost function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSort(pCover *Mvc_Cover_t, pMask *Mvc_Cube_t, pCompareFunc func(*Mvc_Cube_t, *Mvc_Cube_t, *Mvc_Cube_t) int32) {
    var pHead *Mvc_Cube_t;
    var nCubes int32;
    // one cube does not need sorting
    nCubes = Mvc_CoverReadCubeNum(pCover);
    if nCubes <= 1 {
        return; }
    // sort the cubes 
    pHead = Mvc_CoverSort_rec( Mvc_CoverReadCubeHead(pCover), nCubes, pMask, pCompareFunc ); 
    // insert the sorted list into the cover
    Mvc_CoverSetCubeHead( pCover, pHead );
    Mvc_CoverSetCubeTail( pCover, Mvc_ListGetTailFromHead(pHead) );
    // make sure that the list is sorted in the increasing order
    assert(pCompareFunc( Mvc_CoverReadCubeHead(pCover), Mvc_CoverReadCubeTail(pCover), pMask ) <= 0, "pCompareFunc( Mvc_CoverReadCubeHead(pCover), Mvc_CoverReadCubeTail(pCover), pMask ) <= 0");
}

/**Function*************************************************************

  Synopsis    [Recursive part of Mvc_CoverSort()]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSort_rec(pList *Mvc_Cube_t, nItems int32, pMask *Mvc_Cube_t, pCompareFunc func(*Mvc_Cube_t, *Mvc_Cube_t, *Mvc_Cube_t) int32) *Mvc_Cube_t {
    var pList1, pList2 *Mvc_Cube_t;
    var nItems1, nItems2, i int32;

    // trivial case
    if nItems == 1 {
        // inlining of macro Mvc_CubeSetNext
        ((pList).pNext = (nil));
        return pList;
    }

    // select the new sizes
    nItems1 = nItems/2;
    nItems2 = nItems - nItems1;

    // set the new beginnings
    pList1 = pList2 = pList;
    for  i = 0; i < nItems1; i++  {
        // inlining of macro Mvc_CubeReadNext
        pList2 = ((pList2).pNext);
    }

    // solve recursively
    pList1 = Mvc_CoverSort_rec( pList1, nItems1, pMask, pCompareFunc );
    pList2 = Mvc_CoverSort_rec( pList2, nItems2, pMask, pCompareFunc );

    // merge
    return Mvc_CoverSortMerge( pList1, pList2, pMask, pCompareFunc );
}


/**Function*************************************************************

  Synopsis    [Merges two NULL-terminated linked lists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSortMerge(pList1 *Mvc_Cube_t, pList2 *Mvc_Cube_t, pMask *Mvc_Cube_t, pCompareFunc func(*Mvc_Cube_t, *Mvc_Cube_t, *Mvc_Cube_t) int32) *Mvc_Cube_t {
    var pList *Mvc_Cube_t = nil;
    var ppTail **Mvc_Cube_t = &pList;
    var pCube *Mvc_Cube_t;
    for  (pList1 != nil) && (pList2 != nil)  {
        if pCompareFunc( pList1, pList2, pMask ) < 0 {
            pCube = pList1;
            // inlining of macro Mvc_CubeReadNext
            pList1 = ((pList1).pNext);
        
        } else {
            pCube = pList2;
            // inlining of macro Mvc_CubeReadNext
            pList2 = ((pList2).pNext);
        }
        *ppTail = pCube;
        // inlining of macro Mvc_CubeReadNextP
        ppTail = (&(pCube).pNext);
    }
    *ppTail = pList1? pList1: pList2;
    return pList;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

