package dau

import "fmt"

/**CFile****************************************************************

  FileName    [dauCanon.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Canonical form computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauCanon.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/util/utilTruth.h"
//#include "misc/vec/vecMem.h"
//#include "bool/lucky/lucky.h"
//#include <math.h>

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var s_CMasks6 [5]word = [5]word{
    ABC_CONST(0x1111111111111111),
    ABC_CONST(0x0303030303030303),
    ABC_CONST(0x000F000F000F000F),
    ABC_CONST(0x000000FF000000FF),
    ABC_CONST(0x000000000000FFFF),
};

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Compares Cof0 and Cof1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
static inline int Abc_TtCompare1VarCofs( word * pTruth, int nWords, int iVar )
{
    if ( nWords == 1 )
    {
        word Cof0 = pTruth[0] & s_Truths6Neg[iVar];
        word Cof1 = (pTruth[0] >> (1 << iVar)) & s_Truths6Neg[iVar];
        if ( Cof0 != Cof1 )
            return Cof0 < Cof1 ? -1 : 1;
        return 0;
    }
    if ( iVar <= 5 )
    {
        word Cof0, Cof1;
        int w, shift = (1 << iVar);
        for ( w = 0; w < nWords; w++ )
        {
            Cof0 = pTruth[w] & s_Truths6Neg[iVar];
            Cof1 = (pTruth[w] >> shift) & s_Truths6Neg[iVar];
            if ( Cof0 != Cof1 )
                return Cof0 < Cof1 ? -1 : 1;
        }
        return 0;
    }
    // if ( iVar > 5 )
    {
        word * pLimit = pTruth + nWords;
        int i, iStep = Abc_TtWordNum(iVar);
        assert( nWords >= 2 );
        for ( ; pTruth < pLimit; pTruth += 2*iStep )
            for ( i = 0; i < iStep; i++ )
                if ( pTruth[i] != pTruth[i + iStep] )
                    return pTruth[i] < pTruth[i + iStep] ? -1 : 1;
        return 0;
    }   
}
static inline int Abc_TtCompare1VarCofsRev( word * pTruth, int nWords, int iVar )
{
    if ( nWords == 1 )
    {
        word Cof0 = pTruth[0] & s_Truths6Neg[iVar];
        word Cof1 = (pTruth[0] >> (1 << iVar)) & s_Truths6Neg[iVar];
        if ( Cof0 != Cof1 )
            return Cof0 < Cof1 ? -1 : 1;
        return 0;
    }
    if ( iVar <= 5 )
    {
        word Cof0, Cof1;
        int w, shift = (1 << iVar);
        for ( w = nWords - 1; w >= 0; w-- )
        {
            Cof0 = pTruth[w] & s_Truths6Neg[iVar];
            Cof1 = (pTruth[w] >> shift) & s_Truths6Neg[iVar];
            if ( Cof0 != Cof1 )
                return Cof0 < Cof1 ? -1 : 1;
        }
        return 0;
    }
    // if ( iVar > 5 )
    {
        word * pLimit = pTruth + nWords;
        int i, iStep = Abc_TtWordNum(iVar);
        assert( nWords >= 2 );
        for ( pLimit -= 2*iStep; pLimit >= pTruth; pLimit -= 2*iStep )
            for ( i = iStep - 1; i >= 0; i-- )
                if ( pLimit[i] != pLimit[i + iStep] )
                    return pLimit[i] < pLimit[i + iStep] ? -1 : 1;
        return 0;
    }   
}
*/

