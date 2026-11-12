package ivy

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [ivyMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [AIG manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivy_.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

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

  Synopsis    [Starts the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManStart is implemented in ivyLife.go.


/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManStartFrom is implemented in ivyLife.go.


/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManDup(p *Ivy_Man_t) *Ivy_Man_t {
    var vNodes, vLatches *Vec_Int_t;
    var pNew *Ivy_Man_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    // collect latches and nodes in the DFS order
    vNodes = Ivy_ManDfsSeq( p, &vLatches );
    // create the new manager
    pNew = Ivy_ManStart();
    // create the PIs
    Ivy_ManConst1(p).pEquiv = Ivy_ManConst1(pNew);
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pObj.pEquiv = Ivy_ObjCreatePi(pNew);
    }
    // create the fake PIs for latches
    // inlining of macro Ivy_ManForEachNodeVec
    for  i = 0; i < Vec_IntSize(vLatches) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vLatches,i))); i++  {
        pObj.pEquiv = Ivy_ObjCreatePi(pNew);
    }
    // duplicate internal nodes
    // inlining of macro Ivy_ManForEachNodeVec
    for  i = 0; i < Vec_IntSize(vNodes) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        if Ivy_ObjIsBuf(pObj) {
            pObj.pEquiv = Ivy_ObjChild0Equiv(pObj);
        } else {
            pObj.pEquiv = Ivy_And( pNew, Ivy_ObjChild0Equiv(pObj), Ivy_ObjChild1Equiv(pObj) ); }
    }
    // add the POs
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Ivy_ObjCreatePo( pNew, Ivy_ObjChild0Equiv(pObj) );
    }
    // transform additional PI nodes into latches and connect them
    // inlining of macro Ivy_ManForEachNodeVec
    for  i = 0; i < Vec_IntSize(vLatches) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vLatches,i))); i++  {
        assert(!Ivy_ObjFaninC0(pObj), "!Ivy_ObjFaninC0(pObj)");
        pObj.pEquiv.Type = IVY_LATCH;
        pObj.pEquiv.Init = pObj.Init;
        Ivy_ObjConnect( pNew, pObj.pEquiv, Ivy_ObjChild0Equiv(pObj), nil );
    }
    // shrink the arrays
    Vec_PtrShrink( pNew.vPis, Ivy_ManPiNum(p) );
    // update the counters of different objects
    pNew.nObjs[IVY_PI] -= Ivy_ManLatchNum(p);
    pNew.nObjs[IVY_LATCH] += Ivy_ManLatchNum(p);
    // free arrays
    Vec_IntFree( vNodes );
    Vec_IntFree( vLatches );
    // make sure structural hashing did not change anything
    assert(Ivy_ManNodeNum(p)  == Ivy_ManNodeNum(pNew), "Ivy_ManNodeNum(p)  == Ivy_ManNodeNum(pNew)");
    assert(Ivy_ManLatchNum(p) == Ivy_ManLatchNum(pNew), "Ivy_ManLatchNum(p) == Ivy_ManLatchNum(pNew)");
    // check the resulting network
    if !Ivy_ManCheck(pNew) {
        fmt.Printf("Ivy_ManMakeSeq(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManFrames(pMan *Ivy_Man_t, nLatches int32, nFrames int32, fInit int32, pvMapping **Vec_Ptr_t) *Ivy_Man_t {
    var vMapping *Vec_Ptr_t;
    var pNew *Ivy_Man_t;
    var pObj *Ivy_Obj_t;
    var i, f, nPis, nPos, nIdMax int32;
    assert(Ivy_ManLatchNum(pMan) == 0, "Ivy_ManLatchNum(pMan) == 0");
    assert(nFrames > 0, "nFrames > 0");
    // prepare the mapping
    nPis = Ivy_ManPiNum(pMan) - nLatches;
    nPos = Ivy_ManPoNum(pMan) - nLatches;
    nIdMax = Ivy_ManObjIdMax(pMan);
    // create the new manager
    pNew = Ivy_ManStart();
    // set the starting values of latch inputs
    for  i = 0; i < nLatches; i++  {
        Ivy_ManPo(pMan, nPos+i).pEquiv = fInit? Ivy_Not(Ivy_ManConst1(pNew)) : Ivy_ObjCreatePi(pNew);
    }
    // add timeframes
    vMapping = Vec_PtrStart( nIdMax * nFrames + 1 );
    for  f = 0; f < nFrames; f++  {
        // create PIs
        Ivy_ManConst1(pMan).pEquiv = Ivy_ManConst1(pNew);
        for  i = 0; i < nPis; i++  {
            Ivy_ManPi(pMan, i).pEquiv = Ivy_ObjCreatePi(pNew);
        }
        // transfer values to latch outputs
        for  i = 0; i < nLatches; i++  {
            Ivy_ManPi(pMan, nPis+i).pEquiv = Ivy_ManPo(pMan, nPos+i).pEquiv;
        }
        // perform strashing
        // inlining of macro Ivy_ManForEachNode
        // inlining of macro Ivy_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
            pObj.pEquiv = Ivy_And( pNew, Ivy_ObjChild0Equiv(pObj), Ivy_ObjChild1Equiv(pObj) ); }
        }
        // create POs
        for  i = 0; i < nPos; i++  {
            Ivy_ManPo(pMan, i).pEquiv = Ivy_ObjCreatePo( pNew, Ivy_ObjChild0Equiv(Ivy_ManPo(pMan, i)) );
        }
        // set the results of latch inputs
        for  i = 0; i < nLatches; i++  {
            Ivy_ManPo(pMan, nPos+i).pEquiv = Ivy_ObjChild0Equiv(Ivy_ManPo(pMan, nPos+i));
        }
        // save the pointers in this frame
        // inlining of macro Ivy_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            Vec_PtrWriteEntry( vMapping, f * nIdMax + i, pObj.pEquiv ); }
        }
    }
    // connect latches
    if !(fInit != 0) {
        for  i = 0; i < nLatches; i++  {
            Ivy_ObjCreatePo( pNew, Ivy_ManPo(pMan, nPos+i).pEquiv );
        } }
    // remove dangling nodes
    Ivy_ManCleanup(pNew);
    *pvMapping = vMapping;
    // check the resulting network
    if !Ivy_ManCheck(pNew) {
        fmt.Printf("Ivy_ManFrames(): The check has failed.\n"); }
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManStop is implemented in ivyLife.go.


