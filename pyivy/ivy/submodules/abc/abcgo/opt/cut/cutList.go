package cut

/**CFile****************************************************************

  FileName    [cutList.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Implementation of layered listed list of cuts.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cutList.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

           
type Cut_List_t struct {
    Cut_Cut_t *  pHead[CUT_SIZE_MAX+1];
    Cut_Cut_t ** ppTail[CUT_SIZE_MAX+1];
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start the cut list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ListStart(p *Cut_List_t) {
    var i int32;
    for  i = 1; i <= CUT_SIZE_MAX; i++  {
        p.pHead[i] = 0;
        p.ppTail[i] = &p.pHead[i];
    }
}

/**Function*************************************************************

  Synopsis    [Adds one cut to the cut list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ListAdd(p *Cut_List_t, pCut *Cut_Cut_t) {
    assert(pCut.nLeaves > 0 && pCut.nLeaves <= CUT_SIZE_MAX, "pCut.nLeaves > 0 && pCut.nLeaves <= CUT_SIZE_MAX");
    *p.ppTail[pCut.nLeaves] = pCut;
    p.ppTail[pCut.nLeaves] = &pCut.pNext;
}

/**Function*************************************************************

  Synopsis    [Adds one cut to the cut list while preserving order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ListAdd2(p *Cut_List_t, pCut *Cut_Cut_t) {
    extern int Cut_CutCompare( Cut_Cut_t * pCut1, Cut_Cut_t * pCut2 );
    var pTemp *Cut_Cut_t;
    var ppSpot **Cut_Cut_t;
    assert(pCut.nLeaves > 0 && pCut.nLeaves <= CUT_SIZE_MAX, "pCut.nLeaves > 0 && pCut.nLeaves <= CUT_SIZE_MAX");
    if p.pHead[pCut.nLeaves] != nil {
        ppSpot = &p.pHead[pCut.nLeaves];
        for  pTemp = p.pHead[pCut.nLeaves]; (pTemp != nil); pTemp = pTemp.pNext  {
            if Cut_CutCompare(pCut, pTemp) < 0 {
                *ppSpot = pCut;
                pCut.pNext = pTemp;
                return;
            
            } else {
                ppSpot = &pTemp.pNext; }
        }
    }
    *p.ppTail[pCut.nLeaves] = pCut;
    p.ppTail[pCut.nLeaves] = &pCut.pNext;
}

/**Function*************************************************************

  Synopsis    [Derive the super list from the linked list of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ListDerive(p *Cut_List_t, pList *Cut_Cut_t) {
    var pPrev *Cut_Cut_t;
    var nLeaves int32;
    Cut_ListStart( p );
    for  pList != nil  {
        nLeaves = pList.nLeaves;
        p.pHead[nLeaves] = pList;
        for  pPrev = pList, pList = pList.pNext; (pList != nil); pPrev = pList, pList = pList.pNext  {
            if nLeaves < int32(pList.nLeaves) {
                break; }
        }
        p.ppTail[nLeaves] = &pPrev.pNext;
        pPrev.pNext = nil;
    }
}

/**Function*************************************************************

  Synopsis    [Adds the second list to the first list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ListAddList(pOld *Cut_List_t, pNew *Cut_List_t) {
    var i int32;
    for  i = 1; i <= CUT_SIZE_MAX; i++  {
        if pNew.pHead[i] == nil {
            continue; }
        *pOld.ppTail[i] = pNew.pHead[i];
        pOld.ppTail[i] = pNew.ppTail[i];
    }
}

/**Function*************************************************************

  Synopsis    [Returns the cut list linked into one sequence of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ListFinish(p *Cut_List_t) *Cut_Cut_t {
    var pHead *Cut_Cut_t = nil;
    var ppTail **Cut_Cut_t = &pHead;
    var i int32;
    for  i = 1; i <= CUT_SIZE_MAX; i++  {
        if p.pHead[i] == nil {
            continue; }
        *ppTail = p.pHead[i];
        ppTail = p.ppTail[i];
    }
    *ppTail = nil;
    return pHead;
}



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

