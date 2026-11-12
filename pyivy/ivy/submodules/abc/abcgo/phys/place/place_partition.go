package place

import "fmt"

/*===================================================================*/
//  
//     place_partition.c
//
//        Aaron P. Hurst, 2003-2007
//              ahurst@eecs.berkeley.edu
//
/*===================================================================*/

//#include <stdlib.h>
//#include <math.h>
//#include <string.h>
//#include <stdio.h>
//#include <limits.h>
//#include <assert.h>
//#include <sys/stat.h>
//#include <unistd.h>

//#include "place_base.h"
//#include "place_gordian.h"

#if !defined(NO_HMETIS)
#include "libhmetis.h"

ABC_NAMESPACE_IMPL_START

#endif

// --------------------------------------------------------------------
// Global variables
//
// --------------------------------------------------------------------

var g_place_rootPartition *Partition = nil;
var allNetsR2 **ConcreteNet = nil;
var allNetsL2 **ConcreteNet = nil;
var allNetsB2 **ConcreteNet = nil;
var allNetsT2 **ConcreteNet = nil;


// --------------------------------------------------------------------
// Function prototypes and local data structures
//
// --------------------------------------------------------------------

typedef struct FM_cell {
    int loc;
    int gain;
    ConcreteCell *cell;
    struct FM_cell *next, *prev;
    bool locked;
}
 FM_cell;

// void FM_updateGains(ConcreteNet *net, int partition, int inc, 
//                     FM_cell target [], FM_cell *bin [], 
//                     int count_1 [], int count_2 []);



// --------------------------------------------------------------------
// initPartitioning()
//
/// \brief Initializes data structures necessary for partitioning.
//
/// Creates a valid g_place_rootPartition.
///
// --------------------------------------------------------------------
func initPartitioning() {
  var i int32;
  var area float32;

  // create root partition
  g_place_numPartitions = 1;
  if g_place_rootPartition != nil { free(g_place_rootPartition); }
  g_place_rootPartition = malloc(sizeof(Partition));
  g_place_rootPartition.m_level = 0;
  g_place_rootPartition.m_area = 0;
  g_place_rootPartition.m_bounds = g_place_coreBounds;
  g_place_rootPartition.m_vertical = false;
  g_place_rootPartition.m_done = false;
  g_place_rootPartition.m_leaf = true;
      
  // add all of the cells to this partition
  g_place_rootPartition.m_members = malloc(sizeof(ConcreteCell*)*g_place_numCells);
  g_place_rootPartition.m_numMembers = 0;
  for i=0; i<g_place_numCells; i++ { 
    if g_place_concreteCells[i] {
      if !g_place_concreteCells[i].m_fixed {
        area = getCellArea(g_place_concreteCells[i]);
        g_place_rootPartition.m_members[g_place_rootPartition.m_numMembers++] =
          g_place_concreteCells[i];
        g_place_rootPartition.m_area += area;
      }
    }
  }
}


// --------------------------------------------------------------------
// presortNets()
//
/// \brief Sorts nets by corner positions.
//
/// Allocates allNetsX2 structures.
///
// --------------------------------------------------------------------
func presortNets() {
  allNetsL2 = (ConcreteNet**)realloc(allNetsL2, sizeof(ConcreteNet*)*g_place_numNets);
  allNetsR2 = (ConcreteNet**)realloc(allNetsR2, sizeof(ConcreteNet*)*g_place_numNets);
  allNetsB2 = (ConcreteNet**)realloc(allNetsB2, sizeof(ConcreteNet*)*g_place_numNets);
  allNetsT2 = (ConcreteNet**)realloc(allNetsT2, sizeof(ConcreteNet*)*g_place_numNets);
  memcpy(allNetsL2, (size_t)g_place_concreteNets, sizeof(ConcreteNet*)*g_place_numNets);
  memcpy(allNetsR2, (size_t)g_place_concreteNets, sizeof(ConcreteNet*)*g_place_numNets);
  memcpy(allNetsB2, (size_t)g_place_concreteNets, sizeof(ConcreteNet*)*g_place_numNets);
  memcpy(allNetsT2, (size_t)g_place_concreteNets, sizeof(ConcreteNet*)*g_place_numNets);
  qsort(allNetsL2, (size_t)g_place_numNets, sizeof(ConcreteNet*), netSortByL);
  qsort(allNetsR2, (size_t)g_place_numNets, sizeof(ConcreteNet*), netSortByR);
  qsort(allNetsB2, (size_t)g_place_numNets, sizeof(ConcreteNet*), netSortByB);
  qsort(allNetsT2, (size_t)g_place_numNets, sizeof(ConcreteNet*), netSortByT);
}