/**Function*************************************************************

  Synopsis    [Checks equality of pairs of cofactors w.r.t. adjacent variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCheckEqual2VarCofs(pTruth *word, nWords int32, iVar int32, Num1 int32, Num2 int32) int32 {
    assert(Num1 < Num2 && Num2 < 4, "Num1 < Num2 && Num2 < 4");
    if nWords == 1 {
        return ((pTruth[0] >> (Num2 * (1 << iVar))) & s_CMasks6[iVar]) == ((pTruth[0] >> (Num1 * (1 << iVar))) & s_CMasks6[iVar]); }
    if iVar <= 4 {
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            if ((pTruth[w] >> Num2 * shift) & s_CMasks6[iVar]) != ((pTruth[w] >> Num1 * shift) & s_CMasks6[iVar]) {
                return 0; }
        }
        return 1;
    }
    if iVar == 5 {
        var pTruthU *uint32  = (unsigned *)pTruth;
        var pLimitU *uint32  = (unsigned *)(pTruth + nWords);
        assert(nWords >= 2, "nWords >= 2");
        for  ; pTruthU < pLimitU; pTruthU += 4  {
            if pTruthU[Num2] != pTruthU[Num1] {
                return 0; }
        }
        return 1;
    }
    // if ( iVar > 5 )
    {
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        assert(nWords >= 4, "nWords >= 4");
        for  ; pTruth < pLimit; pTruth += 4*iStep  {
            for  i = 0; i < iStep; i++  {
                if pTruth[i+Num2*iStep] != pTruth[i+Num1*iStep] {
                    return 0; }
            }
        }
        return 1;
    }   
}

/**Function*************************************************************

  Synopsis    [Compares pairs of cofactors w.r.t. adjacent variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCompare2VarCofs(pTruth *word, nWords int32, iVar int32, Num1 int32, Num2 int32) int32 {
    assert(Num1 < Num2 && Num2 < 4, "Num1 < Num2 && Num2 < 4");
    if nWords == 1 {
        var Cof1 word  = (pTruth[0] >> (Num1 * (1 << iVar))) & s_CMasks6[iVar];
        var Cof2 word  = (pTruth[0] >> (Num2 * (1 << iVar))) & s_CMasks6[iVar];
        if Cof1 != Cof2 {
            if Cof1 < Cof2 {
                return -1;
            } else {
                return 1;
            } }
        return 0;
    }
    if iVar <= 4 {
        var Cof1, Cof2 word;
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            Cof1 = (pTruth[w] >> Num1 * shift) & s_CMasks6[iVar];
            Cof2 = (pTruth[w] >> Num2 * shift) & s_CMasks6[iVar];
            if Cof1 != Cof2 {
                if Cof1 < Cof2 {
                    return -1;
                } else {
                    return 1;
                } }
        }
        return 0;
    }
    if iVar == 5 {
        var pTruthU *uint32  = (unsigned *)pTruth;
        var pLimitU *uint32  = (unsigned *)(pTruth + nWords);
        assert(nWords >= 2, "nWords >= 2");
        for  ; pTruthU < pLimitU; pTruthU += 4  {
            if pTruthU[Num1] != pTruthU[Num2] {
                if pTruthU[Num1] < pTruthU[Num2] {
                    return -1;
                } else {
                    return 1;
                } }
        }
        return 0;
    }
    // if ( iVar > 5 )
    {
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        var Offset1 int32  = Num1*iStep;
        var Offset2 int32  = Num2*iStep;
        assert(nWords >= 4, "nWords >= 4");
        for  ; pTruth < pLimit; pTruth += 4*iStep  {
            for  i = 0; i < iStep; i++  {
                if pTruth[i + Offset1] != pTruth[i + Offset2] {
                    if pTruth[i + Offset1] < pTruth[i + Offset2] {
                        return -1;
                    } else {
                        return 1;
                    } }
            }
        }
        return 0;
    }   
}
func Abc_TtCompare2VarCofsRev(pTruth *word, nWords int32, iVar int32, Num1 int32, Num2 int32) int32 {
    assert(Num1 < Num2 && Num2 < 4, "Num1 < Num2 && Num2 < 4");
    if nWords == 1 {
        var Cof1 word  = (pTruth[0] >> (Num1 * (1 << iVar))) & s_CMasks6[iVar];
        var Cof2 word  = (pTruth[0] >> (Num2 * (1 << iVar))) & s_CMasks6[iVar];
        if Cof1 != Cof2 {
            if Cof1 < Cof2 {
                return -1;
            } else {
                return 1;
            } }
        return 0;
    }
    if iVar <= 4 {
        var Cof1, Cof2 word;
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = nWords - 1; w >= 0; w--  {
            Cof1 = (pTruth[w] >> Num1 * shift) & s_CMasks6[iVar];
            Cof2 = (pTruth[w] >> Num2 * shift) & s_CMasks6[iVar];
            if Cof1 != Cof2 {
                if Cof1 < Cof2 {
                    return -1;
                } else {
                    return 1;
                } }
        }
        return 0;
    }
    if iVar == 5 {
        var pTruthU *uint32  = (unsigned *)pTruth;
        var pLimitU *uint32  = (unsigned *)(pTruth + nWords);
        assert(nWords >= 2, "nWords >= 2");
        for  pLimitU -= 4; pLimitU >= pTruthU; pLimitU -= 4  {
            if pLimitU[Num1] != pLimitU[Num2] {
                if pLimitU[Num1] < pLimitU[Num2] {
                    return -1;
                } else {
                    return 1;
                } }
        }
        return 0;
    }
    // if ( iVar > 5 )
    {
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        var Offset1 int32  = Num1*iStep;
        var Offset2 int32  = Num2*iStep;
        assert(nWords >= 4, "nWords >= 4");
        for  pLimit -= 4*iStep; pLimit >= pTruth; pLimit -= 4*iStep  {
            for  i = iStep - 1; i >= 0; i--  {
                if pLimit[i + Offset1] != pLimit[i + Offset2] {
                    if pLimit[i + Offset1] < pLimit[i + Offset2] {
                        return -1;
                    } else {
                        return 1;
                    } }
            }
        }
        return 0;
    }   
}

/**Function*************************************************************

  Synopsis    [Minterm counting in all cofactors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtNormalizeSmallTruth(pTruth *word, nVars int32) {
    if nVars < 6 {
        var shift int32;
        var bits int32 = (1 << nVars);
        var base word  = *pTruth = *pTruth & ((((word)1) << bits) - 1);
        for shift = bits; shift < 64; shift += bits {
            *pTruth |= base << shift;
        }
    }
}

func Abc_TtVerifySmallTruth(pTruth *word, nVars int32) {
#ifndef NDEBUG
    if (nVars < 6) {
        word nTruth = *pTruth;
        Abc_TtNormalizeSmallTruth(&nTruth, nVars);
        assert(*pTruth == nTruth);
    }
#endif
}

func Abc_TtCountOnesInTruth(pTruth *word, nVars int32) int32 {   
    var nWords int32  = Abc_TtWordNum( nVars );
    var k, Counter int32;
    Abc_TtVerifySmallTruth(pTruth, nVars);
    for  k = 0; k < nWords; k++  {
        if pTruth[k] {
            Counter += Abc_TtCountOnes( pTruth[k] ); }
    }
    return Counter;
}
func Abc_TtCountOnesInCofs(pTruth *word, nVars int32, pStore *int32) {
    var Temp word;
    var i, k, Counter, nWords int32;
    if nVars <= 6 {
        Abc_TtVerifySmallTruth(pTruth, nVars);
        for  i = 0; i < nVars; i++  {
            pStore[i] = Abc_TtCountOnes( pTruth[0] & s_Truths6Neg[i] );
        }
        return;
    }
    assert(nVars > 6, "nVars > 6");
    nWords = Abc_TtWordNum( nVars );
    memset( pStore, 0, sizeof(int) * nVars );
    for  k = 0; k < nWords; k++  {
        // count 1's for the first six variables
        for  i = 0; i < 6; i++  {
            if Temp = (pTruth[k] & s_Truths6Neg[i]) | ((pTruth[k+1] & s_Truths6Neg[i]) << (1 << i)) {
                pStore[i] += Abc_TtCountOnes( Temp ); }
        }
        // count 1's for all other variables
        if pTruth[k] {
            Counter = Abc_TtCountOnes( pTruth[k] );
            for  i = 6; i < nVars; i++  {
                if (k & (1 << (i-6))) == 0 {
                    pStore[i] += Counter; }
            }
        }
        k++;
        // count 1's for all other variables
        if pTruth[k] {
            Counter = Abc_TtCountOnes( pTruth[k] );
            for  i = 6; i < nVars; i++  {
                if (k & (1 << (i-6))) == 0 {
                    pStore[i] += Counter; }
            }
        }
    } 
}
func Abc_TtCountOnesInCofsSimple(pTruth *word, nVars int32, pStore *int32) int32 {
    Abc_TtCountOnesInCofs( pTruth, nVars, pStore );
    return Abc_TtCountOnesInTruth( pTruth, nVars );
}

// Shifted Cofactor Coefficient
static inline int shiftFunc(int ci)
//{ return ci * ci; }
{ return 1 << ci; }

func Abc_TtScc6(wTruth word, ck int32) int32 {
    var i int32;
    var sum int32;
    if !wTruth { return 0; }
    for i = 0; i < 64; i++ {
        if wTruth & (word)1 << i {
            var ci int32  = Abc_TtBitCount8[i] + ck;
            sum += shiftFunc(ci);
        }
    }
    return sum;
}
func Abc_TtScc(pTruth *word, nVars int32) int32 {
    var k int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    var sum int32;
    Abc_TtNormalizeSmallTruth(pTruth, nVars);
    for k = 0; k < nWords; k++ {
        sum += Abc_TtScc6(pTruth[k], Abc_TtBitCount16(k));
    }
    return sum;
}
func Abc_TtSccInCofs6(wTruth word, nVars int32, ck int32, pStore *int32) {
    var i, j, v int32;
    assert(nVars <= 6, "nVars <= 6");
    for v = 0; v < nVars; v++ {
        var sum int32;
        for i = j = 0; j < 64; j++ {
            if s_Truths6Neg[v] & (word)1 << j {
                if wTruth & (word)1 << j {
                    var ci int32  = Abc_TtBitCount8[i] + ck;
                    sum += shiftFunc(ci);
                }
                i++;
            }
        }
        pStore[v] += sum;
    }
}
func Abc_TtSccInCofs(pTruth *word, nVars int32, pStore *int32) {
    var k, v, nWords int32;
    var kv [10]int32 ;
    memset(pStore, 0, sizeof(int) * nVars);
    if nVars <= 6 {
        Abc_TtNormalizeSmallTruth(pTruth, nVars);
        Abc_TtSccInCofs6(pTruth[0], nVars, 0, pStore);
        return;
    }
    assert(nVars > 6, "nVars > 6");
    nWords = Abc_TtWordNum(nVars);
    for k = 0; k < nWords; k++ {
        // count 1's for the first six variables
        Abc_TtSccInCofs6(pTruth[k], 6, Abc_TtBitCount16(k), pStore);

        // count 1's for all other variables
        for v = 6; v < nVars; v++ {
            if (k & (1 << (v - 6))) == 0 {
                pStore[v] += Abc_TtScc6(pTruth[k], Abc_TtBitCount16(kv[v - 6]));
                kv[v - 6]++;
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Minterm counting in all cofactors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_TtCountOnesInCofsSlow_bit_count [256]int32 = [256]int32{
      0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
      1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
      1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
      2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
      1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
      2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
      2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
      3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
    };
func Abc_TtCountOnesInCofsSlow(pTruth *word, nVars int32, pStore *int32) {
    
    var i, k, nBytes int32;
    unsigned char * pTruthC = (unsigned char *)pTruth;
    nBytes = 8 * Abc_TtWordNum( nVars );
    memset( pStore, 0, sizeof(int) * nVars );
    for  k = 0; k < nBytes; k++  {
        pStore[0] += spStatic_Abc_TtCountOnesInCofsSlow_bit_count[ pTruthC[k] & 0x55 ];
        pStore[1] += spStatic_Abc_TtCountOnesInCofsSlow_bit_count[ pTruthC[k] & 0x33 ];
        pStore[2] += spStatic_Abc_TtCountOnesInCofsSlow_bit_count[ pTruthC[k] & 0x0F ];
        for  i = 3; i < nVars; i++  {
            if (k & (1 << (i-3))) == 0 {
                pStore[i] += spStatic_Abc_TtCountOnesInCofsSlow_bit_count[pTruthC[k]]; }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Minterm counting in all cofactors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_TtCountOnesInCofsFast6_rec_bit_count [256]int32 = [256]int32{
      0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
      1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
      1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
      2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
      1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
      2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
      2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
      3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
    };
func Abc_TtCountOnesInCofsFast6_rec(Truth word, iVar int32, nBytes int32, pStore *int32) int32 {
    
    var nMints0, nMints1 int32;
    if Truth == 0 {
        return 0; }
    if ~Truth == 0 {
        var i int32;
        for  i = 0; i <= iVar; i++  {
            pStore[i] += nBytes * 4;
        }
        return nBytes * 8;
    }
    if nBytes == 1 {
        assert(iVar == 2, "iVar == 2");
        pStore[0] += spStatic_Abc_TtCountOnesInCofsFast6_rec_bit_count[ Truth & 0x55 ];
        pStore[1] += spStatic_Abc_TtCountOnesInCofsFast6_rec_bit_count[ Truth & 0x33 ];
        pStore[2] += spStatic_Abc_TtCountOnesInCofsFast6_rec_bit_count[ Truth & 0x0F ];
        return spStatic_Abc_TtCountOnesInCofsFast6_rec_bit_count[ Truth & 0xFF ];
    }
    nMints0 = Abc_TtCountOnesInCofsFast6_rec( Abc_Tt6Cofactor0(Truth, iVar), iVar - 1, nBytes/2, pStore );
    nMints1 = Abc_TtCountOnesInCofsFast6_rec( Abc_Tt6Cofactor1(Truth, iVar), iVar - 1, nBytes/2, pStore );
    pStore[iVar] += nMints0;
    return nMints0 + nMints1;
}

func Abc_TtCountOnesInCofsFast_rec(pTruth *word, iVar int32, nWords int32, pStore *int32) int32 {
    var nMints0, nMints1 int32;
    if nWords == 1 {
        assert(iVar == 5, "iVar == 5");
        return Abc_TtCountOnesInCofsFast6_rec( pTruth[0], iVar, 8, pStore );
    }
    assert(nWords > 1, "nWords > 1");
    assert(iVar > 5, "iVar > 5");
    if pTruth[0] & 1 {
        if Abc_TtIsConst1( pTruth, nWords ) {
            var i int32;
            for  i = 0; i <= iVar; i++  {
                pStore[i] += nWords * 32;
            }
            return nWords * 64;
        }
    
    } else {
        if Abc_TtIsConst0( pTruth, nWords ) {
            return 0; }
    }
    nMints0 = Abc_TtCountOnesInCofsFast_rec( pTruth,            iVar - 1, nWords/2, pStore );
    nMints1 = Abc_TtCountOnesInCofsFast_rec( pTruth + nWords/2, iVar - 1, nWords/2, pStore );
    pStore[iVar] += nMints0;
    return nMints0 + nMints1;
}
func Abc_TtCountOnesInCofsFast(pTruth *word, nVars int32, pStore *int32) int32 {
    memset( pStore, 0, sizeof(int) * nVars );
    assert(nVars >= 3, "nVars >= 3");
    if nVars <= 6 {
        return Abc_TtCountOnesInCofsFast6_rec( pTruth[0], nVars - 1, Abc_TtByteNum( nVars ), pStore );
    } else {
        return Abc_TtCountOnesInCofsFast_rec( pTruth, nVars - 1, Abc_TtWordNum( nVars ), pStore ); }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtSemiCanonicize(pTruth *word, nVars int32, pCanonPerm string, pStoreOut *int32, fOnlySwap int32) uint32 {
    var fUseOld int32  = 1;
    var fOldSwap int32;
    var pStoreIn [17]int32;
    var pStore *int32;
    if pStoreOut != nil {
        pStore = pStoreOut;
    } else {
        pStore = pStoreIn;
    }
    var i int32;
    var nOnes int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    var uCanonPhase uint32;
    assert(nVars <= 16, "nVars <= 16");
    for  i = 0; i < nVars; i++  {
        pCanonPerm[i] = i;
    }

    if fUseOld != 0 {
        // normalize polarity    
        nOnes = Abc_TtCountOnesInTruth( pTruth, nVars );
        if nOnes > nWords * 32 && !(fOnlySwap != 0) {
            Abc_TtNot( pTruth, nWords );
            nOnes = nWords*64 - nOnes;
            uCanonPhase |= (1 << nVars);
        }
        // normalize phase
        Abc_TtCountOnesInCofs( pTruth, nVars, pStore );
        pStore[nVars] = nOnes;
        for  i = 0; i < nVars; i++  {
            if pStore[i] >= nOnes - pStore[i] || (fOnlySwap != 0) {
                continue; }
            Abc_TtFlip( pTruth, nWords, i );
            uCanonPhase |= (1 << i);
            pStore[i] = nOnes - pStore[i]; 
        }
    
    } else {
        nOnes = Abc_TtCountOnesInCofsQuick( pTruth, nVars, pStore );
        // normalize polarity    
        if nOnes > nWords * 32 && !(fOnlySwap != 0) {
            for  i = 0; i < nVars; i++  {
                pStore[i] = nWords * 32 - pStore[i];
            } 
            Abc_TtNot( pTruth, nWords );
            nOnes = nWords*64 - nOnes;
            uCanonPhase |= (1 << nVars);
        }
        // normalize phase
        pStore[nVars] = nOnes;
        for  i = 0; i < nVars; i++  {
            if pStore[i] >= nOnes - pStore[i] || (fOnlySwap != 0) {
                continue; }
            Abc_TtFlip( pTruth, nWords, i );
            uCanonPhase |= (1 << i);
            pStore[i] = nOnes - pStore[i]; 
        }
    }

    // normalize permutation
    if fOldSwap != 0 {
        var fChange int32;
        do {
            fChange = 0;
            for  i = 0; i < nVars-1; i++  {
                if pStore[i] <= pStore[i+1] {
    //            if ( pStore[i] >= pStore[i+1] )
                    continue; }
                // inlining of macro ABC_SWAP
                { var t int32  = pCanonPerm[i]; pCanonPerm[i] = pCanonPerm[i+1]; pCanonPerm[i+1] = t; };
                // inlining of macro ABC_SWAP
                { var t int32  = pStore[i]; pStore[i] = pStore[i+1]; pStore[i+1] = t; };
                if ((uCanonPhase >> i) & 1) != ((uCanonPhase >> (i+1)) & 1) {
                    uCanonPhase ^= (1 << i);
                    uCanonPhase ^= (1 << (i+1));
                }
                Abc_TtSwapAdjacent( pTruth, nWords, i );            
                fChange = 1;
    //            nSwaps++;
            }
        } 
        while ( fChange );
    
    } else {
        var k, BestK int32;
        for  i = 0; i < nVars - 1; i++  {
            BestK = i + 1;
            for  k = i + 2; k < nVars; k++  {
                if pStore[BestK] > pStore[k] {
    //            if ( pStore[BestK] < pStore[k] )
                    BestK = k; }
            }
            if pStore[i] <= pStore[BestK] {
    //        if ( pStore[i] >= pStore[BestK] )
                continue; }
            // inlining of macro ABC_SWAP
            { var t int32  = pCanonPerm[i]; pCanonPerm[i] = pCanonPerm[BestK]; pCanonPerm[BestK] = t; };
            // inlining of macro ABC_SWAP
            { var t int32  = pStore[i]; pStore[i] = pStore[BestK]; pStore[BestK] = t; };
            if ((uCanonPhase >> i) & 1) != ((uCanonPhase >> BestK) & 1) {
                uCanonPhase ^= (1 << i);
                uCanonPhase ^= (1 << BestK);
            }
            Abc_TtSwapVars( pTruth, nVars, i, BestK );
    //        nSwaps++;
        }
    }
    return uCanonPhase;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_TtCofactorTest10_pCopy1 [1024]word;
var spStatic_Abc_TtCofactorTest10_pCopy2 [1024]word;
func Abc_TtCofactorTest10(pTruth *word, nVars int32, N int32) {
    
    
    var nWords int32  = Abc_TtWordNum( nVars );
    var i int32;
    for  i = 0; i < nVars - 1; i++  {
//        Kit_DsdPrintFromTruth( pTruth, nVars ); printf( "\n" );
        Abc_TtCopy( spStatic_Abc_TtCofactorTest10_pCopy1, pTruth, nWords, 0 );
        Abc_TtSwapAdjacent( spStatic_Abc_TtCofactorTest10_pCopy1, nWords, i );
//        Kit_DsdPrintFromTruth( pCopy1, nVars ); printf( "\n" );
        Abc_TtCopy( spStatic_Abc_TtCofactorTest10_pCopy2, pTruth, nWords, 0 );
        Abc_TtSwapVars( spStatic_Abc_TtCofactorTest10_pCopy2, nVars, i, i+1 );
//        Kit_DsdPrintFromTruth( pCopy2, nVars ); printf( "\n" );
        assert(Abc_TtEqual( spStatic_Abc_TtCofactorTest10_pCopy1, spStatic_Abc_TtCofactorTest10_pCopy2, nWords ), "Abc_TtEqual( spStatic_Abc_TtCofactorTest10_pCopy1, spStatic_Abc_TtCofactorTest10_pCopy2, nWords )");
    }
}

/**Function*************************************************************

  Synopsis    [Naive evaluation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6CofactorPermNaive(pTruth *word, i int32, fSwapOnly int32) int32 {
    if fSwapOnly != 0 {
        var Copy word  = Abc_Tt6SwapAdjacent( pTruth[0], i );
        if pTruth[0] > Copy {
            pTruth[0] = Copy;
            return 4;
        }
        return 0;
    }
    {
        var Copy word  = pTruth[0];
        var Best word  = pTruth[0];
        var Config int32;
        // PXY
        // 001
        Copy = Abc_Tt6Flip( Copy, i );
        if Best > Copy {
            Best = Copy;
Config = 1; }
        // PXY
        // 011
        Copy = Abc_Tt6Flip( Copy, i+1 );
        if Best > Copy {
            Best = Copy;
Config = 3; }
        // PXY
        // 010
        Copy = Abc_Tt6Flip( Copy, i );
        if Best > Copy {
            Best = Copy;
Config = 2; }
        // PXY
        // 110
        Copy = Abc_Tt6SwapAdjacent( Copy, i );
        if Best > Copy {
            Best = Copy;
Config = 6; }
        // PXY
        // 111
        Copy = Abc_Tt6Flip( Copy, i+1 );
        if Best > Copy {
            Best = Copy;
Config = 7; }
        // PXY
        // 101
        Copy = Abc_Tt6Flip( Copy, i );
        if Best > Copy {
            Best = Copy;
Config = 5; }
        // PXY
        // 100
        Copy = Abc_Tt6Flip( Copy, i+1 );
        if Best > Copy {
            Best = Copy;
Config = 4; }
        // PXY
        // 000
        Copy = Abc_Tt6SwapAdjacent( Copy, i );
        assert(Copy == pTruth[0], "Copy == pTruth[0]");
        assert(Best <= pTruth[0], "Best <= pTruth[0]");
        pTruth[0] = Best;
        return Config;
    }
}
var spStatic_Abc_TtCofactorPermNaive_pCopy_ [1024]word;
var spStatic_Abc_TtCofactorPermNaive_pCopy_ [1024]word;
var spStatic_Abc_TtCofactorPermNaive_pBest [1024]word;
func Abc_TtCofactorPermNaive(pTruth *word, i int32, nWords int32, fSwapOnly int32) int32 {
    if fSwapOnly != 0 {
        
        Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pCopy_, pTruth, nWords, 0 );
        Abc_TtSwapAdjacent( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i );
        if Abc_TtCompareRev(pTruth, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords) == 1 {
            Abc_TtCopy( pTruth, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, 0 );
            return 4;
        }
        return 0;
    }
    {
        
        
        var Config int32;
        // save two copies
        Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pCopy_, pTruth, nWords, 0 );
        Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pBest, pTruth, nWords, 0 );
        // PXY
        // 001
        Abc_TtFlip( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i );
        if Abc_TtCompareRev(spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords) == 1 {
            Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, 0 );
Config = 1; }
        // PXY
        // 011
        Abc_TtFlip( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i+1 );
        if Abc_TtCompareRev(spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords) == 1 {
            Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, 0 );
Config = 3; }
        // PXY
        // 010
        Abc_TtFlip( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i );
        if Abc_TtCompareRev(spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords) == 1 {
            Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, 0 );
Config = 2; }
        // PXY
        // 110
        Abc_TtSwapAdjacent( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i );
        if Abc_TtCompareRev(spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords) == 1 {
            Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, 0 );
Config = 6; }
        // PXY
        // 111
        Abc_TtFlip( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i+1 );
        if Abc_TtCompareRev(spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords) == 1 {
            Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, 0 );
Config = 7; }
        // PXY
        // 101
        Abc_TtFlip( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i );
        if Abc_TtCompareRev(spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords) == 1 {
            Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, 0 );
Config = 5; }
        // PXY
        // 100
        Abc_TtFlip( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i+1 );
        if Abc_TtCompareRev(spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords) == 1 {
            Abc_TtCopy( spStatic_Abc_TtCofactorPermNaive_pBest, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, 0 );
Config = 4; }
        // PXY
        // 000
        Abc_TtSwapAdjacent( spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords, i );
        assert(Abc_TtEqual( pTruth, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords ), "Abc_TtEqual( pTruth, spStatic_Abc_TtCofactorPermNaive_pCopy_, nWords )");
        if Config == 0 {
            return 0; }
        assert(Abc_TtCompareRev(pTruth, spStatic_Abc_TtCofactorPermNaive_pBest, nWords) == 1, "Abc_TtCompareRev(pTruth, spStatic_Abc_TtCofactorPermNaive_pBest, nWords) == 1");
        Abc_TtCopy( pTruth, spStatic_Abc_TtCofactorPermNaive_pBest, nWords, 0 );
        return Config;
    }
}


/**Function*************************************************************

  Synopsis    [Smart evaluation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCofactorPermConfig(pTruth *word, i int32, nWords int32, fSwapOnly int32, fNaive int32) int32 {
    if nWords == 1 {
        return Abc_Tt6CofactorPermNaive( pTruth, i, fSwapOnly ); }
    if fNaive != 0 {
        return Abc_TtCofactorPermNaive( pTruth, i, nWords, fSwapOnly ); }
    if fSwapOnly != 0 {
        if Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 1, 2 ) < 0 { // Cof1 < Cof2
        
            Abc_TtSwapAdjacent( pTruth, nWords, i );
            return 4;
        }
        return 0;
    }
    {  
        var fComp01, fComp02, fComp03, fComp12, fComp13, fComp23, Config int32;
        fComp01 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 0, 1 );
        fComp23 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 2, 3 );
        if fComp23 >= 0 { // Cof2 >= Cof3 
        
            if fComp01 >= 0 { // Cof0 >= Cof1
            
                fComp13 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 1, 3 );
                if fComp13 < 0 { // Cof1 < Cof3 
                    Abc_TtFlip( pTruth, nWords, i + 1 );
Config = 2;
                } else if fComp13 == 0 { // Cof1 == Cof3 
                
                    fComp02 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 0, 2 );
                    if fComp02 < 0 {
                        Abc_TtFlip( pTruth, nWords, i + 1 );
Config = 2; }
                }
                // else   Cof1 > Cof3 -- do nothing
            
            } else { // Cof0 < Cof1
            
                fComp03 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 0, 3 );
                if fComp03 < 0 { // Cof0 < Cof3
                
                    Abc_TtFlip( pTruth, nWords, i );
                    Abc_TtFlip( pTruth, nWords, i + 1 );
Config = 3;
                
                } else { //  Cof0 >= Cof3
                
                    if fComp23 == 0 { // can flip Cof0 and Cof1
                        Abc_TtFlip( pTruth, nWords, i );
Config = 1; }
                }
            }
        
        } else { // Cof2 < Cof3 
        
            if fComp01 >= 0 { // Cof0 >= Cof1
            
                fComp12 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 1, 2 );
                if fComp12 > 0 { // Cof1 > Cof2 
                    Abc_TtFlip( pTruth, nWords, i );
Config = 1;
                } else if fComp12 == 0 { // Cof1 == Cof2 
                
                    Abc_TtFlip( pTruth, nWords, i );
                    Abc_TtFlip( pTruth, nWords, i + 1 );
Config = 3;
                
                } else { // Cof1 < Cof2
                
                    Abc_TtFlip( pTruth, nWords, i + 1 );
Config = 2;
                    if fComp01 == 0 {
                        Abc_TtFlip( pTruth, nWords, i );
Config ^= 1; }
                }
            
            } else { // Cof0 < Cof1
            
                fComp02 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 0, 2 );
                if fComp02 == -1 { // Cof0 < Cof2 
                
                    Abc_TtFlip( pTruth, nWords, i );
                    Abc_TtFlip( pTruth, nWords, i + 1 );
Config = 3;
                
                } else if fComp02 == 0 { // Cof0 == Cof2
                
                    fComp13 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 1, 3 );
                    if fComp13 >= 0 { // Cof1 >= Cof3 
                        Abc_TtFlip( pTruth, nWords, i );
Config = 1;
                    } else { // Cof1 < Cof3 
                    
                        Abc_TtFlip( pTruth, nWords, i );
                        Abc_TtFlip( pTruth, nWords, i + 1 );
Config = 3;
                    }
                
                } else { // Cof0 > Cof2
                    Abc_TtFlip( pTruth, nWords, i );
Config = 1; }
            }
        }
        // perform final swap if needed
        fComp12 = Abc_TtCompare2VarCofsRev( pTruth, nWords, i, 1, 2 );
        if fComp12 < 0 { // Cof1 < Cof2
            Abc_TtSwapAdjacent( pTruth, nWords, i );
Config ^= 4; }
        return Config;
    }
}
var spStatic_Abc_TtCofactorPerm_pCopy1 [1024]word;
func Abc_TtCofactorPerm(pTruth *word, i int32, nWords int32, fSwapOnly int32, pCanonPerm string, puCanonPhase *uint32, fNaive int32) int32 {
    if fSwapOnly != 0 {
        var Config int32  = Abc_TtCofactorPermConfig( pTruth, i, nWords, 1, 0 );
        if Config != 0 {
            if ((*puCanonPhase >> i) & 1) != ((*puCanonPhase >> (i+1)) & 1) {
                *puCanonPhase ^= (1 << i);
                *puCanonPhase ^= (1 << (i+1));
            }
            // inlining of macro ABC_SWAP
            { var t int32  = pCanonPerm[i]; pCanonPerm[i] = pCanonPerm[i+1]; pCanonPerm[i+1] = t; };
        }
        return Config;
    }
    {
        
        var Config int32;
        Abc_TtCopy( spStatic_Abc_TtCofactorPerm_pCopy1, pTruth, nWords, 0 );
        Config = Abc_TtCofactorPermConfig( pTruth, i, nWords, 0, fNaive );
        if Config == 0 {
            return 0; }
        if Abc_TtCompareRev(pTruth, spStatic_Abc_TtCofactorPerm_pCopy1, nWords) == 1 { // made it worse
        
            Abc_TtCopy( pTruth, spStatic_Abc_TtCofactorPerm_pCopy1, nWords, 0 );
            return 0;
        }
        // improved
        if Config & 1 {
            *puCanonPhase ^= (1 << i); }
        if Config & 2 {
            *puCanonPhase ^= (1 << (i+1)); }
        if Config & 4 {
            if ((*puCanonPhase >> i) & 1) != ((*puCanonPhase >> (i+1)) & 1) {
                *puCanonPhase ^= (1 << i);
                *puCanonPhase ^= (1 << (i+1));
            }
            // inlining of macro ABC_SWAP
            { var t int32  = pCanonPerm[i]; pCanonPerm[i] = pCanonPerm[i+1]; pCanonPerm[i+1] = t; };
        }
        return Config;
    }
}

/**Function*************************************************************

  Synopsis    [Semi-canonical form computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
#define CANON_VERIFY
var spStatic_Abc_TtCanonicize_pCopy1 [1024]word;
var spStatic_Abc_TtCanonicize_pCopy2 [1024]word;
func Abc_TtCanonicize(pTruth *word, nVars int32, pCanonPerm string) uint32 {
    var pStoreIn [17]int32;
    var uCanonPhase uint32;
    var i int32;
    var k int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    var fNaive int32  = 1;

                   
    var pCanonPermCopy [16]byte;
    
    
    Abc_TtCopy( spStatic_Abc_TtCanonicize_pCopy1, pTruth, nWords, 0 );
      

    uCanonPhase = Abc_TtSemiCanonicize( pTruth, nVars, pCanonPerm, pStoreIn, 0 );
    for  k = 0; k < 5; k++  {
        var fChanges int32;
        for  i = nVars - 2; i >= 0; i--  {
            if pStoreIn[i] == pStoreIn[i+1] {
                fChanges |= Abc_TtCofactorPerm( pTruth, i, nWords, pStoreIn[i] != pStoreIn[nVars]/2, pCanonPerm, &uCanonPhase, fNaive ); }
        }
        if !(fChanges != 0) {
            break; }
        fChanges = 0;
        for  i = 1; i < nVars - 1; i++  {
            if pStoreIn[i] == pStoreIn[i+1] {
                fChanges |= Abc_TtCofactorPerm( pTruth, i, nWords, pStoreIn[i] != pStoreIn[nVars]/2, pCanonPerm, &uCanonPhase, fNaive ); }
        }
        if !(fChanges != 0) {
            break; }
    }

                   
    Abc_TtCopy( spStatic_Abc_TtCanonicize_pCopy2, pTruth, nWords, 0 );
    memcpy( pCanonPermCopy, pCanonPerm, sizeof(char) * nVars );
    Abc_TtImplementNpnConfig( spStatic_Abc_TtCanonicize_pCopy2, nVars, pCanonPermCopy, uCanonPhase );
    if !Abc_TtEqual( spStatic_Abc_TtCanonicize_pCopy1, spStatic_Abc_TtCanonicize_pCopy2, nWords ) {
        fmt.Printf("Canonical form verification failed!\n"); }
      
/*
    if ( !Abc_TtEqual( pCopy1, pCopy2, nWords ) )
    {
        Kit_DsdPrintFromTruth( pCopy1, nVars ); printf( "\n" );
        Kit_DsdPrintFromTruth( pCopy2, nVars ); printf( "\n" );
        i = 0;
    }
*/
    return uCanonPhase;
}

