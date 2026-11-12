package cnf

/**CFile****************************************************************

  FileName    [cnfCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG-to-CNF conversion.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnfCore.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cnf.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var s_pManCnf *Cnf_Man_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManPrepare() {
    if s_pManCnf == nil {
//        printf( "\n\nCreating CNF manager!!!!!\n\n" );
        s_pManCnf = Cnf_ManStart();
    }
}
func Cnf_ManRead() *Cnf_Man_t {
    return s_pManCnf;
}
func Cnf_ManFree() {
    if s_pManCnf == nil {
        return; }
    Cnf_ManStop( s_pManCnf );
    s_pManCnf = nil;
}


/**Function*************************************************************

  Synopsis    [Converts AIG into the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveMappingArray(pAig *Aig_Man_t) *Vec_Int_t {
    var vResult *Vec_Int_t;
    var p *Cnf_Man_t;
    var vMapped *Vec_Ptr_t;
    var pMemCuts *Aig_MmFixed_t;
    var clk abctime;
    // allocate the CNF manager
    p = Cnf_ManStart();
    p.pManAig = pAig;

    // generate cuts for all nodes, assign cost, and find best cuts
clk = Abc_Clock();
    pMemCuts = Dar_ManComputeCuts( pAig, 10, 0, 0 );
p.timeCuts = Abc_Clock() - clk;

    // find the mapping
clk = Abc_Clock();
    Cnf_DeriveMapping( p );
p.timeMap = Abc_Clock() - clk;
//    Aig_ManScanMapping( p, 1 );

    // convert it into CNF
clk = Abc_Clock();
    Cnf_ManTransferCuts( p );
    vMapped = Cnf_ManScanMapping( p, 1, 0 );
    vResult = Cnf_ManWriteCnfMapping( p, vMapped );
    Vec_PtrFree( vMapped );
    Aig_MmFixedStop( pMemCuts, 0 );
p.timeSave = Abc_Clock() - clk;

   // reset reference counters
    Aig_ManResetRefs( pAig );
//ABC_PRT( "Cuts   ", p->timeCuts );
//ABC_PRT( "Map    ", p->timeMap  );
//ABC_PRT( "Saving ", p->timeSave );
    Cnf_ManStop( p );
    return vResult;
}
 
/**Function*************************************************************

  Synopsis    [Converts AIG into the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveWithMan(p *Cnf_Man_t, pAig *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t;
    var vMapped *Vec_Ptr_t;
    var pMemCuts *Aig_MmFixed_t;
    var clk abctime;
    // connect the managers
    p.pManAig = pAig;

    // generate cuts for all nodes, assign cost, and find best cuts
clk = Abc_Clock();
    pMemCuts = Dar_ManComputeCuts( pAig, 10, 0, 0 );
p.timeCuts = Abc_Clock() - clk;

    // find the mapping
clk = Abc_Clock();
    Cnf_DeriveMapping( p );
p.timeMap = Abc_Clock() - clk;
//    Aig_ManScanMapping( p, 1 );

    // convert it into CNF
clk = Abc_Clock();
    Cnf_ManTransferCuts( p );
    vMapped = Cnf_ManScanMapping( p, 1, 1 );
    pCnf = Cnf_ManWriteCnf( p, vMapped, nOutputs );
    Vec_PtrFree( vMapped );
    Aig_MmFixedStop( pMemCuts, 0 );
p.timeSave = Abc_Clock() - clk;

   // reset reference counters
    Aig_ManResetRefs( pAig );
//ABC_PRT( "Cuts   ", p->timeCuts );
//ABC_PRT( "Map    ", p->timeMap  );
//ABC_PRT( "Saving ", p->timeSave );
    return pCnf;
}
func Cnf_Derive(pAig *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
    Cnf_ManPrepare();
    return Cnf_DeriveWithMan( s_pManCnf, pAig, nOutputs );
}
 
/**Function*************************************************************

  Synopsis    [Converts AIG into the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveOtherWithMan(p *Cnf_Man_t, pAig *Aig_Man_t, fSkipTtMin int32) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t;
    var vMapped *Vec_Ptr_t;
    var pMemCuts *Aig_MmFixed_t;
    var clk abctime;
    // connect the managers
    p.pManAig = pAig;

    // generate cuts for all nodes, assign cost, and find best cuts
clk = Abc_Clock();
    pMemCuts = Dar_ManComputeCuts( pAig, 10, fSkipTtMin, 0 );
p.timeCuts = Abc_Clock() - clk;

    // find the mapping
clk = Abc_Clock();
    Cnf_DeriveMapping( p );
p.timeMap = Abc_Clock() - clk;
//    Aig_ManScanMapping( p, 1 );

    // convert it into CNF
clk = Abc_Clock();
    Cnf_ManTransferCuts( p );
    vMapped = Cnf_ManScanMapping( p, 1, 1 );
    pCnf = Cnf_ManWriteCnfOther( p, vMapped );
    pCnf.vMapping = Cnf_ManWriteCnfMapping( p, vMapped );
    Vec_PtrFree( vMapped );
    Aig_MmFixedStop( pMemCuts, 0 );
p.timeSave = Abc_Clock() - clk;

   // reset reference counters
    Aig_ManResetRefs( pAig );
//ABC_PRT( "Cuts   ", p->timeCuts );
//ABC_PRT( "Map    ", p->timeMap  );
//ABC_PRT( "Saving ", p->timeSave );
    return pCnf;
}
func Cnf_DeriveOther(pAig *Aig_Man_t, fSkipTtMin int32) *Cnf_Dat_t {
    Cnf_ManPrepare();
    return Cnf_DeriveOtherWithMan( s_pManCnf, pAig, fSkipTtMin );
}

     

                                                                        

                

                
               
                

                

                                                                        
                                              
 
  
                                    
                                  
     
                  
                                   
                              
                                      
     
  
                     
                                         
                           

                  
                             

                            
                               
  
                            
                               
                            
                               
  
                                     

  
                                         
                                         
                           

               
                         
                            
                
  
                                   
                
 

      


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

