package abci

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [abcResub.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Resubstitution manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcResub.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "bool/dec/dec.h"

//ABC_NAMESPACE_IMPL_START

 
////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const ABC_RS_DIV1_MAX =    150   // the max number of divisors to consider
const ABC_RS_DIV2_MAX =    500   // the max number of pair-wise divisors to consider

   
type Abc_ManRes_t struct {
    // paramers
    nLeavesMax                int32; // the max number of leaves in the cone
    nDivsMax                int32;   // the max number of divisors in the cone
    // representation of the cone
    pRoot         *Abc_Obj_t;      // the root of the cone
    nLeaves                int32;    // the number of leaves
    nDivs                int32;      // the number of all divisor (including leaves)
    nMffc                int32;      // the size of MFFC
    nLastGain                int32;  // the gain the number of nodes
    vDivs         *Vec_Ptr_t;      // the divisors
    // representation of the simulation info
    nBits                int32;      // the number of simulation bits
    nWords                int32;     // the number of unsigneds for siminfo
    vSims         *Vec_Ptr_t;      // simulation info
    pInfo          *uint32;      // pointer to simulation info
    // observability don't-cares
    pCareSet          *uint32;
    // internal divisor storage
    vDivs1UP         *Vec_Ptr_t;   // the single-node unate divisors
    vDivs1UN         *Vec_Ptr_t;   // the single-node unate divisors
    vDivs1B         *Vec_Ptr_t;    // the single-node binate divisors
    vDivs2UP0         *Vec_Ptr_t;  // the double-node unate divisors
    vDivs2UP1         *Vec_Ptr_t;  // the double-node unate divisors
    vDivs2UN0         *Vec_Ptr_t;  // the double-node unate divisors
    vDivs2UN1         *Vec_Ptr_t;  // the double-node unate divisors
    // other data
    vTemp         *Vec_Ptr_t;      // temporary array of nodes
    // runtime statistics
    timeCut            abctime;
    timeTruth            abctime;
    timeRes            abctime;
    timeDiv            abctime;
    timeMffc            abctime;
    timeSim            abctime;
    timeRes1            abctime;
    timeResD            abctime;
    timeRes2            abctime;
    timeRes3            abctime;
    timeNtk            abctime;
    timeTotal            abctime;
    // improvement statistics
    nUsedNodeC                int32;
    nUsedNode0                int32;
    nUsedNode1Or                int32;
    nUsedNode1And                int32;
    nUsedNode2Or                int32;
    nUsedNode2And                int32;
    nUsedNode2OrAnd                int32;
    nUsedNode2AndOr                int32;
    nUsedNode3OrAnd                int32;
    nUsedNode3AndOr                int32;
    nUsedNodeTotal                int32;
    nTotalDivs                int32;
    nTotalLeaves                int32;
    nTotalGain                int32;
    nNodesBeg                int32;
    nNodesEnd                int32;
}

// external procedures
// static Abc_ManRes_t* Abc_ManResubStart( int nLeavesMax, int nDivsMax );

// static void          Abc_ManResubStop( Abc_ManRes_t * p );

// static Dec_Graph_t * Abc_ManResubEval( Abc_ManRes_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, int nSteps, int fUpdateLevel, int fVerbose );

// static void          Abc_ManResubCleanup( Abc_ManRes_t * p );

// static void          Abc_ManResubPrint( Abc_ManRes_t * p );


// other procedures
// static int           Abc_ManResubCollectDivs( Abc_ManRes_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves, int Required );

// static void          Abc_ManResubSimulate( Vec_Ptr_t * vDivs, int nLeaves, Vec_Ptr_t * vSims, int nLeavesMax, int nWords );

// static void          Abc_ManResubPrintDivs( Abc_ManRes_t * p, Abc_Obj_t * pRoot, Vec_Ptr_t * vLeaves );


// static void          Abc_ManResubDivsS( Abc_ManRes_t * p, int Required );

// static void          Abc_ManResubDivsD( Abc_ManRes_t * p, int Required );

// static Dec_Graph_t * Abc_ManResubQuit( Abc_ManRes_t * p );

// static Dec_Graph_t * Abc_ManResubDivs0( Abc_ManRes_t * p );

// static Dec_Graph_t * Abc_ManResubDivs1( Abc_ManRes_t * p, int Required );

// static Dec_Graph_t * Abc_ManResubDivs12( Abc_ManRes_t * p, int Required );

// static Dec_Graph_t * Abc_ManResubDivs2( Abc_ManRes_t * p, int Required );

// static Dec_Graph_t * Abc_ManResubDivs3( Abc_ManRes_t * p, int Required );


// static Vec_Ptr_t *   Abc_CutFactorLarge( Abc_Obj_t * pNode, int nLeavesMax );

// static int           Abc_CutVolumeCheck( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves );


var s_ResubTime abctime;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs incremental resynthesis of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkResubstitute(pNtk *Abc_Ntk_t, nCutMax int32, nStepsMax int32, nLevelsOdc int32, fUpdateLevel int32, fVerbose int32, fVeryVerbose int32) int32 {
    extern void           Dec_GraphUpdateNetwork( Abc_Obj_t * pRoot, Dec_Graph_t * pGraph, int fUpdateLevel, int nGain );
    var pProgress *ProgressBar;
    var pManRes *Abc_ManRes_t;
    var pManCut *Abc_ManCut_t;
    var pManOdc *Odc_Man_t  = nil;
    var pFForm *Dec_Graph_t;
    var vLeaves *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var clk abctime;
    var clkStart abctime = Abc_Clock();
    var i, nNodes int32;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    // cleanup the AIG
    Abc_AigCleanup((Abc_Aig_t *)pNtk.pManFunc);
    // start the managers
    pManCut = Abc_NtkManCutStart( nCutMax, 100000, 100000, 100000 );
    pManRes = Abc_ManResubStart( nCutMax, ABC_RS_DIV1_MAX );
    if nLevelsOdc > 0 {
    pManOdc = Abc_NtkDontCareAlloc( nCutMax, nLevelsOdc, fVerbose, fVeryVerbose ); }

    // compute the reverse levels if level update is requested
    if fUpdateLevel != 0 {
        Abc_NtkStartReverseLevels( pNtk, 0 ); }

    if Abc_NtkLatchNum(pNtk) {
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else {
            pNode.pNext = (Abc_Obj_t *)pNode.pData; }
        }
    }

    // resynthesize each node once
    pManRes.nNodesBeg = Abc_NtkNodeNum(pNtk);
    nNodes = Abc_NtkObjNumMax(pNtk);
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        // skip the constant node
//        if ( Abc_NodeIsConst(pNode) )
//            continue;
        // skip persistant nodes
        if Abc_NodeIsPersistant(pNode) {
            continue; }
        // skip the nodes with many fanouts
        if Abc_ObjFanoutNum(pNode) > 1000 {
            continue; }
        // stop if all nodes have been tried once
        if i >= nNodes {
            break; }

        // compute a reconvergence-driven cut
clk = Abc_Clock();
        vLeaves = Abc_NodeFindCut( pManCut, pNode, 0 );
//        vLeaves = Abc_CutFactorLarge( pNode, nCutMax );
pManRes.timeCut += Abc_Clock() - clk;
/*
        if ( fVerbose && vLeaves )
        printf( "Node %6d : Leaves = %3d. Volume = %3d.\n", pNode->Id, Vec_PtrSize(vLeaves), Abc_CutVolumeCheck(pNode, vLeaves) );
        if ( vLeaves == NULL )
            continue;
*/
        // get the don't-cares
        if pManOdc != nil {
clk = Abc_Clock();
            Abc_NtkDontCareClear( pManOdc );
            Abc_NtkDontCareCompute( pManOdc, pNode, vLeaves, pManRes.pCareSet );
pManRes.timeTruth += Abc_Clock() - clk;
        }

        // evaluate this cut
clk = Abc_Clock();
        pFForm = Abc_ManResubEval( pManRes, pNode, vLeaves, nStepsMax, fUpdateLevel, fVerbose );
//        Vec_PtrFree( vLeaves );
//        Abc_ManResubCleanup( pManRes );
pManRes.timeRes += Abc_Clock() - clk;
        if pFForm == nil {
            continue; }
        pManRes.nTotalGain += pManRes.nLastGain;
/*
        if ( pManRes->nLeaves == 4 && pManRes->nMffc == 2 && pManRes->nLastGain == 1 )
        {
            printf( "%6d :  L = %2d. V = %2d. Mffc = %2d. Divs = %3d.   Up = %3d. Un = %3d. B = %3d.\n", 
                   pNode->Id, pManRes->nLeaves, Abc_CutVolumeCheck(pNode, vLeaves), pManRes->nMffc, pManRes->nDivs, 
                   pManRes->vDivs1UP->nSize, pManRes->vDivs1UN->nSize, pManRes->vDivs1B->nSize );
            Abc_ManResubPrintDivs( pManRes, pNode, vLeaves );
        }
*/
        // acceptable replacement found, update the graph
clk = Abc_Clock();
        Dec_GraphUpdateNetwork( pNode, pFForm, fUpdateLevel, pManRes.nLastGain );
pManRes.timeNtk += Abc_Clock() - clk;
        Dec_GraphFree( pFForm );
    }
    }
    Extra_ProgressBarStop( pProgress );
