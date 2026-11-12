package rwr

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [rwrDec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [Evaluation and decomposition procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwrDec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwr.h"
//#include "bool/dec/dec.h"
//#include "aig/ivy/ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Dec_Graph_t * Rwr_CutEvaluate( Rwr_Man_t * p, Abc_Obj_t * pRoot, Cut_Cut_t * pCut, Vec_Ptr_t * vFaninsCur, int nNodesSaved, int LevelMax, int * pGainBest, int fPlaceEnable );

// static int Rwr_CutIsBoolean( Abc_Obj_t * pObj, Vec_Ptr_t * vLeaves );

// static int Rwr_CutCountNumNodes( Abc_Obj_t * pObj, Cut_Cut_t * pCut );

// static int Rwr_NodeGetDepth_rec( Abc_Obj_t * pObj, Vec_Ptr_t * vLeaves );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs rewriting for one node.]

  Description [This procedure considers all the cuts computed for the node
  and tries to rewrite each of them using the "forest" of different AIG
  structures precomputed and stored in the RWR manager. 
  Determines the best rewriting and computes the gain in the number of AIG
  nodes in the final network. In the end, p->vFanins contains information 
  about the best cut that can be used for rewriting, while p->pGraph gives 
  the decomposition dag (represented using decomposition graph data structure).
  Returns gain in the number of nodes or -1 if node cannot be rewritten.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_NodeRewrite(p *Rwr_Man_t, pManCut *Cut_Man_t, pNode *Abc_Obj_t, fUpdateLevel int32, fUseZeros int32, fPlaceEnable int32) int32 {
    var fVeryVerbose int32;
    var pGraph *Dec_Graph_t;
    var pCut *Cut_Cut_t;//, * pTemp;
    var pFanin *Abc_Obj_t;
    var uPhase uint32;
    var uTruthBest uint32; // Suppress "might be used uninitialized"
    var uTruth uint32;
    var pPerm string;
    var Required, nNodesSaved int32;
    var nNodesSaveCur int32  = -1; // Suppress "might be used uninitialized"
    var i int32;
    var GainCur int32 = -1;
    var GainBest int32 = -1;
    var clk, clk2 abctime;//, Counter;

    p.nNodesConsidered++;
    // get the required times
    if fUpdateLevel != 0 {
        Required = Abc_ObjRequiredLevel(pNode);
    } else {
        Required = ABC_INFINITY;
    }

    // get the node's cuts
clk = Abc_Clock();
    pCut = (Cut_Cut_t *)Abc_NodeGetCutsRecursive( pManCut, pNode, 0, 0 );
    assert(pCut != nil, "pCut != nil");
p.timeCut += Abc_Clock() - clk;

//printf( " %d", Rwr_CutCountNumNodes(pNode, pCut) );
/*
    Counter = 0;
    for ( pTemp = pCut->pNext; pTemp; pTemp = pTemp->pNext )
        Counter++;
    printf( "%d ", Counter );
*/
    // go through the cuts
clk = Abc_Clock();
    for  pCut = pCut.pNext; (pCut != nil); pCut = pCut.pNext  {
        // consider only 4-input cuts
        if pCut.nLeaves < 4 {
            continue; }
//            Cut_CutPrint( pCut, 0 ), printf( "\n" );

        // get the fanin permutation
        uTruth = 0xFFFF & *Cut_CutReadTruth(pCut);
        pPerm = p.pPerms4[ int32(p.pPerms[uTruth]) ];
        uPhase = p.pPhases[uTruth];
        // collect fanins with the corresponding permutation/phase
        Vec_PtrClear( p.vFaninsCur );
        Vec_PtrFill( p.vFaninsCur, int32(pCut.nLeaves), 0 );
        for  i = 0; i < int32(pCut.nLeaves); i++  {
            pFanin = Abc_NtkObj( pNode.pNtk, pCut.pLeaves[int32(pPerm[i])] );
            if pFanin == nil {
                break; }
            pFanin = Abc_ObjNotCond(pFanin, ((uPhase & (1<<i)) > 0) );
            Vec_PtrWriteEntry( p.vFaninsCur, i, pFanin );
        }
        if i != int32(pCut.nLeaves) {
            p.nCutsBad++;
            continue;
        }
        p.nCutsGood++;

        {
            var Counter int32;
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vFaninsCur)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vFaninsCur, i)), 1); i++  {
                if Abc_ObjFanoutNum(Abc_ObjRegular(pFanin)) == 1 {
                    Counter++; }
            }
            if Counter > 2 {
                continue; }
        }

clk2 = Abc_Clock();
/*
        printf( "Considering: (" );
        Vec_PtrForEachEntry( Abc_Obj_t *, p->vFaninsCur, pFanin, i )
            printf( "%d ", Abc_ObjFanoutNum(Abc_ObjRegular(pFanin)) );
        printf( ")\n" );
*/
        // mark the fanin boundary 
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vFaninsCur)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vFaninsCur, i)), 1); i++  {
            Abc_ObjRegular(pFanin).vFanouts.nSize++;
        }

        // label MFFC with current ID
        Abc_NtkIncrementTravId( pNode.pNtk );
        nNodesSaved = Abc_NodeMffcLabelAig( pNode );
        // unmark the fanin boundary
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vFaninsCur)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vFaninsCur, i)), 1); i++  {
            Abc_ObjRegular(pFanin).vFanouts.nSize--;
        }
