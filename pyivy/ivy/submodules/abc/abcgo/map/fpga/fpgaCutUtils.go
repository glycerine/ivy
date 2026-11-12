package fpga

import "fmt"

/**CFile****************************************************************

  FileName    [fpgaCutUtils.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaCutUtils.h,v 1.0 2003/09/08 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fpgaInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutAlloc(p *Fpga_Man_t) *Fpga_Cut_t {
    var pCut *Fpga_Cut_t;
    pCut = (Fpga_Cut_t *)Extra_MmFixedEntryFetch( p.mmCuts );
    memset( pCut, 0, sizeof(Fpga_Cut_t) );
    return pCut;
}

/**Function*************************************************************

  Synopsis    [Duplicates the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutDup(p *Fpga_Man_t, pCutOld *Fpga_Cut_t) *Fpga_Cut_t {
    var pCutNew *Fpga_Cut_t;
    var i int32;
    pCutNew = Fpga_CutAlloc( p );
    pCutNew.pRoot   = pCutOld.pRoot;
    pCutNew.nLeaves = pCutOld.nLeaves;
    for  i = 0; i < pCutOld.nLeaves; i++  {
        pCutNew.ppLeaves[i] = pCutOld.ppLeaves[i];
    }
    return pCutNew;
}

/**Function*************************************************************

  Synopsis    [Deallocates the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutFree(p *Fpga_Man_t, pCut *Fpga_Cut_t) {
    if pCut != nil {
    Extra_MmFixedEntryRecycle( p.mmCuts, (char *)pCut ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutPrint(p *Fpga_Man_t, pRoot *Fpga_Node_t, pCut *Fpga_Cut_t) {
    var i int32;
    fmt.Printf("CUT:  Delay = %4.2f. Area = %4.2f. Nodes = %d -> {", pCut.tArrival, pCut.aFlow, pRoot.Num);
    for  i = 0; i < pCut.nLeaves; i++  {
        fmt.Printf(" %d", pCut.ppLeaves[i].Num);
    }
    fmt.Printf(" } \n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutCreateSimple(p *Fpga_Man_t, pNode *Fpga_Node_t) *Fpga_Cut_t {
    var pCut *Fpga_Cut_t;
    pCut = Fpga_CutAlloc( p );
    pCut.pRoot       = pNode;
    pCut.nLeaves     = 1;
    pCut.ppLeaves[0] = pNode;
    // inlining of macro FPGA_SEQ_SIGN
    pCut.uSign = (1 << (((ABC_PTRUINT_T)pCut.ppLeaves[0])%31));;
    return pCut;
}


/**function*************************************************************

  synopsis    [Computes the exact area associated with the cut.]

  description []
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutGetRootArea(p *Fpga_Man_t, pCut *Fpga_Cut_t) float32 {
    return p.pLutLib.pLutAreas[int32(pCut.nLeaves)];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutListAppend(pSetAll *Fpga_Cut_t, pSets *Fpga_Cut_t) *Fpga_Cut_t {
    var pPrev *Fpga_Cut_t  = nil; // Suppress "might be used uninitialized"
    var pTemp *Fpga_Cut_t;
    if pSetAll == nil {
        return pSets; }
    if pSets == nil {
        return pSetAll; }
    // find the last one
    for  pTemp = pSets; (pTemp != nil); pTemp = pTemp.pNext  {
        pPrev = pTemp;
    }
    // append all the end of the current set
    assert(pPrev.pNext == nil, "pPrev.pNext == nil");
    pPrev.pNext = pSetAll;
    return pSets;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutListRecycle(p *Fpga_Man_t, pSetList *Fpga_Cut_t, pSave *Fpga_Cut_t) {
    var pNext, pTemp *Fpga_Cut_t;
    for  pTemp = pSetList, pNext = pTemp? pTemp.pNext : nil; 
          (pTemp != nil); 
          pTemp = pNext, pNext = pNext? pNext.pNext : nil  {
        if pTemp != pSave {
            Extra_MmFixedEntryRecycle( p.mmCuts, (char *)pTemp ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutListCount(pSets *Fpga_Cut_t) int32 {
    var pTemp *Fpga_Cut_t;
    var i int32;
    for  i, pTemp = 0, pSets; (pTemp != nil); pTemp, i = pTemp.pNext, i+1  {}
    return i;
}

     

                                                                        

                                               

                
               
                

                

                                                                        
                                                                                    
 
                              
             
                                         
     
                                               
                                               
                                               
                      
                                       
                                             
                                                        
                          
     
 

                                                                        

                                               

                
               
                

                

                                                                        
                                                                                    
 
          
                                         
                                                               
 
      

/**Function*************************************************************

  Synopsis    [Computes the arrival time and the area flow of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutGetParameters(pMan *Fpga_Man_t, pCut *Fpga_Cut_t) {
    var pFaninCut *Fpga_Cut_t;
    var i int32;
    pCut.tArrival = -FPGA_FLOAT_LARGE;
    pCut.aFlow    = pMan.pLutLib.pLutAreas[int32(pCut.nLeaves)];
    for  i = 0; i < pCut.nLeaves; i++  {
        pFaninCut = pCut.ppLeaves[i].pCutBest;
        if pCut.tArrival < pFaninCut.tArrival {
             pCut.tArrival = pFaninCut.tArrival; }
        // if the fanout count is not set, assume it to be 1
        if pCut.ppLeaves[i].nRefs == 0 {
            pCut.aFlow += pFaninCut.aFlow;
        } else {
//            pCut->aFlow += pFaninCut->aFlow / pCut->ppLeaves[i]->nRefs;
            pCut.aFlow += pFaninCut.aFlow / pCut.ppLeaves[i].aEstFanouts; }
    }
    // use the first pin to compute the delay of the LUT 
    // (this mapper does not support the variable pin delay model)
    pCut.tArrival += pMan.pLutLib.pLutDelays[int32(pCut.nLeaves)][0];
}


/**function*************************************************************

  synopsis    [Computes the area flow of the cut.]

  description []
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutGetAreaFlow(pMan *Fpga_Man_t, pCut *Fpga_Cut_t) float32 {
    var pCutFanin *Fpga_Cut_t;
    var i int32;
    pCut.aFlow = pMan.pLutLib.pLutAreas[int32(pCut.nLeaves)];
    for  i = 0; i < pCut.nLeaves; i++  {
        // get the cut implementing this phase of the fanin
        pCutFanin  = pCut.ppLeaves[i].pCutBest;   
        assert((pCutFanin != nil), "pCutFanin");
        pCut.aFlow  += pCutFanin.aFlow / pCut.ppLeaves[i].nRefs;
    }
    return pCut.aFlow;
}

/**function*************************************************************

  synopsis    [Computes the exact area associated with the cut.]

  description []
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutGetAreaRefed(pMan *Fpga_Man_t, pCut *Fpga_Cut_t) float32 {
    var aResult, aResult2 float32;
    if pCut.nLeaves == 1 {
        return 0; }
    aResult  = Fpga_CutDeref( pMan, nil, pCut, 0 );
    aResult2 = Fpga_CutRef( pMan, nil, pCut, 0 );
    assert(Fpga_FloatEqual( pMan, aResult, aResult2 ), "Fpga_FloatEqual( pMan, aResult, aResult2 )");
    return aResult;
}

/**function*************************************************************

  synopsis    [Computes the exact area associated with the cut.]

  description []
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutGetAreaDerefed(pMan *Fpga_Man_t, pCut *Fpga_Cut_t) float32 {
    var aResult, aResult2 float32;
    if pCut.nLeaves == 1 {
        return 0; }
    aResult2 = Fpga_CutRef( pMan, nil, pCut, 0 );
    aResult  = Fpga_CutDeref( pMan, nil, pCut, 0 );
    assert(Fpga_FloatEqual( pMan, aResult, aResult2 ), "Fpga_FloatEqual( pMan, aResult, aResult2 )");
    return aResult;
}

/**function*************************************************************

  synopsis    [References the cut.]

  description [This procedure is similar to the procedure NodeReclaim.]
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutRef(pMan *Fpga_Man_t, pNode *Fpga_Node_t, pCut *Fpga_Cut_t, fFanouts int32) float32 {
    var pNodeChild *Fpga_Node_t;
    var aArea float32;
    var i int32;

    // deref the fanouts
//    if ( fFanouts ) 
//        Fpga_CutInsertFanouts( pMan, pNode, pCut );

    // start the area of this cut
    aArea = pMan.pLutLib.pLutAreas[int32(pCut.nLeaves)];
    // go through the children
    for  i = 0; i < pCut.nLeaves; i++  {
        pNodeChild = pCut.ppLeaves[i];
        assert(pNodeChild.nRefs >= 0, "pNodeChild.nRefs >= 0");
        if pNodeChild.nRefs++ > 0 {  
            continue; }
        if !Fpga_NodeIsAnd(pNodeChild) { 
            continue; }
        aArea += Fpga_CutRef( pMan, pNodeChild, pNodeChild.pCutBest, fFanouts );
    }
    return aArea;
}

/**function*************************************************************

  synopsis    [Dereferences the cut.]

  description [This procedure is similar to the procedure NodeRecusiveDeref.]
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_CutDeref(pMan *Fpga_Man_t, pNode *Fpga_Node_t, pCut *Fpga_Cut_t, fFanouts int32) float32 {
    var pNodeChild *Fpga_Node_t;
    var aArea float32;
    var i int32;

    // deref the fanouts
//    if ( fFanouts ) 
//        Fpga_CutRemoveFanouts( pMan, pNode, pCut );

    // start the area of this cut
    aArea = pMan.pLutLib.pLutAreas[int32(pCut.nLeaves)];
    // go through the children
    for  i = 0; i < pCut.nLeaves; i++  {
        pNodeChild = pCut.ppLeaves[i];
        assert(pNodeChild.nRefs > 0, "pNodeChild.nRefs > 0");
        if --pNodeChild.nRefs > 0 {  
            continue; }
        if !Fpga_NodeIsAnd(pNodeChild) { 
            continue; }
        aArea += Fpga_CutDeref( pMan, pNodeChild, pNodeChild.pCutBest, fFanouts );
    }
    return aArea;
}


/**Function*************************************************************

  Synopsis    [Sets the used cuts to be the currently selected ones.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingSetUsedCuts(pMan *Fpga_Man_t) {
    var i int32;
    for  i = 0; i < pMan.vNodesAll.nSize; i++  {
        if pMan.vNodesAll.pArray[i].pCutOld {
            pMan.vNodesAll.pArray[i].pCutBest = pMan.vNodesAll.pArray[i].pCutOld;
            pMan.vNodesAll.pArray[i].pCutOld  = nil;
        }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