var spStatic_Abc_TtCanonicizePerm_pCopy1 [1024]word;
var spStatic_Abc_TtCanonicizePerm_pCopy2 [1024]word;
func Abc_TtCanonicizePerm(pTruth *word, nVars int32, pCanonPerm string) uint32 {
    var pStoreIn [17]int32;
    var uCanonPhase uint32;
    var i int32;
    var k int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    var fNaive int32  = 1;

                   
    var pCanonPermCopy [16]byte;
    
    
    Abc_TtCopy( spStatic_Abc_TtCanonicizePerm_pCopy1, pTruth, nWords, 0 );
      

    assert(nVars <= 16, "nVars <= 16");
    for  i = 0; i < nVars; i++  {
        pCanonPerm[i] = i;
    }

    uCanonPhase = Abc_TtSemiCanonicize( pTruth, nVars, pCanonPerm, pStoreIn, 1 );
    for  k = 0; k < 5; k++  {
        var fChanges int32;
        for  i = nVars - 2; i >= 0; i--  {
            if pStoreIn[i] == pStoreIn[i+1] {
                fChanges |= Abc_TtCofactorPerm( pTruth, i, nWords, 1, pCanonPerm, &uCanonPhase, fNaive ); }
        }
        if !(fChanges != 0) {
            break; }
        fChanges = 0;
        for  i = 1; i < nVars - 1; i++  {
            if pStoreIn[i] == pStoreIn[i+1] {
                fChanges |= Abc_TtCofactorPerm( pTruth, i, nWords, 1, pCanonPerm, &uCanonPhase, fNaive ); }
        }
        if !(fChanges != 0) {
            break; }
    }

                   
    Abc_TtCopy( spStatic_Abc_TtCanonicizePerm_pCopy2, pTruth, nWords, 0 );
    memcpy( pCanonPermCopy, pCanonPerm, sizeof(char) * nVars );
    Abc_TtImplementNpnConfig( spStatic_Abc_TtCanonicizePerm_pCopy2, nVars, pCanonPermCopy, uCanonPhase );
    if !Abc_TtEqual( spStatic_Abc_TtCanonicizePerm_pCopy1, spStatic_Abc_TtCanonicizePerm_pCopy2, nWords ) {
        fmt.Printf("Canonical form verification failed!\n"); }
      

    assert(uCanonPhase == 0, "uCanonPhase == 0");
    return uCanonPhase;
}

