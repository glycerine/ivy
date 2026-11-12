package abci

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [abcRr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Redundancy removal.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcRr.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "proof/fraig/fraig.h"
//#include "opt/sim/sim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Abc_RRMan_t struct {
    // the parameters
    pNtk       *Abc_Ntk_t;             // the network
    nFaninLevels              int32;     // the number of fanin levels
    nFanoutLevels              int32;    // the number of fanout levels
    // the node/fanin/fanout
    pNode       *Abc_Obj_t;            // the node
    pFanin       *Abc_Obj_t;           // the fanin
    pFanout       *Abc_Obj_t;          // the fanout
    // the intermediate cones
    vFaninLeaves       *Vec_Ptr_t;     // the leaves of the fanin cone
    vFanoutRoots       *Vec_Ptr_t;     // the roots of the fanout cone
    // the window
    vLeaves       *Vec_Ptr_t;          // the leaves of the window
    vCone       *Vec_Ptr_t;            // the internal nodes of the window
    vRoots       *Vec_Ptr_t;           // the roots of the window
    pWnd       *Abc_Ntk_t;             // the window derived for the edge
    // the miter 
    pMiter       *Abc_Ntk_t;           // the miter derived from the window
    pParams  *Prove_Params_t;          // the miter proving parameters
    // statistical variables
    nNodesOld              int32;        // the old number of nodes
    nLevelsOld              int32;       // the old number of levels
    nEdgesTried              int32;      // the number of nodes tried
    nEdgesRemoved              int32;    // the number of nodes proved
    timeWindow          abctime;       // the time to construct the window
    timeMiter          abctime;        // the time to construct the miter
    timeProve          abctime;        // the time to prove the miter
    timeUpdate          abctime;       // the network update time
    timeTotal          abctime;        // the total runtime
}

// static Abc_RRMan_t * Abc_RRManStart();

// static void          Abc_RRManStop( Abc_RRMan_t * p );

// static void          Abc_RRManPrintStats( Abc_RRMan_t * p );

// static void          Abc_RRManClean( Abc_RRMan_t * p );

// static int           Abc_NtkRRProve( Abc_RRMan_t * p );

// static int           Abc_NtkRRUpdate( Abc_Ntk_t * pNtk, Abc_Obj_t * pNode, Abc_Obj_t * pFanin, Abc_Obj_t * pFanout );

// static int           Abc_NtkRRWindow( Abc_RRMan_t * p );


// static int           Abc_NtkRRTfi_int( Vec_Ptr_t * vLeaves, int LevelLimit );

// static int           Abc_NtkRRTfo_int( Vec_Ptr_t * vLeaves, Vec_Ptr_t * vRoots, int LevelLimit, Abc_Obj_t * pEdgeFanin, Abc_Obj_t * pEdgeFanout );

// static int           Abc_NtkRRTfo_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vRoots, int LevelLimit );

// static void          Abc_NtkRRTfi_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vCone, int LevelLimit );

// static Abc_Ntk_t *   Abc_NtkWindow( Abc_Ntk_t * pNtk, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vCone, Vec_Ptr_t * vRoots );


// static void          Abc_NtkRRSimulateStart( Abc_Ntk_t * pNtk );

// static void          Abc_NtkRRSimulateStop( Abc_Ntk_t * pNtk );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Removes stuck-at redundancies.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRR(pNtk *Abc_Ntk_t, nFaninLevels int32, nFanoutLevels int32, fUseFanouts int32, fVerbose int32) int32 {
    var pProgress *ProgressBar;
    var p *Abc_RRMan_t;
    var pNode, pFanin, pFanout *Abc_Obj_t;
    var i, k, m, nNodes, RetValue int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    // start the manager
    p = Abc_RRManStart();
    p.pNtk          = pNtk;
    p.nFaninLevels  = nFaninLevels;
    p.nFanoutLevels = nFanoutLevels;
    p.nNodesOld     = Abc_NtkNodeNum(pNtk);
    p.nLevelsOld    = Abc_AigLevel(pNtk);
    // remember latch values
