package ret

import "fmt"

/**CFile****************************************************************

  FileName    [retDelay.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Retiming package.]

  Synopsis    [Incremental retiming for optimum delay.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Oct 31, 2006.]

  Revision    [$Id: retDelay.c,v 1.00 2006/10/31 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "retInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Abc_NtkRetimeMinDelayTry( Abc_Ntk_t * pNtk, int nDelayLim, int fForward, int fInitial, int nIterLimit, int * pIterBest, int fVerbose );

// static int Abc_NtkRetimeTiming( Abc_Ntk_t * pNtk, int fForward, Vec_Ptr_t * vCritical );

// static int Abc_NtkRetimeTiming_rec( Abc_Obj_t * pObj, int fForward );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Retimes incrementally for minimum delay.]

  Description [This procedure cannot be called in the application code
  because it assumes that the network is preprocessed by removing LIs/LOs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeMinDelay(pNtk *Abc_Ntk_t, pNtkCopy *Abc_Ntk_t, nDelayLim int32, nIterLimit int32, fForward int32, fVerbose int32) int32 {
    var IterBest, DelayBest int32;
    var IterBest2, DelayBest2 int32;
    // try to find the best delay iteration on a copy
    DelayBest = Abc_NtkRetimeMinDelayTry( pNtkCopy, nDelayLim, fForward, 0, nIterLimit, &IterBest, fVerbose );
    if IterBest == 0 {
        return 1; }
    // perform the given number of iterations on the original network
    DelayBest2 = Abc_NtkRetimeMinDelayTry( pNtk, nDelayLim, fForward, 1, IterBest, &IterBest2, fVerbose );
    assert(DelayBest == DelayBest2, "DelayBest == DelayBest2");
    assert(IterBest == IterBest2, "IterBest == IterBest2");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns the best delay and the number of best iteration.]

  Description [] 
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeMinDelayTry(pNtk *Abc_Ntk_t, nDelayLim int32, fForward int32, fInitial int32, nIterLimit int32, pIterBest *int32, fVerbose int32) int32 {
    var pNtkNew *Abc_Ntk_t  = nil;
    var vCritical *Vec_Ptr_t;
    var vValues *Vec_Int_t  = nil; // Suppress "might be used uninitialized"
    var pObj *Abc_Obj_t;
    var i, k, IterBest, DelayCur, DelayBest int32;
    var DelayStart int32  = -1; // Suppress "might be used uninitialized"
    var LatchesBest int32;
    // transfer intitial values
    if fInitial != 0 {
        if fForward != 0 {
            Abc_NtkRetimeTranferToCopy( pNtk );
        } else {
            // save initial value of the latches
            vValues = Abc_NtkRetimeCollectLatchValues( pNtk );
            // start the network for initial value computation
            pNtkNew = Abc_NtkRetimeBackwardInitialStart( pNtk );
        }
    }

    if (fVerbose != 0) && !(fInitial != 0) {
        fmt.Printf("Performing analysis:\n"); }
    // find the best iteration
    DelayBest = ABC_INFINITY; IterBest = 0; LatchesBest = Abc_NtkLatchNum(pNtk);
    vCritical = Vec_PtrAlloc( 100 );
    for  i = 0; ; i++  {
        // perform moves for the timing-critical nodes
        DelayCur = Abc_NtkRetimeTiming( pNtk, fForward, vCritical );
        if i == 0 {
            DelayStart = DelayCur; }
        // record this position if it has the best delay
        if DelayBest > DelayCur {
if (fVerbose != 0) && !(fInitial != 0) {
    fmt.Printf("%s Iter = %3d. Delay = %3d. Latches = %5d. Delta = %6.2f. Ratio = %4.2f %%\n", fForward ? "Fwd": "Bwd", i, DelayCur, Abc_NtkLatchNum(pNtk), 1.0*(Abc_NtkLatchNum(pNtk)-LatchesBest)/(DelayBest-DelayCur), 100.0*(Abc_NtkLatchNum(pNtk)-LatchesBest)/Abc_NtkLatchNum(pNtk)/(DelayBest-DelayCur)); }

            DelayBest = DelayCur;
            IterBest = i;
            LatchesBest = Abc_NtkLatchNum(pNtk);
        }
        // quit after timing analysis
        if i == nIterLimit {
            break; }
        // skip if 10 iterations did not give improvement
        if i - IterBest > 20 {
            break; }
        // skip if delay limit is reached
        if nDelayLim > 0 && DelayCur <= nDelayLim {
            break; }
        // try retiming to improve the delay
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vCritical)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vCritical, k)), 1); k++  {
            if Abc_NtkRetimeNodeIsEnabled(pObj, fForward) {
                Abc_NtkRetimeNode( pObj, fForward, fInitial ); }
        }
        // share latches
        if !(fForward != 0) {
            Abc_NtkRetimeShareLatches( pNtk, fInitial ); }    
    }
    Vec_PtrFree( vCritical );
    // transfer the initial state back to the latches
    if fInitial != 0 {
        if fForward != 0 {
            Abc_NtkRetimeTranferFromCopy( pNtk );
        } else {
            Abc_NtkRetimeBackwardInitialFinish( pNtk, pNtkNew, vValues, fVerbose );
            Abc_NtkDelete( pNtkNew );
            Vec_IntFree( vValues );
        }
    }
    if (fVerbose != 0) && !(fInitial != 0) {
        fmt.Printf("%s : Starting delay = %3d.  Final delay = %3d.  IterBest = %2d (out of %2d).\n", fForward? "Forward " : "Backward", DelayStart, DelayBest, IterBest, nIterLimit); }
    *pIterBest = (nIterLimit == 1) ? 1 : IterBest;
    return DelayBest;
}

/**Function*************************************************************

  Synopsis    [Returns the set of timing-critical nodes.]

  Description [Performs static timing analysis on the network. Uses 
  unit-delay model.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeTiming(pNtk *Abc_Ntk_t, fForward int32, vCritical *Vec_Ptr_t) int32 {
    var vLatches *Vec_Ptr_t;
    var pObj, pNext *Abc_Obj_t;
    var i, k, LevelCur, LevelMax int32;
    // mark all objects except nodes
    Abc_NtkIncrementTravId(pNtk);
    vLatches = Vec_PtrAlloc( Abc_NtkLatchNum(pNtk) );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        if Abc_ObjIsLatch(pObj) {
            Vec_PtrPush( vLatches, pObj ); }
        if Abc_ObjIsNode(pObj) {
            continue; }
        pObj.Level = 0;
        Abc_NodeSetTravIdCurrent( pObj );
    }
    }
    // perform analysis from CIs/COs
    if fForward != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vLatches)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLatches, i)), 1); i++  {
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
                LevelCur = Abc_NtkRetimeTiming_rec( pNext, fForward );
                if LevelMax < LevelCur {
                    LevelMax = LevelCur; }
            }
        }
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
                LevelCur = Abc_NtkRetimeTiming_rec( pNext, fForward );
                if LevelMax < LevelCur {
                    LevelMax = LevelCur; }
            }
        }
    
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vLatches)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLatches, i)), 1); i++  {
            LevelCur = Abc_NtkRetimeTiming_rec( Abc_ObjFanin0(pObj), fForward );
            if LevelMax < LevelCur {
                LevelMax = LevelCur; }
        }
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            LevelCur = Abc_NtkRetimeTiming_rec( Abc_ObjFanin0(pObj), fForward );
            if LevelMax < LevelCur {
                LevelMax = LevelCur; }
        }
    }
    // collect timing critical nodes, which should be retimed forward/backward
    Vec_PtrClear( vCritical );
    Abc_NtkIncrementTravId(pNtk);
    if fForward != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vLatches)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLatches, i)), 1); i++  {
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
                if Abc_NodeIsTravIdCurrent(pNext) {
                    continue; }
                if LevelMax != int32(pNext.Level) {
                    continue; }
                // new critical node
                Vec_PtrPush( vCritical, pNext );
                Abc_NodeSetTravIdCurrent( pNext );
            }
        }
    
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vLatches)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLatches, i)), 1); i++  {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, k)), 1); k++  {
                if Abc_NodeIsTravIdCurrent(pNext) {
                    continue; }
                if LevelMax != int32(pNext.Level) {
                    continue; }
                // new critical node
                Vec_PtrPush( vCritical, pNext );
                Abc_NodeSetTravIdCurrent( pNext );
            }
        }
    }
    Vec_PtrFree( vLatches );
    return LevelMax;
}

/**Function*************************************************************

  Synopsis    [Recursively performs timing analysis.]

  Description [Performs static timing analysis on the network. Uses 
  unit-delay model.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeTiming_rec(pObj *Abc_Obj_t, fForward int32) int32 {
    var pNext *Abc_Obj_t;
    var i, LevelCur, LevelMax int32;
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pObj) {
        return pObj.Level; }
    Abc_NodeSetTravIdCurrent(pObj);
    // visit the next nodes
    if fForward != 0 {
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
            LevelCur = Abc_NtkRetimeTiming_rec( pNext, fForward );
            if LevelMax < LevelCur {
                LevelMax = LevelCur; }
        }
    
    } else {
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
            LevelCur = Abc_NtkRetimeTiming_rec( pNext, fForward );
            if LevelMax < LevelCur {
                LevelMax = LevelCur; }
        }
    }
//    printf( "Node %3d -> Level %3d.\n", pObj->Id, LevelMax + 1 );
    pObj.Level = LevelMax + 1;
    return pObj.Level;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

