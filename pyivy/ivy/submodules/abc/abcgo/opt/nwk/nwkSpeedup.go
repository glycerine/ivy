package nwk

import "fmt"

/**CFile****************************************************************

  FileName    [nwkSpeedup.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Netlist representation.]

  Synopsis    [Global delay optimization using structural choices.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkSpeedup.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nwk.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Adds strashed nodes for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSpeedupNode_rec(pAig *Aig_Man_t, pNode *Aig_Obj_t, vNodes *Vec_Ptr_t) int32 {
    if Aig_ObjIsTravIdCurrent(pAig, pNode) {
        return 1; }
    if Aig_ObjIsCi(pNode) {
        return 0; }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    Aig_ObjSetTravIdCurrent( pAig, pNode );
    if !(Aig_ManSpeedupNode_rec( pAig, Aig_ObjFanin0(pNode), vNodes ) != 0) {
        return 0; }
    if !(Aig_ManSpeedupNode_rec( pAig, Aig_ObjFanin1(pNode), vNodes ) != 0) {
        return 0; }
    Vec_PtrPush( vNodes, pNode );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds strashed nodes for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSpeedupNode(pNtk *Nwk_Man_t, pAig *Aig_Man_t, pNode *Nwk_Obj_t, vLeaves *Vec_Ptr_t, vTimes *Vec_Ptr_t) {
    var vNodes *Vec_Ptr_t;
    var pObj, pObj2 *Nwk_Obj_t;
    var ppCofs [32]*Aig_Obj_t;
    var pAnd *Aig_Obj_t;
    var pTemp *Aig_Obj_t;
    var nCofs, i, k, nSkip int32;

    // quit of regulars are the same
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pObj2) = (Nwk_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
        if i != k && Aig_Regular((Aig_Obj_t *)pObj.pCopy) == Aig_Regular((Aig_Obj_t *)pObj2.pCopy) {
//            printf( "Identical after structural hashing!!!\n" );
            return;
        }
    }
    }

    // collect the AIG nodes
    vNodes = Vec_PtrAlloc( 100 );
    Aig_ManIncrementTravId( pAig );
    Aig_ObjSetTravIdCurrent( pAig, Aig_ManConst1(pAig) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pAnd = (Aig_Obj_t *)pObj.pCopy;
        Aig_ObjSetTravIdCurrent( pAig, Aig_Regular(pAnd) );
    }
    // traverse from the root node
    pAnd = (Aig_Obj_t *)pNode.pCopy;
    if !(Aig_ManSpeedupNode_rec( pAig, Aig_Regular(pAnd), vNodes ) != 0) {
//        printf( "Bad node!!!\n" );
        Vec_PtrFree( vNodes );
        return;
    }

    // derive cofactors
    nCofs = (1 << Vec_PtrSize(vTimes));
    for  i = 0; i < nCofs; i++  {
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
            pAnd = (Aig_Obj_t *)pObj.pCopy;
            Aig_Regular(pAnd).pData = Aig_Regular(pAnd);
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vTimes)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vTimes, k)), 1); k++  {
            pAnd = (Aig_Obj_t *)pObj.pCopy;
            Aig_Regular(pAnd).pData = Aig_NotCond( Aig_ManConst1(pAig), ((i & (1<<k)) == 0) );
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes)) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, k)), 1); k++  {
            pTemp.pData = Aig_And( pAig, Aig_ObjChild0Copy(pTemp), Aig_ObjChild1Copy(pTemp) );
        }
        // save the result
        pAnd = (Aig_Obj_t *)pNode.pCopy;
        ppCofs[i] = Aig_NotCond( (Aig_Obj_t *)Aig_Regular(pAnd).pData, Aig_IsComplement(pAnd) );
    }
    Vec_PtrFree( vNodes );

//Nwk_ObjAddFanin( Nwk_ManCreatePo(pAig), ppCofs[0] );
//Nwk_ObjAddFanin( Nwk_ManCreatePo(pAig), ppCofs[1] );

    // collect the resulting tree
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vTimes)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vTimes, k)), 1); k++  {
        for  nSkip = (1<<k), i = 0; i < nCofs; i += 2*nSkip  {
            pAnd = (Aig_Obj_t *)pObj.pCopy;
            ppCofs[i] = Aig_Mux( pAig, Aig_Regular(pAnd), ppCofs[i+nSkip], ppCofs[i] );
        }
    }
//Nwk_ObjAddFanin( Nwk_ManCreatePo(pAig), ppCofs[0] );

    // create choice node
    pAnd  = Aig_Regular((Aig_Obj_t *)pNode.pCopy); // repr
    pTemp = Aig_Regular(ppCofs[0]);    // new
    if Aig_ObjEquiv(pAig, pAnd) == nil && Aig_ObjEquiv(pAig, pTemp) == nil && !Aig_ObjCheckTfi(pAig, pTemp, pAnd) {
        pAig.pEquivs[pAnd.Id] = pTemp; }
}

/**Function*************************************************************

  Synopsis    [Determines timing-critical edges of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDelayTraceTCEdges(pNtk *Nwk_Man_t, pNode *Nwk_Obj_t, tDelta float32, fUseLutLib int32) uint32 {
    var pPinPerm [32]int32;
    var pPinDelays [32]float32;
    var pLutLib *If_LibLut_t;
    if fUseLutLib != 0 {
        pLutLib = pNtk.pLutLib;
    } else {
        pLutLib = nil;
    }
    var pFanin *Nwk_Obj_t;
    var uResult uint32;
    var tRequired float32;
    var pDelays *float32;
    var k int32;
    tRequired = Nwk_ObjRequired(pNode);
    if pLutLib == nil {
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[k]); k++  {
            if tRequired < Nwk_ObjArrival(pFanin) + 1.0 + tDelta {
                uResult |= (1 << k); }
        }
    
    } else if !pLutLib.fVarPinDelays {
        pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pNode)];
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[k]); k++  {
            if tRequired < Nwk_ObjArrival(pFanin) + pDelays[0] + tDelta {
                uResult |= (1 << k); }
        }
    
    } else {
        pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pNode)];
        Nwk_ManDelayTraceSortPins( pNode, pPinPerm, pPinDelays );
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[k]); k++  {
            if tRequired < Nwk_ObjArrival(Nwk_ObjFanin(pNode,pPinPerm[k])) + pDelays[k] + tDelta {
                uResult |= (1 << pPinPerm[k]); }
        }
    }
    return uResult;
}

/**Function*************************************************************

  Synopsis    [Adds choices to speed up the network by the given percentage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManSpeedup(pNtk *Nwk_Man_t, fUseLutLib int32, Percentage int32, Degree int32, fVerbose int32, fVeryVerbose int32) *Aig_Man_t {
    var pAig, pTemp *Aig_Man_t;
    var vTimeCries, vTimeFanins *Vec_Ptr_t;
    var pNode, pFanin, pFanin2 *Nwk_Obj_t;
    var pAnd *Aig_Obj_t;
    var pTempLib *If_LibLut_t  = pNtk.pLutLib;
    var pTempTim *Tim_Man_t  = nil; 
    var tDelta, tArrival float32;
    var i, k, k2, Counter, CounterRes, nTimeCris int32;
    var puTCEdges *uint32;
    // perform delay trace
    if !(fUseLutLib != 0) {
        pNtk.pLutLib = nil;
        if pNtk.pManTime {
            pTempTim = pNtk.pManTime;
            pNtk.pManTime = Tim_ManDup( pTempTim, 1 );
        }
    }
    tArrival = Nwk_ManDelayTraceLut( pNtk );
    if fUseLutLib != 0 {
        tDelta = tArrival*Percentage/100.0;
    } else {
        tDelta = 1.0;
    }
    if fVerbose != 0 {
        fmt.Printf("Max delay = %.2f. Delta = %.2f. ", tArrival, tDelta);
        fmt.Printf("Using %s model. ", fUseLutLib? "LUT library" : "unit-delay");
        if fUseLutLib != 0 {
            fmt.Printf("Percentage = %d. ", Percentage); }
        fmt.Printf("\n");
    }
    // mark the timing critical nodes and edges
    // inlining of macro ABC_ALLOC
    puTCEdges = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Nwk_ManObjNumMax(pNtk))));
    memset( puTCEdges, 0, sizeof(unsigned) * Nwk_ManObjNumMax(pNtk) );
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pNode) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pNode) == nil || !Nwk_ObjIsNode(pNode) {} else {
        if Nwk_ObjSlack(pNode) >= tDelta {
            continue; }
        puTCEdges[pNode.Id] = Nwk_ManDelayTraceTCEdges( pNtk, pNode, tDelta, fUseLutLib );
    }
    }
    if fVerbose != 0 {
        Counter = CounterRes = 0;
        // inlining of macro Nwk_ManForEachNode
        for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pNode) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pNode) == nil || !Nwk_ObjIsNode(pNode) {} else {
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[k]); k++  {
                if !Nwk_ObjIsCi(pFanin) && Nwk_ObjSlack(pFanin) < tDelta {
                    Counter++; }
            }
            CounterRes += Aig_WordCountOnes( puTCEdges[pNode.Id] );
        }
        }
        fmt.Printf("Edges: Total = %7d. 0-slack = %7d. Critical = %7d. Ratio = %4.2f\n", Nwk_ManGetTotalFanins(pNtk), Counter, CounterRes, Counter? 1.0*CounterRes/Counter : 0.0);
    }
    // start the resulting network
    pAig = Nwk_ManStrash( pNtk );
    // inlining of macro ABC_ALLOC
    pAig.pEquivs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(3 * Aig_ManObjNumMax(pAig))));
    memset( pAig.pEquivs, 0, sizeof(Aig_Obj_t *) * 3 * Aig_ManObjNumMax(pAig) );

    // collect nodes to be used for resynthesis
    Counter = CounterRes = 0;
    vTimeCries = Vec_PtrAlloc( 16 );
    vTimeFanins = Vec_PtrAlloc( 16 );
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pNode) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pNode) == nil || !Nwk_ObjIsNode(pNode) {} else {
        if Nwk_ObjSlack(pNode) >= tDelta {
            continue; }
        // count the number of non-PI timing-critical nodes
        nTimeCris = 0;
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[k]); k++  {
            if !Nwk_ObjIsCi(pFanin) && (puTCEdges[pNode.Id] & (1<<k)) {
                nTimeCris++; }
        }
        if !(fVeryVerbose != 0) && nTimeCris == 0 {
            continue; }
        Counter++;
        // count the total number of timing critical second-generation nodes
        Vec_PtrClear( vTimeCries );
        if nTimeCris != 0 {
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[k]); k++  {
                if !Nwk_ObjIsCi(pFanin) && (puTCEdges[pNode.Id] & (1<<k)) {
                    // inlining of macro Nwk_ObjForEachFanin
                    for  k2 = 0; (k2 < int32((pFanin).nFanins)) && ((pFanin2) = (pFanin).pFanio[k2]); k2++  {
                        if puTCEdges[pFanin.Id] & (1<<k2) {
                            Vec_PtrPushUnique( vTimeCries, pFanin2 ); }
                    } }
            }
        }
//        if ( !fVeryVerbose && (Vec_PtrSize(vTimeCries) == 0 || Vec_PtrSize(vTimeCries) > Degree) )
        if Vec_PtrSize(vTimeCries) == 0 || Vec_PtrSize(vTimeCries) > Degree {
            continue; }
        CounterRes++;
        // collect second generation nodes
        Vec_PtrClear( vTimeFanins );
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[k]); k++  {
            if Nwk_ObjIsCi(pFanin) {
                Vec_PtrPushUnique( vTimeFanins, pFanin );
            } else {
                // inlining of macro Nwk_ObjForEachFanin
                for  k2 = 0; (k2 < int32((pFanin).nFanins)) && ((pFanin2) = (pFanin).pFanio[k2]); k2++  {
                    Vec_PtrPushUnique( vTimeFanins, pFanin2 );
                } }                    
        }
        // print the results
        if fVeryVerbose != 0 {
        fmt.Printf("%5d Node %5d : %d %2d %2d  ", Counter, pNode.Id, nTimeCris, Vec_PtrSize(vTimeCries), Vec_PtrSize(vTimeFanins));
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[k]); k++  {
            fmt.Printf("%d(%.2f)%s ", pFanin.Id, Nwk_ObjSlack(pFanin), (puTCEdges[pNode.Id] & (1<<k))? "*":"");
        }
        fmt.Printf("\n");
        }
        // add the node to choices
        if Vec_PtrSize(vTimeCries) == 0 || Vec_PtrSize(vTimeCries) > Degree {
            continue; }
        // order the fanins in the increasing order of criticalily
        if Vec_PtrSize(vTimeCries) > 1 {
            pFanin = (Nwk_Obj_t *)Vec_PtrEntry( vTimeCries, 0 );
            pFanin2 = (Nwk_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
            if Nwk_ObjSlack(pFanin) < Nwk_ObjSlack(pFanin2) {
                Vec_PtrWriteEntry( vTimeCries, 0, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin );
            }
        }
        if Vec_PtrSize(vTimeCries) > 2 {
            pFanin = (Nwk_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
            pFanin2 = (Nwk_Obj_t *)Vec_PtrEntry( vTimeCries, 2 );
            if Nwk_ObjSlack(pFanin) < Nwk_ObjSlack(pFanin2) {
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 2, pFanin );
            }
            pFanin = (Nwk_Obj_t *)Vec_PtrEntry( vTimeCries, 0 );
            pFanin2 = (Nwk_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
            if Nwk_ObjSlack(pFanin) < Nwk_ObjSlack(pFanin2) {
                Vec_PtrWriteEntry( vTimeCries, 0, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin );
            }
        }
        // add choice
        Aig_ManSpeedupNode( pNtk, pAig, pNode, vTimeFanins, vTimeCries );
    }
    }
    Vec_PtrFree( vTimeCries );
    Vec_PtrFree( vTimeFanins );
    puTCEdges = nil;
    if fVerbose != 0 {
        fmt.Printf("Nodes: Total = %7d. 0-slack = %7d. Workable = %7d. Ratio = %4.2f\n", Nwk_ManNodeNum(pNtk), Counter, CounterRes, Counter? 1.0*CounterRes/Counter : 0.0); } 

    // remove invalid choice nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pAnd) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pAnd) == nil || !Aig_ObjIsNode(pAnd) {} else
        if Aig_ObjEquiv(pAig, pAnd) {
            if Aig_ObjRefs(Aig_ObjEquiv(pAig, pAnd)) > 0 {
                pAig.pEquivs[pAnd.Id] = nil; }
        }
    }

    // put back the library
    if !(fUseLutLib != 0) {
        pNtk.pLutLib = pTempLib; }
    if pTempTim != nil {
        Tim_ManStop( pNtk.pManTime );
        pNtk.pManTime = pTempTim;
    }

    // reconstruct the network
    pAig = Aig_ManDupDfs( pTemp = pAig );
    Aig_ManStop( pTemp );
    // reset levels
    Aig_ManChoiceLevel( pAig );
    return pAig;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