p.timeMffc += Abc_Clock() - clk2;

        // evaluate the cut
clk2 = Abc_Clock();
        pGraph = Rwr_CutEvaluate( p, pNode, pCut, p.vFaninsCur, nNodesSaved, Required, &GainCur, fPlaceEnable );
p.timeEval += Abc_Clock() - clk2;

        // check if the cut is better than the current best one
        if pGraph != nil && GainBest < GainCur {
            // save this form
            nNodesSaveCur = nNodesSaved;
            GainBest  = GainCur;
            p.pGraph  = pGraph;
            p.fCompl = ((uPhase & (1<<4)) > 0);
            uTruthBest = 0xFFFF & *Cut_CutReadTruth(pCut);
            // collect fanins in the
            Vec_PtrClear( p.vFanins );
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vFaninsCur)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vFaninsCur, i)), 1); i++  {
                Vec_PtrPush( p.vFanins, pFanin );
            }
        }
    }
p.timeRes += Abc_Clock() - clk;

    if GainBest == -1 {
        return -1; }
/*
    if ( GainBest > 0 )
    {
        printf( "Class %d  ", p->pMap[uTruthBest] );
        printf( "Gain = %d. Node %d : ", GainBest, pNode->Id );
        Vec_PtrForEachEntry( Abc_Obj_t *, p->vFanins, pFanin, i )
            printf( "%d ", Abc_ObjRegular(pFanin)->Id );
        Dec_GraphPrint( stdout, p->pGraph, NULL, NULL );
        printf( "\n" );
    }
*/

//    printf( "%d", nNodesSaveCur - GainBest );
/*
    if ( GainBest > 0 )
    {
        if ( Rwr_CutIsBoolean( pNode, p->vFanins ) )
            printf( "b" );
        else
        {
            printf( "Node %d : ", pNode->Id );
            Vec_PtrForEachEntry( Abc_Obj_t *, p->vFanins, pFanin, i )
                printf( "%d ", Abc_ObjRegular(pFanin)->Id );
            printf( "a" );
        }
    }
*/
/*
    if ( GainBest > 0 )
        if ( p->fCompl )
            printf( "c" );
        else
            printf( "." );
*/

    // copy the leaves
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vFanins, i)), 1); i++  {
        Dec_GraphNode((Dec_Graph_t *)p.pGraph, i).pFunc = pFanin;
    }
