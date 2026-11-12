package scl

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [sclUpsize.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Selective increase of gate sizes.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclUpsize.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclSize.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collect TFO of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindTFO_rec(pObj *Abc_Obj_t, vNodes *Vec_Int_t, vCos *Vec_Int_t) {
    var pNext *Abc_Obj_t;
    var i int32;
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    if Abc_ObjIsCo(pObj) {
        Vec_IntPush( vCos, Abc_ObjId(pObj) );
        return;
    }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Abc_SclFindTFO_rec( pNext, vNodes, vCos );
    }
    if Abc_ObjFaninNum(pObj) > 0 {
        Vec_IntPush( vNodes, Abc_ObjId(pObj) ); }
}
func Abc_SclFindTFO(p *Abc_Ntk_t, vPath *Vec_Int_t) *Vec_Int_t {
    var vNodes, vCos *Vec_Int_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    assert(Vec_IntSize(vPath) > 0, "Vec_IntSize(vPath) > 0");
    vCos = Vec_IntAlloc( 100 );
    vNodes = Vec_IntAlloc( 100 );
    // collect nodes in the TFO
    Abc_NtkIncrementTravId( p ); 
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vPath) && (((pObj) = Abc_NtkObj(p, Vec_IntEntry(vPath,i))), 1); i++  {         if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if Abc_ObjIsNode(pFanin) {
                Abc_SclFindTFO_rec( pFanin, vNodes, vCos ); }
        } }
    }
    // reverse order
    Vec_IntReverseOrder( vNodes );
//    Vec_IntSort( vNodes, 0 );
//Vec_IntPrint( vNodes );
//Vec_IntPrint( vCos );
    Vec_IntAppend( vNodes, vCos );
    Vec_IntFree( vCos );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Collect near-critical COs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindCriticalCoWindow(p *SC_Man, Window int32) *Vec_Int_t {
    var fMaxArr float32  = Abc_SclReadMaxDelay( p ) * (100.0 - Window) / 100.0;
    var vPivots *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i int32;
    vPivots = Vec_IntAlloc( 100 );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        if Abc_SclObjTimeMax(p, pObj) >= fMaxArr {
            Vec_IntPush( vPivots, Abc_ObjId(pObj) ); }
    }
    assert(Vec_IntSize(vPivots) > 0, "Vec_IntSize(vPivots) > 0");
    return vPivots;
}