pManRes.timeTotal = Abc_Clock() - clkStart;
    pManRes.nNodesEnd = Abc_NtkNodeNum(pNtk);

    // print statistics
    if fVerbose != 0 {
    Abc_ManResubPrint( pManRes ); }

    // delete the managers
    Abc_ManResubStop( pManRes );
    Abc_NtkManCutStop( pManCut );
    if pManOdc != nil { Abc_NtkDontCareFree( pManOdc ); }

    // clean the data field
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pNode) == nil {} else {
        pNode.pData = nil; }
    }

    if Abc_NtkLatchNum(pNtk) {
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else {
            pNode.pData = pNode.pNext, pNode.pNext = nil; }
        }
    }

    // put the nodes into the DFS order and reassign their IDs
    Abc_NtkReassignIds( pNtk );
//    Abc_AigCheckFaninOrder( pNtk->pManFunc );
    // fix the levels
    if fUpdateLevel != 0 {
        Abc_NtkStopReverseLevels( pNtk );
    } else {
        Abc_NtkLevel( pNtk ); }
    // check
    if !Abc_NtkCheck( pNtk ) {
        fmt.Printf("Abc_NtkRefactor: The network check has failed.\n");
        return 0;
    }
s_ResubTime = Abc_Clock() - clkStart;
    return 1;
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubStart(nLeavesMax int32, nDivsMax int32) *Abc_ManRes_t {
    var p *Abc_ManRes_t;
    var pData *uint32;
    var i, k int32;
    assert(sizeof(unsigned) == 4, "sizeof(unsigned) == 4");
    // inlining of macro ABC_ALLOC
    p = ((Abc_ManRes_t *) malloc(sizeof(Abc_ManRes_t) * (size_t)(1)));
    *p = Abc_ManRes_t{};
    p.nLeavesMax = nLeavesMax;
    p.nDivsMax   = nDivsMax;
    p.vDivs      = Vec_PtrAlloc( p.nDivsMax );
    // allocate simulation info
    p.nBits      = (1 << p.nLeavesMax);
    if p.nBits <= 32 {
        p.nWords = 1;
    } else {
        p.nWords = (p.nBits / 32);
    }
    // inlining of macro ABC_ALLOC
    p.pInfo      = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nWords * (p.nDivsMax + 1))));
    memset( p.pInfo, 0, sizeof(unsigned) * p.nWords * p.nLeavesMax );
    p.vSims      = Vec_PtrAlloc( p.nDivsMax );
    for  i = 0; i < p.nDivsMax; i++  {
        Vec_PtrPush( p.vSims, p.pInfo + i * p.nWords );
    }
    // assign the care set
    p.pCareSet  = p.pInfo + p.nDivsMax * p.nWords;
    Abc_InfoFill( p.pCareSet, p.nWords );
    // set elementary truth tables
    for  k = 0; k < p.nLeavesMax; k++  {
        pData = (unsigned *)p.vSims.pArray[k];
        for  i = 0; i < p.nBits; i++  {
            if i & (1 << k) {
                pData[i>>5] |= (1 << (i&31)); }
        }
    }
    // create the remaining divisors
    p.vDivs1UP  = Vec_PtrAlloc( p.nDivsMax );
    p.vDivs1UN  = Vec_PtrAlloc( p.nDivsMax );
    p.vDivs1B   = Vec_PtrAlloc( p.nDivsMax );
    p.vDivs2UP0 = Vec_PtrAlloc( p.nDivsMax );
    p.vDivs2UP1 = Vec_PtrAlloc( p.nDivsMax );
    p.vDivs2UN0 = Vec_PtrAlloc( p.nDivsMax );
    p.vDivs2UN1 = Vec_PtrAlloc( p.nDivsMax );
    p.vTemp     = Vec_PtrAlloc( p.nDivsMax );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubStop(p *Abc_ManRes_t) {
    Vec_PtrFree( p.vDivs );
    Vec_PtrFree( p.vSims );
    Vec_PtrFree( p.vDivs1UP );
    Vec_PtrFree( p.vDivs1UN );
    Vec_PtrFree( p.vDivs1B );
    Vec_PtrFree( p.vDivs2UP0 );
    Vec_PtrFree( p.vDivs2UP1 );
    Vec_PtrFree( p.vDivs2UN0 );
    Vec_PtrFree( p.vDivs2UN1 );
    Vec_PtrFree( p.vTemp );
    // inlining of macro ABC_FREE
    ((p.pInfo) ? (free((char *) (p.pInfo)), (p.pInfo) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubPrint(p *Abc_ManRes_t) {
    // inlining of macro ABC_PRT
    fmt.Printf("Used constants    = %6d.             ", p.nUsedNodeC);          Abc_Print(1, "%s =", ("Cuts  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeCut)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    fmt.Printf("Used replacements = %6d.             ", p.nUsedNode0);          Abc_Print(1, "%s =", ("Resub "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeRes)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    fmt.Printf("Used single ORs   = %6d.             ", p.nUsedNode1Or);        Abc_Print(1, "%s =", (" Div  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeDiv)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    fmt.Printf("Used single ANDs  = %6d.             ", p.nUsedNode1And);       Abc_Print(1, "%s =", (" Mffc "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMffc)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    fmt.Printf("Used double ORs   = %6d.             ", p.nUsedNode2Or);        Abc_Print(1, "%s =", (" Sim  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    fmt.Printf("Used double ANDs  = %6d.             ", p.nUsedNode2And);       Abc_Print(1, "%s =", (" 1    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeRes1)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    fmt.Printf("Used OR-AND       = %6d.             ", p.nUsedNode2OrAnd);     Abc_Print(1, "%s =", (" D    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeResD)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    fmt.Printf("Used AND-OR       = %6d.             ", p.nUsedNode2AndOr);     Abc_Print(1, "%s =", (" 2    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeRes2)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    fmt.Printf("Used OR-2ANDs     = %6d.             ", p.nUsedNode3OrAnd);     Abc_Print(1, "%s =", ("Truth "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTruth)))/(float64(int64(time.Second))));; //ABC_PRT( " 3    ", p->timeRes3 );
    // inlining of macro ABC_PRT
    fmt.Printf("Used AND-2ORs     = %6d.             ", p.nUsedNode3AndOr);     Abc_Print(1, "%s =", ("AIG   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeNtk)))/(float64(int64(time.Second))));;
    fmt.Printf("TOTAL             = %6d.             ", p.nUsedNodeC +
                                                         p.nUsedNode0 +
                                                         p.nUsedNode1Or +
                                                         p.nUsedNode1And +
                                                         p.nUsedNode2Or +
                                                         p.nUsedNode2And +
                                                         p.nUsedNode2OrAnd +
                                                         p.nUsedNode2AndOr +
                                                         p.nUsedNode3OrAnd +
                                                         // inlining of macro ABC_PRT
                                                         p.nUsedNode3AndOr);                          Abc_Print(1, "%s =", ("TOTAL "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
    fmt.Printf("Total leaves   = %8d.\n", p.nTotalLeaves);
    fmt.Printf("Total divisors = %8d.\n", p.nTotalDivs);
//    printf( "Total gain     = %8d.\n", p->nTotalGain );
    fmt.Printf("Gain           = %8d. (%6.2f %%).\n", p.nNodesBeg-p.nNodesEnd, 100.0*(p.nNodesBeg-p.nNodesEnd)/p.nNodesBeg);
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubCollectDivs_rec(pNode *Abc_Obj_t, vInternal *Vec_Ptr_t) {
    // skip visited nodes
    if Abc_NodeIsTravIdCurrent(pNode) {
        return; }
    Abc_NodeSetTravIdCurrent(pNode);
    // collect the fanins
    Abc_ManResubCollectDivs_rec( Abc_ObjFanin0(pNode), vInternal );
    Abc_ManResubCollectDivs_rec( Abc_ObjFanin1(pNode), vInternal );
    // collect the internal node
    if pNode.fMarkA == 0 { 
        Vec_PtrPush( vInternal, pNode ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubCollectDivs(p *Abc_ManRes_t, pRoot *Abc_Obj_t, vLeaves *Vec_Ptr_t, Required int32) int32 {
    var pNode, pFanout *Abc_Obj_t;
    var i, k, Limit, Counter int32;

    Vec_PtrClear( p.vDivs1UP );
    Vec_PtrClear( p.vDivs1UN );
    Vec_PtrClear( p.vDivs1B );

    // add the leaves of the cuts to the divisors
    Vec_PtrClear( p.vDivs );
    Abc_NtkIncrementTravId( pRoot.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        Vec_PtrPush( p.vDivs, pNode );
        Abc_NodeSetTravIdCurrent( pNode );        
    }

    // mark nodes in the MFFC
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTemp)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(p.vTemp, i)), 1); i++  {
        pNode.fMarkA = 1;
    }
    // collect the cone (without MFFC)
    Abc_ManResubCollectDivs_rec( pRoot, p.vDivs );
    // unmark the current MFFC
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTemp)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(p.vTemp, i)), 1); i++  {
        pNode.fMarkA = 0;
    }

    // check if the number of divisors is not exceeded
    if Vec_PtrSize(p.vDivs) - Vec_PtrSize(vLeaves) + Vec_PtrSize(p.vTemp) >= Vec_PtrSize(p.vSims) - p.nLeavesMax {
        return 0; }

    // get the number of divisors to collect
    Limit = Vec_PtrSize(p.vSims) - p.nLeavesMax - (Vec_PtrSize(p.vDivs) - Vec_PtrSize(vLeaves) + Vec_PtrSize(p.vTemp));

    // explore the fanouts, which are not in the MFFC
    Counter = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, i)), 1); i++  {
        if Abc_ObjFanoutNum(pNode) > 100 {
//            printf( "%d ", Abc_ObjFanoutNum(pNode) );
            continue;
        }
        // if the fanout has both fanins in the set, add it
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, k)), 1); k++  {
            if Abc_NodeIsTravIdCurrent(pFanout) || Abc_ObjIsCo(pFanout) || int32(pFanout.Level) > Required {
                continue; }
            if Abc_NodeIsTravIdCurrent(Abc_ObjFanin0(pFanout)) && Abc_NodeIsTravIdCurrent(Abc_ObjFanin1(pFanout)) {
                if Abc_ObjFanin0(pFanout) == pRoot || Abc_ObjFanin1(pFanout) == pRoot {
                    continue; }
                Vec_PtrPush( p.vDivs, pFanout );
                Abc_NodeSetTravIdCurrent( pFanout );
                // quit computing divisors if there is too many of them
                if Counter++; Counter == Limit {
                    goto Quits; }
            }
        }
    }

Quits :
    // get the number of divisors
    p.nDivs = Vec_PtrSize(p.vDivs);

    // add the nodes in the MFFC
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTemp)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(p.vTemp, i)), 1); i++  {
        Vec_PtrPush( p.vDivs, pNode );
    }
    assert(pRoot == Vec_PtrEntryLast(p.vDivs), "pRoot == Vec_PtrEntryLast(p.vDivs)");

    assert(Vec_PtrSize(p.vDivs) - Vec_PtrSize(vLeaves) <= Vec_PtrSize(p.vSims) - p.nLeavesMax, "Vec_PtrSize(p.vDivs) - Vec_PtrSize(vLeaves) <= Vec_PtrSize(p.vSims) - p.nLeavesMax");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubPrintDivs(p *Abc_ManRes_t, pRoot *Abc_Obj_t, vLeaves *Vec_Ptr_t) {
    var pFanin, pNode *Abc_Obj_t;
    var i, k int32;
    // print the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, i)), 1); i++  {
        if i < Vec_PtrSize(vLeaves) {
            fmt.Printf("%6d : %c\n", pNode.Id, 'a'+i);
            continue;
        }
        fmt.Printf("%6d : %2d = ", pNode.Id, i);
        // find the first fanin
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(p.vDivs)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, k)), 1); k++  {
            if Abc_ObjFanin0(pNode) == pFanin {
                break; }
        }
        if k < Vec_PtrSize(vLeaves) {
            fmt.Printf("%c", 'a' + k);
        } else {
            fmt.Printf("%d", k); }
        fmt.Printf("%s ", Abc_ObjFaninC0(pNode)? "\'" : "");
        // find the second fanin
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(p.vDivs)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, k)), 1); k++  {
            if Abc_ObjFanin1(pNode) == pFanin {
                break; }
        }
        if k < Vec_PtrSize(vLeaves) {
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

  Synopsis    [Performs simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubSimulate(vDivs *Vec_Ptr_t, nLeaves int32, vSims *Vec_Ptr_t, nLeavesMax int32, nWords int32) {
    var pObj *Abc_Obj_t;
    var puData0, puData1, puData *uint32;
    var i, k int32;
    assert(Vec_PtrSize(vDivs) - nLeaves <= Vec_PtrSize(vSims) - nLeavesMax, "Vec_PtrSize(vDivs) - nLeaves <= Vec_PtrSize(vSims) - nLeavesMax");
    // simulate
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDivs)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vDivs, i)), 1); i++  {
        if i < nLeaves { // initialize the leaf
            pObj.pData = Vec_PtrEntry( vSims, i );
            continue;
        }
        // set storage for the node's simulation info
        pObj.pData = Vec_PtrEntry( vSims, i - nLeaves + nLeavesMax );
        // get pointer to the simulation info
        puData  = (unsigned *)pObj.pData;
        puData0 = (unsigned *)Abc_ObjFanin0(pObj).pData;
        puData1 = (unsigned *)Abc_ObjFanin1(pObj).pData;
        // simulate
        if Abc_ObjFaninC0(pObj) && Abc_ObjFaninC1(pObj) {
            for  k = 0; k < nWords; k++  {
                puData[k] = ~puData0[k] & ~puData1[k];
            }
        } else if Abc_ObjFaninC0(pObj) {
            for  k = 0; k < nWords; k++  {
                puData[k] = ~puData0[k] & puData1[k];
            }
        } else if Abc_ObjFaninC1(pObj) {
            for  k = 0; k < nWords; k++  {
                puData[k] = puData0[k] & ~puData1[k];
            }
        } else { 
            for  k = 0; k < nWords; k++  {
                puData[k] = puData0[k] & puData1[k];
            } }
    }
    // normalize
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDivs)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vDivs, i)), 1); i++  {
        puData = (unsigned *)pObj.pData;
        pObj.fPhase = (puData[0] & 1);
        if pObj.fPhase {
            for  k = 0; k < nWords; k++  {
                puData[k] = ~puData[k];
            } }
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubQuit0(pRoot *Abc_Obj_t, pObj *Abc_Obj_t) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    var eRoot Dec_Edge_t;
    pGraph = Dec_GraphCreate( 1 );
    Dec_GraphNode( pGraph, 0 ).pFunc = pObj;
    eRoot = Dec_EdgeCreate( 0, pObj.fPhase );
    Dec_GraphSetRoot( pGraph, eRoot );
    if pRoot.fPhase {
        Dec_GraphComplement( pGraph ); }
    return pGraph;
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubQuit1(pRoot *Abc_Obj_t, pObj0 *Abc_Obj_t, pObj1 *Abc_Obj_t, fOrGate int32) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    var eRoot, eNode0, eNode1 Dec_Edge_t;
    assert(Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1)");
    pGraph = Dec_GraphCreate( 2 );
    Dec_GraphNode( pGraph, 0 ).pFunc = Abc_ObjRegular(pObj0);
    Dec_GraphNode( pGraph, 1 ).pFunc = Abc_ObjRegular(pObj1);
    eNode0 = Dec_EdgeCreate( 0, Abc_ObjRegular(pObj0).fPhase ^ Abc_ObjIsComplement(pObj0) );
    eNode1 = Dec_EdgeCreate( 1, Abc_ObjRegular(pObj1).fPhase ^ Abc_ObjIsComplement(pObj1) );
    if fOrGate != 0 { 
        eRoot  = Dec_GraphAddNodeOr( pGraph, eNode0, eNode1 );
    } else {
        eRoot  = Dec_GraphAddNodeAnd( pGraph, eNode0, eNode1 ); }
    Dec_GraphSetRoot( pGraph, eRoot );
    if pRoot.fPhase {
        Dec_GraphComplement( pGraph ); }
    return pGraph;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubQuit21(pRoot *Abc_Obj_t, pObj0 *Abc_Obj_t, pObj1 *Abc_Obj_t, pObj2 *Abc_Obj_t, fOrGate int32) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    var eRoot, eNode0, eNode1, eNode2 Dec_Edge_t;
    assert(Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1)");
    assert(Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj2), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj2)");
    assert(Abc_ObjRegular(pObj1) != Abc_ObjRegular(pObj2), "Abc_ObjRegular(pObj1) != Abc_ObjRegular(pObj2)");
    pGraph = Dec_GraphCreate( 3 );
    Dec_GraphNode( pGraph, 0 ).pFunc = Abc_ObjRegular(pObj0);
    Dec_GraphNode( pGraph, 1 ).pFunc = Abc_ObjRegular(pObj1);
    Dec_GraphNode( pGraph, 2 ).pFunc = Abc_ObjRegular(pObj2);
    eNode0 = Dec_EdgeCreate( 0, Abc_ObjRegular(pObj0).fPhase ^ Abc_ObjIsComplement(pObj0) );
    eNode1 = Dec_EdgeCreate( 1, Abc_ObjRegular(pObj1).fPhase ^ Abc_ObjIsComplement(pObj1) );
    eNode2 = Dec_EdgeCreate( 2, Abc_ObjRegular(pObj2).fPhase ^ Abc_ObjIsComplement(pObj2) );
    if fOrGate != 0 {
        eRoot  = Dec_GraphAddNodeOr( pGraph, eNode0, eNode1 );
        eRoot  = Dec_GraphAddNodeOr( pGraph, eNode2, eRoot );
    
    } else {
        eRoot  = Dec_GraphAddNodeAnd( pGraph, eNode0, eNode1 );
        eRoot  = Dec_GraphAddNodeAnd( pGraph, eNode2, eRoot );
    }
    Dec_GraphSetRoot( pGraph, eRoot );
    if pRoot.fPhase {
        Dec_GraphComplement( pGraph ); }
    return pGraph;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubQuit2(pRoot *Abc_Obj_t, pObj0 *Abc_Obj_t, pObj1 *Abc_Obj_t, pObj2 *Abc_Obj_t, fOrGate int32) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    var eRoot, ePrev, eNode0, eNode1, eNode2 Dec_Edge_t;
    assert(Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1)");
    assert(Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj2), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj2)");
    assert(Abc_ObjRegular(pObj1) != Abc_ObjRegular(pObj2), "Abc_ObjRegular(pObj1) != Abc_ObjRegular(pObj2)");
    pGraph = Dec_GraphCreate( 3 );
    Dec_GraphNode( pGraph, 0 ).pFunc = Abc_ObjRegular(pObj0);
    Dec_GraphNode( pGraph, 1 ).pFunc = Abc_ObjRegular(pObj1);
    Dec_GraphNode( pGraph, 2 ).pFunc = Abc_ObjRegular(pObj2);
    eNode0 = Dec_EdgeCreate( 0, Abc_ObjRegular(pObj0).fPhase ^ Abc_ObjIsComplement(pObj0) );
    if Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
        eNode1 = Dec_EdgeCreate( 1, Abc_ObjRegular(pObj1).fPhase );
        eNode2 = Dec_EdgeCreate( 2, Abc_ObjRegular(pObj2).fPhase );
        ePrev  = Dec_GraphAddNodeOr( pGraph, eNode1, eNode2 );
    
    } else {
        eNode1 = Dec_EdgeCreate( 1, Abc_ObjRegular(pObj1).fPhase ^ Abc_ObjIsComplement(pObj1) );
        eNode2 = Dec_EdgeCreate( 2, Abc_ObjRegular(pObj2).fPhase ^ Abc_ObjIsComplement(pObj2) );
        ePrev  = Dec_GraphAddNodeAnd( pGraph, eNode1, eNode2 );
    }
    if fOrGate != 0 { 
        eRoot  = Dec_GraphAddNodeOr( pGraph, eNode0, ePrev );
    } else {
        eRoot  = Dec_GraphAddNodeAnd( pGraph, eNode0, ePrev ); }
    Dec_GraphSetRoot( pGraph, eRoot );
    if pRoot.fPhase {
        Dec_GraphComplement( pGraph ); }
    return pGraph;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubQuit3(pRoot *Abc_Obj_t, pObj0 *Abc_Obj_t, pObj1 *Abc_Obj_t, pObj2 *Abc_Obj_t, pObj3 *Abc_Obj_t, fOrGate int32) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    var eRoot, ePrev0, ePrev1, eNode0, eNode1, eNode2, eNode3 Dec_Edge_t;
    assert(Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1)");
    assert(Abc_ObjRegular(pObj2) != Abc_ObjRegular(pObj3), "Abc_ObjRegular(pObj2) != Abc_ObjRegular(pObj3)");
    pGraph = Dec_GraphCreate( 4 );
    Dec_GraphNode( pGraph, 0 ).pFunc = Abc_ObjRegular(pObj0);
    Dec_GraphNode( pGraph, 1 ).pFunc = Abc_ObjRegular(pObj1);
    Dec_GraphNode( pGraph, 2 ).pFunc = Abc_ObjRegular(pObj2);
    Dec_GraphNode( pGraph, 3 ).pFunc = Abc_ObjRegular(pObj3);
    if Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) {
        eNode0 = Dec_EdgeCreate( 0, Abc_ObjRegular(pObj0).fPhase );
        eNode1 = Dec_EdgeCreate( 1, Abc_ObjRegular(pObj1).fPhase );
        ePrev0 = Dec_GraphAddNodeOr( pGraph, eNode0, eNode1 );
        if Abc_ObjIsComplement(pObj2) && Abc_ObjIsComplement(pObj3) {
            eNode2 = Dec_EdgeCreate( 2, Abc_ObjRegular(pObj2).fPhase );
            eNode3 = Dec_EdgeCreate( 3, Abc_ObjRegular(pObj3).fPhase );
            ePrev1 = Dec_GraphAddNodeOr( pGraph, eNode2, eNode3 );
        
        } else {
            eNode2 = Dec_EdgeCreate( 2, Abc_ObjRegular(pObj2).fPhase ^ Abc_ObjIsComplement(pObj2) );
            eNode3 = Dec_EdgeCreate( 3, Abc_ObjRegular(pObj3).fPhase ^ Abc_ObjIsComplement(pObj3) );
            ePrev1 = Dec_GraphAddNodeAnd( pGraph, eNode2, eNode3 );
        }
    
    } else {
        eNode0 = Dec_EdgeCreate( 0, Abc_ObjRegular(pObj0).fPhase ^ Abc_ObjIsComplement(pObj0) );
        eNode1 = Dec_EdgeCreate( 1, Abc_ObjRegular(pObj1).fPhase ^ Abc_ObjIsComplement(pObj1) );
        ePrev0 = Dec_GraphAddNodeAnd( pGraph, eNode0, eNode1 );
        if Abc_ObjIsComplement(pObj2) && Abc_ObjIsComplement(pObj3) {
            eNode2 = Dec_EdgeCreate( 2, Abc_ObjRegular(pObj2).fPhase );
            eNode3 = Dec_EdgeCreate( 3, Abc_ObjRegular(pObj3).fPhase );
            ePrev1 = Dec_GraphAddNodeOr( pGraph, eNode2, eNode3 );
        
        } else {
            eNode2 = Dec_EdgeCreate( 2, Abc_ObjRegular(pObj2).fPhase ^ Abc_ObjIsComplement(pObj2) );
            eNode3 = Dec_EdgeCreate( 3, Abc_ObjRegular(pObj3).fPhase ^ Abc_ObjIsComplement(pObj3) );
            ePrev1 = Dec_GraphAddNodeAnd( pGraph, eNode2, eNode3 );
        }
    }
    if fOrGate != 0 { 
        eRoot = Dec_GraphAddNodeOr( pGraph, ePrev0, ePrev1 );
    } else {
        eRoot = Dec_GraphAddNodeAnd( pGraph, ePrev0, ePrev1 ); }
    Dec_GraphSetRoot( pGraph, eRoot );
    if pRoot.fPhase {
        Dec_GraphComplement( pGraph ); }
    return pGraph;
}