// --------------------------------------------------------------------
// refinePartitions()
//
/// \brief Splits large leaf partitions.
//
// --------------------------------------------------------------------
func refinePartitions() bool {

  return refinePartition(g_place_rootPartition);
}


// --------------------------------------------------------------------
// reallocPartitions()
//
/// \brief Reallocates the partitions based on placement information.
//
// --------------------------------------------------------------------
func reallocPartitions() {

  reallocPartition(g_place_rootPartition);
}


// --------------------------------------------------------------------
// refinePartition()
//
/// \brief Splits any large leaves within a partition.
//
// --------------------------------------------------------------------
func refinePartition(p *Partition) bool {
  var degenerate bool  = false;
  var nonzeroCount int32;
  var i int32;

  assert((p != nil), "p");

  // is this partition completed?
  if p.m_done { return true; }

  // is this partition a non-leaf node?
  if !p.m_leaf {
    p.m_done = refinePartition(p.m_sub1);
    p.m_done &= refinePartition(p.m_sub2);
    return p.m_done;
  }
  
  // leaf...
  // create two new subpartitions
  g_place_numPartitions++;
  p.m_sub1 = malloc(sizeof(Partition));
  p.m_sub1.m_level = p.m_level+1;
  p.m_sub1.m_leaf = true;
  p.m_sub1.m_done = false;
  p.m_sub1.m_area = 0;
  p.m_sub1.m_vertical = !p.m_vertical;
  p.m_sub1.m_numMembers = 0;
  p.m_sub1.m_members = nil;
  p.m_sub2 = malloc(sizeof(Partition));
  p.m_sub2.m_level = p.m_level+1;
  p.m_sub2.m_leaf = true;
  p.m_sub2.m_done = false;
  p.m_sub2.m_area = 0;
  p.m_sub2.m_vertical = !p.m_vertical;
  p.m_sub2.m_numMembers = 0;
  p.m_sub2.m_members = nil;
  p.m_leaf = false;

  // --- INITIAL PARTITION

  if PARTITION_AREA_ONLY {
    partitionEqualArea(p);
  } else { 
    partitionScanlineMincut(p); }

  resizePartition(p);

  // --- PARTITION IMPROVEMENT

  if p.m_level < REPARTITION_LEVEL_DEPTH {
    if REPARTITION_FM {
      repartitionFM(p);
    } else if REPARTITION_HMETIS {
      repartitionHMetis(p); }
  }
    
  resizePartition(p);
  
  // fix imbalances due to zero-area cells
  for i=0; i<p.m_sub1.m_numMembers; i++ {
    if p.m_sub1.m_members[i] { 
      if getCellArea(p.m_sub1.m_members[i]) > 0 {
        nonzeroCount++;
      } }
  }
  
  // is this leaf now done?
  if nonzeroCount <= LARGEST_FINAL_SIZE {
      p.m_sub1.m_done = true; }
  if nonzeroCount == 0 {
      degenerate = true; }

  nonzeroCount = 0;
  for i=0; i<p.m_sub2.m_numMembers; i++ {
    if p.m_sub2.m_members[i] {
      if getCellArea(p.m_sub2.m_members[i]) > 0 {
        nonzeroCount++;
      } }
  }

  // is this leaf now done?
  if nonzeroCount <= LARGEST_FINAL_SIZE {
      p.m_sub2.m_done = true; }
  if nonzeroCount == 0 {
      degenerate = true; }

  // have we found a degenerate partitioning?
  if degenerate {
    fmt.Printf("QPART-35 : WARNING: degenerate partition generated\n");
    partitionEqualArea(p);
    resizePartition(p);
    p.m_sub1.m_done = true;
    p.m_sub2.m_done = true;
  }
  
  // is this parent now finished?
  if p.m_sub1.m_done && p.m_sub2.m_done { p.m_done = true; }
  
  return p.m_done;
}


