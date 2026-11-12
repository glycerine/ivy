package fpga

/**CFile****************************************************************

  FileName    [fpgaSwitch.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: fpgaSwitch.h,v 1.0 2003/09/08 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fpgaInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**function*************************************************************

  synopsis    [Computes the exact area associated with the cut.]

  description []
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutGetSwitchDerefed(pMan *Fpga_Man_t, pNode *Fpga_Node_t, pCut *Fpga_Cut_t) float32 {
    var aResult, aResult2 float32;
    aResult2 = Fpga_CutRefSwitch( pMan, pNode, pCut, 0 );
    aResult  = Fpga_CutDerefSwitch( pMan, pNode, pCut, 0 );
//    assert( aResult == aResult2 );
    return aResult;
}

/**function*************************************************************

  synopsis    [References the cut.]

  description [This procedure is similar to the procedure NodeReclaim.]
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutRefSwitch(pMan *Fpga_Man_t, pNode *Fpga_Node_t, pCut *Fpga_Cut_t, fFanouts int32) float32 {
    var pNodeChild *Fpga_Node_t;
    var aArea float32;
    var i int32;
    // start the area of this cut
    aArea = pNode.Switching;
    if pCut.nLeaves == 1 {
        return aArea; }
    // go through the children
    for  i = 0; i < pCut.nLeaves; i++  {
        pNodeChild = pCut.ppLeaves[i];
        assert(pNodeChild.nRefs >= 0, "pNodeChild.nRefs >= 0");
        if pNodeChild.nRefs++ > 0 {  
            continue; }
        aArea += Fpga_CutRefSwitch( pMan, pNodeChild, pNodeChild.pCutBest, fFanouts );
    }
    return aArea;
}

/**function*************************************************************

  synopsis    [Dereferences the cut.]

  description [This procedure is similar to the procedure NodeRecusiveDeref.]
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutDerefSwitch(pMan *Fpga_Man_t, pNode *Fpga_Node_t, pCut *Fpga_Cut_t, fFanouts int32) float32 {
    var pNodeChild *Fpga_Node_t;
    var aArea float32;
    var i int32;
    // start the area of this cut
    aArea = pNode.Switching;
    if pCut.nLeaves == 1 {
        return aArea; }
    // go through the children
    for  i = 0; i < pCut.nLeaves; i++  {
        pNodeChild = pCut.ppLeaves[i];
        assert(pNodeChild.nRefs > 0, "pNodeChild.nRefs > 0");
        if --pNodeChild.nRefs > 0 {  
            continue; }
        aArea += Fpga_CutDerefSwitch( pMan, pNodeChild, pNodeChild.pCutBest, fFanouts );
    }
    return aArea;
}

/**Function*************************************************************

  Synopsis    [Computes the array of mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingGetSwitching(pMan *Fpga_Man_t, vMapping *Fpga_NodeVec_t) float32 {
    var pNode *Fpga_Node_t;
    var Switch float32;
    var i int32;
    Switch = 0.0;
    for  i = 0; i < vMapping.nSize; i++  {
        pNode = vMapping.pArray[i];
        // at least one phase has the best cut assigned
        assert(!Fpga_NodeIsAnd(pNode) || pNode.pCutBest != nil, "!Fpga_NodeIsAnd(pNode) || pNode.pCutBest != nil");
        // at least one phase is used in the mapping
        assert(pNode.nRefs > 0, "pNode.nRefs > 0");
        // compute the array due to the supergate
        Switch += pNode.Switching;
    }
    // add buffer for each CO driven by a CI
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        // inlining of macro Fpga_IsComplement
        if Fpga_NodeIsVar(((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) ) && !((int32(((ABC_PTRUINT_T) (pMan.pOutputs[i]) & 01)))) {
            // inlining of macro Fpga_Regular
            Switch += ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) .Switching; }
    }
    return Switch;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

