package nwk

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [nwkMerge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Netlist representation.]

  Synopsis    [LUT merging algorithm.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkMerge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nwk.h"
//#include "nwkMerge.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphAlloc(nVertsMax int32) *Nwk_Grf_t {
    var p *Nwk_Grf_t;
    // inlining of macro ABC_ALLOC
    p = ((Nwk_Grf_t *) malloc(sizeof(Nwk_Grf_t) * (size_t)(1)));
    memset( p, 0, sizeof(Nwk_Grf_t) );
    p.nVertsMax = nVertsMax;
    p.nEdgeHash = Abc_PrimeCudd( 3 * nVertsMax );
    // inlining of macro ABC_CALLOC
    p.pEdgeHash = ((Nwk_Edg_t * *) calloc((size_t)(p.nEdgeHash), sizeof(Nwk_Edg_t *)));
    p.pMemEdges = Aig_MmFixedStart( sizeof(Nwk_Edg_t), p.nEdgeHash );
    p.vPairs    = Vec_IntAlloc( 1000 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocates the graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphFree(p *Nwk_Grf_t) {
    if p.vPairs {    Vec_IntFree( p.vPairs ); }
    if p.pMemEdges { Aig_MmFixedStop( p.pMemEdges, 0 ); }
    if p.pMemVerts { Aig_MmFlexStop( p.pMemVerts, 0 ); }
    // inlining of macro ABC_FREE
    ((p.pVerts) ? (free((char *) (p.pVerts)), (p.pVerts) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pEdgeHash) ? (free((char *) (p.pEdgeHash)), (p.pEdgeHash) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMapLut2Id) ? (free((char *) (p.pMapLut2Id)), (p.pMapLut2Id) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMapId2Lut) ? (free((char *) (p.pMapId2Lut)), (p.pMapId2Lut) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prepares the graph for solving the problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphReportMemoryUsage(p *Nwk_Grf_t) {
    p.nMemBytes1 = 
        sizeof(Nwk_Grf_t) +
        sizeof(void *) * p.nEdgeHash + 
        sizeof(int) * (p.nObjs + p.nVertsMax) +
        sizeof(Nwk_Edg_t) * p.nEdges;
    p.nMemBytes2 = 
        sizeof(Nwk_Vrt_t) * p.nVerts + 
        sizeof(int) * 2 * p.nEdges;
    fmt.Printf("Memory usage stats:  Preprocessing = %.2f MB.  Solving = %.2f MB.\n", 1.0 * p.nMemBytes1 / (1<<20), 1.0 * p.nMemBytes2 / (1<<20));
}


/**Function*************************************************************

  Synopsis    [Finds or adds the edge to the graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphHashEdge(p *Nwk_Grf_t, iLut1 int32, iLut2 int32) {
    var pEntry *Nwk_Edg_t;
    var Key uint32;
    if iLut1 == iLut2 {
        return; }
    if iLut1 > iLut2 {
        Key = iLut1;
        iLut1 = iLut2;
        iLut2 = Key;
    }
    assert(iLut1 < iLut2, "iLut1 < iLut2");
    if p.nObjs < iLut2 {
        p.nObjs = iLut2; }
    Key = uint32((741457 * iLut1 + 4256249 * iLut2)) % p.nEdgeHash;
    for  pEntry = p.pEdgeHash[Key]; (pEntry != nil); pEntry = pEntry.pNext  {
        if pEntry.iNode1 == iLut1 && pEntry.iNode2 == iLut2 {
            return; }
    }
    pEntry = (Nwk_Edg_t *)Aig_MmFixedEntryFetch( p.pMemEdges );
    pEntry.iNode1 = iLut1;
    pEntry.iNode2 = iLut2;
    pEntry.pNext = p.pEdgeHash[Key];
    p.pEdgeHash[Key] = pEntry;
    p.nEdges++;
}

/**Function*************************************************************

  Synopsis    [Adds one entry to the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphListAdd(p *Nwk_Grf_t, pList *int32, pVertex *Nwk_Vrt_t) {
    if *pList {
        var pHead *Nwk_Vrt_t;
        pHead = p.pVerts[*pList];
        pVertex.iPrev = 0;
        pVertex.iNext = pHead.Id;
        pHead.iPrev = pVertex.Id;
    }
    *pList = pVertex.Id;
}

/**Function*************************************************************

  Synopsis    [Deletes one entry from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphListDelete(p *Nwk_Grf_t, pList *int32, pVertex *Nwk_Vrt_t) {
    assert(*pList, "*pList");
    if pVertex.iPrev {
//        assert( p->pVerts[pVertex->iPrev]->iNext == pVertex->Id );
        p.pVerts[pVertex.iPrev].iNext = pVertex.iNext;
    }
    if pVertex.iNext {
//        assert( p->pVerts[pVertex->iNext]->iPrev == pVertex->Id );
        p.pVerts[pVertex.iNext].iPrev = pVertex.iPrev;
    }
    if *pList == pVertex.Id {
        *pList = pVertex.iNext; }
    pVertex.iPrev = pVertex.iNext = 0;
}

/**Function*************************************************************

  Synopsis    [Inserts the edge into one of the linked lists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphListInsert(p *Nwk_Grf_t, pVertex *Nwk_Vrt_t) {
    var pNext *Nwk_Vrt_t;
    assert(pVertex.nEdges > 0, "pVertex.nEdges > 0");

    if pVertex.nEdges == 1 {
        pNext = p.pVerts[ pVertex.pEdges[0] ];
        if pNext.nEdges >= NWK_MAX_LIST {
            Nwk_ManGraphListAdd( p, p.pLists1 + NWK_MAX_LIST, pVertex );
        } else {
            Nwk_ManGraphListAdd( p, p.pLists1 + pNext.nEdges, pVertex ); }
    
    } else {
        if pVertex.nEdges >= NWK_MAX_LIST {
            Nwk_ManGraphListAdd( p, p.pLists2 + NWK_MAX_LIST, pVertex );
        } else {
            Nwk_ManGraphListAdd( p, p.pLists2 + pVertex.nEdges, pVertex ); }
    }
}

/**Function*************************************************************

  Synopsis    [Extracts the edge from one of the linked lists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphListExtract(p *Nwk_Grf_t, pVertex *Nwk_Vrt_t) {
    var pNext *Nwk_Vrt_t;
    assert(pVertex.nEdges > 0, "pVertex.nEdges > 0");

    if pVertex.nEdges == 1 {
        pNext = p.pVerts[ pVertex.pEdges[0] ];
        if pNext.nEdges >= NWK_MAX_LIST {
            Nwk_ManGraphListDelete( p, p.pLists1 + NWK_MAX_LIST, pVertex );
        } else {
            Nwk_ManGraphListDelete( p, p.pLists1 + pNext.nEdges, pVertex ); }
    
    } else {
        if pVertex.nEdges >= NWK_MAX_LIST {
            Nwk_ManGraphListDelete( p, p.pLists2 + NWK_MAX_LIST, pVertex );
        } else {
            Nwk_ManGraphListDelete( p, p.pLists2 + pVertex.nEdges, pVertex ); }
    }
}

/**Function*************************************************************

  Synopsis    [Prepares the graph for solving the problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphPrepare(p *Nwk_Grf_t) {
    var pEntry *Nwk_Edg_t;
    var pVertex *Nwk_Vrt_t;
    var pnEdges *int32;
    var nBytes int32;
    var i int32;
    // allocate memory for the present objects
    // inlining of macro ABC_ALLOC
    p.pMapLut2Id = ((int *) malloc(sizeof(int) * (size_t)(p.nObjs+1)));
    // inlining of macro ABC_ALLOC
    p.pMapId2Lut = ((int *) malloc(sizeof(int) * (size_t)(p.nVertsMax+1)));
    memset( p.pMapLut2Id, 0xff, sizeof(int) * (p.nObjs+1) );
    memset( p.pMapId2Lut, 0xff, sizeof(int) * (p.nVertsMax+1) );
    // mark present objects
    // inlining of macro Nwk_GraphForEachEdge
    for  i = 0; i < p.nEdgeHash; i++  {                        for  pEntry = p.pEdgeHash[i]; (pEntry != nil); pEntry = pEntry.pNext  {
        assert(pEntry.iNode1 <= p.nObjs, "pEntry.iNode1 <= p.nObjs");
        assert(pEntry.iNode2 <= p.nObjs, "pEntry.iNode2 <= p.nObjs");
        p.pMapLut2Id[ pEntry.iNode1 ] = 0;
        p.pMapLut2Id[ pEntry.iNode2 ] = 0;
    }
    }
    // map objects
    p.nVerts = 0;
    for  i = 0; i <= p.nObjs; i++  {
        if p.pMapLut2Id[i] == 0 {
            p.pMapLut2Id[i] = ++p.nVerts;
            p.pMapId2Lut[p.nVerts] = i;
        }
    }
    // count the edges and mark present objects
    // inlining of macro ABC_CALLOC
    pnEdges = ((int *) calloc((size_t)(p.nVerts+1), sizeof(int)));
    // inlining of macro Nwk_GraphForEachEdge
    for  i = 0; i < p.nEdgeHash; i++  {                        for  pEntry = p.pEdgeHash[i]; (pEntry != nil); pEntry = pEntry.pNext  {
        // translate into vertices
        assert(pEntry.iNode1 <= p.nObjs, "pEntry.iNode1 <= p.nObjs");
        assert(pEntry.iNode2 <= p.nObjs, "pEntry.iNode2 <= p.nObjs");
        pEntry.iNode1 = p.pMapLut2Id[pEntry.iNode1];
        pEntry.iNode2 = p.pMapLut2Id[pEntry.iNode2];
        // count the edges
        assert(pEntry.iNode1 <= p.nVerts, "pEntry.iNode1 <= p.nVerts");
        assert(pEntry.iNode2 <= p.nVerts, "pEntry.iNode2 <= p.nVerts");
        pnEdges[pEntry.iNode1]++;
        pnEdges[pEntry.iNode2]++;
    }
    }
    // allocate the real graph
    p.pMemVerts  = Aig_MmFlexStart();
    // inlining of macro ABC_ALLOC
    p.pVerts = ((Nwk_Vrt_t * *) malloc(sizeof(Nwk_Vrt_t *) * (size_t)(p.nVerts + 1)));
    p.pVerts[0] = nil;
    for  i = 1; i <= p.nVerts; i++  {
        assert(pnEdges[i] > 0, "pnEdges[i] > 0");
        nBytes = sizeof(Nwk_Vrt_t) + sizeof(int) * pnEdges[i];
        p.pVerts[i] = (Nwk_Vrt_t *)Aig_MmFlexEntryFetch( p.pMemVerts, nBytes );
        memset( p.pVerts[i], 0, nBytes );
        p.pVerts[i].Id = i;
    }
    // add edges to the real graph
    // inlining of macro Nwk_GraphForEachEdge
    for  i = 0; i < p.nEdgeHash; i++  {                        for  pEntry = p.pEdgeHash[i]; (pEntry != nil); pEntry = pEntry.pNext  {
        pVertex = p.pVerts[pEntry.iNode1];
        pVertex.pEdges[ pVertex.nEdges++ ] = pEntry.iNode2;
        pVertex = p.pVerts[pEntry.iNode2];
        pVertex.pEdges[ pVertex.nEdges++ ] = pEntry.iNode1;
    }
    }
    // put vertices into the data structure
    for  i = 1; i <= p.nVerts; i++  {
        assert(p.pVerts[i].nEdges == pnEdges[i], "p.pVerts[i].nEdges == pnEdges[i]");
        Nwk_ManGraphListInsert( p, p.pVerts[i] );
    }
    // clean up
    Aig_MmFixedStop( p.pMemEdges, 0 ); p.pMemEdges = nil;
    // inlining of macro ABC_FREE
    ((p.pEdgeHash) ? (free((char *) (p.pEdgeHash)), (p.pEdgeHash) = 0) : 0);
//    p->nEdgeHash = 0;
    // inlining of macro ABC_FREE
    ((pnEdges) ? (free((char *) (pnEdges)), (pnEdges) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Sort pairs by the first vertex in the topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphSortPairs(p *Nwk_Grf_t) {
    var nSize int32  = Vec_IntSize(p.vPairs);
    var pIdToPair *int32;
    var i int32;
    // allocate storage
    // inlining of macro ABC_ALLOC
    pIdToPair = ((int *) malloc(sizeof(int) * (size_t)(p.nObjs+1)));
    for  i = 0; i <= p.nObjs; i++  {
        pIdToPair[i] = -1;
    }
    // create mapping
    for  i = 0; i < p.vPairs.nSize; i += 2  {
        assert(pIdToPair[ p.vPairs.pArray[i] ] == -1, "pIdToPair[ p.vPairs.pArray[i] ] == -1");
        pIdToPair[ p.vPairs.pArray[i] ] = p.vPairs.pArray[i+1];
    }
    // recreate pairs
    Vec_IntClear( p.vPairs );
    for  i = 0; i <= p.nObjs; i++  {
        if pIdToPair[i] >= 0 {
            assert(i < pIdToPair[i], "i < pIdToPair[i]");
            Vec_IntPush( p.vPairs, i );
            Vec_IntPush( p.vPairs, pIdToPair[i] );
        }
    }
    assert(nSize == Vec_IntSize(p.vPairs), "nSize == Vec_IntSize(p.vPairs)");
    // inlining of macro ABC_FREE
    ((pIdToPair) ? (free((char *) (pIdToPair)), (pIdToPair) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Updates the problem after pulling out one edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphCheckLists(p *Nwk_Grf_t) {
    var pVertex, pNext *Nwk_Vrt_t;
    var i, j int32;
    assert(p.pLists1[0] == 0, "p.pLists1[0] == 0");
    for  i = 1; i <= NWK_MAX_LIST; i++  {
        if p.pLists1[i] {
            pVertex = p.pVerts[ p.pLists1[i] ];
            assert(pVertex.nEdges == 1, "pVertex.nEdges == 1");
            pNext = p.pVerts[ pVertex.pEdges[0] ];
            assert(pNext.nEdges == i || pNext.nEdges > NWK_MAX_LIST, "pNext.nEdges == i || pNext.nEdges > NWK_MAX_LIST");
        }
    }
    // find the next vertext to extract
    assert(p.pLists2[0] == 0, "p.pLists2[0] == 0");
    assert(p.pLists2[1] == 0, "p.pLists2[1] == 0");
    for  j = 2; j <= NWK_MAX_LIST; j++  {
        if p.pLists2[j] {
            pVertex = p.pVerts[ p.pLists2[j] ];
            assert(pVertex.nEdges == j || pVertex.nEdges > NWK_MAX_LIST, "pVertex.nEdges == j || pVertex.nEdges > NWK_MAX_LIST");
        }
    }
}

/**Function*************************************************************

  Synopsis    [Extracts the edge from one of the linked lists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphVertexRemoveEdge(pThis *Nwk_Vrt_t, pNext *Nwk_Vrt_t) {
    var k int32;
    for  k = 0; k < pThis.nEdges; k++  {
        if pThis.pEdges[k] == pNext.Id {
            break; }
    }
    assert(k < pThis.nEdges, "k < pThis.nEdges");
    pThis.nEdges--;
    for  ; k < pThis.nEdges; k++  {
        pThis.pEdges[k] = pThis.pEdges[k+1];
    }
}

/**Function*************************************************************

  Synopsis    [Updates the problem after pulling out one edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphUpdate(p *Nwk_Grf_t, pVertex *Nwk_Vrt_t, pNext *Nwk_Vrt_t) {
    var pChanged, pOther *Nwk_Vrt_t;
    var i, k int32;
//    Nwk_ManGraphCheckLists( p );
    Nwk_ManGraphListExtract( p, pVertex );
    Nwk_ManGraphListExtract( p, pNext );
    // update neihbors of pVertex
    // inlining of macro Nwk_VertexForEachAdjacent
    for  i = 0; (i < pVertex.nEdges) && (((pChanged) = p.pVerts[pVertex.pEdges[i]]), 1); i++  {
        if pChanged == pNext {
            continue; }
        Nwk_ManGraphListExtract( p, pChanged );
        // move those that use this one
        if pChanged.nEdges > 1 {
        // inlining of macro Nwk_VertexForEachAdjacent
        for  k = 0; (k < pChanged.nEdges) && (((pOther) = p.pVerts[pChanged.pEdges[k]]), 1); k++  {
            if pOther == pVertex || pOther.nEdges > 1 {
                continue; }
            assert(pOther.nEdges == 1, "pOther.nEdges == 1");
            Nwk_ManGraphListExtract( p, pOther );
            pChanged.nEdges--;
            Nwk_ManGraphListInsert( p, pOther );
            pChanged.nEdges++;
        } }
        // remove the edge
        Nwk_ManGraphVertexRemoveEdge( pChanged, pVertex );
        // add the changed vertex back
        if pChanged.nEdges > 0 {
            Nwk_ManGraphListInsert( p, pChanged ); }
    }
    // update neihbors of pNext
    // inlining of macro Nwk_VertexForEachAdjacent
    for  i = 0; (i < pNext.nEdges) && (((pChanged) = p.pVerts[pNext.pEdges[i]]), 1); i++  {
        if pChanged == pVertex {
            continue; }
        Nwk_ManGraphListExtract( p, pChanged );
        // move those that use this one
        if pChanged.nEdges > 1 {
        // inlining of macro Nwk_VertexForEachAdjacent
        for  k = 0; (k < pChanged.nEdges) && (((pOther) = p.pVerts[pChanged.pEdges[k]]), 1); k++  {
            if pOther == pNext || pOther.nEdges > 1 {
                continue; }
            assert(pOther.nEdges == 1, "pOther.nEdges == 1");
            Nwk_ManGraphListExtract( p, pOther );
            pChanged.nEdges--;
            Nwk_ManGraphListInsert( p, pOther );
            pChanged.nEdges++;
        } }
        // remove the edge
        Nwk_ManGraphVertexRemoveEdge( pChanged, pNext );
        // add the changed vertex back
        if pChanged.nEdges > 0 {
            Nwk_ManGraphListInsert( p, pChanged ); }
    }
    // add to the result
    if pVertex.Id < pNext.Id {
        Vec_IntPush( p.vPairs, p.pMapId2Lut[pVertex.Id] ); 
        Vec_IntPush( p.vPairs, p.pMapId2Lut[pNext.Id] ); 
    
    } else {
        Vec_IntPush( p.vPairs, p.pMapId2Lut[pNext.Id] ); 
        Vec_IntPush( p.vPairs, p.pMapId2Lut[pVertex.Id] ); 
    }
//    Nwk_ManGraphCheckLists( p );
}

/**Function*************************************************************

  Synopsis    [Counts the number of entries in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphListLength(p *Nwk_Grf_t, List int32) int32 {
    var pThis *Nwk_Vrt_t;
    var fVerbose int32;
    var Counter int32;
    // inlining of macro Nwk_ListForEachVertex
    for  pThis = List? p.pVerts[List] : nil;  (pThis != nil);             pThis = pThis.iNext? p.pVerts[pThis.iNext] : nil  {
        if (fVerbose != 0) && Counter < 20 {
            fmt.Printf("%d ", p.pVerts[pThis.pEdges[0]].nEdges); }
        Counter++;
    }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns the adjacent vertex with the mininum number of edges.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphListFindMinEdge(p *Nwk_Grf_t, pVert *Nwk_Vrt_t) *Nwk_Vrt_t {
    var pThis *Nwk_Vrt_t;
    var pMinCost *Nwk_Vrt_t = nil;
    var k int32;
    // inlining of macro Nwk_VertexForEachAdjacent
    for  k = 0; (k < pVert.nEdges) && (((pThis) = p.pVerts[pVert.pEdges[k]]), 1); k++  {
        if pMinCost == nil || pMinCost.nEdges > pThis.nEdges {
            pMinCost = pThis; }
    }
    return pMinCost;
}

/**Function*************************************************************

  Synopsis    [Finds the best vertext in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphListFindMin(p *Nwk_Grf_t, List int32) *Nwk_Vrt_t {
    var pThis *Nwk_Vrt_t;
    var pMinCost *Nwk_Vrt_t = nil;
    var k int32;
    var Counter int32 = 10000;
    var BestCost int32 = 1000000;
    // inlining of macro Nwk_ListForEachVertex
    for  pThis = List? p.pVerts[List] : nil;  (pThis != nil);             pThis = pThis.iNext? p.pVerts[pThis.iNext] : nil  {
        for  k = 0; k < pThis.nEdges; k++  {
            if pMinCost == nil || BestCost > p.pVerts[pThis.pEdges[k]].nEdges {
                BestCost = p.pVerts[pThis.pEdges[k]].nEdges;
                pMinCost = pThis;
            }
        }
        if Counter--; Counter == 0 {
            break; }
    }
    return pMinCost;
}

/**Function*************************************************************

  Synopsis    [Solves the problem by extracting one edge at a time.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGraphSolve(p *Nwk_Grf_t) {
    var pVertex, pNext *Nwk_Vrt_t;
    var i, j int32;
    Nwk_ManGraphPrepare( p );
    for {
        // find the next vertex to extract
        assert(p.pLists1[0] == 0, "p.pLists1[0] == 0");
        for  i = 1; i <= NWK_MAX_LIST; i++  {
            if p.pLists1[i] {
//                printf( "%d ", i );
//                printf( "ListA = %2d. Length = %5d.\n", i, Nwk_ManGraphListLength(p,p->pLists1[i]) );
                pVertex = p.pVerts[ p.pLists1[i] ];
                assert(pVertex.nEdges == 1, "pVertex.nEdges == 1");
                pNext = p.pVerts[ pVertex.pEdges[0] ];
                Nwk_ManGraphUpdate( p, pVertex, pNext );
                break;
            }
        }
        if i < NWK_MAX_LIST + 1 {
            continue; }
        // find the next vertex to extract
        assert(p.pLists2[0] == 0, "p.pLists2[0] == 0");
        assert(p.pLists2[1] == 0, "p.pLists2[1] == 0");
        for  j = 2; j <= NWK_MAX_LIST; j++  {
            if p.pLists2[j] {
//                printf( "***%d ", j );
//                printf( "ListB = %2d. Length = %5d.\n", j, Nwk_ManGraphListLength(p,p->pLists2[j]) );
                pVertex = Nwk_ManGraphListFindMin( p, p.pLists2[j] ); 
                assert(pVertex.nEdges == j || j == NWK_MAX_LIST, "pVertex.nEdges == j || j == NWK_MAX_LIST");
                pNext = Nwk_ManGraphListFindMinEdge( p, pVertex );
                Nwk_ManGraphUpdate( p, pVertex, pNext );
                break;
            }
        }
        if j == NWK_MAX_LIST + 1 {
            break; }
    }
    Nwk_ManGraphSortPairs( p );
}

/**Function*************************************************************

  Synopsis    [Reads graph from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManLutMergeReadGraph(pFileName string) *Nwk_Grf_t {
    var p *Nwk_Grf_t;
    var pFile *os.File;
    var Buffer [100]byte;
    var nNodes, nEdges, iNode1, iNode2 int32;
    var RetValue int32;
    pFile = os.Open(pFileName);
    RetValue = fscanf( pFile, "%s %d", Buffer, &nNodes );
    RetValue = fscanf( pFile, "%s %d", Buffer, &nEdges );
    p = Nwk_ManGraphAlloc( nNodes );
    for  fscanf( pFile, "%s %d %d", Buffer, &iNode1, &iNode2 ) == 3  {
        Nwk_ManGraphHashEdge( p, iNode1, iNode2 );
    }
    assert(p.nEdges == nEdges, "p.nEdges == nEdges");
    _ = (pFile).Close();
    return p;
}

/**Function*************************************************************

  Synopsis    [Solves the graph coming from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManLutMergeGraphTest(pFileName string) int32 {
    var nPairs int32;
    var p *Nwk_Grf_t;
    var clk abctime  = Abc_Clock();
    p = Nwk_ManLutMergeReadGraph( pFileName );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Reading"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    clk = Abc_Clock();
    Nwk_ManGraphSolve( p );
    fmt.Printf("GRAPH: Nodes = %6d. Edges = %6d.  Pairs = %6d.  ", p.nVerts, p.nEdges, Vec_IntSize(p.vPairs)/2);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Solving"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    nPairs = Vec_IntSize(p.vPairs)/2;
    Nwk_ManGraphReportMemoryUsage( p );
    Nwk_ManGraphFree( p );
    return nPairs;
}




/**Function*************************************************************

  Synopsis    [Marks the fanins of the node with the current trav ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManMarkFanins_rec(pLut *Nwk_Obj_t, nLevMin int32) {
    var pNext *Nwk_Obj_t;
    var i int32;
    if !Nwk_ObjIsNode(pLut) {
        return; }
    if Nwk_ObjIsTravIdCurrent( pLut ) {
        return; }
    Nwk_ObjSetTravIdCurrent( pLut );
    if Nwk_ObjLevel(pLut) < nLevMin {
        return; }
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pLut).nFanins)) && ((pNext) = (pLut).pFanio[i]); i++  {
        Nwk_ManMarkFanins_rec( pNext, nLevMin );
    }
}

/**Function*************************************************************

  Synopsis    [Marks the fanouts of the node with the current trav ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManMarkFanouts_rec(pLut *Nwk_Obj_t, nLevMax int32, nFanMax int32) {
    var pNext *Nwk_Obj_t;
    var i int32;
    if !Nwk_ObjIsNode(pLut) {
        return; }
    if Nwk_ObjIsTravIdCurrent( pLut ) {
        return; }
    Nwk_ObjSetTravIdCurrent( pLut );
    if Nwk_ObjLevel(pLut) > nLevMax {
        return; }
    if Nwk_ObjFanoutNum(pLut) > nFanMax {
        return; }
    // inlining of macro Nwk_ObjForEachFanout
    for  i = 0; (i < int32((pLut).nFanouts)) && ((pNext) = (pLut).pFanio[(pLut).nFanins+i]); i++  {
        Nwk_ManMarkFanouts_rec( pNext, nLevMax, nFanMax );
    }
}

/**Function*************************************************************

  Synopsis    [Collects the circle of nodes around the given set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCollectCircle(vStart *Vec_Ptr_t, vNext *Vec_Ptr_t, nFanMax int32) {
    var pObj, pNext *Nwk_Obj_t;
    var i, k int32;
    Vec_PtrClear( vNext );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vStart)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vStart, i)), 1); i++  {
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[k]); k++  {
            if !Nwk_ObjIsNode(pNext) {
                continue; }
            if Nwk_ObjIsTravIdCurrent( pNext ) {
                continue; }
            Nwk_ObjSetTravIdCurrent( pNext );
            Vec_PtrPush( vNext, pNext );
        }
        // inlining of macro Nwk_ObjForEachFanout
        for  k = 0; (k < int32((pObj).nFanouts)) && ((pNext) = (pObj).pFanio[(pObj).nFanins+k]); k++  {
            if !Nwk_ObjIsNode(pNext) {
                continue; }
            if Nwk_ObjIsTravIdCurrent( pNext ) {
                continue; }
            Nwk_ObjSetTravIdCurrent( pNext );
            if Nwk_ObjFanoutNum(pNext) > nFanMax {
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
func Nwk_ManCollectNonOverlapCands(pLut *Nwk_Obj_t, vStart *Vec_Ptr_t, vNext *Vec_Ptr_t, vCands *Vec_Ptr_t, pPars *Nwk_LMPars_t) {
    var vTemp *Vec_Ptr_t;
    var pObj *Nwk_Obj_t;
    var i, k int32;
    Vec_PtrClear( vCands );
    if pPars.nMaxSuppSize - Nwk_ObjFaninNum(pLut) <= 1 {
        return; }

    // collect nodes removed by this distance
    assert(pPars.nMaxDistance > 0, "pPars.nMaxDistance > 0");
    Vec_PtrClear( vStart );
    Vec_PtrPush( vStart, pLut );
    Nwk_ManIncrementTravId( pLut.pMan );
    Nwk_ObjSetTravIdCurrent( pLut );
    for  i = 1; i <= pPars.nMaxDistance; i++  {
        Nwk_ManCollectCircle( vStart, vNext, pPars.nMaxFanout );
        vTemp  = vStart;
        vStart = vNext;
        vNext  = vTemp;
        // collect the nodes in vStart
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vStart)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vStart, k)), 1); k++  {
            Vec_PtrPush( vCands, pObj );
        }
    }

    // mark the TFI/TFO nodes
    Nwk_ManIncrementTravId( pLut.pMan );
    if pPars.fUseTfiTfo {
        Nwk_ObjSetTravIdCurrent( pLut );
    } else {
        Nwk_ObjSetTravIdPrevious( pLut );
        Nwk_ManMarkFanins_rec( pLut, Nwk_ObjLevel(pLut) - pPars.nMaxDistance );
        Nwk_ObjSetTravIdPrevious( pLut );
        Nwk_ManMarkFanouts_rec( pLut, Nwk_ObjLevel(pLut) + pPars.nMaxDistance, pPars.nMaxFanout );
    }

    // collect nodes satisfying the following conditions:
    // - they are close enough in terms of distance
    // - they are not in the TFI/TFO of the LUT
    // - they have no more than the given number of fanins
    // - they have no more than the given diff in delay
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCands)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vCands, i)), 1); i++  {
        if Nwk_ObjIsTravIdCurrent(pObj) {
            continue; }
        if Nwk_ObjFaninNum(pLut) + Nwk_ObjFaninNum(pObj) > pPars.nMaxSuppSize {
            continue; }
        if Nwk_ObjLevel(pLut) - Nwk_ObjLevel(pObj) > pPars.nMaxLevelDiff || 
             Nwk_ObjLevel(pObj) - Nwk_ObjLevel(pLut) > pPars.nMaxLevelDiff {
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
func Nwk_ManCountTotalFanins(pLut *Nwk_Obj_t, pCand *Nwk_Obj_t) int32 {
    var pFanin *Nwk_Obj_t;
    var i int32;
    var nCounter int32 = Nwk_ObjFaninNum(pLut);
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pCand).nFanins)) && ((pFanin) = (pCand).pFanio[i]); i++  {
        nCounter += !pFanin.MarkC;
    }
    return nCounter;
}

/**Function*************************************************************

  Synopsis    [Collects overlapping candidates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCollectOverlapCands(pLut *Nwk_Obj_t, vCands *Vec_Ptr_t, pPars *Nwk_LMPars_t) {
    var pFanin, pObj *Nwk_Obj_t;
    var i, k int32;
    // mark fanins of pLut
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pLut).nFanins)) && ((pFanin) = (pLut).pFanio[i]); i++  {
        pFanin.MarkC = 1;
    }
    // collect the matching fanouts of each fanin of the node
    Vec_PtrClear( vCands );
    Nwk_ManIncrementTravId( pLut.pMan );
    Nwk_ObjSetTravIdCurrent( pLut );
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pLut).nFanins)) && ((pFanin) = (pLut).pFanio[i]); i++  {
        if !Nwk_ObjIsNode(pFanin) {
            continue; }
        if Nwk_ObjFanoutNum(pFanin) > pPars.nMaxFanout {
            continue; }
        // inlining of macro Nwk_ObjForEachFanout
        for  k = 0; (k < int32((pFanin).nFanouts)) && ((pObj) = (pFanin).pFanio[(pFanin).nFanins+k]); k++  {
            if !Nwk_ObjIsNode(pObj) {
                continue; }
            if Nwk_ObjIsTravIdCurrent( pObj ) {
                continue; }
            Nwk_ObjSetTravIdCurrent( pObj );
            // check the difference in delay
            if Nwk_ObjLevel(pLut) - Nwk_ObjLevel(pObj) > pPars.nMaxLevelDiff || 
                 Nwk_ObjLevel(pObj) - Nwk_ObjLevel(pLut) > pPars.nMaxLevelDiff {
                 continue; }
            // check the total number of fanins of the node
            if Nwk_ManCountTotalFanins(pLut, pObj) > pPars.nMaxSuppSize {
                continue; }
            Vec_PtrPush( vCands, pObj );
        }
    }
    // unmark fanins of pLut
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pLut).nFanins)) && ((pFanin) = (pLut).pFanio[i]); i++  {
        pFanin.MarkC = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Performs LUT merging with parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManLutMerge(pNtk *Nwk_Man_t, pParsInit any) *Vec_Int_t {
    var pPars *Nwk_LMPars_t  = (Nwk_LMPars_t *)pParsInit;
    var p *Nwk_Grf_t;
    var vResult *Vec_Int_t;
    var vStart, vNext, vCands1, vCands2 *Vec_Ptr_t;
    var pLut, pCand *Nwk_Obj_t;
    var i, k, nVertsMax, nCands int32;
    var clk abctime  = Abc_Clock();
    // count the number of vertices
    nVertsMax = 0;
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pLut) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pLut) == nil || !Nwk_ObjIsNode(pLut) {} else {
        nVertsMax += int32((Nwk_ObjFaninNum(pLut) <= pPars.nMaxLutSize)); }
    }
    p = Nwk_ManGraphAlloc( nVertsMax );
    // create graph
    vStart  = Vec_PtrAlloc( 1000 );
    vNext   = Vec_PtrAlloc( 1000 );
    vCands1 = Vec_PtrAlloc( 1000 );
    vCands2 = Vec_PtrAlloc( 1000 );
    nCands  = 0;
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pLut) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pLut) == nil || !Nwk_ObjIsNode(pLut) {} else {
        if Nwk_ObjFaninNum(pLut) > pPars.nMaxLutSize {
            continue; }
        Nwk_ManCollectOverlapCands( pLut, vCands1, pPars );
        if pPars.fUseDiffSupp {
            Nwk_ManCollectNonOverlapCands( pLut, vStart, vNext, vCands2, pPars ); }
        if Vec_PtrSize(vCands1) == 0 && Vec_PtrSize(vCands2) == 0 {
            continue; }
        nCands += Vec_PtrSize(vCands1) + Vec_PtrSize(vCands2);
        // save candidates
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vCands1)) && (((pCand) = (Nwk_Obj_t *)Vec_PtrEntry(vCands1, k)), 1); k++  {
            Nwk_ManGraphHashEdge( p, Nwk_ObjId(pLut), Nwk_ObjId(pCand) );
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vCands2)) && (((pCand) = (Nwk_Obj_t *)Vec_PtrEntry(vCands2, k)), 1); k++  {
            Nwk_ManGraphHashEdge( p, Nwk_ObjId(pLut), Nwk_ObjId(pCand) );
        }
        // print statistics about this node
        if pPars.fVeryVerbose {
        fmt.Printf("Node %6d : Fanins = %d. Fanouts = %3d.  Cand1 = %3d. Cand2 = %3d.\n", Nwk_ObjId(pLut), Nwk_ObjFaninNum(pLut), Nwk_ObjFaninNum(pLut), Vec_PtrSize(vCands1), Vec_PtrSize(vCands2)); }
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
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    // solve the graph problem
    clk = Abc_Clock();
    Nwk_ManGraphSolve( p );
    if pPars.fVerbose {
        fmt.Printf("GRAPH: Nodes = %6d. Edges = %6d.  Pairs = %6d.  ", p.nVerts, p.nEdges, Vec_IntSize(p.vPairs)/2);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Solving"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
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

