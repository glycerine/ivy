package cut

import "fmt"

/**CFile****************************************************************

  FileName    [cutNode.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [K-feasible cut computation package.]

  Synopsis    [Procedures to compute cuts for a node.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cutNode.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cutInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Cut_NodeMapping( Cut_Man_t * p, Cut_Cut_t * pCuts, int Node, int Node0, int Node1 );

// static int Cut_NodeMapping2( Cut_Man_t * p, Cut_Cut_t * pCuts, int Node, int Node0, int Node1 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns 1 if pDom is contained in pCut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutCheckDominance(pDom *Cut_Cut_t, pCut *Cut_Cut_t) int32 {
    var i, k int32;
    for  i = 0; i < int32(pDom.nLeaves); i++  {
        for  k = 0; k < int32(pCut.nLeaves); k++  {
            if pDom.pLeaves[i] == pCut.pLeaves[k] {
                break; }
        }
        if k == int32(pCut.nLeaves) { // node i in pDom is not contained in pCut
            return 0; }
    }
    // every node in pDom is contained in pCut
    return 1;
}

/**Function*************************************************************

  Synopsis    [Filters cuts using dominance.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutFilter(p *Cut_Man_t, pList *Cut_Cut_t) { 
    var pListR *Cut_Cut_t;
    var ppListR **Cut_Cut_t = &pListR;
    var pCut, pCut2, pDom, pPrev *Cut_Cut_t;
    // save the first cut
    *ppListR = pList, ppListR = &pList.pNext;
    // try to filter out other cuts
    pPrev = pList;
    // inlining of macro Cut_ListForEachCutSafe
    for  pCut = pList.pNext,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
        assert(pCut.nLeaves > 1, "pCut.nLeaves > 1");
        // go through all the previous cuts up to pCut
        // inlining of macro Cut_ListForEachCutStop
        for  pDom = pList.pNext;                                             pDom != pCut;                                            pDom = pDom.pNext  {
            if pDom.nLeaves > pCut.nLeaves {
                continue; }
            if (pDom.uSign & pCut.uSign) != pDom.uSign {
                continue; }
            if Cut_CutCheckDominance( pDom, pCut ) != 0 {
                break; }
        }
        if pDom != pCut { // pDom is contained in pCut - recycle pCut
        
            // make sure cuts are connected after removing
            pPrev.pNext = pCut.pNext;
            // recycle the cut
            Cut_CutRecycle( p, pCut );
        
        } else { // pDom is NOT contained in pCut - save pCut
        
            *ppListR = pCut, ppListR = &pCut.pNext;
            pPrev = pCut;
        }
    }
    *ppListR = nil;
}

/**Function*************************************************************

  Synopsis    [Checks equality of one cut.]

  Description [Returns 1 if the cut is removed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutFilterOneEqual(p *Cut_Man_t, pSuperList *Cut_List_t, pCut *Cut_Cut_t) int32 {
    var pTemp *Cut_Cut_t;
    // inlining of macro Cut_ListForEachCut
    for  pTemp = pSuperList.pHead[pCut.nLeaves];                                             (pTemp != nil);                                                     pTemp = pTemp.pNext  {
        // skip the non-contained cuts
        if pCut.uSign != pTemp.uSign {
            continue; }
        // check containment seriously
        if Cut_CutCheckDominance( pTemp, pCut ) != 0 {
            p.nCutsFilter++;
            Cut_CutRecycle( p, pCut );
            return 1;
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Checks containment for one cut.]

  Description [Returns 1 if the cut is removed.]
               
  SideEffects [May remove other cuts in the set.]

  SeeAlso     []

***********************************************************************/
func Cut_CutFilterOne(p *Cut_Man_t, pSuperList *Cut_List_t, pCut *Cut_Cut_t) int32 {
    var pTemp *Cut_Cut_t;
    var pTemp2 *Cut_Cut_t;
    var ppTail **Cut_Cut_t;
    var a int32;

    // check if this cut is filtered out by smaller cuts
    for  a = 2; a <= int32(pCut.nLeaves); a++  {
        // inlining of macro Cut_ListForEachCut
        for  pTemp = pSuperList.pHead[a];                                             (pTemp != nil);                                                     pTemp = pTemp.pNext  {
            // skip the non-contained cuts
            if (pTemp.uSign & pCut.uSign) != pTemp.uSign {
                continue; }
            // check containment seriously
            if Cut_CutCheckDominance( pTemp, pCut ) != 0 {
                p.nCutsFilter++;
                Cut_CutRecycle( p, pCut );
                return 1;
            }
        }
    }

    // filter out other cuts using this one
    for  a = pCut.nLeaves + 1; a <= int32(pCut.nVarsMax); a++  {
        ppTail = pSuperList.pHead + a;
        // inlining of macro Cut_ListForEachCutSafe
        for  pTemp = pSuperList.pHead[a],                                             pTemp2 = pTemp? pTemp.pNext: nil;                          (pTemp != nil);                                                     pTemp = pTemp2,                                             pTemp2 = pTemp? pTemp.pNext: nil  {
            // skip the non-contained cuts
            if (pTemp.uSign & pCut.uSign) != pCut.uSign {
                ppTail = &pTemp.pNext;
                continue;
            }
            // check containment seriously
            if Cut_CutCheckDominance( pCut, pTemp ) != 0 {
                p.nCutsFilter++;
                p.nNodeCuts--;
                // move the head
                if pSuperList.pHead[a] == pTemp {
                    pSuperList.pHead[a] = pTemp.pNext; }
                // move the tail
                if pSuperList.ppTail[a] == &pTemp.pNext {
                    pSuperList.ppTail[a] = ppTail; }
                // skip the given cut in the list
                *ppTail = pTemp.pNext;
                // recycle pTemp
                Cut_CutRecycle( p, pTemp );
            
            } else {
                ppTail = &pTemp.pNext; }
        }
        assert(ppTail == pSuperList.ppTail[a], "ppTail == pSuperList.ppTail[a]");
        assert(*ppTail == nil, "*ppTail == nil");
    }

    return 0;
}

