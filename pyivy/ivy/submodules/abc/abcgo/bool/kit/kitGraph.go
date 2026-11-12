package kit

import (
	"fmt"
	"strconv"
)

/**CFile****************************************************************

  FileName    [kitGraph.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Computation kit.]

  Synopsis    [Decomposition graph representation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 6, 2006.]

  Revision    [$Id: kitGraph.c,v 1.00 2006/12/06 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates a graph with the given number of leaves.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphCreate(nLeaves int32) *Kit_Graph_t {
    var pGraph *Kit_Graph_t;
    pGraph = new(Kit_Graph_t);
    memset( pGraph, 0, sizeof(Kit_Graph_t) );
    pGraph.nLeaves = nLeaves;
    pGraph.nSize = nLeaves;
    pGraph.nCap = 2 * nLeaves + 50;
    pGraph.pNodes = ABC_ALLOC( Kit_Node_t, pGraph.nCap );
    memset( pGraph.pNodes, 0, sizeof(Kit_Node_t) * pGraph.nSize );
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Creates constant 0 graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphCreateConst0() *Kit_Graph_t {
    var pGraph *Kit_Graph_t;
    pGraph = new(Kit_Graph_t);
    memset( pGraph, 0, sizeof(Kit_Graph_t) );
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
func Kit_GraphCreateConst1() *Kit_Graph_t {
    var pGraph *Kit_Graph_t;
    pGraph = new(Kit_Graph_t);
    memset( pGraph, 0, sizeof(Kit_Graph_t) );
    pGraph.fConst = 1;
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Creates the literal graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphCreateLeaf(iLeaf int32, nLeaves int32, fCompl int32) *Kit_Graph_t {
    var pGraph *Kit_Graph_t;
    assert(0 <= iLeaf && iLeaf < nLeaves, "0 <= iLeaf && iLeaf < nLeaves");
    pGraph = Kit_GraphCreate( nLeaves );
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
func Kit_GraphFree(pGraph *Kit_Graph_t) {
    ABC_FREE( pGraph.pNodes );
    ABC_FREE( pGraph );
}

/**Function*************************************************************

  Synopsis    [Appends a new node to the graph.]

  Description [This procedure is meant for internal use.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphAppendNode(pGraph *Kit_Graph_t) *Kit_Node_t {
    var pNode *Kit_Node_t;
    if pGraph.nSize == pGraph.nCap {
        pGraph.pNodes = ABC_REALLOC( Kit_Node_t, pGraph.pNodes, 2 * pGraph.nCap ); 
        pGraph.nCap   = 2 * pGraph.nCap;
    }
    pNode = pGraph.pNodes + pGraph.nSize++;
    memset( pNode, 0, sizeof(Kit_Node_t) );
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Creates an AND node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphAddNodeAnd(pGraph *Kit_Graph_t, eEdge0 Kit_Edge_t, eEdge1 Kit_Edge_t) Kit_Edge_t {
    var pNode *Kit_Node_t;
    // get the new node
    pNode = Kit_GraphAppendNode( pGraph );
    // set the inputs and other info
    pNode.eEdge0 = eEdge0;
    pNode.eEdge1 = eEdge1;
    pNode.fCompl0 = eEdge0.fCompl;
    pNode.fCompl1 = eEdge1.fCompl;
    return Kit_EdgeCreate( pGraph.nSize - 1, 0 );
}

/**Function*************************************************************

  Synopsis    [Creates an OR node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphAddNodeOr(pGraph *Kit_Graph_t, eEdge0 Kit_Edge_t, eEdge1 Kit_Edge_t) Kit_Edge_t {
    var pNode *Kit_Node_t;
    // get the new node
    pNode = Kit_GraphAppendNode( pGraph );
    // set the inputs and other info
    pNode.eEdge0 = eEdge0;
    pNode.eEdge1 = eEdge1;
    pNode.fCompl0 = eEdge0.fCompl;
    pNode.fCompl1 = eEdge1.fCompl;
    // make adjustments for the OR gate
    pNode.fNodeOr = 1;
    pNode.eEdge0.fCompl = !pNode.eEdge0.fCompl;
    pNode.eEdge1.fCompl = !pNode.eEdge1.fCompl;
    return Kit_EdgeCreate( pGraph.nSize - 1, 1 );
}

/**Function*************************************************************

  Synopsis    [Creates an XOR node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphAddNodeXor(pGraph *Kit_Graph_t, eEdge0 Kit_Edge_t, eEdge1 Kit_Edge_t, Type int32) Kit_Edge_t {
    var eNode0, eNode1, eNode Kit_Edge_t;
    if Type == 0 {
        // derive the first AND
        eEdge0.fCompl ^= 1;
        eNode0 = Kit_GraphAddNodeAnd( pGraph, eEdge0, eEdge1 );
        eEdge0.fCompl ^= 1;
        // derive the second AND
        eEdge1.fCompl ^= 1;
        eNode1 = Kit_GraphAddNodeAnd( pGraph, eEdge0, eEdge1 );
        // derive the final OR
        eNode = Kit_GraphAddNodeOr( pGraph, eNode0, eNode1 );
    
    } else {
        // derive the first AND
        eNode0 = Kit_GraphAddNodeAnd( pGraph, eEdge0, eEdge1 );
        // derive the second AND
        eEdge0.fCompl ^= 1;
        eEdge1.fCompl ^= 1;
        eNode1 = Kit_GraphAddNodeAnd( pGraph, eEdge0, eEdge1 );
        // derive the final OR
        eNode = Kit_GraphAddNodeOr( pGraph, eNode0, eNode1 );
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
func Kit_GraphAddNodeMux(pGraph *Kit_Graph_t, eEdgeC Kit_Edge_t, eEdgeT Kit_Edge_t, eEdgeE Kit_Edge_t, Type int32) Kit_Edge_t {
    var eNode0, eNode1, eNode Kit_Edge_t;
    if Type == 0 {
        // derive the first AND
        eNode0 = Kit_GraphAddNodeAnd( pGraph, eEdgeC, eEdgeT );
        // derive the second AND
        eEdgeC.fCompl ^= 1;
        eNode1 = Kit_GraphAddNodeAnd( pGraph, eEdgeC, eEdgeE );
        // derive the final OR
        eNode = Kit_GraphAddNodeOr( pGraph, eNode0, eNode1 );
    
    } else {
        // complement the arguments
        eEdgeT.fCompl ^= 1;
        eEdgeE.fCompl ^= 1;
        // derive the first AND
        eNode0 = Kit_GraphAddNodeAnd( pGraph, eEdgeC, eEdgeT );
        // derive the second AND
        eEdgeC.fCompl ^= 1;
        eNode1 = Kit_GraphAddNodeAnd( pGraph, eEdgeC, eEdgeE );
        // derive the final OR
        eNode = Kit_GraphAddNodeOr( pGraph, eNode0, eNode1 );
        eNode.fCompl ^= 1;
    }
    return eNode;
}

/**Function*************************************************************

  Synopsis    [Derives the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphToTruth(pGraph *Kit_Graph_t) uint32 {
    var uTruths [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
    var uTruth, uTruth0, uTruth1 uint32;
    var pNode *Kit_Node_t;
    var i int32;

    // sanity checks
    assert(Kit_GraphLeaveNum(pGraph) >= 0, "Kit_GraphLeaveNum(pGraph) >= 0");
    assert(Kit_GraphLeaveNum(pGraph) <= pGraph.nSize, "Kit_GraphLeaveNum(pGraph) <= pGraph.nSize");
    assert(Kit_GraphLeaveNum(pGraph) <= 5, "Kit_GraphLeaveNum(pGraph) <= 5");

    // check for constant function
    if Kit_GraphIsConst(pGraph) {
        if Kit_GraphIsComplement(pGraph) {
            return 0;
        } else {
            return ~(uint32(0));
        } }
    // check for a literal
    if Kit_GraphIsVar(pGraph) {
        if Kit_GraphIsComplement(pGraph) {
            return ~uTruths[Kit_GraphVarInt(pGraph)];
        } else {
            return uTruths[Kit_GraphVarInt(pGraph)];
        } }

    // assign the elementary variables
    // inlining of macro Kit_GraphForEachLeaf
    for  i = 0; (i < (pGraph).nLeaves) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pNode.pFunc = (void *)int64(uTruths[i]);
    }

    // compute the function for each internal node
    // inlining of macro Kit_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        uTruth0 = uint32(int64(Kit_GraphNode(pGraph, pNode.eEdge0.Node).pFunc));
        uTruth1 = uint32(int64(Kit_GraphNode(pGraph, pNode.eEdge1.Node).pFunc));
        if pNode.eEdge0.fCompl {
            uTruth0 = ~uTruth0;
        } else {
            uTruth0 = uTruth0;
        }
        if pNode.eEdge1.fCompl {
            uTruth1 = ~uTruth1;
        } else {
            uTruth1 = uTruth1;
        }
        uTruth = uTruth0 & uTruth1;
        pNode.pFunc = (void *)int64(uTruth);
    }

    // complement the result if necessary
    if Kit_GraphIsComplement(pGraph) {
        return ~uTruth;
    } else {
        return uTruth;
    }
}

/**Function*************************************************************

  Synopsis    [Derives the factored form from the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_TruthToGraph(pTruth *uint32, nVars int32, vMemory *Vec_Int_t) *Kit_Graph_t {
    var pGraph *Kit_Graph_t;
    var RetValue int32;
    // derive SOP
    RetValue = Kit_TruthIsop( pTruth, nVars, vMemory, 1 ); // tried 1 and found not useful in "renode"
    if RetValue == -1 {
        return nil; }
    if Vec_IntSize(vMemory) > (1<<16) {
        return nil; }
//    printf( "Isop size = %d.\n", Vec_IntSize(vMemory) );
    assert(RetValue == 0 || RetValue == 1, "RetValue == 0 || RetValue == 1");
    // derive factored form
    pGraph = Kit_SopFactor( vMemory, RetValue, nVars, vMemory );
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Derives the factored form from the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_TruthToGraph2(pTruth0 *uint32, pTruth1 *uint32, nVars int32, vMemory *Vec_Int_t) *Kit_Graph_t {
    var pGraph *Kit_Graph_t;
    var RetValue int32;
    // derive SOP
    RetValue = Kit_TruthIsop2( pTruth0, pTruth1, nVars, vMemory, 1, 0 ); // tried 1 and found not useful in "renode"
    if RetValue == -1 {
        return nil; }
    if Vec_IntSize(vMemory) > (1<<16) {
        return nil; }
//    printf( "Isop size = %d.\n", Vec_IntSize(vMemory) );
    assert(RetValue == 0 || RetValue == 1, "RetValue == 0 || RetValue == 1");
    // derive factored form
    pGraph = Kit_SopFactor( vMemory, RetValue, nVars, vMemory );
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Derives the maximum depth from the leaf to the root.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphLeafDepth_rec(pGraph *Kit_Graph_t, pNode *Kit_Node_t, pLeaf *Kit_Node_t) int32 {
    var Depth0, Depth1, Depth int32;
    if pNode == pLeaf {
        return 0; }
    if Kit_GraphNodeIsVar(pGraph, pNode) {
        return -100; }
    Depth0 = Kit_GraphLeafDepth_rec( pGraph, Kit_GraphNodeFanin0(pGraph, pNode), pLeaf );
    Depth1 = Kit_GraphLeafDepth_rec( pGraph, Kit_GraphNodeFanin1(pGraph, pNode), pLeaf );
    // inlining of macro KIT_MAX
    
    if (Depth0) > (Depth1) {
        Depth = (Depth0);
    } else {
        Depth = (Depth1);
    }
    if Depth == -100 {
        Depth = -100;
    } else {
        Depth = Depth + 1;
    }
    return Depth;
}

/**Function*************************************************************

  Synopsis    [Derives logic level of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphLevelNum_rec(pGraph *Kit_Graph_t, pNode *Kit_Node_t) int32 {
    var Depth0, Depth1 int32;
    if Kit_GraphNodeIsVar(pGraph, pNode) {
        return 0; }
    Depth0 = Kit_GraphLevelNum_rec( pGraph, Kit_GraphNodeFanin0(pGraph, pNode) );
    Depth1 = Kit_GraphLevelNum_rec( pGraph, Kit_GraphNodeFanin1(pGraph, pNode) );
    // inlining of macro KIT_MAX
    return 1 + (((Depth0) > (Depth1))? (Depth0) : (Depth1));
}

/**Function*************************************************************

  Synopsis    [Returns FF nodes and levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_TruthStats(pTruth *uint32, nVars int32, vMemory *Vec_Int_t) int32 {
    var pGraph *Kit_Graph_t  = Kit_TruthToGraph( pTruth, nVars, vMemory );
    var nNodes int32   = Kit_GraphNodeNum( pGraph );
    var nLevels int32  = Kit_GraphLevelNum_rec( pGraph, Kit_GraphNodeLast(pGraph) );
    Kit_GraphFree( pGraph );
    return (nLevels << 16) | nNodes;
}
func Kit_TruthStatsArray(pArray *uint32, nVars int32, nFuncs int32) *int32 {
    var f int32;
    var pRes *int32 = ABC_CALLOC( int, nFuncs );
    var nInts int32  = Abc_TruthWordNum( nVars );
    var vMemory *Vec_Int_t  = Vec_IntAlloc( 1 << 16 );
    for  f = 0; f < nFuncs; f++  {
        pRes[f] = Kit_TruthStats( pArray + f*nInts, nVars, vMemory );
    }
    Vec_IntFree( vMemory );
    return pRes;
}
func Kit_TruthFindVarNum(pFileName string) int32 {
    var i int32;
    for  i = 0; i < int32(len(pFileName)); i++  {
        if pFileName[i] >= '0' && pFileName[i] <= '9' {
            return strconv.Atoi(pFileName+i); }
    }
    return -1;
}
func Kit_TruthTest(pFileName string) *int32 {
    var clk abctime    = Abc_Clock(); var i int32;
    var nFileSize int32  = Extra_FileSize( pFileName );
    var nVars int32      = Kit_TruthFindVarNum( pFileName );
    var nFuncs int32     = nFileSize / 4 / Abc_TruthWordNum(nVars);
    var pA *uint32  = (unsigned *)Extra_FileReadContents( pFileName );
    var pResult *int32  = Kit_TruthStatsArray( pA, nVars, nFuncs );
    fmt.Printf("Finished proceessing %d functions with %d variables. ", nFuncs, nVars);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    ABC_FREE( pA );
    for  i = 0; i < 5; i++  {
        fmt.Printf("Function %3d :  AND2 = %3d  Lev = %3d\n", i, pResult[i] & 0xFFFF, pResult[i] >> 16);
    }
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Derives the factored form from the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_TruthLitNum(pTruth *uint32, nVars int32, vMemory *Vec_Int_t) int32 {
    var pGraph *Kit_Graph_t;
    var RetValue, nLits int32;
    RetValue = Kit_TruthIsop( pTruth, nVars, vMemory, 1 ); 
    if RetValue == -1 || Vec_IntSize(vMemory) > (1<<16) {
        return -1; }
    assert(RetValue == 0 || RetValue == 1, "RetValue == 0 || RetValue == 1");
    pGraph = Kit_SopFactor( vMemory, RetValue, nVars, vMemory );
    nLits = 1 + Kit_GraphNodeNum( pGraph );
    Kit_GraphFree( pGraph );
    return nLits;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

