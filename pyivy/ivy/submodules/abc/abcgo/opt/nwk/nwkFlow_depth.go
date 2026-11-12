package nwk

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [nwkFlow.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Netlist representation.]

  Synopsis    [Max-flow/min-cut computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkFlow.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nwk.h"

//ABC_NAMESPACE_IMPL_START


/*
    This code is based on the papers:
    A. Hurst, A. Mishchenko, and R. Brayton, "Fast minimum-register retiming 
    via binary maximum-flow", Proc. FMCAD '07, pp. 181-187. 
    A. Hurst, A. Mishchenko, and R. Brayton, "Scalable min-area retiming 
    under simultaneous delay and initial state constraints". Proc. DAC'08. 
*/

var DepthFwd, DepthBwd, DepthFwdMax, DepthBwdMax int32;

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// predecessors
func Nwk_ObjPred(pObj *Nwk_Obj_t) *Nwk_Obj_t { return pObj.pCopy;        }
func Nwk_ObjSetPred(pObj *Nwk_Obj_t, p *Nwk_Obj_t) int32 { pObj.pCopy = p; return 1; }
// sink
func Nwk_ObjIsSink(pObj *Nwk_Obj_t) int32 { return pObj.MarkA;        }
func Nwk_ObjSetSink(pObj *Nwk_Obj_t) { pObj.MarkA = 1;           }
// flow
func Nwk_ObjHasFlow(pObj *Nwk_Obj_t) int32 { return pObj.MarkB;        }
func Nwk_ObjSetFlow(pObj *Nwk_Obj_t) { pObj.MarkB = 1;           }
func Nwk_ObjClearFlow(pObj *Nwk_Obj_t) { pObj.MarkB = 0;           }
 
// representation of visited nodes
// pObj->TravId  < pNtk->nTravIds-2   --- not visited
// pObj->TravId == pNtk->nTravIds-2   --- visited bot only
// pObj->TravId == pNtk->nTravIds-1   --- visited top only
// pObj->TravId == pNtk->nTravIds     --- visited bot and top
func Nwk_ObjVisitedBotOnly(pObj *Nwk_Obj_t) int32 {
    return pObj.TravId == pObj.pMan.nTravIds - 2;
}
func Nwk_ObjVisitedBot(pObj *Nwk_Obj_t) int32 { 
    return pObj.TravId == pObj.pMan.nTravIds - 2 || pObj.TravId == pObj.pMan.nTravIds; 
}
func Nwk_ObjVisitedTop(pObj *Nwk_Obj_t) int32 { 
    return pObj.TravId == pObj.pMan.nTravIds - 1 || pObj.TravId == pObj.pMan.nTravIds; 
}
func Nwk_ObjSetVisitedBot(pObj *Nwk_Obj_t) {
    if pObj.TravId < pObj.pMan.nTravIds - 2 {
        pObj.TravId = pObj.pMan.nTravIds - 2;
    } else if pObj.TravId == pObj.pMan.nTravIds - 1 {
        pObj.TravId = pObj.pMan.nTravIds;
    } else {
        assert(false, "0"); }
}
func Nwk_ObjSetVisitedTop(pObj *Nwk_Obj_t) {
    if pObj.TravId < pObj.pMan.nTravIds - 2 {
        pObj.TravId = pObj.pMan.nTravIds - 1;
    } else if pObj.TravId == pObj.pMan.nTravIds - 2 {
        pObj.TravId = pObj.pMan.nTravIds;
    } else {
        assert(false, "0"); }
}
func Nwk_ManIncrementTravIdFlow(pMan *Nwk_Man_t) {
    Nwk_ManIncrementTravId( pMan );
    Nwk_ManIncrementTravId( pMan );
    Nwk_ManIncrementTravId( pMan );
}

// static int Nwk_ManPushForwardTop_rec( Nwk_Obj_t * pObj, Nwk_Obj_t * pPred );

// static int Nwk_ManPushForwardBot_rec( Nwk_Obj_t * pObj, Nwk_Obj_t * pPred );


// static int Nwk_ManPushBackwardTop_rec( Nwk_Obj_t * pObj, Nwk_Obj_t * pPred );

// static int Nwk_ManPushBackwardBot_rec( Nwk_Obj_t * pObj, Nwk_Obj_t * pPred );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Marks TFI of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManMarkTfiCone_rec(pObj *Nwk_Obj_t) {
    var pNext *Nwk_Obj_t;
    var i int32;
    if pObj.MarkA {
        return; }
    pObj.MarkA = 1;
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[i]); i++  {
        Nwk_ManMarkTfiCone_rec( pNext );
    }
}

