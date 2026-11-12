package lpk

/**CFile****************************************************************

  FileName    [lpkMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkMap.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "lpkInt.h"

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
func Lpk_MapPrimeInternal(pIfMan *If_Man_t, pGraph *Kit_Graph_t) *If_Obj_t {
    var pNode *Kit_Node_t  = nil; // Suppress "might be used uninitialized"
    var pAnd0, pAnd1 *If_Obj_t;
    var i int32;
    // check for constant function
    if Kit_GraphIsConst(pGraph) {
        return If_ManConst1(pIfMan); }
    // check for a literal
    if Kit_GraphIsVar(pGraph) {
        return (If_Obj_t *)Kit_GraphVar(pGraph).pFunc; }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Kit_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pAnd0 = (If_Obj_t *)Kit_GraphNode(pGraph, pNode.eEdge0.Node).pFunc; 
        pAnd1 = (If_Obj_t *)Kit_GraphNode(pGraph, pNode.eEdge1.Node).pFunc; 
        pNode.pFunc = If_ManCreateAnd( pIfMan, 
            If_NotCond( If_Regular(pAnd0), If_IsComplement(pAnd0) ^ pNode.eEdge0.fCompl ), 
            If_NotCond( If_Regular(pAnd1), If_IsComplement(pAnd1) ^ pNode.eEdge1.fCompl ) );
    }
    return (If_Obj_t *)pNode.pFunc;
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_MapPrime(p *Lpk_Man_t, pTruth *uint32, nVars int32, ppLeaves **If_Obj_t) *If_Obj_t {
    var pGraph *Kit_Graph_t;
    var pNode *Kit_Node_t;
    var pRes *If_Obj_t;
    var i int32;
    // derive the factored form
    pGraph = Kit_TruthToGraph( pTruth, nVars, p.vCover );
    if pGraph == nil {
        return nil; }
    // collect the fanins
    // inlining of macro Kit_GraphForEachLeaf
    for  i = 0; (i < (pGraph).nLeaves) && (((pNode) = Kit_GraphNode(pGraph, i)), 1); i++  {
        pNode.pFunc = ppLeaves[i];
    }
    // perform strashing
    pRes = Lpk_MapPrimeInternal( p.pIfMan, pGraph );
    pRes = If_NotCond( pRes, Kit_GraphIsComplement(pGraph) );
    Kit_GraphFree( pGraph );
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Prepares the mapping manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_MapTree_rec(p *Lpk_Man_t, pNtk *Kit_DsdNtk_t, ppLeaves **If_Obj_t, iLit int32, pResult *If_Obj_t) *If_Obj_t {
    var pObj *Kit_DsdObj_t;
    var pObjNew *If_Obj_t = nil;
    var pObjNew2 *If_Obj_t = nil;
    var pFansNew [16]*If_Obj_t;
    var i, iLitFanin uint32;

    assert(iLit >= 0, "iLit >= 0");

    // consider the case of a gate
    pObj = Kit_DsdNtkObj( pNtk, Abc_Lit2Var(iLit) );
    if pObj == nil {
        pObjNew = ppLeaves[Abc_Lit2Var(iLit)];
        return If_NotCond( pObjNew, Abc_LitIsCompl(iLit) );
    }
    if pObj.Type == KIT_DSD_CONST1 {
        return If_NotCond( If_ManConst1(p.pIfMan), Abc_LitIsCompl(iLit) );
    }
    if pObj.Type == KIT_DSD_VAR {
        pObjNew = ppLeaves[Abc_Lit2Var(pObj.pFans[0])];
        return If_NotCond( pObjNew, Abc_LitIsCompl(iLit) ^ Abc_LitIsCompl(pObj.pFans[0]) );
    }
    if pObj.Type == KIT_DSD_AND {
        assert(pObj.nFans == 2, "pObj.nFans == 2");
        pFansNew[0] = Lpk_MapTree_rec( p, pNtk, ppLeaves, pObj.pFans[0], nil );
        pFansNew[1] = pResult? pResult : Lpk_MapTree_rec( p, pNtk, ppLeaves, pObj.pFans[1], nil );
        if pFansNew[0] == nil || pFansNew[1] == nil {
            return nil; }
        pObjNew = If_ManCreateAnd( p.pIfMan, pFansNew[0], pFansNew[1] ); 
        return If_NotCond( pObjNew, Abc_LitIsCompl(iLit) );
    }
    if pObj.Type == KIT_DSD_XOR {
        var fCompl int32  = Abc_LitIsCompl(iLit);
        assert(pObj.nFans == 2, "pObj.nFans == 2");
        pFansNew[0] = Lpk_MapTree_rec( p, pNtk, ppLeaves, pObj.pFans[0], nil );
        pFansNew[1] = pResult? pResult : Lpk_MapTree_rec( p, pNtk, ppLeaves, pObj.pFans[1], nil );
        if pFansNew[0] == nil || pFansNew[1] == nil {
            return nil; }
        fCompl ^= If_IsComplement(pFansNew[0]) ^ If_IsComplement(pFansNew[1]);
        pObjNew = If_ManCreateXor( p.pIfMan, If_Regular(pFansNew[0]), If_Regular(pFansNew[1]) );
        return If_NotCond( pObjNew, fCompl );
    }
    assert(pObj.Type == KIT_DSD_PRIME, "pObj.Type == KIT_DSD_PRIME");
    p.nBlocks[pObj.nFans]++;

    // solve for the inputs
    // inlining of macro Kit_DsdObjForEachFanin
    for  i = 0; (i < int32((pObj).nFans)) && ((iLitFanin) = (pObj).pFans[i], 1); i++  {
        if i == 0 {
            pFansNew[i] = pResult? pResult : Lpk_MapTree_rec( p, pNtk, ppLeaves, iLitFanin, nil );
        } else {
            pFansNew[i] = Lpk_MapTree_rec( p, pNtk, ppLeaves, iLitFanin, nil ); }
        if pFansNew[i] == nil {
            return nil; }
    }
/* 
    if ( !p->fCofactoring && p->pPars->nVarsShared > 0 && (int)pObj->nFans > p->pPars->nLutSize )
    {
        pObjNew = Lpk_MapTreeMulti( p, Kit_DsdObjTruth(pObj), pObj->nFans, pFansNew );
        return If_NotCond( pObjNew, Abc_LitIsCompl(iLit) );
    }
*/
/*
    if ( (int)pObj->nFans > p->pPars->nLutSize )
    {
        pObjNew2 = Lpk_MapTreeMux_rec( p, Kit_DsdObjTruth(pObj), pObj->nFans, pFansNew );
//        if ( pObjNew2 )
//            return If_NotCond( pObjNew2, Abc_LitIsCompl(iLit) );
    }
*/

    // find best cofactoring variable
    if p.pPars.nVarsShared > 0 && int32(pObj.nFans) > p.pPars.nLutSize {
        pObjNew2 = Lpk_MapSuppRedDec_rec( p, Kit_DsdObjTruth(pObj), pObj.nFans, pFansNew );
        if pObjNew2 != nil {
            return If_NotCond( pObjNew2, Abc_LitIsCompl(iLit) ); }
    }

    pObjNew = Lpk_MapPrime( p, Kit_DsdObjTruth(pObj), pObj.nFans, pFansNew );

    // add choice
    if (pObjNew != nil) && (pObjNew2 != nil) {
        If_ObjSetChoice( If_Regular(pObjNew), If_Regular(pObjNew2) );
        If_ManCreateChoice( p.pIfMan, If_Regular(pObjNew) );
    }
    return If_NotCond( pObjNew, Abc_LitIsCompl(iLit) );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