/**Function*************************************************************

  Synopsis    [Semi-canonical form computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCanonicizePhaseVar6(pTruth *word, nVars int32, v int32) int32 {
    var w int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    var s int32;
    var nStep int32 = 1 << (v-6);
    assert(v >= 6, "v >= 6");
    for  w, s = nWords - 1, nWords - nStep; w > 0; w--  {
        if pTruth[w-nStep] == pTruth[w] {
            if w == s { w = s - nStep; s = w - nStep; }
            continue;
        }
        if pTruth[w-nStep] > pTruth[w] {
            return -1; }
        for  ; w > 0; w--  {
            // inlining of macro ABC_SWAP
            { var t word  = pTruth[w-nStep]; pTruth[w-nStep] = pTruth[w]; pTruth[w] = t; };
            if w == s { w = s - nStep; s = w - nStep; }
        }
        assert(w == -1, "w == -1");
        return 1;
    }
    return 0;
}
func Abc_TtCanonicizePhaseVar5(pTruth *word, nVars int32, v int32) int32 {
    var w int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    var Shift int32  = 1 << v;
    var Mask word  = s_Truths6[v];
    assert(v < 6, "v < 6");
    for  w = nWords - 1; w >= 0; w--  {
        if ((pTruth[w] << Shift) & Mask) == (pTruth[w] & Mask) {
            continue; }
        if ((pTruth[w] << Shift) & Mask) > (pTruth[w] & Mask) {
            return -1; }
//        Extra_PrintHex( stdout, (unsigned *)pTruth, nVars ); printf("\n" );
        for  ; w >= 0; w--  {
            pTruth[w] = ((pTruth[w] << Shift) & Mask) | ((pTruth[w] & Mask) >> Shift);
        }
//        Extra_PrintHex( stdout, (unsigned *)pTruth, nVars ); printf( " changed %d", v ), printf("\n" );
        return 1;
    }
    return 0;
}
var spStatic_Abc_TtCanonicizePhase_pCopy1 [1024]word;
var spStatic_Abc_TtCanonicizePhase_pCopy2 [1024]word;
func Abc_TtCanonicizePhase(pTruth *word, nVars int32) uint32 {
    var uCanonPhase uint32;
    var v int32;
    var nWords int32 = Abc_TtWordNum( nVars );
//    static int Counter = 0;
//    Counter++;

                   
    
    
    Abc_TtCopy( spStatic_Abc_TtCanonicizePhase_pCopy1, pTruth, nWords, 0 );
      

    if (pTruth[nWords-1] >> 63) & 1 {
        Abc_TtNot( pTruth, nWords );
        uCanonPhase ^= (1 << nVars);
    }

//    while ( 1 )
//    {
//        unsigned uCanonPhase2 = uCanonPhase;
        for  v = nVars - 1; v >= 6; v--  {
            if Abc_TtCanonicizePhaseVar6( pTruth, nVars, v ) == 1 {
                uCanonPhase ^= 1 << v; }
        }
        for  ; v >= 0; v--  {
            if Abc_TtCanonicizePhaseVar5( pTruth, nVars, v ) == 1 {
                uCanonPhase ^= 1 << v; }
        }
//        if ( uCanonPhase2 == uCanonPhase )
//            break;
//    }

//    for ( v = 5; v >= 0; v-- )
//        assert( Abc_TtCanonicizePhaseVar5( pTruth, nVars, v ) != 1 );

                   
    Abc_TtCopy( spStatic_Abc_TtCanonicizePhase_pCopy2, pTruth, nWords, 0 );
    Abc_TtImplementNpnConfig( spStatic_Abc_TtCanonicizePhase_pCopy2, nVars, nil, uCanonPhase );
    if !Abc_TtEqual( spStatic_Abc_TtCanonicizePhase_pCopy1, spStatic_Abc_TtCanonicizePhase_pCopy2, nWords ) {
        fmt.Printf("Canonical form verification failed!\n"); }
      
    return uCanonPhase;
}


/**Function*************************************************************

  Synopsis    [Hierarchical semi-canonical form computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
const TT_MAX_LEVELS = 5

type Abc_TtHieMan_t_ struct {
    nLastLevel, nWords int32;
    Vec_Mem_t *  vTtMem[TT_MAX_LEVELS];    // truth table memory and hash tables
    Vec_Int_t *  vRepres[TT_MAX_LEVELS];   // pointers to the representatives from the last hierarchical level
    vTruthId [TT_MAX_LEVELS]int32;

    vPhase  *Vec_Int_t;
}

func Abc_TtHieManStart(nVars int32, nLevels int32) *Abc_TtHieMan_t {
    var p *Abc_TtHieMan_t  = nil;
    var i int32;
    if nLevels > TT_MAX_LEVELS { return p; }
    // inlining of macro ABC_CALLOC
    p = &Abc_TtHieMan_t{};
    p.nLastLevel = nLevels - 1;
    p.nWords = Abc_TtWordNum(nVars);
    for i = 0; i < nLevels; i++ {
        p.vTtMem[i] = Vec_MemAlloc(p.nWords, 12);
        Vec_MemHashAlloc(p.vTtMem[i], 10000);
        p.vRepres[i] = Vec_IntAlloc(1);
    }
    p.vPhase = Vec_IntAlloc(2500);
    return p;
}

func Abc_TtHieManStop(p *Abc_TtHieMan_t) {
    var i int32;
    for i = 0; i <= p.nLastLevel; i++ {
        Vec_MemHashFree(p.vTtMem[i]);
        Vec_MemFreeP(&p.vTtMem[i]);
        Vec_IntFree(p.vRepres[i]);
    }
    Vec_IntFree( p.vPhase );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

func Abc_TtHieRetrieveOrInsert(p *Abc_TtHieMan_t, level int32, pTruth *word, pResult *word) int32 {
    var i, iSpot, truthId int32;
    var pRepTruth *word;
    if !(p != nil) { return -1; }
    if level < 0 { level += p.nLastLevel + 1; }
    if level < 0 || level > p.nLastLevel { return -1; }
    iSpot = *Vec_MemHashLookup(p.vTtMem[level], pTruth);
    if iSpot == -1 {
        p.vTruthId[level] = Vec_MemHashInsert(p.vTtMem[level], pTruth);
        if level < p.nLastLevel { return 0; }
        iSpot = p.vTruthId[level];
    }
    // return the class representative
    if level < p.nLastLevel {
        truthId = Vec_IntEntry(p.vRepres[level], iSpot);
    } else {
        truthId = iSpot; }
    for i = 0; i < level; i++ {
        Vec_IntSetEntry(p.vRepres[i], p.vTruthId[i], truthId);
    }

    pRepTruth = Vec_MemReadEntry(p.vTtMem[p.nLastLevel], truthId);
    if level < p.nLastLevel {
        Abc_TtCopy(pResult, pRepTruth, p.nWords, 0);
        return 1;
    }
    assert(Abc_TtEqual(pTruth, pRepTruth, p.nWords), "Abc_TtEqual(pTruth, pRepTruth, p.nWords)");
    if pTruth != pResult {
        Abc_TtCopy(pResult, pRepTruth, p.nWords, 0); }
    return 0;
}

func Abc_TtCanonicizeHie(p *Abc_TtHieMan_t, pTruthInit *word, nVars int32, pCanonPerm string, fExact int32) uint32 {
    var fNaive int32  = 1;
    var pStore [17]int32;
    //static word pTruth[1024];
    var pTruth *word  = pTruthInit;
    var uCanonPhase uint32;
    var nOnes int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    var i, k int32;
    assert(nVars <= 16, "nVars <= 16");

    // handle constant
    if nVars == 0 {
        Abc_TtClear( pTruthInit, nWords );
        return 0;
    }

    //Abc_TtCopy( pTruth, pTruthInit, nWords, 0 );

    for  i = 0; i < nVars; i++  {
        pCanonPerm[i] = i;
    }

    // normalize polarity    
    nOnes = Abc_TtCountOnesInTruth( pTruth, nVars );
    if nOnes > nWords * 32 {
        Abc_TtNot( pTruth, nWords );
        nOnes = nWords*64 - nOnes;
        uCanonPhase |= (1 << nVars);
    }
    // check cache
    if Abc_TtHieRetrieveOrInsert(p, 0, pTruth, pTruthInit) > 0 { return 0; }

    // normalize phase
    Abc_TtCountOnesInCofs( pTruth, nVars, pStore );
    pStore[nVars] = nOnes;
    for  i = 0; i < nVars; i++  {
        if pStore[i] >= nOnes - pStore[i] {
            continue; }
        Abc_TtFlip( pTruth, nWords, i );
        uCanonPhase |= (1 << i);
        pStore[i] = nOnes - pStore[i]; 
    }
    // check cache
    if Abc_TtHieRetrieveOrInsert(p, 1, pTruth, pTruthInit) > 0 { return 0; }

    // normalize permutation
    {
        var k, BestK int32;
        for  i = 0; i < nVars - 1; i++  {
            BestK = i + 1;
            for  k = i + 2; k < nVars; k++  {
                if pStore[BestK] > pStore[k] {
                    BestK = k; }
            }
            if pStore[i] <= pStore[BestK] {
                continue; }
            // inlining of macro ABC_SWAP
            { var t int32  = pCanonPerm[i]; pCanonPerm[i] = pCanonPerm[BestK]; pCanonPerm[BestK] = t; };
            // inlining of macro ABC_SWAP
            { var t int32  = pStore[i]; pStore[i] = pStore[BestK]; pStore[BestK] = t; };
            if ((uCanonPhase >> i) & 1) != ((uCanonPhase >> BestK) & 1) {
                uCanonPhase ^= (1 << i);
                uCanonPhase ^= (1 << BestK);
            }
            Abc_TtSwapVars( pTruth, nVars, i, BestK );
        }
    }
    // check cache
    if Abc_TtHieRetrieveOrInsert(p, 2, pTruth, pTruthInit) > 0 { return 0; }

    // iterate TT permutations for tied variables
    for  k = 0; k < 5; k++  {
        var fChanges int32;
        for  i = nVars - 2; i >= 0; i--  {
            if pStore[i] == pStore[i+1] {
                fChanges |= Abc_TtCofactorPerm( pTruth, i, nWords, pStore[i] != pStore[nVars]/2, pCanonPerm, &uCanonPhase, fNaive ); }
        }
        if !(fChanges != 0) {
            break; }
        fChanges = 0;
        for  i = 1; i < nVars - 1; i++  {
            if pStore[i] == pStore[i+1] {
                fChanges |= Abc_TtCofactorPerm( pTruth, i, nWords, pStore[i] != pStore[nVars]/2, pCanonPerm, &uCanonPhase, fNaive ); }
        }
        if !(fChanges != 0) {
            break; }
    }
    // check cache
    if Abc_TtHieRetrieveOrInsert(p, 3, pTruth, pTruthInit) > 0 { return 0; }

    // perform exact NPN using groups
    if fExact != 0 {
        extern void simpleMinimalGroups(word* x, word* pAux, word* minimal, int* pGroups, int nGroups, permInfo** pis, int nVars, int fFlipOutput, int fFlipInput);
        var pAuxWord, pAuxWord1 [1024]word;
        var pGroups [16]int32;
        var nGroups int32;
        var pis [17]*permInfo;
        // get groups
        pGroups[0] = 0;
        for i = 0; i < nVars - 1; i++ {
            if pStore[i] == pStore[i + 1] {
                pGroups[nGroups]++;
            } else {
                pGroups[nGroups]++;
                nGroups++;
                pGroups[nGroups] = 0;
            }
        }
        pGroups[nGroups]++;
        nGroups++;

        // compute permInfo from 0 to nVars  (incl.)
        for i = 0; i <= nVars; i++ {
            pis[i] = setPermInfoPtr(i);
        }

        // do the exact matching
        if nOnes == nWords * 32 { /* balanced output */
            simpleMinimalGroups(pTruth, pAuxWord, pAuxWord1, pGroups, nGroups, pis, nVars, 1, 1);
        } else if pStore[0] != pStore[1] && pStore[0] == (nOnes - pStore[0]) { /* balanced singleton input */
            simpleMinimalGroups(pTruth, pAuxWord, pAuxWord1, pGroups, nGroups, pis, nVars, 0, 1);
        } else {
            simpleMinimalGroups(pTruth, pAuxWord, pAuxWord1, pGroups, nGroups, pis, nVars, 0, 0); }

        // cleanup
        for i = 0; i <= nVars; i++ {
            freePermInfoPtr(pis[i]);
        }
    }
    // update cache
    Abc_TtHieRetrieveOrInsert(p, 4, pTruth, pTruthInit);
    return 0;
}