// --------------------------------------------------------------------
// repartitionHMetis()
//
/// \brief Repartitions the two subpartitions using the hMetis min-cut library.
///
/// The number of cut nets between the two partitions will be minimized.
//
// --------------------------------------------------------------------
func repartitionHMetis(parent *Partition) {
#if defined(NO_HMETIS)
  printf("QPAR_02 : \t\tERROR: hMetis not available.  Ignoring.\n");
#else

  int n,c,t, i;
  float area;
  int *edgeConnections = NULL;
  int *partitionAssignment = (int *)calloc(g_place_numCells, sizeof(int));
  int *vertexWeights = (int *)calloc(g_place_numCells, sizeof(int));
  int *edgeDegree = (int *)malloc(sizeof(int)*(g_place_numNets+1));
  int numConnections = 0;
  int numEdges = 0;
  float initial_cut;
  int targets = 0;
  ConcreteCell *cell = NULL;
  int options[9];
  int afterCuts = 0;

  assert(parent);
  assert(parent->m_sub1);
  assert(parent->m_sub2);

  printf("QPAR-02 : \t\trepartitioning with hMetis\n");

  // count edges
  edgeDegree[0] = 0;
  for(n=0; n<g_place_numNets; n++) if (g_place_concreteNets[n])
    if (g_place_concreteNets[n]->m_numTerms > 1) {
      numConnections += g_place_concreteNets[n]->m_numTerms;
      edgeDegree[++numEdges] = numConnections;
    }
  
  if (parent->m_vertical) {
    // vertical
    initial_cut = parent->m_sub2->m_bounds.x;
    
    // initialize all cells
    for(c=0; c<g_place_numCells; c++) if (g_place_concreteCells[c]) {
      if (g_place_concreteCells[c]->m_x < initial_cut)
        partitionAssignment[c] = 0;
      else
        partitionAssignment[c] = 1;
    }
  
    // initialize cells in partition 1
    for(t=0; t<parent->m_sub1->m_numMembers; t++) if (parent->m_sub1->m_members[t]) {
      cell = parent->m_sub1->m_members[t];
      vertexWeights[cell->m_id] = getCellArea(cell);
      // pay attention to cells that are close to the cut
      if (abs(cell->m_x-initial_cut) < parent->m_bounds.w*REPARTITION_TARGET_FRACTION) {
        targets++;
        partitionAssignment[cell->m_id] = -1;
      }
    }
    
    // initialize cells in partition 2
    for(t=0; t<parent->m_sub2->m_numMembers; t++) if (parent->m_sub2->m_members[t]) {
      cell = parent->m_sub2->m_members[t];
      vertexWeights[cell->m_id] = getCellArea(cell);
      // pay attention to cells that are close to the cut
      if (abs(cell->m_x-initial_cut) < parent->m_bounds.w*REPARTITION_TARGET_FRACTION) {
        targets++;
        partitionAssignment[cell->m_id] = -1;
      }        
    }
    
  } else {
    // horizontal
    initial_cut = parent->m_sub2->m_bounds.y;
    
    // initialize all cells
    for(c=0; c<g_place_numCells; c++) if (g_place_concreteCells[c]) {
      if (g_place_concreteCells[c]->m_y < initial_cut)
        partitionAssignment[c] = 0;
      else
        partitionAssignment[c] = 1;
    }
    
    // initialize cells in partition 1
    for(t=0; t<parent->m_sub1->m_numMembers; t++) if (parent->m_sub1->m_members[t]) {
      cell = parent->m_sub1->m_members[t];
      vertexWeights[cell->m_id] = getCellArea(cell);
      // pay attention to cells that are close to the cut
      if (abs(cell->m_y-initial_cut) < parent->m_bounds.h*REPARTITION_TARGET_FRACTION) {
        targets++;
        partitionAssignment[cell->m_id] = -1;
      }
    }
    
    // initialize cells in partition 2
    for(t=0; t<parent->m_sub2->m_numMembers; t++) if (parent->m_sub2->m_members[t]) {
      cell = parent->m_sub2->m_members[t];
      vertexWeights[cell->m_id] = getCellArea(cell);
      // pay attention to cells that are close to the cut
      if (abs(cell->m_y-initial_cut) < parent->m_bounds.h*REPARTITION_TARGET_FRACTION) {
        targets++;
        partitionAssignment[cell->m_id] = -1;
      }        
    }
  }

  options[0] = 1;  // any non-default values?
  options[1] = 3; // num bisections
  options[2] = 1;  // grouping scheme
  options[3] = 1;  // refinement scheme
  options[4] = 1;  // cycle refinement scheme
  options[5] = 0;  // reconstruction scheme
  options[6] = 0;  // fixed assignments?
  options[7] = 12261980; // random seed
  options[8] = 0;  // debugging level

  edgeConnections = (int *)malloc(sizeof(int)*numConnections);

  i = 0;
  for(n=0; n<g_place_numNets; n++) if (g_place_concreteNets[n]) {
    if (g_place_concreteNets[n]->m_numTerms > 1)
      for(t=0; t<g_place_concreteNets[n]->m_numTerms; t++)
        edgeConnections[i++] = g_place_concreteNets[n]->m_terms[t]->m_id;
  }

  HMETIS_PartRecursive(g_place_numCells, numEdges, vertexWeights,
               edgeDegree, edgeConnections, NULL,
               2, (int)(100*MAX_PARTITION_NONSYMMETRY),
               options, partitionAssignment, &afterCuts);
    
  /*
  printf("HMET-20 : \t\t\tbalance before %d / %d ... ", parent->m_sub1->m_numMembers,
         parent->m_sub2->m_numMembers);
  */

  // reassign members to subpartitions
  parent->m_sub1->m_numMembers = 0;
  parent->m_sub1->m_area = 0;
  parent->m_sub2->m_numMembers = 0;
  parent->m_sub2->m_area = 0;
  parent->m_sub1->m_members = (ConcreteCell**)realloc(parent->m_sub1->m_members, 
       sizeof(ConcreteCell*)*parent->m_numMembers); 
  parent->m_sub2->m_members = (ConcreteCell**)realloc(parent->m_sub2->m_members, 
       sizeof(ConcreteCell*)*parent->m_numMembers); 
 
  for(t=0; t<parent->m_numMembers; t++) if (parent->m_members[t]) {
    cell = parent->m_members[t];
    area = getCellArea(cell);
    if (partitionAssignment[cell->m_id] == 0) {
      parent->m_sub1->m_members[parent->m_sub1->m_numMembers++] = cell;
      parent->m_sub1->m_area += area;
    }
    else {
      parent->m_sub2->m_members[parent->m_sub2->m_numMembers++] = cell;
      parent->m_sub2->m_area += area;
    }
  }
  /*
  printf("after %d / %d\n", parent->m_sub1->m_numMembers,
         parent->m_sub2->m_numMembers);
  */

  // cout << "HMET-21 : \t\t\tloc: " << initial_cut <<  " targetting: " << targets*100/parent->m_members.length() << "%" << endl;
  // cout << "HMET-22 : \t\t\tstarting cuts= " << beforeCuts << " final cuts= " << afterCuts << endl;

  free(edgeConnections);
  free(vertexWeights);
  free(edgeDegree);
  free(partitionAssignment);
#endif
}


