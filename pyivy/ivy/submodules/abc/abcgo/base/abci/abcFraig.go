package abci

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [abcFraig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures interfacing with the FRAIG package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcFraig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "proof/fraig/fraig.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Abc_Ntk_t *    Abc_NtkFromFraig( Fraig_Man_t * pMan, Abc_Ntk_t * pNtk );

// static Abc_Ntk_t *    Abc_NtkFromFraig2( Fraig_Man_t * pMan, Abc_Ntk_t * pNtk );

// static Abc_Obj_t *    Abc_NodeFromFraig_rec( Abc_Ntk_t * pNtkNew, Fraig_Node_t * pNodeFraig );

// static void           Abc_NtkFromFraig2_rec( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode, Vec_Ptr_t * vNodeReprs );

// extern Fraig_Node_t * Abc_NtkToFraigExdc( Fraig_Man_t * pMan, Abc_Ntk_t * pNtkMain, Abc_Ntk_t * pNtkExdc );

// static void           Abc_NtkFraigRemapUsingExdc( Fraig_Man_t * pMan, Abc_Ntk_t * pNtk );


// static int            Abc_NtkFraigTrustCheck( Abc_Ntk_t * pNtk );

// static void           Abc_NtkFraigTrustOne( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkNew );

// static Abc_Obj_t *    Abc_NodeFraigTrust( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode );

 
////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Interfaces the network with the FRAIG package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraig(pNtk *Abc_Ntk_t, pParams any, fAllNodes int32, fExdc int32) *Abc_Ntk_t {
    var pPars *Fraig_Params_t  = (Fraig_Params_t *)pParams;
    var pNtkNew *Abc_Ntk_t;
    var pMan *Fraig_Man_t; 
    // check if EXDC is present
    if (fExdc != 0) && pNtk.pExdc == nil {
        fExdc = 0;
fmt.Printf("Warning: Networks has no EXDC.\n"); }
    // perform fraiging
    pMan = (Fraig_Man_t *)Abc_NtkToFraig( pNtk, pParams, fAllNodes, fExdc ); 
    // add algebraic choices
//    if ( pPars->fChoicing )
//        Fraig_ManAddChoices( pMan, 0, 6 );
    // prove the miter if asked to
    if pPars.fTryProve {
        Fraig_ManProveMiter( pMan ); }
    // reconstruct FRAIG in the new network
    if fExdc != 0 { 
        pNtkNew = Abc_NtkFromFraig2( pMan, pNtk );
    } else {
        pNtkNew = Abc_NtkFromFraig( pMan, pNtk ); }
    Fraig_ManFree( pMan );
    if pNtk.pExdc {
        pNtkNew.pExdc = Abc_NtkDup( pNtk.pExdc ); }
    // make sure that everything is okay
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkFraig: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Transforms the strashed network into FRAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToFraig(pNtk *Abc_Ntk_t, pParams any, fAllNodes int32, fExdc int32) any {
    var fInternal int32  = ((Fraig_Params_t *)pParams).fInternal;
    var pMan *Fraig_Man_t;
    var pProgress *ProgressBar  = nil;
    var vNodes *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var i int32;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    // create the FRAIG manager
    pMan = Fraig_ManCreate( (Fraig_Params_t *)pParams );

    // map the constant node
    Abc_NtkCleanCopy( pNtk );
    Abc_AigConst1(pNtk).pCopy = (Abc_Obj_t *)Fraig_ManReadConst1(pMan);
    // create PIs and remember them in the old nodes
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pNode.pCopy = (Abc_Obj_t *)Fraig_ManReadIthVar(pMan, i);
    }
 
    // perform strashing
    vNodes = Abc_AigDfs( pNtk, fAllNodes, 0 );
    if !(fInternal != 0) {
        pProgress = Extra_ProgressBarStart( os.Stdout, vNodes.nSize ); }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Abc_ObjFaninNum(pNode) == 0 {
            continue; }
        if pProgress != nil { 
            Extra_ProgressBarUpdate( pProgress, i, nil ); }
        pNode.pCopy = (Abc_Obj_t *)Fraig_NodeAnd( pMan, 
                // inlining of macro Fraig_NotCond
                ((Fraig_Node_t *)((ABC_PTRUINT_T)(Abc_ObjFanin0(pNode).pCopy) ^ (int32(Abc_ObjFaninC0(pNode))))),
                // inlining of macro Fraig_NotCond
                ((Fraig_Node_t *)((ABC_PTRUINT_T)(Abc_ObjFanin1(pNode).pCopy) ^ (int32(Abc_ObjFaninC1(pNode))))) );
    }
    if pProgress != nil {
        Extra_ProgressBarStop( pProgress ); }
    Vec_PtrFree( vNodes );

    // use EXDC to change the mapping of nodes into FRAIG nodes
    if fExdc != 0 {
        Abc_NtkFraigRemapUsingExdc( pMan, pNtk ); }

    // set the primary outputs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Fraig_ManSetPo( pMan, (Fraig_Node_t *)Abc_ObjNotCond( Abc_ObjFanin0(pNode).pCopy, Abc_ObjFaninC0(pNode) ) );
    }
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Derives EXDC node for the given network.]

  Description [Assumes that EXDCs of all POs are the same.
  Returns the EXDC of the first PO.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToFraigExdc(pMan *Fraig_Man_t, pNtkMain *Abc_Ntk_t, pNtkExdc *Abc_Ntk_t) *Fraig_Node_t {
    var pNtkStrash *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var gResult *Fraig_Node_t;
    var ppNames *string;
    var i, k int32;
    // strash the EXDC network
    pNtkStrash = Abc_NtkStrash( pNtkExdc, 0, 0, 0 );
    Abc_NtkCleanCopy( pNtkStrash );
    Abc_AigConst1(pNtkStrash).pCopy = (Abc_Obj_t *)Fraig_ManReadConst1(pMan);
    // set the mapping of the PI nodes
    ppNames = Abc_NtkCollectCioNames( pNtkMain, 0 );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtkStrash)) && (((pObj) = Abc_NtkCi(pNtkStrash, i)), 1); i++  {
        for  k = 0; k < Abc_NtkCiNum(pNtkMain); k++  {
            if int32(strings.Compare(Abc_ObjName(pObj), ppNames[k])) == 0 {
                pObj.pCopy = (Abc_Obj_t *)Fraig_ManReadIthVar(pMan, k);
                break;
            }
        }
        assert(pObj.pCopy != nil, "pObj.pCopy != nil");
    }
    // inlining of macro ABC_FREE
    ((ppNames) ? (free((char *) (ppNames)), (ppNames) = 0) : 0);
    // build FRAIG for each node
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtkStrash).vObjs)) && (((pObj) = Abc_NtkObj(pNtkStrash, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        pObj.pCopy = (Abc_Obj_t *)Fraig_NodeAnd( pMan, 
                // inlining of macro Fraig_NotCond
                ((Fraig_Node_t *)((ABC_PTRUINT_T)(Abc_ObjFanin0(pObj).pCopy) ^ (int32(Abc_ObjFaninC0(pObj))))),
                // inlining of macro Fraig_NotCond
                ((Fraig_Node_t *)((ABC_PTRUINT_T)(Abc_ObjFanin1(pObj).pCopy) ^ (int32(Abc_ObjFaninC1(pObj))))) ); }
    }
    // get the EXDC to be returned
    pObj = Abc_NtkPo( pNtkStrash, 0 );
    // inlining of macro Fraig_NotCond
    gResult = ((Fraig_Node_t *)((ABC_PTRUINT_T)(Abc_ObjFanin0(pObj).pCopy) ^ (int32(Abc_ObjFaninC0(pObj)))));
    Abc_NtkDelete( pNtkStrash );
    return gResult;
}