/**Function*************************************************************

Synopsis    [Adaptive exact/semi-canonical form computation.]

Description []

SideEffects []

SeeAlso     []

***********************************************************************/

typedef struct TiedGroup_
{
    char iStart;                         // index of Abc_TgMan_t::pPerm
    char nGVars;                         // the number of variables in the group
}
 TiedGroup;

typedef struct Abc_TgMan_t_
{
    word *pTruth;
    int nVars;                          // the number of variables
    int nGVars;                         // the number of variables in groups ( symmetric variables purged )
    int nGroups;                        // the number of tied groups
    unsigned uPhase;                    // phase of each variable and the function
    int fPhased;                        // if the phases of all the variables are determined
    char pPerm[16];                     // permutation of variables, symmetric variables purged, for grouping
    char pPermT[16];                    // permutation of variables, symmetric variables expanded, actual transformation for pTruth
    char pPermTRev[16];                 // reverse permutation of pPermT
    signed char pPermDir[16];           // for generating the next permutation
    TiedGroup pGroup[16];               // tied groups
    // symemtric group attributes
    char symPhase[16];                  // phase type of symemtric groups
    signed char symLink[17];            // singly linked list, indicate the variables in symemtric groups

    int  nAlgorithm;                    // 0: AdjCE,  1: AdjSE,  2: E: Cost-Aware
    char pFGrps[16];                    // tied groups to be flipped
    Vec_Int_t * vPhase;                 // candidate phase assignments
}
 Abc_TgMan_t;

                                              
                    
      
/**Function*************************************************************

Synopsis    [Utilities.]

Description []

SideEffects []

SeeAlso     []

***********************************************************************/

// Johnson_CTrotter algorithm
static int Abc_NextPermSwapC(char * pData, signed char * pDir, int size)
{
    var i int32;
    var j int32;
    var k int32 = -1;
    for i = 0; i < size; i++ {
        j = i + pDir[i];
        if j >= 0 && j < size && pData[i] > pData[j] && (k < 0 || pData[i] > pData[k]) {
            k = i; }
    }
    if k < 0 { k = 0; }

    for i = 0; i < size; i++ {
        if pData[i] > pData[k] {
            pDir[i] = -pDir[i]; }
    }

    j = k + pDir[k];
    if j < k {
        return j;
    } else {
        return k;
    }
}

//typedef unsigned(*TtCanonicizeFunc)(Abc_TtHieMan_t * p, word * pTruth, int nVars, char * pCanonPerm, int flag);
var spStatic_Abc_TtCanonicizeWrap_pTruth2 [1024]word;
func Abc_TtCanonicizeWrap(func TtCanonicizeFunc, p *Abc_TtHieMan_t, pTruth *word, nVars int32, pCanonPerm string, flag int32) uint32 {
    var nWords int32  = Abc_TtWordNum(nVars);
    var uCanonPhase1, uCanonPhase2 uint32;
    var pCanonPerm2 [16]byte;
    

    Abc_TtNormalizeSmallTruth(pTruth, nVars);
    if Abc_TtCountOnesInTruth(pTruth, nVars) != nWords * 32 {
        return func(p, pTruth, nVars, pCanonPerm, flag); }
    Abc_TtCopy(spStatic_Abc_TtCanonicizeWrap_pTruth2, pTruth, nWords, 1);
    uCanonPhase1 = func(p, pTruth, nVars, pCanonPerm, flag);
    uCanonPhase2 = func(p, spStatic_Abc_TtCanonicizeWrap_pTruth2, nVars, pCanonPerm2, flag);
    if Abc_TtCompareRev(pTruth, spStatic_Abc_TtCanonicizeWrap_pTruth2, nWords) <= 0 {
        return uCanonPhase1; }
    Abc_TtCopy(pTruth, spStatic_Abc_TtCanonicizeWrap_pTruth2, nWords, 0);
    memcpy(pCanonPerm, pCanonPerm2, nVars);
    return uCanonPhase2;
}

var gpVerCopy [1024]word;
var spStatic_Abc_TtCannonVerify_pCopy2 [1024]word;
func Abc_TtCannonVerify(pTruth *word, nVars int32, pCanonPerm string, uCanonPhase uint32) int32 {
                   
    var nWords int32  = Abc_TtWordNum(nVars);
    var pCanonPermCopy [16]byte;
    
    Abc_TtVerifySmallTruth(pTruth, nVars);
    Abc_TtCopy(spStatic_Abc_TtCannonVerify_pCopy2, pTruth, nWords, 0);
    memcpy(pCanonPermCopy, pCanonPerm, sizeof(char) * nVars);
    Abc_TtImplementNpnConfig(spStatic_Abc_TtCannonVerify_pCopy2, nVars, pCanonPermCopy, uCanonPhase);
    return Abc_TtEqual(gpVerCopy, spStatic_Abc_TtCannonVerify_pCopy2, nWords);
     
             
      
}

/**Function*************************************************************

Synopsis    [Tied group management.]

Description []

SideEffects []

SeeAlso     []

***********************************************************************/

func Abc_TgInitMan(pMan *Abc_TgMan_t, pTruth *word, nVars int32, nAlg int32, vPhase *Vec_Int_t) {
    var i int32;
    pMan.pTruth = pTruth;
    pMan.uPhase = 0;
    pMan.fPhased = 0;
    pMan.nVars = pMan.nGVars = nVars;
    pMan.nGroups = 1;
    pMan.pGroup[0].iStart = 0;
    pMan.pGroup[0].nGVars = nVars;
    for i = 0; i < nVars; i++ {
        pMan.pPerm[i] = i;
        pMan.pPermT[i] = i;
        pMan.pPermTRev[i] = i;
        pMan.symPhase[i] = 1;
        pMan.symLink[i] = -1;
    }
    pMan.symLink[i] = -1;
    pMan.nAlgorithm = nAlg;
    pMan.vPhase = vPhase;
    Vec_IntClear(vPhase);
}

func Abc_TgSplitGroup(pMan *Abc_TgMan_t, pGrp *TiedGroup, pCoef *int32) int32 {
    var i, j, n int32;
    var nGVars int32  = pGrp.nGVars;
    var pVars string  = pMan.pPerm + pGrp.iStart;
    var iGrp int32  = pGrp - pMan.pGroup;

    // sort variables
    for i = 1; i < nGVars; i++ {
        var a int32  = pCoef[i]; var aa byte  = pVars[i];
        for j = i; j > 0 && pCoef[j - 1] > a; j-- {
            pCoef[j] = pCoef[j - 1], pVars[j] = pVars[j - 1];
        }
        pCoef[j] = a; pVars[j] = aa;
    }
    for i = 1; i < nGVars; i++ {
        if pCoef[i] != pCoef[i - 1] { n++; }
    }
    if n == 0 { return 0; }
    memmove(pGrp + n + 1, pGrp + 1, (pMan.nGroups - iGrp - 1) * sizeof(TiedGroup));
    // group variables
    for i = j = 1; i < nGVars; i++ {
        if pCoef[i] == pCoef[i - 1] { continue; }
        pGrp[j].iStart = pGrp.iStart + i;
        pGrp[j - 1].nGVars = pGrp[j].iStart - pGrp[j - 1].iStart;
        j++;
    }
    assert(j == n + 1, "j == n + 1");
    pGrp[n].nGVars = pGrp.iStart + i - pGrp[n].iStart;
    pMan.nGroups += n;
    return n;
}
func Abc_TgManCopy(pDst *Abc_TgMan_t, pDstTruth *word, pSrc *Abc_TgMan_t) {
    *pDst = *pSrc;
    Abc_TtCopy(pDstTruth, pSrc.pTruth, Abc_TtWordNum(pSrc.nVars), 0);
    pDst.pTruth = pDstTruth;
}

func Abc_TgCannonVerify(pMan *Abc_TgMan_t) int32 {
    return Abc_TtCannonVerify(pMan.pTruth, pMan.nVars, pMan.pPermT, pMan.uPhase);
}

// extern int Abc_TgExpendSymmetry(Abc_TgMan_t * pMan, char * pDest);


func CheckConfig(pMan *Abc_TgMan_t) {
#ifndef NDEBUG
    int i;
    char pPermE[16];
    Abc_TgExpendSymmetry(pMan, pPermE);
    for (i = 0; i < pMan->nVars; i++)
    {
        assert(pPermE[i] == pMan->pPermT[i]);
        assert(pMan->pPermTRev[(int)pMan->pPermT[i]] == i);
    }
    assert(Abc_TgCannonVerify(pMan));
#endif
}

/**Function*************************************************************

Synopsis    [Truthtable manipulation.]

Description []

SideEffects []

SeeAlso     []

***********************************************************************/

func Abc_TgFlipVar(pMan *Abc_TgMan_t, iVar int32) {
    var nWords int32  = Abc_TtWordNum(pMan.nVars);
    var ivp int32  = pMan.pPermTRev[iVar];
    Abc_TtFlip(pMan.pTruth, nWords, ivp);
    pMan.uPhase ^= 1 << ivp;
}

func Abc_TgFlipSymGroupByVar(pMan *Abc_TgMan_t, iVar int32) {
    for ; iVar >= 0; iVar = pMan.symLink[iVar] {
        if pMan.symPhase[iVar] {
            Abc_TgFlipVar(pMan, iVar); }
    }
}

func Abc_TgFlipSymGroup(pMan *Abc_TgMan_t, idx int32) {
    Abc_TgFlipSymGroupByVar(pMan, pMan.pPerm[idx]);
}

func Abc_TgClearSymGroupPhase(pMan *Abc_TgMan_t, iVar int32) {
    for ; iVar >= 0; iVar = pMan.symLink[iVar] {
        pMan.symPhase[iVar] = 0;
    }
}

func Abc_TgImplementPerm(pMan *Abc_TgMan_t, pPermDest string) {
    var i int32;
    var nVars int32 = pMan.nVars;
    var pPerm string  = pMan.pPermT;
    var pRev string  = pMan.pPermTRev;
    var uPhase uint32  = pMan.uPhase & (1 << nVars);

    for i = 0; i < nVars; i++ {
        pRev[int32(pPerm[i])] = i;
    }
    for i = 0; i < nVars; i++ {
        pPerm[i] = pRev[int32(pPermDest[i])];
    }
    for i = 0; i < nVars; i++ {
        pRev[int32(pPerm[i])] = i;
    }

    Abc_TtImplementNpnConfig(pMan.pTruth, nVars, pRev, 0);
//  Abc_TtVerifySmallTruth(pMan->pTruth, nVars);

    for i = 0; i < nVars; i++ {
        if pMan.uPhase & (1 << pPerm[i]) {
            uPhase |= (1 << i); }
        pPerm[i] = pPermDest[i];
        pRev[int32(pPerm[i])] = i;
    }
    pMan.uPhase = uPhase;
}

func Abc_TgSwapAdjacentSymGroups(pMan *Abc_TgMan_t, idx int32) {
    var iVar, jVar, ix int32;
    var pPermNew [16]byte;
    assert(idx < pMan.nGVars - 1, "idx < pMan.nGVars - 1");
    iVar = pMan.pPerm[idx];
    jVar = pMan.pPerm[idx + 1];
    pMan.pPerm[idx] = jVar;
    pMan.pPerm[idx + 1] = iVar;
    // inlining of macro ABC_SWAP
    { var t byte  = pMan.pPermDir[idx]; pMan.pPermDir[idx] = pMan.pPermDir[idx + 1]; pMan.pPermDir[idx + 1] = t; };
    if pMan.symLink[iVar] >= 0 || pMan.symLink[jVar] >= 0 {
        Abc_TgExpendSymmetry(pMan, pPermNew);
        Abc_TgImplementPerm(pMan, pPermNew);
        return;
    }
    // plain variable swap
    ix = pMan.pPermTRev[iVar];
    assert(pMan.pPermT[ix] == iVar && pMan.pPermT[ix + 1] == jVar, "pMan.pPermT[ix] == iVar && pMan.pPermT[ix + 1] == jVar");
    Abc_TtSwapAdjacent(pMan.pTruth, Abc_TtWordNum(pMan.nVars), ix);
    pMan.pPermT[ix] = jVar;
    pMan.pPermT[ix + 1] = iVar;
    pMan.pPermTRev[iVar] = ix + 1;
    pMan.pPermTRev[jVar] = ix;
    if ((pMan.uPhase >> ix) & 1) != ((pMan.uPhase >> (ix + 1)) & 1) {
        pMan.uPhase ^= 1 << ix | 1 << (ix + 1); }
    assert((Abc_TgCannonVerify(pMan) != 0), "Abc_TgCannonVerify(pMan)");
}

/**Function*************************************************************

Synopsis    [semmetry of two variables.]

Description []

SideEffects []

SeeAlso     []

***********************************************************************/

var pSymCopy [1024]word;