// --------------------------------------------------------------------
// repartitionFM()
//
/// \brief Fiduccia-Matheyses mincut partitioning algorithm.
//
/// UNIMPLEMENTED (well, un-C-ified)
//
// --------------------------------------------------------------------
func repartitionFM(parent *Partition) {
     
                                                                          

                                                              
                                         
                                                            
                                         
                                                            

                                             
                                                                
                               
                        
                                                     

                     
                                          
                       
                    
                 
                                           
                                                                        
                     

                     
                                  

                           
               

                                             

                           
                                                                                             
                                 
                                        
                                 
            
                                 
                                   
                                 
     

                                      
                                                                                              
                                 
                                                           
                                                                                              
                  
                                
         
     

                                      
                                                                                              
                                 
                                                           
                                                                                              
                  
                                
                 
     

                                                                                
                                                                                            
                                                                                             
                                                              
                                    
                                                                   
                                    
                                                           
                                    
            
                                    
                                                                                  
     
    
            
                 

                                             

                           
                                                                                             
                                 
                                        
                                 
            
                                 
                                   
                                 
     

                                      
                                                                                              
                                 
                                                           
                                                                                              
                  
                                
         
     

                                      
                                                                                              
                                 
                                                           
                                                                                              
                  
                                
                 
     

                                                                                
                                                                                            
                                                                                             
                                                              
                                    
                                                                   
                                    
                                                           
                                    
            
                                    
                                                                                  
     
     

                               
                                                       
                             
                                     
                                     

                                                 
                                                                                              
                                    
                                                  
                                                                             
                                                                             
                                                                             
                                                                             
         

                                                                            

                          
                                                              
                                          

                                                      
                                       
                            
                              
                                         
                                   
     

                     
                               
                      
                 
                                                  
           
                      

                                                 

                                 
                       
                        
                                  
                                

                                                                   
                                          

                               
                        
                         

                              
                                     
                                     
                               

                     
     
                                                      
                                                     
    
                          
                                                         

                                   
              

                                 
                                        

                                     
                                               
                                                                                                          
                                                                                                           

                                                                                         
         
                         
              

                    
                               
                                                                                                                                                                       

                                             
                                
                         
                                                           
                                                           

                                                                    
                                                                              
                            
            
                                                                 
                     
                                    
            
                                    
                                    

                                                                                                                                                                                                    

                                                                                      
                                                                                                                          
                                                                                                                          
            
                                                         
         

                
                         
                                                           
                                                           

                                                         
                                                                              
                            
            
                                                                 
                     
                                    
                                    
                                    

                                                                                                                                                                                                    

                                                                                                                          
                                                                                                                          
        
                                                         
         
         

                                                   

                                 

  
                                            
                                       
                               
                    
                                         
                                       
                               
                    
            
                                    
                                               
  

                                    
                                     
                               
                                            
                               
                                            

  
                                            
                                       
                               
                    
                                         
                                       
                               
                    
            
                                    
                                               
  

                          
                               
                        
                               
                         
        
                                                         

                          
                              
                                                             
     

                                                                                     
     

                                        
                                                                                      
                                                    
                                      
                               
                                      
                               
                                                                                      
                                                                              
                                                 
                                                   
     
          
                                                 
                                                   
     
     
                                                                  
                                                     


                                                                                                                            
                                                                                                    
      
}

