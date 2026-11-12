package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcSpeedup.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Delay trace and speedup.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcSpeedup.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "map/if/if.h"
//#include "aig/aig/aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Abc_ObjArrival(pNode *Abc_Obj_t) float32 { return pNode.pNtk.pLutTimes[3*pNode.Id+0]; }
func Abc_ObjRequired(pNode *Abc_Obj_t) float32 { return pNode.pNtk.pLutTimes[3*pNode.Id+1]; }
func Abc_ObjSlack(pNode *Abc_Obj_t) float32 { return pNode.pNtk.pLutTimes[3*pNode.Id+2]; }

func Abc_ObjSetArrival(pNode *Abc_Obj_t, Time float32) { pNode.pNtk.pLutTimes[3*pNode.Id+0] = Time; }
func Abc_ObjSetRequired(pNode *Abc_Obj_t, Time float32) { pNode.pNtk.pLutTimes[3*pNode.Id+1] = Time; }
func Abc_ObjSetSlack(pNode *Abc_Obj_t, Time float32) { pNode.pNtk.pLutTimes[3*pNode.Id+2] = Time; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sorts the pins in the decreasing order of delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTraceSortPins(pNode *Abc_Obj_t, pPinPerm *int32, pPinDelays *float32) {
    var pFanin *Abc_Obj_t;
    var i, j, best_i, temp int32;
    // start the trivial permutation and collect pin delays
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        pPinPerm[i] = i;
        pPinDelays[i] = Abc_ObjArrival(pFanin);
    }
    // selection sort the pins in the decreasible order of delays
    // this order will match the increasing order of LUT input pins
    for  i = 0; i < Abc_ObjFaninNum(pNode)-1; i++  {
        best_i = i;
        for  j = i+1; j < Abc_ObjFaninNum(pNode); j++  {
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
    assert(Abc_ObjFaninNum(pNode) == 0 || pPinPerm[0] < Abc_ObjFaninNum(pNode), "Abc_ObjFaninNum(pNode) == 0 || pPinPerm[0] < Abc_ObjFaninNum(pNode)");
    for  i = 1; i < Abc_ObjFaninNum(pNode); i++  {
        assert(pPinPerm[i] < Abc_ObjFaninNum(pNode), "pPinPerm[i] < Abc_ObjFaninNum(pNode)");
        assert(pPinDelays[pPinPerm[i-1]] >= pPinDelays[pPinPerm[i]], "pPinDelays[pPinPerm[i-1]] >= pPinDelays[pPinPerm[i]]");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTraceLut(pNtk *Abc_Ntk_t, fUseLutLib int32) float32 {
    var fUseSorting int32  = 1;
    var pPinPerm [32]int32;
    var pPinDelays [32]float32;
    var pLutLib *If_LibLut_t;
    var pNode, pFanin *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var tArrival float32;
    var tRequired float32;
    var tSlack float32;
    var pDelays *float32;
    var i, k int32;

    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    // get the library
    if fUseLutLib != 0 {
        pLutLib = (If_LibLut_t *)Abc_FrameReadLibLut();
    } else {
        pLutLib = nil;
    }
    if (pLutLib != nil) && pLutLib.LutMax < Abc_NtkGetFaninMax(pNtk) {
        fmt.Printf("The max LUT size (%d) is less than the max fanin count (%d).\n", pLutLib.LutMax, Abc_NtkGetFaninMax(pNtk));
        return -ABC_INFINITY;
    }

    // initialize the arrival times
    pNtk.pLutTimes = nil;
    // inlining of macro ABC_ALLOC
    pNtk.pLutTimes = ((float *) malloc(sizeof(float) * (size_t)(3 * Abc_NtkObjNumMax(pNtk))));
    for  i = 0; i < Abc_NtkObjNumMax(pNtk); i++  {
        pNtk.pLutTimes[3*i+0] = pNtk.pLutTimes[3*i+2] = 0;
        pNtk.pLutTimes[3*i+1] = ABC_INFINITY;
    }

    // propagate arrival times
    vNodes = Abc_NtkDfs( pNtk, 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        tArrival = -ABC_INFINITY;
        if pLutLib == nil {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if tArrival < Abc_ObjArrival(pFanin) + 1.0 {
                    tArrival = Abc_ObjArrival(pFanin) + 1.0; }
            }
        
        } else if !pLutLib.fVarPinDelays {
            pDelays = pLutLib.pLutDelays[Abc_ObjFaninNum(pNode)];
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if tArrival < Abc_ObjArrival(pFanin) + pDelays[0] {
                    tArrival = Abc_ObjArrival(pFanin) + pDelays[0]; }
            }
        
        } else {
            pDelays = pLutLib.pLutDelays[Abc_ObjFaninNum(pNode)];
            if fUseSorting != 0 {
                Abc_NtkDelayTraceSortPins( pNode, pPinPerm, pPinDelays );
                // inlining of macro Abc_ObjForEachFanin
                for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  { 
                    if tArrival < Abc_ObjArrival(Abc_ObjFanin(pNode,pPinPerm[k])) + pDelays[k] {
                        tArrival = Abc_ObjArrival(Abc_ObjFanin(pNode,pPinPerm[k])) + pDelays[k]; }
                }
            
            } else {
                // inlining of macro Abc_ObjForEachFanin
                for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                    if tArrival < Abc_ObjArrival(pFanin) + pDelays[k] {
                        tArrival = Abc_ObjArrival(pFanin) + pDelays[k]; }
                }
            }
        }
        if Abc_ObjFaninNum(pNode) == 0 {
            tArrival = 0.0; }
        Abc_ObjSetArrival( pNode, tArrival );
    }
    Vec_PtrFree( vNodes );

    // get the latest arrival times
    tArrival = -ABC_INFINITY;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if tArrival < Abc_ObjArrival(Abc_ObjFanin0(pNode)) {
            tArrival = Abc_ObjArrival(Abc_ObjFanin0(pNode)); }
    }

    // initialize the required times
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if Abc_ObjRequired(Abc_ObjFanin0(pNode)) > tArrival {
            Abc_ObjSetRequired( Abc_ObjFanin0(pNode), tArrival ); }
    }

    // propagate the required times
    vNodes = Abc_NtkDfsReverse( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if pLutLib == nil {
            tRequired = Abc_ObjRequired(pNode) - float32(1.0);
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if Abc_ObjRequired(pFanin) > tRequired {
                    Abc_ObjSetRequired( pFanin, tRequired ); }
            }
        
        } else if !pLutLib.fVarPinDelays {
            pDelays = pLutLib.pLutDelays[Abc_ObjFaninNum(pNode)];
            tRequired = Abc_ObjRequired(pNode) - pDelays[0];
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if Abc_ObjRequired(pFanin) > tRequired {
                    Abc_ObjSetRequired( pFanin, tRequired ); }
            }
        
        } else {
            pDelays = pLutLib.pLutDelays[Abc_ObjFaninNum(pNode)];
            if fUseSorting != 0 {
                Abc_NtkDelayTraceSortPins( pNode, pPinPerm, pPinDelays );
                // inlining of macro Abc_ObjForEachFanin
                for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                    tRequired = Abc_ObjRequired(pNode) - pDelays[k];
                    if Abc_ObjRequired(Abc_ObjFanin(pNode,pPinPerm[k])) > tRequired {
                        Abc_ObjSetRequired( Abc_ObjFanin(pNode,pPinPerm[k]), tRequired ); }
                }
            
            } else {
                // inlining of macro Abc_ObjForEachFanin
                for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                    tRequired = Abc_ObjRequired(pNode) - pDelays[k];
                    if Abc_ObjRequired(pFanin) > tRequired {
                        Abc_ObjSetRequired( pFanin, tRequired ); }
                }
            }
        }
        // set slack for this object
        tSlack = Abc_ObjRequired(pNode) - Abc_ObjArrival(pNode);
        assert(tSlack + 0.001 > 0.0, "tSlack + 0.001 > 0.0");
        Abc_ObjSetSlack( pNode, tSlack < 0.0 ? 0.0 : tSlack );
    }
    Vec_PtrFree( vNodes );
    return tArrival;
}

