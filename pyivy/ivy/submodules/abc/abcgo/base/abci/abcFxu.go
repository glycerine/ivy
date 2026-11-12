package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcFxu.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface with the fast extract package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcFxu.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "opt/fxu/fxu.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
// static int  Abc_NtkFxuCheck( Abc_Ntk_t * pNtk );

// static void Abc_NtkFxuCollectInfo( Abc_Ntk_t * pNtk, Fxu_Data_t * p );

// static void Abc_NtkFxuReconstruct( Abc_Ntk_t * pNtk, Fxu_Data_t * p );


// extern int  Fxu_FastExtract( Fxu_Data_t * pData );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets default values of the FXU parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSetDefaultFxParams(p *Fxu_Data_t) {
    memset( p, 0, sizeof(Fxu_Data_t) );
    p.nSingleMax =  20000;
    p.nPairsMax  =  30000;
    p.nNodesExt  =1000000;
    p.WeightMin  =      0;
    p.LitCountMax=      4;
    p.fOnlyS     =      0;
    p.fOnlyD     =      0;
    p.fUse0      =      0;
    p.fUseCompl  =      1;
    p.fVerbose   =      0;
}

/**Function*************************************************************

  Synopsis    [Performs fast_extract on the current network.]

  Description [Takes the network and the maximum number of nodes to extract.
  Uses the concurrent double-cube and single cube divisor extraction procedure.
  Modifies the network in the end, after extracting all nodes. Note that 
  Ntk_NetworkSweep() may increase the performance of this procedure because 
  the single-literal nodes will not be created in the sparse matrix. Returns 1 
  if the network has been changed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFastExtract(pNtk *Abc_Ntk_t, p *Fxu_Data_t) int32 {
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    // if the network is already in the SOP form, it may come from BLIF file
    // and it may not be SCC-free, in which case FXU will not work correctly
    if Abc_NtkIsSopLogic(pNtk) { // to make sure the SOPs are SCC-free
//        Abc_NtkSopToBdd(pNtk);
//        Abc_NtkBddToSop(pNtk);
    }
    // get the network in the SOP form
    if !Abc_NtkToSop(pNtk, -1, ABC_INFINITY) {
        fmt.Printf("Abc_NtkFastExtract(): Converting to SOPs has failed.\n");
        return 0;
    }
    // check if the network meets the requirements
    if !(Abc_NtkFxuCheck(pNtk) != 0) {
        fmt.Printf("Abc_NtkFastExtract: Nodes have duplicated or complemented fanins. FXU is not performed.\n");
        return 0;
    }
    // sweep removes useless nodes
    Abc_NtkCleanup( pNtk, 0 );
    // collect information about the covers
    Abc_NtkFxuCollectInfo( pNtk, p );
    // call the fast extract procedure
    if Fxu_FastExtract(p) > 0 {
        // update the network
        Abc_NtkFxuReconstruct( pNtk, p );
        // make sure everything is okay
        if !Abc_NtkCheck( pNtk ) {
            fmt.Printf("Abc_NtkFastExtract: The network check has failed.\n"); }
        return 1;
    
    } else {
        fmt.Printf("Warning: The network has not been changed by \"fx\".\n"); }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Makes sure the nodes do not have complemented and duplicated fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFxuCheck(pNtk *Abc_Ntk_t) int32 {
    var pNode, pFanin1, pFanin2 *Abc_Obj_t;
    var n, i, k int32;
    // inlining of macro Abc_NtkForEachNode
    for  n = 0; (n < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, n)), 1); n++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin1) = Abc_ObjFanin(pNode, i)), 1); i++  {
            if i < 2 && Abc_ObjFaninC(pNode, i) {
                return 0; }
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin2) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if i == k {
                    continue; }
                if pFanin1 == pFanin2 {
                    return 0; }
            }
        }
    }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Collect information about the network for fast_extract.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFxuCollectInfo(pNtk *Abc_Ntk_t, p *Fxu_Data_t) {
    var pNode *Abc_Obj_t;
    var i int32;
    // add information to the manager
    p.pManSop    = (Mem_Flex_t *)pNtk.pManFunc;
    p.vSops      = Vec_PtrAlloc(0);
    p.vFanins    = Vec_PtrAlloc(0);
    p.vSopsNew   = Vec_PtrAlloc(0);
    p.vFaninsNew = Vec_PtrAlloc(0);
    Vec_PtrFill( p.vSops,      Abc_NtkObjNumMax(pNtk), nil );
    Vec_PtrFill( p.vFanins,    Abc_NtkObjNumMax(pNtk), nil );
    Vec_PtrFill( p.vSopsNew,   Abc_NtkObjNumMax(pNtk) + p.nNodesExt, nil );
    Vec_PtrFill( p.vFaninsNew, Abc_NtkObjNumMax(pNtk) + p.nNodesExt, nil );
    // add SOPs and fanin array
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if Abc_SopGetVarNum((char *)pNode.pData) < 2 {
            continue; }
        if Abc_SopGetCubeNum((char *)pNode.pData) < 1 {
            continue; }
        p.vSops.pArray[i]   = pNode.pData;
        p.vFanins.pArray[i] = &pNode.vFanins;
    }
    }
    p.nNodesOld = Abc_NtkObjNumMax(pNtk);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFxuFreeInfo(p *Fxu_Data_t) {
    var i int32;
    // free the arrays of new fanins
    if p.vFaninsNew {
        for  i = 0; i < p.vFaninsNew.nSize; i++  {
            if p.vFaninsNew.pArray[i] {
                Vec_IntFree( (Vec_Int_t *)p.vFaninsNew.pArray[i] ); }
        } }
    // free the arrays
    if p.vSops { Vec_PtrFree( p.vSops      ); }
    if p.vSopsNew { Vec_PtrFree( p.vSopsNew   ); }
    if p.vFanins { Vec_PtrFree( p.vFanins    ); }
    if p.vFaninsNew { Vec_PtrFree( p.vFaninsNew ); }
//    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Reconstructs the network after FX.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFxuReconstruct(pNtk *Abc_Ntk_t, p *Fxu_Data_t) {
    var vFanins *Vec_Int_t;
    var pNode, pFanin *Abc_Obj_t;
    var i, k int32;

    assert(p.vFanins.nSize < p.vFaninsNew.nSize, "p.vFanins.nSize < p.vFaninsNew.nSize");
    // create the new nodes
    for  i = p.vFanins.nSize; i < p.vFanins.nSize + p.nNodesNew; i++  {
        // start the node
        pNode = Abc_NtkCreateNode( pNtk );
        assert(i == int32(pNode.Id), "i == int32(pNode.Id)");
    }
    // update the old nodes
    for  i = 0; i < p.vFanins.nSize; i++  {
        // the new array of fanins
        vFanins = (Vec_Int_t *)p.vFaninsNew.pArray[i];
        if vFanins == nil {
            continue; }
        // remove old fanins
        pNode = Abc_NtkObj( pNtk, i );
        Abc_ObjRemoveFanins( pNode );
        // add new fanins
        vFanins = (Vec_Int_t *)p.vFaninsNew.pArray[i];
        for  k = 0; k < vFanins.nSize; k++  {
            pFanin = Abc_NtkObj( pNtk, vFanins.pArray[k] );
            Abc_ObjAddFanin( pNode, pFanin );
        }
        pNode.pData = p.vSopsNew.pArray[i];
        assert(pNode.pData != nil, "pNode.pData != nil");
    }
    // set up the new nodes
    for  i = p.vFanins.nSize; i < p.vFanins.nSize + p.nNodesNew; i++  {
        // get the new node
        pNode = Abc_NtkObj( pNtk, i );
        // add the fanins
        vFanins = (Vec_Int_t *)p.vFaninsNew.pArray[i];
        for  k = 0; k < vFanins.nSize; k++  {
            pFanin = Abc_NtkObj( pNtk, vFanins.pArray[k] );
            Abc_ObjAddFanin( pNode, pFanin );
        }
        pNode.pData = p.vSopsNew.pArray[i];
        assert(pNode.pData != nil, "pNode.pData != nil");
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

