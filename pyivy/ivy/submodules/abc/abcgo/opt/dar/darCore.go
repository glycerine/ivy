package dar

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [darCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [Core of the rewriting package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: darCore.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "darInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// iterator over the nodes in the topological order
// Original macro:
// #define Aig_ManForEachNodeInOrder( p, pObj )                                    \
//     for ( assert(p->pOrderData), p->iPrev = 0, p->iNext = p->pOrderData[1];     \
//           p->iNext && (((pObj) = Aig_ManObj(p, p->iNext)), 1);                  \
//           p->iNext = p->pOrderData[2*p->iPrev+1] )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the structure with default assignment of parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManDefaultRwrParams(pPars *Dar_RwrPar_t) {
    memset( pPars, 0, sizeof(Dar_RwrPar_t) );
    pPars.nCutsMax     =  8; // 8
    pPars.nSubgMax     =  5; // 5 is a "magic number"
    pPars.fFanout      =  1;
    pPars.fUpdateLevel =  0;
    pPars.fUseZeros    =  0;
    pPars.fPower       =  0;
    pPars.fRecycle     =  1;
    pPars.fVerbose     =  0;
    pPars.fVeryVerbose =  0;
}

const MAX_VAL = 10

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManRewrite(pAig *Aig_Man_t, pPars *Dar_RwrPar_t) int32 {
    extern Vec_Int_t * Saig_ManComputeSwitchProbs( Aig_Man_t * p, int nFrames, int nPref, int fProbOne );
    var p *Dar_Man_t;
//    Bar_Progress_t * pProgress;
    var pCut *Dar_Cut_t;
    var pObj, pObjNew *Aig_Obj_t;
    var i, k, nNodesOld, nNodeBefore, nNodeAfter, Required int32;
    var clk, clkStart abctime;
    var Counter int32;
    var nMffcSize int32;//, nMffcGains[MAX_VAL+1][MAX_VAL+1] = {{0}};
    // prepare the library
    Dar_LibPrepare( pPars.nSubgMax ); 
    // create rewriting manager
    p = Dar_ManStart( pAig, pPars );
    if pPars.fPower {
        pAig.vProbs = Saig_ManComputeSwitchProbs( pAig, 48, 16, 1 ); }
    // remove dangling nodes
    Aig_ManCleanup( pAig );
    // if updating levels is requested, start fanout and timing
    if p.pPars.fFanout {
        Aig_ManFanoutStart( pAig ); }
    if p.pPars.fUpdateLevel {
        Aig_ManStartReverseLevels( pAig, 0 ); }
    // set elementary cuts for the PIs
//    Dar_ManCutsStart( p );
    // resynthesize each node once
    clkStart = Abc_Clock();
    p.nNodesInit = Aig_ManNodeNum(pAig);
    nNodesOld = Vec_PtrSize( pAig.vObjs );

//    pProgress = Bar_ProgressStart( stdout, nNodesOld );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
//    pProgress = Bar_ProgressStart( stdout, 100 );
//    Aig_ManOrderStart( pAig );
//    Aig_ManForEachNodeInOrder( pAig, pObj )
    
        if pAig.Time2Quit && !(i & 256) && Abc_Clock() > pAig.Time2Quit {
            break; }
//        Bar_ProgressUpdate( pProgress, 100*pAig->nAndPrev/pAig->nAndTotal, NULL );
//        Bar_ProgressUpdate( pProgress, i, NULL );
        if !Aig_ObjIsNode(pObj) {
            continue; }
        if i > nNodesOld {
//        if ( p->pPars->fUseZeros && i > nNodesOld )
            break; }
        if pPars.fRecycle && ++Counter % 50000 == 0 && Aig_DagSize(pObj) < Vec_PtrSize(p.vCutNodes)/100 {
//            printf( "Counter = %7d.  Node = %7d.  Dag = %5d. Vec = %5d.\n", 
//                Counter, i, Aig_DagSize(pObj), Vec_PtrSize(p->vCutNodes) );
//            fflush( stdout );
            Dar_ManCutsRestart( p, pObj );
        }

        // consider freeing the cuts
//        if ( (i & 0xFFF) == 0 && Aig_MmFixedReadMemUsage(p->pMemCuts)/(1<<20) > 100 )
//            Dar_ManCutsStart( p );

        // compute cuts for the node
        p.nNodesTried++;
clk = Abc_Clock();
        Dar_ObjSetCuts( pObj, nil );
        Dar_ObjComputeCuts_rec( p, pObj );
p.timeCuts += Abc_Clock() - clk;

        // check if there is a trivial cut
        // inlining of macro Dar_ObjForEachCut
        for  (pCut) = Dar_ObjCuts(pObj), k = 0; k < int32((pObj).nCuts); k, pCut = k+1, pCut+1  { if (pCut).fUsed==0 {} else
            if pCut.nLeaves == 0 || (pCut.nLeaves == 1 && pCut.pLeaves[0] != pObj.Id && Aig_ManObj(p.pAig, pCut.pLeaves[0])) {
                break; }
        }
        if k < int32(pObj.nCuts) {
            assert(pCut.nLeaves < 2, "pCut.nLeaves < 2");
            if pCut.nLeaves == 0 { // replace by constant
            
                assert(pCut.uTruth == 0 || pCut.uTruth == 0xFFFF, "pCut.uTruth == 0 || pCut.uTruth == 0xFFFF");
                pObjNew = Aig_NotCond( Aig_ManConst1(p.pAig), pCut.uTruth==0 );
            
            } else {
                assert(pCut.uTruth == 0xAAAA || pCut.uTruth == 0x5555, "pCut.uTruth == 0xAAAA || pCut.uTruth == 0x5555");
                pObjNew = Aig_NotCond( Aig_ManObj(p.pAig, pCut.pLeaves[0]), pCut.uTruth==0x5555 );
            }
            // remove the old cuts
            Dar_ObjSetCuts( pObj, nil );
            // replace the node
            Aig_ObjReplace( pAig, pObj, pObjNew, p.pPars.fUpdateLevel );
            continue;
        }

        // evaluate the cuts
        p.GainBest = -1;
        nMffcSize   = -1;
        if pAig.vLevelR {
            Required = Aig_ObjRequiredLevel(pAig, pObj);
        } else {
            Required = ABC_INFINITY;
        }
        // inlining of macro Dar_ObjForEachCut
        for  (pCut) = Dar_ObjCuts(pObj), k = 0; k < int32((pObj).nCuts); k, pCut = k+1, pCut+1  { if (pCut).fUsed==0 {} else {
            var nLeavesOld int32  = pCut.nLeaves;
            if pCut.nLeaves == 3 {
                pCut.pLeaves[pCut.nLeaves++] = 0; }
            Dar_LibEval( p, pObj, pCut, Required, &nMffcSize );
            pCut.nLeaves = nLeavesOld; 
        }
        }
        // check the best gain
        if !(p.GainBest > 0 || (p.GainBest == 0 && p.pPars.fUseZeros)) {
//            Aig_ObjOrderAdvance( pAig );
            continue;
        }
//        nMffcGains[p->GainBest < MAX_VAL ? p->GainBest : MAX_VAL][nMffcSize < MAX_VAL ? nMffcSize : MAX_VAL]++;
        // remove the old cuts
        Dar_ObjSetCuts( pObj, nil );
        // if we end up here, a rewriting step is accepted
        nNodeBefore = Aig_ManNodeNum( pAig );
        pObjNew = Dar_LibBuildBest( p ); // pObjNew can be complemented!
        pObjNew = Aig_NotCond( pObjNew, Aig_ObjPhaseReal(pObjNew) ^ pObj.fPhase );
        assert(int32(Aig_Regular(pObjNew).Level) <= Required, "int32(Aig_Regular(pObjNew).Level) <= Required");
        // replace the node
        Aig_ObjReplace( pAig, pObj, pObjNew, p.pPars.fUpdateLevel );
        // compare the gains
        nNodeAfter = Aig_ManNodeNum( pAig );
        assert(p.GainBest <= nNodeBefore - nNodeAfter, "p.GainBest <= nNodeBefore - nNodeAfter");
        // count gains of this class
        p.ClassGains[p.ClassBest] += nNodeBefore - nNodeAfter;
    }
    }
