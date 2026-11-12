package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcRewrite.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Technology-independent resynthesis of the AIG based on DAG aware rewriting.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcRewrite.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "opt/rwr/rwr.h"
//#include "bool/dec/dec.h"

//ABC_NAMESPACE_IMPL_START


/*
    The ideas realized in this package are inspired by the paper:
    Per Bjesse, Arne Boralv, "DAG-aware circuit compression for 
    formal verification", Proc. ICCAD 2004, pp. 42-49.
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Cut_Man_t * Abc_NtkStartCutManForRewrite( Abc_Ntk_t * pNtk );

// static void        Abc_NodePrintCuts( Abc_Obj_t * pNode );

// static void        Abc_ManShowCutCone( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves );


// extern void  Abc_PlaceBegin( Abc_Ntk_t * pNtk );

// extern void  Abc_PlaceEnd( Abc_Ntk_t * pNtk );

// extern void  Abc_PlaceUpdate( Vec_Ptr_t * vAddedCells, Vec_Ptr_t * vUpdatedNets );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs incremental rewriting of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRewrite(pNtk *Abc_Ntk_t, fUpdateLevel int32, fUseZeros int32, fVerbose int32, fVeryVerbose int32, fPlaceEnable int32) int32 {
    extern void           Dec_GraphUpdateNetwork( Abc_Obj_t * pRoot, Dec_Graph_t * pGraph, int fUpdateLevel, int nGain );
    var pProgress *ProgressBar;
    var pManCut *Cut_Man_t;
    var pManRwr *Rwr_Man_t;
    var pNode *Abc_Obj_t;
//    Vec_Ptr_t * vAddedCells = NULL, * vUpdatedNets = NULL;
    var pGraph *Dec_Graph_t;
    var i, nNodes, nGain, fCompl int32;
    var clk abctime;
    var clkStart abctime = Abc_Clock();

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // cleanup the AIG
    Abc_AigCleanup((Abc_Aig_t *)pNtk.pManFunc);
/*
    {
        Vec_Vec_t * vParts;
        vParts = Abc_NtkPartitionSmart( pNtk, 50, 1 );
        Vec_VecFree( vParts );
    }
*/

    // start placement package
//    if ( fPlaceEnable )
//    {
//        Abc_PlaceBegin( pNtk );
//        vAddedCells = Abc_AigUpdateStart( pNtk->pManFunc, &vUpdatedNets );
//    }

    // start the rewriting manager
    pManRwr = Rwr_ManStart( 0 );
    if pManRwr == nil {
        return 0; }
    // compute the reverse levels if level update is requested
    if fUpdateLevel != 0 {
        Abc_NtkStartReverseLevels( pNtk, 0 ); }
    // start the cut manager
clk = Abc_Clock();
    pManCut = Abc_NtkStartCutManForRewrite( pNtk );
Rwr_ManAddTimeCuts( pManRwr, Abc_Clock() - clk );
    pNtk.pManCut = pManCut;

    if fVeryVerbose != 0 {
        Rwr_ScoresClean( pManRwr ); }

    // resynthesize each node once
    pManRwr.nNodesBeg = Abc_NtkNodeNum(pNtk);
    nNodes = Abc_NtkObjNumMax(pNtk);
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        // stop if all nodes have been tried once
        if i >= nNodes {
            break; }
        // skip persistant nodes
        if Abc_NodeIsPersistant(pNode) {
            continue; }
        // skip the nodes with many fanouts
        if Abc_ObjFanoutNum(pNode) > 1000 {
            continue; }

        // for each cut, try to resynthesize it
        nGain = Rwr_NodeRewrite( pManRwr, pManCut, pNode, fUpdateLevel, fUseZeros, fPlaceEnable );
        if !(nGain > 0 || (nGain == 0 && (fUseZeros != 0))) {
            continue; }
        // if we end up here, a rewriting step is accepted

        // get hold of the new subgraph to be added to the AIG
        pGraph = (Dec_Graph_t *)Rwr_ManReadDecs(pManRwr);
        fCompl = Rwr_ManReadCompl(pManRwr);

        // reset the array of the changed nodes
        if fPlaceEnable != 0 {
            Abc_AigUpdateReset( (Abc_Aig_t *)pNtk.pManFunc ); }

        // complement the FF if needed
        if fCompl != 0 { Dec_GraphComplement( pGraph ); }
