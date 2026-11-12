package ivy

import "fmt"

/**CFile****************************************************************

  FileName    [ivyRwt.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Rewriting based on precomputation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyRwt.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"
//#include "bool/deco/deco.h"
//#include "opt/rwt/rwt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static unsigned Ivy_NodeGetTruth( Ivy_Obj_t * pObj, int * pNums, int nNums );

// static int Ivy_NodeRewrite( Ivy_Man_t * pMan, Rwt_Man_t * p, Ivy_Obj_t * pNode, int fUpdateLevel, int fUseZeroCost );

// static Dec_Graph_t * Rwt_CutEvaluate( Ivy_Man_t * pMan, Rwt_Man_t * p, Ivy_Obj_t * pRoot, 
//     Vec_Ptr_t * vFaninsCur, int nNodesSaved, int LevelMax, int * pGainBest, unsigned uTruth );


// static int Ivy_GraphToNetworkCount( Ivy_Man_t * p, Ivy_Obj_t * pRoot, Dec_Graph_t * pGraph, int NodeMax, int LevelMax );

// static void Ivy_GraphUpdateNetwork( Ivy_Man_t * p, Ivy_Obj_t * pRoot, Dec_Graph_t * pGraph, int fUpdateLevel, int nGain );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs incremental rewriting of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManRewritePre(p *Ivy_Man_t, fUpdateLevel int32, fUseZeroCost int32, fVerbose int32) int32 {
    var pManRwt *Rwt_Man_t;
    var pNode *Ivy_Obj_t;
    var i, nNodes, nGain int32;
    var clk abctime;
    var clkStart abctime = Abc_Clock();
    // start the rewriting manager
    pManRwt = Rwt_ManStart( 0 );
    p.pData = pManRwt;
    if pManRwt == nil {
        return 0; } 
    // create fanouts
    if (fUpdateLevel != 0) && p.fFanout == 0 {
        Ivy_ManStartFanout( p ); }
    // compute the reverse levels if level update is requested
    if fUpdateLevel != 0 {
        Ivy_ManRequiredLevels( p ); }
    // set the number of levels
//    p->nLevelMax = Ivy_ManLevels( p );
    // resynthesize each node once
    nNodes = Ivy_ManObjIdMax(p);
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pNode) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pNode) == nil {} else if !Ivy_ObjIsNode(pNode) {} else {
        // fix the fanin buffer problem
        Ivy_NodeFixBufferFanins( p, pNode, 1 );
        if Ivy_ObjIsBuf(pNode) {
            continue; }
        // stop if all nodes have been tried once
        if i > nNodes {
            break; }
        // for each cut, try to resynthesize it
        nGain = Ivy_NodeRewrite( p, pManRwt, pNode, fUpdateLevel, fUseZeroCost );
        if nGain > 0 || (nGain == 0 && (fUseZeroCost != 0)) {
            var pGraph *Dec_Graph_t  = (Dec_Graph_t *)Rwt_ManReadDecs(pManRwt);
            var fCompl int32            = Rwt_ManReadCompl(pManRwt);
/*
            {
                Ivy_Obj_t * pObj;
                int i;
                printf( "USING: (" );
                Vec_PtrForEachEntry( Ivy_Obj_t *, Rwt_ManReadLeaves(pManRwt), pObj, i )
                    printf( "%d ", Ivy_ObjFanoutNum(Ivy_Regular(pObj)) );
                printf( ")   Gain = %d.\n", nGain );
            }
            if ( nGain > 0 )
            { // print stats on the MFFC
                extern void Ivy_NodeMffcConeSuppPrint( Ivy_Obj_t * pNode );
                printf( "Node %6d : Gain = %4d  ", pNode->Id, nGain );
                Ivy_NodeMffcConeSuppPrint( pNode );
            }
*/
            // complement the FF if needed
clk = Abc_Clock();
            if fCompl != 0 { Dec_GraphComplement( pGraph ); }
            Ivy_GraphUpdateNetwork( p, pNode, pGraph, fUpdateLevel, nGain );
            if fCompl != 0 { Dec_GraphComplement( pGraph ); }