//    Abc_NtkForEachLatch( pNtk, pNode, i )
//        pNode->pNext = pNode->pData;
    // go through the nodes
    Abc_NtkCleanCopy(pNtk);
    nNodes = Abc_NtkObjNumMax(pNtk);
    Abc_NtkRRSimulateStart(pNtk);
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        // stop if all nodes have been tried once
        if i >= nNodes {
            break; }
        // skip the constant node
//        if ( Abc_NodeIsConst(pNode) )
//            continue;
        // skip persistant nodes
        if Abc_NodeIsPersistant(pNode) {
            continue; }
        // skip the nodes with many fanouts
        if Abc_ObjFanoutNum(pNode) > 1000 {
            continue; }
        // construct the window
        if !(fUseFanouts != 0) {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                // skip the nodes with only one fanout (tree nodes)
                if Abc_ObjFanoutNum(pFanin) == 1 {
                    continue; }
/*
                if ( pFanin->Id == 228 && pNode->Id == 2649 )
                {
                    int k = 0;
                }
*/
                p.nEdgesTried++;
                Abc_RRManClean( p );
                p.pNode   = pNode;
                p.pFanin  = pFanin;
                p.pFanout = nil;

                clk = Abc_Clock();
                RetValue = Abc_NtkRRWindow( p );
                p.timeWindow += Abc_Clock() - clk;
                if !(RetValue != 0) {
                    continue; }
/*
                if ( pFanin->Id == 228 && pNode->Id == 2649 )
                {
                    Abc_NtkShowAig( p->pWnd, 0 );
                }
*/
                clk = Abc_Clock();
                RetValue = Abc_NtkRRProve( p );
                p.timeMiter += Abc_Clock() - clk;
                if !(RetValue != 0) {
                    continue; }
//printf( "%d -> %d (%d)\n", pFanin->Id, pNode->Id, k );

                clk = Abc_Clock();
                Abc_NtkRRUpdate( pNtk, p.pNode, p.pFanin, p.pFanout );
                p.timeUpdate += Abc_Clock() - clk;

                p.nEdgesRemoved++;
                break;
            }
            continue;
        }
        // use the fanouts
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
        // inlining of macro Abc_ObjForEachFanout
        for  m = 0; (m < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, m)), 1); m++  {
            // skip the nodes with only one fanout (tree nodes)
//            if ( Abc_ObjFanoutNum(pFanin) == 1 && Abc_ObjFanoutNum(pNode) == 1 )
//                continue;

            p.nEdgesTried++;
            Abc_RRManClean( p );
            p.pNode   = pNode;
            p.pFanin  = pFanin;
            p.pFanout = pFanout;

            clk = Abc_Clock();
            RetValue = Abc_NtkRRWindow( p );
            p.timeWindow += Abc_Clock() - clk;
            if !(RetValue != 0) {
                continue; }

            clk = Abc_Clock();
            RetValue = Abc_NtkRRProve( p );
            p.timeMiter += Abc_Clock() - clk;
            if !(RetValue != 0) {
                continue; }

            clk = Abc_Clock();
            Abc_NtkRRUpdate( pNtk, p.pNode, p.pFanin, p.pFanout );
            p.timeUpdate += Abc_Clock() - clk;

            p.nEdgesRemoved++;
            break;
        }
        }
    }
    }
    Abc_NtkRRSimulateStop(pNtk);
    Extra_ProgressBarStop( pProgress );
    p.timeTotal = Abc_Clock() - clkTotal;
    if fVerbose != 0 {
        Abc_RRManPrintStats( p ); }
    Abc_RRManStop( p );
    // restore latch values