clk = Abc_Clock();
        Dec_GraphUpdateNetwork( pNode, pGraph, fUpdateLevel, nGain );
Rwr_ManAddTimeUpdate( pManRwr, Abc_Clock() - clk );
        if fCompl != 0 { Dec_GraphComplement( pGraph ); }

        // use the array of changed nodes to update placement
//        if ( fPlaceEnable )
//            Abc_PlaceUpdate( vAddedCells, vUpdatedNets );
    }
    }
    Extra_ProgressBarStop( pProgress );
Rwr_ManAddTimeTotal( pManRwr, Abc_Clock() - clkStart );
    // print stats
    pManRwr.nNodesEnd = Abc_NtkNodeNum(pNtk);
    if fVerbose != 0 {
        Rwr_ManPrintStats( pManRwr ); }
//        Rwr_ManPrintStatsFile( pManRwr );
    if fVeryVerbose != 0 {
        Rwr_ScoresReport( pManRwr ); }
    // delete the managers
    Rwr_ManStop( pManRwr );
    Cut_ManStop( pManCut );
    pNtk.pManCut = nil;

    // start placement package
//    if ( fPlaceEnable )
//    {
//        Abc_PlaceEnd( pNtk );
//        Abc_AigUpdateStop( pNtk->pManFunc );
//    }

    // put the nodes into the DFS order and reassign their IDs
    {
//        abctime clk = Abc_Clock();
    Abc_NtkReassignIds( pNtk );
//        ABC_PRT( "time", Abc_Clock() - clk );
    }
//    Abc_AigCheckFaninOrder( pNtk->pManFunc );
    // fix the levels
    if fUpdateLevel != 0 {
        Abc_NtkStopReverseLevels( pNtk );
    } else {
        Abc_NtkLevel( pNtk ); }
    // check
    if !Abc_NtkCheck( pNtk ) {
        fmt.Printf("Abc_NtkRewrite: The network check has failed.\n");
        return 0;
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Starts the cut manager for rewriting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_NtkStartCutManForRewrite_Params Cut_Params_t;
var spStatic_Abc_NtkStartCutManForRewrite_pParams *Cut_Params_t = &spStatic_Abc_NtkStartCutManForRewrite_Params;
func Abc_NtkStartCutManForRewrite(pNtk *Abc_Ntk_t) *Cut_Man_t {
    
    var pManCut *Cut_Man_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // start the cut manager
    memset( spStatic_Abc_NtkStartCutManForRewrite_pParams, 0, sizeof(Cut_Params_t) );
    spStatic_Abc_NtkStartCutManForRewrite_pParams.nVarsMax  = 4;     // the max cut size ("k" of the k-feasible cuts)
    spStatic_Abc_NtkStartCutManForRewrite_pParams.nKeepMax  = 250;   // the max number of cuts kept at a node
    spStatic_Abc_NtkStartCutManForRewrite_pParams.fTruth    = 1;     // compute truth tables
    spStatic_Abc_NtkStartCutManForRewrite_pParams.fFilter   = 1;     // filter dominated cuts
    spStatic_Abc_NtkStartCutManForRewrite_pParams.fSeq      = 0;     // compute sequential cuts
    spStatic_Abc_NtkStartCutManForRewrite_pParams.fDrop     = 0;     // drop cuts on the fly
    spStatic_Abc_NtkStartCutManForRewrite_pParams.fVerbose  = 0;     // the verbosiness flag
    spStatic_Abc_NtkStartCutManForRewrite_pParams.nIdsMax   = Abc_NtkObjNumMax( pNtk );
    pManCut = Cut_ManStart( spStatic_Abc_NtkStartCutManForRewrite_pParams );
    if spStatic_Abc_NtkStartCutManForRewrite_pParams.fDrop {
        Cut_ManSetFanoutCounts( pManCut, Abc_NtkFanoutCounts(pNtk) ); }
    // set cuts for PIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_ObjFanoutNum(pObj) > 0 {
            Cut_NodeSetTriv( pManCut, pObj.Id ); }
    }
    return pManCut;
}

