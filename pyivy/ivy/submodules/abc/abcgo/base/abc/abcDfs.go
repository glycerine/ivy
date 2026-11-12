package abc

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcDfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures that use depth-first search.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcDfs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"
//#include "proof/cec/cec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfs_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // skip the CI
    if Abc_ObjIsCi(pNode) || (Abc_NtkIsStrash(pNode.pNtk) && Abc_AigNodeIsConst(pNode)) {
        return; }
    assert(Abc_ObjIsNode( pNode ) || Abc_ObjIsBox( pNode ), "Abc_ObjIsNode( pNode ) || Abc_ObjIsBox( pNode )");
    // visit the transitive fanin of the node
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
//        pFanin = Abc_ObjFanin( pNode, Abc_ObjFaninNum(pNode)-1-i );
        Abc_NtkDfs_rec( Abc_ObjFanin0Ntk(pFanin), vNodes );
    }
    // add the node after the fanins have been added
    Vec_PtrPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving out CIs and CO.
  However it marks with the current TravId both CIs and COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfs(pNtk *Abc_Ntk_t, fCollectAll int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    if pNtk.nBarBufs2 > 0 {
        // inlining of macro Abc_NtkForEachBarBuf
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsBarBuf(pObj) {} else {
            Abc_NodeSetTravIdCurrent( pObj );
            Abc_NtkDfs_rec( Abc_ObjFanin0Ntk(Abc_ObjFanin0(pObj)), vNodes );
            Vec_PtrPush( vNodes, pObj );
        }
        }
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
        Abc_NtkDfs_rec( Abc_ObjFanin0Ntk(Abc_ObjFanin0(pObj)), vNodes );
    }
    // collect dangling nodes if asked to
    if fCollectAll != 0 {
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
            if !Abc_NodeIsTravIdCurrent(pObj) {
                Abc_NtkDfs_rec( pObj, vNodes ); }
        }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving out CIs and CO.
  However it marks with the current TravId both CIs and COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfs2(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
    var pObj *Abc_Obj_t;  var i int32;
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
        Abc_NtkDfs_rec( Abc_ObjFanin0Ntk(Abc_ObjFanin0(pObj)), vNodes );
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving out PIs, POs and latches.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsNodes(pNtk *Abc_Ntk_t, ppNodes **Abc_Obj_t, nNodes int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var i int32;
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // go through the PO nodes and call for each of them
    for  i = 0; i < nNodes; i++  {
        if Abc_NtkIsStrash(pNtk) && Abc_AigNodeIsConst(ppNodes[i]) {
            continue; }
        if Abc_ObjIsCo(ppNodes[i]) {
            Abc_NodeSetTravIdCurrent(ppNodes[i]);
            Abc_NtkDfs_rec( Abc_ObjFanin0Ntk(Abc_ObjFanin0(ppNodes[i])), vNodes );
        
        } else if Abc_ObjIsNode(ppNodes[i]) || Abc_ObjIsCi(ppNodes[i]) {
            Abc_NtkDfs_rec( ppNodes[i], vNodes ); }
    }
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsReverse_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // skip the CI
    if Abc_ObjIsCo(pNode) {
        return; }
    assert(Abc_ObjIsNode( pNode ), "Abc_ObjIsNode( pNode )");
    // visit the transitive fanin of the node
    pNode = Abc_ObjFanout0Ntk(pNode);
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
        Abc_NtkDfsReverse_rec( pFanout, vNodes );
    }
    // add the node after the fanins have been added
    Vec_PtrPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the reverse DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving out CIs/COs.
  However it marks both CIs and COs with the current TravId.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsReverse(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj, pFanout *Abc_Obj_t;
    var i, k int32;
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
        pObj = Abc_ObjFanout0Ntk(pObj);
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_NtkDfsReverse_rec( pFanout, vNodes );
        }
    }
    // add constant nodes in the end
    if !Abc_NtkIsStrash(pNtk) {
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
            if Abc_NodeIsConst(pObj) {
                Vec_PtrPush( vNodes, pObj ); }
        }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsReverseNodes_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // skip the CI
    if Abc_ObjIsCo(pNode) {
        return; }
    assert(Abc_ObjIsNode( pNode ), "Abc_ObjIsNode( pNode )");
    // visit the transitive fanin of the node
    pNode = Abc_ObjFanout0Ntk(pNode);
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
        Abc_NtkDfsReverseNodes_rec( pFanout, vNodes );
    }
    // add the node after the fanins have been added
