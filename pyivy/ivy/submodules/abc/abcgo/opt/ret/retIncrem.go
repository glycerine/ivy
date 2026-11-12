package ret

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [retIncrem.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Retiming package.]

  Synopsis    [Incremental retiming in one direction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Oct 31, 2006.]

  Revision    [$Id: retIncrem.c,v 1.00 2006/10/31 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "retInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Abc_NtkRetimeOneWay( Abc_Ntk_t * pNtk, int fForward, int fVerbose );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs retiming in one direction.]

  Description [Currently does not retime over black boxes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeIncremental(pNtk *Abc_Ntk_t, nDelayLim int32, fForward int32, fMinDelay int32, fOneStep int32, fUseOldNames int32, fVerbose int32) int32 {
    var pNtkCopy *Abc_Ntk_t  = nil;
    var vBoxes *Vec_Ptr_t;
    var tLatches *st__table;
    var nLatches int32  = Abc_NtkLatchNum(pNtk);
    var nIdMaxStart int32  = Abc_NtkObjNumMax(pNtk);
    var RetValue int32;
    var nIterLimit int32  = -1; // Suppress "might be used uninitialized"
    if Abc_NtkNodeNum(pNtk) == 0 {
        return 0; }
    // reorder CI/CO/latch inputs
    Abc_NtkOrderCisCos( pNtk );
    if fMinDelay != 0 {
        if fOneStep != 0 {
            nIterLimit = 1;
        } else {
            nIterLimit = 2 * Abc_NtkLevel(pNtk);
        }
        pNtkCopy = Abc_NtkDup( pNtk );
        tLatches = Abc_NtkRetimePrepareLatches( pNtkCopy );
        st__free_table( tLatches );
    }
    // collect latches and remove CIs/COs
    tLatches = Abc_NtkRetimePrepareLatches( pNtk );
    // share the latches
    Abc_NtkRetimeShareLatches( pNtk, 0 );    
    // save boxes
    vBoxes = pNtk.vBoxes;  pNtk.vBoxes = nil;
    // perform the retiming
    if fMinDelay != 0 {
        Abc_NtkRetimeMinDelay( pNtk, pNtkCopy, nDelayLim, nIterLimit, fForward, fVerbose );
    } else {
        Abc_NtkRetimeOneWay( pNtk, fForward, fVerbose ); }
    if fMinDelay != 0 { 
        Abc_NtkDelete( pNtkCopy ); }
    // share the latches
    Abc_NtkRetimeShareLatches( pNtk, 0 );    
    // restore boxes
    pNtk.vBoxes = vBoxes;
    // finalize the latches
    RetValue = Abc_NtkRetimeFinalizeLatches( pNtk, tLatches, nIdMaxStart, fUseOldNames );
    st__free_table( tLatches );
    if RetValue == 0 {
        return 0; }
    // fix the COs
//    Abc_NtkLogicMakeSimpleCos( pNtk, 0 );
    // check for correctness
    if !Abc_NtkCheck( pNtk ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkRetimeForward(): Network check has failed.\n"); }
    // return the number of latches saved
    return nLatches - Abc_NtkLatchNum(pNtk);
}

/**Function*************************************************************

  Synopsis    [Prepares the network for retiming.]

  Description [Hash latches into their number in the original network.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
 func Abc_NtkRetimePrepareLatches(pNtk *Abc_Ntk_t) *st__table {
    var tLatches *st__table;
    var pLatch, pLatchIn, pLatchOut, pFanin *Abc_Obj_t;
    var i int32;
    var nOffSet int32 = Abc_NtkBoxNum(pNtk) - Abc_NtkLatchNum(pNtk);
    // collect latches and remove CIs/COs
    tLatches = st__init_table( st__ptrcmp, st__ptrhash );
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
        // map latch into its true number
        st__insert( tLatches, (char *)(ABC_PTRUINT_T)pLatch, (char *)(ABC_PTRUINT_T)(i-nOffSet) );
        // disconnect LI     
        pLatchIn = Abc_ObjFanin0(pLatch);
        pFanin = Abc_ObjFanin0(pLatchIn);
        Abc_ObjTransferFanout( pLatchIn, pFanin );
        Abc_ObjDeleteFanin( pLatchIn, pFanin );
        // disconnect LO     
        pLatchOut = Abc_ObjFanout0(pLatch);
        pFanin = Abc_ObjFanin0(pLatchOut);
        if Abc_ObjFanoutNum(pLatchOut) > 0 {
            Abc_ObjTransferFanout( pLatchOut, pFanin ); }
        Abc_ObjDeleteFanin( pLatchOut, pFanin );
    }
    }
    return tLatches;
}

/**Function*************************************************************

  Synopsis    [Finalizes the latches after retiming.]

  Description [Reuses the LIs/LOs for old latches.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeFinalizeLatches(pNtk *Abc_Ntk_t, tLatches *st__table, nIdMaxStart int32, fUseOldNames int32) int32 {
    var vCisOld, vCosOld, vBoxesOld, vCisNew, vCosNew, vBoxesNew *Vec_Ptr_t;
    var pObj, pLatch, pLatchIn, pLatchOut *Abc_Obj_t;
    var i, Index int32;
    // create new arrays
    vCisOld   = pNtk.vCis;    pNtk.vCis   = nil;  vCisNew   = Vec_PtrAlloc( 100 );
    vCosOld   = pNtk.vCos;    pNtk.vCos   = nil;  vCosNew   = Vec_PtrAlloc( 100 );  
    vBoxesOld = pNtk.vBoxes;  pNtk.vBoxes = nil;  vBoxesNew = Vec_PtrAlloc( 100 );
    // copy boxes and their CIs/COs
    // inlining of macro Vec_PtrForEachEntryStop
    // inlining of macro st__count
    for  i = 0; (i < Vec_PtrSize(vCisOld) - ((tLatches).num_entries)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vCisOld, i)), 1); i++  {
        Vec_PtrPush( vCisNew, pObj );
    }
    // inlining of macro Vec_PtrForEachEntryStop
    // inlining of macro st__count
    for  i = 0; (i < Vec_PtrSize(vCosOld) - ((tLatches).num_entries)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vCosOld, i)), 1); i++  {
        Vec_PtrPush( vCosNew, pObj );
    }
    // inlining of macro Vec_PtrForEachEntryStop
    // inlining of macro st__count
    for  i = 0; (i < Vec_PtrSize(vBoxesOld) - ((tLatches).num_entries)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vBoxesOld, i)), 1); i++  {
        Vec_PtrPush( vBoxesNew, pObj );
    }
    // go through the latches
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pLatch) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pLatch) == nil {} else {
        if !Abc_ObjIsLatch(pLatch) {
            continue; }
        if Abc_ObjId(pLatch) >= uint32(nIdMaxStart) {
            // this is a new latch 
            pLatchIn  = Abc_NtkCreateBi(pNtk);
            pLatchOut = Abc_NtkCreateBo(pNtk);

            if fUseOldNames != 0 {
                Abc_ObjAssignName( pLatchOut, Abc_ObjName(pLatch), "_out" );
                Abc_ObjAssignName( pLatchIn,  Abc_ObjName(pLatch), "_in" );
            
            } else {
                Abc_ObjAssignName( pLatchOut, Abc_ObjName(Abc_ObjFanin0(pLatch)), "_o2" );
                Abc_ObjAssignName( pLatchIn,  Abc_ObjName(Abc_ObjFanin0(pLatch)), "_i2" );
            }
        
        } else {
            // this is an old latch 
            // get its number in the original order
            if ! st__lookup_int( tLatches, (char *)pLatch, &Index ) {
                fmt.Printf("Abc_NtkRetimeFinalizeLatches(): Internal error.\n");
                return 0;
            }
            // inlining of macro st__count
            assert(pLatch == Vec_PtrEntry(vBoxesOld, Vec_PtrSize(vBoxesOld) - ((tLatches).num_entries) + Index), "pLatch == Vec_PtrEntry(vBoxesOld, Vec_PtrSize(vBoxesOld) - /*__starting_point_inline_macro:st__count__*/((tLatches).num_entries) + Index)");
            // reconnect with the old LIs/LOs
            // inlining of macro st__count
            pLatchIn  = (Abc_Obj_t *)Vec_PtrEntry( vCosOld, Vec_PtrSize(vCosOld) - ((tLatches).num_entries) + Index );
            // inlining of macro st__count
            pLatchOut = (Abc_Obj_t *)Vec_PtrEntry( vCisOld, Vec_PtrSize(vCisOld) - ((tLatches).num_entries) + Index );
        }
        // connect
        Abc_ObjAddFanin( pLatchIn, Abc_ObjFanin0(pLatch) );
        Abc_ObjPatchFanin( pLatch, Abc_ObjFanin0(pLatch), pLatchIn );
        if Abc_ObjFanoutNum(pLatch) > 0 {
            Abc_ObjTransferFanout( pLatch, pLatchOut ); }
        Abc_ObjAddFanin( pLatchOut, pLatch );
        // add to the arrays
        Vec_PtrPush( vCisNew, pLatchOut );
        Vec_PtrPush( vCosNew, pLatchIn );
        Vec_PtrPush( vBoxesNew, pLatch );
    }
    }
    // free useless Cis/Cos
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCisOld)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vCisOld, i)), 1); i++  {
        if !Abc_ObjIsPi(pObj) && Abc_ObjFaninNum(pObj) == 0 && Abc_ObjFanoutNum(pObj) == 0 {
            Abc_NtkDeleteObj(pObj); }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCosOld)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vCosOld, i)), 1); i++  {
        if !Abc_ObjIsPo(pObj) && Abc_ObjFaninNum(pObj) == 0 && Abc_ObjFanoutNum(pObj) == 0 {
            Abc_NtkDeleteObj(pObj); }
    }
    // set the new arrays
    pNtk.vCis   = vCisNew;   Vec_PtrFree( vCisOld );
    pNtk.vCos   = vCosNew;   Vec_PtrFree( vCosOld );
    pNtk.vBoxes = vBoxesNew; Vec_PtrFree( vBoxesOld );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Performs retiming one way, forward or backward.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeOneWay(pNtk *Abc_Ntk_t, fForward int32, fVerbose int32) int32 {
    var pNtkNew *Abc_Ntk_t  = nil; // Suppress "might be used uninitialized"
    var vValues *Vec_Int_t  = nil; // Suppress "might be used uninitialized"
    var pObj *Abc_Obj_t;
    var i int32;
    var fChanges int32;
    var nTotalMoves int32;
    var nTotalMovesLimit int32 = 10000;
    if fForward != 0 {
        Abc_NtkRetimeTranferToCopy( pNtk );
    } else {
        // save initial values of the latches
        vValues = Abc_NtkRetimeCollectLatchValues( pNtk );
        // start the network for initial value computation
        pNtkNew = Abc_NtkRetimeBackwardInitialStart( pNtk );
    }
    // try to move latches forward whenever possible
    do {
        fChanges = 0;
        // inlining of macro Abc_NtkForEachObj
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
            if !Abc_ObjIsNode(pObj) {
                continue; }
            if Abc_NtkRetimeNodeIsEnabled( pObj, fForward ) != 0 {
                Abc_NtkRetimeNode( pObj, fForward, 1 );
                fChanges = 1;
                nTotalMoves++;
                if nTotalMoves >= nTotalMovesLimit {
                    fmt.Printf("Stopped after %d latch moves.\n", nTotalMoves);
                    break;
                }
            }
        }
        }
    } while ( fChanges && nTotalMoves < nTotalMovesLimit );
    // transfer the initial state back to the latches
    if fForward != 0 {
        Abc_NtkRetimeTranferFromCopy( pNtk );
    } else {
        Abc_NtkRetimeBackwardInitialFinish( pNtk, pNtkNew, vValues, fVerbose );
        Abc_NtkDelete( pNtkNew );
        Vec_IntFree( vValues );
    }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if retiming forward/backward is possible.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeNodeIsEnabled(pObj *Abc_Obj_t, fForward int32) int32 {
    var pNext *Abc_Obj_t;
    var i int32;
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    if fForward != 0 {
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
            if !Abc_ObjIsLatch(pNext) {
                return 0; }
        }
    
    } else {
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
            if !Abc_ObjIsLatch(pNext) {
                return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Retimes the node backward or forward.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeNode(pObj *Abc_Obj_t, fForward int32, fInitial int32) {
    var pNtkNew *Abc_Ntk_t  = nil;
    var vNodes *Vec_Ptr_t;
    var pNext, pLatch *Abc_Obj_t;
    var i int32;
    vNodes = Vec_PtrAlloc( 10 );
    if fForward != 0 {
        // compute the initial value
        if fInitial != 0 {
            pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)Abc_ObjSopSimulate( pObj ); }
        // collect fanins
        Abc_NodeCollectFanins( pObj, vNodes );
        // make the node point to the fanins fanins
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNext) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            assert(Abc_ObjIsLatch(pNext), "Abc_ObjIsLatch(pNext)");
            Abc_ObjPatchFanin( pObj, pNext, Abc_ObjFanin0(pNext) );
            if Abc_ObjFanoutNum(pNext) == 0 {
                Abc_NtkDeleteObj(pNext); }            
        }
        // add a new latch on top
        pNext = Abc_NtkCreateLatch(pObj.pNtk);
        if Abc_ObjFanoutNum(pObj) > 0 {
            Abc_ObjTransferFanout( pObj, pNext ); }
        Abc_ObjAddFanin( pNext, pObj );
        // set the initial value
        if fInitial != 0 {
            pNext.pCopy = pObj.pCopy; }
    
    } else {
        // compute the initial value
        if fInitial != 0 {
            pNtkNew = Abc_ObjFanout0(pObj).pCopy.pNtk;
            Abc_NtkDupObj( pNtkNew, pObj, 0 );
            // inlining of macro Abc_ObjForEachFanout
            for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
                assert(Abc_ObjFaninNum(pNext.pCopy) == 0, "Abc_ObjFaninNum(pNext.pCopy) == 0");
                Abc_ObjAddFanin( pNext.pCopy, pObj.pCopy );
            }
        }
        // collect fanouts
        Abc_NodeCollectFanouts( pObj, vNodes );
        // make the fanouts fanouts point to the node
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNext) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            assert(Abc_ObjIsLatch(pNext), "Abc_ObjIsLatch(pNext)");
            Abc_ObjTransferFanout( pNext, pObj );
            Abc_NtkDeleteObj( pNext );  
        }
        // add new latches to the fanins
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
            pLatch = Abc_NtkCreateLatch(pObj.pNtk);
            Abc_ObjPatchFanin( pObj, pNext, pLatch );
            Abc_ObjAddFanin( pLatch, pNext );
            // create buffer isomorphic to this latch
            if fInitial != 0 {
                pLatch.pCopy = Abc_NtkCreateNodeBuf( pNtkNew, nil );
                Abc_ObjAssignName( pLatch.pCopy, Abc_ObjName(pNext), "_buf" );
                Abc_ObjAddFanin( pObj.pCopy, pLatch.pCopy );
            }
        }
    }
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Returns the number of compatible fanout latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeCheckCompatibleLatchFanouts(pObj *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i int32;
    var nLatches int32;
    var Init int32 = -1;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if !Abc_ObjIsLatch(pFanout) {
            continue; }
        if Init == -1 {
            Init = int32((ABC_PTRUINT_T))pObj.pData;
            nLatches++;
        
        } else if Init == int32((ABC_PTRUINT_T))pObj.pData {
            nLatches++; }
    }
    return nLatches;
}