// ----- FM_updateGains()
//   moves a cell between bins
     
                                                              
                                                        
                                                     

                                                                           
                                                            
                               
    
                                 
                      

                                             
                                                                                        
                                    
                                              
                                                                     
                                                                     
                                                                     
                                                                     
         

                           
                                        
                                                        
                                    
                                     
                               
                                            
                               
                                            
                                  
                                                         
                          
                                     
                              
                                     
                               
     
     
    
 
      


// --------------------------------------------------------------------
// partitionEqualArea()
//
/// \brief Splits a partition into two halves of equal area.
//
// --------------------------------------------------------------------
func partitionEqualArea(parent *Partition) {
  var halfArea, area float32;
  var i int32;
  
  // which way to sort?
  if parent.m_vertical {
    // sort by X position
    qsort(parent.m_members, (size_t)parent.m_numMembers, sizeof(ConcreteCell*), cellSortByX);
  } else {
    // sort by Y position
    qsort(parent.m_members, (size_t)parent.m_numMembers, sizeof(ConcreteCell*), cellSortByY); }

  // split the list
  halfArea = parent.m_area*0.5;
  parent.m_sub1.m_area = 0.0;
  parent.m_sub1.m_numMembers = 0;
  parent.m_sub1.m_members = (ConcreteCell**)realloc(parent.m_sub1.m_members, 
                                    sizeof(ConcreteCell*)*parent.m_numMembers);
  parent.m_sub2.m_area = 0.0;
  parent.m_sub2.m_numMembers = 0;
  parent.m_sub2.m_members = (ConcreteCell**)realloc(parent.m_sub2.m_members, 
                                    sizeof(ConcreteCell*)*parent.m_numMembers);

  for ; parent.m_sub1.m_area < halfArea; i++ { 
    if parent.m_members[i] {
      area = getCellArea(parent.m_members[i]);
      parent.m_sub1.m_members[parent.m_sub1.m_numMembers++] = parent.m_members[i];
      parent.m_sub1.m_area += area;
  }
  }
  for ; i<parent.m_numMembers; i++ { 
    if parent.m_members[i] {
      area = getCellArea(parent.m_members[i]);
      parent.m_sub2.m_members[parent.m_sub2.m_numMembers++] = parent.m_members[i];
      parent.m_sub2.m_area += area;
    }
  }
  
}