/**Function*************************************************************

  Synopsis    [Checks if the cut is local and can be removed.]

  Description [Returns 1 if the cut is removed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutFilterGlobal(p *Cut_Man_t, pCut *Cut_Cut_t) int32 {
    var a int32;
    if pCut.nLeaves == 1 {
        return 0; }
    for  a = 0; a < int32(pCut.nLeaves); a++  {
        if Vec_IntEntry( p.vNodeAttrs, pCut.pLeaves[a] ) { // global
            return 0; }
    }
    // there is no global nodes, the cut should be removed
    p.nCutsFilter++;
    Cut_CutRecycle( p, pCut );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Checks containment for one cut.]

  Description [Returns 1 if the cut is removed.]
               
  SideEffects [May remove other cuts in the set.]

  SeeAlso     []

***********************************************************************/
func Cut_CutFilterOld(p *Cut_Man_t, pList *Cut_Cut_t, pCut *Cut_Cut_t) int32 {
    var pPrev *Cut_Cut_t;
    var pTemp *Cut_Cut_t;
    var pTemp2 *Cut_Cut_t;
    var ppTail **Cut_Cut_t;

    // check if this cut is filtered out by smaller cuts
    pPrev = nil;
    // inlining of macro Cut_ListForEachCut
    for  pTemp = pList;                                             (pTemp != nil);                                                     pTemp = pTemp.pNext  {
        if pTemp.nLeaves > pCut.nLeaves {
            break; }
        pPrev = pTemp;
        // skip the non-contained cuts
        if (pTemp.uSign & pCut.uSign) != pTemp.uSign {
            continue; }
        // check containment seriously
        if Cut_CutCheckDominance( pTemp, pCut ) != 0 {
            p.nCutsFilter++;
            Cut_CutRecycle( p, pCut );
            return 1;
        }
    }
    assert(pPrev.pNext == pTemp, "pPrev.pNext == pTemp");

    // filter out other cuts using this one
    ppTail = &pPrev.pNext;
    // inlining of macro Cut_ListForEachCutSafe
    for  pTemp = pTemp,                                             pTemp2 = pTemp? pTemp.pNext: nil;                          (pTemp != nil);                                                     pTemp = pTemp2,                                             pTemp2 = pTemp? pTemp.pNext: nil  {
        // skip the non-contained cuts
        if (pTemp.uSign & pCut.uSign) != pCut.uSign {
            ppTail = &pTemp.pNext;
            continue;
        }
        // check containment seriously
        if Cut_CutCheckDominance( pCut, pTemp ) != 0 {
            p.nCutsFilter++;
            p.nNodeCuts--;
            // skip the given cut in the list
            *ppTail = pTemp.pNext;
            // recycle pTemp
            Cut_CutRecycle( p, pTemp );
        
        } else {
            ppTail = &pTemp.pNext; }
    }
    assert(*ppTail == nil, "*ppTail == nil");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Processes two cuts.]

  Description [Returns 1 if the limit has been reached.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutProcessTwo(p *Cut_Man_t, pCut0 *Cut_Cut_t, pCut1 *Cut_Cut_t, pSuperList *Cut_List_t) int32 {
    var pCut *Cut_Cut_t;
    // merge the cuts
    if pCut0.nLeaves >= pCut1.nLeaves {
        pCut = Cut_CutMergeTwo( p, pCut0, pCut1 );
    } else {
        pCut = Cut_CutMergeTwo( p, pCut1, pCut0 ); }
    if pCut == nil {
        return 0; }
    assert(p.pParams.fSeq || pCut.nLeaves > 1, "p.pParams.fSeq || pCut.nLeaves > 1");
    // set the signature
    pCut.uSign = pCut0.uSign | pCut1.uSign;
    if p.pParams.fRecord {
        pCut.Num0 = pCut0.Num0, pCut.Num1 = pCut1.Num0; }
    // check containment
    if p.pParams.fFilter {
        if Cut_CutFilterOne(p, pSuperList, pCut) != 0 {
//        if ( Cut_CutFilterOneEqual(p, pSuperList, pCut) )
            return 0; }
        if p.pParams.fSeq {
            if p.pCompareOld && (Cut_CutFilterOld(p, p.pCompareOld, pCut) != 0) {
                return 0; }
            if p.pCompareNew && (Cut_CutFilterOld(p, p.pCompareNew, pCut) != 0) {
                return 0; }
        }
    }

    if p.pParams.fGlobal {
        assert(p.vNodeAttrs != nil, "p.vNodeAttrs != nil");
        if Cut_CutFilterGlobal( p, pCut ) != 0 {
            return 0; }
    }

    // compute the truth table
    if p.pParams.fTruth {
        Cut_TruthCompute( p, pCut, pCut0, pCut1, p.fCompl0, p.fCompl1 ); }
    // add to the list
    Cut_ListAdd( pSuperList, pCut );
    // return status (0 if okay; 1 if exceeded the limit)
    return ++p.nNodeCuts == p.pParams.nKeepMax;
}
 
/**Function*************************************************************

  Synopsis    [Computes the cuts by merging cuts at two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeComputeCuts(p *Cut_Man_t, Node int32, Node0 int32, Node1 int32, fCompl0 int32, fCompl1 int32, fTriv int32, TreeCode int32) *Cut_Cut_t {
    var Super Cut_List_t;
    var pSuper *Cut_List_t = &Super;
    var pList, pCut *Cut_Cut_t;
    var clk abctime;
    // start the number of cuts at the node
    p.nNodes++;
    p.nNodeCuts = 0;
    // prepare information for recording
    if p.pParams.fRecord {
        Cut_CutNumberList( Cut_NodeReadCutsNew(p, Node0) );
        Cut_CutNumberList( Cut_NodeReadCutsNew(p, Node1) );
    }
    // compute the cuts
clk = Abc_Clock();
    Cut_ListStart( pSuper );
    Cut_NodeDoComputeCuts( p, pSuper, Node, fCompl0, fCompl1, Cut_NodeReadCutsNew(p, Node0), Cut_NodeReadCutsNew(p, Node1), fTriv, TreeCode );
    pList = Cut_ListFinish( pSuper );
p.timeMerge += Abc_Clock() - clk;
    // verify the result of cut computation
//    Cut_CutListVerify( pList );
    // performing the recording
    if p.pParams.fRecord {
        Vec_IntWriteEntry( p.vNodeStarts, Node, Vec_IntSize(p.vCutPairs) );
        // inlining of macro Cut_ListForEachCut
        for  pCut = pList;                                             (pCut != nil);                                                     pCut = pCut.pNext  {
            Vec_IntPush( p.vCutPairs, ((pCut.Num1 << 16) | pCut.Num0) );
        }
        Vec_IntWriteEntry( p.vNodeCuts, Node, Vec_IntSize(p.vCutPairs) - Vec_IntEntry(p.vNodeStarts, Node) );
    }
    if p.pParams.fRecordAig {
        extern void Aig_RManRecord( unsigned * pTruth, int nVarsInit );
        // inlining of macro Cut_ListForEachCut
        for  pCut = pList;                                             (pCut != nil);                                                     pCut = pCut.pNext  {
            if Cut_CutReadLeaveNum(pCut) > 4 {
                Aig_RManRecord( Cut_CutReadTruth(pCut), Cut_CutReadLeaveNum(pCut) ); }
        }
    }
    // check if the node is over the list
    if p.nNodeCuts == p.pParams.nKeepMax {
        p.nCutsLimit++; }
    // set the list at the node
    Vec_PtrFillExtra( p.vCutsNew, Node + 1, nil );
    assert(Cut_NodeReadCutsNew(p, Node) == nil, "Cut_NodeReadCutsNew(p, Node) == nil");
    /////
//    pList->pNext = NULL;
    /////
    Cut_NodeWriteCutsNew( p, Node, pList );
    // filter the cuts
//clk = Abc_Clock();
//    if ( p->pParams->fFilter )
//        Cut_CutFilter( p, pList0 );
//p->timeFilter += Abc_Clock() - clk;
    // perform mapping of this node with these cuts
clk = Abc_Clock();
    if p.pParams.fMap && !p.pParams.fSeq {
//        int Delay1, Delay2;
//        Delay1 = Cut_NodeMapping( p, pList, Node, Node0, Node1 );    
//        Delay2 = Cut_NodeMapping2( p, pList, Node, Node0, Node1 );   
//        assert( Delay1 >= Delay2 );
        Cut_NodeMapping( p, pList, Node, Node0, Node1 );
    }
p.timeMap += Abc_Clock() - clk;
    return pList;
}

/**Function*************************************************************

  Synopsis    [Returns optimum delay mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeMapping2(p *Cut_Man_t, pCuts *Cut_Cut_t, Node int32, Node0 int32, Node1 int32) int32 {
    var pCut *Cut_Cut_t;
    var DelayMin, DelayCur, i int32;
    if pCuts == nil {
        p.nDelayMin = -1; }
    if p.nDelayMin == -1 {
        return -1; }
    DelayMin = 1000000;
    // inlining of macro Cut_ListForEachCut
    for  pCut = pCuts;                                             (pCut != nil);                                                     pCut = pCut.pNext  {
        if pCut.nLeaves == 1 {
            continue; }
        DelayCur = 0;
        for  i = 0; i < int32(pCut.nLeaves); i++  {
            if DelayCur < Vec_IntEntry(p.vDelays, pCut.pLeaves[i]) {
                DelayCur = Vec_IntEntry(p.vDelays, pCut.pLeaves[i]); }
        }
        if DelayMin > DelayCur {
            DelayMin = DelayCur; }
    }
    if DelayMin == 1000000 {
         p.nDelayMin = -1;
         return -1;
    }
    DelayMin++;
    Vec_IntWriteEntry( p.vDelays, Node, DelayMin );
    if p.nDelayMin < DelayMin {
        p.nDelayMin = DelayMin; }
    return DelayMin;
}

/**Function*************************************************************

  Synopsis    [Returns optimum delay mapping using the largest cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeMapping(p *Cut_Man_t, pCuts *Cut_Cut_t, Node int32, Node0 int32, Node1 int32) int32 {
    var pCut0, pCut1, pCut *Cut_Cut_t;
    var Delay0, Delay1, Delay int32;
    // get the fanin cuts
    Delay0 = Vec_IntEntry( p.vDelays2, Node0 );
    Delay1 = Vec_IntEntry( p.vDelays2, Node1 );
    if Delay0 == 0 {
        pCut0 = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsNew, Node0 );
    } else {
        pCut0 = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsMax, Node0 );
    }
    if Delay1 == 0 {
        pCut1 = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsNew, Node1 );
    } else {
        pCut1 = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsMax, Node1 );
    }
    if Delay0 == Delay1 {
        if Delay0 == 0 {
            Delay = Delay0 + 1;
        } else {
            Delay = Delay0;
        }
    } else if Delay0 > Delay1 {
        Delay = Delay0;
        pCut1 = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsNew, Node1 );
        assert(pCut1.nLeaves == 1, "pCut1.nLeaves == 1");
    
    } else { // if ( Delay0 < Delay1 )
    
        Delay = Delay1;
        pCut0 = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsNew, Node0 );
        assert(pCut0.nLeaves == 1, "pCut0.nLeaves == 1");
    }
    // merge the cuts
    if pCut0.nLeaves < pCut1.nLeaves {
        pCut  = Cut_CutMergeTwo( p, pCut1, pCut0 );
    } else {
        pCut  = Cut_CutMergeTwo( p, pCut0, pCut1 ); }
    if pCut == nil {
        Delay++;
        pCut = Cut_CutAlloc( p );
        pCut.nLeaves = 2;
        pCut.pLeaves[0] = Node0 < Node1 ? Node0 : Node1;
        pCut.pLeaves[1] = Node0 < Node1 ? Node1 : Node0;
    }
    assert(Delay > 0, "Delay > 0");
    Vec_IntWriteEntry( p.vDelays2, Node, Delay );
    Vec_PtrWriteEntry( p.vCutsMax, Node, pCut );
    if p.nDelayMin < Delay {
        p.nDelayMin = Delay; }
    return Delay;
}

/**Function*************************************************************

  Synopsis    [Computes area after mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManMappingArea_rec(p *Cut_Man_t, Node int32) int32 {
    var pCut *Cut_Cut_t;
    var i, Counter int32;
    if p.vCutsMax == nil {
        return 0; }
    pCut = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsMax, Node );
    if pCut == nil || pCut.nLeaves == 1 {
        return 0; }
    Counter = 0;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Counter += Cut_ManMappingArea_rec( p, pCut.pLeaves[i] );
    }
    Vec_PtrWriteEntry( p.vCutsMax, Node, nil );
    return 1 + Counter;
}


/**Function*************************************************************

  Synopsis    [Computes the cuts by merging cuts at two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeDoComputeCuts(p *Cut_Man_t, pSuper *Cut_List_t, Node int32, fCompl0 int32, fCompl1 int32, pList0 *Cut_Cut_t, pList1 *Cut_Cut_t, fTriv int32, TreeCode int32) {
    var pStop0, pStop1, pTemp0, pTemp1 *Cut_Cut_t;
    var pStore0 *Cut_Cut_t = nil;
    var pStore1 *Cut_Cut_t = nil; // Suppress "might be used uninitialized"
    var i, nCutsOld, Limit int32;
    // start with the elementary cut
    if fTriv != 0 {
//        printf( "Creating trivial cut %d.\n", Node );
        pTemp0 = Cut_CutCreateTriv( p, Node );
        Cut_ListAdd( pSuper, pTemp0 );
        p.nNodeCuts++;
    }
    // get the cut lists of children
    if pList0 == nil || pList1 == nil || (p.pParams.fLocal && (TreeCode != 0)) {
        return; }

    // remember the old number of cuts
    nCutsOld = p.nCutsCur;
    Limit = p.pParams.nVarsMax;
    // get the simultation bit of the node
    p.fSimul = (fCompl0 ^ pList0.fSimul) & (fCompl1 ^ pList1.fSimul);
    // set temporary variables
    p.fCompl0 = fCompl0;
    p.fCompl1 = fCompl1;
    // if tree cuts are computed, make sure only the unit cuts propagate over the DAG nodes
    if TreeCode & 1 {
        assert(pList0.nLeaves == 1, "pList0.nLeaves == 1");
        pStore0 = pList0.pNext;
        pList0.pNext = nil;
    }
    if TreeCode & 2 {
        assert(pList1.nLeaves == 1, "pList1.nLeaves == 1");
        pStore1 = pList1.pNext;
        pList1.pNext = nil;
    }
    // find the point in the list where the max-var cuts begin
    // inlining of macro Cut_ListForEachCut
    for  pStop0 = pList0;                                             pStop0;                                                     pStop0 = pStop0.pNext  {
        if pStop0.nLeaves == uint32(Limit) {
            break; }
    }
    // inlining of macro Cut_ListForEachCut
    for  pStop1 = pList1;                                             pStop1;                                                     pStop1 = pStop1.pNext  {
        if pStop1.nLeaves == uint32(Limit) {
            break; }
    }

    // small by small
    // inlining of macro Cut_ListForEachCutStop
    for  pTemp0 = pList0;                                             pTemp0 != pStop0;                                            pTemp0 = pTemp0.pNext  {
    // inlining of macro Cut_ListForEachCutStop
    for  pTemp1 = pList1;                                             pTemp1 != pStop1;                                            pTemp1 = pTemp1.pNext  {
        if Cut_CutProcessTwo( p, pTemp0, pTemp1, pSuper ) != 0 {
            goto Quits; }
    }
    }
    // small by large
    // inlining of macro Cut_ListForEachCutStop
    for  pTemp0 = pList0;                                             pTemp0 != pStop0;                                            pTemp0 = pTemp0.pNext  {
    // inlining of macro Cut_ListForEachCut
    for  pTemp1 = pStop1;                                             (pTemp1 != nil);                                                     pTemp1 = pTemp1.pNext  {
        if (pTemp0.uSign & pTemp1.uSign) != pTemp0.uSign {
            continue; }
        if Cut_CutProcessTwo( p, pTemp0, pTemp1, pSuper ) != 0 {
            goto Quits; }
    }
    }
    // small by large
    // inlining of macro Cut_ListForEachCutStop
    for  pTemp1 = pList1;                                             pTemp1 != pStop1;                                            pTemp1 = pTemp1.pNext  {
    // inlining of macro Cut_ListForEachCut
    for  pTemp0 = pStop0;                                             pTemp0;                                                     pTemp0 = pTemp0.pNext  {
        if (pTemp0.uSign & pTemp1.uSign) != pTemp1.uSign {
            continue; }
        if Cut_CutProcessTwo( p, pTemp0, pTemp1, pSuper ) != 0 {
            goto Quits; }
    }
    }
    // large by large
    // inlining of macro Cut_ListForEachCut
    for  pTemp0 = pStop0;                                             pTemp0;                                                     pTemp0 = pTemp0.pNext  {
    // inlining of macro Cut_ListForEachCut
    for  pTemp1 = pStop1;                                             (pTemp1 != nil);                                                     pTemp1 = pTemp1.pNext  {
        assert(pTemp0.nLeaves == uint32(Limit) && pTemp1.nLeaves == uint32(Limit), "pTemp0.nLeaves == uint32(Limit) && pTemp1.nLeaves == uint32(Limit)");
        if pTemp0.uSign != pTemp1.uSign {
            continue; }
        for  i = 0; i < Limit; i++  {
            if pTemp0.pLeaves[i] != pTemp1.pLeaves[i] {
                break; }
        }
        if i < Limit {
            continue; }
        if Cut_CutProcessTwo( p, pTemp0, pTemp1, pSuper ) != 0 {
            goto Quits; }
    }
    } 
    if p.nNodeCuts == 0 {
        p.nNodesNoCuts++; }
Quits:
    if TreeCode & 1 {
        pList0.pNext = pStore0; }
    if TreeCode & 2 {
        pList1.pNext = pStore1; }
}

/**Function*************************************************************

  Synopsis    [Computes the cuts by unioning cuts at a choice node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeUnionCuts(p *Cut_Man_t, vNodes *Vec_Int_t) *Cut_Cut_t {
    var Super Cut_List_t;
    var pSuper *Cut_List_t = &Super;
    var pList, pListStart, pCut, pCut2 *Cut_Cut_t;
    var pTop *Cut_Cut_t  = nil; // Suppress "might be used uninitialized"
    var i int32;
    var k int32;
    var Node int32;
    var Root int32;
    var Limit int32 = p.pParams.nVarsMax;
    var clk abctime  = Abc_Clock();

    // start the new list
    Cut_ListStart( pSuper );

    // remember the root node to save the resulting cuts
    Root = Vec_IntEntry( vNodes, 0 );
    p.nNodeCuts = 1;

    // collect small cuts first
    Vec_PtrClear( p.vTemp );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((Node) = Vec_IntEntry(vNodes, i)), 1); i++  {
        // get the cuts of this node
        pList = Cut_NodeReadCutsNew( p, Node );
        Cut_NodeWriteCutsNew( p, Node, nil );
        assert((pList != nil), "pList");
        // remember the starting point
        pListStart = pList.pNext;
        pList.pNext = nil;
        // save or recycle the elementary cut
        if i == 0 {
            Cut_ListAdd( pSuper, pList );
pTop = pList;
        } else {
            Cut_CutRecycle( p, pList ); }
        // save all the cuts that are smaller than the limit
        // inlining of macro Cut_ListForEachCutSafe
        for  pCut = pListStart,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
            if pCut.nLeaves == uint32(Limit) {
                Vec_PtrPush( p.vTemp, pCut );
                break;
            }
            // check containment
            if p.pParams.fFilter && (Cut_CutFilterOne( p, pSuper, pCut ) != 0) {
                continue; }
            // set the complemented bit by comparing the first cut with the current cut
            pCut.fCompl = pTop.fSimul ^ pCut.fSimul;
            pListStart = pCut.pNext;
            pCut.pNext = nil;
            // add to the list
            Cut_ListAdd( pSuper, pCut );
            if ++p.nNodeCuts == p.pParams.nKeepMax {
                // recycle the rest of the cuts of this node
                // inlining of macro Cut_ListForEachCutSafe
                for  pCut = pListStart,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
                    Cut_CutRecycle( p, pCut );
                }
                // recycle all cuts of other nodes
                // inlining of macro Vec_IntForEachEntryStart
                for  k = i+1; (k < Vec_IntSize(vNodes)) && (((Node) = Vec_IntEntry(vNodes, k)), 1); k++  {
                    Cut_NodeFreeCuts( p, Node );
                }
                // recycle the saved cuts of other nodes
                // inlining of macro Vec_PtrForEachEntry
                for  k = 0; (k < Vec_PtrSize(p.vTemp)) && (((pList) = (Cut_Cut_t *)Vec_PtrEntry(p.vTemp, k)), 1); k++  {
                    // inlining of macro Cut_ListForEachCutSafe
                    for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
                        Cut_CutRecycle( p, pCut );
                    }
                }
                goto finish;
            }
        }
    } 
    // collect larger cuts next
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTemp)) && (((pList) = (Cut_Cut_t *)Vec_PtrEntry(p.vTemp, i)), 1); i++  {
        // inlining of macro Cut_ListForEachCutSafe
        for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
            // check containment
            if p.pParams.fFilter && (Cut_CutFilterOne( p, pSuper, pCut ) != 0) {
                continue; }
            // set the complemented bit
            pCut.fCompl = pTop.fSimul ^ pCut.fSimul;
            pListStart = pCut.pNext;
            pCut.pNext = nil;
            // add to the list
            Cut_ListAdd( pSuper, pCut );
            if ++p.nNodeCuts == p.pParams.nKeepMax {
                // recycle the rest of the cuts
                // inlining of macro Cut_ListForEachCutSafe
                for  pCut = pListStart,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
                    Cut_CutRecycle( p, pCut );
                }
                // recycle the saved cuts of other nodes
                // inlining of macro Vec_PtrForEachEntryStart
                for  k = i+1; (k < Vec_PtrSize(p.vTemp)) && (((pList) = (Cut_Cut_t *)Vec_PtrEntry(p.vTemp, k)), 1); k++  {
                    // inlining of macro Cut_ListForEachCutSafe
                    for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
                        Cut_CutRecycle( p, pCut );
                    }
                }
                goto finish;
            }
        }
    }
finish :
    // set the cuts at the node
    assert(Cut_NodeReadCutsNew(p, Root) == nil, "Cut_NodeReadCutsNew(p, Root) == nil");
    pList = Cut_ListFinish( pSuper );
    Cut_NodeWriteCutsNew( p, Root, pList );
p.timeUnion += Abc_Clock() - clk;
    // filter the cuts
//clk = Abc_Clock();
//    if ( p->pParams->fFilter )
//        Cut_CutFilter( p, pList );
//p->timeFilter += Abc_Clock() - clk;
    p.nNodes -= vNodes.nSize - 1;
    return pList;
}

/**Function*************************************************************

  Synopsis    [Computes the cuts by unioning cuts at a choice node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_NodeUnionCutsSeq(p *Cut_Man_t, vNodes *Vec_Int_t, CutSetNum int32, fFirst int32) *Cut_Cut_t {
    var Super Cut_List_t;
    var pSuper *Cut_List_t = &Super;
    var pList, pListStart, pCut, pCut2, pTop *Cut_Cut_t;
    var i int32;
    var k int32;
    var Node int32;
    var Root int32;
    var Limit int32 = p.pParams.nVarsMax;
    var clk abctime  = Abc_Clock();

    // start the new list
    Cut_ListStart( pSuper );

    // remember the root node to save the resulting cuts
    Root = Vec_IntEntry( vNodes, 0 );
    p.nNodeCuts = 1;

    // store the original lists for comparison
    p.pCompareOld = Cut_NodeReadCutsOld( p, Root );
    if CutSetNum >= 0 {
        p.pCompareNew = Cut_NodeReadCutsNew( p, Root );
    } else {
        p.pCompareNew = nil;
    }

    // get the topmost cut
    pTop = nil;
    if (pTop = Cut_NodeReadCutsOld( p, Root )) == nil {
        pTop = Cut_NodeReadCutsNew( p, Root ); }
    assert(pTop != nil, "pTop != nil");

    // collect small cuts first
    Vec_PtrClear( p.vTemp );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((Node) = Vec_IntEntry(vNodes, i)), 1); i++  {
        // get the cuts of this node
        if i == 0 && CutSetNum >= 0 {
            pList = Cut_NodeReadCutsTemp( p, CutSetNum );
            Cut_NodeWriteCutsTemp( p, CutSetNum, nil );
        
        } else {
            pList = Cut_NodeReadCutsNew( p, Node );
            Cut_NodeWriteCutsNew( p, Node, nil );
        }
        if pList == nil {
            continue; }

        // process the cuts
        if fFirst != 0 {
            // remember the starting point
            pListStart = pList.pNext;
            pList.pNext = nil;
            // save or recycle the elementary cut
            if i == 0 {
                Cut_ListAdd( pSuper, pList );
            } else {
                Cut_CutRecycle( p, pList ); }
        
        } else {
            pListStart = pList; }

        // save all the cuts that are smaller than the limit
        // inlining of macro Cut_ListForEachCutSafe
        for  pCut = pListStart,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
            if pCut.nLeaves == uint32(Limit) {
                Vec_PtrPush( p.vTemp, pCut );
                break;
            }
            // check containment
//            if ( p->pParams->fFilter && Cut_CutFilterOne( p, pSuper, pCut ) )
//                continue;
            if p.pParams.fFilter {
                if Cut_CutFilterOne(p, pSuper, pCut) != 0 {
                    continue; }
                if p.pParams.fSeq {
                    if p.pCompareOld && (Cut_CutFilterOld(p, p.pCompareOld, pCut) != 0) {
                        continue; }
                    if p.pCompareNew && (Cut_CutFilterOld(p, p.pCompareNew, pCut) != 0) {
                        continue; }
                }
            }

            // set the complemented bit by comparing the first cut with the current cut
            pCut.fCompl = pTop.fSimul ^ pCut.fSimul;
            pListStart = pCut.pNext;
            pCut.pNext = nil;
            // add to the list
            Cut_ListAdd( pSuper, pCut );
            if ++p.nNodeCuts == p.pParams.nKeepMax {
                // recycle the rest of the cuts of this node
                // inlining of macro Cut_ListForEachCutSafe
                for  pCut = pListStart,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
                    Cut_CutRecycle( p, pCut );
                }
                // recycle all cuts of other nodes
                // inlining of macro Vec_IntForEachEntryStart
                for  k = i+1; (k < Vec_IntSize(vNodes)) && (((Node) = Vec_IntEntry(vNodes, k)), 1); k++  {
                    Cut_NodeFreeCuts( p, Node );
                }
                // recycle the saved cuts of other nodes
                // inlining of macro Vec_PtrForEachEntry
                for  k = 0; (k < Vec_PtrSize(p.vTemp)) && (((pList) = (Cut_Cut_t *)Vec_PtrEntry(p.vTemp, k)), 1); k++  {
                    // inlining of macro Cut_ListForEachCutSafe
                    for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
                        Cut_CutRecycle( p, pCut );
                    }
                }
                goto finish;
            }
        }
    } 
    // collect larger cuts next
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTemp)) && (((pList) = (Cut_Cut_t *)Vec_PtrEntry(p.vTemp, i)), 1); i++  {
        // inlining of macro Cut_ListForEachCutSafe
        for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
            // check containment
//            if ( p->pParams->fFilter && Cut_CutFilterOne( p, pSuper, pCut ) )
//                continue;
            if p.pParams.fFilter {
                if Cut_CutFilterOne(p, pSuper, pCut) != 0 {
                    continue; }
                if p.pParams.fSeq {
                    if p.pCompareOld && (Cut_CutFilterOld(p, p.pCompareOld, pCut) != 0) {
                        continue; }
                    if p.pCompareNew && (Cut_CutFilterOld(p, p.pCompareNew, pCut) != 0) {
                        continue; }
                }
            }

            // set the complemented bit
            pCut.fCompl = pTop.fSimul ^ pCut.fSimul;
            pListStart   = pCut.pNext;
            pCut.pNext  = nil;
            // add to the list
            Cut_ListAdd( pSuper, pCut );
            if ++p.nNodeCuts == p.pParams.nKeepMax {
                // recycle the rest of the cuts
                // inlining of macro Cut_ListForEachCutSafe
                for  pCut = pListStart,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
                    Cut_CutRecycle( p, pCut );
                }
                // recycle the saved cuts of other nodes
                // inlining of macro Vec_PtrForEachEntryStart
                for  k = i+1; (k < Vec_PtrSize(p.vTemp)) && (((pList) = (Cut_Cut_t *)Vec_PtrEntry(p.vTemp, k)), 1); k++  {
                    // inlining of macro Cut_ListForEachCutSafe
                    for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
                        Cut_CutRecycle( p, pCut );
                    }
                }
                goto finish;
            }
        }
    }
finish :
    // set the cuts at the node
    pList = Cut_ListFinish( pSuper );

    // set the lists at the node
//    assert( Cut_NodeReadCutsNew(p, Root) == NULL );
//    Cut_NodeWriteCutsNew( p, Root, pList );
    if CutSetNum >= 0 {
        assert(Cut_NodeReadCutsTemp(p, CutSetNum) == nil, "Cut_NodeReadCutsTemp(p, CutSetNum) == nil");
        Cut_NodeWriteCutsTemp( p, CutSetNum, pList );
    
    } else {
        assert(Cut_NodeReadCutsNew(p, Root) == nil, "Cut_NodeReadCutsNew(p, Root) == nil");
        Cut_NodeWriteCutsNew( p, Root, pList );
    }

p.timeUnion += Abc_Clock() - clk;
    // filter the cuts
//clk = Abc_Clock();
//    if ( p->pParams->fFilter )
//        Cut_CutFilter( p, pList );
//p->timeFilter += Abc_Clock() - clk;
//    if ( fFirst )
//        p->nNodes -= vNodes->nSize - 1;
    return pList;
}


/**Function*************************************************************

  Synopsis    [Verifies that the list contains only non-dominated cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutListVerify(pList *Cut_Cut_t) int32 { 
    var pCut, pDom *Cut_Cut_t;
    // inlining of macro Cut_ListForEachCut
    for  pCut = pList;                                             (pCut != nil);                                                     pCut = pCut.pNext  {
        // inlining of macro Cut_ListForEachCutStop
        for  pDom = pList;                                             pDom != pCut;                                            pDom = pDom.pNext  {
            if Cut_CutCheckDominance( pDom, pCut ) != 0 {
                fmt.Printf("******************* These are contained cuts:\n");
                Cut_CutPrint( pDom, 1 );
                Cut_CutPrint( pDom, 1 );
                return 0;
            }
        }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