/**Function*************************************************************

  Synopsis    [Retimes the node backward or forward.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeShareLatches(pNtk *Abc_Ntk_t, fInitial int32) {
    var vNodes *Vec_Ptr_t;
    var pFanin, pLatchTop, pLatchCur *Abc_Obj_t;
    var i, k int32;
    vNodes = Vec_PtrAlloc( 10 );
    // consider latch fanins
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pFanin) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pFanin) == nil {} else {
        if Abc_NtkRetimeCheckCompatibleLatchFanouts(pFanin) <= 1 {
            continue; }
        // get the first latch
        pLatchTop = nil;
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pFanin)) && (((pLatchTop) = Abc_ObjFanout(pFanin, k)), 1); k++  {
            if Abc_ObjIsLatch(pLatchTop) {
                break; }
        }
        assert(pLatchTop && Abc_ObjIsLatch(pLatchTop), "pLatchTop && Abc_ObjIsLatch(pLatchTop)");
        // redirect compatible fanout latches to the first latch
        Abc_NodeCollectFanouts( pFanin, vNodes );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes)) && (((pLatchCur) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, k)), 1); k++  {
            if !Abc_ObjIsLatch(pLatchCur) {
                continue; }
            if pLatchCur == pLatchTop {
                continue; }
            if pLatchCur.pData != pLatchTop.pData {
                continue; }
            // connect the initial state
            if fInitial != 0 {
                Abc_ObjAddFanin( pLatchCur.pCopy, pLatchTop.pCopy ); }
            // redirect the fanouts
            Abc_ObjTransferFanout( pLatchCur, pLatchTop );
            Abc_NtkDeleteObj(pLatchCur);
        }
    }
    }
    Vec_PtrFree( vNodes );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