/**Function*************************************************************

  Synopsis    [Collect near-critical internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindCriticalNodeWindow_rec(p *SC_Man, pObj *Abc_Obj_t, vPath *Vec_Int_t, fSlack float32, fDept int32) {
    var pNext *Abc_Obj_t;
    var fArrMax, fSlackFan float32;
    var i int32;
    if Abc_ObjIsCi(pObj) {
        return; }
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // compute the max arrival time of the fanins
    if fDept != 0 {
//        fArrMax = p->pSlack[Abc_ObjId(pObj)];
        fArrMax = Abc_SclObjGetSlack(p, pObj, p.MaxDelay);
    } else {
        fArrMax = Abc_SclGetMaxDelayNodeFanins( p, pObj ); }
//    assert( fArrMax >= -1 );
    fArrMax = Abc_MaxFloat( fArrMax, 0 );
    // traverse all fanins whose arrival times are within a window
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if Abc_ObjIsCi(pNext) || Abc_ObjFaninNum(pNext) == 0 {
            continue; }
        assert(Abc_ObjIsNode(pNext), "Abc_ObjIsNode(pNext)");
        if fDept != 0 {
//            fSlackFan = fSlack - (p->pSlack[Abc_ObjId(pNext)] - fArrMax);
            fSlackFan = fSlack - (Abc_SclObjGetSlack(p, pNext, p.MaxDelay) - fArrMax);
        } else {
            fSlackFan = fSlack - (fArrMax - Abc_SclObjTimeMax(p, pNext)); }
        if fSlackFan >= 0 {
            Abc_SclFindCriticalNodeWindow_rec( p, pNext, vPath, fSlackFan, fDept ); }
    }
    if Abc_ObjFaninNum(pObj) > 0 {
        Vec_IntPush( vPath, Abc_ObjId(pObj) ); }
}
func Abc_SclFindCriticalNodeWindow(p *SC_Man, vPathCos *Vec_Int_t, Window int32, fDept int32) *Vec_Int_t {
    var fMaxArr float32  = Abc_SclReadMaxDelay( p );
    var fSlackMax float32  = fMaxArr * Window / 100.0;
    var vPath *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_NtkIncrementTravId( p.pNtk ); 
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vPathCos) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vPathCos,i))), 1); i++  {         if (pObj) == nil {} else {
        var fSlackThis float32  = fSlackMax - (fMaxArr - Abc_SclObjTimeMax(p, pObj));
        if fSlackThis >= 0 {
            Abc_SclFindCriticalNodeWindow_rec( p, Abc_ObjFanin0(pObj), vPath, fSlackThis, fDept ); }
    }
    }
    // label critical nodes
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vPathCos) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vPathCos,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.fMarkA = 1; }
    }
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vPath) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vPath,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.fMarkA = 1; }
    }
    return vPath;  
}
func Abc_SclUnmarkCriticalNodeWindow(p *SC_Man, vPath *Vec_Int_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vPath) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vPath,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.fMarkA = 0; }
    }
}
func Abc_SclCountNearCriticalNodes(p *SC_Man) int32 {
    var RetValue int32;
    var vPathPos, vPathNodes *Vec_Int_t;
    vPathPos   = Abc_SclFindCriticalCoWindow( p, 5 );
    vPathNodes = Abc_SclFindCriticalNodeWindow( p, vPathPos, 5, 0 );
    RetValue   = Vec_IntSize(vPathNodes);
    Abc_SclUnmarkCriticalNodeWindow( p, vPathNodes );
    Abc_SclUnmarkCriticalNodeWindow( p, vPathPos );
    Vec_IntFree( vPathPos );
    Vec_IntFree( vPathNodes );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Find the array of nodes to be updated.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindNodesToUpdate(pPivot *Abc_Obj_t, pvNodes **Vec_Int_t, pvEvals **Vec_Int_t, pExtra *Abc_Obj_t) {
    var p *Abc_Ntk_t  = Abc_ObjNtk(pPivot);
    var pObj, pNext, pNext2 *Abc_Obj_t;
    var vNodes *Vec_Int_t  = *pvNodes;
    var vEvals *Vec_Int_t  = *pvEvals;
    var i, k int32;
    assert(Abc_ObjIsNode(pPivot), "Abc_ObjIsNode(pPivot)");
    assert(pPivot.fMarkA, "pPivot.fMarkA");
    // collect fanins, node, and fanouts
    Vec_IntClear( vNodes );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pNext) = Abc_ObjFanin(pPivot, i)), 1); i++  {
//        if ( Abc_ObjIsNode(pNext) && Abc_ObjFaninNum(pNext) > 0 )
        if Abc_ObjIsCi(pNext) || Abc_ObjFaninNum(pNext) > 0 {
            Vec_IntPush( vNodes, Abc_ObjId(pNext) ); }
    }
    Vec_IntPush( vNodes, Abc_ObjId(pPivot) );
    if pExtra != nil {
        Vec_IntPush( vNodes, Abc_ObjId(pExtra) ); }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pPivot)) && (((pNext) = Abc_ObjFanout(pPivot, i)), 1); i++  {
        if Abc_ObjIsNode(pNext) && pNext.fMarkA {
            Vec_IntPush( vNodes, Abc_ObjId(pNext) );
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pNext)) && (((pNext2) = Abc_ObjFanout(pNext, k)), 1); k++  {
                if Abc_ObjIsNode(pNext2) && pNext2.fMarkA {
                    Vec_IntPush( vNodes, Abc_ObjId(pNext2) ); }
            }
        }
    }
    Vec_IntUniqify( vNodes );
    // label nodes
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(p, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        assert(pObj.fMarkB == 0, "pObj.fMarkB == 0");
        pObj.fMarkB = 1;
    }
    }
    // collect nodes visible from the critical paths
    Vec_IntClear( vEvals );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(p, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
            if pNext.fMarkA && !pNext.fMarkB {
//            if ( !pNext->fMarkB )
            
                assert(pObj.fMarkB, "pObj.fMarkB");
                Vec_IntPush( vEvals, Abc_ObjId(pObj) );
                break;
            }
        } }
    }
    assert(Vec_IntSize(vEvals) > 0, "Vec_IntSize(vEvals) > 0");
    // label nodes
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(p, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.fMarkB = 0; }
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindBestCell(p *SC_Man, pObj *Abc_Obj_t, vRecalcs *Vec_Int_t, vEvals *Vec_Int_t, Notches int32, DelayGap int32, pGainBest *float32) int32 {
    var pCellOld, pCellNew *SC_Cell;
    var dGain, dGainBest float32;
    var k, gateBest, NoChange int32;
    // save old gate, timing, fanin load
    pCellOld = Abc_SclObjCell( pObj );
    Abc_SclConeStore( p, vRecalcs );
    Abc_SclEvalStore( p, vEvals );
    Abc_SclLoadStore( p, pObj );
    // try different gate sizes for this node
    gateBest = -1;
    dGainBest = -DelayGap;
    // inlining of macro SC_RingForEachCell
    for  k, pCellNew = 0, pCellOld; k == 0 || pCellNew != pCellOld; pCellNew, k = pCellNew.pNext, k+1  {
        if pCellNew == pCellOld {
            continue; }
        if k > Notches {
            break; }
        // set new cell
        Abc_SclObjSetCell( pObj, pCellNew );
        Abc_SclUpdateLoad( p, pObj, pCellOld, pCellNew );
        // recompute timing
        Abc_SclTimeCone( p, vRecalcs );
        // set old cell
        Abc_SclObjSetCell( pObj, pCellOld );
        Abc_SclLoadRestore( p, pObj );
        // save best gain
        dGain = Abc_SclEvalPerform( p, vEvals );
        if dGainBest < dGain {
            dGainBest = dGain;
            gateBest = pCellNew.Id;
            NoChange = 1;
        
        } else if NoChange != 0 {
            NoChange++; }
        if NoChange == 4 {
            break; }
//        printf( "%.2f ", dGain );
    }
//    printf( "Best = %.2f   ", dGainBest );
//    printf( "\n" );
    // put back old cell and timing
    Abc_SclObjSetCell( pObj, pCellOld );
    Abc_SclConeRestore( p, vRecalcs );
    *pGainBest = dGainBest;
    return gateBest;
}

/**Function*************************************************************

  Synopsis    [Computes the set of gates to upsize.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindBypasses(p *SC_Man, vPathNodes *Vec_Int_t, Ratio int32, Notches int32, iIter int32, DelayGap int32, fVeryVerbose int32) int32 {
    var pCellOld, pCellNew *SC_Cell;
    var vFanouts *Vec_Ptr_t;
    var vRecalcs, vEvals *Vec_Int_t;
    var pBuf, pFanin, pFanout, pExtra *Abc_Obj_t;
    var i, j, iNode, gateBest, gateBest2, fanBest, Counter int32;
    var dGainBest, dGainBest2 float32;

    // compute savings due to bypassing buffers
    vFanouts = Vec_PtrAlloc( 100 );
    vRecalcs = Vec_IntAlloc( 100 );
    vEvals = Vec_IntAlloc( 100 );
    Vec_QueClear( p.vNodeByGain );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vPathNodes) && (((pBuf) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vPathNodes,i))), 1); i++  {         if (pBuf) == nil {} else {
        assert(pBuf.fMarkB == 0, "pBuf.fMarkB == 0");
        if Abc_ObjFaninNum(pBuf) != 1 {
            continue; }
        pFanin = Abc_ObjFanin0(pBuf);
        if !Abc_ObjIsNode(pFanin) {
            continue; }
        pExtra = nil;
        if p.pNtk.vPhases == nil {
            if Abc_SclIsInv(pBuf) {
                if !Abc_SclIsInv(pFanin) {
                    continue; }
                pFanin = Abc_ObjFanin0(pFanin);
                if !Abc_ObjIsNode(pFanin) {
                    continue; }
                pExtra = pBuf;
                // we make pBuf and pFanin are in the same phase and pFanin is a node
            }
        }
        // here we have pBuf and its fanin pFanin, which is a logic node
        // compute nodes to recalculate timing and nodes to evaluate afterwards
        Abc_SclFindNodesToUpdate( pFanin, &vRecalcs, &vEvals, pExtra );
        assert(Vec_IntSize(vEvals) > 0, "Vec_IntSize(vEvals) > 0");
        // consider fanouts of this node
        fanBest    = -1;
        gateBest2  = -1;
        dGainBest2 =  0;
        Abc_NodeCollectFanouts( pBuf, vFanouts );
        // inlining of macro Vec_PtrForEachEntry
        for  j = 0; (j < Vec_PtrSize(vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, j)), 1); j++  {
            // skip COs
            if Abc_ObjIsCo(pFanout) {
                continue; }
            // skip non-critical fanouts
            if !pFanout.fMarkA {
                continue; }
            // skip if fanin already has fanout as a fanout
            if Abc_NodeFindFanin(pFanout, pFanin) >= 0 {
                continue; }
            // skip if fanin already has fanout as a fanout
            if (pExtra != nil) && Abc_NodeFindFanin(pFanout, pExtra) >= 0 {
                continue; }
            // prepare
            Abc_SclLoadStore3( p, pBuf );
            Abc_SclUpdateLoadSplit( p, pBuf, pFanout );
            Abc_ObjPatchFanin( pFanout, pBuf, pFanin );
            // size the fanin
            gateBest = Abc_SclFindBestCell( p, pFanin, vRecalcs, vEvals, Notches, DelayGap, &dGainBest );
            // unprepare
            Abc_SclLoadRestore3( p, pBuf );
            Abc_ObjPatchFanin( pFanout, pFanin, pBuf );
            if gateBest == -1 {
                continue; }
            // compare gain
            if dGainBest2 < dGainBest {
                dGainBest2 = dGainBest;
                gateBest2 = gateBest;
                fanBest = Abc_ObjId(pFanout);
            }
        }
        // remember savings
        if gateBest2 >= 0 {
            assert(dGainBest2 > 0.0, "dGainBest2 > 0.0");
            Vec_FltWriteEntry( p.vNode2Gain, Abc_ObjId(pBuf), dGainBest2 );
            Vec_IntWriteEntry( p.vNode2Gate, Abc_ObjId(pBuf), gateBest2 );
            Vec_QuePush( p.vNodeByGain, Abc_ObjId(pBuf) );
            Vec_IntWriteEntry( p.vBestFans, Abc_ObjId(pBuf), fanBest );
        }
//        if ( ++Counter == 17 )
//            break;
    }
    }
    Vec_PtrFree( vFanouts );
    Vec_IntFree( vRecalcs );
    Vec_IntFree( vEvals );
    if Vec_QueSize(p.vNodeByGain) == 0 {
        return 0; }
    if fVeryVerbose != 0 { 
        fmt.Printf("\n"); }

    // accept changes for that are half above the average and do not overlap
    Counter = 0;
    dGainBest2 = -1;
    vFanouts = Vec_PtrAlloc( 100 );
    for  Vec_QueSize(p.vNodeByGain)  {
        iNode   = Vec_QuePop(p.vNodeByGain);
        pFanout = Abc_NtkObj( p.pNtk, Vec_IntEntry(p.vBestFans, iNode) );
        pBuf    = Abc_NtkObj( p.pNtk, iNode );
        pFanin  = Abc_ObjFanin0(pBuf);
        if pFanout.fMarkB || pBuf.fMarkB {
            continue; }
        if p.pNtk.vPhases == nil {
            // update fanin
            if Abc_SclIsInv(pBuf) {
                if !Abc_SclIsInv(pFanin) {
                    assert(false, "0");
                    continue;
                }
                pFanin = Abc_ObjFanin0(pFanin);
                if !Abc_ObjIsNode(pFanin) {
                    assert(false, "0");
                    continue;
                }
            }
        }
        if pFanin.fMarkB {
            continue; }
        pFanout.fMarkB = 1;
        pBuf.fMarkB = 1;
        pFanin.fMarkB = 1;
        Vec_PtrPush( vFanouts, pFanout );
        Vec_PtrPush( vFanouts, pBuf );
        Vec_PtrPush( vFanouts, pFanin );
        // remember gain
        if dGainBest2 == -1 {
            dGainBest2 = Vec_FltEntry(p.vNode2Gain, iNode); }
//        else if ( dGainBest2 > 2*Vec_FltEntry(p->vNode2Gain, iNode) )
//            break;
        // redirect
        Abc_SclUpdateLoadSplit( p, pBuf, pFanout );
        Abc_SclAddWireLoad( p, pBuf, 1 );
        Abc_SclAddWireLoad( p, pFanin, 1 );
        Abc_ObjPatchFanin( pFanout, pBuf, pFanin );
        Abc_SclAddWireLoad( p, pBuf, 0 );
        Abc_SclAddWireLoad( p, pFanin, 0 );
        Abc_SclTimeIncUpdateLevel( pFanout );
        // remember
        Vec_IntPush( p.vUpdates2, Abc_ObjId(pFanout) );
        Vec_IntPush( p.vUpdates2, Abc_ObjId(pFanin) );
        Vec_IntPush( p.vUpdates2, Abc_ObjId(pBuf) );
        // update cell
        pCellOld = Abc_SclObjCell( pFanin );
        pCellNew = SC_LibCell( p.pLib, Vec_IntEntry(p.vNode2Gate, iNode) );
        p.SumArea += pCellNew.area - pCellOld.area;
        Abc_SclObjSetCell( pFanin, pCellNew );
        Abc_SclUpdateLoad( p, pFanin, pCellOld, pCellNew );
        // record the update
        Vec_IntPush( p.vUpdates, Abc_ObjId(pFanin) );
        Vec_IntPush( p.vUpdates, pCellNew.Id );
        Abc_SclTimeIncInsert( p, pFanout );
        Abc_SclTimeIncInsert( p, pBuf );
        Abc_SclTimeIncInsert( p, pFanin );
        // remember when this node was upsized
        Vec_IntWriteEntry( p.vNodeIter, Abc_ObjId(pFanout), -1 );
        Vec_IntWriteEntry( p.vNodeIter, Abc_ObjId(pBuf), -1 );
        Vec_IntWriteEntry( p.vNodeIter, Abc_ObjId(pFanin), -1 );
        // update polarity
        if p.pNtk.vPhases && Abc_SclIsInv(pBuf) {
            Abc_NodeInvUpdateObjFanoutPolarity( pFanin, pFanout ); }
        // report
        if fVeryVerbose != 0 {
            fmt.Printf("Node %6d  Redir fanout %6d to fanin %6d.  Gain = %7.1f ps. ", Abc_ObjId(pBuf), Abc_ObjId(pFanout), Abc_ObjId(pFanin), Vec_FltEntry(p.vNode2Gain, iNode));
            fmt.Printf("Gate %12s (%2d/%2d)  -> %12s (%2d/%2d) \n", pCellOld.pName, pCellOld.Order, pCellOld.nGates, pCellNew.pName, pCellNew.Order, pCellNew.nGates);
        }
/*
        // check if the node became useless
        if ( Abc_ObjFanoutNum(pBuf) == 0 )
        {
            pCellOld = Abc_SclObjCell( pBuf );
            p->SumArea -= pCellOld->area;
            Abc_NtkDeleteObj_rec( pBuf, 1 );
            printf( "Removed node %d.\n", iNode );
        }
*/
        Counter++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  j = 0; (j < Vec_PtrSize(vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, j)), 1); j++  {
        pFanout.fMarkB = 0;
    }
    Vec_PtrFree( vFanouts );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Check marked fanin/fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Abc_SclObjCheckMarkedFanFans(pObj *Abc_Obj_t) int32 {
    var pNext *Abc_Obj_t;
    var i int32;
    if pObj.fMarkB {
        return 1; }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if pNext.fMarkB {
            return 1; }
    }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if pNext.fMarkB {
            return 1; }
    }
    return 0;
}
func Abc_SclObjMarkFanFans(pObj *Abc_Obj_t, vNodes *Vec_Ptr_t) {
//    Abc_Obj_t * pNext;
//    int i;
    if pObj.fMarkB == 0 {
        Vec_PtrPush( vNodes, pObj );
        pObj.fMarkB = 1;
    }
/*
    Abc_ObjForEachFanin( pObj, pNext, i )
        if ( pNext->fMarkB == 0 )
        {
            Vec_PtrPush( vNodes, pNext );
            pNext->fMarkB = 1;
        }
    Abc_ObjForEachFanout( pObj, pNext, i )
        if ( pNext->fMarkB == 0 )
        {
            Vec_PtrPush( vNodes, pNext );
            pNext->fMarkB = 1;
        }
*/
}