//    Vec_PtrPush( vNodes, pNode );
    Vec_PtrFillExtra( vNodes, pNode.Level + 1, nil );
    pNode.pCopy = (Abc_Obj_t *)Vec_PtrEntry( vNodes, pNode.Level );
    Vec_PtrWriteEntry( vNodes, pNode.Level, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the levelized array of TFO nodes.]

  Description [Collects the levelized array of internal nodes, leaving out CIs/COs.
  However it marks both CIs and COs with the current TravId.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsReverseNodes(pNtk *Abc_Ntk_t, ppNodes **Abc_Obj_t, nNodes int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj, pFanout *Abc_Obj_t;
    var i, k int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrStart( Abc_AigLevel(pNtk) + 1 );
    for  i = 0; i < nNodes; i++  {
        pObj = ppNodes[i];
        assert(Abc_ObjIsCi(pObj), "Abc_ObjIsCi(pObj)");
        Abc_NodeSetTravIdCurrent( pObj );
        pObj = Abc_ObjFanout0Ntk(pObj);
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_NtkDfsReverseNodes_rec( pFanout, vNodes );
        }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns the levelized array of TFO nodes.]

  Description [Collects the levelized array of internal nodes, leaving out CIs/COs.
  However it marks both CIs and COs with the current TravId.
  Collects only the nodes whose support does not exceed the set of given CI nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsReverseNodesContained(pNtk *Abc_Ntk_t, ppNodes **Abc_Obj_t, nNodes int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj, pFanout, pFanin *Abc_Obj_t;
    var i, k, m, nLevels int32;
    // set the levels
    nLevels = Abc_NtkLevel( pNtk );
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrStart( nLevels + 2 );
    for  i = 0; i < nNodes; i++  {
        pObj = ppNodes[i];
        assert(Abc_ObjIsCi(pObj), "Abc_ObjIsCi(pObj)");
        Abc_NodeSetTravIdCurrent( pObj );
        // add to the array
        assert(pObj.Level == 0, "pObj.Level == 0");
        pObj.pCopy = (Abc_Obj_t *)Vec_PtrEntry( vNodes, pObj.Level );
        Vec_PtrWriteEntry( vNodes, pObj.Level, pObj );
    }
    // iterate through the levels
    for  i = 0; i <= nLevels; i++  {
        // iterate through the nodes on each level
        for  pObj = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i); (pObj != nil); pObj = pObj.pCopy  {
            // iterate through the fanouts of each node
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, k)), 1); k++  {
                // skip visited nodes
                if Abc_NodeIsTravIdCurrent(pFanout) {
                    continue; }
                // visit the fanins of this fanout
                // inlining of macro Abc_ObjForEachFanin
                for  m = 0; (m < Abc_ObjFaninNum(pFanout)) && (((pFanin) = Abc_ObjFanin(pFanout, m)), 1); m++  {
                    if !Abc_NodeIsTravIdCurrent(pFanin) {
                        break; }
                }
                if m < Abc_ObjFaninNum(pFanout) {
                    continue; }
                // all fanins are already collected

                // mark the node as visited
                Abc_NodeSetTravIdCurrent( pFanout );
                // handle the COs
                if Abc_ObjIsCo(pFanout) {
                    pFanout.Level = nLevels + 1; }
                // add to the array
                pFanout.pCopy = (Abc_Obj_t *)Vec_PtrEntry( vNodes, pFanout.Level );
                Vec_PtrWriteEntry( vNodes, pFanout.Level, pFanout );
                // handle the COs
                if Abc_ObjIsCo(pFanout) {
                    pFanout.Level = 0; }
            }
        }
    }
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsSeq_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // visit the transitive fanin of the node
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_NtkDfsSeq_rec( pFanin, vNodes );
    }
    // add the node after the fanins have been added
    Vec_PtrPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the array of nodes and latches reachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsSeq(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(!Abc_NtkIsNetlist(pNtk), "!Abc_NtkIsNetlist(pNtk)");
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NtkDfsSeq_rec( pObj, vNodes );
    }
    // mark the PIs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NtkDfsSeq_rec( pObj, vNodes );
    }
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsSeqReverse_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // visit the transitive fanin of the node
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
        Abc_NtkDfsSeqReverse_rec( pFanout, vNodes );
    }
    // add the node after the fanins have been added
    Vec_PtrPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the array of nodes and latches reachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsSeqReverse(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(!Abc_NtkIsNetlist(pNtk), "!Abc_NtkIsNetlist(pNtk)");
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NtkDfsSeqReverse_rec( pObj, vNodes );
    }
    // mark the logic feeding into POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NtkDfsSeq_rec( pObj, vNodes );
    }
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Iterative version of the DFS procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfs_iter(vStack *Vec_Ptr_t, pRoot *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pNode, pFanin *Abc_Obj_t;
    var iFanin int32;
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pRoot ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pRoot );
    // skip the CI
    if Abc_ObjIsCi(pRoot) || (Abc_NtkIsStrash(pRoot.pNtk) && Abc_AigNodeIsConst(pRoot)) {
        return; }
    // add the CI
    Vec_PtrClear( vStack );
    Vec_PtrPush( vStack, pRoot );
    Vec_PtrPush( vStack, (void *)0 );
    for  Vec_PtrSize(vStack) > 0  {
        // get the node and its fanin
        iFanin = int32((ABC_PTRINT_T))Vec_PtrPop(vStack);
        pNode  = (Abc_Obj_t *)Vec_PtrPop(vStack);
        assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
        // add it to the array of nodes if we finished
        if iFanin == Abc_ObjFaninNum(pNode) {
            Vec_PtrPush( vNodes, pNode );
            continue;
        }
        // explore the next fanin
        Vec_PtrPush( vStack, pNode );
        Vec_PtrPush( vStack, (void *)(ABC_PTRINT_T)(iFanin+1) );
        // get the fanin
        pFanin = Abc_ObjFanin0Ntk( Abc_ObjFanin(pNode,iFanin) );
        // if this node is already visited, skip
        if Abc_NodeIsTravIdCurrent( pFanin ) {
            continue; }
        // mark the node as visited
        Abc_NodeSetTravIdCurrent( pFanin );
        // skip the CI
        if Abc_ObjIsCi(pFanin) || (Abc_NtkIsStrash(pFanin.pNtk) && Abc_AigNodeIsConst(pFanin)) {
            continue; }
        Vec_PtrPush( vStack, pFanin );
        Vec_PtrPush( vStack, (void *)0 );   
    }
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving CIs and CO.
  However it marks with the current TravId both CIs and COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsIter(pNtk *Abc_Ntk_t, fCollectAll int32) *Vec_Ptr_t {
    var vNodes, vStack *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 1000 );
    vStack = Vec_PtrAlloc( 1000 );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
        Abc_NtkDfs_iter( vStack, Abc_ObjFanin0Ntk(Abc_ObjFanin0(pObj)), vNodes );
    }
    // collect dangling nodes if asked to
    if fCollectAll != 0 { 
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
            if !Abc_NodeIsTravIdCurrent(pObj) {
                Abc_NtkDfs_iter( vStack, pObj, vNodes ); }
        }
    }
    Vec_PtrFree( vStack );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving CIs and CO.
  However it marks with the current TravId both CIs and COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsIterNodes(pNtk *Abc_Ntk_t, vRoots *Vec_Ptr_t) *Vec_Ptr_t {
    var vNodes, vStack *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_NtkIncrementTravId( pNtk );
    vNodes = Vec_PtrAlloc( 1000 );
    vStack = Vec_PtrAlloc( 1000 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        if !Abc_NodeIsTravIdCurrent(Abc_ObjRegular(pObj)) {
            Abc_NtkDfs_iter( vStack, Abc_ObjRegular(pObj), vNodes ); }
    }
    Vec_PtrFree( vStack );
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsHie_rec(pObj *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pObj );
    // visit the transitive fanin of the node
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_NtkDfsHie_rec( pFanin, vNodes );
    }
    // add the node after the fanins have been added
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of all objects.]

  Description [This procedure collects everything from POs to PIs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsHie(pNtk *Abc_Ntk_t, fCollectAll int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NtkDfsHie_rec( pObj, vNodes );
    }
    // collect dangling nodes if asked to
    if fCollectAll != 0 {
        // inlining of macro Abc_NtkForEachObj
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
            if !Abc_NodeIsTravIdCurrent(pObj) {
                Abc_NtkDfs_rec( pObj, vNodes ); }
        }
    }
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if the ordering of nodes is DFS.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIsDfsOrdered(pNtk *Abc_Ntk_t) int32 {
    var pNode, pFanin *Abc_Obj_t;
    var i, k int32;
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // mark the CIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pNode );
    }
    // go through the nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        // check the fanins of the node
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if !Abc_NodeIsTravIdCurrent(pFanin) {
                return 0; }
        }
        // check the choices of the node
        if Abc_NtkIsStrash(pNtk) && Abc_AigNodeIsChoice(pNode) {
            for  pFanin = (Abc_Obj_t *)pNode.pData; (pFanin != nil); pFanin = (Abc_Obj_t *)pFanin.pData  {
                if !Abc_NodeIsTravIdCurrent(pFanin) {
                    return 0; }
            } }
        // mark the node as visited
        Abc_NodeSetTravIdCurrent( pNode );
    }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Create DFS ordering of nets.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsNets_rec(pNet *Abc_Obj_t, vNets *Vec_Ptr_t) {
    var pNext *Abc_Obj_t;
    var pNode *Abc_Obj_t; var i int32;
    assert(Abc_ObjIsNet(pNet), "Abc_ObjIsNet(pNet)");
    if Abc_NodeIsTravIdCurrent( pNet ) {
        return; }
    Abc_NodeSetTravIdCurrent( pNet );
    pNode = Abc_ObjFanin0( pNet );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pNext) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_NtkDfsNets_rec( pNext, vNets );
    }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pNext) = Abc_ObjFanout(pNode, i)), 1); i++  {
        Vec_PtrPush( vNets, pNext );
    }
}
func Abc_NtkDfsNets(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNets *Vec_Ptr_t;
    var pObj *Abc_Obj_t; var i int32;
    vNets = Vec_PtrAlloc( 100 );
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( Abc_ObjFanout0(pObj) );
    }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Vec_PtrPush( vNets, Abc_ObjFanout0(pObj) );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NtkDfsNets_rec( Abc_ObjFanin0(pObj), vNets );
    }
    return vNets;
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving out CIs and CO.
  However it marks with the current TravId both CIs and COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsWithBoxes_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    if Abc_ObjIsBo(pNode) {
        pNode = Abc_ObjFanin0(pNode); }
    if Abc_ObjIsPi(pNode) {
        return; }
    assert(Abc_ObjIsNode( pNode ) || Abc_ObjIsBox( pNode ), "Abc_ObjIsNode( pNode ) || Abc_ObjIsBox( pNode )");
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    Abc_NodeSetTravIdCurrent( pNode );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if Abc_ObjIsBox(pNode) {
            pFanin = Abc_ObjFanin0(pFanin); }
        assert(Abc_ObjIsNet(pFanin), "Abc_ObjIsNet(pFanin)");
        Abc_NtkDfsWithBoxes_rec( Abc_ObjFanin0Ntk(pFanin), vNodes );
    }
    Vec_PtrPush( vNodes, pNode );
}
func Abc_NtkDfsWithBoxes(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_NtkIncrementTravId( pNtk );
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        assert(Abc_ObjIsNet(Abc_ObjFanin0(pObj)), "Abc_ObjIsNet(Abc_ObjFanin0(pObj))");
        Abc_NtkDfsWithBoxes_rec( Abc_ObjFanin0Ntk(Abc_ObjFanin0(pObj)), vNodes );
    }
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkNodeSupport_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // collect the CI
    if Abc_ObjIsCi(pNode) || (Abc_NtkIsStrash(pNode.pNtk) && Abc_ObjFaninNum(pNode) == 0) {
        Vec_PtrPush( vNodes, pNode );
        return;
    }
    assert(Abc_ObjIsNode( pNode ), "Abc_ObjIsNode( pNode )");
    // visit the transitive fanin of the node
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_NtkNodeSupport_rec( Abc_ObjFanin0Ntk(pFanin), vNodes );
    }
}

