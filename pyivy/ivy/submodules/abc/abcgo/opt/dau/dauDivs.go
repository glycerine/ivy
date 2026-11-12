package dau

import "fmt"

/**CFile****************************************************************

  FileName    [dauDivs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Divisor computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauDivs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/util/utilTruth.h"

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
func Dau_DsdDivisors(pTruth *word, nVars int32) *Vec_Int_t {
    var Copy [DAU_MAX_WORD]word;
    var nWords int32  = Abc_TtWordNum(nVars);
    var nDigits int32  = Abc_TtHexDigitNum(nVars);
    var i int32;
    var j int32;
    var k int32;
    var Digit int32;
    var Counter [5]int32;

    fmt.Printf("     ");
    fmt.Printf(" !a *!b");
    fmt.Printf(" !a * b");
    fmt.Printf("  a *!b");
    fmt.Printf("  a * b");
    fmt.Printf("  a + b");
    fmt.Printf("\n");

    for  i = 0; i < nVars; i++  {
    for  j = i+1; j < nVars; j++  {
        Abc_TtCopy( Copy, pTruth, nWords, 0 );
        Abc_TtSwapVars( Copy, nVars, 0, i );
        Abc_TtSwapVars( Copy, nVars, 1, j );
        for  k = 0; k < 5; k++  {
            Counter[k] = 0;
        }
        for  k = 0; k < nDigits; k++  {
            Digit = Abc_TtGetHex( Copy, k );
            if Digit == 1 || Digit == 14 {
                Counter[0]++;
            } else if Digit == 2 || Digit == 13 {
                Counter[1]++;
            } else if Digit == 4 || Digit == 11 {
                Counter[2]++;
            } else if Digit == 8 || Digit == 7 {
                Counter[3]++;
            } else if Digit == 6 || Digit == 9 {
                Counter[4]++; }
        }
        fmt.Printf("%c %c  ", 'a'+i, 'a'+j);
        for  k = 0; k < 5; k++  {
            fmt.Printf("%7d", Counter[k]);
        }
        fmt.Printf("\n");
    }
    }
    return nil;
}
func Dau_DsdTest000() {
//    char * pDsd = "!(!(abc)!(def))";
//    char * pDsd = "[(abc)(def)]";
    var pDsd string  = "<<abc>d(ef)>";
    var t word  = Dau_Dsd6ToTruth( pDsd );
//    word t = 0xCA88CA88CA88CA88;
//    word t = 0x9ef7a8d9c7193a0f;
    var nVars int32  = Abc_TtSupportSize( &t, 6 );
    return;
//    word t = 0xCACACACACACACACA;
    Dau_DsdDivisors( &t, nVars );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

