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
func Ivy_ObjCreatePi(p *Ivy_Man_t) *Ivy_Obj_t {
    return Ivy_ObjCreate( p, Ivy_ObjCreateGhost(p, nil, nil, IVY_PI, IVY_INIT_NONE) );
}

/**Function*************************************************************

  Synopsis    [Create the new node assuming it does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjCreatePo(p *Ivy_Man_t, pDriver *Ivy_Obj_t) *Ivy_Obj_t {
    return Ivy_ObjCreate( p, Ivy_ObjCreateGhost(p, pDriver, nil, IVY_PO, IVY_INIT_NONE) );
}

/**Function*************************************************************

  Synopsis    [Create the new node assuming it does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjCreate(p *Ivy_Man_t, pGhost *Ivy_Obj_t) *Ivy_Obj_t {
    var pObj *Ivy_Obj_t;
    assert(!Ivy_IsComplement(pGhost), "!Ivy_IsComplement(pGhost)");
    assert(Ivy_ObjIsGhost(pGhost), "Ivy_ObjIsGhost(pGhost)");
    assert(Ivy_TableLookup(p, pGhost) == nil, "Ivy_TableLookup(p, pGhost) == nil");
    // get memory for the new object
    pObj = Ivy_ManFetchMemory( p );
    assert(Ivy_ObjIsNone(pObj), "Ivy_ObjIsNone(pObj)");
    pObj.Id = Vec_PtrSize(p.vObjs);
    Vec_PtrPush( p.vObjs, pObj );
    // add basic info (fanins, compls, type, init)
    pObj.Type = pGhost.Type;
    pObj.Init = pGhost.Init;
    // add connections
    Ivy_ObjConnect( p, pObj, pGhost.pFanin0, pGhost.pFanin1 );
    // compute level
    if Ivy_ObjIsNode(pObj) {
        pObj.Level = Ivy_ObjLevelNew(pObj);
    } else if Ivy_ObjIsLatch(pObj) {
        pObj.Level = 0;
    } else if Ivy_ObjIsOneFanin(pObj) {
        pObj.Level = Ivy_ObjFanin0(pObj).Level;
    } else if !Ivy_ObjIsPi(pObj) {
        assert(false, "0"); }
    // create phase
    if Ivy_ObjIsNode(pObj) {
        pObj.fPhase = Ivy_ObjFaninPhase(Ivy_ObjChild0(pObj)) & Ivy_ObjFaninPhase(Ivy_ObjChild1(pObj));
    } else if Ivy_ObjIsOneFanin(pObj) {
        pObj.fPhase = Ivy_ObjFaninPhase(Ivy_ObjChild0(pObj)); }
    // set the fail TFO flag
    if Ivy_ObjIsNode(pObj) {
        pObj.fFailTfo = Ivy_ObjFanin0(pObj).fFailTfo | Ivy_ObjFanin1(pObj).fFailTfo; }
    // mark the fanins in a special way if the node is EXOR
    if Ivy_ObjIsExor(pObj) {
        Ivy_ObjFanin0(pObj).fExFan = 1;
        Ivy_ObjFanin1(pObj).fExFan = 1;
    }
    // add PIs/POs to the arrays
    if Ivy_ObjIsPi(pObj) {
        Vec_PtrPush( p.vPis, pObj );
    } else if Ivy_ObjIsPo(pObj) {
        Vec_PtrPush( p.vPos, pObj ); }
//    else if ( Ivy_ObjIsBuf(pObj) )
//        Vec_PtrPush( p->vBufs, pObj );
    if p.vRequired && Vec_IntSize(p.vRequired) <= pObj.Id {
        Vec_IntFillExtra( p.vRequired, 2 * Vec_IntSize(p.vRequired), 1000000 ); }
    // update node counters of the manager
    p.nObjs[Ivy_ObjType(pObj)]++;
    p.nCreated++;

//    printf( "Adding %sAIG node: ", p->pHaig==NULL? "H":" " );
//    Ivy_ObjPrintVerbose( p, pObj, p->pHaig==NULL );
//    printf( "\n" );

    // if HAIG is defined, create a corresponding node
    if p.pHaig {
        Ivy_ManHaigCreateObj( p, pObj ); }
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Connect the object to the fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjConnect(p *Ivy_Man_t, pObj *Ivy_Obj_t, pFan0 *Ivy_Obj_t, pFan1 *Ivy_Obj_t) {
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(Ivy_ObjIsPi(pObj) || Ivy_ObjIsOneFanin(pObj) || pFan1 != nil, "Ivy_ObjIsPi(pObj) || Ivy_ObjIsOneFanin(pObj) || pFan1 != nil");
    // add the first fanin
    pObj.pFanin0 = pFan0;
    pObj.pFanin1 = pFan1;
    // increment references of the fanins and add their fanouts
    if Ivy_ObjFanin0(pObj) != nil {
        Ivy_ObjRefsInc( Ivy_ObjFanin0(pObj) );
        if p.fFanout {
            Ivy_ObjAddFanout( p, Ivy_ObjFanin0(pObj), pObj ); }
    }
    if Ivy_ObjFanin1(pObj) != nil {
        Ivy_ObjRefsInc( Ivy_ObjFanin1(pObj) );
        if p.fFanout {
            Ivy_ObjAddFanout( p, Ivy_ObjFanin1(pObj), pObj ); }
    }
    // add the node to the structural hash table
    Ivy_TableInsert( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Connect the object to the fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjDisconnect(p *Ivy_Man_t, pObj *Ivy_Obj_t) {
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(Ivy_ObjIsPi(pObj) || Ivy_ObjIsOneFanin(pObj) || Ivy_ObjFanin1(pObj) != nil, "Ivy_ObjIsPi(pObj) || Ivy_ObjIsOneFanin(pObj) || Ivy_ObjFanin1(pObj) != nil");
    // remove connections
    if pObj.pFanin0 != nil {
        Ivy_ObjRefsDec(Ivy_ObjFanin0(pObj));
        if p.fFanout {
            Ivy_ObjDeleteFanout( p, Ivy_ObjFanin0(pObj), pObj ); }
    }
    if pObj.pFanin1 != nil {
        Ivy_ObjRefsDec(Ivy_ObjFanin1(pObj));
        if p.fFanout {
            Ivy_ObjDeleteFanout( p, Ivy_ObjFanin1(pObj), pObj ); }
    }
    assert(pObj.pNextFan0 == nil, "pObj.pNextFan0 == nil");
    assert(pObj.pNextFan1 == nil, "pObj.pNextFan1 == nil");
    assert(pObj.pPrevFan0 == nil, "pObj.pPrevFan0 == nil");
    assert(pObj.pPrevFan1 == nil, "pObj.pPrevFan1 == nil");
    // remove the node from the structural hash table
    Ivy_TableDelete( p, pObj );
    // add the first fanin
    pObj.pFanin0 = nil;
    pObj.pFanin1 = nil;
}

/**Function*************************************************************

  Synopsis    [Replaces the first fanin of the node by the new fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjPatchFanin0(p *Ivy_Man_t, pObj *Ivy_Obj_t, pFaninNew *Ivy_Obj_t) {
    var pFaninOld *Ivy_Obj_t;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    pFaninOld = Ivy_ObjFanin0(pObj);
    // decrement ref and remove fanout
    Ivy_ObjRefsDec( pFaninOld );
    if p.fFanout {
        Ivy_ObjDeleteFanout( p, pFaninOld, pObj ); }
    // update the fanin
    pObj.pFanin0 = pFaninNew;
    // increment ref and add fanout
    Ivy_ObjRefsInc( Ivy_Regular(pFaninNew) );
    if p.fFanout {
        Ivy_ObjAddFanout( p, Ivy_Regular(pFaninNew), pObj ); }
    // get rid of old fanin
    if !Ivy_ObjIsPi(pFaninOld) && !Ivy_ObjIsConst1(pFaninOld) && Ivy_ObjRefs(pFaninOld) == 0 {
        Ivy_ObjDelete_rec( p, pFaninOld, 1 ); }
}

/**Function*************************************************************

  Synopsis    [Deletes the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjDelete(p *Ivy_Man_t, pObj *Ivy_Obj_t, fFreeTop int32) {
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(Ivy_ObjRefs(pObj) == 0 || !(fFreeTop != 0), "Ivy_ObjRefs(pObj) == 0 || !fFreeTop");
    // update node counters of the manager
    p.nObjs[pObj.Type]--;
    p.nDeleted++;
    // remove connections
    Ivy_ObjDisconnect( p, pObj );
    // remove PIs/POs from the arrays
    if Ivy_ObjIsPi(pObj) {
        Vec_PtrRemove( p.vPis, pObj );
    } else if Ivy_ObjIsPo(pObj) {
        Vec_PtrRemove( p.vPos, pObj );
    } else if p.fFanout && Ivy_ObjIsBuf(pObj) {
        Vec_PtrRemove( p.vBufs, pObj ); }
    // clean and recycle the entry
    if fFreeTop != 0 {
        // free the node
        Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil );
        Ivy_ManRecycleMemory( p, pObj );
    
    } else {
        var nRefsOld int32  = pObj.nRefs;
        var pFanout *Ivy_Obj_t  = pObj.pFanout;
        Ivy_ObjClean( pObj );
        pObj.pFanout = pFanout;
        pObj.nRefs = nRefsOld;
    }
}

/**Function*************************************************************

  Synopsis    [Deletes the MFFC of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjDelete_rec(p *Ivy_Man_t, pObj *Ivy_Obj_t, fFreeTop int32) {
    var pFanin0, pFanin1 *Ivy_Obj_t;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(!Ivy_ObjIsNone(pObj), "!Ivy_ObjIsNone(pObj)");
    if Ivy_ObjIsConst1(pObj) || Ivy_ObjIsPi(pObj) {
        return; }
    pFanin0 = Ivy_ObjFanin0(pObj);
    pFanin1 = Ivy_ObjFanin1(pObj);
    Ivy_ObjDelete( p, pObj, fFreeTop );
    if pFanin0 && !Ivy_ObjIsNone(pFanin0) && Ivy_ObjRefs(pFanin0) == 0 {
        Ivy_ObjDelete_rec( p, pFanin0, 1 ); }
    if (pFanin1 != nil) && !Ivy_ObjIsNone(pFanin1) && Ivy_ObjRefs(pFanin1) == 0 {
        Ivy_ObjDelete_rec( p, pFanin1, 1 ); }
}

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

