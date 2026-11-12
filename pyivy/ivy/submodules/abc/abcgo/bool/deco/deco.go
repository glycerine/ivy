package deco

/**CFile****************************************************************

  FileName    [deco.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [A simple decomposition tree/node data structure and its APIs.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: deco.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

   
type Dec_Edge_t struct {
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */   // the complemented bit
    Node uint32; /* C bitfield: 30 bits; review packing and truncation */   // the decomposition node pointed by the edge
}

   
type Dec_Node_t struct {
    eEdge0        Dec_Edge_t;          // the left child of the node
    eEdge1        Dec_Edge_t;          // the right child of the node
    // other info
    union { int       iFunc;           // the literal of the node (AIG)
    void *            pFunc; };        // the function of the node (BDD or AIG)
    Level uint32; /* C bitfield: 14 bits; review packing and truncation */   // the level of this node in the global AIG
    // printing info 
    fNodeOr uint32; /* C bitfield: 1 bits; review packing and truncation */   // marks the original OR node
    fCompl0 uint32; /* C bitfield: 1 bits; review packing and truncation */   // marks the original complemented edge
    fCompl1 uint32; /* C bitfield: 1 bits; review packing and truncation */   // marks the original complemented edge
    // latch info
    nLat0 uint32; /* C bitfield: 5 bits; review packing and truncation */   // the number of latches on the first edge
    nLat1 uint32; /* C bitfield: 5 bits; review packing and truncation */   // the number of latches on the second edge
    nLat2 uint32; /* C bitfield: 5 bits; review packing and truncation */   // the number of latches on the output edge
}

   
type Dec_Graph_t struct {
    fConst               int32;          // marks the constant 1 graph
    nLeaves               int32;         // the number of leaves
    nSize               int32;           // the number of nodes (including the leaves) 
    nCap               int32;            // the number of allocated nodes
    pNodes       *Dec_Node_t;          // the array of leaves and internal nodes
    eRoot        Dec_Edge_t;           // the pointer to the topmost node
}

   
type Dec_Man_t struct {
    pMvcMem             any;         // memory manager for MVC cover (used for factoring)
    vCubes        *Vec_Int_t;          // storage for cubes
    vLits        *Vec_Int_t;           // storage for literals 
    // precomputation information about 4-variable functions
    unsigned short *  puCanons;        // canonical forms
    pPhases             string;         // canonical phases
    pPerms             string;          // canonical permutations
    unsigned char *   pMap;            // mapping of functions into class numbers
}


////////////////////////////////////////////////////////////////////////
///                        ITERATORS                                 ///
////////////////////////////////////////////////////////////////////////

// iterator through the leaves
// Original macro:
// #define Dec_GraphForEachLeaf( pGraph, pLeaf, i )                                              \
//     for ( i = 0; (i < (pGraph)->nLeaves) && (((pLeaf) = Dec_GraphNode(pGraph, i)), 1); i++ )

// iterator through the internal nodes
// Original macro:
// #define Dec_GraphForEachNode( pGraph, pAnd, i )                                               \
//     for ( i = (pGraph)->nLeaves; (i < (pGraph)->nSize) && (((pAnd) = Dec_GraphNode(pGraph, i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates an edge pointing to the node in the given polarity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_EdgeCreate(Node int32, fCompl int32) Dec_Edge_t {
    var eEdge Dec_Edge_t  = Dec_Edge_t{ uint32(fCompl), uint32(Node) }; 
    return eEdge; 
}

/**Function*************************************************************

  Synopsis    [Converts the edge into unsigned integer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_EdgeToInt(eEdge Dec_Edge_t) uint32 {
    return (eEdge.Node << 1) | eEdge.fCompl; 
}

/**Function*************************************************************

  Synopsis    [Converts unsigned integer into the edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_IntToEdge(Edge uint32) Dec_Edge_t {
    return Dec_EdgeCreate( Edge >> 1, Edge & 1 ); 
}

/**Function*************************************************************

  Synopsis    [Converts the edge into unsigned integer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_EdgeToInt_(m Dec_Edge_t) uint32 { union { var x Dec_Edge_t; var y uint32; }
 v; v.x = m; return v.y;  }
/*
static inline unsigned Dec_EdgeToInt_( Dec_Edge_t eEdge )   
{
    return *(unsigned *)&eEdge;
}
*/

/**Function*************************************************************

  Synopsis    [Converts unsigned integer into the edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_IntToEdge_(m uint32) Dec_Edge_t { union { var x Dec_Edge_t; var y uint32; }
 v; v.y = m; return v.x;  }
/*
static inline Dec_Edge_t Dec_IntToEdge_( unsigned Edge )   
{
    return *(Dec_Edge_t *)&Edge;
}
*/

