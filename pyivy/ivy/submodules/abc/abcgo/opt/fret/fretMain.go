package fret

import (
	"fmt"
	"os"
	. "github.com/glycerine/abcgo/base/abc"
	. "github.com/glycerine/abcgo/aig/hop"
	. "github.com/glycerine/abcgo/map/mio"
	. "github.com/glycerine/abcgo/misc/mem"
	. "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [fretMain.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Flow-based retiming package.]

  Synopsis    [Main file for retiming package.]

  Author      [Aaron Hurst]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2008.]

  Revision    [$Id: fretMain.c,v 1.00 2008/01/01 00:00:00 ahurst Exp $]

***********************************************************************/

//#include "fretime.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Abc_FlowRetime_AddDummyFanin( Abc_Obj_t * pObj );


// static Abc_Ntk_t* Abc_FlowRetime_MainLoop( );


// static void Abc_FlowRetime_MarkBlocks( Abc_Ntk_t * pNtk );

// static void Abc_FlowRetime_MarkReachable_rec( Abc_Obj_t * pObj, char end );

// static int  Abc_FlowRetime_ImplementCut( Abc_Ntk_t * pNtk );

// static void  Abc_FlowRetime_RemoveLatchBubbles( Abc_Obj_t * pLatch );


// static Abc_Ntk_t* Abc_FlowRetime_NtkDup( Abc_Ntk_t * pNtk );


// static void Abc_FlowRetime_VerifyPathLatencies( Abc_Ntk_t * pNtk );

// static int  Abc_FlowRetime_VerifyPathLatencies_rec( Abc_Obj_t * pObj, int markD );


// static void Abc_FlowRetime_UpdateLags_forw_rec( Abc_Obj_t *pObj );

// static void Abc_FlowRetime_UpdateLags_back_rec( Abc_Obj_t *pObj );


// extern void Abc_NtkMarkCone_rec( Abc_Obj_t * pObj, int fForward );


// void
// print_node3(Abc_Obj_t *pObj);


var pManMR *MinRegMan_t;

var fPathError int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs minimum-register retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_MinReg(pNtk *Abc_Ntk_t, fVerbose int32, fComputeInitState int32, fGuaranteeInitState int32, fBlockConst int32, fForwardOnly int32, fBackwardOnly int32, nMaxIters int32, maxDelay int32, fFastButConservative int32) *Abc_Ntk_t {

  var i int32;
  var pObj, pNext *Abc_Obj_t;
  var pData *InitConstraint_t;

  // create manager
  pManMR = new(MinRegMan_t);

  pManMR.pNtk = pNtk;
  pManMR.fVerbose = fVerbose;
  pManMR.fComputeInitState = fComputeInitState;
  pManMR.fGuaranteeInitState = fGuaranteeInitState;
  pManMR.fBlockConst = fBlockConst;
  pManMR.fForwardOnly = fForwardOnly;
  pManMR.fBackwardOnly = fBackwardOnly;
  pManMR.nMaxIters = nMaxIters;
  pManMR.maxDelay = maxDelay;
  pManMR.fComputeInitState = fComputeInitState;
  pManMR.fConservTimingOnly = fFastButConservative;
  pManMR.vNodes = Vec_PtrAlloc(100);
  pManMR.vInitConstraints = Vec_PtrAlloc(2);
  pManMR.pInitNtk = nil;
  pManMR.pInitToOrig = nil;

  vprintf("Flow-based minimum-register retiming...\n");  

  if !Abc_NtkHasOnlyLatchBoxes(pNtk) {
    fmt.Printf("\tERROR: Can not retime with black/white boxes\n");
    return pNtk;
  }

  if maxDelay != 0 {
    vprintf("\tmax delay constraint = %d\n", maxDelay);
    if maxDelay < (i = Abc_NtkLevel(pNtk)) {
      fmt.Printf("ERROR: max delay constraint (%d) must be > current max delay (%d)\n", maxDelay, i);
      return pNtk;
    }
  }

  // print info about type of network
  vprintf("\tnetlist type = ");
  if Abc_NtkIsNetlist( pNtk ) { vprintf("netlist/"); 
  } else if Abc_NtkIsLogic( pNtk ) { vprintf("logic/"); 
  } else if Abc_NtkIsStrash( pNtk ) { vprintf("strash/"); 
  } else { vprintf("***unknown***/"); }
  if Abc_NtkHasSop( pNtk ) { vprintf("sop\n"); 
  } else if Abc_NtkHasBdd( pNtk ) { vprintf("bdd\n"); 
  } else if Abc_NtkHasAig( pNtk ) { vprintf("aig\n"); 
  } else if Abc_NtkHasMapping( pNtk ) { vprintf("mapped\n"); 
  } else { vprintf("***unknown***\n"); }

  vprintf("\tinitial reg count = %d\n", Abc_NtkLatchNum(pNtk));
  vprintf("\tinitial levels = %d\n", Abc_NtkLevel(pNtk));

  // remove bubbles from latch boxes
  if pManMR.fVerbose { Abc_FlowRetime_PrintInitStateInfo(pNtk); }
  vprintf("\tpushing bubbles out of latch boxes\n");
  Abc_NtkForEachLatch( pNtk, pObj, i )
    Abc_FlowRetime_RemoveLatchBubbles(pObj);
  if pManMR.fVerbose { Abc_FlowRetime_PrintInitStateInfo(pNtk); }

  // check for box inputs/outputs
  Abc_NtkForEachLatch( pNtk, pObj, i ) {
    assert(Abc_ObjFaninNum(pObj) == 1, "Abc_ObjFaninNum(pObj) == 1");
    assert(Abc_ObjFanoutNum(pObj) == 1, "Abc_ObjFanoutNum(pObj) == 1");
    assert(!Abc_ObjFaninC0(pObj), "!Abc_ObjFaninC0(pObj)");

    pNext = Abc_ObjFanin0(pObj);
    assert(Abc_ObjIsBi(pNext), "Abc_ObjIsBi(pNext)");
    assert(Abc_ObjFaninNum(pNext) <= 1, "Abc_ObjFaninNum(pNext) <= 1");
    if Abc_ObjFaninNum(pNext) == 0 { // every Bi should have a fanin
      Abc_FlowRetime_AddDummyFanin( pNext ); }
 
    pNext = Abc_ObjFanout0(pObj);
    assert(Abc_ObjIsBo(pNext), "Abc_ObjIsBo(pNext)");
    assert(Abc_ObjFaninNum(pNext) == 1, "Abc_ObjFaninNum(pNext) == 1");
    assert(!Abc_ObjFaninC0(pNext), "!Abc_ObjFaninC0(pNext)");
  }

  pManMR.nLatches = Abc_NtkLatchNum( pNtk );
  pManMR.nNodes = Abc_NtkObjNumMax( pNtk )+1;
   
  // build histogram
  pManMR.vSinkDistHist = Vec_IntStart( pManMR.nNodes*2+10 );

  // initialize timing
  if maxDelay != 0 {
    Abc_FlowRetime_InitTiming( pNtk ); }

  // create lag and Flow_Data structure
  pManMR.vLags = Vec_IntStart(pManMR.nNodes);
  memset(pManMR.vLags.pArray, 0, sizeof(int)*pManMR.nNodes);

  pManMR.pDataArray = ABC_ALLOC( Flow_Data_t, pManMR.nNodes );
  Abc_FlowRetime_ClearFlows( 1 );

  // main loop!
  pNtk = Abc_FlowRetime_MainLoop();

  // cleanup node fields
  Abc_NtkForEachObj( pNtk, pObj, i ) {
    // if not computing init state, set all latches to DC
    if !(fComputeInitState != 0) && Abc_ObjIsLatch(pObj) {
      Abc_LatchSetInitDc(pObj); }
  }

  // deallocate space
  ABC_FREE( pManMR.pDataArray );
  if pManMR.pInitToOrig { ABC_FREE( pManMR.pInitToOrig ); }
  if pManMR.vNodes { Vec_PtrFree(pManMR.vNodes); }
  if pManMR.vLags { Vec_IntFree(pManMR.vLags); }
  if pManMR.vSinkDistHist { Vec_IntFree(pManMR.vSinkDistHist); }
  if pManMR.maxDelay { Abc_FlowRetime_FreeTiming( pNtk ); }
  for  Vec_PtrSize( pManMR.vInitConstraints ) {
    pData = (InitConstraint_t*)Vec_PtrPop( pManMR.vInitConstraints );
    //assert( pData->pBiasNode );
    //Abc_NtkDeleteObj( pData->pBiasNode );
    ABC_FREE( pData.vNodes.pArray );
    ABC_FREE( pData );
  }
  ABC_FREE( pManMR.vInitConstraints );

  // restrash if necessary
  if Abc_NtkIsStrash(pNtk) {
    Abc_NtkReassignIds( pNtk );
    pNtk = Abc_FlowRetime_NtkSilentRestrash( pNtk, 1 );
  }
  
  vprintf("\tfinal reg count = %d\n", Abc_NtkLatchNum(pNtk));
  vprintf("\tfinal levels = %d\n", Abc_NtkLevel(pNtk));

#if defined(DEBUG_CHECK)
  Abc_NtkDoCheck( pNtk );
#endif

  // free manager
  ABC_FREE( pManMR );

  return pNtk;
}

