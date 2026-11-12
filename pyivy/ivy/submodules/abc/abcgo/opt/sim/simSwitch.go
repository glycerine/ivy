package sim

/**CFile****************************************************************

  FileName    [simSwitch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Computes switching activity of nodes in the ABC network.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: simSwitch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "sim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Sim_NodeSimulate( Abc_Obj_t * pNode, Vec_Ptr_t * vSimInfo, int nSimWords );

// static float Sim_ComputeSwitching( unsigned * pSimInfo, int nSimWords );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes switching activity using simulation.]

  Description [Computes switching activity, which is understood as the
  probability of switching under random simulation. Assigns the
  random simulation information at the CI and propagates it through
  the internal nodes of the AIG.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_NtkComputeSwitching(pNtk *Abc_Ntk_t, nPatterns int32) *Vec_Int_t {
    var vSwitching *Vec_Int_t;
    var pSwitching *float32;
    var vNodes *Vec_Ptr_t;
    var vSimInfo *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var pSimInfo *uint32;
    var nSimWords, i int32;

    // allocate space for simulation info of all nodes
    // inlining of macro SIM_NUM_WORDS
    nSimWords = (((nPatterns)>>5) + (((nPatterns)&31) > 0));
    vSimInfo = Sim_UtilInfoAlloc( Abc_NtkObjNumMax(pNtk), nSimWords, 0 );
    // assign the random simulation to the CIs
    vSwitching = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    pSwitching = (float *)vSwitching.pArray;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pSimInfo = (unsigned *)Vec_PtrEntry(vSimInfo, pNode.Id);
        Sim_UtilSetRandom( pSimInfo, nSimWords );
        pSwitching[pNode.Id] = Sim_ComputeSwitching( pSimInfo, nSimWords );
    }
    // simulate the internal nodes
    vNodes  = Abc_AigDfs( pNtk, 1, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pSimInfo = (unsigned *)Vec_PtrEntry(vSimInfo, pNode.Id);
        Sim_UtilSimulateNodeOne( pNode, vSimInfo, nSimWords, 0 );
        pSwitching[pNode.Id] = Sim_ComputeSwitching( pSimInfo, nSimWords );
    }
    Vec_PtrFree( vNodes );
    Sim_UtilInfoFree( vSimInfo );
    return vSwitching;
}

/**Function*************************************************************

  Synopsis    [Computes switching activity of one node.]

  Description [Uses the formula: Switching = 2 * nOnes * nZeros / (nTotal ^ 2) ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ComputeSwitching(pSimInfo *uint32, nSimWords int32) float32 {
    var nOnes, nTotal int32;
    nTotal = 32 * nSimWords;
    nOnes = Sim_UtilCountOnes( pSimInfo, nSimWords );
    return float32(2.0) * nOnes / nTotal * (nTotal - nOnes) / nTotal;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