//    Abc_NtkForEachLatch( pNtk, pNode, i )
//        pNode->pData = pNode->pNext, pNode->pNext = NULL;
    // put the nodes into the DFS order and reassign their IDs
    Abc_NtkReassignIds( pNtk );
    Abc_NtkLevel( pNtk );
    // check
    if !Abc_NtkCheck( pNtk ) {
        fmt.Printf("Abc_NtkRR: The network check has failed.\n");
        return 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Start the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_RRManStart() *Abc_RRMan_t {
    var p *Abc_RRMan_t;
    // inlining of macro ABC_ALLOC
    p = ((Abc_RRMan_t *) malloc(sizeof(Abc_RRMan_t) * (size_t)(1)));
    *p = Abc_RRMan_t{};
    p.vFaninLeaves  = Vec_PtrAlloc( 100 );  // the leaves of the fanin cone
    p.vFanoutRoots  = Vec_PtrAlloc( 100 );  // the roots of the fanout cone
    p.vLeaves       = Vec_PtrAlloc( 100 );  // the leaves of the window
    p.vCone         = Vec_PtrAlloc( 100 );  // the internal nodes of the window
    p.vRoots        = Vec_PtrAlloc( 100 );  // the roots of the window
    // inlining of macro ABC_ALLOC
    p.pParams       = ((Prove_Params_t *) malloc(sizeof(Prove_Params_t) * (size_t)(1)));
    memset( p.pParams, 0, sizeof(Prove_Params_t) );
    Prove_ParamsSetDefault( p.pParams );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stop the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_RRManStop(p *Abc_RRMan_t) {
    Abc_RRManClean( p );
    Vec_PtrFree( p.vFaninLeaves  );  
    Vec_PtrFree( p.vFanoutRoots  );  
    Vec_PtrFree( p.vLeaves );  
    Vec_PtrFree( p.vCone );  
    Vec_PtrFree( p.vRoots );  
    // inlining of macro ABC_FREE
    ((p.pParams) ? (free((char *) (p.pParams)), (p.pParams) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Stop the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_RRManPrintStats(p *Abc_RRMan_t) {
    var Ratio float64  = 100.0*(p.nNodesOld  - Abc_NtkNodeNum(p.pNtk))/p.nNodesOld;
    fmt.Printf("Redundancy removal statistics:\n");
    fmt.Printf("Edges tried     = %6d.\n", p.nEdgesTried);
    fmt.Printf("Edges removed   = %6d. (%5.2f %%)\n", p.nEdgesRemoved, 100.0*p.nEdgesRemoved/p.nEdgesTried);
    fmt.Printf("Node gain       = %6d. (%5.2f %%)\n", p.nNodesOld  - Abc_NtkNodeNum(p.pNtk), Ratio);
    fmt.Printf("Level gain      = %6d.\n", p.nLevelsOld - Abc_AigLevel(p.pNtk));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Windowing      "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeWindow)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Miter          "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMiter)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("    Construct  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMiter - p.timeProve)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("    Prove      "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeProve)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Update         "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeUpdate)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("TOTAL          "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Clean the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_RRManClean(p *Abc_RRMan_t) {
    p.pNode   = nil; 
    p.pFanin  = nil; 
    p.pFanout = nil; 
    Vec_PtrClear( p.vFaninLeaves  );  
    Vec_PtrClear( p.vFanoutRoots  );  
    Vec_PtrClear( p.vLeaves );  
    Vec_PtrClear( p.vCone );  
    Vec_PtrClear( p.vRoots );  
    if p.pWnd {   Abc_NtkDelete( p.pWnd ); }
    if p.pMiter { Abc_NtkDelete( p.pMiter ); }
    p.pWnd   = nil;
    p.pMiter = nil;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the miter is constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRProve(p *Abc_RRMan_t) int32 {
    var pWndCopy *Abc_Ntk_t;
    var RetValue int32;
    var clk abctime;
//    Abc_NtkShowAig( p->pWnd, 0 );
    pWndCopy = Abc_NtkDup( p.pWnd );
    Abc_NtkRRUpdate( pWndCopy, p.pNode.pCopy.pCopy, p.pFanin.pCopy.pCopy, p.pFanout? p.pFanout.pCopy.pCopy : nil );
    if !Abc_NtkIsDfsOrdered(pWndCopy) {
        Abc_NtkReassignIds(pWndCopy); }
    p.pMiter = Abc_NtkMiter( p.pWnd, pWndCopy, 1, 0, 0, 0 );
    Abc_NtkDelete( pWndCopy );
clk = Abc_Clock();
    RetValue  = Abc_NtkMiterProve( &p.pMiter, p.pParams );
p.timeProve += Abc_Clock() - clk;
    if RetValue == 1 {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Updates the network after redundancy removal.]

  Description [This procedure assumes that non-control value of the fanin
  was proved redundant. It is okay to concentrate on non-control values
  because the control values can be seen as redundancy of the fanout edge.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRUpdate(pNtk *Abc_Ntk_t, pNode *Abc_Obj_t, pFanin *Abc_Obj_t, pFanout *Abc_Obj_t) int32 {
    var pNodeNew *Abc_Obj_t = nil;
    var pFanoutNew *Abc_Obj_t = nil;
    assert(pFanout == nil, "pFanout == nil");
    assert(!Abc_ObjIsComplement(pNode), "!Abc_ObjIsComplement(pNode)");
    assert(!Abc_ObjIsComplement(pFanin), "!Abc_ObjIsComplement(pFanin)");
    assert(!Abc_ObjIsComplement(pFanout), "!Abc_ObjIsComplement(pFanout)");
    // find the node after redundancy removal
    if pFanin == Abc_ObjFanin0(pNode) {
        pNodeNew = Abc_ObjChild1(pNode);
    } else if pFanin == Abc_ObjFanin1(pNode) {
        pNodeNew = Abc_ObjChild0(pNode);
    } else { assert(false, "0"); }
    // replace
    if pFanout == nil {
        Abc_AigReplace( (Abc_Aig_t *)pNtk.pManFunc, pNode, pNodeNew, 1 );
        return 1;
    }
    // find the fanout after redundancy removal
    if pNode == Abc_ObjFanin0(pFanout) {
        pFanoutNew = Abc_AigAnd( (Abc_Aig_t *)pNtk.pManFunc, Abc_ObjNotCond(pNodeNew,Abc_ObjFaninC0(pFanout)), Abc_ObjChild1(pFanout) );
    } else if pNode == Abc_ObjFanin1(pFanout) {
        pFanoutNew = Abc_AigAnd( (Abc_Aig_t *)pNtk.pManFunc, Abc_ObjNotCond(pNodeNew,Abc_ObjFaninC1(pFanout)), Abc_ObjChild0(pFanout) );
    } else { assert(false, "0"); }
    // replace
    Abc_AigReplace( (Abc_Aig_t *)pNtk.pManFunc, pFanout, pFanoutNew, 1 );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Constructs window for checking RR.]

  Description [If the window (p->pWnd) with the given scope (p->nFaninLevels, 
  p->nFanoutLevels) cannot be constructed, returns 0. Otherwise, returns 1.
  The levels are measured from the fanin node (pFanin) and the fanout node
  (pEdgeFanout), respectively.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRWindow(p *Abc_RRMan_t) int32 {
    var pObj, pEdgeFanin, pEdgeFanout *Abc_Obj_t;
    var i, LevelMin, LevelMax, RetValue int32;

    // get the edge
    if p.pFanout {
        pEdgeFanout = p.pFanout;
    } else {
        pEdgeFanout = p.pNode;
    }
    if p.pFanout {
        pEdgeFanin = p.pNode;
    } else {
        pEdgeFanin = p.pFanin;
    }
    // get the minimum and maximum levels of the window
    LevelMin = Abc_MaxInt( 0, (int32(p.pFanin.Level)) - p.nFaninLevels );
    LevelMax = int32(pEdgeFanout.Level) + p.nFanoutLevels;

    // start the TFI leaves with the fanin
    Abc_NtkIncrementTravId( p.pNtk );
    Abc_NodeSetTravIdCurrent( p.pFanin );
    Vec_PtrPush( p.vFaninLeaves, p.pFanin );
    // mark the TFI cone and collect the leaves down to the given level
    for  Abc_NtkRRTfi_int(p.vFaninLeaves, LevelMin) != 0  {; }

    // mark the leaves with the new TravId
    Abc_NtkIncrementTravId( p.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFaninLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vFaninLeaves, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
    }
    // traverse the TFO cone of the leaves (while skipping the edge)
    // (a) mark the nodes in the cone using the current TravId
    // (b) collect the nodes that have external fanouts into p->vFanoutRoots
    for  Abc_NtkRRTfo_int(p.vFaninLeaves, p.vFanoutRoots, LevelMax, pEdgeFanin, pEdgeFanout) != 0  {; }

    // mark the fanout roots
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFanoutRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vFanoutRoots, i)), 1); i++  {
        pObj.fMarkA = 1;
    }
    // collect roots reachable from the fanout (p->vRoots)
    RetValue = Abc_NtkRRTfo_rec( pEdgeFanout, p.vRoots, LevelMax + 1 );
    // unmark the fanout roots
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFanoutRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vFanoutRoots, i)), 1); i++  {
        pObj.fMarkA = 0;
    }

    // return if the window is infeasible
    if RetValue == 0 {
        return 0; }

    // collect the DFS-ordered new cone (p->vCone) and new leaves (p->vLeaves)
    // using the previous marks coming from the TFO cone
    Abc_NtkIncrementTravId( p.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vRoots, i)), 1); i++  {
        Abc_NtkRRTfi_rec( pObj, p.vLeaves, p.vCone, LevelMin );
    }

    // create a new network
    p.pWnd = Abc_NtkWindow( p.pNtk, p.vLeaves, p.vCone, p.vRoots );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Marks the nodes in the TFI and collects their leaves.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRTfi_int(vLeaves *Vec_Ptr_t, LevelLimit int32) int32 {
    var pObj, pNext *Abc_Obj_t;
    var i, k, LevelMax, nSize int32;
    assert(LevelLimit >= 0, "LevelLimit >= 0");
    // find the maximum level of leaves
    LevelMax = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        if LevelMax < int32(pObj.Level) {
            LevelMax = pObj.Level; }
    }
    // if the nodes are all PIs, LevelMax == 0
    if LevelMax <= LevelLimit {
        return 0; }
    // expand the nodes with the minimum level
    nSize = Vec_PtrSize(vLeaves);
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < nSize) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        if LevelMax != int32(pObj.Level) {
            continue; }
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if Abc_NodeIsTravIdCurrent(pNext) {
                continue; }
            Abc_NodeSetTravIdCurrent( pNext );
            Vec_PtrPush( vLeaves, pNext );
        }
    }
    // remove old nodes (cannot remove a PI)
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        if LevelMax == int32(pObj.Level) {
            continue; }
        Vec_PtrWriteEntry( vLeaves, k++, pObj );
    }
    Vec_PtrShrink( vLeaves, k );
    if Vec_PtrSize(vLeaves) > 2000 {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Marks the nodes in the TFO and collects their roots.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRTfo_int(vLeaves *Vec_Ptr_t, vRoots *Vec_Ptr_t, LevelLimit int32, pEdgeFanin *Abc_Obj_t, pEdgeFanout *Abc_Obj_t) int32 {
    var pObj, pNext *Abc_Obj_t;
    var i, k, LevelMin, nSize, fObjIsRoot int32;
    // find the minimum level of leaves
    LevelMin = ABC_INFINITY;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        if LevelMin > int32(pObj.Level) {
            LevelMin = pObj.Level; }
    }
    // if the minimum level exceed the limit, we are done
    if LevelMin > LevelLimit {
        return 0; }
    // expand the nodes with the minimum level
    nSize = Vec_PtrSize(vLeaves);
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < nSize) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        if LevelMin != int32(pObj.Level) {
            continue; }
        fObjIsRoot = 0;
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
            // check if the fanout is outside of the cone
            if Abc_ObjIsCo(pNext) || pNext.Level > uint32(LevelLimit) {
                fObjIsRoot = 1;
                continue;
            }
            // skip the edge under check
            if pObj == pEdgeFanin && pNext == pEdgeFanout {
                continue; }
            // skip the visited fanouts
            if Abc_NodeIsTravIdCurrent(pNext) {
                continue; }
            Abc_NodeSetTravIdCurrent( pNext );
            Vec_PtrPush( vLeaves, pNext );
        }
        if fObjIsRoot != 0 {
            Vec_PtrPush( vRoots, pObj ); }
    }
    // remove old nodes
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        if LevelMin == int32(pObj.Level) {
            continue; }
        Vec_PtrWriteEntry( vLeaves, k++, pObj );
    }
    Vec_PtrShrink( vLeaves, k );
    if Vec_PtrSize(vLeaves) > 2000 {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Collects the roots in the TFO of the node.]

  Description [Note that this procedure can be improved by
  marking and skipping the visited nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRTfo_rec(pNode *Abc_Obj_t, vRoots *Vec_Ptr_t, LevelLimit int32) int32 {
    var pFanout *Abc_Obj_t;
    var i int32;
    // if we encountered a node outside of the TFO cone of the fanins, quit
    if Abc_ObjIsCo(pNode) || pNode.Level > uint32(LevelLimit) {
        return 0; }
    // if we encountered a node on the boundary, add it to the roots
    if pNode.fMarkA {
        Vec_PtrPushUnique( vRoots, pNode );
        return 1;
    }
    // mark the node with the current TravId (needed to have all internal nodes marked)
    Abc_NodeSetTravIdCurrent( pNode );
    // traverse the fanouts
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
        if !(Abc_NtkRRTfo_rec( pFanout, vRoots, LevelLimit ) != 0) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Collects the leaves and cone of the roots.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRTfi_rec(pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t, vCone *Vec_Ptr_t, LevelLimit int32) {
    var pFanin *Abc_Obj_t;
    var i int32;
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pNode) {
        return; }
    // add node to leaves if it is not in TFI cone of the leaves (marked before) or below the limit
    if !Abc_NodeIsTravIdPrevious(pNode) || int32(pNode.Level) <= LevelLimit {
        Abc_NodeSetTravIdCurrent( pNode );
        Vec_PtrPush( vLeaves, pNode );
        return;
    }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // call for the node's fanins
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_NtkRRTfi_rec( pFanin, vLeaves, vCone, LevelLimit );
    }
    // add the node to the cone in topological order
    Vec_PtrPush( vCone, pNode );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkWindow(pNtk *Abc_Ntk_t, vLeaves *Vec_Ptr_t, vCone *Vec_Ptr_t, vRoots *Vec_Ptr_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t; 
    var pObj *Abc_Obj_t;
    var fCheck int32  = 1;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // start the network
    pNtkNew = Abc_NtkAlloc( pNtk.ntkType, pNtk.ntkFunc, 1 );
    // duplicate the name and the spec
    pNtkNew.pName = Extra_UtilStrsav( "temp" );
    // map the constant nodes
    Abc_AigConst1(pNtk).pCopy = Abc_AigConst1(pNtkNew);
    // create and map the PIs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.pCopy = Abc_NtkCreatePi(pNtkNew);
    }
    // copy the AND gates
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCone)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vCone, i)), 1); i++  {
        pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) );
    }
    // compare the number of nodes before and after
    if Vec_PtrSize(vCone) != Abc_NtkNodeNum(pNtkNew) {
        fmt.Printf("Warning: Structural hashing during windowing reduced %d nodes (this is a bug).\n", Vec_PtrSize(vCone) - Abc_NtkNodeNum(pNtkNew)); }
    // create the POs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        assert(!Abc_ObjIsComplement(pObj.pCopy), "!Abc_ObjIsComplement(pObj.pCopy)");
        Abc_ObjAddFanin( Abc_NtkCreatePo(pNtkNew), pObj.pCopy );
    }
    // add the PI/PO names
    Abc_NtkAddDummyPiNames( pNtkNew );
    Abc_NtkAddDummyPoNames( pNtkNew );
    // check
    if (fCheck != 0) && !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkWindow: The network check has failed.\n");
        return nil;
    }
    return pNtkNew;
}