/**Function*************************************************************

  Synopsis    [Main loop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_MainLoop() *Abc_Ntk_t {
  var pNtk *Abc_Ntk_t = pManMR.pNtk;
  var pNtkCopy *Abc_Ntk_t = pNtk;
  var pObj *Abc_Obj_t; var i int32;
  var last, flow, cut int32;

  // (i) forward retiming loop
  pManMR.fIsForward = 1;
  pManMR.iteration = 0;

  if !pManMR.fBackwardOnly { for {
    if pManMR.iteration == pManMR.nMaxIters { break; }
    pManMR.subIteration = 0;

    vprintf("\tforward iteration %d\n", pManMR.iteration);
    last = Abc_NtkLatchNum( pNtk );

    Abc_FlowRetime_MarkBlocks( pNtk );

    if pManMR.maxDelay {
      // timing-constrained loop
      Abc_FlowRetime_ConstrainConserv( pNtk );
      for Abc_FlowRetime_RefineConstraints( ) {
        pManMR.subIteration++;
        Abc_FlowRetime_ClearFlows( 0 );
      }
    } else {
      flow = Abc_FlowRetime_PushFlows( pNtk, 1 );
    }

    cut = Abc_FlowRetime_ImplementCut( pNtk );

#if defined (DEBUG_PRINT_LEVELS)
    vprintf("\t\tlevels = %d\n", Abc_NtkLevel(pNtk));
#endif

    Abc_FlowRetime_ClearFlows( 1 );

    pManMR.iteration++;
  
if !( cut != last ) { break }
}  }
  
  // intermediate cleanup (for strashed networks)
  if Abc_NtkIsStrash(pNtk) {
    Abc_NtkReassignIds( pNtk );
    pNtk = pManMR.pNtk = Abc_FlowRetime_NtkSilentRestrash( pNtk, 1 );
  }

  // print info about initial states
  if pManMR.fComputeInitState && pManMR.fVerbose {
    Abc_FlowRetime_PrintInitStateInfo( pNtk ); }

  // (ii) backward retiming loop
  pManMR.fIsForward = 0;

  if !pManMR.fForwardOnly { for {
    // initializability loop
    pManMR.iteration = 0;

    // copy/restore network
    if pManMR.fGuaranteeInitState {
      if pNtk != pNtkCopy {
        Abc_NtkDelete( pNtk ); }
      pNtk = pManMR.pNtk = Abc_FlowRetime_NtkDup( pNtkCopy );
      vprintf("\trestoring network. regs = %d\n", Abc_NtkLatchNum( pNtk ));
    }

    if pManMR.fComputeInitState {
      Abc_FlowRetime_SetupBackwardInit( pNtk );
    }

    for {
      if pManMR.iteration == pManMR.nMaxIters { break; }
      pManMR.subIteration = 0;

      vprintf("\tbackward iteration %d\n", pManMR.iteration);
      last = Abc_NtkLatchNum( pNtk );

      Abc_FlowRetime_AddInitBias( );
      Abc_FlowRetime_MarkBlocks( pNtk );

      if pManMR.maxDelay {
        // timing-constrained loop
        Abc_FlowRetime_ConstrainConserv( pNtk );
        for Abc_FlowRetime_RefineConstraints( ) {
          pManMR.subIteration++;
          Abc_FlowRetime_ClearFlows( 0 );
        }
      } else {
        flow = Abc_FlowRetime_PushFlows( pNtk, 1 );
      }

      Abc_FlowRetime_RemoveInitBias( );
      cut = Abc_FlowRetime_ImplementCut( pNtk );

#if defined(DEBUG_PRINT_LEVELS)
      vprintf("\t\tlevels = %d\n", Abc_NtkLevelReverse(pNtk));
#endif      

      Abc_FlowRetime_ClearFlows( 1 );

      pManMR.iteration++;
    
if !( cut != last ) { break }
} 
    
    // compute initial states
    if !pManMR.fComputeInitState { break; }

    if Abc_FlowRetime_SolveBackwardInit( pNtk ) {
      if pManMR.fVerbose { Abc_FlowRetime_PrintInitStateInfo( pNtk ); }
      break;
    } else {
      if !pManMR.fGuaranteeInitState {
        fmt.Printf("WARNING: no equivalent init state. setting all initial states to don't-cares\n");
        Abc_NtkForEachLatch( pNtk, pObj, i ) Abc_LatchSetInitDc( pObj );
        break;
      }
      Abc_FlowRetime_ConstrainInit( );
    }
    
    Abc_NtkDelete(pManMR.pInitNtk);
    pManMR.pInitNtk = nil;
  
if !(true) { break }
}  }

//  assert(!pManMR->fComputeInitState || pManMR->pInitNtk);
  if pManMR.fComputeInitState {   Abc_NtkDelete(pManMR.pInitNtk); }
//  if (pManMR->fGuaranteeInitState) ; /* Abc_NtkDelete(pNtkCopy); note: original ntk deleted later */

  return pNtk;
}