/**Function*************************************************************

  Synopsis    [Computes the set of gates to upsize.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindUpsizes(p *SC_Man, vPathNodes *Vec_Int_t, Ratio int32, Notches int32, iIter int32, DelayGap int32, fMoreConserf int32) int32 {
    var pCellOld, pCellNew *SC_Cell;
    var vRecalcs, vEvals *Vec_Int_t;
    var vFanouts *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var dGainBest, dGainBest2 float32;
    var i, gateBest, Limit, Counter, iIterLast int32;

    // compute savings due to upsizing each node
    vRecalcs = Vec_IntAlloc( 100 );
    vEvals = Vec_IntAlloc( 100 );
    Vec_QueClear( p.vNodeByGain );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vPathNodes) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vPathNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        assert(pObj.fMarkB == 0, "pObj.fMarkB == 0");
        iIterLast = Vec_IntEntry(p.vNodeIter, Abc_ObjId(pObj));
        if iIterLast >= 0 && iIterLast + 5 > iIter {
            continue; }
        // compute nodes to recalculate timing and nodes to evaluate afterwards
        Abc_SclFindNodesToUpdate( pObj, &vRecalcs, &vEvals, nil );
        assert(Vec_IntSize(vEvals) > 0, "Vec_IntSize(vEvals) > 0");
        //printf( "%d -> %d\n", Vec_IntSize(vRecalcs), Vec_IntSize(vEvals) );
        gateBest = Abc_SclFindBestCell( p, pObj, vRecalcs, vEvals, Notches, DelayGap, &dGainBest );
        // remember savings
        if gateBest >= 0 {
            assert(dGainBest > 0.0, "dGainBest > 0.0");
            Vec_FltWriteEntry( p.vNode2Gain, Abc_ObjId(pObj), dGainBest );
            Vec_IntWriteEntry( p.vNode2Gate, Abc_ObjId(pObj), gateBest );
            Vec_QuePush( p.vNodeByGain, Abc_ObjId(pObj) );
        }
    }
    }
    Vec_IntFree( vRecalcs );
    Vec_IntFree( vEvals );
    if Vec_QueSize(p.vNodeByGain) == 0 {
        return 0; }
/*
    Limit = Abc_MinInt( Vec_QueSize(p->vNodeByGain), Abc_MaxInt((int)(0.01 * Ratio * Vec_IntSize(vPathNodes)), 1) ); 
    //printf( "\nSelecting %d out of %d\n", Limit, Vec_QueSize(p->vNodeByGain) );
    for ( i = 0; i < Limit; i++ )
    {
        // get the object
        pObj = Abc_NtkObj( p->pNtk, Vec_QuePop(p->vNodeByGain) );
        assert( pObj->fMarkA );
        // find old and new gates
        pCellOld = Abc_SclObjCell( pObj );
        pCellNew = SC_LibCell( p->pLib, Vec_IntEntry(p->vNode2Gate, Abc_ObjId(pObj)) );
        assert( pCellNew != NULL );
        //printf( "%6d  %20s -> %20s  ", Abc_ObjId(pObj), pCellOld->pName, pCellNew->pName );
        //printf( "gain is %f\n", Vec_FltEntry(p->vNode2Gain, Abc_ObjId(pObj)) );
        // update gate
        Abc_SclUpdateLoad( p, pObj, pCellOld, pCellNew );
        p->SumArea += pCellNew->area - pCellOld->area;
        Abc_SclObjSetCell( pObj, pCellNew );
        // record the update
        Vec_IntPush( p->vUpdates, Abc_ObjId(pObj) );
        Vec_IntPush( p->vUpdates, pCellNew->Id );
        Abc_SclTimeIncInsert( p, pObj );
        // remember when this node was upsized
        Vec_IntWriteEntry( p->vNodeIter, Abc_ObjId(pObj), iIter );
    }
return Limit;
*/

    Limit = Abc_MinInt( Vec_QueSize(p.vNodeByGain), Abc_MaxInt(int32((0.01 * Ratio * Vec_IntSize(vPathNodes))), 1) ); 
    dGainBest2 = -1;
    Counter = 0;
    vFanouts = Vec_PtrAlloc( 100 );
    for  Vec_QueSize(p.vNodeByGain)  {
        var iNode int32  = Vec_QuePop(p.vNodeByGain);
        var pObj *Abc_Obj_t  = Abc_NtkObj( p.pNtk, iNode );
        assert(pObj.fMarkA, "pObj.fMarkA");
        if Abc_SclObjCheckMarkedFanFans( pObj ) != 0 {
            continue; }
        Abc_SclObjMarkFanFans( pObj, vFanouts );
        // remember gain
        if dGainBest2 == -1 {
            dGainBest2 = Vec_FltEntry(p.vNode2Gain, iNode); }
//        else if ( dGainBest2 > 3*Vec_FltEntry(p->vNode2Gain, iNode) )
//            break;
//        printf( "%.1f ", Vec_FltEntry(p->vNode2Gain, iNode) );

        // find old and new gates
        pCellOld = Abc_SclObjCell( pObj );
        pCellNew = SC_LibCell( p.pLib, Vec_IntEntry(p.vNode2Gate, Abc_ObjId(pObj)) );
        assert(pCellNew != nil, "pCellNew != nil");
        //printf( "%6d  %20s -> %20s  ", Abc_ObjId(pObj), pCellOld->pName, pCellNew->pName );
        //printf( "gain is %f\n", Vec_FltEntry(p->vNode2Gain, Abc_ObjId(pObj)) );
//        if ( pCellOld->Order > 0 )
//            printf( "%.2f  %d -> %d(%d)   ", Vec_FltEntry(p->vNode2Gain, iNode), pCellOld->Order, pCellNew->Order, pCellNew->nGates );
        // update gate
        p.SumArea += pCellNew.area - pCellOld.area;
        Abc_SclObjSetCell( pObj, pCellNew );
        Abc_SclUpdateLoad( p, pObj, pCellOld, pCellNew );
        // record the update
        Vec_IntPush( p.vUpdates, Abc_ObjId(pObj) );
        Vec_IntPush( p.vUpdates, pCellNew.Id );
        Abc_SclTimeIncInsert( p, pObj );
        // remember when this node was upsized
        Vec_IntWriteEntry( p.vNodeIter, Abc_ObjId(pObj), iIter );
        Counter++;
        if Counter == Limit {
            break; }
    }
