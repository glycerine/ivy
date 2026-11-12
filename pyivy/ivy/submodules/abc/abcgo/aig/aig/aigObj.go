package aig

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [aigObj.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Adding/removing objects.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigObj.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

  Synopsis    [Creates primary input.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCreateCi(p *Aig_Man_t) *Aig_Obj_t {
    var pObj *Aig_Obj_t;
    pObj = Aig_ManFetchMemory( p );
    pObj.Type = AIG_OBJ_CI;
    Vec_PtrPush( p.vCis, pObj );
    p.nObjs[AIG_OBJ_CI]++;
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Creates primary output with the given driver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCreateCo(p *Aig_Man_t, pDriver *Aig_Obj_t) *Aig_Obj_t {
    var pObj *Aig_Obj_t;
    pObj = Aig_ManFetchMemory( p );
    pObj.Type = AIG_OBJ_CO;
    Vec_PtrPush( p.vCos, pObj );
    Aig_ObjConnect( p, pObj, pDriver, nil );
    p.nObjs[AIG_OBJ_CO]++;
    return pObj;
}


/**Function*************************************************************

  Synopsis    [Create the new node assuming it does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCreate(p *Aig_Man_t, pGhost *Aig_Obj_t) *Aig_Obj_t {
    var pObj *Aig_Obj_t;
    assert(!Aig_IsComplement(pGhost), "!Aig_IsComplement(pGhost)");
    assert(Aig_ObjIsHash(pGhost), "Aig_ObjIsHash(pGhost)");
//    assert( pGhost == &p->Ghost );
    // get memory for the new object
    pObj = Aig_ManFetchMemory( p );
    pObj.Type = pGhost.Type;
    // add connections
    Aig_ObjConnect( p, pObj, pGhost.pFanin0, pGhost.pFanin1 );
    // update node counters of the manager
    p.nObjs[Aig_ObjType(pObj)]++;
    assert(pObj.pData == nil, "pObj.pData == nil");
    // create the power counter
    if p.vProbs {
        var Prob0 float32  = Abc_Int2Float( Vec_IntEntry( p.vProbs, Aig_ObjFaninId0(pObj) ) );
        var Prob1 float32  = Abc_Int2Float( Vec_IntEntry( p.vProbs, Aig_ObjFaninId1(pObj) ) );
        if Aig_ObjFaninC0(pObj) {
            Prob0 = 1.0 - Prob0;
        } else {
            Prob0 = Prob0;
        }
        if Aig_ObjFaninC1(pObj) {
            Prob1 = 1.0 - Prob1;
        } else {
            Prob1 = Prob1;
        }
        Vec_IntSetEntry( p.vProbs, pObj.Id, Abc_Float2Int(Prob0 * Prob1) );
    }
    return pObj;
}
 
/**Function*************************************************************

  Synopsis    [Connect the object to the fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjConnect(p *Aig_Man_t, pObj *Aig_Obj_t, pFan0 *Aig_Obj_t, pFan1 *Aig_Obj_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(!Aig_ObjIsCi(pObj), "!Aig_ObjIsCi(pObj)");
    // add the first fanin
    pObj.pFanin0 = pFan0;
    pObj.pFanin1 = pFan1;
    // increment references of the fanins and add their fanouts
    if pFan0 != nil {
        assert(Aig_ObjFanin0(pObj).Type > 0, "Aig_ObjFanin0(pObj).Type > 0");
        Aig_ObjRef( Aig_ObjFanin0(pObj) );
        if p.pFanData {
            Aig_ObjAddFanout( p, Aig_ObjFanin0(pObj), pObj ); }
    }
    if pFan1 != nil {
        assert(Aig_ObjFanin1(pObj).Type > 0, "Aig_ObjFanin1(pObj).Type > 0");
        Aig_ObjRef( Aig_ObjFanin1(pObj) );
        if p.pFanData {
            Aig_ObjAddFanout( p, Aig_ObjFanin1(pObj), pObj ); }
    }
    // set level and phase
    pObj.Level = Aig_ObjLevelNew( pObj );
    pObj.fPhase = Aig_ObjPhaseReal(pFan0) & Aig_ObjPhaseReal(pFan1);
    // add the node to the structural hash table
    if p.pTable && Aig_ObjIsHash(pObj) {
        Aig_TableInsert( p, pObj ); }
    // add the node to the dynamically updated topological order
//    if ( p->pOrderData && Aig_ObjIsNode(pObj) )
//        Aig_ObjOrderInsert( p, pObj->Id );
    assert(!Aig_ObjIsNode(pObj) || pObj.Level > 0, "!Aig_ObjIsNode(pObj) || pObj.Level > 0");
}

/**Function*************************************************************

  Synopsis    [Disconnects the object from the fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDisconnect(p *Aig_Man_t, pObj *Aig_Obj_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    // remove connections
    if pObj.pFanin0 != nil {
        if p.pFanData {
            Aig_ObjRemoveFanout( p, Aig_ObjFanin0(pObj), pObj ); }
        Aig_ObjDeref(Aig_ObjFanin0(pObj));
    }
    if pObj.pFanin1 != nil {
        if p.pFanData {
            Aig_ObjRemoveFanout( p, Aig_ObjFanin1(pObj), pObj ); }
        Aig_ObjDeref(Aig_ObjFanin1(pObj));
    }
    // remove the node from the structural hash table
    if p.pTable && Aig_ObjIsHash(pObj) {
        Aig_TableDelete( p, pObj ); }
    // add the first fanin
    pObj.pFanin0 = nil;
    pObj.pFanin1 = nil;
    // remove the node from the dynamically updated topological order
//    if ( p->pOrderData && Aig_ObjIsNode(pObj) )
//        Aig_ObjOrderRemove( p, pObj->Id );
}

/**Function*************************************************************

  Synopsis    [Deletes the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDelete(p *Aig_Man_t, pObj *Aig_Obj_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(!Aig_ObjIsTerm(pObj), "!Aig_ObjIsTerm(pObj)");
    assert(Aig_ObjRefs(pObj) == 0, "Aig_ObjRefs(pObj) == 0");
    if p.pFanData && Aig_ObjIsBuf(pObj) {
        Vec_PtrRemove( p.vBufs, pObj ); }
    p.nObjs[pObj.Type]--;
    Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil );
    Aig_ManRecycleMemory( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Deletes the MFFC of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDelete_rec(p *Aig_Man_t, pObj *Aig_Obj_t, fFreeTop int32) {
    var pFanin0, pFanin1 *Aig_Obj_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Aig_ObjIsConst1(pObj) || Aig_ObjIsCi(pObj) {
        return; }
    assert(!Aig_ObjIsCo(pObj), "!Aig_ObjIsCo(pObj)");
    pFanin0 = Aig_ObjFanin0(pObj);
    pFanin1 = Aig_ObjFanin1(pObj);
    Aig_ObjDisconnect( p, pObj );
    if fFreeTop != 0 {
        Aig_ObjDelete( p, pObj ); }
    if (pFanin0 != nil) && !Aig_ObjIsNone(pFanin0) && Aig_ObjRefs(pFanin0) == 0 {
        Aig_ObjDelete_rec( p, pFanin0, 1 ); }
    if (pFanin1 != nil) && !Aig_ObjIsNone(pFanin1) && Aig_ObjRefs(pFanin1) == 0 {
        Aig_ObjDelete_rec( p, pFanin1, 1 ); }
}

/**Function*************************************************************

  Synopsis    [Deletes the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjDeletePo(p *Aig_Man_t, pObj *Aig_Obj_t) {
    assert(Aig_ObjIsCo(pObj), "Aig_ObjIsCo(pObj)");
    Aig_ObjDeref(Aig_ObjFanin0(pObj));
    pObj.pFanin0 = nil;
    p.nObjs[pObj.Type]--;
    Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil );
    Aig_ManRecycleMemory( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Replaces the first fanin of the node by the new fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjPatchFanin0(p *Aig_Man_t, pObj *Aig_Obj_t, pFaninNew *Aig_Obj_t) {
    var pFaninOld *Aig_Obj_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(Aig_ObjIsCo(pObj), "Aig_ObjIsCo(pObj)");
    pFaninOld = Aig_ObjFanin0(pObj);
    // decrement ref and remove fanout
    if p.pFanData {
        Aig_ObjRemoveFanout( p, pFaninOld, pObj ); }
    Aig_ObjDeref( pFaninOld );
    // update the fanin
    pObj.pFanin0 = pFaninNew;
    pObj.Level = Aig_ObjLevelNew( pObj );
    pObj.fPhase = Aig_ObjPhaseReal(pObj.pFanin0);
    // increment ref and add fanout
    if p.pFanData {
        Aig_ObjAddFanout( p, Aig_ObjFanin0(pObj), pObj ); }
    Aig_ObjRef( Aig_ObjFanin0(pObj) );
    // get rid of old fanin
    if !Aig_ObjIsCi(pFaninOld) && !Aig_ObjIsConst1(pFaninOld) && Aig_ObjRefs(pFaninOld) == 0 {
        Aig_ObjDelete_rec( p, pFaninOld, 1 ); }
}

/**Function*************************************************************

  Synopsis    [Verbose printing of the AIG node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjPrint(p *Aig_Man_t, pObj *Aig_Obj_t) {
    var fShowFanouts int32;
    var pTemp *Aig_Obj_t;
    if pObj == nil {
        fmt.Printf("Object is NULL.");
        return;
    }
    if Aig_IsComplement(pObj) {
        fmt.Printf("Compl ");
        pObj = Aig_Not(pObj);
    }
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    fmt.Printf("Node %4d : ", Aig_ObjId(pObj));
    if Aig_ObjIsConst1(pObj) {
        fmt.Printf("constant 1");
    } else if Aig_ObjIsCi(pObj) {
        fmt.Printf("PI");
    } else if Aig_ObjIsCo(pObj) {
        fmt.Printf("PO( %4d%s )", Aig_ObjFanin0(pObj).Id, (Aig_ObjFaninC0(pObj)? "\'" : " "));
    } else if Aig_ObjIsBuf(pObj) {
        fmt.Printf("BUF( %d%s )", Aig_ObjFanin0(pObj).Id, (Aig_ObjFaninC0(pObj)? "\'" : " "));
    } else {
        fmt.Printf("AND( %4d%s, %4d%s )", Aig_ObjFanin0(pObj).Id, (Aig_ObjFaninC0(pObj)? "\'" : " "), Aig_ObjFanin1(pObj).Id, (Aig_ObjFaninC1(pObj)? "\'" : " ")); }
    fmt.Printf(" (refs = %3d)", Aig_ObjRefs(pObj));
    if (fShowFanouts != 0) && p.pFanData {
        var pFanout *Aig_Obj_t;
        var i int32;
        var iFan int32  = -1; // Suppress "might be used uninitialized"
        fmt.Printf("\nFanouts:\n");
        // inlining of macro Aig_ObjForEachFanout
        for  assert(p.pFanData, "p.pFanData"), i = 0; (i < int32((pObj).nRefs)) &&                         (((iFan) = i? Aig_ObjFanoutNext(p, iFan) : Aig_ObjFanout0Int(p, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(p, iFan>>1)), 1); i++  {
            fmt.Printf("    ");
            fmt.Printf("Node %4d : ", Aig_ObjId(pFanout));
            if Aig_ObjIsCo(pFanout) {
                fmt.Printf("PO( %4d%s )", Aig_ObjFanin0(pFanout).Id, (Aig_ObjFaninC0(pFanout)? "\'" : " "));
            } else if Aig_ObjIsBuf(pFanout) {
                fmt.Printf("BUF( %d%s )", Aig_ObjFanin0(pFanout).Id, (Aig_ObjFaninC0(pFanout)? "\'" : " "));
            } else {
                fmt.Printf("AND( %4d%s, %4d%s )", Aig_ObjFanin0(pFanout).Id, (Aig_ObjFaninC0(pFanout)? "\'" : " "), Aig_ObjFanin1(pFanout).Id, (Aig_ObjFaninC1(pFanout)? "\'" : " ")); }
            fmt.Printf("\n");
        }
        return;
    }
    // there are choices
    if p.pEquivs && p.pEquivs[pObj.Id] {
        // print equivalence class
        fmt.Printf("  { %4d ", pObj.Id);
        for  pTemp = p.pEquivs[pObj.Id]; (pTemp != nil); pTemp = p.pEquivs[pTemp.Id]  {
            fmt.Printf(" %4d%s", pTemp.Id, (pTemp.fPhase != pObj.fPhase)? "\'" : " ");
        }
        fmt.Printf(" }");
        return;
    }
    // this is a secondary node
    if p.pReprs && p.pReprs[pObj.Id] {
        fmt.Printf("  class of %d", pObj.Id); }
}

/**Function*************************************************************

  Synopsis    [Replaces node with a buffer fanin by a node without them.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeFixBufferFanins(p *Aig_Man_t, pObj *Aig_Obj_t, fUpdateLevel int32) {
    var pFanReal0 *Aig_Obj_t;
    var pFanReal1 *Aig_Obj_t;
    var pResult *Aig_Obj_t = nil;
    p.nBufFixes++;
    if Aig_ObjIsCo(pObj) {
        assert(Aig_ObjIsBuf(Aig_ObjFanin0(pObj)), "Aig_ObjIsBuf(Aig_ObjFanin0(pObj))");
        pFanReal0 = Aig_ObjReal_rec( Aig_ObjChild0(pObj) );
        assert(Aig_ObjPhaseReal(Aig_ObjChild0(pObj)) == Aig_ObjPhaseReal(pFanReal0), "Aig_ObjPhaseReal(Aig_ObjChild0(pObj)) == Aig_ObjPhaseReal(pFanReal0)");
        Aig_ObjPatchFanin0( p, pObj, pFanReal0 );
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    assert(Aig_ObjIsBuf(Aig_ObjFanin0(pObj)) || Aig_ObjIsBuf(Aig_ObjFanin1(pObj)), "Aig_ObjIsBuf(Aig_ObjFanin0(pObj)) || Aig_ObjIsBuf(Aig_ObjFanin1(pObj))");
    // get the real fanins
    pFanReal0 = Aig_ObjReal_rec( Aig_ObjChild0(pObj) );
    pFanReal1 = Aig_ObjReal_rec( Aig_ObjChild1(pObj) );
    // get the new node
    if Aig_ObjIsNode(pObj) {
        pResult = Aig_Oper( p, pFanReal0, pFanReal1, Aig_ObjType(pObj) );
//    else if ( Aig_ObjIsLatch(pObj) )
//        pResult = Aig_Latch( p, pFanReal0, Aig_ObjInit(pObj) );
    } else { 
        assert(false, "0"); }
    // replace the node with buffer by the node without buffer
    Aig_ObjReplace( p, pObj, pResult, fUpdateLevel );
}

/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Aig_ManPropagateBuffers(p *Aig_Man_t, fUpdateLevel int32) int32 {
    var pObj *Aig_Obj_t;
    var nSteps int32;
    assert(p.pFanData, "p.pFanData");
    for  nSteps = 0; Vec_PtrSize(p.vBufs) > 0; nSteps++  {
        // get the node with a buffer fanin
        for  pObj = (Aig_Obj_t *)Vec_PtrEntryLast(p.vBufs); Aig_ObjIsBuf(pObj); pObj = Aig_ObjFanout0(p, pObj)  {}
        // replace this node by a node without buffer
        Aig_NodeFixBufferFanins( p, pObj, fUpdateLevel );
        // stop if a cycle occured
        if nSteps > 1000000 {
            fmt.Printf("Error: A cycle is encountered while propagating buffers.\n");
            break;
        }
    }
    return nSteps;
}

/**Function*************************************************************

  Synopsis    [Replaces one object by another.]

  Description [The new object (pObjNew) should be used instead of the old 
  object (pObjOld). If the new object is complemented or used, the buffer 
  is added and the new object remains in the manager; otherwise, the new
  object is deleted.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjReplace(p *Aig_Man_t, pObjOld *Aig_Obj_t, pObjNew *Aig_Obj_t, fUpdateLevel int32) {
    var pObjNewR *Aig_Obj_t  = Aig_Regular(pObjNew);
    // the object to be replaced cannot be complemented
    assert(!Aig_IsComplement(pObjOld), "!Aig_IsComplement(pObjOld)");
    // the object to be replaced cannot be a terminal
    assert(!Aig_ObjIsCi(pObjOld) && !Aig_ObjIsCo(pObjOld), "!Aig_ObjIsCi(pObjOld) && !Aig_ObjIsCo(pObjOld)");
    // the object to be used cannot be a buffer or a PO
    assert(!Aig_ObjIsBuf(pObjNewR) && !Aig_ObjIsCo(pObjNewR), "!Aig_ObjIsBuf(pObjNewR) && !Aig_ObjIsCo(pObjNewR)");
    // the object cannot be the same
    assert(pObjOld != pObjNewR, "pObjOld != pObjNewR");
    // make sure object is not pointing to itself
    assert(pObjOld != Aig_ObjFanin0(pObjNewR), "pObjOld != Aig_ObjFanin0(pObjNewR)");
    assert(pObjOld != Aig_ObjFanin1(pObjNewR), "pObjOld != Aig_ObjFanin1(pObjNewR)");
    if pObjOld == Aig_ObjFanin0(pObjNewR) || pObjOld == Aig_ObjFanin1(pObjNewR) {
        fmt.Printf("Aig_ObjReplace(): Internal error!\n");
        os.Exit(int(1));
    }
    // recursively delete the old node - but leave the object there
    pObjNewR.nRefs++;
    Aig_ObjDelete_rec( p, pObjOld, 0 );
    pObjNewR.nRefs--;
    // if the new object is complemented or already used, create a buffer
    p.nObjs[pObjOld.Type]--;
    if Aig_IsComplement(pObjNew) || Aig_ObjRefs(pObjNew) > 0 || !Aig_ObjIsNode(pObjNew) {
        pObjOld.Type = AIG_OBJ_BUF;
        Aig_ObjConnect( p, pObjOld, pObjNew, nil );
        p.nBufReplaces++;
    
    } else {
        var pFanin0 *Aig_Obj_t  = pObjNew.pFanin0;
        var pFanin1 *Aig_Obj_t  = pObjNew.pFanin1;
        var LevelOld int32  = pObjOld.Level;
        pObjOld.Type = pObjNew.Type;
        Aig_ObjDisconnect( p, pObjNew );
        Aig_ObjConnect( p, pObjOld, pFanin0, pFanin1 );
        // delete the new object
        Aig_ObjDelete( p, pObjNew );
        // update levels
        if p.pFanData {
            pObjOld.Level = LevelOld;
            Aig_ManUpdateLevel( p, pObjOld );
        }
        if fUpdateLevel != 0 {
            Aig_ObjClearReverseLevel( p, pObjOld );
            Aig_ManUpdateReverseLevel( p, pObjOld );
        }
    }
    p.nObjs[pObjOld.Type]++;
    // store buffers if fanout is allocated
    if p.pFanData && Aig_ObjIsBuf(pObjOld) {
        Vec_PtrPush( p.vBufs, pObjOld );
        p.nBufMax = Abc_MaxInt( p.nBufMax, Vec_PtrSize(p.vBufs) );
        Aig_ManPropagateBuffers( p, fUpdateLevel );
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

