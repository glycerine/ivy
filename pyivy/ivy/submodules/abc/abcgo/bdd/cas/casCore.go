package cas

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [casCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CASCADE: Decomposition of shared BDDs into a LUT cascade.]

  Synopsis    [Entrance into the implementation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Spring 2002.]

  Revision    [$Id: casCore.c,v 1.0 2002/01/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>

//#include "base/main/main.h"
//#include "base/cmd/cmd.h"
//#include "bdd/extrab/extraBdd.h"
//#include "cas.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                      static functions                            ///
////////////////////////////////////////////////////////////////////////

// DdNode * GetSingleOutputFunction( DdManager * dd, DdNode ** pbOuts, int nOuts, DdNode ** pbVarsEnc, int nVarsEnc, int fVerbose );

// DdNode * GetSingleOutputFunctionRemapped( DdManager * dd, DdNode ** pOutputs, int nOuts, DdNode ** pbVarsEnc, int nVarsEnc );

// DdNode * GetSingleOutputFunctionRemappedNewDD( DdManager * dd, DdNode ** pOutputs, int nOuts, DdManager ** DdNew );


// extern int CreateDecomposedNetwork( DdManager * dd, DdNode * aFunc, char ** pNames, int nNames, char * FileName, int nLutSize, int fCheck, int fVerbose );


// void WriteSingleOutputFunctionBlif( DdManager * dd, DdNode * aFunc, char ** pNames, int nNames, char * FileName );


// DdNode * Cudd_bddTransferPermute( DdManager * ddSource, DdManager * ddDestination, DdNode * f, int * Permute );


////////////////////////////////////////////////////////////////////////
///                      static varibles                             ///
////////////////////////////////////////////////////////////////////////

//static FILE * pTable = NULL;
//static long s_RemappingTime = 0;

////////////////////////////////////////////////////////////////////////
///                      debugging macros                            ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define PRD(p)       printf( "\nDECOMPOSITION TREE:\n\n" ); PrintDecEntry( (p), 0 ) 

// Original macro:
// #define PRB_(f)       printf( #f " = " ); Cudd_bddPrint(dd,f); printf( "\n" )

// Original macro:
// #define PRK(f,n)     Cudd_PrintKMap(stdout,dd,(f),Cudd_Not(f),(n),NULL,0); printf( "K-map for function" #f "\n\n" )

// Original macro:
// #define PRK2(f,g,n)  Cudd_PrintKMap(stdout,dd,(f),(g),(n),NULL,0); printf( "K-map for function <" #f ", " #g ">\n\n" ) 



////////////////////////////////////////////////////////////////////////
///                     EXTERNAL FUNCTIONS                           ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CascadeExperiment(pFileGeneric string, dd *DdManager, pOutputs **DdNode, nInputs int32, nOutputs int32, nLutSize int32, fCheck int32, fVerbose int32) int32 {
    var i int32;
    var nVars int32  = nInputs;
    var nOuts int32  = nOutputs;
    var clk1 abctime;

    var nVarsEnc int32;              // the number of additional variables to encode outputs
    var pbVarsEnc [MAXOUTPUTS]*DdNode; // the BDDs of the encoding vars

    var nNames int32;               // the total number of all inputs
    var pNames [MAXINPUTS]string;     // the temporary storage for the input (and output encoding) names

    var aFunc *DdNode;                 // the encoded 0-1 BDD containing all the outputs

    var FileNameIni [100]byte;
    var FileNameFin [100]byte;
    var Buffer string // was [100]byte in C


    
//pTable = fopen( "stats.txt", "a+" );
//fprintf( pTable, "%s ", pFileGeneric );
//fprintf( pTable, "%d ", nVars );
//fprintf( pTable, "%d ", nOuts );


    // assign the file names
    strcpy( FileNameIni, pFileGeneric );
    strcat( FileNameIni, "_ENC.blif" );

    strcpy( FileNameFin, pFileGeneric );
    strcat( FileNameFin, "_LUT.blif" );


    // create the variables to encode the outputs
    nVarsEnc = Abc_Base2Log( nOuts );
    for  i = 0; i < nVarsEnc; i++  {
        pbVarsEnc[i] = Cudd_bddNewVarAtLevel( dd, i );
    }


    // store the input names
    nNames  = nVars + nVarsEnc;
    for  i = 0; i < nVars; i++  {
//      pNames[i] = Extra_UtilStrsav( pFunc->pInputNames[i] );
        Buffer = fmt.Sprintf("pi%03d", i);
        pNames[i] = Extra_UtilStrsav( Buffer );
    }
    // set the encoding variable name
    for  ; i < nNames; i++  {       
        Buffer = fmt.Sprintf("OutEnc_%02d", i-nVars);
        pNames[i] = Extra_UtilStrsav( Buffer );
    }


    // print the variable order
//  printf( "\n" );
//  printf( "Variable order is: " );
//  for ( i = 0; i < dd->size; i++ )
//      printf( " %d", dd->invperm[i] );
//  printf( "\n" );

    // derive the single-output function
    clk1 = Abc_Clock();
    aFunc = GetSingleOutputFunction( dd, pOutputs, nOuts, pbVarsEnc, nVarsEnc, fVerbose );  Cudd_Ref( aFunc );
//  aFunc = GetSingleOutputFunctionRemapped( dd, pOutputs, nOuts, pbVarsEnc, nVarsEnc );  Cudd_Ref( aFunc );
//  if ( fVerbose )
//  printf( "Single-output function computation time = %.2f sec\n", (float)(Abc_Clock() - clk1)/(float)(CLOCKS_PER_SEC) );
 
//fprintf( pTable, "%d ", Cudd_SharingSize( pOutputs, nOutputs ) );
//fprintf( pTable, "%d ", Extra_ProfileWidthSharingMax(dd, pOutputs, nOutputs) );

    // dispose of the multiple-output function
//  Extra_Dissolve( pFunc );
 
    // reorder the single output function
//    if ( fVerbose )
//  printf( "Reordering variables...\n");
    clk1 = Abc_Clock();
//  if ( fVerbose )
//  printf( "Node count before = %6d\n", Cudd_DagSize( aFunc ) );
//  Cudd_ReduceHeap(dd, CUDD_REORDER_SIFT,1);
    Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT,1);
    Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT,1);