/**Function*************************************************************

  Synopsis    [Starts simulation to detect non-redundant edges.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRSimulateStart(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var uData, uData0, uData1 uint32;
    var i int32;
    Abc_AigConst1(pNtk).pData = (void *)~(uint32(0));
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pData = (void *)(ABC_PTRUINT_T)SIM_RANDOM_UNSIGNED;
    }
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if i == 0 { continue; }
        uData0 = uint32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pData;
        uData1 = uint32((ABC_PTRUINT_T))Abc_ObjFanin1(pObj).pData;
        if Abc_ObjFaninC0(pObj) {
            uData = ~uData0;
        } else {
            uData = uData0;
        }
        uData &= Abc_ObjFaninC1(pObj)? ~uData1 : uData1;
        assert(pObj.pData == nil, "pObj.pData == nil");
        pObj.pData = (void *)(ABC_PTRUINT_T)uData;
    }
    }
}

/**Function*************************************************************

  Synopsis    [Stops simulation to detect non-redundant edges.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRSimulateStop(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        pObj.pData = nil; }
    }
}







// static void Sim_TraverseNodes_rec( Abc_Obj_t * pRoot, Vec_Str_t * vTargets, Vec_Ptr_t * vNodes );

// static void Sim_CollectNodes_rec( Abc_Obj_t * pRoot, Vec_Ptr_t * vField );

// static void Sim_SimulateCollected( Vec_Str_t * vTargets, Vec_Ptr_t * vNodes, Vec_Ptr_t * vField );


/**Function*************************************************************

  Synopsis    [Simulation to detect non-redundant edges.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRRSimulate(pNtk *Abc_Ntk_t) *Vec_Str_t {
    var vNodes, vField *Vec_Ptr_t;
    var vTargets *Vec_Str_t;
    var pObj *Abc_Obj_t;
    var uData, uData0, uData1 uint32;
    var PrevCi, Phase, i, k int32;

    // start the candidates
    vTargets = Vec_StrStart( Abc_NtkObjNumMax(pNtk) + 1 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Phase = ((Abc_ObjFanoutNum(Abc_ObjFanin1(pObj)) > 1) << 1);
        Phase |= (Abc_ObjFanoutNum(Abc_ObjFanin0(pObj)) > 1);
        Vec_StrWriteEntry( vTargets, pObj.Id, byte(Phase) );
    }
    }

    // simulate patters and store them in copy
    Abc_AigConst1(pNtk).pCopy = (Abc_Obj_t *)~(uint32(0));
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)SIM_RANDOM_UNSIGNED;
    }
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if i == 0 { continue; }
        uData0 = uint32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pData;
        uData1 = uint32((ABC_PTRUINT_T))Abc_ObjFanin1(pObj).pData;
        if Abc_ObjFaninC0(pObj) {
            uData = ~uData0;
        } else {
            uData = uData0;
        }
        uData &= Abc_ObjFaninC1(pObj)? ~uData1 : uData1;
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)uData;
    }
    }
    // store the result in data
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        uData0 = uint32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pData;
        if Abc_ObjFaninC0(pObj) {
            pObj.pData = (void *)(ABC_PTRUINT_T)~uData0;
        } else {
            pObj.pData = (void *)(ABC_PTRUINT_T)uData0; }
    }

    // refine the candidates
    for  PrevCi = 0; PrevCi < Abc_NtkCiNum(pNtk); PrevCi = i  {
        vNodes = Vec_PtrAlloc( 10 );
        Abc_NtkIncrementTravId( pNtk );
        for  i = PrevCi; i < Abc_NtkCiNum(pNtk); i++  {
            Sim_TraverseNodes_rec( Abc_NtkCi(pNtk, i), vTargets, vNodes );
            if Vec_PtrSize(vNodes) > 128 {
                break; }
        }
        // collect the marked nodes in the topological order
        vField = Vec_PtrAlloc( 10 );
        Abc_NtkIncrementTravId( pNtk );
        // inlining of macro Abc_NtkForEachCo
        for  k = 0; (k < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, k)), 1); k++  {
            Sim_CollectNodes_rec( pObj, vField );
        }

        // simulate these nodes
        Sim_SimulateCollected( vTargets, vNodes, vField );
        // prepare for the next loop
        Vec_PtrFree( vNodes );
    }

    // clean
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        pObj.pData = nil; }
    }
    return vTargets;
}

/**Function*************************************************************

  Synopsis    [Collects nodes starting from the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_TraverseNodes_rec(pRoot *Abc_Obj_t, vTargets *Vec_Str_t, vNodes *Vec_Ptr_t) {
    var pFanout *Abc_Obj_t;
    var Entry byte;
    var k int32;
    if Abc_NodeIsTravIdCurrent(pRoot) {
        return; }
    Abc_NodeSetTravIdCurrent( pRoot );
    // save the reached targets
    Entry = Vec_StrEntry(vTargets, pRoot.Id);
    if Entry & 1 {
        Vec_PtrPush( vNodes, Abc_ObjNot(pRoot) ); }
    if Entry & 2 {
        Vec_PtrPush( vNodes, pRoot ); }
    // explore the fanouts
    // inlining of macro Abc_ObjForEachFanout
    for  k = 0; (k < Abc_ObjFanoutNum(pRoot)) && (((pFanout) = Abc_ObjFanout(pRoot, k)), 1); k++  {
        Sim_TraverseNodes_rec( pFanout, vTargets, vNodes );
    }
}

/**Function*************************************************************

  Synopsis    [Collects nodes starting from the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_CollectNodes_rec(pRoot *Abc_Obj_t, vField *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    if Abc_NodeIsTravIdCurrent(pRoot) {
        return; }
    if !Abc_NodeIsTravIdPrevious(pRoot) {
        return; }
    Abc_NodeSetTravIdCurrent( pRoot );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pRoot)) && (((pFanin) = Abc_ObjFanin(pRoot, i)), 1); i++  {
        Sim_CollectNodes_rec( pFanin, vField );
    }
    if !Abc_ObjIsCo(pRoot) {
        pRoot.pData = (void *)(ABC_PTRUINT_T)Vec_PtrSize(vField); }
    Vec_PtrPush( vField, pRoot );
}

/**Function*************************************************************

  Synopsis    [Simulate the given nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SimulateCollected(vTargets *Vec_Str_t, vNodes *Vec_Ptr_t, vField *Vec_Ptr_t) {
    var pObj, pFanin0, pFanin1, pDisproved *Abc_Obj_t;
    var vSims *Vec_Ptr_t;
    var pUnsigned, pUnsignedF *uint32;
    var i, k, Phase, fCompl int32;
    // get simulation info
    vSims = Sim_UtilInfoAlloc( Vec_PtrSize(vField), Vec_PtrSize(vNodes), 0 );
    // simulate the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vField)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vField, i)), 1); i++  {
        if Abc_ObjIsCi(pObj) {
            pUnsigned = (unsigned *)Vec_PtrEntry( vSims, i );
            for  k = 0; k < Vec_PtrSize(vNodes); k++  {
                pUnsigned[k] = uint32((ABC_PTRUINT_T))pObj.pCopy;
            }
            continue;
        }
        if Abc_ObjIsCo(pObj) {
            pUnsigned  = (unsigned *)Vec_PtrEntry( vSims, i );
            pUnsignedF = (unsigned *)Vec_PtrEntry( vSims, int32((ABC_PTRUINT_T))Abc_ObjFanin0(pObj).pData );
            if Abc_ObjFaninC0(pObj) {
                for  k = 0; k < Vec_PtrSize(vNodes); k++  {
                    pUnsigned[k] = ~pUnsignedF[k];
                }
            } else {
                for  k = 0; k < Vec_PtrSize(vNodes); k++  {
                    pUnsigned[k] = pUnsignedF[k];
                } }
            // update targets
            for  k = 0; k < Vec_PtrSize(vNodes); k++  {
                if pUnsigned[k] == uint32((ABC_PTRUINT_T))pObj.pData {
                    continue; }
                pDisproved = (Abc_Obj_t *)Vec_PtrEntry( vNodes, k );
                fCompl = Abc_ObjIsComplement(pDisproved);
                pDisproved = Abc_ObjRegular(pDisproved);
                Phase = Vec_StrEntry( vTargets, pDisproved.Id );
                if fCompl != 0 {
                    Phase = (Phase & 2);
                } else {
                    Phase = (Phase & 1); }
                Vec_StrWriteEntry( vTargets, pDisproved.Id, byte(Phase) );
            }
            continue;
        }
        // simulate the node
        pFanin0 = Abc_ObjFanin0(pObj);
        pFanin1 = Abc_ObjFanin1(pObj);
    }
}



/*
                {
                    unsigned uData;
                    if ( pFanin == Abc_ObjFanin0(pNode) )
                    {
                        uData = (unsigned)Abc_ObjFanin1(pNode)->pData;
                        uData = Abc_ObjFaninC1(pNode)? ~uData : uData;
                    }
                    else if ( pFanin == Abc_ObjFanin1(pNode) )
                    {
                        uData = (unsigned)Abc_ObjFanin0(pNode)->pData;
                        uData = Abc_ObjFaninC0(pNode)? ~uData : uData;
                    }
                    uData ^= (unsigned)pNode->pData;
//                    Extra_PrintBinary( stdout, &uData, 32 ); printf( "\n" );
                    if ( Extra_WordCountOnes(uData) > 8 )
                        continue;
                }
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