/**Function*************************************************************

  Synopsis    [Delay tracing of the LUT mapped network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTracePrint(pNtk *Abc_Ntk_t, fUseLutLib int32, fVerbose int32) {
    var pNode *Abc_Obj_t;
    var pLutLib *If_LibLut_t;
    var i int32;
    var Nodes int32;
    var pCounters *int32;
    var tArrival, tDelta, nSteps, Num float32;
    // get the library
    if fUseLutLib != 0 {
        pLutLib = (If_LibLut_t *)Abc_FrameReadLibLut();
    } else {
        pLutLib = nil;
    }
    if (pLutLib != nil) && pLutLib.LutMax < Abc_NtkGetFaninMax(pNtk) {
        fmt.Printf("The max LUT size (%d) is less than the max fanin count (%d).\n", pLutLib.LutMax, Abc_NtkGetFaninMax(pNtk));
        return;
    }
    // decide how many steps
    if fUseLutLib != 0 {
        nSteps = 20;
    } else {
        nSteps = Abc_NtkLevel(pNtk);
    }
    // inlining of macro ABC_ALLOC
    pCounters = ((int *) malloc(sizeof(int) * (size_t)(nSteps + 1)));
    memset( pCounters, 0, sizeof(int)*(nSteps + 1) );
    // perform delay trace
    tArrival = Abc_NtkDelayTraceLut( pNtk, fUseLutLib );
    tDelta = tArrival / nSteps;
    // count how many nodes have slack in the corresponding intervals
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if Abc_ObjFaninNum(pNode) == 0 {
            continue; }
        Num = Abc_ObjSlack(pNode) / tDelta;
        assert(Num >=0 && Num <= nSteps, "Num >=0 && Num <= nSteps");
        pCounters[int32(Num)]++;
    }
    }
    // print the results
    fmt.Printf("Max delay = %6.2f. Delay trace using %s model:\n", tArrival, fUseLutLib? "LUT library" : "unit-delay");
    Nodes = 0;
    for  i = 0; i < nSteps; i++  {
        Nodes += pCounters[i];
        fmt.Printf("%3d %s : %5d  (%6.2f %%)\n", fUseLutLib? 5*(i+1) : i+1, fUseLutLib? "%":"lev", Nodes, 100.0*Nodes/Abc_NtkNodeNum(pNtk));
    }
    pCounters = nil;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigCheckTfi_rec(pNode *Abc_Obj_t, pOld *Abc_Obj_t) int32 {
    // check the trivial cases
    if pNode == nil {
        return 0; }
    if Abc_ObjIsCi(pNode) {
        return 0; }
    if pNode == pOld {
        return 1; }
    // skip the visited node
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return 0; }
    Abc_NodeSetTravIdCurrent( pNode );
    // check the children
    if Abc_AigCheckTfi_rec( Abc_ObjFanin0(pNode), pOld ) != 0 {
        return 1; }
    if Abc_AigCheckTfi_rec( Abc_ObjFanin1(pNode), pOld ) != 0 {
        return 1; }
    // check equivalent nodes
    return Abc_AigCheckTfi_rec( (Abc_Obj_t *)pNode.pData, pOld );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigCheckTfi(pNew *Abc_Obj_t, pOld *Abc_Obj_t) int32 {
    assert(!Abc_ObjIsComplement(pNew), "!Abc_ObjIsComplement(pNew)");
    assert(!Abc_ObjIsComplement(pOld), "!Abc_ObjIsComplement(pOld)");
    Abc_NtkIncrementTravId( pNew.pNtk );
    return Abc_AigCheckTfi_rec( pNew, pOld );
}

/**Function*************************************************************

  Synopsis    [Adds strashed nodes for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSpeedupNode_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) int32 {
    if Abc_NodeIsTravIdCurrent(pNode) {
        return 1; }
    if Abc_ObjIsCi(pNode) {
        return 0; }
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    Abc_NodeSetTravIdCurrent( pNode );
    if !(Abc_NtkSpeedupNode_rec( Abc_ObjFanin0(pNode), vNodes ) != 0) {
        return 0; }
    if !(Abc_NtkSpeedupNode_rec( Abc_ObjFanin1(pNode), vNodes ) != 0) {
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
func Abc_NtkSpeedupNode(pNtk *Abc_Ntk_t, pAig *Abc_Ntk_t, pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t, vTimes *Vec_Ptr_t) {
    var vNodes *Vec_Ptr_t;
    var pObj, pObj2, pAnd *Abc_Obj_t;
    var ppCofs [32]*Abc_Obj_t;
    var nCofs, i, k, nSkip int32;

    // quit of regulars are the same
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pObj2) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
        if i != k && Abc_ObjRegular(pObj.pCopy) == Abc_ObjRegular(pObj2.pCopy) {
//            printf( "Identical after structural hashing!!!\n" );
            return;
        }
    }
    }

    // collect the AIG nodes
    vNodes = Vec_PtrAlloc( 100 );
    Abc_NtkIncrementTravId( pAig );
    Abc_NodeSetTravIdCurrent( Abc_AigConst1(pAig) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pAnd = pObj.pCopy;
        Abc_NodeSetTravIdCurrent( Abc_ObjRegular(pAnd) );
    }
    // traverse from the root node
    pAnd = pNode.pCopy;
    if !(Abc_NtkSpeedupNode_rec( Abc_ObjRegular(pAnd), vNodes ) != 0) {
//        printf( "Bad node!!!\n" );
        Vec_PtrFree( vNodes );
        return;
    }

    // derive cofactors
    nCofs = (1 << Vec_PtrSize(vTimes));
    for  i = 0; i < nCofs; i++  {
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
            pAnd = pObj.pCopy;
            Abc_ObjRegular(pAnd).pCopy = Abc_ObjRegular(pAnd);
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vTimes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vTimes, k)), 1); k++  {
            pAnd = pObj.pCopy;
            Abc_ObjRegular(pAnd).pCopy = Abc_ObjNotCond( Abc_AigConst1(pAig), ((i & (1<<k)) == 0) );
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, k)), 1); k++  {
            pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pAig.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) );
        }
        // save the result
        pAnd = pNode.pCopy;
        ppCofs[i] = Abc_ObjNotCond( Abc_ObjRegular(pAnd).pCopy, Abc_ObjIsComplement(pAnd) );
    }
    Vec_PtrFree( vNodes );

//Abc_ObjAddFanin( Abc_NtkCreatePo(pAig), ppCofs[0] );
//Abc_ObjAddFanin( Abc_NtkCreatePo(pAig), ppCofs[1] );

    // collect the resulting tree
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vTimes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vTimes, k)), 1); k++  {
        for  nSkip = (1<<k), i = 0; i < nCofs; i += 2*nSkip  {
            pAnd = pObj.pCopy;
            ppCofs[i] = Abc_AigMux( (Abc_Aig_t *)pAig.pManFunc, Abc_ObjRegular(pAnd), ppCofs[i+nSkip], ppCofs[i] );
        }
    }
//Abc_ObjAddFanin( Abc_NtkCreatePo(pAig), ppCofs[0] );

    // create choice node
    pAnd = Abc_ObjRegular(pNode.pCopy); // repr
    pObj = Abc_ObjRegular(ppCofs[0]);    // new
    if pAnd.pData == nil && pObj.pData == nil && !Abc_AigNodeIsConst(pObj) && !(Abc_AigCheckTfi(pObj, pAnd) != 0) {
        pObj.pData = pAnd.pData;
        pAnd.pData = pObj;
    }

}

/**Function*************************************************************

  Synopsis    [Determines timing-critical edges of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTraceTCEdges(pNtk *Abc_Ntk_t, pNode *Abc_Obj_t, tDelta float32, fUseLutLib int32) uint32 {
    var pPinPerm [32]int32;
    var pPinDelays [32]float32;
    var pLutLib *If_LibLut_t;
    var pFanin *Abc_Obj_t;
    var uResult uint32;
    var tRequired float32;
    var pDelays *float32;
    var k int32;
    if fUseLutLib != 0 {
        pLutLib = (If_LibLut_t *)Abc_FrameReadLibLut();
    } else {
        pLutLib = nil;
    }
    tRequired = Abc_ObjRequired(pNode);
    if pLutLib == nil {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if tRequired < Abc_ObjArrival(pFanin) + 1.0 + tDelta {
                uResult |= (1 << k); }
        }
    
    } else if !pLutLib.fVarPinDelays {
        pDelays = pLutLib.pLutDelays[Abc_ObjFaninNum(pNode)];
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if tRequired < Abc_ObjArrival(pFanin) + pDelays[0] + tDelta {
                uResult |= (1 << k); }
        }
    
    } else {
        pDelays = pLutLib.pLutDelays[Abc_ObjFaninNum(pNode)];
        Abc_NtkDelayTraceSortPins( pNode, pPinPerm, pPinDelays );
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if tRequired < Abc_ObjArrival(Abc_ObjFanin(pNode,pPinPerm[k])) + pDelays[k] + tDelta {
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
func Abc_NtkSpeedup(pNtk *Abc_Ntk_t, fUseLutLib int32, Percentage int32, Degree int32, fVerbose int32, fVeryVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var vTimeCries, vTimeFanins *Vec_Ptr_t;
    var pNode, pFanin, pFanin2 *Abc_Obj_t;
    var tDelta, tArrival float32;
    var i, k, k2, Counter, CounterRes, nTimeCris int32;
    var puTCEdges *uint32;
    // perform delay trace
    tArrival = Abc_NtkDelayTraceLut( pNtk, fUseLutLib );
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
    puTCEdges = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Abc_NtkObjNumMax(pNtk))));
    memset( puTCEdges, 0, sizeof(unsigned) * Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if Abc_ObjSlack(pNode) >= tDelta {
            continue; }
        puTCEdges[pNode.Id] = Abc_NtkDelayTraceTCEdges( pNtk, pNode, tDelta, fUseLutLib );
    }
    }
    if fVerbose != 0 {
        Counter = CounterRes = 0;
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if !Abc_ObjIsCi(pFanin) && Abc_ObjSlack(pFanin) < tDelta {
                    Counter++; }
            }
            CounterRes += Extra_WordCountOnes( puTCEdges[pNode.Id] );
        }
        }
        fmt.Printf("Edges: Total = %7d. 0-slack = %7d. Critical = %7d. Ratio = %4.2f\n", Abc_NtkGetTotalFanins(pNtk), Counter, CounterRes, 1.0*CounterRes/Counter);
    }
    // start the resulting network
    pNtkNew = Abc_NtkStrash( pNtk, 0, 1, 0 );

    // collect nodes to be used for resynthesis
    Counter = CounterRes = 0;
    vTimeCries = Vec_PtrAlloc( 16 );
    vTimeFanins = Vec_PtrAlloc( 16 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if Abc_ObjSlack(pNode) >= tDelta {
            continue; }
        // count the number of non-PI timing-critical nodes
        nTimeCris = 0;
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if !Abc_ObjIsCi(pFanin) && (puTCEdges[pNode.Id] & (1<<k)) {
                nTimeCris++; }
        }
        if !(fVeryVerbose != 0) && nTimeCris == 0 {
            continue; }
        Counter++;
        // count the total number of timing critical second-generation nodes
        Vec_PtrClear( vTimeCries );
        if nTimeCris != 0 {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if !Abc_ObjIsCi(pFanin) && (puTCEdges[pNode.Id] & (1<<k)) {
                    // inlining of macro Abc_ObjForEachFanin
                    for  k2 = 0; (k2 < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k2)), 1); k2++  {
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
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if Abc_ObjIsCi(pFanin) {
                Vec_PtrPushUnique( vTimeFanins, pFanin );
            } else {
                // inlining of macro Abc_ObjForEachFanin
                for  k2 = 0; (k2 < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k2)), 1); k2++  {
                    Vec_PtrPushUnique( vTimeFanins, pFanin2 );
                } }                    
        }
        // print the results
        if fVeryVerbose != 0 {
        fmt.Printf("%5d Node %5d : %d %2d %2d  ", Counter, pNode.Id, nTimeCris, Vec_PtrSize(vTimeCries), Vec_PtrSize(vTimeFanins));
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            fmt.Printf("%d(%.2f)%s ", pFanin.Id, Abc_ObjSlack(pFanin), (puTCEdges[pNode.Id] & (1<<k))? "*":"");
        }
        fmt.Printf("\n");
        }
        // add the node to choices
        if Vec_PtrSize(vTimeCries) == 0 || Vec_PtrSize(vTimeCries) > Degree {
            continue; }
        // order the fanins in the increasing order of criticalily
        if Vec_PtrSize(vTimeCries) > 1 {
            pFanin = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 0 );
            pFanin2 = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
            if Abc_ObjSlack(pFanin) < Abc_ObjSlack(pFanin2) {
                Vec_PtrWriteEntry( vTimeCries, 0, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin );
            }
        }
        if Vec_PtrSize(vTimeCries) > 2 {
            pFanin = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
            pFanin2 = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 2 );
            if Abc_ObjSlack(pFanin) < Abc_ObjSlack(pFanin2) {
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 2, pFanin );
            }
            pFanin = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 0 );
            pFanin2 = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
            if Abc_ObjSlack(pFanin) < Abc_ObjSlack(pFanin2) {
                Vec_PtrWriteEntry( vTimeCries, 0, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin );
            }
        }
        // add choice
        Abc_NtkSpeedupNode( pNtk, pNtkNew, pNode, vTimeFanins, vTimeCries );
    }
    }
    Vec_PtrFree( vTimeCries );
    Vec_PtrFree( vTimeFanins );
    puTCEdges = nil;
    if fVerbose != 0 {
        fmt.Printf("Nodes: Total = %7d. 0-slack = %7d. Workable = %7d. Ratio = %4.2f\n", Abc_NtkNodeNum(pNtk), Counter, CounterRes, 1.0*CounterRes/Counter); } 

    // remove invalid choice nodes
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtkNew).vObjs)) && (((pNode) = Abc_NtkObj(pNtkNew, i)), 1); i++  {           if (pNode) == nil || !Abc_AigNodeIsAnd(pNode) {} else
        if pNode.pData {
            if Abc_ObjFanoutNum((Abc_Obj_t *)pNode.pData) > 0 {
                pNode.pData = nil; }
        }
    }

    // return the result
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Marks nodes for power-optimization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPowerEstimate(pNtk *Abc_Ntk_t, fProbOne int32) *Vec_Int_t {
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
    extern Vec_Int_t * Saig_ManComputeSwitchProbs( Aig_Man_t * p, int nFrames, int nPref, int fProbOne );
    var vProbs *Vec_Int_t;
    var vSwitching *Vec_Int_t;
    var pProbability *float32;
    var pSwitching *float32;
    var pNtkStr *Abc_Ntk_t;
    var pAig *Aig_Man_t;
    var pObjAig *Aig_Obj_t;
    var pObjAbc, pObjAbc2 *Abc_Obj_t;
    var i int32;
    // start the resulting array
    vProbs = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    pProbability = (float *)vProbs.pArray;
    // strash the network
    pNtkStr = Abc_NtkStrash( pNtk, 0, 1, 0 );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObjAbc) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObjAbc) == nil {} else
        if Abc_ObjRegular((Abc_Obj_t *)pObjAbc.pTemp).Type == ABC_FUNC_NONE {
            pObjAbc.pTemp = nil; }
    }
    // map network into an AIG
    pAig = Abc_NtkToDar( pNtkStr, 0, int32((Abc_NtkLatchNum(pNtk) > 0)) );
    vSwitching = Saig_ManComputeSwitchProbs( pAig, 48, 16, fProbOne );
    pSwitching = (float *)vSwitching.pArray;
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObjAbc) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObjAbc) == nil {} else {
        if (pObjAbc2 = Abc_ObjRegular((Abc_Obj_t *)pObjAbc.pTemp)) && (pObjAig = Aig_Regular((Aig_Obj_t *)pObjAbc2.pTemp)) {
            pProbability[pObjAbc.Id] = pSwitching[pObjAig.Id]; }
    }
    }
    Vec_IntFree( vSwitching );
    Aig_ManStop( pAig );
    Abc_NtkDelete( pNtkStr );
    return vProbs;
}

/**Function*************************************************************

  Synopsis    [Marks nodes for power-optimization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPowerPrint(pNtk *Abc_Ntk_t, vProbs *Vec_Int_t) {
    var pObj *Abc_Obj_t;
    var pProb *float32;
    var TotalProb float32 = 0.0;
    var ProbThis float32;
    var Probs [6]float32 = [6]float32{0.0};
    var i int32;
    var nNodes int32;
    var nEdges int32;
    var Counter [6]int32 ;
    pProb = (float *)vProbs.pArray;
    assert(Vec_IntSize(vProbs) >= Abc_NtkObjNumMax(pNtk), "Vec_IntSize(vProbs) >= Abc_NtkObjNumMax(pNtk)");
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        if !Abc_ObjIsNode(pObj) && !Abc_ObjIsPi(pObj) {
            continue; }
        nNodes++;
        nEdges += Abc_ObjFanoutNum(pObj);
        ProbThis = pProb[i] * Abc_ObjFanoutNum(pObj);
        TotalProb += ProbThis;
        assert(pProb[i] >= 0.0 && pProb[i] <= 1.0, "pProb[i] >= 0.0 && pProb[i] <= 1.0");
        if pProb[i] >= 0.5 {
            Counter[5]++;
            Probs[5] += ProbThis;
        
        } else if pProb[i] >= 0.4 {
            Counter[4]++;
            Probs[4] += ProbThis;
        
        } else if pProb[i] >= 0.3 {
            Counter[3]++;
            Probs[3] += ProbThis;
        
        } else if pProb[i] >= 0.2 {
            Counter[2]++;
            Probs[2] += ProbThis;
        
        } else if pProb[i] >= 0.1 {
            Counter[1]++;
            Probs[1] += ProbThis;
        
        } else {
            Counter[0]++;
            Probs[0] += ProbThis;
        }
    }
    }
    fmt.Printf("Node  distribution: ");
    for  i = 0; i < 6; i++  {
        fmt.Printf("n%d%d = %6.2f%%  ", i, i+1, 100.0 * Counter[i]/nNodes);
    }
    fmt.Printf("\n");
    fmt.Printf("Power distribution: ");
    for  i = 0; i < 6; i++  {
        fmt.Printf("p%d%d = %6.2f%%  ", i, i+1, 100.0 * Probs[i]/TotalProb);
    }
    fmt.Printf("\n");
    fmt.Printf("Total probs = %7.2f. ", TotalProb);
    fmt.Printf("Total edges = %d. ", nEdges);
    fmt.Printf("Average = %7.2f. ", TotalProb / nEdges);
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Determines timing-critical edges of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPowerCriticalEdges(pNtk *Abc_Ntk_t, pNode *Abc_Obj_t, Limit float32, vProbs *Vec_Int_t) uint32 {
    var pFanin *Abc_Obj_t;
    var pProb *float32  = (float *)vProbs.pArray;
    var uResult uint32;
    var k int32;
    // inlining of macro Abc_ObjForEachFanin
    for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
        if pProb[pFanin.Id] >= Limit {
            uResult |= (1 << k); }
    }
    return uResult;
}

/**Function*************************************************************

  Synopsis    [Adds choices to speed up the network by the given percentage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPowerdown(pNtk *Abc_Ntk_t, fUseLutLib int32, Percentage int32, Degree int32, fVerbose int32, fVeryVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var vProbs *Vec_Int_t;
    var vTimeCries, vTimeFanins *Vec_Ptr_t;
    var pNode, pFanin, pFanin2 *Abc_Obj_t;
    var pProb *float32;
    var Limit float32;
    var i, k, k2, Counter, CounterRes, nTimeCris int32;
    var puPCEdges *uint32;
    // compute the limit
    Limit = 0.5 - (1.0 * Percentage / 100);
    // perform computation of switching probability
    vProbs = Abc_NtkPowerEstimate( pNtk, 0 );
    pProb = (float *)vProbs.pArray;
    // compute percentage of wires of each type
    if fVerbose != 0 {
        Abc_NtkPowerPrint( pNtk, vProbs ); }
    // mark the power critical nodes and edges
    // inlining of macro ABC_ALLOC
    puPCEdges = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Abc_NtkObjNumMax(pNtk))));
    memset( puPCEdges, 0, sizeof(unsigned) * Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if pProb[pNode.Id] < Limit {
            continue; }
        puPCEdges[pNode.Id] = Abc_NtkPowerCriticalEdges( pNtk, pNode, Limit, vProbs );
    }
    }
/*
    if ( fVerbose )
    {
        Counter = CounterRes = 0;
        Abc_NtkForEachNode( pNtk, pNode, i )
        {
            Counter += Abc_ObjFaninNum(pNode);
            CounterRes += Extra_WordCountOnes( puPCEdges[pNode->Id] );
        }
        printf( "Edges: Total = %7d. Critical = %7d. Ratio = %4.2f\n", 
            Counter, CounterRes, 1.0*CounterRes/Counter );
    }
*/
    // start the resulting network
    pNtkNew = Abc_NtkStrash( pNtk, 0, 1, 0 );

    // collect nodes to be used for resynthesis
    Counter = CounterRes = 0;
    vTimeCries = Vec_PtrAlloc( 16 );
    vTimeFanins = Vec_PtrAlloc( 16 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
//        if ( pProb[pNode->Id] < Limit )
//            continue;
        // count the number of non-PI power-critical nodes
        nTimeCris = 0;
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if !Abc_ObjIsCi(pFanin) && (puPCEdges[pNode.Id] & (1<<k)) {
                nTimeCris++; }
        }
        if !(fVeryVerbose != 0) && nTimeCris == 0 {
            continue; }
        Counter++;
        // count the total number of power-critical second-generation nodes
        Vec_PtrClear( vTimeCries );
        if nTimeCris != 0 {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if !Abc_ObjIsCi(pFanin) && (puPCEdges[pNode.Id] & (1<<k)) {
                    // inlining of macro Abc_ObjForEachFanin
                    for  k2 = 0; (k2 < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k2)), 1); k2++  {
                        if puPCEdges[pFanin.Id] & (1<<k2) {
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
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if Abc_ObjIsCi(pFanin) {
                Vec_PtrPushUnique( vTimeFanins, pFanin );
            } else {
                // inlining of macro Abc_ObjForEachFanin
                for  k2 = 0; (k2 < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k2)), 1); k2++  {
                    Vec_PtrPushUnique( vTimeFanins, pFanin2 );
                } }                    
        }
        // print the results
        if fVeryVerbose != 0 {
        fmt.Printf("%5d Node %5d : %d %2d %2d  ", Counter, pNode.Id, nTimeCris, Vec_PtrSize(vTimeCries), Vec_PtrSize(vTimeFanins));
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            fmt.Printf("%d(%.2f)%s ", pFanin.Id, pProb[pFanin.Id], (puPCEdges[pNode.Id] & (1<<k))? "*":"");
        }
        fmt.Printf("\n");
        }
        // add the node to choices
        if Vec_PtrSize(vTimeCries) == 0 || Vec_PtrSize(vTimeCries) > Degree {
            continue; }
        // order the fanins in the increasing order of criticalily
        if Vec_PtrSize(vTimeCries) > 1 {
            pFanin = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 0 );
            pFanin2 = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
