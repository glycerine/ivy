package rsb

import "fmt"

/**CFile****************************************************************

  FileName    [rsbDec6.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Truth-table-based resubstitution.]

  Synopsis    [Implementation of the algorithm.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rsbDec6.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rsbInt.h"

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
func Rsb_DecTry0(c word) uint32 {
    return uint32((c != 0));
}
func Rsb_DecTry1(c word, f1 word) uint32 {
    return (Rsb_DecTry0(c&f1) << 1) | Rsb_DecTry0(c&~f1);
}
func Rsb_DecTry2(c word, f1 word, f2 word) uint32 {
    return (Rsb_DecTry1(c&f2, f1) << 2) | Rsb_DecTry1(c&~f2, f1);
}
func Rsb_DecTry3(c word, f1 word, f2 word, f3 word) uint32 {
    return (Rsb_DecTry2(c&f3, f1, f2) << 4) | Rsb_DecTry2(c&~f3, f1, f2);
}
func Rsb_DecTry4(c word, f1 word, f2 word, f3 word, f4 word) uint32 {
    return (Rsb_DecTry3(c&f4, f1, f2, f3) << 8) | Rsb_DecTry3(c&~f4, f1, f2, f3);
}
func Rsb_DecTry5(c word, f1 word, f2 word, f3 word, f4 word, f5 word) uint32 {
    return (Rsb_DecTry4(c&f5, f1, f2, f3, f4) << 16) | Rsb_DecTry4(c&~f5, f1, f2, f3, f4);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_DecTryCex(g *word, iCexA int32, iCexB int32) int32 {
    return Abc_TtGetBit(g, iCexA) == Abc_TtGetBit(g, iCexB);
}
func Rsb_DecVerifyCex(f *word, g **word, nGs int32, iCexA int32, iCexB int32) {
    var i int32;
    // f distinquished it
    if Rsb_DecTryCex( f, iCexA, iCexB ) != 0 {
        fmt.Printf("Verification of CEX has failed: f(A) == f(B)!!!\n"); }
    // g do not distinguish it
    for  i = 0; i < nGs; i++  {
        if !(Rsb_DecTryCex( g[i], iCexA, iCexB ) != 0) {
            fmt.Printf("Verification of CEX has failed: g[%d](A) != g[%d](B)!!!\n", i, i); }
    }
}
func Rsb_DecRecordCex(g **word, nGs int32, iCexA int32, iCexB int32, pCexes *word, nCexes int32) {
    var i int32;
    assert(nCexes < 64, "nCexes < 64");
    for  i = 0; i < nGs; i++  {
        if Rsb_DecTryCex(g[i], iCexA, iCexB) != 0 {
            Abc_TtSetBit( pCexes + i, nCexes ); }
    }
}

/**Function*************************************************************

  Synopsis    [Checks decomposability of f with divisors g.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_DecCofactor(g **word, nGs int32, w int32, iMint int32) word {
    var i int32;
    var Cof word  = ~(word)0;
    for  i = 0; i < nGs; i++  {
        Cof &=  ((iMint >> i) & 1) ? g[i][w] : ~g[i][w];
    }
    return Cof;
}
func Rsb_DecCheck(nVars int32, f *word, g **word, nGs int32, pPat *uint32, pCexA *int32, pCexB *int32) uint32 {
    var CofA, CofB word;
    var nWords int32  = Abc_TtWordNum( nVars );
    var w int32;
    var k int32;
    var iMint int32;
    var iShift int32 = ( 1 << nGs );
    var uMask uint32  = (~uint32(0)) >> (32-iShift);
    var uTotal uint32;
    assert(nGs > 0 && nGs < 5, "nGs > 0 && nGs < 5");
    for  w = 0; w < nWords; w++  {
        // generate decomposition pattern
        if nGs == 1 {
            pPat[w] = Rsb_DecTry2( ~(word)0, g[0][w], f[w] );
        } else if nGs == 2 {
            pPat[w] = Rsb_DecTry3( ~(word)0, g[0][w], g[1][w], f[w] );
        } else if nGs == 3 {
            pPat[w] = Rsb_DecTry4( ~(word)0, g[0][w], g[1][w], g[2][w], f[w] );
        } else if nGs == 4 {
            pPat[w] = Rsb_DecTry5( ~(word)0, g[0][w], g[1][w], g[2][w], g[3][w], f[w] ); }
        // check if it is consistent
        iMint = Abc_Tt6FirstBit( (pPat[w] >> iShift) & pPat[w] & uMask );
        if iMint >= 0 {   // generate a cex
            CofA = Rsb_DecCofactor( g, nGs, w, iMint );
            assert((~f[w] & CofA) && (f[w] & CofA), "(~f[w] & CofA) && (f[w] & CofA)");
            *pCexA = w * 64 + Abc_Tt6FirstBit( ~f[w] & CofA );
            *pCexB = w * 64 + Abc_Tt6FirstBit(  f[w] & CofA );
            return 0;
        }
        uTotal |= pPat[w];
        if w == 0 {
            continue; }
        // check if it is consistent with other patterns seen so far
        iMint = Abc_Tt6FirstBit( (uTotal >> iShift) & uTotal & uMask );
        if iMint == -1 {
            continue; }
        // find an overlap and generate a cex
        for  k = 0; k < w; k++  {
            iMint = Abc_Tt6FirstBit( ((pPat[k] | pPat[w]) >> iShift) & (pPat[k] | pPat[w]) & uMask );
            if iMint == -1 {
                continue; }
            CofA = Rsb_DecCofactor( g, nGs, k, iMint );
            CofB = Rsb_DecCofactor( g, nGs, w, iMint );
            if (~f[k] & CofA) && (f[w] & CofB) {
                *pCexA = k * 64 + Abc_Tt6FirstBit( ~f[k] & CofA );
                *pCexB = w * 64 + Abc_Tt6FirstBit(  f[w] & CofB );
            
            } else {
                assert((f[k] & CofA) && (~f[w] & CofB), "(f[k] & CofA) && (~f[w] & CofB)");
                *pCexA = k * 64 + Abc_Tt6FirstBit(  f[k] & CofA );
                *pCexB = w * 64 + Abc_Tt6FirstBit( ~f[w] & CofB );
            }
            return 0;
        }
    }
    return uTotal;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_DecPrintTable(pCexes *word, nGs int32, nGsAll int32, vTries *Vec_Int_t) {
    var pCands [16]int32;
    var nCands int32;
    var i, c, Cand, iStart int32;
    if Vec_IntSize(vTries) == 0 {
        return; }

//    printf( "\n" );
    for  i = 0; i < 4; i++  {
        fmt.Printf("    ");
    }
    fmt.Printf("  ");
    for  i = 0; i < nGs; i++  {
        fmt.Printf("%d", (i % 100) / 10);
    }
    fmt.Printf("|");
    for  ; i < nGsAll; i++  {
        fmt.Printf("%d", (i % 100) / 10);
    }
    fmt.Printf("\n");

    for  i = 0; i < 4; i++  {
        fmt.Printf("    ");
    }
    fmt.Printf("  ");
    for  i = 0; i < nGs; i++  {
        fmt.Printf("%d", i % 10);
    }
    fmt.Printf("|");
    for  ; i < nGsAll; i++  {
        fmt.Printf("%d", i % 10);
    }
    fmt.Printf("\n");
    fmt.Printf("\n");

    for  c = 0; iStart < Vec_IntSize(vTries); c++  {
        // collect candidates
        for  i = 0; i < 4; i++  {
            pCands[i] = -1;
        }
        nCands = 0;
        Vec_IntForEachEntryStart( vTries, Cand, i, iStart )
            if Cand == -1 {
                iStart = i + 1;
                break;
            
            } else {
                pCands[nCands++] = Cand; }
        assert(nCands <= 4, "nCands <= 4");
        // print them
        for  i = 0; i < 4; i++  {
            if pCands[i] >= 0 {
                fmt.Printf("%4d", pCands[i]);
            } else {
                fmt.Printf("    "); }
        }
        // print the bit-string
        fmt.Printf("  ");
        for  i = 0; i < nGs; i++  {
            fmt.Printf("%c", Abc_TtGetBit( pCexes + i, c ) ? '.' : '+');
        }
        fmt.Printf("|");
        for  ; i < nGsAll; i++  {
            fmt.Printf("%c", Abc_TtGetBit( pCexes + i, c ) ? '.' : '+');
        }
        fmt.Printf("  %3d\n", c);
    }
    fmt.Printf("\n");

    // write the number of ones
    for  i = 0; i < 4; i++  {
        fmt.Printf("    ");
    }
    fmt.Printf("  ");
    for  i = 0; i < nGs; i++  {
        fmt.Printf("%d", Abc_TtCountOnes(pCexes[i]) / 10);
    }
    fmt.Printf("|");
    for  ; i < nGsAll; i++  {
        fmt.Printf("%d", Abc_TtCountOnes(pCexes[i]) / 10);
    }
    fmt.Printf("\n");

    for  i = 0; i < 4; i++  {
        fmt.Printf("    ");
    }
    fmt.Printf("  ");
    for  i = 0; i < nGs; i++  {
        fmt.Printf("%d", Abc_TtCountOnes(pCexes[i]) % 10);
    }
    fmt.Printf("|");
    for  ; i < nGsAll; i++  {
        fmt.Printf("%d", Abc_TtCountOnes(pCexes[i]) % 10);
    }
    fmt.Printf("\n");
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Init ]

  Description [Returns the numbers of the decomposition functions and 
  the truth table of a function up to 4 variables.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_DecInitCexes(nVars int32, f *word, g **word, nGs int32, nGsAll int32, pCexes *word, vTries *Vec_Int_t) int32 {
    var nWords int32  = Abc_TtWordNum( nVars );
    var ValueB int32  = Abc_TtGetBit( f, 0 );
    var ValueE int32  = Abc_TtGetBit( f, 64*nWords-1 );
    var iCexT0, iCexT1, iCexF0, iCexF1 int32;

    if ValueB != 0 {
        iCexT0 = 0;
    } else {
        iCexT0 = Abc_TtFindFirstBit( f, nVars );
    }
    if ValueE != 0 {
        iCexT1 = 64*nWords-1;
    } else {
        iCexT1 = Abc_TtFindLastBit( f, nVars );
    }

    if !(ValueB != 0) {
        iCexF0 = 0;
    } else {
        iCexF0 = Abc_TtFindFirstZero( f, nVars );
    }
    if !(ValueE != 0) {
        iCexF1 = 64*nWords-1;
    } else {
        iCexF1 = Abc_TtFindLastZero( f, nVars );
    }

    assert(!(Rsb_DecTryCex( f, iCexT0, iCexF0 ) != 0), "!Rsb_DecTryCex( f, iCexT0, iCexF0 )");
    assert(!(Rsb_DecTryCex( f, iCexT0, iCexF1 ) != 0), "!Rsb_DecTryCex( f, iCexT0, iCexF1 )");
    assert(!(Rsb_DecTryCex( f, iCexT1, iCexF0 ) != 0), "!Rsb_DecTryCex( f, iCexT1, iCexF0 )");
    assert(!(Rsb_DecTryCex( f, iCexT1, iCexF1 ) != 0), "!Rsb_DecTryCex( f, iCexT1, iCexF1 )");

    Rsb_DecRecordCex( g, nGsAll, iCexT0, iCexF0, pCexes, 0 );
    Rsb_DecRecordCex( g, nGsAll, iCexT0, iCexF1, pCexes, 1 );
    Rsb_DecRecordCex( g, nGsAll, iCexT1, iCexF0, pCexes, 2 );
    Rsb_DecRecordCex( g, nGsAll, iCexT1, iCexF1, pCexes, 3 );

    if vTries != nil {
    Vec_IntPush( vTries, -1 );
    Vec_IntPush( vTries, -1 );
    Vec_IntPush( vTries, -1 );
    Vec_IntPush( vTries, -1 );
    }
    return 4;
}

/**Function*************************************************************

  Synopsis    [Finds a setset of gs to decompose f.]

  Description [Returns the numbers of the decomposition functions and 
  the truth table of a function up to 4 variables.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_DecPerformInt(p *Rsb_Man_t, nVars int32, f *word, g **word, nGs int32, nGsAll int32, fFindAll int32) uint32 {
    var pCexes *word  = Vec_WrdArray(p.vCexes);
    var pPat *uint32  = (unsigned *)Vec_IntArray(p.vDecPats);
    /*
    The following filtering hueristic can be used:
    after the first round (if there is more than 5 cexes, 
    remove all the divisors, except fanins of the node
    This should lead to a speadup without sacrifying quality.

    Another idea is to precompute several counter-examples
    like the first and last 0 and 1 mints of the function
    which yields 4 cexes.
    */

    var pDivs [16]*word;
    var uTruth uint32;
    var i, k, j, n, iCexA, iCexB, nCexes int32;
    memset( pCexes, 0, sizeof(word) * nGsAll );  
    Vec_IntClear( p.vTries );
    // populate the counter-examples with the three most obvious
