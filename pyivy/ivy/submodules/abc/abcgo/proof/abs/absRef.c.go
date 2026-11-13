package abs

/**CFile****************************************************************

  FileName    [absRef.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Refinement manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absRef.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sat/bsat/satSolver2.h"
//#include "abs.h"
//#include "absRef.h"

//ABC_NAMESPACE_IMPL_START

/*
    Description of the refinement manager

    This refinement manager should be 
    * started by calling Rnm_ManStart()
       this procedure takes one argument, the user's seq miter as a GIA manager
       - the manager should have only one property output
       - this manager should not change while the refinement manager is alive
       - it cannot be used by external applications for any purpose
       - when the refinement manager stop, GIA manager is the same as at the beginning
       - in the meantime, it will have some data-structures attached to its nodes...
    * stopped by calling Rnm_ManStop()
    * between starting and stopping, refinements are obtained by calling Rnm_ManRefine()

    Procedure Rnm_ManRefine() takes the following arguments:
    * the refinement manager previously started by Rnm_ManStart()
    * counter-example (CEX) obtained by abstracting some logic of GIA
    * mapping (vMap) of inputs of the CEX into the object IDs of the GIA manager
       - only PI, flop outputs, and internal AND nodes can be used in vMap
       - the ordering of objects in vMap is not important
       - however, the index of a non-PI object in vMap is used as its priority
         (the smaller the index, the more likely this non-PI object apears in a refinement)
       - only the logic between PO and the objects listed in vMap is traversed by the manager
         (as a result, GIA can be arbitrarily large, but only objects used in the abstraction
         and the pseudo-PI, that is, objects in the cut, will be visited by the manager)
    * flag fPropFanout defines whether value propagation is done through the fanout
       - it this flag is enabled, theoretically refinement should be better (the result smaller)
    * flag fVerbose may print some statistics

    The refinement manager returns a minimal-size array of integer IDs of GIA objects
    which should be added to the abstraction to possibly prevent the given counter-example
       - only flop output and internal AND nodes from vMap may appear in the resulting array
       - if the resulting array is empty, the CEX is a true CEX 
         (in other words, non-PI objects are not needed to set the PO value to 1)

    Verification of the selected refinement is performed by 
       - initializing all PI objects in vMap to value 0 or 1 they have in the CEX
       - initializing all remaining objects in vMap to value X
       - initializing objects used in the refiment to value 0 or 1 they have in the CEX
       - simulating through as many timeframes as required by the CEX
       - if the PO value in the last frame is 1, the refinement is correct
         (however, the minimality of the refinement is not currently checked)
        
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*
static inline int         Rnm_ObjSatVar( Rnm_Man_t * p, Gia_Obj_t * pObj )          { return Vec_IntEntry( p->vSatVars, Gia_ObjId(p->pGia, pObj) );                                                 }
static inline void        Rnm_ObjSetSatVar( Rnm_Man_t * p, Gia_Obj_t * pObj, int c) { Vec_IntWriteEntry( p->vSatVars, Gia_ObjId(p->pGia, pObj), c );                                                }
static inline int         Rnm_ObjFindOrAddSatVar( Rnm_Man_t * p, Gia_Obj_t * pObj)  { if ( Rnm_ObjSatVar(p, pObj) == 0 ) { Rnm_ObjSetSatVar(p, pObj, Vec_IntSize(p->vSat2Ids)); Vec_IntPush(p->vSat2Ids, Gia_ObjId(p->pGia, pObj)); }; return 2*Rnm_ObjSatVar(p, pObj); }
*/
// extern void               Ga2_ManCnfAddStatic( sat_solver2 * pSat, Vec_Int_t * vCnf0, Vec_Int_t * vCnf1, int * pLits, int iLitOut, int ProofId );

// extern Vec_Int_t *        Ga2_ManCnfCompute( unsigned uTruth, int nVars, Vec_Int_t * vCover );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

      

                                                                        

                                                     

                
               
                

                

                                                                        
                                                                                                          
 
                        
                       
          
                                          
               
                                     
                            
     
                                   
                                                      
               
     
                                 
                                       
                                                 
                                                                
                                                
 

                                                                      
 
                               
                                             
                                                          
                               
                                                                
                                                                      
                                                  
                                          
     
                                              
                             
                                       
     
                                                  
                                 
                                   
                                      
                                                   
                                                                        
                                                                        
                                          
                                      
                              
                                
                                                                                        
                           
                                                    
                                          
                             
                                                      
     
                                                 
                                                           
                                                           
                                                                                                     
                                                                                                     
                                                                                                                                
                             
                             
     

                     

                                    
                                                                                                    

                                                        
                                
                                     
                                                                   
     
                                                   
                              
                                                       
         
                                                                                       
                                                                                            
         

                                                      
                     
                                                            
                                                                           
                                                                                               
   
                                                        
                                
                                                      
                                                                
                      
                                     
                                    
            
  

                                                                                                                                                        
                                
                     
                                                      
                                    
                                
                                           

                                     
                                
                             
     
                                     
                                 
                         
                                    
                                  
                   

                                   
                                               
     
                                                        
                                                                            
                                                                             
     
                                          
                                                   
                                         
                                                                
                                 

                     
                            
                                 

                     

                                 
                                                      
     
                                       
                                                 
                                                           
                                             
     
                          
                            
                            
                      
 

      

/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collect internal objects to be used in value propagation.]

  Description [Resulting array vObjs contains RO, AND, PO/RI in a topo order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs sensitization analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Drive implications of the given node towards primary outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