/*
    printf( "(" );
    Vec_PtrForEachEntry( Abc_Obj_t *, p->vFanins, pFanin, i )
        printf( " %d", Abc_ObjRegular(pFanin)->vFanouts.nSize - 1 );
    printf( " )  " );
*/
//    printf( "%d ", Rwr_NodeGetDepth_rec( pNode, p->vFanins ) );

    p.nScores[p.pMap[uTruthBest]]++;
    p.nNodesGained += GainBest;
    if (fUseZeros != 0) || GainBest > 0 {
        p.nNodesRewritten++;
    }

    // report the progress
    if (fVeryVerbose != 0) && GainBest > 0 {
        fmt.Printf("Node %6s :   ", Abc_ObjName(pNode));
        fmt.Printf("Fanins = %d. ", p.vFanins.nSize);
        fmt.Printf("Save = %d.  ", nNodesSaveCur);
        fmt.Printf("Add = %d.  ", nNodesSaveCur-GainBest);
        fmt.Printf("GAIN = %d.  ", GainBest);
        fmt.Printf("Cone = %d.  ", p.pGraph? Dec_GraphNodeNum((Dec_Graph_t *)p.pGraph) : 0);
        fmt.Printf("Class = %d.  ", p.pMap[uTruthBest]);
        fmt.Printf("\n");
    }
    return GainBest;
}

/**Function*************************************************************

  Synopsis    [Evaluates the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_CutEvaluate(p *Rwr_Man_t, pRoot *Abc_Obj_t, pCut *Cut_Cut_t, vFaninsCur *Vec_Ptr_t, nNodesSaved int32, LevelMax int32, pGainBest *int32, fPlaceEnable int32) *Dec_Graph_t {
    extern int            Dec_GraphToNetworkCount( Abc_Obj_t * pRoot, Dec_Graph_t * pGraph, int NodeMax, int LevelMax );
    var vSubgraphs *Vec_Ptr_t;
    var pGraphBest *Dec_Graph_t  = nil; // Suppress "might be used uninitialized"
    var pGraphCur *Dec_Graph_t;
    var pNode, pFanin *Rwr_Node_t;
    var nNodesAdded, GainBest, i, k int32;
    var uTruth uint32;
    var CostBest float32;//, CostCur;
    // find the matching class of subgraphs
    uTruth = 0xFFFF & *Cut_CutReadTruth(pCut);
    vSubgraphs = Vec_VecEntry( p.vClasses, p.pMap[uTruth] );
    p.nSubgraphs += vSubgraphs.nSize;
    // determine the best subgraph
    GainBest = -1;
    CostBest = ABC_INFINITY;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSubgraphs)) && (((pNode) = (Rwr_Node_t *)Vec_PtrEntry(vSubgraphs, i)), 1); i++  {
        // get the current graph
        pGraphCur = (Dec_Graph_t *)pNode.pNext;
        // copy the leaves
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vFaninsCur)) && (((pFanin) = (Rwr_Node_t *)Vec_PtrEntry(vFaninsCur, k)), 1); k++  {
            Dec_GraphNode(pGraphCur, k).pFunc = pFanin;
        }
        // detect how many unlabeled nodes will be reused
        nNodesAdded = Dec_GraphToNetworkCount( pRoot, pGraphCur, nNodesSaved, LevelMax );
        if nNodesAdded == -1 {
            continue; }
        assert(nNodesSaved >= nNodesAdded, "nNodesSaved >= nNodesAdded");
/*
        // evaluate the cut
        if ( fPlaceEnable )
        {
            extern float Abc_PlaceEvaluateCut( Abc_Obj_t * pRoot, Vec_Ptr_t * vFanins );

            float Alpha = 0.5; // ???
            float PlaceCost;

            // get the placement cost of the cut
            PlaceCost = Abc_PlaceEvaluateCut( pRoot, vFaninsCur );

            // get the weigted cost of the cut
            CostCur = nNodesSaved - nNodesAdded + Alpha * PlaceCost;

            // do not allow uphill moves
            if ( nNodesSaved - nNodesAdded < 0 )
                continue;

            // decide what cut to use
            if ( CostBest > CostCur )
            {
                GainBest   = nNodesSaved - nNodesAdded; // pure node cost
                CostBest   = CostCur;                   // cost with placement
                pGraphBest = pGraphCur;                 // subgraph to be used for rewriting

                // score the graph
                if ( nNodesSaved - nNodesAdded > 0 )
                {
                    pNode->nScore++;
                    pNode->nGain += GainBest;
                    pNode->nAdded += nNodesAdded;
                }
            }
        }
        else
*/
        {
            // count the gain at this node
            if GainBest < nNodesSaved - nNodesAdded {
                GainBest   = nNodesSaved - nNodesAdded;
                pGraphBest = pGraphCur;

                // score the graph
                if nNodesSaved - nNodesAdded > 0 {
                    pNode.nScore++;
                    pNode.nGain += GainBest;
                    pNode.nAdded += nNodesAdded;
                }
            }
        }
    }
    if GainBest == -1 {
        return nil; }
    *pGainBest = GainBest;
    return pGraphBest;
}

