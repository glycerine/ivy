package rwr

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [rwrCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [Cut computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwrCut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var pTruths [13719]int32;
var pFreqs [13719]int32;
var pPerm [13719]int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_TempCompare(pNum1 *int32, pNum2 *int32) int32 {
    var Freq1 int32  = pFreqs[*pNum1];
    var Freq2 int32  = pFreqs[*pNum2];
    if Freq1 < Freq2 {
        return 1; }
    if Freq1 > Freq2 {
        return -1; }
    return 0; 
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_Temp() {
    var Buffer [32]byte;
    var nFuncs int32  = 13719;
    var nEntries int32  = 100;
    var uTruth uint32;
    var i, k int32;
    var pFile *os.File;

    pFile = os.Open("nnclass_stats5.txt");
    for  i = 0; i < 13719; i++  {
        var RetValue int32  = fscanf( pFile, "%s%d", Buffer, &pFreqs[i] );
        Extra_ReadHexadecimal( &uTruth, Buffer+2, 5 );
        pTruths[i] = uTruth;
    }
    _ = (pFile).Close();

    for  i = 0; i < 13719; i++  {
        pPerm[i] = i;
    }

    qsort( (void *)pPerm, (size_t)13719, sizeof(int), 
                (int (*)(void *, void *)) Rwr_TempCompare );


    pFile = os.OpenFile("5npn_100.blif", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, "# Most frequent NPN classes of 5 vars.\n");
    fmt.Fprintf(pFile, ".model 5npn\n");
    fmt.Fprintf(pFile, ".inputs a b c d e\n");
    fmt.Fprintf(pFile, ".outputs");
    for  i = 0; i < nEntries; i++  {
        fmt.Fprintf(pFile, " %02d", i);
    }
    fmt.Fprintf(pFile, "\n");

    for  i = 0; i < nEntries; i++  {
        fmt.Fprintf(pFile, ".names a b c d e %02d\n", i);
        uTruth = pTruths[pPerm[i]];
        for  k = 0; k < 32; k++  {
            if uTruth & (1 << k) {
                Extra_PrintBinary( pFile, (unsigned *)&k, 5 );
                fmt.Fprintf(pFile, " 1\n");
            }
        }
    }
    fmt.Fprintf(pFile, ".end\n");
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

