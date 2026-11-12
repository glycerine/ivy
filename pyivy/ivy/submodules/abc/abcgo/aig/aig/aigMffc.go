package aig

/**CFile****************************************************************

  FileName    [aigMffc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Computation of MFFCs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigMffc.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Dereferences the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeDeref_rec(pNode *Aig_Obj_t, LevelMin uint32, pPower *float32, pProbs *float32) int32 {
    var Power0 float32 = 0.0;
    var Power1 float32 = 0.0;
    var pFanin *Aig_Obj_t;
    var Counter int32;
    if pProbs != nil {
        *pPower = 0.0; }
    if Aig_ObjIsCi(pNode) {
        return 0; }
    // consider the first fanin
    pFanin = Aig_ObjFanin0(pNode);
    assert(pFanin.nRefs > 0, "pFanin.nRefs > 0");
    if --pFanin.nRefs == 0 && (!LevelMin || pFanin.Level > LevelMin) {
        Counter += Aig_NodeDeref_rec( pFanin, LevelMin, &Power0, pProbs ); }
    if pProbs != nil {
        *pPower += Power0 + 2.0 * pProbs[pFanin.Id] * (1.0 - pProbs[pFanin.Id]); }
    // skip the buffer
    if Aig_ObjIsBuf(pNode) {
        return Counter; }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    // consider the second fanin
    pFanin = Aig_ObjFanin1(pNode);
    assert(pFanin.nRefs > 0, "pFanin.nRefs > 0");
    if --pFanin.nRefs == 0 && (!LevelMin || pFanin.Level > LevelMin) {
        Counter += Aig_NodeDeref_rec( pFanin, LevelMin, &Power1, pProbs ); }
    if pProbs != nil {
        *pPower += Power1 + 2.0 * pProbs[pFanin.Id] * (1.0 - pProbs[pFanin.Id]); }
    return Counter + 1;
}

/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeRef_rec(pNode *Aig_Obj_t, LevelMin uint32) int32 {
    var pFanin *Aig_Obj_t;
    var Counter int32;
    if Aig_ObjIsCi(pNode) {
        return 0; }
    // consider the first fanin
    pFanin = Aig_ObjFanin0(pNode);
    if pFanin.nRefs++ == 0 && (!LevelMin || pFanin.Level > LevelMin) {
        Counter += Aig_NodeRef_rec( pFanin, LevelMin ); }
    // skip the buffer
    if Aig_ObjIsBuf(pNode) {
        return Counter; }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    // consider the second fanin
    pFanin = Aig_ObjFanin1(pNode);
    if pFanin.nRefs++ == 0 && (!LevelMin || pFanin.Level > LevelMin) {
        Counter += Aig_NodeRef_rec( pFanin, LevelMin ); }
    return Counter + 1;
}

/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeRefLabel_rec(p *Aig_Man_t, pNode *Aig_Obj_t, LevelMin uint32) int32 {
    var pFanin *Aig_Obj_t;
    var Counter int32;
    if Aig_ObjIsCi(pNode) {
        return 0; }
    Aig_ObjSetTravIdCurrent( p, pNode );
    // consider the first fanin
    pFanin = Aig_ObjFanin0(pNode);
    if pFanin.nRefs++ == 0 && (!LevelMin || pFanin.Level > LevelMin) {
        Counter += Aig_NodeRefLabel_rec( p, pFanin, LevelMin ); }
    if Aig_ObjIsBuf(pNode) {
        return Counter; }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    // consider the second fanin
    pFanin = Aig_ObjFanin1(pNode);
    if pFanin.nRefs++ == 0 && (!LevelMin || pFanin.Level > LevelMin) {
        Counter += Aig_NodeRefLabel_rec( p, pFanin, LevelMin ); }
    return Counter + 1;
}

/**Function*************************************************************

  Synopsis    [Collects the internal and boundary nodes in the derefed MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeMffcSupp_rec(p *Aig_Man_t, pNode *Aig_Obj_t, LevelMin uint32, vSupp *Vec_Ptr_t, fTopmost int32, pObjSkip *Aig_Obj_t) {
    // skip visited nodes
    if Aig_ObjIsTravIdCurrent(p, pNode) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pNode);
    // add to the new support nodes
    if !(fTopmost != 0) && pNode != pObjSkip && (Aig_ObjIsCi(pNode) || pNode.nRefs > 0 || pNode.Level <= LevelMin) {
        if vSupp != nil { Vec_PtrPush( vSupp, pNode ); }
        return;
    }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    // recur on the children
    Aig_NodeMffcSupp_rec( p, Aig_ObjFanin0(pNode), LevelMin, vSupp, 0, pObjSkip );
    Aig_NodeMffcSupp_rec( p, Aig_ObjFanin1(pNode), LevelMin, vSupp, 0, pObjSkip );
}

/**Function*************************************************************

  Synopsis    [Collects the support of depth-limited MFFC.]

  Description [Returns the number of internal nodes in the MFFC.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeMffcSupp(p *Aig_Man_t, pNode *Aig_Obj_t, LevelMin int32, vSupp *Vec_Ptr_t) int32 {
    var ConeSize1, ConeSize2 int32;
    if vSupp != nil { Vec_PtrClear( vSupp ); }
    if !Aig_ObjIsNode(pNode) {
        if Aig_ObjIsCi(pNode) && (vSupp != nil) {
            Vec_PtrPush( vSupp, pNode ); }
        return 0;
    }
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    Aig_ManIncrementTravId( p );
    ConeSize1 = Aig_NodeDeref_rec( pNode, LevelMin, nil, nil );
    Aig_NodeMffcSupp_rec( p, pNode, LevelMin, vSupp, 1, nil );
    ConeSize2 = Aig_NodeRef_rec( pNode, LevelMin );
    assert(ConeSize1 == ConeSize2, "ConeSize1 == ConeSize2");
    assert(ConeSize1 > 0, "ConeSize1 > 0");
    return ConeSize1;
}

/**Function*************************************************************

  Synopsis    [Labels the nodes in the MFFC.]

  Description [Returns the number of internal nodes in the MFFC.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeMffcLabel(p *Aig_Man_t, pNode *Aig_Obj_t, pPower *float32) int32 {
    var ConeSize1, ConeSize2 int32;
    assert((pPower != nil) == (p.vProbs != nil), "(pPower != nil) == (p.vProbs != nil)");
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    Aig_ManIncrementTravId( p );
    ConeSize1 = Aig_NodeDeref_rec( pNode, 0, pPower, p.vProbs? (float *)p.vProbs.pArray : nil );
    ConeSize2 = Aig_NodeRefLabel_rec( p, pNode, 0 );
    assert(ConeSize1 == ConeSize2, "ConeSize1 == ConeSize2");
    assert(ConeSize1 > 0, "ConeSize1 > 0");
    return ConeSize1;
}

/**Function*************************************************************

  Synopsis    [Labels the nodes in the MFFC.]

  Description [Returns the number of internal nodes in the MFFC.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeMffcLabelCut(p *Aig_Man_t, pNode *Aig_Obj_t, vLeaves *Vec_Ptr_t) int32 {
    var pObj *Aig_Obj_t;
    var i, ConeSize1, ConeSize2 int32;
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    Aig_ManIncrementTravId( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.nRefs++;
    }
    ConeSize1 = Aig_NodeDeref_rec( pNode, 0, nil, nil );
    ConeSize2 = Aig_NodeRefLabel_rec( p, pNode, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.nRefs--;
    }
    assert(ConeSize1 == ConeSize2, "ConeSize1 == ConeSize2");
    assert(ConeSize1 > 0, "ConeSize1 > 0");
    return ConeSize1;
}

/**Function*************************************************************

  Synopsis    [Expands the cut by adding the most closely related node.]

  Description [Returns 1 if the cut exists.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeMffcExtendCut(p *Aig_Man_t, pNode *Aig_Obj_t, vLeaves *Vec_Ptr_t, vResult *Vec_Ptr_t) int32 {
    var pObj, pLeafBest *Aig_Obj_t;
    var i, LevelMax, ConeSize1, ConeSize2, ConeCur1, ConeCur2, ConeBest int32;
    // dereference the current cut
    LevelMax = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        LevelMax = Abc_MaxInt( LevelMax, int32(pObj.Level) );
    }
    if LevelMax == 0 {
        return 0; }
    // dereference the cut
    ConeSize1 = Aig_NodeDeref_rec( pNode, 0, nil, nil );
    // try expanding each node in the boundary
    ConeBest = ABC_INFINITY;
    pLeafBest = nil;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        if int32(pObj.Level) != LevelMax {
            continue; }
        ConeCur1 = Aig_NodeDeref_rec( pObj, 0, nil, nil );
        if ConeBest > ConeCur1 {
            ConeBest = ConeCur1;
            pLeafBest = pObj;
        }
        ConeCur2 = Aig_NodeRef_rec( pObj, 0 );
        assert(ConeCur1 == ConeCur2, "ConeCur1 == ConeCur2");
    }
    assert(pLeafBest != nil, "pLeafBest != nil");
    assert(Aig_ObjIsNode(pLeafBest), "Aig_ObjIsNode(pLeafBest)");
    // deref the best leaf
    ConeCur1 = Aig_NodeDeref_rec( pLeafBest, 0, nil, nil );
    // collect the cut nodes
    Vec_PtrClear( vResult );
    Aig_ManIncrementTravId( p );
    Aig_NodeMffcSupp_rec( p, pNode, 0, vResult, 1, pLeafBest );
    // ref the nodes
    ConeCur2 = Aig_NodeRef_rec( pLeafBest, 0 );
    assert(ConeCur1 == ConeCur2, "ConeCur1 == ConeCur2");
    // ref the original node
    ConeSize2 = Aig_NodeRef_rec( pNode, 0 );
    assert(ConeSize1 == ConeSize2, "ConeSize1 == ConeSize2");
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