//  Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT,1);
//  Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT,1);
//  Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT,1);
//  Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT,1);
    if fVerbose != 0 {
    fmt.Printf("MTBDD reordered = %6d nodes\n", Cudd_DagSize( aFunc )); }
    if fVerbose != 0 {
    fmt.Printf("Variable reordering time = %.2f sec\n", float32((Abc_Clock() - clk1))/float32((int64(time.Second)))); }
//  printf( "\n" );
//  printf( "Variable order is: " );
//  for ( i = 0; i < dd->size; i++ )
//      printf( " %d", dd->invperm[i] );
//  printf( "\n" );
//fprintf( pTable, "%d ", Cudd_DagSize( aFunc ) );
//fprintf( pTable, "%d ", Extra_ProfileWidthMax(dd, aFunc) );

    // write the single-output function into BLIF for verification
    clk1 = Abc_Clock();
    if fCheck != 0 {
    WriteSingleOutputFunctionBlif( dd, aFunc, pNames, nNames, FileNameIni ); }
//    if ( fVerbose )
//  printf( "Single-output function writing time = %.2f sec\n", (float)(Abc_Clock() - clk1)/(float)(CLOCKS_PER_SEC) );

/*
    ///////////////////////////////////////////////////////////////////
    // verification of single output function
    clk1 = Abc_Clock();
    {
        BFunc g_Func;
        DdNode * aRes;

        g_Func.dd = dd;
        g_Func.FileInput = Extra_UtilStrsav(FileNameIni);

        if ( Extra_ReadFile( &g_Func ) == 0 )
        {
            printf( "\nSomething did not work out while reading the input file for verification\n");
            Extra_Dissolve( &g_Func );
            return;
        } 

        aRes = Cudd_BddToAdd( dd, g_Func.pOutputs[0] );  Cudd_Ref( aRes );

        if ( aRes != aFunc )
            printf( "\nVerification FAILED!\n");
        else
            printf( "\nVerification okay!\n");
        
        Cudd_RecursiveDeref( dd, aRes );

        // delocate
        Extra_Dissolve( &g_Func );
    }
    printf( "Preliminary verification time = %.2f sec\n", (float)(Abc_Clock() - clk1)/(float)(CLOCKS_PER_SEC) );
    ///////////////////////////////////////////////////////////////////
*/

    if !CreateDecomposedNetwork( dd, aFunc, pNames, nNames, FileNameFin, nLutSize, fCheck, fVerbose ) {
        return 0; }

/*
    ///////////////////////////////////////////////////////////////////
    // verification of the decomposed LUT network
    clk1 = Abc_Clock();
    {
        BFunc g_Func;
        DdNode * aRes;

        g_Func.dd = dd;
        g_Func.FileInput = Extra_UtilStrsav(FileNameFin);

        if ( Extra_ReadFile( &g_Func ) == 0 )
        {
            printf( "\nSomething did not work out while reading the input file for verification\n");
            Extra_Dissolve( &g_Func );
            return;
        } 

        aRes = Cudd_BddToAdd( dd, g_Func.pOutputs[0] );  Cudd_Ref( aRes );

        if ( aRes != aFunc )
            printf( "\nFinal verification FAILED!\n");
        else
            printf( "\nFinal verification okay!\n");
        
        Cudd_RecursiveDeref( dd, aRes );
 
        // delocate 
        Extra_Dissolve( &g_Func );
    }
    printf( "Final verification time = %.2f sec\n", (float)(Abc_Clock() - clk1)/(float)(CLOCKS_PER_SEC) );
    ///////////////////////////////////////////////////////////////////
*/
 
    // verify the results
    if fCheck != 0 {
        var Command [300]byte;
        sprintf( Command, "cec %s %s", FileNameIni, FileNameFin );
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
    }

    Cudd_RecursiveDeref( dd, aFunc );

    // release the names
    for  i = 0; i < nNames; i++  {
        ABC_FREE( pNames[i] );
    }