Rwt_ManAddTimeUpdate( pManRwt, Abc_Clock() - clk );
        }
    }
    }
Rwt_ManAddTimeTotal( pManRwt, Abc_Clock() - clkStart );
    // print stats
    if fVerbose != 0 {
        Rwt_ManPrintStats( pManRwt ); }
    // delete the managers
    Rwt_ManStop( pManRwt );
    p.pData = nil;
    // fix the levels
    if fUpdateLevel != 0 {
        Vec_IntFree( p.vRequired );
p.vRequired = nil;
    } else {
        Ivy_ManResetLevels( p ); }
    // check
    if i = Ivy_ManCleanup(p) {
        fmt.Printf("Cleanup after rewriting removed %d dangling nodes.\n", i); }
    if !Ivy_ManCheck(p) {
        fmt.Printf("Ivy_ManRewritePre(): The check has failed.\n"); }
    return 1;
}

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
func Ivy_NodeRewrite(pMan *Ivy_Man_t, p *Rwt_Man_t, pNode *Ivy_Obj_t, fUpdateLevel int32, fUseZeroCost int32) int32 {
    var fVeryVerbose int32;
    var pGraph *Dec_Graph_t;
    var pStore *Ivy_Store_t;
    var pCut *Ivy_Cut_t;
    var pFanin *Ivy_Obj_t;
    var uPhase uint32;
    var uTruthBest uint32; // Suppress "might be used uninitialized"
    var uTruth uint32;
    var pPerm string;
    var Required, nNodesSaved int32;
    var nNodesSaveCur int32  = -1; // Suppress "might be used uninitialized"
    var i int32;
    var c int32;
    var GainCur int32 = -1;
    var GainBest int32 = -1;
    var clk, clk2 abctime;

    p.nNodesConsidered++;
    // get the required times
    if fUpdateLevel != 0 {
        Required = Vec_IntEntry( pMan.vRequired, pNode.Id );
    } else {
        Required = 1000000;
    }
    // get the node's cuts
clk = Abc_Clock();
    pStore = Ivy_NodeFindCutsAll( pMan, pNode, 5 );
p.timeCut += Abc_Clock() - clk;

    // go through the cuts
clk = Abc_Clock();
    for  c = 1; c < pStore.nCuts; c++  {
        pCut = pStore.pCuts + c;
        // consider only 4-input cuts
        if pCut.nSize != 4 {
            continue; }
        // skip the cuts with buffers
        for  i = 0; i < int32(pCut.nSize); i++  {
            if Ivy_ObjIsBuf( Ivy_ManObj(pMan, pCut.pArray[i]) ) {
                break; }
        }
        if i != pCut.nSize {
            p.nCutsBad++;
            continue;
        }
        p.nCutsGood++;
        // get the fanin permutation
clk2 = Abc_Clock();
        uTruth = 0xFFFF & Ivy_NodeGetTruth( pNode, pCut.pArray, pCut.nSize );  // truth table
p.timeTruth += Abc_Clock() - clk2;
        pPerm = p.pPerms4[ int32( p.pPerms[uTruth]) ];
        uPhase = p.pPhases[uTruth];
        // collect fanins with the corresponding permutation/phase
        Vec_PtrClear( p.vFaninsCur );
        Vec_PtrFill( p.vFaninsCur, int32(pCut.nSize), 0 );
        for  i = 0; i < int32(pCut.nSize); i++  {
            pFanin = Ivy_ManObj( pMan, pCut.pArray[int32(pPerm[i])] );
            assert(Ivy_ObjIsNode(pFanin) || Ivy_ObjIsCi(pFanin), "Ivy_ObjIsNode(pFanin) || Ivy_ObjIsCi(pFanin)");
            pFanin = Ivy_NotCond(pFanin, ((uPhase & (1<<i)) > 0) );
            Vec_PtrWriteEntry( p.vFaninsCur, i, pFanin );
        }
clk2 = Abc_Clock();
/*
        printf( "Considering: (" );
        Vec_PtrForEachEntry( Ivy_Obj_t *, p->vFaninsCur, pFanin, i )
            printf( "%d ", Ivy_ObjFanoutNum(Ivy_Regular(pFanin)) );
        printf( ")\n" );
*/
        // mark the fanin boundary 
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vFaninsCur)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(p.vFaninsCur, i)), 1); i++  {
            Ivy_ObjRefsInc( Ivy_Regular(pFanin) );
        }
        // label MFFC with current ID
        Ivy_ManIncrementTravId( pMan );
        nNodesSaved = Ivy_ObjMffcLabel( pMan, pNode );
        // unmark the fanin boundary
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vFaninsCur)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(p.vFaninsCur, i)), 1); i++  {
            Ivy_ObjRefsDec( Ivy_Regular(pFanin) );
        }
