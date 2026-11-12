package if_

import "fmt"

/**CFile****************************************************************

  FileName    [ifUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifUtil.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets all the node copy to NULL.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCleanNodeCopy(p *If_Man_t) {
    var pObj *If_Obj_t;
    var i int32;
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        If_ObjSetCopy( pObj, nil );
    }
}

/**Function*************************************************************

  Synopsis    [Sets all the cut data to NULL.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCleanCutData(p *If_Man_t) {
    var pObj *If_Obj_t;
    var i int32;
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        If_CutSetData( If_ObjCutBest(pObj), nil );
    }
}

/**Function*************************************************************

  Synopsis    [Sets all visited marks to 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCleanMarkV(p *If_Man_t) {
    var pObj *If_Obj_t;
    var i int32;
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        pObj.fVisit = 0;
    }
}
 
     

                                                                        

                                                                         

                
               
                

                

                                                                        
                                                                                 
 
                     
                        
                
          
                                                                    
                   
                                               
                                
                                                
                                
                                                     
                                   
                                  
                                         
                                              
                                                            
                 
 

                                                                        

                                                                         

                                                                        
               
               
                

                

                                                                        
                                       
 
                                
                
          
                                     
                           
                 
                                  
     
                                        
                                           
                           
     
                                        
                                                        
                                                                  
                                                                              
              
                                 
                                                                         
                                                       
                               
                                         
                                                           
                                            
                        
                 
 

                                                                        

                                                                         

                                                                        
               
               
                

                

                                                                        
                                             
 
                                
                
          
                                     
                           
                                  
     
                                        
                                           
                           
     
                                        
                                                        
                                                                  
                                                                              
              
                                 
                                                                         
                                                       
                               
                                           
                                         
                                                           
                                            
                        
                 
 

                                                                        

                                                                         

                
               
                

                

                                                                        
                                                                                    
 
                     
                        
                
                 
                                  
                              
                                                                          
                             
                                                                    
                   
                                               
                                
                                                     
                                   
                                                                       
                                                        
                                                               
                   
                                 
                 
 

                                                                        

                                                                         

                                                                        
               
               
                

                

                                                                        
                                          
 
                    
                
          
                                    
                           
                                  
                        
                                                                              
              
                               
                                 
                                                                               
                 
 

      

/**Function*************************************************************

  Synopsis    [Computes area, references, and nodes used in the mapping.]

  Description [Collects the nodes in reverse topological order in array 
  p->vMapping.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManResetOriginalRefs(p *If_Man_t) {
    var pObj *If_Obj_t;
    var i int32;
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        pObj.nRefs = 0;
    }
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if If_ObjIsAnd(pObj) {
            pObj.pFanin0.nRefs++;
            pObj.pFanin1.nRefs++;
        
        } else if If_ObjIsCo(pObj) {
            pObj.pFanin0.nRefs++; }
    }
}

/**Function*************************************************************

  Synopsis    [Computes cross-cut of the circuit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCrossCut(p *If_Man_t) int32 {
    var pObj, pFanin *If_Obj_t;
    var i, nCutSize, nCutSizeMax int32;
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if !If_ObjIsAnd(pObj) {
            continue; }
        // consider the node
        if nCutSizeMax < ++nCutSize {
            nCutSizeMax = nCutSize; }
        if pObj.nVisits == 0 {
            nCutSize--; }
        // consider the fanins
        pFanin = If_ObjFanin0(pObj);
        if !If_ObjIsCi(pFanin) && --pFanin.nVisits == 0 {
            nCutSize--; }
        pFanin = If_ObjFanin1(pObj);
        if !If_ObjIsCi(pFanin) && --pFanin.nVisits == 0 {
            nCutSize--; }
        // consider the choice class
        if pObj.fRepr {
            for  pFanin = pObj; (pFanin != nil); pFanin = pFanin.pEquiv  {
                if !If_ObjIsCi(pFanin) && --pFanin.nVisits == 0 {
                    nCutSize--; }
            } }
    }
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        assert(If_ObjIsCi(pObj) || pObj.fVisit == 0, "If_ObjIsCi(pObj) || pObj.fVisit == 0");
        pObj.nVisits = pObj.nVisitsCopy;
    }
    assert(nCutSize == 0, "nCutSize == 0");
//    Abc_Print( 1, "Max cross cut size = %6d.\n", nCutSizeMax );
    return nCutSizeMax;
}

/**Function*************************************************************

  Synopsis    [Computes the reverse topological order of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManReverseOrder(p *If_Man_t) *Vec_Ptr_t {
    var vOrder *Vec_Ptr_t;
    var pObj *If_Obj_t;
    var ppStore **If_Obj_t;
    var i int32;
    // allocate place to store the nodes
    // inlining of macro ABC_ALLOC
    ppStore = ((If_Obj_t * *) malloc(sizeof(If_Obj_t *) * (size_t)(p.nLevelMax + 1)));
    memset( ppStore, 0, sizeof(If_Obj_t *) * (p.nLevelMax + 1) );
    // add the nodes
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        assert(pObj.Level >= 0 && pObj.Level <= uint32(p.nLevelMax), "pObj.Level >= 0 && pObj.Level <= uint32(p.nLevelMax)");
        pObj.pCopy = (char *)ppStore[pObj.Level]; 
        ppStore[pObj.Level] = pObj;
    }
    vOrder = Vec_PtrAlloc( If_ManObjNum(p) );
    for  i = p.nLevelMax; i >= 0; i--  {
        for  pObj = ppStore[i]; (pObj != nil); pObj = (If_Obj_t *)pObj.pCopy  {
            Vec_PtrPush( vOrder, pObj );
        }
    }
    // inlining of macro ABC_FREE
    ((ppStore) ? (free((char *) (ppStore)), (ppStore) = 0) : 0);
    // print the order
//    Vec_PtrForEachEntry( If_Obj_t *, vOrder, pObj, i )
//        Abc_Print( 1, "Obj %2d   Type %d  Level = %d\n", pObj->Id, pObj->Type, pObj->Level );
    return vOrder;
}
 
/**Function*************************************************************

  Synopsis    [Computes area, references, and nodes used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManMarkMapping_rec(p *If_Man_t, pObj *If_Obj_t) float32 {
    var pLeaf *If_Obj_t;
    var pCutBest *If_Cut_t;
    var pSwitching *float32;
    if p.vSwitching {
        pSwitching = (float*)p.vSwitching.pArray;
    } else {
        pSwitching = nil;
    }
    var aArea float32;
    var i int32;
    if pObj.nRefs++ || If_ObjIsCi(pObj) || If_ObjIsConst1(pObj) {
        return 0.0; }
    // store the node in the structure by level
    assert(If_ObjIsAnd(pObj), "If_ObjIsAnd(pObj)");
    // visit the transitive fanin of the selected cut
    pCutBest = If_ObjCutBest(pObj);
    p.nNets += pCutBest.nLeaves;
    aArea = If_CutLutArea( p, pCutBest );
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCutBest).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCutBest).pLeaves[i])); i++  {
        p.dPower += pSwitching? pSwitching[pLeaf.Id] : 0.0;
        aArea += If_ManMarkMapping_rec( p, pLeaf );
    }
    return aArea;
}

/**Function*************************************************************

  Synopsis    [Computes area, references, and nodes used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManMarkMapping(p *If_Man_t) {
    var pObj *If_Obj_t;
    var i int32;
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        pObj.Required = IF_FLOAT_LARGE;
        pObj.nVisits = pObj.nVisitsCopy;
        pObj.nRefs = 0;
    }
    p.nNets = 0;
    p.dPower = 0.0;
    p.AreaGlo = 0.0;
    // inlining of macro If_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        p.AreaGlo += If_ManMarkMapping_rec( p, If_ObjFanin0(pObj) );
    }
}

/**Function*************************************************************

  Synopsis    [Collects nodes used in the mapping in the topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCollectMappingDirect(p *If_Man_t) *Vec_Ptr_t {
    var vOrder *Vec_Ptr_t;
    var pObj *If_Obj_t;
    var i int32;
    If_ManMarkMapping( p );
    vOrder = Vec_PtrAlloc( If_ManObjNum(p) );
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if If_ObjIsAnd(pObj) && pObj.nRefs {
            Vec_PtrPush( vOrder, pObj ); }
    }
    return vOrder;
}

/**Function*************************************************************

  Synopsis    [Collects nodes used in the mapping in the topological order.]

  Description [Represents mapping as an array of integers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCollectMappingInt(p *If_Man_t) *Vec_Int_t {
    var vOrder *Vec_Int_t;
    var pCutBest *If_Cut_t;
    var pObj *If_Obj_t;
    var i int32;
    var k int32;
    var nLeaves int32;
    var ppLeaves *int32;
    If_ManMarkMapping( p );
    vOrder = Vec_IntAlloc( If_ManObjNum(p) );
    // inlining of macro If_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if If_ObjIsAnd(pObj) && pObj.nRefs {
            pCutBest = If_ObjCutBest( pObj );
            nLeaves  = If_CutLeaveNum( pCutBest ); 
            ppLeaves = If_CutLeaves( pCutBest );
            // save the number of leaves, the leaves, and finally, the root
            Vec_IntPush( vOrder, nLeaves );
            for  k = 0; k < nLeaves; k++  {
                Vec_IntPush( vOrder, ppLeaves[k] );
            }
            Vec_IntPush( vOrder, pObj.Id );
        }
    }
    return vOrder;
}

/**Function*************************************************************

  Synopsis    [Returns the number of POs pointing to the same internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCountSpecialPos(p *If_Man_t) int32 {
    var pObj *If_Obj_t;
    var i, Counter int32;
    // clean all marks
    // inlining of macro If_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStartStop
    for  i = p.pPars.nLatchesCoBox; (i < If_ManCoNum(p) - p.pPars.nLatchesCo) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        If_ObjFanin0(pObj).fMark = 0;
    }
    // label nodes 
    // inlining of macro If_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStartStop
    for  i = p.pPars.nLatchesCoBox; (i < If_ManCoNum(p) - p.pPars.nLatchesCo) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if !If_ObjFaninC0(pObj) {
            If_ObjFanin0(pObj).fMark = 1; }
    }
    // label nodes 
    // inlining of macro If_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStartStop
    for  i = p.pPars.nLatchesCoBox; (i < If_ManCoNum(p) - p.pPars.nLatchesCo) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if If_ObjFaninC0(pObj) {
            Counter += If_ObjFanin0(pObj).fMark; }
    }
    // clean all marks
    // inlining of macro If_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStartStop
    for  i = p.pPars.nLatchesCoBox; (i < If_ManCoNum(p) - p.pPars.nLatchesCo) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        If_ObjFanin0(pObj).fMark = 0;
    }
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Traverse the cut and counts its volume.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutTraverse_rec(pNode *If_Obj_t, vNodes *Vec_Ptr_t) {
    if pNode.fMark {
        return; }
    pNode.fMark = 1;
//    assert( !If_ObjIsCi(pNode) ); // does not hold with cut minimization
    if If_ObjIsAnd(pNode) {
        If_CutTraverse_rec( If_ObjFanin0(pNode), vNodes ); }
    if If_ObjIsAnd(pNode) {
        If_CutTraverse_rec( If_ObjFanin1(pNode), vNodes ); }
    Vec_PtrPush( vNodes, pNode );
}
func If_CutTraverse(p *If_Man_t, pRoot *If_Obj_t, pCut *If_Cut_t, vNodes *Vec_Ptr_t) {
    var pLeaf *If_Obj_t;
    var i int32;
    // collect the internal nodes of the cut
    Vec_PtrClear( vNodes );
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(p, (pCut).pLeaves[i])); i++  {
        Vec_PtrPush( vNodes, pLeaf );
        assert(pLeaf.fMark == 0, "pLeaf.fMark == 0");
        pLeaf.fMark = 1;
    }
    // collect other nodes
    If_CutTraverse_rec( pRoot, vNodes );
    // clean the mark
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pLeaf) = (If_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pLeaf.fMark = 0;
    }
}
func If_CutTraverseTest(p *If_Man_t, pRoot *If_Obj_t, pCut *If_Cut_t) {
    var vNodes *Vec_Ptr_t;
    vNodes = Vec_PtrAlloc( 1000 );
    If_CutTraverse( p, pRoot, pCut, vNodes );
//if ( Vec_PtrSize(vNodes) > 30 )
//printf( "%d ", Vec_PtrSize(vNodes) );
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ObjPrint(pObj *If_Obj_t) {
    if pObj == nil {
        fmt.Printf("Object is NULL.");
        return;
    }
    fmt.Printf("Obj %4d : ", If_ObjId(pObj));
    if If_ObjIsConst1(pObj) {
        fmt.Printf("constant 1");
    } else if If_ObjIsCi(pObj) {
        fmt.Printf("PI");
    } else if If_ObjIsCo(pObj) {
        fmt.Printf("PO( %4d%s )", If_ObjId(If_ObjFanin0(pObj)), (If_ObjFaninC0(pObj)? "\'" : " "));
    } else {
        fmt.Printf("AND( %4d%s, %4d%s )", If_ObjId(If_ObjFanin0(pObj)), (If_ObjFaninC0(pObj)? "\'" : " "), If_ObjId(If_ObjFanin1(pObj)), (If_ObjFaninC1(pObj)? "\'" : " ")); }
    fmt.Printf(" (refs = %3d)", pObj.nVisitsCopy);
    fmt.Printf("\n");
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