/**Function*************************************************************

  Synopsis    [Derives single-node unate/binate divisors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubDivsS(p *Abc_ManRes_t, Required int32) {
    var fMoreDivs int32  = 1; // bug fix by Siang-Yun Lee
    var pObj *Abc_Obj_t;
    var puData, puDataR *uint32;
    var i, w int32;
    Vec_PtrClear( p.vDivs1UP );
    Vec_PtrClear( p.vDivs1UN );
    Vec_PtrClear( p.vDivs1B );
    puDataR = (unsigned *)p.pRoot.pData;
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < p.nDivs) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, i)), 1); i++  {
        if int32(pObj.Level) > Required - 1 {
            continue; }

        puData = (unsigned *)pObj.pData;
        // check positive containment
        for  w = 0; w < p.nWords; w++  {
//            if ( puData[w] & ~puDataR[w] )
            if puData[w] & ~puDataR[w] & p.pCareSet[w] { // care set
                break; }
        }
        if w == p.nWords {
            Vec_PtrPush( p.vDivs1UP, pObj );
            continue;
        }
        if fMoreDivs != 0 {
            for  w = 0; w < p.nWords; w++  {
    //            if ( ~puData[w] & ~puDataR[w] )
                if ~puData[w] & ~puDataR[w] & p.pCareSet[w] { // care set
                    break; }
            }
            if w == p.nWords {
                Vec_PtrPush( p.vDivs1UP, Abc_ObjNot(pObj) );
                continue;
            }
        }
        // check negative containment
        for  w = 0; w < p.nWords; w++  {
//            if ( ~puData[w] & puDataR[w] )
            if ~puData[w] & puDataR[w] & p.pCareSet[w] { // care set
                break; }
        }
        if w == p.nWords {
            Vec_PtrPush( p.vDivs1UN, pObj );
            continue;
        }
        if fMoreDivs != 0 {
            for  w = 0; w < p.nWords; w++  {
    //            if ( puData[w] & puDataR[w] )
                if puData[w] & puDataR[w] & p.pCareSet[w] { // care set
                    break; }
            }
            if w == p.nWords {
                Vec_PtrPush( p.vDivs1UN, Abc_ObjNot(pObj) );
                continue;
            }
        }
        // add the node to binates
        Vec_PtrPush( p.vDivs1B, pObj );
    }
}

/**Function*************************************************************

  Synopsis    [Derives double-node unate/binate divisors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubDivsD(p *Abc_ManRes_t, Required int32) {
    var pObj0, pObj1 *Abc_Obj_t;
    var puData0, puData1, puDataR *uint32;
    var i, k, w int32;
    Vec_PtrClear( p.vDivs2UP0 );
    Vec_PtrClear( p.vDivs2UP1 );
    Vec_PtrClear( p.vDivs2UN0 );
    Vec_PtrClear( p.vDivs2UN1 );
    puDataR = (unsigned *)p.pRoot.pData;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs1B)) && (((pObj0) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1B, i)), 1); i++  {
        if int32(pObj0.Level) > Required - 2 {
            continue; }

        puData0 = (unsigned *)pObj0.pData;
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = i + 1; (k < Vec_PtrSize(p.vDivs1B)) && (((pObj1) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1B, k)), 1); k++  {
            if int32(pObj1.Level) > Required - 2 {
                continue; }

            puData1 = (unsigned *)pObj1.pData;

            if Vec_PtrSize(p.vDivs2UP0) < ABC_RS_DIV2_MAX {
                // get positive unate divisors
                for  w = 0; w < p.nWords; w++  {
//                    if ( (puData0[w] & puData1[w]) & ~puDataR[w] )
                    if (puData0[w] & puData1[w]) & ~puDataR[w] & p.pCareSet[w] { // care set
                        break; }
                }
                if w == p.nWords {
                    Vec_PtrPush( p.vDivs2UP0, pObj0 );
                    Vec_PtrPush( p.vDivs2UP1, pObj1 );
                }
                for  w = 0; w < p.nWords; w++  {
//                    if ( (~puData0[w] & puData1[w]) & ~puDataR[w] )
                    if (~puData0[w] & puData1[w]) & ~puDataR[w] & p.pCareSet[w] { // care set
                        break; }
                }
                if w == p.nWords {
                    Vec_PtrPush( p.vDivs2UP0, Abc_ObjNot(pObj0) );
                    Vec_PtrPush( p.vDivs2UP1, pObj1 );
                }
                for  w = 0; w < p.nWords; w++  {
//                    if ( (puData0[w] & ~puData1[w]) & ~puDataR[w] )
                    if (puData0[w] & ~puData1[w]) & ~puDataR[w] & p.pCareSet[w] { // care set
                        break; }
                }
                if w == p.nWords {
                    Vec_PtrPush( p.vDivs2UP0, pObj0 );
                    Vec_PtrPush( p.vDivs2UP1, Abc_ObjNot(pObj1) );
                }
                for  w = 0; w < p.nWords; w++  {
//                    if ( (puData0[w] | puData1[w]) & ~puDataR[w] )
                    if (puData0[w] | puData1[w]) & ~puDataR[w] & p.pCareSet[w] { // care set
                        break; }
                }
                if w == p.nWords {
                    Vec_PtrPush( p.vDivs2UP0, Abc_ObjNot(pObj0) );
                    Vec_PtrPush( p.vDivs2UP1, Abc_ObjNot(pObj1) );
                }
            }

            if Vec_PtrSize(p.vDivs2UN0) < ABC_RS_DIV2_MAX {
                // get negative unate divisors
                for  w = 0; w < p.nWords; w++  {
//                    if ( ~(puData0[w] & puData1[w]) & puDataR[w] )
                    if ~(puData0[w] & puData1[w]) & puDataR[w] & p.pCareSet[w] { // care set
                        break; }
                }
                if w == p.nWords {
                    Vec_PtrPush( p.vDivs2UN0, pObj0 );
                    Vec_PtrPush( p.vDivs2UN1, pObj1 );
                }
                for  w = 0; w < p.nWords; w++  {
//                    if ( ~(~puData0[w] & puData1[w]) & puDataR[w] )
                    if ~(~puData0[w] & puData1[w]) & puDataR[w] & p.pCareSet[w] { // care set
                        break; }
                }
                if w == p.nWords {
                    Vec_PtrPush( p.vDivs2UN0, Abc_ObjNot(pObj0) );
                    Vec_PtrPush( p.vDivs2UN1, pObj1 );
                }
                for  w = 0; w < p.nWords; w++  {
//                    if ( ~(puData0[w] & ~puData1[w]) & puDataR[w] )
                    if ~(puData0[w] & ~puData1[w]) & puDataR[w] & p.pCareSet[w] { // care set
                        break; }
                }
                if w == p.nWords {
                    Vec_PtrPush( p.vDivs2UN0, pObj0 );
                    Vec_PtrPush( p.vDivs2UN1, Abc_ObjNot(pObj1) );
                }
                for  w = 0; w < p.nWords; w++  {
//                    if ( ~(puData0[w] | puData1[w]) & puDataR[w] )
                    if ~(puData0[w] | puData1[w]) & puDataR[w] & p.pCareSet[w] { // care set
                        break; }
                }
                if w == p.nWords {
                    Vec_PtrPush( p.vDivs2UN0, Abc_ObjNot(pObj0) );
                    Vec_PtrPush( p.vDivs2UN1, Abc_ObjNot(pObj1) );
                }
            }
        }
    }
//    printf( "%d %d  ", Vec_PtrSize(p->vDivs2UP0), Vec_PtrSize(p->vDivs2UN0) );
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubQuit(p *Abc_ManRes_t) *Dec_Graph_t {
    var pGraph *Dec_Graph_t;
    var upData *uint32;
    var w int32;
    upData = (unsigned *)p.pRoot.pData;
    for  w = 0; w < p.nWords; w++  {
//        if ( upData[w] )
        if upData[w] & p.pCareSet[w] { // care set
            break; }
    }
    if w != p.nWords {
        return nil; }
    // get constant node graph
    if p.pRoot.fPhase {
        pGraph = Dec_GraphCreateConst1();
    } else { 
        pGraph = Dec_GraphCreateConst0(); }
    return pGraph;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubDivs0(p *Abc_ManRes_t) *Dec_Graph_t {
    var pObj *Abc_Obj_t;
    var puData, puDataR *uint32;
    var i, w int32;
    puDataR = (unsigned *)p.pRoot.pData;
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < p.nDivs) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, i)), 1); i++  {
        puData = (unsigned *)pObj.pData;
        for  w = 0; w < p.nWords; w++  {
//            if ( puData[w] != puDataR[w] )
            if (puData[w] ^ puDataR[w]) & p.pCareSet[w] { // care set
                break; }
        }
        if w == p.nWords {
            return Abc_ManResubQuit0( p.pRoot, pObj ); }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubDivs1(p *Abc_ManRes_t, Required int32) *Dec_Graph_t {
    var pObj0, pObj1 *Abc_Obj_t;
    var puData0, puData1, puDataR *uint32;
    var i, k, w int32;
    puDataR = (unsigned *)p.pRoot.pData;
    // check positive unate divisors
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs1UP)) && (((pObj0) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UP, i)), 1); i++  {
        puData0 = (unsigned *)Abc_ObjRegular(pObj0).pData;
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = i + 1; (k < Vec_PtrSize(p.vDivs1UP)) && (((pObj1) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UP, k)), 1); k++  {
            puData1 = (unsigned *)Abc_ObjRegular(pObj1).pData;
            if Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) {
                for  w = 0; w < p.nWords; w++  {
    //                if ( (puData0[w] | puData1[w]) != puDataR[w] )
                    if ((~puData0[w] | ~puData1[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                        break; }
                }
            
            } else if Abc_ObjIsComplement(pObj0) {
                for  w = 0; w < p.nWords; w++  {
    //                if ( (puData0[w] | puData1[w]) != puDataR[w] )
                    if ((~puData0[w] | puData1[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                        break; }
                }
            
            } else if Abc_ObjIsComplement(pObj1) {
                for  w = 0; w < p.nWords; w++  {
    //                if ( (puData0[w] | puData1[w]) != puDataR[w] )
                    if ((puData0[w] | ~puData1[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                        break; }
                }
            
            } else {
                for  w = 0; w < p.nWords; w++  {
    //                if ( (puData0[w] | puData1[w]) != puDataR[w] )
                    if ((puData0[w] | puData1[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                        break; }
                }
            }
            if w == p.nWords {
                p.nUsedNode1Or++;
                return Abc_ManResubQuit1( p.pRoot, pObj0, pObj1, 1 );
            }
        }
    }
    // check negative unate divisors
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs1UN)) && (((pObj0) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UN, i)), 1); i++  {
        puData0 = (unsigned *)Abc_ObjRegular(pObj0).pData;
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = i + 1; (k < Vec_PtrSize(p.vDivs1UN)) && (((pObj1) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UN, k)), 1); k++  {
            puData1 = (unsigned *)Abc_ObjRegular(pObj1).pData;
            if Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) {
                for  w = 0; w < p.nWords; w++  {
    //                if ( (puData0[w] & puData1[w]) != puDataR[w] )
                    if ((~puData0[w] & ~puData1[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                        break; }
                }
            }
            if Abc_ObjIsComplement(pObj0) {
                for  w = 0; w < p.nWords; w++  {
    //                if ( (puData0[w] & puData1[w]) != puDataR[w] )
                    if ((~puData0[w] & puData1[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                        break; }
                }
            }
            if Abc_ObjIsComplement(pObj1) {
                for  w = 0; w < p.nWords; w++  {
    //                if ( (puData0[w] & puData1[w]) != puDataR[w] )
                    if ((puData0[w] & ~puData1[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                        break; }
                }
            
            } else {
                for  w = 0; w < p.nWords; w++  {
    //                if ( (puData0[w] & puData1[w]) != puDataR[w] )
                    if ((puData0[w] & puData1[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                        break; }
                }
            }
            if w == p.nWords {
                p.nUsedNode1And++;
                return Abc_ManResubQuit1( p.pRoot, pObj0, pObj1, 0 );
            }
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubDivs12(p *Abc_ManRes_t, Required int32) *Dec_Graph_t {
    var pObj0 *Abc_Obj_t;
    var pObj1 *Abc_Obj_t;
    var pObj2 *Abc_Obj_t;
    var pObjMax *Abc_Obj_t;
    var pObjMin0 *Abc_Obj_t = nil;
    var pObjMin1 *Abc_Obj_t = nil;
    var puData0, puData1, puData2, puDataR *uint32;
    var i, k, j, w, LevelMax int32;
    puDataR = (unsigned *)p.pRoot.pData;
    // check positive unate divisors
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs1UP)) && (((pObj0) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UP, i)), 1); i++  {
        puData0 = (unsigned *)Abc_ObjRegular(pObj0).pData;
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = i + 1; (k < Vec_PtrSize(p.vDivs1UP)) && (((pObj1) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UP, k)), 1); k++  {
            puData1 = (unsigned *)Abc_ObjRegular(pObj1).pData;
            // inlining of macro Vec_PtrForEachEntryStart
            for  j = k + 1; (j < Vec_PtrSize(p.vDivs1UP)) && (((pObj2) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UP, j)), 1); j++  {
                puData2 = (unsigned *)Abc_ObjRegular(pObj2).pData;
                if Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | puData1[w] | puData2[w]) != puDataR[w] )
                        if ((~puData0[w] | ~puData1[w] | ~puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) && !Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | puData1[w] | puData2[w]) != puDataR[w] )
                        if ((~puData0[w] | ~puData1[w] | puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj0) && !Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | puData1[w] | puData2[w]) != puDataR[w] )
                        if ((~puData0[w] | puData1[w] | ~puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj0) && !Abc_ObjIsComplement(pObj1) && !Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | puData1[w] | puData2[w]) != puDataR[w] )
                        if ((~puData0[w] | puData1[w] | puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if !Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | puData1[w] | puData2[w]) != puDataR[w] )
                        if ((puData0[w] | ~puData1[w] | ~puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if !Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) && !Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | puData1[w] | puData2[w]) != puDataR[w] )
                        if ((puData0[w] | ~puData1[w] | puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if !Abc_ObjIsComplement(pObj0) && !Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | puData1[w] | puData2[w]) != puDataR[w] )
                        if ((puData0[w] | puData1[w] | ~puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if !Abc_ObjIsComplement(pObj0) && !Abc_ObjIsComplement(pObj1) && !Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | puData1[w] | puData2[w]) != puDataR[w] )
                        if ((puData0[w] | puData1[w] | puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else { assert(false, "0"); }
                if w == p.nWords {
                    LevelMax = Abc_MaxInt( Abc_ObjRegular(pObj0).Level, Abc_MaxInt(Abc_ObjRegular(pObj1).Level, Abc_ObjRegular(pObj2).Level) );
                    assert(LevelMax <= Required - 1, "LevelMax <= Required - 1");

                    pObjMax = nil;
                    if int32(Abc_ObjRegular(pObj0).Level) == LevelMax {
                        pObjMax = pObj0;
pObjMin0 = pObj1;
pObjMin1 = pObj2; }
                    if int32(Abc_ObjRegular(pObj1).Level) == LevelMax {
                        if pObjMax != nil { continue; }
                        pObjMax = pObj1;
pObjMin0 = pObj0;
pObjMin1 = pObj2;
                    }
                    if int32(Abc_ObjRegular(pObj2).Level) == LevelMax {
                        if pObjMax != nil { continue; }
                        pObjMax = pObj2;
pObjMin0 = pObj0;
pObjMin1 = pObj1;
                    }

                    p.nUsedNode2Or++;
                    assert((pObjMin0 != nil), "pObjMin0");
                    assert((pObjMin1 != nil), "pObjMin1");
                    return Abc_ManResubQuit21( p.pRoot, pObjMin0, pObjMin1, pObjMax, 1 );
                }
            }
        }
    }
    // check negative unate divisors
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs1UN)) && (((pObj0) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UN, i)), 1); i++  {
        puData0 = (unsigned *)Abc_ObjRegular(pObj0).pData;
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = i + 1; (k < Vec_PtrSize(p.vDivs1UN)) && (((pObj1) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UN, k)), 1); k++  {
            puData1 = (unsigned *)Abc_ObjRegular(pObj1).pData;
            // inlining of macro Vec_PtrForEachEntryStart
            for  j = k + 1; (j < Vec_PtrSize(p.vDivs1UN)) && (((pObj2) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UN, j)), 1); j++  {
                puData2 = (unsigned *)Abc_ObjRegular(pObj2).pData;
                if Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & puData1[w] & puData2[w]) != puDataR[w] )
                        if ((~puData0[w] & ~puData1[w] & ~puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) && !Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & puData1[w] & puData2[w]) != puDataR[w] )
                        if ((~puData0[w] & ~puData1[w] & puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj0) && !Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & puData1[w] & puData2[w]) != puDataR[w] )
                        if ((~puData0[w] & puData1[w] & ~puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj0) && !Abc_ObjIsComplement(pObj1) && !Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & puData1[w] & puData2[w]) != puDataR[w] )
                        if ((~puData0[w] & puData1[w] & puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if !Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & puData1[w] & puData2[w]) != puDataR[w] )
                        if ((puData0[w] & ~puData1[w] & ~puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if !Abc_ObjIsComplement(pObj0) && Abc_ObjIsComplement(pObj1) && !Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & puData1[w] & puData2[w]) != puDataR[w] )
                        if ((puData0[w] & ~puData1[w] & puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if !Abc_ObjIsComplement(pObj0) && !Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & puData1[w] & puData2[w]) != puDataR[w] )
                        if ((puData0[w] & puData1[w] & ~puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if !Abc_ObjIsComplement(pObj0) && !Abc_ObjIsComplement(pObj1) && !Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & puData1[w] & puData2[w]) != puDataR[w] )
                        if ((puData0[w] & puData1[w] & puData2[w]) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else { assert(false, "0"); }
                if w == p.nWords {
                    LevelMax = Abc_MaxInt( Abc_ObjRegular(pObj0).Level, Abc_MaxInt(Abc_ObjRegular(pObj1).Level, Abc_ObjRegular(pObj2).Level) );
                    assert(LevelMax <= Required - 1, "LevelMax <= Required - 1");

                    pObjMax = nil;
                    if int32(Abc_ObjRegular(pObj0).Level) == LevelMax {
                        pObjMax = pObj0;
pObjMin0 = pObj1;
pObjMin1 = pObj2; }
                    if int32(Abc_ObjRegular(pObj1).Level) == LevelMax {
                        if pObjMax != nil { continue; }
                        pObjMax = pObj1;
pObjMin0 = pObj0;
pObjMin1 = pObj2;
                    }
                    if int32(Abc_ObjRegular(pObj2).Level) == LevelMax {
                        if pObjMax != nil { continue; }
                        pObjMax = pObj2;
pObjMin0 = pObj0;
pObjMin1 = pObj1;
                    }

                    p.nUsedNode2And++;
                    assert((pObjMin0 != nil), "pObjMin0");
                    assert((pObjMin1 != nil), "pObjMin1");
                    return Abc_ManResubQuit21( p.pRoot, pObjMin0, pObjMin1, pObjMax, 0 );
                }
            }
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubDivs2(p *Abc_ManRes_t, Required int32) *Dec_Graph_t {
    var pObj0, pObj1, pObj2 *Abc_Obj_t;
    var puData0, puData1, puData2, puDataR *uint32;
    var i, k, w int32;
    puDataR = (unsigned *)p.pRoot.pData;
    // check positive unate divisors
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs1UP)) && (((pObj0) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UP, i)), 1); i++  {
        puData0 = (unsigned *)Abc_ObjRegular(pObj0).pData;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(p.vDivs2UP0)) && (((pObj1) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs2UP0, k)), 1); k++  {
            pObj2 = (Abc_Obj_t *)Vec_PtrEntry( p.vDivs2UP1, k );

            puData1 = (unsigned *)Abc_ObjRegular(pObj1).pData;
            puData2 = (unsigned *)Abc_ObjRegular(pObj2).pData;
            if Abc_ObjIsComplement(pObj0) {
                if Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | (puData1[w] | puData2[w])) != puDataR[w] )
                        if ((~puData0[w] | (puData1[w] | puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj1) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | (~puData1[w] & puData2[w])) != puDataR[w] )
                        if ((~puData0[w] | (~puData1[w] & puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | (puData1[w] & ~puData2[w])) != puDataR[w] )
                        if ((~puData0[w] | (puData1[w] & ~puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | (puData1[w] & puData2[w])) != puDataR[w] )
                        if ((~puData0[w] | (puData1[w] & puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                }
            
            } else {
                if Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | (puData1[w] | puData2[w])) != puDataR[w] )
                        if ((puData0[w] | (puData1[w] | puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj1) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | (~puData1[w] & puData2[w])) != puDataR[w] )
                        if ((puData0[w] | (~puData1[w] & puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | (puData1[w] & ~puData2[w])) != puDataR[w] )
                        if ((puData0[w] | (puData1[w] & ~puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] | (puData1[w] & puData2[w])) != puDataR[w] )
                        if ((puData0[w] | (puData1[w] & puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                }
            }
            if w == p.nWords {
                p.nUsedNode2OrAnd++;
                return Abc_ManResubQuit2( p.pRoot, pObj0, pObj1, pObj2, 1 );
            }
        }
    }
    // check negative unate divisors
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs1UN)) && (((pObj0) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs1UN, i)), 1); i++  {
        puData0 = (unsigned *)Abc_ObjRegular(pObj0).pData;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(p.vDivs2UN0)) && (((pObj1) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs2UN0, k)), 1); k++  {
            pObj2 = (Abc_Obj_t *)Vec_PtrEntry( p.vDivs2UN1, k );

            puData1 = (unsigned *)Abc_ObjRegular(pObj1).pData;
            puData2 = (unsigned *)Abc_ObjRegular(pObj2).pData;
            if Abc_ObjIsComplement(pObj0) {
                if Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & (puData1[w] | puData2[w])) != puDataR[w] )
                        if ((~puData0[w] & (puData1[w] | puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj1) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & (~puData1[w] & puData2[w])) != puDataR[w] )
                        if ((~puData0[w] & (~puData1[w] & puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & (puData1[w] & ~puData2[w])) != puDataR[w] )
                        if ((~puData0[w] & (puData1[w] & ~puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & (puData1[w] & puData2[w])) != puDataR[w] )
                        if ((~puData0[w] & (puData1[w] & puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                }
            
            } else {
                if Abc_ObjIsComplement(pObj1) && Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & (puData1[w] | puData2[w])) != puDataR[w] )
                        if ((puData0[w] & (puData1[w] | puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj1) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & (~puData1[w] & puData2[w])) != puDataR[w] )
                        if ((puData0[w] & (~puData1[w] & puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else if Abc_ObjIsComplement(pObj2) {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & (puData1[w] & ~puData2[w])) != puDataR[w] )
                        if ((puData0[w] & (puData1[w] & ~puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                
                } else {
                    for  w = 0; w < p.nWords; w++  {
    //                    if ( (puData0[w] & (puData1[w] & puData2[w])) != puDataR[w] )
                        if ((puData0[w] & (puData1[w] & puData2[w])) ^ puDataR[w]) & p.pCareSet[w] { // care set
                            break; }
                    }
                }
            }
            if w == p.nWords {
                p.nUsedNode2AndOr++;
                return Abc_ManResubQuit2( p.pRoot, pObj0, pObj1, pObj2, 0 );
            }
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubDivs3(p *Abc_ManRes_t, Required int32) *Dec_Graph_t {
    var pObj0, pObj1, pObj2, pObj3 *Abc_Obj_t;
    var puData0, puData1, puData2, puData3, puDataR *uint32;
    var i, k, w, Flag int32;
    puDataR = (unsigned *)p.pRoot.pData;
    // check positive unate divisors
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs2UP0)) && (((pObj0) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs2UP0, i)), 1); i++  {
        pObj1 = (Abc_Obj_t *)Vec_PtrEntry( p.vDivs2UP1, i );
        puData0 = (unsigned *)Abc_ObjRegular(pObj0).pData;
        puData1 = (unsigned *)Abc_ObjRegular(pObj1).pData;
        Flag = (Abc_ObjIsComplement(pObj0) << 3) | (Abc_ObjIsComplement(pObj1) << 2);

        // inlining of macro Vec_PtrForEachEntryStart
        for  k = i + 1; (k < Vec_PtrSize(p.vDivs2UP0)) && (((pObj2) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs2UP0, k)), 1); k++  {
            pObj3 = (Abc_Obj_t *)Vec_PtrEntry( p.vDivs2UP1, k );
            puData2 = (unsigned *)Abc_ObjRegular(pObj2).pData;
            puData3 = (unsigned *)Abc_ObjRegular(pObj3).pData;

            Flag = (Flag & 12) | (int32(Abc_ObjIsComplement(pObj2)) << 1) | int32(Abc_ObjIsComplement(pObj3));
            assert(Flag < 16, "Flag < 16");
            switch  Flag  {
            case 0: // 0000
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] & puData1[w]) | (puData2[w] & puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] & puData1[w]) | (puData2[w] & puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 1: // 0001
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] & puData1[w]) | (puData2[w] & ~puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] & puData1[w]) | (puData2[w] & ~puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 2: // 0010
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] & puData1[w]) | (~puData2[w] & puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] & puData1[w]) | (~puData2[w] & puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 3: // 0011
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] & puData1[w]) | (puData2[w] | puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] & puData1[w]) | (puData2[w] | puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;

            case 4: // 0100
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] & ~puData1[w]) | (puData2[w] & puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] & ~puData1[w]) | (puData2[w] & puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 5: // 0101
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] & ~puData1[w]) | (puData2[w] & ~puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] & ~puData1[w]) | (puData2[w] & ~puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 6: // 0110
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] & ~puData1[w]) | (~puData2[w] & puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] & ~puData1[w]) | (~puData2[w] & puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 7: // 0111
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] & ~puData1[w]) | (puData2[w] | puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] & ~puData1[w]) | (puData2[w] | puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;

            case 8: // 1000
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((~puData0[w] & puData1[w]) | (puData2[w] & puData3[w])) != puDataR[w] )
                    if ( (((~puData0[w] & puData1[w]) | (puData2[w] & puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 9: // 1001
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((~puData0[w] & puData1[w]) | (puData2[w] & ~puData3[w])) != puDataR[w] )
                    if ( (((~puData0[w] & puData1[w]) | (puData2[w] & ~puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 10: // 1010
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((~puData0[w] & puData1[w]) | (~puData2[w] & puData3[w])) != puDataR[w] )
                    if ( (((~puData0[w] & puData1[w]) | (~puData2[w] & puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 11: // 1011
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((~puData0[w] & puData1[w]) | (puData2[w] | puData3[w])) != puDataR[w] )
                    if ( (((~puData0[w] & puData1[w]) | (puData2[w] | puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;

            case 12: // 1100
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] | puData1[w]) | (puData2[w] & puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] | puData1[w]) | (puData2[w] & puData3[w])) ^ puDataR[w]) & p.pCareSet[w] ) // care set
                        break;
                break;
            case 13: // 1101
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] | puData1[w]) | (puData2[w] & ~puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] | puData1[w]) | (puData2[w] & ~puData3[w])) ^ puDataR[w]) & p.pCareSet[w] )
                        break;
                break;
            case 14: // 1110
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] | puData1[w]) | (~puData2[w] & puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] | puData1[w]) | (~puData2[w] & puData3[w])) ^ puDataR[w]) & p.pCareSet[w] )
                        break;
                break;
            case 15: // 1111
                for ( w = 0; w < p.nWords; w++ )
//                    if ( ((puData0[w] | puData1[w]) | (puData2[w] | puData3[w])) != puDataR[w] )
                    if ( (((puData0[w] | puData1[w]) | (puData2[w] | puData3[w])) ^ puDataR[w]) & p.pCareSet[w] )
                        break;
                break;

            }
            if w == p.nWords {
                p.nUsedNode3OrAnd++;
                return Abc_ManResubQuit3( p.pRoot, pObj0, pObj1, pObj2, pObj3, 1 );
            }
        }
    }

/*
    // check negative unate divisors
    Vec_PtrForEachEntry( Abc_Obj_t *, p->vDivs2UN0, pObj0, i )
    {
        pObj1 = Vec_PtrEntry( p->vDivs2UN1, i );
        puData0 = Abc_ObjRegular(pObj0)->pData;
        puData1 = Abc_ObjRegular(pObj1)->pData;
        Flag = (Abc_ObjIsComplement(pObj0) << 3) | (Abc_ObjIsComplement(pObj1) << 2);

        Vec_PtrForEachEntryStart( Abc_Obj_t *, p->vDivs2UN0, pObj2, k, i + 1 )
        {
            pObj3 = Vec_PtrEntry( p->vDivs2UN1, k );
            puData2 = Abc_ObjRegular(pObj2)->pData;
            puData3 = Abc_ObjRegular(pObj3)->pData;

            Flag = (Flag & 12) | (Abc_ObjIsComplement(pObj2) << 1) | Abc_ObjIsComplement(pObj3);
            assert( Flag < 16 );
            switch( Flag )
            {
            case 0: // 0000
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] & puData1[w]) & (puData2[w] & puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 1: // 0001
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] & puData1[w]) & (puData2[w] & ~puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 2: // 0010
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] & puData1[w]) & (~puData2[w] & puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 3: // 0011
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] & puData1[w]) & (puData2[w] | puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;

            case 4: // 0100
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] & ~puData1[w]) & (puData2[w] & puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 5: // 0101
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] & ~puData1[w]) & (puData2[w] & ~puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 6: // 0110
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] & ~puData1[w]) & (~puData2[w] & puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 7: // 0111
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] & ~puData1[w]) & (puData2[w] | puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;

            case 8: // 1000
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((~puData0[w] & puData1[w]) & (puData2[w] & puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 9: // 1001
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((~puData0[w] & puData1[w]) & (puData2[w] & ~puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 10: // 1010
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((~puData0[w] & puData1[w]) & (~puData2[w] & puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 11: // 1011
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((~puData0[w] & puData1[w]) & (puData2[w] | puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;

            case 12: // 1100
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] | puData1[w]) & (puData2[w] & puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 13: // 1101
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] | puData1[w]) & (puData2[w] & ~puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 14: // 1110
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] | puData1[w]) & (~puData2[w] & puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;
            case 15: // 1111
                for ( w = 0; w < p->nWords; w++ )
                    if ( (((puData0[w] | puData1[w]) & (puData2[w] | puData3[w])) ^ puDataR[w]) & p->pCareSet[w] )
                        break;
                break;

            }
            if ( w == p->nWords )
            {
                p->nUsedNode3AndOr++;
                return Abc_ManResubQuit3( p->pRoot, pObj0, pObj1, pObj2, pObj3, 0 );
            }
        }
    }
*/
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubCleanup(p *Abc_ManRes_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vDivs)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, i)), 1); i++  {
        pObj.pData = nil;
    }
    Vec_PtrClear( p.vDivs );
    p.pRoot = nil;
}