/**Function*************************************************************

  Synopsis    [Returns the set of CI nodes in the support of the given nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSupport(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var i int32;
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // go through the PO nodes and call for each of them
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NtkNodeSupport_rec( Abc_ObjFanin0(pNode), vNodes );
    }
    // add unused CIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if !Abc_NodeIsTravIdCurrent( pNode ) {
            Vec_PtrPush( vNodes, pNode ); }
    }
    assert(Vec_PtrSize(vNodes) == Abc_NtkCiNum(pNtk), "Vec_PtrSize(vNodes) == Abc_NtkCiNum(pNtk)");
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns the set of CI nodes in the support of the given nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkNodeSupport(pNtk *Abc_Ntk_t, ppNodes **Abc_Obj_t, nNodes int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var i int32;
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // go through the PO nodes and call for each of them
    for  i = 0; i < nNodes; i++  {
        if Abc_ObjIsCo(ppNodes[i]) {
            Abc_NtkNodeSupport_rec( Abc_ObjFanin0(ppNodes[i]), vNodes );
        } else {
            Abc_NtkNodeSupport_rec( ppNodes[i], vNodes ); }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns the set of CI node IDs in the support of the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkNodeSupportInt_rec(pNode *Abc_Obj_t, vNodes *Vec_Int_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // collect the CI
    if Abc_ObjIsCi(pNode) || (Abc_NtkIsStrash(pNode.pNtk) && Abc_ObjFaninNum(pNode) == 0) {
        if Abc_ObjIsCi(pNode) {
            Vec_IntPush( vNodes, pNode.iTemp ); }
        return;
    }
    assert(Abc_ObjIsNode( pNode ), "Abc_ObjIsNode( pNode )");
    // visit the transitive fanin of the node
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_NtkNodeSupportInt_rec( Abc_ObjFanin0Ntk(pFanin), vNodes );
    }
}
func Abc_NtkNodeSupportInt(pNtk *Abc_Ntk_t, iCo int32) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Abc_Obj_t; 
    var i int32;
    if iCo < 0 || iCo >= Abc_NtkCoNum(pNtk) {
        return nil; }
    // save node indices in the CI nodes
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.iTemp = i;
    }
    // collect the indexes of CI nodes in the TFI of the CO node
    Abc_NtkIncrementTravId( pNtk );
    pObj = Abc_NtkCo( pNtk, iCo );
    vNodes = Vec_IntAlloc( 100 );
    Abc_NtkNodeSupportInt_rec( Abc_ObjFanin0(pObj), vNodes );
    Vec_IntSort( vNodes, 0 );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Derives GIA comparing two outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFunctionalIsoGia_rec(pNew *Gia_Man_t, pNode *Abc_Obj_t) int32 {
    var iLit0, iLit1 int32;
    if Abc_NodeIsTravIdCurrent(pNode) || Abc_ObjFaninNum(pNode) == 0 || Abc_ObjIsCi(pNode) {
        return pNode.iTemp; }
    assert(Abc_ObjIsNode( pNode ), "Abc_ObjIsNode( pNode )");
    Abc_NodeSetTravIdCurrent( pNode );
    iLit0 = Abc_NtkFunctionalIsoGia_rec( pNew, Abc_ObjFanin0(pNode) );
    iLit1 = Abc_NtkFunctionalIsoGia_rec( pNew, Abc_ObjFanin1(pNode) );
    iLit0 = Abc_LitNotCond( iLit0, Abc_ObjFaninC0(pNode) );
    iLit1 = Abc_LitNotCond( iLit1, Abc_ObjFaninC1(pNode) );
    return (pNode.iTemp = Gia_ManHashAnd(pNew, iLit0, iLit1));
}
func Abc_NtkFunctionalIsoGia(pNtk *Abc_Ntk_t, iCo1 int32, iCo2 int32, fCommon int32) *Gia_Man_t {
    var pNew *Gia_Man_t = nil;
    var pTemp *Gia_Man_t;
    var vSupp1 *Vec_Int_t  = Abc_NtkNodeSupportInt( pNtk, iCo1 );
    var vSupp2 *Vec_Int_t  = Abc_NtkNodeSupportInt( pNtk, iCo2 );
    if Vec_IntSize(vSupp1) == Vec_IntSize(vSupp2) {
        var pObj *Abc_Obj_t;
        var i, iCi, iLit1, iLit2 int32;
        pNew = Gia_ManStart( 1000 );
        pNew.pName = Abc_UtilStrsav( pNtk.pName );
        pNew.pSpec = Abc_UtilStrsav( pNtk.pSpec );
        Gia_ManHashStart( pNew );
        // put commom together
        if fCommon != 0 {
            var vCommon *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(vSupp1) );
            Vec_IntTwoRemoveCommon( vSupp1, vSupp2, vCommon );
            Vec_IntAppend( vSupp1, vCommon );
            Vec_IntAppend( vSupp2, vCommon );
            Vec_IntFree( vCommon );
            assert(Vec_IntSize(vSupp1) == Vec_IntSize(vSupp2), "Vec_IntSize(vSupp1) == Vec_IntSize(vSupp2)");            
        }
        // primary inputs
        Abc_AigConst1(pNtk).iTemp = 1;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vSupp1)) && (((iCi) = Vec_IntEntry(vSupp1, i)), 1); i++  {
            Abc_NtkCi(pNtk, iCi).iTemp = Gia_ManAppendCi(pNew);
        }
        // create the first cone
        Abc_NtkIncrementTravId( pNtk );
        pObj = Abc_NtkCo( pNtk, iCo1 );
        iLit1 = Abc_NtkFunctionalIsoGia_rec( pNew, Abc_ObjFanin0(pObj) );
        iLit1 = Abc_LitNotCond( iLit1, Abc_ObjFaninC0(pObj) );
        // primary inputs
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vSupp2)) && (((iCi) = Vec_IntEntry(vSupp2, i)), 1); i++  {
            Abc_NtkCi(pNtk, iCi).iTemp = Gia_ManCiLit(pNew, i);
        }
        // create the second cone
        Abc_NtkIncrementTravId( pNtk );
        pObj = Abc_NtkCo( pNtk, iCo2 );
        iLit2 = Abc_NtkFunctionalIsoGia_rec( pNew, Abc_ObjFanin0(pObj) );
        iLit2 = Abc_LitNotCond( iLit2, Abc_ObjFaninC0(pObj) );
        Gia_ManAppendCo( pNew, iLit1 );
        Gia_ManAppendCo( pNew, iLit2 );
        // perform cleanup
        pNew = Gia_ManCleanup( pTemp = pNew );
        Gia_ManStop( pTemp );
    }
    Vec_IntFree( vSupp1 );
    Vec_IntFree( vSupp2 );
    return pNew;
}
func Abc_NtkFunctionalIsoInt(pNtk *Abc_Ntk_t, iCo1 int32, iCo2 int32, fCommon int32) int32 {
    var pGia *Gia_Man_t; var Value int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    if iCo1 < 0 || iCo1 >= Abc_NtkCoNum(pNtk) {
        return 0; }
    if iCo2 < 0 || iCo2 >= Abc_NtkCoNum(pNtk) {
        return 0; }
    pGia = Abc_NtkFunctionalIsoGia( pNtk, iCo1, iCo2, fCommon );
    if pGia == nil {
        return 0; }
    Value = Cec_ManVerifySimple( pGia );
    Gia_ManStop( pGia );
    return int32((Value == 1));
}
func Abc_NtkFunctionalIso(pNtk *Abc_Ntk_t, iCo1 int32, iCo2 int32, fCommon int32) int32 {
    var pNtkNew *Abc_Ntk_t; var Result int32;
    if Abc_NtkIsStrash(pNtk) {
        return Abc_NtkFunctionalIsoInt( pNtk, iCo1, iCo2, fCommon ); }
    pNtkNew = Abc_NtkStrash( pNtk, 0, 0, 0 );
    Result = Abc_NtkFunctionalIsoInt( pNtkNew, iCo1, iCo2, fCommon );
    Abc_NtkDelete( pNtkNew );
    return Result;
}


/**Function*************************************************************

  Synopsis    [Computes support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjSuppSize_rec(pObj *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i, Counter int32;
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 0; }
    Abc_NodeSetTravIdCurrent(pObj);
    if Abc_ObjIsPi(pObj) {
        return 1; }
    assert(Abc_ObjIsNode(pObj) || Abc_ObjIsBox(pObj), "Abc_ObjIsNode(pObj) || Abc_ObjIsBox(pObj)");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Counter += Abc_ObjSuppSize_rec( pFanin );
    }
    return Counter;
}
/**Function*************************************************************

  Synopsis    [Computes support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjSuppSize(pObj *Abc_Obj_t) int32 {
    Abc_NtkIncrementTravId( Abc_ObjNtk(pObj) );
    return Abc_ObjSuppSize_rec( pObj );
}
/**Function*************************************************************

  Synopsis    [Computes support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSuppSizeTest(p *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;
    var i, Counter int32;
    var clk abctime  = Abc_Clock();
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjIsNode(pObj) {
            Counter += (Abc_ObjSuppSize(pObj) <= 16); }
    }
    fmt.Printf("Nodes with small support %d (out of %d)\n", Counter, Abc_NtkNodeNum(p));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Computes the sum total of supports of all outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSupportSum(pNtk *Abc_Ntk_t) {
    var vSupp *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, nTotalSupps int32;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        vSupp = Abc_NtkNodeSupport( pNtk, &pObj, 1 );
        nTotalSupps += Vec_PtrSize( vSupp );
        Vec_PtrFree( vSupp );
    }
    fmt.Printf("Total supports = %d.\n", nTotalSupps);
}


/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigDfs_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // skip the PI
    if Abc_ObjIsCi(pNode) || Abc_AigNodeIsConst(pNode) {
        return; }
    assert(Abc_ObjIsNode( pNode ), "Abc_ObjIsNode( pNode )");
    // visit the transitive fanin of the node
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_AigDfs_rec( pFanin, vNodes );
    }
    // visit the equivalent nodes
    if Abc_AigNodeIsChoice( pNode ) {
        for  pFanin = (Abc_Obj_t *)pNode.pData; (pFanin != nil); pFanin = (Abc_Obj_t *)pFanin.pData  {
            Abc_AigDfs_rec( pFanin, vNodes );
        } }
    // add the node after the fanins have been added
    Vec_PtrPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving out CIs/COs.
  However it marks both CIs and COs with the current TravId.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigDfs(pNtk *Abc_Ntk_t, fCollectAll int32, fCollectCos int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // go through the PO nodes and call for each of them
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_AigDfs_rec( Abc_ObjFanin0(pNode), vNodes );
        Abc_NodeSetTravIdCurrent( pNode );
        if fCollectCos != 0 {
            Vec_PtrPush( vNodes, pNode ); }
    }
    // collect dangling nodes if asked to
    if fCollectAll != 0 {
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else
            if !Abc_NodeIsTravIdCurrent(pNode) {
                Abc_AigDfs_rec( pNode, vNodes ); }
        }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of logic nodes.]

  Description [Collects only the internal nodes, leaving out CIs/COs.
  However it marks both CIs and COs with the current TravId.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigDfsMap(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // collect cones of barbufs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if i < Abc_NtkCoNum(pNtk) - pNtk.nBarBufs {
            continue; }
        Abc_AigDfs_rec( Abc_ObjFanin0(pNode), vNodes );
        Abc_NodeSetTravIdCurrent( pNode );
        // collect latch as a placeholder
        assert(Abc_ObjIsLatch(Abc_ObjFanout0(pNode)), "Abc_ObjIsLatch(Abc_ObjFanout0(pNode))");
        Vec_PtrPush( vNodes, Abc_ObjFanout0(pNode) );
    }
    // collect nodes of real POs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if i >= Abc_NtkCoNum(pNtk) - pNtk.nBarBufs {
            break; }
        Abc_AigDfs_rec( Abc_ObjFanin0(pNode), vNodes );
        assert(Abc_ObjIsCo(pNode), "Abc_ObjIsCo(pNode)");
        Abc_NodeSetTravIdCurrent( pNode );
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Collects nodes in the DFS manner by level.]

  Description [The number of levels should be set!!!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DfsLevelizedTfo_rec(pNode *Abc_Obj_t, vLevels *Vec_Vec_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // skip the terminals
    if Abc_ObjIsCo(pNode) {
        return; }
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // add the node to the structure
    Vec_VecPush( vLevels, pNode.Level, pNode );
    // visit the TFO
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
        Abc_DfsLevelizedTfo_rec( pFanout, vLevels );
    }
}

/**Function*************************************************************

  Synopsis    [Collects nodes in the DFS manner by level.]

  Description [The number of levels should be set!!!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DfsLevelized(pNode *Abc_Obj_t, fTfi int32) *Vec_Vec_t {
    var vLevels *Vec_Vec_t;
    var pFanout *Abc_Obj_t;
    var i int32;
    assert(fTfi == 0, "fTfi == 0");
    assert(!Abc_NtkIsNetlist(pNode.pNtk), "!Abc_NtkIsNetlist(pNode.pNtk)");
    // set the traversal ID
    Abc_NtkIncrementTravId( pNode.pNtk );
    vLevels = Vec_VecAlloc( 100 );
    if Abc_ObjIsNode(pNode) {
        Abc_DfsLevelizedTfo_rec( pNode, vLevels );
    } else {
        assert(Abc_ObjIsCi(pNode), "Abc_ObjIsCi(pNode)");
        Abc_NodeSetTravIdCurrent( pNode );
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
            Abc_DfsLevelizedTfo_rec( pFanout, vLevels );
        }
    }
    return vLevels;
}


/**Function*************************************************************

  Synopsis    [Recursively counts the number of logic levels of one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkLevel_rec(pNode *Abc_Obj_t) int32 {
    var pNext *Abc_Obj_t;
    var i, Level int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    // skip the PI
    if Abc_ObjIsCi(pNode) {
        return pNode.Level; }
    assert(Abc_ObjIsNode( pNode ) || pNode.Type == ABC_OBJ_CONST1, "Abc_ObjIsNode( pNode ) || pNode.Type == ABC_OBJ_CONST1");
    // if this node is already visited, return
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return pNode.Level; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // visit the transitive fanin
    pNode.Level = 0;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pNext) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Level = Abc_NtkLevel_rec( Abc_ObjFanin0Ntk(pNext) );
        if pNode.Level < uint32(Level) {
            pNode.Level = Level; }
    }
    if Abc_ObjFaninNum(pNode) > 0 && !Abc_ObjIsBarBuf(pNode) {
        pNode.Level++; }
    return pNode.Level;
}

/**Function*************************************************************

  Synopsis    [Recursively counts the number of logic levels of one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkLevelReverse_rec(pNode *Abc_Obj_t) int32 {
    var pNext *Abc_Obj_t;
    var i, Level int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    // skip the PI
    if Abc_ObjIsCo(pNode) {
        return pNode.Level; }
    assert(Abc_ObjIsNode( pNode ) || pNode.Type == ABC_OBJ_CONST1, "Abc_ObjIsNode( pNode ) || pNode.Type == ABC_OBJ_CONST1");
    // if this node is already visited, return
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return pNode.Level; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // visit the transitive fanin
    pNode.Level = 0;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pNext) = Abc_ObjFanout(pNode, i)), 1); i++  {
        Level = Abc_NtkLevelReverse_rec( Abc_ObjFanout0Ntk(pNext) );
        if pNode.Level < uint32(Level) {
            pNode.Level = Level; }
    }
    if Abc_ObjFaninNum(pNode) > 0 && !Abc_ObjIsBarBuf(pNode) {
        pNode.Level++; }
    return pNode.Level;
}

/**Function*************************************************************

  Synopsis    [Computes the number of logic levels not counting PIs/POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkLevelize(pNtk *Abc_Ntk_t) *Vec_Vec_t {
    var pObj *Abc_Obj_t;
    var vLevels *Vec_Vec_t;
    var nLevels, i int32;
    nLevels = Abc_NtkLevel( pNtk );
    vLevels = Vec_VecStart( nLevels + 1 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        assert(int32(pObj.Level) <= nLevels, "int32(pObj.Level) <= nLevels");
        Vec_VecPush( vLevels, pObj.Level, pObj );
    }
    }
    return vLevels;
}

/**Function*************************************************************

  Synopsis    [Computes the number of logic levels not counting PIs/POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkLevel(pNtk *Abc_Ntk_t) int32 {
    var pNode *Abc_Obj_t;
    var i, LevelsMax int32;
    // set the CI levels
    if pNtk.pManTime == nil || pNtk.AndGateDelay <= 0 {
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
            pNode.Level = 0;
        }
    } else {
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
            pNode.Level = int32((Abc_MaxFloat(0, Abc_NodeReadArrivalWorst(pNode)) / pNtk.AndGateDelay));
        } }
    // perform the traversal
    LevelsMax = 0;
    Abc_NtkIncrementTravId( pNtk );
    if pNtk.nBarBufs == 0 {
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
            Abc_NtkLevel_rec( pNode );
            if LevelsMax < int32(pNode.Level) {
                LevelsMax = int32(pNode.Level); }
        }
        }
    
    } else {
        // inlining of macro Abc_NtkForEachLiPo
        for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i < pNtk.nBarBufs ? Abc_NtkCoNum(pNtk) - pNtk.nBarBufs + i : i - pNtk.nBarBufs)), 1); i++  {
            var pDriver *Abc_Obj_t  = Abc_ObjFanin0(pNode);
            Abc_NtkLevel_rec( pDriver );
            if LevelsMax < int32(pDriver.Level) {
                LevelsMax = int32(pDriver.Level); }
            // transfer the delay
            if i < pNtk.nBarBufs {
                Abc_ObjFanout0(Abc_ObjFanout0(pNode)).Level = pDriver.Level; }
        }
    }
    return LevelsMax;
}

/**Function*************************************************************

  Synopsis    [Computes the number of logic levels not counting PIs/POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkLevelReverse(pNtk *Abc_Ntk_t) int32 {
    var pNode *Abc_Obj_t;
    var i, LevelsMax int32;
    // set the CO levels to zero
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pNode.Level = 0;
    }
    // perform the traversal
    LevelsMax = 0;
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Abc_NtkLevelReverse_rec( pNode );
        if LevelsMax < int32(pNode.Level) {
            LevelsMax = int32(pNode.Level); }
    }
    }
    return LevelsMax;
}

 
/**Function*************************************************************

  Synopsis    [Recursively detects combinational loops.]

  Description []
                
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIsAcyclic_rec(pNode *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    var pFanin *Abc_Obj_t;
    var fAcyclic, i int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    if Abc_ObjIsCi(pNode) || Abc_ObjIsBox(pNode) || (Abc_NtkIsStrash(pNode.pNtk) && Abc_AigNodeIsConst(pNode)) {
        return 1; }
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // make sure the node is not visited
    assert(!Abc_NodeIsTravIdPrevious(pNode), "!Abc_NodeIsTravIdPrevious(pNode)");
    // check if the node is part of the combinational loop
    if Abc_NodeIsTravIdCurrent(pNode) {
        fmt.Fprintf(os.Stdout, "Network \"%s\" contains combinational loop!\n", Abc_NtkName(pNtk));
        fmt.Fprintf(os.Stdout, "Node \"%s\" is encountered twice on the following path to the COs:\n", Abc_ObjName(pNode));
        return 0;
    }
    // mark this node as a node on the current path
    Abc_NodeSetTravIdCurrent( pNode );
    // visit the transitive fanin
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  { 
        pFanin = Abc_ObjFanin0Ntk(pFanin);
        // make sure there is no mixing of networks
        assert(pFanin.pNtk == pNode.pNtk, "pFanin.pNtk == pNode.pNtk");
        // check if the fanin is visited
        if Abc_NodeIsTravIdPrevious(pFanin) { 
            continue; }
        // traverse the fanin's cone searching for the loop
        if fAcyclic = Abc_NtkIsAcyclic_rec(pFanin) {
            continue; }
        // return as soon as the loop is detected
        fmt.Fprintf(os.Stdout, " %s ->", Abc_ObjName(pFanin));
        return 0;
    }
    // visit choices
    if Abc_NtkIsStrash(pNode.pNtk) && Abc_AigNodeIsChoice(pNode) {
        for  pFanin = (Abc_Obj_t *)pNode.pData; (pFanin != nil); pFanin = (Abc_Obj_t *)pFanin.pData  {
            // check if the fanin is visited
            if Abc_NodeIsTravIdPrevious(pFanin) { 
                continue; }
            // traverse the fanin's cone searching for the loop
            if fAcyclic = Abc_NtkIsAcyclic_rec(pFanin) {
                continue; }
            // return as soon as the loop is detected
            fmt.Fprintf(os.Stdout, " %s", Abc_ObjName(pFanin));
            fmt.Fprintf(os.Stdout, " (choice of %s) -> ", Abc_ObjName(pNode));
            return 0;
        }
    }
    // mark this node as a visited node
    Abc_NodeSetTravIdPrevious( pNode );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Detects combinational loops.]

  Description [This procedure is based on the idea suggested by Donald Chai. 
  As we traverse the network and visit the nodes, we need to distinquish 
  three types of nodes: (1) those that are visited for the first time, 
  (2) those that have been visited in this traversal but are currently not 
  on the traversal path, (3) those that have been visited and are currently 
  on the travesal path. When the node of type (3) is encountered, it means 
  that there is a combinational loop. To mark the three types of nodes, 
  two new values of the traversal IDs are used.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIsAcyclic(pNtk *Abc_Ntk_t) int32 {
    var pNode *Abc_Obj_t;
    var fAcyclic int32;
    var i int32;
    // set the traversal ID for this DFS ordering
    Abc_NtkIncrementTravId( pNtk );   
    Abc_NtkIncrementTravId( pNtk );   
    // pNode->TravId == pNet->nTravIds      means "pNode is on the path"
    // pNode->TravId == pNet->nTravIds - 1  means "pNode is visited but is not on the path"
    // pNode->TravId <  pNet->nTravIds - 1  means "pNode is not visited"
    // traverse the network to detect cycles
    fAcyclic = 1;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pNode = Abc_ObjFanin0Ntk(Abc_ObjFanin0(pNode));
        if Abc_NodeIsTravIdPrevious(pNode) {
            continue; }
        // traverse the output logic cone
        if fAcyclic = Abc_NtkIsAcyclic_rec(pNode) {
            continue; }
        // stop as soon as the first loop is detected
        fmt.Fprintf(os.Stdout, " CO \"%s\"\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        break;
    }
    return fAcyclic;
}
 
/**Function*************************************************************

  Synopsis    [Checks for the loops with boxes.]

  Description []
                
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIsAcyclicWithBoxes_rec(pNode *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    var pFanin *Abc_Obj_t;
    var fAcyclic, i int32;
    assert(!Abc_ObjIsNet(pNode), "!Abc_ObjIsNet(pNode)");
    if Abc_ObjIsPi(pNode) || Abc_ObjIsLatch(pNode) || Abc_ObjIsBlackbox(pNode) {
        return 1; }
    assert(Abc_ObjIsNode(pNode) || Abc_ObjIsBox(pNode), "Abc_ObjIsNode(pNode) || Abc_ObjIsBox(pNode)");
    // make sure the node is not visited
    assert(!Abc_NodeIsTravIdPrevious(pNode), "!Abc_NodeIsTravIdPrevious(pNode)");
    // check if the node is part of the combinational loop
    if Abc_NodeIsTravIdCurrent(pNode) {
        fmt.Fprintf(os.Stdout, "Network \"%s\" contains combinational loop!\n", Abc_NtkName(pNtk));
        if Abc_ObjIsBox(pNode) {
            fmt.Fprintf(os.Stdout, "Box \"%s\" is encountered twice on the following path to the COs:\n", Abc_ObjName(pNode));
        } else {
            fmt.Fprintf(os.Stdout, "Node \"%s\" is encountered twice on the following path to the COs:\n", Abc_ObjName(Abc_ObjFanout0(pNode))); }
        return 0;
    }
    // mark this node as a node on the current path
    Abc_NodeSetTravIdCurrent( pNode );
    // visit the transitive fanin
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  { 
        if Abc_ObjIsBox(pNode) {
            pFanin = Abc_ObjFanin0(pFanin); }
        pFanin = Abc_ObjFanin0Ntk(pFanin);
        if Abc_ObjIsBo(pFanin) {
            pFanin = Abc_ObjFanin0(pFanin); }
        // check if the fanin is visited
        if Abc_ObjIsPi(pFanin) || Abc_ObjIsLatch(pFanin) || Abc_ObjIsBlackbox(pFanin) {
            continue; }
        assert(Abc_ObjIsNode(pFanin) || Abc_ObjIsBox(pFanin), "Abc_ObjIsNode(pFanin) || Abc_ObjIsBox(pFanin)");
        if Abc_NodeIsTravIdPrevious(pFanin) { 
            continue; }
        // traverse the fanin's cone searching for the loop
        if fAcyclic = Abc_NtkIsAcyclicWithBoxes_rec(pFanin) {
            continue; }
        // return as soon as the loop is detected
        fmt.Fprintf(os.Stdout, " %s ->", Abc_ObjName( Abc_ObjIsBox(pFanin) ? pFanin : Abc_ObjFanout0(pFanin) ));
        return 0;
    }
    // mark this node as a visited node
    assert(Abc_ObjIsNode(pNode) || Abc_ObjIsBox(pNode), "Abc_ObjIsNode(pNode) || Abc_ObjIsBox(pNode)");
    Abc_NodeSetTravIdPrevious( pNode );
    return 1;
}
func Abc_NtkIsAcyclicWithBoxes(pNtk *Abc_Ntk_t) int32 {
    var pNode *Abc_Obj_t;
    var fAcyclic int32;
    var i int32;
    // set the traversal ID for this DFS ordering
    Abc_NtkIncrementTravId( pNtk );   
    Abc_NtkIncrementTravId( pNtk );   
    // pNode->TravId == pNet->nTravIds      means "pNode is on the path"
    // pNode->TravId == pNet->nTravIds - 1  means "pNode is visited but is not on the path"
    // pNode->TravId <  pNet->nTravIds - 1  means "pNode is not visited"
    // traverse the network to detect cycles
    fAcyclic = 1;
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pNode) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pNode = Abc_ObjFanin0Ntk(Abc_ObjFanin0(pNode));
        if Abc_ObjIsBo(pNode) {
            pNode = Abc_ObjFanin0(pNode); }
        if Abc_NodeIsTravIdPrevious(pNode) {
            continue; }
        // traverse the output logic cone
        if fAcyclic = Abc_NtkIsAcyclicWithBoxes_rec(pNode) {
            continue; }
        // stop as soon as the first loop is detected
        fmt.Fprintf(os.Stdout, " PO \"%s\"\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        break;
    }
    if fAcyclic != 0 {
        // inlining of macro Abc_NtkForEachLatchInput
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtk, i)) && (((pNode) = Abc_ObjFanin0(Abc_NtkBox(pNtk, i))), 1)) {} else {
            pNode = Abc_ObjFanin0Ntk(Abc_ObjFanin0(pNode));
            if Abc_ObjIsBo(pNode) {
                pNode = Abc_ObjFanin0(pNode); }
            if Abc_NodeIsTravIdPrevious(pNode) {
                continue; }
            // traverse the output logic cone
            if fAcyclic = Abc_NtkIsAcyclicWithBoxes_rec(pNode) {
                continue; }
            // stop as soon as the first loop is detected
            fmt.Fprintf(os.Stdout, " PO \"%s\"\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
            break;
        }
        }
    }
    return fAcyclic;
}



/**Function*************************************************************

  Synopsis    [Analyses choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeSetChoiceLevel_rec(pNode *Abc_Obj_t, fMaximum int32) int32 {
    var pTemp *Abc_Obj_t;
    var Level1, Level2, Level, LevelE int32;
    // skip the visited node
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return int32((ABC_PTRINT_T))pNode.pCopy; }
    Abc_NodeSetTravIdCurrent( pNode );
    // compute levels of the children nodes
    Level1 = Abc_NodeSetChoiceLevel_rec( Abc_ObjFanin0(pNode), fMaximum );
    Level2 = Abc_NodeSetChoiceLevel_rec( Abc_ObjFanin1(pNode), fMaximum );
    Level  = 1 + Abc_MaxInt( Level1, Level2 );
    if pNode.pData {
        LevelE = Abc_NodeSetChoiceLevel_rec( (Abc_Obj_t *)pNode.pData, fMaximum );
        if fMaximum != 0 {
            Level = Abc_MaxInt( Level, LevelE );
        } else {
            Level = Abc_MinInt( Level, LevelE ); }
        // set the level of all equivalent nodes to be the same minimum
        for  pTemp = (Abc_Obj_t *)pNode.pData; pTemp; pTemp = (Abc_Obj_t *)pTemp.pData  {
            pTemp.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)Level;
        }
    }
    pNode.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)Level;
    return Level;
}

/**Function*************************************************************

  Synopsis    [Resets the levels of the nodes in the choice graph.]

  Description [Makes the level of the choice nodes to be equal to the
  maximum of the level of the nodes in the equivalence class. This way
  sorting by level leads to the reverse topological order, which is
  needed for the required time computation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigSetChoiceLevels(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;
    var i, LevelMax, LevelCur int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // set the new travid counter
    Abc_NtkIncrementTravId( pNtk );
    // set levels of the CI and constant
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
        pObj.pCopy = nil;
    }
    pObj = Abc_AigConst1( pNtk );
    Abc_NodeSetTravIdCurrent( pObj );
    pObj.pCopy = nil;
    // set levels of all other nodes
    LevelMax = 0;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        LevelCur = Abc_NodeSetChoiceLevel_rec( Abc_ObjFanin0(pObj), 1 );
        LevelMax = Abc_MaxInt( LevelMax, LevelCur );
    }
    return LevelMax;
}

/**Function*************************************************************

  Synopsis    [Returns nodes by level from the smallest to the largest.]

  Description [Correctly handles the case of choice nodes, by first
  spreading them out across several levels and then collecting.]
               
  SideEffects [What happens with dangling nodes???]

  SeeAlso     []

***********************************************************************/
func Abc_AigGetLevelizedOrder(pNtk *Abc_Ntk_t, fCollectCis int32) *Vec_Ptr_t {
    var vNodes, vLevels *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var ppHead **Abc_Obj_t;
    var LevelMax, i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // set the correct levels
    Abc_NtkCleanCopy( pNtk );
    LevelMax = Abc_AigSetChoiceLevels( pNtk );
    // relink nodes by level
    vLevels = Vec_PtrStart( LevelMax + 1 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        ppHead = ((Abc_Obj_t **)vLevels.pArray) + int32((ABC_PTRINT_T))pNode.pCopy;
        pNode.pCopy = *ppHead;
        *ppHead = pNode;
    }
    }
    // recollect nodes
    vNodes = Vec_PtrStart( Abc_NtkNodeNum(pNtk) );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = !fCollectCis; (i < Vec_PtrSize(vLevels)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vLevels, i)), 1); i++  {
        for  ; (pNode != nil); pNode = pNode.pCopy  {
            Vec_PtrPush( vNodes, pNode );
        }
    }
    Vec_PtrFree( vLevels );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Count the number of nodes in the subgraph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjSugraphSize(pObj *Abc_Obj_t) int32 {
    if Abc_ObjIsCi(pObj) {
        return 0; }
    if Abc_ObjFanoutNum(pObj) > 1 {
        return 0; }
    return 1 + Abc_ObjSugraphSize(Abc_ObjFanin0(pObj)) + 
        Abc_ObjSugraphSize(Abc_ObjFanin1(pObj));
}

/**Function*************************************************************

  Synopsis    [Prints subgraphs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrintSubraphSizes(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
        if Abc_ObjFanoutNum(pObj) > 1 && !Abc_NodeIsExorType(pObj) {
            fmt.Printf("%d(%d) ", 1 + Abc_ObjSugraphSize(Abc_ObjFanin0(pObj)) + 
                            Abc_ObjSugraphSize(Abc_ObjFanin1(pObj)), Abc_ObjFanoutNum(pObj)); }
    }
    fmt.Printf("\n");
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

