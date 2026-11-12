package nwk

/**CFile****************************************************************

  FileName    [nwkObj.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [Manipulation of objects.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkObj.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Creates an object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCreateObj(p *Nwk_Man_t, nFanins int32, nFanouts int32) *Nwk_Obj_t {
    var pObj *Nwk_Obj_t;
    pObj = (Nwk_Obj_t *)Aig_MmFlexEntryFetch( p.pMemObjs, sizeof(Nwk_Obj_t) + (nFanins + nFanouts + p.nFanioPlus) * sizeof(Nwk_Obj_t *) );
    memset( pObj, 0, sizeof(Nwk_Obj_t) );
    pObj.pFanio = (Nwk_Obj_t **)((char *)pObj + sizeof(Nwk_Obj_t));
    pObj.Id = Vec_PtrSize( p.vObjs );
    Vec_PtrPush( p.vObjs, pObj );
    pObj.pMan        = p;
    pObj.nFanioAlloc = nFanins + nFanouts + p.nFanioPlus;
    return pObj;
}


/**Function*************************************************************

  Synopsis    [Creates a primary input.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCreateCi(p *Nwk_Man_t, nFanouts int32) *Nwk_Obj_t {
    var pObj *Nwk_Obj_t;
    pObj = Nwk_ManCreateObj( p, 1, nFanouts );
    pObj.PioId = Vec_PtrSize( p.vCis );
    Vec_PtrPush( p.vCis, pObj );
    pObj.Type = NWK_OBJ_CI;
    p.nObjs[NWK_OBJ_CI]++;
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Creates a primary output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCreateCo(p *Nwk_Man_t) *Nwk_Obj_t {
    var pObj *Nwk_Obj_t;
    pObj = Nwk_ManCreateObj( p, 1, 1 );
    pObj.PioId = Vec_PtrSize( p.vCos );
    Vec_PtrPush( p.vCos, pObj );
    pObj.Type = NWK_OBJ_CO;
    p.nObjs[NWK_OBJ_CO]++;
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Creates a latch.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCreateLatch(p *Nwk_Man_t) *Nwk_Obj_t {
    var pObj *Nwk_Obj_t;
    pObj = Nwk_ManCreateObj( p, 1, 1 );
    pObj.Type = NWK_OBJ_LATCH;
    p.nObjs[NWK_OBJ_LATCH]++;
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Creates a node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCreateNode(p *Nwk_Man_t, nFanins int32, nFanouts int32) *Nwk_Obj_t {
    var pObj *Nwk_Obj_t;
    pObj = Nwk_ManCreateObj( p, nFanins, nFanouts );
    pObj.Type = NWK_OBJ_NODE;
    p.nObjs[NWK_OBJ_NODE]++;
    return pObj;
}


/**Function*************************************************************

  Synopsis    [Deletes the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDeleteNode(pObj *Nwk_Obj_t) {
    var vNodes *Vec_Ptr_t  = pObj.pMan.vTemp;
    var pTemp *Nwk_Obj_t;
    var i int32;
    assert(Nwk_ObjFanoutNum(pObj) == 0, "Nwk_ObjFanoutNum(pObj) == 0");
    // delete fanins
    Nwk_ObjCollectFanins( pObj, vNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pTemp) = (Nwk_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Nwk_ObjDeleteFanin( pObj, pTemp );
    }
    // remove from the list of objects
    Vec_PtrWriteEntry( pObj.pMan.vObjs, pObj.Id, nil );
    pObj.pMan.nObjs[pObj.Type]--;
    memset( pObj, 0, sizeof(Nwk_Obj_t) );
    pObj.Id = -1;
}

/**Function*************************************************************

  Synopsis    [Deletes the node and MFFC of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDeleteNode_rec(pObj *Nwk_Obj_t) {
    var vNodes *Vec_Ptr_t;
    var i int32;
    assert(!Nwk_ObjIsCi(pObj), "!Nwk_ObjIsCi(pObj)");
    assert(Nwk_ObjFanoutNum(pObj) == 0, "Nwk_ObjFanoutNum(pObj) == 0");
    vNodes = Vec_PtrAlloc( 100 );
    Nwk_ObjCollectFanins( pObj, vNodes );
    Nwk_ManDeleteNode( pObj );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Nwk_ObjIsNode(pObj) && Nwk_ObjFanoutNum(pObj) == 0 {
            Nwk_ManDeleteNode_rec( pObj ); }
    }
    Vec_PtrFree( vNodes );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