/**Function*************************************************************

  Synopsis    [Evaluates resubstution of one cut.]

  Description [Returns the graph to add if any.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManResubEval(p *Abc_ManRes_t, pRoot *Abc_Obj_t, vLeaves *Vec_Ptr_t, nSteps int32, fUpdateLevel int32, fVerbose int32) *Dec_Graph_t {
    extern int Abc_NodeMffcInside( Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vInside );
    var pGraph *Dec_Graph_t;
    var Required int32;
    var clk abctime;

    if fUpdateLevel != 0 {
        Required = Abc_ObjRequiredLevel(pRoot);
    } else {
        Required = ABC_INFINITY;
    }

    assert(nSteps >= 0, "nSteps >= 0");
    assert(nSteps <= 3, "nSteps <= 3");
    p.pRoot = pRoot;
    p.nLeaves = Vec_PtrSize(vLeaves);
    p.nLastGain = -1;

    // collect the MFFC
clk = Abc_Clock();
    p.nMffc = Abc_NodeMffcInside( pRoot, vLeaves, p.vTemp );
p.timeMffc += Abc_Clock() - clk;
    assert(p.nMffc > 0, "p.nMffc > 0");

    // collect the divisor nodes
clk = Abc_Clock();
    if !(Abc_ManResubCollectDivs( p, pRoot, vLeaves, Required ) != 0) {
        return nil; }
    p.timeDiv += Abc_Clock() - clk;

    p.nTotalDivs   += p.nDivs;
    p.nTotalLeaves += p.nLeaves;

    // simulate the nodes
clk = Abc_Clock();
    Abc_ManResubSimulate( p.vDivs, p.nLeaves, p.vSims, p.nLeavesMax, p.nWords );
p.timeSim += Abc_Clock() - clk;

clk = Abc_Clock();
    // consider constants
    if pGraph = Abc_ManResubQuit( p ) {
        p.nUsedNodeC++;
        p.nLastGain = p.nMffc;
        return pGraph;
    }

    // consider equal nodes
    if pGraph = Abc_ManResubDivs0( p ) {
p.timeRes1 += Abc_Clock() - clk;
        p.nUsedNode0++;
        p.nLastGain = p.nMffc;
        return pGraph;
    }
    if nSteps == 0 || p.nMffc == 1 {
p.timeRes1 += Abc_Clock() - clk;
        return nil;
    }

    // get the one level divisors
    Abc_ManResubDivsS( p, Required );

    // consider one node
    if pGraph = Abc_ManResubDivs1( p, Required ) {
p.timeRes1 += Abc_Clock() - clk;
        p.nLastGain = p.nMffc - 1;
        return pGraph;
    }
p.timeRes1 += Abc_Clock() - clk;
    if nSteps == 1 || p.nMffc == 2 {
        return nil; }

clk = Abc_Clock();
    // consider triples
    if pGraph = Abc_ManResubDivs12( p, Required ) {
p.timeRes2 += Abc_Clock() - clk;
        p.nLastGain = p.nMffc - 2;
        return pGraph;
    }
p.timeRes2 += Abc_Clock() - clk;

    // get the two level divisors
clk = Abc_Clock();
    Abc_ManResubDivsD( p, Required );
p.timeResD += Abc_Clock() - clk;

    // consider two nodes
clk = Abc_Clock();
    if pGraph = Abc_ManResubDivs2( p, Required ) {
p.timeRes2 += Abc_Clock() - clk;
        p.nLastGain = p.nMffc - 2;
        return pGraph;
    }
p.timeRes2 += Abc_Clock() - clk;
    if nSteps == 2 || p.nMffc == 3 {
        return nil; }

    // consider two nodes
clk = Abc_Clock();
    if pGraph = Abc_ManResubDivs3( p, Required ) {
p.timeRes3 += Abc_Clock() - clk;
        p.nLastGain = p.nMffc - 3;
        return pGraph;
    }
p.timeRes3 += Abc_Clock() - clk;
    if nSteps == 3 || p.nLeavesMax == 4 {
        return nil; }
    return nil;
}




/**Function*************************************************************

  Synopsis    [Computes the volume and checks if the cut is feasible.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CutVolumeCheck_rec(pObj *Abc_Obj_t) int32 {
    // quit if the node is visited (or if it is a leaf)
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 0; }
    Abc_NodeSetTravIdCurrent(pObj);
    // report the error
    if Abc_ObjIsCi(pObj) {
        fmt.Printf("Abc_CutVolumeCheck() ERROR: The set of nodes is not a cut!\n"); }
    // count the number of nodes in the leaves
    return 1 + Abc_CutVolumeCheck_rec( Abc_ObjFanin0(pObj) ) +
        Abc_CutVolumeCheck_rec( Abc_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Computes the volume and checks if the cut is feasible.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CutVolumeCheck(pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t) int32 {
    var pObj *Abc_Obj_t;
    var i int32;
    // mark the leaves
    Abc_NtkIncrementTravId( pNode.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
    } 
    // traverse the nodes starting from the given one and count them
    return Abc_CutVolumeCheck_rec( pNode );
}

/**Function*************************************************************

  Synopsis    [Computes the factor cut of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CutFactor_rec(pObj *Abc_Obj_t, vLeaves *Vec_Ptr_t) {
    if pObj.fMarkA {
        return; }
    if Abc_ObjIsCi(pObj) || (Abc_ObjFanoutNum(pObj) > 1 && !Abc_NodeIsMuxControlType(pObj)) {
        Vec_PtrPush( vLeaves, pObj );
        pObj.fMarkA = 1;
        return;
    }
    Abc_CutFactor_rec( Abc_ObjFanin0(pObj), vLeaves );
    Abc_CutFactor_rec( Abc_ObjFanin1(pObj), vLeaves );
}

/**Function*************************************************************

  Synopsis    [Computes the factor cut of the node.]

  Description [Factor-cut is the cut at a node in terms of factor-nodes.
  Factor-nodes are roots of the node trees (MUXes/EXORs are counted as single nodes).
  Factor-cut is unique for the given node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CutFactor(pNode *Abc_Obj_t) *Vec_Ptr_t {
    var vLeaves *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(!Abc_ObjIsCi(pNode), "!Abc_ObjIsCi(pNode)");
    vLeaves  = Vec_PtrAlloc( 10 );
    Abc_CutFactor_rec( Abc_ObjFanin0(pNode), vLeaves );
    Abc_CutFactor_rec( Abc_ObjFanin1(pNode), vLeaves );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.fMarkA = 0;
    }
    return vLeaves;
}

/**Function*************************************************************

  Synopsis    [Cut computation.]

  Description [This cut computation works as follows: 
  It starts with the factor cut at the node. If the factor-cut is large, quit.
  It supports the set of leaves of the cut under construction and labels all nodes
  in the cut under construction, including the leaves.
  It computes the factor-cuts of the leaves and checks if it is easible to add any of them.
  If it is, it randomly chooses one feasible and continues.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CutFactorLarge(pNode *Abc_Obj_t, nLeavesMax int32) *Vec_Ptr_t {
    var vLeaves, vFactors, vFact, vNext *Vec_Ptr_t;
    var vFeasible *Vec_Int_t;
    var pLeaf, pTemp *Abc_Obj_t;
    var i, k, Counter, RandLeaf int32;
    var BestCut, BestShare int32;
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // get one factor-cut
    vLeaves = Abc_CutFactor( pNode );
    if Vec_PtrSize(vLeaves) > nLeavesMax {
        Vec_PtrFree(vLeaves);
        return nil;
    }
    if Vec_PtrSize(vLeaves) == nLeavesMax {
        return vLeaves; }
    // initialize the factor cuts for the leaves
    vFactors = Vec_PtrAlloc( nLeavesMax );
    Abc_NtkIncrementTravId( pNode.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pLeaf) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pLeaf ); 
        if Abc_ObjIsCi(pLeaf) {
            Vec_PtrPush( vFactors, nil );
        } else {
            Vec_PtrPush( vFactors, Abc_CutFactor(pLeaf) ); }
    }
    // construct larger factor cuts
    vFeasible = Vec_IntAlloc( nLeavesMax );
    for {
        BestCut = -1;
BestShare = -1;
        // find the next feasible cut to add
        Vec_IntClear( vFeasible );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vFactors)) && (((vFact) = (Vec_Ptr_t *)Vec_PtrEntry(vFactors, i)), 1); i++  {
            if vFact == nil {
                continue; }
            // count the number of unmarked leaves of this factor cut
            Counter = 0;
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFact)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(vFact, k)), 1); k++  {
                Counter += !Abc_NodeIsTravIdCurrent(pTemp);
            }
            // if the number of new leaves is smaller than the diff, it is feasible
            if Counter <= nLeavesMax - Vec_PtrSize(vLeaves) + 1 {
                Vec_IntPush( vFeasible, i );
                if BestCut == -1 || BestShare < Vec_PtrSize(vFact) - Counter {
                    BestCut = i;
BestShare = Vec_PtrSize(vFact) - Counter; }
            }
        }
        // quit if there is no feasible factor cuts
        if Vec_IntSize(vFeasible) == 0 {
            break; }
        // randomly choose one leaf and get its factor cut
//        RandLeaf = Vec_IntEntry( vFeasible, rand() % Vec_IntSize(vFeasible) );
        // choose the cut that has most sharing with the other cuts
        RandLeaf = BestCut;

        pLeaf = (Abc_Obj_t *)Vec_PtrEntry( vLeaves, RandLeaf );
        vNext = (Vec_Ptr_t *)Vec_PtrEntry( vFactors, RandLeaf );
        // unmark this leaf
        Abc_NodeSetTravIdPrevious( pLeaf ); 
        // remove this cut from the leaves and factor cuts
        for  i = RandLeaf; i < Vec_PtrSize(vLeaves)-1; i++  {
            Vec_PtrWriteEntry( vLeaves,  i, Vec_PtrEntry(vLeaves, i+1) );
            Vec_PtrWriteEntry( vFactors, i, Vec_PtrEntry(vFactors,i+1) );
        }
        Vec_PtrShrink( vLeaves,  Vec_PtrSize(vLeaves) -1 );
        Vec_PtrShrink( vFactors, Vec_PtrSize(vFactors)-1 );
        // add new leaves, compute their factor cuts
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNext)) && (((pLeaf) = (Abc_Obj_t *)Vec_PtrEntry(vNext, i)), 1); i++  {
            if Abc_NodeIsTravIdCurrent(pLeaf) {
                continue; }
            Abc_NodeSetTravIdCurrent( pLeaf ); 
            Vec_PtrPush( vLeaves, pLeaf );
            if Abc_ObjIsCi(pLeaf) {
                Vec_PtrPush( vFactors, nil );
            } else {
                Vec_PtrPush( vFactors, Abc_CutFactor(pLeaf) ); }
        }
        Vec_PtrFree( vNext );
        assert(Vec_PtrSize(vLeaves) <= nLeavesMax, "Vec_PtrSize(vLeaves) <= nLeavesMax");
        if Vec_PtrSize(vLeaves) == nLeavesMax {
            break; }
    }

    // remove temporary storage
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFactors)) && (((vFact) = (Vec_Ptr_t *)Vec_PtrEntry(vFactors, i)), 1); i++  {
        if vFact { Vec_PtrFree( vFact ); }
    }
    Vec_PtrFree( vFactors );
    Vec_IntFree( vFeasible );
    return vLeaves;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

