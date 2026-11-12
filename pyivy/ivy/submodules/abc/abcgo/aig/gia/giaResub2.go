package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaResub2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaResub2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"
//#include "misc/vec/vecHsh.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
   


// extern void Abc_ResubPrepareManager( int nWords );

// extern int Abc_ResubComputeFunction( void ** ppDivs, int nDivs, int nWords, int nLimit, int nDivsMax, int iChoice, int fUseXor, int fDebug, int fVerbose, int ** ppArray );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




// marks MFFC and returns its size


// collects divisors and maps them into nodes
// assumes MFFC is already marked


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ResubPrintDivs(ppDivs *any, nDivs int32) {
    var pDivs **word  = (word **)ppDivs;
    var i int32;
    for  i = 0; i < nDivs; i++  {
        fmt.Printf("Div %2d : ", i);
        Dau_DsdPrintFromTruth( pDivs[i], 6 );
    }
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Creating a window with support composed of primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// returns the number of nodes added to the window when is iPivot is added
// the window nodes in vNodes are labeled with the current traversal ID
// the new node iPivot and its fanout are temporarily labeled and then unlabeled

// find the first PI to add based on the fanout count

// find the next PI to add based on how many nodes will be added to the window

// check if the node has unmarked fanouts

// this is a translation procedure, which converts the array of node IDs (vObjs) 
// into the internal represnetation for the resub engine, which is returned
// (this procedure is not needed when we simply construct a window)

// construct a high-volume window support by the given number (nPis) of primary inputs






/**Function*************************************************************

  Synopsis    [Return initial window for the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






// collect nodes on the path from the meeting point to the root node, excluding the meeting point

// explore the frontier of nodes in the breadth-first traversal




/**Function*************************************************************

  Synopsis    [Grow window for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Grow window for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

// uses levelized structure (vLevels) to collect in array vWin divisors supported by the cut (vIn)

// update the cut until both fanins of AND nodes are not in the cut

                        

// precondition: nodes in vWin and in vIns are marked with the current ID


/**Function*************************************************************

  Synopsis    [Create window for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Derive GIA from the window]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Naive truth-table-based verification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Enumerate windows of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Apply k-resub to one AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Apply k-resub to one AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_RsbTestArray() {
    var Array [1000]int32 = [1000]int32{ 
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 6, 3, 7, 15, 17, 8, 19, 
        5, 20, 5, 12, 8, 24, 4, 12, 9, 28, 27, 31, 23, 32, 4, 13, 8, 36, 5, 
        13, 18, 40, 9, 18, 5, 44, 19, 36, 9, 48, 47, 51, 10, 18, 40, 54, 8, 
        56, 25, 37, 44, 61, 59, 63, 8, 28, 8, 18, 25, 68, 66, 70, 64, 73, 11, 
        19, 8, 13, 76, 78, 10, 19, 40, 82, 9, 84, 81, 87, 20, 61, 19, 28, 30, 
        92, 91, 95, 88, 96, 74, 98, 9, 40, 49, 103, 27, 104, 10, 107, 8, 40, 
        9, 24, 111, 113, 11, 115, 109, 117, 11, 66, 51, 121, 118, 122, 18, 36, 
        18, 110, 93, 127, 10, 131, 129, 133, 11, 38, 32, 137, 103, 138, 19, 141, 
        134, 143, 28, 76, 9, 146, 11, 110, 19, 150, 149, 153, 87, 95, 9, 19, 10, 
        159, 61, 160, 18, 30, 61, 158, 9, 12, 25, 169, 19, 171, 111, 173, 10, 175, 
        167, 177, 18, 102, 4, 20, 18, 171, 183, 185, 11, 187, 181, 189, 178, 190, 
        24, 44, 11, 194, 8, 54, 4, 198, 197, 201, 45, 49, 10, 39, 9, 126, 73, 209, 
        11, 211, 54, 168, 213, 215, 43, 167, 67, 218, 10, 221, 26, 54, 18, 18, 34, 
        34, 38, 38, 40, 40, 42, 42, 52, 52, 100, 100, 124, 124, 126, 126, 144, 144, 
        148, 148, 154, 154, 156, 156, 162, 162, 164, 164, 192, 192, 70, 70, 202, 
        202, 204, 204, 206, 206, 216, 216, 222, 222, 224, 224,
    };
    var i, iFan0, iFan1, nResubs int32;
    var pRes *int32;
    // create the internal array
    var vArray *Vec_Int_t  = Vec_IntAlloc( 100 );
    for  i = 0; i < 50 || Array[i] > 0; i++  {
        Vec_IntPush( vArray, Array[i] );
    }
    Vec_IntPrint( vArray );
    // check the nodes
    fmt.Printf("Constant0 and primary inputs:\n");
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vArray)) && (((iFan0) = Vec_IntEntry(vArray, i)), 1) && (((iFan1) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
        if iFan0 != iFan1 {
            break; }
        fmt.Printf("%2d = %c%2d & %c%2d;\n", i, Abc_LitIsCompl(iFan0) ? '!' : ' ', Abc_Lit2Var(iFan0), Abc_LitIsCompl(iFan1) ? '!' : ' ', Abc_Lit2Var(iFan1));
    }
    fmt.Printf("Primary outputs:\n");
    // inlining of macro Vec_IntForEachEntryDoubleStart
    for  i = 14; (i+1 < Vec_IntSize(vArray)) && (((iFan0) = Vec_IntEntry(vArray, i)), 1) && (((iFan1) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
        if iFan0 != iFan1 {
            continue; }
        fmt.Printf("%2d = %c%2d & %c%2d;\n", i, Abc_LitIsCompl(iFan0) ? '!' : ' ', Abc_Lit2Var(iFan0), Abc_LitIsCompl(iFan1) ? '!' : ' ', Abc_Lit2Var(iFan1));
    }
    // run the resub
    Abc_ResubPrepareManager( 1 );
    Abc_ResubComputeWindow( Vec_IntArray(vArray), Vec_IntSize(vArray)/2, 10, -1, 0, 0, 1, 1, &pRes, &nResubs );
    Abc_ResubPrepareManager( 0 );
    Vec_IntFree( vArray );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

