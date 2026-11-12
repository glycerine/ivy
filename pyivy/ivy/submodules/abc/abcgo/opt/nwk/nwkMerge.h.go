package nwk

/**CFile****************************************************************

  FileName    [nwkMerge.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkMerge.h,v 1.1 2008/05/14 22:13:09 wudenni Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


const NWK_MAX_LIST =  16

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// the LUT merging parameters
    
type Nwk_LMPars_t struct {
    nMaxLutSize    int32;    // the max LUT size for merging (N=5)
    nMaxSuppSize    int32;   // the max total support size after merging (S=5)
    nMaxDistance    int32;   // the max number of nodes separating LUTs
    nMaxLevelDiff    int32;  // the max difference in levels
    nMaxFanout    int32;     // the max number of fanouts to traverse
    fUseDiffSupp    int32;   // enables the use of nodes with different support
    fUseTfiTfo    int32;     // enables the use of TFO/TFO nodes as candidates
    fVeryVerbose    int32;   // enables additional verbose output
    fVerbose    int32;       // enables verbose output
}

// edge of the graph
    
type Nwk_Edg_t struct {
    iNode1             int32;      // the first node
    iNode2             int32;      // the second node
    pNext      *Nwk_Edg_t;       // the next edge
}

// vertex of the graph
    
type Nwk_Vrt_t struct {
    Id             int32;          // the vertex number
    iPrev             int32;       // the previous vertex in the list
    iNext             int32;       // the next vertex in the list
    nEdges             int32;      // the number of edges
    pEdges [0]int32;   // the array of edges
}

// the connectivity graph
    
type Nwk_Grf_t struct {
    // preliminary graph representation
    nObjs             int32;       // the number of objects
    nVertsMax             int32;   // the upper bound on the number of vertices
    nEdgeHash             int32;   // an approximate number of edges
    pEdgeHash     **Nwk_Edg_t;   // hash table for edges
    pMemEdges  *Aig_MmFixed_t;   // memory for edges
    // graph representation
    nEdges             int32;      // the number of edges
    nVerts             int32;      // the number of vertices
    pVerts     **Nwk_Vrt_t;      // the array of vertices
    pMemVerts   *Aig_MmFlex_t;   // memory for vertices
    // intermediate data
    pLists1 [NWK_MAX_LIST+1]int32; // lists of nodes with one edge
    pLists2 [NWK_MAX_LIST+1]int32; // lists of nodes with more than one edge
    // the results of matching
    vPairs      *Vec_Int_t;      // pairs matched in the graph
    // object mappings
    pMapLut2Id            *int32;  // LUT numbers into vertex IDs
    pMapId2Lut            *int32;  // vertex IDs into LUT numbers
    // other things
    nMemBytes1             int32;  // memory usage in bytes
    nMemBytes2             int32;  // memory usage in bytes
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Nwk_GraphForEachEdge( p, pEdge, k )             \
//     for ( k = 0; k < p->nEdgeHash; k++ )                \
//         for ( pEdge = p->pEdgeHash[k]; pEdge; pEdge = pEdge->pNext )


// Original macro:
// #define Nwk_ListForEachVertex( p, List, pVrt )          \
//     for ( pVrt = List? p->pVerts[List] : NULL;  pVrt;   \
//           pVrt = pVrt->iNext? p->pVerts[pVrt->iNext] : NULL )


// Original macro:
// #define Nwk_VertexForEachAdjacent( p, pVrt, pNext, k )  \
//     for ( k = 0; (k < pVrt->nEdges) && (((pNext) = p->pVerts[pVrt->pEdges[k]]), 1); k++ )


////////////////////////////////////////////////////////////////////////
///                      INLINED FUNCTIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         ITERATORS                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== nwkMerge.c ==========================================================*/
extern ABC_DLL Nwk_Grf_t *   Nwk_ManGraphAlloc( int nVertsMax );
extern ABC_DLL void          Nwk_ManGraphFree( Nwk_Grf_t * p );
extern ABC_DLL void          Nwk_ManGraphReportMemoryUsage( Nwk_Grf_t * p );
extern ABC_DLL void          Nwk_ManGraphHashEdge( Nwk_Grf_t * p, int iLut1, int iLut2 );
extern ABC_DLL void          Nwk_ManGraphSolve( Nwk_Grf_t * p );
extern ABC_DLL int           Nwk_ManLutMergeGraphTest( char * pFileName );



//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