/**Function*************************************************************

  Synopsis    [Pushes latch bubbles outside of box.]

  Description [If network is an AIG, a fCompl0 is allowed to remain on
               the BI node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_RemoveLatchBubbles(pLatch *Abc_Obj_t) {
  var bubble int32;
  var pNtk *Abc_Ntk_t  = pManMR.pNtk;
  var pBi, pBo, pInv *Abc_Obj_t;
      
  pBi = Abc_ObjFanin0(pLatch);
  pBo = Abc_ObjFanout0(pLatch);
  assert(!Abc_ObjIsComplement(pBi), "!Abc_ObjIsComplement(pBi)");
  assert(!Abc_ObjIsComplement(pBo), "!Abc_ObjIsComplement(pBo)");

  // push bubbles on BO into latch box
  if Abc_ObjFaninC0(pBo) && Abc_ObjFanoutNum(pBo) > 0 {
    bubble = 1;
    if Abc_LatchIsInit0(pLatch) { Abc_LatchSetInit1(pLatch);
    } else if Abc_LatchIsInit1(pLatch) { Abc_LatchSetInit0(pLatch); }  
  }

  // absorb bubbles on BI
  pBi.fCompl0 ^= bubble ^ Abc_ObjFaninC0(pLatch);

  // convert bubble to INV if not AIG
  if !Abc_NtkIsStrash( pNtk ) && Abc_ObjFaninC0(pBi) {
    pBi.fCompl0 = 0;
    pInv = Abc_NtkCreateNodeInv( pNtk, Abc_ObjFanin0(pBi) );
    Abc_ObjPatchFanin( pBi, Abc_ObjFanin0(pBi), pInv );
  }

  pBo.fCompl0 = 0;
  pLatch.fCompl0 = 0;
}


/**Function*************************************************************

  Synopsis    [Marks nodes in TFO/TFI of PI/PO.]

  Description [Sets flow data flag BLOCK appropriately.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_MarkBlocks(pNtk *Abc_Ntk_t) {
  var i int32;
  var pObj *Abc_Obj_t;

  if pManMR.fIsForward{
    // --- forward retiming : block TFO of inputs    

    // mark the frontier
    Abc_NtkForEachPo( pNtk, pObj, i )
      pObj.fMarkA = 1;
    Abc_NtkForEachLatch( pNtk, pObj, i )
      {
        pObj.fMarkA = 1;
      }
    // mark the nodes reachable from the PIs
    Abc_NtkForEachPi( pNtk, pObj, i )
      Abc_NtkMarkCone_rec( pObj, pManMR.fIsForward );
  } else {
    // --- backward retiming : block TFI of outputs

    // mark the frontier
    Abc_NtkForEachPi( pNtk, pObj, i )
      pObj.fMarkA = 1;
    Abc_NtkForEachLatch( pNtk, pObj, i )
      {
        pObj.fMarkA = 1;
      }
    // mark the nodes reachable from the POs
    Abc_NtkForEachPo( pNtk, pObj, i )
      Abc_NtkMarkCone_rec( pObj, pManMR.fIsForward );
    // block constant nodes (if enabled)
    if pManMR.fBlockConst {
      Abc_NtkForEachObj( pNtk, pObj, i )
        if (Abc_NtkIsStrash(pNtk) && Abc_AigNodeIsConst(pObj)) ||
            (!Abc_NtkIsStrash(pNtk) && Abc_NodeIsConst(pObj)) {
          FSET(pObj, BLOCK);
        }
    }
  }

  // copy marks
  Abc_NtkForEachObj( pNtk, pObj, i ) {
    if pObj.fMarkA {
      pObj.fMarkA = 0;
      if !Abc_ObjIsLatch(pObj) /* && !Abc_ObjIsPi(pObj) */  {
        FSET(pObj, BLOCK); }
    }
  }
}