//    nCexes = Rsb_DecInitCexes( nVars, f, g, nGs, nGsAll, pCexes, vTries );
    // start by checking each function
    p.vFanins.nSize = 1;
    for  i = 0; i < nGs; i++  {
        if pCexes[i] {
            continue; }
        pDivs[0] = g[i];  p.vFanins.pArray[0] = i;
        uTruth = Rsb_DecCheck( nVars, f, pDivs, Vec_IntSize(p.vFanins), pPat, &iCexA, &iCexB );
        if uTruth != 0 {
            if fFindAll != 0 {
                uTruth = uint32(Abc_Tt6Stretch( (word)uTruth, 1 ));
                Kit_DsdPrintFromTruth( &uTruth, 1 ); fmt.Printf("   ");
                Vec_IntPrint(p.vFanins);
                continue;
            
            } else {
                return uTruth; }
        }
        if nCexes == 64 {
            return 0; }
        Rsb_DecVerifyCex( f, pDivs, Vec_IntSize(p.vFanins), iCexA, iCexB );
        Rsb_DecRecordCex( g, nGsAll, iCexA, iCexB, pCexes, nCexes++ );
        if !p.fVerbose {
            continue; }
        Vec_IntPush( p.vTries, i );
        Vec_IntPush( p.vTries, -1 );
    }
    if p.nDecMax == 1 {
        return 0; }
    // continue by checking pairs
    p.vFanins.nSize = 2;
    for  i = 1; i < nGs; i++  {
    for  k = 0; k < i; k++  {
        if pCexes[i] & pCexes[k] {
            continue; }
        pDivs[0] = g[i];  p.vFanins.pArray[0] = i;
        pDivs[1] = g[k];  p.vFanins.pArray[1] = k;
        uTruth = Rsb_DecCheck( nVars, f, pDivs, Vec_IntSize(p.vFanins), pPat, &iCexA, &iCexB );
        if uTruth != 0 {
            if fFindAll != 0 {
                uTruth = uint32(Abc_Tt6Stretch( (word)uTruth, 2 ));
                Kit_DsdPrintFromTruth( &uTruth, 2 ); fmt.Printf("   ");
                Vec_IntPrint(p.vFanins);
                continue;
            
            } else {
                return uTruth; }
        }
        if nCexes == 64 {
            return 0; }
        Rsb_DecVerifyCex( f, pDivs, Vec_IntSize(p.vFanins), iCexA, iCexB );
        Rsb_DecRecordCex( g, nGsAll, iCexA, iCexB, pCexes, nCexes++ );
        if !p.fVerbose {
            continue; }
        Vec_IntPush( p.vTries, i );
        Vec_IntPush( p.vTries, k );
        Vec_IntPush( p.vTries, -1 );
    }
    }
    if p.nDecMax == 2 {
        return 0; }
    // continue by checking triples
    p.vFanins.nSize = 3;
    for  i = 2; i < nGs; i++  {
    for  k = 1; k < i; k++  {
    for  j = 0; j < k; j++  {
        if pCexes[i] & pCexes[k] & pCexes[j] {
            continue; }
        pDivs[0] = g[i];  p.vFanins.pArray[0] = i;
        pDivs[1] = g[k];  p.vFanins.pArray[1] = k;
        pDivs[2] = g[j];  p.vFanins.pArray[2] = j;
        uTruth = Rsb_DecCheck( nVars, f, pDivs, Vec_IntSize(p.vFanins), pPat, &iCexA, &iCexB );
        if uTruth != 0 {
            if fFindAll != 0 {
                uTruth = uint32(Abc_Tt6Stretch( (word)uTruth, 3 ));
                Kit_DsdPrintFromTruth( &uTruth, 3 ); fmt.Printf("   ");
                Vec_IntPrint(p.vFanins);
                continue;
            
            } else {
                return uTruth; }
        }
        if nCexes == 64 {
            return 0; }
        Rsb_DecVerifyCex( f, pDivs, Vec_IntSize(p.vFanins), iCexA, iCexB );
        Rsb_DecRecordCex( g, nGsAll, iCexA, iCexB, pCexes, nCexes++ );
        if !p.fVerbose {
            continue; }
        Vec_IntPush( p.vTries, i );
        Vec_IntPush( p.vTries, k );
        Vec_IntPush( p.vTries, j );
        Vec_IntPush( p.vTries, -1 );
    }
    }
    }
    if p.nDecMax == 3 {
        return 0; }
    // continue by checking quadras
    p.vFanins.nSize = 4;
    for  i = 3; i < nGs; i++  {
    for  k = 2; k < i; k++  {
    for  j = 1; j < k; j++  {
    for  n = 0; n < j; n++  {
        if pCexes[i] & pCexes[k] & pCexes[j] & pCexes[n] {
            continue; }
        pDivs[0] = g[i];  p.vFanins.pArray[0] = i;
        pDivs[1] = g[k];  p.vFanins.pArray[1] = k;
        pDivs[2] = g[j];  p.vFanins.pArray[2] = j;
        pDivs[3] = g[n];  p.vFanins.pArray[3] = n;
        uTruth = Rsb_DecCheck( nVars, f, pDivs, Vec_IntSize(p.vFanins), pPat, &iCexA, &iCexB );
        if uTruth != 0 {
            if fFindAll != 0 {
                uTruth = uint32(Abc_Tt6Stretch( (word)uTruth, 4 ));
                Kit_DsdPrintFromTruth( &uTruth, 4 ); fmt.Printf("   ");
                Vec_IntPrint(p.vFanins);
                continue;
            
            } else {
                return uTruth; }
        }
        if nCexes == 64 {
            return 0; }
        Rsb_DecVerifyCex( f, pDivs, Vec_IntSize(p.vFanins), iCexA, iCexB );
        Rsb_DecRecordCex( g, nGsAll, iCexA, iCexB, pCexes, nCexes++ );
        if !p.fVerbose {
            continue; }
        Vec_IntPush( p.vTries, i );
        Vec_IntPush( p.vTries, k );
        Vec_IntPush( p.vTries, j );
        Vec_IntPush( p.vTries, n );
        Vec_IntPush( p.vTries, -1 );
    }
    }
    }
    }