//fprintf( pTable, "\n" );
//fclose( pTable );

    return 1;
}

     

                                                                        

                

                
               
                

                

                                                                        
                                 
 
                       
                               
                                
                               
              

                                                                                           
                                                                    

                                                                    
                                                                                                      

                                                                                     

                          
                          
                     

                      

                                      
                                               
                                  
                                  


                            
                                              
                                       

                                              
                                       

                                                           
                       
                                                                                                           
                                                                                                           
                                                                                                                         
                                         

                                              
                            

                                         
                                                              
                       
                                                                
                                                
                                                     
                                                     
                                                                
                                                                                                           
                   

                                                  
                                                                


                                                       
                         
                                  
     
                                      
                                               
     



                                                                  
                       
                                                                               
                                                                                                                      


                                                                       
                                             
                       
     
                     
                      

                          
                                                         

                                             
         
                                                                                                    
                                      
                   
          

                                                                             

                            
                                                
            
                                              
        
                                           

                   
                                  
     
                                                                                                                
                                                                       


                                                                                      

  
                                                                       
                                                 
                       
     
                     
                      

                          
                                                         

                                             
         
                                                                                                    
                                      
                   
          

                                                                             

                            
                                                      
            
                                                    
        
                                           

                   
                                  
     
                                                                                                          
                                                                       
  


                                        

                        
                                  
                              


    
                                                                         
                                                    
                                          
                                                                                     
                       

                          
                   

  

      

////////////////////////////////////////////////////////////////////////
///                       SINGLE OUTPUT FUNCTION                     ///
////////////////////////////////////////////////////////////////////////

// the bit count for the first 256 integer numbers
//static unsigned char BitCount8[256] = {
//    0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
//    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
//    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
//    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
//    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
//    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
//    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
//    3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8
//};

/////////////////////////////////////////////////////////////
var s_SuppSize [MAXOUTPUTS]int32;
func CompareSupports(ptrX *int32, ptrY *int32) int32 {
    return ( s_SuppSize[*ptrY] - s_SuppSize[*ptrX] );
}
/////////////////////////////////////////////////////////////

 
/////////////////////////////////////////////////////////////
var s_MintOnes [MAXOUTPUTS]int32;
func CompareMinterms(ptrX *int32, ptrY *int32) int32 {
    return ( s_MintOnes[*ptrY] - s_MintOnes[*ptrX] );
}
/////////////////////////////////////////////////////////////