/**Function*************************************************************

  Synopsis    [Computes maximum flow.]

  Description []
               
  SideEffects [Leaves VISITED flags on source-reachable nodes.]

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_PushFlows(pNtk *Abc_Ntk_t, fVerbose int32) int32 {
  var i, j, flow, last, srcDist int32;
  var pObj, pObj2 *Abc_Obj_t;
//  int clk = clock();

  pManMR.constraintMask |= BLOCK;

  pManMR.fSinkDistTerminate = 0;
  dfsfast_preorder( pNtk );

  // (i) fast max-flow computation
  for !pManMR.fSinkDistTerminate && srcDist < MAX_DIST {
    srcDist = MAX_DIST;
    Abc_NtkForEachLatch( pNtk, pObj, i )
      if (FDATA(pObj).e_dist)    
        srcDist = MIN(srcDist, int32(FDATA(pObj).e_dist));
    
    Abc_NtkForEachLatch( pNtk, pObj, i ) {
      if srcDist == int32(FDATA(pObj).e_dist) &&
          dfsfast_e( pObj, nil ) {
#ifdef DEBUG_PRINT_FLOWS
        printf("\n\n");
#endif
        flow++;
      }
    }
  }

  if fVerbose != 0 { vprintf("\t\tmax-flow1 = %d \t", flow); }

  // (ii) complete max-flow computation
  // also, marks source-reachable nodes
  for {
    last = flow;
    Abc_NtkForEachLatch( pNtk, pObj, i ) {
      if dfsplain_e( pObj, nil ) {
#ifdef DEBUG_PRINT_FLOWS
        printf("\n\n");
#endif
        flow++;
        Abc_NtkForEachObj( pNtk, pObj2, j )
          FUNSET( pObj2, VISITED );
      }
    }
  
if !(flow > last) { break }
} 
  
  if fVerbose != 0 { vprintf("max-flow2 = %d\n", flow); }

//  PRT( "time", clock() - clk );
  return flow;
}


/**Function*************************************************************

  Synopsis    [Restores latch boxes.]

  Description [Latchless BI/BO nodes are removed.  Latch boxes are 
               restored around remaining latches.]
               
  SideEffects [Deletes nodes as appropriate.]

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_FixLatchBoxes(pNtk *Abc_Ntk_t, vBoxIns *Vec_Ptr_t) {
  var i int32;
  var pObj *Abc_Obj_t;
  var pBo *Abc_Obj_t = nil;
  var pBi *Abc_Obj_t = nil;
  var vFreeBi *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
  var vFreeBo *Vec_Ptr_t  = Vec_PtrAlloc( 100 );

  // 1. remove empty bi/bo pairs
  for Vec_PtrSize( vBoxIns ) {
    pBi = (Abc_Obj_t *)Vec_PtrPop( vBoxIns );
    assert(Abc_ObjIsBi(pBi), "Abc_ObjIsBi(pBi)");
    assert(Abc_ObjFanoutNum(pBi) == 1, "Abc_ObjFanoutNum(pBi) == 1");
    // APH: broken by bias nodes assert(Abc_ObjFaninNum(pBi) == 1);
    pBo = Abc_ObjFanout0(pBi);
    assert(!Abc_ObjFaninC0(pBo), "!Abc_ObjFaninC0(pBo)");

    if Abc_ObjIsBo(pBo) {
      // an empty bi/bo pair

      Abc_ObjRemoveFanins( pBo );
      // transfer complement from BI, if present 
      assert(!Abc_ObjIsComplement(Abc_ObjFanin0(pBi)), "!Abc_ObjIsComplement(Abc_ObjFanin0(pBi))");
      Abc_ObjBetterTransferFanout( pBo, Abc_ObjFanin0(pBi), Abc_ObjFaninC0(pBi) );

      Abc_ObjRemoveFanins( pBi );
      pBi.fCompl0 = 0;
      Vec_PtrPush( vFreeBi, pBi );
      Vec_PtrPush( vFreeBo, pBo );

      // free names
      if Nm_ManFindNameById(pNtk.pManName, Abc_ObjId(pBi)) {
        Nm_ManDeleteIdName( pNtk.pManName, Abc_ObjId(pBi)); }
      if Nm_ManFindNameById(pNtk.pManName, Abc_ObjId(pBo)) {
        Nm_ManDeleteIdName( pNtk.pManName, Abc_ObjId(pBo)); }

      // check for complete detachment
      assert(Abc_ObjFaninNum(pBi) == 0, "Abc_ObjFaninNum(pBi) == 0");
      assert(Abc_ObjFanoutNum(pBi) == 0, "Abc_ObjFanoutNum(pBi) == 0");
      assert(Abc_ObjFaninNum(pBo) == 0, "Abc_ObjFaninNum(pBo) == 0");
      assert(Abc_ObjFanoutNum(pBo) == 0, "Abc_ObjFanoutNum(pBo) == 0");
    } else if Abc_ObjIsLatch(pBo) {
    } else {
      Abc_ObjPrint(os.Stdout, pBi);
      Abc_ObjPrint(os.Stdout, pBo);
      assert(false, "0");
    }
  }

  // 2. add bi/bos as necessary for latches
  Abc_NtkForEachLatch( pNtk, pObj, i ) {
    assert(Abc_ObjFaninNum(pObj) == 1, "Abc_ObjFaninNum(pObj) == 1");
    if Abc_ObjFanoutNum(pObj) {
      pBo = Abc_ObjFanout0(pObj);
    } else { pBo = nil; }
    pBi = Abc_ObjFanin0(pObj);
    
    // add BO
    if !(pBo != nil) || !Abc_ObjIsBo(pBo) {
      pBo = (Abc_Obj_t *)Vec_PtrPop( vFreeBo );
      if Abc_ObjFanoutNum(pObj) {  Abc_ObjTransferFanout( pObj, pBo ); }
      Abc_ObjAddFanin( pBo, pObj );
    }
    // add BI
    if !Abc_ObjIsBi(pBi) {
      pBi = (Abc_Obj_t *)Vec_PtrPop( vFreeBi );
      assert(Abc_ObjFaninNum(pBi) == 0, "Abc_ObjFaninNum(pBi) == 0");
      Abc_ObjAddFanin( pBi, Abc_ObjFanin0(pObj) );
      pBi.fCompl0 = pObj.fCompl0;
      Abc_ObjRemoveFanins( pObj );
      Abc_ObjAddFanin( pObj, pBi );
    }
  }

  // delete remaining BIs and BOs
  for Vec_PtrSize( vFreeBi ) {
    pObj = (Abc_Obj_t *)Vec_PtrPop( vFreeBi );
    Abc_NtkDeleteObj( pObj );
  }
  for Vec_PtrSize( vFreeBo ) {
    pObj = (Abc_Obj_t *)Vec_PtrPop( vFreeBo );
    Abc_NtkDeleteObj( pObj );
  }

#if defined(DEBUG_CHECK)
  Abc_NtkForEachObj( pNtk, pObj, i ) {
    if (Abc_ObjIsBo(pObj)) {
      assert(Abc_ObjFaninNum(pObj) == 1);
      assert(Abc_ObjIsLatch(Abc_ObjFanin0(pObj))); 
    }
    if (Abc_ObjIsBi(pObj)) {
      assert(Abc_ObjFaninNum(pObj) == 1);
      assert(Abc_ObjFanoutNum(pObj) == 1);
      assert(Abc_ObjIsLatch(Abc_ObjFanout0(pObj))); 
    }
    if (Abc_ObjIsLatch(pObj)) {
      assert(Abc_ObjFanoutNum(pObj) == 1);
      assert(Abc_ObjFaninNum(pObj) == 1);
    }
  }
#endif

  Vec_PtrFree( vFreeBi );
  Vec_PtrFree( vFreeBo );
}


/**Function*************************************************************

  Synopsis    [Checks register count along all combinational paths.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_VerifyPathLatencies(pNtk *Abc_Ntk_t) {
  var i int32;
  var pObj *Abc_Obj_t;
  fPathError = 0;

  vprintf("\t\tVerifying latency along all paths...");

  Abc_NtkForEachObj( pNtk, pObj, i ) {
    if Abc_ObjIsBo(pObj) {
      Abc_FlowRetime_VerifyPathLatencies_rec( pObj, 0 );
    } else if !pManMR.fIsForward && Abc_ObjIsPi(pObj) {
      Abc_FlowRetime_VerifyPathLatencies_rec( pObj, 0 );
    }

    if fPathError != 0 {
      if Abc_ObjFaninNum(pObj) > 0 {
        fmt.Printf("fanin ");
        print_node(Abc_ObjFanin0(pObj));
      }
      fmt.Printf("\n");
      os.Exit(int(0));
    }
  }

  vprintf(" ok\n");

  Abc_NtkForEachObj( pNtk, pObj, i ) {
    pObj.fMarkA = 0;
    pObj.fMarkB = 0;
    pObj.fMarkC = 0;
  }
}


func Abc_FlowRetime_VerifyPathLatencies_rec(pObj *Abc_Obj_t, markD int32) int32 {
  var i, j int32;
  var pNext *Abc_Obj_t;
  var fCare int32;
  var markC int32  = pObj.fMarkC;

  if !pObj.fMarkB {
    pObj.fMarkB = 1; // visited
    
    if Abc_ObjIsLatch(pObj) {
      markC = 1; }      // latch in output
    
    if !pManMR.fIsForward && !Abc_ObjIsPo(pObj) && !Abc_ObjFanoutNum(pObj) {
      return -1; } // dangling non-PO outputs : don't care what happens
    
    Abc_ObjForEachFanout( pObj, pNext, i ) {
      // reached end of cycle?
      if Abc_ObjIsBo(pNext) ||
           (pManMR.fIsForward && Abc_ObjIsPo(pNext)) {
        if !(markD != 0) && !Abc_ObjIsLatch(pObj) {
          fmt.Printf("\nERROR: no-latch path (end)\n");
          print_node(pNext);
          fmt.Printf("\n");
          fPathError = 1;
        }
      } else if !pManMR.fIsForward && Abc_ObjIsPo(pNext) {
        if (markD != 0) || Abc_ObjIsLatch(pObj) {
          fmt.Printf("\nERROR: extra-latch path to outputs\n");
          print_node(pNext);
          fmt.Printf("\n");
          fPathError = 1;
        }
      } else {
        j = Abc_FlowRetime_VerifyPathLatencies_rec( pNext, markD || Abc_ObjIsLatch(pObj) );
        if j >= 0 {
          markC |= j;
          fCare = 1;
        }
      }

      if fPathError != 0 {
        print_node(pObj);
        fmt.Printf("\n");
        return 0;
      }
    }
  }

  if !(fCare != 0) { return -1; }

  if (markC != 0) && (markD != 0) {
    fmt.Printf("\nERROR: mult-latch path\n");
    print_node(pObj);
    fmt.Printf("\n");
    fPathError = 1;
  }
  if !(markC != 0) && !(markD != 0) {
    fmt.Printf("\nERROR: no-latch path (inter)\n");
    print_node(pObj);
    fmt.Printf("\n");
    fPathError = 1;
  }

  return (pObj.fMarkC = markC);
}


/**Function*************************************************************

  Synopsis    [Copies initial state from latches to BO nodes.]

  Description [Initial states are marked on BO nodes with INIT_0 and
               INIT_1 flags in their Flow_Data structures.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_CopyInitState(pSrc *Abc_Obj_t, pDest *Abc_Obj_t) {
  var pObj *Abc_Obj_t;

  if !pManMR.fComputeInitState { return; }

  assert(Abc_ObjIsLatch(pSrc), "Abc_ObjIsLatch(pSrc)");
  assert(Abc_ObjFanin0(pDest) == pSrc, "Abc_ObjFanin0(pDest) == pSrc");
  assert(!Abc_ObjFaninC0(pDest), "!Abc_ObjFaninC0(pDest)");

  FUNSET(pDest, INIT_CARE);
  if Abc_LatchIsInit0(pSrc) {
    FSET(pDest, INIT_0);
  } else if Abc_LatchIsInit1(pSrc) {
    FSET(pDest, INIT_1);  
  }
  
  if !pManMR.fIsForward {
    pObj = (Abc_Obj_t*)Abc_ObjData(pSrc);
    assert(Abc_ObjIsPi(pObj), "Abc_ObjIsPi(pObj)");
    FDATA(pDest).pInitObj = pObj;
  }
}


/**Function*************************************************************

  Synopsis    [Implements min-cut.]

  Description [Requires source-reachable nodes to be marked VISITED.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_ImplementCut(pNtk *Abc_Ntk_t) int32 {
  var i, j, cut, unmoved int32;
  var pObj *Abc_Obj_t;
  var pReg *Abc_Obj_t;
  var pNext *Abc_Obj_t;
  var pBo *Abc_Obj_t = nil;
  var pBi *Abc_Obj_t = nil;
  var vFreeRegs *Vec_Ptr_t  = Vec_PtrAlloc( Abc_NtkLatchNum(pNtk) );
  var vBoxIns *Vec_Ptr_t  = Vec_PtrAlloc( Abc_NtkLatchNum(pNtk) );
  var vMove *Vec_Ptr_t  = Vec_PtrAlloc( 100 );

  // remove latches from netlist
  Abc_NtkForEachLatch( pNtk, pObj, i ) {
    pBo = Abc_ObjFanout0(pObj);
    pBi = Abc_ObjFanin0(pObj);
    assert(Abc_ObjIsBo(pBo) && Abc_ObjIsBi(pBi), "Abc_ObjIsBo(pBo) && Abc_ObjIsBi(pBi)");
    Vec_PtrPush( vBoxIns, pBi );

    // copy initial state values to BO
    Abc_FlowRetime_CopyInitState( pObj, pBo );

    // re-use latch elsewhere
    Vec_PtrPush( vFreeRegs, pObj );
    FSET(pBo, CROSS_BOUNDARY);

    // cut out of netlist
    Abc_ObjPatchFanin( pBo, pObj, pBi );
    Abc_ObjRemoveFanins( pObj );

    // free name
    if Nm_ManFindNameById(pNtk.pManName, Abc_ObjId(pObj)) {
      Nm_ManDeleteIdName( pNtk.pManName, Abc_ObjId(pObj)); }
  }

  // insert latches into netlist
  Abc_NtkForEachObj( pNtk, pObj, i ) {
    if Abc_ObjIsLatch( pObj ) { continue; }
    if FTEST(pObj, BIAS_NODE) { continue; }
    
    // a latch is required on every node that lies across the min-cit
    assert(!pManMR.fIsForward || !FTEST(pObj, VISITED_E) || FTEST(pObj, VISITED_R), "!pManMR.fIsForward || !FTEST(pObj, VISITED_E) || FTEST(pObj, VISITED_R)");
    if FTEST(pObj, VISITED_R) && !FTEST(pObj, VISITED_E) {
      assert(FTEST(pObj, FLOW), "FTEST(pObj, FLOW)");

      // count size of cut
      cut++;
      if (pManMR.fIsForward && Abc_ObjIsBo(pObj)) || 
          (!pManMR.fIsForward && Abc_ObjIsBi(pObj)) {
        unmoved++; }
      
      // only insert latch between fanouts that lie across min-cut
      // some fanout paths may be cut at deeper points
      Abc_ObjForEachFanout( pObj, pNext, j )
        if (Abc_FlowRetime_IsAcrossCut( pObj, pNext ))
          Vec_PtrPush(vMove, pNext);

      // check that move-set is non-zero
      if Vec_PtrSize(vMove) == 0 {
        print_node(pObj); }
      assert(Vec_PtrSize(vMove) > 0, "Vec_PtrSize(vMove) > 0");
      
      // insert one of re-useable registers
      assert(Vec_PtrSize( vFreeRegs ), "Vec_PtrSize( vFreeRegs )");
      pReg = (Abc_Obj_t *)Vec_PtrPop( vFreeRegs );
      
      Abc_ObjAddFanin(pReg, pObj);
      for Vec_PtrSize( vMove ) {
        pNext = (Abc_Obj_t *)Vec_PtrPop( vMove );
        Abc_ObjPatchFanin( pNext, pObj, pReg );
        if Abc_ObjIsBi(pNext) { assert(Abc_ObjFaninNum(pNext) == 1, "Abc_ObjFaninNum(pNext) == 1"); }

      }
      // APH: broken by bias nodes  if (Abc_ObjIsBi(pObj)) assert(Abc_ObjFaninNum(pObj) == 1);
    }
  }

#if defined(DEBUG_CHECK)        
  Abc_FlowRetime_VerifyPathLatencies( pNtk );
#endif

  // delete remaining latches
  for Vec_PtrSize( vFreeRegs ) {
    pReg = (Abc_Obj_t *)Vec_PtrPop( vFreeRegs );
    Abc_NtkDeleteObj( pReg );
  }
  
  // update initial states
  Abc_FlowRetime_UpdateLags( );
  Abc_FlowRetime_InitState( pNtk );

  // restore latch boxes
  Abc_FlowRetime_FixLatchBoxes( pNtk, vBoxIns );

  Vec_PtrFree( vFreeRegs );
  Vec_PtrFree( vMove );
  Vec_PtrFree( vBoxIns );

  vprintf("\t\tmin-cut = %d (unmoved = %d)\n", cut, unmoved);
  return cut;
}


/**Function*************************************************************

  Synopsis    [Adds dummy fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_AddDummyFanin(pObj *Abc_Obj_t) {
  var pNtk *Abc_Ntk_t  = Abc_ObjNtk( pObj );

  if Abc_NtkIsStrash(pNtk) { 
    Abc_ObjAddFanin(pObj, Abc_AigConst1( pNtk ));
  } else {
    Abc_ObjAddFanin(pObj, Abc_NtkCreateNodeConst0( pNtk )); }
}


/**Function*************************************************************

  Synopsis    [Prints information about a node.]

  Description [Debuging.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func print_node(pObj *Abc_Obj_t) {
  var i int32;
  var pNext *Abc_Obj_t;
  var m [6]byte;

  m[0] = 0;
  if pObj.fMarkA {
    strcat(m, "A"); }
  if pObj.fMarkB {
    strcat(m, "B"); }
  if pObj.fMarkC {
    strcat(m, "C"); }

  printf("node %d type=%d lev=%d tedge=%d (%x%s) fanouts {", Abc_ObjId(pObj), Abc_ObjType(pObj),
           pObj.Level, Vec_PtrSize(FTIMEEDGES(pObj)), FDATA(pObj).mark, m);
  Abc_ObjForEachFanout( pObj, pNext, i )
    fmt.Printf("%d[%d](%d),", Abc_ObjId(pNext), Abc_ObjType(pNext), FDATA(pNext).mark);
  fmt.Printf("} fanins {");
  Abc_ObjForEachFanin( pObj, pNext, i )
    fmt.Printf("%d[%d](%d),", Abc_ObjId(pNext), Abc_ObjType(pNext), FDATA(pNext).mark);
  fmt.Printf("}\n");
}

func print_node2(pObj *Abc_Obj_t) {
  var i int32;
  var pNext *Abc_Obj_t;
  var m [6]byte;

  m[0] = 0;
  if pObj.fMarkA {
    strcat(m, "A"); }
  if pObj.fMarkB {
    strcat(m, "B"); }
  if pObj.fMarkC {
    strcat(m, "C"); }

  printf("node %d type=%d %s fanouts {", Abc_ObjId(pObj), Abc_ObjType(pObj), m);
  Abc_ObjForEachFanout( pObj, pNext, i )
    fmt.Printf("%d ,", Abc_ObjId(pNext));
  fmt.Printf("} fanins {");
  Abc_ObjForEachFanin( pObj, pNext, i )
    fmt.Printf("%d ,", Abc_ObjId(pNext));
  fmt.Printf("} ");
}

func print_node3(pObj *Abc_Obj_t) {
  var i int32;
  var pNext *Abc_Obj_t;
  var m [6]byte;

  m[0] = 0;
  if pObj.fMarkA {
    strcat(m, "A"); }
  if pObj.fMarkB {
    strcat(m, "B"); }
  if pObj.fMarkC {
    strcat(m, "C"); }

  printf("\nnode %d type=%d mark=%d %s\n", Abc_ObjId(pObj), Abc_ObjType(pObj), FDATA(pObj).mark, m);
  fmt.Printf("fanouts\n");
  Abc_ObjForEachFanout( pObj, pNext, i ) {
    print_node(pNext);
    fmt.Printf("\n");
  }
  fmt.Printf("fanins\n");
  Abc_ObjForEachFanin( pObj, pNext, i ) {
    print_node(pNext);
    fmt.Printf("\n");
  }
}


/**Function*************************************************************

  Synopsis    [Transfers fanout.]

  Description [Does not produce an error if there is no fanout.
               Complements as necessary.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjBetterTransferFanout(pFrom *Abc_Obj_t, pTo *Abc_Obj_t, complement int32) {
  var pNext *Abc_Obj_t;
  
  for Abc_ObjFanoutNum(pFrom) > 0 {
    pNext = Abc_ObjFanout0(pFrom);
    Abc_ObjPatchFanin( pNext, pFrom, Abc_ObjNotCond(pTo, complement) );
  }
}


/**Function*************************************************************

  Synopsis    [Returns true is a connection spans the min-cut.]

  Description [pNext is a direct fanout of pObj.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_IsAcrossCut(pObj *Abc_Obj_t, pNext *Abc_Obj_t) int32 {

  if FTEST(pObj, VISITED_R) && !FTEST(pObj, VISITED_E) {
    if pManMR.fIsForward {
      if !FTEST(pNext, VISITED_R) || 
          (FTEST(pNext, BLOCK_OR_CONS) & pManMR.constraintMask)|| 
          FTEST(pNext, CROSS_BOUNDARY) || 
          Abc_ObjIsLatch(pNext) {
        return 1; }
    } else {
      if FTEST(pNext, VISITED_E) ||
          FTEST(pNext, CROSS_BOUNDARY) {
        return 1; }
    }
  }
  
  return 0;
}


/**Function*************************************************************

  Synopsis    [Resets flow problem]

  Description [If fClearAll is true, all marks will be cleared; this is
               typically appropriate after the circuit structure has
               been modified.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_ClearFlows(fClearAll int32) {
	if pManMR == nil {
		return
	}
	n := int(pManMR.nNodes)
	if n > len(pManMR.pDataArray) {
		n = len(pManMR.pDataArray)
	}
	if fClearAll != 0 {
		clear(pManMR.pDataArray[:n])
		return
	}
	// Keep the init-state mapping and all flags outside the flow search state.
	for i := 0; i < n; i++ {
		data := &pManMR.pDataArray[i]
		data.mark &^= VISITED | FLOW
		data.e_dist = 0
		data.r_dist = 0
		data.pred = nil
	}
}


/**Function*************************************************************

  Synopsis    [Duplicates network.]

  Description [Duplicates any type of network. Preserves copy data.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_NtkDup(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
  var pNtkCopy *Abc_Ntk_t;
  var pObj, pObjCopy, pNext, pNextCopy *Abc_Obj_t;
  var i, j int32;

  pNtkCopy = Abc_NtkAlloc( pNtk.ntkType, pNtk.ntkFunc, 1 );
  pNtkCopy.pName = Extra_UtilStrsav(pNtk.pName);
  pNtkCopy.pSpec = Extra_UtilStrsav(pNtk.pSpec);

  // copy each object
  Abc_NtkForEachObj( pNtk, pObj, i) {

    if Abc_NtkIsStrash( pNtk ) && Abc_AigNodeIsConst( pObj ) {
      pObjCopy = Abc_AigConst1( pNtkCopy );
    } else {
      pObjCopy = Abc_NtkDupObj( pNtkCopy, pObj, 0 ); }

    FDATA( pObj ).pCopy = pObjCopy;
    FDATA( pObj ).mark = 0;

    // assert( pManMR->fIsForward || pObj->Id == pObjCopy->Id );

    // copy complementation
    pObjCopy.fCompl0 = pObj.fCompl0;
    pObjCopy.fCompl1 = pObj.fCompl1;
    pObjCopy.fPhase = pObj.fPhase;
  }

  // connect fanin
  Abc_NtkForEachObj( pNtk, pObj, i) {
    pObjCopy = FDATA(pObj).pCopy;
    assert(pObjCopy, "pObjCopy");
    Abc_ObjForEachFanin( pObj, pNext, j ) {
      pNextCopy = FDATA(pNext).pCopy;
      assert((pNextCopy != nil), "pNextCopy");
      assert(pNext.Type == pNextCopy.Type, "pNext.Type == pNextCopy.Type");

      Abc_ObjAddFanin(pObjCopy, pNextCopy);
    }
  }

                             
  Abc_NtkForEachObj( pNtk, pObj, i) { 
    pObjCopy = FDATA(pObj).pCopy;
    assert(Abc_ObjFanoutNum( pObj ) ==  Abc_ObjFanoutNum( pObjCopy ), "Abc_ObjFanoutNum( pObj ) ==  Abc_ObjFanoutNum( pObjCopy )");
    assert(Abc_ObjFaninNum( pObj ) ==  Abc_ObjFaninNum( pObjCopy ), "Abc_ObjFaninNum( pObj ) ==  Abc_ObjFaninNum( pObjCopy )");
  }
      

  assert(Abc_NtkObjNum( pNtk )   == Abc_NtkObjNum( pNtkCopy ), "Abc_NtkObjNum( pNtk )   == Abc_NtkObjNum( pNtkCopy )");
  assert(Abc_NtkLatchNum( pNtk ) == Abc_NtkLatchNum( pNtkCopy ), "Abc_NtkLatchNum( pNtk ) == Abc_NtkLatchNum( pNtkCopy )");
  assert(Abc_NtkPoNum( pNtk )    == Abc_NtkPoNum( pNtkCopy ), "Abc_NtkPoNum( pNtk )    == Abc_NtkPoNum( pNtkCopy )");
  assert(Abc_NtkPiNum( pNtk )    == Abc_NtkPiNum( pNtkCopy ), "Abc_NtkPiNum( pNtk )    == Abc_NtkPiNum( pNtkCopy )");

  return pNtkCopy;
}


/**Function*************************************************************

  Synopsis    [Silent restrash.]

  Description [Same functionality as Abc_NtkRestrash but w/o warnings.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_NtkSilentRestrash(pNtk *Abc_Ntk_t, fCleanup int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, nNodes int32;//, RetValue;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // start the new network (constants and CIs of the old network will point to the their counterparts in the new network)
    pNtkAig = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );
    // restrash the nodes (assuming a topological order of the old network)
    Abc_NtkForEachNode( pNtk, pObj, i )
        pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkAig.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) );
    // finalize the network
    Abc_NtkFinalize( pNtk, pNtkAig );
    // perform cleanup if requested
    if fCleanup != 0 {
      nNodes = Abc_AigCleanup((Abc_Aig_t *)pNtkAig.pManFunc); }
    // duplicate EXDC 
    if pNtk.pExdc {
      pNtkAig.pExdc = Abc_NtkDup( pNtk.pExdc ); }
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkAig ) {
        fmt.Printf("Abc_NtkStrash: The network check has failed.\n");
        Abc_NtkDelete( pNtkAig );
        return nil;
      }
    return pNtkAig;
}



/**Function*************************************************************

  Synopsis    [Updates lag values.]

  Description [Recursive.  Forward retiming.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_UpdateLags_forw_rec(pObj *Abc_Obj_t) {
  var pNext *Abc_Obj_t;
  var i int32;

  assert(!Abc_ObjIsPi(pObj), "!Abc_ObjIsPi(pObj)");
  assert(!Abc_ObjIsLatch(pObj), "!Abc_ObjIsLatch(pObj)");

  if Abc_ObjIsBo(pObj) { return; }
  if Abc_NodeIsTravIdCurrent(pObj) { return; }

  Abc_NodeSetTravIdCurrent(pObj);

  if Abc_ObjIsNode(pObj) {
    Abc_FlowRetime_SetLag( pObj, -1+Abc_FlowRetime_GetLag(pObj) );
  }

  Abc_ObjForEachFanin( pObj, pNext, i ) {
    Abc_FlowRetime_UpdateLags_forw_rec( pNext );
  }
}


/**Function*************************************************************

  Synopsis    [Updates lag values.]

  Description [Recursive.  Backward retiming.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_UpdateLags_back_rec(pObj *Abc_Obj_t) {
  var pNext *Abc_Obj_t;
  var i int32;

  assert(!Abc_ObjIsPo(pObj), "!Abc_ObjIsPo(pObj)");
  assert(!Abc_ObjIsLatch(pObj), "!Abc_ObjIsLatch(pObj)");

  if Abc_ObjIsBo(pObj) { return; }
  if Abc_NodeIsTravIdCurrent(pObj) { return; }

  Abc_NodeSetTravIdCurrent(pObj);

  if Abc_ObjIsNode(pObj) {
    Abc_FlowRetime_SetLag( pObj, 1+Abc_FlowRetime_GetLag(pObj) );
  }

  Abc_ObjForEachFanout( pObj, pNext, i ) {
    Abc_FlowRetime_UpdateLags_back_rec( pNext );
  }
}

/**Function*************************************************************

  Synopsis    [Updates lag values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_UpdateLags() {
  var pObj, pNext *Abc_Obj_t;
  var i, j int32;

  Abc_NtkIncrementTravId( pManMR.pNtk );

  Abc_NtkForEachLatch( pManMR.pNtk, pObj, i )
    if pManMR.fIsForward {
      Abc_ObjForEachFanin( pObj, pNext, j )          
        Abc_FlowRetime_UpdateLags_forw_rec( pNext );
    } else {
      Abc_ObjForEachFanout( pObj, pNext, j )          
        Abc_FlowRetime_UpdateLags_back_rec( pNext );
    }
}


/**Function*************************************************************

  Synopsis    [Gets lag value of a node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_GetLag(pObj *Abc_Obj_t) int32 {
  assert(!Abc_ObjIsLatch(pObj), "!Abc_ObjIsLatch(pObj)");
  assert(int32(Abc_ObjId(pObj)) < Vec_IntSize(pManMR.vLags), "int32(Abc_ObjId(pObj)) < Vec_IntSize(pManMR.vLags)");

  return Vec_IntEntry(pManMR.vLags, Abc_ObjId(pObj));
}

/**Function*************************************************************

  Synopsis    [Sets lag value of a node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FlowRetime_SetLag(pObj *Abc_Obj_t, lag int32) {
  assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
  assert(int32(Abc_ObjId(pObj)) < Vec_IntSize(pManMR.vLags), "int32(Abc_ObjId(pObj)) < Vec_IntSize(pManMR.vLags)");

  Vec_IntWriteEntry(pManMR.vLags, Abc_ObjId(pObj), lag);
}


func Abc_ObjPrintNeighborhood_rec(pObj *Abc_Obj_t, vNodes *Vec_Ptr_t, depth int32) {
  var pObj2 *Abc_Obj_t;
  var i int32;
  
  if pObj.fMarkC || depth < 0 { return; }

  pObj.fMarkC = 1;
  Vec_PtrPush( vNodes, pObj );

  Abc_ObjPrint( os.Stdout, pObj );

  Abc_ObjForEachFanout(pObj, pObj2, i) {
    Abc_ObjPrintNeighborhood_rec( pObj2, vNodes, depth-1 );
  }
  Abc_ObjForEachFanin(pObj, pObj2, i) {
    Abc_ObjPrintNeighborhood_rec( pObj2, vNodes, depth-1 );
  }
}

func Abc_ObjPrintNeighborhood(pObj *Abc_Obj_t, depth int32) {
  var vNodes *Vec_Ptr_t  = Vec_PtrAlloc(100); 
  var pObj2 *Abc_Obj_t;

  Abc_ObjPrintNeighborhood_rec( pObj, vNodes, depth );

  for Vec_PtrSize(vNodes) {
    pObj2 = (Abc_Obj_t*)Vec_PtrPop(vNodes);
    pObj2.fMarkC = 0;
  }

  Vec_PtrFree(vNodes);
}
//ABC_NAMESPACE_IMPL_END

