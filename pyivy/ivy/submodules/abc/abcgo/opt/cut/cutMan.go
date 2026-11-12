package cut

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [cutMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [K-feasible cut computation package.]

  Synopsis    [Cut manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cutMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cutInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern void Npn_StartTruth8( uint8 uTruths[][32] );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the cut manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManStart(pParams *Cut_Params_t) *Cut_Man_t {
    var p *Cut_Man_t;
//    extern int nTruthDsd;
//    nTruthDsd = 0;
    assert(pParams.nVarsMax >= 3 && pParams.nVarsMax <= CUT_SIZE_MAX, "pParams.nVarsMax >= 3 && pParams.nVarsMax <= CUT_SIZE_MAX");
    // inlining of macro ABC_ALLOC
    p = ((Cut_Man_t *) malloc(sizeof(Cut_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Cut_Man_t) );
    // set and correct parameters
    p.pParams = pParams;
    // prepare storage for cuts
    p.vCutsNew = Vec_PtrAlloc( pParams.nIdsMax );
    Vec_PtrFill( p.vCutsNew, pParams.nIdsMax, nil );
    // prepare storage for sequential cuts
    if pParams.fSeq {
        p.pParams.fFilter = 1;
        p.vCutsOld = Vec_PtrAlloc( pParams.nIdsMax );
        Vec_PtrFill( p.vCutsOld, pParams.nIdsMax, nil );
        p.vCutsTemp = Vec_PtrAlloc( pParams.nCutSet );
        Vec_PtrFill( p.vCutsTemp, pParams.nCutSet, nil );
        if pParams.fTruth && pParams.nVarsMax > 5 {
            pParams.fTruth = 0;
            fmt.Printf("Skipping computation of truth tables for sequential cuts with more than 5 inputs.\n");
        }
    }
    // entry size
    p.EntrySize = sizeof(Cut_Cut_t) + pParams.nVarsMax * sizeof(int);
    if pParams.fTruth {
        if pParams.nVarsMax > 14 {
            pParams.fTruth = 0;
            fmt.Printf("Skipping computation of truth table for more than %d inputs.\n", 14);
        
        } else {
            p.nTruthWords = Cut_TruthWords( pParams.nVarsMax );
            p.EntrySize += p.nTruthWords * sizeof(unsigned);
        }
        // inlining of macro ABC_ALLOC
        p.puTemp[0] = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(4 * p.nTruthWords)));
        p.puTemp[1] = p.puTemp[0] + p.nTruthWords;
        p.puTemp[2] = p.puTemp[1] + p.nTruthWords;
        p.puTemp[3] = p.puTemp[2] + p.nTruthWords;
    }
    // enable cut computation recording
    if pParams.fRecord {
        p.vNodeCuts   = Vec_IntStart( pParams.nIdsMax );
        p.vNodeStarts = Vec_IntStart( pParams.nIdsMax );
        p.vCutPairs   = Vec_IntAlloc( 0 );
    }
    // allocate storage for delays
    if pParams.fMap && !p.pParams.fSeq {
        p.vDelays = Vec_IntStart( pParams.nIdsMax );
        p.vDelays2 = Vec_IntStart( pParams.nIdsMax );
        p.vCutsMax = Vec_PtrStart( pParams.nIdsMax );
    }
    // memory for cuts
    p.pMmCuts = Extra_MmFixedStart( p.EntrySize );
    p.vTemp = Vec_PtrAlloc( 100 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the cut manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManStop(p *Cut_Man_t) {
    if p.vCutsNew {    Vec_PtrFree( p.vCutsNew ); }
    if p.vCutsOld {    Vec_PtrFree( p.vCutsOld ); }
    if p.vCutsTemp {   Vec_PtrFree( p.vCutsTemp ); }
    if p.vFanCounts {  Vec_IntFree( p.vFanCounts ); }
    if p.vTemp {       Vec_PtrFree( p.vTemp ); }

    if p.vCutsMax {    Vec_PtrFree( p.vCutsMax ); }
    if p.vDelays {     Vec_IntFree( p.vDelays ); }
    if p.vDelays2 {    Vec_IntFree( p.vDelays2 ); }
    if p.vNodeCuts {   Vec_IntFree( p.vNodeCuts ); }
    if p.vNodeStarts { Vec_IntFree( p.vNodeStarts ); }
    if p.vCutPairs {   Vec_IntFree( p.vCutPairs ); }
    // inlining of macro ABC_FREE
    if p.puTemp[0] {   ((p.puTemp[0]) ? (free((char *) (p.puTemp[0])), (p.puTemp[0]) = 0) : 0); }

    Extra_MmFixedStop( p.pMmCuts );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManPrintStats(p *Cut_Man_t) {
    if p.pReady {
        Cut_CutRecycle( p, p.pReady );
        p.pReady = nil;
    }
    fmt.Printf("Cut computation statistics:\n");
    fmt.Printf("Current cuts      = %8d. (Trivial = %d.)\n", p.nCutsCur-p.nCutsTriv, p.nCutsTriv);
    fmt.Printf("Peak cuts         = %8d.\n", p.nCutsPeak);
    fmt.Printf("Total allocated   = %8d.\n", p.nCutsAlloc);
    fmt.Printf("Total deallocated = %8d.\n", p.nCutsDealloc);
    fmt.Printf("Cuts filtered     = %8d.\n", p.nCutsFilter);
    fmt.Printf("Nodes saturated   = %8d. (Max cuts = %d.)\n", p.nCutsLimit, p.pParams.nKeepMax);
    fmt.Printf("Cuts per node     = %8.1f\n", (float32((p.nCutsCur-p.nCutsTriv)))/p.nNodes);
    fmt.Printf("The cut size      = %8d bytes.\n", p.EntrySize);
    fmt.Printf("Peak memory       = %8.2f MB.\n", float32(p.nCutsPeak) * p.EntrySize / (1<<20));
    fmt.Printf("Total nodes       = %8d.\n", p.nNodes);
    if p.pParams.fDag || p.pParams.fTree {
    fmt.Printf("DAG nodes         = %8d.\n", p.nNodesDag);
    fmt.Printf("Tree nodes        = %8d.\n", p.nNodes - p.nNodesDag);
    }
    fmt.Printf("Nodes w/o cuts    = %8d.\n", p.nNodesNoCuts);
    if p.pParams.fMap && !p.pParams.fSeq {
    fmt.Printf("Mapping delay     = %8d.\n", p.nDelayMin); }

    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Merge "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMerge)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Union "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeUnion)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Filter"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeFilter)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Truth "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTruth)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Map   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMap)))/(float64(int64(time.Second))));;
