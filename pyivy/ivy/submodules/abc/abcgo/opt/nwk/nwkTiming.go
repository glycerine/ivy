package nwk

import "fmt"

/**CFile****************************************************************

  FileName    [nwkTiming.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [Manipulation of timing information.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkTiming.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Cleans timing information for all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCleanTiming(pNtk *Nwk_Man_t) {
    var pObj *Nwk_Obj_t;
    var i int32;
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        pObj.tArrival = pObj.tSlack = 0.0;
        pObj.tRequired = TIM_ETERNITY;
    }
    }
}

/**Function*************************************************************

  Synopsis    [Sorts the pins in the decreasing order of delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDelayTraceSortPins(pNode *Nwk_Obj_t, pPinPerm *int32, pPinDelays *float32) {
    var pFanin *Nwk_Obj_t;
    var i, j, best_i, temp int32;
    // start the trivial permutation and collect pin delays
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pNode).nFanins)) && ((pFanin) = (pNode).pFanio[i]); i++  {
        pPinPerm[i] = i;
        pPinDelays[i] = Nwk_ObjArrival(pFanin);
    }
    // selection sort the pins in the decreasible order of delays
    // this order will match the increasing order of LUT input pins
    for  i = 0; i < Nwk_ObjFaninNum(pNode)-1; i++  {
        best_i = i;
        for  j = i+1; j < Nwk_ObjFaninNum(pNode); j++  {
            if pPinDelays[pPinPerm[j]] > pPinDelays[pPinPerm[best_i]] {
                best_i = j; }
        }
        if best_i == i {
            continue; }
        temp = pPinPerm[i]; 
        pPinPerm[i] = pPinPerm[best_i]; 
        pPinPerm[best_i] = temp;
    }
    // verify
    assert(Nwk_ObjFaninNum(pNode) == 0 || pPinPerm[0] < Nwk_ObjFaninNum(pNode), "Nwk_ObjFaninNum(pNode) == 0 || pPinPerm[0] < Nwk_ObjFaninNum(pNode)");
    for  i = 1; i < Nwk_ObjFaninNum(pNode); i++  {
        assert(pPinPerm[i] < Nwk_ObjFaninNum(pNode), "pPinPerm[i] < Nwk_ObjFaninNum(pNode)");
        assert(pPinDelays[pPinPerm[i-1]] >= pPinDelays[pPinPerm[i]], "pPinDelays[pPinPerm[i-1]] >= pPinDelays[pPinPerm[i]]");
    }
}

/**Function*************************************************************

  Synopsis    [Sorts the pins in the decreasing order of delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManWhereIsPin(pFanout *Nwk_Obj_t, pFanin *Nwk_Obj_t, pPinPerm *int32) int32 {
    var i int32;
    for  i = 0; i < Nwk_ObjFaninNum(pFanout); i++  {
        if Nwk_ObjFanin(pFanout, pPinPerm[i]) == pFanin {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Computes the arrival times for the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodeComputeArrival(pObj *Nwk_Obj_t, fUseSorting int32) float32 {
    var pLutLib *If_LibLut_t  = pObj.pMan.pLutLib;
    var pPinPerm [32]int32;
    var pPinDelays [32]float32;
    var pFanin *Nwk_Obj_t;
    var tArrival float32;
    var pDelays *float32;
    var k int32;
    assert(Nwk_ObjIsNode(pObj) || Nwk_ObjIsCi(pObj) || Nwk_ObjIsCo(pObj), "Nwk_ObjIsNode(pObj) || Nwk_ObjIsCi(pObj) || Nwk_ObjIsCo(pObj)");
    if Nwk_ObjIsCi(pObj) {
        return Nwk_ObjArrival(pObj); }
    if Nwk_ObjIsCo(pObj) {
        return Nwk_ObjArrival( Nwk_ObjFanin0(pObj) ); }
    tArrival = -TIM_ETERNITY;
    if pLutLib == nil {
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  {
            if tArrival < Nwk_ObjArrival(pFanin) + 1.0 {
                tArrival = Nwk_ObjArrival(pFanin) + 1.0; }
        }
    
    } else if !pLutLib.fVarPinDelays {
        pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pObj)];
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  {
            if tArrival < Nwk_ObjArrival(pFanin) + pDelays[0] {
                tArrival = Nwk_ObjArrival(pFanin) + pDelays[0]; }
        }
    
    } else {
        pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pObj)];
        if fUseSorting != 0 {
            Nwk_ManDelayTraceSortPins( pObj, pPinPerm, pPinDelays );
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  { 
                if tArrival < Nwk_ObjArrival(Nwk_ObjFanin(pObj,pPinPerm[k])) + pDelays[k] {
                    tArrival = Nwk_ObjArrival(Nwk_ObjFanin(pObj,pPinPerm[k])) + pDelays[k]; }
            }
        
        } else {
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  {
                if tArrival < Nwk_ObjArrival(pFanin) + pDelays[k] {
                    tArrival = Nwk_ObjArrival(pFanin) + pDelays[k]; }
            }
        }
    }
    if Nwk_ObjFaninNum(pObj) == 0 {
        tArrival = 0.0; }
    return tArrival;
}

/**Function*************************************************************

  Synopsis    [Computes the required times for the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodeComputeRequired(pObj *Nwk_Obj_t, fUseSorting int32) float32 {
    var pLutLib *If_LibLut_t  = pObj.pMan.pLutLib;
    var pPinPerm [32]int32;
    var pPinDelays [32]float32;
    var pFanout *Nwk_Obj_t;
    var tRequired float32;
    var tDelay float32;
    var pDelays *float32;
    var k, iFanin int32;
    assert(Nwk_ObjIsNode(pObj) || Nwk_ObjIsCi(pObj) || Nwk_ObjIsCo(pObj), "Nwk_ObjIsNode(pObj) || Nwk_ObjIsCi(pObj) || Nwk_ObjIsCo(pObj)");
    if Nwk_ObjIsCo(pObj) {
        return Nwk_ObjRequired(pObj); }
    tRequired = TIM_ETERNITY;
    if pLutLib == nil {
        // inlining of macro Nwk_ObjForEachFanout
        for  k = 0; (k < int32((pObj).nFanouts)) && ((pFanout) = (pObj).pFanio[(pObj).nFanins+k]); k++  {
            if Nwk_ObjIsCo(pFanout) {
                tDelay = 0.0;
            } else {
                tDelay = 1.0;
            }
            if tRequired > Nwk_ObjRequired(pFanout) - tDelay {
                tRequired = Nwk_ObjRequired(pFanout) - tDelay; }
        }
    
    } else if !pLutLib.fVarPinDelays {
        // inlining of macro Nwk_ObjForEachFanout
        for  k = 0; (k < int32((pObj).nFanouts)) && ((pFanout) = (pObj).pFanio[(pObj).nFanins+k]); k++  {
            pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pFanout)];
            if Nwk_ObjIsCo(pFanout) {
                tDelay = 0.0;
            } else {
                tDelay = pDelays[0];
            }
            if tRequired > Nwk_ObjRequired(pFanout) - tDelay {
                tRequired = Nwk_ObjRequired(pFanout) - tDelay; }
        }
    
    } else {
        if fUseSorting != 0 {
            // inlining of macro Nwk_ObjForEachFanout
            for  k = 0; (k < int32((pObj).nFanouts)) && ((pFanout) = (pObj).pFanio[(pObj).nFanins+k]); k++  {
                pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pFanout)];
                Nwk_ManDelayTraceSortPins( pFanout, pPinPerm, pPinDelays );
                iFanin = Nwk_ManWhereIsPin( pFanout, pObj, pPinPerm );
                assert(Nwk_ObjFanin(pFanout,pPinPerm[iFanin]) == pObj, "Nwk_ObjFanin(pFanout,pPinPerm[iFanin]) == pObj");
                if Nwk_ObjIsCo(pFanout) {
                    tDelay = 0.0;
                } else {
                    tDelay = pDelays[iFanin];
                }
                if tRequired > Nwk_ObjRequired(pFanout) - tDelay {
                    tRequired = Nwk_ObjRequired(pFanout) - tDelay; }
            }
        
        } else {
            // inlining of macro Nwk_ObjForEachFanout
            for  k = 0; (k < int32((pObj).nFanouts)) && ((pFanout) = (pObj).pFanio[(pObj).nFanins+k]); k++  {
                pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pFanout)];
                iFanin = Nwk_ObjFindFanin( pFanout, pObj );
                assert(Nwk_ObjFanin(pFanout,iFanin) == pObj, "Nwk_ObjFanin(pFanout,iFanin) == pObj");
                if Nwk_ObjIsCo(pFanout) {
                    tDelay = 0.0;
                } else {
                    tDelay = pDelays[iFanin];
                }
                if tRequired > Nwk_ObjRequired(pFanout) - tDelay {
                    tRequired = Nwk_ObjRequired(pFanout) - tDelay; }
            }
        }
    }
    return tRequired;
}

/**Function*************************************************************

  Synopsis    [Propagates the required times through the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodePropagateRequired(pObj *Nwk_Obj_t, fUseSorting int32) float32 {
    var pLutLib *If_LibLut_t  = pObj.pMan.pLutLib;
    var pPinPerm [32]int32;
    var pPinDelays [32]float32;
    var pFanin *Nwk_Obj_t;
    var tRequired float32  = 0.0; // Suppress "might be used uninitialized"
    var pDelays *float32;
    var k int32;
    assert(Nwk_ObjIsNode(pObj), "Nwk_ObjIsNode(pObj)");
    if pLutLib == nil {
        tRequired = Nwk_ObjRequired(pObj) - float32(1.0);
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  {
            if Nwk_ObjRequired(pFanin) > tRequired {
                Nwk_ObjSetRequired( pFanin, tRequired ); }
        }
    
    } else if !pLutLib.fVarPinDelays {
        pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pObj)];
        tRequired = Nwk_ObjRequired(pObj) - pDelays[0];
        // inlining of macro Nwk_ObjForEachFanin
        for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  {
            if Nwk_ObjRequired(pFanin) > tRequired {
                Nwk_ObjSetRequired( pFanin, tRequired ); }
        }
    
    } else {
        pDelays = pLutLib.pLutDelays[Nwk_ObjFaninNum(pObj)];
        if fUseSorting != 0 {
            Nwk_ManDelayTraceSortPins( pObj, pPinPerm, pPinDelays );
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  {
                tRequired = Nwk_ObjRequired(pObj) - pDelays[k];
                if Nwk_ObjRequired(Nwk_ObjFanin(pObj,pPinPerm[k])) > tRequired {
                    Nwk_ObjSetRequired( Nwk_ObjFanin(pObj,pPinPerm[k]), tRequired ); }
            }
        
        } else {
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[k]); k++  {
                tRequired = Nwk_ObjRequired(pObj) - pDelays[k];
                if Nwk_ObjRequired(pFanin) > tRequired {
                    Nwk_ObjSetRequired( pFanin, tRequired ); }
            }
        }
    }
    return tRequired;
}

/**Function*************************************************************

  Synopsis    [Computes the delay trace of the given network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDelayTraceLut(pNtk *Nwk_Man_t) float32 {
    var vObjs *Vec_Ptr_t;
    var fUseSorting int32  = 1;
    var pLutLib *If_LibLut_t  = pNtk.pLutLib;
    var vNodes *Vec_Ptr_t;
    var pObj *Nwk_Obj_t;
    var tArrival, tRequired, tSlack float32;
    var i int32;

    // get the library
    if (pLutLib != nil) && pLutLib.LutMax < Nwk_ManGetFaninMax(pNtk) {
        fmt.Printf("The max LUT size (%d) is less than the max fanin count (%d).\n", pLutLib.LutMax, Nwk_ManGetFaninMax(pNtk));
        return -TIM_ETERNITY;
    }

    // compute the reverse order of all objects
    vNodes = Nwk_ManDfsReverse( pNtk );

    // initialize the arrival times
    Nwk_ManCleanTiming( pNtk );

    // propagate arrival times
    if pNtk.pManTime {
        Tim_ManIncrementTravId( pNtk.pManTime ); }
//    Nwk_ManForEachObj( pNtk, pObj, i )
    vObjs = Nwk_ManDfs( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        tArrival = Nwk_NodeComputeArrival( pObj, fUseSorting );
        if Nwk_ObjIsCi(pObj) && pNtk.pManTime {
            tArrival = Tim_ManGetCiArrival( pNtk.pManTime, pObj.PioId ); }
        if Nwk_ObjIsCo(pObj) && pNtk.pManTime {
            Tim_ManSetCoArrival( pNtk.pManTime, pObj.PioId, tArrival ); }
        Nwk_ObjSetArrival( pObj, tArrival );
    }
    Vec_PtrFree( vObjs );

    // get the latest arrival times
    tArrival = -TIM_ETERNITY;
    // inlining of macro Nwk_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.vCos)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vCos, i)), 1); i++  {                                        if !Nwk_ObjIsPo(pObj) {} else
        if tArrival < Nwk_ObjArrival(pObj) {
            tArrival = Nwk_ObjArrival(pObj); }
    }

    // initialize the required times
    if pNtk.pManTime {
        Tim_ManIncrementTravId( pNtk.pManTime );
        Tim_ManInitPoRequiredAll( pNtk.pManTime, tArrival );
    
    } else {
        // inlining of macro Nwk_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.vCos)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vCos, i)), 1); i++  {
            Nwk_ObjSetRequired( pObj, tArrival );
        }
    }

    // propagate the required times
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Nwk_ObjIsNode(pObj) {
            Nwk_NodePropagateRequired( pObj, fUseSorting );
        
        } else if Nwk_ObjIsCi(pObj) {
            if pNtk.pManTime {
                Tim_ManSetCiRequired( pNtk.pManTime, pObj.PioId, Nwk_ObjRequired(pObj) ); }
        
        } else if Nwk_ObjIsCo(pObj) {
            if pNtk.pManTime {
                tRequired = Tim_ManGetCoRequired( pNtk.pManTime, pObj.PioId );
                Nwk_ObjSetRequired( pObj, tRequired );
            }
            if Nwk_ObjRequired(Nwk_ObjFanin0(pObj)) > Nwk_ObjRequired(pObj) {
                Nwk_ObjSetRequired( Nwk_ObjFanin0(pObj), Nwk_ObjRequired(pObj) ); }
        }

        // set slack for this object
        tSlack = Nwk_ObjRequired(pObj) - Nwk_ObjArrival(pObj);
        assert(tSlack + 0.01 > 0.0, "tSlack + 0.01 > 0.0");
        Nwk_ObjSetSlack( pObj, tSlack < 0.0 ? 0.0 : tSlack );
    }
    Vec_PtrFree( vNodes );
    return tArrival;
}

/**Function*************************************************************

  Synopsis    [Computes the arrival times for the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManVerifyTiming(pNtk *Nwk_Man_t) int32 {
    var pObj *Nwk_Obj_t;
    var tArrival, tRequired float32;
    var i int32;
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        if Nwk_ObjIsCi(pObj) && Nwk_ObjFanoutNum(pObj) == 0 {
            continue; }
        tArrival = Nwk_NodeComputeArrival( pObj, 1 );
        tRequired = Nwk_NodeComputeRequired( pObj, 1 );
        if !Nwk_ManTimeEqual( tArrival, Nwk_ObjArrival(pObj), float32(0.01) ) {
            fmt.Printf("Nwk_ManVerifyTiming(): Object %d has different arrival time (%.2f) from computed (%.2f).\n", pObj.Id, Nwk_ObjArrival(pObj), tArrival); }
        if !Nwk_ManTimeEqual( tRequired, Nwk_ObjRequired(pObj), float32(0.01) ) {
            fmt.Printf("Nwk_ManVerifyTiming(): Object %d has different required time (%.2f) from computed (%.2f).\n", pObj.Id, Nwk_ObjRequired(pObj), tRequired); }
    }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prints the delay trace for the given network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDelayTracePrint(pNtk *Nwk_Man_t) {
    var pLutLib *If_LibLut_t  = pNtk.pLutLib;
    var pNode *Nwk_Obj_t;
    var i int32;
    var Nodes int32;
    var pCounters *int32;
    var tArrival, tDelta, nSteps, Num float32;
    // get the library
    if (pLutLib != nil) && pLutLib.LutMax < Nwk_ManGetFaninMax(pNtk) {
        fmt.Printf("The max LUT size (%d) is less than the max fanin count (%d).\n", pLutLib.LutMax, Nwk_ManGetFaninMax(pNtk));
        return;
    }
    // decide how many steps
    if pLutLib != nil {
        nSteps = 20;
    } else {
        nSteps = Nwk_ManLevelMax(pNtk);
    }
    // inlining of macro ABC_ALLOC
    pCounters = ((int *) malloc(sizeof(int) * (size_t)(nSteps + 1)));
    memset( pCounters, 0, sizeof(int)*(nSteps + 1) );
    // perform delay trace
    tArrival = Nwk_ManDelayTraceLut( pNtk );
    tDelta = tArrival / nSteps;
    // count how many nodes have slack in the corresponding intervals
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pNode) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if (pNode) == nil || !Nwk_ObjIsNode(pNode) {} else {
        if Nwk_ObjFaninNum(pNode) == 0 {
            continue; }
        Num = Nwk_ObjSlack(pNode) / tDelta;
        if Num > nSteps {
            continue; }
        assert(Num >=0 && Num <= nSteps, "Num >=0 && Num <= nSteps");
        pCounters[int32(Num)]++;
    }
    }
    // print the results
    fmt.Printf("Max delay = %6.2f. Delay trace using %s model:\n", tArrival, pLutLib? "LUT library" : "unit-delay");
    Nodes = 0;
    for  i = 0; i < nSteps; i++  {
        Nodes += pCounters[i];
        fmt.Printf("%3d %s : %5d  (%6.2f %%)\n", pLutLib? 5*(i+1) : i+1, pLutLib? "%":"lev", Nodes, 100.0*Nodes/Nwk_ManNodeNum(pNtk));
    }
    pCounters = nil;
}


/**Function*************************************************************

  Synopsis    [Inserts node into the queue of nodes sorted by level.]

  Description [The inserted node should not go before the current position 
  given by iCurrent. If the arrival times are computed, the nodes are sorted
  in the increasing order of levels. If the required times are computed, 
  the nodes are sorted in the decreasing order of levels.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodeUpdateAddToQueue(vQueue *Vec_Ptr_t, pObj *Nwk_Obj_t, iCurrent int32, fArrival int32) {
    var pTemp1, pTemp2 *Nwk_Obj_t;
    var i int32;
    Vec_PtrPush( vQueue, pObj );
    for  i = Vec_PtrSize(vQueue) - 1; i > iCurrent + 1; i--  {
        pTemp1 = (Nwk_Obj_t *)vQueue.pArray[i];
        pTemp2 = (Nwk_Obj_t *)vQueue.pArray[i-1];
        if fArrival != 0 {
            if Nwk_ObjLevel(pTemp2) <= Nwk_ObjLevel(pTemp1) {
                break; }
        
        } else {
            if Nwk_ObjLevel(pTemp2) >= Nwk_ObjLevel(pTemp1) {
                break; }
        }
        vQueue.pArray[i-1] = pTemp1;
        vQueue.pArray[i]   = pTemp2;
    }
    // verification
    for  i = iCurrent + 1; i < Vec_PtrSize(vQueue) - 1; i++  {
        pTemp1 = (Nwk_Obj_t *)vQueue.pArray[i];
        pTemp2 = (Nwk_Obj_t *)vQueue.pArray[i+1];
        if fArrival != 0 {
            assert(Nwk_ObjLevel(pTemp1) <= Nwk_ObjLevel(pTemp2), "Nwk_ObjLevel(pTemp1) <= Nwk_ObjLevel(pTemp2)");
        } else {
            assert(Nwk_ObjLevel(pTemp1) >= Nwk_ObjLevel(pTemp2), "Nwk_ObjLevel(pTemp1) >= Nwk_ObjLevel(pTemp2)"); }
    }
}

/**Function*************************************************************

  Synopsis    [Incrementally updates arrival times of the node.]

  Description [Supports variable-pin delay model and white-boxes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodeUpdateArrival(pObj *Nwk_Obj_t) {
    var pManTime *Tim_Man_t  = pObj.pMan.pManTime;
    var vQueue *Vec_Ptr_t  = pObj.pMan.vTemp;
    var pTemp *Nwk_Obj_t;
    var pNext *Nwk_Obj_t  = nil; // Suppress "might be used uninitialized"
    var tArrival float32;
    var iCur, k, iBox, iTerm1, nTerms int32;
    assert(Nwk_ObjIsNode(pObj), "Nwk_ObjIsNode(pObj)");
    // verify the arrival time
    tArrival = Nwk_NodeComputeArrival( pObj, 1 );
    assert(Nwk_ManTimeLess( tArrival, Nwk_ObjRequired(pObj), float32(0.01) ), "Nwk_ManTimeLess( tArrival, Nwk_ObjRequired(pObj), float32(0.01) )");
    // initialize the queue with the node
    Vec_PtrClear( vQueue );
    Vec_PtrPush( vQueue, pObj );
    pObj.MarkA = 1;
    // process objects
    if pManTime != nil {
        Tim_ManIncrementTravId( pManTime ); }
    // inlining of macro Vec_PtrForEachEntry
    for  iCur = 0; (iCur < Vec_PtrSize(vQueue)) && (((pTemp) = (Nwk_Obj_t *)Vec_PtrEntry(vQueue, iCur)), 1); iCur++  {
        pTemp.MarkA = 0;
        tArrival = Nwk_NodeComputeArrival( pTemp, 1 );
        if Nwk_ObjIsCi(pTemp) && (pManTime != nil) {
            tArrival = Tim_ManGetCiArrival( pManTime, pTemp.PioId ); }
        if Nwk_ManTimeEqual( tArrival, Nwk_ObjArrival(pTemp), float32(0.01) ) {
            continue; }
        Nwk_ObjSetArrival( pTemp, tArrival );
        // add the fanouts to the queue
        if Nwk_ObjIsCo(pTemp) {
            if pManTime != nil {
                iBox = Tim_ManBoxForCo( pManTime, pTemp.PioId );
                if iBox >= 0 { // this CO is an input of the box
                
                    // it may happen that a box-input (CO) was already marked as visited
                    // when some other box-input of the same box was visited - here we undo this
                    if Tim_ManIsCoTravIdCurrent( pManTime, pTemp.PioId ) {
                        Tim_ManSetPreviousTravIdBoxInputs( pManTime, iBox ); }
                    Tim_ManSetCoArrival( pManTime, pTemp.PioId, tArrival );
                    Tim_ManSetCurrentTravIdBoxInputs( pManTime, iBox );
                    iTerm1 = Tim_ManBoxOutputFirst( pManTime, iBox );
                    nTerms = Tim_ManBoxOutputNum( pManTime, iBox );
                    for  k = 0; k < nTerms; k++  {
                        pNext = Nwk_ManCi(pNext.pMan, iTerm1 + k);
                        if pNext.MarkA {
                            continue; }
                        Nwk_NodeUpdateAddToQueue( vQueue, pNext, iCur, 1 );
                        pNext.MarkA = 1;
                    }
                }
            }
        
        } else {
            // inlining of macro Nwk_ObjForEachFanout
            for  k = 0; (k < int32((pTemp).nFanouts)) && ((pNext) = (pTemp).pFanio[(pTemp).nFanins+k]); k++  {
                if pNext.MarkA {
                    continue; }
                Nwk_NodeUpdateAddToQueue( vQueue, pNext, iCur, 1 );
                pNext.MarkA = 1;
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Incrementally updates required times of the node.]

  Description [Supports variable-pin delay model and white-boxes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodeUpdateRequired(pObj *Nwk_Obj_t) {
    var pManTime *Tim_Man_t  = pObj.pMan.pManTime;
    var vQueue *Vec_Ptr_t  = pObj.pMan.vTemp;
    var pTemp *Nwk_Obj_t;
    var pNext *Nwk_Obj_t  = nil; // Suppress "might be used uninitialized"
    var tRequired float32;
    var iCur, k, iBox, iTerm1, nTerms int32;
    assert(Nwk_ObjIsNode(pObj), "Nwk_ObjIsNode(pObj)");
    // make sure the node's required time remained the same
    tRequired = Nwk_NodeComputeRequired( pObj, 1 );
    assert(Nwk_ManTimeEqual( tRequired, Nwk_ObjRequired(pObj), float32(0.01) ), "Nwk_ManTimeEqual( tRequired, Nwk_ObjRequired(pObj), float32(0.01) )");
    // initialize the queue with the node's faninsa and the old node's fanins
    Vec_PtrClear( vQueue );
    // inlining of macro Nwk_ObjForEachFanin
    for  k = 0; (k < int32((pObj).nFanins)) && ((pNext) = (pObj).pFanio[k]); k++  {
        if pNext.MarkA {
            continue; }
        Nwk_NodeUpdateAddToQueue( vQueue, pNext, -1, 0 );
        pNext.MarkA = 1;
    }
    // process objects
    if pManTime != nil {
        Tim_ManIncrementTravId( pManTime ); }
    // inlining of macro Vec_PtrForEachEntry
    for  iCur = 0; (iCur < Vec_PtrSize(vQueue)) && (((pTemp) = (Nwk_Obj_t *)Vec_PtrEntry(vQueue, iCur)), 1); iCur++  {
        pTemp.MarkA = 0;
        tRequired = Nwk_NodeComputeRequired( pTemp, 1 );
        if Nwk_ObjIsCo(pTemp) && (pManTime != nil) {
            tRequired = Tim_ManGetCoRequired( pManTime, pTemp.PioId ); }
        if Nwk_ManTimeEqual( tRequired, Nwk_ObjRequired(pTemp), float32(0.01) ) {
            continue; }
        Nwk_ObjSetRequired( pTemp, tRequired );
        // add the fanins to the queue
        if Nwk_ObjIsCi(pTemp) {
            if pManTime != nil {
                iBox = Tim_ManBoxForCi( pManTime, pTemp.PioId );
                if iBox >= 0 { // this CI is an output of the box
                
                    // it may happen that a box-output (CI) was already marked as visited
                    // when some other box-output of the same box was visited - here we undo this
                    if Tim_ManIsCiTravIdCurrent( pManTime, pTemp.PioId ) {
                        Tim_ManSetPreviousTravIdBoxOutputs( pManTime, iBox ); }
                    Tim_ManSetCiRequired( pManTime, pTemp.PioId, tRequired );
                    Tim_ManSetCurrentTravIdBoxOutputs( pManTime, iBox );
                    iTerm1 = Tim_ManBoxInputFirst( pManTime, iBox );
                    nTerms = Tim_ManBoxInputNum( pManTime, iBox );
                    for  k = 0; k < nTerms; k++  {
                        pNext = Nwk_ManCo(pNext.pMan, iTerm1 + k);
                        if pNext.MarkA {
                            continue; }
                        Nwk_NodeUpdateAddToQueue( vQueue, pNext, iCur, 0 );
                        pNext.MarkA = 1;
                    }
                }
            }
        
        } else {
            // inlining of macro Nwk_ObjForEachFanin
            for  k = 0; (k < int32((pTemp).nFanins)) && ((pNext) = (pTemp).pFanio[k]); k++  {
                if pNext.MarkA {
                    continue; }
                Nwk_NodeUpdateAddToQueue( vQueue, pNext, iCur, 0 );
                pNext.MarkA = 1;
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Computes the level of the node using its fanin levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjLevelNew(pObj *Nwk_Obj_t) int32 {
    var pManTime *Tim_Man_t  = pObj.pMan.pManTime;
    var pFanin *Nwk_Obj_t;
    var i, iBox, iTerm1, nTerms, Level int32;
    if Nwk_ObjIsCi(pObj) || Nwk_ObjIsLatch(pObj) {
        if pManTime != nil {
            iBox = Tim_ManBoxForCi( pManTime, pObj.PioId );
            if iBox >= 0 { // this CI is an output of the box
            
                iTerm1 = Tim_ManBoxInputFirst( pManTime, iBox );
                nTerms = Tim_ManBoxInputNum( pManTime, iBox );
                for  i = 0; i < nTerms; i++  {
                    pFanin = Nwk_ManCo(pObj.pMan, iTerm1 + i);
                    Level = Abc_MaxInt( Level, Nwk_ObjLevel(pFanin) );
                }
                Level++;
            }
        }
        return Level;
    }
    assert(Nwk_ObjIsNode(pObj) || Nwk_ObjIsCo(pObj), "Nwk_ObjIsNode(pObj) || Nwk_ObjIsCo(pObj)");
    // inlining of macro Nwk_ObjForEachFanin
    for  i = 0; (i < int32((pObj).nFanins)) && ((pFanin) = (pObj).pFanio[i]); i++  {
        Level = Abc_MaxInt( Level, Nwk_ObjLevel(pFanin) );
    }
    return Level + (Nwk_ObjIsNode(pObj) && Nwk_ObjFaninNum(pObj) > 0);
}

/**Function*************************************************************

  Synopsis    [Incrementally updates level of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManUpdateLevel(pObj *Nwk_Obj_t) {
    var pManTime *Tim_Man_t  = pObj.pMan.pManTime;
    var vQueue *Vec_Ptr_t  = pObj.pMan.vTemp;
    var pTemp *Nwk_Obj_t;
    var pNext *Nwk_Obj_t  = nil; // Suppress "might be used uninitialized"
    var LevelNew, iCur, k, iBox, iTerm1, nTerms int32;
    assert(Nwk_ObjIsNode(pObj), "Nwk_ObjIsNode(pObj)");
    // initialize the queue with the node
    Vec_PtrClear( vQueue );
    Vec_PtrPush( vQueue, pObj );
    pObj.MarkA = 1;
    // process objects
    // inlining of macro Vec_PtrForEachEntry
    for  iCur = 0; (iCur < Vec_PtrSize(vQueue)) && (((pTemp) = (Nwk_Obj_t *)Vec_PtrEntry(vQueue, iCur)), 1); iCur++  {
        pTemp.MarkA = 0;
        LevelNew = Nwk_ObjLevelNew( pTemp );
        if LevelNew == Nwk_ObjLevel(pTemp) {
            continue; }
        Nwk_ObjSetLevel( pTemp, LevelNew );
        // add the fanouts to the queue
        if Nwk_ObjIsCo(pTemp) {
            if pManTime != nil {
                iBox = Tim_ManBoxForCo( pManTime, pTemp.PioId );
                if iBox >= 0 { // this is not a true PO
                
                    Tim_ManSetCurrentTravIdBoxInputs( pManTime, iBox );
                    iTerm1 = Tim_ManBoxOutputFirst( pManTime, iBox );
                    nTerms = Tim_ManBoxOutputNum( pManTime, iBox );
                    for  k = 0; k < nTerms; k++  {
                        pNext = Nwk_ManCi(pNext.pMan, iTerm1 + k);
                        if pNext.MarkA {
                            continue; }
                        Nwk_NodeUpdateAddToQueue( vQueue, pNext, iCur, 1 );
                        pNext.MarkA = 1;
                    }
                }
            }
        
        } else {
            // inlining of macro Nwk_ObjForEachFanout
            for  k = 0; (k < int32((pTemp).nFanouts)) && ((pNext) = (pTemp).pFanio[(pTemp).nFanins+k]); k++  {
                if pNext.MarkA {
                    continue; }
                Nwk_NodeUpdateAddToQueue( vQueue, pNext, iCur, 1 );
                pNext.MarkA = 1;
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Computes the level of the node using its fanin levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManVerifyLevel(pNtk *Nwk_Man_t) int32 {
    var pObj *Nwk_Obj_t;
    var LevelNew, i int32;
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObj == nil {} else {
        assert(pObj.MarkA == 0, "pObj.MarkA == 0");
        LevelNew = Nwk_ObjLevelNew( pObj );
        if Nwk_ObjLevel(pObj) != LevelNew {
            fmt.Printf("Object %6d: Mismatch betweeh levels: Actual = %d. Correct = %d.\n", i, Nwk_ObjLevel(pObj), LevelNew);
        }
    }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Replaces the node and incrementally updates levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManUpdate(pObj *Nwk_Obj_t, pObjNew *Nwk_Obj_t, vLevels *Vec_Vec_t) {
    assert(pObj.pMan == pObjNew.pMan, "pObj.pMan == pObjNew.pMan");
    assert(pObj != pObjNew, "pObj != pObjNew");
    assert(Nwk_ObjFanoutNum(pObj) > 0, "Nwk_ObjFanoutNum(pObj) > 0");
    assert(Nwk_ObjIsNode(pObj) && !Nwk_ObjIsCo(pObjNew), "Nwk_ObjIsNode(pObj) && !Nwk_ObjIsCo(pObjNew)");
    // transfer fanouts to the old node
    Nwk_ObjTransferFanout( pObj, pObjNew );
    // transfer the timing information
    // (this is needed because updating level happens if the level has changed;
    // when we set the old level, it will be recomputed by the level updating
    // procedure, which will update level of other nodes if there is a difference)
    pObjNew.Level = pObj.Level;
    pObjNew.tArrival = pObj.tArrival;
    pObjNew.tRequired = pObj.tRequired;
    // update required times of the old fanins
    pObj.tRequired = TIM_ETERNITY;
    Nwk_NodeUpdateRequired( pObj );
    // remove the old node
    Nwk_ManDeleteNode_rec( pObj );
    // update the information of the new node
    Nwk_ManUpdateLevel( pObjNew );
    Nwk_NodeUpdateArrival( pObjNew );
    Nwk_NodeUpdateRequired( pObjNew );
//Nwk_ManVerifyTiming( pObjNew->pMan );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