//            if ( Abc_ObjSlack(pFanin) < Abc_ObjSlack(pFanin2) )
            if pProb[pFanin.Id] > pProb[pFanin2.Id] {
                Vec_PtrWriteEntry( vTimeCries, 0, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin );
            }
        }
        if Vec_PtrSize(vTimeCries) > 2 {
            pFanin = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
            pFanin2 = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 2 );
//            if ( Abc_ObjSlack(pFanin) < Abc_ObjSlack(pFanin2) )
            if pProb[pFanin.Id] > pProb[pFanin2.Id] {
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 2, pFanin );
            }
            pFanin = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 0 );
            pFanin2 = (Abc_Obj_t *)Vec_PtrEntry( vTimeCries, 1 );
//            if ( Abc_ObjSlack(pFanin) < Abc_ObjSlack(pFanin2) )
            if pProb[pFanin.Id] > pProb[pFanin2.Id] {
                Vec_PtrWriteEntry( vTimeCries, 0, pFanin2 );
                Vec_PtrWriteEntry( vTimeCries, 1, pFanin );
            }
        }
        // add choice
        Abc_NtkSpeedupNode( pNtk, pNtkNew, pNode, vTimeFanins, vTimeCries );
    }
    }
    Vec_PtrFree( vTimeCries );
    Vec_PtrFree( vTimeFanins );
    puPCEdges = nil;
    if fVerbose != 0 {
        fmt.Printf("Nodes: Total = %7d. Power-critical = %7d. Workable = %7d. Ratio = %4.2f\n", Abc_NtkNodeNum(pNtk), Counter, CounterRes, 1.0*CounterRes/Counter); } 

    // remove invalid choice nodes
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtkNew).vObjs)) && (((pNode) = Abc_NtkObj(pNtkNew, i)), 1); i++  {           if (pNode) == nil || !Abc_AigNodeIsAnd(pNode) {} else
        if pNode.pData {
            if Abc_ObjFanoutNum((Abc_Obj_t *)pNode.pData) > 0 {
                pNode.pData = nil; }
        }
    }

    // return the result
    Vec_IntFree( vProbs );
    return pNtkNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

