package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcRefactor.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Resynthesis based on collapsing and refactoring.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcRefactor.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "bool/dec/dec.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

     
type Abc_ManRef_t struct {
    // user specified parameters
    nNodeSizeMax              int32;      // the limit on the size of the supernode
    nConeSizeMax              int32;      // the limit on the size of the containing cone
    fVerbose              int32;          // the verbosity flag
    // internal data structures
    vVars       *Vec_Ptr_t;             // truth tables
    vFuncs       *Vec_Ptr_t;            // functions
    vMemory       *Vec_Int_t;           // memory
    vCube       *Vec_Str_t;             // temporary
    vForm       *Vec_Int_t;             // temporary
    vVisited       *Vec_Ptr_t;          // temporary
    vLeaves       *Vec_Ptr_t;           // temporary
    // node statistics
    nLastGain              int32;
    nNodesConsidered              int32;
    nNodesRefactored              int32;
    nNodesGained              int32;
    nNodesBeg              int32;
    nNodesEnd              int32;
    // runtime statistics
    timeCut          abctime;
    timeTru          abctime;
    timeDcs          abctime;
    timeSop          abctime;
    timeFact          abctime;
    timeEval          abctime;
    timeRes          abctime;
    timeNtk          abctime;
    timeTotal          abctime;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns function of the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeConeTruth(vVars *Vec_Ptr_t, vFuncs *Vec_Ptr_t, nWordsMax int32, pRoot *Abc_Obj_t, vLeaves *Vec_Ptr_t, vVisited *Vec_Ptr_t) *word {
    var pNode *Abc_Obj_t;
    var pTruth0 *word;
    var pTruth1 *word;
    var pTruth *word = nil;
    var i int32;
    var k int32;
    var nWords int32 = Abc_Truth6WordNum( Vec_PtrSize(vLeaves) );
    // get nodes in the cut without fanins in the DFS order
    Abc_NodeConeCollect( &pRoot, 1, vLeaves, vVisited, 0 );
    // set elementary functions
    Vec_PtrForEachEntry( Abc_Obj_t *, vLeaves, pNode, i )
        pNode.pCopy = (Abc_Obj_t *)Vec_PtrEntry( vVars, i );
    // prepare functions
    for  i = Vec_PtrSize(vFuncs); i < Vec_PtrSize(vVisited); i++  {
        Vec_PtrPush( vFuncs, ABC_ALLOC(word, nWordsMax) );
    }
    // compute functions for the collected nodes
    Vec_PtrForEachEntry( Abc_Obj_t *, vVisited, pNode, i )
    {
        assert(!Abc_ObjIsPi(pNode), "!Abc_ObjIsPi(pNode)");
        pTruth0 = (word *)Abc_ObjFanin0(pNode).pCopy;
        pTruth1 = (word *)Abc_ObjFanin1(pNode).pCopy;
        pTruth  = (word *)Vec_PtrEntry( vFuncs, i );
        if Abc_ObjFaninC0(pNode) {
            if Abc_ObjFaninC1(pNode) {
                for  k = 0; k < nWords; k++  {
                    pTruth[k] = ~pTruth0[k] & ~pTruth1[k];
                }
            } else {
                for  k = 0; k < nWords; k++  {
                    pTruth[k] = ~pTruth0[k] &  pTruth1[k];
                } }
        
        } else {
            if Abc_ObjFaninC1(pNode) {
                for  k = 0; k < nWords; k++  {
                    pTruth[k] =  pTruth0[k] & ~pTruth1[k];
                }
            } else {
                for  k = 0; k < nWords; k++  {
                    pTruth[k] =  pTruth0[k] &  pTruth1[k];
                } }
        }
        pNode.pCopy = (Abc_Obj_t *)pTruth;
    }
    return pTruth;
}
func Abc_NodeConeIsConst0(pTruth *word, nVars int32) int32 {
    var k int32;
    var nWords int32 = Abc_Truth6WordNum( nVars );
    for  k = 0; k < nWords; k++  {
        if pTruth[k] {
            return 0; }
    }
    return 1;
}
func Abc_NodeConeIsConst1(pTruth *word, nVars int32) int32 {
    var k int32;
    var nWords int32 = Abc_Truth6WordNum( nVars );
    for  k = 0; k < nWords; k++  {
        if ~pTruth[k] {
            return 0; }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Resynthesizes the node using refactoring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeRefactor(p *Abc_ManRef_t, pNode *Abc_Obj_t, vFanins *Vec_Ptr_t, fUpdateLevel int32, fUseZeros int32, fUseDcs int32, fVerbose int32) *Dec_Graph_t {
    extern int    Dec_GraphToNetworkCount( Abc_Obj_t * pRoot, Dec_Graph_t * pGraph, int NodeMax, int LevelMax );
    var fVeryVerbose int32;
    var nVars int32  = Vec_PtrSize(vFanins);
    var nWordsMax int32  = Abc_Truth6WordNum(p.nNodeSizeMax);
    var pFForm *Dec_Graph_t;
    var pFanin *Abc_Obj_t;
    var pTruth *word;
    var clk abctime;
    var i, nNodesSaved, nNodesAdded, Required int32;

    p.nNodesConsidered++;

    if fUpdateLevel != 0 {
        Required = Abc_ObjRequiredLevel(pNode);
    } else {
        Required = ABC_INFINITY;
    }

    // get the function of the cut
clk = Abc_Clock();
    pTruth = Abc_NodeConeTruth( p.vVars, p.vFuncs, nWordsMax, pNode, vFanins, p.vVisited );
p.timeTru += Abc_Clock() - clk;
    if pTruth == nil {
        return nil; }

    // always accept the case of constant node
    if (Abc_NodeConeIsConst0(pTruth, nVars) != 0) || (Abc_NodeConeIsConst1(pTruth, nVars) != 0) {
        p.nLastGain = Abc_NodeMffcSize( pNode );
        p.nNodesGained += p.nLastGain;
        p.nNodesRefactored++;
        if Abc_NodeConeIsConst0(pTruth, nVars) != 0 {
            return Dec_GraphCreateConst0();
        } else {
            return Dec_GraphCreateConst1();
        }
    }

    // get the factored form
clk = Abc_Clock();
    pFForm = (Dec_Graph_t *)Kit_TruthToGraph( (unsigned *)pTruth, nVars, p.vMemory );
p.timeFact += Abc_Clock() - clk;

    // mark the fanin boundary 
    // (can mark only essential fanins, belonging to bNodeFunc!)
    Vec_PtrForEachEntry( Abc_Obj_t *, vFanins, pFanin, i )
        pFanin.vFanouts.nSize++;
    // label MFFC with current traversal ID
    Abc_NtkIncrementTravId( pNode.pNtk );
    nNodesSaved = Abc_NodeMffcLabelAig( pNode );
    // unmark the fanin boundary and set the fanins as leaves in the form
    Vec_PtrForEachEntry( Abc_Obj_t *, vFanins, pFanin, i )
    {
        pFanin.vFanouts.nSize--;
        Dec_GraphNode(pFForm, i).pFunc = pFanin;
    }

    // detect how many new nodes will be added (while taking into account reused nodes)
clk = Abc_Clock();
    nNodesAdded = Dec_GraphToNetworkCount( pNode, pFForm, nNodesSaved, Required );
p.timeEval += Abc_Clock() - clk;
    // quit if there is no improvement
    if nNodesAdded == -1 || (nNodesAdded == nNodesSaved && !(fUseZeros != 0)) {
        Dec_GraphFree( pFForm );
        return nil;
    }

    // compute the total gain in the number of nodes
    p.nLastGain = nNodesSaved - nNodesAdded;
    p.nNodesGained += p.nLastGain;
    p.nNodesRefactored++;

    // report the progress
    if fVeryVerbose != 0 {
        fmt.Printf("Node %6s : ", Abc_ObjName(pNode));
        fmt.Printf("Cone = %2d. ", vFanins.nSize);
        fmt.Printf("FF = %2d. ", 1 + Dec_GraphNodeNum(pFForm));
        fmt.Printf("MFFC = %2d. ", nNodesSaved);
        fmt.Printf("Add = %2d. ", nNodesAdded);
        fmt.Printf("GAIN = %2d. ", p.nLastGain);
        fmt.Printf("\n");
    }
    return pFForm;
}


/**Function*************************************************************

  Synopsis    [Starts the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkManRefStart(nNodeSizeMax int32, nConeSizeMax int32, fUseDcs int32, fVerbose int32) *Abc_ManRef_t {
    var p *Abc_ManRef_t;
    p = new(Abc_ManRef_t);
    *p = Abc_ManRef_t{};
    p.vCube        = Vec_StrAlloc( 100 );
    p.vVisited     = Vec_PtrAlloc( 100 );
    p.nNodeSizeMax = nNodeSizeMax;
    p.nConeSizeMax = nConeSizeMax;
    p.fVerbose     = fVerbose;
    p.vVars        = Vec_PtrAllocTruthTables( Abc_MaxInt(nNodeSizeMax, 6) );
    p.vFuncs       = Vec_PtrAlloc( 100 );
    p.vMemory      = Vec_IntAlloc( 1 << 16 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkManRefStop(p *Abc_ManRef_t) {
    Vec_PtrFreeFree( p.vFuncs );
    Vec_PtrFree( p.vVars );
    Vec_IntFree( p.vMemory );
    Vec_PtrFree( p.vVisited );
    Vec_StrFree( p.vCube );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkManRefPrintStats(p *Abc_ManRef_t) {
    fmt.Printf("Refactoring statistics:\n");
    fmt.Printf("Nodes considered  = %8d.\n", p.nNodesConsidered);
    fmt.Printf("Nodes refactored  = %8d.\n", p.nNodesRefactored);
    fmt.Printf("Gain              = %8d. (%6.2f %%).\n", p.nNodesBeg-p.nNodesEnd, 100.0*(p.nNodesBeg-p.nNodesEnd)/p.nNodesBeg);
    ABC_PRT( "Cuts       ", p.timeCut );
    ABC_PRT( "Resynthesis", p.timeRes );
    ABC_PRT( "    BDD    ", p.timeTru );
    ABC_PRT( "    DCs    ", p.timeDcs );
    ABC_PRT( "    SOP    ", p.timeSop );
    ABC_PRT( "    FF     ", p.timeFact );
    ABC_PRT( "    Eval   ", p.timeEval );
    ABC_PRT( "AIG update ", p.timeNtk );
    ABC_PRT( "TOTAL      ", p.timeTotal );
}

/**Function*************************************************************

  Synopsis    [Performs incremental resynthesis of the AIG.]

  Description [Starting from each node, computes a reconvergence-driven cut, 
  derives BDD of the cut function, constructs ISOP, factors the ISOP, 
  and replaces the current implementation of the MFFC of the node by the 
  new factored form, if the number of AIG nodes is reduced and the total
  number of levels of the AIG network is not increated. Returns the
  number of AIG nodes saved.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRefactor(pNtk *Abc_Ntk_t, nNodeSizeMax int32, nConeSizeMax int32, fUpdateLevel int32, fUseZeros int32, fUseDcs int32, fVerbose int32) int32 {
    extern void           Dec_GraphUpdateNetwork( Abc_Obj_t * pRoot, Dec_Graph_t * pGraph, int fUpdateLevel, int nGain );
    var pProgress *ProgressBar;
    var pManRef *Abc_ManRef_t;
    var pManCut *Abc_ManCut_t;
    var pFForm *Dec_Graph_t;
    var vFanins *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var clk abctime;
    var clkStart abctime = Abc_Clock();
    var i, nNodes int32;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // cleanup the AIG
    Abc_AigCleanup((Abc_Aig_t *)pNtk.pManFunc);
    // start the managers
    pManCut = Abc_NtkManCutStart( nNodeSizeMax, nConeSizeMax, 2, 1000 );
    pManRef = Abc_NtkManRefStart( nNodeSizeMax, nConeSizeMax, fUseDcs, fVerbose );
    pManRef.vLeaves   = Abc_NtkManCutReadCutLarge( pManCut );
    // compute the reverse levels if level update is requested
    if fUpdateLevel != 0 {
        Abc_NtkStartReverseLevels( pNtk, 0 ); }

    // resynthesize each node once
    pManRef.nNodesBeg = Abc_NtkNodeNum(pNtk);
    nNodes = Abc_NtkObjNumMax(pNtk);
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    Abc_NtkForEachNode( pNtk, pNode, i )
    {
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
        vFanins = Abc_NodeFindCut( pManCut, pNode, fUseDcs );
pManRef.timeCut += Abc_Clock() - clk;
        // evaluate this cut
clk = Abc_Clock();
        pFForm = Abc_NodeRefactor( pManRef, pNode, vFanins, fUpdateLevel, fUseZeros, fUseDcs, fVerbose );
pManRef.timeRes += Abc_Clock() - clk;
        if pFForm == nil {
            continue; }
        // acceptable replacement found, update the graph
clk = Abc_Clock();
        Dec_GraphUpdateNetwork( pNode, pFForm, fUpdateLevel, pManRef.nLastGain );
pManRef.timeNtk += Abc_Clock() - clk;
        Dec_GraphFree( pFForm );
    }
    Extra_ProgressBarStop( pProgress );
pManRef.timeTotal = Abc_Clock() - clkStart;
    pManRef.nNodesEnd = Abc_NtkNodeNum(pNtk);

    // print statistics of the manager
    if fVerbose != 0 {
        Abc_NtkManRefPrintStats( pManRef ); }
    // delete the managers
    Abc_NtkManCutStop( pManCut );
    Abc_NtkManRefStop( pManRef );
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
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

