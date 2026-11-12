package abci

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [abcMerge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [LUT merging algorithm.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcMerge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/aig/aig.h"
//#include "aig/nwk/nwkMerge.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Marks the fanins of the node with the current trav ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMarkFanins_rec(pLut *Abc_Obj_t, nLevMin int32) {
    var pNext *Abc_Obj_t;
    var i int32;
    if !Abc_ObjIsNode(pLut) {
        return; }
    if Abc_NodeIsTravIdCurrent( pLut ) {
        return; }
    Abc_NodeSetTravIdCurrent( pLut );
    if Abc_ObjLevel(pLut) < nLevMin {
        return; }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pLut)) && (((pNext) = Abc_ObjFanin(pLut, i)), 1); i++  {
        Abc_NtkMarkFanins_rec( pNext, nLevMin );
    }
}

/**Function*************************************************************

  Synopsis    [Marks the fanouts of the node with the current trav ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMarkFanouts_rec(pLut *Abc_Obj_t, nLevMax int32, nFanMax int32) {
    var pNext *Abc_Obj_t;
    var i int32;
    if !Abc_ObjIsNode(pLut) {
        return; }
    if Abc_NodeIsTravIdCurrent( pLut ) {
        return; }
    Abc_NodeSetTravIdCurrent( pLut );
    if Abc_ObjLevel(pLut) > nLevMax {
        return; }
    if Abc_ObjFanoutNum(pLut) > nFanMax {
        return; }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pLut)) && (((pNext) = Abc_ObjFanout(pLut, i)), 1); i++  {
        Abc_NtkMarkFanouts_rec( pNext, nLevMax, nFanMax );
    }
}

/**Function*************************************************************

  Synopsis    [Collects the circle of nodes around the given set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectCircle(vStart *Vec_Ptr_t, vNext *Vec_Ptr_t, nFanMax int32) {
    var pObj, pNext *Abc_Obj_t;
    var i, k int32;
    Vec_PtrClear( vNext );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vStart)) && (((pObj) = (Vec_Int_t *)Vec_PtrEntry(vStart, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if !Abc_ObjIsNode(pNext) {
                continue; }
            if Abc_NodeIsTravIdCurrent( pNext ) {
                continue; }
            Abc_NodeSetTravIdCurrent( pNext );
            Vec_PtrPush( vNext, pNext );
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
            if !Abc_ObjIsNode(pNext) {
                continue; }
            if Abc_NodeIsTravIdCurrent( pNext ) {
                continue; }
            Abc_NodeSetTravIdCurrent( pNext );
            if Abc_ObjFanoutNum(pNext) > nFanMax {
                continue; }
            Vec_PtrPush( vNext, pNext );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Collects the circle of nodes removes from the given one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectNonOverlapCands(pLut *Abc_Obj_t, vStart *Vec_Ptr_t, vNext *Vec_Ptr_t, vCands *Vec_Ptr_t, pPars *Nwk_LMPars_t) {
    var vTemp *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    Vec_PtrClear( vCands );
    if pPars.nMaxSuppSize - Abc_ObjFaninNum(pLut) <= 1 {
        return; }

    // collect nodes removed by this distance
    assert(pPars.nMaxDistance > 0, "pPars.nMaxDistance > 0");
    Vec_PtrClear( vStart );
    Vec_PtrPush( vStart, pLut );
    Abc_NtkIncrementTravId( pLut.pNtk );
    Abc_NodeSetTravIdCurrent( pLut );
    for  i = 1; i <= pPars.nMaxDistance; i++  {
        Abc_NtkCollectCircle( vStart, vNext, pPars.nMaxFanout );
        vTemp  = vStart;
        vStart = vNext;
        vNext  = vTemp;
        // collect the nodes in vStart
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vStart)) && (((pObj) = (Vec_Int_t *)Vec_PtrEntry(vStart, k)), 1); k++  {
            Vec_PtrPush( vCands, pObj );
        }
    }

    // mark the TFI/TFO nodes
    Abc_NtkIncrementTravId( pLut.pNtk );
    if pPars.fUseTfiTfo {
        Abc_NodeSetTravIdCurrent( pLut );
    } else {
        Abc_NodeSetTravIdPrevious( pLut );
        Abc_NtkMarkFanins_rec( pLut, Abc_ObjLevel(pLut) - pPars.nMaxDistance );
        Abc_NodeSetTravIdPrevious( pLut );
        Abc_NtkMarkFanouts_rec( pLut, Abc_ObjLevel(pLut) + pPars.nMaxDistance, pPars.nMaxFanout );
    }

    // collect nodes satisfying the following conditions:
    // - they are close enough in terms of distance
    // - they are not in the TFI/TFO of the LUT
    // - they have no more than the given number of fanins
    // - they have no more than the given diff in delay
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCands)) && (((pObj) = (Vec_Int_t *)Vec_PtrEntry(vCands, i)), 1); i++  {
        if Abc_NodeIsTravIdCurrent(pObj) {
            continue; }
        if Abc_ObjFaninNum(pLut) + Abc_ObjFaninNum(pObj) > pPars.nMaxSuppSize {
            continue; }
        if Abc_ObjLevel(pLut) - Abc_ObjLevel(pObj) > pPars.nMaxLevelDiff || 
             Abc_ObjLevel(pObj) - Abc_ObjLevel(pLut) > pPars.nMaxLevelDiff {
             continue; }
        Vec_PtrWriteEntry( vCands, k++, pObj );
    }
    Vec_PtrShrink( vCands, k );
}


/**Function*************************************************************

  Synopsis    [Count the total number of fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCountTotalFanins(pLut *Abc_Obj_t, pCand *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    var nCounter int32 = Abc_ObjFaninNum(pLut);
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pCand)) && (((pFanin) = Abc_ObjFanin(pCand, i)), 1); i++  {
        nCounter += !pFanin.fMarkC;
    }
    return nCounter;
}

/**Function*************************************************************

  Synopsis    [Collects overlapping candidates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectOverlapCands(pLut *Abc_Obj_t, vCands *Vec_Ptr_t, pPars *Nwk_LMPars_t) {
    var pFanin, pObj *Abc_Obj_t;
    var i, k int32;
    // mark fanins of pLut
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pLut)) && (((pFanin) = Abc_ObjFanin(pLut, i)), 1); i++  {
        pFanin.fMarkC = 1;
    }
    // collect the matching fanouts of each fanin of the node
    Vec_PtrClear( vCands );
    Abc_NtkIncrementTravId( pLut.pNtk );
    Abc_NodeSetTravIdCurrent( pLut );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pLut)) && (((pFanin) = Abc_ObjFanin(pLut, i)), 1); i++  {
        if !Abc_ObjIsNode(pFanin) {
            continue; }
        if Abc_ObjFanoutNum(pFanin) > pPars.nMaxFanout {
            continue; }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pFanin)) && (((pObj) = Abc_ObjFanout(pFanin, k)), 1); k++  {
            if !Abc_ObjIsNode(pObj) {
                continue; }
            if Abc_NodeIsTravIdCurrent( pObj ) {
                continue; }
            Abc_NodeSetTravIdCurrent( pObj );
            // check the difference in delay
            if Abc_ObjLevel(pLut) - Abc_ObjLevel(pObj) > pPars.nMaxLevelDiff || 
                 Abc_ObjLevel(pObj) - Abc_ObjLevel(pLut) > pPars.nMaxLevelDiff {
                 continue; }
            // check the total number of fanins of the node
            if Abc_NtkCountTotalFanins(pLut, pObj) > pPars.nMaxSuppSize {
                continue; }
            Vec_PtrPush( vCands, pObj );
        }
    }
    // unmark fanins of pLut
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pLut)) && (((pFanin) = Abc_ObjFanin(pLut, i)), 1); i++  {
        pFanin.fMarkC = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Performs LUT merging with parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkLutMerge(pNtk *Abc_Ntk_t, pPars *Nwk_LMPars_t) *Vec_Int_t {
    var p *Nwk_Grf_t;
    var vResult *Vec_Int_t;
    var vStart, vNext, vCands1, vCands2 *Vec_Ptr_t;
    var pLut, pCand *Abc_Obj_t;
    var i, k, nVertsMax, nCands int32;
    var clk int64  = time.Now().UnixNano();
    // count the number of vertices
    nVertsMax = 0;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pLut) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pLut) == nil || !Abc_ObjIsNode(pLut) {} else {
        nVertsMax += int32((Abc_ObjFaninNum(pLut) <= pPars.nMaxLutSize)); }
    }
    p = Nwk_ManGraphAlloc( nVertsMax );
    // create graph
    vStart  = Vec_PtrAlloc( 1000 );
    vNext   = Vec_PtrAlloc( 1000 );
    vCands1 = Vec_PtrAlloc( 1000 );
    vCands2 = Vec_PtrAlloc( 1000 );
    nCands  = 0;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pLut) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pLut) == nil || !Abc_ObjIsNode(pLut) {} else {
        if Abc_ObjFaninNum(pLut) > pPars.nMaxLutSize {
            continue; }
        Abc_NtkCollectOverlapCands( pLut, vCands1, pPars );
        if pPars.fUseDiffSupp {
            Abc_NtkCollectNonOverlapCands( pLut, vStart, vNext, vCands2, pPars ); }
        if Vec_PtrSize(vCands1) == 0 && Vec_PtrSize(vCands2) == 0 {
            continue; }
        nCands += Vec_PtrSize(vCands1) + Vec_PtrSize(vCands2);
        // save candidates
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vCands1)) && (((pCand) = (Vec_Int_t *)Vec_PtrEntry(vCands1, k)), 1); k++  {
            Nwk_ManGraphHashEdge( p, Abc_ObjId(pLut), Abc_ObjId(pCand) );
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vCands2)) && (((pCand) = (Vec_Int_t *)Vec_PtrEntry(vCands2, k)), 1); k++  {
            Nwk_ManGraphHashEdge( p, Abc_ObjId(pLut), Abc_ObjId(pCand) );
        }
        // print statistics about this node
        if pPars.fVeryVerbose {
        fmt.Printf("Node %6d : Fanins = %d. Fanouts = %3d.  Cand1 = %3d. Cand2 = %3d.\n", Abc_ObjId(pLut), Abc_ObjFaninNum(pLut), Abc_ObjFaninNum(pLut), Vec_PtrSize(vCands1), Vec_PtrSize(vCands2)); }
    }
    }
    Vec_PtrFree( vStart );
    Vec_PtrFree( vNext );
    Vec_PtrFree( vCands1 );
    Vec_PtrFree( vCands2 );
    if pPars.fVerbose {
        fmt.Printf("Mergable LUTs = %6d. Total cands = %6d. ", p.nVertsMax, nCands);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Deriving graph"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    }
    // solve the graph problem
    clk = time.Now().UnixNano();
    Nwk_ManGraphSolve( p );
    if pPars.fVerbose {
        fmt.Printf("GRAPH: Nodes = %6d. Edges = %6d.  Pairs = %6d.  ", p.nVerts, p.nEdges, Vec_IntSize(p.vPairs)/2);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Solving"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
        Nwk_ManGraphReportMemoryUsage( p );
    }
    vResult = p.vPairs; p.vPairs = nil;
/*
    for ( i = 0; i < vResult->nSize; i += 2 )
        printf( "(%d,%d) ", vResult->pArray[i], vResult->pArray[i+1] );
    printf( "\n" );
*/
    Nwk_ManGraphFree( p );
    return vResult;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

