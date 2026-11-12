package scl

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [sclDnsize.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Selective decrease of gate sizes.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclDnsize.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

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

  Synopsis    [Find the array of nodes to be updated.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindWindow(pPivot *Abc_Obj_t, pvNodes **Vec_Int_t, pvEvals **Vec_Int_t) {
    var p *Abc_Ntk_t  = Abc_ObjNtk(pPivot);
    var pObj, pNext, pNext2 *Abc_Obj_t;
    var vNodes *Vec_Int_t  = *pvNodes;
    var vEvals *Vec_Int_t  = *pvEvals;
    var i, k int32;
    assert(Abc_ObjIsNode(pPivot), "Abc_ObjIsNode(pPivot)");
    // collect fanins, node, and fanouts
    Vec_IntClear( vNodes );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pNext) = Abc_ObjFanin(pPivot, i)), 1); i++  {
//        if ( Abc_ObjIsNode(pNext) && Abc_ObjFaninNum(pNext) > 0 )
        if Abc_ObjIsCi(pNext) || Abc_ObjFaninNum(pNext) > 0 {
            Vec_IntPush( vNodes, Abc_ObjId(pNext) ); }
    }
    Vec_IntPush( vNodes, Abc_ObjId(pPivot) );
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pPivot)) && (((pNext) = Abc_ObjFanout(pPivot, i)), 1); i++  {
        if Abc_ObjIsNode(pNext) {
            Vec_IntPush( vNodes, Abc_ObjId(pNext) );
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pNext)) && (((pNext2) = Abc_ObjFanout(pNext, k)), 1); k++  {
                if Abc_ObjIsNode(pNext2) {
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
            if !pNext.fMarkB {
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

  Synopsis    [Returns 1 if the node can be improved.]

  Description [Updated the node to have a new gate.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclCheckImprovement(p *SC_Man, pObj *Abc_Obj_t, vNodes *Vec_Int_t, vEvals *Vec_Int_t, Notches int32, DelayGap int32) int32 {
    var pTemp *Abc_Obj_t;
    var pCellOld, pCellNew *SC_Cell;
    var dGain, dGainBest float32;
    var i, k, gateBest int32;
    var clk abctime;
clk = Abc_Clock();
//    printf( "%d -> %d\n", Vec_IntSize(vNodes), Vec_IntSize(vEvals) );
    // save old gate, timing, fanin load
    pCellOld = Abc_SclObjCell( pObj );
    Abc_SclConeStore( p, vNodes );
    Abc_SclEvalStore( p, vEvals );
    Abc_SclLoadStore( p, pObj );
    // try different gate sizes for this node
    gateBest = -1;
    dGainBest = -DelayGap;
    // inlining of macro SC_RingForEachCellRev
    for  i, pCellNew = 0, pCellOld; i == 0 || pCellNew != pCellOld; pCellNew, i = pCellNew.pPrev, i+1  {
        if pCellNew.area >= pCellOld.area {
            continue; }
        if i > Notches {
            break; }
        // set new cell
        Abc_SclObjSetCell( pObj, pCellNew );
        Abc_SclUpdateLoad( p, pObj, pCellOld, pCellNew );
        // recompute timing
        Abc_SclTimeCone( p, vNodes );
        // set old cell
        Abc_SclObjSetCell( pObj, pCellOld );
        Abc_SclLoadRestore( p, pObj );
        // evaluate gain
        dGain = Abc_SclEvalPerformLegal( p, vEvals, p.MaxDelay0 );
        if dGain == -1 {
            continue; }
        // save best gain
        if dGainBest < dGain {
            dGainBest = dGain;
            gateBest = pCellNew.Id;
        }
    } 
    // put back old cell and timing
    Abc_SclObjSetCell( pObj, pCellOld );
    Abc_SclConeRestore( p, vNodes );
p.timeSize += Abc_Clock() - clk;
    if gateBest >= 0 {
        pCellNew = SC_LibCell( p.pLib, gateBest );
        Abc_SclObjSetCell( pObj, pCellNew );
        p.SumArea += pCellNew.area - pCellOld.area;
//        printf( "%f   %f -> %f\n", pCellNew->area - pCellOld->area, p->SumArea - (pCellNew->area - pCellOld->area), p->SumArea );
//        printf( "%6d  %20s -> %20s  %f -> %f\n", Abc_ObjId(pObj), pCellOld->pName, pCellNew->pName, pCellOld->area, pCellNew->area );
        // mark used nodes with the current trav ID
        // inlining of macro Abc_NtkForEachObjVec
        for  k = 0; k < Vec_IntSize(vNodes) && (((pTemp) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vNodes,k))), 1); k++  {         if (pTemp) == nil {} else {
            Abc_NodeSetTravIdCurrent( pTemp ); }
        }
        // update load and timing...
        Abc_SclUpdateLoad( p, pObj, pCellOld, pCellNew );
        Abc_SclTimeIncInsert( p, pObj );
        return 1;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Collect nodes by area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectNodesByArea(p *SC_Man, pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Vec_QueSize(p.vNodeByGain) == 0, "Vec_QueSize(p.vNodeByGain) == 0");
    Vec_QueClear( p.vNodeByGain );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
    if Abc_ObjFaninNum(pObj) > 0 {
        Vec_FltWriteEntry( p.vNode2Gain, Abc_ObjId(pObj), Abc_SclObjCell(pObj).area );
        Vec_QuePush( p.vNodeByGain, Abc_ObjId(pObj) );
    }
    }
}
func Abc_SclCheckOverlap(pNtk *Abc_Ntk_t, vNodes *Vec_Int_t) int32 {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else
        if Abc_NodeIsTravIdCurrent(pObj) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Print cumulative statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclDnsizePrint(p *SC_Man, Iter int32, nAttempts int32, nOverlaps int32, nChanges int32, fVerbose int32) {
    if Iter == -1 {
        fmt.Printf("Total : ");
    } else {
        fmt.Printf("%5d : ", Iter); }
    fmt.Printf("Try =%6d  ", nAttempts);
    fmt.Printf("Over =%6d  ", nOverlaps);
    fmt.Printf("Fail =%6d  ", nAttempts-nOverlaps-nChanges);
    fmt.Printf("Win =%6d  ", nChanges);
    fmt.Printf("A: ");
    fmt.Printf("%.2f ", p.SumArea);
    fmt.Printf("(%+5.1f %%)  ", 100.0 * (p.SumArea - p.SumArea0)/ p.SumArea0);
    fmt.Printf("D: ");
    fmt.Printf("%.2f ps ", p.MaxDelay);
    fmt.Printf("(%+5.1f %%)  ", 100.0 * (p.MaxDelay - p.MaxDelay0)/ p.MaxDelay0);
    fmt.Printf("%8.2f sec    ", 1.0*(Abc_Clock() - p.timeTotal)/(int64(time.Second)));
    fmt.Printf("%c", fVerbose ? '\n' : '\r');
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclDnsizePerformInt(pLib *SC_Lib, pNtk *Abc_Ntk_t, pPars *SC_SizePars) {
    var p *SC_Man;
    var pObj *Abc_Obj_t;
    var vNodes, vEvals, vTryLater *Vec_Int_t; 
    var clk abctime;
    var nRuntimeLimit abctime;
    if pPars.TimeOut {
        nRuntimeLimit = pPars.TimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nRuntimeLimit = 0;
    }
    var i, k int32;

    if pPars.fVerbose {
        fmt.Printf("Parameters: ");
        fmt.Printf("Iters =%5d.  ", pPars.nIters);
        fmt.Printf("UseDept =%2d. ", pPars.fUseDept);
        fmt.Printf("UseWL =%2d. ", pPars.fUseWireLoads);
        fmt.Printf("Target =%5d ps. ", pPars.DelayUser);
        fmt.Printf("DelayGap =%3d ps. ", pPars.DelayGap);
        fmt.Printf("Timeout =%4d sec", pPars.TimeOut);
        fmt.Printf("\n");
    }

    // prepare the manager; collect init stats
    p = Abc_SclManStart( pLib, pNtk, pPars.fUseWireLoads, pPars.fUseDept, pPars.DelayUser, pPars.BuffTreeEst );
    p.timeTotal  = Abc_Clock();
    assert(p.vGatesBest == nil, "p.vGatesBest == nil");
    p.vGatesBest = Vec_IntDup( p.pNtk.vGates );

    // perform upsizing
    vNodes = Vec_IntAlloc( 1000 );
    vEvals = Vec_IntAlloc( 1000 );
    vTryLater = Vec_IntAlloc( 1000 );
    for  i = 0; i < pPars.nIters; i++  {
        var nRounds int32;
        var nAttemptAll, nOverlapAll, nChangesAll int32;
        Abc_NtkCollectNodesByArea( p, pNtk );
        for  Vec_QueSize(p.vNodeByGain) > 0  {
            var nAttempt, nOverlap, nChanges int32;
            Vec_IntClear( vTryLater );
            Abc_NtkIncrementTravId( pNtk );
            for  Vec_QueSize(p.vNodeByGain) > 0  {
                clk = Abc_Clock();
                pObj = Abc_NtkObj( p.pNtk, Vec_QuePop(p.vNodeByGain) );
                Abc_SclFindWindow( pObj, &vNodes, &vEvals );
                p.timeCone += Abc_Clock() - clk;
                if Abc_SclCheckOverlap( p.pNtk, vNodes ) != 0 {
                    nOverlap++, Vec_IntPush( vTryLater, Abc_ObjId(pObj) );
                } else { 
                    nChanges += Abc_SclCheckImprovement( p, pObj, vNodes, vEvals, pPars.Notches, pPars.DelayGap ); }
                nAttempt++;
            }
            // inlining of macro Abc_NtkForEachObjVec
            for  k = 0; k < Vec_IntSize(vTryLater) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTryLater,k))), 1); k++  {         if (pObj) == nil {} else {
                Vec_QuePush( p.vNodeByGain, Abc_ObjId(pObj) ); }
            }

            clk = Abc_Clock();
            if Vec_IntSize(p.vChanged) {
                Abc_SclTimeIncUpdate( p );
            } else {
                Abc_SclTimeNtkRecompute( p, &p.SumArea, &p.MaxDelay, pPars.fUseDept, pPars.DelayUser ); }
            p.timeTime += Abc_Clock() - clk;

            p.MaxDelay = Abc_SclReadMaxDelay( p );
            if pPars.fUseDept && pPars.DelayUser > 0 && p.MaxDelay < pPars.DelayUser {
                p.MaxDelay = pPars.DelayUser; }
            Abc_SclDnsizePrint( p, nRounds++, nAttempt, nOverlap, nChanges, pPars.fVeryVerbose ); 
            nAttemptAll += nAttempt; nOverlapAll += nOverlap; nChangesAll += nChanges;
            if nRuntimeLimit && Abc_Clock() > nRuntimeLimit {
                break; }
        }
        // recompute
//        Abc_SclTimeNtkRecompute( p, &p->SumArea, &p->MaxDelay, pPars->fUseDept, pPars->DelayUser );
        if pPars.fVerbose {
            Abc_SclDnsizePrint( p, -1, nAttemptAll, nOverlapAll, nChangesAll, 1 ); } 
        if nRuntimeLimit && Abc_Clock() > nRuntimeLimit {
            break; }
        if nAttemptAll == 0 {
            break; }
    }
    Vec_IntFree( vNodes );
    Vec_IntFree( vEvals );
    Vec_IntFree( vTryLater );
    if !pPars.fVerbose {
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
func Abc_SclDnsizePerform(pLib *SC_Lib, pNtk *Abc_Ntk_t, pPars *SC_SizePars) {
    var pNtkNew *Abc_Ntk_t  = pNtk;
    if pNtk.nBarBufs2 > 0 {
        pNtkNew = Abc_NtkDupDfsNoBarBufs( pNtk ); }
    Abc_SclDnsizePerformInt( pLib, pNtkNew, pPars );
    if pNtk.nBarBufs2 > 0 {
        Abc_SclTransferGates( pNtk, pNtkNew ); }
    if pNtk.nBarBufs2 > 0 {
        Abc_NtkDelete( pNtkNew ); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

