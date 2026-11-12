
package cnf

/**CFile****************************************************************

  FileName    [cnfMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG-to-CNF conversion.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnfMap.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cnf.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes area flow of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutAssignAreaFlow(p *Cnf_Man_t, pCut *Dar_Cut_t, pAreaFlows *int32) {
    var pLeaf *Aig_Obj_t;
    var i int32;
    pCut.Value = 0;
//    pCut->uSign = 100 * Cnf_CutSopCost( p, pCut );
    pCut.uSign = 10 * Cnf_CutSopCost( p, pCut );
    // inlining of macro Dar_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && (((pLeaf) = Aig_ManObj(p.pManAig, (pCut).pLeaves[i])), 1); i++  {
        pCut.Value += pLeaf.nRefs;
        if !Aig_ObjIsNode(pLeaf) {
            continue; }
        assert(pLeaf.nRefs > 0, "pLeaf.nRefs > 0");
        pCut.uSign += pAreaFlows[pLeaf.Id] / (pLeaf.nRefs? pLeaf.nRefs : 1);
    }
}

/**Function*************************************************************

  Synopsis    [Computes area flow of the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutSuperAreaFlow(vSuper *Vec_Ptr_t, pAreaFlows *int32) int32 {
    var pLeaf *Aig_Obj_t;
    var i, nAreaFlow int32;
    nAreaFlow = 100 * (Vec_PtrSize(vSuper) + 1);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pLeaf) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLeaf = Aig_Regular(pLeaf);
        if !Aig_ObjIsNode(pLeaf) {
            continue; }
        assert(pLeaf.nRefs > 0, "pLeaf.nRefs > 0");
        nAreaFlow += pAreaFlows[pLeaf.Id] / (pLeaf.nRefs? pLeaf.nRefs : 1);
    }
    return nAreaFlow;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveMapping(p *Cnf_Man_t) {
    var vSuper *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var pCut, pCutBest *Dar_Cut_t;
    var i int32;
    var k int32;
    var AreaFlow int32;
    var pAreaFlows *int32;
    // allocate area flows
    // inlining of macro ABC_ALLOC
    pAreaFlows = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(p.pManAig))));
    memset( pAreaFlows, 0, sizeof(int) * Aig_ManObjNumMax(p.pManAig) );
    // visit the nodes in the topological order and update their best cuts
    vSuper = Vec_PtrAlloc( 100 );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        // go through the cuts
        pCutBest = nil;
        // inlining of macro Dar_ObjForEachCut
        for  (pCut) = Dar_ObjCuts(pObj), k = 0; k < int32((pObj).nCuts); k, pCut = k+1, pCut+1  { if (pCut).fUsed==0 {} else {
            pCut.fBest = 0;
            if k == 0 {
                continue; }
            Cnf_CutAssignAreaFlow( p, pCut, pAreaFlows );
            if pCutBest == nil || pCutBest.uSign > pCut.uSign || 
                (pCutBest.uSign == pCut.uSign && pCutBest.Value < pCut.Value) {
                 pCutBest = pCut; }
        }
        }
        // check the big cut
//        Aig_ObjCollectSuper( pObj, vSuper );
        // get the area flow of this cut
//        AreaFlow = Cnf_CutSuperAreaFlow( vSuper, pAreaFlows );
        AreaFlow = ABC_INFINITY;
        if AreaFlow >= int32(pCutBest.uSign) {
            pAreaFlows[pObj.Id] = pCutBest.uSign;
            pCutBest.fBest = 1;
        
        } else {
            pAreaFlows[pObj.Id] = AreaFlow;
            pObj.fMarkB = 1; // mark the special node
        }
    }
    }
    Vec_PtrFree( vSuper );
    pAreaFlows = nil;

/*
    // compute the area of mapping
    AreaFlow = 0;
    Aig_ManForEachCo( p->pManAig, pObj, i )
        AreaFlow += Dar_ObjBestCut(Aig_ObjFanin0(pObj))->uSign / 100 / Aig_ObjFanin0(pObj)->nRefs;
    printf( "Area of the network = %d.\n", AreaFlow );
*/
}



     

                                                                        

                                                 

                                           
               
                

                
 
                                                                        
                                                    
  
                      
          
                                           
     
                                   
                                                         
                     
                                                 
     
 

                                                                        

                                                 

                                           
               
                

                

                                                                        
                                                 
 
                      
                              
                                           
     
                                    
                                                         
                     
                                                       
     
                
 

                                                                        

                                               

                
               
                

                

                                                                        
                                                  
 
             
                                 
                            
                
 

                                                                        

                                                      

                
               
                

                

                                                                        
                                                                    
 
                                         
                  
                                                                        
                 
                                      
                    
                                                                               
                   
                                                                      
                    
                                                                      
                                                                 
               
             
 

                                                                        

                                                            

                
               
                

                

                                                                        
                                                  
 
                                 
          
                    
                                      
                                                                      
                            
                    
 

                                                                        

                                              

                
               
                

                

                                                                        
                                                         
 
                      
          
                                      
                      
                      
                                                    
     
                                    
                     
                                
         
                                                      
                           
         
            
         
                                                    
                                   
                
                                              
         
     
 

                                                                        

                                                                             

                
               
                

                

                                                                        
                                     
 
                     
                                 
             
                                                                          
                                             
     
                                
                                        
                                      
                                                   
                          
                                                 

                                         
                                          
                                               
                                                                

                                
                                            
                                      
                                                 
                          
                                               
     
             
 

      

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

