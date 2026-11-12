package mvc

/**CFile****************************************************************

  FileName    [mvcProc.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Various boolean procedures working with covers.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcOpBool.c,v 1.4 2003/04/16 01:55:37 alanmi Exp $]

***********************************************************************/

//#include "mvc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverBooleanOr(pCover1 *Mvc_Cover_t, pCover2 *Mvc_Cover_t) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
    var pCube, pCubeCopy *Mvc_Cube_t;
    // make sure the covers are compatible
    assert(pCover1.nBits == pCover2.nBits, "pCover1.nBits == pCover2.nBits");
    // clone the cover
    pCover = Mvc_CoverClone( pCover1 );
    // create the cubes by making pair-wise products
    // of cubes in pCover1 and pCover2
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover1).lCubes.pHead;          pCube;          pCube = pCube.pNext  {
        pCubeCopy = Mvc_CubeDup( pCover, pCube );
        // inlining of macro Mvc_CoverAddCubeTail
        // inlining of macro Mvc_ListAddCubeTail
        // inlining of macro Mvc_CubeSetNext
        {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
    }
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover2).lCubes.pHead;          pCube;          pCube = pCube.pNext  {
        pCubeCopy = Mvc_CubeDup( pCover, pCube );
        // inlining of macro Mvc_CoverAddCubeTail
        // inlining of macro Mvc_ListAddCubeTail
        // inlining of macro Mvc_CubeSetNext
        {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
    }
    return pCover;
}

     

                                                                        

                

                
               
                

                

                                                                        
                                                                                                 
 
                         
                                               
                                          
                                               
                      
                                       
                                                    
                                      
                                           
     
                                               
         
                                                           
             
                                                    
                                                            
                                                          
             
         
                                                               
                               
                                                   
                                       
     
                  
 

                                                                        

                                                      

                
               
                

                

                                                                        
                                                                                         
 
                         

                                                   
                                         
     
                           
                     
                         
                                
                 
     
                            

                                                   
                                         
     
                           
                     
                         
                                
                 
     
                            

             
 

      

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

