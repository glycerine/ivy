package place

/*===================================================================*/
//  
//     place_inc.c
//
//        Aaron P. Hurst, 2003-2007
//              ahurst@eecs.berkeley.edu
//
/*===================================================================*/

//#include <stdlib.h>
//#include <limits.h>
//#include <assert.h>
//#include <string.h>

//#include "place_base.h"
//#include "place_gordian.h"

//ABC_NAMESPACE_IMPL_START


func sqHashId(id int32, max int32) int32 {
  return ((id * (id+17)) % max);
}

     
                                                                       
              
  
                                                 
                                                                       
                                                     
                                                   
  
                                     
                           
                                                     
                           
                    
                 

                                  

                     
                                             
                          
                                      
                                       

                             
                                
                                                  
                                                  
   

                                       
                          
                                        
                                             
                        
                                  
     

                             
                                                                  
                         
   

                          
                                        
                                             
                        
                                                        
     

                       
  
                      
                                      
                               
                                         
                            
                                         
       
     
 
      

// --------------------------------------------------------------------
// fastEstimate()
//
// --------------------------------------------------------------------
func fastEstimate(cell *ConcreteCell, numNets int32, nets []*ConcreteNet) float32 {
  var len float32;
  var n int32;
  var box Rect;

  assert((cell != nil), "cell");

  for n=0; n<numNets; n++ {
    box = getNetBBox(nets[n]);
    if cell.m_x < box.x { len += (box.x - cell.m_x); }
    if cell.m_x > box.x+box.w { len += (cell.m_x-box.x-box.w); }
    if cell.m_y < box.y { len += (box.x - cell.m_y); }
    if cell.m_y > box.y+box.h { len += (cell.m_y-box.y-box.h); }
  }
  
  return len;
}
//ABC_NAMESPACE_IMPL_END