func GrayCode(BinCode int32) int32 { 
  return BinCode ^ ( BinCode >> 1 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func GetSingleOutputFunction(dd *DdManager, pbOuts **DdNode, nOuts int32, pbVarsEnc **DdNode, nVarsEnc int32, fVerbose int32) *DdNode {
    var i int32;
    var bResult, aResult *DdNode;
    var bCube, bTemp, bProd *DdNode;

    var Order [MAXOUTPUTS]int32;
//  int OrderMint[MAXOUTPUTS];
 
    // sort the output according to their support size
    for  i = 0; i < nOuts; i++  {
        s_SuppSize[i] = Cudd_SupportSize( dd, pbOuts[i] );
//      s_MintOnes[i] = BitCount8[i];
        Order[i]      = i;
//      OrderMint[i]  = i;
    }
    
    // order the outputs
    qsort( (void*)Order,     (size_t)nOuts, sizeof(int), (int(*)(void*, void*)) CompareSupports );
    // order the outputs
//  qsort( (void*)OrderMint, (size_t)nOuts, sizeof(int), (int(*)(const void*, const void*)) CompareMinterms );


    bResult = b0;   Cudd_Ref( bResult );
    for  i = 0; i < nOuts; i++  {
//      bCube   = Cudd_bddBitsToCube( dd, OrderMint[i], nVarsEnc, pbVarsEnc );   Cudd_Ref( bCube );
//      bProd   = Cudd_bddAnd( dd, bCube, pbOuts[Order[nOuts-1-i]] );         Cudd_Ref( bProd );
        bCube   = Extra_bddBitsToCube( dd, i, nVarsEnc, pbVarsEnc, 1 );   Cudd_Ref( bCube );
        bProd   = Cudd_bddAnd( dd, bCube, pbOuts[Order[i]] );         Cudd_Ref( bProd );
        Cudd_RecursiveDeref( dd, bCube );

        bResult = Cudd_bddOr( dd, bProd, bTemp = bResult );           Cudd_Ref( bResult );
        Cudd_RecursiveDeref( dd, bTemp );
        Cudd_RecursiveDeref( dd, bProd );
    }

    // convert to the ADD
if fVerbose != 0 {
fmt.Printf("Single BDD size = %6d nodes\n", Cudd_DagSize(bResult)); }
    aResult = Cudd_BddToAdd( dd, bResult );  Cudd_Ref( aResult );
    Cudd_RecursiveDeref( dd, bResult );
if fVerbose != 0 {
fmt.Printf("MTBDD           = %6d nodes\n", Cudd_DagSize(aResult)); }
    Cudd_Deref( aResult );
    return aResult;
}
/*
DdNode * GetSingleOutputFunction( DdManager * dd, DdNode ** pbOuts, int nOuts, DdNode ** pbVarsEnc, int nVarsEnc )
{
    int i;
    DdNode * bResult, * aResult;
    DdNode * bCube, * bTemp, * bProd;

    bResult = b0;   Cudd_Ref( bResult );
    for ( i = 0; i < nOuts; i++ )
    {
//      bCube   = Extra_bddBitsToCube( dd, i, nVarsEnc, pbVarsEnc );   Cudd_Ref( bCube );
        bCube   = Extra_bddBitsToCube( dd, nOuts-1-i, nVarsEnc, pbVarsEnc );   Cudd_Ref( bCube );
        bProd   = Cudd_bddAnd( dd, bCube, pbOuts[i] );                Cudd_Ref( bProd );
        Cudd_RecursiveDeref( dd, bCube );

        bResult = Cudd_bddOr( dd, bProd, bTemp = bResult );           Cudd_Ref( bResult );
        Cudd_RecursiveDeref( dd, bTemp );
        Cudd_RecursiveDeref( dd, bProd );
    }

    // conver to the ADD
    aResult = Cudd_BddToAdd( dd, bResult );  Cudd_Ref( aResult );
    Cudd_RecursiveDeref( dd, bResult );

    Cudd_Deref( aResult );
    return aResult;
}
*/


////////////////////////////////////////////////////////////////////////
///                        INPUT REMAPPING                           ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
DdNode * GetSingleOutputFunctionRemapped( DdManager * dd, DdNode ** pOutputs, int nOuts, DdNode ** pbVarsEnc, int nVarsEnc )
// returns the ADD of the remapped function
{
    static int Permute[MAXINPUTS];
    static DdNode * pRemapped[MAXOUTPUTS];

    var bSupp, bTemp *DdNode;
    var i, Counter int32;
    var bFunc *DdNode;
    var aFunc *DdNode;

    Cudd_AutodynDisable(dd);

    // perform the remapping
    for  i = 0; i < nOuts; i++  {
        // get support
        bSupp = Cudd_Support( dd, pOutputs[i] );    Cudd_Ref( bSupp );

        // create the variable map
        Counter = 0;
        // inlining of macro cuddT
        for  bTemp = bSupp; bTemp != dd.one; bTemp = ((bTemp).type.kids.T)  {
            Permute[bTemp.index] = Counter++;
        }

        // transfer the BDD and remap it
        pRemapped[i] = Cudd_bddPermute( dd, pOutputs[i], Permute );  Cudd_Ref( pRemapped[i] );

        // remove support
        Cudd_RecursiveDeref( dd, bSupp );
    }
    
    // perform the encoding
    bFunc = Extra_bddEncodingBinary( dd, pRemapped, nOuts, pbVarsEnc, nVarsEnc );   Cudd_Ref( bFunc );

    // convert to ADD
    aFunc = Cudd_BddToAdd( dd, bFunc );  Cudd_Ref( aFunc );
    Cudd_RecursiveDeref( dd, bFunc );

    // deref the intermediate results
    for  i = 0; i < nOuts; i++  {
        Cudd_RecursiveDeref( dd, pRemapped[i] );
    }

    Cudd_Deref( aFunc );
    return aFunc;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
DdNode * GetSingleOutputFunctionRemappedNewDD( DdManager * dd, DdNode ** pOutputs, int nOuts, DdManager ** DdNew )
// returns the ADD of the remapped function
{
    static int Permute[MAXINPUTS];
    static DdNode * pRemapped[MAXOUTPUTS];

    static DdNode * pbVarsEnc[MAXINPUTS];
    var nVarsEnc int32;

    var ddnew *DdManager;

    var bSupp, bTemp *DdNode;
    var i, v, Counter int32;
    var bFunc *DdNode;

    // these are in the new manager
    var bFuncNew *DdNode;
    var aFuncNew *DdNode;

    var nVarsMax int32;

    // perform the remapping and write the DDs into the new manager
    for  i = 0; i < nOuts; i++  {
        // get support
        bSupp = Cudd_Support( dd, pOutputs[i] );    Cudd_Ref( bSupp );

        // create the variable map
        // to remap the DD into the upper part of the manager
        Counter = 0;
        // inlining of macro cuddT
        for  bTemp = bSupp; bTemp != dd.one; bTemp = ((bTemp).type.kids.T)  {
            Permute[bTemp.index] = dd.invperm[Counter++];
        }

        // transfer the BDD and remap it
        pRemapped[i] = Cudd_bddPermute( dd, pOutputs[i], Permute );  Cudd_Ref( pRemapped[i] );

        // remove support
        Cudd_RecursiveDeref( dd, bSupp );


        // determine the largest support size
        if nVarsMax < Counter {
            nVarsMax = Counter; }
    }
    
    // select the encoding variables to follow immediately after the original variables
    nVarsEnc = Abc_Base2Log(nOuts);
/*
    for ( v = 0; v < nVarsEnc; v++ )
        if ( nVarsMax + v < dd->size )
            pbVarsEnc[v] = dd->var[ dd->invperm[nVarsMax+v] ];
        else
            pbVarsEnc[v] = Cudd_bddNewVar( dd );
*/
    // create the new variables on top of the manager
    for  v = 0; v < nVarsEnc; v++  {
        pbVarsEnc[v] = Cudd_bddNewVarAtLevel( dd, v );
    }

//fprintf( pTable, "%d ", Cudd_SharingSize( pRemapped, nOuts ) );
//fprintf( pTable, "%d ", Extra_ProfileWidthSharingMax(dd, pRemapped, nOuts) );


    // perform the encoding
    bFunc = Extra_bddEncodingBinary( dd, pRemapped, nOuts, pbVarsEnc, nVarsEnc );   Cudd_Ref( bFunc );


    // find the cross-manager permutation
    // the variable from the level v in the old manager 
    // should become a variable number v in the new manager
    for  v = 0; v < nVarsMax + nVarsEnc; v++  {
        Permute[dd.invperm[v]] = v;
    }


    ///////////////////////////////////////////////////////////////////////////////
    // start the new manager
    ddnew = Cudd_Init( nVarsMax + nVarsEnc, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0);
//  Cudd_AutodynDisable(ddnew);
    Cudd_AutodynEnable(dd, CUDD_REORDER_SYMM_SIFT);

    // transfer it to the new manager
    bFuncNew = Cudd_bddTransferPermute( dd, ddnew, bFunc, Permute );      Cudd_Ref( bFuncNew );
    ///////////////////////////////////////////////////////////////////////////////


    // deref the intermediate results in the old manager
    Cudd_RecursiveDeref( dd, bFunc );
    for  i = 0; i < nOuts; i++  {
        Cudd_RecursiveDeref( dd, pRemapped[i] );
    }


    ///////////////////////////////////////////////////////////////////////////////
    // convert to ADD in the new manager
    aFuncNew = Cudd_BddToAdd( ddnew, bFuncNew );  Cudd_Ref( aFuncNew );
    Cudd_RecursiveDeref( ddnew, bFuncNew );

    // return the manager
    *DdNew = ddnew;
    ///////////////////////////////////////////////////////////////////////////////

    Cudd_Deref( aFuncNew );
    return aFuncNew;
}

////////////////////////////////////////////////////////////////////////
///                        BLIF WRITING FUNCTIONS                    ///
////////////////////////////////////////////////////////////////////////

// void WriteDDintoBLIFfile( FILE * pFile, DdNode * Func, char * OutputName, char * Prefix, char ** InputNames );



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func WriteSingleOutputFunctionBlif(dd *DdManager, aFunc *DdNode, pNames *string, nNames int32, FileName string) {
    var i int32;
    var pFile *os.File;

    // start the file
    pFile = os.OpenFile(FileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, ".model %s\n", FileName);

    fmt.Fprintf(pFile, ".inputs");
    for  i = 0; i < nNames; i++  {
        fmt.Fprintf(pFile, " %s", pNames[i]);
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".outputs F");
    fmt.Fprintf(pFile, "\n");

    // write the DD into the file
    WriteDDintoBLIFfile( pFile, aFunc, "F", "", pNames );

    fmt.Fprintf(pFile, ".end\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void WriteDDintoBLIFfile( FILE * pFile, DdNode * Func, char * OutputName, char * Prefix, char ** InputNames )
// writes the main part of the BLIF file 
// Func is a BDD or a 0-1 ADD to be written
// OutputName is the name of the output
// Prefix is attached to each intermendiate signal to make it unique
// InputNames are the names of the input signals
// (some part of the code is borrowed from Cudd_DumpDot())
{
    var i int32;
    var visited *st__table;
    var gen *st__generator  = nil;
    var refAddr, diff, mask long;
    var Node, Else, ElseR, Then *DdNode;

    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash );

    /* Collect all the nodes of this DD in the symbol table. */
    // inlining of macro Cudd_Regular
    cuddCollectNodes( ((DdNode *)((ptruint)(Func) & ~01)), visited );

    /* Find how many most significant hex digits are identical
       ** in the addresses of all the nodes. Build a mask based
       ** on this knowledge, so that digits that carry no information
       ** will not be printed. This is done in two steps.
       **  1. We scan the symbol table to find the bits that differ
       **     in at least 2 addresses.
       **  2. We choose one of the possible masks. There are 8 possible
       **     masks for 32-bit integer, and 16 possible masks for 64-bit
       **     integers.
     */

    /* Find the bits that are different. */
    // inlining of macro Cudd_Regular
    refAddr = int64(  ((DdNode *)((ptruint)(Func) & ~01)));
    diff = 0;
    gen = st__init_gen( visited );
    for  st__gen( gen, ( char ** ) &Node, nil )  {
        diff |= refAddr ^ int64(   Node);
    }
    st__free_gen( gen );
    gen = nil;

    /* Choose the mask. */
    for  i = 0; uint32(   i) < 8 * sizeof( long ); i += 4  {
        mask = ( 1 << i ) - 1;
        if diff <= mask {
            break; }
    }


    // write the buffer for the output
    // inlining of macro Cudd_Regular
    fmt.Fprintf(pFile, ".names %s%x %s\n", Prefix, ( mask & int64(((DdNode *)((ptruint)(Func) & ~01))) ) / sizeof(DdNode), OutputName); 
    // inlining of macro Cudd_IsComplement
    fmt.Fprintf(pFile, "%s 1\n", ((int32( ((ptrint) (Func) & 01))))? "0": "1");


    gen = st__init_gen( visited );
    for  st__gen( gen, ( char ** ) &Node, nil )  {
        if Node.index == CUDD_MAXINDEX {
            // write the terminal node
            fmt.Fprintf(pFile, ".names %s%x\n", Prefix, ( mask & int64(Node) ) / sizeof(DdNode));
            // inlining of macro cuddV
            fmt.Fprintf(pFile, " %s\n", (((Node).type.value) == 0.0)? "0": "1");
            continue;
        }

        // inlining of macro cuddE
        Else  = ((Node).type.kids.E);
        // inlining of macro Cudd_Regular
        ElseR = ((DdNode *)((ptruint)(Else) & ~01));
        // inlining of macro cuddT
        Then  = ((Node).type.kids.T);

        assert(InputNames[Node.index], "InputNames[Node.index]");
        if Else == ElseR { // no inverter
            fmt.Fprintf(pFile, ".names %s %s%x %s%x %s%x\n", InputNames[Node.index], Prefix, ( mask & int64(ElseR) ) / sizeof(DdNode), Prefix, ( mask & int64(Then)  ) / sizeof(DdNode), Prefix, ( mask & int64(Node)  ) / sizeof(DdNode));
            fmt.Fprintf(pFile, "01- 1\n");
            fmt.Fprintf(pFile, "1-1 1\n");
        
        } else { // inverter
            var pSlot *int32;
            fmt.Fprintf(pFile, ".names %s %s%x_i %s%x %s%x\n", InputNames[Node.index], Prefix, ( mask & int64(ElseR) ) / sizeof(DdNode), Prefix, ( mask & int64(Then)  ) / sizeof(DdNode), Prefix, ( mask & int64(Node)  ) / sizeof(DdNode));
            fmt.Fprintf(pFile, "01- 1\n");
            fmt.Fprintf(pFile, "1-1 1\n");

            // if the inverter is written, skip
            if ! st__find( visited, (char *)ElseR, (char ***)&pSlot ) {
                assert(false, "0"); }
            if *pSlot {
                continue; }
            *pSlot = 1;

            fmt.Fprintf(pFile, ".names %s%x %s%x_i\n", Prefix, ( mask & int64(ElseR)  ) / sizeof(DdNode), Prefix, ( mask & int64(ElseR)  ) / sizeof(DdNode));
            fmt.Fprintf(pFile, "0 1\n");
        }
    }
    st__free_gen( gen );
    gen = nil;
    st__free_table( visited );
}




var s_ddmin *DdManager;

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void WriteDDintoBLIFfileReorder( DdManager * dd, FILE * pFile, DdNode * Func, char * OutputName, char * Prefix, char ** InputNames )
// writes the main part of the BLIF file 
// Func is a BDD or a 0-1 ADD to be written
// OutputName is the name of the output
// Prefix is attached to each intermendiate signal to make it unique
// InputNames are the names of the input signals
// (some part of the code is borrowed from Cudd_DumpDot())
{
    var i int32;
    var visited *st__table;
    var gen *st__generator  = nil;
    var refAddr, diff, mask long;
    var Node, Else, ElseR, Then *DdNode;


    ///////////////////////////////////////////////////////////////
    var bFmin *DdNode;
    var clk1 abctime;

    if s_ddmin == nil {
        s_ddmin = Cudd_Init( dd.size, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0); }

    clk1 = Abc_Clock();
    bFmin = Cudd_bddTransfer( dd, s_ddmin, Func );  Cudd_Ref( bFmin );

    // reorder
    fmt.Printf("Nodes before = %d.   ", Cudd_DagSize(bFmin)); 
    Cudd_ReduceHeap(s_ddmin,CUDD_REORDER_SYMM_SIFT,1);
//  Cudd_ReduceHeap(s_ddmin,CUDD_REORDER_SYMM_SIFT_CONV,1);
    fmt.Printf("Nodes after  = %d.  \n", Cudd_DagSize(bFmin)); 
    ///////////////////////////////////////////////////////////////



    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash );

    /* Collect all the nodes of this DD in the symbol table. */
    // inlining of macro Cudd_Regular
    cuddCollectNodes( ((DdNode *)((ptruint)(bFmin) & ~01)), visited );

    /* Find how many most significant hex digits are identical
       ** in the addresses of all the nodes. Build a mask based
       ** on this knowledge, so that digits that carry no information
       ** will not be printed. This is done in two steps.
       **  1. We scan the symbol table to find the bits that differ
       **     in at least 2 addresses.
       **  2. We choose one of the possible masks. There are 8 possible
       **     masks for 32-bit integer, and 16 possible masks for 64-bit
       **     integers.
     */

    /* Find the bits that are different. */
    // inlining of macro Cudd_Regular
    refAddr = int64(  ((DdNode *)((ptruint)(bFmin) & ~01)));
    diff = 0;
    gen = st__init_gen( visited );
    for  st__gen( gen, ( char ** ) &Node, nil )  {
        diff |= refAddr ^ int64(   Node);
    }
    st__free_gen( gen );
    gen = nil;

    /* Choose the mask. */
    for  i = 0; uint32(   i) < 8 * sizeof( long ); i += 4  {
        mask = ( 1 << i ) - 1;
        if diff <= mask {
            break; }
    }


    // write the buffer for the output
    // inlining of macro Cudd_Regular
    fmt.Fprintf(pFile, ".names %s%x %s\n", Prefix, ( mask & int64(((DdNode *)((ptruint)(bFmin) & ~01))) ) / sizeof(DdNode), OutputName); 
    // inlining of macro Cudd_IsComplement
    fmt.Fprintf(pFile, "%s 1\n", ((int32( ((ptrint) (bFmin) & 01))))? "0": "1");


    gen = st__init_gen( visited );
    for  st__gen( gen, ( char ** ) &Node, nil )  {
        if Node.index == CUDD_MAXINDEX {
            // write the terminal node
            fmt.Fprintf(pFile, ".names %s%x\n", Prefix, ( mask & int64(Node) ) / sizeof(DdNode));
            // inlining of macro cuddV
            fmt.Fprintf(pFile, " %s\n", (((Node).type.value) == 0.0)? "0": "1");
            continue;
        }

        // inlining of macro cuddE
        Else  = ((Node).type.kids.E);
        // inlining of macro Cudd_Regular
        ElseR = ((DdNode *)((ptruint)(Else) & ~01));
        // inlining of macro cuddT
        Then  = ((Node).type.kids.T);

        assert(InputNames[Node.index], "InputNames[Node.index]");
        if Else == ElseR { // no inverter
            fmt.Fprintf(pFile, ".names %s %s%x %s%x %s%x\n", InputNames[Node.index], Prefix, ( mask & int64(ElseR) ) / sizeof(DdNode), Prefix, ( mask & int64(Then)  ) / sizeof(DdNode), Prefix, ( mask & int64(Node)  ) / sizeof(DdNode));
            fmt.Fprintf(pFile, "01- 1\n");
            fmt.Fprintf(pFile, "1-1 1\n");
        
        } else { // inverter
            fmt.Fprintf(pFile, ".names %s %s%x_i %s%x %s%x\n", InputNames[Node.index], Prefix, ( mask & int64(ElseR) ) / sizeof(DdNode), Prefix, ( mask & int64(Then)  ) / sizeof(DdNode), Prefix, ( mask & int64(Node)  ) / sizeof(DdNode));
            fmt.Fprintf(pFile, "01- 1\n");
            fmt.Fprintf(pFile, "1-1 1\n");

            fmt.Fprintf(pFile, ".names %s%x %s%x_i\n", Prefix, ( mask & int64(ElseR)  ) / sizeof(DdNode), Prefix, ( mask & int64(ElseR)  ) / sizeof(DdNode));
            fmt.Fprintf(pFile, "0 1\n");
        }
    }
    st__free_gen( gen );
    gen = nil;
    st__free_table( visited );


    //////////////////////////////////////////////////
    Cudd_RecursiveDeref( s_ddmin, bFmin );
    //////////////////////////////////////////////////
}




////////////////////////////////////////////////////////////////////////
///                    TRANSFER WITH MAPPING                         ///
////////////////////////////////////////////////////////////////////////
// static DdNode * cuddBddTransferPermuteRecur
// ARGS((DdManager * ddS, DdManager * ddD, DdNode * f, st__table * table, int * Permute ));


// static DdNode * cuddBddTransferPermute
// ARGS((DdManager * ddS, DdManager * ddD, DdNode * f, int * Permute));


/**Function********************************************************************

  Synopsis    [Convert a BDD from a manager to another one.]

  Description [Convert a BDD from a manager to another one. The orders of the
  variables in the two managers may be different. Returns a
  pointer to the BDD in the destination manager if successful; NULL
  otherwise. The i-th entry in the array Permute tells what is the index
  of the i-th variable from the old manager in the new manager.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_bddTransferPermute(ddSource *DdManager, ddDestination *DdManager, f *DdNode, Permute *int32) *DdNode {
    var res *DdNode;
    for
    {
        ddDestination.reordered = 0;
        res = cuddBddTransferPermute( ddSource, ddDestination, f, Permute );
    
if !( ddDestination.reordered == 1 ) { break }
}
    
    return ( res );

}                               /* end of Cudd_bddTransferPermute */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Convert a BDD from a manager to another one.]

  Description [Convert a BDD from a manager to another one. Returns a
  pointer to the BDD in the destination manager if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddTransferPermute]

******************************************************************************/
func cuddBddTransferPermute(ddS *DdManager, ddD *DdManager, f *DdNode, Permute *int32) *DdNode {
    var res *DdNode;
    var table *st__table  = nil;
    var gen *st__generator  = nil;
    var key, value *DdNode;

    table = st__init_table( st__ptrcmp, st__ptrhash );
    if table == nil {
        goto failure; }
    res = cuddBddTransferPermuteRecur( ddS, ddD, f, table, Permute );
    if res != nil {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref); }

    /* Dereference all elements in the table and dispose of the table.
       ** This must be done also if res is NULL to avoid leaks in case of
       ** reordering. */
    gen = st__init_gen( table );
    if gen == nil {
        goto failure; }
    for  st__gen( gen, ( char ** ) &key, ( char ** ) &value )  {
        Cudd_RecursiveDeref( ddD, value );
    }
    st__free_gen( gen );
    gen = nil;
    st__free_table( table );
    table = nil;

    if res != nil {
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref); }
    return ( res );

  failure:
    if table != nil {
        st__free_table( table ); }
    if gen != nil {
        st__free_gen( gen ); }
    return ( nil );

}                               /* end of cuddBddTransferPermute */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddTransferPermute.]

  Description [Performs the recursive step of Cudd_bddTransferPermute.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [cuddBddTransferPermute]

******************************************************************************/
func cuddBddTransferPermuteRecur(ddS *DdManager, ddD *DdManager, f *DdNode, table *st__table, Permute *int32) *DdNode {
    var ft, fe, t, e, var, res *DdNode;
    var one, zero *DdNode;
    var index int32;
    var comple int32;

    statLine( ddD );
    // inlining of macro DD_ONE
    one = ((ddD).one);
    // inlining of macro Cudd_IsComplement
    comple = (int32( ((ptrint) (f) & 01)));

    /* Trivial cases. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if  (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_NotCond
        return ( ((DdNode *)((ptrint)(one) ^ (comple))) ); }

    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_NotCond
    f = ((DdNode *)((ptrint)(f) ^ (comple)));
    /* Now f is a regular pointer to a non-constant node. */

    /* Check the cache. */
    if st__lookup( table, ( char * ) f, ( char ** ) &res ) {
        // inlining of macro Cudd_NotCond
        return ( ((DdNode *)((ptrint)(res) ^ (comple))) ); }

    /* Recursive step. */
    index = Permute[f.index];
    // inlining of macro cuddT
    ft = ((f).type.kids.T);
    // inlining of macro cuddE
    fe = ((f).type.kids.E);

    t = cuddBddTransferPermuteRecur( ddS, ddD, ft, table, Permute );
    if t == nil {
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddTransferPermuteRecur( ddS, ddD, fe, table, Permute );
    if e == nil {
        Cudd_RecursiveDeref( ddD, t );
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));
    var = cuddUniqueInter( ddD, index, one, zero );
    if var == nil {
        Cudd_RecursiveDeref( ddD, t );
        Cudd_RecursiveDeref( ddD, e );
        return ( nil );
    }
    res = cuddBddIteRecur( ddD, var, t, e );
    if res == nil {
        Cudd_RecursiveDeref( ddD, t );
        Cudd_RecursiveDeref( ddD, e );
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref( ddD, t );
    Cudd_RecursiveDeref( ddD, e );

    if st__add_direct( table, ( char * ) f, ( char * ) res ) ==
         st__OUT_OF_MEM {
        Cudd_RecursiveDeref( ddD, res );
        return ( nil );
    }
    // inlining of macro Cudd_NotCond
    return ( ((DdNode *)((ptrint)(res) ^ (comple))) );

}                               /* end of cuddBddTransferPermuteRecur */

////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////




//ABC_NAMESPACE_IMPL_END