//    Aig_ManOrderStop( pAig );
/*
    printf( "Distribution of gain (row) by MFFC size (column) %s 0-costs:\n", p->pPars->fUseZeros? "with":"without" );
    for ( k = 0; k <= MAX_VAL; k++ )
        printf( "<%4d> ", k );
    printf( "\n" );
    for ( i = 0; i <= MAX_VAL; i++ )
    {
        for ( k = 0; k <= MAX_VAL; k++ )
            printf( "%6d ", nMffcGains[i][k] );
        printf( "\n" );
    }
*/

p.timeTotal = Abc_Clock() - clkStart;
p.timeOther = p.timeTotal - p.timeCuts - p.timeEval;

//    Bar_ProgressStop( pProgress );
    Dar_ManCutsFree( p );
    // put the nodes into the DFS order and reassign their IDs
//    Aig_NtkReassignIds( p );
    // fix the levels
//    Aig_ManVerifyLevel( pAig );
    if p.pPars.fFanout {
        Aig_ManFanoutStop( pAig ); }
    if p.pPars.fUpdateLevel {
//        Aig_ManVerifyReverseLevel( pAig );
        Aig_ManStopReverseLevels( pAig );
    }
    if pAig.vProbs {
        Vec_IntFree( pAig.vProbs );
        pAig.vProbs = nil;
    }
    // stop the rewriting manager
    Dar_ManStop( p );
    Aig_ManCheckPhase( pAig );
    // check
    if !Aig_ManCheck( pAig ) {
        fmt.Printf("Aig_ManRewrite: The network check has failed.\n");
        return 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes the total number of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManCutCount(pAig *Aig_Man_t, pnCutsK *int32) int32 {
    var pCut *Dar_Cut_t;
    var pObj *Aig_Obj_t;
    var i, k, nCuts, nCutsK int32;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        // inlining of macro Dar_ObjForEachCut
        for  (pCut) = Dar_ObjCuts(pObj), k = 0; k < int32((pObj).nCuts); k, pCut = k+1, pCut+1  { if (pCut).fUsed==0 {} else {
            nCuts++;
            if pCut.nLeaves == 4 {
                nCutsK++; }
        }
        } }
    }
    if pnCutsK != nil {
        *pnCutsK = nCutsK; }
    return nCuts;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManComputeCuts(pAig *Aig_Man_t, nCutsMax int32, fSkipTtMin int32, fVerbose int32) *Aig_MmFixed_t { 
    var p *Dar_Man_t;
    var Pars Dar_RwrPar_t;
    var pPars *Dar_RwrPar_t = &Pars; 
    var pObj *Aig_Obj_t;
    var pMemCuts *Aig_MmFixed_t;
    var i, nNodes int32;
    var clk abctime  = Abc_Clock();
    // remove dangling nodes
    if nNodes = Aig_ManCleanup( pAig ) {
//        printf( "Removing %d nodes.\n", nNodes );
    }
    // create default parameters
    Dar_ManDefaultRwrParams( pPars );
    pPars.nCutsMax = nCutsMax;
    // create rewriting manager
    p = Dar_ManStart( pAig, pPars );
    // set elementary cuts for the PIs
//    Dar_ManCutsStart( p );
    Aig_MmFixedRestart( p.pMemCuts );
    Dar_ObjPrepareCuts( p, Aig_ManConst1(p.pAig) );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        Dar_ObjPrepareCuts( p, pObj );
    }
    // compute cuts for each nodes in the topological order
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Dar_ObjComputeCuts( p, pObj, fSkipTtMin ); }
    }
    // print verbose stats
    if fVerbose != 0 {
//        Aig_Obj_t * pObj;
        var nCuts, nCutsK int32;//, i;
        nCuts = Dar_ManCutCount( pAig, &nCutsK );
        fmt.Printf("Nodes = %6d. Total cuts = %6d. 4-input cuts = %6d.\n", Aig_ManObjNum(pAig), nCuts, nCutsK);
        fmt.Printf("Cut size = %2d. Truth size = %2d. Total mem = %5.2f MB  ", int32(sizeof(Dar_Cut_t)), int32(4), 1.0*Aig_MmFixedReadMemUsage(p.pMemCuts)/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
/*
        Aig_ManForEachNode( pAig, pObj, i )
            if ( i % 300 == 0 )
                Dar_ObjCutPrint( pAig, pObj );
*/
    }
    // free the cuts
    pMemCuts = p.pMemCuts;
    p.pMemCuts = nil;
//    Dar_ManCutsFree( p );
    // stop the rewriting manager
    Dar_ManStop( p );
    return pMemCuts;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