/**Function*************************************************************

  Synopsis    [Changes mapping of the old nodes into FRAIG nodes using EXDC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigRemapUsingExdc(pMan *Fraig_Man_t, pNtk *Abc_Ntk_t) {
    var gNodeNew, gNodeExdc *Fraig_Node_t;
    var tTable *stmm_table;
    var gen *stmm_generator;
    var pNode, pNodeBest *Abc_Obj_t;
    var pClass *Abc_Obj_t;
    var ppSlot **Abc_Obj_t;
    var vNexts *Vec_Ptr_t;
    var i int32;

    // get the global don't-cares
    assert(pNtk.pExdc, "pNtk.pExdc");
    gNodeExdc = Abc_NtkToFraigExdc( pMan, pNtk, pNtk.pExdc );

    // save the next pointers
    vNexts = Vec_PtrStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Vec_PtrWriteEntry( vNexts, pNode.Id, pNode.pNext ); }
    }

    // find the classes of AIG nodes which have FRAIG nodes assigned
    Abc_NtkCleanNext( pNtk );
    tTable = stmm_init_table(stmm_ptrcmp,stmm_ptrhash);
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else
        if pNode.pCopy {
            // inlining of macro Fraig_Not
            gNodeNew = Fraig_NodeAnd( pMan, (Fraig_Node_t *)pNode.pCopy, ((Fraig_Node_t *)((ABC_PTRUINT_T)(gNodeExdc) ^ 01))  );
            // inlining of macro Fraig_Regular
            if !stmm_find_or_add( tTable, (char *)((Fraig_Node_t *)((ABC_PTRUINT_T)(gNodeNew) & ~01)) , (char ***)&ppSlot ) {
                *ppSlot = nil; }
            pNode.pNext = *ppSlot;
            *ppSlot = pNode;
        }
    }

    // for reach non-trival class, find the node with minimum level, and replace other nodes by it
    Abc_AigSetNodePhases( pNtk );
    // inlining of macro stmm_foreach_item
    for gen=stmm_init_gen(tTable); stmm_gen(gen,(char **)&gNodeNew,(char **)&pClass) || (stmm_free_gen(gen),0); {
        if pClass.pNext == nil {
            continue; }
        // find the node with minimum level
        pNodeBest = pClass;
        for  pNode = pClass.pNext; (pNode != nil); pNode = pNode.pNext  {
            if pNodeBest.Level > pNode.Level {
                 pNodeBest = pNode; }
        }
        // remap the class nodes
        for  pNode = pClass; (pNode != nil); pNode = pNode.pNext  {
            pNode.pCopy = Abc_ObjNotCond( pNodeBest.pCopy, pNode.fPhase ^ pNodeBest.fPhase );
        }
    }
    stmm_free_table( tTable );

    // restore the next pointers
    Abc_NtkCleanNext( pNtk );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        pNode.pNext = (Abc_Obj_t *)Vec_PtrEntry( vNexts, pNode.Id ); }
    }
    Vec_PtrFree( vNexts );
}

/**Function*************************************************************

  Synopsis    [Transforms FRAIG into strashed network with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromFraig(pMan *Fraig_Man_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pProgress *ProgressBar;
    var pNtkNew *Abc_Ntk_t;
    var pNode, pNodeNew *Abc_Obj_t;
    var i int32;
    // create the new network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );
    // make the mapper point to the new network
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Fraig_NodeSetData1( Fraig_ManReadIthVar(pMan, i), (Fraig_Node_t *)pNode.pCopy );
    }
    // set the constant node
    Fraig_NodeSetData1( Fraig_ManReadConst1(pMan), (Fraig_Node_t *)Abc_AigConst1(pNtkNew) );
    // process the nodes in topological order
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkCoNum(pNtk) );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        pNodeNew = Abc_NodeFromFraig_rec( pNtkNew, Fraig_ManReadOutputs(pMan)[i] );
        Abc_ObjAddFanin( pNode.pCopy, pNodeNew );
    }
    Extra_ProgressBarStop( pProgress );
    Abc_NtkReassignIds( pNtkNew );
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Transforms into AIG one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeFromFraig_rec(pNtkNew *Abc_Ntk_t, pNodeFraig *Fraig_Node_t) *Abc_Obj_t {
    var pRes, pRes0, pRes1, pResMin, pResCur *Abc_Obj_t;
    // inlining of macro Fraig_Regular
    Fraig_Node_t * pNodeTemp, * pNodeFraigR = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNodeFraig) & ~01)) ;
    var ppTail *any;
    // check if the node was already considered
    if pRes = (Abc_Obj_t *)Fraig_NodeReadData1(pNodeFraigR) {
        // inlining of macro Fraig_IsComplement
        return Abc_ObjNotCond( pRes, ((int32(((ABC_PTRUINT_T) (pNodeFraig) & 01)))) ); }
    // solve the children
    pRes0 = Abc_NodeFromFraig_rec( pNtkNew, Fraig_NodeReadOne(pNodeFraigR) );
    pRes1 = Abc_NodeFromFraig_rec( pNtkNew, Fraig_NodeReadTwo(pNodeFraigR) );
    // derive the new node
    pRes = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, pRes0, pRes1 );
    pRes.fPhase = Fraig_NodeReadSimInv( pNodeFraigR );
    // if the node has an equivalence class, find its representative
    if Fraig_NodeReadRepr(pNodeFraigR) == nil && Fraig_NodeReadNextE(pNodeFraigR) != nil {
        // go through the FRAIG nodes belonging to this equivalence class
        // and find the representative node (the node with the smallest level)
        pResMin = pRes;
        for  pNodeTemp = Fraig_NodeReadNextE(pNodeFraigR); pNodeTemp; pNodeTemp = Fraig_NodeReadNextE(pNodeTemp)  {
            assert(Fraig_NodeReadData1(pNodeTemp) == nil, "Fraig_NodeReadData1(pNodeTemp) == nil");
            pResCur = Abc_NodeFromFraig_rec( pNtkNew, pNodeTemp );
            if pResMin.Level > pResCur.Level {
                pResMin = pResCur; }
        }
        // link the nodes in such a way that representative goes first
        ppTail = &pResMin.pData;
        if pRes != pResMin {
            *ppTail = pRes;
            ppTail = &pRes.pData;
        }
        for  pNodeTemp = Fraig_NodeReadNextE(pNodeFraigR); pNodeTemp; pNodeTemp = Fraig_NodeReadNextE(pNodeTemp)  {
            pResCur = (Abc_Obj_t *)Fraig_NodeReadData1(pNodeTemp);
            assert((pResCur != nil), "pResCur");
            if pResMin == pResCur {
                continue; }
            *ppTail = pResCur;
            ppTail = &pResCur.pData;
        }
        assert(*ppTail == nil, "*ppTail == nil");

        // update the phase of the node
        pRes = Abc_ObjNotCond( pResMin, (pRes.fPhase ^ pResMin.fPhase) );
    }
    Fraig_NodeSetData1( pNodeFraigR, (Fraig_Node_t *)pRes );
    // inlining of macro Fraig_IsComplement
    return Abc_ObjNotCond( pRes, ((int32(((ABC_PTRUINT_T) (pNodeFraig) & 01)))) );
}

/**Function*************************************************************

  Synopsis    [Transforms FRAIG into strashed network without choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromFraig2(pMan *Fraig_Man_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pProgress *ProgressBar;
    var tTable *stmm_table;
    var vNodeReprs *Vec_Ptr_t;
    var pNtkNew *Abc_Ntk_t;
    var pNode *Abc_Obj_t;
    var pRepr *Abc_Obj_t;
    var ppSlot **Abc_Obj_t;
    var i int32;

    // map the nodes into their lowest level representives
    tTable = stmm_init_table(stmm_ptrcmp,stmm_ptrhash);
    pNode = Abc_AigConst1(pNtk);
    // inlining of macro Fraig_Regular
    if !stmm_find_or_add( tTable, (char *)((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.pCopy) & ~01)) , (char ***)&ppSlot ) {
        *ppSlot = pNode; }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        // inlining of macro Fraig_Regular
        if !stmm_find_or_add( tTable, (char *)((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.pCopy) & ~01)) , (char ***)&ppSlot ) {
            *ppSlot = pNode; }
    }
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else
        if pNode.pCopy {
            // inlining of macro Fraig_Regular
            if !stmm_find_or_add( tTable, (char *)((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.pCopy) & ~01)) , (char ***)&ppSlot ) {
                *ppSlot = pNode;
            } else if (*ppSlot).Level > pNode.Level {
                *ppSlot = pNode; }
        }
    }
    // save representatives for each node
    vNodeReprs = Vec_PtrStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else
        if pNode.pCopy {           
            // inlining of macro Fraig_Regular
            if !stmm_lookup( tTable, (char *)((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.pCopy) & ~01)) , (char **)&pRepr ) {
                assert(false, "0"); }
            if pNode != pRepr {
                Vec_PtrWriteEntry( vNodeReprs, pNode.Id, pRepr ); }
        }
    }
    stmm_free_table( tTable );

    // create the new network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );

    // perform strashing
    Abc_AigSetNodePhases( pNtk );
    Abc_NtkIncrementTravId( pNtk );
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkCoNum(pNtk) );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        Abc_NtkFromFraig2_rec( pNtkNew, Abc_ObjFanin0(pNode), vNodeReprs );
    }
    Extra_ProgressBarStop( pProgress );
    Vec_PtrFree( vNodeReprs );

    // finalize the network
    Abc_NtkFinalize( pNtk, pNtkNew );
    return pNtkNew;
}


/**Function*************************************************************

  Synopsis    [Transforms into AIG one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromFraig2_rec(pNtkNew *Abc_Ntk_t, pNode *Abc_Obj_t, vNodeReprs *Vec_Ptr_t) {
    var pRepr *Abc_Obj_t;
    // skip the PIs and constants
    if Abc_ObjFaninNum(pNode) < 2 {
        return; }
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    assert(Abc_ObjIsNode( pNode ), "Abc_ObjIsNode( pNode )");
    // get the node's representative
    if pRepr = (Abc_Obj_t *)Vec_PtrEntry(vNodeReprs, pNode.Id) {
        Abc_NtkFromFraig2_rec( pNtkNew, pRepr, vNodeReprs );
        pNode.pCopy = Abc_ObjNotCond( pRepr.pCopy, pRepr.fPhase ^ pNode.fPhase );
        return;
    }
    Abc_NtkFromFraig2_rec( pNtkNew, Abc_ObjFanin0(pNode), vNodeReprs );
    Abc_NtkFromFraig2_rec( pNtkNew, Abc_ObjFanin1(pNode), vNodeReprs );
    pNode.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode) );
}



/**Function*************************************************************

  Synopsis    [Interfaces the network with the FRAIG package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigTrust(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;

    if !Abc_NtkIsSopLogic(pNtk) {
        fmt.Printf("Abc_NtkFraigTrust: Trust mode works for netlists and logic SOP networks.\n");
        return nil;
    }

    if !(Abc_NtkFraigTrustCheck(pNtk) != 0) {
        fmt.Printf("Abc_NtkFraigTrust: The network does not look like an AIG with choice nodes.\n");
        return nil;
    }
    
    // perform strashing
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );
    Abc_NtkFraigTrustOne( pNtk, pNtkNew );
    Abc_NtkFinalize( pNtk, pNtkNew );
    Abc_NtkReassignIds( pNtkNew );

    // print a warning about choice nodes
    fmt.Printf("Warning: The resulting AIG contains %d choice nodes.\n", Abc_NtkGetChoiceNum( pNtkNew ));

    // make sure that everything is okay
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkFraigTrust: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Checks whether the node can be processed in the trust mode.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigTrustCheck(pNtk *Abc_Ntk_t) int32 {
    var pNode *Abc_Obj_t;
    var i, nFanins int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        nFanins = Abc_ObjFaninNum(pNode);
        if nFanins < 2 {
            continue; }
        if nFanins == 2 && Abc_SopIsAndType((char *)pNode.pData) {
            continue; }
        if !Abc_SopIsOrType((char *)pNode.pData) {
            return 0; }
    }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Interfaces the network with the FRAIG package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigTrustOne(pNtk *Abc_Ntk_t, pNtkNew *Abc_Ntk_t) {
    var pProgress *ProgressBar;
    var vNodes *Vec_Ptr_t;
    var pNode, pNodeNew, pObj *Abc_Obj_t;
    var i int32;

    // perform strashing
    vNodes = Abc_NtkDfs( pNtk, 0 );
    pProgress = Extra_ProgressBarStart( os.Stdout, vNodes.nSize );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        // get the node
        assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
         // strash the node
        pNodeNew = Abc_NodeFraigTrust( pNtkNew, pNode );
        // get the old object
        if Abc_NtkIsNetlist(pNtk) {
            pObj = Abc_ObjFanout0( pNode ); // the fanout net 
        } else { 
            pObj = pNode; } // the node itself
        // make sure the node is not yet strashed
        assert(pObj.pCopy == nil, "pObj.pCopy == nil");
        // mark the old object with the new AIG node
        pObj.pCopy = pNodeNew;
    }
    Vec_PtrFree( vNodes );
    Extra_ProgressBarStop( pProgress );
}

/**Function*************************************************************

  Synopsis    [Transforms one node into a FRAIG in the trust mode.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeFraigTrust(pNtkNew *Abc_Ntk_t, pNode *Abc_Obj_t) *Abc_Obj_t {
    var pSum, pFanin *Abc_Obj_t;
    var ppTail *any;
    var i, nFanins, fCompl int32;

    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // get the number of node's fanins
    nFanins = Abc_ObjFaninNum( pNode );
    assert(nFanins == Abc_SopGetVarNum((char *)pNode.pData), "nFanins == Abc_SopGetVarNum((char *)pNode.pData)");
    // check if it is a constant
    if nFanins == 0 {
        return Abc_ObjNotCond( Abc_AigConst1(pNtkNew), Abc_SopIsConst0((char *)pNode.pData) ); }
    if nFanins == 1 {
        return Abc_ObjNotCond( Abc_ObjFanin0(pNode).pCopy, Abc_SopIsInv((char *)pNode.pData) ); }
    if nFanins == 2 && Abc_SopIsAndType((char *)pNode.pData) {
        return Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, 
            Abc_ObjNotCond( Abc_ObjFanin0(pNode).pCopy, !Abc_SopGetIthCareLit((char *)pNode.pData,0) ),
            Abc_ObjNotCond( Abc_ObjFanin1(pNode).pCopy, !Abc_SopGetIthCareLit((char *)pNode.pData,1) )  ); }
    assert(Abc_SopIsOrType((char *)pNode.pData), "Abc_SopIsOrType((char *)pNode.pData)");
    fCompl = Abc_SopGetIthCareLit((char *)pNode.pData,0);
    // get the root of the choice node (the first fanin)
    pSum = Abc_ObjFanin0(pNode).pCopy;
    // connect other fanins
    ppTail = &pSum.pData;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if i == 0 {
            continue; }
        *ppTail = pFanin.pCopy;
        ppTail = &pFanin.pCopy.pData;
        // set the complemented bit of this cut
        if fCompl ^ Abc_SopGetIthCareLit((char *)pNode.pData, i) {
            pFanin.pCopy.fPhase = 1; }
    }
    assert(*ppTail == nil, "*ppTail == nil");
    return pSum;
}




/**Function*************************************************************

  Synopsis    [Interfaces the network with the FRAIG package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigStore(pNtkAdd *Abc_Ntk_t) int32 {
    var vStore *Vec_Ptr_t;
    var pNtk *Abc_Ntk_t;
    // create the network to be stored
    pNtk = Abc_NtkStrash( pNtkAdd, 0, 0, 0 );
    if pNtk == nil {
        fmt.Printf("Abc_NtkFraigStore: Initial strashing has failed.\n");
        return 0;
    }
    // get the network currently stored
    vStore = Abc_FrameReadStore();
    if Vec_PtrSize(vStore) > 0 {
        // check that the networks have the same PIs
        extern int Abc_NodeCompareCiCo( Abc_Ntk_t * pNtkOld, Abc_Ntk_t * pNtkNew );
        if !Abc_NodeCompareCiCo(pNtk, (Abc_Ntk_t *)Vec_PtrEntry(vStore, 0)) {
            // reorder PIs of pNtk2 according to pNtk1
            if !Abc_NtkCompareSignals( pNtk, (Abc_Ntk_t *)Vec_PtrEntry(vStore, 0), 1, 1 ) {
                fmt.Printf("Trying to store the network with different primary inputs.\n");
                fmt.Printf("The previously stored networks are deleted and this one is added.\n");
                Abc_NtkFraigStoreClean();
            }
        }
    }
    Vec_PtrPush( vStore, pNtk );
//    printf( "The number of AIG nodes added to storage = %5d.\n", Abc_NtkNodeNum(pNtk) );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Interfaces the network with the FRAIG package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigRestore(nPatsRand int32, nPatsDyna int32, nBTLimit int32) *Abc_Ntk_t {
    extern Abc_Ntk_t * Abc_NtkFraigPartitioned( Vec_Ptr_t * vStore, void * pParams );
    var Params Fraig_Params_t;
    var vStore *Vec_Ptr_t;
    var pNtk, pFraig *Abc_Ntk_t;
    var nWords1, nWords2, nWordsMin int32;
//    abctime clk = Abc_Clock();

    // get the stored network
    vStore = Abc_FrameReadStore();
    if Vec_PtrSize(vStore) == 0 {
        fmt.Printf("There are no network currently in storage.\n");
        return nil;
    }
//    printf( "Currently stored %d networks will be fraiged.\n", Vec_PtrSize(vStore) );
    pNtk = (Abc_Ntk_t *)Vec_PtrEntry( vStore, 0 );

    // swap the first and last network
    // this should lead to the primary choice being "better" because of synthesis
    if Vec_PtrSize(vStore) > 1 {
        pNtk = (Abc_Ntk_t *)Vec_PtrPop( vStore );
        Vec_PtrPush( vStore, Vec_PtrEntry(vStore,0) );
        Vec_PtrWriteEntry( vStore, 0, pNtk );
    }

    // to determine the number of simulation patterns
    // use the following strategy
    // at least 64 words (32 words random and 32 words dynamic)
    // no more than 256M for one circuit (128M + 128M)
    nWords1 = 32;
    nWords2 = (1<<27) / (Abc_NtkNodeNum(pNtk) + Abc_NtkCiNum(pNtk));
    nWordsMin = Abc_MinInt( nWords1, nWords2 );

    // set parameters for fraiging
    Fraig_ParamsSetDefault( &Params );
    if nPatsRand != 0 {
        Params.nPatsRand = nPatsRand;
    } else {
        Params.nPatsRand = nWordsMin * 32;
    }    // the number of words of random simulation info
    if nPatsDyna != 0 {
        Params.nPatsDyna = nPatsDyna;
    } else {
        Params.nPatsDyna = nWordsMin * 32;
    }    // the number of words of dynamic simulation info
    Params.nBTLimit   = nBTLimit;          // the max number of backtracks to perform
    Params.fFuncRed   =    1;              // performs only one level hashing
    Params.fFeedBack  =    1;              // enables solver feedback
    Params.fDist1Pats =    1;              // enables distance-1 patterns
    Params.fDoSparse  =    1;              // performs equiv tests for sparse functions 
    Params.fChoicing  =    1;              // enables recording structural choices
    Params.fTryProve  =    0;              // tries to solve the final miter
    Params.fInternal  =    1;              // does not show progress bar
    Params.fVerbose   =    0;              // the verbosiness flag

    // perform partitioned computation of structural choices
    pFraig = Abc_NtkFraigPartitioned( vStore, &Params );
    Abc_NtkFraigStoreClean();
//ABC_PRT( "Total choicing time", Abc_Clock() - clk );
    return pFraig;
}

/**Function*************************************************************

  Synopsis    [Interfaces the network with the FRAIG package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigStoreClean() {
    var vStore *Vec_Ptr_t;
    var pNtk *Abc_Ntk_t;
    var i int32;
    vStore = Abc_FrameReadStore();
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vStore)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(vStore, i)), 1); i++  {
        Abc_NtkDelete( pNtk );
    }
    Vec_PtrClear( vStore );
}

/**Function*************************************************************

  Synopsis    [Checks the correctness of stored networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigStoreCheck(pFraig *Abc_Ntk_t) {
    var pNode0, pNode1 *Abc_Obj_t;
    var nPoOrig, nPoFinal, nStored int32; 
    var i, k int32;
    // check that the PO functions are correct
    nPoFinal = Abc_NtkPoNum(pFraig);
    nStored  = Abc_FrameReadStoreSize();
    assert(nPoFinal % nStored == 0, "nPoFinal % nStored == 0");
    nPoOrig  = nPoFinal / nStored;
    for  i = 0; i < nPoOrig; i++  {
        pNode0 = Abc_ObjFanin0( Abc_NtkPo(pFraig, i) ); 
        for  k = 1; k < nStored; k++  {
            pNode1 = Abc_ObjFanin0( Abc_NtkPo(pFraig, k*nPoOrig+i) ); 
            if pNode0 != pNode1 {
                fmt.Printf("Verification for PO #%d of network #%d has failed. The PO function is not used.\n", i+1, k+1); }
        }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

