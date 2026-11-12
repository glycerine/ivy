package dec

/**CFile****************************************************************

  FileName    [decAbc.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Interface between the decomposition package and ABC network.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: decAbc.c,v 1.1 2003/05/22 19:20:05 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/ivy/ivy.h"
//#include "dec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description [AIG nodes for the fanins should be assigned to pNode->pFunc
  of the leaves of the graph before calling this procedure.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphToNetwork(pNtk *Abc_Ntk_t, pGraph *Dec_Graph_t) *Abc_Obj_t {
    var pAnd0, pAnd1 *Abc_Obj_t;
    var pNode *Dec_Node_t  = nil; // Suppress "might be used uninitialized"
    var i int32;
    // check for constant function
    if Dec_GraphIsConst(pGraph) {
        return Abc_ObjNotCond( Abc_AigConst1(pNtk), Dec_GraphIsComplement(pGraph) ); }
    // check for a literal
    if Dec_GraphIsVar(pGraph) {
        return Abc_ObjNotCond( (Abc_Obj_t *)Dec_GraphVar(pGraph).pFunc, Dec_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Dec_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = Abc_ObjNotCond( (Abc_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge0.Node).pFunc, pNode.eEdge0.fCompl ); 
        pAnd1 = Abc_ObjNotCond( (Abc_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge1.Node).pFunc, pNode.eEdge1.fCompl ); 
        pNode.pFunc = Abc_AigAnd( (Abc_Aig_t *)pNtk.pManFunc, pAnd0, pAnd1 );
    }
    // complement the result if necessary
    return Abc_ObjNotCond( (Abc_Obj_t *)pNode.pFunc, Dec_GraphIsComplement(pGraph) );
}

/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_SopToAig(pNtk *Abc_Ntk_t, pSop string, vFaninAigs *Vec_Ptr_t) *Abc_Obj_t {
    var pFunc *Abc_Obj_t;
    var pFForm *Dec_Graph_t;
    var pNode *Dec_Node_t;
    var i int32;
    pFForm = Dec_Factor( pSop );
    // inlining of macro Dec_GraphForEachLeaf
    for  i = 0; (i < (pFForm).nLeaves) && (((pNode) = Dec_GraphNode(pFForm, i)), 1); i++  {
        pNode.pFunc = Vec_PtrEntry( vFaninAigs, i );
    }
    pFunc = Dec_GraphToNetwork( pNtk, pFForm );
    Dec_GraphFree( pFForm );
    return pFunc;
}

/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphToAig(pNtk *Abc_Ntk_t, pFForm *Dec_Graph_t, vFaninAigs *Vec_Ptr_t) *Abc_Obj_t {
    var pFunc *Abc_Obj_t;
    var pNode *Dec_Node_t;
    var i int32;
    // inlining of macro Dec_GraphForEachLeaf
    for  i = 0; (i < (pFForm).nLeaves) && (((pNode) = Dec_GraphNode(pFForm, i)), 1); i++  {
        pNode.pFunc = Vec_PtrEntry( vFaninAigs, i );
    }
    pFunc = Dec_GraphToNetwork( pNtk, pFForm );
    return pFunc;
}

/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description [AIG nodes for the fanins should be assigned to pNode->pFunc
  of the leaves of the graph before calling this procedure.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphToNetworkNoStrash(pNtk *Abc_Ntk_t, pGraph *Dec_Graph_t) *Abc_Obj_t {
    var pAnd, pAnd0, pAnd1 *Abc_Obj_t;
    var pNode *Dec_Node_t  = nil; // Suppress "might be used uninitialized"
    var i int32;
    // check for constant function
    if Dec_GraphIsConst(pGraph) {
        return Abc_ObjNotCond( Abc_AigConst1(pNtk), Dec_GraphIsComplement(pGraph) ); }
    // check for a literal
    if Dec_GraphIsVar(pGraph) {
        return Abc_ObjNotCond( (Abc_Obj_t *)Dec_GraphVar(pGraph).pFunc, Dec_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Dec_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = Abc_ObjNotCond( (Abc_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge0.Node).pFunc, pNode.eEdge0.fCompl ); 
        pAnd1 = Abc_ObjNotCond( (Abc_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge1.Node).pFunc, pNode.eEdge1.fCompl ); 
//        pNode->pFunc = Abc_AigAnd( (Abc_Aig_t *)pNtk->pManFunc, pAnd0, pAnd1 );
        pAnd = Abc_NtkCreateNode( pNtk );
        Abc_ObjAddFanin( pAnd, pAnd0 );
        Abc_ObjAddFanin( pAnd, pAnd1 );
        pNode.pFunc = pAnd;
    }
    // complement the result if necessary
    return Abc_ObjNotCond( (Abc_Obj_t *)pNode.pFunc, Dec_GraphIsComplement(pGraph) );
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
func Dec_GraphToNetworkCount(pRoot *Abc_Obj_t, pGraph *Dec_Graph_t, NodeMax int32, LevelMax int32) int32 {
    var pMan *Abc_Aig_t  = (Abc_Aig_t *)pRoot.pNtk.pManFunc;
    var pNode, pNode0, pNode1 *Dec_Node_t;
    var pAnd, pAnd0, pAnd1 *Abc_Obj_t;
    var i, Counter, LevelNew, LevelOld int32;
    // check for constant function or a literal
    if Dec_GraphIsConst(pGraph) || Dec_GraphIsVar(pGraph) {
        return 0; }
    // set the levels of the leaves
    // inlining of macro Dec_GraphForEachLeaf
    for  i = 0; (i < (pGraph).nLeaves) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        pNode.Level = Abc_ObjRegular((Abc_Obj_t *)pNode.pFunc).Level;
    }
    // compute the AIG size after adding the internal nodes
    Counter = 0;
    // inlining of macro Dec_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        // get the children of this node
        pNode0 = Dec_GraphNode( pGraph, pNode.eEdge0.Node );
        pNode1 = Dec_GraphNode( pGraph, pNode.eEdge1.Node );
        // get the AIG nodes corresponding to the children 
        pAnd0 = (Abc_Obj_t *)pNode0.pFunc; 
        pAnd1 = (Abc_Obj_t *)pNode1.pFunc; 
        if pAnd0 && pAnd1 {
            // if they are both present, find the resulting node
            pAnd0 = Abc_ObjNotCond( pAnd0, pNode.eEdge0.fCompl );
            pAnd1 = Abc_ObjNotCond( pAnd1, pNode.eEdge1.fCompl );
            pAnd  = Abc_AigAndLookup( pMan, pAnd0, pAnd1 );
            // return -1 if the node is the same as the original root
            if Abc_ObjRegular(pAnd) == pRoot {
                return -1; }
        
        } else {
            pAnd = nil; }
        // count the number of added nodes
        if pAnd == nil || Abc_NodeIsTravIdCurrent(Abc_ObjRegular(pAnd)) {
            if Counter++; Counter > NodeMax {
                return -1; }
        }
        // count the number of new levels
        LevelNew = 1 + Abc_MaxInt( pNode0.Level, pNode1.Level );
        if pAnd {
            if Abc_ObjRegular(pAnd) == Abc_AigConst1(pRoot.pNtk) {
                LevelNew = 0;
            } else if Abc_ObjRegular(pAnd) == Abc_ObjRegular(pAnd0) {
                LevelNew = int32(Abc_ObjRegular(pAnd0).Level);
            } else if Abc_ObjRegular(pAnd) == Abc_ObjRegular(pAnd1) {
                LevelNew = int32(Abc_ObjRegular(pAnd1).Level); }
            LevelOld = int32(Abc_ObjRegular(pAnd).Level);
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

  Synopsis    [Replaces MFFC of the node by the new factored form.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphUpdateNetwork(pRoot *Abc_Obj_t, pGraph *Dec_Graph_t, fUpdateLevel int32, nGain int32) {
    extern Abc_Obj_t *    Dec_GraphToNetwork( Abc_Ntk_t * pNtk, Dec_Graph_t * pGraph );
    var pRootNew *Abc_Obj_t;
    var pNtk *Abc_Ntk_t  = pRoot.pNtk;
    var nNodesNew, nNodesOld int32;
    nNodesOld = Abc_NtkNodeNum(pNtk);
    // create the new structure of nodes
    pRootNew = Dec_GraphToNetwork( pNtk, pGraph );
    // remove the old nodes
    Abc_AigReplace( (Abc_Aig_t *)pNtk.pManFunc, pRoot, pRootNew, fUpdateLevel );
    // compare the gains
    nNodesNew = Abc_NtkNodeNum(pNtk);
    //assert( nGain <= nNodesOld - nNodesNew );
}


/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphToNetworkAig(pMan *Hop_Man_t, pGraph *Dec_Graph_t) *Hop_Obj_t {
    var pNode *Dec_Node_t  = nil; // Suppress "might be used uninitialized"
    var pAnd0, pAnd1 *Hop_Obj_t;
    var i int32;
    // check for constant function
    if Dec_GraphIsConst(pGraph) {
        return Hop_NotCond( Hop_ManConst1(pMan), Dec_GraphIsComplement(pGraph) ); }
    // check for a literal
    if Dec_GraphIsVar(pGraph) {
        return Hop_NotCond( (Hop_Obj_t *)Dec_GraphVar(pGraph).pFunc, Dec_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Dec_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = Hop_NotCond( (Hop_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge0.Node).pFunc, pNode.eEdge0.fCompl ); 
        pAnd1 = Hop_NotCond( (Hop_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge1.Node).pFunc, pNode.eEdge1.fCompl ); 
        pNode.pFunc = Hop_And( pMan, pAnd0, pAnd1 );
    }
    // complement the result if necessary
    return Hop_NotCond( (Hop_Obj_t *)pNode.pFunc, Dec_GraphIsComplement(pGraph) );
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphFactorSop(pMan *Hop_Man_t, pSop string) *Hop_Obj_t {
    var pFunc *Hop_Obj_t;
    var pFForm *Dec_Graph_t;
    var pNode *Dec_Node_t;
    var i int32;
    // perform factoring
    pFForm = Dec_Factor( pSop );
    // collect the fanins
    // inlining of macro Dec_GraphForEachLeaf
    for  i = 0; (i < (pFForm).nLeaves) && (((pNode) = Dec_GraphNode(pFForm, i)), 1); i++  {
        pNode.pFunc = Hop_IthVar( pMan, i );
    }
    // perform strashing
    pFunc = Dec_GraphToNetworkAig( pMan, pFForm );
    Dec_GraphFree( pFForm );
    return pFunc;
}

/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphToNetworkIvy(pMan *Ivy_Man_t, pGraph *Dec_Graph_t) *Ivy_Obj_t {
    var pNode *Dec_Node_t  = nil; // Suppress "might be used uninitialized"
    var pAnd0, pAnd1 *Ivy_Obj_t;
    var i int32;
    // check for constant function
    if Dec_GraphIsConst(pGraph) {
        return Ivy_NotCond( Ivy_ManConst1(pMan), Dec_GraphIsComplement(pGraph) ); }
    // check for a literal
    if Dec_GraphIsVar(pGraph) {
        return Ivy_NotCond( (Ivy_Obj_t *)Dec_GraphVar(pGraph).pFunc, Dec_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Dec_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = Ivy_NotCond( (Ivy_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge0.Node).pFunc, pNode.eEdge0.fCompl ); 
        pAnd1 = Ivy_NotCond( (Ivy_Obj_t *)Dec_GraphNode(pGraph, pNode.eEdge1.Node).pFunc, pNode.eEdge1.fCompl ); 
        pNode.pFunc = Ivy_And( pMan, pAnd0, pAnd1 );
    }
    // complement the result if necessary
    return Ivy_NotCond( (Ivy_Obj_t *)pNode.pFunc, Dec_GraphIsComplement(pGraph) );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

