package rwt

/**CFile****************************************************************

  FileName    [rwtDec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [Evaluation and decomposition procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwtDec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwt.h"
//#include "bool/deco/deco.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Dec_Graph_t * Rwt_NodePreprocess( Rwt_Man_t * p, Rwt_Node_t * pNode );

// static Dec_Edge_t    Rwt_TravCollect_rec( Rwt_Man_t * p, Rwt_Node_t * pNode, Dec_Graph_t * pGraph );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Preprocesses computed library of subgraphs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManPreprocess(p *Rwt_Man_t) {
    var pGraph *Dec_Graph_t;
    var pNode *Rwt_Node_t;
    var i, k int32;
    // put the nodes into the structure
    // inlining of macro ABC_ALLOC
    p.pMapInv  = ((unsigned short *) malloc(sizeof(unsigned short) * (size_t)(222)));
    memset( p.pMapInv, 0, sizeof(unsigned short) * 222 );
    p.vClasses = Vec_VecStart( 222 );
    for  i = 0; i < p.nFuncs; i++  {
        if p.pTable[i] == nil {
            continue; }
        // consider all implementations of this function
        for  pNode = p.pTable[i]; (pNode != nil); pNode = pNode.pNext  {
            assert(pNode.uTruth == p.pTable[i].uTruth, "pNode.uTruth == p.pTable[i].uTruth");
            assert(p.pMap[pNode.uTruth] < 222, "p.pMap[pNode.uTruth] < 222"); // Always >= 0 b/c unsigned.
            Vec_VecPush( p.vClasses, p.pMap[pNode.uTruth], pNode );
            p.pMapInv[ p.pMap[pNode.uTruth] ] = p.puCanons[pNode.uTruth];
        }
    }
    // compute decomposition forms for each node and verify them
    // inlining of macro Vec_VecForEachEntry
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; i < Vec_VecSize(p.vClasses); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(p.vClasses, i))) && (((pNode) = (Rwt_Node_t *)Vec_PtrEntry(Vec_VecEntry(p.vClasses, i), k)), 1); k++  {
        pGraph = Rwt_NodePreprocess( p, pNode );
        pNode.pNext = (Rwt_Node_t *)pGraph;
//        assert( pNode->uTruth == (Dec_GraphDeriveTruth(pGraph) & 0xFFFF) );
    }
    }
}

/**Function*************************************************************

  Synopsis    [Preprocesses subgraphs rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_NodePreprocess(p *Rwt_Man_t, pNode *Rwt_Node_t) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    var eRoot Dec_Edge_t;
    assert(!Rwt_IsComplement(pNode), "!Rwt_IsComplement(pNode)");
    // consider constant
    if pNode.uTruth == 0 {
        return Dec_GraphCreateConst0(); }
    // consider the case of elementary var
    if pNode.uTruth == 0x00FF {
        return Dec_GraphCreateLeaf( 3, 4, 1 ); }
    // start the subgraphs
    pGraph = Dec_GraphCreate( 4 );
    // collect the nodes
    Rwt_ManIncTravId( p );
    eRoot = Rwt_TravCollect_rec( p, pNode, pGraph );
    Dec_GraphSetRoot( pGraph, eRoot );
    return pGraph;
}

/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_TravCollect_rec(p *Rwt_Man_t, pNode *Rwt_Node_t, pGraph *Dec_Graph_t) Dec_Edge_t {
    var eNode0, eNode1, eNode Dec_Edge_t;
    // elementary variable
    if pNode.fUsed {
        return Dec_EdgeCreate( pNode.Id - 1, 0 ); }
    // previously visited node
    if pNode.TravId == p.nTravIds {
        return Dec_IntToEdge( pNode.Volume ); }
    pNode.TravId = p.nTravIds;
    // solve for children
    eNode0 = Rwt_TravCollect_rec( p, Rwt_Regular(pNode.p0), pGraph );
    if Rwt_IsComplement(pNode.p0) {    
        eNode0.fCompl = !eNode0.fCompl; }
    eNode1 = Rwt_TravCollect_rec( p, Rwt_Regular(pNode.p1), pGraph );
    if Rwt_IsComplement(pNode.p1) {    
        eNode1.fCompl = !eNode1.fCompl; }
    // create the decomposition node(s)
    if pNode.fExor {
        eNode = Dec_GraphAddNodeXor( pGraph, eNode0, eNode1, 0 );
    } else {
        eNode = Dec_GraphAddNodeAnd( pGraph, eNode0, eNode1 ); }
    // save the result
    pNode.Volume = Dec_EdgeToInt( eNode );
    return eNode;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

