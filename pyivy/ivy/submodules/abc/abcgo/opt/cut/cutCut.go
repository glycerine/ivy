package cut

import "fmt"

/**CFile****************************************************************

  FileName    [cutNode.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [K-feasible cut computation package.]

  Synopsis    [Procedures to compute cuts for a node.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cutNode.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cutInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutAlloc(p *Cut_Man_t) *Cut_Cut_t {
    var pCut *Cut_Cut_t;
    // cut allocation
    pCut = (Cut_Cut_t *)Extra_MmFixedEntryFetch( p.pMmCuts );
    memset( pCut, 0, sizeof(Cut_Cut_t) );
    pCut.nVarsMax   = p.pParams.nVarsMax;
    pCut.fSimul     = p.fSimul;
    // statistics
    p.nCutsAlloc++;
    p.nCutsCur++;
    if p.nCutsPeak < p.nCutsAlloc - p.nCutsDealloc {
        p.nCutsPeak = p.nCutsAlloc - p.nCutsDealloc; }
    return pCut;
}

/**Function*************************************************************

  Synopsis    [Recybles the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutRecycle(p *Cut_Man_t, pCut *Cut_Cut_t) {
    p.nCutsDealloc++;
    p.nCutsCur--;
    if pCut.nLeaves == 1 {
        p.nCutsTriv--; }
    Extra_MmFixedEntryRecycle( p.pMmCuts, (char *)pCut );
}

/**Function*************************************************************

  Synopsis    [Compares two cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutCompare(pCut1 *Cut_Cut_t, pCut2 *Cut_Cut_t) int32 {
    var i int32;
    if pCut1.nLeaves < pCut2.nLeaves {
        return -1; }
    if pCut1.nLeaves > pCut2.nLeaves {
        return 1; }
    for  i = 0; i < int32(pCut1.nLeaves); i++  {
        if pCut1.pLeaves[i] < pCut2.pLeaves[i] {
            return -1; }
        if pCut1.pLeaves[i] > pCut2.pLeaves[i] {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Duplicates the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutDupList(p *Cut_Man_t, pList *Cut_Cut_t) *Cut_Cut_t {
    var pHead *Cut_Cut_t = nil;
    var ppTail **Cut_Cut_t = &pHead;
    var pTemp, pCopy *Cut_Cut_t;
    if pList == nil {
        return nil; }
    // inlining of macro Cut_ListForEachCut
    for  pTemp = pList;                                             pTemp;                                                     pTemp = pTemp.pNext  {
        pCopy = (Cut_Cut_t *)Extra_MmFixedEntryFetch( p.pMmCuts );
        memcpy( pCopy, pTemp, (size_t)p.EntrySize );
        *ppTail = pCopy;
        ppTail = &pCopy.pNext;
    }
    *ppTail = nil;
    return pHead;
}

/**Function*************************************************************

  Synopsis    [Recycles the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutRecycleList(p *Cut_Man_t, pList *Cut_Cut_t) {
    var pCut, pCut2 *Cut_Cut_t;
    // inlining of macro Cut_ListForEachCutSafe
    for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
        Extra_MmFixedEntryRecycle( p.pMmCuts, (char *)pCut );
    }
}

/**Function*************************************************************

  Synopsis    [Counts the number of cuts in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutCountList(pList *Cut_Cut_t) int32 {
    var Counter int32;
    // inlining of macro Cut_ListForEachCut
    for  pList = pList;                                             (pList != nil);                                                     pList = pList.pNext  {
        Counter++;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Merges two NULL-terminated linked lists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutMergeLists(pList1 *Cut_Cut_t, pList2 *Cut_Cut_t) *Cut_Cut_t {
    var pList *Cut_Cut_t = nil;
    var ppTail **Cut_Cut_t = &pList;
    var pCut *Cut_Cut_t;
    for  (pList1 != nil) && (pList2 != nil)  {
        if Cut_CutCompare(pList1, pList2) < 0 {
            pCut = pList1;
            pList1 = pList1.pNext;
        
        } else {
            pCut = pList2;
            pList2 = pList2.pNext;
        }
        *ppTail = pCut;
        ppTail = &pCut.pNext;
    }
    *ppTail = pList1? pList1: pList2;
    return pList;
}

/**Function*************************************************************

  Synopsis    [Sets the number of the cuts in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutNumberList(pList *Cut_Cut_t) {
    var pCut *Cut_Cut_t;
    var i int32;
    // inlining of macro Cut_ListForEachCut
    for  pCut = pList;                                             (pCut != nil);                                                     pCut = pCut.pNext  {
        pCut.Num0 = i++;
    }
}

/**Function*************************************************************

  Synopsis    [Creates the trivial cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutCreateTriv(p *Cut_Man_t, Node int32) *Cut_Cut_t {
    var pCut *Cut_Cut_t;
    if p.pParams.fSeq {
        Node <<= CUT_SHIFT; }
    pCut = Cut_CutAlloc( p );
    pCut.nLeaves    = 1;
    pCut.pLeaves[0] = Node;
    pCut.uSign      = Cut_NodeSign( Node );
    if p.pParams.fTruth {
/*
        if ( pCut->nVarsMax == 4 )
            Cut_CutWriteTruth( pCut, p->uTruthVars[0] );
        else
            Extra_BitCopy( pCut->nLeaves, p->uTruths[0], (uint8*)Cut_CutReadTruth(pCut) );
*/
        var pTruth *uint32  = Cut_CutReadTruth(pCut);
        var i int32;
        for  i = 0; i < p.nTruthWords; i++  {
            pTruth[i] = 0xAAAAAAAA;
        }
    }
    p.nCutsTriv++;
    return pCut;
} 


