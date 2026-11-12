package ret

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [retFlow.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Implementation of maximum flow (min-area retiming).]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Oct 31, 2006.]

  Revision    [$Id: retFlow.c,v 1.00 2006/10/31 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "retInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Abc_ObjSetPath(pObj *Abc_Obj_t, pNext *Abc_Obj_t) int32 { pObj.pCopy = pNext; return 1;   }
func Abc_ObjGetPath(pObj *Abc_Obj_t) *Abc_Obj_t { return pObj.pCopy;              }
func Abc_ObjGetFanoutPath(pObj *Abc_Obj_t) *Abc_Obj_t { 
    var pFanout *Abc_Obj_t;
    var i int32;
    assert((Abc_ObjGetPath(pObj) != nil), "Abc_ObjGetPath(pObj)"); 
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if Abc_ObjGetPath(pFanout) == pObj {
            return pFanout; }
    }
    return nil;
}
func Abc_ObjGetFaninPath(pObj *Abc_Obj_t) *Abc_Obj_t { 
    var pFanin *Abc_Obj_t;
    var i int32;
    assert((Abc_ObjGetPath(pObj) != nil), "Abc_ObjGetPath(pObj)"); 
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if Abc_ObjGetPath(pFanin) == pObj {
            return pFanin; }
    }
    return nil;
}
func Abc_ObjGetPredecessorBwd(pObj *Abc_Obj_t) *Abc_Obj_t { 
    var pNext *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if Abc_ObjGetPath(pNext) == pObj {
            return pNext; }
    }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if Abc_ObjGetPath(pNext) == pObj {
            return pNext; }
    }
    return nil;
}
func Abc_ObjGetPredecessorFwd(pObj *Abc_Obj_t) *Abc_Obj_t { 
    var pNext *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if Abc_ObjGetPath(pNext) == pObj {
            return pNext; }
    }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if Abc_ObjGetPath(pNext) == pObj {
            return pNext; }
    }
    return nil;
}

// static int         Abc_NtkMaxFlowBwdPath_rec( Abc_Obj_t * pObj );

// static int         Abc_NtkMaxFlowFwdPath_rec( Abc_Obj_t * pObj );

// static int         Abc_NtkMaxFlowBwdPath2_rec( Abc_Obj_t * pObj );

// static int         Abc_NtkMaxFlowFwdPath2_rec( Abc_Obj_t * pObj );

//static int         Abc_NtkMaxFlowBwdPath3_rec( Abc_Obj_t * pObj );
// static int         Abc_NtkMaxFlowFwdPath3_rec( Abc_Obj_t * pObj, Abc_Obj_t * pPrev, int fFanin );

// static Vec_Ptr_t * Abc_NtkMaxFlowMinCut( Abc_Ntk_t * pNtk, int fForward );

// static void        Abc_NtkMaxFlowMinCutUpdate( Abc_Ntk_t * pNtk, Vec_Ptr_t * vMinCut, int fForward );

// static int         Abc_NtkMaxFlowVerifyCut( Abc_Ntk_t * pNtk, Vec_Ptr_t * vMinCut, int fForward );

// static void        Abc_NtkMaxFlowPrintCut( Abc_Ntk_t * pNtk, Vec_Ptr_t * vMinCut );

// static void        Abc_NtkMaxFlowPrintFlow( Abc_Ntk_t * pNtk, int fForward );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Test-bench for the max-flow computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowTest(pNtk *Abc_Ntk_t) {
    var vMinCut *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;

    // forward flow
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pObj.fMarkA = 1;
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        pObj.fMarkA = Abc_ObjFanin0(pObj).fMarkA = 1; }
    }
//        Abc_ObjFanin0(pObj)->fMarkA = 1;
    vMinCut = Abc_NtkMaxFlow( pNtk, 1, 1 );
    Vec_PtrFree( vMinCut );
    Abc_NtkCleanMarkA( pNtk );

    // backward flow
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pObj.fMarkA = 1;
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        pObj.fMarkA = Abc_ObjFanout0(pObj).fMarkA = 1; }
    }
