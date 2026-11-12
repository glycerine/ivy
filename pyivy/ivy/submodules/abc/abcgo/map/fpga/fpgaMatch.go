package fpga

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [fpgaMatch.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Technology mapping for variable-size-LUT FPGAs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaMatch.c,v 1.7 2004/09/30 21:18:10 satrajit Exp $]

***********************************************************************/

//#include "fpgaInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int          Fpga_MatchNode( Fpga_Man_t * p, Fpga_Node_t * pNode, int fDelayOriented );

// static int          Fpga_MatchNodeArea( Fpga_Man_t * p, Fpga_Node_t * pNode );

// static int          Fpga_MatchNodeSwitch( Fpga_Man_t * p, Fpga_Node_t * pNode );


// static Fpga_Cut_t * Fpga_MappingAreaWithoutNode( Fpga_Man_t * p, Fpga_Node_t * pFanout, Fpga_Node_t * pNodeNo );

// static int          Fpga_MappingMatchesAreaArray( Fpga_Man_t * p, Fpga_NodeVec_t * vNodes );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Finds the best delay assignment of LUTs.]

  Description [This procedure iterates through all the nodes
  of the object graph reachable from the POs and assigns the best
  match to each of them. If the flag fDelayOriented is set to 1, it 
  tries to minimize the arrival time and uses the area flow as a
  tie-breaker. If the flag is set to 0, it considers all the cuts,
  whose arrival times matches the required time at the node, and
  minimizes the area flow using the arrival time as a tie-breaker.
  
  Before this procedure is called, the required times should be set
  and the fanout counts should be computed. In the first iteration,
  the required times are set to very large number (by NodeCreate) 
  and the fanout counts are set to the number of fanouts in the AIG.
  In the following iterations, the required times are set by the
  backward traversal, while the fanouts are estimated approximately.
  
  If the arrival times of the PI nodes are given, they should be 
  assigned to the PIs after the cuts are computed and before this
  procedure is called for the first time.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingMatches(p *Fpga_Man_t, fDelayOriented int32) int32 {
    var pProgress *ProgressBar;
    var pNode *Fpga_Node_t;
    var i, nNodes int32;
    
    // assign the arrival times of the PIs
    for  i = 0; i < p.nInputs; i++  {
        p.pInputs[i].pCutBest.tArrival = p.pInputArrivals[i];
    }

    // match LUTs with nodes in the topological order
    nNodes = p.vAnds.nSize;
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    for  i = 0; i < nNodes; i++  {
        pNode = p.vAnds.pArray[i];
        if !Fpga_NodeIsAnd( pNode ) {
            continue; }
        // skip a secondary node
        if pNode.pRepr {
            continue; }
        // match the node
        Fpga_MatchNode( p, pNode, fDelayOriented );
        Extra_ProgressBarUpdate( pProgress, i, "Matches ..." );
    }
    Extra_ProgressBarStop( pProgress );
/*
    if ( !fDelayOriented )
    {
        float Area = 0.0;
        for ( i = 0; i < p->nOutputs; i++ )
        {
            printf( "%5.2f ", Fpga_Regular(p->pOutputs[i])->pCutBest->aFlow );
            Area += Fpga_Regular(p->pOutputs[i])->pCutBest->aFlow;
        }
        printf( "\nTotal = %5.2f\n", Area );
    }
*/
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes the best matching for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MatchNode(p *Fpga_Man_t, pNode *Fpga_Node_t, fDelayOriented int32) int32 {
    var pCut, pCutBestOld *Fpga_Cut_t;
    var clk int64;
    // make sure that at least one cut other than the trivial is present
    if pNode.pCuts.pNext == nil {
        fmt.Printf("\nError: A node in the mapping graph does not have feasible cuts.\n");
        return 0;
    }

    // estimate the fanouts of the node
    if pNode.aEstFanouts < 0 {
        pNode.aEstFanouts = float32(pNode.nRefs);
    } else {
        pNode.aEstFanouts = float32(((2.0 * pNode.aEstFanouts + pNode.nRefs) / 3.0)); }
//        pNode->aEstFanouts = (float)pNode->nRefs;

    pCutBestOld = pNode.pCutBest;
    pNode.pCutBest = nil;
    for  pCut = pNode.pCuts.pNext; (pCut != nil); pCut = pCut.pNext  {
        // compute the arrival time of the cut and its area flow
clk = time.Now().UnixNano();
        Fpga_CutGetParameters( p, pCut );
//p->time2 += clock() - clk;
        // drop the cut if it does not meet the required times
        if Fpga_FloatMoreThan(p, pCut.tArrival, pNode.tRequired) {
            continue; }
        // if no cut is assigned, use the current one
        if pNode.pCutBest == nil {
            pNode.pCutBest = pCut;
            continue;
        }
        // choose the best cut using one of the two criteria:
        // (1) delay oriented mapping (first traversal), delay first, area-flow as a tie-breaker
        // (2) area recovery (subsequent traversals), area-flow first, delay as a tie-breaker
        if ((fDelayOriented != 0) && 
               (Fpga_FloatMoreThan(p, pNode.pCutBest.tArrival, pCut.tArrival) || 
                (Fpga_FloatEqual(p, pNode.pCutBest.tArrival, pCut.tArrival) && Fpga_FloatMoreThan(p, pNode.pCutBest.aFlow, pCut.aFlow)) )) ||
             (!(fDelayOriented != 0) && 
               (Fpga_FloatMoreThan(p, pNode.pCutBest.aFlow, pCut.aFlow) || 
                (Fpga_FloatEqual(p, pNode.pCutBest.aFlow, pCut.aFlow) && Fpga_FloatMoreThan(p, pNode.pCutBest.tArrival, pCut.tArrival)))) {
            pNode.pCutBest = pCut;
        }
    }

    // make sure the match is found
    if pNode.pCutBest == nil {
        if pCutBestOld == nil {
//            printf( "\nError: Could not match a node in the object graph.\n" );
            return 0;
        }
        pNode.pCutBest = pCutBestOld;
    }
    return 1;
}





/**Function*************************************************************

  Synopsis    [Finds the best area assignment of LUTs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingMatchesArea(p *Fpga_Man_t) int32 {
    var pProgress *ProgressBar;
    var pNode *Fpga_Node_t;
    var i, nNodes int32;
    
    // assign the arrival times of the PIs
    for  i = 0; i < p.nInputs; i++  {
        p.pInputs[i].pCutBest.tArrival = p.pInputArrivals[i];
    }

    // match LUTs with nodes in the topological order
    nNodes = p.vAnds.nSize;
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    for  i = 0; i < nNodes; i++  {
        pNode = p.vAnds.pArray[i];
        if !Fpga_NodeIsAnd( pNode ) {
            continue; }
        // skip a secondary node
        if pNode.pRepr {
            continue; }
        // match the node
        Fpga_MatchNodeArea( p, pNode );
        Extra_ProgressBarUpdate( pProgress, i, "Matches ..." );
    }
    Extra_ProgressBarStop( pProgress );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Finds the best area assignment of LUTs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingMatchesAreaArray(p *Fpga_Man_t, vNodes *Fpga_NodeVec_t) int32 {
    var pNode *Fpga_Node_t;
    var i int32;

    // match LUTs with nodes in the topological order
    for  i = 0; i < vNodes.nSize; i++  {
        pNode = vNodes.pArray[i];
        if !Fpga_NodeIsAnd( pNode ) {
            continue; }
        // skip a secondary node
        if pNode.pRepr {
            continue; }
        // match the node
        if !(Fpga_MatchNodeArea( p, pNode ) != 0) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes the best matching for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MatchNodeArea(p *Fpga_Man_t, pNode *Fpga_Node_t) int32 {
    var pCut, pCutBestOld *Fpga_Cut_t;
    var aAreaCutBest float32;
    var clk int64;
    // make sure that at least one cut other than the trivial is present
    if pNode.pCuts.pNext == nil {
        fmt.Printf("\nError: A node in the mapping graph does not have feasible cuts.\n");
        return 0;
    }

    // remember the old cut
    pCutBestOld = pNode.pCutBest;
    // deref the old cut
    if pNode.nRefs { 
        aAreaCutBest = Fpga_CutDeref( p, pNode, pNode.pCutBest, 0 ); }

    // search for a better cut
    pNode.pCutBest = nil;
    for  pCut = pNode.pCuts.pNext; (pCut != nil); pCut = pCut.pNext  {
        // compute the arrival time of the cut and its area flow
clk = time.Now().UnixNano();
        pCut.tArrival = Fpga_TimeCutComputeArrival( p, pCut );
//p->time2 += clock() - clk;
        // drop the cut if it does not meet the required times
        if Fpga_FloatMoreThan( p, pCut.tArrival, pNode.tRequired ) {
            continue; }
        // get the area of this cut
        pCut.aFlow = Fpga_CutGetAreaDerefed( p, pCut );
        // if no cut is assigned, use the current one
        if pNode.pCutBest == nil {
            pNode.pCutBest = pCut;
            continue;
        }
        // choose the best cut as follows: exact area first, delay as a tie-breaker
        if Fpga_FloatMoreThan(p, pNode.pCutBest.aFlow, pCut.aFlow) || 
             (Fpga_FloatEqual(p, pNode.pCutBest.aFlow, pCut.aFlow) && Fpga_FloatMoreThan(p, pNode.pCutBest.tArrival, pCut.tArrival)) {
            pNode.pCutBest = pCut;
        }
    }

    // make sure the match is found
    if pNode.pCutBest == nil {
        pNode.pCutBest = pCutBestOld; 
        // insert the new cut
        if pNode.nRefs { 
            pNode.pCutBest.aFlow = Fpga_CutRef( p, pNode, pNode.pCutBest, 0 ); }
//        printf( "\nError: Could not match a node in the object graph.\n" );
        return 0;
    }

    // insert the new cut
    // make sure the area selected is not worse then the original area
    if pNode.nRefs {
        pNode.pCutBest.aFlow = Fpga_CutRef( p, pNode, pNode.pCutBest, 0 );
//        assert( pNode->pCutBest->aFlow <= aAreaCutBest );
//        assert( pNode->tRequired < FPGA_FLOAT_LARGE );
    }
    return 1;
}




/**Function*************************************************************

  Synopsis    [Finds the best area assignment of LUTs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingMatchesSwitch(p *Fpga_Man_t) int32 {
    var pProgress *ProgressBar;
    var pNode *Fpga_Node_t;
    var i, nNodes int32;
    
    // assign the arrival times of the PIs
    for  i = 0; i < p.nInputs; i++  {
        p.pInputs[i].pCutBest.tArrival = p.pInputArrivals[i];
    }

    // match LUTs with nodes in the topological order
    nNodes = p.vAnds.nSize;
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    for  i = 0; i < nNodes; i++  {
        pNode = p.vAnds.pArray[i];
        if !Fpga_NodeIsAnd( pNode ) {
            continue; }
        // skip a secondary node
        if pNode.pRepr {
            continue; }
        // match the node
        Fpga_MatchNodeSwitch( p, pNode );
        Extra_ProgressBarUpdate( pProgress, i, "Matches ..." );
    }
    Extra_ProgressBarStop( pProgress );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes the best matching for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MatchNodeSwitch(p *Fpga_Man_t, pNode *Fpga_Node_t) int32 {
    var pCut, pCutBestOld *Fpga_Cut_t;
    var aAreaCutBest float32  = FPGA_FLOAT_LARGE;
    var clk int64;
    // make sure that at least one cut other than the trivial is present
    if pNode.pCuts.pNext == nil {
        fmt.Printf("\nError: A node in the mapping graph does not have feasible cuts.\n");
        return 0;
    }

    // remember the old cut
    pCutBestOld = pNode.pCutBest;
    // deref the old cut
    if pNode.nRefs { 
        aAreaCutBest = Fpga_CutDerefSwitch( p, pNode, pNode.pCutBest, 0 ); }

    // search for a better cut
    pNode.pCutBest = nil;
    for  pCut = pNode.pCuts.pNext; (pCut != nil); pCut = pCut.pNext  {
        // compute the arrival time of the cut and its area flow
clk = time.Now().UnixNano();
        pCut.tArrival = Fpga_TimeCutComputeArrival( p, pCut );
//p->time2 += clock() - clk;
        // drop the cut if it does not meet the required times
        if Fpga_FloatMoreThan( p, pCut.tArrival, pNode.tRequired ) {
            continue; }
        // get the area of this cut
        pCut.aFlow = Fpga_CutGetSwitchDerefed( p, pNode, pCut );
        // if no cut is assigned, use the current one
        if pNode.pCutBest == nil {
            pNode.pCutBest = pCut;
            continue;
        }
        // choose the best cut as follows: exact area first, delay as a tie-breaker
        if Fpga_FloatMoreThan(p, pNode.pCutBest.aFlow, pCut.aFlow) || 
             (Fpga_FloatEqual(p, pNode.pCutBest.aFlow, pCut.aFlow) && Fpga_FloatMoreThan(p, pNode.pCutBest.tArrival, pCut.tArrival)) {
            pNode.pCutBest = pCut;
        }
    }

    // make sure the match is found
    if pNode.pCutBest == nil {
        pNode.pCutBest = pCutBestOld; 
        // insert the new cut
        if pNode.nRefs { 
            pNode.pCutBest.aFlow = Fpga_CutRefSwitch( p, pNode, pNode.pCutBest, 0 ); }
//        printf( "\nError: Could not match a node in the object graph.\n" );
        return 0;
    }

    // insert the new cut
    // make sure the area selected is not worse then the original area
    if pNode.nRefs {
        pNode.pCutBest.aFlow = Fpga_CutRefSwitch( p, pNode, pNode.pCutBest, 0 );
        assert(pNode.pCutBest.aFlow <= aAreaCutBest + 0.001, "pNode.pCutBest.aFlow <= aAreaCutBest + 0.001");
//        assert( pNode->tRequired < FPGA_FLOAT_LARGE );
    }
    return 1;
}


     
                                                                        

                                   

                                                                       
               
                

                

                                                                        
                                      
 
                          
                        
          

                                       
     
                                             
                               
     

                                           
     
                                    
                                       
                     
                                      
                     
                                           
                     
                                          
     

                               
                              
                                          
                   
                                                            

                                           
     
                                   
                         

                                    
                                       
                     
                                      
                     
                                                                                                        
                     

                                     
                                
                     

                                            
                                                  

                                               

                                                    
                                                                
                                         
            
                                            
                                      

                                            
                
     
                                                           
                                                                                                    
 



                                                                        

                                   

                                                                       
               
                

                

                                                                        
                                       
 
                          
                               
                               
                                   
                                               
             

                                       
     
                                             
                               
     

                                           
     
                                    
                                       
                     
                                      
                     
                                           
                     
                                          
     

                               
                              
                                          
                   
                                                            

                  
                                           
     
                                    
                                       
                     
                                      
                     
                                                                                                           
                     

                                   

                                                               
                                                      
         
                                                 
                                             
                                                                            
                                       
                      
         
                                          
         
                                                         
                     
         


                                                        
                 
                                                      
         
                                                 
                            
                                                           
                             
                                             
                          
                                                         
                               
                                  
         

                                      

                               
                 
                                                      
         
                                                 
                            
                                                           
                             
                                             
                          
                                                         
                               
                                  
         
                               


                                                                             
                                                                     

                       
                              
     
                                                           
                                                                                                  
 


                                                                        

                                                                   

                                                                   
               
                

                

                                                                        
                                                                                                      
 
                                                
                       
          
                
                                                                        
                                      
     
                                                                                        
                 
     

                               

                           
                                  
                        
                                                             

                              
                           
                                                                
     
                                                                
              
                                           
                            
                                                              
                                                
                     

                                                  
                                             
                                               
                      
                                 
                     

                                   
                                                     
                                                     
                                      
         
                                   
                     
         
                                                                                   
                                                      
                                                                                                   
         
                                   
         
     

                                   
                                  
     
                                      
                             
                                                                         
                    
     

                              
                                  

                         
                                                                     

                                                                      
                                                     
                                                  
                   
 

      


/**function*************************************************************

  synopsis    [Performs area minimization using a heuristic algorithm.]

  description []
               
  sideeffects []

  seealso     []

***********************************************************************/
func Fpga_FindBestNode(p *Fpga_Man_t, vNodes *Fpga_NodeVec_t, ppNode **Fpga_Node_t, ppCutBest **Fpga_Cut_t) float32 {
    var pNode *Fpga_Node_t;
    var pCut *Fpga_Cut_t;
    var Gain, CutArea1, CutArea2, CutArea3 float32;
    var i int32;

    Gain = 0;
    for  i = 0; i < vNodes.nSize; i++  {
        pNode = vNodes.pArray[i];
        // deref the current cut
        CutArea1 = Fpga_CutDeref( p, pNode, pNode.pCutBest, 0 );

        // ref all the cuts
        for  pCut = pNode.pCuts.pNext; (pCut != nil); pCut = pCut.pNext  {
            if pCut == pNode.pCutBest {
                continue; }
            if pCut.tArrival > pNode.tRequired {
                continue; }

            CutArea2 = Fpga_CutGetAreaDerefed( p, pCut );
            if Gain < CutArea1 - CutArea2 {
                *ppNode = pNode;
                *ppCutBest = pCut;
                Gain = CutArea1 - CutArea2;
            }
        }
        // ref the old cut
        CutArea3 = Fpga_CutRef( p, pNode, pNode.pCutBest, 0 );
        assert(CutArea1 == CutArea3, "CutArea1 == CutArea3");
    }
    if Gain == 0 {
        fmt.Printf("Returning no gain.\n"); }

    return Gain;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

