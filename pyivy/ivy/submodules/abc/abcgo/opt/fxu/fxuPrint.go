package fxu

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [fxuPrint.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Various printing procedures.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuPrint.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"

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
func Fxu_MatrixPrint(pFile *os.File, p *Fxu_Matrix) {
    var pVar *Fxu_Var;
    var pCube *Fxu_Cube;
    var pDiv *Fxu_Double;
    var pSingle *Fxu_Single;
    var pLit *Fxu_Lit;
    var pPair *Fxu_Pair;
    var i, LastNum int32;
    var fStdout int32;

    fStdout = 1;
    if pFile == nil {
        pFile = os.OpenFile("matrix.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        fStdout = 0;
    }

    fmt.Fprintf(pFile, "Matrix has %d vars, %d cubes, %d literals, %d divisors.\n", p.lVars.nItems, p.lCubes.nItems, p.nEntries, p.nDivs);
    fmt.Fprintf(pFile, "Divisors selected so far: single = %d, double = %d.\n", p.nDivs1, p.nDivs2);
    fmt.Fprintf(pFile, "\n");

    // print the numbers on top of the matrix
    for  i = 0; i < 12; i++  {
        fmt.Fprintf(pFile, " ");
    }
    // inlining of macro Fxu_MatrixForEachVariable
    for  pVar = (p).lVars.pHead;          (pVar != nil);          pVar = pVar.pNext  {
        fmt.Fprintf(pFile, "%d", pVar.iVar % 10);
    }
    fmt.Fprintf(pFile, "\n");

    // print the rows
    // inlining of macro Fxu_MatrixForEachCube
    for  pCube = (p).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        fmt.Fprintf(pFile, "%4d", pCube.iCube);
        fmt.Fprintf(pFile, "  ");
        fmt.Fprintf(pFile, "%4d", pCube.pVar.iVar);
        fmt.Fprintf(pFile, "  ");

        // print the literals
        LastNum = -1;
        // inlining of macro Fxu_CubeForEachLiteral
        for  pLit = (pCube).lLits.pHead;          (pLit != nil);          pLit = pLit.pHNext  {
            for  i = LastNum + 1; i < pLit.pVar.iVar; i++  {
                fmt.Fprintf(pFile, ".");
            }
            fmt.Fprintf(pFile, "1");
            LastNum = i;
        }
        for  i = LastNum + 1; i < p.lVars.nItems; i++  {
            fmt.Fprintf(pFile, ".");
        }
        fmt.Fprintf(pFile, "\n");
    }
    fmt.Fprintf(pFile, "\n");

    // print the double-cube divisors
    fmt.Fprintf(pFile, "The double divisors are:\n");
    // inlining of macro Fxu_MatrixForEachDouble
    // inlining of macro Fxu_TableForEachDouble
    for  i = 0; i < (p).nTableSize; i++  {        for  pDiv = (p).pTable[i].pHead;          (pDiv != nil);          pDiv = pDiv.pNext  {
        fmt.Fprintf(pFile, "Divisor #%3d (lit=%d,%d) (w=%2d):  ", pDiv.Num, pDiv.lPairs.pHead.nLits1, pDiv.lPairs.pHead.nLits2, pDiv.Weight);
        // inlining of macro Fxu_DoubleForEachPair
        for  pPair = (pDiv).lPairs.pHead;          (pPair != nil);          pPair = pPair.pDNext  {
            fmt.Fprintf(pFile, " <%d, %d> (b=%d)", pPair.pCube1.iCube, pPair.pCube2.iCube, pPair.nBase);
        }
        fmt.Fprintf(pFile, "\n");
    }
    }
    fmt.Fprintf(pFile, "\n");

    // print the divisors associated with each cube
    fmt.Fprintf(pFile, "The cubes are:\n");
    // inlining of macro Fxu_MatrixForEachCube
    for  pCube = (p).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        fmt.Fprintf(pFile, "Cube #%3d: ", pCube.iCube);
        if pCube.pVar.ppPairs {
            // inlining of macro Fxu_CubeForEachPair
            for  i = 0;        i < pCube.pVar.nCubes && (((pPair) = (pCube).pVar.ppPairs[(pCube).iCube][i]), 1);        i++  {        if pPair == nil {} else {
                fmt.Fprintf(pFile, " <%d %d> (d=%d) (b=%d)", pPair.iCube1, pPair.iCube2, pPair.pDiv.Num, pPair.nBase); }
            }
        }
        fmt.Fprintf(pFile, "\n");
    }
    fmt.Fprintf(pFile, "\n");

    // print the single-cube divisors
    fmt.Fprintf(pFile, "The single divisors are:\n");
    // inlining of macro Fxu_MatrixForEachSingle
    for  pSingle = (p).lSingles.pHead;          (pSingle != nil);          pSingle = pSingle.pNext  {
        fmt.Fprintf(pFile, "Single-cube divisor #%5d: Var1 = %4d. Var2 = %4d. Weight = %2d\n", pSingle.Num, pSingle.pVar1.iVar, pSingle.pVar2.iVar, pSingle.Weight);
    }
    fmt.Fprintf(pFile, "\n");

/*
    {
        int Index;
        fprintf( pFile, "Distribution of divisors in the hash table:\n" );
        for ( Index = 0; Index < p->nTableSize; Index++ )
            fprintf( pFile, " %d", p->pTable[Index].nItems );
        fprintf( pFile, "\n" );
    }
*/
    if !(fStdout != 0) {
        _ = (pFile).Close(); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixPrintDivisorProfile(pFile *os.File, p *Fxu_Matrix) {
    var pDiv *Fxu_Double;
    var WeightMax int32;
    var pProfile *int32;
    var Counter1 int32; // the number of -1 weight
    var CounterL int32; // the number of less than -1 weight
    var i int32;

    WeightMax = Fxu_HeapDoubleReadMaxWeight( p.pHeapDouble );
    // inlining of macro ABC_ALLOC
    pProfile = ((int *) malloc(sizeof(int) * (size_t)((WeightMax + 1))));
    memset( pProfile, 0, sizeof(int) * (WeightMax + 1) );

    Counter1 = 0;
    CounterL = 0;
    // inlining of macro Fxu_MatrixForEachDouble
    // inlining of macro Fxu_TableForEachDouble
    for  i = 0; i < (p).nTableSize; i++  {        for  pDiv = (p).pTable[i].pHead;          (pDiv != nil);          pDiv = pDiv.pNext  {
        assert(pDiv.Weight <= WeightMax, "pDiv.Weight <= WeightMax");
        if pDiv.Weight == -1 {
            Counter1++;
        } else if pDiv.Weight < 0 {
            CounterL++;
        } else {
            pProfile[ pDiv.Weight ]++; }
    }
    }

    fmt.Fprintf(pFile, "The double divisors profile:\n");
    fmt.Fprintf(pFile, "Weight  < -1 divisors = %6d\n", CounterL);
    fmt.Fprintf(pFile, "Weight    -1 divisors = %6d\n", Counter1);
    for  i = 0; i <= WeightMax; i++  {
        if pProfile[i] {
            fmt.Fprintf(pFile, "Weight   %3d divisors = %6d\n", i, pProfile[i]); }
    }
    fmt.Fprintf(pFile, "End of divisor profile printout\n");
    // inlining of macro ABC_FREE
    ((pProfile) ? (free((char *) (pProfile)), (pProfile) = 0) : 0);
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

