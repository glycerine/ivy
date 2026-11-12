package mvc

import "fmt"

/**CFile****************************************************************

  FileName    [mvcPrint.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Printing cubes and covers.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcPrint.c,v 1.6 2003/04/09 18:02:06 alanmi Exp $]

***********************************************************************/

//#include "mvc.h"
//#include "vm.h"
//#include "vmInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Mvc_CubePrintBinary( Mvc_Cover_t * pCover, Mvc_Cube_t * pCube );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverPrint(pCover *Mvc_Cover_t) {
    var pCube *Mvc_Cube_t;
    var i int32;
    // print general statistics
    fmt.Printf("The cover contains %d cubes (%d bits and %d words)\n", pCover.lCubes.nItems, pCover.nBits, pCover.nWords);
    // iterate through the cubes
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        Mvc_CubePrint( pCover, pCube );
    }

    if pCover.pLits {
        for  i = 0; i < pCover.nBits; i++  {
            fmt.Printf(" %d", pCover.pLits[i]);
        }
        fmt.Printf("\n"); 
    }
    fmt.Printf("End of cover printout\n"); 
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CubePrint(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
    var iBit, Value int32;
    // iterate through the literals
//    printf( "Size = %2d   ", Mvc_CubeReadSize(pCube) );
    // inlining of macro Mvc_CubeForEachBit
    // inlining of macro Mvc_CubeBitValue
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    for  iBit = 0;          iBit < pCover.nBits && ((Value = (((pCube).pData[((iBit) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iBit) &  BITS_PER_WORD_MINUS)))) > 0)), 1);          iBit++  {
        fmt.Printf("%c", '0' + Value);
    }
    fmt.Printf("\n");
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverPrintBinary(pCover *Mvc_Cover_t) {
    var pCube *Mvc_Cube_t;
    var i int32;
    // print general statistics
    fmt.Printf("The cover contains %d cubes (%d bits and %d words)\n", pCover.lCubes.nItems, pCover.nBits, pCover.nWords);
    // iterate through the cubes
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        Mvc_CubePrintBinary( pCover, pCube );
    }

    if pCover.pLits {
        for  i = 0; i < pCover.nBits; i++  {
            fmt.Printf(" %d", pCover.pLits[i]);
        }
        fmt.Printf("\n"); 
    }
    fmt.Printf("End of cover printout\n"); 
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CubePrintBinary(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
    var iVar, Value int32;
    // iterate through the literals
//    printf( "Size = %2d   ", Mvc_CubeReadSize(pCube) );
    // inlining of macro Mvc_CubeForEachVarValue
    // inlining of macro Mvc_CubeVarValue
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    for  iVar = 0;          iVar < pCover.nBits/2 && (Value = (((pCube).pData[((2*(iVar)) >> BITS_PER_WORD_LOG)] >> (((2*(iVar)) &  BITS_PER_WORD_MINUS))) & ((Mvc_CubeWord_t)3)));          iVar++  {
        assert(Value != 0, "Value != 0");
        if Value == 3 {
            fmt.Printf("-");
        } else if Value == 1 {
            fmt.Printf("0");
        } else { 
            fmt.Printf("1"); }
    }
    fmt.Printf("\n");
}

     

                                                                        

                

                
               
                

                

                                                                        
                                                                 
 
                       
          
                               
                                                                    
                                                               
                                
                                         
                                                

                        
     
                                             
                                              
                        
     
                                         
 


                                                                        

                

                
               
                

                

                                                                        
                                                                                    
 
                   
                                   
                                                       
             
                                                          
     
                                                       
         
                          
                   
         
                                              
                                                                        
            
                          
     
                   
 

      

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

