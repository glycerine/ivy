package nwk

import "fmt"

/**CFile****************************************************************

  FileName    [nwkFanio.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [Manipulation of fanins/fanouts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkFanio.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Collects fanins of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjCollectFanins(pNode *Nwk_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Nwk_Obj_t;
    var i int32;
    Vec_PtrClear(vNodes);
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[i]); i++  {
        Vec_PtrPush( vNodes, pFanin );
    }
}

/**Function*************************************************************

  Synopsis    [Collects fanouts of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjCollectFanouts(pNode *Nwk_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanout *Nwk_Obj_t;
    var i int32;
    Vec_PtrClear(vNodes);
    // inlining of macro Nwk_ObjForEachFanout
    for  i = 0; (i < int32((pNode).nFanouts)) && ((pFanout) = (pNode).pFanio[(pNode).nFanins+i]); i++  {
        Vec_PtrPush( vNodes, pFanout );
    }
}

/**Function*************************************************************

  Synopsis    [Returns the number of the fanin of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjFindFanin(pObj *Nwk_Obj_t, pFanin *Nwk_Obj_t) int32 {  
    var pTemp *Nwk_Obj_t;
    var i int32;
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pTemp) = (pObj).pFanio[i]); i++  {
        if pTemp == pFanin {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Returns the number of the fanout of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjFindFanout(pObj *Nwk_Obj_t, pFanout *Nwk_Obj_t) int32 {  
    var pTemp *Nwk_Obj_t;
    var i int32;
    // inlining of macro Nwk_ObjForEachFanout
    for  i = 0; (i < int32((pObj).nFanouts)) && ((pTemp) = (pObj).pFanio[(pObj).nFanins+i]); i++  {
        if pTemp == pFanout {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node has to be reallocated.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjReallocIsNeeded(pObj *Nwk_Obj_t) int32 {  
    return pObj.nFanins + pObj.nFanouts == pObj.nFanioAlloc;
}
 
/**Function*************************************************************

  Synopsis    [Reallocates the object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManReallocNode(pObj *Nwk_Obj_t) *Nwk_Obj_t {  
    var pFanioOld **Nwk_Obj_t  = pObj.pFanio;
    assert((Nwk_ObjReallocIsNeeded(pObj) != 0), "Nwk_ObjReallocIsNeeded(pObj)");
    pObj.pFanio = (Nwk_Obj_t **)Aig_MmFlexEntryFetch( pObj.pMan.pMemObjs, 2 * pObj.nFanioAlloc * sizeof(Nwk_Obj_t *) );
    memmove( pObj.pFanio, pFanioOld, pObj.nFanioAlloc * sizeof(Nwk_Obj_t *) );
    pObj.nFanioAlloc *= 2;
    pObj.pMan.nRealloced++;
    return nil;
}

/**Function*************************************************************

  Synopsis    [Creates fanout/fanin relationship between the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjAddFanin(pObj *Nwk_Obj_t, pFanin *Nwk_Obj_t) {
    var i int32;
    assert(pObj.pMan == pFanin.pMan, "pObj.pMan == pFanin.pMan");
    assert(pObj.Id >= 0 && pFanin.Id >= 0, "pObj.Id >= 0 && pFanin.Id >= 0");
    if Nwk_ObjReallocIsNeeded(pObj) != 0 {
        Nwk_ManReallocNode( pObj ); }
    if Nwk_ObjReallocIsNeeded(pFanin) != 0 {
        Nwk_ManReallocNode( pFanin ); }
    for  i = pObj.nFanins + pObj.nFanouts; i > pObj.nFanins; i--  {
        pObj.pFanio[i] = pObj.pFanio[i-1];
    }
    pObj.pFanio[pObj.nFanins++] = pFanin;
    pFanin.pFanio[pFanin.nFanins + pFanin.nFanouts++] = pObj;
    pObj.Level = Abc_MaxInt( pObj.Level, pFanin.Level + Nwk_ObjIsNode(pObj) );
}

/**Function*************************************************************

  Synopsis    [Removes fanout/fanin relationship between the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjDeleteFanin(pObj *Nwk_Obj_t, pFanin *Nwk_Obj_t) {
    var i, k, Limit, fFound int32;
    // remove pFanin from the fanin list of pObj
    Limit = pObj.nFanins + pObj.nFanouts;
    fFound = 0;
    for  k = i = 0; i < Limit; i++  {
        if (fFound != 0) || pObj.pFanio[i] != pFanin {
            pObj.pFanio[k++] = pObj.pFanio[i];
        } else {
            fFound = 1; }
    }
    assert(i == k + 1, "i == k + 1"); // if it fails, likely because of duplicated fanin
    pObj.nFanins--;
    // remove pObj from the fanout list of pFanin
    Limit = pFanin.nFanins + pFanin.nFanouts;
    fFound = 0;
    for  k = i = pFanin.nFanins; i < Limit; i++  {
        if (fFound != 0) || pFanin.pFanio[i] != pObj {
            pFanin.pFanio[k++] = pFanin.pFanio[i];
        } else {
            fFound = 1; }
    }
    assert(i == k + 1, "i == k + 1"); // if it fails, likely because of duplicated fanout
    pFanin.nFanouts--; 
}

/**Function*************************************************************

  Synopsis    [Replaces a fanin of the node.]

  Description [The node is pObj. An old fanin of this node (pFaninOld) has to be
  replaced by a new fanin (pFaninNew). Assumes that the node and the old fanin 
  are not complemented. The new fanin can be complemented. In this case, the
  polarity of the new fanin will change, compared to the polarity of the old fanin.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjPatchFanin(pObj *Nwk_Obj_t, pFaninOld *Nwk_Obj_t, pFaninNew *Nwk_Obj_t) {
    var i, k, iFanin, Limit int32;
    assert(pFaninOld != pFaninNew, "pFaninOld != pFaninNew");
    assert(pObj != pFaninOld, "pObj != pFaninOld");
    assert(pObj != pFaninNew, "pObj != pFaninNew");
    assert(pObj.pMan == pFaninOld.pMan, "pObj.pMan == pFaninOld.pMan");
    assert(pObj.pMan == pFaninNew.pMan, "pObj.pMan == pFaninNew.pMan");
    // update the fanin
    iFanin = Nwk_ObjFindFanin( pObj, pFaninOld );
    if iFanin == -1 {
        fmt.Printf("Nwk_ObjPatchFanin(); Error! Node %d is not among", pFaninOld.Id);
        fmt.Printf(" the fanins of node %d...\n", pObj.Id);
        return;
    }
    pObj.pFanio[iFanin] = pFaninNew;
    // remove pObj from the fanout list of pFaninOld
    Limit = pFaninOld.nFanins + pFaninOld.nFanouts;
    for  k = i = pFaninOld.nFanins; i < Limit; i++  {
        if pFaninOld.pFanio[i] != pObj {
            pFaninOld.pFanio[k++] = pFaninOld.pFanio[i]; }
    }
    pFaninOld.nFanouts--;
    // add pObj to the fanout list of pFaninNew
    if Nwk_ObjReallocIsNeeded(pFaninNew) != 0 {
        Nwk_ManReallocNode( pFaninNew ); }
    pFaninNew.pFanio[pFaninNew.nFanins + pFaninNew.nFanouts++] = pObj;
}


/**Function*************************************************************

  Synopsis    [Transfers fanout from the old node to the new node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjTransferFanout(pNodeFrom *Nwk_Obj_t, pNodeTo *Nwk_Obj_t) {
    var vFanouts *Vec_Ptr_t  = pNodeFrom.pMan.vTemp;
    var pTemp *Nwk_Obj_t;
    var nFanoutsOld, i int32;
    assert(!Nwk_ObjIsCo(pNodeFrom) && !Nwk_ObjIsCo(pNodeTo), "!Nwk_ObjIsCo(pNodeFrom) && !Nwk_ObjIsCo(pNodeTo)");
    assert(pNodeFrom.pMan == pNodeTo.pMan, "pNodeFrom.pMan == pNodeTo.pMan");
    assert(pNodeFrom != pNodeTo, "pNodeFrom != pNodeTo");
    assert(Nwk_ObjFanoutNum(pNodeFrom) > 0, "Nwk_ObjFanoutNum(pNodeFrom) > 0");
    // get the fanouts of the old node
    nFanoutsOld = Nwk_ObjFanoutNum(pNodeTo);
    Nwk_ObjCollectFanouts( pNodeFrom, vFanouts );
    // patch the fanin of each of them
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanouts)) && (((pTemp) = (Nwk_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
        Nwk_ObjPatchFanin( pTemp, pNodeFrom, pNodeTo );
    }
    assert(Nwk_ObjFanoutNum(pNodeFrom) == 0, "Nwk_ObjFanoutNum(pNodeFrom) == 0");
    assert(Nwk_ObjFanoutNum(pNodeTo) == nFanoutsOld + Vec_PtrSize(vFanouts), "Nwk_ObjFanoutNum(pNodeTo) == nFanoutsOld + Vec_PtrSize(vFanouts)");
}

/**Function*************************************************************

  Synopsis    [Replaces the node by a new node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjReplace(pNodeOld *Nwk_Obj_t, pNodeNew *Nwk_Obj_t) {
    assert(pNodeOld.pMan == pNodeNew.pMan, "pNodeOld.pMan == pNodeNew.pMan");
    assert(pNodeOld != pNodeNew, "pNodeOld != pNodeNew");
    assert(Nwk_ObjFanoutNum(pNodeOld) > 0, "Nwk_ObjFanoutNum(pNodeOld) > 0");
    // transfer the fanouts to the old node
    Nwk_ObjTransferFanout( pNodeOld, pNodeNew );
    // remove the old node
    Nwk_ManDeleteNode_rec( pNodeOld );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