/**Function*************************************************************

  Synopsis    [Prints the cuts at the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodePrintCuts(pNode *Abc_Obj_t) {
    var vCuts *Vec_Ptr_t;
    var pCut *Cut_Cut_t;
    var k int32;

    fmt.Printf("\nNode %s\n", Abc_ObjName(pNode));
    vCuts = (Vec_Ptr_t *)pNode.pCopy;
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vCuts)) && (((pCut) = (Cut_Cut_t *)Vec_PtrEntry(vCuts, k)), 1); k++  {
        Extra_PrintBinary( os.Stdout, (unsigned *)&pCut.uSign, 16 ); 
        fmt.Printf("   ");
        Cut_CutPrint( pCut, 0 );   
        fmt.Printf("\n");
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManRewritePrintDivs(vDivs *Vec_Ptr_t, nLeaves int32) {
    var pFanin, pNode, pRoot *Abc_Obj_t;
    var i, k int32;
    pRoot = (Abc_Obj_t *)Vec_PtrEntryLast(vDivs);
    // print the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDivs)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vDivs, i)), 1); i++  {
        if i < nLeaves {
            fmt.Printf("%6d : %c\n", pNode.Id, 'a'+i);
            continue;
        }
        fmt.Printf("%6d : %2d = ", pNode.Id, i);
        // find the first fanin
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vDivs)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(vDivs, k)), 1); k++  {
            if Abc_ObjFanin0(pNode) == pFanin {
                break; }
        }
        if k < nLeaves {
            fmt.Printf("%c", 'a' + k);
        } else {
            fmt.Printf("%d", k); }
        fmt.Printf("%s ", Abc_ObjFaninC0(pNode)? "\'" : "");
        // find the second fanin
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vDivs)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(vDivs, k)), 1); k++  {
            if Abc_ObjFanin1(pNode) == pFanin {
                break; }
        }
        if k < nLeaves {
            fmt.Printf("%c", 'a' + k);
        } else {
            fmt.Printf("%d", k); }
        fmt.Printf("%s ", Abc_ObjFaninC1(pNode)? "\'" : "");
        if pNode == pRoot {
            fmt.Printf(" root"); }
        fmt.Printf("\n");
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManShowCutCone_rec(pNode *Abc_Obj_t, vDivs *Vec_Ptr_t) {
    if Abc_NodeIsTravIdCurrent(pNode) {
        return; }
    Abc_NodeSetTravIdCurrent(pNode);
    Abc_ManShowCutCone_rec( Abc_ObjFanin0(pNode), vDivs );
    Abc_ManShowCutCone_rec( Abc_ObjFanin1(pNode), vDivs );
    Vec_PtrPush( vDivs, pNode );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManShowCutCone(pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t) {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    var pObj *Abc_Obj_t;
    var vDivs *Vec_Ptr_t;
    var i int32;
    vDivs = Vec_PtrAlloc( 100 );
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( Abc_ObjRegular(pObj) );
        Vec_PtrPush( vDivs, Abc_ObjRegular(pObj) );
    }
    Abc_ManShowCutCone_rec( pNode, vDivs );
    Abc_ManRewritePrintDivs( vDivs, Vec_PtrSize(vLeaves) );
    Vec_PtrFree( vDivs );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_RwrExpWithCut_rec(pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t, fUseA int32) {
    if Vec_PtrFind(vLeaves, pNode) >= 0 || Vec_PtrFind(vLeaves, Abc_ObjNot(pNode)) >= 0 {
        if fUseA != 0 {
            Abc_ObjRegular(pNode).fMarkA = 1;
        } else {
            Abc_ObjRegular(pNode).fMarkB = 1; }
        return;
    }
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    Abc_RwrExpWithCut_rec( Abc_ObjFanin0(pNode), vLeaves, fUseA );
    Abc_RwrExpWithCut_rec( Abc_ObjFanin1(pNode), vLeaves, fUseA );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_RwrExpWithCut(pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t) {
    var pObj *Abc_Obj_t;
    var i, CountA, CountB int32;
    Abc_RwrExpWithCut_rec( Abc_ObjFanin0(pNode), vLeaves, 1 );
    Abc_RwrExpWithCut_rec( Abc_ObjFanin1(pNode), vLeaves, 0 );
    CountA = CountB = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        CountA += Abc_ObjRegular(pObj).fMarkA;
        CountB += Abc_ObjRegular(pObj).fMarkB;
        Abc_ObjRegular(pObj).fMarkA = 0;
        Abc_ObjRegular(pObj).fMarkB = 0;
    }
    fmt.Printf("(%d,%d:%d) ", CountA, CountB, CountA+CountB-Vec_PtrSize(vLeaves));
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

