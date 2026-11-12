package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigRepr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Handing node representatives.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigRepr.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

  Synopsis    [Starts the array of representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManReprStart(p *Aig_Man_t, nIdMax int32) {
    assert(Aig_ManBufNum(p) == 0, "Aig_ManBufNum(p) == 0");
    assert(p.pReprs == nil, "p.pReprs == nil");
    p.nReprsAlloc = nIdMax;
    // inlining of macro ABC_ALLOC
    p.pReprs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(p.nReprsAlloc)));
    memset( p.pReprs, 0, sizeof(Aig_Obj_t *) * p.nReprsAlloc );
}

/**Function*************************************************************

  Synopsis    [Stop the array of representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManReprStop(p *Aig_Man_t) {
    assert(p.pReprs != nil, "p.pReprs != nil");
    // inlining of macro ABC_FREE
    ((p.pReprs) ? (free((char *) (p.pReprs)), (p.pReprs) = 0) : 0);
    p.nReprsAlloc = 0;    
}

/**Function*************************************************************

  Synopsis    [Set the representative.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCreateRepr(p *Aig_Man_t, pNode1 *Aig_Obj_t, pNode2 *Aig_Obj_t) {
    assert(p.pReprs != nil, "p.pReprs != nil");
    assert(!Aig_IsComplement(pNode1), "!Aig_IsComplement(pNode1)");
    assert(!Aig_IsComplement(pNode2), "!Aig_IsComplement(pNode2)");
    assert(pNode1.Id < p.nReprsAlloc, "pNode1.Id < p.nReprsAlloc");
    assert(pNode2.Id < p.nReprsAlloc, "pNode2.Id < p.nReprsAlloc");
    assert(pNode1.Id < pNode2.Id, "pNode1.Id < pNode2.Id");
    p.pReprs[pNode2.Id] = pNode1;
}

/**Function*************************************************************

  Synopsis    [Set the representative.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjSetRepr_(p *Aig_Man_t, pNode1 *Aig_Obj_t, pNode2 *Aig_Obj_t) {
    assert(p.pReprs != nil, "p.pReprs != nil");
    assert(!Aig_IsComplement(pNode1), "!Aig_IsComplement(pNode1)");
    assert(!Aig_IsComplement(pNode2), "!Aig_IsComplement(pNode2)");
    assert(pNode1.Id < p.nReprsAlloc, "pNode1.Id < p.nReprsAlloc");
    assert(pNode2.Id < p.nReprsAlloc, "pNode2.Id < p.nReprsAlloc");
    if pNode1 == pNode2 {
        return; }
    if pNode1.Id < pNode2.Id {
        p.pReprs[pNode2.Id] = pNode1;
    } else {
        p.pReprs[pNode1.Id] = pNode2; }
}

/**Function*************************************************************

  Synopsis    [Find representative.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjFindRepr(p *Aig_Man_t, pNode *Aig_Obj_t) *Aig_Obj_t {
    assert(p.pReprs != nil, "p.pReprs != nil");
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert(pNode.Id < p.nReprsAlloc, "pNode.Id < p.nReprsAlloc");
//    assert( !p->pReprs[pNode->Id] || p->pReprs[pNode->Id]->Id < pNode->Id );
    return p.pReprs[pNode.Id];
}

/**Function*************************************************************

  Synopsis    [Clears the representative.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjClearRepr(p *Aig_Man_t, pNode *Aig_Obj_t) {
    assert(p.pReprs != nil, "p.pReprs != nil");
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert(pNode.Id < p.nReprsAlloc, "pNode.Id < p.nReprsAlloc");
    p.pReprs[pNode.Id] = nil;
}

/**Function*************************************************************

  Synopsis    [Find representative transitively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjFindReprTransitive(p *Aig_Man_t, pNode *Aig_Obj_t) *Aig_Obj_t {
    var pNext, pRepr *Aig_Obj_t;
    if pRepr = Aig_ObjFindRepr(p, pNode) {
        for  pNext = Aig_ObjFindRepr(p, pRepr)  {
            pRepr = pNext;
        } }
    return pRepr;
}

/**Function*************************************************************

  Synopsis    [Returns representatives of fanin in approapriate polarity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjGetRepr(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pRepr *Aig_Obj_t;
    if pRepr = Aig_ObjFindRepr(p, pObj) {
        return Aig_NotCond( (Aig_Obj_t *)pRepr.pData, pObj.fPhase ^ pRepr.fPhase ); }
    return (Aig_Obj_t *)pObj.pData;
}
func Aig_ObjChild0Repr(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond( Aig_ObjGetRepr(p, Aig_ObjFanin0(pObj)), Aig_ObjFaninC0(pObj) ); }
func Aig_ObjChild1Repr(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond( Aig_ObjGetRepr(p, Aig_ObjFanin1(pObj)), Aig_ObjFaninC1(pObj) ); }

/**Function*************************************************************

  Synopsis    [Duplicates AIG while substituting representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManTransferRepr(pNew *Aig_Man_t, pOld *Aig_Man_t) {
    var pObj, pRepr *Aig_Obj_t;
    var k int32;
    assert(pNew.pReprs != nil, "pNew.pReprs != nil");
    // extend storage to fix pNew
    if pNew.nReprsAlloc < Aig_ManObjNumMax(pNew) {
        var nReprsAllocNew int32  = 2 * Aig_ManObjNumMax(pNew);
        // inlining of macro ABC_REALLOC
        
        if pNew.pReprs {
            pNew.pReprs = ((Aig_Obj_t * *) realloc((char *)(pNew.pReprs), sizeof(Aig_Obj_t *) * (size_t)(nReprsAllocNew)));
        } else {
            pNew.pReprs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(nReprsAllocNew)));
        }
        memset( pNew.pReprs + pNew.nReprsAlloc, 0, sizeof(Aig_Obj_t *) * (nReprsAllocNew-pNew.nReprsAlloc) );
        pNew.nReprsAlloc = nReprsAllocNew;
    }
    // go through the nodes which have representatives
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(pOld.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pOld.vObjs, k)), 1); k++  { if (pObj) == nil {} else
        if pRepr = Aig_ObjFindRepr(pOld, pObj) {
            Aig_ObjSetRepr_( pNew, Aig_Regular((Aig_Obj_t *)pRepr.pData), Aig_Regular((Aig_Obj_t *)pObj.pData) ); }
    } 
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupRepr_rec(pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pRepr *Aig_Obj_t;
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    if pRepr = Aig_ObjFindRepr(p, pObj) {
        Aig_ManDupRepr_rec( pNew, p, pRepr );
        return (Aig_Obj_t *)(pObj.pData = Aig_NotCond( (Aig_Obj_t *)pRepr.pData, pRepr.fPhase ^ pObj.fPhase ));
    }
    Aig_ManDupRepr_rec( pNew, p, Aig_ObjFanin0(pObj) );
    Aig_ManDupRepr_rec( pNew, p, Aig_ObjFanin1(pObj) );
    return (Aig_Obj_t *)(pObj.pData = Aig_And( pNew, Aig_ObjChild0Repr(p, pObj), Aig_ObjChild1Repr(p, pObj) ));
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while substituting representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupRepr(p *Aig_Man_t, fOrdered int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // start the HOP package
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // map the const and primary inputs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi(pNew);
    }
//    Aig_ManForEachCi( p, pObj, i )
//        pObj->pData = Aig_ObjGetRepr( p, pObj );
    // map the internal nodes
    if fOrdered != 0 {
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Repr(p, pObj), Aig_ObjChild1Repr(p, pObj) ); }
        }
    
    } else {
//        Aig_ManForEachObj( p, pObj, i )
//            if ( p->pReprs[i] )
//                printf( "Substituting %d for %d.\n", p->pReprs[i]->Id, pObj->Id );

        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Aig_ManDupRepr_rec( pNew, p, Aig_ObjFanin0(pObj) );
        }
    }
    // transfer the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Repr(p, pObj) );
    }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the new manager
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupRepr: Check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG with representatives without removing registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupReprBasic(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(p.pReprs != nil, "p.pReprs != nil");
    // reconstruct AIG with representatives
    pNew = Aig_ManDupRepr( p, 0 );
    // perfrom sequential cleanup but do not remove registers
    Aig_ManSeqCleanupBasic( pNew );
    // remove pointers to the dead nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if pObj.pData && Aig_ObjIsNone((Aig_Obj_t *)pObj.pData) {
            pObj.pData = nil; }
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Transfer representatives and return the number of critical fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRemapRepr(p *Aig_Man_t) int32 {
    var pObj, pRepr *Aig_Obj_t;
    var i, nFanouts int32;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pRepr = Aig_ObjFindReprTransitive( p, pObj );
        if pRepr == nil {
            continue; }
        assert(pRepr.Id < pObj.Id, "pRepr.Id < pObj.Id");
        Aig_ObjSetRepr_( p, pObj, pRepr );
        nFanouts += (pObj.nRefs > 0);
    }
    }
    return nFanouts;
}

/**Function*************************************************************

  Synopsis    [Transfer representatives and return the number of critical fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCountReprs(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    if p.pReprs == nil {
        return 0; }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        Counter += (p.pReprs[i] != nil); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCheckTfi_rec(p *Aig_Man_t, pNode *Aig_Obj_t, pOld *Aig_Obj_t) int32 {
    // check the trivial cases
    if pNode == nil {
        return 0; }
    if Aig_ObjIsCi(pNode) {
        return 0; }
//    if ( pNode->Id < pOld->Id ) // cannot use because of choices of pNode
//        return 0;
    if pNode == pOld {
        return 1; }
    // skip the visited node
    if Aig_ObjIsTravIdCurrent( p, pNode ) {
        return 0; }
    Aig_ObjSetTravIdCurrent( p, pNode );
    // check the children
    if Aig_ObjCheckTfi_rec( p, Aig_ObjFanin0(pNode), pOld ) != 0 {
        return 1; }
    if Aig_ObjCheckTfi_rec( p, Aig_ObjFanin1(pNode), pOld ) != 0 {
        return 1; }
    // check equivalent nodes
    return Aig_ObjCheckTfi_rec( p, Aig_ObjEquiv(p, pNode), pOld );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCheckTfi(p *Aig_Man_t, pNew *Aig_Obj_t, pOld *Aig_Obj_t) int32 {
    assert(!Aig_IsComplement(pNew), "!Aig_IsComplement(pNew)");
    assert(!Aig_IsComplement(pOld), "!Aig_IsComplement(pOld)");
    Aig_ManIncrementTravId( p );
    return Aig_ObjCheckTfi_rec( p, pNew, pOld );
}

/**Function*************************************************************

  Synopsis    [Iteratively rehashes the AIG.]

  Description [The input AIG is assumed to have representatives assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRehash(p *Aig_Man_t) *Aig_Man_t {
    var pTemp *Aig_Man_t;
    var i, nFanouts int32;
    assert(p.pReprs != nil, "p.pReprs != nil");
    for  i = 0; (nFanouts = Aig_ManRemapRepr( p )); i++  {
//        printf( "Iter = %3d. Fanouts = %6d. Nodes = %7d.\n", i+1, nFanouts, Aig_ManNodeNum(p) );
        p = Aig_ManDupRepr( pTemp = p, 1 );
        Aig_ManReprStart( p, Aig_ManObjNumMax(p) );
        Aig_ManTransferRepr( p, pTemp );
        Aig_ManStop( pTemp );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Marks the nodes that are Creates choices.]

  Description [The input AIG is assumed to have representatives assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManMarkValidChoices(p *Aig_Man_t) {
    var pObj, pRepr *Aig_Obj_t;
    var i int32;
    assert(p.pReprs != nil, "p.pReprs != nil");
    // create equivalent nodes in the manager
    assert(p.pEquivs == nil, "p.pEquivs == nil");
    // inlining of macro ABC_ALLOC
    p.pEquivs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManObjNumMax(p))));
    memset( p.pEquivs, 0, sizeof(Aig_Obj_t *) * Aig_ManObjNumMax(p) );
    // make the choice nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pRepr = Aig_ObjFindRepr( p, pObj );
        if pRepr == nil {
            continue; }
        // skip constant and PI classes
        if !Aig_ObjIsNode(pRepr) {
            Aig_ObjClearRepr( p, pObj );
            continue;
        }
        // skip choices with combinatinal loops
        if Aig_ObjCheckTfi( p, pObj, pRepr ) != 0 {
            Aig_ObjClearRepr( p, pObj );
            continue;
        }
//printf( "Node %d is represented by node %d.\n", pObj->Id, pRepr->Id );
        // add choice to the choice node
        if pObj.nRefs > 0 {
            Aig_ObjClearRepr( p, pObj );
            continue;
        }
        assert(pObj.nRefs == 0, "pObj.nRefs == 0");
        p.pEquivs[pObj.Id] = p.pEquivs[pRepr.Id];
        p.pEquivs[pRepr.Id] = pObj;
    }
    }
}


/**Function*************************************************************

  Synopsis    [Transfers the classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_TransferMappedClasses(pAig *Aig_Man_t, pPart *Aig_Man_t, pMapBack *int32) int32 {
    var pObj *Aig_Obj_t;
    var nClasses, k int32;
    nClasses = 0;
    if pPart.pReprs {
      // inlining of macro Aig_ManForEachObj
      // inlining of macro Vec_PtrForEachEntry
      for  k = 0; (k < Vec_PtrSize(pPart.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPart.vObjs, k)), 1); k++  { if (pObj) == nil {} else {
          if pPart.pReprs[pObj.Id] == nil {
              continue; }
          nClasses++;
          Aig_ObjSetRepr_( pAig,
              Aig_ManObj(pAig, pMapBack[pObj.Id]),
              Aig_ManObj(pAig, pMapBack[pPart.pReprs[pObj.Id].Id]) );
      }
      }
    }
    return nClasses;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

