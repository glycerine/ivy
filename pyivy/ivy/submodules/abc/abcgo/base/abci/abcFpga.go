package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcFpga.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface with the FPGA mapping package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcFpga.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "map/fpga/fpgaInt.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Fpga_Man_t * Abc_NtkToFpga( Abc_Ntk_t * pNtk, int fRecovery, float * pSwitching, int fLatchPaths, int fVerbose );

// static Abc_Ntk_t *  Abc_NtkFromFpga( Fpga_Man_t * pMan, Abc_Ntk_t * pNtk );

// static Abc_Obj_t *  Abc_NodeFromFpga_rec( Abc_Ntk_t * pNtkNew, Fpga_Node_t * pNodeFpga );

 
////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Interface with the FPGA mapping package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFpga(pNtk *Abc_Ntk_t, DelayTarget float32, fRecovery int32, fSwitching int32, fLatchPaths int32, fVerbose int32) *Abc_Ntk_t {
    var fShowSwitching int32  = 1;
    var pNtkNew *Abc_Ntk_t;
    var pMan *Fpga_Man_t;
    var vSwitching *Vec_Int_t  = nil;
    var pSwitching *float32  = nil;
    var Num int32;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    // print a warning about choice nodes
    if Num = Abc_NtkGetChoiceNum( pNtk ) {
        Abc_Print( 0, "Performing LUT mapping with %d choices.\n", Num ); }

    // compute switching activity
    fShowSwitching |= fSwitching;
    if fShowSwitching != 0 {
        extern Vec_Int_t * Sim_NtkComputeSwitching( Abc_Ntk_t * pNtk, int nPatterns );
        vSwitching = Sim_NtkComputeSwitching( pNtk, 4096 );
        pSwitching = (float *)vSwitching.pArray;
    }

    // perform FPGA mapping
    pMan = Abc_NtkToFpga( pNtk, fRecovery, pSwitching, fLatchPaths, fVerbose );    
    if pSwitching != nil { assert((vSwitching != nil), "vSwitching"); Vec_IntFree( vSwitching ); }
    if pMan == nil {
        return nil; }
    Fpga_ManSetSwitching( pMan, fSwitching );
    Fpga_ManSetLatchPaths( pMan, fLatchPaths );
    Fpga_ManSetLatchNum( pMan, Abc_NtkLatchNum(pNtk) );
    Fpga_ManSetDelayTarget( pMan, DelayTarget );
    if !Fpga_Mapping( pMan ) {
        Fpga_ManFree( pMan );
        return nil;
    }

    // transform the result of mapping into a BDD network
    pNtkNew = Abc_NtkFromFpga( pMan, pNtk );
    if pNtkNew == nil {
        return nil; }
    Fpga_ManFree( pMan );

    // make the network minimum base
    Abc_NtkMinimumBase( pNtkNew );

    if pNtk.pExdc {
        pNtkNew.pExdc = Abc_NtkDup( pNtk.pExdc ); }

    // make sure that everything is okay
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkFpga: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Load the network into FPGA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToFpga(pNtk *Abc_Ntk_t, fRecovery int32, pSwitching *float32, fLatchPaths int32, fVerbose int32) *Fpga_Man_t {
    var pMan *Fpga_Man_t;
    var pProgress *ProgressBar;
    var pNodeFpga *Fpga_Node_t;
    var vNodes *Vec_Ptr_t;
    var pNode, pFanin, pPrev *Abc_Obj_t;
    var pfArrivals *float32;
    var i int32;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    // start the mapping manager and set its parameters
    pMan = Fpga_ManCreate( Abc_NtkCiNum(pNtk), Abc_NtkCoNum(pNtk), fVerbose );
    if pMan == nil {
        return nil; }
    Fpga_ManSetAreaRecovery( pMan, fRecovery );
    Fpga_ManSetOutputNames( pMan, Abc_NtkCollectCioNames(pNtk, 1) );
    pfArrivals = Abc_NtkGetCiArrivalFloats(pNtk);
    if fLatchPaths != 0 {
        for  i = 0; i < Abc_NtkPiNum(pNtk); i++  {
            pfArrivals[i] = -FPGA_FLOAT_LARGE;
        }
    }
    Fpga_ManSetInputArrivals( pMan, pfArrivals );

    // create PIs and remember them in the old nodes
    Abc_NtkCleanCopy( pNtk );
    Abc_AigConst1(pNtk).pCopy = (Abc_Obj_t *)Fpga_ManReadConst1(pMan);
    Abc_NtkForEachCi( pNtk, pNode, i )
    {
        pNodeFpga = Fpga_ManReadInputs(pMan)[i];
        pNode.pCopy = (Abc_Obj_t *)pNodeFpga;
        if pSwitching != nil {
            Fpga_NodeSetSwitching( pNodeFpga, pSwitching[pNode.Id] ); }
    }

    // load the AIG into the mapper
    vNodes = Abc_AigDfs( pNtk, 0, 0 );
    pProgress = Extra_ProgressBarStart( os.Stdout, vNodes.nSize );
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pNode, i )
    {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        // add the node to the mapper
        pNodeFpga = Fpga_NodeAnd( pMan, 
            Fpga_NotCond( Abc_ObjFanin0(pNode).pCopy, Abc_ObjFaninC0(pNode) ),
            Fpga_NotCond( Abc_ObjFanin1(pNode).pCopy, Abc_ObjFaninC1(pNode) ) );
        assert(pNode.pCopy == nil, "pNode.pCopy == nil");
        // remember the node
        pNode.pCopy = (Abc_Obj_t *)pNodeFpga;
        if pSwitching != nil {
            Fpga_NodeSetSwitching( pNodeFpga, pSwitching[pNode.Id] ); }
        // set up the choice node
        if Abc_AigNodeIsChoice( pNode ) {
            for  pPrev = pNode, pFanin = (Abc_Obj_t *)pNode.pData; pFanin; pPrev = pFanin, pFanin = (Abc_Obj_t *)pFanin.pData  {
                Fpga_NodeSetNextE( (Fpga_Node_t *)pPrev.pCopy, (Fpga_Node_t *)pFanin.pCopy );
                Fpga_NodeSetRepr( (Fpga_Node_t *)pFanin.pCopy, (Fpga_Node_t *)pNode.pCopy );
            } }
    }
    Extra_ProgressBarStop( pProgress );
    Vec_PtrFree( vNodes );

    // set the primary outputs without copying the phase
    Abc_NtkForEachCo( pNtk, pNode, i )
        Fpga_ManReadOutputs(pMan)[i] = (Fpga_Node_t *)Abc_ObjFanin0(pNode).pCopy;
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Creates the mapped network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromFpga(pMan *Fpga_Man_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pProgress *ProgressBar;
    var pNtkNew *Abc_Ntk_t;
    var pNode, pNodeNew *Abc_Obj_t;
    var i, nDupGates int32;
    // create the new network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_BDD );
    // make the mapper point to the new network
    Fpga_CutsCleanSign( pMan );
    Fpga_ManCleanData0( pMan );
    Abc_NtkForEachCi( pNtk, pNode, i )
        Fpga_NodeSetData0( Fpga_ManReadInputs(pMan)[i], (char *)pNode.pCopy );
    // set the constant node
//    if ( Fpga_NodeReadRefs(Fpga_ManReadConst1(pMan)) > 0 )
        Fpga_NodeSetData0( Fpga_ManReadConst1(pMan), (char *)Abc_NtkCreateNodeConst1(pNtkNew) );
    // process the nodes in topological order
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkCoNum(pNtk) );
    Abc_NtkForEachCo( pNtk, pNode, i )
    {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        pNodeNew = Abc_NodeFromFpga_rec( pNtkNew, Fpga_ManReadOutputs(pMan)[i] );
        assert(!Abc_ObjIsComplement(pNodeNew), "!Abc_ObjIsComplement(pNodeNew)");
        Abc_ObjFanin0(pNode).pCopy = pNodeNew;
    }
    Extra_ProgressBarStop( pProgress );
    // finalize the new network
    Abc_NtkFinalize( pNtk, pNtkNew );
    // remove the constant node if not used
    pNodeNew = (Abc_Obj_t *)Fpga_NodeReadData0(Fpga_ManReadConst1(pMan));
    if Abc_ObjFanoutNum(pNodeNew) == 0 {
        Abc_NtkDeleteObj( pNodeNew ); }
    // decouple the PO driver nodes to reduce the number of levels
    nDupGates = Abc_NtkLogicMakeSimpleCos( pNtkNew, 1 );
    if (nDupGates != 0) && Fpga_ManReadVerbose(pMan) {
        fmt.Printf("Duplicated %d gates to decouple the CO drivers.\n", nDupGates); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Derive one node after FPGA mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeFromFpga_rec(pNtkNew *Abc_Ntk_t, pNodeFpga *Fpga_Node_t) *Abc_Obj_t {
    var pCutBest *Fpga_Cut_t;
    var ppLeaves **Fpga_Node_t; 
    var pNodeNew *Abc_Obj_t;
    var i, nLeaves int32;
    assert(!Fpga_IsComplement(pNodeFpga), "!Fpga_IsComplement(pNodeFpga)");
    // return if the result if known
    pNodeNew = (Abc_Obj_t *)Fpga_NodeReadData0( pNodeFpga );
    if pNodeNew != nil {
        return pNodeNew; }
    assert(Fpga_NodeIsAnd(pNodeFpga), "Fpga_NodeIsAnd(pNodeFpga)");
    // get the parameters of the best cut
    pCutBest = Fpga_NodeReadCutBest( pNodeFpga );
    ppLeaves = Fpga_CutReadLeaves( pCutBest );
    nLeaves  = Fpga_CutReadLeavesNum( pCutBest ); 
    // create a new node 
    pNodeNew = Abc_NtkCreateNode( pNtkNew );
    for  i = 0; i < nLeaves; i++  {
        Abc_ObjAddFanin( pNodeNew, Abc_NodeFromFpga_rec(pNtkNew, ppLeaves[i]) );
    }
    // derive the function of this node
    pNodeNew.pData = Fpga_TruthsCutBdd( pNtkNew.pManFunc, pCutBest );   Cudd_Ref( (DdNode *)pNodeNew.pData );
    Fpga_NodeSetData0( pNodeFpga, (char *)pNodeNew );
    return pNodeNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