/**Function*************************************************************

  Synopsis    [Checks the type of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_CutIsBoolean_rec(pObj *Abc_Obj_t, vLeaves *Vec_Ptr_t, fMarkA int32) {
    if Vec_PtrFind(vLeaves, pObj) >= 0 || Vec_PtrFind(vLeaves, Abc_ObjNot(pObj)) >= 0 {
        if fMarkA != 0 {
            pObj.fMarkA = 1;
        } else {
            pObj.fMarkB = 1; }
        return;
    }
    assert(!Abc_ObjIsCi(pObj), "!Abc_ObjIsCi(pObj)");
    Rwr_CutIsBoolean_rec( Abc_ObjFanin0(pObj), vLeaves, fMarkA );
    Rwr_CutIsBoolean_rec( Abc_ObjFanin1(pObj), vLeaves, fMarkA );
}

/**Function*************************************************************

  Synopsis    [Checks the type of the cut.]

  Description [Returns 1(0) if the cut is Boolean (algebraic).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_CutIsBoolean(pObj *Abc_Obj_t, vLeaves *Vec_Ptr_t) int32 {
    var pTemp *Abc_Obj_t;
    var i, RetValue int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pTemp = Abc_ObjRegular(pTemp);
        assert(!pTemp.fMarkA && !pTemp.fMarkB, "!pTemp.fMarkA && !pTemp.fMarkB");
    }
    Rwr_CutIsBoolean_rec( Abc_ObjFanin0(pObj), vLeaves, 1 );
    Rwr_CutIsBoolean_rec( Abc_ObjFanin1(pObj), vLeaves, 0 );
    RetValue = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pTemp = Abc_ObjRegular(pTemp);
        RetValue |= pTemp.fMarkA && pTemp.fMarkB;
        pTemp.fMarkA = pTemp.fMarkB = 0;
    }
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Count the nodes in the cut space of a node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_CutCountNumNodes_rec(pObj *Abc_Obj_t, pCut *Cut_Cut_t, vNodes *Vec_Ptr_t) {
    var i int32;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        if pCut.pLeaves[i] == pObj.Id {
            // check if the node is collected
            if pObj.fMarkC == 0 {
                pObj.fMarkC = 1;
                Vec_PtrPush( vNodes, pObj );
            }
            return;
        }
    }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // check if the node is collected
    if pObj.fMarkC == 0 {
        pObj.fMarkC = 1;
        Vec_PtrPush( vNodes, pObj );
    }
    // traverse the fanins
    Rwr_CutCountNumNodes_rec( Abc_ObjFanin0(pObj), pCut, vNodes );
    Rwr_CutCountNumNodes_rec( Abc_ObjFanin1(pObj), pCut, vNodes );
}

/**Function*************************************************************

  Synopsis    [Count the nodes in the cut space of a node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_CutCountNumNodes(pObj *Abc_Obj_t, pCut *Cut_Cut_t) int32 {
    var vNodes *Vec_Ptr_t;
    var i, Counter int32;
    // collect all nodes
    vNodes = Vec_PtrAlloc( 100 );
    for  pCut = pCut.pNext; (pCut != nil); pCut = pCut.pNext  {
        Rwr_CutCountNumNodes_rec( pObj, pCut, vNodes );
    }
    // clean all nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.fMarkC = 0;
    }
    // delete and return
    Counter = Vec_PtrSize(vNodes);
    Vec_PtrFree( vNodes );
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Returns depth of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_NodeGetDepth_rec(pObj *Abc_Obj_t, vLeaves *Vec_Ptr_t) int32 {
    var pLeaf *Abc_Obj_t;
    var i, Depth0, Depth1 int32;
    if Abc_ObjIsCi(pObj) {
        return 0; }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pLeaf) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        if pObj == Abc_ObjRegular(pLeaf) {
            return 0; }
    }
    Depth0 = Rwr_NodeGetDepth_rec( Abc_ObjFanin0(pObj), vLeaves );
    Depth1 = Rwr_NodeGetDepth_rec( Abc_ObjFanin1(pObj), vLeaves );
    return 1 + Abc_MaxInt( Depth0, Depth1 );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ScoresClean(p *Rwr_Man_t) {
    var vSubgraphs *Vec_Ptr_t;
    var pNode *Rwr_Node_t;
    var i, k int32;
    for  i = 0; i < p.vClasses.nSize; i++  {
        vSubgraphs = Vec_VecEntry( p.vClasses, i );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vSubgraphs)) && (((pNode) = (Rwr_Node_t *)Vec_PtrEntry(vSubgraphs, k)), 1); k++  {
            pNode.nScore = pNode.nGain = pNode.nAdded = 0;
        }
    }
}

var Gains [222]int32;

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ScoresCompare(pNum1 *int32, pNum2 *int32) int32 {
    if Gains[*pNum1] > Gains[*pNum2] {
        return -1; }
    if Gains[*pNum1] < Gains[*pNum2] {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ScoresReport(p *Rwr_Man_t) {
    extern void Ivy_TruthDsdComputePrint( unsigned uTruth );
    var Perm [222]int32;
    var vSubgraphs *Vec_Ptr_t;
    var pNode *Rwr_Node_t;
    var i, iNew, k int32;
    var uTruth uint32;
    // collect total gains
    assert(p.vClasses.nSize == 222, "p.vClasses.nSize == 222");
    for  i = 0; i < p.vClasses.nSize; i++  {
        Perm[i] = i;
        Gains[i] = 0;
        vSubgraphs = Vec_VecEntry( p.vClasses, i );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vSubgraphs)) && (((pNode) = (Rwr_Node_t *)Vec_PtrEntry(vSubgraphs, k)), 1); k++  {
            Gains[i] += pNode.nGain;
        }
    }
    // sort the gains
    qsort( Perm, (size_t)222, sizeof(int), (int (*)(void *, void *))Rwr_ScoresCompare );

    // print classes
    for  i = 0; i < p.vClasses.nSize; i++  {
        iNew = Perm[i];
        if Gains[iNew] == 0 {
            break; }
        vSubgraphs = Vec_VecEntry( p.vClasses, iNew );
        fmt.Printf("CLASS %3d: Subgr = %3d. Total gain = %6d.  ", iNew, Vec_PtrSize(vSubgraphs), Gains[iNew]);
        uTruth = uint32(p.pMapInv[iNew]);
        Extra_PrintBinary( os.Stdout, &uTruth, 16 );
        fmt.Printf("  ");
        Ivy_TruthDsdComputePrint( uint32(p.pMapInv[iNew]) | (uint32(p.pMapInv[iNew]) << 16) );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vSubgraphs)) && (((pNode) = (Rwr_Node_t *)Vec_PtrEntry(vSubgraphs, k)), 1); k++  {
            if pNode.nScore == 0 {
                continue; }
            fmt.Printf("    %2d: S=%5d. A=%5d. G=%6d. ", k, pNode.nScore, pNode.nAdded, pNode.nGain);
            Dec_GraphPrint( os.Stdout, (Dec_Graph_t *)pNode.pNext, nil, nil );
        }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