/**Function*************************************************************

  Synopsis    [Removes nodes without fanout.]

  Description [Returns the number of dangling nodes removed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManCleanup is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Marks nodes reachable from the given one.]

  Description [Returns the number of dangling nodes removed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCleanupSeq_rec(pObj *Ivy_Obj_t) {
    if Ivy_ObjIsMarkA(pObj) {
        return; }
    Ivy_ObjSetMarkA(pObj);
    if pObj.pFanin0 != nil {
        Ivy_ManCleanupSeq_rec( Ivy_ObjFanin0(pObj) ); }
    if pObj.pFanin1 != nil {
        Ivy_ManCleanupSeq_rec( Ivy_ObjFanin1(pObj) ); }
}

/**Function*************************************************************

  Synopsis    [Removes logic that does not feed into POs.]

  Description [Returns the number of dangling nodes removed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManCleanupSeq(p *Ivy_Man_t) int32 {
    var vNodes *Vec_Ptr_t;
    var pObj *Ivy_Obj_t;
    var i, RetValue int32;
    // mark the constant and PIs
    Ivy_ObjSetMarkA( Ivy_ManConst1(p) );
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        Ivy_ObjSetMarkA( pObj );
    }
    // mark nodes visited from POs
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Ivy_ManCleanupSeq_rec( pObj );
    }
    // collect unmarked nodes
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Ivy_ObjIsMarkA(pObj) {
            Ivy_ObjClearMarkA(pObj);
        } else {
            Vec_PtrPush( vNodes, pObj ); }
    }
    }
    if Vec_PtrSize(vNodes) == 0 {
        Vec_PtrFree( vNodes );
//printf( "Sequential sweep cleaned out %d nodes.\n", 0 );
        return 0;
    }
    // disconnect the marked objects
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Ivy_ObjDisconnect( p, pObj );
    }
    // remove the dangling objects
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        assert(Ivy_ObjIsNode(pObj) || Ivy_ObjIsLatch(pObj) || Ivy_ObjIsBuf(pObj), "Ivy_ObjIsNode(pObj) || Ivy_ObjIsLatch(pObj) || Ivy_ObjIsBuf(pObj)");
        assert(Ivy_ObjRefs(pObj) == 0, "Ivy_ObjRefs(pObj) == 0");
        // update node counters of the manager
        p.nObjs[pObj.Type]--;
        p.nDeleted++;
        // delete buffer from the array of buffers
        if p.fFanout && Ivy_ObjIsBuf(pObj) {
            Vec_PtrRemove( p.vBufs, pObj ); }
        // free the node
        Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil );
        Ivy_ManRecycleMemory( p, pObj );
    }
    // return the number of nodes freed
    RetValue = Vec_PtrSize(vNodes);
    Vec_PtrFree( vNodes );
//printf( "Sequential sweep cleaned out %d nodes.\n", RetValue );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Checks if latches form self-loop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManLatchIsSelfFeed_rec(pLatch *Ivy_Obj_t, pLatchRoot *Ivy_Obj_t) int32 {
    if !Ivy_ObjIsLatch(pLatch) && !Ivy_ObjIsBuf(pLatch) {
        return 0; }
    if pLatch == pLatchRoot {
        return 1; }
    return Ivy_ManLatchIsSelfFeed_rec( Ivy_ObjFanin0(pLatch), pLatchRoot );
}

/**Function*************************************************************

  Synopsis    [Checks if latches form self-loop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManLatchIsSelfFeed(pLatch *Ivy_Obj_t) int32 {
    if !Ivy_ObjIsLatch(pLatch) {
        return 0; }
    return Ivy_ManLatchIsSelfFeed_rec( Ivy_ObjFanin0(pLatch), pLatch );
}


/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Ivy_ManPropagateBuffers(p *Ivy_Man_t, fUpdateLevel int32) int32 {
    var pNode *Ivy_Obj_t;
    var LimitFactor int32  = 100;
    var NodeBeg int32  = Ivy_ManNodeNum(p);
    var nSteps int32;
    for  nSteps = 0; Vec_PtrSize(p.vBufs) > 0; nSteps++  {
        pNode = (Ivy_Obj_t *)Vec_PtrEntryLast(p.vBufs);
        for  Ivy_ObjIsBuf(pNode)  {
            pNode = Ivy_ObjReadFirstFanout( p, pNode );
        }
        // check if this buffer should remain
        if Ivy_ManLatchIsSelfFeed(pNode) != 0 {
            Vec_PtrPop(p.vBufs);
            continue;
        }
//printf( "Propagating buffer %d with input %d and output %d\n", Ivy_ObjFaninId0(pNode), Ivy_ObjFaninId0(Ivy_ObjFanin0(pNode)), pNode->Id );
//printf( "Latch num %d\n", Ivy_ManLatchNum(p) );
        Ivy_NodeFixBufferFanins( p, pNode, fUpdateLevel );
        if nSteps > NodeBeg * LimitFactor {
            fmt.Printf("Structural hashing is not finished after %d forward latch moves.\n", NodeBeg * LimitFactor);
            fmt.Printf("This circuit cannot be forward-retimed completely. Quitting.\n");
            break;
        }
    }
//    printf( "Number of steps = %d. Nodes beg = %d. Nodes end = %d.\n", nSteps, NodeBeg, Ivy_ManNodeNum(p) );
    return nSteps;
}

/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManPrintStats(p *Ivy_Man_t) {
    fmt.Printf("PI/PO = %d/%d ", Ivy_ManPiNum(p), Ivy_ManPoNum(p));
    fmt.Printf("A = %7d. ", Ivy_ManAndNum(p));
    fmt.Printf("L = %5d. ", Ivy_ManLatchNum(p));
//    printf( "X = %d. ",       Ivy_ManExorNum(p) );
//    printf( "B = %3d. ",       Ivy_ManBufNum(p) );
    fmt.Printf("MaxID = %7d. ", Ivy_ManObjIdMax(p));
//    printf( "Cre = %d. ",     p->nCreated );
//    printf( "Del = %d. ",     p->nDeleted );
    fmt.Printf("Lev = %3d. ", Ivy_ManLatchNum(p)? -1 : Ivy_ManLevels(p));
    fmt.Printf("\n");
    fflush( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Converts a combinational AIG manager into a sequential one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManMakeSeq(p *Ivy_Man_t, nLatches int32, pInits *int32) {
    var pObj, pLatch *Ivy_Obj_t;
    var Init Ivy_Init_t;
    var i int32;
    if nLatches == 0 {
        return; }
    assert(nLatches < Ivy_ManPiNum(p) && nLatches < Ivy_ManPoNum(p), "nLatches < Ivy_ManPiNum(p) && nLatches < Ivy_ManPoNum(p)");
    assert(Ivy_ManPiNum(p) == Vec_PtrSize(p.vPis), "Ivy_ManPiNum(p) == Vec_PtrSize(p.vPis)");
    assert(Ivy_ManPoNum(p) == Vec_PtrSize(p.vPos), "Ivy_ManPoNum(p) == Vec_PtrSize(p.vPos)");
    assert(Vec_PtrSize( p.vBufs ) == 0, "Vec_PtrSize( p.vBufs ) == 0");
    // create fanouts
    if p.fFanout == 0 {
        Ivy_ManStartFanout( p ); }
    // collect the POs to be converted into latches
    for  i = 0; i < nLatches; i++  {
        // get the latch value
        if pInits != nil {
            Init = (Ivy_Init_t)pInits[i];
        } else {
            Init = IVY_INIT_0;
        }
        // create latch
        pObj = Ivy_ManPo( p, Ivy_ManPoNum(p) - nLatches + i );
        pLatch = Ivy_Latch( p, Ivy_ObjChild0(pObj), Init );
        Ivy_ObjDisconnect( p, pObj );
        // recycle the old PO object
        Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil );
        Ivy_ManRecycleMemory( p, pObj );
        // convert the corresponding PI to a buffer and connect it to the latch
        pObj = Ivy_ManPi( p, Ivy_ManPiNum(p) - nLatches + i );
        pObj.Type = IVY_BUF;
        Ivy_ObjConnect( p, pObj, pLatch, nil );
        // save the buffer
        Vec_PtrPush( p.vBufs, pObj );
    }
    // shrink the arrays
    Vec_PtrShrink( p.vPis, Ivy_ManPiNum(p) - nLatches );
    Vec_PtrShrink( p.vPos, Ivy_ManPoNum(p) - nLatches );
    // update the counters of different objects
    p.nObjs[IVY_PI] -= nLatches;
    p.nObjs[IVY_PO] -= nLatches;
    p.nObjs[IVY_BUF] += nLatches;
    p.nDeleted -= 2 * nLatches;
    // remove dangling nodes
    Ivy_ManCleanup(p);
    Ivy_ManCleanupSeq(p);
/* 
    // check for dangling nodes
    Ivy_ManForEachObj( p, pObj, i )
        if ( !Ivy_ObjIsPi(pObj) && !Ivy_ObjIsPo(pObj) && !Ivy_ObjIsConst1(pObj) )
        {
            assert( Ivy_ObjRefs(pObj) > 0 );
            assert( Ivy_ObjRefs(pObj) == Ivy_ObjFanoutNum(p, pObj) );
        }
*/
    // perform hashing by propagating the buffers
    Ivy_ManPropagateBuffers( p, 0 );
    if Ivy_ManBufNum(p) {
        fmt.Printf("The number of remaining buffers is %d.\n", Ivy_ManBufNum(p)); }
    // fix the levels
    Ivy_ManResetLevels( p );
    // check the resulting network
    if !Ivy_ManCheck(p) {
        fmt.Printf("Ivy_ManMakeSeq(): The check has failed.\n"); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