func Abc_TtIsSymmetric(pTruth *word, nVars int32, iVar int32, jVar int32, fPhase int32) int32 {
    var rv int32;
    var nWords int32  = Abc_TtWordNum(nVars);
    Abc_TtCopy(pSymCopy, pTruth, nWords, 0);
    Abc_TtSwapVars(pSymCopy, nVars, iVar, jVar);
    rv = Abc_TtEqual(pTruth, pSymCopy, nWords) * 2;
    if !(fPhase != 0) { return rv; }
    Abc_TtFlip(pSymCopy, nWords, iVar);
    Abc_TtFlip(pSymCopy, nWords, jVar);
    return rv + Abc_TtEqual(pTruth, pSymCopy, nWords);
}

func Abc_TtIsSymmetricHigh(pMan *Abc_TgMan_t, iVar int32, jVar int32, fPhase int32) int32 {
    var rv, iv, jv, n int32;
    var nWords int32  = Abc_TtWordNum(pMan.nVars);
    Abc_TtCopy(pSymCopy, pMan.pTruth, nWords, 0);
    for n, iv, jv = 0, iVar, jVar; iv >= 0 && jv >= 0; iv = pMan.symLink[iv], jv = pMan.symLink[jv], n++ {
        Abc_TtSwapVars(pSymCopy, pMan.nVars, iv, jv);
    }
    assert(iv < 0 && jv < 0, "iv < 0 && jv < 0");   // two symmetric groups must have the same size 
    rv = Abc_TtEqual(pMan.pTruth, pSymCopy, nWords) * 2;
    if !(fPhase != 0) { return rv; }
    for iv, jv = iVar, jVar; iv >= 0 && jv >= 0; iv = pMan.symLink[iv], jv = pMan.symLink[jv] {
        if pMan.symPhase[iv] { Abc_TtFlip(pSymCopy, nWords, iv); }
        if pMan.symPhase[jv] { Abc_TtFlip(pSymCopy, nWords, jv); }
    }
    return rv + Abc_TtEqual(pMan.pTruth, pSymCopy, nWords);
}

/**Function*************************************************************

Synopsis    [Create groups by cofactor signatures]

Description [Similar to Abc_TtSemiCanonicize. 
             Use stable insertion sort to keep the order of the variables in the groups.
             Defer permutation. ]

SideEffects []

SeeAlso     []

***********************************************************************/

func Abc_TgCreateGroups(pMan *Abc_TgMan_t) {
    var pStore [17]int32;
    var i, nOnes int32;
    var nVars int32 = pMan.nVars;
    var nWords int32 = Abc_TtWordNum(nVars);
    //TiedGroup * pGrp = pMan->pGroup;
    assert(nVars <= 16, "nVars <= 16");
    // normalize polarity    
    nOnes = Abc_TtCountOnesInTruth(pMan.pTruth, nVars);
    if nOnes > nWords * 32 {
        Abc_TtNot(pMan.pTruth, nWords);
        nOnes = nWords * 64 - nOnes;
        pMan.uPhase |= (1 << nVars);
    }
    // normalize phase
    Abc_TtCountOnesInCofs(pMan.pTruth, nVars, pStore);
    pStore[nVars] = nOnes;
    for i = 0; i < nVars; i++ {
        if pStore[i] >= nOnes - pStore[i] {
            continue; }
        Abc_TtFlip(pMan.pTruth, nWords, i);
        pMan.uPhase |= (1 << i);
        pStore[i] = nOnes - pStore[i];
    }

    Abc_TgSplitGroup(pMan, pMan.pGroup, pStore);
    pMan.fPhased = pStore[0] * 2 != nOnes;
}

/**Function*************************************************************

Synopsis    [Group symmetric variables.]

Description []

SideEffects []

SeeAlso     []

***********************************************************************/

func Abc_TgGroupSymmetry(pMan *Abc_TgMan_t, pGrp *TiedGroup, fHigh int32) int32 {
    var i, j, iVar, jVar, nsym int32;
    var fDone, scnt, stype [16]int32;
    signed char *symLink = pMan.symLink;
//  char * symPhase = pMan->symPhase;
    var nGVars int32  = pGrp.nGVars;
    var fPhase int32  = (pGrp == pMan.pGroup) && !pMan.fPhased;
    var pVars string  = pMan.pPerm + pGrp.iStart;
    var modified int32;

    for i = 0; i < nGVars; i++ {
        fDone[i] = 0, scnt[i] = 1;
    }

    do {
        modified = 0;
        for i = 0; i < nGVars - 1; i++ {
            iVar = pVars[i];
            if iVar < 0 || fDone[i] { continue; }
//          if (!pGrp->fPhased && !Abc_TtHasVar(pMan->pTruth, pMan->nVars, iVar)) continue;
            // Mark symmetric variables/groups
            for j = i + 1; j < nGVars; j++ {
                jVar = pVars[j];
                if jVar < 0 || scnt[j] != scnt[i] { // || pMan->symPhase[jVar] != pMan->symPhase[iVar])
                    stype[j] = 0;
                } else if scnt[j] == 1 {
                    stype[j] = Abc_TtIsSymmetric(pMan.pTruth, pMan.nVars, iVar, jVar, fPhase);
                } else {
                    stype[j] = Abc_TtIsSymmetricHigh(pMan, iVar, jVar, fPhase); }
            }
            fDone[i] = 1;
            // Merge symmetric groups
            for j = i + 1; j < nGVars; j++ {
                var ii int32;
                jVar = pVars[j];
                switch stype[j] {
                case 1:         // E-Symmetry
                    Abc_TgFlipSymGroupByVar(pMan, jVar);
                    // fallthrough
                fallthrough
case 2:         // NE-Symmetry
                    pMan.symPhase[iVar] += pMan.symPhase[jVar];
                    break;
                case 3:         // multiform Symmetry
                    Abc_TgClearSymGroupPhase(pMan, jVar);
                    break;
                default:        // case 0: No Symmetry
                    continue;
                }

                for ii = iVar; symLink[ii] >= 0; ii = symLink[ii] {}
                symLink[ii] = jVar;
                pVars[j] = -1;
                scnt[i] += scnt[j];
                modified = 1;
                fDone[i] = 0;
                nsym++;
            }
        }
//      if (++order > 3) printf("%d", order);
    } while (fHigh && modified);

    return nsym;
}

func Abc_TgPurgeSymmetry(pMan *Abc_TgMan_t, fHigh int32) {
    var i int32;
    var j int32;
    var k int32;
    var sum int32;
    var nVars int32 = pMan.nVars;
    signed char *symLink = pMan.symLink;
    var gcnt [16]byte ;
    var pPerm string  = pMan.pPerm;

    // purge unsupported variables
    if !pMan.fPhased {
        var iVar int32  = pMan.nVars;
        for j = 0; j < pMan.pGroup[0].nGVars; j++ {
            var jVar int32  = pPerm[j];
            assert(jVar >= 0, "jVar >= 0");
            if !Abc_TtHasVar(pMan.pTruth, nVars, jVar) {
                symLink[jVar] = symLink[iVar];
                symLink[iVar] = jVar;
                pPerm[j] = -1;
                gcnt[0]++;
            }
        }
    }

    for k = 0; k < pMan.nGroups; k++ {
        gcnt[k] += Abc_TgGroupSymmetry(pMan, pMan.pGroup + k, fHigh);
    }

    for i = 0; i < nVars && pPerm[i] >= 0; i++ {}
    for j = i + 1; ; i, j = i+1, j+1 {
        for j < nVars && pPerm[j] < 0 { j++; }
        if j >= nVars { break; }
        pPerm[i] = pPerm[j];
    }
    for k = 0; k < pMan.nGroups; k++ {
        pMan.pGroup[k].nGVars -= gcnt[k];
        pMan.pGroup[k].iStart -= sum;
        sum += gcnt[k];
    }
    if pMan.pGroup[0].nGVars == 0 {
        pMan.nGroups--;
        memmove(pMan.pGroup, pMan.pGroup + 1, sizeof(TiedGroup) * pMan.nGroups);
        assert(pMan.pGroup[0].iStart == 0, "pMan.pGroup[0].iStart == 0");
    }
    pMan.nGVars -= sum;
}

func Abc_TgExpendSymmetry(pMan *Abc_TgMan_t, pDest string) int32 {
    var i, j, k, s int32;
    var pPerm string  = pMan.pPerm;
    for j = 0; j < pMan.nGVars; j++ {
        for k = pPerm[j]; k >= 0; k = pMan.symLink[k] {
            pDest[i++] = k;
        }
    }
    s = i;
    for k = pMan.symLink[pMan.nVars]; k >= 0; k = pMan.symLink[k] {
        pDest[i++] = k;
    }
    assert(i == pMan.nVars, "i == pMan.nVars");
    return s;
}

func Abc_TgResetGroup(pMan *Abc_TgMan_t) {
    var i, j int32;
    var pPerm string  = pMan.pPerm;
    var pPermNew [16]byte;
    var nGVars int32  = pMan.nGVars;
    for i = 1; i < nGVars; i++ {
        var t byte  = pPerm[i];
        for j = i; j > 0 && pPerm[j - 1] > t; j-- {
            pPerm[j] = pPerm[j - 1];
        }
        pPerm[j] = t;
    }
    Abc_TgExpendSymmetry(pMan, pPermNew);
    Abc_TgImplementPerm(pMan, pPermNew);
    pMan.fPhased = 0;
    pMan.nGroups = 1;
    pMan.pGroup.nGVars = nGVars;
    Vec_IntClear(pMan.vPhase);
}

func Abc_TgResetGroup1(pMan *Abc_TgMan_t) {
    var i, j int32;
    var pPermNew [16]byte;
    var nSVars int32  = Abc_TgExpendSymmetry(pMan, pPermNew);
    for i = 1; i < nSVars; i++ {
        var t byte  = pPermNew[i];
        for j = i; j > 0 && pPermNew[j - 1] > t; j-- {
            pPermNew[j] = pPermNew[j - 1];
        }
        pPermNew[j] = t;
    }
    Abc_TgImplementPerm(pMan, pPermNew);
    pMan.fPhased = 0;
    pMan.nGroups = 1;
    pMan.nGVars = pMan.pGroup.nGVars = nSVars;
    for i = 0; i < pMan.nVars; i++ {
        pMan.pPerm[i] = pPermNew[i];
        pMan.symPhase[i] = 1;
        pMan.symLink[i] = -1;
    }
    Vec_IntClear(pMan.vPhase);
}

/**Function*************************************************************

Synopsis    [Semi-canonical form computation.]

Description [simple config enumeration]

SideEffects []

SeeAlso     []

***********************************************************************/
var spStatic_Abc_TgSymGroupPerm_pCopy [1024]word;
var spStatic_Abc_TgSymGroupPerm_pBest [1024]word;
func Abc_TgSymGroupPerm(pMan *Abc_TgMan_t, idx int32, fSwapOnly int32) int32 {
    var pTruth *word  = pMan.pTruth;
    
    
    var Config int32;
    var nWords int32  = Abc_TtWordNum(pMan.nVars);
    var tgManCopy, tgManBest Abc_TgMan_t;

    CheckConfig(pMan);
    if fSwapOnly != 0 {
        Abc_TgManCopy(&tgManCopy, spStatic_Abc_TgSymGroupPerm_pCopy, pMan);
        Abc_TgSwapAdjacentSymGroups(&tgManCopy, idx);
        CheckConfig(&tgManCopy);
        if Abc_TtCompareRev(pTruth, spStatic_Abc_TgSymGroupPerm_pCopy, nWords) < 0 {
            Abc_TgManCopy(pMan, pTruth, &tgManCopy);
            return 4;
        }
        return 0;
    }

    // save two copies
    Abc_TgManCopy(&tgManCopy, spStatic_Abc_TgSymGroupPerm_pCopy, pMan);
    Abc_TgManCopy(&tgManBest, spStatic_Abc_TgSymGroupPerm_pBest, pMan);
    // PXY
    // 001
    Abc_TgFlipSymGroup(&tgManCopy, idx);
    CheckConfig(&tgManCopy);
    if Abc_TtCompareRev(spStatic_Abc_TgSymGroupPerm_pBest, spStatic_Abc_TgSymGroupPerm_pCopy, nWords) == 1 {
        Abc_TgManCopy(&tgManBest, spStatic_Abc_TgSymGroupPerm_pBest, &tgManCopy);
Config = 1; }
    // PXY
    // 011
    Abc_TgFlipSymGroup(&tgManCopy, idx + 1);
    CheckConfig(&tgManCopy);
    if Abc_TtCompareRev(spStatic_Abc_TgSymGroupPerm_pBest, spStatic_Abc_TgSymGroupPerm_pCopy, nWords) == 1 {
        Abc_TgManCopy(&tgManBest, spStatic_Abc_TgSymGroupPerm_pBest, &tgManCopy);
Config = 3; }
    // PXY
    // 010
    Abc_TgFlipSymGroup(&tgManCopy, idx);
    CheckConfig(&tgManCopy);
    if Abc_TtCompareRev(spStatic_Abc_TgSymGroupPerm_pBest, spStatic_Abc_TgSymGroupPerm_pCopy, nWords) == 1 {
        Abc_TgManCopy(&tgManBest, spStatic_Abc_TgSymGroupPerm_pBest, &tgManCopy);
Config = 2; }
    // PXY
    // 110
    Abc_TgSwapAdjacentSymGroups(&tgManCopy, idx);
    CheckConfig(&tgManCopy);
    if Abc_TtCompareRev(spStatic_Abc_TgSymGroupPerm_pBest, spStatic_Abc_TgSymGroupPerm_pCopy, nWords) == 1 {
        Abc_TgManCopy(&tgManBest, spStatic_Abc_TgSymGroupPerm_pBest, &tgManCopy);
Config = 6; }
    // PXY
    // 111
    Abc_TgFlipSymGroup(&tgManCopy, idx + 1);
    CheckConfig(&tgManCopy);
    if Abc_TtCompareRev(spStatic_Abc_TgSymGroupPerm_pBest, spStatic_Abc_TgSymGroupPerm_pCopy, nWords) == 1 {
        Abc_TgManCopy(&tgManBest, spStatic_Abc_TgSymGroupPerm_pBest, &tgManCopy);
Config = 7; }
    // PXY
    // 101
    Abc_TgFlipSymGroup(&tgManCopy, idx);
    CheckConfig(&tgManCopy);
    if Abc_TtCompareRev(spStatic_Abc_TgSymGroupPerm_pBest, spStatic_Abc_TgSymGroupPerm_pCopy, nWords) == 1 {
        Abc_TgManCopy(&tgManBest, spStatic_Abc_TgSymGroupPerm_pBest, &tgManCopy);
Config = 5; }
    // PXY
    // 100
    Abc_TgFlipSymGroup(&tgManCopy, idx + 1);
    CheckConfig(&tgManCopy);
    if Abc_TtCompareRev(spStatic_Abc_TgSymGroupPerm_pBest, spStatic_Abc_TgSymGroupPerm_pCopy, nWords) == 1 {
        Abc_TgManCopy(&tgManBest, spStatic_Abc_TgSymGroupPerm_pBest, &tgManCopy);
Config = 4; }
    // PXY
    // 000
    Abc_TgSwapAdjacentSymGroups(&tgManCopy, idx);
    CheckConfig(&tgManCopy);
    assert(Abc_TtEqual(pTruth, spStatic_Abc_TgSymGroupPerm_pCopy, nWords), "Abc_TtEqual(pTruth, spStatic_Abc_TgSymGroupPerm_pCopy, nWords)");
    if Config == 0 {
        return 0; }
    assert(Abc_TtCompareRev(pTruth, spStatic_Abc_TgSymGroupPerm_pBest, nWords) == 1, "Abc_TtCompareRev(pTruth, spStatic_Abc_TgSymGroupPerm_pBest, nWords) == 1");
    Abc_TgManCopy(pMan, pTruth, &tgManBest);
    return Config;
}

