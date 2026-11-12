package nwk

import "fmt"

/**CFile****************************************************************

  FileName    [nwkDfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [DFS traversals.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkDfs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nwk.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Verifies that the objects are in a topo order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManVerifyTopoOrder(pNtk *Nwk_Man_t) int32 {
    var pObj, pNext *Nwk_Obj_t;
    var i, k, iBox, iTerm1, nTerms int32;
    Nwk_ManIncrementTravId( pNtk );
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        if Nwk_ObjIsNode(pObj) || Nwk_ObjIsCo(pObj) {
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[k]); k++  {
                if !Nwk_ObjIsTravIdCurrent(pNext) {
                    fmt.Printf("Node %d has fanin %d that is not in a topological order.\n", pObj.Id, pNext.Id);
                    return 0;
                }
            }
        
        } else if Nwk_ObjIsCi(pObj) {
            if pNtk.pManTime {
                iBox = Tim_ManBoxForCi( pNtk.pManTime, pObj.PioId );
                if iBox >= 0 { // this is not a true PI
                
                    iTerm1 = Tim_ManBoxInputFirst( pNtk.pManTime, iBox );
                    nTerms = Tim_ManBoxInputNum( pNtk.pManTime, iBox );
                    for  k = 0; k < nTerms; k++  {
                        pNext = Nwk_ManCo( pNtk, iTerm1 + k );
                        if !Nwk_ObjIsTravIdCurrent(pNext) {
                            fmt.Printf("Box %d has input %d that is not in a topological order.\n", iBox, pNext.Id);
                            return 0;
                        }
                    }
                }
            }
        
        } else {
            assert(false, "0"); }
        Nwk_ObjSetTravIdCurrent( pObj );
    }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes the number of logic levels not counting PIs/POs.]

  Description [Assumes that white boxes have unit level.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManLevelBackup(pNtk *Nwk_Man_t) int32 {
    var pManTimeUnit *Tim_Man_t;
    var pObj, pFanin *Nwk_Obj_t;
    var i, k, LevelMax, Level int32;
    assert((Nwk_ManVerifyTopoOrder(pNtk) != 0), "Nwk_ManVerifyTopoOrder(pNtk)");
    // clean the levels
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        Nwk_ObjSetLevel( pObj, 0 ); }
    }
    // perform level computation
    LevelMax = 0;
    if pNtk.pManTime {
        pManTimeUnit = Tim_ManDup( pNtk.pManTime, 1 );
    } else {
        pManTimeUnit = nil;
    }
    if pManTimeUnit != nil {
        Tim_ManIncrementTravId( pManTimeUnit ); }
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        if Nwk_ObjIsCi(pObj) {
            if pManTimeUnit != nil {
                Level = int32(Tim_ManGetCiArrival( pManTimeUnit, pObj.PioId ));
            } else {
                Level = 0;
            }
            Nwk_ObjSetLevel( pObj, Level );
        
        } else if Nwk_ObjIsCo(pObj) {
            Level = Nwk_ObjLevel( Nwk_ObjFanin0(pObj) );
            if pManTimeUnit != nil {
                Tim_ManSetCoArrival( pManTimeUnit, pObj.PioId, float32(Level) ); }
            Nwk_ObjSetLevel( pObj, Level );
            if LevelMax < Nwk_ObjLevel(pObj) {
                LevelMax = Nwk_ObjLevel(pObj); }
        
        } else if Nwk_ObjIsNode(pObj) {
            Level = 0;
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  {
                if Level < Nwk_ObjLevel(pFanin) {
                    Level = Nwk_ObjLevel(pFanin); }
            }
            Nwk_ObjSetLevel( pObj, Level + 1 );
        
        } else {
            assert(false, "0"); }
    }
    }
    // set the old timing manager
    if pManTimeUnit != nil {
        Tim_ManStop( pManTimeUnit ); }
    return LevelMax;
}

/**Function*************************************************************

  Synopsis    [Computes the number of logic levels not counting PIs/POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManLevel_rec(pObj *Nwk_Obj_t) {
    var pManTime *Tim_Man_t  = pObj.pMan.pManTime;
    var pNext *Nwk_Obj_t;
    var i, iBox, iTerm1, nTerms, LevelMax int32;
    if Nwk_ObjIsTravIdCurrent( pObj ) {
        return; }
    Nwk_ObjSetTravIdCurrent( pObj );
    if Nwk_ObjIsCi(pObj) {
        if pManTime != nil {
            iBox = Tim_ManBoxForCi( pManTime, pObj.PioId );
            if iBox >= 0 { // this is not a true PI
            
                iTerm1 = Tim_ManBoxInputFirst( pManTime, iBox );
                nTerms = Tim_ManBoxInputNum( pManTime, iBox );
                for  i = 0; i < nTerms; i++  {
                    pNext = Nwk_ManCo(pObj.pMan, iTerm1 + i);
                    Nwk_ManLevel_rec( pNext );
                    if LevelMax < Nwk_ObjLevel(pNext) {
                        LevelMax = Nwk_ObjLevel(pNext); }
                }
                LevelMax++;
            }
        }
    
    } else if Nwk_ObjIsNode(pObj) || Nwk_ObjIsCo(pObj) {
        // inlining of macro Nwk_ObjForEachFanin
        for  i = 0; (i < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[i]); i++  {
            Nwk_ManLevel_rec( pNext );
            if LevelMax < Nwk_ObjLevel(pNext) {
                LevelMax = Nwk_ObjLevel(pNext); }
        }
        if Nwk_ObjIsNode(pObj) && Nwk_ObjFaninNum(pObj) > 0 {
            LevelMax++; }
    
    } else {
        assert(false, "0"); }
    Nwk_ObjSetLevel( pObj, LevelMax );
}

/**Function*************************************************************

  Synopsis    [Computes the number of logic levels not counting PIs/POs.]

  Description [Does not assume that the objects are in a topo order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManLevel(pNtk *Nwk_Man_t) int32 {
    var pObj *Nwk_Obj_t;
    var i, LevelMax int32;
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        Nwk_ObjSetLevel( pObj, 0 ); }
    }
    Nwk_ManIncrementTravId( pNtk );
    // inlining of macro Nwk_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.vCos)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vCos, i)), 1); i++  {                                        if !Nwk_ObjIsPo(pObj) {} else {
        Nwk_ManLevel_rec( pObj );
        if LevelMax < Nwk_ObjLevel(pObj) {
            LevelMax = Nwk_ObjLevel(pObj); }
    }
    }
    // inlining of macro Nwk_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.vCis)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vCis, i)), 1); i++  {
        Nwk_ManLevel_rec( pObj );
        if LevelMax < Nwk_ObjLevel(pObj) {
            LevelMax = Nwk_ObjLevel(pObj); }
    }
    return LevelMax;
}

/**Function*************************************************************

  Synopsis    [Computes the number of logic levels not counting PIs/POs.]

  Description [Does not assume that the objects are in a topo order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManLevelMax(pNtk *Nwk_Man_t) int32 {
    var pObj *Nwk_Obj_t;
    var i, LevelMax int32;
    // inlining of macro Nwk_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.vCos)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vCos, i)), 1); i++  {                                        if !Nwk_ObjIsPo(pObj) {} else
        if LevelMax < Nwk_ObjLevel(pObj) {
            LevelMax = Nwk_ObjLevel(pObj); }
    }
    return LevelMax;
}

/**Function*************************************************************

  Synopsis    [Returns the array of objects in the AIG manager ordered by level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManLevelize(pNtk *Nwk_Man_t) *Vec_Vec_t {
    var pObj *Nwk_Obj_t;
    var vLevels *Vec_Vec_t;
    var nLevels, i int32;
    assert(Nwk_ManVerifyLevel(pNtk), "Nwk_ManVerifyLevel(pNtk)");
    nLevels = Nwk_ManLevelMax( pNtk );
    vLevels = Vec_VecStart( nLevels + 1 );
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pObj) == nil || !Nwk_ObjIsNode(pObj) {} else {
        assert(Nwk_ObjLevel(pObj) <= nLevels, "Nwk_ObjLevel(pObj) <= nLevels");
        Vec_VecPush( vLevels, Nwk_ObjLevel(pObj), pObj );
    }
    }
    return vLevels;
}



/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDfs_rec(pObj *Nwk_Obj_t, vNodes *Vec_Ptr_t) {
    var pNext *Nwk_Obj_t;
    var i int32;
    if Nwk_ObjIsTravIdCurrent( pObj ) {
        return; }
    Nwk_ObjSetTravIdCurrent( pObj );
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[i]); i++  {
        Nwk_ManDfs_rec( pNext, vNodes );
    }
    Vec_PtrPush( vNodes, pObj );
}
 
/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of all objects except latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDfs(pNtk *Nwk_Man_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Nwk_Obj_t;
    var i int32;
    Nwk_ManIncrementTravId( pNtk );
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        if Nwk_ObjIsCi(pObj) {
            Nwk_ObjSetTravIdCurrent( pObj );
            Vec_PtrPush( vNodes, pObj );
        
        } else if Nwk_ObjIsCo(pObj) {
            Nwk_ManDfs_rec( pObj, vNodes ); }
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
func Nwk_ManDfsNodes_rec(pObj *Nwk_Obj_t, vNodes *Vec_Ptr_t) {
    var pNext *Nwk_Obj_t;
    var i int32;
    if Nwk_ObjIsTravIdCurrent( pObj ) {
        return; }
    Nwk_ObjSetTravIdCurrent( pObj );
    if Nwk_ObjIsCi(pObj) {
        return; }
    assert(Nwk_ObjIsNode(pObj), "Nwk_ObjIsNode(pObj)");
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[i]); i++  {
        Nwk_ManDfsNodes_rec( pNext, vNodes );
    }
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Returns the set of internal nodes rooted in the given nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDfsNodes(pNtk *Nwk_Man_t, ppNodes **Nwk_Obj_t, nNodes int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var i int32;
    // set the traversal ID
    Nwk_ManIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // go through the PO nodes and call for each of them
    for  i = 0; i < nNodes; i++  {
        if Nwk_ObjIsCo(ppNodes[i]) {
            Nwk_ManDfsNodes_rec( Nwk_ObjFanin0(ppNodes[i]), vNodes );
        } else {
            Nwk_ManDfsNodes_rec( ppNodes[i], vNodes ); }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDfsReverse_rec(pObj *Nwk_Obj_t, vNodes *Vec_Ptr_t) {
    var pNext *Nwk_Obj_t;
    var i, iBox, iTerm1, nTerms int32;
    if Nwk_ObjIsTravIdCurrent( pObj ) {
        return; }
    Nwk_ObjSetTravIdCurrent( pObj );
    if Nwk_ObjIsCo(pObj) {
        if pObj.pMan.pManTime {
            iBox = Tim_ManBoxForCo( pObj.pMan.pManTime, pObj.PioId );
            if iBox >= 0 { // this is not a true PO
            
                iTerm1 = Tim_ManBoxOutputFirst( pObj.pMan.pManTime, iBox );
                nTerms = Tim_ManBoxOutputNum( pObj.pMan.pManTime, iBox );
                for  i = 0; i < nTerms; i++  {
                    pNext = Nwk_ManCi(pObj.pMan, iTerm1 + i);
                    Nwk_ManDfsReverse_rec( pNext, vNodes );
                }
            }
        }
    
    } else if Nwk_ObjIsNode(pObj) || Nwk_ObjIsCi(pObj) {
        // inlining of macro Nwk_ObjForEachFanout
        for  i = 0; (i < int32((pObj).nFanouts)) && ((pNext) = (pObj).pFanio[(pObj).nFanins+i]); i++  {
            Nwk_ManDfsReverse_rec( pNext, vNodes );
        }
    
    } else {
        assert(false, "0"); }
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Returns the DFS ordered array of all objects except latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDfsReverse(pNtk *Nwk_Man_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Nwk_Obj_t;
    var i int32;
    Nwk_ManIncrementTravId( pNtk );
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Nwk_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.vCis)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vCis, i)), 1); i++  {                                        if !Nwk_ObjIsPi(pObj) {} else {
        Nwk_ManDfsReverse_rec( pObj, vNodes ); }
    }
    // add nodes without fanins
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pObj) == nil || !Nwk_ObjIsNode(pObj) {} else
        if Nwk_ObjFaninNum(pObj) == 0 && !Nwk_ObjIsTravIdCurrent(pObj) {
            Vec_PtrPush( vNodes, pObj ); }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManSupportNodes_rec(pNode *Nwk_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Nwk_Obj_t;
    var i int32;
    // if this node is already visited, skip
    if Nwk_ObjIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Nwk_ObjSetTravIdCurrent( pNode );
    // collect the CI
    if Nwk_ObjIsCi(pNode) {
        Vec_PtrPush( vNodes, pNode );
        return;
    }
    assert(Nwk_ObjIsNode( pNode ), "Nwk_ObjIsNode( pNode )");
    // visit the transitive fanin of the node
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[i]); i++  {
        Nwk_ManSupportNodes_rec( pFanin, vNodes );
    }
}

/**Function*************************************************************

  Synopsis    [Returns the set of CI nodes in the support of the given nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManSupportNodes(pNtk *Nwk_Man_t, ppNodes **Nwk_Obj_t, nNodes int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var i int32;
    // set the traversal ID
    Nwk_ManIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // go through the PO nodes and call for each of them
    for  i = 0; i < nNodes; i++  {
        if Nwk_ObjIsCo(ppNodes[i]) {
            Nwk_ManSupportNodes_rec( Nwk_ObjFanin0(ppNodes[i]), vNodes );
        } else {
            Nwk_ManSupportNodes_rec( ppNodes[i], vNodes ); }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Computes the sum total of supports of all outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManSupportSum(pNtk *Nwk_Man_t) {
    var vSupp *Vec_Ptr_t;
    var pObj *Nwk_Obj_t;
    var i, nTotalSupps int32;
    // inlining of macro Nwk_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.vCos)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vCos, i)), 1); i++  {
        vSupp = Nwk_ManSupportNodes( pNtk, &pObj, 1 );
        nTotalSupps += Vec_PtrSize( vSupp );
        Vec_PtrFree( vSupp );
    }
    fmt.Printf("Total supports = %d.\n", nTotalSupps);
}


/**Function*************************************************************

  Synopsis    [Dereferences the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjDeref_rec(pNode *Nwk_Obj_t) int32 {
    var pFanin *Nwk_Obj_t;
    var i int32;
    var Counter int32 = 1;
    if Nwk_ObjIsCi(pNode) {
        return 0; }
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[i]); i++  {
        assert(pFanin.nFanouts > 0, "pFanin.nFanouts > 0");
        if --pFanin.nFanouts == 0 {
            Counter += Nwk_ObjDeref_rec( pFanin ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjRef_rec(pNode *Nwk_Obj_t) int32 {
    var pFanin *Nwk_Obj_t;
    var i int32;
    var Counter int32 = 1;
    if Nwk_ObjIsCi(pNode) {
        return 0; }
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[i]); i++  {
        if pFanin.nFanouts++ == 0 {
            Counter += Nwk_ObjRef_rec( pFanin ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collects the internal and boundary nodes in the derefed MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjMffcLabel_rec(pNode *Nwk_Obj_t, fTopmost int32) {
    var pFanin *Nwk_Obj_t;
    var i int32;
    // add to the new support nodes
    if !(fTopmost != 0) && (Nwk_ObjIsCi(pNode) || pNode.nFanouts > 0) {
        return; }
    // skip visited nodes
    if Nwk_ObjIsTravIdCurrent(pNode) {
        return; }
    Nwk_ObjSetTravIdCurrent(pNode);
    // recur on the children
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[i]); i++  {
        Nwk_ObjMffcLabel_rec( pFanin, 0 );
    }
    // collect the internal node
//    printf( "%d ", pNode->Id );
}

/**Function*************************************************************

  Synopsis    [Collects the internal nodes of the MFFC limited by cut.]

  Description []
               
  SideEffects [Increments the trav ID and marks visited nodes.]

  SeeAlso     []

***********************************************************************/
func Nwk_ObjMffcLabel(pNode *Nwk_Obj_t) int32 {
    var Count1, Count2 int32;
    // dereference the node
    Count1 = Nwk_ObjDeref_rec( pNode );
    // collect the nodes inside the MFFC
    Nwk_ManIncrementTravId( pNode.pMan );
    Nwk_ObjMffcLabel_rec( pNode, 1 );
    // reference it back
    Count2 = Nwk_ObjRef_rec( pNode );
    assert(Count1 == Count2, "Count1 == Count2");
    return Count1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

