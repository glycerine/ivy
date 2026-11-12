package ivy

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [ivyUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Various procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyUtil.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

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

  Synopsis    [Increments the current traversal ID of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManIncrementTravId is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Sets the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManCleanTravId is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Computes truth table of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCollectCut_rec(p *Ivy_Man_t, pNode *Ivy_Obj_t, vNodes *Vec_Int_t) {
    if pNode.fMarkA {
        return; }
    pNode.fMarkA = 1;
    assert(Ivy_ObjIsAnd(pNode) || Ivy_ObjIsExor(pNode), "Ivy_ObjIsAnd(pNode) || Ivy_ObjIsExor(pNode)");
    Ivy_ManCollectCut_rec( p, Ivy_ObjFanin0(pNode), vNodes );
    Ivy_ManCollectCut_rec( p, Ivy_ObjFanin1(pNode), vNodes );
    Vec_IntPush( vNodes, pNode.Id );
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the cut.]

  Description [Does not modify the array of leaves. Uses array vTruth to store 
  temporary truth tables. The returned pointer should be used immediately.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCollectCut(p *Ivy_Man_t, pRoot *Ivy_Obj_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t) {
    var i, Leaf int32;
    // collect and mark the leaves
    Vec_IntClear( vNodes );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((Leaf) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        Vec_IntPush( vNodes, Leaf );
        Ivy_ManObj(p, Leaf).fMarkA = 1;
    }
    // collect and mark the nodes
    Ivy_ManCollectCut_rec( p, pRoot, vNodes );
    // clean the nodes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((Leaf) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Ivy_ManObj(p, Leaf).fMarkA = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjGetTruthStore(ObjNum int32, vTruth *Vec_Int_t) *uint32 {
   return ((unsigned *)Vec_IntArray(vTruth)) + 8 * ObjNum;
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCutTruthOne(p *Ivy_Man_t, pNode *Ivy_Obj_t, vTruth *Vec_Int_t, nWords int32) {
    var pTruth, pTruth0, pTruth1 *uint32;
    var i int32;
    pTruth  = Ivy_ObjGetTruthStore( pNode.TravId, vTruth );
    pTruth0 = Ivy_ObjGetTruthStore( Ivy_ObjFanin0(pNode).TravId, vTruth );
    pTruth1 = Ivy_ObjGetTruthStore( Ivy_ObjFanin1(pNode).TravId, vTruth );
    if Ivy_ObjIsExor(pNode) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = pTruth0[i] ^ pTruth1[i];
        }
    } else if !Ivy_ObjFaninC0(pNode) && !Ivy_ObjFaninC1(pNode) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = pTruth0[i] & pTruth1[i];
        }
    } else if !Ivy_ObjFaninC0(pNode) && Ivy_ObjFaninC1(pNode) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = pTruth0[i] & ~pTruth1[i];
        }
    } else if Ivy_ObjFaninC0(pNode) && !Ivy_ObjFaninC1(pNode) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = ~pTruth0[i] & pTruth1[i];
        }
    } else { // if ( Ivy_ObjFaninC0(pNode) && Ivy_ObjFaninC1(pNode) )
        for  i = 0; i < nWords; i++  {
            pTruth[i] = ~pTruth0[i] & ~pTruth1[i];
        } }
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the cut.]

  Description [Does not modify the array of leaves. Uses array vTruth to store 
  temporary truth tables. The returned pointer should be used immediately.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCutTruth(p *Ivy_Man_t, pRoot *Ivy_Obj_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t, vTruth *Vec_Int_t) *uint32 {
    static unsigned uTruths[8][8] = { // elementary truth tables
        { 0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA },
        { 0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC },
        { 0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0 },
        { 0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00 },
        { 0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000 }, 
        { 0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF }, 
        { 0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF,0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF }, 
        { 0x00000000,0x00000000,0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF,0xFFFFFFFF,0xFFFFFFFF } 
    };
    var i, Leaf int32;
    // collect the cut
    Ivy_ManCollectCut( p, pRoot, vLeaves, vNodes );
    // set the node numbers
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((Leaf) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Ivy_ManObj(p, Leaf).TravId = i;
    }
    // alloc enough memory
    Vec_IntClear( vTruth );
    Vec_IntGrow( vTruth, 8 * Vec_IntSize(vNodes) );
    // set the elementary truth tables
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((Leaf) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        memcpy( Ivy_ObjGetTruthStore(i, vTruth), uTruths[i], 8 * sizeof(unsigned) );
    }
    // compute truths for other nodes
    // inlining of macro Vec_IntForEachEntryStart
    for  i = Vec_IntSize(vLeaves); (i < Vec_IntSize(vNodes)) && (((Leaf) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Ivy_ManCutTruthOne( p, Ivy_ManObj(p, Leaf), vTruth, 8 );
    }
    return Ivy_ObjGetTruthStore( pRoot.TravId, vTruth );
}

/**Function*************************************************************

  Synopsis    [Collect the latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManLatches(p *Ivy_Man_t) *Vec_Int_t {
    var vLatches *Vec_Int_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    vLatches = Vec_IntAlloc( Ivy_ManLatchNum(p) );
    // inlining of macro Ivy_ManForEachLatch
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsLatch(pObj) {} else {
        Vec_IntPush( vLatches, pObj.Id ); }
    }
    return vLatches;
}

/**Function*************************************************************

  Synopsis    [Collect the latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManLevels(p *Ivy_Man_t) int32 {
    var pObj *Ivy_Obj_t;
    var i, LevelMax int32;
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        // inlining of macro IVY_MAX
        
        if (LevelMax) > (int32(Ivy_ObjFanin0(pObj).Level)) {
            LevelMax = (LevelMax);
        } else {
            LevelMax = (int32(Ivy_ObjFanin0(pObj).Level));
        }
    }
    return LevelMax;
}

/**Function*************************************************************

  Synopsis    [Collect the latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManResetLevels_rec(pObj *Ivy_Obj_t) int32 {
    if pObj.Level || Ivy_ObjIsCi(pObj) || Ivy_ObjIsConst1(pObj) {
        return pObj.Level; }
    if Ivy_ObjIsBuf(pObj) {
        return pObj.Level = Ivy_ManResetLevels_rec( Ivy_ObjFanin0(pObj) ); }
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    Ivy_ManResetLevels_rec( Ivy_ObjFanin0(pObj) );
    Ivy_ManResetLevels_rec( Ivy_ObjFanin1(pObj) );
    return pObj.Level = Ivy_ObjLevelNew( pObj );
}

/**Function*************************************************************

  Synopsis    [Collect the latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManResetLevels(p *Ivy_Man_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.Level = 0; }
    }
    // inlining of macro Ivy_ManForEachCo
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsCo(pObj) {} else {
        Ivy_ManResetLevels_rec( Ivy_ObjFanin0(pObj) ); }
    }
}

/**Function*************************************************************

  Synopsis    [References/references the node and returns MFFC size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjRefDeref(p *Ivy_Man_t, pNode *Ivy_Obj_t, fReference int32, fLabel int32) int32 {
    var pNode0, pNode1 *Ivy_Obj_t;
    var Counter int32;
    // label visited nodes
    if fLabel != 0 {
        Ivy_ObjSetTravIdCurrent( p, pNode ); }
    // skip the CI
    if Ivy_ObjIsPi(pNode) {
        return 0; }
    assert(Ivy_ObjIsNode(pNode) || Ivy_ObjIsBuf(pNode) || Ivy_ObjIsLatch(pNode), "Ivy_ObjIsNode(pNode) || Ivy_ObjIsBuf(pNode) || Ivy_ObjIsLatch(pNode)");
    // process the internal node
    pNode0 = Ivy_ObjFanin0(pNode);
    pNode1 = Ivy_ObjFanin1(pNode);
    Counter = Ivy_ObjIsNode(pNode);
    if fReference != 0 {
        if pNode0.nRefs++ == 0 {
            Counter += Ivy_ObjRefDeref( p, pNode0, fReference, fLabel ); }
        if pNode1 && pNode1.nRefs++ == 0 {
            Counter += Ivy_ObjRefDeref( p, pNode1, fReference, fLabel ); }
    
    } else {
        assert(pNode0.nRefs > 0, "pNode0.nRefs > 0");
        assert(pNode1 == nil || pNode1.nRefs > 0, "pNode1 == nil || pNode1.nRefs > 0");
        if --pNode0.nRefs == 0 {
            Counter += Ivy_ObjRefDeref( p, pNode0, fReference, fLabel ); }
        if pNode1 && --pNode1.nRefs == 0 {
            Counter += Ivy_ObjRefDeref( p, pNode1, fReference, fLabel ); }
    }
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Labels MFFC with the current label.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjMffcLabel(p *Ivy_Man_t, pNode *Ivy_Obj_t) int32 {
    var nConeSize1, nConeSize2 int32;
    assert(!Ivy_IsComplement( pNode ), "!Ivy_IsComplement( pNode )");
    assert(Ivy_ObjIsNode( pNode ), "Ivy_ObjIsNode( pNode )");
    nConeSize1 = Ivy_ObjRefDeref( p, pNode, 0, 1 ); // dereference
    nConeSize2 = Ivy_ObjRefDeref( p, pNode, 1, 0 ); // reference
    assert(nConeSize1 == nConeSize2, "nConeSize1 == nConeSize2");
    assert(nConeSize1 > 0, "nConeSize1 > 0");
    return nConeSize1;
}

/**Function*************************************************************

  Synopsis    [Recursively updates fanout levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjUpdateLevel_rec(p *Ivy_Man_t, pObj *Ivy_Obj_t) {
    var pFanout *Ivy_Obj_t;
    var vFanouts *Vec_Ptr_t;
    var i, LevelNew int32;
    assert(p.fFanout, "p.fFanout");
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    vFanouts = Vec_PtrAlloc( 10 );
    // inlining of macro Ivy_ObjForEachFanout
    for  i = 0, Ivy_ObjCollectFanouts(p, pObj, vFanouts);                                  i < Vec_PtrSize(vFanouts) && ((pFanout) = (Ivy_Obj_t *)Vec_PtrEntry(vFanouts,i)); i++  {
        if Ivy_ObjIsCo(pFanout) {
//            assert( (int)Ivy_ObjFanin0(pFanout)->Level <= p->nLevelMax );
            continue;
        }
        LevelNew = Ivy_ObjLevelNew( pFanout );
        if int32(pFanout.Level) == LevelNew {
            continue; }
        pFanout.Level = LevelNew;
        Ivy_ObjUpdateLevel_rec( p, pFanout );
    }
    Vec_PtrFree( vFanouts );
}

/**Function*************************************************************

  Synopsis    [Compute the new required level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjLevelRNew(p *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pFanout *Ivy_Obj_t;
    var vFanouts *Vec_Ptr_t;
    var i int32;
    var Required int32;
    var LevelNew int32 = 1000000;
    assert(p.fFanout && p.vRequired, "p.fFanout && p.vRequired");
    vFanouts = Vec_PtrAlloc( 10 );
    // inlining of macro Ivy_ObjForEachFanout
    for  i = 0, Ivy_ObjCollectFanouts(p, pObj, vFanouts);                                  i < Vec_PtrSize(vFanouts) && ((pFanout) = (Ivy_Obj_t *)Vec_PtrEntry(vFanouts,i)); i++  {
        Required = Vec_IntEntry(p.vRequired, pFanout.Id);
        // inlining of macro IVY_MIN
        
        if (LevelNew) < (Required) {
            LevelNew = (LevelNew);
        } else {
            LevelNew = (Required);
        }
    }
    Vec_PtrFree( vFanouts );
    return LevelNew - 1;
}

/**Function*************************************************************

  Synopsis    [Recursively updates fanout levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjUpdateLevelR_rec(p *Ivy_Man_t, pObj *Ivy_Obj_t, ReqNew int32) {
    var pFanin *Ivy_Obj_t;
    if Ivy_ObjIsConst1(pObj) || Ivy_ObjIsCi(pObj) {
        return; }
    assert(Ivy_ObjIsNode(pObj) || Ivy_ObjIsBuf(pObj), "Ivy_ObjIsNode(pObj) || Ivy_ObjIsBuf(pObj)");
    // process the first fanin
    pFanin = Ivy_ObjFanin0(pObj);
    if Vec_IntEntry(p.vRequired, pFanin.Id) > ReqNew - 1 {
        Vec_IntWriteEntry( p.vRequired, pFanin.Id, ReqNew - 1 );
        Ivy_ObjUpdateLevelR_rec( p, pFanin, ReqNew - 1 );
    }
    if Ivy_ObjIsBuf(pObj) {
        return; }
    // process the second fanin
    pFanin = Ivy_ObjFanin1(pObj);
    if Vec_IntEntry(p.vRequired, pFanin.Id) > ReqNew - 1 {
        Vec_IntWriteEntry( p.vRequired, pFanin.Id, ReqNew - 1 );
        Ivy_ObjUpdateLevelR_rec( p, pFanin, ReqNew - 1 );
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjIsMuxType(pNode *Ivy_Obj_t) int32 {
    var pNode0, pNode1 *Ivy_Obj_t;
    // check that the node is regular
    assert(!Ivy_IsComplement(pNode), "!Ivy_IsComplement(pNode)");
    // if the node is not AND, this is not MUX
    if !Ivy_ObjIsAnd(pNode) {
        return 0; }
    // if the children are not complemented, this is not MUX
    if !Ivy_ObjFaninC0(pNode) || !Ivy_ObjFaninC1(pNode) {
        return 0; }
    // get children
    pNode0 = Ivy_ObjFanin0(pNode);
    pNode1 = Ivy_ObjFanin1(pNode);
    // if the children are not ANDs, this is not MUX
    if !Ivy_ObjIsAnd(pNode0) || !Ivy_ObjIsAnd(pNode1) {
        return 0; }
    // otherwise the node is MUX iff it has a pair of equal grandchildren
    return (Ivy_ObjFaninId0(pNode0) == Ivy_ObjFaninId0(pNode1) && (Ivy_ObjFaninC0(pNode0) ^ Ivy_ObjFaninC0(pNode1))) || 
           (Ivy_ObjFaninId0(pNode0) == Ivy_ObjFaninId1(pNode1) && (Ivy_ObjFaninC0(pNode0) ^ Ivy_ObjFaninC1(pNode1))) ||
           (Ivy_ObjFaninId1(pNode0) == Ivy_ObjFaninId0(pNode1) && (Ivy_ObjFaninC1(pNode0) ^ Ivy_ObjFaninC0(pNode1))) ||
           (Ivy_ObjFaninId1(pNode0) == Ivy_ObjFaninId1(pNode1) && (Ivy_ObjFaninC1(pNode0) ^ Ivy_ObjFaninC1(pNode1)));
}

/**Function*************************************************************

  Synopsis    [Recognizes what nodes are control and data inputs of a MUX.]

  Description [If the node is a MUX, returns the control variable C.
  Assigns nodes T and E to be the then and else variables of the MUX. 
  Node C is never complemented. Nodes T and E can be complemented.
  This function also recognizes EXOR/NEXOR gates as MUXes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjRecognizeMux(pNode *Ivy_Obj_t, ppNodeT **Ivy_Obj_t, ppNodeE **Ivy_Obj_t) *Ivy_Obj_t {
    var pNode0, pNode1 *Ivy_Obj_t;
    assert(!Ivy_IsComplement(pNode), "!Ivy_IsComplement(pNode)");
    assert((Ivy_ObjIsMuxType(pNode) != 0), "Ivy_ObjIsMuxType(pNode)");
    // get children
    pNode0 = Ivy_ObjFanin0(pNode);
    pNode1 = Ivy_ObjFanin1(pNode);
    // find the control variable
//    if ( pNode1->p1 == Fraig_Not(pNode2->p1) )
    if Ivy_ObjFaninId0(pNode0) == Ivy_ObjFaninId0(pNode1) && (Ivy_ObjFaninC0(pNode0) ^ Ivy_ObjFaninC0(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p1) )
        if Ivy_ObjFaninC0(pNode0) { // pNode2->p1 is positive phase of C
            *ppNodeT = Ivy_Not(Ivy_ObjChild1(pNode1));//pNode2->p2);
            *ppNodeE = Ivy_Not(Ivy_ObjChild1(pNode0));//pNode1->p2);
            return Ivy_ObjChild0(pNode1);//pNode2->p1;
        
        } else { // pNode1->p1 is positive phase of C
            *ppNodeT = Ivy_Not(Ivy_ObjChild1(pNode0));//pNode1->p2);
            *ppNodeE = Ivy_Not(Ivy_ObjChild1(pNode1));//pNode2->p2);
            return Ivy_ObjChild0(pNode0);//pNode1->p1;
        }
    
//    else if ( pNode1->p1 == Fraig_Not(pNode2->p2) )
    } else if Ivy_ObjFaninId0(pNode0) == Ivy_ObjFaninId1(pNode1) && (Ivy_ObjFaninC0(pNode0) ^ Ivy_ObjFaninC1(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p1) )
        if Ivy_ObjFaninC0(pNode0) { // pNode2->p2 is positive phase of C
            *ppNodeT = Ivy_Not(Ivy_ObjChild0(pNode1));//pNode2->p1);
            *ppNodeE = Ivy_Not(Ivy_ObjChild1(pNode0));//pNode1->p2);
            return Ivy_ObjChild1(pNode1);//pNode2->p2;
        
        } else { // pNode1->p1 is positive phase of C
            *ppNodeT = Ivy_Not(Ivy_ObjChild1(pNode0));//pNode1->p2);
            *ppNodeE = Ivy_Not(Ivy_ObjChild0(pNode1));//pNode2->p1);
            return Ivy_ObjChild0(pNode0);//pNode1->p1;
        }
    
//    else if ( pNode1->p2 == Fraig_Not(pNode2->p1) )
    } else if Ivy_ObjFaninId1(pNode0) == Ivy_ObjFaninId0(pNode1) && (Ivy_ObjFaninC1(pNode0) ^ Ivy_ObjFaninC0(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p2) )
        if Ivy_ObjFaninC1(pNode0) { // pNode2->p1 is positive phase of C
            *ppNodeT = Ivy_Not(Ivy_ObjChild1(pNode1));//pNode2->p2);
            *ppNodeE = Ivy_Not(Ivy_ObjChild0(pNode0));//pNode1->p1);
            return Ivy_ObjChild0(pNode1);//pNode2->p1;
        
        } else { // pNode1->p2 is positive phase of C
            *ppNodeT = Ivy_Not(Ivy_ObjChild0(pNode0));//pNode1->p1);
            *ppNodeE = Ivy_Not(Ivy_ObjChild1(pNode1));//pNode2->p2);
            return Ivy_ObjChild1(pNode0);//pNode1->p2;
        }
    
//    else if ( pNode1->p2 == Fraig_Not(pNode2->p2) )
    } else if Ivy_ObjFaninId1(pNode0) == Ivy_ObjFaninId1(pNode1) && (Ivy_ObjFaninC1(pNode0) ^ Ivy_ObjFaninC1(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p2) )
        if Ivy_ObjFaninC1(pNode0) { // pNode2->p2 is positive phase of C
            *ppNodeT = Ivy_Not(Ivy_ObjChild0(pNode1));//pNode2->p1);
            *ppNodeE = Ivy_Not(Ivy_ObjChild0(pNode0));//pNode1->p1);
            return Ivy_ObjChild1(pNode1);//pNode2->p2;
        
        } else { // pNode1->p2 is positive phase of C
            *ppNodeT = Ivy_Not(Ivy_ObjChild0(pNode0));//pNode1->p1);
            *ppNodeE = Ivy_Not(Ivy_ObjChild0(pNode1));//pNode2->p1);
            return Ivy_ObjChild1(pNode0);//pNode1->p2;
        }
    }
    assert(false, "0"); // this is not MUX
    return nil;
}

/**Function*************************************************************

  Synopsis    [Returns the real fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjReal is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Prints node in HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjPrintVerbose(p *Ivy_Man_t, pObj *Ivy_Obj_t, fHaig int32) {
    var pTemp *Ivy_Obj_t;
    var fShowFanouts int32;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    fmt.Printf("Node %5d : ", Ivy_ObjId(pObj));
    if Ivy_ObjIsConst1(pObj) {
        fmt.Printf("constant 1");
    } else if Ivy_ObjIsPi(pObj) {
        fmt.Printf("PI");
    } else if Ivy_ObjIsPo(pObj) {
        fmt.Printf("PO");
    } else if Ivy_ObjIsLatch(pObj) {
        fmt.Printf("latch (%d%s)", Ivy_ObjFanin0(pObj).Id, (Ivy_ObjFaninC0(pObj)? "\'" : " "));
    } else if Ivy_ObjIsBuf(pObj) {
        fmt.Printf("buffer (%d%s)", Ivy_ObjFanin0(pObj).Id, (Ivy_ObjFaninC0(pObj)? "\'" : " "));
    } else {
        fmt.Printf("AND( %5d%s, %5d%s )", Ivy_ObjFanin0(pObj).Id, (Ivy_ObjFaninC0(pObj)? "\'" : " "), Ivy_ObjFanin1(pObj).Id, (Ivy_ObjFaninC1(pObj)? "\'" : " ")); }
    fmt.Printf(" (refs = %3d)", Ivy_ObjRefs(pObj));
    if fShowFanouts != 0 {
        var vFanouts *Vec_Ptr_t;
        var pFanout *Ivy_Obj_t;
        var i int32;
        vFanouts = Vec_PtrAlloc( 10 );
        fmt.Printf("\nFanouts:\n");
        // inlining of macro Ivy_ObjForEachFanout
        for  i = 0, Ivy_ObjCollectFanouts(p, pObj, vFanouts);                                  i < Vec_PtrSize(vFanouts) && ((pFanout) = (Ivy_Obj_t *)Vec_PtrEntry(vFanouts,i)); i++  {
            fmt.Printf("    ");
            fmt.Printf("Node %5d : ", Ivy_ObjId(pFanout));
            if Ivy_ObjIsPo(pFanout) {
                fmt.Printf("PO");
            } else if Ivy_ObjIsLatch(pFanout) {
                fmt.Printf("latch (%d%s)", Ivy_ObjFanin0(pFanout).Id, (Ivy_ObjFaninC0(pFanout)? "\'" : " "));
            } else if Ivy_ObjIsBuf(pFanout) {
                fmt.Printf("buffer (%d%s)", Ivy_ObjFanin0(pFanout).Id, (Ivy_ObjFaninC0(pFanout)? "\'" : " "));
            } else {
                fmt.Printf("AND( %5d%s, %5d%s )", Ivy_ObjFanin0(pFanout).Id, (Ivy_ObjFaninC0(pFanout)? "\'" : " "), Ivy_ObjFanin1(pFanout).Id, (Ivy_ObjFaninC1(pFanout)? "\'" : " ")); }
            fmt.Printf("\n");
        }
        Vec_PtrFree( vFanouts );
        return;
    }
    if !(fHaig != 0) {
        if pObj.pEquiv == nil {
            fmt.Printf(" HAIG node not given");
        } else {
            fmt.Printf(" HAIG node = %d%s", Ivy_Regular(pObj.pEquiv).Id, (Ivy_IsComplement(pObj.pEquiv)? "\'" : " ")); }
        return;
    }
    if pObj.pEquiv == nil {
        return; }
    // there are choices
    if Ivy_ObjRefs(pObj) > 0 {
        // print equivalence class
        fmt.Printf("  { %5d ", pObj.Id);
        assert(!Ivy_IsComplement(pObj.pEquiv), "!Ivy_IsComplement(pObj.pEquiv)");
        for  pTemp = pObj.pEquiv; pTemp != pObj; pTemp = Ivy_Regular(pTemp.pEquiv)  {
            fmt.Printf(" %5d%s", pTemp.Id, (Ivy_IsComplement(pTemp.pEquiv)? "\'" : " "));
        }
        fmt.Printf(" }");
        return;
    }
    // this is a secondary node
    for  pTemp = Ivy_Regular(pObj.pEquiv); Ivy_ObjRefs(pTemp) == 0; pTemp = Ivy_Regular(pTemp.pEquiv)  {}
    assert(Ivy_ObjRefs(pTemp) > 0, "Ivy_ObjRefs(pTemp) > 0");
    fmt.Printf("  class of %d", pTemp.Id);
}

/**Function*************************************************************

  Synopsis    [Prints node in HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManPrintVerbose(p *Ivy_Man_t, fHaig int32) {
    var vNodes *Vec_Int_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    fmt.Printf("PIs: ");
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        fmt.Printf(" %d", pObj.Id);
    }
    fmt.Printf("\n");
    fmt.Printf("POs: ");
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        fmt.Printf(" %d", pObj.Id);
    }
    fmt.Printf("\n");
    fmt.Printf("Latches: ");
    // inlining of macro Ivy_ManForEachLatch
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsLatch(pObj) {} else {
        fmt.Printf(" %d=%d%s", pObj.Id, Ivy_ObjFanin0(pObj).Id, (Ivy_ObjFaninC0(pObj)? "\'" : " ")); }
    }
    fmt.Printf("\n");
    vNodes = Ivy_ManDfsSeq( p, nil );
    // inlining of macro Ivy_ManForEachNodeVec
    for  i = 0; i < Vec_IntSize(vNodes) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        Ivy_ObjPrintVerbose( p, pObj, fHaig );
fmt.Printf("\n");
    }
    fmt.Printf("\n");
    Vec_IntFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Performs incremental rewriting of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_CutTruthPrint2(p *Ivy_Man_t, pCut *Ivy_Cut_t, uTruth uint32) int32 {
    var i int32;
    fmt.Printf("Trying cut : {");
    for  i = 0; i < pCut.nSize; i++  {
        fmt.Printf(" %6d(%d)", Ivy_LeafId(pCut.pArray[i]), Ivy_LeafLat(pCut.pArray[i]));
    }
    fmt.Printf(" }   ");
    Extra_PrintBinary( os.Stdout, &uTruth, 16 );  fmt.Printf("\n");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs incremental rewriting of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_CutTruthPrint(p *Ivy_Man_t, pCut *Ivy_Cut_t, uTruth uint32) int32 {
    var vArray *Vec_Ptr_t;
    var pObj, pFanout *Ivy_Obj_t;
    var nLatches int32;
    var nPresent int32;
    var i, k int32;
    var fVerbose int32;

    if fVerbose != 0 {
        fmt.Printf("Trying cut : {"); }
    for  i = 0; i < pCut.nSize; i++  {
        if fVerbose != 0 {
            fmt.Printf(" %6d(%d)", Ivy_LeafId(pCut.pArray[i]), Ivy_LeafLat(pCut.pArray[i])); }
        nLatches += Ivy_LeafLat(pCut.pArray[i]);
    }
    if fVerbose != 0 {
        fmt.Printf(" }   "); }
    if fVerbose != 0 {
        fmt.Printf("Latches = %d. ", nLatches); }

    // check if there are latches on the fanout edges
    vArray = Vec_PtrAlloc( 100 );
    for  i = 0; i < pCut.nSize; i++  {
        pObj = Ivy_ManObj( p, Ivy_LeafId(pCut.pArray[i]) );
        // inlining of macro Ivy_ObjForEachFanout
        for  k = 0, Ivy_ObjCollectFanouts(p, pObj, vArray);                                  k < Vec_PtrSize(vArray) && ((pFanout) = (Ivy_Obj_t *)Vec_PtrEntry(vArray,k)); k++  {
            if Ivy_ObjIsLatch(pFanout) {
                nPresent++;
                break;
            }
        }
    }
    Vec_PtrSize( vArray );
    if fVerbose != 0 {
        fmt.Printf("Present = %d. ", nPresent);
        if nLatches > nPresent {
            fmt.Printf("Clauses = %d. ", 2*(nLatches - nPresent)); }
        fmt.Printf("\n");
    }
    if nLatches > nPresent {
        return 2*(nLatches - nPresent);
    } else {
        return 0;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