//    printf( "%d ", nCexes );
    if p.nDecMax == 4 {
        return 0; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Verifies 4-input decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_DecPrintFunc(p *Rsb_Man_t, Truth4 uint32, f *word, ppGs **word, nGs int32, nVarsAll int32) {
    var nVars int32   =  Vec_IntSize(p.vFanins);
    var Copy word   =  Truth4;
    var wOn word    =  Abc_Tt6Stretch( Copy >> (1 << nVars), nVars );
    var wOnDc word  = ~Abc_Tt6Stretch( Copy, nVars );
    var wIsop word  =  Abc_Tt6Isop( wOn, wOnDc, nVars, nil );
    var i int32;

    fmt.Printf("Offset : ");
    Abc_TtPrintBinary( &Copy, nVars );  //printf( "\n" );
    Copy >>= ((word)1 << nVars);
    fmt.Printf("Onset  : ");
    Abc_TtPrintBinary( &Copy, nVars );  //printf( "\n" );
    fmt.Printf("Result : ");
    Abc_TtPrintBinary( &wIsop, nVars ); //printf( "\n" );
    Kit_DsdPrintFromTruth( (unsigned *)&wIsop, nVars );  fmt.Printf("\n");

    // print functions
    fmt.Printf("Func   : ");
    Abc_TtPrintBinary( f, nVarsAll );   //printf( "\n" );
    Kit_DsdPrintFromTruth( (unsigned *)f, nVarsAll );    fmt.Printf("\n");
    for  i = 0; i < nGs; i++  {
        fmt.Printf("Div%3d : ", i);
        Kit_DsdPrintFromTruth( (unsigned *)ppGs[i], nVarsAll );  fmt.Printf("\n");
    }
    fmt.Printf("Solution : ");
    for  i = 0; i < Vec_IntSize(p.vFanins); i++  {
        fmt.Printf("%d ", Vec_IntEntry(p.vFanins, i));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Verifies 4-input decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_DecVerify(p *Rsb_Man_t, nVars int32, f *word, g **word, nGs int32, Truth4 uint32, pTemp1 *word, pTemp2 *word) int32 {
    var pFanins [16]*word;
    var b, m, Num, nSuppSize int32;
    var nWords int32  = Abc_TtWordNum(nVars);
    Truth4 >>= (1 << Vec_IntSize(p.vFanins));
    Truth4 = uint32(Abc_Tt6Stretch( (word)Truth4, Vec_IntSize(p.vFanins) ));
//Kit_DsdPrintFromTruth( (unsigned *)&Truth4, Vec_IntSize(p->vFanins) );
//printf( "\n" );
//    nSuppSize = Rsb_Word6SupportSize( Truth4 );
//    assert( nSuppSize == Vec_IntSize(p->vFanins) );
    nSuppSize = Vec_IntSize(p.vFanins);
    // collect the fanins
    Vec_IntForEachEntry( p.vFanins, Num, b )
    {
        assert(Num < nGs, "Num < nGs");
        pFanins[b] = g[Num];
    }
    // create the or of ands
    Abc_TtClear( pTemp1, nWords );
    for  m = 0; m < (1<<nSuppSize); m++  {
        if ((Truth4 >> m) & 1) == 0 {
            continue; }
        Abc_TtFill( pTemp2, nWords );
        for  b = 0; b < nSuppSize; b++  {
            if (m >> b) & 1 {
                Abc_TtAnd( pTemp2, pTemp2, pFanins[b], nWords, 0 );
            } else {
                Abc_TtSharp( pTemp2, pTemp2, pFanins[b], nWords ); }
        }
        Abc_TtOr( pTemp1, pTemp1, pTemp2, nWords );
    }
    // check the function
    if !Abc_TtEqual( pTemp1, f, nWords ) {
        fmt.Printf("Verification failed.\n"); }
//    else
//        printf( "Verification passed.\n" );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Finds a setset of gs to decompose f.]

  Description [Returns the numbers of the decomposition functions and 
  the truth table of a function up to 4 variables.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_ManPerform(p *Rsb_Man_t, nVars int32, f *word, g **word, nGs int32, nGsAll int32, fVerbose0 int32) uint32 {
    var pCexes *word  = Vec_WrdArray(p.vCexes);
    var pPat *uint32  = (unsigned *)Vec_IntArray(p.vDecPats);
    var fVerbose int32;//(nGs > 40);
    var vTries *Vec_Int_t  = nil;
    var uTruth uint32;

    // verify original decomposition
    if Vec_IntSize(p.vFaninsOld) && Vec_IntSize(p.vFaninsOld) <= 4 {
        var pDivs [8]*word;
        var i, Entry, iCexA, iCexB int32;
        Vec_IntForEachEntry( p.vFaninsOld, Entry, i )
            pDivs[i] = g[Entry];
        uTruth = Rsb_DecCheck( nVars, f, pDivs, Vec_IntSize(p.vFaninsOld), pPat, &iCexA, &iCexB );
//        assert( uTruth != 0 );
        if fVerbose != 0 {
            fmt.Printf("Verified orig decomp with %d vars {", Vec_IntSize(p.vFaninsOld));
            Vec_IntForEachEntry( p.vFaninsOld, Entry, i )
                fmt.Printf(" %d", Entry);
            fmt.Printf(" }\n");
        }
        if uTruth != 0 {
//            if ( fVerbose )
//                Rsb_DecPrintFunc( p, uTruth );
        
        } else {
            fmt.Printf("Verified orig decomp with %d vars {", Vec_IntSize(p.vFaninsOld));
            Vec_IntForEachEntry( p.vFaninsOld, Entry, i )
                fmt.Printf(" %d", Entry);
            fmt.Printf(" }\n");
            fmt.Printf("Verification FAILED.\n");
        }
    }
    // start tries
if fVerbose != 0 {
vTries = Vec_IntAlloc( 100 ); }
    assert(nGs < nGsAll, "nGs < nGsAll");
    uTruth = Rsb_DecPerformInt( p, nVars, f, g, nGs, nGsAll, 0 );

    if uTruth != 0 {
        if fVerbose != 0 {
            var i, Entry int32;
            fmt.Printf("Found decomp with %d vars {", Vec_IntSize(p.vFanins));
            Vec_IntForEachEntry( p.vFanins, Entry, i )
                fmt.Printf(" %d", Entry);
            fmt.Printf(" }\n");
//            Rsb_DecPrintFunc( p, uTruth );
//            Rsb_DecVerify( nVars, f, g, nGs, p->vFanins, uTruth, g[nGsAll], g[nGsAll+1] );
        }
    
    } else {
        Vec_IntShrink( p.vFanins, 0 );
        if fVerbose != 0 {
        fmt.Printf("Did not find decomposition with 4 variables.\n"); }
    }

if fVerbose != 0 {
Rsb_DecPrintTable( pCexes, nGs, nGsAll, vTries ); }
if fVerbose != 0 {
Vec_IntFree( vTries ); }
    return uTruth;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_ManPerformResub6(p *Rsb_Man_t, nVarsAll int32, uTruth word, vDivTruths *Vec_Wrd_t, puTruth0 *word, puTruth1 *word, fVerbose int32) int32 {
    var pGs [200]*word;
    var uTruthRes uint32;
    var i int32;
    var nVars int32;
    var nGs int32 = Vec_WrdSize(vDivTruths);
    assert(nGs < 200, "nGs < 200");
    for  i = 0; i < nGs; i++  {
        pGs[i] = Vec_WrdEntryP(vDivTruths,i);
    }
    uTruthRes = Rsb_DecPerformInt( p, nVarsAll, &uTruth, pGs, nGs, nGs, 0 );
    if uTruthRes == 0 {
        return 0; }

    if fVerbose != 0 {
        Rsb_DecPrintFunc( p, uTruthRes, &uTruth, pGs, nGs, nVarsAll ); }
    if fVerbose != 0 {
        Rsb_DecPrintTable( Vec_WrdArray(p.vCexes), nGs, nGs, p.vTries ); }

    nVars = Vec_IntSize(p.vFanins);
    *puTruth0 = Abc_Tt6Stretch( uTruthRes,                 nVars );
    *puTruth1 = Abc_Tt6Stretch( uTruthRes >> (1 << nVars), nVars );
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_ManPerformResub6Test() {
    var p *Rsb_Man_t;
    var vDivTruths *Vec_Wrd_t;
    var RetValue int32;
    var a word  = s_Truths6[0];
    var b word  = s_Truths6[1];
    var c word  = s_Truths6[2];
    var d word  = s_Truths6[3];
    var e word  = s_Truths6[4];
    var f word  = s_Truths6[5];
    var ab word  = a & b;
    var cd word  = c & d;
    var ef word  = e & f;
    var F word  = ab | cd | ef;
    var uTruth0, uTruth1 word;

    vDivTruths = Vec_WrdAlloc( 100 );
    Vec_WrdPush( vDivTruths, a );
    Vec_WrdPush( vDivTruths, b );
    Vec_WrdPush( vDivTruths, c );
    Vec_WrdPush( vDivTruths, d );
    Vec_WrdPush( vDivTruths, e );
    Vec_WrdPush( vDivTruths, f );
    Vec_WrdPush( vDivTruths, ab );
    Vec_WrdPush( vDivTruths, cd );
    Vec_WrdPush( vDivTruths, ef );

    p = Rsb_ManAlloc( 6, 64, 4, 1 );

    RetValue = Rsb_ManPerformResub6( p, 6, F, vDivTruths, &uTruth0, &uTruth1, 1 );

    Rsb_ManFree( p );
    Vec_WrdFree( vDivTruths );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