//    printf( "Nodes = %d. Multi = %d.  Cuts = %d. Multi = %d.\n", 
//        p->nNodes, p->nNodesMulti, p->nCutsCur-p->nCutsTriv, p->nCutsMulti );
//    printf( "Count0 = %d. Count1 = %d. Count2 = %d.\n\n", p->Count0, p->Count1, p->Count2 );
}

    
/**Function*************************************************************

  Synopsis    [Prints some interesting stats.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManPrintStatsToFile(p *Cut_Man_t, pFileName string, TimeTotal abctime) {
    var pTable *os.File;
    pTable = os.OpenFile("cut_stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    fmt.Fprintf(pTable, "%-20s ", pFileName);
    fmt.Fprintf(pTable, "%8d ", p.nNodes);
    fmt.Fprintf(pTable, "%6.1f ", (float32((p.nCutsCur)))/p.nNodes);
    fmt.Fprintf(pTable, "%6.2f ", (float32((100.0 * p.nCutsLimit)))/p.nNodes);
    fmt.Fprintf(pTable, "%6.2f ", float32(p.nCutsPeak) * p.EntrySize / (1<<20));
    fmt.Fprintf(pTable, "%6.2f ", float32((TimeTotal))/float32((int64(time.Second))));
    fmt.Fprintf(pTable, "\n");
    _ = (pTable).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManSetFanoutCounts(p *Cut_Man_t, vFanCounts *Vec_Int_t) {
    p.vFanCounts = vFanCounts;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManSetNodeAttrs(p *Cut_Man_t, vNodeAttrs *Vec_Int_t) {
    p.vNodeAttrs = vNodeAttrs;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManReadVarsMax(p *Cut_Man_t) int32 {
    return p.pParams.nVarsMax;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManReadParams(p *Cut_Man_t) *Cut_Params_t {
    return p.pParams;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManReadNodeAttrs(p *Cut_Man_t) *Vec_Int_t {
    return p.vNodeAttrs;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_ManIncrementDagNodes(p *Cut_Man_t) {
    p.nNodesDag++;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

