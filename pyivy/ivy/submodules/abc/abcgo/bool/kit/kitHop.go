package kit

import "fmt"

/**CFile****************************************************************

  FileName    [kitHop.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Computation kit.]

  Synopsis    [Procedures involving AIGs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 6, 2006.]

  Revision    [$Id: kitHop.c,v 1.00 2006/12/06 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "kit.h"
//#include "aig/hop/hop.h"
//#include "aig/gia/gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphToGiaInternal(pMan *Gia_Man_t, pGraph *Kit_Graph_t, fHash int32) int32 {
    var pNode *Kit_Node_t  = nil;
    var i, pAnd0, pAnd1 int32;
    // check for constant function
    if Kit_GraphIsConst(pGraph) {
        return Abc_LitNotCond( 1, Kit_GraphIsComplement(pGraph) ); }
    // check for a literal
    if Kit_GraphIsVar(pGraph) {
        return Abc_LitNotCond( Kit_GraphVar(pGraph).iFunc, Kit_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Kit_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = Abc_LitNotCond( Kit_GraphNode(pGraph, pNode.eEdge0.Node).iFunc, pNode.eEdge0.fCompl ); 
        pAnd1 = Abc_LitNotCond( Kit_GraphNode(pGraph, pNode.eEdge1.Node).iFunc, pNode.eEdge1.fCompl ); 
        if fHash != 0 {
            pNode.iFunc = Gia_ManHashAnd( pMan, pAnd0, pAnd1 );
        } else {
            pNode.iFunc = Gia_ManAppendAnd2( pMan, pAnd0, pAnd1 ); }
    }
    // complement the result if necessary
    return Abc_LitNotCond( pNode.iFunc, Kit_GraphIsComplement(pGraph) );
}
func Kit_GraphToGia(pMan *Gia_Man_t, pGraph *Kit_Graph_t, vLeaves *Vec_Int_t, fHash int32) int32 {
    var pNode *Kit_Node_t  = nil;
    var i int32;
    // collect the fanins
    // inlining of macro Kit_GraphForEachLeaf
    for  i = 0; (i < (pGraph).nLeaves) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        if vLeaves != nil {
            pNode.iFunc = Vec_IntEntry(vLeaves, i);
        } else {
            pNode.iFunc = Gia_Obj2Lit(pMan, Gia_ManPi(pMan, i));
        }
    }
    // perform strashing
    return Kit_GraphToGiaInternal( pMan, pGraph, fHash );
}
func Kit_TruthToGia(pMan *Gia_Man_t, pTruth *uint32, nVars int32, vMemory *Vec_Int_t, vLeaves *Vec_Int_t, fHash int32) int32 {
    var iLit int32;
    var pGraph *Kit_Graph_t;
    // transform truth table into the decomposition tree
    if vMemory == nil {
        vMemory = Vec_IntAlloc( 0 );
        pGraph = Kit_TruthToGraph( pTruth, nVars, vMemory );
        Vec_IntFree( vMemory );
    
    } else {
        pGraph = Kit_TruthToGraph( pTruth, nVars, vMemory ); }
    if pGraph == nil {
        fmt.Printf("Kit_TruthToGia(): Converting truth table to AIG has failed for function:\n");
        Kit_DsdPrintFromTruth( pTruth, nVars ); fmt.Printf("\n");
    }
    // derive the AIG for the decomposition tree
    iLit = Kit_GraphToGia( pMan, pGraph, vLeaves, fHash );
    Kit_GraphFree( pGraph );
    return iLit;
}
func Kit_TruthToGia2(pMan *Gia_Man_t, pTruth0 *uint32, pTruth1 *uint32, nVars int32, vMemory *Vec_Int_t, vLeaves *Vec_Int_t, fHash int32) int32 {
    var iLit int32;
    var pGraph *Kit_Graph_t;
    // transform truth table into the decomposition tree
    if vMemory == nil {
        vMemory = Vec_IntAlloc( 0 );
        pGraph = Kit_TruthToGraph2( pTruth0, pTruth1, nVars, vMemory );
        Vec_IntFree( vMemory );
    
    } else {
        pGraph = Kit_TruthToGraph2( pTruth0, pTruth1, nVars, vMemory ); }
    if pGraph == nil {
        fmt.Printf("Kit_TruthToGia2(): Converting truth table to AIG has failed for function:\n");
        Kit_DsdPrintFromTruth( pTruth0, nVars ); fmt.Printf("\n");
        Kit_DsdPrintFromTruth( pTruth1, nVars ); fmt.Printf("\n");
    }
    // derive the AIG for the decomposition tree
    iLit = Kit_GraphToGia( pMan, pGraph, vLeaves, fHash );
    Kit_GraphFree( pGraph );
    return iLit;
}

/**Function*************************************************************

  Synopsis    [Transforms the decomposition graph into the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_GraphToHopInternal(pMan *Hop_Man_t, pGraph *Kit_Graph_t) *Hop_Obj_t {
    var pNode *Kit_Node_t  = nil;
    var pAnd0, pAnd1 *Hop_Obj_t;
    var i int32;
    // check for constant function
    if Kit_GraphIsConst(pGraph) {
        return Hop_NotCond( Hop_ManConst1(pMan), Kit_GraphIsComplement(pGraph) ); }
    // check for a literal
    if Kit_GraphIsVar(pGraph) {
        return Hop_NotCond( (Hop_Obj_t *)Kit_GraphVar(pGraph).pFunc, Kit_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Kit_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = Hop_NotCond( (Hop_Obj_t *)Kit_GraphNode(pGraph, pNode.eEdge0.Node).pFunc, pNode.eEdge0.fCompl ); 
        pAnd1 = Hop_NotCond( (Hop_Obj_t *)Kit_GraphNode(pGraph, pNode.eEdge1.Node).pFunc, pNode.eEdge1.fCompl ); 
        pNode.pFunc = Hop_And( pMan, pAnd0, pAnd1 );
    }
    // complement the result if necessary
    return Hop_NotCond( (Hop_Obj_t *)pNode.pFunc, Kit_GraphIsComplement(pGraph) );
}
func Kit_GraphToHop(pMan *Hop_Man_t, pGraph *Kit_Graph_t) *Hop_Obj_t {
    var pNode *Kit_Node_t  = nil;
    var i int32;
    // collect the fanins
    // inlining of macro Kit_GraphForEachLeaf
    for  i = 0; (i < (pGraph).nLeaves) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pNode.pFunc = Hop_IthVar( pMan, i );
    }
    // perform strashing
    return Kit_GraphToHopInternal( pMan, pGraph );
}
func Kit_TruthToHop(pMan *Hop_Man_t, pTruth *uint32, nVars int32, vMemory *Vec_Int_t) *Hop_Obj_t {
    var pObj *Hop_Obj_t;
    var pGraph *Kit_Graph_t;
    // transform truth table into the decomposition tree
    if vMemory == nil {
        vMemory = Vec_IntAlloc( 0 );
        pGraph = Kit_TruthToGraph( pTruth, nVars, vMemory );
        Vec_IntFree( vMemory );
    
    } else {
        pGraph = Kit_TruthToGraph( pTruth, nVars, vMemory ); }
    if pGraph == nil {
        fmt.Printf("Kit_TruthToHop(): Converting truth table to AIG has failed for function:\n");
        Kit_DsdPrintFromTruth( pTruth, nVars ); fmt.Printf("\n");
    }
    // derive the AIG for the decomposition tree
    pObj = Kit_GraphToHop( pMan, pGraph );
    Kit_GraphFree( pGraph );
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_CoverToHop(pMan *Hop_Man_t, vCover *Vec_Int_t, nVars int32, vMemory *Vec_Int_t) *Hop_Obj_t {
    var pGraph *Kit_Graph_t;
    var pFunc *Hop_Obj_t;
    // perform factoring
    Vec_IntClear( vMemory );
    pGraph = Kit_SopFactor( vCover, 0, nVars, vMemory );
    // convert graph to the AIG
    pFunc = Kit_GraphToHop( pMan, pGraph );
    Kit_GraphFree( pGraph );
    return pFunc;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