//        Abc_ObjFanout0(pObj)->fMarkA = 1;
    vMinCut = Abc_NtkMaxFlow( pNtk, 0, 1 );
    Vec_PtrFree( vMinCut );
    Abc_NtkCleanMarkA( pNtk );

}

/**Function*************************************************************

  Synopsis    [Implementation of max-flow/min-cut computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlow(pNtk *Abc_Ntk_t, fForward int32, fVerbose int32) *Vec_Ptr_t {
    var vMinCut *Vec_Ptr_t;
    var pLatch *Abc_Obj_t;
    var Flow, FlowCur, RetValue, i int32;
    var clk abctime  = Abc_Clock();
    var fUseDirectedFlow int32  = 1;

    // find the max-flow
    Abc_NtkCleanCopy( pNtk );
    Flow = 0;
    Abc_NtkIncrementTravId(pNtk);
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
        if fForward != 0 {
//            assert( !Abc_ObjFanout0(pLatch)->fMarkA );
            FlowCur  = Abc_NtkMaxFlowFwdPath2_rec( Abc_ObjFanout0(pLatch) );
//            FlowCur  = Abc_NtkMaxFlowFwdPath3_rec( Abc_ObjFanout0(pLatch), pLatch, 1 );
            Flow    += FlowCur;
        
        } else {
            assert(!Abc_ObjFanin0(pLatch).fMarkA, "!Abc_ObjFanin0(pLatch).fMarkA");
            FlowCur  = Abc_NtkMaxFlowBwdPath2_rec( Abc_ObjFanin0(pLatch) );
            Flow    += FlowCur;
        }
        if FlowCur {
            Abc_NtkIncrementTravId(pNtk); }
    }
    }

    if !(fUseDirectedFlow != 0) {
        Abc_NtkIncrementTravId(pNtk);
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
            if fForward != 0 {
    //            assert( !Abc_ObjFanout0(pLatch)->fMarkA );
                FlowCur  = Abc_NtkMaxFlowFwdPath_rec( Abc_ObjFanout0(pLatch) );
                Flow    += FlowCur;
            
            } else {
                assert(!Abc_ObjFanin0(pLatch).fMarkA, "!Abc_ObjFanin0(pLatch).fMarkA");
                FlowCur  = Abc_NtkMaxFlowBwdPath_rec( Abc_ObjFanin0(pLatch) );
                Flow    += FlowCur;
            }
            if FlowCur {
                Abc_NtkIncrementTravId(pNtk); }
        }
        }
    }
//    Abc_NtkMaxFlowPrintFlow( pNtk, fForward );

    // mark the nodes reachable from the latches
    Abc_NtkIncrementTravId(pNtk);
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
        if fForward != 0 {
//            assert( !Abc_ObjFanout0(pLatch)->fMarkA );
            if fUseDirectedFlow != 0 {
                RetValue = Abc_NtkMaxFlowFwdPath2_rec( Abc_ObjFanout0(pLatch) );
//                RetValue = Abc_NtkMaxFlowFwdPath3_rec( Abc_ObjFanout0(pLatch), pLatch, 1 );
            } else {
                RetValue = Abc_NtkMaxFlowFwdPath_rec( Abc_ObjFanout0(pLatch) ); }
        
        } else {
            assert(!Abc_ObjFanin0(pLatch).fMarkA, "!Abc_ObjFanin0(pLatch).fMarkA");
            if fUseDirectedFlow != 0 {
                RetValue = Abc_NtkMaxFlowBwdPath2_rec( Abc_ObjFanin0(pLatch) );
            } else {
                RetValue = Abc_NtkMaxFlowBwdPath_rec( Abc_ObjFanin0(pLatch) ); }
        }
        assert(RetValue == 0, "RetValue == 0");
    }
    }

    // find the min-cut with the smallest volume
    vMinCut = Abc_NtkMaxFlowMinCut( pNtk, fForward );
    // verify the cut
    if !(Abc_NtkMaxFlowVerifyCut(pNtk, vMinCut, fForward) != 0) {
        fmt.Printf("Abc_NtkMaxFlow() error! The computed min-cut is not a cut!\n"); }
    // make the cut retimable
    Abc_NtkMaxFlowMinCutUpdate( pNtk, vMinCut, fForward );

    // report the results
    if fVerbose != 0 {
    fmt.Printf("L = %6d. %s max-flow = %6d.  Min-cut = %6d.  ", Abc_NtkLatchNum(pNtk), fForward? "Forward " : "Backward", Flow, Vec_PtrSize(vMinCut));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

//    Abc_NtkMaxFlowPrintCut( pNtk, vMinCut );
    return vMinCut;
}

/**Function*************************************************************

  Synopsis    [Tries to find an augmenting path originating in this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowBwdPath_rec(pObj *Abc_Obj_t) int32 {
    var pNext, pPred *Abc_Obj_t;
    var i int32;
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 0; }
    Abc_NodeSetTravIdCurrent(pObj);
    // get the predecessor
    pPred = Abc_ObjGetPredecessorBwd( pObj );
    // process node without flow
    if !(Abc_ObjGetPath(pObj) != nil) {
        // start the path if we reached a terminal node
        if pObj.fMarkA {
            return Abc_ObjSetPath( pObj, (Abc_Obj_t *)1 ); }
        // explore the fanins
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
            if pNext != pPred && !Abc_ObjIsLatch(pNext) && (Abc_NtkMaxFlowBwdPath_rec(pNext) != 0) {
                return Abc_ObjSetPath( pObj, pNext ); }
        }
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
            if pNext != pPred && !Abc_ObjIsLatch(pNext) && (Abc_NtkMaxFlowBwdPath_rec(pNext) != 0) {
                return Abc_ObjSetPath( pObj, pNext ); }
        }
        return 0;
    }
    // pObj has flow - find the fanout with flow
    if pPred == nil {
        return 0; }
    // go through the successors of the predecessor
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pPred)) && (((pNext) = Abc_ObjFanin(pPred, i)), 1); i++  {
        if !Abc_ObjIsLatch(pNext) && (Abc_NtkMaxFlowBwdPath_rec( pNext ) != 0) {
            return Abc_ObjSetPath( pPred, pNext ); }
    }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pPred)) && (((pNext) = Abc_ObjFanout(pPred, i)), 1); i++  {
        if !Abc_ObjIsLatch(pNext) && (Abc_NtkMaxFlowBwdPath_rec( pNext ) != 0) {
            return Abc_ObjSetPath( pPred, pNext ); }
    }
    // try the fanout
    if Abc_NtkMaxFlowBwdPath_rec( pPred ) != 0 {
        return Abc_ObjSetPath( pPred, nil ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Tries to find an augmenting path originating in this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowFwdPath_rec(pObj *Abc_Obj_t) int32 {
    var pNext, pPred *Abc_Obj_t;
    var i int32;
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 0; }
    Abc_NodeSetTravIdCurrent(pObj);
    // get the predecessor
    pPred = Abc_ObjGetPredecessorFwd( pObj );
    // process node without flow
    if !(Abc_ObjGetPath(pObj) != nil) {
        // start the path if we reached a terminal node
        if pObj.fMarkA {
            return Abc_ObjSetPath( pObj, (Abc_Obj_t *)1 ); }
        // explore the fanins
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
            if pNext != pPred && !Abc_ObjIsLatch(pNext) && (Abc_NtkMaxFlowFwdPath_rec(pNext) != 0) {
                return Abc_ObjSetPath( pObj, pNext ); }
        }
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
            if pNext != pPred && !Abc_ObjIsLatch(pNext) && (Abc_NtkMaxFlowFwdPath_rec(pNext) != 0) {
                return Abc_ObjSetPath( pObj, pNext ); }
        }
        return 0;
    }
    // pObj has flow - find the fanout with flow
    if pPred == nil {
        return 0; }
    // go through the successors of the predecessor
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pPred)) && (((pNext) = Abc_ObjFanout(pPred, i)), 1); i++  {
        if !Abc_ObjIsLatch(pNext) && (Abc_NtkMaxFlowFwdPath_rec( pNext ) != 0) {
            return Abc_ObjSetPath( pPred, pNext ); }
    }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pPred)) && (((pNext) = Abc_ObjFanin(pPred, i)), 1); i++  {
        if !Abc_ObjIsLatch(pNext) && (Abc_NtkMaxFlowFwdPath_rec( pNext ) != 0) {
            return Abc_ObjSetPath( pPred, pNext ); }
    }
    // try the fanout
    if Abc_NtkMaxFlowFwdPath_rec( pPred ) != 0 {
        return Abc_ObjSetPath( pPred, nil ); }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Tries to find an augmenting path originating in this edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowFwdPath3_rec(pObj *Abc_Obj_t, pPrev *Abc_Obj_t, fFanin int32) int32 {
    var pFanin, pFanout *Abc_Obj_t;
    var i int32;
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 0; }
    Abc_NodeSetTravIdCurrent(pObj);
    // skip the fanin which already has flow
    if (fFanin != 0) && (Abc_ObjGetPath(pPrev) != nil) {
        return 0; }
    // if the node has no flow, try to push through the fanouts
    if !(Abc_ObjGetPath(pObj) != nil) {
        // start the path if we reached a terminal node
        if pObj.fMarkA {
            return Abc_ObjSetPath( pObj, (Abc_Obj_t *)1 ); }
        // try to push flow through the fanouts
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
            if Abc_NtkMaxFlowFwdPath3_rec(pFanout, pObj, 1) != 0 {
                if fFanin != 0 {
                    return Abc_ObjSetPath(pPrev, pObj);
                } else {
                    return 1;
                } }
        }
    }
    // try to push through the fanins
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if !Abc_ObjIsLatch(pFanin) && (Abc_NtkMaxFlowFwdPath3_rec(pFanin, pObj, 0) != 0) {
            return Abc_ObjSetPath( pFanin, nil ); }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Tries to find an augmenting path originating in this node.]

  Description [This procedure works for directed graphs only!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowBwdPath2_rec(pObj *Abc_Obj_t) int32 {
    var pFanout, pFanin *Abc_Obj_t;
    var i int32;
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 0; }
    Abc_NodeSetTravIdCurrent(pObj);
    // process node without flow
    if !(Abc_ObjGetPath(pObj) != nil) {
        // start the path if we reached a terminal node
        if pObj.fMarkA {
            return Abc_ObjSetPath( pObj, (Abc_Obj_t *)1 ); }
        // explore the fanins
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
            if Abc_NtkMaxFlowBwdPath2_rec(pFanin) != 0 {
                return Abc_ObjSetPath( pObj, pFanin ); }
        }
        return 0;
    }
    // pObj has flow - find the fanout with flow
    pFanout = Abc_ObjGetFanoutPath( pObj );
    if pFanout == nil {
        return 0; }
    // go through the fanins of the fanout with flow
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pFanout)) && (((pFanin) = Abc_ObjFanin(pFanout, i)), 1); i++  {
        if Abc_NtkMaxFlowBwdPath2_rec( pFanin ) != 0 {
            return Abc_ObjSetPath( pFanout, pFanin ); }
    }
    // try the fanout
    if Abc_NtkMaxFlowBwdPath2_rec( pFanout ) != 0 {
        return Abc_ObjSetPath( pFanout, nil ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Tries to find an augmenting path originating in this node.]

  Description [This procedure works for directed graphs only!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowFwdPath2_rec(pObj *Abc_Obj_t) int32 {
    var pFanout, pFanin *Abc_Obj_t;
    var i int32;
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 0; }
    Abc_NodeSetTravIdCurrent(pObj);
    // process node without flow
    if !(Abc_ObjGetPath(pObj) != nil) { 
        // start the path if we reached a terminal node
        if pObj.fMarkA {
            return Abc_ObjSetPath( pObj, (Abc_Obj_t *)1 ); }
        // explore the fanins
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
            if Abc_NtkMaxFlowFwdPath2_rec(pFanout) != 0 {
                return Abc_ObjSetPath( pObj, pFanout ); }
        }
        return 0;
    }
    // pObj has flow - find the fanout with flow
    pFanin = Abc_ObjGetFaninPath( pObj );
    if pFanin == nil {
        return 0; }
    // go through the fanins of the fanout with flow
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pFanin)) && (((pFanout) = Abc_ObjFanout(pFanin, i)), 1); i++  {
        if Abc_NtkMaxFlowFwdPath2_rec( pFanout ) != 0 {
            return Abc_ObjSetPath( pFanin, pFanout ); }
    }
    // try the fanout
    if Abc_NtkMaxFlowFwdPath2_rec( pFanin ) != 0 {
        return Abc_ObjSetPath( pFanin, nil ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Find minimum-volume minumum cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowMinCut(pNtk *Abc_Ntk_t, fForward int32) *Vec_Ptr_t {
    var vMinCut *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // collect the cut nodes
    vMinCut = Vec_PtrAlloc( Abc_NtkLatchNum(pNtk) );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        // node without flow is not a cut node
        if !(Abc_ObjGetPath(pObj) != nil) {
            continue; }
        // unvisited node is below the cut
        if !Abc_NodeIsTravIdCurrent(pObj) {
            continue; }
        // add terminal with flow or node whose path is not visited
        if pObj.fMarkA || !Abc_NodeIsTravIdCurrent( Abc_ObjGetPath(pObj) ) {
            Vec_PtrPush( vMinCut, pObj ); }
    }
    }
    return vMinCut;
}

/**Function*************************************************************

  Synopsis    [Marks the TFI cone with MarkA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowMarkCut_rec(pObj *Abc_Obj_t) {
    var pNext *Abc_Obj_t;
    var i int32;
    if pObj.fMarkA {
        return; }
    pObj.fMarkA = 1;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_NtkMaxFlowMarkCut_rec( pNext );
    }
}

/**Function*************************************************************

  Synopsis    [Visits the TFI up to marked nodes and collects marked nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowCollectCut_rec(pObj *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pNext *Abc_Obj_t;
    var i int32;
    if Abc_NodeIsTravIdCurrent(pObj) {
        return; }
    Abc_NodeSetTravIdCurrent(pObj);
    if pObj.fMarkA {
        Vec_PtrPush( vNodes, pObj );
        return;
    }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_NtkMaxFlowCollectCut_rec( pNext, vNodes );
    }
}

/**Function*************************************************************

  Synopsis    [Updates the minimum cut to be retimable.]

  Description [This procedure also labels the nodes reachable from
  the latches to the cut with fMarkA.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowMinCutUpdate(pNtk *Abc_Ntk_t, vMinCut *Vec_Ptr_t, fForward int32) {
    var pObj, pNext *Abc_Obj_t;
    var i, k int32;
    // clean marks
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        pObj.fMarkA = 0; }
    }
    // set latch outputs
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        Abc_ObjFanout0(pObj).fMarkA = 1; }
    }
    // traverse from cut nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMinCut)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vMinCut, i)), 1); i++  {
        Abc_NtkMaxFlowMarkCut_rec( pObj );
    }
    if fForward != 0 {
        // change mincut to be nodes with unmarked fanouts
        Vec_PtrClear( vMinCut );
        // inlining of macro Abc_NtkForEachObj
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
            if !pObj.fMarkA {
                continue; }
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
                if pNext.fMarkA {
                    continue; }
                Vec_PtrPush( vMinCut, pObj );
                break;
            }
        }
        }
    
    } else {
        // change mincut to be marked fanins of the unmarked nodes
        Vec_PtrClear( vMinCut );
        Abc_NtkIncrementTravId(pNtk);
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            Abc_NtkMaxFlowCollectCut_rec( Abc_ObjFanin0(pObj), vMinCut ); }
        }
        // transfer the attribute
        // inlining of macro Abc_NtkForEachObj
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
            pObj.fMarkA = Abc_NodeIsTravIdCurrent(pObj); }
        }
        // unmark the cut nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vMinCut)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vMinCut, i)), 1); i++  {
            pObj.fMarkA = 0;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Verifies the min-cut is indeed a cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowVerifyCut_rec(pObj *Abc_Obj_t, fForward int32) int32 {
    var pNext *Abc_Obj_t;
    var i int32;
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 1; }
    Abc_NodeSetTravIdCurrent(pObj);
    // visit the node
    if fForward != 0 {
        if Abc_ObjIsCo(pObj) {
            return 0; }
        // explore the fanouts
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
            if !(Abc_NtkMaxFlowVerifyCut_rec(pNext, fForward) != 0) {
                return 0; }
        }
    
    } else {
        if Abc_ObjIsCi(pObj) {
            return 0; }
        // explore the fanins
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
            if !(Abc_NtkMaxFlowVerifyCut_rec(pNext, fForward) != 0) {
                return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Verifies the min-cut is indeed a cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowVerifyCut(pNtk *Abc_Ntk_t, vMinCut *Vec_Ptr_t, fForward int32) int32 {
    var pObj *Abc_Obj_t;
    var i int32;
    // mark the cut with the current traversal ID
    Abc_NtkIncrementTravId(pNtk);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMinCut)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vMinCut, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
    }
    // search from the latches for a path to the COs/CIs
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        if fForward != 0 {
            if !(Abc_NtkMaxFlowVerifyCut_rec( Abc_ObjFanout0(pObj), fForward ) != 0) {
                return 0; }
        
        } else {
            if !(Abc_NtkMaxFlowVerifyCut_rec( Abc_ObjFanin0(pObj), fForward ) != 0) {
                return 0; }
        }
    }
    }
/*
    {
        // count the volume of the cut
        int Counter = 0;
        Abc_NtkForEachObj( pNtk, pObj, i )
            Counter += Abc_NodeIsTravIdCurrent( pObj );
        printf( "Volume = %d.\n", Counter );
    }
*/
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prints the flows.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowPrintFlow(pNtk *Abc_Ntk_t, fForward int32) {
    var pLatch, pNext *Abc_Obj_t;
    var pPrev *Abc_Obj_t  = nil; // Suppress "might be used uninitialized"
    var i int32;
    if fForward != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.vBoxes)) && (((pLatch) = (Abc_Obj_t *)Vec_PtrEntry(pNtk.vBoxes, i)), 1); i++  {
            assert(!Abc_ObjFanout0(pLatch).fMarkA, "!Abc_ObjFanout0(pLatch).fMarkA");
            if Abc_ObjGetPath(Abc_ObjFanout0(pLatch)) == nil { // no flow through this latch
                continue; }
            fmt.Printf("Path = ");
            for  pNext = Abc_ObjFanout0(pLatch); pNext != (void *)1; pNext = Abc_ObjGetPath(pNext)  {
                fmt.Printf("%s(%d) ", Abc_ObjName(pNext), pNext.Id);
                pPrev = pNext;
            }
            if !Abc_ObjIsPo(pPrev) {
            fmt.Printf("%s(%d) ", Abc_ObjName(Abc_ObjFanout0(pPrev)), Abc_ObjFanout0(pPrev).Id); }
            fmt.Printf("\n");
        }
    
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.vBoxes)) && (((pLatch) = (Abc_Obj_t *)Vec_PtrEntry(pNtk.vBoxes, i)), 1); i++  {
            assert(!Abc_ObjFanin0(pLatch).fMarkA, "!Abc_ObjFanin0(pLatch).fMarkA");
            if Abc_ObjGetPath(Abc_ObjFanin0(pLatch)) == nil { // no flow through this latch
                continue; }
            fmt.Printf("Path = ");
            for  pNext = Abc_ObjFanin0(pLatch); pNext != (void *)1; pNext = Abc_ObjGetPath(pNext)  {
                fmt.Printf("%s(%d) ", Abc_ObjName(pNext), pNext.Id);
                pPrev = pNext;
            }
            if !Abc_ObjIsPi(pPrev) {
            fmt.Printf("%s(%d) ", Abc_ObjName(Abc_ObjFanin0(pPrev)), Abc_ObjFanin0(pPrev).Id); }
            fmt.Printf("\n");
        }
    }
}

/**Function*************************************************************

  Synopsis    [Prints the min-cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMaxFlowPrintCut(pNtk *Abc_Ntk_t, vMinCut *Vec_Ptr_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    fmt.Printf("Min-cut: ");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMinCut)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vMinCut, i)), 1); i++  {
        fmt.Printf("%s(%d) ", Abc_ObjName(pObj), pObj.Id);
    }
    fmt.Printf("\n");
    fmt.Printf("Marked nodes: ");
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if pObj.fMarkA {
            fmt.Printf("%s(%d) ", Abc_ObjName(pObj), pObj.Id); }
    }
    fmt.Printf("\n");
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

