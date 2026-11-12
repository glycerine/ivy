package dar

import "fmt"

/**CFile****************************************************************

  FileName    [darRefact.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [Refactoring.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: darRefact.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "darInt.h"
//#include "bool/kit/kit.h"

//#include "bool/bdc/bdc.h"
//#include "bool/bdc/bdcInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the refactoring manager
              
type Ref_Man_t struct {
    // input data
    pPars    *Dar_RefPar_t;          // rewriting parameters
    pAig       *Aig_Man_t;           // AIG manager 
    // computed cuts
    vCuts       *Vec_Vec_t;          // the storage for cuts
    // truth table and ISOP
    vTruthElem       *Vec_Ptr_t;     // elementary truth tables
    vTruthStore       *Vec_Ptr_t;    // storage for truth tables
    vMemory       *Vec_Int_t;        // storage for ISOP
    vCutNodes       *Vec_Ptr_t;      // storage for internal nodes of the cut
    // various data members
    vLeavesBest       *Vec_Ptr_t;    // the best set of leaves
    pGraphBest     *Kit_Graph_t;     // the best factored form
    GainBest              int32;       // the best gain
    LevelBest              int32;      // the level of node with the best gain
    // bi-decomposition
    DecPars        Bdc_Par_t;        // decomposition parameters
    pManDec       *Bdc_Man_t;        // decomposition manager
    // node statistics
    nNodesInit              int32;     // the initial number of nodes
    nNodesTried              int32;    // the number of nodes tried
    nNodesBelow              int32;    // the number of nodes below the level limit
    nNodesExten              int32;    // the number of nodes with extended cut
    nCutsUsed              int32;      // the number of rewriting steps
    nCutsTried              int32;     // the number of cuts tries
    // timing statistics
    timeCuts          abctime;
    timeEval          abctime;
    timeOther          abctime;
    timeTotal          abctime;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the structure with default assignment of parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManDefaultRefParams(pPars *Dar_RefPar_t) {
    memset( pPars, 0, sizeof(Dar_RefPar_t) );
    pPars.nMffcMin     =  2;  // the min MFFC size for which refactoring is used
    pPars.nLeafMax     = 12;  // the max number of leaves of a cut
    pPars.nCutsMax     =  5;  // the max number of cuts to consider  
    pPars.fUpdateLevel =  0;
    pPars.fUseZeros    =  0;
    pPars.fVerbose     =  0;
    pPars.fVeryVerbose =  0;
}

/**Function*************************************************************

  Synopsis    [Starts the rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManRefStart(pAig *Aig_Man_t, pPars *Dar_RefPar_t) *Ref_Man_t {
    var p *Ref_Man_t;
    // start the manager
    p = new(Ref_Man_t);
    *p = Ref_Man_t{};
    p.pAig         = pAig;
    p.pPars        = pPars;
    // other data
    p.vCuts        = Vec_VecStart( pPars.nCutsMax );
    p.vTruthElem   = Vec_PtrAllocTruthTables( pPars.nLeafMax );
    p.vTruthStore  = Vec_PtrAllocSimInfo( 1024, Kit_TruthWordNum(pPars.nLeafMax) );
    p.vMemory      = Vec_IntAlloc( 1 << 16 );
    p.vCutNodes    = Vec_PtrAlloc( 256 );
    p.vLeavesBest  = Vec_PtrAlloc( pPars.nLeafMax );
    // alloc bi-decomposition manager
    p.DecPars.nVarsMax = pPars.nLeafMax;
    p.DecPars.fVerbose = pPars.fVerbose;
    p.DecPars.fVeryVerbose = 0;
//    p->pManDec = Bdc_ManAlloc( &p->DecPars );
    return p;
}

/**Function*************************************************************

  Synopsis    [Prints out the statistics of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManRefPrintStats(p *Ref_Man_t) {
    var Gain int32  = p.nNodesInit - Aig_ManNodeNum(p.pAig);
    fmt.Printf("NodesBeg = %8d. NodesEnd = %8d. Gain = %6d. (%6.2f %%).\n", p.nNodesInit, Aig_ManNodeNum(p.pAig), Gain, 100.0*Gain/p.nNodesInit);
    fmt.Printf("Tried = %6d. Below = %5d. Extended = %5d.  Used = %5d.  Levels = %4d.\n", p.nNodesTried, p.nNodesBelow, p.nNodesExten, p.nCutsUsed, Aig_ManLevels(p.pAig));
    ABC_PRT( "Cuts  ", p.timeCuts );
    ABC_PRT( "Eval  ", p.timeEval );
    ABC_PRT( "Other ", p.timeOther );
    ABC_PRT( "TOTAL ", p.timeTotal );
}

/**Function*************************************************************

  Synopsis    [Stops the rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManRefStop(p *Ref_Man_t) {
    if p.pManDec {
        Bdc_ManFree( p.pManDec ); }
    if p.pPars.fVerbose {
        Dar_ManRefPrintStats( p ); }
    Vec_VecFree( p.vCuts );
    Vec_PtrFree( p.vTruthElem );
    Vec_PtrFree( p.vTruthStore );
    Vec_PtrFree( p.vLeavesBest );
    Vec_IntFree( p.vMemory );
    Vec_PtrFree( p.vCutNodes );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ref_ObjComputeCuts(pAig *Aig_Man_t, pRoot *Aig_Obj_t, vCuts *Vec_Vec_t) {
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ref_ObjPrint(pObj *Aig_Obj_t) {
    fmt.Printf("%d", pObj? Aig_Regular(pObj).Id : -1);
    if pObj != nil {
        fmt.Printf("(%d) ", Aig_IsComplement(pObj)); }
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
func Dar_RefactTryGraph(pAig *Aig_Man_t, pRoot *Aig_Obj_t, vCut *Vec_Ptr_t, pGraph *Kit_Graph_t, NodeMax int32, LevelMax int32) int32 {
    var pNode, pNode0, pNode1 *Kit_Node_t;
    var pAnd, pAnd0, pAnd1 *Aig_Obj_t;
    var i, Counter, LevelNew, LevelOld int32;
    // check for constant function or a literal
    if Kit_GraphIsConst(pGraph) || Kit_GraphIsVar(pGraph) {
        return 0; }
    // set the levels of the leaves
    // inlining of macro Kit_GraphForEachLeaf
    for  i = 0; (i < (pGraph).nLeaves) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pNode.pFunc = Vec_PtrEntry(vCut, i);
        pNode.Level = Aig_Regular((Aig_Obj_t *)pNode.pFunc).Level;
        assert(Aig_Regular((Aig_Obj_t *)pNode.pFunc).Level < (1<<24)-1, "Aig_Regular((Aig_Obj_t *)pNode.pFunc).Level < (1<<24)-1");
    }
//printf( "Trying:\n" );
    // compute the AIG size after adding the internal nodes
    Counter = 0;
    // inlining of macro Kit_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        // get the children of this node
        pNode0 = Kit_GraphNode( pGraph, pNode.eEdge0.Node );
        pNode1 = Kit_GraphNode( pGraph, pNode.eEdge1.Node );
        // get the AIG nodes corresponding to the children 
        pAnd0 = (Aig_Obj_t *)pNode0.pFunc; 
        pAnd1 = (Aig_Obj_t *)pNode1.pFunc; 
        if pAnd0 && (pAnd1 != nil) {
            // if they are both present, find the resulting node
            pAnd0 = Aig_NotCond( pAnd0, pNode.eEdge0.fCompl );
            pAnd1 = Aig_NotCond( pAnd1, pNode.eEdge1.fCompl );
            pAnd  = Aig_TableLookupTwo( pAig, pAnd0, pAnd1 );
            // return -1 if the node is the same as the original root
            if Aig_Regular(pAnd) == pRoot {
                return -1; }
        
        } else {
            pAnd = nil; }
        // count the number of added nodes
        if pAnd == nil || Aig_ObjIsTravIdCurrent(pAig, Aig_Regular(pAnd)) {
            if Counter++; Counter > NodeMax {
                return -1; }
        }
        // count the number of new levels
        LevelNew = 1 + Abc_MaxInt( pNode0.Level, pNode1.Level );
        if pAnd {
            if Aig_Regular(pAnd) == Aig_ManConst1(pAig) {
                LevelNew = 0;
            } else if Aig_Regular(pAnd) == Aig_Regular(pAnd0) {
                LevelNew = int32(Aig_Regular(pAnd0).Level);
            } else if Aig_Regular(pAnd) == Aig_Regular(pAnd1) {
                LevelNew = int32(Aig_Regular(pAnd1).Level); }
            LevelOld = int32(Aig_Regular(pAnd).Level);
//            assert( LevelNew == LevelOld );
        }
        if LevelNew > LevelMax {
            return -1; }
        pNode.pFunc = pAnd;
        pNode.Level = LevelNew;
/*
printf( "Checking " );
Ref_ObjPrint( pAnd0 );
printf( " and " );
Ref_ObjPrint( pAnd1 );
printf( "  Result " );
Ref_ObjPrint( pNode->pFunc );
printf( "\n" );
*/
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_RefactBuildGraph(pAig *Aig_Man_t, vCut *Vec_Ptr_t, pGraph *Kit_Graph_t) *Aig_Obj_t {
    var pAnd0, pAnd1 *Aig_Obj_t;
    var pNode *Kit_Node_t  = nil;
    var i int32;
    // check for constant function
    if Kit_GraphIsConst(pGraph) {
        return Aig_NotCond( Aig_ManConst1(pAig), Kit_GraphIsComplement(pGraph) ); }
    // set the leaves
    // inlining of macro Kit_GraphForEachLeaf
    for  i = 0; (i < (pGraph).nLeaves) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pNode.pFunc = Vec_PtrEntry(vCut, i);
    }
    // check for a literal
    if Kit_GraphIsVar(pGraph) {
        return Aig_NotCond( (Aig_Obj_t *)Kit_GraphVar(pGraph).pFunc, Kit_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
//printf( "Building (current number %d):\n", Aig_ManObjNumMax(pAig) );
    // inlining of macro Kit_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = Aig_NotCond( (Aig_Obj_t *)Kit_GraphNode(pGraph, pNode.eEdge0.Node).pFunc, pNode.eEdge0.fCompl ); 
        pAnd1 = Aig_NotCond( (Aig_Obj_t *)Kit_GraphNode(pGraph, pNode.eEdge1.Node).pFunc, pNode.eEdge1.fCompl ); 
        pNode.pFunc = Aig_And( pAig, pAnd0, pAnd1 );
/*
printf( "Checking " );
Ref_ObjPrint( pAnd0 );
printf( " and " );
Ref_ObjPrint( pAnd1 );
printf( "  Result " );
Ref_ObjPrint( pNode->pFunc );
printf( "\n" );
*/
    }
    // complement the result if necessary
    return Aig_NotCond( (Aig_Obj_t *)pNode.pFunc, Kit_GraphIsComplement(pGraph) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManRefactorTryCuts(p *Ref_Man_t, pObj *Aig_Obj_t, nNodesSaved int32, Required int32) int32 {
    var vCut *Vec_Ptr_t;
    var pGraphCur *Kit_Graph_t;
    var k, RetValue, GainCur, nNodesAdded int32;
    var pTruth *uint32;

    p.GainBest = -1;
    p.pGraphBest = nil;
    Vec_VecForEachLevel( p.vCuts, vCut, k )
    {
        if Vec_PtrSize(vCut) == 0 {
            continue; }
//        if ( Vec_PtrSize(vCut) != 0 && Vec_PtrSize(Vec_VecEntry(p->vCuts, k+1)) != 0 )
//            continue;

        p.nCutsTried++;
        // get the cut nodes
        Aig_ObjCollectCut( pObj, vCut, p.vCutNodes );
        // get the truth table
        pTruth = Aig_ManCutTruth( pObj, vCut, p.vCutNodes, p.vTruthElem, p.vTruthStore );
        if Kit_TruthIsConst0(pTruth, Vec_PtrSize(vCut)) {
            p.GainBest = Aig_NodeMffcSupp( p.pAig, pObj, 0, nil );
            p.pGraphBest = Kit_GraphCreateConst0();
            Vec_PtrCopy( p.vLeavesBest, vCut );
            return p.GainBest;
        }
        if Kit_TruthIsConst1(pTruth, Vec_PtrSize(vCut)) {
            p.GainBest = Aig_NodeMffcSupp( p.pAig, pObj, 0, nil );
            p.pGraphBest = Kit_GraphCreateConst1();
            Vec_PtrCopy( p.vLeavesBest, vCut );
            return p.GainBest;
        }

        // try the positive phase
        RetValue = Kit_TruthIsop( pTruth, Vec_PtrSize(vCut), p.vMemory, 0 );
        if RetValue > -1 {
            pGraphCur = Kit_SopFactor( p.vMemory, 0, Vec_PtrSize(vCut), p.vMemory );
/*
{
    int RetValue;
    RetValue = Bdc_ManDecompose( p->pManDec, pTruth, NULL, Vec_PtrSize(vCut), NULL, 1000 );
    printf( "Graph = %d. Bidec = %d.\n", Kit_GraphNodeNum(pGraphCur), RetValue );
}
*/
            nNodesAdded = Dar_RefactTryGraph( p.pAig, pObj, vCut, pGraphCur, nNodesSaved - !p.pPars.fUseZeros, Required );
            if nNodesAdded > -1 {
                GainCur = nNodesSaved - nNodesAdded;
                if p.GainBest < GainCur || (p.GainBest == GainCur && 
                    (Kit_GraphIsConst(pGraphCur) || Kit_GraphRootLevel(pGraphCur) < Kit_GraphRootLevel(p.pGraphBest))) {
                    p.GainBest = GainCur;
                    if p.pGraphBest {
                        Kit_GraphFree( p.pGraphBest ); }
                    p.pGraphBest = pGraphCur;
                    Vec_PtrCopy( p.vLeavesBest, vCut );
                
                } else {
                    Kit_GraphFree( pGraphCur ); }
            
            } else {
                Kit_GraphFree( pGraphCur ); }
        }
        // try negative phase
        Kit_TruthNot( pTruth, pTruth, Vec_PtrSize(vCut) );
        RetValue = Kit_TruthIsop( pTruth, Vec_PtrSize(vCut), p.vMemory, 0 );
//        Kit_TruthNot( pTruth, pTruth, Vec_PtrSize(vCut) );
        if RetValue > -1 {
            pGraphCur = Kit_SopFactor( p.vMemory, 1, Vec_PtrSize(vCut), p.vMemory );
/*
{
    int RetValue;
    RetValue = Bdc_ManDecompose( p->pManDec, pTruth, NULL, Vec_PtrSize(vCut), NULL, 1000 );
    printf( "Graph = %d. Bidec = %d.\n", Kit_GraphNodeNum(pGraphCur), RetValue );
}
*/
            nNodesAdded = Dar_RefactTryGraph( p.pAig, pObj, vCut, pGraphCur, nNodesSaved - !p.pPars.fUseZeros, Required );
            if nNodesAdded > -1 {
                GainCur = nNodesSaved - nNodesAdded;
                if p.GainBest < GainCur || (p.GainBest == GainCur && 
                    (Kit_GraphIsConst(pGraphCur) || Kit_GraphRootLevel(pGraphCur) < Kit_GraphRootLevel(p.pGraphBest))) {
                    p.GainBest = GainCur;
                    if p.pGraphBest {
                        Kit_GraphFree( p.pGraphBest ); }
                    p.pGraphBest = pGraphCur;
                    Vec_PtrCopy( p.vLeavesBest, vCut );
                
                } else {
                    Kit_GraphFree( pGraphCur ); }
            
            } else {
                Kit_GraphFree( pGraphCur ); }
        }
    }

    return p.GainBest;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if a non-PI node has nLevelMin or below.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ObjCutLevelAchieved(vCut *Vec_Ptr_t, nLevelMin int32) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    Vec_PtrForEachEntry( Aig_Obj_t *, vCut, pObj, i )
        if ( !Aig_ObjIsCi(pObj) && int32(pObj.Level) <= nLevelMin )
            return 1;
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Dar_ManRefactor(pAig *Aig_Man_t, pPars *Dar_RefPar_t) int32 {
//    Bar_Progress_t * pProgress;
    var p *Ref_Man_t;
    var vCut, vCut2 *Vec_Ptr_t;
    var pObj, pObjNew *Aig_Obj_t;
    var nNodesOld, nNodeBefore, nNodeAfter, nNodesSaved, nNodesSaved2 int32;
    var i, Required, nLevelMin int32;
    var clkStart, clk abctime;

    // start the manager
    p = Dar_ManRefStart( pAig, pPars );
    // remove dangling nodes
    Aig_ManCleanup( pAig );
    // if updating levels is requested, start fanout and timing
    Aig_ManFanoutStart( pAig );
    if p.pPars.fUpdateLevel {
        Aig_ManStartReverseLevels( pAig, 0 ); }

    // resynthesize each node once
    clkStart = Abc_Clock();
    vCut = Vec_VecEntry( p.vCuts, 0 );
    vCut2 = Vec_VecEntry( p.vCuts, 1 );
    p.nNodesInit = Aig_ManNodeNum(pAig);
    nNodesOld = Vec_PtrSize( pAig.vObjs );
//    pProgress = Bar_ProgressStart( stdout, nNodesOld );
    Aig_ManForEachObj( pAig, pObj, i )
    {
//        Bar_ProgressUpdate( pProgress, i, NULL );
        if !Aig_ObjIsNode(pObj) {
            continue; }
        if i > nNodesOld {
            break; }
        if pAig.Time2Quit && !(i & 256) && Abc_Clock() > pAig.Time2Quit {
            break; }
        Vec_VecClear( p.vCuts );

//printf( "\nConsidering node %d.\n", pObj->Id );
        // get the bounded MFFC size
clk = Abc_Clock();
        nLevelMin = Abc_MaxInt( 0, Aig_ObjLevel(pObj) - 10 );
        nNodesSaved = Aig_NodeMffcSupp( pAig, pObj, nLevelMin, vCut );
        if nNodesSaved < p.pPars.nMffcMin { // too small to consider
        
p.timeCuts += Abc_Clock() - clk;
            continue; 
        }
        p.nNodesTried++;
        if Vec_PtrSize(vCut) > p.pPars.nLeafMax { // get one reconv-driven cut
        
            Aig_ManFindCut( pObj, vCut, p.vCutNodes, p.pPars.nLeafMax, 50 );
            nNodesSaved = Aig_NodeMffcLabelCut( p.pAig, pObj, vCut );
        
        } else if Vec_PtrSize(vCut) < p.pPars.nLeafMax - 2 && p.pPars.fExtend {
            if !(Dar_ObjCutLevelAchieved(vCut, nLevelMin) != 0) {
                if Aig_NodeMffcExtendCut( pAig, pObj, vCut, vCut2 ) {
                    nNodesSaved2 = Aig_NodeMffcLabelCut( p.pAig, pObj, vCut );
                    assert(nNodesSaved2 == nNodesSaved, "nNodesSaved2 == nNodesSaved");
                }
                if Vec_PtrSize(vCut2) > p.pPars.nLeafMax {
                    Vec_PtrClear(vCut2); }
                if Vec_PtrSize(vCut2) > 0 {
                    p.nNodesExten++;
//                    printf( "%d(%d) ", Vec_PtrSize(vCut), Vec_PtrSize(vCut2) );
                }
            
            } else {
                p.nNodesBelow++; }
        }
p.timeCuts += Abc_Clock() - clk;

        // try the cuts
clk = Abc_Clock();
        if pAig.vLevelR {
            Required = Aig_ObjRequiredLevel(pAig, pObj);
        } else {
            Required = ABC_INFINITY;
        }
        Dar_ManRefactorTryCuts( p, pObj, nNodesSaved, Required );
p.timeEval += Abc_Clock() - clk;

        // check the best gain
        if !(p.GainBest > 0 || (p.GainBest == 0 && p.pPars.fUseZeros)) {
            if p.pGraphBest {
                Kit_GraphFree( p.pGraphBest ); }
            continue;
        }
//printf( "\n" );

        // if we end up here, a rewriting step is accepted
        nNodeBefore = Aig_ManNodeNum( pAig );
        pObjNew = Dar_RefactBuildGraph( pAig, p.vLeavesBest, p.pGraphBest );
        assert(int32(Aig_Regular(pObjNew).Level) <= Required, "int32(Aig_Regular(pObjNew).Level) <= Required");
        // replace the node
        Aig_ObjReplace( pAig, pObj, pObjNew, p.pPars.fUpdateLevel );
        // compare the gains
        nNodeAfter = Aig_ManNodeNum( pAig );
        assert(p.GainBest <= nNodeBefore - nNodeAfter, "p.GainBest <= nNodeBefore - nNodeAfter");
        Kit_GraphFree( p.pGraphBest );
        p.nCutsUsed++;
//        break;
    }
p.timeTotal = Abc_Clock() - clkStart;
p.timeOther = p.timeTotal - p.timeCuts - p.timeEval;

//    Bar_ProgressStop( pProgress );
    // put the nodes into the DFS order and reassign their IDs
//    Aig_NtkReassignIds( p );
    // fix the levels
    Aig_ManFanoutStop( pAig );
    if p.pPars.fUpdateLevel {
        Aig_ManStopReverseLevels( pAig ); }
/*
    Aig_ManForEachObj( p->pAig, pObj, i )
        if ( Aig_ObjIsNode(pObj) && Aig_ObjRefs(pObj) == 0 )
        {
            printf( "Unreferenced " );
            Aig_ObjPrintVerbose( pObj, 0 );
            printf( "\n" );
        }
*/
    // remove dangling nodes (they should not be here!)
    Aig_ManCleanup( pAig );

    // stop the rewriting manager
    Dar_ManRefStop( p );
//    Aig_ManCheckPhase( pAig );
    if !Aig_ManCheck( pAig ) {
        fmt.Printf("Dar_ManRefactor: The network check has failed.\n");
        return 0;
    }
    return 1;

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

