package sim

/**CFile****************************************************************

  FileName    [simSeq.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Simulation for sequential circuits.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: simUtils.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "sim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Sim_SimulateSeqFrame( Vec_Ptr_t * vInfo, Abc_Ntk_t * pNtk, int iFrames, int nWords, int fTransfer );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Simulates sequential circuit.]

  Description [Takes sequential circuit (pNtk). Simulates the given number
  (nFrames) of the circuit with the given number of machine words (nWords)
  of random simulation data, starting from the initial state. If the initial
  state of some latches is a don't-care, uses random input for that latch.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SimulateSeqRandom(pNtk *Abc_Ntk_t, nFrames int32, nWords int32) *Vec_Ptr_t {
    var vInfo *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    vInfo = Sim_UtilInfoAlloc( Abc_NtkObjNumMax(pNtk), nWords * nFrames, 0 );
    // set the constant data
    pNode = Abc_AigConst1(pNtk);
    // inlining of macro Sim_SimInfoGet
    Sim_UtilSetConst( ((unsigned *)((vInfo).pArray[(pNode).Id])), nWords * nFrames, 1 );
    // set the random PI data
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pNode) = Abc_NtkPi(pNtk, i)), 1); i++  {
        // inlining of macro Sim_SimInfoGet
        Sim_UtilSetRandom( ((unsigned *)((vInfo).pArray[(pNode).Id])), nWords * nFrames );
    }
    // set the initial state data
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else
        if Abc_LatchIsInit0(pNode) {
            // inlining of macro Sim_SimInfoGet
            Sim_UtilSetConst( ((unsigned *)((vInfo).pArray[(pNode).Id])), nWords, 0 );
        } else if Abc_LatchIsInit1(pNode) {
            // inlining of macro Sim_SimInfoGet
            Sim_UtilSetConst( ((unsigned *)((vInfo).pArray[(pNode).Id])), nWords, 1 );
        } else { 
            // inlining of macro Sim_SimInfoGet
            Sim_UtilSetRandom( ((unsigned *)((vInfo).pArray[(pNode).Id])), nWords ); }
    }
    // simulate the nodes for the given number of timeframes
    for  i = 0; i < nFrames; i++  {
        Sim_SimulateSeqFrame( vInfo, pNtk, i, nWords, int32((i < nFrames-1)) );
    }
    return vInfo;
}

/**Function*************************************************************

  Synopsis    [Simulates sequential circuit.]

  Description [Takes sequential circuit (pNtk). Simulates the given number
  (nFrames) of the circuit with the given model. The model is assumed to 
  contain values of PIs for each frame. The latches are initialized to
  the initial state. One word of data is simulated.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SimulateSeqModel(pNtk *Abc_Ntk_t, nFrames int32, pModel *int32) *Vec_Ptr_t {
    var vInfo *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var pUnsigned *uint32;
    var i, k int32;
    vInfo = Sim_UtilInfoAlloc( Abc_NtkObjNumMax(pNtk), nFrames, 0 );
    // set the constant data
    pNode = Abc_AigConst1(pNtk);
    // inlining of macro Sim_SimInfoGet
    Sim_UtilSetConst( ((unsigned *)((vInfo).pArray[(pNode).Id])), nFrames, 1 );
    // set the random PI data
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pNode) = Abc_NtkPi(pNtk, i)), 1); i++  {
        // inlining of macro Sim_SimInfoGet
        pUnsigned = ((unsigned *)((vInfo).pArray[(pNode).Id]));
        for  k = 0; k < nFrames; k++  {
            pUnsigned[k] = pModel[k * Abc_NtkPiNum(pNtk) + i] ? ~(uint32(0)) : 0;
        }
    }
    // set the initial state data
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else {
        // inlining of macro Sim_SimInfoGet
        pUnsigned = ((unsigned *)((vInfo).pArray[(pNode).Id]));
        if Abc_LatchIsInit0(pNode) {
            pUnsigned[0] = 0;
        } else if Abc_LatchIsInit1(pNode) {
            pUnsigned[0] = ~(uint32(0));
        } else { 
            pUnsigned[0] = SIM_RANDOM_UNSIGNED; }
    }
    }
    // simulate the nodes for the given number of timeframes
    for  i = 0; i < nFrames; i++  {
        Sim_SimulateSeqFrame( vInfo, pNtk, i, 1, int32((i < nFrames-1)) );
    }
/*
    // print the simulated values
    for ( i = 0; i < nFrames; i++ )
    {
        printf( "Frame %d : ", i+1 );
        Abc_NtkForEachPi( pNtk, pNode, k )
            printf( "%d", Sim_SimInfoGet(vInfo,pNode)[i] > 0 );
        printf( " " );
        Abc_NtkForEachLatch( pNtk, pNode, k )
            printf( "%d", Sim_SimInfoGet(vInfo,pNode)[i] > 0 );
        printf( " " );
        Abc_NtkForEachPo( pNtk, pNode, k )
            printf( "%d", Sim_SimInfoGet(vInfo,pNode)[i] > 0 );
        printf( "\n" );
    }
    printf( "\n" );
*/
    return vInfo;
}

/**Function*************************************************************

  Synopsis    [Simulates one frame of sequential circuit.]

  Description [Assumes that the latches and POs are already initialized.
  In the end transfers the data to the latches of the next frame.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SimulateSeqFrame(vInfo *Vec_Ptr_t, pNtk *Abc_Ntk_t, iFrames int32, nWords int32, fTransfer int32) {
    var pNode *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Sim_UtilSimulateNodeOne( pNode, vInfo, nWords, iFrames * nWords ); }
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pNode) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Sim_UtilTransferNodeOne( pNode, vInfo, nWords, iFrames * nWords, 0 );
    }
    if !(fTransfer != 0) {
        return; }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else {
        Sim_UtilTransferNodeOne( pNode, vInfo, nWords, iFrames * nWords, 1 ); }
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

