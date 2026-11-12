package scl

import (
	"fmt"
	"math"
)

/**CFile****************************************************************

  FileName    [sclBuffer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Buffering algorithms.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclBuffer.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclSize.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const BUF_SCALE = 1000

   
type Buf_Man_t struct {
    // parameters
    nFanMin            int32;   // the smallest fanout count to consider
    nFanMax            int32;   // the largest fanout count allowed off CP
    fBufPis            int32;   // enables buffing of the combinational inputs
    // internal deta
    pNtk     *Abc_Ntk_t;      // logic network
    vOffsets     *Vec_Int_t;  // offsets into edge delays
    vEdges     *Vec_Int_t;    // edge delays
    vArr     *Vec_Int_t;      // arrival times
    vDep     *Vec_Int_t;      // departure times
    vCounts     *Vec_Flt_t;   // fanout counts
    vQue     *Vec_Que_t;      // queue by fanout count
    nObjStart            int32; // the number of starting objects
    nObjAlloc            int32; // the number of allocated objects
    DelayMax            int32;  // maximum delay (percentage of inverter delay)
    DelayInv          float32;  // inverter delay
    // sorting fanouts
    vOrder     *Vec_Int_t;    // ordering of fanouts
    vDelays     *Vec_Int_t;   // fanout delays
    vNonCrit     *Vec_Int_t;  // non-critical fanouts
    vTfCone     *Vec_Int_t;   // TFI/TFO cone of the node including the node
    vFanouts     *Vec_Ptr_t;  // temp storage for fanouts
    // statistics
    nSeparate            int32;
    nDuplicate            int32;
    nBranch0            int32;
    nBranch1            int32;
    nBranchCrit            int32;
}

func Abc_BufNodeArr(p *Buf_Man_t, pObj *Abc_Obj_t) int32 { return Vec_IntEntry( p.vArr, Abc_ObjId(pObj) );                                   }
func Abc_BufNodeDep(p *Buf_Man_t, pObj *Abc_Obj_t) int32 { return Vec_IntEntry( p.vDep, Abc_ObjId(pObj) );                                   }
func Abc_BufSetNodeArr(p *Buf_Man_t, pObj *Abc_Obj_t, f int32) { Vec_IntWriteEntry( p.vArr, Abc_ObjId(pObj), f );                                  }
func Abc_BufSetNodeDep(p *Buf_Man_t, pObj *Abc_Obj_t, f int32) { Vec_IntWriteEntry( p.vDep, Abc_ObjId(pObj), f );                                  }
func Abc_BufEdgeDelay(p *Buf_Man_t, pObj *Abc_Obj_t, i int32) int32 { return Vec_IntEntry( p.vEdges, Vec_IntEntry(p.vOffsets, Abc_ObjId(pObj)) + i );  }
func Abc_BufSetEdgeDelay(p *Buf_Man_t, pObj *Abc_Obj_t, i int32, f int32) { Vec_IntWriteEntry( p.vEdges, Vec_IntEntry(p.vOffsets, Abc_ObjId(pObj)) + i, f ); }
func Abc_BufNodeSlack(p *Buf_Man_t, pObj *Abc_Obj_t) int32 { return p.DelayMax - Abc_BufNodeArr(p, pObj) - Abc_BufNodeDep(p, pObj);            }
func Abc_BufEdgeSlack(p *Buf_Man_t, pObj *Abc_Obj_t, pFan *Abc_Obj_t) int32 { return p.DelayMax - Abc_BufNodeArr(p, pObj) - Abc_BufNodeDep(p, pFan) - Abc_BufEdgeDelay(p, pFan, Abc_NodeFindFanin(pFan, pObj)); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Make sure fanins of gates are not duplicated.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclReportDupFanins(pNtk *Abc_Ntk_t) {
    var pObj, pFanin, pFanin2 *Abc_Obj_t;
    var i, k, k2 int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            // inlining of macro Abc_ObjForEachFanin
            for  k2 = 0; (k2 < Abc_ObjFaninNum(pObj)) && (((pFanin2) = Abc_ObjFanin(pObj, k2)), 1); k2++  {
                if k != k2 && pFanin == pFanin2 {
                    fmt.Printf("Node %d has dup fanin %d.\n", i, Abc_ObjId(pFanin)); }
            }
        } }
    }    
}

/**Function*************************************************************

  Synopsis    [Removes buffers and inverters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclObjIsBufInv(pObj *Abc_Obj_t) int32 {
    return Abc_ObjIsNode(pObj) && Abc_ObjFaninNum(pObj) == 1;
}
func Abc_SclIsInv(pObj *Abc_Obj_t) int32 {
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    return Mio_GateReadTruth((Mio_Gate_t *)pObj.pData) == ABC_CONST(0x5555555555555555);
}
func Abc_SclGetRealFaninLit(pObj *Abc_Obj_t) int32 {
    var iLit int32;
    if !(Abc_SclObjIsBufInv(pObj) != 0) {
        return Abc_Var2Lit( Abc_ObjId(pObj), 0 ); }
    iLit = Abc_SclGetRealFaninLit( Abc_ObjFanin0(pObj) );
    return Abc_LitNotCond( iLit, Abc_SclIsInv(pObj) );
}
func Abc_SclUnBufferPerform(pNtk *Abc_Ntk_t, fVerbose int32) *Abc_Ntk_t {
    var vLits *Vec_Int_t;
    var pObj, pFanin, pFaninNew *Abc_Obj_t;
    var i int32;
    var k int32;
    var iLit int32;
    var nNodesOld int32 = Abc_NtkObjNumMax(pNtk);
    // assign inverters
    vLits = Vec_IntStartFull( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
        if (Abc_SclIsInv(pObj) != 0) && !(Abc_SclObjIsBufInv(Abc_ObjFanin0(pObj)) != 0) {
            Vec_IntWriteEntry( vLits, Abc_ObjFaninId0(pObj), Abc_ObjId(pObj) ); }
    }
    // transfer fanins
    // inlining of macro Abc_NtkForEachNodeCo
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || (!Abc_ObjIsNode(pObj) && !Abc_ObjIsCo(pObj)) {} else {
        if i >= nNodesOld {
            break; }
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if !(Abc_SclObjIsBufInv(pFanin) != 0) {
                continue; }
            iLit = Abc_SclGetRealFaninLit( pFanin );
            pFaninNew = Abc_NtkObj( pNtk, Abc_Lit2Var(iLit) );
            if Abc_LitIsCompl(iLit) {
                if Vec_IntEntry( vLits, Abc_Lit2Var(iLit) ) == -1 {
                    pFaninNew = Abc_NtkCreateNodeInv( pNtk, pFaninNew );
                    Vec_IntWriteEntry( vLits, Abc_Lit2Var(iLit), Abc_ObjId(pFaninNew) );
                
                } else {
                    pFaninNew = Abc_NtkObj( pNtk, Vec_IntEntry( vLits, Abc_Lit2Var(iLit) ) ); }
                assert(Abc_ObjFaninNum(pFaninNew) == 1, "Abc_ObjFaninNum(pFaninNew) == 1");
            }
            if pFanin != pFaninNew {
                Abc_ObjPatchFanin( pObj, pFanin, pFaninNew ); }
        }
    }
    }
    Vec_IntFree( vLits );
    // duplicate network in topo order
    return Abc_NtkDupDfs( pNtk );
}

/**Function*************************************************************

  Synopsis    [Removes buffers and inverters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclCountMaxPhases(pNtk *Abc_Ntk_t) int32 {
    var vPhLevel *Vec_Int_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k, Max, MaxAll int32;
    vPhLevel = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNodeCo
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || (!Abc_ObjIsNode(pObj) && !Abc_ObjIsCo(pObj)) {} else {
        Max = 0;
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Max = Abc_MaxInt( Max, Vec_IntEntry(vPhLevel, Abc_ObjId(pFanin)) + Abc_ObjFaninPhase(pObj, k) );
        }
        Vec_IntWriteEntry( vPhLevel, i, Max );
        MaxAll = Abc_MaxInt( MaxAll, Max );
    }
    }
    Vec_IntFree( vPhLevel );
    return MaxAll;
}
func Abc_SclBufferPhase(pNtk *Abc_Ntk_t, fVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var vInvs *Vec_Int_t;
    var pObj, pFanin, pFaninNew *Abc_Obj_t;
    var nNodesOld int32  = Abc_NtkObjNumMax(pNtk);
    var i, k, Counter, Counter2, Total int32;
    assert(pNtk.vPhases != nil, "pNtk.vPhases != nil");
    vInvs = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNodeCo
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || (!Abc_ObjIsNode(pObj) && !Abc_ObjIsCo(pObj)) {} else {
        if i >= nNodesOld {
            break; }
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Total++;
            if !Abc_ObjFaninPhase(pObj, k) {
                continue; }
            if Vec_IntEntry(vInvs, Abc_ObjId(pFanin)) == 0 || Abc_ObjIsCi(pFanin) { // allow PIs to have high fanout - to be fixed later
            
                pFaninNew = Abc_NtkCreateNodeInv( pNtk, pFanin );
                Vec_IntWriteEntry( vInvs, Abc_ObjId(pFanin), Abc_ObjId(pFaninNew) );
                Counter++;
            }
            pFaninNew = Abc_NtkObj( pNtk, Vec_IntEntry(vInvs, Abc_ObjId(pFanin)) );
            Abc_ObjPatchFanin( pObj, pFanin, pFaninNew );
            Counter2++;
        }
    }
    }
    if fVerbose != 0 {
        fmt.Printf("Added %d inverters (%.2f %% fanins) (%.2f %% compl fanins).\n", Counter, 100.0 * Counter / Total, 100.0 * Counter2 / Total); }
    Vec_IntFree( vInvs );
    Vec_IntFillExtra( pNtk.vPhases, Abc_NtkObjNumMax(pNtk), 0 );
    // duplicate network in topo order
    vInvs = pNtk.vPhases;
    pNtk.vPhases = nil;
    pNtkNew = Abc_NtkDupDfs( pNtk );
    pNtk.vPhases = vInvs;
    return pNtkNew;
}
func Abc_SclUnBufferPhase(pNtk *Abc_Ntk_t, fVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj, pFanin, pFaninNew *Abc_Obj_t;
    var i, k, iLit, Counter, Total int32;
    assert(pNtk.vPhases == nil, "pNtk.vPhases == nil");
    pNtk.vPhases = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNodeCo
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || (!Abc_ObjIsNode(pObj) && !Abc_ObjIsCo(pObj)) {} else {
        if Abc_SclObjIsBufInv(pObj) != 0 {
            continue; }
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Total++;
            iLit = Abc_SclGetRealFaninLit( pFanin );
            pFaninNew = Abc_NtkObj( pNtk, Abc_Lit2Var(iLit) );
            if pFaninNew == pFanin {
                continue; }
            // skip fanins which are already fanins of the node
            if Abc_NodeFindFanin( pObj, pFaninNew ) >= 0 {
                continue; }
            Abc_ObjPatchFanin( pObj, pFanin, pFaninNew );
            if Abc_LitIsCompl(iLit) {
                Abc_ObjFaninFlipPhase( pObj, k );
Counter++; }
        }
    }
    }
    if fVerbose != 0 {
        fmt.Printf("Saved %d (%.2f %%) fanin phase bits.  ", Counter, 100.0 * Counter / Total); }
    // duplicate network in topo order
    pNtkNew = Abc_NtkDupDfs( pNtk );
    if fVerbose != 0 {
        fmt.Printf("Max depth = %d.\n", Abc_SclCountMaxPhases(pNtkNew)); }
    Abc_SclReportDupFanins( pNtkNew );
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Make sure the network is in topo order without dangling nodes.]

  Description [Returns 1 iff the network is fine.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclCheckNtk(p *Abc_Ntk_t, fVerbose int32) int32 {
    var pObj, pFanin *Abc_Obj_t;
    var i int32;
    var k int32;
    var fFlag int32 = 1;
    Abc_NtkIncrementTravId( p );        
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p)) && (((pObj) = Abc_NtkCi(p, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
    }
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if !Abc_NodeIsTravIdCurrent( pFanin ) {
                fmt.Printf("obj %d and its fanin %d are not in the topo order\n", Abc_ObjId(pObj), Abc_ObjId(pFanin));
fFlag = 0; }
        }
        Abc_NodeSetTravIdCurrent( pObj );
        if Abc_ObjIsBarBuf(pObj) {
            continue; }
        if Abc_ObjFanoutNum(pObj) == 0 {
            fmt.Printf("node %d has no fanout\n", Abc_ObjId(pObj));
fFlag = 0; }
        if !(fFlag != 0) {
            break; }
    }
    }
    if (fFlag != 0) && (fVerbose != 0) {
        fmt.Printf("The network is in topo order and no dangling nodes.\n"); }
    return fFlag;
}

/**Function*************************************************************

  Synopsis    [Performs buffering of the mapped network (old code).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeInvUpdateFanPolarity(pObj *Abc_Obj_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    assert(Abc_ObjFaninNum(pObj) == 0 || (Abc_SclObjIsBufInv(pObj) != 0), "Abc_ObjFaninNum(pObj) == 0 || Abc_SclObjIsBufInv(pObj)");
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        assert(Abc_ObjFaninNum(pFanout) > 0, "Abc_ObjFaninNum(pFanout) > 0");
        if Abc_SclObjIsBufInv(pFanout) != 0 {
            Abc_NodeInvUpdateFanPolarity( pFanout );
        } else {
            Abc_ObjFaninFlipPhase( pFanout, Abc_NodeFindFanin(pFanout, pObj) ); }
    }
}
func Abc_NodeInvUpdateObjFanoutPolarity(pObj *Abc_Obj_t, pFanout *Abc_Obj_t) {
    if Abc_SclObjIsBufInv(pFanout) != 0 {
        Abc_NodeInvUpdateFanPolarity( pFanout );
    } else {
        Abc_ObjFaninFlipPhase( pFanout, Abc_NodeFindFanin(pFanout, pObj) ); }
}
func Abc_NodeCompareLevels(pp1 **Abc_Obj_t, pp2 **Abc_Obj_t) int32 {
    var Diff int32  = Abc_ObjLevel(*pp1) - Abc_ObjLevel(*pp2);
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    Diff = (*pp1).Id - (*pp2).Id; // needed to make qsort() platform-infependent
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}
func Abc_SclComputeReverseLevel(pObj *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i, Level int32;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Level = Abc_MaxInt( Level, pFanout.Level );
    }
    return Level + 1;
}
func Abc_SclPerformBufferingOne(pObj *Abc_Obj_t, Degree int32, fUseInvs int32, fVerbose int32) *Abc_Obj_t {
    var vFanouts *Vec_Ptr_t;
    var pBuffer, pFanout *Abc_Obj_t;
    var i int32;
    var Degree0 int32 = Degree;
    assert(Abc_ObjFanoutNum(pObj) > Degree, "Abc_ObjFanoutNum(pObj) > Degree");
    // collect fanouts and sort by reverse level
    vFanouts = Vec_PtrAlloc( Abc_ObjFanoutNum(pObj) );
    Abc_NodeCollectFanouts( pObj, vFanouts );
    Vec_PtrSort( vFanouts, (int (*)(void))Abc_NodeCompareLevels );
    // select the first Degree fanouts
    if fUseInvs != 0 {
        pBuffer = Abc_NtkCreateNodeInv( pObj.pNtk, nil );
    } else {
        pBuffer = Abc_NtkCreateNodeBuf( pObj.pNtk, nil ); }
    // check if it is possible to not increase level
    if Vec_PtrSize(vFanouts) < 2 * Degree {
        var pFanPrev *Abc_Obj_t  = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, Vec_PtrSize(vFanouts)-1-Degree);
        var pFanThis *Abc_Obj_t  = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, Degree-1);
        var pFanLast *Abc_Obj_t  = (Abc_Obj_t *)Vec_PtrEntryLast(vFanouts);
        if Abc_ObjLevel(pFanThis) == Abc_ObjLevel(pFanLast) &&
             Abc_ObjLevel(pFanPrev) <  Abc_ObjLevel(pFanThis) {
            // find the first one whose level is the same as last
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
                if Abc_ObjLevel(pFanout) == Abc_ObjLevel(pFanLast) {
                    break; }
            }
            assert(i < Vec_PtrSize(vFanouts), "i < Vec_PtrSize(vFanouts)");
            if i > 1 {
                Degree = i; }
        }
        // make the last two more well-balanced
        if Degree == Degree0 && Degree > Vec_PtrSize(vFanouts) - Degree {
            Degree = Vec_PtrSize(vFanouts)/2 + (Vec_PtrSize(vFanouts) & 1); }
        assert(Degree <= Degree0, "Degree <= Degree0");
    }
    // select fanouts
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Degree) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
        Abc_ObjPatchFanin( pFanout, pObj, pBuffer );
    }
    if fVerbose != 0 {
        fmt.Printf("%5d : ", Abc_ObjId(pObj));
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
            fmt.Printf("%d%s ", Abc_ObjLevel(pFanout), i == Degree-1 ? "  " : "");
        }
        fmt.Printf("\n");
    }
    Vec_PtrFree( vFanouts );
    Abc_ObjAddFanin( pBuffer, pObj );
    pBuffer.Level = Abc_SclComputeReverseLevel( pBuffer );
    if fUseInvs != 0 {
        Abc_NodeInvUpdateFanPolarity( pBuffer ); }
    return pBuffer;
}
func Abc_SclPerformBuffering_rec(pObj *Abc_Obj_t, DegreeR int32, Degree int32, fUseInvs int32, fVerbose int32) {
    var vFanouts *Vec_Ptr_t;
    var pBuffer *Abc_Obj_t;
    var pFanout *Abc_Obj_t;
    var i, nOldFanNum int32;
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    pObj.Level = 0;
    if Abc_ObjIsCo(pObj) {
        return; }
    assert(Abc_ObjIsCi(pObj) || Abc_ObjIsNode(pObj), "Abc_ObjIsCi(pObj) || Abc_ObjIsNode(pObj)");
    // buffer fanouts and collect reverse levels
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Abc_SclPerformBuffering_rec( pFanout, DegreeR, Degree, fUseInvs, fVerbose );
    }
    // perform buffering as long as needed
    nOldFanNum = Abc_ObjFanoutNum(pObj);
    for  Abc_ObjFanoutNum(pObj) > Degree  {
        Abc_SclPerformBufferingOne( pObj, Degree, fUseInvs, fVerbose );
    }
    // add yet another level of buffers
    if (DegreeR != 0) && nOldFanNum > DegreeR {
        if fUseInvs != 0 {
            pBuffer = Abc_NtkCreateNodeInv( pObj.pNtk, nil );
        } else {
            pBuffer = Abc_NtkCreateNodeBuf( pObj.pNtk, nil ); }
        vFanouts = Vec_PtrAlloc( Abc_ObjFanoutNum(pObj) );
        Abc_NodeCollectFanouts( pObj, vFanouts );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
            Abc_ObjPatchFanin( pFanout, pObj, pBuffer );
        }
        Vec_PtrFree( vFanouts );
        Abc_ObjAddFanin( pBuffer, pObj );
        pBuffer.Level = Abc_SclComputeReverseLevel( pBuffer );
        if fUseInvs != 0 {
            Abc_NodeInvUpdateFanPolarity( pBuffer ); }
    }
    // compute the new level of the node
    pObj.Level = Abc_SclComputeReverseLevel( pObj );
}
func Abc_SclPerformBuffering(p *Abc_Ntk_t, DegreeR int32, Degree int32, fUseInvs int32, fVerbose int32) *Abc_Ntk_t {
    var vCiLevs *Vec_Int_t;
    var pNew *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkHasMapping(p), "Abc_NtkHasMapping(p)");
    if fUseInvs != 0 {
        fmt.Printf("Warning!!! Using inverters instead of buffers.\n");
        if p.vPhases == nil {
            fmt.Printf("The phases are not given. The result will not verify.\n"); }
    }
    // remember CI levels
    vCiLevs = Vec_IntAlloc( Abc_NtkCiNum(p) );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p)) && (((pObj) = Abc_NtkCi(p, i)), 1); i++  {
        Vec_IntPush( vCiLevs, Abc_ObjLevel(pObj) );
    }
    // perform buffering
    Abc_NtkIncrementTravId( p );        
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p)) && (((pObj) = Abc_NtkCi(p, i)), 1); i++  {
        Abc_SclPerformBuffering_rec( pObj, DegreeR, Degree, fUseInvs, fVerbose );
    }
    // recompute logic levels
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p)) && (((pObj) = Abc_NtkCi(p, i)), 1); i++  {
        pObj.Level = Vec_IntEntry( vCiLevs, i );
    }
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Abc_ObjLevelNew( pObj ); }
    }
    Vec_IntFree( vCiLevs );
    // if phases are present
    if p.vPhases {
        Vec_IntFillExtra( p.vPhases, Abc_NtkObjNumMax(p), 0 ); }
    // duplication in topo order
    pNew = Abc_NtkDupDfs( p );
    Abc_SclCheckNtk( pNew, fVerbose );
//    Abc_NtkDelete( pNew );
    return pNew;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BufComputeArr(p *Buf_Man_t, pObj *Abc_Obj_t) float32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    var DelayF float32;
    var Delay float32 = -ABC_INFINITY;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if Vec_IntEntry(p.vOffsets, Abc_ObjId(pObj)) == -ABC_INFINITY {
            continue; }
        DelayF = Abc_BufNodeArr(p, pFanin) + Abc_BufEdgeDelay(p, pObj, i);
        if Delay < DelayF {
            Delay = DelayF; }
    }
    Abc_BufSetNodeArr( p, pObj, Delay );
    return Delay;
}
func Abc_BufComputeDep(p *Buf_Man_t, pObj *Abc_Obj_t) float32 {
    var pFanout *Abc_Obj_t;
    var i int32;
    var DelayF float32;
    var Delay float32 = -ABC_INFINITY;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if Vec_IntEntry(p.vOffsets, Abc_ObjId(pFanout)) == -ABC_INFINITY {
            continue; }
        DelayF = Abc_BufNodeDep(p, pFanout) + Abc_BufEdgeDelay(p, pFanout, Abc_NodeFindFanin(pFanout, pObj));
        if Delay < DelayF {
            Delay = DelayF; }
    }
    Abc_BufSetNodeDep( p, pObj, Delay );
    return Delay;
}
func Abc_BufUpdateGlobal(p *Buf_Man_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    p.DelayMax = 0;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        p.DelayMax = Abc_MaxInt( p.DelayMax, Abc_BufNodeArr(p, Abc_ObjFanin0(pObj)) );
    }
}
func Abc_BufCreateEdges(p *Buf_Man_t, pObj *Abc_Obj_t) {
    var k int32;
    var pGate *Mio_Gate_t;
    if Abc_ObjIsCo(pObj) {
        pGate = nil;
    } else {
        pGate = (Mio_Gate_t *)pObj.pData;
    }
    Vec_IntWriteEntry( p.vOffsets, Abc_ObjId(pObj), Vec_IntSize(p.vEdges) );
    for  k = 0; k < Abc_ObjFaninNum(pObj); k++  {
        Vec_IntPush( p.vEdges, pGate ? int32((1.0 * BUF_SCALE * Mio_GateReadPinDelay(pGate, k) / p.DelayInv)) : 0 );
    }
}
func Abc_BufAddToQue(p *Buf_Man_t, pObj *Abc_Obj_t) {
    if Abc_ObjFanoutNum(pObj) < p.nFanMin || (!p.fBufPis && Abc_ObjIsCi(pObj)) {
        return; }
    Vec_FltWriteEntry( p.vCounts, Abc_ObjId(pObj), Abc_ObjFanoutNum(pObj) );
    if Vec_QueIsMember(p.vQue, Abc_ObjId(pObj)) {
        Vec_QueUpdate( p.vQue, Abc_ObjId(pObj) );
    } else {
        Vec_QuePush( p.vQue, Abc_ObjId(pObj) ); }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BufCollectTfoCone_rec(pNode *Abc_Obj_t, vNodes *Vec_Int_t) {
    var pNext *Abc_Obj_t;
    var i int32;
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    Abc_NodeSetTravIdCurrent( pNode );
    if Abc_ObjIsCo(pNode) {
        return; }
    assert(Abc_ObjIsCi(pNode) || Abc_ObjIsNode(pNode), "Abc_ObjIsCi(pNode) || Abc_ObjIsNode(pNode)");
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pNext) = Abc_ObjFanout(pNode, i)), 1); i++  {
        Abc_BufCollectTfoCone_rec( pNext, vNodes );
    }
    if Abc_ObjIsNode(pNode) {
        Vec_IntPush( vNodes, Abc_ObjId(pNode) ); }
}
func Abc_BufCollectTfoCone(p *Buf_Man_t, pObj *Abc_Obj_t) {
    Vec_IntClear( p.vTfCone );
    Abc_NtkIncrementTravId( p.pNtk );
    Abc_BufCollectTfoCone_rec( pObj, p.vTfCone );
}
func Abc_BufUpdateArr(p *Buf_Man_t, pObj *Abc_Obj_t) {
    var pNext *Abc_Obj_t;
    var i, Delay int32;
//    assert( Abc_ObjIsNode(pObj) );
    Abc_BufCollectTfoCone( p, pObj );
    Vec_IntReverseOrder( p.vTfCone );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(p.vTfCone) && (((pNext) = Abc_NtkObj(p.pNtk, Vec_IntEntry(p.vTfCone,i))), 1); i++  {         if (pNext) == nil {} else {
        Delay = Abc_BufComputeArr( p, pNext );
        p.DelayMax = Abc_MaxInt( p.DelayMax, Delay );
    }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BufCollectTfiCone_rec(pNode *Abc_Obj_t, vNodes *Vec_Int_t) {
    var pNext *Abc_Obj_t;
    var i int32;
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    Abc_NodeSetTravIdCurrent( pNode );
    if Abc_ObjIsCi(pNode) {
        return; }
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pNext) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_BufCollectTfiCone_rec( pNext, vNodes );
    }
    Vec_IntPush( vNodes, Abc_ObjId(pNode) );
}
func Abc_BufCollectTfiCone(p *Buf_Man_t, pObj *Abc_Obj_t) {
    Vec_IntClear( p.vTfCone );
    Abc_NtkIncrementTravId( p.pNtk );
    Abc_BufCollectTfiCone_rec( pObj, p.vTfCone );
}
func Abc_BufUpdateDep(p *Buf_Man_t, pObj *Abc_Obj_t) {
    var pNext *Abc_Obj_t;
    var i, Delay int32;
//    assert( Abc_ObjIsNode(pObj) );
    Abc_BufCollectTfiCone( p, pObj );
    Vec_IntReverseOrder( p.vTfCone );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(p.vTfCone) && (((pNext) = Abc_NtkObj(p.pNtk, Vec_IntEntry(p.vTfCone,i))), 1); i++  {         if (pNext) == nil {} else {
        Delay = Abc_BufComputeDep( p, pNext );
        p.DelayMax = Abc_MaxInt( p.DelayMax, Delay );
    }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Buf_ManStart(pNtk *Abc_Ntk_t, FanMin int32, FanMax int32, fBufPis int32) *Buf_Man_t {
    
    var pObj *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i int32;
    // inlining of macro ABC_CALLOC
    p := &Buf_Man_t{};
    p.pNtk      = pNtk;
    p.nFanMin   = FanMin;
    p.nFanMax   = FanMax;
    p.fBufPis   = fBufPis;
    // allocate arrays
    p.nObjStart = Abc_NtkObjNumMax(p.pNtk);
    p.nObjAlloc = (6 * Abc_NtkObjNumMax(p.pNtk) / 3) + 100;
    p.vOffsets  = Vec_IntAlloc( p.nObjAlloc );
    p.vArr      = Vec_IntAlloc( p.nObjAlloc );
    p.vDep      = Vec_IntAlloc( p.nObjAlloc );
    p.vCounts   = Vec_FltAlloc( p.nObjAlloc );
    p.vQue      = Vec_QueAlloc( p.nObjAlloc );
    Vec_IntFill( p.vOffsets, p.nObjAlloc, -ABC_INFINITY );
    Vec_IntFill( p.vArr,     p.nObjAlloc, 0 );
    Vec_IntFill( p.vDep,     p.nObjAlloc, 0 );
    Vec_FltFill( p.vCounts,  p.nObjAlloc, -ABC_INFINITY );
    Vec_QueSetPriority( p.vQue, Vec_FltArrayP(p.vCounts) );
    // collect edge delays
    p.DelayInv  = Mio_GateReadPinDelay( Mio_LibraryReadInv((Mio_Library_t *)pNtk.pManFunc), 0 );
    p.vEdges    = Vec_IntAlloc( 1000 );
    // create edges
    vNodes = Abc_NtkDfs( p.pNtk, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_BufCreateEdges( p, pObj );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        Abc_BufCreateEdges( p, pObj );
    }
    // derive delays
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_BufComputeArr( p, pObj );
    }
    // inlining of macro Vec_PtrForEachEntryReverse
    for  i = Vec_PtrSize(vNodes) - 1; (i >= 0) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i--  {
        Abc_BufComputeDep( p, pObj );
    }
    Abc_BufUpdateGlobal( p );
//    Abc_NtkForEachNode( p->pNtk, pObj, i )
//        printf( "%4d : %4d %4d\n", i, Abc_BufNodeArr(p, pObj), Abc_BufNodeDep(p, pObj) );
    // create fanout queue
//    Abc_NtkForEachCi( p->pNtk, pObj, i )
//        Abc_BufAddToQue( p, pObj );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_BufAddToQue( p, pObj );
    }
    Vec_PtrFree( vNodes );
    p.vDelays  = Vec_IntAlloc( 100 );
    p.vOrder   = Vec_IntAlloc( 100 );
    p.vNonCrit = Vec_IntAlloc( 100 );
    p.vTfCone  = Vec_IntAlloc( 100 );
    p.vFanouts = Vec_PtrAlloc( 100 );
    return p;
}
func Buf_ManStop(p *Buf_Man_t) {
    fmt.Printf("Sep = %d. Dup = %d. Br0 = %d. Br1 = %d. BrC = %d.  ", p.nSeparate, p.nDuplicate, p.nBranch0, p.nBranch1, p.nBranchCrit);
    fmt.Printf("Orig = %d. Add = %d. Rem = %d.\n", p.nObjStart, Abc_NtkObjNumMax(p.pNtk) - p.nObjStart, p.nObjAlloc - Abc_NtkObjNumMax(p.pNtk));
    Vec_PtrFree( p.vFanouts );
    Vec_IntFree( p.vTfCone );
    Vec_IntFree( p.vNonCrit );
    Vec_IntFree( p.vDelays );
    Vec_IntFree( p.vOrder );
    Vec_IntFree( p.vOffsets );
    Vec_IntFree( p.vEdges );
    Vec_IntFree( p.vArr );
    Vec_IntFree( p.vDep );
//    Vec_QueCheck( p->vQue );
    Vec_QueFree( p.vQue );
    Vec_FltFree( p.vCounts );
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BufSortByDelay(p *Buf_Man_t, iPivot int32) *Vec_Int_t {
    var pObj, pFanout *Abc_Obj_t;
    var i int32;
    var Slack int32;
    var pOrder *int32;
    Vec_IntClear( p.vDelays );
    pObj = Abc_NtkObj( p.pNtk, iPivot );
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Slack = Abc_BufEdgeSlack(p, pObj, pFanout);
        assert(Slack >= 0, "Slack >= 0");
        Vec_IntPush( p.vDelays, Abc_MaxInt(0, Slack) );
    }
    pOrder = Abc_QuickSortCost( Vec_IntArray(p.vDelays), Vec_IntSize(p.vDelays), 0 );
    Vec_IntClear( p.vOrder );
    for  i = 0; i < Vec_IntSize(p.vDelays); i++  {
        Vec_IntPush( p.vOrder, Abc_ObjId(Abc_ObjFanout(pObj, pOrder[i])) );
    }
    pOrder = nil;
//    for ( i = 0; i < Vec_IntSize(p->vDelays); i++ )
//        printf( "%5d - %5d   ", Vec_IntEntry(p->vOrder, i), Abc_BufEdgeSlack(p, pObj, Abc_NtkObj(p->pNtk, Vec_IntEntry(p->vOrder, i))) );
    return p.vOrder;        
}
func Abc_BufPrintOne(p *Buf_Man_t, iPivot int32) {
    var pObj, pFanout *Abc_Obj_t;
    var vOrder *Vec_Int_t;
    var i, Slack int32;
    pObj = Abc_NtkObj( p.pNtk, iPivot );
    vOrder = Abc_BufSortByDelay( p, iPivot );
    fmt.Printf("Node %5d  Fi = %d  Fo = %3d  Lev = %3d : {", iPivot, Abc_ObjFaninNum(pObj), Abc_ObjFanoutNum(pObj), Abc_ObjLevel(pObj));
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vOrder) && (((pFanout) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vOrder,i))), 1); i++  {         if (pFanout) == nil {} else {
        Slack = Abc_BufEdgeSlack( p, pObj, pFanout );
        fmt.Printf(" %d(%d)", Abc_ObjId(pFanout), Slack);
    }
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BufReplaceBufsByInvs(pNtk *Abc_Ntk_t) {
    var pObj, pInv *Abc_Obj_t;
    var i, Counter int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if !Abc_NodeIsBuf(pObj) {
            continue; }
        assert(pObj.pData == Mio_LibraryReadBuf((Mio_Library_t *)pNtk.pManFunc), "pObj.pData == Mio_LibraryReadBuf((Mio_Library_t *)pNtk.pManFunc)");
        pObj.pData = Mio_LibraryReadInv((Mio_Library_t *)pNtk.pManFunc);
        pInv = Abc_NtkCreateNodeInv( pNtk, Abc_ObjFanin0(pObj) );
        Abc_ObjPatchFanin( pObj, Abc_ObjFanin0(pObj), pInv );
        Counter++;
    }
    }
    fmt.Printf("Replaced %d buffers by invertor pairs.\n", Counter);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BufComputeAverage(p *Buf_Man_t, iPivot int32, vOrder *Vec_Int_t) int32 {
    var pObj, pFanout *Abc_Obj_t;
    var i, Average int32;
    pObj = Abc_NtkObj( p.pNtk, iPivot );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vOrder) && (((pFanout) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vOrder,i))), 1); i++  {         if (pFanout) == nil {} else {
        Average += Abc_BufEdgeSlack( p, pObj, pFanout ); }
    }
    return Average / Vec_IntSize(vOrder);
}
func Abc_BufFindNonBuffDriver(p *Buf_Man_t, pObj *Abc_Obj_t) *Abc_Obj_t {
    if Abc_ObjIsNode(pObj) && Abc_NodeIsBuf(pObj) {
        return Abc_BufFindNonBuffDriver(p, Abc_ObjFanin0(pObj));
    } else {
        return pObj;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BufCountNonCritical(p *Buf_Man_t, pObj *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i int32;
    Vec_IntClear( p.vNonCrit );
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if Abc_BufEdgeSlack( p, pObj, pFanout ) > 7*BUF_SCALE/2 {
            Vec_IntPush( p.vNonCrit, Abc_ObjId(pFanout) ); }
    }
    return Vec_IntSize(p.vNonCrit);
}
func Abc_BufPerformOne(p *Buf_Man_t, iPivot int32, fSkipDup int32, fVerbose int32) {
    var pObj, pFanout *Abc_Obj_t;
    var i, j, nCrit, nNonCrit int32;
//    int DelayMax = p->DelayMax;
    assert(Abc_NtkObjNumMax(p.pNtk) + 30 < p.nObjAlloc, "Abc_NtkObjNumMax(p.pNtk) + 30 < p.nObjAlloc");
    pObj     = Abc_NtkObj( p.pNtk, iPivot );
//    assert( Vec_FltEntry(p->vCounts, iPivot) == (float)Abc_ObjFanoutNum(pObj) );
    nNonCrit = Abc_BufCountNonCritical( p, pObj );
    nCrit    = Abc_ObjFanoutNum(pObj) - nNonCrit;
if fVerbose != 0 {
//Abc_BufPrintOne( p, iPivot );
fmt.Printf("ObjId = %6d : %-10s   FI = %d. FO =%4d.  Crit =%4d.  ", Abc_ObjId(pObj), Mio_GateReadName((Mio_Gate_t *)pObj.pData), Abc_ObjFaninNum(pObj), Abc_ObjFanoutNum(pObj), nCrit);
}
    // consider three cases
    if nCrit > 0 && nNonCrit > 1 {
        // (1) both critical and non-critical are present - split them by adding buffer
        var pBuffer *Abc_Obj_t  = Abc_NtkCreateNodeBuf( p.pNtk, pObj );
        // inlining of macro Abc_NtkForEachObjVec
        for  i = 0; i < Vec_IntSize(p.vNonCrit) && (((pFanout) = Abc_NtkObj(p.pNtk, Vec_IntEntry(p.vNonCrit,i))), 1); i++  {         if (pFanout) == nil {} else {
            Abc_ObjPatchFanin( pFanout, pObj, pBuffer ); }
        }
        // update timing
        Abc_BufCreateEdges( p, pBuffer );
        Abc_BufUpdateArr( p, pBuffer );
        Abc_BufUpdateDep( p, pBuffer );
        Abc_BufAddToQue( p, pObj );
        Abc_BufAddToQue( p, pBuffer );
        Abc_SclTimeIncUpdateLevel( pBuffer );
        p.nSeparate++;
if fVerbose != 0 {
fmt.Printf("Adding buffer\n"); }
    
    } else if !(fSkipDup != 0) && nCrit > 0 && Abc_ObjIsNode(pObj) && Abc_ObjFanoutNum(pObj) > p.nFanMin {//&& Abc_ObjLevel(pObj) < 4 )//&& Abc_ObjFaninNum(pObj) < 2 )
    
        // (2) only critical are present - duplicate
        var pClone *Abc_Obj_t  = Abc_NtkDupObj( p.pNtk, pObj, 0 );
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanout) = Abc_ObjFanin(pObj, i)), 1); i++  {
            Abc_ObjAddFanin( pClone, pFanout );
        }
        Abc_NodeCollectFanouts( pObj, p.vFanouts );
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Vec_PtrSize(p.vFanouts)/2) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(p.vFanouts, i)), 1); i++  {
            Abc_ObjPatchFanin( pFanout, pObj, pClone );
        }
        // update timing
        Abc_BufCreateEdges( p, pClone );
        Abc_BufSetNodeArr( p, pClone, Abc_BufNodeArr(p, pObj) );
        Abc_BufUpdateDep( p, pObj );
        Abc_BufUpdateDep( p, pClone );
        Abc_BufAddToQue( p, pObj );
        Abc_BufAddToQue( p, pClone );
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanout) = Abc_ObjFanin(pObj, i)), 1); i++  {
            Abc_BufAddToQue( p, pFanout );
        }
        Abc_SclTimeIncUpdateLevel( pClone );
        p.nDuplicate++;
//        printf( "Duplicating %s on level %d\n", Mio_GateReadName((Mio_Gate_t *)pObj->pData), Abc_ObjLevel(pObj) );
if fVerbose != 0 {
fmt.Printf("Duplicating node\n"); }
    
    } else if (nCrit > 0 && Abc_ObjFanoutNum(pObj) > 8) || Abc_ObjFanoutNum(pObj) > p.nFanMax {
        // (2) only critical or only non-critical - add buffer/inverter tree
        var nDegree, n1Degree, n1Number, nFirst int32;
        var iFirstBuf int32  = Abc_NtkObjNumMax( p.pNtk );
//        nDegree  = Abc_MinInt( 3, (int)pow(Abc_ObjFanoutNum(pObj), 0.34) );
        nDegree  = Abc_MinInt( 10, int32(math.Pow(float64(Abc_ObjFanoutNum(pObj)), float64(0.5))) );
        n1Degree = Abc_ObjFanoutNum(pObj) / nDegree + 1;
        n1Number = Abc_ObjFanoutNum(pObj) % nDegree;
        nFirst   = n1Degree * n1Number;
        p.nBranchCrit += (nCrit > 0);
        // create inverters
        Abc_NodeCollectFanouts( pObj, p.vFanouts );
        if Abc_ObjIsNode(pObj) && Abc_NodeIsBuf(pObj) {
            p.nBranch0++;
            pObj.pData = Mio_LibraryReadInv((Mio_Library_t *)p.pNtk.pManFunc);
            Abc_BufSetEdgeDelay( p, pObj, 0, BUF_SCALE );
            assert(Abc_NodeIsInv(pObj), "Abc_NodeIsInv(pObj)");
            for  i = 0; i < nDegree; i++  {
                Abc_NtkCreateNodeInv( p.pNtk, pObj );
            }
if fVerbose != 0 {
fmt.Printf("Adding %d inverters\n", nDegree); }
        
        } else {
            p.nBranch1++;
            for  i = 0; i < nDegree; i++  {
                Abc_NtkCreateNodeBuf( p.pNtk, pObj );
            }
if fVerbose != 0 {
fmt.Printf("Adding %d buffers\n", nDegree); }
        }
        // connect inverters
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(p.vFanouts, i)), 1); i++  {
            if i < nFirst {
                j = i/n1Degree;
            } else {
                j = n1Number + ((i - nFirst)/(n1Degree - 1));
            }
            assert(j >= 0 && j < nDegree, "j >= 0 && j < nDegree");
            Abc_ObjPatchFanin( pFanout, pObj, Abc_NtkObj(p.pNtk, iFirstBuf + j) );
        }
        // update timing
        for  i = 0; i < nDegree; i++  {
            Abc_BufCreateEdges( p, Abc_NtkObj(p.pNtk, iFirstBuf + i) );
        }
        Abc_BufUpdateArr( p, pObj );
        for  i = 0; i < nDegree; i++  {
            Abc_BufComputeDep( p, Abc_NtkObj(p.pNtk, iFirstBuf + i) );
        }
        Abc_BufUpdateDep( p, pObj );
        for  i = 0; i < nDegree; i++  {
            Abc_BufAddToQue( p, Abc_NtkObj(p.pNtk, iFirstBuf + i) );
        }
        for  i = 0; i < nDegree; i++  {
            Abc_SclTimeIncUpdateLevel( Abc_NtkObj(p.pNtk, iFirstBuf + i) );
        }
    
    } else {
if fVerbose != 0 {
fmt.Printf("Doing nothing\n"); }
    }
//    if ( DelayMax != p->DelayMax )
//        printf( "%d (%.2f)  ", p->DelayMax, 1.0 * p->DelayMax * p->DelayInv / BUF_SCALE );
}
func Abc_SclBufPerform(pNtk *Abc_Ntk_t, FanMin int32, FanMax int32, fBufPis int32, fSkipDup int32, fVerbose int32) *Abc_Ntk_t {
    var pNew *Abc_Ntk_t;
    var p *Buf_Man_t  = Buf_ManStart( pNtk, FanMin, FanMax, fBufPis );
    var i int32;
    var Limit int32 = ABC_INFINITY;
    Abc_NtkLevel( pNtk );
//    if ( Abc_NtkNodeNum(pNtk) < 1000 )
//        fSkipDup = 1;
    for  i = 0; i < Limit && Vec_QueSize(p.vQue); i++  {
        Abc_BufPerformOne( p, Vec_QuePop(p.vQue), fSkipDup, fVerbose );
    }
    Buf_ManStop( p );
//    Abc_BufReplaceBufsByInvs( pNtk );
    // duplicate network in topo order
    pNew = Abc_NtkDupDfs( pNtk );
    Abc_SclCheckNtk( pNew, fVerbose );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