p.timeMffc += Abc_Clock() - clk2;

        // evaluate the cut
clk2 = Abc_Clock();
        pGraph = Rwt_CutEvaluate( pMan, p, pNode, p.vFaninsCur, nNodesSaved, Required, &GainCur, uTruth );
p.timeEval += Abc_Clock() - clk2;

        // check if the cut is better than the current best one
        if pGraph != nil && GainBest < GainCur {
            // save this form
            nNodesSaveCur = nNodesSaved;
            GainBest  = GainCur;
            p.pGraph  = pGraph;
            p.fCompl = ((uPhase & (1<<4)) > 0);
            uTruthBest = uTruth;
            // collect fanins in the
            Vec_PtrClear( p.vFanins );
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vFaninsCur)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(p.vFaninsCur, i)), 1); i++  {
                Vec_PtrPush( p.vFanins, pFanin );
            }
        }
    }
p.timeRes += Abc_Clock() - clk;

    if GainBest == -1 {
        return -1; }

//    printf( "%d", nNodesSaveCur - GainBest );
/*
    if ( GainBest > 0 )
    {
        if ( Rwt_CutIsintean( pNode, p->vFanins ) )
            printf( "b" );
        else
        {
            printf( "Node %d : ", pNode->Id );
            Vec_PtrForEachEntry( Ivy_Obj_t *, p->vFanins, pFanin, i )
                printf( "%d ", Ivy_Regular(pFanin)->Id );
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
    for  i = 0; (i < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(p.vFanins, i)), 1); i++  {
        Dec_GraphNode((Dec_Graph_t *)p.pGraph, i).pFunc = pFanin;
    }

    p.nScores[p.pMap[uTruthBest]]++;
    p.nNodesGained += GainBest;
    if (fUseZeroCost != 0) || GainBest > 0 {
        p.nNodesRewritten++; }

    // report the progress
    if (fVeryVerbose != 0) && GainBest > 0 {
        fmt.Printf("Node %6d :   ", Ivy_ObjId(pNode));
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

  Synopsis    [Computes the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ivy_NodeGetTruth_rec_uMasks [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
func Ivy_NodeGetTruth_rec(pObj *Ivy_Obj_t, pNums *int32, nNums int32) uint32 {
    
    var uTruth0, uTruth1 uint32;
    var i int32;
    for  i = 0; i < nNums; i++  {
        if pObj.Id == pNums[i] {
            return spStatic_Ivy_NodeGetTruth_rec_uMasks[i]; }
    }
    assert(Ivy_ObjIsNode(pObj) || Ivy_ObjIsBuf(pObj), "Ivy_ObjIsNode(pObj) || Ivy_ObjIsBuf(pObj)");
    uTruth0 = Ivy_NodeGetTruth_rec( Ivy_ObjFanin0(pObj), pNums, nNums );
    if Ivy_ObjFaninC0(pObj) {
        uTruth0 = ~uTruth0; }
    if Ivy_ObjIsBuf(pObj) {
        return uTruth0; }
    uTruth1 = Ivy_NodeGetTruth_rec( Ivy_ObjFanin1(pObj), pNums, nNums );
    if Ivy_ObjFaninC1(pObj) {
        uTruth1 = ~uTruth1; }
    return uTruth0 & uTruth1;
}


/**Function*************************************************************

  Synopsis    [Computes the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeGetTruth(pObj *Ivy_Obj_t, pNums *int32, nNums int32) uint32 {
    assert(nNums < 6, "nNums < 6");
    return Ivy_NodeGetTruth_rec( pObj, pNums, nNums );
}

/**Function*************************************************************

  Synopsis    [Evaluates the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_CutEvaluate(pMan *Ivy_Man_t, p *Rwt_Man_t, pRoot *Ivy_Obj_t, vFaninsCur *Vec_Ptr_t, nNodesSaved int32, LevelMax int32, pGainBest *int32, uTruth uint32) *Dec_Graph_t {
    var vSubgraphs *Vec_Ptr_t;
    var pGraphBest *Dec_Graph_t  = nil; // Suppress "might be used uninitialized"
    var pGraphCur *Dec_Graph_t;
    var pNode, pFanin *Rwt_Node_t;
    var nNodesAdded, GainBest, i, k int32;
    // find the matching class of subgraphs
    vSubgraphs = Vec_VecEntry( p.vClasses, p.pMap[uTruth] );
    p.nSubgraphs += vSubgraphs.nSize;
    // determine the best subgraph
    GainBest = -1;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSubgraphs)) && (((pNode) = (Rwt_Node_t *)Vec_PtrEntry(vSubgraphs, i)), 1); i++  {
        // get the current graph
        pGraphCur = (Dec_Graph_t *)pNode.pNext;
        // copy the leaves
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vFaninsCur)) && (((pFanin) = (Rwt_Node_t *)Vec_PtrEntry(vFaninsCur, k)), 1); k++  {
            Dec_GraphNode(pGraphCur, k).pFunc = pFanin;
        }
        // detect how many unlabeled nodes will be reused
        nNodesAdded = Ivy_GraphToNetworkCount( pMan, pRoot, pGraphCur, nNodesSaved, LevelMax );
        if nNodesAdded == -1 {
            continue; }
        assert(nNodesSaved >= nNodesAdded, "nNodesSaved >= nNodesAdded");
        // count the gain at this node
        if GainBest < nNodesSaved - nNodesAdded {
            GainBest   = nNodesSaved - nNodesAdded;
            pGraphBest = pGraphCur;
        }
    }
    if GainBest == -1 {
        return nil; }
    *pGainBest = GainBest;
    return pGraphBest;
}


/**Function*************************************************************

  Synopsis    [Counts the number of new nodes added when using this graph.]

  Description [AIG nodes for the fanins should be assigned to pNode->pFunc 
  of the leaves of the graph before calling this procedure. 
  Returns -1 if the number of nodes and levels exceeded the given limit or 
  the number of levels exceeded the maximum allowed level.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_GraphToNetworkCount(p *Ivy_Man_t, pRoot *Ivy_Obj_t, pGraph *Dec_Graph_t, NodeMax int32, LevelMax int32) int32 {
    var pNode, pNode0, pNode1 *Dec_Node_t;
    var pAnd, pAnd0, pAnd1 *Ivy_Obj_t;
    var i, Counter, LevelNew, LevelOld int32;
    // check for constant function or a literal
    if Dec_GraphIsConst(pGraph) || Dec_GraphIsVar(pGraph) {
        return 0; }
    // set the levels of the leaves
    // inlining of macro Dec_GraphForEachLeaf
    for  i = 0; (i < (pGraph).nLeaves) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        pNode.Level = Ivy_Regular((Ivy_Obj_t *)pNode.pFunc).Level;
    }
    // compute the AIG size after adding the internal nodes
    Counter = 0;
    // inlining of macro Dec_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        // get the children of this node
        pNode0 = Dec_GraphNode( pGraph, pNode.eEdge0.Node );
        pNode1 = Dec_GraphNode( pGraph, pNode.eEdge1.Node );
        // get the AIG nodes corresponding to the children 
        pAnd0 = (Ivy_Obj_t *)pNode0.pFunc; 
        pAnd1 = (Ivy_Obj_t *)pNode1.pFunc; 
        if pAnd0 && (pAnd1 != nil) {
            // if they are both present, find the resulting node
            pAnd0 = Ivy_NotCond( pAnd0, pNode.eEdge0.fCompl );
            pAnd1 = Ivy_NotCond( pAnd1, pNode.eEdge1.fCompl );
            pAnd  = Ivy_TableLookup( p, Ivy_ObjCreateGhost(p, pAnd0, pAnd1, IVY_AND, IVY_INIT_NONE) );
            // return -1 if the node is the same as the original root
            if Ivy_Regular(pAnd) == pRoot {
                return -1; }
        
        } else {
            pAnd = nil; }
        // count the number of added nodes
        if pAnd == nil || Ivy_ObjIsTravIdCurrent(p, Ivy_Regular(pAnd)) {
            if Counter++; Counter > NodeMax {
                return -1; }
        }
        // count the number of new levels
        // inlining of macro RWT_MAX
        LevelNew = 1 + (((pNode0.Level) > (pNode1.Level))? (pNode0.Level) : (pNode1.Level));
        if pAnd {
            if Ivy_Regular(pAnd) == p.pConst1 {
                LevelNew = 0;
            } else if Ivy_Regular(pAnd) == Ivy_Regular(pAnd0) {
                LevelNew = int32(Ivy_Regular(pAnd0).Level);
            } else if Ivy_Regular(pAnd) == Ivy_Regular(pAnd1) {
                LevelNew = int32(Ivy_Regular(pAnd1).Level); }
            LevelOld = int32(Ivy_Regular(pAnd).Level);
//            assert( LevelNew == LevelOld );
        }
        if LevelNew > LevelMax {
            return -1; }
        pNode.pFunc = pAnd;
        pNode.Level = LevelNew;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description [AIG nodes for the fanins should be assigned to pNode->pFunc
  of the leaves of the graph before calling this procedure.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_GraphToNetwork(p *Ivy_Man_t, pGraph *Dec_Graph_t) *Ivy_Obj_t {
    var pAnd0, pAnd1 *Ivy_Obj_t;
    var pNode *Dec_Node_t  = nil; // Suppress "might be used uninitialized"
    var i int32;
    // check for constant function
    if Dec_GraphIsConst(pGraph) {
        return Ivy_NotCond( Ivy_ManConst1(p), Dec_GraphIsComplement(pGraph) ); }
    // check for a literal
    if Dec_GraphIsVar(pGraph) {
        return Ivy_NotCond( (Ivy_Obj_t *)Dec_GraphVar(pGraph).pFunc, Dec_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Dec_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = Ivy_NotCond( (Ivy_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge0.Node).pFunc, pNode.eEdge0.fCompl ); 
        pAnd1 = Ivy_NotCond( (Ivy_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge1.Node).pFunc, pNode.eEdge1.fCompl ); 
        pNode.pFunc = Ivy_And( p, pAnd0, pAnd1 );
    }
    // complement the result if necessary
    return Ivy_NotCond( (Ivy_Obj_t *)pNode.pFunc, Dec_GraphIsComplement(pGraph) );
}

/**Function*************************************************************

  Synopsis    [Replaces MFFC of the node by the new factored form.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_GraphUpdateNetwork(p *Ivy_Man_t, pRoot *Ivy_Obj_t, pGraph *Dec_Graph_t, fUpdateLevel int32, nGain int32) {
    var pRootNew *Ivy_Obj_t;
    var nNodesNew, nNodesOld, Required int32;
    if fUpdateLevel != 0 {
        Required = Vec_IntEntry( p.vRequired, pRoot.Id );
    } else {
        Required = 1000000;
    }
    nNodesOld = Ivy_ManNodeNum(p);
    // create the new structure of nodes
    pRootNew = Ivy_GraphToNetwork( p, pGraph );
    assert(int32(Ivy_Regular(pRootNew).Level) <= Required, "int32(Ivy_Regular(pRootNew).Level) <= Required");
//    if ( Ivy_Regular(pRootNew)->Level == Required )
//        printf( "Difference %d.\n", Ivy_Regular(pRootNew)->Level - Required );
    // remove the old nodes
//    Ivy_AigReplace( pMan->pManFunc, pRoot, pRootNew, fUpdateLevel );
/*
    if ( Ivy_IsComplement(pRootNew) )
        printf( "c" );
    else
        printf( "d" );
    if ( Ivy_ObjRefs(Ivy_Regular(pRootNew)) > 0 )
        printf( "%d", Ivy_ObjRefs(Ivy_Regular(pRootNew)) );
    printf( " " );
*/
    Ivy_ObjReplace( p, pRoot, pRootNew, 1, 0, 1 );
    // compare the gains
    nNodesNew = Ivy_ManNodeNum(p);
    assert(nGain <= nNodesOld - nNodesNew, "nGain <= nNodesOld - nNodesNew");
    // propagate the buffer
    Ivy_ManPropagateBuffers( p, 1 );
}