var spStatic_Abc_TgPermPhase_pCopy [1024]word;
func Abc_TgPermPhase(pMan *Abc_TgMan_t, iVar int32) int32 {
    
    var nWords int32  = Abc_TtWordNum(pMan.nVars);
    var ivp int32  = pMan.pPermTRev[iVar];
    Abc_TtCopy(spStatic_Abc_TgPermPhase_pCopy, pMan.pTruth, nWords, 0);
    Abc_TtFlip(spStatic_Abc_TgPermPhase_pCopy, nWords, ivp);
    if Abc_TtCompareRev(pMan.pTruth, spStatic_Abc_TgPermPhase_pCopy, nWords) == 1 {
        Abc_TtCopy(pMan.pTruth, spStatic_Abc_TgPermPhase_pCopy, nWords, 0);
        pMan.uPhase ^= 1 << ivp;
        return 16;
    }
    return 0;
}

func Abc_TgSimpleEnumeration(pMan *Abc_TgMan_t) {
    var i, j, k int32;
    var pGid [16]int32;

    for k = j = 0; j < pMan.nGroups; j++ {
        for i = 0; i < pMan.pGroup[j].nGVars; i, k = i+1, k+1 {
            pGid[k] = j;
        }
    }
    assert(k == pMan.nGVars, "k == pMan.nGVars");

    for k = 0; k < 5; k++ {
        var fChanges int32;
        for i = pMan.nGVars - 2; i >= 0; i-- {
            if pGid[i] == pGid[i + 1] {
                fChanges |= Abc_TgSymGroupPerm(pMan, i, pGid[i] > 0 || pMan.fPhased); }
        }
        for i = 1; i < pMan.nGVars - 1; i++ {
            if pGid[i] == pGid[i + 1] {
                fChanges |= Abc_TgSymGroupPerm(pMan, i, pGid[i] > 0 || pMan.fPhased); }
        }

        for i = pMan.nVars - 1; i >= 0; i-- {
            if pMan.symPhase[i] {
                fChanges |= Abc_TgPermPhase(pMan, i); }
        }
        for i = 1; i < pMan.nVars; i++ {
            if pMan.symPhase[i] {
                fChanges |= Abc_TgPermPhase(pMan, i); }
        }
        if !(fChanges != 0) { break; }
    }
    assert((Abc_TgCannonVerify(pMan) != 0), "Abc_TgCannonVerify(pMan)");
}

/**Function*************************************************************

Synopsis    [Exact canonical form computation.]

Description [full config enumeration]

SideEffects []

SeeAlso     []

***********************************************************************/

static int Abc_TgIsInitPerm(char * pData, signed char * pDir, int size)
{
    var i int32;
    if pDir[0] != -1 { return 0; }
    for i = 1; i < size; i++ {
        if pDir[i] != -1 || pData[i] < pData[i - 1] {
            return 0; }
    }
    return 1;
}

func Abc_TgFirstPermutation(pMan *Abc_TgMan_t) {
    var i int32;
    for i = 0; i < pMan.nGVars; i++ {
        pMan.pPermDir[i] = -1;
    }
#ifndef NDEBUG
    for (i = 0; i < pMan->nGroups; i++)
    {
        TiedGroup * pGrp = pMan->pGroup + i;
        int nGvars = pGrp->nGVars;
        char * pVars = pMan->pPerm + pGrp->iStart;
        signed char * pDirs = pMan->pPermDir + pGrp->iStart;
        assert(Abc_TgIsInitPerm(pVars, pDirs, nGvars));
    }
#endif
}

func Abc_TgNextPermutation(pMan *Abc_TgMan_t) int32 {
    var i, j, nGvars int32;
    var pGrp *TiedGroup;
    var pVars string;
    signed char * pDirs;
    for i = 0; i < pMan.nGroups; i++ {
        pGrp = pMan.pGroup + i;
        nGvars = pGrp.nGVars;
        if nGvars == 1 { continue; }
        pVars = pMan.pPerm + pGrp.iStart;
        pDirs = pMan.pPermDir + pGrp.iStart;
        j = Abc_NextPermSwapC(pVars, pDirs, nGvars);
        if j >= 0 {
            Abc_TgSwapAdjacentSymGroups(pMan, j + pGrp.iStart);
            return 1;
        }
        Abc_TgSwapAdjacentSymGroups(pMan, pGrp.iStart);
        assert(Abc_TgIsInitPerm(pVars, pDirs, nGvars), "Abc_TgIsInitPerm(pVars, pDirs, nGvars)");
    }
    return 0;
}

func grayCode(a uint32) uint32 { return a ^ (a >> 1); }

func grayFlip(a uint32) int32 {
    var i int32;
    for i, a = 0, a+1; ; i++ {
        if a & (1 << i) {   return i; }
    }
 }

func Abc_TgSaveBest(pMan *Abc_TgMan_t, pBest *Abc_TgMan_t) {
    if Abc_TtCompareRev(pBest.pTruth, pMan.pTruth, Abc_TtWordNum(pMan.nVars)) == 1 {
        Abc_TgManCopy(pBest, pBest.pTruth, pMan); }
}

func Abc_TgPhaseEnumeration(pMan *Abc_TgMan_t, pBest *Abc_TgMan_t) {
    var pFGrps [16]byte;
    var i int32;
    var j int32;
    var n int32 = pMan.pGroup.nGVars;

    Abc_TgSaveBest(pMan, pBest);
    if pMan.fPhased { return; }

    // sort by symPhase
    for i = 0; i < n; i++ {
        var iv byte  = pMan.pPerm[i];       
        for j = i; j > 0 && pMan.symPhase[int32(pFGrps[j-1])] > pMan.symPhase[int32(iv)]; j-- {
            pFGrps[j] = pFGrps[j - 1];
        }
        pFGrps[j] = iv;
    }

    for i = 0; i < (1 << n) - 1; i++ {
        Abc_TgFlipSymGroupByVar(pMan, pFGrps[grayFlip(i)]);
        Abc_TgSaveBest(pMan, pBest);
    }
}

/**Function*************************************************************

Synopsis    [Hybrid exact canonical form computation.]

Description []

SideEffects []

SeeAlso     []

***********************************************************************/
func Abc_TgCalcScc(pMan *Abc_TgMan_t, pOut *int32, fSort int32) {
    var i, j, k int32;
    var pGrp *TiedGroup;

    Abc_TtSccInCofs(pMan.pTruth, pMan.nVars, pOut);

    for i = j = 0; j < pMan.nGVars; j++ {
        pOut[j] = pOut[i];
        for k = pMan.pPerm[j]; k >= 0; k = pMan.symLink[k] {
            i++;
            assert(pMan.symLink[k] < 0 || pOut[j] == pOut[i], "pMan.symLink[k] < 0 || pOut[j] == pOut[i]");
        }
    }
    if !(fSort != 0) { return; }

    for pGrp = pMan.pGroup; pGrp < pMan.pGroup + pMan.nGroups; pGrp++ {
        var vb int32 = pGrp.iStart;
        var ve int32 = vb + pGrp.nGVars;
        for i = vb + 1; i < ve; i++ {
            var a int32  = pOut[i];
            for j = i; j > vb && pOut[j - 1] > a; j-- {
                pOut[j] = pOut[j - 1];
            }
            pOut[j] = a;
        }
    }
}

func Abc_TgSplitGroupsByScc(pMan *Abc_TgMan_t) {
    var pScc [16]int32;
    var pPermT [16]byte;
    var pGrp *TiedGroup;

    Abc_TgCalcScc(pMan, pScc, 0);
    for pGrp = pMan.pGroup; pGrp < pMan.pGroup + pMan.nGroups; pGrp++ {
        pGrp += Abc_TgSplitGroup(pMan, pGrp, pScc + pGrp.iStart);
    }

    Abc_TgExpendSymmetry(pMan, pPermT);
    Abc_TgImplementPerm(pMan, pPermT);
    assert((Abc_TgCannonVerify(pMan) != 0), "Abc_TgCannonVerify(pMan)");
}

func Abc_TgCompareCoef(pIn1 *int32, pIn2 *int32, n int32) int32 {
    var i int32;
    for i = 0; i < n; i++ {
        if pIn1[i] != pIn2[i] {
            if pIn1[i] < pIn2[i] {
                return -1;
            } else {
                return 1;
            } }
    }
    return 0;
}

// log2(n!)*100
var log2fn [17]int32 = [17]int32{ 0, 0, 100, 258, 458, 691, 949, 1230, 1530, 1847, 2179, 2525, 2884, 3254, 3634, 4025, 4425 };
func Abc_TgPermCostScc(pMan *Abc_TgMan_t, pScc *int32) int32 {
    var i, j, k, cost int32;

    for i = k = 0; i < pMan.nGroups; i++ {
        var n int32  = pMan.pGroup[i].nGVars;
        var d int32  = 1;
        for j, k = 1, k+1; j < n; j, k = j+1, k+1 {
            if pScc[k] == pScc[k - 1] { d++;
            } else {
                cost += log2fn[d];
                d = 1;
            }
        }
        cost += log2fn[d];
    }
    return cost;
}

var spStatic_Abc_TgPermEnumerationScc_pCopy [1024]word;
func Abc_TgPermEnumerationScc(pMan *Abc_TgMan_t, pBest *Abc_TgMan_t) {
    
    var tgManCopy Abc_TgMan_t;
    Abc_TgManCopy(&tgManCopy, spStatic_Abc_TgPermEnumerationScc_pCopy, pMan);
    if pMan.nAlgorithm > 1 {
        Abc_TgSplitGroupsByScc(&tgManCopy); }
    Abc_TgFirstPermutation(&tgManCopy);
    for {
        Abc_TgSaveBest(&tgManCopy, pBest);
    
if !((Abc_TgNextPermutation(&tgManCopy) != 0)) { break }
} 
}

func Abc_TgReorderFGrps(pMan *Abc_TgMan_t) {
    var pFGrps string  = pMan.pFGrps;
    var i int32;
    var j int32;
    var n int32;
    if pMan.fPhased {
        n = 0;
    } else {
        n = pMan.pGroup.nGVars;
    }

    // sort by symPhase
    for i = 0; i < n; i++ {
        var iv byte  = pMan.pPerm[i];
        for j = i; j > 0 && pMan.symPhase[int32(pFGrps[j - 1])] > pMan.symPhase[int32(iv)]; j-- {
            pFGrps[j] = pFGrps[j - 1];
        }
        pFGrps[j] = iv;
    }
}

func ilog2(n int32) int32 {
    var i int32;
    for i = 0; n /= 2; i++ {}
    return i;
}

typedef struct Abc_SccCost_t_ {
    int cNeg, cPhase, cPerm;
}
 Abc_SccCost_t;