/**Function*************************************************************

  Synopsis    [Creates a graph with the given number of leaves.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphCreate(nLeaves int32) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    pGraph = new(Dec_Graph_t);
    *pGraph = Dec_Graph_t{};
    pGraph.nLeaves = nLeaves;
    pGraph.nSize = nLeaves;
    pGraph.nCap = 2 * nLeaves + 50;
    pGraph.pNodes = ABC_ALLOC( Dec_Node_t, pGraph.nCap );
    memset( pGraph.pNodes, 0, sizeof(Dec_Node_t) * pGraph.nSize );
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Creates constant 0 graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphCreateConst0() *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    pGraph = new(Dec_Graph_t);
    *pGraph = Dec_Graph_t{};
    pGraph.fConst = 1;
    pGraph.eRoot.fCompl = 1;
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Creates constant 1 graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphCreateConst1() *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    pGraph = new(Dec_Graph_t);
    *pGraph = Dec_Graph_t{};
    pGraph.fConst = 1;
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Creates the literal graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphCreateLeaf(iLeaf int32, nLeaves int32, fCompl int32) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    assert(0 <= iLeaf && iLeaf < nLeaves, "0 <= iLeaf && iLeaf < nLeaves");
    pGraph = Dec_GraphCreate( nLeaves );
    pGraph.eRoot.Node   = iLeaf;
    pGraph.eRoot.fCompl = fCompl;
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Creates a graph with the given number of leaves.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphFree(pGraph *Dec_Graph_t) {
    ABC_FREE( pGraph.pNodes );
    ABC_FREE( pGraph );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the graph is a constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphIsConst(pGraph *Dec_Graph_t) int32 {
    return pGraph.fConst;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the graph is constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphIsConst0(pGraph *Dec_Graph_t) int32 {
    return pGraph.fConst && pGraph.eRoot.fCompl;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the graph is constant 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphIsConst1(pGraph *Dec_Graph_t) int32 {
    return pGraph.fConst && !pGraph.eRoot.fCompl;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the graph is complemented.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphIsComplement(pGraph *Dec_Graph_t) int32 {
    return pGraph.eRoot.fCompl;
}

/**Function*************************************************************

  Synopsis    [Checks if the graph is complemented.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphComplement(pGraph *Dec_Graph_t) {
    pGraph.eRoot.fCompl ^= 1;
}


/**Function*************************************************************

  Synopsis    [Returns the number of leaves.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphLeaveNum(pGraph *Dec_Graph_t) int32 {
    return pGraph.nLeaves;
}

/**Function*************************************************************

  Synopsis    [Returns the number of internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphNodeNum(pGraph *Dec_Graph_t) int32 {
    return pGraph.nSize - pGraph.nLeaves;
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphNode(pGraph *Dec_Graph_t, i int32) *Dec_Node_t {
    return pGraph.pNodes + i;
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphNodeLast(pGraph *Dec_Graph_t) *Dec_Node_t {
    return pGraph.pNodes + pGraph.nSize - 1;
}

/**Function*************************************************************

  Synopsis    [Returns the number of the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphNodeInt(pGraph *Dec_Graph_t, pNode *Dec_Node_t) int32 {
    return pNode - pGraph.pNodes;
}

/**Function*************************************************************

  Synopsis    [Check if the graph represents elementary variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphIsVar(pGraph *Dec_Graph_t) int32 {
    return pGraph.eRoot.Node < uint32(pGraph.nLeaves);
}

/**Function*************************************************************

  Synopsis    [Check if the graph represents elementary variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphNodeIsVar(pGraph *Dec_Graph_t, pNode *Dec_Node_t) int32 {
    return Dec_GraphNodeInt(pGraph,pNode) < pGraph.nLeaves;
}

/**Function*************************************************************

  Synopsis    [Returns the elementary variable elementary variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphVar(pGraph *Dec_Graph_t) *Dec_Node_t {
    assert((Dec_GraphIsVar( pGraph ) != 0), "Dec_GraphIsVar( pGraph )");
    return Dec_GraphNode( pGraph, pGraph.eRoot.Node );
}

/**Function*************************************************************

  Synopsis    [Returns the number of the elementary variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphVarInt(pGraph *Dec_Graph_t) int32 {
    assert((Dec_GraphIsVar( pGraph ) != 0), "Dec_GraphIsVar( pGraph )");
    return Dec_GraphNodeInt( pGraph, Dec_GraphVar(pGraph) );
}

/**Function*************************************************************

  Synopsis    [Sets the root of the graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphSetRoot(pGraph *Dec_Graph_t, eRoot Dec_Edge_t) {
    pGraph.eRoot = eRoot;
}

/**Function*************************************************************

  Synopsis    [Appends a new node to the graph.]

  Description [This procedure is meant for internal use.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphAppendNode(pGraph *Dec_Graph_t) *Dec_Node_t {
    var pNode *Dec_Node_t;
    if pGraph.nSize == pGraph.nCap {
        pGraph.pNodes = ABC_REALLOC( Dec_Node_t, pGraph.pNodes, 2 * pGraph.nCap ); 
        pGraph.nCap   = 2 * pGraph.nCap;
    }
    pNode = pGraph.pNodes + pGraph.nSize++;
    *pNode = Dec_Node_t{};
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Creates an AND node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphAddNodeAnd(pGraph *Dec_Graph_t, eEdge0 Dec_Edge_t, eEdge1 Dec_Edge_t) Dec_Edge_t {
    var pNode *Dec_Node_t;
    // get the new node
    pNode = Dec_GraphAppendNode( pGraph );
    // set the inputs and other info
    pNode.eEdge0 = eEdge0;
    pNode.eEdge1 = eEdge1;
    pNode.fCompl0 = eEdge0.fCompl;
    pNode.fCompl1 = eEdge1.fCompl;
    return Dec_EdgeCreate( pGraph.nSize - 1, 0 );
}

/**Function*************************************************************

  Synopsis    [Creates an OR node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphAddNodeOr(pGraph *Dec_Graph_t, eEdge0 Dec_Edge_t, eEdge1 Dec_Edge_t) Dec_Edge_t {
    var pNode *Dec_Node_t;
    // get the new node
    pNode = Dec_GraphAppendNode( pGraph );
    // set the inputs and other info
    pNode.eEdge0 = eEdge0;
    pNode.eEdge1 = eEdge1;
    pNode.fCompl0 = eEdge0.fCompl;
    pNode.fCompl1 = eEdge1.fCompl;
    // make adjustments for the OR gate
    pNode.fNodeOr = 1;
    pNode.eEdge0.fCompl = !pNode.eEdge0.fCompl;
    pNode.eEdge1.fCompl = !pNode.eEdge1.fCompl;
    return Dec_EdgeCreate( pGraph.nSize - 1, 1 );
}

/**Function*************************************************************

  Synopsis    [Creates an XOR node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphAddNodeXor(pGraph *Dec_Graph_t, eEdge0 Dec_Edge_t, eEdge1 Dec_Edge_t, Type int32) Dec_Edge_t {
    var eNode0, eNode1, eNode Dec_Edge_t;
    if Type == 0 {
        // derive the first AND
        eEdge0.fCompl ^= 1;
        eNode0 = Dec_GraphAddNodeAnd( pGraph, eEdge0, eEdge1 );
        eEdge0.fCompl ^= 1;
        // derive the second AND
        eEdge1.fCompl ^= 1;
        eNode1 = Dec_GraphAddNodeAnd( pGraph, eEdge0, eEdge1 );
        // derive the final OR
        eNode = Dec_GraphAddNodeOr( pGraph, eNode0, eNode1 );
    
    } else {
        // derive the first AND
        eNode0 = Dec_GraphAddNodeAnd( pGraph, eEdge0, eEdge1 );
        // derive the second AND
        eEdge0.fCompl ^= 1;
        eEdge1.fCompl ^= 1;
        eNode1 = Dec_GraphAddNodeAnd( pGraph, eEdge0, eEdge1 );
        // derive the final OR
        eNode = Dec_GraphAddNodeOr( pGraph, eNode0, eNode1 );
        eNode.fCompl ^= 1;
    }
    return eNode;
}

/**Function*************************************************************

  Synopsis    [Creates an XOR node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphAddNodeMux(pGraph *Dec_Graph_t, eEdgeC Dec_Edge_t, eEdgeT Dec_Edge_t, eEdgeE Dec_Edge_t, Type int32) Dec_Edge_t {
    var eNode0, eNode1, eNode Dec_Edge_t;
    if Type == 0 {
        // derive the first AND
        eNode0 = Dec_GraphAddNodeAnd( pGraph, eEdgeC, eEdgeT );
        // derive the second AND
        eEdgeC.fCompl ^= 1;
        eNode1 = Dec_GraphAddNodeAnd( pGraph, eEdgeC, eEdgeE );
        // derive the final OR
        eNode = Dec_GraphAddNodeOr( pGraph, eNode0, eNode1 );
    
    } else {
        // complement the arguments
        eEdgeT.fCompl ^= 1;
        eEdgeE.fCompl ^= 1;
        // derive the first AND
        eNode0 = Dec_GraphAddNodeAnd( pGraph, eEdgeC, eEdgeT );
        // derive the second AND
        eEdgeC.fCompl ^= 1;
        eNode1 = Dec_GraphAddNodeAnd( pGraph, eEdgeC, eEdgeE );
        // derive the final OR
        eNode = Dec_GraphAddNodeOr( pGraph, eNode0, eNode1 );
        eNode.fCompl ^= 1;
    }
    return eNode;
}



//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