//    printf( "\n" );

    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanouts)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
        pObj.fMarkB = 0;
    }
    Vec_PtrFree( vFanouts );
    return Counter;
}
func Abc_SclApplyUpdateToBest(vGatesBest *Vec_Int_t, vGates *Vec_Int_t, vUpdate *Vec_Int_t) {
    var i, ObjId, GateId, GateId2 int32; 
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vUpdate)) && (((ObjId) = Vec_IntEntry(vUpdate, i)), 1) && (((GateId) = Vec_IntEntry(vUpdate, i+1)), 1); i += 2  {
        Vec_IntWriteEntry( vGatesBest, ObjId, GateId );
    }
    Vec_IntClear( vUpdate );
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vGatesBest)) && (((GateId) = Vec_IntEntry(vGatesBest, i)), 1) && (((GateId2) = Vec_IntEntry(vGates, i)), 1); i++  {
        assert(GateId == GateId2, "GateId == GateId2");
    }
//    Vec_IntClear( vGatesBest );
//    Vec_IntAppend( vGatesBest, vGates );
}
func Abc_SclUndoRecentChanges(pNtk *Abc_Ntk_t, vTrans *Vec_Int_t) {
    var i int32;
    assert(Vec_IntSize(vTrans) % 3 == 0, "Vec_IntSize(vTrans) % 3 == 0");
    for  i = Vec_IntSize(vTrans)/3 - 1; i >= 0; i--  {
        var pFanout *Abc_Obj_t  = Abc_NtkObj( pNtk, Vec_IntEntry(vTrans, 3*i+0) );
        var pFanin *Abc_Obj_t   = Abc_NtkObj( pNtk, Vec_IntEntry(vTrans, 3*i+1) );
        var pObj *Abc_Obj_t     = Abc_NtkObj( pNtk, Vec_IntEntry(vTrans, 3*i+2) );
        // we do not update load here because times will be recomputed
        Abc_ObjPatchFanin( pFanout, pFanin, pObj );
        Abc_SclTimeIncUpdateLevel( pFanout );
//        printf( "Node %6d  Redir fanout %6d from fanin %6d. \n", 
//            Abc_ObjId(pObj), Abc_ObjId(pFanout), Abc_ObjId(pFanin) );
        // update polarity
        if pNtk.vPhases && Abc_SclIsInv(pObj) {
            Abc_NodeInvUpdateObjFanoutPolarity( pObj, pFanout ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclUpsizePrintDiffs(p *SC_Man, pLib *SC_Lib, pNtk *Abc_Ntk_t) {
    var fDiff float32  = float32(0.001);
    var k int32;
    var pObj *Abc_Obj_t;

    // inlining of macro ABC_ALLOC
    var pTimes *SC_Pair  = ((SC_Pair *) malloc(sizeof(SC_Pair) * (size_t)(p.nObjs)));
    // inlining of macro ABC_ALLOC
    var pSlews *SC_Pair  = ((SC_Pair *) malloc(sizeof(SC_Pair) * (size_t)(p.nObjs)));
    // inlining of macro ABC_ALLOC
    var pLoads *SC_Pair  = ((SC_Pair *) malloc(sizeof(SC_Pair) * (size_t)(p.nObjs)));

    memcpy( pTimes, p.pTimes, sizeof(SC_Pair) * p.nObjs );
    memcpy( pSlews, p.pSlews, sizeof(SC_Pair) * p.nObjs );
    memcpy( pLoads, p.pLoads, sizeof(SC_Pair) * p.nObjs );

    Abc_SclTimeNtkRecompute( p, nil, nil, 0, 0 );

    // inlining of macro Abc_NtkForEachNode
    for  k = 0; (k < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, k)), 1); k++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if Abc_AbsFloat(p.pLoads[k].rise - pLoads[k].rise) > fDiff {
            fmt.Printf("%6d : load rise differs %12.6f   %f %f\n", k, p.pLoads[k].rise-pLoads[k].rise, p.pLoads[k].rise, pLoads[k].rise); }
        if Abc_AbsFloat(p.pLoads[k].fall - pLoads[k].fall) > fDiff {
            fmt.Printf("%6d : load fall differs %12.6f   %f %f\n", k, p.pLoads[k].fall-pLoads[k].fall, p.pLoads[k].fall, pLoads[k].fall); }

        if Abc_AbsFloat(p.pSlews[k].rise - pSlews[k].rise) > fDiff {
            fmt.Printf("%6d : slew rise differs %12.6f   %f %f\n", k, p.pSlews[k].rise-pSlews[k].rise, p.pSlews[k].rise, pSlews[k].rise); }
        if Abc_AbsFloat(p.pSlews[k].fall - pSlews[k].fall) > fDiff {
            fmt.Printf("%6d : slew fall differs %12.6f   %f %f\n", k, p.pSlews[k].fall-pSlews[k].fall, p.pSlews[k].fall, pSlews[k].fall); }

        if Abc_AbsFloat(p.pTimes[k].rise - pTimes[k].rise) > fDiff {
            fmt.Printf("%6d : time rise differs %12.6f   %f %f\n", k, p.pTimes[k].rise-pTimes[k].rise, p.pTimes[k].rise, pTimes[k].rise); }
        if Abc_AbsFloat(p.pTimes[k].fall - pTimes[k].fall) > fDiff {
            fmt.Printf("%6d : time fall differs %12.6f   %f %f\n", k, p.pTimes[k].fall-pTimes[k].fall, p.pTimes[k].fall, pTimes[k].fall); }
    }
    }

/*
if ( memcmp( pTimes, p->pTimes, sizeof(SC_Pair) * p->nObjs ) )
    printf( "Times differ!\n" );
if ( memcmp( pSlews, p->pSlews, sizeof(SC_Pair) * p->nObjs ) )
    printf( "Slews differ!\n" );
if ( memcmp( pLoads, p->pLoads, sizeof(SC_Pair) * p->nObjs ) )
    printf( "Loads differ!\n" );
*/

    // inlining of macro ABC_FREE
    ((pTimes) ? (free((char *) (pTimes)), (pTimes) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pSlews) ? (free((char *) (pSlews)), (pSlews) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pLoads) ? (free((char *) (pLoads)), (pLoads) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Print cumulative statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclUpsizePrint(p *SC_Man, Iter int32, win int32, nPathPos int32, nPathNodes int32, nUpsizes int32, nTFOs int32, fVerbose int32) {
    fmt.Printf("%4d ", Iter);
    fmt.Printf("Win:%3d. ", win);
    fmt.Printf("PO:%6d. ", nPathPos);
    fmt.Printf("Path:%7d. ", nPathNodes);
    fmt.Printf("Gate:%5d. ", nUpsizes);
    fmt.Printf("TFO:%7d. ", nTFOs);
    fmt.Printf("A: ");
    fmt.Printf("%.2f ", p.SumArea);
    fmt.Printf("(%+5.1f %%)  ", 100.0 * (p.SumArea - p.SumArea0)/ p.SumArea0);
    fmt.Printf("D: ");
    fmt.Printf("%.2f ps ", p.MaxDelay);
    fmt.Printf("(%+5.1f %%)  ", 100.0 * (p.MaxDelay - p.MaxDelay0)/ p.MaxDelay0);
    fmt.Printf("B: ");
    fmt.Printf("%.2f ps ", p.BestDelay);
    fmt.Printf("(%+5.1f %%)", 100.0 * (p.BestDelay - p.MaxDelay0)/ p.MaxDelay0);
    fmt.Printf("%8.2f sec    ", 1.0*(Abc_Clock() - p.timeTotal)/(int64(time.Second)));
    fmt.Printf("%c", fVerbose ? '\n' : '\r');
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclUpsizeRemoveDangling(p *SC_Man, pNtk *Abc_Ntk_t) {
    var pCell *SC_Cell;
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachNodeNotBarBuf
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || Abc_ObjIsBarBuf(pObj) {} else
        if Abc_ObjFanoutNum(pObj) == 0 {
            pCell = Abc_SclObjCell( pObj );
            p.SumArea -= pCell.area;
            Abc_NtkDeleteObj_rec( pObj, 1 );
//            printf( "Removed node %d.\n", i );
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclUpsizePerformInt(pLib *SC_Lib, pNtk *Abc_Ntk_t, pPars *SC_SizePars) {
    var p *SC_Man;
    var vPathPos *Vec_Int_t  = nil;    // critical POs
    var vPathNodes *Vec_Int_t  = nil;  // critical nodes and PIs
    var clk abctime;
    var nRuntimeLimit abctime;
    if pPars.TimeOut {
        nRuntimeLimit = pPars.TimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nRuntimeLimit = 0;
    }
    var i int32;
    var win int32;
    var nUpsizes int32 = -1;
    var nFramesNoChange int32;
    var nConeSize int32;
    var nAllPos, nAllNodes, nAllTfos, nAllUpsizes int32;
    if pPars.fVerbose {
        fmt.Printf("Parameters: ");
        fmt.Printf("Iters =%5d.  ", pPars.nIters);
        fmt.Printf("Time win =%3d %%. ", pPars.Window);
        fmt.Printf("Update ratio =%3d %%. ", pPars.Ratio);
        fmt.Printf("UseDept =%2d. ", pPars.fUseDept);
        fmt.Printf("UseWL =%2d. ", pPars.fUseWireLoads);
        fmt.Printf("Target =%5d ps. ", pPars.DelayUser);
        fmt.Printf("DelayGap =%3d ps. ", pPars.DelayGap);
        fmt.Printf("Timeout =%4d sec", pPars.TimeOut);
        fmt.Printf("\n");
    }
    // increase window for larger networks
    if pPars.Window == 1 {
        pPars.Window += (Abc_NtkNodeNum(pNtk) > 40000); }
    // prepare the manager; collect init stats
    p = Abc_SclManStart( pLib, pNtk, pPars.fUseWireLoads, pPars.fUseDept, 0, pPars.BuffTreeEst );
    p.timeTotal  = Abc_Clock();
    assert(p.vGatesBest == nil, "p.vGatesBest == nil");
    p.vGatesBest = Vec_IntDup( p.pNtk.vGates );
    p.BestDelay  = p.MaxDelay0;
    // perform upsizing
    nAllPos = nAllNodes = nAllTfos = nAllUpsizes = 0;
    if p.BestDelay <= pPars.DelayUser {
        fmt.Printf("Current delay (%.2f ps) does not exceed the target delay (%.2f ps). Upsizing is not performed.\n", p.BestDelay, float32(pPars.DelayUser));
    } else {
    for  i = 0; i < pPars.nIters; i++  {
        for  win = pPars.Window + ((i % 7) == 6); win <= 100;  win *= 2  {
            // detect critical path
            clk = Abc_Clock();
            vPathPos   = Abc_SclFindCriticalCoWindow( p, win );
            vPathNodes = Abc_SclFindCriticalNodeWindow( p, vPathPos, win, pPars.fUseDept );
            p.timeCone += Abc_Clock() - clk;

            // selectively upsize the nodes
            clk = Abc_Clock();
            if pPars.BypassFreq && (i != 0) && (i % pPars.BypassFreq) == 0 {
                nUpsizes = Abc_SclFindBypasses( p, vPathNodes, pPars.Ratio, pPars.Notches, i, pPars.DelayGap, pPars.fVeryVerbose );
            } else {
                nUpsizes = Abc_SclFindUpsizes( p, vPathNodes, pPars.Ratio, pPars.Notches, i, pPars.DelayGap, (pPars.BypassFreq > 0) ); }
            p.timeSize += Abc_Clock() - clk;

            // unmark critical path
            clk = Abc_Clock();
            Abc_SclUnmarkCriticalNodeWindow( p, vPathNodes );
            Abc_SclUnmarkCriticalNodeWindow( p, vPathPos );
            p.timeCone += Abc_Clock() - clk;
            if nUpsizes > 0 {
                break; }
            Vec_IntFree( vPathPos );
            Vec_IntFree( vPathNodes );
        }
        if nUpsizes == 0 {
            break; }

        // update timing information
        clk = Abc_Clock();
        if pPars.fUseDept {
            if Vec_IntSize(p.vChanged) && !(pPars.BypassFreq && (i != 0) && (i % pPars.BypassFreq) == 0) {
                nConeSize = Abc_SclTimeIncUpdate( p );
            } else {
                Abc_SclTimeNtkRecompute( p, nil, nil, pPars.fUseDept, 0 ); }
        
        } else {
            var vTFO *Vec_Int_t  = Abc_SclFindTFO( p.pNtk, vPathNodes );
            Abc_SclTimeCone( p, vTFO );
            nConeSize = Vec_IntSize( vTFO );
            Vec_IntFree( vTFO );
        }
        p.timeTime += Abc_Clock() - clk;
//        Abc_SclUpsizePrintDiffs( p, pLib, pNtk );

        // save the best network
        p.MaxDelay = Abc_SclReadMaxDelay( p );
        if p.BestDelay > p.MaxDelay {
            p.BestDelay = p.MaxDelay;
            Abc_SclApplyUpdateToBest( p.vGatesBest, p.pNtk.vGates, p.vUpdates );
            Vec_IntClear( p.vUpdates2 );
            nFramesNoChange = 0;
        
        } else {
            nFramesNoChange++; }

        // report and cleanup
        Abc_SclUpsizePrint( p, i, win, Vec_IntSize(vPathPos), Vec_IntSize(vPathNodes), nUpsizes, nConeSize, pPars.fVeryVerbose || (pPars.fVerbose && nFramesNoChange == 0) ); //|| (i == nIters-1) );
        nAllPos     += Vec_IntSize(vPathPos);
        nAllNodes   += Vec_IntSize(vPathNodes);
        nAllTfos    += nConeSize;
        nAllUpsizes += nUpsizes;
        Vec_IntFree( vPathPos );
        Vec_IntFree( vPathNodes );
        // check timeout
        if nRuntimeLimit && Abc_Clock() > nRuntimeLimit {
            break; }
        // check no change
        if nFramesNoChange > pPars.nIterNoChange {
            break; }
        // check best delay
        if p.BestDelay <= pPars.DelayUser {
            break; }
    } }
    // update for best gates and recompute timing
    // inlining of macro ABC_SWAP
    { var t *Vec_Int_t  = p.vGatesBest; p.vGatesBest = p.pNtk.vGates; p.pNtk.vGates = t; };
    if pPars.BypassFreq != 0 {
        Abc_SclUndoRecentChanges( p.pNtk, p.vUpdates2 ); }
    if pPars.BypassFreq != 0 {
        Abc_SclUpsizeRemoveDangling( p, pNtk ); }
    Abc_SclTimeNtkRecompute( p, &p.SumArea, &p.MaxDelay, 0, 0 );
    if pPars.fVerbose {
        Abc_SclUpsizePrint( p, i, pPars.Window, nAllPos/(i?i:1), nAllNodes/(i?i:1), nAllUpsizes/(i?i:1), nAllTfos/(i?i:1), 1 );
    } else {
        fmt.Printf("                                                                                                                                                  \r"); }
    // report runtime
    p.timeTotal = Abc_Clock() - p.timeTotal;
    if pPars.fVerbose {
        p.timeOther = p.timeTotal - p.timeCone - p.timeSize - p.timeTime;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Critical path"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCone)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeCone)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Sizing eval  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSize)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSize)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Timing update"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTime)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTime)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Other        "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: TOTAL        "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
    }
    if pPars.fDumpStats {
        Abc_SclDumpStats( p, "stats2.txt", p.timeTotal ); }
    if nRuntimeLimit && Abc_Clock() > nRuntimeLimit {
        fmt.Printf("Gate sizing timed out at %d seconds.\n", pPars.TimeOut); }

    // save the result and quit
    Abc_SclSclGates2MioGates( pLib, pNtk ); // updates gate pointers
    Abc_SclManFree( p );
//    Abc_NtkCleanMarkAB( pNtk );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclUpsizePerform(pLib *SC_Lib, pNtk *Abc_Ntk_t, pPars *SC_SizePars) {
    var pNtkNew *Abc_Ntk_t  = pNtk;
    if pNtk.nBarBufs2 > 0 {
        pNtkNew = Abc_NtkDupDfsNoBarBufs( pNtk ); }
    Abc_SclUpsizePerformInt( pLib, pNtkNew, pPars );
    if pNtk.nBarBufs2 > 0 {
        Abc_SclTransferGates( pNtk, pNtkNew ); }
    if pNtk.nBarBufs2 > 0 {
        Abc_NtkDelete( pNtkNew ); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