func Abc_TgRecordPhase(pMan *Abc_TgMan_t, mode int32) Abc_SccCost_t {
    var vPhase *Vec_Int_t  = pMan.vPhase;
    var i int32;
    var j int32;
    var n int32 = pMan.pGroup.nGVars;
    var nStart int32;
    if mode == 0 {
        nStart = 1;
    } else {
        nStart = 0;
    }
    var nCoefs int32  = pMan.nGVars + 2 - nStart;
    var pScc, pMinScc [18]int32;
    var ret Abc_SccCost_t;

    if pMan.fPhased {
        ret.cNeg = 0;
        ret.cPhase = 0;
        Abc_TgCalcScc(pMan, pScc + 2, 1);
        ret.cPerm = Abc_TgPermCostScc(pMan, pScc + 2);
        return ret;
    }

    Abc_TgReorderFGrps(pMan);
    pMinScc[1] = Abc_TtScc(pMan.pTruth, pMan.nVars);
    Abc_TgCalcScc(pMan, pMinScc + 2, 1);
    pMinScc[0] = mode == 0 ? 0 : Abc_TgPermCostScc(pMan, pMinScc + 2);
    Vec_IntPush(vPhase, 0);
    for i = 0; (j = grayFlip(i)) < n; i++ {
        Abc_TgFlipSymGroupByVar(pMan, pMan.pFGrps[j]);
        pScc[1] = Abc_TtScc(pMan.pTruth, pMan.nVars);
        if mode == 0 && pScc[1] > pMinScc[1] { continue; }
        Abc_TgCalcScc(pMan, pScc + 2, 1);
        if mode > 0 {
            pScc[0] = Abc_TgPermCostScc(pMan, pScc + 2); }
        if Abc_TgCompareCoef(pScc + nStart, pMinScc + nStart, nCoefs) < 0 {
            memcpy(pMinScc + nStart, pScc + nStart, nCoefs * sizeof(int));
            Vec_IntClear(vPhase);
        }
        if Abc_TgCompareCoef(pScc + nStart, pMinScc + nStart, nCoefs) == 0 {
            Vec_IntPush(vPhase, grayCode(i+1)); }
    }
    Abc_TgFlipSymGroupByVar(pMan, pMan.pFGrps[n - 1]);

    ret.cNeg = n;
    ret.cPhase = ilog2(Vec_IntSize(vPhase));
    ret.cPerm = Abc_TgPermCostScc(pMan, pMinScc + 2);
    return ret;
}

static int Abc_TgRecordPhase1(Abc_TgMan_t * pMan)   // for AdjSE
{
    var vPhase *Vec_Int_t  = pMan.vPhase;
    var i int32;
    var j int32;
    var n int32 = pMan.pGroup.nGVars;
    //int nCoefs = pMan->nGVars + 2;
    var nScc, nMinScc int32;

    assert(Vec_IntSize(vPhase) == 0, "Vec_IntSize(vPhase) == 0");
    if pMan.fPhased { return 0; }

    Abc_TgReorderFGrps(pMan);
    nMinScc = Abc_TtScc(pMan.pTruth, pMan.nVars);
    Vec_IntPush(vPhase, 0);
    for i = 0; (j = grayFlip(i)) < n; i++ {
        Abc_TgFlipSymGroupByVar(pMan, pMan.pFGrps[j]);
        nScc = Abc_TtScc(pMan.pTruth, pMan.nVars);
        if nScc > nMinScc { continue; }
        if nScc < nMinScc {
            nMinScc = nScc;
            Vec_IntClear(vPhase);
        }
        Vec_IntPush(vPhase, grayCode(i + 1));
    }

    Abc_TgFlipSymGroupByVar(pMan, pMan.pFGrps[n - 1]);
    return ilog2(Vec_IntSize(vPhase));
}

var spStatic_Abc_TgPhaseEnumerationScc_pCopy [1024]word;
func Abc_TgPhaseEnumerationScc(pMan *Abc_TgMan_t, pBest *Abc_TgMan_t) {
    var vPhase *Vec_Int_t  = pMan.vPhase;
    var i int32;
    var j int32;
    var n int32 = pMan.pGroup.nGVars;
    var ph0, ph, flp int32;
    
    var tgManCopy Abc_TgMan_t;

    if pMan.fPhased {
        Abc_TgPermEnumerationScc(pMan, pBest);
        return;
    }

    Abc_TgManCopy(&tgManCopy, spStatic_Abc_TgPhaseEnumerationScc_pCopy, pMan);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPhase)) && (((ph) = Vec_IntEntry(vPhase, i)), 1); i++  {
        flp = ph ^ ph0;
        for j = 0; j < n; j++ {
            if flp & (1 << j) {
                Abc_TgFlipSymGroupByVar(&tgManCopy, pMan.pFGrps[j]); }
        }
        ph0 = ph;
        Abc_TgPermEnumerationScc(&tgManCopy, pBest);
    }
}

func Abc_TgFullEnumeration(pWork *Abc_TgMan_t, pBest *Abc_TgMan_t) {
    if pWork.nAlgorithm > 0 {
        Abc_TtFill(pBest.pTruth, Abc_TtWordNum(pBest.nVars));
        Abc_TgPhaseEnumerationScc(pWork, pBest);
    
    } else {
        Abc_TgFirstPermutation(pWork);
        do Abc_TgPhaseEnumeration(pWork, pBest);
        while (Abc_TgNextPermutation(pWork));
    }
    pBest.uPhase |= 1 << 30;
}

// runtime cost for Abc_TgRecordPhase (mode = 1)
func Abc_SccPhaseCost(pMan *Abc_TgMan_t) float64 {
    return pMan.nVars * 0.997 + pMan.nGVars * 1.043 - 15.9;
}

// runtime cost for Abc_TgFullEnumeration
func Abc_SccEnumCost(pMan *Abc_TgMan_t, c Abc_SccCost_t) float64 {
    switch pMan.nAlgorithm {
        case 0: return pMan.nVars + c.cPhase * 1.09 + c.cPerm * 0.01144;
        case 1: return pMan.nVars + c.cPhase * 0.855 + c.cPerm * 0.00797;
        case 2: return pMan.nVars * 0.94 + c.cPhase * 0.885 + c.cPerm * 0.00855 - 20.59;
    }
    return 0;
}

func Abc_TgEnumerationCost(pMan *Abc_TgMan_t) int32 {
    var i int32;
    var c Abc_SccCost_t  = Abc_SccCost_t{ 0,0,0 };
    if pMan.nGroups == 0 { return 0; }

    for i = 0; i < pMan.nGroups; i++ {
        c.cPerm += log2fn[int32(pMan.pGroup[i].nGVars)];
    }

    if pMan.fPhased {
        c.cPhase = 0;
    } else {
        if pMan.nAlgorithm == 0 {
            c.cPhase = pMan.pGroup.nGVars;
        } else {
            c.cPhase = Abc_TgRecordPhase1(pMan);
        }
    }

    // coefficients computed by linear regression
    return Abc_SccEnumCost(pMan, c) + 0.5;
}

/**Function*************************************************************

Synopsis    [Adaptive heuristic/exact canonical form computation.]

Description []

SideEffects []

SeeAlso     []

***********************************************************************/
var spStatic_Abc_TtCanonicizeAda_pCopy [1024]word;
func Abc_TtCanonicizeAda(p *Abc_TtHieMan_t, pTruth *word, nVars int32, pCanonPerm string, iThres int32) uint32 {
    var nWords int32  = Abc_TtWordNum(nVars);
    var fExac uint32;
    var fHash uint32 = 1 << 29;
    
    var tgMan, tgManCopy Abc_TgMan_t;
    var iCost int32;
    int MaxCost = 84;  // maximun posible cost for function with 16 inputs
    int nAlg = iThres >= 1000 ? 1 : 0;
    int fHigh = (iThres % 1000) >= 100, nEnumThres = iThres % 100;

    // handle constant
    if nVars == 0 {
        Abc_TtClear( pTruth, nWords );
        return 0;
    }

    Abc_TtVerifySmallTruth(pTruth, nVars);
                   
    Abc_TtCopy(gpVerCopy, pTruth, nWords, 0);
      

    assert(nVars <= 16, "nVars <= 16");
    assert(!((nAlg != 0) && p == nil), "!(nAlg && p == nil)");
    if (p != nil) && Abc_TtHieRetrieveOrInsert(p, -5, pTruth, pTruth) > 0 { return fHash; }
    Abc_TgInitMan(&tgMan, pTruth, nVars, nAlg, p ? p.vPhase : nil);
    Abc_TgCreateGroups(&tgMan);
    if (p != nil) && Abc_TtHieRetrieveOrInsert(p, -4, pTruth, pTruth) > 0 { return fHash; }
    Abc_TgPurgeSymmetry(&tgMan, fHigh);

    Abc_TgExpendSymmetry(&tgMan, pCanonPerm);
    Abc_TgImplementPerm(&tgMan, pCanonPerm);
    assert((Abc_TgCannonVerify(&tgMan) != 0), "Abc_TgCannonVerify(&tgMan)");

    iCost = Abc_TgEnumerationCost(&tgMan);
    if p == nil || p.nLastLevel == 0 {
        if nEnumThres > MaxCost || iCost < nEnumThres {
            Abc_TgManCopy(&tgManCopy, spStatic_Abc_TtCanonicizeAda_pCopy, &tgMan);
            Abc_TgFullEnumeration(&tgManCopy, &tgMan);
        
        } else {
            Abc_TgSimpleEnumeration(&tgMan); }
    
    } else {
        if nEnumThres > MaxCost || iCost < nEnumThres { fExac = 1 << 30; }
        if Abc_TtHieRetrieveOrInsert(p, -3, pTruth, pTruth) > 0 { return fHash + fExac; }
        Abc_TgManCopy(&tgManCopy, spStatic_Abc_TtCanonicizeAda_pCopy, &tgMan);
        Abc_TgSimpleEnumeration(&tgMan);
        if Abc_TtHieRetrieveOrInsert(p, -2, pTruth, pTruth) > 0 { return fHash + fExac; }
        if fExac != 0 {
            Abc_TgManCopy(&tgMan, pTruth, &tgManCopy);
            Abc_TgFullEnumeration(&tgManCopy, &tgMan);
        }
        Abc_TtHieRetrieveOrInsert(p, -1, pTruth, pTruth);
    }
    memcpy(pCanonPerm, tgMan.pPermT, sizeof(char) * nVars);

                   
    if !(Abc_TgCannonVerify(&tgMan) != 0) {
        fmt.Printf("Canonical form verification failed!\n"); }
      
    return tgMan.uPhase;
}

var spStatic_Abc_TtCanonicizeCA_pCopy [1024]word;
func Abc_TtCanonicizeCA(p *Abc_TtHieMan_t, pTruth *word, nVars int32, pCanonPerm string, fCA int32) uint32 {
    var nWords int32  = Abc_TtWordNum(nVars);
    var fHard uint32;
    var fHash uint32 = 1 << 29;
    
    var tgMan, tgManCopy Abc_TgMan_t;
    var sc Abc_SccCost_t;

    // handle constant
    if nVars == 0 {
        Abc_TtClear(pTruth, nWords);
        return 0;
    }

    Abc_TtVerifySmallTruth(pTruth, nVars);
                   
    Abc_TtCopy(gpVerCopy, pTruth, nWords, 0);
      

    assert(nVars <= 16, "nVars <= 16");
    assert(p != nil, "p != nil");
    if Abc_TtHieRetrieveOrInsert(p, -5, pTruth, pTruth) > 0 { return fHash; }
    Abc_TgInitMan(&tgMan, pTruth, nVars, 2, p.vPhase);

    Abc_TgCreateGroups(&tgMan);
    if (p != nil) && Abc_TtHieRetrieveOrInsert(p, -4, pTruth, pTruth) > 0 { return fHash; }
    Abc_TgPurgeSymmetry(&tgMan, 1);

    Abc_TgExpendSymmetry(&tgMan, pCanonPerm);
    Abc_TgImplementPerm(&tgMan, pCanonPerm);
    assert((Abc_TgCannonVerify(&tgMan) != 0), "Abc_TgCannonVerify(&tgMan)");

    if Abc_TtHieRetrieveOrInsert(p, -3, pTruth, pTruth) > 0 { return fHash; }
    Abc_TgManCopy(&tgManCopy, spStatic_Abc_TtCanonicizeCA_pCopy, &tgMan);
    Abc_TgSimpleEnumeration(&tgMan);
    if Abc_TtHieRetrieveOrInsert(p, -2, pTruth, pTruth) > 0 { return fHash; }
    Abc_TgManCopy(&tgMan, pTruth, &tgManCopy);
    Abc_TtFill(pTruth, nWords);

    assert((Abc_TgCannonVerify(&tgManCopy) != 0), "Abc_TgCannonVerify(&tgManCopy)");
    sc = Abc_TgRecordPhase(&tgManCopy, 0);
    if (fCA != 0) && Abc_SccEnumCost(&tgManCopy, sc) > Abc_SccPhaseCost(&tgManCopy) {
        Abc_TgResetGroup(&tgManCopy);
        sc = Abc_TgRecordPhase(&tgManCopy, 1);
        fHard = 1 << 28;
    }
    Abc_TgFullEnumeration(&tgManCopy, &tgMan);
    Abc_TtHieRetrieveOrInsert(p, -1, pTruth, pTruth);
    memcpy(pCanonPerm, tgMan.pPermT, sizeof(char) * nVars);

                   
    if !(Abc_TgCannonVerify(&tgMan) != 0) {
        fmt.Printf("Canonical form verification failed!\n"); }
      
    return tgMan.uPhase | fHard;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

