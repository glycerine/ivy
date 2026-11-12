package ivy

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [ivyDfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [DFS collection procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyDfs.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collects nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManDfs_rec(p *Ivy_Man_t, pObj *Ivy_Obj_t, vNodes *Vec_Int_t) {
    if Ivy_ObjIsMarkA(pObj) {
        return; }
    Ivy_ObjSetMarkA(pObj);
    if Ivy_ObjIsConst1(pObj) || Ivy_ObjIsCi(pObj) {
        if p.pHaig == nil && pObj.pEquiv {
            Ivy_ManDfs_rec( p, Ivy_Regular(pObj.pEquiv), vNodes ); }
        return;
    }
//printf( "visiting node %d\n", pObj->Id );
/*
    if ( pObj->Id == 87 || pObj->Id == 90 )
    {
        int y = 0;
    }
*/
    assert(Ivy_ObjIsBuf(pObj) || Ivy_ObjIsAnd(pObj) || Ivy_ObjIsExor(pObj), "Ivy_ObjIsBuf(pObj) || Ivy_ObjIsAnd(pObj) || Ivy_ObjIsExor(pObj)");
    Ivy_ManDfs_rec( p, Ivy_ObjFanin0(pObj), vNodes );
    if !Ivy_ObjIsBuf(pObj) {
        Ivy_ManDfs_rec( p, Ivy_ObjFanin1(pObj), vNodes ); }
    if p.pHaig == nil && pObj.pEquiv {
        Ivy_ManDfs_rec( p, Ivy_Regular(pObj.pEquiv), vNodes ); }
    Vec_IntPush( vNodes, pObj.Id );

//printf( "adding node %d with fanins %d and %d and equiv %d (refs = %d)\n", 
//       pObj->Id, Ivy_ObjFanin0(pObj)->Id, Ivy_ObjFanin1(pObj)->Id, 
//       pObj->pEquiv? Ivy_Regular(pObj->pEquiv)->Id: -1, Ivy_ObjRefs(pObj) );
}

/**Function*************************************************************

  Synopsis    [Collects AND/EXOR nodes in the DFS order from CIs to COs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManDfs(p *Ivy_Man_t) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    assert(Ivy_ManLatchNum(p) == 0, "Ivy_ManLatchNum(p) == 0");
    // make sure the nodes are not marked
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!pObj.fMarkA && !pObj.fMarkB, "!pObj.fMarkA && !pObj.fMarkB"); }
    }
    // collect the nodes
    vNodes = Vec_IntAlloc( Ivy_ManNodeNum(p) );
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Ivy_ManDfs_rec( p, Ivy_ObjFanin0(pObj), vNodes );
    }
    // unmark the collected nodes
//    Ivy_ManForEachNodeVec( p, vNodes, pObj, i )
//        Ivy_ObjClearMarkA(pObj);
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        Ivy_ObjClearMarkA(pObj); }
    }
    // make sure network does not have dangling nodes
    assert(Vec_IntSize(vNodes) == Ivy_ManNodeNum(p) + Ivy_ManBufNum(p), "Vec_IntSize(vNodes) == Ivy_ManNodeNum(p) + Ivy_ManBufNum(p)");
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Collects AND/EXOR nodes in the DFS order from CIs to COs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManDfsSeq(p *Ivy_Man_t, pvLatches **Vec_Int_t) *Vec_Int_t {
    var vNodes, vLatches *Vec_Int_t;
    var pObj *Ivy_Obj_t;
    var i int32;
//    assert( Ivy_ManLatchNum(p) > 0 );
    // make sure the nodes are not marked
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!pObj.fMarkA && !pObj.fMarkB, "!pObj.fMarkA && !pObj.fMarkB"); }
    }
    // collect the latches
    vLatches = Vec_IntAlloc( Ivy_ManLatchNum(p) );
    // inlining of macro Ivy_ManForEachLatch
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsLatch(pObj) {} else {
        Vec_IntPush( vLatches, pObj.Id ); }
    }
    // collect the nodes
    vNodes = Vec_IntAlloc( Ivy_ManNodeNum(p) );
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Ivy_ManDfs_rec( p, Ivy_ObjFanin0(pObj), vNodes );
    }
    // inlining of macro Ivy_ManForEachNodeVec
    for  i = 0; i < Vec_IntSize(vLatches) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vLatches,i))); i++  {
        Ivy_ManDfs_rec( p, Ivy_ObjFanin0(pObj), vNodes );
    }
    // unmark the collected nodes
//    Ivy_ManForEachNodeVec( p, vNodes, pObj, i )
//        Ivy_ObjClearMarkA(pObj);
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        Ivy_ObjClearMarkA(pObj); }
    }
    // make sure network does not have dangling nodes
//    assert( Vec_IntSize(vNodes) == Ivy_ManNodeNum(p) + Ivy_ManBufNum(p) );

// temporary!!!

    if pvLatches == nil {
        Vec_IntFree( vLatches );
    } else {
        *pvLatches = vLatches; }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Collects nodes in the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCollectCone_rec(pObj *Ivy_Obj_t, vCone *Vec_Ptr_t) {
    if pObj.fMarkA {
        return; }
    if Ivy_ObjIsBuf(pObj) {
        Ivy_ManCollectCone_rec( Ivy_ObjFanin0(pObj), vCone );
        Vec_PtrPush( vCone, pObj );
        return;
    }
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    Ivy_ManCollectCone_rec( Ivy_ObjFanin0(pObj), vCone );
    Ivy_ManCollectCone_rec( Ivy_ObjFanin1(pObj), vCone );
    Vec_PtrPushUnique( vCone, pObj );
}

/**Function*************************************************************

  Synopsis    [Collects nodes in the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCollectCone(pObj *Ivy_Obj_t, vFront *Vec_Ptr_t, vCone *Vec_Ptr_t) {
    var pTemp *Ivy_Obj_t;
    var i int32;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    // mark the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        Ivy_Regular(pTemp).fMarkA = 1;
    }
    assert(pObj.fMarkA == 0, "pObj.fMarkA == 0");
    // collect the cone
    Vec_PtrClear( vCone );
    Ivy_ManCollectCone_rec( pObj, vCone );
    // unmark the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        Ivy_Regular(pTemp).fMarkA = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Returns the nodes by level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManLevelize(p *Ivy_Man_t) *Vec_Vec_t {
    var vNodes *Vec_Vec_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    vNodes = Vec_VecAlloc( 100 );
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!Ivy_ObjIsBuf(pObj), "!Ivy_ObjIsBuf(pObj)");
        if Ivy_ObjIsNode(pObj) {
            Vec_VecPush( vNodes, pObj.Level, pObj ); }
    }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Computes required levels for each node.]

  Description [Assumes topological ordering of the nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManRequiredLevels(p *Ivy_Man_t) *Vec_Int_t {
    var pObj *Ivy_Obj_t;
    var vLevelsR *Vec_Int_t;
    var vNodes *Vec_Vec_t;
    var i, k, Level, LevelMax int32;
    assert(p.vRequired == nil, "p.vRequired == nil");
    // start the required times
    vLevelsR = Vec_IntStart( Ivy_ManObjIdMax(p) + 1 );
    // iterate through the nodes in the reverse order
    vNodes = Ivy_ManLevelize( p );
    // inlining of macro Vec_VecForEachEntryReverseReverse
    // inlining of macro Vec_PtrForEachEntryReverse
    for  i = Vec_VecSize(vNodes) - 1; i >= 0; i--  {                                                   for  k = Vec_PtrSize(Vec_VecEntry(vNodes, i)) - 1; (k >= 0) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vNodes, i), k)), 1); k--  {
        Level = Vec_IntEntry( vLevelsR, pObj.Id ) + 1 + Ivy_ObjIsExor(pObj);
        if Vec_IntEntry( vLevelsR, Ivy_ObjFaninId0(pObj) ) < Level {
            Vec_IntWriteEntry( vLevelsR, Ivy_ObjFaninId0(pObj), Level ); }
        if Vec_IntEntry( vLevelsR, Ivy_ObjFaninId1(pObj) ) < Level {
            Vec_IntWriteEntry( vLevelsR, Ivy_ObjFaninId1(pObj), Level ); }
    }
    }
    Vec_VecFree( vNodes );
    // convert it into the required times
    LevelMax = Ivy_ManLevels( p );
//printf( "max %5d\n",LevelMax );
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        Level = Vec_IntEntry( vLevelsR, pObj.Id );
        Vec_IntWriteEntry( vLevelsR, pObj.Id, LevelMax - Level );
//printf( "%5d : %5d %5d\n", pObj->Id, Level, LevelMax - Level );
    }
    }
    p.vRequired = vLevelsR;
    return vLevelsR;
}

/**Function*************************************************************

  Synopsis    [Recursively detects combinational loops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManIsAcyclic_rec is implemented in the ported core API group.


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
// Ivy_ManIsAcyclic is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Sets the levels of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManSetLevels_rec(pObj *Ivy_Obj_t, fHaig int32) int32 {
    // quit if the node is visited
    if Ivy_ObjIsMarkA(pObj) {
        return pObj.Level; }
    Ivy_ObjSetMarkA(pObj);
    // quit if this is a CI
    if Ivy_ObjIsConst1(pObj) || Ivy_ObjIsCi(pObj) {
        return 0; }
    assert(Ivy_ObjIsBuf(pObj) || Ivy_ObjIsAnd(pObj) || Ivy_ObjIsExor(pObj), "Ivy_ObjIsBuf(pObj) || Ivy_ObjIsAnd(pObj) || Ivy_ObjIsExor(pObj)");
    // get levels of the fanins
    Ivy_ManSetLevels_rec( Ivy_ObjFanin0(pObj), fHaig );
    if !Ivy_ObjIsBuf(pObj) {
        Ivy_ManSetLevels_rec( Ivy_ObjFanin1(pObj), fHaig ); }
    // get level of the node
    if Ivy_ObjIsBuf(pObj) {
        pObj.Level = 1 + Ivy_ObjFanin0(pObj).Level;
    } else if Ivy_ObjIsNode(pObj) {
        pObj.Level = Ivy_ObjLevelNew( pObj );
    } else { assert(false, "0"); }
    // get level of other choices
    if (fHaig != 0) && pObj.pEquiv && Ivy_ObjRefs(pObj) > 0 {
        var pTemp *Ivy_Obj_t;
        var LevelMax uint32  = pObj.Level;
        for  pTemp = pObj.pEquiv; pTemp != pObj; pTemp = Ivy_Regular(pTemp.pEquiv)  {
            Ivy_ManSetLevels_rec( pTemp, fHaig );
            // inlining of macro IVY_MAX
            
            if (LevelMax) > (pTemp.Level) {
                LevelMax = (LevelMax);
            } else {
                LevelMax = (pTemp.Level);
            }
        }
        // get this level
        pObj.Level = LevelMax;
        for  pTemp = pObj.pEquiv; pTemp != pObj; pTemp = Ivy_Regular(pTemp.pEquiv)  {
            pTemp.Level = LevelMax;
        }
    }
    return pObj.Level;
}

/**Function*************************************************************

  Synopsis    [Sets the levels of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManSetLevels(p *Ivy_Man_t, fHaig int32) int32 {
    var pObj *Ivy_Obj_t;
    var i, LevelMax int32;
    // check if CIs have choices
    if fHaig != 0 {
        // inlining of macro Ivy_ManForEachCi
        // inlining of macro Ivy_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsCi(pObj) {} else
            if pObj.pEquiv {
                fmt.Printf("CI %d has a choice, which will not be visualized.\n", pObj.Id); }
        }
    }
    // clean the levels
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.Level = 0; }
    }
    // compute the levels
    LevelMax = 0;
    // inlining of macro Ivy_ManForEachCo
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsCo(pObj) {} else {
        Ivy_ManSetLevels_rec( Ivy_ObjFanin0(pObj), fHaig );
        // inlining of macro IVY_MAX
        
        if (LevelMax) > (int32(Ivy_ObjFanin0(pObj).Level)) {
            LevelMax = (LevelMax);
        } else {
            LevelMax = (int32(Ivy_ObjFanin0(pObj).Level));
        }
    }
    }
    // compute levels of nodes without fanout
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if (Ivy_ObjIsNode(pObj) || Ivy_ObjIsBuf(pObj)) && Ivy_ObjRefs(pObj) == 0 {
            Ivy_ManSetLevels_rec( pObj, fHaig );
            // inlining of macro IVY_MAX
            
            if (LevelMax) > (int32(pObj.Level)) {
                LevelMax = (LevelMax);
            } else {
                LevelMax = (int32(pObj.Level));
            }
        }
    }
    // clean the marks
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        Ivy_ObjClearMarkA(pObj); }
    }
    return LevelMax;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