/**Function*************************************************************

  Synopsis    [Replaces MFFC of the node by the new factored form.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_GraphUpdateNetwork3(p *Ivy_Man_t, pRoot *Ivy_Obj_t, pGraph *Dec_Graph_t, fUpdateLevel int32, nGain int32) {
    var pRootNew, pFanin *Ivy_Obj_t;
    var nNodesNew, nNodesOld, i, nRefsOld int32;
    nNodesOld = Ivy_ManNodeNum(p);

//printf( "Before = %d. ", Ivy_ManNodeNum(p) );
    // mark the cut
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(((Rwt_Man_t *)p.pData).vFanins)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(((Rwt_Man_t *)p.pData).vFanins, i)), 1); i++  {
        Ivy_ObjRefsInc( Ivy_Regular(pFanin) );
    }
    // deref the old cone
    nRefsOld = pRoot.nRefs;  
    pRoot.nRefs = 0;
    Ivy_ObjDelete_rec( p, pRoot, 0 );
    pRoot.nRefs = nRefsOld;
    // unmark the cut
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(((Rwt_Man_t *)p.pData).vFanins)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(((Rwt_Man_t *)p.pData).vFanins, i)), 1); i++  {
        Ivy_ObjRefsDec( Ivy_Regular(pFanin) );
    }
//printf( "Deref = %d. ", Ivy_ManNodeNum(p) );
 
    // create the new structure of nodes
    pRootNew = Ivy_GraphToNetwork( p, pGraph );
//printf( "Create = %d. ", Ivy_ManNodeNum(p) );
    // remove the old nodes
//    Ivy_AigReplace( pMan->pManFunc, pRoot, pRootNew, fUpdateLevel );
/*
    if ( Ivy_IsComplement(pRootNew) )
        printf( "c" );
    else
        printf( "d" );
    if ( Ivy_ObjRefs(Ivy_Regular(pRootNew)) > 0 )
        printf( "%d", Ivy_ObjRefs(Ivy_Regular(pRootNew)) );
    printf( " " );
*/
    Ivy_ObjReplace( p, pRoot, pRootNew, 0, 0, 1 );
//printf( "Replace = %d. ", Ivy_ManNodeNum(p) );

    // delete remaining dangling nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(((Rwt_Man_t *)p.pData).vFanins)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(((Rwt_Man_t *)p.pData).vFanins, i)), 1); i++  {
        pFanin = Ivy_Regular(pFanin);
        if !Ivy_ObjIsNone(pFanin) && Ivy_ObjRefs(pFanin) == 0 {
            Ivy_ObjDelete_rec( p, pFanin, 1 ); }
    }
//printf( "Deref = %d. ", Ivy_ManNodeNum(p) );
//printf( "\n" );

    // compare the gains
    nNodesNew = Ivy_ManNodeNum(p);
    assert(nGain <= nNodesOld - nNodesNew, "nGain <= nNodesOld - nNodesNew");
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

