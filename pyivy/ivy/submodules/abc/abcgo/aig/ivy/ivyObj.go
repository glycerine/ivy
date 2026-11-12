package ivy

/**CFile****************************************************************

  FileName    [ivyObj.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Adding/removing objects.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyObj.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

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

  Synopsis    [Create the new node assuming it does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjCreatePi is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Create the new node assuming it does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjCreatePo is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Create the new node assuming it does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjCreate is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Connect the object to the fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjConnect is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Connect the object to the fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjDisconnect is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Replaces the first fanin of the node by the new fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjPatchFanin0 is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Deletes the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjDelete is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Deletes the MFFC of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjDelete_rec is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Replaces one object by another.]

  Description [Both objects are currently in the manager. The new object
  (pObjNew) should be used instead of the old object (pObjOld). If the 
  new object is complemented or used, the buffer is added.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjReplace(p *Ivy_Man_t, pObjOld *Ivy_Obj_t, pObjNew *Ivy_Obj_t, fDeleteOld int32, fFreeTop int32, fUpdateLevel int32) {
    var nRefsOld int32;//, clk;
    // the object to be replaced cannot be complemented
    assert(!Ivy_IsComplement(pObjOld), "!Ivy_IsComplement(pObjOld)");
    // the object to be replaced cannot be a terminal
    assert(Ivy_ObjIsNone(pObjOld) || !Ivy_ObjIsPi(pObjOld), "Ivy_ObjIsNone(pObjOld) || !Ivy_ObjIsPi(pObjOld)");
    // the object to be used cannot be a PO or assert
    assert(!Ivy_ObjIsBuf(Ivy_Regular(pObjNew)), "!Ivy_ObjIsBuf(Ivy_Regular(pObjNew))");
    // the object cannot be the same
    assert(pObjOld != Ivy_Regular(pObjNew), "pObjOld != Ivy_Regular(pObjNew)");
//printf( "Replacing %d by %d.\n", Ivy_Regular(pObjOld)->Id, Ivy_Regular(pObjNew)->Id );

    // if HAIG is defined, create the choice node
    if p.pHaig {
//        if ( pObjOld->Id == 31 )
//        {
//            Ivy_ManShow( p, 0 );
//            Ivy_ManShow( p->pHaig, 1 );
//        }
        Ivy_ManHaigCreateChoice( p, pObjOld, pObjNew );
    }
    // if the new object is complemented or already used, add the buffer
    if Ivy_IsComplement(pObjNew) || Ivy_ObjIsLatch(pObjNew) || Ivy_ObjRefs(pObjNew) > 0 || Ivy_ObjIsPi(pObjNew) || Ivy_ObjIsConst1(pObjNew) {
        pObjNew = Ivy_ObjCreate( p, Ivy_ObjCreateGhost(p, pObjNew, nil, IVY_BUF, IVY_INIT_NONE) ); }
    assert(!Ivy_IsComplement(pObjNew), "!Ivy_IsComplement(pObjNew)");
    if fUpdateLevel != 0 {
//clk = Abc_Clock();
        // if the new node's arrival time is different, recursively update arrival time of the fanouts
        if p.fFanout && !Ivy_ObjIsBuf(pObjNew) && pObjOld.Level != pObjNew.Level {
            assert(Ivy_ObjIsNode(pObjOld), "Ivy_ObjIsNode(pObjOld)");
            pObjOld.Level = pObjNew.Level;
            Ivy_ObjUpdateLevel_rec( p, pObjOld );
        }
//p->time1 += Abc_Clock() - clk;
        // if the new node's required time has changed, recursively update required time of the fanins
//clk = Abc_Clock();
        if p.vRequired {
            var ReqNew int32  = Vec_IntEntry(p.vRequired, pObjOld.Id);
            if ReqNew < Vec_IntEntry(p.vRequired, pObjNew.Id) {
                Vec_IntWriteEntry( p.vRequired, pObjNew.Id, ReqNew );
                Ivy_ObjUpdateLevelR_rec( p, pObjNew, ReqNew );
            }
        }
//p->time2 += Abc_Clock() - clk;
    }
    // delete the old object
    if fDeleteOld != 0 {
        Ivy_ObjDelete_rec( p, pObjOld, fFreeTop ); }
    // make sure object is not pointing to itself
    assert(Ivy_ObjFanin0(pObjNew) == nil || pObjOld != Ivy_ObjFanin0(pObjNew), "Ivy_ObjFanin0(pObjNew) == nil || pObjOld != Ivy_ObjFanin0(pObjNew)");
    assert(Ivy_ObjFanin1(pObjNew) == nil || pObjOld != Ivy_ObjFanin1(pObjNew), "Ivy_ObjFanin1(pObjNew) == nil || pObjOld != Ivy_ObjFanin1(pObjNew)");
    // make sure the old node has no fanin fanout pointers
    if p.fFanout {
        assert(pObjOld.pFanout != nil, "pObjOld.pFanout != nil");
        assert(pObjNew.pFanout == nil, "pObjNew.pFanout == nil");
        pObjNew.pFanout = pObjOld.pFanout;
    }
    // transfer the old object
    assert(Ivy_ObjRefs(pObjNew) == 0, "Ivy_ObjRefs(pObjNew) == 0");
    nRefsOld = pObjOld.nRefs;  
    Ivy_ObjOverwrite( pObjOld, pObjNew );
    pObjOld.nRefs = nRefsOld;
    // patch the fanout of the fanins 
    if p.fFanout {
        Ivy_ObjPatchFanout( p, Ivy_ObjFanin0(pObjOld), pObjNew, pObjOld );
        if Ivy_ObjFanin1(pObjOld) {
        Ivy_ObjPatchFanout( p, Ivy_ObjFanin1(pObjOld), pObjNew, pObjOld ); }
    }
    // update the hash table
    Ivy_TableUpdate( p, pObjNew, pObjOld.Id );
    // recycle the object that was taken over by pObjOld
    Vec_PtrWriteEntry( p.vObjs, pObjNew.Id, nil );
    Ivy_ManRecycleMemory( p, pObjNew );
    // if the new node is the buffer propagate it
    if p.fFanout && Ivy_ObjIsBuf(pObjOld) {
        Vec_PtrPush( p.vBufs, pObjOld ); }
//    Ivy_ManCheckFanouts( p );
//    printf( "\n" );
/*
    if ( p->pHaig )
    {
        int x;
        Ivy_ManShow( p, 0, NULL );
        Ivy_ManShow( p->pHaig, 1, NULL );
        x = 0;
    }
*/
//    if ( Ivy_ManCheckFanoutNums(p) )
//    {
//        int x = 0;
//    }
}

