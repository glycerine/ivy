package cut

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [cutOracle.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [K-feasible cut computation package.]

  Synopsis    [Procedures to compute cuts for a node using the oracle.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cutOracle.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cutInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Cut_OracleStruct_t_ struct {
    // cut comptupatation parameters
    pParams      *Cut_Params_t;
    vFanCounts         *Vec_Int_t;
    fSimul                int32;
    // storage for cuts
    vCutsNew         *Vec_Ptr_t;
    vCuts0         *Vec_Ptr_t;
    vCuts1         *Vec_Ptr_t;
    // oracle info 
    vNodeCuts         *Vec_Int_t;
    vNodeStarts         *Vec_Int_t;
    vCutPairs         *Vec_Int_t;
    // memory management
    pMmCuts   *Extra_MmFixed_t;
    EntrySize                int32;
    nTruthWords                int32;
    // stats
    timeTotal            abctime;
    nCuts                int32;
    nCutsTriv                int32;
}

// static Cut_Cut_t * Cut_CutStart( Cut_Oracle_t * p );

// static Cut_Cut_t * Cut_CutTriv( Cut_Oracle_t * p, int Node );

// static Cut_Cut_t * Cut_CutMerge( Cut_Oracle_t * p, Cut_Cut_t * pCut0, Cut_Cut_t * pCut1 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the cut oracle.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_OracleStart(pMan *Cut_Man_t) *Cut_Oracle_t {
    var p *Cut_Oracle_t;

    assert(pMan.pParams.nVarsMax >= 3 && pMan.pParams.nVarsMax <= CUT_SIZE_MAX, "pMan.pParams.nVarsMax >= 3 && pMan.pParams.nVarsMax <= CUT_SIZE_MAX");
    assert(pMan.pParams.fRecord, "pMan.pParams.fRecord");

    // inlining of macro ABC_ALLOC
    p = ((Cut_Oracle_t *) malloc(sizeof(Cut_Oracle_t) * (size_t)(1)));
    memset( p, 0, sizeof(Cut_Oracle_t) );

    // set and correct parameters
    p.pParams     = pMan.pParams;

    // transfer the recording info
    p.vNodeCuts   = pMan.vNodeCuts;    pMan.vNodeCuts   = nil;
    p.vNodeStarts = pMan.vNodeStarts;  pMan.vNodeStarts = nil;
    p.vCutPairs   = pMan.vCutPairs;    pMan.vCutPairs   = nil;

    // prepare storage for cuts
    p.vCutsNew = Vec_PtrAlloc( p.pParams.nIdsMax );
    Vec_PtrFill( p.vCutsNew, p.pParams.nIdsMax, nil );
    p.vCuts0 = Vec_PtrAlloc( 100 );
    p.vCuts1 = Vec_PtrAlloc( 100 );

    // entry size
    p.EntrySize = sizeof(Cut_Cut_t) + p.pParams.nVarsMax * sizeof(int);
    if p.pParams.fTruth {
        if p.pParams.nVarsMax > 8 {
            p.pParams.fTruth = 0;
            fmt.Printf("Skipping computation of truth table for more than 8 inputs.\n");
        
        } else {
            p.nTruthWords = Cut_TruthWords( p.pParams.nVarsMax );
            p.EntrySize += p.nTruthWords * sizeof(unsigned);
        }
    }
    // memory for cuts
    p.pMmCuts = Extra_MmFixedStart( p.EntrySize );
    return p;
}
/**Function*************************************************************

  Synopsis    [Stop the cut oracle.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_OracleStop(p *Cut_Oracle_t) {
//    if ( p->pParams->fVerbose )
    {
        fmt.Printf("Cut computation statistics with oracle:\n");
        fmt.Printf("Current cuts      = %8d. (Trivial = %d.)\n", p.nCuts-p.nCutsTriv, p.nCutsTriv);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total time "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
    }

    if p.vCuts0 {      Vec_PtrFree( p.vCuts0 ); }
    if p.vCuts1 {      Vec_PtrFree( p.vCuts1 ); }
    if p.vCutsNew {    Vec_PtrFree( p.vCutsNew ); }
    if p.vFanCounts {  Vec_IntFree( p.vFanCounts ); }

    if p.vNodeCuts {   Vec_IntFree( p.vNodeCuts ); }
    if p.vNodeStarts { Vec_IntFree( p.vNodeStarts ); }
    if p.vCutPairs {   Vec_IntFree( p.vCutPairs ); }

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
func Cut_OracleSetFanoutCounts(p *Cut_Oracle_t, vFanCounts *Vec_Int_t) {
    p.vFanCounts = vFanCounts;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_OracleReadDrop(p *Cut_Oracle_t) int32 {
    return p.pParams.fDrop;
}

/**Function*************************************************************

  Synopsis    [Sets the trivial cut for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_OracleNodeSetTriv(p *Cut_Oracle_t, Node int32) {
    assert(Vec_PtrEntry( p.vCutsNew, Node ) == nil, "Vec_PtrEntry( p.vCutsNew, Node ) == nil");
    Vec_PtrWriteEntry( p.vCutsNew, Node, Cut_CutTriv(p, Node) );
}



/**Function*************************************************************

  Synopsis    [Allocates the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutStart(p *Cut_Oracle_t) *Cut_Cut_t {
    var pCut *Cut_Cut_t;
    // cut allocation
    pCut = (Cut_Cut_t *)Extra_MmFixedEntryFetch( p.pMmCuts );
    memset( pCut, 0, sizeof(Cut_Cut_t) );
    pCut.nVarsMax   = p.pParams.nVarsMax;
    pCut.fSimul     = p.fSimul;
    p.nCuts++;
    return pCut;
}

/**Function*************************************************************

  Synopsis    [Creates the trivial cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutTriv(p *Cut_Oracle_t, Node int32) *Cut_Cut_t {
    var pCut *Cut_Cut_t;
    pCut = Cut_CutStart( p );
    pCut.nLeaves    = 1;
    pCut.pLeaves[0] = Node;
    if p.pParams.fTruth {
        var pTruth *uint32  = Cut_CutReadTruth(pCut);
        var i int32;
        for  i = 0; i < p.nTruthWords; i++  {
            pTruth[i] = 0xAAAAAAAA;
        }
    }
    p.nCutsTriv++;
    return pCut;
} 

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutMerge(p *Cut_Oracle_t, pCut0 *Cut_Cut_t, pCut1 *Cut_Cut_t) *Cut_Cut_t {
    var pCut *Cut_Cut_t;
    var Limit, i, k, c int32;
    // create the leaves of the new cut
    pCut = Cut_CutStart( p );
    Limit = p.pParams.nVarsMax;
    for  i = k = c = 0; c < Limit; c++  {
        if k == int32(pCut1.nLeaves) {
            if i == int32(pCut0.nLeaves) {
                pCut.nLeaves = c;
                return pCut;
            }
            pCut.pLeaves[c] = pCut0.pLeaves[i]; i++;
            continue;
        }
        if i == int32(pCut0.nLeaves) {
            if k == int32(pCut1.nLeaves) {
                pCut.nLeaves = c;
                return pCut;
            }
            pCut.pLeaves[c] = pCut1.pLeaves[k]; k++;
            continue;
        }
        if pCut0.pLeaves[i] < pCut1.pLeaves[k] {
            pCut.pLeaves[c] = pCut0.pLeaves[i]; i++;
            continue;
        }
        if pCut0.pLeaves[i] > pCut1.pLeaves[k] {
            pCut.pLeaves[c] = pCut1.pLeaves[k]; k++;
            continue;
        }
        pCut.pLeaves[c] = pCut0.pLeaves[i]; i++; 
        k++;
    }
    assert(i == int32(pCut0.nLeaves) && k == int32(pCut1.nLeaves), "i == int32(pCut0.nLeaves) && k == int32(pCut1.nLeaves)");
    pCut.nLeaves = c;
    return pCut;
} 

/**Function*************************************************************

  Synopsis    [Reconstruct the cuts of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_OracleComputeCuts(p *Cut_Oracle_t, Node int32, Node0 int32, Node1 int32, fCompl0 int32, fCompl1 int32) *Cut_Cut_t {
    var pList *Cut_Cut_t = nil;
    var ppTail **Cut_Cut_t = &pList;
    var pCut, pCut0, pCut1, pList0, pList1 *Cut_Cut_t;
    var iCutStart, nCuts, i, Entry int32;
    var clk abctime  = Abc_Clock();

    // get the cuts of the children
    pList0 = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsNew, Node0 );
    pList1 = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsNew, Node1 );
    assert(pList0 && (pList1 != nil), "pList0 && pList1");

    // get the complemented attribute of the cut
    p.fSimul = (fCompl0 ^ pList0.fSimul) & (fCompl1 ^ pList1.fSimul);

    // collect the cuts
    Vec_PtrClear( p.vCuts0 );
    // inlining of macro Cut_ListForEachCut
    for  pCut = pList0;                                             (pCut != nil);                                                     pCut = pCut.pNext  {
        Vec_PtrPush( p.vCuts0, pCut );
    }
    Vec_PtrClear( p.vCuts1 );
    // inlining of macro Cut_ListForEachCut
    for  pCut = pList1;                                             (pCut != nil);                                                     pCut = pCut.pNext  {
        Vec_PtrPush( p.vCuts1, pCut );
    }

    // get the first and last cuts of this node
    nCuts = Vec_IntEntry(p.vNodeCuts, Node);
    iCutStart = Vec_IntEntry(p.vNodeStarts, Node);

    // create trivial cut
    assert(Vec_IntEntry(p.vCutPairs, iCutStart) == 0, "Vec_IntEntry(p.vCutPairs, iCutStart) == 0");
    pCut = Cut_CutTriv( p, Node );
    *ppTail = pCut;
    ppTail = &pCut.pNext;
    // create other cuts
    for  i = 1; i < nCuts; i++  {
        Entry = Vec_IntEntry( p.vCutPairs, iCutStart + i );
        pCut0 = (Cut_Cut_t *)Vec_PtrEntry( p.vCuts0, Entry & 0xFFFF );
        pCut1 = (Cut_Cut_t *)Vec_PtrEntry( p.vCuts1, Entry >> 16 );
        pCut  = Cut_CutMerge( p, pCut0, pCut1 );
        *ppTail = pCut;
        ppTail = &pCut.pNext;
        // compute the truth table
        if p.pParams.fTruth {
            Cut_TruthComputeOld( pCut, pCut0, pCut1, fCompl0, fCompl1 ); }
    }
    *ppTail = nil;

    // write the new cut
    assert(Vec_PtrEntry( p.vCutsNew, Node ) == nil, "Vec_PtrEntry( p.vCutsNew, Node ) == nil");
    Vec_PtrWriteEntry( p.vCutsNew, Node, pList );
p.timeTotal += Abc_Clock() - clk;
    return pList;
}

/**Function*************************************************************

  Synopsis    [Deallocates the cuts at the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_OracleFreeCuts(p *Cut_Oracle_t, Node int32) {
    var pList, pCut, pCut2 *Cut_Cut_t;
    pList = (Cut_Cut_t *)Vec_PtrEntry( p.vCutsNew, Node );
    if pList == nil {
        return; }
    // inlining of macro Cut_ListForEachCutSafe
    for  pCut = pList,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
        Extra_MmFixedEntryRecycle( p.pMmCuts, (char *)pCut );
    }
    Vec_PtrWriteEntry( p.vCutsNew, Node, pList );
}

/**Function*************************************************************

  Synopsis    [Consider dropping cuts if they are useless by now.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_OracleTryDroppingCuts(p *Cut_Oracle_t, Node int32) {
    var nFanouts int32;
    assert(p.vFanCounts, "p.vFanCounts");
    nFanouts = Vec_IntEntry( p.vFanCounts, Node );
    assert(nFanouts > 0, "nFanouts > 0");
    if nFanouts--; nFanouts == 0 {
        Cut_OracleFreeCuts( p, Node ); }
    Vec_IntWriteEntry( p.vFanCounts, Node, nFanouts );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