// --------------------------------------------------------------------
// partitionScanlineMincut()
//
/// \brief Scans the cells within a partition from left to right and chooses the min-cut.
//
// --------------------------------------------------------------------
func partitionScanlineMincut(parent *Partition) {
     
                       
                             
                                        
                                                          
                                                                      
                                     

                                                                             
                        
                                                                      
                         
               
                                                             
                       
                 
                              
                               
       
     
   

                         
                                    
                           
                                                                        
                                     
                                         
                                                 
             
                                                 
          
                                 
                                                                                            
                                 
                     
     
           
       
                                                                                             
                                 
                     
     
           
       
                                        
                                
                              
       
                        
     
   
        
                                    
                           
                                                                        
                                     
                                         
                                                 
             
                                                 
          
                                 
                                                                                           
                                 
                     
     
           
       
                                                                                              
                                 
                     
     
           
       
                                        
                                
                              
       
                        
     
   
                                 
                                      
   
                                                                   
                                    
                             
                                        
                                              
                                               
   
                                    
                             
                    
                                              
                                               
   
      
}


// --------------------------------------------------------------------
// reallocPartition()
//
/// \brief Reallocates a partition and all of its children.
//
// --------------------------------------------------------------------
func reallocPartition(p *Partition) {

  if p.m_leaf {
    return;
  }

  // --- INITIAL PARTITION

  if PARTITION_AREA_ONLY {
    partitionEqualArea(p);
  } else { 
    partitionScanlineMincut(p); }

  resizePartition(p);

  // --- PARTITION IMPROVEMENT
  if p.m_level < REPARTITION_LEVEL_DEPTH {
    if REPARTITION_HMETIS {
      repartitionHMetis(p); }
    
    resizePartition(p);
  }

  reallocPartition(p.m_sub1);
  reallocPartition(p.m_sub2);
}


// --------------------------------------------------------------------
// resizePartition()
//
/// \brief Recomputes the bounding boxes of the child partitions based on their relative areas.
//
// --------------------------------------------------------------------
func resizePartition(p *Partition) {
  // compute the new bounding box
  p.m_sub1.m_bounds.x = p.m_bounds.x;
  p.m_sub1.m_bounds.y = p.m_bounds.y;
  if p.m_vertical {
    p.m_sub1.m_bounds.w = p.m_bounds.w*(p.m_sub1.m_area/p.m_area);
    p.m_sub1.m_bounds.h = p.m_bounds.h;
    p.m_sub2.m_bounds.x = p.m_bounds.x + p.m_sub1.m_bounds.w;
    p.m_sub2.m_bounds.w = p.m_bounds.w*(p.m_sub2.m_area/p.m_area);
    p.m_sub2.m_bounds.y = p.m_bounds.y;
    p.m_sub2.m_bounds.h = p.m_bounds.h;
  } else {
    p.m_sub1.m_bounds.h = p.m_bounds.h*(p.m_sub1.m_area/p.m_area);
    p.m_sub1.m_bounds.w = p.m_bounds.w;
    p.m_sub2.m_bounds.y = p.m_bounds.y + p.m_sub1.m_bounds.h;
    p.m_sub2.m_bounds.h = p.m_bounds.h*(p.m_sub2.m_area/p.m_area);
    p.m_sub2.m_bounds.x = p.m_bounds.x;
    p.m_sub2.m_bounds.w = p.m_bounds.w;
  }
}