/**Function*************************************************************

  Synopsis    [Print the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutPrint(pCut *Cut_Cut_t, fSeq int32) {
    var i int32;
    assert(pCut.nLeaves > 0, "pCut.nLeaves > 0");
    fmt.Printf("%d : {", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        if fSeq != 0 {
            fmt.Printf(" %d", pCut.pLeaves[i] >> CUT_SHIFT);
            if pCut.pLeaves[i] & CUT_MASK {
                fmt.Printf("(%d)", pCut.pLeaves[i] & CUT_MASK); }
        
        } else {
            fmt.Printf(" %d", pCut.pLeaves[i]); }
    }
    fmt.Printf(" }");
//    printf( "\nSign = " );
//    Extra_PrintBinary( stdout, &pCut->uSign, 32 );
}

/**Function*************************************************************

  Synopsis    [Print the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutPrintList(pList *Cut_Cut_t, fSeq int32) {
    var pCut *Cut_Cut_t;
    for  pCut = pList; (pCut != nil); pCut = pCut.pNext  {
        Cut_CutPrint( pCut, fSeq );
fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Consider dropping cuts if they are useless by now.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutPrintMerge(pCut *Cut_Cut_t, pCut0 *Cut_Cut_t, pCut1 *Cut_Cut_t) {
    fmt.Printf("\n");
    fmt.Printf("%d : %5d %5d %5d %5d %5d\n", pCut0.nLeaves, pCut0.nLeaves > 0 ? pCut0.pLeaves[0] : -1, pCut0.nLeaves > 1 ? pCut0.pLeaves[1] : -1, pCut0.nLeaves > 2 ? pCut0.pLeaves[2] : -1, pCut0.nLeaves > 3 ? pCut0.pLeaves[3] : -1, pCut0.nLeaves > 4 ? pCut0.pLeaves[4] : -1);
    fmt.Printf("%d : %5d %5d %5d %5d %5d\n", pCut1.nLeaves, pCut1.nLeaves > 0 ? pCut1.pLeaves[0] : -1, pCut1.nLeaves > 1 ? pCut1.pLeaves[1] : -1, pCut1.nLeaves > 2 ? pCut1.pLeaves[2] : -1, pCut1.nLeaves > 3 ? pCut1.pLeaves[3] : -1, pCut1.nLeaves > 4 ? pCut1.pLeaves[4] : -1);
    if pCut == nil {
        fmt.Printf("Cannot merge\n");
    } else {
        fmt.Printf("%d : %5d %5d %5d %5d %5d\n", pCut.nLeaves, pCut.nLeaves > 0 ? pCut.pLeaves[0] : -1, pCut.nLeaves > 1 ? pCut.pLeaves[1] : -1, pCut.nLeaves > 2 ? pCut.pLeaves[2] : -1, pCut.nLeaves > 3 ? pCut.pLeaves[3] : -1, pCut.nLeaves > 4 ? pCut.pLeaves[4] : -1); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