/**Function*************************************************************

  Synopsis    [Marks TFO of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManMarkTfoCone_rec(pObj *Nwk_Obj_t) {
    var pNext *Nwk_Obj_t;
    var i int32;
    if pObj.MarkA {
        return; }
    pObj.MarkA = 1;
    // inlining of macro Nwk_ObjForEachFanout
    for  i = 0; (i < int32((pObj).nFanouts)) && ((pNext) = (pObj).pFanio[(pObj).nFanins+i]); i++  {
        Nwk_ManMarkTfoCone_rec( pNext );
    }
}

/**Function*************************************************************

  Synopsis    [Fast forward flow pushing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPushForwardFast_rec(pObj *Nwk_Obj_t, pPred *Nwk_Obj_t) int32 {
    var pNext *Nwk_Obj_t;
    var i int32;
    if Nwk_ObjIsTravIdCurrent( pObj ) {
        return 0; }
    Nwk_ObjSetTravIdCurrent( pObj );
    if Nwk_ObjHasFlow(pObj) != 0 {
        return 0; }
    if Nwk_ObjIsSink(pObj) != 0 {
        Nwk_ObjSetFlow(pObj);
        return Nwk_ObjSetPred( pObj, pPred );
    }
    // inlining of macro Nwk_ObjForEachFanout
    for  i = 0; (i < int32((pObj).nFanouts)) && ((pNext) = (pObj).pFanio[(pObj).nFanins+i]); i++  {
        if Nwk_ManPushForwardFast_rec( pNext, pObj ) != 0 {
            Nwk_ObjSetFlow(pObj);
            return Nwk_ObjSetPred( pObj, pPred );
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Fast backward flow pushing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPushBackwardFast_rec(pObj *Nwk_Obj_t, pPred *Nwk_Obj_t) int32 {
    var pNext *Nwk_Obj_t;
    var i int32;
    if Nwk_ObjIsTravIdCurrent( pObj ) {
        return 0; }
    Nwk_ObjSetTravIdCurrent( pObj );
    if Nwk_ObjHasFlow(pObj) != 0 {
        return 0; }
    if Nwk_ObjIsSink(pObj) != 0 {
        Nwk_ObjSetFlow(pObj);
        return Nwk_ObjSetPred( pObj, pPred );
    }
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[i]); i++  {
        if Nwk_ManPushBackwardFast_rec( pNext, pObj ) != 0 {
            Nwk_ObjSetFlow(pObj);
            return Nwk_ObjSetPred( pObj, pPred );
        }
    }
    return 0;
}
  
/**Function*************************************************************

  Synopsis    [Pushing the flow through the bottom part of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPushForwardBot_rec(pObj *Nwk_Obj_t, pPred *Nwk_Obj_t) int32 {
    var pNext *Nwk_Obj_t;
    var i int32;
    if Nwk_ObjVisitedBot(pObj) != 0 {
        return 0; }
    Nwk_ObjSetVisitedBot(pObj);
    DepthFwd++;
    if DepthFwdMax < DepthFwd {
        DepthFwdMax = DepthFwd; }
    // propagate through the internal edge
    if Nwk_ObjHasFlow(pObj) != 0 {
        if Nwk_ObjPred(pObj) != nil {
            if Nwk_ManPushForwardTop_rec( Nwk_ObjPred(pObj), Nwk_ObjPred(pObj) ) != 0 {
                DepthFwd--;
                return Nwk_ObjSetPred( pObj, pPred ); 
            } }
    
    } else if Nwk_ManPushForwardTop_rec(pObj, pObj) != 0 {
        DepthFwd--;
        Nwk_ObjSetFlow( pObj );
        return Nwk_ObjSetPred( pObj, pPred );
    }
    // try to push through the fanins
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[i]); i++  {
        if Nwk_ManPushForwardBot_rec( pNext, pPred ) != 0 {
            DepthFwd--;
            return 1;
        }
    }
    DepthFwd--;
    return 0;
}

/**Function*************************************************************

  Synopsis    [Pushing the flow through the top part of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPushForwardTop_rec(pObj *Nwk_Obj_t, pPred *Nwk_Obj_t) int32 {
    var pNext *Nwk_Obj_t;
    var i int32;
    if Nwk_ObjVisitedTop(pObj) != 0 {
        return 0; }
    Nwk_ObjSetVisitedTop(pObj);
    // check if this is the sink
    if Nwk_ObjIsSink(pObj) != 0 {
        return 1; }
    DepthFwd++;
    if DepthFwdMax < DepthFwd {
        DepthFwdMax = DepthFwd; }
    // try to push through the fanouts
    // inlining of macro Nwk_ObjForEachFanout
    for  i = 0; (i < int32((pObj).nFanouts)) && ((pNext) = (pObj).pFanio[(pObj).nFanins+i]); i++  {
        if Nwk_ManPushForwardBot_rec( pNext, pPred ) != 0 {
            DepthFwd--;
            return 1;
        }
    }
    // redirect the flow
    if (Nwk_ObjHasFlow(pObj) != 0) && !Nwk_ObjIsCi(pObj) {
        if Nwk_ManPushForwardBot_rec( pObj, Nwk_ObjPred(pObj) ) != 0 {
            DepthFwd--;
            Nwk_ObjClearFlow( pObj );
            return Nwk_ObjSetPred( pObj, nil );
        } }
    DepthFwd--;
    return 0;
}
  
/**Function*************************************************************

  Synopsis    [Pushing the flow through the bottom part of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPushBackwardBot_rec(pObj *Nwk_Obj_t, pPred *Nwk_Obj_t) int32 {
    if Nwk_ObjVisitedBot(pObj) != 0 {
        return 0; }
    Nwk_ObjSetVisitedBot(pObj);
    // propagate through the internal edge
    if Nwk_ObjHasFlow(pObj) != 0 {
        if Nwk_ObjPred(pObj) != nil {
            if Nwk_ManPushBackwardTop_rec( Nwk_ObjPred(pObj), Nwk_ObjPred(pObj) ) != 0 {
                return Nwk_ObjSetPred( pObj, pPred ); } } 
    
    } else if Nwk_ManPushBackwardTop_rec(pObj, pObj) != 0 {
        Nwk_ObjSetFlow( pObj );
        return Nwk_ObjSetPred( pObj, pPred );
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Pushing the flow through the top part of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPushBackwardTop_rec(pObj *Nwk_Obj_t, pPred *Nwk_Obj_t) int32 {
    var pNext *Nwk_Obj_t;
    var i int32;
    if Nwk_ObjVisitedTop(pObj) != 0 {
        return 0; }
    Nwk_ObjSetVisitedTop(pObj);
    // check if this is the sink
    if Nwk_ObjIsSink(pObj) != 0 {
        return 1; }
    // try to push through the fanins
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[i]); i++  {
        if Nwk_ManPushBackwardBot_rec( pNext, pPred ) != 0 {
            return 1; }
    }
    // try to push through the fanouts
    // inlining of macro Nwk_ObjForEachFanout
    for  i = 0; (i < int32((pObj).nFanouts)) && ((pNext) = (pObj).pFanio[(pObj).nFanins+i]); i++  {
        if !Nwk_ObjIsCo(pObj) && (Nwk_ManPushBackwardTop_rec( pNext, pPred ) != 0) {
            return 1; }
    }
    // redirect the flow
    if Nwk_ObjHasFlow(pObj) != 0 {
        if (Nwk_ObjPred(pObj) != nil) && (Nwk_ManPushBackwardBot_rec( pObj, Nwk_ObjPred(pObj) ) != 0) {
            Nwk_ObjClearFlow( pObj );
            return Nwk_ObjSetPred( pObj, nil );
        } }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns 0 if there is an unmarked path to a CI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManVerifyCut_rec(pObj *Nwk_Obj_t) int32 {
    var pNext *Nwk_Obj_t;
    var i int32;
    if pObj.MarkA {
        return 1; }
    if Nwk_ObjIsLo(pObj) {
        return 0; }
    if Nwk_ObjIsTravIdCurrent( pObj ) {
        return 1; }
    Nwk_ObjSetTravIdCurrent( pObj );
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[i]); i++  {
        if !(Nwk_ManVerifyCut_rec( pNext ) != 0) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Verifies the forward cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManRetimeVerifyCutForward(pMan *Nwk_Man_t, vNodes *Vec_Ptr_t) int32 {
    var pObj *Nwk_Obj_t;
    var i int32;
    // mark the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        assert(pObj.MarkA == 0, "pObj.MarkA == 0");
        pObj.MarkA = 1;
    }
    // traverse from the COs
    Nwk_ManIncrementTravId( pMan );
    // inlining of macro Nwk_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        if !(Nwk_ManVerifyCut_rec( pObj ) != 0) {
            fmt.Printf("Nwk_ManRetimeVerifyCutForward(): Internal cut verification failed.\n"); }
    }
    // unmark the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.MarkA = 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Verifies the forward cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManRetimeVerifyCutBackward(pMan *Nwk_Man_t, vNodes *Vec_Ptr_t) int32 {
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes minimum cut for forward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManRetimeCutForward(pMan *Nwk_Man_t, nLatches int32, fVerbose int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Nwk_Obj_t;
    var i, RetValue, Counter, Counter2 int32;
    var clk int64  = time.Now().UnixNano();
    // set the sequential parameters
    pMan.nLatches = nLatches;
    pMan.nTruePis = Nwk_ManCiNum(pMan) - nLatches;
    pMan.nTruePos = Nwk_ManCoNum(pMan) - nLatches;
    // mark the COs and the TFO of PIs
    // inlining of macro Nwk_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        pObj.MarkA = 1;
    }
    // inlining of macro Nwk_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < (pMan).nTruePis) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        Nwk_ManMarkTfoCone_rec( pObj );
    }
    // start flow computation from each LO
    Nwk_ManIncrementTravIdFlow( pMan );
    // inlining of macro Nwk_ManForEachLoSeq
    for  i = 0; (i < (pMan).nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCis, i+(pMan).nTruePis)), 1); i++  {
        if !(Nwk_ManPushForwardFast_rec( pObj, nil ) != 0) {
            continue; }
        Nwk_ManIncrementTravIdFlow( pMan );
        Counter++;
    }
    if fVerbose != 0 {
    fmt.Printf("Forward:  Max-flow = %4d -> ", Counter); }
    // continue flow computation from each LO
    DepthFwdMax = DepthFwd = 0;
    Nwk_ManIncrementTravIdFlow( pMan );
    // inlining of macro Nwk_ManForEachLoSeq
    for  i = 0; (i < (pMan).nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCis, i+(pMan).nTruePis)), 1); i++  {
    fmt.Printf("%d ", DepthFwdMax);
        if !(Nwk_ManPushForwardBot_rec( pObj, nil ) != 0) {
            continue; }
        assert(DepthFwd == 0, "DepthFwd == 0");
        Nwk_ManIncrementTravIdFlow( pMan );
        Counter2++;
    }
    fmt.Printf("DepthMax = %d.\n", DepthFwdMax);
    if fVerbose != 0 {
    fmt.Printf("%4d.  ", Counter+Counter2); }
    // repeat flow computation from each LO
    if Counter2 > 0 {
        Nwk_ManIncrementTravIdFlow( pMan );
        // inlining of macro Nwk_ManForEachLoSeq
        for  i = 0; (i < (pMan).nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCis, i+(pMan).nTruePis)), 1); i++  {
            RetValue = Nwk_ManPushForwardBot_rec( pObj, nil );
            assert(!RetValue, "!RetValue");
        }
    }
    // cut is a set of nodes whose bottom is visited but top is not visited
    vNodes = Vec_PtrAlloc( Counter+Counter2 );
    Counter = 0;
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        if Nwk_ObjVisitedBotOnly(pObj) != 0 {
            assert((Nwk_ObjHasFlow(pObj) != 0), "Nwk_ObjHasFlow(pObj)");
            assert(!Nwk_ObjIsCo(pObj), "!Nwk_ObjIsCo(pObj)");
            Vec_PtrPush( vNodes, pObj );
            Counter += Nwk_ObjIsCi(pObj);
        }
    }
    }
    Nwk_ManCleanMarks( pMan );
    assert((Nwk_ManRetimeVerifyCutForward(pMan, vNodes) != 0), "Nwk_ManRetimeVerifyCutForward(pMan, vNodes)");
    if fVerbose != 0 {
    fmt.Printf("Min-cut = %4d.  Unmoved = %4d. ", Vec_PtrSize(vNodes), Counter);
    PRT( "Time", time.Now().UnixNano() - clk );
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Computes minimum cut for backward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManRetimeCutBackward(pMan *Nwk_Man_t, nLatches int32, fVerbose int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Nwk_Obj_t;
    var i, RetValue, Counter, Counter2 int32;
    var clk int64  = time.Now().UnixNano();
    // set the sequential parameters
    pMan.nLatches = nLatches;
    pMan.nTruePis = Nwk_ManCiNum(pMan) - nLatches;
    pMan.nTruePos = Nwk_ManCoNum(pMan) - nLatches;
    // mark the CIs, the TFI of POs, and the constant nodes
    // inlining of macro Nwk_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        pObj.MarkA = 1;
    }
    // inlining of macro Nwk_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < (pMan).nTruePos) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        Nwk_ManMarkTfiCone_rec( pObj );
    }
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  {         if (pObj) == nil || !Nwk_ObjIsNode(pObj) {} else
        if Nwk_ObjFaninNum(pObj) == 0 {
            pObj.MarkA = 1; }
    }
    // start flow computation from each LI driver
    Nwk_ManIncrementTravIdFlow( pMan );
    // inlining of macro Nwk_ManForEachLiSeq
    for  i = 0; (i < (pMan).nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCos, i+(pMan).nTruePos)), 1); i++  {
        if !(Nwk_ManPushBackwardFast_rec( Nwk_ObjFanin0(pObj), nil ) != 0) {
            continue; }
        Nwk_ManIncrementTravIdFlow( pMan );
        Counter++; 
    }
    if fVerbose != 0 { 
    fmt.Printf("Backward: Max-flow = %4d -> ", Counter); }
    // continue flow computation from each LI driver
    Nwk_ManIncrementTravIdFlow( pMan );
    // inlining of macro Nwk_ManForEachLiSeq
    for  i = 0; (i < (pMan).nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCos, i+(pMan).nTruePos)), 1); i++  {
        if !(Nwk_ManPushBackwardBot_rec( Nwk_ObjFanin0(pObj), nil ) != 0) {
            continue; }
        Nwk_ManIncrementTravIdFlow( pMan );
        Counter2++;
    }
    if fVerbose != 0 {
    fmt.Printf("%4d.  ", Counter+Counter2); }
    // repeat flow computation from each LI driver
    if Counter2 > 0 {
        Nwk_ManIncrementTravIdFlow( pMan );
        // inlining of macro Nwk_ManForEachLiSeq
        for  i = 0; (i < (pMan).nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCos, i+(pMan).nTruePos)), 1); i++  {
            RetValue = Nwk_ManPushBackwardBot_rec( Nwk_ObjFanin0(pObj), nil );
            assert(!RetValue, "!RetValue");
        }
    }
    // cut is a set of nodes whose bottom is visited but top is not visited
    vNodes = Vec_PtrAlloc( Counter+Counter2 );
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        if Nwk_ObjVisitedBotOnly(pObj) != 0 {
            assert((Nwk_ObjHasFlow(pObj) != 0), "Nwk_ObjHasFlow(pObj)");
            assert(!Nwk_ObjIsCo(pObj), "!Nwk_ObjIsCo(pObj)");
            Vec_PtrPush( vNodes, pObj );
        }
    }
    }
    // count CO drivers
    Counter = 0;
    // inlining of macro Nwk_ManForEachLiSeq
    for  i = 0; (i < (pMan).nLatches) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pMan.vCos, i+(pMan).nTruePos)), 1); i++  {
        if Nwk_ObjVisitedBotOnly( Nwk_ObjFanin0(pObj) ) != 0 {
            Counter++; }
    }
    Nwk_ManCleanMarks( pMan );
    assert((Nwk_ManRetimeVerifyCutBackward(pMan, vNodes) != 0), "Nwk_ManRetimeVerifyCutBackward(pMan, vNodes)");
    if fVerbose != 0 {
    fmt.Printf("Min-cut = %4d.  Unmoved = %4d. ", Vec_PtrSize(vNodes), Counter);
    PRT( "Time", time.Now().UnixNano() - clk );
    }
    return vNodes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