/**Function*************************************************************

  Synopsis    [Fixes buffer fanins.]

  Description [This situation happens because NodeReplace is a lazy
  procedure, which does not propagate the change to the fanouts but
  instead records the change in the form of a buf/inv node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeFixBufferFanins(p *Ivy_Man_t, pNode *Ivy_Obj_t, fUpdateLevel int32) {
    var pFanReal0 *Ivy_Obj_t;
    var pFanReal1 *Ivy_Obj_t;
    var pResult *Ivy_Obj_t = nil;
    if Ivy_ObjIsPo(pNode) {
        if !Ivy_ObjIsBuf(Ivy_ObjFanin0(pNode)) {
            return; }
        pFanReal0 = Ivy_ObjReal( Ivy_ObjChild0(pNode) );
        Ivy_ObjPatchFanin0( p, pNode, pFanReal0 );
//        Ivy_ManCheckFanouts( p );
        return;
    }
    if !Ivy_ObjIsBuf(Ivy_ObjFanin0(pNode)) && !Ivy_ObjIsBuf(Ivy_ObjFanin1(pNode)) {
        return; }
    // get the real fanins
    pFanReal0 = Ivy_ObjReal( Ivy_ObjChild0(pNode) );
    pFanReal1 = Ivy_ObjReal( Ivy_ObjChild1(pNode) );
    // get the new node
    if Ivy_ObjIsNode(pNode) {
        pResult = Ivy_Oper( p, pFanReal0, pFanReal1, Ivy_ObjType(pNode) );
    } else if Ivy_ObjIsLatch(pNode) {
        pResult = Ivy_Latch( p, pFanReal0, Ivy_ObjInit(pNode) );
    } else { 
        assert(false, "0"); }

//printf( "===== Replacing %d by %d.\n", pNode->Id, pResult->Id );
//Ivy_ObjPrintVerbose( p, pNode, 0 );   printf( "\n" );
//Ivy_ObjPrintVerbose( p, pResult, 0 ); printf( "\n" );

    // perform the replacement
    Ivy_ObjReplace( p, pNode, pResult, 1, 0, fUpdateLevel ); 
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