// --------------------------------------------------------------------
// incrementalSubpartition()
//
/// \brief Adds new cells to an existing partition.  Partition sizes/locations are unchanged.
///
/// The function recurses, adding new cells to appropriate subpartitions.
//
// --------------------------------------------------------------------
func incrementalSubpartition(p *Partition, newCells []*ConcreteCell, numNewCells int32) {
  var c int32;
  var newCells1 **ConcreteCell = (ConcreteCell **)malloc(sizeof(ConcreteCell*)*numNewCells);
  var newCells2 **ConcreteCell = (ConcreteCell **)malloc(sizeof(ConcreteCell*)*numNewCells);
  var numNewCells1, numNewCells2 int32;
  var cut_loc float32;

  assert((p != nil), "p");

  // add new cells to partition list
  p.m_members = (ConcreteCell**)realloc(p.m_members, 
         sizeof(ConcreteCell*)*(p.m_numMembers+numNewCells));
  memcpy(&(p.m_members[p.m_numMembers]), newCells, sizeof(ConcreteCell*)*numNewCells);
  p.m_numMembers += numNewCells;

  // if is a leaf partition, finished
  if p.m_leaf { return; }

  // split new cells into sub-partitions based on location
  if p.m_vertical {
    cut_loc = p.m_sub2.m_bounds.x;
    for c=0; c<numNewCells; c++ {
      if newCells[c].m_x < cut_loc {
        newCells1[numNewCells1++] = newCells[c];
      } else {
        newCells2[numNewCells2++] = newCells[c]; }
    }
  } else {
    cut_loc = p.m_sub2.m_bounds.y;
    for c=0; c<numNewCells; c++ {
      if newCells[c].m_y < cut_loc {
        newCells1[numNewCells1++] = newCells[c];
      } else {
        newCells2[numNewCells2++] = newCells[c]; }
    }    
  }

  if numNewCells1 > 0 { incrementalSubpartition(p.m_sub1, newCells1, numNewCells1); }
  if numNewCells2 > 0 { incrementalSubpartition(p.m_sub2, newCells2, numNewCells2); }

  free(newCells1);
  free(newCells2);
}


// --------------------------------------------------------------------
// incrementalPartition()
//
/// \brief Adds new cells to an existing partition.  Partition sizes/locations are unchanged.
///
/// The function recurses, adding new cells to appropriate subpartitions.
//
// --------------------------------------------------------------------
func incrementalPartition() {
  var c, c2 int32;
  var numNewCells int32;
  var allCells **ConcreteCell = (ConcreteCell **)malloc(sizeof(ConcreteCell*)*g_place_numCells);
  var newCells **ConcreteCell = (ConcreteCell **)malloc(sizeof(ConcreteCell*)*g_place_numCells);

  assert((g_place_rootPartition != nil), "g_place_rootPartition");

  // update cell list of root partition
  memcpy(allCells, (size_t)g_place_concreteCells, sizeof(ConcreteCell*)*g_place_numCells);
  qsort(allCells, (size_t)g_place_numCells, sizeof(ConcreteCell*), cellSortByID);
  qsort(g_place_rootPartition.m_members, (size_t)g_place_rootPartition.m_numMembers,
          sizeof(ConcreteCell*), cellSortByID);

  // scan sorted lists and collect cells not in partitions
  for !allCells[c++] {; }
  for !g_place_rootPartition.m_members[c2++] {; }

  for ; c<g_place_numCells; c, c2 = c+1, c2+1 {
    for c2 < g_place_rootPartition.m_numMembers &&
          allCells[c].m_id > g_place_rootPartition.m_members[c2].m_id { c2++; }
    for c < g_place_numCells && 
          (c2 >= g_place_rootPartition.m_numMembers ||
           allCells[c].m_id < g_place_rootPartition.m_members[c2].m_id) {
      // a new cell!
      newCells[numNewCells++] = allCells[c];
      c++;
    }
  }
  
  fmt.Printf("QPRT-50 : \tincremental partitioning with %d new cells\n", numNewCells);
  if numNewCells>0 { incrementalSubpartition(g_place_rootPartition, newCells, numNewCells); }

  free(allCells);
  free(newCells);
}
//ABC_NAMESPACE_IMPL_END

